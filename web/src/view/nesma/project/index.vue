<template>
  <div class="project-management">
    <!-- 工具栏组件 -->
    <ProjectToolbar
      :selected-count="selectedRows.length"
      :user-options="userOptions"
      :loading="loading"
      @create="handleCreate"
      @batch-import="handleBatchImport"
      @batch-delete="handleBatchDelete"
      @export="handleExport"
      @search="handleSearch"
      @reset="handleReset"
    />

    <!-- 统计卡片组件 -->
    <ProjectStatsCards
      :stats="projectStats"
      :loading="statsLoading"
      @view-all="handleViewAll"
      @filter-active="handleFilterActive"
      @filter-completed="handleFilterCompleted"
      @filter-recent="handleFilterRecent"
      @refresh="refreshStats"
    />

    <!-- 表格组件 -->
    <ProjectTable
      :table-data="tableData"
      :loading="loading"
      :total="total"
      :current-page="page"
      :page-size="pageSize"
      @view="handleView"
      @edit="handleEdit"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
      @size-change="handleSizeChange"
      @current-change="handleCurrentChange"
      @action="handleTableAction"
      @refresh="refreshTableData"
    />

    <!-- 项目表单对话框 -->
    <ProjectFormDialog
      v-model="formDialogVisible"
      :form-data="currentProject"
      :mode="formMode"
      :user-options="userOptions"
      @save="handleSave"
      @cancel="handleFormCancel"
    />

    <!-- 项目详情对话框 -->
    <ProjectDetailDialog
      v-model="detailDialogVisible"
      :project="currentProject"
      @edit="handleEdit"
      @close="handleDetailClose"
    />

    <!-- 周期管理对话框 -->
    <CycleManagementDialog
      v-model="cycleDialogVisible"
      :project="currentProject"
      :active-cycle-id="activeCycleId"
      @save="handleCycleSave"
      @close="handleCycleClose"
      @update-active-cycle="handleActiveeCycleUpdate"
    />

    <!-- 分析进度对话框 -->
    <AnalysisProgressDialog
      v-model="analysisDialogVisible"
      :task-id="currentAnalysisTask"
      :project="currentProject"
      @complete="handleAnalysisComplete"
      @close="handleAnalysisClose"
    />

    <!-- 任务管理对话框 -->
    <TaskManagementDialog
      v-model="taskDialogVisible"
      :project="currentProject"
      @view-progress="handleViewTaskProgress"
      @close="handleTaskClose"
    />

    <!-- 分析选项对话框 -->
    <AnalysisOptionsDialog
      v-model="analysisOptionsVisible"
      :project="currentProject"
      @confirm="handleAnalysisOptionsConfirm"
      @cancel="handleAnalysisOptionsCancel"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'

// 组件导入
import ProjectToolbar from './components/toolbar/ProjectToolbar.vue'
import ProjectStatsCards from './components/stats/ProjectStatsCards.vue'
import ProjectTable from './components/table/ProjectTable.vue'
import ProjectFormDialog from './components/forms/ProjectFormDialog.vue'
import ProjectDetailDialog from './components/ProjectDetailDialog.vue'
import CycleManagementDialog from './components/cycles/CycleManagementDialog.vue'
import AnalysisProgressDialog from './components/analysis/AnalysisProgressDialog.vue'
import AnalysisOptionsDialog from './components/analysis/AnalysisOptionsDialog.vue'
import TaskManagementDialog from './components/tasks/TaskManagementDialog.vue'

// API 导入
import { 
  getNesmaProjectList,
  createNesmaProject,
  updateNesmaProject,
  deleteNesmaProject,
  getNesmaProjectStats,
  executeProjectAnalysis,
  getUnifiedAnalysisProgress,
  getDetailedAnalysisResult,
  // 任务管理相关API
  getProjectAnalysisTasks,
  getAnalysisTaskDetail,
  resumeTaskProgress,
  deleteAnalysisTask,
  retryAnalysisTask,
  getTaskStatistics
} from '@/api/nesma'

// 初始化router
const router = useRouter()

// 响应式数据
const loading = ref(false)
const statsLoading = ref(false)
const tableData = ref([])
const selectedRows = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

