package nesma

import (
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// AIPerformanceMetrics AI性能指标
type AIPerformanceMetrics struct {
	ModelName           string    `json:"model_name"`
	RequestCount        int64     `json:"request_count"`
	SuccessCount        int64     `json:"success_count"`
	FailureCount        int64     `json:"failure_count"`
	TotalTokens         int64     `json:"total_tokens"`
	AvgResponseTime     float64   `json:"avg_response_time_ms"`
	AvgTokensPerRequest float64   `json:"avg_tokens_per_request"`
	AvgConfidence       float64   `json:"avg_confidence"`
	LastUsed            time.Time `json:"last_used"`
	ErrorRate           float64   `json:"error_rate"`
	
	// 响应时间分布
	ResponseTimeP50  float64 `json:"response_time_p50"`
	ResponseTimeP90  float64 `json:"response_time_p90"`
	ResponseTimeP99  float64 `json:"response_time_p99"`
	
	// 质量指标
	HighConfidenceCount   int64   `json:"high_confidence_count"`   // 置信度>0.8的请求数
	MediumConfidenceCount int64   `json:"medium_confidence_count"` // 置信度0.5-0.8的请求数
	LowConfidenceCount    int64   `json:"low_confidence_count"`    // 置信度<0.5的请求数
	
	// 成本统计
	EstimatedCost        float64 `json:"estimated_cost_usd"`
	CostPerThousandTokens float64 `json:"cost_per_thousand_tokens"`
}

// AIRequestRecord AI请求记录
type AIRequestRecord struct {
	ID            string                 `json:"id"`
	ModelName     string                 `json:"model_name"`
	RequestTime   time.Time              `json:"request_time"`
	ResponseTime  time.Duration          `json:"response_time"`
	TokensUsed    int                    `json:"tokens_used"`
	Confidence    float64                `json:"confidence"`
	Success       bool                   `json:"success"`
	ErrorMessage  string                 `json:"error_message,omitempty"`
	RequestType   string                 `json:"request_type"`  // nesma_analysis, description_generation, etc.
	Complexity    AITaskComplexity       `json:"complexity"`
	Context       map[string]interface{} `json:"context,omitempty"`
}

// AIPerformanceMonitor AI性能监控器
type AIPerformanceMonitor struct {
	metrics      map[string]*AIPerformanceMetrics
	records      []AIRequestRecord
	mu           sync.RWMutex
	maxRecords   int
	
	// 实时统计
	responseTimes map[string][]float64  // 按模型存储响应时间
	
	// 配置参数
	highConfidenceThreshold  float64
	mediumConfidenceThreshold float64
	costPerTokens           map[string]float64  // 每个模型的每千token成本
}

// NewAIPerformanceMonitor 创建AI性能监控器
func NewAIPerformanceMonitor() *AIPerformanceMonitor {
	return &AIPerformanceMonitor{
		metrics:                  make(map[string]*AIPerformanceMetrics),
		records:                  make([]AIRequestRecord, 0),
		maxRecords:              10000, // 最多保留1万条记录
		responseTimes:           make(map[string][]float64),
		highConfidenceThreshold: 0.8,
		mediumConfidenceThreshold: 0.5,
		costPerTokens: map[string]float64{
			"deepseek-chat":     0.14 / 1000,    // $0.14 per 1M tokens
			"deepseek-reasoner": 0.55 / 1000,    // $0.55 per 1M tokens (估算)
			"gpt-4":            30.0 / 1000,     // $30 per 1M tokens
			"gpt-4-turbo":      10.0 / 1000,     // $10 per 1M tokens
		},
	}
}

// RecordRequest 记录AI请求
func (m *AIPerformanceMonitor) RecordRequest(record AIRequestRecord) {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// 添加记录
	m.records = append(m.records, record)
	
	// 限制记录数量
	if len(m.records) > m.maxRecords {
		m.records = m.records[len(m.records)-m.maxRecords:]
	}
	
	// 更新指标
	m.updateMetrics(record)
	
	global.GVA_LOG.Debug("记录AI请求性能", 
		zap.String("model", record.ModelName),
		zap.Duration("response_time", record.ResponseTime),
		zap.Int("tokens", record.TokensUsed),
		zap.Float64("confidence", record.Confidence),
		zap.Bool("success", record.Success))
}

// updateMetrics 更新性能指标
func (m *AIPerformanceMonitor) updateMetrics(record AIRequestRecord) {
	modelName := record.ModelName
	
	// 获取或创建模型指标
	if m.metrics[modelName] == nil {
		m.metrics[modelName] = &AIPerformanceMetrics{
			ModelName: modelName,
		}
	}
	
	metrics := m.metrics[modelName]
	
	// 更新基础计数
	metrics.RequestCount++
	if record.Success {
		metrics.SuccessCount++
	} else {
		metrics.FailureCount++
	}
	
	// 更新token统计
	metrics.TotalTokens += int64(record.TokensUsed)
	metrics.AvgTokensPerRequest = float64(metrics.TotalTokens) / float64(metrics.RequestCount)
	
	// 更新响应时间
	responseTimeMs := float64(record.ResponseTime.Nanoseconds()) / 1e6
	if m.responseTimes[modelName] == nil {
		m.responseTimes[modelName] = make([]float64, 0)
	}
	m.responseTimes[modelName] = append(m.responseTimes[modelName], responseTimeMs)
	
	// 限制响应时间记录数量
	if len(m.responseTimes[modelName]) > 1000 {
		m.responseTimes[modelName] = m.responseTimes[modelName][len(m.responseTimes[modelName])-1000:]
	}
	
	// 计算平均响应时间
	totalTime := 0.0
	for _, rt := range m.responseTimes[modelName] {
		totalTime += rt
	}
	metrics.AvgResponseTime = totalTime / float64(len(m.responseTimes[modelName]))
	
	// 计算响应时间分位数
	m.calculatePercentiles(modelName, metrics)
	
	// 更新置信度统计
	if record.Success {
		if record.Confidence >= m.highConfidenceThreshold {
			metrics.HighConfidenceCount++
		} else if record.Confidence >= m.mediumConfidenceThreshold {
			metrics.MediumConfidenceCount++
		} else {
			metrics.LowConfidenceCount++
		}
		
		// 计算平均置信度
		totalConfidenceRequests := metrics.HighConfidenceCount + metrics.MediumConfidenceCount + metrics.LowConfidenceCount
		if totalConfidenceRequests > 0 {
			weightedConfidence := float64(metrics.HighConfidenceCount)*0.9 + 
								 float64(metrics.MediumConfidenceCount)*0.65 + 
								 float64(metrics.LowConfidenceCount)*0.25
			metrics.AvgConfidence = weightedConfidence / float64(totalConfidenceRequests)
		}
	}
	
	// 更新错误率
	if metrics.RequestCount > 0 {
		metrics.ErrorRate = float64(metrics.FailureCount) / float64(metrics.RequestCount)
	}
	
	// 更新成本估算
	if costPerToken, exists := m.costPerTokens[modelName]; exists {
		metrics.CostPerThousandTokens = costPerToken * 1000
		metrics.EstimatedCost = float64(metrics.TotalTokens) * costPerToken
	}
	
	// 更新最后使用时间
	metrics.LastUsed = record.RequestTime
}

// calculatePercentiles 计算响应时间分位数
func (m *AIPerformanceMonitor) calculatePercentiles(modelName string, metrics *AIPerformanceMetrics) {
	times := make([]float64, len(m.responseTimes[modelName]))
	copy(times, m.responseTimes[modelName])
	
	if len(times) == 0 {
		return
	}
	
	// 简单排序计算分位数
	for i := 0; i < len(times); i++ {
		for j := i + 1; j < len(times); j++ {
			if times[i] > times[j] {
				times[i], times[j] = times[j], times[i]
			}
		}
	}
	
	n := len(times)
	metrics.ResponseTimeP50 = times[n*50/100]
	metrics.ResponseTimeP90 = times[n*90/100]
	metrics.ResponseTimeP99 = times[n*99/100]
}

// GetMetrics 获取所有模型的性能指标
func (m *AIPerformanceMonitor) GetMetrics() map[string]*AIPerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 深拷贝指标
	result := make(map[string]*AIPerformanceMetrics)
	for name, metrics := range m.metrics {
		metricsCopy := *metrics
		result[name] = &metricsCopy
	}
	
	return result
}

