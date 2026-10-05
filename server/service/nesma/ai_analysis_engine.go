package nesma

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// AIAnalysisEngine AI分析引擎
type AIAnalysisEngine struct {
	aiService *AIService
	batchService *BatchEvaluationService
	contextBuilder *ProjectContextBuilder
}

// AIAnalysisRequest AI分析请求
type AIAnalysisRequest struct {
	ProjectID       uint                  `json:"projectId"`
	EvaluationID    uint                  `json:"evaluationId"`
	AnalysisType    string                `json:"analysisType"`    // comprehensive, focused, incremental
	FocusAreas      []string              `json:"focusAreas"`      // functional, technical, business, quality, risk
	AnalysisConfig  AnalysisConfiguration `json:"analysisConfig"`
	UserID          uint                  `json:"userId"`
}

// AIAnalysisResult AI分析结果
type AIAnalysisResult struct {
	AnalysisID      uint                      `json:"analysisId"`
	Status          string                    `json:"status"`
	Progress        float64                   `json:"progress"`
	Summary         *AnalysisSummary          `json:"summary"`
	FunctionalAnalysis    *AIFunctionalAnalysisResult    `json:"functionalAnalysis"`
	TechnicalAnalysis     *AITechnicalAnalysisResult     `json:"technicalAnalysis"`
	BusinessAnalysis      *AIBusinessAnalysisResult      `json:"businessAnalysis"`
	QualityAnalysis       *AIQualityAnalysisResult       `json:"qualityAnalysis"`
	RiskAnalysis          *AIRiskAnalysisResult          `json:"riskAnalysis"`
	RecommendationAnalysis []AIRecommendationResult `json:"recommendationAnalysis"`
	ComplianceAnalysis    *AIComplianceAnalysisResult    `json:"complianceAnalysis"`
	ExecutionMetrics      *ExecutionMetrics             `json:"executionMetrics"`
}

// AnalysisSummary 分析摘要
type AnalysisSummary struct {
	OverallScore        float64   `json:"overallScore"`
	ComplexityLevel     string    `json:"complexityLevel"`
	RiskLevel           string    `json:"riskLevel"`
	RecommendationLevel string    `json:"recommendationLevel"`
	KeyFindings         []string  `json:"keyFindings"`
	CriticalIssues      []string  `json:"criticalIssues"`
	MainRecommendations []string  `json:"mainRecommendations"`
	ExecutionTime       time.Duration `json:"executionTime"`
	ProcessedItems      int       `json:"processedItems"`
}

// ExecutionMetrics 执行指标
type ExecutionMetrics struct {
	TotalExecutionTime    time.Duration `json:"totalExecutionTime"`
	AICallCount          int           `json:"aiCallCount"`
	TotalTokensUsed      int           `json:"totalTokensUsed"`
	AverageResponseTime  time.Duration `json:"averageResponseTime"`
	BatchCount           int           `json:"batchCount"`
	ErrorCount           int           `json:"errorCount"`
	SuccessRate          float64       `json:"successRate"`
}

// NewAIAnalysisEngine 创建AI分析引擎
func NewAIAnalysisEngine() *AIAnalysisEngine {
	// 使用传统AI服务保持兼容性，但在分析时使用增强管理器
	aiService := GetAIService()
	
	return &AIAnalysisEngine{
		aiService:      &aiService,
		batchService:   &BatchEvaluationService{},
		contextBuilder: &ProjectContextBuilder{},
	}
}

// StartAnalysis 开始AI分析
func (engine *AIAnalysisEngine) StartAnalysis(request AIAnalysisRequest) (*nesma.AIProjectAnalysis, error) {
	global.GVA_LOG.Info("开始AI项目分析",
		zap.Uint("projectID", request.ProjectID),
		zap.Uint("evaluationID", request.EvaluationID),
		zap.String("analysisType", request.AnalysisType))

	// 1. 创建分析记录
	analysis := &nesma.AIProjectAnalysis{
		ProjectID:       request.ProjectID,
		EvaluationID:    request.EvaluationID,
		AnalysisType:    request.AnalysisType,
		AnalysisVersion: fmt.Sprintf("v%d", time.Now().Unix()),
		Status:          "processing",
		Progress:        0.0,
		StartTime:       time.Now(),
		AIModel:         request.AnalysisConfig.AIModel,
		MaxTokens:       request.AnalysisConfig.MaxTokens,
		Temperature:     request.AnalysisConfig.Temperature,
		BatchSize:       request.AnalysisConfig.BatchSize,
	}

	if err := global.GVA_DB.Create(analysis).Error; err != nil {
		return nil, fmt.Errorf("创建分析记录失败: %w", err)
	}

	// 2. 异步执行分析
	go engine.performAnalysis(analysis, request)

	return analysis, nil
}

