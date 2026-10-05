package nesma

import (
	"errors"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NesmaRequirementVersionService struct{}

// CreateRequirementVersion 创建需求版本
func (s *NesmaRequirementVersionService) CreateRequirementVersion(req *nesmaReq.NesmaRequirementVersionRequest) (*nesma.NesmaRequirementVersion, error) {
	// 验证周期是否存在
	var cycle nesma.NesmaProjectCycle
	if err := global.GVA_DB.First(&cycle, req.CycleID).Error; err != nil {
		return nil, errors.New("项目周期不存在")
	}

	// 检查版本号是否已存在
	var existingVersion nesma.NesmaRequirementVersion
	err := global.GVA_DB.Where("cycle_id = ? AND version = ?", req.CycleID, req.Version).First(&existingVersion).Error
	if err == nil {
		return nil, errors.New("该周期下已存在相同版本号")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// 如果创建的是激活版本，先将其他版本设为非激活
	if req.Status == "active" {
		if err := global.GVA_DB.Model(&nesma.NesmaRequirementVersion{}).
			Where("cycle_id = ? AND status = ?", req.CycleID, "active").
			Update("status", "archived").Error; err != nil {
			return nil, fmt.Errorf("更新其他版本状态失败: %v", err)
		}
	}

	// 创建新版本
	version := &nesma.NesmaRequirementVersion{
		CycleID:     req.CycleID,
		Version:     req.Version,
		VersionType: req.VersionType,
		CreatedBy:   req.CreatedBy,
		Summary:     req.Summary,
		Description: req.Description,
		Status:      req.Status,
	}

	if err := global.GVA_DB.Create(version).Error; err != nil {
		return nil, fmt.Errorf("创建版本失败: %v", err)
	}

	// 预加载关联数据
	if err := global.GVA_DB.Preload("Cycle").First(version, version.ID).Error; err != nil {
		return nil, err
	}

	global.GVA_LOG.Info("创建需求版本成功", 
		zap.Uint("versionId", version.ID),
		zap.String("version", version.Version),
		zap.Uint("cycleId", version.CycleID))

	return version, nil
}

// GetRequirementVersions 获取周期的所有版本
func (s *NesmaRequirementVersionService) GetRequirementVersions(cycleID uint) ([]nesma.NesmaRequirementVersion, error) {
	var versions []nesma.NesmaRequirementVersion
	
	err := global.GVA_DB.Where("cycle_id = ?", cycleID).
		Order("created_at ASC").
		Find(&versions).Error
	
	if err != nil {
		return nil, fmt.Errorf("获取需求版本失败: %v", err)
	}

	return versions, nil
}

// GetRequirementVersion 根据ID获取版本详情
func (s *NesmaRequirementVersionService) GetRequirementVersion(id uint) (*nesma.NesmaRequirementVersion, error) {
	var version nesma.NesmaRequirementVersion
	
	err := global.GVA_DB.Preload("Cycle").
		Preload("Requirements").
		First(&version, id).Error
	
	if err != nil {
		return nil, fmt.Errorf("获取版本详情失败: %v", err)
	}

	return &version, nil
}

// UpdateRequirementVersion 更新需求版本
func (s *NesmaRequirementVersionService) UpdateRequirementVersion(req *nesmaReq.NesmaRequirementVersionRequest) (*nesma.NesmaRequirementVersion, error) {
	var version nesma.NesmaRequirementVersion
	if err := global.GVA_DB.First(&version, req.ID).Error; err != nil {
		return nil, errors.New("版本不存在")
	}

	// 如果更新为激活状态，先将其他版本设为非激活
	if req.Status == "active" && version.Status != "active" {
		if err := global.GVA_DB.Model(&nesma.NesmaRequirementVersion{}).
			Where("cycle_id = ? AND status = ? AND id != ?", version.CycleID, "active", req.ID).
			Update("status", "archived").Error; err != nil {
			return nil, fmt.Errorf("更新其他版本状态失败: %v", err)
		}
	}

	// 更新版本信息
	updates := map[string]interface{}{
		"version":     req.Version,
		"version_type": req.VersionType,
		"created_by":  req.CreatedBy,
		"summary":     req.Summary,
		"description": req.Description,
		"status":      req.Status,
	}

	if err := global.GVA_DB.Model(&version).Updates(updates).Error; err != nil {
		return nil, fmt.Errorf("更新版本失败: %v", err)
	}

	// 重新获取更新后的版本信息
	if err := global.GVA_DB.Preload("Cycle").First(&version, req.ID).Error; err != nil {
		return nil, err
	}

	global.GVA_LOG.Info("更新需求版本成功", 
		zap.Uint("versionId", version.ID),
		zap.String("version", version.Version))

	return &version, nil
}

// DeleteRequirementVersion 删除需求版本
func (s *NesmaRequirementVersionService) DeleteRequirementVersion(id uint) error {
	var version nesma.NesmaRequirementVersion
	if err := global.GVA_DB.First(&version, id).Error; err != nil {
		return errors.New("版本不存在")
	}

	// 检查是否有关联的需求
	var requirementCount int64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("version_id = ?", id).
		Count(&requirementCount)
	
	if requirementCount > 0 {
		return errors.New("该版本下还有需求，无法删除")
	}

	// 删除版本
	if err := global.GVA_DB.Delete(&version).Error; err != nil {
		return fmt.Errorf("删除版本失败: %v", err)
	}

	global.GVA_LOG.Info("删除需求版本成功", 
		zap.Uint("versionId", id),
		zap.String("version", version.Version))

	return nil
}

// GetActiveVersion 获取周期的激活版本
func (s *NesmaRequirementVersionService) GetActiveVersion(cycleID uint) (*nesma.NesmaRequirementVersion, error) {
	var version nesma.NesmaRequirementVersion
	
	err := global.GVA_DB.Where("cycle_id = ? AND status = ?", cycleID, "active").
		First(&version).Error
	
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("该周期没有激活版本")
		}
		return nil, fmt.Errorf("获取激活版本失败: %v", err)
	}

	return &version, nil
}

