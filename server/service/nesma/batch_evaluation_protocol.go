package nesma

import (
	"encoding/json"
	"fmt"
	"time"
)

// BatchEvaluationProtocol 批量评估通信协议
type BatchEvaluationProtocol struct {
	Version   string                 `json:"version"`
	Protocol  string                 `json:"protocol"`
	Timestamp time.Time              `json:"timestamp"`
	RequestID string                 `json:"request_id"`
	Data      map[string]interface{} `json:"data"`
}

// ProtocolMessage 协议消息基类
type ProtocolMessage struct {
	MessageType string                 `json:"message_type"`
	MessageID   string                 `json:"message_id"`
	Timestamp   time.Time              `json:"timestamp"`
	Source      string                 `json:"source"`
	Target      string                 `json:"target"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]string      `json:"metadata,omitempty"`
}

// BatchEvaluationRequest 批量评估请求协议
type BatchEvaluationRequest struct {
	ProtocolMessage
	TaskID          string                   `json:"task_id"`
	ProjectID       uint                     `json:"project_id"`
	CycleID         uint                     `json:"cycle_id"`
	SourceVersionID uint                     `json:"source_version_id"`
	Config          *BatchEvaluationConfig   `json:"config"`
	Requirements    []RequirementData        `json:"requirements"`
	Context         *EvaluationContext       `json:"context,omitempty"`
	Priority        int                      `json:"priority"`
	Callback        *CallbackConfig          `json:"callback,omitempty"`
}

// BatchEvaluationResponse 批量评估响应协议
type BatchEvaluationResponse struct {
	ProtocolMessage
	TaskID         string                    `json:"task_id"`
	Status         string                    `json:"status"`
	Progress       float64                   `json:"progress"`
	Results        []EvaluationResult        `json:"results"`
	Summary        *EvaluationSummaryProtocol        `json:"summary,omitempty"`
	QualityMetrics *QualityMetrics           `json:"quality_metrics,omitempty"`
	Error          *ErrorInfo                `json:"error,omitempty"`
	TokenUsage     *TokenUsageInfo           `json:"token_usage,omitempty"`
	ProcessingTime time.Duration             `json:"processing_time"`
}

// BatchRequest 批次请求协议
type BatchRequest struct {
	ProtocolMessage
	TaskID       string            `json:"task_id"`
	BatchID      string            `json:"batch_id"`
	BatchNumber  int               `json:"batch_number"`
	Requirements []RequirementData `json:"requirements"`
	Context      *BatchContextProtocol     `json:"context"`
	Config       *BatchConfig      `json:"config"`
	Priority     int               `json:"priority"`
}

// BatchResponse 批次响应协议
type BatchResponse struct {
	ProtocolMessage
	TaskID         string             `json:"task_id"`
	BatchID        string             `json:"batch_id"`
	BatchNumber    int                `json:"batch_number"`
	Status         string             `json:"status"`
	Results        []EvaluationResult `json:"results"`
	TokenUsage     *TokenUsageInfo    `json:"token_usage"`
	ProcessingTime time.Duration      `json:"processing_time"`
	QualityScore   float64            `json:"quality_score"`
	Error          *ErrorInfo         `json:"error,omitempty"`
}

// ProgressUpdate 进度更新协议
type ProgressUpdate struct {
	ProtocolMessage
	TaskID         string            `json:"task_id"`
	BatchID        string            `json:"batch_id,omitempty"`
	Progress       float64           `json:"progress"`
	CurrentBatch   int               `json:"current_batch"`
	TotalBatches   int               `json:"total_batches"`
	ProcessedCount int               `json:"processed_count"`
	TotalCount     int               `json:"total_count"`
	Status         string            `json:"status"`
	Message        string            `json:"message"`
	EstimatedTime  *time.Time        `json:"estimated_time,omitempty"`
	Statistics     *ProgressStats    `json:"statistics,omitempty"`
}

// ProgressStats 进度统计
type ProgressStats struct {
	ProcessingRate     float64       `json:"processing_rate"`     // 处理速率 (requirements/minute)
	AverageQuality     float64       `json:"average_quality"`     // 平均质量分数
	TokensPerSecond    float64       `json:"tokens_per_second"`   // Token处理速率
	EstimatedRemaining time.Duration `json:"estimated_remaining"` // 预计剩余时间
	ThroughputTrend    string        `json:"throughput_trend"`    // 吞吐量趋势 (increasing/stable/decreasing)
}

// StatusUpdate 状态更新协议
type StatusUpdate struct {
	ProtocolMessage
	TaskID       string                 `json:"task_id"`
	OldStatus    string                 `json:"old_status"`
	NewStatus    string                 `json:"new_status"`
	Reason       string                 `json:"reason"`
	Details      map[string]interface{} `json:"details,omitempty"`
	ActionNeeded bool                   `json:"action_needed"`
}

// ErrorReport 错误报告协议
type ErrorReport struct {
	ProtocolMessage
	TaskID       string            `json:"task_id"`
	BatchID      string            `json:"batch_id,omitempty"`
	ErrorType    string            `json:"error_type"`
	ErrorCode    string            `json:"error_code"`
	ErrorMessage string            `json:"error_message"`
	Details      map[string]interface{} `json:"details,omitempty"`
	Severity     string            `json:"severity"` // critical/error/warning/info
	Recoverable  bool              `json:"recoverable"`
	Suggestions  []string          `json:"suggestions,omitempty"`
	StackTrace   string            `json:"stack_trace,omitempty"`
}

// CallbackConfig 回调配置
type CallbackConfig struct {
	URL           string            `json:"url"`
	Method        string            `json:"method"`
	Headers       map[string]string `json:"headers,omitempty"`
	RetryPolicy   *RetryPolicyProtocol      `json:"retry_policy,omitempty"`
	Authentication *AuthConfig      `json:"authentication,omitempty"`
	Events        []string          `json:"events"` // 订阅的事件类型
}

// RetryPolicyProtocol 重试策略协议
type RetryPolicyProtocol struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type   string            `json:"type"`   // bearer/basic/apikey
	Token  string            `json:"token,omitempty"`
	ApiKey string            `json:"api_key,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

// RequirementData 需求数据
type RequirementData struct {
	ID               uint                   `json:"id"`
	Title            string                 `json:"title"`
	Description      string                 `json:"description"`
	Level            int                    `json:"level"`
	ParentID         *uint                  `json:"parent_id,omitempty"`
	BusinessContext  string                 `json:"business_context,omitempty"`
	TechnicalContext string                 `json:"technical_context,omitempty"`
	Constraints      []string               `json:"constraints,omitempty"`
	Metadata         map[string]interface{} `json:"metadata,omitempty"`
}

// BatchContextProtocol 批次上下文协议
type BatchContextProtocol struct {
	EvaluationID       string                 `json:"evaluation_id"`
	ProjectInfo        *ProjectInfoProtocol           `json:"project_info"`
	PreviousResults    []EvaluationResult     `json:"previous_results"`
	KnowledgeRefs      []KnowledgeReferenceProtocol   `json:"knowledge_refs"`
	QualityConstraints *QualityConstraints    `json:"quality_constraints"`
	ProcessingHints    []string               `json:"processing_hints"`
	CustomContext      map[string]interface{} `json:"custom_context,omitempty"`
}

// ProjectInfoProtocol 项目信息协议
type ProjectInfoProtocol struct {
	Name        string            `json:"name"`
	Domain      string            `json:"domain"`
	Description string            `json:"description"`
	TechStack   []string          `json:"tech_stack"`
	Standards   []string          `json:"standards"`
	Constraints map[string]string `json:"constraints"`
}

// KnowledgeReferenceProtocol 知识引用协议
type KnowledgeReferenceProtocol struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Content     string  `json:"content"`
	Type        string  `json:"type"`
	Relevance   float64 `json:"relevance"`
	Category    string  `json:"category"`
	Tags        []string `json:"tags"`
}

