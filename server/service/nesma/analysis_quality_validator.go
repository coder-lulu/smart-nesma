package nesma

import (
	"fmt"
	"math"
	"strings"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// AnalysisQualityValidator 分析质量验证器
type AnalysisQualityValidator struct{}

// QualityValidationResult 质量验证结果
type QualityValidationResult struct {
	OverallQuality     float64                    `json:"overallQuality"`
	ConfidenceScore    float64                    `json:"confidenceScore"`
	ValidationIssues   []ValidationIssue          `json:"validationIssues"`
	QualityMetrics     map[string]float64         `json:"qualityMetrics"`
	RecommendedActions []string                   `json:"recommendedActions"`
	ValidationReport   string                     `json:"validationReport"`
	DimensionQuality   map[string]DimensionQuality `json:"dimensionQuality"`
}

// ValidationIssue 验证问题
type ValidationIssue struct {
	Type        string  `json:"type"`
	Severity    string  `json:"severity"` // critical, warning, info
	Dimension   string  `json:"dimension"`
	Description string  `json:"description"`
	Impact      float64 `json:"impact"`
	Suggestion  string  `json:"suggestion"`
}

// DimensionQuality 维度质量
type DimensionQuality struct {
	Score       float64   `json:"score"`
	Confidence  float64   `json:"confidence"`
	Issues      []string  `json:"issues"`
	Strengths   []string  `json:"strengths"`
	Grade       string    `json:"grade"`
}

// ValidateAnalysisQuality 验证分析质量
func (v *AnalysisQualityValidator) ValidateAnalysisQuality(analysisID uint) (*QualityValidationResult, error) {
	global.GVA_LOG.Info("开始验证分析质量", zap.Uint("analysisID", analysisID))

	// 获取分析数据
	data, err := v.getAnalysisData(analysisID)
	if err != nil {
		return nil, fmt.Errorf("获取分析数据失败: %w", err)
	}

	result := &QualityValidationResult{
		ValidationIssues:   []ValidationIssue{},
		QualityMetrics:     make(map[string]float64),
		RecommendedActions: []string{},
		DimensionQuality:   make(map[string]DimensionQuality),
	}

	// 验证各个维度
	v.validateFunctionalAnalysis(data, result)
	v.validateTechnicalAnalysis(data, result)
	v.validateBusinessAnalysis(data, result)
	v.validateQualityAnalysis(data, result)
	v.validateRiskAnalysis(data, result)
	v.validateRecommendationAnalysis(data, result)
	v.validateComplianceAnalysis(data, result)

	// 计算总体质量和置信度
	result.OverallQuality = v.calculateOverallQuality(result)
	result.ConfidenceScore = v.calculateOverallConfidence(result)

	// 生成验证报告
	result.ValidationReport = v.generateValidationReport(result)

	// 生成推荐行动
	result.RecommendedActions = v.generateRecommendedActions(result)

	global.GVA_LOG.Info("分析质量验证完成",
		zap.Uint("analysisID", analysisID),
		zap.Float64("overallQuality", result.OverallQuality),
		zap.Float64("confidenceScore", result.ConfidenceScore),
		zap.Int("issueCount", len(result.ValidationIssues)))

	return result, nil
}

// getAnalysisData 获取分析数据
func (v *AnalysisQualityValidator) getAnalysisData(analysisID uint) (*AnalysisReportData, error) {
	reportService := &AIAnalysisReportService{}
	return reportService.getAnalysisData(analysisID)
}

// validateFunctionalAnalysis 验证功能分析
func (v *AnalysisQualityValidator) validateFunctionalAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if data.FunctionalAnalysis == nil {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "missing_analysis",
			Severity:    "critical",
			Dimension:   "functional",
			Description: "缺少功能性分析结果",
			Impact:      0.8,
			Suggestion:  "重新执行功能性分析",
		})
		return
	}

	fa := data.FunctionalAnalysis
	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	// 验证功能点数量合理性
	if fa.TotalFunctionPoints <= 0 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "invalid_value",
			Severity:    "critical",
			Dimension:   "functional",
			Description: "功能点总数为零或负数",
			Impact:      0.9,
			Suggestion:  "检查功能点识别算法",
		})
		quality.Issues = append(quality.Issues, "功能点总数异常")
	} else {
		quality.Strengths = append(quality.Strengths, "功能点识别正常")
	}

	// 验证数据与事务功能点比例
	if fa.TotalFunctionPoints > 0 {
		dataRatio := fa.DataFunctionPoints / fa.TotalFunctionPoints
		transRatio := fa.TransactionalFunctionPoints / fa.TotalFunctionPoints

		if dataRatio < 0.1 || dataRatio > 0.8 {
			result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
				Type:        "proportion_warning",
				Severity:    "warning",
				Dimension:   "functional",
				Description: fmt.Sprintf("数据功能点比例异常: %.2f", dataRatio),
				Impact:      0.3,
				Suggestion:  "检查数据功能识别准确性",
			})
			quality.Issues = append(quality.Issues, "数据功能点比例异常")
		}

		if transRatio < 0.2 || transRatio > 0.9 {
			result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
				Type:        "proportion_warning",
				Severity:    "warning",
				Dimension:   "functional",
				Description: fmt.Sprintf("事务功能点比例异常: %.2f", transRatio),
				Impact:      0.3,
				Suggestion:  "检查事务功能识别准确性",
			})
			quality.Issues = append(quality.Issues, "事务功能点比例异常")
		}
	}

	// 验证置信度
	if fa.ConfidenceScore < 0.5 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "low_confidence",
			Severity:    "warning",
			Dimension:   "functional",
			Description: fmt.Sprintf("功能分析置信度较低: %.2f", fa.ConfidenceScore),
			Impact:      0.4,
			Suggestion:  "增加更多的分析样本或优化分析算法",
		})
		quality.Issues = append(quality.Issues, "分析置信度偏低")
	} else {
		quality.Strengths = append(quality.Strengths, "分析置信度良好")
	}

	// 计算维度质量评分
	quality.Score = v.calculateDimensionScore(fa.FunctionalCompleteness, fa.FunctionalCoverage, fa.FunctionalConsistency)
	quality.Confidence = fa.ConfidenceScore
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["functional"] = quality
	result.QualityMetrics["functional_score"] = quality.Score
	result.QualityMetrics["functional_confidence"] = quality.Confidence
}

