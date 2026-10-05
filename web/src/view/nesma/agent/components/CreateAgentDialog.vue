<template>
  <el-dialog
    v-model="dialogVisible"
    title="注册Agent"
    width="700px"
    :before-close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      class="agent-form"
    >
      <!-- 基本信息 -->
      <el-card class="form-section">
        <template #header>
          <span>基本信息</span>
        </template>
        
        <el-form-item label="Agent名称" prop="name">
          <el-input 
            v-model="form.name" 
            placeholder="请输入Agent名称"
            maxlength="50"
            show-word-limit
          />
        </el-form-item>
        
        <el-form-item label="Agent描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入Agent描述"
            maxlength="500"
            show-word-limit
          />
        </el-form-item>
        
        <el-form-item label="Agent类型" prop="agentType">
          <el-select 
            v-model="form.agentType" 
            placeholder="选择Agent类型" 
            style="width: 100%"
            @change="handleTypeChange"
          >
            <el-option 
              v-for="type in agentTypes" 
              :key="type.value"
              :label="type.label" 
              :value="type.value"
            >
              <div class="type-option">
                <span class="type-label">{{ type.label }}</span>
                <span class="type-desc">{{ type.description }}</span>
              </div>
            </el-option>
          </el-select>
        </el-form-item>
        
        <el-form-item label="版本号" prop="version">
          <el-input 
            v-model="form.version" 
            placeholder="例如：1.0.0"
            style="width: 200px;"
          />
        </el-form-item>
      </el-card>

      <!-- 运行配置 -->
      <el-card class="form-section">
        <template #header>
          <span>运行配置</span>
        </template>
        
        <el-form-item label="最大并发数" prop="maxConcurrency">
          <el-input-number 
            v-model="form.maxConcurrency" 
            :min="1" 
            :max="20"
            style="width: 200px"
          />
          <span class="form-tip">设置Agent同时处理的最大任务数</span>
        </el-form-item>
        
        <el-form-item label="端点地址" prop="endpoint">
          <el-input 
            v-model="form.endpoint" 
            placeholder="例如：http://localhost:8080/api"
          />
          <span class="form-tip">Agent的HTTP服务端点，留空则使用默认配置</span>
        </el-form-item>
        
        <el-form-item label="超时时间" prop="timeout">
          <el-input-number 
            v-model="form.timeout" 
            :min="5" 
            :max="300"
            style="width: 200px"
          />
          <span class="form-tip">任务超时时间（秒）</span>
        </el-form-item>
        
        <el-form-item label="重试次数" prop="retries">
          <el-input-number 
            v-model="form.retries" 
            :min="0" 
            :max="5"
            style="width: 200px"
          />
          <span class="form-tip">任务失败时的重试次数</span>
        </el-form-item>
      </el-card>

      <!-- 高级配置 -->
      <el-card class="form-section">
        <template #header>
          <div class="section-header">
            <span>高级配置</span>
            <el-switch 
              v-model="showAdvanced" 
              active-text="显示高级配置"
              size="small"
            />
          </div>
        </template>
        
        <div v-show="showAdvanced">
          <el-form-item label="环境变量">
            <div class="env-vars">
              <div 
                v-for="(env, index) in form.envVars" 
                :key="index"
                class="env-var-item"
              >
                <el-input 
                  v-model="env.key" 
                  placeholder="变量名"
                  style="width: 150px; margin-right: 10px;"
                />
                <el-input 
                  v-model="env.value" 
                  placeholder="变量值"
                  style="width: 200px; margin-right: 10px;"
                />
                <el-button 
                  type="danger" 
                  size="small" 
                  icon="Delete"
                  @click="removeEnvVar(index)"
                />
              </div>
              <el-button 
                type="primary" 
                size="small" 
                icon="Plus"
                @click="addEnvVar"
              >
                添加环境变量
              </el-button>
            </div>
          </el-form-item>
          
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
          
          <el-form-item label="配置参数">
            <el-input
              v-model="form.configParams"
              type="textarea"
              :rows="6"
              placeholder="JSON格式的配置参数，例如：&#10;{&#10;  &quot;key1&quot;: &quot;value1&quot;,&#10;  &quot;key2&quot;: &quot;value2&quot;&#10;}"
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
                @click="loadTemplate"
              >
                加载模板
              </el-button>
            </div>
          </el-form-item>
        </div>
      </el-card>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button 
          type="primary" 
          @click="handleSubmit" 
          :loading="loading"
        >
          注册Agent
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { registerAgent } from '@/api/nesma'

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

// Agent类型选项
const agentTypes = [
  {
    value: 'REQUIREMENT_ANALYSIS',
    label: '需求分析Agent',
    description: '负责分析和处理软件需求'
  },
  {
    value: 'NESMA_EVALUATION',
    label: 'NESMA评估Agent',
    description: '执行NESMA功能点分析和评估'
  },
  {
    value: 'PROJECT_MANAGEMENT',
    label: '项目管理Agent',
    description: '协助项目管理和进度跟踪'
  },
  {
    value: 'KNOWLEDGE_MANAGEMENT',
    label: '知识管理Agent',
    description: '管理和维护知识库'
  },
  {
    value: 'QUALITY_ASSURANCE',
    label: '质量保证Agent',
    description: '执行质量检查和验证'
  }
]

