package nesma

import (
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	TaskID              uint      `json:"taskId"`
	StartTime           time.Time `json:"startTime"`
	EndTime             time.Time `json:"endTime"`
	Duration            float64   `json:"duration"`            // 总耗时(秒)
	TotalRequirements   int       `json:"totalRequirements"`   // 总需求数
	ProcessedCount      int       `json:"processedCount"`      // 已处理数
	SuccessCount        int       `json:"successCount"`        // 成功数
	FailedCount         int       `json:"failedCount"`         // 失败数
	AICallCount         int       `json:"aiCallCount"`         // AI调用次数
	WorkerCount         int       `json:"workerCount"`         // 工作协程数
	AvgProcessingTime   float64   `json:"avgProcessingTime"`   // 平均处理时间(秒)
	Throughput          float64   `json:"throughput"`          // 吞吐量(需求/秒)
	SuccessRate         float64   `json:"successRate"`         // 成功率
	ConcurrencyEfficiency float64 `json:"concurrencyEfficiency"` // 并发效率
}

// PerformanceMonitor 性能监控器
type PerformanceMonitor struct {
	metrics map[uint]*PerformanceMetrics
	mutex   sync.RWMutex
}

var globalMonitor = &PerformanceMonitor{
	metrics: make(map[uint]*PerformanceMetrics),
}

// StartMonitoring 开始监控任务
func (pm *PerformanceMonitor) StartMonitoring(taskID uint, totalRequirements, workerCount int) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	pm.metrics[taskID] = &PerformanceMetrics{
		TaskID:            taskID,
		StartTime:         time.Now(),
		TotalRequirements: totalRequirements,
		WorkerCount:       workerCount,
	}
	
	global.GVA_LOG.Info("开始性能监控", 
		zap.Uint("taskId", taskID),
		zap.Int("totalRequirements", totalRequirements),
		zap.Int("workerCount", workerCount))
}

// UpdateProgress 更新进度
func (pm *PerformanceMonitor) UpdateProgress(taskID uint, processedCount, successCount, failedCount, aiCallCount int) {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if metrics, exists := pm.metrics[taskID]; exists {
		metrics.ProcessedCount = processedCount
		metrics.SuccessCount = successCount
		metrics.FailedCount = failedCount
		metrics.AICallCount = aiCallCount
		
		// 计算实时指标
		if processedCount > 0 {
			elapsed := time.Since(metrics.StartTime).Seconds()
			metrics.AvgProcessingTime = elapsed / float64(processedCount)
			metrics.Throughput = float64(processedCount) / elapsed
			metrics.SuccessRate = float64(successCount) / float64(processedCount) * 100
		}
	}
}

// StopMonitoring 停止监控并计算最终指标
func (pm *PerformanceMonitor) StopMonitoring(taskID uint) *PerformanceMetrics {
	pm.mutex.Lock()
	defer pm.mutex.Unlock()
	
	if metrics, exists := pm.metrics[taskID]; exists {
		metrics.EndTime = time.Now()
		metrics.Duration = metrics.EndTime.Sub(metrics.StartTime).Seconds()
		
		// 计算最终指标
		if metrics.ProcessedCount > 0 {
			metrics.AvgProcessingTime = metrics.Duration / float64(metrics.ProcessedCount)
			metrics.Throughput = float64(metrics.ProcessedCount) / metrics.Duration
			metrics.SuccessRate = float64(metrics.SuccessCount) / float64(metrics.ProcessedCount) * 100
			
			// 计算并发效率 (实际吞吐量 / 理论最大吞吐量)
			theoreticalMaxThroughput := float64(metrics.WorkerCount) / metrics.AvgProcessingTime
			if theoreticalMaxThroughput > 0 {
				metrics.ConcurrencyEfficiency = metrics.Throughput / theoreticalMaxThroughput * 100
			}
		}
		
		// 记录性能报告
		pm.logPerformanceReport(metrics)
		
		// 清理内存
		delete(pm.metrics, taskID)
		
		return metrics
	}
	
	return nil
}

