package nesma

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NesmaRequirementService struct{}

// CreateNesmaRequirement 创建需求
func (s *NesmaRequirementService) CreateNesmaRequirement(req *nesmaReq.CreateNesmaRequirementRequest) (*nesma.NesmaRequirement, error) {
	// 唯一性校验：同一项目、同一周期、同一版本、同一层级、同一标题下唯一
	var count int64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("project_id = ? AND cycle_id = ? AND version_id = ? AND level = ? AND title = ?", req.ProjectID, req.CycleID, req.VersionID, req.Level, req.Title).
		Count(&count)
	if count > 0 {
		return nil, errors.New("同一周期、版本、层级下已存在相同标题的需求")
	}

	// 验证父级需求
	if req.ParentID != nil {
		var parent nesma.NesmaRequirement
		if err := global.GVA_DB.First(&parent, *req.ParentID).Error; err != nil {
			return nil, errors.New("父级需求不存在")
		}
		// 验证层级关系
		if req.Level <= parent.Level {
			return nil, errors.New("子级需求层级必须大于父级需求层级")
		}
		if req.Level > parent.Level+1 {
			return nil, errors.New("需求层级不能跨级创建")
		}
		// 验证项目一致性
		if req.ProjectID != parent.ProjectID {
			return nil, errors.New("子级需求必须属于同一项目")
		}
	}

	// 验证项目存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return nil, errors.New("项目不存在")
	}

	// 设置默认值
	if req.Status == "" {
		req.Status = "pending"
	}
	if req.Priority == 0 {
		req.Priority = 3
	}

	// 创建需求
	requirement := nesma.NesmaRequirement{
		ProjectID:          req.ProjectID,
		ParentID:           req.ParentID,
		Level:              req.Level,
		Code:               req.Code,
		Title:              req.Title,
		Description:        req.Description,
		Priority:           req.Priority,
		Status:             req.Status,
		OrderIndex:         req.OrderIndex,
		DomainTags:         req.DomainTags,
		Category:           req.Category,
		Complexity:         req.Complexity,
		EstimateHours:      req.EstimateHours,
		ActualHours:        req.ActualHours,
		BusinessValue:      req.BusinessValue,
		AcceptanceCriteria: req.AcceptanceCriteria,
		Notes:              req.Notes,
		// 新的字段结构
		CycleID:          &req.CycleID,
		VersionID:        &req.VersionID,
		AFP:              req.AFP,
		UFP:              req.UFP,
		FunctionType:     req.FunctionType,
		ReuseLevel:       req.ReuseLevel,
		ModificationType: req.ModificationType,
	}

	if err := global.GVA_DB.Create(&requirement).Error; err != nil {
		return nil, err
	}

	return &requirement, nil
}

// UpdateNesmaRequirement 更新需求
func (s *NesmaRequirementService) UpdateNesmaRequirement(req *nesmaReq.UpdateNesmaRequirementRequest) (*nesma.NesmaRequirement, error) {
	var requirement nesma.NesmaRequirement
	if err := global.GVA_DB.First(&requirement, req.ID).Error; err != nil {
		return nil, errors.New("需求不存在")
	}

	// 验证父级需求（如果有变更）
	if req.ParentID != nil && (requirement.ParentID == nil || *req.ParentID != *requirement.ParentID) {
		var parent nesma.NesmaRequirement
		if err := global.GVA_DB.First(&parent, *req.ParentID).Error; err != nil {
			return nil, errors.New("父级需求不存在")
		}
		// 验证层级关系
		if req.Level <= parent.Level {
			return nil, errors.New("子级需求层级必须大于父级需求层级")
		}
		// 验证项目一致性
		if req.ProjectID != parent.ProjectID {
			return nil, errors.New("子级需求必须属于同一项目")
		}
		// 验证不能设置自己为父级
		if *req.ParentID == req.ID {
			return nil, errors.New("不能设置自己为父级")
		}
	}

	// 更新需求
	updates := map[string]interface{}{
		"project_id":          req.ProjectID,
		"parent_id":           req.ParentID,
		"level":               req.Level,
		"code":                req.Code,
		"title":               req.Title,
		"description":         req.Description,
		"priority":            req.Priority,
		"status":              req.Status,
		"order_index":         req.OrderIndex,
		"domain_tags":         req.DomainTags,
		"category":            req.Category,
		"complexity":          req.Complexity,
		"estimate_hours":      req.EstimateHours,
		"actual_hours":        req.ActualHours,
		"business_value":      req.BusinessValue,
		"acceptance_criteria": req.AcceptanceCriteria,
		"notes":               req.Notes,
		"cycle_id":            &req.CycleID,
		"version_id":          &req.VersionID,
		"afp":                 req.AFP,
		"ufp":                 req.UFP,
		"function_type":       req.FunctionType,
		"reuse_level":         req.ReuseLevel,
		"modification_type":   req.ModificationType,
	}

	if err := global.GVA_DB.Model(&requirement).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新查询获取更新后的数据
	if err := global.GVA_DB.Preload("Project").Preload("Parent").First(&requirement, req.ID).Error; err != nil {
		return nil, err
	}

	return &requirement, nil
}

// DeleteNesmaRequirement 删除需求
// GetRequirementChildren 获取需求的子功能点信息
func (s *NesmaRequirementService) GetRequirementChildren(id uint) (nesmaRes.RequirementChildrenResponse, error) {
	var response nesmaRes.RequirementChildrenResponse
	
	// 获取直接子功能点
	var directChildren []nesma.NesmaRequirement
	err := global.GVA_DB.Where("parent_id = ?", id).Find(&directChildren).Error
	if err != nil {
		return response, err
	}
	
	// 填充直接子功能点信息
	for _, child := range directChildren {
		response.DirectChildren = append(response.DirectChildren, nesmaRes.RequirementChildrenInfo{
			ID:     child.ID,
			Title:  child.Title,
			Level:  child.Level,
			Code:   child.Code,
			Status: child.Status,
		})
	}
	
	// 递归获取所有子功能点
	allChildren := s.getAllChildrenRecursive(id)
	for _, child := range allChildren {
		response.AllChildren = append(response.AllChildren, nesmaRes.RequirementChildrenInfo{
			ID:     child.ID,
			Title:  child.Title,
			Level:  child.Level,
			Code:   child.Code,
			Status: child.Status,
		})
	}
	
	// 统计信息
	response.HasChildren = len(directChildren) > 0
	response.ChildrenCount = len(allChildren)
	
	// 各层级数量统计
	response.LevelBreakdown = make(map[string]int)
	for _, child := range allChildren {
		levelKey := fmt.Sprintf("L%d", child.Level)
		response.LevelBreakdown[levelKey]++
	}
	
	return response, nil
}

// getAllChildrenRecursive 递归获取所有子功能点
func (s *NesmaRequirementService) getAllChildrenRecursive(parentID uint) []nesma.NesmaRequirement {
	var allChildren []nesma.NesmaRequirement
	
	// 获取直接子功能点
	var directChildren []nesma.NesmaRequirement
	global.GVA_DB.Where("parent_id = ?", parentID).Find(&directChildren)
	
	// 添加直接子功能点
	allChildren = append(allChildren, directChildren...)
	
	// 递归获取每个子功能点的子功能点
	for _, child := range directChildren {
		grandChildren := s.getAllChildrenRecursive(child.ID)
		allChildren = append(allChildren, grandChildren...)
	}
	
	return allChildren
}

// DeleteNesmaRequirement 删除需求（支持级联删除）
func (s *NesmaRequirementService) DeleteNesmaRequirement(id uint, forceDelete bool) error {
	var requirement nesma.NesmaRequirement
	if err := global.GVA_DB.First(&requirement, id).Error; err != nil {
		return errors.New("需求不存在")
	}

	// 检查是否有子需求
	var childCount int64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).Where("parent_id = ?", id).Count(&childCount)
	
	if childCount > 0 {
		if !forceDelete {
			return errors.New("存在子功能点，无法删除。请先删除子功能点或使用级联删除")
		} else {
			// 级联删除：先删除所有子功能点
			err := s.deleteChildrenRecursive(id)
			if err != nil {
				return fmt.Errorf("删除子功能点失败: %v", err)
			}
		}
	}

	// 删除自身
	return global.GVA_DB.Delete(&requirement).Error
}

