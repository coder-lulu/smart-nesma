package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaEvaluation NESMA评估记录表
type NesmaEvaluation struct {
	global.GVA_MODEL
	ProjectID              uint   `json:"projectId" gorm:"not null;comment:项目ID"`
	CycleID                *uint  `json:"cycleId" gorm:"comment:项目周期ID;index"`
	RequirementVersionID   *uint  `json:"requirementVersionId" gorm:"comment:需求版本ID;index"`
	AIAnalysisID           *uint  `json:"aiAnalysisId" gorm:"comment:AI分析ID;index"`
	EvaluationName         string `json:"evaluationName" gorm:"type:varchar(200);not null;comment:评估名称"`
	EvaluationVersion      string `json:"evaluationVersion" gorm:"type:varchar(50);not null;comment:评估版本"`
	EvaluationType         string `json:"evaluationType" gorm:"type:varchar(50);not null;comment:评估类型:initial,detailed,final"`
	Status                 string `json:"status" gorm:"type:varchar(50);default:'pending';comment:评估状态:pending,processing,completed,failed"`
	
	// 进度追踪
	Progress               float64 `json:"progress" gorm:"default:0;comment:评估进度百分比0-100"`
	CurrentPhase           string  `json:"currentPhase" gorm:"type:varchar(100);comment:当前阶段"`
	TotalSteps             int     `json:"totalSteps" gorm:"default:0;comment:总步骤数"`
	CompletedSteps         int     `json:"completedSteps" gorm:"default:0;comment:已完成步骤数"`
	ProcessedRequirements  int     `json:"processedRequirements" gorm:"default:0;comment:已处理需求数"`
	TotalRequirements      int     `json:"totalRequirements" gorm:"default:0;comment:总需求数"`
	ProgressMessage        string  `json:"progressMessage" gorm:"type:varchar(500);comment:进度消息"`

	// 评估结果统计
	TotalFunctionPoints    float64 `json:"totalFunctionPoints" gorm:"column:total_function_points;comment:总功能点数"`
	DataFunctionPoints     float64 `json:"dataFunctionPoints" gorm:"column:data_function_points;comment:数据功能点数"`
	TransactionalFP        float64 `json:"transactionalFP" gorm:"column:transactional_fp;comment:事务功能点数"`
	AdjustmentFactor       float64 `json:"adjustmentFactor" gorm:"default:1.0;comment:调整因子"`
	AdjustedFunctionPoints float64 `json:"adjustedFunctionPoints" gorm:"comment:调整后功能点数"`

	// 复杂度统计
	SimpleFunctionCount  int `json:"simpleFunctionCount" gorm:"comment:简单功能数量"`
	AverageFunctionCount int `json:"averageFunctionCount" gorm:"comment:平均功能数量"`
	ComplexFunctionCount int `json:"complexFunctionCount" gorm:"comment:复杂功能数量"`

	// 评估配置
	EvaluationConfig datatypes.JSON `json:"evaluationConfig" gorm:"type:json;comment:评估配置参数"`
	NesmaRules       string         `json:"nesmaRules" gorm:"type:varchar(50);default:'v2.2';comment:NESMA规则版本"`

	// 质量指标
	ConfidenceScore float64 `json:"confidenceScore" gorm:"column:confidence_score;comment:评估置信度分数"`
	AccuracyScore   float64 `json:"accuracyScore" gorm:"column:accuracy_score;comment:评估准确度分数"`
	ComplianceScore float64 `json:"complianceScore" gorm:"column:compliance_score;comment:标准合规度分数"`
	QualityScore    float64 `json:"qualityScore" gorm:"column:quality_score;comment:质量评分"`
	
	// 评估结果
	OverallGrade        string `json:"overallGrade" gorm:"column:overall_grade;type:varchar(50);comment:总体评级"`
	ComplexityLevel     string `json:"complexityLevel" gorm:"column:complexity_level;type:varchar(50);comment:复杂度级别"`
	RiskLevel           string `json:"riskLevel" gorm:"column:risk_level;type:varchar(50);comment:风险级别"`
	EvaluationSummary   string `json:"evaluationSummary" gorm:"column:evaluation_summary;type:text;comment:评估总结"`

	// 评估元数据
	EvaluatorID    uint       `json:"evaluatorId" gorm:"comment:评估人员ID"`
	StartTime      *time.Time `json:"startTime" gorm:"comment:评估开始时间"`
	CompletionTime *time.Time `json:"completionTime" gorm:"comment:评估完成时间"`
	Duration       int        `json:"duration" gorm:"comment:评估耗时(秒)"`

	// 详细结果
	EvaluationDetails  datatypes.JSON `json:"evaluationDetails" gorm:"type:json;comment:详细评估结果"`
	ValidationResults  datatypes.JSON `json:"validationResults" gorm:"type:json;comment:验证结果"`
	RecommendationList datatypes.JSON `json:"recommendationList" gorm:"type:json;comment:改进建议列表"`

	// 错误和异常处理
	ErrorMessage   string     `json:"errorMessage" gorm:"type:text;comment:错误信息"`
	ErrorCode      string     `json:"errorCode" gorm:"type:varchar(50);comment:错误代码"`
	ErrorDetails   string     `json:"errorDetails" gorm:"type:text;comment:错误详情"`
	
	// 审核信息
	ReviewStatus   string     `json:"reviewStatus" gorm:"type:varchar(50);default:'pending';comment:审核状态"`
	ReviewerID     *uint      `json:"reviewerId" gorm:"comment:审核人员ID"`
	ReviewComments string     `json:"reviewComments" gorm:"type:text;comment:审核意见"`
	ReviewTime     *time.Time `json:"reviewTime" gorm:"comment:审核时间"`

	// 关联关系
	Project             NesmaProject                `json:"project" gorm:"foreignKey:ProjectID"`
	Cycle               *NesmaProjectCycle          `json:"cycle" gorm:"foreignKey:CycleID"`
	RequirementVersion  *NesmaRequirementVersion    `json:"requirementVersion" gorm:"foreignKey:RequirementVersionID"`
	FunctionPoints      []NesmaFunctionPoint       `json:"functionPoints" gorm:"foreignKey:EvaluationID"`
	ComplexityMetrics   []NesmaComplexityMetric    `json:"complexityMetrics" gorm:"foreignKey:EvaluationID"`
	ValidationItems     []NesmaValidationItem      `json:"validationItems" gorm:"foreignKey:EvaluationID"`
}

