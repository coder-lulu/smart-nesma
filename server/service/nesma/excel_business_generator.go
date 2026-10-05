package nesma

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// ExcelBusinessSummaryGenerator Excel业务需求汇总表生成器
type ExcelBusinessSummaryGenerator struct {
	storePath string
}

// NewExcelBusinessSummaryGenerator 创建Excel业务汇总表生成器
func NewExcelBusinessSummaryGenerator() *ExcelBusinessSummaryGenerator {
	return &ExcelBusinessSummaryGenerator{
		storePath: filepath.Join(global.GVA_CONFIG.Local.StorePath, "exports"),
	}
}

// GetSupportedType 获取支持的导出类型
func (g *ExcelBusinessSummaryGenerator) GetSupportedType() ExportType {
	return ExportExcelBusinessSummary
}

// Validate 验证数据
func (g *ExcelBusinessSummaryGenerator) Validate(data *ProjectExportData) error {
	if data == nil {
		return NewExportError(ErrDataGeneration, "导出数据为空")
	}
	
	if data.Project.Name == "" {
		return NewExportError(ErrDataGeneration, "项目名称不能为空")
	}
	
	if len(data.Requirements) == 0 {
		return NewExportError(ErrDataGeneration, "项目没有需求数据")
	}
	
	return nil
}

// Generate 生成Excel文档
func (g *ExcelBusinessSummaryGenerator) Generate(data *ProjectExportData, config ExportConfig) (string, error) {
	global.GVA_LOG.Info("开始生成Excel业务需求汇总表", 
		zap.String("项目", data.Project.Name),
		zap.Int("需求数量", len(data.Requirements)))

	// 创建新的Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			global.GVA_LOG.Error("关闭Excel文件失败", zap.Error(err))
		}
	}()

	// 生成各个工作表
	if err := g.generateHierarchySheet(f, data, config); err != nil {
		return "", err
	}

	if err := g.generateWorkloadSheet(f, data, config); err != nil {
		return "", err
	}

	// 删除默认工作表
	f.DeleteSheet("Sheet1")

	// 生成文件路径
	fileName := fmt.Sprintf("%s_业务需求汇总表_%s.xlsx", 
		g.sanitizeFileName(data.Project.Name),
		time.Now().Format("20060102150405"))
	
	filePath := filepath.Join(g.storePath, fileName)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return "", NewExportError(ErrFileGeneration, "保存Excel文件失败", err.Error())
	}

	global.GVA_LOG.Info("Excel业务需求汇总表生成完成", 
		zap.String("文件路径", filePath))

	return filePath, nil
}

// generateHierarchySheet 生成需求层级结构表
func (g *ExcelBusinessSummaryGenerator) generateHierarchySheet(f *excelize.File, data *ProjectExportData, config ExportConfig) error {
	sheetName := "需求层级结构"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// 设置列宽
	columnWidths := []struct{ col, width string }{
		{"A", "6"},   // 层级
		{"B", "15"},  // 需求编号
		{"C", "30"},  // 需求名称
		{"D", "40"},  // 需求描述
		{"E", "40"},  // AI优化描述
		{"F", "25"},  // 业务价值
		{"G", "25"},  // 验收标准
		{"H", "8"},   // 优先级
		{"I", "10"},  // 状态
		{"J", "12"},  // 负责人
	}

	for _, cw := range columnWidths {
		f.SetColWidth(sheetName, cw.col, cw.col, parseFloat(cw.width))
	}

	// 创建样式
	headerStyle := g.createHeaderStyle(f)
	level1Style := g.createLevel1Style(f)
	level2Style := g.createLevel2Style(f)
	level3Style := g.createLevel3Style(f)
	level4Style := g.createLevel4Style(f)
	dataStyle := g.createDataStyle(f)

	// 表头
	headers := []string{
		"层级", "需求编号", "需求名称", "需求描述", "AI优化描述",
		"业务价值", "验收标准", "优先级", "状态", "负责人",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 冻结表头
	if config.ExcelConfig.FreezeHeader {
		f.SetPanes(sheetName, &excelize.Panes{
			Freeze: true,
			YSplit: 1,
		})
	}

	// 按层级和顺序排序需求
	sortedReqs := g.sortRequirementsByHierarchy(data.Requirements)

	// 数据行
	row := 2
	for _, req := range sortedReqs {
		// 选择样式
		var rowStyle int
		switch req.Level {
		case 1:
			rowStyle = level1Style
		case 2:
			rowStyle = level2Style
		case 3:
			rowStyle = level3Style
		case 4:
			rowStyle = level4Style
		default:
			rowStyle = dataStyle
		}

		// 层级
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("%d级", req.Level))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)

		// 需求编号
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), req.Code)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)

		// 需求名称（带缩进显示层级）
		indent := strings.Repeat("  ", req.Level-1) // 根据层级添加缩进
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), indent+req.Title)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), rowStyle)

		// 需求描述
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), req.Description)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), dataStyle)

		// AI优化描述
		if config.IncludeAIAnalysis {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), req.AIDescription)
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), dataStyle)

		// 业务价值
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), req.BusinessValue)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), dataStyle)

		// 验收标准
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), req.AcceptanceCriteria)
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), dataStyle)

		// 优先级
		priorityText := g.getPriorityText(req.Priority)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), priorityText)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), dataStyle)

		// 状态
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), req.Status)
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), dataStyle)

		// 负责人（简化处理）
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), "待分配")
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), dataStyle)

		row++
	}

	return nil
}