// validateTechnicalAnalysis 验证技术分析
func (v *AnalysisQualityValidator) validateTechnicalAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if data.TechnicalAnalysis == nil {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "missing_analysis",
			Severity:    "warning",
			Dimension:   "technical",
			Description: "缺少技术性分析结果",
			Impact:      0.5,
			Suggestion:  "执行技术性分析",
		})
		return
	}

	ta := data.TechnicalAnalysis
	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	// 验证技术可行性
	if ta.TechnicalFeasibility < 0.3 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "high_risk",
			Severity:    "critical",
			Dimension:   "technical",
			Description: fmt.Sprintf("技术可行性过低: %.2f", ta.TechnicalFeasibility),
			Impact:      0.8,
			Suggestion:  "重新评估技术方案或降低技术风险",
		})
		quality.Issues = append(quality.Issues, "技术可行性偏低")
	} else if ta.TechnicalFeasibility > 0.8 {
		quality.Strengths = append(quality.Strengths, "技术可行性良好")
	}

	// 验证实现风险
	if ta.ImplementationRisk > 0.7 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "high_risk",
			Severity:    "warning",
			Dimension:   "technical",
			Description: fmt.Sprintf("实现风险较高: %.2f", ta.ImplementationRisk),
			Impact:      0.6,
			Suggestion:  "制定风险缓解计划",
		})
		quality.Issues = append(quality.Issues, "实现风险较高")
	} else {
		quality.Strengths = append(quality.Strengths, "实现风险可控")
	}

	// 验证工作量估算合理性
	if ta.DevelopmentEffort <= 0 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "invalid_estimate",
			Severity:    "warning",
			Dimension:   "technical",
			Description: "开发工作量估算异常",
			Impact:      0.4,
			Suggestion:  "重新估算开发工作量",
		})
		quality.Issues = append(quality.Issues, "工作量估算异常")
	}

	quality.Score = v.calculateDimensionScore(ta.TechnicalFeasibility, 1.0-ta.ImplementationRisk, ta.TechnicalInnovation)
	quality.Confidence = ta.ConfidenceScore
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["technical"] = quality
	result.QualityMetrics["technical_score"] = quality.Score
	result.QualityMetrics["technical_confidence"] = quality.Confidence
}

