package nesma

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// AIService AI服务接口
type AIService interface {
	GenerateText(ctx context.Context, prompt string, config *AIConfig) (*AIResponse, error)
	ChatCompletion(ctx context.Context, messages []APIMessage, config *AIConfig) (*AIResponse, error)
	IsHealthy(ctx context.Context) bool
}

// AIServiceWithCircuitBreaker 带断路器的AI服务
type AIServiceWithCircuitBreaker struct {
	service AIService
	breaker *CircuitBreaker
	name    string
}

// FallbackAIService 故障转移AI服务
type FallbackAIService struct {
	primary  AIService
	fallback AIService
	name     string
}

// NewAIServiceWithCircuitBreaker 创建带断路器的AI服务
func NewAIServiceWithCircuitBreaker(service AIService, name string) *AIServiceWithCircuitBreaker {
	config := &CircuitBreakerConfig{
		FailureThreshold:   5,
		SuccessThreshold:   3,
		Timeout:           900 * time.Second,
		ResetTimeout:      900 * time.Second,
		MaxConcurrentCalls: 50,
	}
	breaker := NewCircuitBreaker("ai_service_"+name, config)
	
	return &AIServiceWithCircuitBreaker{
		service: service,
		breaker: breaker,
		name:    name,
	}
}

// NewFallbackAIService 创建故障转移AI服务
func NewFallbackAIService(primary, fallback AIService, name string) *FallbackAIService {
	return &FallbackAIService{
		primary:  primary,
		fallback: fallback,
		name:     name,
	}
}

// GenerateText 实现AIService接口 - 带断路器的版本
func (a *AIServiceWithCircuitBreaker) GenerateText(ctx context.Context, prompt string, config *AIConfig) (*AIResponse, error) {
	var result *AIResponse
	err := a.breaker.Execute(func() error {
		var err error
		result, err = a.service.GenerateText(ctx, prompt, config)
		return err
	})
	return result, err
}

// ChatCompletion 实现AIService接口 - 带断路器的版本
func (a *AIServiceWithCircuitBreaker) ChatCompletion(ctx context.Context, messages []APIMessage, config *AIConfig) (*AIResponse, error) {
	var result *AIResponse
	err := a.breaker.Execute(func() error {
		var err error
		result, err = a.service.ChatCompletion(ctx, messages, config)
		return err
	})
	return result, err
}

// IsHealthy 实现AIService接口 - 带断路器的版本
func (a *AIServiceWithCircuitBreaker) IsHealthy(ctx context.Context) bool {
	return a.service.IsHealthy(ctx)
}

// GenerateText 实现AIService接口 - 故障转移版本
func (f *FallbackAIService) GenerateText(ctx context.Context, prompt string, config *AIConfig) (*AIResponse, error) {
	result, err := f.primary.GenerateText(ctx, prompt, config)
	if err != nil {
		global.GVA_LOG.Warn("主服务失败，尝试故障转移", zap.String("service", f.name), zap.Error(err))
		return f.fallback.GenerateText(ctx, prompt, config)
	}
	return result, err
}

// ChatCompletion 实现AIService接口 - 故障转移版本
func (f *FallbackAIService) ChatCompletion(ctx context.Context, messages []APIMessage, config *AIConfig) (*AIResponse, error) {
	result, err := f.primary.ChatCompletion(ctx, messages, config)
	if err != nil {
		global.GVA_LOG.Warn("主服务失败，尝试故障转移", zap.String("service", f.name), zap.Error(err))
		return f.fallback.ChatCompletion(ctx, messages, config)
	}
	return result, err
}

// IsHealthy 实现AIService接口 - 故障转移版本
func (f *FallbackAIService) IsHealthy(ctx context.Context) bool {
	return f.primary.IsHealthy(ctx) || f.fallback.IsHealthy(ctx)
}

// GetStats 获取断路器统计信息
func (a *AIServiceWithCircuitBreaker) GetStats() *CircuitBreakerMetrics {
	return a.breaker.GetMetrics()
}

// ResetCircuitBreaker 重置断路器
func (a *AIServiceWithCircuitBreaker) ResetCircuitBreaker() {
	a.breaker.Reset()
}

// StreamAIService 流式AI服务接口
type StreamAIService interface {
	AIService
	ChatCompletionStream(ctx context.Context, messages []APIMessage, config *AIConfig, callback func(chunk string, isComplete bool, tokens int)) error
}

