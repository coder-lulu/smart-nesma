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

// Level4GenerationResult 四级功能点生成结果
type Level4GenerationResult struct {
	ParentRequirement   *nesma.NesmaRequirement  `json:"parent_requirement"`
	GeneratedLevel4s    []Level4Suggestion       `json:"generated_level4s"`
	AnalysisConfidence  float64                  `json:"analysis_confidence"`
	ProcessingTime      int64                    `json:"processing_time_ms"`
	GenerationNotes     string                   `json:"generation_notes"`
	KnowledgeReferences []KnowledgeReference     `json:"knowledge_refs"`
}

// Level4Suggestion 四级功能点建议
type Level4Suggestion struct {
	SuggestedTitle       string  `json:"suggested_title"`
	SuggestedDescription string  `json:"suggested_description"`
	SuggestedCode        string  `json:"suggested_code"`
	FunctionType         string  `json:"function_type"`         // EI/EO/EQ/ILF/EIF
	BusinessValue        string  `json:"business_value"`
	AcceptanceCriteria   string  `json:"acceptance_criteria"`
	EstimatedComplexity  string  `json:"estimated_complexity"`   // 简单/中等/复杂
	RecommendedAFP       float64 `json:"recommended_afp"`
	RecommendedUFP       float64 `json:"recommended_ufp"`
	Priority             int     `json:"priority"`               // 1-5
	Confidence           float64 `json:"confidence"`             // 0-1
	GenerationReason     string  `json:"generation_reason"`
	RelatedKnowledge     string  `json:"related_knowledge"`
	RelatedRequirements  string  `json:"related_requirements"`   // 相关需求
}

// Level4GeneratorService 四级功能点生成服务
type Level4GeneratorService struct {
	db               *gorm.DB
	aiService        AIService
	knowledgeService *KnowledgeService
}

// NewLevel4GeneratorService 创建四级功能点生成服务
func NewLevel4GeneratorService() *Level4GeneratorService {
	// 检查数据库连接是否可用
	if global.GVA_DB == nil {
		global.GVA_LOG.Error("数据库连接未初始化，无法创建Level4GeneratorService")
		return nil
	}
	
	// 尝试获取AI服务
	var aiService AIService
	aiService = GetAIService()
	
	// 如果获取失败，尝试直接创建DeepSeek服务
	if aiService == nil {
		global.GVA_LOG.Warn("GetAIService返回nil，尝试直接创建DeepSeek服务")
		if global.GVA_CONFIG.AI.DeepSeek.APIKey != "" {
			aiService = NewDeepSeekService(
				global.GVA_CONFIG.AI.DeepSeek.APIKey,
				global.GVA_CONFIG.AI.DeepSeek.BaseURL,
				global.GVA_CONFIG.AI.DeepSeek.Model,
			)
			global.GVA_LOG.Info("直接创建DeepSeek服务成功")
		} else {
			global.GVA_LOG.Error("DeepSeek API Key未配置，无法创建AI服务")
		}
	}
	
	// 创建知识库服务
	knowledgeService := NewKnowledgeService()
	
	service := &Level4GeneratorService{
		db:               global.GVA_DB,
		aiService:        aiService,
		knowledgeService: knowledgeService,
	}
	
	global.GVA_LOG.Info("Level4GeneratorService创建成功", 
		zap.Bool("dbAvailable", service.db != nil),
		zap.Bool("aiServiceAvailable", service.aiService != nil),
		zap.Bool("knowledgeServiceAvailable", service.knowledgeService != nil))
	
	return service
}

// GenerateLevel4Requirements 生成四级功能点
func (s *Level4GeneratorService) GenerateLevel4Requirements(cycleID uint, level3RequirementIDs []uint) ([]*Level4GenerationResult, error) {
	global.GVA_LOG.Info("开始生成四级功能点", zap.Uint("cycleID", cycleID), zap.Any("level3RequirementIDs", level3RequirementIDs))

	var results []*Level4GenerationResult
	
	// 获取周期信息
	var cycle nesma.NesmaProjectCycle
	if err := s.db.Preload("Project").First(&cycle, cycleID).Error; err != nil {
		return nil, fmt.Errorf("获取项目周期失败: %w", err)
	}

	// 获取需要生成四级功能点的三级功能点
	var level3Requirements []nesma.NesmaRequirement
	query := s.db.Where("cycle_id = ? AND level = 3", cycleID)
	if len(level3RequirementIDs) > 0 {
		query = query.Where("id IN ?", level3RequirementIDs)
	}
	
	if err := query.Preload("Children", "level = 4").Find(&level3Requirements).Error; err != nil {
		return nil, fmt.Errorf("获取三级功能点失败: %w", err)
	}

	// 逐个生成四级功能点
	for _, level3Req := range level3Requirements {
		result, err := s.generateLevel4ForRequirement(&level3Req, &cycle)
		if err != nil {
			global.GVA_LOG.Error("生成四级功能点失败", zap.Uint("level3RequirementID", level3Req.ID), zap.Error(err))
			continue
		}
		results = append(results, result)
	}

	return results, nil
}

