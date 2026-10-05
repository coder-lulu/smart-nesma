<template>
  <el-dialog
    v-model="dialogVisible"
    title="创建工作流"
    width="800px"
    :before-close="handleClose"
    class="create-workflow-dialog"
  >
    <div class="dialog-content">
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
        class="workflow-form"
      >
        <!-- 基本信息 -->
        <el-card class="form-section">
          <template #header>
            <div class="section-header">
              <el-icon><Document /></el-icon>
              <span>基本信息</span>
            </div>
          </template>
          
          <el-form-item label="工作流名称" prop="name">
            <el-input 
              v-model="form.name" 
              placeholder="请输入工作流名称"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>
          
          <el-form-item label="描述">
            <el-input
              v-model="form.description"
              type="textarea"
              :rows="3"
              placeholder="请输入工作流描述"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="版本号" prop="version">
                <el-input 
                  v-model="form.version" 
                  placeholder="例如：1.0.0"
                />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="分类" prop="category">
                <el-select 
                  v-model="form.category" 
                  placeholder="选择工作流分类"
                  style="width: 100%"
                >
                  <el-option 
                    v-for="category in categories" 
                    :key="category.value"
                    :label="category.label" 
                    :value="category.value"
                  />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>
          
          <el-form-item label="标签">
            <el-tag
              v-for="tag in form.tags"
              :key="tag"
              closable
              :disable-transitions="false"
              @close="removeTag(tag)"
              style="margin-right: 10px;"
            >
              {{ tag }}
            </el-tag>
            <el-input
              v-if="tagInputVisible"
              ref="tagInputRef"
              v-model="tagInputValue"
              class="tag-input"
              size="small"
              @keyup.enter="handleTagConfirm"
              @blur="handleTagConfirm"
            />
            <el-button 
              v-else 
              class="button-new-tag" 
              size="small" 
              @click="showTagInput"
            >
              + 新标签
            </el-button>
          </el-form-item>
        </el-card>

        <!-- 工作流配置 -->
        <el-card class="form-section">
          <template #header>
            <div class="section-header">
              <el-icon><Setting /></el-icon>
              <span>工作流配置</span>
            </div>
          </template>
          
          <el-form-item label="创建方式">
            <el-radio-group v-model="createMode" @change="handleCreateModeChange">
              <el-radio value="template">从模板创建</el-radio>
              <el-radio value="blank">空白工作流</el-radio>
              <el-radio value="wizard">向导创建</el-radio>
            </el-radio-group>
          </el-form-item>
          
          <!-- 模板选择 -->
          <el-form-item label="选择模板" v-if="createMode === 'template'">
            <div class="template-grid">
              <div 
                v-for="template in templates" 
                :key="template.id"
                class="template-card"
                :class="{ active: selectedTemplate === template.id }"
                @click="selectedTemplate = template.id"
              >
                <div class="template-icon">
                  <el-icon size="24">
                    <component :is="template.icon" />
                  </el-icon>
                </div>
                <div class="template-info">
                  <div class="template-name">{{ template.name }}</div>
                  <div class="template-desc">{{ template.description }}</div>
                </div>
              </div>
            </div>
          </el-form-item>
          
          <!-- 向导配置 -->
          <div v-if="createMode === 'wizard'" class="wizard-config">
            <el-form-item label="业务场景">
              <el-select 
                v-model="wizardConfig.scenario" 
                placeholder="选择业务场景"
                style="width: 100%"
              >
                <el-option label="需求分析流程" value="requirement_analysis" />
                <el-option label="NESMA评估流程" value="nesma_evaluation" />
                <el-option label="文档生成流程" value="document_generation" />
                <el-option label="质量检查流程" value="quality_check" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="复杂度">
              <el-radio-group v-model="wizardConfig.complexity">
                <el-radio value="simple">简单（3-5个步骤）</el-radio>
                <el-radio value="medium">中等（5-8个步骤）</el-radio>
                <el-radio value="complex">复杂（8+个步骤）</el-radio>
              </el-radio-group>
            </el-form-item>
            
            <el-form-item label="包含步骤">
              <el-checkbox-group v-model="wizardConfig.steps">
                <el-checkbox value="requirement_analysis">需求分析</el-checkbox>
                <el-checkbox value="nesma_evaluation">NESMA评估</el-checkbox>
                <el-checkbox value="knowledge_retrieval">知识检索</el-checkbox>
                <el-checkbox value="document_generation">文档生成</el-checkbox>
                <el-checkbox value="quality_check">质量检查</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </div>
          
          <el-form-item label="是否为模板">
            <el-switch 
              v-model="form.isTemplate" 
              active-text="是" 
              inactive-text="否"
            />
            <span class="form-tip">设为模板后可被其他人使用</span>
          </el-form-item>
        </el-card>

        <!-- 高级设置 -->
        <el-card class="form-section">
          <template #header>
            <div class="section-header">
              <el-icon><Tools /></el-icon>
              <span>高级设置</span>
              <el-switch 
                v-model="showAdvanced" 
                active-text="显示高级设置"
                size="small"
                style="margin-left: auto;"
              />
            </div>
          </template>
          
          <div v-show="showAdvanced">
            <el-form-item label="执行超时">
              <el-input-number 
                v-model="form.timeout" 
                :min="60" 
                :max="7200"
                style="width: 200px"
              />
              <span class="form-tip">秒（默认3600秒）</span>
            </el-form-item>
            
            <el-form-item label="最大重试次数">
              <el-input-number 
                v-model="form.maxRetries" 
                :min="0" 
                :max="10"
                style="width: 200px"
              />
            </el-form-item>
            
            <el-form-item label="并发策略">
              <el-select 
                v-model="form.concurrencyStrategy" 
                style="width: 200px"
              >
                <el-option label="串行执行" value="sequential" />
                <el-option label="并行执行" value="parallel" />
                <el-option label="混合执行" value="mixed" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="配置参数">
              <el-input
                v-model="form.configuration"
                type="textarea"
                :rows="6"
                placeholder="JSON格式的配置参数"
              />
              <div class="config-actions">
                <el-button 
                  type="text" 
                  size="small" 
                  @click="formatConfig"
                >
                  格式化JSON
                </el-button>
                <el-button 
                  type="text" 
                  size="small" 
                  @click="validateConfig"
                >
                  验证格式
                </el-button>
                <el-button 
                  type="text" 
                  size="small" 
                  @click="loadDefaultConfig"
                >
                  加载默认配置
                </el-button>
              </div>
            </el-form-item>
          </div>
        </el-card>
      </el-form>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button 
          type="primary" 
          @click="handleSubmit" 
          :loading="loading"
        >
          创建工作流
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Document, Setting, Tools, Operation, CaretRight, Connection, Edit, Clock } from '@element-plus/icons-vue'
import { createWorkflow } from '@/api/nesma'

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

