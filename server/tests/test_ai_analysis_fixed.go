package main

import (
	"context"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
)

func main() {
	// 初始化配置
	global.GVA_VP = core.Viper()
	global.GVA_LOG = core.Zap()

	fmt.Println("=== Smart-NESMA AI分析流程测试 ===")

	// 检查配置是否正确
	if global.GVA_CONFIG.AI.DeepSeek.APIKey == "" {
		fmt.Println("❌ 配置中未找到DeepSeek API Key")
		return
	}

	// 1. 测试AI服务初始化
	fmt.Println("\n1. 测试AI服务初始化...")
	nesma.InitAIServices()
	
	availableServices := nesma.ListAvailableAIServices()
	fmt.Printf("可用AI服务: %v\n", availableServices)

	// 2. 测试DeepSeek服务连接
	fmt.Println("\n2. 测试DeepSeek服务连接...")
	testDeepSeekConnection()

	// 3. 测试AI分析prompt构建和调用
	fmt.Println("\n3. 测试AI分析prompt构建和调用...")
	testPromptBuilding()

	// 4. 测试熔断器功能
	fmt.Println("\n4. 测试熔断器功能...")
	testCircuitBreaker()

	fmt.Println("\n=== 测试完成 ===")
}

func testDeepSeekConnection() {
	// 检查配置
	if global.GVA_CONFIG.AI.DeepSeek.APIKey == "" {
		fmt.Println("⚠️ 未配置DeepSeek API Key，跳过连接测试")
		return
	}

	// 创建DeepSeek服务实例
	deepseekService := nesma.NewDeepSeekService(
		global.GVA_CONFIG.AI.DeepSeek.APIKey,
		global.GVA_CONFIG.AI.DeepSeek.BaseURL,
		global.GVA_CONFIG.AI.DeepSeek.Model,
	)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	config := &nesma.AIConfig{
		MaxTokens:   50,
		Temperature: 0.1,
		TopP:        0.95,
		Model:       "deepseek-chat",
	}

	response, err := deepseekService.GenerateText(ctx, "请回复'连接成功'", config)
	if err != nil {
		fmt.Printf("❌ DeepSeek连接失败: %v\n", err)
		return
	}

	if len(response.Choices) > 0 {
		fmt.Printf("✅ DeepSeek连接成功: %s\n", response.Choices[0].Message.Content)
	} else {
		fmt.Printf("⚠️ DeepSeek响应为空\n")
	}
}

func testPromptBuilding() {
	// 创建测试需求
	testReq := struct {
		Title       string
		Description string
		Level       int
	}{
		Title:       "用户登录功能",
		Description: "用户可以通过用户名和密码登录系统",
		Level:       4,
	}

	// 构建NESMA分析prompt
	prompt := fmt.Sprintf(`你是一位NESMA功能点分析专家，请对以下需求进行专业分析：

需求标题：%s
需求描述：%s
需求级别：%d级

请按照NESMA标准分析该需求，并以JSON格式返回结果：

{
  "optimized_title": "优化后的需求标题",
  "optimized_description": "详细的功能描述，包括输入、处理、输出",
  "function_type": "EI/EO/EQ/ILF/EIF中的一种",
  "complexity_score": 0-100的复杂度评分,
  "confidence_score": 0-1的置信度,
  "recommended_afp": 推荐的AFP值,
  "recommended_ufp": 推荐的UFP值,
  "analysis_notes": "分析说明和改进建议"
}

注意：
1. 优化描述要详细说明功能的业务价值和技术实现要点
2. 功能类型分类要准确：EI(外部输入)、EO(外部输出)、EQ(外部查询)、ILF(内部逻辑文件)、EIF(外部接口文件)
3. 复杂度评分要综合考虑业务复杂度和技术复杂度
4. 功能点推荐要符合NESMA标准`, testReq.Title, testReq.Description, testReq.Level)

	fmt.Printf("✅ 生成的NESMA分析prompt长度: %d字符\n", len(prompt))

	// 检查配置
	if global.GVA_CONFIG.AI.DeepSeek.APIKey == "" {
		fmt.Println("⚠️ 未配置DeepSeek API Key，跳过AI调用测试")
		return
	}

	// 测试AI调用
	deepseekService := nesma.NewDeepSeekService(
		global.GVA_CONFIG.AI.DeepSeek.APIKey,
		global.GVA_CONFIG.AI.DeepSeek.BaseURL,
		global.GVA_CONFIG.AI.DeepSeek.Model,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	config := &nesma.AIConfig{
		MaxTokens:   1000,
		Temperature: 0.7,
		TopP:        0.95,
		Model:       "deepseek-chat",
	}

	fmt.Println("🔄 正在调用DeepSeek API进行NESMA分析...")
	response, err := deepseekService.GenerateText(ctx, prompt, config)
	if err != nil {
		fmt.Printf("❌ AI分析调用失败: %v\n", err)
		return
	}

	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		fmt.Printf("✅ AI分析响应长度: %d字符\n", len(content))
		fmt.Printf("📊 Token使用: prompt=%d, completion=%d, total=%d\n", 
			response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)
		
		// 只显示响应的前200个字符
		if len(content) > 200 {
			fmt.Printf("💬 AI分析响应预览: %s...\n", content[:200])
		} else {
			fmt.Printf("💬 AI分析响应: %s\n", content)
		}
	} else {
		fmt.Printf("⚠️ AI分析响应为空\n")
	}
}

func testCircuitBreaker() {
	fmt.Println("🔧 测试熔断器功能...")
	
	// 获取熔断器统计信息
	stats := nesma.GetCircuitBreakerStats()
	fmt.Printf("📊 熔断器统计信息: %+v\n", stats)
	
	// 获取AI服务健康状态
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	health := nesma.GetAIServiceHealth(ctx)
	fmt.Printf("🏥 AI服务健康状态: %+v\n", health)
}