<template>
  <el-dialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    :title="isEdit ? '编辑功能点' : '新建功能点'"
    width="60%"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      size="default"
    >
      <!-- 基本信息 -->
      <el-card class="form-section">
        <template #header>
          <span>基本信息</span>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="功能名称" prop="functionName">
              <el-input
                v-model="form.functionName"
                placeholder="请输入功能名称"
                clearable
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="功能类型" prop="functionType">
              <el-select
                v-model="form.functionType"
                placeholder="请选择功能类型"
                style="width: 100%"
                @change="onFunctionTypeChange"
              >
                <el-option
                  v-for="type in functionTypes"
                  :key="type.value"
                  :label="type.label"
                  :value="type.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="识别方法" prop="identificationMethod">
              <el-select
                v-model="form.identificationMethod"
                placeholder="请选择识别方法"
                style="width: 100%"
              >
                <el-option
                  v-for="method in identificationMethods"
                  :key="method.value"
                  :label="method.label"
                  :value="method.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="验证状态" prop="validationStatus">
              <el-select
                v-model="form.validationStatus"
                placeholder="请选择验证状态"
                style="width: 100%"
              >
                <el-option
                  v-for="status in validationStatuses"
                  :key="status.value"
                  :label="status.label"
                  :value="status.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-form-item label="功能描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入功能描述"
          />
        </el-form-item>
      </el-card>

      <!-- 技术参数 -->
      <el-card class="form-section">
        <template #header>
          <span>技术参数</span>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <el-form-item label="数据元素类型" prop="dataElementTypes">
              <el-input-number
                v-model="form.dataElementTypes"
                :min="0"
                :max="999"
                style="width: 100%"
                @change="calculateComplexity"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="文件类型引用" prop="fileTypeReferences">
              <el-input-number
                v-model="form.fileTypeReferences"
                :min="0"
                :max="999"
                style="width: 100%"
                @change="calculateComplexity"
              />
            </el-form-item>
          </el-col>
          <el-col :span="8">
            <el-form-item label="记录元素类型" prop="recordElementTypes">
              <el-input-number
                v-model="form.recordElementTypes"
                :min="0"
                :max="999"
                style="width: 100%"
                @change="calculateComplexity"
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="复杂度等级" prop="complexityLevel">
              <el-select
                v-model="form.complexityLevel"
                placeholder="请选择复杂度等级"
                style="width: 100%"
                @change="calculateFunctionPoints"
              >
                <el-option
                  v-for="level in complexityLevels"
                  :key="level.value"
                  :label="level.label"
                  :value="level.value"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="功能点值" prop="functionPointValue">
              <el-input-number
                v-model="form.functionPointValue"
                :min="0"
                :max="999"
                :precision="1"
                style="width: 100%"
                :disabled="autoCalculate"
              />
            </el-form-item>
          </el-col>
        </el-row>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <el-form-item label="置信度" prop="confidenceLevel">
              <el-slider
                v-model="form.confidenceLevel"
                :min="0"
                :max="1"
                :step="0.1"
                :format-tooltip="formatConfidenceTooltip"
                show-input
                :show-input-controls="false"
                input-size="small"
              />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item>
              <el-checkbox
                v-model="autoCalculate"
                label="自动计算功能点值"
                @change="onAutoCalculateChange"
              />
            </el-form-item>
          </el-col>
        </el-row>
      </el-card>

      <!-- 业务规则 -->
      <el-card class="form-section">
        <template #header>
          <span>业务规则</span>
        </template>
        
        <el-form-item label="业务规则" prop="businessRules">
          <el-input
            v-model="form.businessRules"
            type="textarea"
            :rows="3"
            placeholder="请输入业务规则"
          />
        </el-form-item>
        
        <el-form-item label="调整因子" prop="adjustmentFactors">
          <el-input
            v-model="form.adjustmentFactors"
            type="textarea"
            :rows="2"
            placeholder="请输入调整因子说明"
          />
        </el-form-item>
      </el-card>

      <!-- 验证信息 -->
      <el-card class="form-section">
        <template #header>
          <span>验证信息</span>
        </template>
        
        <el-form-item label="验证备注" prop="validationNotes">
          <el-input
            v-model="form.validationNotes"
            type="textarea"
            :rows="2"
            placeholder="请输入验证备注"
          />
        </el-form-item>
        
        <el-form-item label="评审意见" prop="reviewComments">
          <el-input
            v-model="form.reviewComments"
            type="textarea"
            :rows="2"
            placeholder="请输入评审意见"
          />
        </el-form-item>
      </el-card>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="$emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" @click="save" :loading="saving">保存</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { createFunctionPoint, updateFunctionPoint } from '@/api/nesmaEvaluation'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  functionPoint: {
    type: Object,
    default: () => ({})
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'success'])

