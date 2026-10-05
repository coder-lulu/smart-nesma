package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

type PerformanceMonitorApi struct{}

// GetTaskMetrics 获取任务性能指标
func (api *PerformanceMonitorApi) GetTaskMetrics(c *gin.Context) {
	taskIDStr := c.Param("taskId")
	taskID, err := strconv.ParseUint(taskIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("任务ID格式错误", c)
		return
	}

	metrics := nesmaService.GetTaskMetrics(uint(taskID))
	if metrics == nil {
		response.FailWithMessage("任务指标不存在", c)
		return
	}

	response.OkWithData(metrics, c)
}

// GetPerformanceSummary 获取性能摘要
func (api *PerformanceMonitorApi) GetPerformanceSummary(c *gin.Context) {
	summary := nesmaService.GetPerformanceSummary()
	response.OkWithData(summary, c)
}

// GetAllMetrics 获取所有任务指标
func (api *PerformanceMonitorApi) GetAllMetrics(c *gin.Context) {
	// 这个需要通过全局监控器获取
	response.OkWithMessage("功能正在重构中", c)
}

// GetConcurrentConfig 获取并发配置
func (api *PerformanceMonitorApi) GetConcurrentConfig(c *gin.Context) {
	config := nesmaService.GetConcurrentConfig()
	response.OkWithData(config, c)
}

// UpdateConcurrentConfig 更新并发配置
func (api *PerformanceMonitorApi) UpdateConcurrentConfig(c *gin.Context) {
	var config nesmaService.ConcurrentConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		global.GVA_LOG.Error("更新并发配置参数绑定失败!", zap.Error(err))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 验证配置
	if err := nesmaService.ValidateConcurrentConfig(config); err != nil {
		global.GVA_LOG.Error("并发配置验证失败!", zap.Error(err))
		response.FailWithMessage("配置验证失败: "+err.Error(), c)
		return
	}

	// 更新配置
	nesmaService.UpdateConcurrentConfig(config)
	
	global.GVA_LOG.Info("并发配置更新成功", 
		zap.Int("maxAnalysisWorkers", config.MaxAnalysisWorkers),
		zap.Int("maxOptimizationWorkers", config.MaxOptimizationWorkers),
		zap.Int("maxEvaluationWorkers", config.MaxEvaluationWorkers))
	
	response.OkWithMessage("并发配置更新成功", c)
} 