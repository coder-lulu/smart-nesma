package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type DocumentApi struct{}

var documentService = service.ServiceGroupApp.NesmaServiceGroup.DocumentService

// CreateDocument 创建文档生成记录
// @Tags NesmaDocument
// @Summary 创建文档生成记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateNesmaDocumentRequest true "创建文档生成记录"
// @Success 200 {object} response.Response{data=nesma.NesmaDocument} "创建成功"
// @Router /nesma/document [post]
func (a *DocumentApi) CreateDocument(c *gin.Context) {
	var req nesmaReq.CreateNesmaDocumentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	document, err := documentService.CreateDocument(&req, userID)
	if err != nil {
		global.GVA_LOG.Error("创建文档生成记录失败", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	response.OkWithData(document, c)
}

// UpdateDocument 更新文档生成记录
// @Tags NesmaDocument
// @Summary 更新文档生成记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.UpdateNesmaDocumentRequest true "更新文档生成记录"
// @Success 200 {object} response.Response{data=nesma.NesmaDocument} "更新成功"
// @Router /nesma/document [put]
func (a *DocumentApi) UpdateDocument(c *gin.Context) {
	var req nesmaReq.UpdateNesmaDocumentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	document, err := documentService.UpdateDocument(&req)
	if err != nil {
		global.GVA_LOG.Error("更新文档生成记录失败", zap.Error(err))
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}

	response.OkWithData(document, c)
}

// DeleteDocument 删除文档生成记录
// @Tags NesmaDocument
// @Summary 删除文档生成记录
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "文档ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/document/{id} [delete]
func (a *DocumentApi) DeleteDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	err = documentService.DeleteDocument(uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除文档生成记录失败", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

// GetDocument 获取文档详情
// @Tags NesmaDocument
// @Summary 获取文档详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "文档ID"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaDocumentResponse} "获取成功"
// @Router /nesma/document/{id} [get]
func (a *DocumentApi) GetDocument(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	document, err := documentService.GetDocument(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取文档详情失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(document, c)
}

// GetDocumentList 获取文档列表
// @Tags NesmaDocument
// @Summary 获取文档列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query nesmaReq.NesmaDocumentSearch true "获取文档列表"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaDocumentListResponse} "获取成功"
// @Router /nesma/document/list [get]
func (a *DocumentApi) GetDocumentList(c *gin.Context) {
	var req nesmaReq.NesmaDocumentSearch
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	documents, err := documentService.GetDocumentList(&req)
	if err != nil {
		global.GVA_LOG.Error("获取文档列表失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(documents, c)
}

// GenerateDocument 生成文档
// @Tags NesmaDocument
// @Summary 生成文档
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.GenerateDocumentRequest true "生成文档"
// @Success 200 {object} response.Response{data=nesmaRes.DocumentGenerateResponse} "生成成功"
// @Router /nesma/document/generate [post]
func (a *DocumentApi) GenerateDocument(c *gin.Context) {
	var req nesmaReq.GenerateDocumentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := documentService.GenerateDocument(&req)
	if err != nil {
		global.GVA_LOG.Error("生成文档失败", zap.Error(err))
		response.FailWithMessage("生成失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// BatchGenerateDocument 批量生成文档
// @Tags NesmaDocument
// @Summary 批量生成文档
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.BatchGenerateDocumentRequest true "批量生成文档"
// @Success 200 {object} response.Response{data=nesmaRes.BatchGenerateResponse} "生成成功"
// @Router /nesma/document/batch-generate [post]
func (a *DocumentApi) BatchGenerateDocument(c *gin.Context) {
	var req nesmaReq.BatchGenerateDocumentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	result, err := documentService.BatchGenerateDocument(&req, userID)
	if err != nil {
		global.GVA_LOG.Error("批量生成文档失败", zap.Error(err))
		response.FailWithMessage("生成失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// PreviewDocument 预览文档
// @Tags NesmaDocument
// @Summary 预览文档
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.PreviewDocumentRequest true "预览文档"
// @Success 200 {object} response.Response{data=nesmaRes.DocumentPreviewResponse} "预览成功"
// @Router /nesma/document/preview [post]
func (a *DocumentApi) PreviewDocument(c *gin.Context) {
	var req nesmaReq.PreviewDocumentRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	result, err := documentService.PreviewDocument(&req)
	if err != nil {
		global.GVA_LOG.Error("预览文档失败", zap.Error(err))
		response.FailWithMessage("预览失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// DownloadDocument 下载文档
// @Tags NesmaDocument
// @Summary 下载文档
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param documentId query int true "文档ID"
// @Success 200 {file} file "文档文件"
// @Router /nesma/document/download [get]
func (a *DocumentApi) DownloadDocument(c *gin.Context) {
	documentIDStr := c.Query("documentId")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	req := nesmaReq.DownloadDocumentRequest{
		DocumentID: uint(documentID),
	}

	err = documentService.DownloadDocument(&req, c)
	if err != nil {
		global.GVA_LOG.Error("下载文档失败", zap.Error(err))
		response.FailWithMessage("下载失败: "+err.Error(), c)
		return
	}
}

// GetDocumentProgress 获取文档生成进度
// @Tags NesmaDocument
// @Summary 获取文档生成进度
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param documentId query int true "文档ID"
// @Success 200 {object} response.Response{data=nesmaRes.DocumentProgressResponse} "获取成功"
// @Router /nesma/document/progress [get]
func (a *DocumentApi) GetDocumentProgress(c *gin.Context) {
	documentIDStr := c.Query("documentId")
	documentID, err := strconv.ParseUint(documentIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	progress, err := documentService.GetDocumentProgress(uint(documentID))
	if err != nil {
		global.GVA_LOG.Error("获取文档生成进度失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(progress, c)
}

// GetDocumentStats 获取文档统计
// @Tags NesmaDocument
// @Summary 获取文档统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId query int false "项目ID"
// @Success 200 {object} response.Response{data=nesmaRes.DocumentStatsResponse} "获取成功"
// @Router /nesma/document/stats [get]
func (a *DocumentApi) GetDocumentStats(c *gin.Context) {
	projectIDStr := c.Query("projectId")
	var projectID *uint
	if projectIDStr != "" {
		id, err := strconv.ParseUint(projectIDStr, 10, 32)
		if err != nil {
			response.FailWithMessage("项目ID参数错误", c)
			return
		}
		pid := uint(id)
		projectID = &pid
	}

	stats, err := documentService.GetDocumentStats(projectID)
	if err != nil {
		global.GVA_LOG.Error("获取文档统计失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(stats, c)
}

// BatchDeleteDocuments 批量删除文档
// @Tags NesmaDocument
// @Summary 批量删除文档
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.DocumentIdsRequest true "批量删除文档"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/document/batch-delete [post]
func (a *DocumentApi) BatchDeleteDocuments(c *gin.Context) {
	var req nesmaReq.DocumentIdsRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = documentService.BatchDeleteDocuments(&req)
	if err != nil {
		global.GVA_LOG.Error("批量删除文档失败", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}