// performAnalysis 执行分析
func (engine *AIAnalysisEngine) performAnalysis(analysis *nesma.AIProjectAnalysis, request AIAnalysisRequest) {
	startTime := time.Now()
	metrics := &ExecutionMetrics{}

	defer func() {
		metrics.TotalExecutionTime = time.Since(startTime)
		if metrics.AICallCount > 0 {
			metrics.AverageResponseTime = metrics.TotalExecutionTime / time.Duration(metrics.AICallCount)
			metrics.SuccessRate = float64(metrics.AICallCount-metrics.ErrorCount) / float64(metrics.AICallCount)
		}
		
		// 更新分析状态
		engine.updateAnalysisStatus(analysis.ID, "completed", 100.0, metrics)
		
		global.GVA_LOG.Info("AI项目分析完成",
			zap.Uint("analysisID", analysis.ID),
			zap.Duration("executionTime", metrics.TotalExecutionTime),
			zap.Int("aiCallCount", metrics.AICallCount),
			zap.Float64("successRate", metrics.SuccessRate))
	}()

	// 1. 构建项目上下文
	engine.updateAnalysisProgress(analysis.ID, 5.0, "构建项目上下文...")
	context, err := engine.buildProjectContext(analysis, request.AnalysisConfig)
	if err != nil {
		engine.updateAnalysisStatus(analysis.ID, "failed", 0.0, metrics)
		global.GVA_LOG.Error("构建项目上下文失败", zap.Error(err))
		return
	}

	// 2. 执行各维度分析
	progressStep := 95.0 / float64(len(request.FocusAreas))
	currentProgress := 5.0

	for _, focusArea := range request.FocusAreas {
		engine.updateAnalysisProgress(analysis.ID, currentProgress, fmt.Sprintf("执行%s分析...", focusArea))
		
		switch focusArea {
		case "functional":
			err = engine.performFunctionalAnalysis(analysis.ID, context, metrics)
		case "technical":
			err = engine.performTechnicalAnalysis(analysis.ID, context, metrics)
		case "business":
			err = engine.performBusinessAnalysis(analysis.ID, context, metrics)
		case "quality":
			err = engine.performQualityAnalysis(analysis.ID, context, metrics)
		case "risk":
			err = engine.performRiskAnalysis(analysis.ID, context, metrics)
		case "recommendation":
			err = engine.performRecommendationAnalysis(analysis.ID, context, metrics)
		case "compliance":
			err = engine.performComplianceAnalysis(analysis.ID, context, metrics)
		}

		if err != nil {
			metrics.ErrorCount++
			global.GVA_LOG.Error("分析执行失败", 
				zap.String("focusArea", focusArea), 
				zap.Error(err))
		}

		currentProgress += progressStep
		engine.updateAnalysisProgress(analysis.ID, currentProgress, fmt.Sprintf("%s分析完成", focusArea))
	}

	// 3. 生成综合分析摘要
	engine.updateAnalysisProgress(analysis.ID, 98.0, "生成综合分析摘要...")
	err = engine.generateAnalysisSummary(analysis.ID, context, metrics)
	if err != nil {
		global.GVA_LOG.Error("生成分析摘要失败", zap.Error(err))
	}
}

// buildProjectContext 构建项目上下文
func (engine *AIAnalysisEngine) buildProjectContext(analysis *nesma.AIProjectAnalysis, config AnalysisConfiguration) (*ProjectAnalysisContext, error) {
	builder := NewProjectContextBuilder(analysis.ProjectID, analysis.EvaluationID)
	context, err := builder.BuildContext(config)
	if err != nil {
		return nil, err
	}

	// 保存上下文到分析记录
	contextJSON, _ := context.ToJSON()
	global.GVA_DB.Model(analysis).Updates(map[string]interface{}{
		"project_context":        contextJSON,
		"total_requirements":     context.RequirementStats.TotalCount,
		"processed_requirements": 0,
		"requirement_levels":     engine.serializeRequirementLevels(context.RequirementStats.LevelDistribution),
	})

	return context, nil
}

