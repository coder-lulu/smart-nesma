package nesma

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
)

// ExcelNesmaReportGenerator Excel NESMA报告生成器
type ExcelNesmaReportGenerator struct {
	storePath string
}

// NewExcelNesmaReportGenerator 创建Excel NESMA报告生成器
func NewExcelNesmaReportGenerator() *ExcelNesmaReportGenerator {
	return &ExcelNesmaReportGenerator{
		storePath: filepath.Join(global.GVA_CONFIG.Local.StorePath, "exports"),
	}
}

// GetSupportedType 获取支持的导出类型
func (g *ExcelNesmaReportGenerator) GetSupportedType() ExportType {
	return ExportExcelNesmaReport
}

// Validate 验证数据
func (g *ExcelNesmaReportGenerator) Validate(data *ProjectExportData) error {
	if data == nil {
		return NewExportError(ErrDataGeneration, "导出数据为空")
	}
	
	if data.Project.Name == "" {
		return NewExportError(ErrDataGeneration, "项目名称不能为空")
	}
	
	// 检查是否有4级需求（功能点）
	hasLevel4 := false
	for _, req := range data.Requirements {
		if req.Level == 4 {
			hasLevel4 = true
			break
		}
	}
	
	if !hasLevel4 {
		return NewExportError(ErrDataGeneration, "项目没有功能点数据，无法生成NESMA报告")
	}
	
	return nil
}

// Generate 生成Excel文档
func (g *ExcelNesmaReportGenerator) Generate(data *ProjectExportData, config ExportConfig) (string, error) {
	global.GVA_LOG.Info("开始生成Excel NESMA报告", 
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
	if err := g.generateProjectInfoSheet(f, data, config); err != nil {
		return "", err
	}

	if err := g.generateDetailSheet(f, data, config); err != nil {
		return "", err
	}

	if err := g.generateComplexitySheet(f, data, config); err != nil {
		return "", err
	}

	// 删除默认工作表
	f.DeleteSheet("Sheet1")

	// 生成文件路径
	fileName := fmt.Sprintf("%s_NESMA分析报告_%s.xlsx", 
		g.sanitizeFileName(data.Project.Name),
		time.Now().Format("20060102150405"))
	
	filePath := filepath.Join(g.storePath, fileName)

	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return "", NewExportError(ErrFileGeneration, "保存Excel文件失败", err.Error())
	}

	global.GVA_LOG.Info("Excel NESMA报告生成完成", 
		zap.String("文件路径", filePath))

	return filePath, nil
}

