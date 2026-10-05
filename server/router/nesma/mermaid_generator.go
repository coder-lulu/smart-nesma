package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type MermaidGeneratorRouter struct{}

// InitMermaidGeneratorRouter 初始化Mermaid流程图生成路由
func (r *MermaidGeneratorRouter) InitMermaidGeneratorRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	mermaidGeneratorApi := v1.ApiGroupApp.NesmaApiGroup.MermaidGeneratorApi

	// 私有路由组（需要认证）
	privateGroup := Router.Group("nesma/generator/mermaid").Use(middleware.OperationRecord())
	{
		privateGroup.POST("/generate", mermaidGeneratorApi.GenerateMermaidDiagrams)         // 生成Mermaid流程图
		privateGroup.POST("/apply", mermaidGeneratorApi.ApplyMermaidDiagram)               // 应用Mermaid流程图
		privateGroup.POST("/batch-generate", mermaidGeneratorApi.BatchGenerateMermaidDiagrams) // 批量生成Mermaid流程图
		privateGroup.GET("/history/:cycleId", mermaidGeneratorApi.GetMermaidGenerationHistory) // 获取生成历史
		privateGroup.GET("/stats/:cycleId", mermaidGeneratorApi.GetMermaidGenerationStats)   // 获取生成统计
		privateGroup.POST("/validate", mermaidGeneratorApi.ValidateMermaidDiagram)          // 验证Mermaid流程图
		privateGroup.POST("/preview", mermaidGeneratorApi.PreviewMermaidDiagram)           // 预览Mermaid流程图
		
		// 异步流程图生成API
		privateGroup.POST("/generate-async", mermaidGeneratorApi.GenerateMermaidDiagramsAsync)  // 异步生成Mermaid流程图
		privateGroup.GET("/task/:taskId/progress", mermaidGeneratorApi.GetMermaidGenerationTaskProgress) // 获取任务进度
		privateGroup.GET("/task/:taskId/result", mermaidGeneratorApi.GetMermaidGenerationTaskResult)     // 获取任务结果
		
		// 任务管理API
		privateGroup.GET("/tasks", mermaidGeneratorApi.GetMermaidGenerationTasks)               // 获取任务列表
		privateGroup.GET("/task-statistics", mermaidGeneratorApi.GetMermaidTaskStatistics)      // 获取任务统计
		privateGroup.GET("/task/:taskId", mermaidGeneratorApi.GetMermaidTaskDetail)             // 获取任务详情
		privateGroup.POST("/task/:taskId/cancel", mermaidGeneratorApi.CancelMermaidGenerationTask) // 取消任务
		privateGroup.POST("/task/:taskId/retry", mermaidGeneratorApi.RetryMermaidGenerationTask)   // 重试任务
		privateGroup.DELETE("/task/:taskId", mermaidGeneratorApi.DeleteMermaidGenerationTask)      // 删除任务
		privateGroup.POST("/tasks/batch-delete", mermaidGeneratorApi.BatchDeleteMermaidGenerationTasks) // 批量删除任务
	}
}