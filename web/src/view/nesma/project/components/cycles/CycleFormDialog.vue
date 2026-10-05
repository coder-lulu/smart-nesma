<template>
  <el-dialog
    v-model="visible"
    :title="title"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="100px"
      class="cycle-form"
    >
      <el-form-item label="周期名称" prop="name">
        <el-input 
          v-model="form.name" 
          placeholder="请输入周期名称，如：一期、二期等"
          maxlength="100"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="周期描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入周期描述"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="周期状态" prop="status">
            <el-select v-model="form.status" placeholder="请选择周期状态" style="width: 100%">
              <el-option label="规划中" value="planning" />
              <el-option label="进行中" value="active" />
              <el-option label="已完成" value="completed" />
              <el-option label="暂停" value="suspended" />
            </el-select>
          </el-form-item>
        </el-col>
        
        <el-col :span="12">
          <el-form-item label="当前阶段" prop="phase">
            <el-select v-model="form.phase" placeholder="请选择当前阶段" style="width: 100%">
              <el-option label="需求分析" value="requirement" />
              <el-option label="系统设计" value="design" />
              <el-option label="开发实施" value="development" />
              <el-option label="测试验证" value="testing" />
              <el-option label="部署上线" value="deployment" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

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

      <el-form-item label="备注">
        <el-input
          v-model="form.remarks"
          type="textarea"
          :rows="2"
          placeholder="周期相关备注信息"
          maxlength="200"
          show-word-limit
        />
      </el-form-item>
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
import { createProjectCycle, updateProjectCycle } from '@/api/projectCycle'

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
  projectId: {
    type: [Number, String],
    required: true
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
  status: 'planning',
  phase: 'requirement',
  startDate: null,
  endDate: null,
  remarks: ''
})

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const title = computed(() => {
  return props.mode === 'create' ? '创建周期' : '编辑周期'
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入周期名称', trigger: 'blur' },
    { min: 1, max: 100, message: '周期名称长度在 1 到 100 个字符', trigger: 'blur' }
  ],
  description: [
    { max: 500, message: '周期描述长度不能超过 500 个字符', trigger: 'blur' }
  ],
  status: [
    { required: true, message: '请选择周期状态', trigger: 'change' }
  ]
}

// 日期验证
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

// 保存
const handleSave = async () => {
  try {
    if (!formRef.value) {
      ElMessage.error('表单未准备就绪，请稍后再试')
      return
    }
    
    await formRef.value.validate()
    
    saving.value = true
    
    // 组装保存数据
    const saveData = {
      ...form,
      projectId: props.projectId
    }

    let res
    if (props.mode === 'create') {
      res = await createProjectCycle(saveData)
    } else {
      saveData.id = props.formData.id
      res = await updateProjectCycle(saveData)
    }
    
    if (res.code === 0) {
      ElMessage.success(props.mode === 'create' ? '创建成功' : '更新成功')
      emit('save', res.data)
    } else {
      ElMessage.error(res.msg || '操作失败')
    }
  } catch (error) {
    console.error('保存周期失败:', error)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 取消
const handleCancel = () => {
  emit('cancel')
}

// 重置
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  
  Object.assign(form, {
    name: '',
    description: '',
    status: 'planning',
    phase: 'requirement',
    startDate: null,
    endDate: null,
    remarks: ''
  })
}

// 关闭
const handleClose = () => {
  emit('cancel')
}

// 监听表单数据变化
watch(() => props.formData, (newData) => {
  if (newData && Object.keys(newData).length > 0) {
    Object.assign(form, {
      name: '',
      description: '',
      status: 'planning',
      phase: 'requirement',
      startDate: null,
      endDate: null,
      remarks: '',
      ...newData,
      // 处理日期格式
      startDate: newData.startDate ? new Date(newData.startDate) : null,
      endDate: newData.endDate ? new Date(newData.endDate) : null
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
.cycle-form {
  .el-form-item {
    margin-bottom: 20px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>