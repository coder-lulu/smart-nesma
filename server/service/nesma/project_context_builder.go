package nesma

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// ProjectContextBuilder 项目上下文构建器
type ProjectContextBuilder struct {
	projectID    uint
	evaluationID uint
}

// ProjectAnalysisContext 项目分析上下文
type ProjectAnalysisContext struct {
	// 基础项目信息
	Project         *nesma.NesmaProject         `json:"project"`
	Evaluation      *nesma.NesmaEvaluation      `json:"evaluation"`
	ProjectCycle    *nesma.NesmaProjectCycle    `json:"projectCycle,omitempty"`
	RequirementVersion *nesma.NesmaRequirementVersion `json:"requirementVersion,omitempty"`
	
	// 需求相关
	Requirements    []nesma.NesmaRequirement    `json:"requirements"`
	RequirementStats RequirementStatistics      `json:"requirementStats"`
	RequirementTree RequirementTreeNode        `json:"requirementTree"`
	
	// 功能点相关
	ExistingFunctionPoints []nesma.NesmaFunctionPoint `json:"existingFunctionPoints,omitempty"`
	FunctionPointStats     FunctionPointStatistics    `json:"functionPointStats"`
	
	// 知识库相关 (暂时注释，等待知识库功能完善)
	// RelevantKnowledge []nesma.NesmaKnowledge       `json:"relevantKnowledge,omitempty"`
	
	// 项目统计和度量
	ProjectMetrics  ProjectMetrics               `json:"projectMetrics"`
	
	// 历史数据
	HistoricalEvaluations []nesma.NesmaEvaluation   `json:"historicalEvaluations,omitempty"`
	
	// 分析配置
	AnalysisConfig  AnalysisConfiguration        `json:"analysisConfig"`
}

// RequirementStatistics 需求统计
type RequirementStatistics struct {
	TotalCount       int                    `json:"totalCount"`
	LevelDistribution map[int]int           `json:"levelDistribution"`
	TypeDistribution map[string]int        `json:"typeDistribution"`
	FunctionTypeDistribution map[string]int `json:"functionTypeDistribution"`
	ComplexityDistribution map[string]int   `json:"complexityDistribution"`
	ModificationTypeDistribution map[string]int `json:"modificationTypeDistribution"`
	ReuseDistribution map[string]int        `json:"reuseDistribution"`
	AverageDescriptionLength float64        `json:"averageDescriptionLength"`
	KeywordFrequency map[string]int         `json:"keywordFrequency"`
}

// RequirementTreeNode 需求树节点
type RequirementTreeNode struct {
	Requirement nesma.NesmaRequirement   `json:"requirement"`
	Children    []RequirementTreeNode    `json:"children"`
}

// FunctionPointStatistics 功能点统计
type FunctionPointStatistics struct {
	TotalFunctionPoints int                    `json:"totalFunctionPoints"`
	TypeDistribution    map[string]int         `json:"typeDistribution"`
	ComplexityDistribution map[string]int      `json:"complexityDistribution"`
	AverageComplexity   float64                `json:"averageComplexity"`
	QualityMetrics      map[string]float64     `json:"qualityMetrics"`
}

// ProjectMetrics 项目度量
type ProjectMetrics struct {
	ProjectSize         string                 `json:"projectSize"`
	EstimatedDuration   int                    `json:"estimatedDuration"`
	TeamSize            int                    `json:"teamSize"`
	TechnologyComplexity string                `json:"technologyComplexity"`
	BusinessDomain      string                 `json:"businessDomain"`
	IntegrationComplexity string               `json:"integrationComplexity"`
	RiskLevel           string                 `json:"riskLevel"`
	InnovationLevel     string                 `json:"innovationLevel"`
}

// AnalysisConfiguration 分析配置
type AnalysisConfiguration struct {
	AnalysisDepth      string   `json:"analysisDepth"`      // comprehensive, detailed, basic
	FocusAreas         []string `json:"focusAreas"`         // functional, technical, business, quality, risk
	AIModel            string   `json:"aiModel"`
	MaxTokens          int      `json:"maxTokens"`
	Temperature        float64  `json:"temperature"`
	EnableBatchProcessing bool  `json:"enableBatchProcessing"`
	BatchSize          int      `json:"batchSize"`
	IncludeHistoricalData bool  `json:"includeHistoricalData"`
	ComplianceStandards []string `json:"complianceStandards"`
}

