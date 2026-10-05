package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type Level4GeneratorRouter struct{}

// InitLevel4GeneratorRouter 初始化四级功能点生成路由
func (r *Level4GeneratorRouter) InitLevel4GeneratorRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	level4GeneratorApi := v1.ApiGroupApp.NesmaApiGroup.Level4GeneratorApi

	// 私有路由组（需要认证）
	privateGroup := Router.Group("nesma/generator/level4").Use(middleware.OperationRecord())
	{
		privateGroup.POST("/generate", level4GeneratorApi.GenerateLevel4Requirements)      // 生成四级功能点
		privateGroup.POST("/create", level4GeneratorApi.CreateLevel4Requirement)          // 创建四级功能点
		privateGroup.POST("/batch-create", level4GeneratorApi.BatchCreateLevel4Requirements) // 批量创建四级功能点
		privateGroup.GET("/history/:cycleId", level4GeneratorApi.GetLevel4GenerationHistory) // 获取生成历史
		privateGroup.GET("/stats/:cycleId", level4GeneratorApi.GetLevel4GenerationStats)   // 获取生成统计
		privateGroup.POST("/validate", level4GeneratorApi.ValidateLevel4Suggestion)       // 验证四级功能点建议
		
		// 异步L4生成API
		privateGroup.POST("/generate-async", level4GeneratorApi.GenerateLevel4Async)        // 异步生成四级功能点
		privateGroup.GET("/task/:taskId/progress", level4GeneratorApi.GetLevel4GenerationTaskProgress) // 获取任务进度
		privateGroup.GET("/task/:taskId/result", level4GeneratorApi.GetLevel4GenerationTaskResult)     // 获取任务结果
		privateGroup.POST("/confirm", level4GeneratorApi.ConfirmLevel4Requirements)         // 确认并入库L4需求
		
		// L4任务管理API
		privateGroup.GET("/tasks", level4GeneratorApi.GetL4GenerationTasks)                 // 获取L4生成任务列表
		privateGroup.GET("/tasks/statistics", level4GeneratorApi.GetL4TaskStatistics)       // 获取L4任务统计信息
		privateGroup.GET("/task/:taskId/detail", level4GeneratorApi.GetL4TaskDetail)        // 获取L4任务详情
		privateGroup.POST("/task/:taskId/cancel", level4GeneratorApi.CancelL4GenerationTask) // 取消L4生成任务
		privateGroup.POST("/task/:taskId/retry", level4GeneratorApi.RetryL4GenerationTask)   // 重试L4生成任务
		privateGroup.DELETE("/task/:taskId", level4GeneratorApi.DeleteL4GenerationTask)      // 删除L4生成任务
		privateGroup.POST("/tasks/batch-delete", level4GeneratorApi.BatchDeleteL4GenerationTasks) // 批量删除L4生成任务
		privateGroup.POST("/tasks/batch-cancel", level4GeneratorApi.BatchCancelL4GenerationTasks) // 批量取消L4生成任务
	}
}

type Level3AnalysisRouter struct{}

// InitLevel3AnalysisRouter 初始化三级功能点分析路由
func (r *Level3AnalysisRouter) InitLevel3AnalysisRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	level3AnalysisApi := v1.ApiGroupApp.NesmaApiGroup.Level3AnalysisApi

	// 私有路由组（需要认证）
	privateGroup := Router.Group("nesma/analysis/level3").Use(middleware.OperationRecord())
	{
		privateGroup.POST("/analyze", level3AnalysisApi.AnalyzeLevel3Requirements)                   // 分析三级功能点
		privateGroup.POST("/apply-optimization", level3AnalysisApi.ApplyOptimizationSuggestion)     // 应用优化建议
		privateGroup.POST("/create-expansion", level3AnalysisApi.CreateExpansionRequirement)        // 创建扩充功能点
		privateGroup.POST("/batch-apply-optimizations", level3AnalysisApi.BatchApplyOptimizations) // 批量应用优化建议
		privateGroup.POST("/batch-create-expansions", level3AnalysisApi.BatchCreateExpansions)     // 批量创建扩充功能点
		privateGroup.GET("/history/:cycleId", level3AnalysisApi.GetAnalysisHistory)                 // 获取分析历史
		privateGroup.GET("/stats/:cycleId", level3AnalysisApi.GetAnalysisStats)                     // 获取分析统计
	}
}