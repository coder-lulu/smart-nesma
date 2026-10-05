package nesma

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// ContextManager 上下文管理器
type ContextManager struct {
	// 上下文缓存
	contexts map[string]*EvaluationContext
	// 读写锁
	mu sync.RWMutex
	// 上下文过期时间
	contextTTL time.Duration
	// 清理间隔
	cleanupInterval time.Duration
	// 停止信号
	stopChan chan bool
}

// EvaluationContext 评估上下文
type EvaluationContext struct {
	// 基本信息
	EvaluationID    string                 `json:"evaluation_id"`
	ProjectID       uint                   `json:"project_id"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	ExpiresAt       time.Time              `json:"expires_at"`
	
	// 项目信息
	ProjectInfo     *ProjectContextInfo    `json:"project_info"`
	
	// 批次信息
	BatchInfo       *BatchContextInfo      `json:"batch_info"`
	
	// 评估历史
	EvaluationHistory []EvaluationResult   `json:"evaluation_history"`
	
	// 功能类型统计
	FunctionTypeStats map[string]int       `json:"function_type_stats"`
	
	// 复杂度模式
	ComplexityPatterns map[string][]string `json:"complexity_patterns"`
	
	// 质量指标
	QualityMetrics    *QualityMetrics      `json:"quality_metrics"`
	
	// 用户偏好
	UserPreferences   *UserPreferences     `json:"user_preferences"`
	
	// 知识库缓存
	KnowledgeCache    map[string]interface{} `json:"knowledge_cache"`
	
	// 状态信息
	Status            string               `json:"status"`
	
	// 互斥锁
	mu                sync.RWMutex         `json:"-"`
}

// ProjectContextInfo 项目上下文信息
type ProjectContextInfo struct {
	Name         string            `json:"name"`
	Domain       string            `json:"domain"`
	Description  string            `json:"description"`
	Business     string            `json:"business"`
	TechStack    []string          `json:"tech_stack"`
	ProjectType  string            `json:"project_type"`
	Standards    []string          `json:"standards"`
	Constraints  map[string]string `json:"constraints"`
}

// BatchContextInfo 批次上下文信息
type BatchContextInfo struct {
	TotalBatches      int                    `json:"total_batches"`
	CurrentBatch      int                    `json:"current_batch"`
	BatchSize         int                    `json:"batch_size"`
	ProcessedCount    int                    `json:"processed_count"`
	TotalCount        int                    `json:"total_count"`
	BatchProgress     float64                `json:"batch_progress"`
	StartTime         time.Time              `json:"start_time"`
	EstimatedEndTime  time.Time              `json:"estimated_end_time"`
	BatchResults      []BatchResultInfo      `json:"batch_results"`
	ModelUsed         string                 `json:"model_used"`
	TokenUsage        map[string]int         `json:"token_usage"`
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	RequirementID    uint                   `json:"requirement_id"`
	FunctionType     string                 `json:"function_type"`
	ComplexityLevel  string                 `json:"complexity_level"`
	ConfidenceScore  float64                `json:"confidence_score"`
	AFP              float64                `json:"afp"`
	UFP              float64                `json:"ufp"`
	Timestamp        time.Time              `json:"timestamp"`
	BatchID          string                 `json:"batch_id"`
	ModelUsed        string                 `json:"model_used"`
	ProcessingTime   time.Duration          `json:"processing_time"`
	ErrorMsg         string                 `json:"error_msg,omitempty"`
	
	// 原有扩展字段
	Explanation      string                 `json:"explanation,omitempty"`
	DETCount         int                    `json:"det_count,omitempty"`
	RETCount         int                    `json:"ret_count,omitempty"`
	FTRCount         int                    `json:"ftr_count,omitempty"`
	WeightFactor     float64                `json:"weight_factor,omitempty"`
	QualityNotes     string                 `json:"quality_notes,omitempty"`
	
	// 错误处理和降级相关字段
	IsFallback       bool                   `json:"is_fallback,omitempty"`        // 是否为降级结果
	FallbackReason   string                 `json:"fallback_reason,omitempty"`    // 降级原因
	RetryCount       int                    `json:"retry_count,omitempty"`        // 重试次数
	OriginalError    string                 `json:"original_error,omitempty"`     // 原始错误信息
	Confidence       float64                `json:"confidence,omitempty"`         // 置信度（兼容字段）
	Complexity       string                 `json:"complexity,omitempty"`         // 复杂度（兼容字段）
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	OverallAccuracy     float64            `json:"overall_accuracy"`
	ConsistencyScore    float64            `json:"consistency_score"`
	CompletionRate      float64            `json:"completion_rate"`
	ConfidenceAverage   float64            `json:"confidence_average"`
	ErrorRate           float64            `json:"error_rate"`
	ProcessingSpeed     float64            `json:"processing_speed"`
	ModelPerformance    map[string]float64 `json:"model_performance"`
	QualityTrends       []QualityTrend     `json:"quality_trends"`
}

// QualityTrend 质量趋势
type QualityTrend struct {
	Timestamp      time.Time `json:"timestamp"`
	BatchNumber    int       `json:"batch_number"`
	AccuracyScore  float64   `json:"accuracy_score"`
	ConfidenceAvg  float64   `json:"confidence_avg"`
	ProcessingTime float64   `json:"processing_time"`
}

// UserPreferences 用户偏好
type UserPreferences struct {
	PreferredModel      string             `json:"preferred_model"`
	QualityVsSpeed      string             `json:"quality_vs_speed"` // quality/balanced/speed
	AutoApprovalRules   map[string]float64 `json:"auto_approval_rules"`
	NotificationSettings map[string]bool   `json:"notification_settings"`
	CustomPrompts       map[string]string  `json:"custom_prompts"`
}

// BatchResultInfo 批次结果信息
type BatchResultInfo struct {
	BatchID          string            `json:"batch_id"`
	BatchNumber      int               `json:"batch_number"`
	RequirementCount int               `json:"requirement_count"`
	SuccessCount     int               `json:"success_count"`
	FailureCount     int               `json:"failure_count"`
	ProcessingTime   time.Duration     `json:"processing_time"`
	AverageConfidence float64          `json:"average_confidence"`
	TokenUsage       TokenUsage        `json:"token_usage"`
	QualityScore     float64           `json:"quality_score"`
	Status           string            `json:"status"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          time.Time         `json:"end_time"`
}

