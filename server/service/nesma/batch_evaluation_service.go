package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// BatchEvaluationService 批处理评估服务
type BatchEvaluationService struct {
	// 依赖的服务
	db               *gorm.DB
	aiService        AIService
	tokenCalculator  *TokenCalculator
	contextManager   *ContextManager
	knowledgeService *KnowledgeService
	
	// 智能分片算法
	shardingAlgorithm *IntelligentShardingAlgorithm
	
	// 错误处理和重试
	errorHandler      *ErrorHandler
	circuitBreakerManager *CircuitBreakerManager
	
	// 配置
	concurrentConfig ConcurrentConfig
	
	// 运行时管理
	activeEvaluations map[string]*BatchEvaluationTask
	mu               sync.RWMutex
	
	// 工作协程管理
	workerPool       *WorkerPool
	shutdownChan     chan struct{}
	
	// 指标统计
	metrics          *BatchEvaluationMetrics
}

// 全局服务实例
var batchEvaluationServiceInstance *BatchEvaluationService
var batchEvaluationServiceOnce sync.Once

// BatchEvaluationTask 批处理评估任务
type BatchEvaluationTask struct {
	// 任务信息
	ID               string                          `json:"id"`
	EvaluationID     string                          `json:"evaluation_id"`
	ProjectID        uint                            `json:"project_id"`
	CycleID          uint                            `json:"cycle_id"`
	SourceVersionID  uint                            `json:"source_version_id"`
	TargetVersionID  uint                            `json:"target_version_id"`
	
	// 任务配置
	Config           *BatchEvaluationConfig          `json:"config"`
	
	// 批次信息
	Batches          []BatchInfo                     `json:"batches"`
	TotalBatches     int                             `json:"total_batches"`
	CompletedBatches int                             `json:"completed_batches"`
	CurrentBatch     int                             `json:"current_batch"`
	
	// 执行状态
	Status           string                          `json:"status"` // pending/running/completed/failed/cancelled
	Progress         float64                         `json:"progress"`
	StartTime        time.Time                       `json:"start_time"`
	EndTime          *time.Time                      `json:"end_time"`
	CreatedAt        time.Time                       `json:"created_at"`
	EstimatedEndTime *time.Time                      `json:"estimated_end_time"`
	EstimatedDuration time.Duration                  `json:"estimated_duration"`
	ProcessingTime   time.Duration                   `json:"processing_time"`
	StatusMessage    string                          `json:"status_message"`
	
	// 统计信息
	TotalRequirements int                            `json:"total_requirements"`
	ProcessedCount    int                            `json:"processed_count"`
	SuccessCount      int                            `json:"success_count"`
	FailedCount       int                            `json:"failed_count"`
	
	// 结果
	Results          []EvaluationResult              `json:"results"`
	Summary          *EvaluationSummary              `json:"summary"`
	QualityMetrics   *QualityMetrics                 `json:"quality_metrics"`
	TokenUsage       *TokenUsage                     `json:"token_usage"`
	
	// 错误信息
	ErrorMessage     string                          `json:"error_message"`
	
	// 同步控制
	mu               sync.RWMutex                    `json:"-"`
	cancelFunc       context.CancelFunc              `json:"-"`
	
	// 进度回调
	progressCallback func(progress BatchProgress)     `json:"-"`
}

// BatchEvaluationConfig 批处理评估配置
type BatchEvaluationConfig struct {
	// AI模型配置
	AIModel            string                 `json:"ai_model"`
	BackupModel        string                 `json:"backup_model"`
	MaxRetries         int                    `json:"max_retries"`
	TimeoutSeconds     int                    `json:"timeout_seconds"`
	
	// 批处理配置
	BatchStrategy      string                 `json:"batch_strategy"`      // optimal/fixed/adaptive
	FixedBatchSize     int                    `json:"fixed_batch_size"`
	MaxBatchSize       int                    `json:"max_batch_size"`
	MinBatchSize       int                    `json:"min_batch_size"`
	
	// 评估配置
	EvaluationTypes    []string               `json:"evaluation_types"`    // function_type/complexity/afp/ufp
	QualityThreshold   float64                `json:"quality_threshold"`
	ConsistencyCheck   bool                   `json:"consistency_check"`
	
	// 上下文配置
	IncludeHistory     bool                   `json:"include_history"`
	ContextLevels      []int                  `json:"context_levels"`      // 需要包含的需求层级
	KnowledgeCache     bool                   `json:"knowledge_cache"`
	
	// 输出配置
	OutputFormat       string                 `json:"output_format"`       // json/structured
	DetailLevel        string                 `json:"detail_level"`        // basic/detailed/full
	IncludeExplanation bool                   `json:"include_explanation"`
}

// BatchInfo 批次信息
type BatchInfo struct {
	BatchID          string                 `json:"batch_id"`
	BatchNumber      int                    `json:"batch_number"`
	Requirements     []nesma.NesmaRequirement `json:"requirements"`
	EstimatedTokens  int                    `json:"estimated_tokens"`
	ActualTokens     int                    `json:"actual_tokens"`
	Status           string                 `json:"status"`
	StartTime        *time.Time             `json:"start_time"`
	EndTime          *time.Time             `json:"end_time"`
	ModelUsed        string                 `json:"model_used"`
	RetryCount       int                    `json:"retry_count"`
	ErrorMessage     string                 `json:"error_message"`
}

// EvaluationSummary 评估总结
type EvaluationSummary struct {
	TotalRequirements   int                      `json:"total_requirements"`
	ProcessedCount      int                      `json:"processed_count"`
	SuccessRate         float64                  `json:"success_rate"`
	AverageConfidence   float64                  `json:"average_confidence"`
	FunctionTypeStats   map[string]int           `json:"function_type_stats"`
	ComplexityStats     map[string]int           `json:"complexity_stats"`
	TotalAFP            float64                  `json:"total_afp"`
	TotalUFP            float64                  `json:"total_ufp"`
	QualityScore        float64                  `json:"quality_score"`
	ProcessingTime      time.Duration            `json:"processing_time"`
	TokenUsage          TokenUsage               `json:"token_usage"`
	ModelPerformance    map[string]ModelStats    `json:"model_performance"`
	Recommendations     []string                 `json:"recommendations"`
}

// ModelStats 模型统计
type ModelStats struct {
	CallCount        int           `json:"call_count"`
	SuccessCount     int           `json:"success_count"`
	AvgResponseTime  time.Duration `json:"avg_response_time"`
	AvgConfidence    float64       `json:"avg_confidence"`
	TokenUsage       TokenUsage    `json:"token_usage"`
	ErrorRate        float64       `json:"error_rate"`
}

// BatchProgress 批处理进度
type BatchProgress struct {
	EvaluationID     string    `json:"evaluation_id"`
	TotalBatches     int       `json:"total_batches"`
	CompletedBatches int       `json:"completed_batches"`
	CurrentBatch     int       `json:"current_batch"`
	TotalRequirements int      `json:"total_requirements"`
	ProcessedCount   int       `json:"processed_count"`
	SuccessCount     int       `json:"success_count"`
	FailedCount      int       `json:"failed_count"`
	Progress         float64   `json:"progress"`
	EstimatedTimeLeft time.Duration `json:"estimated_time_left"`
	Status           string    `json:"status"`
	CurrentOperation string    `json:"current_operation"`
}

