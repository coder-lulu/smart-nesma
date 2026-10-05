package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"gorm.io/datatypes"
)

// NesmaRequirementSearch 需求搜索请求
type NesmaRequirementSearch struct {
	request.PageInfo
	ProjectID   *uint  `json:"projectId" form:"projectId"`
	ParentID    *uint  `json:"parentId" form:"parentId"`
	Level       *int   `json:"level" form:"level"`
	Status      string `json:"status" form:"status"`
	Keyword     string `json:"keyword" form:"keyword"`
	Title       string `json:"title" form:"title"`
	Description string `json:"description" form:"description"`
	VesionId   *uint  `json:"versionId" form:"versionId"`
}

// NesmaRequirementTreeSearch 需求树形搜索请求
type NesmaRequirementTreeSearch struct {
	ProjectID *uint  `json:"projectId" form:"projectId"`
	Level     *int   `json:"level" form:"level"`
	Status    string `json:"status" form:"status"`
	Keyword   string `json:"keyword" form:"keyword"`
}

// CreateNesmaRequirementRequest 创建需求请求
type CreateNesmaRequirementRequest struct {
	ProjectID          uint           `json:"projectId" binding:"required"`
	ParentID           *uint          `json:"parentId"`
	Level              int            `json:"level" binding:"required,min=1,max=4"`
	Code               string         `json:"code"`
	Title              string         `json:"title" binding:"required"`
	Description        string         `json:"description"`
	Priority           int            `json:"priority" binding:"min=1,max=5"`
	Status             string         `json:"status"`
	OrderIndex         int            `json:"orderIndex"`
	DomainTags         datatypes.JSON `json:"domainTags"`
	Category           string         `json:"category"`
	Complexity         string         `json:"complexity"`
	EstimateHours      *float64       `json:"estimateHours"`
	ActualHours        *float64       `json:"actualHours"`
	BusinessValue      string         `json:"businessValue"`
	AcceptanceCriteria string         `json:"acceptanceCriteria"`
	Notes              string         `json:"notes"`

	// 新增字段
	CycleID          uint    `json:"cycleId"`
	VersionID        uint    `json:"versionId"`
	AFP              float64 `json:"afp"`
	UFP              float64 `json:"ufp"`
	FunctionType     string  `json:"functionType"`
	ReuseLevel       string  `json:"reuseLevel"`
	ModificationType string  `json:"modificationType"`
}

// UpdateNesmaRequirementRequest 更新需求请求
type UpdateNesmaRequirementRequest struct {
	ID                 uint           `json:"id" binding:"required"`
	ProjectID          uint           `json:"projectId" binding:"required"`
	ParentID           *uint          `json:"parentId"`
	Level              int            `json:"level" binding:"required,min=1,max=4"`
	Code               string         `json:"code"`
	Title              string         `json:"title" binding:"required"`
	Description        string         `json:"description"`
	Priority           int            `json:"priority" binding:"min=1,max=5"`
	Status             string         `json:"status"`
	OrderIndex         int            `json:"orderIndex"`
	DomainTags         datatypes.JSON `json:"domainTags"`
	Category           string         `json:"category"`
	Complexity         string         `json:"complexity"`
	EstimateHours      *float64       `json:"estimateHours"`
	ActualHours        *float64       `json:"actualHours"`
	BusinessValue      string         `json:"businessValue"`
	AcceptanceCriteria string         `json:"acceptanceCriteria"`
	Notes              string         `json:"notes"`

	// 新增字段
	CycleID          uint    `json:"cycleId"`
	VersionID        uint    `json:"versionId"`
	AFP              float64 `json:"afp"`
	UFP              float64 `json:"ufp"`
	FunctionType     string  `json:"functionType"`
	ReuseLevel       string  `json:"reuseLevel"`
	ModificationType string  `json:"modificationType"`
}

// NesmaRequirementIdsRequest 批量操作请求
type NesmaRequirementIdsRequest struct {
	IDs []uint `json:"ids" binding:"required"`
}

// ImportNesmaRequirementRequest Excel导入请求
type ImportNesmaRequirementRequest struct {
	ProjectID     uint   `json:"projectId" binding:"required"`
	CycleID       uint   `json:"cycleId" binding:"required"`
	VersionID     uint   `json:"versionId" binding:"required"`
	VersionName   string `json:"versionName"`  // 版本名称
	ImportBatch   string `json:"importBatch"`
	ImportSource  string `json:"importSource"`
	VersionAction string `json:"versionAction"`
}

