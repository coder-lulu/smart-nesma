package nesma

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaRequirementAnalysisTask 需求分析任务表
type NesmaRequirementAnalysisTask struct {
	global.GVA_MODEL
	ProjectID       uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	CycleID         uint           `json:"cycleId" gorm:"not null;comment:周期ID;index"`
	SourceVersionID uint           `json:"sourceVersionId" gorm:"not null;comment:源版本ID"`
	TargetVersionID *uint          `json:"targetVersionId" gorm:"comment:目标版本ID"`
	TaskType        string         `json:"taskType" gorm:"type:varchar(50);not null;comment:任务类型：requirement_analysis-需求分析,description_generation-描述生成,flowchart_generation-流程图生成"`
	Status          string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:任务状态：pending-等待中,running-执行中,completed-已完成,failed-失败,cancelled-已取消"`
	Progress        int            `json:"progress" gorm:"default:0;comment:执行进度0-100"`
	
	// 任务配置
	Config          datatypes.JSON `json:"config" gorm:"type:json;comment:任务配置参数"`
	Priority        int            `json:"priority" gorm:"default:5;comment:任务优先级1-10，数字越小优先级越高"`
	
	// 执行信息
	StartTime       *time.Time     `json:"startTime" gorm:"comment:任务开始时间"`
	EndTime         *time.Time     `json:"endTime" gorm:"comment:任务结束时间"`
	Duration        *int           `json:"duration" gorm:"comment:执行耗时（秒）"`
	
	// 结果统计
	TotalCount      int            `json:"totalCount" gorm:"default:0;comment:总需求数量"`
	ProcessedCount  int            `json:"processedCount" gorm:"default:0;comment:已处理数量"`
	SuccessCount    int            `json:"successCount" gorm:"default:0;comment:成功处理数量"`
	FailedCount     int            `json:"failedCount" gorm:"default:0;comment:失败处理数量"`
	SkippedCount    int            `json:"skippedCount" gorm:"default:0;comment:跳过处理数量"`
	
	// 错误信息
	ErrorMsg        string         `json:"errorMsg" gorm:"type:text;comment:错误信息"`
	ErrorDetails    datatypes.JSON `json:"errorDetails" gorm:"type:json;comment:详细错误信息"`
	
	// 任务结果
	Result          datatypes.JSON `json:"result" gorm:"type:json;comment:任务执行结果"`
	Summary         string         `json:"summary" gorm:"type:text;comment:任务执行摘要"`
	
	// AI相关统计
	AICallCount     int            `json:"aiCallCount" gorm:"default:0;comment:AI调用次数"`
	AICallDuration  *int           `json:"aiCallDuration" gorm:"comment:AI调用总耗时（毫秒）"`
	AITokenUsage    *int           `json:"aiTokenUsage" gorm:"comment:AI Token使用量"`
	
	// 质量评估
	QualityScore    *float64       `json:"qualityScore" gorm:"comment:输出质量评分（0-100）"`
	AccuracyRate    *float64       `json:"accuracyRate" gorm:"comment:准确率（0-1）"`
	
	// 需求ID列表（用于特定需求分析）
	RequirementIDs  datatypes.JSON `json:"requirementIds" gorm:"type:json;comment:需求ID列表"`
	
	// 关联关系
	Project       NesmaProject              `json:"project" gorm:"foreignKey:ProjectID"`
	Cycle         NesmaProjectCycle         `json:"cycle" gorm:"foreignKey:CycleID"`
	SourceVersion NesmaRequirementVersion   `json:"sourceVersion" gorm:"foreignKey:SourceVersionID"`
	TargetVersion *NesmaRequirementVersion  `json:"targetVersion" gorm:"foreignKey:TargetVersionID"`
}

// TableName 自定义表名
func (NesmaRequirementAnalysisTask) TableName() string {
	return "nesma_requirement_analysis_tasks"
}

// IsRunning 判断任务是否正在运行
func (t *NesmaRequirementAnalysisTask) IsRunning() bool {
	return t.Status == "running"
}

// IsCompleted 判断任务是否已完成
func (t *NesmaRequirementAnalysisTask) IsCompleted() bool {
	return t.Status == "completed"
}

// IsFailed 判断任务是否失败
func (t *NesmaRequirementAnalysisTask) IsFailed() bool {
	return t.Status == "failed"
}

