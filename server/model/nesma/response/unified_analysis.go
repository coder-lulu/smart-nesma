package response

import (
	"time"
)

// UnifiedAnalysisResponse 统一分析响应
type UnifiedAnalysisResponse struct {
	TaskID      uint                   `json:"task_id"`
	Type        string                 `json:"type"`
	Status      string                 `json:"status"`
	Progress    int                    `json:"progress"`
	Result      *AnalysisResult        `json:"result,omitempty"`
	Message     string                 `json:"message"`
	Error       string                 `json:"error,omitempty"`
	StartTime   *time.Time             `json:"start_time,omitempty"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Duration    int64                  `json:"duration,omitempty"` // 毫秒
	Metadata    *AnalysisMetadata      `json:"metadata,omitempty"`
}

// AnalysisResult 分析结果详情
type AnalysisResult struct {
	Type                string                    `json:"type"`
	ProjectAnalysis     *ProjectAnalysisResult    `json:"project_analysis,omitempty"`
	RequirementOptimize *RequirementOptimizeResult `json:"requirement_optimize,omitempty"`
	NESMAEvaluation     *NESMAEvaluationResult    `json:"nesma_evaluation,omitempty"`
	Summary             *AnalysisSummary          `json:"summary"`
	Insights            []AnalysisInsight         `json:"insights,omitempty"`
	Recommendations     []AnalysisRecommendation  `json:"recommendations,omitempty"`
	QualityScore        *QualityScore             `json:"quality_score,omitempty"`
}

// ProjectAnalysisResult 项目分析结果
type ProjectAnalysisResult struct {
	ProjectID          uint                     `json:"project_id"`
	ProjectName        string                   `json:"project_name"`
	CycleID            uint                     `json:"cycle_id"`
	CycleName          string                   `json:"cycle_name"`
	TotalRequirements  int                      `json:"total_requirements"`
	AnalyzedCount      int                      `json:"analyzed_count"`
	RequirementsByLevel map[string]int          `json:"requirements_by_level"`
	FunctionPointAnalysis *FunctionPointAnalysis `json:"function_point_analysis,omitempty"`
	ComplexityDistribution map[string]int        `json:"complexity_distribution"`
	CoverageAnalysis   *CoverageAnalysis        `json:"coverage_analysis,omitempty"`
	RiskAssessment     *RiskAssessment          `json:"risk_assessment,omitempty"`
	EstimatedEffort    *EstimatedEffort         `json:"estimated_effort,omitempty"`
}

// RequirementOptimizeResult 需求优化结果  
type RequirementOptimizeResult struct {
	RequirementID      uint                      `json:"requirement_id"`
	OriginalTitle      string                    `json:"original_title"`
	OriginalDescription string                   `json:"original_description"`
	OptimizedTitle     string                    `json:"optimized_title,omitempty"`
	OptimizedDescription string                  `json:"optimized_description,omitempty"`
	OptimizationTypes  []string                  `json:"optimization_types"`
	ImprovementAreas   []ImprovementArea         `json:"improvement_areas"`
	QualityMetrics     *RequirementQualityMetrics `json:"quality_metrics,omitempty"`
	ValidationResults  []ValidationResult        `json:"validation_results,omitempty"`
	ConfidenceScore    float64                   `json:"confidence_score"`
	AppliedChanges     []AppliedChange           `json:"applied_changes,omitempty"`
}

// NESMAEvaluationResult NESMA评估结果
type NESMAEvaluationResult struct {
	ProjectID           uint                      `json:"project_id"`
	CycleID             uint                      `json:"cycle_id"`
	TotalFunctionPoints float64                   `json:"total_function_points"`
	UnadjustedFP        float64                   `json:"unadjusted_fp"`
	AdjustedFP          float64                   `json:"adjusted_fp"`
	AdjustmentFactor    float64                   `json:"adjustment_factor"`
	FunctionTypes       map[string]FunctionTypeAnalysis `json:"function_types"`
	ComplexityBreakdown *ComplexityBreakdown      `json:"complexity_breakdown,omitempty"`
	QualityIndicators   *QualityIndicators        `json:"quality_indicators,omitempty"`
	BenchmarkComparison *BenchmarkComparison      `json:"benchmark_comparison,omitempty"`
	ComplianceCheck     *ComplianceCheck          `json:"compliance_check,omitempty"`
	Recommendations     []NESMARecommendation     `json:"recommendations,omitempty"`
}

// AnalysisSummary 分析摘要
type AnalysisSummary struct {
	TotalItems         int                    `json:"total_items"`
	ProcessedItems     int                    `json:"processed_items"`
	SuccessRate        float64                `json:"success_rate"`
	AverageConfidence  float64                `json:"average_confidence"`
	ProcessingTime     int64                  `json:"processing_time"` // 毫秒
	KeyFindings        []string               `json:"key_findings"`
	OverallScore       float64                `json:"overall_score"`
	StatusDistribution map[string]int         `json:"status_distribution"`
}

// AnalysisInsight 分析洞察
type AnalysisInsight struct {
	Category    string  `json:"category"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"` // High/Medium/Low
	Priority    int     `json:"priority"`
	Confidence  float64 `json:"confidence"`
	Evidence    []string `json:"evidence,omitempty"`
	Actions     []string `json:"actions,omitempty"`
}

