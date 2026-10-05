package nesma

import (
	"time"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// AIProjectAnalysis AI项目全面分析结果主表
type AIProjectAnalysis struct {
	global.GVA_MODEL
	
	// 基础信息
	ProjectID       uint   `json:"projectId" gorm:"not null;comment:项目ID;index"`
	EvaluationID    uint   `json:"evaluationId" gorm:"not null;comment:评估ID;index"`
	AnalysisVersion string `json:"analysisVersion" gorm:"type:varchar(50);comment:分析版本"`
	AnalysisType    string `json:"analysisType" gorm:"type:varchar(50);comment:分析类型:comprehensive,focused,incremental"`
	
	// 分析状态
	Status          string    `json:"status" gorm:"type:varchar(50);default:'processing';comment:分析状态:processing,completed,failed"`
	Progress        float64   `json:"progress" gorm:"comment:分析进度0-100"`
	StartTime       time.Time `json:"startTime" gorm:"comment:分析开始时间"`
	CompletionTime  *time.Time `json:"completionTime" gorm:"comment:分析完成时间"`
	
	// 输入数据统计
	TotalRequirements    int `json:"totalRequirements" gorm:"comment:总需求数"`
	ProcessedRequirements int `json:"processedRequirements" gorm:"comment:已处理需求数"`
	RequirementLevels    string `json:"requirementLevels" gorm:"type:text;comment:需求层级分布JSON"`
	ProjectContext       string `json:"projectContext" gorm:"type:text;comment:项目上下文信息"`
	
	// AI分析配置
	AIModel         string `json:"aiModel" gorm:"type:varchar(100);comment:使用的AI模型"`
	MaxTokens       int    `json:"maxTokens" gorm:"comment:最大Token数"`
	Temperature     float64 `json:"temperature" gorm:"comment:AI温度参数"`
	BatchSize       int    `json:"batchSize" gorm:"comment:分批处理大小"`
	
	// 分析结果摘要
	OverallScore        float64 `json:"overallScore" gorm:"comment:总体质量得分"`
	ComplexityLevel     string  `json:"complexityLevel" gorm:"type:varchar(50);comment:复杂度等级"`
	RiskLevel           string  `json:"riskLevel" gorm:"type:varchar(50);comment:风险等级"`
	RecommendationLevel string  `json:"recommendationLevel" gorm:"type:varchar(50);comment:建议等级"`
	
	// 关联表
	Project                *NesmaProject                `json:"project" gorm:"foreignKey:ProjectID"`
	Evaluation            *NesmaEvaluation             `json:"evaluation" gorm:"foreignKey:EvaluationID"`
	FunctionalAnalysis    []AIFunctionalAnalysis       `json:"functionalAnalysis" gorm:"foreignKey:AnalysisID"`
	TechnicalAnalysis     []AITechnicalAnalysis        `json:"technicalAnalysis" gorm:"foreignKey:AnalysisID"`
	BusinessAnalysis      []AIBusinessAnalysis         `json:"businessAnalysis" gorm:"foreignKey:AnalysisID"`
	QualityAnalysis       []AIQualityAnalysis          `json:"qualityAnalysis" gorm:"foreignKey:AnalysisID"`
	RiskAnalysis          []AIRiskAnalysis             `json:"riskAnalysis" gorm:"foreignKey:AnalysisID"`
	RecommendationAnalysis []AIRecommendationAnalysis  `json:"recommendationAnalysis" gorm:"foreignKey:AnalysisID"`
	ComplianceAnalysis    []AIComplianceAnalysis       `json:"complianceAnalysis" gorm:"foreignKey:AnalysisID"`
}

// AIFunctionalAnalysis 功能性分析
type AIFunctionalAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// NESMA功能点分析
	TotalFunctionPoints      float64 `json:"totalFunctionPoints" gorm:"comment:总功能点数"`
	DataFunctionPoints       float64 `json:"dataFunctionPoints" gorm:"comment:数据功能点数"`
	TransactionalFunctionPoints float64 `json:"transactionalFunctionPoints" gorm:"comment:事务功能点数"`
	
	// 功能点类型分布
	ILFCount int `json:"ilfCount" gorm:"comment:内部逻辑文件数量"`
	EIFCount int `json:"eifCount" gorm:"comment:外部接口文件数量"`
	EICount  int `json:"eiCount" gorm:"comment:外部输入数量"`
	EOCount  int `json:"eoCount" gorm:"comment:外部输出数量"`
	EQCount  int `json:"eqCount" gorm:"comment:外部查询数量"`
	
	// 复杂度分析
	LowComplexityCount    int `json:"lowComplexityCount" gorm:"comment:低复杂度功能数"`
	MediumComplexityCount int `json:"mediumComplexityCount" gorm:"comment:中复杂度功能数"`
	HighComplexityCount   int `json:"highComplexityCount" gorm:"comment:高复杂度功能数"`
	
	// AI分析结果
	FunctionalCoverage    float64 `json:"functionalCoverage" gorm:"comment:功能覆盖度"`
	FunctionalCompleteness float64 `json:"functionalCompleteness" gorm:"comment:功能完整性"`
	FunctionalConsistency  float64 `json:"functionalConsistency" gorm:"comment:功能一致性"`
	
	// 详细分析
	FunctionTypeDistribution string `json:"functionTypeDistribution" gorm:"type:text;comment:功能类型分布分析JSON"`
	ComplexityAnalysis       string `json:"complexityAnalysis" gorm:"type:text;comment:复杂度分析JSON"`
	FunctionalGaps           string `json:"functionalGaps" gorm:"type:text;comment:功能缺口分析JSON"`
	
	// AI评估
	AIAssessment      string  `json:"aiAssessment" gorm:"type:text;comment:AI评估结果"`
	ConfidenceScore   float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
	QualityIndicators string  `json:"qualityIndicators" gorm:"type:text;comment:质量指标JSON"`
}