// ExcelRequirementRow Excel中的需求行数据
type ExcelRequirementRow struct {
	Level              int      `json:"level"`
	Code               string   `json:"code"`
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	Priority           int      `json:"priority"`
	Category           string   `json:"category"`
	Complexity         string   `json:"complexity"`
	EstimateHours      *float64 `json:"estimateHours"`
	BusinessValue      string   `json:"businessValue"`
	AcceptanceCriteria string   `json:"acceptanceCriteria"`
	Notes              string   `json:"notes"`
	ParentTitle        string   `json:"parentTitle"` // 用于匹配父级

	// 新增字段
	ConstructionPeriod string  `json:"constructionPeriod"`
	Version            int     `json:"version"`
	AFP                float64 `json:"afp"`
	UFP                float64 `json:"ufp"`
	FunctionType       string  `json:"functionType"`
	ReuseLevel         string  `json:"reuseLevel"`
	ModificationType   string  `json:"modificationType"`
}

// BatchUpdateOrderRequest 批量更新排序请求
type BatchUpdateOrderRequest struct {
	Items []struct {
		ID         uint `json:"id" binding:"required"`
		OrderIndex int  `json:"orderIndex"`
	} `json:"items" binding:"required"`
}

// MoveRequirementRequest 移动需求请求
type MoveRequirementRequest struct {
	ID       uint  `json:"id" binding:"required"`
	ParentID *uint `json:"parentId"`
	Position int   `json:"position"`
}

// DeleteRequirementsByConditionRequest 根据条件删除需求请求
type DeleteRequirementsByConditionRequest struct {
	ProjectID uint `json:"projectId" binding:"required"`
	CycleID   uint `json:"cycleId" binding:"required"`
	VersionID uint `json:"versionId"`
}

// ApplyOptimizationRequest 应用优化建议请求
type ApplyOptimizationRequest struct {
	RequirementID  uint                    `json:"requirement_id" binding:"required"`
	OptimizationID uint                    `json:"optimization_id" binding:"required"`
	Action         string                  `json:"action" binding:"required,oneof=accept reject modify"`
	Reason         string                  `json:"reason,omitempty"`
	Suggestion     *OptimizationSuggestion `json:"suggestion" binding:"required"`
}

// OptimizationApplication 批量优化应用项
type OptimizationApplication struct {
	RequirementID  uint                    `json:"requirement_id" binding:"required"`
	OptimizationID uint                    `json:"optimization_id" binding:"required"`
	Action         string                  `json:"action" binding:"required,oneof=accept reject modify"`
	Suggestion     *OptimizationSuggestion `json:"suggestion" binding:"required"`
}

// BatchApplyOptimizationsRequest 批量应用优化建议请求
type BatchApplyOptimizationsRequest struct {
	Applications []OptimizationApplication `json:"applications" binding:"required,min=1"`
}

// ==================== 四级功能点生成相关请求 ====================

// Level4GenerationRequest 四级功能点生成请求
type Level4GenerationRequest struct {
	CycleID               uint   `json:"cycle_id" binding:"required"`
	Level3RequirementIDs  []uint `json:"level3RequirementIds"` // 为空时生成所有三级功能点的四级功能点
}

// Level4Suggestion 四级功能点建议
type Level4Suggestion struct {
	SuggestedTitle       string  `json:"suggestedTitle" binding:"required"`
	SuggestedDescription string  `json:"suggestedDescription" binding:"required"`
	SuggestedCode        string  `json:"suggestedCode"`
	FunctionType         string  `json:"functionType" binding:"required,oneof=EI EO EQ ILF EIF"`
	BusinessValue        string  `json:"businessValue"`
	AcceptanceCriteria   string  `json:"acceptanceCriteria"`
	EstimatedComplexity  string  `json:"estimatedComplexity" binding:"required,oneof=简单 中等 复杂"`
	RecommendedAFP       float64 `json:"recommendedAFP" binding:"required,min=0"`
	RecommendedUFP       float64 `json:"recommendedUFP" binding:"required,min=0"`
	Priority             int     `json:"priority" binding:"required,min=1,max=5"`
	Confidence           float64 `json:"confidence" binding:"required,min=0,max=1"`
	GenerationReason     string  `json:"generationReason"`
	RelatedKnowledge     string  `json:"relatedKnowledge"`
}

