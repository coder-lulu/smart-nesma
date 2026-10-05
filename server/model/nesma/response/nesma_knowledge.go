package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
)

// ListKnowledgeEntriesResponse 列出知识条目响应
type ListKnowledgeEntriesResponse struct {
	Entries  []nesma.NesmaKnowledgeEntry `json:"entries"`
	Total    int64                       `json:"total"`
	Page     int                         `json:"page"`
	PageSize int                         `json:"pageSize"`
}

// SearchKnowledgeEntriesResponse 搜索知识条目响应
type SearchKnowledgeEntriesResponse struct {
	Results []SimilarKnowledge `json:"results"`
	Query   string             `json:"query"`
	Count   int                `json:"count"`
}

// SimilarKnowledge 相似知识结构
type SimilarKnowledge struct {
	Knowledge  nesma.NesmaKnowledgeEntry `json:"knowledge"`
	Similarity float32                   `json:"similarity"`
}

// ListKnowledgeRulesResponse 列出知识规则响应
type ListKnowledgeRulesResponse struct {
	Rules    []nesma.NesmaKnowledgeRule `json:"rules"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
}

// KnowledgeStatisticsResponse 知识库统计响应
type KnowledgeStatisticsResponse struct {
	TotalEntries  int64            `json:"totalEntries"`
	TotalRules    int64            `json:"totalRules"`
	CategoryStats map[string]int64 `json:"categoryStats"`
	DomainStats   map[string]int64 `json:"domainStats"`
}

// ListCaseStudiesResponse 列出案例研究响应
type ListCaseStudiesResponse struct {
	Cases    []nesma.NesmaCaseStudy `json:"cases"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"pageSize"`
}

// SearchCaseStudiesResponse 搜索案例研究响应
type SearchCaseStudiesResponse struct {
	Results []SimilarCase `json:"results"`
	Query   string        `json:"query"`
	Count   int           `json:"count"`
}

// SimilarCase 相似案例结构
type SimilarCase struct {
	Case       nesma.NesmaCaseStudy `json:"case"`
	Similarity float32              `json:"similarity"`
}