// deleteChildrenRecursive 递归删除所有子功能点
func (s *NesmaRequirementService) deleteChildrenRecursive(parentID uint) error {
	// 获取直接子功能点
	var directChildren []nesma.NesmaRequirement
	err := global.GVA_DB.Where("parent_id = ?", parentID).Find(&directChildren).Error
	if err != nil {
		return err
	}
	
	// 递归删除每个子功能点的子功能点
	for _, child := range directChildren {
		err := s.deleteChildrenRecursive(child.ID)
		if err != nil {
			return err
		}
		
		// 删除子功能点自身
		err = global.GVA_DB.Delete(&child).Error
		if err != nil {
			return err
		}
	}
	
	return nil
}

// DeleteRequirementsByCondition 根据条件删除需求
func (s *NesmaRequirementService) DeleteRequirementsByCondition(req *nesmaReq.DeleteRequirementsByConditionRequest) error {
	// 验证项目存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return errors.New("项目不存在")
	}

	// 查询符合条件的需求
	var requirements []nesma.NesmaRequirement
	db := global.GVA_DB.Where("project_id = ?", req.ProjectID)
	
	if req.CycleID != 0 {
		db = db.Where("cycle_id = ?", req.CycleID)
	}
	
	if req.VersionID != 0 {
		db = db.Where("version_id = ?", req.VersionID)
	}
	
	if err := db.Find(&requirements).Error; err != nil {
		return fmt.Errorf("查询需求失败: %v", err)
	}

	if len(requirements) == 0 {
		return errors.New("没有找到符合条件的需求")
	}

	// 检查是否有子需求，如果有，一起删除
	var allIDs []uint
	for _, req := range requirements {
		allIDs = append(allIDs, req.ID)
	}

	// 递归查找所有子需求
	if err := s.collectChildRequirements(allIDs, &allIDs); err != nil {
		return fmt.Errorf("收集子需求失败: %v", err)
	}

	// 开始事务删除
	return global.GVA_DB.Transaction(func(tx *gorm.DB) error {
		// 删除所有需求
		if err := tx.Where("id IN ?", allIDs).Delete(&nesma.NesmaRequirement{}).Error; err != nil {
			return fmt.Errorf("删除需求失败: %v", err)
		}

		global.GVA_LOG.Info("根据条件删除需求成功",
			zap.Uint("projectId", req.ProjectID),
			zap.Uint("cycleId", req.CycleID),
			zap.Uint("versionId", req.VersionID),
			zap.Int("deletedCount", len(allIDs)))

		return nil
	})
}

// collectChildRequirements 递归收集子需求ID
func (s *NesmaRequirementService) collectChildRequirements(parentIDs []uint, allIDs *[]uint) error {
	if len(parentIDs) == 0 {
		return nil
	}

	var childRequirements []nesma.NesmaRequirement
	if err := global.GVA_DB.Where("parent_id IN ?", parentIDs).Find(&childRequirements).Error; err != nil {
		return err
	}

	if len(childRequirements) == 0 {
		return nil
	}

	var childIDs []uint
	for _, child := range childRequirements {
		childIDs = append(childIDs, child.ID)
		*allIDs = append(*allIDs, child.ID)
	}

	// 递归查找下一级子需求
	return s.collectChildRequirements(childIDs, allIDs)
}

// GetNesmaRequirement 获取需求详情
func (s *NesmaRequirementService) GetNesmaRequirement(id uint) (*nesmaRes.NesmaRequirementResponse, error) {
	var requirement nesma.NesmaRequirement
	if err := global.GVA_DB.Preload("Project").Preload("Parent").First(&requirement, id).Error; err != nil {
		return nil, errors.New("需求不存在")
	}

	response := &nesmaRes.NesmaRequirementResponse{
		NesmaRequirement: requirement,
		LevelName:        requirement.GetLevelName(),
		FullPath:         requirement.GetFullPath(),
	}

	return response, nil
}

// GetNesmaRequirementList 获取需求列表
func (s *NesmaRequirementService) GetNesmaRequirementList(req *nesmaReq.NesmaRequirementSearch) (*nesmaRes.NesmaRequirementListResponse, error) {
	var requirements []nesma.NesmaRequirement
	var total int64

	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})

	// 条件筛选
	if req.ProjectID != nil {
		db = db.Where("project_id = ?", *req.ProjectID)
	}
	if req.ParentID != nil {
		db = db.Where("parent_id = ?", *req.ParentID)
	}
	if req.Level != nil {
		db = db.Where("level = ?", *req.Level)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.Title != "" {
		db = db.Where("title LIKE ?", "%"+req.Title+"%")
	}

	if req.VesionId != nil {
		db = db.Where("version_id = ?", *req.VesionId)
	}

	if req.Description != "" {
		db = db.Where("description LIKE ?", "%"+req.Description+"%")
	}

	

	// 获取总数
	db.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := db.Preload("Project").Preload("Parent").
		Order("order_index ASC, created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&requirements).Error; err != nil {
		return nil, err
	}

	// 转换响应格式
	var list []nesmaRes.NesmaRequirementResponse
	for _, requirement := range requirements {
		list = append(list, nesmaRes.NesmaRequirementResponse{
			NesmaRequirement: requirement,
			LevelName:        requirement.GetLevelName(),
			FullPath:         requirement.GetFullPath(),
		})
	}

	return &nesmaRes.NesmaRequirementListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// GetRequirementTreeByProject 获取项目需求树形结构（支持多种查询条件）
func (s *NesmaRequirementService) GetRequirementTreeByProject(projectID uint, queryConditions map[string]interface{}) ([]nesma.NesmaRequirement, int64, error) {
	global.GVA_LOG.Info("获取项目需求树", 
		zap.Uint("projectID", projectID),
		zap.Any("conditions", queryConditions))

	var requirements []nesma.NesmaRequirement
	var total int64

	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})

	// 基本项目ID筛选
	db = db.Where("project_id = ?", projectID)

	// 动态添加查询条件
	if cycleID, exists := queryConditions["cycle_id"]; exists && cycleID != nil {
		db = db.Where("cycle_id = ?", cycleID)
	}

	if versionID, exists := queryConditions["version_id"]; exists && versionID != nil {
		db = db.Where("version_id = ?", versionID)
	}

	if level, exists := queryConditions["level"]; exists && level != nil {
		db = db.Where("level = ?", level)
	}

	if status, exists := queryConditions["status"]; exists && status != nil && status != "" {
		db = db.Where("status = ?", status)
	}

	if keyword, exists := queryConditions["keyword"]; exists && keyword != nil && keyword != "" {
		keywordStr := keyword.(string)
		db = db.Where("title LIKE ? OR description LIKE ? OR code LIKE ?",
			"%"+keywordStr+"%", "%"+keywordStr+"%", "%"+keywordStr+"%")
	}

	// 获取总数
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if err := db.Preload("Project").Preload("Parent").Preload("Children").
		Order("level ASC, order_index ASC, created_at ASC").
		Find(&requirements).Error; err != nil {
		return nil, 0, err
	}

	return requirements, total, nil
}

// GetNesmaRequirementTree 获取需求树形结构
func (s *NesmaRequirementService) GetNesmaRequirementTree(projectID uint) ([]nesmaRes.NesmaRequirementTreeResponse, error) {
	var requirements []nesma.NesmaRequirement
	if err := global.GVA_DB.Where("project_id = ?", projectID).
		Order("level ASC, order_index ASC, created_at ASC").
		Find(&requirements).Error; err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildRequirementTree(requirements, nil), nil
}