// AnalysisRecommendation 分析建议
type AnalysisRecommendation struct {
	ID              string                 `json:"id"`
	Type            string                 `json:"type"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Rationale       string                 `json:"rationale"`
	Priority        string                 `json:"priority"` // High/Medium/Low
	Impact          string                 `json:"impact"`
	Effort          string                 `json:"effort"`
	Timeline        string                 `json:"timeline"`
	AcceptanceRate  float64                `json:"acceptance_rate"`
	Implementation  *ImplementationGuide   `json:"implementation,omitempty"`
	Metrics         map[string]interface{} `json:"metrics,omitempty"`
	Tags            []string               `json:"tags,omitempty"`
}

// QualityScore 质量评分
type QualityScore struct {
	OverallScore    float64            `json:"overall_score"`
	CompletenessScore float64          `json:"completeness_score"`
	ClarityScore    float64            `json:"clarity_score"`
	ConsistencyScore float64           `json:"consistency_score"`
	TraceabilityScore float64          `json:"traceability_score"`
	TestabilityScore float64           `json:"testability_score"`
	ScoreBreakdown  map[string]float64 `json:"score_breakdown"`
	Improvements    []string           `json:"improvements,omitempty"`
}

// FunctionPointAnalysis 功能点分析
type FunctionPointAnalysis struct {
	TotalUFP        float64                    `json:"total_ufp"`
	TotalAFP        float64                    `json:"total_afp"`
	ByType          map[string]float64         `json:"by_type"`
	ByComplexity    map[string]float64         `json:"by_complexity"`
	Distribution    *FPDistribution            `json:"distribution,omitempty"`
	Trends          []FPTrend                  `json:"trends,omitempty"`
	Validation      *FPValidation              `json:"validation,omitempty"`
}

// CoverageAnalysis 覆盖度分析
type CoverageAnalysis struct {
	RequirementCoverage   float64            `json:"requirement_coverage"`
	FunctionalCoverage    float64            `json:"functional_coverage"`
	BusinessProcessCoverage float64          `json:"business_process_coverage"`
	TestCoverage          float64            `json:"test_coverage"`
	CoverageByLevel       map[string]float64 `json:"coverage_by_level"`
	GapAnalysis           []CoverageGap      `json:"gap_analysis,omitempty"`
}

// RiskAssessment 风险评估
type RiskAssessment struct {
	OverallRiskLevel    string           `json:"overall_risk_level"`
	RiskCategories      map[string]float64 `json:"risk_categories"`
	HighRiskItems       []RiskItem       `json:"high_risk_items,omitempty"`
	MitigationStrategies []Mitigation    `json:"mitigation_strategies,omitempty"`
	RiskMatrix          *RiskMatrix      `json:"risk_matrix,omitempty"`
}

// EstimatedEffort 工作量估算
type EstimatedEffort struct {
	TotalHours          float64            `json:"total_hours"`
	ByPhase             map[string]float64 `json:"by_phase"`
	ByComplexity        map[string]float64 `json:"by_complexity"`
	ByRole              map[string]float64 `json:"by_role"`
	Confidence          float64            `json:"confidence"`
	Assumptions         []string           `json:"assumptions,omitempty"`
	Contingency         float64            `json:"contingency"`
}

// ImprovementArea 改进区域
type ImprovementArea struct {
	Area            string   `json:"area"`
	CurrentScore    float64  `json:"current_score"`
	TargetScore     float64  `json:"target_score"`
	Importance      string   `json:"importance"`
	Suggestions     []string `json:"suggestions"`
	ExpectedImpact  string   `json:"expected_impact"`
}

// RequirementQualityMetrics 需求质量指标
type RequirementQualityMetrics struct {
	Clarity         float64 `json:"clarity"`
	Completeness    float64 `json:"completeness"`
	Consistency     float64 `json:"consistency"`
	Testability     float64 `json:"testability"`
	Traceability    float64 `json:"traceability"`
	Feasibility     float64 `json:"feasibility"`
	OverallQuality  float64 `json:"overall_quality"`
}

// ValidationResult 验证结果
type ValidationResult struct {
	Rule        string   `json:"rule"`
	Status      string   `json:"status"` // pass/fail/warning
	Message     string   `json:"message"`
	Suggestions []string `json:"suggestions,omitempty"`
}

// AppliedChange 应用的变更
type AppliedChange struct {
	Field       string    `json:"field"`
	OldValue    string    `json:"old_value"`
	NewValue    string    `json:"new_value"`
	Reason      string    `json:"reason"`
	AppliedAt   time.Time `json:"applied_at"`
	AppliedBy   string    `json:"applied_by"`
}

// FunctionTypeAnalysis 功能类型分析
type FunctionTypeAnalysis struct {
	Type        string             `json:"type"`
	Count       int                `json:"count"`
	TotalUFP    float64            `json:"total_ufp"`
	TotalAFP    float64            `json:"total_afp"`
	AvgComplexity float64          `json:"avg_complexity"`
	Distribution map[string]int    `json:"distribution"`
	Examples    []string           `json:"examples,omitempty"`
}

// ComplexityBreakdown 复杂度分解
type ComplexityBreakdown struct {
	Simple      *ComplexityLevel `json:"simple"`
	Average     *ComplexityLevel `json:"average"`
	Complex     *ComplexityLevel `json:"complex"`
	Distribution map[string]float64 `json:"distribution"`
}

// ComplexityLevel 复杂度级别
type ComplexityLevel struct {
	Count       int     `json:"count"`
	Percentage  float64 `json:"percentage"`
	TotalUFP    float64 `json:"total_ufp"`
	TotalAFP    float64 `json:"total_afp"`
}

// QualityIndicators 质量指标
type QualityIndicators struct {
	CompletenessIndex   float64 `json:"completeness_index"`
	ConsistencyIndex    float64 `json:"consistency_index"`
	TraceabilityIndex   float64 `json:"traceability_index"`
	TestabilityIndex    float64 `json:"testability_index"`
	OverallQualityIndex float64 `json:"overall_quality_index"`
}

// BenchmarkComparison 基准对比
type BenchmarkComparison struct {
	IndustryAverage     float64            `json:"industry_average"`
	ProjectScore        float64            `json:"project_score"`
	PerformanceRatio    float64            `json:"performance_ratio"`
	Percentile          float64            `json:"percentile"`
	ComparisonAreas     map[string]float64 `json:"comparison_areas"`
	BenchmarkSource     string             `json:"benchmark_source"`
}

// ComplianceCheck 合规性检查
type ComplianceCheck struct {
	Standard        string             `json:"standard"`
	OverallScore    float64            `json:"overall_score"`
	ComplianceLevel string             `json:"compliance_level"`
	CheckResults    []ComplianceResult `json:"check_results"`
	Recommendations []string           `json:"recommendations,omitempty"`
}

// ComplianceResult 合规性检查结果
type ComplianceResult struct {
	Criterion   string  `json:"criterion"`
	Status      string  `json:"status"` // compliant/non_compliant/partial
	Score       float64 `json:"score"`
	Description string  `json:"description"`
	Evidence    string  `json:"evidence,omitempty"`
}

// NESMARecommendation NESMA建议
type NESMARecommendation struct {
	ID              string   `json:"id"`
	Category        string   `json:"category"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	Impact          string   `json:"impact"`
	Priority        string   `json:"priority"`
	Actions         []string `json:"actions"`
	ExpectedOutcome string   `json:"expected_outcome"`
}

