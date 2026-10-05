package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/router"
	"github.com/gin-gonic/gin"
)

// 占位方法，保证文件可以正确加载，避免go空变量检测报错，请勿删除。
func holder(routers ...*gin.RouterGroup) {
	_ = routers
	_ = router.RouterGroupApp
}

func initBizRouter(routers ...*gin.RouterGroup) {
	privateGroup := routers[0]
	publicGroup := routers[1]

	nesmaRouter := router.RouterGroupApp.Nesma
	{
		nesmaRouter.InitNesmaProjectRouter(privateGroup, publicGroup)     // NESMA项目路由
		nesmaRouter.InitNesmaProjectCycleRouter(privateGroup, publicGroup) // NESMA项目周期路由
		nesmaRouter.InitNesmaRequirementRouter(privateGroup, publicGroup) // NESMA需求路由
		nesmaRouter.InitNesmaRequirementVersionRouter(privateGroup, publicGroup) // NESMA需求版本路由
		nesmaRouter.InitRequirementAnalysisRouter(privateGroup, publicGroup) // MVP一键分析路由
		nesmaRouter.InitIntelligentAnalysisRouter(privateGroup, publicGroup) // 智能分析增强路由
		nesmaRouter.InitKnowledgeRouter(privateGroup, publicGroup)        // 知识库管理路由
		nesmaRouter.InitAgentRouter(privateGroup)                         // Agent管理路由
		nesmaRouter.InitDocumentRouter(privateGroup, publicGroup)         // 文档生成路由
		nesmaRouter.InitAIServiceRouter(privateGroup)                     // AI服务管理路由
		nesmaRouter.InitRecommendationRouter(privateGroup)                // 智能推荐路由
		nesmaRouter.InitChatRouter(privateGroup, publicGroup)             // AI聊天路由
		nesmaRouter.InitEvaluationRouter(privateGroup, publicGroup)      // NESMA评估路由
		nesmaRouter.InitMonitoringRouter(privateGroup)                    // 监控管理路由
		nesmaRouter.InitLevel3AnalysisRouter(privateGroup, publicGroup)   // 三级功能点分析路由
		nesmaRouter.InitLevel4GeneratorRouter(privateGroup, publicGroup)  // 四级功能点生成路由
		nesmaRouter.InitDescriptionGeneratorRouter(privateGroup, publicGroup) // 功能点描述生成路由
		nesmaRouter.InitMermaidGeneratorRouter(privateGroup, publicGroup) // Mermaid流程图生成路由
		nesmaRouter.InitDocumentExportRouter(privateGroup, publicGroup)   // 文档导出路由
		nesmaRouter.InitUnifiedAnalysisWorkflowRouter(privateGroup, publicGroup) // 统一智能分析工作流路由
		nesmaRouter.InitUnifiedAnalysisRouter(privateGroup, publicGroup) // 统一分析路由
		nesmaRouter.InitReportGeneratorRouter(privateGroup, publicGroup) // 报告生成器路由
		nesmaRouter.InitBatchEvaluationRouter(privateGroup, publicGroup) // 批量评估路由
		nesmaRouter.InitErrorHandlingRouter(privateGroup, publicGroup)   // 错误处理路由
		nesmaRouter.InitAIAnalysisRouter(privateGroup, publicGroup)      // AI项目分析路由
	}

	holder(publicGroup, privateGroup)
}
