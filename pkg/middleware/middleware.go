package middleware

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// Middleware 中间件函数类型
type Middleware func(http.Handler) http.Handler

// Chain 中间件链
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Logger 日志中间件
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装响应写入器以捕获状态码
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		// 记录请求信息
		log.Printf(
			"%s %s %s %d %v",
			r.Method,
			r.RequestURI,
			r.RemoteAddr,
			rw.statusCode,
			time.Since(start),
		)
	})
}

// responseWriter 响应写入器
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	body       []byte
}

// WriteHeader 写入状态码
func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Write 写入响应体
func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = b
	return rw.ResponseWriter.Write(b)
}

// Recover 恢复中间件
func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// Timeout 超时中间件
func Timeout(timeout time.Duration) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-done:
				return
			case <-ctx.Done():
				http.Error(w, "Request Timeout", http.StatusRequestTimeout)
			}
		})
	}
}

// CORS CORS中间件
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Auth 认证中间件
func Auth(authFunc func(r *http.Request) bool) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !authFunc(r) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit 限流中间件
func RateLimit(limit int, window time.Duration) Middleware {
	type client struct {
		count     int
		lastReset time.Time
	}

	clients := make(map[string]*client)
	var mu sync.Mutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.RemoteAddr

			mu.Lock()
			c, exists := clients[ip]
			if !exists {
				c = &client{
					lastReset: time.Now(),
				}
				clients[ip] = c
			}

			// 检查是否需要重置计数器
			if time.Since(c.lastReset) > window {
				c.count = 0
				c.lastReset = time.Now()
			}

			// 检查是否超过限制
			if c.count >= limit {
				mu.Unlock()
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}

			c.count++
			mu.Unlock()

			next.ServeHTTP(w, r)
		})
	}
}

// Cache 缓存中间件
func Cache(duration time.Duration) Middleware {
	type cacheEntry struct {
		body       []byte
		headers    http.Header
		expiration time.Time
	}

	cache := make(map[string]*cacheEntry)
	var mu sync.RWMutex

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 只缓存GET请求
			if r.Method != http.MethodGet {
				next.ServeHTTP(w, r)
				return
			}

			// 检查缓存
			mu.RLock()
			if entry, ok := cache[r.URL.String()]; ok && time.Now().Before(entry.expiration) {
				// 复制头部
				for k, v := range entry.headers {
					w.Header()[k] = v
				}
				w.Write(entry.body)
				mu.RUnlock()
				return
			}
			mu.RUnlock()

			// 包装响应写入器
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK,
			}

			next.ServeHTTP(rw, r)

			// 缓存响应
			if rw.statusCode == http.StatusOK {
				mu.Lock()
				cache[r.URL.String()] = &cacheEntry{
					body:       rw.body,
					headers:    rw.Header(),
					expiration: time.Now().Add(duration),
				}
				mu.Unlock()
			}
		})
	}
}

// Metrics 指标中间件
func Metrics(next http.Handler) http.Handler {
	var (
		requests int64
		errors   int64
		latency  time.Duration
		mu       sync.RWMutex
	)

	// 启动指标收集器
	go func() {
		for {
			time.Sleep(time.Minute)
			mu.RLock()
			log.Printf("Metrics: requests=%d errors=%d avg_latency=%v",
				requests, errors, latency/time.Duration(requests))
			mu.RUnlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装响应写入器
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(rw, r)

		// 更新指标
		mu.Lock()
		requests++
		if rw.statusCode >= 400 {
			errors++
		}
		latency += time.Since(start)
		mu.Unlock()
	})
}
