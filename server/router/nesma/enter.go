package nesma

type RouterGroup struct {
	NesmaProjectRouter
	NesmaProjectCycleRouter
	NesmaRequirementRouter
	NesmaRequirementVersionRouter
	RequirementAnalysisRouter
	IntelligentAnalysisRouter
	KnowledgeRouter                   // 知识库路由
	AgentRouter
	DocumentRouter
	AIServiceRouter
	RecommendationRouter
	ChatRouter
	EvaluationRouter
	MonitoringRouter
	Level3AnalysisRouter
	Level4GeneratorRouter
	DescriptionGeneratorRouter
	MermaidGeneratorRouter
	DocumentExportRouter
	UnifiedAnalysisWorkflowRouter
	UnifiedAnalysisRouter
	ReportGeneratorRouter
	BatchEvaluationRouter
	ErrorHandlingRouter
	AIAnalysisRouter
}
