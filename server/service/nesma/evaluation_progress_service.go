package nesma

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaRes "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/response"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProgressEvaluationContext 进度评估过程中的上下文数据
type ProgressEvaluationContext struct {
	Evaluation      *nesma.NesmaEvaluation
	Requirements    []nesma.NesmaRequirement
	FunctionPoints  []nesma.NesmaFunctionPoint
	Config          map[string]interface{}
	ProcessedCount  int
	AnalysisResults map[string]interface{}
	CancelCtx       context.Context    // 取消上下文
	CancelFunc      context.CancelFunc // 取消函数
}

type EvaluationProgressService struct{
	// 存储正在运行的评估的取消函数
	runningEvaluations map[uint]context.CancelFunc
	mu                 sync.RWMutex
}

// 全局单例实例
var (
	progressServiceInstance *EvaluationProgressService
	progressServiceOnce     sync.Once
)

// NewEvaluationProgressService 创建评估进度服务实例
func NewEvaluationProgressService() *EvaluationProgressService {
	return &EvaluationProgressService{
		runningEvaluations: make(map[uint]context.CancelFunc),
	}
}

// GetEvaluationProgressService 获取评估进度服务单例
func GetEvaluationProgressService() *EvaluationProgressService {
	progressServiceOnce.Do(func() {
		progressServiceInstance = NewEvaluationProgressService()
		// 安全日志记录：检查日志是否已初始化
		if global.GVA_LOG != nil {
			global.GVA_LOG.Info("评估进度服务单例初始化完成")
		}
	})
	return progressServiceInstance
}

// registerRunningEvaluation 注册正在运行的评估
func (s *EvaluationProgressService) registerRunningEvaluation(evaluationID uint, cancelFunc context.CancelFunc) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.runningEvaluations == nil {
		s.runningEvaluations = make(map[uint]context.CancelFunc)
	}
	
	s.runningEvaluations[evaluationID] = cancelFunc
	global.GVA_LOG.Info("注册正在运行的评估", zap.Uint("evaluationID", evaluationID))
}

// unregisterRunningEvaluation 注销正在运行的评估
func (s *EvaluationProgressService) unregisterRunningEvaluation(evaluationID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.runningEvaluations != nil {
		delete(s.runningEvaluations, evaluationID)
		global.GVA_LOG.Info("注销正在运行的评估", zap.Uint("evaluationID", evaluationID))
	}
}

// checkCancellation 检查评估是否已被取消
func (s *EvaluationProgressService) checkCancellation(evaluationID uint) error {
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.Select("status").First(&evaluation, evaluationID).Error
	if err != nil {
		return fmt.Errorf("检查取消状态失败: %v", err)
	}
	
	if evaluation.Status == "cancelled" {
		return fmt.Errorf("评估已被取消")
	}
	
	return nil
}

// checkCancellationWithContext 检查上下文是否已被取消
func (s *EvaluationProgressService) checkCancellationWithContext(evalCtx *ProgressEvaluationContext) error {
	// 首先检查上下文取消信号
	select {
	case <-evalCtx.CancelCtx.Done():
		return fmt.Errorf("评估上下文已被取消: %v", evalCtx.CancelCtx.Err())
	default:
		// 继续执行
	}
	
	// 然后检查数据库状态
	return s.checkCancellation(evalCtx.Evaluation.ID)
}

// GetEvaluationProgress 获取评估进度
func (s *EvaluationProgressService) GetEvaluationProgress(evaluationID uint) (*nesmaRes.EvaluationProgressResponse, error) {
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("评估记录不存在")
		}
		return nil, err
	}

	// 计算预计完成时间
	var estimatedTime *time.Time
	if evaluation.StartTime != nil && evaluation.Progress > 0 && evaluation.Progress < 100 {
		elapsed := time.Since(*evaluation.StartTime)
		if evaluation.Progress > 5 { // 避免除零和初期估算不准确
			totalTime := time.Duration(float64(elapsed) / evaluation.Progress * 100)
			estimated := evaluation.StartTime.Add(totalTime)
			estimatedTime = &estimated
		}
	}

	// 构建阶段信息
	phases := s.buildPhaseInfo(&evaluation)
	
	// 计算当前阶段索引（与buildPhaseInfo保持一致）
	currentPhaseIndex := evaluation.GetCurrentPhaseIndex()
	if evaluation.Status == "completed" && currentPhaseIndex != len(phases)-1 {
		currentPhaseIndex = len(phases) - 1
	}

	response := &nesmaRes.EvaluationProgressResponse{
		EvaluationID:          evaluation.ID,
		Status:                evaluation.Status,
		Progress:              evaluation.Progress,
		CurrentPhase:          evaluation.CurrentPhase,
		Message:               evaluation.ProgressMessage,
		CompletedSteps:        evaluation.CompletedSteps,
		TotalSteps:            evaluation.TotalSteps,
		ProcessedRequirements: evaluation.ProcessedRequirements,
		TotalRequirements:     evaluation.TotalRequirements,
		StartTime:             evaluation.StartTime,
		EstimatedTime:         estimatedTime,
		Phases:                phases,
		CurrentPhaseIndex:     currentPhaseIndex,
	}

	// 计算已用时间
	if evaluation.StartTime != nil {
		if evaluation.CompletionTime != nil {
			response.ElapsedTime = int64(evaluation.CompletionTime.Sub(*evaluation.StartTime).Seconds())
		} else {
			response.ElapsedTime = int64(time.Since(*evaluation.StartTime).Seconds())
		}
	}

	return response, nil
}

