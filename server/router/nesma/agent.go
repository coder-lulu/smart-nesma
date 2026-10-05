package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type AgentRouter struct{}

// InitAgentRouter 初始化Agent路由
func (ar *AgentRouter) InitAgentRouter(Router *gin.RouterGroup) {
	agentRouter := Router.Group("agent").Use(middleware.OperationRecord())
	agentRouterWithoutRecord := Router.Group("agent")
	agentApi := v1.ApiGroupApp.NesmaApiGroup.AgentApi
	{
		// Agent管理路由
		agentRouter.POST("register", agentApi.RegisterAgent)              // 注册Agent
		agentRouterWithoutRecord.GET(":id", agentApi.GetAgent)            // 获取Agent信息
		agentRouterWithoutRecord.GET("list", agentApi.ListAgents)         // 获取Agent列表
		agentRouter.PUT(":id/status", agentApi.UpdateAgentStatus)         // 更新Agent状态
		agentRouter.POST(":id/heartbeat", agentApi.AgentHeartbeat)        // Agent心跳
		agentRouterWithoutRecord.GET(":id/tasks", agentApi.GetAgentTasks) // 获取Agent任务列表
		agentRouterWithoutRecord.GET(":id/logs", agentApi.GetAgentLogs)   // 获取Agent日志

		// Agent任务路由
		agentRouter.POST("task", agentApi.CreateTask)                // 创建任务
		agentRouter.POST("task/:taskId/assign", agentApi.AssignTask) // 分配任务

		// Agent消息路由
		agentRouter.POST("message", agentApi.SendMessage) // 发送消息

		// 工作流管理路由
		agentRouter.POST("workflow", agentApi.CreateWorkflow)                        // 创建工作流
		agentRouterWithoutRecord.GET("workflow/:id", agentApi.GetWorkflow)           // 获取工作流
		agentRouterWithoutRecord.GET("workflows", agentApi.ListWorkflows)            // 获取工作流列表
		agentRouter.POST("workflow/:id/execute", agentApi.ExecuteWorkflow)           // 执行工作流
		agentRouterWithoutRecord.GET("execution/:id", agentApi.GetWorkflowExecution) // 获取工作流执行状态
		agentRouter.POST("execution/:id/cancel", agentApi.CancelWorkflowExecution)   // 取消工作流执行

		// 统计和监控路由
		agentRouterWithoutRecord.GET("statistics", agentApi.GetAgentStatistics)             // 获取Agent统计
		agentRouterWithoutRecord.GET("statistics/workflow", agentApi.GetWorkflowStatistics) // 获取工作流统计
		agentRouterWithoutRecord.GET("health", agentApi.GetAgentHealth)                     // 获取Agent健康状态
		agentRouterWithoutRecord.GET("metrics", agentApi.GetSystemMetrics)                  // 获取系统性能指标
	}
}
