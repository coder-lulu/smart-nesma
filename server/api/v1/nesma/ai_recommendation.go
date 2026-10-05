package nesma

import (
	"context"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
)

type AIRecommendationApi struct{}

// AIRecommendationRequest AI推荐请求
type AIRecommendationRequest struct {
	ProjectID       uint   `json:"projectId" binding:"required"`
	ProjectName     string `json:"projectName"`
	Domain          string `json:"domain"`
	RequirementText string `json:"requirementText"`
	Context         map[string]interface{} `json:"context"`
}

// AdoptRecommendationRequest 采纳推荐请求
type AdoptRecommendationRequest struct {
	ProjectID        uint   `json:"projectId" binding:"required"`
	RecommendationID string `json:"recommendationId" binding:"required"`
	Title            string `json:"title" binding:"required"`
	Content          string `json:"content" binding:"required"`
	Category         string `json:"category"`
	Confidence       float64 `json:"confidence"`
	Feedback         string `json:"feedback"`
}

// RejectRecommendationRequest 拒绝推荐请求
type RejectRecommendationRequest struct {
	ProjectID        uint   `json:"projectId" binding:"required"`
	RecommendationID string `json:"recommendationId" binding:"required"`
	Reason           string `json:"reason"`
}

// @Tags AI智能推荐
// @Summary 获取AI智能推荐
// @Description 基于知识库和AI大模型生成智能推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body AIRecommendationRequest true "推荐请求"
// @Success 200 {object} response.Response{data=nesma.AIRecommendationResponse} "获取成功"
// @Router /recommendation/ai [post]
func (r *AIRecommendationApi) GetAIRecommendations(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	var req AIRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	
	// 获取AI推荐服务
	service := nesmaService.GetAIRecommendationService()
	
	// 构建推荐请求
	aiReq := &nesmaService.AIRecommendationRequest{
		UserID:          userID,
		ProjectID:       req.ProjectID,
		ProjectName:     req.ProjectName,
		Domain:          req.Domain,
		RequirementText: req.RequirementText,
		Context:         req.Context,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	// 获取AI推荐
	recommendations, err := service.GetIntelligentRecommendations(ctx, aiReq)
	if err != nil {
		global.GVA_LOG.Error("获取AI推荐失败: " + err.Error())
		response.FailWithMessage("获取AI推荐失败: "+err.Error(), c)
		return
	}
	
	response.OkWithData(recommendations, c)
}

// @Tags AI智能推荐
// @Summary 采纳AI推荐
// @Description 用户采纳AI推荐并将其加入知识库
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body AdoptRecommendationRequest true "采纳请求"
// @Success 200 {object} response.Response{} "采纳成功"
// @Router /recommendation/adopt [post]
func (r *AIRecommendationApi) AdoptRecommendation(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	var req AdoptRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	
	// 获取AI推荐服务
	service := nesmaService.GetAIRecommendationService()
	
	// 构建推荐项目
	recommendation := &nesmaService.AIRecommendationItem{
		ID:         req.RecommendationID,
		Title:      req.Title,
		Content:    req.Content,
		Category:   req.Category,
		Confidence: req.Confidence,
		CreatedAt:  time.Now(),
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 采纳推荐
	err := service.AdoptRecommendation(ctx, userID, req.ProjectID, recommendation, req.Feedback)
	if err != nil {
		global.GVA_LOG.Error("采纳推荐失败: " + err.Error())
		response.FailWithMessage("采纳推荐失败: "+err.Error(), c)
		return
	}
	
	response.OkWithMessage("推荐采纳成功，已加入知识库", c)
}

// @Tags AI智能推荐
// @Summary 拒绝AI推荐
// @Description 用户拒绝AI推荐，用于优化推荐算法
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body RejectRecommendationRequest true "拒绝请求"
// @Success 200 {object} response.Response{} "拒绝成功"
// @Router /recommendation/reject [post]
func (r *AIRecommendationApi) RejectRecommendation(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	var req RejectRecommendationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	
	// 获取AI推荐服务
	service := nesmaService.GetAIRecommendationService()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	// 拒绝推荐
	err := service.RejectRecommendation(ctx, userID, req.ProjectID, req.RecommendationID, req.Reason)
	if err != nil {
		global.GVA_LOG.Error("拒绝推荐失败: " + err.Error())
		response.FailWithMessage("拒绝推荐失败: "+err.Error(), c)
		return
	}
	
	response.OkWithMessage("推荐拒绝成功", c)
}

// @Tags AI智能推荐
// @Summary 获取采纳历史
// @Description 获取用户的推荐采纳历史记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query uint false "项目ID，不填则返回所有项目"
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Success 200 {object} response.Response{data=[]nesma.RecommendationAdoption} "获取成功"
// @Router /recommendation/adoption-history [get]
func (r *AIRecommendationApi) GetAdoptionHistory(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	// 解析参数
	var projectID uint
	if projectIDStr := c.Query("projectId"); projectIDStr != "" {
		if pid, err := strconv.ParseUint(projectIDStr, 10, 32); err == nil {
			projectID = uint(pid)
		}
	}
	
	// 获取AI推荐服务
	service := nesmaService.GetAIRecommendationService()
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 获取采纳历史
	adoptions, err := service.GetAdoptionHistory(ctx, userID, projectID)
	if err != nil {
		global.GVA_LOG.Error("获取采纳历史失败: " + err.Error())
		response.FailWithMessage("获取采纳历史失败: "+err.Error(), c)
		return
	}
	
	response.OkWithData(adoptions, c)
}

// @Tags AI智能推荐
// @Summary 获取AI推荐配置
// @Description 获取AI推荐功能的配置信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /recommendation/ai-config [get]
func (r *AIRecommendationApi) GetAIRecommendationConfig(c *gin.Context) {
	config := map[string]interface{}{
		"supportedCategories": []string{
			"function_analysis",
			"requirement_optimization", 
			"best_practice",
			"risk_warning",
			"process_improvement",
		},
		"supportedFunctionTypes": []string{
			"EI", "EO", "EQ", "ILF", "EIF",
		},
		"confidenceThreshold": 0.6,
		"maxRecommendations": 3,
		"aiModels": []string{
			"DeepSeek-Chat",
			"OpenAI-GPT",
		},
		"knowledgeCategories": []string{
			"NESMA_STANDARD",
			"BEST_PRACTICE", 
			"DOMAIN_KNOWLEDGE",
			"USER_ADOPTED",
		},
		"processingTimeout": "30s",
		"version": "2.0",
		"description": "基于AI大模型和知识库的智能推荐系统",
	}
	
	response.OkWithData(config, c)
}

// @Tags AI智能推荐
// @Summary 快速推荐
// @Description 基于项目ID快速获取推荐，无需详细参数
// @Security ApiKeyAuth
// @accept application/json  
// @Produce application/json
// @Param projectId query uint true "项目ID"
// @Success 200 {object} response.Response{data=nesma.AIRecommendationResponse} "获取成功"
// @Router /recommendation/quick [get]
func (r *AIRecommendationApi) GetQuickRecommendations(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	// 解析项目ID
	projectIDStr := c.Query("projectId")
	if projectIDStr == "" {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}
	
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}
	
	// 获取项目信息
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, projectID).Error; err != nil {
		response.FailWithMessage("项目不存在", c)
		return
	}
	
	// 获取AI推荐服务
	service := nesmaService.GetAIRecommendationService()
	
	// 构建推荐请求
	aiReq := &nesmaService.AIRecommendationRequest{
		UserID:          userID,
		ProjectID:       uint(projectID),
		ProjectName:     project.Name,
		Domain:          project.Domain,
		RequirementText: project.Description,
		Context:         map[string]interface{}{
			"quickMode": true,
		},
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	
	// 获取AI推荐
	recommendations, err := service.GetIntelligentRecommendations(ctx, aiReq)
	if err != nil {
		global.GVA_LOG.Error("获取快速推荐失败: " + err.Error())
		response.FailWithMessage("获取快速推荐失败: "+err.Error(), c)
		return
	}
	
	response.OkWithData(recommendations, c)
}

// @Tags AI智能推荐
// @Summary 批量获取推荐
// @Description 为多个项目批量获取AI推荐
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectIds body []uint true "项目ID列表"
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /recommendation/batch [post]
func (r *AIRecommendationApi) GetBatchRecommendations(c *gin.Context) {
	userID := utils.GetUserID(c)
	
	var projectIDs []uint
	if err := c.ShouldBindJSON(&projectIDs); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}
	
	if len(projectIDs) == 0 {
		response.FailWithMessage("项目ID列表不能为空", c)
		return
	}
	
	if len(projectIDs) > 10 {
		response.FailWithMessage("批量推荐最多支持10个项目", c)
		return
	}
	
	// 获取AI推荐服务
	service := nesmaService.GetAIRecommendationService()
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	// 批量获取推荐
	results := make(map[string]interface{})
	var successCount, failCount int
	
	for _, projectID := range projectIDs {
		// 获取项目信息
		var project nesma.NesmaProject
		if err := global.GVA_DB.First(&project, projectID).Error; err != nil {
			failCount++
			results[strconv.Itoa(int(projectID))] = map[string]interface{}{
				"error": "项目不存在",
			}
			continue
		}
		
		// 构建推荐请求
		aiReq := &nesmaService.AIRecommendationRequest{
			UserID:          userID,
			ProjectID:       projectID,
			ProjectName:     project.Name,
			Domain:          project.Domain,
			RequirementText: project.Description,
			Context:         map[string]interface{}{
				"batchMode": true,
			},
		}
		
		// 获取AI推荐
		recommendations, err := service.GetIntelligentRecommendations(ctx, aiReq)
		if err != nil {
			failCount++
			results[strconv.Itoa(int(projectID))] = map[string]interface{}{
				"error": err.Error(),
			}
		} else {
			successCount++
			results[strconv.Itoa(int(projectID))] = recommendations
		}
	}
	
	response.OkWithData(map[string]interface{}{
		"results":      results,
		"successCount": successCount,
		"failCount":    failCount,
		"total":        len(projectIDs),
	}, c)
}