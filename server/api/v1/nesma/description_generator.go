package nesma

import (
	"fmt"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaModel "github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DescriptionGeneratorApi struct{}

// GenerateRequirementDescriptions 生成需求描述
// @Tags DescriptionGenerator
// @Summary 生成需求描述
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.DescriptionGenerationRequest true "生成请求"
// @Success 200 {object} response.Response{data=[]nesmaService.DescriptionGenerationResult} "生成成功"
// @Router /nesma/generator/description/generate [post]
func (a *DescriptionGeneratorApi) GenerateRequirementDescriptions(c *gin.Context) {
	var req nesmaReq.DescriptionGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	generatorService := nesmaService.GetDescriptionGeneratorService()
	
	// 执行生成
	results, err := generatorService.GenerateRequirementDescriptions(req.CycleID, req.RequirementIDs, req.Levels)
	if err != nil {
		global.GVA_LOG.Error("需求描述生成失败", zap.Error(err))
		response.FailWithMessage("生成失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("需求描述生成完成", 
		zap.Uint("cycleID", req.CycleID),
		zap.Int("结果数量", len(results)),
	)

	response.OkWithData(gin.H{
		"results": results,
		"total":   len(results),
	}, c)
}

// ApplyGeneratedDescription 应用生成的描述
// @Tags DescriptionGenerator
// @Summary 应用生成的描述
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ApplyDescriptionRequest true "应用请求"
// @Success 200 {object} response.Response{} "应用成功"
// @Router /nesma/generator/description/apply [post]
func (a *DescriptionGeneratorApi) ApplyGeneratedDescription(c *gin.Context) {
	var req nesmaReq.ApplyDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.RequirementID == 0 {
		response.FailWithMessage("功能点ID不能为空", c)
		return
	}

	if req.EnhancedDescription == nil {
		response.FailWithMessage("增强描述不能为空", c)
		return
	}

	generatorService := nesmaService.GetDescriptionGeneratorService()
	
	// 应用生成的描述
	err := generatorService.ApplyGeneratedDescription(req.RequirementID, req.EnhancedDescription)
	if err != nil {
		global.GVA_LOG.Error("应用生成描述失败", zap.Error(err))
		response.FailWithMessage("应用失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("生成描述应用成功", 
		zap.Uint("requirementID", req.RequirementID),
	)

	response.OkWithMessage("生成描述应用成功", c)
}

// BatchGenerateDescriptions 批量生成描述
// @Tags DescriptionGenerator
// @Summary 批量生成描述
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchDescriptionGenerationRequest true "批量生成请求"
// @Success 200 {object} response.Response{} "批量生成成功"
// @Router /nesma/generator/description/batch-generate [post]
func (a *DescriptionGeneratorApi) BatchGenerateDescriptions(c *gin.Context) {
	var req nesmaReq.BatchDescriptionGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if len(req.Requests) == 0 {
		response.FailWithMessage("生成请求列表不能为空", c)
		return
	}

	generatorService := nesmaService.GetDescriptionGeneratorService()
	
	var allResults []*nesmaService.DescriptionGenerationResult
	var successCount, failedCount int
	var errors []string

	// 批量生成描述
	for _, genReq := range req.Requests {
		results, err := generatorService.GenerateRequirementDescriptions(genReq.CycleID, genReq.RequirementIDs, genReq.Levels)
		if err != nil {
			failedCount++
			errors = append(errors, err.Error())
			global.GVA_LOG.Error("批量生成描述失败", 
				zap.Uint("cycleID", genReq.CycleID),
				zap.Error(err),
			)
		} else {
			successCount++
			allResults = append(allResults, results...)
		}
	}

	global.GVA_LOG.Info("批量生成描述完成", 
		zap.Int("成功数量", successCount),
		zap.Int("失败数量", failedCount),
		zap.Int("总结果数", len(allResults)),
	)

	response.OkWithData(gin.H{
		"results":       allResults,
		"success_count": successCount,
		"failed_count":  failedCount,
		"errors":        errors,
		"total_results": len(allResults),
	}, c)
}

// GetDescriptionGenerationHistory 获取描述生成历史
// @Tags DescriptionGenerator
// @Summary 获取描述生成历史
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Param requirementId query int false "功能点ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/generator/description/history/{cycleId} [get]
func (a *DescriptionGeneratorApi) GetDescriptionGenerationHistory(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取查询参数
	requirementIDStr := c.Query("requirementId")
	var requirementID uint
	if requirementIDStr != "" {
		if id, err := strconv.ParseUint(requirementIDStr, 10, 32); err == nil {
			requirementID = uint(id)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// 查询描述生成历史
	history := a.getDescriptionHistory(uint(cycleID), requirementID, page, pageSize)

	response.OkWithData(history, c)
}

// getDescriptionHistory 获取描述生成历史
func (a *DescriptionGeneratorApi) getDescriptionHistory(cycleID uint, requirementID uint, page, pageSize int) gin.H {
	var history []gin.H
	var total int64

	// 构建查询
	query := global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND ai_analysis_status = ?", cycleID, "enhanced")

	if requirementID > 0 {
		query = query.Where("id = ?", requirementID)
	}

	// 获取总数
	query.Count(&total)

	// 分页查询
	var requirements []nesmaModel.NesmaRequirement
	offset := (page - 1) * pageSize
	query.Order("ai_analysis_time DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&requirements)

	// 构建历史记录
	for _, req := range requirements {
		history = append(history, gin.H{
			"id":                req.ID,
			"title":             req.Title,
			"level":             req.Level,
			"ai_analysis_time":  req.AIAnalysisTime,
			"ai_confidence":     req.AIConfidenceScore,
			"description_length": len(req.AIDescription),
			"has_enhancement":   req.AIDescription != "",
		})
	}

	return gin.H{
		"list":     history,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
		"cycleId":  cycleID,
	}
}

// GetDescriptionGenerationStats 获取描述生成统计
// @Tags DescriptionGenerator
// @Summary 获取描述生成统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/generator/description/stats/{cycleId} [get]
func (a *DescriptionGeneratorApi) GetDescriptionGenerationStats(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取统计信息
	stats := a.getDescriptionStats(uint(cycleID))

	response.OkWithData(stats, c)
}

// getDescriptionStats 获取描述生成统计
func (a *DescriptionGeneratorApi) getDescriptionStats(cycleID uint) gin.H {
	var totalRequirements int64
	var enhancedRequirements int64
	var avgConfidence float64

	// 统计总需求数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ?", cycleID).
		Count(&totalRequirements)

	// 统计增强描述的需求数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND ai_analysis_status = 'enhanced'", cycleID).
		Count(&enhancedRequirements)

	// 计算平均置信度
	var confidenceResult struct {
		AvgConfidence float64 `json:"avg_confidence"`
	}
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND ai_confidence_score IS NOT NULL", cycleID).
		Select("AVG(ai_confidence_score) as avg_confidence").
		Scan(&confidenceResult)
	
	avgConfidence = confidenceResult.AvgConfidence

	// 按级别统计
	levelStats := a.getLevelStats(cycleID)

	// 描述质量分布
	qualityDistribution := a.getQualityDistribution(cycleID)

	return gin.H{
		"cycle_id":             cycleID,
		"total_requirements":   totalRequirements,
		"enhanced_requirements": enhancedRequirements,
		"enhancement_rate":     func() float64 {
			if totalRequirements == 0 {
				return 0
			}
			return float64(enhancedRequirements) / float64(totalRequirements)
		}(),
		"average_confidence":   avgConfidence,
		"level_stats":          levelStats,
		"quality_distribution": qualityDistribution,
	}
}

// getLevelStats 获取级别统计
func (a *DescriptionGeneratorApi) getLevelStats(cycleID uint) map[string]interface{} {
	var levelResults []struct {
		Level int   `json:"level"`
		Total int64 `json:"total"`
		Enhanced int64 `json:"enhanced"`
	}

	// 查询各级别的统计
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ?", cycleID).
		Select("level, COUNT(*) as total, SUM(CASE WHEN ai_analysis_status = 'enhanced' THEN 1 ELSE 0 END) as enhanced").
		Group("level").
		Scan(&levelResults)

	levelStats := make(map[string]interface{})
	for _, result := range levelResults {
		levelName := fmt.Sprintf("level_%d", result.Level)
		levelStats[levelName] = gin.H{
			"total":    result.Total,
			"enhanced": result.Enhanced,
			"rate":     func() float64 {
				if result.Total == 0 {
					return 0
				}
				return float64(result.Enhanced) / float64(result.Total)
			}(),
		}
	}

	return levelStats
}

// getQualityDistribution 获取质量分布
func (a *DescriptionGeneratorApi) getQualityDistribution(cycleID uint) map[string]int64 {
	var qualityResults []struct {
		Quality string `json:"quality"`
		Count   int64  `json:"count"`
	}

	// 按置信度分组统计
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND ai_confidence_score IS NOT NULL", cycleID).
		Select(`
			CASE 
				WHEN ai_confidence_score >= 0.8 THEN '高质量'
				WHEN ai_confidence_score >= 0.6 THEN '中等质量'
				ELSE '低质量'
			END as quality,
			COUNT(*) as count
		`).
		Group("quality").
		Scan(&qualityResults)

	distribution := make(map[string]int64)
	for _, result := range qualityResults {
		distribution[result.Quality] = result.Count
	}

	return distribution
}

// PreviewDescription 预览描述生成
// @Tags DescriptionGenerator
// @Summary 预览描述生成
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.PreviewDescriptionRequest true "预览请求"
// @Success 200 {object} response.Response{} "预览成功"
// @Router /nesma/generator/description/preview [post]
func (a *DescriptionGeneratorApi) PreviewDescription(c *gin.Context) {
	var req nesmaReq.PreviewDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.RequirementID == 0 {
		response.FailWithMessage("功能点ID不能为空", c)
		return
	}

	generatorService := nesmaService.GetDescriptionGeneratorService()
	
	// 生成预览描述
	results, err := generatorService.GenerateRequirementDescriptions(0, []uint{req.RequirementID}, []int{})
	if err != nil {
		global.GVA_LOG.Error("预览描述生成失败", zap.Error(err))
		response.FailWithMessage("预览失败: "+err.Error(), c)
		return
	}

	if len(results) == 0 {
		response.FailWithMessage("未找到功能点", c)
		return
	}

	global.GVA_LOG.Info("预览描述生成成功", 
		zap.Uint("requirementID", req.RequirementID),
	)

	response.OkWithData(gin.H{
		"preview":    results[0].GeneratedDescription,
		"confidence": results[0].AnalysisConfidence,
		"suggestions": results[0].ImprovementSuggestions,
		"processing_time": results[0].ProcessingTime,
	}, c)
}

// ValidateDescription 验证描述质量
// @Tags DescriptionGenerator
// @Summary 验证描述质量
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ValidateDescriptionRequest true "验证请求"
// @Success 200 {object} response.Response{} "验证成功"
// @Router /nesma/generator/description/validate [post]
func (a *DescriptionGeneratorApi) ValidateDescription(c *gin.Context) {
	var req nesmaReq.ValidateDescriptionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.EnhancedDescription == nil {
		response.FailWithMessage("增强描述不能为空", c)
		return
	}

	// 执行验证逻辑
	validationResult := a.validateDescription(req.EnhancedDescription)

	response.OkWithData(gin.H{
		"is_valid":         validationResult.IsValid,
		"quality_score":    validationResult.QualityScore,
		"completeness":     validationResult.Completeness,
		"issues":           validationResult.Issues,
		"suggestions":      validationResult.Suggestions,
		"quality_metrics":  validationResult.QualityMetrics,
	}, c)
}

// validateDescription 验证描述的逻辑
func (a *DescriptionGeneratorApi) validateDescription(desc *nesmaReq.EnhancedDescription) *DescriptionValidationResult {
	var issues []string
	var suggestions []string
	var qualityScore float64 = 1.0
	var completeness float64 = 0.0

	// 验证基本信息
	if desc.Title == "" {
		issues = append(issues, "标题不能为空")
		qualityScore -= 0.2
	} else {
		completeness += 0.1
	}

	if desc.DetailedDescription == "" {
		issues = append(issues, "详细描述不能为空")
		qualityScore -= 0.2
	} else {
		completeness += 0.1
	}

	// 验证功能流程
	if len(desc.FunctionalFlow.ProcessingSteps) == 0 {
		issues = append(issues, "缺少处理步骤")
		qualityScore -= 0.1
	} else {
		completeness += 0.1
	}

	// 验证数据元素
	if len(desc.DataElements) == 0 {
		issues = append(issues, "缺少数据元素定义")
		qualityScore -= 0.1
	} else {
		completeness += 0.1
	}

	// 验证业务规则
	if len(desc.BusinessRules) == 0 {
		issues = append(issues, "缺少业务规则")
		qualityScore -= 0.1
	} else {
		completeness += 0.1
	}

	// 验证验收标准
	if len(desc.AcceptanceCriteria) == 0 {
		issues = append(issues, "缺少验收标准")
		qualityScore -= 0.1
	} else {
		completeness += 0.1
	}

	// 验证测试场景
	if len(desc.TestScenarios) == 0 {
		issues = append(issues, "缺少测试场景")
		qualityScore -= 0.1
	} else {
		completeness += 0.1
	}

	// 生成改进建议
	if len(issues) > 0 {
		suggestions = append(suggestions, "建议补充缺失的内容以提高描述完整性")
	}

	if len(desc.DataElements) > 0 && len(desc.DataElements) < 3 {
		suggestions = append(suggestions, "建议增加更多数据元素定义")
	}

	if len(desc.BusinessRules) > 0 && len(desc.BusinessRules) < 2 {
		suggestions = append(suggestions, "建议添加更多业务规则")
	}

	// 质量指标
	qualityMetrics := map[string]interface{}{
		"data_elements_count":    len(desc.DataElements),
		"business_rules_count":   len(desc.BusinessRules),
		"acceptance_criteria_count": len(desc.AcceptanceCriteria),
		"test_scenarios_count":   len(desc.TestScenarios),
		"description_length":     len(desc.DetailedDescription),
		"has_business_context":   desc.BusinessContext != "",
	}

	return &DescriptionValidationResult{
		IsValid:         len(issues) == 0,
		QualityScore:    qualityScore,
		Completeness:    completeness,
		Issues:          issues,
		Suggestions:     suggestions,
		QualityMetrics:  qualityMetrics,
	}
}

// DescriptionValidationResult 描述验证结果
type DescriptionValidationResult struct {
	IsValid         bool                   `json:"is_valid"`
	QualityScore    float64                `json:"quality_score"`
	Completeness    float64                `json:"completeness"`
	Issues          []string               `json:"issues"`
	Suggestions     []string               `json:"suggestions"`
	QualityMetrics  map[string]interface{} `json:"quality_metrics"`
}