// AIConfig AI配置
type AIConfig struct {
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	Model       string  `json:"model"`
	Stream      bool    `json:"stream"`
}

// APIMessage API消息格式（用于AI接口调用）
type APIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIResponse AI响应
type AIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
	Error   *AIError `json:"error,omitempty"`
	Text    string   `json:"text,omitempty"` // 添加Text字段
}

// Choice 选择项
type Choice struct {
	Index        int        `json:"index"`
	Message      APIMessage `json:"message"`
	Delta        APIMessage `json:"delta"`
	FinishReason string     `json:"finish_reason"`
}

// Usage 使用情况
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// TokenUsage Token使用统计（添加缺失的定义）
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AIError AI错误
type AIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

// DeepSeekService DeepSeek AI服务实现
type DeepSeekService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Model      string
}

// NewDeepSeekService 创建DeepSeek服务实例
func NewDeepSeekService(apiKey, baseURL, model string) *DeepSeekService {
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	if model == "" {
		model = "deepseek-chat"
	}

	return &DeepSeekService{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 300 * time.Second, // 改为5分钟
			Transport: &http.Transport{
				Dial: (&net.Dialer{
					Timeout: 30 * time.Second, // 连接超时
				}).Dial,
				TLSHandshakeTimeout:   30 * time.Second,  // TLS握手超时
				ResponseHeaderTimeout: 60 * time.Second,  // 响应头超时
				ExpectContinueTimeout: 10 * time.Second,  // Expect: 100-continue超时
				IdleConnTimeout:       90 * time.Second,  // 空闲连接超时
				MaxIdleConns:          100,               // 最大空闲连接数
				MaxIdleConnsPerHost:   10,                // 每个主机最大空闲连接数
			},
		},
	}
}

// GenerateText 生成文本
func (ds *DeepSeekService) GenerateText(ctx context.Context, prompt string, config *AIConfig) (*AIResponse, error) {
	messages := []APIMessage{
		{Role: "user", Content: prompt},
	}
	return ds.ChatCompletion(ctx, messages, config)
}

// ChatCompletion 聊天补全
func (ds *DeepSeekService) ChatCompletion(ctx context.Context, messages []APIMessage, config *AIConfig) (*AIResponse, error) {
	if config == nil {
		config = &AIConfig{
			MaxTokens:   8192,
			Temperature: 0.7,
			TopP:        0.95,
			Model:       ds.Model,
			Stream:      false,
		}
	}

	if config.Model == "" {
		config.Model = ds.Model
	}

	requestBody := map[string]interface{}{
		"model":       config.Model,
		"messages":    messages,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"stream":      config.Stream,
	}

	// 只有在TopP值有效时才包含该参数
	if config.TopP > 0 && config.TopP <= 1.0 {
		requestBody["top_p"] = config.TopP
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 添加详细调试日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("=== AI服务请求开始 ===")
		global.GVA_LOG.Info("DeepSeek API请求详情",
			zap.String("url", ds.BaseURL+"/chat/completions"),
			zap.String("model", config.Model),
			zap.String("apiKey", ds.APIKey[:10]+"..."), // 只显示前10个字符
			zap.String("requestBody", string(jsonData)),
		)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ds.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ds.APIKey)

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("请求头信息",
			zap.String("Content-Type", req.Header.Get("Content-Type")),
			zap.String("Authorization", "Bearer "+ds.APIKey[:10]+"..."),
		)
	}

	resp, err := ds.HTTPClient.Do(req)
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("DeepSeek API请求失败", zap.Error(err))
		}
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 添加详细响应调试日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("=== AI服务响应详情 ===",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.Any("headers", resp.Header),
			zap.String("responseBody", string(body)), // 完整响应内容
			zap.Int("bodyLength", len(body)),
		)
		global.GVA_LOG.Info("=== AI服务响应结束 ===")
	}

	if resp.StatusCode != http.StatusOK {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("DeepSeek API返回错误状态",
				zap.Int("status", resp.StatusCode),
				zap.String("response", string(body)),
			)
		}
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// 检查响应是否是SSE格式 - 如果是SSE格式，说明应该使用流式API
	bodyStr := string(body)
	if strings.HasPrefix(bodyStr, "data: ") {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Info("检测到SSE格式响应，自动转换为非流式响应")
		}

		// 内部处理SSE，组装成完整响应
		return ds.convertSSEToResponse(bodyStr)
	}

	var aiResponse AIResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("DeepSeek API响应解析失败",
				zap.Error(err),
				zap.String("response", string(body)),
			)
		}
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	// 设置Text字段
	if len(aiResponse.Choices) > 0 {
		aiResponse.Text = aiResponse.Choices[0].Message.Content
	}

	return &aiResponse, nil
}