// GetModelMetrics 获取特定模型的性能指标
func (m *AIPerformanceMonitor) GetModelMetrics(modelName string) *AIPerformanceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if metrics, exists := m.metrics[modelName]; exists {
		metricsCopy := *metrics
		return &metricsCopy
	}
	
	return nil
}

// GetRecentRecords 获取最近的请求记录
func (m *AIPerformanceMonitor) GetRecentRecords(limit int) []AIRequestRecord {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if limit <= 0 || limit > len(m.records) {
		limit = len(m.records)
	}
	
	start := len(m.records) - limit
	if start < 0 {
		start = 0
	}
	
	result := make([]AIRequestRecord, limit)
	copy(result, m.records[start:])
	
	return result
}

// GetModelRecommendation 获取模型推荐
func (m *AIPerformanceMonitor) GetModelRecommendation(complexity AITaskComplexity) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	// 根据复杂度和性能指标推荐模型
	switch complexity {
	case ComplexityHigh:
		// 高复杂度任务优先推荐reasoner模型
		if reasonerMetrics := m.metrics["deepseek-reasoner"]; reasonerMetrics != nil && 
			reasonerMetrics.AvgConfidence > 0.8 && reasonerMetrics.ErrorRate < 0.1 {
			return "deepseek-reasoner"
		}
		return "deepseek-chat"
		
	case ComplexityMedium:
		// 中等复杂度任务比较模型性能
		chatMetrics := m.metrics["deepseek-chat"]
		reasonerMetrics := m.metrics["deepseek-reasoner"]
		
		if chatMetrics != nil && reasonerMetrics != nil {
			// 计算性价比分数 (置信度/响应时间)
			chatScore := chatMetrics.AvgConfidence / (chatMetrics.AvgResponseTime / 1000)
			reasonerScore := reasonerMetrics.AvgConfidence / (reasonerMetrics.AvgResponseTime / 1000)
			
			if reasonerScore > chatScore*1.2 { // reasoner比chat好20%以上才推荐
				return "deepseek-reasoner"
			}
		}
		return "deepseek-chat"
		
	case ComplexityLow:
		// 低复杂度任务使用普通模型
		return "deepseek-chat"
		
	default:
		return "deepseek-chat"
	}
}