// QualityConstraints 质量约束
type QualityConstraints struct {
	MinConfidence       float64 `json:"min_confidence"`
	MaxUncertainty      float64 `json:"max_uncertainty"`
	RequiredConsistency float64 `json:"required_consistency"`
	MinCoverage         float64 `json:"min_coverage"`
	QualityGates        []QualityGate `json:"quality_gates"`
}

// QualityGate 质量门禁
type QualityGate struct {
	Name        string  `json:"name"`
	Metric      string  `json:"metric"`
	Threshold   float64 `json:"threshold"`
	Operator    string  `json:"operator"` // gt/gte/lt/lte/eq/ne
	Mandatory   bool    `json:"mandatory"`
	Description string  `json:"description"`
}

// BatchConfig 批次配置
type BatchConfig struct {
	MaxTokens           int                    `json:"max_tokens"`
	Model               string                 `json:"model"`
	Temperature         float64                `json:"temperature"`
	TopP                float64                `json:"top_p"`
	FrequencyPenalty    float64                `json:"frequency_penalty"`
	PresencePenalty     float64                `json:"presence_penalty"`
	Stop                []string               `json:"stop,omitempty"`
	CustomPrompts       map[string]string      `json:"custom_prompts,omitempty"`
	FeatureFlags        map[string]bool        `json:"feature_flags,omitempty"`
	ProcessingOptions   map[string]interface{} `json:"processing_options,omitempty"`
}

