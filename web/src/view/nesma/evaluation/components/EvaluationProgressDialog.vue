<template>
  <el-dialog
    v-model="dialogVisible"
    title="NESMA评估进度"
    width="1000px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @close="handleClose"
    class="progress-dialog"
    destroy-on-close
  >
    <div v-if="loading" class="loading-container">
      <el-skeleton :rows="3" animated />
    </div>
    
    <div v-else-if="progressData" class="progress-container">
      <!-- 顶部信息卡片 -->
      <div class="top-info-card">
        <div class="card-header">
          <div class="evaluation-title">
            <h3>{{ evaluationName }}</h3>
            <div class="status-badges">
              <el-tag :type="getStatusType(progressData.status)" size="large" effect="dark">
                {{ getStatusText(progressData.status) }}
              </el-tag>
              <el-tag type="info" size="small" effect="plain">
                ID: {{ progressData.evaluation_id }}
              </el-tag>
            </div>
          </div>
          <div class="progress-summary">
            <div class="progress-circle">
              <el-progress
                type="circle"
                :percentage="Math.round(progressData.progress)"
                :status="getProgressStatus(progressData.status)"
                :width="80"
                :stroke-width="6"
              >
                <template #default="{ percentage }">
                  <span class="percentage-text">{{ percentage }}%</span>
                </template>
              </el-progress>
            </div>
            <div class="current-phase">
              <span class="phase-label">当前阶段</span>
              <span class="phase-name">{{ progressData.current_phase || '准备中' }}</span>
              <span v-if="progressData.message" class="phase-message">
                {{ progressData.message }}
              </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 主要内容区域使用Tabs -->
      <el-tabs v-model="activeTab" class="progress-tabs" @tab-change="handleTabChange">
        <!-- 进度概览 -->
        <el-tab-pane label="进度概览" name="overview">
          <div class="overview-content">
            <!-- 阶段进度 -->
            <div class="phases-section">
              <h4 class="section-title">
                <el-icon><Guide /></el-icon>
                评估阶段
              </h4>
              <div class="phases-timeline">
                <div
                  v-for="(phase, index) in progressData.phases"
                  :key="index"
                  :class="['timeline-item', getPhaseClass(phase.status)]"
                >
                  <div class="timeline-dot">
                    <el-icon v-if="phase.status === 'completed'" class="success-icon">
                      <Check />
                    </el-icon>
                    <el-icon v-else-if="phase.status === 'processing'" class="loading-icon">
                      <Loading />
                    </el-icon>
                    <el-icon v-else-if="phase.status === 'failed'" class="error-icon">
                      <Close />
                    </el-icon>
                    <span v-else class="step-number">{{ index + 1 }}</span>
                  </div>
                  <div class="timeline-content">
                    <div class="phase-title">{{ phase.name }}</div>
                    <div v-if="phase.description" class="phase-desc">{{ phase.description }}</div>
                    <div v-if="phase.duration_seconds" class="phase-time">
                      <el-icon><Clock /></el-icon>
                      {{ formatDuration(phase.duration_seconds) }}
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 统计卡片 -->
            <div class="stats-section">
              <h4 class="section-title">
                <el-icon><DataAnalysis /></el-icon>
                统计信息
              </h4>
              <el-row :gutter="20">
                <el-col :span="6">
                  <div class="stat-card">
                    <div class="stat-icon steps">
                      <el-icon><List /></el-icon>
                    </div>
                    <div class="stat-content">
                      <div class="stat-value">{{ progressData.completed_steps }}/{{ progressData.total_steps }}</div>
                      <div class="stat-label">步骤进度</div>
                    </div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="stat-card">
                    <div class="stat-icon requirements">
                      <el-icon><Document /></el-icon>
                    </div>
                    <div class="stat-content">
                      <div class="stat-value">{{ progressData.processed_requirements }}/{{ progressData.total_requirements }}</div>
                      <div class="stat-label">需求处理</div>
                    </div>
                  </div>
                </el-col>
                <el-col :span="6">
                  <div class="stat-card">
                    <div class="stat-icon time">
                      <el-icon><Timer /></el-icon>
                    </div>
                    <div class="stat-content">
                      <div class="stat-value">{{ formatDuration(progressData.elapsed_time_seconds) }}</div>
                      <div class="stat-label">已用时间</div>
                    </div>
                  </div>
                </el-col>
                <el-col :span="6" v-if="progressData.estimated_completion_time">
                  <div class="stat-card">
                    <div class="stat-icon estimate">
                      <el-icon><Clock /></el-icon>
                    </div>
                    <div class="stat-content">
                      <div class="stat-value">{{ formatTime(progressData.estimated_completion_time, 'time') }}</div>
                      <div class="stat-label">预计完成</div>
                    </div>
                  </div>
                </el-col>
              </el-row>
            </div>
          </div>
        </el-tab-pane>

        <!-- 时间详情 -->
        <el-tab-pane label="时间详情" name="timing">
          <div class="timing-content">
            <el-row :gutter="20">
              <el-col :span="8">
                <div class="time-card">
                  <div class="time-icon start">
                    <el-icon><VideoPlay /></el-icon>
                  </div>
                  <div class="time-content">
                    <div class="time-label">开始时间</div>
                    <div class="time-value">{{ formatTime(progressData.start_time) }}</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="8" v-if="progressData.estimated_completion_time">
                <div class="time-card">
                  <div class="time-icon estimate">
                    <el-icon><Clock /></el-icon>
                  </div>
                  <div class="time-content">
                    <div class="time-label">预计完成</div>
                    <div class="time-value">{{ formatTime(progressData.estimated_completion_time) }}</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="8" v-if="progressData.status === 'completed'">
                <div class="time-card">
                  <div class="time-icon complete">
                    <el-icon><CircleCheck /></el-icon>
                  </div>
                  <div class="time-content">
                    <div class="time-label">完成时间</div>
                    <div class="time-value">{{ formatTime(progressData.completion_time) }}</div>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>
        </el-tab-pane>

        <!-- 评估日志 -->
        <el-tab-pane label="评估日志" name="logs">
          <div class="logs-content">
            <div class="logs-header">
              <div class="logs-title">
                <el-icon><Document /></el-icon>
                <span>评估日志</span>
                <el-badge :value="logs.length" class="logs-count" />
              </div>
              <el-button size="small" @click="refreshLogs" :loading="logsLoading">
                <el-icon><Refresh /></el-icon>
                刷新
              </el-button>
            </div>
            <div class="logs-list-container">
              <div v-if="logs.length === 0" class="no-logs">
                <el-empty description="暂无日志信息" />
              </div>
              <div v-else class="logs-list">
                <div
                  v-for="(log, index) in logs"
                  :key="index"
                  :class="['log-item', `log-${log.level}`]"
                >
                  <div class="log-indicator">
                    <div :class="['log-dot', `dot-${log.level}`]"></div>
                  </div>
                  <div class="log-content">
                    <div class="log-header">
                      <span class="log-phase">{{ log.phase }}</span>
                      <span class="log-time">{{ formatTime(log.timestamp, 'datetime') }}</span>
                    </div>
                    <div class="log-message">{{ log.message }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <div v-else class="error-container">
      <el-result
        icon="error"
        title="加载失败"
        :sub-title="errorMessage || '无法获取评估进度信息'"
      >
        <template #extra>
          <el-button type="primary" @click="loadProgressData">重新加载</el-button>
        </template>
      </el-result>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <div class="footer-info">
          <span v-if="lastUpdateTime" class="last-update">
            最后更新: {{ formatTime(lastUpdateTime, 'time') }}
          </span>
          <span v-if="progressData?.status === 'processing'" class="auto-refresh">
            <el-icon class="refresh-icon" :class="{ rotating: isRefreshing }"><Refresh /></el-icon>
            自动刷新
          </span>
        </div>
        <div class="footer-actions">
          <el-button @click="handleClose">关闭</el-button>
          <el-button type="primary" @click="refreshProgress" :loading="refreshing">
            <el-icon><Refresh /></el-icon>
            刷新
          </el-button>
          <el-button
            v-if="progressData?.status === 'processing'"
            type="danger"
            @click="handleCancel"
            :loading="cancelling"
          >
            <el-icon><Close /></el-icon>
            取消评估
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Check, Loading, Close, Guide, DataAnalysis, List, Document, 
  Timer, Clock, VideoPlay, CircleCheck, Refresh
} from '@element-plus/icons-vue'
import {
  getEvaluationProgress,
  getEvaluationLogs,
  cancelEvaluation
} from '@/api/nesmaEvaluation'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  evaluationId: {
    type: Number,
    required: true
  },
  evaluationName: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'close'])

