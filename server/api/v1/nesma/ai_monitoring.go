package nesma

import (
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
)

type AIMonitoringApi struct{}

// @Tags NESMA-AI监控
// @Summary 获取AI性能报告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma/ai/performance-report [get]
func (a *AIMonitoringApi) GetPerformanceReport(c *gin.Context) {
	aiManager := nesmaService.GetAIServiceManager()
	report := aiManager.GetPerformanceReport()
	
	response.OkWithData(report, c)
}

// @Tags NESMA-AI监控
// @Summary 获取特定模型的性能指标
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param model_name path string true "模型名称"
// @Success 200 {object} response.Response{data=nesmaService.AIPerformanceMetrics} "获取成功"
// @Router /nesma/ai/model/{model_name}/metrics [get]
func (a *AIMonitoringApi) GetModelMetrics(c *gin.Context) {
	modelName := c.Param("model_name")
	if modelName == "" {
		response.FailWithMessage("模型名称不能为空", c)
		return
	}

	aiManager := nesmaService.GetAIServiceManager()
	metrics := aiManager.GetModelMetrics(modelName)
	
	if metrics == nil {
		response.FailWithMessage("模型未找到或无性能数据", c)
		return
	}

	response.OkWithData(metrics, c)
}

// @Tags NESMA-AI监控
// @Summary 获取所有可用模型列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=[]string} "获取成功"
// @Router /nesma/ai/models [get]
func (a *AIMonitoringApi) GetAvailableModels(c *gin.Context) {
	aiManager := nesmaService.GetAIServiceManager()
	models := aiManager.GetAvailableModels()
	
	response.OkWithData(map[string]interface{}{
		"models": models,
		"total":  len(models),
	}, c)
}

// @Tags NESMA-AI监控
// @Summary 检查AI服务健康状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]bool} "获取成功"
// @Router /nesma/ai/health [get]
func (a *AIMonitoringApi) CheckServiceHealth(c *gin.Context) {
	aiManager := nesmaService.GetAIServiceManager()
	health := aiManager.CheckServiceHealth(c.Request.Context())
	
	allHealthy := true
	for _, isHealthy := range health {
		if !isHealthy {
			allHealthy = false
			break
		}
	}
	
	response.OkWithData(map[string]interface{}{
		"services":    health,
		"all_healthy": allHealthy,
	}, c)
}

// @Tags NESMA-AI监控
// @Summary 获取最近的AI请求记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param limit query int false "记录数量限制" default(50)
// @Success 200 {object} response.Response{data=[]nesmaService.AIRequestRecord} "获取成功"
// @Router /nesma/ai/recent-requests [get]
func (a *AIMonitoringApi) GetRecentRequests(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	
	monitor := nesmaService.GetAIPerformanceMonitor()
	records := monitor.GetRecentRecords(limit)
	
	response.OkWithData(map[string]interface{}{
		"records": records,
		"total":   len(records),
		"limit":   limit,
	}, c)
}

// @Tags NESMA-AI监控
// @Summary 获取所有模型的性能指标概览
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma/ai/metrics-overview [get]
func (a *AIMonitoringApi) GetMetricsOverview(c *gin.Context) {
	monitor := nesmaService.GetAIPerformanceMonitor()
	metrics := monitor.GetMetrics()
	
	// 计算汇总统计
	totalRequests := int64(0)
	totalSuccesses := int64(0)
	totalTokens := int64(0)
	totalCost := 0.0
	avgResponseTime := 0.0
	avgConfidence := 0.0
	modelCount := 0
	
	for _, m := range metrics {
		totalRequests += m.RequestCount
		totalSuccesses += m.SuccessCount
		totalTokens += m.TotalTokens
		totalCost += m.EstimatedCost
		if m.RequestCount > 0 {
			avgResponseTime += m.AvgResponseTime
			avgConfidence += m.AvgConfidence
			modelCount++
		}
	}
	
	if modelCount > 0 {
		avgResponseTime = avgResponseTime / float64(modelCount)
		avgConfidence = avgConfidence / float64(modelCount)
	}
	
	summary := map[string]interface{}{
		"total_requests":    totalRequests,
		"total_successes":   totalSuccesses,
		"total_tokens":      totalTokens,
		"total_cost_usd":    totalCost,
		"success_rate":      func() float64 {
			if totalRequests > 0 {
				return float64(totalSuccesses) / float64(totalRequests)
			}
			return 0
		}(),
		"avg_response_time": avgResponseTime,
		"avg_confidence":    avgConfidence,
		"active_models":     len(metrics),
	}
	
	response.OkWithData(map[string]interface{}{
		"summary": summary,
		"models":  metrics,
	}, c)
}

