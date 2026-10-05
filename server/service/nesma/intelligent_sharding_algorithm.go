package nesma

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// IntelligentShardingAlgorithm 智能分片算法
type IntelligentShardingAlgorithm struct {
	tokenCalculator *TokenCalculator
	contextManager  *ContextManager
	knowledgeService *KnowledgeService
}

// NewIntelligentShardingAlgorithm 创建智能分片算法实例
func NewIntelligentShardingAlgorithm(tokenCalc *TokenCalculator, contextMgr *ContextManager, knowledgeSvc *KnowledgeService) *IntelligentShardingAlgorithm {
	return &IntelligentShardingAlgorithm{
		tokenCalculator: tokenCalc,
		contextManager:  contextMgr,
		knowledgeService: knowledgeSvc,
	}
}

// ShardingStrategy 分片策略枚举
type ShardingStrategy string

const (
	StrategyOptimal     ShardingStrategy = "optimal"     // 最优分片：综合考虑Token、相似度、复杂度
	StrategyBalanced    ShardingStrategy = "balanced"    // 平衡分片：均匀分布Token和任务数
	StrategyToken       ShardingStrategy = "token"       // Token优先：按Token容量分片
	StrategyComplexity  ShardingStrategy = "complexity"  // 复杂度优先：按复杂度分片
	StrategySimilarity  ShardingStrategy = "similarity"  // 相似度优先：按语义相似度分片
	StrategySequential  ShardingStrategy = "sequential"  // 顺序分片：简单顺序分割
)

// ShardingConfig 分片配置
type ShardingConfig struct {
	Strategy            ShardingStrategy `json:"strategy"`
	MaxTokensPerShard   int              `json:"max_tokens_per_shard"`
	MinTokensPerShard   int              `json:"min_tokens_per_shard"`
	MaxRequirements     int              `json:"max_requirements"`
	MinRequirements     int              `json:"min_requirements"`
	SimilarityThreshold float64          `json:"similarity_threshold"`
	ComplexityBalance   bool             `json:"complexity_balance"`
	ContextAwareness    bool             `json:"context_awareness"`
	TokenEfficiency     float64          `json:"token_efficiency"`    // 0.0-1.0
	QualityPriority     float64          `json:"quality_priority"`    // 0.0-1.0
	PerformancePriority float64          `json:"performance_priority"` // 0.0-1.0
}

// RequirementShard 需求分片
type RequirementShard struct {
	ShardID           string                  `json:"shard_id"`
	ShardNumber       int                     `json:"shard_number"`
	Requirements      []nesma.NesmaRequirement `json:"requirements"`
	EstimatedTokens   int                     `json:"estimated_tokens"`
	ActualTokens      int                     `json:"actual_tokens"`
	AverageComplexity float64                 `json:"average_complexity"`
	SimilarityScore   float64                 `json:"similarity_score"`
	ContextScore      float64                 `json:"context_score"`
	QualityScore      float64                 `json:"quality_score"`
	Priority          int                     `json:"priority"`
	RecommendedModel  string                  `json:"recommended_model"`
	ProcessingHints   []string                `json:"processing_hints"`
	Dependencies      []string                `json:"dependencies"`
}

// ShardingResult 分片结果
type ShardingResult struct {
	Strategy          ShardingStrategy      `json:"strategy"`
	Shards            []RequirementShard    `json:"shards"`
	TotalShards       int                   `json:"total_shards"`
	TotalRequirements int                   `json:"total_requirements"`
	TotalTokens       int                   `json:"total_tokens"`
	AverageTokens     float64               `json:"average_tokens"`
	TokenUtilization  float64               `json:"token_utilization"`
	BalanceScore      float64               `json:"balance_score"`
	QualityScore      float64               `json:"quality_score"`
	OptimizationHints []string              `json:"optimization_hints"`
	ProcessingTime    time.Duration         `json:"processing_time"`
}