// NewContextManager 创建上下文管理器
func NewContextManager() *ContextManager {
	cm := &ContextManager{
		contexts:        make(map[string]*EvaluationContext),
		mu:              sync.RWMutex{},
		contextTTL:      24 * time.Hour,   // 24小时过期
		cleanupInterval: 1 * time.Hour,    // 1小时清理一次
		stopChan:        make(chan bool, 1),
	}
	
	// 启动清理goroutine
	go cm.cleanupExpiredContexts()
	
	return cm
}

// CreateContext 创建评估上下文
func (cm *ContextManager) CreateContext(evaluationID string, projectID uint) (*EvaluationContext, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// 检查是否已存在
	if _, exists := cm.contexts[evaluationID]; exists {
		return nil, fmt.Errorf("evaluation context already exists: %s", evaluationID)
	}
	
	// 获取项目信息
	projectInfo, err := cm.loadProjectInfo(projectID)
	if err != nil {
		global.GVA_LOG.Error("加载项目信息失败", zap.Error(err))
		return nil, err
	}
	
	now := time.Now()
	context := &EvaluationContext{
		EvaluationID:       evaluationID,
		ProjectID:          projectID,
		CreatedAt:          now,
		UpdatedAt:          now,
		ExpiresAt:          now.Add(cm.contextTTL),
		ProjectInfo:        projectInfo,
		BatchInfo:          &BatchContextInfo{},
		EvaluationHistory:  make([]EvaluationResult, 0),
		FunctionTypeStats:  make(map[string]int),
		ComplexityPatterns: make(map[string][]string),
		QualityMetrics:     &QualityMetrics{ModelPerformance: make(map[string]float64)},
		UserPreferences:    &UserPreferences{},
		KnowledgeCache:     make(map[string]interface{}),
		Status:             "created",
		mu:                 sync.RWMutex{},
	}
	
	cm.contexts[evaluationID] = context
	
	global.GVA_LOG.Info("创建评估上下文", 
		zap.String("evaluationID", evaluationID),
		zap.Uint("projectID", projectID))
	
	return context, nil
}

// GetContext 获取评估上下文
func (cm *ContextManager) GetContext(evaluationID string) (*EvaluationContext, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	context, exists := cm.contexts[evaluationID]
	if !exists {
		return nil, fmt.Errorf("evaluation context not found: %s", evaluationID)
	}
	
	// 检查是否过期
	if time.Now().After(context.ExpiresAt) {
		delete(cm.contexts, evaluationID)
		return nil, fmt.Errorf("evaluation context expired: %s", evaluationID)
	}
	
	return context, nil
}