// performFunctionalAnalysis 执行功能性分析
func (engine *AIAnalysisEngine) performFunctionalAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始功能性分析", zap.Uint("analysisID", analysisID))

	// 构建功能分析提示词
	prompt := engine.buildFunctionalAnalysisPrompt(context)
	
	// 调用AI进行分析
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	// 解析AI结果
	functionalResult, err := engine.parseFunctionalAnalysisResult(result)
	if err != nil {
		return err
	}

	// 保存分析结果
	functionalAnalysis := &nesma.AIFunctionalAnalysis{
		AnalysisID:                  analysisID,
		TotalFunctionPoints:         functionalResult.TotalFunctionPoints,
		DataFunctionPoints:          functionalResult.DataFunctionPoints,
		TransactionalFunctionPoints: functionalResult.TransactionalFunctionPoints,
		ILFCount:                    functionalResult.ILFCount,
		EIFCount:                    functionalResult.EIFCount,
		EICount:                     functionalResult.EICount,
		EOCount:                     functionalResult.EOCount,
		EQCount:                     functionalResult.EQCount,
		LowComplexityCount:          functionalResult.LowComplexityCount,
		MediumComplexityCount:       functionalResult.MediumComplexityCount,
		HighComplexityCount:         functionalResult.HighComplexityCount,
		FunctionalCoverage:          functionalResult.FunctionalCoverage,
		FunctionalCompleteness:      functionalResult.FunctionalCompleteness,
		FunctionalConsistency:       functionalResult.FunctionalConsistency,
		AIAssessment:                result,
		ConfidenceScore:             functionalResult.ConfidenceScore,
	}

	if functionalResult.FunctionTypeDistribution != nil {
		distData, _ := json.Marshal(functionalResult.FunctionTypeDistribution)
		functionalAnalysis.FunctionTypeDistribution = string(distData)
	}

	if functionalResult.ComplexityAnalysis != nil {
		complexityData, _ := json.Marshal(functionalResult.ComplexityAnalysis)
		functionalAnalysis.ComplexityAnalysis = string(complexityData)
	}

	if functionalResult.QualityIndicators != nil {
		qualityData, _ := json.Marshal(functionalResult.QualityIndicators)
		functionalAnalysis.QualityIndicators = string(qualityData)
	}

	return global.GVA_DB.Create(functionalAnalysis).Error
}

// performTechnicalAnalysis 执行技术性分析
func (engine *AIAnalysisEngine) performTechnicalAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始技术性分析", zap.Uint("analysisID", analysisID))

	prompt := engine.buildTechnicalAnalysisPrompt(context)
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	technicalResult, err := engine.parseTechnicalAnalysisResult(result)
	if err != nil {
		return err
	}

	technicalAnalysis := &nesma.AITechnicalAnalysis{
		AnalysisID:                analysisID,
		ArchitectureType:          technicalResult.ArchitectureType,
		IntegrationComplexity:     technicalResult.IntegrationComplexity,
		DevelopmentEffort:         technicalResult.DevelopmentEffort,
		TestingEffort:             technicalResult.TestingEffort,
		DeploymentComplexity:      technicalResult.DeploymentComplexity,
		TechnicalFeasibility:      technicalResult.TechnicalFeasibility,
		ImplementationRisk:        technicalResult.ImplementationRisk,
		TechnicalInnovation:       technicalResult.TechnicalInnovation,
		MaintenanceComplexity:     technicalResult.MaintenanceComplexity,
		AITechnicalAssessment:     result,
		TechnicalRecommendations:  technicalResult.TechnicalRecommendations,
		ConfidenceScore:           technicalResult.ConfidenceScore,
	}

	if technicalResult.TechnologyStack != nil {
		techData, _ := json.Marshal(technicalResult.TechnologyStack)
		technicalAnalysis.TechnologyStack = string(techData)
	}

	return global.GVA_DB.Create(technicalAnalysis).Error
}

