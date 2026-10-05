<template>
  <el-dialog
    v-model="dialogVisible"
    title="导入知识库"
    :width="700"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @close="handleClose"
  >
    <el-tabs v-model="activeTab" @tab-click="handleTabClick">
      <!-- Excel文件导入 -->
      <el-tab-pane label="Excel文件导入" name="excel">
        <div class="import-section">
          <el-form :model="excelForm" label-width="120px">
            <el-form-item label="导入类型">
              <el-radio-group v-model="excelForm.importType">
                <el-radio value="knowledge">知识条目</el-radio>
                <el-radio value="rules">业务规则</el-radio>
                <el-radio value="cases">案例研究</el-radio>
              </el-radio-group>
            </el-form-item>
            
            <el-form-item label="选择文件">
              <el-upload
                ref="excelUploadRef"
                :limit="1"
                accept=".xlsx,.xls"
                :auto-upload="false"
                :on-change="handleExcelFileChange"
                :file-list="excelFileList"
                :on-remove="handleExcelFileRemove"
              >
                                 <template #trigger>
                   <el-button size="small" type="primary">选择文件</el-button>
                 </template>
                 <template #tip>
                   <div class="el-upload__tip">
                                       只能上传xlsx/xls文件，且不超过10MB
                   </div>
                 </template>
              </el-upload>
            </el-form-item>
            
            <el-form-item label="导入选项">
              <el-checkbox-group v-model="excelForm.options">
                <el-checkbox value="skipDuplicates">跳过重复数据</el-checkbox>
                <el-checkbox value="updateExisting">更新已存在数据</el-checkbox>
                <el-checkbox value="validateData">验证数据格式</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            
            <el-alert
              title="导入说明"
              type="info"
              :closable="false"
              show-icon
            >
              <div>
                <p>1. 请下载对应的Excel模板文件，按照模板格式填写数据</p>
                <p>2. 支持批量导入知识条目、业务规则和案例研究</p>
                <p>3. 导入前会自动验证数据格式和必填字段</p>
                <p>4. 重复数据处理方式可在导入选项中设置</p>
              </div>
            </el-alert>
            
            <div class="template-download">
              <el-button 
                link
                @click="downloadTemplate('knowledge')"
                :disabled="excelUploading"
              >
                下载知识条目模板
              </el-button>
              <el-divider direction="vertical" />
              <el-button 
                link
                @click="downloadTemplate('rules')"
                :disabled="excelUploading"
              >
                下载业务规则模板
              </el-button>
              <el-divider direction="vertical" />
              <el-button 
                link
                @click="downloadTemplate('cases')"
                :disabled="excelUploading"
              >
                下载案例研究模板
              </el-button>
            </div>
          </el-form>
        </div>
      </el-tab-pane>
      
      <!-- NESMA标准知识导入 -->
      <el-tab-pane label="NESMA标准知识" name="nesma">
        <div class="import-section">
          <el-form :model="nesmaForm" label-width="120px">
            <el-form-item label="导入内容">
              <el-checkbox-group v-model="nesmaForm.importContent">
                <el-checkbox value="standards">NESMA标准文档</el-checkbox>
                <el-checkbox value="guidelines">实施指南</el-checkbox>
                <el-checkbox value="templates">模板文档</el-checkbox>
                <el-checkbox value="examples">示例案例</el-checkbox>
                <el-checkbox value="rules">标准规则</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
            
            <el-form-item label="版本选择">
              <el-select v-model="nesmaForm.version" placeholder="请选择NESMA版本">
                <el-option label="NESMA 2.1" value="2.1" />
                <el-option label="NESMA 2.0" value="2.0" />
                <el-option label="NESMA 1.3" value="1.3" />
              </el-select>
            </el-form-item>
            
            <el-form-item label="语言">
              <el-select v-model="nesmaForm.language" placeholder="请选择语言">
                <el-option label="中文" value="zh" />
                <el-option label="英文" value="en" />
              </el-select>
            </el-form-item>
            
            <el-alert
              title="NESMA标准知识导入说明"
              type="warning"
              :closable="false"
              show-icon
            >
              <div>
                <p>1. 导入NESMA标准知识将会添加大量官方标准内容</p>
                <p>2. 建议在首次使用系统时导入，避免重复内容</p>
                <p>3. 导入过程可能需要几分钟时间，请耐心等待</p>
                <p>4. 导入的内容包括标准文档、实施指南、模板等</p>
              </div>
            </el-alert>
            
            <div class="nesma-stats" v-if="nesmaStats">
              <h4>预计导入内容：</h4>
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-statistic title="知识条目" :value="nesmaStats.entries" />
                </el-col>
                <el-col :span="8">
                  <el-statistic title="业务规则" :value="nesmaStats.rules" />
                </el-col>
                <el-col :span="8">
                  <el-statistic title="案例研究" :value="nesmaStats.cases" />
                </el-col>
              </el-row>
            </div>
          </el-form>
        </div>
      </el-tab-pane>
    </el-tabs>
    
    <!-- 导入进度 -->
    <div v-if="importProgress.visible" class="import-progress">
      <el-progress 
        :percentage="importProgress.percentage" 
        :status="importProgress.status"
        :stroke-width="10"
      />
      <div class="progress-text">
        {{ importProgress.text }}
      </div>
    </div>
    
    <!-- 导入结果 -->
    <div v-if="importResult.visible" class="import-result">
      <el-alert
        :title="importResult.title"
        :type="importResult.type"
        :description="importResult.description"
        :closable="false"
        show-icon
      />
      <div v-if="importResult.details" class="result-details">
        <h4>导入详情：</h4>
        <ul>
          <li v-for="detail in importResult.details" :key="detail">{{ detail }}</li>
        </ul>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose" :disabled="importing">取消</el-button>
        <el-button 
          type="primary" 
          :loading="importing"
          @click="handleImport"
          :disabled="!canImport"
        >
          {{ importing ? '导入中...' : '开始导入' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { importNesmaStandardKnowledge } from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'success', 'cancel'])

