package request

// GetEvaluationProgressRequest 获取评估进度请求
type GetEvaluationProgressRequest struct {
	EvaluationID uint `json:"evaluation_id" form:"evaluation_id" binding:"required"`
}

// StartEvaluationProgressRequest 启动评估进度请求
type StartEvaluationProgressRequest struct {
	EvaluationID uint                   `json:"evaluation_id" binding:"required"`
	Config       map[string]interface{} `json:"config,omitempty"`
}

// UpdateEvaluationProgressRequest 更新评估进度请求
type UpdateEvaluationProgressRequest struct {
	EvaluationID  uint    `json:"evaluation_id" binding:"required"`
	Progress      float64 `json:"progress"`
	CurrentPhase  string  `json:"current_phase"`
	Message       string  `json:"message"`
	CompletedSteps int    `json:"completed_steps"`
	TotalSteps    int     `json:"total_steps"`
}

// CancelEvaluationRequest 取消评估请求
type CancelEvaluationRequest struct {
	EvaluationID uint   `json:"evaluation_id" binding:"required"`
	Reason       string `json:"reason,omitempty"`
}

// ForceStopEvaluationRequest 强制停止评估请求
type ForceStopEvaluationRequest struct {
	EvaluationID uint   `json:"evaluation_id" binding:"required"`
	Reason       string `json:"reason,omitempty"`
}