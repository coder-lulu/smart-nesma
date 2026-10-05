package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RecommendationService 推荐服务
type RecommendationService struct {
	db *gorm.DB
}

// RecommendationType 推荐类型
type RecommendationType string

const (
	RecommendationTypeRequirement RecommendationType = "requirement"
	RecommendationTypeTemplate    RecommendationType = "template"
	RecommendationTypeAgent       RecommendationType = "agent"
	RecommendationTypeProject     RecommendationType = "project"
	RecommendationTypeKnowledge   RecommendationType = "knowledge"
)

// RecommendationAlgorithm 推荐算法
type RecommendationAlgorithm string

const (
	AlgorithmCollaborative RecommendationAlgorithm = "collaborative"
	AlgorithmContentBased  RecommendationAlgorithm = "content_based"
	AlgorithmHybrid        RecommendationAlgorithm = "hybrid"
	AlgorithmPopular       RecommendationAlgorithm = "popular"
)

// UserBehavior 用户行为记录
type UserBehavior struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
	UserID     uint           `json:"userId" gorm:"not null;index"`
	Action     string         `json:"action" gorm:"type:varchar(50);not null"` // view, like, share, download, create, edit, delete
	ItemType   string         `json:"itemType" gorm:"type:varchar(50);not null"`
	ItemID     uint           `json:"itemId" gorm:"not null"`
	SessionID  string         `json:"sessionId" gorm:"type:varchar(255)"`
	Context    datatypes.JSON `json:"context" gorm:"type:json"`
	Duration   int            `json:"duration" gorm:"default:0"` // 停留时间(秒)
	DeviceType string         `json:"deviceType" gorm:"type:varchar(50)"`
	UserAgent  string         `json:"userAgent" gorm:"type:varchar(500)"`
	IPAddress  string         `json:"ipAddress" gorm:"type:varchar(50)"`
	CreatedAt  time.Time      `json:"createdAt"`
}

// TableName 自定义表名
func (UserBehavior) TableName() string {
	return "user_behaviors"
}

