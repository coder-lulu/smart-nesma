package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AIServiceRouter struct{}

// InitAIServiceRouter 初始化AI服务路由
func (r *AIServiceRouter) InitAIServiceRouter(Router *gin.RouterGroup) {
	aiServiceRouter := Router.Group("ai").Use(middleware.OperationRecord())
	aiServiceRouterWithoutRecord := Router.Group("ai")
	var aiServiceApi = v1.ApiGroupApp.NesmaApiGroup.AIServiceApi
	{
		// 需要记录操作的路由
		aiServiceRouter.POST("services/switch", aiServiceApi.SwitchDefaultAIService)             // 切换默认AI服务
		aiServiceRouter.POST("services/test", aiServiceApi.TestAIService)                        // 测试AI服务
		aiServiceRouter.POST("services/circuit-breaker/reset", aiServiceApi.ResetCircuitBreaker) // 重置熔断器
	}
	{
		// 不需要记录操作的路由（查询类接口）
		aiServiceRouterWithoutRecord.GET("services", aiServiceApi.GetAIServices)                                // 获取AI服务列表
		aiServiceRouterWithoutRecord.GET("services/health", aiServiceApi.GetAIServiceHealth)                    // 获取AI服务健康状态
		aiServiceRouterWithoutRecord.GET("services/config", aiServiceApi.GetAIServiceConfig)                    // 获取AI服务配置
		aiServiceRouterWithoutRecord.GET("services/circuit-breaker/stats", aiServiceApi.GetCircuitBreakerStats) // 获取熔断器统计
	}
}