// Reactive data
const dialogVisible = ref(false)
const loading = ref(false)
const refreshing = ref(false)
const cancelling = ref(false)
const logsLoading = ref(false)
const progressData = ref(null)
const logs = ref([])
const errorMessage = ref('')
const refreshTimer = ref(null)
const lastUpdateTime = ref(null)
const isRefreshing = ref(false)
const activeTab = ref('overview')

// Watch for dialog visibility
watch(() => props.modelValue, (newVal) => {
  dialogVisible.value = newVal
  if (newVal) {
    loadProgressData()
    startAutoRefresh()
  } else {
    stopAutoRefresh()
  }
})

watch(dialogVisible, (newVal) => {
  emit('update:modelValue', newVal)
})

// Methods
const loadProgressData = async (silent = false) => {
  if (!props.evaluationId) return
  
  if (!silent) {
    loading.value = true
    errorMessage.value = ''
  } else {
    isRefreshing.value = true
  }
  
  try {
    const response = await getEvaluationProgress(props.evaluationId)
    progressData.value = response.data
    lastUpdateTime.value = new Date()
    
    // 只在第一次加载或用户主动刷新时加载日志
    if (!silent || activeTab.value === 'logs') {
      await loadLogs(silent)
    }
  } catch (error) {
    console.error('加载评估进度失败:', error)
    if (!silent) {
      errorMessage.value = error.response?.data?.msg || error.message || '加载失败'
    }
  } finally {
    if (!silent) {
      loading.value = false
    } else {
      isRefreshing.value = false
    }
  }
}

