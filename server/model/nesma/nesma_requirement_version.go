package nesma

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaRequirementVersion 需求版本管理表
type NesmaRequirementVersion struct {
	global.GVA_MODEL
	CycleID     uint           `json:"cycleId" gorm:"not null;comment:所属周期ID;index"`
	Version     string         `json:"version" gorm:"type:varchar(20);not null;comment:版本号，如：v1.0, v1.1, v1.2"`
	VersionType string         `json:"versionType" gorm:"type:varchar(50);not null;comment:版本类型：initial-初始版本,analyzed-AI分析版本,optimized-人工优化版本,finalized-最终版本"`
	CreatedBy   string         `json:"createdBy" gorm:"type:varchar(100);not null;comment:创建方式：user-用户创建,ai_analysis-AI分析,manual_edit-手动编辑"`
	Summary     string         `json:"summary" gorm:"type:varchar(500);comment:版本变更摘要"`
	Description string         `json:"description" gorm:"type:text;comment:详细版本说明"`
	Changes     datatypes.JSON `json:"changes" gorm:"type:json;comment:变更记录详情"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'draft';comment:版本状态：draft-草稿,active-激活,archived-归档"`
	
	// 统计字段
	RequirementCount    int     `json:"requirementCount" gorm:"default:0;comment:该版本需求总数"`
	AnalyzedCount      int     `json:"analyzedCount" gorm:"default:0;comment:已分析需求数"`
	OptimizedCount     int     `json:"optimizedCount" gorm:"default:0;comment:已优化需求数"`
	AnalysisProgress   float64 `json:"analysisProgress" gorm:"default:0;comment:分析进度百分比"`
	
	// AI分析相关
	AnalysisTaskID     *uint   `json:"analysisTaskId" gorm:"comment:关联的分析任务ID"`
	AnalysisStartTime  *time.Time `json:"analysisStartTime" gorm:"comment:分析开始时间"`
	AnalysisEndTime    *time.Time `json:"analysisEndTime" gorm:"comment:分析结束时间"`
	AnalysisDuration   *int    `json:"analysisDuration" gorm:"comment:分析耗时（秒）"`
	
	// 质量评估
	QualityScore       *float64 `json:"qualityScore" gorm:"comment:版本质量评分（0-100）"`
	ConfidenceScore    *float64 `json:"confidenceScore" gorm:"comment:AI分析置信度（0-1）"`
	ReviewStatus       string   `json:"reviewStatus" gorm:"type:varchar(50);default:'pending';comment:评审状态：pending-待评审,approved-已通过,rejected-已拒绝"`
	ReviewNotes        string   `json:"reviewNotes" gorm:"type:text;comment:评审意见"`
	
	// 关联关系
	Cycle        NesmaProjectCycle    `json:"cycle" gorm:"foreignKey:CycleID"`
	Requirements []NesmaRequirement   `json:"requirements" gorm:"foreignKey:VersionID"`
	AnalysisTask *NesmaRequirementAnalysisTask `json:"analysisTask" gorm:"foreignKey:AnalysisTaskID"`
}

// TableName 自定义表名
func (NesmaRequirementVersion) TableName() string {
	return "nesma_requirement_versions"
}

// IsActive 判断版本是否激活
func (v *NesmaRequirementVersion) IsActive() bool {
	return v.Status == "active"
}

// GetVersionTypeText 获取版本类型中文描述
func (v *NesmaRequirementVersion) GetVersionTypeText() string {
	switch v.VersionType {
	case "initial":
		return "初始版本"
	case "analyzed":
		return "AI分析版本"
	case "optimized":
		return "人工优化版本"
	case "finalized":
		return "最终版本"
	default:
		return "未知类型"
	}
}

// GetCreatedByText 获取创建方式中文描述
func (v *NesmaRequirementVersion) GetCreatedByText() string {
	switch v.CreatedBy {
	case "user":
		return "用户创建"
	case "ai_analysis":
		return "AI分析生成"
	case "manual_edit":
		return "手动编辑"
	default:
		return "系统生成"
	}
}

// GetStatusText 获取状态中文描述
func (v *NesmaRequirementVersion) GetStatusText() string {
	switch v.Status {
	case "draft":
		return "草稿"
	case "active":
		return "激活"
	case "archived":
		return "归档"
	default:
		return "未知状态"
	}
}

// GetReviewStatusText 获取评审状态中文描述
func (v *NesmaRequirementVersion) GetReviewStatusText() string {
	switch v.ReviewStatus {
	case "pending":
		return "待评审"
	case "approved":
		return "已通过"
	case "rejected":
		return "已拒绝"
	default:
		return "未评审"
	}
}

// UpdateAnalysisProgress 更新分析进度
func (v *NesmaRequirementVersion) UpdateAnalysisProgress() {
	if v.RequirementCount > 0 {
		v.AnalysisProgress = float64(v.AnalyzedCount) / float64(v.RequirementCount) * 100
	} else {
		v.AnalysisProgress = 0
	}
}

// IsAnalysisCompleted 判断分析是否完成
func (v *NesmaRequirementVersion) IsAnalysisCompleted() bool {
	return v.AnalysisProgress >= 100
}

// GetDurationText 获取分析耗时描述
func (v *NesmaRequirementVersion) GetDurationText() string {
	if v.AnalysisDuration == nil {
		return "未知"
	}
	
	duration := *v.AnalysisDuration
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

// GenerateNextVersion 生成下一个版本号
func (v *NesmaRequirementVersion) GenerateNextVersion() string {
	// 简单的版本号递增逻辑，实际可以更复杂
	switch v.Version {
	case "v1.0":
		return "v1.1"
	case "v1.1":
		return "v1.2"
	case "v1.2":
		return "v1.3"
	default:
		return "v1.0"
	}
}