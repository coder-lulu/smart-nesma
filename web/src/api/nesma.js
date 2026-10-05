import service from '@/utils/request'
import axios from 'axios'

// ==================== 项目管理 ====================

// 创建项目
export const createNesmaProject = (data) => {
  return service({
    url: '/nesma/project',
    method: 'post',
    data
  })
}

// 获取项目列表
export const getNesmaProjectList = (params) => {
  return service({
    url: '/nesma/project/list',
    method: 'get',
    params
  })
}

// 获取项目详情
export const getNesmaProject = (id) => {
  return service({
    url: `/nesma/project/${id}`,
    method: 'get'
  })
}

// 更新项目
export const updateNesmaProject = (data) => {
  return service({
    url: '/nesma/project',
    method: 'put',
    data
  })
}

// 删除项目
export const deleteNesmaProject = (data) => {
  return service({
    url: '/nesma/project',
    method: 'delete',
    data
  })
}

// 批量删除项目
export const deleteNesmaProjectByIds = (data) => {
  return service({
    url: '/nesma/project/delete-batch',
    method: 'delete',
    data
  })
}

// 获取项目统计
export const getNesmaProjectStats = () => {
  return service({
    url: '/nesma/project/stats',
    method: 'get'
  })
}

// 归档项目
export const archiveNesmaProject = (data) => {
  return service({
    url: '/nesma/project/archive',
    method: 'post',
    data
  })
}

// 恢复项目
export const restoreNesmaProject = (data) => {
  return service({
    url: '/nesma/project/restore',
    method: 'post',
    data
  })
}

// ==================== 项目周期管理 ====================

// 创建项目周期
export const createProjectCycle = (data) => {
  return service({
    url: '/nesma/project-cycle',
    method: 'post',
    data
  })
}

// 更新项目周期
export const updateProjectCycle = (id, data) => {
  return service({
    url: `/nesma/project-cycle/${id}`,
    method: 'put',
    data
  })
}

// 删除项目周期
export const deleteProjectCycle = (id) => {
  return service({
    url: `/nesma/project-cycle/${id}`,
    method: 'delete'
  })
}

// 设置当前活跃周期
export const setActiveProjectCycle = (cycleId) => {
  return service({
    url: '/nesma/project-cycle/set-active',
    method: 'post',
    data: {
      cycleId: cycleId
    }
  })
}

// 获取项目周期列表
export const getProjectCycleList = (params) => {
  return service({
    url: '/nesma/project-cycle/list',
    method: 'get',
    params
  })
}

// 获取项目周期详情
export const getProjectCycle = (id) => {
  return service({
    url: `/nesma/project-cycle/${id}`,
    method: 'get'
  })
}

// 更新周期状态
export const updateProjectCycleStatus = (id, data) => {
  return service({
    url: `/nesma/project-cycle/status`,
    method: 'put',
    data
  })
}

// 获取项目的所有周期
export const getProjectCycles = (params) => {
  return service({
    url: '/nesma/project-cycle/by-project',
    method: 'get',
    params
  })
}

// 获取版本列表
export const getVersionList = (params) => {
  return service({
    url: '/nesma/requirement-version/list',
    method: 'get',
    params
  })
}
export const getProjectAllCycles = (projectId) => {
  return service({
    url: `/nesma/project-cycle/project/${projectId}`,
    method: 'get'
  })
}


// ==================== 需求管理 ====================

// 创建需求
export const createNesmaRequirement = (data) => {
  return service({
    url: '/nesma/requirement',
    method: 'post',
    data
  })
}

// 获取需求列表
export const getNesmaRequirementList = (params) => {
  return service({
    url: '/nesma/requirement/list',
    method: 'get',
    params
  })
}

// 获取需求详情
export const getNesmaRequirement = (id) => {
  return service({
    url: `/nesma/requirement/${id}`,
    method: 'get'
  })
}

// 更新需求
export const updateNesmaRequirement = (data) => {
  return service({
    url: '/nesma/requirement',
    method: 'put',
    data
  })
}

// 获取需求的子功能点信息
export const getRequirementChildren = (id) => {
  return service({
    url: `/nesma/requirement/${id}/children`,
    method: 'get'
  })
}

// 删除需求（支持级联删除）
export const deleteNesmaRequirement = (id, forceDelete = false) => {
  return service({
    url: `/nesma/requirement/${id}`,
    method: 'delete',
    params: { force: forceDelete }
  })
}

// 批量删除需求
export const batchDeleteNesmaRequirements = (data) => {
  return service({
    url: '/nesma/requirement/batch-delete',
    method: 'post',
    data
  })
}

// 条件删除需求
export const deleteRequirementsByCondition = (data) => {
  return service({
    url: '/nesma/requirement/delete-by-condition',
    method: 'post',
    data
  })
}

// 获取项目最大版本号
export const getProjectMaxVersion = (projectId) => {
  return service({
    url: `/nesma/requirement/max-version/${projectId}`,
    method: 'get'
  })
}

// 获取需求树
export const getNesmaRequirementTree = (params) => {
  return service({
    url: '/nesma/requirement/tree',
    method: 'get',
    params
  })
}

// 获取需求统计
export const getNesmaRequirementStats = (params) => {
  return service({
    url: '/nesma/requirement/stats',
    method: 'get',
    params
  })
}

// 获取父需求选项
export const getParentRequirementOptions = (params) => {
  return service({
    url: '/nesma/requirement/parent-options',
    method: 'get',
    params
  })
}

// 从Excel导入需求
export const importRequirementsFromExcel = (data) => {
  return service({
    url: '/nesma/requirement/import-excel',
    method: 'post',
    data
  })
}

// 批量更新排序
export const batchUpdateOrder = (data) => {
  return service({
    url: '/nesma/requirement/batch-update-order',
    method: 'post',
    data
  })
}

// 移动需求
export const moveRequirement = (data) => {
  return service({
    url: '/nesma/requirement/move',
    method: 'post',
    data
  })
}

// 需求分析相关API (已迁移到单个需求AI分析专用接口)

// 批量分析需求
export const batchAnalyzeRequirements = (data) => {
  return service({
    url: '/requirement-analysis/batch-analyze',
    method: 'post',
    data
  })
}

// 快速分析
export const quickAnalyzeRequirement = (data) => {
  return service({
    url: '/requirement-analysis/quick-analyze',
    method: 'post',
    data
  })
}

