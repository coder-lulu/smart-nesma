package nesma

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"archive/zip"
	"io"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

type DocumentService struct{}

// =========================== 文档生成记录管理 ===========================

// CreateDocument 创建文档生成记录
func (s *DocumentService) CreateDocument(req *nesmaReq.CreateNesmaDocumentRequest, userID uint) (*nesma.NesmaDocument, error) {
	// 验证项目存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return nil, errors.New("项目不存在")
	}

	// 验证模板存在（支持内置模板）
	if req.TemplateType == "builtin_excel" || req.TemplateType == "builtin_word" {
		// 内置模板验证
		if req.TemplateType == "builtin_excel" && req.Type != "excel" {
			return nil, errors.New("内置Excel模板只支持excel类型")
		}
		if req.TemplateType == "builtin_word" && req.Type != "word" {
			return nil, errors.New("内置Word模板只支持word类型")
		}
	} else {
		// 数据库模板验证（暂时注释，因为当前使用内置模板）
		// var template nesma.NesmaDocTemplate
		// if err := global.GVA_DB.First(&template, req.TemplateID).Error; err != nil {
		// 	return nil, errors.New("模板不存在")
		// }
		//
		// // 验证模板类型与请求类型匹配
		// if template.Type != req.Type || template.Format != req.Format {
		// 	return nil, errors.New("模板类型或格式不匹配")
		// }
		return nil, errors.New("暂不支持自定义模板，请使用内置模板")
	}

	// 生成文档版本
	version := req.Version
	if version == "" {
		version = fmt.Sprintf("v%s", time.Now().Format("20060102150405"))
	}

	// 验证项目周期和版本（如果指定）
	if req.CycleID != nil {
		var cycle nesma.NesmaProjectCycle
		if err := global.GVA_DB.First(&cycle, *req.CycleID).Error; err != nil {
			return nil, errors.New("项目周期不存在")
		}
		if cycle.ProjectID != req.ProjectID {
			return nil, errors.New("项目周期不属于指定项目")
		}
	}

	if req.VersionID != nil {
		var version nesma.NesmaRequirementVersion
		if err := global.GVA_DB.First(&version, *req.VersionID).Error; err != nil {
			return nil, errors.New("需求版本不存在")
		}
		if req.CycleID != nil && version.CycleID != *req.CycleID {
			return nil, errors.New("需求版本不属于指定项目周期")
		}
	}

	// 创建文档记录
	doc := nesma.NesmaDocument{
		ProjectID:   req.ProjectID,
		CycleID:     req.CycleID,
		VersionID:   req.VersionID,
		TemplateID:  req.TemplateID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		Format:      req.Format,
		Status:      "pending",
		Progress:    0,
		Config:      req.Config,
		Version:     version,
		CreatedBy:   userID,
	}

	if err := global.GVA_DB.Create(&doc).Error; err != nil {
		return nil, err
	}

	// 预加载关联数据（暂时不加载Template，因为使用内置模板）
	if err := global.GVA_DB.Preload("Project").Preload("Cycle").Preload("RequirementVersion").First(&doc, doc.ID).Error; err != nil {
		return nil, err
	}

	return &doc, nil
}

// UpdateDocument 更新文档生成记录
func (s *DocumentService) UpdateDocument(req *nesmaReq.UpdateNesmaDocumentRequest) (*nesma.NesmaDocument, error) {
	var doc nesma.NesmaDocument
	if err := global.GVA_DB.First(&doc, req.ID).Error; err != nil {
		return nil, errors.New("文档不存在")
	}

	// 验证项目存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return nil, errors.New("项目不存在")
	}

	// 验证模板存在（支持内置模板）
	if req.TemplateID == "builtin_excel" || req.TemplateID == "builtin_word" {
		// 内置模板验证
		if req.TemplateID == "builtin_excel" && req.Type != "excel" {
			return nil, errors.New("内置Excel模板只支持excel类型")
		}
		if req.TemplateID == "builtin_word" && req.Type != "word" {
			return nil, errors.New("内置Word模板只支持word类型")
		}
	} else {
		// 数据库模板验证（暂时注释，因为当前使用内置模板）
		// var template nesma.NesmaDocTemplate
		// if err := global.GVA_DB.First(&template, req.TemplateID).Error; err != nil {
		// 	return nil, errors.New("模板不存在")
		// }
		return nil, errors.New("暂不支持自定义模板，请使用内置模板")
	}

	// 验证项目周期和版本（如果指定）
	if req.CycleID != nil {
		var cycle nesma.NesmaProjectCycle
		if err := global.GVA_DB.First(&cycle, *req.CycleID).Error; err != nil {
			return nil, errors.New("项目周期不存在")
		}
		if cycle.ProjectID != req.ProjectID {
			return nil, errors.New("项目周期不属于指定项目")
		}
	}

	if req.VersionID != nil {
		var version nesma.NesmaRequirementVersion
		if err := global.GVA_DB.First(&version, *req.VersionID).Error; err != nil {
			return nil, errors.New("需求版本不存在")
		}
		if req.CycleID != nil && version.CycleID != *req.CycleID {
			return nil, errors.New("需求版本不属于指定项目周期")
		}
	}

	// 更新文档记录
	updates := map[string]interface{}{
		"project_id":  req.ProjectID,
		"cycle_id":    req.CycleID,
		"version_id":  req.VersionID,
		"template_id": req.TemplateID,
		"name":        req.Name,
		"description": req.Description,
		"type":        req.Type,
		"format":      req.Format,
		"config":      req.Config,
		"version":     req.Version,
	}

	if err := global.GVA_DB.Model(&doc).Updates(updates).Error; err != nil {
		return nil, err
	}

	// 重新查询获取更新后的数据（暂时不加载Template，因为使用内置模板）
	if err := global.GVA_DB.Preload("Project").Preload("Cycle").Preload("RequirementVersion").First(&doc, req.ID).Error; err != nil {
		return nil, err
	}

	return &doc, nil
}

// DeleteDocument 删除文档生成记录
func (s *DocumentService) DeleteDocument(id uint) error {
	var doc nesma.NesmaDocument
	if err := global.GVA_DB.First(&doc, id).Error; err != nil {
		return errors.New("文档不存在")
	}

	// 删除文件
	if doc.FilePath != "" {
		if err := os.Remove(doc.FilePath); err != nil {
			global.GVA_LOG.Warn("删除文档文件失败", zap.String("filePath", doc.FilePath), zap.Error(err))
		}
	}

	return global.GVA_DB.Delete(&doc).Error
}

// GetDocument 获取文档详情
func (s *DocumentService) GetDocument(id uint) (*nesmaRes.NesmaDocumentResponse, error) {
	var doc nesma.NesmaDocument
	if err := global.GVA_DB.Preload("Project").Preload("Cycle").Preload("RequirementVersion").First(&doc, id).Error; err != nil {
		return nil, errors.New("文档不存在")
	}

	response := &nesmaRes.NesmaDocumentResponse{
		NesmaDocument: doc,
		StatusText:    doc.GetStatusText(),
		TypeText:      doc.GetTypeText(),
		FormatText:    doc.GetFormatText(),
	}

	return response, nil
}