// performBusinessAnalysis 执行业务性分析
func (engine *AIAnalysisEngine) performBusinessAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始业务性分析", zap.Uint("analysisID", analysisID))

	prompt := engine.buildBusinessAnalysisPrompt(context)
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	businessResult, err := engine.parseBusinessAnalysisResult(result)
	if err != nil {
		return err
	}

	businessAnalysis := &nesma.AIBusinessAnalysis{
		AnalysisID:                analysisID,
		BusinessValue:             businessResult.BusinessValue,
		ROIEstimation:             businessResult.ROIEstimation,
		StrategicAlignment:        businessResult.StrategicAlignment,
		UserExperienceScore:       businessResult.UserExperienceScore,
		ProcessEfficiency:         businessResult.ProcessEfficiency,
		ProcessAutomation:         businessResult.ProcessAutomation,
		BusinessLogicComplexity:   businessResult.BusinessLogicComplexity,
		MarketFit:                 businessResult.MarketFit,
		CompetitiveAdvantage:      businessResult.CompetitiveAdvantage,
		InnovationLevel:           businessResult.InnovationLevel,
		RegulatoryCompliance:      businessResult.RegulatoryCompliance,
		AIBusinessAssessment:      result,
		BusinessRecommendations:   businessResult.BusinessRecommendations,
		ConfidenceScore:           businessResult.ConfidenceScore,
	}

	return global.GVA_DB.Create(businessAnalysis).Error
}

// performQualityAnalysis 执行质量分析
func (engine *AIAnalysisEngine) performQualityAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始质量分析", zap.Uint("analysisID", analysisID))

	prompt := engine.buildQualityAnalysisPrompt(context)
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	qualityResult, err := engine.parseQualityAnalysisResult(result)
	if err != nil {
		return err
	}

	qualityAnalysis := &nesma.AIQualityAnalysis{
		AnalysisID:                analysisID,
		OverallQuality:            qualityResult.OverallQuality,
		RequirementQuality:        qualityResult.RequirementQuality,
		DesignQuality:             qualityResult.DesignQuality,
		Correctness:               qualityResult.Correctness,
		Completeness:              qualityResult.Completeness,
		Consistency:               qualityResult.Consistency,
		Clarity:                   qualityResult.Clarity,
		Traceability:              qualityResult.Traceability,
		Maintainability:           qualityResult.Maintainability,
		Modularity:                qualityResult.Modularity,
		Reusability:               qualityResult.Reusability,
		Testability:               qualityResult.Testability,
		AIQualityAssessment:       result,
		QualityRecommendations:    qualityResult.QualityRecommendations,
		ConfidenceScore:           qualityResult.ConfidenceScore,
	}

	return global.GVA_DB.Create(qualityAnalysis).Error
}

// performRiskAnalysis 执行风险分析
func (engine *AIAnalysisEngine) performRiskAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始风险分析", zap.Uint("analysisID", analysisID))

	prompt := engine.buildRiskAnalysisPrompt(context)
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	riskResult, err := engine.parseRiskAnalysisResult(result)
	if err != nil {
		return err
	}

	riskAnalysis := &nesma.AIRiskAnalysis{
		AnalysisID:           analysisID,
		OverallRisk:          riskResult.OverallRisk,
		RiskLevel:            riskResult.RiskLevel,
		TechnicalRisk:        riskResult.TechnicalRisk,
		ImplementationRisk:   riskResult.ImplementationRisk,
		IntegrationRisk:      riskResult.IntegrationRisk,
		ScheduleRisk:         riskResult.ScheduleRisk,
		BudgetRisk:           riskResult.BudgetRisk,
		ResourceRisk:         riskResult.ResourceRisk,
		BusinessRisk:         riskResult.BusinessRisk,
		MarketRisk:           riskResult.MarketRisk,
		ComplianceRisk:       riskResult.ComplianceRisk,
		AIRiskAssessment:     result,
		RiskRecommendations:  riskResult.RiskRecommendations,
		ConfidenceScore:      riskResult.ConfidenceScore,
	}

	return global.GVA_DB.Create(riskAnalysis).Error
}

