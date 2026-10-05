package nesma

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
	"sync" // Added for concurrent processing

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MermaidGenerationResult Mermaid流程图生成结果
type MermaidGenerationResult struct {
	RequirementID        uint                     `json:"requirement_id"`
	OriginalRequirement  *nesma.NesmaRequirement  `json:"original_requirement"`
	GeneratedMermaid     MermaidDiagram           `json:"generated_mermaid"`
	AnalysisConfidence   float64                  `json:"analysis_confidence"`
	ProcessingTime       int64                    `json:"processing_time_ms"`
	GenerationNotes      string                   `json:"generation_notes"`
	KnowledgeReferences  []KnowledgeReference     `json:"knowledge_refs"`
	ValidationResults    MermaidValidationResult  `json:"validation_results"`
}

// MermaidDiagram Mermaid流程图
type MermaidDiagram struct {
	DiagramType     string           `json:"diagram_type"`     // flowchart/sequence/class/state
	Title           string           `json:"title"`
	Description     string           `json:"description"`
	MermaidCode     string           `json:"mermaid_code"`
	Nodes           []MermaidNode    `json:"nodes"`
	Edges           []MermaidEdge    `json:"edges"`
	Subgraphs       []MermaidSubgraph `json:"subgraphs"`
	Styling         MermaidStyling   `json:"styling"`
	Metadata        MermaidMetadata  `json:"metadata"`
}

// MermaidNode Mermaid节点
type MermaidNode struct {
	ID          string            `json:"id"`
	Label       string            `json:"label"`
	Type        string            `json:"type"`        // start/end/process/decision/data/connector
	Shape       string            `json:"shape"`       // rectangle/circle/rhombus/cylinder/etc
	Category    string            `json:"category"`    // input/process/output/decision/exception
	Properties  map[string]string `json:"properties"`
	Validation  NodeValidation    `json:"validation"`
}

// MermaidEdge Mermaid边
type MermaidEdge struct {
	ID         string            `json:"id"`
	From       string            `json:"from"`
	To         string            `json:"to"`
	Label      string            `json:"label"`
	Type       string            `json:"type"`        // arrow/dotted/thick/etc
	Condition  string            `json:"condition"`   // for decision nodes
	Properties map[string]string `json:"properties"`
}

// MermaidSubgraph Mermaid子图
type MermaidSubgraph struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Nodes       []string `json:"nodes"`
	Style       string   `json:"style"`
}

// MermaidStyling Mermaid样式
type MermaidStyling struct {
	Theme          string                 `json:"theme"`           // default/dark/forest/etc
	NodeStyles     map[string]string      `json:"node_styles"`
	EdgeStyles     map[string]string      `json:"edge_styles"`
	CustomCSS      string                 `json:"custom_css"`
	ColorScheme    map[string]string      `json:"color_scheme"`
}

// MermaidMetadata Mermaid元数据
type MermaidMetadata struct {
	ComplexityLevel string    `json:"complexity_level"`  // simple/medium/complex
	NodeCount       int       `json:"node_count"`
	EdgeCount       int       `json:"edge_count"`
	MaxDepth        int       `json:"max_depth"`
	GeneratedAt     time.Time `json:"generated_at"`
	Version         string    `json:"version"`
}

// NodeValidation 节点验证
type NodeValidation struct {
	IsValid    bool     `json:"is_valid"`
	Issues     []string `json:"issues"`
	Warnings   []string `json:"warnings"`
}

// MermaidValidationResult Mermaid验证结果
type MermaidValidationResult struct {
	IsValid        bool     `json:"is_valid"`
	SyntaxValid    bool     `json:"syntax_valid"`
	StructureValid bool     `json:"structure_valid"`
	ComplexityOK   bool     `json:"complexity_ok"`
	Issues         []string `json:"issues"`
	Warnings       []string `json:"warnings"`
	Suggestions    []string `json:"suggestions"`
}

// MermaidGeneratorService Mermaid流程图生成服务
type MermaidGeneratorService struct {
	db               *gorm.DB
	aiService        AIService
	knowledgeService *KnowledgeService
	maxNodes         int
	maxDepth         int
}

