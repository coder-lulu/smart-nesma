package request

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/request"
	"gorm.io/datatypes"
)

// CreateNesmaDocumentRequest 创建文档生成请求
type CreateNesmaDocumentRequest struct {
	ProjectID   uint           `json:"projectId" binding:"required" comment:"项目ID"`
	CycleID     *uint          `json:"cycleId" comment:"项目周期ID"`
	VersionID   *uint          `json:"versionId" comment:"需求版本ID"`
	TemplateType string         `json:"templateType" comment:"模板类型"`
	TemplateID   string         `json:"templateId" binding:"required" comment:"模板ID"`
	Name        string         `json:"name" binding:"required" comment:"文档名称"`
	Description string         `json:"description" comment:"文档描述"`
	Type        string         `json:"type" binding:"required,oneof=word excel pdf" comment:"文档类型"`
	Format      string         `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"文档格式"`
	Config      datatypes.JSON `json:"config" comment:"生成配置"`
	Version     string         `json:"version" comment:"文档版本"`
}

// UpdateNesmaDocumentRequest 更新文档生成请求
type UpdateNesmaDocumentRequest struct {
	ID          uint           `json:"id" binding:"required" comment:"文档ID"`
	ProjectID   uint           `json:"projectId" binding:"required" comment:"项目ID"`
	CycleID     *uint          `json:"cycleId" comment:"项目周期ID"`
	VersionID   *uint          `json:"versionId" comment:"需求版本ID"`
	TemplateID  string         `json:"templateId" binding:"required" comment:"模板ID"`
	Name        string         `json:"name" binding:"required" comment:"文档名称"`
	Description string         `json:"description" comment:"文档描述"`
	Type        string         `json:"type" binding:"required,oneof=word excel pdf" comment:"文档类型"`
	Format      string         `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"文档格式"`
	Config      datatypes.JSON `json:"config" comment:"生成配置"`
	Version     string         `json:"version" comment:"文档版本"`
}

// NesmaDocumentSearch 文档搜索请求
type NesmaDocumentSearch struct {
	request.PageInfo
	ProjectID  *uint  `json:"projectId" form:"projectId" comment:"项目ID"`
	CycleID    *uint  `json:"cycleId" form:"cycleId" comment:"项目周期ID"`
	VersionID  *uint  `json:"versionId" form:"versionId" comment:"需求版本ID"`
	TemplateID *uint  `json:"templateId" form:"templateId" comment:"模板ID"`
	Type       string `json:"type" form:"type" comment:"文档类型"`
	Format     string `json:"format" form:"format" comment:"文档格式"`
	Status     string `json:"status" form:"status" comment:"状态"`
	Keyword    string `json:"keyword" form:"keyword" comment:"关键词"`
	CreatedBy  *uint  `json:"createdBy" form:"createdBy" comment:"创建人ID"`
}

// GenerateDocumentRequest 生成文档请求
type GenerateDocumentRequest struct {
	DocumentID uint           `json:"documentId" binding:"required" comment:"文档ID"`
	Config     datatypes.JSON `json:"config" comment:"生成配置"`
	Async      bool           `json:"async" comment:"是否异步生成"`
}

// BatchGenerateDocumentRequest 批量生成文档请求
type BatchGenerateDocumentRequest struct {
	ProjectID   uint             `json:"projectId" binding:"required" comment:"项目ID"`
	Name        string           `json:"name" binding:"required" comment:"批次名称"`
	Description string           `json:"description" comment:"批次描述"`
	Documents   []DocumentConfig `json:"documents" binding:"required" comment:"文档配置列表"`
	Config      datatypes.JSON   `json:"config" comment:"批次配置"`
}

// DocumentConfig 文档配置
type DocumentConfig struct {
	TemplateID  string         `json:"templateId" binding:"required" comment:"模板ID"`
	Name        string         `json:"name" binding:"required" comment:"文档名称"`
	Description string         `json:"description" comment:"文档描述"`
	Type        string         `json:"type" binding:"required,oneof=word excel pdf" comment:"文档类型"`
	Format      string         `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"文档格式"`
	Config      datatypes.JSON `json:"config" comment:"生成配置"`
	Version     string         `json:"version" comment:"文档版本"`
}