// AnalysisMetadata 分析元数据
type AnalysisMetadata struct {
	Version           string                 `json:"version"`
	AnalysisEngine    string                 `json:"analysis_engine"`
	AIModel           string                 `json:"ai_model"`
	Parameters        map[string]interface{} `json:"parameters,omitempty"`
	KnowledgeBase     *KnowledgeBaseInfo     `json:"knowledge_base,omitempty"`
	ProcessingStats   *ProcessingStats       `json:"processing_stats,omitempty"`
	QualityControls   []QualityControl       `json:"quality_controls,omitempty"`
}

// 辅助结构体
type FPDistribution struct {
	ByLevel map[string]float64 `json:"by_level"`
	ByType  map[string]float64 `json:"by_type"`
}

type FPTrend struct {
	Period string  `json:"period"`
	UFP    float64 `json:"ufp"`
	AFP    float64 `json:"afp"`
}

type FPValidation struct {
	Status    string   `json:"status"`
	Issues    []string `json:"issues,omitempty"`
	Warnings  []string `json:"warnings,omitempty"`
}

type CoverageGap struct {
	Area        string   `json:"area"`
	GapType     string   `json:"gap_type"`
	Description string   `json:"description"`
	Impact      string   `json:"impact"`
	Suggestions []string `json:"suggestions"`
}

