package request

import (
	"gorm.io/datatypes"
)

// CreateKnowledgeEntryRequest 创建知识条目请求
type CreateKnowledgeEntryRequest struct {
	Title           string         `json:"title" binding:"required"`
	Content         string         `json:"content" binding:"required"`
	Category        string         `json:"category" binding:"required"`
	Domain          string         `json:"domain"`
	Tags            datatypes.JSON `json:"tags"`
	ConfidenceScore float64        `json:"confidenceScore"`
	Version         string         `json:"version"`
	Source          string         `json:"source"`
	Author          string         `json:"author"`
}

// UpdateKnowledgeEntryRequest 更新知识条目请求
type UpdateKnowledgeEntryRequest struct {
	Title           string         `json:"title"`
	Content         string         `json:"content"`
	Category        string         `json:"category"`
	Domain          string         `json:"domain"`
	Tags            datatypes.JSON `json:"tags"`
	ConfidenceScore float64        `json:"confidenceScore"`
	Version         string         `json:"version"`
	Status          string         `json:"status"`
	Source          string         `json:"source"`
	Author          string         `json:"author"`
}

// ListKnowledgeEntriesRequest 列出知识条目请求
type ListKnowledgeEntriesRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Category string `json:"category" form:"category"`
	Domain   string `json:"domain" form:"domain"`
	Status   string `json:"status" form:"status"`
	Keyword  string `json:"keyword" form:"keyword"`
}

// SearchKnowledgeEntriesRequest 搜索知识条目请求
type SearchKnowledgeEntriesRequest struct {
	Query     string  `json:"query" binding:"required"`
	Category  string  `json:"category"`
	Limit     int     `json:"limit"`
	Threshold float32 `json:"threshold"`
}

// CreateKnowledgeRuleRequest 创建知识规则请求
type CreateKnowledgeRuleRequest struct {
	RuleName            string `json:"ruleName" binding:"required"`
	ConditionExpression string `json:"conditionExpression" binding:"required"`
	ActionExpression    string `json:"actionExpression" binding:"required"`
	Priority            int    `json:"priority"`
	IsActive            bool   `json:"isActive"`
	Description         string `json:"description"`
	Category            string `json:"category"`
}

// UpdateKnowledgeRuleRequest 更新知识规则请求
type UpdateKnowledgeRuleRequest struct {
	RuleName            string `json:"ruleName"`
	ConditionExpression string `json:"conditionExpression"`
	ActionExpression    string `json:"actionExpression"`
	Priority            int    `json:"priority"`
	IsActive            *bool  `json:"isActive"`
	Description         string `json:"description"`
	Category            string `json:"category"`
}

// ListKnowledgeRulesRequest 列出知识规则请求
type ListKnowledgeRulesRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Category string `json:"category" form:"category"`
	IsActive *bool  `json:"isActive" form:"isActive"`
}

// CreateCaseStudyRequest 创建案例研究请求
type CreateCaseStudyRequest struct {
	ProjectName         string      `json:"projectName" binding:"required"`
	Domain              string      `json:"domain"`
	RequirementsSummary string      `json:"requirementsSummary"`
	NesmaResults        interface{} `json:"nesmaResults"`
	LessonsLearned      string      `json:"lessonsLearned"`
	TotalFunctionPoints float64     `json:"totalFunctionPoints"`
	ProjectDuration     int         `json:"projectDuration"`
	TeamSize            int         `json:"teamSize"`
	Complexity          string      `json:"complexity"`
	Success             bool        `json:"success"`
}

// UpdateCaseStudyRequest 更新案例研究请求
type UpdateCaseStudyRequest struct {
	ProjectName         string      `json:"projectName"`
	Domain              string      `json:"domain"`
	RequirementsSummary string      `json:"requirementsSummary"`
	NesmaResults        interface{} `json:"nesmaResults"`
	LessonsLearned      string      `json:"lessonsLearned"`
	TotalFunctionPoints float64     `json:"totalFunctionPoints"`
	ProjectDuration     int         `json:"projectDuration"`
	TeamSize            int         `json:"teamSize"`
	Complexity          string      `json:"complexity"`
	Success             *bool       `json:"success"`
}

// ListCaseStudiesRequest 列出案例研究请求
type ListCaseStudiesRequest struct {
	Page       int    `json:"page" form:"page"`
	PageSize   int    `json:"pageSize" form:"pageSize"`
	Domain     string `json:"domain" form:"domain"`
	Complexity string `json:"complexity" form:"complexity"`
	Success    *bool  `json:"success" form:"success"`
}

// SearchCaseStudiesRequest 搜索案例研究请求
type SearchCaseStudiesRequest struct {
	Query     string  `json:"query" binding:"required"`
	Domain    string  `json:"domain"`
	Limit     int     `json:"limit"`
	Threshold float32 `json:"threshold"`
}

// Agent相关请求结构

// RegisterAgentRequest 注册Agent请求
type RegisterAgentRequest struct {
	AgentID        string         `json:"agentId" binding:"required"`
	Name           string         `json:"name" binding:"required"`
	Description    string         `json:"description"`
	AgentType      string         `json:"agentType" binding:"required"`
	Capabilities   datatypes.JSON `json:"capabilities"`
	Configuration  datatypes.JSON `json:"configuration"`
	MaxConcurrency int            `json:"maxConcurrency"`
	Version        string         `json:"version"`
	Endpoint       string         `json:"endpoint" binding:"required"`
}

// UpdateAgentRequest 更新Agent请求
type UpdateAgentRequest struct {
	Name           string         `json:"name"`
	Description    string         `json:"description"`
	Configuration  datatypes.JSON `json:"configuration"`
	MaxConcurrency int            `json:"maxConcurrency"`
	Version        string         `json:"version"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	TaskType   string         `json:"taskType" binding:"required"`
	Priority   int            `json:"priority"`
	InputData  datatypes.JSON `json:"inputData"`
	SessionID  string         `json:"sessionId"`
	UserID     uint           `json:"userId"`
	ProjectID  uint           `json:"projectId"`
	MaxRetries int            `json:"maxRetries"`
	Metadata   datatypes.JSON `json:"metadata"`
}

// AssignTaskRequest 分配任务请求
type AssignTaskRequest struct {
	AgentID string `json:"agentId" binding:"required"`
}

// SendMessageRequest 发送消息请求
type SendMessageRequest struct {
	Type      string                 `json:"type" binding:"required"`
	From      string                 `json:"from" binding:"required"`
	To        string                 `json:"to" binding:"required"`
	SessionID string                 `json:"sessionId"`
	Content   map[string]interface{} `json:"content"`
	Priority  int                    `json:"priority"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// UpdateStatusRequest 更新状态请求
type UpdateStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
