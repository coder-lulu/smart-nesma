package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type MonitoringRouter struct{}

// InitMonitoringRouter 初始化监控路由
func (m *MonitoringRouter) InitMonitoringRouter(Router *gin.RouterGroup) {
	monitoringRouter := Router.Group("monitoring").Use(middleware.OperationRecord())
	monitoringRouterWithoutRecord := Router.Group("monitoring")
	var monitoringApi = v1.ApiGroupApp.NesmaApiGroup.MonitoringApi

	{
		// 服务性能统计
		monitoringRouter.GET("service/stats", monitoringApi.GetServiceStats)

		// 错误分析
		monitoringRouter.GET("error/analysis", monitoringApi.GetErrorAnalysis)

		// Token使用统计
		monitoringRouter.GET("token/stats", monitoringApi.GetTokenUsageStats)

		// 时间序列统计
		monitoringRouter.GET("time-series", monitoringApi.GetTimeSeriesStats)

		// 监控指标
		monitoringRouter.GET("metrics", monitoringApi.GetMonitoringMetrics)

		// 需求统计
		monitoringRouter.GET("requirement-stats", monitoringApi.GetRequirementStats)

		// 服务健康状态
		monitoringRouterWithoutRecord.GET("health", monitoringApi.GetServiceHealth)

		// 系统资源使用情况
		monitoringRouterWithoutRecord.GET("system/resources", monitoringApi.GetSystemResources)

		// 模型性能对比
		monitoringRouter.GET("model/comparison", monitoringApi.GetModelPerformanceComparison)

		// 响应时间分布
		monitoringRouter.GET("response-time/distribution", monitoringApi.GetResponseTimeDistribution)

		// 成本分析
		monitoringRouter.GET("cost/analysis", monitoringApi.GetCostAnalysis)

		// 导出监控报告
		monitoringRouter.GET("report/export", monitoringApi.ExportMonitoringReport)
	}
}
