package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
)

// NesmaEvaluationFactors NESMA评估因子配置表
type NesmaEvaluationFactors struct {
	global.GVA_MODEL
	EvaluationID  uint   `json:"evaluation_id" gorm:"not null;comment:评估ID;index:idx_evaluation_factors_evaluation_id"`
	FactorConfig  string `json:"factor_config" gorm:"type:text;not null;comment:因子配置JSON"`
	NesmaVersion  string `json:"nesma_version" gorm:"type:varchar(20);default:'v2.2';comment:NESMA版本;index:idx_evaluation_factors_nesma_version"`
}

// TableName 设置表名
func (NesmaEvaluationFactors) TableName() string {
	return "nesma_evaluation_factors"
}