// WorkerPool 工作协程池
type WorkerPool struct {
	workerCount  int
	taskChan     chan BatchWorkerTask
	resultChan   chan BatchWorkerResult
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// BatchWorkerTask 批处理工作任务
type BatchWorkerTask struct {
	EvaluationID string
	BatchInfo    BatchInfo
	Context      map[string]interface{}
	Config       *BatchEvaluationConfig
}

// BatchWorkerResult 批处理工作结果
type BatchWorkerResult struct {
	EvaluationID string
	BatchID      string
	Results      []EvaluationResult
	Error        error
	TokenUsage   TokenUsage
	Duration     time.Duration
	ModelUsed    string
}

// BatchEvaluationMetrics 批处理评估指标
type BatchEvaluationMetrics struct {
	TotalEvaluations    int64               `json:"total_evaluations"`
	CompletedEvaluations int64              `json:"completed_evaluations"`
	FailedEvaluations   int64               `json:"failed_evaluations"`
	TotalRequirements   int64               `json:"total_requirements"`
	TotalTokens         int64               `json:"total_tokens"`
	TotalCost           float64             `json:"total_cost"`
	AverageProcessingTime time.Duration     `json:"average_processing_time"`
	ModelUsageStats     map[string]int64    `json:"model_usage_stats"`
	mu                  sync.RWMutex        `json:"-"`
}

// NewBatchEvaluationService 创建批处理评估服务
func NewBatchEvaluationService(db *gorm.DB, aiService AIService, knowledgeService *KnowledgeService) *BatchEvaluationService {
	tokenCalculator := NewTokenCalculator()
	contextManager := NewContextManager()
	
	// 初始化错误处理和断路器
	errorHandler := NewErrorHandler()
	circuitBreakerManager := NewCircuitBreakerManager()
	
	service := &BatchEvaluationService{
		db:               db,
		aiService:        aiService,
		tokenCalculator:  tokenCalculator,
		contextManager:   contextManager,
		knowledgeService: knowledgeService,
		errorHandler:     errorHandler,
		circuitBreakerManager: circuitBreakerManager,
		concurrentConfig: GetConcurrentConfig(),
		activeEvaluations: make(map[string]*BatchEvaluationTask),
		mu:               sync.RWMutex{},
		shutdownChan:     make(chan struct{}),
		metrics:          &BatchEvaluationMetrics{
			ModelUsageStats: make(map[string]int64),
		},
	}
	
	// 初始化智能分片算法
	service.shardingAlgorithm = NewIntelligentShardingAlgorithm(tokenCalculator, contextManager, knowledgeService)
	
	// 初始化工作协程池
	service.workerPool = NewWorkerPool(service.concurrentConfig.MaxEvaluationWorkers)
	
	// 设置断路器
	service.setupCircuitBreakers()
	
	// 启动工作协程
	go service.startWorkers()
	
	global.GVA_LOG.Info("批处理评估服务初始化完成",
		zap.Int("最大工作协程数", service.concurrentConfig.MaxEvaluationWorkers))
	
	return service
}

// CreateBatchEvaluation 创建批处理评估任务
func (s *BatchEvaluationService) CreateBatchEvaluation(projectID, cycleID, sourceVersionID uint, config *BatchEvaluationConfig) (*BatchEvaluationTask, error) {
	// 生成评估ID
	evaluationID := fmt.Sprintf("eval_%d_%d_%d", projectID, cycleID, time.Now().UnixNano())
	
	// 创建上下文
	_, err := s.contextManager.CreateContext(evaluationID, projectID)
	if err != nil {
		return nil, fmt.Errorf("创建上下文失败: %w", err)
	}
	
	// 获取需求列表
	requirements, err := s.getRequirements(cycleID, sourceVersionID)
	if err != nil {
		return nil, fmt.Errorf("获取需求列表失败: %w", err)
	}
	
	if len(requirements) == 0 {
		return nil, fmt.Errorf("没有找到需要评估的需求")
	}
	
	// 创建目标版本
	targetVersionID, err := s.createTargetVersion(cycleID, sourceVersionID)
	if err != nil {
		return nil, fmt.Errorf("创建目标版本失败: %w", err)
	}
	
	// 计算批次分配
	batches, err := s.calculateBatches(requirements, config)
	if err != nil {
		return nil, fmt.Errorf("计算批次失败: %w", err)
	}
	
	// 创建批处理任务
	now := time.Now()
	task := &BatchEvaluationTask{
		ID:                evaluationID,
		EvaluationID:      evaluationID,
		ProjectID:         projectID,
		CycleID:           cycleID,
		SourceVersionID:   sourceVersionID,
		TargetVersionID:   targetVersionID,
		Config:            config,
		Batches:           batches,
		TotalBatches:      len(batches),
		CompletedBatches:  0,
		CurrentBatch:      0,
		Status:            "pending",
		Progress:          0.0,
		StartTime:         now,
		CreatedAt:         now,
		EstimatedDuration: time.Duration(len(batches) * 5) * time.Minute, // 预估5分钟/批次
		StatusMessage:     "任务已创建，等待启动",
		TotalRequirements: len(requirements),
		ProcessedCount:    0,
		SuccessCount:      0,
		FailedCount:       0,
		Results:           make([]EvaluationResult, 0),
		mu:                sync.RWMutex{},
	}
	
	// 保存到活跃任务列表
	s.mu.Lock()
	s.activeEvaluations[evaluationID] = task
	s.mu.Unlock()
	
	global.GVA_LOG.Info("创建批处理评估任务",
		zap.String("evaluationID", evaluationID),
		zap.Uint("projectID", projectID),
		zap.Int("需求数量", len(requirements)),
		zap.Int("批次数量", len(batches)))
	
	return task, nil
}

// StartBatchEvaluation 开始批处理评估
func (s *BatchEvaluationService) StartBatchEvaluation(evaluationID string, progressCallback func(BatchProgress)) error {
	s.mu.RLock()
	task, exists := s.activeEvaluations[evaluationID]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("评估任务不存在: %s", evaluationID)
	}
	
	task.mu.Lock()
	if task.Status != "pending" {
		task.mu.Unlock()
		return fmt.Errorf("任务状态不正确: %s", task.Status)
	}
	
	task.Status = "running"
	task.StartTime = time.Now()
	task.progressCallback = progressCallback
	
	// 创建取消上下文
	ctx, cancel := context.WithCancel(context.Background())
	task.cancelFunc = cancel
	task.mu.Unlock()
	
	// 异步执行批处理
	go s.executeBatchEvaluation(ctx, task)
	
	global.GVA_LOG.Info("开始批处理评估",
		zap.String("evaluationID", evaluationID),
		zap.Int("总批次数", task.TotalBatches))
	
	return nil
}

// executeBatchEvaluation 执行批处理评估
func (s *BatchEvaluationService) executeBatchEvaluation(ctx context.Context, task *BatchEvaluationTask) {
	defer func() {
		if r := recover(); r != nil {
			s.handleTaskError(task, fmt.Errorf("批处理执行异常: %v", r))
		}
	}()
	
	// 更新指标
	s.metrics.mu.Lock()
	s.metrics.TotalEvaluations++
	s.metrics.mu.Unlock()
	
	// 按批次顺序处理
	for i, batch := range task.Batches {
		select {
		case <-ctx.Done():
			s.handleTaskCancellation(task)
			return
		default:
		}
		
		// 更新当前批次状态
		task.mu.Lock()
		task.Batches[i].Status = "running"
		now := time.Now()
		task.Batches[i].StartTime = &now
		task.mu.Unlock()
		
		// 发送进度回调
		s.sendProgress(task, fmt.Sprintf("处理批次 %d/%d", i+1, task.TotalBatches))
		
		// 处理批次
		err := s.processBatch(ctx, task, &task.Batches[i])
		
		task.mu.Lock()
		if err != nil {
			task.Batches[i].Status = "failed"
			task.Batches[i].ErrorMessage = err.Error()
			task.FailedCount++
			
			global.GVA_LOG.Error("批次处理失败",
				zap.String("evaluationID", task.EvaluationID),
				zap.String("batchID", batch.BatchID),
				zap.Error(err))
			
			// 根据配置决定是否继续
			if !s.shouldContinueOnBatchFailure(task.Config) {
				task.mu.Unlock()
				s.handleTaskError(task, fmt.Errorf("批次处理失败: %w", err))
				return
			}
		} else {
			task.Batches[i].Status = "completed"
			task.CompletedBatches++
		}
		
		endTime := time.Now()
		task.Batches[i].EndTime = &endTime
		task.mu.Unlock()
		
		// 发送进度回调
		s.sendProgress(task, fmt.Sprintf("批次 %d/%d 完成", i+1, task.TotalBatches))
	}
	
	// 所有批次处理完成，开始最终评估
	s.sendProgress(task, "开始最终评估")
	err := s.performFinalEvaluation(ctx, task)
	if err != nil {
		s.handleTaskError(task, fmt.Errorf("最终评估失败: %w", err))
		return
	}
	
	// 完成任务
	s.completeTask(task)
}

// processBatch 处理单个批次
func (s *BatchEvaluationService) processBatch(ctx context.Context, task *BatchEvaluationTask, batch *BatchInfo) error {
	// 获取批次上下文
	batchContext, err := s.contextManager.GetBatchContext(task.EvaluationID, batch.BatchNumber)
	if err != nil {
		return fmt.Errorf("获取批次上下文失败: %w", err)
	}
	
	// 构建批次任务
	workerTask := BatchWorkerTask{
		EvaluationID: task.EvaluationID,
		BatchInfo:    *batch,
		Context:      batchContext,
		Config:       task.Config,
	}
	
	// 发送到工作协程池
	select {
	case s.workerPool.taskChan <- workerTask:
	case <-ctx.Done():
		return ctx.Err()
	}
	
	// 等待结果
	select {
	case result := <-s.workerPool.resultChan:
		if result.Error != nil {
			return result.Error
		}
		
		// 更新批次结果
		batch.ActualTokens = result.TokenUsage.TotalTokens
		batch.ModelUsed = result.ModelUsed
		
		// 更新任务结果
		task.mu.Lock()
		task.Results = append(task.Results, result.Results...)
		task.ProcessedCount += len(result.Results)
		
		// 统计成功数量
		for _, r := range result.Results {
			if r.ErrorMsg == "" {
				task.SuccessCount++
			}
		}
		task.mu.Unlock()
		
		// 更新上下文管理器
		for _, r := range result.Results {
			s.contextManager.AddEvaluationResult(task.EvaluationID, r)
		}
		
		return nil
		
	case <-ctx.Done():
		return ctx.Err()
	}
}