// RecommendationHistory 推荐历史
type RecommendationHistory struct {
	ID                 uint           `json:"id" gorm:"primaryKey"`
	UserID             uint           `json:"userId" gorm:"not null;index"`
	RecommendationType string         `json:"recommendationType" gorm:"type:varchar(50);not null"`
	ItemType           string         `json:"itemType" gorm:"type:varchar(50);not null"`
	ItemID             uint           `json:"itemId" gorm:"not null"`
	Score              float64        `json:"score" gorm:"not null"`
	Reason             string         `json:"reason" gorm:"type:text"`
	Algorithm          string         `json:"algorithm" gorm:"type:varchar(50)"`
	Context            datatypes.JSON `json:"context" gorm:"type:json"`
	IsClicked          bool           `json:"isClicked" gorm:"default:false"`
	IsInteracted       bool           `json:"isInteracted" gorm:"default:false"`
	FeedbackScore      *float64       `json:"feedbackScore" gorm:"default:null"`
	FeedbackComment    string         `json:"feedbackComment" gorm:"type:text"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

// TableName 自定义表名
func (RecommendationHistory) TableName() string {
	return "recommendation_histories"
}

// RecommendationRequest 推荐请求
type RecommendationRequest struct {
	UserID             uint                   `json:"userId" binding:"required"`
	RecommendationType string                 `json:"recommendationType"` // requirement, template, agent, project, knowledge, all
	Algorithm          string                 `json:"algorithm"`          // collaborative, content_based, hybrid, popular
	Limit              int                    `json:"limit"`              // 推荐数量限制
	Context            map[string]interface{} `json:"context"`            // 上下文信息
	Filters            map[string]interface{} `json:"filters"`            // 过滤条件
	ExcludeItems       []uint                 `json:"excludeItems"`       // 排除的项目ID
	IncludeHistory     bool                   `json:"includeHistory"`     // 是否包含历史记录
	Diversify          bool                   `json:"diversify"`          // 是否多样化
	MinScore           float64                `json:"minScore"`           // 最低分数阈值
}

// RecommendationItem 推荐项目
type RecommendationItem struct {
	Type        string                 `json:"type"`
	ItemID      uint                   `json:"itemId"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Score       float64                `json:"score"`
	Reason      string                 `json:"reason"`
	Category    string                 `json:"category"`
	Tags        []string               `json:"tags"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

// RecommendationResponse 推荐响应
type RecommendationResponse struct {
	Items     []RecommendationItem   `json:"items"`
	Total     int                    `json:"total"`
	Algorithm string                 `json:"algorithm"`
	Context   map[string]interface{} `json:"context"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewRecommendationService 创建推荐服务
func NewRecommendationService() *RecommendationService {
	return &RecommendationService{
		db: global.GVA_DB,
	}
}

// GetRecommendations 获取推荐
func (rs *RecommendationService) GetRecommendations(ctx context.Context, req RecommendationRequest) (*RecommendationResponse, error) {
	// 参数验证
	if req.UserID == 0 {
		return nil, errors.New("用户ID不能为空")
	}

	// 设置默认值
	if req.Limit <= 0 {
		req.Limit = 10
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	if req.Algorithm == "" {
		req.Algorithm = string(AlgorithmHybrid)
	}

	if req.RecommendationType == "" {
		req.RecommendationType = "all"
	}

	if req.MinScore == 0 {
		req.MinScore = 0.1
	}

	// 获取推荐结果
	var recommendations []RecommendationItem
	var err error

	switch RecommendationAlgorithm(req.Algorithm) {
	case AlgorithmCollaborative:
		recommendations, err = rs.getCollaborativeRecommendations(ctx, req)
	case AlgorithmContentBased:
		recommendations, err = rs.getContentBasedRecommendations(ctx, req)
	case AlgorithmHybrid:
		recommendations, err = rs.getHybridRecommendations(ctx, req)
	case AlgorithmPopular:
		recommendations, err = rs.getPopularRecommendations(ctx, req)
	default:
		return nil, fmt.Errorf("不支持的推荐算法: %s", req.Algorithm)
	}

	if err != nil {
		return nil, err
	}

	// 过滤和排序
	recommendations = rs.filterAndSortRecommendations(recommendations, req)

	// 多样化
	if req.Diversify {
		recommendations = rs.diversifyRecommendations(recommendations, req.Limit)
	}

	// 限制数量
	if len(recommendations) > req.Limit {
		recommendations = recommendations[:req.Limit]
	}

	// 记录推荐历史
	rs.saveRecommendationHistory(req.UserID, recommendations, req.Algorithm, req.Context)

	return &RecommendationResponse{
		Items:     recommendations,
		Total:     len(recommendations),
		Algorithm: req.Algorithm,
		Context:   req.Context,
		Timestamp: time.Now(),
	}, nil
}

// getHybridRecommendations 获取混合推荐
func (rs *RecommendationService) getHybridRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var allRecommendations []RecommendationItem

	// 获取不同类型的推荐
	if req.RecommendationType == "all" || req.RecommendationType == "requirement" {
		reqRecommendations, _ := rs.getRequirementRecommendations(ctx, req)
		allRecommendations = append(allRecommendations, reqRecommendations...)
	}

	if req.RecommendationType == "all" || req.RecommendationType == "template" {
		templateRecommendations, _ := rs.getTemplateRecommendations(ctx, req)
		allRecommendations = append(allRecommendations, templateRecommendations...)
	}

	if req.RecommendationType == "all" || req.RecommendationType == "agent" {
		agentRecommendations, _ := rs.getAgentRecommendations(ctx, req)
		allRecommendations = append(allRecommendations, agentRecommendations...)
	}

	if req.RecommendationType == "all" || req.RecommendationType == "project" {
		projectRecommendations, _ := rs.getProjectRecommendations(ctx, req)
		allRecommendations = append(allRecommendations, projectRecommendations...)
	}

	if req.RecommendationType == "all" || req.RecommendationType == "knowledge" {
		knowledgeRecommendations, _ := rs.getKnowledgeRecommendations(ctx, req)
		allRecommendations = append(allRecommendations, knowledgeRecommendations...)
	}

	// 合并和重新评分
	recommendations := rs.mergeAndRescoreRecommendations(allRecommendations, req)

	return recommendations, nil
}

// getAgentRecommendations 获取Agent推荐
func (rs *RecommendationService) getAgentRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var items []RecommendationItem
	var agents []nesma.NesmaAgent

	// 获取活跃的Agent
	rs.db.Where("status = ?", "active").Find(&agents)

	userBehaviors := rs.getUserBehaviors(req.UserID, "agent", 30)

	for _, agent := range agents {
		score := rs.calculateAgentScore(agent, userBehaviors, req.Context)
		if score > req.MinScore {
			reason := rs.generateAgentReason(agent, userBehaviors)

			// 处理 capabilities JSON
			var capabilities []string
			if agent.Capabilities != nil {
				capabilityData, _ := agent.Capabilities.MarshalJSON()
				json.Unmarshal(capabilityData, &capabilities)
			}

			items = append(items, RecommendationItem{
				Type:        string(RecommendationTypeAgent),
				ItemID:      agent.ID,
				Title:       agent.Name,
				Description: agent.Description,
				Score:       score,
				Reason:      reason,
				Category:    agent.AgentType,
				Tags:        capabilities,
				Metadata: map[string]interface{}{
					"agentType":           agent.AgentType,
					"status":              agent.Status,
					"version":             agent.Version,
					"maxConcurrency":      agent.MaxConcurrency,
					"averageResponseTime": agent.AverageResponseTime,
				},
				CreatedAt: agent.CreatedAt,
				UpdatedAt: agent.UpdatedAt,
			})
		}
	}

	return items, nil
}

