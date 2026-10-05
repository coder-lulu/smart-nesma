package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UnifiedAnalysisApi struct{}

// ExecuteProjectAnalysis 执行项目分析 - 简化版本
// @Tags UnifiedAnalysis
// @Summary 执行项目分析
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body request.UnifiedAnalysisRequest true "项目分析请求"
// @Success 200 {object} response.Response{data=object} "成功"
// @Router /nesma/unified-analysis/project-analysis [post]
func (u *UnifiedAnalysisApi) ExecuteProjectAnalysis(c *gin.Context) {
	var req request.UnifiedAnalysisRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 获取简化的统一分析服务
	analysisService := service.ServiceGroupApp.NesmaServiceGroup.UnifiedAnalysisService
	
	// 执行项目分析
	result, err := analysisService.ExecuteProjectAnalysis(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("执行项目分析失败", zap.Error(err))
		response.FailWithMessage("执行项目分析失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// GetAnalysisProgress 获取分析进度
// @Tags UnifiedAnalysis
// @Summary 获取分析进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param task_id query uint true "任务ID"
// @Success 200 {object} response.Response{data=object} "成功"
// @Router /nesma/unified-analysis/progress [get]
func (u *UnifiedAnalysisApi) GetAnalysisProgress(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的任务ID", c)
		return
	}

	// 获取统一分析服务
	analysisService := service.ServiceGroupApp.NesmaServiceGroup.UnifiedAnalysisService
	
	// 获取进度
	progress, err := analysisService.GetAnalysisProgress(uint(taskID))
	if err != nil {
		global.GVA_LOG.Error("获取分析进度失败", zap.Error(err))
		response.FailWithMessage("获取进度失败: "+err.Error(), c)
		return
	}

	response.OkWithData(progress, c)
}

// GetDetailedAnalysisResult 获取详细分析结果
// @Tags UnifiedAnalysis
// @Summary 获取详细分析结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param task_id query uint true "任务ID"
// @Param detail_level query string false "详细级别：summary/detailed" default("summary")
// @Success 200 {object} response.Response{data=object} "成功"
// @Router /nesma/unified-analysis/detailed-result [get]
func (u *UnifiedAnalysisApi) GetDetailedAnalysisResult(c *gin.Context) {
	taskIDStr := c.Query("task_id")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的任务ID", c)
		return
	}

	detailLevel := c.DefaultQuery("detail_level", "summary")

	// 获取统一分析服务
	analysisService := service.ServiceGroupApp.NesmaServiceGroup.UnifiedAnalysisService
	
	// 获取详细结果
	result, err := analysisService.GetDetailedAnalysisResult(uint(taskID), detailLevel)
	if err != nil {
		global.GVA_LOG.Error("获取详细分析结果失败", 
			zap.Uint("taskId", uint(taskID)),
			zap.String("detailLevel", detailLevel),
			zap.Error(err))
		response.FailWithMessage("获取分析结果失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}