// ExecuteSharding 执行智能分片
func (a *IntelligentShardingAlgorithm) ExecuteSharding(requirements []nesma.NesmaRequirement, config *ShardingConfig) (*ShardingResult, error) {
	startTime := time.Now()
	
	global.GVA_LOG.Info("开始智能分片算法", 
		zap.String("strategy", string(config.Strategy)),
		zap.Int("requirements_count", len(requirements)),
		zap.Int("max_tokens_per_shard", config.MaxTokensPerShard))

	// 1. 预处理：分析需求特征
	analyzedReqs, err := a.analyzeRequirements(requirements)
	if err != nil {
		return nil, fmt.Errorf("需求分析失败: %w", err)
	}

	// 2. 根据策略执行分片
	var shards []RequirementShard
	switch config.Strategy {
	case StrategyOptimal:
		shards, err = a.optimalSharding(analyzedReqs, config)
	case StrategyBalanced:
		shards, err = a.balancedSharding(analyzedReqs, config)
	case StrategyToken:
		shards, err = a.tokenBasedSharding(analyzedReqs, config)
	case StrategyComplexity:
		shards, err = a.complexityBasedSharding(analyzedReqs, config)
	case StrategySimilarity:
		shards, err = a.similarityBasedSharding(analyzedReqs, config)
	case StrategySequential:
		shards, err = a.sequentialSharding(analyzedReqs, config)
	default:
		return nil, fmt.Errorf("不支持的分片策略: %s", config.Strategy)
	}

	if err != nil {
		return nil, fmt.Errorf("分片执行失败: %w", err)
	}

	// 3. 后处理：优化和验证
	shards, err = a.postProcessShards(shards, config)
	if err != nil {
		return nil, fmt.Errorf("分片后处理失败: %w", err)
	}

	// 4. 计算分片质量指标
	result := &ShardingResult{
		Strategy:          config.Strategy,
		Shards:            shards,
		TotalShards:       len(shards),
		TotalRequirements: len(requirements),
		TotalTokens:       a.calculateTotalTokens(shards),
		ProcessingTime:    time.Since(startTime),
	}

	a.calculateQualityMetrics(result)
	a.generateOptimizationHints(result, config)

	global.GVA_LOG.Info("智能分片算法完成", 
		zap.Int("total_shards", result.TotalShards),
		zap.Int("total_tokens", result.TotalTokens),
		zap.Float64("balance_score", result.BalanceScore),
		zap.Float64("quality_score", result.QualityScore),
		zap.Duration("processing_time", result.ProcessingTime))

	return result, nil
}

// analyzeRequirements 分析需求特征
func (a *IntelligentShardingAlgorithm) analyzeRequirements(requirements []nesma.NesmaRequirement) ([]AnalyzedRequirement, error) {
	analyzed := make([]AnalyzedRequirement, len(requirements))
	
	for i, req := range requirements {
		// 计算Token数量
		tokens := a.tokenCalculator.EstimateTokens(req.Title + " " + req.Description, "deepseek-chat")
		
		// 计算复杂度评分
		complexity := a.calculateComplexityScore(req)
		
		// 计算语义向量（简化版）
		semanticVector := a.calculateSemanticVector(req)
		
		analyzed[i] = AnalyzedRequirement{
			Requirement:     req,
			TokenCount:      tokens,
			ComplexityScore: complexity,
			SemanticVector:  semanticVector,
			Priority:        a.calculatePriority(req),
			ProcessingHints: a.generateProcessingHints(req),
		}
	}
	
	return analyzed, nil
}

// optimalSharding 最优分片算法
func (a *IntelligentShardingAlgorithm) optimalSharding(requirements []AnalyzedRequirement, config *ShardingConfig) ([]RequirementShard, error) {
	// 1. 聚类分析：按语义相似度分组
	clusters := a.performClustering(requirements, config)
	
	// 2. 负载均衡：调整分片大小
	balancedShards := a.balanceShardLoad(clusters, config)
	
	// 3. 质量优化：微调分片组合
	optimizedShards := a.optimizeShardQuality(balancedShards, config)
	
	return optimizedShards, nil
}

