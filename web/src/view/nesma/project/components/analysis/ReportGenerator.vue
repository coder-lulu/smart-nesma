<template>
  <div class="enhanced-report-generator">
    <!-- 报告生成选项对话框 -->
    <el-dialog
      v-model="showOptions"
      title="智能报告生成"
      width="800px"
      :close-on-click-modal="false"
      class="report-generator-dialog"
    >
      <div class="report-config-container">
        <!-- 左侧：报告类型选择 -->
        <div class="report-type-section">
          <h3>报告类型</h3>
          <el-card class="type-card">
            <el-radio-group v-model="reportConfig.type" @change="onReportTypeChange">
              <el-radio-button label="analysis">分析报告</el-radio-button>
              <el-radio-button label="summary">摘要报告</el-radio-button>
              <el-radio-button label="detailed">详细报告</el-radio-button>
              <el-radio-button label="custom">自定义报告</el-radio-button>
            </el-radio-group>
          </el-card>
        </div>

        <!-- 右侧：格式和选项 -->
        <div class="report-options-section">
          <el-tabs v-model="activeTab" type="card">
            <el-tab-pane label="基础配置" name="basic">
              <el-form :model="reportConfig" label-width="120px">
                <el-form-item label="报告格式">
                  <el-select v-model="reportConfig.format" placeholder="选择报告格式">
                    <el-option-group label="文档格式">
                      <el-option label="PDF报告" value="pdf" />
                      <el-option label="Word文档" value="word" />
                      <el-option label="HTML网页" value="html" />
                    </el-option-group>
                    <el-option-group label="数据格式">
                      <el-option label="Excel数据" value="excel" />
                      <el-option label="JSON数据" value="json" />
                      <el-option label="CSV数据" value="csv" />
                    </el-option-group>
                  </el-select>
                </el-form-item>
                
                <el-form-item label="报告标题">
                  <el-input v-model="reportConfig.title" placeholder="输入报告标题" />
                </el-form-item>
                
                <el-form-item label="报告模板">
                  <el-select v-model="reportConfig.templateId" placeholder="选择报告模板">
                    <el-option 
                      v-for="template in filteredTemplates" 
                      :key="template.id || template.ID"
                      :label="template.name || template.title" 
                      :value="template.id || template.ID" 
                    />
                  </el-select>
                  <el-button type="text" @click="openTemplateManager">管理模板</el-button>
                </el-form-item>
                
                <el-form-item label="详细程度">
                  <el-slider
                    v-model="reportConfig.detailLevel"
                    :marks="detailLevelMarks"
                    :min="1"
                    :max="5"
                    show-stops
                  />
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <el-tab-pane label="内容选择" name="content">
              <div class="content-selection">
                <h4>报告章节</h4>
                <el-tree
                  ref="sectionTree"
                  :data="sectionTreeData"
                  node-key="id"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  :default-checked-keys="reportConfig.sections"
                  @check="onSectionCheck"
                  class="section-tree"
                >
                  <template #default="{ node, data }">
                    <span class="tree-node">
                      <el-icon><Document /></el-icon>
                      <span>{{ data.name }}</span>
                      <span class="node-description">{{ data.description }}</span>
                    </span>
                  </template>
                </el-tree>
              </div>
            </el-tab-pane>

            <el-tab-pane label="高级选项" name="advanced">
              <el-form :model="reportConfig.advanced" label-width="120px">
                <el-form-item label="图表样式">
                  <el-select v-model="reportConfig.advanced.chartStyle">
                    <el-option label="简洁风格" value="simple" />
                    <el-option label="商务风格" value="business" />
                    <el-option label="学术风格" value="academic" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="数据精度">
                  <el-input-number v-model="reportConfig.advanced.precision" :min="0" :max="6" />
                </el-form-item>
                
                <el-form-item label="语言">
                  <el-select v-model="reportConfig.advanced.language">
                    <el-option label="中文" value="zh-CN" />
                    <el-option label="英文" value="en-US" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="导出选项">
                  <el-checkbox-group v-model="reportConfig.advanced.exportOptions">
                    <el-checkbox label="includeRawData">包含原始数据</el-checkbox>
                    <el-checkbox label="includeMetadata">包含元数据</el-checkbox>
                    <el-checkbox label="enableSignature">启用数字签名</el-checkbox>
                    <el-checkbox label="enableEncryption">启用加密</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
              </el-form>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showOptions = false">取消</el-button>
          <el-button type="primary" @click="previewReport">预览</el-button>
          <el-button type="success" @click="generateReport">生成报告</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 报告预览对话框 -->
    <el-dialog
      v-model="showPreview"
      title="报告预览"
      width="90%"
      :close-on-click-modal="false"
      class="report-preview-dialog"
    >
      <div class="preview-container">
        <div class="preview-toolbar">
          <el-button-group>
            <el-button :type="previewMode === 'html' ? 'primary' : 'default'" @click="previewMode = 'html'">
              <el-icon><Monitor /></el-icon>
              HTML预览
            </el-button>
            <el-button :type="previewMode === 'pdf' ? 'primary' : 'default'" @click="previewMode = 'pdf'">
              <el-icon><Document /></el-icon>
              PDF预览
            </el-button>
            <el-button :type="previewMode === 'mobile' ? 'primary' : 'default'" @click="previewMode = 'mobile'">
              <el-icon><Cellphone /></el-icon>
              移动端预览
            </el-button>
          </el-button-group>
          
          <el-button type="primary" @click="downloadPreview">
            <el-icon><Download /></el-icon>
            下载预览
          </el-button>
        </div>
        
        <div :class="['preview-content', previewMode]">
          <div v-if="previewLoading" class="preview-loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>正在生成预览...</span>
          </div>
          <div v-else-if="previewContent" v-html="previewContent" class="preview-html"></div>
          <div v-else class="preview-error">预览生成失败</div>
        </div>
      </div>
    </el-dialog>

    <!-- 报告生成进度对话框 -->
    <el-dialog
      v-model="showProgress"
      title="报告生成进度"
      width="500px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="progress-container">
        <div class="progress-info">
          <div class="current-step">
            <el-icon class="step-icon"><Loading /></el-icon>
            <span>{{ currentStep }}</span>
          </div>
          <el-progress
            :percentage="progressPercentage"
            :stroke-width="12"
            :color="progressColor"
            class="progress-bar"
          />
          <div class="progress-details">
            <p>已完成：{{ completedSteps }}/{{ totalSteps }}</p>
            <p>预计剩余时间：{{ estimatedTime }}</p>
          </div>
        </div>
        
        <div class="progress-log">
          <h4>生成日志</h4>
          <div class="log-content">
            <div v-for="log in generationLogs" :key="log.id" class="log-item">
              <span class="log-time">{{ log.timestamp }}</span>
              <span :class="['log-level', log.level]">{{ log.level }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <el-button @click="cancelGeneration" :disabled="generationCompleted">取消生成</el-button>
        <el-button v-if="generationCompleted" type="primary" @click="downloadReport">下载报告</el-button>
      </template>
    </el-dialog>

    <!-- 模板管理对话框 -->
    <ReportTemplateManager
      v-model="showTemplateManager"
      @template-updated="refreshTemplates"
    />

    <!-- 批量生成对话框 -->
    <BatchReportGenerator
      v-model="showBatchGenerator"
      :projects="availableProjects"
      @batch-started="onBatchStarted"
    />
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage, ElLoading, ElMessageBox } from 'element-plus'
import { 
  Document, 
  Monitor, 
  Cellphone, 
  Download, 
  Loading 
} from '@element-plus/icons-vue'
import ReportTemplateManager from './ReportTemplateManager.vue'
import BatchReportGenerator from './BatchReportGenerator.vue'
import { 
  exportUnifiedAnalysisReport,
  getReportTemplates,
  generateReportPreview,
  getReportGenerationProgress,
  cancelReportGeneration 
} from '@/api/nesma'