// validateBusinessAnalysis 验证业务分析
func (v *AnalysisQualityValidator) validateBusinessAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if data.BusinessAnalysis == nil {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "missing_analysis",
			Severity:    "warning",
			Dimension:   "business",
			Description: "缺少业务性分析结果",
			Impact:      0.4,
			Suggestion:  "执行业务性分析",
		})
		return
	}

	ba := data.BusinessAnalysis
	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	// 验证业务价值
	if ba.BusinessValue < 0.4 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "low_value",
			Severity:    "warning",
			Dimension:   "business",
			Description: fmt.Sprintf("业务价值偏低: %.2f", ba.BusinessValue),
			Impact:      0.5,
			Suggestion:  "重新评估业务价值或调整项目方向",
		})
		quality.Issues = append(quality.Issues, "业务价值偏低")
	} else {
		quality.Strengths = append(quality.Strengths, "业务价值合理")
	}

	// 验证ROI
	if ba.ROIEstimation < 0.1 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "low_roi",
			Severity:    "warning",
			Dimension:   "business",
			Description: fmt.Sprintf("投资回报率过低: %.2f", ba.ROIEstimation),
			Impact:      0.6,
			Suggestion:  "重新评估成本效益或优化投资结构",
		})
		quality.Issues = append(quality.Issues, "投资回报率偏低")
	} else {
		quality.Strengths = append(quality.Strengths, "投资回报率合理")
	}

	quality.Score = v.calculateDimensionScore(ba.BusinessValue, ba.ROIEstimation, ba.StrategicAlignment)
	quality.Confidence = ba.ConfidenceScore
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["business"] = quality
	result.QualityMetrics["business_score"] = quality.Score
	result.QualityMetrics["business_confidence"] = quality.Confidence
}

// validateQualityAnalysis 验证质量分析
func (v *AnalysisQualityValidator) validateQualityAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if data.QualityAnalysis == nil {
		return // 质量分析是可选的
	}

	qa := data.QualityAnalysis
	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	// 验证质量指标
	if qa.OverallQuality < 0.5 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "quality_concern",
			Severity:    "warning",
			Dimension:   "quality",
			Description: fmt.Sprintf("总体质量评分偏低: %.2f", qa.OverallQuality),
			Impact:      0.5,
			Suggestion:  "制定质量改进计划",
		})
		quality.Issues = append(quality.Issues, "总体质量偏低")
	} else {
		quality.Strengths = append(quality.Strengths, "质量评分良好")
	}

	quality.Score = qa.OverallQuality
	quality.Confidence = qa.ConfidenceScore
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["quality"] = quality
	result.QualityMetrics["quality_score"] = quality.Score
	result.QualityMetrics["quality_confidence"] = quality.Confidence
}

// validateRiskAnalysis 验证风险分析
func (v *AnalysisQualityValidator) validateRiskAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if data.RiskAnalysis == nil {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "missing_analysis",
			Severity:    "warning",
			Dimension:   "risk",
			Description: "缺少风险分析结果",
			Impact:      0.4,
			Suggestion:  "执行风险分析",
		})
		return
	}

	ra := data.RiskAnalysis
	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	// 验证风险水平
	if ra.OverallRisk > 0.8 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "high_risk",
			Severity:    "critical",
			Dimension:   "risk",
			Description: fmt.Sprintf("总体风险过高: %.2f", ra.OverallRisk),
			Impact:      0.9,
			Suggestion:  "制定全面的风险缓解策略",
		})
		quality.Issues = append(quality.Issues, "总体风险过高")
	} else if ra.OverallRisk < 0.3 {
		quality.Strengths = append(quality.Strengths, "风险水平可控")
	}

	quality.Score = 1.0 - ra.OverallRisk // 风险越低，质量评分越高
	quality.Confidence = ra.ConfidenceScore
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["risk"] = quality
	result.QualityMetrics["risk_score"] = quality.Score
	result.QualityMetrics["risk_confidence"] = quality.Confidence
}