// generateWorkloadSheet 生成工作量评估表
func (g *ExcelBusinessSummaryGenerator) generateWorkloadSheet(f *excelize.File, data *ProjectExportData, config ExportConfig) error {
	sheetName := "工作量评估"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// 设置列宽
	columnWidths := []struct{ col, width string }{
		{"A", "15"},  // 需求编号
		{"B", "30"},  // 需求名称
		{"C", "12"},  // 功能类型
		{"D", "10"},  // 复杂度
		{"E", "10"},  // 预估工时
		{"F", "10"},  // 实际工时
		{"G", "10"},  // 工时偏差
		{"H", "10"},  // 完成度
		{"I", "20"},  // 备注
	}

	for _, cw := range columnWidths {
		f.SetColWidth(sheetName, cw.col, cw.col, parseFloat(cw.width))
	}

	// 创建样式
	headerStyle := g.createHeaderStyle(f)
	dataStyle := g.createDataStyle(f)
	numberStyle := g.createNumberStyle(f)
	percentStyle := g.createPercentStyle(f)

	// 表头
	headers := []string{
		"需求编号", "需求名称", "功能类型", "复杂度",
		"预估工时", "实际工时", "工时偏差", "完成度", "备注",
	}

	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 冻结表头
	if config.ExcelConfig.FreezeHeader {
		f.SetPanes(sheetName, &excelize.Panes{
			Freeze: true,
			YSplit: 1,
		})
	}

	// 数据行 - 只显示4级需求（功能点）
	row := 2
	totalEstimate := 0.0
	totalActual := 0.0

	for _, req := range data.Requirements {
		if req.Level != 4 {
			continue
		}

		// 需求编号
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), req.Code)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)

		// 需求名称
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), req.Title)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)

		// 功能类型
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), req.FunctionType)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), dataStyle)

		// 复杂度
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), req.Complexity)
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), dataStyle)

		// 预估工时
		if req.EstimateHours != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), *req.EstimateHours)
			totalEstimate += *req.EstimateHours
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), 0)
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), numberStyle)

		// 实际工时
		if req.ActualHours != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), *req.ActualHours)
			totalActual += *req.ActualHours
		} else {
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), 0)
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), numberStyle)

		// 工时偏差（公式计算）
		f.SetCellFormula(sheetName, fmt.Sprintf("G%d", row), 
			fmt.Sprintf("IF(E%d>0,(F%d-E%d)/E%d,\"\")", row, row, row, row))
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), percentStyle)

		// 完成度（简化处理，基于状态）
		completionPercent := g.getCompletionPercent(req.Status)
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), completionPercent)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), percentStyle)

		// 备注
		notes := req.Notes
		if config.IncludeAIAnalysis && req.AIConfidenceScore != nil {
			if notes != "" {
				notes += "; "
			}
			notes += fmt.Sprintf("AI置信度: %.1f%%", *req.AIConfidenceScore*100)
		}
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), notes)
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), dataStyle)

		row++
	}

	// 汇总行
	if row > 2 {
		summaryRow := row + 1
		
		// 标签
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", summaryRow), "合计")
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", summaryRow), fmt.Sprintf("B%d", summaryRow), headerStyle)

		// 预估工时汇总
		f.SetCellFormula(sheetName, fmt.Sprintf("E%d", summaryRow), 
			fmt.Sprintf("SUM(E2:E%d)", row-1))
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", summaryRow), fmt.Sprintf("E%d", summaryRow), headerStyle)

		// 实际工时汇总
		f.SetCellFormula(sheetName, fmt.Sprintf("F%d", summaryRow), 
			fmt.Sprintf("SUM(F2:F%d)", row-1))
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", summaryRow), fmt.Sprintf("F%d", summaryRow), headerStyle)

		// 总体偏差
		f.SetCellFormula(sheetName, fmt.Sprintf("G%d", summaryRow), 
			fmt.Sprintf("IF(E%d>0,(F%d-E%d)/E%d,\"\")", summaryRow, summaryRow, summaryRow, summaryRow))
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", summaryRow), fmt.Sprintf("G%d", summaryRow), headerStyle)
	}

	return nil
}

