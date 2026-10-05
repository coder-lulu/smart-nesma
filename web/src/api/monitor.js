// 统一监控API接口
import request from '@/utils/request'

// 系统概览相关接口
export const getSystemOverview = () => {
  return request({
    url: '/api/v1/monitor/overview',
    method: 'get'
  })
}

export const getProjectStats = () => {
  return request({
    url: '/api/v1/monitor/project-stats',
    method: 'get'
  })
}

export const getRequirementStats = () => {
  return request({
    url: '/api/v1/monitor/requirement-stats',
    method: 'get'
  })
}

export const getSystemStats = () => {
  return request({
    url: '/api/v1/monitor/system-stats',
    method: 'get'
  })
}

// AI服务监控相关接口
export const getAIServiceStats = () => {
  return request({
    url: '/api/v1/monitor/ai-service-stats',
    method: 'get'
  })
}

export const getAIServices = () => {
  return request({
    url: '/api/v1/monitor/ai-services',
    method: 'get'
  })
}

export const getAIServiceAlerts = () => {
  return request({
    url: '/api/v1/monitor/ai-service-alerts',
    method: 'get'
  })
}

export const testAIService = (serviceId) => {
  return request({
    url: `/api/v1/monitor/ai-service/${serviceId}/test`,
    method: 'post'
  })
}

export const restartAIService = (serviceId) => {
  return request({
    url: `/api/v1/monitor/ai-service/${serviceId}/restart`,
    method: 'post'
  })
}

export const getAIServiceLogs = (serviceId) => {
  return request({
    url: `/api/v1/monitor/ai-service/${serviceId}/logs`,
    method: 'get'
  })
}

export const getAIServiceConfig = (serviceId) => {
  return request({
    url: `/api/v1/monitor/ai-service/${serviceId}/config`,
    method: 'get'
  })
}

// 性能监控相关接口
export const getPerformanceStats = () => {
  return request({
    url: '/api/v1/monitor/performance-stats',
    method: 'get'
  })
}

export const getAPIStats = (params) => {
  return request({
    url: '/api/v1/monitor/api-stats',
    method: 'get',
    params
  })
}

export const getResourceStats = () => {
  return request({
    url: '/api/v1/monitor/resource-stats',
    method: 'get'
  })
}

export const getPerformanceData = (params) => {
  return request({
    url: '/api/v1/monitor/performance-data',
    method: 'get',
    params
  })
}

// Agent监控相关接口
export const getAgentStats = () => {
  return request({
    url: '/api/v1/monitor/agent-stats',
    method: 'get'
  })
}

export const getAgents = () => {
  return request({
    url: '/api/v1/monitor/agents',
    method: 'get'
  })
}

export const getTaskQueue = () => {
  return request({
    url: '/api/v1/monitor/task-queue',
    method: 'get'
  })
}

export const startAgent = (agentId) => {
  return request({
    url: `/api/v1/monitor/agent/${agentId}/start`,
    method: 'post'
  })
}

export const stopAgent = (agentId) => {
  return request({
    url: `/api/v1/monitor/agent/${agentId}/stop`,
    method: 'post'
  })
}

export const restartAgent = (agentId) => {
  return request({
    url: `/api/v1/monitor/agent/${agentId}/restart`,
    method: 'post'
  })
}

export const getAgentLogs = (agentId) => {
  return request({
    url: `/api/v1/monitor/agent/${agentId}/logs`,
    method: 'get'
  })
}

export const getAgentConfig = (agentId) => {
  return request({
    url: `/api/v1/monitor/agent/${agentId}/config`,
    method: 'get'
  })
}

export const cancelTask = (taskId) => {
  return request({
    url: `/api/v1/monitor/task/${taskId}/cancel`,
    method: 'post'
  })
}

export const getTaskDetail = (taskId) => {
  return request({
    url: `/api/v1/monitor/task/${taskId}`,
    method: 'get'
  })
}

// 实时日志相关接口
export const getLogs = (params) => {
  return request({
    url: '/api/v1/monitor/logs',
    method: 'get',
    params
  })
}

export const getLogStats = () => {
  return request({
    url: '/api/v1/monitor/log-stats',
    method: 'get'
  })
}

export const clearLogs = () => {
  return request({
    url: '/api/v1/monitor/logs/clear',
    method: 'delete'
  })
}

export const exportLogs = (params) => {
  return request({
    url: '/api/v1/monitor/logs/export',
    method: 'post',
    data: params,
    responseType: 'blob'
  })
}