// 响应式数据
const formRef = ref(null)
const saving = ref(false)
const autoCalculate = ref(true)

// 计算属性
const isEdit = computed(() => {
  return props.functionPoint && props.functionPoint.id
})

// 表单数据
const form = reactive({
  id: null,
  functionName: '',
  functionType: '',
  identificationMethod: 'manual',
  validationStatus: 'pending',
  description: '',
  dataElementTypes: 0,
  fileTypeReferences: 0,
  recordElementTypes: 0,
  complexityLevel: 'Average',
  functionPointValue: 0,
  confidenceLevel: 0.8,
  businessRules: '',
  adjustmentFactors: '',
  validationNotes: '',
  reviewComments: ''
})

// 选项数据
const functionTypes = [
  { value: 'ILF', label: '内部逻辑文件 (ILF)' },
  { value: 'EIF', label: '外部接口文件 (EIF)' },
  { value: 'EI', label: '外部输入 (EI)' },
  { value: 'EO', label: '外部输出 (EO)' },
  { value: 'EQ', label: '外部查询 (EQ)' }
]

const identificationMethods = [
  { value: 'manual', label: '手工识别' },
  { value: 'automated', label: '自动识别' },
  { value: 'hybrid', label: '混合识别' }
]

const validationStatuses = [
  { value: 'pending', label: '待验证' },
  { value: 'approved', label: '已通过' },
  { value: 'rejected', label: '已拒绝' },
  { value: 'reviewed', label: '已评审' }
]

const complexityLevels = [
  { value: 'Low', label: '低复杂度' },
  { value: 'Average', label: '中等复杂度' },
  { value: 'High', label: '高复杂度' }
]

// 功能点值映射表
const functionPointValues = {
  'ILF': { Low: 7, Average: 10, High: 15 },
  'EIF': { Low: 5, Average: 7, High: 10 },
  'EI': { Low: 3, Average: 4, High: 6 },
  'EO': { Low: 4, Average: 5, High: 7 },
  'EQ': { Low: 3, Average: 4, High: 6 }
}

// 复杂度判断规则
const complexityRules = {
  'ILF': (det, ret) => {
    if (ret <= 1) return det <= 19 ? 'Low' : det <= 50 ? 'Average' : 'High'
    if (ret <= 5) return det <= 20 ? 'Low' : det <= 50 ? 'Average' : 'High'
    return det <= 20 ? 'Average' : 'High'
  },
  'EIF': (det, ret) => {
    if (ret <= 1) return det <= 19 ? 'Low' : det <= 50 ? 'Average' : 'High'
    if (ret <= 5) return det <= 20 ? 'Low' : det <= 50 ? 'Average' : 'High'
    return det <= 20 ? 'Average' : 'High'
  },
  'EI': (det, ftr) => {
    if (ftr <= 1) return det <= 4 ? 'Low' : det <= 15 ? 'Average' : 'High'
    if (ftr <= 2) return det <= 5 ? 'Low' : det <= 15 ? 'Average' : 'High'
    return det <= 5 ? 'Average' : 'High'
  },
  'EO': (det, ftr) => {
    if (ftr <= 1) return det <= 5 ? 'Low' : det <= 19 ? 'Average' : 'High'
    if (ftr <= 2) return det <= 5 ? 'Low' : det <= 19 ? 'Average' : 'High'
    return det <= 5 ? 'Average' : 'High'
  },
  'EQ': (det, ftr) => {
    if (ftr <= 1) return det <= 5 ? 'Low' : det <= 19 ? 'Average' : 'High'
    if (ftr <= 2) return det <= 5 ? 'Low' : det <= 19 ? 'Average' : 'High'
    return det <= 5 ? 'Average' : 'High'
  }
}