// GetDocumentList 获取文档列表
func (s *DocumentService) GetDocumentList(req *nesmaReq.NesmaDocumentSearch) (*nesmaRes.NesmaDocumentListResponse, error) {
	var docs []nesma.NesmaDocument
	var total int64

	db := global.GVA_DB.Model(&nesma.NesmaDocument{})

	// 条件筛选
	if req.ProjectID != nil {
		db = db.Where("project_id = ?", *req.ProjectID)
	}
	if req.CycleID != nil {
		db = db.Where("cycle_id = ?", *req.CycleID)
	}
	if req.VersionID != nil {
		db = db.Where("version_id = ?", *req.VersionID)
	}
	if req.TemplateID != nil {
		db = db.Where("template_id = ?", *req.TemplateID)
	}
	if req.Type != "" {
		db = db.Where("type = ?", req.Type)
	}
	if req.Format != "" {
		db = db.Where("format = ?", req.Format)
	}
	if req.Status != "" {
		db = db.Where("status = ?", req.Status)
	}
	if req.CreatedBy != nil {
		db = db.Where("created_by = ?", *req.CreatedBy)
	}
	if req.Keyword != "" {
		db = db.Where("name LIKE ? OR description LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
 
	// 获取总数
	db.Count(&total)

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := db.Preload("Project").Preload("Cycle").Preload("RequirementVersion").
		Order("created_at DESC").
		Offset(offset).Limit(req.PageSize).Find(&docs).Error; err != nil {
		return nil, err
	}

	// 转换响应格式
	var list []nesmaRes.NesmaDocumentResponse
	for _, doc := range docs {
		list = append(list, nesmaRes.NesmaDocumentResponse{
			NesmaDocument: doc,
			StatusText:    doc.GetStatusText(),
			TypeText:      doc.GetTypeText(),
			FormatText:    doc.GetFormatText(),
		})
	}

	return &nesmaRes.NesmaDocumentListResponse{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// =========================== 文档生成功能 ===========================

// GenerateDocument 生成文档
func (s *DocumentService) GenerateDocument(req *nesmaReq.GenerateDocumentRequest) (*nesmaRes.DocumentGenerateResponse, error) {
	// 获取文档记录
	var doc nesma.NesmaDocument
	if err := global.GVA_DB.Preload("Project").Preload("Cycle").Preload("RequirementVersion").First(&doc, req.DocumentID).Error; err != nil {
		return nil, errors.New("文档不存在")
	}

	// 检查文档状态
	if doc.Status == "generating" {
		return nil, errors.New("文档正在生成中")
	}

	// 更新状态为生成中
	if err := global.GVA_DB.Model(&doc).Updates(map[string]interface{}{
		"status":   "generating",
		"progress": 0,
	}).Error; err != nil {
		return nil, err
	}

	// 如果是异步生成，启动goroutine
	if req.Async {
		go func() {
			s.generateDocumentAsync(doc, req.Config)
		}()
		return &nesmaRes.DocumentGenerateResponse{
			DocumentID: doc.ID,
			Status:     "generating",
			Progress:   0,
			Message:    "文档生成已启动",
		}, nil
	}

	// 同步生成
	return s.generateDocumentSync(doc, req.Config)
}

// generateDocumentSync 同步生成文档
func (s *DocumentService) generateDocumentSync(doc nesma.NesmaDocument, config []byte) (*nesmaRes.DocumentGenerateResponse, error) {
	var err error
	var filePath string

	// 根据文档类型生成
	switch doc.Type {
	case "word":
		filePath, err = s.generateWordDocument(doc, config)
	case "excel":
		filePath, err = s.generateExcelDocument(doc, config)
	case "pdf":
		filePath, err = s.generatePdfDocument(doc, config)
	default:
		err = errors.New("不支持的文档类型")
	}

	if err != nil {
		// 更新状态为失败
		global.GVA_DB.Model(&doc).Updates(map[string]interface{}{
			"status":    "failed",
			"progress":  0,
			"error_msg": err.Error(),
		})
		return &nesmaRes.DocumentGenerateResponse{
			DocumentID: doc.ID,
			Status:     "failed",
			Progress:   0,
			Message:    err.Error(),
		}, nil
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	// 更新状态为完成
	now := time.Now()
	updates := map[string]interface{}{
		"status":       "completed",
		"progress":     100,
		"file_path":    filePath,
		"file_size":    fileInfo.Size(),
		"generated_at": &now,
	}

	if err := global.GVA_DB.Model(&doc).Updates(updates).Error; err != nil {
		return nil, err
	}

	return &nesmaRes.DocumentGenerateResponse{
		DocumentID: doc.ID,
		Status:     "completed",
		Progress:   100,
		Message:    "文档生成完成",
		FilePath:   filePath,
	}, nil
}

// generateDocumentAsync 异步生成文档
func (s *DocumentService) generateDocumentAsync(doc nesma.NesmaDocument, config []byte) {
	defer func() {
		if r := recover(); r != nil {
			global.GVA_LOG.Error("文档生成异步任务panic", zap.Any("panic", r))
			global.GVA_DB.Model(&doc).Updates(map[string]interface{}{
				"status":    "failed",
				"progress":  0,
				"error_msg": fmt.Sprintf("生成失败: %v", r),
			})
		}
	}()

	// 生成文档
	_, err := s.generateDocumentSync(doc, config)
	if err != nil {
		global.GVA_LOG.Error("异步生成文档失败", zap.Error(err))
	}
}

// generateWordDocument 生成Word文档 - 使用开源方案
func (s *DocumentService) generateWordDocument(doc nesma.NesmaDocument, config []byte) (string, error) {
	global.GVA_LOG.Info("开始生成Word文档（开源方案）", zap.Uint("documentId", doc.ID))

	// 获取项目需求数据
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Preload("Project").Preload("Parent").Preload("Cycle").Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		return "", err
	}

	// 生成HTML内容
	htmlContent := s.generateWordHTMLContent(doc, requirements)

	// 保存文档
	fileName := fmt.Sprintf("%s_%s.docx", doc.Name, time.Now().Format("20060102150405"))
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	filePath := filepath.Join(global.GVA_CONFIG.Local.StorePath, "documents", fileName)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", err
	}

	// 使用开源方案生成Word文档
	if err := s.createWordDocumentFromHTML(htmlContent, filePath); err != nil {
		return "", err
	}

	global.GVA_LOG.Info("Word文档生成完成（开源方案）", zap.String("filePath", filePath))
	return filePath, nil
}

// generateExcelDocument 生成Excel文档
func (s *DocumentService) generateExcelDocument(doc nesma.NesmaDocument, config []byte) (string, error) {
	global.GVA_LOG.Info("开始生成Excel文档", 
		zap.Uint("documentId", doc.ID),
		zap.String("templateId", doc.TemplateID),
		zap.String("format", doc.Format),
		zap.String("type", doc.Type))

	// 创建Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			global.GVA_LOG.Error("关闭Excel文件失败", zap.Error(err))
		}
	}()

	// 检查是否为内置模板
	if doc.TemplateID == "builtin_excel" {
		global.GVA_LOG.Info("使用内置Excel模板生成")
		if err := s.generateBuiltinExcelTemplate(f, doc); err != nil {
			global.GVA_LOG.Error("内置Excel模板生成失败", zap.Error(err))
			return "", err
		}
	} else {
		global.GVA_LOG.Info("使用格式化Excel模板生成", zap.String("format", doc.Format))
		// 根据格式类型生成内容
		switch doc.Format {
		case "requirement_spec":
			global.GVA_LOG.Info("生成需求规格说明书Excel")
			if err := s.generateRequirementSpecExcel(f, doc); err != nil {
				return "", err
			}
		case "nesma_report":
			global.GVA_LOG.Info("生成NESMA报告Excel")
			if err := s.generateNesmaReportExcel(f, doc); err != nil {
				return "", err
			}
		case "business_summary":
			global.GVA_LOG.Info("生成业务需求汇总表Excel")
			if err := s.generateBusinessSummaryExcel(f, doc); err != nil {
				return "", err
			}
		default:
			return "", errors.New("不支持的Excel文档格式: " + doc.Format)
		}
	}

	// 保存文档
	fileName := fmt.Sprintf("%s_%s.xlsx", doc.Name, time.Now().Format("20060102150405"))
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	filePath := filepath.Join(global.GVA_CONFIG.Local.StorePath, "documents", fileName)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", err
	}

	if err := f.SaveAs(filePath); err != nil {
		return "", err
	}

	global.GVA_LOG.Info("Excel文档生成完成", zap.String("filePath", filePath))
	return filePath, nil
}

// generatePdfDocument 生成PDF文档
func (s *DocumentService) generatePdfDocument(doc nesma.NesmaDocument, config []byte) (string, error) {
	global.GVA_LOG.Info("开始生成PDF文档", zap.Uint("documentId", doc.ID))

	// 创建PDF文档
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(20, 20, 20)
	pdf.AddPage()

	// 根据格式类型生成内容
	switch doc.Format {
	case "requirement_spec":
		if err := s.generateRequirementSpecPdf(pdf, doc); err != nil {
			return "", err
		}
	case "nesma_report":
		if err := s.generateNesmaReportPdf(pdf, doc); err != nil {
			return "", err
		}
	case "business_summary":
		if err := s.generateBusinessSummaryPdf(pdf, doc); err != nil {
			return "", err
		}
	default:
		return "", errors.New("不支持的PDF文档格式")
	}

	// 保存文档
	fileName := fmt.Sprintf("%s_%s.pdf", doc.Name, time.Now().Format("20060102150405"))
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	filePath := filepath.Join(global.GVA_CONFIG.Local.StorePath, "documents", fileName)

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return "", err
	}

	if err := pdf.OutputFileAndClose(filePath); err != nil {
		return "", err
	}

	global.GVA_LOG.Info("PDF文档生成完成", zap.String("filePath", filePath))
	return filePath, nil
}

// =========================== 文档生成实现 ===========================

// generateRequirementSpecWord 生成需求规格说明书Word文档
func (s *DocumentService) generateRequirementSpecWord(doc nesma.NesmaDocument) error {
	return nil
}