// balancedSharding 平衡分片算法
func (a *IntelligentShardingAlgorithm) balancedSharding(requirements []AnalyzedRequirement, config *ShardingConfig) ([]RequirementShard, error) {
	totalTokens := 0
	for _, req := range requirements {
		totalTokens += req.TokenCount
	}
	
	// 计算理想分片数
	idealShards := int(math.Ceil(float64(totalTokens) / float64(config.MaxTokensPerShard)))
	avgTokensPerShard := totalTokens / idealShards
	
	var shards []RequirementShard
	currentShard := RequirementShard{
		ShardID:     fmt.Sprintf("shard_%d", len(shards)+1),
		ShardNumber: len(shards) + 1,
	}
	
	currentTokens := 0
	
	for _, req := range requirements {
		// 如果添加当前需求会超过平均Token数，并且当前分片不为空
		if currentTokens+req.TokenCount > avgTokensPerShard && len(currentShard.Requirements) > 0 {
			// 完成当前分片
			currentShard.EstimatedTokens = currentTokens
			currentShard.AverageComplexity = a.calculateAverageComplexity(currentShard.Requirements)
			shards = append(shards, currentShard)
			
			// 开始新分片
			currentShard = RequirementShard{
				ShardID:     fmt.Sprintf("shard_%d", len(shards)+1),
				ShardNumber: len(shards) + 1,
			}
			currentTokens = 0
		}
		
		currentShard.Requirements = append(currentShard.Requirements, req.Requirement)
		currentTokens += req.TokenCount
	}
	
	// 处理最后一个分片
	if len(currentShard.Requirements) > 0 {
		currentShard.EstimatedTokens = currentTokens
		currentShard.AverageComplexity = a.calculateAverageComplexity(currentShard.Requirements)
		shards = append(shards, currentShard)
	}
	
	return shards, nil
}

// tokenBasedSharding 基于Token的分片算法
func (a *IntelligentShardingAlgorithm) tokenBasedSharding(requirements []AnalyzedRequirement, config *ShardingConfig) ([]RequirementShard, error) {
	var shards []RequirementShard
	currentShard := RequirementShard{
		ShardID:     fmt.Sprintf("shard_%d", len(shards)+1),
		ShardNumber: len(shards) + 1,
	}
	
	currentTokens := 0
	
	for _, req := range requirements {
		// 检查是否需要创建新分片
		if currentTokens+req.TokenCount > config.MaxTokensPerShard && len(currentShard.Requirements) > 0 {
			// 完成当前分片
			currentShard.EstimatedTokens = currentTokens
			currentShard.AverageComplexity = a.calculateAverageComplexity(currentShard.Requirements)
			shards = append(shards, currentShard)
			
			// 开始新分片
			currentShard = RequirementShard{
				ShardID:     fmt.Sprintf("shard_%d", len(shards)+1),
				ShardNumber: len(shards) + 1,
			}
			currentTokens = 0
		}
		
		currentShard.Requirements = append(currentShard.Requirements, req.Requirement)
		currentTokens += req.TokenCount
	}
	
	// 处理最后一个分片
	if len(currentShard.Requirements) > 0 {
		currentShard.EstimatedTokens = currentTokens
		currentShard.AverageComplexity = a.calculateAverageComplexity(currentShard.Requirements)
		shards = append(shards, currentShard)
	}
	
	return shards, nil
}

// complexityBasedSharding 基于复杂度的分片算法
func (a *IntelligentShardingAlgorithm) complexityBasedSharding(requirements []AnalyzedRequirement, config *ShardingConfig) ([]RequirementShard, error) {
	// 按复杂度排序
	sort.Slice(requirements, func(i, j int) bool {
		return requirements[i].ComplexityScore > requirements[j].ComplexityScore
	})
	
	var shards []RequirementShard
	
	// 使用轮询方式分配，确保每个分片的复杂度相对均衡
	shardCount := int(math.Ceil(float64(len(requirements)) / float64(config.MaxRequirements)))
	if shardCount == 0 {
		shardCount = 1
	}
	
	// 初始化分片
	for i := 0; i < shardCount; i++ {
		shards = append(shards, RequirementShard{
			ShardID:     fmt.Sprintf("shard_%d", i+1),
			ShardNumber: i + 1,
		})
	}
	
	// 轮询分配需求
	for i, req := range requirements {
		shardIndex := i % shardCount
		shards[shardIndex].Requirements = append(shards[shardIndex].Requirements, req.Requirement)
		shards[shardIndex].EstimatedTokens += req.TokenCount
	}
	
	// 计算每个分片的平均复杂度
	for i := range shards {
		shards[i].AverageComplexity = a.calculateAverageComplexity(shards[i].Requirements)
	}
	
	return shards, nil
}

