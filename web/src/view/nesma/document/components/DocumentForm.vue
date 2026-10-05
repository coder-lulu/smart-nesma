<template>
  <div class="document-form">
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      label-position="right"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="文档名称" prop="name">
            <el-input
              v-model="form.name"
              placeholder="请输入文档名称"
              maxlength="100"
              show-word-limit
              :disabled="dialogType === 'view'"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="所属项目" prop="projectId">
            <el-select
              v-model="form.projectId"
              placeholder="请选择项目"
              style="width: 100%"
              filterable
              :disabled="dialogType === 'view'"
              @change="handleProjectChange"
            >
              <el-option
                v-for="project in projectOptions"
                :key="project.value"
                :label="project.label"
                :value="project.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 新增：项目周期和版本选择 -->
      <el-row :gutter="20" v-if="form.projectId">
        <el-col :span="12">
          <el-form-item label="项目周期" prop="cycleId">
            <el-select
              v-model="form.cycleId"
              placeholder="请选择项目周期"
              style="width: 100%"
              filterable
              :disabled="dialogType === 'view'"
              @change="handleCycleChange"
            >
              <el-option
                v-for="cycle in cycleOptions"
                :key="cycle.value"
                :label="cycle.label"
                :value="cycle.value"
              >
                <div class="cycle-option">
                  <span>{{ cycle.label }}</span>
                  <el-tag v-if="cycle.status === 'active'" size="small" type="success">进行中</el-tag>
                  <el-tag v-else-if="cycle.status === 'completed'" size="small" type="info">已完成</el-tag>
                  <el-tag v-else size="small" type="warning">{{ cycle.statusText }}</el-tag>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="需求版本" prop="versionId">
            <el-select
              v-model="form.versionId"
              placeholder="请选择需求版本"
              style="width: 100%"
              filterable
              :loading="versionLoading"
              :disabled="dialogType === 'view'"
              @change="handleVersionChange"
            >
              <el-option
                v-for="version in versionOptions"
                :key="version.value"
                :label="version.label"
                :value="version.value"
              >
                <div class="version-option">
                  <span>{{ version.label }}</span>
                  <el-tag v-if="version.createdBy === 'ai_analysis'" size="small" type="primary">AI分析</el-tag>
                  <el-tag v-else size="small" type="info">{{ version.createdBy }}</el-tag>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="文档类型" prop="type">
            <el-select
              v-model="form.type"
              placeholder="请选择文档类型"
              style="width: 100%"
              :disabled="dialogType === 'view'"
              @change="handleTypeChange"
            >
              <el-option label="Excel表格" value="excel" />
              <el-option label="Word文档" value="word" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="文档格式" prop="format">
            <el-select
              v-model="form.format"
              placeholder="请选择文档格式"
              style="width: 100%"
              :disabled="dialogType === 'view'"
              @change="handleFormatChange"
            >
              <el-option
                v-for="option in formatOptions"
                :key="option.value"
                :label="option.label"
                :value="option.value"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="文档模板" prop="templateId">
            <el-select
              v-model="form.templateId"
              placeholder="请选择模板"
              style="width: 100%"
              filterable
              :loading="templateLoading"
              :disabled="dialogType === 'view'"
            >
              <el-option
                v-for="template in templateOptions"
                :key="template.value"
                :label="template.label"
                :value="template.value"
              >
                <div class="template-option">
                  <span>{{ template.label }}</span>
                  <el-tag v-if="template.isDefault" size="small" type="success">默认</el-tag>
                </div>
              </el-option>
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="文档版本" prop="version">
            <el-input
              v-model="form.version"
              placeholder="如：v1.0.0"
              maxlength="20"
              :disabled="dialogType === 'view'"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="文档描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入文档描述"
          maxlength="500"
          show-word-limit
          :disabled="dialogType === 'view'"
        />
      </el-form-item>

      <el-form-item label="生成配置" v-if="dialogType !== 'view'">
        <el-card class="config-card">
          <template #header>
            <div class="config-header">
              <span>文档生成参数</span>
              <el-button size="small" type="primary" text @click="resetConfig">
                重置为默认配置
              </el-button>
            </div>
          </template>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="是否包含图表" label-width="100px">
                <el-switch v-model="form.config.includeCharts" />
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="是否包含统计" label-width="100px">
                <el-switch v-model="form.config.includeStatistics" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20" v-if="form.format === 'nesma_report'">
            <el-col :span="12">
              <el-form-item label="评估详细度" label-width="100px">
                <el-select v-model="form.config.assessmentDetail" style="width: 100%">
                  <el-option label="简要" value="brief" />
                  <el-option label="标准" value="standard" />
                  <el-option label="详细" value="detailed" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="包含建议" label-width="100px">
                <el-switch v-model="form.config.includeRecommendations" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-row :gutter="20" v-if="form.format === 'requirement_spec'">
            <el-col :span="12">
              <el-form-item label="需求层级" label-width="100px">
                              <el-checkbox-group v-model="form.config.requirementLevels">
                <el-checkbox value="1">一级需求</el-checkbox>
                <el-checkbox value="2">二级需求</el-checkbox>
                <el-checkbox value="3">三级需求</el-checkbox>
                <el-checkbox value="4">功能点</el-checkbox>
              </el-checkbox-group>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="包含用例图" label-width="100px">
                <el-switch v-model="form.config.includeUseCaseDiagram" />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="自定义变量" label-width="100px" v-if="templateVariables.length > 0">
            <div class="custom-variables">
              <div 
                v-for="variable in templateVariables" 
                :key="variable.name"
                class="variable-item"
              >
                <el-form-item 
                  :label="variable.description || variable.name" 
                  label-width="120px"
                  style="margin-bottom: 12px;"
                >
                  <el-input
                    v-if="variable.type === 'string'"
                    v-model="form.config.customVariables[variable.name]"
                    :placeholder="`请输入${variable.description || variable.name}`"
                  />
                  <el-input-number
                    v-else-if="variable.type === 'number'"
                    v-model="form.config.customVariables[variable.name]"
                    :min="variable.min || 0"
                    :max="variable.max || 9999"
                    style="width: 100%;"
                  />
                  <el-switch
                    v-else-if="variable.type === 'boolean'"
                    v-model="form.config.customVariables[variable.name]"
                  />
                  <el-select
                    v-else-if="variable.type === 'select'"
                    v-model="form.config.customVariables[variable.name]"
                    style="width: 100%;"
                  >
                    <el-option
                      v-for="option in variable.options"
                      :key="option"
                      :label="option"
                      :value="option"
                    />
                  </el-select>
                </el-form-item>
              </div>
            </div>
          </el-form-item>
        </el-card>
      </el-form-item>

      <el-form-item label="生成方式" v-if="dialogType !== 'view'">
        <el-radio-group v-model="form.generateMode">
          <el-radio value="sync">
            <div class="generate-mode-option">
              <div class="mode-title">同步生成</div>
              <div class="mode-desc">立即生成并等待完成（适用于小文档）</div>
            </div>
          </el-radio>
          <el-radio value="async">
            <div class="generate-mode-option">
              <div class="mode-title">异步生成</div>
              <div class="mode-desc">后台生成，完成后通知（适用于大文档）</div>
            </div>
          </el-radio>
        </el-radio-group>
      </el-form-item>
    </el-form>

    <div class="form-actions">
      <el-button @click="handleCancel">{{ dialogType === 'view' ? '关闭' : '取消' }}</el-button>
      <el-button 
        v-if="dialogType !== 'view'" 
        type="primary" 
        @click="handleSubmit" 
        :loading="submitLoading"
      >
        {{ dialogType === 'create' ? '创建并生成' : '更新文档' }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted, computed } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createDocument,
  updateDocument,
  generateDocument,
  getNesmaProjectList,
  getRequirementVersions,
  getProjectCycles,
  getTemplateOptions,
  getTemplateVariables
} from '@/api/nesma'
import { number } from 'echarts/core'

