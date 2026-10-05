package request

// WorkflowStep 工作流步骤定义
type WorkflowStep struct {
	ID            string                 `json:"id" binding:"required"`
	Name          string                 `json:"name" binding:"required"`
	AgentType     string                 `json:"agentType" binding:"required"`
	Order         int                    `json:"order"`
	Dependencies  []string               `json:"dependencies"`
	Configuration map[string]interface{} `json:"configuration"`
	Timeout       int                    `json:"timeout"`
	Retries       int                    `json:"retries"`
	Condition     string                 `json:"condition"`
	Description   string                 `json:"description"`
}

// CreateWorkflowRequest 创建工作流请求
type CreateWorkflowRequest struct {
	WorkflowID  string         `json:"workflowId"`
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Steps       []WorkflowStep `json:"steps" binding:"required"`
	Version     string         `json:"version"`
	Category    string         `json:"category"`
}

// ExecuteWorkflowRequest 执行工作流请求
type ExecuteWorkflowRequest struct {
	InputData map[string]interface{} `json:"inputData"`
}

// GetTaskListRequest 获取任务列表请求
type GetTaskListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Priority string `json:"priority" form:"priority"`
	AgentID  string `json:"agentId" form:"agentId"`
}

// GetAgentListRequest 获取Agent列表请求
type GetAgentListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
	Type     string `json:"type" form:"type"`
}

// GetMessageListRequest 获取消息列表请求
type GetMessageListRequest struct {
	Page      int    `json:"page" form:"page"`
	PageSize  int    `json:"pageSize" form:"pageSize"`
	Keyword   string `json:"keyword" form:"keyword"`
	Status    string `json:"status" form:"status"`
	SessionID string `json:"sessionId" form:"sessionId"`
}

// GetWorkflowListRequest 获取工作流列表请求
type GetWorkflowListRequest struct {
	Page     int    `json:"page" form:"page"`
	PageSize int    `json:"pageSize" form:"pageSize"`
	Keyword  string `json:"keyword" form:"keyword"`
	Status   string `json:"status" form:"status"`
}
