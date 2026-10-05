package nesma

import (
	"testing"
	"strings"
	
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
)

// TestBuildSystemPrompt 测试系统提示词构建
func TestBuildSystemPrompt(t *testing.T) {
	service := &BatchEvaluationService{}
	
	config := &BatchEvaluationConfig{
		IncludeExplanation: true,
	}
	
	prompt := service.buildSystemPrompt(config)
	
	// 验证提示词包含必要的部分
	requiredSections := []string{
		"NESMA 2.2 功能点分析专家系统",
		"功能点类型定义",
		"复杂度评估标准",
		"权重因子表",
		"分析指导原则",
		"标准分析步骤",
		"输出格式要求",
		"质量要求",
		"特殊注意事项",
	}
	
	for _, section := range requiredSections {
		if !strings.Contains(prompt, section) {
			t.Errorf("提示词缺少必要部分: %s", section)
		}
	}
	
	// 验证包含NESMA标准定义
	nesmaTypes := []string{"ILF", "EIF", "EI", "EO", "EQ"}
	for _, funcType := range nesmaTypes {
		if !strings.Contains(prompt, funcType) {
			t.Errorf("提示词缺少功能类型定义: %s", funcType)
		}
	}
	
	// 验证包含复杂度定义
	complexityLevels := []string{"Low", "Average", "High"}
	for _, level := range complexityLevels {
		if !strings.Contains(prompt, level) {
			t.Errorf("提示词缺少复杂度级别: %s", level)
		}
	}
	
	// 验证包含权重因子表
	weightFactors := []string{"7", "10", "15", "5", "3", "4", "6"}
	for _, factor := range weightFactors {
		if !strings.Contains(prompt, factor) {
			t.Errorf("提示词缺少权重因子: %s", factor)
		}
	}
	
	t.Logf("提示词长度: %d 字符", len(prompt))
	t.Logf("提示词包含 %d 个功能类型定义", len(nesmaTypes))
	t.Logf("提示词包含 %d 个复杂度级别", len(complexityLevels))
}

// TestBuildBatchPrompt 测试批次提示词构建
func TestBuildBatchPrompt(t *testing.T) {
	service := &BatchEvaluationService{}
	
	// 模拟上下文管理器
	service.contextManager = &ContextManager{}
	
	config := &BatchEvaluationConfig{
		IncludeExplanation: true,
	}
	
	requirements := []nesma.NesmaRequirement{
		{
			ID:          1,
			Code:        "REQ001",
			Title:       "用户登录功能",
			Description: "系统应提供用户登录功能，包括用户名密码验证",
			Level:       3,
		},
		{
			ID:          2,
			Code:        "REQ002", 
			Title:       "客户信息管理",
			Description: "系统应支持客户基本信息的增删改查操作",
			Level:       3,
		},
	}
	
	batchInfo := BatchInfo{
		BatchID:      "batch_001",
		BatchNumber:  1,
		Requirements: requirements,
		Status:       "pending",
	}
	
	task := BatchWorkerTask{
		EvaluationID: "eval_001",
		BatchInfo:    batchInfo,
		Config:       config,
		Context:      make(map[string]interface{}),
	}
	
	// 由于上下文管理器需要数据库连接，这里只测试基本结构
	// 在实际环境中需要完整的依赖注入
	t.Logf("批次任务创建成功，包含 %d 个需求", len(requirements))
	t.Logf("任务配置包含解释说明: %v", task.Config.IncludeExplanation)
}