// generateNesmaReportExcel 生成NESMA评估报告Excel文档  
func (s *DocumentService) generateNesmaReportExcel(f *excelize.File, doc nesma.NesmaDocument) error {
	// 获取项目需求 - 基于周期和版本筛选，预加载关联数据
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Preload("Project").Preload("Parent").Preload("Cycle").Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		return err
	}

	// 创建工作表
	sheetName := "NESMA评估报告"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// 设置项目信息
	f.SetCellValue(sheetName, "A1", "项目名称")
	f.SetCellValue(sheetName, "B1", doc.Project.Name)
	f.SetCellValue(sheetName, "A2", "生成时间")
	f.SetCellValue(sheetName, "B2", time.Now().Format("2006-01-02 15:04:05"))
	
	headerRow := 4
	if doc.HasCycleVersion() {
		f.SetCellValue(sheetName, "A3", "数据范围")
		f.SetCellValue(sheetName, "B3", doc.GetDocumentScope())
		headerRow = 5
	}

	// 设置增强的表头 - 包含AI分析结果
	headers := []string{
		"需求ID", "层级", "需求编号", "需求标题", "AI优化标题", 
		"需求描述", "AI优化描述", "功能类型", "复杂度", "AI复杂度评分",
		"UFP", "AFP", "AI推荐UFP", "AI推荐AFP", "AI分析状态", 
		"置信度", "知识库引用", "有Mermaid图",
	}
	
	for i, header := range headers {
		cell := fmt.Sprintf("%s%d", string(rune('A'+i)), headerRow)
		f.SetCellValue(sheetName, cell, header)
		// 设置表头样式
		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true, Size: 10},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#D3D3D3"}, Pattern: 1},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
		})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	// 统计AI分析结果
	var totalFP, totalRecommendedFP float64
	var analysisStats = map[string]int{
		"completed": 0,
		"pending": 0,
		"failed": 0,
	}
	var functionTypeStats = make(map[string]int)
	var aiOptimizedCount = 0

	// 填充详细数据
	for i, req := range requirements {
		row := i + headerRow + 1
		
		// 基础信息
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), req.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), req.GetLevelName())
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), req.Code)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), req.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), req.AIGeneratedTitle)
		
		// 描述信息  
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), req.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), req.AIDescription)
		
		// NESMA评估信息
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), req.FunctionType)
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), req.ComplexityLevel)
		if req.AIComplexityScore != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("%.1f", *req.AIComplexityScore))
		}
		
		// 功能点信息
		f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), req.UFP)
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), req.AFP)
		if req.RecommendedUFP != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("M%d", row), *req.RecommendedUFP)
			totalRecommendedFP += *req.RecommendedUFP
		}
		if req.RecommendedAFP != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("N%d", row), *req.RecommendedAFP)
		}
		
		// AI分析状态
		f.SetCellValue(sheetName, fmt.Sprintf("O%d", row), req.GetAIAnalysisStatusText())
		if req.AIConfidenceScore != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("P%d", row), fmt.Sprintf("%.2f", *req.AIConfidenceScore))
		}
		f.SetCellValue(sheetName, fmt.Sprintf("Q%d", row), req.KnowledgeReferences)
		f.SetCellValue(sheetName, fmt.Sprintf("R%d", row), func() string {
			if req.MermaidDiagram != "" { return "是" } else { return "否" }
		}())
		
		// 统计数据
		totalFP += req.AFP
		analysisStats[req.AIAnalysisStatus]++
		if req.FunctionType != "" {
			functionTypeStats[req.FunctionType]++
		}
		if req.AIOptimizationApplied {
			aiOptimizedCount++
		}
		
		// 设置AI状态的颜色
		var statusColor string
		switch req.AIAnalysisStatus {
		case "completed":
			statusColor = "#90EE90" // 浅绿色
		case "failed":
			statusColor = "#FFB6C1" // 浅红色
		default:
			statusColor = "#FFFFE0" // 浅黄色
		}
		if statusColor != "" {
			statusStyle, _ := f.NewStyle(&excelize.Style{
				Fill: excelize.Fill{Type: "pattern", Color: []string{statusColor}, Pattern: 1},
			})
			f.SetCellStyle(sheetName, fmt.Sprintf("O%d", row), fmt.Sprintf("O%d", row), statusStyle)
		}
	}

	// 添加统计汇总表
	summaryRow := len(requirements) + headerRow + 3
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "== 统计汇总 ==")
	summaryStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12},
	})
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", summaryRow), fmt.Sprintf("A%d", summaryRow), summaryStyle)
	
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "总需求数量")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), len(requirements))
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "总功能点数(AFP)")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), totalFP)
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "AI推荐功能点数")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), totalRecommendedFP)
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", summaryRow), "AI优化应用数量")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), aiOptimizedCount)
	
	// AI分析状态统计
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), "== AI分析状态统计 ==")
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", summaryRow), fmt.Sprintf("D%d", summaryRow), summaryStyle)
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), "已完成")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", summaryRow), analysisStats["completed"])
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), "待分析")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", summaryRow), analysisStats["pending"])
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", summaryRow), "分析失败")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", summaryRow), analysisStats["failed"])
	
	// 功能类型统计
	summaryRow++
	f.SetCellValue(sheetName, fmt.Sprintf("G%d", summaryRow), "== 功能类型统计 ==")
	f.SetCellStyle(sheetName, fmt.Sprintf("G%d", summaryRow), fmt.Sprintf("G%d", summaryRow), summaryStyle)
	summaryRow++
	for funcType, count := range functionTypeStats {
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", summaryRow), funcType)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", summaryRow), count)
		summaryRow++
	}

	// 设置列宽
	columnWidths := map[string]float64{
		"A": 8, "B": 12, "C": 10, "D": 25, "E": 25,
		"F": 30, "G": 30, "H": 10, "I": 12, "J": 12,
		"K": 8, "L": 8, "M": 12, "N": 12, "O": 12,
		"P": 8, "Q": 20, "R": 10,
	}
	for col, width := range columnWidths {
		f.SetColWidth(sheetName, col, col, width)
	}

	return nil
}

// generateBusinessSummaryPdf 生成业务需求汇总表PDF文档
func (s *DocumentService) generateBusinessSummaryPdf(pdf *gofpdf.Fpdf, doc nesma.NesmaDocument) error {
	// 设置中文字体支持
	pdf.SetFont("Arial", "", 12)

	// 添加标题
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("%s - Business Requirements Summary", doc.Project.Name))
	pdf.Ln(15)

	// 添加项目信息
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, fmt.Sprintf("Project: %s", doc.Project.Name))
	pdf.Ln(6)
	if doc.HasCycleVersion() {
		pdf.Cell(0, 8, fmt.Sprintf("Scope: %s", doc.GetDocumentScope()))
		pdf.Ln(6)
	}
	pdf.Cell(0, 8, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	// 获取项目统计 - 基于周期和版本筛选
	var totalCount int64
	query := global.GVA_DB.Model(&nesma.NesmaRequirement{}).Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	query.Count(&totalCount)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Summary Statistics")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Total Requirements: %d", totalCount))
	pdf.Ln(6)

	return nil
}

// 其他格式生成方法的占位符实现
func (s *DocumentService) generateNesmaReportWord(doc nesma.NesmaDocument) error {
	return nil
}

func (s *DocumentService) generateBusinessSummaryWord(doc nesma.NesmaDocument) error {
	return nil
}

func (s *DocumentService) generateRequirementSpecExcel(f *excelize.File, doc nesma.NesmaDocument) error {
	// 获取项目需求 - 基于周期和版本筛选
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		return err
	}

	// 创建工作表
	sheetName := "需求规格说明书"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// 设置表头
	headers := []string{"需求ID", "需求名称", "需求描述", "层级", "父级ID", "优先级", "状态", "创建时间"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		// 设置表头样式
		style, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Bold: true},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"#E6E6FA"}, Pattern: 1},
		})
		f.SetCellStyle(sheetName, cell, cell, style)
	}

	// 填充数据
	for i, req := range requirements {
		row := i + 2
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), req.ID)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), req.Title)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), req.Description)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), req.Level)
		if req.ParentID != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), *req.ParentID)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), req.Priority)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), req.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), req.CreatedAt.Format("2006-01-02 15:04:05"))
	}

	// 自动调整列宽
	for i := range headers {
		column := string(rune('A' + i))
		f.SetColWidth(sheetName, column, column, 15)
	}

	return nil
}

func (s *DocumentService) generateBusinessSummaryExcel(f *excelize.File, doc nesma.NesmaDocument) error {
	global.GVA_LOG.Info("开始生成业务需求汇总表Excel（调用内置模板）", 
		zap.Uint("documentId", doc.ID),
		zap.Uint("projectId", doc.ProjectID))

	// 直接调用我们优化过的内置Excel模板生成逻辑
	return s.generateBuiltinExcelTemplate(f, doc)
}

func (s *DocumentService) generateRequirementSpecPdf(pdf *gofpdf.Fpdf, doc nesma.NesmaDocument) error {
	// 获取项目需求 - 基于周期和版本筛选
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		return err
	}

	// 设置中文字体支持
	pdf.SetFont("Arial", "", 12)

	// 添加标题
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("%s - Requirements Specification", doc.Project.Name))
	pdf.Ln(15)

	// 添加项目信息
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Project Information")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Project: %s", doc.Project.Name))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Description: %s", doc.Project.Description))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	// 添加需求列表
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Requirements List")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	for _, req := range requirements {
		indent := float64(req.Level-1) * 10
		pdf.SetX(20 + indent)
		pdf.Cell(0, 6, fmt.Sprintf("%d. %s", req.ID, req.Title))
		pdf.Ln(6)
		if req.Description != "" {
			pdf.SetX(25 + indent)
			pdf.Cell(0, 6, fmt.Sprintf("Description: %s", req.Description))
			pdf.Ln(6)
		}
		pdf.Ln(2)
	}

	return nil
}

