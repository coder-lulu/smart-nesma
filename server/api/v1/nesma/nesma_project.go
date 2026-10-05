package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NesmaProjectApi struct{}

var nesmaProjectService = service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectService

// GetProjectOptions 获取项目选项列表
// @Tags NesmaProject
// @Summary 获取项目选项列表
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /nesma/project/options [get]
func (nesmaProjectApi *NesmaProjectApi) GetProjectOptions(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	projects, err := nesmaProjectService.GetProjectOptions(userID)
	if err != nil {
		global.GVA_LOG.Error("获取项目选项失败", zap.Error(err))
		response.FailWithMessage("获取项目选项失败: "+err.Error(), c)
		return
	}
	
	response.OkWithData(projects, c)
}

// CreateNesmaProject 创建NESMA项目
// @Tags NesmaProject
// @Summary 创建NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectCreate true "创建NESMA项目"
// @Success 200 {object} response.Response{msg=string} "创建成功"
// @Router /nesma/project [post]
func (nesmaProjectApi *NesmaProjectApi) CreateNesmaProject(c *gin.Context) {
	var req nesmaReq.NesmaProjectCreate
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 获取当前用户ID
	userID := utils.GetUserID(c)

	err = nesmaProjectService.CreateNesmaProjectFromRequest(&req, userID)
	if err != nil {
		global.GVA_LOG.Error("创建失败!", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("创建成功", c)
}

// DeleteNesmaProject 删除NESMA项目
// @Tags NesmaProject
// @Summary 删除NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectByID true "删除NESMA项目"
// @Success 200 {object} response.Response{msg=string} "删除成功"
// @Router /nesma/project [delete]
func (nesmaProjectApi *NesmaProjectApi) DeleteNesmaProject(c *gin.Context) {
	var req nesmaReq.NesmaProjectByID
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	err = nesmaProjectService.DeleteNesmaProject(req.ID, userID)
	if err != nil {
		global.GVA_LOG.Error("删除失败!", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

// DeleteNesmaProjectByIds 批量删除NESMA项目
// @Tags NesmaProject
// @Summary 批量删除NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectDeleteByIds true "批量删除NESMA项目"
// @Success 200 {object} response.Response{msg=string} "批量删除成功"
// @Router /nesma/project/delete-batch [delete]
func (nesmaProjectApi *NesmaProjectApi) DeleteNesmaProjectByIds(c *gin.Context) {
	var req nesmaReq.NesmaProjectDeleteByIds
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	err = nesmaProjectService.DeleteNesmaProjectByIds(req.IDs, userID)
	if err != nil {
		global.GVA_LOG.Error("批量删除失败!", zap.Error(err))
		response.FailWithMessage("批量删除失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("批量删除成功", c)
}

// UpdateNesmaProject 更新NESMA项目
// @Tags NesmaProject
// @Summary 更新NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectUpdate true "更新NESMA项目"
// @Success 200 {object} response.Response{msg=string} "更新成功"
// @Router /nesma/project [put]
func (nesmaProjectApi *NesmaProjectApi) UpdateNesmaProject(c *gin.Context) {
	var req nesmaReq.NesmaProjectUpdate
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)

	err = nesmaProjectService.UpdateNesmaProjectFromRequest(&req, userID)
	if err != nil {
		global.GVA_LOG.Error("更新失败!", zap.Error(err))
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("更新成功", c)
}

// FindNesmaProject 用ID查询NESMA项目
// @Tags NesmaProject
// @Summary 用ID查询NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param ID path int true "项目ID"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaProjectResponse,msg=string} "查询成功"
// @Router /nesma/project/{ID} [get]
func (nesmaProjectApi *NesmaProjectApi) FindNesmaProject(c *gin.Context) {
	ID := c.Param("ID")
	projectID, err := strconv.ParseUint(ID, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	userID := utils.GetUserID(c)
	project, err := nesmaProjectService.GetNesmaProject(uint(projectID), userID)
	if err != nil {
		global.GVA_LOG.Error("查询失败!", zap.Error(err))
		response.FailWithMessage("查询失败: "+err.Error(), c)
		return
	}

	// 获取项目的需求统计
	stats, err := nesmaProjectService.GetProjectRequirementStats(uint(projectID))
	if err != nil {
		global.GVA_LOG.Error("获取项目需求统计失败!", zap.Error(err))
		// 如果获取统计失败，使用空统计
		stats = nesmaRes.RequirementStats{}
	}

	response.OkWithDetailed(nesmaRes.NesmaProjectResponse{
		Project:          project,
		RequirementStats: stats,
	}, "查询成功", c)
}

// GetNesmaProjectList 分页获取NESMA项目列表
// @Tags NesmaProject
// @Summary 分页获取NESMA项目列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query nesmaReq.NesmaProjectSearch true "分页获取NESMA项目列表"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaProjectListResponse,msg=string} "获取成功"
// @Router /nesma/project/list [get]
func (nesmaProjectApi *NesmaProjectApi) GetNesmaProjectList(c *gin.Context) {
	var pageInfo nesmaReq.NesmaProjectSearch
	err := c.ShouldBindQuery(&pageInfo)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	list, total, err := nesmaProjectService.GetNesmaProjectInfoList(pageInfo, userID)
	if err != nil {
		global.GVA_LOG.Error("获取失败!", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	// 为每个项目添加需求统计
	var listWithStats []nesmaRes.NesmaProjectWithStats
	for _, project := range list {
		stats, err := nesmaProjectService.GetProjectRequirementStats(project.ID)
		if err != nil {
			global.GVA_LOG.Error("获取项目需求统计失败!", zap.Error(err))
			// 如果获取统计失败，使用空统计
			stats = nesmaRes.RequirementStats{}
		}

		projectWithStats := nesmaRes.NesmaProjectWithStats{
			NesmaProject:     project,
			RequirementStats: stats,
		}
		listWithStats = append(listWithStats, projectWithStats)
	}

	response.OkWithDetailed(nesmaRes.NesmaProjectListResponse{
		List:     listWithStats,
		Total:    total,
		Page:     pageInfo.Page,
		PageSize: pageInfo.PageSize,
	}, "获取成功", c)
}

// GetNesmaProjectStats 获取项目统计信息
// @Tags NesmaProject
// @Summary 获取项目统计信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID，不传则获取所有项目统计"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaProjectStatsResponse,msg=string} "获取成功"
// @Router /nesma/project/stats [get]
func (nesmaProjectApi *NesmaProjectApi) GetNesmaProjectStats(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	// 检查是否传递了项目ID
	var projectID *uint
	if projectIDStr := c.Query("projectId"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			projectIDVal := uint(id)
			projectID = &projectIDVal
		} else {
			response.FailWithMessage("项目ID格式错误", c)
			return
		}
	} 
	
	stats, err := nesmaProjectService.GetNesmaProjectStats(userID, projectID)
	if err != nil {
		global.GVA_LOG.Error("获取统计信息失败!", zap.Error(err))
		response.FailWithMessage("获取统计信息失败: "+err.Error(), c)
		return
	}

	// 转换统计数据格式
	statsResponse := nesmaRes.NesmaProjectStatsResponse{
		TotalProjects:  stats["totalProjects"].(int64),
		RecentProjects: stats["recentProjects"].(int64),
	}

	// 添加项目列表到响应中（用于工作台显示）
	if recentProjectsList, ok := stats["recentProjectsList"].([]nesma.NesmaProject); ok {
		for _, project := range recentProjectsList {
			// 获取项目的需求统计
			requirementStats, err := nesmaProjectService.GetProjectRequirementStats(project.ID)
			var reqCount, fpCount int64
			if err == nil {
				reqCount = requirementStats.TotalCount
				fpCount = requirementStats.Level4 // 功能点计数项数
			}
			
			statsResponse.RecentProjectsList = append(statsResponse.RecentProjectsList, nesmaRes.RecentProject{
				ID:               project.ID,
				Name:             project.Name,
				Status:           project.Status,
				RequirementCount: reqCount,
				FunctionPoints:   fpCount,
			})
		}
	}

	// 添加所有项目列表到响应中（用于批量操作）
	if allProjectsList, ok := stats["allProjectsList"].([]nesma.NesmaProject); ok {
		for _, project := range allProjectsList {
			statsResponse.Projects = append(statsResponse.Projects, nesmaRes.ProjectBasicInfo{
				ID:     project.ID,
				Name:   project.Name,
				Status: project.Status,
			})
		}
	}

	// 转换状态统计
	if statusStats, ok := stats["statusStats"].([]struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}); ok {
		for _, stat := range statusStats {
			statsResponse.StatusStats = append(statsResponse.StatusStats, nesmaRes.StatusStat{
				Status: stat.Status,
				Count:  stat.Count,
			})
		}
	}

	// 转换领域统计
	if domainStats, ok := stats["domainStats"].([]struct {
		Domain string `json:"domain"`
		Count  int64  `json:"count"`
	}); ok {
		for _, stat := range domainStats {
			statsResponse.DomainStats = append(statsResponse.DomainStats, nesmaRes.DomainStat{
				Domain: stat.Domain,
				Count:  stat.Count,
			})
		}
	}

	response.OkWithDetailed(statsResponse, "获取成功", c)
}

// ArchiveNesmaProject 归档项目
// @Tags NesmaProject
// @Summary 归档项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectByID true "归档项目"
// @Success 200 {object} response.Response{msg=string} "归档成功"
// @Router /nesma/project/archive [post]
func (nesmaProjectApi *NesmaProjectApi) ArchiveNesmaProject(c *gin.Context) {
	var req nesmaReq.NesmaProjectByID
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	err = nesmaProjectService.ArchiveNesmaProject(req.ID, userID)
	if err != nil {
		global.GVA_LOG.Error("归档失败!", zap.Error(err))
		response.FailWithMessage("归档失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("归档成功", c)
}

// RestoreNesmaProject 恢复项目
// @Tags NesmaProject
// @Summary 恢复项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectByID true "恢复项目"
// @Success 200 {object} response.Response{msg=string} "恢复成功"
// @Router /nesma/project/restore [post]
func (nesmaProjectApi *NesmaProjectApi) RestoreNesmaProject(c *gin.Context) {
	var req nesmaReq.NesmaProjectByID
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	err = nesmaProjectService.RestoreNesmaProject(req.ID, userID)
	if err != nil {
		global.GVA_LOG.Error("恢复失败!", zap.Error(err))
		response.FailWithMessage("恢复失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("恢复成功", c)
}

// OneClickAnalysis 一键分析项目
// @Tags NesmaProject
// @Summary 一键分析项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaOneClickAnalysisRequest true "一键分析项目"
// @Success 200 {object} response.Response{data=gin.H,msg=string} "分析任务启动成功"
// @Router /nesma/project/one-click-analysis [post]
func (nesmaProjectApi *NesmaProjectApi) OneClickAnalysis(c *gin.Context) {
	var req nesmaReq.NesmaOneClickAnalysisRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)

	// 验证项目权限
	_, err = nesmaProjectService.GetNesmaProject(req.ProjectID, userID)
	if err != nil {
		global.GVA_LOG.Error("项目权限验证失败!", zap.Error(err))
		response.FailWithMessage("项目权限验证失败: "+err.Error(), c)
		return
	}

	// 构建分析请求
	analysisReq := nesmaReq.NesmaAnalyzeRequest{
		ProjectID: req.ProjectID,
		CycleID:   req.CycleID,
		Config: struct {
			AnalysisType    string   `json:"analysisType"`
			TargetLevels    []int    `json:"targetLevels"`
			AIModel         string   `json:"aiModel"`
			EnableMermaid   bool     `json:"enableMermaid"`
			BatchSize       int      `json:"batchSize"`
			MaxRetries      int      `json:"maxRetries"`
			SkipCompleted   bool     `json:"skipCompleted"`
		}{
			AnalysisType:  "requirement_optimization",
			TargetLevels:  []int{3, 4}, // 分析3-4级需求
			AIModel:       "deepseek",
			EnableMermaid: true,
			BatchSize:     10,
			MaxRetries:    3,
			SkipCompleted: false,
		},
	}

	// 获取需求分析服务
	requirementAnalysisService := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService

	// 启动分析任务
	task, err := requirementAnalysisService.StartAnalysis(&analysisReq)
	if err != nil {
		global.GVA_LOG.Error("启动一键分析失败!", zap.Error(err))
		response.FailWithMessage("启动一键分析失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"taskId":     task.ID,
		"status":     task.Status,
		"progress":   task.Progress,
		"projectId":  req.ProjectID,
		"cycleId":    req.CycleID,
		"message":    "一键分析任务已启动",
		"startTime":  task.StartTime,
	}, c)
}
