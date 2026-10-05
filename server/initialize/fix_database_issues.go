package initialize

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// FixDatabaseIssues 修复数据库表结构缺失问题
func FixDatabaseIssues() error {
	db := global.GVA_DB
	if db == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	global.GVA_LOG.Info("开始修复NESMA数据库表结构问题...")

	// 1. 检查nesma_evaluations表是否存在error_message列
	if !db.Migrator().HasColumn("nesma_evaluations", "error_message") {
		global.GVA_LOG.Info("添加nesma_evaluations表缺失的错误处理列...")
		
		// 添加错误处理相关列
		err := db.Exec(`
			ALTER TABLE nesma_evaluations 
			ADD COLUMN IF NOT EXISTS error_message TEXT,
			ADD COLUMN IF NOT EXISTS error_code VARCHAR(50),
			ADD COLUMN IF NOT EXISTS error_details TEXT
		`).Error
		
		if err != nil {
			global.GVA_LOG.Error("添加nesma_evaluations错误处理列失败", zap.Error(err))
			return err
		}
		
		global.GVA_LOG.Info("nesma_evaluations表错误处理列添加成功")
	}

	// 2. 检查并创建AI分析相关表
	aiTables := []struct {
		name     string
		checkSQL string
	}{
		{"ai_project_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_project_analyses'"},
		{"ai_functional_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_functional_analyses'"},
		{"ai_technical_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_technical_analyses'"},
		{"ai_business_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_business_analyses'"},
		{"ai_quality_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_quality_analyses'"},
		{"ai_risk_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_risk_analyses'"},
		{"ai_recommendation_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_recommendation_analyses'"},
		{"ai_compliance_analyses", "SELECT 1 FROM information_schema.tables WHERE table_name = 'ai_compliance_analyses'"},
		{"nesma_evaluation_factors", "SELECT 1 FROM information_schema.tables WHERE table_name = 'nesma_evaluation_factors'"},
	}

	missingTables := []string{}
	for _, table := range aiTables {
		var count int
		err := db.Raw(table.checkSQL).Scan(&count).Error
		if err != nil || count == 0 {
			missingTables = append(missingTables, table.name)
		}
	}

	if len(missingTables) > 0 {
		global.GVA_LOG.Info("检测到缺失的AI分析表", zap.Strings("tables", missingTables))
		
		// 直接通过GORM创建缺失的表
		err := createMissingTables(missingTables)
		if err != nil {
			global.GVA_LOG.Error("创建缺失表失败", zap.Error(err))
			return err
		}
		
		global.GVA_LOG.Info("AI分析表创建完成")
	}

	// 3. 验证修复结果
	err := validateDatabaseFix()
	if err != nil {
		global.GVA_LOG.Error("数据库修复验证失败", zap.Error(err))
		return err
	}

	global.GVA_LOG.Info("NESMA数据库表结构修复完成")
	return nil
}

// createMissingTables 直接创建缺失的表
func createMissingTables(missingTables []string) error {
	db := global.GVA_DB
	
	for _, tableName := range missingTables {
		switch tableName {
		case "nesma_evaluation_factors":
			// 创建评估因子配置表
			sql := `CREATE TABLE IF NOT EXISTS nesma_evaluation_factors (
				id BIGSERIAL PRIMARY KEY,
				created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
				deleted_at TIMESTAMPTZ NULL,
				evaluation_id BIGINT NOT NULL,
				factor_config TEXT NOT NULL,
				nesma_version VARCHAR(20) DEFAULT 'v2.2'
			)`
			err := db.Exec(sql).Error
			if err != nil {
				return fmt.Errorf("创建nesma_evaluation_factors表失败: %v", err)
			}
			
			// 创建索引
			err = db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_evaluation_factors_evaluation_id ON nesma_evaluation_factors(evaluation_id)").Error
			if err != nil {
				global.GVA_LOG.Warn("创建索引失败", zap.Error(err))
			}
			
			err = db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_evaluation_factors_nesma_version ON nesma_evaluation_factors(nesma_version)").Error
			if err != nil {
				global.GVA_LOG.Warn("创建索引失败", zap.Error(err))
			}
			
			global.GVA_LOG.Info("nesma_evaluation_factors表创建成功")
		default:
			global.GVA_LOG.Warn("未知的缺失表", zap.String("table", tableName))
		}
	}
	
	return nil
}

