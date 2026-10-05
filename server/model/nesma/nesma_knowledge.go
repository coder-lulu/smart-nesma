package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaKnowledgeEntry 知识库条目表
type NesmaKnowledgeEntry struct {
	global.GVA_MODEL
	Title           string         `json:"title" gorm:"type:varchar(500);not null;comment:知识条目标题"`
	Content         string         `json:"content" gorm:"type:text;not null;comment:知识内容"`
	Category        string         `json:"category" gorm:"type:varchar(100);not null;comment:分类：NESMA_STANDARD,BEST_PRACTICE,CASE_STUDY,RULE"`
	Domain          string         `json:"domain" gorm:"type:varchar(100);comment:适用领域"`
	Tags            datatypes.JSON `json:"tags" gorm:"type:json;comment:标签"`
	Embedding       string         `json:"embedding" gorm:"type:text;comment:向量表示(JSON格式)"`
	ConfidenceScore float64        `json:"confidenceScore" gorm:"default:1.0;comment:置信度评分"`
	UsageCount      int            `json:"usageCount" gorm:"default:0;comment:使用次数"`
	Version         string         `json:"version" gorm:"type:varchar(50);comment:版本号"`
	Status          string         `json:"status" gorm:"type:varchar(50);default:'active';comment:状态"`
	Source          string         `json:"source" gorm:"type:varchar(255);comment:来源"`
	Author          string         `json:"author" gorm:"type:varchar(255);comment:作者"`
}

// TableName 自定义表名
func (NesmaKnowledgeEntry) TableName() string {
	return "nesma_knowledge_entries"
}

// NesmaKnowledgeRule 知识库规则表
type NesmaKnowledgeRule struct {
	global.GVA_MODEL
	RuleName            string `json:"ruleName" gorm:"type:varchar(255);not null;comment:规则名称"`
	ConditionExpression string `json:"conditionExpression" gorm:"type:text;not null;comment:规则条件表达式"`
	ActionExpression    string `json:"actionExpression" gorm:"type:text;not null;comment:规则动作表达式"`
	Priority            int    `json:"priority" gorm:"default:1;comment:优先级"`
	IsActive            bool   `json:"isActive" gorm:"default:true;comment:是否激活"`
	Description         string `json:"description" gorm:"type:text;comment:规则描述"`
	Category            string `json:"category" gorm:"type:varchar(100);comment:规则分类"`
}

// TableName 自定义表名
func (NesmaKnowledgeRule) TableName() string {
	return "nesma_knowledge_rules"
}

// NesmaCaseStudy 案例库表
type NesmaCaseStudy struct {
	global.GVA_MODEL
	ProjectName         string         `json:"projectName" gorm:"type:varchar(255);not null;comment:项目名称"`
	Domain              string         `json:"domain" gorm:"type:varchar(100);comment:项目领域"`
	RequirementsSummary string         `json:"requirementsSummary" gorm:"type:text;comment:需求摘要"`
	NesmaResults        datatypes.JSON `json:"nesmaResults" gorm:"type:json;comment:NESMA评估结果"`
	LessonsLearned      string         `json:"lessonsLearned" gorm:"type:text;comment:经验教训"`
	Embedding           string         `json:"embedding" gorm:"type:text;comment:向量表示(JSON格式)"`
	TotalFunctionPoints float64        `json:"totalFunctionPoints" gorm:"comment:总功能点数"`
	ProjectDuration     int            `json:"projectDuration" gorm:"comment:项目持续时间(天)"`
	TeamSize            int            `json:"teamSize" gorm:"comment:团队规模"`
	Complexity          string         `json:"complexity" gorm:"type:varchar(50);comment:项目复杂度"`
	Success             bool           `json:"success" gorm:"default:true;comment:项目是否成功"`
}

// TableName 自定义表名
func (NesmaCaseStudy) TableName() string {
	return "nesma_case_studies"
}
