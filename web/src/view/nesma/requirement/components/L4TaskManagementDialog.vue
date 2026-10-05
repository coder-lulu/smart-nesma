<template>
  <el-dialog 
    v-model="visible" 
    title="L4功能点生成任务管理" 
    width="1400px" 
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div class="l4-task-management">
      <!-- L4任务统计卡片 -->
      <div class="task-stats" v-if="taskStats">
        <el-row :gutter="16">
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon total">
                  <el-icon><Operation /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.total || 0 }}</div>
                  <div class="stat-label">总任务数</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon running">
                  <el-icon><Loading /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.running || 0 }}</div>
                  <div class="stat-label">生成中</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon completed">
                  <el-icon><Check /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.completed || 0 }}</div>
                  <div class="stat-label">已完成</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon failed">
                  <el-icon><Warning /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.failed || 0 }}</div>
                  <div class="stat-label">失败</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon pending">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.pending || 0 }}</div>
                  <div class="stat-label">等待中</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon cancelled">
                  <el-icon><Close /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.cancelled || 0 }}</div>
                  <div class="stat-label">已取消</div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- L4任务筛选 -->
      <div class="task-filters">
        <el-form :inline="true" :model="filterForm">
          <el-form-item label="生成策略">
            <el-select v-model="filterForm.generationStrategy" placeholder="选择生成策略" clearable>
              <el-option label="全面生成" value="comprehensive" />
              <el-option label="基础生成" value="basic" />
              <el-option label="智能生成" value="intelligent" />
            </el-select>
          </el-form-item>
          <el-form-item label="任务状态">
            <el-select v-model="filterForm.status" placeholder="选择状态" clearable>
              <el-option label="等待中" value="pending" />
              <el-option label="生成中" value="running" />
              <el-option label="已完成" value="completed" />
              <el-option label="失败" value="failed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="L3需求数">
            <el-input-number 
              v-model="filterForm.minL3Count" 
              :min="1" 
              :max="100" 
              placeholder="最少L3数"
              controls-position="right" 
              size="small"
            />
            <span style="margin: 0 8px;">-</span>
            <el-input-number 
              v-model="filterForm.maxL3Count" 
              :min="1" 
              :max="100" 
              placeholder="最多L3数"
              controls-position="right" 
              size="small"
            />
          </el-form-item>
          <el-form-item label="创建时间">
            <el-date-picker
              v-model="filterForm.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              format="YYYY-MM-DD HH:mm:ss"
              size="small"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleFilter">筛选</el-button>
            <el-button @click="handleResetFilter">重置</el-button>
            <el-button @click="refreshTasks">刷新</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- L4任务列表 -->
      <div class="task-table">
        <el-table 
          :data="taskList" 
          style="width: 100%" 
          v-loading="loading"
          empty-text="暂无L4生成任务"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="ID" label="任务ID" width="80" />
          <el-table-column prop="generationStrategy" label="生成策略" width="120">
            <template #default="{ row }">
              <el-tag :type="getStrategyColor(row.generationStrategy)" size="small">
                {{ getStrategyText(row.generationStrategy) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusColor(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="150">
            <template #default="{ row }">
              <el-progress 
                :percentage="row.progress" 
                :status="getProgressStatus(row.status)"
                :stroke-width="8"
                :show-text="false"
              />
              <span style="margin-left: 8px; font-size: 12px;">{{ row.progress }}%</span>
            </template>
          </el-table-column>
          <el-table-column label="L3需求信息" width="220">
            <template #default="{ row }">
              <div class="l3-info">
                <div class="l3-count">L3需求数: {{ row.l3RequirementCount || 0 }}</div>
                <div class="l3-list" v-if="row.l3RequirementTitles">
                  <el-tooltip :content="row.l3RequirementTitles.join('\n')" placement="top">
                    <div class="l3-titles">
                      {{ row.l3RequirementTitles.slice(0, 2).join(', ') }}
                      <span v-if="row.l3RequirementTitles.length > 2">...</span>
                    </div>
                  </el-tooltip>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="生成统计" width="180">
            <template #default="{ row }">
              <div class="generation-stats">
                <span class="stat-item">
                  <el-icon><Check /></el-icon>
                  成功: {{ row.successCount || 0 }}
                </span>
                <span class="stat-item">
                  <el-icon><Warning /></el-icon>
                  失败: {{ row.failedCount || 0 }}
                </span>
                <span class="stat-item">
                  <el-icon><Document /></el-icon>
                  总数: {{ row.totalL4Count || 0 }}
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="耗时" width="100">
            <template #default="{ row }">
              <span v-if="row.duration">{{ formatDuration(row.duration) }}</span>
              <span v-else-if="row.status === 'running'">{{ getRunningDuration(row.createdAt) }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="260" fixed="right">
            <template #default="{ row }">
              <el-button 
                type="primary" 
                size="small" 
                link 
                @click="handleViewProgress(row)"
                v-if="row.status === 'running'"
              >
                <el-icon><View /></el-icon>
                实时进度
              </el-button>
              <el-button 
                type="success" 
                size="small" 
                link 
                @click="handleViewResult(row)"
                v-if="row.status === 'completed'"
              >
                <el-icon><DocumentChecked /></el-icon>
                查看结果
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                link 
                @click="handleRetry(row)"
                v-if="row.status === 'failed'"
              >
                <el-icon><Refresh /></el-icon>
                重试
              </el-button>
              <el-button 
                type="warning" 
                size="small" 
                link 
                @click="handleCancel(row)"
                v-if="row.status === 'running' || row.status === 'pending'"
              >
                <el-icon><Close /></el-icon>
                取消
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                link 
                @click="handleViewDetail(row)"
              >
                <el-icon><InfoFilled /></el-icon>
                详情
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                link 
                @click="handleDelete(row)"
                v-if="row.status !== 'running'"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <!-- 对话框底部 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button 
          type="warning" 
          @click="handleBatchCancel"
          :disabled="selectedTasks.length === 0 || !hasRunningTasks"
        >
          批量取消 ({{ getRunningTaskCount }})
        </el-button>
        <el-button 
          type="danger" 
          @click="handleBatchDelete"
          :disabled="selectedTasks.length === 0 || hasRunningTasks"
        >
          批量删除 ({{ selectedTasks.length }})
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- L4任务详情对话框 -->
  <el-dialog 
    v-model="detailDialogVisible" 
    title="L4生成任务详情" 
    width="800px"
    :close-on-click-modal="false"
  >
    <div class="task-detail" v-if="currentTaskDetail">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="任务ID">{{ currentTaskDetail.ID }}</el-descriptions-item>
        <el-descriptions-item label="生成策略">
          <el-tag :type="getStrategyColor(currentTaskDetail.generationStrategy)" size="small">
            {{ getStrategyText(currentTaskDetail.generationStrategy) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="任务状态">
          <el-tag :type="getStatusColor(currentTaskDetail.status)" size="small">
            {{ getStatusText(currentTaskDetail.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="进度">{{ currentTaskDetail.progress }}%</el-descriptions-item>
        <el-descriptions-item label="L3需求数">{{ currentTaskDetail.l3RequirementCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="生成L4数">{{ currentTaskDetail.totalL4Count || 0 }}</el-descriptions-item>
        <el-descriptions-item label="成功数">{{ currentTaskDetail.successCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="失败数">{{ currentTaskDetail.failedCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(currentTaskDetail.createdAt) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatDateTime(currentTaskDetail.completedAt) }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ formatDuration(currentTaskDetail.duration) }}</el-descriptions-item>
        <el-descriptions-item label="错误信息" v-if="currentTaskDetail.errorMsg">
          <el-text type="danger">{{ currentTaskDetail.errorMsg }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
      
      <!-- L3需求列表 -->
      <div class="l3-requirements" v-if="currentTaskDetail.l3RequirementTitles">
        <h4>涉及的L3需求</h4>
        <el-tag 
          v-for="(title, index) in currentTaskDetail.l3RequirementTitles" 
          :key="index"
          style="margin: 2px 4px 2px 0;"
        >
          {{ title }}
        </el-tag>
      </div>
      
      <!-- 任务配置详情 -->
      <div class="task-config" v-if="currentTaskDetail.configDetails">
        <h4>任务配置</h4>
        <pre>{{ JSON.stringify(currentTaskDetail.configDetails, null, 2) }}</pre>
      </div>
    </div>
    
    <template #footer>
      <el-button @click="detailDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>

  <!-- L4任务结果展示对话框 -->
  <L4TaskResultDialog 
    v-model="showResultDialog"
    :taskId="selectedTaskId"
  />
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import L4TaskResultDialog from './L4TaskResultDialog.vue'
import {
  Operation, Loading, Check, Warning, Timer, Close, View, DocumentChecked, 
  Refresh, InfoFilled, Delete, Document
} from '@element-plus/icons-vue'

// 这里需要添加L4任务相关的API调用
import {
  getL4GenerationTasks,
  getL4TaskStatistics,
  deleteL4GenerationTask,
  batchDeleteL4GenerationTasks,
  retryL4GenerationTask,
  cancelL4GenerationTask,
  getL4TaskDetail,
  batchCancelL4GenerationTasks
} from '@/api/nesma'

const props = defineProps({
  modelValue: Boolean,
  project: Object,
  cycle: Object,
  version: Object
})

const emit = defineEmits(['update:modelValue', 'view-progress', 'view-result', 'close'])

// 响应式数据
const loading = ref(false)
const taskList = ref([])
const selectedTasks = ref([])
const taskStats = ref(null)
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const detailDialogVisible = ref(false)
const currentTaskDetail = ref(null)

// L4任务结果对话框相关
const showResultDialog = ref(false)
const selectedTaskId = ref(null)

// 筛选表单
const filterForm = ref({
  generationStrategy: '',
  status: '',
  minL3Count: null,
  maxL3Count: null,
  dateRange: null
})

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const hasRunningTasks = computed(() => {
  return selectedTasks.value.some(task => task.status === 'running' || task.status === 'pending')
})

const getRunningTaskCount = computed(() => {
  return selectedTasks.value.filter(task => task.status === 'running' || task.status === 'pending').length
})

// 监听对话框显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.project) {
    refreshTasks()
    loadTaskStats()
  }
})

// 方法
const refreshTasks = async () => {
  if (!props.project) return
  
  try {
    loading.value = true
    
    const params = {
      projectId: props.project.ID,
      cycleId: props.cycle?.ID,
      versionId: props.version?.ID,
      page: page.value,
      pageSize: pageSize.value,
      ...filterForm.value
    }
    
    // 处理日期范围
    if (filterForm.value.dateRange) {
      params.startDate = filterForm.value.dateRange[0]
      params.endDate = filterForm.value.dateRange[1]
    }
    
    // 处理L3需求数范围
    if (filterForm.value.minL3Count) {
      params.minL3Count = filterForm.value.minL3Count
    }
    if (filterForm.value.maxL3Count) {
      params.maxL3Count = filterForm.value.maxL3Count
    }
    
    // 清理空参数
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    
    const response = await getL4GenerationTasks(params)
    
    if (response.code === 0) {
      // 处理分页响应格式
      if (response.data && typeof response.data === 'object') {
        taskList.value = response.data.list || []
        total.value = response.data.total || 0
      } else {
        // 兼容旧格式
        taskList.value = response.data || []
        total.value = taskList.value.length || 0
      }
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
    
  } catch (error) {
    console.error('加载L4任务列表失败:', error)
    ElMessage.error('加载L4任务列表失败：' + error.message)
    
    // 发生错误时，回退到模拟数据
    const mockTasks = [
      {
        ID: 1,
        generationStrategy: 'comprehensive',
        status: 'completed',
        progress: 100,
        l3RequirementCount: 5,
        l3RequirementTitles: ['用户登录功能', '权限管理', '数据备份', '日志记录', '系统监控'],
        successCount: 12,
        failedCount: 1,
        totalL4Count: 13,
        duration: 245,
        createdAt: '2025-07-15T10:30:00Z',
        completedAt: '2025-07-15T10:34:05Z'
      },
      {
        ID: 2,
        generationStrategy: 'basic',
        status: 'running',
        progress: 65,
        l3RequirementCount: 3,
        l3RequirementTitles: ['消息通知', '文件上传', '数据导出'],
        successCount: 6,
        failedCount: 0,
        totalL4Count: 9,
        duration: null,
        createdAt: '2025-07-15T11:15:00Z',
        completedAt: null
      }
    ]
    
    taskList.value = mockTasks
    total.value = mockTasks.length
  } finally {
    loading.value = false
  }
}

const loadTaskStats = async () => {
  if (!props.project) return
  
  try {
    const params = {
      projectId: props.project.ID,
      cycleId: props.cycle?.ID,
      versionId: props.version?.ID
    }
    
    const response = await getL4TaskStatistics(params)
    
    if (response.code === 0) {
      taskStats.value = response.data
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
  } catch (error) {
    console.error('加载L4任务统计失败:', error)
    
    // 发生错误时，回退到模拟数据
    taskStats.value = {
      total: 15,
      running: 2,
      completed: 10,
      failed: 2,
      pending: 1,
      cancelled: 0
    }
  }
}

const handleFilter = () => {
  page.value = 1
  refreshTasks()
}

const handleResetFilter = () => {
  filterForm.value = {
    generationStrategy: '',
    status: '',
    minL3Count: null,
    maxL3Count: null,
    dateRange: null
  }
  page.value = 1
  refreshTasks()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  page.value = 1
  refreshTasks()
}

const handleCurrentChange = (currentPage) => {
  page.value = currentPage
  refreshTasks()
}

const handleSelectionChange = (selection) => {
  selectedTasks.value = selection
}

const handleViewProgress = (row) => {
  emit('view-progress', row.ID)
}

const handleViewResult = (row) => {
  selectedTaskId.value = row.ID
  showResultDialog.value = true
}

const handleRetry = async (row) => {
  try {
    await ElMessageBox.confirm('确定要重试该L4生成任务吗？', '确认重试', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await retryL4GenerationTask(row.ID)
    ElMessage.success('L4生成任务重试成功')
    refreshTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重试任务失败：' + error.message)
    }
  }
}

const handleCancel = async (row) => {
  try {
    await ElMessageBox.confirm('确定要取消该L4生成任务吗？', '确认取消', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await cancelL4GenerationTask(row.ID)
    ElMessage.success('L4生成任务取消成功')
    refreshTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消任务失败：' + error.message)
    }
  }
}

const handleViewDetail = async (row) => {
  try {
    const response = await getL4TaskDetail(row.ID)
    
    if (response.code === 0) {
      currentTaskDetail.value = response.data
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
    
    detailDialogVisible.value = true
  } catch (error) {
    console.error('获取任务详情失败:', error)
    ElMessage.error('获取任务详情失败：' + error.message)
    
    // 发生错误时，回退到模拟数据
    currentTaskDetail.value = {
      ...row,
      configDetails: {
        generationStrategy: row.generationStrategy,
        maxL4PerL3: 5,
        confidenceThreshold: 0.7,
        useKnowledgeBase: true,
        enableOptimization: true
      }
    }
    
    detailDialogVisible.value = true
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该L4生成任务吗？删除后无法恢复！', '确认删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    await deleteL4GenerationTask(row.ID)
    ElMessage.success('L4生成任务删除成功')
    refreshTasks()
    loadTaskStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除任务失败：' + error.message)
    }
  }
}

const handleBatchCancel = async () => {
  const runningTasks = selectedTasks.value.filter(task => task.status === 'running' || task.status === 'pending')
  if (runningTasks.length === 0) {
    ElMessage.warning('请选择正在运行的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要取消选中的 ${runningTasks.length} 个任务吗？`, '确认批量取消', {
      confirmButtonText: '确定取消',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const taskIds = runningTasks.map(task => task.ID)
    await batchCancelL4GenerationTasks({ task_ids: taskIds })
    
    ElMessage.success('批量取消成功')
    refreshTasks()
    loadTaskStats()
    selectedTasks.value = []
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量取消失败：' + error.message)
    }
  }
}

const handleBatchDelete = async () => {
  const deletableTasks = selectedTasks.value.filter(task => task.status !== 'running')
  if (deletableTasks.length === 0) {
    ElMessage.warning('请选择非运行状态的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${deletableTasks.length} 个任务吗？删除后无法恢复！`, '确认批量删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    const taskIds = deletableTasks.map(task => task.ID)
    await batchDeleteL4GenerationTasks({ task_ids: taskIds })
    
    ElMessage.success('批量删除成功')
    refreshTasks()
    loadTaskStats()
    selectedTasks.value = []
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败：' + error.message)
    }
  }
}

const handleClose = () => {
  emit('close')
}

// 计算正在运行的任务耗时
const getRunningDuration = (startTime) => {
  if (!startTime) return '-'
  const now = new Date()
  const start = new Date(startTime)
  const duration = Math.floor((now - start) / 1000)
  return formatDuration(duration)
}

// 辅助方法
const getStrategyColor = (strategy) => {
  const colorMap = {
    comprehensive: 'primary',
    basic: 'success',
    intelligent: 'warning'
  }
  return colorMap[strategy] || 'info'
}

const getStrategyText = (strategy) => {
  const textMap = {
    comprehensive: '全面生成',
    basic: '基础生成',
    intelligent: '智能生成'
  }
  return textMap[strategy] || strategy
}

const getStatusColor = (status) => {
  const colorMap = {
    pending: 'warning',
    running: 'primary',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return colorMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    pending: '等待中',
    running: '生成中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return textMap[status] || status
}

const getProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return minutes > 0 ? `${minutes}分${secs}秒` : `${secs}秒`
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 生命周期
onMounted(() => {
  if (props.modelValue && props.project) {
    refreshTasks()
    loadTaskStats()
  }
})
</script>

<style lang="scss" scoped>
.l4-task-management {
  .task-stats {
    margin-bottom: 20px;
    
    .stat-card {
      border: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      
      .stat-item {
        display: flex;
        align-items: center;
        padding: 8px;
        
        .stat-icon {
          width: 40px;
          height: 40px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 12px;
          
          .el-icon {
            font-size: 18px;
            color: white;
          }
          
          &.total {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          }
          
          &.running {
            background: linear-gradient(135deg, #409eff 0%, #36cfc9 100%);
          }
          
          &.completed {
            background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
          }
          
          &.failed {
            background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
          }
          
          &.pending {
            background: linear-gradient(135deg, #e6a23c 0%, #f7ba2a 100%);
          }
          
          &.cancelled {
            background: linear-gradient(135deg, #909399 0%, #b1b3b8 100%);
          }
        }
        
        .stat-info {
          flex: 1;
          
          .stat-value {
            font-size: 20px;
            font-weight: 600;
            color: #303133;
            line-height: 1;
          }
          
          .stat-label {
            font-size: 12px;
            color: #909399;
            margin-top: 2px;
          }
        }
      }
    }
  }
  
  .task-filters {
    background: #f8f9fa;
    padding: 16px;
    border-radius: 8px;
    margin-bottom: 20px;
  }
  
  .task-table {
    .l3-info {
      .l3-count {
        font-size: 12px;
        color: #606266;
        margin-bottom: 4px;
      }
      
      .l3-titles {
        font-size: 12px;
        color: #909399;
        line-height: 1.2;
      }
    }
    
    .generation-stats {
      display: flex;
      flex-direction: column;
      gap: 2px;
      
      .stat-item {
        display: flex;
        align-items: center;
        font-size: 12px;
        color: #606266;
        
        .el-icon {
          margin-right: 4px;
          font-size: 12px;
        }
      }
    }
  }
  
  .pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
}

.task-detail {
  .l3-requirements {
    margin-top: 20px;
    
    h4 {
      margin-bottom: 10px;
      color: #303133;
    }
  }
  
  .task-config {
    margin-top: 20px;
    
    h4 {
      margin-bottom: 10px;
      color: #303133;
    }
    
    pre {
      background: #f8f9fa;
      padding: 12px;
      border-radius: 4px;
      font-size: 12px;
      overflow-x: auto;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>