// ChatCompletionStream 流式聊天补全
func (ds *DeepSeekService) ChatCompletionStream(ctx context.Context, messages []APIMessage, config *AIConfig, callback func(chunk string, isComplete bool, tokens int)) error {
	if config == nil {
		config = &AIConfig{
			MaxTokens:   8192,
			Temperature: 0.7,
			TopP:        0.95,
			Model:       ds.Model,
			Stream:      true,
		}
	}

	if config.Model == "" {
		config.Model = ds.Model
	}

	// 强制启用流式
	config.Stream = true

	requestBody := map[string]interface{}{
		"model":       config.Model,
		"messages":    messages,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"stream":      true,
	}

	// 只有在TopP值有效时才包含该参数
	if config.TopP > 0 && config.TopP <= 1.0 {
		requestBody["top_p"] = config.TopP
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("DeepSeek流式API请求",
			zap.String("url", ds.BaseURL+"/chat/completions"),
			zap.String("model", config.Model),
		)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", ds.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ds.APIKey)
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	resp, err := ds.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	// 解析SSE流
	scanner := bufio.NewScanner(resp.Body)
	scanner.Split(bufio.ScanLines)

	totalTokens := 0

	for scanner.Scan() {
		line := scanner.Text()

		// 跳过空行
		if strings.TrimSpace(line) == "" {
			continue
		}

		// 处理SSE数据行
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")

			// 检查结束标记
			if strings.TrimSpace(data) == "[DONE]" {
				callback("", true, totalTokens)
				break
			}

			// 解析JSON数据
			var streamResponse struct {
				ID      string `json:"id"`
				Object  string `json:"object"`
				Created int64  `json:"created"`
				Model   string `json:"model"`
				Choices []struct {
					Index int `json:"index"`
					Delta struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"delta"`
					Message struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"message"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(data), &streamResponse); err != nil {
				if global.GVA_LOG != nil {
					global.GVA_LOG.Warn("解析流式响应失败", zap.Error(err), zap.String("data", data))
				}
				continue
			}

			// 提取内容和token信息
			if len(streamResponse.Choices) > 0 {
				choice := streamResponse.Choices[0]
				content := choice.Delta.Content

				if streamResponse.Usage != nil {
					totalTokens = streamResponse.Usage.TotalTokens
				}

				// 检查是否完成
				isComplete := choice.FinishReason != nil && *choice.FinishReason == "stop"

				if content != "" || isComplete {
					callback(content, isComplete, totalTokens)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading stream: %w", err)
	}

	return nil
}

// IsHealthy 检查服务健康状态
func (ds *DeepSeekService) IsHealthy(ctx context.Context) bool {
	messages := []APIMessage{
		{Role: "user", Content: "ping"},
	}
	config := &AIConfig{
		MaxTokens:   10,
		Temperature: 0.1,
		TopP:        0.95,
		Model:       ds.Model,
	}

	_, err := ds.ChatCompletion(ctx, messages, config)
	return err == nil
}

// convertSSEToResponse 将SSE格式响应转换为标准响应
func (ds *DeepSeekService) convertSSEToResponse(sseData string) (*AIResponse, error) {
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("开始转换SSE响应为标准格式")
	}

	var content strings.Builder
	var usage Usage
	var model string
	var messageID string

	// 将SSE数据按行分割，正确处理换行符
	lines := strings.Split(strings.ReplaceAll(sseData, "\r\n", "\n"), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 处理data:开头的行
		if strings.HasPrefix(line, "data: ") {
			dataStr := strings.TrimPrefix(line, "data: ")

			// 跳过[DONE]标记
			if strings.TrimSpace(dataStr) == "[DONE]" {
				continue
			}

			// 解析JSON数据
			var streamResponse struct {
				ID      string `json:"id"`
				Object  string `json:"object"`
				Created int64  `json:"created"`
				Model   string `json:"model"`
				Choices []struct {
					Index int `json:"index"`
					Delta struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"delta"`
					Message struct {
						Role    string `json:"role"`
						Content string `json:"content"`
					} `json:"message"`
					FinishReason *string `json:"finish_reason"`
				} `json:"choices"`
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}

			if err := json.Unmarshal([]byte(dataStr), &streamResponse); err != nil {
				if global.GVA_LOG != nil {
					global.GVA_LOG.Warn("解析SSE数据失败", zap.Error(err), zap.String("data", dataStr))
				}
				continue
			}

			// 提取信息
			if streamResponse.ID != "" {
				messageID = streamResponse.ID
			}
			if streamResponse.Model != "" {
				model = streamResponse.Model
			}

			// 提取内容 - 只累积delta.content
			if len(streamResponse.Choices) > 0 {
				choice := streamResponse.Choices[0]

				// 优先使用delta中的内容（流式数据）
				if choice.Delta.Content != "" {
					content.WriteString(choice.Delta.Content)
				}
			}

			// 提取usage信息（通常在最后一个chunk中）
			if streamResponse.Usage != nil {
				usage = Usage{
					PromptTokens:     streamResponse.Usage.PromptTokens,
					CompletionTokens: streamResponse.Usage.CompletionTokens,
					TotalTokens:      streamResponse.Usage.TotalTokens,
				}
			}
		}
	}

	// 构建响应
	finalContent := content.String()
	aiResponse := &AIResponse{
		ID:      messageID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   model,
		Choices: []Choice{
			{
				Index: 0,
				Message: APIMessage{
					Role:    "assistant",
					Content: finalContent,
				},
				FinishReason: "stop",
			},
		},
		Usage: usage,
		Text:  finalContent,
	}

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("SSE转换完成",
			zap.String("content", finalContent),
			zap.Int("contentLength", len(finalContent)),
			zap.Int("totalTokens", usage.TotalTokens),
		)
	}

	return aiResponse, nil
}

