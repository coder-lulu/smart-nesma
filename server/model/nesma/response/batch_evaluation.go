package response

import (
	"time"
)

// BatchEvaluationTaskResponse 批量评估任务响应
type BatchEvaluationTaskResponse struct {
	TaskID            string                       `json:"task_id"`
	EvaluationID      string                       `json:"evaluation_id"`
	ProjectID         uint                         `json:"project_id"`
	CycleID           uint                         `json:"cycle_id"`
	Status            string                       `json:"status"`
	Progress          float64                      `json:"progress"`
	TotalBatches      int                          `json:"total_batches"`
	TotalRequirements int                          `json:"total_requirements"`
	CreatedAt         time.Time                    `json:"created_at"`
	StartTime         *time.Time                   `json:"start_time,omitempty"`
	EndTime           *time.Time                   `json:"end_time,omitempty"`
	EstimatedDuration time.Duration                `json:"estimated_duration"`
	ProcessingTime    time.Duration                `json:"processing_time"`
	Config            map[string]interface{}       `json:"config"`
}

// BatchEvaluationProgressResponse 批量评估进度响应
type BatchEvaluationProgressResponse struct {
	TaskID            string                    `json:"task_id"`
	EvaluationID      string                    `json:"evaluation_id"`
	Status            string                    `json:"status"`
	Progress          float64                   `json:"progress"`
	CurrentBatch      int                       `json:"current_batch"`
	TotalBatches      int                       `json:"total_batches"`
	ProcessedCount    int                       `json:"processed_count"`
	TotalRequirements int                       `json:"total_requirements"`
	StartTime         *time.Time                `json:"start_time,omitempty"`
	EstimatedEndTime  *time.Time                `json:"estimated_end_time,omitempty"`
	Message           string                    `json:"message"`
	BatchDetails      []BatchDetailResponse     `json:"batch_details,omitempty"`
	RecentResults     []EvaluationResultResponse `json:"recent_results,omitempty"`
}

// BatchDetailResponse 批次详情响应
type BatchDetailResponse struct {
	BatchID         string        `json:"batch_id"`
	BatchNumber     int           `json:"batch_number"`
	Status          string        `json:"status"`
	RequirementCount int           `json:"requirement_count"`
	TokenEstimate   int           `json:"token_estimate"`
	ActualTokens    int           `json:"actual_tokens"`
	ProcessingTime  time.Duration `json:"processing_time"`
	StartTime       *time.Time    `json:"start_time,omitempty"`
	EndTime         *time.Time    `json:"end_time,omitempty"`
	ErrorMessage    string        `json:"error_message,omitempty"`
	ModelUsed       string        `json:"model_used"`
	RetryCount      int           `json:"retry_count"`
	QualityScore    float64       `json:"quality_score"`
	ConfidenceAvg   float64       `json:"confidence_avg"`
}

// EvaluationResultResponse 评估结果响应
type EvaluationResultResponse struct {
	RequirementID    uint          `json:"requirement_id"`
	FunctionType     string        `json:"function_type"`
	ComplexityLevel  string        `json:"complexity_level"`
	ConfidenceScore  float64       `json:"confidence_score"`
	AFP              float64       `json:"afp"`
	UFP              float64       `json:"ufp"`
	Timestamp        time.Time     `json:"timestamp"`
	BatchID          string        `json:"batch_id"`
	ModelUsed        string        `json:"model_used"`
	ProcessingTime   time.Duration `json:"processing_time"`
	ErrorMsg         string        `json:"error_msg,omitempty"`
	RequirementTitle string        `json:"requirement_title,omitempty"`
	RequirementDesc  string        `json:"requirement_desc,omitempty"`
	OptimizedDesc    string        `json:"optimized_desc,omitempty"`
	FunctionDetail   string        `json:"function_detail,omitempty"`
}

// BatchEvaluationResultResponse 批量评估结果响应
type BatchEvaluationResultResponse struct {
	TaskID           string                       `json:"task_id"`
	EvaluationID     string                       `json:"evaluation_id"`
	Status           string                       `json:"status"`
	TotalResults     int64                        `json:"total_results"`
	Results          []EvaluationResultResponse   `json:"results"`
	Summary          *BatchEvaluationSummary      `json:"summary,omitempty"`
	QualityMetrics   *BatchEvaluationQualityMetrics `json:"quality_metrics,omitempty"`
	TokenUsage       *BatchEvaluationTokenUsage   `json:"token_usage,omitempty"`
	ProcessingTime   time.Duration                `json:"processing_time"`
	Page             int                          `json:"page"`
	PageSize         int                          `json:"page_size"`
}

