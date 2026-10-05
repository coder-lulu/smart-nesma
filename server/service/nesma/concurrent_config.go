package nesma

import "fmt"

// ConcurrentConfig 并发配置
type ConcurrentConfig struct {
	// AI分析相关配置
	MaxAnalysisWorkers    int `json:"max_analysis_workers"`    // 最大分析工作协程数
	MaxOptimizationWorkers int `json:"max_optimization_workers"` // 最大优化工作协程数
	MaxEvaluationWorkers   int `json:"max_evaluation_workers"`   // 最大评估工作协程数
	
	// 批量更新配置
	AnalysisBatchSize      int `json:"analysis_batch_size"`      // 分析批量更新大小
	OptimizationBatchSize  int `json:"optimization_batch_size"`  // 优化批量更新大小
	EvaluationBatchSize    int `json:"evaluation_batch_size"`    // 评估批量更新大小
	
	// 通道缓冲区大小
	TaskChannelBuffer      int `json:"task_channel_buffer"`      // 任务通道缓冲区大小
	ResultChannelBuffer    int `json:"result_channel_buffer"`    // 结果通道缓冲区大小
	
	// 超时配置
	WorkerTimeout          int `json:"worker_timeout"`           // 工作协程超时时间(秒)
	BatchUpdateTimeout     int `json:"batch_update_timeout"`     // 批量更新超时时间(秒)
}

// DefaultConcurrentConfig 默认并发配置
var DefaultConcurrentConfig = ConcurrentConfig{
	MaxAnalysisWorkers:      5,  // 5个并发分析工作协程
	MaxOptimizationWorkers:  5,  // 5个并发优化工作协程
	MaxEvaluationWorkers:    5,  // 5个并发评估工作协程
	
	AnalysisBatchSize:       10, // 每10个结果批量更新一次
	OptimizationBatchSize:   10, // 每10个结果批量更新一次
	EvaluationBatchSize:     10, // 每10个结果批量更新一次
	
	TaskChannelBuffer:       100, // 任务通道缓冲区100
	ResultChannelBuffer:     100, // 结果通道缓冲区100
	
	WorkerTimeout:           60,  // 60秒超时
	BatchUpdateTimeout:      30,  // 30秒批量更新超时
}

// GetConcurrentConfig 获取并发配置
func GetConcurrentConfig() ConcurrentConfig {
	return DefaultConcurrentConfig
}

// UpdateConcurrentConfig 更新并发配置
func UpdateConcurrentConfig(config ConcurrentConfig) {
	DefaultConcurrentConfig = config
}

// ValidateConcurrentConfig 验证并发配置
func ValidateConcurrentConfig(config ConcurrentConfig) error {
	if config.MaxAnalysisWorkers <= 0 {
		return fmt.Errorf("max_analysis_workers must be greater than 0")
	}
	if config.MaxOptimizationWorkers <= 0 {
		return fmt.Errorf("max_optimization_workers must be greater than 0")
	}
	if config.MaxEvaluationWorkers <= 0 {
		return fmt.Errorf("max_evaluation_workers must be greater than 0")
	}
	if config.AnalysisBatchSize <= 0 {
		return fmt.Errorf("analysis_batch_size must be greater than 0")
	}
	if config.OptimizationBatchSize <= 0 {
		return fmt.Errorf("optimization_batch_size must be greater than 0")
	}
	if config.EvaluationBatchSize <= 0 {
		return fmt.Errorf("evaluation_batch_size must be greater than 0")
	}
	return nil
} 