const props = defineProps({
  taskId: [String, Number],
  project: Object,
  analysisData: Object
})

const emit = defineEmits(['close', 'download', 'report-generated'])

// 响应式数据
const showOptions = ref(false)
const showPreview = ref(false)
const showProgress = ref(false)
const showTemplateManager = ref(false)
const showBatchGenerator = ref(false)
const activeTab = ref('basic')
const previewMode = ref('html')
const previewLoading = ref(false)
const previewContent = ref('')
const currentGenerationId = ref(null)

// 报告配置
const reportConfig = reactive({
  type: 'analysis',
  format: 'pdf',
  title: '',
  templateId: null,
  detailLevel: 3,
  sections: [],
  advanced: {
    chartStyle: 'business',
    precision: 2,
    language: 'zh-CN',
    exportOptions: ['includeRawData']
  }
})

// 报告模板
const reportTemplates = ref([])

// 过滤有效的模板（排除null值）
const filteredTemplates = computed(() => {
  return reportTemplates.value.filter(template => 
    template && 
    (template.id || template.ID) && 
    (template.name || template.title)
  )
})

// 章节树数据
const sectionTreeData = ref([
  {
    id: 'executive_summary',
    name: '执行摘要',
    description: '高层管理概述',
    children: [
      { id: 'key_findings', name: '关键发现', description: '主要分析结果' },
      { id: 'recommendations', name: '建议事项', description: '改进建议' }
    ]
  },
  {
    id: 'analysis_results',
    name: '分析结果',
    description: '详细分析数据',
    children: [
      { id: 'requirement_analysis', name: '需求分析', description: '需求详细分析' },
      { id: 'function_points', name: '功能点计算', description: 'NESMA功能点分析' },
      { id: 'complexity_analysis', name: '复杂度分析', description: '复杂度评估' }
    ]
  },
  {
    id: 'quality_assessment',
    name: '质量评估',
    description: '质量指标评估',
    children: [
      { id: 'quality_metrics', name: '质量指标', description: '各项质量指标' },
      { id: 'risk_analysis', name: '风险分析', description: '潜在风险识别' }
    ]
  },
  {
    id: 'technical_details',
    name: '技术细节',
    description: '技术实现细节',
    children: [
      { id: 'architecture', name: '架构设计', description: '系统架构' },
      { id: 'implementation', name: '实现方案', description: '技术实现' }
    ]
  },
  {
    id: 'appendices',
    name: '附录',
    description: '补充材料',
    children: [
      { id: 'raw_data', name: '原始数据', description: '分析原始数据' },
      { id: 'references', name: '参考文献', description: '引用资料' }
    ]
  }
])