// GetNesmaRequirementTreeWithFilter 获取需求树形结构（带搜索条件）
func (s *NesmaRequirementService) GetNesmaRequirementTreeWithFilter(req *nesmaReq.NesmaRequirementTreeSearch) ([]nesmaRes.NesmaRequirementTreeResponse, error) {
	global.GVA_LOG.Info("开始获取需求树（带过滤条件）",
		zap.Any("projectID", req.ProjectID),
		zap.Any("level", req.Level),
		zap.String("status", req.Status),
		zap.String("keyword", req.Keyword))

	var requirements []nesma.NesmaRequirement

	// 如果有关键词搜索，需要特殊处理以包含上级节点
	if req.Keyword != "" {
		global.GVA_LOG.Info("使用关键词搜索分支")
		return s.getRequirementTreeWithKeywordSearch(req)
	}

	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})

	// 项目ID筛选
	if req.ProjectID != nil {
		db = db.Where("project_id = ?", *req.ProjectID)
	}

	// 层级筛选
	if req.Level != nil {
		db = db.Where("level = ?", *req.Level)
	}

	// 状态筛选
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	if err := db.Order("level ASC, order_index ASC, created_at ASC").
		Find(&requirements).Error; err != nil {
		return nil, err
	} 

	// 构建树形结构
	return s.buildRequirementTree(requirements, nil), nil
}

// getRequirementTreeWithKeywordSearch 关键词搜索时获取需求树（包含上级节点）
func (s *NesmaRequirementService) getRequirementTreeWithKeywordSearch(req *nesmaReq.NesmaRequirementTreeSearch) ([]nesmaRes.NesmaRequirementTreeResponse, error) {
	global.GVA_LOG.Info("开始关键词搜索", zap.String("keyword", req.Keyword))

	// 首先查找所有匹配关键词的需求
	var matchedRequirements []nesma.NesmaRequirement
	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})

	if req.ProjectID != nil {
		db = db.Where("project_id = ?", *req.ProjectID)
	}

	// 关键词筛选
	db = db.Where("title LIKE ? OR description LIKE ? OR code LIKE ?",
		"%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")

	// 层级和状态筛选
	if req.Level != nil {
		db = db.Where("level = ?", *req.Level)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	if err := db.Find(&matchedRequirements).Error; err != nil {
		global.GVA_LOG.Error("关键词搜索查询失败", zap.Error(err))
		return nil, err
	}

	global.GVA_LOG.Info("关键词匹配结果", zap.Int("count", len(matchedRequirements)))

	if len(matchedRequirements) == 0 {
		return []nesmaRes.NesmaRequirementTreeResponse{}, nil
	}

	// 收集所有需要的ID（包括匹配项和其上级节点）
	neededIDs := make(map[uint]bool)

	// 添加匹配的需求ID
	for _, req := range matchedRequirements {
		neededIDs[req.ID] = true
	}

	// 查找并添加所有上级节点
	for _, req := range matchedRequirements {
		if err := s.addParentIDs(req.ParentID, neededIDs); err != nil {
			return nil, err
		}
	}

	// 将map转换为slice
	var ids []uint
	for id := range neededIDs {
		ids = append(ids, id)
	}

	// 查询所有需要的需求（包括匹配项和上级节点）
	var allRequirements []nesma.NesmaRequirement
	if err := global.GVA_DB.Where("id IN ?", ids).
		Order("level ASC, order_index ASC, created_at ASC").
		Find(&allRequirements).Error; err != nil {
		return nil, err
	}

	// 构建树形结构
	return s.buildRequirementTree(allRequirements, nil), nil
}

// addParentIDs 递归添加父级节点ID
func (s *NesmaRequirementService) addParentIDs(parentID *uint, neededIDs map[uint]bool) error {
	if parentID == nil {
		return nil
	}

	// 如果已经添加过，避免重复查询
	if neededIDs[*parentID] {
		return nil
	}

	// 添加父级ID
	neededIDs[*parentID] = true

	// 查询父级节点
	var parent nesma.NesmaRequirement
	err := global.GVA_DB.First(&parent, *parentID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 忽略不存在的父级节点，但记录日志
			global.GVA_LOG.Warn("父级节点不存在", zap.Uint("parentID", *parentID))
			return nil
		}
		global.GVA_LOG.Error("查询父级节点失败", zap.Uint("parentID", *parentID), zap.Error(err))
		return err
	}

	// 递归添加祖父级节点
	return s.addParentIDs(parent.ParentID, neededIDs)
}

// buildRequirementTree 构建需求树
func (s *NesmaRequirementService) buildRequirementTree(requirements []nesma.NesmaRequirement, parentID *uint) []nesmaRes.NesmaRequirementTreeResponse {
	var tree []nesmaRes.NesmaRequirementTreeResponse

	for _, req := range requirements {
		if (parentID == nil && req.ParentID == nil) || (parentID != nil && req.ParentID != nil && *req.ParentID == *parentID) {
			node := nesmaRes.NesmaRequirementTreeResponse{
				NesmaRequirement: req,
				LevelName:        req.GetLevelName(),
				FullPath:         req.GetFullPath(),
				Children:         s.buildRequirementTree(requirements, &req.ID),
			}
			tree = append(tree, node)
		}
	}

	return tree
}

// GetNesmaRequirementStats 获取需求统计
func (s *NesmaRequirementService) GetNesmaRequirementStats(projectID *uint) (*nesmaRes.NesmaRequirementStatsResponse, error) {
	var totalCount int64
	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})
	
	// 如果指定了项目ID，则按项目过滤；否则获取全部项目统计
	if projectID != nil {
		db = db.Where("project_id = ?", *projectID)
	}
	db.Count(&totalCount)

	// 层级统计
	var levelStats []nesmaRes.RequirementLevelStats
	levelQuery := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Select("level, COUNT(*) as count")
	if projectID != nil {
		levelQuery = levelQuery.Where("project_id = ?", *projectID)
	}
	rows, err := levelQuery.Group("level").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var level int
		var count int64
		rows.Scan(&level, &count)
		levelName := ""
		switch level {
		case 1:
			levelName = "一级功能模块"
		case 2:
			levelName = "二级功能模块"
		case 3:
			levelName = "三级功能模块"
		case 4:
			levelName = "功能点计数项"
		}
		levelStats = append(levelStats, nesmaRes.RequirementLevelStats{
			Level:     level,
			LevelName: levelName,
			Count:     count,
		})
	}

	// 状态统计
	var statusStats []nesmaRes.RequirementStatusStats
	statusQuery := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Select("status, COUNT(*) as count")
	if projectID != nil {
		statusQuery = statusQuery.Where("project_id = ?", *projectID)
	}
	rows, err = statusQuery.Group("status").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int64
		rows.Scan(&status, &count)
		statusStats = append(statusStats, nesmaRes.RequirementStatusStats{
			Status: status,
			Count:  count,
		})
	}

	// 优先级统计
	var priorityStats []nesmaRes.RequirementPriorityStats
	priorityQuery := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Select("priority, COUNT(*) as count")
	if projectID != nil {
		priorityQuery = priorityQuery.Where("project_id = ?", *projectID)
	}
	rows, err = priorityQuery.Group("priority").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var priority int
		var count int64
		rows.Scan(&priority, &count)
		priorityStats = append(priorityStats, nesmaRes.RequirementPriorityStats{
			Priority: priority,
			Count:    count,
		})
	}

	// 复杂度统计
	var complexityStats []nesmaRes.RequirementComplexityStats
	complexityQuery := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Select("complexity, COUNT(*) as count").
		Where("complexity != ''")
	if projectID != nil {
		complexityQuery = complexityQuery.Where("project_id = ?", *projectID)
	}
	rows, err = complexityQuery.Group("complexity").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var complexity string
		var count int64
		rows.Scan(&complexity, &count)
		complexityStats = append(complexityStats, nesmaRes.RequirementComplexityStats{
			Complexity: complexity,
			Count:      count,
		})
	}

	return &nesmaRes.NesmaRequirementStatsResponse{
		ProjectID:       projectID,
		TotalCount:      totalCount,
		LevelStats:      levelStats,
		StatusStats:     statusStats,
		PriorityStats:   priorityStats,
		ComplexityStats: complexityStats,
	}, nil
}

