package request

// UnifiedAnalysisRequest 统一分析请求
type UnifiedAnalysisRequest struct {
	Type          string                 `json:"type" binding:"required,oneof=project_analysis requirement_optimize nesma_evaluation"`
	ProjectID     uint                   `json:"project_id" binding:"required"`
	CycleID       *uint                  `json:"cycle_id,omitempty"`
	VersionID     *uint                  `json:"version_id,omitempty"`
	RequirementID *uint                  `json:"requirement_id,omitempty"`
	Parameters    map[string]interface{} `json:"parameters,omitempty"`
	Priority      int                    `json:"priority" binding:"min=1,max=5"`
}

// ProjectAnalysisRequest 项目一键分析请求
type ProjectAnalysisRequest struct {
	ProjectID   uint                   `json:"project_id" binding:"required"`
	CycleID     uint                   `json:"cycle_id" binding:"required"`
	AnalysisScope string               `json:"analysis_scope" binding:"omitempty,oneof=full incremental specific"`
	IncludeLevels []int                `json:"include_levels,omitempty"`
	GenerateReports []string           `json:"generate_reports,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// RequirementOptimizeRequest 需求优化请求
type RequirementOptimizeRequest struct {
	RequirementID   uint                   `json:"requirement_id" binding:"required"`
	OptimizeType    string                 `json:"optimize_type" binding:"omitempty,oneof=title description structure quality"`
	IncludeContext  bool                   `json:"include_context"`
	KnowledgeWeight float64                `json:"knowledge_weight" binding:"min=0,max=1"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
}

// NESMAEvaluationRequest NESMA评估请求
type NESMAEvaluationRequest struct {
	ProjectID       uint                   `json:"project_id" binding:"required"`
	CycleID         uint                   `json:"cycle_id" binding:"required"`
	EvaluationType  string                 `json:"evaluation_type" binding:"omitempty,oneof=comprehensive compliance benchmark"`
	EvaluationScope string                 `json:"evaluation_scope" binding:"omitempty,oneof=all selected functional"`
	TargetRequirements []uint              `json:"target_requirements,omitempty"`
	IncludeBenchmark bool                  `json:"include_benchmark"`
	Parameters      map[string]interface{} `json:"parameters,omitempty"`
}

// GetAnalysisProgressRequest 获取分析进度请求
type GetAnalysisProgressRequest struct {
	TaskID uint `json:"task_id" form:"task_id" binding:"required"`
}

// GetAnalysisResultRequest 获取分析结果请求
type GetAnalysisResultRequest struct {
	TaskID     uint   `json:"task_id" form:"task_id" binding:"required"`
	Format     string `json:"format" form:"format" binding:"omitempty,oneof=json summary detailed"`
	IncludeRaw bool   `json:"include_raw" form:"include_raw"`
}

// CancelAnalysisRequest 取消分析请求
type CancelAnalysisRequest struct {
	TaskID uint   `json:"task_id" binding:"required"`
	Reason string `json:"reason,omitempty"`
}

// ApplyUnifiedOptimizationRequest 应用统一优化建议请求（避免与nesma_requirement.go冲突）
type ApplyUnifiedOptimizationRequest struct {
	TaskID          uint   `json:"task_id" binding:"required"`
	OptimizationID  uint   `json:"optimization_id" binding:"required"`
	SuggestionID    string `json:"suggestion_id" binding:"required"`
	Action          string `json:"action" binding:"required,oneof=accept reject modify"`
	ModifiedContent string `json:"modified_content,omitempty"`
	Reason          string `json:"reason,omitempty"`
	Modifications   map[string]interface{} `json:"modifications,omitempty"`
}

// BatchApplyUnifiedOptimizationsRequest 批量应用统一优化建议请求（避免与nesma_requirement.go冲突）
type BatchApplyUnifiedOptimizationsRequest struct {
	TaskID           uint   `json:"task_id" binding:"required"`
	OptimizationIDs  []uint `json:"optimization_ids" binding:"required,min=1"`
	Action           string `json:"action" binding:"required,oneof=accept reject modify"`
	ModifiedContent  string `json:"modified_content,omitempty"`
	Reason           string `json:"reason,omitempty"`
}

// GetOptimizationSuggestionsRequest 获取优化建议请求
type GetOptimizationSuggestionsRequest struct {
	ProjectID     *uint    `json:"project_id" form:"project_id"`
	RequirementID *uint    `json:"requirement_id" form:"requirement_id"`
	Category      string   `json:"category" form:"category"`
	Priority      string   `json:"priority" form:"priority"`
	Status        string   `json:"status" form:"status"`
	Page          int      `json:"page" form:"page"`
	PageSize      int      `json:"page_size" form:"page_size"`
}

// CreateOptimizationSuggestionRequest 创建优化建议请求
type CreateOptimizationSuggestionRequest struct {
	ProjectID           uint                   `json:"project_id" binding:"required"`
	RequirementID       *uint                  `json:"requirement_id,omitempty"`
	SuggestionType      string                 `json:"suggestion_type" binding:"required,oneof=project requirement process"`
	Category            string                 `json:"category" binding:"required,oneof=quality efficiency risk compliance"`
	Title               string                 `json:"title" binding:"required"`
	Description         string                 `json:"description" binding:"required"`
	Rationale           string                 `json:"rationale,omitempty"`
	ExpectedOutcome     string                 `json:"expected_outcome,omitempty"`
	Priority            string                 `json:"priority" binding:"required,oneof=High Medium Low"`
	Impact              string                 `json:"impact" binding:"omitempty,oneof=High Medium Low"`
	Effort              string                 `json:"effort" binding:"omitempty,oneof=High Medium Low"`
	ActionItems         []string               `json:"action_items,omitempty"`
	Timeline            string                 `json:"timeline,omitempty"`
	ResourceRequirements string                `json:"resource_requirements,omitempty"`
	GeneratedBy         string                 `json:"generated_by" binding:"omitempty,oneof=ai manual"`
	AIModel             string                 `json:"ai_model,omitempty"`
	ConfidenceScore     float64                `json:"confidence_score,omitempty"`
}

