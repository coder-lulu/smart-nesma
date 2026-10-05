package request

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// CreateEvaluationRequest 创建评估请求
type CreateEvaluationRequest struct {
	ProjectID              uint                   `json:"projectId" binding:"required"`
	CycleID                *uint                  `json:"cycleId"`
	RequirementVersionID   *uint                  `json:"requirementVersionId"`
	EvaluationName         string                 `json:"evaluationName" binding:"required"`
	EvaluationVersion      string                 `json:"evaluationVersion" binding:"required"`
	EvaluationType         string                 `json:"evaluationType" binding:"required"`
	EvaluatorID            uint                   `json:"evaluatorId" binding:"required"`
	NesmaRules             string                 `json:"nesmaRules"`
	EvaluationConfig       map[string]interface{} `json:"evaluationConfig"`
	Description            string                 `json:"description"`
}

// UpdateEvaluationRequest 更新评估请求
type UpdateEvaluationRequest struct {
	ID                uint                   `json:"id" binding:"required"`
	EvaluationName    string                 `json:"evaluationName"`
	EvaluationVersion string                 `json:"evaluationVersion"`
	EvaluationType    string                 `json:"evaluationType"`
	Status            string                 `json:"status"`
	EvaluationConfig  map[string]interface{} `json:"evaluationConfig"`
}

// EvaluationSearchRequest 评估搜索请求
type EvaluationSearchRequest struct {
	ProjectID              uint       `json:"projectId"`
	CycleID                *uint      `json:"cycleId"`
	RequirementVersionID   *uint      `json:"requirementVersionId"`
	EvaluatorID            uint       `json:"evaluatorId"`
	Status                 string     `json:"status"`
	EvaluationType         string     `json:"evaluationType"`
	StartTime              *time.Time `json:"startTime"`
	EndTime                *time.Time `json:"endTime"`
	request.PageInfo
}

// StartEvaluationRequest 开始评估请求
type StartEvaluationRequest struct {
	EvaluationID uint `json:"evaluationId" binding:"required"`
}

// ReviewEvaluationRequest 评估审核请求
type ReviewEvaluationRequest struct {
	EvaluationID   uint   `json:"evaluationId" binding:"required"`
	ReviewStatus   string `json:"reviewStatus" binding:"required"`
	ReviewerID     uint   `json:"reviewerId" binding:"required"`
	ReviewComments string `json:"reviewComments"`
}

// FunctionPointRequest 功能点请求
type FunctionPointRequest struct {
	EvaluationID         uint    `json:"evaluationId" binding:"required"`
	RequirementID        *uint   `json:"requirementId"`
	FunctionType         string  `json:"functionType" binding:"required"`
	FunctionName         string  `json:"functionName" binding:"required"`
	FunctionDesc         string  `json:"functionDesc"`
	DataElements         int     `json:"dataElements"`
	FileTypes            int     `json:"fileTypes"`
	RecordElements       int     `json:"recordElements"`
	ComplexityLevel      string  `json:"complexityLevel" binding:"required"`
	WeightFactor         float64 `json:"weightFactor"`
	IdentificationMethod string  `json:"identificationMethod"`
	ConfidenceLevel      float64 `json:"confidenceLevel"`
	IsValidated          bool    `json:"isValidated"`
	ValidationNotes      string  `json:"validationNotes"`
}

// ValidationItemRequest 验证项请求
type ValidationItemRequest struct {
	EvaluationID     uint   `json:"evaluationId" binding:"required"`
	ValidationRule   string `json:"validationRule" binding:"required"`
	RuleDescription  string `json:"ruleDescription"`
	ValidationResult string `json:"validationResult"`
	ExpectedValue    string `json:"expectedValue"`
	ActualValue      string `json:"actualValue"`
	DeviationLevel   string `json:"deviationLevel"`
	ValidationNotes  string `json:"validationNotes"`
	IsCritical       bool   `json:"isCritical"`
	ResolutionStatus string `json:"resolutionStatus"`
	ResolutionNotes  string `json:"resolutionNotes"`
}

// ComplexityMetricRequest 复杂度指标请求
type ComplexityMetricRequest struct {
	EvaluationID       uint    `json:"evaluationId" binding:"required"`
	MetricType         string  `json:"metricType" binding:"required"`
	MetricName         string  `json:"metricName" binding:"required"`
	MetricValue        float64 `json:"metricValue"`
	ThresholdValue     float64 `json:"thresholdValue"`
	Score              float64 `json:"score"`
	Weight             float64 `json:"weight"`
	CalculationFormula string  `json:"calculationFormula"`
	QualityIndicator   string  `json:"qualityIndicator"`
	ImpactLevel        string  `json:"impactLevel"`
}