// CreateLevel4Request 创建四级功能点请求
type CreateLevel4Request struct {
	CycleID    uint              `json:"cycleId" binding:"required"`
	ParentID   uint              `json:"parentId" binding:"required"`
	Suggestion *Level4Suggestion `json:"suggestion" binding:"required"`
}

// Level4CreationRequest 四级功能点创建请求
type Level4CreationRequest struct {
	ParentID   uint              `json:"parentId" binding:"required"`
	Suggestion *Level4Suggestion `json:"suggestion" binding:"required"`
}

// BatchCreateLevel4Request 批量创建四级功能点请求
type BatchCreateLevel4Request struct {
	CycleID   uint                    `json:"cycleId" binding:"required"`
	Creations []Level4CreationRequest `json:"creations" binding:"required"`
}

// AsyncLevel4GenerationRequest 异步L4生成请求
type AsyncLevel4GenerationRequest struct {
	ProjectID            uint   `json:"projectId" binding:"required"`
	CycleID              uint   `json:"cycleId" binding:"required"`
	VersionID            uint   `json:"versionId" binding:"required"`
	L3RequirementIDs     []uint `json:"l3RequirementIds" binding:"required,min=1"`
	GenerationStrategy   string `json:"generationStrategy" binding:"omitempty,oneof=comprehensive decomposition scenario_extension process_refinement"`
	ComplexityLevel      string `json:"complexityLevel" binding:"omitempty,oneof=simple moderate complex"`
	MaxL4Count          int    `json:"maxL4Count" binding:"omitempty,min=1,max=20"`
	IncludeKnowledgeBase bool   `json:"includeKnowledgeBase"`
	AutoSave            bool   `json:"autoSave"`
}

// ConfirmLevel4RequirementsRequest 确认L4需求入库请求
type ConfirmLevel4RequirementsRequest struct {
	L4Requirements []ConfirmLevel4RequirementItem `json:"l4Requirements" binding:"required,min=1"`
}

// ConfirmLevel4RequirementItem 确认L4需求项
type ConfirmLevel4RequirementItem struct {
	ParentID           uint    `json:"parentId" binding:"required"`
	Code               string  `json:"code" binding:"required"`
	Title              string  `json:"title" binding:"required"`
	Description        string  `json:"description" binding:"required"`
	FunctionType       string  `json:"functionType" binding:"required"`
	BusinessValue      string  `json:"businessValue"`
	AcceptanceCriteria string  `json:"acceptanceCriteria"`
	Complexity         string  `json:"complexity" binding:"required"`
	AFP                float64 `json:"afp" binding:"required"`
	UFP                float64 `json:"ufp" binding:"required"`
	Priority           int     `json:"priority" binding:"required,min=1,max=5"`
	Confidence         float64 `json:"confidence" binding:"required,min=0,max=1"`
	Notes              string  `json:"notes"`
}

// ValidateLevel4SuggestionRequest 验证四级功能点建议请求
type ValidateLevel4SuggestionRequest struct {
	Suggestion *Level4Suggestion `json:"suggestion" binding:"required"`
}

// ==================== 三级功能点分析相关请求 ====================

// Level3AnalysisRequest 三级功能点分析请求
type Level3AnalysisRequest struct {
	CycleID        uint   `json:"cycleId" binding:"required"`
	RequirementIDs []uint `json:"requirementIds"` // 为空时分析所有三级功能点
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type           string  `json:"type" binding:"required,oneof=title description category"`
	Field          string  `json:"field" binding:"required"`
	CurrentValue   string  `json:"currentValue"`
	SuggestedValue string  `json:"suggestedValue" binding:"required"`
	Reason         string  `json:"reason" binding:"required"`
	Priority       int     `json:"priority" binding:"required,min=1,max=5"`
	Confidence     float64 `json:"confidence" binding:"required,min=0,max=1"`
	Category       string  `json:"category"`
}

// ExpansionSuggestion 扩充建议
type ExpansionSuggestion struct {
	SuggestedTitle       string  `json:"suggestedTitle" binding:"required"`
	SuggestedDescription string  `json:"suggestedDescription" binding:"required"`
	SuggestedCategory    string  `json:"suggestedCategory"`
	Justification        string  `json:"justification" binding:"required"`
	Priority             int     `json:"priority" binding:"required,min=1,max=5"`
	Confidence           float64 `json:"confidence" binding:"required,min=0,max=1"`
	RelatedKnowledge     string  `json:"relatedKnowledge"`
	EstimatedComplexity  string  `json:"estimatedComplexity" binding:"required,oneof=简单 中等 复杂"`
}