// UpdateOptimizationSuggestionRequest 更新优化建议请求
type UpdateOptimizationSuggestionRequest struct {
	Status              string     `json:"status" binding:"omitempty,oneof=pending in_progress completed rejected"`
	AssignedTo          string     `json:"assigned_to,omitempty"`
	StartDate           *string    `json:"start_date,omitempty"`
	DueDate             *string    `json:"due_date,omitempty"`
	ActualOutcome       string     `json:"actual_outcome,omitempty"`
	EffectivenessScore  *float64   `json:"effectiveness_score,omitempty"`
	ImplementationNotes string     `json:"implementation_notes,omitempty"`
}

// GetProjectAnalysisHistoryRequest 获取项目分析历史请求
type GetProjectAnalysisHistoryRequest struct {
	ProjectID     uint   `json:"project_id" form:"project_id" binding:"required"`
	CycleID       uint   `json:"cycle_id" form:"cycle_id"`
	AnalysisType  string `json:"analysis_type" form:"analysis_type"`
	StartDate     string `json:"start_date" form:"start_date"`
	EndDate       string `json:"end_date" form:"end_date"`
	Page          int    `json:"page" form:"page"`
	PageSize      int    `json:"page_size" form:"page_size"`
}

// GetNESMAEvaluationHistoryRequest 获取NESMA评估历史请求
type GetNESMAEvaluationHistoryRequest struct {
	ProjectID       uint   `json:"project_id" form:"project_id" binding:"required"`
	CycleID         uint   `json:"cycle_id" form:"cycle_id"`
	EvaluationType  string `json:"evaluation_type" form:"evaluation_type"`
	StartDate       string `json:"start_date" form:"start_date"`
	EndDate         string `json:"end_date" form:"end_date"`
	Page            int    `json:"page" form:"page"`
	PageSize        int    `json:"page_size" form:"page_size"`
}

// CompareAnalysisVersionsRequest 分析版本对比请求
type CompareAnalysisVersionsRequest struct {
	BaseTaskID     uint     `json:"base_task_id" binding:"required"`
	CompareTaskID  uint     `json:"compare_task_id" binding:"required"`
	CompareFields  []string `json:"compare_fields,omitempty"`
	ShowDiffOnly   bool     `json:"show_diff_only"`
	IncludeDetails bool     `json:"include_details"`
}

// ExportAnalysisRequest 导出分析结果请求
type ExportAnalysisRequest struct {
	TaskID         uint     `json:"task_id" form:"task_id" binding:"required"`
	ExportFormat   string   `json:"export_format" form:"export_format" binding:"required,oneof=excel pdf word json"`
	IncludeSections []string `json:"include_sections" form:"include_sections"`
	Language       string   `json:"language" form:"language" binding:"omitempty,oneof=zh-CN en-US"`
	Template       string   `json:"template" form:"template"` // 恢复为string类型
	CustomFields   []string `json:"custom_fields" form:"custom_fields"`
}

// UnifiedBatchAnalysisRequest 统一批量分析请求（避免与intelligent_analysis.go冲突）
type UnifiedBatchAnalysisRequest struct {
	ProjectID      uint                   `json:"project_id" binding:"required"`
	CycleID        uint                   `json:"cycle_id" binding:"required"`
	AnalysisTypes  []string               `json:"analysis_types" binding:"required"`
	BatchSize      int                    `json:"batch_size" binding:"min=1,max=100"`
	ParallelCount  int                    `json:"parallel_count" binding:"min=1,max=10"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
}

// UnifiedScheduleAnalysisRequest 统一定时分析请求（避免与intelligent_analysis.go冲突）
type UnifiedScheduleAnalysisRequest struct {
	ProjectID      uint                   `json:"project_id" binding:"required"`
	CycleID        uint                   `json:"cycle_id" binding:"required"`
	AnalysisType   string                 `json:"analysis_type" binding:"required"`
	ScheduleType   string                 `json:"schedule_type" binding:"required,oneof=once daily weekly monthly"`
	ScheduleTime   string                 `json:"schedule_time" binding:"required"`
	TimeZone       string                 `json:"time_zone"`
	Enabled        bool                   `json:"enabled"`
	Parameters     map[string]interface{} `json:"parameters,omitempty"`
	NotifyUsers    []uint                 `json:"notify_users,omitempty"`
	NotifyChannels []string               `json:"notify_channels,omitempty"`
}

// AnalysisQualityFeedbackRequest 分析质量反馈请求
type AnalysisQualityFeedbackRequest struct {
	TaskID          uint   `json:"task_id" binding:"required"`
	RequirementID   uint   `json:"requirement_id,omitempty"`
	OverallRating   int    `json:"overall_rating" binding:"min=1,max=5"`
	AccuracyRating  int    `json:"accuracy_rating" binding:"min=1,max=5"`
	UsefulnessRating int   `json:"usefulness_rating" binding:"min=1,max=5"`
	FeedbackText    string `json:"feedback_text"`
	ImprovementSuggestions []string `json:"improvement_suggestions,omitempty"`
}