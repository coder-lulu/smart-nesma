import service from '@/utils/request'

// ==================== 统一AI服务API ====================

// ==================== AI对话服务 ====================

// 创建会话
export const createChatSession = (data) => {
  return service({
    url: '/api/v1/ai/chat/sessions',
    method: 'post',
    data
  })
}

// 获取会话列表
export const getChatSessionList = (params) => {
  return service({
    url: '/api/v1/ai/chat/sessions',
    method: 'get',
    params
  })
}

// 获取会话详情
export const getChatSession = (sessionId) => {
  return service({
    url: `/api/v1/ai/chat/sessions/${sessionId}`,
    method: 'get'
  })
}

// 更新会话
export const updateChatSession = (sessionId, data) => {
  return service({
    url: `/api/v1/ai/chat/sessions/${sessionId}`,
    method: 'put',
    data
  })
}

// 删除会话
export const deleteChatSession = (sessionId) => {
  return service({
    url: `/api/v1/ai/chat/sessions/${sessionId}`,
    method: 'delete'
  })
}

// 发送消息
export const sendMessage = (data) => {
  return service({
    url: '/api/v1/ai/chat/messages',
    method: 'post',
    data
  })
}

// 获取消息历史
export const getMessageHistory = (sessionId, params) => {
  return service({
    url: `/api/v1/ai/chat/sessions/${sessionId}/messages`,
    method: 'get',
    params
  })
}

// 导出会话
export const exportChatSession = (sessionId, format = 'json') => {
  return service({
    url: `/api/v1/ai/chat/sessions/${sessionId}/export`,
    method: 'get',
    params: { format },
    responseType: 'blob'
  })
}

// ==================== AI分析服务 ====================

// 启动需求分析
export const startRequirementAnalysis = (data) => {
  return service({
    url: '/api/v1/ai/analysis/requirements',
    method: 'post',
    data
  })
}

// 批量分析需求
export const batchAnalyzeRequirements = (data) => {
  return service({
    url: '/api/v1/ai/analysis/requirements/batch',
    method: 'post',
    data
  })
}

// 获取分析进度
export const getAnalysisProgress = (taskId) => {
  return service({
    url: `/api/v1/ai/analysis/tasks/${taskId}/progress`,
    method: 'get'
  })
}

// 获取分析结果
export const getAnalysisResult = (taskId) => {
  return service({
    url: `/api/v1/ai/analysis/tasks/${taskId}/result`,
    method: 'get'
  })
}

// 取消分析任务
export const cancelAnalysisTask = (taskId) => {
  return service({
    url: `/api/v1/ai/analysis/tasks/${taskId}/cancel`,
    method: 'post'
  })
}

// 获取分析任务列表
export const getAnalysisTasks = (params) => {
  return service({
    url: '/api/v1/ai/analysis/tasks',
    method: 'get',
    params
  })
}

// 智能分析工作区状态
export const getWorkspaceStatus = () => {
  return service({
    url: '/api/v1/ai/analysis/workspace/status',
    method: 'get'
  })
}

// 获取分析统计信息
export const getAnalysisStats = (params) => {
  return service({
    url: '/api/v1/ai/analysis/stats',
    method: 'get',
    params
  })
}

// ==================== AI生成服务 ====================

// 通用内容生成
export const generateContent = (data) => {
  return service({
    url: '/api/v1/ai/generation/content',
    method: 'post',
    data
  })
}

// 生成需求描述
export const generateDescription = (data) => {
  return service({
    url: '/api/v1/ai/generation/description',
    method: 'post',
    data
  })
}

// 生成Mermaid图表
export const generateMermaidChart = (data) => {
  return service({
    url: '/api/v1/ai/generation/mermaid',
    method: 'post',
    data
  })
}

// 生成L4级需求
export const generateLevel4Requirements = (data) => {
  return service({
    url: '/api/v1/ai/generation/level4',
    method: 'post',
    data
  })
}

// 批量生成内容
export const batchGenerateContent = (data) => {
  return service({
    url: '/api/v1/ai/generation/batch',
    method: 'post',
    data
  })
}

// 获取生成任务状态
export const getGenerationTaskStatus = (taskId) => {
  return service({
    url: `/api/v1/ai/generation/tasks/${taskId}/status`,
    method: 'get'
  })
}

