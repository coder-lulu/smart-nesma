package nesma

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// CircuitBreakerState 断路器状态
type CircuitBreakerState int32

const (
	StateClosed   CircuitBreakerState = iota // 关闭状态：正常工作
	StateOpen                                // 开启状态：快速失败
	StateHalfOpen                            // 半开状态：尝试恢复
)

func (s CircuitBreakerState) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreakerConfig 断路器配置
type CircuitBreakerConfig struct {
	FailureThreshold   int           `json:"failure_threshold"`    // 失败阈值
	SuccessThreshold   int           `json:"success_threshold"`    // 成功阈值（半开状态）
	Timeout           time.Duration `json:"timeout"`              // 超时时间
	ResetTimeout      time.Duration `json:"reset_timeout"`        // 重置超时时间
	MaxConcurrentCalls int          `json:"max_concurrent_calls"` // 最大并发调用数
}

// DefaultCircuitBreakerConfig 默认断路器配置
var DefaultCircuitBreakerConfig = &CircuitBreakerConfig{
	FailureThreshold:   5,
	SuccessThreshold:   3,
	Timeout:           60 * time.Second,
	ResetTimeout:      30 * time.Second,
	MaxConcurrentCalls: 100,
}

// CircuitBreakerMetrics 断路器指标
type CircuitBreakerMetrics struct {
	TotalRequests     int64 `json:"total_requests"`
	SuccessfulRequests int64 `json:"successful_requests"`
	FailedRequests    int64 `json:"failed_requests"`
	RejectedRequests  int64 `json:"rejected_requests"`
	StateChanges      int64 `json:"state_changes"`
	LastStateChange   time.Time `json:"last_state_change"`
}

// CircuitBreaker 断路器实现
type CircuitBreaker struct {
	name           string
	config         *CircuitBreakerConfig
	state          int32 // CircuitBreakerState的原子操作
	failures       int64
	successes      int64
	requests       int64
	lastFailTime   time.Time
	mutex          sync.RWMutex
	metrics        *CircuitBreakerMetrics
	onStateChange  func(name string, from, to CircuitBreakerState)
}

// NewCircuitBreaker 创建断路器
func NewCircuitBreaker(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = DefaultCircuitBreakerConfig
	}
	
	cb := &CircuitBreaker{
		name:    name,
		config:  config,
		state:   int32(StateClosed),
		metrics: &CircuitBreakerMetrics{
			LastStateChange: time.Now(),
		},
	}
	
	return cb
}

// SetOnStateChange 设置状态变化回调
func (cb *CircuitBreaker) SetOnStateChange(callback func(name string, from, to CircuitBreakerState)) {
	cb.onStateChange = callback
}

// Execute 执行操作
func (cb *CircuitBreaker) Execute(operation func() error) error {
	return cb.ExecuteWithFallback(operation, nil)
}

// ExecuteWithFallback 执行操作（带降级）
func (cb *CircuitBreaker) ExecuteWithFallback(operation func() error, fallback func() error) error {
	// 检查是否可以执行
	if !cb.canExecute() {
		atomic.AddInt64(&cb.metrics.RejectedRequests, 1)
		if fallback != nil {
			global.GVA_LOG.Warn("断路器开启，执行降级操作",
				zap.String("circuit_breaker", cb.name),
				zap.String("state", cb.GetState().String()),
			)
			return fallback()
		}
		return fmt.Errorf("circuit breaker '%s' is %s", cb.name, cb.GetState().String())
	}
	
	// 执行操作
	atomic.AddInt64(&cb.metrics.TotalRequests, 1)
	atomic.AddInt64(&cb.requests, 1)
	
	err := operation()
	
	if err != nil {
		cb.onFailure()
		atomic.AddInt64(&cb.metrics.FailedRequests, 1)
		return err
	}
	
	cb.onSuccess()
	atomic.AddInt64(&cb.metrics.SuccessfulRequests, 1)
	return nil
}

// canExecute 检查是否可以执行
func (cb *CircuitBreaker) canExecute() bool {
	state := cb.GetState()
	
	switch state {
	case StateClosed:
		return true
	case StateOpen:
		return cb.shouldAttemptReset()
	case StateHalfOpen:
		return cb.getRequests() < int64(cb.config.MaxConcurrentCalls)
	default:
		return false
	}
}

// shouldAttemptReset 是否应该尝试重置
func (cb *CircuitBreaker) shouldAttemptReset() bool {
	cb.mutex.RLock()
	defer cb.mutex.RUnlock()
	
	return time.Since(cb.lastFailTime) >= cb.config.ResetTimeout
}

// onSuccess 成功处理
func (cb *CircuitBreaker) onSuccess() {
	state := cb.GetState()
	
	switch state {
	case StateClosed:
		// 重置失败计数
		atomic.StoreInt64(&cb.failures, 0)
		
	case StateHalfOpen:
		successes := atomic.AddInt64(&cb.successes, 1)
		if successes >= int64(cb.config.SuccessThreshold) {
			cb.setState(StateClosed)
		}
	}
}

// onFailure 失败处理
func (cb *CircuitBreaker) onFailure() {
	failures := atomic.AddInt64(&cb.failures, 1)
	
	cb.mutex.Lock()
	cb.lastFailTime = time.Now()
	cb.mutex.Unlock()
	
	state := cb.GetState()
	
	switch state {
	case StateClosed:
		if failures >= int64(cb.config.FailureThreshold) {
			cb.setState(StateOpen)
		}
		
	case StateHalfOpen:
		cb.setState(StateOpen)
	}
}