// Props
const props = defineProps({
  formData: {
    type: Object,
    default: () => ({})
  },
  dialogType: {
    type: String,
    default: 'create'
  }
})

// Emits
const emit = defineEmits(['submit', 'cancel'])

// 响应式数据
const formRef = ref(null)
const submitLoading = ref(false)
const templateLoading = ref(false)
const versionLoading = ref(false)
const projectOptions = ref([])
const projectData = ref([]) // 存储完整的项目数据
const cycleOptions = ref([])
const versionOptions = ref([])
const templateOptions = ref([])
const templateVariables = ref([])

// 表单数据
const form = reactive({
  name: '',
  description: '',
  projectId: null,
  cycleId: null,
  versionId: null,
  templateId: null,
  templateType: null,
  type: 'excel',
  format: 'business_summary',
  version: 'v1.0.0',
  generateMode: 'async',
  config: {
    includeCharts: true,
    includeStatistics: true,
    includeRecommendations: true,
    assessmentDetail: 'standard',
    requirementLevels: ['1', '2', '3'],
    includeUseCaseDiagram: false,
    customVariables: {}
  }
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入文档名称', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  projectId: [
    { required: true, message: '请选择项目', trigger: 'change' }
  ],
  cycleId: [
    { required: true, message: '请选择项目周期', trigger: 'change' }
  ],
  versionId: [
    { required: true, message: '请选择需求版本', trigger: 'change' }
  ],
  type: [
    { required: true, message: '请选择文档类型', trigger: 'change' }
  ],
  format: [
    { required: true, message: '请选择文档格式', trigger: 'change' }
  ]
}

