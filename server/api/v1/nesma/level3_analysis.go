package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// 类型转换函数
func convertOptimizationSuggestion(req *nesmaReq.OptimizationSuggestion) *nesmaService.OptimizationSuggestion {
	if req == nil {
		return nil
	}
	return &nesmaService.OptimizationSuggestion{
		Type:           req.Type,
		Field:          req.Field,
		CurrentValue:   req.CurrentValue,
		SuggestedValue: req.SuggestedValue,
		Reason:         req.Reason,
		Priority:       req.Priority,
		Confidence:     req.Confidence,
		Category:       req.Category,
	}
}

func convertExpansionSuggestion(req *nesmaReq.ExpansionSuggestion) *nesmaService.ExpansionSuggestion {
	if req == nil {
		return nil
	}
	return &nesmaService.ExpansionSuggestion{
		SuggestedTitle:       req.SuggestedTitle,
		SuggestedDescription: req.SuggestedDescription,
		SuggestedCategory:    req.SuggestedCategory,
		Justification:        req.Justification,
		Priority:             req.Priority,
		Confidence:           req.Confidence,
		RelatedKnowledge:     req.RelatedKnowledge,
		EstimatedComplexity:  req.EstimatedComplexity,
	}
}

type Level3AnalysisApi struct{}

// AnalyzeLevel3Requirements 分析三级功能点
// @Tags Level3Analysis
// @Summary 分析三级功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.Level3AnalysisRequest true "分析请求"
// @Success 200 {object} response.Response{data=[]nesmaService.Level3AnalysisResult} "分析成功"
// @Router /nesma/analysis/level3/analyze [post]
func (a *Level3AnalysisApi) AnalyzeLevel3Requirements(c *gin.Context) {
	var req nesmaReq.Level3AnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	analyzerService := nesmaService.GetLevel3AnalyzerService()
	
	// 执行分析
	results, err := analyzerService.AnalyzeLevel3Requirements(req.CycleID, req.RequirementIDs)
	if err != nil {
		global.GVA_LOG.Error("三级功能点分析失败", zap.Error(err))
		response.FailWithMessage("分析失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("三级功能点分析完成", 
		zap.Uint("cycleID", req.CycleID),
		zap.Int("结果数量", len(results)),
	)

	response.OkWithData(gin.H{
		"results": results,
		"total":   len(results),
	}, c)
}

// ApplyOptimizationSuggestion 应用优化建议
// @Tags Level3Analysis
// @Summary 应用优化建议
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ApplyOptimizationRequest true "应用优化请求"
// @Success 200 {object} response.Response{} "应用成功"
// @Router /nesma/analysis/level3/apply-optimization [post]
func (a *Level3AnalysisApi) ApplyOptimizationSuggestion(c *gin.Context) {
	var req nesmaReq.ApplyOptimizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.RequirementID == 0 {
		response.FailWithMessage("功能点ID不能为空", c)
		return
	}

	if req.Suggestion == nil {
		response.FailWithMessage("优化建议不能为空", c)
		return
	}

	analyzerService := nesmaService.GetLevel3AnalyzerService()
	
	// 应用优化建议
	err := analyzerService.ApplyOptimizationSuggestion(req.RequirementID, convertOptimizationSuggestion(req.Suggestion))
	if err != nil {
		global.GVA_LOG.Error("应用优化建议失败", zap.Error(err))
		response.FailWithMessage("应用失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("优化建议应用成功", 
		zap.Uint("requirementID", req.RequirementID),
		zap.String("type", req.Suggestion.Type),
	)

	response.OkWithMessage("优化建议应用成功", c)
}

// CreateExpansionRequirement 创建扩充功能点
// @Tags Level3Analysis
// @Summary 创建扩充功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateExpansionRequest true "创建扩充请求"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirement} "创建成功"
// @Router /nesma/analysis/level3/create-expansion [post]
func (a *Level3AnalysisApi) CreateExpansionRequirement(c *gin.Context) {
	var req nesmaReq.CreateExpansionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	if req.ParentID == 0 {
		response.FailWithMessage("父功能点ID不能为空", c)
		return
	}

	if req.Suggestion == nil {
		response.FailWithMessage("扩充建议不能为空", c)
		return
	}

	analyzerService := nesmaService.GetLevel3AnalyzerService()
	
	// 创建扩充功能点
	newRequirement, err := analyzerService.CreateExpansionRequirement(req.CycleID, req.ParentID, convertExpansionSuggestion(req.Suggestion))
	if err != nil {
		global.GVA_LOG.Error("创建扩充功能点失败", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("扩充功能点创建成功", 
		zap.Uint("cycleID", req.CycleID),
		zap.Uint("parentID", req.ParentID),
		zap.Uint("newRequirementID", newRequirement.ID),
	)

	response.OkWithData(newRequirement, c)
}

// BatchApplyOptimizations 批量应用优化建议
// @Tags Level3Analysis
// @Summary 批量应用优化建议
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchApplyOptimizationsRequest true "批量应用请求"
// @Success 200 {object} response.Response{} "批量应用成功"
// @Router /nesma/analysis/level3/batch-apply-optimizations [post]
func (a *Level3AnalysisApi) BatchApplyOptimizations(c *gin.Context) {
	var req nesmaReq.BatchApplyOptimizationsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if len(req.Applications) == 0 {
		response.FailWithMessage("应用列表不能为空", c)
		return
	}

	analyzerService := nesmaService.GetLevel3AnalyzerService()
	
	var successCount int
	var failedCount int
	var errors []string

	// 批量应用优化建议
	for _, app := range req.Applications {
		err := analyzerService.ApplyOptimizationSuggestion(app.RequirementID, convertOptimizationSuggestion(app.Suggestion))
		if err != nil {
			failedCount++
			errors = append(errors, err.Error())
			global.GVA_LOG.Error("批量应用优化建议失败", 
				zap.Uint("requirementID", app.RequirementID),
				zap.Error(err),
			)
		} else {
			successCount++
		}
	}

	global.GVA_LOG.Info("批量应用优化建议完成", 
		zap.Int("成功数量", successCount),
		zap.Int("失败数量", failedCount),
	)

	response.OkWithData(gin.H{
		"success_count": successCount,
		"failed_count":  failedCount,
		"errors":        errors,
	}, c)
}

// BatchCreateExpansions 批量创建扩充功能点
// @Tags Level3Analysis
// @Summary 批量创建扩充功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchCreateExpansionsRequest true "批量创建请求"
// @Success 200 {object} response.Response{} "批量创建成功"
// @Router /nesma/analysis/level3/batch-create-expansions [post]
func (a *Level3AnalysisApi) BatchCreateExpansions(c *gin.Context) {
	var req nesmaReq.BatchCreateExpansionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if len(req.Expansions) == 0 {
		response.FailWithMessage("扩充列表不能为空", c)
		return
	}

	analyzerService := nesmaService.GetLevel3AnalyzerService()
	
	var successCount int
	var failedCount int
	var errors []string
	var createdRequirements []uint

	// 批量创建扩充功能点
	for _, exp := range req.Expansions {
		newRequirement, err := analyzerService.CreateExpansionRequirement(exp.CycleID, exp.ParentID, convertExpansionSuggestion(exp.Suggestion))
		if err != nil {
			failedCount++
			errors = append(errors, err.Error())
			global.GVA_LOG.Error("批量创建扩充功能点失败", 
				zap.Uint("cycleID", exp.CycleID),
				zap.Uint("parentID", exp.ParentID),
				zap.Error(err),
			)
		} else {
			successCount++
			createdRequirements = append(createdRequirements, newRequirement.ID)
		}
	}

	global.GVA_LOG.Info("批量创建扩充功能点完成", 
		zap.Int("成功数量", successCount),
		zap.Int("失败数量", failedCount),
	)

	response.OkWithData(gin.H{
		"success_count":         successCount,
		"failed_count":          failedCount,
		"created_requirements":  createdRequirements,
		"errors":                errors,
	}, c)
}

// GetAnalysisHistory 获取分析历史
// @Tags Level3Analysis
// @Summary 获取分析历史
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/analysis/level3/history/{cycleId} [get]
func (a *Level3AnalysisApi) GetAnalysisHistory(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	_, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// TODO: 实现分析历史查询逻辑
	// 这里先返回模拟数据
	response.OkWithData(gin.H{
		"list":     []interface{}{},
		"total":    0,
		"page":     page,
		"pageSize": pageSize,
	}, c)
}

// GetAnalysisStats 获取分析统计
// @Tags Level3Analysis
// @Summary 获取分析统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/analysis/level3/stats/{cycleId} [get]
func (a *Level3AnalysisApi) GetAnalysisStats(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// TODO: 实现统计查询逻辑
	// 这里先返回模拟数据
	response.OkWithData(gin.H{
		"cycle_id":               cycleID,
		"total_requirements":     0,
		"analyzed_requirements":  0,
		"optimization_suggestions": 0,
		"expansion_suggestions":  0,
		"applied_optimizations":  0,
		"created_expansions":     0,
		"average_confidence":     0.0,
	}, c)
}