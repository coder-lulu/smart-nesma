package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ChatApi struct{}

// @Tags AI对话
// @Summary 发送消息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesma.ChatRequest true "聊天请求"
// @Success 200 {object} response.Response{data=nesma.ChatResponse} "发送成功"
// @Router /chat/send [post]
func (c *ChatApi) SendMessage(ctx *gin.Context) {
	var req nesma.ChatRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请求参数错误: "+err.Error(), ctx)
		return
	}

	// 获取用户ID
	userID := utils.GetUserID(ctx)
	if userID == 0 {
		response.FailWithMessage("用户未登录", ctx)
		return
	}

	// 添加详细调试日志
	global.GVA_LOG.Info("=== 聊天API收到请求 ===",
		zap.Uint("userID", userID),
		zap.Any("sessionID", req.SessionID),
		zap.Any("projectID", req.ProjectID),
		zap.String("model", req.ModelName),
		zap.String("message", req.Message),
		zap.Any("context", req.Context),
		zap.Bool("stream", req.Stream),
		zap.Int("messageLength", len(req.Message)),
	)

	// 检查AI服务健康状态
	aiService := nesma.GetAIServiceWithProtection(req.ModelName)
	if aiService == nil {
		global.GVA_LOG.Error("AI服务不可用", zap.String("model", req.ModelName))
		response.FailWithMessage("AI服务不可用，请稍后重试", ctx)
		return
	}

	// 检查AI服务健康状态
	if !aiService.IsHealthy(ctx) {
		global.GVA_LOG.Error("AI服务健康检查失败", zap.String("model", req.ModelName))
		response.FailWithMessage("AI服务暂时不可用，请稍后重试", ctx)
		return
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()
	if chatService == nil {
		global.GVA_LOG.Error("聊天服务不可用")
		response.FailWithMessage("聊天服务不可用", ctx)
		return
	}

	// 发送消息
	chatResponse, err := chatService.SendMessage(ctx, userID, req)
	if err != nil {
		global.GVA_LOG.Error("发送消息失败", zap.Error(err))

		// 根据错误类型返回不同的错误信息
		errorMsg := "发送消息失败"
		if strings.Contains(err.Error(), "invalid character '<'") {
			errorMsg = "AI服务配置错误，请检查API密钥和服务地址"
		} else if strings.Contains(err.Error(), "context deadline exceeded") {
			errorMsg = "AI服务响应超时，请稍后重试"
		} else if strings.Contains(err.Error(), "connection refused") {
			errorMsg = "无法连接到AI服务，请检查网络连接"
		} else if strings.Contains(err.Error(), "unauthorized") {
			errorMsg = "AI服务认证失败，请检查API密钥"
		}

		response.FailWithMessage(errorMsg+": "+err.Error(), ctx)
		return
	}

	global.GVA_LOG.Info("聊天响应成功",
		zap.Uint("sessionID", chatResponse.SessionID),
		zap.Uint("messageID", chatResponse.MessageID),
		zap.String("model", chatResponse.ModelName),
		zap.Int("tokenCount", chatResponse.TokenCount),
		zap.Int64("elapsedTime", chatResponse.ElapsedTime),
	)

	response.OkWithData(chatResponse, ctx)
}

