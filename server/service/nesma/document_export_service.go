package nesma

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// DocumentExportService 文档导出服务 - 简化版
type DocumentExportService struct {
	dataProvider      *ExportDataProvider
	generators        map[ExportType]Generator
	generatorsMutex   sync.RWMutex
	exportResults     map[string]*ExportResult // 内存中的导出结果缓存
	resultsMutex      sync.RWMutex
}

// NewDocumentExportService 创建文档导出服务
func NewDocumentExportService() *DocumentExportService {
	service := &DocumentExportService{
		dataProvider:  NewExportDataProvider(),
		generators:    make(map[ExportType]Generator),
		exportResults: make(map[string]*ExportResult),
	}

	// 注册三种固定生成器
	service.registerGenerators()

	return service
}

// registerGenerators 注册生成器
func (s *DocumentExportService) registerGenerators() {
	// 注册Word需求规格书生成器
	s.RegisterGenerator(NewWordRequirementSpecGenerator())
	
	// 注册Excel NESMA报告生成器
	s.RegisterGenerator(NewExcelNesmaReportGenerator())
	
	// 注册Excel业务汇总表生成器
	s.RegisterGenerator(NewExcelBusinessSummaryGenerator())

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("文档生成器注册完成", 
			zap.Int("生成器数量", len(s.generators)))
	}
}

// RegisterGenerator 注册生成器
func (s *DocumentExportService) RegisterGenerator(generator Generator) {
	s.generatorsMutex.Lock()
	defer s.generatorsMutex.Unlock()
	
	exportType := generator.GetSupportedType()
	s.generators[exportType] = generator
	
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("注册文档生成器", 
			zap.String("类型", string(exportType)))
	}
}

// Export 统一导出接口 - 核心方法
func (s *DocumentExportService) Export(req *ExportRequest) (*ExportResult, error) {
	global.GVA_LOG.Info("开始导出文档", 
		zap.Uint("projectId", req.ProjectID),
		zap.String("exportType", string(req.ExportType)),
		zap.Uint("userId", req.UserID))

	// 1. 验证请求
	if err := s.dataProvider.ValidateExportRequest(req); err != nil {
		global.GVA_LOG.Error("导出请求验证失败", zap.Error(err))
		return nil, err
	}

	// 2. 检查生成器是否存在
	s.generatorsMutex.RLock()
	generator, exists := s.generators[req.ExportType]
	s.generatorsMutex.RUnlock()
	
	if !exists {
		return nil, NewExportError(ErrInvalidExportType, 
			"不支持的导出类型: "+string(req.ExportType))
	}

	// 3. 创建导出结果记录
	result := &ExportResult{
		ID:         uint(time.Now().UnixNano()), // 简单的ID生成
		ProjectID:  req.ProjectID,
		ExportType: req.ExportType,
		Status:     "generating",
		Progress:   0,
		Message:    "正在准备数据...",
		CreatedAt:  time.Now(),
	}

	// 4. 缓存结果
	resultKey := s.getResultKey(req.ProjectID, req.ExportType)
	s.cacheResult(resultKey, result)

	// 5. 获取导出数据
	result.Message = "正在获取项目数据..."
	result.Progress = 20
	s.updateResult(resultKey, result)

	exportData, err := s.dataProvider.GetProjectExportData(req)
	if err != nil {
		result.Status = "failed"
		result.ErrorMsg = err.Error()
		result.Message = "获取数据失败: " + err.Error()
		s.updateResult(resultKey, result)
		return result, err
	}

	// 6. 验证数据
	result.Message = "正在验证数据..."
	result.Progress = 40
	s.updateResult(resultKey, result)

	if err := generator.Validate(exportData); err != nil {
		result.Status = "failed"
		result.ErrorMsg = err.Error()
		result.Message = "数据验证失败: " + err.Error()
		s.updateResult(resultKey, result)
		return result, err
	}

	// 7. 生成文档
	result.Message = "正在生成文档..."
	result.Progress = 60
	s.updateResult(resultKey, result)

	filePath, err := generator.Generate(exportData, req.Config)
	if err != nil {
		result.Status = "failed"
		result.ErrorMsg = err.Error()
		result.Message = "文档生成失败: " + err.Error()
		s.updateResult(resultKey, result)
		return result, err
	}

	// 8. 完成导出
	result.Message = "文档生成完成"
	result.Progress = 100
	result.Status = "completed"
	result.FilePath = filePath
	result.FileName = req.ExportType.GetFileName(exportData.Project.Name)
	result.ProcessedCount = len(exportData.Requirements)
	
	// 获取文件大小
	if fileInfo, err := os.Stat(filePath); err == nil {
		result.FileSize = fileInfo.Size()
	}
	
	completedAt := time.Now()
	result.CompletedAt = &completedAt
	
	// 生成下载URL（简化版）
	result.DownloadURL = fmt.Sprintf("/api/v1/nesma/document/download/%d/%s", 
		req.ProjectID, string(req.ExportType))

	s.updateResult(resultKey, result)

	global.GVA_LOG.Info("文档导出完成", 
		zap.String("filePath", filePath),
		zap.String("fileName", result.FileName),
		zap.Int64("fileSize", result.FileSize),
		zap.Int("processedCount", result.ProcessedCount))

	return result, nil
}