// GetTaskTypeText 获取任务类型中文描述
func (t *NesmaRequirementAnalysisTask) GetTaskTypeText() string {
	switch t.TaskType {
	case "requirement_analysis":
		return "需求分析"
	case "description_generation":
		return "描述生成"
	case "flowchart_generation":
		return "流程图生成"
	default:
		return "未知任务"
	}
}

// GetStatusText 获取状态中文描述
func (t *NesmaRequirementAnalysisTask) GetStatusText() string {
	switch t.Status {
	case "pending":
		return "等待中"
	case "running":
		return "执行中"
	case "completed":
		return "已完成"
	case "failed":
		return "失败"
	case "cancelled":
		return "已取消"
	default:
		return "未知状态"
	}
}

// GetProgressText 获取进度描述
func (t *NesmaRequirementAnalysisTask) GetProgressText() string {
	if t.TotalCount == 0 {
		return fmt.Sprintf("%d%%", t.Progress)
	}
	return fmt.Sprintf("%d%% (%d/%d)", t.Progress, t.ProcessedCount, t.TotalCount)
}

// GetDurationText 获取耗时描述
func (t *NesmaRequirementAnalysisTask) GetDurationText() string {
	if t.Duration == nil {
		if t.StartTime != nil && t.EndTime == nil && t.IsRunning() {
			// 任务正在运行，计算当前耗时
			duration := int(time.Since(*t.StartTime).Seconds())
			return t.formatDuration(duration) + " (进行中)"
		}
		return "未知"
	}
	return t.formatDuration(*t.Duration)
}

// formatDuration 格式化耗时
func (t *NesmaRequirementAnalysisTask) formatDuration(duration int) string {
	if duration < 60 {
		return fmt.Sprintf("%d秒", duration)
	} else if duration < 3600 {
		return fmt.Sprintf("%d分%d秒", duration/60, duration%60)
	} else {
		hours := duration / 3600
		minutes := (duration % 3600) / 60
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	}
}

// GetSuccessRate 获取成功率
func (t *NesmaRequirementAnalysisTask) GetSuccessRate() float64 {
	if t.ProcessedCount == 0 {
		return 0
	}
	return float64(t.SuccessCount) / float64(t.ProcessedCount) * 100
}

// GetAIEfficiency 获取AI调用效率
func (t *NesmaRequirementAnalysisTask) GetAIEfficiency() string {
	if t.AICallCount == 0 || t.AICallDuration == nil {
		return "无数据"
	}
	avgDuration := *t.AICallDuration / t.AICallCount
	return fmt.Sprintf("平均%dms/次", avgDuration)
}

// UpdateProgress 更新任务进度
func (t *NesmaRequirementAnalysisTask) UpdateProgress() {
	if t.TotalCount > 0 {
		// 使用更精确的进度计算
		percentage := float64(t.ProcessedCount) / float64(t.TotalCount) * 100
		t.Progress = int(percentage)
		
		// 特殊处理：如果有进度但计算结果为0，设置为1确保显示
		if t.ProcessedCount > 0 && t.Progress == 0 {
			t.Progress = 1
		}
		
		// 为了更好的用户体验，设置最小进度值
		if t.ProcessedCount > 0 && t.Progress < 1 {
			t.Progress = 1
		}
	}
	
	// 确保进度不超过100
	if t.Progress > 100 {
		t.Progress = 100
	}
}

// SetCompleted 设置任务完成
func (t *NesmaRequirementAnalysisTask) SetCompleted() {
	t.Status = "completed"
	t.Progress = 100
	now := time.Now()
	t.EndTime = &now
	
	if t.StartTime != nil {
		duration := int(now.Sub(*t.StartTime).Seconds())
		t.Duration = &duration
	}
}

// SetFailed 设置任务失败
func (t *NesmaRequirementAnalysisTask) SetFailed(errorMsg string) {
	t.Status = "failed"
	t.ErrorMsg = errorMsg
	now := time.Now()
	t.EndTime = &now
	
	if t.StartTime != nil {
		duration := int(now.Sub(*t.StartTime).Seconds())
		t.Duration = &duration
	}
}

// SetRunning 设置任务运行中
func (t *NesmaRequirementAnalysisTask) SetRunning() {
	t.Status = "running"
	now := time.Now()
	t.StartTime = &now
}