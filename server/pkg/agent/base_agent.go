package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// BaseAgent 基础Agent实现
type BaseAgent struct {
	mu              sync.RWMutex
	info            AgentInfo
	messageHandlers map[MessageType]MessageHandler
	taskProcessors  map[string]TaskProcessor
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	isRunning       bool
	lastHealthCheck time.Time
	config          map[string]interface{}
}

// NewBaseAgent 创建基础Agent
func NewBaseAgent(agentType AgentType, name, description string, logger *zap.Logger) *BaseAgent {
	ctx, cancel := context.WithCancel(context.Background())

	agent := &BaseAgent{
		info: AgentInfo{
			ID:           fmt.Sprintf("%s-%s", agentType, uuid.New().String()[:8]),
			Type:         agentType,
			Name:         name,
			Description:  description,
			Version:      "1.0.0",
			Status:       "initialized",
			Capabilities: make([]string, 0),
			Config:       make(map[string]interface{}),
			Health: HealthStatus{
				Status:    "unknown",
				LastCheck: time.Now(),
			},
			RegisteredAt: time.Now(),
			LastSeen:     time.Now(),
		},
		messageHandlers: make(map[MessageType]MessageHandler),
		taskProcessors:  make(map[string]TaskProcessor),
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
		isRunning:       false,
		lastHealthCheck: time.Now(),
		config:          make(map[string]interface{}),
	}

	return agent
}

// GetInfo 获取Agent信息
func (a *BaseAgent) GetInfo() AgentInfo {
	a.mu.RLock()
	defer a.mu.RUnlock()

	a.info.LastSeen = time.Now()
	return a.info
}

// Start 启动Agent
func (a *BaseAgent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.isRunning {
		return fmt.Errorf("agent %s is already running", a.info.ID)
	}

	a.isRunning = true
	a.info.Status = "running"
	a.logger.Info("Agent started", zap.String("id", a.info.ID))

	return nil
}

// Stop 停止Agent
func (a *BaseAgent) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.isRunning {
		return fmt.Errorf("agent %s is not running", a.info.ID)
	}

	a.cancel()
	a.isRunning = false
	a.info.Status = "stopped"
	a.logger.Info("Agent stopped", zap.String("id", a.info.ID))

	return nil
}

// Health 获取健康状态
func (a *BaseAgent) Health() HealthStatus {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	a.lastHealthCheck = now

	status := "healthy"
	details := "Agent is running normally"
	responseTime := int64(1) // 模拟响应时间

	if !a.isRunning {
		status = "unhealthy"
		details = "Agent is not running"
		responseTime = 0
	}

	a.info.Health = HealthStatus{
		Status:       status,
		LastCheck:    now,
		Details:      details,
		ResponseTime: responseTime,
	}

	return a.info.Health
}

// HandleMessage 处理消息
func (a *BaseAgent) HandleMessage(ctx context.Context, msg *Message) (*Message, error) {
	a.mu.RLock()
	handler, exists := a.messageHandlers[msg.Type]
	a.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("no handler for message type %s", msg.Type)
	}

	if !handler.CanHandle(msg.Type) {
		return nil, fmt.Errorf("handler cannot process message type %s", msg.Type)
	}

	a.logger.Debug("Handling message",
		zap.String("agentId", a.info.ID),
		zap.String("messageId", msg.ID),
		zap.String("type", string(msg.Type)))

	return handler.Handle(ctx, msg)
}

// ProcessTask 处理任务
func (a *BaseAgent) ProcessTask(ctx context.Context, task *Task) error {
	a.mu.RLock()
	processor, exists := a.taskProcessors[task.Type]
	a.mu.RUnlock()

	if !exists {
		return fmt.Errorf("no processor for task type %s", task.Type)
	}

	if !processor.CanProcess(task.Type) {
		return fmt.Errorf("processor cannot handle task type %s", task.Type)
	}

	a.logger.Debug("Processing task",
		zap.String("agentId", a.info.ID),
		zap.String("taskId", task.ID),
		zap.String("type", task.Type))

	return processor.Process(ctx, task)
}