// ExportAsync 异步导出 - 用于大文档
func (s *DocumentExportService) ExportAsync(req *ExportRequest) (*ExportResult, error) {
	// 创建初始结果
	result := &ExportResult{
		ID:         uint(time.Now().UnixNano()),
		ProjectID:  req.ProjectID,
		ExportType: req.ExportType,
		Status:     "generating",
		Progress:   0,
		Message:    "导出任务已启动...",
		CreatedAt:  time.Now(),
	}

	resultKey := s.getResultKey(req.ProjectID, req.ExportType)
	s.cacheResult(resultKey, result)

	// 异步执行导出
	go func() {
		finalResult, err := s.Export(req)
		if err != nil {
			global.GVA_LOG.Error("异步导出失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("异步导出完成", 
				zap.String("fileName", finalResult.FileName))
		}
	}()

	return result, nil
}

// GetExportProgress 获取导出进度
func (s *DocumentExportService) GetExportProgress(projectID uint, exportType ExportType) (*ExportResult, error) {
	resultKey := s.getResultKey(projectID, exportType)
	
	s.resultsMutex.RLock()
	result, exists := s.exportResults[resultKey]
	s.resultsMutex.RUnlock()
	
	if !exists {
		return nil, NewExportError(ErrDataGeneration, "导出任务不存在")
	}

	// 返回副本，避免并发修改
	resultCopy := *result
	return &resultCopy, nil
}

// ListSupportedTypes 列出支持的导出类型
func (s *DocumentExportService) ListSupportedTypes() []ExportType {
	s.generatorsMutex.RLock()
	defer s.generatorsMutex.RUnlock()
	
	types := make([]ExportType, 0, len(s.generators))
	for exportType := range s.generators {
		types = append(types, exportType)
	}
	
	return types
}

// GetExportTypeInfo 获取导出类型信息
func (s *DocumentExportService) GetExportTypeInfo() map[ExportType]string {
	return map[ExportType]string{
		ExportWordRequirementSpec:  "Word需求规格说明书 - 包含完整的项目需求文档，层级结构清晰，包含AI分析结果和流程图",
		ExportExcelNesmaReport:     "Excel NESMA分析报告 - 功能点统计、UFP/AFP计算、复杂度分析等专业NESMA评估报告",
		ExportExcelBusinessSummary: "Excel业务需求汇总表 - 需求层级结构、工作量评估、进度追踪等业务管理表格",
	}
}

// DownloadFile 下载文件
func (s *DocumentExportService) DownloadFile(projectID uint, exportType ExportType) (string, error) {
	resultKey := s.getResultKey(projectID, exportType)
	
	s.resultsMutex.RLock()
	result, exists := s.exportResults[resultKey]
	s.resultsMutex.RUnlock()
	
	if !exists {
		return "", NewExportError(ErrDataGeneration, "导出文件不存在")
	}

	if result.Status != "completed" {
		return "", NewExportError(ErrDataGeneration, "文件尚未生成完成")
	}

	// 检查文件是否存在
	if _, err := os.Stat(result.FilePath); os.IsNotExist(err) {
		return "", NewExportError(ErrFileGeneration, "文件已被删除或移动")
	}

	return result.FilePath, nil
}