// generateLevel4ForRequirement 为单个三级功能点生成四级功能点
func (s *Level4GeneratorService) generateLevel4ForRequirement(level3Req *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) (*Level4GenerationResult, error) {
	startTime := time.Now()
	
	// 1. 搜索相关知识库内容
	knowledgeRefs, err := s.searchRelevantKnowledge(level3Req, cycle)
	if err != nil {
		global.GVA_LOG.Warn("搜索相关知识失败", zap.Error(err))
		knowledgeRefs = []KnowledgeReference{}
	}

	// 2. 生成四级功能点建议
	level4Suggestions, err := s.generateLevel4Suggestions(level3Req, cycle, knowledgeRefs)
	if err != nil {
		return nil, fmt.Errorf("生成四级功能点建议失败: %w", err)
	}

	// 3. 计算整体置信度
	confidence := s.calculateGenerationConfidence(level4Suggestions, knowledgeRefs)

	// 4. 构建生成结果
	result := &Level4GenerationResult{
		ParentRequirement:   level3Req,
		GeneratedLevel4s:    level4Suggestions,
		AnalysisConfidence:  confidence,
		ProcessingTime:      time.Since(startTime).Milliseconds(),
		GenerationNotes:     s.generateGenerationNotes(level3Req, level4Suggestions),
		KnowledgeReferences: knowledgeRefs,
	}

	global.GVA_LOG.Info("四级功能点生成完成", 
		zap.Uint("level3RequirementID", level3Req.ID),
		zap.Int("生成数量", len(level4Suggestions)),
		zap.Float64("置信度", confidence),
	)

	return result, nil
}