// similarityBasedSharding 基于相似度的分片算法
func (a *IntelligentShardingAlgorithm) similarityBasedSharding(requirements []AnalyzedRequirement, config *ShardingConfig) ([]RequirementShard, error) {
	// 执行层次聚类
	clusters := a.performHierarchicalClustering(requirements, config.SimilarityThreshold)
	
	var shards []RequirementShard
	
	for i, cluster := range clusters {
		shard := RequirementShard{
			ShardID:     fmt.Sprintf("shard_%d", i+1),
			ShardNumber: i + 1,
		}
		
		totalTokens := 0
		for _, req := range cluster {
			shard.Requirements = append(shard.Requirements, req.Requirement)
			totalTokens += req.TokenCount
		}
		
		shard.EstimatedTokens = totalTokens
		shard.AverageComplexity = a.calculateAverageComplexity(shard.Requirements)
		shard.SimilarityScore = a.calculateClusterSimilarity(cluster)
		
		shards = append(shards, shard)
	}
	
	return shards, nil
}

// sequentialSharding 顺序分片算法
func (a *IntelligentShardingAlgorithm) sequentialSharding(requirements []AnalyzedRequirement, config *ShardingConfig) ([]RequirementShard, error) {
	var shards []RequirementShard
	
	shardSize := config.MaxRequirements
	if shardSize == 0 {
		shardSize = 10 // 默认大小
	}
	
	for i := 0; i < len(requirements); i += shardSize {
		end := i + shardSize
		if end > len(requirements) {
			end = len(requirements)
		}
		
		shard := RequirementShard{
			ShardID:     fmt.Sprintf("shard_%d", len(shards)+1),
			ShardNumber: len(shards) + 1,
		}
		
		totalTokens := 0
		for j := i; j < end; j++ {
			shard.Requirements = append(shard.Requirements, requirements[j].Requirement)
			totalTokens += requirements[j].TokenCount
		}
		
		shard.EstimatedTokens = totalTokens
		shard.AverageComplexity = a.calculateAverageComplexity(shard.Requirements)
		
		shards = append(shards, shard)
	}
	
	return shards, nil
}

// AnalyzedRequirement 分析后的需求
type AnalyzedRequirement struct {
	Requirement     nesma.NesmaRequirement `json:"requirement"`
	TokenCount      int                    `json:"token_count"`
	ComplexityScore float64                `json:"complexity_score"`
	SemanticVector  []float64              `json:"semantic_vector"`
	Priority        int                    `json:"priority"`
	ProcessingHints []string               `json:"processing_hints"`
}

// 辅助方法实现
func (a *IntelligentShardingAlgorithm) calculateComplexityScore(req nesma.NesmaRequirement) float64 {
	score := 0.0
	
	// 基于描述长度
	descLength := len(req.Description)
	if descLength > 500 {
		score += 0.3
	} else if descLength > 200 {
		score += 0.2
	} else {
		score += 0.1
	}
	
	// 基于层级
	switch req.Level {
	case 1:
		score += 0.1
	case 2:
		score += 0.2
	case 3:
		score += 0.4
	case 4:
		score += 0.6
	}
	
	// 基于关键词
	keywords := []string{"算法", "复杂", "集成", "接口", "数据库", "安全", "性能", "并发"}
	desc := strings.ToLower(req.Description)
	for _, keyword := range keywords {
		if strings.Contains(desc, keyword) {
			score += 0.1
		}
	}
	
	return math.Min(score, 1.0)
}

