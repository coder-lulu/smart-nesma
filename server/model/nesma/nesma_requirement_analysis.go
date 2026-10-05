package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaRequirementAnalysis NESMA需求分析结果表
type NesmaRequirementAnalysis struct {
	global.GVA_MODEL
	RequirementID     uint           `json:"requirementId" gorm:"not null;comment:需求ID"`
	AIExpandedContent string         `json:"aiExpandedContent" gorm:"type:text;comment:AI扩展的需求内容"`
	NesmaCategory     string         `json:"nesmaCategory" gorm:"type:varchar(100);comment:NESMA分类(ILF,EIF,EI,EO,EQ)"`
	ComplexityScore   float64        `json:"complexityScore" gorm:"comment:复杂度评分"`
	FunctionPoints    float64        `json:"functionPoints" gorm:"comment:功能点数"`
	ConfidenceLevel   float64        `json:"confidenceLevel" gorm:"comment:置信度"`
	KnowledgeBaseRefs datatypes.JSON `json:"knowledgeBaseRefs" gorm:"type:json;comment:引用的知识库条目"`
	AnalysisVersion   string         `json:"analysisVersion" gorm:"type:varchar(50);comment:分析版本"`
	AnalysisDetails   datatypes.JSON `json:"analysisDetails" gorm:"type:json;comment:详细分析结果"`
	ReviewStatus      string         `json:"reviewStatus" gorm:"type:varchar(50);default:'pending';comment:审核状态"`
	ReviewComments    string         `json:"reviewComments" gorm:"type:text;comment:审核意见"`

	// 关联关系
	Requirement NesmaRequirement `json:"requirement" gorm:"foreignKey:RequirementID"`
}

// TableName 自定义表名
func (NesmaRequirementAnalysis) TableName() string {
	return "nesma_requirement_analysis"
}
