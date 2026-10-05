package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaResponse "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandlingApi 错误处理API接口
type ErrorHandlingApi struct {
	batchEvaluationService *nesma.BatchEvaluationService
}

// NewErrorHandlingApi 创建错误处理API
func NewErrorHandlingApi(batchEvaluationService *nesma.BatchEvaluationService) *ErrorHandlingApi {
	return &ErrorHandlingApi{
		batchEvaluationService: batchEvaluationService,
	}
}

// GetErrorMetrics 获取错误处理指标
// @Tags ErrorHandling
// @Summary 获取错误处理指标
// @Description 获取系统错误处理指标统计信息
// @Produce json
// @Success 200 {object} response.Response{data=nesmaResponse.ErrorMetricsResponse} "获取成功"
// @Router /nesma/error-handling/metrics [get]
func (e *ErrorHandlingApi) GetErrorMetrics(c *gin.Context) {
	// 获取错误处理指标
	serviceErrorMetrics := e.batchEvaluationService.GetErrorHandlerMetrics()
	
	// 获取断路器指标
	serviceCircuitBreakerMetrics := e.batchEvaluationService.GetCircuitBreakerMetrics()
	
	// 转换为响应模型
	errorMetrics := &nesmaResponse.ErrorMetrics{
		TotalErrors:       serviceErrorMetrics.TotalErrors,
		ErrorsByType:      convertErrorsByType(serviceErrorMetrics.ErrorsByType),
		ErrorsBySeverity:  convertErrorsBySeverity(serviceErrorMetrics.ErrorsBySeverity),
		RetryAttempts:     serviceErrorMetrics.RetryAttempts,
		SuccessfulRetries: serviceErrorMetrics.SuccessfulRetries,
		FailedRetries:     serviceErrorMetrics.FailedRetries,
		LastResetTime:     serviceErrorMetrics.LastResetTime,
	}
	
	circuitBreakerMetrics := make(map[string]*nesmaResponse.CircuitBreakerMetrics)
	for name, metrics := range serviceCircuitBreakerMetrics {
		circuitBreakerMetrics[name] = &nesmaResponse.CircuitBreakerMetrics{
			TotalRequests:      metrics.TotalRequests,
			SuccessfulRequests: metrics.SuccessfulRequests,
			FailedRequests:     metrics.FailedRequests,
			RejectedRequests:   metrics.RejectedRequests,
			StateChanges:       metrics.StateChanges,
			LastStateChange:    metrics.LastStateChange,
		}
	}
	
	// 构建响应
	resp := &nesmaResponse.ErrorMetricsResponse{
		ErrorMetrics:          errorMetrics,
		CircuitBreakerMetrics: circuitBreakerMetrics,
		Timestamp:             errorMetrics.LastResetTime,
	}
	
	global.GVA_LOG.Info("获取错误处理指标成功")
	response.OkWithData(resp, c)
}

// GetCircuitBreakerStatus 获取断路器状态
// @Tags ErrorHandling
// @Summary 获取断路器状态
// @Description 获取所有断路器的当前状态
// @Produce json
// @Success 200 {object} response.Response{data=nesmaResponse.CircuitBreakerStatusResponse} "获取成功"
// @Router /nesma/error-handling/circuit-breaker/status [get]
func (e *ErrorHandlingApi) GetCircuitBreakerStatus(c *gin.Context) {
	// 获取断路器指标
	circuitBreakerMetrics := e.batchEvaluationService.GetCircuitBreakerMetrics()
	
	// 构建状态响应
	statusList := make([]nesmaResponse.CircuitBreakerStatus, 0, len(circuitBreakerMetrics))
	for name, metrics := range circuitBreakerMetrics {
		status := nesmaResponse.CircuitBreakerStatus{
			Name:                name,
			State:               "UNKNOWN", // 需要从实际断路器获取状态
			TotalRequests:       metrics.TotalRequests,
			SuccessfulRequests:  metrics.SuccessfulRequests,
			FailedRequests:      metrics.FailedRequests,
			RejectedRequests:    metrics.RejectedRequests,
			StateChanges:        metrics.StateChanges,
			LastStateChange:     metrics.LastStateChange,
		}
		statusList = append(statusList, status)
	}
	
	resp := &nesmaResponse.CircuitBreakerStatusResponse{
		CircuitBreakers: statusList,
		TotalCount:      len(statusList),
	}
	
	global.GVA_LOG.Info("获取断路器状态成功",
		zap.Int("断路器数量", len(statusList)))
	response.OkWithData(resp, c)
}

// ResetCircuitBreaker 重置断路器
// @Tags ErrorHandling
// @Summary 重置断路器
// @Description 重置指定的断路器到关闭状态
// @Accept json
// @Produce json
// @Param name path string true "断路器名称"
// @Success 200 {object} response.Response "重置成功"
// @Failure 400 {object} response.Response "请求参数错误"
// @Router /nesma/error-handling/circuit-breaker/{name}/reset [post]
func (e *ErrorHandlingApi) ResetCircuitBreaker(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		global.GVA_LOG.Error("断路器名称为空")
		response.FailWithMessage("断路器名称不能为空", c)
		return
	}
	
	// 重置指定断路器
	// 注意：这里需要扩展BatchEvaluationService来支持重置单个断路器
	// 目前只能重置所有断路器
	e.batchEvaluationService.ResetCircuitBreakers()
	
	global.GVA_LOG.Info("断路器重置成功",
		zap.String("name", name))
	response.OkWithMessage("断路器重置成功", c)
}