// BatchEvaluationSummary 批量评估摘要
type BatchEvaluationSummary struct {
	TotalRequirements    int                    `json:"total_requirements"`
	ProcessedRequirements int                   `json:"processed_requirements"`
	SuccessRate          float64               `json:"success_rate"`
	AverageConfidence    float64               `json:"average_confidence"`
	FunctionTypeStats    map[string]int        `json:"function_type_stats"`
	ComplexityStats      map[string]int        `json:"complexity_stats"`
	TotalAFP             float64               `json:"total_afp"`
	TotalUFP             float64               `json:"total_ufp"`
	QualityDistribution  map[string]int        `json:"quality_distribution"`
	ProcessingStats      ProcessingStatsResponse `json:"processing_stats"`
}

// ProcessingStatsResponse 处理统计响应
type ProcessingStatsResponse struct {
	TotalBatches      int           `json:"total_batches"`
	CompletedBatches  int           `json:"completed_batches"`
	FailedBatches     int           `json:"failed_batches"`
	AverageBatchTime  time.Duration `json:"average_batch_time"`
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	ThroughputPerHour int           `json:"throughput_per_hour"`
}

// BatchEvaluationQualityMetrics 批量评估质量指标
type BatchEvaluationQualityMetrics struct {
	OverallAccuracy     float64                `json:"overall_accuracy"`
	ConsistencyScore    float64                `json:"consistency_score"`
	CompletionRate      float64                `json:"completion_rate"`
	ConfidenceAverage   float64                `json:"confidence_average"`
	ErrorRate           float64                `json:"error_rate"`
	ProcessingSpeed     float64                `json:"processing_speed"`
	ModelPerformance    map[string]float64     `json:"model_performance"`
	QualityTrends       []QualityTrendResponse `json:"quality_trends"`
	BatchQualityScores  []float64              `json:"batch_quality_scores"`
}

// QualityTrendResponse 质量趋势响应
type QualityTrendResponse struct {
	Timestamp      time.Time `json:"timestamp"`
	BatchNumber    int       `json:"batch_number"`
	AccuracyScore  float64   `json:"accuracy_score"`
	ConfidenceAvg  float64   `json:"confidence_avg"`
	ProcessingTime float64   `json:"processing_time"`
	QualityScore   float64   `json:"quality_score"`
}

// BatchEvaluationTokenUsage 批量评估Token使用情况
type BatchEvaluationTokenUsage struct {
	TotalTokens        int                    `json:"total_tokens"`
	InputTokens        int                    `json:"input_tokens"`
	OutputTokens       int                    `json:"output_tokens"`
	TokensPerBatch     []int                  `json:"tokens_per_batch"`
	TokensPerModel     map[string]int         `json:"tokens_per_model"`
	EstimatedCost      float64                `json:"estimated_cost"`
	CostPerModel       map[string]float64     `json:"cost_per_model"`
	TokenEfficiency    float64                `json:"token_efficiency"`
	CostEfficiency     float64                `json:"cost_efficiency"`
}

// BatchEvaluationListResponse 批量评估列表响应
type BatchEvaluationListResponse struct {
	Tasks    []BatchEvaluationTaskResponse `json:"tasks"`
	Total    int64                         `json:"total"`
	Page     int                           `json:"page"`
	PageSize int                           `json:"page_size"`
}