// 对比分析结果
export const compareAnalysisResults = (data) => {
  return service({
    url: '/requirement-analysis/compare',
    method: 'post',
    data
  })
}

// 获取分析历史
export const getRequirementAnalysisHistory = (params) => {
  return service({
    url: '/requirement-analysis/history',
    method: 'get',
    params
  })
}

// 获取分析结果 (已迁移到单个需求AI分析专用接口)

// 获取分析配置
export const getAnalysisConfig = () => {
  return service({
    url: '/requirement-analysis/config',
    method: 'get'
  })
}

// 获取分析统计
export const getAnalysisStats = (params) => {
  return service({
    url: '/requirement-analysis/stats',
    method: 'get',
    params
  })
}

// ==================== Agent管理 ====================

// 注册Agent
export const registerAgent = (data) => {
  return service({
    url: '/agent/register',
    method: 'post',
    data
  })
}

// 获取Agent列表
export const getAgentList = (params) => {
  return service({
    url: '/agent/list',
    method: 'get',
    params
  })
}

// 获取Agent详情
export const getAgent = (id) => {
  return service({
    url: `/agent/${id}`,
    method: 'get'
  })
}

// 更新Agent状态
export const updateAgentStatus = (id, data) => {
  return service({
    url: `/agent/${id}/status`,
    method: 'put',
    data
  })
}

// Agent心跳
export const agentHeartbeat = (id) => {
  return service({
    url: `/agent/${id}/heartbeat`,
    method: 'post'
  })
}

// 获取Agent统计
export const getAgentStatistics = () => {
  return service({
    url: '/agent/statistics',
    method: 'get'
  })
}

// 获取Agent健康状态
export const getAgentHealth = () => {
  return service({
    url: '/agent/health',
    method: 'get'
  })
}

// 获取系统性能指标
export const getSystemMetrics = () => {
  return service({
    url: '/agent/metrics',
    method: 'get'
  })
}

// 创建任务
export const createTask = (data) => {
  return service({
    url: '/agent/task',
    method: 'post',
    data
  })
}

// 分配任务
export const assignTask = (taskId, data) => {
  return service({
    url: `/agent/task/${taskId}/assign`,
    method: 'post',
    data
  })
}

// 更新Agent
export const updateAgent = (data) => {
  return service({
    url: '/agent',
    method: 'put',
    data
  })
}

// 删除Agent
export const deleteAgent = (id) => {
  return service({
    url: `/agent/${id}`,
    method: 'delete'
  })
}

// 获取可用Agent
export const getAvailableAgents = (params) => {
  return service({
    url: '/agent/available',
    method: 'get',
    params
  })
}

// 获取任务详情
export const getTask = (taskId) => {
  return service({
    url: `/agent/task/${taskId}`,
    method: 'get'
  })
}

// 更新任务
export const updateTask = (data) => {
  return service({
    url: '/agent/task',
    method: 'put',
    data
  })
}

// 执行任务
export const executeTask = (data) => {
  return service({
    url: '/agent/task/execute',
    method: 'post',
    data
  })
}

// 获取Agent的任务列表
export const getTasksByAgent = (agentId, params) => {
  return service({
    url: `/agent/${agentId}/tasks`,
    method: 'get',
    params
  })
}

// 根据状态获取任务列表
export const getTasksByStatus = (params) => {
  return service({
    url: '/agent/tasks',
    method: 'get',
    params
  })
}

// 获取消息列表
export const getMessages = (sessionId, params) => {
  return service({
    url: `/agent/messages/${sessionId}`,
    method: 'get',
    params
  })
}

// 获取Agent日志
export const getAgentLogs = (agentId, params) => {
  return service({
    url: `/agent/${agentId}/logs`,
    method: 'get',
    params
  })
}

// 发送Agent消息
export const sendAgentMessage = (data) => {
  return service({
    url: '/agent/message',
    method: 'post',
    data
  })
}

// ==================== 工作流管理 ====================

// 创建工作流
export const createWorkflow = (data) => {
  return service({
    url: '/agent/workflow',
    method: 'post',
    data
  })
}

// 获取工作流列表
export const getWorkflowList = (params) => {
  return service({
    url: '/agent/workflows',
    method: 'get',
    params
  })
}

// 获取工作流详情
export const getWorkflow = (id) => {
  return service({
    url: `/agent/workflow/${id}`,
    method: 'get'
  })
}

// 更新工作流
export const updateWorkflow = (data) => {
  return service({
    url: '/agent/workflow',
    method: 'put',
    data
  })
}

// 删除工作流
export const deleteWorkflow = (id) => {
  return service({
    url: `/agent/workflow/${id}`,
    method: 'delete'
  })
}

// 执行工作流
export const executeWorkflow = (id, data) => {
  return service({
    url: `/agent/workflow/${id}/execute`,
    method: 'post',
    data
  })
}

// 获取工作流执行详情
export const getWorkflowExecution = (id) => {
  return service({
    url: `/agent/execution/${id}`,
    method: 'get'
  })
}

// 获取工作流的所有执行记录
export const getWorkflowExecutions = (workflowId, params) => {
  return service({
    url: `/agent/workflow/${workflowId}/executions`,
    method: 'get',
    params
  })
}

// 取消Agent执行
export const cancelAgentExecution = (id) => {
  return service({
    url: `/agent/execution/${id}/cancel`,
    method: 'post'
  })
}

// 取消工作流执行
export const cancelWorkflowExecution = (id) => {
  return service({
    url: `/nesma/workflow/execution/${id}/cancel`,
    method: 'post'
  })
}

// 获取工作流统计
export const getWorkflowStatistics = () => {
  return service({
    url: '/agent/statistics/workflow',
    method: 'get'
  })
}

// ==================== 知识库管理 ====================

// 创建知识条目
export const createKnowledgeEntry = (data) => {
  return service({
    url: '/knowledge',
    method: 'post',
    data
  })
}

// 更新知识条目
export const updateKnowledgeEntry = (data) => {
  return service({
    url: `/knowledge/${data.id}`,
    method: 'put',
    data
  })
}

// 删除知识条目
export const deleteKnowledgeEntry = (id) => {
  return service({
    url: `/knowledge/${id}`,
    method: 'delete'
  })
}

