package response

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
)

// EvaluationResponse 评估响应
type EvaluationResponse struct {
	ID                     uint       `json:"id"`
	ProjectID              uint       `json:"projectId"`
	EvaluationName         string     `json:"evaluationName"`
	EvaluationVersion      string     `json:"evaluationVersion"`
	EvaluationType         string     `json:"evaluationType"`
	Status                 string     `json:"status"`
	TotalFunctionPoints    float64    `json:"totalFunctionPoints"`
	DataFunctionPoints     float64    `json:"dataFunctionPoints"`
	TransactionalFP        float64    `json:"transactionalFP"`
	AdjustedFunctionPoints float64    `json:"adjustedFunctionPoints"`
	AdjustmentFactor       float64    `json:"adjustmentFactor"`
	SimpleFunctionCount    int        `json:"simpleFunctionCount"`
	AverageFunctionCount   int        `json:"averageFunctionCount"`
	ComplexFunctionCount   int        `json:"complexFunctionCount"`
	ConfidenceScore        float64    `json:"confidenceScore"`
	AccuracyScore          float64    `json:"accuracyScore"`
	ComplianceScore        float64    `json:"complianceScore"`
	EvaluatorID            uint       `json:"evaluatorId"`
	StartTime              *time.Time `json:"startTime"`
	CompletionTime         *time.Time `json:"completionTime"`
	Duration               int        `json:"duration"`
	ReviewStatus           string     `json:"reviewStatus"`
	ReviewerID             *uint      `json:"reviewerId"`
	ReviewComments         string     `json:"reviewComments"`
	ReviewTime             *time.Time `json:"reviewTime"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`

	// 前端需要的关联字段
	ProjectName            string     `json:"projectName"`
	EvaluatorName          string     `json:"evaluatorName"`
	CycleName              string     `json:"cycleName"`
	RequirementVersionName string     `json:"requirementVersionName"`
	
	// 前端评估详情页面需要的字段
	NesmaRules             string     `json:"nesmaRules"`             // NESMA规则版本
	UnadjustedFunctionPoints float64  `json:"unadjustedFunctionPoints"` // 未调整功能点
	FunctionPointCount     int        `json:"functionPointCount"`    // 功能点条目数

	// 关联数据
	Project           *nesma.NesmaProject           `json:"project,omitempty"`
	FunctionPoints    []nesma.NesmaFunctionPoint    `json:"functionPoints,omitempty"`
	ComplexityMetrics []nesma.NesmaComplexityMetric `json:"complexityMetrics,omitempty"`
	ValidationItems   []nesma.NesmaValidationItem   `json:"validationItems,omitempty"`
}

// EvaluationListResponse 评估列表响应
type EvaluationListResponse struct {
	List     []EvaluationResponse `json:"list"`
	Total    int64                `json:"total"`
	Page     int                  `json:"page"`
	PageSize int                  `json:"pageSize"`
}

// FunctionPointResponse 功能点响应
type FunctionPointResponse struct {
	ID                   uint      `json:"id"`
	EvaluationID         uint      `json:"evaluationId"`
	RequirementID        *uint     `json:"requirementId"`
	FunctionType         string    `json:"functionType"`
	FunctionName         string    `json:"functionName"`
	FunctionDesc         string    `json:"functionDesc"`
	DataElements         int       `json:"dataElements"`
	FileTypes            int       `json:"fileTypes"`
	RecordElements       int       `json:"recordElements"`
	ComplexityLevel      string    `json:"complexityLevel"`
	WeightFactor         float64   `json:"weightFactor"`
	CalculatedPoints     float64   `json:"calculatedPoints"`
	IdentificationMethod string    `json:"identificationMethod"`
	ConfidenceLevel      float64   `json:"confidenceLevel"`
	IsValidated          bool      `json:"isValidated"`
	ValidationStatus     string    `json:"validationStatus"`
	ValidationNotes      string    `json:"validationNotes"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`

	// 关联数据
	Evaluation  *nesma.NesmaEvaluation  `json:"evaluation,omitempty"`
	Requirement *nesma.NesmaRequirement `json:"requirement,omitempty"`
}

// ComplexityMetricResponse 复杂度指标响应
type ComplexityMetricResponse struct {
	ID                 uint      `json:"id"`
	EvaluationID       uint      `json:"evaluationId"`
	MetricType         string    `json:"metricType"`
	MetricName         string    `json:"metricName"`
	MetricValue        float64   `json:"metricValue"`
	ThresholdValue     float64   `json:"thresholdValue"`
	Score              float64   `json:"score"`
	Weight             float64   `json:"weight"`
	CalculationFormula string    `json:"calculationFormula"`
	QualityIndicator   string    `json:"qualityIndicator"`
	ImpactLevel        string    `json:"impactLevel"`
	CreatedAt          time.Time `json:"createdAt"`
	UpdatedAt          time.Time `json:"updatedAt"`
}

// ValidationItemResponse 验证项响应
type ValidationItemResponse struct {
	ID               uint      `json:"id"`
	EvaluationID     uint      `json:"evaluationId"`
	ValidationRule   string    `json:"validationRule"`
	RuleDescription  string    `json:"ruleDescription"`
	ValidationResult string    `json:"validationResult"`
	ExpectedValue    string    `json:"expectedValue"`
	ActualValue      string    `json:"actualValue"`
	DeviationLevel   string    `json:"deviationLevel"`
	ValidationNotes  string    `json:"validationNotes"`
	IsCritical       bool      `json:"isCritical"`
	ResolutionStatus string    `json:"resolutionStatus"`
	ResolutionNotes  string    `json:"resolutionNotes"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// EvaluationSummaryResponse 评估摘要响应
type EvaluationSummaryResponse struct {
	ProjectID              uint       `json:"projectId"`
	ProjectName            string     `json:"projectName"`
	TotalEvaluations       int        `json:"totalEvaluations"`
	CompletedEvaluations   int        `json:"completedEvaluations"`
	AverageFunctionPoints  float64    `json:"averageFunctionPoints"`
	LatestEvaluationID     uint       `json:"latestEvaluationId"`
	LatestEvaluationName   string     `json:"latestEvaluationName"`
	LatestEvaluationTime   *time.Time `json:"latestEvaluationTime"`
	AverageConfidenceScore float64    `json:"averageConfidenceScore"`
	AverageAccuracyScore   float64    `json:"averageAccuracyScore"`
	AverageComplianceScore float64    `json:"averageComplianceScore"`
}

// EvaluationStatsResponse 评估统计响应
type EvaluationStatsResponse struct {
	TotalFunctionPoints       float64 `json:"totalFunctionPoints"`
	DataFunctionPointsRatio   float64 `json:"dataFunctionPointsRatio"`
	TransactionalFPRatio      float64 `json:"transactionalFPRatio"`
	AverageComplexityScore    float64 `json:"averageComplexityScore"`
	HighConfidencePointsRatio float64 `json:"highConfidencePointsRatio"`
	ValidationPassRate        float64 `json:"validationPassRate"`
	CriticalIssuesCount       int     `json:"criticalIssuesCount"`

	// 功能点类型分布
	ILFCount int `json:"ilfCount"`
	EIFCount int `json:"eifCount"`
	EICount  int `json:"eiCount"`
	EOCount  int `json:"eoCount"`
	EQCount  int `json:"eqCount"`

	// 复杂度级别分布
	LowComplexityCount     int `json:"lowComplexityCount"`
	AverageComplexityCount int `json:"averageComplexityCount"`
	HighComplexityCount    int `json:"highComplexityCount"`
}
