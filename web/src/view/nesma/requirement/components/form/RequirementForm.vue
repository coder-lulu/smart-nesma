<template>
  <el-dialog
    :model-value="modelValue"
    :title="dialogTitle"
    width="1000px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @update:model-value="$emit('update:modelValue', $event)"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      @submit.prevent
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="所属项目" prop="projectId">
            <el-select 
              v-model="form.projectId" 
              placeholder="选择项目" 
              style="width: 100%"
              :disabled="isEdit || !projects.length"
              @change="handleProjectChange"
            >
              <el-option
                v-for="project in projects"
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
              v-model="form.level" 
              placeholder="选择层级" 
              style="width: 100%"
              :disabled="isEdit"
              @change="handleLevelChange"
            >
              <el-option label="L1 - 一级模块" :value="1" />
              <el-option label="L2 - 二级模块" :value="2" />
              <el-option label="L3 - 三级模块" :value="3" />
              <el-option label="L4 - 功能点" :value="4" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="建设周期" prop="cycleId">
            <el-select 
              v-model="form.cycleId" 
              placeholder="选择周期" 
              style="width: 100%"
              :disabled="isEdit || !cycles.length"
            >
              <el-option
                v-for="cycle in cycles"
                :key="cycle.id"
                :label="cycle.name"
                :value="cycle.id"
              />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="需求版本" prop="versionId">
            <el-select 
              v-model="form.versionId" 
              placeholder="选择版本" 
              style="width: 100%"
              :disabled="isEdit || !versions.length"
            >
              <el-option
                v-for="version in versions"
                :key="version.id"
                :label="version.version"
                :value="version.id"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item v-if="form.level > 1" label="父级需求" prop="parentId">
        <el-select 
          v-model="form.parentId" 
          placeholder="选择父级需求" 
          style="width: 100%"
          clearable
          filterable
        >
          <el-option
            v-for="parent in parentOptions"
            :key="parent.id"
            :label="parent.fullPath || parent.title"
            :value="parent.id"
          >
            <div class="parent-option">
              <span class="parent-title">{{ parent.title }}</span>
              <el-tag size="small" :type="getLevelType(parent.level)">
                L{{ parent.level }}
              </el-tag>
            </div>
          </el-option>
        </el-select>
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="需求编号" prop="code">
            <el-input 
              v-model="form.code" 
              placeholder="请输入需求编号（可选）" 
              clearable
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="排序索引" prop="orderIndex">
            <el-input-number 
              v-model="form.orderIndex" 
              :min="0" 
              :max="9999"
              style="width: 100%"
              placeholder="用于控制显示顺序"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="需求标题" prop="title">
        <el-input 
          v-model="form.title" 
          placeholder="请输入需求标题" 
          clearable
          maxlength="500"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="需求描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="4"
          placeholder="请输入需求描述"
          maxlength="2000"
          show-word-limit
        />
      </el-form-item>

      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="优先级" prop="priority">
            <el-rate 
              v-model="form.priority" 
              :max="5" 
              show-text
              :texts="['很低', '较低', '一般', '较高', '很高']"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="状态" prop="status">
            <el-select v-model="form.status" placeholder="选择状态" style="width: 100%">
              <el-option label="待处理" value="pending" />
              <el-option label="进行中" value="in_progress" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="复杂度" prop="complexity">
            <el-select v-model="form.complexity" placeholder="选择复杂度" style="width: 100%">
              <el-option label="简单" value="简单" />
              <el-option label="中等" value="中等" />
              <el-option label="复杂" value="复杂" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="预估工时" prop="estimateHours">
            <el-input-number 
              v-model="form.estimateHours" 
              :min="0" 
              :precision="1"
              style="width: 100%"
              placeholder="小时"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="实际工时" prop="actualHours">
            <el-input-number 
              v-model="form.actualHours" 
              :min="0" 
              :precision="1"
              style="width: 100%"
              placeholder="小时"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- NESMA相关字段 -->
      <el-divider content-position="left">
        <span class="divider-title">NESMA功能点信息</span>
      </el-divider>

      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="功能类型" prop="functionType">
            <el-select v-model="form.functionType" placeholder="选择功能类型" style="width: 100%">
              <el-option label="EI - 外部输入" value="EI" />
              <el-option label="EO - 外部输出" value="EO" />
              <el-option label="EQ - 外部查询" value="EQ" />
              <el-option label="ILF - 内部逻辑文件" value="ILF" />
              <el-option label="EIF - 外部接口文件" value="EIF" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="重用程度" prop="reuseLevel">
            <el-select v-model="form.reuseLevel" placeholder="选择重用程度" style="width: 100%">
              <el-option label="高" value="高" />
              <el-option label="中" value="中" />
              <el-option label="低" value="低" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="修改类型" prop="modificationType">
            <el-select v-model="form.modificationType" placeholder="选择修改类型" style="width: 100%">
              <el-option label="新增" value="新增" />
              <el-option label="优化" value="优化" />
              <el-option label="删除" value="删除" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="AFP值" prop="afp">
            <el-input-number 
              v-model="form.afp" 
              :min="0" 
              :precision="2"
              style="width: 100%"
              placeholder="调整后功能点数"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="UFP值" prop="ufp">
            <el-input-number 
              v-model="form.ufp" 
              :min="0" 
              :precision="2"
              style="width: 100%"
              placeholder="未调整功能点数"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 业务信息 -->
      <el-divider content-position="left">
        <span class="divider-title">业务信息</span>
      </el-divider>

      <el-form-item label="业务价值" prop="businessValue">
        <el-input
          v-model="form.businessValue"
          type="textarea"
          :rows="2"
          placeholder="请描述该需求的业务价值"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="验收标准" prop="acceptanceCriteria">
        <el-input
          v-model="form.acceptanceCriteria"
          type="textarea"
          :rows="3"
          placeholder="请输入验收标准，明确完成条件"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="备注" prop="notes">
        <el-input
          v-model="form.notes"
          type="textarea"
          :rows="2"
          placeholder="其他备注信息"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <span class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleSubmit"
          :loading="loading"
        >
          {{ isEdit ? '更新' : '创建' }}
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
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
  isEdit: {
    type: Boolean,
    default: false
  },
  projects: {
    type: Array,
    default: () => []
  },
  cycles: {
    type: Array,
    default: () => []
  },
  versions: {
    type: Array,
    default: () => []
  },
  parentOptions: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'submit', 'cancel'])

