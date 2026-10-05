<template>
  <el-dialog
    :model-value="modelValue"
    title="Excel导入"
    width="600px"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="import-dialog">
      <el-steps :active="currentStep" finish-status="success">
        <el-step title="选择文件" description="上传Excel文件" />
        <el-step title="配置参数" description="设置导入选项" />
        <el-step title="导入结果" description="查看导入结果" />
      </el-steps>

      <!-- 步骤1: 文件选择 -->
      <div v-if="currentStep === 0" class="step-content">
        <el-upload
          ref="uploadRef"
          class="upload-demo"
          drag
          action="#"
          :auto-upload="false"
          :on-change="handleFileChange"
          :before-upload="() => false"
          accept=".xlsx,.xls"
        >
          <el-icon class="el-icon--upload"><upload-filled /></el-icon>
          <div class="el-upload__text">
            将Excel文件拖到此处，或<em>点击上传</em>
          </div>
          <template #tip>
            <div class="el-upload__tip">
              只能上传 .xlsx/.xls 文件
            </div>
          </template>
        </el-upload>

        <div class="template-download">
          <el-button type="text" @click="downloadTemplate">
            <el-icon><download /></el-icon>
            下载导入模板
          </el-button>
        </div>
      </div>

      <!-- 步骤2: 配置参数 -->
      <div v-if="currentStep === 1" class="step-content">
        <el-form :model="importConfig" label-width="120px">
          <el-form-item label="导入周期" required>
            <el-select 
              v-model="importConfig.cycleId" 
              placeholder="请选择导入周期"
              :loading="loading.cycles"
              @change="handleCycleChange"
            >
              <el-option 
                v-for="item in cycleOptions" 
                :key="item.ID" 
                :label="item.name" 
                :value="item.ID" 
              />
            </el-select>
          </el-form-item>
          
          <el-form-item label="导入版本" required>
            <el-input 
              v-model="importConfig.versionName" 
              placeholder="自动生成版本号"
              disabled 
            />
            <div class="version-info">
              <small>基于当前最大版本号 {{ currentMaxVersion }} 自动生成新版本</small>
            </div>
          </el-form-item>
          
          <el-form-item label="导入批次">
            <el-input v-model="importConfig.importBatch" placeholder="自动生成" disabled />
          </el-form-item>
          
          <el-form-item label="导入来源">
            <el-input v-model="importConfig.importSource" placeholder="请输入导入来源" disabled />
          </el-form-item>
          
          <el-form-item label="覆盖策略">
            <el-radio-group v-model="importConfig.overwriteStrategy">
              <el-radio value="skip">跳过重复项</el-radio>
              <el-radio value="update">更新重复项</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-form>
      </div>

      <!-- 步骤3: 导入结果 -->
      <div v-if="currentStep === 2" class="step-content">
        <div v-if="importResult" class="import-result">
          <el-result
            :icon="importResult.success ? 'success' : 'warning'"
            :title="importResult.success ? '导入成功' : '导入部分成功'"
            :sub-title="`共处理 ${importResult.totalRows} 行，成功 ${importResult.successRows} 行，失败 ${importResult.failedRows} 行`"
          />
          
          <div v-if="importResult.errors?.length" class="error-list">
            <h4>错误详情：</h4>
            <ul>
              <li v-for="(error, index) in importResult.errors" :key="index">
                {{ error }}
              </li>
            </ul>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <span class="dialog-footer">
        <el-button v-if="currentStep > 0" @click="prevStep">上一步</el-button>
        <el-button @click="handleClose">取消</el-button>
        <el-button 
          v-if="currentStep < 2" 
          type="primary" 
          @click="nextStep"
          :disabled="!canNextStep"
        >
          下一步
        </el-button>
        <el-button 
          v-if="currentStep === 1" 
          type="primary" 
          @click="handleImport"
          :loading="importing"
        >
          开始导入
        </el-button>
        <el-button 
          v-if="currentStep === 2" 
          type="primary" 
          @click="handleComplete"
        >
          完成
        </el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { UploadFilled, Download } from '@element-plus/icons-vue'
import { importFromExcel } from '@/api/nesma/requirement'
import { getProjectCycles } from '@/api/projectCycle'
import { getProjectMaxVersion } from '@/api/nesma/requirement'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  projectId: {
    type: [Number, String],
    default: null
  },
  cycleId: {
    type: [Number, String],
    default: null
  },
  versionId: {
    type: [Number, String],
    default: null
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'import-completed'])

// 响应式数据
const currentStep = ref(0)
const selectedFile = ref(null)
const importing = ref(false)
const importResult = ref(null)
const loading = reactive({
  cycles: false,
  version: false
})

// 周期选项
const cycleOptions = ref([])
const currentMaxVersion = ref('v1.0')

const importConfig = reactive({
  cycleId: null,
  versionName: '',
  importBatch: `import_${Date.now()}`,
  importSource: 'Excel导入',
  overwriteStrategy: 'skip'
})

// 计算属性
const canNextStep = computed(() => {
  if (currentStep.value === 0) return !!selectedFile.value
  if (currentStep.value === 1) return !!importConfig.cycleId && !!importConfig.versionName
  return false
})

// 监听对话框打开
watch(() => props.modelValue, (newVal) => {
  if (newVal) {
    // 对话框打开时加载数据
    loadProjectCycles()
  }
})

// 方法
const handleFileChange = (file) => {
  selectedFile.value = file
}

const downloadTemplate = () => {
  // 这里可以实现模板下载功能
  ElMessage.success('模板下载功能开发中...')
}