// validateRecommendationAnalysis 验证建议分析
func (v *AnalysisQualityValidator) validateRecommendationAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if len(data.RecommendationAnalysis) == 0 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "missing_recommendations",
			Severity:    "info",
			Dimension:   "recommendation",
			Description: "缺少改进建议",
			Impact:      0.2,
			Suggestion:  "生成项目改进建议",
		})
		return
	}

	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	highPriorityCount := 0
	for _, rec := range data.RecommendationAnalysis {
		if rec.Priority == "high" {
			highPriorityCount++
		}
	}

	if highPriorityCount > 10 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "too_many_high_priority",
			Severity:    "warning",
			Dimension:   "recommendation",
			Description: fmt.Sprintf("高优先级建议过多: %d", highPriorityCount),
			Impact:      0.3,
			Suggestion:  "重新评估建议优先级",
		})
		quality.Issues = append(quality.Issues, "高优先级建议过多")
	} else {
		quality.Strengths = append(quality.Strengths, "建议优先级分布合理")
	}

	quality.Score = math.Min(1.0, float64(len(data.RecommendationAnalysis))/10.0) // 建议数量合理性
	quality.Confidence = 0.8 // 建议分析的置信度相对固定
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["recommendation"] = quality
	result.QualityMetrics["recommendation_score"] = quality.Score
	result.QualityMetrics["recommendation_confidence"] = quality.Confidence
}

// validateComplianceAnalysis 验证合规分析
func (v *AnalysisQualityValidator) validateComplianceAnalysis(data *AnalysisReportData, result *QualityValidationResult) {
	if data.ComplianceAnalysis == nil {
		return // 合规分析是可选的
	}

	ca := data.ComplianceAnalysis
	quality := DimensionQuality{
		Issues:    []string{},
		Strengths: []string{},
	}

	if ca.OverallCompliance < 0.6 {
		result.ValidationIssues = append(result.ValidationIssues, ValidationIssue{
			Type:        "compliance_issue",
			Severity:    "warning",
			Dimension:   "compliance",
			Description: fmt.Sprintf("NESMA合规性不足: %.2f", ca.OverallCompliance),
			Impact:      0.4,
			Suggestion:  "改进分析方法以提高NESMA合规性",
		})
		quality.Issues = append(quality.Issues, "合规性不足")
	} else {
		quality.Strengths = append(quality.Strengths, "NESMA合规性良好")
	}

	quality.Score = ca.OverallCompliance
	quality.Confidence = ca.ConfidenceScore
	quality.Grade = v.getGrade(quality.Score)

	result.DimensionQuality["compliance"] = quality
	result.QualityMetrics["compliance_score"] = quality.Score
	result.QualityMetrics["compliance_confidence"] = quality.Confidence
}

// calculateOverallQuality 计算总体质量
func (v *AnalysisQualityValidator) calculateOverallQuality(result *QualityValidationResult) float64 {
	weights := map[string]float64{
		"functional":     0.35,
		"technical":      0.25,
		"business":       0.20,
		"quality":        0.10,
		"risk":           0.05,
		"recommendation": 0.03,
		"compliance":     0.02,
	}

	totalScore := 0.0
	totalWeight := 0.0

	for dimension, weight := range weights {
		if quality, exists := result.DimensionQuality[dimension]; exists {
			totalScore += quality.Score * weight
			totalWeight += weight
		}
	}

	if totalWeight > 0 {
		baseScore := totalScore / totalWeight
		
		// 根据问题严重程度调整评分
		penalty := 0.0
		for _, issue := range result.ValidationIssues {
			switch issue.Severity {
			case "critical":
				penalty += issue.Impact * 0.3
			case "warning":
				penalty += issue.Impact * 0.1
			case "info":
				penalty += issue.Impact * 0.02
			}
		}
		
		return math.Max(0.0, baseScore-penalty)
	}

	return 0.0
}

