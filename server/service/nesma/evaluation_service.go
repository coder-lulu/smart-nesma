package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/system"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type EvaluationService struct{}

// 功能点识别规则配置
type FunctionPointRules struct {
	ILFRules []DetectionRule `json:"ilfRules"`
	EIFRules []DetectionRule `json:"eifRules"`
	EIRules  []DetectionRule `json:"eiRules"`
	EORules  []DetectionRule `json:"eoRules"`
	EQRules  []DetectionRule `json:"eqRules"`
}

// 检测规则
type DetectionRule struct {
	RuleName    string   `json:"ruleName"`
	Keywords    []string `json:"keywords"`
	Patterns    []string `json:"patterns"`
	Weight      float64  `json:"weight"`
	Confidence  float64  `json:"confidence"`
	Description string   `json:"description"`
}

// 评估配置
type EvaluationConfig struct {
	NesmaVersion           string  `json:"nesmaVersion"`
	UseAIAssisted          bool    `json:"useAIAssisted"`
	ConfidenceThreshold    float64 `json:"confidenceThreshold"`
	ValidationLevel        string  `json:"validationLevel"`
	IncludeRecommendations bool    `json:"includeRecommendations"`
}

// 复杂度计算结果
type ComplexityCalculationResult struct {
	FunctionType    string  `json:"functionType"`
	DataElements    int     `json:"dataElements"`
	FileTypes       int     `json:"fileTypes"`
	RecordElements  int     `json:"recordElements"`
	ComplexityLevel string  `json:"complexityLevel"`
	WeightFactor    float64 `json:"weightFactor"`
	ConfidenceScore float64 `json:"confidenceScore"`
	DetectionMethod string  `json:"detectionMethod"`
}

// 创建NESMA评估
func (e *EvaluationService) CreateEvaluation(req nesmaReq.CreateEvaluationRequest) (nesmaRes.EvaluationResponse, error) {
	var evaluation nesma.NesmaEvaluation

	// 基本信息设置
	evaluation.ProjectID = req.ProjectID
	evaluation.CycleID = req.CycleID
	evaluation.RequirementVersionID = req.RequirementVersionID
	evaluation.EvaluationName = req.EvaluationName
	evaluation.EvaluationVersion = req.EvaluationVersion
	evaluation.EvaluationType = req.EvaluationType
	evaluation.Status = "pending"
	evaluation.EvaluatorID = req.EvaluatorID
	evaluation.NesmaRules = req.NesmaRules

	// 设置评估配置
	configData, _ := json.Marshal(req.EvaluationConfig)
	evaluation.EvaluationConfig = configData

	now := time.Now()
	evaluation.StartTime = &now

	// 保存评估记录
	if err := global.GVA_DB.Create(&evaluation).Error; err != nil {
		global.GVA_LOG.Error("创建评估记录失败", zap.Error(err))
		return nesmaRes.EvaluationResponse{}, err
	}

	global.GVA_LOG.Info("NESMA评估创建成功", zap.Uint("evaluationId", evaluation.ID))

	return e.buildEvaluationResponse(evaluation), nil
}

// 更新NESMA评估
func (e *EvaluationService) UpdateEvaluation(req nesmaReq.UpdateEvaluationRequest) (nesmaRes.EvaluationResponse, error) {
	// 获取现有评估记录
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.First(&evaluation, req.ID).Error; err != nil {
		global.GVA_LOG.Error("获取评估记录失败", zap.Error(err))
		return nesmaRes.EvaluationResponse{}, err
	}

	// 检查是否可以更新（已完成的评估不能更新基本信息）
	if evaluation.Status == "completed" && req.EvaluationType != "" {
		return nesmaRes.EvaluationResponse{}, fmt.Errorf("已完成的评估不能修改基本信息")
	}

	// 更新基本信息
	if req.EvaluationName != "" {
		evaluation.EvaluationName = req.EvaluationName
	}
	if req.EvaluationVersion != "" {
		evaluation.EvaluationVersion = req.EvaluationVersion
	}
	if req.EvaluationType != "" {
		evaluation.EvaluationType = req.EvaluationType
	}
	if req.Status != "" {
		evaluation.Status = req.Status
	}

	// 更新评估配置
	if req.EvaluationConfig != nil {
		configData, _ := json.Marshal(req.EvaluationConfig)
		evaluation.EvaluationConfig = configData
	}

	// 保存更新
	if err := global.GVA_DB.Save(&evaluation).Error; err != nil {
		global.GVA_LOG.Error("更新评估记录失败", zap.Error(err))
		return nesmaRes.EvaluationResponse{}, err
	}

	global.GVA_LOG.Info("NESMA评估更新成功", zap.Uint("evaluationId", evaluation.ID))

	return e.buildEvaluationResponse(evaluation), nil
}

// 开始自动评估
func (e *EvaluationService) StartAutoEvaluation(evaluationID uint) error {
	// 更新状态为处理中
	if err := global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evaluationID).
		Update("status", "processing").Error; err != nil {
		return err
	}

	// 获取评估记录
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("Project").First(&evaluation, evaluationID).Error; err != nil {
		return err
	}

	global.GVA_LOG.Info("开始AI驱动的NESMA评估",
		zap.Uint("evaluationId", evaluationID),
		zap.Uint("projectId", evaluation.ProjectID))

	// 使用新的AI分析引擎进行评估
	go e.performAIDrivenEvaluation(evaluationID, evaluation)

	return nil
}

// 执行AI驱动的评估
func (e *EvaluationService) performAIDrivenEvaluation(evaluationID uint, evaluation nesma.NesmaEvaluation) {
	global.GVA_LOG.Info("开始AI驱动评估", zap.Uint("evaluationId", evaluationID))

	// 解析评估配置
	var config EvaluationConfig
	if err := json.Unmarshal(evaluation.EvaluationConfig, &config); err != nil {
		global.GVA_LOG.Error("解析评估配置失败", zap.Error(err))
		e.updateEvaluationStatus(evaluationID, "failed", "配置解析失败")
		return
	}

	// 创建AI分析引擎
	aiEngine := NewAIAnalysisEngine()

	// 构建AI分析请求
	analysisRequest := AIAnalysisRequest{
		ProjectID:    evaluation.ProjectID,
		EvaluationID: evaluationID,
		AnalysisType: "comprehensive",
		FocusAreas:   []string{"functional", "technical", "business", "quality", "risk", "compliance"},
		AnalysisConfig: AnalysisConfiguration{
			AnalysisDepth:         "comprehensive",
			AIModel:               "deepseek-chat",
			MaxTokens:             4000,
			Temperature:           0.1,
			EnableBatchProcessing: true,
			BatchSize:             20,
			IncludeHistoricalData: config.ValidationLevel == "comprehensive",
			ComplianceStandards:   []string{"NESMA 2.2"},
		},
		UserID: evaluation.EvaluatorID,
	}

	// 启动AI分析
	aiAnalysis, err := aiEngine.StartAnalysis(analysisRequest)
	if err != nil {
		global.GVA_LOG.Error("启动AI分析失败", zap.Error(err))
		e.updateEvaluationStatus(evaluationID, "failed", "AI分析启动失败: "+err.Error())
		return
	}

	// 更新评估记录，关联AI分析 
	// 重新获取最新的评估记录
	if err := global.GVA_DB.First(&evaluation, evaluationID).Error; err != nil {
		global.GVA_LOG.Error("获取评估记录失败", zap.Error(err))
		return
	}
	
	// 更新字段
	evaluation.AIAnalysisID = &aiAnalysis.ID
	evaluation.Status = "processing"
	evaluation.Progress = 10.0
	
	if err := global.GVA_DB.Save(&evaluation).Error; err != nil {
		global.GVA_LOG.Error("更新评估状态失败", zap.Error(err))
	}

	// 监控AI分析进度并更新评估状态
	go e.monitorAIAnalysisProgress(evaluationID, aiAnalysis.ID)

	global.GVA_LOG.Info("AI分析已启动",
		zap.Uint("evaluationId", evaluationID),
		zap.Uint("aiAnalysisId", aiAnalysis.ID))
}

// 监控AI分析进度
func (e *EvaluationService) monitorAIAnalysisProgress(evaluationID, aiAnalysisID uint) {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次进度
	defer ticker.Stop()

	maxWaitTime := 30 * time.Minute // 最长等待30分钟
	timeout := time.After(maxWaitTime)

	for {
		select {
		case <-ticker.C:
			// 检查AI分析状态
			var aiAnalysis nesma.AIProjectAnalysis
			if err := global.GVA_DB.First(&aiAnalysis, aiAnalysisID).Error; err != nil {
				global.GVA_LOG.Error("获取AI分析状态失败", zap.Error(err))
				continue
			}

			// 更新评估进度
			progress := 10.0 + (aiAnalysis.Progress * 0.8) // 10% + AI进度的80%
			e.updateEvaluationProgress(evaluationID, progress)

			// 检查是否完成
			if aiAnalysis.Status == "completed" {
				e.processAIAnalysisResults(evaluationID, aiAnalysisID)
				return
			} else if aiAnalysis.Status == "failed" {
				e.updateEvaluationStatus(evaluationID, "failed", "AI分析失败")
				return
			}

		case <-timeout:
			global.GVA_LOG.Error("AI分析超时", zap.Uint("evaluationId", evaluationID))
			e.updateEvaluationStatus(evaluationID, "failed", "AI分析超时")
			return
		}
	}
}

// 处理AI分析结果
func (e *EvaluationService) processAIAnalysisResults(evaluationID, aiAnalysisID uint) {
	global.GVA_LOG.Info("开始处理AI分析结果",
		zap.Uint("evaluationId", evaluationID),
		zap.Uint("aiAnalysisId", aiAnalysisID))

	// 获取AI功能分析结果
	var functionalAnalysis nesma.AIFunctionalAnalysis
	if err := global.GVA_DB.Where("analysis_id = ?", aiAnalysisID).First(&functionalAnalysis).Error; err != nil {
		global.GVA_LOG.Error("获取功能分析结果失败", zap.Error(err))
		e.updateEvaluationStatus(evaluationID, "failed", "获取分析结果失败")
		return
	}

	// 获取AI技术分析结果
	var technicalAnalysis nesma.AITechnicalAnalysis
	global.GVA_DB.Where("analysis_id = ?", aiAnalysisID).First(&technicalAnalysis)

	// 获取AI业务分析结果
	var businessAnalysis nesma.AIBusinessAnalysis
	global.GVA_DB.Where("analysis_id = ?", aiAnalysisID).First(&businessAnalysis)

	// 获取AI质量分析结果
	var qualityAnalysis nesma.AIQualityAnalysis
	global.GVA_DB.Where("analysis_id = ?", aiAnalysisID).First(&qualityAnalysis)

	// 获取AI风险分析结果
	var riskAnalysis nesma.AIRiskAnalysis
	global.GVA_DB.Where("analysis_id = ?", aiAnalysisID).First(&riskAnalysis)

	// 计算综合评估结果
	evaluationResult := e.calculateComprehensiveEvaluation(
		functionalAnalysis,
		technicalAnalysis,
		businessAnalysis,
		qualityAnalysis,
		riskAnalysis)

	// 更新评估记录
	now := time.Now()
	updates := map[string]interface{}{
		"status":                    "completed",
		"progress":                  100.0,
		"completion_time":           &now,
		"total_function_points":     functionalAnalysis.TotalFunctionPoints,
		"data_function_points":      functionalAnalysis.DataFunctionPoints,
		"transactional_fp":              functionalAnalysis.TransactionalFunctionPoints,
		"overall_grade":             evaluationResult.OverallGrade,
		"quality_score":             evaluationResult.QualityScore,
		"complexity_level":          evaluationResult.ComplexityLevel,
		"risk_level":                evaluationResult.RiskLevel,
		"evaluation_summary":        evaluationResult.Summary,
		"confidence_score":          evaluationResult.ConfidenceScore,
	}

	if err := global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evaluationID).Updates(updates).Error; err != nil {
		global.GVA_LOG.Error("更新评估结果失败", zap.Error(err))
		return
	}

	global.GVA_LOG.Info("AI驱动的NESMA评估完成",
		zap.Uint("evaluationId", evaluationID),
		zap.Float64("totalFunctionPoints", functionalAnalysis.TotalFunctionPoints),
		zap.String("overallGrade", evaluationResult.OverallGrade))
}

// 计算综合评估结果
func (e *EvaluationService) calculateComprehensiveEvaluation(
	functional nesma.AIFunctionalAnalysis,
	technical nesma.AITechnicalAnalysis,
	business nesma.AIBusinessAnalysis,
	quality nesma.AIQualityAnalysis,
	risk nesma.AIRiskAnalysis) AIEvaluationResult {

	// 计算总体评分 (加权平均)
	functionalWeight := 0.35
	technicalWeight := 0.20
	businessWeight := 0.20
	qualityWeight := 0.15
	riskWeight := 0.10

	overallScore := functional.FunctionalCompleteness*functionalWeight +
		technical.TechnicalFeasibility*technicalWeight +
		business.BusinessValue*businessWeight +
		quality.OverallQuality*qualityWeight +
		(1.0-risk.OverallRisk)*riskWeight

	// 确定总体等级
	var overallGrade string
	switch {
	case overallScore >= 0.9:
		overallGrade = "excellent"
	case overallScore >= 0.8:
		overallGrade = "good"
	case overallScore >= 0.7:
		overallGrade = "satisfactory"
	case overallScore >= 0.6:
		overallGrade = "acceptable"
	default:
		overallGrade = "needs_improvement"
	}

	// 确定复杂度等级
	var complexityLevel string
	if functional.HighComplexityCount > functional.LowComplexityCount {
		complexityLevel = "high"
	} else if functional.MediumComplexityCount > functional.LowComplexityCount {
		complexityLevel = "medium"
	} else {
		complexityLevel = "low"
	}

	// 确定风险等级
	var riskLevel string
	if risk.OverallRisk >= 0.7 {
		riskLevel = "high"
	} else if risk.OverallRisk >= 0.4 {
		riskLevel = "medium"
	} else {
		riskLevel = "low"
	}

	// 生成评估摘要
	summary := fmt.Sprintf("AI驱动的NESMA评估完成。功能点总数: %.1f, 功能完整性: %.2f, "+
		"技术可行性: %.2f, 业务价值: %.2f, 质量评分: %.2f, 风险等级: %s",
		functional.TotalFunctionPoints,
		functional.FunctionalCompleteness,
		technical.TechnicalFeasibility,
		business.BusinessValue,
		quality.OverallQuality,
		riskLevel)

	// 计算综合置信度
	confidenceScore := (functional.ConfidenceScore +
		technical.ConfidenceScore +
		business.ConfidenceScore +
		quality.ConfidenceScore +
		risk.ConfidenceScore) / 5.0

	return AIEvaluationResult{
		OverallGrade:    overallGrade,
		QualityScore:    overallScore,
		ComplexityLevel: complexityLevel,
		RiskLevel:       riskLevel,
		Summary:         summary,
		ConfidenceScore: confidenceScore,
	}
}

// AI评估结果结构
type AIEvaluationResult struct {
	OverallGrade    string  `json:"overallGrade"`
	QualityScore    float64 `json:"qualityScore"`
	ComplexityLevel string  `json:"complexityLevel"`
	RiskLevel       string  `json:"riskLevel"`
	Summary         string  `json:"summary"`
	ConfidenceScore float64 `json:"confidenceScore"`
}