// SetActiveVersion 设置激活版本
func (s *NesmaRequirementVersionService) SetActiveVersion(versionID uint) error {
	var version nesma.NesmaRequirementVersion
	if err := global.GVA_DB.First(&version, versionID).Error; err != nil {
		return errors.New("版本不存在")
	}

	// 开始事务
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 将同周期的其他版本设为非激活
	if err := tx.Model(&nesma.NesmaRequirementVersion{}).
		Where("cycle_id = ? AND status = ?", version.CycleID, "active").
		Update("status", "archived").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("更新其他版本状态失败: %v", err)
	}

	// 设置当前版本为激活
	if err := tx.Model(&version).Update("status", "active").Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("设置激活版本失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交事务失败: %v", err)
	}

	global.GVA_LOG.Info("设置激活版本成功", 
		zap.Uint("versionId", versionID),
		zap.String("version", version.Version))

	return nil
}

// CreateOrGetVersion 创建或获取版本（用于导入）
func (s *NesmaRequirementVersionService) CreateOrGetVersion(tx *gorm.DB, cycleID uint, versionName string, createdBy string) (*nesma.NesmaRequirementVersion, error) {
	// 检查版本是否已存在
	var existingVersion nesma.NesmaRequirementVersion
	err := tx.Where("cycle_id = ? AND version = ?", cycleID, versionName).First(&existingVersion).Error
	
	if err == nil {
		// 版本已存在，返回现有版本
		global.GVA_LOG.Info("版本已存在，使用现有版本", 
			zap.Uint("versionId", existingVersion.ID),
			zap.String("version", existingVersion.Version),
			zap.Uint("cycleId", cycleID))
		return &existingVersion, nil
	}
	
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("查询版本失败: %v", err)
	}

	// 版本不存在，创建新版本
	version := &nesma.NesmaRequirementVersion{
		CycleID:     cycleID,
		Version:     versionName,
		VersionType: "import", // 导入版本
		CreatedBy:   createdBy,
		Summary:     fmt.Sprintf("Excel导入版本 %s", versionName),
		Description: fmt.Sprintf("通过Excel导入创建的版本 %s", versionName),
		Status:      "active", // 导入的版本设为激活状态
	}

	if err := tx.Create(version).Error; err != nil {
		return nil, fmt.Errorf("创建版本失败: %v", err)
	}

	global.GVA_LOG.Info("创建导入版本成功", 
		zap.Uint("versionId", version.ID),
		zap.String("version", version.Version),
		zap.Uint("cycleId", cycleID))

	return version, nil
}