// NewMermaidGeneratorService 创建Mermaid流程图生成服务
func NewMermaidGeneratorService() *MermaidGeneratorService {
	aiService := GetAIService()
	if aiService == nil {
		// 避免在logger未初始化时调用log，可能导致空指针
		if global.GVA_CONFIG.AI.DeepSeek.APIKey != "" {
			aiService = NewDeepSeekService(
				global.GVA_CONFIG.AI.DeepSeek.APIKey,
				global.GVA_CONFIG.AI.DeepSeek.BaseURL,
				global.GVA_CONFIG.AI.DeepSeek.Model,
			)
		}
	}
	
	// 安全获取配置值
	maxNodes := global.GVA_CONFIG.AI.Mermaid.MaxNodes
	maxDepth := global.GVA_CONFIG.AI.Mermaid.MaxDepth
	if maxNodes == 0 {
		maxNodes = 50 // 默认值
	}
	if maxDepth == 0 {
		maxDepth = 10 // 默认值
	}
	
	return &MermaidGeneratorService{
		db:               global.GVA_DB,
		aiService:        aiService,
		knowledgeService: NewKnowledgeService(),
		maxNodes:         maxNodes,
		maxDepth:         maxDepth,
	}
}

// GenerateMermaidDiagrams 生成Mermaid流程图 - 多线程版本
func (s *MermaidGeneratorService) GenerateMermaidDiagrams(cycleID uint, requirementIDs []uint, diagramType string) ([]*MermaidGenerationResult, error) {
	global.GVA_LOG.Info("开始生成Mermaid流程图", zap.Uint("cycleID", cycleID), zap.Any("requirementIDs", requirementIDs), zap.String("diagramType", diagramType))

	// 使用Mermaid专用配置
	mermaidConfig := global.GVA_CONFIG.AI.Mermaid
	maxWorkers := mermaidConfig.MaxWorkers
	batchSize := mermaidConfig.BatchSize
	
	global.GVA_LOG.Info("Mermaid生成配置", 
		zap.Int("maxWorkers", maxWorkers), 
		zap.Int("batchSize", batchSize),
		zap.Int("maxNodes", mermaidConfig.MaxNodes),
		zap.Int("maxDepth", mermaidConfig.MaxDepth))

	var results []*MermaidGenerationResult
	
	// 获取周期信息
	var cycle nesma.NesmaProjectCycle
	if err := s.db.Preload("Project").First(&cycle, cycleID).Error; err != nil {
		return nil, fmt.Errorf("获取项目周期失败: %w", err)
	}

	// 获取需要生成流程图的功能点
	var requirements []nesma.NesmaRequirement
	query := s.db.Where("cycle_id = ? AND level = 4", cycleID)
	
	if len(requirementIDs) > 0 {
		query = query.Where("id IN ?", requirementIDs)
	}
	
	if err := query.Find(&requirements).Error; err != nil {
		return nil, fmt.Errorf("获取功能点失败: %w", err)
	}

	// 使用多线程并发处理
	if len(requirements) > 1 && maxWorkers > 1 {
		return s.generateMermaidDiagramsConcurrently(requirements, &cycle, diagramType, maxWorkers, batchSize)
	} else {
		// 单线程处理
		for _, requirement := range requirements {
			result, err := s.generateMermaidDiagram(&requirement, &cycle, diagramType)
			if err != nil {
				global.GVA_LOG.Error("生成Mermaid流程图失败", zap.Uint("requirementID", requirement.ID), zap.Error(err))
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// generateMermaidDiagramsConcurrently 并发生成Mermaid流程图
func (s *MermaidGeneratorService) generateMermaidDiagramsConcurrently(
	requirements []nesma.NesmaRequirement, 
	cycle *nesma.NesmaProjectCycle, 
	diagramType string, 
	maxWorkers int, 
	batchSize int) ([]*MermaidGenerationResult, error) {

	// 创建任务通道和结果通道
	taskChan := make(chan *nesma.NesmaRequirement, len(requirements))
	resultChan := make(chan *MermaidGenerationResult, len(requirements))
	
	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go s.mermaidWorker(i, taskChan, resultChan, cycle, diagramType, &wg)
	}
	
	// 发送任务到通道
	for i := range requirements {
		taskChan <- &requirements[i]
	}
	close(taskChan)
	
	// 等待所有工作协程完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()
	
	// 收集结果
	var results []*MermaidGenerationResult
	for result := range resultChan {
		results = append(results, result)
	}

	return results, nil
}

// mermaidWorker Mermaid生成工作协程
func (s *MermaidGeneratorService) mermaidWorker(
	workerID int, 
	taskChan <-chan *nesma.NesmaRequirement, 
	resultChan chan<- *MermaidGenerationResult, 
	cycle *nesma.NesmaProjectCycle, 
	diagramType string, 
	wg *sync.WaitGroup) {
	
	defer wg.Done()
	
	// 每个工作协程创建独立的AI服务实例
	aiService := NewDeepSeekService(global.GVA_CONFIG.AI.DeepSeek.APIKey, global.GVA_CONFIG.AI.DeepSeek.BaseURL, "deepseek-chat")
	if aiService == nil {
		global.GVA_LOG.Error("工作协程AI服务初始化失败", zap.Int("workerID", workerID))
		return
	}
	
	global.GVA_LOG.Info("Mermaid生成工作协程启动", zap.Int("workerID", workerID))
	
	for req := range taskChan {
		global.GVA_LOG.Debug("工作协程开始处理Mermaid生成", 
			zap.Int("workerID", workerID),
			zap.Uint("requirementID", req.ID),
			zap.String("title", req.Title))
		
		result, err := s.generateMermaidDiagramWithAI(req, cycle, diagramType, aiService)
		if err != nil {
			global.GVA_LOG.Error("Mermaid生成失败", 
				zap.Int("workerID", workerID),
				zap.Error(err), 
				zap.String("title", req.Title))
			continue
		}
		
		resultChan <- result
	}
	
	global.GVA_LOG.Info("Mermaid生成工作协程结束", zap.Int("workerID", workerID))
}

// generateMermaidDiagram 生成单个Mermaid流程图
func (s *MermaidGeneratorService) generateMermaidDiagram(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, diagramType string) (*MermaidGenerationResult, error) {
	startTime := time.Now()
	
	// 1. 搜索相关知识库内容
	knowledgeRefs, err := s.searchRelevantKnowledge(requirement, cycle)
	if err != nil {
		global.GVA_LOG.Warn("搜索相关知识失败", zap.Error(err))
		knowledgeRefs = []KnowledgeReference{}
	}

	// 2. 生成Mermaid流程图
	mermaidDiagram, err := s.generateMermaidStructure(requirement, cycle, diagramType, knowledgeRefs)
	if err != nil {
		return nil, fmt.Errorf("生成Mermaid流程图失败: %w", err)
	}

	// 3. 验证流程图
	validationResult := s.validateMermaidDiagram(mermaidDiagram)

	// 4. 计算置信度
	confidence := s.calculateMermaidConfidence(mermaidDiagram, validationResult, knowledgeRefs)

	// 5. 构建生成结果
	result := &MermaidGenerationResult{
		RequirementID:       requirement.ID,
		OriginalRequirement: requirement,
		GeneratedMermaid:    *mermaidDiagram,
		AnalysisConfidence:  confidence,
		ProcessingTime:      time.Since(startTime).Milliseconds(),
		GenerationNotes:     s.generateMermaidNotes(requirement, mermaidDiagram),
		KnowledgeReferences: knowledgeRefs,
		ValidationResults:   validationResult,
	}

	global.GVA_LOG.Info("Mermaid流程图生成完成", 
		zap.Uint("requirementID", requirement.ID),
		zap.Float64("置信度", confidence),
		zap.Int("节点数", len(mermaidDiagram.Nodes)),
		zap.Int("边数", len(mermaidDiagram.Edges)),
	)

	return result, nil
}

// generateMermaidDiagramWithAI 使用指定AI服务生成Mermaid流程图
func (s *MermaidGeneratorService) generateMermaidDiagramWithAI(
	requirement *nesma.NesmaRequirement, 
	cycle *nesma.NesmaProjectCycle, 
	diagramType string, 
	aiService AIService) (*MermaidGenerationResult, error) {
	
	startTime := time.Now()
	
	// 搜索相关知识库内容
	knowledgeRefs, err := s.searchRelevantKnowledge(requirement, cycle)
	if err != nil {
		global.GVA_LOG.Warn("搜索相关知识失败", zap.Error(err))
		knowledgeRefs = []KnowledgeReference{}
	}

	// 生成Mermaid流程图
	mermaidDiagram, err := s.generateMermaidStructureWithAI(requirement, cycle, diagramType, knowledgeRefs, aiService)
	if err != nil {
		return nil, fmt.Errorf("生成Mermaid流程图失败: %w", err)
	}

	// 验证流程图
	validationResult := s.validateMermaidDiagram(mermaidDiagram)

	// 计算置信度
	confidence := s.calculateMermaidConfidence(mermaidDiagram, validationResult, knowledgeRefs)

	// 构建生成结果
	result := &MermaidGenerationResult{
		RequirementID:       requirement.ID,
		OriginalRequirement: requirement,
		GeneratedMermaid:    *mermaidDiagram,
		AnalysisConfidence:  confidence,
		ProcessingTime:      time.Since(startTime).Milliseconds(),
		GenerationNotes:     s.generateMermaidNotes(requirement, mermaidDiagram),
		KnowledgeReferences: knowledgeRefs,
		ValidationResults:   validationResult,
	}

	return result, nil
}

// searchRelevantKnowledge 搜索相关知识库内容
func (s *MermaidGeneratorService) searchRelevantKnowledge(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle) ([]KnowledgeReference, error) {
	// 空指针检查
	if s.knowledgeService == nil {
		global.GVA_LOG.Warn("知识库服务未初始化，跳过知识库搜索")
		return []KnowledgeReference{}, nil
	}
	
	// 构建搜索查询
	searchQuery := fmt.Sprintf("%s %s %s 流程图 业务流程", requirement.Title, requirement.Description, cycle.Project.Domain)
	
	// 调用知识库搜索服务
	searchReq := &request.SearchKnowledgeEntriesRequest{
		Query: searchQuery,
		Limit: 5,
	}
	
	searchResp, err := s.knowledgeService.SearchKnowledgeEntries(context.Background(), searchReq)
	if err != nil {
		global.GVA_LOG.Warn("搜索知识库失败", zap.Error(err))
		return []KnowledgeReference{}, nil // 返回空列表而不是错误，让流程继续
	}

	// 转换为知识库引用格式
	var references []KnowledgeReference
	for _, result := range searchResp.Results {
		references = append(references, KnowledgeReference{
			ID:        result.Knowledge.ID,
			Title:     result.Knowledge.Title,
			Category:  result.Knowledge.Category,
			Relevance: float64(result.Similarity),
		})
	}

	return references, nil
}

// generateMermaidStructure 生成Mermaid结构
func (s *MermaidGeneratorService) generateMermaidStructure(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, diagramType string, knowledgeRefs []KnowledgeReference) (*MermaidDiagram, error) {
	// 构建AI分析prompt
	prompt := s.buildMermaidPrompt(requirement, cycle, diagramType, knowledgeRefs)
	
	// 调用AI服务
	messages := []APIMessage{
		{Role: "system", Content: `你是一个专业的NESMA功能点分析师和业务流程建模专家，具备以下专业能力：

1. **NESMA功能点分析专业知识**
   - 精通EI（外部输入）、EO（外部输出）、EQ（外部查询）、ILF（内部逻辑文件）、EIF（外部接口文件）五大功能类型
   - 理解每种功能类型的特点、计数规则和复杂度评估标准
   - 能够识别功能点的边界和数据流向

2. **业务流程建模技能**
   - 擅长使用Mermaid语法创建清晰、准确的业务流程图
   - 能够将复杂业务逻辑转化为标准化的流程图
   - 注重流程的完整性、逻辑性和可维护性

3. **质量保证标准**
   - 确保生成的流程图符合NESMA功能点分析要求
   - 包含必要的异常处理和错误分支
   - 使用标准化的节点形状和连接线样式
   - 注重业务场景的真实性和实用性

4. **输出规范**
   - 严格按照指定的JSON格式输出结果
   - 确保Mermaid语法正确无误
   - 提供详细的节点和边的属性信息
   - 包含完整的样式和元数据信息

请根据提供的功能点信息，生成符合NESMA标准的专业流程图。`},
		{Role: "user", Content: prompt},
	}

	config := &AIConfig{
		MaxTokens:   8192,
		Temperature: 0.3,  // 降低温度以获得更一致和准确的输出
		TopP:        0.8,  // 稍微降低top_p以获得更聚焦的输出
		Model:       "deepseek-chat",
	}

	response, err := s.aiService.ChatCompletion(context.Background(), messages, config)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	return s.parseMermaidResponse(response.Text, diagramType)
}

// generateMermaidStructureWithAI 使用指定AI服务生成Mermaid结构
func (s *MermaidGeneratorService) generateMermaidStructureWithAI(
	requirement *nesma.NesmaRequirement, 
	cycle *nesma.NesmaProjectCycle, 
	diagramType string, 
	knowledgeRefs []KnowledgeReference, 
	aiService AIService) (*MermaidDiagram, error) {
	
	// 构建AI分析prompt
	prompt := s.buildMermaidPrompt(requirement, cycle, diagramType, knowledgeRefs)
	
	// 使用Mermaid专用AI配置
	mermaidConfig := global.GVA_CONFIG.AI.Mermaid
	config := &AIConfig{
		MaxTokens:   mermaidConfig.MaxTokens,
		Temperature: mermaidConfig.Temperature,
		TopP:        mermaidConfig.TopP,
		Model:       mermaidConfig.Model,
	}
	
	// 调用AI服务
	messages := []APIMessage{
		{Role: "system", Content: `你是一个专业的NESMA功能点分析师和业务流程建模专家，具备以下专业能力：

1. **NESMA功能点分析专业知识**
   - 精通EI（外部输入）、EO（外部输出）、EQ（外部查询）、ILF（内部逻辑文件）、EIF（外部接口文件）五大功能类型
   - 理解每种功能类型的特点、计数规则和复杂度评估标准
   - 能够识别功能点的边界和数据流向

2. **业务流程建模技能**
   - 擅长使用Mermaid语法创建清晰、准确的业务流程图
   - 能够将复杂业务逻辑转化为标准化的流程图
   - 注重流程的完整性、逻辑性和可维护性

3. **质量保证标准**
   - 确保生成的流程图符合NESMA功能点分析要求
   - 包含必要的异常处理和错误分支
   - 使用标准化的节点形状和连接线样式
   - 注重业务场景的真实性和实用性

4. **输出规范**
   - 严格按照指定的JSON格式输出结果
   - 确保Mermaid语法正确无误
   - 提供详细的节点和边的属性信息
   - 包含完整的样式和元数据信息

请根据提供的功能点信息，生成符合NESMA标准的专业流程图。`},
		{Role: "user", Content: prompt},
	}

	response, err := aiService.ChatCompletion(context.Background(), messages, config)
	if err != nil {
		return nil, fmt.Errorf("AI分析失败: %w", err)
	}

	// 解析AI响应
	return s.parseMermaidResponse(response.Text, diagramType)
}

// buildMermaidPrompt 构建Mermaid生成的prompt
func (s *MermaidGeneratorService) buildMermaidPrompt(requirement *nesma.NesmaRequirement, cycle *nesma.NesmaProjectCycle, diagramType string, knowledgeRefs []KnowledgeReference) string {
	var knowledgeContext strings.Builder
	for _, ref := range knowledgeRefs {
		knowledgeContext.WriteString(fmt.Sprintf("- %s: %s\n", ref.Title, ref.Category))
	}

	// 根据功能类型提供特定的指导
	var functionTypeGuidance string
	switch requirement.FunctionType {
	case "EI":
		functionTypeGuidance = `
【EI功能点特点】
- 外部输入功能：从系统外部接收数据
- 必须包含数据验证和错误处理
- 通常涉及数据转换和格式化
- 需要记录操作日志和审计信息
- 重点关注数据完整性和业务规则校验`
	case "EO":
		functionTypeGuidance = `
【EO功能点特点】
- 外部输出功能：向系统外部发送数据
- 包含数据处理和格式化逻辑
- 通常涉及报表生成或数据导出
- 需要数据聚合和计算处理
- 重点关注输出格式和性能优化`
	case "EQ":
		functionTypeGuidance = `
【EQ功能点特点】
- 外部查询功能：响应用户查询请求
- 包含查询条件处理和结果过滤
- 通常涉及多表关联和复杂查询
- 需要分页和排序功能
- 重点关注查询性能和用户体验`
	case "ILF":
		functionTypeGuidance = `
【ILF功能点特点】
- 内部逻辑文件：维护系统内部数据
- 包含数据增删改查操作
- 通常涉及数据完整性约束
- 需要版本控制和变更管理
- 重点关注数据一致性和安全性`
	case "EIF":
		functionTypeGuidance = `
【EIF功能点特点】
- 外部接口文件：与外部系统交互
- 包含接口协议和数据格式转换
- 通常涉及API调用和响应处理
- 需要错误重试和熔断机制
- 重点关注接口稳定性和性能`
	default:
		functionTypeGuidance = `
【通用功能点特点】
- 包含完整的业务流程处理
- 需要数据验证和错误处理
- 涉及用户交互和系统响应
- 重点关注业务逻辑完整性`
	}

	return fmt.Sprintf(`
请为以下四级功能点生成专业的Mermaid流程图，该流程图将用于NESMA功能点分析：

【功能点详细信息】
标题: %s
描述: %s
功能类型: %s (NESMA标准)
复杂度: %s
业务价值: %s
项目领域: %s
验收标准: %s
优先级: %d

【相关知识库内容】
%s

%s

【流程图设计要求】
1. 使用Mermaid %s语法，确保语法正确性
2. 节点数量控制在6-10个之间，避免过于复杂
3. 清晰展示数据流向：输入 → 处理 → 输出
4. 包含关键的业务决策点和异常处理分支
5. 体现功能类型的特点和要求
6. 使用标准化的节点形状和连接线样式
7. 确保流程逻辑符合实际业务场景

【节点设计规范】
- 开始/结束节点：使用圆形 (circle)
- 处理节点：使用矩形 (rectangle)
- 决策节点：使用菱形 (rhombus)
- 数据节点：使用圆柱形 (cylinder)
- 异常处理：使用六边形 (hexagon)

【连接线设计规范】
- 正常流程：实线箭头
- 异常流程：虚线箭头
- 条件分支：带标签的箭头
- 并行处理：粗线箭头

【输出格式要求】
请严格按照以下JSON格式返回结果：
{
  "diagram_type": "%s",
  "title": "功能流程图标题",
  "description": "详细的功能流程描述",
  "mermaid_code": "完整的Mermaid代码，确保语法正确",
  "nodes": [
    {
      "id": "节点唯一标识",
      "label": "节点显示文本",
      "type": "start/end/process/decision/data/exception",
      "shape": "circle/rectangle/rhombus/cylinder/hexagon",
      "category": "input/process/output/decision/exception/validation",
      "properties": {
        "business_rule": "相关业务规则",
        "data_validation": "数据验证要求",
        "error_handling": "异常处理方式"
      }
    }
  ],
  "edges": [
    {
      "id": "边的唯一标识",
      "from": "起始节点ID",
      "to": "目标节点ID",
      "label": "流程标签或条件",
      "type": "arrow/dotted/thick",
      "condition": "决策条件描述"
    }
  ],
  "styling": {
    "theme": "default",
    "node_styles": {
      "start": "fill:#e1f5fe,stroke:#01579b",
      "process": "fill:#f3e5f5,stroke:#4a148c",
      "decision": "fill:#fff3e0,stroke:#e65100",
      "exception": "fill:#ffebee,stroke:#c62828",
      "end": "fill:#e8f5e8,stroke:#2e7d32"
    },
    "edge_styles": {
      "normal": "stroke:#333,stroke-width:2",
      "exception": "stroke:#f44336,stroke-width:2,stroke-dasharray:5,5"
    },
    "color_scheme": {
      "primary": "#1976d2",
      "success": "#4caf50",
      "warning": "#ff9800",
      "error": "#f44336"
    }
  },
  "metadata": {
    "complexity_level": "simple/medium/complex",
    "node_count": 节点总数,
    "edge_count": 边总数,
    "max_depth": 最大流程深度,
    "nesma_function_type": "%s",
    "business_scenario": "业务场景描述",
    "data_flow": "数据流向说明",
    "version": "1.0"
  }
}

【Mermaid语法示例】
graph TD
    A[开始] --> B{数据验证}
    B -->|验证通过| C[业务处理]
    B -->|验证失败| D[错误处理]
    C --> E[数据保存]
    E --> F[结果返回]
    D --> G[错误响应]
    F --> H[结束]
    G --> H

【质量标准】
- 流程逻辑完整且符合业务实际
- 节点命名规范，使用中文描述
- 包含必要的异常处理和错误分支
- 体现功能类型的特点和要求
- 数据流向清晰，便于理解
- 符合NESMA功能点分析标准
- 易于维护和扩展`,
		requirement.Title,
		requirement.Description,
		requirement.FunctionType,
		requirement.Complexity,
		requirement.BusinessValue,
		cycle.Project.Domain,
		requirement.AcceptanceCriteria,
		requirement.Priority,
		knowledgeContext.String(),
		functionTypeGuidance,
		diagramType,
		diagramType,
		requirement.FunctionType,
	)
}

// parseMermaidResponse 解析Mermaid响应
func (s *MermaidGeneratorService) parseMermaidResponse(response string, diagramType string) (*MermaidDiagram, error) {
	var diagram MermaidDiagram

	// 尝试直接解析JSON
	if err := json.Unmarshal([]byte(response), &diagram); err != nil {
		// 如果解析失败，尝试提取JSON部分
		jsonStart := strings.Index(response, "{")
		jsonEnd := strings.LastIndex(response, "}")
		if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
			jsonContent := response[jsonStart : jsonEnd+1]
			if err := json.Unmarshal([]byte(jsonContent), &diagram); err != nil {
				global.GVA_LOG.Warn("解析Mermaid响应失败", zap.String("response", response), zap.Error(err))
				return s.createFallbackMermaidDiagram(response, diagramType), nil
			}
		} else {
			global.GVA_LOG.Warn("无法从响应中提取JSON", zap.String("response", response))
			return s.createFallbackMermaidDiagram(response, diagramType), nil
		}
	}

	// 设置默认值
	if diagram.DiagramType == "" {
		diagram.DiagramType = diagramType
	}
	if diagram.Metadata.Version == "" {
		diagram.Metadata.Version = "1.0"
	}
	diagram.Metadata.GeneratedAt = time.Now()

	return &diagram, nil
}

// createFallbackMermaidDiagram 创建备用Mermaid流程图
func (s *MermaidGeneratorService) createFallbackMermaidDiagram(response string, diagramType string) *MermaidDiagram {
	// 尝试从响应中提取Mermaid代码
	mermaidCode := s.extractMermaidCode(response)
	if mermaidCode == "" {
		mermaidCode = s.generateSimpleMermaidCode()
	}

	return &MermaidDiagram{
		DiagramType: diagramType,
		Title:       "功能流程图",
		Description: "基础功能流程图",
		MermaidCode: mermaidCode,
		Nodes: []MermaidNode{
			{ID: "A", Label: "开始", Type: "start", Shape: "circle", Category: "input"},
			{ID: "B", Label: "处理", Type: "process", Shape: "rectangle", Category: "process"},
			{ID: "C", Label: "结束", Type: "end", Shape: "circle", Category: "output"},
		},
		Edges: []MermaidEdge{
			{ID: "AB", From: "A", To: "B", Label: "", Type: "arrow"},
			{ID: "BC", From: "B", To: "C", Label: "", Type: "arrow"},
		},
		Styling: MermaidStyling{
			Theme: "default",
			NodeStyles: map[string]string{
				"A": "fill:#e1f5fe",
				"B": "fill:#f3e5f5",
				"C": "fill:#e8f5e8",
			},
			ColorScheme: map[string]string{
				"primary": "#007bff",
			},
		},
		Metadata: MermaidMetadata{
			ComplexityLevel: "simple",
			NodeCount:       3,
			EdgeCount:       2,
			MaxDepth:        3,
			GeneratedAt:     time.Now(),
			Version:         "1.0",
		},
	}
}

// extractMermaidCode 从响应中提取Mermaid代码
func (s *MermaidGeneratorService) extractMermaidCode(response string) string {
	// 使用正则表达式提取Mermaid代码块
	mermaidRegex := regexp.MustCompile(`(?s)` + "`" + `mermaid\n(.*?)\n` + "`")
	matches := mermaidRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// 尝试提取graph TD开头的代码
	graphRegex := regexp.MustCompile(`(?s)graph TD\n(.*?)(?:\n\n|\z)`)
	matches = graphRegex.FindStringSubmatch(response)
	if len(matches) > 1 {
		return fmt.Sprintf("graph TD\n%s", strings.TrimSpace(matches[1]))
	}

	return ""
}

// generateSimpleMermaidCode 生成简单的Mermaid代码
func (s *MermaidGeneratorService) generateSimpleMermaidCode() string {
	return `graph TD
    A[开始] --> B[处理]
    B --> C[结束]
    
    classDef startNode fill:#e1f5fe
    classDef processNode fill:#f3e5f5
    classDef endNode fill:#e8f5e8
    
    class A startNode
    class B processNode
    class C endNode`
}

// validateMermaidDiagram 验证Mermaid流程图
func (s *MermaidGeneratorService) validateMermaidDiagram(diagram *MermaidDiagram) MermaidValidationResult {
	var issues []string
	var warnings []string
	var suggestions []string

	// 验证基本结构
	syntaxValid := s.validateMermaidSyntax(diagram.MermaidCode)
	structureValid := s.validateMermaidStructure(diagram)
	complexityOK := s.validateMermaidComplexity(diagram)

	// 语法验证
	if !syntaxValid {
		issues = append(issues, "Mermaid语法不正确")
	}

	// 结构验证
	if !structureValid {
		issues = append(issues, "流程图结构不完整")
	}

	// 复杂度验证
	if !complexityOK {
		warnings = append(warnings, "流程图过于复杂，建议简化")
	}

	// 节点验证
	if len(diagram.Nodes) == 0 {
		issues = append(issues, "流程图缺少节点")
	}

	if len(diagram.Nodes) > s.maxNodes {
		warnings = append(warnings, fmt.Sprintf("节点数量超过推荐值(%d)", s.maxNodes))
	}

	// 边验证
	if len(diagram.Edges) == 0 {
		issues = append(issues, "流程图缺少连接")
	}

	// 生成建议
	if len(diagram.Nodes) > 0 && len(diagram.Edges) > 0 {
		suggestions = append(suggestions, "建议添加更多的节点标签以提高可读性")
		suggestions = append(suggestions, "建议使用不同的颜色来区分不同类型的节点")
	}

	return MermaidValidationResult{
		IsValid:        len(issues) == 0,
		SyntaxValid:    syntaxValid,
		StructureValid: structureValid,
		ComplexityOK:   complexityOK,
		Issues:         issues,
		Warnings:       warnings,
		Suggestions:    suggestions,
	}
}

// validateMermaidSyntax 验证Mermaid语法
func (s *MermaidGeneratorService) validateMermaidSyntax(mermaidCode string) bool {
	// 基本语法检查
	if mermaidCode == "" {
		return false
	}

	// 检查是否包含graph声明
	if !strings.Contains(mermaidCode, "graph") {
		return false
	}

	// 检查基本的箭头语法
	arrowRegex := regexp.MustCompile(`-->|->`)
	if !arrowRegex.MatchString(mermaidCode) {
		return false
	}

	return true
}

// validateMermaidStructure 验证Mermaid结构
func (s *MermaidGeneratorService) validateMermaidStructure(diagram *MermaidDiagram) bool {
	// 检查是否有开始和结束节点
	hasStart := false
	hasEnd := false

	for _, node := range diagram.Nodes {
		if node.Type == "start" {
			hasStart = true
		}
		if node.Type == "end" {
			hasEnd = true
		}
	}

	return hasStart && hasEnd
}

// validateMermaidComplexity 验证Mermaid复杂度
func (s *MermaidGeneratorService) validateMermaidComplexity(diagram *MermaidDiagram) bool {
	mermaidConfig := global.GVA_CONFIG.AI.Mermaid
	
	// 检查节点数量
	if len(diagram.Nodes) > mermaidConfig.MaxNodes {
		return false
	}

	// 检查深度
	if diagram.Metadata.MaxDepth > mermaidConfig.MaxDepth {
		return false
	}

	// 检查边数
	if len(diagram.Edges) > mermaidConfig.MaxEdges {
		return false
	}

	return true
}

// calculateMermaidConfidence 计算Mermaid生成置信度
func (s *MermaidGeneratorService) calculateMermaidConfidence(diagram *MermaidDiagram, validation MermaidValidationResult, knowledgeRefs []KnowledgeReference) float64 {
	var score float64 = 0.5
	mermaidConfig := global.GVA_CONFIG.AI.Mermaid

	// 验证结果影响置信度
	if validation.IsValid {
		score += 0.2
	}
	if validation.SyntaxValid {
		score += 0.1
	}
	if validation.StructureValid {
		score += 0.1
	}
	if validation.ComplexityOK {
		score += 0.05
	}

	// 内容完整性影响置信度
	if len(diagram.Nodes) > 0 {
		score += 0.05
	}
	if len(diagram.Edges) > 0 {
		score += 0.05
	}

	// 知识库支持度
	knowledgeBonus := float64(len(knowledgeRefs)) * 0.03
	if knowledgeBonus > 0.1 {
		knowledgeBonus = 0.1
	}

	finalScore := score + knowledgeBonus
	
	// 检查是否达到最小置信度要求
	if finalScore < mermaidConfig.MinConfidence {
		global.GVA_LOG.Warn("Mermaid生成置信度过低", 
			zap.Float64("confidence", finalScore),
			zap.Float64("minConfidence", mermaidConfig.MinConfidence))
	}

	return finalScore
}

// generateMermaidNotes 生成Mermaid备注
func (s *MermaidGeneratorService) generateMermaidNotes(requirement *nesma.NesmaRequirement, diagram *MermaidDiagram) string {
	var notes strings.Builder
	
	notes.WriteString(fmt.Sprintf("为「%s」生成%s流程图；", requirement.Title, diagram.DiagramType))
	notes.WriteString(fmt.Sprintf("包含%d个节点、%d条边；", len(diagram.Nodes), len(diagram.Edges)))
	notes.WriteString(fmt.Sprintf("复杂度：%s；", diagram.Metadata.ComplexityLevel))
	
	return notes.String()
}

// ApplyMermaidDiagram 应用Mermaid流程图
func (s *MermaidGeneratorService) ApplyMermaidDiagram(requirementID uint, mermaidDiagram *MermaidDiagram) error {
	var requirement nesma.NesmaRequirement
	if err := s.db.First(&requirement, requirementID).Error; err != nil {
		return fmt.Errorf("获取功能点失败: %w", err)
	}

	// 将Mermaid流程图保存到需求中
	_, _ = json.Marshal(mermaidDiagram) // 序列化但不使用，保留以备后用
	
	updates := map[string]interface{}{
		"ai_description": fmt.Sprintf("%s\n\n流程图：\n%s", requirement.AIDescription, mermaidDiagram.MermaidCode),
		"ai_analysis_status": "mermaid_generated",
	}

	now := time.Now()
	updates["ai_analysis_time"] = &now

	return s.db.Model(&requirement).Updates(updates).Error
}

// GetMermaidGeneratorService 获取Mermaid生成服务实例
func GetMermaidGeneratorService() *MermaidGeneratorService {
	return NewMermaidGeneratorService()
}