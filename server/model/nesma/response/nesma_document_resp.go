package response

import (
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
)

// NesmaDocumentResponse 文档响应
type NesmaDocumentResponse struct {
	nesma.NesmaDocument
	StatusText string `json:"statusText"`
	TypeText   string `json:"typeText"`
	FormatText string `json:"formatText"`
}

// NesmaDocumentListResponse 文档列表响应
type NesmaDocumentListResponse struct {
	List     []NesmaDocumentResponse `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
}

// NesmaDocTemplateResponse 模板响应
type NesmaDocTemplateResponse struct {
	nesma.NesmaDocTemplate
	TypeText     string `json:"typeText"`
	FormatText   string `json:"formatText"`
	CategoryText string `json:"categoryText"`
}

// NesmaDocTemplateListResponse 模板列表响应
type NesmaDocTemplateListResponse struct {
	List     []NesmaDocTemplateResponse `json:"list"`
	Total    int64                      `json:"total"`
	Page     int                        `json:"page"`
	PageSize int                        `json:"pageSize"`
}

// NesmaDocBatchResponse 批次响应
type NesmaDocBatchResponse struct {
	nesma.NesmaDocBatch
	StatusText string `json:"statusText"`
}

// NesmaDocBatchListResponse 批次列表响应
type NesmaDocBatchListResponse struct {
	List     []NesmaDocBatchResponse `json:"list"`
	Total    int64                   `json:"total"`
	Page     int                     `json:"page"`
	PageSize int                     `json:"pageSize"`
}

// DocumentGenerateResponse 文档生成响应
type DocumentGenerateResponse struct {
	DocumentID uint   `json:"documentId"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Message    string `json:"message"`
	FilePath   string `json:"filePath,omitempty"`
}

// BatchGenerateResponse 批量生成响应
type BatchGenerateResponse struct {
	BatchID      uint   `json:"batchId"`
	Status       string `json:"status"`
	Progress     int    `json:"progress"`
	TotalCount   int    `json:"totalCount"`
	SuccessCount int    `json:"successCount"`
	FailedCount  int    `json:"failedCount"`
	Message      string `json:"message"`
}

// DocumentPreviewResponse 文档预览响应
type DocumentPreviewResponse struct {
	PreviewURL string `json:"previewUrl"`
	Content    string `json:"content"`
	Format     string `json:"format"`
}

// DocumentDownloadResponse 文档下载响应
type DocumentDownloadResponse struct {
	FileName string `json:"fileName"`
	FilePath string `json:"filePath"`
	FileSize int64  `json:"fileSize"`
	MimeType string `json:"mimeType"`
}

// DocumentStatsResponse 文档统计响应
type DocumentStatsResponse struct {
	TotalCount         int64                  `json:"totalCount"`
	CompletedCount     int64                  `json:"completedCount"`
	PendingCount       int64                  `json:"pendingCount"`
	FailedCount        int64                  `json:"failedCount"`
	TypeStats          []DocumentTypeStats    `json:"typeStats"`
	FormatStats        []DocumentFormatStats  `json:"formatStats"`
	MonthlyStats       []DocumentMonthlyStats `json:"monthlyStats"`
	TemplateUsageStats []TemplateUsageStats   `json:"templateUsageStats"`
}

// DocumentTypeStats 文档类型统计
type DocumentTypeStats struct {
	Type     string `json:"type"`
	TypeText string `json:"typeText"`
	Count    int64  `json:"count"`
}

// DocumentFormatStats 文档格式统计
type DocumentFormatStats struct {
	Format     string `json:"format"`
	FormatText string `json:"formatText"`
	Count      int64  `json:"count"`
}

// DocumentMonthlyStats 文档月度统计
type DocumentMonthlyStats struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

// TemplateUsageStats 模板使用统计
type TemplateUsageStats struct {
	TemplateID   uint   `json:"templateId"`
	TemplateName string `json:"templateName"`
	Count        int64  `json:"count"`
}

// TemplateOptionsResponse 模板选项响应
type TemplateOptionsResponse struct {
	Options []TemplateOption `json:"options"`
}

// TemplateOption 模板选项
type TemplateOption struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	Category    string `json:"category"`
	IsDefault   bool   `json:"isDefault"`
}

// DocumentValidationResponse 文档验证响应
type DocumentValidationResponse struct {
	IsValid bool                      `json:"isValid"`
	Errors  []DocumentValidationError `json:"errors"`
}

// DocumentValidationError 文档验证错误
type DocumentValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// TemplateVariablesResponse 模板变量响应
type TemplateVariablesResponse struct {
	Variables []TemplateVariable `json:"variables"`
}

// TemplateVariable 模板变量
type TemplateVariable struct {
	Name         string      `json:"name"`
	Description  string      `json:"description"`
	Type         string      `json:"type"`
	Required     bool        `json:"required"`
	DefaultValue interface{} `json:"defaultValue"`
	Options      []string    `json:"options,omitempty"`
}

// DocumentProgressResponse 文档进度响应
type DocumentProgressResponse struct {
	DocumentID uint   `json:"documentId"`
	Status     string `json:"status"`
	Progress   int    `json:"progress"`
	Message    string `json:"message"`
	StartTime  string `json:"startTime"`
	EndTime    string `json:"endTime,omitempty"`
}

// BatchProgressResponse 批次进度响应
type BatchProgressResponse struct {
	BatchID        uint                       `json:"batchId"`
	Status         string                     `json:"status"`
	Progress       int                        `json:"progress"`
	TotalCount     int                        `json:"totalCount"`
	CompletedCount int                        `json:"completedCount"`
	FailedCount    int                        `json:"failedCount"`
	Documents      []DocumentProgressResponse `json:"documents"`
}

// UploadTemplateResponse 上传模板响应
type UploadTemplateResponse struct {
	TemplateID   uint   `json:"templateId"`
	TemplatePath string `json:"templatePath"`
	Message      string `json:"message"`
}

// ExportDocumentResponse 导出文档响应
type ExportDocumentResponse struct {
	ExportID string `json:"exportId"`
	Status   string `json:"status"`
	Message  string `json:"message"`
}
