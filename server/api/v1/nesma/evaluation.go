package nesma

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type EvaluationApi struct{}

// GetUserOptions 获取用户选项列表
// @Tags NESMA评估
// @Summary 获取用户选项列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /nesma/user/options [get]
func (e *EvaluationApi) GetUserOptions(c *gin.Context) {
	evaluationService := &nesma.EvaluationService{}
	
	userOptions, err := evaluationService.GetUserOptions(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("获取用户选项失败", zap.Error(err))
		response.FailWithMessage("获取用户选项失败: "+err.Error(), c)
		return
	}
	
	response.OkWithData(userOptions, c)
}

// GetProjectCycles 获取项目周期选项列表
// @Tags NESMA评估
// @Summary 获取项目周期选项列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int true "项目ID"
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /nesma-evaluation/project-cycles [get]
func (e *EvaluationApi) GetProjectCycles(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	if projectIDStr == "" {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的项目ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	cycles, err := evaluationService.GetProjectCycles(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("获取项目周期失败", zap.Error(err))
		response.FailWithMessage("获取项目周期失败: "+err.Error(), c)
		return
	}

	response.OkWithData(cycles, c)
}

// GetCycleVersions 获取周期版本选项列表
// @Tags NESMA评估
// @Summary 获取周期版本选项列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId query int true "周期ID"
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /nesma-evaluation/cycle-versions [get]
func (e *EvaluationApi) GetCycleVersions(c *gin.Context) {
	cycleIDStr := c.Query("cycleId")
	if cycleIDStr == "" {
		response.FailWithMessage("周期ID不能为空", c)
		return
	}

	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的周期ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	versions, err := evaluationService.GetCycleVersions(uint(cycleID))
	if err != nil {
		global.GVA_LOG.Error("获取周期版本失败", zap.Error(err))
		response.FailWithMessage("获取周期版本失败: "+err.Error(), c)
		return
	}

	response.OkWithData(versions, c)
}

// @Tags NESMA评估
// @Summary 创建NESMA评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateEvaluationRequest true "创建评估请求"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationResponse} "创建成功"
// @Router /nesma-evaluation/create [post]
func (e *EvaluationApi) CreateEvaluation(c *gin.Context) {
	var req nesmaReq.CreateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证评估类型
	validTypes := []string{"initial", "detailed", "final"}
	isValidType := false
	for _, vt := range validTypes {
		if req.EvaluationType == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		response.FailWithMessage("无效的评估类型", c)
		return
	}

	// 获取当前用户ID作为评估人员
	userID := utils.GetUserID(c)
	req.EvaluatorID = userID

	// 设置默认NESMA规则版本
	if req.NesmaRules == "" {
		req.NesmaRules = "v2.2"
	}

	// 创建评估
	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.CreateEvaluation(req)
	if err != nil {
		global.GVA_LOG.Error("创建NESMA评估失败", zap.Error(err))
		response.FailWithMessage("创建评估失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 更新NESMA评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Param data body nesmaReq.UpdateEvaluationRequest true "更新评估请求"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationResponse} "更新成功"
// @Router /nesma-evaluation/{id} [put]
func (e *EvaluationApi) UpdateEvaluation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	var req nesmaReq.UpdateEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置评估ID
	req.ID = uint(id)

	// 验证评估类型（如果提供）
	if req.EvaluationType != "" {
		validTypes := []string{"initial", "detailed", "final"}
		isValidType := false
		for _, vt := range validTypes {
			if req.EvaluationType == vt {
				isValidType = true
				break
			}
		}
		if !isValidType {
			response.FailWithMessage("无效的评估类型", c)
			return
		}
	}

	// 更新评估
	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.UpdateEvaluation(req)
	if err != nil {
		global.GVA_LOG.Error("更新NESMA评估失败", zap.Error(err))
		response.FailWithMessage("更新评估失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 开始自动评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.StartEvaluationRequest true "开始评估请求"
// @Success 200 {object} response.Response{} "开始成功"
// @Router /nesma-evaluation/start [post]
func (e *EvaluationApi) StartEvaluation(c *gin.Context) {
	var req nesmaReq.StartEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err := evaluationService.StartAutoEvaluation(req.EvaluationID)
	if err != nil {
		global.GVA_LOG.Error("开始NESMA评估失败", zap.Error(err))
		response.FailWithMessage("开始评估失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估已开始，正在后台处理", c)
}

// @Tags NESMA评估
// @Summary 获取评估详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationResponse} "获取成功"
// @Router /nesma-evaluation/{id} [get]
func (e *EvaluationApi) GetEvaluation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetEvaluation(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取NESMA评估失败", zap.Error(err))
		response.FailWithMessage("获取评估失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 获取评估列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID"
// @Param status query string false "评估状态"
// @Param evaluationType query string false "评估类型"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationListResponse} "获取成功"
// @Router /nesma-evaluation/list [get]
func (e *EvaluationApi) GetEvaluationList(c *gin.Context) {
	var req nesmaReq.EvaluationSearchRequest

	// 解析查询参数
	if projectIDStr := c.Query("projectId"); projectIDStr != "" {
		if projectID, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			req.ProjectID = uint(projectID)
		}
	}

	req.Status = c.Query("status")
	req.EvaluationType = c.Query("evaluationType")

	// 解析分页参数
	req.Page = 1
	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	req.PageSize = 20
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			req.PageSize = pageSize
		}
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetEvaluationList(req)
	if err != nil {
		global.GVA_LOG.Error("获取NESMA评估列表失败", zap.Error(err))
		response.FailWithMessage("获取评估列表失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 获取功能点列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param evaluationId query int true "评估ID"
// @Param functionType query string false "功能类型"
// @Param complexityLevel query string false "复杂度级别"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(50)
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma-evaluation/function-points [get]
func (e *EvaluationApi) GetFunctionPoints(c *gin.Context) {
	evaluationIDStr := c.Query("evaluationId")
	global.GVA_LOG.Info("GetFunctionPoints API called", zap.String("evaluationId", evaluationIDStr))
	
	if evaluationIDStr == "" {
		response.FailWithMessage("评估ID不能为空", c)
		return
	}

	evaluationID, err := strconv.ParseUint(evaluationIDStr, 10, 32)
	if err != nil {
		global.GVA_LOG.Error("解析评估ID失败", zap.String("evaluationId", evaluationIDStr), zap.Error(err))
		response.FailWithMessage("无效的评估ID", c)
		return
	}
	
	global.GVA_LOG.Info("解析评估ID成功", zap.Uint64("evaluationID", evaluationID))

	functionType := c.Query("functionType")
	complexityLevel := c.Query("complexityLevel")

	// 解析分页参数
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 50
	if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 200 {
			pageSize = ps
		}
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetFunctionPoints(uint(evaluationID), functionType, complexityLevel, page, pageSize)
	if err != nil {
		global.GVA_LOG.Error("获取功能点列表失败", zap.Error(err))
		response.FailWithMessage("获取功能点列表失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 更新功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "功能点ID"
// @Param data body nesmaReq.FunctionPointRequest true "更新功能点请求"
// @Success 200 {object} response.Response{} "更新成功"
// @Router /nesma-evaluation/function-point/{id} [put]
func (e *EvaluationApi) UpdateFunctionPoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的功能点ID", c)
		return
	}

	var req nesmaReq.FunctionPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.UpdateFunctionPoint(uint(id), req)
	if err != nil {
		global.GVA_LOG.Error("更新功能点失败", zap.Error(err))
		response.FailWithMessage("更新功能点失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("功能点更新成功", c)
}

// @Tags NESMA评估
// @Summary 添加功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.FunctionPointRequest true "添加功能点请求"
// @Success 200 {object} response.Response{data=nesmaRes.FunctionPointResponse} "添加成功"
// @Router /nesma-evaluation/function-point [post]
func (e *EvaluationApi) CreateFunctionPoint(c *gin.Context) {
	var req nesmaReq.FunctionPointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证功能类型
	validTypes := []string{"ILF", "EIF", "EI", "EO", "EQ"}
	isValidType := false
	for _, vt := range validTypes {
		if req.FunctionType == vt {
			isValidType = true
			break
		}
	}
	if !isValidType {
		response.FailWithMessage("无效的功能类型", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.CreateFunctionPoint(req)
	if err != nil {
		global.GVA_LOG.Error("添加功能点失败", zap.Error(err))
		response.FailWithMessage("添加功能点失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 删除功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "功能点ID"
// @Success 200 {object} response.Response{} "删除成功"
// @Router /nesma-evaluation/function-point/{id} [delete]
func (e *EvaluationApi) DeleteFunctionPoint(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的功能点ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.DeleteFunctionPoint(uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除功能点失败", zap.Error(err))
		response.FailWithMessage("删除功能点失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("功能点删除成功", c)
}

// @Tags NESMA评估
// @Summary 获取评估统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationStatsResponse} "获取成功"
// @Router /nesma-evaluation/{id}/stats [get]
func (e *EvaluationApi) GetEvaluationStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetEvaluationStats(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取评估统计失败", zap.Error(err))
		response.FailWithMessage("获取评估统计失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 获取复杂度指标
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param evaluationId query int true "评估ID"
// @Success 200 {object} response.Response{data=[]nesmaRes.ComplexityMetricResponse} "获取成功"
// @Router /nesma-evaluation/complexity-metrics [get]
func (e *EvaluationApi) GetComplexityMetrics(c *gin.Context) {
	evaluationIDStr := c.Query("evaluationId")
	if evaluationIDStr == "" {
		response.FailWithMessage("评估ID不能为空", c)
		return
	}

	evaluationID, err := strconv.ParseUint(evaluationIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetComplexityMetrics(uint(evaluationID))
	if err != nil {
		global.GVA_LOG.Error("获取复杂度指标失败", zap.Error(err))
		response.FailWithMessage("获取复杂度指标失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 获取验证项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param evaluationId query int true "评估ID"
// @Param validationResult query string false "验证结果"
// @Success 200 {object} response.Response{data=[]nesmaRes.ValidationItemResponse} "获取成功"
// @Router /nesma-evaluation/validation-items [get]
func (e *EvaluationApi) GetValidationItems(c *gin.Context) {
	evaluationIDStr := c.Query("evaluationId")
	if evaluationIDStr == "" {
		response.FailWithMessage("评估ID不能为空", c)
		return
	}

	evaluationID, err := strconv.ParseUint(evaluationIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	validationResult := c.Query("validationResult")

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetValidationItems(uint(evaluationID), validationResult)
	if err != nil {
		global.GVA_LOG.Error("获取验证项目失败", zap.Error(err))
		response.FailWithMessage("获取验证项目失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 更新验证项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "验证项ID"
// @Param data body nesmaReq.ValidationItemRequest true "更新验证项请求"
// @Success 200 {object} response.Response{} "更新成功"
// @Router /nesma-evaluation/validation-item/{id} [put]
func (e *EvaluationApi) UpdateValidationItem(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的验证项ID", c)
		return
	}

	var req nesmaReq.ValidationItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.UpdateValidationItem(uint(id), req)
	if err != nil {
		global.GVA_LOG.Error("更新验证项失败", zap.Error(err))
		response.FailWithMessage("更新验证项失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("验证项更新成功", c)
}

// @Tags NESMA评估
// @Summary 评估审核
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ReviewEvaluationRequest true "评估审核请求"
// @Success 200 {object} response.Response{} "审核成功"
// @Router /nesma-evaluation/review [post]
func (e *EvaluationApi) ReviewEvaluation(c *gin.Context) {
	var req nesmaReq.ReviewEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证审核状态
	validStatuses := []string{"approved", "rejected", "needs_revision"}
	isValidStatus := false
	for _, vs := range validStatuses {
		if req.ReviewStatus == vs {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		response.FailWithMessage("无效的审核状态", c)
		return
	}

	// 获取当前用户ID作为审核人员
	userID := utils.GetUserID(c)
	req.ReviewerID = userID

	evaluationService := nesma.EvaluationService{}
	err := evaluationService.ReviewEvaluation(req)
	if err != nil {
		global.GVA_LOG.Error("评估审核失败", zap.Error(err))
		response.FailWithMessage("评估审核失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估审核完成", c)
}

// @Tags NESMA评估
// @Summary 导出评估报告
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Param format query string false "导出格式" default(pdf)
// @Success 200 {object} response.Response{data=map[string]interface{}} "导出成功"
// @Router /nesma-evaluation/{id}/export [get]
func (e *EvaluationApi) ExportEvaluationReport(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "pdf"
	}

	// 验证导出格式
	validFormats := []string{"pdf", "excel", "json"}
	isValidFormat := false
	for _, vf := range validFormats {
		if format == vf {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		response.FailWithMessage("不支持的导出格式", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.ExportEvaluationReport(uint(id), format)
	if err != nil {
		global.GVA_LOG.Error("导出评估报告失败", zap.Error(err))
		response.FailWithMessage("导出评估报告失败: "+err.Error(), c)
		return
	}

	response.OkWithData(map[string]interface{}{
		"downloadUrl": result["downloadUrl"],
		"fileName":    result["fileName"],
		"fileSize":    result["fileSize"],
		"exportTime":  time.Now(),
	}, c)
}

// @Tags NESMA评估
// @Summary 获取项目评估摘要
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId path int true "项目ID"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationSummaryResponse} "获取成功"
// @Router /nesma-evaluation/project/{projectId}/summary [get]
func (e *EvaluationApi) GetProjectEvaluationSummary(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的项目ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetProjectEvaluationSummary(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("获取项目评估摘要失败", zap.Error(err))
		response.FailWithMessage("获取项目评估摘要失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 重新计算评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Success 200 {object} response.Response{} "重新计算成功"
// @Router /nesma-evaluation/{id}/recalculate [post]
func (e *EvaluationApi) RecalculateEvaluation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.RecalculateEvaluation(uint(id))
	if err != nil {
		global.GVA_LOG.Error("重新计算评估失败", zap.Error(err))
		response.FailWithMessage("重新计算评估失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估重新计算已开始", c)
}

// @Tags NESMA评估
// @Summary 删除评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Success 200 {object} response.Response{} "删除成功"
// @Router /nesma-evaluation/{id} [delete]
func (e *EvaluationApi) DeleteEvaluation(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.DeleteEvaluation(uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除评估失败", zap.Error(err))
		response.FailWithMessage("删除评估失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估删除成功", c)
}

// ==================== 评估因子配置相关API ====================

// @Tags NESMA评估
// @Summary 获取评估因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma-evaluation/{id}/factors [get]
func (e *EvaluationApi) GetEvaluationFactors(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetEvaluationFactors(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取评估因子配置失败", zap.Error(err))
		response.FailWithMessage("获取评估因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 保存评估因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.EvaluationFactorsRequest true "评估因子配置"
// @Success 200 {object} response.Response{} "保存成功"
// @Router /nesma-evaluation/factors [post]
func (e *EvaluationApi) SaveEvaluationFactors(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证评估ID
	evaluationId, ok := req["evaluationId"].(float64)
	if !ok || evaluationId == 0 {
		response.FailWithMessage("评估ID不能为空", c)
		return
	}

	// 转换为服务层需要的结构
	factors := &nesma.EvaluationFactors{
		EvaluationID: uint(evaluationId),
	}

	// 解析数据功能配置
	if dataFunctions, ok := req["dataFunctions"].(map[string]interface{}); ok {
		// 初始化数据功能切片，确保它们是空的
		factors.DataFunctions.ILF = []nesma.FunctionWeight{}
		factors.DataFunctions.EIF = []nesma.FunctionWeight{}
		
		if ilf, ok := dataFunctions["ilf"].([]interface{}); ok {
			for _, item := range ilf {
				if weightItem, ok := item.(map[string]interface{}); ok {
					// 安全的类型断言，避免nil值导致panic
					complexity, _ := weightItem["complexity"].(string)
					weightVal, _ := weightItem["weight"].(float64)
					description, _ := weightItem["description"].(string)
					
					weight := nesma.FunctionWeight{
						Complexity:  complexity,
						Weight:      int(weightVal),
						Description: description,
					}
					factors.DataFunctions.ILF = append(factors.DataFunctions.ILF, weight)
				}
			}
		}
		if eif, ok := dataFunctions["eif"].([]interface{}); ok {
			for _, item := range eif {
				if weightItem, ok := item.(map[string]interface{}); ok {
					// 安全的类型断言，避免nil值导致panic
					complexity, _ := weightItem["complexity"].(string)
					weightVal, _ := weightItem["weight"].(float64)
					description, _ := weightItem["description"].(string)
					
					weight := nesma.FunctionWeight{
						Complexity:  complexity,
						Weight:      int(weightVal),
						Description: description,
					}
					factors.DataFunctions.EIF = append(factors.DataFunctions.EIF, weight)
				}
			}
		}
	}

	// 解析事务功能配置
	if transactionFunctions, ok := req["transactionFunctions"].(map[string]interface{}); ok {
		// 初始化事务功能切片，确保它们是空的
		factors.TransactionFunctions.EI = []nesma.FunctionWeight{}
		factors.TransactionFunctions.EO = []nesma.FunctionWeight{}
		factors.TransactionFunctions.EQ = []nesma.FunctionWeight{}
		
		for funcType, funcData := range transactionFunctions {
			if funcArray, ok := funcData.([]interface{}); ok {
				for _, item := range funcArray {
					if weightItem, ok := item.(map[string]interface{}); ok {
						// 安全的类型断言，避免nil值导致panic
						complexity, _ := weightItem["complexity"].(string)
						weightVal, _ := weightItem["weight"].(float64)
						description, _ := weightItem["description"].(string)
						
						weight := nesma.FunctionWeight{
							Complexity:  complexity,
							Weight:      int(weightVal),
							Description: description,
						}
						switch funcType {
						case "ei":
							factors.TransactionFunctions.EI = append(factors.TransactionFunctions.EI, weight)
						case "eo":
							factors.TransactionFunctions.EO = append(factors.TransactionFunctions.EO, weight)
						case "eq":
							factors.TransactionFunctions.EQ = append(factors.TransactionFunctions.EQ, weight)
						}
					}
				}
			}
		}
	}

	// 解析调整因子
	if adjustmentFactors, ok := req["adjustmentFactors"].([]interface{}); ok {
		// 初始化调整因子切片，确保它是空的
		factors.AdjustmentFactors = []nesma.AdjustmentFactor{}
		
		for _, item := range adjustmentFactors {
			if adjItem, ok := item.(map[string]interface{}); ok {
				// 安全的类型断言，避免nil值导致panic
				name, _ := adjItem["name"].(string)
				value, _ := adjItem["value"].(float64)
				description, _ := adjItem["description"].(string)
				
				factor := nesma.AdjustmentFactor{
					Name:        name,
					Value:       int(value),
					Description: description,
				}
				factors.AdjustmentFactors = append(factors.AdjustmentFactors, factor)
			}
		}
	}

	evaluationService := nesma.EvaluationService{}
	err := evaluationService.SaveEvaluationFactors(factors)
	if err != nil {
		global.GVA_LOG.Error("保存评估因子配置失败", zap.Error(err))
		response.FailWithMessage("保存评估因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估因子配置保存成功", c)
}

// @Tags NESMA评估
// @Summary 更新评估因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Param data body nesmaReq.EvaluationFactorsRequest true "评估因子配置"
// @Success 200 {object} response.Response{} "更新成功"
// @Router /nesma-evaluation/{id}/factors [put]
func (e *EvaluationApi) UpdateEvaluationFactors(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 转换为服务层需要的结构
	factors := &nesma.EvaluationFactors{
		EvaluationID: uint(id),
	}

	// 调试日志：记录UpdateEvaluationFactors开始时的状态
	global.GVA_LOG.Info("UpdateEvaluationFactors开始", 
		zap.Uint("evaluationId", uint(id)),
		zap.Int("初始EO切片长度", len(factors.TransactionFunctions.EO)))

	// 解析数据功能配置
	if dataFunctions, ok := req["dataFunctions"].(map[string]interface{}); ok {
		// 初始化数据功能切片，确保它们是空的
		factors.DataFunctions.ILF = []nesma.FunctionWeight{}
		factors.DataFunctions.EIF = []nesma.FunctionWeight{}
		
		if ilf, ok := dataFunctions["ilf"].([]interface{}); ok {
			for _, item := range ilf {
				if weightItem, ok := item.(map[string]interface{}); ok {
					// 安全的类型断言
					complexity, _ := weightItem["complexity"].(string)
					weight, _ := weightItem["weight"].(float64)
					description, _ := weightItem["description"].(string)
					
					weightStruct := nesma.FunctionWeight{
						Complexity:  complexity,
						Weight:      int(weight),
						Description: description,
					}
					factors.DataFunctions.ILF = append(factors.DataFunctions.ILF, weightStruct)
				}
			}
		}
		if eif, ok := dataFunctions["eif"].([]interface{}); ok {
			for _, item := range eif {
				if weightItem, ok := item.(map[string]interface{}); ok {
					// 安全的类型断言
					complexity, _ := weightItem["complexity"].(string)
					weight, _ := weightItem["weight"].(float64)
					description, _ := weightItem["description"].(string)
					
					weightStruct := nesma.FunctionWeight{
						Complexity:  complexity,
						Weight:      int(weight),
						Description: description,
					}
					factors.DataFunctions.EIF = append(factors.DataFunctions.EIF, weightStruct)
				}
			}
		}
	}

	// 解析事务功能配置
	if transactionFunctions, ok := req["transactionFunctions"].(map[string]interface{}); ok {
		// 初始化事务功能切片，确保它们是空的
		factors.TransactionFunctions.EI = []nesma.FunctionWeight{}
		factors.TransactionFunctions.EO = []nesma.FunctionWeight{}
		factors.TransactionFunctions.EQ = []nesma.FunctionWeight{}
		
		for funcType, funcData := range transactionFunctions {
			if funcArray, ok := funcData.([]interface{}); ok {
				for _, item := range funcArray {
					if weightItem, ok := item.(map[string]interface{}); ok {
						// 安全的类型断言
						complexity, _ := weightItem["complexity"].(string)
						weight, _ := weightItem["weight"].(float64)
						description, _ := weightItem["description"].(string)
						
						weightStruct := nesma.FunctionWeight{
							Complexity:  complexity,
							Weight:      int(weight),
							Description: description,
						}
						switch funcType {
						case "ei":
							factors.TransactionFunctions.EI = append(factors.TransactionFunctions.EI, weightStruct)
						case "eo":
							factors.TransactionFunctions.EO = append(factors.TransactionFunctions.EO, weightStruct)
						case "eq":
							factors.TransactionFunctions.EQ = append(factors.TransactionFunctions.EQ, weightStruct)
						}
					}
				}
			}
		}
	}

	// 解析调整因子
	if adjustmentFactors, ok := req["adjustmentFactors"].([]interface{}); ok {
		// 初始化调整因子切片，确保它是空的
		factors.AdjustmentFactors = []nesma.AdjustmentFactor{}
		
		for _, item := range adjustmentFactors {
			if adjItem, ok := item.(map[string]interface{}); ok {
				// 安全的类型断言，避免nil值导致panic
				name, _ := adjItem["name"].(string)
				value, _ := adjItem["value"].(float64)
				description, _ := adjItem["description"].(string)
				
				factor := nesma.AdjustmentFactor{
					Name:        name,
					Value:       int(value),
					Description: description,
				}
				factors.AdjustmentFactors = append(factors.AdjustmentFactors, factor)
			}
		}
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.UpdateEvaluationFactors(uint(id), factors)
	if err != nil {
		global.GVA_LOG.Error("更新评估因子配置失败", zap.Error(err))
		response.FailWithMessage("更新评估因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估因子配置更新成功", c)
}

// @Tags NESMA评估
// @Summary 获取默认NESMA因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param version query string false "NESMA规则版本" default(v2.2)
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /nesma-evaluation/default-factors [get]
func (e *EvaluationApi) GetDefaultNESMAFactors(c *gin.Context) {
	version := c.DefaultQuery("version", "v2.2")

	// 验证版本
	validVersions := []string{"v2.0", "v2.1", "v2.2"}
	isValidVersion := false
	for _, vv := range validVersions {
		if version == vv {
			isValidVersion = true
			break
		}
	}
	if !isValidVersion {
		response.FailWithMessage("不支持的NESMA规则版本", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetDefaultNESMAFactors(version)
	if err != nil {
		global.GVA_LOG.Error("获取默认NESMA因子配置失败", zap.Error(err))
		response.FailWithMessage("获取默认NESMA因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 验证评估因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.EvaluationFactorsRequest true "评估因子配置"
// @Success 200 {object} response.Response{data=map[string]interface{}} "验证成功"
// @Router /nesma-evaluation/validate-factors [post]
func (e *EvaluationApi) ValidateEvaluationFactors(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 转换为服务层需要的结构
	factors := &nesma.EvaluationFactors{}
	// 这里可以添加类似的解析逻辑，但为了简化，我们直接调用验证
	
	evaluationService := nesma.EvaluationService{}
	err := evaluationService.ValidateEvaluationFactors(factors)
	if err != nil {
		global.GVA_LOG.Error("验证评估因子配置失败", zap.Error(err))
		response.FailWithMessage("验证评估因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估因子配置验证通过", c)
}

// @Tags NESMA评估
// @Summary 计算调整因子
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.AdjustmentFactorsRequest true "调整因子数据"
// @Success 200 {object} response.Response{data=map[string]interface{}} "计算成功"
// @Router /nesma-evaluation/calculate-adjustment [post]
func (e *EvaluationApi) CalculateAdjustmentFactor(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 解析调整因子
	var adjustmentFactors []nesma.AdjustmentFactor
	if factors, ok := req["adjustmentFactors"].([]interface{}); ok {
		for _, item := range factors {
			if adjItem, ok := item.(map[string]interface{}); ok {
				// 安全的类型断言，避免nil值导致panic
				name, _ := adjItem["name"].(string)
				value, _ := adjItem["value"].(float64)
				description, _ := adjItem["description"].(string)
				
				factor := nesma.AdjustmentFactor{
					Name:        name,
					Value:       int(value),
					Description: description,
				}
				adjustmentFactors = append(adjustmentFactors, factor)
			}
		}
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.CalculateAdjustmentFactor(adjustmentFactors)
	if err != nil {
		global.GVA_LOG.Error("计算调整因子失败", zap.Error(err))
		response.FailWithMessage("计算调整因子失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 预览功能点计算
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.FunctionPointPreviewRequest true "功能点预览数据"
// @Success 200 {object} response.Response{data=map[string]interface{}} "预览成功"
// @Router /nesma-evaluation/preview-calculation [post]
func (e *EvaluationApi) PreviewFunctionPointCalculation(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.PreviewFunctionPointCalculation(req)
	if err != nil {
		global.GVA_LOG.Error("预览功能点计算失败", zap.Error(err))
		response.FailWithMessage("预览功能点计算失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 应用评估因子到项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Param data body nesmaReq.ApplyFactorsRequest true "应用因子请求"
// @Success 200 {object} response.Response{} "应用成功"
// @Router /nesma-evaluation/{id}/apply-factors [post]
func (e *EvaluationApi) ApplyFactorsToProject(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	err = evaluationService.ApplyFactorsToProject(uint(id), req)
	if err != nil {
		global.GVA_LOG.Error("应用评估因子到项目失败", zap.Error(err))
		response.FailWithMessage("应用评估因子到项目失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估因子已成功应用到项目", c)
}

// @Tags NESMA评估
// @Summary 获取评估因子历史记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /nesma-evaluation/{id}/factors-history [get]
func (e *EvaluationApi) GetFactorsHistory(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.GetFactorsHistory(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取评估因子历史记录失败", zap.Error(err))
		response.FailWithMessage("获取评估因子历史记录失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 导出评估因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "评估ID"
// @Param format query string false "导出格式" default(json)
// @Success 200 {object} response.Response{data=map[string]interface{}} "导出成功"
// @Router /nesma-evaluation/{id}/export-factors [get]
func (e *EvaluationApi) ExportEvaluationFactors(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	format := c.DefaultQuery("format", "json")

	// 验证导出格式
	validFormats := []string{"json", "excel", "csv"}
	isValidFormat := false
	for _, vf := range validFormats {
		if format == vf {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		response.FailWithMessage("不支持的导出格式", c)
		return
	}

	evaluationService := nesma.EvaluationService{}
	result, err := evaluationService.ExportEvaluationFactors(uint(id), format)
	if err != nil {
		global.GVA_LOG.Error("导出评估因子配置失败", zap.Error(err))
		response.FailWithMessage("导出评估因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// @Tags NESMA评估
// @Summary 导入评估因子配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param file formData file true "配置文件"
// @Param evaluationId formData int true "评估ID"
// @Success 200 {object} response.Response{} "导入成功"
// @Router /nesma-evaluation/import-factors [post]
func (e *EvaluationApi) ImportEvaluationFactors(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("文件上传失败: "+err.Error(), c)
		return
	}

	// 获取评估ID
	evaluationIdStr := c.PostForm("evaluationId")
	evaluationId, err := strconv.ParseUint(evaluationIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	// 打开文件
	src, err := file.Open()
	if err != nil {
		response.FailWithMessage("打开文件失败: "+err.Error(), c)
		return
	}
	defer src.Close()

	// 读取文件内容
	fileContent := make([]byte, file.Size)
	_, err = src.Read(fileContent)
	if err != nil {
		response.FailWithMessage("读取文件失败: "+err.Error(), c)
		return
	}

	// 验证文件格式（JSON）
	if !isValidJSON(fileContent) {
		response.FailWithMessage("文件格式错误，请上传有效的JSON文件", c)
		return
	}

	// 调用服务层导入
	evaluationService := nesma.EvaluationService{}
	err = evaluationService.ImportEvaluationFactors(uint(evaluationId), fileContent)
	if err != nil {
		global.GVA_LOG.Error("导入评估因子配置失败", zap.Error(err))
		response.FailWithMessage("导入评估因子配置失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估因子配置导入成功", c)
}

// isValidJSON 验证是否为有效的JSON格式
func isValidJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}