// setState 设置状态
func (cb *CircuitBreaker) setState(newState CircuitBreakerState) {
	oldState := CircuitBreakerState(atomic.SwapInt32(&cb.state, int32(newState)))
	
	if oldState != newState {
		atomic.StoreInt64(&cb.failures, 0)
		atomic.StoreInt64(&cb.successes, 0)
		atomic.StoreInt64(&cb.requests, 0)
		
		cb.metrics.StateChanges++
		cb.metrics.LastStateChange = time.Now()
		
		global.GVA_LOG.Info("断路器状态变化",
			zap.String("circuit_breaker", cb.name),
			zap.String("from", oldState.String()),
			zap.String("to", newState.String()),
		)
		
		// 状态变化为半开时，启动监控
		if newState == StateHalfOpen {
			go cb.monitorHalfOpenState()
		}
		
		// 调用状态变化回调
		if cb.onStateChange != nil {
			cb.onStateChange(cb.name, oldState, newState)
		}
	}
}

// monitorHalfOpenState 监控半开状态
func (cb *CircuitBreaker) monitorHalfOpenState() {
	time.Sleep(cb.config.ResetTimeout)
	
	if cb.GetState() == StateHalfOpen {
		// 如果还在半开状态，检查是否应该切换到开启状态
		if atomic.LoadInt64(&cb.requests) == 0 {
			cb.setState(StateOpen)
		}
	}
}

// GetState 获取当前状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	state := CircuitBreakerState(atomic.LoadInt32(&cb.state))
	
	// 检查是否需要从开启状态切换到半开状态
	if state == StateOpen && cb.shouldAttemptReset() {
		// 尝试原子性地切换到半开状态
		if atomic.CompareAndSwapInt32(&cb.state, int32(StateOpen), int32(StateHalfOpen)) {
			cb.setState(StateHalfOpen) // 触发状态变化逻辑
		}
		return StateHalfOpen
	}
	
	return state
}

// GetName 获取断路器名称
func (cb *CircuitBreaker) GetName() string {
	return cb.name
}

// GetMetrics 获取指标
func (cb *CircuitBreaker) GetMetrics() *CircuitBreakerMetrics {
	return cb.metrics
}

// getRequests 获取当前请求数
func (cb *CircuitBreaker) getRequests() int64 {
	return atomic.LoadInt64(&cb.requests)
}

// Reset 重置断路器
func (cb *CircuitBreaker) Reset() {
	cb.setState(StateClosed)
	atomic.StoreInt64(&cb.failures, 0)
	atomic.StoreInt64(&cb.successes, 0)
	atomic.StoreInt64(&cb.requests, 0)
	
	global.GVA_LOG.Info("断路器已重置",
		zap.String("circuit_breaker", cb.name),
	)
}

// CircuitBreakerManager 断路器管理器
type CircuitBreakerManager struct {
	breakers map[string]*CircuitBreaker
	mutex    sync.RWMutex
}

// NewCircuitBreakerManager 创建断路器管理器
func NewCircuitBreakerManager() *CircuitBreakerManager {
	return &CircuitBreakerManager{
		breakers: make(map[string]*CircuitBreaker),
	}
}

// GetOrCreate 获取或创建断路器
func (m *CircuitBreakerManager) GetOrCreate(name string, config *CircuitBreakerConfig) *CircuitBreaker {
	m.mutex.RLock()
	if breaker, exists := m.breakers[name]; exists {
		m.mutex.RUnlock()
		return breaker
	}
	m.mutex.RUnlock()
	
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	// 双重检查
	if breaker, exists := m.breakers[name]; exists {
		return breaker
	}
	
	breaker := NewCircuitBreaker(name, config)
	m.breakers[name] = breaker
	
	global.GVA_LOG.Info("创建新的断路器",
		zap.String("name", name),
		zap.Int("failure_threshold", config.FailureThreshold),
		zap.Duration("reset_timeout", config.ResetTimeout),
	)
	
	return breaker
}

// Get 获取断路器
func (m *CircuitBreakerManager) Get(name string) (*CircuitBreaker, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	breaker, exists := m.breakers[name]
	return breaker, exists
}

// GetAll 获取所有断路器
func (m *CircuitBreakerManager) GetAll() map[string]*CircuitBreaker {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	result := make(map[string]*CircuitBreaker, len(m.breakers))
	for name, breaker := range m.breakers {
		result[name] = breaker
	}
	
	return result
}

// Remove 移除断路器
func (m *CircuitBreakerManager) Remove(name string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	delete(m.breakers, name)
	
	global.GVA_LOG.Info("移除断路器",
		zap.String("name", name),
	)
}

// ResetAll 重置所有断路器
func (m *CircuitBreakerManager) ResetAll() {
	m.mutex.RLock()
	breakers := make([]*CircuitBreaker, 0, len(m.breakers))
	for _, breaker := range m.breakers {
		breakers = append(breakers, breaker)
	}
	m.mutex.RUnlock()
	
	for _, breaker := range breakers {
		breaker.Reset()
	}
	
	global.GVA_LOG.Info("重置所有断路器",
		zap.Int("count", len(breakers)),
	)
}

// GetMetricsReport 获取指标报告
func (m *CircuitBreakerManager) GetMetricsReport() map[string]*CircuitBreakerMetrics {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	report := make(map[string]*CircuitBreakerMetrics, len(m.breakers))
	for name, breaker := range m.breakers {
		report[name] = breaker.GetMetrics()
	}
	
	return report
}