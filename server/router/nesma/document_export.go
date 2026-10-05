package nesma

import (
	"github.com/gin-gonic/gin"
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
)

type DocumentExportRouter struct {}

// InitDocumentExportRouter 初始化文档导出路由
func (d *DocumentExportRouter) InitDocumentExportRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	documentExportRouter := Router.Group("export")
	documentExportPublicRouter := PublicRouter.Group("export")
	
	documentExportApi := v1.ApiGroupApp.NesmaApiGroup.DocumentExportApi
	
	// 私有路由（需要认证）
	{
		documentExportRouter.POST("new", documentExportApi.ExportDocument)                          // 统一文档导出接口
		documentExportRouter.GET("types", documentExportApi.GetSupportedTypes)                     // 获取支持的导出类型
		documentExportRouter.GET("download/:projectId/:exportType", documentExportApi.DownloadExportFile) // 下载导出文件
	}
	
	// 公共路由（快速导出）
	{
		documentExportPublicRouter.GET("quick", documentExportApi.QuickExport)                      // 快速导出接口
	}
}