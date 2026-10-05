<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="800px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-width="120px"
      @submit.prevent
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="需求名称" prop="name">
            <el-input
              v-model="formData.name"
              placeholder="请输入需求名称"
              clearable
              maxlength="200"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="需求编号" prop="code">
            <el-input
              v-model="formData.code"
              placeholder="请输入需求编号"
              clearable
              maxlength="50"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="所属项目" prop="projectId">
            <el-select
              v-model="formData.projectId"
              placeholder="请选择项目"
              style="width: 100%"
              @change="handleProjectChange"
            >
              <el-option
                v-for="project in projectOptions"
                :key="project.ID"
                :label="project.name"
                :value="project.ID"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="需求层级" prop="level">
            <el-select
              v-model="formData.level"
              placeholder="请选择需求层级"
              style="width: 100%"
              @change="handleLevelChange"
            >
              <el-option label="L1 - 主要功能分组" :value="1" />
              <el-option label="L2 - 功能子分组" :value="2" />
              <el-option label="L3 - 具体功能" :value="3" />
              <el-option label="L4 - 功能点" :value="4" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="父级需求" prop="parentId">
            <el-select
              v-model="formData.parentId"
              placeholder="请选择父级需求"
              style="width: 100%"
              clearable
              :disabled="!formData.projectId || formData.level === 1"
            >
              <el-option
                v-for="parent in parentOptions"
                :key="parent.ID"
                :label="`${parent.name} (${parent.code})`"
                :value="parent.ID"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="需求状态" prop="status">
            <el-select
              v-model="formData.status"
              placeholder="请选择状态"
              style="width: 100%"
            >
              <el-option label="待处理" value="pending" />
              <el-option label="进行中" value="in_progress" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="优先级" prop="priority">
            <el-select
              v-model="formData.priority"
              placeholder="请选择优先级"
              style="width: 100%"
            >
              <el-option label="低" value="low" />
              <el-option label="中" value="medium" />
              <el-option label="高" value="high" />
              <el-option label="紧急" value="urgent" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="排序" prop="sort">
            <el-input-number
              v-model="formData.sort"
              :min="0"
              :max="999"
              style="width: 100%"
              placeholder="输入排序值"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 功能点特有字段 -->
      <div v-if="formData.level === 4" class="function-point-fields">
        <el-divider content-position="left">功能点信息</el-divider>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="功能点类型" prop="functionPointType">
              <el-select
                v-model="formData.functionPointType"
                placeholder="请选择功能点类型"
                style="width: 100%"
              >
                <el-option label="数据功能" value="data" />
                <el-option label="事务功能" value="transaction" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="复杂度" prop="complexity">
              <el-select
                v-model="formData.complexity"
                placeholder="请选择复杂度"
                style="width: 100%"
              >
                <el-option label="简单" value="low" />
                <el-option label="中等" value="average" />
                <el-option label="复杂" value="high" />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="功能点数" prop="functionPoints">
              <el-input-number
                v-model="formData.functionPoints"
                :min="0"
                :precision="2"
                style="width: 100%"
                placeholder="输入功能点数"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="调整因子" prop="adjustmentFactor">
              <el-input-number
                v-model="formData.adjustmentFactor"
                :min="0"
                :max="2"
                :precision="2"
                style="width: 100%"
                placeholder="输入调整因子"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </div>

      <el-form-item label="需求描述" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          :rows="4"
          placeholder="请输入需求描述"
          maxlength="2000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="验收标准" prop="acceptanceCriteria">
        <el-input
          v-model="formData.acceptanceCriteria"
          type="textarea"
          :rows="3"
          placeholder="请输入验收标准"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="备注" prop="notes">
        <el-input
          v-model="formData.notes"
          type="textarea"
          :rows="2"
          placeholder="请输入备注"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button type="primary" @click="handleConfirm" :loading="loading">
          确定
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  createNesmaRequirement, 
  updateNesmaRequirement, 
  getNesmaProjectList,
  getParentRequirementOptions
} from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  requirement: {
    type: Object,
    default: () => null
  },
  mode: {
    type: String,
    default: 'create' // create, edit, view
  }
})