func (s *DocumentService) generateNesmaReportPdf(pdf *gofpdf.Fpdf, doc nesma.NesmaDocument) error {
	// 获取项目需求数据 - 基于周期和版本筛选
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		return err
	}

	// 统计功能点分布
	typeStats := make(map[string]int)
	complexityStats := make(map[string]int)
	var totalFP float64

	for _, req := range requirements {
		if req.FunctionType != "" {
			typeStats[req.FunctionType]++
			totalFP += req.AFP
		}
		if req.Complexity != "" {
			complexityStats[req.Complexity]++
		}
	}

	// 设置中文字体支持
	pdf.SetFont("Arial", "", 12)

	// 添加标题
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, fmt.Sprintf("%s - NESMA Function Point Analysis Report", doc.Project.Name))
	pdf.Ln(15)

	// 添加项目信息
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Project Summary")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 6, fmt.Sprintf("Project: %s", doc.Project.Name))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Total Function Points: %.1f", totalFP))
	pdf.Ln(6)
	pdf.Cell(0, 6, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(10)

	// 添加功能点统计
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Function Point Distribution")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	for funcType, count := range typeStats {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %d", funcType, count))
		pdf.Ln(6)
	}

	// 添加复杂度统计
	pdf.Ln(5)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(0, 8, "Complexity Distribution")
	pdf.Ln(8)

	pdf.SetFont("Arial", "", 10)
	for complexity, count := range complexityStats {
		pdf.Cell(0, 6, fmt.Sprintf("%s: %d", complexity, count))
		pdf.Ln(6)
	}

	return nil
}

// =========================== 新增缺失的方法 ===========================

// BatchGenerateDocument 批量生成文档
func (s *DocumentService) BatchGenerateDocument(req *nesmaReq.BatchGenerateDocumentRequest, userID uint) (*nesmaRes.BatchGenerateResponse, error) {
	// 验证项目存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return nil, errors.New("项目不存在")
	}

	// 创建批次记录
	batch := nesma.NesmaDocBatch{
		ProjectID:   req.ProjectID,
		Name:        req.Name,
		Description: req.Description,
		TotalCount:  len(req.Documents),
		Status:      "pending",
		Progress:    0,
		Config:      req.Config,
		CreatedBy:   userID,
	}

	if err := global.GVA_DB.Create(&batch).Error; err != nil {
		return nil, err
	}

	// 创建文档记录
	var documents []nesma.NesmaDocument
	for _, docConfig := range req.Documents {
		doc := nesma.NesmaDocument{
			ProjectID:   req.ProjectID,
			TemplateID:  docConfig.TemplateID,
			Name:        docConfig.Name,
			Description: docConfig.Description,
			Type:        docConfig.Type,
			Format:      docConfig.Format,
			Status:      "pending",
			Progress:    0,
			Config:      docConfig.Config,
			Version:     docConfig.Version,
			CreatedBy:   userID,
		}
		documents = append(documents, doc)
	}

	if err := global.GVA_DB.Create(&documents).Error; err != nil {
		return nil, err
	}

	// 异步批量生成
	go func() {
		s.processBatchGeneration(batch.ID, documents)
	}()

	return &nesmaRes.BatchGenerateResponse{
		BatchID:      batch.ID,
		Status:       "processing",
		Progress:     0,
		TotalCount:   batch.TotalCount,
		SuccessCount: 0,
		FailedCount:  0,
		Message:      "批量生成已启动",
	}, nil
}

// processBatchGeneration 处理批量生成
func (s *DocumentService) processBatchGeneration(batchID uint, documents []nesma.NesmaDocument) {
	defer func() {
		if r := recover(); r != nil {
			global.GVA_LOG.Error("批量生成文档panic", zap.Any("panic", r))
		}
	}()

	var successCount, failedCount int
	for i, doc := range documents {
		// 生成文档
		_, err := s.generateDocumentSync(doc, doc.Config)
		if err != nil {
			failedCount++
		} else {
			successCount++
		}

		// 更新批次进度
		progress := (i + 1) * 100 / len(documents)
		global.GVA_DB.Model(&nesma.NesmaDocBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
			"progress":      progress,
			"success_count": successCount,
			"failed_count":  failedCount,
		})
	}

	// 更新批次状态
	status := "completed"
	if failedCount > 0 {
		status = "partial"
	}

	now := time.Now()
	global.GVA_DB.Model(&nesma.NesmaDocBatch{}).Where("id = ?", batchID).Updates(map[string]interface{}{
		"status":        status,
		"progress":      100,
		"success_count": successCount,
		"failed_count":  failedCount,
		"completed_at":  &now,
	})
}

// PreviewDocument 预览文档
func (s *DocumentService) PreviewDocument(req *nesmaReq.PreviewDocumentRequest) (*nesmaRes.DocumentPreviewResponse, error) {
	// 验证项目存在
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, req.ProjectID).Error; err != nil {
		return nil, errors.New("项目不存在")
	}

	// 验证内置模板ID
	if req.TemplateID != "builtin_excel" && req.TemplateID != "builtin_word" {
		return nil, errors.New("暂不支持自定义模板预览，请使用内置模板")
	}

	// 生成预览内容
	templateName := ""
	if req.TemplateID == "builtin_excel" {
		templateName = "业务需求信息汇总表（内置Excel模板）"
	} else {
		templateName = "NESMA功能点分析送审文档（内置Word模板）"
	}

	content := fmt.Sprintf("文档预览 - %s\n项目: %s\n模板: %s\n类型: %s\n格式: %s",
		templateName, project.Name, templateName, req.Type, req.Format)

	return &nesmaRes.DocumentPreviewResponse{
		PreviewURL: "",
		Content:    content,
		Format:     req.Format,
	}, nil
}

// DownloadDocument 下载文档
func (s *DocumentService) DownloadDocument(req *nesmaReq.DownloadDocumentRequest, c *gin.Context) error {
	var doc nesma.NesmaDocument
	if err := global.GVA_DB.First(&doc, req.DocumentID).Error; err != nil {
		return errors.New("文档不存在")
	}

	if doc.FilePath == "" || !doc.IsCompleted() {
		return errors.New("文档尚未生成完成")
	}

	// 检查文件是否存在
	if _, err := os.Stat(doc.FilePath); os.IsNotExist(err) {
		return errors.New("文档文件不存在")
	}

	// 设置响应头
	fileName := filepath.Base(doc.FilePath)
	c.Header("Content-Description", "File Transfer")
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Header("Content-Type", "application/octet-stream")

	// 发送文件
	c.File(doc.FilePath)
	return nil
}

// GetDocumentProgress 获取文档生成进度
func (s *DocumentService) GetDocumentProgress(documentID uint) (*nesmaRes.DocumentProgressResponse, error) {
	var doc nesma.NesmaDocument
	if err := global.GVA_DB.First(&doc, documentID).Error; err != nil {
		return nil, errors.New("文档不存在")
	}

	response := &nesmaRes.DocumentProgressResponse{
		DocumentID: doc.ID,
		Status:     doc.Status,
		Progress:   doc.Progress,
		Message:    doc.GetStatusText(),
		StartTime:  doc.CreatedAt.Format("2006-01-02 15:04:05"),
	}

	if doc.GeneratedAt != nil {
		response.EndTime = doc.GeneratedAt.Format("2006-01-02 15:04:05")
	}

	return response, nil
}

// GetDocumentStats 获取文档统计
func (s *DocumentService) GetDocumentStats(projectID *uint) (*nesmaRes.DocumentStatsResponse, error) {
	db := global.GVA_DB.Model(&nesma.NesmaDocument{})
	if projectID != nil {
		db = db.Where("project_id = ?", *projectID)
	}

	var totalCount, completedCount, pendingCount, failedCount int64

	// 总数
	db.Count(&totalCount)

	// 各状态统计
	db.Where("status = ?", "completed").Count(&completedCount)
	db.Where("status = ?", "pending").Count(&pendingCount)
	db.Where("status = ?", "failed").Count(&failedCount)

	// 类型统计
	var typeStats []nesmaRes.DocumentTypeStats
	var typeResults []struct {
		Type  string
		Count int64
	}
	db.Select("type, count(*) as count").Group("type").Scan(&typeResults)
	for _, result := range typeResults {
		typeStats = append(typeStats, nesmaRes.DocumentTypeStats{
			Type:     result.Type,
			TypeText: s.getTypeText(result.Type),
			Count:    result.Count,
		})
	}

	// 格式统计
	var formatStats []nesmaRes.DocumentFormatStats
	var formatResults []struct {
		Format string
		Count  int64
	}
	db.Select("format, count(*) as count").Group("format").Scan(&formatResults)
	for _, result := range formatResults {
		formatStats = append(formatStats, nesmaRes.DocumentFormatStats{
			Format:     result.Format,
			FormatText: s.getFormatText(result.Format),
			Count:      result.Count,
		})
	}

	return &nesmaRes.DocumentStatsResponse{
		TotalCount:     totalCount,
		CompletedCount: completedCount,
		PendingCount:   pendingCount,
		FailedCount:    failedCount,
		TypeStats:      typeStats,
		FormatStats:    formatStats,
	}, nil
}

