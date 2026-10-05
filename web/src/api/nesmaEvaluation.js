import service from '@/utils/request'

// 创建NESMA评估
export const createEvaluation = (data) => {
  return service({
    url: '/nesma-evaluation/create',
    method: 'post',
    data
  })
}

// 更新NESMA评估
export const updateEvaluation = (data) => {
  return service({
    url: `/nesma-evaluation/${data.id}`,
    method: 'put',
    data
  })
}

// 开始评估
export const startEvaluation = (id) => {
  return service({
    url: '/nesma-evaluation/start',
    method: 'post',
    data: { evaluationId: id }
  })
}

// 获取评估详情
export const getEvaluation = (id) => {
  return service({
    url: `/nesma-evaluation/${id}`,
    method: 'get'
  })
}

// 获取评估列表
export const getEvaluationList = (params) => {
  return service({
    url: '/nesma-evaluation/list',
    method: 'get',
    params
  })
}

// 获取功能点列表
export const getFunctionPoints = (params) => {
  return service({
    url: '/nesma-evaluation/function-points',
    method: 'get',
    params
  })
}

// 创建功能点
export const createFunctionPoint = (data) => {
  return service({
    url: '/nesma-evaluation/function-point',
    method: 'post',
    data
  })
}

// 更新功能点
export const updateFunctionPoint = (data) => {
  return service({
    url: `/nesma-evaluation/function-point/${data.id}`,
    method: 'put',
    data
  })
}

// 删除功能点
export const deleteFunctionPoint = (id) => {
  return service({
    url: `/nesma-evaluation/function-point/${id}`,
    method: 'delete'
  })
}

// 获取评估统计
export const getEvaluationStats = (id) => {
  return service({
    url: `/nesma-evaluation/${id}/stats`,
    method: 'get'
  })
}

// 获取复杂度指标
export const getComplexityMetrics = (params) => {
  return service({
    url: '/nesma-evaluation/complexity-metrics',
    method: 'get',
    params
  })
}

// 获取验证项目
export const getValidationItems = (params) => {
  return service({
    url: '/nesma-evaluation/validation-items',
    method: 'get',
    params
  })
}

// 更新验证项目
export const updateValidationItem = (data) => {
  return service({
    url: `/nesma-evaluation/validation-item/${data.id}`,
    method: 'put',
    data
  })
}

// 评估审核
export const reviewEvaluation = (data) => {
  return service({
    url: '/nesma-evaluation/review',
    method: 'post',
    data
  })
}

