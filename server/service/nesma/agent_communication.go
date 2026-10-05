package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MessagePriority 消息优先级
type MessagePriority int

const (
	PriorityLow MessagePriority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// MessageStatus 消息状态
type MessageStatus string

const (
	MessageStatusPending   MessageStatus = "pending"
	MessageStatusSent      MessageStatus = "sent"
	MessageStatusDelivered MessageStatus = "delivered"
	MessageStatusRead      MessageStatus = "read"
	MessageStatusFailed    MessageStatus = "failed"
	MessageStatusExpired   MessageStatus = "expired"
)

// AgentMessage 增强的Agent消息结构
type AgentMessage struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	From         string                 `json:"from"`
	To           string                 `json:"to"`
	SessionID    string                 `json:"sessionId"`
	Content      map[string]interface{} `json:"content"`
	Priority     MessagePriority        `json:"priority"`
	Status       MessageStatus          `json:"status"`
	CreatedAt    time.Time              `json:"createdAt"`
	ExpiresAt    *time.Time             `json:"expiresAt"`
	RetryCount   int                    `json:"retryCount"`
	MaxRetries   int                    `json:"maxRetries"`
	ResponseToID string                 `json:"responseToId"`
	Metadata     map[string]interface{} `json:"metadata"`
	ErrorMessage string                 `json:"errorMessage"`
	DeliveredAt  *time.Time             `json:"deliveredAt"`
	ReadAt       *time.Time             `json:"readAt"`
}

// MessageRouter 消息路由器
type MessageRouter struct {
	routes map[string]string // agentID -> endpoint mapping
	mu     sync.RWMutex
}

// NewMessageRouter 创建消息路由器
func NewMessageRouter() *MessageRouter {
	return &MessageRouter{
		routes: make(map[string]string),
	}
}

// RegisterRoute 注册路由
func (r *MessageRouter) RegisterRoute(agentID, endpoint string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes[agentID] = endpoint
}

// UnregisterRoute 取消注册路由
func (r *MessageRouter) UnregisterRoute(agentID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.routes, agentID)
}

// GetRoute 获取路由
func (r *MessageRouter) GetRoute(agentID string) (string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	endpoint, exists := r.routes[agentID]
	return endpoint, exists
}

// MessageQueue 消息队列
type MessageQueue struct {
	messages []*AgentMessage
	mu       sync.RWMutex
	cond     *sync.Cond
}

// NewMessageQueue 创建消息队列
func NewMessageQueue() *MessageQueue {
	mq := &MessageQueue{
		messages: make([]*AgentMessage, 0),
	}
	mq.cond = sync.NewCond(&mq.mu)
	return mq
}

