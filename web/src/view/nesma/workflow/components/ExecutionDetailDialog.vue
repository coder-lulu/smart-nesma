<template>
  <el-dialog
    v-model="dialogVisible"
    title="执行详情"
    width="1000px"
    :before-close="handleClose"
    class="execution-detail-dialog"
  >
    <div class="detail-content" v-loading="loading">
      <div v-if="executionDetail" class="execution-detail">
        <!-- 执行基本信息 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <el-icon><Operation /></el-icon>
              <span>执行信息</span>
              <div class="header-actions">
                <el-tag :type="getStatusTagType(executionDetail.status)" size="large">
                  {{ getStatusLabel(executionDetail.status) }}
                </el-tag>
              </div>
            </div>
          </template>
          
          <el-descriptions :column="2" border>
            <el-descriptions-item label="执行ID">
              <el-text class="execution-id">{{ executionDetail.executionId }}</el-text>
              <el-button 
                size="small" 
                text 
                @click="copyToClipboard(executionDetail.executionId)"
                style="margin-left: 8px"
              >
                复制
              </el-button>
            </el-descriptions-item>
            <el-descriptions-item label="工作流ID">{{ executionDetail.workflowId }}</el-descriptions-item>
            <el-descriptions-item label="当前步骤">{{ executionDetail.currentStep || '-' }}</el-descriptions-item>
            <el-descriptions-item label="会话ID">{{ executionDetail.sessionId || '-' }}</el-descriptions-item>
            <el-descriptions-item label="开始时间">{{ formatTime(executionDetail.startTime) }}</el-descriptions-item>
            <el-descriptions-item label="结束时间">{{ formatTime(executionDetail.endTime) }}</el-descriptions-item>
            <el-descriptions-item label="总耗时" v-if="executionDetail.duration">
              {{ formatDuration(executionDetail.duration) }}
            </el-descriptions-item>
            <el-descriptions-item label="错误信息" v-if="executionDetail.errorMessage" :span="2">
              <el-text type="danger">{{ executionDetail.errorMessage }}</el-text>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 执行进度 -->
        <el-card class="progress-card">
          <template #header>
            <div class="card-header">
              <el-icon><TrendCharts /></el-icon>
              <span>执行进度</span>
            </div>
          </template>
          
          <div class="progress-content">
            <div class="progress-bar">
              <el-progress 
                :percentage="getOverallProgress()" 
                :status="getProgressStatus()"
                :stroke-width="12"
                :show-text="true"
              />
            </div>
            
            <!-- 步骤进度 -->
            <div class="steps-progress">
              <div 
                v-for="(step, index) in workflowSteps" 
                :key="step.id"
                class="step-progress"
                :class="getStepClass(step, index)"
              >
                <div class="step-number">{{ index + 1 }}</div>
                <div class="step-info">
                  <div class="step-name">{{ step.name }}</div>
                  <div class="step-status">{{ getStepStatus(step, index) }}</div>
                </div>
                <div class="step-duration" v-if="step.duration">
                  {{ formatDuration(step.duration) }}
                </div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 输入输出数据 -->
        <el-card class="data-card">
          <template #header>
            <div class="card-header">
              <el-icon><Files /></el-icon>
              <span>输入输出数据</span>
            </div>
          </template>
          
          <el-tabs v-model="activeDataTab">
            <el-tab-pane label="输入数据" name="input">
              <div class="data-content">
                <div class="data-actions">
                  <el-button size="small" @click="formatInputData">格式化</el-button>
                  <el-button size="small" @click="copyInputData">复制</el-button>
                </div>
                <el-input
                  v-model="inputDataFormatted"
                  type="textarea"
                  :rows="12"
                  readonly
                  class="data-textarea"
                />
              </div>
            </el-tab-pane>
            <el-tab-pane label="输出数据" name="output">
              <div class="data-content">
                <div class="data-actions">
                  <el-button size="small" @click="formatOutputData">格式化</el-button>
                  <el-button size="small" @click="copyOutputData">复制</el-button>
                  <el-button size="small" @click="downloadOutputData">下载</el-button>
                </div>
                <el-input
                  v-model="outputDataFormatted"
                  type="textarea"
                  :rows="12"
                  readonly
                  class="data-textarea"
                />
              </div>
            </el-tab-pane>
          </el-tabs>
        </el-card>

        <!-- 执行日志 -->
        <el-card class="logs-card">
          <template #header>
            <div class="card-header">
              <el-icon><Document /></el-icon>
              <span>执行日志</span>
              <div class="header-actions">
                <el-button size="small" @click="refreshLogs">刷新</el-button>
                <el-button size="small" @click="downloadLogs">下载日志</el-button>
                <el-switch 
                  v-model="autoRefreshLogs" 
                  active-text="自动刷新"
                  size="small"
                />
              </div>
            </div>
          </template>
          
          <div class="logs-content">
            <div class="logs-filters">
              <el-select 
                v-model="logLevelFilter" 
                placeholder="日志级别" 
                size="small"
                style="width: 120px"
                @change="filterLogs"
              >
                <el-option label="全部" value="" />
                <el-option label="INFO" value="info" />
                <el-option label="WARN" value="warn" />
                <el-option label="ERROR" value="error" />
              </el-select>
              <el-input
                v-model="logSearchKeyword"
                placeholder="搜索日志内容"
                size="small"
                style="width: 200px; margin-left: 12px"
                @input="filterLogs"
              />
            </div>
            
            <div class="logs-list" ref="logsContainer">
              <div 
                v-for="(log, index) in filteredLogs" 
                :key="index"
                class="log-item"
                :class="log.level"
              >
                <span class="log-time">{{ formatTime(log.timestamp) }}</span>
                <span class="log-level">{{ log.level.toUpperCase() }}</span>
                <span class="log-step" v-if="log.step">{{ log.step }}</span>
                <span class="log-message">{{ log.message }}</span>
              </div>
            </div>
            
            <div v-if="filteredLogs.length === 0" class="no-logs">
              <el-empty description="暂无日志记录" />
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button 
          v-if="executionDetail?.status === 'running'" 
          type="danger" 
          @click="cancelExecution"
        >
          取消执行
        </el-button>
        <el-button 
          v-if="executionDetail?.status === 'failed'" 
          type="primary" 
          @click="retryExecution"
        >
          重新执行
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Operation, TrendCharts, Files, Document } from '@element-plus/icons-vue'