// AIServiceManager AI服务管理器
type AIServiceManager struct {
	services             map[string]AIService
	defaultService       AIService
	circuitBreakers      map[string]*AIServiceWithCircuitBreaker
	fallbackServices     map[string]*FallbackAIService
	enableCircuitBreaker bool
	enableFallback       bool
}

// NewAIServiceManager 创建AI服务管理器
func NewAIServiceManager() *AIServiceManager {
	return &AIServiceManager{
		services:             make(map[string]AIService),
		circuitBreakers:      make(map[string]*AIServiceWithCircuitBreaker),
		fallbackServices:     make(map[string]*FallbackAIService),
		enableCircuitBreaker: true,  // 默认启用熔断器
		enableFallback:       false, // 默认不启用故障转移
	}
}

// RegisterService 注册AI服务
func (m *AIServiceManager) RegisterService(name string, service AIService) {
	m.services[name] = service
}

// SetDefault 设置默认服务
func (m *AIServiceManager) SetDefault(service AIService) {
	m.defaultService = service
}

// GetService 获取AI服务
func (m *AIServiceManager) GetService(name string) (AIService, bool) {
	service, exists := m.services[name]
	return service, exists
}

// GetDefaultService 获取默认AI服务
func (m *AIServiceManager) GetDefaultService() AIService {
	return m.defaultService
}

// GetAllServices 获取所有已注册的AI服务
func (m *AIServiceManager) GetAllServices() map[string]AIService {
	return m.services
}

// ListAvailableServices 列出所有可用的AI服务名称
func (m *AIServiceManager) ListAvailableServices() []string {
	var services []string
	for name := range m.services {
		services = append(services, name)
	}
	return services
}

// SwitchDefaultService 切换默认AI服务
func (m *AIServiceManager) SwitchDefaultService(name string) bool {
	if service, exists := m.services[name]; exists {
		m.defaultService = service
		return true
	}
	return false
}

// GetServiceHealth 获取所有服务的健康状态
func (m *AIServiceManager) GetServiceHealth(ctx context.Context) map[string]bool {
	health := make(map[string]bool)
	for name, service := range m.services {
		health[name] = service.IsHealthy(ctx)
	}
	return health
}

// 全局AI服务管理器
var GlobalAIServiceManager *AIServiceManager