// Push 推送消息
func (mq *MessageQueue) Push(msg *AgentMessage) {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// 插入消息并按优先级排序
	inserted := false
	for i, existingMsg := range mq.messages {
		if msg.Priority > existingMsg.Priority {
			mq.messages = append(mq.messages[:i], append([]*AgentMessage{msg}, mq.messages[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		mq.messages = append(mq.messages, msg)
	}

	mq.cond.Signal()
}

// Pop 弹出消息
func (mq *MessageQueue) Pop() *AgentMessage {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	for len(mq.messages) == 0 {
		mq.cond.Wait()
	}

	msg := mq.messages[0]
	mq.messages = mq.messages[1:]
	return msg
}

// PopWithTimeout 带超时的弹出消息
func (mq *MessageQueue) PopWithTimeout(timeout time.Duration) *AgentMessage {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if len(mq.messages) > 0 {
		msg := mq.messages[0]
		mq.messages = mq.messages[1:]
		return msg
	}

	// 创建超时channel
	timeoutChan := time.After(timeout)
	done := make(chan struct{})

	go func() {
		mq.cond.Wait()
		close(done)
	}()

	select {
	case <-done:
		if len(mq.messages) > 0 {
			msg := mq.messages[0]
			mq.messages = mq.messages[1:]
			return msg
		}
		return nil
	case <-timeoutChan:
		return nil
	}
}

// Size 获取队列大小
func (mq *MessageQueue) Size() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return len(mq.messages)
}

// CommunicationManager 通信管理器
type CommunicationManager struct {
	router          *MessageRouter
	messageQueue    *MessageQueue
	messageHandlers map[string]func(*AgentMessage) error
	retryQueue      *MessageQueue
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	logger          *zap.Logger
}

// NewCommunicationManager 创建通信管理器
func NewCommunicationManager(logger *zap.Logger) *CommunicationManager {
	ctx, cancel := context.WithCancel(context.Background())

	cm := &CommunicationManager{
		router:          NewMessageRouter(),
		messageQueue:    NewMessageQueue(),
		messageHandlers: make(map[string]func(*AgentMessage) error),
		retryQueue:      NewMessageQueue(),
		ctx:             ctx,
		cancel:          cancel,
		logger:          logger,
	}

	// 启动消息处理器
	go cm.processMessages()
	go cm.processRetryMessages()

	return cm
}

// RegisterMessageHandler 注册消息处理器
func (cm *CommunicationManager) RegisterMessageHandler(messageType string, handler func(*AgentMessage) error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.messageHandlers[messageType] = handler
}

// RegisterAgent 注册Agent
func (cm *CommunicationManager) RegisterAgent(agentID, endpoint string) {
	cm.router.RegisterRoute(agentID, endpoint)
	if cm.logger != nil {
		cm.logger.Info("Agent registered", zap.String("agentId", agentID), zap.String("endpoint", endpoint))
	}
}

// UnregisterAgent 取消注册Agent
func (cm *CommunicationManager) UnregisterAgent(agentID string) {
	cm.router.UnregisterRoute(agentID)
	if cm.logger != nil {
		cm.logger.Info("Agent unregistered", zap.String("agentId", agentID))
	}
}

// SendMessage 发送消息
func (cm *CommunicationManager) SendMessage(ctx context.Context, msg *AgentMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}

	msg.CreatedAt = time.Now()
	msg.Status = MessageStatusPending

	// 检查消息是否过期
	if msg.ExpiresAt != nil && time.Now().After(*msg.ExpiresAt) {
		msg.Status = MessageStatusExpired
		return errors.New("message expired")
	}

	// 保存消息到数据库
	if err := cm.saveMessage(msg); err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}

	// 推送到消息队列
	cm.messageQueue.Push(msg)

	return nil
}

// BroadcastMessage 广播消息
func (cm *CommunicationManager) BroadcastMessage(ctx context.Context, msg *AgentMessage) error {
	// 获取所有注册的Agent
	cm.router.mu.RLock()
	agents := make([]string, 0, len(cm.router.routes))
	for agentID := range cm.router.routes {
		agents = append(agents, agentID)
	}
	cm.router.mu.RUnlock()

	// 为每个Agent创建消息副本
	for _, agentID := range agents {
		if agentID != msg.From { // 不发送给自己
			msgCopy := *msg
			msgCopy.ID = uuid.New().String()
			msgCopy.To = agentID

			if err := cm.SendMessage(ctx, &msgCopy); err != nil {
				if cm.logger != nil {
					cm.logger.Error("Failed to broadcast message",
						zap.String("messageId", msgCopy.ID),
						zap.String("to", agentID),
						zap.Error(err))
				}
			}
		}
	}

	return nil
}

// processMessages 处理消息
func (cm *CommunicationManager) processMessages() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
			msg := cm.messageQueue.PopWithTimeout(5 * time.Second)
			if msg == nil {
				continue
			}

			if err := cm.deliverMessage(msg); err != nil {
				if cm.logger != nil {
					cm.logger.Error("Failed to deliver message",
						zap.String("messageId", msg.ID),
						zap.Error(err))
				}

				// 重试逻辑
				if msg.RetryCount < msg.MaxRetries {
					msg.RetryCount++
					msg.Status = MessageStatusPending
					time.Sleep(time.Duration(msg.RetryCount) * time.Second) // 指数退避
					cm.retryQueue.Push(msg)
				} else {
					msg.Status = MessageStatusFailed
					msg.ErrorMessage = err.Error()
					cm.updateMessageStatus(msg)
				}
			}
		}
	}
}

// processRetryMessages 处理重试消息
func (cm *CommunicationManager) processRetryMessages() {
	for {
		select {
		case <-cm.ctx.Done():
			return
		default:
			msg := cm.retryQueue.PopWithTimeout(10 * time.Second)
			if msg == nil {
				continue
			}

			if err := cm.deliverMessage(msg); err != nil {
				if cm.logger != nil {
					cm.logger.Error("Failed to deliver retry message",
						zap.String("messageId", msg.ID),
						zap.Error(err))
				}

				if msg.RetryCount < msg.MaxRetries {
					msg.RetryCount++
					time.Sleep(time.Duration(msg.RetryCount*2) * time.Second) // 指数退避
					cm.retryQueue.Push(msg)
				} else {
					msg.Status = MessageStatusFailed
					msg.ErrorMessage = err.Error()
					cm.updateMessageStatus(msg)
				}
			}
		}
	}
}

// deliverMessage 投递消息
func (cm *CommunicationManager) deliverMessage(msg *AgentMessage) error {
	// 检查消息是否过期
	if msg.ExpiresAt != nil && time.Now().After(*msg.ExpiresAt) {
		msg.Status = MessageStatusExpired
		cm.updateMessageStatus(msg)
		return errors.New("message expired")
	}

	// 获取目标Agent的路由
	endpoint, exists := cm.router.GetRoute(msg.To)
	if !exists {
		return fmt.Errorf("agent %s not found", msg.To)
	}

	// 检查是否有对应的消息处理器
	cm.mu.RLock()
	handler, exists := cm.messageHandlers[msg.Type]
	cm.mu.RUnlock()

	if exists {
		// 使用注册的处理器处理消息
		if err := handler(msg); err != nil {
			return err
		}
	} else {
		// 默认处理逻辑（可以是HTTP调用、gRPC调用等）
		if cm.logger != nil {
			cm.logger.Info("Delivering message",
				zap.String("messageId", msg.ID),
				zap.String("to", msg.To),
				zap.String("endpoint", endpoint))
		}
	}

	// 更新消息状态
	msg.Status = MessageStatusDelivered
	now := time.Now()
	msg.DeliveredAt = &now

	return cm.updateMessageStatus(msg)
}

