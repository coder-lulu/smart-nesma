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

// Level3AnalysisResult 三级功能点分析结果
type Level3AnalysisResult struct {
	OriginalRequirement     *nesma.NesmaRequirement   `json:"original"`
	OptimizationSuggestions []OptimizationSuggestion  `json:"optimizations"`
	ExpansionSuggestions    []ExpansionSuggestion     `json:"expansions"`
	KnowledgeReferences     []KnowledgeReference      `json:"knowledge_refs"`
	AnalysisConfidence      float64                   `json:"confidence"`
	ProcessingTime          int64                     `json:"processing_time_ms"`
	AnalysisNotes           string                    `json:"analysis_notes"`
}

// OptimizationSuggestion 优化建议
type OptimizationSuggestion struct {
	Type           string  `json:"type"`            // title/description/category
	Field          string  `json:"field"`           // 要优化的字段名
	CurrentValue   string  `json:"current_value"`   // 当前值
	SuggestedValue string  `json:"suggested_value"` // 建议值
	Reason         string  `json:"reason"`          // 优化原因
	Priority       int     `json:"priority"`        // 优先级 1-5
	Confidence     float64 `json:"confidence"`      // 置信度 0-1
	Category       string  `json:"category"`        // 优化类别
}

// ExpansionSuggestion 扩充建议
type ExpansionSuggestion struct {
	SuggestedTitle       string  `json:"suggested_title"`
	SuggestedDescription string  `json:"suggested_description"`
	SuggestedCategory    string  `json:"suggested_category"`
	Justification        string  `json:"justification"`
	Priority             int     `json:"priority"`
	Confidence           float64 `json:"confidence"`
	RelatedKnowledge     string  `json:"related_knowledge"`
	EstimatedComplexity  string  `json:"estimated_complexity"`
}

// Level3AnalyzerService 三级功能点分析服务
type Level3AnalyzerService struct {
	db            *gorm.DB
	aiService     AIService
	knowledgeService *KnowledgeService
}

// NewLevel3AnalyzerService 创建三级功能点分析服务
func NewLevel3AnalyzerService() *Level3AnalyzerService {
	aiService := GetAIService()
	if aiService == nil {
		// 避免在logger未初始化时调用log，可能导致空指针
		// 在运行时会再次检查AI服务状态
		if global.GVA_CONFIG.AI.DeepSeek.APIKey != "" {
			aiService = NewDeepSeekService(
				global.GVA_CONFIG.AI.DeepSeek.APIKey,
				global.GVA_CONFIG.AI.DeepSeek.BaseURL,
				global.GVA_CONFIG.AI.DeepSeek.Model,
			)
		}
	}
	
	return &Level3AnalyzerService{
		db:               global.GVA_DB,
		aiService:        aiService,
		knowledgeService: NewKnowledgeService(),
	}
}