// UpdateContext 更新评估上下文
func (cm *ContextManager) UpdateContext(evaluationID string, updateFunc func(*EvaluationContext) error) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	context, exists := cm.contexts[evaluationID]
	if !exists {
		return fmt.Errorf("evaluation context not found: %s", evaluationID)
	}
	
	context.mu.Lock()
	defer context.mu.Unlock()
	
	// 执行更新
	if err := updateFunc(context); err != nil {
		return err
	}
	
	// 更新时间戳
	context.UpdatedAt = time.Now()
	
	return nil
}

// AddEvaluationResult 添加评估结果
func (cm *ContextManager) AddEvaluationResult(evaluationID string, result EvaluationResult) error {
	return cm.UpdateContext(evaluationID, func(ctx *EvaluationContext) error {
		// 添加到历史记录
		ctx.EvaluationHistory = append(ctx.EvaluationHistory, result)
		
		// 更新统计信息
		if result.FunctionType != "" {
			ctx.FunctionTypeStats[result.FunctionType]++
		}
		
		// 更新复杂度模式
		if result.ComplexityLevel != "" {
			patterns := ctx.ComplexityPatterns[result.ComplexityLevel]
			// 这里可以添加模式识别逻辑
			ctx.ComplexityPatterns[result.ComplexityLevel] = patterns
		}
		
		// 更新质量指标
		cm.updateQualityMetrics(ctx, result)
		
		return nil
	})
}

// UpdateBatchInfo 更新批次信息
func (cm *ContextManager) UpdateBatchInfo(evaluationID string, batchInfo BatchContextInfo) error {
	return cm.UpdateContext(evaluationID, func(ctx *EvaluationContext) error {
		ctx.BatchInfo = &batchInfo
		return nil
	})
}

// AddBatchResult 添加批次结果
func (cm *ContextManager) AddBatchResult(evaluationID string, batchResult BatchResultInfo) error {
	return cm.UpdateContext(evaluationID, func(ctx *EvaluationContext) error {
		ctx.BatchInfo.BatchResults = append(ctx.BatchInfo.BatchResults, batchResult)
		ctx.BatchInfo.ProcessedCount += batchResult.SuccessCount
		ctx.BatchInfo.CurrentBatch = batchResult.BatchNumber
		
		// 更新进度
		if ctx.BatchInfo.TotalCount > 0 {
			ctx.BatchInfo.BatchProgress = float64(ctx.BatchInfo.ProcessedCount) / float64(ctx.BatchInfo.TotalCount)
		}
		
		// 更新Token使用统计
		if ctx.BatchInfo.TokenUsage == nil {
			ctx.BatchInfo.TokenUsage = make(map[string]int)
		}
		ctx.BatchInfo.TokenUsage["total"] += batchResult.TokenUsage.TotalTokens
		ctx.BatchInfo.TokenUsage["input"] += batchResult.TokenUsage.PromptTokens
		ctx.BatchInfo.TokenUsage["output"] += batchResult.TokenUsage.CompletionTokens
		
		return nil
	})
}

// GetBatchContext 获取批次上下文
func (cm *ContextManager) GetBatchContext(evaluationID string, batchNumber int) (map[string]interface{}, error) {
	context, err := cm.GetContext(evaluationID)
	if err != nil {
		return nil, err
	}
	
	context.mu.RLock()
	defer context.mu.RUnlock()
	
	batchContext := map[string]interface{}{
		"evaluation_id":       context.EvaluationID,
		"project_info":        context.ProjectInfo,
		"batch_number":        batchNumber,
		"total_batches":       context.BatchInfo.TotalBatches,
		"processed_count":     context.BatchInfo.ProcessedCount,
		"function_type_stats": context.FunctionTypeStats,
		"complexity_patterns": context.ComplexityPatterns,
		"quality_metrics":     context.QualityMetrics,
		"user_preferences":    context.UserPreferences,
		"knowledge_cache":     context.KnowledgeCache,
	}
	
	// 添加相关历史结果
	recentResults := make([]EvaluationResult, 0)
	for _, result := range context.EvaluationHistory {
		if len(recentResults) < 10 { // 最近10个结果
			recentResults = append(recentResults, result)
		}
	}
	batchContext["recent_results"] = recentResults
	
	return batchContext, nil
}

