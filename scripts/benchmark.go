package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

// BenchmarkConfig 性能测试配置
type BenchmarkConfig struct {
	BaseURL     string `json:"base_url"`
	Concurrency int    `json:"concurrency"`
	Requests    int    `json:"requests"`
	Timeout     int    `json:"timeout"`
}

// BenchmarkResult 性能测试结果
type BenchmarkResult struct {
	TotalRequests   int           `json:"total_requests"`
	SuccessRequests int           `json:"success_requests"`
	FailedRequests  int           `json:"failed_requests"`
	TotalTime       time.Duration `json:"total_time"`
	AvgTime         time.Duration `json:"avg_time"`
	MinTime         time.Duration `json:"min_time"`
	MaxTime         time.Duration `json:"max_time"`
	RequestsPerSec  float64       `json:"requests_per_sec"`
}

// RequestResult 单个请求结果
type RequestResult struct {
	Duration   time.Duration
	StatusCode int
	Error      error
}

func main() {
	config := &BenchmarkConfig{
		BaseURL:     "http://localhost:8080",
		Concurrency: 10,
		Requests:    1000,
		Timeout:     30,
	}

	fmt.Println("🚀 开始家族树系统性能测试...")
	fmt.Printf("配置: 并发数=%d, 请求数=%d, 超时=%ds\n",
		config.Concurrency, config.Requests, config.Timeout)

	// 测试不同的端点
	endpoints := []struct {
		name   string
		method string
		path   string
		body   interface{}
	}{
		{"健康检查", "GET", "/health", nil},
		{"搜索个人", "GET", "/api/v1/individuals?q=张", nil},
		{"获取个人详情", "GET", "/api/v1/individuals/1", nil},
		{"获取家族树", "GET", "/api/v1/individuals/1/family-tree?generations=3", nil},
		{"创建个人", "POST", "/api/v1/individuals", map[string]interface{}{
			"full_name": "测试用户",
			"gender":    "male",
		}},
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\n📊 测试端点: %s %s\n", endpoint.method, endpoint.path)
		result := runBenchmark(config, endpoint.method, endpoint.path, endpoint.body)
		printResult(result)
	}

	fmt.Println("\n✅ 性能测试完成!")
}

func runBenchmark(config *BenchmarkConfig, method, path string, body interface{}) *BenchmarkResult {
	var wg sync.WaitGroup
	results := make(chan RequestResult, config.Requests)

	// 限制并发数
	semaphore := make(chan struct{}, config.Concurrency)

	startTime := time.Now()

	// 发送请求
	for i := 0; i < config.Requests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := makeRequest(config.BaseURL+path, method, body, config.Timeout)
			results <- result
		}()
	}

	// 等待所有请求完成
	wg.Wait()
	close(results)

	totalTime := time.Since(startTime)

	// 统计结果
	var successCount, failedCount int
	var totalDuration, minDuration, maxDuration time.Duration
	minDuration = time.Hour // 初始化为一个大值

	for result := range results {
		if result.Error == nil && result.StatusCode < 400 {
			successCount++
		} else {
			failedCount++
		}

		totalDuration += result.Duration
		if result.Duration < minDuration {
			minDuration = result.Duration
		}
		if result.Duration > maxDuration {
			maxDuration = result.Duration
		}
	}

	avgDuration := totalDuration / time.Duration(config.Requests)
	requestsPerSec := float64(config.Requests) / totalTime.Seconds()

	return &BenchmarkResult{
		TotalRequests:   config.Requests,
		SuccessRequests: successCount,
		FailedRequests:  failedCount,
		TotalTime:       totalTime,
		AvgTime:         avgDuration,
		MinTime:         minDuration,
		MaxTime:         maxDuration,
		RequestsPerSec:  requestsPerSec,
	}
}

func makeRequest(url, method string, body interface{}, timeoutSec int) RequestResult {
	client := &http.Client{
		Timeout: time.Duration(timeoutSec) * time.Second,
	}

	var reqBody []byte
	if body != nil {
		var err error
		reqBody, err = json.Marshal(body)
		if err != nil {
			return RequestResult{Error: err}
		}
	}

	start := time.Now()

	req, err := http.NewRequest(method, url, bytes.NewBuffer(reqBody))
	if err != nil {
		return RequestResult{Error: err, Duration: time.Since(start)}
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return RequestResult{Error: err, Duration: duration}
	}
	defer resp.Body.Close()

	// 读取响应体以确保完整的请求周期
	ioutil.ReadAll(resp.Body)

	return RequestResult{
		Duration:   duration,
		StatusCode: resp.StatusCode,
	}
}

func printResult(result *BenchmarkResult) {
	fmt.Printf("总请求数: %d\n", result.TotalRequests)
	fmt.Printf("成功请求: %d\n", result.SuccessRequests)
	fmt.Printf("失败请求: %d\n", result.FailedRequests)
	fmt.Printf("总耗时: %v\n", result.TotalTime)
	fmt.Printf("平均响应时间: %v\n", result.AvgTime)
	fmt.Printf("最小响应时间: %v\n", result.MinTime)
	fmt.Printf("最大响应时间: %v\n", result.MaxTime)
	fmt.Printf("每秒请求数: %.2f\n", result.RequestsPerSec)
	fmt.Printf("成功率: %.2f%%\n", float64(result.SuccessRequests)/float64(result.TotalRequests)*100)
}
