import service from '@/utils/request'

// ==================== 智能分析API ====================

// ==================== AI驱动分析 (新功能) ====================

// 启动AI需求分析
export const startAIAnalysis = (data) => {
  return service({
    url: '/nesma/analysis/analyze',
    method: 'post',
    data
  })
}

// 获取AI分析进度
export const getAIAnalysisProgress = (taskId) => {
  return service({
    url: `/nesma/analysis/progress/${taskId}`,
    method: 'get'
  })
}

// 获取AI分析任务列表
export const getAIAnalysisTasks = (params) => {
  return service({
    url: '/nesma/analysis/tasks',
    method: 'get',
    params
  })
}

// 获取AI分析任务统计
export const getAIAnalysisTaskStatistics = () => {
  return service({
    url: '/nesma/analysis/tasks/statistics',
    method: 'get'
  })
}

// 取消AI分析任务
export const cancelAIAnalysis = (taskId) => {
  return service({
    url: `/nesma/analysis/cancel/${taskId}`,
    method: 'post'
  })
}

// 删除AI分析任务
export const deleteAIAnalysisTask = (taskId) => {
  return service({
    url: `/nesma/analysis/tasks/${taskId}`,
    method: 'delete'
  })
}

// AI服务健康状态
export const getAIServiceHealth = () => {
  return service({
    url: '/ai/services/health',
    method: 'get'
  })
}

// 测试AI服务
export const testAIService = (data) => {
  return service({
    url: '/ai/services/test',
    method: 'post',
    data
  })
}

// 重置AI服务断路器
export const resetAICircuitBreaker = () => {
  return service({
    url: '/ai/services/circuit-breaker/reset',
    method: 'post'
  })
}

// 三级功能点分析
export const analyzeLevel3Requirements = (data) => {
  return service({
    url: '/nesma/analysis/level3/analyze',
    method: 'post',
    data
  })
}

// 批量三级功能点分析
export const batchAnalyzeLevel3Requirements = (data) => {
  return service({
    url: '/nesma/analysis/level3/batch-analyze',
    method: 'post',
    data
  })
}

// 应用三级优化建议
export const applyLevel3Optimization = (data) => {
  return service({
    url: '/nesma/analysis/level3/apply',
    method: 'post',
    data
  })
}

// 获取三级分析历史
export const getLevel3AnalysisHistory = (cycleId, params) => {
  return service({
    url: `/nesma/analysis/level3/history/${cycleId}`,
    method: 'get',
    params
  })
}

// 四级功能点生成
export const generateLevel4Requirements = (data) => {
  return service({
    url: '/nesma/generator/level4/generate',
    method: 'post',
    data
  })
}

// 批量四级功能点生成
export const batchGenerateLevel4Requirements = (data) => {
  return service({
    url: '/nesma/generator/level4/batch-generate',
    method: 'post',
    data
  })
}

// 应用四级功能点
export const applyLevel4Requirement = (data) => {
  return service({
    url: '/nesma/generator/level4/apply',
    method: 'post',
    data
  })
}

// 获取四级生成历史
export const getLevel4GenerationHistory = (cycleId, params) => {
  return service({
    url: `/nesma/generator/level4/history/${cycleId}`,
    method: 'get',
    params
  })
}

// 功能点描述生成
export const generateRequirementDescriptions = (data) => {
  return service({
    url: '/nesma/generator/description/generate',
    method: 'post',
    data
  })
}

// 批量描述生成
export const batchGenerateDescriptions = (data) => {
  return service({
    url: '/nesma/generator/description/batch-generate',
    method: 'post',
    data
  })
}

// 应用描述
export const applyDescription = (data) => {
  return service({
    url: '/nesma/generator/description/apply',
    method: 'post',
    data
  })
}

// 获取描述生成历史
export const getDescriptionGenerationHistory = (cycleId, params) => {
  return service({
    url: `/nesma/generator/description/history/${cycleId}`,
    method: 'get',
    params
  })
}

// Mermaid流程图生成
export const generateMermaidDiagrams = (data) => {
  return service({
    url: '/nesma/generator/mermaid/generate',
    method: 'post',
    data
  })
}