// GetParentRequirementOptions 获取父需求选项
func (s *NesmaRequirementService) GetParentRequirementOptions(projectID uint, level int) (*nesmaRes.ParentRequirementOptionsResponse, error) {
	var requirements []nesma.NesmaRequirement

	// 只能选择比当前层级小一级的需求作为父级（子需求只能选择父需求，不能选择爷爷）
	parentLevel := level - 1
	if parentLevel < 1 {
		// 如果是一级需求，没有父级选项
		return &nesmaRes.ParentRequirementOptionsResponse{
			Options: []nesmaRes.ParentRequirementOption{},
		}, nil
	}

	if err := global.GVA_DB.Where("project_id = ? AND level = ?", projectID, parentLevel).
		Order("level ASC, order_index ASC, created_at ASC").
		Find(&requirements).Error; err != nil {
		return nil, err
	}

	var options []nesmaRes.ParentRequirementOption
	for _, req := range requirements {
		options = append(options, nesmaRes.ParentRequirementOption{
			ID:       req.ID,
			Title:    req.Title,
			Level:    req.Level,
			FullPath: req.GetFullPath(),
		})
	}

	return &nesmaRes.ParentRequirementOptionsResponse{
		Options: options,
	}, nil
}

// BatchDeleteNesmaRequirements 批量删除需求
func (s *NesmaRequirementService) BatchDeleteNesmaRequirements(req *nesmaReq.NesmaRequirementIdsRequest) error {
	// 检查是否有子需求
	for _, id := range req.IDs {
		var childCount int64
		global.GVA_DB.Model(&nesma.NesmaRequirement{}).Where("parent_id = ?", id).Count(&childCount)
		if childCount > 0 {
			return errors.New("存在子需求，无法删除")
		}
	}

	return global.GVA_DB.Delete(&nesma.NesmaRequirement{}, req.IDs).Error
}

// ImportFromExcel 从Excel导入需求
func (s *NesmaRequirementService) ImportFromExcel(filePath string, req *nesmaReq.ImportNesmaRequirementRequest) (*nesmaRes.ImportResultResponse, error) {
	file, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("Excel文件中没有工作表")
	}

	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}

	if len(rows) < 2 {
		return nil, errors.New("Excel文件中没有数据行")
	}

	// 生成导入批次号
	importBatch := req.ImportBatch
	if importBatch == "" {
		importBatch = fmt.Sprintf("import_%d", time.Now().Unix())
	}

	result := &nesmaRes.ImportResultResponse{
		TotalRows:   len(rows) - 1, // 排除标题行
		ImportBatch: importBatch,
		Errors:      []string{},
	}

	// 存储需求映射，key为标题，value为需求ID
	requirementMap := make(map[string]uint)

	// 维护当前的父级模块信息（处理合并单元格）
	var currentLevel1, currentLevel2, currentLevel3 string
	var version *nesma.NesmaRequirementVersion

	if req.VersionAction == "new" {
			// 创建版本记录
		version = &nesma.NesmaRequirementVersion{
		CycleID:     req.CycleID,
		Version:     req.VersionName,
		VersionType: "initial",
		CreatedBy:   "import",
		Summary:     fmt.Sprintf("从Excel导入，批次：%s", importBatch),
		}
		if err := global.GVA_DB.Create(version).Error; err != nil {
			return nil, fmt.Errorf("创建版本记录失败: %v", err)
		}
	} else {
		// 获取版本记录
		if err := global.GVA_DB.Where("cycle_id = ? AND version = ?", req.CycleID, req.VersionName).First(&version).Error; err != nil {
			return nil, fmt.Errorf("获取版本记录失败: %v", err)
		}
	}



	// 事务处理
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i, row := range rows[1:] { // 跳过标题行
		rowNum := i + 2

		// 确保行有足够的列，不够的补空字符串（防止数据错位）
		for len(row) < 9 {
			row = append(row, "")
		}

		// 检查是否是表头行（只检测明确的表头关键词）
		if len(row) > 3 && (row[3] == "一级模块" || row[3] == "一级功能模块") {
			continue // 跳过表头行
		}

		// 检查是否至少有一个模块列或功能点列有数据
		hasValidData := false
		// 检查模块列（3-5列）和功能点列（6列）
		for j := 3; j <= 6; j++ {
			if j < len(row) && row[j] != "" {
				hasValidData = true
				break
			}
		}
		if !hasValidData {
			continue // 跳过没有有效数据的行
		}

		// 获取基本信息（确保使用正确的列索引）
		code := ""
		if len(row) > 0 {
			code = row[0]
		}
		category := ""
		if len(row) > 1 {
			category = row[1]
		}
		// row[2] 是子系统，保留
		
		// 对于1-3级模块，字段列作为描述；对于4级功能点，字段列作为标题
		fieldDescription := ""
		if len(row) > 6 {
			fieldDescription = row[6] // 字段列
		}
		notes := ""
		if len(row) > 7 {
			notes = row[7] // 功能设计&数据库列（作为备注）
		}

		// 更新当前模块信息（处理合并单元格）
		if len(row) > 3 && row[3] != "" {
			currentLevel1 = row[3]
		}
		if len(row) > 4 && row[4] != "" {
			currentLevel2 = row[4]
		} else if len(row) > 3 && row[3] != "" {
			// 如果一级模块有值但二级模块为空，重置二级模块
			currentLevel2 = ""
		}
		if len(row) > 5 && row[5] != "" {
			currentLevel3 = row[5]
		} else if (len(row) > 4 && row[4] != "") || (len(row) > 3 && row[3] != "") {
			// 如果上级模块有值但三级模块为空，重置三级模块
			currentLevel3 = ""
		}

		// 创建一级需求（只有当前行有一级模块数据时才创建）
		if len(row) > 3 && row[3] != "" {
			err := s.createOrUpdateRequirement(tx, 1, currentLevel1, "", code, category, fieldDescription, notes, req, importBatch, i*10, requirementMap, version.ID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("第%d行创建一级需求失败: %v", rowNum, err))
				result.FailedRows++
				continue
			}
		}

		// 创建二级需求（只有当前行有二级模块数据时才创建）
		if len(row) > 4 && row[4] != "" && currentLevel1 != "" {
			err := s.createOrUpdateRequirement(tx, 2, currentLevel2, currentLevel1, code, category, fieldDescription, notes, req, importBatch, i*10+1, requirementMap, version.ID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("第%d行创建二级需求失败: %v", rowNum, err))
				result.FailedRows++
				continue
			}
		}

		// 创建三级需求（只有当前行有三级模块数据时才创建）
		if len(row) > 5 && row[5] != "" && currentLevel2 != "" {
			err := s.createOrUpdateRequirement(tx, 3, currentLevel3, currentLevel2, code, category, fieldDescription, notes, req, importBatch, i*10+2, requirementMap, version.ID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("第%d行创建三级需求失败: %v", rowNum, err))
				result.FailedRows++
				continue
			}
		}

		// 创建四级功能点（功能点计数项）- 字段列作为功能点
		if len(row) > 6 && row[6] != "" {
			// 确定父级需求
			var parentTitle string
			if currentLevel3 != "" {
				parentTitle = currentLevel3
			} else if currentLevel2 != "" {
				parentTitle = currentLevel2
			} else if currentLevel1 != "" {
				parentTitle = currentLevel1
			}

			global.GVA_LOG.Info("准备创建四级功能点", 
				zap.String("functionPointTitle", row[6]),
				zap.String("parentTitle", parentTitle),
				zap.String("currentLevel1", currentLevel1),
				zap.String("currentLevel2", currentLevel2),
				zap.String("currentLevel3", currentLevel3))

			if parentTitle != "" {
				// 使用字段列作为功能点标题，notes作为功能点描述
				functionPointTitle := row[6]
				functionPointDescription := notes // 使用notes作为功能点描述
				err := s.createOrUpdateRequirement(tx, 4, functionPointTitle, parentTitle, code, category, functionPointDescription, notes, req, importBatch, i*10+3, requirementMap, version.ID)
				if err != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("第%d行创建功能点失败: %v", rowNum, err))
					result.FailedRows++
					continue
				}
			} else {
				global.GVA_LOG.Warn("四级功能点没有父级需求", 
					zap.String("functionPointTitle", row[6]),
					zap.Int("rowNum", rowNum))
			}
		}

		result.SuccessRows++
	}

	if len(result.Errors) > 0 {
		result.Success = false
	} else {
		result.Success = true
	}

	if result.FailedRows > 0 && result.SuccessRows == 0 {
		tx.Rollback()
		return result, errors.New("导入失败，所有数据都有问题")
	}

	tx.Commit()
	return result, nil
}