const nextStep = () => {
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

// 加载项目周期列表
const loadProjectCycles = async () => {
  if (!props.projectId) {
    ElMessage.warning('项目ID不能为空')
    return
  }
  
  try {
    loading.cycles = true
    const response = await getProjectCycles(props.projectId)
    
    if (response.code === 0) {
      cycleOptions.value = response.data || []
      
      // 如果有默认周期ID，设置它
      if (props.cycleId && cycleOptions.value.length > 0) {
        const defaultCycle = cycleOptions.value.find(c => c.ID == props.cycleId)
        if (defaultCycle) {
          importConfig.cycleId = defaultCycle.ID
          await handleCycleChange(defaultCycle.ID)
        }
      }
    } else {
      ElMessage.error('加载项目周期失败：' + response.msg)
    }
  } catch (error) {
    console.error('加载项目周期失败:', error)
    ElMessage.error('加载项目周期失败：' + error.message)
  } finally {
    loading.cycles = false
  }
}

// 处理周期变化
const handleCycleChange = async (cycleId) => {
  if (!cycleId) {
    currentMaxVersion.value = 'v1.0'
    importConfig.versionName = ''
    return
  }
  
  try {
    loading.version = true
    
    // 获取项目最大版本号
    const response = await getProjectMaxVersion(props.projectId)
    
    if (response.code === 0) {
      if (response.data.maxVersion == 0) {
        currentMaxVersion.value = 'v1.0'
        importConfig.versionName = 'v1.0'
        return
      } else {
        const maxVersion = response.data.maxVersion || 1
        currentMaxVersion.value = `v${maxVersion}.0`
        // 生成新版本号（在最大版本基础上+0.1）
      const newVersion = maxVersion + 0.1
      importConfig.versionName = `v${newVersion.toFixed(1)}`
      }
    } else {
      ElMessage.error('获取版本信息失败：' + response.msg)
      currentMaxVersion.value = 'v1.0'
      importConfig.versionName = 'v1.1'
    }
  } catch (error) {
    console.error('获取版本信息失败:', error)
    ElMessage.error('获取版本信息失败：' + error.message)
    currentMaxVersion.value = 'v1.0'
    importConfig.versionName = 'v1.1'
  } finally {
    loading.version = false
  }
}

const handleImport = async () => {
  if (!selectedFile.value) {
    ElMessage.warning('请先选择文件')
    return
  }
  
  if (!props.projectId || !importConfig.cycleId || !importConfig.versionName) {
    ElMessage.warning('项目、周期或版本信息缺失')
    return
  }
  
  try {
    importing.value = true
    
    // 构造FormData
    const formData = new FormData()
    formData.append('file', selectedFile.value.raw)
    formData.append('projectId', props.projectId)
    formData.append('cycleId', importConfig.cycleId)
    formData.append('versionName', importConfig.versionName)
    formData.append('importBatch', importConfig.importBatch)
    formData.append('importSource', importConfig.importSource)
    formData.append('overwriteStrategy', importConfig.overwriteStrategy)
    
    console.log('开始导入，参数:', {
      projectId: props.projectId,
      cycleId: importConfig.cycleId,
      versionName: importConfig.versionName,
      importBatch: importConfig.importBatch,
      importSource: importConfig.importSource,
      overwriteStrategy: importConfig.overwriteStrategy
    })
    
    const response = await importFromExcel(formData)
    
    if (response.code === 0) {
      importResult.value = {
        success: true,
        totalRows: response.data.totalRows || 0,
        successRows: response.data.successRows || 0,
        failedRows: response.data.failedRows || 0,
        errors: response.data.errors || []
      }
      
      currentStep.value = 2
      ElMessage.success('导入完成')
    } else {
      throw new Error(response.msg || '导入失败')
    }
  } catch (error) {
    console.error('导入失败:', error)
    ElMessage.error('导入失败: ' + error.message)
    
    importResult.value = {
      success: false,
      totalRows: 0,
      successRows: 0,
      failedRows: 0,
      errors: [error.message]
    }
    currentStep.value = 2
  } finally {
    importing.value = false
  }
}

const handleComplete = () => {
  emit('import-completed', importResult.value)
  handleClose()
}

const handleClose = () => {
  // 重置状态
  currentStep.value = 0
  selectedFile.value = null
  importing.value = false
  importResult.value = null
  cycleOptions.value = []
  currentMaxVersion.value = 'v1.0'
  
  importConfig.cycleId = null
  importConfig.versionName = ''
  importConfig.importBatch = `import_${Date.now()}`
  importConfig.importSource = 'Excel导入'
  importConfig.overwriteStrategy = 'skip'
  
  emit('update:modelValue', false)
}

// 组件挂载时加载数据
onMounted(() => {
  if (props.modelValue) {
    loadProjectCycles()
    
  }
})
</script>

<style lang="scss" scoped>
.import-dialog {
  .step-content {
    margin: 30px 0;
    min-height: 200px;
  }
  
  .template-download {
    margin-top: 20px;
    text-align: center;
  }
  
  .version-info {
    margin-top: 5px;
    color: #909399;
  }
  
  .import-result {
    .error-list {
      margin-top: 20px;
      
      h4 {
        color: #e6a23c;
        margin-bottom: 10px;
      }
      
      ul {
        background: #fdf6ec;
        border: 1px solid #f5dab1;
        border-radius: 4px;
        padding: 10px 20px;
        max-height: 200px;
        overflow-y: auto;
        
        li {
          color: #e6a23c;
          line-height: 1.5;
          margin-bottom: 5px;
        }
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>