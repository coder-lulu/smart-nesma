package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type EvaluationRouter struct{}

// InitEvaluationRouter 初始化NESMA评估路由
func (r *EvaluationRouter) InitEvaluationRouter(privateGroup, publicGroup *gin.RouterGroup) {
	evaluationApi := nesma.EvaluationApi{}
	// 用户选项路由（无需认证）
	publicGroup.GET("nesma/user/options", evaluationApi.GetUserOptions)                               // 获取用户选项列表
	
	// 兼容性路由 - 支持旧的nesma-evaluation路径
	legacyEvaluationRouter := privateGroup.Group("nesma-evaluation").Use(middleware.OperationRecord())
	{
		legacyEvaluationRouter.POST("create", evaluationApi.CreateEvaluation)                         // 创建评估（兼容旧路径）
		legacyEvaluationRouter.PUT(":id", evaluationApi.UpdateEvaluation)                             // 更新评估（兼容旧路径）
		legacyEvaluationRouter.POST("start", evaluationApi.StartEvaluation)                           // 开始评估（兼容旧路径）
		legacyEvaluationRouter.GET(":id", evaluationApi.GetEvaluation)                                // 获取评估详情（兼容旧路径）
		legacyEvaluationRouter.GET("list", evaluationApi.GetEvaluationList)                           // 获取评估列表（兼容旧路径）
		legacyEvaluationRouter.DELETE(":id", evaluationApi.DeleteEvaluation)                          // 删除评估（兼容旧路径）
		legacyEvaluationRouter.GET("project-cycles", evaluationApi.GetProjectCycles)                  // 获取项目周期（兼容旧路径）
		legacyEvaluationRouter.GET("cycle-versions", evaluationApi.GetCycleVersions)                  // 获取周期版本（兼容旧路径）
		
		// 评估因子相关路由（兼容旧路径）
		legacyEvaluationRouter.GET(":id/factors", evaluationApi.GetEvaluationFactors)                 // 获取评估因子配置
		legacyEvaluationRouter.POST("factors", evaluationApi.SaveEvaluationFactors)                   // 保存评估因子配置
		legacyEvaluationRouter.PUT(":id/factors", evaluationApi.UpdateEvaluationFactors)              // 更新评估因子配置
		legacyEvaluationRouter.GET("default-factors", evaluationApi.GetDefaultNESMAFactors)           // 获取默认NESMA因子配置
		legacyEvaluationRouter.POST("validate-factors", evaluationApi.ValidateEvaluationFactors)      // 验证评估因子配置
		legacyEvaluationRouter.POST("calculate-adjustment", evaluationApi.CalculateAdjustmentFactor)  // 计算调整因子
		legacyEvaluationRouter.POST("preview-calculation", evaluationApi.PreviewFunctionPointCalculation) // 预览功能点计算
		legacyEvaluationRouter.POST(":id/apply-factors", evaluationApi.ApplyFactorsToProject)         // 应用评估因子到项目
		legacyEvaluationRouter.GET(":id/factors-history", evaluationApi.GetFactorsHistory)            // 获取评估因子历史记录
		legacyEvaluationRouter.POST("import-factors", evaluationApi.ImportEvaluationFactors)          // 导入评估因子配置
		legacyEvaluationRouter.GET(":id/export-factors", evaluationApi.ExportEvaluationFactors)       // 导出评估因子配置
	}
	
	evaluationRouter := privateGroup.Group("nesma/evaluation").Use(middleware.OperationRecord())
	{
		evaluationRouter.POST("create", evaluationApi.CreateEvaluation)                               // 创建评估
		evaluationRouter.PUT(":id", evaluationApi.UpdateEvaluation)                                   // 更新评估
		evaluationRouter.POST("start", evaluationApi.StartEvaluation)                                 // 开始评估
		evaluationRouter.GET(":id", evaluationApi.GetEvaluation)                                      // 获取评估详情
		evaluationRouter.GET("list", evaluationApi.GetEvaluationList)                                 // 获取评估列表
		evaluationRouter.DELETE(":id", evaluationApi.DeleteEvaluation)                                // 删除评估
		evaluationRouter.GET("project-cycles", evaluationApi.GetProjectCycles)                        // 获取项目周期
		evaluationRouter.GET("cycle-versions", evaluationApi.GetCycleVersions)                        // 获取周期版本
		evaluationRouter.GET("function-points", evaluationApi.GetFunctionPoints)                      // 获取功能点列表
		evaluationRouter.POST("function-point", evaluationApi.CreateFunctionPoint)                    // 添加功能点
		evaluationRouter.PUT("function-point/:id", evaluationApi.UpdateFunctionPoint)                 // 更新功能点
		evaluationRouter.DELETE("function-point/:id", evaluationApi.DeleteFunctionPoint)              // 删除功能点
		evaluationRouter.GET(":id/stats", evaluationApi.GetEvaluationStats)                           // 获取评估统计
		evaluationRouter.GET("complexity-metrics", evaluationApi.GetComplexityMetrics)                // 获取复杂度指标
		evaluationRouter.GET("validation-items", evaluationApi.GetValidationItems)                    // 获取验证项目
		evaluationRouter.PUT("validation-item/:id", evaluationApi.UpdateValidationItem)               // 更新验证项目
		evaluationRouter.POST("review", evaluationApi.ReviewEvaluation)                               // 评估审核
		evaluationRouter.GET(":id/export", evaluationApi.ExportEvaluationReport)                      // 导出评估报告
		evaluationRouter.GET("project/:projectId/summary", evaluationApi.GetProjectEvaluationSummary) // 获取项目评估摘要
		evaluationRouter.POST(":id/recalculate", evaluationApi.RecalculateEvaluation)                 // 重新计算评估
		
		// 评估进度相关路由
		evaluationProgressApi := nesma.EvaluationProgressApi{}
		evaluationRouter.GET(":id/progress", evaluationProgressApi.GetEvaluationProgress)             // 获取评估进度
		evaluationRouter.POST("progress/start", evaluationProgressApi.StartEvaluation)                // 启动评估进度跟踪
		evaluationRouter.POST("progress/cancel", evaluationProgressApi.CancelEvaluation)              // 取消评估
		evaluationRouter.GET(":id/status", evaluationProgressApi.GetEvaluationStatus)                 // 获取评估状态
		evaluationRouter.GET("progress/stats", evaluationProgressApi.GetEvaluationStats)              // 获取评估统计
		evaluationRouter.GET(":id/logs", evaluationProgressApi.GetEvaluationLogs)                     // 获取评估日志
		evaluationRouter.POST("progress/update", evaluationProgressApi.UpdateEvaluationProgress)      // 手动更新进度(内部接口)
		evaluationRouter.GET("progress/running", evaluationProgressApi.GetRunningEvaluations)         // 获取运行中的评估列表
		evaluationRouter.POST("progress/force-stop", evaluationProgressApi.ForceStopEvaluation)       // 强制停止评估
		
		// 评估因子相关路由
		evaluationRouter.GET(":id/factors", evaluationApi.GetEvaluationFactors)                 // 获取评估因子配置
		evaluationRouter.POST("factors", evaluationApi.SaveEvaluationFactors)                   // 保存评估因子配置
		evaluationRouter.PUT(":id/factors", evaluationApi.UpdateEvaluationFactors)              // 更新评估因子配置
		evaluationRouter.GET("default-factors", evaluationApi.GetDefaultNESMAFactors)           // 获取默认NESMA因子配置
		evaluationRouter.POST("validate-factors", evaluationApi.ValidateEvaluationFactors)      // 验证评估因子配置
		evaluationRouter.POST("calculate-adjustment", evaluationApi.CalculateAdjustmentFactor)  // 计算调整因子
		evaluationRouter.POST("preview-calculation", evaluationApi.PreviewFunctionPointCalculation) // 预览功能点计算
		evaluationRouter.POST(":id/apply-factors", evaluationApi.ApplyFactorsToProject)         // 应用评估因子到项目
		evaluationRouter.GET(":id/factors-history", evaluationApi.GetFactorsHistory)            // 获取评估因子历史记录
		evaluationRouter.POST("import-factors", evaluationApi.ImportEvaluationFactors)          // 导入评估因子配置
		evaluationRouter.GET(":id/export-factors", evaluationApi.ExportEvaluationFactors)       // 导出评估因子配置
		
		// AI监控相关路由
		aiMonitoringApi := nesma.AIMonitoringApi{}
		evaluationRouter.GET("ai/performance-report", aiMonitoringApi.GetPerformanceReport)         // 获取AI性能报告
		evaluationRouter.GET("ai/model/:model_name/metrics", aiMonitoringApi.GetModelMetrics)       // 获取特定模型性能指标
		evaluationRouter.GET("ai/models", aiMonitoringApi.GetAvailableModels)                       // 获取可用模型列表
		evaluationRouter.GET("ai/health", aiMonitoringApi.CheckServiceHealth)                       // 检查AI服务健康状态
		evaluationRouter.GET("ai/recent-requests", aiMonitoringApi.GetRecentRequests)               // 获取最近AI请求记录
		evaluationRouter.GET("ai/metrics-overview", aiMonitoringApi.GetMetricsOverview)             // 获取性能指标概览
		evaluationRouter.GET("ai/model-recommendation", aiMonitoringApi.GetModelRecommendation)     // 获取模型推荐
		evaluationRouter.GET("ai/usage-trends", aiMonitoringApi.GetUsageTrends)                     // 获取使用趋势
	}
}
