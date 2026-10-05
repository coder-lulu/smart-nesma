<template>
  <el-dialog
    v-model="dialogVisible"
    title="执行历史"
    width="1200px"
    :before-close="handleClose"
    class="execution-history-dialog"
  >
    <div class="history-content">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="toolbar-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索执行ID..."
            :prefix-icon="Search"
            clearable
            @change="handleSearch"
            style="width: 250px"
          />
          <el-select 
            v-model="statusFilter" 
            placeholder="状态筛选" 
            clearable 
            @change="handleSearch"
            style="width: 120px; margin-left: 12px"
          >
            <el-option label="全部" value="" />
            <el-option label="运行中" value="running" />
            <el-option label="已完成" value="completed" />
            <el-option label="已失败" value="failed" />
            <el-option label="已取消" value="cancelled" />
          </el-select>
          <el-date-picker
            v-model="dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            format="YYYY-MM-DD"
            value-format="YYYY-MM-DD"
            @change="handleSearch"
            style="width: 240px; margin-left: 12px"
          />
        </div>
        <div class="toolbar-right">
          <el-button @click="refreshData" :icon="Refresh">刷新</el-button>
          <el-button @click="exportExecutions" :icon="Download">导出</el-button>
        </div>
      </div>

      <!-- 统计信息 -->
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-icon total">
            <el-icon size="20"><Operation /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.total || 0 }}</div>
            <div class="stat-label">总执行次数</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon success">
            <el-icon size="20"><CircleCheck /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.successful || 0 }}</div>
            <div class="stat-label">成功</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon failed">
            <el-icon size="20"><CircleClose /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.failed || 0 }}</div>
            <div class="stat-label">失败</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon running">
            <el-icon size="20"><Loading /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ stats.running || 0 }}</div>
            <div class="stat-label">运行中</div>
          </div>
        </div>
      </div>

      <!-- 执行记录表格 -->
      <div class="executions-table">
        <el-table 
          :data="executionList" 
          v-loading="loading"
          @row-click="handleRowClick"
          row-key="executionId"
          :default-sort="{ prop: 'startTime', order: 'descending' }"
        >
          <el-table-column 
            prop="executionId" 
            label="执行ID" 
            width="200"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <div class="execution-id-cell">
                <el-text class="execution-id">{{ row.executionId }}</el-text>
                <el-button 
                  size="small" 
                  text 
                  @click.stop="copyToClipboard(row.executionId)"
                  style="margin-left: 8px"
                >
                  复制
                </el-button>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column 
            prop="status" 
            label="状态" 
            width="100"
            :filters="statusFilters"
            :filter-method="filterStatus"
          >
            <template #default="{ row }">
              <el-tag :type="getStatusTagType(row.status)">
                {{ getStatusLabel(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          
          <el-table-column 
            prop="currentStep" 
            label="当前步骤" 
            width="150"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <div class="current-step">
                <span>{{ row.currentStep || '-' }}</span>
                <el-progress 
                  v-if="row.status === 'running'" 
                  :percentage="getStepProgress(row)"
                  :stroke-width="6"
                  :show-text="false"
                  style="margin-top: 4px"
                />
              </div>
            </template>
          </el-table-column>
          
          <el-table-column 
            prop="startTime" 
            label="开始时间" 
            width="180"
            sortable
          >
            <template #default="{ row }">
              {{ formatTime(row.startTime) }}
            </template>
          </el-table-column>
          
          <el-table-column 
            prop="endTime" 
            label="结束时间" 
            width="180"
            sortable
          >
            <template #default="{ row }">
              {{ row.endTime ? formatTime(row.endTime) : '-' }}
            </template>
          </el-table-column>
          
          <el-table-column 
            prop="duration" 
            label="耗时" 
            width="100"
            sortable
          >
            <template #default="{ row }">
              <span class="duration">{{ formatDuration(row.duration) }}</span>
            </template>
          </el-table-column>
          
          <el-table-column 
            prop="inputData" 
            label="输入数据" 
            width="200"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <div class="input-data">
                <el-tooltip :content="formatInputData(row.inputData)" placement="top">
                  <el-text truncated>{{ getInputDataSummary(row.inputData) }}</el-text>
                </el-tooltip>
              </div>
            </template>
          </el-table-column>
          
          <el-table-column 
            label="错误信息" 
            width="200"
            show-overflow-tooltip
          >
            <template #default="{ row }">
              <div class="error-message" v-if="row.status === 'failed' && row.errorMessage">
                <el-text type="danger" truncated>{{ row.errorMessage }}</el-text>
              </div>
              <span v-else>-</span>
            </template>
          </el-table-column>
          
          <el-table-column 
            label="操作" 
            width="180" 
            fixed="right"
          >
            <template #default="{ row }">
              <div class="action-buttons">
                <el-button 
                  size="small" 
                  @click.stop="viewExecutionDetail(row)"
                  :icon="View"
                >
                  详情
                </el-button>
                <el-button 
                  v-if="row.status === 'running'" 
                  size="small" 
                  type="danger" 
                  @click.stop="cancelExecution(row)"
                  :icon="CircleClose"
                >
                  取消
                </el-button>
                <el-dropdown @command="handleCommand" trigger="click">
                  <el-button size="small" :icon="MoreFilled" />
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="`logs:${row.executionId}`" :icon="Document">
                        查看日志
                      </el-dropdown-item>
                      <el-dropdown-item :command="`rerun:${row.executionId}`" :icon="Refresh">
                        重新执行
                      </el-dropdown-item>
                      <el-dropdown-item :command="`export:${row.executionId}`" :icon="Download">
                        导出结果
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination" v-if="total > 0">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>

      <!-- 空状态 -->
      <div v-if="executionList.length === 0 && !loading" class="empty-state">
        <el-empty description="暂无执行记录" />
      </div>
    </div>

    <!-- 执行详情对话框 -->
    <ExecutionDetailDialog
      v-model:visible="detailDialogVisible"
      :execution-id="selectedExecutionId"
    />
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Search, Refresh, Download, Operation, CircleCheck, CircleClose, Loading,
  View, MoreFilled, Document
} from '@element-plus/icons-vue'

import { 
  getWorkflowExecutions,
  cancelWorkflowExecution as cancelWorkflowExecutionAPI
} from '@/api/nesma'

import ExecutionDetailDialog from './ExecutionDetailDialog.vue'

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  workflowId: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['update:visible'])

// 响应式数据
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const loading = ref(false)
const executionList = ref([])
const stats = ref({})
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)

