package nesma

import (
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaRequirement NESMA需求表
type NesmaRequirement struct {
	global.GVA_MODEL
	ProjectID   uint           `json:"projectId" gorm:"not null;comment:项目ID;index"`
	ParentID    *uint          `json:"parentId" gorm:"comment:父需求ID，支持层级结构"`
	Level       int            `json:"level" gorm:"not null;comment:需求层级：1-一级功能模块,2-二级功能模块,3-三级功能模块,4-功能点计数项"`
	Code        string         `json:"code" gorm:"type:varchar(100);comment:需求编号"`
	Title       string         `json:"title" gorm:"type:varchar(500);not null;comment:需求标题"`
	Description string         `json:"description" gorm:"type:text;comment:需求描述"`
	Priority    int            `json:"priority" gorm:"default:3;comment:优先级1-5"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:需求状态"`
	OrderIndex  int            `json:"orderIndex" gorm:"default:0;comment:排序索引"`
	DomainTags  datatypes.JSON `json:"domainTags" gorm:"type:json;comment:领域标签"`

	// 扩展字段，支持Excel导入
	Category           string   `json:"category" gorm:"type:varchar(100);comment:功能分类"`
	Complexity         string   `json:"complexity" gorm:"type:varchar(50);comment:复杂度：简单/中等/复杂"`
	EstimateHours      *float64 `json:"estimateHours" gorm:"comment:预估工时"`
	ActualHours        *float64 `json:"actualHours" gorm:"comment:实际工时"`
	BusinessValue      string   `json:"businessValue" gorm:"type:text;comment:业务价值"`
	AcceptanceCriteria string   `json:"acceptanceCriteria" gorm:"type:text;comment:验收标准"`
	Notes              string   `json:"notes" gorm:"type:text;comment:备注"`

	// 导入相关字段
	ImportBatch  string `json:"importBatch" gorm:"type:varchar(100);comment:导入批次号"`
	ImportSource string `json:"importSource" gorm:"type:varchar(100);comment:导入来源"`

	// 新增字段
	ConstructionPeriod string  `json:"constructionPeriod" gorm:"type:varchar(100);comment:建设周期"`
	Version            float64 `json:"version" gorm:"default:0;comment:版本"`
	AFP                float64 `json:"afp" gorm:"comment:调整后功能点数"`
	UFP                float64 `json:"ufp" gorm:"comment:未调整功能点数"`
	FunctionType       string  `json:"functionType" gorm:"type:varchar(10);comment:功能类别：EQ,EI,EO,ILF,EIF"`
	ReuseLevel         string  `json:"reuseLevel" gorm:"type:varchar(20);comment:重用程度：高,中,低"`
	ModificationType   string  `json:"modificationType" gorm:"type:varchar(20);comment:修改类型：新增,优化,删除"`

	// 版本管理字段  
	CycleID   *uint `json:"cycleId" gorm:"comment:所属项目周期ID;index"`
	VersionID *uint `json:"versionId" gorm:"comment:所属需求版本ID;index"`

	// AI分析相关字段
	AIAnalysisStatus  string         `json:"aiAnalysisStatus" gorm:"type:varchar(50);default:'pending';comment:AI分析状态：pending-待分析,analyzing-分析中,completed-已完成,failed-失败"`
	AIDescription     string         `json:"aiDescription" gorm:"type:text;comment:AI生成的需求描述"`
	AIGeneratedTitle  string         `json:"aiGeneratedTitle" gorm:"type:varchar(500);comment:AI优化的需求标题"`
	AIComplexityScore *float64       `json:"aiComplexityScore" gorm:"comment:AI评估的复杂度得分（0-100）"`
	AIConfidenceScore *float64       `json:"aiConfidenceScore" gorm:"comment:AI分析置信度（0-1）"`
	AIAnalysisTime    *time.Time     `json:"aiAnalysisTime" gorm:"comment:AI分析完成时间"`
	AIAnalysisLog     datatypes.JSON `json:"aiAnalysisLog" gorm:"type:json;comment:AI分析过程日志"`
	
	// Mermaid流程图字段
	MermaidDiagram    string         `json:"mermaidDiagram" gorm:"type:text;comment:Mermaid流程图代码"`
	MermaidGeneratedAt *time.Time    `json:"mermaidGeneratedAt" gorm:"comment:流程图生成时间"`

	// 智能推荐字段
	SimilarRequirements   datatypes.JSON `json:"similarRequirements" gorm:"type:json;comment:相似需求推荐"`
	KnowledgeReferences   string         `json:"knowledgeReferences" gorm:"type:text;comment:知识库引用"`
	RecommendedAFP        *float64       `json:"recommendedAFP" gorm:"comment:AI推荐的AFP值"`
	RecommendedUFP        *float64       `json:"recommendedUFP" gorm:"comment:AI推荐的UFP值"`

	// 优化相关字段
	AIOptimizationApplied   bool       `json:"aiOptimizationApplied" gorm:"default:false;comment:是否已应用AI优化建议"`
	AIOptimizationAppliedAt *time.Time `json:"aiOptimizationAppliedAt" gorm:"comment:AI优化应用时间"`
	ComplexityLevel         string     `json:"complexityLevel" gorm:"type:varchar(50);comment:复杂度级别"`

	// 关联关系
	Project  NesmaProject               `json:"project" gorm:"foreignKey:ProjectID"`
	Parent   *NesmaRequirement          `json:"parent" gorm:"foreignKey:ParentID"`
	Children []NesmaRequirement         `json:"children" gorm:"foreignKey:ParentID"`
	Analysis []NesmaRequirementAnalysis `json:"analysis" gorm:"foreignKey:RequirementID"`
	Cycle    *NesmaProjectCycle         `json:"cycle" gorm:"foreignKey:CycleID"`
}

// TableName 自定义表名
func (NesmaRequirement) TableName() string {
	return "nesma_requirements"
}

// GetLevelName 获取层级名称
func (r *NesmaRequirement) GetLevelName() string {
	switch r.Level {
	case 1:
		return "一级功能模块"
	case 2:
		return "二级功能模块"
	case 3:
		return "三级功能模块"
	case 4:
		return "功能点计数项"
	default:
		return "未知层级"
	}
}

// GetFullPath 获取完整路径
func (r *NesmaRequirement) GetFullPath() string {
	if r.Parent == nil {
		return r.Title
	}
	return r.Parent.GetFullPath() + " > " + r.Title
}

// IsAIAnalyzed 判断是否已完成AI分析
func (r *NesmaRequirement) IsAIAnalyzed() bool {
	return r.AIAnalysisStatus == "completed"
}

// IsAIAnalyzing 判断是否正在AI分析中
func (r *NesmaRequirement) IsAIAnalyzing() bool {
	return r.AIAnalysisStatus == "analyzing"
}

// GetAIAnalysisStatusText 获取AI分析状态中文描述
func (r *NesmaRequirement) GetAIAnalysisStatusText() string {
	switch r.AIAnalysisStatus {
	case "pending":
		return "待分析"
	case "analyzing":
		return "分析中"
	case "completed":
		return "已完成"
	case "failed":
		return "失败"
	default:
		return "未知状态"
	}
}

// GetDisplayTitle 获取显示标题（优先使用AI优化的标题）
func (r *NesmaRequirement) GetDisplayTitle() string {
	if r.AIGeneratedTitle != "" {
		return r.AIGeneratedTitle
	}
	return r.Title
}

// GetDisplayDescription 获取显示描述（优先使用AI生成的描述）
func (r *NesmaRequirement) GetDisplayDescription() string {
	if r.AIDescription != "" {
		return r.AIDescription
	}
	return r.Description
}

// GetRecommendedFunctionPoints 获取推荐的功能点数
func (r *NesmaRequirement) GetRecommendedFunctionPoints() (afp, ufp float64) {
	if r.RecommendedAFP != nil && r.RecommendedUFP != nil {
		return *r.RecommendedAFP, *r.RecommendedUFP
	}
	return r.AFP, r.UFP
}

// HasAIRecommendations 判断是否有AI推荐
func (r *NesmaRequirement) HasAIRecommendations() bool {
	return r.RecommendedAFP != nil || r.RecommendedUFP != nil || r.AIComplexityScore != nil
}

// SetAIAnalysisCompleted 设置AI分析完成
func (r *NesmaRequirement) SetAIAnalysisCompleted() {
	r.AIAnalysisStatus = "completed"
	now := time.Now()
	r.AIAnalysisTime = &now
}

// SetAIAnalysisFailed 设置AI分析失败
func (r *NesmaRequirement) SetAIAnalysisFailed() {
	r.AIAnalysisStatus = "failed"
}

// SetAIAnalyzing 设置AI分析中
func (r *NesmaRequirement) SetAIAnalyzing() {
	r.AIAnalysisStatus = "analyzing"
}
