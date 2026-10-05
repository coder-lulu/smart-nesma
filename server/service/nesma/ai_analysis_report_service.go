package nesma

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// AIAnalysisReportService AI分析报告服务
type AIAnalysisReportService struct{}

// AnalysisReportData 分析报告数据
type AnalysisReportData struct {
	Analysis               *nesma.AIProjectAnalysis     `json:"analysis"`
	Project                *nesma.NesmaProject          `json:"project"`
	Evaluation             *nesma.NesmaEvaluation       `json:"evaluation"`
	FunctionalAnalysis     *nesma.AIFunctionalAnalysis  `json:"functionalAnalysis"`
	TechnicalAnalysis      *nesma.AITechnicalAnalysis   `json:"technicalAnalysis"`
	BusinessAnalysis       *nesma.AIBusinessAnalysis    `json:"businessAnalysis"`
	QualityAnalysis        *nesma.AIQualityAnalysis     `json:"qualityAnalysis"`
	RiskAnalysis           *nesma.AIRiskAnalysis        `json:"riskAnalysis"`
	RecommendationAnalysis []nesma.AIRecommendationAnalysis `json:"recommendationAnalysis"`
	ComplianceAnalysis     *nesma.AIComplianceAnalysis  `json:"complianceAnalysis"`
}

// GenerateComprehensiveReport 生成综合分析报告
func (s *AIAnalysisReportService) GenerateComprehensiveReport(analysisID uint, format string) (*AnalysisReport, error) {
	global.GVA_LOG.Info("开始生成AI分析报告", zap.Uint("analysisID", analysisID), zap.String("format", format))

	// 获取分析数据
	data, err := s.getAnalysisData(analysisID)
	if err != nil {
		return nil, fmt.Errorf("获取分析数据失败: %w", err)
	}

	// 根据格式生成报告
	switch format {
	case "markdown":
		return s.generateMarkdownReport(data)
	case "json":
		return s.generateJSONReport(data)
	case "html":
		return s.generateHTMLReport(data)
	default:
		return s.generateJSONReport(data)
	}
}

// getAnalysisData 获取分析数据
func (s *AIAnalysisReportService) getAnalysisData(analysisID uint) (*AnalysisReportData, error) {
	data := &AnalysisReportData{}

	// 获取主分析记录
	var analysis nesma.AIProjectAnalysis
	if err := global.GVA_DB.Preload("Project").Preload("Evaluation").
		First(&analysis, analysisID).Error; err != nil {
		return nil, err
	}
	data.Analysis = &analysis
	data.Project = analysis.Project
	data.Evaluation = analysis.Evaluation

	// 获取功能分析
	var functional nesma.AIFunctionalAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).First(&functional).Error; err == nil {
		data.FunctionalAnalysis = &functional
	}

	// 获取技术分析
	var technical nesma.AITechnicalAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).First(&technical).Error; err == nil {
		data.TechnicalAnalysis = &technical
	}

	// 获取业务分析
	var business nesma.AIBusinessAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).First(&business).Error; err == nil {
		data.BusinessAnalysis = &business
	}

	// 获取质量分析
	var quality nesma.AIQualityAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).First(&quality).Error; err == nil {
		data.QualityAnalysis = &quality
	}

	// 获取风险分析
	var risk nesma.AIRiskAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).First(&risk).Error; err == nil {
		data.RiskAnalysis = &risk
	}

	// 获取建议分析
	var recommendations []nesma.AIRecommendationAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).Find(&recommendations).Error; err == nil {
		data.RecommendationAnalysis = recommendations
	}

	// 获取合规分析
	var compliance nesma.AIComplianceAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", analysisID).First(&compliance).Error; err == nil {
		data.ComplianceAnalysis = &compliance
	}

	return data, nil
}