// 搜索和筛选
const searchKeyword = ref('')
const statusFilter = ref('')
const dateRange = ref(null)

// 对话框状态
const detailDialogVisible = ref(false)
const selectedExecutionId = ref('')

// 状态筛选选项
const statusFilters = [
  { text: '运行中', value: 'running' },
  { text: '已完成', value: 'completed' },
  { text: '已失败', value: 'failed' },
  { text: '已取消', value: 'cancelled' }
]

// 获取执行记录
const getExecutions = async () => {
  if (!props.workflowId) return
  
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      status: statusFilter.value,
      startDate: dateRange.value?.[0],
      endDate: dateRange.value?.[1]
    }
    
    const res = await getWorkflowExecutions(props.workflowId, params)
    if (res.code === 0) {
      executionList.value = res.data?.executions || []
      total.value = res.data?.total || 0
      
      // 计算统计信息
      calculateStats()
    }
  } catch (error) {
    ElMessage.error('获取执行记录失败')
  } finally {
    loading.value = false
  }
}

// 计算统计信息
const calculateStats = () => {
  const executions = executionList.value
  const total = executions.length
  const successful = executions.filter(e => e.status === 'completed').length
  const failed = executions.filter(e => e.status === 'failed').length
  const running = executions.filter(e => e.status === 'running').length
  
  stats.value = {
    total,
    successful,
    failed,
    running
  }
}

// 搜索处理
const handleSearch = () => {
  currentPage.value = 1
  getExecutions()
}

// 刷新数据
const refreshData = () => {
  getExecutions()
}

// 分页处理
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  getExecutions()
}

const handleCurrentChange = (page) => {
  currentPage.value = page
  getExecutions()
}

// 表格行点击
const handleRowClick = (row) => {
  viewExecutionDetail(row)
}

// 查看执行详情
const viewExecutionDetail = (execution) => {
  selectedExecutionId.value = execution.executionId
  detailDialogVisible.value = true
}

// 取消执行
const cancelExecution = async (execution) => {
  try {
    await ElMessageBox.confirm(
      `确定要取消执行 "${execution.executionId}" 吗？`,
      '取消确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await cancelWorkflowExecutionAPI(execution.executionId)
    ElMessage.success('取消成功')
    getExecutions()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消失败')
    }
  }
}

// 下拉菜单命令处理
const handleCommand = (command) => {
  const [action, executionId] = command.split(':')
  
  switch (action) {
    case 'logs':
      viewExecutionLogs(executionId)
      break
    case 'rerun':
      rerunExecution(executionId)
      break
    case 'export':
      exportExecutionResult(executionId)
      break
  }
}

// 查看执行日志
const viewExecutionLogs = (executionId) => {
  ElMessage.info('查看日志功能开发中')
}

// 重新执行
const rerunExecution = (executionId) => {
  ElMessage.info('重新执行功能开发中')
}

// 导出执行结果
const exportExecutionResult = (executionId) => {
  ElMessage.info('导出结果功能开发中')
}

// 导出执行记录
const exportExecutions = () => {
  ElMessage.info('导出功能开发中')
}

// 状态筛选
const filterStatus = (value, row) => {
  return row.status === value
}