// generateProjectInfoSheet 生成项目信息与汇总表
func (g *ExcelNesmaReportGenerator) generateProjectInfoSheet(f *excelize.File, data *ProjectExportData, config ExportConfig) error {
	sheetName := "项目信息与汇总"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}
	f.SetActiveSheet(index)

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 15)
	f.SetColWidth(sheetName, "B", "B", 25)
	f.SetColWidth(sheetName, "C", "C", 8)
	f.SetColWidth(sheetName, "D", "D", 15)
	f.SetColWidth(sheetName, "E", "E", 8)
	f.SetColWidth(sheetName, "F", "F", 12)

	// 创建样式
	titleStyle := g.createTitleStyle(f)
	headerStyle := g.createHeaderStyle(f)
	dataStyle := g.createDataStyle(f)
	numberStyle := g.createNumberStyle(f)

	// 报告标题
	f.SetCellValue(sheetName, "A1", data.Project.Name+" NESMA功能点分析报告")
	f.SetCellStyle(sheetName, "A1", "A1", titleStyle)
	f.MergeCell(sheetName, "A1", "F1")

	// 项目基本信息
	infoStartRow := 3
	infoItems := [][]string{
		{"项目名称", data.Project.Name},
		{"项目领域", data.Project.Domain},
		{"项目状态", data.Project.Status},
		{"创建时间", data.GeneratedAt.Format("2006-01-02 15:04:05")},
	}

	if data.Cycle != nil {
		infoItems = append(infoItems, []string{"分析周期", data.Cycle.Name})
	}
	if data.Version != nil {
		infoItems = append(infoItems, []string{"分析版本", data.Version.Version})
	}

	for i, item := range infoItems {
		row := infoStartRow + i
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), item[0])
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), item[1])
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), headerStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)
	}

	// 功能点统计汇总
	statsStartRow := infoStartRow + len(infoItems) + 2
	
	// 统计表标题
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", statsStartRow-1), "功能点统计汇总")
	f.SetCellStyle(sheetName, fmt.Sprintf("D%d", statsStartRow-1), fmt.Sprintf("D%d", statsStartRow-1), titleStyle)

	// 统计表头
	statsHeaders := []string{"功能类型", "数量", "UFP小计"}
	for i, header := range statsHeaders {
		cell := fmt.Sprintf("%s%d", string(rune('D'+i)), statsStartRow)
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 统计数据
	stats := data.Statistics
	statsData := [][]interface{}{
		{"EI外部输入", stats.EICount, stats.EIUFP},
		{"EO外部输出", stats.EOCount, stats.EOUFP},
		{"EQ外部查询", stats.EQCount, stats.EQUFP},
		{"ILF内部文件", stats.ILFCount, stats.ILFUFP},
		{"EIF外部文件", stats.EIFCount, stats.EIFUFP},
	}

	for i, rowData := range statsData {
		row := statsStartRow + 1 + i
		for j, value := range rowData {
			cell := fmt.Sprintf("%s%d", string(rune('D'+j)), row)
			f.SetCellValue(sheetName, cell, value)
			if j == 0 {
				f.SetCellStyle(sheetName, cell, cell, dataStyle)
			} else {
				f.SetCellStyle(sheetName, cell, cell, numberStyle)
			}
		}
	}

	// 合计行
	totalRow := statsStartRow + 1 + len(statsData)
	f.SetCellValue(sheetName, fmt.Sprintf("D%d", totalRow), "合计")
	f.SetCellValue(sheetName, fmt.Sprintf("E%d", totalRow), stats.Level4Count)
	f.SetCellValue(sheetName, fmt.Sprintf("F%d", totalRow), stats.TotalUFP)
	
	// 设置合计行样式
	for col := 'D'; col <= 'F'; col++ {
		cell := fmt.Sprintf("%s%d", string(col), totalRow)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 调整因子信息
	adjustmentStartRow := totalRow + 2
	f.SetCellValue(sheetName, fmt.Sprintf("A%d", adjustmentStartRow), "调整因子")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", adjustmentStartRow), stats.AdjustmentFactor)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", adjustmentStartRow), fmt.Sprintf("A%d", adjustmentStartRow), headerStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", adjustmentStartRow), fmt.Sprintf("B%d", adjustmentStartRow), numberStyle)

	f.SetCellValue(sheetName, fmt.Sprintf("A%d", adjustmentStartRow+1), "调整后AFP")
	f.SetCellValue(sheetName, fmt.Sprintf("B%d", adjustmentStartRow+1), stats.TotalAFP)
	f.SetCellStyle(sheetName, fmt.Sprintf("A%d", adjustmentStartRow+1), fmt.Sprintf("A%d", adjustmentStartRow+1), headerStyle)
	f.SetCellStyle(sheetName, fmt.Sprintf("B%d", adjustmentStartRow+1), fmt.Sprintf("B%d", adjustmentStartRow+1), numberStyle)

	if stats.TotalEstimateHours > 0 {
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", adjustmentStartRow+2), "预估工时")
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", adjustmentStartRow+2), fmt.Sprintf("%.1f小时", stats.TotalEstimateHours))
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", adjustmentStartRow+2), fmt.Sprintf("A%d", adjustmentStartRow+2), headerStyle)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", adjustmentStartRow+2), fmt.Sprintf("B%d", adjustmentStartRow+2), dataStyle)
	}

	return nil
}