const loadLogs = async (silent = false) => {
  if (!props.evaluationId) return
  
  if (!silent) {
    logsLoading.value = true
  }
  
  try {
    const response = await getEvaluationLogs(props.evaluationId, 1, 50)
    logs.value = response.data.logs || []
  } catch (error) {
    console.error('加载评估日志失败:', error)
  } finally {
    if (!silent) {
      logsLoading.value = false
    }
  }
}

const refreshProgress = async () => {
  refreshing.value = true
  try {
    await loadProgressData()
    ElMessage.success('进度刷新成功')
  } finally {
    refreshing.value = false
  }
}

const refreshLogs = async () => {
  await loadLogs()
}

const handleTabChange = (tabName) => {
  if (tabName === 'logs' && logs.value.length === 0) {
    loadLogs()
  }
}

const handleCancel = async () => {
  try {
    await ElMessageBox.confirm('确定要取消此评估吗？取消后评估将停止执行。', '确认取消评估', {
      type: 'warning',
      confirmButtonText: '确定取消',
      cancelButtonText: '继续评估'
    })
    
    cancelling.value = true
    try {
      await cancelEvaluation(props.evaluationId, '用户在进度页面取消')
      ElMessage.success('评估已取消')
      await loadProgressData() // 刷新进度状态
    } catch (error) {
      console.error('取消评估失败:', error)
      ElMessage.error('取消评估失败: ' + (error.response?.data?.msg || error.message))
    } finally {
      cancelling.value = false
    }
  } catch (error) {
    // 用户取消操作
  }
}

const handleClose = () => {
  stopAutoRefresh()
  emit('close')
}

// Auto refresh for processing evaluations - 优化刷新逻辑，避免闪烁
const startAutoRefresh = () => {
  stopAutoRefresh()
  refreshTimer.value = setInterval(async () => {
    if (progressData.value?.status === 'processing') {
      // 使用silent模式刷新，避免UI闪烁
      await loadProgressData(true)
    } else {
      stopAutoRefresh()
    }
  }, 3000) // 改为3秒刷新一次，频率更高但不会闪烁
}

const stopAutoRefresh = () => {
  if (refreshTimer.value) {
    clearInterval(refreshTimer.value)
    refreshTimer.value = null
  }
  isRefreshing.value = false
}

// Utility methods
const getStatusType = (status) => {
  const statusMap = {
    'pending': 'warning',
    'processing': 'primary',
    'completed': 'success',
    'failed': 'danger',
    'cancelled': 'info'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    'pending': '待开始',
    'processing': '进行中',
    'completed': '已完成',
    'failed': '失败',
    'cancelled': '已取消'
  }
  return statusMap[status] || '未知'
}

const getProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const getPhaseClass = (status) => {
  return {
    'phase-completed': status === 'completed',
    'phase-processing': status === 'processing',
    'phase-failed': status === 'failed',
    'phase-pending': status === 'pending'
  }
}

const formatTime = (timeString, format = 'datetime') => {
  if (!timeString) return '-'
  const date = new Date(timeString)
  
  switch (format) {
    case 'time':
      return date.toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit',
        second: '2-digit'
      })
    case 'date':
      return date.toLocaleDateString('zh-CN')
    case 'datetime':
    default:
      return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
      })
  }
}

