package initialize

import (
	"context"
	"encoding/json"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	nesmaModel "github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// InitNesmaServices 初始化NESMA相关服务
func InitNesmaServices() {
	// 初始化传统AI服务
	nesma.InitAIServices()
	
	// 初始化增强AI服务管理器 (支持deepseek-reasoner)
	aiServiceManager := nesma.GetAIServiceManager()
	global.GVA_LOG.Info("增强AI服务管理器初始化完成", 
		zap.Bool("deepseek_reasoner_available", aiServiceManager != nil))

	// 初始化推荐服务
	nesma.InitRecommendationService()
	
	// 初始化AI推荐服务
	nesma.InitAIRecommendationService()

	// 初始化聊天服务
	nesma.InitChatService()

	// 初始化监控服务
	nesma.InitMonitoringService()

	// 初始化需求分析服务
	// nesma.InitRequirementAnalysisService()  // 暂时注释掉未实现的服务

	// 初始化工作流引擎
	nesma.InitWorkflowEngine()

	// 初始化知识搜索引擎
	initKnowledgeSearchEngine()

	// 初始化批量评估服务
	initBatchEvaluationService()

	// 初始化其他服务
	initializeOtherServices()
}

// initializeOtherServices 初始化其他服务
func initializeOtherServices() {
	// 初始化内置Agent
	initBuiltinAgents()

	// 修复现有评估记录的周期和版本关联
	FixExistingEvaluations()

	// 可以在这里初始化其他NESMA相关服务
	zap.L().Info("NESMA服务初始化完成")
}

// initBuiltinAgents 初始化内置Agent
func initBuiltinAgents() {
	global.GVA_LOG.Info("开始初始化内置Agent系统...")

	if global.GVA_DB == nil {
		global.GVA_LOG.Error("数据库连接不可用，跳过内置Agent初始化")
		return
	}

	agentService := nesma.NewAgentService()
	ctx := context.Background()

	// 定义5个内置Agent
	builtinAgents := []nesmaModel.NesmaAgent{
		{
			AgentID:      "requirement-analysis-agent",
			Name:         "需求细化Agent",
			Description:  "负责需求扩充与细化，结合知识库提供上下文，分析用户输入的一二三级需求，生成详细的需求规格说明",
			AgentType:    "REQUIREMENT_ANALYSIS",
			Capabilities: createCapabilities([]string{"requirement_expansion", "context_analysis", "requirement_refinement", "knowledge_integration"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: createConfiguration(map[string]interface{}{
				"prompt_template": "requirement_analysis_prompt",
				"max_tokens":      2000,
				"temperature":     0.7,
				"model":           "deepseek-chat",
			}),
			MaxConcurrency: 5,
		},
		{
			AgentID:      "nesma-evaluation-agent",
			Name:         "NESMA评估Agent",
			Description:  "负责功能点识别与度量，应用NESMA标准规则，对需求进行功能点计算和复杂度评估",
			AgentType:    "NESMA_EVALUATION",
			Capabilities: createCapabilities([]string{"function_point_identification", "complexity_assessment", "nesma_calculation", "standard_compliance"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: createConfiguration(map[string]interface{}{
				"nesma_version":    "2.2",
				"calculation_mode": "detailed",
				"validation_level": "strict",
				"prompt_template":  "nesma_evaluation_prompt",
			}),
			MaxConcurrency: 3,
		},
		{
			AgentID:      "document-generation-agent",
			Name:         "文档生成Agent",
			Description:  "负责多格式文档生成，遵循最佳实践模板，生成需求规格说明书、NESMA评估报告等标准文档",
			AgentType:    "DOCUMENT_GENERATION",
			Capabilities: createCapabilities([]string{"document_templating", "multi_format_export", "content_structuring", "format_validation"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: createConfiguration(map[string]interface{}{
				"supported_formats": []string{"docx", "pdf", "excel", "html"},
				"template_version":  "1.0",
				"quality_check":     true,
				"auto_formatting":   true,
			}),
			MaxConcurrency: 4,
		},
		{
			AgentID:      "memory-agent",
			Name:         "记忆Agent",
			Description:  "负责长期记忆、上下文检索与智能遗忘，管理系统的记忆存储和检索，提供上下文感知能力",
			AgentType:    "MEMORY_MANAGEMENT",
			Capabilities: createCapabilities([]string{"vector_storage", "context_retrieval", "intelligent_forgetting", "memory_compression"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: createConfiguration(map[string]interface{}{
				"memory_types":         []string{"vector", "compact", "session"},
				"retention_policy":     "adaptive",
				"similarity_threshold": 0.7,
				"max_memory_size":      10000,
			}),
			MaxConcurrency: 10,
		},
		{
			AgentID:      "knowledge-base-agent",
			Name:         "知识库Agent",
			Description:  "负责知识检索、规则匹配、案例推荐，管理NESMA标准库、最佳实践库和历史案例库",
			AgentType:    "KNOWLEDGE_RETRIEVAL",
			Capabilities: createCapabilities([]string{"knowledge_search", "rule_matching", "case_recommendation", "content_validation"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: createConfiguration(map[string]interface{}{
				"knowledge_sources": []string{"nesma_standards", "best_practices", "case_studies", "rules"},
				"search_algorithm":  "semantic",
				"ranking_method":    "relevance_score",
				"cache_enabled":     true,
			}),
			MaxConcurrency: 8,
		},
	}

	// 注册所有内置Agent
	for i := range builtinAgents {
		agent := &builtinAgents[i]

		// 检查Agent是否已存在
		var existingAgent nesmaModel.NesmaAgent
		err := global.GVA_DB.Where("agent_id = ?", agent.AgentID).First(&existingAgent).Error
		if err == gorm.ErrRecordNotFound {
			// Agent不存在，注册新Agent
			err = agentService.RegisterAgent(ctx, agent)
			if err != nil {
				global.GVA_LOG.Error("注册内置Agent失败: " + agent.Name + ", 错误: " + err.Error())
			} else {
				global.GVA_LOG.Info("成功注册内置Agent: " + agent.Name)
			}
		} else if err == nil {
			// Agent已存在，更新配置
			existingAgent.Description = agent.Description
			existingAgent.Capabilities = agent.Capabilities
			existingAgent.Configuration = agent.Configuration
			existingAgent.Status = "active"

			err = agentService.UpdateAgent(ctx, &existingAgent)
			if err != nil {
				global.GVA_LOG.Error("更新内置Agent失败: " + agent.Name + ", 错误: " + err.Error())
			} else {
				global.GVA_LOG.Info("成功更新内置Agent: " + agent.Name)
			}
		} else {
			global.GVA_LOG.Error("检查内置Agent时出错: " + agent.Name + ", 错误: " + err.Error())
		}
	}

	global.GVA_LOG.Info("内置Agent系统初始化完成")
}

// createCapabilities 创建Agent能力配置
func createCapabilities(capabilities []string) datatypes.JSON {
	data, _ := json.Marshal(capabilities)
	return datatypes.JSON(data)
}

// createConfiguration 创建Agent配置
func createConfiguration(config map[string]interface{}) datatypes.JSON {
	data, _ := json.Marshal(config)
	return datatypes.JSON(data)
}

// initKnowledgeSearchEngine 初始化知识搜索引擎
func initKnowledgeSearchEngine() {
	global.GVA_LOG.Info("开始初始化知识搜索引擎...")

	// 创建知识搜索引擎实例
	// searchEngine := nesma.NewKnowledgeSearchEngine() // 暂时注释掉
	
	// 异步构建知识库向量索引
	/*
	go func() {
		ctx := context.Background()
		if err := searchEngine.BuildKnowledgeVectorIndex(ctx); err != nil {
			global.GVA_LOG.Error("构建知识库向量索引失败", zap.Error(err))
		} else {
			global.GVA_LOG.Info("知识库向量索引构建完成")
		}
	}()
	*/

	global.GVA_LOG.Info("知识搜索引擎初始化完成")
}

// initBatchEvaluationService 初始化批量评估服务
func initBatchEvaluationService() {
	global.GVA_LOG.Info("开始初始化批量评估服务...")

	// 获取依赖服务
	aiService := nesma.GetAIService()
	if aiService == nil {
		global.GVA_LOG.Error("AI服务未初始化，无法创建批量评估服务")
		return
	}

	// 这里应该获取知识库服务实例，暂时使用nil
	var knowledgeService *nesma.KnowledgeService

	// 创建批量评估服务实例
	batchEvaluationService := nesma.NewBatchEvaluationService(
		global.GVA_DB,
		aiService,
		knowledgeService,
	)

	// 设置全局服务实例
	nesma.SetBatchEvaluationService(batchEvaluationService)

	global.GVA_LOG.Info("批量评估服务初始化完成")
}