// 进度相关数据
const progressPercentage = ref(0)
const currentStep = ref('')
const completedSteps = ref(0)
const totalSteps = ref(0)
const estimatedTime = ref('')
const generationLogs = ref([])
const generationCompleted = ref(false)
const availableProjects = ref([])

// 计算属性
const defaultTitle = computed(() => {
  if (!props.project) return '分析报告'
  return `${reportConfig.type === 'analysis' ? '分析报告' : '摘要报告'}_${props.project.name}_${new Date().toLocaleDateString()}`
})

const detailLevelMarks = computed(() => ({
  1: '概要',
  2: '简要',
  3: '标准',
  4: '详细',
  5: '完整'
}))

const progressColor = computed(() => {
  if (progressPercentage.value < 30) return '#f56c6c'
  if (progressPercentage.value < 70) return '#e6a23c'
  return '#67c23a'
})

// 方法
const openReportOptions = () => {
  reportConfig.title = defaultTitle.value
  showOptions.value = true
}

const onReportTypeChange = (type) => {
  // 根据报告类型调整默认章节
  const defaultSections = {
    analysis: ['executive_summary', 'analysis_results', 'quality_assessment'],
    summary: ['executive_summary', 'key_findings'],
    detailed: ['executive_summary', 'analysis_results', 'quality_assessment', 'technical_details'],
    custom: []
  }
  reportConfig.sections = defaultSections[type] || []
}

