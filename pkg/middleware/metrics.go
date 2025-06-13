package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"
)

// Metrics 指标收集器
type Metrics struct {
	mu              sync.RWMutex
	RequestCount    int64
	RequestDuration time.Duration
	ErrorCount      int64
	StatusCodes     map[int]int64
	EndpointMetrics map[string]*EndpointMetric
}

// EndpointMetric 端点指标
type EndpointMetric struct {
	Count      int64
	Duration   time.Duration
	ErrorCount int64
	LastAccess time.Time
}

// NewMetrics 创建新的指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		StatusCodes:     make(map[int]int64),
		EndpointMetrics: make(map[string]*EndpointMetric),
	}
}

// MetricsMiddleware 指标收集中间件
func (m *Metrics) MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装响应写入器
		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		// 处理请求
		next.ServeHTTP(rw, r)

		// 记录指标
		duration := time.Since(start)
		m.recordMetrics(r.Method+" "+r.URL.Path, rw.statusCode, duration)
	})
}

// recordMetrics 记录指标
func (m *Metrics) recordMetrics(endpoint string, statusCode int, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 总体指标
	m.RequestCount++
	m.RequestDuration += duration
	m.StatusCodes[statusCode]++

	if statusCode >= 400 {
		m.ErrorCount++
	}

	// 端点指标
	if metric, exists := m.EndpointMetrics[endpoint]; exists {
		metric.Count++
		metric.Duration += duration
		metric.LastAccess = time.Now()
		if statusCode >= 400 {
			metric.ErrorCount++
		}
	} else {
		m.EndpointMetrics[endpoint] = &EndpointMetric{
			Count:      1,
			Duration:   duration,
			LastAccess: time.Now(),
		}
		if statusCode >= 400 {
			m.EndpointMetrics[endpoint].ErrorCount = 1
		}
	}
}

// GetMetrics 获取指标数据
func (m *Metrics) GetMetrics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	avgDuration := time.Duration(0)
	if m.RequestCount > 0 {
		avgDuration = m.RequestDuration / time.Duration(m.RequestCount)
	}

	endpoints := make(map[string]interface{})
	for endpoint, metric := range m.EndpointMetrics {
		avgEndpointDuration := time.Duration(0)
		if metric.Count > 0 {
			avgEndpointDuration = metric.Duration / time.Duration(metric.Count)
		}

		endpoints[endpoint] = map[string]interface{}{
			"count":        metric.Count,
			"avg_duration": avgEndpointDuration.String(),
			"error_count":  metric.ErrorCount,
			"last_access":  metric.LastAccess.Format(time.RFC3339),
		}
	}

	return map[string]interface{}{
		"total_requests": m.RequestCount,
		"avg_duration":   avgDuration.String(),
		"error_count":    m.ErrorCount,
		"status_codes":   m.StatusCodes,
		"endpoints":      endpoints,
		"uptime":         time.Since(time.Now()).String(),
	}
}

// Reset 重置指标
func (m *Metrics) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.RequestCount = 0
	m.RequestDuration = 0
	m.ErrorCount = 0
	m.StatusCodes = make(map[int]int64)
	m.EndpointMetrics = make(map[string]*EndpointMetric)
}

// MetricsHandler 指标查看处理器
func (m *Metrics) MetricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	metrics := m.GetMetrics()

	// 简单的JSON编码
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("{\n"))

	first := true
	for key, value := range metrics {
		if !first {
			w.Write([]byte(",\n"))
		}
		first = false

		w.Write([]byte("  \"" + key + "\": "))

		switch v := value.(type) {
		case string:
			w.Write([]byte("\"" + v + "\""))
		case int64:
			w.Write([]byte(strconv.FormatInt(v, 10)))
		case map[int]int64:
			w.Write([]byte("{\n"))
			firstInner := true
			for k, val := range v {
				if !firstInner {
					w.Write([]byte(",\n"))
				}
				firstInner = false
				w.Write([]byte("    \"" + strconv.Itoa(k) + "\": " + strconv.FormatInt(val, 10)))
			}
			w.Write([]byte("\n  }"))
		case map[string]interface{}:
			w.Write([]byte("{\n"))
			firstInner := true
			for k, val := range v {
				if !firstInner {
					w.Write([]byte(",\n"))
				}
				firstInner = false
				w.Write([]byte("    \"" + k + "\": "))

				switch innerV := val.(type) {
				case map[string]interface{}:
					w.Write([]byte("{\n"))
					firstInnerInner := true
					for kk, vv := range innerV {
						if !firstInnerInner {
							w.Write([]byte(",\n"))
						}
						firstInnerInner = false
						w.Write([]byte("      \"" + kk + "\": "))

						switch vvv := vv.(type) {
						case string:
							w.Write([]byte("\"" + vvv + "\""))
						case int64:
							w.Write([]byte(strconv.FormatInt(vvv, 10)))
						}
					}
					w.Write([]byte("\n    }"))
				}
			}
			w.Write([]byte("\n  }"))
		}
	}

	w.Write([]byte("\n}\n"))
}