// performFinalEvaluation 执行最终评估
func (s *BatchEvaluationService) performFinalEvaluation(ctx context.Context, task *BatchEvaluationTask) error {
	// 构建最终评估提示词
	finalPrompt, err := s.buildFinalEvaluationPrompt(task)
	if err != nil {
		return fmt.Errorf("构建最终评估提示词失败: %w", err)
	}
	
	// 调用AI进行最终评估
	response, err := s.aiService.GenerateText(ctx, finalPrompt, &AIConfig{
		Model:     task.Config.AIModel,
		MaxTokens: 8192,
	})
	if err != nil {
		return fmt.Errorf("AI最终评估失败: %w", err)
	}
	
	// 解析评估结果
	summary, err := s.parseFinalEvaluationResult(response.Text)
	if err != nil {
		return fmt.Errorf("解析最终评估结果失败: %w", err)
	}
	
	// 更新任务摘要
	task.mu.Lock()
	task.Summary = summary
	task.mu.Unlock()
	
	global.GVA_LOG.Info("最终评估完成",
		zap.String("evaluationID", task.EvaluationID),
		zap.Float64("质量评分", summary.QualityScore),
		zap.Float64("成功率", summary.SuccessRate))
	
	return nil
}

// buildFinalEvaluationPrompt 构建最终评估提示词
func (s *BatchEvaluationService) buildFinalEvaluationPrompt(task *BatchEvaluationTask) (string, error) {
	var promptBuilder strings.Builder
	
	// 添加任务摘要
	promptBuilder.WriteString("## 批次评估任务摘要\n")
	promptBuilder.WriteString(fmt.Sprintf("- 项目ID: %d\n", task.ProjectID))
	promptBuilder.WriteString(fmt.Sprintf("- 总需求数: %d\n", task.TotalRequirements))
	promptBuilder.WriteString(fmt.Sprintf("- 处理批次数: %d\n", task.TotalBatches))
	promptBuilder.WriteString(fmt.Sprintf("- 处理成功数: %d\n", task.SuccessCount))
	promptBuilder.WriteString("\n")
	
	// 添加功能类型统计
	functionStats := make(map[string]int)
	complexityStats := make(map[string]int)
	
	for _, result := range task.Results {
		if result.FunctionType != "" {
			functionStats[result.FunctionType]++
		}
		if result.ComplexityLevel != "" {
			complexityStats[result.ComplexityLevel]++
		}
	}
	
	if len(functionStats) > 0 {
		promptBuilder.WriteString("## 功能类型分布\n")
		for funcType, count := range functionStats {
			promptBuilder.WriteString(fmt.Sprintf("- %s: %d个\n", funcType, count))
		}
		promptBuilder.WriteString("\n")
	}
	
	if len(complexityStats) > 0 {
		promptBuilder.WriteString("## 复杂度分布\n")
		for complexity, count := range complexityStats {
			promptBuilder.WriteString(fmt.Sprintf("- %s: %d个\n", complexity, count))
		}
		promptBuilder.WriteString("\n")
	}
	
	// 添加最终评估指令
	promptBuilder.WriteString("## 最终评估指令\n")
	promptBuilder.WriteString("请基于以上批次评估结果，提供以下JSON格式的最终评估:\n")
	promptBuilder.WriteString("```json\n")
	promptBuilder.WriteString("{\n")
	promptBuilder.WriteString("  \"quality_score\": 0.0,\n")
	promptBuilder.WriteString("  \"consistency_score\": 0.0,\n")
	promptBuilder.WriteString("  \"accuracy_assessment\": \"评估准确性分析\",\n")
	promptBuilder.WriteString("  \"recommendations\": [\"建议1\", \"建议2\"],\n")
	promptBuilder.WriteString("  \"improvement_areas\": [\"改进点1\", \"改进点2\"],\n")
	promptBuilder.WriteString("  \"overall_assessment\": \"整体评估总结\"\n")
	promptBuilder.WriteString("}\n")
	promptBuilder.WriteString("```\n")
	
	return promptBuilder.String(), nil
}

// parseFinalEvaluationResult 解析最终评估结果
func (s *BatchEvaluationService) parseFinalEvaluationResult(response string) (*EvaluationSummary, error) {
	// 尝试提取JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("响应中未找到JSON格式")
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var finalResult struct {
		QualityScore       float64  `json:"quality_score"`
		ConsistencyScore   float64  `json:"consistency_score"`
		AccuracyAssessment string   `json:"accuracy_assessment"`
		Recommendations    []string `json:"recommendations"`
		ImprovementAreas   []string `json:"improvement_areas"`
		OverallAssessment  string   `json:"overall_assessment"`
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &finalResult); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}
	
	// 构建摘要
	summary := &EvaluationSummary{
		QualityScore:     finalResult.QualityScore,
		Recommendations:  finalResult.Recommendations,
	}
	
	return summary, nil
}

// calculateBatches 计算批次分配（使用智能分片算法）
func (s *BatchEvaluationService) calculateBatches(requirements []nesma.NesmaRequirement, config *BatchEvaluationConfig) ([]BatchInfo, error) {
	// 构建分片配置
	shardingConfig := s.buildShardingConfig(config)
	
	// 执行智能分片
	shardingResult, err := s.shardingAlgorithm.ExecuteSharding(requirements, shardingConfig)
	if err != nil {
		global.GVA_LOG.Error("智能分片失败", zap.Error(err))
		// 降级到基础分片方法
		return s.calculateBatchesBasic(requirements, config)
	}
	
	// 转换分片结果为BatchInfo
	var batches []BatchInfo
	for _, shard := range shardingResult.Shards {
		batch := BatchInfo{
			BatchID:         shard.ShardID,
			BatchNumber:     shard.ShardNumber,
			Requirements:    shard.Requirements,
			EstimatedTokens: shard.EstimatedTokens,
			Status:          "pending",
			ModelUsed:       shard.RecommendedModel,
		}
		batches = append(batches, batch)
	}
	
	global.GVA_LOG.Info("智能分片完成", 
		zap.String("strategy", string(shardingResult.Strategy)),
		zap.Int("total_shards", shardingResult.TotalShards),
		zap.Float64("balance_score", shardingResult.BalanceScore),
		zap.Float64("quality_score", shardingResult.QualityScore),
		zap.Duration("processing_time", shardingResult.ProcessingTime))
	
	return batches, nil
}

// buildShardingConfig 构建分片配置
func (s *BatchEvaluationService) buildShardingConfig(config *BatchEvaluationConfig) *ShardingConfig {
	return &ShardingConfig{
		Strategy:            ShardingStrategy(config.BatchStrategy),
		MaxTokensPerShard:   8192, // 默认最大Token数
		MinTokensPerShard:   1000, // 默认最小Token数
		MaxRequirements:     config.MaxBatchSize,
		MinRequirements:     config.MinBatchSize,
		SimilarityThreshold: 0.7,
		ComplexityBalance:   config.ConsistencyCheck,
		ContextAwareness:    config.IncludeHistory,
		TokenEfficiency:     0.8,
		QualityPriority:     config.QualityThreshold,
		PerformancePriority: 0.7,
	}
}

// calculateBatchesBasic 基础批次计算（降级方法）
func (s *BatchEvaluationService) calculateBatchesBasic(requirements []nesma.NesmaRequirement, config *BatchEvaluationConfig) ([]BatchInfo, error) {
	// 构建系统提示词
	systemPrompt := s.buildSystemPrompt(config)
	
	// 计算批次容量
	batchTokenInfos := s.tokenCalculator.CalculateBatchCapacity(requirements, config.AIModel, systemPrompt)
	
	var batches []BatchInfo
	currentReqIndex := 0
	
	for i, tokenInfo := range batchTokenInfos {
		if currentReqIndex >= len(requirements) {
			break
		}
		
		// 获取当前批次的需求
		endIndex := currentReqIndex + tokenInfo.RequirementsCount
		if endIndex > len(requirements) {
			endIndex = len(requirements)
		}
		
		batchRequirements := requirements[currentReqIndex:endIndex]
		
		batch := BatchInfo{
			BatchID:         tokenInfo.BatchID,
			BatchNumber:     i + 1,
			Requirements:    batchRequirements,
			EstimatedTokens: tokenInfo.EstimatedTokens,
			Status:          "pending",
			ModelUsed:       tokenInfo.RecommendedModel,
		}
		
		batches = append(batches, batch)
		currentReqIndex = endIndex
	}
	
	return batches, nil
}

