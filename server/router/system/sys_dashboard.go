package system

import (
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DashboardRouter struct{}

// InitDashboardRouter 初始化 Dashboard 路由信息
func (s *DashboardRouter) InitDashboardRouter(Router *gin.RouterGroup) {
	dashboardRouter := Router.Group("dashboard").Use(middleware.OperationRecord())
	dashboardRouterWithoutRecord := Router.Group("dashboard")
	{
		dashboardRouterWithoutRecord.GET("stats", dashboardApi.GetDashboardStats)       // 获取仪表板统计信息
		dashboardRouterWithoutRecord.GET("activities", dashboardApi.GetDashboardActivities) // 获取仪表板活动记录
	}
	{
		// 需要记录操作日志的接口（如果有的话）
		_ = dashboardRouter
	}
}