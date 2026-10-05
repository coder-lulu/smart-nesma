import service from '@/utils/request'

// ==================== AI对话相关API ====================

// 发送消息
export const sendMessage = (data) => {
  return service({
    url: '/chat/send',
    method: 'post',
    data
  })
}

// 创建会话
export const createChatSession = (data) => {
  return service({
    url: '/chat/session',
    method: 'post',
    data
  })
}

// 获取会话列表
export const getChatSessionList = (params) => {
  return service({
    url: '/chat/sessions',
    method: 'get',
    params
  })
}

// 获取会话详情
export const getChatSession = (sessionId) => {
  return service({
    url: `/chat/session/${sessionId}`,
    method: 'get'
  })
}

// 更新会话
export const updateChatSession = (sessionId, data) => {
  return service({
    url: `/chat/session/${sessionId}`,
    method: 'put',
    data
  })
}

// 删除会话
export const deleteChatSession = (sessionId) => {
  return service({
    url: `/chat/session/${sessionId}`,
    method: 'delete'
  })
}

// 获取消息列表
export const getChatMessages = (sessionId, params) => {
  return service({
    url: `/chat/session/${sessionId}/messages`,
    method: 'get',
    params
  })
}

// 清空会话消息
export const clearChatMessages = (sessionId) => {
  return service({
    url: `/chat/session/${sessionId}/clear`,
    method: 'post'
  })
}

// 获取聊天统计
export const getChatStats = (params) => {
  return service({
    url: '/chat/stats',
    method: 'get',
    params
  })
}

// 导出聊天记录
export const exportChatHistory = (sessionId, params) => {
  return service({
    url: `/chat/session/${sessionId}/export`,
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 重置AI服务熔断器
export const resetCircuitBreaker = () => {
  return service({
    url: '/chat/reset-circuit-breaker',
    method: 'post'
  })
}

// 获取AI服务健康状态
export const getAIHealth = () => {
  return service({
    url: '/chat/health',
    method: 'get'
  })
}

// 测试API连通性
export const testAPIConnectivity = () => {
  return service({
    url: '/chat/health',
    method: 'get',
    timeout: 10000 // 10秒超时
  })
} 