// WebSocket日志订阅
export const createLogWebSocket = (callback) => {
  const wsUrl = `${window.location.protocol === 'https:' ? 'wss:' : 'ws:'}//${window.location.host}/api/v1/monitor/logs/ws`
  const ws = new WebSocket(wsUrl)
  
  ws.onmessage = (event) => {
    try {
      const logData = JSON.parse(event.data)
      callback(logData)
    } catch (error) {
      console.error('解析日志数据失败:', error)
    }
  }
  
  ws.onerror = (error) => {
    console.error('WebSocket连接错误:', error)
  }
  
  return ws
}

// 监控设置相关接口
export const getMonitorSettings = () => {
  return request({
    url: '/api/v1/monitor/settings',
    method: 'get'
  })
}

export const saveMonitorSettings = (settings) => {
  return request({
    url: '/api/v1/monitor/settings',
    method: 'post',
    data: settings
  })
}

export const getAlertRules = () => {
  return request({
    url: '/api/v1/monitor/alert-rules',
    method: 'get'
  })
}

export const saveAlertRule = (rule) => {
  return request({
    url: '/api/v1/monitor/alert-rules',
    method: 'post',
    data: rule
  })
}

export const updateAlertRule = (ruleId, rule) => {
  return request({
    url: `/api/v1/monitor/alert-rules/${ruleId}`,
    method: 'put',
    data: rule
  })
}

export const deleteAlertRule = (ruleId) => {
  return request({
    url: `/api/v1/monitor/alert-rules/${ruleId}`,
    method: 'delete'
  })
}

// 数据导出相关接口
export const exportMonitorReport = (params) => {
  return request({
    url: '/api/v1/monitor/export/report',
    method: 'post',
    data: params,
    responseType: 'blob'
  })
}

export const exportPerformanceData = (params) => {
  return request({
    url: '/api/v1/monitor/export/performance',
    method: 'post',
    data: params,
    responseType: 'blob'
  })
}

// 健康检查接口
export const healthCheck = () => {
  return request({
    url: '/api/v1/monitor/health',
    method: 'get',
    timeout: 5000
  })
}

// 实时数据刷新
export const refreshAllData = () => {
  return Promise.all([
    getSystemOverview(),
    getAIServiceStats(),
    getPerformanceStats(),
    getAgentStats(),
    getLogStats()
  ])
}

// 批量操作接口
export const batchOperateAgents = (operation, agentIds) => {
  return request({
    url: '/api/v1/monitor/agents/batch',
    method: 'post',
    data: {
      operation,
      agentIds
    }
  })
}

export const batchOperateTasks = (operation, taskIds) => {
  return request({
    url: '/api/v1/monitor/tasks/batch',
    method: 'post',
    data: {
      operation,
      taskIds
    }
  })
}

// 告警处理接口
export const resolveAlert = (alertId) => {
  return request({
    url: `/api/v1/monitor/alerts/${alertId}/resolve`,
    method: 'post'
  })
}

export const dismissAlert = (alertId) => {
  return request({
    url: `/api/v1/monitor/alerts/${alertId}/dismiss`,
    method: 'post'
  })
}

export const getActiveAlerts = () => {
  return request({
    url: '/api/v1/monitor/alerts/active',
    method: 'get'
  })
}

// 监控数据聚合接口
export const getMonitoringSummary = (timeRange = '24h') => {
  return request({
    url: '/api/v1/monitor/summary',
    method: 'get',
    params: { timeRange }
  })
}

export default {
  // 系统概览
  getSystemOverview,
  getProjectStats,
  getRequirementStats,
  getSystemStats,
  
  // AI服务监控
  getAIServiceStats,
  getAIServices,
  getAIServiceAlerts,
  testAIService,
  restartAIService,
  getAIServiceLogs,
  getAIServiceConfig,
  
  // 性能监控
  getPerformanceStats,
  getAPIStats,
  getResourceStats,
  getPerformanceData,
  
  // Agent监控
  getAgentStats,
  getAgents,
  getTaskQueue,
  startAgent,
  stopAgent,
  restartAgent,
  getAgentLogs,
  getAgentConfig,
  cancelTask,
  getTaskDetail,
  
  // 实时日志
  getLogs,
  getLogStats,
  clearLogs,
  exportLogs,
  createLogWebSocket,
  
  // 监控设置
  getMonitorSettings,
  saveMonitorSettings,
  getAlertRules,
  saveAlertRule,
  updateAlertRule,
  deleteAlertRule,
  
  // 数据导出
  exportMonitorReport,
  exportPerformanceData,
  
  // 工具函数
  healthCheck,
  refreshAllData,
  batchOperateAgents,
  batchOperateTasks,
  resolveAlert,
  dismissAlert,
  getActiveAlerts,
  getMonitoringSummary
}