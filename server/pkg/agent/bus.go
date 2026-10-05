package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MessageBus Agent消息总线
type MessageBus struct {
	mu              sync.RWMutex
	agents          map[string]Agent         // 注册的Agent
	messageHandlers map[string]chan *Message // 消息处理通道
	taskQueue       chan *Task               // 任务队列
	messageQueue    chan *Message            // 消息队列
	taskStore       map[string]*Task         // 任务存储
	sessionStore    map[string]*Session      // 会话存储
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup

	// 配置
	maxWorkers        int
	taskTimeout       time.Duration
	messageTimeout    time.Duration
	heartbeatInterval time.Duration
}

// Session 会话信息
type Session struct {
	ID         string                 `json:"id"`
	UserID     uint                   `json:"userId"`
	CreatedAt  time.Time              `json:"createdAt"`
	LastActive time.Time              `json:"lastActive"`
	Context    map[string]interface{} `json:"context"`
	Tasks      []string               `json:"tasks"`
	mu         sync.RWMutex
}

// BusConfig 总线配置
type BusConfig struct {
	MaxWorkers        int           `json:"maxWorkers"`
	TaskTimeout       time.Duration `json:"taskTimeout"`
	MessageTimeout    time.Duration `json:"messageTimeout"`
	HeartbeatInterval time.Duration `json:"heartbeatInterval"`
}

// NewMessageBus 创建消息总线
func NewMessageBus(config BusConfig, logger *zap.Logger) *MessageBus {
	ctx, cancel := context.WithCancel(context.Background())

	return &MessageBus{
		agents:            make(map[string]Agent),
		messageHandlers:   make(map[string]chan *Message),
		taskQueue:         make(chan *Task, 1000),
		messageQueue:      make(chan *Message, 1000),
		taskStore:         make(map[string]*Task),
		sessionStore:      make(map[string]*Session),
		logger:            logger,
		ctx:               ctx,
		cancel:            cancel,
		maxWorkers:        config.MaxWorkers,
		taskTimeout:       config.TaskTimeout,
		messageTimeout:    config.MessageTimeout,
		heartbeatInterval: config.HeartbeatInterval,
	}
}

// Start 启动消息总线
func (bus *MessageBus) Start() error {
	bus.logger.Info("Starting message bus")

	// 启动消息处理工作器
	for i := 0; i < bus.maxWorkers; i++ {
		bus.wg.Add(1)
		go bus.messageWorker()
	}

	// 启动任务处理工作器
	for i := 0; i < bus.maxWorkers; i++ {
		bus.wg.Add(1)
		go bus.taskWorker()
	}

	// 启动心跳检查
	bus.wg.Add(1)
	go bus.heartbeatWorker()

	// 启动清理工作器
	bus.wg.Add(1)
	go bus.cleanupWorker()

	return nil
}

// Stop 停止消息总线
func (bus *MessageBus) Stop() error {
	bus.logger.Info("Stopping message bus")

	bus.cancel()
	bus.wg.Wait()

	return nil
}

// RegisterAgent 注册Agent
func (bus *MessageBus) RegisterAgent(agent Agent) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	info := agent.GetInfo()
	if _, exists := bus.agents[info.ID]; exists {
		return fmt.Errorf("agent %s already registered", info.ID)
	}

	bus.agents[info.ID] = agent
	bus.messageHandlers[info.ID] = make(chan *Message, 100)

	// 启动Agent消息处理协程
	bus.wg.Add(1)
	go bus.agentMessageHandler(info.ID)

	bus.logger.Info("Agent registered", zap.String("id", info.ID), zap.String("type", string(info.Type)))
	return nil
}

// UnregisterAgent 注销Agent
func (bus *MessageBus) UnregisterAgent(agentID string) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	if _, exists := bus.agents[agentID]; !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	delete(bus.agents, agentID)
	if ch, exists := bus.messageHandlers[agentID]; exists {
		close(ch)
		delete(bus.messageHandlers, agentID)
	}

	bus.logger.Info("Agent unregistered", zap.String("id", agentID))
	return nil
}

// SendMessage 发送消息
func (bus *MessageBus) SendMessage(msg *Message) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	msg.Timestamp = time.Now()

	select {
	case bus.messageQueue <- msg:
		return nil
	case <-time.After(bus.messageTimeout):
		return fmt.Errorf("message queue timeout")
	}
}

// SubmitTask 提交任务
func (bus *MessageBus) SubmitTask(task *Task) error {
	if task.ID == "" {
		task.ID = uuid.New().String()
	}
	task.Status = TaskStatusPending
	task.CreatedAt = time.Now()

	bus.mu.Lock()
	bus.taskStore[task.ID] = task
	bus.mu.Unlock()

	select {
	case bus.taskQueue <- task:
		return nil
	case <-time.After(bus.taskTimeout):
		return fmt.Errorf("task queue timeout")
	}
}

// GetTask 获取任务
func (bus *MessageBus) GetTask(taskID string) (*Task, bool) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	task, exists := bus.taskStore[taskID]
	return task, exists
}

// CreateSession 创建会话
func (bus *MessageBus) CreateSession(userID uint) *Session {
	session := &Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		Context:    make(map[string]interface{}),
		Tasks:      make([]string, 0),
	}

	bus.mu.Lock()
	bus.sessionStore[session.ID] = session
	bus.mu.Unlock()

	return session
}

// GetSession 获取会话
func (bus *MessageBus) GetSession(sessionID string) (*Session, bool) {
	bus.mu.RLock()
	defer bus.mu.RUnlock()

	session, exists := bus.sessionStore[sessionID]
	return session, exists
}

// 消息处理工作器
func (bus *MessageBus) messageWorker() {
	defer bus.wg.Done()

	for {
		select {
		case <-bus.ctx.Done():
			return
		case msg := <-bus.messageQueue:
			bus.routeMessage(msg)
		}
	}
}

