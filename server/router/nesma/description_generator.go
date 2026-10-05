package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DescriptionGeneratorRouter struct{}

// InitDescriptionGeneratorRouter 初始化功能点描述生成路由
func (r *DescriptionGeneratorRouter) InitDescriptionGeneratorRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	descriptionGeneratorApi := v1.ApiGroupApp.NesmaApiGroup.DescriptionGeneratorApi

	// 私有路由组（需要认证）
	privateGroup := Router.Group("nesma/generator/description").Use(middleware.OperationRecord())
	{
		privateGroup.POST("/generate", descriptionGeneratorApi.GenerateRequirementDescriptions)       // 生成需求描述
		privateGroup.POST("/apply", descriptionGeneratorApi.ApplyGeneratedDescription)               // 应用生成的描述
		privateGroup.POST("/batch-generate", descriptionGeneratorApi.BatchGenerateDescriptions)      // 批量生成描述
		privateGroup.GET("/history/:cycleId", descriptionGeneratorApi.GetDescriptionGenerationHistory) // 获取生成历史
		privateGroup.GET("/stats/:cycleId", descriptionGeneratorApi.GetDescriptionGenerationStats)   // 获取生成统计
		privateGroup.POST("/preview", descriptionGeneratorApi.PreviewDescription)                   // 预览描述生成
		privateGroup.POST("/validate", descriptionGeneratorApi.ValidateDescription)                 // 验证描述质量
	}
}