// NesmaFunctionPoint 功能点详细记录表
type NesmaFunctionPoint struct {
	global.GVA_MODEL
	EvaluationID  uint   `json:"evaluationId" gorm:"not null;comment:评估ID"`
	RequirementID *uint  `json:"requirementId" gorm:"comment:关联需求ID"`
	FunctionType  string `json:"functionType" gorm:"type:varchar(50);not null;comment:功能类型:ILF,EIF,EI,EO,EQ"`
	FunctionName  string `json:"functionName" gorm:"type:varchar(200);not null;comment:功能名称"`
	FunctionDesc  string `json:"functionDesc" gorm:"type:text;comment:功能描述"`

	// 计算要素
	DataElements   int `json:"dataElements" gorm:"comment:数据元素数"`
	FileTypes      int `json:"fileTypes" gorm:"comment:文件类型数"`
	RecordElements int `json:"recordElements" gorm:"comment:记录元素数"`

	// 复杂度和权重
	ComplexityLevel  string  `json:"complexityLevel" gorm:"type:varchar(20);not null;comment:复杂度级别:Low,Average,High"`
	WeightFactor     float64 `json:"weightFactor" gorm:"not null;comment:权重因子"`
	CalculatedPoints float64 `json:"calculatedPoints" gorm:"not null;comment:计算得出的功能点数"`

	// 识别信息
	IdentificationMethod string         `json:"identificationMethod" gorm:"type:varchar(50);comment:识别方法:manual,auto,ai_assisted"`
	ConfidenceLevel      float64        `json:"confidenceLevel" gorm:"comment:识别置信度"`
	DetectionRules       datatypes.JSON `json:"detectionRules" gorm:"type:json;comment:检测规则"`

	// 验证信息
	IsValidated      bool   `json:"isValidated" gorm:"default:false;comment:是否已验证"`
	ValidationStatus string `json:"validationStatus" gorm:"type:varchar(50);default:'pending';comment:验证状态"`
	ValidationNotes  string `json:"validationNotes" gorm:"type:text;comment:验证备注"`

	// 关联关系
	Evaluation  NesmaEvaluation   `json:"evaluation" gorm:"foreignKey:EvaluationID"`
	Requirement *NesmaRequirement `json:"requirement" gorm:"foreignKey:RequirementID"`
}

