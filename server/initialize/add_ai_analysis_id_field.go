package initialize

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// AddAIAnalysisIDField 添加AI分析ID字段到评估表
func AddAIAnalysisIDField() {
	db := global.GVA_DB
	
	// 要添加的字段列表
	fieldsToAdd := []struct {
		Name string
		SQL  string
	}{
		{"ai_analysis_id", "ALTER TABLE nesma_evaluations ADD COLUMN ai_analysis_id bigint"},
		{"quality_score", "ALTER TABLE nesma_evaluations ADD COLUMN quality_score double precision DEFAULT 0"},
		{"overall_grade", "ALTER TABLE nesma_evaluations ADD COLUMN overall_grade varchar(50)"},
		{"complexity_level", "ALTER TABLE nesma_evaluations ADD COLUMN complexity_level varchar(50)"},
		{"risk_level", "ALTER TABLE nesma_evaluations ADD COLUMN risk_level varchar(50)"},
		{"evaluation_summary", "ALTER TABLE nesma_evaluations ADD COLUMN evaluation_summary text"},
		{"total_function_points", "ALTER TABLE nesma_evaluations ADD COLUMN total_function_points double precision DEFAULT 0"},
		{"data_function_points", "ALTER TABLE nesma_evaluations ADD COLUMN data_function_points double precision DEFAULT 0"},
		{"transactional_fp", "ALTER TABLE nesma_evaluations ADD COLUMN transactional_fp double precision DEFAULT 0"},
		{"confidence_score", "ALTER TABLE nesma_evaluations ADD COLUMN confidence_score double precision DEFAULT 0"},
		{"accuracy_score", "ALTER TABLE nesma_evaluations ADD COLUMN accuracy_score double precision DEFAULT 0"},
		{"compliance_score", "ALTER TABLE nesma_evaluations ADD COLUMN compliance_score double precision DEFAULT 0"},
	}
	
	// 逐个检查并添加字段
	for _, field := range fieldsToAdd {
		if db.Migrator().HasColumn("nesma_evaluations", field.Name) {
			global.GVA_LOG.Info("字段已存在，跳过添加", zap.String("field", field.Name))
			continue
		}
		
		// 添加字段
		err := db.Exec(field.SQL).Error
		if err != nil {
			global.GVA_LOG.Error("添加字段失败", zap.String("field", field.Name), zap.Error(err))
			continue
		}
		
		global.GVA_LOG.Info("字段添加成功", zap.String("field", field.Name))
	}
	
	// 添加索引
	indexSQL := "CREATE INDEX IF NOT EXISTS idx_nesma_evaluations_ai_analysis_id ON nesma_evaluations(ai_analysis_id)"
	err := db.Exec(indexSQL).Error
	if err != nil {
		global.GVA_LOG.Error("添加AI分析ID索引失败", zap.Error(err))
	}
	
	global.GVA_LOG.Info("评估表字段更新完成")
}