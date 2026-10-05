package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
)

// NesmaProjectResponse 项目响应
type NesmaProjectResponse struct {
	Project          nesma.NesmaProject `json:"project"`
	RequirementStats RequirementStats   `json:"requirementStats"`
}

// NesmaProjectWithStats 带统计信息的项目
type NesmaProjectWithStats struct {
	nesma.NesmaProject
	RequirementStats RequirementStats `json:"requirementStats"`
}

// RequirementStats 需求统计
type RequirementStats struct {
	TotalCount int64 `json:"totalCount"` // 总需求数
	Level1     int64 `json:"level1"`     // 一级功能模块数
	Level2     int64 `json:"level2"`     // 二级功能模块数
	Level3     int64 `json:"level3"`     // 三级功能模块数
	Level4     int64 `json:"level4"`     // 功能点计数项数
	Completed  int64 `json:"completed"`  // 已完成数
}

// NesmaProjectListResponse 项目列表响应
type NesmaProjectListResponse struct {
	List     []NesmaProjectWithStats `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
}

// NesmaProjectStatsResponse 项目统计响应
type NesmaProjectStatsResponse struct {
	TotalProjects       int64                `json:"totalProjects"`       // 总项目数
	StatusStats         []StatusStat         `json:"statusStats"`         // 按状态统计
	DomainStats         []DomainStat         `json:"domainStats"`         // 按领域统计
	RecentProjects      int64                `json:"recentProjects"`      // 最近30天项目数
	RecentProjectsList  []RecentProject      `json:"recentProjectsList"`  // 最近项目列表（用于工作台）
	Projects            []ProjectBasicInfo   `json:"projects"`            // 所有项目基本信息（用于批量操作）
}

// RecentProject 最近项目信息
type RecentProject struct {
	ID               uint   `json:"id"`
	Name             string `json:"name"`
	Status           string `json:"status"`
	RequirementCount int64  `json:"requirementCount"`
	FunctionPoints   int64  `json:"functionPoints"`
}

// ProjectBasicInfo 项目基本信息
type ProjectBasicInfo struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// StatusStat 状态统计
type StatusStat struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// DomainStat 领域统计
type DomainStat struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
}
