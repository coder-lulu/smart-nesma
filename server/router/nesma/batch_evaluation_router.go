package nesma

import (
	"time"
	
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// BatchEvaluationRouter 批量评估路由
type BatchEvaluationRouter struct{}

// InitBatchEvaluationRouter 初始化批量评估路由
func (r *BatchEvaluationRouter) InitBatchEvaluationRouter(privateGroup, publicGroup *gin.RouterGroup) {
	batchEvaluationApi := nesma.BatchEvaluationApi{}
	batchEvaluationRouterWithoutAuth := publicGroup.Group("nesma/batch-evaluation")
	batchEvaluationRouterWithAuth := privateGroup.Group("nesma/batch-evaluation").Use(middleware.JWTAuth())

	{
		// 不需要认证的路由
		batchEvaluationRouterWithoutAuth.GET("/ws", nesma.GlobalWebSocketManager.HandleWebSocket) // WebSocket连接
		batchEvaluationRouterWithoutAuth.POST("/webhook", batchEvaluationApi.HandleWebhook)       // Webhook回调
	}

	{
		// 需要认证的路由

		// 任务管理
		batchEvaluationRouterWithAuth.POST("/create", batchEvaluationApi.CreateBatchEvaluation)       // 创建批量评估任务
		batchEvaluationRouterWithAuth.POST("/start/:taskId", batchEvaluationApi.StartBatchEvaluation) // 启动批量评估任务
		batchEvaluationRouterWithAuth.POST("/cancel/:taskId", batchEvaluationApi.CancelBatchEvaluation) // 取消批量评估任务
		batchEvaluationRouterWithAuth.POST("/retry/:taskId", batchEvaluationApi.RetryBatchEvaluation)   // 重试批量评估任务
		batchEvaluationRouterWithAuth.DELETE("/delete/:taskId", batchEvaluationApi.DeleteBatchEvaluation) // 删除批量评估任务

		// 状态查询
		batchEvaluationRouterWithAuth.GET("/status/:taskId", batchEvaluationApi.GetBatchEvaluationStatus) // 获取批量评估任务状态
		batchEvaluationRouterWithAuth.GET("/detail/:taskId", batchEvaluationApi.GetBatchEvaluationDetail) // 获取批量评估任务详情
		batchEvaluationRouterWithAuth.GET("/result/:taskId", batchEvaluationApi.GetBatchEvaluationResult) // 获取批量评估任务结果
		batchEvaluationRouterWithAuth.GET("/list", batchEvaluationApi.GetBatchEvaluationList)             // 获取批量评估任务列表

		// 批次管理
		batchEvaluationRouterWithAuth.GET("/batches/:taskId", batchEvaluationApi.GetBatchDetails)       // 获取批次详情
		batchEvaluationRouterWithAuth.POST("/batches/retry", batchEvaluationApi.RetryBatches)           // 重试特定批次
		batchEvaluationRouterWithAuth.GET("/batches/status/:batchId", batchEvaluationApi.GetBatchStatus) // 获取批次状态

		// 结果分析
		batchEvaluationRouterWithAuth.GET("/analysis/:taskId", batchEvaluationApi.GetAnalysisResult)      // 获取分析结果
		batchEvaluationRouterWithAuth.POST("/compare", batchEvaluationApi.CompareBatchEvaluations)        // 比较批量评估结果
		batchEvaluationRouterWithAuth.GET("/metrics/:taskId", batchEvaluationApi.GetEvaluationMetrics)    // 获取评估指标
		batchEvaluationRouterWithAuth.GET("/quality/:taskId", batchEvaluationApi.GetQualityMetrics)       // 获取质量指标
		
		// 智能分片测试
		batchEvaluationRouterWithAuth.POST("/test-sharding", batchEvaluationApi.TestSharding)            // 测试智能分片算法

		// 导出功能
		batchEvaluationRouterWithAuth.POST("/export/:taskId", batchEvaluationApi.ExportBatchEvaluation)  // 导出批量评估结果
		batchEvaluationRouterWithAuth.GET("/export/status/:exportId", batchEvaluationApi.GetExportStatus) // 获取导出状态
		batchEvaluationRouterWithAuth.GET("/export/download/:exportId", batchEvaluationApi.DownloadExport) // 下载导出文件

		// 模板管理
		batchEvaluationRouterWithAuth.POST("/templates", batchEvaluationApi.CreateTemplate)              // 创建模板
		batchEvaluationRouterWithAuth.GET("/templates", batchEvaluationApi.GetTemplateList)              // 获取模板列表
		batchEvaluationRouterWithAuth.GET("/templates/:templateId", batchEvaluationApi.GetTemplate)      // 获取模板详情
		batchEvaluationRouterWithAuth.PUT("/templates/:templateId", batchEvaluationApi.UpdateTemplate)   // 更新模板
		batchEvaluationRouterWithAuth.DELETE("/templates/:templateId", batchEvaluationApi.DeleteTemplate) // 删除模板

		// 调度管理
		batchEvaluationRouterWithAuth.POST("/schedules", batchEvaluationApi.CreateSchedule)              // 创建调度
		batchEvaluationRouterWithAuth.GET("/schedules", batchEvaluationApi.GetScheduleList)              // 获取调度列表
		batchEvaluationRouterWithAuth.GET("/schedules/:scheduleId", batchEvaluationApi.GetSchedule)      // 获取调度详情
		batchEvaluationRouterWithAuth.PUT("/schedules/:scheduleId", batchEvaluationApi.UpdateSchedule)   // 更新调度
		batchEvaluationRouterWithAuth.DELETE("/schedules/:scheduleId", batchEvaluationApi.DeleteSchedule) // 删除调度
		batchEvaluationRouterWithAuth.POST("/schedules/:scheduleId/enable", batchEvaluationApi.EnableSchedule) // 启用调度
		batchEvaluationRouterWithAuth.POST("/schedules/:scheduleId/disable", batchEvaluationApi.DisableSchedule) // 禁用调度

		// 通知管理
		batchEvaluationRouterWithAuth.POST("/notifications", batchEvaluationApi.SendNotification)         // 发送通知
		batchEvaluationRouterWithAuth.GET("/notifications", batchEvaluationApi.GetNotificationList)       // 获取通知列表
		batchEvaluationRouterWithAuth.GET("/notifications/:notificationId", batchEvaluationApi.GetNotification) // 获取通知详情

		// 系统监控
		batchEvaluationRouterWithAuth.GET("/monitoring/system", batchEvaluationApi.GetSystemMonitoring)   // 获取系统监控
		batchEvaluationRouterWithAuth.GET("/monitoring/metrics", batchEvaluationApi.GetMetrics)           // 获取系统指标
		batchEvaluationRouterWithAuth.GET("/monitoring/health", batchEvaluationApi.GetHealthStatus)       // 获取健康状态

		// 配置管理
		batchEvaluationRouterWithAuth.GET("/config", batchEvaluationApi.GetConfig)                        // 获取配置
		batchEvaluationRouterWithAuth.PUT("/config", batchEvaluationApi.UpdateConfig)                     // 更新配置
		batchEvaluationRouterWithAuth.POST("/config/validate", batchEvaluationApi.ValidateConfig)         // 验证配置
		batchEvaluationRouterWithAuth.POST("/config/reset", batchEvaluationApi.ResetConfig)               // 重置配置

		// 日志管理
		batchEvaluationRouterWithAuth.GET("/logs/:taskId", batchEvaluationApi.GetTaskLogs)                // 获取任务日志
		batchEvaluationRouterWithAuth.GET("/logs/system", batchEvaluationApi.GetSystemLogs)               // 获取系统日志
		batchEvaluationRouterWithAuth.GET("/logs/error", batchEvaluationApi.GetErrorLogs)                 // 获取错误日志

		// 统计报表
		batchEvaluationRouterWithAuth.GET("/reports/summary", batchEvaluationApi.GetSummaryReport)        // 获取摘要报表
		batchEvaluationRouterWithAuth.GET("/reports/performance", batchEvaluationApi.GetPerformanceReport) // 获取性能报表
		batchEvaluationRouterWithAuth.GET("/reports/quality", batchEvaluationApi.GetQualityReport)        // 获取质量报表
		batchEvaluationRouterWithAuth.GET("/reports/usage", batchEvaluationApi.GetUsageReport)            // 获取使用报表
	}
}

// InitBatchEvaluationWebSocketRouter 初始化批量评估WebSocket路由
func (r *BatchEvaluationRouter) InitBatchEvaluationWebSocketRouter(Router *gin.RouterGroup) {
	batchEvaluationWSRouterWithoutAuth := Router.Group("batch-evaluation/ws")
	{
		// WebSocket连接
		batchEvaluationWSRouterWithoutAuth.GET("/connect", nesma.GlobalWebSocketManager.HandleWebSocket)
		
		// WebSocket状态
		batchEvaluationWSRouterWithoutAuth.GET("/status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"connected_clients": nesma.GlobalWebSocketManager.GetConnectedClients(),
				"server_time":      time.Now(),
			})
		})
	}
}

// 添加中间件支持
func (r *BatchEvaluationRouter) InitBatchEvaluationMiddleware(Router *gin.RouterGroup) {
	// 添加自定义中间件
	Router.Use(middleware.JWTAuth())
	Router.Use(middleware.CasbinHandler())
	Router.Use(RateLimitMiddleware())
	Router.Use(ValidationMiddleware())
	Router.Use(LoggingMiddleware())
}

// RateLimitMiddleware 限流中间件
func RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 实现限流逻辑
		// 可以使用redis或内存存储实现
		c.Next()
	}
}

// ValidationMiddleware 验证中间件
func ValidationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 实现请求验证逻辑
		c.Next()
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 实现请求日志记录
		start := time.Now()
		c.Next()
		
		// 记录请求信息
		global.GVA_LOG.Info("BatchEvaluation API Request",
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Duration("duration", time.Since(start)),
			zap.Int("status", c.Writer.Status()),
		)
	}
}