const formatDuration = (seconds) => {
  if (!seconds || seconds < 0) return '0秒'
  
  const hours = Math.floor(seconds / 3600)
  const minutes = Math.floor((seconds % 3600) / 60)
  const secs = Math.floor(seconds % 60)
  
  if (hours > 0) {
    return `${hours}小时${minutes}分${secs}秒`
  } else if (minutes > 0) {
    return `${minutes}分${secs}秒`
  } else {
    return `${secs}秒`
  }
}

// Lifecycle
onUnmounted(() => {
  stopAutoRefresh()
})
</script>

<style lang="scss" scoped>
// 对话框整体样式
:deep(.progress-dialog) {
  .el-dialog__body {
    padding: 20px 24px;
    max-height: 70vh;
    overflow: hidden;
  }
  
  .el-dialog__footer {
    padding: 16px 24px;
    border-top: 1px solid #e4e7ed;
  }
}

.progress-container {
  min-height: 500px;
}

// 顶部信息卡片
.top-info-card {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 12px;
  padding: 24px;
  margin-bottom: 24px;
  color: white;
  
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
  
  .evaluation-title {
    h3 {
      margin: 0 0 12px 0;
      font-size: 20px;
      font-weight: 600;
      color: white;
    }
    
    .status-badges {
      display: flex;
      gap: 12px;
      align-items: center;
    }
  }
  
  .progress-summary {
    display: flex;
    align-items: center;
    gap: 24px;
    
    .progress-circle {
      :deep(.el-progress-circle__text) {
        color: white !important;
      }
    }
    
    .percentage-text {
      font-size: 16px;
      font-weight: 600;
      color: white;
    }
    
    .current-phase {
      display: flex;
      flex-direction: column;
      gap: 4px;
      
      .phase-label {
        font-size: 12px;
        opacity: 0.8;
      }
      
      .phase-name {
        font-size: 16px;
        font-weight: 600;
      }
      
      .phase-message {
        font-size: 12px;
        opacity: 0.8;
      }
    }
  }
}

// Tabs样式
.progress-tabs {
  :deep(.el-tabs__header) {
    margin: 0 0 20px 0;
    border-bottom: 2px solid #f0f2f5;
  }
  
  :deep(.el-tabs__item) {
    padding: 0 20px;
    font-weight: 500;
    
    &.is-active {
      color: #409eff;
    }
  }
  
  :deep(.el-tabs__content) {
    overflow: visible;
  }
  
  :deep(.el-tab-pane) {
    max-height: 400px;
    overflow-y: auto;
    padding-right: 8px;
    
    &::-webkit-scrollbar {
      width: 6px;
    }
    
    &::-webkit-scrollbar-track {
      background: #f1f1f1;
      border-radius: 3px;
    }
    
    &::-webkit-scrollbar-thumb {
      background: #c1c1c1;
      border-radius: 3px;
      
      &:hover {
        background: #a8a8a8;
      }
    }
  }
}

// 概览内容
.overview-content {
  .section-title {
    display: flex;
    align-items: center;
    gap: 8px;
    margin: 0 0 16px 0;
    font-size: 16px;
    font-weight: 600;
    color: #303133;
  }
}

// 阶段时间线
.phases-section {
  margin-bottom: 32px;
  
  .phases-timeline {
    .timeline-item {
      display: flex;
      align-items: flex-start;
      padding: 16px 0;
      position: relative;
      
      &:not(:last-child):after {
        content: '';
        position: absolute;
        left: 20px;
        top: 60px;
        width: 2px;
        height: calc(100% - 20px);
        background: #e4e7ed;
        z-index: 1;
      }
      
      &.timeline-completed:after {
        background: #67c23a;
      }
      
      &.timeline-processing:after {
        background: #409eff;
      }
      
      .timeline-dot {
        width: 40px;
        height: 40px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 16px;
        flex-shrink: 0;
        position: relative;
        z-index: 2;
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
        
        .success-icon {
          background: #67c23a;
          color: white;
          border-radius: 50%;
          padding: 8px;
        }
        
        .loading-icon {
          background: #409eff;
          color: white;
          border-radius: 50%;
          padding: 8px;
          animation: spin 1s linear infinite;
        }
        
        .error-icon {
          background: #f56c6c;
          color: white;
          border-radius: 50%;
          padding: 8px;
        }
        
        .step-number {
          background: #e4e7ed;
          color: #909399;
          border-radius: 50%;
          width: 100%;
          height: 100%;
          display: flex;
          align-items: center;
          justify-content: center;
          font-weight: 600;
        }
      }
      
      &.timeline-completed .timeline-dot {
        background: #67c23a;
        color: white;
      }
      
      &.timeline-processing .timeline-dot {
        background: #409eff;
        color: white;
      }
      
      &.timeline-failed .timeline-dot {
        background: #f56c6c;
        color: white;
      }
      
      .timeline-content {
        flex: 1;
        
        .phase-title {
          font-size: 16px;
          font-weight: 600;
          color: #303133;
          margin-bottom: 8px;
        }
        
        .phase-desc {
          font-size: 14px;
          color: #606266;
          margin-bottom: 8px;
        }
        
        .phase-time {
          display: flex;
          align-items: center;
          gap: 4px;
          font-size: 12px;
          color: #909399;
        }
      }
    }
  }
}