// AITechnicalAnalysis 技术性分析
type AITechnicalAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// 技术架构分析
	ArchitectureType      string  `json:"architectureType" gorm:"type:varchar(100);comment:架构类型"`
	TechnologyStack       string  `json:"technologyStack" gorm:"type:text;comment:技术栈JSON"`
	IntegrationComplexity string  `json:"integrationComplexity" gorm:"type:varchar(50);comment:集成复杂度"`
	
	// 性能分析
	PerformanceRequirements string  `json:"performanceRequirements" gorm:"type:text;comment:性能需求JSON"`
	ScalabilityAnalysis     string  `json:"scalabilityAnalysis" gorm:"type:text;comment:可扩展性分析JSON"`
	ReliabilityAnalysis     string  `json:"reliabilityAnalysis" gorm:"type:text;comment:可靠性分析JSON"`
	
	// 技术债务和风险
	TechnicalDebt          string  `json:"technicalDebt" gorm:"type:text;comment:技术债务分析JSON"`
	SecurityConsiderations string  `json:"securityConsiderations" gorm:"type:text;comment:安全考虑JSON"`
	MaintenanceComplexity  float64 `json:"maintenanceComplexity" gorm:"comment:维护复杂度"`
	
	// 开发估算
	DevelopmentEffort      float64 `json:"developmentEffort" gorm:"comment:开发工作量估算"`
	TestingEffort          float64 `json:"testingEffort" gorm:"comment:测试工作量估算"`
	DeploymentComplexity   float64 `json:"deploymentComplexity" gorm:"comment:部署复杂度"`
	
	// AI技术评估
	TechnicalFeasibility   float64 `json:"technicalFeasibility" gorm:"comment:技术可行性"`
	ImplementationRisk     float64 `json:"implementationRisk" gorm:"comment:实现风险"`
	TechnicalInnovation    float64 `json:"technicalInnovation" gorm:"comment:技术创新度"`
	
	// 详细分析
	AITechnicalAssessment  string  `json:"aiTechnicalAssessment" gorm:"type:text;comment:AI技术评估"`
	TechnicalRecommendations string `json:"technicalRecommendations" gorm:"type:text;comment:技术建议"`
	ConfidenceScore        float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
}

