package nesma

import (
	"errors"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"gorm.io/gorm"
	"go.uber.org/zap"
)

type NesmaProjectService struct{}

// convertCustomTimeToTimePtr 将CustomTime转换为*time.Time
func convertCustomTimeToTimePtr(ct *nesma.CustomTime) *time.Time {
	if ct == nil || ct.Time.IsZero() {
		return nil
	}
	return &ct.Time
}

// CreateNesmaProject 创建NESMA项目
func (nesmaProjectService *NesmaProjectService) CreateNesmaProject(project *nesma.NesmaProject) (err error) {
	// 检查项目名称是否重复
	var count int64
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).Where("name = ? AND owner_id = ?", project.Name, project.OwnerID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("项目名称已存在")
	}

	// 设置默认值
	if project.Status == "" {
		project.Status = "active"
	}
	if project.StartDate == nil {
		now := time.Now()
		project.StartDate = &now
	}

	err = global.GVA_DB.Create(project).Error
	return err
}

// CreateNesmaProjectFromRequest 从请求创建NESMA项目
func (nesmaProjectService *NesmaProjectService) CreateNesmaProjectFromRequest(req *nesmaReq.NesmaProjectCreate, userID uint) (err error) {
	// 创建项目对象
	project := nesma.NesmaProject{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     userID,
		Domain:      req.Domain,
		DomainTags:  req.DomainTags,
		Settings:    req.Settings,
		StartDate:   convertCustomTimeToTimePtr(req.StartDate),
		EndDate:     convertCustomTimeToTimePtr(req.EndDate),
	}

	return nesmaProjectService.CreateNesmaProject(&project)
}

// DeleteNesmaProject 删除NESMA项目
func (nesmaProjectService *NesmaProjectService) DeleteNesmaProject(ID uint, userID uint) (err error) {
	// 检查项目是否存在且用户有权限
	var project nesma.NesmaProject
	err = global.GVA_DB.Where("id = ? AND owner_id = ?", ID, userID).First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在或无权限删除")
		}
		return err
	}

	// 项目状态如果为归档状态则不允许删除
	if project.Status == "archived" {
		return errors.New("归档项目不允许删除")
	}

	// 删除项目下的所有周期的所有版本，然后删除所有周期，最后删除项目
	// 先获取项目下的所有周期ID
	var cycleIDs []uint
	err = global.GVA_DB.Model(&nesma.NesmaProjectCycle{}).Where("project_id = ?", ID).Pluck("id", &cycleIDs).Error
	if err != nil {
		return err
	}
	
	// 删除这些周期下的所有版本
	if len(cycleIDs) > 0 {
		err = global.GVA_DB.Delete(&nesma.NesmaRequirementVersion{}, "cycle_id IN ?", cycleIDs).Error
		if err != nil {
			return err
		}
	}
	
	// 删除项目下的所有周期
	err = global.GVA_DB.Delete(&nesma.NesmaProjectCycle{}, "project_id = ?", ID).Error
	if err != nil {
		return err
	}
	
	// 删除项目下的所有需求
	err = global.GVA_DB.Delete(&nesma.NesmaRequirement{}, "project_id = ?", ID).Error
	if err != nil {
		return err
	}

	err = global.GVA_DB.Delete(&project).Error
	return err
}

// DeleteNesmaProjectByIds 批量删除NESMA项目
func (nesmaProjectService *NesmaProjectService) DeleteNesmaProjectByIds(IDs []uint, userID uint) (err error) {
	// 检查所有项目是否存在且用户有权限
	var count int64
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).Where("id IN ? AND owner_id = ?", IDs, userID).Count(&count).Error
	if err != nil {
		return err
	}
	if int(count) != len(IDs) {
		return errors.New("部分项目不存在或无权限删除")
	}

	// 检查项目下是否有需求
	var requirementCount int64
	err = global.GVA_DB.Model(&nesma.NesmaRequirement{}).Where("project_id IN ?", IDs).Count(&requirementCount).Error
	if err != nil {
		return err
	}
	if requirementCount > 0 {
		return errors.New("部分项目下存在需求，无法删除")
	}

	err = global.GVA_DB.Delete(&[]nesma.NesmaProject{}, "id IN ? AND owner_id = ?", IDs, userID).Error
	return err
}

// UpdateNesmaProject 更新NESMA项目
func (nesmaProjectService *NesmaProjectService) UpdateNesmaProject(project nesma.NesmaProject, userID uint) (err error) {
	// 检查项目是否存在且用户有权限
	var existingProject nesma.NesmaProject
	err = global.GVA_DB.Where("id = ? AND owner_id = ?", project.ID, userID).First(&existingProject).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在或无权限修改")
		}
		return err
	}

	// 检查项目名称是否重复（排除自己）
	var count int64
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).Where("name = ? AND owner_id = ? AND id != ?", project.Name, userID, project.ID).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("项目名称已存在")
	}

	// 更新项目信息
	err = global.GVA_DB.Model(&existingProject).Updates(map[string]interface{}{
		"name":        project.Name,
		"description": project.Description,
		"status":      project.Status,
		"domain":      project.Domain,
		"domain_tags": project.DomainTags,
		"settings":    project.Settings,
		"start_date":  project.StartDate,
		"end_date":    project.EndDate,
	}).Error
	return err
}

