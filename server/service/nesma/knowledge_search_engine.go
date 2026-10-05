package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"go.uber.org/zap"
)

// KnowledgeSearchEngine 知识库智能搜索引擎
type KnowledgeSearchEngine struct {
	embedder         *AIEmbeddingService
	vectorStore      *EnhancedVectorStore  
	semanticReasoner *SemanticReasoningEngine
	queryProcessor   *QueryProcessor
	resultRanker     *SemanticResultRanker
	cache            *SearchCache
	monitor          *SearchMonitor
	mu               sync.RWMutex
}

// AIEmbeddingService AI嵌入服务
type AIEmbeddingService struct {
	deepseeekClient *DeepSeekClient
	cache          *EmbeddingCache
	batchSize      int
}

// EnhancedVectorStore 增强向量存储
type EnhancedVectorStore struct {
	vectors    map[string]*KnowledgeVector
	index      *SemanticIndex
	metadata   *VectorMetadata
	mu         sync.RWMutex
}

// SemanticReasoningEngine 语义推理引擎
type SemanticReasoningEngine struct {
	ruleBase     *NESMARuleBase
	ontology     *NESMAOntology
	inferEngine  *LogicalInferenceEngine
	explainer    *ReasoningExplainer
}

// QueryProcessor 查询处理器
type QueryProcessor struct {
	analyzer    *SemanticAnalyzer
	expander    *QueryExpander
	optimizer   *QueryOptimizer
	normalizer  *QueryNormalizer
}

// SemanticResultRanker 语义结果排序器
type SemanticResultRanker struct {
	weights    map[string]float64
	boosters   []RelevanceBooster
	diversifier *ResultDiversifier
}

// KnowledgeVector 知识向量
type KnowledgeVector struct {
	ID          string                 `json:"id"`
	Content     string                 `json:"content"`
	Embedding   []float64              `json:"embedding"`
	Type        string                 `json:"type"`          // requirement, standard, case_study, rule
	Category    string                 `json:"category"`      // NESMA_STANDARD, BEST_PRACTICE, etc.
	Domain      string                 `json:"domain"`        // 业务领域
	Concepts    []string               `json:"concepts"`      // 包含的概念
	Relations   map[string]interface{} `json:"relations"`     // 关系信息
	Quality     float64                `json:"quality"`       // 质量分数
	Freshness   time.Time              `json:"freshness"`     // 新鲜度
	Usage       int                    `json:"usage"`         // 使用次数
	Metadata    map[string]interface{} `json:"metadata"`
}

// KnowledgeSearchRequest 知识搜索请求
type KnowledgeSearchRequest struct {
	Query         string                 `json:"query"`
	Type          string                 `json:"type"`           // semantic, vector, hybrid, reasoning
	Domain        string                 `json:"domain"`         // 限制搜索域
	Category      []string               `json:"category"`       // 限制类别
	Concepts      []string               `json:"concepts"`       // 相关概念
	MinRelevance  float64                `json:"min_relevance"`  // 最小相关度
	MaxResults    int                    `json:"max_results"`
	IncludeRules  bool                   `json:"include_rules"`  // 是否包含推理规则
	UseCache      bool                   `json:"use_cache"`
	Context       map[string]interface{} `json:"context"`        // 上下文信息
}

// KnowledgeSearchResult 知识搜索结果
type KnowledgeSearchResult struct {
	ID             string                 `json:"id"`
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Type           string                 `json:"type"`
	Category       string                 `json:"category"`
	Domain         string                 `json:"domain"`
	RelevanceScore float64                `json:"relevance_score"`
	SemanticScore  float64                `json:"semantic_score"`
	QualityScore   float64                `json:"quality_score"`
	FinalScore     float64                `json:"final_score"`
	Explanation    string                 `json:"explanation"`    // 相关性解释
	Highlights     []string               `json:"highlights"`     // 高亮片段
	RelatedConcepts []string              `json:"related_concepts"`
	Usage          int                    `json:"usage"`
	Source         string                 `json:"source"`
	Author         string                 `json:"author"`
	CreatedAt      time.Time              `json:"created_at"`
	Metadata       map[string]interface{} `json:"metadata"`
}