// CreateExpansionRequest 创建扩充功能点请求
type CreateExpansionRequest struct {
	CycleID    uint                 `json:"cycleId" binding:"required"`
	ParentID   uint                 `json:"parentId" binding:"required"`
	Suggestion *ExpansionSuggestion `json:"suggestion" binding:"required"`
}

// ExpansionCreation 扩充创建
type ExpansionCreation struct {
	CycleID    uint                 `json:"cycleId" binding:"required"`
	ParentID   uint                 `json:"parentId" binding:"required"`
	Suggestion *ExpansionSuggestion `json:"suggestion" binding:"required"`
}

// BatchCreateExpansionsRequest 批量创建扩充功能点请求
type BatchCreateExpansionsRequest struct {
	Expansions []ExpansionCreation `json:"expansions" binding:"required"`
}

// ==================== 功能点描述生成相关请求 ====================

// DescriptionGenerationRequest 功能点描述生成请求
type DescriptionGenerationRequest struct {
	CycleID        uint   `json:"cycleId" binding:"required"`
	RequirementIDs []uint `json:"requirementIds"` // 为空时生成所有功能点的描述
	Levels         []int  `json:"levels"`         // 指定级别，为空时包含所有级别
}

// ApplyDescriptionRequest 应用描述请求
type ApplyDescriptionRequest struct {
	RequirementID        uint                    `json:"requirementId" binding:"required"`
	EnhancedDescription  *EnhancedDescription    `json:"enhancedDescription" binding:"required"`
}

// EnhancedDescription 增强的功能描述
type EnhancedDescription struct {
	Title                string                 `json:"title" binding:"required"`
	Summary              string                 `json:"summary"`
	DetailedDescription  string                 `json:"detailedDescription" binding:"required"`
	BusinessContext      string                 `json:"businessContext"`
	FunctionalFlow       FunctionalFlow         `json:"functionalFlow"`
	DataElements         []DataElement          `json:"dataElements"`
	BusinessRules        []BusinessRule         `json:"businessRules"`
	QualityAttributes    []QualityAttribute     `json:"qualityAttributes"`
	AcceptanceCriteria   []AcceptanceCriterion  `json:"acceptanceCriteria"`
	TestScenarios        []TestScenario         `json:"testScenarios"`
	Dependencies         []string               `json:"dependencies"`
	Assumptions          []string               `json:"assumptions"`
	Constraints          []string               `json:"constraints"`
	RiskFactors          []string               `json:"riskFactors"`
}

// FunctionalFlow 功能流程
type FunctionalFlow struct {
	InputSources      []string `json:"inputSources"`
	ProcessingSteps   []string `json:"processingSteps"`
	OutputTargets     []string `json:"outputTargets"`
	ExceptionHandling []string `json:"exceptionHandling"`
}

// DataElement 数据元素
type DataElement struct {
	Name        string `json:"name" binding:"required"`
	Type        string `json:"type" binding:"required"`
	Description string `json:"description"`
	Mandatory   bool   `json:"mandatory"`
	Format      string `json:"format"`
	Validation  string `json:"validation"`
}

// BusinessRule 业务规则
type BusinessRule struct {
	RuleID      string `json:"ruleId" binding:"required"`
	Description string `json:"description" binding:"required"`
	Condition   string `json:"condition"`
	Action      string `json:"action"`
	Priority    string `json:"priority" binding:"required,oneof=高 中 低"`
}

// QualityAttribute 质量属性
type QualityAttribute struct {
	Attribute   string `json:"attribute" binding:"required"`
	Requirement string `json:"requirement" binding:"required"`
	Measurement string `json:"measurement"`
}

// AcceptanceCriterion 验收标准
type AcceptanceCriterion struct {
	CriterionID string `json:"criterionId" binding:"required"`
	Description string `json:"description" binding:"required"`
	Priority    string `json:"priority" binding:"required,oneof=高 中 低"`
	Testable    bool   `json:"testable"`
}