// performRecommendationAnalysis 执行建议分析
func (engine *AIAnalysisEngine) performRecommendationAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始建议分析", zap.Uint("analysisID", analysisID))

	prompt := engine.buildRecommendationAnalysisPrompt(context)
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	recommendations, err := engine.parseRecommendationAnalysisResult(result)
	if err != nil {
		return err
	}

	// 创建多个建议记录
	for _, rec := range recommendations {
		recommendationAnalysis := &nesma.AIRecommendationAnalysis{
			AnalysisID:               analysisID,
			RecommendationType:       rec.RecommendationType,
			Priority:                 rec.Priority,
			ImpactLevel:              rec.ImpactLevel,
			Title:                    rec.Title,
			Description:              rec.Description,
			Rationale:                rec.Rationale,
			EstimatedEffort:          rec.EstimatedEffort,
			ExpectedBenefit:          rec.ExpectedBenefit,
			AIGeneratedRecommendation: result,
			ConfidenceScore:          rec.ConfidenceScore,
			Urgency:                  rec.Urgency,
		}

		if err := global.GVA_DB.Create(recommendationAnalysis).Error; err != nil {
			global.GVA_LOG.Error("保存建议分析失败", zap.Error(err))
		}
	}

	return nil
}

// performComplianceAnalysis 执行合规性分析
func (engine *AIAnalysisEngine) performComplianceAnalysis(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("开始合规性分析", zap.Uint("analysisID", analysisID))

	prompt := engine.buildComplianceAnalysisPrompt(context)
	result, err := engine.callAIForAnalysis(prompt, context.AnalysisConfig, metrics)
	if err != nil {
		return err
	}

	complianceResult, err := engine.parseComplianceAnalysisResult(result)
	if err != nil {
		return err
	}

	complianceAnalysis := &nesma.AIComplianceAnalysis{
		AnalysisID:                  analysisID,
		ComplianceStandard:          complianceResult.ComplianceStandard,
		StandardVersion:             complianceResult.StandardVersion,
		OverallCompliance:           complianceResult.OverallCompliance,
		ComplianceLevel:             complianceResult.ComplianceLevel,
		AIComplianceAssessment:      result,
		ComplianceRecommendations:   complianceResult.ComplianceRecommendations,
		ConfidenceScore:             complianceResult.ConfidenceScore,
	}

	return global.GVA_DB.Create(complianceAnalysis).Error
}

// generateAnalysisSummary 生成分析摘要
func (engine *AIAnalysisEngine) generateAnalysisSummary(analysisID uint, context *ProjectAnalysisContext, metrics *ExecutionMetrics) error {
	global.GVA_LOG.Info("生成分析摘要", zap.Uint("analysisID", analysisID))

	// 获取所有分析结果
	var functional nesma.AIFunctionalAnalysis
	var technical nesma.AITechnicalAnalysis
	var business nesma.AIBusinessAnalysis
	var quality nesma.AIQualityAnalysis
	var risk nesma.AIRiskAnalysis

	global.GVA_DB.Where("analysis_id = ?", analysisID).First(&functional)
	global.GVA_DB.Where("analysis_id = ?", analysisID).First(&technical)
	global.GVA_DB.Where("analysis_id = ?", analysisID).First(&business)
	global.GVA_DB.Where("analysis_id = ?", analysisID).First(&quality)
	global.GVA_DB.Where("analysis_id = ?", analysisID).First(&risk)

	// 计算总体评分
	overallScore := engine.calculateOverallScore(functional, technical, business, quality, risk)
	
	// 确定复杂度等级
	complexityLevel := engine.determineComplexityLevel(context, functional, technical)
	
	// 确定风险等级
	riskLevel := engine.determineRiskLevel(risk, technical, business)

	// 更新分析记录
	updates := map[string]interface{}{
		"overall_score":        overallScore,
		"complexity_level":     complexityLevel,
		"risk_level":           riskLevel,
		"recommendation_level": "medium", // 基于风险和质量确定
	}

	return global.GVA_DB.Model(&nesma.AIProjectAnalysis{}).
		Where("id = ?", analysisID).Updates(updates).Error
}

// Helper methods for AI calls and result parsing will be implemented in the next part
// 这些方法将在下一部分实现，包括：
// - callAIForAnalysis
// - buildXXXAnalysisPrompt 系列方法
// - parseXXXAnalysisResult 系列方法
// - updateAnalysisProgress
// - updateAnalysisStatus

// 工具方法
func (engine *AIAnalysisEngine) serializeRequirementLevels(levels map[int]int) string {
	data, _ := json.Marshal(levels)
	return string(data)
}

