package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// MigrateNesmaTables 迁移NESMA相关表结构
func MigrateNesmaTables() {
	db := global.GVA_DB
	
	// 按顺序迁移表结构，确保外键依赖正确
	err := db.AutoMigrate(
		&nesma.NesmaProject{},           // 项目表（基础表）
		&nesma.NesmaProjectCycle{},      // 项目周期表
		&nesma.NesmaRequirementVersion{}, // 需求版本表
		&nesma.NesmaRequirement{},       // 需求表（更新字段）
		&nesma.NesmaRequirementAnalysisTask{}, // 分析任务表
		&nesma.NesmaEvaluationFactors{}, // 评估因子配置表
		
		// 新增的分析相关表结构
		&nesma.NesmaProjectAnalysis{},        // 项目分析记录表
		&nesma.NesmaRequirementOptimization{}, // 需求优化建议表
		&nesma.NesmaNESMAEvaluation{},        // NESMA评估记录表
		&nesma.NesmaOptimizationSuggestion{}, // 优化建议表
		&nesma.NesmaAnalysisHistory{},        // 分析历史记录表
		
		// AI项目分析相关表结构
		&nesma.AIProjectAnalysis{},        // AI项目分析主表
		&nesma.AIFunctionalAnalysis{},     // AI功能性分析表
		&nesma.AITechnicalAnalysis{},      // AI技术性分析表
		&nesma.AIBusinessAnalysis{},       // AI业务性分析表
		&nesma.AIQualityAnalysis{},        // AI质量分析表
		&nesma.AIRiskAnalysis{},           // AI风险分析表
		&nesma.AIRecommendationAnalysis{}, // AI建议分析表
		&nesma.AIComplianceAnalysis{},     // AI合规性分析表
	)
	
	if err != nil {
		global.GVA_LOG.Error("NESMA表结构迁移失败", zap.Error(err))
		return
	}
	
	global.GVA_LOG.Info("NESMA表结构迁移成功")
	
	// 添加缺失的字段
	AddAIAnalysisIDField()
	
	// 创建必要的索引
	createNesmaIndexes()
	
	// 插入默认数据
	insertNesmaDefaultData()
}

// createNesmaIndexes 创建NESMA相关索引
func createNesmaIndexes() {
	db := global.GVA_DB
	
	// 为需求表创建复合索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirements_project_cycle ON nesma_requirements(project_id, cycle_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirements_version_level ON nesma_requirements(version_id, level)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirements_ai_status ON nesma_requirements(ai_analysis_status)")
	
	// 为项目周期表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_project_cycles_status ON nesma_project_cycles(status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_project_cycles_phase ON nesma_project_cycles(phase)")
	
	// 为需求版本表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirement_versions_type ON nesma_requirement_versions(version_type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirement_versions_status ON nesma_requirement_versions(status)")
	
	// 为分析任务表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_analysis_tasks_type_status ON nesma_requirement_analysis_tasks(task_type, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_analysis_tasks_priority ON nesma_requirement_analysis_tasks(priority)")
	
	// 为项目分析记录表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_project_analyses_project_cycle ON nesma_project_analyses(project_id, cycle_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_project_analyses_analysis_type ON nesma_project_analyses(analysis_type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_project_analyses_overall_grade ON nesma_project_analyses(overall_grade)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_project_analyses_analysis_date ON nesma_project_analyses(analysis_date)")
	
	// 为需求优化建议表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirement_optimizations_requirement ON nesma_requirement_optimizations(requirement_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirement_optimizations_decision ON nesma_requirement_optimizations(user_decision)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirement_optimizations_generated_at ON nesma_requirement_optimizations(generated_at)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirement_optimizations_applied ON nesma_requirement_optimizations(applied_to_requirement)")
	
	// 为NESMA评估记录表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_nesma_evaluations_project_cycle ON nesma_nesma_evaluations(project_id, cycle_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_nesma_evaluations_evaluation_type ON nesma_nesma_evaluations(evaluation_type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_nesma_evaluations_overall_grade ON nesma_nesma_evaluations(overall_grade)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_nesma_evaluations_evaluation_date ON nesma_nesma_evaluations(evaluation_date)")
	
	// 为优化建议表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_optimization_suggestions_project ON nesma_optimization_suggestions(project_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_optimization_suggestions_type_category ON nesma_optimization_suggestions(suggestion_type, category)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_optimization_suggestions_priority_status ON nesma_optimization_suggestions(priority, status)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_optimization_suggestions_assigned_to ON nesma_optimization_suggestions(assigned_to)")
	
	// 为分析历史记录表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_analysis_histories_project_type ON nesma_analysis_histories(project_id, analysis_type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_analysis_histories_analysis_date ON nesma_analysis_histories(analysis_date)")
	
	// 为AI项目分析主表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_project_evaluation ON ai_project_analyses(project_id, evaluation_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_status_progress ON ai_project_analyses(status, progress)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_analysis_type ON ai_project_analyses(analysis_type)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_overall_score ON ai_project_analyses(overall_score)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_project_analyses_start_time ON ai_project_analyses(start_time)")
	
	// 为AI各维度分析表创建索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_functional_analyses_analysis_id ON ai_functional_analyses(analysis_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_technical_analyses_analysis_id ON ai_technical_analyses(analysis_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_business_analyses_analysis_id ON ai_business_analyses(analysis_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_quality_analyses_analysis_id ON ai_quality_analyses(analysis_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_risk_analyses_analysis_id ON ai_risk_analyses(analysis_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_recommendation_analyses_analysis_id ON ai_recommendation_analyses(analysis_id)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_compliance_analyses_analysis_id ON ai_compliance_analyses(analysis_id)")
	
	// 为AI建议分析表创建更多索引
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_recommendation_analyses_type_priority ON ai_recommendation_analyses(recommendation_type, priority)")
	db.Exec("CREATE INDEX IF NOT EXISTS idx_ai_recommendation_analyses_impact_urgency ON ai_recommendation_analyses(impact_level, urgency)")
	
	global.GVA_LOG.Info("NESMA索引创建完成")
}

