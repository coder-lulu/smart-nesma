package nesma

import (
	"context"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
)

type AIServiceApi struct{}

// AIServiceInfo AI服务信息
type AIServiceInfo struct {
	Name        string `json:"name"`
	IsDefault   bool   `json:"isDefault"`
	IsHealthy   bool   `json:"isHealthy"`
	Model       string `json:"model"`
	Description string `json:"description"`
}

// AIServiceListResponse 服务列表响应
type AIServiceListResponse struct {
	Services []AIServiceInfo `json:"services"`
	Total    int             `json:"total"`
}

// AIServiceSwitchRequest 切换服务请求
type AIServiceSwitchRequest struct {
	ServiceName string `json:"serviceName" binding:"required"`
}

// AITestRequest AI测试请求
type AITestRequest struct {
	ServiceName string `json:"serviceName"`
	Prompt      string `json:"prompt" binding:"required"`
}

// AITestResponse AI测试响应
type AITestResponse struct {
	ServiceName string `json:"serviceName"`
	Response    string `json:"response"`
	TokenUsage  int    `json:"tokenUsage"`
	Duration    int64  `json:"duration"`
}

// @Tags AI服务管理
// @Summary 获取AI服务列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=AIServiceListResponse} "获取成功"
// @Router /ai/services [get]
func (a *AIServiceApi) GetAIServices(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 获取所有可用的AI服务
	serviceNames := nesma.ListAvailableAIServices()

	// 获取服务健康状态
	healthStatus := nesma.GetAIServiceHealth(ctx)

	// 获取当前默认服务
	defaultService := nesma.GetAIService()
	var defaultServiceName string
	if defaultService != nil {
		// 通过比较服务实例找到默认服务名称
		for _, name := range serviceNames {
			service := nesma.GetAIService(name)
			if service == defaultService {
				defaultServiceName = name
				break
			}
		}
	}

	// 构建服务信息
	var services []AIServiceInfo
	for _, name := range serviceNames {
		// 类型断言获取健康状态
		isHealthy := false
		if healthData, ok := healthStatus[name]; ok {
			if healthMap, ok := healthData.(map[string]interface{}); ok {
				if available, ok := healthMap["available"].(bool); ok {
					isHealthy = available
				}
			}
		}

		service := AIServiceInfo{
			Name:      name,
			IsDefault: name == defaultServiceName,
			IsHealthy: isHealthy,
		}

		// 添加服务描述和模型信息
		switch name {
		case "deepseek":
			service.Description = "DeepSeek AI - 高性能中文AI模型"
			service.Model = "deepseek-chat"
		case "openai":
			service.Description = "OpenAI GPT - 强大的通用AI模型"
			service.Model = "gpt-3.5-turbo"
		case "claude":
			service.Description = "Claude AI - Anthropic开发的安全AI助手"
			service.Model = "claude-3-sonnet-20240229"
		default:
			service.Description = "AI服务"
			service.Model = "未知"
		}

		services = append(services, service)
	}

	resp := AIServiceListResponse{
		Services: services,
		Total:    len(services),
	}

	response.OkWithData(resp, c)
}

// @Tags AI服务管理
// @Summary 切换默认AI服务
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body AIServiceSwitchRequest true "切换服务请求"
// @Success 200 {object} response.Response{} "切换成功"
// @Router /ai/services/switch [post]
func (a *AIServiceApi) SwitchDefaultAIService(c *gin.Context) {
	var req AIServiceSwitchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 切换默认服务
	if nesma.SwitchDefaultAIService(req.ServiceName) {
		global.GVA_LOG.Info("AI服务已切换: " + req.ServiceName)
		response.OkWithMessage("AI服务已切换为: "+req.ServiceName, c)
	} else {
		response.FailWithMessage("切换失败，服务不存在: "+req.ServiceName, c)
	}
}

// @Tags AI服务管理
// @Summary 获取AI服务健康状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]bool} "获取成功"
// @Router /ai/services/health [get]
func (a *AIServiceApi) GetAIServiceHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	health := nesma.GetAIServiceHealth(ctx)
	response.OkWithData(health, c)
}