// 统计数据
const projectStats = ref({
  totalProjects: 0,
  activeProjects: 0,
  completedProjects: 0,
  recentProjects: 0,
  pausedProjects: 0,
  archivedProjects: 0,
  domainDistribution: {},
  monthlyTrend: []
})

// 对话框状态
const formDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const cycleDialogVisible = ref(false)
const analysisDialogVisible = ref(false)
const analysisOptionsVisible = ref(false)
const taskDialogVisible = ref(false)

// 当前操作的数据
const currentProject = ref({})
const formMode = ref('create') // create | edit
const currentAnalysisTask = ref(null)
const activeCycleId = ref(null)

// 用户选项（用于筛选）
const userOptions = ref([])

// 搜索参数
const searchParams = ref({})

// 计算属性
const selectedIds = computed(() => selectedRows.value.map(row => row.ID))

// 工具栏事件处理
const handleCreate = () => {
  currentProject.value = {}
  formMode.value = 'create'
  formDialogVisible.value = true
}

const handleBatchImport = () => {
  ElMessage.info('批量导入功能开发中...')
}

const handleBatchDelete = async () => {
  if (selectedIds.value.length === 0) {
    ElMessage.warning('请选择要删除的项目')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedIds.value.length} 个项目吗？`,
      '批量删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    loading.value = true
    // 批量删除逻辑
    await Promise.all(selectedIds.value.map(id => deleteNesmaProject({ ID: id })))
    
    ElMessage.success('批量删除成功')
    await refreshTableData()
    await refreshStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败: ' + error.message)
    }
  } finally {
    loading.value = false
  }
}

const handleExport = () => {
  ElMessage.info('导出功能开发中...')
}

const handleSearch = (params) => {
  searchParams.value = params
  page.value = 1
  refreshTableData()
}

const handleReset = () => {
  searchParams.value = {}
  page.value = 1
  refreshTableData()
}

// 统计卡片事件处理
const handleViewAll = () => {
  handleReset()
}

const handleFilterActive = () => {
  handleSearch({ status: 'active' })
}

const handleFilterCompleted = () => {
  handleSearch({ status: 'archived' })
}

const handleFilterRecent = () => {
  const thirtyDaysAgo = new Date()
  thirtyDaysAgo.setDate(thirtyDaysAgo.getDate() - 30)
  
  const endDate = new Date()
  
  // 使用简单的日期格式 YYYY-MM-DD，如果后端要求完整时间，会自动解析为00:00:00
  const formatDate = (date) => {
    return date.toISOString().split('T')[0] + 'T00:00:00Z'
  }
  
  console.log('筛选近30天项目，日期范围:', {
    startDate: formatDate(thirtyDaysAgo),
    endDate: formatDate(endDate)
  })
  
  handleSearch({ 
    startDate: formatDate(thirtyDaysAgo),
    endDate: formatDate(endDate)
  })
}

// 表格事件处理
const handleView = (row) => {
  currentProject.value = row
  detailDialogVisible.value = true
}

const handleEdit = (row) => {
  currentProject.value = { ...row }
  formMode.value = 'edit'
  formDialogVisible.value = true
}

const handleSelectionChange = (rows) => {
  selectedRows.value = rows
}

const handleSortChange = (sortInfo) => {
  // 处理排序逻辑
  refreshTableData()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  page.value = 1
  refreshTableData()
}

const handleCurrentChange = (currentPage) => {
  page.value = currentPage
  refreshTableData()
}

const handleTableAction = ({ action, row, data }) => {
  switch (action) {
    case 'requirements':
      handleRequirements(row)
      break
    case 'cycles':
      handleCycleManagement(row)
      break
    case 'analysis':
      handleOneClickAnalysis(row)
      break
    case 'tasks':
      handleTaskManagement(row)
      break
    case 'duplicate':
      handleDuplicate(row)
      break
    case 'export':
      handleExportProject(row)
      break
    case 'archive':
      handleArchive(row)
      break
    case 'restore':
      handleRestore(row)
      break
    case 'delete':
      handleDelete(row)
      break
    case 'export-table':
      handleExportTable(data)
      break
    default:
      console.warn('未知操作:', action)
  }
}

// 业务操作方法
const handleRequirements = (row) => {
  // 使用路由跳转到需求管理页面
  router.push({
    path: '/layout/nesma/requirement',
    query: {
      projectId: row.ID,
      activeCycleId: row.activeCycleId,
    }
  })
}

const handleCycleManagement = (row) => {
  currentProject.value = row
  activeCycleId.value = row.activeCycleId
  cycleDialogVisible.value = true
}

const handleOneClickAnalysis = (row) => {
  if (!row.activeCycleId) {
    ElMessage.warning('请先设置活跃周期')
    return
  }

  if (row.status !== 'active') {
    ElMessage.warning('只有活跃状态的项目才能进行分析')
    return
  }

  // 设置当前项目并弹出选择对话框
  currentProject.value = {
    ...row,
    activeCycleName: row.activeCycleName || '默认周期'
  }
  analysisOptionsVisible.value = true
}

// 处理分析选项确认
const handleAnalysisOptionsConfirm = async (analysisConfig) => {
  try {
    analysisOptionsVisible.value = false
    
    // 显示加载提示
    const loadingMessage = ElMessage({
      message: '正在启动智能分析任务...',
      type: 'info',
      duration: 0
    })

    // 调用统一分析API
    const response = await executeProjectAnalysis(analysisConfig)
    
    // 关闭加载提示
    loadingMessage.close()
    
    // 启动分析任务监控
    currentAnalysisTask.value = response.data.task_id || response.data.taskId
    analysisDialogVisible.value = true
    
    // 构建成功消息
    const selectedItems = []
    if (analysisConfig.includeDescription) selectedItems.push('需求描述优化')
    if (analysisConfig.includeLevel4) selectedItems.push('L4功能点生成')
    if (analysisConfig.includeMermaid) selectedItems.push('流程图生成')
    
    ElMessage.success(`智能分析任务已启动！\n分析项目: ${selectedItems.join('、')}\n任务ID: ${currentAnalysisTask.value}`)
    
  } catch (error) {
    console.error('启动分析失败:', error)
    
    // 分类处理错误
    let errorMessage = '启动分析失败'
    if (error.response) {
      switch (error.response.status) {
        case 400:
          errorMessage = '请求参数错误: ' + (error.response.data?.message || '参数验证失败')
          break
        case 403:
          errorMessage = '没有权限执行此操作'
          break
        case 404:
          errorMessage = '分析服务不可用，请稍后重试'
          break
        case 500:
          errorMessage = '服务器内部错误，请联系管理员'
          break
        default:
          errorMessage = error.response.data?.message || '未知错误'
      }
    } else if (error.code === 'NETWORK_ERROR') {
      errorMessage = '网络连接失败，请检查网络设置'
    }
    
    ElMessage.error(errorMessage)
  }
}

// 处理分析选项取消
const handleAnalysisOptionsCancel = () => {
  analysisOptionsVisible.value = false
  currentProject.value = {}
}

const handleDuplicate = async (row) => {
  try {
    const duplicatedProject = {
      ...row,
      name: `${row.name} - 副本`,
      ID: undefined,
      createdAt: undefined,
      updatedAt: undefined
    }
    
    await createNesmaProject(duplicatedProject)
    ElMessage.success('项目复制成功')
    refreshTableData()
  } catch (error) {
    ElMessage.error('项目复制失败: ' + error.message)
  }
}

const handleExportProject = (row) => {
  ElMessage.info(`导出项目 "${row.name}" 的数据...`)
}

const handleArchive = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要归档项目 "${row.name}" 吗？归档后项目将不可编辑。`,
      '归档确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    await updateNesmaProject({ ...row, status: 'archived' })
    ElMessage.success('项目归档成功')
    refreshTableData()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('项目归档失败: ' + error.message)
    }
  }
}