// InitAIServices 初始化AI服务
func InitAIServices() {
	GlobalAIServiceManager = NewAIServiceManager()

	// 从配置文件读取AI服务配置
	aiConfig := global.GVA_CONFIG.AI

	var primaryService AIService
	var fallbackService AIService

	// 初始化DeepSeek服务
	if aiConfig.DeepSeek.APIKey != "" {
		deepseekService := NewDeepSeekService(
			aiConfig.DeepSeek.APIKey,
			aiConfig.DeepSeek.BaseURL,
			aiConfig.DeepSeek.Model,
		)
		GlobalAIServiceManager.RegisterServiceWithCircuitBreaker("deepseek", deepseekService)

		// 如果没有设置默认服务，则设置DeepSeek为默认服务
		if GlobalAIServiceManager.GetDefaultService() == nil {
			GlobalAIServiceManager.SetDefault(deepseekService)
			primaryService = deepseekService
		}
	}

	// 初始化OpenAI服务
	if aiConfig.OpenAI.APIKey != "" {
		openaiService := NewOpenAIService(
			aiConfig.OpenAI.APIKey,
			aiConfig.OpenAI.BaseURL,
			aiConfig.OpenAI.Model,
		)
		GlobalAIServiceManager.RegisterServiceWithCircuitBreaker("openai", openaiService)

		// 如果没有设置默认服务，则设置OpenAI为默认服务
		if GlobalAIServiceManager.GetDefaultService() == nil {
			GlobalAIServiceManager.SetDefault(openaiService)
			primaryService = openaiService
		} else if fallbackService == nil {
			fallbackService = openaiService
		}
	}

	// 初始化Claude服务
	if aiConfig.Claude.APIKey != "" {
		claudeService := NewClaudeService(
			aiConfig.Claude.APIKey,
			aiConfig.Claude.BaseURL,
			aiConfig.Claude.Model,
		)
		GlobalAIServiceManager.RegisterServiceWithCircuitBreaker("claude", claudeService)

		// 如果没有设置默认服务，则设置Claude为默认服务
		if GlobalAIServiceManager.GetDefaultService() == nil {
			GlobalAIServiceManager.SetDefault(claudeService)
			primaryService = claudeService
		} else if fallbackService == nil {
			fallbackService = claudeService
		}
	}

	// 如果有多个服务，设置故障转移
	if primaryService != nil && fallbackService != nil {
		GlobalAIServiceManager.EnableFallback(true)
		// 为主要服务设置故障转移
		for name, service := range GlobalAIServiceManager.services {
			if service == primaryService {
				GlobalAIServiceManager.RegisterFallbackService(name+"_with_fallback", primaryService, fallbackService)
				break
			}
		}
	}

	// 只有在日志对象已经初始化后才输出日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("AI services initialized successfully with circuit breaker and fallback protection")
	}
}

// OpenAIService OpenAI服务实现
type OpenAIService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Model      string
}

// NewOpenAIService 创建OpenAI服务实例
func NewOpenAIService(apiKey, baseURL, model string) *OpenAIService {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-3.5-turbo"
	}

	return &OpenAIService{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second, // 增加到3分钟，适应复杂任务
		},
	}
}

// GenerateText 生成文本
func (os *OpenAIService) GenerateText(ctx context.Context, prompt string, config *AIConfig) (*AIResponse, error) {
	messages := []APIMessage{
		{Role: "user", Content: prompt},
	}
	return os.ChatCompletion(ctx, messages, config)
}