// StartEvaluation 启动评估
func (s *EvaluationProgressService) StartEvaluation(evaluationID uint, config map[string]interface{}) (*nesmaRes.EvaluationProgressResponse, error) {
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationID).Error
	if err != nil {
		return nil, fmt.Errorf("评估记录不存在: %v", err)
	}

	// 检查评估状态
	if evaluation.Status == "processing" {
		return nil, fmt.Errorf("评估已在运行中")
	}

	if evaluation.Status == "completed" {
		return nil, fmt.Errorf("评估已完成，无法重新启动")
	}

	// 更新评估状态
	now := time.Now()
	evaluation.Status = "processing"
	evaluation.StartTime = &now
	evaluation.SetPhase("初始化评估")
	evaluation.ProgressMessage = "正在初始化评估流程..."

	err = global.GVA_DB.Save(&evaluation).Error
	if err != nil {
		return nil, fmt.Errorf("更新评估状态失败: %v", err)
	}

	// 异步启动评估流程
	go s.performEvaluation(evaluationID, config)

	return s.GetEvaluationProgress(evaluationID)
}

// performEvaluation 执行评估流程
func (s *EvaluationProgressService) performEvaluation(evaluationID uint, config map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			global.GVA_LOG.Error("评估流程发生panic", 
				zap.Uint("evaluationId", evaluationID),
				zap.Any("panic", r))
			s.handleEvaluationError(evaluationID, fmt.Errorf("评估流程异常: %v", r))
		}
	}()

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel() // 确保在函数结束时取消上下文
		s.unregisterRunningEvaluation(evaluationID) // 注销运行中的评估
	}()
	
	// 注册正在运行的评估
	s.registerRunningEvaluation(evaluationID, cancel)
	
	// 获取评估记录
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.Preload("Project").Preload("Cycle").First(&evaluation, evaluationID).Error
	if err != nil {
		s.handleEvaluationError(evaluationID, fmt.Errorf("获取评估记录失败: %v", err))
		return
	}

	// 初始化评估上下文
	evalCtx := &ProgressEvaluationContext{
		Evaluation:      &evaluation,
		Requirements:    []nesma.NesmaRequirement{},
		FunctionPoints:  []nesma.NesmaFunctionPoint{},
		Config:          config,
		ProcessedCount:  0,
		AnalysisResults: make(map[string]interface{}),
		CancelCtx:       ctx,
		CancelFunc:      cancel,
	}

	global.GVA_LOG.Info("开始执行NESMA评估", 
		zap.Uint("evaluationId", evaluationID),
		zap.String("evaluationName", evaluation.EvaluationName))

	// 阶段0: 初始化评估
	err = s.executePhase(ctx, evaluationID, "初始化评估", func() error {
		return s.initializeEvaluationWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段1: 获取需求数据
	err = s.executePhase(ctx, evaluationID, "获取需求数据", func() error {
		return s.loadRequirementsWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段2: 功能点识别
	err = s.executePhase(ctx, evaluationID, "功能点识别", func() error {
		return s.identifyFunctionPointsWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段3: 复杂度分析
	err = s.executePhase(ctx, evaluationID, "复杂度分析", func() error {
		return s.analyzeComplexityWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段4: 权重计算
	err = s.executePhase(ctx, evaluationID, "权重计算", func() error {
		return s.calculateWeightsWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段5: 调整因子计算
	err = s.executePhase(ctx, evaluationID, "调整因子计算", func() error {
		return s.calculateAdjustmentFactorsWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段6: 验证检查
	err = s.executePhase(ctx, evaluationID, "验证检查", func() error {
		return s.performValidationWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段7: 结果汇总
	err = s.executePhase(ctx, evaluationID, "结果汇总", func() error {
		return s.summarizeResultsWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	// 阶段8: 完成评估
	err = s.executePhase(ctx, evaluationID, "完成评估", func() error {
		return s.completeEvaluationWithContext(evalCtx)
	})
	if err != nil {
		return
	}

	global.GVA_LOG.Info("NESMA评估完成", 
		zap.Uint("evaluationId", evaluationID),
		zap.Float64("totalFP", evaluation.TotalFunctionPoints))
}

// executePhase 执行单个阶段
func (s *EvaluationProgressService) executePhase(ctx context.Context, evaluationID uint, phase string, fn func() error) error {
	// 检查取消状态
	if err := s.checkCancellation(evaluationID); err != nil {
		global.GVA_LOG.Info("阶段执行被取消", 
			zap.String("phase", phase), 
			zap.Uint("evaluationID", evaluationID))
		return err
	}
	
	// 检查上下文取消
	select {
	case <-ctx.Done():
		return fmt.Errorf("评估上下文已取消: %v", ctx.Err())
	default:
		// 继续执行
	}

	// 开始阶段 - 更新为正在执行状态
	err := s.UpdateProgress(evaluationID, -1, phase, fmt.Sprintf("正在执行: %s", phase))
	if err != nil {
		global.GVA_LOG.Error("更新阶段开始状态失败", zap.Error(err))
	}

	// 执行阶段逻辑
	start := time.Now()
	err = fn()
	duration := time.Since(start)

	if err != nil {
		global.GVA_LOG.Error("阶段执行失败", 
			zap.String("phase", phase),
			zap.Duration("duration", duration),
			zap.Error(err))
		s.handleEvaluationError(evaluationID, fmt.Errorf("阶段 %s 执行失败: %v", phase, err))
		return err
	}

	// 阶段完成 - 计算当前进度并更新步骤计数
	var evaluation nesma.NesmaEvaluation
	if err := global.GVA_DB.First(&evaluation, evaluationID).Error; err == nil {
		phases := evaluation.GetProgressPhases()
		currentIndex := -1
		for i, p := range phases {
			if p == phase {
				currentIndex = i
				break
			}
		}
		
		if currentIndex >= 0 {
			// 更新步骤统计
			evaluation.TotalSteps = len(phases)
			evaluation.CompletedSteps = currentIndex + 1 // 完成的步骤数
			
			// 计算阶段完成后的进度百分比
			progressPercent := float64(evaluation.CompletedSteps) / float64(evaluation.TotalSteps) * 100
			
			// 保存评估统计
			updateData := map[string]interface{}{
				"progress":         progressPercent,
				"current_phase":    phase,
				"progress_message": fmt.Sprintf("已完成: %s", phase),
				"total_steps":      evaluation.TotalSteps,
				"completed_steps":  evaluation.CompletedSteps,
			}
			
			err = global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
				Where("id = ?", evaluationID).Updates(updateData).Error
			if err != nil {
				global.GVA_LOG.Error("更新阶段完成状态失败", zap.Error(err))
			} else {
				global.GVA_LOG.Info("阶段进度更新", 
					zap.String("phase", phase),
					zap.Int("completedSteps", evaluation.CompletedSteps),
					zap.Int("totalSteps", evaluation.TotalSteps),
					zap.Float64("progress", progressPercent))
			}
		}
	}

	global.GVA_LOG.Info("阶段执行成功", 
		zap.String("phase", phase),
		zap.Duration("duration", duration))

	return nil
}

// initializeEvaluation 初始化评估
func (s *EvaluationProgressService) initializeEvaluation(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	global.GVA_LOG.Info("初始化评估环境", zap.Uint("evaluationId", evaluationID))
	
	// 初始化评估统计
	updates := map[string]interface{}{
		"status":              "processing",
		"total_steps":         9, // 总共9个阶段
		"completed_steps":     0,
		"processed_requirements": 0,
		"total_requirements":  0,
		"progress_message":    "评估初始化完成",
	}
	
	err := global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evaluationID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("初始化评估失败: %v", err)
	}
	
	// 模拟初始化耗时
	time.Sleep(500 * time.Millisecond)
	
	global.GVA_LOG.Info("评估初始化完成", zap.Uint("evaluationId", evaluationID))
	return nil
}

// loadRequirements 加载需求数据
func (s *EvaluationProgressService) loadRequirements(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	// 获取关联的需求数据
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Where("project_id = ?", evaluation.ProjectID)
	
	if evaluation.CycleID != nil {
		query = query.Where("cycle_id = ?", *evaluation.CycleID)
	}
	
	if evaluation.RequirementVersionID != nil {
		query = query.Where("version_id = ?", *evaluation.RequirementVersionID)
	}

	err := query.Find(&requirements).Error
	if err != nil {
		return fmt.Errorf("获取需求数据失败: %v", err)
	}

	// 更新需求统计
	evaluation.TotalRequirements = len(requirements)
	evaluation.ProcessedRequirements = 0

	err = global.GVA_DB.Save(evaluation).Error
	if err != nil {
		return fmt.Errorf("更新需求统计失败: %v", err)
	}

	global.GVA_LOG.Info("需求数据加载完成", 
		zap.Uint("evaluationId", evaluationID),
		zap.Int("requirementCount", len(requirements)))

	return nil
}

// identifyFunctionPoints 功能点识别
func (s *EvaluationProgressService) identifyFunctionPoints(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	// 调用现有的评估服务进行功能点识别
	evaluationService := EvaluationService{}
	evaluationService.performAutoEvaluation(evaluationID, []nesma.NesmaRequirement{})
	return nil
}

// analyzeComplexity 复杂度分析
func (s *EvaluationProgressService) analyzeComplexity(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	evaluationService := EvaluationService{}
	evaluationService.calculateComplexityMetrics(evaluationID, []nesma.NesmaRequirement{})
	return nil
}

// calculateWeights 权重计算
func (s *EvaluationProgressService) calculateWeights(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	// 获取功能点记录
	var functionPoints []nesma.NesmaFunctionPoint
	err := global.GVA_DB.Where("evaluation_id = ?", evaluationID).Find(&functionPoints).Error
	if err != nil {
		return fmt.Errorf("获取功能点记录失败: %v", err)
	}

	// 更新权重计算
	for i := range functionPoints {
		fp := &functionPoints[i]
		// 这里已经在创建时使用了配置的权重因子
		fp.CalculatedPoints = fp.WeightFactor
	}

	// 批量更新
	for _, fp := range functionPoints {
		err = global.GVA_DB.Save(&fp).Error
		if err != nil {
			global.GVA_LOG.Error("更新功能点权重失败", zap.Error(err))
		}
	}

	return nil
}

// calculateAdjustmentFactors 调整因子计算
func (s *EvaluationProgressService) calculateAdjustmentFactors(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	evaluationService := EvaluationService{}
	adjustment := evaluationService.calculateAdjustmentFactor(evaluationID)
	
	evaluation.AdjustmentFactor = adjustment
	evaluation.AdjustedFunctionPoints = evaluation.TotalFunctionPoints * adjustment
	
	return global.GVA_DB.Save(evaluation).Error
}

// performValidation 执行验证检查
func (s *EvaluationProgressService) performValidation(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	evaluationService := EvaluationService{}
	evaluationService.performValidationChecks(evaluationID)
	return nil
}

// summarizeResults 汇总结果
func (s *EvaluationProgressService) summarizeResults(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	evaluationService := EvaluationService{}
	evaluationService.recalculateEvaluationTotals(evaluationID)
	return nil
}

// completeEvaluation 完成评估
func (s *EvaluationProgressService) completeEvaluation(evaluationID uint, evaluation *nesma.NesmaEvaluation) error {
	now := time.Now()
	
	// 设置完成状态
	evaluation.Status = "completed"
	evaluation.Progress = 100
	evaluation.CompletionTime = &now
	evaluation.CurrentPhase = "完成评估"
	evaluation.ProgressMessage = "评估已完成"
	
	// 确保所有步骤都标记为完成
	phases := evaluation.GetProgressPhases()
	evaluation.TotalSteps = len(phases)
	evaluation.CompletedSteps = len(phases) // 所有步骤都已完成
	
	if evaluation.StartTime != nil {
		evaluation.Duration = int(now.Sub(*evaluation.StartTime).Seconds())
	}
	
	global.GVA_LOG.Info("评估完成", 
		zap.Uint("evaluationId", evaluationID),
		zap.String("currentPhase", evaluation.CurrentPhase),
		zap.Int("completedSteps", evaluation.CompletedSteps),
		zap.Int("totalSteps", evaluation.TotalSteps),
		zap.Float64("progress", evaluation.Progress))

	return global.GVA_DB.Save(evaluation).Error
}

// handleEvaluationError 处理评估错误
func (s *EvaluationProgressService) handleEvaluationError(evaluationID uint, err error) {
	var evaluation nesma.NesmaEvaluation
	if dbErr := global.GVA_DB.First(&evaluation, evaluationID).Error; dbErr != nil {
		global.GVA_LOG.Error("获取评估记录失败", zap.Error(dbErr))
		return
	}

	evaluation.Status = "failed"
	evaluation.ErrorMessage = err.Error()
	evaluation.ProgressMessage = "评估失败: " + err.Error()

	if saveErr := global.GVA_DB.Save(&evaluation).Error; saveErr != nil {
		global.GVA_LOG.Error("保存错误状态失败", zap.Error(saveErr))
	}

	global.GVA_LOG.Error("评估流程失败", 
		zap.Uint("evaluationId", evaluationID),
		zap.Error(err))
}

// CancelEvaluation 取消评估
func (s *EvaluationProgressService) CancelEvaluation(evaluationID uint, reason string) error {
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationID).Error
	if err != nil {
		return fmt.Errorf("评估记录不存在: %v", err)
	}

	if evaluation.Status != "processing" {
		return fmt.Errorf("只能取消正在运行的评估")
	}

	// 首先更新数据库状态
	evaluation.Status = "cancelled"
	evaluation.ErrorMessage = reason
	evaluation.ProgressMessage = "评估已被取消"
	
	err = global.GVA_DB.Save(&evaluation).Error
	if err != nil {
		return fmt.Errorf("更新评估状态失败: %v", err)
	}

	// 然后取消正在运行的评估上下文
	s.mu.RLock()
	cancelFunc, exists := s.runningEvaluations[evaluationID]
	s.mu.RUnlock()
	
	if exists && cancelFunc != nil {
		global.GVA_LOG.Info("取消正在运行的评估上下文", 
			zap.Uint("evaluationID", evaluationID),
			zap.String("reason", reason))
		cancelFunc() // 调用取消函数
	} else {
		global.GVA_LOG.Info("评估上下文不存在或已结束", 
			zap.Uint("evaluationID", evaluationID))
	}

	return nil
}

// GetRunningEvaluations 获取正在运行的评估列表
func (s *EvaluationProgressService) GetRunningEvaluations() ([]nesmaRes.RunningEvaluationResponse, error) {
	s.mu.RLock()
	runningIDs := make([]uint, 0, len(s.runningEvaluations))
	for evaluationID := range s.runningEvaluations {
		runningIDs = append(runningIDs, evaluationID)
	}
	s.mu.RUnlock()
	
	if len(runningIDs) == 0 {
		return []nesmaRes.RunningEvaluationResponse{}, nil
	}
	
	// 查询数据库获取详细信息
	var evaluations []nesma.NesmaEvaluation
	err := global.GVA_DB.Preload("Project").Where("id IN ?", runningIDs).Find(&evaluations).Error
	if err != nil {
		return nil, fmt.Errorf("查询运行中评估失败: %v", err)
	}
	
	runningEvaluations := make([]nesmaRes.RunningEvaluationResponse, 0, len(evaluations))
	for _, eval := range evaluations {
		elapsedTime := int64(0)
		if eval.StartTime != nil {
			elapsedTime = int64(time.Since(*eval.StartTime).Seconds())
		}
		
		// 估算完成时间
		var estimatedTime *time.Time
		if eval.Progress > 0 && eval.StartTime != nil {
			totalEstimated := time.Duration(float64(elapsedTime) / eval.Progress * 100) * time.Second
			estimated := eval.StartTime.Add(totalEstimated)
			estimatedTime = &estimated
		}
		
		runningEval := nesmaRes.RunningEvaluationResponse{
			EvaluationID:     eval.ID,
			ProjectID:        eval.ProjectID,
			ProjectName:      eval.Project.Name,
			Status:          eval.Status,
			Progress:        eval.Progress,
			CurrentPhase:    eval.CurrentPhase,
			StartTime:       *eval.StartTime,
			ElapsedTime:     elapsedTime,
			EstimatedTime:   estimatedTime,
			CanCancel:       eval.Status == "processing",
			ThreadID:        fmt.Sprintf("thread-%d", eval.ID),
		}
		runningEvaluations = append(runningEvaluations, runningEval)
	}
	
	return runningEvaluations, nil
}

// IsEvaluationRunning 检查指定评估是否正在运行
func (s *EvaluationProgressService) IsEvaluationRunning(evaluationID uint) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	_, exists := s.runningEvaluations[evaluationID]
	return exists
}

// GetEvaluationStatus 获取评估状态
func (s *EvaluationProgressService) GetEvaluationStatus(evaluationID uint) (*nesmaRes.EvaluationStatusResponse, error) {
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationID).Error
	if err != nil {
		return nil, err
	}

	response := &nesmaRes.EvaluationStatusResponse{
		EvaluationID: evaluation.ID,
		Status:       evaluation.Status,
		Progress:     evaluation.Progress,
		CurrentPhase: evaluation.CurrentPhase,
		IsRunning:    evaluation.Status == "processing",
		CanCancel:    evaluation.Status == "processing",
		CanRestart:   evaluation.Status == "failed" || evaluation.Status == "cancelled",
		LastUpdate:   evaluation.UpdatedAt,
	}

	return response, nil
}

// GetEvaluationStats 获取评估统计
func (s *EvaluationProgressService) GetEvaluationStats() (*nesmaRes.EvaluationSystemStatsResponse, error) {
	stats := &nesmaRes.EvaluationSystemStatsResponse{}

	// 总数
	err := global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Count(&stats.TotalEvaluations).Error
	if err != nil {
		return nil, err
	}

	// 按状态统计
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("status = ?", "processing").Count(&stats.RunningEvaluations)
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("status = ?", "completed").Count(&stats.CompletedEvaluations)
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("status = ?", "failed").Count(&stats.FailedEvaluations)
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Where("status = ?", "cancelled").Count(&stats.CancelledEvaluations)

	// 平均进度
	var avgProgress float64
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).Select("AVG(progress)").Scan(&avgProgress)
	stats.AverageProgress = avgProgress

	// 平均持续时间 (计算已完成评估的持续时间)
	var avgDuration float64
	global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("status IN ? AND start_time IS NOT NULL AND completion_time IS NOT NULL", []string{"completed", "failed"}).
		Select("AVG(EXTRACT(EPOCH FROM (completion_time - start_time)))").
		Scan(&avgDuration)
	stats.AverageDuration = int64(avgDuration)

	// 系统负载和内存使用 (模拟数据，实际可以从系统监控获取)
	stats.SystemLoad = float64(len(s.runningEvaluations)) / 10.0 // 假设最大并发10个
	stats.MemoryUsage = int64(len(s.runningEvaluations) * 50)    // 每个评估假设50MB

	return stats, nil
}

// GetEvaluationLogs 获取评估日志
func (s *EvaluationProgressService) GetEvaluationLogs(evaluationID uint, page, pageSize int, level string) (*nesmaRes.EvaluationLogResponse, error) {
	// 这里实现一个简化的日志查询
	// 实际项目中可能需要从专门的日志系统获取
	logs := []nesmaRes.EvaluationLogEntry{
		{
			Timestamp: time.Now().Add(-time.Hour),
			Level:     "info",
			Phase:     "初始化评估",
			Message:   "评估开始",
		},
		{
			Timestamp: time.Now().Add(-time.Minute * 30),
			Level:     "info",
			Phase:     "功能点识别",
			Message:   "正在识别功能点",
		},
	}

	response := &nesmaRes.EvaluationLogResponse{
		EvaluationID: evaluationID,
		Logs:         logs,
		TotalCount:   len(logs),
		PageSize:     pageSize,
		CurrentPage:  page,
	}

	return response, nil
}

// UpdateProgress 更新进度
func (s *EvaluationProgressService) UpdateProgress(evaluationID uint, progress float64, currentPhase, message string) error {
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationID).Error
	if err != nil {
		return err
	}

	// 更新进度：-1表示不更新进度，只更新阶段状态
	if progress >= 0 {
		evaluation.Progress = progress
	}
	
	// 更新当前阶段
	if currentPhase != "" {
		evaluation.SetPhase(currentPhase)
		// 如果没有明确指定进度，根据阶段自动计算
		if progress < 0 {
			evaluation.UpdateProgress()
		}
	}
	
	// 更新消息
	if message != "" {
		evaluation.ProgressMessage = message
	}

	// 记录进度更新日志
	global.GVA_LOG.Info("更新评估进度", 
		zap.Uint("evaluationId", evaluationID),
		zap.Float64("progress", evaluation.Progress),
		zap.String("phase", evaluation.CurrentPhase),
		zap.String("message", message))

	return global.GVA_DB.Save(&evaluation).Error
}

// buildPhaseInfo 构建阶段信息
func (s *EvaluationProgressService) buildPhaseInfo(evaluation *nesma.NesmaEvaluation) []nesmaRes.PhaseInfo {
	phases := evaluation.GetProgressPhases()
	currentIndex := evaluation.GetCurrentPhaseIndex()
	
	// 特殊处理：如果评估已完成但当前阶段索引不正确，则设置为最后一个阶段
	if evaluation.Status == "completed" && currentIndex != len(phases)-1 {
		currentIndex = len(phases) - 1
	}
	
	phaseInfos := make([]nesmaRes.PhaseInfo, len(phases))
	for i, phase := range phases {
		status := "pending"
		
		// 根据评估整体状态和当前阶段索引确定每个阶段的状态
		switch evaluation.Status {
		case "completed":
			// 所有阶段都已完成
			status = "completed"
		case "failed":
			if i < currentIndex {
				status = "completed"
			} else if i == currentIndex {
				status = "failed"
			} else {
				status = "pending"
			}
		case "processing":
			if i < currentIndex {
				status = "completed"
			} else if i == currentIndex {
				status = "processing"
			} else {
				status = "pending"
			}
		default: // pending状态
			if i < currentIndex {
				status = "completed"
			} else if i == currentIndex {
				status = "processing"
			} else {
				status = "pending"
			}
		}

		phaseInfos[i] = nesmaRes.PhaseInfo{
			Name:   phase,
			Index:  i,
			Status: status,
		}
	}

	return phaseInfos
}

// ==================== 带上下文的评估阶段方法 ====================

// initializeEvaluationWithContext 初始化评估(带上下文)
func (s *EvaluationProgressService) initializeEvaluationWithContext(evalCtx *ProgressEvaluationContext) error {
	global.GVA_LOG.Info("初始化评估环境", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	// 初始化评估统计
	updates := map[string]interface{}{
		"status":              "processing",
		"total_steps":         9, // 总共9个阶段
		"completed_steps":     0,
		"processed_requirements": 0,
		"total_requirements":  0,
		"progress_message":    "评估初始化完成",
	}
	
	err := global.GVA_DB.Model(&nesma.NesmaEvaluation{}).
		Where("id = ?", evalCtx.Evaluation.ID).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("初始化评估失败: %v", err)
	}
	
	// 模拟初始化耗时
	time.Sleep(500 * time.Millisecond)
	
	global.GVA_LOG.Info("评估初始化完成", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	return nil
}

// loadRequirementsWithContext 加载需求数据(带上下文)
func (s *EvaluationProgressService) loadRequirementsWithContext(evalCtx *ProgressEvaluationContext) error {
	// 检查取消状态
	if err := s.checkCancellationWithContext(evalCtx); err != nil {
		return err
	}
	
	// 获取关联的需求数据
	var requirements []nesma.NesmaRequirement
	query := global.GVA_DB.Where("project_id = ?", evalCtx.Evaluation.ProjectID)
	
	if evalCtx.Evaluation.CycleID != nil {
		query = query.Where("cycle_id = ?", *evalCtx.Evaluation.CycleID)
	}
	
	if evalCtx.Evaluation.RequirementVersionID != nil {
		query = query.Where("version_id = ?", *evalCtx.Evaluation.RequirementVersionID)
	}

	err := query.Find(&requirements).Error
	if err != nil {
		return fmt.Errorf("获取需求数据失败: %v", err)
	}

	// 再次检查取消状态
	if err := s.checkCancellationWithContext(evalCtx); err != nil {
		return err
	}

	// 保存需求数据到上下文中
	evalCtx.Requirements = requirements
	
	// 更新需求统计
	evalCtx.Evaluation.TotalRequirements = len(requirements)
	evalCtx.Evaluation.ProcessedRequirements = 0

	err = global.GVA_DB.Save(evalCtx.Evaluation).Error
	if err != nil {
		return fmt.Errorf("更新需求统计失败: %v", err)
	}

	global.GVA_LOG.Info("需求数据加载完成", 
		zap.Uint("evaluationId", evalCtx.Evaluation.ID),
		zap.Int("requirementCount", len(requirements)))

	return nil
}

// identifyFunctionPointsWithContext 功能点识别(带上下文)
func (s *EvaluationProgressService) identifyFunctionPointsWithContext(evalCtx *ProgressEvaluationContext) error {
	// 检查取消状态
	if err := s.checkCancellationWithContext(evalCtx); err != nil {
		return err
	}
	
	global.GVA_LOG.Info("开始功能点识别", 
		zap.Uint("evaluationId", evalCtx.Evaluation.ID),
		zap.Int("requirementCount", len(evalCtx.Requirements)))
	
	// 现在我们有了实际的需求数据！
	if len(evalCtx.Requirements) == 0 {
		global.GVA_LOG.Warn("没有找到需求数据进行分析", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
		return nil
	}
	
	// 在长时间操作前再次检查取消状态
	if err := s.checkCancellationWithContext(evalCtx); err != nil {
		return err
	}
	
	// 调用现有的评估服务，但传入实际的需求数据
	// TODO: 这里应该将evaluationService也改为支持取消的版本
	evaluationService := EvaluationService{}
	evaluationService.performAutoEvaluation(evalCtx.Evaluation.ID, evalCtx.Requirements)
	
	// 完成后检查取消状态
	if err := s.checkCancellationWithContext(evalCtx); err != nil {
		return err
	}
	
	global.GVA_LOG.Info("功能点识别完成", 
		zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	return nil
}

// analyzeComplexityWithContext 复杂度分析(带上下文)
func (s *EvaluationProgressService) analyzeComplexityWithContext(evalCtx *ProgressEvaluationContext) error {
	global.GVA_LOG.Info("开始复杂度分析", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	evaluationService := EvaluationService{}
	evaluationService.calculateComplexityMetrics(evalCtx.Evaluation.ID, evalCtx.Requirements)
	
	return nil
}

// calculateWeightsWithContext 权重计算(带上下文)
func (s *EvaluationProgressService) calculateWeightsWithContext(evalCtx *ProgressEvaluationContext) error {
	global.GVA_LOG.Info("开始权重计算", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	// 模拟权重计算处理
	time.Sleep(1 * time.Second)
	
	return nil
}

// calculateAdjustmentFactorsWithContext 调整因子计算(带上下文)
func (s *EvaluationProgressService) calculateAdjustmentFactorsWithContext(evalCtx *ProgressEvaluationContext) error {
	global.GVA_LOG.Info("开始调整因子计算", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	// 模拟调整因子计算处理
	time.Sleep(1 * time.Second)
	
	return nil
}

// performValidationWithContext 验证检查(带上下文)
func (s *EvaluationProgressService) performValidationWithContext(evalCtx *ProgressEvaluationContext) error {
	global.GVA_LOG.Info("开始验证检查", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	// 模拟验证检查处理
	time.Sleep(1 * time.Second)
	
	return nil
}

// summarizeResultsWithContext 结果汇总(带上下文)
func (s *EvaluationProgressService) summarizeResultsWithContext(evalCtx *ProgressEvaluationContext) error {
	global.GVA_LOG.Info("开始结果汇总", zap.Uint("evaluationId", evalCtx.Evaluation.ID))
	
	// 模拟结果汇总处理
	time.Sleep(1 * time.Second)
	
	return nil
}

// completeEvaluationWithContext 完成评估(带上下文)
func (s *EvaluationProgressService) completeEvaluationWithContext(evalCtx *ProgressEvaluationContext) error {
	now := time.Now()
	
	// 设置完成状态
	evalCtx.Evaluation.Status = "completed"
	evalCtx.Evaluation.Progress = 100
	evalCtx.Evaluation.CompletionTime = &now
	evalCtx.Evaluation.CurrentPhase = "完成评估"
	evalCtx.Evaluation.ProgressMessage = "评估已完成"
	
	// 确保所有步骤都标记为完成
	phases := evalCtx.Evaluation.GetProgressPhases()
	evalCtx.Evaluation.TotalSteps = len(phases)
	evalCtx.Evaluation.CompletedSteps = len(phases) // 所有步骤都已完成
	
	if evalCtx.Evaluation.StartTime != nil {
		evalCtx.Evaluation.Duration = int(now.Sub(*evalCtx.Evaluation.StartTime).Seconds())
	}
	
	global.GVA_LOG.Info("评估完成", 
		zap.Uint("evaluationId", evalCtx.Evaluation.ID),
		zap.String("currentPhase", evalCtx.Evaluation.CurrentPhase),
		zap.Int("completedSteps", evalCtx.Evaluation.CompletedSteps),
		zap.Int("totalSteps", evalCtx.Evaluation.TotalSteps),
		zap.Float64("progress", evalCtx.Evaluation.Progress))

	return global.GVA_DB.Save(evalCtx.Evaluation).Error
}

// ForceStopEvaluation 强制停止评估
func (s *EvaluationProgressService) ForceStopEvaluation(evaluationID uint, reason string) error {
	global.GVA_LOG.Info("强制停止评估", 
		zap.Uint("evaluationID", evaluationID),
		zap.String("reason", reason))

	// 更新数据库状态
	var evaluation nesma.NesmaEvaluation
	err := global.GVA_DB.First(&evaluation, evaluationID).Error
	if err != nil {
		return fmt.Errorf("查找评估记录失败: %v", err)
	}

	// 强制停止：无论当前状态如何都停止
	evaluation.Status = "cancelled"
	evaluation.ErrorMessage = fmt.Sprintf("强制停止: %s", reason)
	evaluation.ProgressMessage = "评估已被强制停止"
	evaluation.CompletionTime = &[]time.Time{time.Now()}[0]
	
	if evaluation.StartTime != nil {
		evaluation.Duration = int(time.Since(*evaluation.StartTime).Seconds())
	}
	
	err = global.GVA_DB.Save(&evaluation).Error
	if err != nil {
		return fmt.Errorf("更新评估状态失败: %v", err)
	}

	// 停止正在运行的评估上下文
	s.mu.Lock()
	cancelFunc, exists := s.runningEvaluations[evaluationID]
	if exists {
		delete(s.runningEvaluations, evaluationID) // 立即移除
		s.mu.Unlock()
		
		if cancelFunc != nil {
			global.GVA_LOG.Info("强制取消正在运行的评估上下文", 
				zap.Uint("evaluationID", evaluationID))
			cancelFunc() // 调用取消函数
		}
	} else {
		s.mu.Unlock()
		global.GVA_LOG.Info("评估上下文不存在，仅更新数据库状态", 
			zap.Uint("evaluationID", evaluationID))
	}

	return nil
}