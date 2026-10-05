package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// KnowledgeService 知识库管理服务
type KnowledgeService struct {
	vectorService VectorService
	aiService     AIService
}

// NewKnowledgeService 创建知识库管理服务
func NewKnowledgeService() *KnowledgeService {
	return &KnowledgeService{
		vectorService: NewVectorService(),
		aiService:     GetAIService(),
	}
}

// CreateKnowledgeEntry 创建知识条目
func (s *KnowledgeService) CreateKnowledgeEntry(ctx context.Context, req *request.CreateKnowledgeEntryRequest) (*nesma.NesmaKnowledgeEntry, error) {
	// 检查必填字段
	if req.Title == "" || req.Content == "" || req.Category == "" {
		return nil, errors.New("标题、内容、类别不能为空")
	}

	// 检查类别是否有效
	validCategories := []string{"NESMA_STANDARD", "BEST_PRACTICE", "CASE_STUDY", "RULE"}
	isValidCategory := false
	for _, category := range validCategories {
		if req.Category == category {
			isValidCategory = true
			break
		}
	}
	if !isValidCategory {
		return nil, errors.New("无效的类别")
	}

	// 生成向量表示 - 暂时跳过，后续集成AI服务时实现
	var embedding []float64

	// 创建知识条目
	entry := &nesma.NesmaKnowledgeEntry{
		Title:           req.Title,
		Content:         req.Content,
		Category:        req.Category,
		Domain:          req.Domain,
		Tags:            req.Tags,
		ConfidenceScore: req.ConfidenceScore,
		Version:         req.Version,
		Status:          "active",
		Source:          req.Source,
		Author:          req.Author,
		UsageCount:      0,
	}

	// 保存向量表示
	if embedding != nil {
		embeddingJSON, _ := json.Marshal(embedding)
		entry.Embedding = string(embeddingJSON)
	}

	// 保存到数据库
	if err := global.GVA_DB.Create(entry).Error; err != nil {
		return nil, fmt.Errorf("创建知识条目失败: %w", err)
	}

	return entry, nil
}

// GetKnowledgeEntry 获取知识条目
func (s *KnowledgeService) GetKnowledgeEntry(ctx context.Context, id uint) (*nesma.NesmaKnowledgeEntry, error) {
	var entry nesma.NesmaKnowledgeEntry
	if err := global.GVA_DB.First(&entry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("知识条目不存在")
		}
		return nil, fmt.Errorf("获取知识条目失败: %w", err)
	}

	// 增加使用次数
	global.GVA_DB.Model(&entry).Update("usage_count", gorm.Expr("usage_count + 1"))

	return &entry, nil
}

// UpdateKnowledgeEntry 更新知识条目
func (s *KnowledgeService) UpdateKnowledgeEntry(ctx context.Context, id uint, req *request.UpdateKnowledgeEntryRequest) (*nesma.NesmaKnowledgeEntry, error) {
	var entry nesma.NesmaKnowledgeEntry
	if err := global.GVA_DB.First(&entry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("知识条目不存在")
		}
		return nil, fmt.Errorf("获取知识条目失败: %w", err)
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Content != "" {
		updates["content"] = req.Content
		// 如果内容更新了，重新生成向量 - 暂时跳过
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	}
	if req.Tags != nil {
		updates["tags"] = req.Tags
	}
	if req.ConfidenceScore > 0 {
		updates["confidence_score"] = req.ConfidenceScore
	}
	if req.Version != "" {
		updates["version"] = req.Version
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Source != "" {
		updates["source"] = req.Source
	}
	if req.Author != "" {
		updates["author"] = req.Author
	}

	// 更新数据库
	if err := global.GVA_DB.Model(&entry).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新知识条目失败: %w", err)
	}

	// 重新获取更新后的数据
	if err := global.GVA_DB.First(&entry, id).Error; err != nil {
		return nil, fmt.Errorf("获取更新后的知识条目失败: %w", err)
	}

	return &entry, nil
}

// DeleteKnowledgeEntry 删除知识条目
func (s *KnowledgeService) DeleteKnowledgeEntry(ctx context.Context, id uint) error {
	var entry nesma.NesmaKnowledgeEntry
	if err := global.GVA_DB.First(&entry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("知识条目不存在")
		}
		return fmt.Errorf("获取知识条目失败: %w", err)
	}

	// 删除数据库记录
	if err := global.GVA_DB.Delete(&entry).Error; err != nil {
		return fmt.Errorf("删除知识条目失败: %w", err)
	}

	return nil
}