// KnowledgeSearchResponse 知识搜索响应
type KnowledgeSearchResponse struct {
	Results       []*KnowledgeSearchResult `json:"results"`
	Total         int                      `json:"total"`
	QueryTime     time.Duration            `json:"query_time"`
	Method        string                   `json:"method"`
	ProcessedQuery string                  `json:"processed_query"`
	Concepts      []string                 `json:"extracted_concepts"`
	Suggestions   []string                 `json:"suggestions"`
	RelatedQueries []string                `json:"related_queries"`
	Quality       *SearchQualityMetrics    `json:"quality"`
	Metadata      map[string]interface{}   `json:"metadata"`
}

// SearchQualityMetrics 搜索质量指标
type SearchQualityMetrics struct {
	Precision      float64 `json:"precision"`
	Coverage       float64 `json:"coverage"`
	Diversity      float64 `json:"diversity"`
	SemanticDepth  float64 `json:"semantic_depth"`
	Freshness      float64 `json:"freshness"`
	Confidence     float64 `json:"confidence"`
}

// NewKnowledgeSearchEngine 创建知识搜索引擎
func NewKnowledgeSearchEngine() *KnowledgeSearchEngine {
	return &KnowledgeSearchEngine{
		embedder:         NewAIEmbeddingService(),
		vectorStore:      NewEnhancedVectorStore(),
		semanticReasoner: NewSemanticReasoningEngine(),
		queryProcessor:   NewQueryProcessor(),
		resultRanker:     NewSemanticResultRanker(),
		cache:            NewSearchCache(),
		monitor:          NewSearchMonitor(),
	}
}

// Search 执行智能知识搜索
func (kse *KnowledgeSearchEngine) Search(ctx context.Context, req *KnowledgeSearchRequest) (*KnowledgeSearchResponse, error) {
	startTime := time.Now()
	
	// 记录搜索请求
	kse.monitor.RecordRequest(req)
	
	// 生成缓存键
	cacheKey := kse.generateCacheKey(req)
	
	// 检查缓存
	if req.UseCache {
		if cached := kse.cache.Get(cacheKey); cached != nil {
			kse.monitor.RecordCacheHit()
			return cached.(*KnowledgeSearchResponse), nil
		}
	}
	
	// 1. 查询预处理
	processedQuery, concepts, err := kse.queryProcessor.Process(req.Query, req.Context)
	if err != nil {
		return nil, fmt.Errorf("查询预处理失败: %w", err)
	}
	
	// 2. 根据搜索类型执行搜索
	var results []*KnowledgeSearchResult
	var method string
	
	switch req.Type {
	case "vector":
		results, err = kse.vectorSearch(ctx, processedQuery, req)
		method = "vector"
	case "semantic":
		results, err = kse.semanticSearch(ctx, processedQuery, concepts, req)
		method = "semantic"
	case "reasoning":
		results, err = kse.reasoningSearch(ctx, processedQuery, concepts, req)
		method = "reasoning"
	case "hybrid":
		results, err = kse.hybridSearch(ctx, processedQuery, concepts, req)
		method = "hybrid"
	default:
		results, err = kse.hybridSearch(ctx, processedQuery, concepts, req)
		method = "hybrid_default"
	}
	
	if err != nil {
		kse.monitor.RecordError(err)
		return nil, err
	}
	
	// 3. 结果过滤和排序
	results = kse.filterResults(results, req)
	results = kse.resultRanker.Rank(results, req)
	
	// 4. 应用限制
	if req.MaxResults > 0 && len(results) > req.MaxResults {
		results = results[:req.MaxResults]
	}
	
	// 5. 生成相关建议
	suggestions := kse.generateSuggestions(processedQuery, concepts, results)
	relatedQueries := kse.generateRelatedQueries(processedQuery, concepts)
	
	// 6. 计算质量指标
	quality := kse.calculateQualityMetrics(results, req)
	
	// 7. 构建响应
	response := &KnowledgeSearchResponse{
		Results:        results,
		Total:          len(results),
		QueryTime:      time.Since(startTime),
		Method:         method,
		ProcessedQuery: processedQuery,
		Concepts:       concepts,
		Suggestions:    suggestions,
		RelatedQueries: relatedQueries,
		Quality:        quality,
		Metadata: map[string]interface{}{
			"original_query": req.Query,
			"search_domain":  req.Domain,
			"use_reasoning":  req.Type == "reasoning" || req.Type == "hybrid",
		},
	}
	
	// 8. 缓存结果
	if req.UseCache {
		kse.cache.Set(cacheKey, response, 15*time.Minute)
	}
	
	// 9. 记录搜索结果
	kse.monitor.RecordResult(req, response)
	
	return response, nil
}