// NewProjectContextBuilder 创建项目上下文构建器
func NewProjectContextBuilder(projectID, evaluationID uint) *ProjectContextBuilder {
	return &ProjectContextBuilder{
		projectID:    projectID,
		evaluationID: evaluationID,
	}
}

// BuildContext 构建完整的项目分析上下文
func (b *ProjectContextBuilder) BuildContext(config AnalysisConfiguration) (*ProjectAnalysisContext, error) {
	global.GVA_LOG.Info("开始构建项目分析上下文",
		zap.Uint("projectID", b.projectID),
		zap.Uint("evaluationID", b.evaluationID))

	context := &ProjectAnalysisContext{
		AnalysisConfig: config,
	}

	// 1. 获取基础项目信息
	if err := b.loadBasicProjectInfo(context); err != nil {
		return nil, fmt.Errorf("加载基础项目信息失败: %w", err)
	}

	// 2. 获取需求信息
	if err := b.loadRequirementInfo(context); err != nil {
		return nil, fmt.Errorf("加载需求信息失败: %w", err)
	}

	// 3. 获取功能点信息
	if err := b.loadFunctionPointInfo(context); err != nil {
		return nil, fmt.Errorf("加载功能点信息失败: %w", err)
	}

	// 4. 获取相关知识库信息 (暂时跳过)
	// if err := b.loadRelevantKnowledge(context); err != nil {
	// 	global.GVA_LOG.Warn("加载知识库信息失败", zap.Error(err))
	// }

	// 5. 计算项目度量
	b.calculateProjectMetrics(context)

	// 6. 获取历史数据（如果启用）
	if config.IncludeHistoricalData {
		if err := b.loadHistoricalData(context); err != nil {
			global.GVA_LOG.Warn("加载历史数据失败", zap.Error(err))
		}
	}

	global.GVA_LOG.Info("项目分析上下文构建完成",
		zap.Int("requirementCount", len(context.Requirements)),
		zap.Int("functionPointCount", len(context.ExistingFunctionPoints)))

	return context, nil
}

// loadBasicProjectInfo 加载基础项目信息
func (b *ProjectContextBuilder) loadBasicProjectInfo(context *ProjectAnalysisContext) error {
	// 加载项目信息
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, b.projectID).Error; err != nil {
		return err
	}
	context.Project = &project

	// 加载评估信息
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("Cycle").Preload("RequirementVersion").
		First(&evaluation, b.evaluationID).Error; err != nil {
		return err
	}
	context.Evaluation = &evaluation

	// 加载关联的周期和版本信息
	if evaluation.CycleID != nil {
		context.ProjectCycle = evaluation.Cycle
	}
	if evaluation.RequirementVersionID != nil {
		context.RequirementVersion = evaluation.RequirementVersion
	}

	return nil
}

// loadRequirementInfo 加载需求信息
func (b *ProjectContextBuilder) loadRequirementInfo(context *ProjectAnalysisContext) error {
	// 构建需求查询
	query := global.GVA_DB.Where("project_id = ?", b.projectID)
	
	// 根据评估配置过滤需求范围
	if context.Evaluation.CycleID != nil {
		query = query.Where("cycle_id = ?", *context.Evaluation.CycleID)
	}
	if context.Evaluation.RequirementVersionID != nil {
		query = query.Where("version_id = ?", *context.Evaluation.RequirementVersionID)
	}

	// 获取需求列表
	var requirements []nesma.NesmaRequirement
	if err := query.Order("level, code").Find(&requirements).Error; err != nil {
		return err
	}
	context.Requirements = requirements

	// 计算需求统计
	context.RequirementStats = b.calculateRequirementStatistics(requirements)

	// 构建需求树
	context.RequirementTree = b.buildRequirementTree(requirements)

	return nil
}

// loadFunctionPointInfo 加载功能点信息
func (b *ProjectContextBuilder) loadFunctionPointInfo(context *ProjectAnalysisContext) error {
	// 获取已存在的功能点
	var functionPoints []nesma.NesmaFunctionPoint
	if err := global.GVA_DB.Where("evaluation_id = ?", b.evaluationID).
		Find(&functionPoints).Error; err != nil {
		return err
	}
	context.ExistingFunctionPoints = functionPoints

	// 计算功能点统计
	context.FunctionPointStats = b.calculateFunctionPointStatistics(functionPoints)

	return nil
}