// GetRequirements 获取需求列表
func (s *BatchEvaluationService) GetRequirements(cycleID, sourceVersionID uint) ([]nesma.NesmaRequirement, error) {
	var requirements []nesma.NesmaRequirement
	
	// 查询需求列表
	query := s.db.Where("cycle_id = ?", cycleID)
	if sourceVersionID > 0 {
		query = query.Where("version_id = ?", sourceVersionID)
	}
	
	err := query.Find(&requirements).Error
	if err != nil {
		return nil, fmt.Errorf("查询需求列表失败: %w", err)
	}
	
	return requirements, nil
}

// TestSharding 测试分片算法
func (s *BatchEvaluationService) TestSharding(requirements []nesma.NesmaRequirement, strategy string, config map[string]interface{}) (*ShardingResult, error) {
	// 构建分片配置
	shardingConfig := &ShardingConfig{
		Strategy:            ShardingStrategy(strategy),
		MaxTokensPerShard:   8192,
		MinTokensPerShard:   1000,
		MaxRequirements:     20,
		MinRequirements:     2,
		SimilarityThreshold: 0.7,
		ComplexityBalance:   true,
		ContextAwareness:    true,
		TokenEfficiency:     0.8,
		QualityPriority:     0.7,
		PerformancePriority: 0.6,
	}
	
	// 从配置中覆盖参数
	if maxTokens, ok := config["max_tokens_per_shard"].(float64); ok {
		shardingConfig.MaxTokensPerShard = int(maxTokens)
	}
	if minTokens, ok := config["min_tokens_per_shard"].(float64); ok {
		shardingConfig.MinTokensPerShard = int(minTokens)
	}
	if maxReqs, ok := config["max_requirements"].(float64); ok {
		shardingConfig.MaxRequirements = int(maxReqs)
	}
	if minReqs, ok := config["min_requirements"].(float64); ok {
		shardingConfig.MinRequirements = int(minReqs)
	}
	if threshold, ok := config["similarity_threshold"].(float64); ok {
		shardingConfig.SimilarityThreshold = threshold
	}
	
	// 执行分片算法
	result, err := s.shardingAlgorithm.ExecuteSharding(requirements, shardingConfig)
	if err != nil {
		return nil, fmt.Errorf("执行分片算法失败: %w", err)
	}
	
	return result, nil
}

// buildSystemPrompt 构建系统提示词
func (s *BatchEvaluationService) buildSystemPrompt(config *BatchEvaluationConfig) string {
	var promptBuilder strings.Builder
	
	// 专业身份和标准说明
	promptBuilder.WriteString("# NESMA 2.2 功能点分析专家系统\n\n")
	promptBuilder.WriteString("你是一位拥有20年以上软件度量经验的NESMA 2.2国际标准专家，具备ISO/IEC 14143标准认证资质。你在功能点识别、复杂度评估、质量分析方面有深厚造诣，在金融、制造、医疗、政府、电商等15+个行业有丰富的项目实践。\n\n")
	
	// NESMA 2.2标准核心概念
	promptBuilder.WriteString("## NESMA 2.2 标准核心概念\n\n")
	promptBuilder.WriteString("### 功能点类型定义\n")
	promptBuilder.WriteString("- **ILF (内部逻辑文件)**: 由应用维护的用户可识别的逻辑相关数据组，如客户信息、订单数据等\n")
	promptBuilder.WriteString("- **EIF (外部接口文件)**: 由其他应用维护但本应用引用的用户可识别的逻辑相关数据组，如外部系统的主数据\n")
	promptBuilder.WriteString("- **EI (外部输入)**: 处理来自应用边界外的数据或控制信息的基本过程，如用户登录、数据录入等\n")
	promptBuilder.WriteString("- **EO (外部输出)**: 向应用边界外发送数据或控制信息的基本过程，如报表生成、数据导出等\n")
	promptBuilder.WriteString("- **EQ (外部查询)**: 从应用边界外发送输入并接收输出的基本过程，如数据查询、状态检查等\n\n")
	
	// 复杂度评估标准
	promptBuilder.WriteString("### 复杂度评估标准\n\n")
	promptBuilder.WriteString("#### 数据功能 (ILF/EIF) 复杂度\n")
	promptBuilder.WriteString("- **Low (低)**: DET ≤ 19 且 RET ≤ 2\n")
	promptBuilder.WriteString("- **Average (中)**: DET 20-50 且 RET 3-5，或 DET ≤ 19 且 RET ≥ 6\n")
	promptBuilder.WriteString("- **High (高)**: DET ≥ 51 且 RET ≥ 6，或 DET ≥ 20 且 RET ≥ 6\n\n")
	
	promptBuilder.WriteString("#### 事务功能 (EI/EO/EQ) 复杂度\n")
	promptBuilder.WriteString("- **Low (低)**: DET ≤ 15 且 FTR ≤ 1\n")
	promptBuilder.WriteString("- **Average (中)**: DET 16-45 且 FTR 2-3，或 DET ≤ 15 且 FTR ≥ 4\n")
	promptBuilder.WriteString("- **High (高)**: DET ≥ 46 且 FTR ≥ 4，或 DET ≥ 16 且 FTR ≥ 4\n\n")
	
	// 权重因子表
	promptBuilder.WriteString("### 权重因子表\n\n")
	promptBuilder.WriteString("| 功能类型 | 复杂度 | 权重因子 |\n")
	promptBuilder.WriteString("|---------|--------|----------|\n")
	promptBuilder.WriteString("| ILF | Low | 7 |\n")
	promptBuilder.WriteString("| ILF | Average | 10 |\n")
	promptBuilder.WriteString("| ILF | High | 15 |\n")
	promptBuilder.WriteString("| EIF | Low | 5 |\n")
	promptBuilder.WriteString("| EIF | Average | 7 |\n")
	promptBuilder.WriteString("| EIF | High | 10 |\n")
	promptBuilder.WriteString("| EI | Low | 3 |\n")
	promptBuilder.WriteString("| EI | Average | 4 |\n")
	promptBuilder.WriteString("| EI | High | 6 |\n")
	promptBuilder.WriteString("| EO | Low | 4 |\n")
	promptBuilder.WriteString("| EO | Average | 5 |\n")
	promptBuilder.WriteString("| EO | High | 7 |\n")
	promptBuilder.WriteString("| EQ | Low | 3 |\n")
	promptBuilder.WriteString("| EQ | Average | 4 |\n")
	promptBuilder.WriteString("| EQ | High | 6 |\n\n")
	
	// 分析指导原则
	promptBuilder.WriteString("## 分析指导原则\n\n")
	promptBuilder.WriteString("### 1. 功能识别原则\n")
	promptBuilder.WriteString("- **用户视角**: 从最终用户角度识别功能，而非技术实现角度\n")
	promptBuilder.WriteString("- **业务价值**: 每个功能必须为最终用户提供可识别的业务价值\n")
	promptBuilder.WriteString("- **独立性**: 功能之间应相对独立，避免重复计算\n")
	promptBuilder.WriteString("- **完整性**: 确保所有用户可识别的功能都被正确识别\n\n")
	
	promptBuilder.WriteString("### 2. 复杂度评估原则\n")
	promptBuilder.WriteString("- **数据元素(DET)**: 用户可识别的、非重复的数据字段\n")
	promptBuilder.WriteString("- **记录元素(RET)**: 用户可识别的、逻辑相关的数据子组\n")
	promptBuilder.WriteString("- **文件类型(FTR)**: 被事务功能访问的ILF或EIF\n")
	promptBuilder.WriteString("- **保守原则**: 当复杂度边界模糊时，选择较低的复杂度级别\n\n")
	
	promptBuilder.WriteString("### 3. 质量保证原则\n")
	promptBuilder.WriteString("- **一致性**: 相似功能应使用相同的分类和复杂度标准\n")
	promptBuilder.WriteString("- **可追溯性**: 每个功能点都应有明确的业务需求支撑\n")
	promptBuilder.WriteString("- **可验证性**: 分析结果应能被其他专家验证和重现\n")
	promptBuilder.WriteString("- **完整性**: 确保所有相关功能都被识别和计算\n\n")
	
	// 分析步骤
	promptBuilder.WriteString("## 标准分析步骤\n\n")
	promptBuilder.WriteString("### 步骤1: 功能类型识别\n")
	promptBuilder.WriteString("1. 分析需求描述中的数据处理模式\n")
	promptBuilder.WriteString("2. 识别是数据维护还是事务处理\n")
	promptBuilder.WriteString("3. 确定是内部数据还是外部接口\n")
	promptBuilder.WriteString("4. 区分输入、输出、查询功能\n\n")
	
	promptBuilder.WriteString("### 步骤2: 复杂度评估\n")
	promptBuilder.WriteString("1. 估算数据元素数量(DET)\n")
	promptBuilder.WriteString("2. 估算记录元素数量(RET)\n")
	promptBuilder.WriteString("3. 估算文件类型数量(FTR)\n")
	promptBuilder.WriteString("4. 根据标准表格确定复杂度级别\n\n")
	
	promptBuilder.WriteString("### 步骤3: 功能点计算\n")
	promptBuilder.WriteString("1. 根据功能类型和复杂度查找权重因子\n")
	promptBuilder.WriteString("2. 计算AFP (调整功能点) = UFP (未调整功能点) × 权重因子\n")
	promptBuilder.WriteString("3. 验证计算结果的合理性\n\n")
	
	promptBuilder.WriteString("### 步骤4: 质量评估\n")
	promptBuilder.WriteString("1. 评估分析的置信度(0.0-1.0)\n")
	promptBuilder.WriteString("2. 识别潜在的不确定性因素\n")
	promptBuilder.WriteString("3. 提供分析说明和推理过程\n\n")
	
	// 输出要求
	promptBuilder.WriteString("## 输出格式要求\n\n")
	promptBuilder.WriteString("请严格按照以下JSON格式输出分析结果：\n")
	promptBuilder.WriteString("```json\n")
	promptBuilder.WriteString("[\n")
	promptBuilder.WriteString("  {\n")
	promptBuilder.WriteString("    \"requirement_id\": \"需求唯一标识\",\n")
	promptBuilder.WriteString("    \"function_type\": \"EI|EO|EQ|ILF|EIF\",\n")
	promptBuilder.WriteString("    \"complexity_level\": \"Low|Average|High\",\n")
	promptBuilder.WriteString("    \"afp\": 调整功能点数(浮点数),\n")
	promptBuilder.WriteString("    \"ufp\": 未调整功能点数(浮点数),\n")
	promptBuilder.WriteString("    \"confidence_score\": 置信度(0.0-1.0),\n")
	promptBuilder.WriteString("    \"explanation\": \"详细的分析推理过程\",\n")
	promptBuilder.WriteString("    \"det_count\": 估算的数据元素数量,\n")
	promptBuilder.WriteString("    \"ret_count\": 估算的记录元素数量,\n")
	promptBuilder.WriteString("    \"ftr_count\": 估算的文件类型数量,\n")
	promptBuilder.WriteString("    \"weight_factor\": 使用的权重因子,\n")
	promptBuilder.WriteString("    \"quality_notes\": \"质量评估备注\"\n")
	promptBuilder.WriteString("  }\n")
	promptBuilder.WriteString("]\n")
	promptBuilder.WriteString("```\n\n")
	
	// 质量要求
	promptBuilder.WriteString("## 质量要求\n\n")
	promptBuilder.WriteString("- **准确性**: 功能类型识别必须基于NESMA 2.2标准\n")
	promptBuilder.WriteString("- **一致性**: 相似功能应使用相同的评估标准\n")
	promptBuilder.WriteString("- **完整性**: 所有用户可识别的功能都应被分析\n")
	promptBuilder.WriteString("- **可追溯性**: 每个分析结果都应有明确的推理过程\n")
	promptBuilder.WriteString("- **保守性**: 在边界情况下选择较低的复杂度级别\n\n")
	
	// 特殊注意事项
	promptBuilder.WriteString("## 特殊注意事项\n\n")
	promptBuilder.WriteString("- **边界情况**: 当功能类型或复杂度边界模糊时，选择更保守的评估\n")
	promptBuilder.WriteString("- **复合功能**: 如果一个需求包含多个功能，应分别识别和计算\n")
	promptBuilder.WriteString("- **技术细节**: 忽略技术实现细节，专注于业务功能识别\n")
	promptBuilder.WriteString("- **用户价值**: 确保每个识别的功能都为最终用户提供价值\n")
	promptBuilder.WriteString("- **标准遵循**: 严格遵循NESMA 2.2和ISO/IEC 14143标准\n\n")
	
	// 开始分析指令
	promptBuilder.WriteString("## 开始分析\n\n")
	promptBuilder.WriteString("请基于以上NESMA 2.2标准，对提供的需求进行专业的功能点分析。确保每个分析结果都符合国际标准，具有高质量和可验证性。\n\n")
	
	return promptBuilder.String()
}

