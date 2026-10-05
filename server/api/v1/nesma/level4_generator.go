package nesma

import (
	"encoding/json"
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

// 类型转换函数
func convertLevel4Suggestion(req *nesmaReq.Level4Suggestion) *nesmaService.Level4Suggestion {
	if req == nil {
		return nil
	}
	return &nesmaService.Level4Suggestion{
		SuggestedTitle:       req.SuggestedTitle,
		SuggestedDescription: req.SuggestedDescription,
		SuggestedCode:        req.SuggestedCode,
		FunctionType:         req.FunctionType,
		BusinessValue:        req.BusinessValue,
		AcceptanceCriteria:   req.AcceptanceCriteria,
		EstimatedComplexity:  req.EstimatedComplexity,
		RecommendedAFP:       req.RecommendedAFP,
		RecommendedUFP:       req.RecommendedUFP,
		Priority:             req.Priority,
		Confidence:           req.Confidence,
		GenerationReason:     req.GenerationReason,
		RelatedKnowledge:     req.RelatedKnowledge,
	}
}

func convertLevel4CreationRequest(req nesmaReq.Level4CreationRequest) nesmaService.Level4CreationRequest {
	return nesmaService.Level4CreationRequest{
		ParentID:   req.ParentID,
		Suggestion: convertLevel4Suggestion(req.Suggestion),
	}
}

func convertLevel4CreationRequests(reqs []nesmaReq.Level4CreationRequest) []nesmaService.Level4CreationRequest {
	var result []nesmaService.Level4CreationRequest
	for _, req := range reqs {
		result = append(result, convertLevel4CreationRequest(req))
	}
	return result
}

type Level4GeneratorApi struct{}

// GenerateLevel4Requirements 生成四级功能点
// @Tags Level4Generator
// @Summary 生成四级功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.Level4GenerationRequest true "生成请求"
// @Success 200 {object} response.Response{data=[]nesmaService.Level4GenerationResult} "生成成功"
// @Router /nesma/generator/level4/generate [post]
func (a *Level4GeneratorApi) GenerateLevel4Requirements(c *gin.Context) {
	var req nesmaReq.Level4GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	generatorService := nesmaService.GetLevel4GeneratorService()
	
	// 执行生成
	results, err := generatorService.GenerateLevel4Requirements(req.CycleID, req.Level3RequirementIDs)
	if err != nil {
		global.GVA_LOG.Error("四级功能点生成失败", zap.Error(err))
		response.FailWithMessage("生成失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("四级功能点生成完成", 
		zap.Uint("cycleID", req.CycleID),
		zap.Int("结果数量", len(results)),
	)

	response.OkWithData(gin.H{
		"results": results,
		"total":   len(results),
	}, c)
}

// CreateLevel4Requirement 创建四级功能点
// @Tags Level4Generator
// @Summary 创建四级功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateLevel4Request true "创建请求"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirement} "创建成功"
// @Router /nesma/generator/level4/create [post]
func (a *Level4GeneratorApi) CreateLevel4Requirement(c *gin.Context) {
	var req nesmaReq.CreateLevel4Request
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
		response.FailWithMessage("功能点建议不能为空", c)
		return
	}

	generatorService := nesmaService.GetLevel4GeneratorService()
	
	// 创建四级功能点
	newRequirement, err := generatorService.CreateLevel4Requirement(req.CycleID, req.ParentID, convertLevel4Suggestion(req.Suggestion))
	if err != nil {
		global.GVA_LOG.Error("创建四级功能点失败", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("四级功能点创建成功", 
		zap.Uint("cycleID", req.CycleID),
		zap.Uint("parentID", req.ParentID),
		zap.Uint("newRequirementID", newRequirement.ID),
	)

	response.OkWithData(newRequirement, c)
}

// BatchCreateLevel4Requirements 批量创建四级功能点
// @Tags Level4Generator
// @Summary 批量创建四级功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchCreateLevel4Request true "批量创建请求"
// @Success 200 {object} response.Response{} "批量创建成功"
// @Router /nesma/generator/level4/batch-create [post]
func (a *Level4GeneratorApi) BatchCreateLevel4Requirements(c *gin.Context) {
	var req nesmaReq.BatchCreateLevel4Request
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.CycleID == 0 {
		response.FailWithMessage("项目周期ID不能为空", c)
		return
	}

	if len(req.Creations) == 0 {
		response.FailWithMessage("创建列表不能为空", c)
		return
	}

	generatorService := nesmaService.GetLevel4GeneratorService()
	
	// 批量创建四级功能点
	result, err := generatorService.BatchCreateLevel4Requirements(req.CycleID, convertLevel4CreationRequests(req.Creations))
	if err != nil {
		global.GVA_LOG.Error("批量创建四级功能点失败", zap.Error(err))
		response.FailWithMessage("批量创建失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("批量创建四级功能点完成", 
		zap.Int("成功数量", result.SuccessCount),
		zap.Int("失败数量", result.FailedCount),
	)

	response.OkWithData(result, c)
}

// GetLevel4GenerationHistory 获取四级功能点生成历史
// @Tags Level4Generator
// @Summary 获取四级功能点生成历史
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Param level3Id query int false "三级功能点ID"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/generator/level4/history/{cycleId} [get]
func (a *Level4GeneratorApi) GetLevel4GenerationHistory(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取查询参数
	level3IDStr := c.Query("level3Id")
	var level3ID uint
	if level3IDStr != "" {
		if id, err := strconv.ParseUint(level3IDStr, 10, 32); err == nil {
			level3ID = uint(id)
		}
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))

	// TODO: 实现生成历史查询逻辑
	// 这里先返回模拟数据
	response.OkWithData(gin.H{
		"list":      []interface{}{},
		"total":     0,
		"page":      page,
		"pageSize":  pageSize,
		"cycleId":   uint(cycleID),
		"level3Id":  level3ID,
	}, c)
}

// GetLevel4GenerationStats 获取四级功能点生成统计
// @Tags Level4Generator
// @Summary 获取四级功能点生成统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Success 200 {object} response.Response{} "获取成功"
// @Router /nesma/generator/level4/stats/{cycleId} [get]
func (a *Level4GeneratorApi) GetLevel4GenerationStats(c *gin.Context) {
	cycleIDStr := c.Param("cycleId")
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	// 获取统计信息
	var totalLevel3Count int64
	var totalLevel4Count int64
	var generatedLevel4Count int64
	var avgConfidence float64

	// 统计三级功能点总数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 3", cycleID).
		Count(&totalLevel3Count)

	// 统计四级功能点总数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4", cycleID).
		Count(&totalLevel4Count)

	// 统计AI生成的四级功能点数
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4 AND ai_analysis_status = 'generated'", cycleID).
		Count(&generatedLevel4Count)

	// 计算平均置信度
	var confidenceResult struct {
		AvgConfidence float64 `json:"avg_confidence"`
	}
	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4 AND ai_confidence_score IS NOT NULL", cycleID).
		Select("AVG(ai_confidence_score) as avg_confidence").
		Scan(&confidenceResult)
	
	avgConfidence = confidenceResult.AvgConfidence

	response.OkWithData(gin.H{
		"cycle_id":                cycleID,
		"total_level3_count":      totalLevel3Count,
		"total_level4_count":      totalLevel4Count,
		"generated_level4_count":  generatedLevel4Count,
		"generation_coverage":     func() float64 {
			if totalLevel3Count == 0 {
				return 0
			}
			return float64(generatedLevel4Count) / float64(totalLevel3Count)
		}(),
		"average_confidence":      avgConfidence,
		"function_type_distribution": a.getFunctionTypeDistribution(uint(cycleID)),
	}, c)
}

// getFunctionTypeDistribution 获取功能类型分布
func (a *Level4GeneratorApi) getFunctionTypeDistribution(cycleID uint) map[string]int64 {
	var results []struct {
		FunctionType string `json:"function_type"`
		Count        int64  `json:"count"`
	}

	global.GVA_DB.Model(&nesmaModel.NesmaRequirement{}).
		Where("cycle_id = ? AND level = 4 AND function_type IS NOT NULL", cycleID).
		Select("function_type, COUNT(*) as count").
		Group("function_type").
		Scan(&results)

	distribution := make(map[string]int64)
	for _, result := range results {
		distribution[result.FunctionType] = result.Count
	}

	return distribution
}

// ValidateLevel4Suggestion 验证四级功能点建议
// @Tags Level4Generator
// @Summary 验证四级功能点建议
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ValidateLevel4SuggestionRequest true "验证请求"
// @Success 200 {object} response.Response{} "验证成功"
// @Router /nesma/generator/level4/validate [post]
func (a *Level4GeneratorApi) ValidateLevel4Suggestion(c *gin.Context) {
	var req nesmaReq.ValidateLevel4SuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证必要参数
	if req.Suggestion == nil {
		response.FailWithMessage("功能点建议不能为空", c)
		return
	}

	// 执行验证逻辑
	validationResult := a.validateSuggestion(convertLevel4Suggestion(req.Suggestion))

	response.OkWithData(gin.H{
		"is_valid":         validationResult.IsValid,
		"validation_score": validationResult.ValidationScore,
		"issues":           validationResult.Issues,
		"suggestions":      validationResult.Suggestions,
	}, c)
}

// validateSuggestion 验证建议的逻辑
func (a *Level4GeneratorApi) validateSuggestion(suggestion *nesmaService.Level4Suggestion) *ValidationResult {
	var issues []string
	var suggestions []string
	score := 1.0

	// 验证标题
	if suggestion.SuggestedTitle == "" {
		issues = append(issues, "标题不能为空")
		score -= 0.2
	} else if len(suggestion.SuggestedTitle) < 5 {
		issues = append(issues, "标题过短")
		score -= 0.1
	}

	// 验证描述
	if suggestion.SuggestedDescription == "" {
		issues = append(issues, "描述不能为空")
		score -= 0.2
	} else if len(suggestion.SuggestedDescription) < 20 {
		issues = append(issues, "描述过于简单")
		score -= 0.1
	}

	// 验证功能类型
	validTypes := []string{"EI", "EO", "EQ", "ILF", "EIF"}
	isValidType := false
	for _, validType := range validTypes {
		if suggestion.FunctionType == validType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		issues = append(issues, "功能类型不符合NESMA标准")
		score -= 0.15
	}

	// 验证复杂度
	validComplexities := []string{"简单", "中等", "复杂"}
	isValidComplexity := false
	for _, validComplexity := range validComplexities {
		if suggestion.EstimatedComplexity == validComplexity {
			isValidComplexity = true
			break
		}
	}
	if !isValidComplexity {
		issues = append(issues, "复杂度评估不规范")
		score -= 0.1
	}

	// 验证功能点数
	if suggestion.RecommendedAFP <= 0 || suggestion.RecommendedUFP <= 0 {
		issues = append(issues, "功能点数必须大于0")
		score -= 0.15
	}

	// 验证置信度
	if suggestion.Confidence < 0 || suggestion.Confidence > 1 {
		issues = append(issues, "置信度必须在0-1之间")
		score -= 0.1
	}

	// 生成建议
	if len(issues) > 0 {
		suggestions = append(suggestions, "建议完善功能点描述，确保包含输入、处理、输出等关键信息")
		suggestions = append(suggestions, "建议参考NESMA标准进行功能类型分类")
		suggestions = append(suggestions, "建议添加更详细的验收标准")
	}

	return &ValidationResult{
		IsValid:         len(issues) == 0,
		ValidationScore: score,
		Issues:          issues,
		Suggestions:     suggestions,
	}
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid         bool     `json:"is_valid"`
	ValidationScore float64  `json:"validation_score"`
	Issues          []string `json:"issues"`
	Suggestions     []string `json:"suggestions"`
}

// ==================== 异步L4生成API ====================

// GenerateLevel4Async 异步生成四级功能点
// @Tags Level4Generator
// @Summary 异步生成四级功能点
// @Description 异步生成四级功能点，支持批量处理和实时进度跟踪
// @Accept application/json
// @Produce application/json
// @Param data body nesmaReq.AsyncLevel4GenerationRequest true "生成参数"
// @Success 200 {object} response.Response{data=gin.H} "任务创建成功，返回taskId"
// @Router /nesma/generator/level4/generate-async [post]
func (a *Level4GeneratorApi) GenerateLevel4Async(c *gin.Context) {
	var req nesmaReq.AsyncLevel4GenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置默认值
	if req.GenerationStrategy == "" {
		req.GenerationStrategy = "comprehensive"
	}
	if req.ComplexityLevel == "" {
		req.ComplexityLevel = "moderate"
	}
	if req.MaxL4Count == 0 {
		req.MaxL4Count = 8
	}

	global.GVA_LOG.Info("收到异步L4生成请求", 
		zap.Uint("projectId", req.ProjectID),
		zap.Uint("cycleId", req.CycleID),
		zap.Int("l3Count", len(req.L3RequirementIDs)))

	// 调用服务
	generatorService := nesmaService.GetLevel4GeneratorService()
	task, err := generatorService.GenerateLevel4Async(&req)
	if err != nil {
		global.GVA_LOG.Error("异步L4生成失败", zap.Error(err))
		response.FailWithMessage("异步L4生成失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"taskId":      task.ID,
		"message":     "L4生成任务已创建，正在后台处理",
		"totalCount":  task.TotalCount,
		"status":      task.Status,
	}, c)
}

// GetLevel4GenerationTaskProgress 获取L4生成任务进度
// @Tags Level4Generator
// @Summary 获取L4生成任务进度
// @Description 获取异步L4生成任务的实时进度和状态
// @Accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=nesmaModel.NesmaRequirementAnalysisTask} "获取成功"
// @Router /nesma/generator/level4/task/{taskId}/progress [get]
func (a *Level4GeneratorApi) GetLevel4GenerationTaskProgress(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'level4_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	response.OkWithData(task, c)
}

// GetLevel4GenerationTaskResult 获取L4生成任务结果
// @Tags Level4Generator
// @Summary 获取L4生成任务结果
// @Description 获取已完成的L4生成任务的详细结果
// @Accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/level4/task/{taskId}/result [get]
func (a *Level4GeneratorApi) GetLevel4GenerationTaskResult(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'level4_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	if task.Status != "completed" {
		response.FailWithMessage("任务尚未完成", c)
		return
	}

	// 获取生成的L4需求列表
	var createdRequirements []nesmaModel.NesmaRequirement
	if task.Result != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(task.Result, &result); err == nil {
			if createdIDs, ok := result["created_requirements"].([]interface{}); ok {
				var ids []uint
				for _, id := range createdIDs {
					if intId, ok := id.(float64); ok {
						ids = append(ids, uint(intId))
					}
				}
				if len(ids) > 0 {
					global.GVA_DB.Where("id IN ?", ids).Find(&createdRequirements)
				}
			}
		}
	}

	response.OkWithData(gin.H{
		"task":                task,
		"createdRequirements": createdRequirements,
	}, c)
}

// ConfirmLevel4Requirements 确认并入库L4需求
// @Tags Level4Generator
// @Summary 确认并入库L4需求
// @Description 用户确认AI生成的L4建议后，入库到对应的L3需求下
// @Accept application/json
// @Produce application/json
// @Param data body nesmaReq.ConfirmLevel4RequirementsRequest true "确认请求"
// @Success 200 {object} response.Response{data=gin.H} "确认成功"
// @Router /nesma/generator/level4/confirm [post]
func (a *Level4GeneratorApi) ConfirmLevel4Requirements(c *gin.Context) {
	var req nesmaReq.ConfirmLevel4RequirementsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("收到L4需求确认请求", zap.Int("l4Count", len(req.L4Requirements)))

	// 调用服务
	generatorService := nesmaService.GetLevel4GeneratorService()
	result, err := generatorService.ConfirmLevel4Requirements(&req)
	if err != nil {
		global.GVA_LOG.Error("L4需求确认失败", zap.Error(err))
		response.FailWithMessage("L4需求确认失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"successCount":        result.SuccessCount,
		"failedCount":         result.FailedCount,
		"createdRequirements": result.CreatedRequirements,
		"errors":              result.Errors,
		"message":             fmt.Sprintf("成功确认 %d 个L4需求，失败 %d 个", result.SuccessCount, result.FailedCount),
	}, c)
}

// ==================== L4任务管理API ====================

// GetL4GenerationTasks 获取L4生成任务列表
// @Tags Level4Generator
// @Summary 获取L4生成任务列表
// @Description 获取L4生成任务列表，支持分页和筛选
// @Accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID"
// @Param cycleId query int false "周期ID"
// @Param versionId query int false "版本ID"
// @Param generationStrategy query string false "生成策略"
// @Param status query string false "任务状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/level4/tasks [get]
func (a *Level4GeneratorApi) GetL4GenerationTasks(c *gin.Context) {
	// 获取查询参数
	projectIDStr := c.Query("projectId")
	cycleIDStr := c.Query("cycleId")
	versionIDStr := c.Query("versionId")
	generationStrategy := c.Query("generationStrategy")
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	
	// 构建查询条件
	db := global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{})
	db = db.Where("task_type = ?", "level4_generation")
	
	if projectIDStr != "" {
		if projectID, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			db = db.Where("project_id = ?", projectID)
		}
	}
	
	if cycleIDStr != "" {
		if cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32); err == nil {
			db = db.Where("cycle_id = ?", cycleID)
		}
	}
	
	if versionIDStr != "" {
		if versionID, err := strconv.ParseUint(versionIDStr, 10, 32); err == nil {
			db = db.Where("source_version_id = ?", versionID)
		}
	}
	
	if generationStrategy != "" {
		db = db.Where("JSON_EXTRACT(result, '$.generation_strategy') = ?", generationStrategy)
	}
	
	if status != "" {
		db = db.Where("status = ?", status)
	}
	
	// 获取总数
	var total int64
	db.Count(&total)
	
	// 分页查询
	var tasks []nesmaModel.NesmaRequirementAnalysisTask
	offset := (page - 1) * pageSize
	db = db.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tasks)
	
	// 转换为前端需要的格式
	var taskList []gin.H
	for _, task := range tasks {
		taskItem := gin.H{
			"ID":                task.ID,
			"projectId":         task.ProjectID,
			"cycleId":           task.CycleID,
			"versionId":         task.SourceVersionID,
			"status":            task.Status,
			"progress":          task.Progress,
			"createdAt":         task.CreatedAt,
			"updatedAt":         task.UpdatedAt,
			"totalCount":        task.TotalCount,
			"successCount":      task.SuccessCount,
			"failedCount":       task.FailedCount,
			"generationStrategy": "comprehensive", // 默认值
			"l3RequirementCount": 0,
			"l3RequirementTitles": []string{},
			"totalL4Count":      0,
			"duration":          nil,
			"completedAt":       nil,
		}
		
		// 解析任务结果获取更多信息
		if task.Result != nil {
			var result map[string]interface{}
			if err := json.Unmarshal(task.Result, &result); err == nil {
				if strategy, ok := result["generation_strategy"].(string); ok {
					taskItem["generationStrategy"] = strategy
				}
				if l3Count, ok := result["l3_requirement_count"].(float64); ok {
					taskItem["l3RequirementCount"] = int(l3Count)
				}
				if l3Titles, ok := result["l3_requirement_titles"].([]interface{}); ok {
					titles := make([]string, 0, len(l3Titles))
					for _, title := range l3Titles {
						if titleStr, ok := title.(string); ok {
							titles = append(titles, titleStr)
						}
					}
					taskItem["l3RequirementTitles"] = titles
				}
				if l4Count, ok := result["total_l4_count"].(float64); ok {
					taskItem["totalL4Count"] = int(l4Count)
				}
				if completedAtStr, ok := result["completed_at"].(string); ok {
					taskItem["completedAt"] = completedAtStr
				}
				if durationVal, ok := result["duration"].(float64); ok {
					taskItem["duration"] = int(durationVal)
				}
			}
		}
		
		taskList = append(taskList, taskItem)
	}
	
	response.OkWithData(gin.H{
		"list":     taskList,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, c)
}

