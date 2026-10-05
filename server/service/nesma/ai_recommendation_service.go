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
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AIRecommendationService AI驱动的推荐服务
type AIRecommendationService struct {
	db             *gorm.DB
	aiService      AIService
	vectorService  VectorService
	knowledgeService *KnowledgeService
}

// RecommendationAdoption 推荐采纳记录
type RecommendationAdoption struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	UserID          uint           `json:"userId" gorm:"not null;index"`
	ProjectID       uint           `json:"projectId" gorm:"not null"`
	RecommendationID string        `json:"recommendationId" gorm:"not null"`
	Title           string         `json:"title" gorm:"not null"`
	Content         string         `json:"content" gorm:"type:text"`
	Category        string         `json:"category" gorm:"type:varchar(100)"`
	Confidence      float64        `json:"confidence" gorm:"default:0.0"`
	KnowledgeRefs   datatypes.JSON `json:"knowledgeRefs" gorm:"type:json"`
	UserFeedback    string         `json:"userFeedback" gorm:"type:text"`
	AdoptedAt       time.Time      `json:"adoptedAt"`
	CreatedAt       time.Time      `json:"createdAt"`
}

// TableName 自定义表名
func (RecommendationAdoption) TableName() string {
	return "recommendation_adoptions"
}

// AIRecommendationRequest AI推荐请求
type AIRecommendationRequest struct {
	UserID          uint                   `json:"userId" binding:"required"`
	ProjectID       uint                   `json:"projectId" binding:"required"`
	RequirementText string                 `json:"requirementText"`
	ProjectName     string                 `json:"projectName"`
	Domain          string                 `json:"domain"`
	RecommendationType string              `json:"recommendationType"`  // 推荐类型
	Algorithm       string                 `json:"algorithm"`          // 推荐算法
	Requirements    []ProjectRequirement   `json:"requirements"`       // 项目功能点
	Context         map[string]interface{} `json:"context"`
}

// ProjectRequirement 项目需求功能点
type ProjectRequirement struct {
	ID          uint                   `json:"id"`
	Level       int                    `json:"level"`     // 层级：1,2,3,4
	Code        string                 `json:"code"`      // 需求编号
	Title       string                 `json:"title"`     // 需求标题
	Description string                 `json:"description"` // 需求描述
	Children    []ProjectRequirement   `json:"children"`  // 子级需求
}

// AIRecommendationItem AI推荐项目
type AIRecommendationItem struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Content         string                 `json:"content"`
	Category        string                 `json:"category"`
	Confidence      float64                `json:"confidence"`
	KnowledgeRefs   []KnowledgeReference   `json:"knowledgeRefs"`
	Improvements    []string               `json:"improvements"`
	FunctionType    string                 `json:"functionType"`
	RiskWarnings    []string               `json:"riskWarnings"`
	BestPractices   []string               `json:"bestPractices"`
	CreatedAt       time.Time              `json:"createdAt"`
}

// KnowledgeReference 知识库引用
type KnowledgeReference struct {
	ID       uint    `json:"id"`
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Relevance float64 `json:"relevance"`
}

// AIRecommendationResponse AI推荐响应
type AIRecommendationResponse struct {
	Recommendations []AIRecommendationItem `json:"recommendations"`
	GeneratedBy     string                 `json:"generatedBy"`
	ProcessingTime  time.Duration          `json:"processingTime"`
	KnowledgeUsed   int                    `json:"knowledgeUsed"`
	Confidence      float64                `json:"confidence"`
	CreatedAt       time.Time              `json:"createdAt"`
}

// RecommendationAnalysisResult AI推荐分析结果
type RecommendationAnalysisResult struct {
	FunctionPointAdvice     string   `json:"functionPointAdvice"`
	OptimizationSuggestions []string `json:"optimizationSuggestions"`
	BestPractices          []string `json:"bestPractices"`
	RiskWarnings           []string `json:"riskWarnings"`
	FunctionType           string   `json:"functionType"`
	Confidence             float64  `json:"confidence"`
}

// NewAIRecommendationService 创建AI推荐服务
func NewAIRecommendationService() *AIRecommendationService {
	return &AIRecommendationService{
		db:               global.GVA_DB,
		aiService:        GetAIService(),
		vectorService:    GetVectorService(),
		knowledgeService: NewKnowledgeService(),
	}
}