// 更新评估状态
func (e *EvaluationService) updateEvaluationStatus(evaluationID uint, status, message string) {
	updates := map[string]interface{}{
		"status": status,
	}
	if message != "" {
		updates["error_message"] = message
	}
	if status == "failed" {
		now := time.Now()
		updates["completion_time"] = &now
	}

	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evaluationID).Updates(updates)
}

// 更新评估进度
func (e *EvaluationService) updateEvaluationProgress(evaluationID uint, progress float64) {
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evaluationID).
		Update("progress", progress)
}

// 执行自动评估 (保留原有函数以向后兼容)
func (e *EvaluationService) performAutoEvaluation(evaluationID uint, requirements []nesma.NesmaRequirement) {
	var evaluation nesma.NesmaEvaluation
	global.GVA_DB.First(&evaluation, evaluationID)

	// 解析评估配置
	var config EvaluationConfig
	if err := json.Unmarshal(evaluation.EvaluationConfig, &config); err != nil {
		global.GVA_LOG.Error("解析评估配置失败", zap.Error(err))
		return
	}

	// 统计变量
	var totalFP float64
	var dataFP float64
	var transactionalFP float64
	var simpleFunctions, avgFunctions, complexFunctions int

	// 加载功能点识别规则
	rules := e.loadFunctionPointRules()

	// 遍历需求进行功能点识别
	for _, req := range requirements {
		if req.Level == 4 { // 只处理功能点级别的需求
			// 自动识别功能点类型
			result := e.identifyFunctionPoint(evaluationID, req, rules, config)

			if result.FunctionType != "" {
				// 创建功能点记录
				fp := nesma.NesmaFunctionPoint{
					EvaluationID:         evaluationID,
					RequirementID:        &req.ID,
					FunctionType:         result.FunctionType,
					FunctionName:         req.Title,
					FunctionDesc:         req.Description,
					DataElements:         result.DataElements,
					FileTypes:            result.FileTypes,
					RecordElements:       result.RecordElements,
					ComplexityLevel:      result.ComplexityLevel,
					WeightFactor:         result.WeightFactor,
					CalculatedPoints:     result.WeightFactor,
					IdentificationMethod: result.DetectionMethod,
					ConfidenceLevel:      result.ConfidenceScore,
					IsValidated:          result.ConfidenceScore > config.ConfidenceThreshold,
				}

				// 设置检测规则
				ruleData, _ := json.Marshal(result)
				fp.DetectionRules = ruleData

				// 保存功能点
				global.GVA_DB.Create(&fp)

				// 累计统计
				totalFP += fp.CalculatedPoints
				if result.FunctionType == "ILF" || result.FunctionType == "EIF" {
					dataFP += fp.CalculatedPoints
				} else {
					transactionalFP += fp.CalculatedPoints
				}

				// 复杂度统计
				switch result.ComplexityLevel {
				case "Low":
					simpleFunctions++
				case "Average":
					avgFunctions++
				case "High":
					complexFunctions++
				}
			}
		}
	}

	// 计算复杂度指标
	e.calculateComplexityMetrics(evaluationID, requirements)

	// 执行验证检查
	e.performValidationChecks(evaluationID)

	// 计算调整因子
	adjustmentFactor := e.calculateAdjustmentFactor(evaluationID)
	adjustedFP := totalFP * adjustmentFactor

	// 计算质量分数
	confidenceScore := e.calculateConfidenceScore(evaluationID)
	accuracyScore := e.calculateAccuracyScore(evaluationID)
	complianceScore := e.calculateComplianceScore(evaluationID)

	// 更新评估结果
	completionTime := time.Now()
	duration := int(completionTime.Sub(*evaluation.StartTime).Seconds())

	updates := map[string]interface{}{
		"status":                   "completed",
		"total_function_points":    totalFP,
		"data_function_points":     dataFP,
		"transactional_fp":         transactionalFP,
		"adjustment_factor":        adjustmentFactor,
		"adjusted_function_points": adjustedFP,
		"simple_function_count":    simpleFunctions,
		"average_function_count":   avgFunctions,
		"complex_function_count":   complexFunctions,
		"confidence_score":         confidenceScore,
		"accuracy_score":           accuracyScore,
		"compliance_score":         complianceScore,
		"completion_time":          completionTime,
		"duration":                 duration,
	}

	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("id = ?", evaluationID).Updates(updates)

	// 生成改进建议
	if config.IncludeRecommendations {
		recommendations := e.generateRecommendations(evaluationID)
		recData, _ := json.Marshal(recommendations)
		global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
			Where("id = ?", evaluationID).
			Update("recommendation_list", recData)
	}

	global.GVA_LOG.Info("NESMA自动评估完成",
		zap.Uint("evaluationId", evaluationID),
		zap.Float64("totalFP", totalFP),
		zap.Float64("adjustedFP", adjustedFP),
		zap.Float64("confidenceScore", confidenceScore))
}

// 识别功能点类型
func (e *EvaluationService) identifyFunctionPoint(evaluationID uint, req nesma.NesmaRequirement, rules FunctionPointRules, config EvaluationConfig) ComplexityCalculationResult {
	content := strings.ToLower(req.Title + " " + req.Description)
	result := ComplexityCalculationResult{
		DetectionMethod: "auto",
	}

	// 计算各种功能类型的匹配得分
	scores := map[string]float64{
		"ILF": e.calculateRuleScore(content, rules.ILFRules),
		"EIF": e.calculateRuleScore(content, rules.EIFRules),
		"EI":  e.calculateRuleScore(content, rules.EIRules),
		"EO":  e.calculateRuleScore(content, rules.EORules),
		"EQ":  e.calculateRuleScore(content, rules.EQRules),
	}

	// 找到最高分的类型
	maxScore := 0.0
	bestType := ""
	for funcType, score := range scores {
		if score > maxScore {
			maxScore = score
			bestType = funcType
		}
	}

	// 如果置信度不够，使用AI辅助
	if maxScore < config.ConfidenceThreshold && config.UseAIAssisted {
		result = e.aiAssistedIdentification(evaluationID, req, result)
		result.DetectionMethod = "ai_assisted"
	} else if maxScore >= config.ConfidenceThreshold {
		result.FunctionType = bestType
		result.ConfidenceScore = maxScore

		// 估算数据元素等
		result.DataElements = e.estimateDataElements(content, bestType)
		result.FileTypes = e.estimateFileTypes(content, bestType)
		result.RecordElements = e.estimateRecordElements(content, bestType)

		// 计算复杂度级别和权重
		result.ComplexityLevel = e.calculateComplexityLevel(bestType, result.DataElements, result.FileTypes, result.RecordElements)
		result.WeightFactor = e.getWeightFactorFromEvaluation(evaluationID, bestType, result.ComplexityLevel)
	}

	return result
}

// 计算规则匹配分数
func (e *EvaluationService) calculateRuleScore(content string, rules []DetectionRule) float64 {
	totalScore := 0.0
	totalWeight := 0.0

	for _, rule := range rules {
		score := 0.0

		// 关键词匹配
		for _, keyword := range rule.Keywords {
			if strings.Contains(content, strings.ToLower(keyword)) {
				score += rule.Weight
			}
		}

		// 正则模式匹配
		for _, pattern := range rule.Patterns {
			if matched, _ := regexp.MatchString(pattern, content); matched {
				score += rule.Weight * 1.5 // 模式匹配权重更高
			}
		}

		if score > 0 {
			totalScore += score * rule.Confidence
			totalWeight += rule.Weight
		}
	}

	if totalWeight == 0 {
		return 0
	}

	return math.Min(totalScore/totalWeight, 1.0)
}

// AI辅助识别
func (e *EvaluationService) aiAssistedIdentification(evaluationID uint, req nesma.NesmaRequirement, result ComplexityCalculationResult) ComplexityCalculationResult {
	aiService := GetAIService()
	if aiService == nil {
		return result
	}

	prompt := fmt.Sprintf(`
作为国际认证的NESMA功能点评估专家，请基于以下专业背景分析需求并确定其功能点类型：

## 专业背景
你是一位拥有20年以上软件度量经验的NESMA 2.2国际标准专家，具备ISO/IEC 14143标准认证资质。你在功能点识别、复杂度评估、质量分析方面有深厚造诣，在金融、制造、医疗、政府、电商等15+个行业有丰富的项目实践。

## 分析要求
请基于NESMA 2.2标准，运用专业的分析方法，对以下需求进行精确的功能点类型识别：

需求标题：%s
需求描述：%s

## 功能类型定义（基于NESMA 2.2标准）
- **ILF (内部逻辑文件)**: 由应用维护的用户可识别的逻辑相关数据组
- **EIF (外部接口文件)**: 由其他应用维护但本应用引用的用户可识别的逻辑相关数据组
- **EI (外部输入)**: 处理来自应用边界外的数据或控制信息的基本过程
- **EO (外部输出)**: 向应用边界外发送数据或控制信息的基本过程
- **EQ (外部查询)**: 从应用边界外发送输入并接收输出的基本过程

## 分析维度
- **功能特征分析**：数据处理类型、业务流程复杂度、用户交互方式
- **技术实现考量**：接口复杂度、数据存储需求、计算逻辑难度
- **业务价值评估**：用户价值、业务重要性、实现优先级

请以JSON格式回复：
{
  "functionType": "精确识别的功能类型",
  "confidence": 0.0-1.0的置信度,
  "dataElements": 基于DET标准的估算数量,
  "fileTypes": 基于FTR标准的估算数量,
  "recordElements": 基于RET标准的估算数量,
  "reasoning": "详细的分析推理过程",
  "complexityFactors": "复杂度评估因素",
  "qualityScore": 0-100的质量评分
}`, req.Title, req.Description)

	ctx := context.Background()
	config := &AIConfig{
		MaxTokens:   500,
		Temperature: 0.3,
		Model:       "deepseek-chat",
	}

	response, err := aiService.GenerateText(ctx, prompt, config)
	if err != nil {
		global.GVA_LOG.Error("AI辅助识别失败", zap.Error(err))
		return result
	}

	// 解析AI响应
	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		var aiResult map[string]interface{}

		if err := json.Unmarshal([]byte(content), &aiResult); err == nil {
			if funcType, ok := aiResult["functionType"].(string); ok {
				result.FunctionType = funcType
			}
			if confidence, ok := aiResult["confidence"].(float64); ok {
				result.ConfidenceScore = confidence
			}
			if dataElements, ok := aiResult["dataElements"].(float64); ok {
				result.DataElements = int(dataElements)
			}
			if fileTypes, ok := aiResult["fileTypes"].(float64); ok {
				result.FileTypes = int(fileTypes)
			}
			if recordElements, ok := aiResult["recordElements"].(float64); ok {
				result.RecordElements = int(recordElements)
			}

			// 计算复杂度和权重
			result.ComplexityLevel = e.calculateComplexityLevel(result.FunctionType, result.DataElements, result.FileTypes, result.RecordElements)
			result.WeightFactor = e.getWeightFactorFromEvaluation(evaluationID, result.FunctionType, result.ComplexityLevel)
		}
	}

	return result
}

// 估算数据元素数量
func (e *EvaluationService) estimateDataElements(content, functionType string) int {
	// 基于内容中的字段、属性等关键词估算
	keywords := []string{"字段", "属性", "参数", "数据", "信息", "field", "attribute", "data", "property"}
	count := 0

	for _, keyword := range keywords {
		count += strings.Count(content, keyword)
	}

	// 基于功能类型调整基数
	base := map[string]int{
		"ILF": 8, "EIF": 6,
		"EI": 5, "EO": 7,
		"EQ": 4,
	}

	if baseCount, exists := base[functionType]; exists {
		return baseCount + count
	}

	return 5 + count
}

// 估算文件类型数量
func (e *EvaluationService) estimateFileTypes(content, functionType string) int {
	// 基于表、实体等关键词估算
	keywords := []string{"表", "实体", "文件", "数据库", "table", "entity", "file", "database"}
	count := 1 // 至少有一个文件类型

	for _, keyword := range keywords {
		count += strings.Count(content, keyword)
	}

	return count
}

// 估算记录元素数量
func (e *EvaluationService) estimateRecordElements(content, functionType string) int {
	// 基于记录、条目等关键词估算
	keywords := []string{"记录", "条目", "行", "项", "record", "entry", "row", "item"}
	count := 1

	for _, keyword := range keywords {
		count += strings.Count(content, keyword)
	}

	return count
}

// 计算复杂度级别
func (e *EvaluationService) calculateComplexityLevel(functionType string, dataElements, fileTypes, recordElements int) string {
	switch functionType {
	case "ILF", "EIF":
		if recordElements <= 1 && dataElements <= 19 {
			return "Low"
		} else if recordElements <= 5 && dataElements <= 50 {
			return "Average"
		} else {
			return "High"
		}

	case "EI":
		if fileTypes <= 1 && dataElements <= 4 {
			return "Low"
		} else if fileTypes <= 2 && dataElements <= 15 {
			return "Average"
		} else {
			return "High"
		}

	case "EO", "EQ":
		if fileTypes <= 2 && dataElements <= 5 {
			return "Low"
		} else if fileTypes <= 3 && dataElements <= 19 {
			return "Average"
		} else {
			return "High"
		}
	}

	return "Average"
}

// 获取权重因子
func (e *EvaluationService) getWeightFactor(functionType, complexityLevel string) float64 {
	// 这是旧的硬编码方法，应该被 getWeightFactorFromEvaluation 替代
	// 保留作为后备选项
	weights := map[string]map[string]float64{
		"ILF": {"Low": 7, "Average": 10, "High": 15},
		"EIF": {"Low": 5, "Average": 7, "High": 10},
		"EI":  {"Low": 3, "Average": 4, "High": 6},
		"EO":  {"Low": 4, "Average": 5, "High": 7},
		"EQ":  {"Low": 3, "Average": 4, "High": 6},
	}

	if functionWeights, exists := weights[functionType]; exists {
		if weight, exists := functionWeights[complexityLevel]; exists {
			return weight
		}
	}

	return 1.0
}

