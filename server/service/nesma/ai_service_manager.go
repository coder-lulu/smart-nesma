package nesma

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// EnhancedAIServiceManager 增强的AI服务管理器
type EnhancedAIServiceManager struct {
	services map[string]AIService
	mu       sync.RWMutex
	
	// 模型策略配置
	defaultModel      string
	reasoningModel    string
	complexAnalysisModel string
}

// AIModelType AI模型类型
type AIModelType string

const (
	ModelTypeDefault          AIModelType = "default"          // 普通对话和简单分析
	ModelTypeReasoning        AIModelType = "reasoning"        // 深度推理和复杂分析
	ModelTypeComplexAnalysis  AIModelType = "complex"          // 复杂的NESMA分析
)

// AITaskComplexity 任务复杂度
type AITaskComplexity string

const (
	ComplexityLow    AITaskComplexity = "low"     // 简单任务
	ComplexityMedium AITaskComplexity = "medium"  // 中等复杂度
	ComplexityHigh   AITaskComplexity = "high"    // 高复杂度
)

// EnhancedAIAnalysisRequest 增强的AI分析请求
type EnhancedAIAnalysisRequest struct {
	Prompt      string            `json:"prompt"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Complexity  AITaskComplexity  `json:"complexity"`
	ModelType   AIModelType       `json:"model_type,omitempty"`
	MaxTokens   int              `json:"max_tokens,omitempty"`
	Temperature float64          `json:"temperature,omitempty"`
	TopP        float64          `json:"top_p,omitempty"`
	Timeout     time.Duration    `json:"timeout,omitempty"`
}

// AIAnalysisResponse AI分析响应
type AIAnalysisResponse struct {
	Content       string                 `json:"content"`
	Model         string                 `json:"model"`
	TokensUsed    int                   `json:"tokens_used"`
	ProcessingTime time.Duration        `json:"processing_time"`
	Confidence    float64               `json:"confidence,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

// GetAIServiceManager 获取AI服务管理器单例
var enhancedAIServiceManagerInstance *EnhancedAIServiceManager
var enhancedAIServiceManagerOnce sync.Once

func GetAIServiceManager() *EnhancedAIServiceManager {
	enhancedAIServiceManagerOnce.Do(func() {
		enhancedAIServiceManagerInstance = NewEnhancedAIServiceManager()
	})
	return enhancedAIServiceManagerInstance
}

// NewEnhancedAIServiceManager 创建增强的AI服务管理器
func NewEnhancedAIServiceManager() *EnhancedAIServiceManager {
	manager := &EnhancedAIServiceManager{
		services:             make(map[string]AIService),
		defaultModel:         "deepseek-chat",
		reasoningModel:       "deepseek-reasoner",
		complexAnalysisModel: "deepseek-reasoner",
	}
	
	// 初始化服务
	manager.initializeServices()
	
	return manager
}

// initializeServices 初始化AI服务
func (m *EnhancedAIServiceManager) initializeServices() {
	// DeepSeek Chat服务
	if global.GVA_CONFIG.AI.DeepSeek.APIKey != "" {
		deepseekService := NewDeepSeekService(
			global.GVA_CONFIG.AI.DeepSeek.APIKey,
			global.GVA_CONFIG.AI.DeepSeek.BaseURL,
			global.GVA_CONFIG.AI.DeepSeek.Model,
		)
		m.registerService("deepseek-chat", deepseekService)
		global.GVA_LOG.Info("DeepSeek Chat服务已注册", zap.String("model", global.GVA_CONFIG.AI.DeepSeek.Model))
	}
	
	// DeepSeek Reasoner服务
	if global.GVA_CONFIG.AI.DeepSeekReasoner.APIKey != "" {
		reasonerService := NewDeepSeekService(
			global.GVA_CONFIG.AI.DeepSeekReasoner.APIKey,
			global.GVA_CONFIG.AI.DeepSeekReasoner.BaseURL,
			global.GVA_CONFIG.AI.DeepSeekReasoner.Model,
		)
		m.registerService("deepseek-reasoner", reasonerService)
		global.GVA_LOG.Info("DeepSeek Reasoner服务已注册", zap.String("model", global.GVA_CONFIG.AI.DeepSeekReasoner.Model))
	}
	
	// OpenAI备用服务
	if global.GVA_CONFIG.AI.OpenAI.APIKey != "" {
		openaiService := NewOpenAIService(
			global.GVA_CONFIG.AI.OpenAI.APIKey,
			global.GVA_CONFIG.AI.OpenAI.BaseURL,
			global.GVA_CONFIG.AI.OpenAI.Model,
		)
		m.registerService("openai", openaiService)
		global.GVA_LOG.Info("OpenAI备用服务已注册", zap.String("model", global.GVA_CONFIG.AI.OpenAI.Model))
	}
}

// registerService 注册AI服务
func (m *EnhancedAIServiceManager) registerService(name string, service AIService) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.services[name] = service
}