type RiskItem struct {
	ID           string  `json:"id"`
	Category     string  `json:"category"`
	Description  string  `json:"description"`
	Probability  float64 `json:"probability"`
	Impact       float64 `json:"impact"`
	RiskScore    float64 `json:"risk_score"`
	Mitigation   string  `json:"mitigation,omitempty"`
}

type Mitigation struct {
	RiskID      string   `json:"risk_id"`
	Strategy    string   `json:"strategy"`
	Actions     []string `json:"actions"`
	Timeline    string   `json:"timeline"`
	Owner       string   `json:"owner,omitempty"`
	Effectiveness float64 `json:"effectiveness"`
}

type RiskMatrix struct {
	High   []RiskItem `json:"high"`
	Medium []RiskItem `json:"medium"`
	Low    []RiskItem `json:"low"`
}

type ImplementationGuide struct {
	Steps        []string               `json:"steps"`
	Prerequisites []string              `json:"prerequisites,omitempty"`
	Resources    []string               `json:"resources,omitempty"`
	Timeline     string                 `json:"timeline"`
	Checklist    []ChecklistItem        `json:"checklist,omitempty"`
	Examples     map[string]interface{} `json:"examples,omitempty"`
}

type ChecklistItem struct {
	Task        string `json:"task"`
	Completed   bool   `json:"completed"`
	Priority    string `json:"priority"`
	Owner       string `json:"owner,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
}

type KnowledgeBaseInfo struct {
	Version     string    `json:"version"`
	LastUpdate  time.Time `json:"last_update"`
	EntriesUsed int       `json:"entries_used"`
	Coverage    float64   `json:"coverage"`
}

type ProcessingStats struct {
	TotalRequests    int                    `json:"total_requests"`
	SuccessfulCalls  int                    `json:"successful_calls"`
	FailedCalls      int                    `json:"failed_calls"`
	AvgResponseTime  int64                  `json:"avg_response_time"` // 毫秒
	TokenUsage       *TokenUsage            `json:"token_usage,omitempty"`
	CacheHitRate     float64                `json:"cache_hit_rate"`
	RetryAttempts    int                    `json:"retry_attempts"`
}

type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type QualityControl struct {
	Rule        string    `json:"rule"`
	Status      string    `json:"status"`
	Score       float64   `json:"score"`
	Message     string    `json:"message"`
	AppliedAt   time.Time `json:"applied_at"`
}

// GetDetailedAnalysisResponse 获取详细分析响应请求
type GetDetailedAnalysisResponse struct {
	TaskID         uint     `json:"task_id" form:"task_id" binding:"required"`
	IncludeSections []string `json:"include_sections" form:"include_sections"` // 包含的章节
	DetailLevel    string   `json:"detail_level" form:"detail_level"`         // basic/standard/detailed
	Format         string   `json:"format" form:"format"`                     // json/summary/report
}