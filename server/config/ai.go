package config

// AI AI服务配置
type AI struct {
	DeepSeek         DeepSeek  `mapstructure:"deepseek" json:"deepseek" yaml:"deepseek"`
	DeepSeekReasoner DeepSeek  `mapstructure:"deepseek-reasoner" json:"deepseek-reasoner" yaml:"deepseek-reasoner"`
	OpenAI           OpenAI    `mapstructure:"openai" json:"openai" yaml:"openai"`
	Claude           Claude    `mapstructure:"claude" json:"claude" yaml:"claude"`
	Embedding        Embedding `mapstructure:"embedding" json:"embedding" yaml:"embedding"`
	Vector           Vector    `mapstructure:"vector" json:"vector" yaml:"vector"`
	Mermaid          Mermaid   `mapstructure:"mermaid" json:"mermaid" yaml:"mermaid"` // 新增Mermaid配置
}

// DeepSeek DeepSeek配置
type DeepSeek struct {
	APIKey      string  `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseURL     string  `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Model       string  `mapstructure:"model" json:"model" yaml:"model"`
	MaxTokens   int     `mapstructure:"max-tokens" json:"max-tokens" yaml:"max-tokens"`
	Temperature float64 `mapstructure:"temperature" json:"temperature" yaml:"temperature"`
	TopP        float64 `mapstructure:"top-p" json:"top-p" yaml:"top-p"`
	Timeout     int     `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

// OpenAI OpenAI配置
type OpenAI struct {
	APIKey      string  `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseURL     string  `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Model       string  `mapstructure:"model" json:"model" yaml:"model"`
	MaxTokens   int     `mapstructure:"max-tokens" json:"max-tokens" yaml:"max-tokens"`
	Temperature float64 `mapstructure:"temperature" json:"temperature" yaml:"temperature"`
	TopP        float64 `mapstructure:"top-p" json:"top-p" yaml:"top-p"`
	Timeout     int     `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

// Claude Claude配置
type Claude struct {
	APIKey      string  `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseURL     string  `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Model       string  `mapstructure:"model" json:"model" yaml:"model"`
	MaxTokens   int     `mapstructure:"max-tokens" json:"max-tokens" yaml:"max-tokens"`
	Temperature float64 `mapstructure:"temperature" json:"temperature" yaml:"temperature"`
	TopP        float64 `mapstructure:"top-p" json:"top-p" yaml:"top-p"`
	Timeout     int     `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

// Embedding 嵌入服务配置
type Embedding struct {
	Provider   string `mapstructure:"provider" json:"provider" yaml:"provider"`
	Model      string `mapstructure:"model" json:"model" yaml:"model"`
	APIKey     string `mapstructure:"api-key" json:"api-key" yaml:"api-key"`
	BaseURL    string `mapstructure:"base-url" json:"base-url" yaml:"base-url"`
	Dimensions int    `mapstructure:"dimensions" json:"dimensions" yaml:"dimensions"`
	Timeout    int    `mapstructure:"timeout" json:"timeout" yaml:"timeout"`
}

// Vector 向量数据库配置
type Vector struct {
	Provider            string  `mapstructure:"provider" json:"provider" yaml:"provider"`
	Dimensions          int     `mapstructure:"dimensions" json:"dimensions" yaml:"dimensions"`
	SimilarityThreshold float32 `mapstructure:"similarity-threshold" json:"similarity-threshold" yaml:"similarity-threshold"`
	MaxResults          int     `mapstructure:"max-results" json:"max-results" yaml:"max-results"`
	IndexType           string  `mapstructure:"index-type" json:"index-type" yaml:"index-type"`
}

// Mermaid Mermaid生成器配置
type Mermaid struct {
	// 并发配置
	MaxWorkers         int `mapstructure:"max-workers" json:"max-workers" yaml:"max-workers"`                   // 最大工作协程数
	BatchSize          int `mapstructure:"batch-size" json:"batch-size" yaml:"batch-size"`                     // 批量处理大小
	TaskChannelBuffer  int `mapstructure:"task-channel-buffer" json:"task-channel-buffer" yaml:"task-channel-buffer"`   // 任务通道缓冲区
	ResultChannelBuffer int `mapstructure:"result-channel-buffer" json:"result-channel-buffer" yaml:"result-channel-buffer"` // 结果通道缓冲区
	
	// 超时配置
	WorkerTimeout      int `mapstructure:"worker-timeout" json:"worker-timeout" yaml:"worker-timeout"`         // 工作协程超时时间(秒)
	BatchUpdateTimeout int `mapstructure:"batch-update-timeout" json:"batch-update-timeout" yaml:"batch-update-timeout"` // 批量更新超时时间(秒)
	
	// AI参数配置
	MaxTokens         int     `mapstructure:"max-tokens" json:"max-tokens" yaml:"max-tokens"`           // 最大token数
	Temperature       float64 `mapstructure:"temperature" json:"temperature" yaml:"temperature"`         // 温度参数
	TopP              float64 `mapstructure:"top-p" json:"top-p" yaml:"top-p"`                         // Top-P参数
	Model             string  `mapstructure:"model" json:"model" yaml:"model"`                         // 使用的模型
	
	// 流程图配置
	MaxNodes          int `mapstructure:"max-nodes" json:"max-nodes" yaml:"max-nodes"`                 // 最大节点数
	MaxDepth          int `mapstructure:"max-depth" json:"max-depth" yaml:"max-depth"`                 // 最大深度
	MaxEdges          int `mapstructure:"max-edges" json:"max-edges" yaml:"max-edges"`                 // 最大边数
	
	// 质量配置
	MinConfidence     float64 `mapstructure:"min-confidence" json:"min-confidence" yaml:"min-confidence"` // 最小置信度
	EnableValidation  bool    `mapstructure:"enable-validation" json:"enable-validation" yaml:"enable-validation"` // 是否启用验证
	EnableStyling     bool    `mapstructure:"enable-styling" json:"enable-styling" yaml:"enable-styling"`       // 是否启用样式
	
	// 主题配置
	DefaultTheme      string `mapstructure:"default-theme" json:"default-theme" yaml:"default-theme"`   // 默认主题
	ColorScheme       string `mapstructure:"color-scheme" json:"color-scheme" yaml:"color-scheme"`     // 颜色方案
}