// ChatCompletion 聊天补全
func (os *OpenAIService) ChatCompletion(ctx context.Context, messages []APIMessage, config *AIConfig) (*AIResponse, error) {
	if config == nil {
		config = &AIConfig{
			MaxTokens:   8192,
			Temperature: 0.7,
			TopP:        0.95,
			Model:       os.Model,
			Stream:      false,
		}
	}

	if config.Model == "" {
		config.Model = os.Model
	}

	requestBody := map[string]interface{}{
		"model":       config.Model,
		"messages":    messages,
		"max_tokens":  config.MaxTokens,
		"temperature": config.Temperature,
		"stream":      config.Stream,
	}

	// 只有在TopP值有效时才包含该参数
	if config.TopP > 0 && config.TopP <= 1.0 {
		requestBody["top_p"] = config.TopP
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 添加详细调试日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("=== OpenAI服务请求开始 ===")
		global.GVA_LOG.Info("OpenAI API请求详情",
			zap.String("url", os.BaseURL+"/chat/completions"),
			zap.String("model", config.Model),
			zap.String("apiKey", os.APIKey[:10]+"..."), // 只显示前10个字符
			zap.String("requestBody", string(jsonData)),
		)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", os.BaseURL+"/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+os.APIKey)

	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("OpenAI请求头信息",
			zap.String("Content-Type", req.Header.Get("Content-Type")),
			zap.String("Authorization", "Bearer "+os.APIKey[:10]+"..."),
		)
	}

	resp, err := os.HTTPClient.Do(req)
	if err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("OpenAI API请求失败", zap.Error(err))
		}
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// 添加详细响应调试日志
	if global.GVA_LOG != nil {
		global.GVA_LOG.Info("=== OpenAI服务响应详情 ===",
			zap.Int("statusCode", resp.StatusCode),
			zap.String("status", resp.Status),
			zap.Any("headers", resp.Header),
			zap.String("responseBody", string(body)), // 完整响应内容
			zap.Int("bodyLength", len(body)),
		)
		global.GVA_LOG.Info("=== OpenAI服务响应结束 ===")
	}

	var aiResponse AIResponse
	if err := json.Unmarshal(body, &aiResponse); err != nil {
		if global.GVA_LOG != nil {
			global.GVA_LOG.Error("OpenAI响应JSON解析失败",
				zap.Error(err),
				zap.String("responseBody", string(body)),
			)
		}
		return nil, fmt.Errorf("failed to unmarshal response: %w, body: %s", err, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		if aiResponse.Error != nil {
			return nil, fmt.Errorf("API error: %s", aiResponse.Error.Message)
		}
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	return &aiResponse, nil
}

// IsHealthy 检查服务健康状态
func (os *OpenAIService) IsHealthy(ctx context.Context) bool {
	messages := []APIMessage{
		{Role: "user", Content: "ping"},
	}
	config := &AIConfig{
		MaxTokens:   10,
		Temperature: 0.1,
		TopP:        0.95,
		Model:       os.Model,
	}

	_, err := os.ChatCompletion(ctx, messages, config)
	return err == nil
}

// ClaudeService Claude服务实现
type ClaudeService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Model      string
}

// NewClaudeService 创建Claude服务实例
func NewClaudeService(apiKey, baseURL, model string) *ClaudeService {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if model == "" {
		model = "claude-3-sonnet-20240229"
	}

	return &ClaudeService{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Model:   model,
		HTTPClient: &http.Client{
			Timeout: 180 * time.Second, // 增加到3分钟，适应复杂任务
		},
	}
}

// GenerateText 生成文本
func (cs *ClaudeService) GenerateText(ctx context.Context, prompt string, config *AIConfig) (*AIResponse, error) {
	messages := []APIMessage{
		{Role: "user", Content: prompt},
	}
	return cs.ChatCompletion(ctx, messages, config)
}

