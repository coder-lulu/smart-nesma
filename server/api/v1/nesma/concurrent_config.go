package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ConcurrentConfigApi struct{}

// GetConcurrentConfig 获取当前并发配置
func (api *ConcurrentConfigApi) GetConcurrentConfig(c *gin.Context) {
	config := nesmaService.GetConcurrentConfig()
	response.OkWithData(config, c)
}

// UpdateConcurrentConfig 更新并发配置
func (api *ConcurrentConfigApi) UpdateConcurrentConfig(c *gin.Context) {
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

// ResetConcurrentConfig 重置为默认并发配置
func (api *ConcurrentConfigApi) ResetConcurrentConfig(c *gin.Context) {
	defaultConfig := nesmaService.DefaultConcurrentConfig
	nesmaService.UpdateConcurrentConfig(defaultConfig)
	
	global.GVA_LOG.Info("并发配置已重置为默认值")
	response.OkWithMessage("并发配置已重置为默认值", c)
} 