<template>
  <el-dialog
    v-model="dialogVisible"
    title="Agent详情"
    width="900px"
    :before-close="handleClose"
  >
    <div v-if="agentDetail" class="agent-detail">
      <!-- 基本信息 -->
      <el-card class="info-card">
        <template #header>
          <div class="card-header">
            <span>基本信息</span>
            <el-tag :type="getStatusTagType(agentDetail.status)" size="small">
              {{ getStatusLabel(agentDetail.status) }}
            </el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="Agent ID" :span="2">
            <el-text copyable>{{ agentDetail.agentId }}</el-text>
          </el-descriptions-item>
          <el-descriptions-item label="Agent名称">
            {{ agentDetail.name }}
          </el-descriptions-item>
          <el-descriptions-item label="Agent类型">
            <el-tag :type="getAgentTypeColor(agentDetail.agentType)">
              {{ getAgentTypeLabel(agentDetail.agentType) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="当前状态">
            <div class="status-item">
              <el-tag :type="getStatusTagType(agentDetail.status)">
                {{ getStatusLabel(agentDetail.status) }}
              </el-tag>
              <el-button 
                size="small" 
                type="primary" 
                @click="handleToggleStatus"
                :loading="statusLoading"
              >
                {{ agentDetail.status === 'active' ? '停止' : '启动' }}
              </el-button>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="版本号">
            {{ agentDetail.version }}
          </el-descriptions-item>
          <el-descriptions-item label="最大并发数">
            {{ agentDetail.maxConcurrency }}
          </el-descriptions-item>
          <el-descriptions-item label="当前处理任务">
            <el-progress
              :percentage="getLoadPercentage(agentDetail)"
              :color="getLoadColor(agentDetail)"
              :stroke-width="6"
            >
              <span>{{ agentDetail.processingCount || 0 }}/{{ agentDetail.maxConcurrency || 0 }}</span>
            </el-progress>
          </el-descriptions-item>
          <el-descriptions-item label="注册时间">
            {{ formatDateTime(agentDetail.createdAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="最后心跳">
            <div class="heartbeat-item">
              <el-icon :style="{ color: getHeartbeatColor(agentDetail.lastHeartbeat) }">
                <Notification />
              </el-icon>
              {{ formatDateTime(agentDetail.lastHeartbeat) }}
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="端点地址" :span="2">
            <el-text copyable>{{ agentDetail.endpoint || '-' }}</el-text>
          </el-descriptions-item>
          <el-descriptions-item label="Agent描述" :span="2">
            {{ agentDetail.description }}
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 性能指标 -->
      <el-card class="metrics-card">
        <template #header>
          <div class="card-header">
            <span>性能指标</span>
            <el-button 
              type="text" 
              icon="Refresh" 
              @click="refreshMetrics"
              :loading="metricsLoading"
            />
          </div>
        </template>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-icon">
                <el-icon><DataBoard /></el-icon>
              </div>
              <div class="metric-info">
                <div class="metric-value">{{ agentDetail.totalProcessed || 0 }}</div>
                <div class="metric-label">总处理任务</div>
              </div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-icon">
                <el-icon><Timer /></el-icon>
              </div>
              <div class="metric-info">
                <div class="metric-value">{{ agentDetail.averageResponseTime || 0 }}ms</div>
                <div class="metric-label">平均响应时间</div>
              </div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-icon">
                <el-icon><SuccessFilled /></el-icon>
              </div>
              <div class="metric-info">
                <div class="metric-value">{{ getSuccessRate(agentDetail) }}%</div>
                <div class="metric-label">成功率</div>
              </div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-icon">
                <el-icon><TrendCharts /></el-icon>
              </div>
              <div class="metric-info">
                <div class="metric-value">{{ agentDetail.efficiency || 0 }}%</div>
                <div class="metric-label">运行效率</div>
              </div>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 任务管理 -->
      <el-card class="tasks-card">
        <template #header>
          <div class="card-header">
            <span>任务管理</span>
            <div class="header-actions">
              <el-select
                v-model="taskFilter"
                placeholder="任务状态"
                size="small"
                style="width: 120px; margin-right: 10px;"
                @change="getTaskList"
              >
                <el-option label="全部" value="" />
                <el-option label="待处理" value="pending" />
                <el-option label="执行中" value="processing" />
                <el-option label="已完成" value="completed" />
                <el-option label="失败" value="failed" />
              </el-select>
              <el-button 
                type="primary" 
                size="small" 
                icon="Plus" 
                @click="handleCreateTask"
              >
                创建任务
              </el-button>
            </div>
          </div>
        </template>
        <el-table :data="taskList" v-loading="taskLoading" max-height="300">
          <el-table-column prop="taskId" label="任务ID" width="180" show-overflow-tooltip />
          <el-table-column prop="taskType" label="任务类型" width="120" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="scope">
              <el-tag :type="getTaskStatusType(scope.row.status)" size="small">
                {{ getTaskStatusLabel(scope.row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="priority" label="优先级" width="100">
            <template #default="scope">
              <el-rate v-model="scope.row.priority" disabled :max="5" size="small" />
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="160">
            <template #default="scope">
              {{ formatDateTime(scope.row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column prop="processingTime" label="处理时间" width="100">
            <template #default="scope">
              {{ scope.row.processingTime || 0 }}ms
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="scope">
              <el-button size="small" @click="handleViewTask(scope.row)">查看</el-button>
              <el-button 
                size="small" 
                type="primary" 
                @click="handleExecuteTask(scope.row)"
                v-if="scope.row.status === 'pending'"
              >
                执行
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 实时日志 -->
      <el-card class="logs-card">
        <template #header>
          <div class="card-header">
            <span>实时日志</span>
            <div class="header-actions">
              <el-select
                v-model="logLevel"
                placeholder="日志级别"
                size="small"
                style="width: 120px; margin-right: 10px;"
                @change="getAgentLogs"
              >
                <el-option label="全部" value="" />
                <el-option label="DEBUG" value="debug" />
                <el-option label="INFO" value="info" />
                <el-option label="WARN" value="warn" />
                <el-option label="ERROR" value="error" />
              </el-select>
              <el-button 
                type="text" 
                icon="Refresh" 
                @click="getAgentLogs"
                :loading="logsLoading"
              />
            </div>
          </div>
        </template>
        <div class="logs-content" ref="logsContainer">
          <div 
            v-for="log in agentLogs" 
            :key="log.id" 
            :class="['log-item', `log-${log.level}`]"
          >
            <span class="log-time">{{ formatTime(log.timestamp) }}</span>
            <span class="log-level">{{ log.level.toUpperCase() }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
          <div v-if="agentLogs.length === 0" class="empty-logs">
            暂无日志数据
          </div>
        </div>
      </el-card>
    </div>

    <!-- 加载状态 -->
    <div v-else class="loading-container">
      <el-skeleton :rows="10" animated />
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="primary" @click="handleSendMessage">发送消息</el-button>
        <el-button type="warning" @click="handleRestartAgent">重启Agent</el-button>
        <el-button type="danger" @click="handleDeleteAgent">删除Agent</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  getAgent, 
  updateAgentStatus, 
  getTasksByAgent,
  getAgentLogs as getAgentLogsApi,
  agentHeartbeat,
  executeTask,
  deleteAgent,
  updateAgent
} from '@/api/nesma'

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  agentId: {
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

const agentDetail = ref(null)
const taskList = ref([])
const agentLogs = ref([])
const statusLoading = ref(false)
const metricsLoading = ref(false)
const taskLoading = ref(false)
const logsLoading = ref(false)
const taskFilter = ref('')
const logLevel = ref('')
const logsContainer = ref(null)

// 获取Agent详情
const getAgentDetail = async () => {
  if (!props.agentId) return
  
  try {
    const res = await getAgent(props.agentId)
    agentDetail.value = res.data
  } catch (error) {
    ElMessage.error('获取Agent详情失败')
  }
}

// 获取任务列表
const getTaskList = async () => {
  if (!props.agentId) return
  
  try {
    taskLoading.value = true
    const res = await getTasksByAgent(props.agentId, {
      status: taskFilter.value,
      page: 1,
      pageSize: 50
    })
    taskList.value = res.data.list || []
  } catch (error) {
    ElMessage.error('获取任务列表失败')
  } finally {
    taskLoading.value = false
  }
}

// 获取Agent日志
const getAgentLogs = async () => {
  if (!props.agentId) return
  
  try {
    logsLoading.value = true
    const res = await getAgentLogsApi(props.agentId, {
      level: logLevel.value,
      limit: 100
    })
    agentLogs.value = res.data.list || []
    
    // 滚动到底部
    nextTick(() => {
      if (logsContainer.value) {
        logsContainer.value.scrollTop = logsContainer.value.scrollHeight
      }
    })
  } catch (error) {
    ElMessage.error('获取日志失败')
  } finally {
    logsLoading.value = false
  }
}

// 刷新性能指标
const refreshMetrics = async () => {
  metricsLoading.value = true
  await getAgentDetail()
  metricsLoading.value = false
}

// 切换Agent状态
const handleToggleStatus = async () => {
  const newStatus = agentDetail.value.status === 'active' ? 'offline' : 'active'
  const action = newStatus === 'active' ? '启动' : '停止'
  
  try {
    await ElMessageBox.confirm(`确定要${action}此Agent吗？`, '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    statusLoading.value = true
    await updateAgentStatus(props.agentId, { status: newStatus })
    
    ElMessage.success(`${action}成功`)
    agentDetail.value.status = newStatus
    emit('refresh')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(`${action}失败`)
    }
  } finally {
    statusLoading.value = false
  }
}

// 重启Agent
const handleRestartAgent = async () => {
  try {
    await ElMessageBox.confirm('确定要重启此Agent吗？', '确认操作', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 先停止再启动
    await updateAgentStatus(props.agentId, { status: 'offline' })
    setTimeout(async () => {
      await updateAgentStatus(props.agentId, { status: 'active' })
      ElMessage.success('重启成功')
      await getAgentDetail()
      emit('refresh')
    }, 1000)
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重启失败')
    }
  }
}

// 创建任务
const handleCreateTask = () => {
  ElMessage.info('创建任务功能开发中...')
}

// 查看任务详情
const handleViewTask = (task) => {
  ElMessage.info('查看任务详情功能开发中...')
}

// 执行任务
const handleExecuteTask = async (task) => {
  try {
    await executeTask({ taskId: task.taskId })
    ElMessage.success('任务执行成功')
    await getTaskList()
  } catch (error) {
    ElMessage.error('执行任务失败')
  }
}

// 发送消息
const handleSendMessage = () => {
  // 触发父组件的发送消息功能
  emit('send-message', props.agentId)
}

// 删除Agent
const handleDeleteAgent = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要删除Agent "${agentDetail.value?.name}" 吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteAgent(props.agentId)
    ElMessage.success('删除成功')
    emit('refresh')
    handleClose()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  agentDetail.value = null
  taskList.value = []
  agentLogs.value = []
}

// 工具方法
const getAgentTypeLabel = (type) => {
  const labels = {
    'REQUIREMENT_ANALYSIS': '需求分析',
    'NESMA_EVALUATION': 'NESMA评估',
    'PROJECT_MANAGEMENT': '项目管理',
    'KNOWLEDGE_MANAGEMENT': '知识管理',
    'QUALITY_ASSURANCE': '质量检查'
  }
  return labels[type] || type
}

const getAgentTypeColor = (type) => {
  const colors = {
    'REQUIREMENT_ANALYSIS': '',
    'NESMA_EVALUATION': 'success',
    'PROJECT_MANAGEMENT': 'warning',
    'KNOWLEDGE_MANAGEMENT': 'info',
    'QUALITY_ASSURANCE': 'danger'
  }
  return colors[type] || ''
}

const getStatusLabel = (status) => {
  const labels = {
    'active': '活跃',
    'busy': '忙碌',
    'offline': '离线',
    'error': '错误'
  }
  return labels[status] || status
}

const getStatusTagType = (status) => {
  const types = {
    'active': 'success',
    'busy': 'warning',
    'offline': 'info',
    'error': 'danger'
  }
  return types[status] || 'info'
}

const getLoadPercentage = (agent) => {
  if (!agent) return 0
  const max = agent.maxConcurrency || 1
  const current = agent.processingCount || 0
  return Math.round((current / max) * 100)
}

const getLoadColor = (agent) => {
  const percentage = getLoadPercentage(agent)
  if (percentage >= 90) return '#f56c6c'
  if (percentage >= 70) return '#e6a23c'
  return '#67c23a'
}

const getSuccessRate = (agent) => {
  if (!agent || !agent.totalProcessed) return 100
  const errorCount = agent.errorCount || 0
  return Math.round(((agent.totalProcessed - errorCount) / agent.totalProcessed) * 100)
}

const getHeartbeatColor = (lastHeartbeat) => {
  if (!lastHeartbeat) return '#f56c6c'
  
  const now = new Date()
  const heartbeat = new Date(lastHeartbeat)
  const diff = now - heartbeat
  
  if (diff < 60000) return '#67c23a' // 1分钟内 - 绿色
  if (diff < 300000) return '#e6a23c' // 5分钟内 - 橙色
  return '#f56c6c' // 超过5分钟 - 红色
}

const getTaskStatusType = (status) => {
  const types = {
    'pending': 'info',
    'processing': 'warning',
    'completed': 'success',
    'failed': 'danger'
  }
  return types[status] || 'info'
}

const getTaskStatusLabel = (status) => {
  const labels = {
    'pending': '待处理',
    'processing': '执行中',
    'completed': '已完成',
    'failed': '失败'
  }
  return labels[status] || status
}

const formatDateTime = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return `${date.getHours().toString().padStart(2, '0')}:${date.getMinutes().toString().padStart(2, '0')}:${date.getSeconds().toString().padStart(2, '0')}`
}

// 监听对话框显示状态
watch(() => props.visible, (newVal) => {
  if (newVal && props.agentId) {
    getAgentDetail()
    getTaskList()
    getAgentLogs()
  }
})

// 监听agentId变化
watch(() => props.agentId, (newVal) => {
  if (newVal && props.visible) {
    getAgentDetail()
    getTaskList()
    getAgentLogs()
  }
})
</script>

<style scoped>
.agent-detail {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card,
.metrics-card,
.tasks-card,
.logs-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
}

.header-actions {
  display: flex;
  align-items: center;
}

.status-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.heartbeat-item {
  display: flex;
  align-items: center;
  gap: 5px;
}

.metric-card {
  display: flex;
  align-items: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  gap: 15px;
}

.metric-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
}

.metric-info {
  flex: 1;
}

.metric-value {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 5px;
}

.metric-label {
  font-size: 14px;
  color: #909399;
}

.logs-content {
  max-height: 300px;
  overflow-y: auto;
  background: #000;
  border-radius: 4px;
  padding: 10px;
  font-family: 'Courier New', monospace;
}

.log-item {
  display: flex;
  margin-bottom: 5px;
  font-size: 12px;
  line-height: 1.4;
}

.log-time {
  color: #909399;
  margin-right: 10px;
  min-width: 80px;
}

.log-level {
  margin-right: 10px;
  min-width: 50px;
  font-weight: bold;
}

.log-message {
  flex: 1;
  word-break: break-all;
}

.log-debug .log-level {
  color: #909399;
}

.log-info .log-level {
  color: #67c23a;
}

.log-warn .log-level {
  color: #e6a23c;
}

.log-error .log-level {
  color: #f56c6c;
}

.log-debug .log-message {
  color: #909399;
}

.log-info .log-message {
  color: #e1e1e1;
}

.log-warn .log-message {
  color: #e6a23c;
}

.log-error .log-message {
  color: #f56c6c;
}

.empty-logs {
  text-align: center;
  color: #909399;
  padding: 50px 0;
}

.loading-container {
  padding: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 滚动条样式 */
.logs-content::-webkit-scrollbar {
  width: 6px;
}

.logs-content::-webkit-scrollbar-track {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
}

.logs-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.3);
  border-radius: 3px;
}

.logs-content::-webkit-scrollbar-thumb:hover {
  background: rgba(255, 255, 255, 0.5);
}
</style> 