// 统计卡片
.stats-section {
  .stat-card {
    display: flex;
    align-items: center;
    padding: 20px;
    background: #f8fafc;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    transition: all 0.3s ease;
    
    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }
    
    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 16px;
      color: white;
      
      &.steps {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }
      
      &.requirements {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      }
      
      &.time {
        background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
      }
      
      &.estimate {
        background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
      }
    }
    
    .stat-content {
      .stat-value {
        font-size: 18px;
        font-weight: 700;
        color: #1a202c;
        margin-bottom: 4px;
      }
      
      .stat-label {
        font-size: 12px;
        color: #718096;
        font-weight: 500;
      }
    }
  }
}

// 时间卡片
.timing-content {
  .time-card {
    display: flex;
    align-items: center;
    padding: 24px;
    background: white;
    border-radius: 12px;
    border: 1px solid #e2e8f0;
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
    transition: all 0.3s ease;
    
    &:hover {
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
    }
    
    .time-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 16px;
      color: white;
      
      &.start {
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
      }
      
      &.estimate {
        background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
      }
      
      &.complete {
        background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
      }
    }
    
    .time-content {
      .time-label {
        font-size: 14px;
        color: #718096;
        margin-bottom: 4px;
      }
      
      .time-value {
        font-size: 16px;
        font-weight: 600;
        color: #1a202c;
      }
    }
  }
}

// 日志内容
.logs-content {
  .logs-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    
    .logs-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      
      .logs-count {
        :deep(.el-badge__content) {
          font-size: 10px;
        }
      }
    }
  }
  
  .logs-list-container {
    .logs-list {
      .log-item {
        display: flex;
        align-items: flex-start;
        padding: 12px 0;
        border-bottom: 1px solid #f0f2f5;
        
        &:last-child {
          border-bottom: none;
        }
        
        .log-indicator {
          margin-right: 12px;
          padding-top: 4px;
          
          .log-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            
            &.dot-info {
              background: #409eff;
            }
            
            &.dot-warning {
              background: #e6a23c;
            }
            
            &.dot-error {
              background: #f56c6c;
            }
          }
        }
        
        .log-content {
          flex: 1;
          
          .log-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 4px;
            
            .log-phase {
              font-weight: 600;
              color: #303133;
              font-size: 14px;
            }
            
            .log-time {
              font-size: 12px;
              color: #909399;
            }
          }
          
          .log-message {
            font-size: 13px;
            color: #606266;
            line-height: 1.4;
          }
        }
      }
    }
  }
}

// 底部样式
.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  
  .footer-info {
    display: flex;
    align-items: center;
    gap: 16px;
    
    .last-update {
      font-size: 12px;
      color: #909399;
    }
    
    .auto-refresh {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;
      color: #409eff;
      
      .refresh-icon {
        &.rotating {
          animation: spin 1s linear infinite;
        }
      }
    }
  }
  
  .footer-actions {
    display: flex;
    gap: 12px;
  }
}

// 加载和错误状态
.loading-container {
  padding: 40px 20px;
}

.error-container {
  padding: 40px 20px;
}

// 动画
@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 响应式设计
@media (max-width: 768px) {
  .top-info-card .card-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 16px;
  }
  
  .progress-summary {
    flex-direction: column;
    align-items: flex-start !important;
    gap: 16px !important;
  }
  
  .stat-card, .time-card {
    flex-direction: column;
    text-align: center;
    
    .stat-icon, .time-icon {
      margin-right: 0;
      margin-bottom: 12px;
    }
  }
}
</style>