// vectorSearch 向量搜索
func (kse *KnowledgeSearchEngine) vectorSearch(ctx context.Context, query string, req *KnowledgeSearchRequest) ([]*KnowledgeSearchResult, error) {
	// 1. 生成查询向量
	queryEmbedding, err := kse.embedder.GenerateEmbedding(query)
	if err != nil {
		return nil, fmt.Errorf("生成查询向量失败: %w", err)
	}
	
	// 2. 向量相似度搜索
	candidates := kse.vectorStore.SearchSimilar(queryEmbedding, req.MaxResults*2, req.MinRelevance)
	
	// 3. 转换为搜索结果
	var results []*KnowledgeSearchResult
	for _, candidate := range candidates {
		result := &KnowledgeSearchResult{
			ID:             candidate.ID,
			Content:        candidate.Content,
			Type:           candidate.Type,
			Category:       candidate.Category,
			Domain:         candidate.Domain,
			RelevanceScore: candidate.Quality,
			SemanticScore:  kse.calculateSemanticSimilarity(query, candidate.Content),
			QualityScore:   candidate.Quality,
			Usage:          candidate.Usage,
			CreatedAt:      candidate.Freshness,
			Metadata:       candidate.Metadata,
		}
		result.FinalScore = kse.calculateFinalScore(result)
		results = append(results, result)
	}
	
	return results, nil
}

// semanticSearch 语义搜索
func (kse *KnowledgeSearchEngine) semanticSearch(ctx context.Context, query string, concepts []string, req *KnowledgeSearchRequest) ([]*KnowledgeSearchResult, error) {
	// 1. 概念匹配搜索
	conceptResults := kse.searchByConcepts(concepts, req)
	
	// 2. 语义关系搜索
	relationResults := kse.searchBySemanticRelations(query, concepts, req)
	
	// 3. 合并结果
	merged := kse.mergeResults(conceptResults, relationResults)
	
	// 4. 语义相关性评分
	for _, result := range merged {
		result.SemanticScore = kse.calculateSemanticRelevance(query, concepts, result)
		result.FinalScore = kse.calculateFinalScore(result)
	}
	
	return merged, nil
}

