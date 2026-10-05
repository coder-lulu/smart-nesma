package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// DescriptionGenerationResult 功能点描述生成结果
type DescriptionGenerationResult struct {
	RequirementID        uint                     `json:"requirement_id"`
	OriginalRequirement  *nesma.NesmaRequirement  `json:"original_requirement"`
	GeneratedDescription request.EnhancedDescription      `json:"generated_description"`
	AnalysisConfidence   float64                  `json:"analysis_confidence"`
	ProcessingTime       int64                    `json:"processing_time_ms"`
	GenerationNotes      string                   `json:"generation_notes"`
	KnowledgeReferences  []KnowledgeReference     `json:"knowledge_refs"`
	ImprovementSuggestions []string               `json:"improvement_suggestions"`
}

// DescriptionGeneratorService 功能点描述生成服务
type DescriptionGeneratorService struct {
	db               *gorm.DB
	aiService        AIService
	knowledgeService *KnowledgeService
}

// NewDescriptionGeneratorService 创建功能点描述生成服务
func NewDescriptionGeneratorService() *DescriptionGeneratorService {
	return &DescriptionGeneratorService{
		db:               global.GVA_DB,
		aiService:        GetAIService(),
		knowledgeService: NewKnowledgeService(),
	}
}

// GenerateRequirementDescriptions 生成需求描述
func (s *DescriptionGeneratorService) GenerateRequirementDescriptions(cycleID uint, requirementIDs []uint, levels []int) ([]*DescriptionGenerationResult, error) {
	global.GVA_LOG.Info("开始生成需求描述", zap.Uint("cycleID", cycleID), zap.Any("requirementIDs", requirementIDs), zap.Any("levels", levels))

	var results []*DescriptionGenerationResult
	
	// 获取周期信息
	var cycle nesma.NesmaProjectCycle
	if err := s.db.Preload("Project").First(&cycle, cycleID).Error; err != nil {
		return nil, fmt.Errorf("获取项目周期失败: %w", err)
	}

	// 获取需要生成描述的功能点
	var requirements []nesma.NesmaRequirement
	query := s.db.Where("cycle_id = ?", cycleID)
	
	if len(requirementIDs) > 0 {
		query = query.Where("id IN ?", requirementIDs)
	}
	
	if len(levels) > 0 {
		query = query.Where("level IN ?", levels)
	}
	
	if err := query.Find(&requirements).Error; err != nil {
		return nil, fmt.Errorf("获取功能点失败: %w", err)
	}

	// 逐个生成描述
	for _, requirement := range requirements {
		result, err := s.generateRequirementDescription(&requirement, &cycle)
		if err != nil {
			global.GVA_LOG.Error("生成需求描述失败", zap.Uint("requirementID", requirement.ID), zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// generateRequirementDescription 生成单个需求描述
func (s *DescriptionGeneratorService) generateRequirementDescription(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) (*DescriptionGenerationResult, error) {
	startTime := time.Now()
	
	// 1. 搜索相关知识库内容
	knowledgeRefs, err := s.searchRelevantKnowledge(requirement, cycle)
	if err != nil {
		global.GVA_LOG.Warn("搜索相关知识失败", zap.Error(err))
		knowledgeRefs = []KnowledgeReference{}
	}

	// 2. 生成增强描述
	enhancedDesc, err := s.generateEnhancedDescription(requirement, cycle, knowledgeRefs)
	if err != nil {
		return nil, fmt.Errorf("生成增强描述失败: %w", err)
	}

	// 3. 生成改进建议
	improvementSuggestions := s.generateImprovementSuggestions(requirement, enhancedDesc)

	// 4. 计算置信度
	confidence := s.calculateDescriptionConfidence(enhancedDesc, knowledgeRefs)

	// 5. 构建生成结果
	result := &DescriptionGenerationResult{
		RequirementID:          requirement.ID,
		OriginalRequirement:    requirement,
		GeneratedDescription:   *enhancedDesc,
		AnalysisConfidence:     confidence,
		ProcessingTime:         time.Since(startTime).Milliseconds(),
		GenerationNotes:        s.generateDescriptionNotes(requirement, enhancedDesc),
		KnowledgeReferences:    knowledgeRefs,
		ImprovementSuggestions: improvementSuggestions,
	}

	global.GVA_LOG.Info("需求描述生成完成", 
		zap.Uint("requirementID", requirement.ID),
		zap.Float64("置信度", confidence),
		zap.Int("数据元素数", len(enhancedDesc.DataElements)),
		zap.Int("业务规则数", len(enhancedDesc.BusinessRules)),
	)

	return result, nil
}

// searchRelevantKnowledge 搜索相关知识库内容
func (s *DescriptionGeneratorService) searchRelevantKnowledge(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) ([]KnowledgeReference, error) {
	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s %s %s 功能描述 业务规则", requirement.Title, requirement.Description, cycle.Project.Domain)
	
	// 调用知识库搜索服务
	searchReq := &request.SearchKnowledgeEntriesRequest{
		Query: searchQuery,
		Limit: 8,
	}
	
	searchResp, err := s.knowledgeService.SearchKnowledgeEntries(context.Background(), searchReq)
	if err != nil {
		return nil, fmt.Errorf("搜索知识库失败: %w", err)
	}

	// 转换为知识库引用格式
	var references []KnowledgeReference
	for _, result := range searchResp.Results {
		references = append(references, KnowledgeReference{
			ID:        result.Knowledge.ID,
			Title:     result.Knowledge.Title,
			Category:  result.Knowledge.Category,
			Relevance: float64(result.Similarity),
		})
	}

	return references, nil
}

// generateEnhancedDescription 生成增强描述
func (s *DescriptionGeneratorService) generateEnhancedDescription(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, knowledgeRefs []KnowledgeReference) (*request.EnhancedDescription, error) {
	// 构建AI分析prompt
	prompt := s.buildDescriptionPrompt(requirement, cycle, knowledgeRefs)
	
	// 调用AI服务
	messages := []APIMessage{
		{Role: "system", Content: "你是一个专业的业务分析师和需求工程师，擅长编写详细的功能需求文档。"},
		{Role: "user", Content: prompt},
	}

	config := &AIConfig{
		MaxTokens:   4000,
		Temperature: 0.7,
		TopP:        0.9,
		Model:       "deepseek-chat",
	}

	response, err := s.aiService.ChatCompletion(context.Background(), messages, config)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	return s.parseDescriptionResponse(response.Text)
}

// buildDescriptionPrompt 构建描述生成的prompt
func (s *DescriptionGeneratorService) buildDescriptionPrompt(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, knowledgeRefs []KnowledgeReference) string {
	var knowledgeContext strings.Builder
	for _, ref := range knowledgeRefs {
		knowledgeContext.WriteString(fmt.Sprintf("- %s: %s\n", ref.Title, ref.Category))
	}

	return fmt.Sprintf(`
请为以下功能点生成详细的功能描述文档：

【功能点信息】
标题: %s
当前描述: %s
功能级别: %s
功能类型: %s
复杂度: %s
业务价值: %s
项目领域: %s

【相关知识库内容】
%s

【生成要求】
请生成一份专业的功能需求文档，包含以下内容：
1. 功能概述和业务上下文
2. 详细的功能流程（输入-处理-输出）
3. 数据元素和格式要求
4. 业务规则和约束条件
5. 质量属性要求
6. 验收标准和测试场景
7. 依赖关系和假设条件
8. 风险因素识别

【输出格式】
请以JSON格式返回结果：
{
  "title": "优化后的功能标题",
  "summary": "功能概述",
  "detailed_description": "详细功能描述",
  "business_context": "业务上下文和价值",
  "functional_flow": {
    "input_sources": ["输入来源1", "输入来源2"],
    "processing_steps": ["处理步骤1", "处理步骤2"],
    "output_targets": ["输出目标1", "输出目标2"],
    "exception_handling": ["异常处理1", "异常处理2"]
  },
  "data_elements": [
    {
      "name": "数据元素名称",
      "type": "数据类型",
      "description": "元素描述",
      "mandatory": true,
      "format": "格式要求",
      "validation": "验证规则"
    }
  ],
  "business_rules": [
    {
      "rule_id": "BR001",
      "description": "业务规则描述",
      "condition": "触发条件",
      "action": "执行动作",
      "priority": "高/中/低"
    }
  ],
  "quality_attributes": [
    {
      "attribute": "质量属性名称",
      "requirement": "质量要求",
      "measurement": "度量方法"
    }
  ],
  "acceptance_criteria": [
    {
      "criterion_id": "AC001",
      "description": "验收标准描述",
      "priority": "高/中/低",
      "testable": true
    }
  ],
  "test_scenarios": [
    {
      "scenario_id": "TS001",
      "description": "测试场景描述",
      "preconditions": ["前提条件1"],
      "steps": ["测试步骤1", "测试步骤2"],
      "expected": "期望结果"
    }
  ],
  "dependencies": ["依赖项1", "依赖项2"],
  "assumptions": ["假设条件1", "假设条件2"],
  "constraints": ["约束条件1", "约束条件2"],
  "risk_factors": ["风险因素1", "风险因素2"]
}

【质量标准】
- 描述要专业、准确、可理解
- 业务规则要具体、可操作
- 验收标准要可测试、可验证
- 数据元素要完整、规范
- 测试场景要覆盖主要业务流程`,
		requirement.Title,
		requirement.Description,
		requirement.GetLevelName(),
		requirement.FunctionType,
		requirement.Complexity,
		requirement.BusinessValue,
		cycle.Project.Domain,
		knowledgeContext.String(),
	)
}

// parseDescriptionResponse 解析描述响应
func (s *DescriptionGeneratorService) parseDescriptionResponse(response string) (*request.EnhancedDescription, error) {
	var enhancedDesc request.EnhancedDescription

	// 尝试直接解析JSON
	if err := json.Unmarshal([]byte(response), &enhancedDesc); err != nil {
		// 如果解析失败，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonContent := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonContent), &enhancedDesc); err != nil {
				global.GVA_LOG.Warn("解析描述响应失败", zap.String("response", response), zap.Error(err))
				return s.createFallbackDescription(response), nil
			}
		} else {
			global.GVA_LOG.Warn("无法从响应中提取JSON", zap.String("response", response))
			return s.createFallbackDescription(response), nil
		}
	}

	return &enhancedDesc, nil
}

// createFallbackDescription 创建备用描述
func (s *DescriptionGeneratorService) createFallbackDescription(response string) *request.EnhancedDescription {
	return &request.EnhancedDescription{
		Title:               "功能点",
		Summary:             "基础功能描述",
		DetailedDescription: fmt.Sprintf("AI响应内容：\n%s", response),
		BusinessContext:     "业务功能支持",
		FunctionalFlow: request.FunctionalFlow{
			InputSources:      []string{"用户输入"},
			ProcessingSteps:   []string{"数据处理"},
			OutputTargets:     []string{"结果输出"},
			ExceptionHandling: []string{"异常处理"},
		},
		DataElements: []request.DataElement{
			{
				Name:        "基础数据",
				Type:        "string",
				Description: "基础数据元素",
				Mandatory:   true,
				Format:      "文本格式",
				Validation:  "非空验证",
			},
		},
		BusinessRules: []request.BusinessRule{
			{
				RuleID:      "BR001",
				Description: "基础业务规则",
				Condition:   "满足条件",
				Action:      "执行动作",
				Priority:    "中",
			},
		},
		QualityAttributes: []request.QualityAttribute{
			{
				Attribute:   "可用性",
				Requirement: "系统可用性要求",
				Measurement: "可用性度量",
			},
		},
		AcceptanceCriteria: []request.AcceptanceCriterion{
			{
				CriterionID: "AC001",
				Description: "基础验收标准",
				Priority:    "高",
				Testable:    true,
			},
		},
		TestScenarios: []request.TestScenario{
			{
				ScenarioID:    "TS001",
				Description:   "基础测试场景",
				Preconditions: []string{"系统正常"},
				Steps:         []string{"执行操作"},
				Expected:      "得到预期结果",
			},
		},
		Dependencies:  []string{"系统依赖"},
		Assumptions:   []string{"基础假设"},
		Constraints:   []string{"约束条件"},
		RiskFactors:   []string{"风险因素"},
	}
}

// generateImprovementSuggestions 生成改进建议
func (s *DescriptionGeneratorService) generateImprovementSuggestions(requirement *nesma.NesmaRequirement, enhancedDesc *request.EnhancedDescription) []string {
	var suggestions []string

	// 基于生成结果的质量评估
	if len(enhancedDesc.DataElements) < 2 {
		suggestions = append(suggestions, "建议补充更多数据元素定义")
	}

	if len(enhancedDesc.BusinessRules) < 1 {
		suggestions = append(suggestions, "建议添加具体的业务规则")
	}

	if len(enhancedDesc.AcceptanceCriteria) < 3 {
		suggestions = append(suggestions, "建议增加更多验收标准")
	}

	if len(enhancedDesc.TestScenarios) < 2 {
		suggestions = append(suggestions, "建议添加更多测试场景")
	}

	if enhancedDesc.BusinessContext == "" {
		suggestions = append(suggestions, "建议补充业务上下文信息")
	}

	if len(enhancedDesc.Dependencies) == 0 {
		suggestions = append(suggestions, "建议识别功能依赖关系")
	}

	return suggestions
}

// calculateDescriptionConfidence 计算描述生成置信度
func (s *DescriptionGeneratorService) calculateDescriptionConfidence(enhancedDesc *request.EnhancedDescription, knowledgeRefs []KnowledgeReference) float64 {
	var score float64 = 0.5 // 基础分数

	// 内容完整性评分
	if enhancedDesc.DetailedDescription != "" {
		score += 0.1
	}
	if enhancedDesc.BusinessContext != "" {
		score += 0.1
	}
	if len(enhancedDesc.DataElements) > 0 {
		score += 0.1
	}
	if len(enhancedDesc.BusinessRules) > 0 {
		score += 0.1
	}
	if len(enhancedDesc.AcceptanceCriteria) > 0 {
		score += 0.1
	}

	// 知识库支持度
	knowledgeBonus := float64(len(knowledgeRefs)) * 0.05
	if knowledgeBonus > 0.2 {
		knowledgeBonus = 0.2
	}

	return score + knowledgeBonus
}

// generateDescriptionNotes 生成描述备注
func (s *DescriptionGeneratorService) generateDescriptionNotes(requirement *nesma.NesmaRequirement, enhancedDesc *request.EnhancedDescription) string {
	var notes strings.Builder
	
	notes.WriteString(fmt.Sprintf("为「%s」生成详细功能描述；", requirement.Title))
	notes.WriteString(fmt.Sprintf("包含%d个数据元素、%d个业务规则、%d个验收标准、%d个测试场景；", 
		len(enhancedDesc.DataElements), 
		len(enhancedDesc.BusinessRules), 
		len(enhancedDesc.AcceptanceCriteria), 
		len(enhancedDesc.TestScenarios)))
	
	return notes.String()
}

// ApplyGeneratedDescription 应用生成的描述
func (s *DescriptionGeneratorService) ApplyGeneratedDescription(requirementID uint, enhancedDesc *request.EnhancedDescription) error {
	var requirement nesma.NesmaRequirement
	if err := s.db.First(&requirement, requirementID).Error; err != nil {
		return fmt.Errorf("获取功能点失败: %w", err)
	}

	// 将生成的描述应用到需求上
	descriptionJSON, _ := json.Marshal(enhancedDesc)
	
	updates := map[string]interface{}{
		"title":       enhancedDesc.Title,
		"description": enhancedDesc.DetailedDescription,
		"business_value": enhancedDesc.BusinessContext,
		"acceptance_criteria": enhancedDesc.AcceptanceCriteria,
		"ai_description": string(descriptionJSON),
		"ai_analysis_status": "enhanced",
	}

	now := time.Now()
	updates["ai_analysis_time"] = &now

	return s.db.Model(&requirement).Updates(updates).Error
}

// GetDescriptionGeneratorService 获取描述生成服务实例
func GetDescriptionGeneratorService() *DescriptionGeneratorService {
	return NewDescriptionGeneratorService()
}