// GetIntelligentRecommendations 获取智能推荐
func (s *AIRecommendationService) GetIntelligentRecommendations(ctx context.Context, req *AIRecommendationRequest) (*AIRecommendationResponse, error) {
	startTime := time.Now()
	
	// 1. 根据用户输入搜索相关知识库
	relatedKnowledge, err := s.searchRelatedKnowledge(ctx, req)
	if err != nil {
		global.GVA_LOG.Error("搜索相关知识失败", zap.Error(err))
		return nil, fmt.Errorf("搜索相关知识失败: %w", err)
	}

	// 2. 构建包含知识库内容的Prompt
	prompt := s.buildRecommendationPrompt(req, relatedKnowledge)

	// 3. 调用AI大模型生成推荐
	messages := []APIMessage{
		{Role: "user", Content: prompt},
	}
	config := &AIConfig{
		MaxTokens:   8192,  // 修复：DeepSeek-Chat模型的最大token数
		Temperature: 0.7,
		TopP:        0.9,   // 修复：设置有效的top_p值 (0, 1.0]
		Model:       "deepseek-chat",
	}
	
	aiResponse, err := s.aiService.ChatCompletion(ctx, messages, config)
	if err != nil {
		global.GVA_LOG.Error("AI推荐生成失败", zap.Error(err))
		return nil, fmt.Errorf("AI推荐生成失败: %w", err)
	}
	
	// 提取AI响应内容
	var responseContent string
	if len(aiResponse.Choices) > 0 {
		responseContent = aiResponse.Choices[0].Message.Content
	} else {
		return nil, fmt.Errorf("AI响应为空")
	}

	// 4. 解析AI推荐结果
	recommendations, err := s.parseAIRecommendations(responseContent, relatedKnowledge)
	if err != nil {
		global.GVA_LOG.Error("解析AI推荐结果失败", zap.Error(err))
		return nil, fmt.Errorf("解析AI推荐结果失败: %w", err)
	}

	// 5. 计算平均置信度
	avgConfidence := s.calculateAverageConfidence(recommendations)

	processingTime := time.Since(startTime)
	
	global.GVA_LOG.Info("AI推荐生成成功", 
		zap.Uint("userId", req.UserID),
		zap.Uint("projectId", req.ProjectID),
		zap.Int("knowledgeUsed", len(relatedKnowledge)),
		zap.Float64("confidence", avgConfidence),
		zap.Duration("processingTime", processingTime))

	return &AIRecommendationResponse{
		Recommendations: recommendations,
		GeneratedBy:     "AI-DeepSeek",
		ProcessingTime:  processingTime,
		KnowledgeUsed:   len(relatedKnowledge),
		Confidence:      avgConfidence,
		CreatedAt:       time.Now(),
	}, nil
}

// searchRelatedKnowledge 搜索相关知识库
func (s *AIRecommendationService) searchRelatedKnowledge(ctx context.Context, req *AIRecommendationRequest) ([]nesma.NesmaKnowledgeEntry, error) {
	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s %s %s", req.ProjectName, req.Domain, req.RequirementText)
	
	// 使用向量服务搜索相关知识（限制返回5个最相关的）
	if s.vectorService != nil {
		results, err := s.vectorService.SearchByContent(ctx, searchQuery, 5)
		if err != nil {
			global.GVA_LOG.Error("向量搜索失败", zap.Error(err))
		} else if len(results) > 0 {
			// 从向量搜索结果中提取相关内容，搜索知识库
			var searchTerms []string
			for _, result := range results {
				if result.Vector != nil && result.Vector.Content != "" {
					searchTerms = append(searchTerms, result.Vector.Content)
				}
			}
			
			if len(searchTerms) > 0 {
				var knowledge []nesma.NesmaKnowledgeEntry
				// 基于向量搜索结果的内容搜索知识库
				query := s.db.Where("status = ?", "active")
				for _, term := range searchTerms {
					query = query.Or("title LIKE ? OR content LIKE ?", "%"+term+"%", "%"+term+"%")
				}
				
				err := query.Order("confidence_score DESC, usage_count DESC").
					Limit(5).
					Find(&knowledge).Error
				
				if err == nil && len(knowledge) > 0 {
					return knowledge, nil
				}
			}
		}
	}
	
	// 如果向量服务不可用，使用传统文本搜索
	var knowledge []nesma.NesmaKnowledgeEntry
	err := s.db.Where("status = ? AND (title LIKE ? OR content LIKE ? OR category LIKE ?)", 
		"active", "%"+req.Domain+"%", "%"+req.RequirementText+"%", "%"+req.Domain+"%").
		Order("confidence_score DESC, usage_count DESC").
		Limit(5).
		Find(&knowledge).Error
	
	return knowledge, err
}

