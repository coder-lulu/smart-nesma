package initialize

import (
	"os"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/example"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

func Gorm() *gorm.DB {
	switch global.GVA_CONFIG.System.DbType {
	case "mysql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mysql.Dbname
		return GormMysql()
	case "pgsql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Pgsql.Dbname
		return GormPgSql()
	case "oracle":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Oracle.Dbname
		return GormOracle()
	case "mssql":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mssql.Dbname
		return GormMssql()
	case "sqlite":
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Sqlite.Dbname
		return GormSqlite()
	default:
		global.GVA_ACTIVE_DBNAME = &global.GVA_CONFIG.Mysql.Dbname
		return GormMysql()
	}
}

func RegisterTables() {
	db := global.GVA_DB
	err := db.AutoMigrate(

		system.SysApi{},
		system.SysIgnoreApi{},
		system.SysUser{},
		system.SysBaseMenu{},
		system.JwtBlacklist{},
		system.SysAuthority{},
		system.SysDictionary{},
		system.SysOperationRecord{},
		system.SysAutoCodeHistory{},
		system.SysDictionaryDetail{},
		system.SysBaseMenuParameter{},
		system.SysBaseMenuBtn{},
		system.SysAuthorityBtn{},
		system.SysAutoCodePackage{},
		system.SysExportTemplate{},
		system.Condition{},
		system.JoinTemplate{},
		system.SysParams{},

		example.ExaFile{},
		example.ExaCustomer{},
		example.ExaFileChunk{},
		example.ExaFileUploadAndDownload{},
		example.ExaAttachmentCategory{},

		// NESMA相关模型 - 项目管理
		nesma.NesmaProject{},
		nesma.NesmaProjectCycle{},

		// NESMA相关模型 - 需求管理
		nesma.NesmaRequirement{},
		nesma.NesmaRequirementVersion{},
		nesma.NesmaRequirementAnalysis{},
		nesma.NesmaRequirementAnalysisTask{},

		// NESMA相关模型 - 评估引擎
		nesma.NesmaEvaluation{},
		nesma.NesmaFunctionPoint{},
		nesma.NesmaComplexityMetric{},
		nesma.NesmaValidationItem{},
		nesma.NesmaEvaluationHistory{},

		// NESMA相关模型 - 文档生成
		nesma.NesmaDocument{},
		nesma.NesmaDocTemplate{},
		nesma.NesmaDocBatch{},

		// NESMA相关模型 - 知识库
		nesma.NesmaKnowledgeEntry{},
		nesma.NesmaKnowledgeRule{},
		nesma.NesmaCaseStudy{},

		// NESMA相关模型 - 记忆系统
		nesma.NesmaVectorMemory{},
		nesma.NesmaCompactMemory{},
		nesma.NesmaSessionMemory{},
		nesma.NesmaAgentInteraction{},

		// NESMA相关模型 - Agent系统
		nesma.NesmaAgent{},
		nesma.NesmaAgentTask{},
		nesma.NesmaAgentMessage{},
		nesma.NesmaWorkflow{},
		nesma.NesmaWorkflowExecution{},

		// NESMA相关模型 - 向量存储
		nesma.NesmaVectorStore{},
		nesma.NesmaVectorIndex{},
		nesma.NesmaVectorQuery{},
		nesma.NesmaEmbedding{},
		
		// AI项目分析相关模型
		nesma.AIProjectAnalysis{},
		nesma.AIFunctionalAnalysis{},
		nesma.AITechnicalAnalysis{},
		nesma.AIBusinessAnalysis{},
		nesma.AIQualityAnalysis{},
		nesma.AIRiskAnalysis{},
		nesma.AIRecommendationAnalysis{},
		nesma.AIComplianceAnalysis{},
	)
	if err != nil {
		global.GVA_LOG.Error("register table failed", zap.Error(err))
		os.Exit(0)
	}

	// 注册AI服务相关模型
	err = db.AutoMigrate(
		// AI服务监控模型
		nesma.AIServiceMetrics{},

		// 推荐系统模型
		nesma.UserBehavior{},
		nesma.RecommendationHistory{},

		// 聊天系统模型
		nesma.ChatSession{},
		nesma.ChatMessage{},

		// 需求分析系统模型
		nesma.RequirementAnalysisResult{},
	)
	if err != nil {
		global.GVA_LOG.Error("register ai service tables failed", zap.Error(err))
		os.Exit(0)
	}

	err = bizModel()

	if err != nil {
		global.GVA_LOG.Error("register biz_table failed", zap.Error(err))
		os.Exit(0)
	}
	global.GVA_LOG.Info("register table success")

	// 执行NESMA自定义数据迁移
	MigrateNesmaTables()

	// 修复数据库表结构缺失问题
	if err := FixDatabaseIssues(); err != nil {
		global.GVA_LOG.Error("修复数据库问题失败", zap.Error(err))
		// 不退出程序，继续运行，但记录错误
	}

	// 检查NESMA评估系统健康状态
	if !CheckNesmaEvaluationSystemHealth() {
		global.GVA_LOG.Warn("NESMA评估系统健康检查未完全通过，请检查数据库配置")
	}

	// 初始化NESMA示例模板数据
	InitNesmaSampleTemplates()
}