// 获取知识条目详情
export const getKnowledgeEntry = (id) => {
  return service({
    url: `/knowledge/${id}`,
    method: 'get'
  })
}

// 获取知识条目列表
export const getKnowledgeEntryList = (params) => {
  return service({
    url: '/knowledge',
    method: 'get',
    params
  })
}

// 搜索知识条目
export const searchKnowledgeEntries = (data) => {
  return service({
    url: '/knowledge/entries/search',
    method: 'post',
    data
  })
}

// 创建知识规则
export const createKnowledgeRule = (data) => {
  return service({
    url: '/knowledge/rule',
    method: 'post',
    data
  })
}

// 更新知识规则
export const updateKnowledgeRule = (data) => {
  return service({
    url: `/knowledge/rule/${data.id}`,
    method: 'put',
    data
  })
}

// 删除知识规则
export const deleteKnowledgeRule = (id) => {
  return service({
    url: `/knowledge/rule/${id}`,
    method: 'delete'
  })
}

// 获取知识规则详情
export const getKnowledgeRule = (id) => {
  return service({
    url: `/knowledge/rule/${id}`,
    method: 'get'
  })
}

// 获取知识规则列表
export const getKnowledgeRuleList = (params) => {
  return service({
    url: '/knowledge/rules',
    method: 'get',
    params
  })
}

// 创建案例研究
export const createCaseStudy = (data) => {
  return service({
    url: '/knowledge/case',
    method: 'post',
    data
  })
}

// 更新案例研究
export const updateCaseStudy = (data) => {
  return service({
    url: `/knowledge/case/${data.id}`,
    method: 'put',
    data
  })
}

// 删除案例研究
export const deleteCaseStudy = (id) => {
  return service({
    url: `/knowledge/case/${id}`,
    method: 'delete'
  })
}

// 获取案例研究详情
export const getCaseStudy = (id) => {
  return service({
    url: `/knowledge/case/${id}`,
    method: 'get'
  })
}

// 获取案例研究列表
export const getCaseStudyList = (params) => {
  return service({
    url: '/knowledge/cases',
    method: 'get',
    params
  })
}

// 搜索案例研究
export const searchCaseStudies = (data) => {
  return service({
    url: '/knowledge/cases/search',
    method: 'post',
    data
  })
}

// 获取知识库统计
export const getKnowledgeStatistics = () => {
  return service({
    url: '/knowledge/statistics',
    method: 'get'
  })
}

// 导入NESMA标准知识
export const importNesmaStandardKnowledge = () => {
  return service({
    url: '/knowledge/import/standard',
    method: 'post'
  })
}

// 获取热门标签
export const getPopularTags = (params) => {
  return service({
    url: '/knowledge/tags/popular',
    method: 'get',
    params
  })
}

// 获取推荐知识
export const getRecommendedKnowledge = (params) => {
  return service({
    url: '/knowledge/recommended',
    method: 'get',
    params
  })
}

// ==================== 文档生成管理 ====================

// 创建文档生成记录
export const createDocument = (data) => {
  return service({
    url: '/nesma/document',
    method: 'post',
    data
  })
}

// 更新文档生成记录
export const updateDocument = (data) => {
  return service({
    url: '/nesma/document',
    method: 'put',
    data
  })
}

// 删除文档生成记录
export const deleteDocument = (id) => {
  return service({
    url: `/nesma/document/${id}`,
    method: 'delete'
  })
}

// 获取文档详情
export const getDocument = (id) => {
  return service({
    url: `/nesma/document/${id}`,
    method: 'get'
  })
}

// 获取文档列表
export const getDocumentList = (params) => {
  return service({
    url: '/nesma/document/list',
    method: 'get',
    params
  })
}

// 生成文档
export const generateDocument = (data) => {
  return service({
    url: '/nesma/document/generate',
    method: 'post',
    data
  })
}

// 批量生成文档
export const batchGenerateDocument = (data) => {
  return service({
    url: '/nesma/document/batch-generate',
    method: 'post',
    data
  })
}

// 预览文档
export const previewDocument = (data) => {
  return service({
    url: '/nesma/document/preview',
    method: 'post',
    data
  })
}

// 下载文档 - 使用独立的axios实例避免baseURL前缀问题
export const downloadDocument = (params) => {
  return axios({
    url: '/nesma/document/download',
    method: 'get',
    params,
    responseType: 'blob',
    timeout: 300000
  })
}

// 获取文档生成进度
export const getDocumentProgress = (params) => {
  return service({
    url: '/nesma/document/progress',
    method: 'get',
    params
  })
}

// 获取文档统计
export const getDocumentStats = (params) => {
  return service({
    url: '/nesma/document/stats',
    method: 'get',
    params
  })
}

// 批量删除文档
export const batchDeleteDocuments = (data) => {
  return service({
    url: '/nesma/document/batch-delete',
    method: 'post',
    data
  })
}

// ==================== 文档模板管理 ====================

// 创建模板
export const createTemplate = (data) => {
  return service({
    url: '/nesma/template',
    method: 'post',
    data
  })
}

// 更新模板
export const updateTemplate = (data) => {
  return service({
    url: '/nesma/template',
    method: 'put',
    data
  })
}

// 删除模板
export const deleteTemplate = (id) => {
  return service({
    url: `/nesma/template/${id}`,
    method: 'delete'
  })
}

// 获取模板详情
export const getTemplate = (id) => {
  return service({
    url: `/nesma/template/${id}`,
    method: 'get'
  })
}

// 获取模板列表
export const getTemplateList = (params) => {
  return service({
    url: '/nesma/template/list',
    method: 'get',
    params
  })
}

// 获取模板选项
export const getTemplateOptions = (params) => {
  return service({
    url: '/nesma/template/options',
    method: 'get',
    params
  })
}

// 获取模板变量
export const getTemplateVariables = (params) => {
  return service({
    url: '/nesma/template/variables',
    method: 'get',
    params
  })
}

// 设置默认模板
export const setDefaultTemplate = (id) => {
  return service({
    url: `/nesma/template/${id}/default`,
    method: 'post'
  })
}

// 激活/停用模板
export const activateTemplate = (id, data) => {
  return service({
    url: `/nesma/template/${id}/activate`,
    method: 'post',
    data
  })
}

