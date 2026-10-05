package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type KnowledgeApi struct{}

var knowledgeService *nesma.KnowledgeService
var knowledgeSearchEngine *nesma.KnowledgeSearchEngine

// getKnowledgeService 获取知识库服务实例，支持延迟初始化
func getKnowledgeService() *nesma.KnowledgeService {
	if knowledgeService == nil {
		knowledgeService = nesma.NewKnowledgeService()
	}
	return knowledgeService
}

// getKnowledgeSearchEngine 获取知识搜索引擎实例
func getKnowledgeSearchEngine() *nesma.KnowledgeSearchEngine {
	if knowledgeSearchEngine == nil {
		knowledgeSearchEngine = nesma.NewKnowledgeSearchEngine()
	}
	return knowledgeSearchEngine
}

// CreateKnowledgeEntry 创建知识条目
func (k *KnowledgeApi) CreateKnowledgeEntry(c *gin.Context) {
	var req nesmaReq.CreateKnowledgeEntryRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	entry, err := getKnowledgeService().CreateKnowledgeEntry(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("创建知识条目失败!", zap.Error(err))
		response.FailWithMessage("创建知识条目失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(entry, "创建知识条目成功", c)
}

// GetKnowledgeEntry 获取知识条目
func (k *KnowledgeApi) GetKnowledgeEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	entry, err := getKnowledgeService().GetKnowledgeEntry(c.Request.Context(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取知识条目失败!", zap.Error(err))
		response.FailWithMessage("获取知识条目失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(entry, "获取知识条目成功", c)
}

// ListKnowledgeEntries 列出知识条目
func (k *KnowledgeApi) ListKnowledgeEntries(c *gin.Context) {
	var req nesmaReq.ListKnowledgeEntriesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := getKnowledgeService().ListKnowledgeEntries(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("获取知识条目列表失败!", zap.Error(err))
		response.FailWithMessage("获取知识条目列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(list, "获取知识条目列表成功", c)
}

// SearchKnowledgeEntries 搜索知识条目
func (k *KnowledgeApi) SearchKnowledgeEntries(c *gin.Context) {
	var req nesmaReq.SearchKnowledgeEntriesRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := getKnowledgeService().SearchKnowledgeEntries(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("搜索知识条目失败!", zap.Error(err))
		response.FailWithMessage("搜索知识条目失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(result, "搜索知识条目成功", c)
}

// GetKnowledgeStatistics 获取知识库统计信息
func (k *KnowledgeApi) GetKnowledgeStatistics(c *gin.Context) {
	stats, err := getKnowledgeService().GetKnowledgeStatistics(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("获取知识库统计失败!", zap.Error(err))
		response.FailWithMessage("获取知识库统计失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(stats, "获取知识库统计成功", c)
}

// ImportNesmaStandardKnowledge 导入NESMA标准知识
func (k *KnowledgeApi) ImportNesmaStandardKnowledge(c *gin.Context) {
	err := getKnowledgeService().ImportNesmaStandardKnowledge(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("导入NESMA标准知识失败!", zap.Error(err))
		response.FailWithMessage("导入NESMA标准知识失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("导入NESMA标准知识成功", c)
}

// UpdateKnowledgeEntry 更新知识条目
func (k *KnowledgeApi) UpdateKnowledgeEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var req nesmaReq.UpdateKnowledgeEntryRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	entry, err := getKnowledgeService().UpdateKnowledgeEntry(c.Request.Context(), uint(id), &req)
	if err != nil {
		global.GVA_LOG.Error("更新知识条目失败!", zap.Error(err))
		response.FailWithMessage("更新知识条目失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(entry, "更新知识条目成功", c)
}

// DeleteKnowledgeEntry 删除知识条目
func (k *KnowledgeApi) DeleteKnowledgeEntry(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	err = getKnowledgeService().DeleteKnowledgeEntry(c.Request.Context(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除知识条目失败!", zap.Error(err))
		response.FailWithMessage("删除知识条目失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除知识条目成功", c)
}

// GetPopularTags 获取热门标签
func (k *KnowledgeApi) GetPopularTags(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	tags, err := getKnowledgeService().GetPopularTags(c.Request.Context(), limit)
	if err != nil {
		global.GVA_LOG.Error("获取热门标签失败!", zap.Error(err))
		response.FailWithMessage("获取热门标签失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(tags, "获取热门标签成功", c)
}

// GetRecommendedKnowledge 获取推荐知识
func (k *KnowledgeApi) GetRecommendedKnowledge(c *gin.Context) {
	// 从JWT token中获取用户ID
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.FailWithMessage("无效的用户ID", c)
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	knowledge, err := getKnowledgeService().GetRecommendedKnowledge(c.Request.Context(), userID, limit)
	if err != nil {
		global.GVA_LOG.Error("获取推荐知识失败!", zap.Error(err))
		response.FailWithMessage("获取推荐知识失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(knowledge, "获取推荐知识成功", c)
}

// CreateKnowledgeRule 创建知识规则
func (k *KnowledgeApi) CreateKnowledgeRule(c *gin.Context) {
	var req nesmaReq.CreateKnowledgeRuleRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	rule, err := getKnowledgeService().CreateKnowledgeRule(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("创建知识规则失败!", zap.Error(err))
		response.FailWithMessage("创建知识规则失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(rule, "创建知识规则成功", c)
}

// GetKnowledgeRule 获取知识规则
func (k *KnowledgeApi) GetKnowledgeRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	rule, err := getKnowledgeService().GetKnowledgeRule(c.Request.Context(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取知识规则失败!", zap.Error(err))
		response.FailWithMessage("获取知识规则失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(rule, "获取知识规则成功", c)
}

// UpdateKnowledgeRule 更新知识规则
func (k *KnowledgeApi) UpdateKnowledgeRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var req nesmaReq.UpdateKnowledgeRuleRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	rule, err := getKnowledgeService().UpdateKnowledgeRule(c.Request.Context(), uint(id), &req)
	if err != nil {
		global.GVA_LOG.Error("更新知识规则失败!", zap.Error(err))
		response.FailWithMessage("更新知识规则失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(rule, "更新知识规则成功", c)
}

// DeleteKnowledgeRule 删除知识规则
func (k *KnowledgeApi) DeleteKnowledgeRule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	err = getKnowledgeService().DeleteKnowledgeRule(c.Request.Context(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除知识规则失败!", zap.Error(err))
		response.FailWithMessage("删除知识规则失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除知识规则成功", c)
}

// ListKnowledgeRules 列出知识规则
func (k *KnowledgeApi) ListKnowledgeRules(c *gin.Context) {
	var req nesmaReq.ListKnowledgeRulesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := getKnowledgeService().ListKnowledgeRules(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("获取知识规则列表失败!", zap.Error(err))
		response.FailWithMessage("获取知识规则列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(list, "获取知识规则列表成功", c)
}

// CreateCaseStudy 创建案例研究
func (k *KnowledgeApi) CreateCaseStudy(c *gin.Context) {
	var req nesmaReq.CreateCaseStudyRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	caseStudy, err := getKnowledgeService().CreateCaseStudy(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("创建案例研究失败!", zap.Error(err))
		response.FailWithMessage("创建案例研究失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(caseStudy, "创建案例研究成功", c)
}

// GetCaseStudy 获取案例研究
func (k *KnowledgeApi) GetCaseStudy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	caseStudy, err := getKnowledgeService().GetCaseStudy(c.Request.Context(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取案例研究失败!", zap.Error(err))
		response.FailWithMessage("获取案例研究失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(caseStudy, "获取案例研究成功", c)
}

// UpdateCaseStudy 更新案例研究
func (k *KnowledgeApi) UpdateCaseStudy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	var req nesmaReq.UpdateCaseStudyRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	caseStudy, err := getKnowledgeService().UpdateCaseStudy(c.Request.Context(), uint(id), &req)
	if err != nil {
		global.GVA_LOG.Error("更新案例研究失败!", zap.Error(err))
		response.FailWithMessage("更新案例研究失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(caseStudy, "更新案例研究成功", c)
}

// DeleteCaseStudy 删除案例研究
func (k *KnowledgeApi) DeleteCaseStudy(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("无效的ID", c)
		return
	}

	err = getKnowledgeService().DeleteCaseStudy(c.Request.Context(), uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除案例研究失败!", zap.Error(err))
		response.FailWithMessage("删除案例研究失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除案例研究成功", c)
}

// ListCaseStudies 列出案例研究
func (k *KnowledgeApi) ListCaseStudies(c *gin.Context) {
	var req nesmaReq.ListCaseStudiesRequest
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, err := getKnowledgeService().ListCaseStudies(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("获取案例研究列表失败!", zap.Error(err))
		response.FailWithMessage("获取案例研究列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(list, "获取案例研究列表成功", c)
}

// SearchCaseStudies 搜索案例研究
func (k *KnowledgeApi) SearchCaseStudies(c *gin.Context) {
	var req nesmaReq.SearchCaseStudiesRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := getKnowledgeService().SearchCaseStudies(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("搜索案例研究失败!", zap.Error(err))
		response.FailWithMessage("搜索案例研究失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(result, "搜索案例研究成功", c)
}

// IntelligentKnowledgeSearch 智能知识搜索
func (k *KnowledgeApi) IntelligentKnowledgeSearch(c *gin.Context) {
	var req nesma.KnowledgeSearchRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 设置默认值
	if req.Type == "" {
		req.Type = "hybrid"
	}
	if req.MaxResults <= 0 {
		req.MaxResults = 10
	}
	if req.MinRelevance <= 0 {
		req.MinRelevance = 0.1
	}

	result, err := getKnowledgeSearchEngine().Search(c.Request.Context(), &req)
	if err != nil {
		global.GVA_LOG.Error("智能知识搜索失败!", zap.Error(err))
		response.FailWithMessage("智能知识搜索失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(result, "智能知识搜索成功", c)
}

// BuildKnowledgeIndex 构建知识库向量索引
func (k *KnowledgeApi) BuildKnowledgeIndex(c *gin.Context) {
	err := getKnowledgeSearchEngine().BuildKnowledgeVectorIndex(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("构建知识库索引失败!", zap.Error(err))
		response.FailWithMessage("构建知识库索引失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("知识库索引构建已开始，请查看日志了解进度", c)
}

// GetKnowledgeSearchSuggestions 获取知识搜索建议
func (k *KnowledgeApi) GetKnowledgeSearchSuggestions(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		response.FailWithMessage("查询内容不能为空", c)
		return
	}

	// 模拟搜索建议生成
	suggestions := []string{
		query + " NESMA标准",
		query + " 功能点分析",
		query + " 最佳实践",
		query + " 案例研究",
	}

	response.OkWithDetailed(map[string]interface{}{
		"query":       query,
		"suggestions": suggestions,
	}, "获取搜索建议成功", c)
}

// GetKnowledgeSearchHistory 获取知识搜索历史
func (k *KnowledgeApi) GetKnowledgeSearchHistory(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		response.FailWithMessage("无效的用户ID", c)
		return
	}

	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	// TODO: 实现搜索历史查询逻辑
	history := []map[string]interface{}{
		{
			"id":         1,
			"query":      "NESMA功能点分析",
			"type":       "hybrid",
			"results":    8,
			"created_at": "2025-01-08T10:30:00Z",
		},
	}

	response.OkWithDetailed(map[string]interface{}{
		"history": history,
		"total":   len(history),
		"limit":   limit,
	}, "获取搜索历史成功", c)
}

// 🔧 添加路由中缺失的方法
// BatchCreateKnowledgeEntries 批量创建知识条目
func (k *KnowledgeApi) BatchCreateKnowledgeEntries(c *gin.Context) {
	var req []nesmaReq.CreateKnowledgeEntryRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// TODO: 实现批量创建逻辑
	response.OkWithMessage("批量创建知识条目功能待实现", c)
}

// BatchDeleteKnowledgeEntries 批量删除知识条目
func (k *KnowledgeApi) BatchDeleteKnowledgeEntries(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids"`
	}
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// TODO: 实现批量删除逻辑
	response.OkWithMessage("批量删除知识条目功能待实现", c)
}

// ImportKnowledgeEntries 导入知识条目
func (k *KnowledgeApi) ImportKnowledgeEntries(c *gin.Context) {
	// TODO: 实现导入逻辑
	response.OkWithMessage("导入知识条目功能待实现", c)
}

// GetKnowledgeEntryList 获取知识条目列表（别名）
func (k *KnowledgeApi) GetKnowledgeEntryList(c *gin.Context) {
	k.ListKnowledgeEntries(c)
}

// GetKnowledgeCategories 获取知识分类
func (k *KnowledgeApi) GetKnowledgeCategories(c *gin.Context) {
	// TODO: 实现获取分类逻辑
	categories := []string{"NESMA标准", "功能点分析", "最佳实践", "案例研究"}
	response.OkWithDetailed(categories, "获取知识分类成功", c)
}

// GetKnowledgeTags 获取知识标签
func (k *KnowledgeApi) GetKnowledgeTags(c *gin.Context) {
	k.GetPopularTags(c)
}

// GetKnowledgeStats 获取知识库统计
func (k *KnowledgeApi) GetKnowledgeStats(c *gin.Context) {
	k.GetKnowledgeStatistics(c)
}

// GetKnowledgeRecommendations 获取知识推荐
func (k *KnowledgeApi) GetKnowledgeRecommendations(c *gin.Context) {
	k.GetRecommendedKnowledge(c)
}

// VectorSearchKnowledge 向量搜索知识
func (k *KnowledgeApi) VectorSearchKnowledge(c *gin.Context) {
	// TODO: 实现向量搜索逻辑
	response.OkWithMessage("向量搜索功能待实现", c)
}

// SemanticSearchKnowledge 语义搜索知识
func (k *KnowledgeApi) SemanticSearchKnowledge(c *gin.Context) {
	k.IntelligentKnowledgeSearch(c)
}

// GetRelatedKnowledge 获取相关知识
func (k *KnowledgeApi) GetRelatedKnowledge(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		response.FailWithMessage("无效的ID", c)
		return
	}

	// TODO: 实现获取相关知识逻辑
	response.OkWithMessage("获取相关知识功能待实现", c)
}

// GetPublicKnowledgeEntries 获取公开知识条目
func (k *KnowledgeApi) GetPublicKnowledgeEntries(c *gin.Context) {
	// TODO: 实现获取公开知识条目逻辑
	response.OkWithMessage("获取公开知识条目功能待实现", c)
}

// GetKnowledgeHelp 获取知识库帮助
func (k *KnowledgeApi) GetKnowledgeHelp(c *gin.Context) {
	help := map[string]interface{}{
		"title":       "NESMA知识库帮助",
		"description": "NESMA知识库提供功能点分析的标准、最佳实践和案例研究",
		"features": []string{
			"知识条目搜索",
			"智能推荐",
			"案例研究",
			"规则管理",
		},
	}
	response.OkWithDetailed(help, "获取知识库帮助成功", c)
}