const emit = defineEmits(['update:modelValue', 'success', 'cancel'])

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const dialogTitle = computed(() => {
  switch (props.mode) {
    case 'create':
      return '新建需求'
    case 'edit':
      return '编辑需求'
    case 'view':
      return '查看需求'
    default:
      return '需求'
  }
})

const formRef = ref(null)
const loading = ref(false)
const projectOptions = ref([])
const parentOptions = ref([])

const formData = ref({
  name: '',
  code: '',
  projectId: null,
  level: null,
  parentId: null,
  status: 'pending',
  priority: 'medium',
  sort: 0,
  functionPointType: '',
  complexity: '',
  functionPoints: null,
  adjustmentFactor: 1.0,
  description: '',
  acceptanceCriteria: '',
  notes: ''
})

const rules = ref({
  name: [
    { required: true, message: '请输入需求名称', trigger: 'blur' },
    { min: 1, max: 200, message: '需求名称长度在1到200个字符', trigger: 'blur' }
  ],
  code: [
    { required: true, message: '请输入需求编号', trigger: 'blur' },
    { min: 1, max: 50, message: '需求编号长度在1到50个字符', trigger: 'blur' }
  ],
  projectId: [
    { required: true, message: '请选择项目', trigger: 'change' }
  ],
  level: [
    { required: true, message: '请选择需求层级', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择需求状态', trigger: 'change' }
  ],
  priority: [
    { required: true, message: '请选择优先级', trigger: 'change' }
  ],
  functionPointType: [
    { 
      validator: (rule, value, callback) => {
        if (formData.value.level === 4 && !value) {
          callback(new Error('功能点类型不能为空'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ],
  complexity: [
    { 
      validator: (rule, value, callback) => {
        if (formData.value.level === 4 && !value) {
          callback(new Error('复杂度不能为空'))
        } else {
          callback()
        }
      },
      trigger: 'change'
    }
  ]
})

// 监听对话框打开
watch(dialogVisible, (newVal) => {
  if (newVal) {
    initForm()
    loadProjectOptions()
  }
})

// 监听项目变化
const handleProjectChange = () => {
  formData.value.parentId = null
  if (formData.value.projectId) {
    loadParentOptions()
  }
}

// 监听层级变化
const handleLevelChange = () => {
  formData.value.parentId = null
  if (formData.value.projectId) {
    loadParentOptions()
  }
}

// 初始化表单
const initForm = () => {
  if (props.requirement && props.mode !== 'create') {
    formData.value = {
      ...formData.value,
      ...props.requirement
    }
  } else {
    formData.value = {
      name: '',
      code: '',
      projectId: null,
      level: null,
      parentId: null,
      status: 'pending',
      priority: 'medium',
      sort: 0,
      functionPointType: '',
      complexity: '',
      functionPoints: null,
      adjustmentFactor: 1.0,
      description: '',
      acceptanceCriteria: '',
      notes: ''
    }
  }
}

// 加载项目选项
const loadProjectOptions = async () => {
  try {
    const res = await getNesmaProjectList({ page: 1, pageSize: 1000 })
    projectOptions.value = res.data.list || []
  } catch (error) {
    console.error('加载项目列表失败:', error)
    ElMessage.error('加载项目列表失败')
  }
}

// 加载父级需求选项
const loadParentOptions = async () => {
  if (!formData.value.projectId || formData.value.level === 1) {
    parentOptions.value = []
    return
  }
  
  try {
    const res = await getParentRequirementOptions({
      projectId: formData.value.projectId,
      level: formData.value.level - 1
    })
    parentOptions.value = res.data || []
  } catch (error) {
    console.error('加载父级需求选项失败:', error)
    ElMessage.error('加载父级需求选项失败')
  }
}

// 确认操作
const handleConfirm = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    loading.value = true
    
    const data = { ...formData.value }
    
    if (props.mode === 'create') {
      await createNesmaRequirement(data)
      ElMessage.success('创建需求成功')
    } else if (props.mode === 'edit') {
      await updateNesmaRequirement(data)
      ElMessage.success('更新需求成功')
    }
    
    emit('success')
    handleClose()
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  emit('cancel')
  emit('update:modelValue', false)
}

onMounted(() => {
  loadProjectOptions()
})
</script>

<style scoped>
.function-point-fields {
  margin-top: 20px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}
</style> 