package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AIAnalysisApi struct{}

// StartAIAnalysis 启动AI分析
// @Tags AI分析
// @Summary 启动AI项目全面分析
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaService.AIAnalysisRequest true "AI分析请求参数"
// @Success 200 {object} response.Response{data=nesma.AIProjectAnalysis,msg=string} "启动成功"
// @Router /api/v1/nesma/ai-analysis/start [post]
func (api *AIAnalysisApi) StartAIAnalysis(c *gin.Context) {
	var req nesmaService.AIAnalysisRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 设置默认分析配置
	if req.AnalysisConfig.AIModel == "" {
		req.AnalysisConfig.AIModel = "deepseek-chat"
	}
	if req.AnalysisConfig.MaxTokens == 0 {
		req.AnalysisConfig.MaxTokens = 4000
	}
	if req.AnalysisConfig.Temperature == 0 {
		req.AnalysisConfig.Temperature = 0.1
	}
	if req.AnalysisConfig.BatchSize == 0 {
		req.AnalysisConfig.BatchSize = 10
	}
	if len(req.FocusAreas) == 0 {
		req.FocusAreas = []string{"functional", "technical", "business", "quality", "risk"}
	}

	engine := nesmaService.NewAIAnalysisEngine()
	analysis, err := engine.StartAnalysis(req)
	if err != nil {
		global.GVA_LOG.Error("启动AI分析失败", zap.Error(err))
		response.FailWithMessage("启动AI分析失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(analysis, "AI分析已启动", c)
}

// GetAnalysisStatus 获取分析状态
// @Tags AI分析
// @Summary 获取AI分析状态和进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param analysisId path int true "分析ID"
// @Success 200 {object} response.Response{data=nesma.AIProjectAnalysis,msg=string} "获取成功"
// @Router /api/v1/nesma/ai-analysis/{analysisId}/status [get]
func (api *AIAnalysisApi) GetAnalysisStatus(c *gin.Context) {
	analysisIdStr := c.Param("analysisId")
	analysisId, err := strconv.ParseUint(analysisIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("分析ID格式错误", c)
		return
	}

	var analysis nesma.AIProjectAnalysis
	err = global.GVA_DB.Preload("Project").Preload("Evaluation").
		First(&analysis, uint(analysisId)).Error
	if err != nil {
		response.FailWithMessage("分析记录不存在", c)
		return
	}

	response.OkWithData(analysis, c)
}

// GetAnalysisResult 获取完整分析结果
// @Tags AI分析
// @Summary 获取AI分析完整结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param analysisId path int true "分析ID"
// @Success 200 {object} response.Response{data=nesmaService.AIAnalysisResult,msg=string} "获取成功"
// @Router /api/v1/nesma/ai-analysis/{analysisId}/result [get]
func (api *AIAnalysisApi) GetAnalysisResult(c *gin.Context) {
	analysisIdStr := c.Param("analysisId")
	analysisId, err := strconv.ParseUint(analysisIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("分析ID格式错误", c)
		return
	}

	// 获取主分析记录
	var analysis nesma.AIProjectAnalysis
	err = global.GVA_DB.Preload("Project").Preload("Evaluation").
		First(&analysis, uint(analysisId)).Error
	if err != nil {
		response.FailWithMessage("分析记录不存在", c)
		return
	}

	// 获取各维度分析结果
	result := &nesmaService.AIAnalysisResult{
		AnalysisID: uint(analysisId),
		Status:     analysis.Status,
		Progress:   analysis.Progress,
	}

	// 功能分析
	var functionalAnalysis nesma.AIFunctionalAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).First(&functionalAnalysis).Error; err == nil {
		result.FunctionalAnalysis = &nesmaService.AIFunctionalAnalysisResult{
			TotalFunctionPoints:         functionalAnalysis.TotalFunctionPoints,
			DataFunctionPoints:          functionalAnalysis.DataFunctionPoints,
			TransactionalFunctionPoints: functionalAnalysis.TransactionalFunctionPoints,
			ILFCount:                    functionalAnalysis.ILFCount,
			EIFCount:                    functionalAnalysis.EIFCount,
			EICount:                     functionalAnalysis.EICount,
			EOCount:                     functionalAnalysis.EOCount,
			EQCount:                     functionalAnalysis.EQCount,
			LowComplexityCount:          functionalAnalysis.LowComplexityCount,
			MediumComplexityCount:       functionalAnalysis.MediumComplexityCount,
			HighComplexityCount:         functionalAnalysis.HighComplexityCount,
			FunctionalCoverage:          functionalAnalysis.FunctionalCoverage,
			FunctionalCompleteness:      functionalAnalysis.FunctionalCompleteness,
			FunctionalConsistency:       functionalAnalysis.FunctionalConsistency,
			ConfidenceScore:             functionalAnalysis.ConfidenceScore,
		}
	}

	// 技术分析
	var technicalAnalysis nesma.AITechnicalAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).First(&technicalAnalysis).Error; err == nil {
		result.TechnicalAnalysis = &nesmaService.AITechnicalAnalysisResult{
			ArchitectureType:         technicalAnalysis.ArchitectureType,
			IntegrationComplexity:    technicalAnalysis.IntegrationComplexity,
			DevelopmentEffort:        technicalAnalysis.DevelopmentEffort,
			TestingEffort:            technicalAnalysis.TestingEffort,
			DeploymentComplexity:     technicalAnalysis.DeploymentComplexity,
			TechnicalFeasibility:     technicalAnalysis.TechnicalFeasibility,
			ImplementationRisk:       technicalAnalysis.ImplementationRisk,
			TechnicalInnovation:      technicalAnalysis.TechnicalInnovation,
			MaintenanceComplexity:    technicalAnalysis.MaintenanceComplexity,
			TechnicalRecommendations: technicalAnalysis.TechnicalRecommendations,
			ConfidenceScore:          technicalAnalysis.ConfidenceScore,
		}
	}

	// 业务分析
	var businessAnalysis nesma.AIBusinessAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).First(&businessAnalysis).Error; err == nil {
		result.BusinessAnalysis = &nesmaService.AIBusinessAnalysisResult{
			BusinessValue:           businessAnalysis.BusinessValue,
			ROIEstimation:           businessAnalysis.ROIEstimation,
			StrategicAlignment:      businessAnalysis.StrategicAlignment,
			UserExperienceScore:     businessAnalysis.UserExperienceScore,
			ProcessEfficiency:       businessAnalysis.ProcessEfficiency,
			ProcessAutomation:       businessAnalysis.ProcessAutomation,
			BusinessLogicComplexity: businessAnalysis.BusinessLogicComplexity,
			MarketFit:               businessAnalysis.MarketFit,
			CompetitiveAdvantage:    businessAnalysis.CompetitiveAdvantage,
			InnovationLevel:         businessAnalysis.InnovationLevel,
			RegulatoryCompliance:    businessAnalysis.RegulatoryCompliance,
			BusinessRecommendations: businessAnalysis.BusinessRecommendations,
			ConfidenceScore:         businessAnalysis.ConfidenceScore,
		}
	}

	// 质量分析
	var qualityAnalysis nesma.AIQualityAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).First(&qualityAnalysis).Error; err == nil {
		result.QualityAnalysis = &nesmaService.AIQualityAnalysisResult{
			OverallQuality:         qualityAnalysis.OverallQuality,
			RequirementQuality:     qualityAnalysis.RequirementQuality,
			DesignQuality:          qualityAnalysis.DesignQuality,
			Correctness:            qualityAnalysis.Correctness,
			Completeness:           qualityAnalysis.Completeness,
			Consistency:            qualityAnalysis.Consistency,
			Clarity:                qualityAnalysis.Clarity,
			Traceability:           qualityAnalysis.Traceability,
			Maintainability:        qualityAnalysis.Maintainability,
			Modularity:             qualityAnalysis.Modularity,
			Reusability:            qualityAnalysis.Reusability,
			Testability:            qualityAnalysis.Testability,
			QualityRecommendations: qualityAnalysis.QualityRecommendations,
			ConfidenceScore:        qualityAnalysis.ConfidenceScore,
		}
	}

	// 风险分析
	var riskAnalysis nesma.AIRiskAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).First(&riskAnalysis).Error; err == nil {
		result.RiskAnalysis = &nesmaService.AIRiskAnalysisResult{
			OverallRisk:         riskAnalysis.OverallRisk,
			RiskLevel:           riskAnalysis.RiskLevel,
			TechnicalRisk:       riskAnalysis.TechnicalRisk,
			ImplementationRisk:  riskAnalysis.ImplementationRisk,
			IntegrationRisk:     riskAnalysis.IntegrationRisk,
			ScheduleRisk:        riskAnalysis.ScheduleRisk,
			BudgetRisk:          riskAnalysis.BudgetRisk,
			ResourceRisk:        riskAnalysis.ResourceRisk,
			BusinessRisk:        riskAnalysis.BusinessRisk,
			MarketRisk:          riskAnalysis.MarketRisk,
			ComplianceRisk:      riskAnalysis.ComplianceRisk,
			RiskRecommendations: riskAnalysis.RiskRecommendations,
			ConfidenceScore:     riskAnalysis.ConfidenceScore,
		}
	}

	// 获取建议列表
	var recommendations []nesma.AIRecommendationAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Find(&recommendations).Error; err == nil {
		for _, rec := range recommendations {
			recResult := nesmaService.AIRecommendationResult{
				RecommendationType: rec.RecommendationType,
				Priority:           rec.Priority,
				ImpactLevel:        rec.ImpactLevel,
				Title:              rec.Title,
				Description:        rec.Description,
				Rationale:          rec.Rationale,
				EstimatedEffort:    rec.EstimatedEffort,
				ExpectedBenefit:    rec.ExpectedBenefit,
				ConfidenceScore:    rec.ConfidenceScore,
				Urgency:            rec.Urgency,
			}
			result.RecommendationAnalysis = append(result.RecommendationAnalysis, recResult)
		}
	}

	// 合规性分析
	var complianceAnalysis nesma.AIComplianceAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisId).First(&complianceAnalysis).Error; err == nil {
		result.ComplianceAnalysis = &nesmaService.AIComplianceAnalysisResult{
			ComplianceStandard:        complianceAnalysis.ComplianceStandard,
			StandardVersion:           complianceAnalysis.StandardVersion,
			OverallCompliance:         complianceAnalysis.OverallCompliance,
			ComplianceLevel:           complianceAnalysis.ComplianceLevel,
			ComplianceRecommendations: complianceAnalysis.ComplianceRecommendations,
			ConfidenceScore:           complianceAnalysis.ConfidenceScore,
		}
	}

	// 生成摘要信息
	result.Summary = &nesmaService.AnalysisSummary{
		OverallScore:        analysis.OverallScore,
		ComplexityLevel:     analysis.ComplexityLevel,
		RiskLevel:           analysis.RiskLevel,
		RecommendationLevel: analysis.RecommendationLevel,
		ProcessedItems:      analysis.ProcessedRequirements,
	}

	response.OkWithData(result, c)
}

