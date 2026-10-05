<template>
  <div class="workflow-management">
    <!-- 统计卡片 -->
    <div class="stats-section">
      <div class="stat-card blue">
        <div class="stat-icon">
          <el-icon><Document /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ totalCount }}</div>
          <div class="stat-label">工作流总数</div>
        </div>
      </div>
      <div class="stat-card green">
        <div class="stat-icon">
          <el-icon><CircleCheckFilled /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ activeCount }}</div>
          <div class="stat-label">活跃流程</div>
        </div>
      </div>
      <div class="stat-card orange">
        <div class="stat-icon">
          <el-icon><Timer /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ executionCount }}</div>
          <div class="stat-label">执行次数</div>
        </div>
      </div>
      <div class="stat-card red">
        <div class="stat-icon">
          <el-icon><Setting /></el-icon>
        </div>
        <div class="stat-content">
          <div class="stat-value">{{ filteredWorkflows.length }}</div>
          <div class="stat-label">当前显示</div>
        </div>
      </div>
    </div>

    <!-- 操作工具栏 -->
    <div class="action-toolbar">
      <div class="toolbar-left">
        <el-button type="primary" :icon="Plus" @click="handleCreateWorkflow">
          新建工作流
        </el-button>
        <el-button :icon="Download" @click="handleBatchExport">
          导出工作流
        </el-button>
        <el-button type="danger" :icon="Delete" @click="handleBatchDelete" :disabled="!selectedWorkflows.length">
          批量删除
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-dropdown trigger="click">
          <el-button>
            更多操作 <el-icon class="el-icon--right"><MoreFilled /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="handleBatchActivate">批量激活</el-dropdown-item>
              <el-dropdown-item @click="handleBatchDeactivate">批量停用</el-dropdown-item>
              <el-dropdown-item divided @click="refreshWorkflowList">刷新列表</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        <el-select v-model="categoryFilter" placeholder="工作流类别" clearable @change="handleFilter">
          <el-option label="全部类别" value="" />
          <el-option label="需求分析" value="requirement_analysis" />
          <el-option label="NESMA评估" value="nesma_evaluation" />
          <el-option label="文档生成" value="document_generation" />
          <el-option label="质量检查" value="quality_check" />
        </el-select>
        <el-select v-model="statusFilter" placeholder="工作流状态" clearable @change="handleFilter">
          <el-option label="全部状态" value="" />
          <el-option label="活跃" value="active" />
          <el-option label="停用" value="inactive" />
        </el-select>
        <el-input
          v-model="searchKeyword"
          placeholder="搜索工作流名称"
          :prefix-icon="Search"
          clearable
          @input="handleSearch"
          style="width: 200px"
        />
        <el-button type="primary" :icon="Search">
          搜索
        </el-button>
      </div>
    </div>

    <!-- 分类标签页 -->
    <div class="category-tabs">
      <div class="tab-item" :class="{ active: activeTab === 'all' }" @click="activeTab = 'all'">
        工作流列表
      </div>
      <div class="tab-item" :class="{ active: activeTab === 'analytics' }" @click="activeTab = 'analytics'">
        执行统计
      </div>
      <div class="tab-item" :class="{ active: activeTab === 'templates' }" @click="activeTab = 'templates'">
        模板管理
      </div>
    </div>

    <!-- 标签页内容 -->
    <div class="workflow-container">

      <!-- 工作流列表 -->
      <div v-if="activeTab === 'all'" class="tab-content">
        <div class="workflow-table-container" v-loading="loading">
          <el-table 
            :data="filteredWorkflows" 
            style="width: 100%" 
            @selection-change="handleSelectionChange"
            stripe
            row-key="workflowId"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="name" label="工作流标题" min-width="200">
              <template #default="{ row }">
                <div class="workflow-title">
                  <el-link type="primary" @click="handleViewWorkflow(row)">
                    {{ row.name }}
                  </el-link>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="category" label="工作流类别" width="120">
              <template #default="{ row }">
                <el-tag size="small" effect="light" type="primary">
                  {{ getCategoryName(row.category) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="domain" label="工作流领域" width="120">
              <template #default="{ row }">
                <el-tag size="small" effect="light" type="success">
                  {{ row.domain || '通用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="tags" label="标签" width="150">
              <template #default="{ row }">
                <div class="tags-wrapper">
                  <el-tag 
                    v-for="tag in (row.tags || ['自动化', '流程'])" 
                    :key="tag" 
                    size="small" 
                    effect="light"
                    style="margin-right: 4px;"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="confidence" label="置信度" width="120">
              <template #default="{ row }">
                <div class="confidence-wrapper">
                  <el-progress 
                    :percentage="row.confidence || 100" 
                    :show-text="false" 
                    :stroke-width="6"
                    :color="getConfidenceColor(row.confidence || 100)"
                  />
                  <span class="confidence-text">{{ row.confidence || 100 }}%</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button size="small" type="primary" @click="handleViewWorkflow(row)">
                  查看
                </el-button>
                <el-button size="small" type="success" @click="handleEditWorkflow(row)">
                  编辑
                </el-button>
                <el-button size="small" type="danger" @click="handleDeleteWorkflow(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <!-- 空状态 -->
          <div v-if="!loading && filteredWorkflows.length === 0" class="empty-state">
            <div class="empty-illustration">
              <el-icon size="64"><DocumentCopy /></el-icon>
            </div>
            <h3>暂无工作流</h3>
            <p>创建您的第一个AI工作流，开始自动化业务流程</p>
            <el-button type="primary" :icon="Plus" @click="handleCreateWorkflow">
              创建工作流
            </el-button>
          </div>
        </div>
      </div>

      <!-- 执行统计 -->
      <div v-else-if="activeTab === 'analytics'" class="tab-content">
        <div class="analytics-container">
          <!-- 统计图表区域 -->
          <div class="analytics-charts">
            <div class="chart-section">
              <div class="section-header">
                <h3>执行趋势分析</h3>
                <el-select v-model="analyticsTimeRange" size="small">
                  <el-option label="最近7天" value="7d" />
                  <el-option label="最近30天" value="30d" />
                  <el-option label="最近90天" value="90d" />
                </el-select>
              </div>
              <div class="chart-placeholder">
                <el-icon size="48"><TrendCharts /></el-icon>
                <p>执行趋势图表</p>
                <el-button type="primary" size="small">查看详细报告</el-button>
              </div>
            </div>
            
            <div class="chart-section">
              <div class="section-header">
                <h3>工作流性能分析</h3>
              </div>
              <div class="performance-metrics">
                <div class="metric-item">
                  <div class="metric-label">平均执行时间</div>
                  <div class="metric-value">2.5分钟</div>
                </div>
                <div class="metric-item">
                  <div class="metric-label">成功率</div>
                  <div class="metric-value success">98.5%</div>
                </div>
                <div class="metric-item">
                  <div class="metric-label">错误率</div>
                  <div class="metric-value error">1.5%</div>
                </div>
              </div>
            </div>
          </div>

          <!-- 执行记录表格 -->
          <div class="execution-records">
            <div class="section-header">
              <h3>最近执行记录</h3>
              <el-button type="primary" size="small" @click="handleViewAllExecutions">
                查看全部
              </el-button>
            </div>
            <el-table :data="recentExecutions" style="width: 100%">
              <el-table-column prop="executionId" label="执行ID" width="120" />
              <el-table-column prop="workflowName" label="工作流名称" min-width="200" />
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag 
                    :type="row.status === 'success' ? 'success' : row.status === 'failed' ? 'danger' : 'warning'"
                    size="small"
                  >
                    {{ getExecutionStatusName(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="duration" label="执行时长" width="120" />
              <el-table-column prop="startTime" label="开始时间" width="180">
                <template #default="{ row }">
                  {{ formatTime(row.startTime) }}
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100">
                <template #default="{ row }">
                  <el-button size="small" type="primary" @click="handleViewExecution(row)">
                    查看
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
      </div>

      <!-- 模板管理 -->
      <div v-else-if="activeTab === 'templates'" class="tab-content">
        <div class="templates-container">
          <div class="template-categories">
            <div class="category-card" v-for="category in templateCategories" :key="category.id">
              <div class="category-icon">
                <el-icon size="32"><component :is="category.icon" /></el-icon>
              </div>
              <div class="category-info">
                <h4>{{ category.name }}</h4>
                <p>{{ category.description }}</p>
                <div class="category-stats">
                  <span>{{ category.templateCount }} 个模板</span>
                </div>
              </div>
              <div class="category-actions">
                <el-button type="primary" size="small" @click="handleViewTemplates(category)">
                  查看模板
                </el-button>
                <el-button size="small" @click="handleCreateTemplate(category)">
                  创建模板
                </el-button>
              </div>
            </div>
          </div>

          <!-- 推荐模板 -->
          <div class="recommended-templates">
            <div class="section-header">
              <h3>推荐模板</h3>
              <el-button type="primary" size="small" @click="handleViewAllTemplates">
                查看全部
              </el-button>
            </div>
            <div class="template-grid">
              <div class="template-card" v-for="template in recommendedTemplates" :key="template.id">
                <div class="template-header">
                  <div class="template-icon">
                    <el-icon><component :is="template.icon" /></el-icon>
                  </div>
                  <el-tag size="small" :type="template.category === 'official' ? 'success' : 'info'">
                    {{ template.category === 'official' ? '官方' : '社区' }}
                  </el-tag>
                </div>
                <h4>{{ template.name }}</h4>
                <p>{{ template.description }}</p>
                <div class="template-stats">
                  <span><el-icon><Download /></el-icon> {{ template.downloads }}</span>
                  <span><el-icon><View /></el-icon> {{ template.views }}</span>
                </div>
                <div class="template-actions">
                  <el-button type="primary" size="small" @click="handleUseTemplate(template)">
                    使用模板
                  </el-button>
                  <el-button size="small" @click="handlePreviewTemplate(template)">
                    预览
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

      <!-- 分页 -->
      <div class="pagination-wrapper" v-if="totalCount > 0 && activeTab === 'all'">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[12, 24, 48, 96]"
          :total="totalCount"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <!-- 对话框组件 -->
    <CreateWorkflowDialog 
      v-model="showCreateDialog"
      @success="handleCreateSuccess"
    />
    
    <WorkflowDesignerDrawer
      v-model="showDesignerDrawer"
      :workflow="selectedWorkflow"
      @save="handleDesignerSave"
    />
    
    <WorkflowDetailDialog
      v-model="showDetailDialog"
      :workflow="selectedWorkflow"
    />
    
    <ExecutionDetailDialog
      v-model="showExecutionDialog"
      :execution="currentExecution"
    />
    
    <ExecutionHistoryDialog
      v-model="showHistoryDialog"
      :workflow-id="selectedWorkflowId"
    />
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Plus, Refresh, Search, Grid, List, EditPen, VideoPlay, MoreFilled, 
  View, CopyDocument, Clock, Download, Delete, Document, Files,
  FolderOpened, Timer, CircleCheckFilled, VideoPause, 
  TrendCharts, Setting, DocumentCopy
} from '@element-plus/icons-vue'
import { 
  getWorkflowList, 
  deleteWorkflow,
  executeWorkflow as executeWorkflowAPI,
  getWorkflowExecution,
  createWorkflow as createWorkflowAPI
} from '@/api/nesma'
import CreateWorkflowDialog from './components/CreateWorkflowDialog.vue'
import WorkflowDesignerDrawer from './components/WorkflowDesignerDrawer.vue'
import WorkflowDetailDialog from './components/WorkflowDetailDialog.vue'
import ExecutionDetailDialog from './components/ExecutionDetailDialog.vue'
import ExecutionHistoryDialog from './components/ExecutionHistoryDialog.vue'

// 响应式数据
const loading = ref(false)
const searchKeyword = ref('')
const statusFilter = ref('')
const categoryFilter = ref('')
const currentPage = ref(1)
const pageSize = ref(12)
const totalCount = ref(0)
const workflowList = ref([])
const selectedWorkflow = ref(null)
const selectedWorkflowId = ref('')
const currentExecution = ref(null)
const selectedWorkflows = ref([])
const activeTab = ref('all')
const analyticsTimeRange = ref('30d')

// 模拟数据
const recentExecutions = ref([
  {
    executionId: 'exec-001',
    workflowName: '测试工作流',
    status: 'success',
    duration: '2分15秒',
    startTime: new Date(Date.now() - 2 * 60 * 60 * 1000)
  }
])

const templateCategories = ref([
  {
    id: 1,
    name: '数据处理',
    description: '用于数据清洗、转换和分析的工作流模板',
    icon: 'TrendCharts',
    templateCount: 12
  },
  {
    id: 2,
    name: '文档生成',
    description: '自动生成各类文档和报告的模板',
    icon: 'Document',
    templateCount: 8
  },
  {
    id: 3,
    name: '业务自动化',
    description: '常见业务流程自动化模板',
    icon: 'Setting',
    templateCount: 15
  }
])

const recommendedTemplates = ref([
  {
    id: 1,
    name: 'NESMA评估模板',
    description: '标准的NESMA功能点评估流程',
    category: 'official',
    icon: 'TrendCharts',
    downloads: 245,
    views: 1200
  },
  {
    id: 2,
    name: '需求分析模板',
    description: '系统需求分析和整理模板',
    category: 'community',
    icon: 'Files',
    downloads: 156,
    views: 890
  }
])

// 对话框状态
const showCreateDialog = ref(false)
const showDesignerDrawer = ref(false)
const showDetailDialog = ref(false)
const showExecutionDialog = ref(false)
const showHistoryDialog = ref(false)

// 计算属性
const filteredWorkflows = computed(() => {
  let filtered = workflowList.value
  
  // 搜索过滤
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase()
    filtered = filtered.filter(workflow => 
      workflow.name.toLowerCase().includes(keyword) ||
      (workflow.description && workflow.description.toLowerCase().includes(keyword))
    )
  }
  
  // 状态过滤
  if (statusFilter.value) {
    filtered = filtered.filter(workflow => workflow.status === statusFilter.value)
  }
  
  // 分类过滤
  if (categoryFilter.value) {
    filtered = filtered.filter(workflow => workflow.category === categoryFilter.value)
  }
  
  return filtered
})

const activeCount = computed(() => {
  return workflowList.value.filter(w => w.status === 'active').length
})

const executionCount = computed(() => {
  return workflowList.value.reduce((sum, w) => sum + (w.executionCount || 0), 0)
})

// 页面方法
const loadWorkflowList = async () => {
  loading.value = true
  try {
    const response = await getWorkflowList({
      page: currentPage.value,
      pageSize: pageSize.value
    })
    workflowList.value = response.data.workflows || []
    totalCount.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('获取工作流列表失败')
  } finally {
    loading.value = false
  }
}

const refreshWorkflowList = () => {
  loadWorkflowList()
}

const handleSearch = () => {
  // 搜索时重置到第一页
  currentPage.value = 1
}

const handleFilter = () => {
  // 筛选时重置到第一页
  currentPage.value = 1
}

const handleSizeChange = (size) => {
  pageSize.value = size
  loadWorkflowList()
}

const handleCurrentChange = (page) => {
  currentPage.value = page
  loadWorkflowList()
}

// 工作流操作
const handleCreateWorkflow = () => {
  showCreateDialog.value = true
}

const handleViewWorkflow = (workflow) => {
  selectedWorkflow.value = workflow
  showDetailDialog.value = true
}

const handleEditWorkflow = (workflow) => {
  selectedWorkflow.value = workflow
  showDesignerDrawer.value = true
}

const handleExecuteWorkflow = async (workflow) => {
  try {
    ElMessageBox.confirm('确定要执行此工作流吗？', '执行确认', {
      type: 'warning'
    }).then(async () => {
      const response = await executeWorkflowAPI(workflow.workflowId, {
        inputData: {
          requirement: '用户需求',
          priority: 'normal'
        }
      })
      
      currentExecution.value = response.data
      showExecutionDialog.value = true
      
      ElMessage.success('工作流已开始执行')
    })
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('执行工作流失败')
    }
  }
}

const handleWorkflowAction = async (command) => {
  const [action, workflowId] = command.split('_')
  const workflow = workflowList.value.find(w => w.workflowId === workflowId)
  
  switch (action) {
    case 'view':
      handleViewWorkflow(workflow)
      break
    case 'copy':
      await handleCopyWorkflow(workflow)
      break
    case 'history':
      selectedWorkflowId.value = workflowId
      showHistoryDialog.value = true
      break
    case 'export':
      await handleExportWorkflow(workflow)
      break
    case 'delete':
      await handleDeleteWorkflow(workflow)
      break
  }
}

const handleCopyWorkflow = async (workflow) => {
  try {
    // 创建副本
    const copyData = {
      name: `${workflow.name} - 副本`,
      description: workflow.description,
      category: workflow.category,
      definition: workflow.definition
    }
    
    selectedWorkflow.value = { ...workflow, ...copyData }
    showCreateDialog.value = true
  } catch (error) {
    ElMessage.error('复制工作流失败')
  }
}

const handleExportWorkflow = async (workflow) => {
  try {
    const exportData = {
      name: workflow.name,
      description: workflow.description,
      category: workflow.category,
      definition: workflow.definition,
      version: workflow.version
    }
    
    const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${workflow.name}.json`
    a.click()
    URL.revokeObjectURL(url)
    
    ElMessage.success('工作流配置已导出')
  } catch (error) {
    ElMessage.error('导出工作流失败')
  }
}

const handleDeleteWorkflow = async (workflow) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除工作流"${workflow.name}"吗？此操作不可恢复。`,
      '删除确认',
      { type: 'warning' }
    )
    
    await deleteWorkflow(workflow.workflowId)
    ElMessage.success('工作流已删除')
    loadWorkflowList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除工作流失败')
    }
  }
}

// 批量操作
const handleBatchExport = () => {
  if (selectedWorkflows.value.length === 0) {
    ElMessage.warning('请选择要导出的工作流')
    return
  }
  ElMessage.success('导出功能开发中...')
}

const handleBatchDelete = async () => {
  if (selectedWorkflows.value.length === 0) {
    ElMessage.warning('请选择要删除的工作流')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedWorkflows.value.length} 个工作流吗？此操作不可恢复。`,
      '批量删除确认',
      { type: 'warning' }
    )
    
    // 这里应该调用批量删除API
    ElMessage.success('批量删除功能开发中...')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

const handleBatchActivate = () => {
  if (selectedWorkflows.value.length === 0) {
    ElMessage.warning('请选择要激活的工作流')
    return
  }
  ElMessage.success('批量激活功能开发中...')
}

const handleBatchDeactivate = () => {
  if (selectedWorkflows.value.length === 0) {
    ElMessage.warning('请选择要停用的工作流')
    return
  }
  ElMessage.success('批量停用功能开发中...')
}

const handleSelectionChange = (selection) => {
  selectedWorkflows.value = selection
}

// 工具函数
const getConfidenceColor = (confidence) => {
  if (confidence >= 80) return '#67c23a'
  if (confidence >= 60) return '#e6a23c'
  return '#f56c6c'
}

// 事件处理
const handleCreateSuccess = () => {
  loadWorkflowList()
}

const handleDesignerSave = () => {
  loadWorkflowList()
}

// 工具函数
const getCategoryName = (category) => {
  const categoryMap = {
    requirement_analysis: '需求分析',
    nesma_evaluation: 'NESMA评估',
    document_generation: '文档生成',
    quality_check: '质量检查'
  }
  return categoryMap[category] || '未分类'
}

const getCategoryIcon = (category) => {
  const iconMap = {
    requirement_analysis: Files,
    nesma_evaluation: TrendCharts,
    document_generation: Document,
    quality_check: CircleCheckFilled
  }
  return iconMap[category] || Document
}

const getCategoryIconClass = (category) => {
  const classMap = {
    requirement_analysis: 'category-icon primary',
    nesma_evaluation: 'category-icon success', 
    document_generation: 'category-icon warning',
    quality_check: 'category-icon danger'
  }
  return classMap[category] || 'category-icon'
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString()
}

// 执行统计相关方法
const getExecutionStatusName = (status) => {
  const statusMap = {
    'success': '成功',
    'failed': '失败',
    'running': '运行中',
    'pending': '等待中'
  }
  return statusMap[status] || '未知'
}

const handleViewAllExecutions = () => {
  ElMessage.info('查看全部执行记录功能开发中')
}

const handleViewExecution = (execution) => {
  ElMessage.info(`查看执行记录 ${execution.executionId}`)
}

// 模板管理相关方法
const handleViewTemplates = (category) => {
  ElMessage.info(`查看 ${category.name} 模板`)
}

const handleCreateTemplate = (category) => {
  ElMessage.info(`创建 ${category.name} 模板`)
}

const handleViewAllTemplates = () => {
  ElMessage.info('查看全部模板功能开发中')
}

const handleUseTemplate = (template) => {
  ElMessage.info(`使用模板: ${template.name}`)
}

const handlePreviewTemplate = (template) => {
  ElMessage.info(`预览模板: ${template.name}`)
}

// 生命周期
onMounted(() => {
  loadWorkflowList()
})
</script>

<style scoped>
.workflow-management {
  min-height: 100vh;
  background: #f5f5f5;
  padding: 20px;
}

/* 统计卡片 */
.stats-section {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 20px;
  margin-bottom: 20px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
  transition: all 0.3s ease;
}

.stat-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
}

.stat-card.blue .stat-icon {
  background: #3b82f6;
  color: white;
}

.stat-card.green .stat-icon {
  background: #10b981;
  color: white;
}

.stat-card.orange .stat-icon {
  background: #f59e0b;
  color: white;
}

.stat-card.red .stat-icon {
  background: #ef4444;
  color: white;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 20px;
  flex-shrink: 0;
}

.stat-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #1f2937;
  line-height: 1;
}

.stat-label {
  font-size: 14px;
  color: #6b7280;
  font-weight: 500;
}

/* 操作工具栏 */
.action-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: white;
  border-radius: 8px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
}

.toolbar-left {
  display: flex;
  gap: 12px;
  align-items: center;
}

.toolbar-right {
  display: flex;
  gap: 12px;
  align-items: center;
}

/* 分类标签页 */
.category-tabs {
  display: flex;
  background: white;
  border-radius: 8px;
  margin-bottom: 20px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.tab-item {
  padding: 12px 24px;
  cursor: pointer;
  border-bottom: 3px solid transparent;
  color: #6b7280;
  font-weight: 500;
  transition: all 0.3s ease;
  position: relative;
}

.tab-item:first-child {
  border-top-left-radius: 8px;
  border-bottom-left-radius: 8px;
}

.tab-item:last-child {
  border-top-right-radius: 8px;
  border-bottom-right-radius: 8px;
}

.tab-item:hover {
  color: #3b82f6;
  background: #f8fafc;
}

.tab-item.active {
  color: #3b82f6;
  border-bottom-color: #3b82f6;
  background: #f8fafc;
}

.workflow-container {
  padding: 0;
}

.tab-content {
  background: white;
  border-radius: 8px;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  border: 1px solid #e5e7eb;
}

/* 执行统计样式 */
.analytics-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.analytics-charts {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.chart-section {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #e5e7eb;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.section-header h3 {
  margin: 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.chart-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px;
  color: #6b7280;
  text-align: center;
  gap: 12px;
}

.chart-placeholder .el-icon {
  color: #9ca3af;
}

.performance-metrics {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
}

.metric-item {
  text-align: center;
  padding: 16px;
  background: white;
  border-radius: 6px;
  border: 1px solid #e5e7eb;
}

.metric-label {
  font-size: 12px;
  color: #6b7280;
  margin-bottom: 4px;
}

.metric-value {
  font-size: 18px;
  font-weight: 600;
  color: #1f2937;
}

.metric-value.success {
  color: #10b981;
}

.metric-value.error {
  color: #ef4444;
}

.execution-records {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #e5e7eb;
}

/* 模板管理样式 */
.templates-container {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.template-categories {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 20px;
}

.category-card {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #e5e7eb;
  transition: all 0.3s ease;
}

.category-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.category-icon {
  color: #3b82f6;
  margin-bottom: 12px;
}

.category-info h4 {
  margin: 0 0 8px 0;
  color: #1f2937;
  font-size: 16px;
  font-weight: 600;
}

.category-info p {
  margin: 0 0 12px 0;
  color: #6b7280;
  font-size: 14px;
  line-height: 1.5;
}

.category-stats {
  font-size: 12px;
  color: #9ca3af;
  margin-bottom: 16px;
}

.category-actions {
  display: flex;
  gap: 8px;
}

.recommended-templates {
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #e5e7eb;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 16px;
}

.template-card {
  background: white;
  border-radius: 8px;
  padding: 16px;
  border: 1px solid #e5e7eb;
  transition: all 0.3s ease;
}

.template-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
}

.template-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.template-icon {
  color: #3b82f6;
}

.template-card h4 {
  margin: 0 0 8px 0;
  color: #1f2937;
  font-size: 14px;
  font-weight: 600;
}

.template-card p {
  margin: 0 0 12px 0;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.4;
}

.template-stats {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
  font-size: 12px;
  color: #9ca3af;
}

.template-stats span {
  display: flex;
  align-items: center;
  gap: 4px;
}

.template-actions {
  display: flex;
  gap: 8px;
}

/* 工作流表格 */
.workflow-table-container {
  background: white;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  overflow: hidden;
}

.workflow-title {
  font-weight: 500;
  color: #1f2937;
}

.tags-wrapper {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.confidence-wrapper {
  display: flex;
  align-items: center;
  gap: 8px;
}

.confidence-text {
  font-size: 12px;
  color: #6b7280;
  min-width: 36px;
}

:deep(.el-table) {
  border: none;
}

:deep(.el-table th) {
  background: #f8fafc;
  color: #374151;
  font-weight: 600;
  border-bottom: 1px solid #e5e7eb;
}

:deep(.el-table td) {
  border-bottom: 1px solid #f3f4f6;
  padding: 12px 0;
}

:deep(.el-table--striped .el-table__body tr.el-table__row--striped td) {
  background: #f8fafc;
}

:deep(.el-table__body tr:hover > td) {
  background: #f0f9ff !important;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: #6b7280;
}

.empty-illustration {
  margin-bottom: 16px;
  color: #d1d5db;
}

.empty-state h3 {
  margin: 0 0 8px 0;
  font-size: 16px;
  color: #374151;
  font-weight: 500;
}

.empty-state p {
  margin: 0 0 20px 0;
  color: #6b7280;
  font-size: 14px;
}

/* 分页 */
.pagination-wrapper {
  display: flex;
  justify-content: center;
  padding: 20px 0;
}

:deep(.danger-item) {
  color: #dc2626;
}

/* 响应式设计 */
@media (max-width: 1024px) {
  .stats-section {
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
    gap: 16px;
  }
  
  .action-toolbar {
    flex-direction: column;
    gap: 12px;
    align-items: stretch;
  }
  
  .toolbar-left,
  .toolbar-right {
    justify-content: center;
    flex-wrap: wrap;
  }
}

@media (max-width: 768px) {
  .workflow-management {
    padding: 12px;
  }
  
  .stats-section {
    grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
    gap: 12px;
    margin-bottom: 16px;
  }
  
  .stat-card {
    padding: 16px;
    gap: 12px;
  }
  
  .stat-icon {
    width: 40px;
    height: 40px;
    font-size: 18px;
  }
  
  .stat-value {
    font-size: 24px;
  }
  
  .stat-label {
    font-size: 12px;
  }
  
  .action-toolbar {
    padding: 12px;
  }
  
  .toolbar-left,
  .toolbar-right {
    flex-direction: column;
    gap: 8px;
    width: 100%;
  }
  
  .category-tabs {
    flex-direction: column;
  }
  
  .tab-item {
    padding: 8px 16px;
    text-align: center;
    border-bottom: 1px solid #e5e7eb;
    border-radius: 0;
  }
  
  .tab-item:first-child {
    border-radius: 8px 8px 0 0;
  }
  
  .tab-item:last-child {
    border-radius: 0 0 8px 8px;
    border-bottom: none;
  }
  
  .tab-item.active {
    border-bottom: 1px solid #e5e7eb;
    border-left: 3px solid #3b82f6;
  }
}
</style> 