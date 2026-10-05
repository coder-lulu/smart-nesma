package nesma

type ServiceGroup struct {
	NesmaProjectService
	NesmaProjectCycleService
	NesmaRequirementService
	NesmaRequirementVersionService
	RequirementAnalysisService
	SingleRequirementAnalysisService
	WorkflowEngine
	DocumentService
	*DocumentExportService
	TemplateService
	*UnifiedAnalysisService
	KnowledgeService
	// *KnowledgeSearchEngine // 暂时注释掉，避免编译错误
}
