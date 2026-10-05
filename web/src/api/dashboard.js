import service from '@/utils/request'

// ==================== 统一仪表盘API ====================

// 获取仪表盘概览数据
export const getDashboardOverview = () => {
  return service({
    url: '/nesma/dashboard/overview',
    method: 'get'
  })
}

// 获取工作区统计数据
export const getWorkspaceStats = () => {
  return service({
    url: '/nesma/dashboard/workspace/stats',
    method: 'get'
  })
}

// 获取图表数据
export const getChartsData = (params) => {
  return service({
    url: '/nesma/dashboard/charts',
    method: 'get',
    params
  })
}

// 获取最近活动
export const getRecentActivities = (params) => {
  return service({
    url: '/nesma/dashboard/activities/recent',
    method: 'get',
    params
  })
}

// 获取活跃工作流
export const getActiveWorkflows = (params) => {
  return service({
    url: '/nesma/dashboard/workflows/active',
    method: 'get',
    params
  })
}

// 获取工作流模板
export const getWorkflowTemplates = () => {
  return service({
    url: '/nesma/dashboard/workflows/templates',
    method: 'get'
  })
}

// 创建工作流
export const createDashboardWorkflow = (data) => {
  return service({
    url: '/nesma/dashboard/workflows',
    method: 'post',
    data
  })
}

// 执行工作流操作
export const executeWorkflowAction = (action, workflowId, data) => {
  return service({
    url: `/nesma/dashboard/workflows/${workflowId}/actions/${action}`,
    method: 'post',
    data
  })
}

// 获取系统状态
export const getSystemStatus = () => {
  return service({
    url: '/nesma/dashboard/system/status',
    method: 'get'
  })
}

// 获取性能数据
export const getPerformanceData = (params) => {
  return service({
    url: '/nesma/dashboard/system/performance',
    method: 'get',
    params
  })
}

// 获取实时性能指标
export const getRealTimeMetrics = () => {
  return service({
    url: '/nesma/dashboard/system/metrics/realtime',
    method: 'get'
  })
}

