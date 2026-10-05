package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type ChatRouter struct{}

// InitChatRouter 初始化AI聊天路由
func (r *ChatRouter) InitChatRouter(PrivateRouter *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	chatRouter := PrivateRouter.Group("chat").Use(middleware.OperationRecord())
	chatRouterWithoutRecord := PrivateRouter.Group("chat")
	chatPublicRouter := PublicRouter.Group("chat") // 公共路由，不需要认证
	var chatApi = v1.ApiGroupApp.NesmaApiGroup.ChatApi
	{
		// 需要记录操作的路由
		chatRouter.POST("send", chatApi.SendMessage)                          // 发送消息
		chatRouter.POST("session", chatApi.CreateSession)                     // 创建会话
		chatRouter.PUT("session/:sessionId", chatApi.UpdateSession)           // 更新会话
		chatRouter.DELETE("session/:sessionId", chatApi.DeleteSession)        // 删除会话
		chatRouter.POST("session/:sessionId/clear", chatApi.ClearMessages)    // 清空会话消息
		chatRouter.POST("reset-circuit-breaker", chatApi.ResetCircuitBreaker) // 重置熔断器
	}
	{
		// 不需要记录操作的路由（查询类接口）
		chatRouterWithoutRecord.GET("sessions", chatApi.GetSessionList)                     // 获取会话列表
		chatRouterWithoutRecord.GET("session/:sessionId", chatApi.GetSession)               // 获取会话详情
		chatRouterWithoutRecord.GET("session/:sessionId/messages", chatApi.GetMessages)     // 获取会话消息
		chatRouterWithoutRecord.GET("stats", chatApi.GetChatStats)                          // 获取聊天统计
		chatRouterWithoutRecord.GET("session/:sessionId/export", chatApi.ExportChatHistory) // 导出聊天历史
		chatRouterWithoutRecord.GET("stream", chatApi.SendMessageStream)                    // 流式发送消息
		chatRouterWithoutRecord.GET("health", chatApi.GetAIHealth)                          // 获取AI服务健康状态
	}
	{
		// 公共路由（不需要JWT中间件的路由）
		chatPublicRouter.GET("ws", chatApi.WebSocketUpgrade) // WebSocket聊天接口，通过查询参数传递token
	}
}