// CreateNesmaDocTemplateRequest 创建文档模板请求
type CreateNesmaDocTemplateRequest struct {
	Name         string         `json:"name" binding:"required" comment:"模板名称"`
	Description  string         `json:"description" comment:"模板描述"`
	Type         string         `json:"type" binding:"required,oneof=word excel pdf" comment:"模板类型"`
	Format       string         `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"模板格式"`
	Category     string         `json:"category" comment:"模板分类"`
	TemplatePath string         `json:"templatePath" comment:"模板文件路径"`
	PreviewPath  string         `json:"previewPath" comment:"预览图路径"`
	Config       datatypes.JSON `json:"config" comment:"模板配置"`
	Variables    datatypes.JSON `json:"variables" comment:"模板变量"`
	IsDefault    bool           `json:"isDefault" comment:"是否默认模板"`
	IsActive     bool           `json:"isActive" comment:"是否激活"`
}

// UpdateNesmaDocTemplateRequest 更新文档模板请求
type UpdateNesmaDocTemplateRequest struct {
	ID           uint           `json:"id" binding:"required" comment:"模板ID"`
	Name         string         `json:"name" binding:"required" comment:"模板名称"`
	Description  string         `json:"description" comment:"模板描述"`
	Type         string         `json:"type" binding:"required,oneof=word excel pdf" comment:"模板类型"`
	Format       string         `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"模板格式"`
	Category     string         `json:"category" comment:"模板分类"`
	TemplatePath string         `json:"templatePath" comment:"模板文件路径"`
	PreviewPath  string         `json:"previewPath" comment:"预览图路径"`
	Config       datatypes.JSON `json:"config" comment:"模板配置"`
	Variables    datatypes.JSON `json:"variables" comment:"模板变量"`
	IsDefault    bool           `json:"isDefault" comment:"是否默认模板"`
	IsActive     bool           `json:"isActive" comment:"是否激活"`
}

// NesmaDocTemplateSearch 模板搜索请求
type NesmaDocTemplateSearch struct {
	request.PageInfo
	Type      string `json:"type" form:"type" comment:"模板类型"`
	Format    string `json:"format" form:"format" comment:"模板格式"`
	Category  string `json:"category" form:"category" comment:"模板分类"`
	IsDefault *bool  `json:"isDefault" form:"isDefault" comment:"是否默认模板"`
	IsActive  *bool  `json:"isActive" form:"isActive" comment:"是否激活"`
	Keyword   string `json:"keyword" form:"keyword" comment:"关键词"`
	CreatedBy *uint  `json:"createdBy" form:"createdBy" comment:"创建人ID"`
}

// NesmaDocBatchSearch 批次搜索请求
type NesmaDocBatchSearch struct {
	request.PageInfo
	ProjectID *uint  `json:"projectId" form:"projectId" comment:"项目ID"`
	Status    string `json:"status" form:"status" comment:"状态"`
	Keyword   string `json:"keyword" form:"keyword" comment:"关键词"`
	CreatedBy *uint  `json:"createdBy" form:"createdBy" comment:"创建人ID"`
}

// DocumentIdsRequest 文档ID列表请求
type DocumentIdsRequest struct {
	IDs []uint `json:"ids" binding:"required" comment:"文档ID列表"`
}

// TemplateIdsRequest 模板ID列表请求
type TemplateIdsRequest struct {
	IDs []uint `json:"ids" binding:"required" comment:"模板ID列表"`
}

// BatchIdsRequest 批次ID列表请求
type BatchIdsRequest struct {
	IDs []uint `json:"ids" binding:"required" comment:"批次ID列表"`
}

// PreviewDocumentRequest 预览文档请求
type PreviewDocumentRequest struct {
	ProjectID  uint           `json:"projectId" binding:"required" comment:"项目ID"`
	TemplateID string         `json:"templateId" binding:"required" comment:"模板ID"`
	Type       string         `json:"type" binding:"required,oneof=word excel pdf" comment:"文档类型"`
	Format     string         `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"文档格式"`
	Config     datatypes.JSON `json:"config" comment:"生成配置"`
}

// DownloadDocumentRequest 下载文档请求
type DownloadDocumentRequest struct {
	DocumentID uint `json:"documentId" binding:"required" comment:"文档ID"`
}

// UploadTemplateRequest 上传模板请求
type UploadTemplateRequest struct {
	Name        string `json:"name" binding:"required" comment:"模板名称"`
	Description string `json:"description" comment:"模板描述"`
	Type        string `json:"type" binding:"required,oneof=word excel pdf" comment:"模板类型"`
	Format      string `json:"format" binding:"required,oneof=requirement_spec nesma_report business_summary" comment:"模板格式"`
	Category    string `json:"category" comment:"模板分类"`
	IsDefault   bool   `json:"isDefault" comment:"是否默认模板"`
}