import { 
  getWorkflowExecution,
  cancelWorkflowExecution as cancelWorkflowExecutionAPI
} from '@/api/nesma'

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  executionId: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['update:visible', 'refresh'])

// 响应式数据
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const loading = ref(false)
const executionDetail = ref(null)
const workflowSteps = ref([])
const activeDataTab = ref('input')
const inputDataFormatted = ref('')
const outputDataFormatted = ref('')
const autoRefreshLogs = ref(false)
const logLevelFilter = ref('')
const logSearchKeyword = ref('')
const logs = ref([])
const logsContainer = ref(null)

// 模拟日志数据
const mockLogs = ref([
  {
    timestamp: new Date('2024-01-15T10:00:00'),
    level: 'info',
    step: '开始',
    message: '工作流开始执行'
  },
  {
    timestamp: new Date('2024-01-15T10:00:05'),
    level: 'info',
    step: '需求分析',
    message: '开始分析用户需求'
  },
  {
    timestamp: new Date('2024-01-15T10:00:15'),
    level: 'info',
    step: '需求分析',
    message: '需求分析完成，识别到5个功能模块'
  },
  {
    timestamp: new Date('2024-01-15T10:00:20'),
    level: 'info',
    step: 'NESMA评估',
    message: '开始NESMA功能点评估'
  },
  {
    timestamp: new Date('2024-01-15T10:00:45'),
    level: 'warn',
    step: 'NESMA评估',
    message: '检测到复杂度较高的功能点，需要人工确认'
  }
])

let refreshTimer = null

// 计算属性
const filteredLogs = computed(() => {
  let result = logs.value

  if (logLevelFilter.value) {
    result = result.filter(log => log.level === logLevelFilter.value)
  }

  if (logSearchKeyword.value) {
    const keyword = logSearchKeyword.value.toLowerCase()
    result = result.filter(log => 
      log.message.toLowerCase().includes(keyword) ||
      log.step?.toLowerCase().includes(keyword)
    )
  }

  return result
})