const formRef = ref(null)
const tagInputRef = ref(null)
const loading = ref(false)
const showAdvanced = ref(false)
const tagInputVisible = ref(false)
const tagInputValue = ref('')
const createMode = ref('template')
const selectedTemplate = ref('')

// 向导配置
const wizardConfig = reactive({
  scenario: '',
  complexity: 'medium',
  steps: []
})

// 分类选项
const categories = [
  { label: '需求分析', value: 'requirement_analysis' },
  { label: 'NESMA评估', value: 'nesma_evaluation' },
  { label: '文档生成', value: 'document_generation' },
  { label: '质量检查', value: 'quality_check' },
  { label: '其他', value: 'other' }
]

// 模板选项
const templates = [
  {
    id: 'requirement_analysis_basic',
    name: '基础需求分析',
    description: '简单的需求分析流程模板',
    icon: 'Document'
  },
  {
    id: 'nesma_evaluation_standard',
    name: '标准NESMA评估',
    description: '标准的NESMA功能点评估流程',
    icon: 'Operation'
  },
  {
    id: 'document_generation_auto',
    name: '自动文档生成',
    description: '自动化文档生成流程模板',
    icon: 'Edit'
  },
  {
    id: 'quality_check_comprehensive',
    name: '全面质量检查',
    description: '综合性质量检查流程模板',
    icon: 'VideoPlay'
  }
]

// 表单数据
const form = reactive({
  name: '',
  description: '',
  version: '1.0.0',
  category: '',
  tags: [],
  isTemplate: false,
  timeout: 3600,
  maxRetries: 3,
  concurrencyStrategy: 'sequential',
  configuration: ''
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入工作流名称', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在2到100个字符', trigger: 'blur' }
  ],
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' },
    { pattern: /^\d+\.\d+\.\d+$/, message: '版本号格式错误，应为x.x.x', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择分类', trigger: 'change' }
  ]
}