// GetL4TaskStatistics 获取L4任务统计信息
// @Tags Level4Generator
// @Summary 获取L4任务统计信息
// @Description 获取L4生成任务的统计信息
// @Accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID"
// @Param cycleId query int false "周期ID"
// @Param versionId query int false "版本ID"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/level4/tasks/statistics [get]
func (a *Level4GeneratorApi) GetL4TaskStatistics(c *gin.Context) {
	// 获取查询参数
	projectIDStr := c.Query("projectId")
	cycleIDStr := c.Query("cycleId")
	versionIDStr := c.Query("versionId")
	
	// 构建基础查询条件
	db := global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{})
	db = db.Where("task_type = ?", "level4_generation")
	
	if projectIDStr != "" {
		if projectID, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			db = db.Where("project_id = ?", projectID)
		}
	}
	
	if cycleIDStr != "" {
		if cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32); err == nil {
			db = db.Where("cycle_id = ?", cycleID)
		}
	}
	
	if versionIDStr != "" {
		if versionID, err := strconv.ParseUint(versionIDStr, 10, 32); err == nil {
			db = db.Where("source_version_id = ?", versionID)
		}
	}
	
	// 统计各种状态的任务数量
	var total int64
	var pending int64
	var running int64
	var completed int64
	var failed int64
	var cancelled int64
	
	db.Count(&total)
	db.Where("status = ?", "pending").Count(&pending)
	db.Where("status = ?", "running").Count(&running)
	db.Where("status = ?", "completed").Count(&completed)
	db.Where("status = ?", "failed").Count(&failed)
	db.Where("status = ?", "cancelled").Count(&cancelled)
	
	response.OkWithData(gin.H{
		"total":     total,
		"pending":   pending,
		"running":   running,
		"completed": completed,
		"failed":    failed,
		"cancelled": cancelled,
	}, c)
}

