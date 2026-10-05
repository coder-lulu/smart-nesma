import service from '@/utils/request'

// ==================== 知识图谱管理API (统一接口) ====================

// 获取知识实体列表
export const getKnowledgeEntities = (params) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge-entities',
    method: 'get',
    params
  })
}

// 获取知识关系列表
export const getKnowledgeRelations = (params) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge-relations', 
    method: 'get',
    params
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

// 创建知识实体
export const createKnowledgeEntity = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge-entities',
    method: 'post',
    data
  })
}

// 更新知识实体
export const updateKnowledgeEntity = (data) => {
  return service({
    url: '/nesma/knowledge/entity',
    method: 'put',
    data
  })
}

// 删除知识实体
export const deleteKnowledgeEntity = (id) => {
  return service({
    url: `/nesma/knowledge/entity/${id}`,
    method: 'delete'
  })
}

// 获取实体详情
export const getKnowledgeEntityDetail = (id) => {
  return service({
    url: `/nesma/knowledge/entity/${id}`,
    method: 'get'
  })
}

// 创建实体关系
export const createKnowledgeRelation = (data) => {
  return service({
    url: '/nesma/knowledge/relation',
    method: 'post',
    data
  })
}

// 更新实体关系
export const updateKnowledgeRelation = (data) => {
  return service({
    url: '/nesma/knowledge/relation',
    method: 'put',
    data
  })
}

// 删除实体关系
export const deleteKnowledgeRelation = (id) => {
  return service({
    url: `/nesma/knowledge/relation/${id}`,
    method: 'delete'
  })
}

// 向量搜索
export const vectorSearch = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/vector/search',
    method: 'post',
    data
  })
}

// 搜索知识库
export const searchKnowledgeBase = (data) => {
  return service({
    url: '/nesma/intelligent-analysis/knowledge/search',
    method: 'post',
    data
  })
}

// 搜索相似实体
export const searchSimilarEntities = (params) => {
  return service({
    url: '/nesma/knowledge/similar-entities',
    method: 'get',
    params
  })
}

// 语义推理
export const semanticReasoning = (data) => {
  return service({
    url: '/nesma/knowledge/semantic-reasoning',
    method: 'post',
    data
  })
}

// 获取知识图谱统计信息
export const getKnowledgeGraphStats = () => {
  return service({
    url: '/nesma/knowledge/graph/stats',
    method: 'get'
  })
}

// 导出知识图谱
export const exportKnowledgeGraph = (params) => {
  return service({
    url: '/nesma/intelligent-analysis/export',
    method: 'get',
    params,
    responseType: 'blob'
  })
}

// 导入知识图谱
export const importKnowledgeGraph = (data) => {
  return service({
    url: '/nesma/knowledge/graph/import',
    method: 'post',
    data,
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
}

// 获取知识图谱建议
export const getKnowledgeGraphRecommendations = (params) => {
  return service({
    url: '/nesma/knowledge/graph/recommendations',
    method: 'get',
    params
  })
}

// 验证知识图谱一致性
export const validateKnowledgeGraph = () => {
  return service({
    url: '/nesma/knowledge/graph/validate',
    method: 'post'
  })
}

// 获取相关节点
export const getRelatedNodes = (nodeId, depth = 1) => {
  return service({
    url: `/nesma/knowledge/node/${nodeId}/related`,
    method: 'get',
    params: { depth }
  })
}

// 图谱布局算法设置
export const updateGraphLayout = (layoutType, params) => {
  return service({
    url: '/nesma/knowledge/graph/layout',
    method: 'post',
    data: { 
      layoutType,
      params 
    }
  })
}