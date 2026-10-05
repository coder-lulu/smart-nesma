package response

import (
	"time"
)

// ErrorMetrics 错误指标
type ErrorMetrics struct {
	TotalErrors      int64                    `json:"total_errors"`
	ErrorsByType     map[string]int64         `json:"errors_by_type"`
	ErrorsBySeverity map[string]int64         `json:"errors_by_severity"`
	RetryAttempts    int64                    `json:"retry_attempts"`
	SuccessfulRetries int64                   `json:"successful_retries"`
	FailedRetries    int64                    `json:"failed_retries"`
	LastResetTime    time.Time                `json:"last_reset_time"`
}

// CircuitBreakerMetrics 断路器指标
type CircuitBreakerMetrics struct {
	TotalRequests      int64     `json:"total_requests"`
	SuccessfulRequests int64     `json:"successful_requests"`
	FailedRequests     int64     `json:"failed_requests"`
	RejectedRequests   int64     `json:"rejected_requests"`
	StateChanges       int64     `json:"state_changes"`
	LastStateChange    time.Time `json:"last_state_change"`
}

// ErrorMetricsResponse 错误指标响应
type ErrorMetricsResponse struct {
	ErrorMetrics          *ErrorMetrics                      `json:"error_metrics"`
	CircuitBreakerMetrics map[string]*CircuitBreakerMetrics  `json:"circuit_breaker_metrics"`
	Timestamp             time.Time                          `json:"timestamp"`
}

// CircuitBreakerStatus 断路器状态
type CircuitBreakerStatus struct {
	Name                string    `json:"name"`
	State               string    `json:"state"`               // CLOSED/OPEN/HALF_OPEN
	TotalRequests       int64     `json:"total_requests"`
	SuccessfulRequests  int64     `json:"successful_requests"`
	FailedRequests      int64     `json:"failed_requests"`
	RejectedRequests    int64     `json:"rejected_requests"`
	StateChanges        int64     `json:"state_changes"`
	LastStateChange     time.Time `json:"last_state_change"`
}

// CircuitBreakerStatusResponse 断路器状态响应
type CircuitBreakerStatusResponse struct {
	CircuitBreakers []CircuitBreakerStatus `json:"circuit_breakers"`
	TotalCount      int                    `json:"total_count"`
}

// ErrorTrend 错误趋势
type ErrorTrend struct {
	ErrorType   string  `json:"error_type"`
	Count       int64   `json:"count"`
	Percentage  float64 `json:"percentage"`
	Trend       string  `json:"trend"`      // increasing/decreasing/stable
}

// ErrorTrendSummary 错误趋势摘要
type ErrorTrendSummary struct {
	TotalErrors       int64  `json:"total_errors"`
	RetryAttempts     int64  `json:"retry_attempts"`
	SuccessfulRetries int64  `json:"successful_retries"`
	FailedRetries     int64  `json:"failed_retries"`
	OverallTrend      string `json:"overall_trend"`
}

// ErrorTrendsResponse 错误趋势响应
type ErrorTrendsResponse struct {
	TimeRange int                `json:"time_range"`
	Trends    []ErrorTrend       `json:"trends"`
	Summary   ErrorTrendSummary  `json:"summary"`
}

// TestErrorHandlingResponse 测试错误处理响应
type TestErrorHandlingResponse struct {
	TestType    string `json:"test_type"`
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	RetryCount  int    `json:"retry_count"`
	Duration    int64  `json:"duration"`    // 毫秒
	Details     string `json:"details"`
}