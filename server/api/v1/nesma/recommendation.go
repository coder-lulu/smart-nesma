package nesma

import (
	"context"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type RecommendationApi struct{}

// @Tags 智能推荐
// @Summary 获取推荐列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param recommendationType query string false "推荐类型" Enums(all,requirement,template,agent,project,knowledge) default(all)
// @Param algorithm query string false "推荐算法" Enums(hybrid,collaborative,content_based,popular) default(hybrid)
// @Param limit query int false "推荐数量" default(20)
// @Param diversify query bool false "是否多样化" default(true)
// @Param minScore query number false "最小分数" default(0.1)
// @Success 200 {object} response.Response{data=nesma.RecommendationResponse} "获取成功"
// @Router /recommendation/items [get]
func (r *RecommendationApi) GetRecommendations(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 解析参数 - 支持前端发送的参数名称
	recommendationType := c.Query("recommendationType")
	if recommendationType == "" {
		recommendationType = c.Query("type") // 兼容旧参数名
	}
	if recommendationType == "" {
		recommendationType = "all" // 默认值
	}

	algorithm := c.Query("algorithm")
	if algorithm == "" {
		algorithm = "hybrid" // 默认混合算法
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}
	// 兼容旧参数名
	if maxResultsStr := c.Query("maxResults"); maxResultsStr != "" {
		if max, err := strconv.Atoi(maxResultsStr); err == nil && max > 0 {
			limit = max
		}
	}

	minScore := 0.1
	if minScoreStr := c.Query("minScore"); minScoreStr != "" {
		if score, err := strconv.ParseFloat(minScoreStr, 64); err == nil && score >= 0 {
			minScore = score
		}
	}

	diversify := true
	if diversifyStr := c.Query("diversify"); diversifyStr != "" {
		if div, err := strconv.ParseBool(diversifyStr); err == nil {
			diversify = div
		}
	}

	// 构建推荐请求
	req := nesma.RecommendationRequest{
		UserID:             userID,
		RecommendationType: recommendationType,
		Algorithm:          algorithm,
		Limit:              limit,
		Context:            make(map[string]interface{}),
		MinScore:           minScore,
		Diversify:          diversify,
	}

	// 添加上下文信息
	if domain := c.Query("domain"); domain != "" {
		req.Context["domain"] = domain
	}
	if category := c.Query("category"); category != "" {
		req.Context["category"] = category
	}
	if projectType := c.Query("projectType"); projectType != "" {
		req.Context["projectType"] = projectType
	}
	if taskType := c.Query("taskType"); taskType != "" {
		req.Context["taskType"] = taskType
	}

	// 获取推荐服务
	service := nesma.GetRecommendationService()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取推荐
	recommendations, err := service.GetRecommendations(ctx, req)
	if err != nil {
		global.GVA_LOG.Error("获取推荐失败: " + err.Error())
		response.FailWithMessage("获取推荐失败: "+err.Error(), c)
		return
	}

	response.OkWithData(recommendations, c)
}

// UserBehaviorRequest 用户行为记录请求
type UserBehaviorRequest struct {
	Action   string                 `json:"action" binding:"required"`   // 行为类型
	ItemType string                 `json:"itemType" binding:"required"` // 项目类型
	ItemID   uint                   `json:"itemId" binding:"required"`   // 项目ID
	Context  map[string]interface{} `json:"context"`                     // 上下文信息
}

// @Tags 智能推荐
// @Summary 记录用户行为
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body UserBehaviorRequest true "用户行为"
// @Success 200 {object} response.Response{} "记录成功"
// @Router /recommendation/behavior [post]
func (r *RecommendationApi) RecordUserBehavior(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req UserBehaviorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 获取推荐服务
	service := nesma.GetRecommendationService()

	// 记录用户行为
	err := service.RecordUserBehavior(userID, req.Action, req.ItemType, req.ItemID, req.Context)
	if err != nil {
		global.GVA_LOG.Error("记录用户行为失败: " + err.Error())
		response.FailWithMessage("记录失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("行为记录成功", c)
}

// RecommendationFeedbackRequest 推荐反馈请求
type RecommendationFeedbackRequest struct {
	RecommendationID uint   `json:"recommendationId" binding:"required"` // 推荐ID
	ItemType         string `json:"itemType" binding:"required"`         // 项目类型
	ItemID           uint   `json:"itemId" binding:"required"`           // 项目ID
	Clicked          bool   `json:"clicked"`                             // 是否点击
	Liked            *bool  `json:"liked"`                               // 是否喜欢 (null表示未评价)
}

// @Tags 智能推荐
// @Summary 记录推荐反馈
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body RecommendationFeedbackRequest true "推荐反馈"
// @Success 200 {object} response.Response{} "记录成功"
// @Router /recommendation/feedback [post]
func (r *RecommendationApi) RecordRecommendationFeedback(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req RecommendationFeedbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 获取推荐服务
	service := nesma.GetRecommendationService()

	// 记录推荐反馈
	err := service.RecordRecommendationFeedback(userID, req.RecommendationID, req.ItemType, req.ItemID, req.Clicked, req.Liked)
	if err != nil {
		global.GVA_LOG.Error("记录推荐反馈失败: " + err.Error())
		response.FailWithMessage("记录失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("反馈记录成功", c)
}

// @Tags 智能推荐
// @Summary 获取推荐统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param days query int false "统计天数" default(30)
// @Param userId query int false "用户ID（管理员可查看其他用户）"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /recommendation/stats [get]
func (r *RecommendationApi) GetRecommendationStats(c *gin.Context) {
	userID := utils.GetUserID(c)

	// 解析参数
	days := 30
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	var targetUserID *uint
	if userIDStr := c.Query("userId"); userIDStr != "" {
		if id, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			targetUserIDUint := uint(id)
			targetUserID = &targetUserIDUint
		}
	}

	// 如果查询其他用户统计，需要管理员权限
	if targetUserID != nil && *targetUserID != userID {
		// 这里可以添加权限检查
		// if !hasAdminPermission(userID) {
		//     response.FailWithMessage("无权限查看其他用户统计", c)
		//     return
		// }
	} else {
		targetUserID = &userID
	}

	// 获取推荐服务
	service := nesma.GetRecommendationService()

	// 获取统计信息
	stats, err := service.GetRecommendationStats(*targetUserID, days)
	if err != nil {
		global.GVA_LOG.Error("获取推荐统计失败: " + err.Error())
		response.FailWithMessage("获取统计失败: "+err.Error(), c)
		return
	}

	response.OkWithData(stats, c)
}

// PersonalizedRecommendationRequest 个性化推荐请求
type PersonalizedRecommendationRequest struct {
	Types      []string               `json:"types" binding:"required"` // 推荐类型列表
	ProjectID  *uint                  `json:"projectId"`                // 项目ID
	Context    map[string]interface{} `json:"context"`                  // 上下文信息
	MaxResults int                    `json:"maxResults"`               // 最大结果数
	MinScore   float64                `json:"minScore"`                 // 最小分数
	Diversify  bool                   `json:"diversify"`                // 是否多样化
}

// @Tags 智能推荐
// @Summary 获取个性化混合推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body PersonalizedRecommendationRequest true "推荐请求"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /recommendation/personalized [post]
func (r *RecommendationApi) GetPersonalizedRecommendations(c *gin.Context) {
	userID := utils.GetUserID(c)

	var req PersonalizedRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 设置默认值
	if req.MaxResults <= 0 {
		req.MaxResults = 20
	}
	if req.MinScore <= 0 {
		req.MinScore = 0.1
	}
	if req.Context == nil {
		req.Context = make(map[string]interface{})
	}

	// 获取推荐服务
	service := nesma.GetRecommendationService()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 为每种类型获取推荐
	allRecommendations := make(map[string]*nesma.RecommendationResponse)
	var totalItems []nesma.RecommendationItem

	for _, recType := range req.Types {
		recReq := nesma.RecommendationRequest{
			UserID:             userID,
			RecommendationType: recType,
			Algorithm:          "hybrid",
			Limit:              req.MaxResults / len(req.Types), // 平均分配结果数
			Context:            req.Context,
			MinScore:           req.MinScore,
			Diversify:          req.Diversify,
		}

		recommendations, err := service.GetRecommendations(ctx, recReq)
		if err != nil {
			global.GVA_LOG.Error("获取推荐失败: " + err.Error())
			continue
		}

		allRecommendations[recType] = recommendations
		totalItems = append(totalItems, recommendations.Items...)
	}

	// 如果需要多样化，重新排序混合结果
	if req.Diversify {
		totalItems = r.diversifyRecommendations(totalItems, req.MaxResults)
	} else {
		// 按分数排序
		for i := 0; i < len(totalItems)-1; i++ {
			for j := 0; j < len(totalItems)-i-1; j++ {
				if totalItems[j].Score < totalItems[j+1].Score {
					totalItems[j], totalItems[j+1] = totalItems[j+1], totalItems[j]
				}
			}
		}

		// 限制结果数量
		if len(totalItems) > req.MaxResults {
			totalItems = totalItems[:req.MaxResults]
		}
	}

	result := map[string]interface{}{
		"mixed":       totalItems,
		"byType":      allRecommendations,
		"total":       len(totalItems),
		"requestId":   time.Now().UnixNano(),
		"diversified": req.Diversify,
	}

	response.OkWithData(result, c)
}

// diversifyRecommendations 多样化推荐结果
func (r *RecommendationApi) diversifyRecommendations(items []nesma.RecommendationItem, maxResults int) []nesma.RecommendationItem {
	if len(items) <= maxResults {
		return items
	}

	// 按类型分组
	typeGroups := make(map[string][]nesma.RecommendationItem)
	for _, item := range items {
		typeGroups[item.Type] = append(typeGroups[item.Type], item)
	}

	// 对每组按分数排序
	for itemType := range typeGroups {
		group := typeGroups[itemType]
		for i := 0; i < len(group)-1; i++ {
			for j := 0; j < len(group)-i-1; j++ {
				if group[j].Score < group[j+1].Score {
					group[j], group[j+1] = group[j+1], group[j]
				}
			}
		}
		typeGroups[itemType] = group
	}

	// 轮流从每个类型中选择项目
	result := make([]nesma.RecommendationItem, 0, maxResults)
	typeList := make([]string, 0, len(typeGroups))
	for itemType := range typeGroups {
		typeList = append(typeList, itemType)
	}

	typeIndex := 0
	for len(result) < maxResults {
		hasItem := false

		for i := 0; i < len(typeList); i++ {
			currentType := typeList[(typeIndex+i)%len(typeList)]
			if len(typeGroups[currentType]) > 0 {
				// 取第一个（分数最高的）
				result = append(result, typeGroups[currentType][0])
				typeGroups[currentType] = typeGroups[currentType][1:]
				hasItem = true
			}
		}

		if !hasItem {
			break // 所有类型都没有项目了
		}

		typeIndex++
	}

	return result
}

// @Tags 智能推荐
// @Summary 获取推荐配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /recommendation/config [get]
func (r *RecommendationApi) GetRecommendationConfig(c *gin.Context) {
	config := map[string]interface{}{
		"supportedTypes": []string{
			"requirement", "template", "agent", "project", "knowledge",
		},
		"supportedActions": []string{
			"view", "click", "use", "download", "edit", "favorite",
			"share", "bookmark", "reference", "configure", "collaborate",
		},
		"defaultMaxResults": 10,
		"defaultMinScore":   0.1,
		"algorithms": []string{
			"Collaborative Filtering",
			"Content-based",
			"Popularity-based",
			"Hybrid",
		},
		"contextFields": map[string][]string{
			"project":   {"domain", "projectType"},
			"template":  {"format", "type", "category"},
			"agent":     {"taskType", "capabilities"},
			"knowledge": {"category", "priority"},
		},
	}

	response.OkWithData(config, c)
}