// getWeightFactorFromEvaluation 从评估因子配置中获取权重因子
func (e *EvaluationService) getWeightFactorFromEvaluation(evaluationId uint, functionType, complexityLevel string) float64 {
	// 获取评估因子配置
	factors, err := e.GetEvaluationFactors(evaluationId)
	if err != nil {
		global.GVA_LOG.Warn("获取评估因子配置失败，使用默认权重", 
			zap.Uint("evaluationId", evaluationId),
			zap.Error(err))
		return e.getWeightFactor(functionType, complexityLevel)
	}

	// 从配置中查找对应的权重
	switch functionType {
	case "ILF":
		for _, weight := range factors.DataFunctions.ILF {
			if strings.ToLower(weight.Complexity) == strings.ToLower(complexityLevel) ||
			   (strings.ToLower(weight.Complexity) == "low" && strings.ToLower(complexityLevel) == "low") ||
			   (strings.ToLower(weight.Complexity) == "average" && strings.ToLower(complexityLevel) == "average") ||
			   (strings.ToLower(weight.Complexity) == "high" && strings.ToLower(complexityLevel) == "high") {
				return float64(weight.Weight)
			}
		}
	case "EIF":
		for _, weight := range factors.DataFunctions.EIF {
			if strings.ToLower(weight.Complexity) == strings.ToLower(complexityLevel) ||
			   (strings.ToLower(weight.Complexity) == "low" && strings.ToLower(complexityLevel) == "low") ||
			   (strings.ToLower(weight.Complexity) == "average" && strings.ToLower(complexityLevel) == "average") ||
			   (strings.ToLower(weight.Complexity) == "high" && strings.ToLower(complexityLevel) == "high") {
				return float64(weight.Weight)
			}
		}
	case "EI":
		for _, weight := range factors.TransactionFunctions.EI {
			if strings.ToLower(weight.Complexity) == strings.ToLower(complexityLevel) ||
			   (strings.ToLower(weight.Complexity) == "low" && strings.ToLower(complexityLevel) == "low") ||
			   (strings.ToLower(weight.Complexity) == "average" && strings.ToLower(complexityLevel) == "average") ||
			   (strings.ToLower(weight.Complexity) == "high" && strings.ToLower(complexityLevel) == "high") {
				return float64(weight.Weight)
			}
		}
	case "EO":
		for _, weight := range factors.TransactionFunctions.EO {
			if strings.ToLower(weight.Complexity) == strings.ToLower(complexityLevel) ||
			   (strings.ToLower(weight.Complexity) == "low" && strings.ToLower(complexityLevel) == "low") ||
			   (strings.ToLower(weight.Complexity) == "average" && strings.ToLower(complexityLevel) == "average") ||
			   (strings.ToLower(weight.Complexity) == "high" && strings.ToLower(complexityLevel) == "high") {
				return float64(weight.Weight)
			}
		}
	case "EQ":
		for _, weight := range factors.TransactionFunctions.EQ {
			if strings.ToLower(weight.Complexity) == strings.ToLower(complexityLevel) ||
			   (strings.ToLower(weight.Complexity) == "low" && strings.ToLower(complexityLevel) == "low") ||
			   (strings.ToLower(weight.Complexity) == "average" && strings.ToLower(complexityLevel) == "average") ||
			   (strings.ToLower(weight.Complexity) == "high" && strings.ToLower(complexityLevel) == "high") {
				return float64(weight.Weight)
			}
		}
	}

	// 如果没有找到配置，使用默认权重
	global.GVA_LOG.Warn("未找到对应的权重配置，使用默认权重", 
		zap.Uint("evaluationId", evaluationId),
		zap.String("functionType", functionType),
		zap.String("complexityLevel", complexityLevel))
	return e.getWeightFactor(functionType, complexityLevel)
}

// 加载功能点识别规则
func (e *EvaluationService) loadFunctionPointRules() FunctionPointRules {
	return FunctionPointRules{
		ILFRules: []DetectionRule{
			{
				RuleName:    "数据存储识别",
				Keywords:    []string{"存储", "保存", "数据库", "表", "维护", "管理"},
				Patterns:    []string{`.*存储.*数据.*`, `.*维护.*信息.*`},
				Weight:      1.0,
				Confidence:  0.8,
				Description: "识别内部逻辑文件特征",
			},
			{
				RuleName:    "实体管理识别",
				Keywords:    []string{"用户", "客户", "产品", "订单", "实体"},
				Patterns:    []string{`.*管理.*`, `.*维护.*`},
				Weight:      0.9,
				Confidence:  0.7,
				Description: "识别实体管理功能",
			},
		},
		EIFRules: []DetectionRule{
			{
				RuleName:    "外部数据源",
				Keywords:    []string{"外部", "接口", "第三方", "API", "调用"},
				Patterns:    []string{`.*外部.*数据.*`, `.*第三方.*接口.*`},
				Weight:      1.0,
				Confidence:  0.8,
				Description: "识别外部接口文件",
			},
		},
		EIRules: []DetectionRule{
			{
				RuleName:    "输入处理",
				Keywords:    []string{"输入", "录入", "添加", "创建", "新增", "提交"},
				Patterns:    []string{`.*输入.*`, `.*录入.*`, `.*添加.*`},
				Weight:      1.0,
				Confidence:  0.9,
				Description: "识别外部输入功能",
			},
		},
		EORules: []DetectionRule{
			{
				RuleName:    "输出生成",
				Keywords:    []string{"输出", "生成", "报表", "导出", "打印", "计算"},
				Patterns:    []string{`.*生成.*报表.*`, `.*导出.*`, `.*计算.*`},
				Weight:      1.0,
				Confidence:  0.8,
				Description: "识别外部输出功能",
			},
		},
		EQRules: []DetectionRule{
			{
				RuleName:    "查询检索",
				Keywords:    []string{"查询", "检索", "搜索", "浏览", "列表", "显示"},
				Patterns:    []string{`.*查询.*`, `.*检索.*`, `.*列表.*`},
				Weight:      1.0,
				Confidence:  0.9,
				Description: "识别外部查询功能",
			},
		},
	}
}

// 构建评估响应
func (e *EvaluationService) buildEvaluationResponse(evaluation nesma.NesmaEvaluation) nesmaRes.EvaluationResponse {
	response := nesmaRes.EvaluationResponse{
		ID:                     evaluation.ID,
		ProjectID:              evaluation.ProjectID,
		EvaluationName:         evaluation.EvaluationName,
		EvaluationVersion:      evaluation.EvaluationVersion,
		EvaluationType:         evaluation.EvaluationType,
		Status:                 evaluation.Status,
		TotalFunctionPoints:    evaluation.TotalFunctionPoints,
		DataFunctionPoints:     evaluation.DataFunctionPoints,
		TransactionalFP:        evaluation.TransactionalFP,
		AdjustedFunctionPoints: evaluation.AdjustedFunctionPoints,
		ConfidenceScore:        evaluation.ConfidenceScore,
		AccuracyScore:          evaluation.AccuracyScore,
		ComplianceScore:        evaluation.ComplianceScore,
		EvaluatorID:            evaluation.EvaluatorID,
		StartTime:              evaluation.StartTime,
		CompletionTime:         evaluation.CompletionTime,
		Duration:               evaluation.Duration,
		CreatedAt:              evaluation.CreatedAt,
		UpdatedAt:              evaluation.UpdatedAt,
	}

	// 填充关联字段
	if evaluation.Project.Name != "" {
		response.ProjectName = evaluation.Project.Name
	}

	if evaluation.Cycle != nil {
		response.CycleName = evaluation.Cycle.Name
	}

	if evaluation.RequirementVersion != nil {
		response.RequirementVersionName = evaluation.RequirementVersion.Version
	}

	// 获取评估人姓名
	if evaluation.EvaluatorID > 0 {
		var evaluator struct {
			NickName string `json:"nickName"`
		}
		if err := global.GVA_DB.Model(&system.SysUser{}).
			Select("nick_name").
			Where("id = ?", evaluation.EvaluatorID).
			First(&evaluator).Error; err == nil {
			response.EvaluatorName = evaluator.NickName
		}
	}

	// 填充评估详情页面需要的额外字段
	response.NesmaRules = "NESMA 2.1"  // 默认NESMA规则版本
	response.UnadjustedFunctionPoints = evaluation.TotalFunctionPoints / evaluation.AdjustmentFactor
	if evaluation.AdjustmentFactor == 0 {
		response.UnadjustedFunctionPoints = evaluation.TotalFunctionPoints
	}
	
	// 计算功能点条目数
	var functionPointCount int64
	global.GVA_DB.Model(&nesma.NesmaFunctionPoint{}).
		Where("evaluation_id = ?", evaluation.ID).
		Count(&functionPointCount)
	response.FunctionPointCount = int(functionPointCount)

	return response
}

// 计算复杂度指标
func (e *EvaluationService) calculateComplexityMetrics(evaluationID uint, requirements []nesma.NesmaRequirement) {
	// 获取评估记录
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("FunctionPoints").First(&evaluation, evaluationID).Error; err != nil {
		global.GVA_LOG.Error("获取评估记录失败", zap.Error(err))
		return
	}

	global.GVA_LOG.Info("开始计算复杂度指标", zap.Uint("evaluationId", evaluationID))

	// 1. 计算功能点分布指标
	e.calculateFunctionPointDistribution(evaluationID, evaluation.FunctionPoints)

	// 2. 计算复杂度平衡指标
	e.calculateComplexityBalance(evaluationID, evaluation.FunctionPoints)

	// 3. 计算质量指标
	e.calculateQualityMetrics(evaluationID, evaluation.FunctionPoints)

	// 4. 计算项目复杂度指标
	e.calculateProjectComplexityMetrics(evaluationID, requirements)

	// 5. 计算估算准确度指标
	e.calculateEstimationAccuracy(evaluationID, evaluation.FunctionPoints)

	global.GVA_LOG.Info("复杂度指标计算完成", zap.Uint("evaluationId", evaluationID))
}

// 计算功能点分布指标
func (e *EvaluationService) calculateFunctionPointDistribution(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	// 统计各类型功能点数量和比例
	typeCounts := make(map[string]int)
	complexityCounts := make(map[string]int)
	totalFP := 0.0

	for _, fp := range functionPoints {
		typeCounts[fp.FunctionType]++
		complexityCounts[fp.ComplexityLevel]++
		totalFP += fp.CalculatedPoints
	}

	// 数据功能点比例
	dataFP := float64(typeCounts["ILF"]+typeCounts["EIF"]) / float64(len(functionPoints))
	transactionalFP := float64(typeCounts["EI"]+typeCounts["EO"]+typeCounts["EQ"]) / float64(len(functionPoints))

	// 创建分布指标
	metrics := []nesma.NesmaComplexityMetric{
		{
			EvaluationID:       evaluationID,
			MetricType:         "DISTRIBUTION",
			MetricName:         "数据功能点比例",
			MetricValue:        dataFP,
			ThresholdValue:     0.3, // 建议数据功能点占比30%以上
			Score:              e.calculateDistributionScore(dataFP, 0.3),
			Weight:             1.0,
			QualityIndicator:   e.getQualityIndicator(dataFP, 0.3),
			ImpactLevel:        "medium",
			CalculationFormula: "(ILF + EIF) / 总功能点数",
		},
		{
			EvaluationID:       evaluationID,
			MetricType:         "DISTRIBUTION",
			MetricName:         "事务功能点比例",
			MetricValue:        transactionalFP,
			ThresholdValue:     0.7, // 建议事务功能点占比70%以下
			Score:              e.calculateDistributionScore(transactionalFP, 0.7),
			Weight:             1.0,
			QualityIndicator:   e.getQualityIndicator(transactionalFP, 0.7),
			ImpactLevel:        "medium",
			CalculationFormula: "(EI + EO + EQ) / 总功能点数",
		},
	}

	// 复杂度分布指标
	for level, count := range complexityCounts {
		ratio := float64(count) / float64(len(functionPoints))
		metrics = append(metrics, nesma.NesmaComplexityMetric{
			EvaluationID:       evaluationID,
			MetricType:         "COMPLEXITY_DISTRIBUTION",
			MetricName:         fmt.Sprintf("%s复杂度功能点比例", level),
			MetricValue:        ratio,
			ThresholdValue:     0.33, // 理想状态下各复杂度级别应该相对平衡
			Score:              e.calculateDistributionScore(ratio, 0.33),
			Weight:             0.8,
			QualityIndicator:   e.getComplexityDistributionQuality(level, ratio),
			ImpactLevel:        "low",
			CalculationFormula: fmt.Sprintf("%s复杂度功能点数 / 总功能点数", level),
		})
	}

	// 批量保存指标
	for _, metric := range metrics {
		global.GVA_DB.Create(&metric)
	}
}

// 计算复杂度平衡指标
func (e *EvaluationService) calculateComplexityBalance(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	// 计算各类型功能点的平均复杂度
	typeComplexity := make(map[string][]float64)

	for _, fp := range functionPoints {
		typeComplexity[fp.FunctionType] = append(typeComplexity[fp.FunctionType], fp.WeightFactor)
	}

	var balanceScore float64
	for funcType, weights := range typeComplexity {
		avgWeight := e.calculateAverage(weights)
		variance := e.calculateVariance(weights, avgWeight)

		// 创建平衡指标
		metric := nesma.NesmaComplexityMetric{
			EvaluationID:       evaluationID,
			MetricType:         "BALANCE",
			MetricName:         fmt.Sprintf("%s复杂度平衡度", funcType),
			MetricValue:        variance,
			ThresholdValue:     2.0,                         // 方差阈值
			Score:              math.Max(0, 1-variance/4.0), // 方差越小分数越高
			Weight:             1.0,
			QualityIndicator:   e.getBalanceQuality(variance),
			ImpactLevel:        "medium",
			CalculationFormula: "权重因子方差计算",
		}

		balanceScore += metric.Score
		global.GVA_DB.Create(&metric)
	}

	// 总体平衡指标
	overallBalance := nesma.NesmaComplexityMetric{
		EvaluationID:       evaluationID,
		MetricType:         "BALANCE",
		MetricName:         "整体复杂度平衡度",
		MetricValue:        balanceScore / float64(len(typeComplexity)),
		ThresholdValue:     0.7,
		Score:              balanceScore / float64(len(typeComplexity)),
		Weight:             1.5,
		QualityIndicator:   e.getQualityIndicator(balanceScore/float64(len(typeComplexity)), 0.7),
		ImpactLevel:        "high",
		CalculationFormula: "各类型复杂度平衡度平均值",
	}

	global.GVA_DB.Create(&overallBalance)
}

// 计算质量指标
func (e *EvaluationService) calculateQualityMetrics(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	// 计算置信度分布
	var totalConfidence float64
	var lowConfidenceCount int
	var highConfidenceCount int

	for _, fp := range functionPoints {
		totalConfidence += fp.ConfidenceLevel
		if fp.ConfidenceLevel < 0.6 {
			lowConfidenceCount++
		} else if fp.ConfidenceLevel > 0.8 {
			highConfidenceCount++
		}
	}

	avgConfidence := totalConfidence / float64(len(functionPoints))
	lowConfidenceRatio := float64(lowConfidenceCount) / float64(len(functionPoints))
	highConfidenceRatio := float64(highConfidenceCount) / float64(len(functionPoints))

	// 创建质量指标
	qualityMetrics := []nesma.NesmaComplexityMetric{
		{
			EvaluationID:       evaluationID,
			MetricType:         "QUALITY",
			MetricName:         "平均置信度",
			MetricValue:        avgConfidence,
			ThresholdValue:     0.75,
			Score:              avgConfidence,
			Weight:             2.0,
			QualityIndicator:   e.getQualityIndicator(avgConfidence, 0.75),
			ImpactLevel:        "critical",
			CalculationFormula: "所有功能点置信度平均值",
		},
		{
			EvaluationID:       evaluationID,
			MetricType:         "QUALITY",
			MetricName:         "低置信度功能点比例",
			MetricValue:        lowConfidenceRatio,
			ThresholdValue:     0.2, // 低置信度功能点应少于20%
			Score:              math.Max(0, 1-lowConfidenceRatio/0.2),
			Weight:             1.5,
			QualityIndicator:   e.getQualityIndicator(1-lowConfidenceRatio, 0.8),
			ImpactLevel:        "high",
			CalculationFormula: "置信度<0.6的功能点数 / 总功能点数",
		},
		{
			EvaluationID:       evaluationID,
			MetricType:         "QUALITY",
			MetricName:         "高置信度功能点比例",
			MetricValue:        highConfidenceRatio,
			ThresholdValue:     0.5, // 高置信度功能点应超过50%
			Score:              math.Min(1, highConfidenceRatio/0.5),
			Weight:             1.2,
			QualityIndicator:   e.getQualityIndicator(highConfidenceRatio, 0.5),
			ImpactLevel:        "medium",
			CalculationFormula: "置信度>0.8的功能点数 / 总功能点数",
		},
	}

	// 批量保存质量指标
	for _, metric := range qualityMetrics {
		global.GVA_DB.Create(&metric)
	}
}

