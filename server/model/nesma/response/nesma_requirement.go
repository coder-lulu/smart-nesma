package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
)

// NesmaRequirementResponse 需求响应
type NesmaRequirementResponse struct {
	nesma.NesmaRequirement
	LevelName string `json:"levelName"`
	FullPath  string `json:"fullPath"`
}

// NesmaRequirementTreeResponse 需求树形响应
type NesmaRequirementTreeResponse struct {
	nesma.NesmaRequirement
	LevelName string                         `json:"levelName"`
	FullPath  string                         `json:"fullPath"`
	Children  []NesmaRequirementTreeResponse `json:"children"`
}

// NesmaRequirementListResponse 需求列表响应
type NesmaRequirementListResponse struct {
	List     []NesmaRequirementResponse `json:"list"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
}

// NesmaRequirementStatsResponse 需求统计响应
type NesmaRequirementStatsResponse struct {
	ProjectID       *uint                        `json:"projectId"` // 为nil时表示全部项目
	TotalCount      int64                        `json:"totalCount"`
	LevelStats      []RequirementLevelStats      `json:"levelStats"`
	StatusStats     []RequirementStatusStats     `json:"statusStats"`
	PriorityStats   []RequirementPriorityStats   `json:"priorityStats"`
	ComplexityStats []RequirementComplexityStats `json:"complexityStats"`
}

// RequirementLevelStats 层级统计
type RequirementLevelStats struct {
	Level     int    `json:"level"`
	LevelName string `json:"levelName"`
	Count     int64  `json:"count"`
}

// RequirementStatusStats 状态统计
type RequirementStatusStats struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
}

// RequirementPriorityStats 优先级统计
type RequirementPriorityStats struct {
	Priority int   `json:"priority"`
	Count    int64 `json:"count"`
}

// RequirementComplexityStats 复杂度统计
type RequirementComplexityStats struct {
	Complexity string `json:"complexity"`
	Count      int64  `json:"count"`
}

// ImportResultResponse 导入结果响应
type ImportResultResponse struct {
	Success     bool     `json:"success"`
	TotalRows   int      `json:"totalRows"`
	SuccessRows int      `json:"successRows"`
	FailedRows  int      `json:"failedRows"`
	Errors      []string `json:"errors"`
	ImportBatch string   `json:"importBatch"`
}

// ParentRequirementOption 父需求选项
type ParentRequirementOption struct {
	ID       uint   `json:"ID"`
	Title    string `json:"title"`
	Level    int    `json:"level"`
	FullPath string `json:"fullPath"`
}

// ParentRequirementOptionsResponse 父需求选项响应
type ParentRequirementOptionsResponse struct {
	Options []ParentRequirementOption `json:"options"`
}

// RequirementChildrenInfo 子功能点信息
type RequirementChildrenInfo struct {
	ID       uint   `json:"ID"`
	Title    string `json:"title"`
	Level    int    `json:"level"`
	Code     string `json:"code"`
	Status   string `json:"status"`
}

// RequirementChildrenResponse 子功能点响应
type RequirementChildrenResponse struct {
	HasChildren     bool                      `json:"hasChildren"`     // 是否有子功能点
	ChildrenCount   int                       `json:"childrenCount"`   // 子功能点总数
	DirectChildren  []RequirementChildrenInfo `json:"directChildren"`  // 直接子功能点
	AllChildren     []RequirementChildrenInfo `json:"allChildren"`     // 所有子功能点（递归）
	LevelBreakdown  map[string]int            `json:"levelBreakdown"`  // 各层级数量统计
}

// RequirementAIAnalysisResponse 需求AI分析结果响应
type RequirementAIAnalysisResponse struct {
	RequirementID       uint    `json:"requirementId"`
	OriginalTitle       string  `json:"originalTitle"`
	OriginalDescription string  `json:"originalDescription"`
	AIAnalysisStatus    string  `json:"aiAnalysisStatus"`
	AIGeneratedTitle    string  `json:"aiGeneratedTitle"`
	AIDescription       string  `json:"aiDescription"`
	AIComplexityScore   *float64 `json:"aiComplexityScore"`
	AIConfidenceScore   *float64 `json:"aiConfidenceScore"`
	RecommendedAFP      *float64 `json:"recommendedAFP"`
	RecommendedUFP      *float64 `json:"recommendedUFP"`
	FunctionType        string  `json:"functionType"`
	AIAnalysisTime      *string `json:"aiAnalysisTime"`
	OptimizationNotes   string  `json:"optimizationNotes"`
	SimilarRequirements []SimilarRequirement `json:"similarRequirements"`
}

// SimilarRequirement 相似需求
type SimilarRequirement struct {
	ID          uint    `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Similarity  float64 `json:"similarity"`
	ProjectName string  `json:"projectName"`
}

// ProjectAIAnalysisStats 项目AI分析统计
type ProjectAIAnalysisStats struct {
	ProjectID             uint                     `json:"projectId"`
	CycleID               *uint                    `json:"cycleId"`
	TotalRequirements     int64                    `json:"totalRequirements"`
	AnalyzedRequirements  int64                    `json:"analyzedRequirements"`
	PendingRequirements   int64                    `json:"pendingRequirements"`
	AnalyzingRequirements int64                    `json:"analyzingRequirements"`
	FailedRequirements    int64                    `json:"failedRequirements"`
	AnalysisProgress      float64                  `json:"analysisProgress"`
	StatusDistribution    []AIAnalysisStatusStats  `json:"statusDistribution"`
	LevelDistribution     []AIAnalysisLevelStats   `json:"levelDistribution"`
	ComplexityDistribution []AIComplexityStats     `json:"complexityDistribution"`
	AverageConfidence     float64                  `json:"averageConfidence"`
	OptimizationSummary   AIOptimizationSummary    `json:"optimizationSummary"`
}

