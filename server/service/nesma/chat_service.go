package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ChatService 聊天服务
type ChatService struct {
	db *gorm.DB
}

// ChatSession 聊天会话
type ChatSession struct {
	ID          uint          `json:"id" gorm:"primaryKey"`
	UserID      uint          `json:"userId" gorm:"not null;index"`
	ProjectID   *uint         `json:"projectId" gorm:"index"`
	Title       string        `json:"title" gorm:"type:varchar(255);not null"`
	Description string        `json:"description" gorm:"type:text"`
	ModelName   string        `json:"modelName" gorm:"type:varchar(50);not null"`
	IsActive    bool          `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time     `json:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt"`
	Messages    []ChatMessage `json:"messages" gorm:"foreignKey:SessionID"`
}

// ChatMessage 聊天消息
type ChatMessage struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	SessionID  uint      `json:"sessionId" gorm:"not null;index"`
	Role       string    `json:"role" gorm:"type:varchar(20);not null"` // user, assistant, system
	Content    string    `json:"content" gorm:"type:text;not null"`
	Metadata   string    `json:"metadata" gorm:"type:json"`
	TokenCount int       `json:"tokenCount" gorm:"default:0"`
	CreatedAt  time.Time `json:"createdAt"`
}

// ChatRequest 聊天请求
type ChatRequest struct {
	SessionID *uint                  `json:"sessionId"`
	ProjectID *uint                  `json:"projectId"`
	Message   string                 `json:"message" binding:"required"`
	ModelName string                 `json:"modelName"`
	Context   map[string]interface{} `json:"context"`
	Stream    bool                   `json:"stream"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	SessionID   uint                   `json:"sessionId"`
	MessageID   uint                   `json:"messageId"`
	Content     string                 `json:"content"`
	Role        string                 `json:"role"`
	TokenCount  int                    `json:"tokenCount"`
	ModelName   string                 `json:"modelName"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"createdAt"`
	ElapsedTime int64                  `json:"elapsedTime"`
}

// SessionListRequest 会话列表请求
type SessionListRequest struct {
	UserID    uint  `json:"userId"`
	ProjectID *uint `json:"projectId"`
	Page      int   `json:"page"`
	PageSize  int   `json:"pageSize"`
	IsActive  *bool `json:"isActive"`
}

// SessionListResponse 会话列表响应
type SessionListResponse struct {
	Sessions []ChatSession `json:"sessions"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"pageSize"`
}

// NewChatService 创建聊天服务
func NewChatService() *ChatService {
	return &ChatService{
		db: global.GVA_DB,
	}
}

// CreateSession 创建新会话
func (cs *ChatService) CreateSession(userID uint, projectID *uint, title string, modelName string) (*ChatSession, error) {
	if title == "" {
		title = fmt.Sprintf("Chat Session %s", time.Now().Format("2006-01-02 15:04"))
	}

	if modelName == "" {
		modelName = "deepseek" // 默认模型
	}

	session := &ChatSession{
		UserID:    userID,
		ProjectID: projectID,
		Title:     title,
		ModelName: modelName,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := cs.db.Create(session).Error
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetSession 获取会话
func (cs *ChatService) GetSession(sessionID uint, userID uint) (*ChatSession, error) {
	var session ChatSession
	err := cs.db.Preload("Messages").Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetSessionList 获取会话列表
func (cs *ChatService) GetSessionList(req SessionListRequest) (*SessionListResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := cs.db.Model(&ChatSession{}).Where("user_id = ?", req.UserID)

	if req.ProjectID != nil {
		query = query.Where("project_id = ?", *req.ProjectID)
	}

	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, err
	}

	var sessions []ChatSession
	err = query.Order("updated_at DESC").
		Offset((req.Page - 1) * req.PageSize).
		Limit(req.PageSize).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}

	return &SessionListResponse{
		Sessions: sessions,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// UpdateSession 更新会话
func (cs *ChatService) UpdateSession(sessionID uint, userID uint, updates map[string]interface{}) error {
	updates["updated_at"] = time.Now()
	return cs.db.Model(&ChatSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Updates(updates).Error
}

// DeleteSession 删除会话
func (cs *ChatService) DeleteSession(sessionID uint, userID uint) error {
	// 软删除：设置为非活跃状态
	return cs.db.Model(&ChatSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Update("is_active", false).Error
}

// SendMessage 发送消息
func (cs *ChatService) SendMessage(ctx context.Context, userID uint, req ChatRequest) (*ChatResponse, error) {
	start := time.Now()

	// 添加详细调试日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("=== 聊天服务开始处理消息 ===",
			zap.Uint("userID", userID),
			zap.Any("sessionID", req.SessionID),
			zap.Any("projectID", req.ProjectID),
			zap.String("message", req.Message),
			zap.String("modelName", req.ModelName),
			zap.Any("context", req.Context),
			zap.Bool("stream", req.Stream),
		)
	}

	// 如果没有指定会话，创建新会话
	var session *ChatSession
	var err error

	if req.SessionID == nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Info("创建新会话", zap.String("modelName", req.ModelName))
		}
		session, err = cs.CreateSession(userID, req.ProjectID, "", req.ModelName)
		if err != nil {
			return nil, fmt.Errorf("创建会话失败: %w", err)
		}
	} else {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Info("获取现有会话", zap.Uint("sessionID", *req.SessionID))
		}
		session, err = cs.GetSession(*req.SessionID, userID)
		if err != nil {
			return nil, fmt.Errorf("获取会话失败: %w", err)
		}
	}

	// 保存用户消息
	userMessage := &ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}

	// 设置metadata为有效JSON格式
	if req.Context != nil && len(req.Context) > 0 {
		if metadataBytes, err := json.Marshal(req.Context); err == nil {
			userMessage.Metadata = string(metadataBytes)
		} else {
			userMessage.Metadata = "{}"
		}
	} else {
		userMessage.Metadata = "{}" // 设置为空JSON对象而不是空字符串
	}

	err = cs.db.Create(userMessage).Error
	if err != nil {
		return nil, fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 准备对话历史
	var messages []ChatMessage
	err = cs.db.Where("session_id = ?", session.ID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return nil, fmt.Errorf("获取对话历史失败: %w", err)
	}

	// 转换为AI服务需要的格式
	apiMessages := make([]APIMessage, len(messages))
	for i, msg := range messages {
		apiMessages[i] = APIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 调用AI服务
	modelName := req.ModelName
	if modelName == "" {
		modelName = session.ModelName
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("准备调用AI服务",
			zap.String("modelName", modelName),
			zap.Int("messageCount", len(apiMessages)),
		)

		// 打印所有对话历史
		for i, msg := range apiMessages {
			global.GVA_LOG.Info("对话历史",
				zap.Int("index", i),
				zap.String("role", msg.Role),
				zap.String("content", msg.Content),
			)
		}
	}

	aiService := GetAIServiceWithProtection(modelName)
	if aiService == nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("AI服务不可用", zap.String("modelName", modelName))
		}
		return nil, fmt.Errorf("AI服务不可用: %s", modelName)
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("成功获取AI服务", zap.String("modelName", modelName))
	}

	aiConfig := &AIConfig{
		MaxTokens:   2000,
		Temperature: 0.7,
		TopP:        0.9,
		Model:       modelName,
		Stream:      false,
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("AI配置",
			zap.Int("maxTokens", aiConfig.MaxTokens),
			zap.Float64("temperature", aiConfig.Temperature),
			zap.Float64("topP", aiConfig.TopP),
		)
	}

	aiResponse, err := aiService.ChatCompletion(ctx, apiMessages, aiConfig)
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("AI服务调用失败",
				zap.Error(err),
				zap.String("modelName", modelName),
			)
		}
		return nil, fmt.Errorf("AI服务调用失败: %w", err)
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("AI服务调用成功",
			zap.String("modelName", modelName),
			zap.String("responseText", aiResponse.Text),
			zap.Int("totalTokens", aiResponse.Usage.TotalTokens),
		)
	}

	// 保存AI响应
	assistantMessage := &ChatMessage{
		SessionID:  session.ID,
		Role:       "assistant",
		Content:    aiResponse.Text,
		TokenCount: aiResponse.Usage.CompletionTokens,
		CreatedAt:  time.Now(),
	}

	// 设置metadata为有效JSON格式
	if aiResponse.Usage.TotalTokens > 0 {
		metadata := map[string]interface{}{
			"model":            aiResponse.Model,
			"promptTokens":     aiResponse.Usage.PromptTokens,
			"completionTokens": aiResponse.Usage.CompletionTokens,
			"totalTokens":      aiResponse.Usage.TotalTokens,
		}
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			assistantMessage.Metadata = string(metadataBytes)
		} else {
			assistantMessage.Metadata = "{}"
		}
	} else {
		// 即使没有token信息，也要设置为有效的JSON
		metadata := map[string]interface{}{
			"model": aiResponse.Model,
		}
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			assistantMessage.Metadata = string(metadataBytes)
		} else {
			assistantMessage.Metadata = "{}"
		}
	}

	err = cs.db.Create(assistantMessage).Error
	if err != nil {
		return nil, fmt.Errorf("保存AI响应失败: %w", err)
	}

	// 更新会话时间
	cs.UpdateSession(session.ID, userID, map[string]interface{}{
		"updated_at": time.Now(),
	})

	// 记录用户行为
	if recommendationService := GetRecommendationService(); recommendationService != nil {
		recommendationService.RecordUserBehavior(userID, "chat", "ai_model", session.ID, req.Context)
	}

	elapsedTime := time.Since(start).Milliseconds()

	// 解析元数据
	var metadata map[string]interface{}
	if assistantMessage.Metadata != "" {
		json.Unmarshal([]byte(assistantMessage.Metadata), &metadata)
	}

	return &ChatResponse{
		SessionID:   session.ID,
		MessageID:   assistantMessage.ID,
		Content:     assistantMessage.Content,
		Role:        assistantMessage.Role,
		TokenCount:  assistantMessage.TokenCount,
		ModelName:   modelName,
		Metadata:    metadata,
		CreatedAt:   assistantMessage.CreatedAt,
		ElapsedTime: elapsedTime,
	}, nil
}

// GetMessages 获取会话消息
func (cs *ChatService) GetMessages(sessionID uint, userID uint, page, pageSize int) ([]ChatMessage, int64, error) {
	// 验证会话所有权
	var session ChatSession
	err := cs.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		return nil, 0, fmt.Errorf("会话不存在或无权限访问")
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50
	}

	var total int64
	err = cs.db.Model(&ChatMessage{}).Where("session_id = ?", sessionID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var messages []ChatMessage
	err = cs.db.Where("session_id = ?", sessionID).
		Order("created_at ASC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&messages).Error
	if err != nil {
		return nil, 0, err
	}

	return messages, total, nil
}

// ClearMessages 清空会话消息
func (cs *ChatService) ClearMessages(sessionID uint, userID uint) error {
	// 验证会话所有权
	var session ChatSession
	err := cs.db.Where("id = ? AND user_id = ?", sessionID, userID).First(&session).Error
	if err != nil {
		return fmt.Errorf("会话不存在或无权限访问")
	}

	// 删除所有消息
	return cs.db.Where("session_id = ?", sessionID).Delete(&ChatMessage{}).Error
}

// GetChatStats 获取聊天统计
func (cs *ChatService) GetChatStats(userID uint, days int) (map[string]interface{}, error) {
	if days <= 0 {
		days = 30
	}

	startTime := time.Now().AddDate(0, 0, -days)

	// 会话统计
	var sessionCount int64
	err := cs.db.Model(&ChatSession{}).Where("user_id = ? AND created_at >= ?", userID, startTime).Count(&sessionCount).Error
	if err != nil {
		return nil, err
	}

	// 消息统计
	var messageCount int64
	err = cs.db.Model(&ChatMessage{}).
		Joins("JOIN chat_sessions ON chat_messages.session_id = chat_sessions.id").
		Where("chat_sessions.user_id = ? AND chat_messages.created_at >= ?", userID, startTime).
		Count(&messageCount).Error
	if err != nil {
		return nil, err
	}

	// Token统计
	var tokenSum int64
	err = cs.db.Model(&ChatMessage{}).
		Joins("JOIN chat_sessions ON chat_messages.session_id = chat_sessions.id").
		Where("chat_sessions.user_id = ? AND chat_messages.created_at >= ?", userID, startTime).
		Select("COALESCE(SUM(token_count), 0)").
		Scan(&tokenSum).Error
	if err != nil {
		return nil, err
	}

	// 活跃会话统计
	var activeSessionCount int64
	err = cs.db.Model(&ChatSession{}).Where("user_id = ? AND is_active = ? AND updated_at >= ?", userID, true, startTime).Count(&activeSessionCount).Error
	if err != nil {
		return nil, err
	}

	// 模型使用统计
	var modelStats []struct {
		ModelName string `json:"model_name"`
		Count     int64  `json:"count"`
	}
	err = cs.db.Model(&ChatSession{}).
		Where("user_id = ? AND created_at >= ?", userID, startTime).
		Group("model_name").
		Select("model_name, COUNT(*) as count").
		Find(&modelStats).Error
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"period":             fmt.Sprintf("%d days", days),
		"sessionCount":       sessionCount,
		"messageCount":       messageCount,
		"tokenCount":         tokenSum,
		"activeSessionCount": activeSessionCount,
		"modelStats":         modelStats,
		"averageTokensPerMessage": func() float64 {
			if messageCount > 0 {
				return float64(tokenSum) / float64(messageCount)
			}
			return 0
		}(),
	}, nil
}

// ExportChatHistory 导出聊天历史
func (cs *ChatService) ExportChatHistory(sessionID uint, userID uint, format string) ([]byte, error) {
	session, err := cs.GetSession(sessionID, userID)
	if err != nil {
		return nil, fmt.Errorf("获取会话失败: %w", err)
	}

	switch format {
	case "json":
		return json.MarshalIndent(session, "", "  ")
	case "txt":
		var content string
		content += fmt.Sprintf("会话标题: %s\n", session.Title)
		content += fmt.Sprintf("创建时间: %s\n", session.CreatedAt.Format("2006-01-02 15:04:05"))
		content += fmt.Sprintf("模型: %s\n", session.ModelName)
		content += "\n对话记录:\n"
		content += strings.Repeat("=", 50) + "\n"

		for _, msg := range session.Messages {
			content += fmt.Sprintf("\n[%s] %s:\n%s\n",
				msg.CreatedAt.Format("15:04:05"),
				msg.Role,
				msg.Content)
		}

		return []byte(content), nil
	default:
		return nil, fmt.Errorf("不支持的导出格式: %s", format)
	}
}

// SendMessageStream 流式发送消息
func (cs *ChatService) SendMessageStream(ctx context.Context, userID uint, req ChatRequest, callback func(string, bool, int)) error {
	start := time.Now()

	// 如果没有指定会话，创建新会话
	var session *ChatSession
	var err error

	if req.SessionID == nil {
		session, err = cs.CreateSession(userID, req.ProjectID, "", req.ModelName)
		if err != nil {
			return fmt.Errorf("创建会话失败: %w", err)
		}
	} else {
		session, err = cs.GetSession(*req.SessionID, userID)
		if err != nil {
			return fmt.Errorf("获取会话失败: %w", err)
		}
	}

	// 保存用户消息
	userMessage := &ChatMessage{
		SessionID: session.ID,
		Role:      "user",
		Content:   req.Message,
		CreatedAt: time.Now(),
	}

	// 设置metadata为有效JSON格式
	if req.Context != nil && len(req.Context) > 0 {
		if metadataBytes, err := json.Marshal(req.Context); err == nil {
			userMessage.Metadata = string(metadataBytes)
		} else {
			userMessage.Metadata = "{}"
		}
	} else {
		userMessage.Metadata = "{}" // 设置为空JSON对象而不是空字符串
	}

	err = cs.db.Create(userMessage).Error
	if err != nil {
		return fmt.Errorf("保存用户消息失败: %w", err)
	}

	// 准备对话历史
	var messages []ChatMessage
	err = cs.db.Where("session_id = ?", session.ID).Order("created_at ASC").Find(&messages).Error
	if err != nil {
		return fmt.Errorf("获取对话历史失败: %w", err)
	}

	// 转换为AI服务需要的格式
	apiMessages := make([]APIMessage, len(messages))
	for i, msg := range messages {
		apiMessages[i] = APIMessage{
			Role:    msg.Role,
			Content: msg.Content,
		}
	}

	// 调用AI服务
	modelName := req.ModelName
	if modelName == "" {
		modelName = session.ModelName
	}

	aiService := GetAIServiceWithProtection(modelName)
	if aiService == nil {
		return fmt.Errorf("AI服务不可用: %s", modelName)
	}

	aiConfig := &AIConfig{
		MaxTokens:   2000,
		Temperature: 0.7,
		TopP:        0.9,
		Stream:      true,
	}

	// 使用流式响应
	responseContent := ""
	tokenCount := 0

	err = cs.streamChatCompletion(ctx, aiService, apiMessages, aiConfig, func(chunk string, isComplete bool, tokens int) {
		responseContent += chunk
		tokenCount = tokens
		callback(chunk, isComplete, tokens)
	})

	if err != nil {
		return fmt.Errorf("AI服务调用失败: %w", err)
	}

	// 保存AI响应
	assistantMessage := &ChatMessage{
		SessionID:  session.ID,
		Role:       "assistant",
		Content:    responseContent,
		TokenCount: tokenCount,
		CreatedAt:  time.Now(),
	}

	// 设置metadata为有效JSON格式
	if tokenCount > 0 {
		metadata := map[string]interface{}{
			"model":       modelName,
			"totalTokens": tokenCount,
			"elapsedTime": time.Since(start).Milliseconds(),
		}
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			assistantMessage.Metadata = string(metadataBytes)
		} else {
			assistantMessage.Metadata = "{}"
		}
	} else {
		// 即使没有token信息，也要设置为有效的JSON
		metadata := map[string]interface{}{
			"model": modelName,
		}
		if metadataBytes, err := json.Marshal(metadata); err == nil {
			assistantMessage.Metadata = string(metadataBytes)
		} else {
			assistantMessage.Metadata = "{}"
		}
	}

	err = cs.db.Create(assistantMessage).Error
	if err != nil {
		return fmt.Errorf("保存AI响应失败: %w", err)
	}

	// 更新会话时间
	cs.UpdateSession(session.ID, userID, map[string]interface{}{
		"updated_at": time.Now(),
	})

	// 记录用户行为
	if recommendationService := GetRecommendationService(); recommendationService != nil {
		recommendationService.RecordUserBehavior(userID, "chat", "ai_model", session.ID, req.Context)
	}

	return nil
}

// streamChatCompletion 流式聊天完成
func (cs *ChatService) streamChatCompletion(ctx context.Context, service AIService, messages []APIMessage, config *AIConfig, callback func(string, bool, int)) error {
	// 添加详细日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("开始流式聊天调用",
			zap.String("serviceType", fmt.Sprintf("%T", service)),
			zap.Bool("configStream", config.Stream),
		)
	}

	// 尝试使用真正的流式调用
	if streamService, ok := service.(StreamAIService); ok {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Info("使用真正的流式AI服务")
		}

		return streamService.ChatCompletionStream(ctx, messages, config, callback)
	}

	// 降级到模拟流式（为兼容性保留）
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("流式接口不可用，使用模拟流式AI服务")
	}

	aiResponse, err := service.ChatCompletion(ctx, messages, config)
	if err != nil {
		return err
	}

	// 模拟流式输出
	content := aiResponse.Text
	if content == "" && len(aiResponse.Choices) > 0 {
		content = aiResponse.Choices[0].Message.Content
	}

	// 按字符分块输出
	chunkSize := 3
	for i := 0; i < len(content); i += chunkSize {
		end := i + chunkSize
		if end > len(content) {
			end = len(content)
		}

		chunk := content[i:end]
		callback(chunk, false, 0)

		// 添加小延迟模拟真实的流式效果
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(30 * time.Millisecond):
		}
	}

	// 发送完成信号
	callback("", true, aiResponse.Usage.TotalTokens)
	return nil
}

// 全局聊天服务实例
var GlobalChatService *ChatService

// InitChatService 初始化聊天服务
func InitChatService() {
	GlobalChatService = NewChatService()

	// 自动迁移数据库表
	err := global.GVA_DB.AutoMigrate(&ChatSession{}, &ChatMessage{})
	if err != nil {
		global.GVA_LOG.Error("Failed to migrate chat tables", zap.Error(err))
	} else {
		global.GVA_LOG.Info("Chat service initialized successfully")
	}
}

// GetChatService 获取聊天服务
func GetChatService() *ChatService {
	if GlobalChatService == nil {
		InitChatService()
	}
	return GlobalChatService
}