// getService 获取AI服务
func (m *EnhancedAIServiceManager) getService(name string) (AIService, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	service, exists := m.services[name]
	return service, exists
}

// selectOptimalModel 根据任务复杂度选择最优模型
func (m *EnhancedAIServiceManager) selectOptimalModel(complexity AITaskComplexity, modelType AIModelType) string {
	// 优先使用指定的模型类型
	if modelType == ModelTypeReasoning || modelType == ModelTypeComplexAnalysis {
		if _, exists := m.getService("deepseek-reasoner"); exists {
			return "deepseek-reasoner"
		}
	}
	
	// 根据复杂度选择模型
	switch complexity {
	case ComplexityHigh:
		// 高复杂度任务优先使用reasoner模型
		if _, exists := m.getService("deepseek-reasoner"); exists {
			return "deepseek-reasoner"
		}
		return "deepseek-chat"
	case ComplexityMedium:
		// 中等复杂度任务根据任务类型选择
		if modelType == ModelTypeReasoning {
			if _, exists := m.getService("deepseek-reasoner"); exists {
				return "deepseek-reasoner"
			}
		}
		return "deepseek-chat"
	case ComplexityLow:
		// 低复杂度任务使用普通模型
		return "deepseek-chat"
	default:
		return "deepseek-chat"
	}
}

// AnalyzeWithOptimalModel 使用最优模型进行分析
func (m *EnhancedAIServiceManager) AnalyzeWithOptimalModel(ctx context.Context, req *EnhancedAIAnalysisRequest) (*AIAnalysisResponse, error) {
	startTime := time.Now()
	
	// 获取性能监控器
	monitor := GetAIPerformanceMonitor()
	
	// 根据历史性能数据选择最优模型
	recommendedModel := monitor.GetModelRecommendation(req.Complexity)
	selectedModel := m.selectOptimalModelWithHistory(req.Complexity, req.ModelType, recommendedModel)
	
	// 获取服务
	service, exists := m.getService(selectedModel)
	if !exists {
		// 降级到默认服务
		if service, exists = m.getService("deepseek-chat"); !exists {
			return nil, fmt.Errorf("没有可用的AI服务")
		}
		selectedModel = "deepseek-chat"
	}
	
	// 构建AI配置
	config := m.buildAIConfig(req, selectedModel)
	
	// 执行分析
	global.GVA_LOG.Info("开始AI分析", 
		zap.String("model", selectedModel),
		zap.String("recommended_model", recommendedModel),
		zap.String("complexity", string(req.Complexity)),
		zap.Int("prompt_length", len(req.Prompt)))
	
	response, err := service.GenerateText(ctx, req.Prompt, config)
	
	processingTime := time.Since(startTime)
	success := err == nil
	
	// 记录性能数据
	requestRecord := AIRequestRecord{
		ID:           fmt.Sprintf("%d", startTime.UnixNano()),
		ModelName:    selectedModel,
		RequestTime:  startTime,
		ResponseTime: processingTime,
		Success:      success,
		RequestType:  "nesma_analysis",
		Complexity:   req.Complexity,
		Context:      req.Context,
	}
	
	if err != nil {
		requestRecord.ErrorMessage = err.Error()
		monitor.RecordRequest(requestRecord)
		
		global.GVA_LOG.Error("AI分析失败", 
			zap.String("model", selectedModel),
			zap.Error(err))
		return nil, fmt.Errorf("AI分析失败: %v", err)
	}
	
	// 补充记录信息
	requestRecord.TokensUsed = response.Usage.TotalTokens
	requestRecord.Confidence = m.calculateConfidence(response, selectedModel)
	
	// 记录成功的请求
	monitor.RecordRequest(requestRecord)
	
	// 构建响应
	content := response.Text
	if content == "" && len(response.Choices) > 0 {
		content = response.Choices[0].Message.Content
	}
	
	result := &AIAnalysisResponse{
		Content:        content,
		Model:          selectedModel,
		TokensUsed:     response.Usage.TotalTokens,
		ProcessingTime: processingTime,
		Confidence:     requestRecord.Confidence,
		Metadata: map[string]interface{}{
			"complexity":       string(req.Complexity),
			"model_type":      string(req.ModelType),
			"processing_time": processingTime.Seconds(),
			"input_tokens":    response.Usage.PromptTokens,
			"output_tokens":   response.Usage.CompletionTokens,
			"recommended_model": recommendedModel,
		},
	}
	
	global.GVA_LOG.Info("AI分析完成", 
		zap.String("model", selectedModel),
		zap.Duration("processing_time", processingTime),
		zap.Int("tokens_used", response.Usage.TotalTokens),
		zap.Float64("confidence", result.Confidence))
	
	return result, nil
}