// 复制到剪贴板
const copyToClipboard = async (text) => {
  try {
    await navigator.clipboard.writeText(text)
    ElMessage.success('已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 工具函数
const getStatusTagType = (status) => {
  const typeMap = {
    completed: 'success',
    failed: 'danger',
    running: 'primary',
    cancelled: 'warning',
    pending: 'info'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    completed: '已完成',
    failed: '已失败',
    running: '运行中',
    cancelled: '已取消',
    pending: '等待中'
  }
  return labelMap[status] || status
}

const getStepProgress = (execution) => {
  // 简单的进度计算逻辑
  if (execution.status === 'completed') return 100
  if (execution.status === 'failed') return 0
  if (execution.currentStep) {
    // 根据当前步骤计算进度
    const steps = ['开始', '需求分析', 'NESMA评估', '文档生成', '结束']
    const currentIndex = steps.indexOf(execution.currentStep)
    return currentIndex >= 0 ? (currentIndex / steps.length) * 100 : 50
  }
  return 10
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleString('zh-CN')
}

const formatDuration = (duration) => {
  if (!duration) return '-'
  
  const hours = Math.floor(duration / 3600)
  const minutes = Math.floor((duration % 3600) / 60)
  const seconds = duration % 60
  
  if (hours > 0) {
    return `${hours}h ${minutes}m ${seconds}s`
  } else if (minutes > 0) {
    return `${minutes}m ${seconds}s`
  } else {
    return `${seconds}s`
  }
}

const getInputDataSummary = (inputData) => {
  if (!inputData) return '-'
  
  try {
    const data = typeof inputData === 'string' ? JSON.parse(inputData) : inputData
    const keys = Object.keys(data)
    if (keys.length === 0) return '空数据'
    if (keys.length === 1) return `${keys[0]}: ${data[keys[0]]}`
    return `${keys.length}个参数`
  } catch (error) {
    return '无效数据'
  }
}

const formatInputData = (inputData) => {
  if (!inputData) return '无输入数据'
  
  try {
    const data = typeof inputData === 'string' ? JSON.parse(inputData) : inputData
    return JSON.stringify(data, null, 2)
  } catch (error) {
    return '无效的JSON数据'
  }
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  // 重置搜索条件
  searchKeyword.value = ''
  statusFilter.value = ''
  dateRange.value = null
  currentPage.value = 1
}

// 监听props变化
watch(() => props.workflowId, (newId) => {
  if (newId && props.visible) {
    getExecutions()
  }
})

watch(() => props.visible, (visible) => {
  if (visible && props.workflowId) {
    getExecutions()
  }
})
</script>

<style scoped>
.execution-history-dialog {
  --el-dialog-border-radius: 16px;
}

.history-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20px 24px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 12px;
}

.stats-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 16px;
}

.stat-card {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  border: 1px solid #f0f2f5;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.stat-icon.total { background: linear-gradient(135deg, #667eea, #764ba2); }
.stat-icon.success { background: linear-gradient(135deg, #43e97b, #38f9d7); }
.stat-icon.failed { background: linear-gradient(135deg, #fa709a, #fee140); }
.stat-icon.running { background: linear-gradient(135deg, #4facfe, #00f2fe); }

.stat-content {
  flex: 1;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #1f2937;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 13px;
  color: #6b7280;
  font-weight: 500;
}

.executions-table {
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.execution-id-cell {
  display: flex;
  align-items: center;
}

.execution-id {
  font-family: 'Monaco', 'Consolas', monospace;
  background: #f3f4f6;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  max-width: 120px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.current-step {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.duration {
  font-weight: 500;
  color: #3b82f6;
}

.input-data {
  max-width: 180px;
}

.error-message {
  max-width: 180px;
}

.action-buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pagination {
  display: flex;
  justify-content: center;
  padding: 20px;
}

.empty-state {
  background: white;
  border-radius: 12px;
  padding: 60px 20px;
  text-align: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
}

/* 响应式设计 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    gap: 16px;
  }
  
  .toolbar-left {
    width: 100%;
    flex-direction: column;
    gap: 12px;
  }
  
  .stats-row {
    grid-template-columns: 1fr;
  }
}

/* 表格样式增强 */
:deep(.el-table) {
  border-radius: 12px;
  overflow: hidden;
}

:deep(.el-table__header) {
  background: #f8fafc;
}

:deep(.el-table th) {
  background: #f8fafc !important;
  color: #374151;
  font-weight: 600;
  border-bottom: 1px solid #e5e7eb;
}

:deep(.el-table td) {
  border-bottom: 1px solid #f3f4f6;
}

:deep(.el-table__row:hover) {
  background-color: #f8fafc;
  cursor: pointer;
}

:deep(.el-table__row.current-row > td) {
  background-color: #e3f2fd !important;
}

:deep(.el-progress-bar__outer) {
  background-color: #e5e7eb;
}

:deep(.el-progress-bar__inner) {
  background: linear-gradient(90deg, #3b82f6, #06b6d4);
}
</style> 