func (a *IntelligentShardingAlgorithm) calculateSemanticVector(req nesma.NesmaRequirement) []float64 {
	// 简化的语义向量计算
	text := req.Title + " " + req.Description
	words := strings.Fields(strings.ToLower(text))
	
	// 使用简单的词频向量
	wordFreq := make(map[string]int)
	for _, word := range words {
		wordFreq[word]++
	}
	
	// 转换为固定长度向量
	vector := make([]float64, 50)
	i := 0
	for _, freq := range wordFreq {
		if i >= 50 {
			break
		}
		vector[i] = float64(freq)
		i++
	}
	
	return vector
}

func (a *IntelligentShardingAlgorithm) calculatePriority(req nesma.NesmaRequirement) int {
	// 根据需求特征计算优先级
	priority := 50 // 默认优先级
	
	// 基于层级调整
	priority += (req.Level - 1) * 10
	
	// 基于关键词调整
	highPriorityKeywords := []string{"核心", "关键", "重要", "安全", "性能"}
	desc := strings.ToLower(req.Description)
	for _, keyword := range highPriorityKeywords {
		if strings.Contains(desc, keyword) {
			priority += 20
		}
	}
	
	return priority
}

func (a *IntelligentShardingAlgorithm) generateProcessingHints(req nesma.NesmaRequirement) []string {
	var hints []string
	
	desc := strings.ToLower(req.Description)
	
	if strings.Contains(desc, "数据库") {
		hints = append(hints, "数据库相关功能")
	}
	if strings.Contains(desc, "接口") {
		hints = append(hints, "API接口功能")
	}
	if strings.Contains(desc, "用户") {
		hints = append(hints, "用户交互功能")
	}
	if strings.Contains(desc, "报表") {
		hints = append(hints, "报表生成功能")
	}
	if strings.Contains(desc, "计算") {
		hints = append(hints, "计算处理功能")
	}
	
	return hints
}

func (a *IntelligentShardingAlgorithm) calculateAverageComplexity(requirements []nesma.NesmaRequirement) float64 {
	if len(requirements) == 0 {
		return 0.0
	}
	
	total := 0.0
	for _, req := range requirements {
		total += a.calculateComplexityScore(req)
	}
	
	return total / float64(len(requirements))
}

func (a *IntelligentShardingAlgorithm) performClustering(requirements []AnalyzedRequirement, config *ShardingConfig) [][]AnalyzedRequirement {
	// 简化的聚类实现
	var clusters [][]AnalyzedRequirement
	
	// 按复杂度分组
	lowComplexity := make([]AnalyzedRequirement, 0)
	mediumComplexity := make([]AnalyzedRequirement, 0)
	highComplexity := make([]AnalyzedRequirement, 0)
	
	for _, req := range requirements {
		if req.ComplexityScore < 0.3 {
			lowComplexity = append(lowComplexity, req)
		} else if req.ComplexityScore < 0.7 {
			mediumComplexity = append(mediumComplexity, req)
		} else {
			highComplexity = append(highComplexity, req)
		}
	}
	
	if len(lowComplexity) > 0 {
		clusters = append(clusters, lowComplexity)
	}
	if len(mediumComplexity) > 0 {
		clusters = append(clusters, mediumComplexity)
	}
	if len(highComplexity) > 0 {
		clusters = append(clusters, highComplexity)
	}
	
	return clusters
}

func (a *IntelligentShardingAlgorithm) balanceShardLoad(clusters [][]AnalyzedRequirement, config *ShardingConfig) []RequirementShard {
	var shards []RequirementShard
	
	for i, cluster := range clusters {
		// 如果簇太大，进一步分割
		if len(cluster) > config.MaxRequirements {
			subShards := a.splitLargeCluster(cluster, config)
			shards = append(shards, subShards...)
		} else {
			shard := RequirementShard{
				ShardID:     fmt.Sprintf("shard_%d", i+1),
				ShardNumber: i + 1,
			}
			
			totalTokens := 0
			for _, req := range cluster {
				shard.Requirements = append(shard.Requirements, req.Requirement)
				totalTokens += req.TokenCount
			}
			
			shard.EstimatedTokens = totalTokens
			shard.AverageComplexity = a.calculateAverageComplexity(shard.Requirements)
			shards = append(shards, shard)
		}
	}
	
	return shards
}

