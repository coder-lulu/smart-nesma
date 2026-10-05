package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// AI分析结果数据结构定义
type AIFunctionalAnalysisResult struct {
	TotalFunctionPoints         float64                `json:"totalFunctionPoints"`
	DataFunctionPoints          float64                `json:"dataFunctionPoints"`
	TransactionalFunctionPoints float64                `json:"transactionalFunctionPoints"`
	ILFCount                    int                    `json:"ilfCount"`
	EIFCount                    int                    `json:"eifCount"`
	EICount                     int                    `json:"eiCount"`
	EOCount                     int                    `json:"eoCount"`
	EQCount                     int                    `json:"eqCount"`
	LowComplexityCount          int                    `json:"lowComplexityCount"`
	MediumComplexityCount       int                    `json:"mediumComplexityCount"`
	HighComplexityCount         int                    `json:"highComplexityCount"`
	FunctionalCoverage          float64                `json:"functionalCoverage"`
	FunctionalCompleteness      float64                `json:"functionalCompleteness"`
	FunctionalConsistency       float64                `json:"functionalConsistency"`
	FunctionTypeDistribution    map[string]interface{} `json:"functionTypeDistribution"`
	ComplexityAnalysis          map[string]interface{} `json:"complexityAnalysis"`
	QualityIndicators           map[string]interface{} `json:"qualityIndicators"`
	ConfidenceScore             float64                `json:"confidenceScore"`
}

type AITechnicalAnalysisResult struct {
	ArchitectureType            string                 `json:"architectureType"`
	TechnologyStack             map[string]interface{} `json:"technologyStack"`
	IntegrationComplexity       string                 `json:"integrationComplexity"`
	DevelopmentEffort           float64                `json:"developmentEffort"`
	TestingEffort               float64                `json:"testingEffort"`
	DeploymentComplexity        float64                `json:"deploymentComplexity"`
	TechnicalFeasibility        float64                `json:"technicalFeasibility"`
	ImplementationRisk          float64                `json:"implementationRisk"`
	TechnicalInnovation         float64                `json:"technicalInnovation"`
	MaintenanceComplexity       float64                `json:"maintenanceComplexity"`
	TechnicalRecommendations    string                 `json:"technicalRecommendations"`
	ConfidenceScore             float64                `json:"confidenceScore"`
}

type AIBusinessAnalysisResult struct {
	BusinessValue            float64 `json:"businessValue"`
	ROIEstimation            float64 `json:"roiEstimation"`
	StrategicAlignment       float64 `json:"strategicAlignment"`
	UserExperienceScore      float64 `json:"userExperienceScore"`
	ProcessEfficiency        float64 `json:"processEfficiency"`
	ProcessAutomation        float64 `json:"processAutomation"`
	BusinessLogicComplexity  float64 `json:"businessLogicComplexity"`
	MarketFit                float64 `json:"marketFit"`
	CompetitiveAdvantage     float64 `json:"competitiveAdvantage"`
	InnovationLevel          float64 `json:"innovationLevel"`
	RegulatoryCompliance     float64 `json:"regulatoryCompliance"`
	BusinessRecommendations  string  `json:"businessRecommendations"`
	ConfidenceScore          float64 `json:"confidenceScore"`
}

type AIQualityAnalysisResult struct {
	OverallQuality         float64 `json:"overallQuality"`
	RequirementQuality     float64 `json:"requirementQuality"`
	DesignQuality          float64 `json:"designQuality"`
	Correctness            float64 `json:"correctness"`
	Completeness           float64 `json:"completeness"`
	Consistency            float64 `json:"consistency"`
	Clarity                float64 `json:"clarity"`
	Traceability           float64 `json:"traceability"`
	Maintainability        float64 `json:"maintainability"`
	Modularity             float64 `json:"modularity"`
	Reusability            float64 `json:"reusability"`
	Testability            float64 `json:"testability"`
	QualityRecommendations string  `json:"qualityRecommendations"`
	ConfidenceScore        float64 `json:"confidenceScore"`
}

type AIRiskAnalysisResult struct {
	OverallRisk         float64 `json:"overallRisk"`
	RiskLevel           string  `json:"riskLevel"`
	TechnicalRisk       float64 `json:"technicalRisk"`
	ImplementationRisk  float64 `json:"implementationRisk"`
	IntegrationRisk     float64 `json:"integrationRisk"`
	ScheduleRisk        float64 `json:"scheduleRisk"`
	BudgetRisk          float64 `json:"budgetRisk"`
	ResourceRisk        float64 `json:"resourceRisk"`
	BusinessRisk        float64 `json:"businessRisk"`
	MarketRisk          float64 `json:"marketRisk"`
	ComplianceRisk      float64 `json:"complianceRisk"`
	RiskRecommendations string  `json:"riskRecommendations"`
	ConfidenceScore     float64 `json:"confidenceScore"`
}