// getKnowledgeRecommendations 获取知识推荐
func (rs *RecommendationService) getKnowledgeRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var items []RecommendationItem
	var knowledgeEntries []nesma.NesmaKnowledgeEntry

	// 获取活跃的知识条目
	rs.db.Where("status = ?", "active").Find(&knowledgeEntries)

	userBehaviors := rs.getUserBehaviors(req.UserID, "knowledge", 30)

	for _, knowledge := range knowledgeEntries {
		score := rs.calculateKnowledgeScore(knowledge, userBehaviors, req.Context)
		if score > req.MinScore {
			reason := rs.generateKnowledgeReason(knowledge, userBehaviors)

			// 处理 tags JSON
			var tags []string
			if knowledge.Tags != nil {
				tagData, _ := knowledge.Tags.MarshalJSON()
				json.Unmarshal(tagData, &tags)
			}

			items = append(items, RecommendationItem{
				Type:        string(RecommendationTypeKnowledge),
				ItemID:      knowledge.ID,
				Title:       knowledge.Title,
				Description: knowledge.Content,
				Score:       score,
				Reason:      reason,
				Category:    knowledge.Category,
				Tags:        tags,
				Metadata: map[string]interface{}{
					"category":        knowledge.Category,
					"domain":          knowledge.Domain,
					"confidenceScore": knowledge.ConfidenceScore,
					"usageCount":      knowledge.UsageCount,
					"version":         knowledge.Version,
					"source":          knowledge.Source,
					"author":          knowledge.Author,
				},
				CreatedAt: knowledge.CreatedAt,
				UpdatedAt: knowledge.UpdatedAt,
			})
		}
	}

	return items, nil
}

// calculateAgentScore 计算Agent分数
func (rs *RecommendationService) calculateAgentScore(agent nesma.NesmaAgent, behaviors []nesma.UserBehavior, context map[string]interface{}) float64 {
	score := 0.4 // 基础分数

	// 根据Agent类型匹配
	if agentType, exists := context["agentType"]; exists {
		if agentType == agent.AgentType {
			score += 0.3
		}
	}

	// 根据用户行为
	for _, behavior := range behaviors {
		if behavior.ItemID == agent.ID {
			if behavior.Action == "view" {
				score += 0.05
			} else if behavior.Action == "use" {
				score += 0.1
			}
		}
	}

	// 根据Agent状态
	if agent.Status == "active" {
		score += 0.1
	}

	// 根据能力匹配
	if capabilities, exists := context["capabilities"]; exists {
		if capStr, ok := capabilities.(string); ok {
			// 处理 capabilities JSON
			var agentCapabilities []string
			if agent.Capabilities != nil {
				capabilityData, _ := agent.Capabilities.MarshalJSON()
				json.Unmarshal(capabilityData, &agentCapabilities)

				for _, capability := range agentCapabilities {
					if strings.Contains(capStr, capability) {
						score += 0.15
					}
				}
			}
		}
	}

	return math.Min(score, 1.0)
}