// reasoningSearch 推理搜索
func (kse *KnowledgeSearchEngine) reasoningSearch(ctx context.Context, query string, concepts []string, req *KnowledgeSearchRequest) ([]*KnowledgeSearchResult, error) {
	// 1. 语义推理
	reasoningResult, err := kse.semanticReasoner.PerformReasoning(query, concepts, req.Context)
	if err != nil {
		global.GVA_LOG.Warn("语义推理失败", zap.Error(err))
		// 降级到语义搜索
		return kse.semanticSearch(ctx, query, concepts, req)
	}
	
	// 2. 基于推理结果搜索
	var results []*KnowledgeSearchResult
	
	// 🔧 修复：从ReasoningResult中安全获取数据
	resultMap, ok := reasoningResult.Result.(map[string]interface{})
	if !ok {
		resultMap = make(map[string]interface{})
	}
	
	// 添加推理得到的新概念
	explanations, _ := resultMap["explanations"].([]string)
	expandedConcepts := append(concepts, explanations...)
	
	// 搜索推理结果相关的知识
	inferences, _ := resultMap["inferences"].([]string)
	for _, inference := range inferences {
		inferenceResults := kse.searchByInference(inference, req)
		results = append(results, inferenceResults...)
	}
	
	// 搜索扩展概念相关的知识
	conceptResults := kse.searchByConcepts(expandedConcepts, req)
	results = append(results, conceptResults...)
	
	// 3. 去重和评分
	results = kse.deduplicateResults(results)
	for _, result := range results {
		result.SemanticScore = kse.calculateSemanticRelevance(query, expandedConcepts, result)
		result.Explanation = kse.generateReasoningExplanation(result, reasoningResult)
		result.FinalScore = kse.calculateFinalScore(result)
	}
	
	return results, nil
}