// loadRelevantKnowledge 加载相关知识库信息 (暂时注释)
// func (b *ProjectContextBuilder) loadRelevantKnowledge(context *ProjectAnalysisContext) error {
// 	// 基于项目领域和类型获取相关知识
// 	var knowledge []nesma.NesmaKnowledge
// 	query := global.GVA_DB.Where("status = ?", "active")
// 	
// 	if context.Project.Domain != "" {
// 		query = query.Where("domain LIKE ? OR domain = ''", "%"+context.Project.Domain+"%")
// 	}
// 	
// 	if err := query.Limit(50).Find(&knowledge).Error; err != nil {
// 		return err
// 	}
// 	context.RelevantKnowledge = knowledge
//
// 	return nil
// }

// calculateProjectMetrics 计算项目度量
func (b *ProjectContextBuilder) calculateProjectMetrics(context *ProjectAnalysisContext) {
	metrics := ProjectMetrics{
		BusinessDomain: context.Project.Domain,
	}

	// 基于需求数量判断项目规模
	requirementCount := len(context.Requirements)
	switch {
	case requirementCount < 50:
		metrics.ProjectSize = "small"
	case requirementCount < 200:
		metrics.ProjectSize = "medium"
	case requirementCount < 500:
		metrics.ProjectSize = "large"
	default:
		metrics.ProjectSize = "enterprise"
	}

	// 基于功能复杂度判断技术复杂度
	avgComplexity := context.FunctionPointStats.AverageComplexity
	switch {
	case avgComplexity < 0.3:
		metrics.TechnologyComplexity = "low"
	case avgComplexity < 0.7:
		metrics.TechnologyComplexity = "medium"
	default:
		metrics.TechnologyComplexity = "high"
	}

	// 评估集成复杂度
	eifCount := context.RequirementStats.FunctionTypeDistribution["EIF"]
	if eifCount > 10 {
		metrics.IntegrationComplexity = "high"
	} else if eifCount > 5 {
		metrics.IntegrationComplexity = "medium"
	} else {
		metrics.IntegrationComplexity = "low"
	}

	context.ProjectMetrics = metrics
}

// loadHistoricalData 加载历史数据
func (b *ProjectContextBuilder) loadHistoricalData(context *ProjectAnalysisContext) error {
	var historicalEvaluations []nesma.NesmaEvaluation
	if err := global.GVA_DB.Where("project_id = ? AND id != ?", b.projectID, b.evaluationID).
		Order("created_at DESC").Limit(5).Find(&historicalEvaluations).Error; err != nil {
		return err
	}
	context.HistoricalEvaluations = historicalEvaluations
	return nil
}

// calculateRequirementStatistics 计算需求统计
func (b *ProjectContextBuilder) calculateRequirementStatistics(requirements []nesma.NesmaRequirement) RequirementStatistics {
	stats := RequirementStatistics{
		TotalCount:                   len(requirements),
		LevelDistribution:            make(map[int]int),
		TypeDistribution:             make(map[string]int),
		FunctionTypeDistribution:     make(map[string]int),
		ComplexityDistribution:       make(map[string]int),
		ModificationTypeDistribution: make(map[string]int),
		ReuseDistribution:            make(map[string]int),
		KeywordFrequency:             make(map[string]int),
	}

	totalLength := 0
	keywords := []string{"用户", "系统", "数据", "查询", "添加", "删除", "修改", "管理", "报表", "接口"}

	for _, req := range requirements {
		// 层级分布
		stats.LevelDistribution[req.Level]++

		// 功能类型分布
		if req.FunctionType != "" {
			stats.FunctionTypeDistribution[req.FunctionType]++
		}

		// 修改类型分布
		if req.ModificationType != "" {
			stats.ModificationTypeDistribution[req.ModificationType]++
		}

		// 重用程度分布
		if req.ReuseLevel != "" {
			stats.ReuseDistribution[req.ReuseLevel]++
		}

		// 描述长度统计
		totalLength += len(req.Description)

		// 关键词频率
		content := strings.ToLower(req.Title + " " + req.Description)
		for _, keyword := range keywords {
			if strings.Contains(content, keyword) {
				stats.KeywordFrequency[keyword]++
			}
		}
	}

	if len(requirements) > 0 {
		stats.AverageDescriptionLength = float64(totalLength) / float64(len(requirements))
	}

	return stats
}