// GenerateNextVersion 生成下一个版本号
func (s *NesmaRequirementVersionService) GenerateNextVersion(cycleID uint, versionType string) (string, error) {
	// 获取当前周期的最新版本
	var latestVersion nesma.NesmaRequirementVersion
	err := global.GVA_DB.Where("cycle_id = ?", cycleID).
		Order("created_at DESC").
		First(&latestVersion).Error
	
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// 没有版本，返回初始版本
		return "v1.0", nil
	}
	
	if err != nil {
		return "", fmt.Errorf("获取最新版本失败: %v", err)
	}

	// 解析版本号并生成下一个版本
	nextVersion := s.parseAndIncrementVersion(latestVersion.Version, versionType)
	return nextVersion, nil
}

// parseAndIncrementVersion 解析版本号并递增
func (s *NesmaRequirementVersionService) parseAndIncrementVersion(currentVersion, versionType string) string {
	// 默认版本号处理
	if currentVersion == "" {
		return "v1.0"
	}

	// 解析版本号 (如 v1.1 -> [1, 1])
	var major, minor int
	if _, err := fmt.Sscanf(currentVersion, "v%d.%d", &major, &minor); err != nil {
		// 如果解析失败，返回默认版本
		return "v1.0"
	}

	// 根据版本类型决定递增策略
	switch versionType {
	case "initial":
		// 初始版本，小版本号+1
		return fmt.Sprintf("v%d.%d", major, minor+1)
	case "analyzed":
		// AI分析版本，小版本号+1
		return fmt.Sprintf("v%d.%d", major, minor+1)
	case "optimized":
		// 人工优化版本，小版本号+1
		return fmt.Sprintf("v%d.%d", major, minor+1)
	case "finalized":
		// 最终版本，主版本号+1，小版本号重置为0
		return fmt.Sprintf("v%d.0", major+1)
	default:
		// 默认策略，小版本号+1
		return fmt.Sprintf("v%d.%d", major, minor+1)
	}
}

// UpdateVersionStats 更新版本统计信息
func (s *NesmaRequirementVersionService) UpdateVersionStats(versionID uint) error {
	var version nesma.NesmaRequirementVersion
	if err := global.GVA_DB.First(&version, versionID).Error; err != nil {
		return errors.New("版本不存在")
	}

	// 统计该版本的需求数量
	var requirementCount int64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("version_id = ?", versionID).
		Count(&requirementCount)

	// 统计已分析的需求数量
	var analyzedCount int64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("version_id = ? AND ai_analysis_status = ?", versionID, "completed").
		Count(&analyzedCount)

	// 统计已优化的需求数量
	var optimizedCount int64
	global.GVA_DB.Model(&nesma.NesmaRequirement{}).
		Where("version_id = ? AND ai_analysis_status = ? AND ai_description IS NOT NULL AND ai_description != ''", versionID, "completed").
		Count(&optimizedCount)

	// 计算分析进度
	var analysisProgress float64
	if requirementCount > 0 {
		analysisProgress = float64(analyzedCount) / float64(requirementCount) * 100
	}

	// 更新版本统计信息
	updates := map[string]interface{}{
		"requirement_count":  int(requirementCount),
		"analyzed_count":     int(analyzedCount),
		"optimized_count":    int(optimizedCount),
		"analysis_progress":  analysisProgress,
	}

	if err := global.GVA_DB.Model(&version).Updates(updates).Error; err != nil {
		return fmt.Errorf("更新版本统计失败: %v", err)
	}

	global.GVA_LOG.Info("更新版本统计成功", 
		zap.Uint("versionId", versionID),
		zap.Int64("requirementCount", requirementCount),
		zap.Int64("analyzedCount", analyzedCount),
		zap.Float64("analysisProgress", analysisProgress))

	return nil
} 