// 计算项目复杂度指标
func (e *EvaluationService) calculateProjectComplexityMetrics(evaluationID uint, requirements []nesma.NesmaRequirement) {
	// 需求层次分析
	levelCounts := make(map[int]int)
	for _, req := range requirements {
		levelCounts[req.Level]++
	}

	// 计算需求复杂度
	reqComplexity := float64(levelCounts[3]*1+levelCounts[4]*2+levelCounts[5]*3) / float64(len(requirements))

	// 创建项目复杂度指标
	projectMetrics := []nesma.NesmaComplexityMetric{
		{
			EvaluationID:       evaluationID,
			MetricType:         "PROJECT_COMPLEXITY",
			MetricName:         "需求层次复杂度",
			MetricValue:        reqComplexity,
			ThresholdValue:     2.0,
			Score:              math.Min(1, reqComplexity/2.0),
			Weight:             1.0,
			QualityIndicator:   e.getComplexityQuality(reqComplexity),
			ImpactLevel:        "medium",
			CalculationFormula: "加权需求层次复杂度计算",
		},
		{
			EvaluationID:       evaluationID,
			MetricType:         "PROJECT_COMPLEXITY",
			MetricName:         "需求总数",
			MetricValue:        float64(len(requirements)),
			ThresholdValue:     100,
			Score:              math.Min(1, float64(len(requirements))/100),
			Weight:             0.8,
			QualityIndicator:   e.getComplexityQuality(float64(len(requirements)) / 50),
			ImpactLevel:        "low",
			CalculationFormula: "项目总需求数量",
		},
	}

	// 批量保存项目复杂度指标
	for _, metric := range projectMetrics {
		global.GVA_DB.Create(&metric)
	}
}

// 计算估算准确度指标
func (e *EvaluationService) calculateEstimationAccuracy(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	// 计算AI辅助识别的准确度
	var aiAssistedCount int
	var totalAIConfidence float64

	for _, fp := range functionPoints {
		if fp.IdentificationMethod == "ai_assisted" {
			aiAssistedCount++
			totalAIConfidence += fp.ConfidenceLevel
		}
	}

	var aiAccuracy float64
	if aiAssistedCount > 0 {
		aiAccuracy = totalAIConfidence / float64(aiAssistedCount)
	}

	// 创建准确度指标
	accuracyMetric := nesma.NesmaComplexityMetric{
		EvaluationID:       evaluationID,
		MetricType:         "ACCURACY",
		MetricName:         "AI辅助识别准确度",
		MetricValue:        aiAccuracy,
		ThresholdValue:     0.8,
		Score:              aiAccuracy,
		Weight:             1.0,
		QualityIndicator:   e.getQualityIndicator(aiAccuracy, 0.8),
		ImpactLevel:        "medium",
		CalculationFormula: "AI辅助识别功能点平均置信度",
	}

	global.GVA_DB.Create(&accuracyMetric)
}

// 辅助计算函数
func (e *EvaluationService) calculateAverage(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func (e *EvaluationService) calculateVariance(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += (v - mean) * (v - mean)
	}
	return sum / float64(len(values))
}

func (e *EvaluationService) calculateDistributionScore(value, threshold float64) float64 {
	if value >= threshold {
		return 1.0
	}
	return value / threshold
}

func (e *EvaluationService) getQualityIndicator(value, threshold float64) string {
	ratio := value / threshold
	if ratio >= 1.0 {
		return "excellent"
	} else if ratio >= 0.8 {
		return "good"
	} else if ratio >= 0.6 {
		return "average"
	} else {
		return "poor"
	}
}

func (e *EvaluationService) getComplexityDistributionQuality(level string, ratio float64) string {
	// 理想的复杂度分布应该是平衡的
	ideal := 0.33
	deviation := math.Abs(ratio - ideal)

	if deviation <= 0.1 {
		return "excellent"
	} else if deviation <= 0.2 {
		return "good"
	} else if deviation <= 0.3 {
		return "average"
	} else {
		return "poor"
	}
}

func (e *EvaluationService) getBalanceQuality(variance float64) string {
	if variance <= 1.0 {
		return "excellent"
	} else if variance <= 2.0 {
		return "good"
	} else if variance <= 3.0 {
		return "average"
	} else {
		return "poor"
	}
}

func (e *EvaluationService) getComplexityQuality(complexity float64) string {
	if complexity <= 1.0 {
		return "excellent"
	} else if complexity <= 2.0 {
		return "good"
	} else if complexity <= 3.0 {
		return "average"
	} else {
		return "poor"
	}
}

// 执行验证检查
func (e *EvaluationService) performValidationChecks(evaluationID uint) {
	global.GVA_LOG.Info("开始执行NESMA标准合规检查", zap.Uint("evaluationId", evaluationID))

	// 获取评估数据
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("FunctionPoints").First(&evaluation, evaluationID).Error; err != nil {
		global.GVA_LOG.Error("获取评估数据失败", zap.Error(err))
		return
	}

	// 1. 功能点分类验证
	e.validateFunctionPointClassification(evaluationID, evaluation.FunctionPoints)

	// 2. 复杂度级别验证
	e.validateComplexityLevels(evaluationID, evaluation.FunctionPoints)

	// 3. 权重因子验证
	e.validateWeightFactors(evaluationID, evaluation.FunctionPoints)

	// 4. 数据一致性验证
	e.validateDataConsistency(evaluationID, evaluation)

	// 5. NESMA标准合规验证
	e.validateNesmaCompliance(evaluationID, evaluation)

	global.GVA_LOG.Info("NESMA标准合规检查完成", zap.Uint("evaluationId", evaluationID))
}

// 验证功能点分类
func (e *EvaluationService) validateFunctionPointClassification(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	validationItems := []nesma.NesmaValidationItem{}

	// 检查功能点类型分布是否合理
	typeCounts := make(map[string]int)
	for _, fp := range functionPoints {
		typeCounts[fp.FunctionType]++
	}

	totalFP := len(functionPoints)
	if totalFP == 0 {
		return
	}

	// 验证数据功能点比例
	dataFP := typeCounts["ILF"] + typeCounts["EIF"]
	dataRatio := float64(dataFP) / float64(totalFP)

	validation := nesma.NesmaValidationItem{
		EvaluationID:     evaluationID,
		ValidationRule:   "数据功能点比例检查",
		RuleDescription:  "数据功能点(ILF+EIF)应占总功能点的20%-40%",
		ExpectedValue:    "0.2-0.4",
		ActualValue:      fmt.Sprintf("%.2f", dataRatio),
		ValidationNotes:  "数据功能点比例影响系统架构复杂度",
		IsCritical:       false,
		ResolutionStatus: "open",
	}

	if dataRatio < 0.2 {
		validation.ValidationResult = "warning"
		validation.DeviationLevel = "medium"
		validation.ValidationNotes = "数据功能点比例偏低，可能缺少数据存储需求"
	} else if dataRatio > 0.4 {
		validation.ValidationResult = "warning"
		validation.DeviationLevel = "medium"
		validation.ValidationNotes = "数据功能点比例偏高，可能存在数据冗余"
	} else {
		validation.ValidationResult = "pass"
		validation.DeviationLevel = "none"
	}

	validationItems = append(validationItems, validation)

	// 验证事务功能点分布
	transactionTypes := []string{"EI", "EO", "EQ"}
	for _, transType := range transactionTypes {
		count := typeCounts[transType]
		ratio := float64(count) / float64(totalFP)

		validation := nesma.NesmaValidationItem{
			EvaluationID:     evaluationID,
			ValidationRule:   fmt.Sprintf("%s功能点比例检查", transType),
			RuleDescription:  fmt.Sprintf("%s功能点比例应在合理范围内", transType),
			ExpectedValue:    "0.1-0.5",
			ActualValue:      fmt.Sprintf("%.2f", ratio),
			ValidationNotes:  fmt.Sprintf("%s功能点分布影响系统交互复杂度", transType),
			IsCritical:       false,
			ResolutionStatus: "open",
		}

		if ratio < 0.05 {
			validation.ValidationResult = "warning"
			validation.DeviationLevel = "low"
			validation.ValidationNotes = fmt.Sprintf("%s功能点过少，可能遗漏相关需求", transType)
		} else if ratio > 0.6 {
			validation.ValidationResult = "fail"
			validation.DeviationLevel = "high"
			validation.ValidationNotes = fmt.Sprintf("%s功能点过多，需要重新审查分类", transType)
			validation.IsCritical = true
		} else {
			validation.ValidationResult = "pass"
			validation.DeviationLevel = "none"
		}

		validationItems = append(validationItems, validation)
	}

	// 批量保存验证结果
	for _, item := range validationItems {
		global.GVA_DB.Create(&item)
	}
}

// 验证复杂度级别
func (e *EvaluationService) validateComplexityLevels(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	validationItems := []nesma.NesmaValidationItem{}

	// 统计各复杂度级别分布
	complexityCounts := make(map[string]int)
	for _, fp := range functionPoints {
		complexityCounts[fp.ComplexityLevel]++
	}

	totalFP := len(functionPoints)
	if totalFP == 0 {
		return
	}

	// 检查复杂度级别分布是否合理
	for level, count := range complexityCounts {
		ratio := float64(count) / float64(totalFP)

		validation := nesma.NesmaValidationItem{
			EvaluationID:     evaluationID,
			ValidationRule:   fmt.Sprintf("%s复杂度级别分布检查", level),
			RuleDescription:  "各复杂度级别应该相对平衡，避免过度集中",
			ExpectedValue:    "0.2-0.5",
			ActualValue:      fmt.Sprintf("%.2f", ratio),
			ValidationNotes:  "复杂度级别分布影响开发工作量评估",
			IsCritical:       false,
			ResolutionStatus: "open",
		}

		if ratio > 0.7 {
			validation.ValidationResult = "warning"
			validation.DeviationLevel = "medium"
			validation.ValidationNotes = fmt.Sprintf("%s复杂度功能点过于集中，建议重新评估", level)
		} else if ratio < 0.1 && count > 0 {
			validation.ValidationResult = "pass"
			validation.DeviationLevel = "none"
		} else {
			validation.ValidationResult = "pass"
			validation.DeviationLevel = "none"
		}

		validationItems = append(validationItems, validation)
	}

	// 验证复杂度级别与权重因子的一致性
	for _, fp := range functionPoints {
		expectedWeight := e.getWeightFactorFromEvaluation(evaluationID, fp.FunctionType, fp.ComplexityLevel)

		if math.Abs(fp.WeightFactor-expectedWeight) > 0.1 {
			validation := nesma.NesmaValidationItem{
				EvaluationID:     evaluationID,
				ValidationRule:   "权重因子一致性检查",
				RuleDescription:  "权重因子应与复杂度级别匹配",
				ExpectedValue:    fmt.Sprintf("%.1f", expectedWeight),
				ActualValue:      fmt.Sprintf("%.1f", fp.WeightFactor),
				ValidationResult: "fail",
				DeviationLevel:   "high",
				ValidationNotes:  fmt.Sprintf("功能点ID:%d的权重因子与复杂度级别不匹配", fp.ID),
				IsCritical:       true,
				ResolutionStatus: "open",
			}
			validationItems = append(validationItems, validation)
		}
	}

	// 批量保存验证结果
	for _, item := range validationItems {
		global.GVA_DB.Create(&item)
	}
}

// 验证权重因子
func (e *EvaluationService) validateWeightFactors(evaluationID uint, functionPoints []nesma.NesmaFunctionPoint) {
	validationItems := []nesma.NesmaValidationItem{}

	// NESMA标准权重因子表
	standardWeights := map[string]map[string]float64{
		"ILF": {"Low": 7, "Average": 10, "High": 15},
		"EIF": {"Low": 5, "Average": 7, "High": 10},
		"EI":  {"Low": 3, "Average": 4, "High": 6},
		"EO":  {"Low": 4, "Average": 5, "High": 7},
		"EQ":  {"Low": 3, "Average": 4, "High": 6},
	}

	// 检查每个功能点的权重因子
	for _, fp := range functionPoints {
		if typeWeights, exists := standardWeights[fp.FunctionType]; exists {
			if expectedWeight, exists := typeWeights[fp.ComplexityLevel]; exists {

				validation := nesma.NesmaValidationItem{
					EvaluationID:     evaluationID,
					ValidationRule:   "标准权重因子检查",
					RuleDescription:  "权重因子必须符合NESMA标准",
					ExpectedValue:    fmt.Sprintf("%.0f", expectedWeight),
					ActualValue:      fmt.Sprintf("%.1f", fp.WeightFactor),
					ValidationNotes:  fmt.Sprintf("功能点:%s (%s-%s)", fp.FunctionName, fp.FunctionType, fp.ComplexityLevel),
					IsCritical:       true,
					ResolutionStatus: "open",
				}

				if fp.WeightFactor == expectedWeight {
					validation.ValidationResult = "pass"
					validation.DeviationLevel = "none"
				} else {
					validation.ValidationResult = "fail"
					validation.DeviationLevel = "high"
					validation.ValidationNotes += fmt.Sprintf(" - 权重因子不符合标准")
				}

				validationItems = append(validationItems, validation)
			}
		}
	}

	// 批量保存验证结果
	for _, item := range validationItems {
		global.GVA_DB.Create(&item)
	}
}

// 验证数据一致性
func (e *EvaluationService) validateDataConsistency(evaluationID uint, evaluation nesma.NesmaEvaluation) {
	validationItems := []nesma.NesmaValidationItem{}

	// 验证总功能点数计算
	var calculatedTotal float64
	for _, fp := range evaluation.FunctionPoints {
		calculatedTotal += fp.CalculatedPoints
	}

	if math.Abs(calculatedTotal-evaluation.TotalFunctionPoints) > 0.1 {
		validation := nesma.NesmaValidationItem{
			EvaluationID:     evaluationID,
			ValidationRule:   "总功能点数一致性检查",
			RuleDescription:  "总功能点数应等于各功能点之和",
			ExpectedValue:    fmt.Sprintf("%.1f", calculatedTotal),
			ActualValue:      fmt.Sprintf("%.1f", evaluation.TotalFunctionPoints),
			ValidationResult: "fail",
			DeviationLevel:   "high",
			ValidationNotes:  "总功能点数计算不一致",
			IsCritical:       true,
			ResolutionStatus: "open",
		}
		validationItems = append(validationItems, validation)
	}

	// 验证数据功能点与事务功能点分类
	var actualDataFP, actualTransFP float64
	for _, fp := range evaluation.FunctionPoints {
		if fp.FunctionType == "ILF" || fp.FunctionType == "EIF" {
			actualDataFP += fp.CalculatedPoints
		} else {
			actualTransFP += fp.CalculatedPoints
		}
	}

	if math.Abs(actualDataFP-evaluation.DataFunctionPoints) > 0.1 {
		validation := nesma.NesmaValidationItem{
			EvaluationID:     evaluationID,
			ValidationRule:   "数据功能点分类一致性检查",
			RuleDescription:  "数据功能点数应等于ILF和EIF功能点之和",
			ExpectedValue:    fmt.Sprintf("%.1f", actualDataFP),
			ActualValue:      fmt.Sprintf("%.1f", evaluation.DataFunctionPoints),
			ValidationResult: "fail",
			DeviationLevel:   "medium",
			ValidationNotes:  "数据功能点分类计算不一致",
			IsCritical:       false,
			ResolutionStatus: "open",
		}
		validationItems = append(validationItems, validation)
	}

	// 批量保存验证结果
	for _, item := range validationItems {
		global.GVA_DB.Create(&item)
	}
}