// calculateKnowledgeScore 计算知识库分数
func (rs *RecommendationService) calculateKnowledgeScore(knowledge nesma.NesmaKnowledgeEntry, behaviors []nesma.UserBehavior, context map[string]interface{}) float64 {
	score := 0.4 // 基础分数

	// 根据分类匹配
	if category, exists := context["category"]; exists {
		if category == knowledge.Category {
			score += 0.3
		}
	}

	// 根据标签匹配
	if tags, exists := context["tags"]; exists {
		if tagStr, ok := tags.(string); ok && knowledge.Tags != nil {
			// 处理 tags JSON
			var knowledgeTags []string
			tagData, _ := knowledge.Tags.MarshalJSON()
			json.Unmarshal(tagData, &knowledgeTags)

			for _, tag := range knowledgeTags {
				if strings.Contains(tagStr, tag) {
					score += 0.1
				}
			}
		}
	}

	// 根据置信度
	score += knowledge.ConfidenceScore * 0.2

	// 根据使用次数
	if knowledge.UsageCount > 0 {
		score += math.Min(float64(knowledge.UsageCount)*0.01, 0.2)
	}

	return math.Min(score, 1.0)
}

// generateKnowledgeReason 生成知识库推荐理由
func (rs *RecommendationService) generateKnowledgeReason(knowledge nesma.NesmaKnowledgeEntry, behaviors []nesma.UserBehavior) string {
	reasons := []string{}

	// 根据分类
	if knowledge.Category == "NESMA_STANDARD" {
		reasons = append(reasons, "NESMA标准知识")
	} else if knowledge.Category == "BEST_PRACTICE" {
		reasons = append(reasons, "最佳实践")
	}

	// 根据置信度
	if knowledge.ConfidenceScore > 0.8 {
		reasons = append(reasons, "高置信度内容")
	}

	// 根据使用次数
	if knowledge.UsageCount > 10 {
		reasons = append(reasons, "热门内容")
	}

	if len(reasons) == 0 {
		return "相关知识推荐"
	}

	return strings.Join(reasons, "、")
}

// getRequirementRecommendations 获取需求推荐
func (rs *RecommendationService) getRequirementRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var items []RecommendationItem

	// 1. 基于协同过滤的推荐
	collaborativeItems := rs.getCollaborativeFilteringRecommendations(req.UserID, "requirement")

	// 2. 基于内容的推荐
	contentItems := rs.getContentBasedRecommendationsOld(req.UserID, "requirement", req.Context)

	// 3. 基于热门度的推荐
	popularItems := rs.getPopularityBasedRecommendations("requirement")

	// 合并推荐结果
	items = rs.mergeRecommendations(collaborativeItems, contentItems, popularItems)

	return items, nil
}

