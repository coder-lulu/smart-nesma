package nesma

import (
	"time"
	"gorm.io/datatypes"
)

// ExportType 导出类型枚举 - 只支持三种固定类型
type ExportType string

const (
	ExportWordRequirementSpec  ExportType = "word_requirement_spec"   // Word需求规格说明书
	ExportExcelNesmaReport     ExportType = "excel_nesma_report"      // Excel NESMA报告  
	ExportExcelBusinessSummary ExportType = "excel_business_summary"  // Excel业务需求汇总表
)

// IsValid 验证导出类型是否有效
func (et ExportType) IsValid() bool {
	switch et {
	case ExportWordRequirementSpec, ExportExcelNesmaReport, ExportExcelBusinessSummary:
		return true
	default:
		return false
	}
}

// GetFileName 获取文件名后缀
func (et ExportType) GetFileName(projectName string) string {
	timestamp := time.Now().Format("20060102150405")
	switch et {
	case ExportWordRequirementSpec:
		return projectName + "_需求规格说明书_" + timestamp + ".docx"
	case ExportExcelNesmaReport:
		return projectName + "_NESMA分析报告_" + timestamp + ".xlsx"
	case ExportExcelBusinessSummary:
		return projectName + "_业务需求汇总表_" + timestamp + ".xlsx"
	default:
		return projectName + "_文档_" + timestamp
	}
}

// ExportRequest 统一的导出请求
type ExportRequest struct {
	ProjectID  uint              `json:"projectId" binding:"required"`
	CycleID    *uint             `json:"cycleId" comment:"项目周期ID，为空则导出所有周期"`
	VersionID  *uint             `json:"versionId" comment:"需求版本ID，为空则导出最新版本"`
	ExportType ExportType        `json:"exportType" binding:"required"`
	Config     ExportConfig      `json:"config" comment:"导出配置"`
	UserID     uint              `json:"userId" comment:"请求用户ID"`
}

// ExportConfig 导出配置
type ExportConfig struct {
	// 通用配置
	IncludeAIAnalysis   bool `json:"includeAiAnalysis" default:"true" comment:"是否包含AI分析结果"`
	IncludeMermaidDiag  bool `json:"includeMermaidDiag" default:"true" comment:"是否包含Mermaid流程图"`
	IncludeStatistics   bool `json:"includeStatistics" default:"true" comment:"是否包含统计信息"`
	
	// Word文档特定配置
	WordConfig WordExportConfig `json:"wordConfig"`
	
	// Excel文档特定配置
	ExcelConfig ExcelExportConfig `json:"excelConfig"`
}

// WordExportConfig Word导出配置
type WordExportConfig struct {
	IncludeCoverPage    bool   `json:"includeCoverPage" default:"true" comment:"是否包含封面"`
	IncludeTableOfContents bool `json:"includeTableOfContents" default:"true" comment:"是否包含目录"`
	FontSize            int    `json:"fontSize" default:"12" comment:"字体大小"`
	LineSpacing         string `json:"lineSpacing" default:"1.5" comment:"行距"`
	DetailLevel         int    `json:"detailLevel" default:"4" comment:"详细级别1-4"`
}

// ExcelExportConfig Excel导出配置
type ExcelExportConfig struct {
	IncludeHeader       bool `json:"includeHeader" default:"true" comment:"是否包含表头"`
	IncludeStatSheet    bool `json:"includeStatSheet" default:"true" comment:"是否包含统计表"`
	FreezeHeader        bool `json:"freezeHeader" default:"true" comment:"是否冻结表头"`
	AutoColumnWidth     bool `json:"autoColumnWidth" default:"true" comment:"是否自动调整列宽"`
}

// ExportResult 导出结果
type ExportResult struct {
	ID            uint      `json:"id"`
	ProjectID     uint      `json:"projectId"`
	ExportType    ExportType `json:"exportType"`
	FileName      string    `json:"fileName"`
	FilePath      string    `json:"filePath"`
	FileSize      int64     `json:"fileSize"`
	Status        string    `json:"status"` // generating, completed, failed
	Progress      int       `json:"progress"` // 0-100
	Message       string    `json:"message"`
	ProcessedCount int      `json:"processedCount" comment:"处理的需求数量"`
	ErrorMsg      string    `json:"errorMsg,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	CompletedAt   *time.Time `json:"completedAt,omitempty"`
	DownloadURL   string    `json:"downloadUrl,omitempty"`
}

// ProjectExportData 项目导出数据 - 统一的数据结构
type ProjectExportData struct {
	Project      ProjectInfo        `json:"project"`
	Cycle        *CycleInfo         `json:"cycle,omitempty"`
	Version      *VersionInfo       `json:"version,omitempty"`
	Requirements []RequirementInfo  `json:"requirements"`
	Statistics   StatisticsInfo     `json:"statistics"`
	GeneratedAt  time.Time          `json:"generatedAt"`
}

// ProjectInfo 项目信息
type ProjectInfo struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Domain      string                 `json:"domain"`
	Status      string                 `json:"status"`
	Owner       string                 `json:"owner"`
	StartDate   *time.Time             `json:"startDate"`
	EndDate     *time.Time             `json:"endDate"`
	Settings    datatypes.JSON         `json:"settings"`
}

// CycleInfo 周期信息
type CycleInfo struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"startDate"`
	EndDate     *time.Time `json:"endDate"`
}