// AIBusinessAnalysis 业务性分析
type AIBusinessAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// 业务价值分析
	BusinessValue         float64 `json:"businessValue" gorm:"comment:业务价值评分"`
	ROIEstimation         float64 `json:"roiEstimation" gorm:"comment:投资回报率估算"`
	StrategicAlignment    float64 `json:"strategicAlignment" gorm:"comment:战略一致性"`
	
	// 用户体验分析
	UserExperienceScore   float64 `json:"userExperienceScore" gorm:"comment:用户体验评分"`
	UsabilityAnalysis     string  `json:"usabilityAnalysis" gorm:"type:text;comment:可用性分析JSON"`
	AccessibilityAnalysis string  `json:"accessibilityAnalysis" gorm:"type:text;comment:可访问性分析JSON"`
	
	// 业务流程分析
	ProcessEfficiency     float64 `json:"processEfficiency" gorm:"comment:流程效率"`
	ProcessAutomation     float64 `json:"processAutomation" gorm:"comment:流程自动化程度"`
	BusinessLogicComplexity float64 `json:"businessLogicComplexity" gorm:"comment:业务逻辑复杂度"`
	
	// 市场和竞争分析
	MarketFit             float64 `json:"marketFit" gorm:"comment:市场适配度"`
	CompetitiveAdvantage  float64 `json:"competitiveAdvantage" gorm:"comment:竞争优势"`
	InnovationLevel       float64 `json:"innovationLevel" gorm:"comment:创新水平"`
	
	// 业务风险
	BusinessRisk          string  `json:"businessRisk" gorm:"type:text;comment:业务风险分析JSON"`
	RegulatoryCompliance  float64 `json:"regulatoryCompliance" gorm:"comment:法规合规性"`
	
	// AI业务评估
	AIBusinessAssessment  string  `json:"aiBusinessAssessment" gorm:"type:text;comment:AI业务评估"`
	BusinessRecommendations string `json:"businessRecommendations" gorm:"type:text;comment:业务建议"`
	ConfidenceScore       float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
}

// AIQualityAnalysis 质量分析
type AIQualityAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// 质量指标
	OverallQuality        float64 `json:"overallQuality" gorm:"comment:总体质量"`
	RequirementQuality    float64 `json:"requirementQuality" gorm:"comment:需求质量"`
	DesignQuality         float64 `json:"designQuality" gorm:"comment:设计质量"`
	
	// 质量特征分析
	Correctness           float64 `json:"correctness" gorm:"comment:正确性"`
	Completeness          float64 `json:"completeness" gorm:"comment:完整性"`
	Consistency           float64 `json:"consistency" gorm:"comment:一致性"`
	Clarity               float64 `json:"clarity" gorm:"comment:清晰性"`
	Traceability          float64 `json:"traceability" gorm:"comment:可追溯性"`
	
	// 可维护性分析
	Maintainability       float64 `json:"maintainability" gorm:"comment:可维护性"`
	Modularity            float64 `json:"modularity" gorm:"comment:模块化程度"`
	Reusability           float64 `json:"reusability" gorm:"comment:可重用性"`
	Testability           float64 `json:"testability" gorm:"comment:可测试性"`
	
	// 质量问题识别
	CriticalIssues        string  `json:"criticalIssues" gorm:"type:text;comment:关键问题JSON"`
	QualityGaps           string  `json:"qualityGaps" gorm:"type:text;comment:质量缺口JSON"`
	ImprovementAreas      string  `json:"improvementAreas" gorm:"type:text;comment:改进领域JSON"`
	
	// AI质量评估
	AIQualityAssessment   string  `json:"aiQualityAssessment" gorm:"type:text;comment:AI质量评估"`
	QualityRecommendations string `json:"qualityRecommendations" gorm:"type:text;comment:质量建议"`
	ConfidenceScore       float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
}

// AIRiskAnalysis 风险分析
type AIRiskAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// 总体风险评估
	OverallRisk           float64 `json:"overallRisk" gorm:"comment:总体风险"`
	RiskLevel             string  `json:"riskLevel" gorm:"type:varchar(50);comment:风险等级"`
	
	// 技术风险
	TechnicalRisk         float64 `json:"technicalRisk" gorm:"comment:技术风险"`
	ImplementationRisk    float64 `json:"implementationRisk" gorm:"comment:实现风险"`
	IntegrationRisk       float64 `json:"integrationRisk" gorm:"comment:集成风险"`
	
	// 项目风险
	ScheduleRisk          float64 `json:"scheduleRisk" gorm:"comment:进度风险"`
	BudgetRisk            float64 `json:"budgetRisk" gorm:"comment:预算风险"`
	ResourceRisk          float64 `json:"resourceRisk" gorm:"comment:资源风险"`
	
	// 业务风险
	BusinessRisk          float64 `json:"businessRisk" gorm:"comment:业务风险"`
	MarketRisk            float64 `json:"marketRisk" gorm:"comment:市场风险"`
	ComplianceRisk        float64 `json:"complianceRisk" gorm:"comment:合规风险"`
	
	// 风险详情
	IdentifiedRisks       string  `json:"identifiedRisks" gorm:"type:text;comment:识别的风险JSON"`
	RiskMitigationPlans   string  `json:"riskMitigationPlans" gorm:"type:text;comment:风险缓解计划JSON"`
	ContingencyPlans      string  `json:"contingencyPlans" gorm:"type:text;comment:应急计划JSON"`
	
	// AI风险评估
	AIRiskAssessment      string  `json:"aiRiskAssessment" gorm:"type:text;comment:AI风险评估"`
	RiskRecommendations   string  `json:"riskRecommendations" gorm:"type:text;comment:风险建议"`
	ConfidenceScore       float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
}

