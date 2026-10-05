package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type UnifiedAnalysisRouter struct{}

// InitUnifiedAnalysisRouter 初始化统一分析路由信息 - 简化版本
func (s *UnifiedAnalysisRouter) InitUnifiedAnalysisRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	unifiedAnalysisRouter := Router.Group("nesma/unified-analysis").Use(middleware.OperationRecord())
	unifiedAnalysisRouterWithoutRecord := Router.Group("nesma/unified-analysis")
	unifiedAnalysisApi := v1.ApiGroupApp.NesmaApiGroup.UnifiedAnalysisApi

	{
		// 需要操作记录的接口
		unifiedAnalysisRouter.POST("project-analysis", unifiedAnalysisApi.ExecuteProjectAnalysis) // 项目一键分析
	}
	{
		// 不需要操作记录的查询接口
		unifiedAnalysisRouterWithoutRecord.GET("progress", unifiedAnalysisApi.GetAnalysisProgress)             // 获取分析进度
		unifiedAnalysisRouterWithoutRecord.GET("detailed-result", unifiedAnalysisApi.GetDetailedAnalysisResult) // 获取详细分析结果
	}
}