package nesma

import (
	"time"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaProjectAnalysis 项目分析记录
type NesmaProjectAnalysis struct {
	global.GVA_MODEL
	ProjectID       uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	CycleID         uint           `json:"cycleId" gorm:"not null;comment:周期ID;index"`
	VersionID       uint           `json:"versionId" gorm:"not null;comment:版本ID;index"`
	
	// 分析基本信息
	AnalysisType    string         `json:"analysisType" gorm:"type:varchar(50);not null;comment:分析类型：project_overview-项目概览,requirement_analysis-需求分析,quality_assessment-质量评估"`
	AnalysisScope   string         `json:"analysisScope" gorm:"type:varchar(100);comment:分析范围"`
	AnalysisDate    time.Time      `json:"analysisDate" gorm:"not null;comment:分析日期"`
	
	// 分析结果概要
	OverallGrade    string         `json:"overallGrade" gorm:"type:varchar(10);comment:总体评级：A+,A,A-,B+,B,B-,C+,C,C-,D,F"`
	HealthScore     float64        `json:"healthScore" gorm:"comment:项目健康分(0-100)"`
	SuccessProbability float64     `json:"successProbability" gorm:"comment:成功概率(0-100)"`
	QualityScore    float64        `json:"qualityScore" gorm:"comment:质量评分(0-100)"`
	
	// 统计数据
	TotalRequirements   int        `json:"totalRequirements" gorm:"comment:总需求数量"`
	TotalFunctionPoints float64    `json:"totalFunctionPoints" gorm:"comment:总功能点数"`
	ComplexityIndex     float64    `json:"complexityIndex" gorm:"comment:复杂度指数"`
	
	// 分析详细结果 (JSON格式)
	AnalysisResult  datatypes.JSON `json:"analysisResult" gorm:"type:json;comment:详细分析结果"`
	
	// 改进建议
	KeyStrengths    datatypes.JSON `json:"keyStrengths" gorm:"type:json;comment:核心优势"`
	MajorRisks      datatypes.JSON `json:"majorRisks" gorm:"type:json;comment:主要风险"`
	Recommendations datatypes.JSON `json:"recommendations" gorm:"type:json;comment:改进建议"`
	
	// AI分析信息
	AIModel         string         `json:"aiModel" gorm:"type:varchar(50);comment:使用的AI模型"`
	ConfidenceScore float64        `json:"confidenceScore" gorm:"comment:AI置信度(0-1)"`
	
	// 关联关系
	Project         NesmaProject          `json:"project" gorm:"foreignKey:ProjectID"`
	Cycle           NesmaProjectCycle     `json:"cycle" gorm:"foreignKey:CycleID"`
	Version         NesmaRequirementVersion `json:"version" gorm:"foreignKey:VersionID"`
}

// TableName 自定义表名
func (NesmaProjectAnalysis) TableName() string {
	return "nesma_project_analyses"
}

// NesmaRequirementOptimization 需求优化建议
type NesmaRequirementOptimization struct {
	global.GVA_MODEL
	RequirementID   uint           `json:"requirementId" gorm:"not null;comment:需求ID;index"`
	ProjectID       uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	OptimizationType string        `json:"optimizationType" gorm:"type:varchar(50);not null;comment:优化类型：title-标题优化,description-描述优化,structure-结构优化,quality-质量提升"`
	
	// 原始信息
	OriginalTitle       string         `json:"originalTitle" gorm:"type:varchar(500);comment:原始标题"`
	OriginalDescription string         `json:"originalDescription" gorm:"type:text;comment:原始描述"`
	OriginalQualityScore float64       `json:"originalQualityScore" gorm:"comment:原始质量评分"`
	
	// 优化后信息
	OptimizedTitle      string         `json:"optimizedTitle" gorm:"type:varchar(500);comment:优化后标题"`
	OptimizedDescription string        `json:"optimizedDescription" gorm:"type:text;comment:优化后描述"`
	OptimizedQualityScore float64      `json:"optimizedQualityScore" gorm:"comment:优化后质量评分"`
	
	// 优化分析
	ImprovementSummary  datatypes.JSON `json:"improvementSummary" gorm:"type:json;comment:改进摘要"`
	QualityComparison   datatypes.JSON `json:"qualityComparison" gorm:"type:json;comment:质量对比"`
	OptimizationRationale string       `json:"optimizationRationale" gorm:"type:text;comment:优化理由"`
	
	// 用户决策
	UserDecisionRequired bool          `json:"userDecisionRequired" gorm:"default:false;comment:是否需要用户决策"`
	DecisionPoints      datatypes.JSON `json:"decisionPoints" gorm:"type:json;comment:决策点"`
	UserDecision        string         `json:"userDecision" gorm:"type:varchar(20);comment:用户决策：accepted-接受,rejected-拒绝,pending-待决策"`
	UserDecisionTime    *time.Time     `json:"userDecisionTime" gorm:"comment:用户决策时间"`
	UserDecisionReason  string         `json:"userDecisionReason" gorm:"type:text;comment:用户决策理由"`
	
	// 应用状态
	AppliedToRequirement bool          `json:"appliedToRequirement" gorm:"default:false;comment:是否已应用到需求"`
	AppliedTime         *time.Time     `json:"appliedTime" gorm:"comment:应用时间"`
	
	// AI分析信息
	AIModel             string         `json:"aiModel" gorm:"type:varchar(50);comment:使用的AI模型"`
	AnalysisConfidence  float64        `json:"analysisConfidence" gorm:"comment:分析置信度(0-1)"`
	GeneratedAt         time.Time      `json:"generatedAt" gorm:"not null;comment:生成时间"`
	
	// 关联关系
	Requirement         NesmaRequirement  `json:"requirement" gorm:"foreignKey:RequirementID"`
	Project             NesmaProject      `json:"project" gorm:"foreignKey:ProjectID"`
}

// TableName 自定义表名
func (NesmaRequirementOptimization) TableName() string {
	return "nesma_requirement_optimizations"
}

// IsAccepted 判断是否已被接受
func (ro *NesmaRequirementOptimization) IsAccepted() bool {
	return ro.UserDecision == "accepted"
}

// IsRejected 判断是否已被拒绝
func (ro *NesmaRequirementOptimization) IsRejected() bool {
	return ro.UserDecision == "rejected"
}

// IsPending 判断是否待决策
func (ro *NesmaRequirementOptimization) IsPending() bool {
	return ro.UserDecision == "pending" || ro.UserDecision == ""
}

// AcceptOptimization 接受优化建议
func (ro *NesmaRequirementOptimization) AcceptOptimization(reason string) {
	ro.UserDecision = "accepted"
	now := time.Now()
	ro.UserDecisionTime = &now
	ro.UserDecisionReason = reason
}

// RejectOptimization 拒绝优化建议
func (ro *NesmaRequirementOptimization) RejectOptimization(reason string) {
	ro.UserDecision = "rejected"
	now := time.Now()
	ro.UserDecisionTime = &now
	ro.UserDecisionReason = reason
}

// NesmaNESMAEvaluation NESMA评估记录
type NesmaNESMAEvaluation struct {
	global.GVA_MODEL
	ProjectID       uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	CycleID         uint           `json:"cycleId" gorm:"not null;comment:周期ID;index"`
	EvaluationType  string         `json:"evaluationType" gorm:"type:varchar(50);not null;comment:评估类型：comprehensive-全面评估,compliance-合规检查,benchmark-基准对比"`
	
	// 评估概要
	OverallGrade        string         `json:"overallGrade" gorm:"type:varchar(10);comment:总体评级"`
	TotalAFP            float64        `json:"totalAFP" gorm:"comment:总调整功能点数"`
	TotalUFP            float64        `json:"totalUFP" gorm:"comment:总未调整功能点数"`
	QualityScore        float64        `json:"qualityScore" gorm:"comment:质量评分(0-100)"`
	NESMACompliance     float64        `json:"nesmaCompliance" gorm:"comment:NESMA合规评分(0-100)"`
	IndustryRanking     string         `json:"industryRanking" gorm:"type:varchar(50);comment:行业排名"`
	InvestmentGrade     string         `json:"investmentGrade" gorm:"type:varchar(20);comment:投资级别"`
	
	// 功能点分析
	EICount             int            `json:"eiCount" gorm:"comment:外部输入数量"`
	EOCount             int            `json:"eoCount" gorm:"comment:外部输出数量"`
	EQCount             int            `json:"eqCount" gorm:"comment:外部查询数量"`
	ILFCount            int            `json:"ilfCount" gorm:"comment:内部逻辑文件数量"`
	EIFCount            int            `json:"eifCount" gorm:"comment:外部接口文件数量"`
	
	// 复杂度分布
	LowComplexityCount    int          `json:"lowComplexityCount" gorm:"comment:低复杂度数量"`
	AvgComplexityCount    int          `json:"avgComplexityCount" gorm:"comment:中等复杂度数量"`
	HighComplexityCount   int          `json:"highComplexityCount" gorm:"comment:高复杂度数量"`
	
	// 质量指标
	MeasurementConsistency   float64   `json:"measurementConsistency" gorm:"comment:测量一致性(0-100)"`
	DocumentationCompleteness float64 `json:"documentationCompleteness" gorm:"comment:文档完整性(0-100)"`
	TraceabilityScore       float64   `json:"traceabilityScore" gorm:"comment:可追溯性评分(0-100)"`
	
	// 评估详细结果
	EvaluationResult    datatypes.JSON `json:"evaluationResult" gorm:"type:json;comment:详细评估结果"`
	FunctionPointAnalysis datatypes.JSON `json:"functionPointAnalysis" gorm:"type:json;comment:功能点分析"`
	QualityEvaluation   datatypes.JSON `json:"qualityEvaluation" gorm:"type:json;comment:质量评估"`
	ComplianceCheck     datatypes.JSON `json:"complianceCheck" gorm:"type:json;comment:合规检查"`
	BenchmarkingAnalysis datatypes.JSON `json:"benchmarkingAnalysis" gorm:"type:json;comment:基准对比分析"`
	
	// 改进建议
	KeyFindings         datatypes.JSON `json:"keyFindings" gorm:"type:json;comment:关键发现"`
	CriticalRecommendations datatypes.JSON `json:"criticalRecommendations" gorm:"type:json;comment:关键建议"`
	ImprovementRoadmap  datatypes.JSON `json:"improvementRoadmap" gorm:"type:json;comment:改进路线图"`
	
	// 评估元信息
	EvaluatorModel      string         `json:"evaluatorModel" gorm:"type:varchar(50);comment:评估AI模型"`
	EvaluationStandard  string         `json:"evaluationStandard" gorm:"type:varchar(50);comment:评估标准"`
	ConfidenceLevel     float64        `json:"confidenceLevel" gorm:"comment:置信水平(0-1)"`
	EvaluationDate      time.Time      `json:"evaluationDate" gorm:"not null;comment:评估日期"`
	
	// 关联关系
	Project             NesmaProject      `json:"project" gorm:"foreignKey:ProjectID"`
	Cycle               NesmaProjectCycle `json:"cycle" gorm:"foreignKey:CycleID"`
}

// TableName 自定义表名
func (NesmaNESMAEvaluation) TableName() string {
	return "nesma_nesma_evaluations"
}

// GetGradeText 获取评级中文描述
func (ne *NesmaNESMAEvaluation) GetGradeText() string {
	gradeMap := map[string]string{
		"A+": "优秀+",
		"A":  "优秀",
		"A-": "优秀-",
		"B+": "良好+",
		"B":  "良好",
		"B-": "良好-",
		"C+": "一般+",
		"C":  "一般",
		"C-": "一般-",
		"D":  "较差",
		"F":  "不合格",
	}
	if text, exists := gradeMap[ne.OverallGrade]; exists {
		return text
	}
	return "未评级"
}

// GetInvestmentGradeText 获取投资级别中文描述
func (ne *NesmaNESMAEvaluation) GetInvestmentGradeText() string {
	gradeMap := map[string]string{
		"Excellent": "优秀投资",
		"Good":      "良好投资",
		"Fair":      "一般投资",
		"Poor":      "较差投资",
	}
	if text, exists := gradeMap[ne.InvestmentGrade]; exists {
		return text
	}
	return "未评级"
}

// CalculateFunctionPointDistribution 计算功能点分布
func (ne *NesmaNESMAEvaluation) CalculateFunctionPointDistribution() map[string]float64 {
	total := float64(ne.EICount + ne.EOCount + ne.EQCount + ne.ILFCount + ne.EIFCount)
	if total == 0 {
		return map[string]float64{
			"EI": 0, "EO": 0, "EQ": 0, "ILF": 0, "EIF": 0,
		}
	}
	
	return map[string]float64{
		"EI":  float64(ne.EICount) / total * 100,
		"EO":  float64(ne.EOCount) / total * 100,
		"EQ":  float64(ne.EQCount) / total * 100,
		"ILF": float64(ne.ILFCount) / total * 100,
		"EIF": float64(ne.EIFCount) / total * 100,
	}
}

// CalculateComplexityDistribution 计算复杂度分布
func (ne *NesmaNESMAEvaluation) CalculateComplexityDistribution() map[string]float64 {
	total := float64(ne.LowComplexityCount + ne.AvgComplexityCount + ne.HighComplexityCount)
	if total == 0 {
		return map[string]float64{
			"Low": 0, "Average": 0, "High": 0,
		}
	}
	
	return map[string]float64{
		"Low":     float64(ne.LowComplexityCount) / total * 100,
		"Average": float64(ne.AvgComplexityCount) / total * 100,
		"High":    float64(ne.HighComplexityCount) / total * 100,
	}
}

// NesmaOptimizationSuggestion 优化建议
type NesmaOptimizationSuggestion struct {
	global.GVA_MODEL
	ProjectID           uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	RequirementID       *uint          `json:"requirementId" gorm:"comment:需求ID(针对单个需求的建议);index"`
	SuggestionType      string         `json:"suggestionType" gorm:"type:varchar(50);not null;comment:建议类型：project-项目级,requirement-需求级,process-流程级"`
	Category            string         `json:"category" gorm:"type:varchar(50);not null;comment:建议类别：quality-质量改进,efficiency-效率提升,risk-风险控制,compliance-合规性"`
	
	// 建议内容
	Title               string         `json:"title" gorm:"type:varchar(200);not null;comment:建议标题"`
	Description         string         `json:"description" gorm:"type:text;not null;comment:建议描述"`
	Rationale           string         `json:"rationale" gorm:"type:text;comment:建议理由"`
	ExpectedOutcome     string         `json:"expectedOutcome" gorm:"type:text;comment:预期结果"`
	
	// 优先级和影响
	Priority            string         `json:"priority" gorm:"type:varchar(20);not null;comment:优先级：High-高,Medium-中,Low-低"`
	Impact              string         `json:"impact" gorm:"type:varchar(20);comment:影响程度：High-高,Medium-中,Low-低"`
	Effort              string         `json:"effort" gorm:"type:varchar(20);comment:实施工作量：High-高,Medium-中,Low-低"`
	
	// 实施信息
	ActionItems         datatypes.JSON `json:"actionItems" gorm:"type:json;comment:行动项列表"`
	ImplementationNotes string         `json:"implementationNotes" gorm:"type:text;comment:实施注意事项"`
	Timeline            string         `json:"timeline" gorm:"type:varchar(100);comment:时间线"`
	ResourceRequirements string        `json:"resourceRequirements" gorm:"type:text;comment:资源需求"`
	
	// 状态跟踪
	Status              string         `json:"status" gorm:"type:varchar(20);default:'pending';comment:状态：pending-待处理,in_progress-进行中,completed-已完成,rejected-已拒绝"`
	AssignedTo          string         `json:"assignedTo" gorm:"type:varchar(100);comment:指派给"`
	StartDate           *time.Time     `json:"startDate" gorm:"comment:开始日期"`
	DueDate             *time.Time     `json:"dueDate" gorm:"comment:截止日期"`
	CompletedDate       *time.Time     `json:"completedDate" gorm:"comment:完成日期"`
	
	// 效果评估
	ActualOutcome       string         `json:"actualOutcome" gorm:"type:text;comment:实际结果"`
	EffectivenessScore  *float64       `json:"effectivenessScore" gorm:"comment:有效性评分(0-100)"`
	
	// 生成信息
	GeneratedBy         string         `json:"generatedBy" gorm:"type:varchar(50);comment:生成方式：ai-AI生成,manual-手动添加"`
	AIModel             string         `json:"aiModel" gorm:"type:varchar(50);comment:AI模型"`
	ConfidenceScore     float64        `json:"confidenceScore" gorm:"comment:置信度(0-1)"`
	
	// 关联关系
	Project             NesmaProject     `json:"project" gorm:"foreignKey:ProjectID"`
	Requirement         *NesmaRequirement `json:"requirement" gorm:"foreignKey:RequirementID"`
}

// TableName 自定义表名
func (NesmaOptimizationSuggestion) TableName() string {
	return "nesma_optimization_suggestions"
}

// GetPriorityText 获取优先级中文描述
func (os *NesmaOptimizationSuggestion) GetPriorityText() string {
	priorityMap := map[string]string{
		"High":   "高",
		"Medium": "中",
		"Low":    "低",
	}
	if text, exists := priorityMap[os.Priority]; exists {
		return text
	}
	return "未设置"
}

// GetStatusText 获取状态中文描述
func (os *NesmaOptimizationSuggestion) GetStatusText() string {
	statusMap := map[string]string{
		"pending":     "待处理",
		"in_progress": "进行中",
		"completed":   "已完成",
		"rejected":    "已拒绝",
	}
	if text, exists := statusMap[os.Status]; exists {
		return text
	}
	return "未知状态"
}

// IsOverdue 判断是否逾期
func (os *NesmaOptimizationSuggestion) IsOverdue() bool {
	if os.DueDate == nil || os.Status == "completed" || os.Status == "rejected" {
		return false
	}
	return time.Now().After(*os.DueDate)
}

// MarkAsCompleted 标记为已完成
func (os *NesmaOptimizationSuggestion) MarkAsCompleted(outcome string, effectivenessScore float64) {
	os.Status = "completed"
	now := time.Now()
	os.CompletedDate = &now
	os.ActualOutcome = outcome
	os.EffectivenessScore = &effectivenessScore
}

// MarkAsRejected 标记为已拒绝
func (os *NesmaOptimizationSuggestion) MarkAsRejected(reason string) {
	os.Status = "rejected"
	os.ActualOutcome = reason
}

// NesmaAnalysisHistory 分析历史记录
type NesmaAnalysisHistory struct {
	global.GVA_MODEL
	ProjectID       uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	AnalysisType    string         `json:"analysisType" gorm:"type:varchar(50);not null;comment:分析类型"`
	AnalysisDate    time.Time      `json:"analysisDate" gorm:"not null;comment:分析日期;index"`
	
	// 快照数据
	QualityScore    float64        `json:"qualityScore" gorm:"comment:质量评分"`
	FunctionPoints  float64        `json:"functionPoints" gorm:"comment:功能点数"`
	RequirementCount int           `json:"requirementCount" gorm:"comment:需求数量"`
	ComplexityIndex float64        `json:"complexityIndex" gorm:"comment:复杂度指数"`
	
	// 变化对比
	QualityDelta    float64        `json:"qualityDelta" gorm:"comment:质量变化"`
	FPDelta         float64        `json:"fpDelta" gorm:"comment:功能点变化"`
	RequirementDelta int           `json:"requirementDelta" gorm:"comment:需求数变化"`
	
	// 关联的分析记录ID
	ProjectAnalysisID   *uint      `json:"projectAnalysisId" gorm:"comment:项目分析ID"`
	EvaluationID        *uint      `json:"evaluationId" gorm:"comment:评估ID"`
	
	// 关联关系
	Project         NesmaProject      `json:"project" gorm:"foreignKey:ProjectID"`
}

// TableName 自定义表名
func (NesmaAnalysisHistory) TableName() string {
	return "nesma_analysis_histories"
}