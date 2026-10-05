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

type TemplateApi struct{}

var templateService = service.ServiceGroupApp.NesmaServiceGroup.TemplateService

// CreateTemplate 创建文档模板
// @Tags NesmaTemplate
// @Summary 创建文档模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.CreateNesmaDocTemplateRequest true "创建文档模板"
// @Success 200 {object} response.Response{data=nesma.NesmaDocTemplate} "创建成功"
// @Router /nesma/template [post]
func (a *TemplateApi) CreateTemplate(c *gin.Context) {
	var req nesmaReq.CreateNesmaDocTemplateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	template, err := templateService.CreateTemplate(&req, userID)
	if err != nil {
		global.GVA_LOG.Error("创建文档模板失败", zap.Error(err))
		response.FailWithMessage("创建失败: "+err.Error(), c)
		return
	}

	response.OkWithData(template, c)
}

// UpdateTemplate 更新文档模板
// @Tags NesmaTemplate
// @Summary 更新文档模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.UpdateNesmaDocTemplateRequest true "更新文档模板"
// @Success 200 {object} response.Response{data=nesma.NesmaDocTemplate} "更新成功"
// @Router /nesma/template [put]
func (a *TemplateApi) UpdateTemplate(c *gin.Context) {
	var req nesmaReq.UpdateNesmaDocTemplateRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	template, err := templateService.UpdateTemplate(&req)
	if err != nil {
		global.GVA_LOG.Error("更新文档模板失败", zap.Error(err))
		response.FailWithMessage("更新失败: "+err.Error(), c)
		return
	}

	response.OkWithData(template, c)
}