// TestScenario 测试场景
type TestScenario struct {
	ScenarioID    string   `json:"scenarioId" binding:"required"`
	Description   string   `json:"description" binding:"required"`
	Preconditions []string `json:"preconditions"`
	Steps         []string `json:"steps" binding:"required"`
	Expected      string   `json:"expected" binding:"required"`
}

// BatchDescriptionGenerationRequest 批量描述生成请求
type BatchDescriptionGenerationRequest struct {
	Requests []DescriptionGenerationRequest `json:"requests" binding:"required"`
}

// PreviewDescriptionRequest 预览描述请求
type PreviewDescriptionRequest struct {
	RequirementID uint `json:"requirementId" binding:"required"`
}

// ValidateDescriptionRequest 验证描述请求
type ValidateDescriptionRequest struct {
	EnhancedDescription *EnhancedDescription `json:"enhancedDescription" binding:"required"`
}

// ==================== Mermaid流程图生成相关请求 ====================

// MermaidGenerationRequest Mermaid流程图生成请求
type MermaidGenerationRequest struct {
	CycleID        uint   `json:"cycleId" binding:"required"`
	RequirementIDs []uint `json:"requirementIds"` // 为空时生成所有四级功能点的流程图
	DiagramType    string `json:"diagramType"`    // flowchart/sequence/class/state，默认flowchart
}

// AsyncMermaidGenerationRequest 异步Mermaid流程图生成请求
type AsyncMermaidGenerationRequest struct {
	ProjectID      uint   `json:"projectId" binding:"required"`
	CycleID        uint   `json:"cycleId" binding:"required"`
	VersionID      uint   `json:"versionId" binding:"required"`
	RequirementIDs []uint `json:"requirementIds"` // 为空时生成所有四级功能点的流程图
	DiagramType    string `json:"diagramType"`    // flowchart/sequence/class/state，默认flowchart
	DetailLevel    string `json:"detailLevel"`    // basic/detailed/complete，默认detailed
	Options        struct {
		IncludeSubRequirements bool `json:"includeSubRequirements"` // 是否包含子需求
		AutoLayout             bool `json:"autoLayout"`             // 是否自动布局
		AddAnnotations         bool `json:"addAnnotations"`         // 是否添加注释
		GroupByLevel           bool `json:"groupByLevel"`           // 是否按层级分组
	} `json:"options"`
}

// ApplyMermaidRequest 应用Mermaid流程图请求
type ApplyMermaidRequest struct {
	RequirementID   uint            `json:"requirementId" binding:"required"`
	MermaidDiagram  *MermaidDiagram `json:"mermaidDiagram" binding:"required"`
}

// MermaidDiagram Mermaid流程图
type MermaidDiagram struct {
	DiagramType  string           `json:"diagramType" binding:"required"`
	Title        string           `json:"title" binding:"required"`
	Description  string           `json:"description"`
	MermaidCode  string           `json:"mermaidCode" binding:"required"`
	Nodes        []MermaidNode    `json:"nodes"`
	Edges        []MermaidEdge    `json:"edges"`
	Subgraphs    []MermaidSubgraph `json:"subgraphs"`
	Styling      MermaidStyling   `json:"styling"`
	Metadata     MermaidMetadata  `json:"metadata"`
}

// MermaidNode Mermaid节点
type MermaidNode struct {
	ID         string            `json:"id" binding:"required"`
	Label      string            `json:"label" binding:"required"`
	Type       string            `json:"type" binding:"required,oneof=start end process decision data connector"`
	Shape      string            `json:"shape" binding:"required"`
	Category   string            `json:"category" binding:"required,oneof=input process output decision exception"`
	Properties map[string]string `json:"properties"`
}

// MermaidEdge Mermaid边
type MermaidEdge struct {
	ID         string            `json:"id" binding:"required"`
	From       string            `json:"from" binding:"required"`
	To         string            `json:"to" binding:"required"`
	Label      string            `json:"label"`
	Type       string            `json:"type" binding:"required"`
	Condition  string            `json:"condition"`
	Properties map[string]string `json:"properties"`
}

// MermaidSubgraph Mermaid子图
type MermaidSubgraph struct {
	ID          string   `json:"id" binding:"required"`
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Nodes       []string `json:"nodes"`
	Style       string   `json:"style"`
}