// generateMarkdownReport 生成Markdown格式报告
func (s *AIAnalysisReportService) generateMarkdownReport(data *AnalysisReportData) (*AnalysisReport, error) {
	var content strings.Builder

	// 报告头部
	content.WriteString(fmt.Sprintf("# AI驱动的NESMA项目分析报告\n\n"))
	content.WriteString(fmt.Sprintf("**项目名称**: %s\n", data.Project.Name))
	content.WriteString(fmt.Sprintf("**项目描述**: %s\n", data.Project.Description))
	content.WriteString(fmt.Sprintf("**分析时间**: %s\n", data.Analysis.StartTime.Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("**分析版本**: %s\n", data.Analysis.AnalysisVersion))
	content.WriteString(fmt.Sprintf("**总体评分**: %.2f\n", data.Analysis.OverallScore))
	content.WriteString(fmt.Sprintf("**复杂度等级**: %s\n", data.Analysis.ComplexityLevel))
	content.WriteString(fmt.Sprintf("**风险等级**: %s\n\n", data.Analysis.RiskLevel))

	// 执行摘要
	content.WriteString("## 📊 执行摘要\n\n")
	if data.Analysis.OverallScore >= 0.8 {
		content.WriteString("✅ **项目整体评估良好**，AI分析显示项目具备较好的实施条件和成功前景。\n\n")
	} else if data.Analysis.OverallScore >= 0.6 {
		content.WriteString("⚠️ **项目评估中等**，存在一些需要关注的问题，建议进行优化改进。\n\n")
	} else {
		content.WriteString("❌ **项目存在较多问题**，建议详细评估风险并制定改进计划。\n\n")
	}

	// 功能性分析
	if data.FunctionalAnalysis != nil {
		content.WriteString("## 🎯 功能性分析\n\n")
		content.WriteString(fmt.Sprintf("- **总功能点数**: %.1f\n", data.FunctionalAnalysis.TotalFunctionPoints))
		content.WriteString(fmt.Sprintf("- **数据功能点**: %.1f\n", data.FunctionalAnalysis.DataFunctionPoints))
		content.WriteString(fmt.Sprintf("- **事务功能点**: %.1f\n", data.FunctionalAnalysis.TransactionalFunctionPoints))
		content.WriteString(fmt.Sprintf("- **功能覆盖度**: %.2f\n", data.FunctionalAnalysis.FunctionalCoverage))
		content.WriteString(fmt.Sprintf("- **功能完整性**: %.2f\n", data.FunctionalAnalysis.FunctionalCompleteness))
		content.WriteString(fmt.Sprintf("- **功能一致性**: %.2f\n\n", data.FunctionalAnalysis.FunctionalConsistency))

		content.WriteString("### 功能点分布\n")
		content.WriteString(fmt.Sprintf("- ILF (内部逻辑文件): %d\n", data.FunctionalAnalysis.ILFCount))
		content.WriteString(fmt.Sprintf("- EIF (外部接口文件): %d\n", data.FunctionalAnalysis.EIFCount))
		content.WriteString(fmt.Sprintf("- EI (外部输入): %d\n", data.FunctionalAnalysis.EICount))
		content.WriteString(fmt.Sprintf("- EO (外部输出): %d\n", data.FunctionalAnalysis.EOCount))
		content.WriteString(fmt.Sprintf("- EQ (外部查询): %d\n\n", data.FunctionalAnalysis.EQCount))

		content.WriteString("### 复杂度分布\n")
		content.WriteString(fmt.Sprintf("- 低复杂度: %d\n", data.FunctionalAnalysis.LowComplexityCount))
		content.WriteString(fmt.Sprintf("- 中复杂度: %d\n", data.FunctionalAnalysis.MediumComplexityCount))
		content.WriteString(fmt.Sprintf("- 高复杂度: %d\n\n", data.FunctionalAnalysis.HighComplexityCount))
	}

	// 技术性分析
	if data.TechnicalAnalysis != nil {
		content.WriteString("## 🔧 技术性分析\n\n")
		content.WriteString(fmt.Sprintf("- **架构类型**: %s\n", data.TechnicalAnalysis.ArchitectureType))
		content.WriteString(fmt.Sprintf("- **集成复杂度**: %s\n", data.TechnicalAnalysis.IntegrationComplexity))
		content.WriteString(fmt.Sprintf("- **技术可行性**: %.2f\n", data.TechnicalAnalysis.TechnicalFeasibility))
		content.WriteString(fmt.Sprintf("- **实现风险**: %.2f\n", data.TechnicalAnalysis.ImplementationRisk))
		content.WriteString(fmt.Sprintf("- **技术创新度**: %.2f\n", data.TechnicalAnalysis.TechnicalInnovation))
		content.WriteString(fmt.Sprintf("- **维护复杂度**: %.2f\n\n", data.TechnicalAnalysis.MaintenanceComplexity))

		content.WriteString("### 工作量估算\n")
		content.WriteString(fmt.Sprintf("- **开发工作量**: %.1f 人月\n", data.TechnicalAnalysis.DevelopmentEffort))
		content.WriteString(fmt.Sprintf("- **测试工作量**: %.1f 人月\n", data.TechnicalAnalysis.TestingEffort))
		content.WriteString(fmt.Sprintf("- **部署复杂度**: %.2f\n\n", data.TechnicalAnalysis.DeploymentComplexity))

		if data.TechnicalAnalysis.TechnicalRecommendations != "" {
			content.WriteString("### 技术建议\n")
			content.WriteString(data.TechnicalAnalysis.TechnicalRecommendations)
			content.WriteString("\n\n")
		}
	}

	// 业务性分析
	if data.BusinessAnalysis != nil {
		content.WriteString("## 💼 业务性分析\n\n")
		content.WriteString(fmt.Sprintf("- **业务价值**: %.2f\n", data.BusinessAnalysis.BusinessValue))
		content.WriteString(fmt.Sprintf("- **ROI估算**: %.2f\n", data.BusinessAnalysis.ROIEstimation))
		content.WriteString(fmt.Sprintf("- **战略一致性**: %.2f\n", data.BusinessAnalysis.StrategicAlignment))
		content.WriteString(fmt.Sprintf("- **用户体验评分**: %.2f\n", data.BusinessAnalysis.UserExperienceScore))
		content.WriteString(fmt.Sprintf("- **流程效率**: %.2f\n", data.BusinessAnalysis.ProcessEfficiency))
		content.WriteString(fmt.Sprintf("- **流程自动化**: %.2f\n", data.BusinessAnalysis.ProcessAutomation))
		content.WriteString(fmt.Sprintf("- **市场适配度**: %.2f\n", data.BusinessAnalysis.MarketFit))
		content.WriteString(fmt.Sprintf("- **竞争优势**: %.2f\n", data.BusinessAnalysis.CompetitiveAdvantage))
		content.WriteString(fmt.Sprintf("- **创新水平**: %.2f\n\n", data.BusinessAnalysis.InnovationLevel))

		if data.BusinessAnalysis.BusinessRecommendations != "" {
			content.WriteString("### 业务建议\n")
			content.WriteString(data.BusinessAnalysis.BusinessRecommendations)
			content.WriteString("\n\n")
		}
	}

	// 质量分析
	if data.QualityAnalysis != nil {
		content.WriteString("## 🏆 质量分析\n\n")
		content.WriteString(fmt.Sprintf("- **总体质量**: %.2f\n", data.QualityAnalysis.OverallQuality))
		content.WriteString(fmt.Sprintf("- **需求质量**: %.2f\n", data.QualityAnalysis.RequirementQuality))
		content.WriteString(fmt.Sprintf("- **设计质量**: %.2f\n", data.QualityAnalysis.DesignQuality))
		content.WriteString(fmt.Sprintf("- **正确性**: %.2f\n", data.QualityAnalysis.Correctness))
		content.WriteString(fmt.Sprintf("- **完整性**: %.2f\n", data.QualityAnalysis.Completeness))
		content.WriteString(fmt.Sprintf("- **一致性**: %.2f\n", data.QualityAnalysis.Consistency))
		content.WriteString(fmt.Sprintf("- **清晰性**: %.2f\n", data.QualityAnalysis.Clarity))
		content.WriteString(fmt.Sprintf("- **可追溯性**: %.2f\n", data.QualityAnalysis.Traceability))
		content.WriteString(fmt.Sprintf("- **可维护性**: %.2f\n", data.QualityAnalysis.Maintainability))
		content.WriteString(fmt.Sprintf("- **可重用性**: %.2f\n", data.QualityAnalysis.Reusability))
		content.WriteString(fmt.Sprintf("- **可测试性**: %.2f\n\n", data.QualityAnalysis.Testability))

		if data.QualityAnalysis.QualityRecommendations != "" {
			content.WriteString("### 质量改进建议\n")
			content.WriteString(data.QualityAnalysis.QualityRecommendations)
			content.WriteString("\n\n")
		}
	}

	// 风险分析
	if data.RiskAnalysis != nil {
		content.WriteString("## ⚠️ 风险分析\n\n")
		content.WriteString(fmt.Sprintf("- **总体风险**: %.2f\n", data.RiskAnalysis.OverallRisk))
		content.WriteString(fmt.Sprintf("- **风险等级**: %s\n", data.RiskAnalysis.RiskLevel))
		content.WriteString(fmt.Sprintf("- **技术风险**: %.2f\n", data.RiskAnalysis.TechnicalRisk))
		content.WriteString(fmt.Sprintf("- **实现风险**: %.2f\n", data.RiskAnalysis.ImplementationRisk))
		content.WriteString(fmt.Sprintf("- **集成风险**: %.2f\n", data.RiskAnalysis.IntegrationRisk))
		content.WriteString(fmt.Sprintf("- **进度风险**: %.2f\n", data.RiskAnalysis.ScheduleRisk))
		content.WriteString(fmt.Sprintf("- **预算风险**: %.2f\n", data.RiskAnalysis.BudgetRisk))
		content.WriteString(fmt.Sprintf("- **资源风险**: %.2f\n", data.RiskAnalysis.ResourceRisk))
		content.WriteString(fmt.Sprintf("- **业务风险**: %.2f\n", data.RiskAnalysis.BusinessRisk))
		content.WriteString(fmt.Sprintf("- **合规风险**: %.2f\n\n", data.RiskAnalysis.ComplianceRisk))

		if data.RiskAnalysis.RiskRecommendations != "" {
			content.WriteString("### 风险缓解建议\n")
			content.WriteString(data.RiskAnalysis.RiskRecommendations)
			content.WriteString("\n\n")
		}
	}

	// 改进建议
	if len(data.RecommendationAnalysis) > 0 {
		content.WriteString("## 💡 改进建议\n\n")
		
		// 按优先级分组
		highPriority := []nesma.AIRecommendationAnalysis{}
		mediumPriority := []nesma.AIRecommendationAnalysis{}
		lowPriority := []nesma.AIRecommendationAnalysis{}

		for _, rec := range data.RecommendationAnalysis {
			switch rec.Priority {
			case "high":
				highPriority = append(highPriority, rec)
			case "medium":
				mediumPriority = append(mediumPriority, rec)
			case "low":
				lowPriority = append(lowPriority, rec)
			}
		}

		if len(highPriority) > 0 {
			content.WriteString("### 🔴 高优先级建议\n")
			for _, rec := range highPriority {
				content.WriteString(fmt.Sprintf("- **%s** (%s)\n", rec.Title, rec.RecommendationType))
				content.WriteString(fmt.Sprintf("  - %s\n", rec.Description))
				content.WriteString(fmt.Sprintf("  - 预期收益: %s\n", rec.ExpectedBenefit))
				content.WriteString(fmt.Sprintf("  - 估算工作量: %.1f 人天\n\n", rec.EstimatedEffort))
			}
		}

		if len(mediumPriority) > 0 {
			content.WriteString("### 🟡 中优先级建议\n")
			for _, rec := range mediumPriority {
				content.WriteString(fmt.Sprintf("- **%s** (%s)\n", rec.Title, rec.RecommendationType))
				content.WriteString(fmt.Sprintf("  - %s\n", rec.Description))
				content.WriteString(fmt.Sprintf("  - 预期收益: %s\n\n", rec.ExpectedBenefit))
			}
		}

		if len(lowPriority) > 0 {
			content.WriteString("### 🟢 低优先级建议\n")
			for _, rec := range lowPriority {
				content.WriteString(fmt.Sprintf("- **%s** (%s)\n", rec.Title, rec.RecommendationType))
				content.WriteString(fmt.Sprintf("  - %s\n\n", rec.Description))
			}
		}
	}

	// 合规性分析
	if data.ComplianceAnalysis != nil {
		content.WriteString("## ✅ NESMA合规性分析\n\n")
		content.WriteString(fmt.Sprintf("- **合规标准**: %s %s\n", data.ComplianceAnalysis.ComplianceStandard, data.ComplianceAnalysis.StandardVersion))
		content.WriteString(fmt.Sprintf("- **总体合规性**: %.2f\n", data.ComplianceAnalysis.OverallCompliance))
		content.WriteString(fmt.Sprintf("- **合规等级**: %s\n\n", data.ComplianceAnalysis.ComplianceLevel))

		if data.ComplianceAnalysis.ComplianceRecommendations != "" {
			content.WriteString("### 合规改进建议\n")
			content.WriteString(data.ComplianceAnalysis.ComplianceRecommendations)
			content.WriteString("\n\n")
		}
	}

	// 报告尾部
	content.WriteString("## 📝 报告说明\n\n")
	content.WriteString("本报告由AI驱动的NESMA分析系统自动生成，基于DeepSeek大模型的专业分析能力，\n")
	content.WriteString("结合NESMA 2.2国际标准，对项目进行了全方位的智能评估。\n\n")
	content.WriteString(fmt.Sprintf("**报告生成时间**: %s\n", time.Now().Format("2006-01-02 15:04:05")))
	content.WriteString(fmt.Sprintf("**分析置信度**: %.2f\n", (data.FunctionalAnalysis.ConfidenceScore+data.TechnicalAnalysis.ConfidenceScore+data.BusinessAnalysis.ConfidenceScore)/3.0))

	return &AnalysisReport{
		ID:          fmt.Sprintf("report_%d_%d", data.Analysis.ID, time.Now().Unix()),
		AnalysisID:  data.Analysis.ID,
		Format:      "markdown",
		Title:       fmt.Sprintf("%s - AI分析报告", data.Project.Name),
		Content:     content.String(),
		GeneratedAt: time.Now(),
		Size:        len(content.String()),
	}, nil
}

// generateJSONReport 生成JSON格式报告
func (s *AIAnalysisReportService) generateJSONReport(data *AnalysisReportData) (*AnalysisReport, error) {
	content, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}

	return &AnalysisReport{
		ID:          fmt.Sprintf("report_%d_%d", data.Analysis.ID, time.Now().Unix()),
		AnalysisID:  data.Analysis.ID,
		Format:      "json",
		Title:       fmt.Sprintf("%s - AI分析数据", data.Project.Name),
		Content:     string(content),
		GeneratedAt: time.Now(),
		Size:        len(content),
	}, nil
}