// sortRequirementsByHierarchy 按层级和顺序排序需求
func (g *ExcelBusinessSummaryGenerator) sortRequirementsByHierarchy(requirements []RequirementInfo) []RequirementInfo {
	// 创建层级映射
	level1Map := make(map[uint][]RequirementInfo)
	level2Map := make(map[uint][]RequirementInfo)
	level3Map := make(map[uint][]RequirementInfo)
	level4Map := make(map[uint][]RequirementInfo)

	// 分组
	for _, req := range requirements {
		switch req.Level {
		case 1:
			level1Map[req.ID] = append(level1Map[req.ID], req)
		case 2:
			if req.ParentID != nil {
				level2Map[*req.ParentID] = append(level2Map[*req.ParentID], req)
			}
		case 3:
			if req.ParentID != nil {
				level3Map[*req.ParentID] = append(level3Map[*req.ParentID], req)
			}
		case 4:
			if req.ParentID != nil {
				level4Map[*req.ParentID] = append(level4Map[*req.ParentID], req)
			}
		}
	}

	// 递归构建层级结构
	var result []RequirementInfo
	
	// 获取所有1级需求
	var level1Reqs []RequirementInfo
	for _, req := range requirements {
		if req.Level == 1 {
			level1Reqs = append(level1Reqs, req)
		}
	}

	// 按OrderIndex排序
	g.sortByOrderIndex(level1Reqs)

	// 递归添加子需求
	for _, level1 := range level1Reqs {
		result = append(result, level1)
		result = append(result, g.getChildRequirements(level1.ID, requirements, 2)...)
	}

	return result
}

// getChildRequirements 递归获取子需求
func (g *ExcelBusinessSummaryGenerator) getChildRequirements(parentID uint, allReqs []RequirementInfo, level int) []RequirementInfo {
	var children []RequirementInfo
	var directChildren []RequirementInfo

	// 找到直接子需求
	for _, req := range allReqs {
		if req.Level == level && req.ParentID != nil && *req.ParentID == parentID {
			directChildren = append(directChildren, req)
		}
	}

	// 排序
	g.sortByOrderIndex(directChildren)

	// 递归添加子需求
	for _, child := range directChildren {
		children = append(children, child)
		if level < 4 {
			children = append(children, g.getChildRequirements(child.ID, allReqs, level+1)...)
		}
	}

	return children
}

// sortByOrderIndex 按OrderIndex排序
func (g *ExcelBusinessSummaryGenerator) sortByOrderIndex(reqs []RequirementInfo) {
	for i := 0; i < len(reqs)-1; i++ {
		for j := i + 1; j < len(reqs); j++ {
			if reqs[i].OrderIndex > reqs[j].OrderIndex {
				reqs[i], reqs[j] = reqs[j], reqs[i]
			}
		}
	}
}

// getPriorityText 获取优先级文本
func (g *ExcelBusinessSummaryGenerator) getPriorityText(priority int) string {
	switch priority {
	case 1:
		return "最高"
	case 2:
		return "高"
	case 3:
		return "中"
	case 4:
		return "低"
	case 5:
		return "最低"
	default:
		return "中"
	}
}

// getCompletionPercent 获取完成度百分比
func (g *ExcelBusinessSummaryGenerator) getCompletionPercent(status string) float64 {
	switch strings.ToLower(status) {
	case "completed", "完成":
		return 1.0
	case "in_progress", "进行中":
		return 0.5
	case "testing", "测试中":
		return 0.8
	case "pending", "待开始":
		return 0.0
	default:
		return 0.0
	}
}

// 样式创建方法
func (g *ExcelBusinessSummaryGenerator) createHeaderStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center", WrapText: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"C4D79B"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createLevel1Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E6F3FF"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createLevel2Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 10, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F0F8FF"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createLevel3Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 10, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"F8F8FF"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createLevel4Style(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 9, Family: "宋体"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createDataStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 9, Family: "宋体"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center", WrapText: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createNumberStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 9, Family: "宋体"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		NumFmt: 2, // 数字格式
	})
	return style
}

func (g *ExcelBusinessSummaryGenerator) createPercentStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Size: 9, Family: "宋体"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
		NumFmt: 10, // 百分比格式
	})
	return style
}

// sanitizeFileName 清理文件名
func (g *ExcelBusinessSummaryGenerator) sanitizeFileName(name string) string {
	replacements := map[string]string{
		"/": "_", "\\": "_", ":": "_", "*": "_", "?": "_",
		"\"": "_", "<": "_", ">": "_", "|": "_",
	}
	
	result := name
	for old, new := range replacements {
		result = strings.Replace(result, old, new, -1)
	}
	
	return result
}