// 导出评估报告
export const exportEvaluationReport = (id, params) => {
  return service({
    url: `/nesma-evaluation/${id}/export`,
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 生成评估报告（兼容性别名）
export const generateEvaluationReport = (id, format = 'pdf') => {
  return exportEvaluationReport(id, { format })
}

// 获取项目评估摘要
export const getProjectEvaluationSummary = (projectId) => {
  return service({
    url: `/nesma-evaluation/project/${projectId}/summary`,
    method: 'get'
  })
}

// 重新计算评估
export const recalculateEvaluation = (id) => {
  return service({
    url: `/nesma-evaluation/${id}/recalculate`,
    method: 'post'
  })
}

// 删除评估（新增）
export const deleteEvaluation = (id) => {
  return service({
    url: `/nesma-evaluation/${id}`,
    method: 'delete'
  })
}

// 获取评估配置选项
export const getEvaluationConfig = () => {
  return service({
    url: '/nesma-evaluation/config',
    method: 'get'
  })
}

// 批量删除评估
export const deleteEvaluations = (ids) => {
  return service({
    url: '/nesma-evaluation/batch-delete',
    method: 'post',
    data: { ids }
  })
}

// 克隆评估
export const cloneEvaluation = (id, data) => {
  return service({
    url: `/nesma-evaluation/${id}/clone`,
    method: 'post',
    data
  })
}

// 获取评估历史
export const getEvaluationHistory = (id) => {
  return service({
    url: `/nesma-evaluation/${id}/history`,
    method: 'get'
  })
}

// 比较评估结果
export const compareEvaluations = (data) => {
  return service({
    url: '/nesma-evaluation/compare',
    method: 'post',
    data
  })
}

// ==================== 评估因子配置相关API ====================

// 获取评估因子配置
export const getEvaluationFactors = (evaluationId) => {
  return service({
    url: `/nesma-evaluation/${evaluationId}/factors`,
    method: 'get'
  })
}

// 保存评估因子配置
export const saveEvaluationFactors = (data) => {
  return service({
    url: '/nesma-evaluation/factors',
    method: 'post',
    data
  })
}

// 更新评估因子配置
export const updateEvaluationFactors = (evaluationId, data) => {
  return service({
    url: `/nesma-evaluation/${evaluationId}/factors`,
    method: 'put',
    data
  })
}

// 获取默认NESMA因子配置
export const getDefaultNESMAFactors = (nesmaVersion = 'v2.2') => {
  return service({
    url: '/nesma-evaluation/default-factors',
    method: 'get',
    params: { version: nesmaVersion }
  })
}

// 验证评估因子配置
export const validateEvaluationFactors = (data) => {
  return service({
    url: '/nesma-evaluation/validate-factors',
    method: 'post',
    data
  })
}

// 计算调整因子
export const calculateAdjustmentFactor = (adjustmentFactors) => {
  return service({
    url: '/nesma-evaluation/calculate-adjustment',
    method: 'post',
    data: { adjustmentFactors }
  })
}

// 预览功能点计算
export const previewFunctionPointCalculation = (data) => {
  return service({
    url: '/nesma-evaluation/preview-calculation',
    method: 'post',
    data
  })
}

// 应用评估因子到项目
export const applyFactorsToProject = (evaluationId, data) => {
  return service({
    url: `/nesma-evaluation/${evaluationId}/apply-factors`,
    method: 'post',
    data
  })
}

// 获取评估因子历史记录
export const getFactorsHistory = (evaluationId) => {
  return service({
    url: `/nesma-evaluation/${evaluationId}/factors-history`,
    method: 'get'
  })
}

// 导入评估因子配置
export const importEvaluationFactors = (evaluationId, file) => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('evaluationId', evaluationId)
  
  return service({
    url: '/nesma-evaluation/import-factors',
    method: 'post',
    data: formData,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 导出评估因子配置
export const exportEvaluationFactors = (evaluationId, format = 'json') => {
  return service({
    url: `/nesma-evaluation/${evaluationId}/export-factors`,
    method: 'get',
    params: { format },
    responseType: format === 'json' ? 'json' : 'blob'
  })
}

// ==================== 评估进度相关API ====================

// 获取评估进度
export const getEvaluationProgress = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/progress`,
    method: 'get'
  })
}

// 启动评估（进度跟踪版本）
export const startEvaluationProgress = (evaluationId, config = {}) => {
  return service({
    url: '/nesma/evaluation/progress/start',
    method: 'post',
    data: {
      evaluation_id: evaluationId,
      config
    }
  })
}

// 取消评估
export const cancelEvaluation = (evaluationId, reason = '') => {
  return service({
    url: '/nesma/evaluation/progress/cancel',
    method: 'post',
    data: {
      evaluation_id: evaluationId,
      reason
    }
  })
}

// 获取评估状态
export const getEvaluationStatus = (evaluationId) => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/status`,
    method: 'get'
  })
}

// 获取评估统计（进度版本）
export const getEvaluationProgressStats = () => {
  return service({
    url: '/nesma/evaluation/progress/stats',
    method: 'get'
  })
}

// 获取评估日志
export const getEvaluationLogs = (evaluationId, page = 1, pageSize = 20, level = '') => {
  return service({
    url: `/nesma/evaluation/${evaluationId}/logs`,
    method: 'get',
    params: {
      page,
      page_size: pageSize,
      level
    }
  })
}

// 手动更新评估进度（内部接口）
export const updateEvaluationProgress = (evaluationId, progressData) => {
  return service({
    url: '/nesma/evaluation/progress/update',
    method: 'post',
    data: {
      evaluation_id: evaluationId,
      ...progressData
    }
  })
} 