// BatchEvaluationDetailResponse 批量评估详情响应
type BatchEvaluationDetailResponse struct {
	TaskID            string                             `json:"task_id"`
	EvaluationID      string                             `json:"evaluation_id"`
	ProjectID         uint                               `json:"project_id"`
	CycleID           uint                               `json:"cycle_id"`
	SourceVersionID   uint                               `json:"source_version_id"`
	TargetVersionID   uint                               `json:"target_version_id"`
	Status            string                             `json:"status"`
	Progress          float64                            `json:"progress"`
	CurrentBatch      int                                `json:"current_batch"`
	TotalBatches      int                                `json:"total_batches"`
	CompletedBatches  int                                `json:"completed_batches"`
	ProcessedCount    int                                `json:"processed_count"`
	TotalRequirements int                                `json:"total_requirements"`
	CreatedAt         time.Time                          `json:"created_at"`
	StartTime         *time.Time                         `json:"start_time,omitempty"`
	EndTime           *time.Time                         `json:"end_time,omitempty"`
	EstimatedEndTime  *time.Time                         `json:"estimated_end_time,omitempty"`
	EstimatedDuration time.Duration                      `json:"estimated_duration"`
	ProcessingTime    time.Duration                      `json:"processing_time"`
	StatusMessage     string                             `json:"status_message"`
	Config            map[string]interface{}             `json:"config"`
	Summary           *BatchEvaluationSummary            `json:"summary,omitempty"`
	QualityMetrics    *BatchEvaluationQualityMetrics     `json:"quality_metrics,omitempty"`
	TokenUsage        *BatchEvaluationTokenUsage         `json:"token_usage,omitempty"`
	BatchDetails      []BatchDetailResponse              `json:"batch_details"`
	Results           []EvaluationResultResponse         `json:"results"`
}

// BatchEvaluationCompareResponse 批量评估比较响应
type BatchEvaluationCompareResponse struct {
	Task1         BatchEvaluationTaskResponse  `json:"task1"`
	Task2         BatchEvaluationTaskResponse  `json:"task2"`
	Comparison    ComparisonResult             `json:"comparison"`
	Differences   []DifferenceItem             `json:"differences"`
	Similarities  []SimilarityItem             `json:"similarities"`
	Recommendations []string                   `json:"recommendations"`
}

// ComparisonResult 比较结果
type ComparisonResult struct {
	AccuracyDiff      float64                `json:"accuracy_diff"`
	ConfidenceDiff    float64                `json:"confidence_diff"`
	ProcessingTimeDiff time.Duration         `json:"processing_time_diff"`
	TokenUsageDiff    int                    `json:"token_usage_diff"`
	QualityScoreDiff  float64                `json:"quality_score_diff"`
	FunctionTypeChanges map[string]int       `json:"function_type_changes"`
	ComplexityChanges map[string]int         `json:"complexity_changes"`
	OverallImprovement float64               `json:"overall_improvement"`
}

// DifferenceItem 差异项
type DifferenceItem struct {
	RequirementID uint    `json:"requirement_id"`
	Field         string  `json:"field"`
	OldValue      string  `json:"old_value"`
	NewValue      string  `json:"new_value"`
	Impact        string  `json:"impact"`
	Confidence    float64 `json:"confidence"`
}

// SimilarityItem 相似项
type SimilarityItem struct {
	RequirementID uint    `json:"requirement_id"`
	Field         string  `json:"field"`
	Value         string  `json:"value"`
	Confidence    float64 `json:"confidence"`
}

// BatchEvaluationMetricsResponse 批量评估指标响应
type BatchEvaluationMetricsResponse struct {
	Overview       MetricsOverview              `json:"overview"`
	TimeSeriesData []TimeSeriesDataPoint        `json:"time_series_data"`
	ModelMetrics   []ModelMetricsItem           `json:"model_metrics"`
	ProjectMetrics []ProjectMetricsItem         `json:"project_metrics"`
	Trends         []TrendItem                  `json:"trends"`
}

// MetricsOverview 指标概览
type MetricsOverview struct {
	TotalTasks          int           `json:"total_tasks"`
	CompletedTasks      int           `json:"completed_tasks"`
	FailedTasks         int           `json:"failed_tasks"`
	AverageAccuracy     float64       `json:"average_accuracy"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	TotalTokensUsed     int           `json:"total_tokens_used"`
	TotalCost           float64       `json:"total_cost"`
	ThroughputPerDay    int           `json:"throughput_per_day"`
}

// TimeSeriesDataPoint 时间序列数据点
type TimeSeriesDataPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	TasksCount   int       `json:"tasks_count"`
	Accuracy     float64   `json:"accuracy"`
	ProcessingTime time.Duration `json:"processing_time"`
	TokenUsage   int       `json:"token_usage"`
	Cost         float64   `json:"cost"`
}

// ModelMetricsItem 模型指标项
type ModelMetricsItem struct {
	ModelName      string        `json:"model_name"`
	TasksCount     int           `json:"tasks_count"`
	AverageAccuracy float64      `json:"average_accuracy"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	TokenUsage     int           `json:"token_usage"`
	Cost           float64       `json:"cost"`
	ErrorRate      float64       `json:"error_rate"`
}