// buildRecommendationPrompt 构建推荐Prompt
func (s *AIRecommendationService) buildRecommendationPrompt(req *AIRecommendationRequest, knowledge []nesma.NesmaKnowledgeEntry) string {
	knowledgeText := s.formatKnowledgeForPrompt(knowledge)
	requirementsText := s.formatRequirementsForPrompt(req.Requirements)
	
	// 根据用户选择的推荐类型定制建议内容
	var recommendationFocus string
	switch req.RecommendationType {
	case "requirement":
		recommendationFocus = "需求模板和功能点分析"
	case "template":
		recommendationFocus = "文档模板和格式规范"
	case "agent":
		recommendationFocus = "AI助手和自动化工具"
	case "knowledge":
		recommendationFocus = "知识库内容和最佳实践"
	case "project":
		recommendationFocus = "项目管理和流程优化"
	default:
		recommendationFocus = "综合性推荐建议"
	}
	
	// 根据算法类型定制分析方法
	var algorithmApproach string
	switch req.Algorithm {
	case "hybrid":
		algorithmApproach = "结合多种方法进行综合分析"
	case "collaborative":
		algorithmApproach = "基于相似项目的协同经验"
	case "content_based":
		algorithmApproach = "基于内容特征的深度分析"
	case "popular":
		algorithmApproach = "基于热门实践的通用建议"
	default:
		algorithmApproach = "采用标准NESMA方法论"
	}
	
	return fmt.Sprintf(`
你是一位资深的NESMA功能点分析专家，请基于以下详细信息为用户提供专业的智能推荐：

## 项目基本信息
- **项目名称**: %s
- **业务领域**: %s
- **推荐重点**: %s
- **分析方法**: %s

## 当前需求描述
%s

## 项目功能点结构
%s

## 相关知识库内容
%s

## 专业分析任务
请作为NESMA专家，重点关注以下方面：

### 1. 功能点分析建议
- 基于项目的二级和三级功能点，识别EI/EO/EQ/ILF/EIF类型
- 分析每个功能点的DET（数据元素）和RET（记录元素）数量
- 提供功能点复杂度评估（简单/平均/复杂）
- 给出初步的功能点计数建议

### 2. 需求优化建议
- 针对现有功能点描述的改进建议
- 识别缺失的功能点或边界模糊的地方
- 提供更标准化的NESMA功能点描述

### 3. 最佳实践建议
- 基于相关知识库内容的实践建议
- 类似项目的成功经验分享
- NESMA标准在该领域的应用要点

### 4. 风险提示
- 功能点边界定义不清的风险
- 可能遗漏的功能点类型
- 复杂度评估可能存在的偏差

### 5. 功能类型建议
- 为核心功能推荐最合适的NESMA功能类型
- 解释推荐理由和判断依据

## 输出要求
请以纯JSON格式返回（不要使用markdown代码块），包含以下字段：
{
  "functionPointAdvice": "详细的功能点分析建议，包含具体的计数方法和复杂度评估",
  "optimizationSuggestions": ["具体的优化建议1", "具体的优化建议2", "具体的优化建议3"],
  "bestPractices": ["最佳实践1", "最佳实践2", "最佳实践3"],
  "riskWarnings": ["风险提示1", "风险提示2"],
  "functionType": "最推荐的功能类型(EI/EO/EQ/ILF/EIF)",
  "confidence": 0.85
}

## 注意事项
- 所有建议必须基于NESMA 2.2标准
- 考虑项目的实际功能点层级结构
- 置信度应基于知识库内容的相关性
- 建议要具体可操作，避免过于抽象
- 如果信息不足，请在建议中明确说明
`, req.ProjectName, req.Domain, recommendationFocus, algorithmApproach, 
   req.RequirementText, requirementsText, knowledgeText)
}

// formatRequirementsForPrompt 格式化项目需求为Prompt
func (s *AIRecommendationService) formatRequirementsForPrompt(requirements []ProjectRequirement) string {
	if len(requirements) == 0 {
		return "暂无详细功能点信息"
	}
	
	var formatted strings.Builder
	
	for _, req := range requirements {
		s.formatRequirementRecursive(&formatted, req, 0)
	}
	
	return formatted.String()
}

