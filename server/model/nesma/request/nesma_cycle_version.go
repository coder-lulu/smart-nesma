package request

import (
	"time"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
)

// NesmaProjectCycleRequest 项目周期请求结构
type NesmaProjectCycleRequest struct {
	ID          uint       `json:"id"`
	ProjectID   uint       `json:"projectId" binding:"required"`
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
	Status      string     `json:"status"`
	Phase       string     `json:"phase"`
}

// NesmaProjectCycleSearch 项目周期搜索结构
type NesmaProjectCycleSearch struct {
	request.PageInfo
	ProjectID *uint  `json:"projectId" form:"projectId"`
	Status    string `json:"status" form:"status"`
	Keyword   string `json:"keyword" form:"keyword"`
}

// NesmaRequirementVersionRequest 需求版本请求结构
type NesmaRequirementVersionRequest struct {
	ID          uint   `json:"id"`
	CycleID     uint   `json:"cycleId" binding:"required"`
	Version     string `json:"version" binding:"required"`
	VersionType string `json:"versionType" binding:"required"`
	CreatedBy   string `json:"createdBy" binding:"required"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// NesmaRequirementVersionSearch 需求版本搜索结构
type NesmaRequirementVersionSearch struct {
	request.PageInfo
	CycleID     *uint  `json:"cycleId" form:"cycleId"`
	VersionType string `json:"versionType" form:"versionType"`
	Status      string `json:"status" form:"status"`
	CreatedBy   string `json:"createdBy" form:"createdBy"`
}

// NesmaAnalyzeRequest 一键分析请求结构
type NesmaAnalyzeRequest struct {
	ProjectID      uint   `json:"projectId" binding:"required"`
	CycleID        uint   `json:"cycleId" binding:"required"`
	RequirementIDs []uint `json:"requirementIds"` // 特定需求ID列表，如果为空则分析整个周期
	Config         struct {
		AnalysisType    string   `json:"analysisType"`    // requirement_analysis, description_generation
		TargetLevels    []int    `json:"targetLevels"`    // 目标层级：[3,4]
		AIModel         string   `json:"aiModel"`         // deepseek, openai
		EnableMermaid   bool     `json:"enableMermaid"`   // 是否生成流程图
		BatchSize       int      `json:"batchSize"`       // 批次大小
		MaxRetries      int      `json:"maxRetries"`      // 最大重试次数
		SkipCompleted   bool     `json:"skipCompleted"`   // 跳过已分析的需求
	} `json:"config"`
}