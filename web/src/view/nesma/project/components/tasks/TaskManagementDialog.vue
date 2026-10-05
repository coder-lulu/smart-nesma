<template>
  <el-dialog 
    v-model="visible" 
    title="任务管理" 
    width="1200px" 
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div class="task-management">
      <!-- 任务统计卡片 -->
      <div class="task-stats" v-if="taskStats">
        <el-row :gutter="16">
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon total">
                  <el-icon><Document /></el-icon>
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
                  <div class="stat-label">进行中</div>
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

      <!-- 任务筛选 -->
      <div class="task-filters">
        <el-form :inline="true" :model="filterForm">
          <el-form-item label="任务类型">
            <el-select v-model="filterForm.taskType" placeholder="选择任务类型" clearable>
              <el-option label="需求分析" value="requirement_analysis" />
              <el-option label="描述生成" value="description_generation" />
              <el-option label="流程图生成" value="flowchart_generation" />
              <el-option label="4级需求生成" value="level4_generation" />
              <el-option label="3级需求生成" value="level3_generation" />
              <el-option label="项目分析" value="project_analysis" />
              <el-option label="需求优化" value="requirement_optimization" />
              <el-option label="NESMA评估" value="nesma_evaluation" />
            </el-select>
          </el-form-item>
          <el-form-item label="任务状态">
            <el-select v-model="filterForm.status" placeholder="选择状态" clearable>
              <el-option label="等待中" value="pending" />
              <el-option label="进行中" value="running" />
              <el-option label="已完成" value="completed" />
              <el-option label="失败" value="failed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="创建时间">
            <el-date-picker
              v-model="filterForm.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              format="YYYY-MM-DD HH:mm:ss"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleFilter">筛选</el-button>
            <el-button @click="handleResetFilter">重置</el-button>
            <el-button @click="refreshTasks">刷新</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 任务列表 -->
      <div class="task-table">
        <el-table 
          :data="taskList" 
          style="width: 100%" 
          v-loading="loading"
          empty-text="暂无任务数据"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="ID" label="任务ID" width="80" />
          <el-table-column prop="taskType" label="任务类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getTaskTypeColor(row.taskType)" size="small">
                {{ getTaskTypeText(row.taskType) }}
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
          <el-table-column prop="progress" label="进度" width="120">
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
          <el-table-column label="处理统计" width="150">
            <template #default="{ row }">
              <div class="process-stats">
                <span class="stat-item">总数: {{ row.totalCount || 0 }}</span>
                <span class="stat-item">成功: {{ row.successCount || 0 }}</span>
                <span class="stat-item">失败: {{ row.failedCount || 0 }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="耗时" width="100">
            <template #default="{ row }">
              <span v-if="row.duration">{{ formatDuration(row.duration) }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="updatedAt" label="更新时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.updatedAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" fixed="right">
            <template #default="{ row }">
              <el-button 
                type="primary" 
                size="small" 
                link 
                @click="handleViewProgress(row)"
                v-if="row.status === 'running' || row.status === 'completed'"
              >
                查看进度
              </el-button>
              <el-button 
                type="success" 
                size="small" 
                link 
                @click="handleRetry(row)"
                v-if="row.status === 'failed'"
              >
                重试
              </el-button>
              <el-button 
                type="warning" 
                size="small" 
                link 
                @click="handleCancel(row)"
                v-if="row.status === 'running' || row.status === 'pending'"
              >
                取消
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                link 
                @click="handleViewDetail(row)"
              >
                详情
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                link 
                @click="handleDelete(row)"
                v-if="row.status !== 'running'"
              >
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
          type="danger" 
          @click="handleBatchDelete"
          :disabled="selectedTasks.length === 0"
        >
          批量删除 ({{ selectedTasks.length }})
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document, Loading, Check, Warning, Timer, Close, Search, Refresh
} from '@element-plus/icons-vue'
import {
  getProjectAnalysisTasks,
  getTaskStatistics,
  deleteAnalysisTask,
  batchDeleteAnalysisTasks,
  retryAnalysisTask,
  cancelUnifiedAnalysis
} from '@/api/nesma'

const props = defineProps({
  modelValue: Boolean,
  project: Object
})

const emit = defineEmits(['update:modelValue', 'view-progress', 'close'])

// 响应式数据
const loading = ref(false)
const taskList = ref([])
const selectedTasks = ref([])
const taskStats = ref(null)
const total = ref(0)
const page = ref(1)
const pageSize = ref(5)

// 筛选表单
const filterForm = ref({
  taskType: '',
  status: '',
  dateRange: null
})

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
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
      page: page.value,
      pageSize: pageSize.value,
      ...filterForm.value
    }
    
    // 处理日期范围
    if (filterForm.value.dateRange) {
      params.startDate = filterForm.value.dateRange[0]
      params.endDate = filterForm.value.dateRange[1]
    }
    
    const response = await getProjectAnalysisTasks(null, params)
    
    if (response.code === 0) {
      // 处理新的分页响应格式
      if (response.data && typeof response.data === 'object') {
        taskList.value = response.data.list || []
        total.value = response.data.total || 0
      } else {
        // 兼容旧格式（如果后端还没更新）
        taskList.value = response.data || []
        total.value = taskList.value.length || 0
      }
    }
  } catch (error) {
    ElMessage.error('加载任务列表失败：' + error.message)
  } finally {
    loading.value = false
  }
}

