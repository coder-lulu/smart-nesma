package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaAgent Agent配置表
type NesmaAgent struct {
	global.GVA_MODEL
	AgentID             string         `json:"agentId" gorm:"type:varchar(100);not null;unique;comment:Agent唯一标识"`
	Name                string         `json:"name" gorm:"type:varchar(200);not null;comment:Agent名称"`
	Description         string         `json:"description" gorm:"type:text;comment:Agent描述"`
	AgentType           string         `json:"agentType" gorm:"type:varchar(50);not null;comment:Agent类型:REQUIREMENT_ANALYSIS,NESMA_EVALUATION,KNOWLEDGE_RETRIEVAL,DOCUMENT_GENERATION,COORDINATION"`
	Capabilities        datatypes.JSON `json:"capabilities" gorm:"type:json;comment:Agent能力配置"`
	Status              string         `json:"status" gorm:"type:varchar(50);default:'active';comment:Agent状态:active,inactive,busy,error"`
	Version             string         `json:"version" gorm:"type:varchar(50);default:'1.0.0';comment:Agent版本"`
	Configuration       datatypes.JSON `json:"configuration" gorm:"type:json;comment:Agent配置参数"`
	MaxConcurrency      int            `json:"maxConcurrency" gorm:"default:3;comment:最大并发处理数"`
	LastHeartbeat       time.Time      `json:"lastHeartbeat" gorm:"comment:最后心跳时间"`
	ProcessingCount     int            `json:"processingCount" gorm:"default:0;comment:当前处理中的任务数"`
	TotalProcessed      int            `json:"totalProcessed" gorm:"default:0;comment:总处理任务数"`
	AverageResponseTime int            `json:"averageResponseTime" gorm:"default:0;comment:平均响应时间(毫秒)"`
}

// TableName 自定义表名
func (NesmaAgent) TableName() string {
	return "nesma_agents"
}

// NesmaAgentTask Agent任务表
type NesmaAgentTask struct {
	global.GVA_MODEL
	TaskID         string         `json:"taskId" gorm:"type:varchar(100);not null;unique;comment:任务唯一标识"`
	AgentID        string         `json:"agentId" gorm:"type:varchar(100);not null;comment:负责的Agent ID"`
	TaskType       string         `json:"taskType" gorm:"type:varchar(50);not null;comment:任务类型"`
	Priority       int            `json:"priority" gorm:"default:1;comment:任务优先级:1-低,2-中,3-高,4-紧急"`
	Status         string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:任务状态:pending,processing,completed,failed,cancelled"`
	InputData      datatypes.JSON `json:"inputData" gorm:"type:json;comment:任务输入数据"`
	OutputData     datatypes.JSON `json:"outputData" gorm:"type:json;comment:任务输出数据"`
	ErrorMessage   string         `json:"errorMessage" gorm:"type:text;comment:错误信息"`
	SessionID      string         `json:"sessionId" gorm:"type:varchar(255);comment:会话ID"`
	UserID         uint           `json:"userId" gorm:"comment:用户ID"`
	ProjectID      uint           `json:"projectId" gorm:"comment:项目ID"`
	ParentTaskID   string         `json:"parentTaskId" gorm:"type:varchar(100);comment:父任务ID"`
	RetryCount     int            `json:"retryCount" gorm:"default:0;comment:重试次数"`
	MaxRetries     int            `json:"maxRetries" gorm:"default:3;comment:最大重试次数"`
	StartTime      *time.Time     `json:"startTime" gorm:"comment:开始时间"`
	EndTime        *time.Time     `json:"endTime" gorm:"comment:结束时间"`
	ProcessingTime int            `json:"processingTime" gorm:"comment:处理时间(毫秒)"`
	Metadata       datatypes.JSON `json:"metadata" gorm:"type:json;comment:任务元数据"`
}

// TableName 自定义表名
func (NesmaAgentTask) TableName() string {
	return "nesma_agent_tasks"
}