// EvaluationSummaryProtocol 评估摘要协议
type EvaluationSummaryProtocol struct {
	TotalRequirements   int                    `json:"total_requirements"`
	ProcessedRequirements int                  `json:"processed_requirements"`
	SuccessRate         float64                `json:"success_rate"`
	AverageConfidence   float64                `json:"average_confidence"`
	AverageQuality      float64                `json:"average_quality"`
	FunctionTypeStats   map[string]int         `json:"function_type_stats"`
	ComplexityStats     map[string]int         `json:"complexity_stats"`
	TotalAFP            float64                `json:"total_afp"`
	TotalUFP            float64                `json:"total_ufp"`
	QualityDistribution map[string]int         `json:"quality_distribution"`
	ProcessingStats     *ProcessingStatistics  `json:"processing_stats"`
	Recommendations     []string               `json:"recommendations"`
}

// ProcessingStatistics 处理统计
type ProcessingStatistics struct {
	TotalBatches        int           `json:"total_batches"`
	CompletedBatches    int           `json:"completed_batches"`
	FailedBatches       int           `json:"failed_batches"`
	AverageBatchTime    time.Duration `json:"average_batch_time"`
	TotalProcessingTime time.Duration `json:"total_processing_time"`
	ThroughputPerHour   int           `json:"throughput_per_hour"`
	PeakThroughput      int           `json:"peak_throughput"`
	BottleneckAnalysis  []string      `json:"bottleneck_analysis"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code        string                 `json:"code"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Recoverable bool                   `json:"recoverable"`
	Suggestions []string               `json:"suggestions,omitempty"`
}

// WebSocketProtocol WebSocket协议定义
type WebSocketProtocol struct {
	// 消息类型
	MessageTypes struct {
		Connection string `json:"connection"`
		Subscribe  string `json:"subscribe"`
		Unsubscribe string `json:"unsubscribe"`
		Status     string `json:"status"`
		Update     string `json:"update"`
		Error      string `json:"error"`
		Ping       string `json:"ping"`
		Pong       string `json:"pong"`
	} `json:"message_types"`
	
	// 事件类型
	EventTypes struct {
		Connected       string `json:"connected"`
		Disconnected    string `json:"disconnected"`
		TaskCreated     string `json:"task_created"`
		TaskStarted     string `json:"task_started"`
		TaskCompleted   string `json:"task_completed"`
		TaskFailed      string `json:"task_failed"`
		TaskCancelled   string `json:"task_cancelled"`
		BatchStarted    string `json:"batch_started"`
		BatchCompleted  string `json:"batch_completed"`
		BatchFailed     string `json:"batch_failed"`
		ProgressUpdate  string `json:"progress_update"`
		StatusUpdate    string `json:"status_update"`
		ErrorOccurred   string `json:"error_occurred"`
	} `json:"event_types"`
	
	// 状态定义
	StatusTypes struct {
		Pending     string `json:"pending"`
		Running     string `json:"running"`
		Completed   string `json:"completed"`
		Failed      string `json:"failed"`
		Cancelled   string `json:"cancelled"`
		Paused      string `json:"paused"`
		Retrying    string `json:"retrying"`
	} `json:"status_types"`
}

// ProtocolValidator 协议验证器
type ProtocolValidator struct {
	version string
	schemas map[string]interface{}
}

// NewProtocolValidator 创建协议验证器
func NewProtocolValidator(version string) *ProtocolValidator {
	return &ProtocolValidator{
		version: version,
		schemas: make(map[string]interface{}),
	}
}

