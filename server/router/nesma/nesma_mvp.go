package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type NesmaProjectCycleRouter struct{}

// InitNesmaProjectCycleRouter 初始化项目周期路由
func (s *NesmaProjectCycleRouter) InitNesmaProjectCycleRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	cycleRouter := Router.Group("nesma/project-cycle").Use(middleware.OperationRecord())
	cycleRouterWithoutRecord := Router.Group("nesma/project-cycle")

	var cycleApi = v1.ApiGroupApp.NesmaApiGroup.NesmaProjectCycleApi
	{
		cycleRouter.POST("", cycleApi.CreateNesmaProjectCycle)      // 创建项目周期
		cycleRouter.PUT("", cycleApi.UpdateNesmaProjectCycle)       // 更新项目周期
		cycleRouter.DELETE(":id", cycleApi.DeleteNesmaProjectCycle) // 删除项目周期
		cycleRouter.PUT("status", cycleApi.UpdateCycleStatus)       // 更新周期状态
		cycleRouter.POST("set-active", cycleApi.SetActiveProjectCycle) // 设置激活周期
	}
	{
		cycleRouterWithoutRecord.GET(":id", cycleApi.GetNesmaProjectCycle)     // 根据ID获取周期
		cycleRouterWithoutRecord.GET("list", cycleApi.GetNesmaProjectCycleList) // 获取周期列表
		cycleRouterWithoutRecord.GET("project/:projectId", cycleApi.GetProjectCycles) // 获取项目的所有周期
	}
}

type RequirementAnalysisRouter struct{}

// InitRequirementAnalysisRouter 初始化需求分析路由
func (s *RequirementAnalysisRouter) InitRequirementAnalysisRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	analysisRouter := Router.Group("nesma/analysis").Use(middleware.OperationRecord())
	analysisRouterWithoutRecord := Router.Group("nesma/analysis")

	var analysisApi = v1.ApiGroupApp.NesmaApiGroup.RequirementAnalysisApi
	{
		// MVP一键分析功能
		analysisRouter.POST("analyze", analysisApi.MVPStartAnalysis)      // 启动一键分析
		analysisRouter.POST("cancel/:taskId", analysisApi.MVPCancelAnalysis)    // 取消分析任务
	}
	{
		// 查询功能
		analysisRouterWithoutRecord.GET("progress/:taskId", analysisApi.MVPGetAnalysisProgress) // 获取分析进度
		analysisRouterWithoutRecord.GET("tasks", analysisApi.MVPGetAnalysisTasks)               // 获取分析任务列表
		analysisRouterWithoutRecord.GET("tasks/statistics", analysisApi.MVPGetTaskStatistics)  // 获取任务统计信息
		analysisRouterWithoutRecord.DELETE("tasks/:taskId", analysisApi.MVPDeleteAnalysisTask)       // 删除分析任务
	}
}

type IntelligentAnalysisRouter struct{}

// InitIntelligentAnalysisRouter 初始化智能分析路由
func (s *IntelligentAnalysisRouter) InitIntelligentAnalysisRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	intelligentAnalysisRouter := Router.Group("nesma/intelligent-analysis").Use(middleware.OperationRecord())
	intelligentAnalysisRouterWithoutRecord := Router.Group("nesma/intelligent-analysis")
	
	var intelligentAnalysisApi = v1.ApiGroupApp.NesmaApiGroup.IntelligentAnalysisApi
	
	{
		// 需要记录操作日志的接口
		intelligentAnalysisRouter.POST("start", intelligentAnalysisApi.StartIntelligentAnalysis)           // 启动智能分析
		intelligentAnalysisRouter.POST("apply-recommendation", intelligentAnalysisApi.ApplyAnalysisRecommendation) // 应用分析建议
		intelligentAnalysisRouter.POST("ai/test", intelligentAnalysisApi.TestAIModel)                         // 测试AI模型
		intelligentAnalysisRouter.POST("knowledge/search", intelligentAnalysisApi.SearchKnowledgeBase)        // 搜索知识库
		intelligentAnalysisRouter.POST("vector/search", intelligentAnalysisApi.VectorSearch)                  // 向量搜索
	}
	
	{
		// 不需要记录操作日志的接口
		intelligentAnalysisRouterWithoutRecord.GET("requirement/tree-enhanced", intelligentAnalysisApi.GetRequirementTree)     // 获取需求树
		intelligentAnalysisRouterWithoutRecord.GET("progress/:taskId", intelligentAnalysisApi.GetAnalysisProgress)             // 获取分析进度
		intelligentAnalysisRouterWithoutRecord.GET("result/:taskId", intelligentAnalysisApi.GetAnalysisResult)                 // 获取分析结果
		intelligentAnalysisRouterWithoutRecord.GET("recommendations/:taskId", intelligentAnalysisApi.GetAnalysisRecommendations) // 获取分析建议
		intelligentAnalysisRouterWithoutRecord.GET("ai/status", intelligentAnalysisApi.GetAIServiceStatus)                     // 获取AI服务状态
		intelligentAnalysisRouterWithoutRecord.GET("ai/models", intelligentAnalysisApi.GetAvailableAIModels)                   // 获取可用AI模型
		intelligentAnalysisRouterWithoutRecord.GET("knowledge/graph", intelligentAnalysisApi.GetKnowledgeGraph)                // 获取知识图谱
		intelligentAnalysisRouterWithoutRecord.GET("knowledge-entities", intelligentAnalysisApi.GetKnowledgeEntities)          // 获取知识实体
		intelligentAnalysisRouterWithoutRecord.GET("knowledge-relations", intelligentAnalysisApi.GetKnowledgeRelations)        // 获取知识关系
		intelligentAnalysisRouterWithoutRecord.GET("export", intelligentAnalysisApi.ExportIntelligentAnalysisReport)           // 导出分析报告
	}
}