// getRequirements 获取需求列表
func (s *BatchEvaluationService) getRequirements(cycleID, versionID uint) ([]nesma.NesmaRequirement, error) {
	var requirements []nesma.NesmaRequirement
	
	query := s.db.Where("cycle_id = ? AND version_id = ?", cycleID, versionID)
	
	// 只获取叶子节点需求（3-4级）
	query = query.Where("level IN (3, 4)")
	
	err := query.Find(&requirements).Error
	if err != nil {
		return nil, err
	}
	
	// 按照层级和编号排序
	sort.Slice(requirements, func(i, j int) bool {
		if requirements[i].Level != requirements[j].Level {
			return requirements[i].Level < requirements[j].Level
		}
		return requirements[i].Code < requirements[j].Code
	})
	
	return requirements, nil
}

// createTargetVersion 创建目标版本
func (s *BatchEvaluationService) createTargetVersion(cycleID, sourceVersionID uint) (uint, error) {
	// 获取源版本信息
	var sourceVersion nesma.NesmaRequirementVersion
	if err := s.db.First(&sourceVersion, sourceVersionID).Error; err != nil {
		return 0, err
	}
	
	// 创建新版本
	newVersion := nesma.NesmaRequirementVersion{
		CycleID:     cycleID,
		Version:     fmt.Sprintf("%s_batch_%d", sourceVersion.Version, time.Now().Unix()),
		VersionType: "ai_batch_analyzed",
		CreatedBy:   "batch_evaluation_service",
		Summary:     "批处理AI分析结果",
	}
	
	if err := s.db.Create(&newVersion).Error; err != nil {
		return 0, err
	}
	
	return newVersion.ID, nil
}

// 辅助方法
func (s *BatchEvaluationService) shouldContinueOnBatchFailure(config *BatchEvaluationConfig) bool {
	// 简化实现，可以根据配置决定
	return config.MaxRetries > 0
}

func (s *BatchEvaluationService) sendProgress(task *BatchEvaluationTask, operation string) {
	if task.progressCallback == nil {
		return
	}
	
	progress := BatchProgress{
		EvaluationID:     task.EvaluationID,
		TotalBatches:     task.TotalBatches,
		CompletedBatches: task.CompletedBatches,
		TotalRequirements: task.TotalRequirements,
		ProcessedCount:   task.ProcessedCount,
		SuccessCount:     task.SuccessCount,
		FailedCount:      task.FailedCount,
		Status:           task.Status,
		CurrentOperation: operation,
	}
	
	if task.TotalRequirements > 0 {
		progress.Progress = float64(task.ProcessedCount) / float64(task.TotalRequirements)
	}
	
	task.progressCallback(progress)
}

func (s *BatchEvaluationService) handleTaskError(task *BatchEvaluationTask, err error) {
	task.mu.Lock()
	defer task.mu.Unlock()
	
	task.Status = "failed"
	task.ErrorMessage = err.Error()
	now := time.Now()
	task.EndTime = &now
	
	// 更新指标
	s.metrics.mu.Lock()
	s.metrics.FailedEvaluations++
	s.metrics.mu.Unlock()
	
	// 发送最终进度
	s.sendProgress(task, "任务失败")
	
	global.GVA_LOG.Error("批处理评估任务失败",
		zap.String("evaluationID", task.EvaluationID),
		zap.Error(err))
}

func (s *BatchEvaluationService) handleTaskCancellation(task *BatchEvaluationTask) {
	task.mu.Lock()
	defer task.mu.Unlock()
	
	task.Status = "cancelled"
	now := time.Now()
	task.EndTime = &now
	
	s.sendProgress(task, "任务已取消")
	
	global.GVA_LOG.Info("批处理评估任务已取消",
		zap.String("evaluationID", task.EvaluationID))
}

func (s *BatchEvaluationService) completeTask(task *BatchEvaluationTask) {
	task.mu.Lock()
	defer task.mu.Unlock()
	
	task.Status = "completed"
	now := time.Now()
	task.EndTime = &now
	
	// 更新指标
	s.metrics.mu.Lock()
	s.metrics.CompletedEvaluations++
	s.metrics.TotalRequirements += int64(task.TotalRequirements)
	s.metrics.mu.Unlock()
	
	s.sendProgress(task, "任务完成")
	
	global.GVA_LOG.Info("批处理评估任务完成",
		zap.String("evaluationID", task.EvaluationID),
		zap.Int("处理需求数", task.ProcessedCount),
		zap.Int("成功数", task.SuccessCount),
		zap.Float64("成功率", float64(task.SuccessCount)/float64(task.ProcessedCount)*100))
}