// AnalyzeLevel3Requirements 分析三级功能点
func (s *Level3AnalyzerService) AnalyzeLevel3Requirements(cycleID uint, requirementIDs []uint) ([]*Level3AnalysisResult, error) {
	global.GVA_LOG.Info("开始分析三级功能点", zap.Uint("cycleID", cycleID), zap.Any("requirementIDs", requirementIDs))

	var results []*Level3AnalysisResult
	
	// 获取周期信息
	var cycle nesma.NesmaProjectCycle
	if err := s.db.Preload("Project").First(&cycle, cycleID).Error; err != nil {
		return nil, fmt.Errorf("获取项目周期失败: %w", err)
	}

	// 获取需要分析的三级功能点
	var requirements []nesma.NesmaRequirement
	query := s.db.Where("cycle_id = ? AND level = 3", cycleID)
	if len(requirementIDs) > 0 {
		query = query.Where("id IN ?", requirementIDs)
	}
	
	if err := query.Find(&requirements).Error; err != nil {
		return nil, fmt.Errorf("获取三级功能点失败: %w", err)
	}

	// 逐个分析每个三级功能点
	for _, requirement := range requirements {
		result, err := s.analyzeLevel3Requirement(&requirement, &cycle)
		if err != nil {
			global.GVA_LOG.Error("分析三级功能点失败", zap.Uint("requirementID", requirement.ID), zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// analyzeLevel3Requirement 分析单个三级功能点
func (s *Level3AnalyzerService) analyzeLevel3Requirement(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) (*Level3AnalysisResult, error) {
	startTime := time.Now()
	
	// 1. 搜索相关知识库内容
	knowledgeRefs, err := s.searchRelevantKnowledge(requirement, cycle)
	if err != nil {
		global.GVA_LOG.Warn("搜索相关知识失败", zap.Error(err))
		knowledgeRefs = []KnowledgeReference{} // 继续执行，但没有知识库支持
	}

	// 2. 生成优化建议
	optimizationSuggestions, err := s.generateOptimizationSuggestions(requirement, knowledgeRefs)
	if err != nil {
		return nil, fmt.Errorf("生成优化建议失败: %w", err)
	}

	// 3. 生成扩充建议
	expansionSuggestions, err := s.generateExpansionSuggestions(requirement, cycle, knowledgeRefs)
	if err != nil {
		return nil, fmt.Errorf("生成扩充建议失败: %w", err)
	}

	// 4. 计算整体置信度
	confidence := s.calculateAnalysisConfidence(optimizationSuggestions, expansionSuggestions, knowledgeRefs)

	// 5. 构建分析结果
	result := &Level3AnalysisResult{
		OriginalRequirement:     requirement,
		OptimizationSuggestions: optimizationSuggestions,
		ExpansionSuggestions:    expansionSuggestions,
		KnowledgeReferences:     knowledgeRefs,
		AnalysisConfidence:      confidence,
		ProcessingTime:          time.Since(startTime).Milliseconds(),
		AnalysisNotes:           s.generateAnalysisNotes(optimizationSuggestions, expansionSuggestions),
	}

	global.GVA_LOG.Info("三级功能点分析完成", 
		zap.Uint("requirementID", requirement.ID),
		zap.Int("优化建议数", len(optimizationSuggestions)),
		zap.Int("扩充建议数", len(expansionSuggestions)),
		zap.Float64("置信度", confidence),
	)

	return result, nil
}

// searchRelevantKnowledge 搜索相关知识库内容
func (s *Level3AnalyzerService) searchRelevantKnowledge(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) ([]KnowledgeReference, error) {
	// 空指针检查
	if s.knowledgeService == nil {
		global.GVA_LOG.Warn("知识库服务未初始化，跳过知识库搜索")
		return []KnowledgeReference{}, nil
	}
	
	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s %s %s", requirement.Title, requirement.Description, cycle.Project.Domain)
	
	// 调用知识库搜索服务
	searchReq := &request.SearchKnowledgeEntriesRequest{
		Query: searchQuery,
		Limit: 5,
	}
	
	searchResp, err := s.knowledgeService.SearchKnowledgeEntries(context.Background(), searchReq)
	if err != nil {
		global.GVA_LOG.Warn("搜索知识库失败", zap.Error(err))
		return []KnowledgeReference{}, nil // 返回空列表而不是错误，让流程继续
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

// generateOptimizationSuggestions 生成优化建议
func (s *Level3AnalyzerService) generateOptimizationSuggestions(requirement *nesma.NesmaRequirement, knowledgeRefs []KnowledgeReference) ([]OptimizationSuggestion, error) {
	// 构建AI分析prompt
	prompt := s.buildOptimizationPrompt(requirement, knowledgeRefs)
	
	// 调用AI服务
	messages := []APIMessage{
		{Role: "system", Content: `你是一位国际认证的NESMA（荷兰软件度量协会）功能点分析专家，拥有20年以上的软件度量和需求分析经验。你的专业资质包括：

## 核心专业能力
- NESMA 2.2国际标准专家：精通NESMA功能点分析的所有细节、规则和最佳实践
- ISO/IEC 14143标准认证：具备国际软件度量标准认证资质
- 多行业项目经验：在金融、制造、医疗、政府、电商等15+个行业有丰富的项目实践
- 软件工程专家：深度理解软件开发生命周期、架构设计、质量保证体系
- 需求工程大师：擅长需求获取、分析、建模、验证和管理的全流程

## 专业技能矩阵
- 精确功能点识别：准确识别EI/EO/EQ/ILF/EIF五种功能类型，准确率>95%
- 科学复杂度评估：基于DET/RET/FTR的科学评估方法，考虑技术、业务、数据三个维度
- 质量评估：从完整性、一致性、可追溯性、可测试性等多维度评估需求质量
- 风险识别：预判技术风险、业务风险、项目风险，提供缓解策略
- 优化建议：基于行业最佳实践，提供具体可操作的改进方案

## 分析方法论
- 系统性思维：从项目整体角度分析需求的价值和影响
- 数据驱动：基于历史数据和行业基准进行科学评估
- 价值导向：始终关注业务价值和投资回报率
- 持续改进：提供可执行的改进路径和监控机制

## 行业洞察力
- 跨行业对比：能够将不同行业的成功经验进行横向对比和借鉴
- 趋势预测：基于行业发展趋势预测技术演进和需求变化
- 最佳实践：掌握各行业在需求分析、项目管理、质量控制方面的最佳实践
- 创新思维：善于发现新的分析方法和优化机会

你的使命是运用专业知识和丰富经验，分析和优化三级功能点定义，确保每个功能点都能为项目成功提供可靠的数据支撑和决策依据。`},
		{Role: "user", Content: prompt},
	}

	config := &AIConfig{
		MaxTokens:   32000,
		Temperature: 0.7,
		TopP:        0.9,
		Model:       "deepseek-reasoner",
	}

	// 运行时安全检查：确保AI服务可用
	if s.aiService == nil {
		global.GVA_LOG.Warn("AI服务不可用，返回空的优化建议")
		return []OptimizationSuggestion{}, nil
	}

	response, err := s.aiService.ChatCompletion(context.Background(), messages, config)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	return s.parseOptimizationResponse(response.Text)
}

// generateExpansionSuggestions 生成扩充建议
func (s *Level3AnalyzerService) generateExpansionSuggestions(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, knowledgeRefs []KnowledgeReference) ([]ExpansionSuggestion, error) {
	// 构建AI分析prompt
	prompt := s.buildExpansionPrompt(requirement, cycle, knowledgeRefs)
	
	// 调用AI服务
	messages := []APIMessage{
		{Role: "system", Content: `你是一位国际认证的NESMA（荷兰软件度量协会）功能点分析专家，拥有20年以上的软件度量和需求分析经验。你的专业资质包括：

## 核心专业能力
- NESMA 2.2国际标准专家：精通NESMA功能点分析的所有细节、规则和最佳实践
- ISO/IEC 14143标准认证：具备国际软件度量标准认证资质
- 多行业项目经验：在金融、制造、医疗、政府、电商等15+个行业有丰富的项目实践
- 软件工程专家：深度理解软件开发生命周期、架构设计、质量保证体系
- 需求工程大师：擅长需求获取、分析、建模、验证和管理的全流程

## 专业技能矩阵
- 精确功能点识别：准确识别EI/EO/EQ/ILF/EIF五种功能类型，准确率>95%
- 科学复杂度评估：基于DET/RET/FTR的科学评估方法，考虑技术、业务、数据三个维度
- 质量评估：从完整性、一致性、可追溯性、可测试性等多维度评估需求质量
- 风险识别：预判技术风险、业务风险、项目风险，提供缓解策略
- 优化建议：基于行业最佳实践，提供具体可操作的改进方案

## 分析方法论
- 系统性思维：从项目整体角度分析需求的价值和影响
- 数据驱动：基于历史数据和行业基准进行科学评估
- 价值导向：始终关注业务价值和投资回报率
- 持续改进：提供可执行的改进路径和监控机制

## 行业洞察力
- 跨行业对比：能够将不同行业的成功经验进行横向对比和借鉴
- 趋势预测：基于行业发展趋势预测技术演进和需求变化
- 最佳实践：掌握各行业在需求分析、项目管理、质量控制方面的最佳实践
- 创新思维：善于发现新的分析方法和优化机会

你的使命是运用专业知识和丰富经验，基于业务需求扩充和完善三级功能点，确保每个功能点都能为项目成功提供可靠的数据支撑和决策依据。`},
		{Role: "user", Content: prompt},
	}

	config := &AIConfig{
		MaxTokens:   32000,
		Temperature: 0.8,
		TopP:        0.9,
		Model:       "deepseek-reasoner",
	}

	// 运行时安全检查：确保AI服务可用
	if s.aiService == nil {
		global.GVA_LOG.Warn("AI服务不可用，返回空的扩充建议")
		return []ExpansionSuggestion{}, nil
	}

	response, err := s.aiService.ChatCompletion(context.Background(), messages, config)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	return s.parseExpansionResponse(response.Text)
}

// buildOptimizationPrompt 构建优化建议的prompt
func (s *Level3AnalyzerService) buildOptimizationPrompt(requirement *nesma.NesmaRequirement, knowledgeRefs []KnowledgeReference) string {
	var knowledgeContext strings.Builder
	for _, ref := range knowledgeRefs {
		knowledgeContext.WriteString(fmt.Sprintf("- %s: %s\n", ref.Title, ref.Category))
	}

	return fmt.Sprintf(`
请分析以下三级功能点并提供优化建议：

【当前功能点信息】
标题: %s
描述: %s
分类: %s
业务价值: %s

【相关知识库内容】
%s

【分析要求】
1. 分析标题是否清晰、准确、符合NESMA标准
2. 分析描述是否完整、具体、可实现
3. 分析分类是否合理
4. 提供具体的优化建议

【输出格式】
请以JSON格式返回分析结果：
{
  "optimizations": [
    {
      "type": "title|description|category",
      "field": "字段名",
      "current_value": "当前值",
      "suggested_value": "建议值",
      "reason": "优化原因",
      "priority": 1-5,
      "confidence": 0.0-1.0,
      "category": "优化类别"
    }
  ]
}`,
		requirement.Title,
		requirement.Description,
		requirement.Category,
		requirement.BusinessValue,
		knowledgeContext.String(),
	)
}

// buildExpansionPrompt 构建扩充建议的prompt
func (s *Level3AnalyzerService) buildExpansionPrompt(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, knowledgeRefs []KnowledgeReference) string {
	var knowledgeContext strings.Builder
	for _, ref := range knowledgeRefs {
		knowledgeContext.WriteString(fmt.Sprintf("- %s: %s\n", ref.Title, ref.Category))
	}

	return fmt.Sprintf(`
请基于以下三级功能点分析是否需要扩充相关功能点：

【当前功能点信息】
标题: %s
描述: %s
分类: %s
项目领域: %s
项目描述: %s

【相关知识库内容】
%s

【分析要求】
1. 分析当前功能点是否完整，是否遗漏了相关功能
2. 基于项目领域和业务特点，建议补充的功能点
3. 确保建议的功能点合理、可实现、有价值
4. 每个建议都要有充分的理由和知识库支持

【输出格式】
请以JSON格式返回分析结果：
{
  "expansions": [
    {
      "suggested_title": "建议的功能点标题",
      "suggested_description": "建议的功能点描述",
      "suggested_category": "建议的功能点分类",
      "justification": "建议理由",
      "priority": 1-5,
      "confidence": 0.0-1.0,
      "related_knowledge": "相关知识库内容",
      "estimated_complexity": "简单|中等|复杂"
    }
  ]
}`,
		requirement.Title,
		requirement.Description,
		requirement.Category,
		cycle.Project.Domain,
		cycle.Project.Description,
		knowledgeContext.String(),
	)
}

// parseOptimizationResponse 解析优化建议响应
func (s *Level3AnalyzerService) parseOptimizationResponse(response string) ([]OptimizationSuggestion, error) {
	// 简化的JSON解析
	var result struct {
		Optimizations []OptimizationSuggestion `json:"optimizations"`
	}

	// 尝试直接解析JSON
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果解析失败，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonContent := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
				global.GVA_LOG.Warn("解析优化建议响应失败", zap.String("response", response), zap.Error(err))
				return []OptimizationSuggestion{}, nil
			}
		} else {
			global.GVA_LOG.Warn("无法从响应中提取JSON", zap.String("response", response))
			return []OptimizationSuggestion{}, nil
		}
	}

	return result.Optimizations, nil
}

// parseExpansionResponse 解析扩充建议响应
func (s *Level3AnalyzerService) parseExpansionResponse(response string) ([]ExpansionSuggestion, error) {
	// 简化的JSON解析
	var result struct {
		Expansions []ExpansionSuggestion `json:"expansions"`
	}

	// 尝试直接解析JSON
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果解析失败，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonContent := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
				global.GVA_LOG.Warn("解析扩充建议响应失败", zap.String("response", response), zap.Error(err))
				return []ExpansionSuggestion{}, nil
			}
		} else {
			global.GVA_LOG.Warn("无法从响应中提取JSON", zap.String("response", response))
			return []ExpansionSuggestion{}, nil
		}
	}

	return result.Expansions, nil
}