// GeneratePerformanceReport 生成性能报告
func (m *AIPerformanceMonitor) GeneratePerformanceReport() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	report := make(map[string]interface{})
	
	// 总体统计
	totalRequests := int64(0)
	totalSuccesses := int64(0)
	totalTokens := int64(0)
	totalCost := 0.0
	
	for _, metrics := range m.metrics {
		totalRequests += metrics.RequestCount
		totalSuccesses += metrics.SuccessCount
		totalTokens += metrics.TotalTokens
		totalCost += metrics.EstimatedCost
	}
	
	report["summary"] = map[string]interface{}{
		"total_requests":    totalRequests,
		"total_successes":   totalSuccesses,
		"total_tokens":      totalTokens,
		"total_cost_usd":    totalCost,
		"overall_success_rate": func() float64 {
			if totalRequests > 0 {
				return float64(totalSuccesses) / float64(totalRequests)
			}
			return 0
		}(),
		"active_models": len(m.metrics),
	}
	
	// 各模型详细指标
	report["models"] = m.GetMetrics()
	
	// 最近24小时统计
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	recentRequests := 0
	recentSuccesses := 0
	
	for _, record := range m.records {
		if record.RequestTime.After(oneDayAgo) {
			recentRequests++
			if record.Success {
				recentSuccesses++
			}
		}
	}
	
	report["recent_24h"] = map[string]interface{}{
		"requests":     recentRequests,
		"successes":    recentSuccesses,
		"success_rate": func() float64 {
			if recentRequests > 0 {
				return float64(recentSuccesses) / float64(recentRequests)
			}
			return 0
		}(),
	}
	
	// 性能建议
	recommendations := m.generateRecommendations()
	report["recommendations"] = recommendations
	
	return report
}

// generateRecommendations 生成性能改进建议
func (m *AIPerformanceMonitor) generateRecommendations() []string {
	recommendations := make([]string, 0)
	
	for modelName, metrics := range m.metrics {
		// 检查错误率
		if metrics.ErrorRate > 0.05 { // 错误率超过5%
			recommendations = append(recommendations, 
				fmt.Sprintf("模型 %s 错误率较高(%.2f%%)，建议检查API配置和网络连接", 
					modelName, metrics.ErrorRate*100))
		}
		
		// 检查响应时间
		if metrics.AvgResponseTime > 30000 { // 超过30秒
			recommendations = append(recommendations, 
				fmt.Sprintf("模型 %s 响应时间较长(%.2fs)，建议优化提示词长度或考虑使用更快的模型", 
					modelName, metrics.AvgResponseTime/1000))
		}
		
		// 检查置信度
		if metrics.AvgConfidence < 0.7 {
			recommendations = append(recommendations, 
				fmt.Sprintf("模型 %s 平均置信度较低(%.2f)，建议优化提示词设计或增加上下文信息", 
					modelName, metrics.AvgConfidence))
		}
		
		// 检查成本效率
		if metrics.EstimatedCost > 100 { // 成本超过100美元
			recommendations = append(recommendations, 
				fmt.Sprintf("模型 %s 使用成本较高($%.2f)，建议考虑使用更经济的模型或优化token使用", 
					modelName, metrics.EstimatedCost))
		}
	}
	
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "AI服务运行状态良好，性能指标都在正常范围内")
	}
	
	return recommendations
}

// 全局AI性能监控器实例
var globalAIPerformanceMonitor *AIPerformanceMonitor
var aiPerformanceMonitorOnce sync.Once

// GetAIPerformanceMonitor 获取AI性能监控器单例
func GetAIPerformanceMonitor() *AIPerformanceMonitor {
	aiPerformanceMonitorOnce.Do(func() {
		globalAIPerformanceMonitor = NewAIPerformanceMonitor()
	})
	return globalAIPerformanceMonitor
}