// startWorkers 启动工作协程
func (s *BatchEvaluationService) startWorkers() {
	for i := 0; i < s.concurrentConfig.MaxEvaluationWorkers; i++ {
		go s.batchWorker()
	}
}

// batchWorker 批处理工作协程
func (s *BatchEvaluationService) batchWorker() {
	for {
		select {
		case task := <-s.workerPool.taskChan:
			result := s.processBatchWorkerTask(task)
			s.workerPool.resultChan <- result
		case <-s.shutdownChan:
			return
		}
	}
}

// processBatchWorkerTask 处理批处理工作任务
func (s *BatchEvaluationService) processBatchWorkerTask(task BatchWorkerTask) BatchWorkerResult {
	startTime := time.Now()
	
	// 构建批次提示词
	prompt, err := s.buildBatchPrompt(task)
	if err != nil {
		return BatchWorkerResult{
			EvaluationID: task.EvaluationID,
			BatchID:      task.BatchInfo.BatchID,
			Error:        err,
		}
	}
	
	// 调用AI
	response, err := s.aiService.GenerateText(context.Background(), prompt, &AIConfig{
		Model:     task.Config.AIModel,
		MaxTokens: 8192,
	})
	if err != nil {
		return BatchWorkerResult{
			EvaluationID: task.EvaluationID,
			BatchID:      task.BatchInfo.BatchID,
			Error:        err,
		}
	}
	
	// 解析结果
	results, err := s.parseBatchResponse(response.Text, task.BatchInfo.Requirements)
	if err != nil {
		return BatchWorkerResult{
			EvaluationID: task.EvaluationID,
			BatchID:      task.BatchInfo.BatchID,
			Error:        err,
		}
	}
	
	// 计算Token使用量
	tokenUsage := s.tokenCalculator.EstimateTokens(prompt+response.Text, task.Config.AIModel)
	
	return BatchWorkerResult{
		EvaluationID: task.EvaluationID,
		BatchID:      task.BatchInfo.BatchID,
		Results:      results,
		TokenUsage:   TokenUsage{
			TotalTokens: tokenUsage,
		},
		ModelUsed:    task.Config.AIModel,
		Duration: time.Since(startTime),
	}
}

// buildBatchPrompt 构建批次提示词
func (s *BatchEvaluationService) buildBatchPrompt(task BatchWorkerTask) (string, error) {
	var promptBuilder strings.Builder
	
	// 添加上下文信息
	contextPrompt, err := s.contextManager.GenerateContextPrompt(task.EvaluationID, task.BatchInfo.BatchNumber)
	if err != nil {
		return "", err
	}
	promptBuilder.WriteString(contextPrompt)
	
	// 添加系统指令
	systemPrompt := s.buildSystemPrompt(task.Config)
	promptBuilder.WriteString(systemPrompt)
	
	// 添加当前批次需求
	promptBuilder.WriteString(fmt.Sprintf("\n## 当前批次需求 (批次 %d)\n", task.BatchInfo.BatchNumber))
	for i, req := range task.BatchInfo.Requirements {
		promptBuilder.WriteString(fmt.Sprintf("### 需求 %d\n", i+1))
		promptBuilder.WriteString(fmt.Sprintf("- 编号: %s\n", req.Code))
		promptBuilder.WriteString(fmt.Sprintf("- 标题: %s\n", req.Title))
		promptBuilder.WriteString(fmt.Sprintf("- 描述: %s\n", req.Description))
		promptBuilder.WriteString(fmt.Sprintf("- 层级: %d\n", req.Level))
		promptBuilder.WriteString("\n")
	}
	
	// 添加输出格式要求
	promptBuilder.WriteString("## 输出格式要求\n")
	promptBuilder.WriteString("请严格按照以下JSON格式输出分析结果：\n")
	promptBuilder.WriteString("```json\n")
	promptBuilder.WriteString("[\n")
	promptBuilder.WriteString("  {\n")
	promptBuilder.WriteString("    \"requirement_id\": \"需求唯一标识\",\n")
	promptBuilder.WriteString("    \"function_type\": \"EI|EO|EQ|ILF|EIF\",\n")
	promptBuilder.WriteString("    \"complexity_level\": \"Low|Average|High\",\n")
	promptBuilder.WriteString("    \"afp\": 调整功能点数(浮点数),\n")
	promptBuilder.WriteString("    \"ufp\": 未调整功能点数(浮点数),\n")
	promptBuilder.WriteString("    \"confidence_score\": 置信度(0.0-1.0),\n")
	promptBuilder.WriteString("    \"explanation\": \"详细的分析推理过程\",\n")
	promptBuilder.WriteString("    \"det_count\": 估算的数据元素数量,\n")
	promptBuilder.WriteString("    \"ret_count\": 估算的记录元素数量,\n")
	promptBuilder.WriteString("    \"ftr_count\": 估算的文件类型数量,\n")
	promptBuilder.WriteString("    \"weight_factor\": 使用的权重因子,\n")
	promptBuilder.WriteString("    \"quality_notes\": \"质量评估备注\"\n")
	promptBuilder.WriteString("  }\n")
	promptBuilder.WriteString("]\n")
	promptBuilder.WriteString("```\n")
	
	return promptBuilder.String(), nil
}

// parseBatchResponse 解析批次响应
func (s *BatchEvaluationService) parseBatchResponse(response string, requirements []nesma.NesmaRequirement) ([]EvaluationResult, error) {
	// 提取JSON
	jsonStart := strings.Index(response, "[")
	jsonEnd := strings.LastIndex(response, "]")
	if jsonStart == -1 || jsonEnd == -1 {
		return nil, fmt.Errorf("响应中未找到JSON数组")
	}
	
	jsonStr := response[jsonStart : jsonEnd+1]
	
	var batchResults []struct {
		RequirementID    string  `json:"requirement_id"`
		FunctionType     string  `json:"function_type"`
		ComplexityLevel  string  `json:"complexity_level"`
		AFP              float64 `json:"afp"`
		UFP              float64 `json:"ufp"`
		ConfidenceScore  float64 `json:"confidence_score"`
		Explanation      string  `json:"explanation"`
		DETCount         int     `json:"det_count"`
		RETCount         int     `json:"ret_count"`
		FTRCount         int     `json:"ftr_count"`
		WeightFactor     float64 `json:"weight_factor"`
		QualityNotes     string  `json:"quality_notes"`
	}
	
	if err := json.Unmarshal([]byte(jsonStr), &batchResults); err != nil {
		return nil, fmt.Errorf("解析JSON失败: %w", err)
	}
	
	// 转换为EvaluationResult
	results := make([]EvaluationResult, 0, len(batchResults))
	
	for i, result := range batchResults {
		if i < len(requirements) {
			evalResult := EvaluationResult{
				RequirementID:   requirements[i].ID,
				FunctionType:    result.FunctionType,
				ComplexityLevel: result.ComplexityLevel,
				AFP:             result.AFP,
				UFP:             result.UFP,
				ConfidenceScore: result.ConfidenceScore,
				Timestamp:       time.Now(),
				// 新增字段映射
				Explanation:     result.Explanation,
				DETCount:        result.DETCount,
				RETCount:        result.RETCount,
				FTRCount:        result.FTRCount,
				WeightFactor:    result.WeightFactor,
				QualityNotes:    result.QualityNotes,
			}
			results = append(results, evalResult)
		}
	}
	
	return results, nil
}

// GetBatchEvaluationStatus 获取批处理评估状态
func (s *BatchEvaluationService) GetBatchEvaluationStatus(evaluationID string) (*BatchEvaluationTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	task, exists := s.activeEvaluations[evaluationID]
	if !exists {
		return nil, fmt.Errorf("评估任务不存在: %s", evaluationID)
	}
	
	return task, nil
}

// CancelBatchEvaluation 取消批处理评估
func (s *BatchEvaluationService) CancelBatchEvaluation(evaluationID string) error {
	s.mu.RLock()
	task, exists := s.activeEvaluations[evaluationID]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("评估任务不存在: %s", evaluationID)
	}
	
	task.mu.Lock()
	if task.cancelFunc != nil {
		task.cancelFunc()
	}
	task.mu.Unlock()
	
	return nil
}

// GetBatchEvaluationMetrics 获取批处理评估指标
func (s *BatchEvaluationService) GetBatchEvaluationMetrics() *BatchEvaluationMetrics {
	s.metrics.mu.RLock()
	defer s.metrics.mu.RUnlock()
	
	// 返回副本
	metrics := *s.metrics
	return &metrics
}

// NewWorkerPool 创建工作协程池
func NewWorkerPool(workerCount int) *WorkerPool {
	return &WorkerPool{
		workerCount: workerCount,
		taskChan:    make(chan BatchWorkerTask, 100),
		resultChan:  make(chan BatchWorkerResult, 100),
		stopChan:    make(chan struct{}),
	}
}