// MermaidStyling Mermaid样式
type MermaidStyling struct {
	Theme       string            `json:"theme"`
	NodeStyles  map[string]string `json:"nodeStyles"`
	EdgeStyles  map[string]string `json:"edgeStyles"`
	CustomCSS   string            `json:"customCss"`
	ColorScheme map[string]string `json:"colorScheme"`
}

// MermaidMetadata Mermaid元数据
type MermaidMetadata struct {
	ComplexityLevel string `json:"complexityLevel" binding:"required,oneof=simple medium complex"`
	NodeCount       int    `json:"nodeCount" binding:"required,min=1"`
	EdgeCount       int    `json:"edgeCount" binding:"required,min=0"`
	MaxDepth        int    `json:"maxDepth" binding:"required,min=1"`
	Version         string `json:"version"`
}

// BatchMermaidGenerationRequest 批量Mermaid生成请求
type BatchMermaidGenerationRequest struct {
	Requests []MermaidGenerationRequest `json:"requests" binding:"required"`
}

// ValidateMermaidRequest 验证Mermaid请求
type ValidateMermaidRequest struct {
	MermaidDiagram *MermaidDiagram `json:"mermaidDiagram" binding:"required"`
}

// PreviewMermaidRequest 预览Mermaid请求
type PreviewMermaidRequest struct {
	RequirementID uint   `json:"requirementId" binding:"required"`
	DiagramType   string `json:"diagramType"` // 默认flowchart
}

// ==================== 文档导出相关请求 ====================

// DocumentExportRequest 文档导出请求
type DocumentExportRequest struct {
	CycleID        uint   `json:"cycleId" binding:"required"`
	ExportType     string `json:"exportType"` // comprehensive/requirement_only/analysis_only
	DocumentFormat string `json:"documentFormat"` // word/excel/pdf
	IncludeSections []string `json:"includeSections"` // 包含的章节
	ExcludeSections []string `json:"excludeSections"` // 排除的章节
	CustomTemplate  string `json:"customTemplate"` // 自定义模板
	OutputPath      string `json:"outputPath"` // 输出路径
	FileName        string `json:"fileName"` // 文件名
	CompressionType string `json:"compressionType"` // 压缩类型
	Watermark       string `json:"watermark"` // 水印
	Password        string `json:"password"` // 密码保护
}

// BatchDocumentExportRequest 批量文档导出请求
type BatchDocumentExportRequest struct {
	ExportRequests []DocumentExportRequest `json:"exportRequests" binding:"required"`
}

// ExportPreviewRequest 导出预览请求
type ExportPreviewRequest struct {
	CycleID        uint   `json:"cycleId" binding:"required"`
	ExportType     string `json:"exportType"` // comprehensive/requirement_only/analysis_only
	DocumentFormat string `json:"documentFormat"` // word/excel/pdf
}

// ==================== 统一智能分析工作流相关请求 ====================

// CreateWorkflowFromTemplateRequest 从模板创建工作流请求
type CreateWorkflowFromTemplateRequest struct {
	TemplateID          string                 `json:"templateId" binding:"required"`
	CycleID             uint                   `json:"cycleId" binding:"required"`
	ProjectID           uint                   `json:"projectId" binding:"required"`
	CustomConfiguration map[string]interface{} `json:"customConfiguration"`
	WorkflowName        string                 `json:"workflowName"`
	WorkflowDescription string                 `json:"workflowDescription"`
}

// WorkflowExecutionRequest 工作流执行请求
type WorkflowExecutionRequest struct {
	WorkflowID           string                         `json:"workflowId" binding:"required"`
	CycleID              uint                           `json:"cycleId" binding:"required"`
	ProjectID            uint                           `json:"projectId" binding:"required"`
	ExecutionMode        string                         `json:"executionMode"` // sequential/parallel/adaptive
	EnabledPhases        []string                       `json:"enabledPhases"`
	PhaseConfigurations  map[string]PhaseConfigRequest  `json:"phaseConfigurations"`
	QualityThresholds    QualityThresholdsRequest       `json:"qualityThresholds"`
	RetryPolicy          RetryPolicyRequest             `json:"retryPolicy"`
	NotificationSettings NotificationSettingsRequest    `json:"notificationSettings"`
	TimeoutMinutes       int                            `json:"timeoutMinutes"`
}