// createOrUpdateRequirement 创建或更新需求的辅助方法
func (s *NesmaRequirementService) createOrUpdateRequirement(tx *gorm.DB, level int, title, parentTitle, code, category, description, notes string, req *nesmaReq.ImportNesmaRequirementRequest, importBatch string, orderIndex int, requirementMap map[string]uint, versionID uint) error {
	if title == "" {
		return errors.New("需求标题不能为空")
	}

	// 查找父级需求ID
	var parentID *uint
	if parentTitle != "" {
		if parentReqID, exists := requirementMap[parentTitle]; exists {
			parentID = &parentReqID
		} else {
			return fmt.Errorf("找不到父级需求: %s", parentTitle)
		}
	}

	// 检查是否已存在同名需求（使用新的唯一性约束）
	var existingReq nesma.NesmaRequirement
	err := tx.Where("project_id = ? AND level = ? AND title = ? AND cycle_id = ? AND version_id = ?", 
		req.ProjectID, level, title, req.CycleID, versionID).First(&existingReq).Error

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fmt.Errorf("查询需求失败: %v", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 不存在，创建新需求
		newReq := nesma.NesmaRequirement{
			ProjectID:    req.ProjectID,
			ParentID:     parentID,
			Level:        level,
			Code:         code,
			Title:        title,
			Description:  description,
			Priority:     3,
			Status:       "pending",
			Category:     category,
			Complexity:   "中等",
			Notes:        notes,
			ImportBatch:  importBatch,
			ImportSource: req.ImportSource,
			OrderIndex:   orderIndex,
			// 新增字段设置
			CycleID:          &req.CycleID,
			VersionID:        &versionID,
			AFP:              0,
			UFP:              0,
			FunctionType:     "",
			ReuseLevel:       "中",
			ModificationType: "新增",
		}

		if err := tx.Create(&newReq).Error; err != nil {
			return fmt.Errorf("创建需求失败: %v", err)
		}

		requirementMap[title] = newReq.ID
	} else {
		// 已存在，更新信息
		updates := map[string]interface{}{
			"code":         code,
			"description":  description,
			"category":     category,
			"notes":        notes,
			"import_batch": importBatch,
			"parent_id":    parentID,
		}

		if err := tx.Model(&existingReq).Updates(updates).Error; err != nil {
			return fmt.Errorf("更新需求失败: %v", err)
		}

		requirementMap[title] = existingReq.ID
	}

	return nil
}

// parseExcelRow 解析Excel行数据，新格式：序号类别、项目、子系统、一级模块、二级模块、三级模块、字段、功能设计&数据库
func (s *NesmaRequirementService) parseExcelRow(row []string, rowIndex int) (*nesmaReq.ExcelRequirementRow, error) {
	if len(row) < 7 {
		return nil, errors.New("数据列数不足，需要至少7列")
	}

	excelRow := &nesmaReq.ExcelRequirementRow{}

	// 从模块列判断层级和标题
	// 列索引：0-序号类别, 1-项目, 2-子系统, 3-一级模块, 4-二级模块, 5-三级模块, 6-字段, 7-功能设计&数据库

	var level int
	var title string
	var parentTitle string

	// 检查三级模块
	if len(row) > 5 && row[5] != "" {
		level = 3
		title = row[5]
		if len(row) > 4 && row[4] != "" {
			parentTitle = row[4] // 二级模块作为父级
		}
	} else if len(row) > 4 && row[4] != "" {
		// 检查二级模块
		level = 2
		title = row[4]
		if len(row) > 3 && row[3] != "" {
			parentTitle = row[3] // 一级模块作为父级
		}
	} else if len(row) > 3 && row[3] != "" {
		// 检查一级模块
		level = 1
		title = row[3]
		// 一级模块没有父级
	} else {
		return nil, errors.New("无法确定需求层级，模块列都为空")
	}

	if title == "" {
		return nil, errors.New("标题不能为空")
	}

	excelRow.Level = level
	excelRow.Title = title
	excelRow.ParentTitle = parentTitle

	// 描述（字段列）
	if len(row) > 6 {
		excelRow.Description = row[6]
	}

	// 编号（使用序号类别）
	if len(row) > 0 {
		excelRow.Code = row[0]
	}

	// 分类（使用项目列）
	if len(row) > 1 {
		excelRow.Category = row[1]
	}

	// 备注（使用功能设计&数据库列）
	if len(row) > 7 {
		excelRow.Notes = row[7]
	}

	// 设置默认值
	excelRow.Priority = 3
	excelRow.Complexity = "中等"

	return excelRow, nil
}

// parseRowRequirements 解析一行数据中的所有需求（处理层级结构）
func (s *NesmaRequirementService) parseRowRequirements(row []string, rowNum int, projectID uint, importBatch, importSource string, orderIndex int) ([]nesma.NesmaRequirement, error) {
	var requirements []nesma.NesmaRequirement

	// 列索引：0-序号类别, 1-项目, 2-子系统, 3-一级模块, 4-二级模块, 5-三级模块, 6-字段, 7-功能设计&数据库

	// 基本信息
	code := ""
	if len(row) > 0 {
		code = row[0]
	}

	category := ""
	if len(row) > 1 {
		category = row[1]
	}

	description := ""
	if len(row) > 6 {
		description = row[6]
	}

	notes := ""
	if len(row) > 7 {
		notes = row[7]
	}

	// 处理一级模块
	if len(row) > 3 && row[3] != "" {
		req := nesma.NesmaRequirement{
			ProjectID:    projectID,
			Level:        1,
			Code:         code,
			Title:        row[3],
			Description:  description,
			Priority:     3,
			Status:       "pending",
			Category:     category,
			Complexity:   "中等",
			Notes:        notes,
			ImportBatch:  importBatch,
			ImportSource: importSource,
			// 新增字段设置默认值
			ConstructionPeriod: "默认周期",
			Version:            0,
			AFP:                0,
			UFP:                0,
			FunctionType:       "",
			ReuseLevel:         "中",
			ModificationType:   "新增",
			OrderIndex:         orderIndex * 10, // 给子级留空间
		}
		requirements = append(requirements, req)
	}

	// 处理二级模块
	if len(row) > 4 && row[4] != "" {
		req := nesma.NesmaRequirement{
			ProjectID:    projectID,
			Level:        2,
			Code:         code,
			Title:        row[4],
			Description:  description,
			Priority:     3,
			Status:       "pending",
			Category:     category,
			Complexity:   "中等",
			Notes:        notes,
			ImportBatch:  importBatch,
			ImportSource: importSource,
			// 新增字段设置默认值
			ConstructionPeriod: "默认周期",
			Version:            0,
			AFP:                0,
			UFP:                0,
			FunctionType:       "",
			ReuseLevel:         "中",
			ModificationType:   "新增",
			OrderIndex:         orderIndex*10 + 1,
		}
		requirements = append(requirements, req)
	}

	// 处理三级模块
	if len(row) > 5 && row[5] != "" {
		req := nesma.NesmaRequirement{
			ProjectID:    projectID,
			Level:        3,
			Code:         code,
			Title:        row[5],
			Description:  description,
			Priority:     3,
			Status:       "pending",
			Category:     category,
			Complexity:   "中等",
			Notes:        notes,
			ImportBatch:  importBatch,
			ImportSource: importSource,
			// 新增字段设置默认值
			ConstructionPeriod: "默认周期",
			Version:            0,
			AFP:                0,
			UFP:                0,
			FunctionType:       "",
			ReuseLevel:         "中",
			ModificationType:   "新增",
			OrderIndex:         orderIndex*10 + 2,
		}
		requirements = append(requirements, req)
	}

	if len(requirements) == 0 {
		return nil, errors.New("无法解析任何需求，模块列都为空")
	}

	return requirements, nil
}