// 获取执行详情
const getExecutionDetail = async () => {
  if (!props.executionId) return
  
  loading.value = true
  try {
    const res = await getWorkflowExecution(props.executionId)
    if (res.code === 0) {
      executionDetail.value = res.data
      
      // 格式化输入输出数据
      formatDataForDisplay()
      
      // 模拟工作流步骤
      workflowSteps.value = [
        { id: 'start', name: '开始', status: 'completed' },
        { id: 'analyze', name: '需求分析', status: 'completed' },
        { id: 'nesma', name: 'NESMA评估', status: 'running' },
        { id: 'document', name: '文档生成', status: 'pending' },
        { id: 'end', name: '结束', status: 'pending' }
      ]
      
      // 使用模拟日志数据
      logs.value = mockLogs.value
    }
  } catch (error) {
    ElMessage.error('获取执行详情失败')
  } finally {
    loading.value = false
  }
}

// 格式化数据显示
const formatDataForDisplay = () => {
  try {
    if (executionDetail.value?.inputData) {
      const inputData = typeof executionDetail.value.inputData === 'string' 
        ? JSON.parse(executionDetail.value.inputData)
        : executionDetail.value.inputData
      inputDataFormatted.value = JSON.stringify(inputData, null, 2)
    }
    
    if (executionDetail.value?.outputData) {
      const outputData = typeof executionDetail.value.outputData === 'string'
        ? JSON.parse(executionDetail.value.outputData)
        : executionDetail.value.outputData
      outputDataFormatted.value = JSON.stringify(outputData, null, 2)
    }
  } catch (error) {
    console.error('格式化数据失败:', error)
  }
}

// 获取整体进度
const getOverallProgress = () => {
  if (!workflowSteps.value.length) return 0
  
  const completedSteps = workflowSteps.value.filter(step => step.status === 'completed').length
  return Math.round((completedSteps / workflowSteps.value.length) * 100)
}

// 获取进度状态
const getProgressStatus = () => {
  if (!executionDetail.value) return ''
  
  switch (executionDetail.value.status) {
    case 'completed': return 'success'
    case 'failed': return 'exception'
    default: return ''
  }
}

// 获取步骤类别
const getStepClass = (step, index) => {
  return {
    'step-completed': step.status === 'completed',
    'step-running': step.status === 'running',
    'step-pending': step.status === 'pending',
    'step-failed': step.status === 'failed'
  }
}

// 获取步骤状态
const getStepStatus = (step, index) => {
  const statusMap = {
    completed: '已完成',
    running: '运行中',
    pending: '等待中',
    failed: '已失败'
  }
  return statusMap[step.status] || '未知'
}

// 刷新日志
const refreshLogs = () => {
  // 这里可以添加实际的日志刷新逻辑
  ElMessage.success('日志已刷新')
}

// 过滤日志
const filterLogs = () => {
  // 计算属性会自动重新计算
}

