package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AIAnalysisRouter struct{}

// InitAIAnalysisRouter 初始化AI分析路由
func (r *AIAnalysisRouter) InitAIAnalysisRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	aiAnalysisRouter := Router.Group("ai-analysis").Use(middleware.OperationRecord())
	aiAnalysisRouterWithoutRecord := Router.Group("ai-analysis")
	aiAnalysisPublicRouter := PublicRouter.Group("ai-analysis")

	aiAnalysisApi := v1.ApiGroupApp.NesmaApiGroup.AIAnalysisApi

	{
		// 需要记录操作的路由
		aiAnalysisRouter.POST("start", aiAnalysisApi.StartAIAnalysis)                               // 启动AI分析
		aiAnalysisRouter.DELETE(":analysisId", aiAnalysisApi.DeleteAnalysis)                       // 删除分析记录
	}
	{
		// 不需要记录操作的路由
		aiAnalysisRouterWithoutRecord.GET(":analysisId/status", aiAnalysisApi.GetAnalysisStatus)   // 获取分析状态
		aiAnalysisRouterWithoutRecord.GET(":analysisId/result", aiAnalysisApi.GetAnalysisResult)   // 获取分析结果
		aiAnalysisRouterWithoutRecord.GET(":analysisId/report", aiAnalysisApi.GetAnalysisReport)   // 生成分析报告
		aiAnalysisRouterWithoutRecord.GET("project/:projectId/history", aiAnalysisApi.GetAnalysisHistory) // 获取项目分析历史
	}
	{
		// 公开路由（如果需要的话，目前AI分析功能都需要登录）
		_ = aiAnalysisPublicRouter
	}
}