// ListKnowledgeEntries 列出知识条目
func (s *KnowledgeService) ListKnowledgeEntries(ctx context.Context, req *request.ListKnowledgeEntriesRequest) (*response.ListKnowledgeEntriesResponse, error) {
	var entries []nesma.NesmaKnowledgeEntry
	var total int64

	// 构建查询
	query := global.GVA_DB.Model(&nesma.NesmaKnowledgeEntry{})

	// 添加过滤条件
	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.Domain != "" {
		query = query.Where("domain = ?", req.Domain)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Keyword != "" {
		query = query.Where("title LIKE ? OR content LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取知识条目总数失败: %w", err)
	}

	// 分页查询
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("获取知识条目列表失败: %w", err)
	}

	return &response.ListKnowledgeEntriesResponse{
		Entries:  entries,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetKnowledgeStatistics 获取知识库统计信息
func (s *KnowledgeService) GetKnowledgeStatistics(ctx context.Context) (*response.KnowledgeStatisticsResponse, error) {
	var totalEntries int64
	var totalRules int64

	// 获取总条目数
	if err := global.GVA_DB.Model(&nesma.NesmaKnowledgeEntry{}).Count(&totalEntries).Error; err != nil {
		return nil, fmt.Errorf("获取知识条目总数失败: %w", err)
	}

	// 获取总规则数
	if err := global.GVA_DB.Model(&nesma.NesmaKnowledgeRule{}).Count(&totalRules).Error; err != nil {
		return nil, fmt.Errorf("获取知识规则总数失败: %w", err)
	}

	// 获取分类统计
	categoryStats := make(map[string]int64)
	var categoryResults []struct {
		Category string
		Count    int64
	}
	if err := global.GVA_DB.Model(&nesma.NesmaKnowledgeEntry{}).
		Select("category, COUNT(*) as count").
		Group("category").
		Scan(&categoryResults).Error; err != nil {
		return nil, fmt.Errorf("获取分类统计失败: %w", err)
	}
	for _, result := range categoryResults {
		categoryStats[result.Category] = result.Count
	}

	// 获取领域统计
	domainStats := make(map[string]int64)
	var domainResults []struct {
		Domain string
		Count  int64
	}
	if err := global.GVA_DB.Model(&nesma.NesmaKnowledgeEntry{}).
		Select("domain, COUNT(*) as count").
		Where("domain IS NOT NULL AND domain != ''").
		Group("domain").
		Scan(&domainResults).Error; err != nil {
		return nil, fmt.Errorf("获取领域统计失败: %w", err)
	}
	for _, result := range domainResults {
		domainStats[result.Domain] = result.Count
	}

	return &response.KnowledgeStatisticsResponse{
		TotalEntries:  totalEntries,
		TotalRules:    totalRules,
		CategoryStats: categoryStats,
		DomainStats:   domainStats,
	}, nil
}

// ImportNesmaStandardKnowledge 导入NESMA标准知识
func (s *KnowledgeService) ImportNesmaStandardKnowledge(ctx context.Context) error {
	// 预定义的NESMA标准知识
	standardKnowledge := []struct {
		Title    string
		Content  string
		Category string
		Domain   string
		Tags     []string
	}{
		{
			Title:    "NESMA功能点分析基础",
			Content:  "NESMA功能点分析是一种用于估算软件规模的方法，通过分析软件的功能需求来确定其大小。主要包括数据功能（ILF和EIF）和事务功能（EI、EO、EQ）。",
			Category: "NESMA_STANDARD",
			Domain:   "软件度量",
			Tags:     []string{"功能点", "估算", "基础知识"},
		},
		{
			Title:    "内部逻辑文件(ILF)计算规则",
			Content:  "内部逻辑文件是由应用程序维护的逻辑相关数据组。计算时需要考虑数据元素类型(DET)和记录元素类型(RET)的数量。",
			Category: "NESMA_STANDARD",
			Domain:   "数据功能",
			Tags:     []string{"ILF", "数据功能", "计算规则"},
		},
		{
			Title:    "外部输入(EI)识别准则",
			Content:  "外部输入是从应用程序边界外部接收数据或控制信息的基本过程。必须具有唯一的处理逻辑，并且维护一个或多个ILF。",
			Category: "NESMA_STANDARD",
			Domain:   "事务功能",
			Tags:     []string{"EI", "事务功能", "识别准则"},
		},
	}

	// 导入知识
	for _, knowledge := range standardKnowledge {
		tagsJSON, _ := json.Marshal(knowledge.Tags)

		req := &request.CreateKnowledgeEntryRequest{
			Title:           knowledge.Title,
			Content:         knowledge.Content,
			Category:        knowledge.Category,
			Domain:          knowledge.Domain,
			Tags:            datatypes.JSON(tagsJSON),
			ConfidenceScore: 1.0,
			Version:         "1.0",
			Source:          "NESMA标准",
			Author:          "NESMA官方",
		}

		if _, err := s.CreateKnowledgeEntry(ctx, req); err != nil {
			global.GVA_LOG.Error("导入标准知识失败: " + err.Error())
			continue
		}
	}

	return nil
}

// SearchKnowledgeEntries 搜索知识条目
func (s *KnowledgeService) SearchKnowledgeEntries(ctx context.Context, req *request.SearchKnowledgeEntriesRequest) (*response.SearchKnowledgeEntriesResponse, error) {
	// 暂时使用内容搜索，后续实现向量搜索
	searchResults, err := s.vectorService.SearchByContent(ctx, req.Query, req.Limit)
	if err != nil {
		return nil, fmt.Errorf("内容搜索失败: %w", err)
	}

	// 根据搜索结果获取知识条目
	var results []response.SimilarKnowledge
	for _, result := range searchResults {
		if result.Vector != nil && result.Vector.ProjectID != nil && *result.Vector.ProjectID > 0 {
			var entry nesma.NesmaKnowledgeEntry
			if err := global.GVA_DB.First(&entry, *result.Vector.ProjectID).Error; err == nil {
				results = append(results, response.SimilarKnowledge{
					Knowledge:  entry,
					Similarity: float32(result.Similarity),
				})
			}
		}
	}

	return &response.SearchKnowledgeEntriesResponse{
		Results: results,
		Query:   req.Query,
		Count:   len(results),
	}, nil
}

// ============= 知识规则管理 =============

// CreateKnowledgeRule 创建知识规则
func (s *KnowledgeService) CreateKnowledgeRule(ctx context.Context, req *request.CreateKnowledgeRuleRequest) (*nesma.NesmaKnowledgeRule, error) {
	if req.RuleName == "" || req.ConditionExpression == "" || req.ActionExpression == "" {
		return nil, errors.New("规则名称、条件表达式、动作表达式不能为空")
	}

	rule := &nesma.NesmaKnowledgeRule{
		RuleName:            req.RuleName,
		ConditionExpression: req.ConditionExpression,
		ActionExpression:    req.ActionExpression,
		Priority:            req.Priority,
		IsActive:            req.IsActive,
		Description:         req.Description,
		Category:            req.Category,
	}

	if err := global.GVA_DB.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("创建知识规则失败: %w", err)
	}

	return rule, nil
}

// GetKnowledgeRule 获取知识规则
func (s *KnowledgeService) GetKnowledgeRule(ctx context.Context, id uint) (*nesma.NesmaKnowledgeRule, error) {
	var rule nesma.NesmaKnowledgeRule
	if err := global.GVA_DB.First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("知识规则不存在")
		}
		return nil, fmt.Errorf("获取知识规则失败: %w", err)
	}

	return &rule, nil
}