// NesmaComplexityMetric 复杂度指标表
type NesmaComplexityMetric struct {
	global.GVA_MODEL
	EvaluationID   uint    `json:"evaluationId" gorm:"not null;comment:评估ID"`
	MetricType     string  `json:"metricType" gorm:"type:varchar(100);not null;comment:指标类型"`
	MetricName     string  `json:"metricName" gorm:"type:varchar(200);not null;comment:指标名称"`
	MetricValue    float64 `json:"metricValue" gorm:"not null;comment:指标值"`
	ThresholdValue float64 `json:"thresholdValue" gorm:"comment:阈值"`
	Score          float64 `json:"score" gorm:"comment:评分"`
	Weight         float64 `json:"weight" gorm:"default:1.0;comment:权重"`

	// 计算信息
	CalculationFormula string         `json:"calculationFormula" gorm:"type:varchar(500);comment:计算公式"`
	CalculationDetails datatypes.JSON `json:"calculationDetails" gorm:"type:json;comment:计算详情"`

	// 质量评估
	QualityIndicator string `json:"qualityIndicator" gorm:"type:varchar(50);comment:质量指示器:excellent,good,average,poor"`
	ImpactLevel      string `json:"impactLevel" gorm:"type:varchar(50);comment:影响级别:low,medium,high,critical"`

	// 关联关系
	Evaluation NesmaEvaluation `json:"evaluation" gorm:"foreignKey:EvaluationID"`
}

// NesmaValidationItem 验证项目表
type NesmaValidationItem struct {
	global.GVA_MODEL
	EvaluationID     uint   `json:"evaluationId" gorm:"not null;comment:评估ID"`
	ValidationRule   string `json:"validationRule" gorm:"type:varchar(200);not null;comment:验证规则"`
	RuleDescription  string `json:"ruleDescription" gorm:"type:text;comment:规则描述"`
	ValidationResult string `json:"validationResult" gorm:"type:varchar(50);not null;comment:验证结果:pass,fail,warning,skip"`

	// 验证详情
	ExpectedValue     string         `json:"expectedValue" gorm:"type:varchar(200);comment:期望值"`
	ActualValue       string         `json:"actualValue" gorm:"type:varchar(200);comment:实际值"`
	DeviationLevel    string         `json:"deviationLevel" gorm:"type:varchar(50);comment:偏差级别"`
	ValidationNotes   string         `json:"validationNotes" gorm:"type:text;comment:验证备注"`
	ValidationDetails datatypes.JSON `json:"validationDetails" gorm:"type:json;comment:验证详情"`

	// 处理信息
	IsCritical       bool   `json:"isCritical" gorm:"default:false;comment:是否关键问题"`
	ResolutionStatus string `json:"resolutionStatus" gorm:"type:varchar(50);default:'open';comment:解决状态"`
	ResolutionNotes  string `json:"resolutionNotes" gorm:"type:text;comment:解决备注"`

	// 关联关系
	Evaluation NesmaEvaluation `json:"evaluation" gorm:"foreignKey:EvaluationID"`
}

// NesmaEvaluationHistory 评估历史对比表
type NesmaEvaluationHistory struct {
	global.GVA_MODEL
	ProjectID      uint `json:"projectId" gorm:"not null;comment:项目ID"`
	CurrentEvalID  uint `json:"currentEvalId" gorm:"not null;comment:当前评估ID"`
	PreviousEvalID uint `json:"previousEvalId" gorm:"not null;comment:之前评估ID"`

	// 变化指标
	FPDifference     float64 `json:"fpDifference" gorm:"comment:功能点差异"`
	FPChangePercent  float64 `json:"fpChangePercent" gorm:"comment:功能点变化百分比"`
	ComplexityChange float64 `json:"complexityChange" gorm:"comment:复杂度变化"`

	// 分析结果
	ChangeCategory string         `json:"changeCategory" gorm:"type:varchar(100);comment:变化类别"`
	ChangeSummary  string         `json:"changeSummary" gorm:"type:text;comment:变化总结"`
	ChangeReasons  datatypes.JSON `json:"changeReasons" gorm:"type:json;comment:变化原因"`
	ImpactAnalysis datatypes.JSON `json:"impactAnalysis" gorm:"type:json;comment:影响分析"`

	// 关联关系
	Project      NesmaProject    `json:"project" gorm:"foreignKey:ProjectID"`
	CurrentEval  NesmaEvaluation `json:"currentEval" gorm:"foreignKey:CurrentEvalID"`
	PreviousEval NesmaEvaluation `json:"previousEval" gorm:"foreignKey:PreviousEvalID"`
}

// TableName 自定义表名
func (NesmaEvaluation) TableName() string {
	return "nesma_evaluations"
}

func (NesmaFunctionPoint) TableName() string {
	return "nesma_function_points"
}

func (NesmaComplexityMetric) TableName() string {
	return "nesma_complexity_metrics"
}

func (NesmaValidationItem) TableName() string {
	return "nesma_validation_items"
}

func (NesmaEvaluationHistory) TableName() string {
	return "nesma_evaluation_history"
}