// ProjectMetricsItem 项目指标项
type ProjectMetricsItem struct {
	ProjectID      uint          `json:"project_id"`
	ProjectName    string        `json:"project_name"`
	TasksCount     int           `json:"tasks_count"`
	AverageAccuracy float64      `json:"average_accuracy"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	TokenUsage     int           `json:"token_usage"`
	Cost           float64       `json:"cost"`
}

// TrendItem 趋势项
type TrendItem struct {
	Metric     string    `json:"metric"`
	Trend      string    `json:"trend"` // increasing/decreasing/stable
	Change     float64   `json:"change"`
	Period     string    `json:"period"`
	Confidence float64   `json:"confidence"`
}

// BatchEvaluationExportResponse 批量评估导出响应
type BatchEvaluationExportResponse struct {
	TaskID        string    `json:"task_id"`
	ExportID      string    `json:"export_id"`
	Format        string    `json:"format"`
	Status        string    `json:"status"`
	DownloadURL   string    `json:"download_url,omitempty"`
	ExpiresAt     time.Time `json:"expires_at"`
	FileSize      int64     `json:"file_size,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}

// BatchEvaluationTemplateResponse 批量评估模板响应
type BatchEvaluationTemplateResponse struct {
	ID          string                       `json:"id"`
	Name        string                       `json:"name"`
	Description string                       `json:"description"`
	Config      map[string]interface{}       `json:"config"`
	Tags        []string                     `json:"tags"`
	Public      bool                         `json:"public"`
	UsageCount  int                          `json:"usage_count"`
	CreatedBy   string                       `json:"created_by"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
	Metadata    map[string]interface{}       `json:"metadata"`
}

// BatchEvaluationScheduleResponse 批量评估调度响应
type BatchEvaluationScheduleResponse struct {
	ID              string                       `json:"id"`
	ProjectID       uint                         `json:"project_id"`
	CycleID         uint                         `json:"cycle_id"`
	SourceVersionID uint                         `json:"source_version_id"`
	Config          map[string]interface{}       `json:"config"`
	ScheduleType    string                       `json:"schedule_type"`
	ScheduleTime    string                       `json:"schedule_time"`
	Enabled         bool                         `json:"enabled"`
	NextRun         time.Time                    `json:"next_run"`
	LastRun         *time.Time                   `json:"last_run,omitempty"`
	RunCount        int                          `json:"run_count"`
	Description     string                       `json:"description"`
	CreatedAt       time.Time                    `json:"created_at"`
	UpdatedAt       time.Time                    `json:"updated_at"`
	Metadata        map[string]interface{}       `json:"metadata"`
}

// BatchEvaluationWebhookResponse Webhook响应
type BatchEvaluationWebhookResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
}

// ShardingTestResponse 分片测试响应
type ShardingTestResponse struct {
	Strategy          string    `json:"strategy"`
	TotalShards       int       `json:"total_shards"`
	TotalRequirements int       `json:"total_requirements"`
	TotalTokens       int       `json:"total_tokens"`
	AverageTokens     float64   `json:"average_tokens"`
	TokenUtilization  float64   `json:"token_utilization"`
	BalanceScore      float64   `json:"balance_score"`
	QualityScore      float64   `json:"quality_score"`
	ProcessingTime    string    `json:"processing_time"`
	OptimizationHints []string  `json:"optimization_hints"`
	Shards            []ShardInfo `json:"shards"`
}

// ShardInfo 分片信息
type ShardInfo struct {
	ShardID           string    `json:"shard_id"`
	ShardNumber       int       `json:"shard_number"`
	RequirementCount  int       `json:"requirement_count"`
	EstimatedTokens   int       `json:"estimated_tokens"`
	AverageComplexity float64   `json:"average_complexity"`
	SimilarityScore   float64   `json:"similarity_score"`
	QualityScore      float64   `json:"quality_score"`
	RecommendedModel  string    `json:"recommended_model"`
	ProcessingHints   []string  `json:"processing_hints"`
}