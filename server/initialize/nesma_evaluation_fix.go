package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// FixExistingEvaluations 修复现有评估记录的周期和版本关联
func FixExistingEvaluations() {
	// 检查是否需要修复
	var countWithNull int64
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("cycle_id IS NULL OR requirement_version_id IS NULL").
		Count(&countWithNull)

	if countWithNull == 0 {
		global.GVA_LOG.Info("无需修复评估记录")
		return
	}

	global.GVA_LOG.Info("开始修复现有评估记录", zap.Int64("需要修复的记录数", countWithNull))

	// 获取所有项目
	var projects []nesma.NesmaProject
	if err := global.GVA_DB.Find(&projects).Error; err != nil {
		global.GVA_LOG.Error("获取项目列表失败", zap.Error(err))
		return
	}

	for _, project := range projects {
		// 为每个项目创建默认周期（如果不存在）
		var cycleCount int64
		global.GVA_DB.Model(&nesma.NesmaProjectCycle{}).
			Where("project_id = ?", project.ID).
			Count(&cycleCount)

		var cycle nesma.NesmaProjectCycle
		if cycleCount == 0 {
			// 创建默认周期
			cycle = nesma.NesmaProjectCycle{
				ProjectID:   project.ID,
				Name:        "默认周期",
				Description: "系统自动创建的默认周期，用于兼容历史评估数据",
				Status:      "active",
				Phase:       "requirement",
			}
			if err := global.GVA_DB.Create(&cycle).Error; err != nil {
				global.GVA_LOG.Error("创建默认周期失败", zap.Error(err), zap.Uint("projectId", project.ID))
				continue
			}
		} else {
			// 获取第一个周期
			global.GVA_DB.Where("project_id = ?", project.ID).
				Order("created_at ASC").
				First(&cycle)
		}

		// 为周期创建默认版本（如果不存在）
		var versionCount int64
		global.GVA_DB.Model(&nesma.NesmaRequirementVersion{}).
			Where("cycle_id = ?", cycle.ID).
			Count(&versionCount)

		var version nesma.NesmaRequirementVersion
		if versionCount == 0 {
			// 创建默认版本
			version = nesma.NesmaRequirementVersion{
				CycleID:     cycle.ID,
				Version:     "v1.0",
				VersionType: "initial",
				CreatedBy:   "system",
				Summary:     "系统自动创建的默认版本，用于兼容历史评估数据",
				Status:      "active",
			}
			if err := global.GVA_DB.Create(&version).Error; err != nil {
				global.GVA_LOG.Error("创建默认版本失败", zap.Error(err), zap.Uint("cycleId", cycle.ID))
				continue
			}
		} else {
			// 获取第一个版本
			global.GVA_DB.Where("cycle_id = ?", cycle.ID).
				Order("created_at ASC").
				First(&version)
		}

		// 更新该项目下的所有NULL评估记录
		result := global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
			Where("project_id = ? AND (cycle_id IS NULL OR requirement_version_id IS NULL)", project.ID).
			Updates(map[string]interface{}{
				"cycle_id":                &cycle.ID,
				"requirement_version_id":  &version.ID,
			})

		if result.Error != nil {
			global.GVA_LOG.Error("更新评估记录失败", zap.Error(result.Error), zap.Uint("projectId", project.ID))
		} else if result.RowsAffected > 0 {
			global.GVA_LOG.Info("成功修复评估记录", 
				zap.Uint("projectId", project.ID),
				zap.String("projectName", project.Name),
				zap.Int64("修复数量", result.RowsAffected))
		}
	}

	// 验证修复结果
	var finalCountWithNull int64
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("cycle_id IS NULL OR requirement_version_id IS NULL").
		Count(&finalCountWithNull)

	global.GVA_LOG.Info("评估记录修复完成", 
		zap.Int64("修复前NULL记录数", countWithNull),
		zap.Int64("修复后NULL记录数", finalCountWithNull))
}