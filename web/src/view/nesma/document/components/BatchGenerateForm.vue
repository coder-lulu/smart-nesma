<template>
  <div class="batch-generate-form">
    <el-steps :active="currentStep" align-center>
      <el-step title="选择项目" description="选择要生成文档的项目" />
      <el-step title="配置文档" description="配置文档类型和模板" />
      <el-step title="确认生成" description="确认配置并开始生成" />
    </el-steps>

    <!-- 步骤1: 选择项目 -->
    <div v-if="currentStep === 0" class="step-content">
      <el-card class="step-card">
        <template #header>
          <div class="step-header">
            <el-icon><FolderOpened /></el-icon>
            <span>选择项目</span>
          </div>
        </template>

        <div class="project-selection">
          <div class="selection-toolbar">
            <el-input
              v-model="projectSearch"
              placeholder="搜索项目名称"
              style="width: 300px;"
              clearable
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
            <el-button @click="selectAllProjects">全选</el-button>
            <el-button @click="clearAllProjects">清空</el-button>
          </div>

          <el-table
            ref="projectTableRef"
            v-loading="projectLoading"
            :data="filteredProjects"
            style="width: 100%; margin-top: 16px;"
            @selection-change="handleProjectSelectionChange"
            max-height="400"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="name" label="项目名称" min-width="200" />
            <el-table-column prop="description" label="项目描述" min-width="250" show-overflow-tooltip />
            <el-table-column prop="domain" label="领域" width="120">
              <template #default="scope">
                <el-tag size="small">{{ scope.row.domain }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="需求统计" width="150">
              <template #default="scope">
                <div class="requirement-count">
                  <span>{{ scope.row.requirementCount || 0 }} 个需求</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag :type="getProjectStatusType(scope.row.status)" size="small">
                  {{ getProjectStatusLabel(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>

          <div class="selection-summary">
            <el-alert
              :title="`已选择 ${selectedProjects.length} 个项目`"
              type="info"
              :closable="false"
              show-icon
            >
              <template #default>
                <div v-if="selectedProjects.length > 0">
                  <div class="selected-projects">
                    <el-tag
                      v-for="project in selectedProjects.slice(0, 5)"
                      :key="project.ID"
                      size="small"
                      style="margin-right: 8px; margin-bottom: 4px;"
                    >
                      {{ project.name }}
                    </el-tag>
                    <span v-if="selectedProjects.length > 5">
                      等 {{ selectedProjects.length }} 个项目
                    </span>
                  </div>
                </div>
              </template>
            </el-alert>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 步骤2: 配置文档 -->
    <div v-if="currentStep === 1" class="step-content">
      <el-card class="step-card">
        <template #header>
          <div class="step-header">
            <el-icon><Setting /></el-icon>
            <span>配置文档</span>
          </div>
        </template>

        <el-form
          ref="configFormRef"
          :model="batchConfig"
          :rules="configRules"
          label-width="120px"
        >
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="文档类型" prop="types">
                <el-checkbox-group v-model="batchConfig.types">
                  <el-checkbox value="word">Word文档</el-checkbox>
                  <el-checkbox value="excel">Excel表格</el-checkbox>
                  <el-checkbox value="pdf">PDF文档</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="文档格式" prop="formats">
                <el-checkbox-group v-model="batchConfig.formats">
                  <el-checkbox value="requirement_spec">需求规格说明书</el-checkbox>
                  <el-checkbox value="nesma_report">NESMA评估报告</el-checkbox>
                  <el-checkbox value="business_summary">业务需求汇总表</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="命名规则" prop="namingRule">
            <el-input
              v-model="batchConfig.namingRule"
              placeholder="如：{项目名称}_{文档格式}_{日期}"
            />
            <div class="naming-hint">
              <small>可用变量：{项目名称}、{文档格式}、{文档类型}、{日期}、{时间}</small>
            </div>
          </el-form-item>

          <el-form-item label="版本号" prop="version">
            <el-input
              v-model="batchConfig.version"
              placeholder="如：v1.0.0"
              style="width: 200px;"
            />
          </el-form-item>

          <el-form-item label="生成配置">
            <el-card class="config-section">
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-form-item label="包含图表" label-width="80px">
                    <el-switch v-model="batchConfig.config.includeCharts" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="包含统计" label-width="80px">
                    <el-switch v-model="batchConfig.config.includeStatistics" />
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="详细模式" label-width="80px">
                    <el-switch v-model="batchConfig.config.detailedMode" />
                  </el-form-item>
                </el-col>
              </el-row>
            </el-card>
          </el-form-item>

          <el-form-item label="生成策略">
            <el-radio-group v-model="batchConfig.strategy">
              <el-radio value="parallel">
                <div class="strategy-option">
                  <div class="strategy-title">并行生成</div>
                  <div class="strategy-desc">同时生成多个文档，速度快但消耗资源多</div>
                </div>
              </el-radio>
              <el-radio value="sequential">
                <div class="strategy-option">
                  <div class="strategy-title">顺序生成</div>
                  <div class="strategy-desc">依次生成文档，速度慢但资源消耗少</div>
                </div>
              </el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="并发数量" v-if="batchConfig.strategy === 'parallel'">
            <el-slider
              v-model="batchConfig.concurrency"
              :min="1"
              :max="10"
              :step="1"
              show-stops
              show-input
              style="width: 300px;"
            />
            <div class="concurrency-hint">
              <small>建议并发数量不超过5个，避免服务器负载过重</small>
            </div>
          </el-form-item>
        </el-form>
      </el-card>
    </div>

    <!-- 步骤3: 确认生成 -->
    <div v-if="currentStep === 2" class="step-content">
      <el-card class="step-card">
        <template #header>
          <div class="step-header">
            <el-icon><Check /></el-icon>
            <span>确认生成</span>
          </div>
        </template>

        <div class="generation-summary">
          <el-descriptions title="生成概要" :column="2" border>
            <el-descriptions-item label="选择项目">
              {{ selectedProjects.length }} 个
            </el-descriptions-item>
            <el-descriptions-item label="文档类型">
              {{ batchConfig.types.join(', ') }}
            </el-descriptions-item>
            <el-descriptions-item label="文档格式">
              {{ batchConfig.formats.map(f => getFormatLabel(f)).join(', ') }}
            </el-descriptions-item>
            <el-descriptions-item label="生成策略">
              {{ batchConfig.strategy === 'parallel' ? '并行生成' : '顺序生成' }}
            </el-descriptions-item>
            <el-descriptions-item label="预计文档数">
              {{ estimatedDocumentCount }} 个
            </el-descriptions-item>
            <el-descriptions-item label="预计时间">
              {{ estimatedTime }}
            </el-descriptions-item>
          </el-descriptions>

          <el-alert
            title="注意事项"
            type="warning"
            :closable="false"
            style="margin-top: 20px;"
          >
            <ul>
              <li>批量生成可能需要较长时间，请耐心等待</li>
              <li>生成过程中请不要关闭浏览器</li>
              <li>文档生成完成后会自动通知</li>
              <li>可在文档列表中查看生成进度</li>
            </ul>
          </el-alert>

          <div class="document-preview" v-if="estimatedDocumentCount > 0">
            <h4>将要生成的文档：</h4>
            <el-table
              :data="documentPreview"
              style="width: 100%; margin-top: 10px;"
              max-height="300"
            >
              <el-table-column prop="projectName" label="项目" width="200" />
              <el-table-column prop="type" label="类型" width="100">
                <template #default="scope">
                  <el-tag size="small">{{ scope.row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="format" label="格式" width="150">
                <template #default="scope">
                  <el-tag size="small" type="info">{{ getFormatLabel(scope.row.format) }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="name" label="文档名称" min-width="250" />
            </el-table>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 操作按钮 -->
    <div class="step-actions">
      <el-button v-if="currentStep > 0" @click="handlePrevStep">上一步</el-button>
      <el-button @click="handleCancel">取消</el-button>
      <el-button
        v-if="currentStep < 2"
        type="primary"
        @click="handleNextStep"
        :disabled="!canNextStep"
      >
        下一步
      </el-button>
      <el-button
        v-if="currentStep === 2"
        type="success"
        @click="handleStartGeneration"
        :loading="generating"
      >
        开始生成
      </el-button>
    </div>

    <!-- 生成进度对话框 -->
    <el-dialog
      v-model="progressDialogVisible"
      title="批量生成进度"
      width="800px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="generation-progress">
        <div class="overall-progress">
          <div class="progress-header">
            <span>总体进度</span>
            <span>{{ completedCount }}/{{ totalCount }}</span>
          </div>
          <el-progress
            :percentage="overallProgress"
            :stroke-width="10"
            :text-inside="true"
          />
        </div>

        <div class="detail-progress">
          <el-table
            :data="generationTasks"
            style="width: 100%; margin-top: 20px;"
            max-height="300"
          >
            <el-table-column prop="documentName" label="文档名称" min-width="200" />
            <el-table-column prop="status" label="状态" width="100">
              <template #default="scope">
                <el-tag :type="getTaskStatusType(scope.row.status)" size="small">
                  {{ getTaskStatusLabel(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="progress" label="进度" width="150">
              <template #default="scope">
                <el-progress
                  :percentage="scope.row.progress"
                  :stroke-width="6"
                  :show-text="false"
                />
              </template>
            </el-table-column>
            <el-table-column prop="error" label="错误信息" min-width="200">
              <template #default="scope">
                <span v-if="scope.row.error" class="error-text">{{ scope.row.error }}</span>
                <span v-else>-</span>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>

      <template #footer>
        <el-button @click="handleCloseProgress" :disabled="generating">
          {{ generating ? '生成中...' : '关闭' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  FolderOpened,
  Setting,
  Check,
  Search
} from '@element-plus/icons-vue'
import {
  getNesmaProjectList,
  batchGenerateDocument
} from '@/api/nesma'

// Emits
const emit = defineEmits(['submit', 'cancel'])

// 响应式数据
const currentStep = ref(0)
const projectLoading = ref(false)
const generating = ref(false)
const progressDialogVisible = ref(false)

const projectTableRef = ref(null)
const configFormRef = ref(null)

const projectSearch = ref('')
const allProjects = ref([])
const selectedProjects = ref([])

// 批量配置
const batchConfig = reactive({
  types: ['word'],
  formats: ['requirement_spec'],
  namingRule: '{项目名称}_{文档格式}_{日期}',
  version: 'v1.0.0',
  strategy: 'parallel',
  concurrency: 3,
  config: {
    includeCharts: true,
    includeStatistics: true,
    detailedMode: false
  }
})

// 生成任务
const generationTasks = ref([])
const completedCount = ref(0)
const totalCount = ref(0)

// 表单验证规则
const configRules = {
  types: [
    { required: true, message: '请选择至少一种文档类型', trigger: 'change' }
  ],
  formats: [
    { required: true, message: '请选择至少一种文档格式', trigger: 'change' }
  ],
  namingRule: [
    { required: true, message: '请输入命名规则', trigger: 'blur' }
  ],
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' }
  ]
}

// 计算属性
const filteredProjects = computed(() => {
  if (!projectSearch.value) return allProjects.value
  return allProjects.value.filter(project =>
    project.name.toLowerCase().includes(projectSearch.value.toLowerCase())
  )
})

const canNextStep = computed(() => {
  if (currentStep.value === 0) {
    return selectedProjects.value.length > 0
  }
  if (currentStep.value === 1) {
    return batchConfig.types.length > 0 && batchConfig.formats.length > 0
  }
  return true
})

const estimatedDocumentCount = computed(() => {
  return selectedProjects.value.length * batchConfig.types.length * batchConfig.formats.length
})

const estimatedTime = computed(() => {
  const count = estimatedDocumentCount.value
  const timePerDoc = batchConfig.strategy === 'parallel' ? 30 : 60 // 秒
  const totalSeconds = Math.ceil(count * timePerDoc / (batchConfig.concurrency || 1))
  
  if (totalSeconds < 60) return `约 ${totalSeconds} 秒`
  if (totalSeconds < 3600) return `约 ${Math.ceil(totalSeconds / 60)} 分钟`
  return `约 ${Math.ceil(totalSeconds / 3600)} 小时`
})

const documentPreview = computed(() => {
  const preview = []
  selectedProjects.value.forEach(project => {
    batchConfig.types.forEach(type => {
      batchConfig.formats.forEach(format => {
        preview.push({
          projectName: project.name,
          type: type,
          format: format,
          name: generateDocumentName(project.name, format, type)
        })
      })
    })
  })
  return preview.slice(0, 50) // 最多显示50个
})

const overallProgress = computed(() => {
  if (totalCount.value === 0) return 0
  return Math.round((completedCount.value / totalCount.value) * 100)
})

// 生命周期
onMounted(() => {
  loadProjects()
})

// 加载项目列表
const loadProjects = async () => {
  projectLoading.value = true
  try {
    const res = await getNesmaProjectList({ 
      pageSize: 1000,
      status: 'active'
    })
    if (res.code === 0) {
      allProjects.value = res.data.list || []
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
    ElMessage.error('加载项目列表失败')
  } finally {
    projectLoading.value = false
  }
}

// 事件处理
const handleProjectSelectionChange = (selection) => {
  selectedProjects.value = selection
}

const selectAllProjects = () => {
  projectTableRef.value?.toggleAllSelection()
}

const clearAllProjects = () => {
  projectTableRef.value?.clearSelection()
}

const handlePrevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const handleNextStep = async () => {
  if (currentStep.value === 1) {
    // 验证配置表单
    const valid = await configFormRef.value?.validate().catch(() => false)
    if (!valid) return
  }
  
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

const handleCancel = () => {
  emit('cancel')
}

const handleStartGeneration = async () => {
  try {
    await ElMessageBox.confirm(
      `确定要生成 ${estimatedDocumentCount.value} 个文档吗？`,
      '批量生成确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 准备生成任务
    const tasks = []
    selectedProjects.value.forEach(project => {
      batchConfig.types.forEach(type => {
        batchConfig.formats.forEach(format => {
          tasks.push({
            projectId: project.ID,
            projectName: project.name,
            type: type,
            format: format,
            name: generateDocumentName(project.name, format, type),
            templateId: null, // 使用默认模板
            config: batchConfig.config
          })
        })
      })
    })

    // 开始批量生成
    generating.value = true
    progressDialogVisible.value = true
    generationTasks.value = tasks.map(task => ({
      ...task,
      status: 'pending',
      progress: 0,
      error: null
    }))
    totalCount.value = tasks.length
    completedCount.value = 0

    const res = await batchGenerateDocument({
      tasks: tasks,
      strategy: batchConfig.strategy,
      concurrency: batchConfig.concurrency
    })

    if (res.code === 0) {
      ElMessage.success('批量生成任务已提交')
      // 这里可以添加轮询逻辑来更新进度
      simulateProgress()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量生成失败:', error)
      ElMessage.error('批量生成失败')
    }
  }
}

// 模拟进度更新（实际项目中应该通过WebSocket或轮询获取真实进度）
const simulateProgress = () => {
  const interval = setInterval(() => {
    generationTasks.value.forEach(task => {
      if (task.status === 'pending') {
        task.status = 'generating'
        task.progress = 10
      } else if (task.status === 'generating') {
        task.progress = Math.min(task.progress + Math.random() * 20, 100)
        if (task.progress >= 100) {
          task.status = 'completed'
          completedCount.value++
        }
      }
    })

    if (completedCount.value >= totalCount.value) {
      clearInterval(interval)
      generating.value = false
      ElMessage.success('批量生成完成')
    }
  }, 1000)
}

const handleCloseProgress = () => {
  if (!generating.value) {
    progressDialogVisible.value = false
    emit('submit')
  }
}

// 工具函数
const generateDocumentName = (projectName, format, type) => {
  const formatLabels = {
    requirement_spec: '需求规格说明书',
    nesma_report: 'NESMA评估报告',
    business_summary: '业务需求汇总表'
  }
  
  let name = batchConfig.namingRule
  name = name.replace('{项目名称}', projectName)
  name = name.replace('{文档格式}', formatLabels[format] || format)
  name = name.replace('{文档类型}', type.toUpperCase())
  name = name.replace('{日期}', new Date().toISOString().split('T')[0])
  name = name.replace('{时间}', new Date().toTimeString().split(' ')[0])
  
  return name
}

const getFormatLabel = (format) => {
  const labels = {
    requirement_spec: '需求规格说明书',
    nesma_report: 'NESMA评估报告',
    business_summary: '业务需求汇总表'
  }
  return labels[format] || format
}

const getProjectStatusLabel = (status) => {
  const labels = {
    active: '活跃',
    paused: '暂停',
    completed: '完成',
    archived: '归档'
  }
  return labels[status] || status
}

const getProjectStatusType = (status) => {
  const types = {
    active: 'success',
    paused: 'warning',
    completed: 'info',
    archived: ''
  }
  return types[status] || ''
}

const getTaskStatusLabel = (status) => {
  const labels = {
    pending: '等待中',
    generating: '生成中',
    completed: '已完成',
    failed: '失败'
  }
  return labels[status] || status
}

const getTaskStatusType = (status) => {
  const types = {
    pending: 'info',
    generating: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return types[status] || ''
}
</script>

<style scoped>
.batch-generate-form {
  padding: 20px;
}

.step-content {
  margin: 30px 0;
}

.step-card {
  border-radius: 8px;
}

.step-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
}

.project-selection {
  margin-top: 20px;
}

.selection-toolbar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.selection-summary {
  margin-top: 16px;
}

.selected-projects {
  margin-top: 8px;
}

.config-section {
  background: var(--el-fill-color-light);
  border: none;
}

.naming-hint {
  margin-top: 4px;
  color: var(--el-text-color-secondary);
}

.strategy-option {
  margin-left: 8px;
}

.strategy-title {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.strategy-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.concurrency-hint {
  margin-top: 8px;
  color: var(--el-text-color-secondary);
}

.generation-summary {
  margin-top: 20px;
}

.document-preview {
  margin-top: 20px;
}

.document-preview h4 {
  margin: 0 0 10px 0;
  color: var(--el-text-color-primary);
}

.step-actions {
  display: flex;
  justify-content: center;
  gap: 12px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color);
}

.generation-progress {
  padding: 10px 0;
}

.overall-progress {
  margin-bottom: 20px;
}

.progress-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 10px;
  font-weight: 500;
}

.detail-progress {
  margin-top: 20px;
}

.error-text {
  color: var(--el-color-danger);
  font-size: 12px;
}

.requirement-count {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style> 