// searchRelevantKnowledge 搜索相关知识库内容
func (s *Level4GeneratorService) searchRelevantKnowledge(level3Req *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) ([]KnowledgeReference, error) {
	// 空指针检查
	if s.knowledgeService == nil {
		global.GVA_LOG.Warn("知识库服务未初始化，跳过知识库搜索")
		return []KnowledgeReference{}, nil
	}
	
	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s %s %s 四级功能点", level3Req.Title, level3Req.Description, cycle.Project.Domain)
	
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

// generateLevel4Suggestions 生成四级功能点建议
func (s *Level4GeneratorService) generateLevel4Suggestions(level3Req *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, knowledgeRefs []KnowledgeReference) ([]Level4Suggestion, error) {
	// 检查AI服务是否可用
	if s.aiService == nil {
		global.GVA_LOG.Error("AI服务未初始化，无法生成四级功能点建议")
		return nil, fmt.Errorf("AI服务未初始化")
	}
	
	// 构建AI分析prompt
	prompt := s.buildLevel4GenerationPrompt(level3Req, cycle, knowledgeRefs)
	
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

你的使命是运用专业知识和丰富经验，基于三级功能点为团队生成最具价值、最准确、最全面的四级功能点建议，确保每个功能点都能为项目成功提供可靠的数据支撑和决策依据。`},
		{Role: "user", Content: prompt},
	}

	config := &AIConfig{
		MaxTokens:   32000,
		Temperature: 0.8,
		TopP:        0.9,
		Model:       "deepseek-reasoner",
	}

	// 创建带超时的context (180秒)
	ctx, cancel := context.WithTimeout(context.Background(), 900*time.Second)
	defer cancel()

	global.GVA_LOG.Info("开始调用AI服务生成四级功能点建议", 
		zap.Uint("level3RequirementID", level3Req.ID),
		zap.String("aiServiceType", fmt.Sprintf("%T", s.aiService)))

	response, err := s.aiService.ChatCompletion(ctx, messages, config)
	if err != nil {
		global.GVA_LOG.Error("AI服务调用失败", 
			zap.Uint("level3RequirementID", level3Req.ID),
			zap.Error(err))
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	global.GVA_LOG.Info("AI服务调用成功，开始解析响应", 
		zap.Uint("level3RequirementID", level3Req.ID))

	// 解析AI响应
	return s.parseLevel4Response(response.Text)
}

// buildLevel4GenerationPrompt 构建四级功能点生成的prompt
func (s *Level4GeneratorService) buildLevel4GenerationPrompt(level3Req *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, knowledgeRefs []KnowledgeReference) string {
	var knowledgeContext strings.Builder
	for _, ref := range knowledgeRefs {
		knowledgeContext.WriteString(fmt.Sprintf("- %s: %s\n", ref.Title, ref.Category))
	}

	// 获取现有的四级功能点作为参考
	existingLevel4s := s.getExistingLevel4s(level3Req.ID)
	var existingContext strings.Builder
	for _, existing := range existingLevel4s {
		existingContext.WriteString(fmt.Sprintf("- %s: %s\n", existing.Title, existing.Description))
	}

	return fmt.Sprintf(`
请基于以下三级功能点生成详细的四级功能点：

【三级功能点信息】
标题: %s
描述: %s
分类: %s
业务价值: %s
项目领域: %s

【现有四级功能点】
%s

【相关知识库内容】
%s

【生成要求】
1. 基于三级功能点的业务逻辑，拆分出具体的四级功能点
2. 每个四级功能点应该是独立的、可计数的功能单元
3. 确保四级功能点覆盖三级功能点的所有核心业务场景
4. 按照NESMA标准进行功能类型分类（EI/EO/EQ/ILF/EIF）
5. 如果已有四级功能点，应该补充缺失的或优化现有的
6. 每个四级功能点都要有明确的业务价值和验收标准
7. 每个四级功能点都要有明确的复杂度评估（简单/中等/复杂）
8. 每个四级功能点都要有明确的推荐AFP和UFP值
9. 每个四级功能点都要有明确的优先级（1-5，1最高）
10. 每个四级功能点都要有明确的置信度（0.0-1.0）
11. 每个四级功能点都要有明确的生成理由
12. 在生成四级功能点时如果可以扩展除了标准功能分类之外的功能类型，可以结合项目信息和三级功能点信息进行扩展
13. 建议生成3-8个四级功能点，确保覆盖完整
14. 每个四级功能点得描述要详细，但又不冗长，要符合人类自然语言，涵盖到功能点所涉及业务场景，并描述清楚功能点所涉及的业务逻辑，不要写输入、处理和输出，使用自然语言描述，输出长度在100-200字左右

【输出格式】
请以JSON格式返回生成结果：
{
  "level4_suggestions": [
    {
      "suggested_title": "四级功能点标题",
      "suggested_description": "详细功能描述，描述信息要符合人类自然语言，涵盖到功能点所涉及业务场景，并描述清楚功能点所涉及的业务逻辑，要详细但不冗长",
      "suggested_code": "功能编码",
      "function_type": "EI/EO/EQ/ILF/EIF",
      "business_value": "业务价值说明",
      "acceptance_criteria": "验收标准",
      "estimated_complexity": "简单/中等/复杂",
      "recommended_afp": 3.0,
      "recommended_ufp": 3.0,
      "priority": 1,
      "confidence": 0.8,
      "generation_reason": "生成理由，理由详细充分，理由要符合人类自然语言，要详细但不冗长",
      "related_knowledge": "相关知识库内容",
      "related_requirements": "相关需求"
    }
  ]
}

【注意事项】
- 四级功能点应该是原子级别的功能单元
- 确保每个功能点都有明确的输入、处理和输出，但在描述功能是不要写输入、处理和输出，使用自然语言描述，详细但又不冗长
- 功能类型分类要准确，符合NESMA标准：
  * EI（External Input）：外部输入，从用户或其他系统接收数据
  * EO（External Output）：外部输出，向用户或其他系统发送数据
  * EQ（External Query）：外部查询，从系统获取数据但不修改
  * ILF（Internal Logical File）：内部逻辑文件，系统内部维护的数据
  * EIF（External Interface File）：外部接口文件，其他系统维护的数据
- 复杂度评估要基于功能的业务复杂性和技术实现难度
- 功能点数推荐要符合NESMA计数规则
- AFP和UFP值应该根据功能复杂度和数据元素类型来确定`,
		level3Req.Title,
		level3Req.Description,
		level3Req.Category,
		level3Req.BusinessValue,
		cycle.Project.Domain,
		existingContext.String(),
		knowledgeContext.String(),
	)
}

// getExistingLevel4s 获取现有的四级功能点
func (s *Level4GeneratorService) getExistingLevel4s(parentID uint) []nesma.NesmaRequirement {
	var level4s []nesma.NesmaRequirement
	s.db.Where("parent_id = ? AND level = 4", parentID).Find(&level4s)
	return level4s
}

// parseLevel4Response 解析四级功能点响应
func (s *Level4GeneratorService) parseLevel4Response(response string) ([]Level4Suggestion, error) {
	// 简化的JSON解析
	var result struct {
		Level4Suggestions []Level4Suggestion `json:"level4_suggestions"`
	}

	// 尝试直接解析JSON
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		// 如果解析失败，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonContent := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
				global.GVA_LOG.Warn("解析四级功能点响应失败", zap.String("response", response), zap.Error(err))
				return s.createFallbackLevel4Suggestions(), nil
			}
		} else {
			global.GVA_LOG.Warn("无法从响应中提取JSON", zap.String("response", response))
			return s.createFallbackLevel4Suggestions(), nil
		}
	}

	return result.Level4Suggestions, nil
}

// createFallbackLevel4Suggestions 创建备用的四级功能点建议
func (s *Level4GeneratorService) createFallbackLevel4Suggestions() []Level4Suggestion {
	return []Level4Suggestion{
		{
			SuggestedTitle:       "数据查询功能",
			SuggestedDescription: "提供基础的数据查询和检索功能",
			SuggestedCode:        "L4_001",
			FunctionType:         "EQ",
			BusinessValue:        "支持用户快速获取所需信息",
			AcceptanceCriteria:   "能够准确返回查询结果",
			EstimatedComplexity:  "简单",
			RecommendedAFP:       3.0,
			RecommendedUFP:       3.0,
			Priority:             3,
			Confidence:           0.6,
			GenerationReason:     "AI响应解析失败，使用默认功能点",
			RelatedKnowledge:     "通用查询功能模板",
			RelatedRequirements:  "基础查询需求",
		},
	}
}

// calculateGenerationConfidence 计算生成置信度
func (s *Level4GeneratorService) calculateGenerationConfidence(suggestions []Level4Suggestion, knowledgeRefs []KnowledgeReference) float64 {
	var totalConfidence float64
	var count int

	// 计算建议的平均置信度
	for _, suggestion := range suggestions {
		totalConfidence += suggestion.Confidence
		count++
	}

	// 知识库支持度影响置信度
	knowledgeBonus := float64(len(knowledgeRefs)) * 0.08
	if knowledgeBonus > 0.25 {
		knowledgeBonus = 0.25
	}

	if count == 0 {
		return 0.5 + knowledgeBonus
	}

	baseConfidence := totalConfidence / float64(count)
	return baseConfidence + knowledgeBonus
}

// generateGenerationNotes 生成生成备注
func (s *Level4GeneratorService) generateGenerationNotes(level3Req *nesma.NesmaRequirement, suggestions []Level4Suggestion) string {
	var notes strings.Builder
	
	notes.WriteString(fmt.Sprintf("基于三级功能点「%s」生成 %d 个四级功能点；", level3Req.Title, len(suggestions)))
	
	// 统计功能类型分布
	typeCount := make(map[string]int)
	for _, suggestion := range suggestions {
		typeCount[suggestion.FunctionType]++
	}
	
	notes.WriteString("功能类型分布：")
	for funcType, count := range typeCount {
		notes.WriteString(fmt.Sprintf("%s:%d个，", funcType, count))
	}
	
	if len(suggestions) == 0 {
		notes.WriteString("未能生成有效的四级功能点建议。")
	}

	return notes.String()
}

// CreateLevel4Requirement 创建四级功能点
func (s *Level4GeneratorService) CreateLevel4Requirement(cycleID uint, parentID uint, suggestion *Level4Suggestion) (*nesma.NesmaRequirement, error) {
	// 获取父功能点信息
	var parentRequirement nesma.NesmaRequirement
	if err := s.db.First(&parentRequirement, parentID).Error; err != nil {
		return nil, fmt.Errorf("获取父功能点失败: %w", err)
	}

	// 验证父功能点是三级
	if parentRequirement.Level != 3 {
		return nil, fmt.Errorf("父功能点必须是三级功能点")
	}

	// 创建新的四级功能点
	newRequirement := &nesma.NesmaRequirement{
		ProjectID:            parentRequirement.ProjectID,
		CycleID:              &cycleID,
		ParentID:             &parentID,
		Level:                4,
		Code:                 suggestion.SuggestedCode,
		Title:                suggestion.SuggestedTitle,
		Description:          suggestion.SuggestedDescription,
		Category:             parentRequirement.Category,
		Status:               "pending",
		Priority:             suggestion.Priority,
		Complexity:           suggestion.EstimatedComplexity,
		BusinessValue:        suggestion.BusinessValue,
		AcceptanceCriteria:   suggestion.AcceptanceCriteria,
		FunctionType:         suggestion.FunctionType,
		AFP:                  suggestion.RecommendedAFP,
		UFP:                  suggestion.RecommendedUFP,
		AIAnalysisStatus:     "generated",
		AIDescription:        suggestion.SuggestedDescription,
		AIGeneratedTitle:     suggestion.SuggestedTitle,
		AIConfidenceScore:    &suggestion.Confidence,
		RecommendedAFP:       &suggestion.RecommendedAFP,
		RecommendedUFP:       &suggestion.RecommendedUFP,
		Notes:                suggestion.GenerationReason,
	}

	now := time.Now()
	newRequirement.AIAnalysisTime = &now

	if err := s.db.Create(newRequirement).Error; err != nil {
		return nil, fmt.Errorf("创建四级功能点失败: %w", err)
	}

	return newRequirement, nil
}

// BatchCreateLevel4Requirements 批量创建四级功能点
func (s *Level4GeneratorService) BatchCreateLevel4Requirements(cycleID uint, creations []Level4CreationRequest) (*Level4BatchCreationResult, error) {
	var successCount, failedCount int
	var createdRequirements []uint
	var errors []string

	for _, creation := range creations {
		newRequirement, err := s.CreateLevel4Requirement(cycleID, creation.ParentID, creation.Suggestion)
		if err != nil {
			failedCount++
			errors = append(errors, err.Error())
			global.GVA_LOG.Error("批量创建四级功能点失败", 
				zap.Uint("cycleID", cycleID),
				zap.Uint("parentID", creation.ParentID),
				zap.Error(err),
			)
		} else {
			successCount++
			createdRequirements = append(createdRequirements, newRequirement.ID)
		}
	}

	global.GVA_LOG.Info("批量创建四级功能点完成", 
		zap.Int("成功数量", successCount),
		zap.Int("失败数量", failedCount),
	)

	return &Level4BatchCreationResult{
		SuccessCount:         successCount,
		FailedCount:          failedCount,
		CreatedRequirements:  createdRequirements,
		Errors:               errors,
	}, nil
}

// 辅助结构体
type Level4CreationRequest struct {
	ParentID   uint              `json:"parent_id"`
	Suggestion *Level4Suggestion `json:"suggestion"`
}

type Level4BatchCreationResult struct {
	SuccessCount        int      `json:"success_count"`
	FailedCount         int      `json:"failed_count"`
	CreatedRequirements []uint   `json:"created_requirements"`
	Errors              []string `json:"errors"`
}

// Level4GenerationTaskResult L4生成任务结果
type Level4GenerationTaskResult struct {
	TotalCount           int                      `json:"total_count"`
	SuccessCount         int                      `json:"success_count"`
	FailedCount          int                      `json:"failed_count"`
	L4SuggestionsResults []map[string]interface{} `json:"l4_suggestions_results"` // L4建议结果
	Errors               []string                 `json:"errors"`
	CreatedRequirements  []nesma.NesmaRequirement `json:"created_requirements,omitempty"` // 创建的需求列表
}

// GetLevel4GeneratorService 获取四级功能点生成服务实例
func GetLevel4GeneratorService() *Level4GeneratorService {
	service := NewLevel4GeneratorService()
	if service == nil {
		global.GVA_LOG.Error("无法创建Level4GeneratorService，数据库连接可能未初始化")
		return nil
	}
	
	// 验证服务的关键组件
	if service.db == nil {
		global.GVA_LOG.Error("Level4GeneratorService数据库连接为空")
		return nil
	}
	
	if service.aiService == nil {
		global.GVA_LOG.Error("Level4GeneratorService AI服务为空")
		return nil
	}
	
	global.GVA_LOG.Info("获取Level4GeneratorService实例成功")
	return service
}

// ==================== 异步任务支持 ====================

// GenerateLevel4Async 异步生成四级功能点
func (s *Level4GeneratorService) GenerateLevel4Async(req *request.AsyncLevel4GenerationRequest) (*nesma.NesmaRequirementAnalysisTask, error) {
	global.GVA_LOG.Info("开始异步生成四级功能点", 
		zap.Uint("cycleID", req.CycleID), 
		zap.Uint("versionID", req.VersionID),
		zap.Int("l3Count", len(req.L3RequirementIDs)))

	// 创建任务
	task := &nesma.NesmaRequirementAnalysisTask{
		ProjectID:       req.ProjectID,
		CycleID:         req.CycleID,
		SourceVersionID: req.VersionID,
		TaskType:        "level4_generation",
		Status:          "pending",
		Progress:        0,
		TotalCount:      len(req.L3RequirementIDs),
		Priority:        5,
	}

	// 保存任务配置
	configData := map[string]interface{}{
		"l3_requirement_ids":    req.L3RequirementIDs,
		"generation_strategy":   req.GenerationStrategy,
		"complexity_level":      req.ComplexityLevel,
		"max_l4_count":         req.MaxL4Count,
		"include_knowledge_base": req.IncludeKnowledgeBase,
		"auto_save":            req.AutoSave,
	}
	configJSON, _ := json.Marshal(configData)
	task.Config = configJSON

	if err := s.db.Create(task).Error; err != nil {
		return nil, fmt.Errorf("创建异步任务失败: %w", err)
	}

	// 启动异步处理
	go s.processLevel4GenerationTask(task)

	return task, nil
}

// processLevel4GenerationTask 处理L4生成任务
func (s *Level4GeneratorService) processLevel4GenerationTask(task *nesma.NesmaRequirementAnalysisTask) {
	// 更新任务状态为运行中
	task.Status = "running"
	now := time.Now()
	task.StartTime = &now
	s.db.Save(task)

	global.GVA_LOG.Info("开始处理L4生成任务", zap.Uint("taskID", task.ID))

	// 解析任务配置
	var config map[string]interface{}
	if err := json.Unmarshal(task.Config, &config); err != nil {
		s.failTask(task, fmt.Sprintf("解析任务配置失败: %v", err))
		return
	}

	l3RequirementIDs := config["l3_requirement_ids"].([]interface{})
	var l3IDs []uint
	for _, id := range l3RequirementIDs {
		l3IDs = append(l3IDs, uint(id.(float64)))
	}

	// 获取周期信息
	var cycle nesma.NesmaProjectCycle
	if err := s.db.Preload("Project").First(&cycle, task.CycleID).Error; err != nil {
		s.failTask(task, fmt.Sprintf("获取项目周期失败: %v", err))
		return
	}

	// 获取L3需求
	var level3Requirements []nesma.NesmaRequirement
	if err := s.db.Where("id IN ? AND level = 3", l3IDs).Find(&level3Requirements).Error; err != nil {
		s.failTask(task, fmt.Sprintf("获取L3需求失败: %v", err))
		return
	}

	// 更新总数
	task.TotalCount = len(level3Requirements)
	s.db.Save(task)

	// 使用worker pool并发处理
	s.processL3RequirementsConcurrently(task, level3Requirements, &cycle, config)
}

// processL3RequirementsConcurrently 并发处理L3需求
func (s *Level4GeneratorService) processL3RequirementsConcurrently(
	task *nesma.NesmaRequirementAnalysisTask, 
	level3Requirements []nesma.NesmaRequirement, 
	cycle *nesma.NesmaProjectCycle,
	config map[string]interface{}) {

	// 创建worker pool（限制并发数为3，避免API频率限制）
	const maxWorkers = 3
	jobs := make(chan *nesma.NesmaRequirement, len(level3Requirements))
	results := make(chan *Level4ProcessResult, len(level3Requirements))

	// 启动workers
	for w := 0; w < maxWorkers; w++ {
		go s.level4Worker(jobs, results, cycle, config, task)
	}

	// 发送任务到worker
	for i := range level3Requirements {
		jobs <- &level3Requirements[i]
	}
	close(jobs)

	// 收集结果
	var allResults []*Level4ProcessResult
	var successCount, failedCount int
	var l4SuggestionsResults []map[string]interface{}
	var errors []string

	var createdRequirements []nesma.NesmaRequirement

	for i := 0; i < len(level3Requirements); i++ {
		result := <-results
		allResults = append(allResults, result)

		if result.Success {
			successCount++
			// 保存L4建议结果（不入库）
			l4SuggestionsResults = append(l4SuggestionsResults, map[string]interface{}{
				"l3_requirement_id": result.L3RequirementID,
				"l3_requirement":    result.L3Requirement,
				"l4_suggestions":    result.L4Suggestions,
			})
			createdRequirements = append(createdRequirements, result.CreatedRequirements...)
		} else {
			failedCount++
			errors = append(errors, result.Error)
		}

		// 更新进度
		progress := int(float64(i+1) / float64(len(level3Requirements)) * 100)
		task.Progress = progress
		task.ProcessedCount = i + 1
		task.SuccessCount = successCount
		task.FailedCount = failedCount
		s.db.Save(task)
	}

	// 完成任务
	s.completeTask(task, &Level4GenerationTaskResult{
		TotalCount:          len(level3Requirements),
		SuccessCount:        successCount,
		FailedCount:         failedCount,
		L4SuggestionsResults: l4SuggestionsResults,
		Errors:              errors,
		CreatedRequirements: createdRequirements,
	})
}

// level4Worker worker处理函数
func (s *Level4GeneratorService) level4Worker(
	jobs <-chan *nesma.NesmaRequirement, 
	results chan<- *Level4ProcessResult, 
	cycle *nesma.NesmaProjectCycle,
	config map[string]interface{},
	task *nesma.NesmaRequirementAnalysisTask) {

	for l3Req := range jobs {
		result := &Level4ProcessResult{L3RequirementID: l3Req.ID}

		// 生成L4功能点建议（不入库）
		l4Result, err := s.generateLevel4ForRequirement(l3Req, cycle)
		if err != nil {
			result.Success = false
			result.Error = fmt.Sprintf("L3需求[%d]生成失败: %v", l3Req.ID, err)
			results <- result
			continue
		}

		// 成功生成建议，不自动入库
		result.Success = true
		result.L4Suggestions = l4Result.GeneratedLevel4s
		result.L3Requirement = *l3Req  // 解引用指针
		
		results <- result
	}
}

// Level4ProcessResult L4处理结果
type Level4ProcessResult struct {
	L3RequirementID     uint                      `json:"l3_requirement_id"`
	Success             bool                      `json:"success"`
	Error               string                    `json:"error,omitempty"`
	L4Suggestions       []Level4Suggestion        `json:"l4_suggestions,omitempty"`
	L3Requirement       nesma.NesmaRequirement    `json:"l3_requirement,omitempty"`
	CreatedRequirements []nesma.NesmaRequirement  `json:"created_requirements,omitempty"` // 用于后续入库时使用
}

// saveLevel4RequirementsToDatabase 保存L4需求到数据库
func (s *Level4GeneratorService) saveLevel4RequirementsToDatabase(result *Level4GenerationResult, versionID uint) ([]uint, error) {
	var createdIDs []uint

	for _, suggestion := range result.GeneratedLevel4s {
		// 创建L4需求记录
		l4Requirement := &nesma.NesmaRequirement{
			ProjectID:             result.ParentRequirement.ProjectID,
			ParentID:              &result.ParentRequirement.ID,
			Level:                 4,
			Code:                  suggestion.SuggestedCode,
			Title:                 suggestion.SuggestedTitle,
			Description:           suggestion.SuggestedDescription,
			BusinessValue:         suggestion.BusinessValue,
			AcceptanceCriteria:    suggestion.AcceptanceCriteria,
			FunctionType:          suggestion.FunctionType,
			Complexity:            suggestion.EstimatedComplexity,
			AFP:                   suggestion.RecommendedAFP,
			UFP:                   suggestion.RecommendedUFP,
			Priority:              suggestion.Priority,
			Status:                "completed",
			CycleID:               result.ParentRequirement.CycleID,
			VersionID:             &versionID,
			AIAnalysisStatus:      "completed",
			AIDescription:         suggestion.SuggestedDescription,
			AIConfidenceScore:     &suggestion.Confidence,
		}

		if err := s.db.Create(l4Requirement).Error; err != nil {
			return nil, fmt.Errorf("保存L4需求失败: %w", err)
		}

		createdIDs = append(createdIDs, l4Requirement.ID)
	}

	return createdIDs, nil
}

// completeTask 完成任务
func (s *Level4GeneratorService) completeTask(task *nesma.NesmaRequirementAnalysisTask, result *Level4GenerationTaskResult) {
	task.Status = "completed"
	task.Progress = 100
	now := time.Now()
	task.EndTime = &now
	if task.StartTime != nil {
		duration := int(now.Sub(*task.StartTime).Seconds())
		task.Duration = &duration
	}

	// 保存结果
	resultJSON, _ := json.Marshal(result)
	task.Result = resultJSON
	task.Summary = fmt.Sprintf("成功生成 %d 个L3需求的L4建议，失败 %d 个", result.SuccessCount, result.FailedCount)

	s.db.Save(task)
	global.GVA_LOG.Info("L4生成任务完成", zap.Uint("taskID", task.ID), zap.Int("successCount", result.SuccessCount))
}

// failTask 任务失败
func (s *Level4GeneratorService) failTask(task *nesma.NesmaRequirementAnalysisTask, errorMsg string) {
	task.Status = "failed"
	task.ErrorMsg = errorMsg
	now := time.Now()
	task.EndTime = &now
	if task.StartTime != nil {
		duration := int(now.Sub(*task.StartTime).Seconds())
		task.Duration = &duration
	}

	s.db.Save(task)
	global.GVA_LOG.Error("L4生成任务失败", zap.Uint("taskID", task.ID), zap.String("error", errorMsg))
}

// ConfirmLevel4Requirements 确认并入库L4需求
func (s *Level4GeneratorService) ConfirmLevel4Requirements(confirmReq *request.ConfirmLevel4RequirementsRequest) (*Level4BatchCreationResult, error) {
	var successCount, failedCount int
	var createdRequirements []uint
	var errors []string

	for _, l4Req := range confirmReq.L4Requirements {
		// 获取父级L3需求信息，用于继承版本和周期信息
		var parentL3 nesma.NesmaRequirement
		if err := s.db.First(&parentL3, l4Req.ParentID).Error; err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("获取父级L3需求[%d]失败: %v", l4Req.ParentID, err))
			continue
		}

		// 验证父需求确实是L3级别
		if parentL3.Level != 3 {
			failedCount++
			errors = append(errors, fmt.Sprintf("父需求[%d]不是L3级别需求", l4Req.ParentID))
			continue
		}

		// 创建L4需求，继承父级的版本和周期信息
		newRequirement := &nesma.NesmaRequirement{
			ProjectID:             parentL3.ProjectID,
			ParentID:              &l4Req.ParentID,
			CycleID:               parentL3.CycleID,    // 继承父级周期
			VersionID:             parentL3.VersionID,  // 继承父级版本
			Level:                 4,
			Code:                  l4Req.Code,
			Title:                 l4Req.Title,
			Description:           l4Req.Description,
			Category:              parentL3.Category,   // 继承父级分类
			Status:                "confirmed",         // 确认状态
			Priority:              l4Req.Priority,
			Complexity:            l4Req.Complexity,
			BusinessValue:         l4Req.BusinessValue,
			AcceptanceCriteria:    l4Req.AcceptanceCriteria,
			FunctionType:          l4Req.FunctionType,
			AFP:                   l4Req.AFP,
			UFP:                   l4Req.UFP,
			AIAnalysisStatus:      "confirmed",
			AIDescription:         l4Req.Description,
			AIGeneratedTitle:      l4Req.Title,
			AIConfidenceScore:     &l4Req.Confidence,
			RecommendedAFP:        &l4Req.AFP,
			RecommendedUFP:        &l4Req.UFP,
			Notes:                 l4Req.Notes,
		}

		now := time.Now()
		newRequirement.AIAnalysisTime = &now

		if err := s.db.Create(newRequirement).Error; err != nil {
			failedCount++
			errors = append(errors, fmt.Sprintf("创建L4需求失败: %v", err))
			global.GVA_LOG.Error("确认入库L4需求失败", 
				zap.Uint("parentID", l4Req.ParentID),
				zap.String("title", l4Req.Title),
				zap.Error(err),
			)
		} else {
			successCount++
			createdRequirements = append(createdRequirements, newRequirement.ID)
			global.GVA_LOG.Info("成功确认入库L4需求", 
				zap.Uint("id", newRequirement.ID),
				zap.String("title", l4Req.Title),
				zap.Uint("parentID", l4Req.ParentID),
			)
		}
	}

	global.GVA_LOG.Info("L4需求确认入库完成", 
		zap.Int("成功数量", successCount),
		zap.Int("失败数量", failedCount),
	)

	return &Level4BatchCreationResult{
		SuccessCount:         successCount,
		FailedCount:          failedCount,
		CreatedRequirements:  createdRequirements,
		Errors:               errors,
	}, nil
}