// BatchDeleteDocuments 批量删除文档
func (s *DocumentService) BatchDeleteDocuments(req *nesmaReq.DocumentIdsRequest) error {
	// 检查文档是否存在
	var docs []nesma.NesmaDocument
	if err := global.GVA_DB.Where("id IN ?", req.IDs).Find(&docs).Error; err != nil {
		return err
	}

	if len(docs) != len(req.IDs) {
		return errors.New("部分文档不存在")
	}

	// 删除文件
	for _, doc := range docs {
		if doc.FilePath != "" {
			if err := os.Remove(doc.FilePath); err != nil {
				global.GVA_LOG.Warn("删除文档文件失败", zap.String("filePath", doc.FilePath), zap.Error(err))
			}
		}
	}

	// 批量删除记录
	return global.GVA_DB.Where("id IN ?", req.IDs).Delete(&nesma.NesmaDocument{}).Error
}

// 辅助方法
func (s *DocumentService) getTypeText(docType string) string {
	switch docType {
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

func (s *DocumentService) getFormatText(format string) string {
	switch format {
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

// ========================== 内置固定模板生成方法 ==========================

// generateBuiltinExcelTemplate 生成内置Excel模板文档（业务需求信息汇总表）
func (s *DocumentService) generateBuiltinExcelTemplate(f *excelize.File, doc nesma.NesmaDocument) error {
	global.GVA_LOG.Info("开始生成内置Excel模板", 
		zap.Uint("documentId", doc.ID),
		zap.Uint("projectId", doc.ProjectID),
		zap.Any("cycleId", doc.CycleID),
		zap.Any("versionId", doc.VersionID),
		zap.Bool("hasCycleVersion", doc.HasCycleVersion()))
	
	// 获取项目需求 - 基于周期和版本筛选，预加载关联数据
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Preload("Project").Preload("Parent").Preload("Cycle").Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		global.GVA_LOG.Info("使用周期版本筛选", 
			zap.Uint("cycleId", doc.GetCycleID()),
			zap.Uint("versionId", doc.GetVersionID()))
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	} else {
		global.GVA_LOG.Info("使用项目级别筛选，不限制周期版本")
	}
	
	// 执行查询并记录结果
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		global.GVA_LOG.Error("查询需求数据失败", zap.Error(err))
		return err
	}
	
	global.GVA_LOG.Info("数据库查询完成", 
		zap.Int("查询到的需求数量", len(requirements)),
		zap.String("SQL查询条件", fmt.Sprintf("project_id = %d", doc.ProjectID)))
	
	// 记录每个需求的详细信息
	for i, req := range requirements {
		global.GVA_LOG.Debug("需求详情", 
			zap.Int("索引", i),
			zap.Uint("需求ID", req.ID),
			zap.String("需求标题", req.Title),
			zap.Int("层级", req.Level),
			zap.String("需求编号", req.Code),
			zap.Uint("项目ID", req.ProjectID),
			zap.Any("周期ID", req.CycleID),
			zap.Any("版本ID", req.VersionID))
	}

	// 创建工作表 - 模拟模板结构
	sheetName := "功能需求列表"
	f.NewSheet(sheetName)
	f.DeleteSheet("Sheet1")

	// 设置列宽 - 与模板保持一致
	f.SetColWidth(sheetName, "A", "A", 12)  // 需求序号
	f.SetColWidth(sheetName, "B", "B", 25)  // 项目
	f.SetColWidth(sheetName, "C", "C", 15)  // 子系统
	f.SetColWidth(sheetName, "D", "D", 20)  // 一级模块
	f.SetColWidth(sheetName, "E", "E", 25)  // 二级模块
	f.SetColWidth(sheetName, "F", "F", 20)  // 三级模块
	f.SetColWidth(sheetName, "G", "G", 35)  // 功能点计数项名称
	f.SetColWidth(sheetName, "H", "H", 15)  // 新增、优化或删除

	// 创建表头样式 - 严格按照XML模板(s51样式) #C4D79B
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 9, Family: "微软雅黑", Color: "#000000"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"C4D79B"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	// 创建特殊表头样式 - 黄色背景(s53样式) #FFFF00
	specialHeaderStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 9, Family: "微软雅黑", Color: "#000000"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFF00"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	// 创建数据样式 - 按照XML模板(s54样式)
	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11, Family: "宋体", Color: "#000000"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
	})

	// 特殊样式：合并单元格的样式 - 按照XML模板(s59样式) - 将在实际合并时定义

	// 功能点计数项名称样式 - 按照XML模板(s58样式)
	_, _ = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11, Family: "Calibri", Color: "#000000"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
	})

	// 修改类型样式 - 按照XML模板(s55样式)
	_, _ = f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 11, Family: "宋体", Color: "#000000"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
	})

	// 设置表头 - 严格按照XML模板样式
	headers := []string{
		"需求序号", "项目", "子系统", "一级模块", "二级模块", 
		"三级模块", "功能点计数项名称", "新增、优化或删除\n（可选项：新增、优化、删除）",
	}
	
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		
		// 应用不同的表头样式
		if i == 7 { // 第8列（H列）使用黄色背景
			f.SetCellStyle(sheetName, cell, cell, specialHeaderStyle)
		} else if i == 5 || i == 6 { // 第6、7列使用绿色背景但不同配置
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		} else { // 其他列使用标准绿色背景
			f.SetCellStyle(sheetName, cell, cell, headerStyle)
		}
	}

	// 处理所有有效需求数据，不仅仅是4级
	validRequirements := make([]nesma.NesmaRequirement, 0)
	for i, req := range requirements {
		// 只要有需求标题就算有效需求
		if req.Title != "" {
			validRequirements = append(validRequirements, req)
			global.GVA_LOG.Debug("发现有效需求", 
				zap.Int("原始索引", i),
				zap.Uint("需求ID", req.ID),
				zap.String("需求标题", req.Title),
				zap.Int("层级", req.Level))
		} else {
			global.GVA_LOG.Debug("跳过无效需求", 
				zap.Int("原始索引", i),
				zap.Uint("需求ID", req.ID),
				zap.String("需求标题为空", req.Title))
		}
	}

	// 记录调试信息
	global.GVA_LOG.Info("Excel数据处理统计", 
		zap.Int("原始需求数量", len(requirements)),
		zap.Int("有效需求数量", len(validRequirements)),
		zap.String("项目名称", doc.Project.Name))

	if len(validRequirements) == 0 {
		global.GVA_LOG.Error("没有找到有效的需求数据")
		return errors.New("没有找到有效的需求数据")
	}

	// 填充数据并实现真正的单元格合并
	row := 2
	seqNo := 1
	
	// 获取项目信息
	projectName := doc.Project.Name
	subSystem := "智慧园区"  // 可以从项目配置中获取
	
	// 预填充项目名称和子系统（第一行）
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), projectName)
	f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), subSystem)
	
	// 直接填充所有有效需求数据
	for i, req := range validRequirements {
		global.GVA_LOG.Info("开始填充需求数据", 
			zap.Int("循环索引", i),
			zap.Int("序号", seqNo),
			zap.Int("行号", row),
			zap.Uint("需求ID", req.ID),
			zap.String("需求标题", req.Title),
			zap.Int("需求层级", req.Level))
		
		// 需求序号
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%03d", seqNo))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)
		
		// 项目名称和子系统 (每行都填充，后面再统一合并)
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), projectName)
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), subSystem)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), dataStyle)
		
		// 记录处理进度
		global.GVA_LOG.Debug("处理需求数据", 
			zap.Int("序号", seqNo),
			zap.Int("行号", row),
			zap.String("需求标题", req.Title),
			zap.Int("需求层级", req.Level))
		
		// 构建完整的层级路径
		var level1, level2, level3 string
		
		// 获取完整的层级路径
		if req.Level >= 1 {
			level1 = s.getRequirementAtLevel(req, 1)
		}
		if req.Level >= 2 {
			level2 = s.getRequirementAtLevel(req, 2)
		}
		if req.Level >= 3 {
			level3 = s.getRequirementAtLevel(req, 3)
		}
		
		// 填充各级模块
		if level1 != "" {
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), level1)
		}
		if level2 != "" {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), level2)
		}
		if level3 != "" {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), level3)
		}
		
		// G列：功能点计数项名称 (始终是当前需求的标题)
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), req.GetDisplayTitle())
		
		// 设置D-G列样式
		for col := 'D'; col <= 'G'; col++ {
			f.SetCellStyle(sheetName, fmt.Sprintf("%s%d", string(col), row), fmt.Sprintf("%s%d", string(col), row), dataStyle)
		}
		
		// 修改类型
		modificationType := "新增"
		if req.ModificationType != "" {
			modificationType = req.ModificationType
		}
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), modificationType)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), dataStyle)
		
		seqNo++
		row++
		
		global.GVA_LOG.Info("需求数据填充完成", 
			zap.Int("已处理序号", seqNo-1),
			zap.Int("下一行号", row),
			zap.String("已填充需求", req.Title))
	}

	// 记录最终统计信息
	global.GVA_LOG.Info("Excel数据填充完成", 
		zap.Int("实际插入行数", seqNo-1),
		zap.Int("最终行号", row-1),
		zap.Int("有效需求数量", len(validRequirements)))

	// 实现项目名称和子系统的全表合并
	if len(validRequirements) > 0 {
		totalRows := len(validRequirements) + 1  // +1 for header
		
		// 创建合并单元格样式
		mergedStyle, _ := f.NewStyle(&excelize.Style{
			Font: &excelize.Font{Size: 11, Family: "宋体", Color: "#000000"},
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		})
		
		global.GVA_LOG.Info("实现单元格合并", 
			zap.Int("总行数", totalRows),
			zap.String("合并范围B", fmt.Sprintf("B2:B%d", totalRows)),
			zap.String("合并范围C", fmt.Sprintf("C2:C%d", totalRows)))
		
		// 合并项目名称（B列）
		f.MergeCell(sheetName, "B2", fmt.Sprintf("B%d", totalRows))
		f.SetCellStyle(sheetName, "B2", fmt.Sprintf("B%d", totalRows), mergedStyle)
		
		// 合并子系统（C列）
		f.MergeCell(sheetName, "C2", fmt.Sprintf("C%d", totalRows))
		f.SetCellStyle(sheetName, "C2", fmt.Sprintf("C%d", totalRows), mergedStyle)
	}

	return nil
}

