<template>
  <div class="template-form">
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      label-position="right"
    >
      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="模板名称" prop="name">
            <el-input
              v-model="form.name"
              placeholder="请输入模板名称"
              maxlength="100"
              show-word-limit
            />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="模板版本" prop="version">
            <el-input
              v-model="form.version"
              placeholder="如：v1.0.0"
              maxlength="20"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="模板类型" prop="type">
            <el-select
              v-model="form.type"
              placeholder="请选择模板类型"
              style="width: 100%"
              @change="handleTypeChange"
            >
              <el-option label="Word文档" value="word" />
              <el-option label="Excel表格" value="excel" />
              <el-option label="PDF文档" value="pdf" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="文档格式" prop="format">
            <el-select
              v-model="form.format"
              placeholder="请选择文档格式"
              style="width: 100%"
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

      <el-form-item label="模板描述" prop="description">
        <el-input
          v-model="form.description"
          type="textarea"
          :rows="3"
          placeholder="请输入模板描述"
          maxlength="500"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="模板配置">
        <el-card class="config-card">
          <template #header>
            <div class="config-header">
              <span>模板参数</span>
            </div>
          </template>
          
          <el-row :gutter="20">
            <el-col :span="12">
              <el-form-item label="是否默认" label-width="100px">
                <el-switch v-model="form.isDefault" />
                <div class="form-tip">
                  <small>设为默认模板后，创建文档时自动选择</small>
                </div>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item label="模板状态" label-width="100px">
                <el-switch 
                  v-model="form.status" 
                  active-text="启用"
                  inactive-text="停用"
                  active-value="active"
                  inactive-value="inactive"
                />
              </el-form-item>
            </el-col>
          </el-row>

          <el-form-item label="模板变量" label-width="100px">
            <div class="variables-section">
              <div class="variables-header">
                <span>自定义变量</span>
                <el-button size="small" type="primary" text @click="addVariable">
                  <el-icon><Plus /></el-icon>
                  添加变量
                </el-button>
              </div>
              
              <div v-if="form.variables.length === 0" class="empty-variables">
                <el-empty 
                  description="暂无自定义变量" 
                  :image-size="60"
                />
              </div>
              
              <div v-else class="variables-list">
                <div 
                  v-for="(variable, index) in form.variables" 
                  :key="index"
                  class="variable-item"
                >
                  <el-row :gutter="12">
                    <el-col :span="5">
                      <el-input
                        v-model="variable.name"
                        placeholder="变量名称"
                        maxlength="50"
                      />
                    </el-col>
                    <el-col :span="5">
                      <el-input
                        v-model="variable.label"
                        placeholder="显示标签"
                        maxlength="50"
                      />
                    </el-col>
                    <el-col :span="4">
                      <el-select
                        v-model="variable.type"
                        placeholder="数据类型"
                        style="width: 100%"
                      >
                        <el-option label="文本" value="string" />
                        <el-option label="数字" value="number" />
                        <el-option label="布尔" value="boolean" />
                        <el-option label="选择" value="select" />
                      </el-select>
                    </el-col>
                    <el-col :span="5">
                      <el-input
                        v-model="variable.defaultValue"
                        placeholder="默认值"
                        maxlength="100"
                      />
                    </el-col>
                    <el-col :span="4">
                      <el-input
                        v-model="variable.placeholder"
                        placeholder="提示文本"
                        maxlength="100"
                      />
                    </el-col>
                    <el-col :span="1">
                      <el-button 
                        size="small" 
                        type="danger" 
                        icon="Delete"
                        @click="removeVariable(index)"
                      />
                    </el-col>
                  </el-row>
                  
                  <!-- 选择类型的选项配置 -->
                  <div v-if="variable.type === 'select'" class="select-options">
                    <div class="options-header">
                      <small>选项配置：</small>
                      <el-button size="small" type="primary" text @click="addOption(index)">
                        添加选项
                      </el-button>
                    </div>
                    <div class="options-list">
                      <div 
                        v-for="(option, optionIndex) in variable.options" 
                        :key="optionIndex"
                        class="option-item"
                      >
                        <el-input
                          v-model="option.label"
                          placeholder="选项标签"
                          style="width: 150px; margin-right: 8px;"
                        />
                        <el-input
                          v-model="option.value"
                          placeholder="选项值"
                          style="width: 150px; margin-right: 8px;"
                        />
                        <el-button 
                          size="small" 
                          type="danger" 
                          icon="Delete"
                          @click="removeOption(index, optionIndex)"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-form-item>
        </el-card>
      </el-form-item>

      <el-form-item label="模板内容">
        <el-card class="content-card">
          <template #header>
            <div class="content-header">
              <span>模板内容编辑</span>
              <div class="content-actions">
                <el-button size="small" @click="handlePreview" icon="View">
                  预览
                </el-button>
                <el-button size="small" @click="handleImport" icon="Upload">
                  导入模板
                </el-button>
              </div>
            </div>
          </template>
          
          <div class="content-editor">
            <el-input
              v-model="form.content"
              type="textarea"
              :rows="12"
              placeholder="请输入模板内容，可使用变量如：{{变量名称}}"
              maxlength="10000"
              show-word-limit
            />
            <div class="editor-tips">
              <div class="tips-title">模板语法说明：</div>
              <ul class="tips-list">
                <li><code v-pre>{{变量名称}}</code> - 替换为用户输入的变量值</li>
                <li><code v-pre>{{项目名称}}</code> - 自动替换为项目名称</li>
                <li><code v-pre>{{需求列表}}</code> - 自动生成需求列表</li>
                <li><code v-pre>{{统计信息}}</code> - 自动生成统计图表</li>
                <li><code v-pre>{{生成时间}}</code> - 自动替换为文档生成时间</li>
              </ul>
            </div>
          </div>
        </el-card>
      </el-form-item>
    </el-form>

    <div class="form-actions">
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="primary" @click="handleSubmit" :loading="submitLoading">
        {{ dialogType === 'create' ? '创建模板' : '更新模板' }}
      </el-button>
    </div>

    <!-- 导入文件对话框 -->
    <input
      ref="fileInputRef"
      type="file"
      accept=".docx,.xlsx,.pdf,.txt"
      style="display: none"
      @change="handleFileImport"
    />

    <!-- 预览对话框 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="模板预览"
      width="80%"
      :close-on-click-modal="false"
    >
      <div class="template-preview">
        <pre class="preview-content">{{ form.content }}</pre>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, watch, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Delete, View, Upload } from '@element-plus/icons-vue'