// calculateAnalysisConfidence 计算分析置信度
func (s *Level3AnalyzerService) calculateAnalysisConfidence(optimizations []OptimizationSuggestion, expansions []ExpansionSuggestion, knowledgeRefs []KnowledgeReference) float64 {
	var totalConfidence float64
	var count int

	// 计算优化建议的平均置信度
	for _, opt := range optimizations {
		totalConfidence += opt.Confidence
		count++
	}

	// 计算扩充建议的平均置信度
	for _, exp := range expansions {
		totalConfidence += exp.Confidence
		count++
	}

	// 知识库支持度影响置信度
	knowledgeBonus := float64(len(knowledgeRefs)) * 0.05
	if knowledgeBonus > 0.2 {
		knowledgeBonus = 0.2
	}

	if count == 0 {
		return 0.5 + knowledgeBonus
	}

	baseConfidence := totalConfidence / float64(count)
	return baseConfidence + knowledgeBonus
}

// generateAnalysisNotes 生成分析备注
func (s *Level3AnalyzerService) generateAnalysisNotes(optimizations []OptimizationSuggestion, expansions []ExpansionSuggestion) string {
	var notes strings.Builder
	
	if len(optimizations) > 0 {
		notes.WriteString(fmt.Sprintf("发现 %d 个优化建议；", len(optimizations)))
	}
	
	if len(expansions) > 0 {
		notes.WriteString(fmt.Sprintf("发现 %d 个扩充建议；", len(expansions)))
	}
	
	if len(optimizations) == 0 && len(expansions) == 0 {
		notes.WriteString("当前功能点定义较为完善，无需特别优化。")
	}

	return notes.String()
}