// hybridSearch 混合搜索
func (kse *KnowledgeSearchEngine) hybridSearch(ctx context.Context, query string, concepts []string, req *KnowledgeSearchRequest) ([]*KnowledgeSearchResult, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	var allResults []*KnowledgeSearchResult
	
	// 并行执行多种搜索
	searchMethods := []struct {
		name   string
		search func() ([]*KnowledgeSearchResult, error)
	}{
		{"vector", func() ([]*KnowledgeSearchResult, error) { 
			return kse.vectorSearch(ctx, query, req) 
		}},
		{"semantic", func() ([]*KnowledgeSearchResult, error) { 
			return kse.semanticSearch(ctx, query, concepts, req) 
		}},
	}
	
	// 如果启用推理搜索
	if req.IncludeRules {
		searchMethods = append(searchMethods, struct {
			name   string
			search func() ([]*KnowledgeSearchResult, error)
		}{"reasoning", func() ([]*KnowledgeSearchResult, error) { 
			return kse.reasoningSearch(ctx, query, concepts, req) 
		}})
	}
	
	for _, method := range searchMethods {
		wg.Add(1)
		go func(m struct {
			name   string
			search func() ([]*KnowledgeSearchResult, error)
		}) {
			defer wg.Done()
			
			results, err := m.search()
			if err != nil {
				global.GVA_LOG.Warn("搜索方法执行失败", 
					zap.String("method", m.name), 
					zap.Error(err))
				return
			}
			
			// 标记搜索方法
			for _, result := range results {
				if result.Metadata == nil {
					result.Metadata = make(map[string]interface{})
				}
				result.Metadata["search_method"] = m.name
			}
			
			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(method)
	}
	
	wg.Wait()
	
	// 融合结果
	fusedResults := kse.fuseResults(allResults)
	
	return fusedResults, nil
}

// BuildKnowledgeVectorIndex 构建知识库向量索引
func (kse *KnowledgeSearchEngine) BuildKnowledgeVectorIndex(ctx context.Context) error {
	global.GVA_LOG.Info("开始构建知识库向量索引")
	startTime := time.Now()
	
	// 1. 获取所有知识条目
	var knowledgeEntries []nesma.NesmaKnowledgeEntry
	if err := global.GVA_DB.Where("status = ?", "active").Find(&knowledgeEntries).Error; err != nil {
		return fmt.Errorf("获取知识条目失败: %w", err)
	}
	
	// 2. 批量生成向量
	batchSize := 10
	for i := 0; i < len(knowledgeEntries); i += batchSize {
		end := i + batchSize
		if end > len(knowledgeEntries) {
			end = len(knowledgeEntries)
		}
		
		batch := knowledgeEntries[i:end]
		if err := kse.processBatch(batch); err != nil {
			global.GVA_LOG.Error("批量处理失败", 
				zap.Int("batch_start", i), 
				zap.Error(err))
			continue
		}
		
		// 进度记录
		if (i/batchSize+1)%10 == 0 {
			global.GVA_LOG.Info("索引构建进度", 
				zap.Int("processed", i+len(batch)),
				zap.Int("total", len(knowledgeEntries)))
		}
	}
	
	// 3. 构建索引
	if err := kse.vectorStore.BuildIndex(); err != nil {
		return fmt.Errorf("构建向量索引失败: %w", err)
	}
	
	global.GVA_LOG.Info("知识库向量索引构建完成", 
		zap.Duration("耗时", time.Since(startTime)),
		zap.Int("条目数量", len(knowledgeEntries)))
	
	return nil
}

// processBatch 处理批量知识条目
func (kse *KnowledgeSearchEngine) processBatch(entries []nesma.NesmaKnowledgeEntry) error {
	var texts []string
	for _, entry := range entries {
		content := fmt.Sprintf("%s %s", entry.Title, entry.Content)
		texts = append(texts, content)
	}
	
	// 批量生成嵌入向量
	embeddings, err := kse.embedder.GenerateBatchEmbeddings(texts)
	if err != nil {
		return fmt.Errorf("批量生成嵌入向量失败: %w", err)
	}
	
	// 更新向量存储
	for i, entry := range entries {
		if i >= len(embeddings) {
			break
		}
		
		vector := &KnowledgeVector{
			ID:        fmt.Sprintf("knowledge_%d", entry.ID),
			Content:   entry.Content,
			Embedding: embeddings[i],
			Type:      "knowledge",
			Category:  entry.Category,
			Domain:    entry.Domain,
			Concepts:  kse.extractConcepts(entry.Content),
			Quality:   entry.ConfidenceScore,
			Freshness: entry.CreatedAt,
			Usage:     int(entry.UsageCount),
			Metadata: map[string]interface{}{
				"entry_id": entry.ID,
				"title":    entry.Title,
				"source":   entry.Source,
				"author":   entry.Author,
				"tags":     entry.Tags,
			},
		}
		
		kse.vectorStore.AddVector(vector)
		
		// 更新数据库中的嵌入向量
		embeddingJSON, _ := json.Marshal(embeddings[i])
		global.GVA_DB.Model(&entry).Update("embedding", string(embeddingJSON))
	}
	
	return nil
}

// 辅助方法实现
func (kse *KnowledgeSearchEngine) generateCacheKey(req *KnowledgeSearchRequest) string {
	return fmt.Sprintf("ks_%s_%s_%s_%d_%f", 
		req.Query, req.Type, req.Domain, req.MaxResults, req.MinRelevance)
}

func (kse *KnowledgeSearchEngine) filterResults(results []*KnowledgeSearchResult, req *KnowledgeSearchRequest) []*KnowledgeSearchResult {
	var filtered []*KnowledgeSearchResult
	
	for _, result := range results {
		// 域过滤
		if req.Domain != "" && result.Domain != req.Domain {
			continue
		}
		
		// 类别过滤
		if len(req.Category) > 0 {
			categoryMatch := false
			for _, cat := range req.Category {
				if result.Category == cat {
					categoryMatch = true
					break
				}
			}
			if !categoryMatch {
				continue
			}
		}
		
		// 相关性阈值过滤
		if result.RelevanceScore < req.MinRelevance {
			continue
		}
		
		filtered = append(filtered, result)
	}
	
	return filtered
}

func (kse *KnowledgeSearchEngine) calculateSemanticSimilarity(query, content string) float64 {
	// 简化的语义相似度计算
	queryWords := strings.Fields(strings.ToLower(query))
	contentWords := strings.Fields(strings.ToLower(content))
	
	intersection := 0
	for _, qw := range queryWords {
		for _, cw := range contentWords {
			if qw == cw {
				intersection++
				break
			}
		}
	}
	
	union := len(queryWords) + len(contentWords) - intersection
	if union == 0 {
		return 0
	}
	
	return float64(intersection) / float64(union)
}

func (kse *KnowledgeSearchEngine) calculateFinalScore(result *KnowledgeSearchResult) float64 {
	// 综合评分计算
	weights := map[string]float64{
		"relevance": 0.4,
		"semantic":  0.3,
		"quality":   0.2,
		"usage":     0.1,
	}
	
	usageScore := math.Min(float64(result.Usage)/100.0, 1.0)
	
	finalScore := weights["relevance"]*result.RelevanceScore +
		weights["semantic"]*result.SemanticScore +
		weights["quality"]*result.QualityScore +
		weights["usage"]*usageScore
	
	return math.Min(finalScore, 1.0)
}

func (kse *KnowledgeSearchEngine) extractConcepts(content string) []string {
	// 简化的概念提取
	words := strings.Fields(strings.ToLower(content))
	concepts := make(map[string]bool)
	
	nesmaTerms := []string{
		"功能点", "数据功能", "事务功能", "外部输入", "外部输出", "外部查询",
		"内部逻辑文件", "外部接口文件", "复杂度", "低复杂度", "中等复杂度", "高复杂度",
		"用户需求", "软件需求", "功能需求", "非功能需求", "业务规则",
	}
	
	for _, word := range words {
		for _, term := range nesmaTerms {
			if strings.Contains(term, word) || strings.Contains(word, term) {
				concepts[term] = true
			}
		}
	}
	
	var result []string
	for concept := range concepts {
		result = append(result, concept)
	}
	
	return result
}

// 占位符方法实现
func NewAIEmbeddingService() *AIEmbeddingService {
	return &AIEmbeddingService{
		deepseeekClient: NewDeepSeekClient(),
		cache:          &EmbeddingCache{},
		batchSize:      10,
	}
}

func NewEnhancedVectorStore() *EnhancedVectorStore {
	return &EnhancedVectorStore{
		vectors:  make(map[string]*KnowledgeVector),
		index:    &SemanticIndex{},
		metadata: &VectorMetadata{},
	}
}

func NewSemanticReasoningEngine() *SemanticReasoningEngine {
	return &SemanticReasoningEngine{
		ruleBase:    &NESMARuleBase{},
		ontology:    &NESMAOntology{},
		inferEngine: &LogicalInferenceEngine{},
		explainer:   &ReasoningExplainer{},
	}
}

func NewQueryProcessor() *QueryProcessor {
	return &QueryProcessor{
		analyzer:   &SemanticAnalyzer{},
		expander:   &QueryExpander{},
		optimizer:  &QueryOptimizer{},
		normalizer: &QueryNormalizer{},
	}
}

func NewSemanticResultRanker() *SemanticResultRanker {
	return &SemanticResultRanker{
		weights: map[string]float64{
			"relevance": 0.4,
			"semantic":  0.3,
			"quality":   0.2,
			"freshness": 0.1,
		},
		boosters:    make([]RelevanceBooster, 0),
		diversifier: &ResultDiversifier{},
	}
}

func NewSearchCache() *SearchCache {
	return &SearchCache{}
}

func NewSearchMonitor() *SearchMonitor {
	return &SearchMonitor{}
}

func (aes *AIEmbeddingService) GenerateEmbedding(text string) ([]float64, error) {
	// TODO: 集成DeepSeek API生成真实嵌入向量
	return generateMockEmbedding(text, 1536), nil
}

func (aes *AIEmbeddingService) GenerateBatchEmbeddings(texts []string) ([][]float64, error) {
	var embeddings [][]float64
	for _, text := range texts {
		embedding, err := aes.GenerateEmbedding(text)
		if err != nil {
			return nil, err
		}
		embeddings = append(embeddings, embedding)
	}
	return embeddings, nil
}

func generateMockEmbedding(text string, dim int) []float64 {
	// 生成模拟嵌入向量
	hash := 0
	for _, c := range text {
		hash = hash*31 + int(c)
	}
	
	vector := make([]float64, dim)
	for i := range vector {
		vector[i] = math.Sin(float64(hash+i)) * 0.3
	}
	
	return vector
}

// 占位符类型
type DeepSeekClient struct{}
type SemanticIndex struct{}
type VectorMetadata struct{}
type NESMARuleBase struct{}
type LogicalInferenceEngine struct{}
type SemanticAnalyzer struct{}
type QueryExpander struct{}
type QueryOptimizer struct{}
type QueryNormalizer struct{}
type RelevanceBooster interface{}
type ResultDiversifier struct{}
type SearchCache struct{}
type SearchMonitor struct{}

// 🔧 添加缺失的类型定义
type EmbeddingCache struct {
	cache map[string][]float64
	mu    sync.RWMutex
}

type NESMAOntology struct {
	concepts    map[string]*Concept
	relations   map[string]*Relation
	hierarchy   map[string][]string
}

type ReasoningExplainer struct {
	rules       []ReasoningRule
	explanations map[string]string
}

type Concept struct {
	ID          string
	Name        string
	Description string
	Properties  map[string]interface{}
}

type Relation struct {
	ID     string
	Source string
	Target string
	Type   string
}

type ReasoningRule struct {
	ID        string
	Condition string
	Action    string
	Weight    float64
}

type ReasoningResult struct {
	Result      interface{}
	Confidence  float64
	Explanation string
}

// 占位符方法
func NewDeepSeekClient() *DeepSeekClient { return &DeepSeekClient{} }

func (qp *QueryProcessor) Process(query string, context map[string]interface{}) (string, []string, error) {
	concepts := qp.analyzer.ExtractConcepts(query)
	expandedQuery := qp.expander.Expand(query, concepts)
	return expandedQuery, concepts, nil
}

func (sa *SemanticAnalyzer) ExtractConcepts(query string) []string {
	words := strings.Fields(strings.ToLower(query))
	nesmaTerms := []string{"功能点", "数据功能", "事务功能", "复杂度", "需求"}
	
	var concepts []string
	for _, word := range words {
		for _, term := range nesmaTerms {
			if strings.Contains(word, term) {
				concepts = append(concepts, term)
			}
		}
	}
	return concepts
}

func (qe *QueryExpander) Expand(query string, concepts []string) string {
	return query // 简化实现
}

func (evs *EnhancedVectorStore) SearchSimilar(embedding []float64, limit int, threshold float64) []*KnowledgeVector {
	// 简化的相似度搜索
	var results []*KnowledgeVector
	for _, vector := range evs.vectors {
		similarity := calculateCosineSimilarity(embedding, vector.Embedding)
		if similarity >= threshold {
			results = append(results, vector)
		}
	}
	
	// 按相似度排序
	sort.Slice(results, func(i, j int) bool {
		sim1 := calculateCosineSimilarity(embedding, results[i].Embedding)
		sim2 := calculateCosineSimilarity(embedding, results[j].Embedding)
		return sim1 > sim2
	})
	
	if len(results) > limit {
		results = results[:limit]
	}
	
	return results
}

func (evs *EnhancedVectorStore) AddVector(vector *KnowledgeVector) {
	evs.mu.Lock()
	defer evs.mu.Unlock()
	evs.vectors[vector.ID] = vector
}

func (evs *EnhancedVectorStore) BuildIndex() error {
	// 构建向量索引
	return nil
}

func (kse *KnowledgeSearchEngine) searchByConcepts(concepts []string, req *KnowledgeSearchRequest) []*KnowledgeSearchResult {
	return []*KnowledgeSearchResult{}
}

func (kse *KnowledgeSearchEngine) searchBySemanticRelations(query string, concepts []string, req *KnowledgeSearchRequest) []*KnowledgeSearchResult {
	return []*KnowledgeSearchResult{}
}

func (kse *KnowledgeSearchEngine) mergeResults(results1, results2 []*KnowledgeSearchResult) []*KnowledgeSearchResult {
	return append(results1, results2...)
}

func (kse *KnowledgeSearchEngine) calculateSemanticRelevance(query string, concepts []string, result *KnowledgeSearchResult) float64 {
	return 0.8
}

func (kse *KnowledgeSearchEngine) searchByInference(inference string, req *KnowledgeSearchRequest) []*KnowledgeSearchResult {
	return []*KnowledgeSearchResult{}
}

func (kse *KnowledgeSearchEngine) deduplicateResults(results []*KnowledgeSearchResult) []*KnowledgeSearchResult {
	seen := make(map[string]bool)
	var unique []*KnowledgeSearchResult
	
	for _, result := range results {
		if !seen[result.ID] {
			seen[result.ID] = true
			unique = append(unique, result)
		}
	}
	
	return unique
}

func (kse *KnowledgeSearchEngine) generateReasoningExplanation(result *KnowledgeSearchResult, reasoning interface{}) string {
	return "基于语义推理匹配的相关知识"
}

func (kse *KnowledgeSearchEngine) fuseResults(allResults []*KnowledgeSearchResult) []*KnowledgeSearchResult {
	// 简化的结果融合
	return kse.deduplicateResults(allResults)
}

func (kse *KnowledgeSearchEngine) generateSuggestions(query string, concepts []string, results []*KnowledgeSearchResult) []string {
	return []string{}
}

func (kse *KnowledgeSearchEngine) generateRelatedQueries(query string, concepts []string) []string {
	return []string{}
}

func (kse *KnowledgeSearchEngine) calculateQualityMetrics(results []*KnowledgeSearchResult, req *KnowledgeSearchRequest) *SearchQualityMetrics {
	return &SearchQualityMetrics{
		Precision:     0.85,
		Coverage:      0.75,
		Diversity:     0.70,
		SemanticDepth: 0.80,
		Freshness:     0.90,
		Confidence:    0.82,
	}
}

func (srr *SemanticResultRanker) Rank(results []*KnowledgeSearchResult, req *KnowledgeSearchRequest) []*KnowledgeSearchResult {
	sort.Slice(results, func(i, j int) bool {
		return results[i].FinalScore > results[j].FinalScore
	})
	return results
}

func (sre *SemanticReasoningEngine) PerformReasoning(query string, concepts []string, context map[string]interface{}) (*ReasoningResult, error) {
	// 🔧 修复：使用正确的ReasoningResult结构
	result := map[string]interface{}{
		"query":        query,
		"entities":     []string{},
		"relations":    []string{},
		"inferences":   []string{"基于NESMA标准的推理结果"},
		"explanations": concepts,
		"method":       "semantic_inference",
		"metadata":     context,
	}
	
	return &ReasoningResult{
		Result:      result,
		Confidence:  0.8,
		Explanation: "基于NESMA标准的语义推理分析",
	}, nil
}

func (sc *SearchCache) Get(key string) interface{} { return nil }
func (sc *SearchCache) Set(key string, value interface{}, ttl time.Duration) {}

func (sm *SearchMonitor) RecordRequest(req *KnowledgeSearchRequest) {}
func (sm *SearchMonitor) RecordCacheHit() {}
func (sm *SearchMonitor) RecordError(err error) {}
func (sm *SearchMonitor) RecordResult(req *KnowledgeSearchRequest, resp *KnowledgeSearchResponse) {}

// 🔧 添加缺失的calculateCosineSimilarity函数
func calculateCosineSimilarity(a, b []float64) float64 {
	if len(a) != len(b) {
		return 0.0
	}
	
	var dotProduct, normA, normB float64
	
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	
	normA = math.Sqrt(normA)
	normB = math.Sqrt(normB)
	
	if normA == 0 || normB == 0 {
		return 0.0
	}
	
	return dotProduct / (normA * normB)
}