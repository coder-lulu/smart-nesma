package request

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"gorm.io/datatypes"
)

// NesmaProjectSearch 项目搜索结构体
type NesmaProjectSearch struct {
	request.PageInfo
	Name      string     `json:"name" form:"name"`           // 项目名称
	Status    string     `json:"status" form:"status"`       // 项目状态
	Domain    string     `json:"domain" form:"domain"`       // 项目领域
	StartDate *time.Time `json:"startDate" form:"startDate"` // 开始日期
	EndDate   *time.Time `json:"endDate" form:"endDate"`     // 结束日期
}

// NesmaProjectCreate 创建项目请求
type NesmaProjectCreate struct {
	Name        string            `json:"name" binding:"required" comment:"项目名称"`
	Description string            `json:"description" comment:"项目描述"`
	Domain      string            `json:"domain" comment:"项目领域"`
	DomainTags  datatypes.JSON    `json:"domainTags" comment:"领域标签"`
	Settings    datatypes.JSON    `json:"settings" comment:"项目配置"`
	StartDate   *nesma.CustomTime `json:"startDate" comment:"项目开始时间"`
	EndDate     *nesma.CustomTime `json:"endDate" comment:"项目结束时间"`
}

// NesmaProjectUpdate 更新项目请求
type NesmaProjectUpdate struct {
	ID          uint              `json:"id" binding:"required" comment:"项目ID"`
	Name        string            `json:"name" binding:"required" comment:"项目名称"`
	Description string            `json:"description" comment:"项目描述"`
	Status      string            `json:"status" comment:"项目状态"`
	Domain      string            `json:"domain" comment:"项目领域"`
	DomainTags  datatypes.JSON    `json:"domainTags" comment:"领域标签"`
	Settings    datatypes.JSON    `json:"settings" comment:"项目配置"`
	StartDate   *nesma.CustomTime `json:"startDate" comment:"项目开始时间"`
	EndDate     *nesma.CustomTime `json:"endDate" comment:"项目结束时间"`
}

// NesmaProjectByID 根据ID获取项目请求
type NesmaProjectByID struct {
	ID uint `json:"id" form:"id" binding:"required" comment:"项目ID"`
}

// NesmaProjectDeleteByIds 批量删除项目请求
type NesmaProjectDeleteByIds struct {
	IDs []uint `json:"ids" form:"ids" binding:"required" comment:"项目ID列表"`
}

// NesmaOneClickAnalysisRequest 一键分析项目请求
type NesmaOneClickAnalysisRequest struct {
	ProjectID uint `json:"projectId" form:"projectId" binding:"required" comment:"项目ID"`
	CycleID   uint `json:"cycleId" form:"cycleId" binding:"required" comment:"项目周期ID"`
}
