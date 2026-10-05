package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DocumentExportApi 简化文档导出API  
type DocumentExportApi struct{}

// 使用新的导出服务
var exportService = nesmaService.NewDocumentExportService()

// ExportDocument 统一文档导出接口
// @Tags NesmaDocumentExport
// @Summary 统一文档导出接口
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body ExportRequest true "导出请求"
// @Success 200 {object} response.Response{data=ExportResult} "导出成功"
// @Router /nesma/export/new [post]
func (a *DocumentExportApi) ExportDocument(c *gin.Context) {
	var req nesmaService.ExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("请求参数错误: "+err.Error(), c)
		return
	}

	// 设置用户ID
	req.UserID = utils.GetUserID(c)

	global.GVA_LOG.Info("收到文档导出请求",
		zap.Uint("projectId", req.ProjectID),
		zap.String("exportType", string(req.ExportType)),
		zap.Uint("userId", req.UserID))

	// 调用导出服务
	result, err := exportService.Export(&req)
	if err != nil {
		global.GVA_LOG.Error("文档导出失败", zap.Error(err))
		
		// 根据错误类型返回不同的错误信息
		if exportErr, ok := err.(*nesmaService.ExportError); ok {
			switch exportErr.Code {
			case nesmaService.ErrInvalidExportType:
				response.FailWithMessage("不支持的导出类型", c)
			case nesmaService.ErrProjectNotFound:
				response.FailWithMessage("项目不存在", c)
			case nesmaService.ErrCycleNotFound:
				response.FailWithMessage("项目周期不存在", c)
			case nesmaService.ErrVersionNotFound:
				response.FailWithMessage("需求版本不存在", c)
			case nesmaService.ErrDataGeneration:
				response.FailWithMessage("数据生成失败: "+exportErr.Message, c)
			case nesmaService.ErrFileGeneration:
				response.FailWithMessage("文件生成失败: "+exportErr.Message, c)
			default:
				response.FailWithMessage("导出失败: "+exportErr.Message, c)
			}
		} else {
			response.FailWithMessage("导出失败: "+err.Error(), c)
		}
		return
	}

	response.OkWithData(result, c)
}

// GetSupportedTypes 获取支持的导出类型
// @Tags NesmaDocumentExport
// @Summary 获取支持的导出类型
// @Security ApiKeyAuth
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]string} "获取成功"
// @Router /nesma/export/types [get]
func (a *DocumentExportApi) GetSupportedTypes(c *gin.Context) {
	types := exportService.GetExportTypeInfo()
	response.OkWithData(types, c)
}

// DownloadExportFile 下载导出文件
// @Tags NesmaDocumentExport
// @Summary 下载导出文件
// @Security ApiKeyAuth
// @Param projectId path int true "项目ID"
// @Param exportType path string true "导出类型"
// @Success 200 {file} binary "文件下载"
// @Router /nesma/export/download/{projectId}/{exportType} [get]
func (a *DocumentExportApi) DownloadExportFile(c *gin.Context) {
	projectIDStr := c.Param("projectId")
	exportTypeStr := c.Param("exportType")

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	exportType := nesmaService.ExportType(exportTypeStr)
	if !exportType.IsValid() {
		response.FailWithMessage("不支持的导出类型", c)
		return
	}

	// 获取文件路径
	filePath, err := exportService.DownloadFile(uint(projectID), exportType)
	if err != nil {
		global.GVA_LOG.Error("获取下载文件失败", zap.Error(err))
		response.FailWithMessage("文件不存在或未生成完成", c)
		return
	}

	// 生成文件名
	fileName := exportType.GetFileName("项目文档")

	global.GVA_LOG.Info("开始下载文件",
		zap.String("filePath", filePath),
		zap.String("fileName", fileName))

	// 设置响应头
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(filePath)
}

// QuickExport 快速导出接口 - 简化版，用于快速测试
// @Tags NesmaDocumentExport
// @Summary 快速导出接口
// @Security ApiKeyAuth
// @Param projectId query int true "项目ID"
// @Param type query string true "导出类型" Enums(word_requirement_spec,excel_nesma_report,excel_business_summary)
// @Success 200 {file} binary "文件下载"
// @Router /nesma/export/quick [get]
func (a *DocumentExportApi) QuickExport(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	exportTypeStr := c.Query("type")

	if projectIDStr == "" || exportTypeStr == "" {
		response.FailWithMessage("缺少必要参数", c)
		return
	}

	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	exportType := nesmaService.ExportType(exportTypeStr)
	if !exportType.IsValid() {
		response.FailWithMessage("不支持的导出类型", c)
		return
	}

	// 创建默认导出请求
	req := &nesmaService.ExportRequest{
		ProjectID:  uint(projectID),
		ExportType: exportType,
		UserID:     utils.GetUserID(c),
		Config: nesmaService.ExportConfig{
			IncludeAIAnalysis:  true,
			IncludeMermaidDiag: true,
			IncludeStatistics:  true,
			WordConfig: nesmaService.WordExportConfig{
				IncludeCoverPage:       true,
				IncludeTableOfContents: true,
				FontSize:               12,
				LineSpacing:            "1.5",
				DetailLevel:            4,
			},
			ExcelConfig: nesmaService.ExcelExportConfig{
				IncludeHeader:    true,
				IncludeStatSheet: true,
				FreezeHeader:     true,
				AutoColumnWidth:  true,
			},
		},
	}

	global.GVA_LOG.Info("收到快速导出请求",
		zap.Uint("projectId", req.ProjectID),
		zap.String("exportType", string(req.ExportType)))

	// 同步导出
	result, err := exportService.Export(req)
	if err != nil {
		global.GVA_LOG.Error("快速导出失败", zap.Error(err))
		response.FailWithMessage("导出失败: "+err.Error(), c)
		return
	}

	if result.Status != "completed" {
		response.FailWithMessage("导出未完成: "+result.Message, c)
		return
	}

	// 直接下载文件
	fileName := exportType.GetFileName("快速导出")
	
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")

	c.File(result.FilePath)
}