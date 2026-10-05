package request

// TestErrorHandlingRequest 测试错误处理请求
type TestErrorHandlingRequest struct {
	ErrorType   string `json:"error_type" binding:"required"`   // 错误类型：network/api/timeout/rate_limit/invalid_data/system
	MaxRetries  int    `json:"max_retries" binding:"min=0"`     // 最大重试次数
	Simulate    bool   `json:"simulate"`                        // 是否模拟错误
	Duration    int    `json:"duration"`                        // 测试持续时间（秒）
}