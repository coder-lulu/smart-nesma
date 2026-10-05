<template>
  <el-dialog
    v-model="dialogVisible"
    :title="isEdit ? '编辑评估' : '创建评估'"
    width="600px"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      label-position="left"
    >
      <el-form-item label="评估名称" prop="evaluationName">
        <el-input
          v-model="form.evaluationName"
          placeholder="请输入评估名称"
          maxlength="200"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="所属项目" prop="projectId">
        <el-select
          v-model="form.projectId"
          placeholder="请选择项目"
          style="width: 100%"
          filterable
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

      <el-form-item label="建设周期" prop="cycleId">
        <el-select
          v-model="form.cycleId"
          placeholder="请选择建设周期"
          style="width: 100%"
          filterable
          :disabled="!form.projectId"
          @change="handleCycleChange"
        >
          <el-option
            v-for="cycle in cycleOptions"
            :key="cycle.value"
            :label="cycle.label"
            :value="cycle.value"
          >
            <span>{{ cycle.label }}</span>
            <span style="float: right; color: #8492a6; font-size: 12px;">
              {{ cycle.statusText }} - {{ cycle.phaseText }}
            </span>
          </el-option>
        </el-select>
      </el-form-item>

      <el-form-item label="需求版本" prop="requirementVersionId">
        <el-select
          v-model="form.requirementVersionId"
          placeholder="请选择需求版本"
          style="width: 100%"
          filterable
          :disabled="!form.cycleId"
          @change="handleVersionChange"
        >
          <el-option
            v-for="version in versionOptions"
            :key="version.value"
            :label="version.label"
            :value="version.value"
          >
            <span>{{ version.label }}</span>
            <span style="float: right; color: #8492a6; font-size: 12px;">
              {{ version.versionTypeText }} - {{ version.statusText }}
            </span>
          </el-option>
        </el-select>
      </el-form-item>

      <el-form-item label="评估版本" prop="evaluationVersion">
        <el-input
          v-model="form.evaluationVersion"
          placeholder="如：v1.0.0"
          maxlength="50"
        />
      </el-form-item>

      <el-form-item label="评估类型" prop="evaluationType">
        <el-radio-group v-model="form.evaluationType">
          <el-radio value="initial">初步评估</el-radio>
          <el-radio value="detailed">详细评估</el-radio>
          <el-radio value="final">最终评估</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="NESMA规则" prop="nesmaRules">
        <el-select v-model="form.nesmaRules" placeholder="选择NESMA规则版本">
          <el-option label="NESMA v2.2" value="v2.2" />
          <el-option label="NESMA v2.1" value="v2.1" />
          <el-option label="NESMA v2.0" value="v2.0" />
        </el-select>
      </el-form-item>

      <el-form-item label="评估人" prop="evaluatorId">
        <el-select
          v-model="form.evaluatorId"
          placeholder="请选择评估人"
          style="width: 100%"
          filterable
        >
          <el-option
            v-for="user in userOptions"
            :key="user.value"
            :label="user.label"
            :value="user.value"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="评估配置">
        <el-card class="config-card">
          <el-form-item label="使用AI辅助" prop="useAIAssisted" style="margin-bottom: 10px;">
            <el-switch
              v-model="form.evaluationConfig.useAIAssisted"
              active-text="开启"
              inactive-text="关闭"
            />
          </el-form-item>

          <el-form-item label="置信度阈值" prop="confidenceThreshold" style="margin-bottom: 10px;">
            <el-slider
              v-model="form.evaluationConfig.confidenceThreshold"
              :min="0.5"
              :max="1.0"
              :step="0.05"
              :format-tooltip="formatPercentage"
              show-input
              style="margin-right: 20px;"
            />
          </el-form-item>

          <el-form-item label="验证级别" prop="validationLevel" style="margin-bottom: 10px;">
            <el-radio-group v-model="form.evaluationConfig.validationLevel">
              <el-radio value="basic">基础验证</el-radio>
              <el-radio value="standard">标准验证</el-radio>
              <el-radio value="strict">严格验证</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="包含改进建议" prop="includeRecommendations" style="margin-bottom: 0;">
            <el-switch
              v-model="form.evaluationConfig.includeRecommendations"
              active-text="是"
              inactive-text="否"
            />
          </el-form-item>
        </el-card>
      </el-form-item>

      <el-form-item label="备注">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入评估备注（可选）"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="submitting"
          @click="handleSubmit"
        >
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { createEvaluation, updateEvaluation } from '@/api/nesmaEvaluation'
import { getNesmaProjectList } from '@/api/nesma'
import { getUserList } from '@/api/user'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  editData: {
    type: Object,
    default: null
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'success'])

// 响应式数据
const formRef = ref(null)
const submitting = ref(false)
const projectOptions = ref([])
const cycleOptions = ref([])
const versionOptions = ref([])
const userOptions = ref([])

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const isEdit = computed(() => !!props.editData)

// 表单数据
const defaultForm = {
  evaluationName: '',
  projectId: '',
  cycleId: '',
  requirementVersionId: '',
  evaluationVersion: 'v1.0.0',
  evaluationType: 'detailed',
  nesmaRules: 'v2.2',
  evaluatorId: '',
  description: '',
  evaluationConfig: {
    useAIAssisted: true,
    confidenceThreshold: 0.75,
    validationLevel: 'standard',
    includeRecommendations: true
  }
}

const form = reactive({ ...defaultForm })