// 验证NESMA标准合规
func (e *EvaluationService) validateNesmaCompliance(evaluationID uint, evaluation nesma.NesmaEvaluation) {
	validationItems := []nesma.NesmaValidationItem{}

	// 检查是否包含必需的功能点类型
	typeExists := make(map[string]bool)
	for _, fp := range evaluation.FunctionPoints {
		typeExists[fp.FunctionType] = true
	}

	requiredTypes := []string{"ILF", "EI"} // 最基本的功能点类型
	for _, reqType := range requiredTypes {
		validation := nesma.NesmaValidationItem{
			EvaluationID:     evaluationID,
			ValidationRule:   fmt.Sprintf("%s功能点存在性检查", reqType),
			RuleDescription:  fmt.Sprintf("项目应包含%s类型的功能点", reqType),
			ExpectedValue:    "存在",
			ValidationNotes:  "基本功能点类型检查",
			IsCritical:       false,
			ResolutionStatus: "open",
		}

		if typeExists[reqType] {
			validation.ValidationResult = "pass"
			validation.ActualValue = "存在"
			validation.DeviationLevel = "none"
		} else {
			validation.ValidationResult = "warning"
			validation.ActualValue = "不存在"
			validation.DeviationLevel = "medium"
			validation.ValidationNotes = fmt.Sprintf("项目缺少%s类型功能点，可能需要补充相关需求", reqType)
		}

		validationItems = append(validationItems, validation)
	}

	// 检查项目规模合理性
	totalFP := evaluation.TotalFunctionPoints
	validation := nesma.NesmaValidationItem{
		EvaluationID:     evaluationID,
		ValidationRule:   "项目规模合理性检查",
		RuleDescription:  "项目功能点数应在合理范围内",
		ExpectedValue:    ">10",
		ActualValue:      fmt.Sprintf("%.1f", totalFP),
		ValidationNotes:  "项目规模影响开发策略制定",
		IsCritical:       false,
		ResolutionStatus: "open",
	}

	if totalFP < 10 {
		validation.ValidationResult = "warning"
		validation.DeviationLevel = "low"
		validation.ValidationNotes = "项目规模较小，建议确认是否遗漏需求"
	} else if totalFP > 1000 {
		validation.ValidationResult = "warning"
		validation.DeviationLevel = "medium"
		validation.ValidationNotes = "项目规模较大，建议分阶段实施"
	} else {
		validation.ValidationResult = "pass"
		validation.DeviationLevel = "none"
	}

	validationItems = append(validationItems, validation)

	// 批量保存验证结果
	for _, item := range validationItems {
		global.GVA_DB.Create(&item)
	}
}

// 计算调整因子
func (e *EvaluationService) calculateAdjustmentFactor(evaluationID uint) float64 {
	// 获取评估因子配置
	factors, err := e.GetEvaluationFactors(evaluationID)
	if err != nil {
		global.GVA_LOG.Error("获取评估因子配置失败，使用默认计算方法", zap.Error(err))
		return e.calculateLegacyAdjustmentFactor(evaluationID)
	}

	// 使用NESMA标准的14个调整因子计算VAF
	totalDegreeOfInfluence := 0
	for _, factor := range factors.AdjustmentFactors {
		// 确保值在有效范围内 (0-5)
		value := factor.Value
		if value < 0 {
			value = 0
		} else if value > 5 {
			value = 5
		}
		totalDegreeOfInfluence += value
	}

	// NESMA标准公式：VAF = (TDI * 0.01) + 0.65
	// 其中TDI是所有调整因子的总和
	vaf := (float64(totalDegreeOfInfluence) * 0.01) + 0.65

	global.GVA_LOG.Info("计算NESMA调整因子", 
		zap.Uint("evaluationId", evaluationID),
		zap.Int("totalDegreeOfInfluence", totalDegreeOfInfluence),
		zap.Float64("vaf", vaf))

	return vaf
}

// calculateLegacyAdjustmentFactor 原有的调整因子计算方法（作为后备）
func (e *EvaluationService) calculateLegacyAdjustmentFactor(evaluationID uint) float64 {
	// 获取评估数据
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("FunctionPoints").First(&evaluation, evaluationID).Error; err != nil {
		global.GVA_LOG.Error("获取评估数据失败", zap.Error(err))
		return 1.0
	}

	// 基础调整因子
	adjustmentFactor := 1.0

	// 1. 项目复杂度调整
	totalFP := evaluation.TotalFunctionPoints
	if totalFP > 500 {
		adjustmentFactor += 0.1 // 大型项目复杂度调整
	} else if totalFP > 200 {
		adjustmentFactor += 0.05 // 中型项目调整
	} else if totalFP < 50 {
		adjustmentFactor -= 0.05 // 小型项目调整
	}

	// 2. 数据功能点比例调整
	dataFPRatio := evaluation.DataFunctionPoints / evaluation.TotalFunctionPoints
	if dataFPRatio > 0.5 {
		adjustmentFactor += 0.05 // 数据密集型系统
	} else if dataFPRatio < 0.2 {
		adjustmentFactor -= 0.03 // 处理密集型系统
	}

	// 3. 复杂度分布调整
	complexityCounts := make(map[string]int)
	for _, fp := range evaluation.FunctionPoints {
		complexityCounts[fp.ComplexityLevel]++
	}

	highComplexityRatio := float64(complexityCounts["High"]) / float64(len(evaluation.FunctionPoints))
	if highComplexityRatio > 0.4 {
		adjustmentFactor += 0.08 // 高复杂度项目调整
	} else if highComplexityRatio < 0.1 {
		adjustmentFactor -= 0.03 // 低复杂度项目调整
	}

	// 4. 置信度调整
	confidenceScore := e.calculateConfidenceScore(evaluationID)
	if confidenceScore < 0.7 {
		adjustmentFactor += 0.1 // 低置信度需要增加缓冲
	} else if confidenceScore > 0.9 {
		adjustmentFactor -= 0.02 // 高置信度可适当降低
	}

	// 5. 技术架构复杂度调整（基于功能点类型分布）
	typeCounts := make(map[string]int)
	for _, fp := range evaluation.FunctionPoints {
		typeCounts[fp.FunctionType]++
	}

	// 外部接口文件较多表示系统集成复杂度高
	eifRatio := float64(typeCounts["EIF"]) / float64(len(evaluation.FunctionPoints))
	if eifRatio > 0.3 {
		adjustmentFactor += 0.05 // 高集成复杂度调整
	}

	// 输出功能较多表示报表系统
	eoRatio := float64(typeCounts["EO"]) / float64(len(evaluation.FunctionPoints))
	if eoRatio > 0.3 {
		adjustmentFactor += 0.03 // 报表系统复杂度调整
	}

	// 6. 范围限制
	if adjustmentFactor < 0.8 {
		adjustmentFactor = 0.8
	} else if adjustmentFactor > 1.3 {
		adjustmentFactor = 1.3
	}

	return adjustmentFactor
}

// 计算置信度分数
func (e *EvaluationService) calculateConfidenceScore(evaluationID uint) float64 {
	// 获取功能点数据
	var functionPoints []nesma.NesmaFunctionPoint
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&functionPoints).Error; err != nil {
		global.GVA_LOG.Error("获取功能点数据失败", zap.Error(err))
		return 0.0
	}

	if len(functionPoints) == 0 {
		return 0.0
	}

	// 计算总体置信度分数
	var totalConfidence float64
	var weightedConfidence float64
	var totalWeight float64

	for _, fp := range functionPoints {
		// 基础置信度
		confidence := fp.ConfidenceLevel

		// 根据识别方法调整置信度
		switch fp.IdentificationMethod {
		case "manual":
			confidence *= 1.0 // 人工识别最可靠
		case "auto":
			confidence *= 0.8 // 自动识别稍低
		case "ai_assisted":
			confidence *= 0.9 // AI辅助识别较高
		}

		// 根据验证状态调整置信度
		if fp.IsValidated {
			confidence *= 1.1 // 已验证的功能点置信度提升
		}

		// 权重计算（高复杂度功能点权重更高）
		weight := fp.WeightFactor
		weightedConfidence += confidence * weight
		totalWeight += weight
		totalConfidence += confidence
	}

	// 计算加权平均置信度
	var finalScore float64
	if totalWeight > 0 {
		finalScore = weightedConfidence / totalWeight
	} else {
		finalScore = totalConfidence / float64(len(functionPoints))
	}

	// 确保分数在0-1之间
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// 计算准确度分数
func (e *EvaluationService) calculateAccuracyScore(evaluationID uint) float64 {
	// 获取验证结果
	var validationItems []nesma.NesmaValidationItem
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&validationItems).Error; err != nil {
		global.GVA_LOG.Error("获取验证结果失败", zap.Error(err))
		return 0.0
	}

	if len(validationItems) == 0 {
		return 0.8 // 默认值
	}

	// 计算验证通过率
	var passCount int
	var criticalFailCount int
	var totalScore float64

	for _, item := range validationItems {
		switch item.ValidationResult {
		case "pass":
			passCount++
			totalScore += 1.0
		case "warning":
			totalScore += 0.7
		case "fail":
			if item.IsCritical {
				criticalFailCount++
			}
			totalScore += 0.3
		}
	}

	// 基础准确度分数
	baseScore := totalScore / float64(len(validationItems))

	// 关键失败项严重影响准确度
	criticalPenalty := float64(criticalFailCount) * 0.2
	finalScore := baseScore - criticalPenalty

	// 确保分数在0-1之间
	if finalScore < 0 {
		finalScore = 0
	}
	if finalScore > 1.0 {
		finalScore = 1.0
	}

	return finalScore
}

// 计算合规度分数
func (e *EvaluationService) calculateComplianceScore(evaluationID uint) float64 {
	// 获取功能点数据
	var functionPoints []nesma.NesmaFunctionPoint
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&functionPoints).Error; err != nil {
		global.GVA_LOG.Error("获取功能点数据失败", zap.Error(err))
		return 0.0
	}

	if len(functionPoints) == 0 {
		return 0.0
	}

	var complianceScore float64
	totalChecks := 0

	// 1. 权重因子合规检查
	standardWeights := map[string]map[string]float64{
		"ILF": {"Low": 7, "Average": 10, "High": 15},
		"EIF": {"Low": 5, "Average": 7, "High": 10},
		"EI":  {"Low": 3, "Average": 4, "High": 6},
		"EO":  {"Low": 4, "Average": 5, "High": 7},
		"EQ":  {"Low": 3, "Average": 4, "High": 6},
	}

	var correctWeights int
	for _, fp := range functionPoints {
		if typeWeights, exists := standardWeights[fp.FunctionType]; exists {
			if expectedWeight, exists := typeWeights[fp.ComplexityLevel]; exists {
				totalChecks++
				if fp.WeightFactor == expectedWeight {
					correctWeights++
				}
			}
		}
	}

	if totalChecks > 0 {
		complianceScore += (float64(correctWeights) / float64(totalChecks)) * 0.4
	}

	// 2. 功能点类型分布合规检查
	typeCounts := make(map[string]int)
	for _, fp := range functionPoints {
		typeCounts[fp.FunctionType]++
	}

	dataFP := typeCounts["ILF"] + typeCounts["EIF"]
	dataRatio := float64(dataFP) / float64(len(functionPoints))

	// 数据功能点比例应在20%-40%之间
	if dataRatio >= 0.2 && dataRatio <= 0.4 {
		complianceScore += 0.2
	} else if dataRatio >= 0.15 && dataRatio <= 0.5 {
		complianceScore += 0.1
	}

	// 3. 复杂度分布合规检查
	complexityCounts := make(map[string]int)
	for _, fp := range functionPoints {
		complexityCounts[fp.ComplexityLevel]++
	}

	// 检查复杂度分布是否过于集中
	maxRatio := 0.0
	for _, count := range complexityCounts {
		ratio := float64(count) / float64(len(functionPoints))
		if ratio > maxRatio {
			maxRatio = ratio
		}
	}

	if maxRatio <= 0.6 {
		complianceScore += 0.2
	} else if maxRatio <= 0.8 {
		complianceScore += 0.1
	}

	// 4. 基本功能点类型存在性检查
	requiredTypes := []string{"ILF", "EI"}
	existingTypes := 0
	for _, reqType := range requiredTypes {
		if typeCounts[reqType] > 0 {
			existingTypes++
		}
	}
	complianceScore += (float64(existingTypes) / float64(len(requiredTypes))) * 0.2

	// 确保分数在0-1之间
	if complianceScore > 1.0 {
		complianceScore = 1.0
	}

	return complianceScore
}

