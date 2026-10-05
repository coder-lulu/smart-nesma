package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type DocumentRouter struct{}

// InitDocumentRouter 初始化文档生成路由
func (s *DocumentRouter) InitDocumentRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	documentApi := v1.ApiGroupApp.NesmaApiGroup.DocumentApi
	templateApi := v1.ApiGroupApp.NesmaApiGroup.TemplateApi

	// 需要认证的路由
	documentRouter := Router.Group("nesma/document").Use(middleware.OperationRecord())
	{
		// 文档生成记录管理
		documentRouter.POST("", documentApi.CreateDocument)      // 创建文档生成记录
		documentRouter.PUT("", documentApi.UpdateDocument)       // 更新文档生成记录
		documentRouter.DELETE(":id", documentApi.DeleteDocument) // 删除文档生成记录
		documentRouter.GET(":id", documentApi.GetDocument)       // 获取文档详情
		documentRouter.GET("list", documentApi.GetDocumentList)  // 获取文档列表

		// 文档生成功能
		documentRouter.POST("generate", documentApi.GenerateDocument)            // 生成文档
		documentRouter.POST("batch-generate", documentApi.BatchGenerateDocument) // 批量生成文档
		documentRouter.POST("preview", documentApi.PreviewDocument)              // 预览文档
		documentRouter.GET("download", documentApi.DownloadDocument)             // 下载文档

		// 文档状态和统计
		documentRouter.GET("progress", documentApi.GetDocumentProgress) // 获取文档生成进度
		documentRouter.GET("stats", documentApi.GetDocumentStats)       // 获取文档统计

		// 批量操作
		documentRouter.POST("batch-delete", documentApi.BatchDeleteDocuments) // 批量删除文档
	}
	

	// 模板管理路由
	templateRouter := Router.Group("nesma/template").Use(middleware.OperationRecord())
	{
		// 模板CRUD
		templateRouter.POST("", templateApi.CreateTemplate)      // 创建模板
		templateRouter.PUT("", templateApi.UpdateTemplate)       // 更新模板
		templateRouter.DELETE(":id", templateApi.DeleteTemplate) // 删除模板
		templateRouter.GET(":id", templateApi.GetTemplate)       // 获取模板详情
		templateRouter.GET("list", templateApi.GetTemplateList)  // 获取模板列表

		// 模板选项和变量
		templateRouter.GET("options", templateApi.GetTemplateOptions)     // 获取模板选项
		templateRouter.GET("variables", templateApi.GetTemplateVariables) // 获取模板变量

		// 模板管理
		templateRouter.POST(":id/default", templateApi.SetDefaultTemplate) // 设置默认模板
		templateRouter.POST(":id/activate", templateApi.ActivateTemplate)  // 激活/停用模板
		templateRouter.POST("upload", templateApi.UploadTemplate)          // 上传模板文件

		// 批量操作
		templateRouter.POST("batch-delete", templateApi.BatchDeleteTemplates) // 批量删除模板
	}
}
