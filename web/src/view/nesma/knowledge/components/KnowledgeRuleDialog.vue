<template>
  <el-dialog
    v-model="dialogVisible"
    :title="getDialogTitle()"
    :width="900"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="120px"
      :disabled="mode === 'view'"
    >
      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="规则名称" prop="ruleName">
            <el-input 
              v-model="formData.ruleName" 
              placeholder="请输入规则名称"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="规则类别" prop="category">
            <el-select
              v-model="formData.category"
              placeholder="请选择规则类别"
              style="width: 100%"
            >
              <el-option label="验证规则" value="VALIDATION" />
              <el-option label="计算规则" value="CALCULATION" />
              <el-option label="业务规则" value="BUSINESS" />
              <el-option label="系统规则" value="SYSTEM" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="优先级" prop="priority">
            <el-select
              v-model="formData.priority"
              placeholder="请选择优先级"
              style="width: 100%"
            >
              <el-option label="高" value="HIGH" />
              <el-option label="中" value="MEDIUM" />
              <el-option label="低" value="LOW" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="状态" prop="status">
            <el-radio-group v-model="formData.status">
              <el-radio value="ACTIVE">启用</el-radio>
              <el-radio value="INACTIVE">禁用</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="作者" prop="author">
            <el-input v-model="formData.author" placeholder="请输入作者" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="规则描述" prop="description">
            <el-input
              v-model="formData.description"
              type="textarea"
              :rows="3"
              placeholder="请输入规则描述"
              maxlength="500"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="条件表达式" prop="conditionExpression">
            <el-input
              v-model="formData.conditionExpression"
              type="textarea"
              :rows="4"
              placeholder="请输入条件表达式，支持JavaScript语法"
              maxlength="2000"
              show-word-limit
            />
            <div class="expression-help">
              <el-icon><InfoFilled /></el-icon>
              <span>支持JavaScript语法，例如: data.functionPoints > 100 && data.complexity === 'HIGH'</span>
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="动作表达式" prop="actionExpression">
            <el-input
              v-model="formData.actionExpression"
              type="textarea"
              :rows="4"
              placeholder="请输入动作表达式，支持JavaScript语法"
              maxlength="2000"
              show-word-limit
            />
            <div class="expression-help">
              <el-icon><InfoFilled /></el-icon>
              <span>支持JavaScript语法，例如: result.adjustmentFactor = 1.2; result.message = '复杂度较高，建议增加工作量'</span>
            </div>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="规则标签" prop="tags">
            <el-select
              v-model="selectedTags"
              multiple
              filterable
              allow-create
              default-first-option
              placeholder="请选择或输入标签"
              style="width: 100%"
            >
              <el-option
                v-for="tag in commonTags"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 查看模式下显示额外信息 -->
      <template v-if="mode === 'view' && formData.id">
        <el-divider content-position="left">执行统计</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-statistic title="执行次数" :value="formData.executionCount || 0" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="成功次数" :value="formData.successCount || 0" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="失败次数" :value="formData.errorCount || 0" />
          </el-col>
        </el-row>
        <el-row :gutter="20" style="margin-top: 20px;">
          <el-col :span="8">
            <el-statistic title="创建时间" :value="formatDate(formData.createdAt)" value-style="font-size: 14px;" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="更新时间" :value="formatDate(formData.updatedAt)" value-style="font-size: 14px;" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="最后执行" :value="formatDate(formData.lastExecutedAt)" value-style="font-size: 14px;" />
          </el-col>
        </el-row>
      </template>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">{{ mode === 'view' ? '关闭' : '取消' }}</el-button>
        <el-button 
          v-if="mode !== 'view'"
          type="primary" 
          :loading="saving"
          @click="handleSave"
        >
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { createKnowledgeRule, updateKnowledgeRule } from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  rule: {
    type: Object,
    default: null
  },
  mode: {
    type: String,
    default: 'create' // create, edit, view
  }
})

const emit = defineEmits(['update:modelValue', 'save', 'cancel'])

// 响应式数据
const formRef = ref(null)
const saving = ref(false)
const selectedTags = ref([])

// 常用标签
const commonTags = [
  'NESMA', '功能点', '复杂度', '验证', '计算', '业务规则', '系统规则',
  '数据功能', '事务功能', '质量管理', '自动化', '人工审核'
]

// 表单数据
const formData = reactive({
  id: null,
  ruleName: '',
  description: '',
  category: '',
  priority: 'MEDIUM',
  status: 'ACTIVE',
  conditionExpression: '',
  actionExpression: '',
  author: '',
  executionCount: 0,
  successCount: 0,
  errorCount: 0,
  createdAt: null,
  updatedAt: null,
  lastExecutedAt: null
})

// 表单验证规则
const formRules = {
  ruleName: [
    { required: true, message: '请输入规则名称', trigger: 'blur' },
    { min: 2, max: 200, message: '规则名称长度应在2-200个字符', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择规则类别', trigger: 'change' }
  ],
  priority: [
    { required: true, message: '请选择优先级', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择状态', trigger: 'change' }
  ],
  conditionExpression: [
    { required: true, message: '请输入条件表达式', trigger: 'blur' },
    { min: 10, max: 2000, message: '条件表达式长度应在10-2000个字符', trigger: 'blur' }
  ],
  actionExpression: [
    { required: true, message: '请输入动作表达式', trigger: 'blur' },
    { min: 10, max: 2000, message: '动作表达式长度应在10-2000个字符', trigger: 'blur' }
  ],
  author: [
    { required: true, message: '请输入作者', trigger: 'blur' }
  ]
}

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    id: null,
    ruleName: '',
    description: '',
    category: '',
    priority: 'MEDIUM',
    status: 'ACTIVE',
    conditionExpression: '',
    actionExpression: '',
    author: '',
    executionCount: 0,
    successCount: 0,
    errorCount: 0,
    createdAt: null,
    updatedAt: null,
    lastExecutedAt: null
  })
  selectedTags.value = []
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

// 解析标签数组
const parseTagsArray = (tags) => {
  if (!tags) return []
  try {
    return Array.isArray(tags) ? tags : JSON.parse(tags)
  } catch {
    return []
  }
}

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// 监听props变化
watch(() => props.rule, (newRule) => {
  if (newRule) {
    Object.assign(formData, newRule)
    selectedTags.value = parseTagsArray(newRule.tags)
  } else {
    resetForm()
  }
}, { immediate: true })

// 监听标签变化
watch(selectedTags, (newTags) => {
  formData.tags = JSON.stringify(newTags)
}, { deep: true })

// 格式化日期
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleString('zh-CN')
}

// 获取对话框标题
const getDialogTitle = () => {
  const titleMap = {
    create: '新建知识规则',
    edit: '编辑知识规则',
    view: '查看知识规则'
  }
  return titleMap[props.mode] || '知识规则'
}

// 保存规则
const handleSave = async () => {
  if (!formRef.value) return
  
  try {
    const valid = await formRef.value.validate()
    if (!valid) return
    
    saving.value = true
    
    const saveData = {
      ...formData,
      tags: JSON.stringify(selectedTags.value)
    }
    
    if (props.mode === 'create') {
      await createKnowledgeRule(saveData)
      ElMessage.success('规则创建成功')
    } else {
      await updateKnowledgeRule(saveData)
      ElMessage.success('规则更新成功')
    }
    
    emit('save', saveData)
    handleClose()
  } catch (error) {
    console.error('保存规则失败:', error)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  emit('update:modelValue', false)
  emit('cancel')
  nextTick(() => {
    resetForm()
  })
}
</script>

<style scoped>
.expression-help {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style> 