// UpdateKnowledgeRule 更新知识规则
func (s *KnowledgeService) UpdateKnowledgeRule(ctx context.Context, id uint, req *request.UpdateKnowledgeRuleRequest) (*nesma.NesmaKnowledgeRule, error) {
	var rule nesma.NesmaKnowledgeRule
	if err := global.GVA_DB.First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("知识规则不存在")
		}
		return nil, fmt.Errorf("获取知识规则失败: %w", err)
	}

	updates := make(map[string]interface{})
	if req.RuleName != "" {
		updates["rule_name"] = req.RuleName
	}
	if req.ConditionExpression != "" {
		updates["condition_expression"] = req.ConditionExpression
	}
	if req.ActionExpression != "" {
		updates["action_expression"] = req.ActionExpression
	}
	if req.Priority > 0 {
		updates["priority"] = req.Priority
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Category != "" {
		updates["category"] = req.Category
	}

	if err := global.GVA_DB.Model(&rule).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新知识规则失败: %w", err)
	}

	if err := global.GVA_DB.First(&rule, id).Error; err != nil {
		return nil, fmt.Errorf("获取更新后的知识规则失败: %w", err)
	}

	return &rule, nil
}

// DeleteKnowledgeRule 删除知识规则
func (s *KnowledgeService) DeleteKnowledgeRule(ctx context.Context, id uint) error {
	var rule nesma.NesmaKnowledgeRule
	if err := global.GVA_DB.First(&rule, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("知识规则不存在")
		}
		return fmt.Errorf("获取知识规则失败: %w", err)
	}

	if err := global.GVA_DB.Delete(&rule).Error; err != nil {
		return fmt.Errorf("删除知识规则失败: %w", err)
	}

	return nil
}