// buildRequirementTree 构建需求树
func (b *ProjectContextBuilder) buildRequirementTree(requirements []nesma.NesmaRequirement) RequirementTreeNode {
	// 创建需求映射
	reqMap := make(map[uint]*RequirementTreeNode)
	var roots []RequirementTreeNode

	// 初始化所有节点
	for _, req := range requirements {
		node := RequirementTreeNode{
			Requirement: req,
			Children:    []RequirementTreeNode{},
		}
		reqMap[req.ID] = &node
	}

	// 构建树结构
	for _, req := range requirements {
		node := reqMap[req.ID]
		if req.ParentID == nil {
			// 根节点
			roots = append(roots, *node)
		} else {
			// 子节点
			if parent, exists := reqMap[*req.ParentID]; exists {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	// 返回虚拟根节点
	return RequirementTreeNode{
		Requirement: nesma.NesmaRequirement{
			Title: "项目需求根节点",
			Level: 0,
		},
		Children: roots,
	}
}

// calculateFunctionPointStatistics 计算功能点统计
func (b *ProjectContextBuilder) calculateFunctionPointStatistics(functionPoints []nesma.NesmaFunctionPoint) FunctionPointStatistics {
	stats := FunctionPointStatistics{
		TotalFunctionPoints:    len(functionPoints),
		TypeDistribution:       make(map[string]int),
		ComplexityDistribution: make(map[string]int),
		QualityMetrics:         make(map[string]float64),
	}

	if len(functionPoints) == 0 {
		return stats
	}

	totalComplexity := 0.0
	totalConfidence := 0.0

	for _, fp := range functionPoints {
		// 类型分布
		stats.TypeDistribution[fp.FunctionType]++

		// 复杂度分布
		stats.ComplexityDistribution[fp.ComplexityLevel]++

		// 累计计算
		switch fp.ComplexityLevel {
		case "Low":
			totalComplexity += 1.0
		case "Average":
			totalComplexity += 2.0
		case "High":
			totalComplexity += 3.0
		}

		totalConfidence += fp.ConfidenceLevel
	}

	stats.AverageComplexity = totalComplexity / float64(len(functionPoints))
	stats.QualityMetrics["averageConfidence"] = totalConfidence / float64(len(functionPoints))

	return stats
}

// GetContextSummary 获取上下文摘要
func (context *ProjectAnalysisContext) GetContextSummary() string {
	summary := fmt.Sprintf(`项目分析上下文摘要:
项目名称: %s
项目领域: %s  
评估类型: %s
需求总数: %d
功能点总数: %d
项目规模: %s
技术复杂度: %s
集成复杂度: %s`,
		context.Project.Name,
		context.Project.Domain,
		context.Evaluation.EvaluationType,
		context.RequirementStats.TotalCount,
		context.FunctionPointStats.TotalFunctionPoints,
		context.ProjectMetrics.ProjectSize,
		context.ProjectMetrics.TechnologyComplexity,
		context.ProjectMetrics.IntegrationComplexity)

	return summary
}

// ToJSON 转换为JSON格式
func (context *ProjectAnalysisContext) ToJSON() (string, error) {
	jsonData, err := json.MarshalIndent(context, "", "  ")
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

// GetRequirementsByLevel 按层级获取需求
func (context *ProjectAnalysisContext) GetRequirementsByLevel(level int) []nesma.NesmaRequirement {
	var filtered []nesma.NesmaRequirement
	for _, req := range context.Requirements {
		if req.Level == level {
			filtered = append(filtered, req)
		}
	}
	return filtered
}

// GetRequirementsByFunctionType 按功能类型获取需求
func (context *ProjectAnalysisContext) GetRequirementsByFunctionType(functionType string) []nesma.NesmaRequirement {
	var filtered []nesma.NesmaRequirement
	for _, req := range context.Requirements {
		if req.FunctionType == functionType {
			filtered = append(filtered, req)
		}
	}
	return filtered
}