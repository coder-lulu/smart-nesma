package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/api/v1/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type KnowledgeRouter struct{}

// InitKnowledgeRouter 初始化知识库路由
func (r *KnowledgeRouter) InitKnowledgeRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	knowledgeApi := nesma.KnowledgeApi{}

	knowledgeRouter := Router.Group("knowledge").Use(middleware.OperationRecord())
	knowledgeRouterWithoutRecord := Router.Group("knowledge")
	knowledgePublicRouter := PublicRouter.Group("knowledge")
	
	{
		// 需要记录操作的路由
		knowledgeRouter.POST("", knowledgeApi.CreateKnowledgeEntry)                     // 创建知识条目
		knowledgeRouter.PUT("/:id", knowledgeApi.UpdateKnowledgeEntry)                  // 更新知识条目
		knowledgeRouter.DELETE("/:id", knowledgeApi.DeleteKnowledgeEntry)               // 删除知识条目
		knowledgeRouter.POST("/batch", knowledgeApi.BatchCreateKnowledgeEntries)        // 批量创建知识条目
		knowledgeRouter.DELETE("/batch", knowledgeApi.BatchDeleteKnowledgeEntries)      // 批量删除知识条目
		knowledgeRouter.POST("/import", knowledgeApi.ImportKnowledgeEntries)            // 导入知识条目
	}

	{
		// 不需要记录操作的查询路由
		knowledgeRouterWithoutRecord.GET("", knowledgeApi.GetKnowledgeEntryList)                    // 获取知识条目列表
		knowledgeRouterWithoutRecord.GET("/:id", knowledgeApi.GetKnowledgeEntry)                    // 获取知识条目详情
		knowledgeRouterWithoutRecord.GET("/search", knowledgeApi.SearchKnowledgeEntries)            // 搜索知识条目
		knowledgeRouterWithoutRecord.GET("/categories", knowledgeApi.GetKnowledgeCategories)        // 获取知识分类
		knowledgeRouterWithoutRecord.GET("/tags", knowledgeApi.GetKnowledgeTags)                    // 获取知识标签
		knowledgeRouterWithoutRecord.GET("/stats", knowledgeApi.GetKnowledgeStats)                  // 获取知识库统计
		knowledgeRouterWithoutRecord.GET("/recommendations", knowledgeApi.GetKnowledgeRecommendations) // 获取知识推荐
		knowledgeRouterWithoutRecord.POST("/vector-search", knowledgeApi.VectorSearchKnowledge)     // 向量搜索
		knowledgeRouterWithoutRecord.POST("/semantic-search", knowledgeApi.SemanticSearchKnowledge) // 语义搜索
		knowledgeRouterWithoutRecord.GET("/related/:id", knowledgeApi.GetRelatedKnowledge)          // 获取相关知识
	}

	{
		// 公开路由（如果需要）
		knowledgePublicRouter.GET("/public", knowledgeApi.GetPublicKnowledgeEntries)    // 获取公开知识条目
		knowledgePublicRouter.GET("/help", knowledgeApi.GetKnowledgeHelp)               // 获取知识库帮助
	}
}