// 生成改进建议
func (e *EvaluationService) generateRecommendations(evaluationID uint) []map[string]interface{} {
	recommendations := []map[string]interface{}{}

	// 获取评估数据
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("FunctionPoints").First(&evaluation, evaluationID).Error; err != nil {
		global.GVA_LOG.Error("获取评估数据失败", zap.Error(err))
		return recommendations
	}

	// 获取验证结果
	var validationItems []nesma.NesmaValidationItem
	global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&validationItems)

	// 1. 基于置信度的建议
	confidenceScore := e.calculateConfidenceScore(evaluationID)
	if confidenceScore < 0.7 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "quality",
			"title":       "提升评估置信度",
			"description": fmt.Sprintf("当前置信度为%.2f，建议通过以下方式提升：\n1. 补充详细的需求描述\n2. 增加AI辅助识别\n3. 进行人工验证", confidenceScore),
			"priority":    "high",
			"category":    "质量改进",
			"impact":      "高",
			"effort":      "中",
		})
	}

	// 2. 基于验证结果的建议
	criticalFailures := 0
	for _, item := range validationItems {
		if item.IsCritical && item.ValidationResult == "fail" {
			criticalFailures++
			recommendations = append(recommendations, map[string]interface{}{
				"type":        "critical",
				"title":       "修复关键验证失败",
				"description": fmt.Sprintf("验证规则 '%s' 失败：%s", item.ValidationRule, item.ValidationNotes),
				"priority":    "critical",
				"category":    "合规性",
				"impact":      "严重",
				"effort":      "高",
			})
		}
	}

	// 3. 基于功能点分布的建议
	typeCounts := make(map[string]int)
	complexityCounts := make(map[string]int)
	var lowConfidenceCount int

	for _, fp := range evaluation.FunctionPoints {
		typeCounts[fp.FunctionType]++
		complexityCounts[fp.ComplexityLevel]++
		if fp.ConfidenceLevel < 0.6 {
			lowConfidenceCount++
		}
	}

	// 数据功能点比例检查
	dataFPRatio := float64(typeCounts["ILF"]+typeCounts["EIF"]) / float64(len(evaluation.FunctionPoints))
	if dataFPRatio < 0.2 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "architecture",
			"title":       "补充数据功能点",
			"description": fmt.Sprintf("数据功能点比例为%.2f，低于建议的20%%，建议：\n1. 检查是否遗漏数据存储需求\n2. 确认外部接口文件识别的完整性", dataFPRatio),
			"priority":    "medium",
			"category":    "架构完整性",
			"impact":      "中",
			"effort":      "中",
		})
	} else if dataFPRatio > 0.5 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "architecture",
			"title":       "优化数据架构",
			"description": fmt.Sprintf("数据功能点比例为%.2f，高于建议的40%%，建议：\n1. 检查数据存储设计是否存在冗余\n2. 考虑数据标准化和整合", dataFPRatio),
			"priority":    "low",
			"category":    "架构优化",
			"impact":      "低",
			"effort":      "高",
		})
	}

	// 复杂度分布检查
	highComplexityRatio := float64(complexityCounts["High"]) / float64(len(evaluation.FunctionPoints))
	if highComplexityRatio > 0.5 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "complexity",
			"title":       "复杂度管理",
			"description": fmt.Sprintf("高复杂度功能点比例为%.2f，建议：\n1. 分解复杂功能模块\n2. 考虑分阶段实施\n3. 增加开发资源投入", highComplexityRatio),
			"priority":    "high",
			"category":    "风险管理",
			"impact":      "高",
			"effort":      "高",
		})
	}

	// 4. 基于项目规模的建议
	totalFP := evaluation.TotalFunctionPoints
	if totalFP > 500 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "management",
			"title":       "大型项目管理",
			"description": fmt.Sprintf("项目规模为%.1f功能点，属于大型项目，建议：\n1. 采用敏捷开发方法\n2. 分模块并行开发\n3. 加强项目监控", totalFP),
			"priority":    "medium",
			"category":    "项目管理",
			"impact":      "中",
			"effort":      "中",
		})
	} else if totalFP < 50 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "scope",
			"title":       "确认项目范围",
			"description": fmt.Sprintf("项目规模为%.1f功能点，相对较小，建议：\n1. 确认需求收集的完整性\n2. 检查是否遗漏功能模块", totalFP),
			"priority":    "medium",
			"category":    "需求管理",
			"impact":      "中",
			"effort":      "低",
		})
	}

	// 5. 基于AI辅助识别的建议
	aiAssistedCount := 0
	for _, fp := range evaluation.FunctionPoints {
		if fp.IdentificationMethod == "ai_assisted" {
			aiAssistedCount++
		}
	}

	if aiAssistedCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "validation",
			"title":       "AI识别结果验证",
			"description": fmt.Sprintf("有%d个功能点使用AI辅助识别，建议进行人工验证以确保准确性", aiAssistedCount),
			"priority":    "medium",
			"category":    "质量保证",
			"impact":      "中",
			"effort":      "中",
		})
	}

	// 6. 基于外部接口的建议
	eifCount := typeCounts["EIF"]
	if eifCount > len(evaluation.FunctionPoints)/3 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "integration",
			"title":       "系统集成复杂度管理",
			"description": fmt.Sprintf("外部接口文件数量为%d，占比较高，建议：\n1. 制定详细的接口规范\n2. 考虑接口标准化\n3. 增加集成测试", eifCount),
			"priority":    "medium",
			"category":    "系统集成",
			"impact":      "中",
			"effort":      "中",
		})
	}

	// 7. 基于低置信度功能点的建议
	if lowConfidenceCount > 0 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "improvement",
			"title":       "提升功能点识别准确性",
			"description": fmt.Sprintf("有%d个功能点置信度较低，建议：\n1. 补充详细的功能描述\n2. 进行专家评审\n3. 使用AI辅助识别", lowConfidenceCount),
			"priority":    "medium",
			"category":    "质量改进",
			"impact":      "中",
			"effort":      "中",
		})
	}

	// 8. 基于合规性得分的建议
	complianceScore := e.calculateComplianceScore(evaluationID)
	if complianceScore < 0.8 {
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "compliance",
			"title":       "提升NESMA标准合规性",
			"description": fmt.Sprintf("当前合规性得分为%.2f，建议：\n1. 检查权重因子设置\n2. 确认功能点分类准确性\n3. 参考NESMA标准指南", complianceScore),
			"priority":    "high",
			"category":    "标准合规",
			"impact":      "高",
			"effort":      "中",
		})
	}

	// 9. 成本和工期估算建议
	if totalFP > 0 {
		estimatedDays := int(totalFP * 0.5) // 基础估算：每功能点0.5天
		recommendations = append(recommendations, map[string]interface{}{
			"type":        "estimation",
			"title":       "开发工期估算",
			"description": fmt.Sprintf("基于%.1f功能点，预估开发工期约%d天，建议：\n1. 考虑团队技能水平调整\n2. 增加20%%的缓冲时间\n3. 制定详细的开发计划", totalFP, estimatedDays),
			"priority":    "low",
			"category":    "项目规划",
			"impact":      "低",
			"effort":      "低",
		})
	}

	// 按优先级排序
	priorityOrder := map[string]int{"critical": 1, "high": 2, "medium": 3, "low": 4}

	// 简单的排序逻辑
	for i := 0; i < len(recommendations); i++ {
		for j := i + 1; j < len(recommendations); j++ {
			priI := priorityOrder[recommendations[i]["priority"].(string)]
			priJ := priorityOrder[recommendations[j]["priority"].(string)]
			if priI > priJ {
				recommendations[i], recommendations[j] = recommendations[j], recommendations[i]
			}
		}
	}

	return recommendations
}

// 获取评估详情
func (e *EvaluationService) GetEvaluation(evaluationID uint) (nesmaRes.EvaluationResponse, error) {
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Preload("Project").
		Preload("Cycle").
		Preload("RequirementVersion").
		Preload("FunctionPoints").
		Preload("ComplexityMetrics").
		Preload("ValidationItems").
		First(&evaluation, evaluationID).Error; err != nil {
		return nesmaRes.EvaluationResponse{}, err
	}

	return e.buildEvaluationResponse(evaluation), nil
}

// 获取评估列表
func (e *EvaluationService) GetEvaluationList(req nesmaReq.EvaluationSearchRequest) (nesmaRes.EvaluationListResponse, error) {
	var evaluations []nesma.NesmaEvaluation
	var total int64

	query := global.GVA_DB.Model(&nesma.NesmaEvaluation{})

	// 添加筛选条件
	if req.ProjectID != 0 {
		query = query.Where("project_id = ?", req.ProjectID)
	}
	if req.CycleID != nil && *req.CycleID != 0 {
		query = query.Where("cycle_id = ?", *req.CycleID)
	}
	if req.RequirementVersionID != nil && *req.RequirementVersionID != 0 {
		query = query.Where("requirement_version_id = ?", *req.RequirementVersionID)
	}
	if req.EvaluatorID != 0 {
		query = query.Where("evaluator_id = ?", req.EvaluatorID)
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.EvaluationType != "" {
		query = query.Where("evaluation_type = ?", req.EvaluationType)
	}
	if req.StartTime != nil {
		query = query.Where("start_time >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("completion_time <= ?", req.EndTime)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nesmaRes.EvaluationListResponse{}, err
	}

	// 分页查询
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("Project").
		Preload("Cycle").
		Preload("RequirementVersion").
		Order("updated_at desc").
		Limit(req.PageSize).
		Offset(offset).
		Find(&evaluations).Error; err != nil {
		return nesmaRes.EvaluationListResponse{}, err
	}

	// 构建响应
	var responses []nesmaRes.EvaluationResponse
	for _, eval := range evaluations {
		responses = append(responses, e.buildEvaluationResponse(eval))
	}

	return nesmaRes.EvaluationListResponse{
		List:     responses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, nil
}

// 获取评估统计
func (e *EvaluationService) GetEvaluationStats(evaluationID uint) (nesmaRes.EvaluationStatsResponse, error) {
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.First(&evaluation, evaluationID).Error; err != nil {
		return nesmaRes.EvaluationStatsResponse{}, err
	}

	// 获取功能点数据
	var functionPoints []nesma.NesmaFunctionPoint
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&functionPoints).Error; err != nil {
		return nesmaRes.EvaluationStatsResponse{}, err
	}

	// 计算统计数据
	var stats nesmaRes.EvaluationStatsResponse
	stats.TotalFunctionPoints = evaluation.TotalFunctionPoints
	stats.DataFunctionPointsRatio = evaluation.DataFunctionPoints / evaluation.TotalFunctionPoints * 100
	stats.TransactionalFPRatio = evaluation.TransactionalFP / evaluation.TotalFunctionPoints * 100

	// 功能点类型统计
	for _, fp := range functionPoints {
		switch fp.FunctionType {
		case "ILF":
			stats.ILFCount++
		case "EIF":
			stats.EIFCount++
		case "EI":
			stats.EICount++
		case "EO":
			stats.EOCount++
		case "EQ":
			stats.EQCount++
		}

		// 复杂度级别统计
		switch fp.ComplexityLevel {
		case "Low":
			stats.LowComplexityCount++
		case "Average":
			stats.AverageComplexityCount++
		case "High":
			stats.HighComplexityCount++
		}
	}

	// 验证通过率
	var validationItems []nesma.NesmaValidationItem
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&validationItems).Error; err == nil {
		passCount := 0
		criticalCount := 0
		for _, item := range validationItems {
			if item.ValidationResult == "pass" {
				passCount++
			}
			if item.IsCritical {
				criticalCount++
			}
		}
		if len(validationItems) > 0 {
			stats.ValidationPassRate = float64(passCount) / float64(len(validationItems)) * 100
		}
		stats.CriticalIssuesCount = criticalCount
	}

	// 计算置信度比率
	highConfidenceCount := 0
	for _, fp := range functionPoints {
		if fp.ConfidenceLevel > 0.8 {
			highConfidenceCount++
		}
	}
	if len(functionPoints) > 0 {
		stats.HighConfidencePointsRatio = float64(highConfidenceCount) / float64(len(functionPoints)) * 100
	}

	// 计算平均复杂度分数
	if len(functionPoints) > 0 {
		complexitySum := 0.0
		for _, fp := range functionPoints {
			complexitySum += fp.WeightFactor
		}
		stats.AverageComplexityScore = complexitySum / float64(len(functionPoints))
	}

	return stats, nil
}

// 获取功能点列表
func (e *EvaluationService) GetFunctionPoints(evaluationID uint, functionType, complexityLevel string, page, pageSize int) (map[string]interface{}, error) {
	global.GVA_LOG.Info("GetFunctionPoints服务开始", 
		zap.Uint("evaluationID", evaluationID),
		zap.String("functionType", functionType),
		zap.String("complexityLevel", complexityLevel),
		zap.Int("page", page),
		zap.Int("pageSize", pageSize))
	
	// 验证评估是否存在
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.First(&evaluation, evaluationID).Error; err != nil {
		global.GVA_LOG.Error("查询评估记录失败", 
			zap.Uint("evaluationID", evaluationID), 
			zap.Error(err))
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("评估记录不存在")
		}
		return nil, fmt.Errorf("查询评估记录失败: %v", err)
	}
	
	global.GVA_LOG.Info("评估记录查询成功", 
		zap.Uint("evaluationID", evaluationID),
		zap.String("evaluationName", evaluation.EvaluationName),
		zap.String("status", evaluation.Status))

	var functionPoints []nesma.NesmaFunctionPoint
	var total int64

	query := global.GVA_DB.Model(&nesma.NesmaFunctionPoint{}).Where("evaluation_id = ?", evaluationID)

	// 添加筛选条件
	if functionType != "" {
		query = query.Where("function_type = ?", functionType)
	}
	if complexityLevel != "" {
		query = query.Where("complexity_level = ?", complexityLevel)
	}

	// 获取总数
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	if err := query.Preload("Requirement").
		Order("created_at desc").
		Limit(pageSize).
		Offset(offset).
		Find(&functionPoints).Error; err != nil {
		return nil, err
	}

	// 构建响应
	var responses []nesmaRes.FunctionPointResponse
	for _, fp := range functionPoints {
		responses = append(responses, nesmaRes.FunctionPointResponse{
			ID:                   fp.ID,
			EvaluationID:         fp.EvaluationID,
			RequirementID:        fp.RequirementID,
			FunctionType:         fp.FunctionType,
			FunctionName:         fp.FunctionName,
			FunctionDesc:         fp.FunctionDesc,
			DataElements:         fp.DataElements,
			FileTypes:            fp.FileTypes,
			RecordElements:       fp.RecordElements,
			ComplexityLevel:      fp.ComplexityLevel,
			WeightFactor:         fp.WeightFactor,
			CalculatedPoints:     fp.CalculatedPoints,
			IdentificationMethod: fp.IdentificationMethod,
			ConfidenceLevel:      fp.ConfidenceLevel,
			IsValidated:          fp.IsValidated,
			ValidationStatus:     fp.ValidationStatus,
			ValidationNotes:      fp.ValidationNotes,
			CreatedAt:            fp.CreatedAt,
			UpdatedAt:            fp.UpdatedAt,
			Requirement:          fp.Requirement,
		})
	}

	return map[string]interface{}{
		"list":     responses,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	}, nil
}

// 创建功能点
func (e *EvaluationService) CreateFunctionPoint(req nesmaReq.FunctionPointRequest) (nesmaRes.FunctionPointResponse, error) {
	// 计算权重因子
	if req.WeightFactor == 0 {
		req.WeightFactor = e.getWeightFactorFromEvaluation(req.EvaluationID, req.FunctionType, req.ComplexityLevel)
	}

	functionPoint := nesma.NesmaFunctionPoint{
		EvaluationID:         req.EvaluationID,
		RequirementID:        req.RequirementID,
		FunctionType:         req.FunctionType,
		FunctionName:         req.FunctionName,
		FunctionDesc:         req.FunctionDesc,
		DataElements:         req.DataElements,
		FileTypes:            req.FileTypes,
		RecordElements:       req.RecordElements,
		ComplexityLevel:      req.ComplexityLevel,
		WeightFactor:         req.WeightFactor,
		CalculatedPoints:     req.WeightFactor,
		IdentificationMethod: req.IdentificationMethod,
		ConfidenceLevel:      req.ConfidenceLevel,
		IsValidated:          req.IsValidated,
		ValidationNotes:      req.ValidationNotes,
	}

	if err := global.GVA_DB.Create(&functionPoint).Error; err != nil {
		return nesmaRes.FunctionPointResponse{}, err
	}

	// 重新计算评估总分
	e.recalculateEvaluationTotals(req.EvaluationID)

	return nesmaRes.FunctionPointResponse{
		ID:                   functionPoint.ID,
		EvaluationID:         functionPoint.EvaluationID,
		RequirementID:        functionPoint.RequirementID,
		FunctionType:         functionPoint.FunctionType,
		FunctionName:         functionPoint.FunctionName,
		FunctionDesc:         functionPoint.FunctionDesc,
		DataElements:         functionPoint.DataElements,
		FileTypes:            functionPoint.FileTypes,
		RecordElements:       functionPoint.RecordElements,
		ComplexityLevel:      functionPoint.ComplexityLevel,
		WeightFactor:         functionPoint.WeightFactor,
		CalculatedPoints:     functionPoint.CalculatedPoints,
		IdentificationMethod: functionPoint.IdentificationMethod,
		ConfidenceLevel:      functionPoint.ConfidenceLevel,
		IsValidated:          functionPoint.IsValidated,
		ValidationStatus:     functionPoint.ValidationStatus,
		ValidationNotes:      functionPoint.ValidationNotes,
		CreatedAt:            functionPoint.CreatedAt,
		UpdatedAt:            functionPoint.UpdatedAt,
	}, nil
}