// getTemplateRecommendations 获取模板推荐
func (rs *RecommendationService) getTemplateRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var items []RecommendationItem
	var templates []nesma.NesmaDocTemplate

	// 获取用户行为历史
	userBehaviors := rs.getUserBehaviors(req.UserID, "template", 50)

	// 查询模板
	query := rs.db.Model(&nesma.NesmaDocTemplate{}).Where("is_active = ?", true)

	// 如果有项目上下文，优先推荐项目相关模板
	if projectID, exists := req.Context["projectId"]; exists {
		if pid, ok := projectID.(uint); ok {
			query = query.Where("project_id = ? OR project_id IS NULL", pid)
		}
	}

	query.Find(&templates)

	for _, template := range templates {
		score := rs.calculateTemplateScore(template, userBehaviors, req.Context)

		reason := rs.generateTemplateReason(template, userBehaviors)

		items = append(items, RecommendationItem{
			ItemID:      template.ID,
			Type:        string(RecommendationTypeTemplate),
			Title:       template.Name,
			Description: template.Description,
			Score:       score,
			Reason:      reason,
			Category:    template.Type,
			Tags:        []string{template.Format},
			Metadata: map[string]interface{}{
				"format":   template.Format,
				"type":     template.Type,
				"isActive": template.IsActive,
			},
			CreatedAt: template.CreatedAt,
			UpdatedAt: template.UpdatedAt,
		})
	}

	return items, nil
}

// getProjectRecommendations 获取项目推荐
func (rs *RecommendationService) getProjectRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var items []RecommendationItem
	var projects []nesma.NesmaProject

	// 获取用户可访问的项目
	rs.db.Model(&nesma.NesmaProject{}).Where("status = ?", "active").Find(&projects)

	userBehaviors := rs.getUserBehaviors(req.UserID, "project", 20)

	for _, project := range projects {
		score := rs.calculateProjectScore(project, userBehaviors, req.Context)

		reason := rs.generateProjectReason(project, userBehaviors)

		items = append(items, RecommendationItem{
			ItemID:      project.ID,
			Type:        string(RecommendationTypeProject),
			Title:       project.Name,
			Description: project.Description,
			Score:       score,
			Reason:      reason,
			Category:    project.Domain,
			Tags:        []string{project.Domain, project.Status},
			Metadata: map[string]interface{}{
				"domain":  project.Domain,
				"status":  project.Status,
				"ownerId": project.OwnerID,
			},
			CreatedAt: project.CreatedAt,
			UpdatedAt: project.UpdatedAt,
		})
	}

	return items, nil
}

// getCollaborativeFilteringRecommendations 协同过滤推荐
func (rs *RecommendationService) getCollaborativeFilteringRecommendations(userID uint, targetType string) []RecommendationItem {
	// 简化的协同过滤实现
	// 1. 找到相似用户
	// 2. 推荐相似用户喜欢的项目

	var items []RecommendationItem
	// 这里可以实现更复杂的协同过滤算法
	return items
}

// getContentBasedRecommendationsOld 基于内容的推荐（旧版本）
func (rs *RecommendationService) getContentBasedRecommendationsOld(userID uint, targetType string, context map[string]interface{}) []RecommendationItem {
	var items []RecommendationItem
	// 基于用户历史偏好和内容特征进行推荐
	return items
}

// getPopularityBasedRecommendations 基于热门度的推荐
func (rs *RecommendationService) getPopularityBasedRecommendations(targetType string) []RecommendationItem {
	var items []RecommendationItem
	// 推荐热门项目
	return items
}

// mergeRecommendations 合并推荐结果
func (rs *RecommendationService) mergeRecommendations(lists ...[]RecommendationItem) []RecommendationItem {
	merged := make([]RecommendationItem, 0)
	seen := make(map[string]bool)

	for _, list := range lists {
		for _, item := range list {
			key := fmt.Sprintf("%s_%d", item.Type, item.ItemID)
			if !seen[key] {
				merged = append(merged, item)
				seen[key] = true
			}
		}
	}

	return merged
}

// calculateTemplateScore 计算模板分数
func (rs *RecommendationService) calculateTemplateScore(template nesma.NesmaDocTemplate, behaviors []nesma.UserBehavior, context map[string]interface{}) float64 {
	score := 0.5 // 基础分数

	// 基于用户历史行为
	for _, behavior := range behaviors {
		if behavior.ItemID == template.ID {
			switch behavior.Action {
			case "view":
				score += 0.1
			case "use":
				score += 0.3
			case "download":
				score += 0.2
			case "favorite":
				score += 0.4
			}
		}
	}

	// 基于模板流行度（使用次数）
	var usageCount int64
	rs.db.Model(&nesma.UserBehavior{}).Where("item_type = ? AND item_id = ? AND action IN ?",
		"template", template.ID, []string{"use", "download"}).Count(&usageCount)
	popularityScore := math.Log(float64(usageCount+1)) * 0.1
	score += popularityScore

	// 基于上下文匹配
	if context != nil {
		if projectType, exists := context["projectType"]; exists && projectType == template.Type {
			score += 0.3
		}
		if format, exists := context["preferredFormat"]; exists && format == template.Format {
			score += 0.2
		}
	}

	// 归一化分数
	return math.Min(score, 1.0)
}

