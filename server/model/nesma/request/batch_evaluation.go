package request

import (
	"fmt"
	"time"
)

// CreateBatchEvaluationRequest 创建批量评估请求
type CreateBatchEvaluationRequest struct {
	ProjectID       uint                              `json:"project_id" binding:"required" validate:"min=1"`
	CycleID         uint                              `json:"cycle_id" binding:"required" validate:"min=1"`
	SourceVersionID uint                              `json:"source_version_id" binding:"required" validate:"min=1"`
	Config          *BatchEvaluationConfigRequest     `json:"config" binding:"required"`
	Description     string                            `json:"description,omitempty"`
	Priority        string                            `json:"priority,omitempty" validate:"oneof=low medium high"`
	ScheduledTime   *time.Time                        `json:"scheduled_time,omitempty"`
	Callback        *BatchEvaluationCallbackConfig    `json:"callback,omitempty"`
	Metadata        map[string]interface{}           `json:"metadata,omitempty"`
}

// BatchEvaluationCallbackConfig 批量评估回调配置
type BatchEvaluationCallbackConfig struct {
	WebhookURL      string            `json:"webhook_url,omitempty"`
	OnProgress      bool              `json:"on_progress,omitempty"`
	OnComplete      bool              `json:"on_complete,omitempty"`
	OnError         bool              `json:"on_error,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	RetryCount      int               `json:"retry_count,omitempty"`
	RetryInterval   time.Duration     `json:"retry_interval,omitempty"`
}

// Validate 验证请求参数
func (r *CreateBatchEvaluationRequest) Validate() error {
	if r.ProjectID == 0 {
		return fmt.Errorf("项目ID不能为空")
	}
	if r.CycleID == 0 {
		return fmt.Errorf("周期ID不能为空")
	}
	if r.SourceVersionID == 0 {
		return fmt.Errorf("源版本ID不能为空")
	}
	if r.Config == nil {
		return fmt.Errorf("配置不能为空")
	}
	if err := r.Config.Validate(); err != nil {
		return fmt.Errorf("配置验证失败: %v", err)
	}
	return nil
}

// RetryBatchEvaluationRequest 重试批量评估请求
type RetryBatchEvaluationRequest struct {
	BatchIDs []string                     `json:"batch_ids,omitempty"`
	Config   *BatchEvaluationConfigRequest `json:"config,omitempty"`
	Reason   string                       `json:"reason,omitempty"`
}

// BatchEvaluationQuery 批量评估查询参数
type BatchEvaluationQuery struct {
	ProjectID    uint   `form:"project_id"`
	CycleID      uint   `form:"cycle_id"`
	Status       string `form:"status"`
	StartTime    string `form:"start_time"`
	EndTime      string `form:"end_time"`
	Page         int    `form:"page"`
	PageSize     int    `form:"page_size"`
	SortBy       string `form:"sort_by"`
	SortOrder    string `form:"sort_order"`
	SearchKeyword string `form:"search_keyword"`
}

// BatchEvaluationResultQuery 批量评估结果查询参数
type BatchEvaluationResultQuery struct {
	TaskID         string `form:"task_id" binding:"required"`
	BatchID        string `form:"batch_id"`
	RequirementID  uint   `form:"requirement_id"`
	FunctionType   string `form:"function_type"`
	ComplexityLevel string `form:"complexity_level"`
	MinConfidence  float64 `form:"min_confidence"`
	MaxConfidence  float64 `form:"max_confidence"`
	Page           int    `form:"page"`
	PageSize       int    `form:"page_size"`
	SortBy         string `form:"sort_by"`
	SortOrder      string `form:"sort_order"`
}

// BatchEvaluationExportRequest 批量评估导出请求
type BatchEvaluationExportRequest struct {
	TaskID      string   `json:"task_id" binding:"required"`
	Format      string   `json:"format" binding:"required,oneof=excel word pdf"`
	Sections    []string `json:"sections,omitempty"`
	Template    string   `json:"template,omitempty"`
	Options     *ExportOptions `json:"options,omitempty"`
}

// ExportOptions 导出选项
type ExportOptions struct {
	IncludeCharts     bool   `json:"include_charts,omitempty"`
	IncludeDetails    bool   `json:"include_details,omitempty"`
	IncludeMetrics    bool   `json:"include_metrics,omitempty"`
	Language          string `json:"language,omitempty"`
	Theme             string `json:"theme,omitempty"`
	CustomTitle       string `json:"custom_title,omitempty"`
	ShowBatchDetails  bool   `json:"show_batch_details,omitempty"`
	ShowTokenUsage    bool   `json:"show_token_usage,omitempty"`
}

// BatchEvaluationCompareRequest 批量评估比较请求
type BatchEvaluationCompareRequest struct {
	TaskID1      string   `json:"task_id_1" binding:"required"`
	TaskID2      string   `json:"task_id_2" binding:"required"`
	CompareBy    []string `json:"compare_by,omitempty"`
	GroupBy      string   `json:"group_by,omitempty"`
	ShowDetails  bool     `json:"show_details,omitempty"`
}

// BatchEvaluationNotificationRequest 批量评估通知请求
type BatchEvaluationNotificationRequest struct {
	TaskID       string                 `json:"task_id" binding:"required"`
	NotifyType   string                 `json:"notify_type" binding:"required,oneof=email sms webhook"`
	Recipients   []string               `json:"recipients" binding:"required"`
	Template     string                 `json:"template,omitempty"`
	CustomFields map[string]interface{} `json:"custom_fields,omitempty"`
}

// BatchEvaluationScheduleRequest 批量评估调度请求
type BatchEvaluationScheduleRequest struct {
	ProjectID       uint                         `json:"project_id" binding:"required"`
	CycleID         uint                         `json:"cycle_id" binding:"required"`
	SourceVersionID uint                         `json:"source_version_id" binding:"required"`
	Config          *BatchEvaluationConfigRequest `json:"config" binding:"required"`
	ScheduleType    string                       `json:"schedule_type" binding:"required,oneof=once daily weekly monthly"`
	ScheduleTime    string                       `json:"schedule_time" binding:"required"`
	Enabled         bool                         `json:"enabled"`
	Description     string                       `json:"description,omitempty"`
	Metadata        map[string]interface{}       `json:"metadata,omitempty"`
}

// BatchEvaluationConfigRequest 批量评估配置请求
type BatchEvaluationConfigRequest struct {
	AIModel                string                 `json:"ai_model" binding:"required"`
	MaxTokensPerBatch      int                    `json:"max_tokens_per_batch" binding:"required,min=1000,max=200000"`
	BatchSize              int                    `json:"batch_size" binding:"required,min=1,max=100"`
	ConcurrentBatches      int                    `json:"concurrent_batches" binding:"required,min=1,max=10"`
	MaxRetries             int                    `json:"max_retries" binding:"min=0,max=5"`
	RetryDelay             int                    `json:"retry_delay" binding:"min=1,max=300"`
	EnableFallback         bool                   `json:"enable_fallback"`
	FallbackModels         []string               `json:"fallback_models,omitempty"`
	QualityThreshold       float64                `json:"quality_threshold" binding:"min=0.0,max=1.0"`
	ConfidenceThreshold    float64                `json:"confidence_threshold" binding:"min=0.0,max=1.0"`
	EnableContextTracking  bool                   `json:"enable_context_tracking"`
	ContextWindow          int                    `json:"context_window" binding:"min=1000,max=200000"`
	EnableProgressCallback bool                   `json:"enable_progress_callback"`
	CallbackInterval       int                    `json:"callback_interval" binding:"min=1,max=60"`
	CustomPrompts          map[string]string      `json:"custom_prompts,omitempty"`
	FeatureFlags           map[string]bool        `json:"feature_flags,omitempty"`
	Metadata               map[string]interface{} `json:"metadata,omitempty"`
}

// Validate 验证配置
func (r *BatchEvaluationConfigRequest) Validate() error {
	if r.AIModel == "" {
		return fmt.Errorf("AI模型不能为空")
	}
	if r.MaxTokensPerBatch < 1000 || r.MaxTokensPerBatch > 200000 {
		return fmt.Errorf("每批最大Token数应在1000-200000之间")
	}
	if r.BatchSize < 1 || r.BatchSize > 100 {
		return fmt.Errorf("批次大小应在1-100之间")
	}
	if r.ConcurrentBatches < 1 || r.ConcurrentBatches > 10 {
		return fmt.Errorf("并发批次数应在1-10之间")
	}
	if r.MaxRetries < 0 || r.MaxRetries > 5 {
		return fmt.Errorf("最大重试次数应在0-5之间")
	}
	if r.QualityThreshold < 0.0 || r.QualityThreshold > 1.0 {
		return fmt.Errorf("质量阈值应在0.0-1.0之间")
	}
	if r.ConfidenceThreshold < 0.0 || r.ConfidenceThreshold > 1.0 {
		return fmt.Errorf("置信度阈值应在0.0-1.0之间")
	}
	return nil
}

// ToServiceConfig 转换为服务配置（需要在API层实现转换）
func (r *BatchEvaluationConfigRequest) ToServiceConfig() map[string]interface{} {
	return map[string]interface{}{
		"ai_model":                r.AIModel,
		"max_tokens_per_batch":    r.MaxTokensPerBatch,
		"batch_size":              r.BatchSize,
		"concurrent_batches":      r.ConcurrentBatches,
		"max_retries":             r.MaxRetries,
		"retry_delay":             time.Duration(r.RetryDelay) * time.Second,
		"enable_fallback":         r.EnableFallback,
		"fallback_models":         r.FallbackModels,
		"quality_threshold":       r.QualityThreshold,
		"confidence_threshold":    r.ConfidenceThreshold,
		"enable_context_tracking": r.EnableContextTracking,
		"context_window":          r.ContextWindow,
		"enable_progress_callback": r.EnableProgressCallback,
		"callback_interval":       time.Duration(r.CallbackInterval) * time.Second,
		"custom_prompts":          r.CustomPrompts,
		"feature_flags":           r.FeatureFlags,
		"metadata":                r.Metadata,
	}
}

// BatchEvaluationWebhookRequest Webhook请求
type BatchEvaluationWebhookRequest struct {
	TaskID     string                 `json:"task_id"`
	EventType  string                 `json:"event_type"`
	Timestamp  time.Time              `json:"timestamp"`
	Status     string                 `json:"status"`
	Progress   float64                `json:"progress"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data,omitempty"`
	Signature  string                 `json:"signature,omitempty"`
}

