package main

import (
	"fmt"
	"log"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
)

func main() {
	// 初始化配置
	global.GVA_VP = initialize.Viper() // 先初始化Viper
	initialize.OtherInit()
	global.GVA_LOG = initialize.Zap()
	global.GVA_DB = initialize.Gorm()

	// 执行迁移
	if err := migrateUniqueConstraint(); err != nil {
		log.Fatalf("迁移失败: %v", err)
	}

	fmt.Println("迁移完成")
}

func migrateUniqueConstraint() error {
	db := global.GVA_DB

	// 1. 删除旧的唯一性约束
	if err := db.Exec("DROP INDEX IF EXISTS idx_unique_requirement").Error; err != nil {
		return fmt.Errorf("删除旧约束失败: %v", err)
	}
	fmt.Println("删除旧约束成功")

	// 2. 创建新的唯一性约束
	if err := db.Exec("CREATE UNIQUE INDEX idx_unique_requirement_v2 ON nesma_requirements (project_id, level, title, cycle_id, version_id)").Error; err != nil {
		return fmt.Errorf("创建新约束失败: %v", err)
	}
	fmt.Println("创建新约束成功")

	// 3. 添加必要的索引
	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirements_cycle_id ON nesma_requirements (cycle_id)").Error; err != nil {
		return fmt.Errorf("创建cycle_id索引失败: %v", err)
	}
	fmt.Println("创建cycle_id索引成功")

	if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_nesma_requirements_version_id ON nesma_requirements (version_id)").Error; err != nil {
		return fmt.Errorf("创建version_id索引失败: %v", err)
	}
	fmt.Println("创建version_id索引成功")

	return nil
} 