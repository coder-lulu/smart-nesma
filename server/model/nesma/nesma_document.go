package nesma

import (
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"gorm.io/datatypes"
)

// NesmaDocument 文档生成记录表
type NesmaDocument struct {
	global.GVA_MODEL
	ProjectID   uint           `json:"projectId" gorm:"not null;comment:项目ID"`
	CycleID     *uint          `json:"cycleId" gorm:"comment:项目周期ID"`
	VersionID   *uint          `json:"versionId" gorm:"comment:需求版本ID"`
	TemplateID  string         `json:"templateId" gorm:"type:varchar(100);not null;comment:模板ID（支持内置模板如builtin_excel）"`
	BatchID     *uint          `json:"batchId" gorm:"comment:批次ID（可选）"`
	Name        string         `json:"name" gorm:"type:varchar(255);not null;comment:文档名称"`
	Description string         `json:"description" gorm:"type:text;comment:文档描述"`
	Type        string         `json:"type" gorm:"type:varchar(50);not null;comment:文档类型：word/excel/pdf"`
	Format      string         `json:"format" gorm:"type:varchar(50);not null;comment:文档格式：requirement_spec/nesma_report/business_summary"`
	FilePath    string         `json:"filePath" gorm:"type:varchar(500);comment:文件路径"`
	FileSize    int64          `json:"fileSize" gorm:"comment:文件大小"`
	Status      string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:生成状态：pending/generating/completed/failed"`
	Progress    int            `json:"progress" gorm:"default:0;comment:生成进度0-100"`
	ErrorMsg    string         `json:"errorMsg" gorm:"type:text;comment:错误信息"`
	Config      datatypes.JSON `json:"config" gorm:"type:json;comment:生成配置"`
	Version     string         `json:"version" gorm:"type:varchar(50);comment:文档版本"`
	GeneratedAt *time.Time     `json:"generatedAt" gorm:"comment:生成完成时间"`
	CreatedBy   uint           `json:"createdBy" gorm:"not null;comment:创建人ID"`

	// 关联关系
	Project         NesmaProject           `json:"project" gorm:"foreignKey:ProjectID"`
	Cycle           *NesmaProjectCycle     `json:"cycle" gorm:"foreignKey:CycleID"`
	RequirementVersion *NesmaRequirementVersion `json:"requirementVersion" gorm:"foreignKey:VersionID"`
	// Template NesmaDocTemplate       `json:"template" gorm:"foreignKey:TemplateID"` // 暂时注释，因为现在使用内置模板
	Batch           *NesmaDocBatch         `json:"batch" gorm:"foreignKey:BatchID"`
}

// NesmaDocTemplate 文档模板表
type NesmaDocTemplate struct {
	global.GVA_MODEL
	Name         string         `json:"name" gorm:"type:varchar(255);not null;comment:模板名称"`
	Description  string         `json:"description" gorm:"type:text;comment:模板描述"`
	Type         string         `json:"type" gorm:"type:varchar(50);not null;comment:模板类型：word/excel/pdf"`
	Format       string         `json:"format" gorm:"type:varchar(50);not null;comment:模板格式：requirement_spec/nesma_report/business_summary"`
	Category     string         `json:"category" gorm:"type:varchar(100);comment:模板分类"`
	Content      string         `json:"content" gorm:"type:text;comment:模板内容"`
	TemplatePath string         `json:"templatePath" gorm:"type:varchar(500);comment:模板文件路径"`
	PreviewPath  string         `json:"previewPath" gorm:"type:varchar(500);comment:预览图路径"`
	Config       datatypes.JSON `json:"config" gorm:"type:json;comment:模板配置"`
	Variables    datatypes.JSON `json:"variables" gorm:"type:json;comment:模板变量"`
	IsDefault    bool           `json:"isDefault" gorm:"default:false;comment:是否默认模板"`
	IsActive     bool           `json:"isActive" gorm:"default:true;comment:是否激活"`
	UsageCount   int            `json:"usageCount" gorm:"default:0;comment:使用次数"`
	CreatedBy    uint           `json:"createdBy" gorm:"not null;comment:创建人ID"`

	// 关联关系
	Documents []NesmaDocument `json:"documents" gorm:"foreignKey:TemplateID"`
}

// NesmaDocBatch 批量文档生成记录表
type NesmaDocBatch struct {
	global.GVA_MODEL
	ProjectID    uint           `json:"projectId" gorm:"not null;comment:项目ID"`
	Name         string         `json:"name" gorm:"type:varchar(255);not null;comment:批次名称"`
	Description  string         `json:"description" gorm:"type:text;comment:批次描述"`
	TotalCount   int            `json:"totalCount" gorm:"not null;comment:总文档数"`
	SuccessCount int            `json:"successCount" gorm:"default:0;comment:成功数量"`
	FailedCount  int            `json:"failedCount" gorm:"default:0;comment:失败数量"`
	Status       string         `json:"status" gorm:"type:varchar(50);default:'pending';comment:批次状态：pending/processing/completed/failed"`
	Progress     int            `json:"progress" gorm:"default:0;comment:批次进度0-100"`
	Config       datatypes.JSON `json:"config" gorm:"type:json;comment:批次配置"`
	StartedAt    *time.Time     `json:"startedAt" gorm:"comment:开始时间"`
	CompletedAt  *time.Time     `json:"completedAt" gorm:"comment:完成时间"`
	CreatedBy    uint           `json:"createdBy" gorm:"not null;comment:创建人ID"`

	// 关联关系
	Project   NesmaProject    `json:"project" gorm:"foreignKey:ProjectID"`
	Documents []NesmaDocument `json:"documents" gorm:"foreignKey:BatchID"`
}

