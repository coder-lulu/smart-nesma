package nesma

import (
	"fmt"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// TokenCalculator 智能Token计算器
type TokenCalculator struct {
	// 不同模型的Token比例配置
	ModelTokenRatios map[string]float64
	// 预留Token比例（用于输出）
	ReservedTokenRatio float64
	// 安全边界比例
	SafetyMargin float64
}

// TokenUsageInfo Token使用情况
type TokenUsageInfo struct {
	InputTokens       int     `json:"input_tokens"`
	OutputTokens      int     `json:"output_tokens"`
	TotalTokens       int     `json:"total_tokens"`
	EstimatedCost     float64 `json:"estimated_cost"`
	ModelUsed         string  `json:"model_used"`
	ContextWindow     int     `json:"context_window"`
	UtilizationRatio  float64 `json:"utilization_ratio"`
}

// BatchTokenInfo 批次Token信息
type BatchTokenInfo struct {
	BatchID           string `json:"batch_id"`
	RequirementsCount int    `json:"requirements_count"`
	EstimatedTokens   int    `json:"estimated_tokens"`
	ActualTokens      int    `json:"actual_tokens"`
	CanFitInBatch     bool   `json:"can_fit_in_batch"`
	RecommendedModel  string `json:"recommended_model"`
}

// NewTokenCalculator 创建Token计算器
func NewTokenCalculator() *TokenCalculator {
	return &TokenCalculator{
		ModelTokenRatios: map[string]float64{
			"gpt-3.5-turbo":        0.75,  // 英文为主，中文token比例更高
			"gpt-4":               0.75,   // 英文为主，中文token比例更高
			"gpt-4-turbo":         0.75,   // 英文为主，中文token比例更高
			"gpt-4o":              0.75,   // 英文为主，中文token比例更高
			"o1-preview":          0.75,   // 英文为主，中文token比例更高
			"o1-mini":             0.75,   // 英文为主，中文token比例更高
			"deepseek-chat":       0.8,    // 对中文友好
			"deepseek-coder":      0.8,    // 对中文友好
			"claude-3-haiku":      0.75,   // 英文为主
			"claude-3-sonnet":     0.75,   // 英文为主
			"claude-3-opus":       0.75,   // 英文为主
			"claude-3.5-sonnet":   0.75,   // 英文为主
		},
		ReservedTokenRatio: 0.3,   // 预留30%给输出
		SafetyMargin:       0.1,   // 10%安全边界
	}
}

// EstimateTokens 估算文本的Token数量
func (tc *TokenCalculator) EstimateTokens(text string, modelName string) int {
	if text == "" {
		return 0
	}

	// 获取模型的Token比例
	ratio, exists := tc.ModelTokenRatios[modelName]
	if !exists {
		ratio = 0.75 // 默认比例
	}

	// 基础字符计数
	charCount := utf8.RuneCountInString(text)
	
	// 根据文本类型调整计算
	tokenCount := tc.calculateTokensByType(text, charCount, ratio)

	global.GVA_LOG.Debug("Token估算", 
		zap.String("模型", modelName),
		zap.Int("字符数", charCount),
		zap.Float64("比例", ratio),
		zap.Int("估算Token", tokenCount))

	return tokenCount
}

// calculateTokensByType 根据文本类型计算Token
func (tc *TokenCalculator) calculateTokensByType(text string, charCount int, ratio float64) int {
	// 基础Token计算
	baseTokens := int(float64(charCount) * ratio)

	// JSON结构化文本的额外Token
	if tc.isJSONLike(text) {
		baseTokens = int(float64(baseTokens) * 1.2) // JSON结构增加20%
	}

	// 代码块的额外Token
	if tc.hasCodeBlocks(text) {
		baseTokens = int(float64(baseTokens) * 1.1) // 代码块增加10%
	}

	// 表格结构的额外Token
	if tc.hasTableStructure(text) {
		baseTokens = int(float64(baseTokens) * 1.15) // 表格结构增加15%
	}

	// 确保最小Token数
	if baseTokens < 1 {
		baseTokens = 1
	}

	return baseTokens
}

// isJSONLike 判断是否为JSON类似结构
func (tc *TokenCalculator) isJSONLike(text string) bool {
	jsonMarkers := []string{`"`, `{`, `}`, `[`, `]`, `:`}
	markerCount := 0
	for _, marker := range jsonMarkers {
		markerCount += strings.Count(text, marker)
	}
	return markerCount > len(text)/20 // 如果结构化字符超过5%，认为是JSON类似
}

// hasCodeBlocks 判断是否包含代码块
func (tc *TokenCalculator) hasCodeBlocks(text string) bool {
	codePatterns := []string{
		"```", "```go", "```python", "```java", "```sql",
		"func ", "class ", "def ", "SELECT ", "INSERT ",
	}
	for _, pattern := range codePatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// hasTableStructure 判断是否包含表格结构
func (tc *TokenCalculator) hasTableStructure(text string) bool {
	tablePatterns := []string{"|", "---|", "表格", "┌", "┐", "└", "┘", "│"}
	for _, pattern := range tablePatterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

// CalculateRequirementTokens 计算需求的Token使用量
func (tc *TokenCalculator) CalculateRequirementTokens(req nesma.NesmaRequirement, modelName string) TokenUsageInfo {
	// 构建完整的需求文本
	var textParts []string
	
	if req.Code != "" {
		textParts = append(textParts, fmt.Sprintf("需求编号: %s", req.Code))
	}
	if req.Title != "" {
		textParts = append(textParts, fmt.Sprintf("需求标题: %s", req.Title))
	}
	if req.Description != "" {
		textParts = append(textParts, fmt.Sprintf("需求描述: %s", req.Description))
	}
	if req.AIDescription != "" {
		textParts = append(textParts, fmt.Sprintf("AI描述: %s", req.AIDescription))
	}
	if req.AcceptanceCriteria != "" {
		textParts = append(textParts, fmt.Sprintf("验收标准: %s", req.AcceptanceCriteria))
	}

	fullText := strings.Join(textParts, "\n")
	inputTokens := tc.EstimateTokens(fullText, modelName)

	// 估算输出Token（基于预期的JSON响应）
	outputTokens := tc.estimateOutputTokens(req, modelName)

	// 计算总Token
	totalTokens := inputTokens + outputTokens

	// 获取模型的上下文窗口
	contextWindow := tc.getContextWindow(modelName)

	// 计算使用率
	utilizationRatio := float64(totalTokens) / float64(contextWindow)

	return TokenUsageInfo{
		InputTokens:      inputTokens,
		OutputTokens:     outputTokens,
		TotalTokens:      totalTokens,
		EstimatedCost:    tc.calculateCost(inputTokens, outputTokens, modelName),
		ModelUsed:        modelName,
		ContextWindow:    contextWindow,
		UtilizationRatio: utilizationRatio,
	}
}

// estimateOutputTokens 估算输出Token数量
func (tc *TokenCalculator) estimateOutputTokens(req nesma.NesmaRequirement, modelName string) int {
	// 基础JSON响应结构的Token
	baseOutputTokens := 500

	// 根据需求复杂度调整
	if len(req.Description) > 1000 {
		baseOutputTokens = int(float64(baseOutputTokens) * 1.5)
	}

	// 如果需要生成详细分析
	if req.Level >= 3 {
		baseOutputTokens = int(float64(baseOutputTokens) * 1.3)
	}

	return baseOutputTokens
}

// calculateCost 计算Token成本
func (tc *TokenCalculator) calculateCost(inputTokens, outputTokens int, modelName string) float64 {
	// 简化的成本计算（实际应该从配置文件读取）
	costPerInput := 0.0001  // 每1000个输入Token的成本
	costPerOutput := 0.0002 // 每1000个输出Token的成本

	// 根据模型调整成本
	switch modelName {
	case "gpt-4", "gpt-4-turbo", "gpt-4o":
		costPerInput *= 10
		costPerOutput *= 10
	case "o1-preview", "o1-mini":
		costPerInput *= 15
		costPerOutput *= 15
	case "claude-3-opus":
		costPerInput *= 8
		costPerOutput *= 8
	case "deepseek-chat":
		costPerInput *= 0.1
		costPerOutput *= 0.1
	}

	return (float64(inputTokens)/1000)*costPerInput + (float64(outputTokens)/1000)*costPerOutput
}

// getContextWindow 获取模型的上下文窗口大小
func (tc *TokenCalculator) getContextWindow(modelName string) int {
	contextWindows := map[string]int{
		"gpt-3.5-turbo":     4096,
		"gpt-4":             8192,
		"gpt-4-turbo":       32768,
		"gpt-4o":            32768,
		"o1-preview":        32768,
		"o1-mini":           32768,
		"deepseek-chat":     32768,
		"deepseek-coder":    32768,
		"claude-3-haiku":    200000,
		"claude-3-sonnet":   200000,
		"claude-3-opus":     200000,
		"claude-3.5-sonnet": 200000,
	}

	if window, exists := contextWindows[modelName]; exists {
		return window
	}
	return 4096 // 默认窗口大小
}

// CalculateBatchCapacity 计算批次容量
func (tc *TokenCalculator) CalculateBatchCapacity(requirements []nesma.NesmaRequirement, modelName string, systemPrompt string) []BatchTokenInfo {
	contextWindow := tc.getContextWindow(modelName)
	
	// 计算系统提示词的Token
	systemTokens := tc.EstimateTokens(systemPrompt, modelName)
	
	// 计算可用Token（减去系统提示词和预留空间）
	availableTokens := int(float64(contextWindow) * (1 - tc.ReservedTokenRatio - tc.SafetyMargin))
	availableTokens -= systemTokens

	global.GVA_LOG.Info("批次容量计算",
		zap.String("模型", modelName),
		zap.Int("上下文窗口", contextWindow),
		zap.Int("系统提示词Token", systemTokens),
		zap.Int("可用Token", availableTokens))

	var batches []BatchTokenInfo
	currentBatch := BatchTokenInfo{
		BatchID:           fmt.Sprintf("batch_%d", len(batches)+1),
		RequirementsCount: 0,
		EstimatedTokens:   0,
		CanFitInBatch:     true,
		RecommendedModel:  modelName,
	}

	for _, req := range requirements {
		reqTokens := tc.CalculateRequirementTokens(req, modelName)
		
		// 检查是否可以添加到当前批次
		if currentBatch.EstimatedTokens+reqTokens.TotalTokens <= availableTokens {
			currentBatch.RequirementsCount++
			currentBatch.EstimatedTokens += reqTokens.TotalTokens
		} else {
			// 当前批次已满，创建新批次
			if currentBatch.RequirementsCount > 0 {
				batches = append(batches, currentBatch)
			}
			
			currentBatch = BatchTokenInfo{
				BatchID:           fmt.Sprintf("batch_%d", len(batches)+1),
				RequirementsCount: 1,
				EstimatedTokens:   reqTokens.TotalTokens,
				CanFitInBatch:     reqTokens.TotalTokens <= availableTokens,
				RecommendedModel:  modelName,
			}
			
			// 如果单个需求超过批次容量，建议使用更大上下文的模型
			if !currentBatch.CanFitInBatch {
				recommendedModel := tc.recommendLargerModel(modelName)
				currentBatch.RecommendedModel = recommendedModel
				global.GVA_LOG.Warn("需求超过批次容量",
					zap.String("需求ID", fmt.Sprintf("%d", req.ID)),
					zap.Int("需求Token", reqTokens.TotalTokens),
					zap.Int("可用Token", availableTokens),
					zap.String("推荐模型", recommendedModel))
			}
		}
	}

	// 添加最后一个批次
	if currentBatch.RequirementsCount > 0 {
		batches = append(batches, currentBatch)
	}

	global.GVA_LOG.Info("批次容量计算完成",
		zap.Int("总需求数", len(requirements)),
		zap.Int("批次数", len(batches)),
		zap.String("模型", modelName))

	return batches
}

// recommendLargerModel 推荐更大上下文的模型
func (tc *TokenCalculator) recommendLargerModel(currentModel string) string {
	modelHierarchy := map[string][]string{
		"gpt-3.5-turbo":   {"gpt-4-turbo", "gpt-4o", "claude-3.5-sonnet"},
		"gpt-4":           {"gpt-4-turbo", "gpt-4o", "claude-3.5-sonnet"},
		"gpt-4-turbo":     {"claude-3.5-sonnet", "claude-3-opus"},
		"gpt-4o":          {"claude-3.5-sonnet", "claude-3-opus"},
		"deepseek-chat":   {"claude-3.5-sonnet", "claude-3-opus"},
		"deepseek-coder":  {"claude-3.5-sonnet", "claude-3-opus"},
		"claude-3-haiku":  {"claude-3-sonnet", "claude-3.5-sonnet"},
		"claude-3-sonnet": {"claude-3.5-sonnet", "claude-3-opus"},
	}

	if recommendations, exists := modelHierarchy[currentModel]; exists && len(recommendations) > 0 {
		return recommendations[0]
	}
	return "claude-3.5-sonnet" // 默认推荐
}

// ValidateTokenLimits 验证Token限制
func (tc *TokenCalculator) ValidateTokenLimits(requirements []nesma.NesmaRequirement, modelName string) (bool, string) {
	totalTokens := 0
	contextWindow := tc.getContextWindow(modelName)

	for _, req := range requirements {
		tokenUsage := tc.CalculateRequirementTokens(req, modelName)
		totalTokens += tokenUsage.TotalTokens
	}

	safeLimit := int(float64(contextWindow) * (1 - tc.SafetyMargin))
	
	if totalTokens > safeLimit {
		return false, fmt.Sprintf("总Token数 %d 超过安全限制 %d，建议使用分批处理", totalTokens, safeLimit)
	}

	return true, ""
}

// GetOptimalBatchSize 获取最优批次大小
func (tc *TokenCalculator) GetOptimalBatchSize(requirements []nesma.NesmaRequirement, modelName string, systemPrompt string) int {
	if len(requirements) == 0 {
		return 0
	}

	// 计算平均Token使用量
	totalTokens := 0
	for _, req := range requirements {
		tokenUsage := tc.CalculateRequirementTokens(req, modelName)
		totalTokens += tokenUsage.TotalTokens
	}

	avgTokensPerReq := totalTokens / len(requirements)
	contextWindow := tc.getContextWindow(modelName)
	systemTokens := tc.EstimateTokens(systemPrompt, modelName)
	
	// 计算可用Token
	availableTokens := int(float64(contextWindow) * (1 - tc.ReservedTokenRatio - tc.SafetyMargin))
	availableTokens -= systemTokens

	// 计算最优批次大小
	optimalBatchSize := availableTokens / avgTokensPerReq
	
	// 确保批次大小在合理范围内
	if optimalBatchSize < 1 {
		optimalBatchSize = 1
	} else if optimalBatchSize > 50 {
		optimalBatchSize = 50 // 限制最大批次大小
	}

	global.GVA_LOG.Info("最优批次大小计算",
		zap.String("模型", modelName),
		zap.Int("平均Token", avgTokensPerReq),
		zap.Int("可用Token", availableTokens),
		zap.Int("最优批次大小", optimalBatchSize))

	return optimalBatchSize
}

// GenerateTokenReport 生成Token使用报告
func (tc *TokenCalculator) GenerateTokenReport(requirements []nesma.NesmaRequirement, modelName string) map[string]interface{} {
	report := map[string]interface{}{
		"model_name":        modelName,
		"context_window":    tc.getContextWindow(modelName),
		"total_requirements": len(requirements),
		"token_statistics":  make(map[string]interface{}),
		"cost_analysis":     make(map[string]interface{}),
		"recommendations":   make([]string, 0),
	}

	// 计算统计信息
	totalInputTokens := 0
	totalOutputTokens := 0
	totalCost := 0.0
	maxTokens := 0
	minTokens := math.MaxInt32

	for _, req := range requirements {
		tokenUsage := tc.CalculateRequirementTokens(req, modelName)
		totalInputTokens += tokenUsage.InputTokens
		totalOutputTokens += tokenUsage.OutputTokens
		totalCost += tokenUsage.EstimatedCost
		
		if tokenUsage.TotalTokens > maxTokens {
			maxTokens = tokenUsage.TotalTokens
		}
		if tokenUsage.TotalTokens < minTokens {
			minTokens = tokenUsage.TotalTokens
		}
	}

	report["token_statistics"] = map[string]interface{}{
		"total_input_tokens":  totalInputTokens,
		"total_output_tokens": totalOutputTokens,
		"total_tokens":        totalInputTokens + totalOutputTokens,
		"avg_tokens_per_req":  (totalInputTokens + totalOutputTokens) / len(requirements),
		"max_tokens_per_req":  maxTokens,
		"min_tokens_per_req":  minTokens,
	}

	report["cost_analysis"] = map[string]interface{}{
		"total_estimated_cost": totalCost,
		"avg_cost_per_req":     totalCost / float64(len(requirements)),
	}

	// 生成推荐
	recommendations := []string{}
	contextWindow := tc.getContextWindow(modelName)
	if totalInputTokens+totalOutputTokens > int(float64(contextWindow)*0.8) {
		recommendations = append(recommendations, "建议使用分批处理以避免超过上下文窗口限制")
	}
	
	if totalCost > 10.0 {
		recommendations = append(recommendations, "成本较高，建议考虑使用更经济的模型")
	}

	report["recommendations"] = recommendations

	return report
}