import {
  createTemplate,
  updateTemplate
} from '@/api/nesma'

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
const fileInputRef = ref(null)
const submitLoading = ref(false)
const previewDialogVisible = ref(false)

// 表单数据
const form = reactive({
  name: '',
  description: '',
  type: 'word',
  format: 'requirement_spec',
  version: 'v1.0.0',
  status: 'active',
  isDefault: false,
  content: '',
  variables: []
})

// 表单验证规则
const rules = {
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' },
    { min: 2, max: 100, message: '长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择模板类型', trigger: 'change' }
  ],
  format: [
    { required: true, message: '请选择文档格式', trigger: 'change' }
  ],
  version: [
    { required: true, message: '请输入版本号', trigger: 'blur' }
  ],
  content: [
    { required: true, message: '请输入模板内容', trigger: 'blur' }
  ]
}

// 计算属性
const formatOptions = computed(() => {
  const allOptions = [
    { label: '需求规格说明书', value: 'requirement_spec' },
    { label: 'NESMA评估报告', value: 'nesma_report' },
    { label: '业务需求汇总表', value: 'business_summary' }
  ]

  // 根据模板类型过滤格式选项
  if (form.type === 'word') {
    return allOptions.filter(opt => opt.value !== 'business_summary')
  } else if (form.type === 'excel') {
    return allOptions.filter(opt => opt.value !== 'requirement_spec')
  } else if (form.type === 'pdf') {
    return allOptions
  }
  
  return allOptions
})

// 监听props变化
watch(() => props.formData, (newData) => {
  if (newData && Object.keys(newData).length > 0) {
    Object.assign(form, {
      ...newData,
      variables: newData.variables || []
    })
  }
}, { immediate: true, deep: true })

