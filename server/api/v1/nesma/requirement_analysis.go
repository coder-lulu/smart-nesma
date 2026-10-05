package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RequirementAnalysisApi struct{}

// MVPStartAnalysis MVP版本：一键分析需求
func (ra *RequirementAnalysisApi) MVPStartAnalysis(c *gin.Context) {
	var req nesmaReq.NesmaAnalyzeRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	task, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.StartAnalysis(&req)
	if err != nil {
		global.GVA_LOG.Error("启动需求分析失败!", zap.Error(err))
		response.FailWithMessage("启动需求分析失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{
		"taskId":     task.ID,
		"status":     task.Status,
		"progress":   task.Progress,
		"message":    "需求分析任务已启动",
	}, c)
}

// MVPGetAnalysisProgress MVP版本：获取分析进度
func (ra *RequirementAnalysisApi) MVPGetAnalysisProgress(c *gin.Context) {
	taskIdStr := c.Param("taskId")
	taskId, err := strconv.ParseUint(taskIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	task, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.GetAnalysisProgress(uint(taskId))
	if err != nil {
		global.GVA_LOG.Error("获取分析进度失败!", zap.Error(err))
		response.FailWithMessage("获取分析进度失败", c)
		return
	}

	response.OkWithData(gin.H{
		"task_id":        task.ID,
		"status":         task.Status,
		"progress":       task.Progress,
		"total_count":    task.TotalCount,
		"processed_count": task.ProcessedCount,
		"success_count":  task.SuccessCount,
		"failed_count":   task.FailedCount,
		"ai_call_count":  task.AICallCount,
		"start_time":     task.StartTime,
		"end_time":       task.EndTime,
		"duration":       task.Duration,
		"summary":        task.Summary,
		"result":         task.Result,
		"created_at":     task.CreatedAt,
		"updated_at":     task.UpdatedAt,
	}, c)
}

// MVPGetAnalysisTasks MVP版本：获取分析任务列表（支持分页和筛选）
func (ra *RequirementAnalysisApi) MVPGetAnalysisTasks(c *gin.Context) {
	var req nesmaReq.AnalysisTaskListRequest
	
	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	
	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // 限制最大分页大小
	}
	
	global.GVA_LOG.Info("获取分析任务列表", 
		zap.Int("page", req.Page),
		zap.Int("pageSize", req.PageSize),
		zap.Any("projectId", req.ProjectID),
		zap.Any("cycleId", req.CycleID),
		zap.String("taskType", req.TaskType),
		zap.String("status", req.Status),
		zap.String("startDate", req.StartDate),
		zap.String("endDate", req.EndDate))

	tasks, total, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.GetAnalysisTasksWithPagination(&req)
	if err != nil {
		global.GVA_LOG.Error("获取分析任务列表失败!", zap.Error(err))
		response.FailWithMessage("获取分析任务列表失败: "+err.Error(), c)
		return
	}

	// 计算分页信息
	totalPages := int((total + int64(req.PageSize) - 1) / int64(req.PageSize))
	
	response.OkWithData(gin.H{
		"list":        tasks,
		"total":       total,
		"page":        req.Page,
		"pageSize":    req.PageSize,
		"totalPages":  totalPages,
		"hasMore":     req.Page < totalPages,
	}, c)
}

// MVPCancelAnalysis MVP版本：取消分析任务
func (ra *RequirementAnalysisApi) MVPCancelAnalysis(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.CancelAnalysis(uint(taskID))
	if err != nil {
		global.GVA_LOG.Error("取消分析任务失败!", zap.Error(err))
		response.FailWithMessage("取消分析失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("分析任务已取消", c)
}

// MVPCancelAnalysis MVP版本：删除分析任务
func (ra *RequirementAnalysisApi) MVPDeleteAnalysisTask(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.DeleteAnalysisTask(uint(taskID))
	if err != nil {
		global.GVA_LOG.Error("删除分析任务失败!", zap.Error(err))
		response.FailWithMessage("删除分析任务失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("分析任务已删除", c)
}

// MVPGetTaskStatistics MVP版本：获取任务统计信息
func (ra *RequirementAnalysisApi) MVPGetTaskStatistics(c *gin.Context) {
	var req nesmaReq.TaskStatisticsRequest
	
	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		response.FailWithMessage("参数绑定失败: "+err.Error(), c)
		return
	}
	
	var projectId, cycleId uint
	if req.ProjectID != nil {
		projectId = *req.ProjectID
	}
	if req.CycleID != nil {
		cycleId = *req.CycleID
	}

	stats, err := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService.GetTaskStatistics(projectId, cycleId)
	if err != nil {
		global.GVA_LOG.Error("获取任务统计失败!", zap.Error(err))
		response.FailWithMessage("获取任务统计失败: "+err.Error(), c)
		return
	}

	response.OkWithData(stats, c)
}