// UpdateConfig 更新配置
func (a *BaseAgent) UpdateConfig(config map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	for key, value := range config {
		a.config[key] = value
		a.info.Config[key] = value
	}

	a.logger.Info("Config updated", zap.String("agentId", a.info.ID))
	return nil
}

// RegisterMessageHandler 注册消息处理器
func (a *BaseAgent) RegisterMessageHandler(msgType MessageType, handler MessageHandler) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.messageHandlers[msgType] = handler
	a.logger.Debug("Message handler registered",
		zap.String("agentId", a.info.ID),
		zap.String("messageType", string(msgType)))
}

// RegisterTaskProcessor 注册任务处理器
func (a *BaseAgent) RegisterTaskProcessor(taskType string, processor TaskProcessor) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.taskProcessors[taskType] = processor
	a.logger.Debug("Task processor registered",
		zap.String("agentId", a.info.ID),
		zap.String("taskType", taskType))
}

// AddCapability 添加能力
func (a *BaseAgent) AddCapability(capability string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查是否已存在
	for _, cap := range a.info.Capabilities {
		if cap == capability {
			return
		}
	}

	a.info.Capabilities = append(a.info.Capabilities, capability)
	a.logger.Debug("Capability added",
		zap.String("agentId", a.info.ID),
		zap.String("capability", capability))
}

// GetConfig 获取配置值
func (a *BaseAgent) GetConfig(key string) (interface{}, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	value, exists := a.config[key]
	return value, exists
}

// SetConfig 设置配置值
func (a *BaseAgent) SetConfig(key string, value interface{}) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.config[key] = value
	a.info.Config[key] = value
}

// IsRunning 检查是否运行中
func (a *BaseAgent) IsRunning() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	return a.isRunning
}

// GetLogger 获取日志记录器
func (a *BaseAgent) GetLogger() *zap.Logger {
	return a.logger
}

// GetContext 获取上下文
func (a *BaseAgent) GetContext() context.Context {
	return a.ctx
}

// DefaultMessageHandler 默认消息处理器
type DefaultMessageHandler struct {
	handlerFunc func(ctx context.Context, msg *Message) (*Message, error)
	messageType MessageType
}

// NewDefaultMessageHandler 创建默认消息处理器
func NewDefaultMessageHandler(msgType MessageType, handler func(ctx context.Context, msg *Message) (*Message, error)) *DefaultMessageHandler {
	return &DefaultMessageHandler{
		handlerFunc: handler,
		messageType: msgType,
	}
}

// CanHandle 检查是否可以处理消息
func (h *DefaultMessageHandler) CanHandle(msgType MessageType) bool {
	return h.messageType == msgType
}

// Handle 处理消息
func (h *DefaultMessageHandler) Handle(ctx context.Context, msg *Message) (*Message, error) {
	if h.handlerFunc == nil {
		return nil, fmt.Errorf("no handler function provided")
	}
	return h.handlerFunc(ctx, msg)
}

// DefaultTaskProcessor 默认任务处理器
type DefaultTaskProcessor struct {
	processorFunc func(ctx context.Context, task *Task) error
	taskType      string
}

// NewDefaultTaskProcessor 创建默认任务处理器
func NewDefaultTaskProcessor(taskType string, processor func(ctx context.Context, task *Task) error) *DefaultTaskProcessor {
	return &DefaultTaskProcessor{
		processorFunc: processor,
		taskType:      taskType,
	}
}

// CanProcess 检查是否可以处理任务
func (p *DefaultTaskProcessor) CanProcess(taskType string) bool {
	return p.taskType == taskType
}

// Process 处理任务
func (p *DefaultTaskProcessor) Process(ctx context.Context, task *Task) error {
	if p.processorFunc == nil {
		return fmt.Errorf("no processor function provided")
	}
	return p.processorFunc(ctx, task)
}