// 事件处理
const handleTypeChange = () => {
  // 重置格式
  form.format = formatOptions.value[0]?.value || ''
}

// 变量管理
const addVariable = () => {
  form.variables.push({
    name: '',
    label: '',
    type: 'string',
    defaultValue: '',
    placeholder: '',
    options: []
  })
}

const removeVariable = (index) => {
  form.variables.splice(index, 1)
}

const addOption = (variableIndex) => {
  if (!form.variables[variableIndex].options) {
    form.variables[variableIndex].options = []
  }
  form.variables[variableIndex].options.push({
    label: '',
    value: ''
  })
}

const removeOption = (variableIndex, optionIndex) => {
  form.variables[variableIndex].options.splice(optionIndex, 1)
}

// 预览
const handlePreview = () => {
  if (!form.content) {
    ElMessage.warning('请先输入模板内容')
    return
  }
  previewDialogVisible.value = true
}

// 导入模板
const handleImport = () => {
  fileInputRef.value?.click()
}

const handleFileImport = (event) => {
  const file = event.target.files[0]
  if (!file) return

  const reader = new FileReader()
  reader.onload = (e) => {
    try {
      const content = e.target.result
      form.content = content
      ElMessage.success('导入成功')
    } catch (error) {
      console.error('导入失败:', error)
      ElMessage.error('导入失败')
    }
  }
  reader.readAsText(file)
  
  // 清空文件输入
  event.target.value = ''
}

// 表单提交
const handleSubmit = async () => {
  if (!formRef.value) return
  
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return
  
  // 验证变量配置
  for (let i = 0; i < form.variables.length; i++) {
    const variable = form.variables[i]
    if (!variable.name || !variable.label) {
      ElMessage.error(`第 ${i + 1} 个变量的名称和标签不能为空`)
      return
    }
    if (variable.type === 'select' && (!variable.options || variable.options.length === 0)) {
      ElMessage.error(`第 ${i + 1} 个变量为选择类型，必须配置选项`)
      return
    }
  }
  
  submitLoading.value = true
  try {
    let res
    if (props.dialogType === 'create') {
      res = await createTemplate(form)
      if (res.code === 0) {
        ElMessage.success('模板创建成功')
      }
    } else {
      res = await updateTemplate({ ...form, ID: props.formData.ID })
      if (res.code === 0) {
        ElMessage.success('模板更新成功')
      }
    }
    
    if (res.code === 0) {
      emit('submit')
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
</script>

<style scoped>
.template-form {
  padding: 20px;
}

.config-card,
.content-card {
  width: 100%;
}

.config-header,
.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.content-actions {
  display: flex;
  gap: 8px;
}

.form-tip {
  margin-top: 4px;
}

.form-tip small {
  color: var(--el-text-color-secondary);
}

.variables-section {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 16px;
  background: var(--el-fill-color-light);
}

.variables-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  font-weight: 500;
}

.empty-variables {
  text-align: center;
  padding: 20px;
}

.variables-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.variable-item {
  border: 1px solid var(--el-border-color-light);
  border-radius: 6px;
  padding: 12px;
  background: white;
}

.select-options {
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px dashed var(--el-border-color-light);
}

.options-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.options-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.option-item {
  display: flex;
  align-items: center;
}

.content-editor {
  position: relative;
}

.editor-tips {
  margin-top: 16px;
  padding: 12px;
  background: var(--el-fill-color-light);
  border-radius: 6px;
  font-size: 12px;
}

.tips-title {
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}

.tips-list {
  margin: 0;
  padding-left: 20px;
  color: var(--el-text-color-secondary);
}

.tips-list li {
  margin-bottom: 4px;
}

.tips-list code {
  background: var(--el-color-primary-light-9);
  color: var(--el-color-primary);
  padding: 2px 4px;
  border-radius: 3px;
  font-size: 11px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  margin-top: 30px;
  padding-top: 20px;
  border-top: 1px solid var(--el-border-color);
}

.template-preview {
  max-height: 500px;
  overflow: auto;
}

.preview-content {
  background: var(--el-fill-color-light);
  padding: 16px;
  border-radius: 6px;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
  line-height: 1.5;
}
</style> 