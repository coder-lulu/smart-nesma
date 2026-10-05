package nesma

import (
	v1 "github.com/flipped-aurora/gin-vue-admin/server/api/v1"
	"github.com/flipped-aurora/gin-vue-admin/server/middleware"
	"github.com/gin-gonic/gin"
)

type NesmaProjectRouter struct{}

// InitNesmaProjectRouter 初始化 NESMA项目 路由信息
func (s *NesmaProjectRouter) InitNesmaProjectRouter(Router *gin.RouterGroup, PublicRouter *gin.RouterGroup) {
	nesmaProjectRouter := Router.Group("nesma/project").Use(middleware.OperationRecord())
	nesmaProjectRouterWithoutRecord := Router.Group("nesma/project")

	var nesmaProjectApi = v1.ApiGroupApp.NesmaApiGroup.NesmaProjectApi
	{
		nesmaProjectRouter.POST("", nesmaProjectApi.CreateNesmaProject)                    // 新建NESMA项目
		nesmaProjectRouter.DELETE("", nesmaProjectApi.DeleteNesmaProject)                  // 删除NESMA项目
		nesmaProjectRouter.DELETE("delete-batch", nesmaProjectApi.DeleteNesmaProjectByIds) // 批量删除NESMA项目
		nesmaProjectRouter.PUT("", nesmaProjectApi.UpdateNesmaProject)                     // 更新NESMA项目
		nesmaProjectRouter.POST("archive", nesmaProjectApi.ArchiveNesmaProject)            // 归档项目
		nesmaProjectRouter.POST("restore", nesmaProjectApi.RestoreNesmaProject)            // 恢复项目
		nesmaProjectRouter.POST("one-click-analysis", nesmaProjectApi.OneClickAnalysis)    // 一键分析项目
	}
	{
		nesmaProjectRouterWithoutRecord.GET("options", nesmaProjectApi.GetProjectOptions)  // 获取项目选项列表（必须在:ID之前）
		nesmaProjectRouterWithoutRecord.GET("list", nesmaProjectApi.GetNesmaProjectList)   // 获取NESMA项目列表
		nesmaProjectRouterWithoutRecord.GET("stats", nesmaProjectApi.GetNesmaProjectStats) // 获取项目统计信息
		nesmaProjectRouterWithoutRecord.GET(":ID", nesmaProjectApi.FindNesmaProject)       // 根据ID获取NESMA项目（必须放在最后）
	}
}