// calculateProjectScore 计算项目分数
func (rs *RecommendationService) calculateProjectScore(project nesma.NesmaProject, behaviors []nesma.UserBehavior, context map[string]interface{}) float64 {
	score := 0.3 // 基础分数

	// 基于用户行为
	for _, behavior := range behaviors {
		if behavior.ItemID == project.ID {
			switch behavior.Action {
			case "view":
				score += 0.1
			case "edit":
				score += 0.3
			case "collaborate":
				score += 0.4
			}
		}
	}

	// 基于项目活跃度
	var recentActivity int64
	rs.db.Model(&nesma.UserBehavior{}).Where("item_type = ? AND item_id = ? AND created_at > ?",
		"project", project.ID, time.Now().AddDate(0, 0, -30)).Count(&recentActivity)

	activityScore := math.Min(float64(recentActivity)*0.05, 0.3)
	score += activityScore

	// 基于领域匹配
	if context != nil {
		if domain, exists := context["domain"]; exists && domain == project.Domain {
			score += 0.3
		}
	}

	return math.Min(score, 1.0)
}

// generateTemplateReason 生成模板推荐理由
func (rs *RecommendationService) generateTemplateReason(template nesma.NesmaDocTemplate, behaviors []nesma.UserBehavior) string {
	reasons := []string{}

	for _, behavior := range behaviors {
		if behavior.ItemID == template.ID {
			switch behavior.Action {
			case "use":
				reasons = append(reasons, "您使用过类似模板")
			case "download":
				reasons = append(reasons, "您下载过此模板")
			case "favorite":
				reasons = append(reasons, "您收藏过类似模板")
			}
		}
	}

	if template.Type == "STANDARD" {
		reasons = append(reasons, "标准模板推荐")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "基于模板类型匹配")
	}

	return strings.Join(reasons, "，")
}

// generateProjectReason 生成项目推荐理由
func (rs *RecommendationService) generateProjectReason(project nesma.NesmaProject, behaviors []nesma.UserBehavior) string {
	reasons := []string{}

	for _, behavior := range behaviors {
		if behavior.ItemID == project.ID {
			switch behavior.Action {
			case "collaborate":
				reasons = append(reasons, "您参与过类似项目")
			case "view":
				reasons = append(reasons, "您关注过此项目")
			}
		}
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "项目领域匹配您的兴趣")
	}

	return strings.Join(reasons, "，")
}

// getUserBehaviors 获取用户行为历史
func (rs *RecommendationService) getUserBehaviors(userID uint, targetType string, limit int) []nesma.UserBehavior {
	var behaviors []nesma.UserBehavior
	rs.db.Where("user_id = ? AND item_type = ?", userID, targetType).
		Order("created_at DESC").
		Limit(limit).
		Find(&behaviors)
	return behaviors
}

// saveRecommendationHistory 记录推荐历史
func (rs *RecommendationService) saveRecommendationHistory(userID uint, recommendations []RecommendationItem, algorithm string, context map[string]interface{}) {
	for _, item := range recommendations {
		var contextData datatypes.JSON
		if context != nil {
			if data, err := json.Marshal(context); err == nil {
				contextData = datatypes.JSON(data)
			}
		}

		history := RecommendationHistory{
			UserID:             userID,
			RecommendationType: string(RecommendationTypeRequirement),
			ItemType:           string(RecommendationTypeRequirement),
			ItemID:             item.ItemID,
			Score:              item.Score,
			Reason:             item.Reason,
			Algorithm:          algorithm,
			Context:            contextData,
			IsClicked:          false,
			IsInteracted:       false,
			CreatedAt:          time.Now(),
		}
		rs.db.Create(&history)
	}
}