// 表单数据
const form = reactive({
  name: '',
  description: '',
  agentType: '',
  version: '1.0.0',
  maxConcurrency: 3,
  endpoint: '',
  timeout: 30,
  retries: 3,
  envVars: [],
  tags: [],
  configParams: ''
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入Agent名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在2到50个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入Agent描述', trigger: 'blur' },
    { max: 500, message: '描述不能超过500个字符', trigger: 'blur' }
  ],
  agentType: [
    { required: true, message: '请选择Agent类型', trigger: 'change' }
  ],
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' },
    { pattern: /^\d+\.\d+\.\d+$/, message: '版本号格式错误，应为x.x.x', trigger: 'blur' }
  ],
  maxConcurrency: [
    { required: true, message: '请设置最大并发数', trigger: 'change' }
  ],
  timeout: [
    { required: true, message: '请设置超时时间', trigger: 'change' }
  ],
  retries: [
    { required: true, message: '请设置重试次数', trigger: 'change' }
  ],
  endpoint: [
    { 
      pattern: /^https?:\/\/.+/, 
      message: '端点地址格式错误', 
      trigger: 'blur' 
    }
  ]
}

// 处理Agent类型变化
const handleTypeChange = (type) => {
  // 根据类型设置默认配置
  const typeConfigs = {
    'REQUIREMENT_ANALYSIS': {
      timeout: 60,
      maxConcurrency: 3,
      configParams: JSON.stringify({
        analysisDepth: 'detailed',
        includeNonFunctional: true
      }, null, 2)
    },
    'NESMA_EVALUATION': {
      timeout: 120,
      maxConcurrency: 2,
      configParams: JSON.stringify({
        evaluationMethod: 'standard',
        includeDataFunctions: true,
        includeTransactionFunctions: true
      }, null, 2)
    },
    'PROJECT_MANAGEMENT': {
      timeout: 30,
      maxConcurrency: 5,
      configParams: JSON.stringify({
        trackingInterval: 'daily',
        notificationEnabled: true
      }, null, 2)
    },
    'KNOWLEDGE_MANAGEMENT': {
      timeout: 45,
      maxConcurrency: 4,
      configParams: JSON.stringify({
        searchEngine: 'vector',
        updateFrequency: 'realtime'
      }, null, 2)
    },
    'QUALITY_ASSURANCE': {
      timeout: 90,
      maxConcurrency: 2,
      configParams: JSON.stringify({
        checkLevel: 'comprehensive',
        autoFix: false
      }, null, 2)
    }
  }
  
  const config = typeConfigs[type]
  if (config) {
    Object.assign(form, config)
  }
}

// 添加环境变量
const addEnvVar = () => {
  form.envVars.push({ key: '', value: '' })
}

// 移除环境变量
const removeEnvVar = (index) => {
  form.envVars.splice(index, 1)
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
    if (form.configParams) {
      const parsed = JSON.parse(form.configParams)
      form.configParams = JSON.stringify(parsed, null, 2)
      ElMessage.success('JSON格式化成功')
    }
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

// 验证配置JSON
const validateConfig = () => {
  try {
    if (form.configParams) {
      JSON.parse(form.configParams)
      ElMessage.success('JSON格式正确')
    } else {
      ElMessage.info('配置为空')
    }
  } catch (error) {
    ElMessage.error('JSON格式错误：' + error.message)
  }
}

// 加载配置模板
const loadTemplate = () => {
  const template = {
    maxRetries: 3,
    logLevel: 'info',
    enableMetrics: true,
    customSettings: {}
  }
  form.configParams = JSON.stringify(template, null, 2)
}

// 重置表单
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(form, {
    name: '',
    description: '',
    agentType: '',
    version: '1.0.0',
    maxConcurrency: 3,
    endpoint: '',
    timeout: 30,
    retries: 3,
    envVars: [],
    tags: [],
    configParams: ''
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    
    // 验证配置参数
    let configParams = null
    if (form.configParams) {
      try {
        configParams = JSON.parse(form.configParams)
      } catch (error) {
        ElMessage.error('配置参数格式错误，请检查JSON格式')
        return
      }
    }
    
    loading.value = true
    
    const requestData = {
      name: form.name,
      description: form.description,
      agentType: form.agentType,
      version: form.version,
      maxConcurrency: form.maxConcurrency,
      endpoint: form.endpoint || undefined,
      timeout: form.timeout,
      retries: form.retries,
      envVars: form.envVars.filter(env => env.key && env.value),
      tags: form.tags,
      configParams
    }
    
    const res = await registerAgent(requestData)
    
    ElMessage.success('Agent注册成功')
    emit('refresh')
    handleClose()
  } catch (error) {
    ElMessage.error('注册失败：' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  handleReset()
}
</script>

<style scoped>
.agent-form {
  padding: 0 10px;
}

.form-section {
  margin-bottom: 20px;
  border: none;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.type-option {
  display: flex;
  flex-direction: column;
}

.type-label {
  font-weight: 500;
  color: #303133;
}

.type-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-left: 10px;
}

.env-vars {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.env-var-item {
  display: flex;
  align-items: center;
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

.config-actions {
  margin-top: 10px;
  display: flex;
  gap: 10px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .env-var-item {
    flex-direction: column;
    align-items: stretch;
    gap: 10px;
  }
  
  .env-var-item .el-input {
    width: 100% !important;
    margin-right: 0 !important;
  }
}
</style> 