func (engine *AIAnalysisEngine) calculateOverallScore(functional nesma.AIFunctionalAnalysis, technical nesma.AITechnicalAnalysis, business nesma.AIBusinessAnalysis, quality nesma.AIQualityAnalysis, risk nesma.AIRiskAnalysis) float64 {
	// 加权计算总体评分
	functionalWeight := 0.25
	technicalWeight := 0.20
	businessWeight := 0.20
	qualityWeight := 0.25
	riskWeight := 0.10

	score := functional.FunctionalCompleteness*functionalWeight +
		technical.TechnicalFeasibility*technicalWeight +
		business.BusinessValue*businessWeight +
		quality.OverallQuality*qualityWeight +
		(1.0-risk.OverallRisk)*riskWeight

	return math.Min(math.Max(score, 0.0), 1.0)
}

func (engine *AIAnalysisEngine) determineComplexityLevel(context *ProjectAnalysisContext, functional nesma.AIFunctionalAnalysis, technical nesma.AITechnicalAnalysis) string {
	// 基于多个因素确定复杂度等级
	complexityScore := 0.0
	
	// 项目规模因子
	if context.ProjectMetrics.ProjectSize == "enterprise" {
		complexityScore += 0.4
	} else if context.ProjectMetrics.ProjectSize == "large" {
		complexityScore += 0.3
	} else if context.ProjectMetrics.ProjectSize == "medium" {
		complexityScore += 0.2
	} else {
		complexityScore += 0.1
	}

	// 功能复杂度因子
	if functional.HighComplexityCount > functional.LowComplexityCount {
		complexityScore += 0.3
	} else {
		complexityScore += 0.1
	}

	// 技术复杂度因子
	complexityScore += technical.MaintenanceComplexity * 0.3

	if complexityScore >= 0.7 {
		return "high"
	} else if complexityScore >= 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

func (engine *AIAnalysisEngine) determineRiskLevel(risk nesma.AIRiskAnalysis, technical nesma.AITechnicalAnalysis, business nesma.AIBusinessAnalysis) string {
	riskScore := risk.OverallRisk

	if riskScore >= 0.7 {
		return "high"
	} else if riskScore >= 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

// updateAnalysisProgress 更新分析进度
func (engine *AIAnalysisEngine) updateAnalysisProgress(analysisID uint, progress float64, message string) {
	// 更新数据库进度
	global.GVA_DB.Model(&nesma.AIProjectAnalysis{}).
		Where("id = ?", analysisID).
		Updates(map[string]interface{}{
			"progress": progress,
		})

	// 获取分析记录以获取项目ID和用户ID
	var analysis nesma.AIProjectAnalysis
	if err := global.GVA_DB.Preload("Evaluation").First(&analysis, analysisID).Error; err == nil {
		// 集成WebSocket发送进度更新
		wsIntegration := NewAIAnalysisWebSocketIntegration()
		wsIntegration.NotifyAnalysisProgress(
			analysisID,
			analysis.ProjectID,
			analysis.Evaluation.EvaluatorID,
			progress,
			engine.getCurrentStepName(progress),
			message,
		)
	}
		
	global.GVA_LOG.Info("分析进度更新",
		zap.Uint("analysisID", analysisID),
		zap.Float64("progress", progress),
		zap.String("message", message))
}

// updateAnalysisStatus 更新分析状态
func (engine *AIAnalysisEngine) updateAnalysisStatus(analysisID uint, status string, progress float64, metrics *ExecutionMetrics) {
	updates := map[string]interface{}{
		"status":   status,
		"progress": progress,
	}

	if status == "completed" || status == "failed" {
		now := time.Now()
		updates["completion_time"] = &now
	}

	global.GVA_DB.Model(&nesma.AIProjectAnalysis{}).
		Where("id = ?", analysisID).Updates(updates)
}

// getCurrentStepName 根据进度获取当前步骤名称
func (engine *AIAnalysisEngine) getCurrentStepName(progress float64) string {
	switch {
	case progress < 10:
		return "构建项目上下文"
	case progress < 25:
		return "功能性分析"
	case progress < 40:
		return "技术性分析"
	case progress < 55:
		return "业务性分析"
	case progress < 70:
		return "质量分析"
	case progress < 85:
		return "风险分析"
	case progress < 95:
		return "建议分析"
	case progress < 100:
		return "生成综合分析摘要"
	default:
		return "分析完成"
	}
}