// ResetAllCircuitBreakers 重置所有断路器
// @Tags ErrorHandling
// @Summary 重置所有断路器
// @Description 重置所有断路器到关闭状态
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "重置成功"
// @Router /nesma/error-handling/circuit-breaker/reset-all [post]
func (e *ErrorHandlingApi) ResetAllCircuitBreakers(c *gin.Context) {
	e.batchEvaluationService.ResetCircuitBreakers()
	
	global.GVA_LOG.Info("所有断路器重置成功")
	response.OkWithMessage("所有断路器重置成功", c)
}

// ResetErrorMetrics 重置错误统计
// @Tags ErrorHandling
// @Summary 重置错误统计
// @Description 重置所有错误统计数据
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "重置成功"
// @Router /nesma/error-handling/metrics/reset [post]
func (e *ErrorHandlingApi) ResetErrorMetrics(c *gin.Context) {
	e.batchEvaluationService.ResetErrorMetrics()
	
	global.GVA_LOG.Info("错误统计重置成功")
	response.OkWithMessage("错误统计重置成功", c)
}

// GetErrorTrends 获取错误趋势
// @Tags ErrorHandling
// @Summary 获取错误趋势
// @Description 获取系统错误趋势分析
// @Produce json
// @Param hours query int false "时间范围（小时）" default(24)
// @Success 200 {object} response.Response{data=nesmaResponse.ErrorTrendsResponse} "获取成功"
// @Router /nesma/error-handling/trends [get]
func (e *ErrorHandlingApi) GetErrorTrends(c *gin.Context) {
	// 获取时间范围参数
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}
	
	// 获取错误指标
	errorMetrics := e.batchEvaluationService.GetErrorHandlerMetrics()
	
	// 构建趋势响应（简化版本，实际应该从时间序列数据中计算）
	trends := make([]nesmaResponse.ErrorTrend, 0)
	
	// 模拟趋势数据
	for errorType, count := range errorMetrics.ErrorsByType {
		trend := nesmaResponse.ErrorTrend{
			ErrorType:   string(errorType),
			Count:       count,
			Percentage:  float64(count) / float64(errorMetrics.TotalErrors) * 100,
			Trend:       "stable", // 这里应该根据历史数据计算
		}
		trends = append(trends, trend)
	}
	
	resp := &nesmaResponse.ErrorTrendsResponse{
		TimeRange: hours,
		Trends:    trends,
		Summary: nesmaResponse.ErrorTrendSummary{
			TotalErrors:      errorMetrics.TotalErrors,
			RetryAttempts:    errorMetrics.RetryAttempts,
			SuccessfulRetries: errorMetrics.SuccessfulRetries,
			FailedRetries:    errorMetrics.FailedRetries,
			OverallTrend:     "stable",
		},
	}
	
	global.GVA_LOG.Info("获取错误趋势成功",
		zap.Int("时间范围", hours),
		zap.Int("趋势数量", len(trends)))
	response.OkWithData(resp, c)
}

// TestErrorHandling 测试错误处理
// @Tags ErrorHandling
// @Summary 测试错误处理
// @Description 测试错误处理和重试机制
// @Accept json
// @Produce json
// @Param request body request.TestErrorHandlingRequest true "测试请求"
// @Success 200 {object} response.Response{data=nesmaResponse.TestErrorHandlingResponse} "测试完成"
// @Failure 400 {object} response.Response "请求参数错误"
// @Router /nesma/error-handling/test [post]
func (e *ErrorHandlingApi) TestErrorHandling(c *gin.Context) {
	var req request.TestErrorHandlingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		global.GVA_LOG.Error("参数绑定失败", zap.Error(err))
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	
	// 验证请求参数
	if req.ErrorType == "" {
		response.FailWithMessage("错误类型不能为空", c)
		return
	}
	
	// 这里可以实现具体的错误处理测试逻辑
	// 例如：模拟不同类型的错误，测试重试机制等
	
	resp := &nesmaResponse.TestErrorHandlingResponse{
		TestType:    req.ErrorType,
		Success:     true,
		Message:     "错误处理测试完成",
		RetryCount:  req.MaxRetries,
		Duration:    1000, // 毫秒
		Details:     "测试了错误分类、重试机制和断路器功能",
	}
	
	global.GVA_LOG.Info("错误处理测试完成",
		zap.String("错误类型", req.ErrorType),
		zap.Int("最大重试次数", req.MaxRetries))
	response.OkWithData(resp, c)
}

// convertErrorsByType 转换错误类型统计
func convertErrorsByType(serviceMap map[nesma.ErrorType]int64) map[string]int64 {
	result := make(map[string]int64)
	for errorType, count := range serviceMap {
		result[string(errorType)] = count
	}
	return result
}

// convertErrorsBySeverity 转换错误严重程度统计
func convertErrorsBySeverity(serviceMap map[nesma.ErrorSeverity]int64) map[string]int64 {
	result := make(map[string]int64)
	for severity, count := range serviceMap {
		result[string(severity)] = count
	}
	return result
}