const handleRestore = async (row) => {
  try {
    await updateNesmaProject({ ...row, status: 'active' })
    ElMessage.success('项目恢复成功')
    refreshTableData()
  } catch (error) {
    ElMessage.error('项目恢复失败: ' + error.message)
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除项目 "${row.name}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'danger'
      }
    )

    await deleteNesmaProject({ ID: row.ID })
    refreshTableData()
    refreshStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('项目删除失败: ' + error.message)
    }
  }
}

const handleExportTable = (selectedData) => {
  const data = selectedData.length > 0 ? selectedData : tableData.value
  ElMessage.info(`导出 ${data.length} 条项目数据...`)
}

// 表单对话框事件处理
const handleSave = async (formData) => {
  try {
    if (formMode.value === 'create') {
      await createNesmaProject(formData)
      ElMessage.success('项目创建成功')
    } else {
      await updateNesmaProject(formData)
      ElMessage.success('项目更新成功')
    }
    
    formDialogVisible.value = false
    await refreshTableData()
    await refreshStats()
  } catch (error) {
    ElMessage.error(`项目${formMode.value === 'create' ? '创建' : '更新'}失败: ` + error.message)
  }
}

const handleFormCancel = () => {
  formDialogVisible.value = false
}

// 详情对话框事件处理
const handleDetailClose = () => {
  detailDialogVisible.value = false
}