// @Tags AI对话
// @Summary 流式发送消息
// @Security ApiKeyAuth
// @accept application/json
// @Produce text/event-stream
// @Param sessionId query int false "会话ID"
// @Param projectId query int false "项目ID"
// @Param message query string true "消息内容"
// @Param modelName query string false "模型名称"
// @Param token query string false "访问令牌（可选，优先级高于Header中的token）"
// @Success 200 {string} string "流式响应"
// @Router /chat/stream [get]
func (c *ChatApi) SendMessageStream(ctx *gin.Context) {
	var userID uint

	// 优先从查询参数获取token（为了支持EventSource）
	token := ctx.Query("token")
	if token != "" {
		// 解析token获取用户ID
		j := utils.NewJWT()
		claims, err := j.ParseToken(token)
		if err != nil {
			global.GVA_LOG.Error("SSE token解析失败", zap.Error(err))
			ctx.String(http.StatusUnauthorized, "data: {\"type\":\"error\",\"message\":\"无效的访问令牌\"}\n\n")
			return
		}
		userID = claims.BaseClaims.ID
	} else {
		// 回退到标准的JWT中间件
		userID = utils.GetUserID(ctx)
	}

	if userID == 0 {
		ctx.String(http.StatusUnauthorized, "data: {\"type\":\"error\",\"message\":\"用户未登录\"}\n\n")
		return
	}

	// 解析参数
	var sessionID *uint
	if sessionIDStr := ctx.Query("sessionId"); sessionIDStr != "" {
		if id, err := strconv.ParseUint(sessionIDStr, 10, 32); err == nil {
			sessionIDUint := uint(id)
			sessionID = &sessionIDUint
		}
	}

	var projectID *uint
	if projectIDStr := ctx.Query("projectId"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			projectIDUint := uint(id)
			projectID = &projectIDUint
		}
	}

	message := ctx.Query("message")
	if message == "" {
		ctx.String(400, "data: {\"type\":\"error\",\"message\":\"message parameter is required\"}\n\n")
		return
	}

	modelName := ctx.Query("modelName")
	if modelName == "" {
		modelName = "deepseek-chat"
	}

	global.GVA_LOG.Info("=== SSE流式请求 ===",
		zap.Uint("userID", userID),
		zap.Any("sessionID", sessionID),
		zap.Any("projectID", projectID),
		zap.String("model", modelName),
		zap.String("message", message),
	)

	// 设置SSE响应头
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")
	ctx.Header("Access-Control-Allow-Headers", "Cache-Control")

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 创建上下文
	requestCtx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	// 创建聊天请求
	req := nesma.ChatRequest{
		SessionID: sessionID,
		ProjectID: projectID,
		Message:   message,
		ModelName: modelName,
		Stream:    true,
	}

	// 发送消息并处理流式响应
	err := chatService.SendMessageStream(requestCtx, userID, req, func(chunk string, isComplete bool, tokenCount int) {
		if isComplete {
			// 发送完成事件
			data := map[string]interface{}{
				"type":       "done",
				"tokenCount": tokenCount,
			}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
			ctx.Writer.Flush()
		} else {
			// 发送内容块
			data := map[string]interface{}{
				"type":    "content",
				"content": chunk,
			}
			jsonData, _ := json.Marshal(data)
			fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
			ctx.Writer.Flush()
		}
	})

	if err != nil {
		global.GVA_LOG.Error("SSE流式发送失败", zap.Error(err))
		// 发送错误事件
		data := map[string]interface{}{
			"type":    "error",
			"message": err.Error(),
		}
		jsonData, _ := json.Marshal(data)
		fmt.Fprintf(ctx.Writer, "data: %s\n\n", jsonData)
		ctx.Writer.Flush()
	}
}

// CreateSessionRequest 创建会话请求
type CreateSessionRequest struct {
	ProjectID   *uint  `json:"projectId"`
	Title       string `json:"title"`
	ModelName   string `json:"modelName"`
	Description string `json:"description"`
}

// @Tags AI对话
// @Summary 创建会话
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body CreateSessionRequest true "创建会话请求"
// @Success 200 {object} response.Response{data=nesma.ChatSession} "创建成功"
// @Router /chat/session [post]
func (c *ChatApi) CreateSession(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	var req CreateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), ctx)
		return
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 创建会话
	session, err := chatService.CreateSession(userID, req.ProjectID, req.Title, req.ModelName)
	if err != nil {
		global.GVA_LOG.Error("创建会话失败: " + err.Error())
		response.FailWithMessage("创建会话失败: "+err.Error(), ctx)
		return
	}

	// 如果有描述，更新会话
	if req.Description != "" {
		err = chatService.UpdateSession(session.ID, userID, map[string]interface{}{
			"description": req.Description,
		})
		if err != nil {
			global.GVA_LOG.Warn("更新会话描述失败: " + err.Error())
		}
	}

	response.OkWithData(session, ctx)
}

