package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	Logger(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}
}

func TestRecover(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	Recover(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d; got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestTimeout(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	Timeout(50*time.Millisecond)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusRequestTimeout {
		t.Errorf("expected status %d; got %d", http.StatusRequestTimeout, rec.Code)
	}
}

func TestCORS(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/test", nil)
			rec := httptest.NewRecorder()

			CORS(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Header().Get("Access-Control-Allow-Origin") != "*" {
				t.Error("expected CORS headers not set")
			}
		})
	}
}

func TestAuth(t *testing.T) {
	authFunc := func(r *http.Request) bool {
		return r.Header.Get("Authorization") == "valid"
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		auth           string
		expectedStatus int
	}{
		{
			name:           "valid auth",
			auth:           "valid",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid auth",
			auth:           "invalid",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set("Authorization", tt.auth)
			rec := httptest.NewRecorder()

			Auth(authFunc)(handler).ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d; got %d", tt.expectedStatus, rec.Code)
			}
		})
	}
}

func TestRateLimit(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// 测试限流
	rateLimited := RateLimit(2, time.Second)(handler)

	// 第一次请求应该成功
	rateLimited.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	// 第二次请求应该成功
	rec = httptest.NewRecorder()
	rateLimited.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	// 第三次请求应该被限流
	rec = httptest.NewRecorder()
	rateLimited.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Errorf("expected status %d; got %d", http.StatusTooManyRequests, rec.Code)
	}
}

func TestCache(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"test": "data"}`))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// 测试缓存
	cached := Cache(time.Second)(handler)

	// 第一次请求应该从处理器获取
	cached.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	// 第二次请求应该从缓存获取
	rec = httptest.NewRecorder()
	cached.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	// POST请求不应该被缓存
	req = httptest.NewRequest("POST", "/test", nil)
	rec = httptest.NewRecorder()
	cached.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}
}

func TestMetrics(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	// 测试指标收集
	metrics := Metrics(handler)

	// 正常请求
	metrics.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d; got %d", http.StatusOK, rec.Code)
	}

	// 错误请求
	errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	rec = httptest.NewRecorder()
	Metrics(errorHandler).ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d; got %d", http.StatusInternalServerError, rec.Code)
	}
}