// logPerformanceReport 记录性能报告
func (pm *PerformanceMonitor) logPerformanceReport(metrics *PerformanceMetrics) {
	global.GVA_LOG.Info("性能监控报告",
		zap.Uint("taskId", metrics.TaskID),
		zap.Float64("duration", metrics.Duration),
		zap.Int("totalRequirements", metrics.TotalRequirements),
		zap.Int("processedCount", metrics.ProcessedCount),
		zap.Int("successCount", metrics.SuccessCount),
		zap.Int("failedCount", metrics.FailedCount),
		zap.Int("aiCallCount", metrics.AICallCount),
		zap.Int("workerCount", metrics.WorkerCount),
		zap.Float64("avgProcessingTime", metrics.AvgProcessingTime),
		zap.Float64("throughput", metrics.Throughput),
		zap.Float64("successRate", metrics.SuccessRate),
		zap.Float64("concurrencyEfficiency", metrics.ConcurrencyEfficiency))
}

// GetMetrics 获取任务指标
func (pm *PerformanceMonitor) GetMetrics(taskID uint) *PerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	return pm.metrics[taskID]
}

// GetAllMetrics 获取所有指标
func (pm *PerformanceMonitor) GetAllMetrics() map[uint]*PerformanceMetrics {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	result := make(map[uint]*PerformanceMetrics)
	for k, v := range pm.metrics {
		result[k] = v
	}
	return result
}

// GetPerformanceSummary 获取性能摘要
func (pm *PerformanceMonitor) GetPerformanceSummary() map[string]interface{} {
	pm.mutex.RLock()
	defer pm.mutex.RUnlock()
	
	var totalTasks int
	var totalRequirements int
	var totalProcessed int
	var totalSuccess int
	var totalFailed int
	var totalAICalls int
	var totalDuration float64
	
	for _, metrics := range pm.metrics {
		totalTasks++
		totalRequirements += metrics.TotalRequirements
		totalProcessed += metrics.ProcessedCount
		totalSuccess += metrics.SuccessCount
		totalFailed += metrics.FailedCount
		totalAICalls += metrics.AICallCount
		totalDuration += metrics.Duration
	}
	
	avgSuccessRate := 0.0
	if totalProcessed > 0 {
		avgSuccessRate = float64(totalSuccess) / float64(totalProcessed) * 100
	}
	
	avgThroughput := 0.0
	if totalDuration > 0 {
		avgThroughput = float64(totalProcessed) / totalDuration
	}
	
	return map[string]interface{}{
		"activeTasks":        totalTasks,
		"totalRequirements":  totalRequirements,
		"totalProcessed":     totalProcessed,
		"totalSuccess":       totalSuccess,
		"totalFailed":        totalFailed,
		"totalAICalls":       totalAICalls,
		"totalDuration":      totalDuration,
		"avgSuccessRate":     avgSuccessRate,
		"avgThroughput":      avgThroughput,
	}
}

// StartTaskMonitoring 开始任务监控的便捷方法
func StartTaskMonitoring(taskID uint, totalRequirements, workerCount int) {
	globalMonitor.StartMonitoring(taskID, totalRequirements, workerCount)
}

// UpdateTaskProgress 更新任务进度的便捷方法
func UpdateTaskProgress(taskID uint, processedCount, successCount, failedCount, aiCallCount int) {
	globalMonitor.UpdateProgress(taskID, processedCount, successCount, failedCount, aiCallCount)
}

// StopTaskMonitoring 停止任务监控的便捷方法
func StopTaskMonitoring(taskID uint) *PerformanceMetrics {
	return globalMonitor.StopMonitoring(taskID)
}

// GetTaskMetrics 获取任务指标的便捷方法
func GetTaskMetrics(taskID uint) *PerformanceMetrics {
	return globalMonitor.GetMetrics(taskID)
}

// GetPerformanceSummary 获取性能摘要的便捷方法
func GetPerformanceSummary() map[string]interface{} {
	return globalMonitor.GetPerformanceSummary()
} 