package nesma

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type UnifiedAnalysisWorkflowRouter struct {}

// InitUnifiedAnalysisWorkflowRouter 初始化统一智能分析工作流路由
func (u *UnifiedAnalysisWorkflowRouter) InitUnifiedAnalysisWorkflowRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	workflowRouter := Router.Group("workflow")
	workflowPublicRouter := PublicRouter.Group("workflow")
	
	workflowApi := v1.ApiGroupApp.NesmaApiGroup.UnifiedAnalysisWorkflowApi
	
	// 私有路由（需要认证）
	{
		workflowRouter.POST("execute", workflowApi.ExecuteFullAnalysisWorkflow)           // 执行完整分析工作流
		workflowRouter.GET("status/:executionId", workflowApi.GetWorkflowExecutionStatus) // 获取工作流执行状态
		workflowRouter.POST("cancel/:executionId", workflowApi.CancelWorkflowExecution)   // 取消工作流执行
		workflowRouter.GET("history/:projectId", workflowApi.GetWorkflowExecutionHistory) // 获取工作流执行历史
		workflowRouter.GET("stats/:projectId", workflowApi.GetWorkflowExecutionStats)     // 获取工作流执行统计
		workflowRouter.POST("validate", workflowApi.ValidateWorkflowConfiguration)        // 验证工作流配置
		workflowRouter.GET("templates", workflowApi.GetWorkflowTemplates)                 // 获取工作流模板
		workflowRouter.POST("create-from-template", workflowApi.CreateWorkflowFromTemplate) // 从模板创建工作流
	}
	
	// 公共路由（如果需要）
	{
		// 暂无公共路由
		_ = workflowPublicRouter
	}
}