// filterAndSortRecommendations 过滤和排序推荐结果
func (rs *RecommendationService) filterAndSortRecommendations(recommendations []RecommendationItem, req RecommendationRequest) []RecommendationItem {
	// 过滤低分项
	filteredItems := make([]RecommendationItem, 0)
	for _, item := range recommendations {
		if item.Score >= req.MinScore {
			filteredItems = append(filteredItems, item)
		}
	}

	// 排序并限制结果数量
	sort.Slice(filteredItems, func(i, j int) bool {
		return filteredItems[i].Score > filteredItems[j].Score
	})

	if len(filteredItems) > req.Limit {
		filteredItems = filteredItems[:req.Limit]
	}

	return filteredItems
}

// diversifyRecommendations 多样化推荐结果
func (rs *RecommendationService) diversifyRecommendations(recommendations []RecommendationItem, limit int) []RecommendationItem {
	// 实现多样化推荐逻辑
	return recommendations
}

// generateAgentReason 生成Agent推荐理由
func (rs *RecommendationService) generateAgentReason(agent nesma.NesmaAgent, behaviors []nesma.UserBehavior) string {
	reasons := []string{}

	for _, behavior := range behaviors {
		if behavior.ItemID == agent.ID {
			switch behavior.Action {
			case "use":
				reasons = append(reasons, "您使用过类似Agent")
			case "configure":
				reasons = append(reasons, "您配置过此Agent")
			}
		}
	}

	if agent.AgentType == "REQUIREMENT_ANALYSIS" {
		reasons = append(reasons, "需求分析专用Agent")
	} else if agent.AgentType == "NESMA_EVALUATION" {
		reasons = append(reasons, "NESMA评估专用Agent")
	}

	if len(reasons) == 0 {
		reasons = append(reasons, "基于Agent能力匹配")
	}

	return strings.Join(reasons, "，")
}

// 全局推荐服务实例
var GlobalRecommendationService *RecommendationService

// InitRecommendationService 初始化推荐服务
func InitRecommendationService() {
	GlobalRecommendationService = NewRecommendationService()

	// 自动迁移数据库表
	err := global.GVA_DB.AutoMigrate(&UserBehavior{}, &RecommendationHistory{})
	if err != nil {
		global.GVA_LOG.Error("Failed to migrate recommendation tables", zap.Error(err))
	} else {
		global.GVA_LOG.Info("Recommendation service initialized successfully")
	}
}

// GetRecommendationService 获取推荐服务
func GetRecommendationService() *RecommendationService {
	if GlobalRecommendationService == nil {
		InitRecommendationService()
	}
	return GlobalRecommendationService
}

// getCollaborativeRecommendations 获取协同过滤推荐
func (rs *RecommendationService) getCollaborativeRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var recommendations []RecommendationItem
	// 简化的协同过滤实现
	// 1. 找到相似用户
	// 2. 推荐相似用户喜欢的项目
	return recommendations, nil
}

// getContentBasedRecommendations 获取基于内容的推荐
func (rs *RecommendationService) getContentBasedRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var recommendations []RecommendationItem
	// 基于用户历史偏好和内容特征进行推荐
	return recommendations, nil
}

// getPopularRecommendations 获取热门推荐
func (rs *RecommendationService) getPopularRecommendations(ctx context.Context, req RecommendationRequest) ([]RecommendationItem, error) {
	var recommendations []RecommendationItem
	// 推荐热门项目
	return recommendations, nil
}

// mergeAndRescoreRecommendations 合并并重新评分推荐
func (rs *RecommendationService) mergeAndRescoreRecommendations(recommendations []RecommendationItem, req RecommendationRequest) []RecommendationItem {
	// 去重
	seen := make(map[string]bool)
	merged := make([]RecommendationItem, 0)

	for _, item := range recommendations {
		key := fmt.Sprintf("%s_%d", item.Type, item.ItemID)
		if !seen[key] {
			merged = append(merged, item)
			seen[key] = true
		}
	}

	// 重新评分
	for i := range merged {
		merged[i].Score = merged[i].Score * 0.8 // 混合算法权重调整
	}

	return merged
}

