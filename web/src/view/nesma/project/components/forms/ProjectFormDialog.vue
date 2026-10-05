<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="900px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      class="project-form"
    >
      <!-- 基础信息 -->
      <div class="form-section">
        <h4 class="section-title">基础信息</h4>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="项目名称" prop="name">
              <el-input 
                v-model="form.name" 
                placeholder="请输入项目名称"
                maxlength="255"
                show-word-limit
              />
            </el-form-item>
          </el-col>
          
          <el-col :span="12">
            <el-form-item label="项目领域" prop="domain">
              <el-select v-model="form.domain" placeholder="请选择项目领域" style="width: 100%">
                <el-option label="金融" value="finance" />
                <el-option label="电商" value="ecommerce" />
                <el-option label="医疗" value="healthcare" />
                <el-option label="教育" value="education" />
                <el-option label="政务" value="government" />
                <el-option label="制造" value="manufacturing" />
                <el-option label="物流" value="logistics" />
                <el-option label="其他" value="other" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item label="项目描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入项目描述，建议包含项目目标、功能范围等信息"
            maxlength="1000"
            show-word-limit
          />
        </el-form-item>

        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="项目状态" prop="status">
              <el-select v-model="form.status" placeholder="请选择项目状态" style="width: 100%">
                <el-option label="活跃" value="active" />
                <el-option label="暂停" value="paused" />
                <el-option label="完成" value="completed" />
                <el-option label="归档" value="archived" />
                <el-option label="规划中" value="planning" />
              </el-select>
            </el-form-item>
          </el-col>
          
        </el-row>
      </div>

      <!-- 时间计划 -->
      <div class="form-section">
        <h4 class="section-title">时间计划</h4>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="开始时间" prop="startDate">
              <el-date-picker
                v-model="form.startDate"
                type="date"
                placeholder="请选择开始时间"
                style="width: 100%"
                :disabled-date="disabledStartDate"
              />
            </el-form-item>
          </el-col>
          
          <el-col :span="12">
            <el-form-item label="结束时间" prop="endDate">
              <el-date-picker
                v-model="form.endDate"
                type="date"
                placeholder="请选择结束时间"
                style="width: 100%"
                :disabled-date="disabledEndDate"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </div>

      <!-- 领域标签 -->
      <div class="form-section">
        <h4 class="section-title">领域标签</h4>
        
        <el-form-item label="领域标签">
          <el-select
            v-model="form.domainTags"
            multiple
            filterable
            allow-create
            placeholder="请选择或输入领域标签"
            style="width: 100%"
          >
            <el-option label="后端开发" value="backend" />
            <el-option label="前端开发" value="frontend" />
            <el-option label="数据库设计" value="database" />
            <el-option label="用户界面" value="ui" />
            <el-option label="数据分析" value="analytics" />
            <el-option label="人工智能" value="ai" />
            <el-option label="微服务" value="microservice" />
            <el-option label="云原生" value="cloud-native" />
          </el-select>
        </el-form-item>
      </div>

      <!-- 项目配置 -->
      <div class="form-section">
        <h4 class="section-title">项目配置</h4>
        
        <el-form-item label="项目配置">
          <el-checkbox v-model="form.settings.autoAnalysis">启用自动分析</el-checkbox>
          <el-checkbox v-model="form.settings.notifications">邮件通知</el-checkbox>
          <el-checkbox v-model="form.settings.publicVisible">公开可见</el-checkbox>
          <el-checkbox v-model="form.settings.templateEnabled">启用模板</el-checkbox>
        </el-form-item>
      </div>

      <!-- 项目周期 -->
      <div class="form-section" v-if="mode === 'edit'">
        <h4 class="section-title">项目周期</h4>
        
        <el-form-item label="当前活跃周期">
          <el-select v-model="form.activeCycleId" placeholder="请选择当前活跃周期" style="width: 100%">
            <el-option
              v-for="cycle in cycleOptions"
              :key="cycle.id"
              :label="cycle.name"
              :value="cycle.id"
            />
          </el-select>
        </el-form-item>
      </div>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">取消</el-button>
        <el-button @click="handleReset" v-if="mode === 'create'">重置</el-button>
        <el-button type="primary" @click="handleSave" :loading="saving">
          {{ mode === 'create' ? '创建' : '更新' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  formData: {
    type: Object,
    default: () => ({})
  },
  mode: {
    type: String,
    default: 'create' // create | edit
  },
  userOptions: {
    type: Array,
    default: () => []
  },
  cycleOptions: {
    type: Array,
    default: () => []
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'save', 'cancel'])

// 响应式数据
const formRef = ref(null)
const saving = ref(false)

const form = reactive({
  name: '',
  description: '',
  domain: '',
  status: 'active',
  startDate: null,
  endDate: null,
  ownerId: null,
  domainTags: [],
  settings: {
    autoAnalysis: true,
    notifications: true,
    publicVisible: false,
    templateEnabled: false
  },
  activeCycleId: null
})

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const title = computed(() => {
  return props.mode === 'create' ? '创建项目' : '编辑项目'
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入项目名称', trigger: 'blur' },
    { min: 2, max: 255, message: '项目名称长度在 2 到 255 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入项目描述', trigger: 'blur' },
    { min: 10, max: 1000, message: '项目描述长度在 10 到 1000 个字符', trigger: 'blur' }
  ],
  domain: [
    { required: true, message: '请选择项目领域', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择项目状态', trigger: 'change' }
  ],
  ownerId: [
    { required: true, message: '请选择项目负责人', trigger: 'change' }
  ]
}

// 方法
const disabledStartDate = (time) => {
  if (form.endDate) {
    const endDate = new Date(form.endDate)
    return time.getTime() > endDate.getTime()
  }
  return false
}

const disabledEndDate = (time) => {
  if (form.startDate) {
    const startDate = new Date(form.startDate)
    return time.getTime() < startDate.getTime()
  }
  return false
}

const handleSave = async () => {
  try {
    if (!formRef.value) {
      ElMessage.error('表单未准备就绪，请稍后再试')
      return
    }
    
    await formRef.value.validate()
    
    saving.value = true
    
    // 组装保存数据，确保与后端模型匹配
    const saveData = {
      ...form,
      // 处理domainTags数组
      domainTags: form.domainTags,
      // 处理settings对象
      settings: form.settings
    }
    
    emit('save', saveData)
  } catch (error) {
    ElMessage.error('请检查表单信息')
  } finally {
    saving.value = false
  }
}

const handleCancel = () => {
  emit('cancel')
}

const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  
  Object.assign(form, {
    name: '',
    description: '',
    domain: '',
    status: 'active',
    startDate: null,
    endDate: null,
    ownerId: null,
    domainTags: [],
    settings: {
      autoAnalysis: true,
      notifications: true,
      publicVisible: false,
      templateEnabled: false
    },
    activeCycleId: null
  })
}

const handleClose = () => {
  emit('cancel')
}

// 监听表单数据变化
watch(() => props.formData, (newData) => {
  if (newData && Object.keys(newData).length > 0) {
    Object.assign(form, {
      name: '',
      description: '',
      domain: '',
      status: 'active',
      startDate: null,
      endDate: null,
      ownerId: null,
      domainTags: [],
      settings: {
        autoAnalysis: true,
        notifications: true,
        publicVisible: false,
        templateEnabled: false
      },
      activeCycleId: null,
      ...newData,
      // 特殊处理JSON字段
      domainTags: newData.domainTags || [],
      settings: {
        autoAnalysis: true,
        notifications: true,
        publicVisible: false,
        templateEnabled: false,
        ...(newData.settings || {})
      }
    })
  }
}, { immediate: true, deep: true })

// 监听对话框显示状态
watch(visible, (newVisible) => {
  if (newVisible && props.mode === 'create') {
    setTimeout(() => {
      handleReset()
    }, 100)
  }
})
</script>

<style lang="scss" scoped>
.project-form {
  .form-section {
    margin-bottom: 24px;
    
    &:last-child {
      margin-bottom: 0;
    }

    .section-title {
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      margin: 0 0 16px 0;
      padding-bottom: 8px;
      border-bottom: 1px solid #ebeef5;
    }
  }

  .user-option {
    display: flex;
    justify-content: space-between;
    align-items: center;

    .user-role {
      font-size: 12px;
      color: #909399;
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

// 响应式设计
@media (max-width: 768px) {
  .project-form {
    .form-section {
      margin-bottom: 16px;
    }
  }
}
</style>