type ConcurrentConfigRouter struct{}

// InitConcurrentConfigRouter 初始化并发配置路由
func (s *ConcurrentConfigRouter) InitConcurrentConfigRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	configRouter := Router.Group("nesma/concurrent-config").Use(middleware.OperationRecord())
	configRouterWithoutRecord := Router.Group("nesma/concurrent-config")

	var configApi = v1.ApiGroupApp.NesmaApiGroup.ConcurrentConfigApi
	{
		configRouter.PUT("", configApi.UpdateConcurrentConfig)     // 更新并发配置
		configRouter.POST("reset", configApi.ResetConcurrentConfig) // 重置并发配置
	}
	{
		configRouterWithoutRecord.GET("", configApi.GetConcurrentConfig) // 获取并发配置
	}
}

type PerformanceMonitorRouter struct{}

// InitPerformanceMonitorRouter 初始化性能监控路由
func (s *PerformanceMonitorRouter) InitPerformanceMonitorRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	monitorRouter := Router.Group("nesma/performance").Use(middleware.OperationRecord())
	monitorRouterWithoutRecord := Router.Group("nesma/performance")

	var monitorApi = v1.ApiGroupApp.NesmaApiGroup.PerformanceMonitorApi
	{
		monitorRouter.PUT("config", monitorApi.UpdateConcurrentConfig) // 更新并发配置
	}
	{
		monitorRouterWithoutRecord.GET("metrics/:taskId", monitorApi.GetTaskMetrics)        // 获取任务性能指标
		monitorRouterWithoutRecord.GET("summary", monitorApi.GetPerformanceSummary)          // 获取性能摘要
		monitorRouterWithoutRecord.GET("metrics", monitorApi.GetAllMetrics)                  // 获取所有任务指标
		monitorRouterWithoutRecord.GET("config", monitorApi.GetConcurrentConfig)            // 获取并发配置
	}
}

type ReportGeneratorRouter struct{}

// InitReportGeneratorRouter 初始化报告生成路由
func (s *ReportGeneratorRouter) InitReportGeneratorRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	reportRouter := Router.Group("nesma/report").Use(middleware.OperationRecord())
	reportRouterWithoutRecord := Router.Group("nesma/report")

	// var unifiedAnalysisApi = v1.ApiGroupApp.NesmaApiGroup.UnifiedAnalysisApi // 暂时注释未使用
	var documentApi = v1.ApiGroupApp.NesmaApiGroup.DocumentApi
	var templateApi = v1.ApiGroupApp.NesmaApiGroup.TemplateApi
	{
		// 报告生成相关接口（需要操作日志）
		// reportRouter.POST("generate", unifiedAnalysisApi.ExportAnalysisReport) // 生成报告 - 暂时注释
		reportRouter.POST("batch-generate", documentApi.BatchGenerateDocument) // 批量生成报告
	}
	{
		// 报告模板和查询相关接口（不需要操作日志）
		reportRouterWithoutRecord.GET("templates", templateApi.GetTemplateList)  // 获取报告模板列表
		reportRouterWithoutRecord.GET("progress/:id", documentApi.GetDocumentProgress) // 获取报告生成进度
		reportRouterWithoutRecord.GET("download/:id", documentApi.DownloadDocument) // 下载报告
		reportRouterWithoutRecord.GET("stats", documentApi.GetDocumentStats)    // 获取报告统计
	}
}