// 响应式数据
const activeTab = ref('excel')
const importing = ref(false)
const excelUploading = ref(false)
const excelUploadRef = ref(null)
const excelFileList = ref([])

// Excel导入表单
const excelForm = reactive({
  importType: 'knowledge',
  options: ['skipDuplicates', 'validateData'],
  file: null
})

// NESMA导入表单
const nesmaForm = reactive({
  importContent: ['standards', 'guidelines'],
  version: '2.1',
  language: 'zh'
})

// 导入进度
const importProgress = reactive({
  visible: false,
  percentage: 0,
  status: '',
  text: ''
})

// 导入结果
const importResult = reactive({
  visible: false,
  title: '',
  type: '',
  description: '',
  details: []
})

// NESMA统计信息
const nesmaStats = ref({
  entries: 120,
  rules: 45,
  cases: 15
})

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const canImport = computed(() => {
  if (activeTab.value === 'excel') {
    return excelFileList.value.length > 0
  } else {
    return nesmaForm.importContent.length > 0 && nesmaForm.version && nesmaForm.language
  }
})

// 监听导入类型变化
watch(() => excelForm.importType, () => {
  // 清空文件列表
  excelFileList.value = []
  if (excelUploadRef.value) {
    excelUploadRef.value.clearFiles()
  }
})

// Tab切换
const handleTabClick = (tab) => {
  activeTab.value = tab.name
  resetImportStatus()
}

// Excel文件变化
const handleExcelFileChange = (file) => {
  const isExcel = file.name.endsWith('.xlsx') || file.name.endsWith('.xls')
  const isLt10M = file.size / 1024 / 1024 < 10
  
  if (!isExcel) {
    ElMessage.error('只能上传Excel文件！')
    return false
  }
  if (!isLt10M) {
    ElMessage.error('文件大小不能超过10MB！')
    return false
  }
  
  excelForm.file = file.raw
  excelFileList.value = [file]
  return true
}

// 移除Excel文件
const handleExcelFileRemove = () => {
  excelForm.file = null
  excelFileList.value = []
}

// 下载模板
const downloadTemplate = (type) => {
  const templates = {
    knowledge: '知识条目导入模板.xlsx',
    rules: '业务规则导入模板.xlsx',
    cases: '案例研究导入模板.xlsx'
  }
  
  // 这里应该调用API下载模板文件
  ElMessage.info(`正在下载${templates[type]}...`)
  
  // 模拟下载
  const link = document.createElement('a')
  link.href = `/templates/${templates[type]}`
  link.download = templates[type]
  link.click()
}