// GetL4TaskDetail 获取L4任务详情
// @Tags Level4Generator
// @Summary 获取L4任务详情
// @Description 获取L4生成任务的详细信息
// @Accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "获取成功"
// @Router /nesma/generator/level4/task/{taskId}/detail [get]
func (a *Level4GeneratorApi) GetL4TaskDetail(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}
	
	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'level4_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	
	// 构建详情数据
	taskDetail := gin.H{
		"ID":           task.ID,
		"projectId":    task.ProjectID,
		"cycleId":      task.CycleID,
		"versionId":    task.SourceVersionID,
		"status":       task.Status,
		"progress":     task.Progress,
		"createdAt":    task.CreatedAt,
		"updatedAt":    task.UpdatedAt,
		"totalCount":   task.TotalCount,
		"successCount": task.SuccessCount,
		"failedCount":  task.FailedCount,
		"errorMsg":     "",
		"configDetails": gin.H{
			"generationStrategy":  "comprehensive",
			"maxL4PerL3":         5,
			"confidenceThreshold": 0.7,
			"useKnowledgeBase":   true,
			"enableOptimization": true,
		},
	}
	
	// 解析任务结果获取更多信息
	if task.Result != nil {
		var result map[string]interface{}
		if err := json.Unmarshal(task.Result, &result); err == nil {
			taskDetail["configDetails"] = result
		}
	}
	
	response.OkWithData(taskDetail, c)
}