// GetComplexityLevel 根据权重因子获取复杂度级别
func (fp *NesmaFunctionPoint) GetComplexityLevel() string {
	switch fp.FunctionType {
	case "ILF", "EIF": // 内部逻辑文件, 外部接口文件
		if fp.WeightFactor <= 7 {
			return "Low"
		} else if fp.WeightFactor <= 10 {
			return "Average"
		} else {
			return "High"
		}
	case "EI": // 外部输入
		if fp.WeightFactor <= 3 {
			return "Low"
		} else if fp.WeightFactor <= 4 {
			return "Average"
		} else {
			return "High"
		}
	case "EO", "EQ": // 外部输出, 外部查询
		if fp.WeightFactor <= 4 {
			return "Low"
		} else if fp.WeightFactor <= 5 {
			return "Average"
		} else {
			return "High"
		}
	}
	return "Average"
}

// UpdateProgress 更新评估进度
func (e *NesmaEvaluation) UpdateProgress() {
	if e.TotalSteps > 0 {
		e.Progress = float64(e.CompletedSteps) / float64(e.TotalSteps) * 100
	} else if e.TotalRequirements > 0 {
		e.Progress = float64(e.ProcessedRequirements) / float64(e.TotalRequirements) * 100
	} else {
		e.Progress = 0
	}
	
	// 确保进度在0-100范围内
	if e.Progress < 0 {
		e.Progress = 0
	} else if e.Progress > 100 {
		e.Progress = 100
	}
}

// IsCompleted 检查评估是否已完成
func (e *NesmaEvaluation) IsCompleted() bool {
	return e.Status == "completed" || e.Progress >= 100
}

// IsProcessing 检查评估是否正在处理中
func (e *NesmaEvaluation) IsProcessing() bool {
	return e.Status == "processing"
}

// GetProgressPhases 获取评估阶段列表
func (e *NesmaEvaluation) GetProgressPhases() []string {
	return []string{
		"初始化评估",
		"获取需求数据",
		"功能点识别",
		"复杂度分析",
		"权重计算",
		"调整因子计算", 
		"验证检查",
		"结果汇总",
		"完成评估",
	}
}

// GetCurrentPhaseIndex 获取当前阶段索引
func (e *NesmaEvaluation) GetCurrentPhaseIndex() int {
	phases := e.GetProgressPhases()
	for i, phase := range phases {
		if phase == e.CurrentPhase {
			return i
		}
	}
	return 0
}

// SetPhase 设置当前阶段
func (e *NesmaEvaluation) SetPhase(phase string) {
	e.CurrentPhase = phase
	phases := e.GetProgressPhases()
	for i, p := range phases {
		if p == phase {
			e.CompletedSteps = i
			e.TotalSteps = len(phases)
			break
		}
	}
	e.UpdateProgress()
}

// CalculateFunctionPoints 计算功能点数
func (fp *NesmaFunctionPoint) CalculateFunctionPoints() float64 {
	// NESMA权重表
	weights := map[string]map[string]float64{
		"ILF": {"Low": 7, "Average": 10, "High": 15}, // 内部逻辑文件
		"EIF": {"Low": 5, "Average": 7, "High": 10},  // 外部接口文件
		"EI":  {"Low": 3, "Average": 4, "High": 6},   // 外部输入
		"EO":  {"Low": 4, "Average": 5, "High": 7},   // 外部输出
		"EQ":  {"Low": 3, "Average": 4, "High": 6},   // 外部查询
	}

	if functionWeights, exists := weights[fp.FunctionType]; exists {
		if weight, exists := functionWeights[fp.ComplexityLevel]; exists {
			fp.WeightFactor = weight
			fp.CalculatedPoints = weight
			return weight
		}
	}

	return 0
}

// GetQualityIndicator 获取质量指示器
func (cm *NesmaComplexityMetric) GetQualityIndicator() string {
	if cm.Score >= 0.9 {
		return "excellent"
	} else if cm.Score >= 0.7 {
		return "good"
	} else if cm.Score >= 0.5 {
		return "average"
	} else {
		return "poor"
	}
}

// IsPass 验证是否通过
func (vi *NesmaValidationItem) IsPass() bool {
	return vi.ValidationResult == "pass"
}

// GetSeverityLevel 获取严重级别
func (vi *NesmaValidationItem) GetSeverityLevel() string {
	if vi.IsCritical {
		return "critical"
	}

	switch vi.ValidationResult {
	case "fail":
		return "high"
	case "warning":
		return "medium"
	case "pass":
		return "low"
	default:
		return "unknown"
	}
}