// @Tags NESMA-AI监控
// @Summary 获取模型推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param complexity query string true "任务复杂度" Enums(low,medium,high)
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma/ai/model-recommendation [get]
func (a *AIMonitoringApi) GetModelRecommendation(c *gin.Context) {
	complexityStr := c.Query("complexity")
	if complexityStr == "" {
		response.FailWithMessage("请指定任务复杂度", c)
		return
	}
	
	var complexity nesmaService.AITaskComplexity
	switch complexityStr {
	case "low":
		complexity = nesmaService.ComplexityLow
	case "medium":
		complexity = nesmaService.ComplexityMedium
	case "high":
		complexity = nesmaService.ComplexityHigh
	default:
		response.FailWithMessage("无效的复杂度参数，支持：low/medium/high", c)
		return
	}
	
	monitor := nesmaService.GetAIPerformanceMonitor()
	recommendedModel := monitor.GetModelRecommendation(complexity)
	
	// 获取推荐模型的性能指标
	metrics := monitor.GetModelMetrics(recommendedModel)
	
	response.OkWithData(map[string]interface{}{
		"recommended_model": recommendedModel,
		"complexity":       complexityStr,
		"model_metrics":    metrics,
		"reason":          fmt.Sprintf("基于%s复杂度任务的历史性能数据推荐", complexityStr),
	}, c)
}

// @Tags NESMA-AI监控
// @Summary 获取AI使用统计趋势
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param hours query int false "统计时间范围(小时)" default(24)
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma/ai/usage-trends [get]
func (a *AIMonitoringApi) GetUsageTrends(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, err := strconv.Atoi(hoursStr)
	if err != nil || hours <= 0 {
		hours = 24
	}
	
	monitor := nesmaService.GetAIPerformanceMonitor()
	
	// 获取最近的记录进行趋势分析
	allRecords := monitor.GetRecentRecords(1000) // 获取最近1000条记录
	
	// 按小时分组统计
	hourlyStats := make(map[string]map[string]interface{})
	
	cutoffTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	for _, record := range allRecords {
		if record.RequestTime.Before(cutoffTime) {
			continue
		}
		
		hourKey := record.RequestTime.Format("2006-01-02 15:00")
		if hourlyStats[hourKey] == nil {
			hourlyStats[hourKey] = map[string]interface{}{
				"requests":     0,
				"successes":    0,
				"total_tokens": 0,
				"avg_confidence": 0.0,
				"models":      make(map[string]int),
			}
		}
		
		stats := hourlyStats[hourKey]
		stats["requests"] = stats["requests"].(int) + 1
		if record.Success {
			stats["successes"] = stats["successes"].(int) + 1
		}
		stats["total_tokens"] = stats["total_tokens"].(int) + record.TokensUsed
		
		// 更新模型使用统计
		models := stats["models"].(map[string]int)
		models[record.ModelName]++
		
		// 计算平均置信度（简化计算）
		if record.Success {
			currentAvg := stats["avg_confidence"].(float64)
			requests := stats["requests"].(int)
			stats["avg_confidence"] = (currentAvg*float64(requests-1) + record.Confidence) / float64(requests)
		}
	}
	
	response.OkWithData(map[string]interface{}{
		"time_range_hours": hours,
		"hourly_stats":     hourlyStats,
		"total_hours":      len(hourlyStats),
	}, c)
}