// ChatCompletion 聊天补全
func (cs *ClaudeService) ChatCompletion(ctx context.Context, messages []APIMessage, config *AIConfig) (*AIResponse, error) {
	if config == nil {
		config = &AIConfig{
			MaxTokens:   8192,
			Temperature: 0.7,
			TopP:        0.95,
			Model:       cs.Model,
			Stream:      false,
		}
	}

	if config.Model == "" {
		config.Model = cs.Model
	}

	// Claude使用不同的API格式
	requestBody := map[string]interface{}{
		"model":      config.Model,
		"max_tokens": config.MaxTokens,
		"messages":   messages,
	}

	if config.Temperature > 0 {
		requestBody["temperature"] = config.Temperature
	}

	if config.TopP > 0 && config.TopP < 1 {
		requestBody["top_p"] = config.TopP
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", cs.BaseURL+"/v1/messages", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", cs.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := cs.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Claude返回格式与OpenAI不同，需要转换
	var claudeResponse map[string]interface{}
	if err := json.Unmarshal(body, &claudeResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if errorMsg, exists := claudeResponse["error"]; exists {
			return nil, fmt.Errorf("API error: %v", errorMsg)
		}
		return nil, fmt.Errorf("API error: status %d", resp.StatusCode)
	}

	// 转换Claude响应格式为标准格式
	aiResponse := &AIResponse{
		ID:      fmt.Sprintf("claude-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   config.Model,
		Choices: []Choice{},
		Usage:   Usage{},
	}

	// 解析Claude响应内容
	if content, exists := claudeResponse["content"]; exists {
		if contentArray, ok := content.([]interface{}); ok && len(contentArray) > 0 {
			if contentItem, ok := contentArray[0].(map[string]interface{}); ok {
				if text, exists := contentItem["text"]; exists {
					aiResponse.Choices = append(aiResponse.Choices, Choice{
						Index: 0,
						Message: APIMessage{
							Role:    "assistant",
							Content: text.(string),
						},
						FinishReason: "stop",
					})
				}
			}
		}
	}

	// 解析usage信息
	if usage, exists := claudeResponse["usage"]; exists {
		if usageMap, ok := usage.(map[string]interface{}); ok {
			if inputTokens, exists := usageMap["input_tokens"]; exists {
				if tokens, ok := inputTokens.(float64); ok {
					aiResponse.Usage.PromptTokens = int(tokens)
				}
			}
			if outputTokens, exists := usageMap["output_tokens"]; exists {
				if tokens, ok := outputTokens.(float64); ok {
					aiResponse.Usage.CompletionTokens = int(tokens)
				}
			}
			aiResponse.Usage.TotalTokens = aiResponse.Usage.PromptTokens + aiResponse.Usage.CompletionTokens
		}
	}

	return aiResponse, nil
}

// IsHealthy 检查服务健康状态
func (cs *ClaudeService) IsHealthy(ctx context.Context) bool {
	messages := []APIMessage{
		{Role: "user", Content: "ping"},
	}
	config := &AIConfig{
		MaxTokens:   10,
		Temperature: 0.1,
		TopP:        0.95,
		Model:       cs.Model,
	}

	_, err := cs.ChatCompletion(ctx, messages, config)
	return err == nil
}

// GetAIService 获取AI服务
func GetAIService(name ...string) AIService {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}

	if len(name) > 0 {
		if service, exists := GlobalAIServiceManager.GetService(name[0]); exists {
			return service
		}
	}

	return GlobalAIServiceManager.GetDefaultService()
}

// ListAvailableAIServices 列出所有可用的AI服务
func ListAvailableAIServices() []string {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}
	return GlobalAIServiceManager.ListAvailableServices()
}

// SwitchDefaultAIService 切换默认AI服务
func SwitchDefaultAIService(name string) bool {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}
	return GlobalAIServiceManager.SwitchDefaultService(name)
}

// GetAIServiceHealth 获取所有AI服务的健康状态
func GetAIServiceHealth(ctx context.Context) map[string]interface{} {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}

	health := make(map[string]interface{})
	healthMap := GlobalAIServiceManager.GetServiceHealth(ctx)

	// 转换为接口类型并添加详细信息
	for name, isHealthy := range healthMap {
		health[name] = map[string]interface{}{
			"available": isHealthy,
			"service":   name,
		}
	}

	// 计算整体状态
	overallStatus := "healthy"
	healthyCount := 0
	totalCount := len(healthMap)

	for _, isHealthy := range healthMap {
		if isHealthy {
			healthyCount++
		}
	}

	if healthyCount == 0 {
		overallStatus = "unhealthy"
	} else if healthyCount < totalCount {
		overallStatus = "degraded"
	}

	health["status"] = overallStatus
	health["healthy_services"] = healthyCount
	health["total_services"] = totalCount

	return health
}

// EnableCircuitBreaker 启用/禁用熔断器
func (m *AIServiceManager) EnableCircuitBreaker(enable bool) {
	m.enableCircuitBreaker = enable
}

// EnableFallback 启用/禁用故障转移
func (m *AIServiceManager) EnableFallback(enable bool) {
	m.enableFallback = enable
}

// RegisterServiceWithCircuitBreaker 注册带熔断器的服务
func (m *AIServiceManager) RegisterServiceWithCircuitBreaker(name string, service AIService) {
	m.services[name] = service

	// 创建带熔断器的服务
	if m.enableCircuitBreaker {
		cbService := NewAIServiceWithCircuitBreaker(service, name)
		m.circuitBreakers[name] = cbService
	}
}