type AIRecommendationResult struct {
	RecommendationType  string  `json:"recommendationType"`
	Priority            string  `json:"priority"`
	ImpactLevel         string  `json:"impactLevel"`
	Title               string  `json:"title"`
	Description         string  `json:"description"`
	Rationale           string  `json:"rationale"`
	EstimatedEffort     float64 `json:"estimatedEffort"`
	ExpectedBenefit     string  `json:"expectedBenefit"`
	ConfidenceScore     float64 `json:"confidenceScore"`
	Urgency             string  `json:"urgency"`
}

type AIComplianceAnalysisResult struct {
	ComplianceStandard        string  `json:"complianceStandard"`
	StandardVersion           string  `json:"standardVersion"`
	OverallCompliance         float64 `json:"overallCompliance"`
	ComplianceLevel           string  `json:"complianceLevel"`
	ComplianceRecommendations string  `json:"complianceRecommendations"`
	ConfidenceScore           float64 `json:"confidenceScore"`
}

// buildFunctionalAnalysisPrompt 构建功能分析提示词
func (engine *AIAnalysisEngine) buildFunctionalAnalysisPrompt(context *ProjectAnalysisContext) string {
	requirementSummary := engine.buildRequirementSummary(context.Requirements)
	
	prompt := fmt.Sprintf(`# NESMA功能性分析专家任务

## 专家背景
您是一位拥有20年以上经验的国际认证NESMA功能点分析专家，具备ISO/IEC 14143标准认证资质，在软件度量、功能点识别、复杂度评估方面有深厚造诣。您精通NESMA 2.2国际标准，在金融、制造、医疗、政府、电商等15+个行业有丰富的项目实践经验。

## 分析任务
请对以下项目进行全面的功能性分析，严格按照NESMA 2.2标准，运用专业的分析方法，提供准确、详细的功能点评估和分析。

## 项目基本信息
- **项目名称**: %s
- **业务领域**: %s
- **项目规模**: %s
- **技术复杂度**: %s
- **需求总数**: %d
- **已识别功能点**: %d

## 项目上下文
%s

## 需求分析数据
### 需求层级分布
%s

### 功能类型分布
%s

### 详细需求列表
%s

## 分析要求

### 1. NESMA功能点识别与分类
请按照NESMA 2.2标准，对所有需求进行功能点识别和分类：

**数据功能（Data Functions）**：
- **ILF (内部逻辑文件)**: 由应用维护的用户可识别的逻辑相关数据组
- **EIF (外部接口文件)**: 由其他应用维护但本应用引用的用户可识别的逻辑相关数据组

**事务功能（Transactional Functions）**：
- **EI (外部输入)**: 处理来自应用边界外的数据或控制信息的基本过程
- **EO (外部输出)**: 向应用边界外发送数据或控制信息的基本过程
- **EQ (外部查询)**: 从应用边界外发送输入并接收输出的基本过程

### 2. 复杂度评估
对每种功能类型进行复杂度评估（低、中、高），并计算相应的功能点数：

**复杂度权重表**：
- ILF: 低=7, 中=10, 高=15
- EIF: 低=5, 中=7, 高=10
- EI: 低=3, 中=4, 高=6
- EO: 低=4, 中=5, 高=7
- EQ: 低=3, 中=4, 高=6

### 3. 质量评估
评估功能性要求的质量特征：
- **功能覆盖度**: 需求对业务功能的覆盖程度
- **功能完整性**: 功能需求的完整程度
- **功能一致性**: 功能规范的一致性

### 4. 深度分析
- **功能缺口识别**: 识别潜在的功能缺口和遗漏
- **复杂度分布分析**: 分析复杂度分布的合理性
- **功能依赖关系**: 识别功能间的依赖关系

## 输出格式要求

请严格按照以下JSON格式返回分析结果：

{
  "totalFunctionPoints": 总功能点数,
  "dataFunctionPoints": 数据功能点数,
  "transactionalFunctionPoints": 事务功能点数,
  "ilfCount": ILF数量,
  "eifCount": EIF数量,
  "eiCount": EI数量,
  "eoCount": EO数量,
  "eqCount": EQ数量,
  "lowComplexityCount": 低复杂度功能数,
  "mediumComplexityCount": 中复杂度功能数,
  "highComplexityCount": 高复杂度功能数,
  "functionalCoverage": 功能覆盖度评分(0-1),
  "functionalCompleteness": 功能完整性评分(0-1),
  "functionalConsistency": 功能一致性评分(0-1),
  "functionTypeDistribution": {
    "分析": "功能类型分布的详细分析",
    "recommendations": ["改进建议列表"]
  },
  "complexityAnalysis": {
    "distribution": "复杂度分布分析",
    "patterns": "复杂度模式识别",
    "optimization": "复杂度优化建议"
  },
  "qualityIndicators": {
    "strengths": ["功能性优势列表"],
    "weaknesses": ["功能性不足列表"],
    "improvements": ["改进方向列表"]
  },
  "confidenceScore": 分析置信度(0-1)
}

## 专业要求
1. **严格遵循NESMA 2.2标准**
2. **基于实际需求内容进行分析，不得凭空推测**
3. **提供具体的改进建议和优化方案**
4. **确保分析结果的准确性和专业性**
5. **置信度评估要基于需求的清晰度和完整度**`,
		context.Project.Name,
		context.Project.Domain,
		context.ProjectMetrics.ProjectSize,
		context.ProjectMetrics.TechnologyComplexity,
		context.RequirementStats.TotalCount,
		context.FunctionPointStats.TotalFunctionPoints,
		context.GetContextSummary(),
		engine.formatDistribution(context.RequirementStats.LevelDistribution),
		engine.formatStringDistribution(context.RequirementStats.FunctionTypeDistribution),
		requirementSummary)

	return prompt
}