// generateDetailSheet 生成功能点详细清单
func (g *ExcelNesmaReportGenerator) generateDetailSheet(f *excelize.File, data *ProjectExportData, config ExportConfig) error {
	sheetName := "功能点详细清单"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// 设置列宽
	columnWidths := []struct{ col, width string }{
		{"A", "6"},   // 序号
		{"B", "15"},  // 需求编号
		{"C", "20"},  // 一级模块
		{"D", "20"},  // 二级模块
		{"E", "20"},  // 三级模块
		{"F", "25"},  // 功能点名称
		{"G", "8"},   // 功能类型
		{"H", "8"},   // 复杂度
		{"I", "8"},   // UFP
		{"J", "8"},   // AFP
		{"K", "10"},  // AI置信度
		{"L", "20"},  // 备注
	}

	for _, cw := range columnWidths {
		f.SetColWidth(sheetName, cw.col, cw.col, parseFloat(cw.width))
	}

	// 创建样式
	headerStyle := g.createHeaderStyle(f)
	dataStyle := g.createDataStyle(f)
	numberStyle := g.createNumberStyle(f)

	// 表头
	headers := []string{
		"序号", "需求编号", "一级模块", "二级模块", "三级模块",
		"功能点名称", "功能类型", "复杂度", "UFP", "AFP", "AI置信度", "备注",
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

	// 数据行 - 只输出4级需求（功能点）
	row := 2
	index4 := 1
	for _, req := range data.Requirements {
		if req.Level != 4 {
			continue
		}

		// 序号
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), index4)
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), numberStyle)

		// 需求编号
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), req.Code)
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)

		// 层级标题
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), req.Level1Title)
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), req.Level2Title)
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), req.Level3Title)
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("E%d", row), dataStyle)

		// 功能点名称
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), req.Title)
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), dataStyle)

		// 功能类型
		f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), req.FunctionType)
		f.SetCellStyle(sheetName, fmt.Sprintf("G%d", row), fmt.Sprintf("G%d", row), dataStyle)

		// 复杂度
		f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), req.Complexity)
		f.SetCellStyle(sheetName, fmt.Sprintf("H%d", row), fmt.Sprintf("H%d", row), dataStyle)

		// UFP
		f.SetCellValue(sheetName, fmt.Sprintf("I%d", row), req.UFP)
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), numberStyle)

		// AFP
		f.SetCellValue(sheetName, fmt.Sprintf("J%d", row), req.AFP)
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), numberStyle)

		// AI置信度
		if req.AIConfidenceScore != nil {
			f.SetCellValue(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("%.1f%%", *req.AIConfidenceScore*100))
		}
		f.SetCellStyle(sheetName, fmt.Sprintf("K%d", row), fmt.Sprintf("K%d", row), dataStyle)

		// 备注
		notes := req.Notes
		if config.IncludeAIAnalysis && req.AIDescription != "" {
			if notes != "" {
				notes += "; "
			}
			notes += "AI优化：" + req.AIDescription[:min(50, len(req.AIDescription))] + "..."
		}
		f.SetCellValue(sheetName, fmt.Sprintf("L%d", row), notes)
		f.SetCellStyle(sheetName, fmt.Sprintf("L%d", row), fmt.Sprintf("L%d", row), dataStyle)

		row++
		index4++
	}

	// 合计行
	if row > 2 {
		f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), "合计")
		f.SetCellStyle(sheetName, fmt.Sprintf("F%d", row), fmt.Sprintf("F%d", row), headerStyle)

		// UFP合计公式
		f.SetCellFormula(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("SUM(I2:I%d)", row-1))
		f.SetCellStyle(sheetName, fmt.Sprintf("I%d", row), fmt.Sprintf("I%d", row), headerStyle)

		// AFP合计公式
		f.SetCellFormula(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("SUM(J2:J%d)", row-1))
		f.SetCellStyle(sheetName, fmt.Sprintf("J%d", row), fmt.Sprintf("J%d", row), headerStyle)
	}

	return nil
}

