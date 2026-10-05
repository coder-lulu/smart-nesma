package nesma

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// ErrorType 错误类型枚举 (使用现有的定义)
type ErrorType string

const (
	ErrorTypeNetwork      ErrorType = "network_error"       // 网络错误
	ErrorTypeAPI          ErrorType = "api_error"           // API调用错误
	ErrorTypeInvalidData  ErrorType = "validation_error"    // 数据错误
	ErrorTypeUnknown      ErrorType = "unknown_error"       // 未知错误
)

// ErrorSeverity 错误严重程度
type ErrorSeverity string

const (
	SeverityLow      ErrorSeverity = "low"      // 低级错误，可重试
	SeverityMedium   ErrorSeverity = "medium"   // 中级错误，限制重试
	SeverityHigh     ErrorSeverity = "high"     // 高级错误，停止重试
	SeverityCritical ErrorSeverity = "critical" // 严重错误，立即停止
)

// RetryableError 可重试错误结构
type RetryableError struct {
	Type         ErrorType     `json:"type"`
	Severity     ErrorSeverity `json:"severity"`
	Message      string        `json:"message"`
	OriginalErr  error         `json:"-"`
	Retryable    bool          `json:"retryable"`
	RetryAfter   time.Duration `json:"retry_after"`
	Context      map[string]interface{} `json:"context"`
	Timestamp    time.Time     `json:"timestamp"`
}

func (e *RetryableError) Error() string {
	return fmt.Sprintf("[%s:%s] %s", e.Type, e.Severity, e.Message)
}