// buildTechnicalAnalysisPrompt 构建技术分析提示词
func (engine *AIAnalysisEngine) buildTechnicalAnalysisPrompt(context *ProjectAnalysisContext) string {
	
	prompt := fmt.Sprintf(`# 软件技术架构分析专家任务

## 专家背景
您是一位拥有15年以上经验的软件架构师和技术评估专家，精通各种软件架构模式、技术栈选型、系统集成、性能优化等领域。您具备多个行业的大型项目架构设计和技术风险评估经验，对现代软件开发技术、DevOps、云计算、微服务等有深入理解。

## 分析任务
请对以下项目进行全面的技术性分析，从架构设计、技术选型、实现复杂度、开发工作量等多个维度提供专业的技术评估和建议。

## 项目基本信息
- **项目名称**: %s
- **业务领域**: %s
- **项目规模**: %s
- **集成复杂度**: %s
- **需求总数**: %d
- **外部接口数**: %d

## 项目技术上下文
%s

## 需求技术特征
### 数据处理需求
%s

### 集成接口需求
%s

### 性能和非功能需求
%s

## 分析维度

### 1. 架构设计分析
- **架构模式**: 分析适合的架构模式（单体、微服务、分层架构等）
- **技术栈选型**: 评估技术栈的合理性和先进性
- **系统边界**: 分析系统边界和外部依赖

### 2. 技术复杂度评估
- **开发复杂度**: 评估开发实现的技术难度
- **集成复杂度**: 评估系统集成的复杂程度
- **维护复杂度**: 评估系统后期维护的复杂度

### 3. 工作量评估
- **开发工作量**: 估算开发所需的工作量（人月）
- **测试工作量**: 估算测试所需的工作量（人月）
- **部署复杂度**: 评估部署和运维的复杂度

### 4. 技术风险评估
- **技术可行性**: 评估技术方案的可行性
- **实现风险**: 识别技术实现的主要风险
- **技术创新度**: 评估项目的技术创新程度

### 5. 技术建议
- **架构优化建议**: 提供架构设计的优化建议
- **技术选型建议**: 推荐合适的技术栈
- **风险缓解措施**: 提供技术风险的缓解方案

## 输出格式要求

请严格按照以下JSON格式返回分析结果：

{
  "architectureType": "推荐的架构类型",
  "technologyStack": {
    "backend": "后端技术栈建议",
    "frontend": "前端技术栈建议", 
    "database": "数据库技术建议",
    "middleware": "中间件技术建议"
  },
  "integrationComplexity": "集成复杂度等级(low/medium/high)",
  "developmentEffort": 开发工作量估算(人月),
  "testingEffort": 测试工作量估算(人月),
  "deploymentComplexity": 部署复杂度评分(0-1),
  "technicalFeasibility": 技术可行性评分(0-1),
  "implementationRisk": 实现风险评分(0-1),
  "technicalInnovation": 技术创新度评分(0-1),
  "maintenanceComplexity": 维护复杂度评分(0-1),
  "technicalRecommendations": "详细的技术建议和优化方案",
  "confidenceScore": 分析置信度(0-1)
}

## 专业要求
1. **基于实际需求特征进行技术分析**
2. **考虑项目规模和业务特点**
3. **提供具体可执行的技术建议**
4. **评估要客观、准确、有依据**
5. **考虑技术的成熟度和团队技能匹配度**`,
		context.Project.Name,
		context.Project.Domain,
		context.ProjectMetrics.ProjectSize,
		context.ProjectMetrics.IntegrationComplexity,
		context.RequirementStats.TotalCount,
		context.RequirementStats.FunctionTypeDistribution["EIF"],
		context.GetContextSummary(),
		engine.analyzeDataRequirements(context),
		engine.analyzeIntegrationRequirements(context),
		engine.analyzeNonFunctionalRequirements(context))

	return prompt
}

