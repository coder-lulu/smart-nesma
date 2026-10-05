import service from '@/utils/request'

// 获取AI智能推荐
export const getAIRecommendations = (data) => {
  return service({
    url: '/nesma/recommendation/ai',
    method: 'post',
    data
  })
}

// 采纳AI推荐
export const adoptRecommendation = (data) => {
  return service({
    url: '/nesma/recommendation/adopt',
    method: 'post',
    data
  })
}

// 拒绝AI推荐
export const rejectRecommendation = (data) => {
  return service({
    url: '/nesma/recommendation/reject',
    method: 'post',
    data
  })
}

// 获取快速推荐
export const getQuickRecommendations = (projectId) => {
  return service({
    url: '/nesma/recommendation/quick',
    method: 'get',
    params: { projectId }
  })
}

// 获取采纳历史
export const getAdoptionHistory = (params) => {
  return service({
    url: '/nesma/recommendation/adoption-history',
    method: 'get',
    params
  })
}

// 获取批量推荐
export const getBatchRecommendations = (projectIds) => {
  return service({
    url: '/nesma/recommendation/batch',
    method: 'post',
    data: projectIds
  })
}

// 获取AI推荐配置
export const getAIRecommendationConfig = () => {
  return service({
    url: '/nesma/recommendation/ai-config',
    method: 'get'
  })
}

// 以下是传统推荐API（保留兼容性）
export const getRecommendations = (params) => {
  return service({
    url: '/nesma/recommendation/items',
    method: 'get',
    params
  })
}

// 记录推荐反馈
export const recordRecommendationFeedback = (data) => {
  return service({
    url: '/nesma/recommendation/feedback',
    method: 'post',
    data
  })
}

// 提交推荐反馈 (别名，为了兼容性)
export const submitRecommendationFeedback = recordRecommendationFeedback

// 记录用户行为
export const recordUserBehavior = (data) => {
  return service({
    url: '/nesma/recommendation/behavior',
    method: 'post',
    data
  })
}

export const getPersonalizedRecommendations = (data) => {
  return service({
    url: '/nesma/recommendation/personalized',
    method: 'post',
    data
  })
}

export const getRecommendationStats = (params) => {
  return service({
    url: '/nesma/recommendation/stats',
    method: 'get',
    params
  })
}

export const getRecommendationConfig = () => {
  return service({
    url: '/nesma/recommendation/config',
    method: 'get'
  })
}