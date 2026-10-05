<template>
  <el-dialog
    v-model="dialogVisible"
    title="执行工作流"
    width="800px"
    :before-close="handleClose"
  >
    <div class="workflow-execution">
      <!-- 选择工作流 -->
      <el-card class="section-card">
        <template #header>
          <span>选择工作流</span>
        </template>
        <el-form label-width="120px">
          <el-form-item label="工作流">
            <el-select
              v-model="selectedWorkflowId"
              placeholder="请选择要执行的工作流"
              style="width: 100%"
              filterable
              @change="handleWorkflowChange"
            >
              <el-option
                v-for="workflow in workflowList"
                :key="workflow.workflowId"
                :label="workflow.name"
                :value="workflow.workflowId"
              >
                <div class="workflow-option">
                  <div class="workflow-name">{{ workflow.name }}</div>
                  <div class="workflow-desc">{{ workflow.description || '暂无描述' }}</div>
                  <div class="workflow-meta">
                    <el-tag size="small" type="info">{{ workflow.steps?.length || 0 }} 个步骤</el-tag>
                    <el-tag size="small" :type="getStatusTagType(workflow.status)">
                      {{ getStatusText(workflow.status) }}
                    </el-tag>
                  </div>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 工作流详情 -->
      <el-card class="section-card" v-if="selectedWorkflow">
        <template #header>
          <span>工作流详情</span>
        </template>
        <el-descriptions :column="2" border>
          <el-descriptions-item label="工作流名称">{{ selectedWorkflow.name }}</el-descriptions-item>
          <el-descriptions-item label="步骤数量">{{ selectedWorkflow.steps?.length || 0 }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusTagType(selectedWorkflow.status)">
              {{ getStatusText(selectedWorkflow.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="成功率">{{ selectedWorkflow.successRate || 0 }}%</el-descriptions-item>
          <el-descriptions-item label="描述" :span="2">
            {{ selectedWorkflow.description || '暂无描述' }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 工作流步骤 -->
        <div class="workflow-steps" v-if="selectedWorkflow.steps && selectedWorkflow.steps.length > 0">
          <h4>执行步骤</h4>
          <div class="steps-timeline">
            <div
              v-for="(step, index) in selectedWorkflow.steps"
              :key="index"
              class="step-item"
            >
              <div class="step-number">{{ index + 1 }}</div>
              <div class="step-content">
                <div class="step-name">{{ step.name || `步骤${index + 1}` }}</div>
                <div class="step-description">{{ step.description || '暂无描述' }}</div>
                <div class="step-meta">
                  <el-tag size="small">{{ getAgentTypeLabel(step.agentType) }}</el-tag>
                  <el-tag size="small" type="info">{{ step.timeout || 30 }}s 超时</el-tag>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-card>

      <!-- 执行参数 -->
      <el-card class="section-card" v-if="selectedWorkflow">
        <template #header>
          <span>执行参数</span>
        </template>
        <el-form :model="executionForm" label-width="120px">
          <el-form-item label="执行名称">
            <el-input
              v-model="executionForm.name"
              placeholder="可选，为此次执行命名"
              maxlength="100"
            />
          </el-form-item>
          
          <el-form-item label="优先级">
            <el-select v-model="executionForm.priority" style="width: 200px">
              <el-option label="低" value="low" />
              <el-option label="普通" value="normal" />
              <el-option label="高" value="high" />
              <el-option label="紧急" value="urgent" />
            </el-select>
          </el-form-item>

          <el-form-item label="输入参数">
            <el-input
              v-model="executionForm.inputData"
              type="textarea"
              :rows="8"
              placeholder="请输入JSON格式的执行参数"
            />
            <div class="form-actions">
                          <el-button type="text" size="small" @click="validateInputData">验证JSON</el-button>
            <el-button type="text" size="small" @click="formatInputData">格式化JSON</el-button>
            <el-button type="text" size="small" @click="loadSampleData">加载示例</el-button>
            </div>
          </el-form-item>

          <el-form-item label="执行选项">
            <el-checkbox-group v-model="executionForm.options">
              <el-checkbox value="enableDetailLogs">启用详细日志</el-checkbox>
              <el-checkbox value="stopOnError">出错时停止</el-checkbox>
              <el-checkbox value="notifyOnComplete">完成时通知</el-checkbox>
              <el-checkbox value="saveIntermediate">保存中间结果</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </el-form>
      </el-card>

      <!-- 执行预览 -->
      <el-card class="section-card" v-if="selectedWorkflow && executionPreview">
        <template #header>
          <span>执行预览</span>
        </template>
        <el-alert
          title="预览信息"
          type="info"
          :closable="false"
          style="margin-bottom: 16px"
        >
          <template #default>
            <p>执行时间预估：{{ executionPreview.estimatedDuration }}秒</p>
            <p>所需Agent类型：{{ executionPreview.requiredAgentTypes.join(', ') }}</p>
            <p>可用Agent数量：{{ executionPreview.availableAgents }}</p>
          </template>
        </el-alert>

        <div class="execution-warning" v-if="executionPreview.warnings.length > 0">
          <el-alert
            v-for="(warning, index) in executionPreview.warnings"
            :key="index"
            :title="warning"
            type="warning"
            :closable="false"
            style="margin-bottom: 8px"
          />
        </div>
      </el-card>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button @click="handleRefresh" :icon="Refresh">刷新工作流</el-button>
        <el-button 
          type="primary" 
          @click="handleExecute" 
          :loading="executing"
          :disabled="!selectedWorkflow || executing"
        >
          开始执行
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { 
  getWorkflowList, 
  getWorkflow, 
  executeWorkflow,
  getAvailableAgents
} from '@/api/nesma'

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
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
const executing = ref(false)
const workflowList = ref([])
const selectedWorkflowId = ref('')
const selectedWorkflow = ref(null)
const executionPreview = ref(null)

// 执行表单
const executionForm = reactive({
  name: '',
  priority: 'normal',
  inputData: '{}',
  options: ['stopOnError']
})

// 获取工作流列表
const getWorkflows = async () => {
  try {
    loading.value = true
    const res = await getWorkflowList({
      status: 'published', // 只获取已发布的工作流
      page: 1,
      pageSize: 100
    })
    workflowList.value = res.data?.workflows || []
  } catch (error) {
    ElMessage.error('获取工作流列表失败')
  } finally {
    loading.value = false
  }
}

// 处理工作流选择变化
const handleWorkflowChange = async (workflowId) => {
  if (!workflowId) {
    selectedWorkflow.value = null
    executionPreview.value = null
    return
  }

  try {
    const res = await getWorkflow(workflowId)
    selectedWorkflow.value = res.data
    
    // 生成执行预览
    await generateExecutionPreview()
  } catch (error) {
    ElMessage.error('获取工作流详情失败')
  }
}

// 生成执行预览
const generateExecutionPreview = async () => {
  if (!selectedWorkflow.value) return

  try {
    const workflow = selectedWorkflow.value
    const steps = workflow.steps || []
    
    // 计算预估执行时间
    const estimatedDuration = steps.reduce((total, step) => {
      return total + (step.timeout || 30)
    }, 0)

    // 获取所需Agent类型
    const requiredAgentTypes = [...new Set(steps.map(step => step.agentType))]
    
    // 检查可用Agent
    const availableAgentsRes = await getAvailableAgents()
    const availableAgents = availableAgentsRes.data?.length || 0
    
    // 生成警告
    const warnings = []
    if (availableAgents === 0) {
      warnings.push('当前没有可用的Agent，工作流可能无法执行')
    }
    if (estimatedDuration > 300) {
      warnings.push('预估执行时间较长，请确保有足够的时间等待')
    }

    executionPreview.value = {
      estimatedDuration,
      requiredAgentTypes: requiredAgentTypes.map(type => getAgentTypeLabel(type)),
      availableAgents,
      warnings
    }
  } catch (error) {
    console.error('生成执行预览失败:', error)
  }
}

// 验证输入数据
const validateInputData = () => {
  try {
    JSON.parse(executionForm.inputData)
    ElMessage.success('JSON格式正确')
  } catch (error) {
    ElMessage.error('JSON格式错误：' + error.message)
  }
}

// 格式化输入数据
const formatInputData = () => {
  try {
    const parsed = JSON.parse(executionForm.inputData)
    executionForm.inputData = JSON.stringify(parsed, null, 2)
    ElMessage.success('格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

// 加载示例数据
const loadSampleData = () => {
  const sampleData = {
    projectId: 'example-project-001',
    requirements: [
      {
        id: 'req-001',
        name: '用户登录功能',
        description: '实现用户登录、注册、密码重置功能'
      }
    ],
    analysisOptions: {
      includeNonFunctional: true,
      detailLevel: 'comprehensive'
    }
  }
  executionForm.inputData = JSON.stringify(sampleData, null, 2)
}

// 刷新工作流列表
const handleRefresh = () => {
  getWorkflows()
}

// 执行工作流
const handleExecute = async () => {
  if (!selectedWorkflow.value) {
    ElMessage.error('请选择要执行的工作流')
    return
  }

  try {
    // 验证输入数据
    let inputData = {}
    if (executionForm.inputData) {
      try {
        inputData = JSON.parse(executionForm.inputData)
      } catch (error) {
        ElMessage.error('输入参数格式错误，请检查JSON格式')
        return
      }
    }

    executing.value = true
    
    const executionData = {
      name: executionForm.name || `执行-${Date.now()}`,
      priority: executionForm.priority,
      inputData,
      options: executionForm.options
    }
    
    const res = await executeWorkflow(selectedWorkflow.value.workflowId, executionData)
    
    ElMessage.success('工作流执行成功')
    emit('refresh')
    handleClose()
  } catch (error) {
    ElMessage.error('执行失败：' + (error.message || '未知错误'))
  } finally {
    executing.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  selectedWorkflowId.value = ''
  selectedWorkflow.value = null
  executionPreview.value = null
  Object.assign(executionForm, {
    name: '',
    priority: 'normal',
    inputData: '{}',
    options: ['stopOnError']
  })
}

// 工具方法
const getStatusText = (status) => {
  const statusMap = {
    draft: '草稿',
    published: '已发布',
    running: '运行中',
    paused: '已暂停',
    completed: '已完成',
    failed: '失败'
  }
  return statusMap[status] || status
}

const getStatusTagType = (status) => {
  const typeMap = {
    draft: 'info',
    published: 'success',
    running: 'warning',
    paused: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return typeMap[status] || ''
}

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

// 监听对话框显示状态
watch(() => props.visible, (newVal) => {
  if (newVal) {
    getWorkflows()
  }
})

onMounted(() => {
  if (props.visible) {
    getWorkflows()
  }
})
</script>

<style scoped>
.workflow-execution {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.section-card {
  border: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.workflow-option {
  padding: 8px 0;
}

.workflow-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.workflow-desc {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
  line-height: 1.3;
}

.workflow-meta {
  display: flex;
  gap: 8px;
}

.workflow-steps {
  margin-top: 20px;
}

.workflow-steps h4 {
  color: #303133;
  margin-bottom: 16px;
  font-size: 16px;
}

.steps-timeline {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.step-item {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
  background: #fafafa;
}

.step-number {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #409eff;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 500;
  font-size: 14px;
  flex-shrink: 0;
}

.step-content {
  flex: 1;
}

.step-name {
  font-weight: 500;
  color: #303133;
  margin-bottom: 4px;
}

.step-description {
  font-size: 13px;
  color: #606266;
  margin-bottom: 8px;
  line-height: 1.4;
}

.step-meta {
  display: flex;
  gap: 8px;
}

.form-actions {
  margin-top: 8px;
  display: flex;
  gap: 16px;
}

.execution-warning {
  margin-top: 16px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .step-item {
    flex-direction: column;
    align-items: stretch;
  }
  
  .step-number {
    align-self: flex-start;
  }
  
  .step-meta {
    flex-direction: column;
    gap: 4px;
  }
  
  .form-actions {
    flex-direction: column;
    gap: 8px;
  }
}
</style> 