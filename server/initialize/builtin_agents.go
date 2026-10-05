package initialize

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

const initOrderBuiltinAgents = system.InitOrderExternal + 1

type builtinAgents struct{}

// auto run
func init() {
	system.RegisterInit(initOrderBuiltinAgents, &builtinAgents{})
}

func (b *builtinAgents) InitializerName() string {
	return "builtin_agents"
}

func (b *builtinAgents) MigrateTable(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (b *builtinAgents) TableCreated(ctx context.Context) bool {
	return true
}

func (b *builtinAgents) InitializeData(ctx context.Context) (context.Context, error) {
	global.GVA_LOG.Info("开始初始化内置Agent系统...")

	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		global.GVA_LOG.Error("无法获取数据库上下文")
		return ctx, system.ErrMissingDBContext
	}

	agentService := nesmaService.NewAgentService()
	global.GVA_LOG.Info("Agent服务已创建")

	// 定义5个内置Agent
	builtinAgents := []nesma.NesmaAgent{
		{
			AgentID:      "requirement-analysis-agent",
			Name:         "需求细化Agent",
			Description:  "负责需求扩充与细化，结合知识库提供上下文，分析用户输入的一二三级需求，生成详细的需求规格说明",
			AgentType:    "REQUIREMENT_ANALYSIS",
			Capabilities: b.createCapabilities([]string{"requirement_expansion", "context_analysis", "requirement_refinement", "knowledge_integration"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: b.createConfiguration(map[string]interface{}{
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
			Capabilities: b.createCapabilities([]string{"function_point_identification", "complexity_assessment", "nesma_calculation", "standard_compliance"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: b.createConfiguration(map[string]interface{}{
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
			Capabilities: b.createCapabilities([]string{"document_templating", "multi_format_export", "content_structuring", "format_validation"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: b.createConfiguration(map[string]interface{}{
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
			Capabilities: b.createCapabilities([]string{"vector_storage", "context_retrieval", "intelligent_forgetting", "memory_compression"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: b.createConfiguration(map[string]interface{}{
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
			Capabilities: b.createCapabilities([]string{"knowledge_search", "rule_matching", "case_recommendation", "content_validation"}),
			Status:       "active",
			Version:      "1.0.0",
			Configuration: b.createConfiguration(map[string]interface{}{
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
		var existingAgent nesma.NesmaAgent
		err := db.Where("agent_id = ?", agent.AgentID).First(&existingAgent).Error
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

	return ctx, nil
}

func (b *builtinAgents) DataInserted(ctx context.Context) bool {
	db, ok := ctx.Value("db").(*gorm.DB)
	if !ok {
		global.GVA_LOG.Error("DataInserted: 无法获取数据库上下文")
		return false
	}

	// 检查所有内置Agent是否都已存在
	agentIDs := []string{
		"requirement-analysis-agent",
		"nesma-evaluation-agent",
		"document-generation-agent",
		"memory-agent",
		"knowledge-base-agent",
	}

	var count int64
	err := db.Model(&nesma.NesmaAgent{}).Where("agent_id IN ?", agentIDs).Count(&count).Error
	if err != nil {
		global.GVA_LOG.Error("DataInserted: 查询Agent数量失败: " + err.Error())
		return false
	}

	isInserted := count == int64(len(agentIDs))
	global.GVA_LOG.Info(fmt.Sprintf("DataInserted检查结果: 找到%d个Agent，期望%d个，数据已插入: %v", count, len(agentIDs), isInserted))

	return isInserted
}

// createCapabilities 创建Agent能力配置
func (b *builtinAgents) createCapabilities(capabilities []string) datatypes.JSON {
	data, _ := json.Marshal(capabilities)
	return datatypes.JSON(data)
}

// createConfiguration 创建Agent配置
func (b *builtinAgents) createConfiguration(config map[string]interface{}) datatypes.JSON {
	data, _ := json.Marshal(config)
	return datatypes.JSON(data)
}