// generateHTMLReport 生成HTML格式报告
func (s *AIAnalysisReportService) generateHTMLReport(data *AnalysisReportData) (*AnalysisReport, error) {
	var content strings.Builder

	// HTML头部
	content.WriteString(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI驱动的NESMA项目分析报告</title>
    <style>
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; margin: 40px; line-height: 1.6; }
        .header { border-bottom: 3px solid #007bff; padding-bottom: 20px; margin-bottom: 30px; }
        .section { margin-bottom: 30px; }
        .metric { display: inline-block; margin: 10px 20px 10px 0; }
        .high-priority { color: #dc3545; }
        .medium-priority { color: #ffc107; }
        .low-priority { color: #28a745; }
        .score-excellent { color: #28a745; font-weight: bold; }
        .score-good { color: #17a2b8; font-weight: bold; }
        .score-warning { color: #ffc107; font-weight: bold; }
        .score-danger { color: #dc3545; font-weight: bold; }
        table { width: 100%; border-collapse: collapse; margin: 20px 0; }
        th, td { border: 1px solid #ddd; padding: 12px; text-align: left; }
        th { background-color: #f8f9fa; }
    </style>
</head>
<body>`)

	// 报告标题
	content.WriteString(fmt.Sprintf(`
    <div class="header">
        <h1>🤖 AI驱动的NESMA项目分析报告</h1>
        <h2>%s</h2>
        <p><strong>分析时间:</strong> %s</p>
        <p><strong>分析版本:</strong> %s</p>
        <p><strong>总体评分:</strong> <span class="%s">%.2f</span></p>
    </div>`,
		data.Project.Name,
		data.Analysis.StartTime.Format("2006-01-02 15:04:05"),
		data.Analysis.AnalysisVersion,
		s.getScoreClass(data.Analysis.OverallScore),
		data.Analysis.OverallScore))

	// 功能性分析
	if data.FunctionalAnalysis != nil {
		content.WriteString(`
    <div class="section">
        <h2>🎯 功能性分析</h2>
        <table>
            <tr><th>指标</th><th>数值</th><th>评价</th></tr>`)
		content.WriteString(fmt.Sprintf(`
            <tr><td>总功能点数</td><td>%.1f</td><td>%s</td></tr>
            <tr><td>数据功能点</td><td>%.1f</td><td>-</td></tr>
            <tr><td>事务功能点</td><td>%.1f</td><td>-</td></tr>
            <tr><td>功能覆盖度</td><td>%.2f</td><td>%s</td></tr>
            <tr><td>功能完整性</td><td>%.2f</td><td>%s</td></tr>`,
			data.FunctionalAnalysis.TotalFunctionPoints,
			s.evaluateFunctionPoints(data.FunctionalAnalysis.TotalFunctionPoints),
			data.FunctionalAnalysis.DataFunctionPoints,
			data.FunctionalAnalysis.TransactionalFunctionPoints,
			data.FunctionalAnalysis.FunctionalCoverage,
			s.evaluateScore(data.FunctionalAnalysis.FunctionalCoverage),
			data.FunctionalAnalysis.FunctionalCompleteness,
			s.evaluateScore(data.FunctionalAnalysis.FunctionalCompleteness)))
		content.WriteString(`
        </table>
    </div>`)
	}

	// 技术性分析
	if data.TechnicalAnalysis != nil {
		content.WriteString(`
    <div class="section">
        <h2>🔧 技术性分析</h2>
        <table>
            <tr><th>指标</th><th>数值</th><th>评价</th></tr>`)
		content.WriteString(fmt.Sprintf(`
            <tr><td>架构类型</td><td>%s</td><td>-</td></tr>
            <tr><td>技术可行性</td><td>%.2f</td><td>%s</td></tr>
            <tr><td>实现风险</td><td>%.2f</td><td>%s</td></tr>
            <tr><td>开发工作量</td><td>%.1f 人月</td><td>-</td></tr>
            <tr><td>测试工作量</td><td>%.1f 人月</td><td>-</td></tr>`,
			data.TechnicalAnalysis.ArchitectureType,
			data.TechnicalAnalysis.TechnicalFeasibility,
			s.evaluateScore(data.TechnicalAnalysis.TechnicalFeasibility),
			data.TechnicalAnalysis.ImplementationRisk,
			s.evaluateRiskScore(data.TechnicalAnalysis.ImplementationRisk),
			data.TechnicalAnalysis.DevelopmentEffort,
			data.TechnicalAnalysis.TestingEffort))
		content.WriteString(`
        </table>
    </div>`)
	}

	// 建议分析
	if len(data.RecommendationAnalysis) > 0 {
		content.WriteString(`
    <div class="section">
        <h2>💡 改进建议</h2>
        <table>
            <tr><th>优先级</th><th>类型</th><th>标题</th><th>描述</th><th>工作量</th></tr>`)
		
		for _, rec := range data.RecommendationAnalysis {
			priorityClass := ""
			switch rec.Priority {
			case "high":
				priorityClass = "high-priority"
			case "medium":
				priorityClass = "medium-priority"
			case "low":
				priorityClass = "low-priority"
			}
			
			content.WriteString(fmt.Sprintf(`
            <tr>
                <td><span class="%s">%s</span></td>
                <td>%s</td>
                <td>%s</td>
                <td>%s</td>
                <td>%.1f 人天</td>
            </tr>`,
				priorityClass, rec.Priority,
				rec.RecommendationType,
				rec.Title,
				rec.Description,
				rec.EstimatedEffort))
		}
		content.WriteString(`
        </table>
    </div>`)
	}

	// HTML尾部
	content.WriteString(fmt.Sprintf(`
    <div class="section">
        <h2>📝 报告说明</h2>
        <p>本报告由AI驱动的NESMA分析系统自动生成，基于DeepSeek大模型的专业分析能力，结合NESMA 2.2国际标准。</p>
        <p><strong>报告生成时间:</strong> %s</p>
    </div>
</body>
</html>`, time.Now().Format("2006-01-02 15:04:05")))

	return &AnalysisReport{
		ID:          fmt.Sprintf("report_%d_%d", data.Analysis.ID, time.Now().Unix()),
		AnalysisID:  data.Analysis.ID,
		Format:      "html",
		Title:       fmt.Sprintf("%s - AI分析报告", data.Project.Name),
		Content:     content.String(),
		GeneratedAt: time.Now(),
		Size:        len(content.String()),
	}, nil
}

// 工具方法
func (s *AIAnalysisReportService) getScoreClass(score float64) string {
	if score >= 0.9 {
		return "score-excellent"
	} else if score >= 0.8 {
		return "score-good"
	} else if score >= 0.6 {
		return "score-warning"
	} else {
		return "score-danger"
	}
}

func (s *AIAnalysisReportService) evaluateScore(score float64) string {
	if score >= 0.9 {
		return "优秀"
	} else if score >= 0.8 {
		return "良好"
	} else if score >= 0.6 {
		return "一般"
	} else {
		return "需改进"
	}
}

func (s *AIAnalysisReportService) evaluateRiskScore(score float64) string {
	if score >= 0.7 {
		return "高风险"
	} else if score >= 0.4 {
		return "中风险"
	} else {
		return "低风险"
	}
}

func (s *AIAnalysisReportService) evaluateFunctionPoints(fp float64) string {
	if fp < 100 {
		return "小型项目"
	} else if fp < 500 {
		return "中型项目"
	} else if fp < 1000 {
		return "大型项目"
	} else {
		return "超大型项目"
	}
}

// AnalysisReport 分析报告
type AnalysisReport struct {
	ID          string    `json:"id"`
	AnalysisID  uint      `json:"analysisId"`
	Format      string    `json:"format"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	GeneratedAt time.Time `json:"generatedAt"`
	Size        int       `json:"size"`
}