// CancelL4GenerationTask 取消L4生成任务
// @Tags Level4Generator
// @Summary 取消L4生成任务
// @Description 取消正在运行的L4生成任务
// @Accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "取消成功"
// @Router /nesma/generator/level4/task/{taskId}/cancel [post]
func (a *Level4GeneratorApi) CancelL4GenerationTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}
	
	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'level4_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	
	// 只能取消运行中或等待中的任务
	if task.Status != "running" && task.Status != "pending" {
		response.FailWithMessage("只能取消运行中或等待中的任务", c)
		return
	}
	
	// 更新任务状态为已取消
	if err := global.GVA_DB.Model(&task).Update("status", "cancelled").Error; err != nil {
		response.FailWithMessage("取消任务失败", c)
		return
	}
	
	response.OkWithData(gin.H{
		"message": "任务已取消",
		"taskId":  taskID,
	}, c)
}

// RetryL4GenerationTask 重试L4生成任务
// @Tags Level4Generator
// @Summary 重试L4生成任务
// @Description 重试失败的L4生成任务
// @Accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "重试成功"
// @Router /nesma/generator/level4/task/{taskId}/retry [post]
func (a *Level4GeneratorApi) RetryL4GenerationTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}
	
	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'level4_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	
	// 只能重试失败的任务
	if task.Status != "failed" {
		response.FailWithMessage("只能重试失败的任务", c)
		return
	}
	
	// 重置任务状态
	updates := map[string]interface{}{
		"status":       "pending",
		"progress":     0,
		"success_count": 0,
		"failed_count":  0,
	}
	
	if err := global.GVA_DB.Model(&task).Updates(updates).Error; err != nil {
		response.FailWithMessage("重试任务失败", c)
		return
	}
	
	response.OkWithData(gin.H{
		"message": "任务已重置为等待状态",
		"taskId":  taskID,
	}, c)
}