// 响应式数据
const formRef = ref(null)
const form = reactive({
  projectId: null,
  cycleId: null,
  versionId: null,
  parentId: null,
  level: 1,
  code: '',
  title: '',
  description: '',
  priority: 3,
  status: 'pending',
  orderIndex: 0,
  complexity: '中等',
  estimateHours: null,
  actualHours: null,
  businessValue: '',
  acceptanceCriteria: '',
  notes: '',
  // NESMA字段
  functionType: '',
  reuseLevel: '中',
  modificationType: '新增',
  afp: 0,
  ufp: 0
})

// 表单验证规则
const rules = {
  projectId: [
    { required: true, message: '请选择所属项目', trigger: 'change' }
  ],
  cycleId: [
    { required: true, message: '请选择建设周期', trigger: 'change' }
  ],
  versionId: [
    { required: true, message: '请选择需求版本', trigger: 'change' }
  ],
  level: [
    { required: true, message: '请选择需求层级', trigger: 'change' }
  ],
  title: [
    { required: true, message: '请输入需求标题', trigger: 'blur' },
    { min: 2, max: 500, message: '标题长度应在2-500字符之间', trigger: 'blur' }
  ],
  priority: [
    { required: true, message: '请选择优先级', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ],
  parentId: [
    { required: true, message: '请选择父级需求', trigger: 'change' }
  ]
}

// 计算属性
const dialogTitle = computed(() => {
  return props.isEdit ? '编辑需求' : '新建需求'
})

// 方法
const getLevelType = (level) => {
  const typeMap = {
    1: 'danger',
    2: 'warning',
    3: 'primary',
    4: 'success'
  }
  return typeMap[level] || 'info'
}

const handleProjectChange = () => {
  // 项目变更时清空周期和版本
  form.cycleId = null
  form.versionId = null
}

const handleLevelChange = () => {
  // 层级变更时清空父级需求
  form.parentId = null
  
  // 动态设置父级需求验证规则
  if (form.level > 1) {
    rules.parentId = [
      { required: true, message: '请选择父级需求', trigger: 'change' }
    ]
  } else {
    delete rules.parentId
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    
    // 清理数据
    const submitData = { ...form }
    
    // 如果是一级需求，清空parentId
    if (submitData.level === 1) {
      submitData.parentId = null
    }
    
    // 清理空值
    Object.keys(submitData).forEach(key => {
      if (submitData[key] === '' || submitData[key] === null) {
        if (['estimateHours', 'actualHours', 'afp', 'ufp'].includes(key)) {
          submitData[key] = null
        } else if (!['priority', 'orderIndex', 'level'].includes(key)) {
          delete submitData[key]
        }
      }
    })
    
    emit('submit', submitData)
  } catch (error) {
    ElMessage.error('请检查表单填写是否正确')
  }
}

const handleClose = () => {
  emit('cancel')
}

const resetForm = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  
  // 重置为默认值
  Object.assign(form, {
    projectId: null,
    cycleId: null,
    versionId: null,
    parentId: null,
    level: 1,
    code: '',
    title: '',
    description: '',
    priority: 3,
    status: 'pending',
    orderIndex: 0,
    complexity: '中等',
    estimateHours: null,
    actualHours: null,
    businessValue: '',
    acceptanceCriteria: '',
    notes: '',
    functionType: '',
    reuseLevel: '中',
    modificationType: '新增',
    afp: 0,
    ufp: 0
  })
}

// 监听器
watch(() => props.formData, (newData) => {
  if (newData && Object.keys(newData).length > 0) {
    Object.assign(form, newData)
  }
}, { deep: true, immediate: true })

watch(() => props.modelValue, (newValue) => {
  if (newValue) {
    nextTick(() => {
      if (formRef.value) {
        formRef.value.clearValidate()
      }
    })
  } else {
    resetForm()
  }
})
</script>

<style lang="scss" scoped>
.parent-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;

  .parent-title {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
}

.divider-title {
  font-weight: 600;
  color: #303133;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-form-item__label) {
  font-weight: 500;
}

:deep(.el-textarea__inner) {
  resize: vertical;
}
</style>