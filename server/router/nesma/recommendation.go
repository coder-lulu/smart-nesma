package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type RecommendationRouter struct{}

// InitRecommendationRouter 初始化推荐系统路由
func (r *RecommendationRouter) InitRecommendationRouter(Router *gin.RouterGroup) {
	recommendationRouter := Router.Group("nesma/recommendation").Use(middleware.OperationRecord())
	recommendationRouterWithoutRecord := Router.Group("nesma/recommendation")
	var recommendationApi = v1.ApiGroupApp.NesmaApiGroup.RecommendationApi
	var aiRecommendationApi = v1.ApiGroupApp.NesmaApiGroup.AIRecommendationApi
	
	{
		// 传统推荐系统路由（保留兼容性）
		recommendationRouter.POST("behavior", recommendationApi.RecordUserBehavior)                 // 记录用户行为
		recommendationRouter.POST("feedback", recommendationApi.RecordRecommendationFeedback)       // 记录推荐反馈
		recommendationRouter.POST("personalized", recommendationApi.GetPersonalizedRecommendations) // 个性化推荐
		
		// AI智能推荐路由（新功能）
		recommendationRouter.POST("ai", aiRecommendationApi.GetAIRecommendations)           // 获取AI推荐
		recommendationRouter.POST("adopt", aiRecommendationApi.AdoptRecommendation)         // 采纳推荐
		recommendationRouter.POST("reject", aiRecommendationApi.RejectRecommendation)       // 拒绝推荐
		recommendationRouter.POST("batch", aiRecommendationApi.GetBatchRecommendations)     // 批量推荐
	}
	{
		// 不需要记录操作的路由（查询类接口）
		recommendationRouterWithoutRecord.GET("items", recommendationApi.GetRecommendations)       // 获取推荐列表
		recommendationRouterWithoutRecord.GET("stats", recommendationApi.GetRecommendationStats)   // 获取推荐统计
		recommendationRouterWithoutRecord.GET("config", recommendationApi.GetRecommendationConfig) // 获取推荐配置
		
		// AI推荐查询路由
		recommendationRouterWithoutRecord.GET("quick", aiRecommendationApi.GetQuickRecommendations)     // 快速推荐
		recommendationRouterWithoutRecord.GET("adoption-history", aiRecommendationApi.GetAdoptionHistory) // 采纳历史
		recommendationRouterWithoutRecord.GET("ai-config", aiRecommendationApi.GetAIRecommendationConfig) // AI推荐配置
	}
}
