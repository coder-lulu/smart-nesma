package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type NesmaRequirementRouter struct{}

// InitNesmaRequirementRouter 初始化需求管理路由
func (r *NesmaRequirementRouter) InitNesmaRequirementRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	nesmaRequirementApi := v1.ApiGroupApp.NesmaApiGroup.NesmaRequirementApi

	// 私有路由组（需要认证）
	privateGroup := Router.Group("nesma/requirement").Use(middleware.OperationRecord())
	{
		privateGroup.POST("", nesmaRequirementApi.CreateNesmaRequirement)                            // 创建需求
		privateGroup.PUT("", nesmaRequirementApi.UpdateNesmaRequirement)                             // 更新需求
		privateGroup.DELETE("/:id", nesmaRequirementApi.DeleteNesmaRequirement)                      // 删除需求
		privateGroup.GET("/:id/children", nesmaRequirementApi.GetRequirementChildren)               // 获取需求的子功能点信息
		privateGroup.POST("/delete-by-condition", nesmaRequirementApi.DeleteRequirementsByCondition) // 根据条件删除需求
		privateGroup.GET("/:id", nesmaRequirementApi.GetNesmaRequirement)                            // 获取需求详情
		privateGroup.GET("/list", nesmaRequirementApi.GetNesmaRequirementList)                       // 获取需求列表
		privateGroup.GET("/tree", nesmaRequirementApi.GetNesmaRequirementTree)                       // 获取需求树
		privateGroup.GET("/stats", nesmaRequirementApi.GetNesmaRequirementStats)                     // 获取需求统计
		privateGroup.GET("/parent-options", nesmaRequirementApi.GetParentRequirementOptions)         // 获取父需求选项
		privateGroup.POST("/batch-delete", nesmaRequirementApi.BatchDeleteNesmaRequirements)         // 批量删除需求
		privateGroup.POST("/import-excel", nesmaRequirementApi.ImportFromExcel)                      // Excel导入
		privateGroup.POST("/batch-update-order", nesmaRequirementApi.BatchUpdateOrder)               // 批量更新排序
		privateGroup.POST("/move", nesmaRequirementApi.MoveRequirement)                              // 移动需求
		privateGroup.GET("/max-version/:projectId", nesmaRequirementApi.GetProjectMaxVersion)        // 获取项目最大版本号
		
		// AI分析相关路由
		privateGroup.POST("/ai-analysis", nesmaRequirementApi.AnalyzeRequirement)                    // 启动需求AI分析
		privateGroup.POST("/apply-analysis", nesmaRequirementApi.ApplyRequirementAnalysis)           // 应用分析结果
		privateGroup.GET("/:id/ai-analysis", nesmaRequirementApi.GetRequirementAIAnalysis)           // 获取需求AI分析结果
		privateGroup.GET("/ai-analysis-stats", nesmaRequirementApi.GetProjectAIAnalysisStats)        // 获取项目AI分析统计
		privateGroup.POST("/compare-versions", nesmaRequirementApi.CompareRequirementVersions)       // 对比需求版本
		privateGroup.PUT("/ai-status", nesmaRequirementApi.UpdateRequirementAIStatus)               // 更新需求AI分析状态
		privateGroup.GET("/ai-optimized", nesmaRequirementApi.GetAIOptimizedRequirements)           // 获取AI优化的需求列表
		
		// 需求分析进度和结果路由
		privateGroup.GET("/analysis-progress/:taskId", nesmaRequirementApi.GetAnalysisProgress)      // 获取分析进度
		privateGroup.GET("/analysis-result/:taskId", nesmaRequirementApi.GetAnalysisResult)          // 获取分析结果详情
	}
}