// insertNesmaDefaultData 插入NESMA默认数据
func insertNesmaDefaultData() {
	db := global.GVA_DB
	
	// 检查是否需要为现有项目创建默认周期
	var existingProjects []nesma.NesmaProject
	db.Find(&existingProjects)
	
	for _, project := range existingProjects {
		// 检查项目是否已有周期
		var cycleCount int64
		db.Model(&nesma.NesmaProjectCycle{}).Where("project_id = ?", project.ID).Count(&cycleCount)
		
		if cycleCount == 0 {
			// 为项目创建默认周期
			defaultCycle := nesma.NesmaProjectCycle{
				ProjectID:   project.ID,
				Name:        "一期",
				Description: "系统建设第一期",
				Status:      "planning",
				Phase:       "requirement",
			}
			
			if err := db.Create(&defaultCycle).Error; err != nil {
				global.GVA_LOG.Error("创建默认项目周期失败", zap.Error(err), zap.Uint("projectId", project.ID))
				continue
			}
			
			// 为周期创建初始版本
			initialVersion := nesma.NesmaRequirementVersion{
				CycleID:     defaultCycle.ID,
				Version:     "v1.0",
				VersionType: "initial",
				CreatedBy:   "system",
				Summary:     "系统初始版本",
				Description: "系统自动创建的初始需求版本",
				Status:      "active",
			}
			
			if err := db.Create(&initialVersion).Error; err != nil {
				global.GVA_LOG.Error("创建初始版本失败", zap.Error(err), zap.Uint("cycleId", defaultCycle.ID))
				continue
			}
			
			// 将该项目的现有需求关联到新创建的周期和版本
			db.Model(&nesma.NesmaRequirement{}).
				Where("project_id = ? AND cycle_id IS NULL", project.ID).
				Updates(map[string]interface{}{
					"cycle_id":   defaultCycle.ID,
					"version_id": initialVersion.ID,
				})
			
			global.GVA_LOG.Info("为项目创建默认周期和版本", 
				zap.Uint("projectId", project.ID), 
				zap.String("projectName", project.Name))
		}
	}
	
	global.GVA_LOG.Info("NESMA默认数据插入完成")
}

// CheckNesmaMigrationStatus 检查NESMA迁移状态
func CheckNesmaMigrationStatus() bool {
	db := global.GVA_DB
	
	// 检查基础表是否存在
	baseTables := []string{
		"nesma_projects",
		"nesma_project_cycles",
		"nesma_requirement_versions",
		"nesma_requirements",
		"nesma_requirement_analysis_tasks",
		"nesma_evaluation_factors",
	}
	
	// 检查新增的分析表是否存在
	analysisTables := []string{
		"nesma_project_analyses",        // 项目分析记录表
		"nesma_requirement_optimizations", // 需求优化建议表
		"nesma_nesma_evaluations",        // NESMA评估记录表
		"nesma_optimization_suggestions", // 优化建议表
		"nesma_analysis_histories",       // 分析历史记录表
	}
	
	allTables := append(baseTables, analysisTables...)
	
	for _, tableName := range allTables {
		if !db.Migrator().HasTable(tableName) {
			global.GVA_LOG.Error("NESMA表不存在", zap.String("table", tableName))
			return false
		}
	}
	
	// 检查需求表是否有新字段
	requiredColumns := []string{
		"cycle_id",
		"version_id", 
		"ai_analysis_status",
		"ai_description",
		"ai_generated_title",
		"ai_complexity_score",
		"ai_confidence_score",
		"ai_analysis_time",
		"ai_analysis_log",
		"similar_requirements",
		"recommended_afp",
		"recommended_ufp",
	}
	
	for _, column := range requiredColumns {
		if !db.Migrator().HasColumn(&nesma.NesmaRequirement{}, column) {
			global.GVA_LOG.Error("需求表缺少必要字段", zap.String("column", column))
			return false
		}
	}
	
	global.GVA_LOG.Info("NESMA迁移状态检查通过")
	return true
}