// 新增API接口扩展方法

// GetBatchEvaluationResults 获取批量评估结果（分页）
func (s *BatchEvaluationService) GetBatchEvaluationResults(taskID string, page, pageSize int) ([]EvaluationResult, int64, error) {
	s.mu.RLock()
	task, exists := s.activeEvaluations[taskID]
	s.mu.RUnlock()
	
	if !exists {
		return nil, 0, fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.mu.RLock()
	results := task.Results
	task.mu.RUnlock()
	
	total := int64(len(results))
	
	// 计算分页范围
	start := (page - 1) * pageSize
	end := start + pageSize
	
	if start > len(results) {
		return []EvaluationResult{}, total, nil
	}
	
	if end > len(results) {
		end = len(results)
	}
	
	return results[start:end], total, nil
}

// GetBatchEvaluationDetail 获取批量评估详情
func (s *BatchEvaluationService) GetBatchEvaluationDetail(taskID string) (*BatchEvaluationTask, error) {
	s.mu.RLock()
	task, exists := s.activeEvaluations[taskID]
	s.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("任务不存在: %s", taskID)
	}
	
	// 创建副本以避免并发问题
	task.mu.RLock()
	defer task.mu.RUnlock()
	
	detailTask := &BatchEvaluationTask{
		ID:                task.ID,
		EvaluationID:      task.EvaluationID,
		ProjectID:         task.ProjectID,
		CycleID:           task.CycleID,
		SourceVersionID:   task.SourceVersionID,
		TargetVersionID:   task.TargetVersionID,
		Status:            task.Status,
		Progress:          task.Progress,
		CurrentBatch:      task.CurrentBatch,
		TotalBatches:      task.TotalBatches,
		CompletedBatches:  task.CompletedBatches,
		ProcessedCount:    task.ProcessedCount,
		TotalRequirements: task.TotalRequirements,
		CreatedAt:         task.CreatedAt,
		StartTime:         task.StartTime,
		EndTime:           task.EndTime,
		EstimatedEndTime:  task.EstimatedEndTime,
		EstimatedDuration: task.EstimatedDuration,
		ProcessingTime:    task.ProcessingTime,
		StatusMessage:     task.StatusMessage,
		Config:            task.Config,
		Summary:           task.Summary,
		QualityMetrics:    task.QualityMetrics,
		TokenUsage:        task.TokenUsage,
		Batches:           make([]BatchInfo, len(task.Batches)),
		Results:           make([]EvaluationResult, len(task.Results)),
	}
	
	// 复制批次信息
	copy(detailTask.Batches, task.Batches)
	copy(detailTask.Results, task.Results)
	
	return detailTask, nil
}