// DeleteL4GenerationTask 删除L4生成任务
// @Tags Level4Generator
// @Summary 删除L4生成任务
// @Description 删除L4生成任务记录
// @Accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=gin.H} "删除成功"
// @Router /nesma/generator/level4/task/{taskId} [delete]
func (a *Level4GeneratorApi) DeleteL4GenerationTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}
	
	var task nesmaModel.NesmaRequirementAnalysisTask
	if err := global.GVA_DB.Where("id = ? AND task_type = 'level4_generation'", taskID).First(&task).Error; err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}
	
	// 不能删除正在运行的任务
	if task.Status == "running" {
		response.FailWithMessage("不能删除正在运行的任务", c)
		return
	}
	
	// 删除任务
	if err := global.GVA_DB.Delete(&task).Error; err != nil {
		response.FailWithMessage("删除任务失败", c)
		return
	}
	
	response.OkWithData(gin.H{
		"message": "任务已删除",
		"taskId":  taskID,
	}, c)
}

// BatchDeleteL4GenerationTasks 批量删除L4生成任务
// @Tags Level4Generator
// @Summary 批量删除L4生成任务
// @Description 批量删除L4生成任务记录
// @Accept application/json
// @Produce application/json
// @Param data body gin.H true "任务ID列表"
// @Success 200 {object} response.Response{data=gin.H} "删除成功"
// @Router /nesma/generator/level4/tasks/batch-delete [post]
func (a *Level4GeneratorApi) BatchDeleteL4GenerationTasks(c *gin.Context) {
	var req struct {
		TaskIDs []uint `json:"task_ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	
	if len(req.TaskIDs) == 0 {
		response.FailWithMessage("任务ID列表不能为空", c)
		return
	}
	
	// 检查是否有正在运行的任务
	var runningCount int64
	global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{}).
		Where("id IN ? AND task_type = 'level4_generation' AND status = 'running'", req.TaskIDs).
		Count(&runningCount)
	
	if runningCount > 0 {
		response.FailWithMessage("不能删除正在运行的任务", c)
		return
	}
	
	// 批量删除
	result := global.GVA_DB.Where("id IN ? AND task_type = 'level4_generation'", req.TaskIDs).
		Delete(&nesmaModel.NesmaRequirementAnalysisTask{})
	
	if result.Error != nil {
		response.FailWithMessage("批量删除失败", c)
		return
	}
	
	response.OkWithData(gin.H{
		"message":      "批量删除成功",
		"deletedCount": result.RowsAffected,
	}, c)
}

// BatchCancelL4GenerationTasks 批量取消L4生成任务
// @Tags Level4Generator
// @Summary 批量取消L4生成任务
// @Description 批量取消正在运行的L4生成任务
// @Accept application/json
// @Produce application/json
// @Param data body gin.H true "任务ID列表"
// @Success 200 {object} response.Response{data=gin.H} "取消成功"
// @Router /nesma/generator/level4/tasks/batch-cancel [post]
func (a *Level4GeneratorApi) BatchCancelL4GenerationTasks(c *gin.Context) {
	var req struct {
		TaskIDs []uint `json:"task_ids" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	
	if len(req.TaskIDs) == 0 {
		response.FailWithMessage("任务ID列表不能为空", c)
		return
	}
	
	// 批量取消（只取消运行中和等待中的任务）
	result := global.GVA_DB.Model(&nesmaModel.NesmaRequirementAnalysisTask{}).
		Where("id IN ? AND task_type = 'level4_generation' AND status IN ?", req.TaskIDs, []string{"running", "pending"}).
		Update("status", "cancelled")
	
	if result.Error != nil {
		response.FailWithMessage("批量取消失败", c)
		return
	}
	
	response.OkWithData(gin.H{
		"message":       "批量取消成功",
		"cancelledCount": result.RowsAffected,
	}, c)
}