// buildBusinessAnalysisPrompt 构建业务分析提示词
func (engine *AIAnalysisEngine) buildBusinessAnalysisPrompt(context *ProjectAnalysisContext) string {
	prompt := fmt.Sprintf(`# 业务价值分析专家任务

## 专家背景
您是一位拥有10年以上经验的业务分析师和投资评估专家，精通业务价值评估、投资回报分析、市场竞争分析、用户体验评估等领域。您具备多个行业的数字化转型和业务流程优化经验，对业务战略、市场趋势、用户需求有深入洞察。

## 分析任务
请对以下项目进行全面的业务价值分析，从业务价值、投资回报、战略意义、用户体验、市场竞争等多个维度提供专业的业务评估。

## 项目业务信息
- **项目名称**: %s
- **业务领域**: %s
- **项目规模**: %s
- **目标用户**: 根据需求分析推断
- **业务流程**: 基于功能需求分析

## 项目业务上下文
%s

## 业务需求分析
%s

## 分析维度

### 1. 业务价值评估
- **直接价值**: 项目带来的直接业务价值
- **间接价值**: 项目产生的间接效益
- **长期价值**: 项目的长期战略价值

### 2. 投资回报分析
- **成本效益**: 分析投入产出比
- **回报周期**: 估算投资回报周期
- **风险收益**: 评估风险与收益的平衡

### 3. 用户体验分析
- **用户满意度**: 预期的用户满意度
- **易用性**: 系统的易用性评估
- **用户采纳度**: 用户接受和使用的可能性

### 4. 市场竞争分析
- **市场定位**: 在市场中的定位分析
- **竞争优势**: 相对于竞争对手的优势
- **创新程度**: 业务模式或功能的创新性

## 输出格式要求

{
  "businessValue": 业务价值评分(0-1),
  "roiEstimation": ROI估算值,
  "strategicAlignment": 战略一致性评分(0-1),
  "userExperienceScore": 用户体验评分(0-1),
  "processEfficiency": 流程效率评分(0-1),
  "processAutomation": 流程自动化程度(0-1),
  "businessLogicComplexity": 业务逻辑复杂度(0-1),
  "marketFit": 市场适配度(0-1),
  "competitiveAdvantage": 竞争优势评分(0-1),
  "innovationLevel": 创新水平评分(0-1),
  "regulatoryCompliance": 法规合规性评分(0-1),
  "businessRecommendations": "详细的业务优化建议",
  "confidenceScore": 分析置信度(0-1)
}`,
		context.Project.Name,
		context.Project.Domain,
		context.ProjectMetrics.ProjectSize,
		context.GetContextSummary(),
		engine.analyzeBusinessRequirements(context))

	return prompt
}

// buildQualityAnalysisPrompt 构建质量分析提示词
func (engine *AIAnalysisEngine) buildQualityAnalysisPrompt(context *ProjectAnalysisContext) string {
	prompt := fmt.Sprintf(`# 软件质量分析专家任务

## 专家背景
您是一位拥有15年以上经验的软件质量保证专家，精通ISO/IEC 25010软件质量模型、软件质量度量、质量管理体系等领域。您具备大型软件项目的质量规划、质量控制、质量改进的丰富经验。

## 分析任务
请对以下项目进行全面的质量分析，从需求质量、设计质量、可维护性、可测试性等多个维度评估项目的软件质量状况。

## 项目质量信息
- **项目名称**: %s
- **项目规模**: %s
- **需求总数**: %d
- **需求平均描述长度**: %.1f字符
- **需求层级分布**: %s

## 质量分析数据
%s

## 分析维度

### 1. 需求质量
- **正确性**: 需求的正确性和准确性
- **完整性**: 需求的完整程度
- **一致性**: 需求间的一致性
- **清晰性**: 需求描述的清晰度
- **可追溯性**: 需求的可追溯性

### 2. 设计质量
- **模块化**: 系统设计的模块化程度
- **内聚性**: 模块内的内聚程度
- **耦合性**: 模块间的耦合程度
- **可扩展性**: 设计的可扩展性

### 3. 可维护性
- **可理解性**: 代码和设计的可理解性
- **可修改性**: 系统的可修改程度
- **稳定性**: 修改对系统的影响程度

## 输出格式要求

{
  "overallQuality": 总体质量评分(0-1),
  "requirementQuality": 需求质量评分(0-1),
  "designQuality": 设计质量评分(0-1),
  "correctness": 正确性评分(0-1),
  "completeness": 完整性评分(0-1),
  "consistency": 一致性评分(0-1),
  "clarity": 清晰性评分(0-1),
  "traceability": 可追溯性评分(0-1),
  "maintainability": 可维护性评分(0-1),
  "modularity": 模块化程度(0-1),
  "reusability": 可重用性评分(0-1),
  "testability": 可测试性评分(0-1),
  "qualityRecommendations": "详细的质量改进建议",
  "confidenceScore": 分析置信度(0-1)
}`,
		context.Project.Name,
		context.ProjectMetrics.ProjectSize,
		context.RequirementStats.TotalCount,
		context.RequirementStats.AverageDescriptionLength,
		engine.formatDistribution(context.RequirementStats.LevelDistribution),
		engine.analyzeQualityMetrics(context))

	return prompt
}

