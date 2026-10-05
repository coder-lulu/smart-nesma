<template>
  <el-dialog
    v-model="dialogVisible"
    title="工作流详情"
    width="1200px"
    :before-close="handleClose"
    class="workflow-detail-dialog"
  >
    <div class="detail-content" v-loading="loading">
      <div v-if="workflowDetail" class="workflow-detail">
        <!-- 工作流基本信息 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <div class="header-left">
                <el-icon size="20"><Document /></el-icon>
                <span>基本信息</span>
              </div>
              <div class="header-right">
                <el-tag :type="getStatusTagType(workflowDetail.status)">
                  {{ getStatusLabel(workflowDetail.status) }}
                </el-tag>
                <el-tag v-if="workflowDetail.isTemplate" type="info">模板</el-tag>
              </div>
            </div>
          </template>
          
          <el-descriptions :column="2" border>
            <el-descriptions-item label="工作流ID">
              <el-text class="workflow-id">{{ workflowDetail.workflowId }}</el-text>
              <el-button 
                size="small" 
                text 
                @click="copyToClipboard(workflowDetail.workflowId)"
                style="margin-left: 8px"
              >
                复制
              </el-button>
            </el-descriptions-item>
            <el-descriptions-item label="名称">{{ workflowDetail.name }}</el-descriptions-item>
            <el-descriptions-item label="版本">v{{ workflowDetail.version }}</el-descriptions-item>
            <el-descriptions-item label="分类">
              <el-tag size="small" :type="getCategoryTagType(workflowDetail.category)">
                {{ getCategoryLabel(workflowDetail.category) }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="使用次数">{{ workflowDetail.usageCount || 0 }}</el-descriptions-item>
            <el-descriptions-item label="创建时间">{{ formatTime(workflowDetail.createdAt) }}</el-descriptions-item>
            <el-descriptions-item label="描述" :span="2">
              {{ workflowDetail.description || '暂无描述' }}
            </el-descriptions-item>
          </el-descriptions>
          
          <!-- 标签 -->
          <div class="tags-section" v-if="workflowDetail.tags && workflowDetail.tags.length > 0">
            <div class="tags-label">标签：</div>
            <div class="tags-list">
              <el-tag 
                v-for="tag in workflowDetail.tags" 
                :key="tag" 
                size="small"
                style="margin-right: 8px;"
              >
                {{ tag }}
              </el-tag>
            </div>
          </div>
        </el-card>

        <!-- 执行统计 -->
        <el-card class="stats-card">
          <template #header>
            <div class="card-header">
              <el-icon size="20"><TrendCharts /></el-icon>
              <span>执行统计</span>
            </div>
          </template>
          
          <div class="stats-grid">
            <div class="stat-item">
              <div class="stat-icon total">
                <el-icon size="24"><Operation /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ executionStats.totalExecutions || 0 }}</div>
                <div class="stat-label">总执行次数</div>
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-icon success">
                <el-icon size="24"><CircleCheck /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ executionStats.successfulExecutions || 0 }}</div>
                <div class="stat-label">成功执行</div>
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-icon failed">
                <el-icon size="24"><CircleClose /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ executionStats.failedExecutions || 0 }}</div>
                <div class="stat-label">失败执行</div>
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-icon rate">
                <el-icon size="24"><Odometer /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ executionStats.successRate || 0 }}%</div>
                <div class="stat-label">成功率</div>
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-icon time">
                <el-icon size="24"><Timer /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ executionStats.avgDuration || 0 }}s</div>
                <div class="stat-label">平均时长</div>
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-icon last">
                <el-icon size="24"><Clock /></el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ formatTime(executionStats.lastExecutionTime, 'relative') }}</div>
                <div class="stat-label">最近执行</div>
              </div>
            </div>
          </div>
        </el-card>

        <!-- 工作流定义 -->
        <el-card class="definition-card">
          <template #header>
            <div class="card-header">
              <el-icon size="20"><Connection /></el-icon>
              <span>工作流定义</span>
              <div class="header-actions">
                <el-button size="small" @click="toggleDefinitionView">
                  {{ showDefinitionJson ? '图形视图' : 'JSON视图' }}
                </el-button>
              </div>
            </div>
          </template>
          
          <!-- 图形化显示 -->
          <div v-if="!showDefinitionJson" class="workflow-steps">
            <div 
              v-for="(step, index) in workflowSteps" 
              :key="step.id"
              class="step-item"
            >
              <div class="step-connector" v-if="index > 0"></div>
              <div class="step-node" :class="getStepNodeClass(step.agentType)">
                <div class="step-number">{{ index + 1 }}</div>
                <div class="step-icon">
                  <el-icon size="20">
                    <component :is="getStepIcon(step.agentType)" />
                  </el-icon>
                </div>
                <div class="step-info">
                  <div class="step-name">{{ step.name }}</div>
                  <div class="step-type">{{ getStepTypeLabel(step.agentType) }}</div>
                </div>
                <div class="step-details">
                  <el-tag size="small">{{ step.timeout || 30 }}s 超时</el-tag>
                  <el-tag size="small" type="info">{{ step.retries || 3 }} 次重试</el-tag>
                </div>
              </div>
              <div class="step-description" v-if="step.description">
                {{ step.description }}
              </div>
            </div>
          </div>
          
          <!-- JSON显示 -->
          <div v-else class="definition-json">
            <el-input
              v-model="definitionJson"
              type="textarea"
              :rows="20"
              readonly
              class="json-textarea"
            />
            <div class="json-actions">
              <el-button size="small" @click="copyDefinition">复制定义</el-button>
              <el-button size="small" @click="downloadDefinition">下载定义</el-button>
            </div>
          </div>
        </el-card>

        <!-- 最近执行记录 -->
        <el-card class="executions-card">
          <template #header>
            <div class="card-header">
              <el-icon size="20"><List /></el-icon>
              <span>最近执行记录</span>
              <div class="header-actions">
                <el-button size="small" @click="refreshExecutions">刷新</el-button>
                <el-button size="small" @click="viewAllExecutions">查看全部</el-button>
              </div>
            </div>
          </template>
          
          <el-table :data="recentExecutions" v-loading="executionsLoading">
            <el-table-column prop="executionId" label="执行ID" width="200">
              <template #default="{ row }">
                <el-text class="execution-id">{{ row.executionId.substring(0, 8) }}...</el-text>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="getExecutionStatusType(row.status)">
                  {{ getExecutionStatusLabel(row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="currentStep" label="当前步骤" width="150">
              <template #default="{ row }">
                {{ row.currentStep || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="startTime" label="开始时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.startTime) }}
              </template>
            </el-table-column>
            <el-table-column prop="totalDuration" label="耗时" width="100">
              <template #default="{ row }">
                {{ row.totalDuration ? `${row.totalDuration}s` : '-' }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="150">
              <template #default="{ row }">
                <el-button size="small" @click="viewExecution(row)">查看详情</el-button>
                <el-button 
                  v-if="row.status === 'running'" 
                  size="small" 
                  type="danger" 
                  @click="cancelExecution(row)"
                >
                  取消
                </el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <div v-if="recentExecutions.length === 0" class="no-executions">
            <el-empty description="暂无执行记录" />
          </div>
        </el-card>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="primary" @click="executeWorkflow">执行工作流</el-button>
        <el-button @click="editWorkflow">编辑工作流</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Document, TrendCharts, Connection, List, Operation, CircleCheck, CircleClose,
  Odometer, Timer, Clock, VideoPlay, Edit, Plus, Download, Search
} from '@element-plus/icons-vue'

import { 
  getWorkflow,
  getWorkflowExecutions,
  cancelWorkflowExecution as cancelWorkflowExecutionAPI
} from '@/api/nesma'

const router = useRouter()

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
const emit = defineEmits(['update:visible', 'refresh'])

// 响应式数据
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const loading = ref(false)
const executionsLoading = ref(false)
const workflowDetail = ref(null)
const executionStats = ref({})
const recentExecutions = ref([])
const showDefinitionJson = ref(false)
const definitionJson = ref('')
const workflowSteps = ref([])

// 获取工作流详情
const getWorkflowDetail = async () => {
  if (!props.workflowId) return
  
  loading.value = true
  try {
    const res = await getWorkflow(props.workflowId)
    if (res.code === 0) {
      workflowDetail.value = res.data
      
      // 解析工作流定义
      if (res.data.definition) {
        const definition = JSON.parse(res.data.definition)
        workflowSteps.value = definition.steps || []
        definitionJson.value = JSON.stringify(definition, null, 2)
      }
      
      // 计算执行统计
      calculateExecutionStats()
    }
  } catch (error) {
    ElMessage.error('获取工作流详情失败')
  } finally {
    loading.value = false
  }
}

// 获取执行记录
const getExecutions = async () => {
  if (!props.workflowId) return
  
  executionsLoading.value = true
  try {
    const res = await getWorkflowExecutions(props.workflowId, {
      page: 1,
      pageSize: 10,
      orderBy: 'startTime',
      order: 'desc'
    })
    if (res.code === 0) {
      recentExecutions.value = res.data?.executions || []
    }
  } catch (error) {
    console.error('获取执行记录失败:', error)
  } finally {
    executionsLoading.value = false
  }
}

// 计算执行统计
const calculateExecutionStats = () => {
  const executions = recentExecutions.value
  const total = executions.length
  const successful = executions.filter(e => e.status === 'completed').length
  const failed = executions.filter(e => e.status === 'failed').length
  const successRate = total > 0 ? Math.round((successful / total) * 100) : 0
  
  const durations = executions
    .filter(e => e.totalDuration)
    .map(e => e.totalDuration)
  const avgDuration = durations.length > 0 
    ? Math.round(durations.reduce((a, b) => a + b, 0) / durations.length)
    : 0
  
  const lastExecution = executions[0]
  
  executionStats.value = {
    totalExecutions: total,
    successfulExecutions: successful,
    failedExecutions: failed,
    successRate,
    avgDuration,
    lastExecutionTime: lastExecution?.startTime
  }
}

// 刷新执行记录
const refreshExecutions = () => {
  getExecutions()
}

// 查看所有执行记录
const viewAllExecutions = () => {
  emit('view-all-executions', props.workflowId)
}

// 查看执行详情
const viewExecution = (execution) => {
  emit('view-execution', execution.executionId)
}

// 取消执行
const cancelExecution = async (execution) => {
  try {
    await cancelWorkflowExecutionAPI(execution.executionId)
    ElMessage.success('取消成功')
    getExecutions()
  } catch (error) {
    ElMessage.error('取消失败')
  }
}

// 执行工作流
const executeWorkflow = () => {
  emit('execute-workflow', workflowDetail.value)
  handleClose()
}

// 编辑工作流
const editWorkflow = () => {
  router.push(`/nesma/workflow/designer?id=${props.workflowId}`)
  handleClose()
}

// 切换定义视图
const toggleDefinitionView = () => {
  showDefinitionJson.value = !showDefinitionJson.value
}

// 复制定义
const copyDefinition = async () => {
  try {
    await navigator.clipboard.writeText(definitionJson.value)
    ElMessage.success('定义已复制到剪贴板')
  } catch (error) {
    ElMessage.error('复制失败')
  }
}

// 下载定义
const downloadDefinition = () => {
  const blob = new Blob([definitionJson.value], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${workflowDetail.value.name}_definition.json`
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
    active: 'success',
    inactive: 'info',
    draft: 'warning'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    active: '活跃',
    inactive: '非活跃',
    draft: '草稿'
  }
  return labelMap[status] || status
}

const getCategoryTagType = (category) => {
  const typeMap = {
    requirement_analysis: 'primary',
    nesma_evaluation: 'success',
    document_generation: 'warning',
    quality_check: 'danger'
  }
  return typeMap[category] || 'info'
}

const getCategoryLabel = (category) => {
  const labelMap = {
    requirement_analysis: '需求分析',
    nesma_evaluation: 'NESMA评估',
    document_generation: '文档生成',
    quality_check: '质量检查'
  }
  return labelMap[category] || category
}

const getExecutionStatusType = (status) => {
  const typeMap = {
    completed: 'success',
    failed: 'danger',
    running: 'primary',
    cancelled: 'warning',
    pending: 'info'
  }
  return typeMap[status] || 'info'
}

const getExecutionStatusLabel = (status) => {
  const labelMap = {
    completed: '成功',
    failed: '失败',
    running: '运行中',
    cancelled: '已取消',
    pending: '等待中'
  }
  return labelMap[status] || status
}

const getStepNodeClass = (agentType) => {
  const classMap = {
    'START': 'start-node',
    'END': 'end-node',
    'REQUIREMENT_ANALYSIS': 'requirement-node',
    'NESMA_EVALUATION': 'nesma-node',
    'KNOWLEDGE_RETRIEVAL': 'knowledge-node',
    'DOCUMENT_GENERATION': 'document-node'
  }
  return classMap[agentType] || 'default-node'
}

const getStepIcon = (agentType) => {
  const iconMap = {
    'START': 'VideoPlay',
    'END': 'CircleCheck',
    'REQUIREMENT_ANALYSIS': 'Document',
    'NESMA_EVALUATION': 'TrendCharts',
    'KNOWLEDGE_RETRIEVAL': 'Search',
    'DOCUMENT_GENERATION': 'Edit'
  }
  return iconMap[agentType] || 'Operation'
}

const getStepTypeLabel = (agentType) => {
  const labelMap = {
    'START': '开始节点',
    'END': '结束节点',
    'REQUIREMENT_ANALYSIS': '需求分析',
    'NESMA_EVALUATION': 'NESMA评估',
    'KNOWLEDGE_RETRIEVAL': '知识检索',
    'DOCUMENT_GENERATION': '文档生成'
  }
  return labelMap[agentType] || agentType
}

const formatTime = (time, type = 'absolute') => {
  if (!time) return '-'
  
  const date = new Date(time)
  
  if (type === 'relative') {
    const now = new Date()
    const diff = now - date
    const days = Math.floor(diff / (1000 * 60 * 60 * 24))
    const hours = Math.floor(diff / (1000 * 60 * 60))
    const minutes = Math.floor(diff / (1000 * 60))
    
    if (days > 0) return `${days}天前`
    if (hours > 0) return `${hours}小时前`
    if (minutes > 0) return `${minutes}分钟前`
    return '刚刚'
  }
  
  return date.toLocaleString('zh-CN')
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
}

// 监听props变化
watch(() => props.workflowId, (newId) => {
  if (newId && props.visible) {
    getWorkflowDetail()
    getExecutions()
  }
})

watch(() => props.visible, (visible) => {
  if (visible && props.workflowId) {
    getWorkflowDetail()
    getExecutions()
  }
})
</script>

<style scoped>
.workflow-detail-dialog {
  --el-dialog-border-radius: 16px;
}

.detail-content {
  max-height: 80vh;
  overflow-y: auto;
}

.workflow-detail {
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.info-card, .stats-card, .definition-card, .executions-card {
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

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.workflow-id {
  font-family: 'Monaco', 'Consolas', monospace;
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.tags-section {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 16px;
  padding-top: 16px;
  border-top: 1px solid #e5e7eb;
}

.tags-label {
  font-weight: 500;
  color: #374151;
  flex-shrink: 0;
}

.tags-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 20px;
}

.stat-item {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  background: #f8fafc;
  border-radius: 12px;
  border: 1px solid #e5e7eb;
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
.stat-icon.rate { background: linear-gradient(135deg, #a8edea, #fed6e3); }
.stat-icon.time { background: linear-gradient(135deg, #ffecd2, #fcb69f); }
.stat-icon.last { background: linear-gradient(135deg, #89f7fe, #66a6ff); }

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

.workflow-steps {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.step-item {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-connector {
  position: absolute;
  top: -10px;
  left: 50px;
  width: 2px;
  height: 20px;
  background: #d1d5db;
}

.step-node {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 20px;
  border-radius: 12px;
  border: 2px solid #e5e7eb;
  background: white;
}

.step-node.start-node { border-color: #10b981; background: #ecfdf5; }
.step-node.end-node { border-color: #ef4444; background: #fef2f2; }
.step-node.requirement-node { border-color: #3b82f6; background: #eff6ff; }
.step-node.nesma-node { border-color: #8b5cf6; background: #f5f3ff; }
.step-node.knowledge-node { border-color: #f59e0b; background: #fffbeb; }
.step-node.document-node { border-color: #06b6d4; background: #f0fdfa; }

.step-number {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: #3b82f6;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  flex-shrink: 0;
}

.step-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  flex-shrink: 0;
}

.step-info {
  flex: 1;
}

.step-info .step-name {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.step-info .step-type {
  font-size: 13px;
  color: #6b7280;
}

.step-details {
  display: flex;
  gap: 8px;
  flex-shrink: 0;
}

.step-description {
  margin-left: 88px;
  padding: 12px 16px;
  background: #f8fafc;
  border-radius: 8px;
  font-size: 13px;
  color: #6b7280;
  line-height: 1.5;
}

.definition-json {
  position: relative;
}

.json-textarea {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
}

.json-actions {
  margin-top: 12px;
  display: flex;
  gap: 12px;
}

.execution-id {
  font-family: 'Monaco', 'Consolas', monospace;
  background: #f3f4f6;
  padding: 2px 6px;
  border-radius: 4px;
  font-size: 12px;
}

.no-executions {
  padding: 40px 0;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 0 0 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: 1fr;
  }
  
  .step-node {
    flex-direction: column;
    text-align: center;
    gap: 12px;
  }
  
  .step-details {
    justify-content: center;
  }
  
  .step-description {
    margin-left: 0;
  }
}

/* 表格样式增强 */
:deep(.el-table) {
  border-radius: 8px;
  overflow: hidden;
}

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
</style> 