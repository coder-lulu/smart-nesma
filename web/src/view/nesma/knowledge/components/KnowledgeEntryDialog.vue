<template>
  <el-dialog
    v-model="dialogVisible"
    :title="getDialogTitle()"
    :width="800"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
      :disabled="mode === 'view'"
    >
      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="知识标题" prop="title">
            <el-input 
              v-model="formData.title" 
              placeholder="请输入知识标题"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="知识类别" prop="category">
            <el-select
              v-model="formData.category"
              placeholder="请选择知识类别"
              style="width: 100%"
            >
              <el-option label="NESMA标准" value="NESMA_STANDARD" />
              <el-option label="最佳实践" value="BEST_PRACTICE" />
              <el-option label="案例研究" value="CASE_STUDY" />
              <el-option label="业务规则" value="RULE" />
              <el-option label="用户采纳" value="用户采纳" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="知识领域" prop="domain">
            <el-input v-model="formData.domain" placeholder="请输入知识领域" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="置信度" prop="confidenceScore">
            <el-slider
              v-model="formData.confidenceScore"
              :min="0"
              :max="1"
              :step="0.1"
              :format-tooltip="formatTooltip"
              style="margin-top: 12px;"
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="版本" prop="version">
            <el-input v-model="formData.version" placeholder="请输入版本号" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="作者" prop="author">
            <el-input v-model="formData.author" placeholder="请输入作者" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="来源" prop="source">
            <el-input v-model="formData.source" placeholder="请输入来源" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="标签" prop="tags">
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
                v-for="tag in popularTags"
                :key="tag"
                :label="tag"
                :value="tag"
              />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="知识内容" prop="content">
            <el-input
              v-model="formData.content"
              type="textarea"
              :rows="8"
              placeholder="请输入知识内容"
              maxlength="5000"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <!-- 查看模式下显示额外信息 -->
      <template v-if="mode === 'view' && formData.id">
        <el-divider content-position="left">统计信息</el-divider>
        <el-row :gutter="20">
          <el-col :span="8">
            <el-statistic title="使用次数" :value="formData.usageCount || 0" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="创建时间" :value="formatDate(formData.createdAt)" value-style="font-size: 14px;" />
          </el-col>
          <el-col :span="8">
            <el-statistic title="更新时间" :value="formatDate(formData.updatedAt)" value-style="font-size: 14px;" />
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
import { createKnowledgeEntry, updateKnowledgeEntry, getPopularTags } from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  entry: {
    type: Object,
    default: null
  },
  mode: {
    type: String,
    default: 'create' // create, edit, view
  }
})

const emit = defineEmits(['update:modelValue', 'success', 'cancel'])

// 响应式数据
const formRef = ref(null)
const saving = ref(false)
const popularTags = ref([])
const selectedTags = ref([])

// 表单数据
const formData = reactive({
  id: null,
  title: '',
  content: '',
  category: '',
  domain: '',
  confidenceScore: 0.8,
  version: '1.0',
  author: '',
  source: '',
  usageCount: 0,
  createdAt: null,
  updatedAt: null
})

// 表单验证规则
const formRules = {
  title: [
    { required: true, message: '请输入知识标题', trigger: 'blur' },
    { min: 2, max: 200, message: '标题长度应在2-200个字符之间', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入知识内容', trigger: 'blur' },
    { min: 10, max: 5000, message: '内容长度应在10-5000个字符之间', trigger: 'blur' }
  ],
  category: [
    { required: true, message: '请选择知识类别', trigger: 'change' }
  ],
  domain: [
    { required: true, message: '请输入知识领域', trigger: 'blur' }
  ],
  confidenceScore: [
    { required: true, message: '请设置置信度', trigger: 'change' },
    { type: 'number', min: 0, max: 1, message: '置信度应在0-1之间', trigger: 'change' }
  ]
}

// 重置表单
const resetForm = (data = null) => {
  if (data) {
    Object.assign(formData, {
      id: data.id,
      title: data.title || '',
      content: data.content || '',
      category: data.category || '',
      domain: data.domain || '',
      confidenceScore: data.confidenceScore || 0.8,
      version: data.version || '1.0',
      author: data.author || '',
      source: data.source || '',
      usageCount: data.usageCount || 0,
      createdAt: data.createdAt,
      updatedAt: data.updatedAt
    })
    
    // 解析标签
    try {
      const tags = data.tags ? (Array.isArray(data.tags) ? data.tags : JSON.parse(data.tags)) : []
      selectedTags.value = tags
    } catch {
      selectedTags.value = []
    }
  } else {
    Object.assign(formData, {
      id: null,
      title: '',
      content: '',
      category: '',
      domain: '',
      confidenceScore: 0.8,
      version: '1.0',
      author: '',
      source: '',
      usageCount: 0,
      createdAt: null,
      updatedAt: null
    })
    selectedTags.value = []
  }
}

// 格式化提示
const formatTooltip = (value) => {
  return `${Math.round(value * 100)}%`
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleString('zh-CN')
}

// 获取对话框标题
const getDialogTitle = () => {
  const titleMap = {
    create: '新建知识条目',
    edit: '编辑知识条目',
    view: '查看知识条目'
  }
  return titleMap[props.mode] || '知识条目'
}

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 监听props变化
watch(() => props.entry, (newEntry) => {
  if (newEntry) {
    resetForm(newEntry)
  } else {
    resetForm()
  }
}, { immediate: true })

watch(() => props.modelValue, (visible) => {
  if (visible) {
    loadPopularTags()
    nextTick(() => {
      formRef.value?.clearValidate()
    })
  }
})

// 加载热门标签
const loadPopularTags = async () => {
  try {
    const res = await getPopularTags({ limit: 20 })
    popularTags.value = res.data || []
  } catch (error) {
    console.error('获取热门标签失败:', error)
  }
}

// 保存
const handleSave = async () => {
  try {
    await formRef.value?.validate()
    
    saving.value = true
    
    const data = {
      ...formData,
      tags: selectedTags.value
    }
    
    if (props.mode === 'create') {
      await createKnowledgeEntry(data)
    } else {
      await updateKnowledgeEntry(data.id, data)
    }
    
    emit('success')
    emit('update:modelValue', false)

  } catch (error) {
    if (error && error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    saving.value = false
  }
}

// 关闭
const handleClose = () => {
  emit('update:modelValue', false)
  emit('cancel')
}
</script>

<style scoped>
.dialog-footer {
  text-align: right;
}

:deep(.el-slider__runway) {
  margin: 16px 0;
}

:deep(.el-statistic__content) {
  font-size: 14px;
}

:deep(.el-textarea__inner) {
  resize: vertical;
}
</style> 