const onSectionCheck = (data, checked) => {
  const checkedKeys = checked.checkedKeys
  const halfCheckedKeys = checked.halfCheckedKeys
  reportConfig.sections = [...checkedKeys, ...halfCheckedKeys]
}

const previewReport = async () => {
  if (!props.taskId) {
    ElMessage.error('任务ID不能为空')
    return
  }

  previewLoading.value = true
  showPreview.value = true
  
  try {
    const response = await generateReportPreview({
      task_id: props.taskId,
      config: reportConfig
    })
    
    previewContent.value = response.data.html_content
  } catch (error) {
    ElMessage.error('生成预览失败: ' + error.message)
    previewContent.value = ''
  } finally {
    previewLoading.value = false
  }
}

const generateReport = async () => {
  if (!props.taskId) {
    ElMessage.error('任务ID不能为空')
    return
  }

  showOptions.value = false
  showProgress.value = true
  
  // 重置进度状态
  progressPercentage.value = 0
  currentStep.value = '准备生成报告...'
  completedSteps.value = 0
  totalSteps.value = 8
  generationLogs.value = []
  generationCompleted.value = false
  
  try {
    // 启动报告生成 - 修复数据格式匹配后端期望
    const response = await exportUnifiedAnalysisReport({
      task_id: props.taskId,
      export_format: reportConfig.format, // 修复：确保格式字段存在
      include_sections: reportConfig.sections || [],
      language: reportConfig.advanced.language || 'zh-CN',
      template: String(reportConfig.templateId || 'default'), // 修复：转换为字符串
      custom_fields: []
    })
    
    // 处理不同的响应格式
    let blob = null, filename = ''
    
    if (response.data.download_url) {
      // 如果有下载链接，直接跳转下载
      window.open(response.data.download_url, '_blank')
      filename = response.data.file_name || `analysis_report_${props.taskId}.${reportConfig.format}`
    } else if (response.data.file_content) {
      // 如果有文件内容（Base64编码）
      const binaryString = atob(response.data.file_content)
      const bytes = new Uint8Array(binaryString.length)
      for (let i = 0; i < binaryString.length; i++) {
        bytes[i] = binaryString.charCodeAt(i)
      }
      blob = new Blob([bytes], { type: response.data.content_type || 'application/octet-stream' })
      filename = response.data.file_name || `analysis_report_${props.taskId}.${reportConfig.format}`
    } else if (response.data.html_content) {
      // HTML内容，直接预览
      previewContent.value = response.data.html_content
      showPreview.value = true
      showOptions.value = false
      return
    } else if (reportConfig.format === 'json') {
      // JSON格式，直接下载
      blob = new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json' })
      filename = `analysis_report_${props.taskId}.json`
    } else {
      // 其他情况，将整个响应作为JSON下载
      blob = new Blob([JSON.stringify(response.data, null, 2)], { type: 'application/json' })
      filename = `analysis_report_${props.taskId}.json`
    }
    
    // 统一下载逻辑
    if (blob) {
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.setAttribute('download', filename)
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.URL.revokeObjectURL(url)
      
      emit('download', { format: reportConfig.format, filename })
    }
    
    ElMessage.success(`报告生成成功！格式: ${reportConfig.format.toUpperCase()}`)
    showOptions.value = false
    
  } catch (error) {
    ElMessage.error('启动报告生成失败: ' + error.message)
    showProgress.value = false
  }
}