// buildRiskAnalysisPrompt 构建风险分析提示词
func (engine *AIAnalysisEngine) buildRiskAnalysisPrompt(context *ProjectAnalysisContext) string {
	prompt := fmt.Sprintf(`# 项目风险分析专家任务

## 专家背景
您是一位拥有12年以上经验的项目风险管理专家，精通项目管理、风险识别、风险评估、风险缓解等领域。您具备多个行业大型软件项目的风险管理经验，对技术风险、业务风险、管理风险有深入的理解和实践经验。

## 分析任务
请对以下项目进行全面的风险分析，识别技术风险、业务风险、管理风险等，并提供风险缓解建议。

## 项目风险信息
- **项目名称**: %s
- **项目规模**: %s
- **技术复杂度**: %s
- **集成复杂度**: %s
- **业务领域**: %s

## 风险分析上下文
%s

## 分析维度

### 1. 技术风险
- **技术可行性风险**: 技术方案的可行性
- **技术复杂度风险**: 技术实现的复杂度
- **集成风险**: 系统集成的风险

### 2. 项目管理风险
- **进度风险**: 项目进度延期的风险
- **资源风险**: 人力资源不足的风险
- **成本风险**: 项目成本超支的风险

### 3. 业务风险
- **需求变更风险**: 业务需求变更的风险
- **市场风险**: 市场环境变化的风险
- **合规风险**: 法规合规的风险

## 输出格式要求

{
  "overallRisk": 总体风险评分(0-1),
  "riskLevel": "风险等级(low/medium/high)",
  "technicalRisk": 技术风险评分(0-1),
  "implementationRisk": 实现风险评分(0-1),
  "integrationRisk": 集成风险评分(0-1),
  "scheduleRisk": 进度风险评分(0-1),
  "budgetRisk": 预算风险评分(0-1),
  "resourceRisk": 资源风险评分(0-1),
  "businessRisk": 业务风险评分(0-1),
  "marketRisk": 市场风险评分(0-1),
  "complianceRisk": 合规风险评分(0-1),
  "riskRecommendations": "详细的风险缓解建议",
  "confidenceScore": 分析置信度(0-1)
}`,
		context.Project.Name,
		context.ProjectMetrics.ProjectSize,
		context.ProjectMetrics.TechnologyComplexity,
		context.ProjectMetrics.IntegrationComplexity,
		context.Project.Domain,
		context.GetContextSummary())

	return prompt
}

// buildRecommendationAnalysisPrompt 构建建议分析提示词
func (engine *AIAnalysisEngine) buildRecommendationAnalysisPrompt(context *ProjectAnalysisContext) string {
	prompt := fmt.Sprintf(`# 项目改进建议专家任务

## 专家背景
您是一位拥有20年以上经验的软件项目顾问和改进专家，精通软件开发最佳实践、项目管理、质量改进、团队协作等领域。您具备多个行业的数字化转型和流程优化经验，对项目成功要素有深入理解。

## 分析任务
基于项目的全面分析，提供具体可行的改进建议，包括功能优化、技术改进、质量提升、风险缓解等方面的专业建议。

## 项目改进上下文
%s

## 建议分析要求

### 1. 功能优化建议
- 基于功能分析结果提供功能优化建议
- 识别功能缺口和改进机会

### 2. 技术改进建议  
- 基于技术分析结果提供技术架构优化建议
- 推荐技术选型和实现方案改进

### 3. 质量提升建议
- 基于质量分析结果提供质量改进措施
- 建议质量保证和质量控制方法

### 4. 风险缓解建议
- 基于风险分析结果提供风险应对策略
- 建议风险监控和预防措施

## 输出格式要求

请返回建议列表的JSON数组，每个建议包含以下字段：

[
  {
    "recommendationType": "建议类型(functional/technical/quality/risk/process)",
    "priority": "优先级(high/medium/low)",
    "impactLevel": "影响程度(high/medium/low)",
    "title": "建议标题",
    "description": "详细描述",
    "rationale": "建议理由和依据",
    "estimatedEffort": 预估工作量(人天),
    "expectedBenefit": "预期收益描述",
    "confidenceScore": 建议可行性置信度(0-1),
    "urgency": "紧急程度(urgent/normal/low)"
  }
]`,
		context.GetContextSummary())

	return prompt
}

