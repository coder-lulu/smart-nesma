package nesma

import (
	nesmaApi "github.com/flipped-aurora/gin-vue-admin/server/api/v1/nesma"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
)

// ErrorHandlingRouter 错误处理路由
type ErrorHandlingRouter struct{}

// InitErrorHandlingRouter 初始化错误处理路由
func (r *ErrorHandlingRouter) InitErrorHandlingRouter(privateGroup, publicGroup *gin.RouterGroup) {
	// 获取批量评估服务实例
	batchEvaluationService := nesmaService.GetBatchEvaluationService()
	if batchEvaluationService == nil {
		// 如果服务未初始化，直接返回
		return
	}
	
	errorHandlingApi := nesmaApi.NewErrorHandlingApi(batchEvaluationService)
	
	// 私有路由组（需要认证）
	errorHandlingRouterWithAuth := privateGroup.Group("nesma/error-handling")
	{
		// 错误指标相关
		errorHandlingRouterWithAuth.GET("metrics", errorHandlingApi.GetErrorMetrics)
		errorHandlingRouterWithAuth.POST("metrics/reset", errorHandlingApi.ResetErrorMetrics)
		errorHandlingRouterWithAuth.GET("trends", errorHandlingApi.GetErrorTrends)
		
		// 断路器相关
		errorHandlingRouterWithAuth.GET("circuit-breaker/status", errorHandlingApi.GetCircuitBreakerStatus)
		errorHandlingRouterWithAuth.POST("circuit-breaker/:name/reset", errorHandlingApi.ResetCircuitBreaker)
		errorHandlingRouterWithAuth.POST("circuit-breaker/reset-all", errorHandlingApi.ResetAllCircuitBreakers)
		
		// 测试相关
		errorHandlingRouterWithAuth.POST("test", errorHandlingApi.TestErrorHandling)
	}
	
	// 公共路由组（无需认证）- 只提供基本的状态查询
	errorHandlingRouterPublic := publicGroup.Group("nesma/error-handling")
	{
		// 基本状态查询
		errorHandlingRouterPublic.GET("status", errorHandlingApi.GetCircuitBreakerStatus)
	}
}