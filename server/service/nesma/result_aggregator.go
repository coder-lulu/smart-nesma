package nesma

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ResultAggregator 结果聚合器
type ResultAggregator struct {
	db                *gorm.DB
	tokenCalculator   *TokenCalculator
	
	// 配置
	config            *AggregatorConfig
	
	// 缓存
	aggregationCache  map[string]*AggregationResult
	cacheMu          sync.RWMutex
	
	// 指标
	metrics          *AggregatorMetrics
}

// AggregatorConfig 聚合器配置
type AggregatorConfig struct {
	// 聚合策略
	AggregationStrategy    string  `json:"aggregation_strategy"`    // weighted/average/consensus/ml_based
	ConsistencyThreshold   float64 `json:"consistency_threshold"`   // 一致性阈值
	QualityThreshold       float64 `json:"quality_threshold"`       // 质量阈值
	
	// 权重配置
	ConfidenceWeight       float64 `json:"confidence_weight"`       // 置信度权重
	ModelPerformanceWeight float64 `json:"model_performance_weight"` // 模型性能权重
	HistoryWeight          float64 `json:"history_weight"`          // 历史记录权重
	
	// 冲突解决
	ConflictResolution     string  `json:"conflict_resolution"`     // vote/highest_confidence/model_priority
	MinConsensusRatio      float64 `json:"min_consensus_ratio"`     // 最小共识比例
	
	// 质量控制
	EnableQualityCheck     bool    `json:"enable_quality_check"`
	EnableConsistencyCheck bool    `json:"enable_consistency_check"`
	EnableOutlierDetection bool    `json:"enable_outlier_detection"`
	
	// 输出控制
	DetailLevel           string   `json:"detail_level"`            // summary/detailed/full
	IncludeMetrics        bool     `json:"include_metrics"`
	ExportFormat          []string `json:"export_format"`           // json/excel/pdf
}

// AggregationResult 聚合结果
type AggregationResult struct {
	// 基本信息
	EvaluationID        string                    `json:"evaluation_id"`
	ProjectID           uint                      `json:"project_id"`
	CycleID             uint                      `json:"cycle_id"`
	AggregationTime     time.Time                 `json:"aggregation_time"`
	
	// 聚合配置
	Strategy            string                    `json:"strategy"`
	QualityScore        float64                   `json:"quality_score"`
	ConsistencyScore    float64                   `json:"consistency_score"`
	
	// 需求结果
	RequirementResults  []AggregatedRequirement   `json:"requirement_results"`
	
	// 统计信息
	TotalRequirements   int                       `json:"total_requirements"`
	ProcessedCount      int                       `json:"processed_count"`
	SuccessCount        int                       `json:"success_count"`
	ConflictCount       int                       `json:"conflict_count"`
	
	// 功能点统计
	FunctionTypeStats   map[string]FunctionTypeStats `json:"function_type_stats"`
	ComplexityStats     map[string]ComplexityStats   `json:"complexity_stats"`
	AFPTotal            float64                      `json:"afp_total"`
	UFPTotal            float64                      `json:"ufp_total"`
	
	// 质量分析
	QualityAnalysis     *QualityAnalysis             `json:"quality_analysis"`
	ConsistencyAnalysis *ConsistencyAnalysis         `json:"consistency_analysis"`
	
	// 异常检测
	OutlierDetection    *OutlierDetection            `json:"outlier_detection"`
	
	// 改进建议
	Recommendations     []Recommendation             `json:"recommendations"`
	
	// 元数据
	Metadata            map[string]interface{}       `json:"metadata"`
}

// AggregatedRequirement 聚合需求结果
type AggregatedRequirement struct {
	// 需求信息
	RequirementID       uint      `json:"requirement_id"`
	RequirementCode     string    `json:"requirement_code"`
	RequirementTitle    string    `json:"requirement_title"`
	Level               int       `json:"level"`
	
	// 聚合结果
	FunctionType        string    `json:"function_type"`
	ComplexityLevel     string    `json:"complexity_level"`
	AFP                 float64   `json:"afp"`
	UFP                 float64   `json:"ufp"`
	
	// 置信度和质量
	ConfidenceScore     float64   `json:"confidence_score"`
	QualityScore        float64   `json:"quality_score"`
	ConsistencyScore    float64   `json:"consistency_score"`
	
	// 聚合统计
	EvaluationCount     int       `json:"evaluation_count"`
	ConsensusRatio      float64   `json:"consensus_ratio"`
	
	// 候选结果
	CandidateResults    []CandidateResult `json:"candidate_results"`
	
	// 冲突信息
	HasConflict         bool      `json:"has_conflict"`
	ConflictReason      string    `json:"conflict_reason"`
	ConflictResolution  string    `json:"conflict_resolution"`
	
	// 改进建议
	Suggestions         []string  `json:"suggestions"`
	
	// 元数据
	Metadata            map[string]interface{} `json:"metadata"`
}

