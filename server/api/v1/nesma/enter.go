package nesma

type ApiGroup struct {
	NesmaProjectApi
	NesmaProjectCycleApi
	NesmaRequirementApi
	NesmaRequirementVersionApi
	// KnowledgeApi // 暂时注释掉
	AgentApi
	AgentTestApi
	DocumentApi
	TemplateApi
	AIServiceApi
	RecommendationApi
	AIRecommendationApi
	ChatApi
	RequirementAnalysisApi
	IntelligentAnalysisApi
	EvaluationApi
	MonitoringApi
	Level3AnalysisApi
	Level4GeneratorApi
	DescriptionGeneratorApi
	MermaidGeneratorApi
	DocumentExportApi
	UnifiedAnalysisWorkflowApi
	UnifiedAnalysisApi
	ConcurrentConfigApi
	PerformanceMonitorApi
	AIAnalysisApi
}