func (e *RetryableError) Unwrap() error {
	return e.OriginalErr
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries      int           `json:"max_retries"`       // 最大重试次数
	InitialDelay    time.Duration `json:"initial_delay"`     // 初始延迟
	MaxDelay        time.Duration `json:"max_delay"`         // 最大延迟
	BackoffFactor   float64       `json:"backoff_factor"`    // 退避因子
	Jitter          bool          `json:"jitter"`            // 是否添加随机抖动
	TimeoutPerRetry time.Duration `json:"timeout_per_retry"` // 每次重试超时时间
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = &RetryConfig{
	MaxRetries:      3,
	InitialDelay:    1 * time.Second,
	MaxDelay:        30 * time.Second,
	BackoffFactor:   2.0,
	Jitter:          true,
	TimeoutPerRetry: 60 * time.Second,
}

// ErrorHandler 错误处理器
type ErrorHandler struct {
	retryConfigs map[ErrorType]*RetryConfig
	metrics      *ErrorMetrics
}

// ErrorMetrics 错误统计
type ErrorMetrics struct {
	TotalErrors     int64                    `json:"total_errors"`
	ErrorsByType    map[ErrorType]int64      `json:"errors_by_type"`
	ErrorsBySeverity map[ErrorSeverity]int64 `json:"errors_by_severity"`
	RetryAttempts   int64                    `json:"retry_attempts"`
	SuccessfulRetries int64                  `json:"successful_retries"`
	FailedRetries   int64                    `json:"failed_retries"`
	LastResetTime   time.Time               `json:"last_reset_time"`
}

// NewErrorHandler 创建错误处理器
func NewErrorHandler() *ErrorHandler {
	handler := &ErrorHandler{
		retryConfigs: make(map[ErrorType]*RetryConfig),
		metrics: &ErrorMetrics{
			ErrorsByType:     make(map[ErrorType]int64),
			ErrorsBySeverity: make(map[ErrorSeverity]int64),
			LastResetTime:    time.Now(),
		},
	}

	// 设置默认重试配置
	handler.SetRetryConfig(ErrorTypeNetwork, &RetryConfig{
		MaxRetries:      5,
		InitialDelay:    2 * time.Second,
		MaxDelay:        60 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		TimeoutPerRetry: 30 * time.Second,
	})

	handler.SetRetryConfig(ErrorTypeAPI, &RetryConfig{
		MaxRetries:      3,
		InitialDelay:    1 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		TimeoutPerRetry: 60 * time.Second,
	})

	handler.SetRetryConfig(ErrorType(ErrorTypeTimeout), &RetryConfig{
		MaxRetries:      2,
		InitialDelay:    5 * time.Second,
		MaxDelay:        30 * time.Second,
		BackoffFactor:   1.5,
		Jitter:          true,
		TimeoutPerRetry: 120 * time.Second,
	})

	handler.SetRetryConfig(ErrorType(ErrorTypeRateLimit), &RetryConfig{
		MaxRetries:      5,
		InitialDelay:    10 * time.Second,
		MaxDelay:        300 * time.Second,
		BackoffFactor:   2.0,
		Jitter:          true,
		TimeoutPerRetry: 60 * time.Second,
	})

	return handler
}

// SetRetryConfig 设置重试配置
func (h *ErrorHandler) SetRetryConfig(errorType ErrorType, config *RetryConfig) {
	h.retryConfigs[errorType] = config
}

// ClassifyError 错误分类
func (h *ErrorHandler) ClassifyError(err error) *RetryableError {
	if err == nil {
		return nil
	}

	retryableErr := &RetryableError{
		OriginalErr: err,
		Message:     err.Error(),
		Timestamp:   time.Now(),
		Context:     make(map[string]interface{}),
	}

	// 根据错误内容进行分类
	errorMsg := err.Error()
	
	switch {
	case containsAny(errorMsg, []string{"connection", "network", "dns", "dial"}):
		retryableErr.Type = ErrorTypeNetwork
		retryableErr.Severity = SeverityMedium
		retryableErr.Retryable = true
		
	case containsAny(errorMsg, []string{"timeout", "deadline exceeded"}):
		retryableErr.Type = ErrorType(ErrorTypeTimeout)
		retryableErr.Severity = SeverityMedium
		retryableErr.Retryable = true
		
	case containsAny(errorMsg, []string{"rate limit", "too many requests", "429"}):
		retryableErr.Type = ErrorType(ErrorTypeRateLimit)
		retryableErr.Severity = SeverityLow
		retryableErr.Retryable = true
		retryableErr.RetryAfter = 30 * time.Second
		
	case containsAny(errorMsg, []string{"400", "401", "403", "invalid"}):
		retryableErr.Type = ErrorTypeInvalidData
		retryableErr.Severity = SeverityHigh
		retryableErr.Retryable = false
		
	case containsAny(errorMsg, []string{"500", "502", "503", "504"}):
		retryableErr.Type = ErrorTypeAPI
		retryableErr.Severity = SeverityMedium
		retryableErr.Retryable = true
		
	default:
		retryableErr.Type = ErrorTypeUnknown
		retryableErr.Severity = SeverityMedium
		retryableErr.Retryable = true
	}

	// 记录错误统计
	h.recordError(retryableErr)
	
	return retryableErr
}

// ExecuteWithRetry 执行带重试的操作
func (h *ErrorHandler) ExecuteWithRetry(ctx context.Context, operation func() error) error {
	return h.ExecuteWithRetryTyped(ctx, operation, ErrorTypeUnknown)
}

// ExecuteWithRetryTyped 执行指定错误类型的带重试操作
func (h *ErrorHandler) ExecuteWithRetryTyped(ctx context.Context, operation func() error, errorType ErrorType) error {
	config := h.getRetryConfig(errorType)
	
	var lastErr error
	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// 创建带超时的上下文
		_, cancel := context.WithTimeout(ctx, config.TimeoutPerRetry)
		
		// 执行操作
		err := operation()
		cancel()
		
		if err == nil {
			if attempt > 0 {
				h.metrics.SuccessfulRetries++
				global.GVA_LOG.Info("重试成功",
					zap.Int("attempt", attempt),
					zap.String("error_type", string(errorType)),
				)
			}
			return nil
		}
		
		// 分类错误
		retryableErr := h.ClassifyError(err)
		lastErr = retryableErr
		
		// 检查是否应该重试
		if !retryableErr.Retryable || retryableErr.Severity == SeverityCritical {
			global.GVA_LOG.Error("错误不可重试，停止重试",
				zap.Error(err),
				zap.String("type", string(retryableErr.Type)),
				zap.String("severity", string(retryableErr.Severity)),
			)
			h.metrics.FailedRetries++
			return retryableErr
		}
		
		// 检查是否达到最大重试次数
		if attempt >= config.MaxRetries {
			global.GVA_LOG.Error("达到最大重试次数",
				zap.Error(err),
				zap.Int("max_retries", config.MaxRetries),
				zap.String("error_type", string(errorType)),
			)
			h.metrics.FailedRetries++
			return retryableErr
		}
		
		// 计算延迟时间
		delay := h.calculateDelay(config, attempt, retryableErr.RetryAfter)
		
		global.GVA_LOG.Warn("操作失败，准备重试",
			zap.Error(err),
			zap.Int("attempt", attempt+1),
			zap.Int("max_retries", config.MaxRetries),
			zap.Duration("delay", delay),
			zap.String("error_type", string(errorType)),
		)
		
		// 等待后重试
		h.metrics.RetryAttempts++
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
			// 继续重试
		}
	}
	
	return lastErr
}

