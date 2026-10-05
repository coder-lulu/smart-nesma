package config

// Agent Agent配置
type Agent struct {
	Registry  AgentRegistry  `mapstructure:"registry" json:"registry" yaml:"registry"`
	Workflows AgentWorkflows `mapstructure:"workflows" json:"workflows" yaml:"workflows"`
	Message   AgentMessage   `mapstructure:"message" json:"message" yaml:"message"`
	Health    AgentHealth    `mapstructure:"health" json:"health" yaml:"health"`
}

// AgentRegistry Agent注册配置
type AgentRegistry struct {
	HeartbeatInterval   int  `mapstructure:"heartbeat-interval" json:"heartbeat-interval" yaml:"heartbeat-interval"`          // 心跳间隔（秒）
	HealthCheckInterval int  `mapstructure:"health-check-interval" json:"health-check-interval" yaml:"health-check-interval"` // 健康检查间隔（秒）
	MaxConcurrentTasks  int  `mapstructure:"max-concurrent-tasks" json:"max-concurrent-tasks" yaml:"max-concurrent-tasks"`    // 最大并发任务数
	MessageQueueSize    int  `mapstructure:"message-queue-size" json:"message-queue-size" yaml:"message-queue-size"`          // 消息队列大小
	AgentTimeout        int  `mapstructure:"agent-timeout" json:"agent-timeout" yaml:"agent-timeout"`                         // Agent超时时间（秒）
	AutoCleanupTimeout  int  `mapstructure:"auto-cleanup-timeout" json:"auto-cleanup-timeout" yaml:"auto-cleanup-timeout"`    // 自动清理超时时间（秒）
	EnableMetrics       bool `mapstructure:"enable-metrics" json:"enable-metrics" yaml:"enable-metrics"`                      // 是否启用指标收集
	EnableDistributed   bool `mapstructure:"enable-distributed" json:"enable-distributed" yaml:"enable-distributed"`          // 是否启用分布式模式
}

// AgentWorkflows 工作流配置
type AgentWorkflows struct {
	MaxExecutionTime   int    `mapstructure:"max-execution-time" json:"max-execution-time" yaml:"max-execution-time"`       // 最大执行时间（秒）
	RetryAttempts      int    `mapstructure:"retry-attempts" json:"retry-attempts" yaml:"retry-attempts"`                   // 重试次数
	RetryDelay         int    `mapstructure:"retry-delay" json:"retry-delay" yaml:"retry-delay"`                            // 重试延迟（秒）
	EnablePersistence  bool   `mapstructure:"enable-persistence" json:"enable-persistence" yaml:"enable-persistence"`       // 是否启用持久化
	DefaultEngine      string `mapstructure:"default-engine" json:"default-engine" yaml:"default-engine"`                   // 默认工作流引擎
	MaxConcurrentFlows int    `mapstructure:"max-concurrent-flows" json:"max-concurrent-flows" yaml:"max-concurrent-flows"` // 最大并发工作流数
}

// AgentMessage 消息配置
type AgentMessage struct {
	MaxMessageSize    int    `mapstructure:"max-message-size" json:"max-message-size" yaml:"max-message-size"`       // 最大消息大小（字节）
	MessageTimeout    int    `mapstructure:"message-timeout" json:"message-timeout" yaml:"message-timeout"`          // 消息超时时间（秒）
	EnableCompression bool   `mapstructure:"enable-compression" json:"enable-compression" yaml:"enable-compression"` // 是否启用消息压缩
	EnableEncryption  bool   `mapstructure:"enable-encryption" json:"enable-encryption" yaml:"enable-encryption"`    // 是否启用消息加密
	QueueType         string `mapstructure:"queue-type" json:"queue-type" yaml:"queue-type"`                         // 队列类型：memory, redis, kafka
	MaxRetries        int    `mapstructure:"max-retries" json:"max-retries" yaml:"max-retries"`                      // 最大重试次数
	RetryInterval     int    `mapstructure:"retry-interval" json:"retry-interval" yaml:"retry-interval"`             // 重试间隔（秒）
}

