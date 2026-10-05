package nesma

import (
	"errors"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"gorm.io/gorm"
)

type NesmaProjectCycleService struct{}

// CreateNesmaProjectCycle 创建项目周期
func (s *NesmaProjectCycleService) CreateNesmaProjectCycle(req *nesmaReq.NesmaProjectCycleRequest) (*nesma.NesmaProjectCycle, error) {
	// 检查项目是否存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return nil, errors.New("项目不存在")
	}

	cycle := &nesma.NesmaProjectCycle{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Status:      "planning", // 默认状态
		Phase:       "requirement", // 默认阶段
	}

	if err := global.GVA_DB.Create(cycle).Error; err != nil {
		return nil, err
	}

	// // 创建默认初始版本
	// initialVersion := &nesma.NesmaRequirementVersion{
	// 	CycleID:     cycle.ID,
	// 	Version:     "v1.0",
	// 	VersionType: "initial",
	// 	CreatedBy:   "user",
	// 	Summary:     "初始版本",
	// 	Description: "用户创建的初始需求版本",
	// 	Status:      "active",
	// }

	// if err := global.GVA_DB.Create(initialVersion).Error; err != nil {
	// 	return nil, err
	// }

	// 预加载关联数据
	global.GVA_DB.Preload("Project").First(cycle, cycle.ID)
	
	return cycle, nil
}

// GetNesmaProjectCycle 根据ID获取项目周期
func (s *NesmaProjectCycleService) GetNesmaProjectCycle(id uint) (*nesma.NesmaProjectCycle, error) {
	var cycle nesma.NesmaProjectCycle
	err := global.GVA_DB.Preload("Project").First(&cycle, id).Error
	return &cycle, err
}

// GetNesmaProjectCycleList 获取项目周期列表
func (s *NesmaProjectCycleService) GetNesmaProjectCycleList(req *nesmaReq.NesmaProjectCycleSearch) ([]nesma.NesmaProjectCycle, int64, error) {
	var cycles []nesma.NesmaProjectCycle
	var total int64

	db := global.GVA_DB.Model(&nesma.NesmaProjectCycle{})

	// 项目过滤
	if req.ProjectID != nil {
		db = db.Where("project_id = ?", *req.ProjectID)
	}

	// 状态过滤
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}

	// 关键词搜索
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}

	// 计算总数
	db.Count(&total)

	// 分页查询
	// 修复：当分页参数为空时设置默认值
	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 50 // 设置默认页大小
	}
	
	offset := (page - 1) * pageSize
	err := db.Preload("Project").
		Offset(offset).
		Limit(pageSize).
		Order("created_at DESC").
		Find(&cycles).Error

	return cycles, total, err
}

// UpdateNesmaProjectCycle 更新项目周期
func (s *NesmaProjectCycleService) UpdateNesmaProjectCycle(req *nesmaReq.NesmaProjectCycleRequest) (*nesma.NesmaProjectCycle, error) {
	var cycle nesma.NesmaProjectCycle
	if err := global.GVA_DB.First(&cycle, req.ID).Error; err != nil {
		return nil, errors.New("项目周期不存在")
	}

	// 更新字段
	cycle.Name = req.Name
	cycle.Description = req.Description
	cycle.StartDate = req.StartDate
	cycle.EndDate = req.EndDate
	if req.Status != "" {
		cycle.Status = req.Status
	}
	if req.Phase != "" {
		cycle.Phase = req.Phase
	}

	if err := global.GVA_DB.Save(&cycle).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据
	global.GVA_DB.Preload("Project").First(&cycle, cycle.ID)
	
	return &cycle, nil
}

// DeleteNesmaProjectCycle 删除项目周期
func (s *NesmaProjectCycleService) DeleteNesmaProjectCycle(id uint) error {
	// 检查周期是否存在
	var cycle nesma.NesmaProjectCycle
	if err := global.GVA_DB.First(&cycle, id).Error; err != nil {
		return errors.New("项目周期不存在")
	}

	// 检查并清空项目的活跃周期
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, cycle.ProjectID).Error; err == nil {
		if project.ActiveCycleID != nil && *project.ActiveCycleID == id {
			// 置空活跃周期
			global.GVA_DB.Model(&project).Update("active_cycle_id", 0)
		}
	}

	// 删除所有版本下需求
	err := global.GVA_DB.Delete(&nesma.NesmaRequirement{}, "version_id IN (SELECT id FROM nesma_requirement_versions WHERE cycle_id = ?)", id).Error
	if err != nil {
		return err
	}
	// 删除副本时先删除副本下的所有版本
	err = global.GVA_DB.Delete(&nesma.NesmaRequirementVersion{}, "cycle_id = ?", id).Error
	if err != nil {
		return err
	}

	return global.GVA_DB.Delete(&cycle).Error
}

// GetProjectCycles 获取指定项目的所有周期
func (s *NesmaProjectCycleService) GetProjectCycles(projectID uint) ([]nesma.NesmaProjectCycle, error) {
	var cycles []nesma.NesmaProjectCycle
	err := global.GVA_DB.Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&cycles).Error
	return cycles, err
}

// UpdateCycleStatus 更新周期状态
func (s *NesmaProjectCycleService) UpdateCycleStatus(id uint, status string) error {
	// 验证状态值
	validStatuses := []string{"planning", "active", "completed", "suspended"}
	isValid := false
	for _, s := range validStatuses {
		if s == status {
			isValid = true
			break
		}
	}
	if !isValid {
		return errors.New("无效的状态值")
	}

	return global.GVA_DB.Model(&nesma.NesmaProjectCycle{}).
		Where("id = ?", id).
		Update("status", status).Error
}

// GetActiveCycle 获取项目的当前活跃周期
func (s *NesmaProjectCycleService) GetActiveCycle(projectID uint) (*nesma.NesmaProjectCycle, error) {
	var cycle nesma.NesmaProjectCycle
	err := global.GVA_DB.Where("project_id = ? AND status = ?", projectID, "active").
		First(&cycle).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果没有活跃周期，返回最新的周期
			err = global.GVA_DB.Where("project_id = ?", projectID).
				Order("created_at DESC").
				First(&cycle).Error
		}
	}
	return &cycle, err
}

// SetActiveProjectCycle 设置项目的激活周期
func (s *NesmaProjectCycleService) SetActiveProjectCycle(cycleID uint) error {
	// 验证周期是否属于该项目
	var cycle nesma.NesmaProjectCycle
	err := global.GVA_DB.Where("id = ?", cycleID).First(&cycle).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("周期不存在或不属于该项目")
		}
		return err
	}

	// 更新项目的激活周期ID
	err = global.GVA_DB.Model(&nesma.NesmaProject{}).
		Where("id = ?", cycle.ProjectID).
		Update("active_cycle_id", cycleID).Error
	return err
}