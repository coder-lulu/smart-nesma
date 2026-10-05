package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaVectorMemory 向量记忆表
type NesmaVectorMemory struct {
	global.GVA_MODEL
	UserID         uint           `json:"userId" gorm:"not null;comment:用户ID"`
	ProjectID      *uint          `json:"projectId" gorm:"comment:项目ID"`
	Content        string         `json:"content" gorm:"type:text;not null;comment:记忆内容"`
	Embedding      string         `json:"embedding" gorm:"type:text;comment:向量表示(JSON格式)"`
	Metadata       datatypes.JSON `json:"metadata" gorm:"type:json;comment:元数据"`
	RelevanceScore float64        `json:"relevanceScore" gorm:"default:1.0;comment:相关性评分"`
	MemoryType     string         `json:"memoryType" gorm:"type:varchar(50);comment:记忆类型"`
	ExpiresAt      *time.Time     `json:"expiresAt" gorm:"comment:过期时间"`
}

// TableName 自定义表名
func (NesmaVectorMemory) TableName() string {
	return "nesma_vector_memories"
}

// NesmaCompactMemory 紧凑记忆表
type NesmaCompactMemory struct {
	global.GVA_MODEL
	UserID           uint      `json:"userId" gorm:"not null;comment:用户ID"`
	ProjectID        *uint     `json:"projectId" gorm:"comment:项目ID"`
	Summary          string    `json:"summary" gorm:"type:text;not null;comment:记忆摘要"`
	InteractionCount int       `json:"interactionCount" gorm:"default:0;comment:交互次数"`
	LastUpdated      time.Time `json:"lastUpdated" gorm:"default:CURRENT_TIMESTAMP;comment:最后更新时间"`
	Importance       float64   `json:"importance" gorm:"default:1.0;comment:重要性评分"`
}

// TableName 自定义表名
func (NesmaCompactMemory) TableName() string {
	return "nesma_compact_memories"
}

// NesmaSessionMemory 会话记忆表
type NesmaSessionMemory struct {
	global.GVA_MODEL
	SessionID           string         `json:"sessionId" gorm:"type:varchar(255);not null;comment:会话ID"`
	UserID              uint           `json:"userId" gorm:"not null;comment:用户ID"`
	ConversationHistory datatypes.JSON `json:"conversationHistory" gorm:"type:json;comment:对话历史"`
	ExpiresAt           time.Time      `json:"expiresAt" gorm:"comment:过期时间"`
	IsActive            bool           `json:"isActive" gorm:"default:true;comment:是否激活"`
}

// TableName 自定义表名
func (NesmaSessionMemory) TableName() string {
	return "nesma_session_memories"
}

// NesmaAgentInteraction Agent协作记录表
type NesmaAgentInteraction struct {
	global.GVA_MODEL
	SessionID        string         `json:"sessionId" gorm:"type:varchar(255);not null;comment:会话ID"`
	SourceAgent      string         `json:"sourceAgent" gorm:"type:varchar(100);not null;comment:源Agent"`
	TargetAgent      string         `json:"targetAgent" gorm:"type:varchar(100);not null;comment:目标Agent"`
	MessageType      string         `json:"messageType" gorm:"type:varchar(50);not null;comment:消息类型"`
	MessageContent   datatypes.JSON `json:"messageContent" gorm:"type:json;comment:消息内容"`
	ResponseContent  datatypes.JSON `json:"responseContent" gorm:"type:json;comment:响应内容"`
	ProcessingTimeMs int            `json:"processingTimeMs" gorm:"comment:处理时间(毫秒)"`
	Status           string         `json:"status" gorm:"type:varchar(50);default:'success';comment:状态"`
	ErrorMessage     string         `json:"errorMessage" gorm:"type:text;comment:错误信息"`
}

// TableName 自定义表名
func (NesmaAgentInteraction) TableName() string {
	return "nesma_agent_interactions"
}