func (a *IntelligentShardingAlgorithm) splitLargeCluster(cluster []AnalyzedRequirement, config *ShardingConfig) []RequirementShard {
	var shards []RequirementShard
	
	subShardSize := config.MaxRequirements
	for i := 0; i < len(cluster); i += subShardSize {
		end := i + subShardSize
		if end > len(cluster) {
			end = len(cluster)
		}
		
		shard := RequirementShard{
			ShardID:     fmt.Sprintf("shard_%d", len(shards)+1),
			ShardNumber: len(shards) + 1,
		}
		
		totalTokens := 0
		for j := i; j < end; j++ {
			shard.Requirements = append(shard.Requirements, cluster[j].Requirement)
			totalTokens += cluster[j].TokenCount
		}
		
		shard.EstimatedTokens = totalTokens
		shard.AverageComplexity = a.calculateAverageComplexity(shard.Requirements)
		shards = append(shards, shard)
	}
	
	return shards
}

func (a *IntelligentShardingAlgorithm) optimizeShardQuality(shards []RequirementShard, config *ShardingConfig) []RequirementShard {
	// 简化的质量优化
	for i := range shards {
		shards[i].QualityScore = a.calculateShardQualityScore(shards[i])
		shards[i].RecommendedModel = a.selectOptimalModel(shards[i])
		shards[i].ProcessingHints = a.generateShardProcessingHints(shards[i])
	}
	
	return shards
}

func (a *IntelligentShardingAlgorithm) calculateShardQualityScore(shard RequirementShard) float64 {
	score := 0.0
	
	// Token利用率
	tokenUtilization := float64(shard.EstimatedTokens) / 4000.0 // 假设4000是理想Token数
	if tokenUtilization > 0.8 && tokenUtilization < 1.0 {
		score += 0.3
	} else if tokenUtilization > 0.6 {
		score += 0.2
	} else {
		score += 0.1
	}
	
	// 复杂度均衡
	if shard.AverageComplexity > 0.3 && shard.AverageComplexity < 0.7 {
		score += 0.3
	} else {
		score += 0.1
	}
	
	// 需求数量
	reqCount := len(shard.Requirements)
	if reqCount >= 5 && reqCount <= 15 {
		score += 0.2
	} else {
		score += 0.1
	}
	
	// 相似度
	score += shard.SimilarityScore * 0.2
	
	return math.Min(score, 1.0)
}

func (a *IntelligentShardingAlgorithm) selectOptimalModel(shard RequirementShard) string {
	if shard.AverageComplexity > 0.7 {
		return "gpt-4"
	} else if shard.AverageComplexity > 0.4 {
		return "gpt-3.5-turbo"
	} else {
		return "deepseek-chat"
	}
}

func (a *IntelligentShardingAlgorithm) generateShardProcessingHints(shard RequirementShard) []string {
	var hints []string
	
	if shard.AverageComplexity > 0.7 {
		hints = append(hints, "高复杂度分片，建议使用高性能模型")
	}
	
	if shard.EstimatedTokens > 3000 {
		hints = append(hints, "Token数量较多，可能需要分割")
	}
	
	if len(shard.Requirements) > 20 {
		hints = append(hints, "需求数量较多，建议并行处理")
	}
	
	return hints
}

func (a *IntelligentShardingAlgorithm) performHierarchicalClustering(requirements []AnalyzedRequirement, threshold float64) [][]AnalyzedRequirement {
	// 简化的层次聚类
	var clusters [][]AnalyzedRequirement
	
	// 按相似度分组（简化实现）
	processed := make([]bool, len(requirements))
	
	for i := 0; i < len(requirements); i++ {
		if processed[i] {
			continue
		}
		
		cluster := []AnalyzedRequirement{requirements[i]}
		processed[i] = true
		
		for j := i + 1; j < len(requirements); j++ {
			if processed[j] {
				continue
			}
			
			similarity := a.calculateSimilarity(requirements[i], requirements[j])
			if similarity >= threshold {
				cluster = append(cluster, requirements[j])
				processed[j] = true
			}
		}
		
		clusters = append(clusters, cluster)
	}
	
	return clusters
}

