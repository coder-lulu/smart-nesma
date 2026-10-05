package request

// IntelligentAnalysisRequest 智能分析请求
type IntelligentAnalysisRequest struct {
	ProjectID       uint                   `json:"project_id" binding:"required"`
	CycleID         uint                   `json:"cycle_id" binding:"required"`
	RequirementIDs  []uint                 `json:"requirement_ids" binding:"required"`
	Config          map[string]interface{} `json:"config"`
	Priority        string                 `json:"priority" binding:"omitempty,oneof=low medium high"`
	AnalysisType    string                 `json:"analysis_type" binding:"omitempty,oneof=standard enhanced comprehensive"`
	GenerateContent []string               `json:"generate_content"`
}

// GetRequirementTreeRequest 获取需求树请求
type GetRequirementTreeRequest struct {
	ProjectID       uint `json:"project_id" form:"project_id"`
	CycleID         uint `json:"cycle_id" form:"cycle_id"`
	IncludeAnalyzed bool `json:"include_analyzed" form:"include_analyzed"`
	MinLevel        int  `json:"min_level" form:"min_level"`
	MaxLevel        int  `json:"max_level" form:"max_level"`
}

// ApplyRecommendationRequest 应用推荐建议请求
type ApplyRecommendationRequest struct {
	TaskID           uint                   `json:"task_id" binding:"required"`
	RecommendationID string                 `json:"recommendation_id" binding:"required"`
	RequirementID    uint                   `json:"requirement_id" binding:"required"`
	ActionType       string                 `json:"action_type" binding:"required,oneof=accept reject modify"`
	ModifiedContent  map[string]interface{} `json:"modified_content"`
	ApplyReason      string                 `json:"apply_reason"`
}

// TestAIModelRequest 测试AI模型请求
type TestAIModelRequest struct {
	ModelID     string                 `json:"model_id" binding:"required"`
	APIKey      string                 `json:"api_key"`
	BaseURL     string                 `json:"base_url"`
	TestMessage string                 `json:"test_message"`
	Config      map[string]interface{} `json:"config"`
}

// SearchKnowledgeRequest 搜索知识库请求
type SearchKnowledgeRequest struct {
	Query      string   `json:"query" binding:"required"`
	Categories []string `json:"categories"`
	Limit      int      `json:"limit"`
	Offset     int      `json:"offset"`
	Filter     string   `json:"filter"`
}

// GetKnowledgeGraphRequest 获取知识图谱请求
type GetKnowledgeGraphRequest struct {
	NodeType   string   `json:"node_type" form:"node_type"`
	Categories []string `json:"categories" form:"categories"`
	Depth      int      `json:"depth" form:"depth"`
	Limit      int      `json:"limit" form:"limit"`
	CenterNode string   `json:"center_node" form:"center_node"`
}

// VectorSearchRequest 向量搜索请求
type VectorSearchRequest struct {
	Query      string                 `json:"query" binding:"required"`
	TopK       int                    `json:"top_k"`
	Threshold  float64                `json:"threshold"`
	Filter     map[string]interface{} `json:"filter"`
	SearchType string                 `json:"search_type" binding:"omitempty,oneof=similarity hybrid semantic"`
}

// UpdateKnowledgeNodeRequest 更新知识节点请求
type UpdateKnowledgeNodeRequest struct {
	NodeID      string                 `json:"node_id" binding:"required"`
	NodeType    string                 `json:"node_type" binding:"required"`
	Properties  map[string]interface{} `json:"properties"`
	Labels      []string               `json:"labels"`
	Description string                 `json:"description"`
}

// CreateKnowledgeRelationRequest 创建知识关系请求
type CreateKnowledgeRelationRequest struct {
	FromNodeID   string                 `json:"from_node_id" binding:"required"`
	ToNodeID     string                 `json:"to_node_id" binding:"required"`
	RelationType string                 `json:"relation_type" binding:"required"`
	Properties   map[string]interface{} `json:"properties"`
	Weight       float64                `json:"weight"`
	Description  string                 `json:"description"`
}

// AnalysisConfigRequest 分析配置请求
type AnalysisConfigRequest struct {
	Mode             string   `json:"mode" binding:"required,oneof=standard enhanced comprehensive"`
	AIModel          string   `json:"ai_model" binding:"required"`
	KnowledgeWeight  int      `json:"knowledge_weight"`
	GenerateContent  []string `json:"generate_content"`
	CustomPrompt     string   `json:"custom_prompt"`
	AnalysisDepth    int      `json:"analysis_depth"`
	IncludeExamples  bool     `json:"include_examples"`
	ValidationLevel  string   `json:"validation_level" binding:"omitempty,oneof=basic standard strict"`
}

// BatchAnalysisRequest 批量分析请求
type BatchAnalysisRequest struct {
	IntelligentAnalysisRequest
	BatchSize     int                    `json:"batch_size"`
	ParallelCount int                    `json:"parallel_count"`
	RetryCount    int                    `json:"retry_count"`
	Timeout       int                    `json:"timeout"`
	ContinueOnError bool                 `json:"continue_on_error"`
}