// @Tags AI对话
// @Summary 获取会话列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(20)
// @Param isActive query bool false "是否活跃"
// @Success 200 {object} response.Response{data=nesma.SessionListResponse} "获取成功"
// @Router /chat/sessions [get]
func (c *ChatApi) GetSessionList(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	// 解析参数
	var projectID *uint
	if projectIDStr := ctx.Query("projectId"); projectIDStr != "" {
		if id, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			projectIDUint := uint(id)
			projectID = &projectIDUint
		}
	}

	page := 1
	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 20
	if pageSizeStr := ctx.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	var isActive *bool
	if isActiveStr := ctx.Query("isActive"); isActiveStr != "" {
		if active, err := strconv.ParseBool(isActiveStr); err == nil {
			isActive = &active
		}
	}

	req := nesma.SessionListRequest{
		UserID:    userID,
		ProjectID: projectID,
		Page:      page,
		PageSize:  pageSize,
		IsActive:  isActive,
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 获取会话列表
	sessionList, err := chatService.GetSessionList(req)
	if err != nil {
		global.GVA_LOG.Error("获取会话列表失败: " + err.Error())
		response.FailWithMessage("获取会话列表失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(sessionList, ctx)
}

// @Tags AI对话
// @Summary 获取会话详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param sessionId path int true "会话ID"
// @Success 200 {object} response.Response{data=nesma.ChatSession} "获取成功"
// @Router /chat/session/{sessionId} [get]
func (c *ChatApi) GetSession(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	sessionIDStr := ctx.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("会话ID无效", ctx)
		return
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 获取会话
	session, err := chatService.GetSession(uint(sessionID), userID)
	if err != nil {
		global.GVA_LOG.Error("获取会话失败: " + err.Error())
		response.FailWithMessage("获取会话失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(session, ctx)
}

// UpdateSessionRequest 更新会话请求
type UpdateSessionRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	ModelName   *string `json:"modelName"`
}

// @Tags AI对话
// @Summary 更新会话
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param sessionId path int true "会话ID"
// @Param data body UpdateSessionRequest true "更新会话请求"
// @Success 200 {object} response.Response{} "更新成功"
// @Router /chat/session/{sessionId} [put]
func (c *ChatApi) UpdateSession(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	sessionIDStr := ctx.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("会话ID无效", ctx)
		return
	}

	var req UpdateSessionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), ctx)
		return
	}

	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.ModelName != nil {
		updates["model_name"] = *req.ModelName
	}

	if len(updates) == 0 {
		response.FailWithMessage("没有要更新的字段", ctx)
		return
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 更新会话
	err = chatService.UpdateSession(uint(sessionID), userID, updates)
	if err != nil {
		global.GVA_LOG.Error("更新会话失败: " + err.Error())
		response.FailWithMessage("更新会话失败: "+err.Error(), ctx)
		return
	}

	response.OkWithMessage("更新成功", ctx)
}

// @Tags AI对话
// @Summary 删除会话
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param sessionId path int true "会话ID"
// @Success 200 {object} response.Response{} "删除成功"
// @Router /chat/session/{sessionId} [delete]
func (c *ChatApi) DeleteSession(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	sessionIDStr := ctx.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("会话ID无效", ctx)
		return
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 删除会话
	err = chatService.DeleteSession(uint(sessionID), userID)
	if err != nil {
		global.GVA_LOG.Error("删除会话失败: " + err.Error())
		response.FailWithMessage("删除会话失败: "+err.Error(), ctx)
		return
	}

	response.OkWithMessage("删除成功", ctx)
}