// BatchEvaluationFilter 批量评估过滤器
type BatchEvaluationFilter struct {
	ProjectID uint   `json:"project_id"`
	CycleID   uint   `json:"cycle_id"`
	Status    string `json:"status"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

// GetBatchEvaluationList 获取批量评估任务列表
func (s *BatchEvaluationService) GetBatchEvaluationList(filter *BatchEvaluationFilter) ([]BatchEvaluationTask, int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var filteredTasks []BatchEvaluationTask
	
	for _, task := range s.activeEvaluations {
		task.mu.RLock()
		
		// 应用过滤条件
		if filter.ProjectID != 0 && task.ProjectID != filter.ProjectID {
			task.mu.RUnlock()
			continue
		}
		
		if filter.CycleID != 0 && task.CycleID != filter.CycleID {
			task.mu.RUnlock()
			continue
		}
		
		if filter.Status != "" && task.Status != filter.Status {
			task.mu.RUnlock()
			continue
		}
		
		// 创建任务副本
		filteredTask := BatchEvaluationTask{
			ID:                task.ID,
			EvaluationID:      task.EvaluationID,
			ProjectID:         task.ProjectID,
			CycleID:           task.CycleID,
			Status:            task.Status,
			Progress:          task.Progress,
			TotalBatches:      task.TotalBatches,
			TotalRequirements: task.TotalRequirements,
			CreatedAt:         task.CreatedAt,
			StartTime:         task.StartTime,
			EndTime:           task.EndTime,
			EstimatedDuration: task.EstimatedDuration,
			ProcessingTime:    task.ProcessingTime,
			Config:            task.Config,
		}
		
		filteredTasks = append(filteredTasks, filteredTask)
		task.mu.RUnlock()
	}
	
	total := int64(len(filteredTasks))
	
	// 分页处理
	if filter.Page > 0 && filter.PageSize > 0 {
		start := (filter.Page - 1) * filter.PageSize
		end := start + filter.PageSize
		
		if start > len(filteredTasks) {
			return []BatchEvaluationTask{}, total, nil
		}
		
		if end > len(filteredTasks) {
			end = len(filteredTasks)
		}
		
		filteredTasks = filteredTasks[start:end]
	}
	
	return filteredTasks, total, nil
}

// CancelBatchEvaluationTask 取消批量评估任务
func (s *BatchEvaluationService) CancelBatchEvaluationTask(taskID string) error {
	s.mu.RLock()
	task, exists := s.activeEvaluations[taskID]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.mu.Lock()
	defer task.mu.Unlock()
	
	// 检查任务状态
	if task.Status == "completed" || task.Status == "failed" || task.Status == "cancelled" {
		return fmt.Errorf("任务已完成或已取消，无法取消")
	}
	
	// 设置取消标志
	task.Status = "cancelled"
	now := time.Now()
	task.EndTime = &now
	
	// 发送取消信号
	select {
	case s.shutdownChan <- struct{}{}:
	default:
	}
	
	s.sendProgress(task, "任务已取消")
	
	global.GVA_LOG.Info("批量评估任务已取消", 
		zap.String("taskID", taskID),
		zap.String("evaluationID", task.EvaluationID))
	
	return nil
}

// RetryBatchEvaluation 重试批量评估任务
func (s *BatchEvaluationService) RetryBatchEvaluation(taskID string, batchIDs []string, config *BatchEvaluationConfig) error {
	s.mu.RLock()
	task, exists := s.activeEvaluations[taskID]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.mu.Lock()
	defer task.mu.Unlock()
	
	// 检查任务状态
	if task.Status == "running" {
		return fmt.Errorf("任务正在运行中，无法重试")
	}
	
	// 更新配置（如果提供）
	if config != nil {
		task.Config = config
	}
	
	// 重置指定批次的状态
	if len(batchIDs) > 0 {
		for i, batch := range task.Batches {
			for _, batchID := range batchIDs {
				if batch.BatchID == batchID {
					task.Batches[i].Status = "pending"
					task.Batches[i].RetryCount++
					task.Batches[i].ErrorMessage = ""
					task.Batches[i].StartTime = nil
					task.Batches[i].EndTime = nil
				}
			}
		}
	} else {
		// 重试所有失败的批次
		for i, batch := range task.Batches {
			if batch.Status == "failed" {
				task.Batches[i].Status = "pending"
				task.Batches[i].RetryCount++
				task.Batches[i].ErrorMessage = ""
				task.Batches[i].StartTime = nil
				task.Batches[i].EndTime = nil
			}
		}
	}
	
	// 重置任务状态
	task.Status = "pending"
	task.EndTime = nil
	task.StatusMessage = "准备重试"
	
	// 重新启动任务
	go func() {
		if err := s.StartBatchEvaluationTask(taskID); err != nil {
			global.GVA_LOG.Error("重试任务启动失败", 
				zap.String("taskID", taskID),
				zap.Error(err))
		}
	}()
	
	global.GVA_LOG.Info("批量评估任务重试", 
		zap.String("taskID", taskID),
		zap.Strings("batchIDs", batchIDs))
	
	return nil
}

// StartBatchEvaluationTask 启动批量评估任务
func (s *BatchEvaluationService) StartBatchEvaluationTask(taskID string) error {
	s.mu.RLock()
	task, exists := s.activeEvaluations[taskID]
	s.mu.RUnlock()
	
	if !exists {
		return fmt.Errorf("任务不存在: %s", taskID)
	}
	
	task.mu.Lock()
	defer task.mu.Unlock()
	
	// 检查任务状态
	if task.Status == "running" {
		return fmt.Errorf("任务已在运行中")
	}
	
	if task.Status == "completed" {
		return fmt.Errorf("任务已完成")
	}
	
	// 设置任务状态
	task.Status = "running"
	now := time.Now()
	task.StartTime = now
	
	// 异步执行任务
	go func() {
		ctx, cancel := context.WithCancel(context.Background())
		task.cancelFunc = cancel
		s.executeBatchEvaluation(ctx, task)
	}()
	
	s.sendProgress(task, "任务已启动")
	
	global.GVA_LOG.Info("批量评估任务已启动", 
		zap.String("taskID", taskID),
		zap.String("evaluationID", task.EvaluationID))
	
	return nil
}

// Shutdown 关闭服务
func (s *BatchEvaluationService) Shutdown() {
	close(s.shutdownChan)
	s.contextManager.Stop()
	
	global.GVA_LOG.Info("批处理评估服务已关闭")
}

// setupCircuitBreakers 设置断路器
func (s *BatchEvaluationService) setupCircuitBreakers() {
	// AI服务断路器
	aiConfig := &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:           60 * time.Second,
		ResetTimeout:      30 * time.Second,
		MaxConcurrentCalls: 50,
	}
	s.circuitBreakerManager.GetOrCreate("ai_service", aiConfig)
	
	// 数据库断路器
	dbConfig := &CircuitBreakerConfig{
		FailureThreshold:   3,
		SuccessThreshold:   2,
		Timeout:           30 * time.Second,
		ResetTimeout:      15 * time.Second,
		MaxConcurrentCalls: 100,
	}
	s.circuitBreakerManager.GetOrCreate("database", dbConfig)
	
	// 知识库服务断路器
	knowledgeConfig := &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:           45 * time.Second,
		ResetTimeout:      20 * time.Second,
		MaxConcurrentCalls: 30,
	}
	s.circuitBreakerManager.GetOrCreate("knowledge_service", knowledgeConfig)
	
	global.GVA_LOG.Info("断路器设置完成",
		zap.Int("断路器数量", 3))
}

// executeRequirementAnalysisWithRetry 执行带重试的需求分析
func (s *BatchEvaluationService) executeRequirementAnalysisWithRetry(ctx context.Context, requirement nesma.NesmaRequirement, batchContext map[string]interface{}) (*EvaluationResult, error) {
	// 获取AI服务断路器
	aiBreaker, _ := s.circuitBreakerManager.Get("ai_service")
	
	var result *EvaluationResult
	
	// 使用断路器执行操作
	err := aiBreaker.ExecuteWithFallback(
		func() error {
			// 使用错误处理器进行重试
			return s.errorHandler.ExecuteWithRetryTyped(ctx, func() error {
				var err error
				result, err = s.executeRequirementAnalysis(ctx, requirement, batchContext)
				return err
			}, ErrorTypeAPI)
		},
		func() error {
			// 降级操作：使用默认值
			global.GVA_LOG.Warn("AI服务不可用，使用降级策略",
				zap.Uint("requirement_id", requirement.ID),
				zap.String("title", requirement.Title))
			
			result = s.createFallbackResult(requirement)
			return nil
		},
	)
	
	return result, err
}

// createFallbackResult 创建降级结果
func (s *BatchEvaluationService) createFallbackResult(requirement nesma.NesmaRequirement) *EvaluationResult {
	return &EvaluationResult{
		RequirementID:    requirement.ID,
		FunctionType:     "EI", // 默认为外部输入
		Complexity:       "Average", // 默认中等复杂度
		AFP:              3.0, // 默认AFP值
		UFP:              3.0, // 默认UFP值
		Confidence:       0.3, // 低置信度表示这是降级结果
		ModelUsed:        "fallback",
		ProcessingTime:   time.Millisecond * 100,
		IsFallback:       true,
		FallbackReason:   "AI服务不可用，使用默认值",
		Explanation:      "由于AI服务暂时不可用，系统自动提供了保守的评估结果",
		Timestamp:        time.Now(),
	}
}

// executeWithRetryAndCircuitBreaker 执行带重试和断路器保护的操作
func (s *BatchEvaluationService) executeWithRetryAndCircuitBreaker(ctx context.Context, operation func() error, circuitBreakerName string, errorType ErrorType) error {
	breaker, exists := s.circuitBreakerManager.Get(circuitBreakerName)
	if !exists {
		// 如果断路器不存在，直接使用错误处理器重试
		return s.errorHandler.ExecuteWithRetryTyped(ctx, operation, errorType)
	}
	
	return breaker.Execute(func() error {
		return s.errorHandler.ExecuteWithRetryTyped(ctx, operation, errorType)
	})
}

// recoverFromPanic 从panic中恢复
func (s *BatchEvaluationService) recoverFromPanic(taskID string) {
	if r := recover(); r != nil {
		global.GVA_LOG.Error("批量评估任务发生panic",
			zap.String("task_id", taskID),
			zap.Any("panic", r))
		
		// 更新任务状态
		s.mu.RLock()
		task, exists := s.activeEvaluations[taskID]
		s.mu.RUnlock()
		
		if exists {
			task.mu.Lock()
			task.Status = "failed"
			task.ErrorMessage = fmt.Sprintf("系统内部错误: %v", r)
			now := time.Now()
			task.EndTime = &now
			task.mu.Unlock()
			
			s.sendProgress(task, "任务因系统错误终止")
		}
	}
}

// validateBatchData 验证批次数据
func (s *BatchEvaluationService) validateBatchData(batch BatchInfo) error {
	if len(batch.Requirements) == 0 {
		return &RetryableError{
			Type:        ErrorTypeInvalidData,
			Severity:    SeverityHigh,
			Message:     "批次中没有需求数据",
			Retryable:   false,
			Timestamp:   time.Now(),
		}
	}
	
	if batch.EstimatedTokens <= 0 {
		return &RetryableError{
			Type:        ErrorTypeInvalidData,
			Severity:    SeverityLow,
			Message:     "估算Token数量无效",
			Retryable:   true,
			Timestamp:   time.Now(),
		}
	}
	
	return nil
}

// handleBatchFailure 处理批次失败
func (s *BatchEvaluationService) handleBatchFailure(task *BatchEvaluationTask, batchIndex int, err error) {
	task.mu.Lock()
	defer task.mu.Unlock()
	
	// 分类错误
	retryableErr := s.errorHandler.ClassifyError(err)
	
	if batchIndex < len(task.Batches) {
		batch := &task.Batches[batchIndex]
		batch.Status = "failed"
		batch.ErrorMessage = retryableErr.Message
		batch.RetryCount++
		
		now := time.Now()
		batch.EndTime = &now
	}
	
	task.FailedCount++
	
	// 根据错误严重程度决定是否继续
	switch retryableErr.Severity {
	case SeverityCritical:
		task.Status = "failed"
		task.ErrorMessage = retryableErr.Message
		now := time.Now()
		task.EndTime = &now
		
		global.GVA_LOG.Error("严重错误，停止批量评估",
			zap.String("task_id", task.ID),
			zap.String("error", retryableErr.Message))
		
	case SeverityHigh:
		// 高级错误：跳过当前批次，继续处理其他批次
		global.GVA_LOG.Warn("高级错误，跳过当前批次",
			zap.String("task_id", task.ID),
			zap.Int("batch_index", batchIndex),
			zap.String("error", retryableErr.Message))
		
	default:
		// 中低级错误：记录日志，继续处理
		global.GVA_LOG.Warn("批次处理失败",
			zap.String("task_id", task.ID),
			zap.Int("batch_index", batchIndex),
			zap.String("error", retryableErr.Message))
	}
}

// GetErrorHandlerMetrics 获取错误处理指标
func (s *BatchEvaluationService) GetErrorHandlerMetrics() *ErrorMetrics {
	return s.errorHandler.GetMetrics()
}

// GetCircuitBreakerMetrics 获取断路器指标
func (s *BatchEvaluationService) GetCircuitBreakerMetrics() map[string]*CircuitBreakerMetrics {
	return s.circuitBreakerManager.GetMetricsReport()
}

// ResetErrorMetrics 重置错误统计
func (s *BatchEvaluationService) ResetErrorMetrics() {
	s.errorHandler.ResetMetrics()
}

// ResetCircuitBreakers 重置所有断路器
func (s *BatchEvaluationService) ResetCircuitBreakers() {
	s.circuitBreakerManager.ResetAll()
}

// SetBatchEvaluationService 设置全局服务实例
func SetBatchEvaluationService(service *BatchEvaluationService) {
	batchEvaluationServiceInstance = service
}

// GetBatchEvaluationService 获取全局服务实例
func GetBatchEvaluationService() *BatchEvaluationService {
	return batchEvaluationServiceInstance
}

// executeRequirementAnalysis 执行需求分析
func (s *BatchEvaluationService) executeRequirementAnalysis(ctx context.Context, requirement nesma.NesmaRequirement, batchContext map[string]interface{}) (*EvaluationResult, error) {
	// 这里应该实现具体的需求分析逻辑
	// 为了编译通过，先提供一个简单的实现
	return &EvaluationResult{
		RequirementID:    requirement.ID,
		FunctionType:     "EI",
		ComplexityLevel:  "Average",
		ConfidenceScore:  0.8,
		AFP:              3.0,
		UFP:              3.0,
		Timestamp:        time.Now(),
		ModelUsed:        "default",
		ProcessingTime:   time.Millisecond * 100,
	}, nil
}