// AIAnalysisStatusStats AI分析状态统计
type AIAnalysisStatusStats struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
	Percentage float64 `json:"percentage"`
}

// AIAnalysisLevelStats AI分析层级统计
type AIAnalysisLevelStats struct {
	Level      int     `json:"level"`
	LevelName  string  `json:"levelName"`
	Total      int64   `json:"total"`
	Analyzed   int64   `json:"analyzed"`
	Percentage float64 `json:"percentage"`
}

// AIComplexityStats AI复杂度统计
type AIComplexityStats struct {
	ComplexityRange string  `json:"complexityRange"` // "0-20", "21-40", "41-60", "61-80", "81-100"
	Count           int64   `json:"count"`
	Percentage      float64 `json:"percentage"`
}

// AIOptimizationSummary AI优化总结
type AIOptimizationSummary struct {
	TotalOptimizations    int64   `json:"totalOptimizations"`
	TitleOptimizations    int64   `json:"titleOptimizations"`
	DescriptionOptimizations int64 `json:"descriptionOptimizations"`
	FunctionPointUpdates  int64   `json:"functionPointUpdates"`
	AverageImprovement    float64 `json:"averageImprovement"`
}

// RequirementVersionCompareResponse 需求版本对比响应
type RequirementVersionCompareResponse struct {
	RequirementID uint                    `json:"requirementId"`
	Version1      RequirementVersionInfo  `json:"version1"`
	Version2      RequirementVersionInfo  `json:"version2"`
	Differences   []RequirementDifference `json:"differences"`
	Summary       VersionCompareSummary   `json:"summary"`
}

// RequirementVersionInfo 需求版本信息
type RequirementVersionInfo struct {
	VersionID         uint     `json:"versionId"`
	Version           string   `json:"version"`
	Title             string   `json:"title"`
	Description       string   `json:"description"`
	AIAnalysisStatus  string   `json:"aiAnalysisStatus"`
	AIGeneratedTitle  string   `json:"aiGeneratedTitle"`
	AIDescription     string   `json:"aiDescription"`
	AIComplexityScore *float64 `json:"aiComplexityScore"`
	AIConfidenceScore *float64 `json:"aiConfidenceScore"`
	RecommendedAFP    *float64 `json:"recommendedAFP"`
	RecommendedUFP    *float64 `json:"recommendedUFP"`
	FunctionType      string   `json:"functionType"`
	CreatedAt         string   `json:"createdAt"`
}

// RequirementDifference 需求差异
type RequirementDifference struct {
	Field       string `json:"field"`
	FieldName   string `json:"fieldName"`
	OldValue    string `json:"oldValue"`
	NewValue    string `json:"newValue"`
	ChangeType  string `json:"changeType"` // "added", "removed", "modified"
	Significance string `json:"significance"` // "low", "medium", "high"
}

// VersionCompareSummary 版本对比总结
type VersionCompareSummary struct {
	TotalDifferences   int     `json:"totalDifferences"`
	HighSignificance   int     `json:"highSignificance"`
	MediumSignificance int     `json:"mediumSignificance"`
	LowSignificance    int     `json:"lowSignificance"`
	OverallChange      float64 `json:"overallChange"` // 整体变化程度 (0-100)
	RecommendAction    string  `json:"recommendAction"` // "accept", "review", "reject"
}

// AIOptimizedRequirementsResponse AI优化需求列表响应
type AIOptimizedRequirementsResponse struct {
	List     []AIOptimizedRequirement `json:"list"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"pageSize"`
	Summary  AIOptimizedSummary       `json:"summary"`
}

// AIOptimizedRequirement AI优化的需求
type AIOptimizedRequirement struct {
	nesma.NesmaRequirement
	LevelName           string               `json:"levelName"`
	FullPath            string               `json:"fullPath"`
	HasAIOptimization   bool                 `json:"hasAIOptimization"`
	OptimizationScore   float64              `json:"optimizationScore"`
	OptimizationChanges []OptimizationChange `json:"optimizationChanges"`
}

// OptimizationChange 优化变更
type OptimizationChange struct {
	Type        string `json:"type"`        // "title", "description", "afp", "ufp", "complexity"
	OriginalValue string `json:"originalValue"`
	OptimizedValue string `json:"optimizedValue"`
	Improvement   float64 `json:"improvement"` // 改进程度 (0-100)
}

// AIOptimizedSummary AI优化总结
type AIOptimizedSummary struct {
	TotalCount            int64   `json:"totalCount"`
	OptimizedCount        int64   `json:"optimizedCount"`
	PendingCount          int64   `json:"pendingCount"`
	FailedCount           int64   `json:"failedCount"`
	OptimizationRate      float64 `json:"optimizationRate"`
	AverageImprovement    float64 `json:"averageImprovement"`
	TotalFunctionPoints   float64 `json:"totalFunctionPoints"`
	OptimizedFunctionPoints float64 `json:"optimizedFunctionPoints"`
}