// CleanupOldExports 清理旧的导出文件
func (s *DocumentExportService) CleanupOldExports(maxAge time.Duration) error {
	s.resultsMutex.Lock()
	defer s.resultsMutex.Unlock()

	cutoffTime := time.Now().Add(-maxAge)
	deletedCount := 0

	for key, result := range s.exportResults {
		if result.CreatedAt.Before(cutoffTime) {
			// 删除文件
			if result.FilePath != "" {
				if err := os.Remove(result.FilePath); err != nil {
					global.GVA_LOG.Warn("删除导出文件失败", 
						zap.String("filePath", result.FilePath), 
						zap.Error(err))
				}
			}
			
			// 从缓存中删除
			delete(s.exportResults, key)
			deletedCount++
		}
	}

	global.GVA_LOG.Info("清理旧导出文件完成", 
		zap.Int("删除数量", deletedCount),
		zap.Duration("保留时间", maxAge))

	return nil
}

// 私有方法

func (s *DocumentExportService) getResultKey(projectID uint, exportType ExportType) string {
	return fmt.Sprintf("%d_%s", projectID, string(exportType))
}

func (s *DocumentExportService) cacheResult(key string, result *ExportResult) {
	s.resultsMutex.Lock()
	defer s.resultsMutex.Unlock()
	s.exportResults[key] = result
}

func (s *DocumentExportService) updateResult(key string, result *ExportResult) {
	s.resultsMutex.Lock()
	defer s.resultsMutex.Unlock()
	if existing, exists := s.exportResults[key]; exists {
		// 更新现有结果
		existing.Status = result.Status
		existing.Progress = result.Progress
		existing.Message = result.Message
		existing.ErrorMsg = result.ErrorMsg
		existing.FilePath = result.FilePath
		existing.FileName = result.FileName
		existing.FileSize = result.FileSize
		existing.ProcessedCount = result.ProcessedCount
		existing.CompletedAt = result.CompletedAt
		existing.DownloadURL = result.DownloadURL
	}
}

// ensureStorageDir 确保存储目录存在
func (s *DocumentExportService) ensureStorageDir() string {
	storePath := filepath.Join(global.GVA_CONFIG.Local.StorePath, "exports")
	if err := os.MkdirAll(storePath, 0755); err != nil {
		global.GVA_LOG.Error("创建导出目录失败", zap.Error(err))
		// 回退到临时目录
		storePath = filepath.Join(os.TempDir(), "nesma_exports")
		os.MkdirAll(storePath, 0755)
	}
	return storePath
}

// GetStoragePath 获取存储路径
func (s *DocumentExportService) GetStoragePath() string {
	return s.ensureStorageDir()
}

// GenerateFileName 生成文件名
func (s *DocumentExportService) GenerateFileName(projectName string, exportType ExportType) string {
	// 清理项目名称中的非法字符
	cleanName := s.sanitizeFileName(projectName)
	return exportType.GetFileName(cleanName)
}

// sanitizeFileName 清理文件名中的非法字符
func (s *DocumentExportService) sanitizeFileName(name string) string {
	// 替换常见的非法字符
	replacements := map[string]string{
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}
	
	result := name
	for old, new := range replacements {
		result = strings.Replace(result, old, new, -1)
	}
	
	return result
}

// GetExportHistory 获取导出历史（简化版，从内存缓存读取）
func (s *DocumentExportService) GetExportHistory(projectID uint, limit int) []*ExportResult {
	s.resultsMutex.RLock()
	defer s.resultsMutex.RUnlock()
	
	var results []*ExportResult
	for _, result := range s.exportResults {
		if result.ProjectID == projectID {
			resultCopy := *result
			results = append(results, &resultCopy)
		}
	}
	
	// 简单排序（按创建时间倒序）
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].CreatedAt.Before(results[j].CreatedAt) {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
	
	// 限制数量
	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	
	return results
}