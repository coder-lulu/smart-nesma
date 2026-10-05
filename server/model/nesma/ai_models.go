package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// AIServiceMetrics AI服务监控指标
type AIServiceMetrics struct {
	global.GVA_MODEL
	ServiceName      string         `json:"serviceName" gorm:"type:varchar(100);not null;index"`
	MetricType       string         `json:"metricType" gorm:"type:varchar(50);not null"` // response_time, token_usage, error_rate, success_rate
	MetricValue      float64        `json:"metricValue" gorm:"not null"`
	RequestCount     int64          `json:"requestCount" gorm:"default:0"`
	ErrorCount       int64          `json:"errorCount" gorm:"default:0"`
	TotalTokens      int64          `json:"totalTokens" gorm:"default:0"`
	PromptTokens     int64          `json:"promptTokens" gorm:"default:0"`
	CompletionTokens int64          `json:"completionTokens" gorm:"default:0"`
	AverageLatency   float64        `json:"averageLatency" gorm:"default:0"`
	P95Latency       float64        `json:"p95Latency" gorm:"default:0"`
	P99Latency       float64        `json:"p99Latency" gorm:"default:0"`
	ErrorMessages    datatypes.JSON `json:"errorMessages" gorm:"type:json"`
	Metadata         datatypes.JSON `json:"metadata" gorm:"type:json"`
	RecordedAt       time.Time      `json:"recordedAt" gorm:"index"`
}

// TableName 自定义表名
func (AIServiceMetrics) TableName() string {
	return "ai_service_metrics"
}

// UserBehavior 用户行为记录
type UserBehavior struct {
	global.GVA_MODEL
	UserID     uint           `json:"userId" gorm:"not null;index"`
	Action     string         `json:"action" gorm:"type:varchar(50);not null"` // view, like, share, download, create, edit, delete
	ItemType   string         `json:"itemType" gorm:"type:varchar(50);not null"`
	ItemID     uint           `json:"itemId" gorm:"not null"`
	SessionID  string         `json:"sessionId" gorm:"type:varchar(255)"`
	Context    datatypes.JSON `json:"context" gorm:"type:json"`
	Duration   int            `json:"duration" gorm:"default:0"` // 停留时间(秒)
	DeviceType string         `json:"deviceType" gorm:"type:varchar(50)"`
	UserAgent  string         `json:"userAgent" gorm:"type:varchar(500)"`
	IPAddress  string         `json:"ipAddress" gorm:"type:varchar(50)"`
}

// TableName 自定义表名
func (UserBehavior) TableName() string {
	return "user_behaviors"
}

// RecommendationHistory 推荐历史
type RecommendationHistory struct {
	global.GVA_MODEL
	UserID             uint           `json:"userId" gorm:"not null;index"`
	RecommendationType string         `json:"recommendationType" gorm:"type:varchar(50);not null"`
	ItemType           string         `json:"itemType" gorm:"type:varchar(50);not null"`
	ItemID             uint           `json:"itemId" gorm:"not null"`
	Score              float64        `json:"score" gorm:"not null"`
	Reason             string         `json:"reason" gorm:"type:text"`
	Algorithm          string         `json:"algorithm" gorm:"type:varchar(50)"`
	Context            datatypes.JSON `json:"context" gorm:"type:json"`
	IsClicked          bool           `json:"isClicked" gorm:"default:false"`
	IsInteracted       bool           `json:"isInteracted" gorm:"default:false"`
	FeedbackScore      *float64       `json:"feedbackScore" gorm:"default:null"`
	FeedbackComment    string         `json:"feedbackComment" gorm:"type:text"`
}

// TableName 自定义表名
func (RecommendationHistory) TableName() string {
	return "recommendation_histories"
}

// ChatSession 聊天会话
type ChatSession struct {
	global.GVA_MODEL
	UserID      uint          `json:"userId" gorm:"not null;index"`
	ProjectID   *uint         `json:"projectId" gorm:"index"`
	Title       string        `json:"title" gorm:"type:varchar(255);not null"`
	Description string        `json:"description" gorm:"type:text"`
	ModelName   string        `json:"modelName" gorm:"type:varchar(50);not null"`
	IsActive    bool          `json:"isActive" gorm:"default:true"`
	Messages    []ChatMessage `json:"messages" gorm:"foreignKey:SessionID"`
}

// TableName 自定义表名
func (ChatSession) TableName() string {
	return "chat_sessions"
}

// ChatMessage 聊天消息
type ChatMessage struct {
	global.GVA_MODEL
	SessionID  uint   `json:"sessionId" gorm:"not null;index"`
	Role       string `json:"role" gorm:"type:varchar(20);not null"` // user, assistant, system
	Content    string `json:"content" gorm:"type:text;not null"`
	Metadata   string `json:"metadata" gorm:"type:json"`
	TokenCount int    `json:"tokenCount" gorm:"default:0"`
}

// TableName 自定义表名
func (ChatMessage) TableName() string {
	return "chat_messages"
}

// RequirementAnalysisResult 需求分析结果
type RequirementAnalysisResult struct {
	global.GVA_MODEL
	UserID         uint           `json:"userId" gorm:"not null;index"`
	ProjectID      *uint          `json:"projectId" gorm:"index"`
	RequirementID  *uint          `json:"requirementId" gorm:"index"`
	AnalysisType   string         `json:"analysisType" gorm:"type:varchar(50);not null"` // refinement, quality, decomposition, risk
	InputText      string         `json:"inputText" gorm:"type:text;not null"`
	Result         datatypes.JSON `json:"result" gorm:"type:json;not null"`
	Suggestions    datatypes.JSON `json:"suggestions" gorm:"type:json"`
	QualityScore   float64        `json:"qualityScore" gorm:"default:0"`
	ProcessingTime int64          `json:"processingTime" gorm:"default:0"` // 处理时间(毫秒)
	ModelName      string         `json:"modelName" gorm:"type:varchar(50)"`
	Version        string         `json:"version" gorm:"type:varchar(20);default:'1.0'"`
	Status         string         `json:"status" gorm:"type:varchar(20);default:'completed'"`
	ErrorMessage   string         `json:"errorMessage" gorm:"type:text"`
}

// TableName 自定义表名
func (RequirementAnalysisResult) TableName() string {
	return "requirement_analysis_results"
}
