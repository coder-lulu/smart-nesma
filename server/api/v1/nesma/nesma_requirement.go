package nesma

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NesmaRequirementApi struct{}

// CreateNesmaRequirement 创建需求
// @Tags NesmaRequirement
// @Summary 创建需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateNesmaRequirementRequest true "创建需求请求"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirement} "创建成功"
// @Router /nesma/requirement [post]
func (a *NesmaRequirementApi) CreateNesmaRequirement(c *gin.Context) {
	var req nesmaReq.CreateNesmaRequirementRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	requirement, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.CreateNesmaRequirement(&req)
	if err != nil {
		global.GVA_LOG.Error("创建需求失败!", zap.Error(err))
		response.FailWithMessage("创建需求失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(requirement, "创建需求成功", c)
}

// UpdateNesmaRequirement 更新需求
// @Tags NesmaRequirement
// @Summary 更新需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.UpdateNesmaRequirementRequest true "更新需求请求"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirement} "更新成功"
// @Router /nesma/requirement [put]
func (a *NesmaRequirementApi) UpdateNesmaRequirement(c *gin.Context) {
	var req nesmaReq.UpdateNesmaRequirementRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	requirement, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.UpdateNesmaRequirement(&req)
	if err != nil {
		global.GVA_LOG.Error("更新需求失败!", zap.Error(err))
		response.FailWithMessage("更新需求失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(requirement, "更新需求成功", c)
}

// GetRequirementChildren 获取需求的子功能点信息
// @Tags NesmaRequirement
// @Summary 获取需求的子功能点信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "需求ID"
// @Success 200 {object} response.Response{data=nesmaRes.RequirementChildrenResponse} "获取成功"
// @Router /nesma/requirement/{id}/children [get]
func (a *NesmaRequirementApi) GetRequirementChildren(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	children, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetRequirementChildren(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取子功能点失败!", zap.Error(err))
		response.FailWithMessage("获取子功能点失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(children, "获取子功能点成功", c)
}

// DeleteNesmaRequirement 删除需求
// @Tags NesmaRequirement
// @Summary 删除需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "需求ID"
// @Param force query bool false "是否强制删除（包括子功能点）"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/requirement/{id} [delete]
func (a *NesmaRequirementApi) DeleteNesmaRequirement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	// 获取force参数，决定是否级联删除子功能点
	forceDelete := c.Query("force") == "true"

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.DeleteNesmaRequirement(uint(id), forceDelete)
	if err != nil {
		global.GVA_LOG.Error("删除需求失败!", zap.Error(err))
		response.FailWithMessage("删除需求失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除需求成功", c)
}

// DeleteRequirementsByCondition 根据条件删除需求
// @Tags NesmaRequirement
// @Summary 根据条件删除需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.DeleteRequirementsByConditionRequest true "删除条件"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/requirement/delete-by-condition [post]
func (a *NesmaRequirementApi) DeleteRequirementsByCondition(c *gin.Context) {
	var req nesmaReq.DeleteRequirementsByConditionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.DeleteRequirementsByCondition(&req)
	if err != nil {
		global.GVA_LOG.Error("根据条件删除需求失败!", zap.Error(err))
		response.FailWithMessage("根据条件删除需求失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("根据条件删除需求成功", c)
}

// GetNesmaRequirement 获取需求详情
// @Tags NesmaRequirement
// @Summary 获取需求详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "需求ID"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaRequirementResponse} "获取成功"
// @Router /nesma/requirement/{id} [get]
func (a *NesmaRequirementApi) GetNesmaRequirement(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	requirement, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetNesmaRequirement(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取需求详情失败!", zap.Error(err))
		response.FailWithMessage("获取需求详情失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(requirement, "获取需求详情成功", c)
}

// GetNesmaRequirementList 获取需求列表
// @Tags NesmaRequirement
// @Summary 获取需求列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query nesmaReq.NesmaRequirementSearch true "分页查询参数"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaRequirementListResponse} "获取成功"
// @Router /nesma/requirement/list [get]
func (a *NesmaRequirementApi) GetNesmaRequirementList(c *gin.Context) {
	var req nesmaReq.NesmaRequirementSearch
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetNesmaRequirementList(&req)
	if err != nil {
		global.GVA_LOG.Error("获取需求列表失败!", zap.Error(err))
		response.FailWithMessage("获取需求列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(list, "获取需求列表成功", c)
}

// GetNesmaRequirementTree 获取需求树形结构
// @Tags NesmaRequirement
// @Summary 获取需求树形结构
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int true "项目ID"
// @Param level query int false "需求层级"
// @Param status query string false "状态"
// @Param keyword query string false "关键词"
// @Success 200 {object} response.Response{data=[]nesmaRes.NesmaRequirementTreeResponse} "获取成功"
// @Router /nesma/requirement/tree [get]
func (a *NesmaRequirementApi) GetNesmaRequirementTree(c *gin.Context) {
	var req nesmaReq.NesmaRequirementTreeSearch
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 添加调试日志
	global.GVA_LOG.Info("GetNesmaRequirementTree 请求参数",
		zap.Any("projectID", req.ProjectID),
		zap.Any("level", req.Level),
		zap.String("status", req.Status),
		zap.String("keyword", req.Keyword))

	if req.ProjectID == nil {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}

	tree, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetNesmaRequirementTreeWithFilter(&req)
	if err != nil {
		global.GVA_LOG.Error("获取需求树失败!", zap.Error(err))
		response.FailWithMessage("获取需求树失败: "+err.Error(), c)
		return
	}

	global.GVA_LOG.Info("GetNesmaRequirementTree 返回结果数量", zap.Int("count", len(tree)))
	response.OkWithDetailed(tree, "获取需求树成功", c)
}

// GetNesmaRequirementStats 获取需求统计
// @Tags NesmaRequirement
// @Summary 获取需求统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID，不传则获取全部项目统计"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaRequirementStatsResponse} "获取成功"
// @Router /nesma/requirement/stats [get]
func (a *NesmaRequirementApi) GetNesmaRequirementStats(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	global.GVA_LOG.Info("GetNesmaRequirementStats - projectIDStr: " + projectIDStr)

	var projectID *uint
	if projectIDStr != "" {
		projectIDUint, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			global.GVA_LOG.Error("项目ID格式错误", zap.String("projectIDStr", projectIDStr), zap.Error(err))
			response.FailWithMessage("项目ID格式错误: "+err.Error(), c)
			return
		}
		temp := uint(projectIDUint)
		projectID = &temp
	}

	stats, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetNesmaRequirementStats(projectID)
	if err != nil {
		global.GVA_LOG.Error("获取需求统计失败!", zap.Error(err))
		response.FailWithMessage("获取需求统计失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(stats, "获取需求统计成功", c)
}

// GetParentRequirementOptions 获取父需求选项
// @Tags NesmaRequirement
// @Summary 获取父需求选项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int true "项目ID"
// @Param level query int true "当前层级"
// @Success 200 {object} response.Response{data=nesmaRes.ParentRequirementOptionsResponse} "获取成功"
// @Router /nesma/requirement/parent-options [get]
func (a *NesmaRequirementApi) GetParentRequirementOptions(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	levelStr := c.Query("level")
	global.GVA_LOG.Info("GetParentRequirementOptions - projectIDStr: " + projectIDStr + ", levelStr: " + levelStr)

	if projectIDStr == "" {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		global.GVA_LOG.Error("项目ID格式错误", zap.String("projectIDStr", projectIDStr), zap.Error(err))
		response.FailWithMessage("项目ID格式错误: "+err.Error(), c)
		return
	}

	if levelStr == "" {
		response.FailWithMessage("层级不能为空", c)
		return
	}

	level, err := strconv.Atoi(levelStr)
	if err != nil {
		global.GVA_LOG.Error("层级格式错误", zap.String("levelStr", levelStr), zap.Error(err))
		response.FailWithMessage("层级格式错误: "+err.Error(), c)
		return
	}

	options, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetParentRequirementOptions(uint(projectID), level)
	if err != nil {
		global.GVA_LOG.Error("获取父需求选项失败!", zap.Error(err))
		response.FailWithMessage("获取父需求选项失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(options, "获取父需求选项成功", c)
}

// BatchDeleteNesmaRequirements 批量删除需求
// @Tags NesmaRequirement
// @Summary 批量删除需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaRequirementIdsRequest true "批量删除请求"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/requirement/batch-delete [post]
func (a *NesmaRequirementApi) BatchDeleteNesmaRequirements(c *gin.Context) {
	var req nesmaReq.NesmaRequirementIdsRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.BatchDeleteNesmaRequirements(&req)
	if err != nil {
		global.GVA_LOG.Error("批量删除需求失败!", zap.Error(err))
		response.FailWithMessage("批量删除需求失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("批量删除需求成功", c)
}

// ImportFromExcel 从Excel导入需求
// @Tags NesmaRequirement
// @Summary 从Excel导入需求
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Param file formData file true "Excel文件"
// @Param projectId formData int true "项目ID"
// @Param cycleId formData int true "周期ID"
// @Param versionName formData string false "版本名称"
// @Param importSource formData string false "导入来源"
// @Success 200 {object} response.Response{data=nesmaRes.ImportResultResponse} "导入成功"
// @Router /nesma/requirement/import-excel [post]
func (a *NesmaRequirementApi) ImportFromExcel(c *gin.Context) {
	projectIDStr := c.PostForm("projectId")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	cycleIDStr := c.PostForm("cycleId")
	if cycleIDStr == "" {
		response.FailWithMessage("周期ID不能为空", c)
		return
	}
	
	cycleID, err := strconv.ParseUint(cycleIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	versionName := c.PostForm("versionName")


	versionAction := c.PostForm("versionAction")
	if versionAction == "" {
		versionAction = "new" // 默认创建新版本
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("文件上传失败", c)
		return
	}

	// 保存上传的文件
	dst := "./uploads/" + file.Filename
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.FailWithMessage("文件保存失败", c)
		return
	}

	// 获取导入来源
	importSource := c.PostForm("importSource")
	if importSource == "" {
		importSource = "Excel导入"
	}

	// 创建导入请求 - 使用正确的字段结构
	req := &nesmaReq.ImportNesmaRequirementRequest{
		ProjectID:    uint(projectID),
		CycleID:      uint(cycleID),
		VersionName:  versionName,
		ImportBatch:  fmt.Sprintf("batch_%d", time.Now().Unix()),
		ImportSource: importSource,
		VersionAction: versionAction,
	}

	result, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.ImportFromExcel(dst, req)
	if err != nil {
		global.GVA_LOG.Error("导入需求失败!", zap.Error(err))
		response.FailWithMessage("导入需求失败: "+err.Error(), c)
		return
	}

	// 删除临时文件
	os.Remove(dst)

	response.OkWithDetailed(result, "导入需求完成", c)
}

// BatchUpdateOrder 批量更新排序
// @Tags NesmaRequirement
// @Summary 批量更新排序
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchUpdateOrderRequest true "批量更新排序请求"
// @Success 200 {object} response.Response "更新成功"
// @Router /nesma/requirement/batch-update-order [post]
func (a *NesmaRequirementApi) BatchUpdateOrder(c *gin.Context) {
	var req nesmaReq.BatchUpdateOrderRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.BatchUpdateOrder(&req)
	if err != nil {
		global.GVA_LOG.Error("批量更新排序失败!", zap.Error(err))
		response.FailWithMessage("批量更新排序失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("批量更新排序成功", c)
}

// MoveRequirement 移动需求
// @Tags NesmaRequirement
// @Summary 移动需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.MoveRequirementRequest true "移动需求请求"
// @Success 200 {object} response.Response "移动成功"
// @Router /nesma/requirement/move [post]
func (a *NesmaRequirementApi) MoveRequirement(c *gin.Context) {
	var req nesmaReq.MoveRequirementRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.MoveRequirement(&req)
	if err != nil {
		global.GVA_LOG.Error("移动需求失败!", zap.Error(err))
		response.FailWithMessage("移动需求失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("移动需求成功", c)
}

// GetProjectMaxVersion 获取项目最大版本号
// @Tags NesmaRequirement
// @Summary 获取项目最大版本号
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId path int true "项目ID"
// @Success 200 {object} response.Response{data=map[string]int} "获取成功"
// @Router /nesma/requirement/max-version/{projectId} [get]
func (a *NesmaRequirementApi) GetProjectMaxVersion(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	maxVersion, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetProjectMaxVersion(uint(projectId))
	if err != nil {
		global.GVA_LOG.Error("获取项目最大版本号失败!", zap.Error(err))
		response.FailWithMessage("获取项目最大版本号失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(map[string]float64{"maxVersion": maxVersion}, "获取项目最大版本号成功", c)
}

// GetRequirementAIAnalysis 获取需求AI分析结果
// @Tags NesmaRequirement
// @Summary 获取需求AI分析结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "需求ID"
// @Success 200 {object} response.Response{data=nesmaRes.RequirementAIAnalysisResponse} "获取成功"
// @Router /nesma/requirement/{id}/ai-analysis [get]
func (a *NesmaRequirementApi) GetRequirementAIAnalysis(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	analysis, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetRequirementAIAnalysis(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取需求AI分析失败!", zap.Error(err))
		response.FailWithMessage("获取需求AI分析失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(analysis, "获取需求AI分析成功", c)
}

// GetProjectAIAnalysisStats 获取项目AI分析统计
// @Tags NesmaRequirement
// @Summary 获取项目AI分析统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int true "项目ID"
// @Param cycleId query int false "周期ID"
// @Success 200 {object} response.Response{data=nesmaRes.ProjectAIAnalysisStats} "获取成功"
// @Router /nesma/requirement/ai-analysis-stats [get]
func (a *NesmaRequirementApi) GetProjectAIAnalysisStats(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	if projectIDStr == "" {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	var cycleID *uint
	cycleIDStr := c.Query("cycleId")
	if cycleIDStr != "" {
		cycleIDUint, err := strconv.ParseUint(cycleIDStr, 10, 32)
		if err != nil {
			response.FailWithMessage("周期ID格式错误", c)
			return
		}
		temp := uint(cycleIDUint)
		cycleID = &temp
	}

	stats, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetProjectAIAnalysisStats(uint(projectID), cycleID)
	if err != nil {
		global.GVA_LOG.Error("获取项目AI分析统计失败!", zap.Error(err))
		response.FailWithMessage("获取项目AI分析统计失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(stats, "获取项目AI分析统计成功", c)
}

// CompareRequirementVersions 对比需求版本
// @Tags NesmaRequirement
// @Summary 对比需求版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CompareRequirementVersionsRequest true "对比需求版本请求"
// @Success 200 {object} response.Response{data=nesmaRes.RequirementVersionCompareResponse} "对比成功"
// @Router /nesma/requirement/compare-versions [post]
func (a *NesmaRequirementApi) CompareRequirementVersions(c *gin.Context) {
	var req nesmaReq.CompareRequirementVersionsRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	comparison, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.CompareRequirementVersions(&req)
	if err != nil {
		global.GVA_LOG.Error("对比需求版本失败!", zap.Error(err))
		response.FailWithMessage("对比需求版本失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(comparison, "对比需求版本成功", c)
}

// UpdateRequirementAIStatus 更新需求AI分析状态
// @Tags NesmaRequirement
// @Summary 更新需求AI分析状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.UpdateRequirementAIStatusRequest true "更新AI状态请求"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /nesma/requirement/ai-status [put]
func (a *NesmaRequirementApi) UpdateRequirementAIStatus(c *gin.Context) {
	var req nesmaReq.UpdateRequirementAIStatusRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.UpdateRequirementAIStatus(&req)
	if err != nil {
		global.GVA_LOG.Error("更新需求AI状态失败!", zap.Error(err))
		response.FailWithMessage("更新需求AI状态失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("更新需求AI状态成功", c)
}

// GetAIOptimizedRequirements 获取AI优化的需求列表
// @Tags NesmaRequirement
// @Summary 获取AI优化的需求列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int true "项目ID"
// @Param cycleId query int false "周期ID"
// @Param versionId query int false "版本ID"
// @Param status query string false "AI分析状态"
// @Success 200 {object} response.Response{data=nesmaRes.AIOptimizedRequirementsResponse} "获取成功"
// @Router /nesma/requirement/ai-optimized [get]
func (a *NesmaRequirementApi) GetAIOptimizedRequirements(c *gin.Context) {
	var req nesmaReq.GetAIOptimizedRequirementsRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	if req.ProjectID == nil {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}

	requirements, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementService.GetAIOptimizedRequirements(&req)
	if err != nil {
		global.GVA_LOG.Error("获取AI优化需求列表失败!", zap.Error(err))
		response.FailWithMessage("获取AI优化需求列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(requirements, "获取AI优化需求列表成功", c)
}

// AnalyzeRequirement 分析单个需求
// @Tags NesmaRequirement
// @Summary 分析单个需求
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "分析需求请求"
// @Success 200 {object} response.Response{data=object} "分析成功"
// @Router /nesma/requirement/ai-analysis [post]
func (a *NesmaRequirementApi) AnalyzeRequirement(c *gin.Context) {
	var req struct {
		RequirementId uint `json:"requirementId" binding:"required"`
		ProjectId     uint `json:"projectId" binding:"required"`
		CycleId       uint `json:"cycleId" binding:"required"`
	}
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 构建单需求分析请求
	analysisReq := &nesmaService.SingleAnalysisRequest{
		RequirementID: req.RequirementId,
		ProjectID:     req.ProjectId,
		CycleID:       req.CycleId,
	}

	// 调用单需求分析服务（异步）
	task, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.AnalyzeSingleRequirementAsync(analysisReq)
	if err != nil {
		global.GVA_LOG.Error("启动单需求分析失败!", zap.Error(err))
		response.FailWithMessage("启动需求分析失败: "+err.Error(), c)
		return
	}

	// 返回任务信息
	response.OkWithData(gin.H{
		"taskId":         task.ID,
		"requirementId":  task.RequirementID,
		"status":         task.Status,
		"progress":       task.Progress,
		"message":        "需求分析任务已启动",
	}, c)
}

// ApplyRequirementAnalysis 应用需求分析结果
// @Tags NesmaRequirement
// @Summary 应用需求分析结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "应用分析结果请求"
// @Success 200 {object} response.Response{msg=string} "应用成功"
// @Router /nesma/requirement/apply-analysis [post]
func (a *NesmaRequirementApi) ApplyRequirementAnalysis(c *gin.Context) {
	var req struct {
		RequirementId uint                             `json:"requirementId" binding:"required"`
		OptimizedInfo *nesmaService.AnalysisRequirementInfo `json:"optimizedInfo" binding:"required"`
		AnalysisNote  string                          `json:"analysisNote"`
		UpdateStatus  bool                            `json:"updateStatus"`  // 新增：是否更新状态为已完成
	}
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 获取原需求信息
	var requirement nesma.NesmaRequirement
	if err := global.GVA_DB.First(&requirement, req.RequirementId).Error; err != nil {
		response.FailWithMessage("需求不存在", c)
		return
	}

	// 备份原始信息
	originalTitle := requirement.Title

	// 应用优化信息
	requirement.Title = req.OptimizedInfo.Title
	requirement.Description = req.OptimizedInfo.Description
	requirement.FunctionType = req.OptimizedInfo.FunctionType
	requirement.Complexity = req.OptimizedInfo.Complexity
	requirement.BusinessValue = req.OptimizedInfo.BusinessValue
	
	if req.OptimizedInfo.RecommendedAFP != nil {
		requirement.AFP = *req.OptimizedInfo.RecommendedAFP
	}
	if req.OptimizedInfo.RecommendedUFP != nil {
		requirement.UFP = *req.OptimizedInfo.RecommendedUFP
	}

	// 更新分析相关字段
	requirement.AIAnalysisStatus = "applied"
	requirement.AIGeneratedTitle = req.OptimizedInfo.Title
	requirement.AIDescription = req.OptimizedInfo.Description
	analysisTime := time.Now()
	requirement.AIAnalysisTime = &analysisTime
	requirement.Notes = req.AnalysisNote

	// 如果需要更新状态为已完成
	if req.UpdateStatus {
		requirement.Status = "completed"
	}

	// 保存更新
	if err := global.GVA_DB.Save(&requirement).Error; err != nil {
		global.GVA_LOG.Error("应用分析结果失败!", zap.Error(err))
		response.FailWithMessage("应用分析结果失败: "+err.Error(), c)
		return
	}

	// 记录操作日志
	global.GVA_LOG.Info("应用需求分析结果", 
		zap.Uint("requirementId", req.RequirementId),
		zap.String("originalTitle", originalTitle),
		zap.String("newTitle", requirement.Title))

	response.OkWithMessage("需求分析结果已成功应用", c)
}

// GetAnalysisProgress 获取分析进度
// @Tags NesmaRequirement
// @Summary 获取分析进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Router /nesma/requirement/analysis-progress/{taskId} [get]
func (a *NesmaRequirementApi) GetAnalysisProgress(c *gin.Context) {
	taskIdStr := c.Param("taskId")
	taskId, err := strconv.ParseUint(taskIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	// 先尝试获取单需求分析任务
	singleTask, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.GetSingleAnalysisTask(uint(taskId))
	if err == nil {
		// 返回单需求分析任务信息（包含动画数据）
		responseData := gin.H{
			"taskId":         singleTask.ID,
			"requirementId":  singleTask.RequirementID,
			"status":         singleTask.Status,
			"progress":       singleTask.Progress,
			"currentStage":   singleTask.CurrentStage,
			"stageDesc":      singleTask.StageDesc,
			"animationType":  singleTask.AnimationType,
			"errorMsg":       singleTask.ErrorMsg,
			"startTime":      singleTask.StartTime,
			"endTime":        singleTask.EndTime,
			
			// 阶段历史（用于展示分析步骤）
			"stageHistory":   singleTask.StageHistory,
			
			// 动画提示数据
			"animationData": gin.H{
				"currentStage":    singleTask.CurrentStage,
				"animationType":   singleTask.AnimationType,
				"progressPercent": singleTask.Progress,
				"isRunning":       singleTask.Status == "running",
				"isCompleted":     singleTask.Status == "completed",
				"isFailed":        singleTask.Status == "failed",
				
				// 阶段图标映射（前端可用于显示不同图标）
				"stageIcon": a.getStageIcon(singleTask.CurrentStage),
				"stageColor": a.getStageColor(singleTask.CurrentStage, singleTask.Status),
				
				// 预估剩余时间（基于历史阶段）
				"estimatedRemaining": a.calculateEstimatedTime(singleTask),
			},
			
			// 实时状态描述
			"statusText": a.getStatusText(singleTask),
		}

		// 如果任务完成，返回分析结果
		if singleTask.Status == "completed" && singleTask.Result != nil {
			result := singleTask.Result
			responseData["result"] = gin.H{
				"requirementId":  result.RequirementID,
				"analysisNote":   result.AnalysisNote,
				"confidence":     result.Confidence,
				"timestamp":      result.Timestamp,
				
				// 原始信息
				"originalInfo":   result.OriginalInfo,
				
				// 优化信息
				"optimizedInfo":  result.OptimizedInfo,
				
				// 对比摘要（便于前端显示）
				"comparisonSummary": gin.H{
					"titleChanged":       result.OriginalInfo.Title != result.OptimizedInfo.Title,
					"descriptionChanged": result.OriginalInfo.Description != result.OptimizedInfo.Description,
					"functionTypeChanged": result.OriginalInfo.FunctionType != result.OptimizedInfo.FunctionType,
					"complexityChanged":  result.OriginalInfo.Complexity != result.OptimizedInfo.Complexity,
					"businessValueChanged": result.OriginalInfo.BusinessValue != result.OptimizedInfo.BusinessValue,
					"afpChanged": (result.OriginalInfo.RecommendedAFP == nil && result.OptimizedInfo.RecommendedAFP != nil) ||
								 (result.OriginalInfo.RecommendedAFP != nil && result.OptimizedInfo.RecommendedAFP != nil && 
								  *result.OriginalInfo.RecommendedAFP != *result.OptimizedInfo.RecommendedAFP),
					"ufpChanged": (result.OriginalInfo.RecommendedUFP == nil && result.OptimizedInfo.RecommendedUFP != nil) ||
								 (result.OriginalInfo.RecommendedUFP != nil && result.OptimizedInfo.RecommendedUFP != nil && 
								  *result.OriginalInfo.RecommendedUFP != *result.OptimizedInfo.RecommendedUFP),
				},
				
				// 改进亮点（便于前端展示）
				"improvements": gin.H{
					"hasTitle":        result.OptimizedInfo.Title != "" && result.OptimizedInfo.Title != result.OriginalInfo.Title,
					"hasDescription":  result.OptimizedInfo.Description != "" && result.OptimizedInfo.Description != result.OriginalInfo.Description,
					"hasFunctionType": result.OptimizedInfo.FunctionType != "" && result.OptimizedInfo.FunctionType != result.OriginalInfo.FunctionType,
					"hasBusinessValue": result.OptimizedInfo.BusinessValue != "" && result.OptimizedInfo.BusinessValue != result.OriginalInfo.BusinessValue,
					"hasAfpRecommendation": result.OptimizedInfo.RecommendedAFP != nil && *result.OptimizedInfo.RecommendedAFP > 0,
					"hasUfpRecommendation": result.OptimizedInfo.RecommendedUFP != nil && *result.OptimizedInfo.RecommendedUFP > 0,
				},
				
				// 建议采纳理由
				"adoptionReasons": []string{
					"AI基于项目背景和上级需求进行了深度分析",
					"参考了同级需求保持一致性",
					fmt.Sprintf("分析置信度达到 %.0f%%", result.Confidence*100),
					"优化了功能描述的完整性和准确性",
				},
			}
		}

		response.OkWithData(responseData, c)
		return
	}

	// 如果不是单需求分析任务，尝试获取批量分析任务
	task, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.GetAnalysisProgress(uint(taskId))
	if err != nil {
		global.GVA_LOG.Error("获取分析进度失败!", zap.Error(err))
		response.FailWithMessage("获取分析进度失败", c)
		return
	}

	response.OkWithData(gin.H{
		"taskId":         task.ID,
		"status":         task.Status,
		"progress":       task.Progress,
		"totalCount":     task.TotalCount,
		"processedCount": task.ProcessedCount,
		"successCount":   task.SuccessCount,
		"failedCount":    task.FailedCount,
		"startTime":      task.StartTime,
		"endTime":        task.EndTime,
		"duration":       task.Duration,
		"summary":        task.Summary,
		"errorMsg":       task.ErrorMsg,
		"targetVersionId": task.TargetVersionID,
	}, c)
}

// GetAnalysisResult 获取分析结果详情
// @Tags NesmaRequirement
// @Summary 获取分析结果详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Router /nesma/requirement/analysis-result/{taskId} [get]
func (a *NesmaRequirementApi) GetAnalysisResult(c *gin.Context) {
	taskIdStr := c.Param("taskId")
	taskId, err := strconv.ParseUint(taskIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	// 获取单需求分析任务
	singleTask, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.GetSingleAnalysisTask(uint(taskId))
	if err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	if singleTask.Status != "completed" || singleTask.Result == nil {
		response.FailWithMessage("分析任务尚未完成", c)
		return
	}

	result := singleTask.Result
	
	// 构建详细的结果对比数据
	responseData := gin.H{
		"taskId":        singleTask.ID,
		"requirementId": result.RequirementID,
		"analysisTime":  result.Timestamp,
		"confidence":    result.Confidence,
		"analysisNote":  result.AnalysisNote,
		
		// 原始需求信息
		"original": gin.H{
			"title":          result.OriginalInfo.Title,
			"description":    result.OriginalInfo.Description,
			"functionType":   result.OriginalInfo.FunctionType,
			"complexity":     result.OriginalInfo.Complexity,
			"businessValue":  result.OriginalInfo.BusinessValue,
			"recommendedAFP": result.OriginalInfo.RecommendedAFP,
			"recommendedUFP": result.OriginalInfo.RecommendedUFP,
		},
		
		// AI优化建议
		"optimized": gin.H{
			"title":          result.OptimizedInfo.Title,
			"description":    result.OptimizedInfo.Description,
			"functionType":   result.OptimizedInfo.FunctionType,
			"complexity":     result.OptimizedInfo.Complexity,
			"businessValue":  result.OptimizedInfo.BusinessValue,
			"recommendedAFP": result.OptimizedInfo.RecommendedAFP,
			"recommendedUFP": result.OptimizedInfo.RecommendedUFP,
		},
		
		// 变更统计
		"changes": gin.H{
			"totalFields":   7, // 总字段数
			"changedFields": a.countChangedFields(result.OriginalInfo, result.OptimizedInfo),
			"improvementRate": a.calculateImprovementRate(result.OriginalInfo, result.OptimizedInfo),
		},
		
		// 改进亮点
		"highlights": a.buildImprovementHighlights(result.OriginalInfo, result.OptimizedInfo),
		
		// 置信度评级
		"confidenceLevel": a.getConfidenceLevel(result.Confidence),
		
		// 建议操作
		"recommendation": gin.H{
			"shouldAdopt": result.Confidence >= 0.7,
			"reasons": a.buildAdoptionReasons(result),
			"warnings": a.buildAdoptionWarnings(result),
		},
	}

	response.OkWithData(responseData, c)
}

// countChangedFields 计算变更字段数量
func (a *NesmaRequirementApi) countChangedFields(original, optimized *nesmaService.AnalysisRequirementInfo) int {
	count := 0
	if original.Title != optimized.Title { count++ }
	if original.Description != optimized.Description { count++ }
	if original.FunctionType != optimized.FunctionType { count++ }
	if original.Complexity != optimized.Complexity { count++ }
	if original.BusinessValue != optimized.BusinessValue { count++ }
	
	// AFP变更
	if (original.RecommendedAFP == nil && optimized.RecommendedAFP != nil) ||
	   (original.RecommendedAFP != nil && optimized.RecommendedAFP != nil && *original.RecommendedAFP != *optimized.RecommendedAFP) {
		count++
	}
	
	// UFP变更
	if (original.RecommendedUFP == nil && optimized.RecommendedUFP != nil) ||
	   (original.RecommendedUFP != nil && optimized.RecommendedUFP != nil && *original.RecommendedUFP != *optimized.RecommendedUFP) {
		count++
	}
	
	return count
}

// calculateImprovementRate 计算改进率
func (a *NesmaRequirementApi) calculateImprovementRate(original, optimized *nesmaService.AnalysisRequirementInfo) float64 {
	originalEmpty := 0
	optimizedFilled := 0
	
	if original.Title == "" { originalEmpty++ }
	if original.Description == "" { originalEmpty++ }
	if original.FunctionType == "" { originalEmpty++ }
	if original.BusinessValue == "" { originalEmpty++ }
	if original.RecommendedAFP == nil || *original.RecommendedAFP == 0 { originalEmpty++ }
	if original.RecommendedUFP == nil || *original.RecommendedUFP == 0 { originalEmpty++ }
	
	if optimized.Title != "" { optimizedFilled++ }
	if optimized.Description != "" { optimizedFilled++ }
	if optimized.FunctionType != "" { optimizedFilled++ }
	if optimized.BusinessValue != "" { optimizedFilled++ }
	if optimized.RecommendedAFP != nil && *optimized.RecommendedAFP > 0 { optimizedFilled++ }
	if optimized.RecommendedUFP != nil && *optimized.RecommendedUFP > 0 { optimizedFilled++ }
	
	if originalEmpty == 0 { return 0 }
	return float64(optimizedFilled) / float64(originalEmpty+optimizedFilled) * 100
}

// buildImprovementHighlights 构建改进亮点
func (a *NesmaRequirementApi) buildImprovementHighlights(original, optimized *nesmaService.AnalysisRequirementInfo) []string {
	highlights := []string{}
	
	if original.Title != optimized.Title && optimized.Title != "" {
		highlights = append(highlights, "优化了需求标题，更加准确明确")
	}
	if original.Description == "" && optimized.Description != "" {
		highlights = append(highlights, "补充了详细的功能描述，包含输入输出流程")
	}
	if original.FunctionType == "" && optimized.FunctionType != "" {
		highlights = append(highlights, fmt.Sprintf("确定了NESMA功能类型：%s", optimized.FunctionType))
	}
	if original.BusinessValue == "" && optimized.BusinessValue != "" {
		highlights = append(highlights, "明确了业务价值和效益")
	}
	if (original.RecommendedAFP == nil || *original.RecommendedAFP == 0) && 
	   (optimized.RecommendedAFP != nil && *optimized.RecommendedAFP > 0) {
		highlights = append(highlights, fmt.Sprintf("推荐AFP值：%.0f", *optimized.RecommendedAFP))
	}
	if (original.RecommendedUFP == nil || *original.RecommendedUFP == 0) && 
	   (optimized.RecommendedUFP != nil && *optimized.RecommendedUFP > 0) {
		highlights = append(highlights, fmt.Sprintf("推荐UFP值：%.0f", *optimized.RecommendedUFP))
	}
	
	return highlights
}

// getConfidenceLevel 获取置信度等级
func (a *NesmaRequirementApi) getConfidenceLevel(confidence float64) string {
	if confidence >= 0.9 { return "非常高" }
	if confidence >= 0.8 { return "高" }
	if confidence >= 0.7 { return "中等" }
	if confidence >= 0.6 { return "中低" }
	return "低"
}

// buildAdoptionReasons 构建采纳理由
func (a *NesmaRequirementApi) buildAdoptionReasons(result *nesmaService.SingleAnalysisResult) []string {
	reasons := []string{}
	
	if result.Confidence >= 0.8 {
		reasons = append(reasons, fmt.Sprintf("AI分析置信度高达%.0f%%", result.Confidence*100))
	}
	if result.OptimizedInfo.Description != "" && result.OriginalInfo.Description == "" {
		reasons = append(reasons, "补充了完整的功能描述")
	}
	if result.OptimizedInfo.FunctionType != "" && result.OriginalInfo.FunctionType == "" {
		reasons = append(reasons, "明确了NESMA功能类型分类")
	}
	if result.OptimizedInfo.BusinessValue != "" && result.OriginalInfo.BusinessValue == "" {
		reasons = append(reasons, "明确了业务价值和收益")
	}
	
	reasons = append(reasons, "基于项目背景和上级需求进行分析")
	reasons = append(reasons, "参考同级需求保持一致性")
	
	return reasons
}

// buildAdoptionWarnings 构建采纳警告
func (a *NesmaRequirementApi) buildAdoptionWarnings(result *nesmaService.SingleAnalysisResult) []string {
	warnings := []string{}
	
	if result.Confidence < 0.7 {
		warnings = append(warnings, "AI分析置信度较低，建议人工审核")
	}
	if result.OptimizedInfo.FunctionType != result.OriginalInfo.FunctionType && result.OriginalInfo.FunctionType != "" {
		warnings = append(warnings, "功能类型发生变更，请确认是否合适")
	}
	
	if len(warnings) == 0 {
		warnings = append(warnings, "请仔细核对优化内容是否符合实际需求")
	}
	
	return warnings
}

// getStageIcon 获取阶段图标
func (a *NesmaRequirementApi) getStageIcon(stage string) string {
	iconMap := map[string]string{
		"initializing":     "Setting",
		"validating":       "Search",
		"fetching":         "Download", 
		"context_building": "Connection",
		"ai_analyzing":     "Cpu",
		"ai_calling":       "CloudDownload",
		"parsing":          "DocumentCopy",
		"finalizing":       "Check",
		"completed":        "SuccessFilled",
		"failed":           "CloseBold",
	}
	
	if icon, exists := iconMap[stage]; exists {
		return icon
	}
	return "Loading"
}

// getStageColor 获取阶段颜色
func (a *NesmaRequirementApi) getStageColor(stage, status string) string {
	if status == "failed" {
		return "#F56C6C" // 红色
	}
	if status == "completed" {
		return "#67C23A" // 绿色
	}
	
	// 运行中的颜色
	colorMap := map[string]string{
		"initializing":     "#409EFF", // 蓝色
		"validating":       "#E6A23C", // 橙色
		"fetching":         "#409EFF", // 蓝色
		"context_building": "#909399", // 灰色
		"ai_analyzing":     "#722ED1", // 紫色
		"ai_calling":       "#13C2C2", // 青色
		"parsing":          "#52C41A", // 浅绿
		"finalizing":       "#1890FF", // 天蓝
	}
	
	if color, exists := colorMap[stage]; exists {
		return color
	}
	return "#409EFF" // 默认蓝色
}

// calculateEstimatedTime 计算预估剩余时间（秒）
func (a *NesmaRequirementApi) calculateEstimatedTime(task *nesmaService.SingleAnalysisTask) int {
	if task.Status != "running" {
		return 0
	}
	
	// 基于当前进度估算
	if task.Progress >= 95 {
		return 2
	} else if task.Progress >= 85 {
		return 5
	} else if task.Progress >= 70 {
		return 10
	} else if task.Progress >= 50 {
		return 15
	} else if task.Progress >= 25 {
		return 20
	}
	return 25
}

// getStatusText 获取状态描述文本
func (a *NesmaRequirementApi) getStatusText(task *nesmaService.SingleAnalysisTask) string {
	switch task.Status {
	case "pending":
		return "任务已创建，等待开始..."
	case "running":
		return fmt.Sprintf("%s (进度 %d%%)", task.StageDesc, task.Progress)
	case "completed":
		return "✨ 分析完成！AI已为您优化需求内容"
	case "failed":
		return fmt.Sprintf("❌ 分析失败：%s", task.ErrorMsg)
	default:
		return "未知状态"
	}
}

// AnalyzeHierarchicalRequirement 启动层次化需求分析
// @Tags NesmaRequirement  
// @Summary 启动层次化需求分析
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "层次化分析请求"
// @Success 200 {object} response.Response{data=object} "分析启动成功"
// @Router /nesma/requirement/hierarchical-analysis [post]
func (a *NesmaRequirementApi) AnalyzeHierarchicalRequirement(c *gin.Context) {
	var req nesmaService.HierarchicalAnalysisRequest
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 设置默认值
	if req.AnalysisType == "" {
		req.AnalysisType = "single"
	}
	if req.MaxL4Count == 0 {
		req.MaxL4Count = 8
	}

	// 调用层次化分析服务
	task, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.AnalyzeHierarchicalRequirementAsync(&req)
	if err != nil {
		global.GVA_LOG.Error("启动层次化需求分析失败!", zap.Error(err))
		response.FailWithMessage("启动层次化分析失败: "+err.Error(), c)
		return
	}

	// 返回任务信息
	response.OkWithData(gin.H{
		"taskId":       task.ID,
		"requirementId": task.RequirementID,
		"analysisType": task.AnalysisType,
		"status":       task.Status,
		"progress":     task.Progress,
		"message":      "层次化需求分析任务已启动",
		"config":       task.Config,
	}, c)
}

// GetHierarchicalAnalysisProgress 获取层次化分析进度
// @Tags NesmaRequirement
// @Summary 获取层次化分析进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Router /nesma/requirement/hierarchical-analysis-progress/{taskId} [get]
func (a *NesmaRequirementApi) GetHierarchicalAnalysisProgress(c *gin.Context) {
	taskIdStr := c.Param("taskId")
	taskId, err := strconv.ParseUint(taskIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	// 获取层次化分析任务
	task, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.GetHierarchicalAnalysisTask(uint(taskId))
	if err != nil {
		global.GVA_LOG.Error("获取层次化分析进度失败!", zap.Error(err))
		response.FailWithMessage("获取分析进度失败", c)
		return
	}

	// 构建详细的进度响应
	responseData := gin.H{
		"taskId":        task.ID,
		"requirementId": task.RequirementID,
		"analysisType":  task.AnalysisType,
		"status":        task.Status,
		"progress":      task.Progress,
		"currentStep":   task.CurrentStep,
		"stepDesc":      task.StepDesc,
		"animationType": task.AnimationType,
		"errorMsg":      task.ErrorMsg,
		"startTime":     task.StartTime,
		"endTime":       task.EndTime,
		"config":        task.Config,
		
		// 处理步骤详情
		"processingSteps": task.ProcessingSteps,
		
		// 增强的动画数据
		"animationData": gin.H{
			"currentStep":     task.CurrentStep,
			"animationType":   task.AnimationType,
			"progressPercent": task.Progress,
			"isRunning":       task.Status == "running",
			"isCompleted":     task.Status == "completed",
			"isFailed":        task.Status == "failed",
			"stageIcon":       a.getHierarchicalStageIcon(task.CurrentStep),
			"stageColor":      a.getHierarchicalStageColor(task.CurrentStep, task.Status),
			"estimatedRemaining": a.calculateHierarchicalEstimatedTime(task),
		},
		
		// 实时状态描述
		"statusText": a.getHierarchicalStatusText(task),
	}

	// 如果任务完成，返回详细结果
	if task.Status == "completed" && task.Result != nil {
		result := task.Result
		responseData["result"] = gin.H{
			"requirementId":       result.RequirementID,
			"analysisType":        result.AnalysisType,
			"analysisNote":        result.AnalysisNote,
			"confidence":          result.Confidence,
			"timestamp":           result.Timestamp,
			"originalInfo":        result.OriginalInfo,
			"optimizedInfo":       result.OptimizedInfo,
			"knowledgeReferences": result.KnowledgeReferences,
			
			// L4生成结果
			"l4GenerationResult": result.L4GenerationResult,
			
			// 分析摘要
			"analysisSummary": gin.H{
				"hasL4Generation":     result.L4GenerationResult != nil,
				"l4GenerationStrategy": func() string {
					if result.L4GenerationResult != nil {
						return result.L4GenerationResult.Strategy
					}
					return ""
				}(),
				"generatedL4Count": func() int {
					if result.L4GenerationResult != nil {
						return result.L4GenerationResult.GeneratedL4Count
					}
					return 0
				}(),
				"optimizedL4Count": func() int {
					if result.L4GenerationResult != nil {
						return result.L4GenerationResult.OptimizedL4Count
					}
					return 0
				}(),
				"totalImprovements": a.countHierarchicalImprovements(result),
			},
			
			// 推荐操作
			"recommendedActions": func() []string {
				if result.L4GenerationResult != nil {
					return result.L4GenerationResult.RecommendedActions
				}
				return []string{"请仔细审核分析结果"}
			}(),
		}
	}

	response.OkWithData(responseData, c)
}

// GetHierarchicalAnalysisResult 获取层次化分析结果详情
// @Tags NesmaRequirement
// @Summary 获取层次化分析结果详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path int true "任务ID"
// @Success 200 {object} response.Response{data=object} "获取成功"
// @Router /nesma/requirement/hierarchical-analysis-result/{taskId} [get]
func (a *NesmaRequirementApi) GetHierarchicalAnalysisResult(c *gin.Context) {
	taskIdStr := c.Param("taskId")
	taskId, err := strconv.ParseUint(taskIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	// 获取层次化分析任务
	task, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.GetHierarchicalAnalysisTask(uint(taskId))
	if err != nil {
		response.FailWithMessage("任务不存在", c)
		return
	}

	if task.Status != "completed" || task.Result == nil {
		response.FailWithMessage("分析任务尚未完成", c)
		return
	}

	result := task.Result
	
	// 构建详细的结果数据
	responseData := gin.H{
		"taskId":        task.ID,
		"requirementId": result.RequirementID,
		"analysisType":  result.AnalysisType,
		"analysisTime":  result.Timestamp,
		"confidence":    result.Confidence,
		"analysisNote":  result.AnalysisNote,
		
		// L3需求分析结果
		"requirementAnalysis": gin.H{
			"original":  result.OriginalInfo,
			"optimized": result.OptimizedInfo,
			"hasChanges": a.hasRequirementChanges(result.OriginalInfo, result.OptimizedInfo),
		},
		
		// L4生成结果
		"l4GenerationResult": result.L4GenerationResult,
		
		// 知识库引用
		"knowledgeReferences": result.KnowledgeReferences,
		
		// 分析统计
		"analysisStats": gin.H{
			"totalL4Suggestions": func() int {
				if result.L4GenerationResult != nil {
					return len(result.L4GenerationResult.L4Suggestions)
				}
				return 0
			}(),
			"confidenceLevel": a.getConfidenceLevel(result.Confidence),
			"processingTime": func() string {
				if task.EndTime != nil {
					duration := task.EndTime.Sub(task.StartTime)
					return fmt.Sprintf("%.1f秒", duration.Seconds())
				}
				return "未知"
			}(),
		},
		
		// 推荐操作
		"recommendations": gin.H{
			"shouldAdoptL3": result.Confidence >= 0.7,
			"shouldCreateL4": result.L4GenerationResult != nil && len(result.L4GenerationResult.L4Suggestions) > 0,
			"adoptionReasons": a.buildHierarchicalAdoptionReasons(result),
			"warnings": a.buildHierarchicalAdoptionWarnings(result),
		},
	}

	response.OkWithData(responseData, c)
}

// CreateL4Requirements 创建L4功能点
// @Tags NesmaRequirement
// @Summary 创建L4功能点
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "L4创建请求"
// @Success 200 {object} response.Response{data=object} "创建成功"
// @Router /nesma/requirement/create-l4 [post]
func (a *NesmaRequirementApi) CreateL4Requirements(c *gin.Context) {
	var req struct {
		ParentID    uint                         `json:"parentId" binding:"required"`
		CycleID     uint                         `json:"cycleId" binding:"required"`
		Suggestions []nesmaService.L4Suggestion `json:"suggestions" binding:"required"`
	}
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 获取L4生成器服务
	l4Generator := nesmaService.GetLevel4GeneratorService()
	
	// 构建创建请求
	creations := make([]nesmaService.Level4CreationRequest, len(req.Suggestions))
	for i, suggestion := range req.Suggestions {
		creations[i] = nesmaService.Level4CreationRequest{
			ParentID:   req.ParentID,
			Suggestion: &nesmaService.Level4Suggestion{
				SuggestedTitle:       suggestion.SuggestedTitle,
				SuggestedDescription: suggestion.SuggestedDescription,
				SuggestedCode:        suggestion.SuggestedCode,
				FunctionType:         suggestion.FunctionType,
				BusinessValue:        suggestion.BusinessValue,
				AcceptanceCriteria:   suggestion.AcceptanceCriteria,
				EstimatedComplexity:  suggestion.EstimatedComplexity,
				RecommendedAFP:       suggestion.RecommendedAFP,
				RecommendedUFP:       suggestion.RecommendedUFP,
				Priority:             suggestion.Priority,
				Confidence:           suggestion.Confidence,
				GenerationReason:     suggestion.GenerationReason,
				RelatedKnowledge:     suggestion.RelatedKnowledge,
			},
		}
	}
	
	// 批量创建L4功能点
	result, err := l4Generator.BatchCreateLevel4Requirements(req.CycleID, creations)
	if err != nil {
		global.GVA_LOG.Error("创建L4功能点失败!", zap.Error(err))
		response.FailWithMessage("创建L4功能点失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"successCount":        result.SuccessCount,
		"failedCount":         result.FailedCount,
		"createdRequirements": result.CreatedRequirements,
		"errors":              result.Errors,
		"message":             fmt.Sprintf("成功创建 %d 个L4功能点", result.SuccessCount),
	}, c)
}

// ApplyHierarchicalAnalysisResult 应用层次化分析结果
// @Tags NesmaRequirement
// @Summary 应用层次化分析结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "应用分析结果请求"
// @Success 200 {object} response.Response{msg=string} "应用成功"
// @Router /nesma/requirement/apply-hierarchical-analysis [post]
func (a *NesmaRequirementApi) ApplyHierarchicalAnalysisResult(c *gin.Context) {
	var req struct {
		TaskID               uint                        `json:"taskId" binding:"required"`
		RequirementID        uint                        `json:"requirementId" binding:"required"`
		ApplyL3Optimization  bool                        `json:"applyL3Optimization"`
		ApplyL4Generation    bool                        `json:"applyL4Generation"`
		SelectedL4Suggestions []int                      `json:"selectedL4Suggestions"`
		UpdateStatus         bool                        `json:"updateStatus"`
		OptimizedInfo        *nesmaService.AnalysisRequirementInfo `json:"optimizedInfo"`
		AnalysisNote         string                      `json:"analysisNote"`
	}
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 获取分析任务和结果
	task, err := service.ServiceGroupApp.NesmaServiceGroup.SingleRequirementAnalysisService.GetHierarchicalAnalysisTask(req.TaskID)
	if err != nil || task.Result == nil {
		response.FailWithMessage("分析任务不存在或未完成", c)
		return
	}

	result := task.Result
	
	// 应用L3优化（如果选择）
	if req.ApplyL3Optimization && req.OptimizedInfo != nil {
		var requirement nesma.NesmaRequirement
		if err := global.GVA_DB.First(&requirement, req.RequirementID).Error; err != nil {
			response.FailWithMessage("需求不存在", c)
			return
		}

		// 应用优化信息
		requirement.Title = req.OptimizedInfo.Title
		requirement.Description = req.OptimizedInfo.Description
		requirement.FunctionType = req.OptimizedInfo.FunctionType
		requirement.Complexity = req.OptimizedInfo.Complexity
		requirement.BusinessValue = req.OptimizedInfo.BusinessValue
		
		if req.OptimizedInfo.RecommendedAFP != nil {
			requirement.AFP = *req.OptimizedInfo.RecommendedAFP
		}
		if req.OptimizedInfo.RecommendedUFP != nil {
			requirement.UFP = *req.OptimizedInfo.RecommendedUFP
		}

		// 更新分析相关字段
		requirement.AIAnalysisStatus = "applied"
		requirement.AIGeneratedTitle = req.OptimizedInfo.Title
		requirement.AIDescription = req.OptimizedInfo.Description
		analysisTime := time.Now()
		requirement.AIAnalysisTime = &analysisTime
		requirement.Notes = req.AnalysisNote

		// 如果需要更新状态为已完成
		if req.UpdateStatus {
			requirement.Status = "completed"
		}

		// 保存更新
		if err := global.GVA_DB.Save(&requirement).Error; err != nil {
			global.GVA_LOG.Error("应用L3优化失败!", zap.Error(err))
			response.FailWithMessage("应用L3优化失败: "+err.Error(), c)
			return
		}
	}

	// 应用L4生成（如果选择）
	var createdL4Count int
	if req.ApplyL4Generation && result.L4GenerationResult != nil {
		l4Generator := nesmaService.GetLevel4GeneratorService()
		
		// 筛选选中的L4建议
		selectedSuggestions := []nesmaService.L4Suggestion{}
		for _, index := range req.SelectedL4Suggestions {
			if index >= 0 && index < len(result.L4GenerationResult.L4Suggestions) {
				selectedSuggestions = append(selectedSuggestions, result.L4GenerationResult.L4Suggestions[index])
			}
		}

		// 构建创建请求
		creations := make([]nesmaService.Level4CreationRequest, len(selectedSuggestions))
		for i, suggestion := range selectedSuggestions {
			creations[i] = nesmaService.Level4CreationRequest{
				ParentID: req.RequirementID,
				Suggestion: &nesmaService.Level4Suggestion{
					SuggestedTitle:       suggestion.SuggestedTitle,
					SuggestedDescription: suggestion.SuggestedDescription,
					SuggestedCode:        suggestion.SuggestedCode,
					FunctionType:         suggestion.FunctionType,
					BusinessValue:        suggestion.BusinessValue,
					AcceptanceCriteria:   suggestion.AcceptanceCriteria,
					EstimatedComplexity:  suggestion.EstimatedComplexity,
					RecommendedAFP:       suggestion.RecommendedAFP,
					RecommendedUFP:       suggestion.RecommendedUFP,
					Priority:             suggestion.Priority,
					Confidence:           suggestion.Confidence,
					GenerationReason:     suggestion.GenerationReason,
					RelatedKnowledge:     suggestion.RelatedKnowledge,
				},
			}
		}
		
		// 批量创建L4功能点
		batchResult, err := l4Generator.BatchCreateLevel4Requirements(task.CycleID, creations)
		if err != nil {
			global.GVA_LOG.Error("创建L4功能点失败!", zap.Error(err))
		} else {
			createdL4Count = batchResult.SuccessCount
		}
	}

	// 记录操作日志
	global.GVA_LOG.Info("应用层次化分析结果", 
		zap.Uint("taskId", req.TaskID),
		zap.Uint("requirementId", req.RequirementID),
		zap.Bool("applyL3Optimization", req.ApplyL3Optimization),
		zap.Bool("applyL4Generation", req.ApplyL4Generation),
		zap.Int("createdL4Count", createdL4Count))

	response.OkWithData(gin.H{
		"message": "层次化分析结果已成功应用",
		"appliedL3Optimization": req.ApplyL3Optimization,
		"appliedL4Generation": req.ApplyL4Generation,
		"createdL4Count": createdL4Count,
	}, c)
}

// 层次化分析相关的辅助方法

// getHierarchicalStageIcon 获取层次化分析阶段图标
func (a *NesmaRequirementApi) getHierarchicalStageIcon(stage string) string {
	iconMap := map[string]string{
		"validating_l3":        "Search",
		"analyzing_existing_l4": "List",
		"building_context":     "Connection",
		"searching_knowledge":  "Search",
		"generating_l4":        "Magic",
		"analyzing_l3":         "Edit",
		"integrating_results":  "Check",
		"fetching_existing_l4": "Download",
		"optimizing_l4s":       "Refresh",
		"initializing":         "Setting",
		"validating":           "Search",
		"context_building":     "Connection",
		"ai_analyzing":         "Cpu",
		"ai_calling":           "CloudDownload",
		"parsing":              "DocumentCopy",
		"finalizing":           "Check",
		"completed":            "SuccessFilled",
		"failed":               "CloseBold",
	}
	
	if icon, exists := iconMap[stage]; exists {
		return icon
	}
	return "Loading"
}

// getHierarchicalStageColor 获取层次化分析阶段颜色
func (a *NesmaRequirementApi) getHierarchicalStageColor(stage, status string) string {
	if status == "failed" {
		return "#F56C6C" // 红色
	}
	if status == "completed" {
		return "#67C23A" // 绿色
	}
	
	// 运行中的颜色
	colorMap := map[string]string{
		"validating_l3":        "#409EFF", // 蓝色
		"analyzing_existing_l4": "#E6A23C", // 橙色
		"building_context":     "#909399", // 灰色
		"searching_knowledge":  "#13C2C2", // 青色
		"generating_l4":        "#722ED1", // 紫色
		"analyzing_l3":         "#52C41A", // 浅绿
		"integrating_results":  "#1890FF", // 天蓝
		"fetching_existing_l4": "#409EFF", // 蓝色
		"optimizing_l4s":       "#F5A623", // 黄色
		"initializing":         "#409EFF", // 蓝色
		"validating":           "#E6A23C", // 橙色
		"context_building":     "#909399", // 灰色
		"ai_analyzing":         "#722ED1", // 紫色
		"ai_calling":           "#13C2C2", // 青色
		"parsing":              "#52C41A", // 浅绿
		"finalizing":           "#1890FF", // 天蓝
	}
	
	if color, exists := colorMap[stage]; exists {
		return color
	}
	return "#409EFF" // 默认蓝色
}

// calculateHierarchicalEstimatedTime 计算层次化分析预估剩余时间
func (a *NesmaRequirementApi) calculateHierarchicalEstimatedTime(task *nesmaService.HierarchicalAnalysisTask) int {
	if task.Status != "running" {
		return 0
	}
	
	// 根据分析类型和当前进度估算
	switch task.AnalysisType {
	case "generate_l4":
		if task.Progress >= 90 {
			return 5
		} else if task.Progress >= 70 {
			return 15
		} else if task.Progress >= 50 {
			return 30
		} else if task.Progress >= 30 {
			return 45
		}
		return 60
	case "optimize_l4":
		if task.Progress >= 90 {
			return 3
		} else if task.Progress >= 60 {
			return 10
		} else if task.Progress >= 30 {
			return 20
		}
		return 30
	default:
		// 普通单需求分析
		if task.Progress >= 85 {
			return 5
		} else if task.Progress >= 70 {
			return 10
		} else if task.Progress >= 50 {
			return 15
		}
		return 20
	}
}

// getHierarchicalStatusText 获取层次化分析状态描述
func (a *NesmaRequirementApi) getHierarchicalStatusText(task *nesmaService.HierarchicalAnalysisTask) string {
	switch task.Status {
	case "pending":
		return "任务已创建，等待开始..."
	case "running":
		return fmt.Sprintf("%s (进度 %d%%)", task.StepDesc, task.Progress)
	case "completed":
		switch task.AnalysisType {
		case "generate_l4":
			return "✨ L4生成完成！AI已为您生成功能点建议"
		case "optimize_l4":
			return "✨ L4优化完成！AI已为您优化现有功能点"
		default:
			return "✨ 增强分析完成！AI已为您优化需求内容"
		}
	case "failed":
		return fmt.Sprintf("❌ 分析失败：%s", task.ErrorMsg)
	default:
		return "未知状态"
	}
}

// countHierarchicalImprovements 计算层次化分析改进数量
func (a *NesmaRequirementApi) countHierarchicalImprovements(result *nesmaService.HierarchicalAnalysisResult) int {
	count := 0
	
	// L3改进计数
	if result.OriginalInfo != nil && result.OptimizedInfo != nil {
		if result.OriginalInfo.Title != result.OptimizedInfo.Title {
			count++
		}
		if result.OriginalInfo.Description != result.OptimizedInfo.Description {
			count++
		}
		if result.OriginalInfo.FunctionType != result.OptimizedInfo.FunctionType {
			count++
		}
		if result.OriginalInfo.BusinessValue != result.OptimizedInfo.BusinessValue {
			count++
		}
	}
	
	// L4生成计数
	if result.L4GenerationResult != nil {
		count += len(result.L4GenerationResult.L4Suggestions)
	}
	
	return count
}

// hasRequirementChanges 检查需求是否有变更
func (a *NesmaRequirementApi) hasRequirementChanges(original, optimized *nesmaService.AnalysisRequirementInfo) bool {
	if original == nil || optimized == nil {
		return false
	}
	
	return original.Title != optimized.Title ||
		   original.Description != optimized.Description ||
		   original.FunctionType != optimized.FunctionType ||
		   original.Complexity != optimized.Complexity ||
		   original.BusinessValue != optimized.BusinessValue
}

// buildHierarchicalAdoptionReasons 构建层次化分析采纳理由
func (a *NesmaRequirementApi) buildHierarchicalAdoptionReasons(result *nesmaService.HierarchicalAnalysisResult) []string {
	reasons := []string{}
	
	if result.Confidence >= 0.8 {
		reasons = append(reasons, fmt.Sprintf("AI分析置信度高达%.0f%%", result.Confidence*100))
	}
	
	if result.L4GenerationResult != nil {
		switch result.L4GenerationResult.Strategy {
		case "create_new":
			reasons = append(reasons, "基于L3需求全新生成L4功能点")
		case "expand":
			reasons = append(reasons, "在现有L4基础上智能扩展补充")
		case "optimize_existing":
			reasons = append(reasons, "优化现有L4功能点描述和分类")
		}
		
		if len(result.L4GenerationResult.L4Suggestions) > 0 {
			reasons = append(reasons, fmt.Sprintf("生成了%d个高质量L4功能点建议", len(result.L4GenerationResult.L4Suggestions)))
		}
	}
	
	if len(result.KnowledgeReferences) > 0 {
		reasons = append(reasons, "基于丰富的知识库内容进行分析")
	}
	
	reasons = append(reasons, "充分考虑项目背景和需求层级关系")
	
	return reasons
}

// buildHierarchicalAdoptionWarnings 构建层次化分析采纳警告
func (a *NesmaRequirementApi) buildHierarchicalAdoptionWarnings(result *nesmaService.HierarchicalAnalysisResult) []string {
	warnings := []string{}
	
	if result.Confidence < 0.7 {
		warnings = append(warnings, "AI分析置信度较低，建议人工审核")
	}
	
	if result.L4GenerationResult != nil && len(result.L4GenerationResult.L4Suggestions) > 10 {
		warnings = append(warnings, "生成的L4功能点较多，建议筛选重要的进行创建")
	}
	
	if result.AnalysisType == "generate_l4" {
		warnings = append(warnings, "L4功能点创建后将影响项目结构，请谨慎选择")
	}
	
	if len(warnings) == 0 {
		warnings = append(warnings, "请仔细核对分析内容是否符合实际需求")
	}
	
	return warnings
}