// VersionInfo 版本信息
type VersionInfo struct {
	ID          uint       `json:"id"`
	Version     string     `json:"version"`
	VersionType string     `json:"versionType"`
	Summary     string     `json:"summary"`
	CreatedBy   string     `json:"createdBy"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// RequirementInfo 需求信息 - 扁平化结构便于模板使用
type RequirementInfo struct {
	ID                uint       `json:"id"`
	Code              string     `json:"code"`
	Title             string     `json:"title"`
	Description       string     `json:"description"`
	Level             int        `json:"level"`
	ParentID          *uint      `json:"parentId"`
	OrderIndex        int        `json:"orderIndex"`
	
	// 层级信息 - 便于模板直接使用
	Level1Title       string     `json:"level1Title"`
	Level2Title       string     `json:"level2Title"`
	Level3Title       string     `json:"level3Title"`
	Level4Title       string     `json:"level4Title"`
	
	// NESMA相关字段
	FunctionType      string     `json:"functionType"`
	Complexity        string     `json:"complexity"`
	ComplexityLevel   string     `json:"complexityLevel"`
	UFP               float64    `json:"ufp"`
	AFP               float64    `json:"afp"`
	ReuseLevel        string     `json:"reuseLevel"`
	ModificationType  string     `json:"modificationType"`
	
	// AI分析相关
	AIDescription     string     `json:"aiDescription"`
	AIGeneratedTitle  string     `json:"aiGeneratedTitle"`
	AIComplexityScore *float64   `json:"aiComplexityScore"`
	AIConfidenceScore *float64   `json:"aiConfidenceScore"`
	AIAnalysisTime    *time.Time `json:"aiAnalysisTime"`
	MermaidDiagram    string     `json:"mermaidDiagram"`
	
	// 业务相关
	BusinessValue     string     `json:"businessValue"`
	AcceptanceCriteria string    `json:"acceptanceCriteria"`
	Priority          int        `json:"priority"`
	Status            string     `json:"status"`
	Category          string     `json:"category"`
	
	// 工时相关
	EstimateHours     *float64   `json:"estimateHours"`
	ActualHours       *float64   `json:"actualHours"`
	Notes             string     `json:"notes"`
}

// StatisticsInfo 统计信息
type StatisticsInfo struct {
	// 需求数量统计
	TotalRequirements int `json:"totalRequirements"`
	Level1Count       int `json:"level1Count"`
	Level2Count       int `json:"level2Count"`
	Level3Count       int `json:"level3Count"`
	Level4Count       int `json:"level4Count"`
	
	// NESMA功能类型统计
	EICount           int     `json:"eiCount"`
	EOCount           int     `json:"eoCount"`
	EQCount           int     `json:"eqCount"`
	ILFCount          int     `json:"ilfCount"`
	EIFCount          int     `json:"eifCount"`
	
	// UFP统计
	EIUFP             float64 `json:"eiUfp"`
	EOUFP             float64 `json:"eoUfp"`
	EQUFP             float64 `json:"eqUfp"`
	ILFUFP            float64 `json:"ilfUfp"`
	EIFUFP            float64 `json:"eifUfp"`
	TotalUFP          float64 `json:"totalUfp"`
	
	// AFP统计
	TotalAFP          float64 `json:"totalAfp"`
	AdjustmentFactor  float64 `json:"adjustmentFactor"`
	
	// 复杂度统计
	SimpleCount       int     `json:"simpleCount"`
	AverageCount      int     `json:"averageCount"`
	ComplexCount      int     `json:"complexCount"`
	
	// 工作量统计
	TotalEstimateHours float64 `json:"totalEstimateHours"`
	TotalActualHours   float64 `json:"totalActualHours"`
	
	// AI分析统计
	AIAnalyzedCount    int     `json:"aiAnalyzedCount"`
	AvgConfidenceScore float64 `json:"avgConfidenceScore"`
}

// Generator 导出生成器接口
type Generator interface {
	Generate(data *ProjectExportData, config ExportConfig) (string, error)
	Validate(data *ProjectExportData) error
	GetSupportedType() ExportType
}

// ErrorCode 错误代码
type ErrorCode string

const (
	ErrInvalidExportType    ErrorCode = "INVALID_EXPORT_TYPE"
	ErrProjectNotFound      ErrorCode = "PROJECT_NOT_FOUND"
	ErrCycleNotFound        ErrorCode = "CYCLE_NOT_FOUND"
	ErrVersionNotFound      ErrorCode = "VERSION_NOT_FOUND"
	ErrDataGeneration       ErrorCode = "DATA_GENERATION_ERROR"
	ErrFileGeneration       ErrorCode = "FILE_GENERATION_ERROR"
	ErrPermissionDenied     ErrorCode = "PERMISSION_DENIED"
	ErrInvalidConfig        ErrorCode = "INVALID_CONFIG"
)

// ExportError 导出错误
type ExportError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

func (e ExportError) Error() string {
	return string(e.Code) + ": " + e.Message
}

// NewExportError 创建导出错误
func NewExportError(code ErrorCode, message string, details ...string) *ExportError {
	err := &ExportError{
		Code:    code,
		Message: message,
	}
	if len(details) > 0 {
		err.Details = details[0]
	}
	return err
}