// buildComplianceAnalysisPrompt 构建合规性分析提示词
func (engine *AIAnalysisEngine) buildComplianceAnalysisPrompt(context *ProjectAnalysisContext) string {
	prompt := fmt.Sprintf(`# NESMA合规性分析专家任务

## 专家背景
您是一位NESMA 2.2国际标准认证专家，具备ISO/IEC 14143软件度量标准的深厚理解，拥有15年以上的功能点分析和NESMA合规性评估经验。

## 分析任务
请对以下项目进行NESMA 2.2标准的合规性分析，评估项目在功能点识别、计算方法、文档规范等方面的合规程度。

## 合规分析信息
- **项目名称**: %s
- **评估标准**: NESMA 2.2
- **功能点总数**: %d
- **需求层级**: %s

## 合规分析上下文
%s

## 合规检查要点

### 1. 功能点识别合规性
- 功能点类型识别的准确性
- 边界确定的正确性
- 计数规则的遵循程度

### 2. 计算方法合规性
- 复杂度评估的标准性
- 权重分配的正确性
- 调整因子的合理性

### 3. 文档规范合规性
- 需求描述的规范性
- 分析过程的完整性
- 结果记录的标准性

## 输出格式要求

{
  "complianceStandard": "NESMA 2.2",
  "standardVersion": "2.2",
  "overallCompliance": 总体合规性评分(0-1),
  "complianceLevel": "合规等级(high/medium/low)",
  "complianceRecommendations": "详细的合规改进建议",
  "confidenceScore": 分析置信度(0-1)
}`,
		context.Project.Name,
		context.FunctionPointStats.TotalFunctionPoints,
		engine.formatDistribution(context.RequirementStats.LevelDistribution),
		context.GetContextSummary())

	return prompt
}

// callAIForAnalysis 调用AI进行分析
func (engine *AIAnalysisEngine) callAIForAnalysis(prompt string, config AnalysisConfiguration, metrics *ExecutionMetrics) (string, error) {
	defer func() {
		metrics.AICallCount++
	}()

	// 使用新的增强AI服务管理器进行智能模型选择
	aiServiceManager := GetAIServiceManager()
	
	// 创建AI分析请求，评估任务复杂度为高复杂度NESMA分析
	analysisRequest := &EnhancedAIAnalysisRequest{
		Prompt:      prompt,
		Complexity:  ComplexityHigh,     // NESMA分析是高复杂度任务
		ModelType:   ModelTypeComplexAnalysis, // 使用复杂分析模型类型
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		Timeout:     5 * time.Minute,    // 给deepseek-reasoner足够的时间进行推理
	}
	
	// 使用增强AI管理器进行智能分析
	response, err := aiServiceManager.AnalyzeWithOptimalModel(context.Background(), analysisRequest)
	if err != nil {
		metrics.ErrorCount++
		// 如果新的服务失败，降级到原有服务
		global.GVA_LOG.Warn("增强AI分析失败，降级到传统服务", zap.Error(err))
		return engine.fallbackToLegacyAI(prompt, config, metrics)
	}

	// 检查响应
	if response == nil || response.Content == "" {
		metrics.ErrorCount++
		return "", fmt.Errorf("AI返回空响应")
	}

	// 记录性能指标
	global.GVA_LOG.Info("AI分析完成",
		zap.String("model", response.Model),
		zap.Float64("confidence", response.Confidence),
		zap.Duration("response_time", response.ProcessingTime),
		zap.Int("tokens_used", response.TokensUsed))

	return response.Content, nil
}

// fallbackToLegacyAI 降级到传统AI服务
func (engine *AIAnalysisEngine) fallbackToLegacyAI(prompt string, config AnalysisConfiguration, metrics *ExecutionMetrics) (string, error) {
	// 构建AI配置
	aiConfig := &AIConfig{
		MaxTokens:   config.MaxTokens,
		Temperature: config.Temperature,
		Model:       config.AIModel,
	}

	// 调用传统AI服务
	response, err := (*engine.aiService).GenerateText(context.Background(), prompt, aiConfig)
	if err != nil {
		return "", err
	}

	// 检查响应
	if response == nil || len(response.Choices) == 0 {
		return "", fmt.Errorf("AI返回空响应")
	}

	content := response.Choices[0].Message.Content
	global.GVA_LOG.Info("使用传统AI服务完成分析", zap.String("model", config.AIModel))
	
	return content, nil
}

// 工具方法
func (engine *AIAnalysisEngine) buildRequirementSummary(requirements []nesma.NesmaRequirement) string {
	var summary strings.Builder
	
	// 只取前20个需求作为示例，避免上下文过长
	displayCount := len(requirements)
	if displayCount > 20 {
		displayCount = 20
	}

	for i := 0; i < displayCount; i++ {
		req := requirements[i]
		summary.WriteString(fmt.Sprintf("- [L%d] %s: %s (功能类型: %s)\n", 
			req.Level, req.Title, 
			engine.truncateString(req.Description, 100),
			req.FunctionType))
	}

	if len(requirements) > 20 {
		summary.WriteString(fmt.Sprintf("... 还有%d个需求\n", len(requirements)-20))
	}

	return summary.String()
}

func (engine *AIAnalysisEngine) formatDistribution(dist map[int]int) string {
	var result strings.Builder
	for key, value := range dist {
		result.WriteString(fmt.Sprintf("- %d: %d个\n", key, value))
	}
	return result.String()
}