// formatRequirementRecursive 递归格式化需求结构
func (s *AIRecommendationService) formatRequirementRecursive(builder *strings.Builder, req ProjectRequirement, depth int) {
	indent := strings.Repeat("  ", depth)
	
	builder.WriteString(fmt.Sprintf("%s### L%d: %s\n", indent, req.Level, req.Title))
	builder.WriteString(fmt.Sprintf("%s编号: %s\n", indent, req.Code))
	if req.Description != "" {
		builder.WriteString(fmt.Sprintf("%s描述: %s\n", indent, req.Description))
	}
	builder.WriteString("\n")
	
	// 递归处理子级需求
	for _, child := range req.Children {
		s.formatRequirementRecursive(builder, child, depth+1)
	}
}

// formatKnowledgeForPrompt 格式化知识库内容为Prompt
func (s *AIRecommendationService) formatKnowledgeForPrompt(knowledge []nesma.NesmaKnowledgeEntry) string {
	if len(knowledge) == 0 {
		return "暂无相关知识库内容"
	}
	
	var formatted string
	for i, k := range knowledge {
		formatted += fmt.Sprintf("### 知识%d：%s\n", i+1, k.Title)
		formatted += fmt.Sprintf("分类：%s\n", k.Category)
		formatted += fmt.Sprintf("置信度：%.2f\n", k.ConfidenceScore)
		formatted += fmt.Sprintf("内容：%s\n\n", k.Content)
	}
	
	return formatted
}

// parseAIRecommendations 解析AI推荐结果
func (s *AIRecommendationService) parseAIRecommendations(aiResponse string, knowledge []nesma.NesmaKnowledgeEntry) ([]AIRecommendationItem, error) {
	// 处理AI响应中的markdown格式JSON
	cleanedResponse := aiResponse
	
	// 移除markdown代码块标记
	if strings.Contains(cleanedResponse, "```json") {
		cleanedResponse = strings.ReplaceAll(cleanedResponse, "```json", "")
		cleanedResponse = strings.ReplaceAll(cleanedResponse, "```", "")
	}
	
	// 提取JSON部分（去掉注释）
	lines := strings.Split(cleanedResponse, "\n")
	var jsonLines []string
	inJSON := false
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") {
			inJSON = true
		}
		if inJSON {
			jsonLines = append(jsonLines, line)
		}
		if strings.HasSuffix(line, "}") && inJSON {
			break
		}
	}
	
	jsonStr := strings.Join(jsonLines, "\n")
	
	var result RecommendationAnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		global.GVA_LOG.Error("JSON解析失败", zap.String("cleanedResponse", jsonStr), zap.Error(err))
		return nil, fmt.Errorf("解析AI响应失败: %w", err)
	}

	// 构建知识库引用
	knowledgeRefs := make([]KnowledgeReference, 0, len(knowledge))
	for _, k := range knowledge {
		knowledgeRefs = append(knowledgeRefs, KnowledgeReference{
			ID:        k.ID,
			Title:     k.Title,
			Category:  k.Category,
			Relevance: k.ConfidenceScore,
		})
	}

	// 创建推荐项目
	recommendations := []AIRecommendationItem{
		{
			ID:              fmt.Sprintf("ai_rec_%d", time.Now().UnixNano()),
			Title:           "AI智能推荐 - 功能点分析建议",
			Content:         result.FunctionPointAdvice,
			Category:        "function_analysis",
			Confidence:      result.Confidence,
			KnowledgeRefs:   knowledgeRefs,
			Improvements:    result.OptimizationSuggestions,
			FunctionType:    result.FunctionType,
			RiskWarnings:    result.RiskWarnings,
			BestPractices:   result.BestPractices,
			CreatedAt:       time.Now(),
		},
	}

	return recommendations, nil
}

// calculateAverageConfidence 计算平均置信度
func (s *AIRecommendationService) calculateAverageConfidence(recommendations []AIRecommendationItem) float64 {
	if len(recommendations) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, rec := range recommendations {
		total += rec.Confidence
	}
	
	return total / float64(len(recommendations))
}