// PhaseConfigRequest 阶段配置请求
type PhaseConfigRequest struct {
	Enabled              bool                   `json:"enabled"`
	Priority             int                    `json:"priority"`
	MaxRetries           int                    `json:"maxRetries"`
	TimeoutSeconds       int                    `json:"timeoutSeconds"`
	QualityThreshold     float64                `json:"qualityThreshold"`
	ParallelProcessing   bool                   `json:"parallelProcessing"`
	BatchSize            int                    `json:"batchSize"`
	CustomParameters     map[string]interface{} `json:"customParameters"`
}

// QualityThresholdsRequest 质量阈值请求
type QualityThresholdsRequest struct {
	Level3AnalysisConfidence   float64 `json:"level3AnalysisConfidence"`
	Level4GenerationConfidence float64 `json:"level4GenerationConfidence"`
	DescriptionQualityScore    float64 `json:"descriptionQualityScore"`
	MermaidValidationScore     float64 `json:"mermaidValidationScore"`
	OverallQualityScore        float64 `json:"overallQualityScore"`
}

// RetryPolicyRequest 重试策略请求
type RetryPolicyRequest struct {
	MaxRetries         int      `json:"maxRetries"`
	RetryDelaySeconds  int      `json:"retryDelaySeconds"`
	ExponentialBackoff bool     `json:"exponentialBackoff"`
	RetryableErrors    []string `json:"retryableErrors"`
}

// NotificationSettingsRequest 通知设置请求
type NotificationSettingsRequest struct {
	EnableNotifications  bool     `json:"enableNotifications"`
	NotificationChannels []string `json:"notificationChannels"`
	NotificationTriggers []string `json:"notificationTriggers"`
	WebhookURL          string   `json:"webhookUrl"`
	EmailRecipients     []string `json:"emailRecipients"`
}

// CompareRequirementVersionsRequest 对比需求版本请求
type CompareRequirementVersionsRequest struct {
	RequirementID uint `json:"requirementId" binding:"required"`
	VersionID1    uint `json:"versionId1" binding:"required"`
	VersionID2    uint `json:"versionId2" binding:"required"`
}

// UpdateRequirementAIStatusRequest 更新需求AI分析状态请求
type UpdateRequirementAIStatusRequest struct {
	RequirementIDs    []uint `json:"requirementIds" binding:"required"`
	AIAnalysisStatus  string `json:"aiAnalysisStatus" binding:"required"`
	AIDescription     string `json:"aiDescription"`
	AIGeneratedTitle  string `json:"aiGeneratedTitle"`
	AIComplexityScore *float64 `json:"aiComplexityScore"`
	AIConfidenceScore *float64 `json:"aiConfidenceScore"`
	RecommendedAFP    *float64 `json:"recommendedAFP"`
	RecommendedUFP    *float64 `json:"recommendedUFP"`
}

// GetAIOptimizedRequirementsRequest 获取AI优化需求列表请求
type GetAIOptimizedRequirementsRequest struct {
	request.PageInfo
	ProjectID *uint  `json:"projectId" form:"projectId"`
	CycleID   *uint  `json:"cycleId" form:"cycleId"`
	VersionID *uint  `json:"versionId" form:"versionId"`
	Status    string `json:"status" form:"status"` // pending, analyzing, completed, failed
}

// ==================== 分析任务管理相关请求 ====================

// AnalysisTaskListRequest 分析任务列表查询请求
type AnalysisTaskListRequest struct {
	request.PageInfo
	ProjectID *uint  `json:"projectId" form:"projectId" binding:"omitempty"`
	CycleID   *uint  `json:"cycleId" form:"cycleId" binding:"omitempty"`
	TaskType  string `json:"taskType" form:"taskType" binding:"omitempty,oneof=requirement_analysis description_generation flowchart_generation mermaid_generation level3_analysis level4_generation"` // 任务类型
	Status    string `json:"status" form:"status" binding:"omitempty,oneof=pending running completed failed cancelled"`       // 任务状态
	StartDate string `json:"startDate" form:"startDate" binding:"omitempty"` // 创建开始时间 YYYY-MM-DD HH:mm:ss
	EndDate   string `json:"endDate" form:"endDate" binding:"omitempty"`     // 创建结束时间 YYYY-MM-DD HH:mm:ss
}

// TaskStatisticsRequest 任务统计查询请求
type TaskStatisticsRequest struct {
	ProjectID *uint `json:"projectId" form:"projectId" binding:"omitempty"`
	CycleID   *uint `json:"cycleId" form:"cycleId" binding:"omitempty"`
}