// CandidateResult 候选结果
type CandidateResult struct {
	FunctionType     string    `json:"function_type"`
	ComplexityLevel  string    `json:"complexity_level"`
	AFP              float64   `json:"afp"`
	UFP              float64   `json:"ufp"`
	ConfidenceScore  float64   `json:"confidence_score"`
	Weight           float64   `json:"weight"`
	Source           string    `json:"source"`
	BatchID          string    `json:"batch_id"`
	ModelUsed        string    `json:"model_used"`
	Timestamp        time.Time `json:"timestamp"`
}

// FunctionTypeStats 功能类型统计
type FunctionTypeStats struct {
	Count            int     `json:"count"`
	Percentage       float64 `json:"percentage"`
	TotalAFP         float64 `json:"total_afp"`
	TotalUFP         float64 `json:"total_ufp"`
	AvgComplexity    float64 `json:"avg_complexity"`
	AvgConfidence    float64 `json:"avg_confidence"`
	QualityScore     float64 `json:"quality_score"`
}

// ComplexityStats 复杂度统计
type ComplexityStats struct {
	Count            int     `json:"count"`
	Percentage       float64 `json:"percentage"`
	TotalAFP         float64 `json:"total_afp"`
	TotalUFP         float64 `json:"total_ufp"`
	AvgConfidence    float64 `json:"avg_confidence"`
	QualityScore     float64 `json:"quality_score"`
}

// QualityAnalysis 质量分析
type QualityAnalysis struct {
	OverallQuality      float64                `json:"overall_quality"`
	QualityDistribution map[string]int         `json:"quality_distribution"`
	QualityTrends       []QualityTrend         `json:"quality_trends"`
	QualityIssues       []QualityIssue         `json:"quality_issues"`
	ModelPerformance    map[string]float64     `json:"model_performance"`
	ImprovementAreas    []string               `json:"improvement_areas"`
}

// ConsistencyAnalysis 一致性分析
type ConsistencyAnalysis struct {
	OverallConsistency   float64                `json:"overall_consistency"`
	ConsistencyByType    map[string]float64     `json:"consistency_by_type"`
	ConsistencyByLevel   map[int]float64        `json:"consistency_by_level"`
	InconsistentResults  []InconsistentResult   `json:"inconsistent_results"`
	ConsistencyTrends    []ConsistencyTrend     `json:"consistency_trends"`
	ConsistencyIssues    []ConsistencyIssue     `json:"consistency_issues"`
}

// OutlierDetection 异常检测
type OutlierDetection struct {
	OutlierCount        int                 `json:"outlier_count"`
	OutlierPercentage   float64             `json:"outlier_percentage"`
	OutlierResults      []OutlierResult     `json:"outlier_results"`
	DetectionMethods    []string            `json:"detection_methods"`
	Threshold           float64             `json:"threshold"`
	ActionTaken         string              `json:"action_taken"`
}

// QualityIssue 质量问题
type QualityIssue struct {
	IssueType       string    `json:"issue_type"`
	Severity        string    `json:"severity"`
	Description     string    `json:"description"`
	AffectedCount   int       `json:"affected_count"`
	RequirementIDs  []uint    `json:"requirement_ids"`
	Recommendations []string  `json:"recommendations"`
}

// InconsistentResult 不一致结果
type InconsistentResult struct {
	RequirementID    uint      `json:"requirement_id"`
	RequirementCode  string    `json:"requirement_code"`
	ConflictType     string    `json:"conflict_type"`
	CandidateCount   int       `json:"candidate_count"`
	VarianceScore    float64   `json:"variance_score"`
	Resolution       string    `json:"resolution"`
	Confidence       float64   `json:"confidence"`
}

// ConsistencyTrend 一致性趋势
type ConsistencyTrend struct {
	BatchNumber      int       `json:"batch_number"`
	ConsistencyScore float64   `json:"consistency_score"`
	RequirementCount int       `json:"requirement_count"`
	Timestamp        time.Time `json:"timestamp"`
}

// ConsistencyIssue 一致性问题
type ConsistencyIssue struct {
	IssueType       string    `json:"issue_type"`
	Severity        string    `json:"severity"`
	Description     string    `json:"description"`
	AffectedCount   int       `json:"affected_count"`
	RequirementIDs  []uint    `json:"requirement_ids"`
	Recommendations []string  `json:"recommendations"`
}

// OutlierResult 异常结果
type OutlierResult struct {
	RequirementID    uint      `json:"requirement_id"`
	RequirementCode  string    `json:"requirement_code"`
	OutlierType      string    `json:"outlier_type"`
	DeviationScore   float64   `json:"deviation_score"`
	ExpectedValue    float64   `json:"expected_value"`
	ActualValue      float64   `json:"actual_value"`
	DetectionMethod  string    `json:"detection_method"`
	ActionTaken      string    `json:"action_taken"`
}