// executeFixDatabaseSQL 执行数据库修复SQL脚本
func executeFixDatabaseSQL() error {
	db := global.GVA_DB
	
	// 读取SQL文件
	sqlFilePath := filepath.Join("migration", "fix_database_issues.sql")
	sqlContent, err := ioutil.ReadFile(sqlFilePath)
	if err != nil {
		global.GVA_LOG.Error("读取数据库修复SQL文件失败", zap.Error(err))
		return err
	}

	// 分割SQL语句（按分号分割）
	sqlStatements := strings.Split(string(sqlContent), ";")
	
	// 执行每个SQL语句
	for i, stmt := range sqlStatements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		
		global.GVA_LOG.Debug("执行SQL语句", zap.Int("index", i), zap.String("sql", stmt[:min(len(stmt), 100)]))
		
		err := db.Exec(stmt).Error
		if err != nil {
			// 忽略某些预期的错误（如表已存在）
			if strings.Contains(err.Error(), "already exists") || 
			   strings.Contains(err.Error(), "duplicate column") {
				global.GVA_LOG.Debug("忽略预期错误", zap.Error(err))
				continue
			}
			
			global.GVA_LOG.Error("执行SQL语句失败", 
				zap.Int("index", i), 
				zap.String("sql", stmt), 
				zap.Error(err))
			return err
		}
	}
	
	return nil
}

// validateDatabaseFix 验证数据库修复结果
func validateDatabaseFix() error {
	db := global.GVA_DB
	
	// 1. 验证nesma_evaluations表的error_message列
	if !db.Migrator().HasColumn("nesma_evaluations", "error_message") {
		return fmt.Errorf("nesma_evaluations表仍然缺少error_message列")
	}
	
	// 2. 验证AI分析表是否存在
	requiredTables := []string{
		"ai_project_analyses",
		"ai_functional_analyses", 
		"ai_technical_analyses",
		"ai_business_analyses",
		"ai_quality_analyses",
		"ai_risk_analyses",
		"ai_recommendation_analyses",
		"ai_compliance_analyses",
	}
	
	for _, tableName := range requiredTables {
		if !db.Migrator().HasTable(tableName) {
			return fmt.Errorf("表 %s 仍然不存在", tableName)
		}
	}
	
	global.GVA_LOG.Info("数据库修复验证通过", zap.Strings("verified_tables", requiredTables))
	return nil
}

// min 辅助函数
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CheckNesmaEvaluationSystemHealth 检查NESMA评估系统健康状态
func CheckNesmaEvaluationSystemHealth() bool {
	db := global.GVA_DB
	if db == nil {
		global.GVA_LOG.Error("数据库连接未初始化")
		return false
	}

	// 检查关键表是否存在
	criticalTables := map[string]string{
		"nesma_evaluations":        "NESMA评估记录表",
		"ai_project_analyses":      "AI项目分析主表",
		"ai_functional_analyses":   "AI功能性分析表",
		"nesma_projects":           "NESMA项目表",
		"nesma_requirements":       "NESMA需求表",
	}

	healthy := true
	for tableName, description := range criticalTables {
		if !db.Migrator().HasTable(tableName) {
			global.GVA_LOG.Error("关键表不存在", 
				zap.String("table", tableName), 
				zap.String("description", description))
			healthy = false
		}
	}

	// 检查nesma_evaluations表的关键列
	if db.Migrator().HasTable("nesma_evaluations") {
		requiredColumns := []string{"error_message", "error_code", "status", "project_id"}
		for _, column := range requiredColumns {
			if !db.Migrator().HasColumn("nesma_evaluations", column) {
				global.GVA_LOG.Error("nesma_evaluations表缺少关键列", zap.String("column", column))
				healthy = false
			}
		}
	}

	if healthy {
		global.GVA_LOG.Info("NESMA评估系统健康检查通过")
	} else {
		global.GVA_LOG.Error("NESMA评估系统健康检查失败")
	}

	return healthy
}