// 周期管理对话框事件处理
const handleCycleSave = () => {
  cycleDialogVisible.value = false
  refreshTableData()
}

const handleCycleClose = () => {
  cycleDialogVisible.value = false
}

// 处理激活周期更新
const handleActiveeCycleUpdate = (newActiveCycleId) => {
  // 更新本地的激活周期ID
  activeCycleId.value = newActiveCycleId
  
  // 更新当前项目对象中的激活周期ID
  if (currentProject.value) {
    currentProject.value.activeCycleId = newActiveCycleId
  }
  
  // 刷新项目列表以确保数据同步
  refreshTableData()
}

// 分析进度对话框事件处理
const handleAnalysisComplete = () => {
  analysisDialogVisible.value = false
  refreshTableData()
  ElMessage.success('需求分析完成')
}

const handleAnalysisClose = () => {
  analysisDialogVisible.value = false
}

// 任务管理相关方法
const handleTaskManagement = (row) => {
  currentProject.value = row
  taskDialogVisible.value = true
}

const handleViewTaskProgress = (taskId) => {
  // 从任务管理对话框查看任务进度
  console.log('查看任务进度:', taskId)
  
  // 设置当前分析任务ID
  currentAnalysisTask.value = taskId
  
  // 关闭任务管理对话框
  taskDialogVisible.value = false
  
  // 打开分析进度对话框
  analysisDialogVisible.value = true
  
  ElMessage.info(`正在加载任务 ${taskId} 的进度信息...`)
}

const handleTaskClose = () => {
  taskDialogVisible.value = false
}

// 数据获取方法
const refreshTableData = async () => {
  try {
    loading.value = true
    
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      ...searchParams.value
    }
    
    const response = await getNesmaProjectList(params)
    
    tableData.value = response.data.list || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取项目列表失败: ' + error.message)
    tableData.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const refreshStats = async () => {
  try {
    statsLoading.value = true
    const response = await getNesmaProjectStats()
    const rawData = response.data || {}
    
    // 处理状态统计数据
    const statusStats = rawData.statusStats || []
    const processedStats = {
      totalProjects: rawData.totalProjects || 0,
      recentProjects: rawData.recentProjects || 0,
      activeProjects: 0,
      completedProjects: 0,
      pausedProjects: 0,
      archivedProjects: 0,
      domainDistribution: {},
      monthlyTrend: rawData.monthlyTrend || []
    }
    
    // 从statusStats数组中提取各状态的项目数量
    statusStats.forEach(stat => {
      switch(stat.status) {
        case 'active':
          processedStats.activeProjects = stat.count || 0
          break
        case 'completed':
          processedStats.completedProjects = stat.count || 0
          break
        case 'paused':
          processedStats.pausedProjects = stat.count || 0
          break
        case 'archived':
          processedStats.archivedProjects = stat.count || 0
          break
      }
    })
    
    // 处理领域分布数据
    const domainStats = rawData.domainStats || []
    domainStats.forEach(stat => {
      processedStats.domainDistribution[stat.domain] = stat.count || 0
    })
    
    projectStats.value = processedStats
    console.log('处理后的统计数据:', processedStats)
  } catch (error) {
    console.error('获取项目统计失败:', error)
    // 使用默认值
  } finally {
    statsLoading.value = false
  }
}

// 生命周期
onMounted(() => {
  refreshTableData()
  refreshStats()
})
</script>

<style lang="scss" scoped>
.project-management {
  padding: 20px;
  background: #f5f7fa;
  min-height: 100vh;

  // 组件间距
  > * {
    margin-bottom: 16px;

    &:last-child {
      margin-bottom: 0;
    }
  }
}
</style>