// 上传模板文件
export const uploadTemplate = (data) => {
  return service({
    url: '/nesma/template/upload',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 批量删除模板
export const batchDeleteTemplates = (data) => {
  return service({
    url: '/nesma/template/batch-delete',
    method: 'post',
    data
  })
}

// ==================== NESMA评估管理 ====================

// 创建评估
export const createEvaluation = (data) => {
  return service({
    url: '/nesma/evaluation',
    method: 'post',
    data
  })
}

// 更新评估
export const updateEvaluation = (data) => {
  return service({
    url: '/nesma/evaluation',
    method: 'put',
    data
  })
}

// 删除评估
export const deleteEvaluation = (id) => {
  return service({
    url: `/nesma/evaluation/${id}`,
    method: 'delete'
  })
}

// 获取评估详情
export const getEvaluation = (id) => {
  return service({
    url: `/nesma/evaluation/${id}`,
    method: 'get'
  })
}

// 获取评估列表
export const getEvaluationList = (params) => {
  return service({
    url: '/nesma/evaluation/list',
    method: 'get',
    params
  })
}

// 开始评估
export const startEvaluation = (id) => {
  return service({
    url: `/nesma/evaluation/${id}/start`,
    method: 'post'
  })
}

// 停止评估
export const stopEvaluation = (id) => {
  return service({
    url: `/nesma/evaluation/${id}/stop`,
    method: 'post'
  })
}

// 审核评估
export const reviewEvaluation = (id, data) => {
  return service({
    url: `/nesma/evaluation/${id}/review`,
    method: 'post',
    data
  })
}

// 获取评估统计
export const getEvaluationStats = (params) => {
  return service({
    url: '/nesma/evaluation/stats',
    method: 'get',
    params
  })
}

// 获取评估摘要
export const getEvaluationSummary = (projectId) => {
  return service({
    url: `/nesma/evaluation/summary/${projectId}`,
    method: 'get'
  })
}

// ==================== 功能点管理 ====================

// 创建功能点
export const createFunctionPoint = (data) => {
  return service({
    url: '/nesma/function-point',
    method: 'post',
    data
  })
}

// 更新功能点
export const updateFunctionPoint = (data) => {
  return service({
    url: '/nesma/function-point',
    method: 'put',
    data
  })
}

// 删除功能点
export const deleteFunctionPoint = (id) => {
  return service({
    url: `/nesma/function-point/${id}`,
    method: 'delete'
  })
}

// 获取功能点详情
export const getFunctionPoint = (id) => {
  return service({
    url: `/nesma/function-point/${id}`,
    method: 'get'
  })
}

// 获取功能点列表
export const getFunctionPointList = (params) => {
  return service({
    url: '/nesma/function-point/list',
    method: 'get',
    params
  })
}

// 验证功能点
export const validateFunctionPoint = (id, data) => {
  return service({
    url: `/nesma/function-point/${id}/validate`,
    method: 'post',
    data
  })
}

// 批量更新功能点
export const batchUpdateFunctionPoints = (data) => {
  return service({
    url: '/nesma/function-point/batch-update',
    method: 'post',
    data
  })
}

// 批量删除功能点
export const batchDeleteFunctionPoints = (data) => {
  return service({
    url: '/nesma/function-point/batch-delete',
    method: 'post',
    data
  })
}

// ==================== 复杂度指标管理 ====================

// 获取复杂度指标
export const getComplexityMetrics = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/complexity-metrics`,
    method: 'get'
  })
}

// 更新复杂度指标
export const updateComplexityMetric = (data) => {
  return service({
    url: '/nesma/complexity-metric',
    method: 'put',
    data
  })
}

// 重新计算复杂度指标
export const recalculateComplexityMetrics = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/recalculate-metrics`,
    method: 'post'
  })
}

// ==================== 验证项管理 ====================

// 获取验证项列表
export const getValidationItems = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/validation-items`,
    method: 'get'
  })
}

// 更新验证项
export const updateValidationItem = (data) => {
  return service({
    url: '/nesma/validation-item',
    method: 'put',
    data
  })
}

// 解决验证项
export const resolveValidationItem = (id, data) => {
  return service({
    url: `/nesma/validation-item/${id}/resolve`,
    method: 'post',
    data
  })
}

// 重新执行验证检查
export const revalidateEvaluation = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/revalidate`,
    method: 'post'
  })
}

// ==================== 改进建议 ====================

// 获取改进建议
export const getRecommendations = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/recommendations`,
    method: 'get'
  })
}

// 标记建议为已处理
export const markRecommendationAsHandled = (evaluationId, recommendationIndex) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/recommendation/${recommendationIndex}/handled`,
    method: 'post'
  })
}

// 生成评估报告
export const generateEvaluationReport = (evaluationId, format = 'pdf') => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/report`,
    method: 'post',
    data: { format },
    responseType: 'blob'
  })
}

// 导出评估数据
export const exportEvaluationData = (evaluationId, format = 'excel') => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/export`,
    method: 'get',
    params: { format },
    responseType: 'blob'
  })
}

// ==================== 需求分析管理 ====================

// 启动需求分析（已迁移到统一智能分析接口 startIntelligentAnalysis）

// 获取分析进度（旧版本接口，已迁移到统一接口）
// export const getAnalysisProgress - 已在统一智能分析接口中定义

// 获取分析任务列表（旧版本接口）
export const getAnalysisTasksOld = (params) => {
  return service({
    url: '/nesma/analysis/tasks',
    method: 'get',
    params
  })
}

// 获取分析结果对比
export const getAnalysisComparison = (originalVersionId, analyzedVersionId) => {
  return service({
    url: '/nesma/analysis/comparison',
    method: 'get',
    params: {
      originalVersionId,
      analyzedVersionId
    }
  })
}

// 获取需求版本列表
export const getRequirementVersions = (cycleId) => {
  return service({
    url: `/nesma/requirement-version/cycle/${cycleId}`,
    method: 'get'
  })
}

// 切换需求版本
export const switchRequirementVersion = (versionId) => {
  return service({
    url: `/nesma/requirement-version/active/${versionId}`,
    method: 'put'
  })
}

// ==================== 需求版本管理 ====================

// 创建需求版本
export const createRequirementVersion = (data) => {
  return service({
    url: '/nesma/requirement-version',
    method: 'post',
    data
  })
}