// ListKnowledgeRules 列出知识规则
func (s *KnowledgeService) ListKnowledgeRules(ctx context.Context, req *request.ListKnowledgeRulesRequest) (*response.ListKnowledgeRulesResponse, error) {
	var rules []nesma.NesmaKnowledgeRule
	var total int64

	query := global.GVA_DB.Model(&nesma.NesmaKnowledgeRule{})

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}
	if req.IsActive != nil {
		query = query.Where("is_active = ?", *req.IsActive)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取知识规则总数失败: %w", err)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("priority DESC, created_at DESC").Find(&rules).Error; err != nil {
		return nil, fmt.Errorf("获取知识规则列表失败: %w", err)
	}

	return &response.ListKnowledgeRulesResponse{
		Rules:    rules,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// ============= 案例研究管理 =============

// CreateCaseStudy 创建案例研究
func (s *KnowledgeService) CreateCaseStudy(ctx context.Context, req *request.CreateCaseStudyRequest) (*nesma.NesmaCaseStudy, error) {
	if req.ProjectName == "" {
		return nil, errors.New("项目名称不能为空")
	}

	var nesmaResultsJSON datatypes.JSON
	if req.NesmaResults != nil {
		resultBytes, _ := json.Marshal(req.NesmaResults)
		nesmaResultsJSON = datatypes.JSON(resultBytes)
	}

	caseStudy := &nesma.NesmaCaseStudy{
		ProjectName:         req.ProjectName,
		Domain:              req.Domain,
		RequirementsSummary: req.RequirementsSummary,
		NesmaResults:        nesmaResultsJSON,
		LessonsLearned:      req.LessonsLearned,
		TotalFunctionPoints: req.TotalFunctionPoints,
		ProjectDuration:     req.ProjectDuration,
		TeamSize:            req.TeamSize,
		Complexity:          req.Complexity,
		Success:             req.Success,
	}

	if err := global.GVA_DB.Create(caseStudy).Error; err != nil {
		return nil, fmt.Errorf("创建案例研究失败: %w", err)
	}

	return caseStudy, nil
}

// GetCaseStudy 获取案例研究
func (s *KnowledgeService) GetCaseStudy(ctx context.Context, id uint) (*nesma.NesmaCaseStudy, error) {
	var caseStudy nesma.NesmaCaseStudy
	if err := global.GVA_DB.First(&caseStudy, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("案例研究不存在")
		}
		return nil, fmt.Errorf("获取案例研究失败: %w", err)
	}

	return &caseStudy, nil
}

// UpdateCaseStudy 更新案例研究
func (s *KnowledgeService) UpdateCaseStudy(ctx context.Context, id uint, req *request.UpdateCaseStudyRequest) (*nesma.NesmaCaseStudy, error) {
	var caseStudy nesma.NesmaCaseStudy
	if err := global.GVA_DB.First(&caseStudy, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("案例研究不存在")
		}
		return nil, fmt.Errorf("获取案例研究失败: %w", err)
	}

	updates := make(map[string]interface{})
	if req.ProjectName != "" {
		updates["project_name"] = req.ProjectName
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	}
	if req.RequirementsSummary != "" {
		updates["requirements_summary"] = req.RequirementsSummary
	}
	if req.NesmaResults != nil {
		resultBytes, _ := json.Marshal(req.NesmaResults)
		updates["nesma_results"] = datatypes.JSON(resultBytes)
	}
	if req.LessonsLearned != "" {
		updates["lessons_learned"] = req.LessonsLearned
	}
	if req.TotalFunctionPoints > 0 {
		updates["total_function_points"] = req.TotalFunctionPoints
	}
	if req.ProjectDuration > 0 {
		updates["project_duration"] = req.ProjectDuration
	}
	if req.TeamSize > 0 {
		updates["team_size"] = req.TeamSize
	}
	if req.Complexity != "" {
		updates["complexity"] = req.Complexity
	}
	if req.Success != nil {
		updates["success"] = *req.Success
	}

	if err := global.GVA_DB.Model(&caseStudy).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新案例研究失败: %w", err)
	}

	if err := global.GVA_DB.First(&caseStudy, id).Error; err != nil {
		return nil, fmt.Errorf("获取更新后的案例研究失败: %w", err)
	}

	return &caseStudy, nil
}