// 更新功能点
func (e *EvaluationService) UpdateFunctionPoint(fpID uint, req nesmaReq.FunctionPointRequest) error {
	var functionPoint nesma.NesmaFunctionPoint
	if err := global.GVA_DB.First(&functionPoint, fpID).Error; err != nil {
		return err
	}

	// 计算权重因子
	if req.WeightFactor == 0 {
		req.WeightFactor = e.getWeightFactorFromEvaluation(req.EvaluationID, req.FunctionType, req.ComplexityLevel)
	}

	// 更新字段
	functionPoint.FunctionType = req.FunctionType
	functionPoint.FunctionName = req.FunctionName
	functionPoint.FunctionDesc = req.FunctionDesc
	functionPoint.DataElements = req.DataElements
	functionPoint.FileTypes = req.FileTypes
	functionPoint.RecordElements = req.RecordElements
	functionPoint.ComplexityLevel = req.ComplexityLevel
	functionPoint.WeightFactor = req.WeightFactor
	functionPoint.CalculatedPoints = req.WeightFactor
	functionPoint.IdentificationMethod = req.IdentificationMethod
	functionPoint.ConfidenceLevel = req.ConfidenceLevel
	functionPoint.IsValidated = req.IsValidated
	functionPoint.ValidationNotes = req.ValidationNotes

	if err := global.GVA_DB.Save(&functionPoint).Error; err != nil {
		return err
	}

	// 重新计算评估总分
	e.recalculateEvaluationTotals(functionPoint.EvaluationID)

	return nil
}

// 删除功能点
func (e *EvaluationService) DeleteFunctionPoint(fpID uint) error {
	var functionPoint nesma.NesmaFunctionPoint
	if err := global.GVA_DB.First(&functionPoint, fpID).Error; err != nil {
		return err
	}

	evaluationID := functionPoint.EvaluationID

	if err := global.GVA_DB.Delete(&functionPoint).Error; err != nil {
		return err
	}

	// 重新计算评估总分
	e.recalculateEvaluationTotals(evaluationID)

	return nil
}

// 删除评估
func (e *EvaluationService) DeleteEvaluation(evaluationID uint) error {
	// 开始事务
	tx := global.GVA_DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 检查评估是否存在
	var evaluation nesma.NesmaEvaluation
	if err := tx.First(&evaluation, evaluationID).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("评估不存在: %v", err)
	}

	// 2. 检查评估状态，如果正在进行中则不允许删除
	if evaluation.Status == "processing" {
		tx.Rollback()
		return fmt.Errorf("不能删除正在进行的评估")
	}

	// 3. 删除关联的验证项
	if err := tx.Where("evaluation_id = ?", evaluationID).Delete(&nesma.NesmaValidationItem{}).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("删除验证项失败", zap.Error(err))
		return fmt.Errorf("删除验证项失败: %v", err)
	}

	// 4. 删除关联的复杂度指标
	if err := tx.Where("evaluation_id = ?", evaluationID).Delete(&nesma.NesmaComplexityMetric{}).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("删除复杂度指标失败", zap.Error(err))
		return fmt.Errorf("删除复杂度指标失败: %v", err)
	}

	// 5. 删除关联的功能点
	if err := tx.Where("evaluation_id = ?", evaluationID).Delete(&nesma.NesmaFunctionPoint{}).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("删除功能点失败", zap.Error(err))
		return fmt.Errorf("删除功能点失败: %v", err)
	}

	// 6. 最后删除评估本身
	if err := tx.Delete(&evaluation).Error; err != nil {
		tx.Rollback()
		global.GVA_LOG.Error("删除评估失败", zap.Error(err))
		return fmt.Errorf("删除评估失败: %v", err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("提交删除事务失败: %v", err)
	}

	global.GVA_LOG.Info("评估删除成功", zap.Uint("evaluationID", evaluationID), zap.String("evaluationName", evaluation.EvaluationName))
	return nil
}

// 获取复杂度指标
func (e *EvaluationService) GetComplexityMetrics(evaluationID uint) ([]nesmaRes.ComplexityMetricResponse, error) {
	var metrics []nesma.NesmaComplexityMetric
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&metrics).Error; err != nil {
		return nil, err
	}

	var responses []nesmaRes.ComplexityMetricResponse
	for _, metric := range metrics {
		responses = append(responses, nesmaRes.ComplexityMetricResponse{
			ID:                 metric.ID,
			EvaluationID:       metric.EvaluationID,
			MetricType:         metric.MetricType,
			MetricName:         metric.MetricName,
			MetricValue:        metric.MetricValue,
			ThresholdValue:     metric.ThresholdValue,
			Score:              metric.Score,
			Weight:             metric.Weight,
			CalculationFormula: metric.CalculationFormula,
			QualityIndicator:   metric.QualityIndicator,
			ImpactLevel:        metric.ImpactLevel,
			CreatedAt:          metric.CreatedAt,
			UpdatedAt:          metric.UpdatedAt,
		})
	}

	return responses, nil
}

// 获取验证项目
func (e *EvaluationService) GetValidationItems(evaluationID uint, validationResult string) ([]nesmaRes.ValidationItemResponse, error) {
	var items []nesma.NesmaValidationItem
	query := global.GVA_DB.Where("evaluation_id = ?", evaluationID)

	if validationResult != "" {
		query = query.Where("validation_result = ?", validationResult)
	}

	if err := query.Find(&items).Error; err != nil {
		return nil, err
	}

	var responses []nesmaRes.ValidationItemResponse
	for _, item := range items {
		responses = append(responses, nesmaRes.ValidationItemResponse{
			ID:               item.ID,
			EvaluationID:     item.EvaluationID,
			ValidationRule:   item.ValidationRule,
			RuleDescription:  item.RuleDescription,
			ValidationResult: item.ValidationResult,
			ExpectedValue:    item.ExpectedValue,
			ActualValue:      item.ActualValue,
			DeviationLevel:   item.DeviationLevel,
			ValidationNotes:  item.ValidationNotes,
			IsCritical:       item.IsCritical,
			ResolutionStatus: item.ResolutionStatus,
			ResolutionNotes:  item.ResolutionNotes,
			CreatedAt:        item.CreatedAt,
			UpdatedAt:        item.UpdatedAt,
		})
	}

	return responses, nil
}

// 更新验证项目
func (e *EvaluationService) UpdateValidationItem(itemID uint, req nesmaReq.ValidationItemRequest) error {
	var item nesma.NesmaValidationItem
	if err := global.GVA_DB.First(&item, itemID).Error; err != nil {
		return err
	}

	// 更新字段
	item.ValidationResult = req.ValidationResult
	item.ExpectedValue = req.ExpectedValue
	item.ActualValue = req.ActualValue
	item.DeviationLevel = req.DeviationLevel
	item.ValidationNotes = req.ValidationNotes
	item.IsCritical = req.IsCritical
	item.ResolutionStatus = req.ResolutionStatus
	item.ResolutionNotes = req.ResolutionNotes

	return global.GVA_DB.Save(&item).Error
}

// 评估审核
func (e *EvaluationService) ReviewEvaluation(req nesmaReq.ReviewEvaluationRequest) error {
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.First(&evaluation, req.EvaluationID).Error; err != nil {
		return err
	}

	// 更新审核信息
	evaluation.ReviewStatus = req.ReviewStatus
	evaluation.ReviewerID = &req.ReviewerID
	evaluation.ReviewComments = req.ReviewComments
	now := time.Now()
	evaluation.ReviewTime = &now

	return global.GVA_DB.Save(&evaluation).Error
}

// 导出评估报告
func (e *EvaluationService) ExportEvaluationReport(evaluationID uint, format string) (map[string]interface{}, error) {
	// 获取评估数据
	_, err := e.GetEvaluation(evaluationID)
	if err != nil {
		return nil, err
	}

	// 这里应该实现具体的报告生成逻辑
	// 暂时返回模拟数据
	filename := fmt.Sprintf("evaluation_%d_%s.%s", evaluationID, time.Now().Format("20060102150405"), format)

	return map[string]interface{}{
		"downloadUrl": fmt.Sprintf("/uploads/reports/%s", filename),
		"fileName":    filename,
		"fileSize":    "1.2MB",
	}, nil
}

// 获取项目评估摘要
func (e *EvaluationService) GetProjectEvaluationSummary(projectID uint) (nesmaRes.EvaluationSummaryResponse, error) {
	var summary nesmaRes.EvaluationSummaryResponse
	summary.ProjectID = projectID

	// 获取项目信息
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, projectID).Error; err != nil {
		return summary, err
	}
	summary.ProjectName = project.Name

	// 统计评估数量
	var totalEvaluations int64
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("project_id = ?", projectID).Count(&totalEvaluations)
	summary.TotalEvaluations = int(totalEvaluations)

	var completedEvaluations int64
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("project_id = ? AND status = ?", projectID, "completed").Count(&completedEvaluations)
	summary.CompletedEvaluations = int(completedEvaluations)

	// 获取最新评估
	var latestEvaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.Where("project_id = ?", projectID).Order("created_at desc").First(&latestEvaluation).Error; err == nil {
		summary.LatestEvaluationID = latestEvaluation.ID
		summary.LatestEvaluationName = latestEvaluation.EvaluationName
		summary.LatestEvaluationTime = &latestEvaluation.CreatedAt
	}

	// 计算平均值
	if completedEvaluations > 0 {
		var avgFP, avgConfidence, avgAccuracy, avgCompliance float64
		global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
			Where("project_id = ? AND status = ?", projectID, "completed").
			Select("AVG(total_function_points) as avg_fp, AVG(confidence_score) as avg_confidence, AVG(accuracy_score) as avg_accuracy, AVG(compliance_score) as avg_compliance").
			Scan(&struct {
				AvgFP         float64 `json:"avg_fp"`
				AvgConfidence float64 `json:"avg_confidence"`
				AvgAccuracy   float64 `json:"avg_accuracy"`
				AvgCompliance float64 `json:"avg_compliance"`
			}{
				AvgFP:         avgFP,
				AvgConfidence: avgConfidence,
				AvgAccuracy:   avgAccuracy,
				AvgCompliance: avgCompliance,
			})

		summary.AverageFunctionPoints = avgFP
		summary.AverageConfidenceScore = avgConfidence
		summary.AverageAccuracyScore = avgAccuracy
		summary.AverageComplianceScore = avgCompliance
	}

	return summary, nil
}

// 重新计算评估
func (e *EvaluationService) RecalculateEvaluation(evaluationID uint) error {
	// 重新启动评估计算
	return e.StartAutoEvaluation(evaluationID)
}

// 重新计算评估总分
func (e *EvaluationService) recalculateEvaluationTotals(evaluationID uint) {
	var functionPoints []nesma.NesmaFunctionPoint
	if err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&functionPoints).Error; err != nil {
		return
	}

	var totalFP, dataFP, transactionalFP float64
	var simpleFunctions, avgFunctions, complexFunctions int

	for _, fp := range functionPoints {
		totalFP += fp.CalculatedPoints

		if fp.FunctionType == "ILF" || fp.FunctionType == "EIF" {
			dataFP += fp.CalculatedPoints
		} else {
			transactionalFP += fp.CalculatedPoints
		}

		switch fp.ComplexityLevel {
		case "Low":
			simpleFunctions++
		case "Average":
			avgFunctions++
		case "High":
			complexFunctions++
		}
	}

	// 更新评估记录
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evaluationID).
		Updates(map[string]interface{}{
			"total_function_points":    totalFP,
			"data_function_points":     dataFP,
			"transactional_fp":         transactionalFP,
			"simple_function_count":    simpleFunctions,
			"average_function_count":   avgFunctions,
			"complex_function_count":   complexFunctions,
			"adjusted_function_points": totalFP, // 这里可以加入调整因子计算
		})
}

// GetUserOptions 获取用户选项列表
// GetProjectCycles 获取项目周期选项列表
func (e *EvaluationService) GetProjectCycles(projectID uint) ([]map[string]interface{}, error) {
	var cycles []nesma.NesmaProjectCycle
	if err := global.GVA_DB.Where("project_id = ?", projectID).
		Order("created_at ASC").
		Find(&cycles).Error; err != nil {
		return nil, err
	}

	var options []map[string]interface{}
	for _, cycle := range cycles {
		options = append(options, map[string]interface{}{
			"value":       cycle.ID,
			"label":       cycle.Name,
			"description": cycle.Description,
			"status":      cycle.Status,
			"statusText":  cycle.GetStatusText(),
			"phase":       cycle.Phase,
			"phaseText":   cycle.GetPhaseText(),
			"progress":    cycle.Progress,
			"startDate":   cycle.StartDate,
			"endDate":     cycle.EndDate,
		})
	}
	return options, nil
}

// GetCycleVersions 获取周期版本选项列表
func (e *EvaluationService) GetCycleVersions(cycleID uint) ([]map[string]interface{}, error) {
	var versions []nesma.NesmaRequirementVersion
	if err := global.GVA_DB.Where("cycle_id = ?", cycleID).
		Order("created_at ASC").
		Find(&versions).Error; err != nil {
		return nil, err
	}

	var options []map[string]interface{}
	for _, version := range versions {
		options = append(options, map[string]interface{}{
			"value":            version.ID,
			"label":            version.Version,
			"versionType":      version.VersionType,
			"versionTypeText":  version.GetVersionTypeText(),
			"createdBy":        version.CreatedBy,
			"createdByText":    version.GetCreatedByText(),
			"summary":          version.Summary,
			"status":           version.Status,
			"statusText":       version.GetStatusText(),
			"requirementCount": version.RequirementCount,
			"analysisProgress": version.AnalysisProgress,
			"qualityScore":     version.QualityScore,
			"confidenceScore":  version.ConfidenceScore,
			"reviewStatus":     version.ReviewStatus,
			"reviewStatusText": version.GetReviewStatusText(),
		})
	}
	return options, nil
}

func (es *EvaluationService) GetUserOptions(ctx context.Context) ([]map[string]interface{}, error) {
	var users []map[string]interface{}
	
	// 查询系统用户表，获取用户选项
	err := global.GVA_DB.Table("sys_users").
		Select("id as value, username as label, nick_name").
		Where("enable = ?", 1).
		Scan(&users).Error
	
	if err != nil {
		global.GVA_LOG.Error("查询用户选项失败", zap.Error(err))
		// 返回默认用户选项
		return []map[string]interface{}{
			{
				"value": 1,
				"label": "admin",
				"nick_name": "超级管理员",
			},
		}, nil
	}
	
	return users, nil
}

// ==================== 评估因子配置相关方法 ====================