// 获取版本详情
export const getRequirementVersion = (id) => {
  return service({
    url: `/nesma/requirement-version/${id}`,
    method: 'get'
  })
}

// 更新需求版本
export const updateRequirementVersion = (data) => {
  return service({
    url: '/nesma/requirement-version',
    method: 'put',
    data
  })
}

// 删除需求版本
export const deleteRequirementVersion = (id) => {
  return service({
    url: `/nesma/requirement-version/${id}`,
    method: 'delete'
  })
}

// 获取周期的激活版本
export const getActiveVersion = (cycleId) => {
  return service({
    url: `/nesma/requirement-version/active/${cycleId}`,
    method: 'get'
  })
}

// 设置激活版本
export const setActiveVersion = (id) => {
  return service({
    url: `/nesma/requirement-version/active/${id}`,
    method: 'put'
  })
}

// 生成下一个版本号
export const generateNextVersion = (cycleId, versionType = 'initial') => {
  return service({
    url: `/nesma/requirement-version/next/${cycleId}`,
    method: 'get',
    params: { versionType }
  })
}

// 更新版本统计信息
export const updateVersionStats = (id) => {
  return service({
    url: `/nesma/requirement-version/stats/${id}`,
    method: 'put'
  })
}

// ==================== 单个需求AI分析专用接口 ====================

// 启动单个需求AI分析
export const analyzeRequirement = (data) => {
  return service({
    url: '/nesma/requirement/ai-analysis',
    method: 'post',
    data
  })
}

// 获取需求分析进度
export const getRequirementAnalysisProgress = (taskId) => {
  return service({
    url: `/nesma/requirement/analysis-progress/${taskId}`,
    method: 'get'
  })
}

// 获取需求分析结果详情
export const getRequirementAnalysisResult = (taskId) => {
  return service({
    url: `/nesma/requirement/analysis-result/${taskId}`,
    method: 'get'
  })
}

// 应用AI分析结果
export const applyRequirementAnalysisResult = (data) => {
  return service({
    url: `/nesma/requirement/apply-analysis`,
    method: 'post',
    data
  })
}


// ==================== 统一智能分析接口（批量分析使用） ====================

// 获取需求树（用于智能分析）
export const getRequirementTree = (params) => {
  return service({
    url: '/nesma/intelligent-analysis/requirement/tree-enhanced',
    method: 'get',
    params
  })
}

// 启动智能分析（统一接口）
export const startIntelligentAnalysis = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/start',
    method: 'post',
    data
  })
}

// 获取分析进度（统一接口，支持新旧格式的taskId）
export const getAnalysisProgress = (taskId) => {
  return service({
    url: `/nesma/intelligent-analysis/progress/${taskId}`,
    method: 'get'
  })
}

// 获取分析结果详情
export const getAnalysisResult = (taskId) => {
  return service({
    url: `/nesma/intelligent-analysis/result/${taskId}`,
    method: 'get'
  })
}

// 获取分析建议
export const getAnalysisRecommendations = (taskId) => {
  return service({
    url: `/nesma/intelligent-analysis/recommendations/${taskId}`,
    method: 'get'
  })
}

// 应用分析建议
export const applyAnalysisRecommendation = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/apply-recommendation',
    method: 'post',
    data
  })
}

// 获取可用AI模型列表
export const getAvailableAIModels = () => {
  return service({
    url: '/nesma/intelligent-analysis/ai/models',
    method: 'get'
  })
}

// 测试AI模型连接
export const testAIModel = (modelConfig) => {
  return service({
    url: '/nesma/intelligent-analysis/ai/test',
    method: 'post',
    data: modelConfig
  })
}

// 获取AI服务状态
export const getAIServiceStatus = () => {
  return service({
    url: '/nesma/intelligent-analysis/ai/status',
    method: 'get'
  })
}

// 获取知识库搜索结果
export const searchKnowledgeBase = (query) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge/search',
    method: 'post',
    data: { query }
  })
}

// 获取知识图谱数据
export const getKnowledgeGraph = (params) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge/graph',
    method: 'get',
    params
  })
}

// 获取知识实体列表
export const getKnowledgeEntities = () => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge-entities',
    method: 'get'
  })
}

// 获取知识关系列表
export const getKnowledgeRelations = () => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge-relations',
    method: 'get'
  })
}

// 创建知识实体
export const createKnowledgeEntity = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge-entities',
    method: 'post',
    data
  })
}

// 更新知识图谱节点
export const updateKnowledgeNode = (data) => {
  return service({
    url: '/nesma/knowledge/node',
    method: 'put',
    data
  })
}

// 创建知识图谱关系
export const createKnowledgeRelation = (data) => {
  return service({
    url: '/nesma/knowledge/relation',
    method: 'post',
    data
  })
}

// 获取向量搜索结果
export const vectorSearch = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/vector/search',
    method: 'post',
    data
  })
}