// AgentHealth 健康检查配置
type AgentHealth struct {
	CheckInterval     int    `mapstructure:"check-interval" json:"check-interval" yaml:"check-interval"`                // 检查间隔（秒）
	Timeout           int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`                                     // 超时时间（秒）
	FailureThreshold  int    `mapstructure:"failure-threshold" json:"failure-threshold" yaml:"failure-threshold"`       // 失败阈值
	RecoveryTimeout   int    `mapstructure:"recovery-timeout" json:"recovery-timeout" yaml:"recovery-timeout"`          // 恢复超时时间（秒）
	EnableAutoRestart bool   `mapstructure:"enable-auto-restart" json:"enable-auto-restart" yaml:"enable-auto-restart"` // 是否启用自动重启
	HealthEndpoint    string `mapstructure:"health-endpoint" json:"health-endpoint" yaml:"health-endpoint"`             // 健康检查端点
}

// AgentType Agent类型配置
type AgentType struct {
	Name         string                 `mapstructure:"name" json:"name" yaml:"name"`                            // Agent类型名称
	Description  string                 `mapstructure:"description" json:"description" yaml:"description"`       // 描述
	Capabilities []string               `mapstructure:"capabilities" json:"capabilities" yaml:"capabilities"`    // 能力列表
	Config       map[string]interface{} `mapstructure:"config" json:"config" yaml:"config"`                      // 特定配置
	MaxInstances int                    `mapstructure:"max-instances" json:"max-instances" yaml:"max-instances"` // 最大实例数
	MinInstances int                    `mapstructure:"min-instances" json:"min-instances" yaml:"min-instances"` // 最小实例数
	AutoScale    bool                   `mapstructure:"auto-scale" json:"auto-scale" yaml:"auto-scale"`          // 是否自动扩缩容
}

// GetDefaultAgentConfig 获取默认Agent配置
func GetDefaultAgentConfig() Agent {
	return Agent{
		Registry: AgentRegistry{
			HeartbeatInterval:   30,
			HealthCheckInterval: 30,
			MaxConcurrentTasks:  3,
			MessageQueueSize:    100,
			AgentTimeout:        300,
			AutoCleanupTimeout:  3600,
			EnableMetrics:       true,
			EnableDistributed:   false,
		},
		Workflows: AgentWorkflows{
			MaxExecutionTime:   300,
			RetryAttempts:      3,
			RetryDelay:         5,
			EnablePersistence:  true,
			DefaultEngine:      "simple",
			MaxConcurrentFlows: 10,
		},
		Message: AgentMessage{
			MaxMessageSize:    1024 * 1024, // 1MB
			MessageTimeout:    30,
			EnableCompression: false,
			EnableEncryption:  false,
			QueueType:         "memory",
			MaxRetries:        3,
			RetryInterval:     5,
		},
		Health: AgentHealth{
			CheckInterval:     30,
			Timeout:           10,
			FailureThreshold:  3,
			RecoveryTimeout:   60,
			EnableAutoRestart: true,
			HealthEndpoint:    "/health",
		},
	}
}

// GetAgentTypeConfigs 获取预定义的Agent类型配置
func GetAgentTypeConfigs() map[string]AgentType {
	return map[string]AgentType{
		"requirement_refine": {
			Name:        "需求细化Agent",
			Description: "负责分析和细化用户需求，提供更详细的功能点描述",
			Capabilities: []string{
				"requirement_analysis",
				"requirement_refinement",
				"requirement_validation",
				"requirement_classification",
			},
			Config: map[string]interface{}{
				"ai_model":        "deepseek-chat",
				"max_tokens":      2000,
				"temperature":     0.7,
				"context_window":  4000,
				"prompt_template": "requirement_refine_v1",
			},
			MaxInstances: 5,
			MinInstances: 1,
			AutoScale:    true,
		},
		"nesma_evaluation": {
			Name:        "NESMA评估Agent",
			Description: "执行NESMA功能点评估，计算软件规模和复杂度",
			Capabilities: []string{
				"function_point_calculation",
				"complexity_analysis",
				"size_estimation",
				"quality_assessment",
			},
			Config: map[string]interface{}{
				"evaluation_rules": "nesma_v2.2",
				"precision_level":  "high",
				"validation_mode":  "strict",
				"report_format":    "standard",
			},
			MaxInstances: 3,
			MinInstances: 1,
			AutoScale:    false,
		},
		"document_generate": {
			Name:        "文档生成Agent",
			Description: "基于分析结果生成各类技术文档",
			Capabilities: []string{
				"document_generation",
				"template_processing",
				"content_formatting",
				"quality_checking",
			},
			Config: map[string]interface{}{
				"ai_model":          "deepseek-chat",
				"template_engine":   "jinja2",
				"output_formats":    []string{"markdown", "pdf", "docx"},
				"quality_threshold": 0.8,
			},
			MaxInstances: 5,
			MinInstances: 1,
			AutoScale:    true,
		},
		"memory": {
			Name:        "记忆Agent",
			Description: "管理和维护系统记忆，提供上下文信息",
			Capabilities: []string{
				"memory_storage",
				"context_retrieval",
				"information_indexing",
				"relevance_scoring",
			},
			Config: map[string]interface{}{
				"storage_type":     "redis",
				"index_algorithm":  "faiss",
				"retention_period": 7 * 24 * 3600, // 7天
				"max_memory_size":  1000000,       // 100万条记录
			},
			MaxInstances: 2,
			MinInstances: 1,
			AutoScale:    false,
		},
		"knowledge": {
			Name:        "知识库Agent",
			Description: "提供知识检索和推荐服务",
			Capabilities: []string{
				"knowledge_retrieval",
				"semantic_search",
				"recommendation",
				"knowledge_validation",
			},
			Config: map[string]interface{}{
				"search_engine":        "elasticsearch",
				"embedding_model":      "text-embedding-ada-002",
				"similarity_threshold": 0.7,
				"max_results":          20,
			},
			MaxInstances: 3,
			MinInstances: 1,
			AutoScale:    true,
		},
	}
}