// ApplyOptimizationSuggestion 应用优化建议
func (s *Level3AnalyzerService) ApplyOptimizationSuggestion(requirementID uint, suggestion *OptimizationSuggestion) error {
	var requirement nesma.NesmaRequirement
	if err := s.db.First(&requirement, requirementID).Error; err != nil {
		return fmt.Errorf("获取功能点失败: %w", err)
	}

	// 根据建议类型应用优化
	switch suggestion.Type {
	case "title":
		requirement.Title = suggestion.SuggestedValue
	case "description":
		requirement.Description = suggestion.SuggestedValue
	case "category":
		requirement.Category = suggestion.SuggestedValue
	}

	// 更新AI分析状态
	now := time.Now()
	requirement.AIAnalysisStatus = "optimized"
	requirement.AIAnalysisTime = &now

	return s.db.Save(&requirement).Error
}

// CreateExpansionRequirement 创建扩充功能点
func (s *Level3AnalyzerService) CreateExpansionRequirement(cycleID uint, parentID uint, suggestion *ExpansionSuggestion) (*nesma.NesmaRequirement, error) {
	// 获取父功能点信息
	var parentRequirement nesma.NesmaRequirement
	if err := s.db.First(&parentRequirement, parentID).Error; err != nil {
		return nil, fmt.Errorf("获取父功能点失败: %w", err)
	}

	// 创建新的三级功能点
	newRequirement := &nesma.NesmaRequirement{
		ProjectID:          parentRequirement.ProjectID,
		CycleID:            &cycleID,
		ParentID:           &parentID,
		Level:              3,
		Title:              suggestion.SuggestedTitle,
		Description:        suggestion.SuggestedDescription,
		Category:           suggestion.SuggestedCategory,
		Status:             "pending",
		Priority:           suggestion.Priority,
		Complexity:         suggestion.EstimatedComplexity,
		AIAnalysisStatus:   "generated",
		AIDescription:      suggestion.SuggestedDescription,
		AIGeneratedTitle:   suggestion.SuggestedTitle,
		AIConfidenceScore:  &suggestion.Confidence,
	}

	now := time.Now()
	newRequirement.AIAnalysisTime = &now

	if err := s.db.Create(newRequirement).Error; err != nil {
		return nil, fmt.Errorf("创建扩充功能点失败: %w", err)
	}

	return newRequirement, nil
}

// GetLevel3AnalyzerService 获取三级功能点分析服务实例
func GetLevel3AnalyzerService() *Level3AnalyzerService {
	return NewLevel3AnalyzerService()
}