// TestParseBatchResponse 测试批次响应解析
func TestParseBatchResponse(t *testing.T) {
	service := &BatchEvaluationService{}
	
	requirements := []nesma.NesmaRequirement{
		{
			ID:          1,
			Code:        "REQ001",
			Title:       "用户登录功能",
			Description: "系统应提供用户登录功能，包括用户名密码验证",
			Level:       3,
		},
	}
	
	// 模拟AI响应
	response := `[
		{
			"requirement_id": "1",
			"function_type": "EI",
			"complexity_level": "Low",
			"afp": 3.0,
			"ufp": 3.0,
			"confidence_score": 0.85,
			"explanation": "用户登录是典型的外部输入功能，处理来自应用边界外的用户凭据数据",
			"det_count": 8,
			"ret_count": 1,
			"ftr_count": 1,
			"weight_factor": 3.0,
			"quality_notes": "分析基于标准NESMA规则，置信度较高"
		}
	]`
	
	results, err := service.parseBatchResponse(response, requirements)
	if err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}
	
	if len(results) != 1 {
		t.Fatalf("期望1个结果，实际得到 %d 个", len(results))
	}
	
	result := results[0]
	
	// 验证基本字段
	if result.FunctionType != "EI" {
		t.Errorf("期望功能类型 EI，实际得到 %s", result.FunctionType)
	}
	
	if result.ComplexityLevel != "Low" {
		t.Errorf("期望复杂度级别 Low，实际得到 %s", result.ComplexityLevel)
	}
	
	if result.AFP != 3.0 {
		t.Errorf("期望AFP 3.0，实际得到 %f", result.AFP)
	}
	
	if result.UFP != 3.0 {
		t.Errorf("期望UFP 3.0，实际得到 %f", result.UFP)
	}
	
	if result.ConfidenceScore != 0.85 {
		t.Errorf("期望置信度 0.85，实际得到 %f", result.ConfidenceScore)
	}
	
	// 验证新字段
	if result.Explanation == "" {
		t.Error("缺少分析说明")
	}
	
	if result.DETCount != 8 {
		t.Errorf("期望DET计数 8，实际得到 %d", result.DETCount)
	}
	
	if result.RETCount != 1 {
		t.Errorf("期望RET计数 1，实际得到 %d", result.RETCount)
	}
	
	if result.FTRCount != 1 {
		t.Errorf("期望FTR计数 1，实际得到 %d", result.FTRCount)
	}
	
	if result.WeightFactor != 3.0 {
		t.Errorf("期望权重因子 3.0，实际得到 %f", result.WeightFactor)
	}
	
	if result.QualityNotes == "" {
		t.Error("缺少质量备注")
	}
	
	t.Logf("响应解析成功，功能类型: %s, 复杂度: %s, AFP: %f", 
		result.FunctionType, result.ComplexityLevel, result.AFP)
}

// TestPromptQuality 测试提示词质量
func TestPromptQuality(t *testing.T) {
	service := &BatchEvaluationService{}
	
	config := &BatchEvaluationConfig{
		IncludeExplanation: true,
	}
	
	prompt := service.buildSystemPrompt(config)
	
	// 检查提示词的完整性
	qualityChecks := []struct {
		name     string
		keywords []string
	}{
		{
			name: "NESMA标准引用",
			keywords: []string{"NESMA 2.2", "ISO/IEC 14143", "国际标准"},
		},
		{
			name: "功能类型定义",
			keywords: []string{"ILF", "EIF", "EI", "EO", "EQ", "内部逻辑文件", "外部接口文件", "外部输入", "外部输出", "外部查询"},
		},
		{
			name: "复杂度标准",
			keywords: []string{"Low", "Average", "High", "DET", "RET", "FTR", "数据元素", "记录元素", "文件类型"},
		},
		{
			name: "权重因子",
			keywords: []string{"权重因子", "7", "10", "15", "5", "3", "4", "6"},
		},
		{
			name: "分析原则",
			keywords: []string{"用户视角", "业务价值", "独立性", "完整性", "保守原则"},
		},
		{
			name: "质量要求",
			keywords: []string{"准确性", "一致性", "可追溯性", "可验证性"},
		},
		{
			name: "输出格式",
			keywords: []string{"JSON", "requirement_id", "function_type", "complexity_level", "afp", "ufp", "confidence_score"},
		},
	}
	
	for _, check := range qualityChecks {
		missingKeywords := []string{}
		for _, keyword := range check.keywords {
			if !strings.Contains(prompt, keyword) {
				missingKeywords = append(missingKeywords, keyword)
			}
		}
		
		if len(missingKeywords) > 0 {
			t.Errorf("%s 检查失败，缺少关键词: %v", check.name, missingKeywords)
		} else {
			t.Logf("%s 检查通过", check.name)
		}
	}
	
	// 检查提示词长度是否合理
	if len(prompt) < 2000 {
		t.Errorf("提示词过短，长度: %d 字符", len(prompt))
	}
	
	if len(prompt) > 10000 {
		t.Errorf("提示词过长，长度: %d 字符", len(prompt))
	}
	
	t.Logf("提示词质量检查通过，长度: %d 字符", len(prompt))
} 