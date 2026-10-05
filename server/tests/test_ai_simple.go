package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
)

func main() {
	fmt.Println("=== Smart-NESMA AI分析流程测试 ===")

	// 简单的配置检查
	deepseekAPIKey := os.Getenv("DEEPSEEK_API_KEY")
	if deepseekAPIKey == "" {
		// 没有配置密钥时不发起外部请求
		fmt.Println("请设置 DEEPSEEK_API_KEY 后运行此测试")
		return
	}

	// 1. 测试DeepSeek服务创建
	fmt.Println("\n1. 测试DeepSeek服务创建...")
	testDeepSeekService(deepseekAPIKey)

	// 2. 测试AI分析prompt构建
	fmt.Println("\n2. 测试AI分析prompt构建...")
	testPromptBuilding()

	// 3. 测试端到端AI分析
	fmt.Println("\n3. 测试端到端AI分析...")
	testEndToEndAnalysis(deepseekAPIKey)

	fmt.Println("\n=== 测试完成 ===")
}

func testDeepSeekService(apiKey string) {
	// 创建DeepSeek服务实例
	deepseekService := nesma.NewDeepSeekService(
		apiKey,
		"https://api.deepseek.com",
		"deepseek-chat",
	)

	if deepseekService == nil {
		fmt.Println("❌ DeepSeek服务创建失败")
		return
	}

	fmt.Println("✅ DeepSeek服务创建成功")

	// 测试健康检查（简单测试，可能因为API Key而失败）
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	isHealthy := deepseekService.IsHealthy(ctx)
	if isHealthy {
		fmt.Println("✅ DeepSeek服务健康检查通过")
	} else {
		fmt.Println("⚠️ DeepSeek服务健康检查失败（可能是API Key或网络问题）")
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
		Description: "用户可以通过用户名和密码登录系统，验证成功后进入主界面",
		Level:       4,
	}

	// 构建NESMA分析prompt
	prompt := buildNESMAAnalysisPrompt(testReq)

	fmt.Printf("✅ 生成的NESMA分析prompt:\n")
	fmt.Printf("📏 Prompt长度: %d字符\n", len(prompt))
	fmt.Printf("📝 Prompt预览: %s...\n", prompt[:min(200, len(prompt))])

	// 验证prompt关键组件
	requiredComponents := []string{
		"NESMA功能点分析专家",
		"需求标题",
		"需求描述",
		"JSON格式",
		"function_type",
		"complexity_score",
		"confidence_score",
		"EI/EO/EQ/ILF/EIF",
	}

	fmt.Println("\n🔍 检查Prompt关键组件:")
	for _, component := range requiredComponents {
		if contains(prompt, component) {
			fmt.Printf("  ✅ %s\n", component)
		} else {
			fmt.Printf("  ❌ %s\n", component)
		}
	}
}

func testEndToEndAnalysis(apiKey string) {
	if apiKey == "" {
		fmt.Println("⚠️ 未配置有效的API Key，跳过端到端测试")
		return
	}

	// 创建DeepSeek服务
	deepseekService := nesma.NewDeepSeekService(apiKey, "", "")

	// 创建测试需求
	testReq := struct {
		Title       string
		Description string
		Level       int
	}{
		Title:       "商品搜索功能",
		Description: "用户可以输入关键词搜索商品，系统返回匹配的商品列表",
		Level:       4,
	}

	// 构建prompt
	prompt := buildNESMAAnalysisPrompt(testReq)

	// 配置AI请求
	config := &nesma.AIConfig{
		MaxTokens:   800,
		Temperature: 0.7,
		Model:       "deepseek-chat",
	}

	// 执行AI分析
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("🔄 正在调用DeepSeek API进行NESMA分析...")

	response, err := deepseekService.GenerateText(ctx, prompt, config)
	if err != nil {
		fmt.Printf("❌ AI分析调用失败: %v\n", err)
		return
	}

	if len(response.Choices) > 0 {
		content := response.Choices[0].Message.Content
		fmt.Printf("✅ AI分析响应成功\n")
		fmt.Printf("📊 Token使用: prompt=%d, completion=%d, total=%d\n",
			response.Usage.PromptTokens, response.Usage.CompletionTokens, response.Usage.TotalTokens)
		fmt.Printf("📝 响应长度: %d字符\n", len(content))

		// 简单验证响应格式
		if contains(content, "{") && contains(content, "}") {
			fmt.Println("✅ 响应包含JSON格式")
		} else {
			fmt.Println("⚠️ 响应可能不是JSON格式")
		}

		if contains(content, "function_type") {
			fmt.Println("✅ 响应包含功能类型分析")
		}

		if contains(content, "complexity_score") {
			fmt.Println("✅ 响应包含复杂度评分")
		}

		fmt.Printf("\n💬 AI分析响应预览:\n%s\n", content[:min(500, len(content))])
	} else {
		fmt.Println("❌ AI分析响应为空")
	}
}

// 辅助函数
func buildNESMAAnalysisPrompt(req interface{}) string {
	testReq := req.(struct {
		Title       string
		Description string
		Level       int
	})

	return fmt.Sprintf(`你是一位NESMA功能点分析专家，请对以下需求进行专业分析：

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
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (len(substr) == 0 || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			if s[i+j] != substr[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
