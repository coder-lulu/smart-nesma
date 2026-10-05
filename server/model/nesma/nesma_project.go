package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaProject NESMA项目表
type NesmaProject struct {
	global.GVA_MODEL
	Name        string         `json:"name" gorm:"type:varchar(255);not null;comment:项目名称"`
	Description string         `json:"description" gorm:"type:text;comment:项目描述"`
	OwnerID     uint           `json:"ownerId" gorm:"not null;comment:项目负责人ID"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'active';comment:项目状态"`
	Domain      string         `json:"domain" gorm:"type:varchar(100);comment:项目领域（金融、电商、医疗等）"`
	DomainTags  datatypes.JSON `json:"domainTags" gorm:"type:json;comment:领域标签"`
	Settings    datatypes.JSON `json:"settings" gorm:"type:json;comment:项目配置"`
	StartDate     *time.Time     `json:"startDate" gorm:"comment:项目开始时间"`
	EndDate       *time.Time     `json:"endDate" gorm:"comment:项目结束时间"`
	ActiveCycleID *uint          `json:"activeCycleId" gorm:"comment:当前激活的周期ID"`

	// 关联关系
	ActiveCycle   *NesmaProjectCycle `json:"activeCycle" gorm:"foreignKey:ActiveCycleID"`
	Requirements []NesmaRequirement   `json:"requirements" gorm:"foreignKey:ProjectID"`
	Cycles       []NesmaProjectCycle  `json:"cycles" gorm:"foreignKey:ProjectID"`
}

// TableName 自定义表名
func (NesmaProject) TableName() string {
	return "nesma_projects"
}