// getParentTitle 获取指定层级的父级标题
func (s *NesmaRequirementService) getParentTitle(row []string, level int) string {
	// 列索引：0-序号类别, 1-项目, 2-子系统, 3-一级模块, 4-二级模块, 5-三级模块, 6-字段, 7-功能设计&数据库
	switch level {
	case 1:
		if len(row) > 3 {
			return row[3]
		}
	case 2:
		if len(row) > 4 {
			return row[4]
		}
	case 3:
		if len(row) > 5 {
			return row[5]
		}
	}
	return ""
}

// BatchUpdateOrder 批量更新排序
func (s *NesmaRequirementService) BatchUpdateOrder(req *nesmaReq.BatchUpdateOrderRequest) error {
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for _, item := range req.Items {
		if err := tx.Model(&nesma.NesmaRequirement{}).
			Where("id = ?", item.ID).
			Update("order_index", item.OrderIndex).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit().Error
}

// MoveRequirement 移动需求
func (s *NesmaRequirementService) MoveRequirement(req *nesmaReq.MoveRequirementRequest) error {
	var requirement nesma.NesmaRequirement
	if err := global.GVA_DB.First(&requirement, req.ID).Error; err != nil {
		return errors.New("需求不存在")
	}

	// 验证父级需求
	if req.ParentID != nil {
		var parent nesma.NesmaRequirement
		if err := global.GVA_DB.First(&parent, *req.ParentID).Error; err != nil {
			return errors.New("父级需求不存在")
		}
		// 验证项目一致性
		if requirement.ProjectID != parent.ProjectID {
			return errors.New("不能移动到其他项目")
		}
	}

	updates := map[string]interface{}{
		"parent_id":   req.ParentID,
		"order_index": req.Position,
	}

	return global.GVA_DB.Model(&requirement).Updates(updates).Error
}

// GetProjectMaxVersion 获取项目最大版本号
// 重构后的逻辑：先获取项目当前活跃周期，再从需求版本表中获取最大版本号
func (s *NesmaRequirementService) GetProjectMaxVersion(projectID uint) (float64, error) {
	// 1. 先获取项目的当前活跃周期
	cycleService := NesmaProjectCycleService{}
	activeCycle, err := cycleService.GetActiveCycle(projectID)
	if err != nil {
		global.GVA_LOG.Error("获取项目活跃周期失败", 
			zap.Uint("projectID", projectID), 
			zap.Error(err))
		return 0, fmt.Errorf("获取项目活跃周期失败: %v", err)
	}

	// 2. 根据周期ID到nesma_requirement_versions表中获取最大版本号
	var maxVersionStr string
	err = global.GVA_DB.Model(&nesma.NesmaRequirementVersion{}).
		Where("cycle_id = ?", activeCycle.ID).
		Select("COALESCE(MAX(version), 'v0.0')").
		Scan(&maxVersionStr).Error

	if err != nil {
		global.GVA_LOG.Error("查询最大版本号失败", 
			zap.Uint("cycleID", activeCycle.ID), 
			zap.Error(err))
		return 0, fmt.Errorf("查询最大版本号失败: %v", err)
	}

	// 3. 解析版本号字符串，提取主版本号
	// 支持的版本格式：v1.0, v1.1, v1.2, v2.0 等
	maxVersion := s.parseVersionNumber(maxVersionStr)
	
	global.GVA_LOG.Info("获取项目最大版本号成功", 
		zap.Uint("projectID", projectID),
		zap.Uint("cycleID", activeCycle.ID),
		zap.String("maxVersionStr", maxVersionStr),
		zap.Float64("maxVersion", maxVersion))

	return maxVersion, nil
}

// parseVersionNumber 解析版本号字符串，提取主版本号
// 支持格式：v1.0, v1.1, v2.0, 1.0, 1.1 等
func (s *NesmaRequirementService) parseVersionNumber(versionStr string) float64 {
	if versionStr == "" || versionStr == "v0.0" {
		return 0.0
	}

	// 去掉前缀v或V
	if versionStr[0] == 'v' || versionStr[0] == 'V' {
		versionStr = versionStr[1:]
	}

	// 直接尝试解析为浮点数
	if val, err := strconv.ParseFloat(versionStr, 64); err == nil {
		return val
	}

	global.GVA_LOG.Warn("无法解析版本号，使用默认值0", 
		zap.String("versionStr", versionStr))
	return 0
}

// ==================== AI分析展示相关服务方法 ====================

// GetRequirementAIAnalysis 获取需求AI分析结果
func (s *NesmaRequirementService) GetRequirementAIAnalysis(requirementID uint) (*nesmaRes.RequirementAIAnalysisResponse, error) {
	var requirement nesma.NesmaRequirement
	if err := global.GVA_DB.First(&requirement, requirementID).Error; err != nil {
		return nil, errors.New("需求不存在")
	}

	// 查询相似需求
	var similarRequirements []nesmaRes.SimilarRequirement
	if requirement.SimilarRequirements != nil {
		// 解析JSON数据
		// 这里简化处理，实际应该根据项目需求实现相似度计算
		similarRequirements = []nesmaRes.SimilarRequirement{
			{
				ID:          requirement.ID + 1,
				Title:       "相似需求示例",
				Description: "这是一个相似需求的示例",
				Similarity:  0.85,
				ProjectName: "相关项目",
			},
		}
	}

	var analysisTime *string
	if requirement.AIAnalysisTime != nil {
		timeStr := requirement.AIAnalysisTime.Format("2006-01-02 15:04:05")
		analysisTime = &timeStr
	}

	response := &nesmaRes.RequirementAIAnalysisResponse{
		RequirementID:       requirement.ID,
		OriginalTitle:       requirement.Title,
		OriginalDescription: requirement.Description,
		AIAnalysisStatus:    requirement.AIAnalysisStatus,
		AIGeneratedTitle:    requirement.AIGeneratedTitle,
		AIDescription:       requirement.AIDescription,
		AIComplexityScore:   requirement.AIComplexityScore,
		AIConfidenceScore:   requirement.AIConfidenceScore,
		RecommendedAFP:      requirement.RecommendedAFP,
		RecommendedUFP:      requirement.RecommendedUFP,
		FunctionType:        requirement.FunctionType,
		AIAnalysisTime:      analysisTime,
		OptimizationNotes:   requirement.Notes,
		SimilarRequirements: similarRequirements,
	}

	return response, nil
}

// GetProjectAIAnalysisStats 获取项目AI分析统计
func (s *NesmaRequirementService) GetProjectAIAnalysisStats(projectID uint, cycleID *uint) (*nesmaRes.ProjectAIAnalysisStats, error) {
	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})
	
	// 项目过滤
	db = db.Where("project_id = ?", projectID)
	
	// 周期过滤
	if cycleID != nil {
		db = db.Where("cycle_id = ?", *cycleID)
	}

	// 总需求数
	var totalRequirements int64
	db.Count(&totalRequirements)

	// 各状态统计
	var statusStats []nesmaRes.AIAnalysisStatusStats
	rows, err := db.Select("ai_analysis_status, COUNT(*) as count").
		Group("ai_analysis_status").Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var analyzedCount, pendingCount, analyzingCount, failedCount int64
	for rows.Next() {
		var status string
		var count int64
		rows.Scan(&status, &count)
		
		percentage := float64(count) / float64(totalRequirements) * 100
		statusStats = append(statusStats, nesmaRes.AIAnalysisStatusStats{
			Status:     status,
			Count:      count,
			Percentage: percentage,
		})

		switch status {
		case "completed":
			analyzedCount = count
		case "pending":
			pendingCount = count
		case "analyzing":
			analyzingCount = count
		case "failed":
			failedCount = count
		}
	}

	// 层级分布统计
	var levelStats []nesmaRes.AIAnalysisLevelStats
	levelRows, err := db.Select("level, COUNT(*) as total, SUM(CASE WHEN ai_analysis_status = 'completed' THEN 1 ELSE 0 END) as analyzed").
		Group("level").Rows()
	if err != nil {
		return nil, err
	}
	defer levelRows.Close()

	for levelRows.Next() {
		var level int
		var total, analyzed int64
		levelRows.Scan(&level, &total, &analyzed)
		
		levelName := ""
		switch level {
		case 1:
			levelName = "一级功能模块"
		case 2:
			levelName = "二级功能模块"
		case 3:
			levelName = "三级功能模块"
		case 4:
			levelName = "功能点计数项"
		}
		
		percentage := float64(analyzed) / float64(total) * 100
		levelStats = append(levelStats, nesmaRes.AIAnalysisLevelStats{
			Level:      level,
			LevelName:  levelName,
			Total:      total,
			Analyzed:   analyzed,
			Percentage: percentage,
		})
	}

	// 复杂度分布统计
	var complexityStats []nesmaRes.AIComplexityStats
	complexityRanges := []struct {
		Range string
		Min   float64
		Max   float64
	}{
		{"0-20", 0, 20},
		{"21-40", 21, 40},
		{"41-60", 41, 60},
		{"61-80", 61, 80},
		{"81-100", 81, 100},
	}

	for _, cr := range complexityRanges {
		var count int64
		db.Where("ai_complexity_score >= ? AND ai_complexity_score <= ?", cr.Min, cr.Max).Count(&count)
		percentage := float64(count) / float64(totalRequirements) * 100
		complexityStats = append(complexityStats, nesmaRes.AIComplexityStats{
			ComplexityRange: cr.Range,
			Count:           count,
			Percentage:      percentage,
		})
	}

	// 平均置信度
	var avgConfidence float64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("project_id = ? AND ai_confidence_score IS NOT NULL", projectID).
		Select("AVG(ai_confidence_score)").Scan(&avgConfidence)

	// 优化总结
	var optimizationSummary nesmaRes.AIOptimizationSummary
	var totalOptimizations, titleOptimizations, descriptionOptimizations, functionPointUpdates int64
	
	db.Where("ai_generated_title != '' OR ai_description != ''").Count(&totalOptimizations)
	db.Where("ai_generated_title != ''").Count(&titleOptimizations)
	db.Where("ai_description != ''").Count(&descriptionOptimizations)
	db.Where("recommended_afp IS NOT NULL OR recommended_ufp IS NOT NULL").Count(&functionPointUpdates)

	optimizationSummary = nesmaRes.AIOptimizationSummary{
		TotalOptimizations:       totalOptimizations,
		TitleOptimizations:       titleOptimizations,
		DescriptionOptimizations: descriptionOptimizations,
		FunctionPointUpdates:     functionPointUpdates,
		AverageImprovement:       75.5, // 模拟数据
	}

	analysisProgress := float64(analyzedCount) / float64(totalRequirements) * 100

	response := &nesmaRes.ProjectAIAnalysisStats{
		ProjectID:              projectID,
		CycleID:                cycleID,
		TotalRequirements:      totalRequirements,
		AnalyzedRequirements:   analyzedCount,
		PendingRequirements:    pendingCount,
		AnalyzingRequirements:  analyzingCount,
		FailedRequirements:     failedCount,
		AnalysisProgress:       analysisProgress,
		StatusDistribution:     statusStats,
		LevelDistribution:      levelStats,
		ComplexityDistribution: complexityStats,
		AverageConfidence:      avgConfidence,
		OptimizationSummary:    optimizationSummary,
	}

	return response, nil
}