// RegisterFallbackService 注册故障转移服务
func (m *AIServiceManager) RegisterFallbackService(name string, primary, fallback AIService) {
	// 注册主服务
	m.RegisterServiceWithCircuitBreaker(name, primary)

	// 创建故障转移服务
	if m.enableFallback {
		fallbackService := NewFallbackAIService(primary, fallback, name)
		m.fallbackServices[name] = fallbackService
	}
}

// GetServiceWithProtection 获取带保护机制的AI服务
func (m *AIServiceManager) GetServiceWithProtection(name string) AIService {
	// 模型名称映射
	serviceName := m.mapModelNameToService(name)

	// 优先返回故障转移服务
	if m.enableFallback && m.fallbackServices[serviceName] != nil {
		return m.fallbackServices[serviceName]
	}

	// 其次返回带熔断器的服务
	if m.enableCircuitBreaker && m.circuitBreakers[serviceName] != nil {
		return m.circuitBreakers[serviceName]
	}

	// 最后返回原始服务
	return m.services[serviceName]
}

// mapModelNameToService 将模型名称映射到服务名称
func (m *AIServiceManager) mapModelNameToService(modelName string) string {
	// 模型名称到服务名称的映射
	modelToServiceMap := map[string]string{
		"deepseek-chat":     "deepseek",
		"deepseek-coder":    "deepseek",
		"gpt-3.5-turbo":     "openai",
		"gpt-4":             "openai",
		"gpt-4o":            "openai",
		"gpt-4o-mini":       "openai",
		"o1-mini":           "openai",
		"o1-mini-all":       "openai",
		"o1-preview":        "openai",
		"claude-3-opus":     "claude",
		"claude-3-sonnet":   "claude",
		"claude-3-haiku":    "claude",
		"claude-3-5-sonnet": "claude",
		"claude-3-5-haiku":  "claude",
	}

	if serviceName, exists := modelToServiceMap[modelName]; exists {
		return serviceName
	}

	// 如果没有映射，检查是否直接是服务名称
	if _, exists := m.services[modelName]; exists {
		return modelName
	}

	// 默认返回输入的名称
	return modelName
}

// GetDefaultServiceWithProtection 获取带保护机制的默认AI服务
func (m *AIServiceManager) GetDefaultServiceWithProtection() AIService {
	// 找到默认服务的名称
	var defaultName string
	for name, service := range m.services {
		if service == m.defaultService {
			defaultName = name
			break
		}
	}

	if defaultName != "" {
		return m.GetServiceWithProtection(defaultName)
	}

	return m.defaultService
}

// GetCircuitBreakerStats 获取熔断器统计信息
func (m *AIServiceManager) GetCircuitBreakerStats() map[string]interface{} {
	stats := make(map[string]interface{})

	for name, cb := range m.circuitBreakers {
		stats[name] = cb.GetStats()
	}

	return stats
}

// ResetCircuitBreaker 重置指定服务的熔断器
func (m *AIServiceManager) ResetCircuitBreaker(name string) bool {
	if cb, exists := m.circuitBreakers[name]; exists {
		cb.ResetCircuitBreaker()
		return true
	}
	return false
}

// ResetAllCircuitBreakers 重置所有熔断器
func (m *AIServiceManager) ResetAllCircuitBreakers() {
	for _, cb := range m.circuitBreakers {
		cb.ResetCircuitBreaker()
	}
}

// GetAIServiceWithProtection 获取带保护机制的AI服务（全局函数）
func GetAIServiceWithProtection(name ...string) AIService {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}

	if len(name) > 0 {
		return GlobalAIServiceManager.GetServiceWithProtection(name[0])
	}

	return GlobalAIServiceManager.GetDefaultServiceWithProtection()
}

// GetCircuitBreakerStats 获取熔断器统计信息（全局函数）
func GetCircuitBreakerStats() map[string]interface{} {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}
	return GlobalAIServiceManager.GetCircuitBreakerStats()
}

// ResetCircuitBreaker 重置熔断器（全局函数）
func ResetCircuitBreaker(name string) bool {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}
	return GlobalAIServiceManager.ResetCircuitBreaker(name)
}

// ResetAllCircuitBreakers 重置所有熔断器（全局函数）
func ResetAllCircuitBreakers() {
	if GlobalAIServiceManager == nil {
		InitAIServices()
	}
	GlobalAIServiceManager.ResetAllCircuitBreakers()
}