// GenerateContextPrompt 生成上下文提示
func (cm *ContextManager) GenerateContextPrompt(evaluationID string, batchNumber int) (string, error) {
	batchContext, err := cm.GetBatchContext(evaluationID, batchNumber)
	if err != nil {
		return "", err
	}
	
	var promptBuilder strings.Builder
	
	// 项目信息
	if projectInfo, ok := batchContext["project_info"].(*ProjectContextInfo); ok {
		promptBuilder.WriteString("## 项目上下文信息\n")
		promptBuilder.WriteString(fmt.Sprintf("- 项目名称: %s\n", projectInfo.Name))
		promptBuilder.WriteString(fmt.Sprintf("- 业务领域: %s\n", projectInfo.Domain))
		promptBuilder.WriteString(fmt.Sprintf("- 项目类型: %s\n", projectInfo.ProjectType))
		if len(projectInfo.TechStack) > 0 {
			promptBuilder.WriteString(fmt.Sprintf("- 技术栈: %s\n", strings.Join(projectInfo.TechStack, ", ")))
		}
		promptBuilder.WriteString("\n")
	}
	
	// 批次信息
	promptBuilder.WriteString("## 批次处理信息\n")
	promptBuilder.WriteString(fmt.Sprintf("- 当前批次: %d/%d\n", batchNumber, batchContext["total_batches"]))
	promptBuilder.WriteString(fmt.Sprintf("- 已处理需求: %d\n", batchContext["processed_count"]))
	promptBuilder.WriteString("\n")
	
	// 功能类型统计
	if stats, ok := batchContext["function_type_stats"].(map[string]int); ok && len(stats) > 0 {
		promptBuilder.WriteString("## 已识别功能类型统计\n")
		for funcType, count := range stats {
			promptBuilder.WriteString(fmt.Sprintf("- %s: %d个\n", funcType, count))
		}
		promptBuilder.WriteString("\n")
	}
	
	// 质量指标
	if metrics, ok := batchContext["quality_metrics"].(*QualityMetrics); ok {
		promptBuilder.WriteString("## 质量指标参考\n")
		promptBuilder.WriteString(fmt.Sprintf("- 整体准确率: %.2f%%\n", metrics.OverallAccuracy*100))
		promptBuilder.WriteString(fmt.Sprintf("- 一致性评分: %.2f%%\n", metrics.ConsistencyScore*100))
		promptBuilder.WriteString(fmt.Sprintf("- 平均置信度: %.2f%%\n", metrics.ConfidenceAverage*100))
		promptBuilder.WriteString("\n")
	}
	
	// 用户偏好
	if prefs, ok := batchContext["user_preferences"].(*UserPreferences); ok {
		promptBuilder.WriteString("## 用户偏好设置\n")
		promptBuilder.WriteString(fmt.Sprintf("- 首选模型: %s\n", prefs.PreferredModel))
		promptBuilder.WriteString(fmt.Sprintf("- 质量vs速度: %s\n", prefs.QualityVsSpeed))
		promptBuilder.WriteString("\n")
	}
	
	return promptBuilder.String(), nil
}

// CleanupContext 清理上下文
func (cm *ContextManager) CleanupContext(evaluationID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if _, exists := cm.contexts[evaluationID]; !exists {
		return fmt.Errorf("evaluation context not found: %s", evaluationID)
	}
	
	delete(cm.contexts, evaluationID)
	
	global.GVA_LOG.Info("清理评估上下文", zap.String("evaluationID", evaluationID))
	
	return nil
}

// GetContextStatus 获取上下文状态
func (cm *ContextManager) GetContextStatus(evaluationID string) (map[string]interface{}, error) {
	context, err := cm.GetContext(evaluationID)
	if err != nil {
		return nil, err
	}
	
	context.mu.RLock()
	defer context.mu.RUnlock()
	
	status := map[string]interface{}{
		"evaluation_id":    context.EvaluationID,
		"project_id":       context.ProjectID,
		"status":           context.Status,
		"created_at":       context.CreatedAt,
		"updated_at":       context.UpdatedAt,
		"expires_at":       context.ExpiresAt,
		"batch_progress":   context.BatchInfo.BatchProgress,
		"processed_count":  context.BatchInfo.ProcessedCount,
		"total_count":      context.BatchInfo.TotalCount,
		"current_batch":    context.BatchInfo.CurrentBatch,
		"total_batches":    context.BatchInfo.TotalBatches,
		"quality_score":    context.QualityMetrics.OverallAccuracy,
		"results_count":    len(context.EvaluationHistory),
	}
	
	return status, nil
}

// loadProjectInfo 加载项目信息
func (cm *ContextManager) loadProjectInfo(projectID uint) (*ProjectContextInfo, error) {
	var project nesma.NesmaProject
	if err := global.GVA_DB.First(&project, projectID).Error; err != nil {
		return nil, err
	}
	
	projectInfo := &ProjectContextInfo{
		Name:        project.Name,
		Domain:      project.Domain,
		Description: project.Description,
		Business:    "general", // 默认业务类型
		ProjectType: "web",     // 默认项目类型
		TechStack:   make([]string, 0),
		Standards:   make([]string, 0),
		Constraints: make(map[string]string),
	}
	
	// 可以从项目配置或其他表中加载更多信息
	// 这里简化处理
	projectInfo.Standards = append(projectInfo.Standards, "NESMA 2.2")
	
	return projectInfo, nil
}