// AdoptRecommendation 采纳推荐
func (s *AIRecommendationService) AdoptRecommendation(ctx context.Context, userID uint, projectID uint, recommendation *AIRecommendationItem, feedback string) error {
	// 1. 记录用户采纳行为
	knowledgeRefsJSON, _ := json.Marshal(recommendation.KnowledgeRefs)
	
	adoption := &RecommendationAdoption{
		UserID:          userID,
		ProjectID:       projectID,
		RecommendationID: recommendation.ID,
		Title:           recommendation.Title,
		Content:         recommendation.Content,
		Category:        recommendation.Category,
		Confidence:      recommendation.Confidence,
		KnowledgeRefs:   datatypes.JSON(knowledgeRefsJSON),
		UserFeedback:    feedback,
		AdoptedAt:       time.Now(),
		CreatedAt:       time.Now(),
	}
	
	if err := s.db.Create(adoption).Error; err != nil {
		return fmt.Errorf("记录采纳失败: %w", err)
	}
	
	// 获取用户信息以设置正确的作者
	var user struct {
		Username string `json:"username"`
		NickName string `json:"nickName"`
	}
	
	if err := s.db.Table("sys_users").Select("username, nick_name").Where("id = ?", userID).First(&user).Error; err != nil {
		global.GVA_LOG.Warn("获取用户信息失败，使用默认作者", zap.Error(err))
		user.Username = fmt.Sprintf("用户%d", userID)
		user.NickName = fmt.Sprintf("用户%d", userID)
	}
	
	// 优先使用昵称，如果昵称为空或默认值，则使用用户名
	authorName := user.NickName
	if authorName == "" || authorName == "系统用户" {
		authorName = user.Username
	}
	
	// 2. 将采纳的推荐内容入库到知识库
	knowledgeEntry := &nesma.NesmaKnowledgeEntry{
		Title:           fmt.Sprintf("用户采纳推荐: %s", recommendation.Title),
		Content:         recommendation.Content,
		Category:        "用户采纳",                    // 中文分类
		Domain:          getCategoryDisplayName(recommendation.Category), // 转换为中文域名
		Source:          "AI智能推荐",                   // 中文来源
		Author:          authorName,                   // 使用真实用户名
		ConfidenceScore: recommendation.Confidence,
		Status:          "active",
	}
	
	if err := s.db.Create(knowledgeEntry).Error; err != nil {
		global.GVA_LOG.Error("推荐内容入库失败", zap.Error(err))
		return fmt.Errorf("推荐内容入库失败: %w", err)
	}
	
	global.GVA_LOG.Info("用户采纳推荐成功", 
		zap.Uint("userId", userID),
		zap.Uint("projectId", projectID),
		zap.String("recommendationId", recommendation.ID),
		zap.String("feedback", feedback))
	
	return nil
}

// RejectRecommendation 拒绝推荐
func (s *AIRecommendationService) RejectRecommendation(ctx context.Context, userID uint, projectID uint, recommendationID string, reason string) error {
	// 记录拒绝行为用于改进推荐算法
	global.GVA_LOG.Info("用户拒绝推荐", 
		zap.Uint("userId", userID),
		zap.Uint("projectId", projectID),
		zap.String("recommendationId", recommendationID),
		zap.String("reason", reason))
	
	return nil
}

// GetAdoptionHistory 获取采纳历史
func (s *AIRecommendationService) GetAdoptionHistory(ctx context.Context, userID uint, projectID uint) ([]RecommendationAdoption, error) {
	var adoptions []RecommendationAdoption
	query := s.db.Where("user_id = ?", userID)
	
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	
	err := query.Order("adopted_at DESC").Find(&adoptions).Error
	return adoptions, err
}

// 全局AI推荐服务实例
var GlobalAIRecommendationService *AIRecommendationService

// InitAIRecommendationService 初始化AI推荐服务
func InitAIRecommendationService() {
	GlobalAIRecommendationService = NewAIRecommendationService()
	
	// 自动迁移数据库表
	err := global.GVA_DB.AutoMigrate(&RecommendationAdoption{})
	if err != nil {
		global.GVA_LOG.Error("Failed to migrate AI recommendation tables", zap.Error(err))
	} else {
		global.GVA_LOG.Info("AI recommendation service initialized successfully")
	}
}

// getCategoryDisplayName 将英文分类转换为中文显示名称
func getCategoryDisplayName(category string) string {
	categoryMap := map[string]string{
		"function_analysis":      "功能点分析",
		"requirement":            "需求管理", 
		"template":               "模板管理",
		"agent":                  "AI助手",
		"project":                "项目管理",
		"knowledge":              "知识管理",
		"ai_generated":           "AI生成",
		"best_practice":          "最佳实践",
		"risk_warning":           "风险提示",
		"process_improvement":    "流程改进",
		"nesma_standard":         "NESMA标准",
		"domain_knowledge":       "领域知识",
		"user_feedback":          "用户反馈",
	}
	
	if displayName, exists := categoryMap[category]; exists {
		return displayName
	}
	
	// 如果没有找到对应的中文名称，返回原始分类名
	return category
}

// GetAIRecommendationService 获取AI推荐服务
func GetAIRecommendationService() *AIRecommendationService {
	if GlobalAIRecommendationService == nil {
		InitAIRecommendationService()
	}
	return GlobalAIRecommendationService
}