// 导出性能数据
export const exportPerformanceData = (params) => {
  return service({
    url: '/nesma/dashboard/system/performance/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 刷新系统状态
export const refreshSystemStatus = () => {
  return service({
    url: '/nesma/dashboard/system/status/refresh',
    method: 'post'
  })
}

// 获取用户偏好设置
export const getUserPreferences = () => {
  return service({
    url: '/nesma/dashboard/user/preferences',
    method: 'get'
  })
}

// 更新用户偏好设置
export const updateUserPreferences = (data) => {
  return service({
    url: '/nesma/dashboard/user/preferences',
    method: 'put',
    data
  })
}

// 获取个性化推荐
export const getPersonalizedRecommendations = () => {
  return service({
    url: '/nesma/dashboard/recommendations/personalized',
    method: 'get'
  })
}

// 获取系统通知
export const getSystemNotifications = (params) => {
  return service({
    url: '/nesma/dashboard/notifications',
    method: 'get',
    params
  })
}

// 标记通知为已读
export const markNotificationAsRead = (notificationId) => {
  return service({
    url: `/nesma/dashboard/notifications/${notificationId}/read`,
    method: 'put'
  })
}

// 获取快速操作配置
export const getQuickActionsConfig = () => {
  return service({
    url: '/nesma/dashboard/quick-actions/config',
    method: 'get'
  })
}

// 更新快速操作配置
export const updateQuickActionsConfig = (data) => {
  return service({
    url: '/nesma/dashboard/quick-actions/config',
    method: 'put',
    data
  })
}

// 获取小部件配置
export const getWidgetConfig = () => {
  return service({
    url: '/nesma/dashboard/widgets/config',
    method: 'get'
  })
}

// 更新小部件配置
export const updateWidgetConfig = (data) => {
  return service({
    url: '/nesma/dashboard/widgets/config',
    method: 'put',
    data
  })
}

// 获取布局配置
export const getLayoutConfig = () => {
  return service({
    url: '/nesma/dashboard/layout/config',
    method: 'get'
  })
}

// 更新布局配置
export const updateLayoutConfig = (data) => {
  return service({
    url: '/nesma/dashboard/layout/config',
    method: 'put',
    data
  })
}

// ==================== 数据聚合API ====================

// 获取聚合统计数据
export const getAggregatedStats = (params) => {
  return service({
    url: '/nesma/dashboard/stats/aggregated',
    method: 'get',
    params
  })
}

// 获取趋势分析数据
export const getTrendAnalysis = (params) => {
  return service({
    url: '/nesma/dashboard/trends/analysis',
    method: 'get',
    params
  })
}

// 获取对比分析数据
export const getComparisonAnalysis = (params) => {
  return service({
    url: '/nesma/dashboard/comparison/analysis',
    method: 'get',
    params
  })
}

// 获取预测分析数据
export const getPredictiveAnalysis = (params) => {
  return service({
    url: '/nesma/dashboard/predictive/analysis',
    method: 'get',
    params
  })
}

// ==================== 缓存管理API ====================

// 刷新仪表盘缓存
export const refreshDashboardCache = () => {
  return service({
    url: '/nesma/dashboard/cache/refresh',
    method: 'post'
  })
}

// 清理仪表盘缓存
export const clearDashboardCache = () => {
  return service({
    url: '/nesma/dashboard/cache/clear',
    method: 'post'
  })
}

// 获取缓存状态
export const getCacheStatus = () => {
  return service({
    url: '/nesma/dashboard/cache/status',
    method: 'get'
  })
}

// ==================== 实时数据API ====================

// 建立WebSocket连接获取实时数据
export const establishWebSocketConnection = (endpoint) => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = window.location.host
  return new WebSocket(`${protocol}//${host}/api/v1/nesma/dashboard/ws/${endpoint}`)
}

// 获取实时统计数据
export const getRealTimeStats = () => {
  return service({
    url: '/nesma/dashboard/realtime/stats',
    method: 'get'
  })
}

// 获取实时活动流
export const getRealTimeActivityStream = () => {
  return service({
    url: '/nesma/dashboard/realtime/activities',
    method: 'get'
  })
}

// 获取实时系统指标
export const getRealTimeSystemMetrics = () => {
  return service({
    url: '/nesma/dashboard/realtime/system-metrics',
    method: 'get'
  })
}

// ==================== 数据导出API ====================

// 导出仪表盘数据
export const exportDashboardData = (params) => {
  return service({
    url: '/nesma/dashboard/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 导出统计报告
export const exportStatsReport = (params) => {
  return service({
    url: '/nesma/dashboard/export/stats-report',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 导出图表数据
export const exportChartData = (params) => {
  return service({
    url: '/nesma/dashboard/export/chart-data',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// ==================== 健康检查API ====================

// 系统健康检查
export const performHealthCheck = () => {
  return service({
    url: '/nesma/dashboard/health/check',
    method: 'post'
  })
}

// 获取健康检查历史
export const getHealthCheckHistory = (params) => {
  return service({
    url: '/nesma/dashboard/health/history',
    method: 'get',
    params
  })
}

// 获取服务依赖状态
export const getServiceDependencies = () => {
  return service({
    url: '/nesma/dashboard/health/dependencies',
    method: 'get'
  })
}

// ==================== 告警管理API ====================

// 获取告警列表
export const getAlerts = (params) => {
  return service({
    url: '/nesma/dashboard/alerts',
    method: 'get',
    params
  })
}

// 创建告警规则
export const createAlertRule = (data) => {
  return service({
    url: '/nesma/dashboard/alerts/rules',
    method: 'post',
    data
  })
}

// 更新告警规则
export const updateAlertRule = (data) => {
  return service({
    url: '/nesma/dashboard/alerts/rules',
    method: 'put',
    data
  })
}

// 删除告警规则
export const deleteAlertRule = (ruleId) => {
  return service({
    url: `/nesma/dashboard/alerts/rules/${ruleId}`,
    method: 'delete'
  })
}

// 确认告警
export const acknowledgeAlert = (alertId, data) => {
  return service({
    url: `/nesma/dashboard/alerts/${alertId}/acknowledge`,
    method: 'post',
    data
  })
}

// 关闭告警
export const closeAlert = (alertId, data) => {
  return service({
    url: `/nesma/dashboard/alerts/${alertId}/close`,
    method: 'post',
    data
  })
}

// ==================== 集成API ====================

// 获取外部系统集成状态
export const getIntegrationStatus = () => {
  return service({
    url: '/nesma/dashboard/integrations/status',
    method: 'get'
  })
}

// 测试外部系统连接
export const testIntegrationConnection = (integrationId) => {
  return service({
    url: `/nesma/dashboard/integrations/${integrationId}/test`,
    method: 'post'
  })
}

// 同步外部系统数据
export const syncIntegrationData = (integrationId, data) => {
  return service({
    url: `/nesma/dashboard/integrations/${integrationId}/sync`,
    method: 'post',
    data
  })
}

// ==================== 审计日志API ====================

// 获取审计日志
export const getAuditLogs = (params) => {
  return service({
    url: '/nesma/dashboard/audit/logs',
    method: 'get',
    params
  })
}

// 获取用户操作日志
export const getUserActionLogs = (params) => {
  return service({
    url: '/nesma/dashboard/audit/user-actions',
    method: 'get',
    params
  })
}

// 获取系统事件日志
export const getSystemEventLogs = (params) => {
  return service({
    url: '/nesma/dashboard/audit/system-events',
    method: 'get',
    params
  })
}

// 导出审计日志
export const exportAuditLogs = (params) => {
  return service({
    url: '/nesma/dashboard/audit/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}