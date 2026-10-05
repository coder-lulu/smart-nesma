<template>
  <div class="agent-container">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleCreateAgent">
          注册Agent
        </el-button>
        <el-button icon="VideoPlay" @click="handleExecuteWorkflow">
          执行工作流
        </el-button>
        <el-button icon="MessageBox" @click="handleBroadcastMessage">
          广播消息
        </el-button>
        <el-button 
          icon="Delete" 
          type="danger" 
          :disabled="!selectedAgents.length"
          @click="handleBatchDelete"
        >
          批量删除
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchForm.keyword"
          placeholder="搜索Agent名称/ID"
          style="width: 200px; margin-right: 10px;"
          clearable
          @keyup.enter="getAgentList"
        />
        <el-select
          v-model="searchForm.agentType"
          placeholder="Agent类型"
          style="width: 150px; margin-right: 10px;"
          clearable
        >
          <el-option label="需求分析" value="REQUIREMENT_ANALYSIS" />
          <el-option label="NESMA评估" value="NESMA_EVALUATION" />
          <el-option label="项目管理" value="PROJECT_MANAGEMENT" />
          <el-option label="知识管理" value="KNOWLEDGE_MANAGEMENT" />
          <el-option label="质量检查" value="QUALITY_ASSURANCE" />
        </el-select>
        <el-select
          v-model="searchForm.status"
          placeholder="状态"
          style="width: 120px; margin-right: 10px;"
          clearable
        >
          <el-option label="活跃" value="active" />
          <el-option label="忙碌" value="busy" />
          <el-option label="离线" value="offline" />
          <el-option label="错误" value="error" />
        </el-select>
        <el-button icon="Search" @click="getAgentList">搜索</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div class="stats-cards" v-if="agentStats">
      <el-row :gutter="20">
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-icon">
                <el-icon><Connection /></el-icon>
              </div>
              <div class="stats-info">
                <div class="stats-number">{{ agentStats.totalAgents }}</div>
                <div class="stats-label">总Agent数</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-icon active">
                <el-icon><VideoPlay /></el-icon>
              </div>
              <div class="stats-info">
                <div class="stats-number">{{ agentStats.activeAgents }}</div>
                <div class="stats-label">活跃Agent</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-icon processing">
                <el-icon><Loading /></el-icon>
              </div>
              <div class="stats-info">
                <div class="stats-number">{{ agentStats.processingTasks }}</div>
                <div class="stats-label">处理中任务</div>
              </div>
            </div>
          </el-card>
        </el-col>
        <el-col :span="6">
          <el-card class="stats-card">
            <div class="stats-content">
              <div class="stats-icon completed">
                <el-icon><Check /></el-icon>
              </div>
              <div class="stats-info">
                <div class="stats-number">{{ agentStats.completedTasks }}</div>
                <div class="stats-label">已完成任务</div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 主要内容区域 -->
    <el-row :gutter="20">
      <!-- Agent列表 -->
      <el-col :span="16">
        <el-card class="table-card">
          <template #header>
            <div class="card-header">
              <span>Agent列表</span>
              <el-button 
                type="text" 
                icon="Refresh" 
                @click="getAgentList"
                :loading="loading"
              />
            </div>
          </template>
          <el-table
            v-loading="loading"
            :data="agentList"
            style="width: 100%"
            @selection-change="handleSelectionChange"
            @row-click="handleRowClick"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="agentId" label="Agent ID" width="120" show-overflow-tooltip />
            <el-table-column prop="name" label="名称" min-width="150" show-overflow-tooltip />
            <el-table-column prop="agentType" label="类型" width="120">
              <template #default="scope">
                <el-tag size="small" :type="getAgentTypeColor(scope.row.agentType)">
                  {{ getAgentTypeLabel(scope.row.agentType) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag 
                  :type="getStatusTagType(scope.row.status)" 
                  size="small"
                  :effect="scope.row.status === 'active' ? 'light' : 'plain'"
                >
                  <el-icon v-if="scope.row.status === 'active'" class="status-icon">
                    <CircleCheck />
                  </el-icon>
                  <el-icon v-else-if="scope.row.status === 'busy'" class="status-icon">
                    <Loading />
                  </el-icon>
                  <el-icon v-else-if="scope.row.status === 'offline'" class="status-icon">
                    <CircleClose />
                  </el-icon>
                  <el-icon v-else class="status-icon">
                    <Warning />
                  </el-icon>
                  {{ getStatusLabel(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="负载" width="120">
              <template #default="scope">
                <div class="load-info">
                  <span>{{ scope.row.processingCount || 0 }}/{{ scope.row.maxConcurrency || 0 }}</span>
                  <el-progress
                    :percentage="getLoadPercentage(scope.row)"
                    :color="getLoadColor(scope.row)"
                    :stroke-width="3"
                    :show-text="false"
                    style="margin-top: 2px;"
                  />
                </div>
              </template>
            </el-table-column>
            <el-table-column label="性能" width="120">
              <template #default="scope">
                <div class="performance-info">
                  <div class="performance-item">
                    <span class="performance-label">响应时间:</span>
                    <span class="performance-value">{{ scope.row.averageResponseTime || 0 }}ms</span>
                  </div>
                  <div class="performance-item">
                    <span class="performance-label">处理总数:</span>
                    <span class="performance-value">{{ scope.row.totalProcessed || 0 }}</span>
                  </div>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="最后心跳" width="120">
              <template #default="scope">
                <span class="heartbeat-time">{{ formatTime(scope.row.lastHeartbeat) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150" fixed="right">
              <template #default="scope">
                <el-button-group>
                  <el-button 
                    type="primary" 
                    size="small" 
                    @click.stop="handleViewAgent(scope.row)"
                  >
                    详情
                  </el-button>
                  <el-dropdown @command="(command) => handleDropdownCommand(command, scope.row)">
                    <el-button type="primary" size="small">
                      更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="message">发送消息</el-dropdown-item>
                        <el-dropdown-item command="assign">分配任务</el-dropdown-item>
                        <el-dropdown-item command="config">配置</el-dropdown-item>
                        <el-dropdown-item 
                          command="delete" 
                          style="color: #f56c6c;"
                        >
                          删除
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </el-button-group>
              </template>
            </el-table-column>
          </el-table>
          
          <!-- 分页 -->
          <div class="pagination-wrapper">
            <el-pagination
              v-model:current-page="pageInfo.page"
              v-model:page-size="pageInfo.pageSize"
              :page-sizes="[10, 20, 50, 100]"
              :total="pageInfo.total"
              layout="total, sizes, prev, pager, next, jumper"
              @size-change="handleSizeChange"
              @current-change="handleCurrentChange"
            />
          </div>
        </el-card>
      </el-col>
      
      <!-- 右侧面板 -->
      <el-col :span="8">
        <!-- 系统监控 -->
        <el-card class="monitor-card">
          <template #header>
            <div class="card-header">
              <span>系统监控</span>
              <el-button 
                type="text" 
                icon="Refresh" 
                @click="getSystemMetrics"
                :loading="metricsLoading"
              />
            </div>
          </template>
          <div class="metrics-content">
            <div class="metric-item">
              <div class="metric-label">CPU使用率</div>
              <div class="metric-value">
                <el-progress
                  :percentage="systemMetrics.cpuUsage"
                  :color="systemMetrics.cpuUsage > 80 ? '#f56c6c' : '#67c23a'"
                  :stroke-width="8"
                />
              </div>
            </div>
            <div class="metric-item">
              <div class="metric-label">内存使用率</div>
              <div class="metric-value">
                <el-progress
                  :percentage="systemMetrics.memoryUsage"
                  :color="systemMetrics.memoryUsage > 80 ? '#f56c6c' : '#67c23a'"
                  :stroke-width="8"
                />
              </div>
            </div>
            <div class="metric-item">
              <div class="metric-label">消息队列</div>
              <div class="metric-value">
                <span class="queue-size">{{ systemMetrics.messageQueueSize }}</span>
              </div>
            </div>
            <div class="metric-item">
              <div class="metric-label">重试队列</div>
              <div class="metric-value">
                <span class="queue-size">{{ systemMetrics.retryQueueSize }}</span>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 最近活动 -->
        <el-card class="activity-card">
          <template #header>
            <div class="card-header">
              <span>最近活动</span>
              <el-button 
                type="text" 
                icon="Refresh" 
                @click="getRecentActivities"
                :loading="activityLoading"
              />
            </div>
          </template>
          <div class="activity-content">
            <el-timeline>
              <el-timeline-item
                v-for="activity in recentActivities"
                :key="activity.id"
                :timestamp="formatTime(activity.timestamp)"
                :type="getActivityType(activity.type)"
              >
                <div class="activity-item">
                  <div class="activity-title">{{ activity.title }}</div>
                  <div class="activity-description">{{ activity.description }}</div>
                  <div class="activity-agent">{{ activity.agentName }}</div>
                </div>
              </el-timeline-item>
            </el-timeline>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- Agent详情对话框 -->
    <agent-detail-dialog
      v-model:visible="showAgentDetail"
      :agent-id="selectedAgentId"
      @refresh="getAgentList"
    />

    <!-- 创建Agent对话框 -->
    <create-agent-dialog
      v-model:visible="showCreateAgent"
      @refresh="getAgentList"
    />

    <!-- 发送消息对话框 -->
    <send-message-dialog
      v-model:visible="showSendMessage"
      :agent-id="selectedAgentId"
    />

    <!-- 工作流执行对话框 -->
    <workflow-execution-dialog
      v-model:visible="showWorkflowExecution"
      @refresh="getAgentList"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  getAgentList as getAgentListApi, 
  getAgentStatistics,
  getSystemMetrics as getSystemMetricsApi,
  agentHeartbeat,
  updateAgentStatus,
  deleteAgent
} from '@/api/nesma'
import AgentDetailDialog from './components/AgentDetailDialog.vue'
import CreateAgentDialog from './components/CreateAgentDialog.vue'
import SendMessageDialog from './components/SendMessageDialog.vue'
import WorkflowExecutionDialog from './components/WorkflowExecutionDialog.vue'

// 响应式数据
const loading = ref(false)
const metricsLoading = ref(false)
const activityLoading = ref(false)
const agentList = ref([])
const selectedAgents = ref([])
const selectedAgentId = ref('')
const agentStats = ref(null)
const systemMetrics = ref({
  cpuUsage: 0,
  memoryUsage: 0,
  messageQueueSize: 0,
  retryQueueSize: 0
})
const recentActivities = ref([])

// 搜索表单
const searchForm = reactive({
  keyword: '',
  agentType: '',
  status: ''
})

// 分页信息
const pageInfo = reactive({
  page: 1,
  pageSize: 10,
  total: 0
})

// 对话框显示状态
const showAgentDetail = ref(false)
const showCreateAgent = ref(false)
const showSendMessage = ref(false)
const showWorkflowExecution = ref(false)

// 获取Agent列表
const getAgentList = async () => {
  try {
    loading.value = true
    const res = await getAgentListApi({
      page: pageInfo.page,
      pageSize: pageInfo.pageSize,
      ...searchForm
    })
    agentList.value = res.data.agents || []
    pageInfo.total = res.data.total || res.data.agents?.length || 0
  } catch (error) {
    ElMessage.error('获取Agent列表失败')
  } finally {
    loading.value = false
  }
}

// 获取Agent统计信息
const getAgentStats = async () => {
  try {
    const res = await getAgentStatistics()
    agentStats.value = res.data
  } catch (error) {
    ElMessage.error('获取统计信息失败')
  }
}

// 获取系统监控指标
const getSystemMetrics = async () => {
  try {
    metricsLoading.value = true
    const res = await getSystemMetricsApi()
    systemMetrics.value = res.data
  } catch (error) {
    ElMessage.error('获取系统监控指标失败')
  } finally {
    metricsLoading.value = false
  }
}

// 获取最近活动
const getRecentActivities = async () => {
  try {
    activityLoading.value = true
    // 模拟数据，实际应该从API获取
    recentActivities.value = [
      {
        id: 1,
        title: 'Agent任务完成',
        description: '需求分析任务已完成',
        agentName: 'Agent-001',
        timestamp: new Date(),
        type: 'success'
      },
      {
        id: 2,
        title: '新Agent注册',
        description: '新的NESMA评估Agent已注册',
        agentName: 'Agent-002',
        timestamp: new Date(Date.now() - 300000),
        type: 'info'
      }
    ]
  } catch (error) {
    ElMessage.error('获取最近活动失败')
  } finally {
    activityLoading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  Object.assign(searchForm, {
    keyword: '',
    agentType: '',
    status: ''
  })
  getAgentList()
}

// 处理选择变化
const handleSelectionChange = (selection) => {
  selectedAgents.value = selection
}

// 处理行点击
const handleRowClick = (row) => {
  selectedAgentId.value = row.agentId
  showAgentDetail.value = true
}

// 处理分页变化
const handleSizeChange = (size) => {
  pageInfo.pageSize = size
  getAgentList()
}

const handleCurrentChange = (page) => {
  pageInfo.page = page
  getAgentList()
}

// 处理创建Agent
const handleCreateAgent = () => {
  showCreateAgent.value = true
}

// 处理查看Agent详情
const handleViewAgent = (row) => {
  selectedAgentId.value = row.agentId
  showAgentDetail.value = true
}

// 处理下拉菜单命令
const handleDropdownCommand = (command, row) => {
  selectedAgentId.value = row.agentId
  switch (command) {
    case 'message':
      showSendMessage.value = true
      break
    case 'assign':
      // 实现分配任务逻辑
      break
    case 'config':
      // 实现配置逻辑
      break
    case 'delete':
      handleDeleteAgent(row)
      break
  }
}

// 处理删除Agent
const handleDeleteAgent = async (row) => {
  try {
    await ElMessageBox.confirm(`确定要删除Agent "${row.name}" 吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteAgent(row.agentId)
    ElMessage.success('删除成功')
    getAgentList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除失败')
    }
  }
}

// 处理批量删除
const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedAgents.value.length} 个Agent吗？`, '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    // 并行删除所有选中的Agent
    const deletePromises = selectedAgents.value.map(agent => deleteAgent(agent.agentId))
    await Promise.all(deletePromises)
    
    ElMessage.success('批量删除成功')
    selectedAgents.value = []
    getAgentList()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
    }
  }
}

// 处理执行工作流
const handleExecuteWorkflow = () => {
  showWorkflowExecution.value = true
}

// 处理广播消息
const handleBroadcastMessage = () => {
  selectedAgentId.value = 'broadcast'
  showSendMessage.value = true
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

const getActivityType = (type) => {
  const types = {
    'success': 'success',
    'info': 'primary',
    'warning': 'warning',
    'error': 'danger'
  }
  return types[type] || 'primary'
}

const formatTime = (time) => {
  if (!time) return '-'
  const date = new Date(time)
  return `${date.getHours()}:${date.getMinutes().toString().padStart(2, '0')}`
}

// 定时刷新数据
const startAutoRefresh = () => {
  setInterval(() => {
    getAgentList()
    getAgentStats()
    getSystemMetrics()
  }, 10000) // 每10秒刷新一次
}

// 组件挂载时获取数据
onMounted(() => {
  getAgentList()
  getAgentStats()
  getSystemMetrics()
  getRecentActivities()
  startAutoRefresh()
})
</script>

<style scoped>
.agent-container {
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 15px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.toolbar-left {
  display: flex;
  gap: 10px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.stats-cards {
  margin-bottom: 20px;
}

.stats-card {
  border: none;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.stats-content {
  display: flex;
  align-items: center;
  gap: 15px;
}

.stats-icon {
  width: 50px;
  height: 50px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  color: #fff;
  background: #409eff;
}

.stats-icon.active {
  background: #67c23a;
}

.stats-icon.processing {
  background: #e6a23c;
}

.stats-icon.completed {
  background: #f56c6c;
}

.stats-info {
  flex: 1;
}

.stats-number {
  font-size: 28px;
  font-weight: bold;
  color: #303133;
  line-height: 1;
}

.stats-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.table-card {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  border: none;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.load-info {
  font-size: 12px;
  color: #606266;
}

.performance-info {
  font-size: 11px;
  color: #909399;
}

.performance-item {
  display: flex;
  justify-content: space-between;
  margin-bottom: 2px;
}

.performance-label {
  color: #909399;
}

.performance-value {
  color: #303133;
  font-weight: 500;
}

.heartbeat-time {
  font-size: 12px;
  color: #909399;
}

.pagination-wrapper {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.monitor-card {
  margin-bottom: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  border: none;
}

.metrics-content {
  padding: 10px 0;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.metric-label {
  font-size: 14px;
  color: #606266;
  min-width: 80px;
}

.metric-value {
  flex: 1;
  margin-left: 10px;
}

.queue-size {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
}

.activity-card {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  border: none;
}

.activity-content {
  max-height: 400px;
  overflow-y: auto;
}

.activity-item {
  padding: 5px 0;
}

.activity-title {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 3px;
}

.activity-description {
  font-size: 12px;
  color: #606266;
  margin-bottom: 3px;
}

.activity-agent {
  font-size: 11px;
  color: #909399;
}

.status-icon {
  margin-right: 3px;
}

.el-button-group {
  display: flex;
}

.el-button-group .el-button {
  margin-left: 0;
}

.el-button-group .el-button + .el-button {
  margin-left: -1px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .toolbar {
    flex-direction: column;
    gap: 10px;
  }
  
  .toolbar-left,
  .toolbar-right {
    width: 100%;
    justify-content: center;
  }
  
  .stats-cards .el-col {
    margin-bottom: 10px;
  }
}
</style> 