// NesmaAgentMessage Agent消息表
type NesmaAgentMessage struct {
	global.GVA_MODEL
	MessageID     string         `json:"messageId" gorm:"type:varchar(100);not null;unique;comment:消息唯一标识"`
	SessionID     string         `json:"sessionId" gorm:"type:varchar(255);not null;comment:会话ID"`
	SourceAgentID string         `json:"sourceAgentId" gorm:"type:varchar(100);not null;comment:源Agent ID"`
	TargetAgentID string         `json:"targetAgentId" gorm:"type:varchar(100);comment:目标Agent ID"`
	MessageType   string         `json:"messageType" gorm:"type:varchar(50);not null;comment:消息类型:REQUEST,RESPONSE,NOTIFICATION,BROADCAST"`
	ContentType   string         `json:"contentType" gorm:"type:varchar(50);comment:内容类型:text,json,binary"`
	Content       datatypes.JSON `json:"content" gorm:"type:json;not null;comment:消息内容"`
	Status        string         `json:"status" gorm:"type:varchar(50);default:'sent';comment:消息状态:sent,delivered,read,failed"`
	Priority      int            `json:"priority" gorm:"default:1;comment:消息优先级"`
	ResponseToID  string         `json:"responseToId" gorm:"type:varchar(100);comment:响应的消息ID"`
	ExpiresAt     *time.Time     `json:"expiresAt" gorm:"comment:过期时间"`
	Metadata      datatypes.JSON `json:"metadata" gorm:"type:json;comment:消息元数据"`
	DeliveredAt   *time.Time     `json:"deliveredAt" gorm:"comment:送达时间"`
	ReadAt        *time.Time     `json:"readAt" gorm:"comment:阅读时间"`
}

// TableName 自定义表名
func (NesmaAgentMessage) TableName() string {
	return "nesma_agent_messages"
}

// NesmaWorkflow 工作流表
type NesmaWorkflow struct {
	global.GVA_MODEL
	WorkflowID  string         `json:"workflowId" gorm:"type:varchar(100);not null;unique;comment:工作流唯一标识"`
	Name        string         `json:"name" gorm:"type:varchar(200);not null;comment:工作流名称"`
	Description string         `json:"description" gorm:"type:text;comment:工作流描述"`
	Definition  datatypes.JSON `json:"definition" gorm:"type:json;not null;comment:工作流定义(JSON格式)"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'active';comment:工作流状态:active,inactive,draft"`
	Version     string         `json:"version" gorm:"type:varchar(50);default:'1.0.0';comment:工作流版本"`
	Category    string         `json:"category" gorm:"type:varchar(100);comment:工作流分类"`
	Tags        datatypes.JSON `json:"tags" gorm:"type:json;comment:标签"`
	IsTemplate  bool           `json:"isTemplate" gorm:"default:false;comment:是否为模板"`
	UsageCount  int            `json:"usageCount" gorm:"default:0;comment:使用次数"`
	CreatedBy   uint           `json:"createdBy" gorm:"comment:创建者ID"`
}

// TableName 自定义表名
func (NesmaWorkflow) TableName() string {
	return "nesma_workflows"
}

// NesmaWorkflowExecution 工作流执行表
type NesmaWorkflowExecution struct {
	global.GVA_MODEL
	ExecutionID   string         `json:"executionId" gorm:"type:varchar(100);not null;unique;comment:执行唯一标识"`
	WorkflowID    string         `json:"workflowId" gorm:"type:varchar(100);not null;comment:工作流ID"`
	SessionID     string         `json:"sessionId" gorm:"type:varchar(255);not null;comment:会话ID"`
	Status        string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:执行状态:pending,running,completed,failed,cancelled"`
	InputData     datatypes.JSON `json:"inputData" gorm:"type:json;comment:输入数据"`
	OutputData    datatypes.JSON `json:"outputData" gorm:"type:json;comment:输出数据"`
	CurrentStep   string         `json:"currentStep" gorm:"type:varchar(100);comment:当前步骤"`
	ExecutionLog  datatypes.JSON `json:"executionLog" gorm:"type:json;comment:执行日志"`
	ErrorMessage  string         `json:"errorMessage" gorm:"type:text;comment:错误信息"`
	StartTime     *time.Time     `json:"startTime" gorm:"comment:开始时间"`
	EndTime       *time.Time     `json:"endTime" gorm:"comment:结束时间"`
	TotalDuration int            `json:"totalDuration" gorm:"comment:总耗时(毫秒)"`
	UserID        uint           `json:"userId" gorm:"comment:用户ID"`
	ProjectID     uint           `json:"projectId" gorm:"comment:项目ID"`
	Metadata      datatypes.JSON `json:"metadata" gorm:"type:json;comment:执行元数据"`
}

// TableName 自定义表名
func (NesmaWorkflowExecution) TableName() string {
	return "nesma_workflow_executions"
}