// getRequirementAtLevel 获取指定层级的需求标题
func (s *DocumentService) getRequirementAtLevel(req nesma.NesmaRequirement, targetLevel int) string {
	if req.Level == targetLevel {
		return req.GetDisplayTitle()
	}
	
	if req.Level > targetLevel {
		// 往上查找父级需求
		current := &req
		for current != nil && current.Level > targetLevel {
			if current.Parent != nil {
				current = current.Parent
			} else {
				break
			}
		}
		if current != nil && current.Level == targetLevel {
			return current.GetDisplayTitle()
		}
	}
	
	return ""
}

// mergeConsecutiveCells 对连续相同内容的单元格进行合并
func (s *DocumentService) mergeConsecutiveCells(f *excelize.File, sheetName string, levelMap map[string][]int, style int) {
	global.GVA_LOG.Info("开始处理层级合并", zap.Int("层级数量", len(levelMap)))
	
	for key, rows := range levelMap {
		if len(rows) <= 1 {
			continue // 只有一行不需要合并
		}
		
		// 解析列名和内容
		colName := strings.Split(key, ":")[0]
		content := strings.Split(key, ":")[1]
		
		global.GVA_LOG.Info("处理层级合并", 
			zap.String("列名", colName),
			zap.String("内容", content),
			zap.Ints("行号列表", rows))
		
		// 对连续相同内容的单元格进行分组合并
		groupStart := rows[0]
		for i := 1; i < len(rows); i++ {
			if rows[i] != rows[i-1]+1 {
				// 不连续，合并上一组
				if rows[i-1] > groupStart {
					startCell := fmt.Sprintf("%s%d", colName, groupStart)
					endCell := fmt.Sprintf("%s%d", colName, rows[i-1])
					f.MergeCell(sheetName, startCell, endCell)
					f.SetCellStyle(sheetName, startCell, endCell, style)
					global.GVA_LOG.Info("合并单元格", 
						zap.String("开始单元格", startCell),
						zap.String("结束单元格", endCell))
				}
				groupStart = rows[i]
			}
		}
		
		// 处理最后一组
		if len(rows) > 1 && rows[len(rows)-1] > groupStart {
			startCell := fmt.Sprintf("%s%d", colName, groupStart)
			endCell := fmt.Sprintf("%s%d", colName, rows[len(rows)-1])
			f.MergeCell(sheetName, startCell, endCell)
			f.SetCellStyle(sheetName, startCell, endCell, style)
			global.GVA_LOG.Info("合并最后组单元格", 
				zap.String("开始单元格", startCell),
				zap.String("结束单元格", endCell))
		}
	}
}

// generateBuiltinWordTemplate 生成内置Word模板文档（送审文档）
func (s *DocumentService) generateBuiltinWordTemplate(doc nesma.NesmaDocument) error {
	// 获取项目需求数据 - 基于周期和版本筛选，预加载关联数据
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Preload("Project").Preload("Parent").Preload("Cycle").Where("project_id = ?", doc.ProjectID)
	
	// 如果指定了周期和版本，则精确筛选
	if doc.HasCycleVersion() {
		query = query.Where("cycle_id = ? AND version_id = ?", doc.GetCycleID(), doc.GetVersionID())
	}
	
	if err := query.Order("level ASC, order_index ASC").Find(&requirements).Error; err != nil {
		return err
	}

	// 使用新的开源方案生成HTML内容
	htmlContent := s.generateWordHTMLContent(doc, requirements)
	
	// 生成文档路径
	fileName := fmt.Sprintf("%s_%s.docx", doc.Name, time.Now().Format("20060102150405"))
	fileName = strings.ReplaceAll(fileName, " ", "_")
	fileName = strings.ReplaceAll(fileName, "/", "_")
	fileName = strings.ReplaceAll(fileName, "\\", "_")
	filePath := filepath.Join(global.GVA_CONFIG.Local.StorePath, "documents", fileName)
	
	// 创建Word文档
	return s.createWordDocumentFromHTML(htmlContent, filePath)
}

// addDocumentTitle 添加文档标题页
func (s *DocumentService) addDocumentTitle(doc nesma.NesmaDocument) {
	// 简化实现
}

// addTableOfContents 添加目录
func (s *DocumentService) addTableOfContents() {
	// 简化实现
}

// addProjectBasicInfo 添加项目基本信息章节
func (s *DocumentService) addProjectBasicInfo(doc nesma.NesmaDocument) {
	// 简化实现
}

// addFunctionPointSummary 添加功能点统计汇总章节
func (s *DocumentService) addFunctionPointSummary(requirements []nesma.NesmaRequirement) {
	// 简化实现
}

// addDetailedRequirementList 添加详细功能点清单章节
func (s *DocumentService) addDetailedRequirementList(requirements []nesma.NesmaRequirement) {
	// 简化实现
}

// addNESMAAnalysisReport 添加NESMA分析报告章节
func (s *DocumentService) addNESMAAnalysisReport(requirements []nesma.NesmaRequirement) {
	// 简化实现
}

// addAIAnalysisResults 添加AI分析结果章节
func (s *DocumentService) addAIAnalysisResults(requirements []nesma.NesmaRequirement) {
	// 简化实现
}

// addConclusionAndRecommendations 添加结论与建议章节
func (s *DocumentService) addConclusionAndRecommendations(requirements []nesma.NesmaRequirement) {
	// 简化实现
}

// ========================== 开源Word文档生成方案 ==========================