// GetAnalysisHistory 获取项目分析历史
// @Tags AI分析
// @Summary 获取项目AI分析历史记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId path int true "项目ID"
// @Success 200 {object} response.Response{data=[]nesma.AIProjectAnalysis,msg=string} "获取成功"
// @Router /api/v1/nesma/ai-analysis/project/{projectId}/history [get]
func (api *AIAnalysisApi) GetAnalysisHistory(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	var analyses []nesma.AIProjectAnalysis
	err = global.GVA_DB.Preload("Evaluation").
		Where("project_id = ?", uint(projectId)).
		Order("created_at DESC").
		Find(&analyses).Error
	if err != nil {
		response.FailWithMessage("获取分析历史失败", c)
		return
	}

	response.OkWithData(analyses, c)
}

// DeleteAnalysis 删除分析记录
// @Tags AI分析
// @Summary 删除AI分析记录及相关数据
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param analysisId path int true "分析ID"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /api/v1/nesma/ai-analysis/{analysisId} [delete]
func (api *AIAnalysisApi) DeleteAnalysis(c *gin.Context) {
	analysisIdStr := c.Param("analysisId")
	analysisId, err := strconv.ParseUint(analysisIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("分析ID格式错误", c)
		return
	}

	// 使用事务删除所有相关记录
	err = global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 删除各维度分析结果
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AIFunctionalAnalysis{}).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AITechnicalAnalysis{}).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AIBusinessAnalysis{}).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AIQualityAnalysis{}).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AIRiskAnalysis{}).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AIRecommendationAnalysis{}).Error; err != nil {
			return err
		}
		if err := global.GVA_DB.Where("analysis_id = ?", analysisId).Delete(&nesma.AIComplianceAnalysis{}).Error; err != nil {
			return err
		}

		// 删除主分析记录
		if err := global.GVA_DB.Delete(&nesma.AIProjectAnalysis{}, uint(analysisId)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		global.GVA_LOG.Error("删除分析记录失败", zap.Error(err))
		response.FailWithMessage("删除分析记录失败", c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

// GetAnalysisReport 生成分析报告
// @Tags AI分析
// @Summary 生成AI分析报告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param analysisId path int true "分析ID"
// @Param format query string false "报告格式" Enums(json,markdown,pdf)
// @Success 200 {object} response.Response{data=string,msg=string} "生成成功"
// @Router /api/v1/nesma/ai-analysis/{analysisId}/report [get]
func (api *AIAnalysisApi) GetAnalysisReport(c *gin.Context) {
	analysisIdStr := c.Param("analysisId")
	analysisId, err := strconv.ParseUint(analysisIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("分析ID格式错误", c)
		return
	}

	format := c.DefaultQuery("format", "json")
	
	// 获取完整分析结果（复用GetAnalysisResult的逻辑）
	// 这里可以调用报告生成服务来格式化输出
	// 暂时返回JSON格式
	
	var analysis nesma.AIProjectAnalysis
	err = global.GVA_DB.Preload("Project").Preload("Evaluation").
		First(&analysis, uint(analysisId)).Error
	if err != nil {
		response.FailWithMessage("分析记录不存在", c)
		return
	}

	// 根据格式生成不同的报告
	switch format {
	case "markdown":
		// TODO: 实现Markdown格式报告
		response.OkWithDetailed("Markdown报告生成功能待实现", "报告已生成", c)
	case "pdf":
		// TODO: 实现PDF格式报告  
		response.OkWithDetailed("PDF报告生成功能待实现", "报告已生成", c)
	default:
		// 返回JSON格式
		response.OkWithData(analysis, c)
	}
}