import service from '@/utils/request'

// 导出综合文档
export const exportComprehensiveDocument = (data) => {
  return service({
    url: '/nesma/document/export/comprehensive',
    method: 'post',
    data
  })
}

// 获取项目和周期信息
export const getProjectCycleInfo = (cycleId) => {
  return service({
    url: `/nesma/document/project-cycle-info/${cycleId}`,
    method: 'get'
  })
}

// 获取报告模板列表
export const getDocumentTemplates = (params) => {
  return service({
    url: '/nesma/document/templates',
    method: 'get',
    params
  })
}

// 创建报告模板
export const createDocumentTemplate = (data) => {
  return service({
    url: '/nesma/document/templates',
    method: 'post',
    data
  })
}

// 更新报告模板
export const updateDocumentTemplate = (id, data) => {
  return service({
    url: `/nesma/document/templates/${id}`,
    method: 'put',
    data
  })
}

// 删除报告模板
export const deleteDocumentTemplate = (id) => {
  return service({
    url: `/nesma/document/templates/${id}`,
    method: 'delete'
  })
}

// 生成报告预览
export const generateDocumentPreview = (data) => {
  return service({
    url: '/nesma/document/preview',
    method: 'post',
    data
  })
}

// 获取导出进度
export const getExportProgress = (exportId) => {
  return service({
    url: `/nesma/document/export/progress/${exportId}`,
    method: 'get'
  })
}

// 取消导出
export const cancelDocumentExport = (exportId) => {
  return service({
    url: `/nesma/document/export/cancel/${exportId}`,
    method: 'post'
  })
}

// 下载导出文件
export const downloadExportedFile = (exportId, fileName) => {
  return service({
    url: `/nesma/document/export/download/${exportId}/${fileName}`,
    method: 'get',
    responseType: 'blob'
  })
}

// 获取导出历史
export const getExportHistory = (params) => {
  return service({
    url: '/nesma/document/export/history',
    method: 'get',
    params
  })
}

// 批量导出
export const batchExportDocuments = (data) => {
  return service({
    url: '/nesma/document/export/batch',
    method: 'post',
    data
  })
}

// 获取可用的导出格式
export const getAvailableFormats = () => {
  return service({
    url: '/nesma/document/export/formats',
    method: 'get'
  })
}

// 验证模板
export const validateTemplate = (data) => {
  return service({
    url: '/nesma/document/templates/validate',
    method: 'post',
    data
  })
}

// 获取导出统计
export const getExportStatistics = (params) => {
  return service({
    url: '/nesma/document/export/statistics',
    method: 'get',
    params
  })
}