// DeleteTemplate 删除文档模板
// @Tags NesmaTemplate
// @Summary 删除文档模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "模板ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/template/{id} [delete]
func (a *TemplateApi) DeleteTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	err = templateService.DeleteTemplate(uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除文档模板失败", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}

// GetTemplate 获取模板详情
// @Tags NesmaTemplate
// @Summary 获取模板详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "模板ID"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaDocTemplateResponse} "获取成功"
// @Router /nesma/template/{id} [get]
func (a *TemplateApi) GetTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	template, err := templateService.GetTemplate(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取模板详情失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(template, c)
}

// GetTemplateList 获取模板列表
// @Tags NesmaTemplate
// @Summary 获取模板列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query nesmaReq.NesmaDocTemplateSearch true "获取模板列表"
// @Success 200 {object} response.Response{data=nesmaRes.NesmaDocTemplateListResponse} "获取成功"
// @Router /nesma/template/list [get]
func (a *TemplateApi) GetTemplateList(c *gin.Context) {
	var req nesmaReq.NesmaDocTemplateSearch
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

	templates, err := templateService.GetTemplateList(&req)
	if err != nil {
		global.GVA_LOG.Error("获取模板列表失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(templates, c)
}

// GetTemplateOptions 获取模板选项
// @Tags NesmaTemplate
// @Summary 获取模板选项
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param type query string false "文档类型"
// @Param format query string false "文档格式"
// @Success 200 {object} response.Response{data=nesmaRes.TemplateOptionsResponse} "获取成功"
// @Router /nesma/template/options [get]
func (a *TemplateApi) GetTemplateOptions(c *gin.Context) {
	docType := c.Query("type")
	format := c.Query("format")

	options, err := templateService.GetTemplateOptions(docType, format)
	if err != nil {
		global.GVA_LOG.Error("获取模板选项失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(options, c)
}

// GetTemplateVariables 获取模板变量
// @Tags NesmaTemplate
// @Summary 获取模板变量
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param templateId query int true "模板ID"
// @Success 200 {object} response.Response{data=nesmaRes.TemplateVariablesResponse} "获取成功"
// @Router /nesma/template/variables [get]
func (a *TemplateApi) GetTemplateVariables(c *gin.Context) {
	templateIDStr := c.Query("templateId")
	templateID, err := strconv.ParseUint(templateIDStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	variables, err := templateService.GetTemplateVariables(uint(templateID))
	if err != nil {
		global.GVA_LOG.Error("获取模板变量失败", zap.Error(err))
		response.FailWithMessage("获取失败: "+err.Error(), c)
		return
	}

	response.OkWithData(variables, c)
}

// SetDefaultTemplate 设置默认模板
// @Tags NesmaTemplate
// @Summary 设置默认模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "模板ID"
// @Success 200 {object} response.Response "设置成功"
// @Router /nesma/template/{id}/default [post]
func (a *TemplateApi) SetDefaultTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	err = templateService.SetDefaultTemplate(uint(id))
	if err != nil {
		global.GVA_LOG.Error("设置默认模板失败", zap.Error(err))
		response.FailWithMessage("设置失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("设置成功", c)
}

// ActivateTemplate 激活/停用模板
// @Tags NesmaTemplate
// @Summary 激活/停用模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "模板ID"
// @Param isActive query bool true "是否激活"
// @Success 200 {object} response.Response "操作成功"
// @Router /nesma/template/{id}/activate [post]
func (a *TemplateApi) ActivateTemplate(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("参数错误", c)
		return
	}

	isActiveStr := c.Query("isActive")
	isActive := isActiveStr == "true"

	err = templateService.ActivateTemplate(uint(id), isActive)
	if err != nil {
		global.GVA_LOG.Error("激活/停用模板失败", zap.Error(err))
		response.FailWithMessage("操作失败: "+err.Error(), c)
		return
	}

	message := "停用成功"
	if isActive {
		message = "激活成功"
	}
	response.OkWithMessage(message, c)
}

// UploadTemplate 上传模板文件
// @Tags NesmaTemplate
// @Summary 上传模板文件
// @Security ApiKeyAuth
// @accept multipart/form-data
// @Produce application/json
// @Param file formData file true "模板文件"
// @Param name formData string true "模板名称"
// @Param description formData string false "模板描述"
// @Param type formData string true "模板类型"
// @Param format formData string true "模板格式"
// @Param category formData string false "模板分类"
// @Param isDefault formData bool false "是否默认模板"
// @Success 200 {object} response.Response{data=nesmaRes.UploadTemplateResponse} "上传成功"
// @Router /nesma/template/upload [post]
func (a *TemplateApi) UploadTemplate(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		response.FailWithMessage("文件上传失败: "+err.Error(), c)
		return
	}

	// 获取表单参数
	req := nesmaReq.UploadTemplateRequest{
		Name:        c.PostForm("name"),
		Description: c.PostForm("description"),
		Type:        c.PostForm("type"),
		Format:      c.PostForm("format"),
		Category:    c.PostForm("category"),
		IsDefault:   c.PostForm("isDefault") == "true",
	}

	// 验证必填字段
	if req.Name == "" || req.Type == "" || req.Format == "" {
		response.FailWithMessage("模板名称、类型和格式为必填项", c)
		return
	}

	// 保存上传的文件
	uploadPath := global.GVA_CONFIG.Local.StorePath + "/temp/" + file.Filename
	err = c.SaveUploadedFile(file, uploadPath)
	if err != nil {
		global.GVA_LOG.Error("保存上传文件失败", zap.Error(err))
		response.FailWithMessage("保存文件失败: "+err.Error(), c)
		return
	}

	userID := utils.GetUserID(c)
	result, err := templateService.UploadTemplate(&req, uploadPath, userID)
	if err != nil {
		global.GVA_LOG.Error("上传模板失败", zap.Error(err))
		response.FailWithMessage("上传失败: "+err.Error(), c)
		return
	}

	response.OkWithData(result, c)
}

// BatchDeleteTemplates 批量删除模板
// @Tags NesmaTemplate
// @Summary 批量删除模板
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.TemplateIdsRequest true "批量删除模板"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/template/batch-delete [post]
func (a *TemplateApi) BatchDeleteTemplates(c *gin.Context) {
	var req nesmaReq.TemplateIdsRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = templateService.BatchDeleteTemplates(&req)
	if err != nil {
		global.GVA_LOG.Error("批量删除模板失败", zap.Error(err))
		response.FailWithMessage("删除失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除成功", c)
}
