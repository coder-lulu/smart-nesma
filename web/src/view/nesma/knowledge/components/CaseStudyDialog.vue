<template>
  <el-dialog
    v-model="dialogVisible"
    :title="getDialogTitle()"
    :width="1000"
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
          <el-form-item label="项目名称" prop="projectName">
            <el-input 
              v-model="formData.projectName" 
              placeholder="请输入项目名称"
              maxlength="200"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="12">
          <el-form-item label="应用领域" prop="domain">
            <el-select
              v-model="formData.domain"
              placeholder="请选择应用领域"
              style="width: 100%"
              filterable
              allow-create
            >
              <el-option label="金融保险" value="金融保险" />
              <el-option label="制造业" value="制造业" />
              <el-option label="电子商务" value="电子商务" />
              <el-option label="教育培训" value="教育培训" />
              <el-option label="医疗健康" value="医疗健康" />
              <el-option label="政府机关" value="政府机关" />
              <el-option label="物流运输" value="物流运输" />
              <el-option label="通信电信" value="通信电信" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="组织类型" prop="organizationType">
            <el-select
              v-model="formData.organizationType"
              placeholder="请选择组织类型"
              style="width: 100%"
            >
              <el-option label="大型企业" value="大型企业" />
              <el-option label="中小企业" value="中小企业" />
              <el-option label="政府机构" value="政府机构" />
              <el-option label="事业单位" value="事业单位" />
              <el-option label="非盈利组织" value="非盈利组织" />
            </el-select>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="项目规模" prop="projectScale">
            <el-select
              v-model="formData.projectScale"
              placeholder="请选择项目规模"
              style="width: 100%"
            >
              <el-option label="小型 (<100FP)" value="SMALL" />
              <el-option label="中型 (100-500FP)" value="MEDIUM" />
              <el-option label="大型 (>500FP)" value="LARGE" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="项目状态" prop="status">
            <el-select
              v-model="formData.status"
              placeholder="请选择项目状态"
              style="width: 100%"
            >
              <el-option label="已完成" value="COMPLETED" />
              <el-option label="进行中" value="IN_PROGRESS" />
              <el-option label="已取消" value="CANCELLED" />
              <el-option label="草稿" value="DRAFT" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="完成时间" prop="completionDate">
            <el-date-picker
              v-model="formData.completionDate"
              type="date"
              placeholder="请选择完成时间"
              style="width: 100%"
              value-format="YYYY-MM-DD"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="需求摘要" prop="requirementSummary">
            <el-input
              v-model="formData.requirementSummary"
              type="textarea"
              :rows="4"
              placeholder="请输入需求摘要"
              maxlength="1000"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-divider content-position="left">NESMA度量结果</el-divider>
      
      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="NESMA结果" prop="nesmaResult">
            <el-input-number
              v-model="formData.nesmaResult"
              :min="0"
              :max="10000"
              :precision="2"
              placeholder="功能点数"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="实际工作量" prop="actualEffort">
            <el-input-number
              v-model="formData.actualEffort"
              :min="0"
              :max="10000"
              :precision="2"
              placeholder="人天"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="生产率" prop="productivityRatio">
            <el-input-number
              v-model="formData.productivityRatio"
              :min="0"
              :max="100"
              :precision="4"
              placeholder="FP/人天"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="8">
          <el-form-item label="数据功能点" prop="dataFunctionPoints">
            <el-input-number
              v-model="formData.dataFunctionPoints"
              :min="0"
              :max="10000"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="事务功能点" prop="transactionFunctionPoints">
            <el-input-number
              v-model="formData.transactionFunctionPoints"
              :min="0"
              :max="10000"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="调整因子" prop="adjustmentFactor">
            <el-input-number
              v-model="formData.adjustmentFactor"
              :min="0"
              :max="5"
              :precision="2"
              style="width: 100%"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="技术特征" prop="technicalCharacteristics">
            <el-checkbox-group v-model="selectedTechnicalCharacteristics">
              <el-checkbox value="数据通讯">数据通讯</el-checkbox>
              <el-checkbox value="分布式数据处理">分布式数据处理</el-checkbox>
              <el-checkbox value="性能">性能</el-checkbox>
              <el-checkbox value="重负载配置">重负载配置</el-checkbox>
              <el-checkbox value="事务率">事务率</el-checkbox>
              <el-checkbox value="在线数据录入">在线数据录入</el-checkbox>
              <el-checkbox value="终端用户效率">终端用户效率</el-checkbox>
              <el-checkbox value="在线更新">在线更新</el-checkbox>
              <el-checkbox value="复杂处理">复杂处理</el-checkbox>
              <el-checkbox value="可重用性">可重用性</el-checkbox>
              <el-checkbox value="安装易用性">安装易用性</el-checkbox>
              <el-checkbox value="操作易用性">操作易用性</el-checkbox>
              <el-checkbox value="多站点">多站点</el-checkbox>
              <el-checkbox value="便于变更">便于变更</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="项目经验" prop="projectExperience">
            <el-input
              v-model="formData.projectExperience"
              type="textarea"
              :rows="4"
              placeholder="请描述项目实施过程中的经验和教训"
              maxlength="2000"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="最佳实践" prop="bestPractices">
            <el-input
              v-model="formData.bestPractices"
              type="textarea"
              :rows="4"
              placeholder="请描述项目中的最佳实践和可复用的经验"
              maxlength="2000"
              show-word-limit
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="20">
        <el-col :span="24">
          <el-form-item label="问题与挑战" prop="challengesAndSolutions">
            <el-input
              v-model="formData.challengesAndSolutions"
              type="textarea"
              :rows="4"
              placeholder="请描述项目中遇到的问题和解决方案"
              maxlength="2000"
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
            <el-statistic title="引用次数" :value="formData.referenceCount || 0" />
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
import { createCaseStudy, updateCaseStudy } from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  caseStudy: {
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
const selectedTechnicalCharacteristics = ref([])

// 表单数据
const formData = reactive({
  id: null,
  projectName: '',
  domain: '',
  organizationType: '',
  projectScale: '',
  status: 'DRAFT',
  completionDate: '',
  requirementSummary: '',
  nesmaResult: 0,
  actualEffort: 0,
  productivityRatio: 0,
  dataFunctionPoints: 0,
  transactionFunctionPoints: 0,
  adjustmentFactor: 1.0,
  technicalCharacteristics: '',
  projectExperience: '',
  bestPractices: '',
  challengesAndSolutions: '',
  referenceCount: 0,
  createdAt: null,
  updatedAt: null
})

// 表单验证规则
const formRules = {
  projectName: [
    { required: true, message: '请输入项目名称', trigger: 'blur' },
    { min: 2, max: 200, message: '项目名称长度应在2-200个字符', trigger: 'blur' }
  ],
  domain: [
    { required: true, message: '请选择应用领域', trigger: 'change' }
  ],
  organizationType: [
    { required: true, message: '请选择组织类型', trigger: 'change' }
  ],
  projectScale: [
    { required: true, message: '请选择项目规模', trigger: 'change' }
  ],
  status: [
    { required: true, message: '请选择项目状态', trigger: 'change' }
  ],
  requirementSummary: [
    { required: true, message: '请输入需求摘要', trigger: 'blur' },
    { min: 10, max: 1000, message: '需求摘要长度应在10-1000个字符', trigger: 'blur' }
  ],
  nesmaResult: [
    { required: true, message: '请输入NESMA结果', trigger: 'blur' },
    { type: 'number', min: 0, message: 'NESMA结果不能为负数', trigger: 'blur' }
  ],
  actualEffort: [
    { required: true, message: '请输入实际工作量', trigger: 'blur' },
    { type: 'number', min: 0, message: '实际工作量不能为负数', trigger: 'blur' }
  ]
}

// 重置表单
const resetForm = () => {
  Object.assign(formData, {
    id: null,
    projectName: '',
    domain: '',
    organizationType: '',
    projectScale: '',
    status: 'DRAFT',
    completionDate: '',
    requirementSummary: '',
    nesmaResult: 0,
    actualEffort: 0,
    productivityRatio: 0,
    dataFunctionPoints: 0,
    transactionFunctionPoints: 0,
    adjustmentFactor: 1.0,
    technicalCharacteristics: '',
    projectExperience: '',
    bestPractices: '',
    challengesAndSolutions: '',
    referenceCount: 0,
    createdAt: null,
    updatedAt: null
  })
  selectedTechnicalCharacteristics.value = []
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

// 解析数组字段
const parseArrayField = (field) => {
  if (!field) return []
  try {
    return Array.isArray(field) ? field : JSON.parse(field)
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
watch(() => props.caseStudy, (newCaseStudy) => {
  if (newCaseStudy) {
    Object.assign(formData, newCaseStudy)
    selectedTechnicalCharacteristics.value = parseArrayField(newCaseStudy.technicalCharacteristics)
  } else {
    resetForm()
  }
}, { immediate: true })

// 监听技术特征变化
watch(selectedTechnicalCharacteristics, (newValue) => {
  formData.technicalCharacteristics = JSON.stringify(newValue)
}, { deep: true })

// 监听生产率计算
watch([() => formData.nesmaResult, () => formData.actualEffort], ([nesmaResult, actualEffort]) => {
  if (nesmaResult > 0 && actualEffort > 0) {
    formData.productivityRatio = parseFloat((nesmaResult / actualEffort).toFixed(4))
  }
})

// 格式化日期
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleString('zh-CN')
}

// 获取对话框标题
const getDialogTitle = () => {
  const titleMap = {
    create: '新建案例研究',
    edit: '编辑案例研究',
    view: '查看案例研究'
  }
  return titleMap[props.mode] || '案例研究'
}

// 保存案例
const handleSave = async () => {
  if (!formRef.value) return
  
  try {
    const valid = await formRef.value.validate()
    if (!valid) return
    
    saving.value = true
    
    const saveData = {
      ...formData,
      technicalCharacteristics: JSON.stringify(selectedTechnicalCharacteristics.value)
    }
    
    if (props.mode === 'create') {
      await createCaseStudy(saveData)
      ElMessage.success('案例创建成功')
    } else {
      await updateCaseStudy(saveData)
      ElMessage.success('案例更新成功')
    }
    
    emit('save', saveData)
    handleClose()
  } catch (error) {
    console.error('保存案例失败:', error)
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
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-checkbox-group) {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

:deep(.el-checkbox) {
  margin-right: 0;
}
</style> 