const pollGenerationProgress = async () => {
  if (!currentGenerationId.value) return
  
  const interval = setInterval(async () => {
    try {
      const response = await getReportGenerationProgress(currentGenerationId.value)
      const progress = response.data
      
      progressPercentage.value = progress.percentage
      currentStep.value = progress.current_step
      completedSteps.value = progress.completed_steps
      totalSteps.value = progress.total_steps
      estimatedTime.value = progress.estimated_time
      
      // 更新日志
      if (progress.logs && progress.logs.length > 0) {
        generationLogs.value = progress.logs
      }
      
      if (progress.status === 'completed') {
        clearInterval(interval)
        generationCompleted.value = true
        currentStep.value = '报告生成完成'
        ElMessage.success('报告生成完成!')
        emit('report-generated', {
          generationId: currentGenerationId.value,
          config: reportConfig
        })
      } else if (progress.status === 'failed') {
        clearInterval(interval)
        ElMessage.error('报告生成失败')
        showProgress.value = false
      }
      
    } catch (error) {
      clearInterval(interval)
      ElMessage.error('获取生成进度失败')
      showProgress.value = false
    }
  }, 2000)
}

const cancelGeneration = async () => {
  if (!currentGenerationId.value) return
  
  try {
    await ElMessageBox.confirm('确定要取消报告生成吗？', '确认取消', {
      type: 'warning'
    })
    
    await cancelReportGeneration(currentGenerationId.value)
    showProgress.value = false
    ElMessage.success('报告生成已取消')
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消生成失败')
    }
  }
}

const downloadReport = () => {
  if (!currentGenerationId.value) return
  
  // 创建下载链接
  const downloadUrl = `/api/v1/nesma/report/download/${currentGenerationId.value}`
  const link = document.createElement('a')
  link.href = downloadUrl
  link.download = `${reportConfig.title}.${reportConfig.format}`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  
  showProgress.value = false
  emit('download', {
    generationId: currentGenerationId.value,
    format: reportConfig.format
  })
}