// BatchEvaluationMetricsRequest 批量评估指标请求
type BatchEvaluationMetricsRequest struct {
	TaskIDs    []string  `json:"task_ids,omitempty"`
	ProjectID  uint      `json:"project_id,omitempty"`
	CycleID    uint      `json:"cycle_id,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	MetricType string    `json:"metric_type,omitempty"`
	GroupBy    string    `json:"group_by,omitempty"`
}

// BatchEvaluationTemplateRequest 批量评估模板请求
type BatchEvaluationTemplateRequest struct {
	Name        string                       `json:"name" binding:"required"`
	Description string                       `json:"description,omitempty"`
	Config      *BatchEvaluationConfigRequest `json:"config" binding:"required"`
	Tags        []string                     `json:"tags,omitempty"`
	Public      bool                         `json:"public"`
	Metadata    map[string]interface{}       `json:"metadata,omitempty"`
}

// TestShardingRequest 测试分片请求
type TestShardingRequest struct {
	ProjectID       uint                   `json:"project_id" binding:"required"`
	CycleID         uint                   `json:"cycle_id" binding:"required"`
	SourceVersionID uint                   `json:"source_version_id"`
	Strategy        string                 `json:"strategy" binding:"required,oneof=optimal balanced token complexity similarity sequential"`
	Config          map[string]interface{} `json:"config,omitempty"`
}