// 计算属性
const formatOptions = computed(() => {
  // 根据文档类型返回对应的内置模板
  if (form.type === 'excel') {
    return [
      { label: '业务需求信息汇总表', value: 'business_summary' }
    ]
  } else if (form.type === 'word') {
    return [
      { label: '送审文档完整版', value: 'requirement_spec' }
    ]
  }
  
  return []
})

// 加载项目选项
const loadProjectOptions = async () => {
  try {
    const res = await getNesmaProjectList({ pageSize: 1000 })
    if (res.code === 0) {
      projectData.value = res.data.list // 存储完整的项目数据
      projectOptions.value = res.data.list.map(item => ({
        label: item.name,
        value: item.ID
      }))
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
  }
}

// 加载项目周期选项
const loadCycleOptions = () => {
  if (!form.projectId) return
  
  // 从已加载的项目数据中获取周期信息
  const project = projectData.value.find(p => p.ID === form.projectId)
  if (project && project.cycles) {
    cycleOptions.value = project.cycles.map(cycle => ({
      label: cycle.name,
      value: cycle.ID,
      status: cycle.status,
      statusText: getCycleStatusText(cycle.status)
    }))
  } else {
    cycleOptions.value = []
  }
}

// 为查看模式加载周期选项
const loadCycleOptionsForView = async () => {
  if (!form.projectId) return
  
  try {
    // 获取项目的所有周期
    const res = await getProjectCycles({ projectId: form.projectId })
    if (res.code === 0) {
      cycleOptions.value = res.data.map(cycle => ({
        label: cycle.name,
        value: cycle.ID,
        status: cycle.status,
        statusText: getCycleStatusText(cycle.status)
      }))
    }
  } catch (error) {
    console.error('加载项目周期失败:', error)
  }
}

// 获取周期状态文本
const getCycleStatusText = (status) => {
  const statusMap = {
    'planning': '规划中',
    'active': '进行中',
    'completed': '已完成',
    'paused': '暂停'
  }
  return statusMap[status] || status
}

// 加载版本选项
const loadVersionOptions = async () => {
  if (!form.cycleId) return
  
  versionLoading.value = true
  try {
    const res = await getRequirementVersions(form.cycleId)
    if (res.code === 0) {
      versionOptions.value = res.data.map(item => ({
        label: `${item.version} - ${item.summary || ''}`,
        value: item.ID,
        createdBy: item.createdBy,
        versionType: item.versionType
      }))
    }
  } catch (error) {
    console.error('加载版本列表失败:', error)
  } finally {
    versionLoading.value = false
  }
}

// 为查看模式加载版本选项
const loadVersionOptionsForView = async () => {
  if (!form.cycleId) return
  
  try {
    const res = await getRequirementVersions(form.cycleId)
    if (res.code === 0) {
      versionOptions.value = res.data.map(item => ({
        label: `${item.version} - ${item.summary || ''}`,
        value: item.ID,
        createdBy: item.createdBy,
        versionType: item.versionType
      }))
    }
  } catch (error) {
    console.error('加载版本列表失败:', error)
  }
}

// 加载模板选项
const loadTemplateOptions = async () => {
  if (!form.type || !form.format) return
  
  templateLoading.value = true
  try {
    const res = await getTemplateOptions({
      type: form.type,
      format: form.format
    })
    if (res.code === 0) {
      // 确保 res.data 是数组格式
      const dataList = Array.isArray(res.data) ? res.data : (res.data?.options || [])
      templateOptions.value = dataList.map(item => ({
        label: item.name,
        value: String(item.ID || item.id), // 确保转换为字符串
        isDefault: item.isDefault
      }))
      
      // 自动选择默认模板
      const defaultTemplate = templateOptions.value.find(t => t.isDefault)
      if (defaultTemplate && !form.templateId) {
        form.templateId = defaultTemplate.value
        loadTemplateVariables()
      }
    }
  } catch (error) {
    console.error('加载模板选项失败:', error)
  } finally {
    templateLoading.value = false
  }
}

// 为查看模式加载模板选项
const loadTemplateOptionsForView = async () => {
  if (!form.type || !form.format) return
  
  try {
    const res = await getTemplateOptions({
      type: form.type,
      format: form.format
    })
    if (res.code === 0) {
      const dataList = Array.isArray(res.data) ? res.data : (res.data?.options || [])
      templateOptions.value = dataList.map(item => ({
        label: item.name,
        value: String(item.ID || item.id),
        isDefault: item.isDefault
      }))
    }
  } catch (error) {
    console.error('加载模板选项失败:', error)
  }
}

// 监听props变化
watch(() => props.formData, async (newData) => {
  if (newData && Object.keys(newData).length > 0) {
    Object.assign(form, {
      ...newData,
      config: {
        ...form.config,
        ...(newData.config || {})
      }
    })
    
    // 如果是查看模式，需要加载关联的项目、周期和版本数据
    if (props.dialogType === 'view' && newData.projectId) {
      await loadProjectOptions()
      await loadCycleOptionsForView()
      await loadVersionOptionsForView()
      await loadTemplateOptionsForView()
    }
  }
}, { immediate: true, deep: true })

// 生命周期
onMounted(() => {
  loadProjectOptions()
  loadTemplateOptions()
})

// 加载模板变量
const loadTemplateVariables = async () => {
  if (!form.templateId) return
  
  try {
    const res = await getTemplateVariables({ templateId: form.templateId })
    if (res.code === 0) {
      // 确保 res.data.variables 是数组，如果不存在则使用空数组
      const variables = res.data?.variables || []
      templateVariables.value = Array.isArray(variables) ? variables : []
      
      // 初始化自定义变量
      templateVariables.value.forEach(variable => {
        if (!(variable.name in form.config.customVariables)) {
          form.config.customVariables[variable.name] = variable.defaultValue
        }
      })
    }
  } catch (error) {
    console.error('加载模板变量失败:', error)
    // 出错时设置为空数组
    templateVariables.value = []
  }
}

// 事件处理
const handleProjectChange = () => {
  // 项目变更时重置周期和版本
  form.cycleId = null
  form.versionId = null
  cycleOptions.value = []
  versionOptions.value = []
  
  // 从已加载的项目数据中获取周期信息
  if (form.projectId) {
    loadCycleOptions()
  }
}

const handleCycleChange = () => {
  // 周期变更时重置版本
  form.versionId = null
  versionOptions.value = []
  console.log('form.cycleId', form.cycleId)
  // 加载版本列表
  if (form.cycleId) {
    loadVersionOptions()
  }
}

const handleVersionChange = () => {
  // 版本变更时的处理逻辑
  console.log('Selected version:', form.versionId)
}

const handleTypeChange = () => {
  // 重置格式（自动选择对应的内置模板）
  const firstOption = formatOptions.value[0]
  if (firstOption) {
    form.format = firstOption.value
    // 设置对应的内置模板ID
    setBuiltinTemplateId()
  }
  
  templateOptions.value = []
  templateVariables.value = []
  form.config.customVariables = {}
}

const handleFormatChange = () => {
  // 设置对应的内置模板ID
  setBuiltinTemplateId()
  
  templateOptions.value = []
  templateVariables.value = []
  form.config.customVariables = {}
}

// 设置内置模板ID的辅助函数
const setBuiltinTemplateId = () => {
  if (form.type === 'excel' && form.format === 'business_summary') {
    form.templateType = 'builtin_excel'
  } else if (form.type === 'word' && (form.format === 'requirement_spec' || form.format === 'nesma_report')) {
    form.templateType = 'builtin_word'
  }
  loadTemplateOptions()
}

// 重置配置
const resetConfig = () => {
  form.config = {
    includeCharts: true,
    includeStatistics: true,
    includeRecommendations: true,
    assessmentDetail: 'standard',
    requirementLevels: ['1', '2', '3'],
    includeUseCaseDiagram: false,
    customVariables: {}
  }
  
  // 重新初始化模板变量默认值
  if (Array.isArray(templateVariables.value)) {
    templateVariables.value.forEach(variable => {
      form.config.customVariables[variable.name] = variable.defaultValue
    })
  }
}

// 表单提交
const handleSubmit = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  
  submitLoading.value = true
  try {
    // 准备提交数据，确保内置模板使用正确的templateId
    const submitData = {
      ...form,
      cycleId: form.cycleId,
      versionId: form.versionId,
      templateType: 'builtin_excel',
      // templateId: Number(form.templateId)
    }
    
    // 根据格式自动设置内置模板ID
    if (form.format === 'business_summary' && form.type === 'excel') {
      submitData.templateType = 'builtin_excel'
    } else if (form.format === 'requirement_spec' && form.type === 'word') {
      submitData.templateType = 'builtin_word'
    } else if (form.format === 'nesma_report' && form.type === 'word') {
      submitData.templateType = 'builtin_word'
    }
    
    let res
    if (props.dialogType === 'create') {
      // 创建文档
      res = await createDocument(submitData)
      if (res.code === 0) {
        ElMessage.success('文档创建成功')
        
        // 如果是同步生成，立即生成文档
        if (form.generateMode === 'sync') {
          const generateRes = await generateDocument({
            documentId: res.data.ID,
            async: false
          })
          if (generateRes.code === 0) {
            ElMessage.success('文档生成完成')
          }
        } else {
          // 异步生成
          await generateDocument({
            documentId: res.data.ID,
            async: true
          })
          ElMessage.success('文档已加入生成队列')
        }
        
        // 创建成功后立即关闭窗口
        emit('submit')
      }
    } else {
      // 更新文档
      res = await updateDocument({ 
        ...submitData, 
        ID: props.formData.ID
      })
      if (res.code === 0) {
        ElMessage.success('文档更新成功')
        emit('submit')
      }
    }
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  } finally {
    submitLoading.value = false
    
  }
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
}

// 监听模板变化
watch(() => form.templateId, () => {
  if (form.templateId) {
    loadTemplateVariables()
  } else {
    templateVariables.value = []
    form.config.customVariables = {}
  }
})

// 监听类型和格式变化
watch([() => form.type, () => form.format], () => {
  if (form.type && form.format) {
    // 对于内置模板格式，不需要加载外部模板选项
    const isBuiltinTemplate = (form.type === 'excel' && form.format === 'business_summary') ||
                             (form.type === 'word' && (form.format === 'requirement_spec' || form.format === 'nesma_report'))
    
    if (!isBuiltinTemplate) {
      loadTemplateOptions()
    } else {
      // 确保内置模板ID设置正确
      setBuiltinTemplateId()
    }
  }
})
</script>

<style scoped>
.document-form {
  padding: 20px;
}

.config-card {
  width: 100%;
}

.config-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.template-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.generate-mode-option {
  margin-left: 8px;
}

.mode-title {
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.mode-desc {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
}

.custom-variables {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 16px;
  background: var(--el-fill-color-light);
}

.variable-item {
  margin-bottom: 8px;
}

.variable-item:last-child {
  margin-bottom: 0;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color);
}
</style> 