// UpdateNesmaProjectFromRequest 从请求更新NESMA项目
func (nesmaProjectService *NesmaProjectService) UpdateNesmaProjectFromRequest(req *nesmaReq.NesmaProjectUpdate, userID uint) (err error) {
	// 创建更新对象
	project := nesma.NesmaProject{
		GVA_MODEL:   global.GVA_MODEL{ID: req.ID},
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Domain:      req.Domain,
		DomainTags:  req.DomainTags,
		Settings:    req.Settings,
		StartDate:   convertCustomTimeToTimePtr(req.StartDate),
		EndDate:     convertCustomTimeToTimePtr(req.EndDate),
	}

	return nesmaProjectService.UpdateNesmaProject(project, userID)
}

// GetNesmaProject 根据ID获取NESMA项目
func (nesmaProjectService *NesmaProjectService) GetNesmaProject(ID uint, userID uint) (project nesma.NesmaProject, err error) {
	// 添加调试日志
	global.GVA_LOG.Info("正在获取项目详情", 
		zap.Uint("projectId", ID),
		zap.Uint("userId", userID))
	
	err = global.GVA_DB.Preload("Cycles").Preload("ActiveCycle").Where("id = ? AND owner_id = ?", ID, userID).First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return project, errors.New("项目不存在或无权限访问")
		}
		global.GVA_LOG.Error("获取项目详情失败", zap.Error(err))
		return project, err
	}
	
	// 添加调试日志：查看预加载结果
	global.GVA_LOG.Info("项目详情获取成功", 
		zap.Uint("projectId", project.ID),
		zap.String("projectName", project.Name),
		zap.Int("cyclesCount", len(project.Cycles)))
	
	return
}

// GetNesmaProjectInfoList 分页获取NESMA项目列表
func (nesmaProjectService *NesmaProjectService) GetNesmaProjectInfoList(info nesmaReq.NesmaProjectSearch, userID uint) (list []nesma.NesmaProject, total int64, err error) {
	limit := info.PageSize
	offset := info.PageSize * (info.Page - 1)

	// 构建查询条件
	db := global.GVA_DB.Model(&nesma.NesmaProject{}).Where("owner_id = ?", userID)

	// 添加搜索条件
	if info.Name != "" {
		db = db.Where("name LIKE ?", "%"+info.Name+"%")
	}
	if info.Status != "" {
		db = db.Where("status = ?", info.Status)
	}
	if info.Domain != "" {
		db = db.Where("domain = ?", info.Domain)
	}
	if info.StartDate != nil && info.EndDate != nil {
		db = db.Where("created_at BETWEEN ? AND ?", info.StartDate, info.EndDate)
	}

	// 获取总数
	err = db.Count(&total).Error
	if err != nil {
		return
	}

	// 获取列表数据，包含预加载的周期信息
	err = db.Preload("Cycles").Preload("ActiveCycle").Limit(limit).Offset(offset).Order("created_at DESC").Find(&list).Error
	return list, total, err
}