// 表单验证规则
const rules = {
  functionName: [
    { required: true, message: '请输入功能名称', trigger: 'blur' },
    { min: 2, max: 100, message: '功能名称长度在 2 到 100 个字符', trigger: 'blur' }
  ],
  functionType: [
    { required: true, message: '请选择功能类型', trigger: 'change' }
  ],
  identificationMethod: [
    { required: true, message: '请选择识别方法', trigger: 'change' }
  ],
  validationStatus: [
    { required: true, message: '请选择验证状态', trigger: 'change' }
  ],
  dataElementTypes: [
    { required: true, message: '请输入数据元素类型数量', trigger: 'blur' },
    { type: 'number', min: 0, message: '数据元素类型数量不能小于0', trigger: 'blur' }
  ],
  fileTypeReferences: [
    { required: true, message: '请输入文件类型引用数量', trigger: 'blur' },
    { type: 'number', min: 0, message: '文件类型引用数量不能小于0', trigger: 'blur' }
  ],
  recordElementTypes: [
    { required: true, message: '请输入记录元素类型数量', trigger: 'blur' },
    { type: 'number', min: 0, message: '记录元素类型数量不能小于0', trigger: 'blur' }
  ],
  complexityLevel: [
    { required: true, message: '请选择复杂度等级', trigger: 'change' }
  ],
  functionPointValue: [
    { required: true, message: '请输入功能点值', trigger: 'blur' },
    { type: 'number', min: 0, message: '功能点值不能小于0', trigger: 'blur' }
  ],
  confidenceLevel: [
    { required: true, message: '请设置置信度', trigger: 'blur' },
    { type: 'number', min: 0, max: 1, message: '置信度范围在0到1之间', trigger: 'blur' }
  ]
}

// 监听弹窗打开
watch(() => props.modelValue, (visible) => {
  if (visible) {
    resetForm()
    if (isEdit.value) {
      loadFunctionPoint()
    }
  }
})

// 方法
const resetForm = () => {
  Object.assign(form, {
    id: null,
    functionName: '',
    functionType: '',
    identificationMethod: 'manual',
    validationStatus: 'pending',
    description: '',
    dataElementTypes: 0,
    fileTypeReferences: 0,
    recordElementTypes: 0,
    complexityLevel: 'Average',
    functionPointValue: 0,
    confidenceLevel: 0.8,
    businessRules: '',
    adjustmentFactors: '',
    validationNotes: '',
    reviewComments: ''
  })
  autoCalculate.value = true
}

const loadFunctionPoint = () => {
  if (!props.functionPoint) return
  
  Object.assign(form, {
    id: props.functionPoint.id,
    functionName: props.functionPoint.functionName || '',
    functionType: props.functionPoint.functionType || '',
    identificationMethod: props.functionPoint.identificationMethod || 'manual',
    validationStatus: props.functionPoint.validationStatus || 'pending',
    description: props.functionPoint.description || '',
    dataElementTypes: props.functionPoint.dataElementTypes || 0,
    fileTypeReferences: props.functionPoint.fileTypeReferences || 0,
    recordElementTypes: props.functionPoint.recordElementTypes || 0,
    complexityLevel: props.functionPoint.complexityLevel || 'Average',
    functionPointValue: props.functionPoint.functionPointValue || 0,
    confidenceLevel: props.functionPoint.confidenceLevel || 0.8,
    businessRules: props.functionPoint.businessRules || '',
    adjustmentFactors: props.functionPoint.adjustmentFactors || '',
    validationNotes: props.functionPoint.validationNotes || '',
    reviewComments: props.functionPoint.reviewComments || ''
  })
}

const onFunctionTypeChange = () => {
  if (autoCalculate.value) {
    calculateComplexity()
  }
}

const calculateComplexity = () => {
  if (!form.functionType) return
  
  const det = form.dataElementTypes || 0
  const ftr = form.fileTypeReferences || 0
  const ret = form.recordElementTypes || 0
  
  const rule = complexityRules[form.functionType]
  if (rule) {
    const complexity = ['ILF', 'EIF'].includes(form.functionType) 
      ? rule(det, ret) 
      : rule(det, ftr)
    form.complexityLevel = complexity
    
    if (autoCalculate.value) {
      calculateFunctionPoints()
    }
  }
}

const calculateFunctionPoints = () => {
  if (!form.functionType || !form.complexityLevel) return
  
  const values = functionPointValues[form.functionType]
  if (values && values[form.complexityLevel]) {
    form.functionPointValue = values[form.complexityLevel]
  }
}

const onAutoCalculateChange = () => {
  if (autoCalculate.value) {
    calculateComplexity()
  }
}

const formatConfidenceTooltip = (value) => {
  return `${(value * 100).toFixed(0)}%`
}

const save = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    saving.value = true
    
    const formData = { ...form }
    
    if (isEdit.value) {
      await updateFunctionPoint(formData)
      ElMessage.success('功能点更新成功')
    } else {
      await createFunctionPoint(formData)
      ElMessage.success('功能点创建成功')
    }
    
    emit('success')
    emit('update:modelValue', false)
  } catch (error) {
    console.error('保存功能点失败:', error)
    ElMessage.error('保存失败，请重试')
  } finally {
    saving.value = false
  }
}
</script>

<style scoped>
.form-section {
  margin-bottom: 20px;
}

.form-section:last-child {
  margin-bottom: 0;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-slider) {
  margin-right: 20px;
}

:deep(.el-slider__input) {
  width: 80px;
}
</style> 