import service from '@/utils/request'

// 创建需求版本
export const createRequirementVersion = (data) => {
  return service({
    url: '/nesma/requirement-version',
    method: 'post',
    data
  })
}

// 获取周期的所有版本
export const getRequirementVersions = (cycleId) => {
  return service({
    url: `/nesma/requirement-version/cycle/${cycleId}`,
    method: 'get'
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