// updateQualityMetrics 更新质量指标
func (cm *ContextManager) updateQualityMetrics(ctx *EvaluationContext, result EvaluationResult) {
	if ctx.QualityMetrics == nil {
		ctx.QualityMetrics = &QualityMetrics{
			ModelPerformance: make(map[string]float64),
		}
	}
	
	// 更新整体准确率（简化计算）
	successCount := 0
	totalCount := len(ctx.EvaluationHistory)
	
	for _, r := range ctx.EvaluationHistory {
		if r.ErrorMsg == "" && r.ConfidenceScore > 0.5 {
			successCount++
		}
	}
	
	if totalCount > 0 {
		ctx.QualityMetrics.OverallAccuracy = float64(successCount) / float64(totalCount)
	}
	
	// 更新平均置信度
	confidenceSum := 0.0
	for _, r := range ctx.EvaluationHistory {
		confidenceSum += r.ConfidenceScore
	}
	if totalCount > 0 {
		ctx.QualityMetrics.ConfidenceAverage = confidenceSum / float64(totalCount)
	}
	
	// 更新模型性能
	if result.ModelUsed != "" {
		ctx.QualityMetrics.ModelPerformance[result.ModelUsed] = result.ConfidenceScore
	}
	
	// 添加质量趋势
	trend := QualityTrend{
		Timestamp:      time.Now(),
		BatchNumber:    ctx.BatchInfo.CurrentBatch,
		AccuracyScore:  ctx.QualityMetrics.OverallAccuracy,
		ConfidenceAvg:  ctx.QualityMetrics.ConfidenceAverage,
		ProcessingTime: float64(result.ProcessingTime.Milliseconds()),
	}
	ctx.QualityMetrics.QualityTrends = append(ctx.QualityMetrics.QualityTrends, trend)
}

// cleanupExpiredContexts 清理过期上下文
func (cm *ContextManager) cleanupExpiredContexts() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cm.performCleanup()
		case <-cm.stopChan:
			return
		}
	}
}

// performCleanup 执行清理
func (cm *ContextManager) performCleanup() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	now := time.Now()
	expiredContexts := make([]string, 0)
	
	for id, context := range cm.contexts {
		if now.After(context.ExpiresAt) {
			expiredContexts = append(expiredContexts, id)
		}
	}
	
	for _, id := range expiredContexts {
		delete(cm.contexts, id)
	}
	
	if len(expiredContexts) > 0 {
		global.GVA_LOG.Info("清理过期上下文", zap.Int("count", len(expiredContexts)))
	}
}

// Stop 停止上下文管理器
func (cm *ContextManager) Stop() {
	close(cm.stopChan)
}

// GetAllContexts 获取所有上下文状态（用于监控）
func (cm *ContextManager) GetAllContexts() []map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	contexts := make([]map[string]interface{}, 0, len(cm.contexts))
	
	for _, ctx := range cm.contexts {
		ctx.mu.RLock()
		contextInfo := map[string]interface{}{
			"evaluation_id":   ctx.EvaluationID,
			"project_id":      ctx.ProjectID,
			"status":          ctx.Status,
			"created_at":      ctx.CreatedAt,
			"batch_progress":  ctx.BatchInfo.BatchProgress,
			"processed_count": ctx.BatchInfo.ProcessedCount,
			"quality_score":   ctx.QualityMetrics.OverallAccuracy,
		}
		ctx.mu.RUnlock()
		contexts = append(contexts, contextInfo)
	}
	
	return contexts
}

// SerializeContext 序列化上下文
func (cm *ContextManager) SerializeContext(evaluationID string) ([]byte, error) {
	context, err := cm.GetContext(evaluationID)
	if err != nil {
		return nil, err
	}
	
	context.mu.RLock()
	defer context.mu.RUnlock()
	
	return json.Marshal(context)
}

// DeserializeContext 反序列化上下文
func (cm *ContextManager) DeserializeContext(data []byte) (*EvaluationContext, error) {
	var context EvaluationContext
	if err := json.Unmarshal(data, &context); err != nil {
		return nil, err
	}
	
	// 重新设置锁
	context.mu = sync.RWMutex{}
	
	// 检查过期时间
	if time.Now().After(context.ExpiresAt) {
		return nil, fmt.Errorf("context expired")
	}
	
	return &context, nil
}