func (a *IntelligentShardingAlgorithm) calculateSimilarity(req1, req2 AnalyzedRequirement) float64 {
	// 简化的相似度计算
	if req1.Requirement.Level == req2.Requirement.Level {
		return 0.8
	}
	
	// 基于描述关键词
	desc1 := strings.ToLower(req1.Requirement.Description)
	desc2 := strings.ToLower(req2.Requirement.Description)
	
	words1 := strings.Fields(desc1)
	words2 := strings.Fields(desc2)
	
	common := 0
	for _, word1 := range words1 {
		for _, word2 := range words2 {
			if word1 == word2 {
				common++
			}
		}
	}
	
	return float64(common) / float64(len(words1)+len(words2)-common)
}

func (a *IntelligentShardingAlgorithm) calculateClusterSimilarity(cluster []AnalyzedRequirement) float64 {
	if len(cluster) <= 1 {
		return 1.0
	}
	
	totalSimilarity := 0.0
	comparisons := 0
	
	for i := 0; i < len(cluster); i++ {
		for j := i + 1; j < len(cluster); j++ {
			totalSimilarity += a.calculateSimilarity(cluster[i], cluster[j])
			comparisons++
		}
	}
	
	if comparisons == 0 {
		return 1.0
	}
	
	return totalSimilarity / float64(comparisons)
}

func (a *IntelligentShardingAlgorithm) postProcessShards(shards []RequirementShard, config *ShardingConfig) ([]RequirementShard, error) {
	// 后处理：合并小分片、分割大分片
	var processedShards []RequirementShard
	
	for _, shard := range shards {
		if shard.EstimatedTokens > config.MaxTokensPerShard {
			// 分割大分片
			splitShards := a.splitLargeShard(shard, config)
			processedShards = append(processedShards, splitShards...)
		} else {
			processedShards = append(processedShards, shard)
		}
	}
	
	// 合并小分片
	processedShards = a.mergeSmallShards(processedShards, config)
	
	return processedShards, nil
}

func (a *IntelligentShardingAlgorithm) splitLargeShard(shard RequirementShard, config *ShardingConfig) []RequirementShard {
	var newShards []RequirementShard
	
	targetSize := config.MaxTokensPerShard
	currentShard := RequirementShard{
		ShardID:     fmt.Sprintf("%s_1", shard.ShardID),
		ShardNumber: shard.ShardNumber,
	}
	
	currentTokens := 0
	shardCounter := 1
	
	for _, req := range shard.Requirements {
		reqTokens := a.tokenCalculator.EstimateTokens(req.Title + " " + req.Description, "deepseek-chat")
		
		if currentTokens+reqTokens > targetSize && len(currentShard.Requirements) > 0 {
			currentShard.EstimatedTokens = currentTokens
			currentShard.AverageComplexity = a.calculateAverageComplexity(currentShard.Requirements)
			newShards = append(newShards, currentShard)
			
			shardCounter++
			currentShard = RequirementShard{
				ShardID:     fmt.Sprintf("%s_%d", shard.ShardID, shardCounter),
				ShardNumber: shard.ShardNumber,
			}
			currentTokens = 0
		}
		
		currentShard.Requirements = append(currentShard.Requirements, req)
		currentTokens += reqTokens
	}
	
	if len(currentShard.Requirements) > 0 {
		currentShard.EstimatedTokens = currentTokens
		currentShard.AverageComplexity = a.calculateAverageComplexity(currentShard.Requirements)
		newShards = append(newShards, currentShard)
	}
	
	return newShards
}