// 导出智能分析报告
export const exportIntelligentAnalysisReport = (params) => {
  return service({
    url: '/nesma/intelligent-analysis/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// ==================== 统一AI分析接口 (新增) ====================

// 执行统一分析
export const executeUnifiedAnalysis = (data) => {
  return service({
    url: '/nesma/unified-analysis/execute',
    method: 'post',
    data
  })
}

// 获取分析进度
export const getUnifiedAnalysisProgress = (taskId) => {
  return service({
    url: '/nesma/unified-analysis/progress',
    method: 'get',
    params: { task_id: taskId }
  })
}

// 获取分析结果
export const getUnifiedAnalysisResult = (taskId) => {
  return service({
    url: '/nesma/unified-analysis/result',
    method: 'get',
    params: { task_id: taskId }
  })
}

// 获取详细分析结果
export const getDetailedAnalysisResult = (taskId, detailLevel = 'standard') => {
  return service({
    url: '/nesma/unified-analysis/detailed-result',
    method: 'get',
    params: { 
      task_id: taskId,
      detail_level: detailLevel 
    }
  })
}

// 取消分析
export const cancelUnifiedAnalysis = (data) => {
  return service({
    url: '/nesma/unified-analysis/cancel',
    method: 'post',
    data
  })
}

// 项目一键分析
// export const executeProjectAnalysis = (data) => {
//   return service({
//     url: '/nesma/project/one-click-analysis',
//     method: 'post',
//     data
//   })
// }

export const executeProjectAnalysis = (data) => {
  return service({
    url: '/nesma/unified-analysis/project-analysis',
    method: 'post',
    data
  })
}

// 需求优化分析
export const executeRequirementOptimization = (data) => {
  return service({
    url: '/nesma/unified-analysis/requirement-optimize',
    method: 'post',
    data
  })
}

// NESMA评估
export const executeNESMAEvaluation = (data) => {
  return service({
    url: '/nesma/unified-analysis/nesma-evaluation',
    method: 'post',
    data
  })
}

// 获取优化建议
export const getOptimizationSuggestions = (params) => {
  return service({
    url: '/nesma/unified-analysis/optimization-suggestions',
    method: 'get',
    params
  })
}

// 应用优化建议
export const applyOptimization = (data) => {
  return service({
    url: '/nesma/unified-analysis/apply-optimization',
    method: 'post',
    data
  })
}

// 批量应用优化建议
export const batchApplyOptimizations = (data) => {
  return service({
    url: '/nesma/unified-analysis/batch-apply-optimizations',
    method: 'post',
    data
  })
}

// 获取需求优化详情
export const getRequirementOptimizationDetail = (taskId, requirementId) => {
  return service({
    url: '/nesma/unified-analysis/requirement-optimization-detail',
    method: 'get',
    params: { task_id: taskId, requirement_id: requirementId }
  })
}

// 导出分析报告
export const exportUnifiedAnalysisReport = (data) => {
  return service({
    url: '/nesma/unified-analysis/export',
    method: 'post',
    data
  })
}

// ==================== 完善的报告生成接口 ====================

// 获取报告模板列表
export const getReportTemplates = () => {
  return service({
    url: '/nesma/report/templates',
    method: 'get'
  })
}

// 创建报告模板
export const createReportTemplate = (data) => {
  return service({
    url: '/nesma/report/templates',
    method: 'post',
    data
  })
}

// 更新报告模板
export const updateReportTemplate = (data) => {
  return service({
    url: '/nesma/report/templates',
    method: 'put',
    data
  })
}

// 删除报告模板
export const deleteReportTemplate = (id) => {
  return service({
    url: `/nesma/report/templates/${id}`,
    method: 'delete'
  })
}

// 复制报告模板
export const duplicateReportTemplate = (id) => {
  return service({
    url: `/nesma/report/templates/${id}/duplicate`,
    method: 'post'
  })
}

// 生成报告预览
export const generateReportPreview = (data) => {
  return service({
    url: '/nesma/report/preview',
    method: 'post',
    data
  })
}

// 获取报告生成进度
export const getReportGenerationProgress = (generationId) => {
  return service({
    url: `/nesma/report/generation/${generationId}/progress`,
    method: 'get'
  })
}

// 取消报告生成
export const cancelReportGeneration = (generationId) => {
  return service({
    url: `/nesma/report/generation/${generationId}/cancel`,
    method: 'post'
  })
}

// 批量报告生成
export const startBatchReportGeneration = (data) => {
  return service({
    url: '/nesma/report/batch/start',
    method: 'post',
    data
  })
}

// 获取批量生成进度
export const getBatchGenerationProgress = (batchId) => {
  return service({
    url: `/nesma/report/batch/${batchId}/progress`,
    method: 'get'
  })
}

// 取消批量生成
export const cancelBatchGeneration = (batchId) => {
  return service({
    url: `/nesma/report/batch/${batchId}/cancel`,
    method: 'post'
  })
}

// 获取报告生成历史
export const getReportGenerationHistory = (params) => {
  return service({
    url: '/nesma/report/history',
    method: 'get',
    params
  })
}

// 获取报告生成统计
export const getReportGenerationStats = (params) => {
  return service({
    url: '/nesma/report/stats',
    method: 'get',
    params
  })
}

// 下载报告
export const downloadGeneratedReport = (generationId) => {
  return service({
    url: `/nesma/report/download/${generationId}`,
    method: 'get',
    responseType: 'blob'
  })
}

// 报告调度管理
export const createReportSchedule = (data) => {
  return service({
    url: '/nesma/report/schedule',
    method: 'post',
    data
  })
}

// 获取报告调度列表
export const getReportSchedules = (params) => {
  return service({
    url: '/nesma/report/schedules',
    method: 'get',
    params
  })
}

// 更新报告调度
export const updateReportSchedule = (data) => {
  return service({
    url: '/nesma/report/schedule',
    method: 'put',
    data
  })
}

// 删除报告调度
export const deleteReportSchedule = (id) => {
  return service({
    url: `/nesma/report/schedule/${id}`,
    method: 'delete'
  })
}

// 启用/禁用报告调度
export const toggleReportSchedule = (id, enabled) => {
  return service({
    url: `/nesma/report/schedule/${id}/toggle`,
    method: 'post',
    data: { enabled }
  })
}

// 获取分析历史
export const getUnifiedAnalysisHistory = (params) => {
  return service({
    url: '/nesma/unified-analysis/history',
    method: 'get',
    params
  })
}

// ==================== 任务管理接口 (新增) ====================

// 获取分析任务列表
export const getAnalysisTaskList = (params) => {
  return service({
    url: '/nesma/analysis/tasks',
    method: 'get',
    params
  })
}

// 获取项目的分析任务列表
export const getProjectAnalysisTasks = (projectId, params = {}) => {
  return service({
    url: `/nesma/analysis/tasks`,
    method: 'get',
    params: {
      projectId: projectId,
      ...params
    }
  })
}

// 获取任务详情
export const getAnalysisTaskDetail = (taskId) => {
  return service({
    url: `/nesma/analysis/task/${taskId}`,
    method: 'get'
  })
}

// 恢复任务进度查看
export const resumeTaskProgress = (taskId) => {
  return service({
    url: `/nesma/analysis/task/${taskId}/resume`,
    method: 'post'
  })
}

// 获取任务执行日志
export const getTaskExecutionLogs = (taskId, params = {}) => {
  return service({
    url: `/nesma/analysis/task/${taskId}/logs`,
    method: 'get',
    params
  })
}

// 删除任务
export const deleteAnalysisTask = (taskId) => {
  return service({
    url: `/nesma/analysis/tasks/${taskId}`,
    method: 'delete'
  })
}

// 批量删除任务
export const batchDeleteAnalysisTasks = (data) => {
  return service({
    url: '/nesma/analysis/tasks/batch-delete',
    method: 'post',
    data
  })
}

// 重新执行任务
export const retryAnalysisTask = (taskId) => {
  return service({
    url: `/nesma/analysis/task/${taskId}/retry`,
    method: 'post'
  })
}

// 获取任务统计信息
export const getTaskStatistics = (params = {}) => {
  return service({
    url: '/nesma/analysis/tasks/statistics',
    method: 'get',
    params
  })
}

// ==================== 统一工作流接口 (新增) ====================

// 执行完整分析工作流
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

// 获取工作流执行历史
export const getWorkflowExecutionHistory = (projectId, params = {}) => {
  return service({
    url: `/nesma/workflow/history/${projectId}`,
    method: 'get',
    params
  })
}

// 获取工作流执行统计
export const getWorkflowExecutionStats = (projectId, params = {}) => {
  return service({
    url: `/nesma/workflow/stats/${projectId}`,
    method: 'get',
    params
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

// 获取工作流模板
export const getWorkflowTemplates = () => {
  return service({
    url: '/nesma/workflow/templates',
    method: 'get'
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

// ==================== 兼容性API别名 ====================

// 为了保持向后兼容，提供别名
export const analyzeProject = executeProjectAnalysis
export const optimizeRequirement = executeRequirementOptimization
export const evaluateNESMA = executeNESMAEvaluation
export const startWorkflow = executeFullAnalysisWorkflow
export const getWorkflowStatus = getWorkflowExecutionStatus

// 兼容旧的intelligent-analysis接口
export const startIntelligentAnalysisCompat = startIntelligentAnalysis
export const getAnalysisProgressCompat = getAnalysisProgress
export const getAnalysisResultCompat = getAnalysisResult

// ==================== 兼容性接口（旧版本支持） ====================

// 注意：以下函数已在上方统一智能分析接口中重新定义，这些是兼容性声明
// 启动需求分析（兼容旧接口）- 已迁移到 startIntelligentAnalysis

// 获取分析任务列表（兼容旧接口）- 功能保持独立
export const getAnalysisTasksLegacy = (params) => {
  return service({
    url: '/nesma/analysis/tasks',
    method: 'get',
    params
  })
}

// ==================== 选项获取接口 ====================

// 获取项目选项（用于下拉框）
export const getProjectOptions = () => {
  return service({
    url: '/nesma/project/options',
    method: 'get'
  })
}

// 获取用户选项（用于下拉框）
export const getUserOptions = () => {
  return service({
    url: '/nesma/user/options',
    method: 'get'
  })
}

// ==================== 通用列表和删除接口 ====================

// 获取知识列表（兼容接口）
export const getKnowledgeList = (params) => {
  return getKnowledgeEntryList(params)
}

// 删除知识（兼容接口）
export const deleteKnowledge = (id) => {
  return deleteKnowledgeEntry(id)
}

// 创建知识（兼容接口）
export const createKnowledge = (data) => {
  return createKnowledgeEntry(data)
}

// 更新知识（兼容接口）
export const updateKnowledge = (data) => {
  return updateKnowledgeEntry(data)
}

// ==================== Mermaid流程图生成器 ====================

// 生成Mermaid流程图
export const generateMermaidDiagrams = (data) => {
  return service({
    url: '/nesma/generator/mermaid/generate',
    method: 'post',
    data
  })
}

// 异步生成Mermaid流程图
export const generateMermaidDiagramsAsync = (data) => {
  return service({
    url: '/nesma/generator/mermaid/generate-async',
    method: 'post',
    data
  })
}

// 获取Mermaid生成任务进度
export const getMermaidGenerationTaskProgress = (taskId) => {
  return service({
    url: `/nesma/generator/mermaid/task/${taskId}/progress`,
    method: 'get'
  })
}

// 获取Mermaid生成任务结果
export const getMermaidGenerationTaskResult = (taskId) => {
  return service({
    url: `/nesma/generator/mermaid/task/${taskId}/result`,
    method: 'get'
  })
}

// 批量生成Mermaid流程图
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

// 验证Mermaid流程图
export const validateMermaidDiagram = (data) => {
  return service({
    url: '/nesma/generator/mermaid/validate',
    method: 'post',
    data
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

// 获取Mermaid生成历史
export const getMermaidGenerationHistory = (cycleId) => {
  return service({
    url: `/nesma/generator/mermaid/history/${cycleId}`,
    method: 'get'
  })
}

// 获取Mermaid生成统计
export const getMermaidGenerationStats = (cycleId) => {
  return service({
    url: `/nesma/generator/mermaid/stats/${cycleId}`,
    method: 'get'
  })
}

// 获取Mermaid生成任务列表
export const getMermaidGenerationTasks = (params) => {
  return service({
    url: '/nesma/generator/mermaid/tasks',
    method: 'get',
    params
  })
}

// 获取Mermaid任务统计信息
export const getMermaidTaskStatistics = (params) => {
  return service({
    url: '/nesma/generator/mermaid/task-statistics',
    method: 'get',
    params
  })
}

// 获取Mermaid任务详情
export const getMermaidTaskDetail = (taskId) => {
  return service({
    url: `/nesma/generator/mermaid/task/${taskId}`,
    method: 'get'
  })
}

// 取消Mermaid生成任务
export const cancelMermaidGenerationTask = (taskId) => {
  return service({
    url: `/nesma/generator/mermaid/task/${taskId}/cancel`,
    method: 'post'
  })
}

// 重试Mermaid生成任务
export const retryMermaidGenerationTask = (taskId) => {
  return service({
    url: `/nesma/generator/mermaid/task/${taskId}/retry`,
    method: 'post'
  })
}

// 删除Mermaid生成任务
export const deleteMermaidGenerationTask = (taskId) => {
  return service({
    url: `/nesma/generator/mermaid/task/${taskId}`,
    method: 'delete'
  })
}

// 批量删除Mermaid生成任务
export const batchDeleteMermaidGenerationTasks = (data) => {
  return service({
    url: '/nesma/generator/mermaid/tasks/batch-delete',
    method: 'post',
    data
  })
}

// ==================== L4任务管理接口 ====================

// 获取L4生成任务列表
export const getL4GenerationTasks = (params) => {
  return service({
    url: '/nesma/generator/level4/tasks',
    method: 'get',
    params
  })
}

// 获取L4任务统计信息
export const getL4TaskStatistics = (params) => {
  return service({
    url: '/nesma/generator/level4/tasks/statistics',
    method: 'get',
    params
  })
}

// 获取L4任务详情
export const getL4TaskDetail = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/detail`,
    method: 'get'
  })
}

// 取消L4生成任务
export const cancelL4GenerationTask = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/cancel`,
    method: 'post'
  })
}

// 重试L4生成任务
export const retryL4GenerationTask = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/retry`,
    method: 'post'
  })
}

// 删除L4生成任务
export const deleteL4GenerationTask = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}`,
    method: 'delete'
  })
}

// 批量删除L4生成任务
export const batchDeleteL4GenerationTasks = (data) => {
  return service({
    url: '/nesma/generator/level4/tasks/batch-delete',
    method: 'post',
    data
  })
}

// 批量取消L4生成任务
export const batchCancelL4GenerationTasks = (data) => {
  return service({
    url: '/nesma/generator/level4/tasks/batch-cancel',
    method: 'post',
    data
  })
}

// 获取L4任务实时进度（支持轮询）
export const getL4TaskProgress = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/progress`,
    method: 'get'
  })
}

// 获取L4任务执行日志
export const getL4TaskLogs = (taskId, params = {}) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/logs`,
    method: 'get',
    params
  })
}

// 生成L4功能点
export const generateLevel4Requirements = (data) => {
  return service({
    url: '/nesma/generator/level4/generate',
    method: 'post',
    data
  })
}

// 创建L4功能点
export const createLevel4Requirement = (data) => {
  return service({
    url: '/nesma/generator/level4/create',
    method: 'post',
    data
  })
}

// 批量创建L4功能点
export const batchCreateLevel4Requirements = (data) => {
  return service({
    url: '/nesma/generator/level4/batch-create',
    method: 'post',
    data
  })
}

// 验证L4功能点建议
export const validateLevel4Suggestion = (data) => {
  return service({
    url: '/nesma/generator/level4/validate',
    method: 'post',
    data
  })
}

// ==================== 异步L4生成API ====================

// 异步生成L4功能点
export const generateLevel4Async = (data) => {
  return service({
    url: '/nesma/generator/level4/generate-async',
    method: 'post',
    data
  })
}

// 获取L4生成任务进度
export const getLevel4GenerationTaskProgress = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/progress`,
    method: 'get'
  })
}

// 获取L4生成任务结果
export const getLevel4GenerationTaskResult = (taskId) => {
  return service({
    url: `/nesma/generator/level4/task/${taskId}/result`,
    method: 'get'
  })
}

// 获取L4生成历史
export const getLevel4GenerationHistory = (cycleId) => {
  return service({
    url: `/nesma/generator/level4/history/${cycleId}`,
    method: 'get'
  })
}

// 获取L4生成统计
export const getLevel4GenerationStats = (cycleId) => {
  return service({
    url: `/nesma/generator/level4/stats/${cycleId}`,
    method: 'get'
  })
}

// ==================== 批量分析接口 ====================

// 启动批量分析
export const startBatchAnalysis = (data) => {
  return service({
    url: '/nesma/analysis/batch-start',
    method: 'post',
    data
  })
}

// 获取批量分析进度
export const getBatchAnalysisProgress = (batchId) => {
  return service({
    url: `/nesma/analysis/batch-progress/${batchId}`,
    method: 'get'
  })
}

// 取消批量分析
export const cancelBatchAnalysis = (batchId) => {
  return service({
    url: `/nesma/analysis/batch-cancel/${batchId}`,
    method: 'post'
  })
}

// 获取批量分析结果
export const getBatchAnalysisResult = (batchId) => {
  return service({
    url: `/nesma/analysis/batch-result/${batchId}`,
    method: 'get'
  })
}

// ==================== 系统活动和监控 ====================

// ==================== L3需求分析器 ====================

// 分析L3需求
export const analyzeLevel3Requirements = (data) => {
  return service({
    url: '/nesma/analyzer/level3/analyze',
    method: 'post',
    data
  })
}

// 应用优化建议
export const applyOptimizationSuggestion = (data) => {
  return service({
    url: '/nesma/analyzer/level3/apply-optimization',
    method: 'post',
    data
  })
}

// 创建扩充功能点
export const createExpansionRequirement = (data) => {
  return service({
    url: '/nesma/analyzer/level3/create-expansion',
    method: 'post',
    data
  })
}

// 批量应用优化建议
export const batchApplyLevel3Optimizations = (data) => {
  return service({
    url: '/nesma/analyzer/level3/batch-apply-optimizations',
    method: 'post',
    data
  })
}

// 批量创建扩充功能点
export const batchCreateExpansions = (data) => {
  return service({
    url: '/nesma/analyzer/level3/batch-create-expansions',
    method: 'post',
    data
  })
}

// 获取L3分析历史
export const getLevel3AnalysisHistory = (cycleId) => {
  return service({
    url: `/nesma/analyzer/level3/history/${cycleId}`,
    method: 'get'
  })
}

// 获取L3分析统计
export const getLevel3AnalysisStats = (cycleId) => {
  return service({
    url: `/nesma/analyzer/level3/stats/${cycleId}`,
    method: 'get'
  })
}

// ==================== 工作台统计 ====================

// 获取工作台统计数据
export const getDashboardStats = (params) => {
  return service({
    url: '/dashboard/stats',
    method: 'get',
    params
  })
}

// 清除工作台缓存
export const clearDashboardCache = () => {
  return service({
    url: '/dashboard/cache',
    method: 'delete'
  })
}

// 获取系统活动列表
export const getSystemActivities = (params) => {
  return service({
    url: '/dashboard/activities',
    method: 'get',
    params
  })
}

// 创建系统活动记录
export const createSystemActivity = (data) => {
  return service({
    url: '/system/activities',
    method: 'post',
    data
  })
}

// ==================== 系统状态监控 ====================

// 获取系统状态
export const getSystemStatus = () => {
  return service({
    url: '/system/status',
    method: 'get'
  })
}

// 获取系统健康检查
export const getSystemHealth = () => {
  return service({
    url: '/system/health',
    method: 'get'
  })
}

// 获取系统性能指标
export const getSystemPerformance = () => {
  return service({
    url: '/system/performance',
    method: 'get'
  })
}