func (engine *AIAnalysisEngine) formatStringDistribution(dist map[string]int) string {
	var result strings.Builder
	for key, value := range dist {
		result.WriteString(fmt.Sprintf("- %s: %d个\n", key, value))
	}
	return result.String()
}

func (engine *AIAnalysisEngine) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (engine *AIAnalysisEngine) analyzeDataRequirements(context *ProjectAnalysisContext) string {
	ilfCount := context.RequirementStats.FunctionTypeDistribution["ILF"]
	eifCount := context.RequirementStats.FunctionTypeDistribution["EIF"]
	
	return fmt.Sprintf("数据处理需求: ILF=%d, EIF=%d, 数据复杂度评估中等", ilfCount, eifCount)
}

func (engine *AIAnalysisEngine) analyzeIntegrationRequirements(context *ProjectAnalysisContext) string {
	eifCount := context.RequirementStats.FunctionTypeDistribution["EIF"]
	
	integrationLevel := "低"
	if eifCount > 10 {
		integrationLevel = "高"
	} else if eifCount > 5 {
		integrationLevel = "中"
	}
	
	return fmt.Sprintf("集成需求: 外部接口%d个, 集成复杂度%s", eifCount, integrationLevel)
}

func (engine *AIAnalysisEngine) analyzeNonFunctionalRequirements(context *ProjectAnalysisContext) string {
	return "性能需求: 基于项目规模推断为中等性能要求, 需要标准的响应时间和并发支持"
}

func (engine *AIAnalysisEngine) analyzeBusinessRequirements(context *ProjectAnalysisContext) string {
	return fmt.Sprintf("业务需求分析: %s领域项目, 规模%s, 预期具有良好的业务价值和用户体验", 
		context.Project.Domain, context.ProjectMetrics.ProjectSize)
}

func (engine *AIAnalysisEngine) analyzeQualityMetrics(context *ProjectAnalysisContext) string {
	return fmt.Sprintf("质量度量: 需求%d个, 平均描述长度%.1f字符, 需求结构化程度待评估", 
		context.RequirementStats.TotalCount, context.RequirementStats.AverageDescriptionLength)
}

// AI结果解析函数

// parseFunctionalAnalysisResult 解析功能分析结果
func (engine *AIAnalysisEngine) parseFunctionalAnalysisResult(aiResponse string) (*AIFunctionalAnalysisResult, error) {
	// 提取JSON内容
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var result AIFunctionalAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析功能分析结果失败: %w", err)
	}

	// 验证必要字段
	if result.TotalFunctionPoints < 0 {
		result.TotalFunctionPoints = 0
	}
	if result.ConfidenceScore == 0 {
		result.ConfidenceScore = 0.5 // 默认置信度
	}

	return &result, nil
}

// parseTechnicalAnalysisResult 解析技术分析结果
func (engine *AIAnalysisEngine) parseTechnicalAnalysisResult(aiResponse string) (*AITechnicalAnalysisResult, error) {
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var result AITechnicalAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析技术分析结果失败: %w", err)
	}

	// 设置默认值
	if result.ArchitectureType == "" {
		result.ArchitectureType = "未指定"
	}
	if result.ConfidenceScore == 0 {
		result.ConfidenceScore = 0.5
	}

	return &result, nil
}

// parseBusinessAnalysisResult 解析业务分析结果
func (engine *AIAnalysisEngine) parseBusinessAnalysisResult(aiResponse string) (*AIBusinessAnalysisResult, error) {
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var result AIBusinessAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析业务分析结果失败: %w", err)
	}

	// 验证评分范围
	result.BusinessValue = engine.validateScore(result.BusinessValue)
	result.ROIEstimation = engine.validateScore(result.ROIEstimation)
	result.StrategicAlignment = engine.validateScore(result.StrategicAlignment)
	
	if result.ConfidenceScore == 0 {
		result.ConfidenceScore = 0.5
	}

	return &result, nil
}

// parseQualityAnalysisResult 解析质量分析结果
func (engine *AIAnalysisEngine) parseQualityAnalysisResult(aiResponse string) (*AIQualityAnalysisResult, error) {
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var result AIQualityAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析质量分析结果失败: %w", err)
	}

	// 验证所有质量评分
	result.OverallQuality = engine.validateScore(result.OverallQuality)
	result.RequirementQuality = engine.validateScore(result.RequirementQuality)
	result.DesignQuality = engine.validateScore(result.DesignQuality)
	result.Correctness = engine.validateScore(result.Correctness)
	result.Completeness = engine.validateScore(result.Completeness)
	result.Consistency = engine.validateScore(result.Consistency)
	result.Clarity = engine.validateScore(result.Clarity)
	result.Traceability = engine.validateScore(result.Traceability)
	result.Maintainability = engine.validateScore(result.Maintainability)
	result.Modularity = engine.validateScore(result.Modularity)
	result.Reusability = engine.validateScore(result.Reusability)
	result.Testability = engine.validateScore(result.Testability)

	if result.ConfidenceScore == 0 {
		result.ConfidenceScore = 0.5
	}

	return &result, nil
}