func (a *IntelligentShardingAlgorithm) mergeSmallShards(shards []RequirementShard, config *ShardingConfig) []RequirementShard {
	var mergedShards []RequirementShard
	
	i := 0
	for i < len(shards) {
		currentShard := shards[i]
		
		// 如果分片太小，尝试与下一个分片合并
		if currentShard.EstimatedTokens < config.MinTokensPerShard && i+1 < len(shards) {
			nextShard := shards[i+1]
			
			// 检查合并后是否超过最大限制
			if currentShard.EstimatedTokens+nextShard.EstimatedTokens <= config.MaxTokensPerShard {
				// 合并分片
				mergedShard := RequirementShard{
					ShardID:     fmt.Sprintf("merged_%d", len(mergedShards)+1),
					ShardNumber: len(mergedShards) + 1,
				}
				
				mergedShard.Requirements = append(mergedShard.Requirements, currentShard.Requirements...)
				mergedShard.Requirements = append(mergedShard.Requirements, nextShard.Requirements...)
				mergedShard.EstimatedTokens = currentShard.EstimatedTokens + nextShard.EstimatedTokens
				mergedShard.AverageComplexity = a.calculateAverageComplexity(mergedShard.Requirements)
				
				mergedShards = append(mergedShards, mergedShard)
				i += 2 // 跳过已合并的两个分片
			} else {
				mergedShards = append(mergedShards, currentShard)
				i++
			}
		} else {
			mergedShards = append(mergedShards, currentShard)
			i++
		}
	}
	
	return mergedShards
}

func (a *IntelligentShardingAlgorithm) calculateTotalTokens(shards []RequirementShard) int {
	total := 0
	for _, shard := range shards {
		total += shard.EstimatedTokens
	}
	return total
}

func (a *IntelligentShardingAlgorithm) calculateQualityMetrics(result *ShardingResult) {
	if len(result.Shards) == 0 {
		return
	}
	
	// 计算平均Token数
	result.AverageTokens = float64(result.TotalTokens) / float64(len(result.Shards))
	
	// 计算Token利用率
	maxPossibleTokens := len(result.Shards) * 4000 // 假设4000是理想Token数
	result.TokenUtilization = float64(result.TotalTokens) / float64(maxPossibleTokens)
	
	// 计算平衡分数
	result.BalanceScore = a.calculateBalanceScore(result.Shards)
	
	// 计算质量分数
	totalQuality := 0.0
	for _, shard := range result.Shards {
		totalQuality += shard.QualityScore
	}
	result.QualityScore = totalQuality / float64(len(result.Shards))
}

func (a *IntelligentShardingAlgorithm) calculateBalanceScore(shards []RequirementShard) float64 {
	if len(shards) == 0 {
		return 0.0
	}
	
	// 计算Token数量的标准差
	tokens := make([]int, len(shards))
	sum := 0
	for i, shard := range shards {
		tokens[i] = shard.EstimatedTokens
		sum += shard.EstimatedTokens
	}
	
	mean := float64(sum) / float64(len(shards))
	variance := 0.0
	
	for _, token := range tokens {
		variance += math.Pow(float64(token)-mean, 2)
	}
	
	variance /= float64(len(shards))
	stdDev := math.Sqrt(variance)
	
	// 标准差越小，平衡分数越高
	balanceScore := 1.0 - (stdDev / mean)
	return math.Max(0.0, balanceScore)
}

func (a *IntelligentShardingAlgorithm) generateOptimizationHints(result *ShardingResult, config *ShardingConfig) {
	var hints []string
	
	if result.BalanceScore < 0.7 {
		hints = append(hints, "分片不够均衡，建议使用平衡分片策略")
	}
	
	if result.TokenUtilization < 0.7 {
		hints = append(hints, "Token利用率偏低，可以增加每个分片的Token数量")
	}
	
	if result.QualityScore < 0.6 {
		hints = append(hints, "分片质量偏低，建议使用最优分片策略")
	}
	
	if len(result.Shards) > 50 {
		hints = append(hints, "分片数量过多，建议增加分片大小")
	}
	
	if len(result.Shards) < 5 {
		hints = append(hints, "分片数量过少，建议减少分片大小以提高并行度")
	}
	
	result.OptimizationHints = hints
}