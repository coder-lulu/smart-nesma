package agent

import (
	"context"
	"time"
)

// AgentType Agent类型枚举
type AgentType string

const (
	AgentTypeRequirementRefine AgentType = "requirement_refine" // 需求细化Agent
	AgentTypeNesmaEvaluation   AgentType = "nesma_evaluation"   // NESMA评估Agent
	AgentTypeDocumentGenerate  AgentType = "document_generate"  // 文档生成Agent
	AgentTypeMemory            AgentType = "memory"             // 记忆Agent
	AgentTypeKnowledge         AgentType = "knowledge"          // 知识库Agent
)

// MessageType 消息类型
type MessageType string

const (
	MessageTypeRequest   MessageType = "request"   // 请求消息
	MessageTypeResponse  MessageType = "response"  // 响应消息
	MessageTypeNotify    MessageType = "notify"    // 通知消息
	MessageTypeHeartbeat MessageType = "heartbeat" // 心跳消息
	MessageTypeError     MessageType = "error"     // 错误消息
)

// TaskStatus 任务状态
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"    // 待处理
	TaskStatusProcessing TaskStatus = "processing" // 处理中
	TaskStatusCompleted  TaskStatus = "completed"  // 已完成
	TaskStatusFailed     TaskStatus = "failed"     // 失败
	TaskStatusCancelled  TaskStatus = "cancelled"  // 已取消
)

// Message Agent间通信消息
type Message struct {
	ID        string                 `json:"id"`        // 消息ID
	Type      MessageType            `json:"type"`      // 消息类型
	From      string                 `json:"from"`      // 发送方Agent ID
	To        string                 `json:"to"`        // 接收方Agent ID
	SessionID string                 `json:"sessionId"` // 会话ID
	Payload   map[string]interface{} `json:"payload"`   // 消息内容
	Timestamp time.Time              `json:"timestamp"` // 时间戳
	Timeout   time.Duration          `json:"timeout"`   // 超时时间
	ReplyTo   string                 `json:"replyTo"`   // 回复消息ID
	Priority  int                    `json:"priority"`  // 优先级
}

// Task Agent任务
type Task struct {
	ID           string                 `json:"id"`           // 任务ID
	Type         string                 `json:"type"`         // 任务类型
	Status       TaskStatus             `json:"status"`       // 任务状态
	AgentID      string                 `json:"agentId"`      // 负责的Agent ID
	SessionID    string                 `json:"sessionId"`    // 会话ID
	Input        map[string]interface{} `json:"input"`        // 输入参数
	Output       map[string]interface{} `json:"output"`       // 输出结果
	Error        string                 `json:"error"`        // 错误信息
	CreatedAt    time.Time              `json:"createdAt"`    // 创建时间
	StartedAt    *time.Time             `json:"startedAt"`    // 开始时间
	CompletedAt  *time.Time             `json:"completedAt"`  // 完成时间
	Dependencies []string               `json:"dependencies"` // 依赖的任务ID
	Timeout      time.Duration          `json:"timeout"`      // 超时时间
	Retries      int                    `json:"retries"`      // 重试次数
	MaxRetries   int                    `json:"maxRetries"`   // 最大重试次数
}

// AgentInfo Agent信息
type AgentInfo struct {
	ID           string                 `json:"id"`           // Agent ID
	Type         AgentType              `json:"type"`         // Agent类型
	Name         string                 `json:"name"`         // Agent名称
	Description  string                 `json:"description"`  // Agent描述
	Version      string                 `json:"version"`      // 版本号
	Status       string                 `json:"status"`       // 状态
	Capabilities []string               `json:"capabilities"` // 能力列表
	Config       map[string]interface{} `json:"config"`       // 配置信息
	Health       HealthStatus           `json:"health"`       // 健康状态
	RegisteredAt time.Time              `json:"registeredAt"` // 注册时间
	LastSeen     time.Time              `json:"lastSeen"`     // 最后活跃时间
}

// HealthStatus 健康状态
type HealthStatus struct {
	Status       string    `json:"status"`       // healthy, unhealthy, unknown
	LastCheck    time.Time `json:"lastCheck"`    // 最后检查时间
	Details      string    `json:"details"`      // 详细信息
	ResponseTime int64     `json:"responseTime"` // 响应时间(ms)
}

// Agent 接口定义
type Agent interface {
	// 基础方法
	GetInfo() AgentInfo
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	Health() HealthStatus

	// 消息处理
	HandleMessage(ctx context.Context, msg *Message) (*Message, error)

	// 任务处理
	ProcessTask(ctx context.Context, task *Task) error

	// 配置更新
	UpdateConfig(config map[string]interface{}) error
}

// MessageHandler 消息处理器接口
type MessageHandler interface {
	CanHandle(msgType MessageType) bool
	Handle(ctx context.Context, msg *Message) (*Message, error)
}

// TaskProcessor 任务处理器接口
type TaskProcessor interface {
	CanProcess(taskType string) bool
	Process(ctx context.Context, task *Task) error
}
