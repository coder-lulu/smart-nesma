package nesma

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketMessage WebSocket消息结构
type WebSocketMessage struct {
	Type      string                 `json:"type"` // message, status, error, complete
	SessionID uint                   `json:"sessionId"`
	Content   string                 `json:"content"`
	Role      string                 `json:"role"`
	MessageID uint                   `json:"messageId"`
	Metadata  map[string]interface{} `json:"metadata"`
	Error     string                 `json:"error,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// ChatWebSocketRequest 聊天WebSocket请求
type ChatWebSocketRequest struct {
	Action    string                 `json:"action"` // send_message, join_session, leave_session
	SessionID uint                   `json:"sessionId"`
	ProjectID *uint                  `json:"projectId"`
	Message   string                 `json:"message"`
	ModelName string                 `json:"modelName"`
	Context   map[string]interface{} `json:"context"`
}

// WebSocket升级器
var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// 允许跨域连接
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// ChatWebSocketHandler WebSocket聊天处理器
func (c *ChatApi) ChatWebSocketHandler(ctx *gin.Context) {
	// 从查询参数获取token
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "缺少访问令牌"})
		return
	}

	// 解析token获取用户ID
	j := utils.NewJWT()
	claims, err := j.ParseToken(token)
	if err != nil {
		global.GVA_LOG.Error("WebSocket token解析失败", zap.Error(err))
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的访问令牌"})
		return
	}

	userID := claims.BaseClaims.ID
	if userID == 0 {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "无效的用户ID"})
		return
	}

	// 升级到WebSocket连接
	conn, err := wsUpgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		global.GVA_LOG.Error("WebSocket升级失败", zap.Error(err))
		return
	}
	defer conn.Close()

	global.GVA_LOG.Info("WebSocket连接建立", zap.Uint("userID", userID))

	// 发送连接成功消息
	c.sendWebSocketMessage(conn, WebSocketMessage{
		Type:      "status",
		Content:   "连接成功",
		Timestamp: time.Now(),
	})

	// 处理WebSocket消息
	for {
		var req ChatWebSocketRequest
		err := conn.ReadJSON(&req)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				global.GVA_LOG.Error("WebSocket读取错误", zap.Error(err))
			}
			break
		}

		switch req.Action {
		case "send_message":
			c.handleWebSocketMessage(conn, userID, req)
		case "ping":
			c.sendWebSocketMessage(conn, WebSocketMessage{
				Type:      "pong",
				Content:   "pong",
				Timestamp: time.Now(),
			})
		default:
			c.sendWebSocketMessage(conn, WebSocketMessage{
				Type:      "error",
				Error:     "未知的操作类型",
				Timestamp: time.Now(),
			})
		}
	}

	global.GVA_LOG.Info("WebSocket连接关闭", zap.Uint("userID", userID))
}

// handleWebSocketMessage 处理WebSocket消息
func (c *ChatApi) handleWebSocketMessage(conn *websocket.Conn, userID uint, req ChatWebSocketRequest) {
	// 获取聊天服务
	chatService := nesma.GetChatService()
	if chatService == nil {
		c.sendWebSocketMessage(conn, WebSocketMessage{
			Type:      "error",
			Error:     "聊天服务不可用",
			Timestamp: time.Now(),
		})
		return
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// 构建聊天请求
	chatRequest := nesma.ChatRequest{
		SessionID: &req.SessionID,
		ProjectID: req.ProjectID,
		Message:   req.Message,
		ModelName: req.ModelName,
		Context:   req.Context,
		Stream:    true,
	}

	// 如果没有会话ID，先创建会话
	if req.SessionID == 0 {
		session, err := chatService.CreateSession(userID, req.ProjectID, "", req.ModelName)
		if err != nil {
			c.sendWebSocketMessage(conn, WebSocketMessage{
				Type:      "error",
				Error:     fmt.Sprintf("创建会话失败: %v", err),
				Timestamp: time.Now(),
			})
			return
		}
		req.SessionID = session.ID
		chatRequest.SessionID = &session.ID
	}

	// 发送用户消息确认
	c.sendWebSocketMessage(conn, WebSocketMessage{
		Type:      "message",
		SessionID: req.SessionID,
		Content:   req.Message,
		Role:      "user",
		MessageID: 0, // 临时ID
		Timestamp: time.Now(),
	})

	// 发送开始生成状态
	c.sendWebSocketMessage(conn, WebSocketMessage{
		Type:      "status",
		SessionID: req.SessionID,
		Content:   "AI正在思考...",
		Timestamp: time.Now(),
	})

	// 流式发送消息
	err := chatService.SendMessageStream(ctx, userID, chatRequest, func(chunk string, isComplete bool, tokenCount int) {
		if isComplete {
			// 发送完成消息
			c.sendWebSocketMessage(conn, WebSocketMessage{
				Type:      "complete",
				SessionID: req.SessionID,
				Content:   "",
				Role:      "assistant",
				Metadata: map[string]interface{}{
					"tokenCount": tokenCount,
					"modelName":  req.ModelName,
				},
				Timestamp: time.Now(),
			})
		} else {
			// 发送流式内容
			c.sendWebSocketMessage(conn, WebSocketMessage{
				Type:      "message",
				SessionID: req.SessionID,
				Content:   chunk,
				Role:      "assistant",
				Metadata: map[string]interface{}{
					"streaming": true,
				},
				Timestamp: time.Now(),
			})
		}
	})

	if err != nil {
		global.GVA_LOG.Error("流式消息发送失败", zap.Error(err))
		c.sendWebSocketMessage(conn, WebSocketMessage{
			Type:      "error",
			SessionID: req.SessionID,
			Error:     fmt.Sprintf("消息发送失败: %v", err),
			Timestamp: time.Now(),
		})
	}
}

// sendWebSocketMessage 发送WebSocket消息
func (c *ChatApi) sendWebSocketMessage(conn *websocket.Conn, message WebSocketMessage) {
	if err := conn.WriteJSON(message); err != nil {
		global.GVA_LOG.Error("WebSocket消息发送失败", zap.Error(err))
	}
}

// @Tags AI对话
// @Summary WebSocket聊天接口
// @Security ApiKeyAuth
// @Router /chat/ws [get]
func (c *ChatApi) WebSocketUpgrade(ctx *gin.Context) {
	c.ChatWebSocketHandler(ctx)
}
