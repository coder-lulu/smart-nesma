import { defineStore } from 'pinia'
import { ref, computed, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import {
  getRequirementTree,
  getRequirementList,
  getRequirementStats,
  createRequirement,
  updateRequirement,
  deleteRequirement,
  batchDeleteRequirements,
  getParentRequirementOptions,
  analyzeRequirement,
  batchAnalyzeRequirements,
  moveRequirement,
  copyRequirement,
  exportRequirements,
  importFromExcel
} from '@/api/nesma/requirement'
import {
  getRequirementVersions,
  createRequirementVersion,
  setActiveVersion
} from '@/api/nesma/requirementVersion'

export const useRequirementStore = defineStore('requirement', () => {
  // ==================== 状态定义 ====================
  
  // 基础数据状态
  const treeData = ref([])
  const tableData = ref([])
  const stats = ref({})
  const loading = ref(false)
  const submitting = ref(false)
  
  // 分页状态
  const pagination = reactive({
    page: 1,
    pageSize: 20,
    total: 0
  })
  
  // 搜索条件状态
  const searchForm = reactive({
    projectId: null,
    cycleId: null,
    versionId: null,
    keyword: '',
    level: null,
    status: '',
    functionType: '',
    complexity: [],
    analysisStatus: [],
    dateRange: null,
    minFunctionPoints: null,
    maxFunctionPoints: null
  })
  
  // 选择状态
  const selectedRows = ref([])
  const currentNode = ref(null)
  
  // 视图状态
  const viewMode = ref('tree') // tree | table | kanban
  const showAdvancedFilter = ref(false)
  
  // 版本管理状态
  const versions = ref([])
  const activeVersion = ref(null)
  
  // 父级选项缓存
  const parentOptionsCache = reactive(new Map())
  
  // 分析任务状态
  const analysisTasks = ref([])
  const analysisProgress = reactive({})
  
  // ==================== 计算属性 ====================
  
  const hasSearchConditions = computed(() => {
    return !!(
      searchForm.keyword ||
      searchForm.level ||
      searchForm.status ||
      searchForm.functionType ||
      searchForm.complexity?.length ||
      searchForm.analysisStatus?.length ||
      searchForm.dateRange ||
      searchForm.minFunctionPoints ||
      searchForm.maxFunctionPoints
    )
  })
  
  const filteredCount = computed(() => {
    if (viewMode.value === 'tree') {
      return countTreeNodes(treeData.value)
    } else {
      return tableData.value.length
    }
  })
  
  const selectedCount = computed(() => selectedRows.value.length)
  
  const canAnalyze = computed(() => {
    return selectedRows.value.some(row => 
      row.level >= 3 && (!row.aiAnalysisStatus || row.aiAnalysisStatus === 'failed')
    )
  })
  
  // ==================== 辅助方法 ====================
  
  const countTreeNodes = (nodes) => {
    let count = 0
    nodes.forEach(node => {
      count++
      if (node.children?.length) {
        count += countTreeNodes(node.children)
      }
    })
    return count
  }
  
  const buildSearchParams = () => {
    const params = { ...searchForm }
    
    // 处理日期范围
    if (params.dateRange?.length === 2) {
      params.startDate = params.dateRange[0]
      params.endDate = params.dateRange[1]
    }
    delete params.dateRange
    
    // 过滤空值
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || 
          (Array.isArray(params[key]) && params[key].length === 0)) {
        delete params[key]
      }
    })
    
    return params
  }
  
  // ==================== 数据加载方法 ====================
  
  const loadTreeData = async (forceRefresh = false) => {
    if (!searchForm.projectId) {
      treeData.value = []
      return
    }
    
    try {
      loading.value = true
      const params = buildSearchParams()
      
      const response = await getRequirementTree(params)
      treeData.value = response.data || []
      
      // 同时更新统计信息
      if (forceRefresh) {
        await loadStats()
      }
    } catch (error) {
      console.error('加载需求树失败:', error)
      ElMessage.error('加载需求树失败: ' + (error.message || '未知错误'))
      treeData.value = []
    } finally {
      loading.value = false
    }
  }
  
  const loadTableData = async (page = pagination.page, pageSize = pagination.pageSize) => {
    if (!searchForm.projectId) {
      tableData.value = []
      pagination.total = 0
      return
    }
    
    try {
      loading.value = true
      const params = {
        ...buildSearchParams(),
        page,
        pageSize
      }
      
      const response = await getRequirementList(params)
      const result = response.data || {}
      
      tableData.value = result.list || []
      pagination.page = result.page || page
      pagination.pageSize = result.pageSize || pageSize
      pagination.total = result.total || 0
    } catch (error) {
      console.error('加载需求列表失败:', error)
      ElMessage.error('加载需求列表失败: ' + (error.message || '未知错误'))
      tableData.value = []
      pagination.total = 0
    } finally {
      loading.value = false
    }
  }
  
  const loadStats = async () => {
    try {
      // 如果没有指定项目ID，则获取全部项目的统计
      const response = await getRequirementStats(searchForm.projectId)
      stats.value = response.data || {}
    } catch (error) {
      console.error('加载统计信息失败:', error)
      stats.value = {}
    }
  }
  
  const loadVersions = async () => {
    if (!searchForm.cycleId) {
      versions.value = []
      activeVersion.value = null
      return
    }
    
    try {
      const response = await getRequirementVersions(searchForm.cycleId)
      versions.value = response.data || []
      
      // 查找激活版本
      activeVersion.value = versions.value.find(v => v.isActive) || null
      if (activeVersion.value) {
        searchForm.versionId = activeVersion.value.id
      }
    } catch (error) {
      console.error('加载版本列表失败:', error)
      versions.value = []
      activeVersion.value = null
    }
  }
  
  const loadParentOptions = async (projectId, level) => {
    if (!projectId || level <= 1) return []
    
    const cacheKey = `${projectId}-${level}`
    if (parentOptionsCache.has(cacheKey)) {
      return parentOptionsCache.get(cacheKey)
    }
    
    try {
      const response = await getParentRequirementOptions(projectId, level)
      const options = response.data?.options || []
      
      // 缓存结果
      parentOptionsCache.set(cacheKey, options)
      return options
    } catch (error) {
      console.error('加载父级选项失败:', error)
      return []
    }
  }
  
  // ==================== 数据操作方法 ====================
  
  const createNew = async (data) => {
    try {
      submitting.value = true
      const response = await createRequirement({
        ...data,
        projectId: searchForm.projectId,
        cycleId: searchForm.cycleId,
        versionId: searchForm.versionId
      })
      
      ElMessage.success('需求创建成功')
      await refresh()
      return response.data
    } catch (error) {
      console.error('创建需求失败:', error)
      ElMessage.error('创建需求失败: ' + (error.message || '未知错误'))
      throw error
    } finally {
      submitting.value = false
    }
  }
  
  const update = async (id, data) => {
    try {
      submitting.value = true
      const response = await updateRequirement({
        id,
        ...data,
        projectId: searchForm.projectId,
        cycleId: searchForm.cycleId,
        versionId: searchForm.versionId
      })
      
      ElMessage.success('需求更新成功')
      await refresh()
      return response.data
    } catch (error) {
      console.error('更新需求失败:', error)
      ElMessage.error('更新需求失败: ' + (error.message || '未知错误'))
      throw error
    } finally {
      submitting.value = false
    }
  }
  
  const remove = async (id) => {
    try {
      await deleteRequirement(id)
      ElMessage.success('需求删除成功')
      await refresh()
    } catch (error) {
      console.error('删除需求失败:', error)
      ElMessage.error('删除需求失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  const batchRemove = async (ids) => {
    try {
      await batchDeleteRequirements(ids)
      ElMessage.success(`成功删除 ${ids.length} 个需求`)
      selectedRows.value = []
      await refresh()
    } catch (error) {
      console.error('批量删除需求失败:', error)
      ElMessage.error('批量删除需求失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  const move = async (dragNode, dropNode, dropType) => {
    try {
      const data = {
        id: dragNode.id,
        parentId: dropType === 'inner' ? dropNode.id : dropNode.parentId,
        position: dropType === 'before' ? dropNode.orderIndex : dropNode.orderIndex + 1
      }
      
      await moveRequirement(data)
      ElMessage.success('需求移动成功')
      await refresh()
    } catch (error) {
      console.error('移动需求失败:', error)
      ElMessage.error('移动需求失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  const copy = async (sourceId, targetData) => {
    try {
      const response = await copyRequirement({
        sourceId,
        ...targetData,
        projectId: searchForm.projectId,
        cycleId: searchForm.cycleId,
        versionId: searchForm.versionId
      })
      
      ElMessage.success('需求复制成功')
      await refresh()
      return response.data
    } catch (error) {
      console.error('复制需求失败:', error)
      ElMessage.error('复制需求失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  // ==================== AI分析方法 ====================
  
  const analyzeRequirementItem = async (requirementId) => {
    try {
      const response = await analyzeRequirement({
        requirementId,
        projectId: searchForm.projectId,
        cycleId: searchForm.cycleId,
        versionId: searchForm.versionId
      })
      
      const taskId = response.data?.taskId
      if (taskId) {
        analysisTasks.value.push(taskId)
        analysisProgress[taskId] = { progress: 0, status: 'running' }
      }
      
      ElMessage.success('AI分析任务已启动')
      return taskId
    } catch (error) {
      console.error('启动AI分析失败:', error)
      ElMessage.error('启动AI分析失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  const batchAnalyze = async (requirementIds) => {
    try {
      const response = await batchAnalyzeRequirements({
        requirementIds,
        projectId: searchForm.projectId,
        cycleId: searchForm.cycleId,
        versionId: searchForm.versionId
      })
      
      const taskId = response.data?.taskId
      if (taskId) {
        analysisTasks.value.push(taskId)
        analysisProgress[taskId] = { progress: 0, status: 'running' }
      }
      
      ElMessage.success(`已启动 ${requirementIds.length} 个需求的批量AI分析`)
      return taskId
    } catch (error) {
      console.error('启动批量AI分析失败:', error)
      ElMessage.error('启动批量AI分析失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  // ==================== 版本管理方法 ====================
  
  const createVersion = async (versionData) => {
    try {
      const response = await createRequirementVersion({
        ...versionData,
        cycleId: searchForm.cycleId
      })
      
      ElMessage.success('版本创建成功')
      await loadVersions()
      return response.data
    } catch (error) {
      console.error('创建版本失败:', error)
      ElMessage.error('创建版本失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  const switchVersion = async (versionId) => {
    try {
      await setActiveVersion(versionId)
      searchForm.versionId = versionId
      
      ElMessage.success('版本切换成功')
      await refresh()
    } catch (error) {
      console.error('切换版本失败:', error)
      ElMessage.error('切换版本失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  // ==================== 导入导出方法 ====================
  
  const importExcel = async (file, options = {}) => {
    try {
      const formData = new FormData()
      formData.append('file', file)
      formData.append('projectId', searchForm.projectId)
      formData.append('cycleId', searchForm.cycleId)
      formData.append('versionId', searchForm.versionId)
      
      Object.keys(options).forEach(key => {
        formData.append(key, options[key])
      })
      
      const response = await importFromExcel(formData)
      const result = response.data || {}
      
      if (result.success) {
        ElMessage.success(`导入成功：共处理 ${result.totalRows} 行，成功 ${result.successRows} 行`)
      } else {
        ElMessage.warning(`导入部分成功：成功 ${result.successRows} 行，失败 ${result.failedRows} 行`)
      }
      
      await refresh()
      return result
    } catch (error) {
      console.error('Excel导入失败:', error)
      ElMessage.error('Excel导入失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  const exportExcel = async (options = {}) => {
    try {
      const params = {
        ...buildSearchParams(),
        ...options
      }
      
      const response = await exportRequirements(params)
      
      // 创建下载链接
      const blob = new Blob([response.data], {
        type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet'
      })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = `需求列表_${new Date().toISOString().slice(0, 10)}.xlsx`
      link.click()
      window.URL.revokeObjectURL(url)
      
      ElMessage.success('导出成功')
    } catch (error) {
      console.error('Excel导出失败:', error)
      ElMessage.error('Excel导出失败: ' + (error.message || '未知错误'))
      throw error
    }
  }
  
  // ==================== 搜索和筛选方法 ====================
  
  const search = async () => {
    pagination.page = 1 // 重置到第一页
    if (viewMode.value === 'tree') {
      await loadTreeData()
    } else {
      await loadTableData()
    }
  }
  
  const resetSearch = () => {
    // 保留项目、周期、版本选择，重置其他条件
    const { projectId, cycleId, versionId } = searchForm
    Object.assign(searchForm, {
      projectId,
      cycleId,
      versionId,
      keyword: '',
      level: null,
      status: '',
      functionType: '',
      complexity: [],
      analysisStatus: [],
      dateRange: null,
      minFunctionPoints: null,
      maxFunctionPoints: null
    })
    
    showAdvancedFilter.value = false
    search()
  }
  
  const setSearchCondition = (key, value) => {
    searchForm[key] = value
    search()
  }
  
  const toggleAdvancedFilter = () => {
    showAdvancedFilter.value = !showAdvancedFilter.value
  }
  
  // ==================== 状态管理方法 ====================
  
  const setViewMode = (mode) => {
    viewMode.value = mode
    selectedRows.value = [] // 切换视图时清空选择
    
    if (mode === 'tree') {
      loadTreeData()
    } else {
      loadTableData()
    }
  }
  
  const setSelection = (rows) => {
    selectedRows.value = rows
  }
  
  const setCurrentNode = (node) => {
    currentNode.value = node
  }
  
  const setProject = async (projectId) => {
    if (searchForm.projectId !== projectId) {
      searchForm.projectId = projectId
      searchForm.cycleId = null
      searchForm.versionId = null
      
      // 清空缓存
      parentOptionsCache.clear()
      
      // 重新加载数据
      if (projectId) {
        await Promise.all([
          loadStats(),
          viewMode.value === 'tree' ? loadTreeData() : loadTableData()
        ])
      } else {
        treeData.value = []
        tableData.value = []
        stats.value = {}
      }
    }
  }
  
  const setCycle = async (cycleId) => {
    if (searchForm.cycleId !== cycleId) {
      searchForm.cycleId = cycleId
      searchForm.versionId = null
      
      // 清空版本相关缓存
      versions.value = []
      activeVersion.value = null
      
      if (cycleId) {
        await loadVersions()
      }
      
      await refresh()
    }
  }
  
  const setVersion = async (versionId) => {
    if (searchForm.versionId !== versionId) {
      searchForm.versionId = versionId
      await refresh()
    }
  }
  
  // ==================== 刷新方法 ====================
  
  const refresh = async () => {
    const promises = []
    
    if (searchForm.projectId) {
      promises.push(loadStats())
      
      if (viewMode.value === 'tree') {
        promises.push(loadTreeData())
      } else {
        promises.push(loadTableData())
      }
    }
    
    await Promise.all(promises)
  }
  
  const refreshStats = () => {
    if (searchForm.projectId) {
      loadStats()
    }
  }
  
  // ==================== 清理方法 ====================
  
  const clearCache = () => {
    parentOptionsCache.clear()
  }
  
  const reset = () => {
    // 重置所有状态
    treeData.value = []
    tableData.value = []
    stats.value = {}
    selectedRows.value = []
    currentNode.value = null
    
    Object.assign(searchForm, {
      projectId: null,
      cycleId: null,
      versionId: null,
      keyword: '',
      level: null,
      status: '',
      functionType: '',
      complexity: [],
      analysisStatus: [],
      dateRange: null,
      minFunctionPoints: null,
      maxFunctionPoints: null
    })
    
    Object.assign(pagination, {
      page: 1,
      pageSize: 20,
      total: 0
    })
    
    viewMode.value = 'tree'
    showAdvancedFilter.value = false
    versions.value = []
    activeVersion.value = null
    analysisTasks.value = []
    Object.keys(analysisProgress).forEach(key => {
      delete analysisProgress[key]
    })
    
    clearCache()
  }
  
  // ==================== 导出store ====================
  
  return {
    // 状态
    treeData,
    tableData,
    stats,
    loading,
    submitting,
    pagination,
    searchForm,
    selectedRows,
    currentNode,
    viewMode,
    showAdvancedFilter,
    versions,
    activeVersion,
    analysisTasks,
    analysisProgress,
    
    // 计算属性
    hasSearchConditions,
    filteredCount,
    selectedCount,
    canAnalyze,
    
    // 数据加载方法
    loadTreeData,
    loadTableData,
    loadStats,
    loadVersions,
    loadParentOptions,
    
    // 数据操作方法
    createNew,
    update,
    remove,
    batchRemove,
    move,
    copy,
    
    // AI分析方法
    analyzeRequirementItem,
    batchAnalyze,
    
    // 版本管理方法
    createVersion,
    switchVersion,
    
    // 导入导出方法
    importExcel,
    exportExcel,
    
    // 搜索和筛选方法
    search,
    resetSearch,
    setSearchCondition,
    toggleAdvancedFilter,
    
    // 状态管理方法
    setViewMode,
    setSelection,
    setCurrentNode,
    setProject,
    setCycle,
    setVersion,
    
    // 刷新方法
    refresh,
    refreshStats,
    
    // 清理方法
    clearCache,
    reset
  }
})