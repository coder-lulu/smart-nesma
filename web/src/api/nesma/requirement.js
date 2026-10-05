import service from '@/utils/request'

// 创建需求
export const createRequirement = (data) => {
  return service({
    url: '/nesma/requirement',
    method: 'post',
    data
  })
}

// 更新需求
export const updateRequirement = (data) => {
  return service({
    url: '/nesma/requirement',
    method: 'put',
    data
  })
}

// 删除需求
export const deleteRequirement = (id) => {
  return service({
    url: `/nesma/requirement/${id}`,
    method: 'delete'
  })
}

// 批量删除需求
export const batchDeleteRequirements = (ids) => {
  return service({
    url: '/nesma/requirement/batch-delete',
    method: 'post',
    data: { ids }
  })
}

// 获取需求详情
export const getRequirement = (id) => {
  return service({
    url: `/nesma/requirement/${id}`,
    method: 'get'
  })
}

// 获取需求列表（分页）
export const getRequirementList = (params) => {
  return service({
    url: '/nesma/requirement/list',
    method: 'get',
    params
  })
}

// 获取需求树形结构
export const getRequirementTree = (params) => {
  return service({
    url: '/nesma/requirement/tree',
    method: 'get',
    params
  })
}

// 获取需求统计信息
export const getRequirementStats = (projectId) => {
  return service({
    url: `/nesma/requirement/stats`,
    method: 'get',
    params: { projectId }
  })
}

// 获取父级需求选项
export const getParentRequirementOptions = (projectId, level) => {
  return service({
    url: '/nesma/requirement/parent-options',
    method: 'get',
    params: { projectId, level }
  })
}

// 获取需求的子功能点信息
export const getRequirementChildren = (id) => {
  return service({
    url: `/nesma/requirement/${id}/children`,
    method: 'get'
  })
}

// 批量更新排序
export const batchUpdateOrder = (items) => {
  return service({
    url: '/nesma/requirement/batch-update-order',
    method: 'post',
    data: { items }
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

// Excel导入
export const importFromExcel = (formData) => {
  return service({
    url: '/nesma/requirement/import-excel',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 根据条件删除需求
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

// AI分析相关API
export const analyzeRequirement = (data) => {
  return service({
    url: '/nesma/requirement/ai-analysis',
    method: 'post',
    data
  })
}

// 批量AI分析
export const batchAnalyzeRequirements = (data) => {
  return service({
    url: '/nesma/requirement/batch-ai-analysis',
    method: 'post',
    data
  })
}

// 获取需求分析进度
export const getAnalysisProgress = (taskId) => {
  return service({
    url: `/nesma/requirement/analysis-progress/${taskId}`,
    method: 'get'
  })
}

// 应用需求分析结果
export const applyRequirementAnalysis = (data) => {
  return service({
    url: '/nesma/requirement/apply-analysis',
    method: 'post',
    data
  })
}

// 复制需求
export const copyRequirement = (data) => {
  return service({
    url: '/nesma/requirement/copy',
    method: 'post',
    data
  })
}

// 导出需求
export const exportRequirements = (params) => {
  return service({
    url: '/nesma/requirement/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 搜索需求（支持复杂条件）
export const searchRequirements = (params) => {
  return service({
    url: '/nesma/requirement/search',
    method: 'post',
    data: params
  })
}

// 获取需求版本对比
export const compareRequirementVersions = (params) => {
  return service({
    url: '/nesma/requirement/version-compare',
    method: 'get',
    params
  })
}

// 获取需求依赖关系
export const getRequirementDependencies = (id) => {
  return service({
    url: `/nesma/requirement/${id}/dependencies`,
    method: 'get'
  })
}

// 设置需求依赖关系
export const setRequirementDependencies = (data) => {
  return service({
    url: '/nesma/requirement/dependencies',
    method: 'post',
    data
  })
}