package response

import "time"

// EvaluationProgressResponse 评估进度响应
type EvaluationProgressResponse struct {
	EvaluationID uint    `json:"evaluation_id"`
	Status       string  `json:"status"`
	Progress     float64 `json:"progress"`
	CurrentPhase string  `json:"current_phase"`
	Message      string  `json:"message"`
	
	// 步骤信息
	CompletedSteps int `json:"completed_steps"`
	TotalSteps     int `json:"total_steps"`
	
	// 需求处理信息
	ProcessedRequirements int `json:"processed_requirements"`
	TotalRequirements     int `json:"total_requirements"`
	
	// 时间信息
	StartTime      *time.Time `json:"start_time"`
	EstimatedTime  *time.Time `json:"estimated_completion_time"`
	ElapsedTime    int64      `json:"elapsed_time_seconds"`
	
	// 阶段列表
	Phases         []PhaseInfo `json:"phases"`
	CurrentPhaseIndex int      `json:"current_phase_index"`
}

// PhaseInfo 阶段信息
type PhaseInfo struct {
	Name        string `json:"name"`
	Index       int    `json:"index"`
	Status      string `json:"status"` // pending, processing, completed, failed
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	Duration    int64  `json:"duration_seconds,omitempty"`
	Description string `json:"description,omitempty"`
}

// EvaluationStatusResponse 评估状态响应
type EvaluationStatusResponse struct {
	EvaluationID   uint     `json:"evaluation_id"`
	Status         string   `json:"status"`
	Progress       float64  `json:"progress"`
	CurrentPhase   string   `json:"current_phase"`
	IsRunning      bool     `json:"is_running"`
	CanCancel      bool     `json:"can_cancel"`
	CanRestart     bool     `json:"can_restart"`
	LastUpdate     time.Time `json:"last_update"`
}

// EvaluationProgressStatsResponse 评估进度统计响应
type EvaluationProgressStatsResponse struct {
	TotalEvaluations     int64   `json:"total_evaluations"`
	RunningEvaluations   int64   `json:"running_evaluations"`
	CompletedEvaluations int64   `json:"completed_evaluations"`
	FailedEvaluations    int64   `json:"failed_evaluations"`
	AverageProgress      float64 `json:"average_progress"`
	AverageDuration      int64   `json:"average_duration_seconds"`
}

// EvaluationLogResponse 评估日志响应
type EvaluationLogResponse struct {
	EvaluationID uint              `json:"evaluation_id"`
	Logs         []EvaluationLogEntry `json:"logs"`
	TotalCount   int               `json:"total_count"`
	PageSize     int               `json:"page_size"`
	CurrentPage  int               `json:"current_page"`
}

// EvaluationLogEntry 评估日志条目
type EvaluationLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // info, warning, error
	Phase     string    `json:"phase"`
	Message   string    `json:"message"`
	Details   interface{} `json:"details,omitempty"`
}

// RunningEvaluationResponse 运行中评估响应
type RunningEvaluationResponse struct {
	EvaluationID     uint      `json:"evaluation_id"`
	ProjectID        uint      `json:"project_id"`
	ProjectName      string    `json:"project_name"`
	Status          string     `json:"status"`
	Progress        float64    `json:"progress"`
	CurrentPhase    string     `json:"current_phase"`
	StartTime       time.Time  `json:"start_time"`
	ElapsedTime     int64      `json:"elapsed_time_seconds"`
	EstimatedTime   *time.Time `json:"estimated_completion_time,omitempty"`
	CanCancel       bool       `json:"can_cancel"`
	ThreadID        string     `json:"thread_id,omitempty"`
}

// EvaluationSystemStatsResponse 评估系统统计响应
type EvaluationSystemStatsResponse struct {
	TotalEvaluations     int64   `json:"total_evaluations"`
	RunningEvaluations   int64   `json:"running_evaluations"`
	CompletedEvaluations int64   `json:"completed_evaluations"`
	FailedEvaluations    int64   `json:"failed_evaluations"`
	CancelledEvaluations int64   `json:"cancelled_evaluations"`
	AverageProgress      float64 `json:"average_progress"`
	AverageDuration      int64   `json:"average_duration_seconds"`
	SystemLoad           float64 `json:"system_load"`
	MemoryUsage          int64   `json:"memory_usage_mb"`
}