// RecordUserBehavior 记录用户行为
func (rs *RecommendationService) RecordUserBehavior(userID uint, action, itemType string, itemID uint, context map[string]interface{}) error {
	var contextData datatypes.JSON
	if context != nil {
		if data, err := json.Marshal(context); err == nil {
			contextData = datatypes.JSON(data)
		}
	}

	behavior := nesma.UserBehavior{
		UserID:   userID,
		Action:   action,
		ItemType: itemType,
		ItemID:   itemID,
		Context:  contextData,
	}

	return rs.db.Create(&behavior).Error
}

// RecordRecommendationFeedback 记录推荐反馈
func (rs *RecommendationService) RecordRecommendationFeedback(userID uint, recommendationID uint, itemType string, itemID uint, clicked bool, liked *bool) error {
	// 查找推荐历史记录
	var history RecommendationHistory
	if err := rs.db.Where("user_id = ? AND item_type = ? AND item_id = ?", userID, itemType, itemID).First(&history).Error; err != nil {
		// 如果找不到历史记录，创建一个新的
		history = RecommendationHistory{
			UserID:             userID,
			RecommendationType: itemType,
			ItemType:           itemType,
			ItemID:             itemID,
			Score:              0.5,
			Reason:             "用户反馈",
			Algorithm:          "feedback",
			IsClicked:          clicked,
			IsInteracted:       true,
			CreatedAt:          time.Now(),
		}
	}

	// 更新反馈信息
	history.IsClicked = clicked
	history.IsInteracted = true
	if liked != nil {
		if *liked {
			history.FeedbackScore = &[]float64{5.0}[0]
		} else {
			history.FeedbackScore = &[]float64{1.0}[0]
		}
	}

	return rs.db.Save(&history).Error
}

// GetRecommendationStats 获取推荐统计
func (rs *RecommendationService) GetRecommendationStats(userID uint, days int) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// 统计时间范围
	since := time.Now().AddDate(0, 0, -days)

	// 推荐总数
	var totalRecommendations int64
	rs.db.Model(&RecommendationHistory{}).Where("user_id = ? AND created_at > ?", userID, since).Count(&totalRecommendations)

	// 点击率
	var clickedRecommendations int64
	rs.db.Model(&RecommendationHistory{}).Where("user_id = ? AND created_at > ? AND is_clicked = ?", userID, since, true).Count(&clickedRecommendations)

	// 互动率
	var interactedRecommendations int64
	rs.db.Model(&RecommendationHistory{}).Where("user_id = ? AND created_at > ? AND is_interacted = ?", userID, since, true).Count(&interactedRecommendations)

	// 计算比率
	var clickRate, interactionRate float64
	if totalRecommendations > 0 {
		clickRate = float64(clickedRecommendations) / float64(totalRecommendations)
		interactionRate = float64(interactedRecommendations) / float64(totalRecommendations)
	}

	// 按类型统计
	var typeStats []struct {
		ItemType string
		Count    int64
	}
	rs.db.Model(&RecommendationHistory{}).Select("item_type, count(*) as count").Where("user_id = ? AND created_at > ?", userID, since).Group("item_type").Scan(&typeStats)

	// 用户行为统计
	var behaviorStats []struct {
		Action string
		Count  int64
	}
	rs.db.Model(&nesma.UserBehavior{}).Select("action, count(*) as count").Where("user_id = ? AND created_at > ?", userID, since).Group("action").Scan(&behaviorStats)

	stats["totalRecommendations"] = totalRecommendations
	stats["clickedRecommendations"] = clickedRecommendations
	stats["interactedRecommendations"] = interactedRecommendations
	stats["clickRate"] = clickRate
	stats["interactionRate"] = interactionRate
	stats["typeStats"] = typeStats
	stats["behaviorStats"] = behaviorStats
	stats["period"] = days

	return stats, nil
}