// Recommendation 改进建议
type Recommendation struct {
	Category        string    `json:"category"`
	Priority        string    `json:"priority"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	ActionItems     []string  `json:"action_items"`
	ExpectedImpact  string    `json:"expected_impact"`
	ImplementationTime string `json:"implementation_time"`
}

// AggregatorMetrics 聚合器指标
type AggregatorMetrics struct {
	TotalAggregations      int64               `json:"total_aggregations"`
	SuccessfulAggregations int64               `json:"successful_aggregations"`
	FailedAggregations     int64               `json:"failed_aggregations"`
	TotalRequirements      int64               `json:"total_requirements"`
	TotalConflicts         int64               `json:"total_conflicts"`
	AverageQualityScore    float64             `json:"average_quality_score"`
	AverageConsistencyScore float64            `json:"average_consistency_score"`
	AverageProcessingTime  time.Duration       `json:"average_processing_time"`
	StrategyUsageStats     map[string]int64    `json:"strategy_usage_stats"`
	mu                     sync.RWMutex        `json:"-"`
}

// NewResultAggregator 创建结果聚合器
func NewResultAggregator(db *gorm.DB, tokenCalculator *TokenCalculator) *ResultAggregator {
	config := &AggregatorConfig{
		AggregationStrategy:    "weighted",
		ConsistencyThreshold:   0.8,
		QualityThreshold:       0.7,
		ConfidenceWeight:       0.4,
		ModelPerformanceWeight: 0.3,
		HistoryWeight:          0.3,
		ConflictResolution:     "highest_confidence",
		MinConsensusRatio:      0.6,
		EnableQualityCheck:     true,
		EnableConsistencyCheck: true,
		EnableOutlierDetection: true,
		DetailLevel:           "detailed",
		IncludeMetrics:        true,
		ExportFormat:          []string{"json"},
	}
	
	return &ResultAggregator{
		db:              db,
		tokenCalculator: tokenCalculator,
		config:          config,
		aggregationCache: make(map[string]*AggregationResult),
		cacheMu:         sync.RWMutex{},
		metrics: &AggregatorMetrics{
			StrategyUsageStats: make(map[string]int64),
		},
	}
}

// AggregateResults 聚合结果
func (r *ResultAggregator) AggregateResults(evaluationID string, batchResults []EvaluationResult) (*AggregationResult, error) {
	startTime := time.Now()
	
	// 更新指标
	r.metrics.mu.Lock()
	r.metrics.TotalAggregations++
	r.metrics.StrategyUsageStats[r.config.AggregationStrategy]++
	r.metrics.mu.Unlock()
	
	global.GVA_LOG.Info("开始结果聚合",
		zap.String("evaluationID", evaluationID),
		zap.Int("批次结果数", len(batchResults)),
		zap.String("聚合策略", r.config.AggregationStrategy))
	
	// 检查缓存
	r.cacheMu.RLock()
	if cached, exists := r.aggregationCache[evaluationID]; exists {
		r.cacheMu.RUnlock()
		return cached, nil
	}
	r.cacheMu.RUnlock()
	
	// 按需求ID分组
	resultsByRequirement := r.groupResultsByRequirement(batchResults)
	
	// 获取项目信息
	projectID, cycleID, err := r.getProjectInfo(evaluationID)
	if err != nil {
		return nil, fmt.Errorf("获取项目信息失败: %w", err)
	}
	
	// 创建聚合结果
	aggregationResult := &AggregationResult{
		EvaluationID:       evaluationID,
		ProjectID:          projectID,
		CycleID:            cycleID,
		AggregationTime:    time.Now(),
		Strategy:           r.config.AggregationStrategy,
		RequirementResults: make([]AggregatedRequirement, 0),
		FunctionTypeStats:  make(map[string]FunctionTypeStats),
		ComplexityStats:    make(map[string]ComplexityStats),
		Metadata:           make(map[string]interface{}),
	}
	
	// 处理每个需求的结果
	for requirementID, results := range resultsByRequirement {
		aggregatedReq, err := r.aggregateRequirementResults(requirementID, results)
		if err != nil {
			global.GVA_LOG.Error("聚合需求结果失败",
				zap.Uint("requirementID", requirementID),
				zap.Error(err))
			continue
		}
		
		aggregationResult.RequirementResults = append(aggregationResult.RequirementResults, *aggregatedReq)
		aggregationResult.ProcessedCount++
		
		if !aggregatedReq.HasConflict {
			aggregationResult.SuccessCount++
		} else {
			aggregationResult.ConflictCount++
		}
	}
	
	// 计算统计信息
	r.calculateStatistics(aggregationResult)
	
	// 质量分析
	if r.config.EnableQualityCheck {
		aggregationResult.QualityAnalysis = r.performQualityAnalysis(aggregationResult)
	}
	
	// 一致性分析
	if r.config.EnableConsistencyCheck {
		aggregationResult.ConsistencyAnalysis = r.performConsistencyAnalysis(aggregationResult)
	}
	
	// 异常检测
	if r.config.EnableOutlierDetection {
		aggregationResult.OutlierDetection = r.performOutlierDetection(aggregationResult)
	}
	
	// 生成建议
	aggregationResult.Recommendations = r.generateRecommendations(aggregationResult)
	
	// 保存到缓存
	r.cacheMu.Lock()
	r.aggregationCache[evaluationID] = aggregationResult
	r.cacheMu.Unlock()
	
	// 更新指标
	r.metrics.mu.Lock()
	r.metrics.SuccessfulAggregations++
	r.metrics.TotalRequirements += int64(aggregationResult.ProcessedCount)
	r.metrics.TotalConflicts += int64(aggregationResult.ConflictCount)
	r.metrics.AverageProcessingTime = time.Since(startTime)
	r.metrics.mu.Unlock()
	
	global.GVA_LOG.Info("结果聚合完成",
		zap.String("evaluationID", evaluationID),
		zap.Int("处理需求数", aggregationResult.ProcessedCount),
		zap.Int("成功数", aggregationResult.SuccessCount),
		zap.Int("冲突数", aggregationResult.ConflictCount),
		zap.Float64("质量评分", aggregationResult.QualityScore),
		zap.Duration("处理时间", time.Since(startTime)))
	
	return aggregationResult, nil
}

// groupResultsByRequirement 按需求ID分组结果
func (r *ResultAggregator) groupResultsByRequirement(results []EvaluationResult) map[uint][]EvaluationResult {
	grouped := make(map[uint][]EvaluationResult)
	
	for _, result := range results {
		grouped[result.RequirementID] = append(grouped[result.RequirementID], result)
	}
	
	return grouped
}

// aggregateRequirementResults 聚合单个需求的结果
func (r *ResultAggregator) aggregateRequirementResults(requirementID uint, results []EvaluationResult) (*AggregatedRequirement, error) {
	if len(results) == 0 {
		return nil, fmt.Errorf("没有结果可以聚合")
	}
	
	// 获取需求信息
	req, err := r.getRequirementInfo(requirementID)
	if err != nil {
		return nil, fmt.Errorf("获取需求信息失败: %w", err)
	}
	
	// 创建聚合需求
	aggregatedReq := &AggregatedRequirement{
		RequirementID:    requirementID,
		RequirementCode:  req.Code,
		RequirementTitle: req.Title,
		Level:           req.Level,
		EvaluationCount: len(results),
		CandidateResults: make([]CandidateResult, 0),
		Suggestions:     make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}
	
	// 转换为候选结果
	for _, result := range results {
		candidate := CandidateResult{
			FunctionType:     result.FunctionType,
			ComplexityLevel:  result.ComplexityLevel,
			AFP:              result.AFP,
			UFP:              result.UFP,
			ConfidenceScore:  result.ConfidenceScore,
			Source:           "batch_evaluation",
			BatchID:          result.BatchID,
			ModelUsed:        result.ModelUsed,
			Timestamp:        result.Timestamp,
		}
		
		// 计算权重
		candidate.Weight = r.calculateCandidateWeight(candidate)
		
		aggregatedReq.CandidateResults = append(aggregatedReq.CandidateResults, candidate)
	}
	
	// 执行聚合
	switch r.config.AggregationStrategy {
	case "weighted":
		r.aggregateWeighted(aggregatedReq)
	case "average":
		r.aggregateAverage(aggregatedReq)
	case "consensus":
		r.aggregateConsensus(aggregatedReq)
	default:
		return nil, fmt.Errorf("不支持的聚合策略: %s", r.config.AggregationStrategy)
	}
	
	// 检测冲突
	r.detectConflicts(aggregatedReq)
	
	// 计算质量评分
	r.calculateQualityScore(aggregatedReq)
	
	// 生成建议
	r.generateRequirementSuggestions(aggregatedReq)
	
	return aggregatedReq, nil
}

// aggregateWeighted 加权聚合
func (r *ResultAggregator) aggregateWeighted(req *AggregatedRequirement) {
	if len(req.CandidateResults) == 0 {
		return
	}
	
	// 计算加权平均
	var totalWeight float64
	var weightedAFP, weightedUFP, weightedConfidence float64
	
	functionTypeVotes := make(map[string]float64)
	complexityVotes := make(map[string]float64)
	
	for _, candidate := range req.CandidateResults {
		weight := candidate.Weight
		totalWeight += weight
		
		weightedAFP += candidate.AFP * weight
		weightedUFP += candidate.UFP * weight
		weightedConfidence += candidate.ConfidenceScore * weight
		
		functionTypeVotes[candidate.FunctionType] += weight
		complexityVotes[candidate.ComplexityLevel] += weight
	}
	
	if totalWeight > 0 {
		req.AFP = weightedAFP / totalWeight
		req.UFP = weightedUFP / totalWeight
		req.ConfidenceScore = weightedConfidence / totalWeight
	}
	
	// 选择得票最高的类型
	req.FunctionType = r.getHighestVote(functionTypeVotes)
	req.ComplexityLevel = r.getHighestVote(complexityVotes)
}

// aggregateAverage 平均聚合
func (r *ResultAggregator) aggregateAverage(req *AggregatedRequirement) {
	if len(req.CandidateResults) == 0 {
		return
	}
	
	count := float64(len(req.CandidateResults))
	var totalAFP, totalUFP, totalConfidence float64
	
	functionTypeVotes := make(map[string]int)
	complexityVotes := make(map[string]int)
	
	for _, candidate := range req.CandidateResults {
		totalAFP += candidate.AFP
		totalUFP += candidate.UFP
		totalConfidence += candidate.ConfidenceScore
		
		functionTypeVotes[candidate.FunctionType]++
		complexityVotes[candidate.ComplexityLevel]++
	}
	
	req.AFP = totalAFP / count
	req.UFP = totalUFP / count
	req.ConfidenceScore = totalConfidence / count
	
	// 选择得票最多的类型
	req.FunctionType = r.getHighestVoteInt(functionTypeVotes)
	req.ComplexityLevel = r.getHighestVoteInt(complexityVotes)
}

// aggregateConsensus 共识聚合
func (r *ResultAggregator) aggregateConsensus(req *AggregatedRequirement) {
	if len(req.CandidateResults) == 0 {
		return
	}
	
	// 计算共识比例
	functionTypeVotes := make(map[string]int)
	complexityVotes := make(map[string]int)
	
	for _, candidate := range req.CandidateResults {
		functionTypeVotes[candidate.FunctionType]++
		complexityVotes[candidate.ComplexityLevel]++
	}
	
	total := len(req.CandidateResults)
	
	// 检查是否达到共识
	maxFunctionVotes := 0
	maxComplexityVotes := 0
	
	for _, votes := range functionTypeVotes {
		if votes > maxFunctionVotes {
			maxFunctionVotes = votes
		}
	}
	
	for _, votes := range complexityVotes {
		if votes > maxComplexityVotes {
			maxComplexityVotes = votes
		}
	}
	
	functionConsensus := float64(maxFunctionVotes) / float64(total)
	complexityConsensus := float64(maxComplexityVotes) / float64(total)
	
	req.ConsensusRatio = math.Min(functionConsensus, complexityConsensus)
	
	if req.ConsensusRatio >= r.config.MinConsensusRatio {
		// 有共识，使用共识结果
		req.FunctionType = r.getHighestVoteInt(functionTypeVotes)
		req.ComplexityLevel = r.getHighestVoteInt(complexityVotes)
		
		// 计算共识结果的平均值
		var consensusAFP, consensusUFP, consensusConfidence float64
		var consensusCount int
		
		for _, candidate := range req.CandidateResults {
			if candidate.FunctionType == req.FunctionType && candidate.ComplexityLevel == req.ComplexityLevel {
				consensusAFP += candidate.AFP
				consensusUFP += candidate.UFP
				consensusConfidence += candidate.ConfidenceScore
				consensusCount++
			}
		}
		
		if consensusCount > 0 {
			req.AFP = consensusAFP / float64(consensusCount)
			req.UFP = consensusUFP / float64(consensusCount)
			req.ConfidenceScore = consensusConfidence / float64(consensusCount)
		}
	} else {
		// 无共识，使用冲突解决策略
		r.resolveConflict(req)
	}
}

// calculateCandidateWeight 计算候选结果权重
func (r *ResultAggregator) calculateCandidateWeight(candidate CandidateResult) float64 {
	weight := 0.0
	
	// 置信度权重
	weight += candidate.ConfidenceScore * r.config.ConfidenceWeight
	
	// 模型性能权重（简化实现）
	modelPerformance := r.getModelPerformance(candidate.ModelUsed)
	weight += modelPerformance * r.config.ModelPerformanceWeight
	
	// 历史权重（基于时间新鲜度）
	timeFactor := r.calculateTimeFactor(candidate.Timestamp)
	weight += timeFactor * r.config.HistoryWeight
	
	return weight
}

// detectConflicts 检测冲突
func (r *ResultAggregator) detectConflicts(req *AggregatedRequirement) {
	if len(req.CandidateResults) <= 1 {
		return
	}
	
	// 检查功能类型冲突
	functionTypes := make(map[string]bool)
	complexityLevels := make(map[string]bool)
	
	var afpValues []float64
	var ufpValues []float64
	
	for _, candidate := range req.CandidateResults {
		functionTypes[candidate.FunctionType] = true
		complexityLevels[candidate.ComplexityLevel] = true
		afpValues = append(afpValues, candidate.AFP)
		ufpValues = append(ufpValues, candidate.UFP)
	}
	
	// 检查类型冲突
	if len(functionTypes) > 1 {
		req.HasConflict = true
		req.ConflictReason = "功能类型不一致"
	}
	
	if len(complexityLevels) > 1 {
		req.HasConflict = true
		if req.ConflictReason != "" {
			req.ConflictReason += ", "
		}
		req.ConflictReason += "复杂度级别不一致"
	}
	
	// 检查数值冲突
	afpVariance := r.calculateVariance(afpValues)
	ufpVariance := r.calculateVariance(ufpValues)
	
	if afpVariance > 0.5 || ufpVariance > 0.5 {
		req.HasConflict = true
		if req.ConflictReason != "" {
			req.ConflictReason += ", "
		}
		req.ConflictReason += "功能点数值差异过大"
	}
	
	// 如果有冲突，尝试解决
	if req.HasConflict {
		r.resolveConflict(req)
	}
}

// resolveConflict 解决冲突
func (r *ResultAggregator) resolveConflict(req *AggregatedRequirement) {
	switch r.config.ConflictResolution {
	case "highest_confidence":
		r.resolveByHighestConfidence(req)
	case "vote":
		r.resolveByVote(req)
	case "model_priority":
		r.resolveByModelPriority(req)
	default:
		r.resolveByHighestConfidence(req)
	}
	
	req.ConflictResolution = r.config.ConflictResolution
}

// resolveByHighestConfidence 按最高置信度解决冲突
func (r *ResultAggregator) resolveByHighestConfidence(req *AggregatedRequirement) {
	if len(req.CandidateResults) == 0 {
		return
	}
	
	// 找到置信度最高的结果
	maxConfidence := 0.0
	bestCandidate := req.CandidateResults[0]
	
	for _, candidate := range req.CandidateResults {
		if candidate.ConfidenceScore > maxConfidence {
			maxConfidence = candidate.ConfidenceScore
			bestCandidate = candidate
		}
	}
	
	req.FunctionType = bestCandidate.FunctionType
	req.ComplexityLevel = bestCandidate.ComplexityLevel
	req.AFP = bestCandidate.AFP
	req.UFP = bestCandidate.UFP
	req.ConfidenceScore = bestCandidate.ConfidenceScore
}

// resolveByVote 按投票解决冲突
func (r *ResultAggregator) resolveByVote(req *AggregatedRequirement) {
	functionTypeVotes := make(map[string]int)
	complexityVotes := make(map[string]int)
	
	for _, candidate := range req.CandidateResults {
		functionTypeVotes[candidate.FunctionType]++
		complexityVotes[candidate.ComplexityLevel]++
	}
	
	req.FunctionType = r.getHighestVoteInt(functionTypeVotes)
	req.ComplexityLevel = r.getHighestVoteInt(complexityVotes)
	
	// 计算投票结果的平均值
	var totalAFP, totalUFP, totalConfidence float64
	var count int
	
	for _, candidate := range req.CandidateResults {
		if candidate.FunctionType == req.FunctionType && candidate.ComplexityLevel == req.ComplexityLevel {
			totalAFP += candidate.AFP
			totalUFP += candidate.UFP
			totalConfidence += candidate.ConfidenceScore
			count++
		}
	}
	
	if count > 0 {
		req.AFP = totalAFP / float64(count)
		req.UFP = totalUFP / float64(count)
		req.ConfidenceScore = totalConfidence / float64(count)
	}
}

// resolveByModelPriority 按模型优先级解决冲突
func (r *ResultAggregator) resolveByModelPriority(req *AggregatedRequirement) {
	// 模型优先级（简化实现）
	modelPriority := map[string]int{
		"gpt-4o":              10,
		"gpt-4":               9,
		"claude-3.5-sonnet":   8,
		"claude-3-sonnet":     7,
		"deepseek-chat":       6,
		"gpt-3.5-turbo":       5,
	}
	
	bestPriority := -1
	bestCandidate := req.CandidateResults[0]
	
	for _, candidate := range req.CandidateResults {
		priority := modelPriority[candidate.ModelUsed]
		if priority > bestPriority {
			bestPriority = priority
			bestCandidate = candidate
		}
	}
	
	req.FunctionType = bestCandidate.FunctionType
	req.ComplexityLevel = bestCandidate.ComplexityLevel
	req.AFP = bestCandidate.AFP
	req.UFP = bestCandidate.UFP
	req.ConfidenceScore = bestCandidate.ConfidenceScore
}

// calculateStatistics 计算统计信息
func (r *ResultAggregator) calculateStatistics(result *AggregationResult) {
	result.TotalRequirements = len(result.RequirementResults)
	
	// 统计功能类型
	functionTypeCount := make(map[string]int)
	functionTypeAFP := make(map[string]float64)
	functionTypeUFP := make(map[string]float64)
	functionTypeConfidence := make(map[string]float64)
	
	// 统计复杂度
	complexityCount := make(map[string]int)
	complexityAFP := make(map[string]float64)
	complexityUFP := make(map[string]float64)
	complexityConfidence := make(map[string]float64)
	
	var totalConfidence float64
	
	for _, req := range result.RequirementResults {
		// 功能类型统计
		functionTypeCount[req.FunctionType]++
		functionTypeAFP[req.FunctionType] += req.AFP
		functionTypeUFP[req.FunctionType] += req.UFP
		functionTypeConfidence[req.FunctionType] += req.ConfidenceScore
		
		// 复杂度统计
		complexityCount[req.ComplexityLevel]++
		complexityAFP[req.ComplexityLevel] += req.AFP
		complexityUFP[req.ComplexityLevel] += req.UFP
		complexityConfidence[req.ComplexityLevel] += req.ConfidenceScore
		
		// 总计
		result.AFPTotal += req.AFP
		result.UFPTotal += req.UFP
		totalConfidence += req.ConfidenceScore
	}
	
	// 计算功能类型统计
	for funcType, count := range functionTypeCount {
		stats := FunctionTypeStats{
			Count:         count,
			Percentage:    float64(count) / float64(result.TotalRequirements) * 100,
			TotalAFP:      functionTypeAFP[funcType],
			TotalUFP:      functionTypeUFP[funcType],
			AvgConfidence: functionTypeConfidence[funcType] / float64(count),
		}
		result.FunctionTypeStats[funcType] = stats
	}
	
	// 计算复杂度统计
	for complexity, count := range complexityCount {
		stats := ComplexityStats{
			Count:         count,
			Percentage:    float64(count) / float64(result.TotalRequirements) * 100,
			TotalAFP:      complexityAFP[complexity],
			TotalUFP:      complexityUFP[complexity],
			AvgConfidence: complexityConfidence[complexity] / float64(count),
		}
		result.ComplexityStats[complexity] = stats
	}
	
	// 计算整体质量和一致性评分
	if result.TotalRequirements > 0 {
		result.QualityScore = totalConfidence / float64(result.TotalRequirements)
		result.ConsistencyScore = float64(result.SuccessCount) / float64(result.TotalRequirements)
	}
}

// 辅助方法实现
func (r *ResultAggregator) getProjectInfo(evaluationID string) (uint, uint, error) {
	// 从数据库获取项目信息（简化实现）
	// 这里应该根据evaluationID查询相关的项目和周期信息
	return 1, 1, nil
}

func (r *ResultAggregator) getRequirementInfo(requirementID uint) (*nesma.NesmaRequirement, error) {
	var req nesma.NesmaRequirement
	if err := r.db.First(&req, requirementID).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func (r *ResultAggregator) getModelPerformance(modelName string) float64 {
	// 简化实现，实际应从历史数据计算
	modelPerformance := map[string]float64{
		"gpt-4o":              0.95,
		"gpt-4":               0.9,
		"claude-3.5-sonnet":   0.88,
		"claude-3-sonnet":     0.85,
		"deepseek-chat":       0.8,
		"gpt-3.5-turbo":       0.75,
	}
	
	if performance, exists := modelPerformance[modelName]; exists {
		return performance
	}
	return 0.7
}

func (r *ResultAggregator) calculateTimeFactor(timestamp time.Time) float64 {
	// 基于时间新鲜度计算权重
	hoursSince := time.Since(timestamp).Hours()
	if hoursSince < 1 {
		return 1.0
	} else if hoursSince < 24 {
		return 0.8
	} else {
		return 0.6
	}
}

func (r *ResultAggregator) getHighestVote(votes map[string]float64) string {
	var maxVotes float64
	var winner string
	
	for option, vote := range votes {
		if vote > maxVotes {
			maxVotes = vote
			winner = option
		}
	}
	
	return winner
}

func (r *ResultAggregator) getHighestVoteInt(votes map[string]int) string {
	var maxVotes int
	var winner string
	
	for option, vote := range votes {
		if vote > maxVotes {
			maxVotes = vote
			winner = option
		}
	}
	
	return winner
}

func (r *ResultAggregator) calculateVariance(values []float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	
	// 计算平均值
	var sum float64
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	
	// 计算方差
	var variance float64
	for _, v := range values {
		variance += (v - mean) * (v - mean)
	}
	variance = variance / float64(len(values))
	
	return variance
}

func (r *ResultAggregator) calculateQualityScore(req *AggregatedRequirement) {
	// 简化的质量评分计算
	req.QualityScore = req.ConfidenceScore
	
	// 根据一致性调整
	if req.HasConflict {
		req.QualityScore *= 0.8
	}
	
	// 根据评估数量调整
	if req.EvaluationCount > 1 {
		req.QualityScore *= 1.1
	}
	
	// 确保在0-1范围内
	if req.QualityScore > 1 {
		req.QualityScore = 1
	}
}

func (r *ResultAggregator) generateRequirementSuggestions(req *AggregatedRequirement) {
	if req.HasConflict {
		req.Suggestions = append(req.Suggestions, "存在评估冲突，建议人工复核")
	}
	
	if req.ConfidenceScore < 0.7 {
		req.Suggestions = append(req.Suggestions, "置信度较低，建议提供更详细的需求描述")
	}
	
	if req.EvaluationCount < 2 {
		req.Suggestions = append(req.Suggestions, "评估次数较少，建议增加评估样本")
	}
}

// 质量分析、一致性分析和异常检测的实现
func (r *ResultAggregator) performQualityAnalysis(result *AggregationResult) *QualityAnalysis {
	analysis := &QualityAnalysis{
		QualityDistribution: make(map[string]int),
		QualityTrends:       make([]QualityTrend, 0),
		QualityIssues:       make([]QualityIssue, 0),
		ModelPerformance:    make(map[string]float64),
		ImprovementAreas:    make([]string, 0),
	}
	
	// 简化实现
	analysis.OverallQuality = result.QualityScore
	
	return analysis
}

func (r *ResultAggregator) performConsistencyAnalysis(result *AggregationResult) *ConsistencyAnalysis {
	analysis := &ConsistencyAnalysis{
		ConsistencyByType:   make(map[string]float64),
		ConsistencyByLevel:  make(map[int]float64),
		InconsistentResults: make([]InconsistentResult, 0),
		ConsistencyTrends:   make([]ConsistencyTrend, 0),
		ConsistencyIssues:   make([]ConsistencyIssue, 0),
	}
	
	// 简化实现
	analysis.OverallConsistency = result.ConsistencyScore
	
	return analysis
}

func (r *ResultAggregator) performOutlierDetection(result *AggregationResult) *OutlierDetection {
	detection := &OutlierDetection{
		OutlierResults:   make([]OutlierResult, 0),
		DetectionMethods: []string{"z-score", "iqr"},
		Threshold:        2.0,
		ActionTaken:      "flagged",
	}
	
	// 简化实现
	detection.OutlierCount = 0
	detection.OutlierPercentage = 0.0
	
	return detection
}

func (r *ResultAggregator) generateRecommendations(result *AggregationResult) []Recommendation {
	recommendations := make([]Recommendation, 0)
	
	// 基于质量评分生成建议
	if result.QualityScore < 0.7 {
		recommendations = append(recommendations, Recommendation{
			Category:     "质量改进",
			Priority:     "高",
			Title:        "提升评估质量",
			Description:  "当前评估质量较低，建议改进评估方法",
			ActionItems:  []string{"优化提示词", "增加评估样本", "使用更好的模型"},
			ExpectedImpact: "提升20%评估准确性",
			ImplementationTime: "1-2周",
		})
	}
	
	// 基于一致性评分生成建议
	if result.ConsistencyScore < 0.8 {
		recommendations = append(recommendations, Recommendation{
			Category:     "一致性改进",
			Priority:     "中",
			Title:        "提升评估一致性",
			Description:  "评估结果一致性有待改进",
			ActionItems:  []string{"统一评估标准", "增加训练数据", "使用集成方法"},
			ExpectedImpact: "提升15%一致性",
			ImplementationTime: "2-3周",
		})
	}
	
	return recommendations
}

// GetAggregationResult 获取聚合结果
func (r *ResultAggregator) GetAggregationResult(evaluationID string) (*AggregationResult, error) {
	r.cacheMu.RLock()
	defer r.cacheMu.RUnlock()
	
	if result, exists := r.aggregationCache[evaluationID]; exists {
		return result, nil
	}
	
	return nil, fmt.Errorf("聚合结果不存在: %s", evaluationID)
}

// GetAggregatorMetrics 获取聚合器指标
func (r *ResultAggregator) GetAggregatorMetrics() *AggregatorMetrics {
	r.metrics.mu.RLock()
	defer r.metrics.mu.RUnlock()
	
	// 返回副本
	metrics := *r.metrics
	return &metrics
}

// UpdateConfig 更新配置
func (r *ResultAggregator) UpdateConfig(config *AggregatorConfig) {
	r.config = config
	global.GVA_LOG.Info("聚合器配置已更新", zap.Any("config", config))
}

// ClearCache 清理缓存
func (r *ResultAggregator) ClearCache() {
	r.cacheMu.Lock()
	defer r.cacheMu.Unlock()
	
	r.aggregationCache = make(map[string]*AggregationResult)
	global.GVA_LOG.Info("聚合器缓存已清理")
}