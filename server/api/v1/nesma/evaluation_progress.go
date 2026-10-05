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

type EvaluationProgressApi struct{}

// @Tags NESMA评估进度
// @Summary 获取评估进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param evaluation_id path int true "评估ID"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationProgressResponse} "获取成功"
// @Router /nesma/evaluation/{evaluation_id}/progress [get]
func (e *EvaluationProgressApi) GetEvaluationProgress(c *gin.Context) {
	evaluationIdStr := c.Param("id")
	evaluationId, err := strconv.ParseUint(evaluationIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesmaService.GetEvaluationProgressService()
	progressData, err := evaluationService.GetEvaluationProgress(uint(evaluationId))
	if err != nil {
		global.GVA_LOG.Error("获取评估进度失败", zap.Error(err))
		response.FailWithMessage("获取评估进度失败: "+err.Error(), c)
		return
	}

	response.OkWithData(progressData, c)
}

// @Tags NESMA评估进度
// @Summary 启动评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.StartEvaluationProgressRequest true "启动评估请求"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationProgressResponse} "启动成功"
// @Router /nesma/evaluation/start [post]
func (e *EvaluationProgressApi) StartEvaluation(c *gin.Context) {
	var req nesmaReq.StartEvaluationProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesmaService.GetEvaluationProgressService()
	progressData, err := evaluationService.StartEvaluation(req.EvaluationID, req.Config)
	if err != nil {
		global.GVA_LOG.Error("启动评估失败", zap.Error(err))
		response.FailWithMessage("启动评估失败: "+err.Error(), c)
		return
	}

	response.OkWithData(progressData, c)
}

// @Tags NESMA评估进度
// @Summary 取消评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CancelEvaluationRequest true "取消评估请求"
// @Success 200 {object} response.Response{} "取消成功"
// @Router /nesma/evaluation/cancel [post]
func (e *EvaluationProgressApi) CancelEvaluation(c *gin.Context) {
	var req nesmaReq.CancelEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesmaService.GetEvaluationProgressService()
	err := evaluationService.CancelEvaluation(req.EvaluationID, req.Reason)
	if err != nil {
		global.GVA_LOG.Error("取消评估失败", zap.Error(err))
		response.FailWithMessage("取消评估失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估已取消", c)
}

// @Tags NESMA评估进度
// @Summary 获取评估状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param evaluation_id path int true "评估ID"
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationStatusResponse} "获取成功"
// @Router /nesma/evaluation/{evaluation_id}/status [get]
func (e *EvaluationProgressApi) GetEvaluationStatus(c *gin.Context) {
	evaluationIdStr := c.Param("id")
	evaluationId, err := strconv.ParseUint(evaluationIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	evaluationService := nesmaService.GetEvaluationProgressService()
	statusData, err := evaluationService.GetEvaluationStatus(uint(evaluationId))
	if err != nil {
		global.GVA_LOG.Error("获取评估状态失败", zap.Error(err))
		response.FailWithMessage("获取评估状态失败: "+err.Error(), c)
		return
	}

	response.OkWithData(statusData, c)
}

// @Tags NESMA评估进度
// @Summary 获取评估统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationSystemStatsResponse} "获取成功"
// @Router /nesma/evaluation/stats [get]
func (e *EvaluationProgressApi) GetEvaluationStats(c *gin.Context) {
	evaluationService := nesmaService.GetEvaluationProgressService()
	statsData, err := evaluationService.GetEvaluationStats()
	if err != nil {
		global.GVA_LOG.Error("获取评估统计失败", zap.Error(err))
		response.FailWithMessage("获取评估统计失败: "+err.Error(), c)
		return
	}

	response.OkWithData(statsData, c)
}

// @Tags NESMA评估进度
// @Summary 获取评估日志
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param evaluation_id path int true "评估ID"
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Param level query string false "日志级别" Enums(info,warning,error)
// @Success 200 {object} response.Response{data=nesmaRes.EvaluationLogResponse} "获取成功"
// @Router /nesma/evaluation/{evaluation_id}/logs [get]
func (e *EvaluationProgressApi) GetEvaluationLogs(c *gin.Context) {
	evaluationIdStr := c.Param("id")
	evaluationId, err := strconv.ParseUint(evaluationIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的评估ID", c)
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	level := c.Query("level")

	evaluationService := nesmaService.GetEvaluationProgressService()
	logsData, err := evaluationService.GetEvaluationLogs(uint(evaluationId), page, pageSize, level)
	if err != nil {
		global.GVA_LOG.Error("获取评估日志失败", zap.Error(err))
		response.FailWithMessage("获取评估日志失败: "+err.Error(), c)
		return
	}

	response.OkWithData(logsData, c)
}

// @Tags NESMA评估进度
// @Summary 手动更新评估进度(内部接口)
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.UpdateEvaluationProgressRequest true "更新进度请求"
// @Success 200 {object} response.Response{} "更新成功"
// @Router /nesma/evaluation/progress/update [post]
func (e *EvaluationProgressApi) UpdateEvaluationProgress(c *gin.Context) {
	var req nesmaReq.UpdateEvaluationProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesmaService.GetEvaluationProgressService()
	err := evaluationService.UpdateProgress(req.EvaluationID, req.Progress, req.CurrentPhase, req.Message)
	if err != nil {
		global.GVA_LOG.Error("更新评估进度失败", zap.Error(err))
		response.FailWithMessage("更新评估进度失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("进度更新成功", c)
}

// @Tags NESMA评估进度
// @Summary 获取运行中的评估列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=[]nesmaRes.RunningEvaluationResponse} "获取成功"
// @Router /nesma/evaluation/running [get]
func (e *EvaluationProgressApi) GetRunningEvaluations(c *gin.Context) {
	evaluationService := nesmaService.GetEvaluationProgressService()
	runningEvaluations, err := evaluationService.GetRunningEvaluations()
	if err != nil {
		global.GVA_LOG.Error("获取运行中评估列表失败", zap.Error(err))
		response.FailWithMessage("获取运行中评估列表失败: "+err.Error(), c)
		return
	}

	response.OkWithData(runningEvaluations, c)
}

// @Tags NESMA评估进度
// @Summary 强制停止评估
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.ForceStopEvaluationRequest true "强制停止评估请求"
// @Success 200 {object} response.Response{} "停止成功"
// @Router /nesma/evaluation/force-stop [post]
func (e *EvaluationProgressApi) ForceStopEvaluation(c *gin.Context) {
	var req nesmaReq.ForceStopEvaluationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	evaluationService := nesmaService.GetEvaluationProgressService()
	err := evaluationService.ForceStopEvaluation(req.EvaluationID, req.Reason)
	if err != nil {
		global.GVA_LOG.Error("强制停止评估失败", zap.Error(err))
		response.FailWithMessage("强制停止评估失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("评估已强制停止", c)
}