// generateComplexitySheet 生成复杂度分析表
func (g *ExcelNesmaReportGenerator) generateComplexitySheet(f *excelize.File, data *ProjectExportData, config ExportConfig) error {
	sheetName := "复杂度分析"
	_, err := f.NewSheet(sheetName)
	if err != nil {
		return err
	}

	// 设置列宽
	f.SetColWidth(sheetName, "A", "A", 12) // 功能类型
	f.SetColWidth(sheetName, "B", "E", 15) // 复杂度列

	// 创建样式
	headerStyle := g.createHeaderStyle(f)
	dataStyle := g.createDataStyle(f)
	numberStyle := g.createNumberStyle(f)

	// 表头
	headers := []string{"功能类型", "简单", "中等", "复杂", "小计UFP"}
	for i, header := range headers {
		cell := fmt.Sprintf("%s1", string(rune('A'+i)))
		f.SetCellValue(sheetName, cell, header)
		f.SetCellStyle(sheetName, cell, cell, headerStyle)
	}

	// 统计各功能类型的复杂度分布
	complexityStats := g.calculateComplexityStats(data.Requirements)

	row := 2
	functionTypes := []string{"EI", "EO", "EQ", "ILF", "EIF"}
	
	for _, funcType := range functionTypes {
		stats, exists := complexityStats[funcType]
		if !exists {
			stats = FunctionComplexityStats{}
		}

		// 功能类型名称
		typeNames := map[string]string{
			"EI":  "EI外部输入",
			"EO":  "EO外部输出", 
			"EQ":  "EQ外部查询",
			"ILF": "ILF内部文件",
			"EIF": "EIF外部文件",
		}
		
		f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), typeNames[funcType])
		f.SetCellStyle(sheetName, fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), dataStyle)

		// 简单
		f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), 
			fmt.Sprintf("%d(%.1f)", stats.SimpleCount, stats.SimpleUFP))
		f.SetCellStyle(sheetName, fmt.Sprintf("B%d", row), fmt.Sprintf("B%d", row), dataStyle)

		// 中等
		f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), 
			fmt.Sprintf("%d(%.1f)", stats.AverageCount, stats.AverageUFP))
		f.SetCellStyle(sheetName, fmt.Sprintf("C%d", row), fmt.Sprintf("C%d", row), dataStyle)

		// 复杂
		f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), 
			fmt.Sprintf("%d(%.1f)", stats.ComplexCount, stats.ComplexUFP))
		f.SetCellStyle(sheetName, fmt.Sprintf("D%d", row), fmt.Sprintf("D%d", row), dataStyle)

		// 小计UFP
		totalUFP := stats.SimpleUFP + stats.AverageUFP + stats.ComplexUFP
		f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), totalUFP)
		f.SetCellStyle(sheetName, fmt.Sprintf("E%d", row), fmt.Sprintf("E%d", row), numberStyle)

		row++
	}

	return nil
}

// FunctionComplexityStats 功能复杂度统计结构
type FunctionComplexityStats struct {
	SimpleCount   int
	AverageCount  int
	ComplexCount  int
	SimpleUFP     float64
	AverageUFP    float64
	ComplexUFP    float64
}

// calculateComplexityStats 计算复杂度统计
func (g *ExcelNesmaReportGenerator) calculateComplexityStats(requirements []RequirementInfo) map[string]FunctionComplexityStats {
	stats := make(map[string]FunctionComplexityStats)

	for _, req := range requirements {
		if req.Level != 4 || req.FunctionType == "" {
			continue
		}

		funcType := strings.ToUpper(req.FunctionType)
		stat := stats[funcType]

		switch strings.ToLower(req.Complexity) {
		case "简单", "low", "simple":
			stat.SimpleCount++
			stat.SimpleUFP += req.UFP
		case "中等", "average", "medium":
			stat.AverageCount++
			stat.AverageUFP += req.UFP
		case "复杂", "high", "complex":
			stat.ComplexCount++
			stat.ComplexUFP += req.UFP
		}

		stats[funcType] = stat
	}

	return stats
}

// 样式创建方法
func (g *ExcelNesmaReportGenerator) createTitleStyle(f *excelize.File) int {
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 14, Family: "微软雅黑"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E6F3FF"}, Pattern: 1},
	})
	return style
}

func (g *ExcelNesmaReportGenerator) createHeaderStyle(f *excelize.File) int {
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

func (g *ExcelNesmaReportGenerator) createDataStyle(f *excelize.File) int {
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

func (g *ExcelNesmaReportGenerator) createNumberStyle(f *excelize.File) int {
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

// 辅助函数
func (g *ExcelNesmaReportGenerator) sanitizeFileName(name string) string {
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

func parseFloat(s string) float64 {
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return 10.0
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}