// TableName 自定义表名
func (NesmaDocument) TableName() string {
	return "nesma_documents"
}

func (NesmaDocTemplate) TableName() string {
	return "nesma_doc_templates"
}

func (NesmaDocBatch) TableName() string {
	return "nesma_doc_batches"
}

// GetStatusText 获取状态文本
func (d *NesmaDocument) GetStatusText() string {
	switch d.Status {
	case "pending":
		return "待生成"
	case "generating":
		return "生成中"
	case "completed":
		return "已完成"
	case "failed":
		return "生成失败"
	default:
		return "未知状态"
	}
}

// GetTypeText 获取类型文本
func (d *NesmaDocument) GetTypeText() string {
	switch d.Type {
	case "word":
		return "Word文档"
	case "excel":
		return "Excel表格"
	case "pdf":
		return "PDF文档"
	default:
		return "未知类型"
	}
}

// GetFormatText 获取格式文本
func (d *NesmaDocument) GetFormatText() string {
	switch d.Format {
	case "requirement_spec":
		return "需求规格说明书"
	case "nesma_report":
		return "NESMA评估报告"
	case "business_summary":
		return "业务需求汇总表"
	default:
		return "未知格式"
	}
}

// IsCompleted 是否已完成
func (d *NesmaDocument) IsCompleted() bool {
	return d.Status == "completed"
}

// IsFailed 是否失败
func (d *NesmaDocument) IsFailed() bool {
	return d.Status == "failed"
}

// IsProcessing 是否正在处理
func (d *NesmaDocument) IsProcessing() bool {
	return d.Status == "generating"
}

// IsBatchDocument 是否为批量生成的文档
func (d *NesmaDocument) IsBatchDocument() bool {
	return d.BatchID != nil
}

// GetBatchID 获取批次ID
func (d *NesmaDocument) GetBatchID() uint {
	if d.BatchID != nil {
		return *d.BatchID
	}
	return 0
}

// GetCycleID 获取周期ID
func (d *NesmaDocument) GetCycleID() uint {
	if d.CycleID != nil {
		return *d.CycleID
	}
	return 0
}

// GetVersionID 获取版本ID
func (d *NesmaDocument) GetVersionID() uint {
	if d.VersionID != nil {
		return *d.VersionID
	}
	return 0
}

// HasCycleVersion 是否关联了周期和版本
func (d *NesmaDocument) HasCycleVersion() bool {
	return d.CycleID != nil && d.VersionID != nil
}

// GetDocumentScope 获取文档范围描述
func (d *NesmaDocument) GetDocumentScope() string {
	if d.HasCycleVersion() {
		return fmt.Sprintf("项目周期%d-版本%d", d.GetCycleID(), d.GetVersionID())
	}
	return "全项目范围"
}

// GetBatchStatusText 获取批次状态文本
func (b *NesmaDocBatch) GetBatchStatusText() string {
	switch b.Status {
	case "pending":
		return "待处理"
	case "processing":
		return "处理中"
	case "completed":
		return "已完成"
	case "failed":
		return "处理失败"
	default:
		return "未知状态"
	}
}

// IsCompleted 批次是否已完成
func (b *NesmaDocBatch) IsCompleted() bool {
	return b.Status == "completed"
}

// IsFailed 批次是否失败
func (b *NesmaDocBatch) IsFailed() bool {
	return b.Status == "failed"
}

// IsProcessing 批次是否正在处理
func (b *NesmaDocBatch) IsProcessing() bool {
	return b.Status == "processing"
}

// GetSuccessRate 获取成功率
func (b *NesmaDocBatch) GetSuccessRate() float64 {
	if b.TotalCount == 0 {
		return 0.0
	}
	return float64(b.SuccessCount) / float64(b.TotalCount) * 100
}

// UpdateProgress 更新批次进度
func (b *NesmaDocBatch) UpdateProgress() {
	processedCount := b.SuccessCount + b.FailedCount
	if b.TotalCount > 0 {
		b.Progress = int(float64(processedCount) / float64(b.TotalCount) * 100)
	}

	// 更新状态
	if processedCount == b.TotalCount {
		if b.FailedCount == 0 {
			b.Status = "completed"
		} else if b.SuccessCount == 0 {
			b.Status = "failed"
		} else {
			b.Status = "completed" // 部分成功也算完成
		}
	} else if processedCount > 0 {
		b.Status = "processing"
	}
}