const downloadPreview = () => {
  if (!previewContent.value) return
  
  const blob = new Blob([previewContent.value], { type: 'text/html' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${reportConfig.title}_preview.html`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
}

const openTemplateManager = () => {
  showTemplateManager.value = true
}

const refreshTemplates = async () => {
  try {
    const response = await getReportTemplates()
    console.log('模板API响应:', response)
    
    // 确保响应数据是数组
    let templates = []
    if (response && response.data) {
      if (Array.isArray(response.data)) {
        templates = response.data
      } else if (response.data.list && Array.isArray(response.data.list)) {
        templates = response.data.list
      } else if (response.data.templates && Array.isArray(response.data.templates)) {
        templates = response.data.templates
      } else {
        console.warn('模板数据格式不匹配:', response.data)
        // 如果没有模板数据，使用默认模板
        templates = [
          { id: 1, name: '标准分析报告', type: 'analysis', format: 'pdf' },
          { id: 2, name: '详细需求文档', type: 'requirement', format: 'word' },
          { id: 3, name: 'NESMA评估报告', type: 'nesma', format: 'excel' }
        ]
      }
    } else {
      console.warn('模板API响应为空，使用默认模板')
      // 使用默认模板
      templates = [
        { id: 1, name: '标准分析报告', type: 'analysis', format: 'pdf' },
        { id: 2, name: '详细需求文档', type: 'requirement', format: 'word' },
        { id: 3, name: 'NESMA评估报告', type: 'nesma', format: 'excel' }
      ]
    }
    
    reportTemplates.value = templates
    console.log('设置模板数据:', templates)
    
  } catch (error) {
    console.error('刷新模板失败:', error)
    ElMessage.warning('获取模板列表失败，使用默认模板')
    
    // 失败时使用默认模板
    reportTemplates.value = [
      { id: 1, name: '标准分析报告', type: 'analysis', format: 'pdf' },
      { id: 2, name: '详细需求文档', type: 'requirement', format: 'word' },
      { id: 3, name: 'NESMA评估报告', type: 'nesma', format: 'excel' }
    ]
  }
}

const onBatchStarted = (batchInfo) => {
  ElMessage.success(`批量生成已启动，共 ${batchInfo.total} 个报告`)
}

// 生命周期
onMounted(() => {
  refreshTemplates()
})

// 暴露方法给父组件
defineExpose({
  openReportOptions,
  generateReport
})
</script>

<style lang="scss" scoped>
.enhanced-report-generator {
  .report-generator-dialog {
    .report-config-container {
      display: flex;
      gap: 20px;
      min-height: 500px;
      
      .report-type-section {
        flex: 0 0 200px;
        
        h3 {
          margin-bottom: 16px;
          color: #303133;
          font-size: 16px;
        }
        
        .type-card {
          .el-radio-group {
            display: flex;
            flex-direction: column;
            gap: 8px;
            
            .el-radio-button {
              width: 100%;
              
              :deep(.el-radio-button__inner) {
                width: 100%;
                text-align: left;
              }
            }
          }
        }
      }
      
      .report-options-section {
        flex: 1;
        
        .el-tabs {
          height: 100%;
          
          .content-selection {
            .section-tree {
              max-height: 300px;
              overflow-y: auto;
              
              .tree-node {
                display: flex;
                align-items: center;
                gap: 8px;
                
                .node-description {
                  color: #909399;
                  font-size: 12px;
                  margin-left: auto;
                }
              }
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
  }
  
  .report-preview-dialog {
    .preview-container {
      .preview-toolbar {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
        padding: 12px;
        background: #f5f7fa;
        border-radius: 6px;
      }
      
      .preview-content {
        min-height: 600px;
        border: 1px solid #dcdfe6;
        border-radius: 6px;
        overflow: hidden;
        
        &.html {
          .preview-html {
            padding: 20px;
            background: white;
          }
        }
        
        &.pdf {
          background: #525659;
          .preview-html {
            background: white;
            margin: 20px;
            padding: 20px;
            box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
          }
        }
        
        &.mobile {
          max-width: 375px;
          margin: 0 auto;
          
          .preview-html {
            font-size: 14px;
            line-height: 1.6;
          }
        }
        
        .preview-loading {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 300px;
          color: #909399;
          
          .el-icon {
            font-size: 32px;
            margin-bottom: 16px;
          }
        }
        
        .preview-error {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 300px;
          color: #f56c6c;
          font-size: 16px;
        }
      }
    }
  }
  
  .progress-container {
    .progress-info {
      margin-bottom: 24px;
      
      .current-step {
        display: flex;
        align-items: center;
        gap: 8px;
        margin-bottom: 12px;
        font-size: 14px;
        color: #606266;
        
        .step-icon {
          color: #409eff;
        }
      }
      
      .progress-bar {
        margin-bottom: 12px;
      }
      
      .progress-details {
        display: flex;
        justify-content: space-between;
        font-size: 12px;
        color: #909399;
      }
    }
    
    .progress-log {
      h4 {
        margin-bottom: 12px;
        color: #303133;
        font-size: 14px;
      }
      
      .log-content {
        max-height: 200px;
        overflow-y: auto;
        padding: 12px;
        background: #f5f7fa;
        border-radius: 4px;
        font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
        font-size: 12px;
        line-height: 1.5;
        
        .log-item {
          display: flex;
          gap: 8px;
          margin-bottom: 4px;
          
          .log-time {
            color: #909399;
            flex-shrink: 0;
          }
          
          .log-level {
            font-weight: bold;
            flex-shrink: 0;
            
            &.INFO {
              color: #409eff;
            }
            
            &.WARN {
              color: #e6a23c;
            }
            
            &.ERROR {
              color: #f56c6c;
            }
          }
          
          .log-message {
            flex: 1;
            color: #303133;
          }
        }
      }
    }
  }
}
</style>