// 处理创建方式变化
const handleCreateModeChange = (mode) => {
  if (mode === 'template') {
    selectedTemplate.value = templates[0]?.id
  } else {
    selectedTemplate.value = ''
  }
  
  if (mode === 'wizard') {
    wizardConfig.scenario = ''
    wizardConfig.complexity = 'medium'
    wizardConfig.steps = []
  }
}

// 显示标签输入框
const showTagInput = () => {
  tagInputVisible.value = true
  nextTick(() => {
    tagInputRef.value?.focus()
  })
}

// 确认添加标签
const handleTagConfirm = () => {
  if (tagInputValue.value && !form.tags.includes(tagInputValue.value)) {
    form.tags.push(tagInputValue.value)
  }
  tagInputVisible.value = false
  tagInputValue.value = ''
}

// 移除标签
const removeTag = (tag) => {
  const index = form.tags.indexOf(tag)
  if (index > -1) {
    form.tags.splice(index, 1)
  }
}

// 格式化配置JSON
const formatConfig = () => {
  try {
    if (form.configuration) {
      const parsed = JSON.parse(form.configuration)
      form.configuration = JSON.stringify(parsed, null, 2)
      ElMessage.success('JSON格式化成功')
    }
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

// 验证配置JSON
const validateConfig = () => {
  try {
    if (form.configuration) {
      JSON.parse(form.configuration)
      ElMessage.success('JSON格式正确')
    } else {
      ElMessage.info('配置为空')
    }
  } catch (error) {
    ElMessage.error('JSON格式错误：' + error.message)
  }
}

// 加载默认配置
const loadDefaultConfig = () => {
  const defaultConfig = {
    timeout: 3600,
    maxRetries: 3,
    enableLogging: true,
    autoSave: true,
    notifications: {
      onSuccess: true,
      onFailure: true
    }
  }
  form.configuration = JSON.stringify(defaultConfig, null, 2)
}

// 重置表单
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(form, {
    name: '',
    description: '',
    version: '1.0.0',
    category: '',
    tags: [],
    isTemplate: false,
    timeout: 3600,
    maxRetries: 3,
    concurrencyStrategy: 'sequential',
    configuration: ''
  })
  createMode.value = 'template'
  selectedTemplate.value = templates[0]?.id
  showAdvanced.value = false
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    
    loading.value = true
    
    // 根据创建方式生成工作流定义
    let definition = {}
    
    if (createMode.value === 'template' && selectedTemplate.value) {
      definition = generateFromTemplate(selectedTemplate.value)
    } else if (createMode.value === 'wizard') {
      definition = generateFromWizard(wizardConfig)
    } else {
      definition = { steps: [] }
    }
    
    // 验证配置参数
    let configuration = null
    if (form.configuration) {
      try {
        configuration = JSON.parse(form.configuration)
      } catch (error) {
        ElMessage.error('配置参数格式错误，请检查JSON格式')
        return
      }
    }
    
    const requestData = {
      name: form.name,
      description: form.description,
      version: form.version,
      category: form.category,
      tags: form.tags,
      isTemplate: form.isTemplate,
      definition: JSON.stringify(definition),
      configuration,
      timeout: form.timeout,
      maxRetries: form.maxRetries,
      concurrencyStrategy: form.concurrencyStrategy
    }
    
    await createWorkflow(requestData)
    
    ElMessage.success('工作流创建成功')
    emit('refresh')
    handleClose()
  } catch (error) {
    ElMessage.error('创建失败：' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 从模板生成定义
const generateFromTemplate = (templateId) => {
  const templates = {
    requirement_analysis_basic: {
      steps: [
        {
          id: 'start',
          name: '开始',
          agentType: 'START',
          order: 1,
          dependencies: []
        },
        {
          id: 'analyze',
          name: '需求分析',
          agentType: 'REQUIREMENT_ANALYSIS',
          order: 2,
          dependencies: ['start']
        },
        {
          id: 'end',
          name: '结束',
          agentType: 'END',
          order: 3,
          dependencies: ['analyze']
        }
      ]
    },
    nesma_evaluation_standard: {
      steps: [
        {
          id: 'start',
          name: '开始',
          agentType: 'START',
          order: 1,
          dependencies: []
        },
        {
          id: 'requirement',
          name: '需求分析',
          agentType: 'REQUIREMENT_ANALYSIS',
          order: 2,
          dependencies: ['start']
        },
        {
          id: 'nesma',
          name: 'NESMA评估',
          agentType: 'NESMA_EVALUATION',
          order: 3,
          dependencies: ['requirement']
        },
        {
          id: 'document',
          name: '生成报告',
          agentType: 'DOCUMENT_GENERATION',
          order: 4,
          dependencies: ['nesma']
        },
        {
          id: 'end',
          name: '结束',
          agentType: 'END',
          order: 5,
          dependencies: ['document']
        }
      ]
    }
  }
  
  return templates[templateId] || { steps: [] }
}

// 从向导生成定义
const generateFromWizard = (config) => {
  const steps = [
    {
      id: 'start',
      name: '开始',
      agentType: 'START',
      order: 1,
      dependencies: []
    }
  ]
  
  let order = 2
  let lastStepId = 'start'
  
  config.steps.forEach(stepType => {
    const stepId = `step_${order}`
    steps.push({
      id: stepId,
      name: getStepName(stepType),
      agentType: stepType.toUpperCase(),
      order,
      dependencies: [lastStepId]
    })
    lastStepId = stepId
    order++
  })
  
  steps.push({
    id: 'end',
    name: '结束',
    agentType: 'END',
    order,
    dependencies: [lastStepId]
  })
  
  return { steps }
}

const getStepName = (stepType) => {
  const names = {
    requirement_analysis: '需求分析',
    nesma_evaluation: 'NESMA评估',
    knowledge_retrieval: '知识检索',
    document_generation: '文档生成',
    quality_check: '质量检查'
  }
  return names[stepType] || stepType
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  handleReset()
}
</script>

<style scoped>
.create-workflow-dialog {
  --el-dialog-border-radius: 16px;
}

.dialog-content {
  max-height: 70vh;
  overflow-y: auto;
  padding: 0 4px;
}

.workflow-form {
  gap: 24px;
  display: flex;
  flex-direction: column;
}

.form-section {
  border: none;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border-radius: 12px;
  margin-bottom: 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  color: #374151;
}

.form-tip {
  font-size: 12px;
  color: #6b7280;
  margin-left: 8px;
}

.tag-input {
  width: 90px;
  margin-left: 10px;
  vertical-align: bottom;
}

.button-new-tag {
  margin-left: 10px;
  height: 32px;
  line-height: 30px;
  padding-top: 0;
  padding-bottom: 0;
}

.template-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
  gap: 16px;
  margin-top: 12px;
}

.template-card {
  border: 2px solid #e5e7eb;
  border-radius: 12px;
  padding: 20px;
  cursor: pointer;
  transition: all 0.3s ease;
  background: #fafbfc;
}

.template-card:hover {
  border-color: #3b82f6;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.15);
  transform: translateY(-2px);
}

.template-card.active {
  border-color: #3b82f6;
  background: #eff6ff;
  box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
}

.template-icon {
  width: 48px;
  height: 48px;
  border-radius: 12px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  margin-bottom: 12px;
}

.template-name {
  font-weight: 600;
  color: #1f2937;
  margin-bottom: 4px;
}

.template-desc {
  font-size: 13px;
  color: #6b7280;
  line-height: 1.4;
}

.wizard-config {
  padding: 16px;
  background: #f8fafc;
  border-radius: 8px;
  margin-top: 12px;
}

.config-actions {
  margin-top: 8px;
  display: flex;
  gap: 16px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 20px 0 0 0;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .template-grid {
    grid-template-columns: 1fr;
  }
  
  .dialog-content {
    max-height: 60vh;
  }
  
  .config-actions {
    flex-direction: column;
    gap: 8px;
  }
}

/* 表单项样式增强 */
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #374151;
}

:deep(.el-input__wrapper) {
  border-radius: 8px;
}

:deep(.el-textarea__inner) {
  border-radius: 8px;
}

:deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}

:deep(.el-card__header) {
  padding: 20px 24px;
  background: #f8fafc;
  border-bottom: 1px solid #e5e7eb;
}

:deep(.el-card__body) {
  padding: 24px;
}
</style> 