// ExportAnalysisReportRequest 导出分析报告请求
type ExportAnalysisReportRequest struct {
	TaskID        uint     `json:"task_id" form:"task_id" binding:"required"`
	Format        string   `json:"format" form:"format" binding:"required,oneof=excel pdf word"`
	IncludeSections []string `json:"include_sections" form:"include_sections"`
	CustomTemplate string   `json:"custom_template" form:"custom_template"`
	Language      string   `json:"language" form:"language"`
}

// GetAnalysisHistoryRequest 获取分析历史请求
type GetAnalysisHistoryRequest struct {
	ProjectID   uint   `json:"project_id" form:"project_id"`
	CycleID     uint   `json:"cycle_id" form:"cycle_id"`
	UserID      uint   `json:"user_id" form:"user_id"`
	Status      string `json:"status" form:"status"`
	StartDate   string `json:"start_date" form:"start_date"`
	EndDate     string `json:"end_date" form:"end_date"`
	Page        int    `json:"page" form:"page"`
	PageSize    int    `json:"page_size" form:"page_size"`
	OrderBy     string `json:"order_by" form:"order_by"`
	OrderDir    string `json:"order_dir" form:"order_dir"`
}

// CompareAnalysisRequest 分析对比请求
type CompareAnalysisRequest struct {
	BaseTaskID    uint     `json:"base_task_id" binding:"required"`
	CompareTaskID uint     `json:"compare_task_id" binding:"required"`
	CompareFields []string `json:"compare_fields"`
	ShowDiffOnly  bool     `json:"show_diff_only"`
}

// CustomAnalysisRequest 自定义分析请求
type CustomAnalysisRequest struct {
	IntelligentAnalysisRequest
	CustomRules     []AnalysisRule   `json:"custom_rules"`
	CustomPrompts   map[string]string      `json:"custom_prompts"`
	CustomMetrics   []string               `json:"custom_metrics"`
	ExternalData    map[string]interface{} `json:"external_data"`
	ValidationRules []ValidationRule `json:"validation_rules"`
}

// AnalysisRule 分析规则
type AnalysisRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Condition   string                 `json:"condition"`
	Action      string                 `json:"action"`
	Parameters  map[string]interface{} `json:"parameters"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
}

// ValidationRule 验证规则
type ValidationRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        string                 `json:"type"`
	Rule        string                 `json:"rule"`
	ErrorMsg    string                 `json:"error_msg"`
	Severity    string                 `json:"severity"`
	Parameters  map[string]interface{} `json:"parameters"`
	Enabled     bool                   `json:"enabled"`
}

// AnalysisFilterRequest 分析过滤请求
type AnalysisFilterRequest struct {
	RequirementLevels []int    `json:"requirement_levels"`
	FunctionTypes     []string `json:"function_types"`
	Complexities      []string `json:"complexities"`
	Domains           []string `json:"domains"`
	Tags              []string `json:"tags"`
	DateRange         struct {
		Start string `json:"start"`
		End   string `json:"end"`
	} `json:"date_range"`
	ConfidenceRange struct {
		Min float64 `json:"min"`
		Max float64 `json:"max"`
	} `json:"confidence_range"`
}

// OptimizeAnalysisRequest 优化分析请求
type OptimizeAnalysisRequest struct {
	TaskID          uint                   `json:"task_id" binding:"required"`
	OptimizeFields  []string               `json:"optimize_fields"`
	OptimizeTargets map[string]interface{} `json:"optimize_targets"`
	ConstraintsConfig map[string]interface{} `json:"constraints_config"`
	IterationCount  int                    `json:"iteration_count"`
	LearningRate    float64                `json:"learning_rate"`
}

// FeedbackAnalysisRequest 分析反馈请求
type FeedbackAnalysisRequest struct {
	TaskID          uint                   `json:"task_id" binding:"required"`
	RequirementID   uint                   `json:"requirement_id" binding:"required"`
	FeedbackType    string                 `json:"feedback_type" binding:"required,oneof=positive negative suggestion"`
	FeedbackContent string                 `json:"feedback_content" binding:"required"`
	Rating          int                    `json:"rating"`
	Categories      []string               `json:"categories"`
	Suggestions     map[string]interface{} `json:"suggestions"`
}

// ScheduleAnalysisRequest 定时分析请求
type ScheduleAnalysisRequest struct {
	IntelligentAnalysisRequest
	ScheduleType   string `json:"schedule_type" binding:"required,oneof=once daily weekly monthly"`
	ScheduleTime   string `json:"schedule_time" binding:"required"`
	TimeZone       string `json:"time_zone"`
	Enabled        bool   `json:"enabled"`
	NotifyUsers    []uint `json:"notify_users"`
	NotifyChannels []string `json:"notify_channels"`
}