// saveMessage 保存消息到数据库
func (cm *CommunicationManager) saveMessage(msg *AgentMessage) error {
	// 转换为数据库模型
	dbMsg := &nesma.NesmaAgentMessage{
		MessageID:     msg.ID,
		SessionID:     msg.SessionID,
		SourceAgentID: msg.From,
		TargetAgentID: msg.To,
		MessageType:   msg.Type,
		ContentType:   "json",
		Status:        string(msg.Status),
		Priority:      int(msg.Priority),
		ResponseToID:  msg.ResponseToID,
		ExpiresAt:     msg.ExpiresAt,
	}

	// 序列化内容
	if contentBytes, err := json.Marshal(msg.Content); err == nil {
		dbMsg.Content = contentBytes
	}

	// 序列化元数据
	if metadataBytes, err := json.Marshal(msg.Metadata); err == nil {
		dbMsg.Metadata = metadataBytes
	}

	return global.GVA_DB.Create(dbMsg).Error
}

// updateMessageStatus 更新消息状态
func (cm *CommunicationManager) updateMessageStatus(msg *AgentMessage) error {
	updates := map[string]interface{}{
		"status":        string(msg.Status),
		"retry_count":   msg.RetryCount,
		"error_message": msg.ErrorMessage,
		"delivered_at":  msg.DeliveredAt,
		"read_at":       msg.ReadAt,
	}

	return global.GVA_DB.Model(&nesma.NesmaAgentMessage{}).
		Where("message_id = ?", msg.ID).
		Updates(updates).Error
}

// GetMessageHistory 获取消息历史
func (cm *CommunicationManager) GetMessageHistory(ctx context.Context, sessionID string, limit int) ([]*AgentMessage, error) {
	var dbMessages []nesma.NesmaAgentMessage

	err := global.GVA_DB.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&dbMessages).Error

	if err != nil {
		return nil, err
	}

	messages := make([]*AgentMessage, 0, len(dbMessages))
	for _, dbMsg := range dbMessages {
		msg := &AgentMessage{
			ID:           dbMsg.MessageID,
			Type:         dbMsg.MessageType,
			From:         dbMsg.SourceAgentID,
			To:           dbMsg.TargetAgentID,
			SessionID:    dbMsg.SessionID,
			Priority:     MessagePriority(dbMsg.Priority),
			Status:       MessageStatus(dbMsg.Status),
			CreatedAt:    dbMsg.CreatedAt,
			ExpiresAt:    dbMsg.ExpiresAt,
			ResponseToID: dbMsg.ResponseToID,
			DeliveredAt:  dbMsg.DeliveredAt,
			ReadAt:       dbMsg.ReadAt,
		}

		// 反序列化内容
		if len(dbMsg.Content) > 0 {
			json.Unmarshal(dbMsg.Content, &msg.Content)
		}

		// 反序列化元数据
		if len(dbMsg.Metadata) > 0 {
			json.Unmarshal(dbMsg.Metadata, &msg.Metadata)
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// Stop 停止通信管理器
func (cm *CommunicationManager) Stop() {
	cm.cancel()
}

// GetQueueStats 获取队列统计信息
func (cm *CommunicationManager) GetQueueStats() map[string]interface{} {
	return map[string]interface{}{
		"message_queue_size": cm.messageQueue.Size(),
		"retry_queue_size":   cm.retryQueue.Size(),
		"registered_agents":  len(cm.router.routes),
	}
}

// 全局通信管理器
var GlobalCommunicationManager *CommunicationManager

// InitCommunicationManager 初始化通信管理器
func InitCommunicationManager() {
	logger := global.GVA_LOG
	if logger == nil {
		logger = zap.NewNop()
	}

	GlobalCommunicationManager = NewCommunicationManager(logger)

	// 注册默认消息处理器
	GlobalCommunicationManager.RegisterMessageHandler("ping", func(msg *AgentMessage) error {
		// Ping消息处理
		return nil
	})

	GlobalCommunicationManager.RegisterMessageHandler("heartbeat", func(msg *AgentMessage) error {
		// 心跳消息处理
		return nil
	})
}

// GetCommunicationManager 获取通信管理器
func GetCommunicationManager() *CommunicationManager {
	if GlobalCommunicationManager == nil {
		InitCommunicationManager()
	}
	return GlobalCommunicationManager
}