// AIRecommendationAnalysis 改进建议分析
type AIRecommendationAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// 建议分类
	RecommendationType    string  `json:"recommendationType" gorm:"type:varchar(100);comment:建议类型"`
	Priority              string  `json:"priority" gorm:"type:varchar(50);comment:优先级"`
	ImpactLevel           string  `json:"impactLevel" gorm:"type:varchar(50);comment:影响程度"`
	
	// 建议内容
	Title                 string  `json:"title" gorm:"type:varchar(200);comment:建议标题"`
	Description           string  `json:"description" gorm:"type:text;comment:建议描述"`
	Rationale             string  `json:"rationale" gorm:"type:text;comment:建议理由"`
	
	// 实施相关
	ImplementationSteps   string  `json:"implementationSteps" gorm:"type:text;comment:实施步骤JSON"`
	EstimatedEffort       float64 `json:"estimatedEffort" gorm:"comment:估算工作量"`
	ExpectedBenefit       string  `json:"expectedBenefit" gorm:"type:text;comment:预期收益"`
	
	// 依赖和约束
	Dependencies          string  `json:"dependencies" gorm:"type:text;comment:依赖关系JSON"`
	Constraints           string  `json:"constraints" gorm:"type:text;comment:约束条件JSON"`
	Risks                 string  `json:"risks" gorm:"type:text;comment:相关风险JSON"`
	
	// AI生成
	AIGeneratedRecommendation string  `json:"aiGeneratedRecommendation" gorm:"type:text;comment:AI生成的建议"`
	ConfidenceScore       float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
	Urgency               string  `json:"urgency" gorm:"type:varchar(50);comment:紧急程度"`
}

// AIComplianceAnalysis 合规性分析
type AIComplianceAnalysis struct {
	global.GVA_MODEL
	AnalysisID uint `json:"analysisId" gorm:"not null;comment:分析ID;index"`
	
	// 合规标准
	ComplianceStandard    string  `json:"complianceStandard" gorm:"type:varchar(100);comment:合规标准"`
	StandardVersion       string  `json:"standardVersion" gorm:"type:varchar(50);comment:标准版本"`
	
	// 合规评估
	OverallCompliance     float64 `json:"overallCompliance" gorm:"comment:总体合规性"`
	ComplianceLevel       string  `json:"complianceLevel" gorm:"type:varchar(50);comment:合规等级"`
	
	// 具体合规项
	ComplianceItems       string  `json:"complianceItems" gorm:"type:text;comment:合规项目JSON"`
	NonComplianceItems    string  `json:"nonComplianceItems" gorm:"type:text;comment:不合规项目JSON"`
	PartialComplianceItems string `json:"partialComplianceItems" gorm:"type:text;comment:部分合规项目JSON"`
	
	// 合规建议
	ComplianceGaps        string  `json:"complianceGaps" gorm:"type:text;comment:合规缺口JSON"`
	RemediationPlan       string  `json:"remediationPlan" gorm:"type:text;comment:整改计划JSON"`
	
	// AI合规评估
	AIComplianceAssessment string  `json:"aiComplianceAssessment" gorm:"type:text;comment:AI合规评估"`
	ComplianceRecommendations string `json:"complianceRecommendations" gorm:"type:text;comment:合规建议"`
	ConfidenceScore       float64 `json:"confidenceScore" gorm:"comment:置信度分数"`
}

// TableName 设置表名
func (AIProjectAnalysis) TableName() string {
	return "ai_project_analyses"
}

func (AIFunctionalAnalysis) TableName() string {
	return "ai_functional_analyses"
}

func (AITechnicalAnalysis) TableName() string {
	return "ai_technical_analyses"
}

func (AIBusinessAnalysis) TableName() string {
	return "ai_business_analyses"
}

func (AIQualityAnalysis) TableName() string {
	return "ai_quality_analyses"
}

func (AIRiskAnalysis) TableName() string {
	return "ai_risk_analyses"
}

func (AIRecommendationAnalysis) TableName() string {
	return "ai_recommendation_analyses"
}

func (AIComplianceAnalysis) TableName() string {
	return "ai_compliance_analyses"
}