// getRetryConfig 获取重试配置
func (h *ErrorHandler) getRetryConfig(errorType ErrorType) *RetryConfig {
	if config, exists := h.retryConfigs[errorType]; exists {
		return config
	}
	return DefaultRetryConfig
}

// calculateDelay 计算延迟时间
func (h *ErrorHandler) calculateDelay(config *RetryConfig, attempt int, retryAfter time.Duration) time.Duration {
	var delay time.Duration
	
	// 如果有明确的重试时间建议，使用它
	if retryAfter > 0 {
		delay = retryAfter
	} else {
		// 指数退避算法
		delay = time.Duration(float64(config.InitialDelay) * math.Pow(config.BackoffFactor, float64(attempt)))
	}
	
	// 限制最大延迟
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	
	// 添加随机抖动
	if config.Jitter {
		jitter := time.Duration(rand.Float64() * float64(delay) * 0.1) // 10%的抖动
		delay += jitter
	}
	
	return delay
}

// recordError 记录错误统计
func (h *ErrorHandler) recordError(err *RetryableError) {
	h.metrics.TotalErrors++
	h.metrics.ErrorsByType[err.Type]++
	h.metrics.ErrorsBySeverity[err.Severity]++
}

// GetMetrics 获取错误统计
func (h *ErrorHandler) GetMetrics() *ErrorMetrics {
	return h.metrics
}

// ResetMetrics 重置错误统计
func (h *ErrorHandler) ResetMetrics() {
	h.metrics = &ErrorMetrics{
		ErrorsByType:     make(map[ErrorType]int64),
		ErrorsBySeverity: make(map[ErrorSeverity]int64),
		LastResetTime:    time.Now(),
	}
}

// containsAny 检查字符串是否包含任一关键词
func containsAny(s string, keywords []string) bool {
	for _, keyword := range keywords {
		if len(s) >= len(keyword) {
			for i := 0; i <= len(s)-len(keyword); i++ {
				if s[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// RecoverableOperation 可恢复操作接口
type RecoverableOperation interface {
	Execute() error
	Rollback() error
	GetOperationType() string
}

// TransactionalExecutor 事务执行器
type TransactionalExecutor struct {
	errorHandler *ErrorHandler
}

// NewTransactionalExecutor 创建事务执行器
func NewTransactionalExecutor(errorHandler *ErrorHandler) *TransactionalExecutor {
	return &TransactionalExecutor{
		errorHandler: errorHandler,
	}
}

// ExecuteWithRollback 执行带回滚的操作
func (e *TransactionalExecutor) ExecuteWithRollback(ctx context.Context, operation RecoverableOperation) error {
	return e.errorHandler.ExecuteWithRetry(ctx, func() error {
		err := operation.Execute()
		if err != nil {
			// 尝试回滚操作
			if rollbackErr := operation.Rollback(); rollbackErr != nil {
				global.GVA_LOG.Error("回滚操作失败",
					zap.Error(rollbackErr),
					zap.String("operation", operation.GetOperationType()),
					zap.Error(err),
				)
			}
		}
		return err
	})
}