// buildAIConfig 构建AI配置
func (m *EnhancedAIServiceManager) buildAIConfig(req *EnhancedAIAnalysisRequest, model string) *AIConfig {
	config := &AIConfig{}
	
	// 根据模型设置默认参数
	switch model {
	case "deepseek-reasoner":
		config.MaxTokens = 32768
		config.Temperature = 0.3
		config.TopP = 0.95
	case "deepseek-chat":
		config.MaxTokens = 8192
		config.Temperature = 0.7
		config.TopP = 0.95
	default:
		config.MaxTokens = 8192
		config.Temperature = 0.7
		config.TopP = 0.95
	}
	
	// 根据复杂度调整参数
	switch req.Complexity {
	case ComplexityHigh:
		config.Temperature = 0.3  // 更保守的温度
		if model == "deepseek-reasoner" {
			config.MaxTokens = 32768
		} else {
			config.MaxTokens = 8192
		}
	case ComplexityMedium:
		config.Temperature = 0.5
		config.MaxTokens = 8192
	case ComplexityLow:
		config.Temperature = 0.7
		config.MaxTokens = 4096
	}
	
	// 应用请求中的覆盖参数
	if req.MaxTokens > 0 {
		config.MaxTokens = req.MaxTokens
	}
	if req.Temperature >= 0 {
		config.Temperature = req.Temperature
	}
	if req.TopP > 0 {
		config.TopP = req.TopP
	}
	
	return config
}

// calculateConfidence 计算置信度
func (m *EnhancedAIServiceManager) calculateConfidence(response *AIResponse, model string) float64 {
	confidence := 0.8 // 基础置信度
	
	// 根据模型调整置信度
	switch model {
	case "deepseek-reasoner":
		confidence = 0.95  // Reasoner模型置信度更高
	case "deepseek-chat":
		confidence = 0.85
	default:
		confidence = 0.8
	}
	
	// 根据响应长度调整（更长的响应通常更详细）
	content := response.Text
	if content == "" && len(response.Choices) > 0 {
		content = response.Choices[0].Message.Content
	}
	if len(content) > 1000 {
		confidence += 0.05
	} else if len(content) < 200 {
		confidence -= 0.1
	}
	
	// 确保置信度在合理范围内
	if confidence > 1.0 {
		confidence = 1.0
	} else if confidence < 0.1 {
		confidence = 0.1
	}
	
	return confidence
}

// GetAvailableModels 获取可用模型列表
func (m *EnhancedAIServiceManager) GetAvailableModels() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	models := make([]string, 0, len(m.services))
	for name := range m.services {
		models = append(models, name)
	}
	return models
}

// CheckServiceHealth 检查服务健康状态
func (m *EnhancedAIServiceManager) CheckServiceHealth(ctx context.Context) map[string]bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	health := make(map[string]bool)
	for name, service := range m.services {
		health[name] = service.IsHealthy(ctx)
	}
	return health
}

// AnalyzeNESMARequirement 专门用于NESMA需求分析的高级接口
func (m *EnhancedAIServiceManager) AnalyzeNESMARequirement(ctx context.Context, prompt string, requirementContext map[string]interface{}) (*AIAnalysisResponse, error) {
	req := &EnhancedAIAnalysisRequest{
		Prompt:     prompt,
		Context:    requirementContext,
		Complexity: ComplexityHigh,
		ModelType:  ModelTypeComplexAnalysis,
		Timeout:    300 * time.Second, // 5分钟超时
	}
	
	return m.AnalyzeWithOptimalModel(ctx, req)
}

// GenerateRequirementDescription 生成需求描述的专用接口
func (m *EnhancedAIServiceManager) GenerateRequirementDescription(ctx context.Context, prompt string) (*AIAnalysisResponse, error) {
	req := &EnhancedAIAnalysisRequest{
		Prompt:     prompt,
		Complexity: ComplexityMedium,
		ModelType:  ModelTypeDefault,
		Timeout:    180 * time.Second, // 3分钟超时
	}
	
	return m.AnalyzeWithOptimalModel(ctx, req)
}

// selectOptimalModelWithHistory 根据历史性能数据选择最优模型
func (m *EnhancedAIServiceManager) selectOptimalModelWithHistory(complexity AITaskComplexity, modelType AIModelType, recommendedModel string) string {
	// 优先使用性能监控推荐的模型
	if recommendedModel != "" && recommendedModel != "deepseek-chat" {
		if _, exists := m.getService(recommendedModel); exists {
			return recommendedModel
		}
	}
	
	// 降级到原有的选择逻辑
	return m.selectOptimalModel(complexity, modelType)
}

// GetPerformanceReport 获取AI性能报告
func (m *EnhancedAIServiceManager) GetPerformanceReport() map[string]interface{} {
	monitor := GetAIPerformanceMonitor()
	return monitor.GeneratePerformanceReport()
}

// GetModelMetrics 获取模型性能指标
func (m *EnhancedAIServiceManager) GetModelMetrics(modelName string) *AIPerformanceMetrics {
	monitor := GetAIPerformanceMonitor()
	return monitor.GetModelMetrics(modelName)
}