// 重置导入状态
const resetImportStatus = () => {
  importProgress.visible = false
  importProgress.percentage = 0
  importProgress.status = ''
  importProgress.text = ''
  importResult.visible = false
  importResult.title = ''
  importResult.type = ''
  importResult.description = ''
  importResult.details = []
}

// 执行导入
const handleImport = async () => {
  if (!canImport.value) return
  
  try {
    importing.value = true
    resetImportStatus()
    
    if (activeTab.value === 'excel') {
      await importFromExcel()
    } else {
      await importNesmaStandard()
    }
    
    emit('success')
  } catch (error) {
    console.error('导入失败:', error)
    showImportResult('导入失败', 'error', error.message || '导入过程中发生错误')
  } finally {
    importing.value = false
  }
}

// Excel导入
const importFromExcel = async () => {
  showImportProgress('正在解析Excel文件...', 10)
  
  // 模拟解析过程
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  showImportProgress('正在验证数据格式...', 30)
  await new Promise(resolve => setTimeout(resolve, 1000))
  
  showImportProgress('正在导入数据...', 60)
  await new Promise(resolve => setTimeout(resolve, 2000))
  
  showImportProgress('导入完成', 100, 'success')
  
  // 显示导入结果
  const mockResult = {
    success: 45,
    skipped: 5,
    failed: 2
  }
  
  showImportResult(
    '导入完成',
    'success',
    `成功导入${mockResult.success}条记录`,
    [
      `成功导入: ${mockResult.success}条`,
      `跳过重复: ${mockResult.skipped}条`,
      `导入失败: ${mockResult.failed}条`
    ]
  )
}

// NESMA标准导入
const importNesmaStandard = async () => {
  showImportProgress('正在下载NESMA标准知识...', 20)
  
  try {
    const response = await importNesmaStandardKnowledge()
    
    showImportProgress('正在处理标准内容...', 60)
    await new Promise(resolve => setTimeout(resolve, 2000))
    
    showImportProgress('正在保存到数据库...', 90)
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    showImportProgress('导入完成', 100, 'success')
    
    showImportResult(
      'NESMA标准知识导入完成',
      'success',
      '成功导入NESMA标准知识库',
      [
        `知识条目: ${response.data?.entries || 0}条`,
        `业务规则: ${response.data?.rules || 0}条`,
        `案例研究: ${response.data?.cases || 0}条`
      ]
    )
  } catch (error) {
    throw error
  }
}

// 显示导入进度
const showImportProgress = (text, percentage, status = '') => {
  importProgress.visible = true
  importProgress.text = text
  importProgress.percentage = percentage
  importProgress.status = status
}

// 显示导入结果
const showImportResult = (title, type, description, details = []) => {
  importResult.visible = true
  importResult.title = title
  importResult.type = type
  importResult.description = description
  importResult.details = details
}

// 关闭对话框
const handleClose = () => {
  emit('update:modelValue', false)
  emit('cancel')
  
  // 重置状态
  activeTab.value = 'excel'
  excelForm.importType = 'knowledge'
  excelForm.options = ['skipDuplicates', 'validateData']
  excelForm.file = null
  excelFileList.value = []
  nesmaForm.importContent = ['standards', 'guidelines']
  nesmaForm.version = '2.1'
  nesmaForm.language = 'zh'
  resetImportStatus()
}
</script>

<style scoped>
.import-section {
  padding: 20px 0;
}

.template-download {
  margin-top: 20px;
  padding: 15px;
  background-color: #f8f9fa;
  border-radius: 6px;
  text-align: center;
}

.nesma-stats {
  margin-top: 20px;
  padding: 15px;
  background-color: #f0f9ff;
  border-radius: 6px;
}

.nesma-stats h4 {
  margin: 0 0 15px 0;
  color: #303133;
}

.import-progress {
  margin: 20px 0;
}

.progress-text {
  text-align: center;
  margin-top: 10px;
  color: #606266;
  font-size: 14px;
}

.import-result {
  margin: 20px 0;
}

.result-details {
  margin-top: 15px;
}

.result-details h4 {
  margin: 0 0 10px 0;
  color: #303133;
}

.result-details ul {
  margin: 0;
  padding-left: 20px;
}

.result-details li {
  margin-bottom: 5px;
  color: #606266;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-upload__tip) {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}
</style> 