// ValidateMessage 验证消息格式
func (v *ProtocolValidator) ValidateMessage(messageType string, data []byte) error {
	// 实现消息验证逻辑
	var message ProtocolMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return err
	}
	
	// 检查必要字段
	if message.MessageType == "" {
		return fmt.Errorf("message_type is required")
	}
	
	if message.MessageID == "" {
		return fmt.Errorf("message_id is required")
	}
	
	if message.Timestamp.IsZero() {
		return fmt.Errorf("timestamp is required")
	}
	
	// 根据消息类型进行特定验证
	switch messageType {
	case "batch_evaluation_request":
		return v.validateBatchEvaluationRequest(data)
	case "batch_request":
		return v.validateBatchRequest(data)
	case "progress_update":
		return v.validateProgressUpdate(data)
	// 添加更多消息类型验证
	}
	
	return nil
}

// validateBatchEvaluationRequest 验证批量评估请求
func (v *ProtocolValidator) validateBatchEvaluationRequest(data []byte) error {
	var req BatchEvaluationRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	
	if req.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	
	if req.ProjectID == 0 {
		return fmt.Errorf("project_id is required")
	}
	
	if req.CycleID == 0 {
		return fmt.Errorf("cycle_id is required")
	}
	
	if req.Config == nil {
		return fmt.Errorf("config is required")
	}
	
	return nil
}

// validateBatchRequest 验证批次请求
func (v *ProtocolValidator) validateBatchRequest(data []byte) error {
	var req BatchRequest
	if err := json.Unmarshal(data, &req); err != nil {
		return err
	}
	
	if req.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	
	if req.BatchID == "" {
		return fmt.Errorf("batch_id is required")
	}
	
	if len(req.Requirements) == 0 {
		return fmt.Errorf("requirements are required")
	}
	
	return nil
}

// validateProgressUpdate 验证进度更新
func (v *ProtocolValidator) validateProgressUpdate(data []byte) error {
	var update ProgressUpdate
	if err := json.Unmarshal(data, &update); err != nil {
		return err
	}
	
	if update.TaskID == "" {
		return fmt.Errorf("task_id is required")
	}
	
	if update.Progress < 0 || update.Progress > 1 {
		return fmt.Errorf("progress must be between 0 and 1")
	}
	
	return nil
}

// ProtocolEncoder 协议编码器
type ProtocolEncoder struct {
	version string
}

// NewProtocolEncoder 创建协议编码器
func NewProtocolEncoder(version string) *ProtocolEncoder {
	return &ProtocolEncoder{version: version}
}

// EncodeMessage 编码消息
func (e *ProtocolEncoder) EncodeMessage(message interface{}) ([]byte, error) {
	return json.Marshal(message)
}

// DecodeMessage 解码消息
func (e *ProtocolEncoder) DecodeMessage(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}

// 协议常量定义
const (
	// 协议版本
	ProtocolVersion = "1.0"
	
	// 消息类型
	MessageTypeBatchEvaluationRequest  = "batch_evaluation_request"
	MessageTypeBatchEvaluationResponse = "batch_evaluation_response"
	MessageTypeBatchRequest            = "batch_request"
	MessageTypeBatchResponse           = "batch_response"
	MessageTypeProgressUpdate          = "progress_update"
	MessageTypeStatusUpdate            = "status_update"
	MessageTypeErrorReport             = "error_report"
	
	// 事件类型
	EventTypeTaskCreated    = "task_created"
	EventTypeTaskStarted    = "task_started"
	EventTypeTaskCompleted  = "task_completed"
	EventTypeTaskFailed     = "task_failed"
	EventTypeTaskCancelled  = "task_cancelled"
	EventTypeBatchStarted   = "batch_started"
	EventTypeBatchCompleted = "batch_completed"
	EventTypeBatchFailed    = "batch_failed"
	EventTypeProgressUpdate = "progress_update"
	EventTypeStatusUpdate   = "status_update"
	EventTypeErrorOccurred  = "error_occurred"
	
	// 状态类型
	StatusTypePending   = "pending"
	StatusTypeRunning   = "running"
	StatusTypeCompleted = "completed"
	StatusTypeFailed    = "failed"
	StatusTypeCancelled = "cancelled"
	StatusTypePaused    = "paused"
	StatusTypeRetrying  = "retrying"
	
	// 错误类型
	ErrorTypeValidation    = "validation_error"
	ErrorTypeProcessing    = "processing_error"
	ErrorTypeTimeout       = "timeout_error"
	ErrorTypeRateLimit     = "rate_limit_error"
	ErrorTypeAuthentication = "authentication_error"
	ErrorTypeAuthorization = "authorization_error"
	ErrorTypeSystemError   = "system_error"
	ErrorTypeNetworkError  = "network_error"
)