// @Tags AI服务管理
// @Summary 测试AI服务
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body AITestRequest true "测试请求"
// @Success 200 {object} response.Response{data=AITestResponse} "测试成功"
// @Router /ai/services/test [post]
func (a *AIServiceApi) TestAIService(c *gin.Context) {
	var req AITestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	start := time.Now()

	// 获取指定的AI服务
	var service nesma.AIService
	if req.ServiceName != "" {
		service = nesma.GetAIService(req.ServiceName)
	} else {
		service = nesma.GetAIService()
		req.ServiceName = "default"
	}

	if service == nil {
		response.FailWithMessage("AI服务不可用", c)
		return
	}

	// 调用AI服务
	config := &nesma.AIConfig{
		MaxTokens:   500,
		Temperature: 0.7,
		TopP:        0.95,
	}

	resp, err := service.GenerateText(ctx, req.Prompt, config)
	if err != nil {
		global.GVA_LOG.Error("AI服务测试失败: " + err.Error())
		response.FailWithMessage("AI服务测试失败: "+err.Error(), c)
		return
	}

	duration := time.Since(start).Milliseconds()

	// 提取响应内容
	var responseText string
	var tokenUsage int
	if len(resp.Choices) > 0 {
		responseText = resp.Choices[0].Message.Content
	}
	tokenUsage = resp.Usage.TotalTokens

	testResp := AITestResponse{
		ServiceName: req.ServiceName,
		Response:    responseText,
		TokenUsage:  tokenUsage,
		Duration:    duration,
	}

	response.OkWithData(testResp, c)
}

// @Tags AI服务管理
// @Summary 获取AI服务配置
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /ai/services/config [get]
func (a *AIServiceApi) GetAIServiceConfig(c *gin.Context) {
	config := map[string]interface{}{
		"available_models": map[string][]string{
			"deepseek": {"deepseek-chat", "deepseek-coder"},
			"openai":   {"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"},
			"claude":   {"claude-3-sonnet-20240229", "claude-3-haiku-20240307", "claude-3-opus-20240229"},
		},
		"default_config": map[string]interface{}{
			"max_tokens":  2000,
			"temperature": 0.7,
			"top_p":       0.95,
		},
	}

	response.OkWithData(config, c)
}

// @Tags AI服务管理
// @Summary 获取熔断器统计信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=map[string]interface{}} "获取成功"
// @Router /ai/services/circuit-breaker/stats [get]
func (a *AIServiceApi) GetCircuitBreakerStats(c *gin.Context) {
	stats := nesma.GetCircuitBreakerStats()
	response.OkWithData(stats, c)
}

// CircuitBreakerResetRequest 重置熔断器请求
type CircuitBreakerResetRequest struct {
	ServiceName string `json:"serviceName"`
	ResetAll    bool   `json:"resetAll"`
}

// @Tags AI服务管理
// @Summary 重置熔断器
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body CircuitBreakerResetRequest true "重置请求"
// @Success 200 {object} response.Response{} "重置成功"
// @Router /ai/services/circuit-breaker/reset [post]
func (a *AIServiceApi) ResetCircuitBreaker(c *gin.Context) {
	var req CircuitBreakerResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	if req.ResetAll {
		nesma.ResetAllCircuitBreakers()
		global.GVA_LOG.Info("All circuit breakers reset")
		response.OkWithMessage("所有熔断器已重置", c)
	} else if req.ServiceName != "" {
		if nesma.ResetCircuitBreaker(req.ServiceName) {
			global.GVA_LOG.Info("Circuit breaker reset for service: " + req.ServiceName)
			response.OkWithMessage("熔断器已重置: "+req.ServiceName, c)
		} else {
			response.FailWithMessage("服务不存在: "+req.ServiceName, c)
		}
	} else {
		response.FailWithMessage("请指定服务名称或选择重置所有", c)
	}
}
