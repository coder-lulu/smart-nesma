package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaProjectCycle 项目建设周期表
type NesmaProjectCycle struct {
	global.GVA_MODEL
	ProjectID   uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	Name        string         `json:"name" gorm:"type:varchar(100);not null;comment:周期名称，如：一期、二期、三期"`
	Description string         `json:"description" gorm:"type:text;comment:周期描述"`
	StartDate   *time.Time     `json:"startDate" gorm:"comment:周期开始时间"`
	EndDate     *time.Time     `json:"endDate" gorm:"comment:周期结束时间"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'planning';comment:周期状态：planning-规划中,active-进行中,completed-已完成,suspended-暂停"`
	Phase       string         `json:"phase" gorm:"type:varchar(50);comment:当前阶段：需求分析,设计,开发,测试,上线"`
	Settings    datatypes.JSON `json:"settings" gorm:"type:json;comment:周期特有配置"`
	
	// 统计字段
	RequirementCount int     `json:"requirementCount" gorm:"default:0;comment:需求总数"`
	CompletedCount   int     `json:"completedCount" gorm:"default:0;comment:已完成需求数"`
	Progress         float64 `json:"progress" gorm:"default:0;comment:完成进度百分比"`
	
	// 关联关系
	Project  NesmaProject           `json:"project" gorm:"foreignKey:ProjectID"`
	Versions []NesmaRequirementVersion `json:"versions" gorm:"foreignKey:CycleID"`
}

// TableName 自定义表名
func (NesmaProjectCycle) TableName() string {
	return "nesma_project_cycles"
}

// IsActive 判断周期是否激活
func (c *NesmaProjectCycle) IsActive() bool {
	return c.Status == "active"
}

// GetStatusText 获取状态中文描述
func (c *NesmaProjectCycle) GetStatusText() string {
	switch c.Status {
	case "planning":
		return "规划中"
	case "active":
		return "进行中"
	case "completed":
		return "已完成"
	case "suspended":
		return "暂停"
	default:
		return "未知状态"
	}
}

// GetPhaseText 获取阶段中文描述
func (c *NesmaProjectCycle) GetPhaseText() string {
	switch c.Phase {
	case "requirement":
		return "需求分析"
	case "design":
		return "系统设计"
	case "development":
		return "开发实施"
	case "testing":
		return "测试验证"
	case "deployment":
		return "部署上线"
	default:
		return "未设置"
	}
}

// UpdateProgress 更新进度
func (c *NesmaProjectCycle) UpdateProgress() {
	if c.RequirementCount > 0 {
		c.Progress = float64(c.CompletedCount) / float64(c.RequirementCount) * 100
	} else {
		c.Progress = 0
	}
}