// CompareRequirementVersions 对比需求版本
func (s *NesmaRequirementService) CompareRequirementVersions(req *nesmaReq.CompareRequirementVersionsRequest) (*nesmaRes.RequirementVersionCompareResponse, error) {
	// 查询两个版本的需求
	var requirement1, requirement2 nesma.NesmaRequirement
	if err := global.GVA_DB.Where("id = ? AND version_id = ?", req.RequirementID, req.VersionID1).First(&requirement1).Error; err != nil {
		return nil, errors.New("版本1需求不存在")
	}
	if err := global.GVA_DB.Where("id = ? AND version_id = ?", req.RequirementID, req.VersionID2).First(&requirement2).Error; err != nil {
		return nil, errors.New("版本2需求不存在")
	}

	// 构建版本信息
	version1 := nesmaRes.RequirementVersionInfo{
		VersionID:         req.VersionID1,
		Version:           fmt.Sprintf("v%d", req.VersionID1),
		Title:             requirement1.Title,
		Description:       requirement1.Description,
		AIAnalysisStatus:  requirement1.AIAnalysisStatus,
		AIGeneratedTitle:  requirement1.AIGeneratedTitle,
		AIDescription:     requirement1.AIDescription,
		AIComplexityScore: requirement1.AIComplexityScore,
		AIConfidenceScore: requirement1.AIConfidenceScore,
		RecommendedAFP:    requirement1.RecommendedAFP,
		RecommendedUFP:    requirement1.RecommendedUFP,
		FunctionType:      requirement1.FunctionType,
		CreatedAt:         requirement1.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	version2 := nesmaRes.RequirementVersionInfo{
		VersionID:         req.VersionID2,
		Version:           fmt.Sprintf("v%d", req.VersionID2),
		Title:             requirement2.Title,
		Description:       requirement2.Description,
		AIAnalysisStatus:  requirement2.AIAnalysisStatus,
		AIGeneratedTitle:  requirement2.AIGeneratedTitle,
		AIDescription:     requirement2.AIDescription,
		AIComplexityScore: requirement2.AIComplexityScore,
		AIConfidenceScore: requirement2.AIConfidenceScore,
		RecommendedAFP:    requirement2.RecommendedAFP,
		RecommendedUFP:    requirement2.RecommendedUFP,
		FunctionType:      requirement2.FunctionType,
		CreatedAt:         requirement2.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	// 计算差异
	var differences []nesmaRes.RequirementDifference
	
	// 标题差异
	if requirement1.Title != requirement2.Title {
		differences = append(differences, nesmaRes.RequirementDifference{
			Field:        "title",
			FieldName:    "需求标题",
			OldValue:     requirement1.Title,
			NewValue:     requirement2.Title,
			ChangeType:   "modified",
			Significance: "high",
		})
	}

	// 描述差异
	if requirement1.Description != requirement2.Description {
		differences = append(differences, nesmaRes.RequirementDifference{
			Field:        "description",
			FieldName:    "需求描述",
			OldValue:     requirement1.Description,
			NewValue:     requirement2.Description,
			ChangeType:   "modified",
			Significance: "medium",
		})
	}

	// AI分析状态差异
	if requirement1.AIAnalysisStatus != requirement2.AIAnalysisStatus {
		differences = append(differences, nesmaRes.RequirementDifference{
			Field:        "aiAnalysisStatus",
			FieldName:    "AI分析状态",
			OldValue:     requirement1.AIAnalysisStatus,
			NewValue:     requirement2.AIAnalysisStatus,
			ChangeType:   "modified",
			Significance: "low",
		})
	}

	// 功能点差异
	if requirement1.AFP != requirement2.AFP {
		differences = append(differences, nesmaRes.RequirementDifference{
			Field:        "afp",
			FieldName:    "调整后功能点",
			OldValue:     fmt.Sprintf("%.2f", requirement1.AFP),
			NewValue:     fmt.Sprintf("%.2f", requirement2.AFP),
			ChangeType:   "modified",
			Significance: "high",
		})
	}

	// 计算总结
	var highSignificance, mediumSignificance, lowSignificance int
	for _, diff := range differences {
		switch diff.Significance {
		case "high":
			highSignificance++
		case "medium":
			mediumSignificance++
		case "low":
			lowSignificance++
		}
	}

	overallChange := float64(len(differences)) * 10.0 // 简化的整体变化计算
	if overallChange > 100 {
		overallChange = 100
	}

	recommendAction := "accept"
	if highSignificance > 2 {
		recommendAction = "review"
	} else if len(differences) > 10 {
		recommendAction = "review"
	}

	summary := nesmaRes.VersionCompareSummary{
		TotalDifferences:   len(differences),
		HighSignificance:   highSignificance,
		MediumSignificance: mediumSignificance,
		LowSignificance:    lowSignificance,
		OverallChange:      overallChange,
		RecommendAction:    recommendAction,
	}

	response := &nesmaRes.RequirementVersionCompareResponse{
		RequirementID: req.RequirementID,
		Version1:      version1,
		Version2:      version2,
		Differences:   differences,
		Summary:       summary,
	}

	return response, nil
}

// UpdateRequirementAIStatus 更新需求AI分析状态
func (s *NesmaRequirementService) UpdateRequirementAIStatus(req *nesmaReq.UpdateRequirementAIStatusRequest) error {
	// 构建更新字段
	updates := map[string]interface{}{
		"ai_analysis_status": req.AIAnalysisStatus,
	}

	if req.AIDescription != "" {
		updates["ai_description"] = req.AIDescription
	}
	if req.AIGeneratedTitle != "" {
		updates["ai_generated_title"] = req.AIGeneratedTitle
	}
	if req.AIComplexityScore != nil {
		updates["ai_complexity_score"] = req.AIComplexityScore
	}
	if req.AIConfidenceScore != nil {
		updates["ai_confidence_score"] = req.AIConfidenceScore
	}
	if req.RecommendedAFP != nil {
		updates["recommended_afp"] = req.RecommendedAFP
	}
	if req.RecommendedUFP != nil {
		updates["recommended_ufp"] = req.RecommendedUFP
	}

	// 如果状态为完成，更新分析时间
	if req.AIAnalysisStatus == "completed" {
		updates["ai_analysis_time"] = time.Now()
	}

	// 批量更新
	if err := global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("id IN ?", req.RequirementIDs).
		Updates(updates).Error; err != nil {
		return fmt.Errorf("更新需求AI状态失败: %v", err)
	}

	global.GVA_LOG.Info("批量更新需求AI状态成功", 
		zap.Any("requirementIDs", req.RequirementIDs),
		zap.String("status", req.AIAnalysisStatus))

	return nil
}

// GetAIOptimizedRequirements 获取AI优化的需求列表
func (s *NesmaRequirementService) GetAIOptimizedRequirements(req *nesmaReq.GetAIOptimizedRequirementsRequest) (*nesmaRes.AIOptimizedRequirementsResponse, error) {
	var requirements []nesma.NesmaRequirement
	var total int64

	db := global.GVA_DB.Model(&nesma.NesmaRequirement{})

	// 基础过滤条件
	if req.ProjectID != nil {
		db = db.Where("project_id = ?", *req.ProjectID)
	}
	if req.CycleID != nil {
		db = db.Where("cycle_id = ?", *req.CycleID)
	}
	if req.VersionID != nil {
		db = db.Where("version_id = ?", *req.VersionID)
	}
	if req.Status != "" {
		db = db.Where("ai_analysis_status = ?", req.Status)
	}

	// 只查询有AI优化的需求
	db = db.Where("ai_analysis_status IN ('completed', 'analyzing') OR ai_generated_title != '' OR ai_description != ''")

	// 获取总数
	db.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := db.Preload("Project").Preload("Parent").
		Order("ai_analysis_time DESC, updated_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&requirements).Error; err != nil {
		return nil, err
	}

	// 转换为响应格式
	var list []nesmaRes.AIOptimizedRequirement
	for _, req := range requirements {
		optimizationChanges := []nesmaRes.OptimizationChange{}
		optimizationScore := 0.0
		hasOptimization := false

		// 检查标题优化
		if req.AIGeneratedTitle != "" && req.AIGeneratedTitle != req.Title {
			optimizationChanges = append(optimizationChanges, nesmaRes.OptimizationChange{
				Type:           "title",
				OriginalValue:  req.Title,
				OptimizedValue: req.AIGeneratedTitle,
				Improvement:    80.0,
			})
			optimizationScore += 80.0
			hasOptimization = true
		}

		// 检查描述优化
		if req.AIDescription != "" && req.AIDescription != req.Description {
			optimizationChanges = append(optimizationChanges, nesmaRes.OptimizationChange{
				Type:           "description",
				OriginalValue:  req.Description,
				OptimizedValue: req.AIDescription,
				Improvement:    70.0,
			})
			optimizationScore += 70.0
			hasOptimization = true
		}

		// 检查功能点优化
		if req.RecommendedAFP != nil && *req.RecommendedAFP != req.AFP {
			optimizationChanges = append(optimizationChanges, nesmaRes.OptimizationChange{
				Type:           "afp",
				OriginalValue:  fmt.Sprintf("%.2f", req.AFP),
				OptimizedValue: fmt.Sprintf("%.2f", *req.RecommendedAFP),
				Improvement:    60.0,
			})
			optimizationScore += 60.0
			hasOptimization = true
		}

		// 计算平均优化分数
		if len(optimizationChanges) > 0 {
			optimizationScore = optimizationScore / float64(len(optimizationChanges))
		}

		list = append(list, nesmaRes.AIOptimizedRequirement{
			NesmaRequirement:    req,
			LevelName:           req.GetLevelName(),
			FullPath:            req.GetFullPath(),
			HasAIOptimization:   hasOptimization,
			OptimizationScore:   optimizationScore,
			OptimizationChanges: optimizationChanges,
		})
	}

	// 计算总结统计
	var totalCount, optimizedCount, pendingCount, failedCount int64
	totalCount = total

	// 统计各状态数量
	if req.ProjectID != nil {
		baseQuery := global.GVA_DB.Model(&nesma.NesmaRequirement{}).Where("project_id = ?", *req.ProjectID)
		if req.CycleID != nil {
			baseQuery = baseQuery.Where("cycle_id = ?", *req.CycleID)
		}
		
		baseQuery.Where("ai_analysis_status = 'completed'").Count(&optimizedCount)
		baseQuery.Where("ai_analysis_status = 'pending'").Count(&pendingCount)
		baseQuery.Where("ai_analysis_status = 'failed'").Count(&failedCount)
	}

	optimizationRate := float64(optimizedCount) / float64(totalCount) * 100
	averageImprovement := 75.5 // 模拟数据

	var totalFunctionPoints, optimizedFunctionPoints float64
	if req.ProjectID != nil {
		global.GVA_DB.Model(&nesma.NesmaRequirement{}).
			Where("project_id = ?", *req.ProjectID).
			Select("SUM(afp)").Scan(&totalFunctionPoints)
		global.GVA_DB.Model(&nesma.NesmaRequirement{}).
			Where("project_id = ? AND recommended_afp IS NOT NULL", *req.ProjectID).
			Select("SUM(recommended_afp)").Scan(&optimizedFunctionPoints)
	}

	summary := nesmaRes.AIOptimizedSummary{
		TotalCount:              totalCount,
		OptimizedCount:          optimizedCount,
		PendingCount:            pendingCount,
		FailedCount:             failedCount,
		OptimizationRate:        optimizationRate,
		AverageImprovement:      averageImprovement,
		TotalFunctionPoints:     totalFunctionPoints,
		OptimizedFunctionPoints: optimizedFunctionPoints,
	}

	response := &nesmaRes.AIOptimizedRequirementsResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
		Summary:  summary,
	}

	return response, nil
}