// GetProjectRequirementStats 获取项目的需求统计
func (nesmaProjectService *NesmaProjectService) GetProjectRequirementStats(projectID uint) (stats nesmaRes.RequirementStats, err error) {
	// 总需求数
	err = global.GVA_DB.Model(&nesma.NesmaRequirement{}).Where("project_id = ?", projectID).Count(&stats.TotalCount).Error
	if err != nil {
		return stats, err
	}

	// 按层级统计
	var levelStats []struct {
		Level int   `json:"level"`
		Count int64 `json:"count"`
	}
	err = global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Select("level, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("level").
		Scan(&levelStats).Error
	if err != nil {
		return stats, err
	}

	// 分配到对应层级
	for _, stat := range levelStats {
		switch stat.Level {
		case 1:
			stats.Level1 = stat.Count
		case 2:
			stats.Level2 = stat.Count
		case 3:
			stats.Level3 = stat.Count
		case 4:
			stats.Level4 = stat.Count
		}
	}

	// 已完成数
	err = global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("project_id = ? AND status = ?", projectID, "completed").
		Count(&stats.Completed).Error
	if err != nil {
		return stats, err
	}

	return stats, nil
}

// GetNesmaProjectPublic 获取公共项目信息（无需权限验证）
func (nesmaProjectService *NesmaProjectService) GetNesmaProjectPublic() {
	// 这里可以实现获取公共项目的逻辑
	// 比如获取公开的项目模板等
}

// GetNesmaProjectStats 获取项目统计信息
func (nesmaProjectService *NesmaProjectService) GetNesmaProjectStats(userID uint, projectID *uint) (stats map[string]interface{}, err error) {
	stats = make(map[string]interface{})

	// 如果指定了项目ID，则获取特定项目的统计信息
	if projectID != nil {
		// 验证项目权限
		var project nesma.NesmaProject
		err = global.GVA_DB.Where("id = ? AND owner_id = ?", *projectID, userID).First(&project).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("项目不存在或无权限访问")
			}
			return nil, err
		}

		// 返回单个项目的信息
		stats["totalProjects"] = int64(1)
		stats["recentProjects"] = int64(1)
		stats["project"] = project
		
		// 获取项目的需求统计
		requirementStats, err := nesmaProjectService.GetProjectRequirementStats(*projectID)
		if err == nil {
			stats["requirementStats"] = requirementStats
		}

		return stats, nil
	}

	// 获取所有项目的统计信息
	// 总项目数
	var totalProjects int64
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).Where("owner_id = ? and deleted_at is null", userID).Count(&totalProjects).Error
	if err != nil {
		return nil, err
	}
	stats["totalProjects"] = totalProjects

	// 按状态统计
	var statusStats []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).
		Select("status, COUNT(*) as count").
		Where("owner_id = ?", userID).
		Group("status").
		Scan(&statusStats).Error
	if err != nil {
		return nil, err
	}
	stats["statusStats"] = statusStats

	// 按领域统计
	var domainStats []struct {
		Domain string `json:"domain"`
		Count  int64  `json:"count"`
	}
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).
		Select("domain, COUNT(*) as count").
		Where("owner_id = ? AND domain != ''", userID).
		Group("domain").
		Scan(&domainStats).Error
	if err != nil {
		return nil, err
	}
	stats["domainStats"] = domainStats

	// 最近30天创建的项目数
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	var recentProjects int64
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).
		Where("owner_id = ? AND created_at >= ?", userID, thirtyDaysAgo).
		Count(&recentProjects).Error
	if err != nil {
		return nil, err
	}
	stats["recentProjects"] = recentProjects

	// 获取最近的项目列表（用于工作台显示）
	var recentProjectsList []nesma.NesmaProject
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).
		Where("owner_id = ?", userID).
		Order("updated_at DESC").
		Limit(5).
		Find(&recentProjectsList).Error
	if err != nil {
		global.GVA_LOG.Error("获取最近项目列表失败", zap.Error(err))
	} else {
		stats["recentProjectsList"] = recentProjectsList
	}

	// 获取所有项目列表（用于批量操作的下拉选择）
	var allProjectsList []nesma.NesmaProject
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).
		Select("id, name, status").
		Where("owner_id = ? AND status != 'archived'", userID).
		Order("name ASC").
		Find(&allProjectsList).Error
	if err != nil {
		global.GVA_LOG.Error("获取项目列表失败", zap.Error(err))
	} else {
		stats["allProjectsList"] = allProjectsList
	}

	return stats, nil
}

// ArchiveNesmaProject 归档项目
func (nesmaProjectService *NesmaProjectService) ArchiveNesmaProject(ID uint, userID uint) (err error) {
	// 检查项目是否存在且用户有权限
	var project nesma.NesmaProject
	err = global.GVA_DB.Where("id = ? AND owner_id = ?", ID, userID).First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在或无权限操作")
		}
		return err
	}

	// 更新项目状态为归档
	err = global.GVA_DB.Model(&project).Update("status", "archived").Error
	return err
}

// RestoreNesmaProject 恢复项目
func (nesmaProjectService *NesmaProjectService) RestoreNesmaProject(ID uint, userID uint) (err error) {
	// 检查项目是否存在且用户有权限
	var project nesma.NesmaProject
	err = global.GVA_DB.Where("id = ? AND owner_id = ?", ID, userID).First(&project).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("项目不存在或无权限操作")
		}
		return err
	}

	// 更新项目状态为活跃
	err = global.GVA_DB.Model(&project).Update("status", "active").Error
	return err
}

// GetProjectOptions 获取项目选项列表
func (nesmaProjectService *NesmaProjectService) GetProjectOptions(userID uint) ([]map[string]interface{}, error) {
	var projects []map[string]interface{}
	
	// 查询用户有权限的项目，返回选项格式
	err := global.GVA_DB.Table("nesma_projects").
		Select("id as value, name as label, description").
		Where("owner_id = ? AND status != ?", userID, "archived").
		Order("created_at DESC").
		Scan(&projects).Error
	
	if err != nil {
		global.GVA_LOG.Error("查询项目选项失败", zap.Error(err))
		// 返回默认选项
		return []map[string]interface{}{
			{
				"value":       1,
				"label":       "默认项目",
				"description": "系统默认项目",
			},
		}, nil
	}
	
	return projects, nil
}