// parseRiskAnalysisResult 解析风险分析结果
func (engine *AIAnalysisEngine) parseRiskAnalysisResult(aiResponse string) (*AIRiskAnalysisResult, error) {
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var result AIRiskAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析风险分析结果失败: %w", err)
	}

	// 验证所有风险评分
	result.OverallRisk = engine.validateScore(result.OverallRisk)
	result.TechnicalRisk = engine.validateScore(result.TechnicalRisk)
	result.ImplementationRisk = engine.validateScore(result.ImplementationRisk)
	result.IntegrationRisk = engine.validateScore(result.IntegrationRisk)
	result.ScheduleRisk = engine.validateScore(result.ScheduleRisk)
	result.BudgetRisk = engine.validateScore(result.BudgetRisk)
	result.ResourceRisk = engine.validateScore(result.ResourceRisk)
	result.BusinessRisk = engine.validateScore(result.BusinessRisk)
	result.MarketRisk = engine.validateScore(result.MarketRisk)
	result.ComplianceRisk = engine.validateScore(result.ComplianceRisk)

	// 设置风险等级
	if result.RiskLevel == "" {
		if result.OverallRisk >= 0.7 {
			result.RiskLevel = "high"
		} else if result.OverallRisk >= 0.4 {
			result.RiskLevel = "medium"
		} else {
			result.RiskLevel = "low"
		}
	}

	if result.ConfidenceScore == 0 {
		result.ConfidenceScore = 0.5
	}

	return &result, nil
}

// parseRecommendationAnalysisResult 解析建议分析结果
func (engine *AIAnalysisEngine) parseRecommendationAnalysisResult(aiResponse string) ([]AIRecommendationResult, error) {
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var results []AIRecommendationResult
	if err := json.Unmarshal([]byte(jsonContent), &results); err != nil {
		return nil, fmt.Errorf("解析建议分析结果失败: %w", err)
	}

	// 验证和清理数据
	for i := range results {
		if results[i].Title == "" {
			results[i].Title = fmt.Sprintf("建议 #%d", i+1)
		}
		if results[i].Priority == "" {
			results[i].Priority = "medium"
		}
		if results[i].ImpactLevel == "" {
			results[i].ImpactLevel = "medium"
		}
		if results[i].RecommendationType == "" {
			results[i].RecommendationType = "general"
		}
		if results[i].Urgency == "" {
			results[i].Urgency = "normal"
		}
		if results[i].ConfidenceScore == 0 {
			results[i].ConfidenceScore = 0.5
		}
	}

	return results, nil
}

// parseComplianceAnalysisResult 解析合规性分析结果
func (engine *AIAnalysisEngine) parseComplianceAnalysisResult(aiResponse string) (*AIComplianceAnalysisResult, error) {
	jsonContent := engine.extractJSONFromResponse(aiResponse)
	if jsonContent == "" {
		return nil, fmt.Errorf("未找到有效的JSON响应内容")
	}

	var result AIComplianceAnalysisResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		return nil, fmt.Errorf("解析合规性分析结果失败: %w", err)
	}

	// 设置默认值
	if result.ComplianceStandard == "" {
		result.ComplianceStandard = "NESMA 2.2"
	}
	if result.StandardVersion == "" {
		result.StandardVersion = "2.2"
	}

	// 验证评分
	result.OverallCompliance = engine.validateScore(result.OverallCompliance)

	// 设置合规等级
	if result.ComplianceLevel == "" {
		if result.OverallCompliance >= 0.8 {
			result.ComplianceLevel = "high"
		} else if result.OverallCompliance >= 0.6 {
			result.ComplianceLevel = "medium"
		} else {
			result.ComplianceLevel = "low"
		}
	}

	if result.ConfidenceScore == 0 {
		result.ConfidenceScore = 0.5
	}

	return &result, nil
}

// 工具函数

// extractJSONFromResponse 从AI响应中提取JSON内容
func (engine *AIAnalysisEngine) extractJSONFromResponse(response string) string {
	// 查找JSON代码块
	if start := strings.Index(response, "```json"); start != -1 {
		start += 7 // 跳过 "```json"
		if end := strings.Index(response[start:], "```"); end != -1 {
			return strings.TrimSpace(response[start : start+end])
		}
	}

	// 查找裸JSON
	if start := strings.Index(response, "{"); start != -1 {
		// 找到最后一个大括号
		if end := strings.LastIndex(response, "}"); end != -1 && end > start {
			return strings.TrimSpace(response[start : end+1])
		}
	}

	// 查找JSON数组
	if start := strings.Index(response, "["); start != -1 {
		if end := strings.LastIndex(response, "]"); end != -1 && end > start {
			return strings.TrimSpace(response[start : end+1])
		}
	}

	return ""
}

// validateScore 验证评分范围 (0-1)
func (engine *AIAnalysisEngine) validateScore(score float64) float64 {
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}