// DeleteCaseStudy 删除案例研究
func (s *KnowledgeService) DeleteCaseStudy(ctx context.Context, id uint) error {
	var caseStudy nesma.NesmaCaseStudy
	if err := global.GVA_DB.First(&caseStudy, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("案例研究不存在")
		}
		return fmt.Errorf("获取案例研究失败: %w", err)
	}

	if err := global.GVA_DB.Delete(&caseStudy).Error; err != nil {
		return fmt.Errorf("删除案例研究失败: %w", err)
	}

	return nil
}

// ListCaseStudies 列出案例研究
func (s *KnowledgeService) ListCaseStudies(ctx context.Context, req *request.ListCaseStudiesRequest) (*response.ListCaseStudiesResponse, error) {
	var cases []nesma.NesmaCaseStudy
	var total int64

	query := global.GVA_DB.Model(&nesma.NesmaCaseStudy{})

	if req.Domain != "" {
		query = query.Where("domain = ?", req.Domain)
	}
	if req.Complexity != "" {
		query = query.Where("complexity = ?", req.Complexity)
	}
	if req.Success != nil {
		query = query.Where("success = ?", *req.Success)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("获取案例研究总数失败: %w", err)
	}

	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("获取案例研究列表失败: %w", err)
	}

	return &response.ListCaseStudiesResponse{
		Cases:    cases,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// SearchCaseStudies 搜索案例研究
func (s *KnowledgeService) SearchCaseStudies(ctx context.Context, req *request.SearchCaseStudiesRequest) (*response.SearchCaseStudiesResponse, error) {
	// 暂时使用简单的数据库搜索，后续实现向量搜索
	var cases []nesma.NesmaCaseStudy
	query := global.GVA_DB.Model(&nesma.NesmaCaseStudy{})

	if req.Query != "" {
		query = query.Where("project_name LIKE ? OR requirements_summary LIKE ? OR lessons_learned LIKE ?",
			"%"+req.Query+"%", "%"+req.Query+"%", "%"+req.Query+"%")
	}

	if req.Domain != "" {
		query = query.Where("domain = ?", req.Domain)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	if err := query.Limit(limit).Find(&cases).Error; err != nil {
		return nil, fmt.Errorf("搜索案例研究失败: %w", err)
	}

	var results []response.SimilarCase
	for _, caseStudy := range cases {
		results = append(results, response.SimilarCase{
			Case:       caseStudy,
			Similarity: 1.0, // 暂时设置为1.0，后续实现真正的相似度计算
		})
	}

	return &response.SearchCaseStudiesResponse{
		Results: results,
		Query:   req.Query,
		Count:   len(results),
	}, nil
}

// GetPopularTags 获取热门标签
func (s *KnowledgeService) GetPopularTags(ctx context.Context, limit int) ([]string, error) {
	var entries []nesma.NesmaKnowledgeEntry
	if err := global.GVA_DB.Select("tags").Where("tags IS NOT NULL").Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("获取标签失败: %w", err)
	}

	tagCount := make(map[string]int)
	for _, entry := range entries {
		var tags []string
		if err := json.Unmarshal(entry.Tags, &tags); err == nil {
			for _, tag := range tags {
				tagCount[tag]++
			}
		}
	}

	// 简单排序，取前N个
	var result []string
	for tag := range tagCount {
		result = append(result, tag)
		if len(result) >= limit {
			break
		}
	}

	return result, nil
}

// GetRecommendedKnowledge 获取推荐知识
func (s *KnowledgeService) GetRecommendedKnowledge(ctx context.Context, userID uint, limit int) ([]nesma.NesmaKnowledgeEntry, error) {
	var entries []nesma.NesmaKnowledgeEntry

	if err := global.GVA_DB.
		Where("status = ?", "active").
		Order("usage_count DESC, created_at DESC").
		Limit(limit).
		Find(&entries).Error; err != nil {
		return nil, fmt.Errorf("获取推荐知识失败: %w", err)
	}

	return entries, nil
}