// @Tags AI对话
// @Summary 获取会话消息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param sessionId path int true "会话ID"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(50)
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /chat/session/{sessionId}/messages [get]
func (c *ChatApi) GetMessages(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	sessionIDStr := ctx.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("会话ID无效", ctx)
		return
	}

	page := 1
	if pageStr := ctx.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	pageSize := 50
	if pageSizeStr := ctx.Query("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 获取消息
	messages, total, err := chatService.GetMessages(uint(sessionID), userID, page, pageSize)
	if err != nil {
		global.GVA_LOG.Error("获取消息失败: " + err.Error())
		response.FailWithMessage("获取消息失败: "+err.Error(), ctx)
		return
	}

	result := map[string]interface{}{
		"messages": messages,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}

	response.OkWithData(result, ctx)
}

// @Tags AI对话
// @Summary 清空会话消息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param sessionId path int true "会话ID"
// @Success 200 {object} response.Response{} "清空成功"
// @Router /chat/session/{sessionId}/clear [post]
func (c *ChatApi) ClearMessages(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	sessionIDStr := ctx.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("会话ID无效", ctx)
		return
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 清空消息
	err = chatService.ClearMessages(uint(sessionID), userID)
	if err != nil {
		global.GVA_LOG.Error("清空消息失败: " + err.Error())
		response.FailWithMessage("清空消息失败: "+err.Error(), ctx)
		return
	}

	response.OkWithMessage("清空成功", ctx)
}

// @Tags AI对话
// @Summary 获取聊天统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param days query int false "统计天数" default(30)
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /chat/stats [get]
func (c *ChatApi) GetChatStats(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	days := 30
	if daysStr := ctx.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 获取统计
	stats, err := chatService.GetChatStats(userID, days)
	if err != nil {
		global.GVA_LOG.Error("获取聊天统计失败: " + err.Error())
		response.FailWithMessage("获取聊天统计失败: "+err.Error(), ctx)
		return
	}

	response.OkWithData(stats, ctx)
}

// @Tags AI对话
// @Summary 导出聊天历史
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param sessionId path int true "会话ID"
// @Param format query string false "导出格式" Enums(json,txt) default(json)
// @Success 200 {object} response.Response{data=map[string]interface{}} "导出成功"
// @Router /chat/session/{sessionId}/export [get]
func (c *ChatApi) ExportChatHistory(ctx *gin.Context) {
	userID := utils.GetUserID(ctx)

	sessionIDStr := ctx.Param("sessionId")
	sessionID, err := strconv.ParseUint(sessionIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("会话ID无效", ctx)
		return
	}

	format := ctx.Query("format")
	if format == "" {
		format = "json"
	}

	// 获取聊天服务
	chatService := nesma.GetChatService()

	// 导出历史
	data, err := chatService.ExportChatHistory(uint(sessionID), userID, format)
	if err != nil {
		global.GVA_LOG.Error("导出聊天历史失败: " + err.Error())
		response.FailWithMessage("导出聊天历史失败: "+err.Error(), ctx)
		return
	}

	result := map[string]interface{}{
		"data":      string(data),
		"format":    format,
		"timestamp": time.Now().Unix(),
	}

	response.OkWithData(result, ctx)
}

// @Tags AI对话
// @Summary 重置AI服务熔断器
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response "重置成功"
// @Router /chat/reset-circuit-breaker [post]
func (c *ChatApi) ResetCircuitBreaker(ctx *gin.Context) {
	// 重置所有熔断器
	nesma.ResetAllCircuitBreakers()

	// 获取熔断器状态
	stats := nesma.GetCircuitBreakerStats()

	response.OkWithDetailed(gin.H{
		"message": "熔断器已重置",
		"stats":   stats,
	}, "重置成功", ctx)
}

// @Tags AI对话
// @Summary 获取AI服务健康状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response "获取成功"
// @Router /chat/health [get]
func (c *ChatApi) GetAIHealth(ctx *gin.Context) {
	requestCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	health := nesma.GetAIServiceHealth(requestCtx)
	stats := nesma.GetCircuitBreakerStats()

	response.OkWithDetailed(gin.H{
		"health": health,
		"stats":  stats,
	}, "获取成功", ctx)
}
