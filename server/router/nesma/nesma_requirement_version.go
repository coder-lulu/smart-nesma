package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type NesmaRequirementVersionRouter struct{}

// InitNesmaRequirementVersionRouter 初始化需求版本管理路由
func (r *NesmaRequirementVersionRouter) InitNesmaRequirementVersionRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	versionApi := v1.ApiGroupApp.NesmaApiGroup.NesmaRequirementVersionApi

	// 私有路由组（需要认证）
	privateGroup := Router.Group("nesma/requirement-version").Use(middleware.OperationRecord())
	{
		privateGroup.POST("", versionApi.CreateRequirementVersion)              // 创建需求版本
		privateGroup.PUT("", versionApi.UpdateRequirementVersion)               // 更新需求版本
		privateGroup.DELETE("/:id", versionApi.DeleteRequirementVersion)        // 删除需求版本
		privateGroup.GET("/:id", versionApi.GetRequirementVersion)              // 获取版本详情
		privateGroup.GET("/cycle/:cycleId", versionApi.GetRequirementVersions)  // 获取周期的所有版本
		privateGroup.GET("/active/:cycleId", versionApi.GetActiveVersion)       // 获取周期的激活版本
		privateGroup.PUT("/active/:id", versionApi.SetActiveVersion)            // 设置激活版本
		privateGroup.GET("/next/:cycleId", versionApi.GenerateNextVersion)      // 生成下一个版本号
		privateGroup.PUT("/stats/:id", versionApi.UpdateVersionStats)           // 更新版本统计信息
	}
} 