// EvaluationFactors 评估因子配置结构
type EvaluationFactors struct {
	EvaluationID     uint                   `json:"evaluationId"`
	DataFunctions    DataFunctionWeights    `json:"dataFunctions"`
	TransactionFunctions TransactionFunctionWeights `json:"transactionFunctions"`
	AdjustmentFactors []AdjustmentFactor    `json:"adjustmentFactors"`
	CalculatedValues CalculatedValues      `json:"calculatedValues"`
	CreatedAt        time.Time             `json:"createdAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
}

// DataFunctionWeights 数据功能权重配置
type DataFunctionWeights struct {
	ILF []FunctionWeight `json:"ilf"`
	EIF []FunctionWeight `json:"eif"`
}

// TransactionFunctionWeights 事务功能权重配置
type TransactionFunctionWeights struct {
	EI []FunctionWeight `json:"ei"`
	EO []FunctionWeight `json:"eo"`
	EQ []FunctionWeight `json:"eq"`
}

// FunctionWeight 功能权重
type FunctionWeight struct {
	Complexity  string  `json:"complexity"`
	Weight      int     `json:"weight"`
	Description string  `json:"description"`
}

// AdjustmentFactor 调整因子
type AdjustmentFactor struct {
	Name        string `json:"name"`
	Value       int    `json:"value"`
	Description string `json:"description"`
}

// CalculatedValues 计算值
type CalculatedValues struct {
	TotalDegreeOfInfluence  int     `json:"totalDegreeOfInfluence"`
	ValueAdjustmentFactor   float64 `json:"valueAdjustmentFactor"`
}

// GetEvaluationFactors 获取评估因子配置
func (es *EvaluationService) GetEvaluationFactors(evaluationId uint) (*EvaluationFactors, error) {
	// 查询评估基本信息
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationId).Error
	if err != nil {
		global.GVA_LOG.Error("查询评估信息失败", zap.Error(err))
		return nil, err
	}

	// 查询已保存的因子配置
	var factorRecord nesma.NesmaEvaluationFactors
	err = global.GVA_DB.Where("evaluation_id = ?", evaluationId).
		First(&factorRecord).Error
	
	if err != nil {
		// 如果没有找到配置记录，返回默认配置
		global.GVA_LOG.Info("未找到已保存的评估因子配置，返回默认配置", zap.Uint("evaluationId", evaluationId))
		return es.getDefaultEvaluationFactors(evaluationId), nil
	}

	// 解析已保存的配置
	factors := &EvaluationFactors{}
	if factorRecord.FactorConfig != "" {
		err = json.Unmarshal([]byte(factorRecord.FactorConfig), factors)
		if err != nil {
			global.GVA_LOG.Error("解析评估因子配置失败", zap.Error(err))
			return es.getDefaultEvaluationFactors(evaluationId), nil
		}
	}

	// 调试日志：记录读取到的EO数据
	global.GVA_LOG.Info("获取评估因子配置后的EO数据", 
		zap.Uint("evaluationId", evaluationId),
		zap.Int("eoCount", len(factors.TransactionFunctions.EO)),
		zap.Any("eoData", factors.TransactionFunctions.EO))

	factors.EvaluationID = evaluationId
	return factors, nil
}

// SaveEvaluationFactors 保存评估因子配置
func (es *EvaluationService) SaveEvaluationFactors(factors *EvaluationFactors) error {
	// 验证配置数据
	if err := es.validateEvaluationFactors(factors); err != nil {
		return fmt.Errorf("配置数据验证失败: %v", err)
	}

	// 调试日志：记录保存前的EO数据
	global.GVA_LOG.Info("保存评估因子配置前的EO数据", 
		zap.Uint("evaluationId", factors.EvaluationID),
		zap.Int("eoCount", len(factors.TransactionFunctions.EO)),
		zap.Any("eoData", factors.TransactionFunctions.EO))

	// 序列化配置为JSON
	configJson, err := json.Marshal(factors)
	if err != nil {
		global.GVA_LOG.Error("序列化评估因子配置失败", zap.Error(err))
		return err
	}

	// 查询是否已存在配置记录
	var existingRecord nesma.NesmaEvaluationFactors
	err = global.GVA_DB.Where("evaluation_id = ?", factors.EvaluationID).
		First(&existingRecord).Error
	
	if err != nil {
		// 创建新记录
		factorRecord := nesma.NesmaEvaluationFactors{
			EvaluationID:  factors.EvaluationID,
			FactorConfig:  string(configJson),
			NesmaVersion:  "v2.2",
		}
		
		err = global.GVA_DB.Create(&factorRecord).Error
		if err != nil {
			global.GVA_LOG.Error("创建评估因子配置记录失败", zap.Error(err))
			return err
		}
	} else {
		// 更新现有记录
		existingRecord.FactorConfig = string(configJson)
		err = global.GVA_DB.Save(&existingRecord).Error
		if err != nil {
			global.GVA_LOG.Error("更新评估因子配置记录失败", zap.Error(err))
			return err
		}
	}

	global.GVA_LOG.Info("评估因子配置保存成功", zap.Uint("evaluationId", factors.EvaluationID))
	return nil
}

// UpdateEvaluationFactors 更新评估因子配置
func (es *EvaluationService) UpdateEvaluationFactors(evaluationId uint, factors *EvaluationFactors) error {
	factors.EvaluationID = evaluationId
	return es.SaveEvaluationFactors(factors)
}

// GetDefaultNESMAFactors 获取默认NESMA因子配置
func (es *EvaluationService) GetDefaultNESMAFactors(nesmaVersion string) (*EvaluationFactors, error) {
	defaultFactors := es.getDefaultEvaluationFactors(0)
	return defaultFactors, nil
}

// ValidateEvaluationFactors 验证评估因子配置
func (es *EvaluationService) ValidateEvaluationFactors(factors *EvaluationFactors) error {
	return es.validateEvaluationFactors(factors)
}

// CalculateAdjustmentFactor 计算调整因子
func (es *EvaluationService) CalculateAdjustmentFactor(adjustmentFactors []AdjustmentFactor) (*CalculatedValues, error) {
	totalInfluence := 0
	for _, factor := range adjustmentFactors {
		if factor.Value < 0 || factor.Value > 5 {
			return nil, fmt.Errorf("调整因子值必须在0-5之间: %s = %d", factor.Name, factor.Value)
		}
		totalInfluence += factor.Value
	}

	valueAdjustmentFactor := float64(totalInfluence)*0.01 + 0.65

	return &CalculatedValues{
		TotalDegreeOfInfluence: totalInfluence,
		ValueAdjustmentFactor:  valueAdjustmentFactor,
	}, nil
}

// PreviewFunctionPointCalculation 预览功能点计算
func (es *EvaluationService) PreviewFunctionPointCalculation(data map[string]interface{}) (map[string]interface{}, error) {
	// 解析输入数据
	factorsData, ok := data["factors"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("缺少factors数据")
	}

	// 模拟计算功能点
	var totalUFP float64 = 0
	var breakdown = make(map[string]interface{})

	// 计算数据功能点
	if dataFunctions, ok := factorsData["dataFunctions"].(map[string]interface{}); ok {
		if ilf, ok := dataFunctions["ilf"].([]interface{}); ok {
			var ilfTotal float64 = 0
			for _, item := range ilf {
				if weightItem, ok := item.(map[string]interface{}); ok {
					if weight, ok := weightItem["weight"].(float64); ok {
						ilfTotal += weight
					}
				}
			}
			breakdown["ilf"] = ilfTotal
			totalUFP += ilfTotal
		}
	}

	// 计算调整因子
	adjustmentFactor := 1.0
	if adjustments, ok := factorsData["adjustmentFactors"].([]interface{}); ok {
		totalInfluence := 0
		for _, item := range adjustments {
			if adjItem, ok := item.(map[string]interface{}); ok {
				if value, ok := adjItem["value"].(float64); ok {
					totalInfluence += int(value)
				}
			}
		}
		adjustmentFactor = float64(totalInfluence)*0.01 + 0.65
	}

	adjustedFP := totalUFP * adjustmentFactor

	result := map[string]interface{}{
		"unfunctionPoints":      totalUFP,
		"adjustmentFactor":      adjustmentFactor,
		"adjustedFunctionPoints": adjustedFP,
		"breakdown":             breakdown,
		"calculatedAt":          time.Now(),
	}

	return result, nil
}

// ApplyFactorsToProject 应用评估因子到项目
func (es *EvaluationService) ApplyFactorsToProject(evaluationId uint, data map[string]interface{}) error {
	// 获取评估信息
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationId).Error
	if err != nil {
		return err
	}

	// 应用因子配置到项目评估
	// 这里可以实现具体的应用逻辑
	global.GVA_LOG.Info("应用评估因子到项目", 
		zap.Uint("evaluationId", evaluationId),
		zap.Uint("projectId", evaluation.ProjectID))

	return nil
}

// GetFactorsHistory 获取评估因子历史记录
func (es *EvaluationService) GetFactorsHistory(evaluationId uint) ([]map[string]interface{}, error) {
	var history []map[string]interface{}
	
	err := global.GVA_DB.Table("nesma_evaluation_factors").
		Select("id, evaluation_id, nesma_version, created_at, updated_at").
		Where("evaluation_id = ?", evaluationId).
		Order("updated_at DESC").
		Scan(&history).Error
	
	if err != nil {
		global.GVA_LOG.Error("查询评估因子历史失败", zap.Error(err))
		return nil, err
	}

	return history, nil
}

// ImportEvaluationFactors 导入评估因子配置
func (es *EvaluationService) ImportEvaluationFactors(evaluationId uint, factorsData []byte) error {
	var factors EvaluationFactors
	err := json.Unmarshal(factorsData, &factors)
	if err != nil {
		return fmt.Errorf("解析导入数据失败: %v", err)
	}

	factors.EvaluationID = evaluationId
	return es.SaveEvaluationFactors(&factors)
}

// ExportEvaluationFactors 导出评估因子配置
func (es *EvaluationService) ExportEvaluationFactors(evaluationId uint, format string) ([]byte, error) {
	factors, err := es.GetEvaluationFactors(evaluationId)
	if err != nil {
		return nil, err
	}

	switch format {
	case "json":
		return json.MarshalIndent(factors, "", "  ")
	default:
		return json.Marshal(factors)
	}
}

// 私有辅助方法

// getDefaultEvaluationFactors 获取默认评估因子配置
func (es *EvaluationService) getDefaultEvaluationFactors(evaluationId uint) *EvaluationFactors {
	return &EvaluationFactors{
		EvaluationID: evaluationId,
		DataFunctions: DataFunctionWeights{
			ILF: []FunctionWeight{
				{Complexity: "low", Weight: 7, Description: "1个RET，1-19个DET 或者 2-5个RET，1-19个DET"},
				{Complexity: "average", Weight: 10, Description: "1个RET，20-50个DET 或者 2-5个RET，20-50个DET 或者 6+个RET，1-19个DET"},
				{Complexity: "high", Weight: 15, Description: "1个RET，51+个DET 或者 2-5个RET，51+个DET 或者 6+个RET，20+个DET"},
			},
			EIF: []FunctionWeight{
				{Complexity: "low", Weight: 5, Description: "1个RET，1-19个DET 或者 2-5个RET，1-19个DET"},
				{Complexity: "average", Weight: 7, Description: "1个RET，20-50个DET 或者 2-5个RET，20-50个DET 或者 6+个RET，1-19个DET"},
				{Complexity: "high", Weight: 10, Description: "1个RET，51+个DET 或者 2-5个RET，51+个DET 或者 6+个RET，20+个DET"},
			},
		},
		TransactionFunctions: TransactionFunctionWeights{
			EI: []FunctionWeight{
				{Complexity: "low", Weight: 3, Description: "1-4个DET，0-1个FTR 或者 5-15个DET，0-1个FTR"},
				{Complexity: "average", Weight: 4, Description: "1-4个DET，2个FTR 或者 5-15个DET，2个FTR 或者 16+个DET，0-1个FTR"},
				{Complexity: "high", Weight: 6, Description: "5-15个DET，3+个FTR 或者 16+个DET，2+个FTR"},
			},
			EO: []FunctionWeight{
				{Complexity: "low", Weight: 4, Description: "1-5个DET，0-1个FTR 或者 6-19个DET，0-1个FTR"},
				{Complexity: "average", Weight: 5, Description: "1-5个DET，2-3个FTR 或者 6-19个DET，2-3个FTR 或者 20+个DET，0-1个FTR"},
				{Complexity: "high", Weight: 7, Description: "6-19个DET，4+个FTR 或者 20+个DET，2+个FTR"},
			},
			EQ: []FunctionWeight{
				{Complexity: "low", Weight: 3, Description: "1-5个DET，0-1个FTR 或者 6-19个DET，0-1个FTR"},
				{Complexity: "average", Weight: 4, Description: "1-5个DET，2-3个FTR 或者 6-19个DET，2-3个FTR 或者 20+个DET，0-1个FTR"},
				{Complexity: "high", Weight: 6, Description: "6-19个DET，4+个FTR 或者 20+个DET，2+个FTR"},
			},
		},
		AdjustmentFactors: []AdjustmentFactor{
			{Name: "数据通信", Value: 3, Description: "应用程序与其用户直接通信的程度"},
			{Name: "分布式数据处理", Value: 3, Description: "分布式数据和处理功能的特征程度"},
			{Name: "性能", Value: 3, Description: "应用程序性能目标（响应时间、吞吐量）对设计的影响"},
			{Name: "系统配置负荷严重", Value: 3, Description: "处理器利用率对应用程序设计的影响"},
			{Name: "交易量", Value: 3, Description: "预期的交易量对应用程序设计的影响"},
			{Name: "在线数据录入", Value: 3, Description: "通过交互式交易进行在线数据录入的数量"},
			{Name: "最终用户效率", Value: 3, Description: "为用户效率而设计的在线功能"},
			{Name: "在线更新", Value: 3, Description: "应用程序提供的在线更新能力"},
			{Name: "复杂的处理", Value: 3, Description: "应用程序中复杂处理的程度"},
			{Name: "可重用性", Value: 3, Description: "应用程序代码专门为在其他应用程序中重用而开发"},
			{Name: "安装容易性", Value: 3, Description: "转换和安装应用程序的难易程度"},
			{Name: "操作容易性", Value: 3, Description: "应用程序的有效和/或自动启动、备份和恢复过程"},
			{Name: "多站点", Value: 3, Description: "应用程序专门为支持多个站点和组织而设计和开发"},
			{Name: "便于变更", Value: 3, Description: "应用程序专门为便于变更而设计、开发和支持"},
		},
		CalculatedValues: CalculatedValues{
			TotalDegreeOfInfluence: 42, // 14个因子 * 3的默认值
			ValueAdjustmentFactor:  1.07, // 42 * 0.01 + 0.65
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// validateEvaluationFactors 验证评估因子配置
func (es *EvaluationService) validateEvaluationFactors(factors *EvaluationFactors) error {
	if factors.EvaluationID == 0 {
		return fmt.Errorf("评估ID不能为空")
	}

	// 验证数据功能权重
	if len(factors.DataFunctions.ILF) != 3 {
		return fmt.Errorf("ILF权重配置必须包含3个复杂度级别")
	}
	if len(factors.DataFunctions.EIF) != 3 {
		return fmt.Errorf("EIF权重配置必须包含3个复杂度级别")
	}

	// 验证事务功能权重
	if len(factors.TransactionFunctions.EI) != 3 {
		return fmt.Errorf("EI权重配置必须包含3个复杂度级别")
	}
	if len(factors.TransactionFunctions.EO) != 3 {
		return fmt.Errorf("EO权重配置必须包含3个复杂度级别")
	}
	if len(factors.TransactionFunctions.EQ) != 3 {
		return fmt.Errorf("EQ权重配置必须包含3个复杂度级别")
	}

	// 验证调整因子
	if len(factors.AdjustmentFactors) != 14 {
		return fmt.Errorf("调整因子必须包含14个GSC特征")
	}

	for _, factor := range factors.AdjustmentFactors {
		if factor.Value < 0 || factor.Value > 5 {
			return fmt.Errorf("调整因子值必须在0-5之间: %s = %d", factor.Name, factor.Value)
		}
	}

	return nil
}