// 表单验证规则
const rules = {
  evaluationName: [
    { required: true, message: '请输入评估名称', trigger: 'blur' },
    { min: 2, max: 200, message: '评估名称长度在 2 到 200 个字符', trigger: 'blur' }
  ],
  projectId: [
    { required: true, message: '请选择项目', trigger: 'change' }
  ],
  cycleId: [
    { required: true, message: '请选择建设周期', trigger: 'change' }
  ],
  requirementVersionId: [
    { required: true, message: '请选择需求版本', trigger: 'change' }
  ],
  evaluationVersion: [
    { required: true, message: '请输入评估版本', trigger: 'blur' }
  ],
  evaluationType: [
    { required: true, message: '请选择评估类型', trigger: 'change' }
  ],
  nesmaRules: [
    { required: true, message: '请选择NESMA规则版本', trigger: 'change' }
  ],
  evaluatorId: [
    { required: true, message: '请选择评估人', trigger: 'change' }
  ]
}

// 方法
const resetForm = () => {
  Object.keys(defaultForm).forEach(key => {
    if (key === 'evaluationConfig') {
      form[key] = { ...defaultForm[key] }
    } else {
      form[key] = defaultForm[key]
    }
  })
  formRef.value?.clearValidate()
}

// 监听编辑数据变化
watch(() => props.editData, (newData) => {
  if (newData) {
    // 编辑模式，填充表单数据
    Object.keys(defaultForm).forEach(key => {
      if (key === 'evaluationConfig') {
        form[key] = { ...defaultForm[key], ...(newData[key] || {}) }
      } else {
        form[key] = newData[key] || defaultForm[key]
      }
    })
  } else {
    // 新建模式，重置表单
    resetForm()
  }
}, { immediate: true })

const handleProjectChange = async (projectId) => {
  const project = projectOptions.value.find(p => p.value === projectId)
  if (project) {
    // 根据项目自动生成评估名称
    if (!form.evaluationName) {
      const now = new Date()
      const dateStr = now.toISOString().slice(0, 10).replace(/-/g, '')
      form.evaluationName = `${project.label}_NESMA评估_${dateStr}`
    }
    
    // 重置后续选择
    form.cycleId = ''
    form.requirementVersionId = ''
    cycleOptions.value = []
    versionOptions.value = []
    
    // 加载项目周期
    await loadProjectCycles(projectId)
  }
}

const handleCycleChange = async (cycleId) => {
  // 重置版本选择
  form.requirementVersionId = ''
  versionOptions.value = []
  
  if (cycleId) {
    // 加载周期版本
    await loadCycleVersions(cycleId)
  }
}

const handleVersionChange = (versionId) => {
  const version = versionOptions.value.find(v => v.value === versionId)
  if (version) {
    // 可以根据版本信息调整评估配置
    console.log('选择版本:', version)
  }
}

const handleSubmit = async () => {
  try {
    await formRef.value.validate()
    submitting.value = true

    const submitData = {
      ...form,
      evaluationConfig: form.evaluationConfig
    }

    if (isEdit.value) {
      submitData.id = props.editData.id
      await updateEvaluation(submitData)
      ElMessage.success('评估更新成功')
    } else {
      await createEvaluation(submitData)
      ElMessage.success('评估创建成功')
    }

    emit('success')
    dialogVisible.value = false
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error(isEdit.value ? '更新失败' : '创建失败')
  } finally {
    submitting.value = false
  }
}

const handleClosed = () => {
  resetForm()
}

const formatPercentage = (value) => {
  return `${(value * 100).toFixed(0)}%`
}

// 加载项目周期
const loadProjectCycles = async (projectId) => {
  try {
    const response = await fetch(`/api/nesma-evaluation/project-cycles?projectId=${projectId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })
    const result = await response.json()
    if (result.code === 0) {
      cycleOptions.value = result.data || []
    } else {
      console.error('获取项目周期失败:', result.msg)
    }
  } catch (error) {
    console.error('获取项目周期失败:', error)
  }
}

// 加载周期版本
const loadCycleVersions = async (cycleId) => {
  try {
    const response = await fetch(`/api/nesma-evaluation/cycle-versions?cycleId=${cycleId}`, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${localStorage.getItem('token')}`,
        'Content-Type': 'application/json'
      }
    })
    const result = await response.json()
    if (result.code === 0) {
      versionOptions.value = result.data || []
    } else {
      console.error('获取周期版本失败:', result.msg)
    }
  } catch (error) {
    console.error('获取周期版本失败:', error)
  }
}

// 获取项目选项
const getProjectOptions = async () => {
  try {
    const response = await getNesmaProjectList({ page: 1, pageSize: 1000 })
    projectOptions.value = response.data?.list?.filter(item => 
      item && item.name && (item.ID !== undefined && item.ID !== null)
    ).map(item => ({
      label: item.name,
      value: item.ID
    })) || []
  } catch (error) {
    console.error('获取项目列表失败:', error)
  }
}

// 获取用户选项
const getUserOptions = async () => {
  try {
    // getUserList API 使用POST方法，需要传递data参数
    const data = { pageSize: 1000, page: 1 }
    const response = await getUserList(data)
    userOptions.value = response.data?.list?.map(item => ({
      label: item.nickName,
      value: item.ID
    })) || []
  } catch (error) {
    console.error('获取用户列表失败:', error)
  }
}

// 生命周期
onMounted(() => {
  getProjectOptions()
  getUserOptions()
})
</script>

<style scoped>
.config-card {
  background-color: #fafafa;
  border: 1px solid #e6e6e6;
}

.config-card :deep(.el-card__body) {
  padding: 15px;
}

.config-card .el-form-item {
  margin-bottom: 10px;
}

.config-card .el-form-item:last-child {
  margin-bottom: 0;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-slider) {
  margin-right: 20px;
}
</style> 