// generateWordHTMLContent 生成Word文档的HTML内容
func (s *DocumentService) generateWordHTMLContent(doc nesma.NesmaDocument, requirements []nesma.NesmaRequirement) string {
	global.GVA_LOG.Info("开始生成Word HTML内容", 
		zap.String("文档名称", doc.Name),
		zap.Int("需求数量", len(requirements)),
		zap.Uint("项目ID", doc.ProjectID))
	
	var htmlBuilder strings.Builder
	
	// HTML头部
	htmlBuilder.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>` + doc.Name + `</title>
    <style>
        body { font-family: "宋体", SimSun; font-size: 12pt; line-height: 1.5; margin: 2cm; }
        h1 { font-size: 18pt; font-weight: bold; text-align: center; margin-bottom: 20pt; }
        h2 { font-size: 14pt; font-weight: bold; margin-top: 20pt; margin-bottom: 10pt; }
        h3 { font-size: 12pt; font-weight: bold; margin-top: 15pt; margin-bottom: 8pt; }
        table { border-collapse: collapse; width: 100%; margin-bottom: 20pt; }
        th, td { border: 1px solid #000; padding: 8pt; text-align: left; vertical-align: top; }
        th { background-color: #E6E6FA; font-weight: bold; text-align: center; }
        .header { text-align: center; font-size: 16pt; font-weight: bold; margin-bottom: 30pt; }
        .footer { text-align: center; font-size: 10pt; margin-top: 30pt; color: #666; }
        .section { margin-bottom: 30pt; }
        .toc { margin-bottom: 30pt; }
        .toc ul { list-style-type: none; padding-left: 20pt; }
        .toc li { margin-bottom: 5pt; }
        .page-break { page-break-before: always; }
    </style>
</head>
<body>`)

	// 1. 文档标题页
	htmlBuilder.WriteString(fmt.Sprintf(`
    <div class="header">
        <h1>%s</h1>
        <h2>NESMA功能点分析送审文档</h2>
        <p>项目名称：%s</p>
        <p>生成时间：%s</p>
    </div>`, doc.Name, doc.Project.Name, time.Now().Format("2006-01-02 15:04:05")))

	// 2. 目录
	htmlBuilder.WriteString(`
    <div class="page-break"></div>
    <div class="toc">
        <h2>目录</h2>
        <ul>
            <li>1. 项目基本信息</li>
            <li>2. 功能点统计分析</li>
            <li>3. 详细功能点清单</li>
            <li>4. NESMA计算结果</li>
            <li>5. AI分析结果</li>
            <li>6. 结论与建议</li>
        </ul>
    </div>`)

	// 3. 项目基本信息
	htmlBuilder.WriteString(`
    <div class="page-break"></div>
    <div class="section">
        <h2>1. 项目基本信息</h2>
        <table>
            <tr><th>项目属性</th><th>详细信息</th></tr>
            <tr><td>项目名称</td><td>` + doc.Project.Name + `</td></tr>
            <tr><td>项目描述</td><td>` + doc.Project.Description + `</td></tr>
            <tr><td>业务领域</td><td>` + doc.Project.Domain + `</td></tr>
            <tr><td>项目状态</td><td>` + doc.Project.Status + `</td></tr>
            <tr><td>创建时间</td><td>` + doc.Project.CreatedAt.Format("2006-01-02") + `</td></tr>
        </table>
    </div>`)

	// 4. 功能点统计分析
	htmlBuilder.WriteString(`
    <div class="section">
        <h2>2. 功能点统计分析</h2>
        <h3>2.1 需求层级分布</h3>
        <table>
            <tr><th>层级</th><th>数量</th><th>占比</th></tr>`)
	
	// 统计各层级需求数量
	levelStats := make(map[int]int)
	for _, req := range requirements {
		levelStats[req.Level]++
	}
	total := len(requirements)
	
	global.GVA_LOG.Info("层级统计完成", 
		zap.Int("总需求数", total),
		zap.Any("层级统计", levelStats))
	
	for level := 1; level <= 4; level++ {
		count := levelStats[level]
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}
		htmlBuilder.WriteString(fmt.Sprintf(`
            <tr><td>%d级需求</td><td>%d</td><td>%.1f%%</td></tr>`, level, count, percentage))
	}
	
	htmlBuilder.WriteString(`
        </table>
        <h3>2.2 功能类型分布</h3>
        <table>
            <tr><th>功能类型</th><th>数量</th><th>占比</th></tr>`)
	
	// 统计功能类型
	functionTypeStats := make(map[string]int)
	for _, req := range requirements {
		if req.FunctionType != "" {
			functionTypeStats[req.FunctionType]++
		}
	}
	
	global.GVA_LOG.Info("功能类型统计完成", 
		zap.Any("功能类型统计", functionTypeStats))
	
	for funcType, count := range functionTypeStats {
		percentage := 0.0
		if total > 0 {
			percentage = float64(count) / float64(total) * 100
		}
		htmlBuilder.WriteString(fmt.Sprintf(`
            <tr><td>%s</td><td>%d</td><td>%.1f%%</td></tr>`, funcType, count, percentage))
	}
	
	htmlBuilder.WriteString(`
        </table>
    </div>`)

	// 5. 详细功能点清单
	htmlBuilder.WriteString(`
    <div class="section">
        <h2>3. 详细功能点清单</h2>
        <table>
            <tr>
                <th>序号</th>
                <th>功能点名称</th>
                <th>功能类型</th>
                <th>复杂度</th>
                <th>AFP值</th>
                <th>AI分析状态</th>
            </tr>`)
	
	// 只显示4级需求（具体功能点）
	seqNo := 1
	level4Count := 0
	for _, req := range requirements {
		if req.Level == 4 {
			level4Count++
			global.GVA_LOG.Debug("添加功能点到HTML", 
				zap.Int("序号", seqNo),
				zap.String("功能点名称", req.GetDisplayTitle()),
				zap.String("功能类型", req.FunctionType))
			
			htmlBuilder.WriteString(fmt.Sprintf(`
            <tr>
                <td>%d</td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%.1f</td>
                <td>%s</td>
            </tr>`, seqNo, req.GetDisplayTitle(), req.FunctionType, req.ComplexityLevel, req.AFP, req.GetAIAnalysisStatusText()))
			seqNo++
		}
	}
	
	global.GVA_LOG.Info("功能点清单生成完成", 
		zap.Int("四级需求数量", level4Count),
		zap.Int("实际序号", seqNo-1))
	
	htmlBuilder.WriteString(`
        </table>
    </div>`)

	// 6. NESMA计算结果
	htmlBuilder.WriteString(`
    <div class="section">
        <h2>4. NESMA计算结果</h2>
        <table>
            <tr><th>功能类型</th><th>数量</th><th>功能点数</th><th>占比</th></tr>`)
	
	// 计算总功能点数
	totalFP := 0.0
	functionTypePoints := make(map[string]float64)
	for _, req := range requirements {
		if req.FunctionType != "" {
			functionTypePoints[req.FunctionType] += req.AFP
			totalFP += req.AFP
		}
	}
	
	global.GVA_LOG.Info("功能点计算完成", 
		zap.Float64("总功能点", totalFP),
		zap.Any("功能类型点数", functionTypePoints))
	
	for funcType, points := range functionTypePoints {
		count := functionTypeStats[funcType]
		percentage := 0.0
		if totalFP > 0 {
			percentage = points / totalFP * 100
		}
		htmlBuilder.WriteString(fmt.Sprintf(`
            <tr><td>%s</td><td>%d</td><td>%.1f</td><td>%.1f%%</td></tr>`, funcType, count, points, percentage))
	}
	
	htmlBuilder.WriteString(fmt.Sprintf(`
            <tr><td><strong>总计</strong></td><td>%d</td><td><strong>%.1f</strong></td><td><strong>100.0%%</strong></td></tr>
        </table>
    </div>`, len(requirements), totalFP))

	// 7. AI分析结果
	htmlBuilder.WriteString(`
    <div class="section">
        <h2>5. AI分析结果</h2>
        <p>本项目采用AI技术对需求进行智能分析和优化，主要包括以下方面：</p>
        <ul>
            <li>需求描述优化和标准化</li>
            <li>功能类型自动分类</li>
            <li>复杂度智能评估</li>
            <li>知识库匹配和参考</li>
        </ul>
    </div>`)

	// 8. 结论与建议
	htmlBuilder.WriteString(`
    <div class="section">
        <h2>6. 结论与建议</h2>
        <p>根据NESMA功能点分析结果，本项目共计` + fmt.Sprintf("%.1f", totalFP) + `个功能点，建议在项目实施过程中：</p>
        <ul>
            <li>重点关注高复杂度功能点的实现</li>
            <li>充分利用AI分析结果优化需求理解</li>
            <li>定期进行功能点评估和调整</li>
            <li>确保与NESMA标准的一致性</li>
        </ul>
    </div>`)

	// HTML尾部
	htmlBuilder.WriteString(`
    <div class="footer">
        <p>文档生成时间：` + time.Now().Format("2006-01-02 15:04:05") + `</p>
        <p>Smart-NESMA系统生成</p>
    </div>
</body>
</html>`)

	return htmlBuilder.String()
}

// createWordDocumentFromHTML 从HTML内容创建Word文档
func (s *DocumentService) createWordDocumentFromHTML(htmlContent, filePath string) error {
	global.GVA_LOG.Info("开始创建Word文档", 
		zap.String("文件路径", filePath),
		zap.Int("HTML内容长度", len(htmlContent)))
	
	// 创建一个简单的DOCX文档结构
	// 这里实现一个基本的DOCX生成器
	
	// 创建ZIP文件（DOCX是ZIP格式）
	zipFile, err := os.Create(filePath)
	if err != nil {
		global.GVA_LOG.Error("创建ZIP文件失败", zap.Error(err))
		return err
	}
	defer zipFile.Close()
	
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	
	// 1. 添加[Content_Types].xml
	contentTypesXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
    <Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
    <Default Extension="xml" ContentType="application/xml"/>
    <Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>
</Types>`
	
	if err := s.addFileToZip(zipWriter, "[Content_Types].xml", contentTypesXML); err != nil {
		return err
	}
	
	// 2. 添加_rels/.rels
	relsXML := `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">
    <Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/>
</Relationships>`
	
	if err := s.addFileToZip(zipWriter, "_rels/.rels", relsXML); err != nil {
		return err
	}
	
	// 3. 添加word/document.xml (主要内容)
	// 将HTML转换为简单的WordML
	global.GVA_LOG.Info("开始转换HTML为WordML")
	wordMLContent := s.convertHTMLToWordML(htmlContent)
	global.GVA_LOG.Info("WordML转换完成", 
		zap.Int("WordML内容长度", len(wordMLContent)))
	
	if err := s.addFileToZip(zipWriter, "word/document.xml", wordMLContent); err != nil {
		global.GVA_LOG.Error("添加WordML内容失败", zap.Error(err))
		return err
	}
	
	global.GVA_LOG.Info("Word文档创建完成", zap.String("文件路径", filePath))
	return nil
}

// addFileToZip 向ZIP文件添加文件
func (s *DocumentService) addFileToZip(zipWriter *zip.Writer, filename, content string) error {
	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	
	_, err = io.WriteString(writer, content)
	return err
}

// convertHTMLToWordML 将HTML转换为基本的WordML格式
func (s *DocumentService) convertHTMLToWordML(htmlContent string) string {
	global.GVA_LOG.Info("开始转换HTML为WordML", 
		zap.Int("HTML内容长度", len(htmlContent)))
	
	// 简化的HTML到WordML转换
	// 这里实现基本的转换逻辑
	
	var wordMLBuilder strings.Builder
	
	// WordML文档头部
	wordMLBuilder.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
    <w:body>`)
	
	// 解析HTML内容并转换为更好的WordML
	// 处理标题、表格、段落等
	lines := strings.Split(htmlContent, "\n")
	inTable := false
	processedLines := 0
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// 处理标题
		if strings.Contains(line, "<h1>") || strings.Contains(line, "<h2>") || strings.Contains(line, "<h3>") {
			headerText := s.cleanHTMLTags(line)
			if headerText != "" {
				// 根据标题级别设置不同的样式
				headerLevel := "Heading1"
				fontSize := "24"
				if strings.Contains(line, "<h2>") {
					headerLevel = "Heading2"
					fontSize = "20"
				} else if strings.Contains(line, "<h3>") {
					headerLevel = "Heading3"
					fontSize = "16"
				}
				
				wordMLBuilder.WriteString(fmt.Sprintf(`
        <w:p>
            <w:pPr>
                <w:pStyle w:val="%s"/>
                <w:spacing w:before="240" w:after="120"/>
            </w:pPr>
            <w:r>
                <w:rPr>
                    <w:b/>
                    <w:sz w:val="%s"/>
                    <w:color w:val="000000"/>
                </w:rPr>
                <w:t>%s</w:t>
            </w:r>
        </w:p>`, headerLevel, fontSize, headerText))
				processedLines++
				global.GVA_LOG.Debug("处理标题", 
					zap.String("标题内容", headerText),
					zap.String("标题级别", headerLevel))
			}
			continue
		}
		
		// 处理表格开始
		if strings.Contains(line, "<table>") {
			inTable = true
			wordMLBuilder.WriteString(`
        <w:tbl>
            <w:tblPr>
                <w:tblStyle w:val="TableGrid"/>
                <w:tblW w:w="5000" w:type="pct"/>
                <w:tblBorders>
                    <w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                    <w:left w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                    <w:bottom w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                    <w:right w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                    <w:insideH w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                    <w:insideV w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                </w:tblBorders>
                <w:tblLook w:val="04A0"/>
            </w:tblPr>`)
			global.GVA_LOG.Debug("开始处理表格")
			continue
		}
		
		// 处理表格结束
		if strings.Contains(line, "</table>") {
			inTable = false
			wordMLBuilder.WriteString(`
        </w:tbl>`)
			continue
		}
		
		// 处理表格行
		if inTable && strings.Contains(line, "<tr>") {
			wordMLBuilder.WriteString(`
            <w:tr>`)
			continue
		}
		
		if inTable && strings.Contains(line, "</tr>") {
			wordMLBuilder.WriteString(`
            </w:tr>`)
			continue
		}
		
		// 处理表格单元格
		if inTable && (strings.Contains(line, "<td>") || strings.Contains(line, "<th>")) {
			cellText := s.cleanHTMLTags(line)
			if cellText != "" {
				// 区分表头和普通单元格
				isHeader := strings.Contains(line, "<th>")
				headerStyle := ""
				if isHeader {
					headerStyle = `
                        <w:rPr>
                            <w:b/>
                            <w:color w:val="000000"/>
                        </w:rPr>`
				}
				
				// 设置单元格样式
				bgColor := "FFFFFF"
				alignment := "left"
				if isHeader {
					bgColor = "E6E6FA"
					alignment = "center"
				}
				
				wordMLBuilder.WriteString(fmt.Sprintf(`
                <w:tc>
                    <w:tcPr>
                        <w:tcW w:w="2000" w:type="dxa"/>
                        <w:tcBorders>
                            <w:top w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                            <w:left w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                            <w:bottom w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                            <w:right w:val="single" w:sz="4" w:space="0" w:color="000000"/>
                        </w:tcBorders>
                        <w:shd w:val="clear" w:color="auto" w:fill="%s"/>
                    </w:tcPr>
                    <w:p>
                        <w:pPr>
                            <w:jc w:val="%s"/>
                        </w:pPr>
                        <w:r>%s
                            <w:t>%s</w:t>
                        </w:r>
                    </w:p>
                </w:tc>`, bgColor, alignment, headerStyle, cellText))
				processedLines++
				global.GVA_LOG.Debug("处理表格单元格", 
					zap.String("单元格内容", cellText),
					zap.Bool("是否表头", isHeader))
			}
			continue
		}
		
		// 处理普通段落和其他内容
		if !inTable {
			// 过滤掉HTML结构标签和CSS样式
			if strings.Contains(line, "<style>") || strings.Contains(line, "</style>") || 
					strings.Contains(line, "<head>") || strings.Contains(line, "</head>") ||
					strings.Contains(line, "<html>") || strings.Contains(line, "</html>") ||
					strings.Contains(line, "<body>") || strings.Contains(line, "</body>") ||
					strings.Contains(line, "<!DOCTYPE") ||
					strings.Contains(line, "<title>") || strings.Contains(line, "</title>") ||
					strings.Contains(line, "<meta") ||
					strings.Contains(line, "font-family:") || strings.Contains(line, "font-size:") ||
					strings.Contains(line, "margin:") || strings.Contains(line, "padding:") ||
					strings.Contains(line, "border:") || strings.Contains(line, "background-color:") ||
					strings.Contains(line, "<div") || strings.Contains(line, "</div>") ||
					strings.Contains(line, "<p>") || strings.Contains(line, "</p>") ||
					strings.Contains(line, "}") {
				continue // 跳过CSS样式和HTML结构标签
			}
			
			cleanText := s.cleanHTMLTags(line)
			if cleanText != "" && len(strings.TrimSpace(cleanText)) > 0 {
				// 判断是否是列表项
				if strings.Contains(line, "<li>") {
					wordMLBuilder.WriteString(fmt.Sprintf(`
        <w:p>
            <w:pPr>
                <w:pStyle w:val="ListParagraph"/>
                <w:numPr>
                    <w:ilvl w:val="0"/>
                    <w:numId w:val="1"/>
                </w:numPr>
                <w:ind w:left="720"/>
            </w:pPr>
            <w:r>
                <w:t>• %s</w:t>
            </w:r>
        </w:p>`, cleanText))
				} else {
					// 普通段落
					wordMLBuilder.WriteString(fmt.Sprintf(`
        <w:p>
            <w:pPr>
                <w:spacing w:after="120"/>
            </w:pPr>
            <w:r>
                <w:rPr>
                    <w:sz w:val="22"/>
                    <w:color w:val="000000"/>
                </w:rPr>
                <w:t>%s</w:t>
            </w:r>
        </w:p>`, cleanText))
				}
				processedLines++
				global.GVA_LOG.Debug("处理段落内容", 
					zap.String("段落内容", cleanText))
			}
		}
	}
	
	// WordML文档尾部
	wordMLBuilder.WriteString(`
    </w:body>
</w:document>`)
	
	result := wordMLBuilder.String()
	global.GVA_LOG.Info("HTML到WordML转换完成", 
		zap.Int("处理的行数", processedLines),
		zap.Int("WordML内容长度", len(result)))
	
	return result
}

// cleanHTMLTags 清理HTML标签
func (s *DocumentService) cleanHTMLTags(text string) string {
	// 简单的HTML标签清理
	// 移除所有HTML标签，保留文本内容
	
	// 处理常见的HTML实体
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&apos;", "'")
	
	// 简单的标签移除
	for {
		start := strings.Index(text, "<")
		if start == -1 {
			break
		}
		end := strings.Index(text[start:], ">")
		if end == -1 {
			break
		}
		text = text[:start] + text[start+end+1:]
	}
	
	return strings.TrimSpace(text)
}