// 下载日志
const downloadLogs = () => {
  const logContent = filteredLogs.value
    .map(log => `[${formatTime(log.timestamp)}] [${log.level.toUpperCase()}] ${log.step ? `[${log.step}] ` : ''}${log.message}`)
    .join('\n')
  
  const blob = new Blob([logContent], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `execution_${props.executionId}_logs.txt`
  a.click()
  URL.revokeObjectURL(url)
}

// 取消执行
const cancelExecution = async () => {
  try {
    await cancelWorkflowExecutionAPI(props.executionId)
    ElMessage.success('取消成功')
    emit('refresh')
    getExecutionDetail()
  } catch (error) {
    ElMessage.error('取消失败')
  }
}

// 重新执行
const retryExecution = () => {
  ElMessage.info('重新执行功能开发中')
}

// 数据操作
const formatInputData = () => {
  try {
    const data = JSON.parse(inputDataFormatted.value)
    inputDataFormatted.value = JSON.stringify(data, null, 2)
  } catch (error) {
    ElMessage.error('数据格式错误')
  }
}

const formatOutputData = () => {
  try {
    const data = JSON.parse(outputDataFormatted.value)
    outputDataFormatted.value = JSON.stringify(data, null, 2)
  } catch (error) {
    ElMessage.error('数据格式错误')
  }
}

const copyInputData = async () => {
  try {
    await navigator.clipboard.writeText(inputDataFormatted.value)
    ElMessage.success('输入数据已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const copyOutputData = async () => {
  try {
    await navigator.clipboard.writeText(outputDataFormatted.value)
    ElMessage.success('输出数据已复制')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

const downloadOutputData = () => {
  const blob = new Blob([outputDataFormatted.value], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `execution_${props.executionId}_output.json`
  a.click()
  URL.revokeObjectURL(url)
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

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  
  // 清理定时器
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

// 监听自动刷新
watch(() => autoRefreshLogs.value, (enabled) => {
  if (enabled && executionDetail.value?.status === 'running') {
    refreshTimer = setInterval(() => {
      refreshLogs()
    }, 5000)
  } else if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
})

// 监听props变化
watch(() => props.executionId, (newId) => {
  if (newId && props.visible) {
    getExecutionDetail()
  }
})

watch(() => props.visible, (visible) => {
  if (visible && props.executionId) {
    getExecutionDetail()
  } else {
    handleClose()
  }
})

// 生命周期
onMounted(() => {
  if (props.visible && props.executionId) {
    getExecutionDetail()
  }
})

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer)
  }
})
</script>

<style scoped>
.execution-detail-dialog {
  --el-dialog-border-radius: 16px;
}

.detail-content {
  max-height: 80vh;
  overflow-y: auto;
}

.execution-detail {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.info-card, .progress-card, .data-card, .logs-card {
  border: none;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border-radius: 12px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  font-weight: 600;
  color: #374151;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.execution-id {
  font-family: 'Monaco', 'Consolas', monospace;
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.progress-content {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.progress-bar {
  margin-bottom: 16px;
}

.steps-progress {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.step-progress {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e5e7eb;
  background: white;
}

.step-progress.step-completed {
  background: #f0f9ff;
  border-color: #3b82f6;
}

.step-progress.step-running {
  background: #fef3c7;
  border-color: #f59e0b;
}

.step-progress.step-failed {
  background: #fef2f2;
  border-color: #ef4444;
}

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  color: white;
  flex-shrink: 0;
}

.step-completed .step-number {
  background: #3b82f6;
}

.step-running .step-number {
  background: #f59e0b;
}

.step-failed .step-number {
  background: #ef4444;
}

.step-pending .step-number {
  background: #9ca3af;
}

.step-info {
  flex: 1;
}

.step-name {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.step-status {
  font-size: 13px;
  color: #6b7280;
}

.step-duration {
  font-size: 13px;
  color: #3b82f6;
  font-weight: 500;
}

.data-content {
  position: relative;
}

.data-actions {
  margin-bottom: 12px;
  display: flex;
  gap: 8px;
}

.data-textarea {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 13px;
}

.logs-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.logs-filters {
  display: flex;
  align-items: center;
  gap: 12px;
}

.logs-list {
  max-height: 400px;
  overflow-y: auto;
  background: #f8fafc;
  border-radius: 8px;
  padding: 16px;
}

.log-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 8px 0;
  font-size: 13px;
  font-family: 'Monaco', 'Consolas', monospace;
  border-bottom: 1px solid #e5e7eb;
}

.log-item:last-child {
  border-bottom: none;
}

.log-time {
  color: #6b7280;
  min-width: 160px;
  flex-shrink: 0;
}

.log-level {
  min-width: 50px;
  font-weight: 600;
  flex-shrink: 0;
}

.log-level.info { color: #3b82f6; }
.log-level.warn { color: #f59e0b; }
.log-level.error { color: #ef4444; }

.log-step {
  min-width: 80px;
  color: #8b5cf6;
  font-weight: 500;
  flex-shrink: 0;
}

.log-message {
  color: #374151;
  line-height: 1.5;
  flex: 1;
}

.no-logs {
  padding: 40px 0;
  text-align: center;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 0 0 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .step-progress {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
  
  .logs-filters {
    flex-direction: column;
    align-items: stretch;
    gap: 8px;
  }
  
  .log-item {
    flex-direction: column;
    gap: 4px;
  }
  
  .log-time, .log-level, .log-step {
    min-width: auto;
  }
}

/* 表格和卡片样式增强 */
:deep(.el-descriptions__label) {
  font-weight: 500;
  color: #374151;
}

:deep(.el-card__header) {
  padding: 20px 24px;
  background: #f8fafc;
}

:deep(.el-card__body) {
  padding: 24px;
}

:deep(.el-progress-bar__outer) {
  background-color: #e5e7eb;
  border-radius: 6px;
}

:deep(.el-progress-bar__inner) {
  border-radius: 6px;
}

:deep(.el-tabs__nav-wrap) {
  margin-bottom: 16px;
}
</style> 