// calculateOverallConfidence 计算总体置信度
func (v *AnalysisQualityValidator) calculateOverallConfidence(result *QualityValidationResult) float64 {
	totalConfidence := 0.0
	count := 0

	for _, quality := range result.DimensionQuality {
		totalConfidence += quality.Confidence
		count++
	}

	if count > 0 {
		baseConfidence := totalConfidence / float64(count)
		
		// 根据问题数量调整置信度
		issueCount := len(result.ValidationIssues)
		confidencePenalty := math.Min(0.3, float64(issueCount)*0.05)
		
		return math.Max(0.0, baseConfidence-confidencePenalty)
	}

	return 0.0
}

// generateValidationReport 生成验证报告
func (v *AnalysisQualityValidator) generateValidationReport(result *QualityValidationResult) string {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("# AI分析质量验证报告\n\n"))
	report.WriteString(fmt.Sprintf("**总体质量评分**: %.2f\n", result.OverallQuality))
	report.WriteString(fmt.Sprintf("**总体置信度**: %.2f\n", result.ConfidenceScore))
	report.WriteString(fmt.Sprintf("**验证问题数**: %d\n\n", len(result.ValidationIssues)))

	// 维度质量摘要
	report.WriteString("## 维度质量摘要\n\n")
	for dimension, quality := range result.DimensionQuality {
		report.WriteString(fmt.Sprintf("- **%s**: %.2f (%s) - 置信度: %.2f\n",
			dimension, quality.Score, quality.Grade, quality.Confidence))
	}
	report.WriteString("\n")

	// 关键问题
	if len(result.ValidationIssues) > 0 {
		report.WriteString("## 发现的问题\n\n")
		for _, issue := range result.ValidationIssues {
			severity := ""
			switch issue.Severity {
			case "critical":
				severity = "🔴 严重"
			case "warning":
				severity = "🟡 警告"
			case "info":
				severity = "🔵 信息"
			}
			
			report.WriteString(fmt.Sprintf("### %s - %s\n", severity, issue.Description))
			report.WriteString(fmt.Sprintf("- **维度**: %s\n", issue.Dimension))
			report.WriteString(fmt.Sprintf("- **影响程度**: %.2f\n", issue.Impact))
			report.WriteString(fmt.Sprintf("- **建议**: %s\n\n", issue.Suggestion))
		}
	}

	return report.String()
}

// generateRecommendedActions 生成推荐行动
func (v *AnalysisQualityValidator) generateRecommendedActions(result *QualityValidationResult) []string {
	actions := []string{}

	// 基于问题生成行动建议
	criticalIssues := 0
	warningIssues := 0
	
	for _, issue := range result.ValidationIssues {
		switch issue.Severity {
		case "critical":
			criticalIssues++
		case "warning":
			warningIssues++
		}
	}

	if criticalIssues > 0 {
		actions = append(actions, fmt.Sprintf("立即处理 %d 个严重问题", criticalIssues))
	}

	if warningIssues > 3 {
		actions = append(actions, fmt.Sprintf("制定计划处理 %d 个警告问题", warningIssues))
	}

	if result.OverallQuality < 0.7 {
		actions = append(actions, "整体质量偏低，建议重新分析或优化分析参数")
	}

	if result.ConfidenceScore < 0.6 {
		actions = append(actions, "分析置信度不足，建议增加更多数据源或调整分析方法")
	}

	// 维度特定建议
	for dimension, quality := range result.DimensionQuality {
		if quality.Score < 0.5 {
			actions = append(actions, fmt.Sprintf("重点改进%s维度的分析质量", dimension))
		}
	}

	if len(actions) == 0 {
		actions = append(actions, "分析质量良好，建议定期验证以保持质量标准")
	}

	return actions
}

// 工具方法
func (v *AnalysisQualityValidator) calculateDimensionScore(values ...float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func (v *AnalysisQualityValidator) getGrade(score float64) string {
	switch {
	case score >= 0.9:
		return "优秀"
	case score >= 0.8:
		return "良好"
	case score >= 0.7:
		return "合格"
	case score >= 0.6:
		return "一般"
	default:
		return "需改进"
	}
}