// 任务处理工作器
func (bus *MessageBus) taskWorker() {
	defer bus.wg.Done()

	for {
		select {
		case <-bus.ctx.Done():
			return
		case task := <-bus.taskQueue:
			bus.processTask(task)
		}
	}
}

// Agent消息处理器
func (bus *MessageBus) agentMessageHandler(agentID string) {
	defer bus.wg.Done()

	bus.mu.RLock()
	ch, exists := bus.messageHandlers[agentID]
	agent, agentExists := bus.agents[agentID]
	bus.mu.RUnlock()

	if !exists || !agentExists {
		return
	}

	for {
		select {
		case <-bus.ctx.Done():
			return
		case msg, ok := <-ch:
			if !ok {
				return
			}

			// 处理消息
			ctx, cancel := context.WithTimeout(bus.ctx, bus.messageTimeout)
			response, err := agent.HandleMessage(ctx, msg)
			cancel()

			if err != nil {
				bus.logger.Error("Agent message handling failed",
					zap.String("agentId", agentID),
					zap.String("messageId", msg.ID),
					zap.Error(err))

				// 发送错误响应
				if msg.ReplyTo != "" {
					errorMsg := &Message{
						ID:        uuid.New().String(),
						Type:      MessageTypeError,
						From:      agentID,
						To:        msg.From,
						SessionID: msg.SessionID,
						Payload:   map[string]interface{}{"error": err.Error()},
						ReplyTo:   msg.ID,
						Timestamp: time.Now(),
					}
					bus.SendMessage(errorMsg)
				}
				continue
			}

			// 发送响应
			if response != nil {
				bus.SendMessage(response)
			}
		}
	}
}

// 路由消息
func (bus *MessageBus) routeMessage(msg *Message) {
	bus.mu.RLock()
	ch, exists := bus.messageHandlers[msg.To]
	bus.mu.RUnlock()

	if !exists {
		bus.logger.Error("Target agent not found", zap.String("to", msg.To))
		return
	}

	select {
	case ch <- msg:
		// 消息已发送
	case <-time.After(bus.messageTimeout):
		bus.logger.Error("Message delivery timeout", zap.String("messageId", msg.ID))
	}
}

// 处理任务
func (bus *MessageBus) processTask(task *Task) {
	bus.mu.RLock()
	agent, exists := bus.agents[task.AgentID]
	bus.mu.RUnlock()

	if !exists {
		task.Status = TaskStatusFailed
		task.Error = fmt.Sprintf("agent %s not found", task.AgentID)
		bus.logger.Error("Task agent not found", zap.String("agentId", task.AgentID))
		return
	}

	// 更新任务状态
	task.Status = TaskStatusProcessing
	now := time.Now()
	task.StartedAt = &now

	// 处理任务
	ctx, cancel := context.WithTimeout(bus.ctx, task.Timeout)
	defer cancel()

	err := agent.ProcessTask(ctx, task)
	completedAt := time.Now()
	task.CompletedAt = &completedAt

	if err != nil {
		task.Status = TaskStatusFailed
		task.Error = err.Error()
		bus.logger.Error("Task processing failed",
			zap.String("taskId", task.ID),
			zap.String("agentId", task.AgentID),
			zap.Error(err))
	} else {
		task.Status = TaskStatusCompleted
		bus.logger.Info("Task completed",
			zap.String("taskId", task.ID),
			zap.String("agentId", task.AgentID))
	}
}

// 心跳检查工作器
func (bus *MessageBus) heartbeatWorker() {
	defer bus.wg.Done()

	ticker := time.NewTicker(bus.heartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bus.ctx.Done():
			return
		case <-ticker.C:
			bus.checkAgentHealth()
		}
	}
}

// 检查Agent健康状态
func (bus *MessageBus) checkAgentHealth() {
	bus.mu.RLock()
	agents := make(map[string]Agent)
	for id, agent := range bus.agents {
		agents[id] = agent
	}
	bus.mu.RUnlock()

	for id, agent := range agents {
		health := agent.Health()
		if health.Status != "healthy" {
			bus.logger.Warn("Agent unhealthy",
				zap.String("agentId", id),
				zap.String("status", health.Status),
				zap.String("details", health.Details))
		}
	}
}

// 清理工作器
func (bus *MessageBus) cleanupWorker() {
	defer bus.wg.Done()

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-bus.ctx.Done():
			return
		case <-ticker.C:
			bus.cleanup()
		}
	}
}

// 清理过期数据
func (bus *MessageBus) cleanup() {
	now := time.Now()
	expiredTasks := make([]string, 0)
	expiredSessions := make([]string, 0)

	bus.mu.RLock()
	// 清理过期任务（24小时前）
	for id, task := range bus.taskStore {
		if task.CompletedAt != nil && now.Sub(*task.CompletedAt) > 24*time.Hour {
			expiredTasks = append(expiredTasks, id)
		}
	}

	// 清理过期会话（1小时未活跃）
	for id, session := range bus.sessionStore {
		if now.Sub(session.LastActive) > time.Hour {
			expiredSessions = append(expiredSessions, id)
		}
	}
	bus.mu.RUnlock()

	// 删除过期数据
	bus.mu.Lock()
	for _, id := range expiredTasks {
		delete(bus.taskStore, id)
	}
	for _, id := range expiredSessions {
		delete(bus.sessionStore, id)
	}
	bus.mu.Unlock()

	if len(expiredTasks) > 0 || len(expiredSessions) > 0 {
		bus.logger.Info("Cleanup completed",
			zap.Int("expiredTasks", len(expiredTasks)),
			zap.Int("expiredSessions", len(expiredSessions)))
	}
}