const loadTaskStats = async () => {
  if (!props.project) return
  
  try {
    const response = await getTaskStatistics({
      projectId: props.project.ID
    })
    
    if (response.code === 0) {
      taskStats.value = response.data
    }
  } catch (error) {
    console.error('加载任务统计失败:', error)
  }
}

const handleFilter = () => {
  page.value = 1
  refreshTasks()
}

const handleResetFilter = () => {
  filterForm.value = {
    taskType: '',
    status: '',
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
  console.log('查看任务进度:', row.ID, '任务类型:', row.taskType, '状态:', row.status)
  emit('view-progress', row.ID)
}

const handleRetry = async (row) => {
  try {
    await ElMessageBox.confirm('确定要重试该任务吗？', '确认重试', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await retryAnalysisTask(row.ID)
    ElMessage.success('任务重试成功')
    refreshTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重试任务失败：' + error.message)
    }
  }
}

const handleCancel = async (row) => {
  try {
    await ElMessageBox.confirm('确定要取消该任务吗？', '确认取消', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await cancelUnifiedAnalysis({ task_id: row.ID })
    ElMessage.success('任务取消成功')
    refreshTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消任务失败：' + error.message)
    }
  }
}

const handleViewDetail = (row) => {
  // 可以实现任务详情查看对话框
  ElMessage.info('任务详情功能开发中...')
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该任务吗？删除后无法恢复！', '确认删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    await deleteAnalysisTask(row.ID)
    ElMessage.success('任务删除成功')
    refreshTasks()
    loadTaskStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除任务失败：' + error.message)
    }
  }
}

const handleBatchDelete = async () => {
  if (selectedTasks.value.length === 0) {
    ElMessage.warning('请选择要删除的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedTasks.value.length} 个任务吗？删除后无法恢复！`, '确认批量删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    const taskIds = selectedTasks.value.map(task => task.ID)
    await batchDeleteAnalysisTasks({ task_ids: taskIds })
    
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

// 辅助方法
const getTaskTypeColor = (taskType) => {
  const colorMap = {
    requirement_analysis: 'primary',
    description_generation: 'success',
    flowchart_generation: 'warning',
    level4_generation: 'success',
    level3_generation: 'warning',
    project_analysis: 'primary',
    requirement_optimization: 'success',
    nesma_evaluation: 'warning'
  }
  return colorMap[taskType] || 'info'
}

const getTaskTypeText = (taskType) => {
  const textMap = {
    requirement_analysis: '需求分析',
    description_generation: '描述生成',
    flowchart_generation: '流程图生成',
    level4_generation: '4级需求生成',
    level3_generation: '3级需求生成',
    project_analysis: '项目分析',
    requirement_optimization: '需求优化',
    nesma_evaluation: 'NESMA评估'
  }
  return textMap[taskType] || taskType
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
    running: '进行中',
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
.task-management {
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
    .process-stats {
      display: flex;
      flex-direction: column;
      gap: 2px;
      
      .stat-item {
        font-size: 12px;
        color: #606266;
      }
    }
  }
  
  .pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>