// 批量Mermaid生成
export const batchGenerateMermaidDiagrams = (data) => {
  return service({
    url: '/nesma/generator/mermaid/batch-generate',
    method: 'post',
    data
  })
}

// 应用Mermaid流程图
export const applyMermaidDiagram = (data) => {
  return service({
    url: '/nesma/generator/mermaid/apply',
    method: 'post',
    data
  })
}

// 获取Mermaid生成历史
export const getMermaidGenerationHistory = (cycleId, params) => {
  return service({
    url: `/nesma/generator/mermaid/history/${cycleId}`,
    method: 'get',
    params
  })
}

// 预览Mermaid流程图
export const previewMermaidDiagram = (data) => {
  return service({
    url: '/nesma/generator/mermaid/preview',
    method: 'post',
    data
  })
}

// 文档导出
export const exportComprehensiveDocument = (data) => {
  return service({
    url: '/nesma/export/comprehensive',
    method: 'post',
    data
  })
}

// 批量文档导出
export const batchExportDocuments = (data) => {
  return service({
    url: '/nesma/export/batch',
    method: 'post',
    data
  })
}

// 获取导出历史
export const getExportHistory = (projectId, params) => {
  return service({
    url: `/nesma/export/history/${projectId}`,
    method: 'get',
    params
  })
}

// 获取导出预览
export const getExportPreview = (data) => {
  return service({
    url: '/nesma/export/preview',
    method: 'post',
    data
  })
}

// 统一工作流执行
export const executeFullAnalysisWorkflow = (data) => {
  return service({
    url: '/nesma/workflow/execute',
    method: 'post',
    data
  })
}

// 获取工作流执行状态
export const getWorkflowExecutionStatus = (executionId) => {
  return service({
    url: `/nesma/workflow/status/${executionId}`,
    method: 'get'
  })
}

// 取消工作流执行
export const cancelWorkflowExecution = (executionId) => {
  return service({
    url: `/nesma/workflow/cancel/${executionId}`,
    method: 'post'
  })
}

// 获取工作流执行历史
export const getWorkflowExecutionHistory = (projectId, params) => {
  return service({
    url: `/nesma/workflow/history/${projectId}`,
    method: 'get',
    params
  })
}

// 获取工作流模板
export const getWorkflowTemplates = () => {
  return service({
    url: '/nesma/workflow/templates',
    method: 'get'
  })
}

// 验证工作流配置
export const validateWorkflowConfiguration = (data) => {
  return service({
    url: '/nesma/workflow/validate',
    method: 'post',
    data
  })
}

// 从模板创建工作流
export const createWorkflowFromTemplate = (data) => {
  return service({
    url: '/nesma/workflow/create-from-template',
    method: 'post',
    data
  })
}

// 测试用的默认数据
export const getTestData = () => {
  return {
    testRequirements: [
      {
        id: 1,
        title: "用户登录功能",
        description: "用户可以通过用户名和密码登录系统",
        level: 3,
        cycleId: 1,
        projectId: 1
      },
      {
        id: 2,
        title: "数据查询功能",
        description: "用户可以查询和筛选业务数据",
        level: 3,
        cycleId: 1,
        projectId: 1
      },
      {
        id: 3,
        title: "报表生成功能",
        description: "系统可以生成各种业务报表",
        level: 3,
        cycleId: 1,
        projectId: 1
      }
    ],
    testWorkflowConfig: {
      workflowId: "test_workflow_001",
      cycleId: 1,
      projectId: 1,
      executionMode: "sequential",
      enabledPhases: ["level3_analysis", "level4_generation", "description_generation", "mermaid_generation", "document_export"],
      qualityThresholds: {
        level3AnalysisConfidence: 0.8,
        level4GenerationConfidence: 0.8,
        descriptionQualityScore: 0.85,
        mermaidValidationScore: 0.75,
        overallQualityScore: 0.8
      },
      retryPolicy: {
        maxRetries: 3,
        retryDelaySeconds: 5,
        exponentialBackoff: true,
        retryableErrors: ["network_error", "ai_service_error"]
      },
      notificationSettings: {
        enableNotifications: false,
        notificationChannels: [],
        notificationTriggers: [],
        webhookUrl: "",
        emailRecipients: []
      },
      timeoutMinutes: 30
    }
  }
}