// 获取生成模板
export const getGenerationTemplates = (type) => {
  return service({
    url: '/api/v1/ai/generation/templates',
    method: 'get',
    params: { type }
  })
}

// ==================== AI推荐服务 ====================

// 获取AI智能推荐
export const getAIRecommendations = (data) => {
  return service({
    url: '/api/v1/ai/recommendation/get',
    method: 'post',
    data
  })
}

// 获取个性化推荐
export const getPersonalizedRecommendations = (userId, params) => {
  return service({
    url: `/api/v1/ai/recommendation/personal/${userId}`,
    method: 'get',
    params
  })
}

// 采纳AI推荐
export const adoptRecommendation = (data) => {
  return service({
    url: '/api/v1/ai/recommendation/adopt',
    method: 'post',
    data
  })
}

// 拒绝AI推荐
export const rejectRecommendation = (data) => {
  return service({
    url: '/api/v1/ai/recommendation/reject',
    method: 'post',
    data
  })
}

// 获取推荐反馈
export const getRecommendationFeedback = (recommendationId) => {
  return service({
    url: `/api/v1/ai/recommendation/${recommendationId}/feedback`,
    method: 'get'
  })
}

// 提交推荐反馈
export const submitRecommendationFeedback = (data) => {
  return service({
    url: '/api/v1/ai/recommendation/feedback',
    method: 'post',
    data
  })
}

// 获取推荐统计
export const getRecommendationStats = (params) => {
  return service({
    url: '/api/v1/ai/recommendation/stats',
    method: 'get',
    params
  })
}

// ==================== AI服务管理 ====================

// 获取AI服务状态
export const getAIServiceStatus = () => {
  return service({
    url: '/api/v1/ai/service/status',
    method: 'get'
  })
}

// 测试AI服务连通性
export const testAIServiceConnection = (serviceType) => {
  return service({
    url: '/api/v1/ai/service/test',
    method: 'post',
    data: { serviceType }
  })
}

// 获取AI模型列表
export const getAIModels = () => {
  return service({
    url: '/api/v1/ai/service/models',
    method: 'get'
  })
}

// 切换AI模型
export const switchAIModel = (modelId) => {
  return service({
    url: '/api/v1/ai/service/models/switch',
    method: 'post',
    data: { modelId }
  })
}

// 获取AI服务配置
export const getAIServiceConfig = () => {
  return service({
    url: '/api/v1/ai/service/config',
    method: 'get'
  })
}

// 更新AI服务配置
export const updateAIServiceConfig = (data) => {
  return service({
    url: '/api/v1/ai/service/config',
    method: 'put',
    data
  })
}

// 获取AI服务使用统计
export const getAIServiceUsageStats = (params) => {
  return service({
    url: '/api/v1/ai/service/usage/stats',
    method: 'get',
    params
  })
}

// 获取AI服务性能指标
export const getAIServiceMetrics = (params) => {
  return service({
    url: '/api/v1/ai/service/metrics',
    method: 'get',
    params
  })
}

// ==================== 知识图谱服务 ====================

// 获取知识图谱数据
export const getKnowledgeGraphData = (params) => {
  return service({
    url: '/api/v1/ai/knowledge-graph/data',
    method: 'get',
    params
  })
}

// 搜索知识节点
export const searchKnowledgeNodes = (query) => {
  return service({
    url: '/api/v1/ai/knowledge-graph/search',
    method: 'get',
    params: { query }
  })
}

// 获取节点关系
export const getNodeRelations = (nodeId) => {
  return service({
    url: `/api/v1/ai/knowledge-graph/nodes/${nodeId}/relations`,
    method: 'get'
  })
}

// 更新知识图谱
export const updateKnowledgeGraph = (data) => {
  return service({
    url: '/api/v1/ai/knowledge-graph/update',
    method: 'post',
    data
  })
}

// ==================== WebSocket连接 ====================

// 建立AI服务WebSocket连接
export const createAIWebSocketConnection = (sessionId) => {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${wsProtocol}//${window.location.host}/api/v1/ai/chat/ws/${sessionId}`
  return new WebSocket(wsUrl)
}

// 建立分析进度WebSocket连接
export const createAnalysisWebSocketConnection = (taskId) => {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const wsUrl = `${wsProtocol}//${window.location.host}/api/v1/ai/analysis/ws/${taskId}`
  return new WebSocket(wsUrl)
}