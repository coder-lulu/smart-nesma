<template>
  <div class="enhanced-document-generator">
    <!-- 文档生成选项对话框 -->
    <el-dialog
      v-model="showOptions"
      title="智能文档导出"
      width="1000px"
      :close-on-click-modal="false"
      class="document-generator-dialog"
    >
      <div class="document-config-container">
        <!-- 左侧：导出类型选择 -->
        <div class="export-type-section">
          <h3>导出类型</h3>
          <el-card class="type-card">
            <el-radio-group v-model="exportConfig.type" @change="onExportTypeChange">
              <el-radio-button label="comprehensive">综合文档</el-radio-button>
              <el-radio-button label="requirement_only">需求文档</el-radio-button>
              <el-radio-button label="analysis_only">分析报告</el-radio-button>
            </el-radio-group>
          </el-card>
          
          <!-- 快速预设 -->
          <div class="quick-presets">
            <h4>快速预设</h4>
            <el-card class="preset-card">
              <div class="preset-item" @click="applyPreset('management')">
                <el-icon><Avatar /></el-icon>
                <span>管理报告</span>
              </div>
              <div class="preset-item" @click="applyPreset('technical')">
                <el-icon><Tools /></el-icon>
                <span>技术文档</span>
              </div>
              <div class="preset-item" @click="applyPreset('audit')">
                <el-icon><Document /></el-icon>
                <span>审计报告</span>
              </div>
            </el-card>
          </div>
        </div>

        <!-- 右侧：详细配置 -->
        <div class="export-options-section">
          <el-tabs v-model="activeTab" type="card">
            <!-- 基础配置 -->
            <el-tab-pane label="基础配置" name="basic">
              <el-form :model="exportConfig" label-width="120px">
                <el-form-item label="文档格式">
                  <el-select v-model="exportConfig.format" placeholder="选择文档格式">
                    <el-option-group label="文档格式">
                      <el-option label="Word文档 (.docx)" value="word">
                        <span>Word文档 (.docx)</span>
                        <span class="option-desc">适合编辑和审阅</span>
                      </el-option>
                      <el-option label="PDF文档 (.pdf)" value="pdf">
                        <span>PDF文档 (.pdf)</span>
                        <span class="option-desc">适合分发和存档</span>
                      </el-option>
                      <el-option label="HTML网页 (.html)" value="html">
                        <span>HTML网页 (.html)</span>
                        <span class="option-desc">适合在线查看</span>
                      </el-option>
                    </el-option-group>
                    <el-option-group label="数据格式">
                      <el-option label="Excel表格 (.xlsx)" value="excel">
                        <span>Excel表格 (.xlsx)</span>
                        <span class="option-desc">适合数据分析</span>
                      </el-option>
                      <el-option label="JSON数据 (.json)" value="json">
                        <span>JSON数据 (.json)</span>
                        <span class="option-desc">适合程序处理</span>
                      </el-option>
                    </el-option-group>
                  </el-select>
                </el-form-item>
                
                <el-form-item label="文档标题">
                  <el-input 
                    v-model="exportConfig.title" 
                    placeholder="输入文档标题"
                    :suffix-icon="Edit"
                  />
                </el-form-item>
                
                <el-form-item label="文档模板">
                  <el-select v-model="exportConfig.templateId" placeholder="选择文档模板">
                    <el-option 
                      v-for="template in filteredTemplates" 
                      :key="template.id || template.ID"
                      :label="template.name || template.title" 
                      :value="template.id || template.ID"
                    >
                      <div class="template-option">
                        <span class="template-name">{{ template.name || template.title }}</span>
                        <el-tag size="small" :type="getTemplateTypeColor(template.type)">{{ template.type }}</el-tag>
                      </div>
                    </el-option>
                  </el-select>
                  <el-button type="text" @click="openTemplateManager" class="template-manage-btn">
                    <el-icon><Setting /></el-icon>
                    管理模板
                  </el-button>
                </el-form-item>
                
                <el-form-item label="详细程度">
                  <el-slider
                    v-model="exportConfig.detailLevel"
                    :marks="detailLevelMarks"
                    :min="1"
                    :max="5"
                    show-stops
                    show-tooltip
                  />
                  <div class="detail-level-desc">
                    {{ getDetailLevelDescription(exportConfig.detailLevel) }}
                  </div>
                </el-form-item>
                
                <el-form-item label="语言版本">
                  <el-radio-group v-model="exportConfig.language">
                    <el-radio-button label="zh-CN">中文</el-radio-button>
                    <el-radio-button label="en-US">英文</el-radio-button>
                    <el-radio-button label="both">双语</el-radio-button>
                  </el-radio-group>
                </el-form-item>
              </el-form>
            </el-tab-pane>

            <!-- 内容选择 -->
            <el-tab-pane label="内容选择" name="content">
              <div class="content-selection">
                <div class="section-header">
                  <h4>文档章节</h4>
                  <div class="section-actions">
                    <el-button size="small" @click="selectAllSections">全选</el-button>
                    <el-button size="small" @click="clearAllSections">清空</el-button>
                    <el-button size="small" @click="resetToDefault">重置默认</el-button>
                  </div>
                </div>
                
                <el-tree
                  ref="sectionTree"
                  :data="sectionTreeData"
                  node-key="id"
                  :props="{ label: 'name', children: 'children' }"
                  show-checkbox
                  :default-checked-keys="exportConfig.sections"
                  @check="onSectionCheck"
                  class="section-tree"
                >
                  <template #default="{ node, data }">
                    <div class="tree-node">
                      <el-icon class="node-icon">{{ getSectionIcon(data.type) }}</el-icon>
                      <span class="node-name">{{ data.name }}</span>
                      <span class="node-description">{{ data.description }}</span>
                      <el-tag v-if="data.required" size="small" type="danger">必需</el-tag>
                    </div>
                  </template>
                </el-tree>
              </div>
            </el-tab-pane>

            <!-- 高级选项 -->
            <el-tab-pane label="高级选项" name="advanced">
              <el-form :model="exportConfig.advanced" label-width="120px">
                <el-form-item label="页面设置">
                  <el-select v-model="exportConfig.advanced.pageSize">
                    <el-option label="A4" value="A4" />
                    <el-option label="A3" value="A3" />
                    <el-option label="Letter" value="Letter" />
                    <el-option label="Legal" value="Legal" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="页面方向">
                  <el-radio-group v-model="exportConfig.advanced.orientation">
                    <el-radio label="portrait">纵向</el-radio>
                    <el-radio label="landscape">横向</el-radio>
                  </el-radio-group>
                </el-form-item>
                
                <el-form-item label="图表样式">
                  <el-select v-model="exportConfig.advanced.chartStyle">
                    <el-option label="简洁风格" value="simple" />
                    <el-option label="商务风格" value="business" />
                    <el-option label="学术风格" value="academic" />
                    <el-option label="技术风格" value="technical" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="数据精度">
                  <el-input-number 
                    v-model="exportConfig.advanced.precision" 
                    :min="0" 
                    :max="6"
                    controls-position="right"
                  />
                  <span class="form-tip">小数点后位数</span>
                </el-form-item>
                
                <el-form-item label="文档选项">
                  <el-checkbox-group v-model="exportConfig.advanced.documentOptions">
                    <el-checkbox label="includeTableOfContents">包含目录</el-checkbox>
                    <el-checkbox label="includePageNumbers">包含页码</el-checkbox>
                    <el-checkbox label="includeHeader">包含页眉</el-checkbox>
                    <el-checkbox label="includeFooter">包含页脚</el-checkbox>
                    <el-checkbox label="includeWatermark">包含水印</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
                
                <el-form-item label="安全选项">
                  <el-checkbox-group v-model="exportConfig.advanced.securityOptions">
                    <el-checkbox label="enablePassword">启用密码保护</el-checkbox>
                    <el-checkbox label="enableDigitalSignature">启用数字签名</el-checkbox>
                    <el-checkbox label="preventCopy">禁止复制</el-checkbox>
                    <el-checkbox label="preventEdit">禁止编辑</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
                
                <el-form-item v-if="exportConfig.advanced.securityOptions.includes('enablePassword')" label="文档密码">
                  <el-input 
                    v-model="exportConfig.advanced.password" 
                    type="password" 
                    placeholder="输入文档密码"
                    show-password
                  />
                </el-form-item>
              </el-form>
            </el-tab-pane>
            
            <!-- 批量导出 -->
            <el-tab-pane label="批量导出" name="batch">
              <div class="batch-export-section">
                <el-alert
                  title="批量导出"
                  description="同时导出多个项目或周期的文档"
                  type="info"
                  show-icon
                  :closable="false"
                />
                
                <el-form :model="batchConfig" label-width="120px" style="margin-top: 20px;">
                  <el-form-item label="导出范围">
                    <el-radio-group v-model="batchConfig.scope">
                      <el-radio label="currentProject">当前项目所有周期</el-radio>
                      <el-radio label="selectedProjects">选择项目</el-radio>
                      <el-radio label="allProjects">所有项目</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  
                  <el-form-item v-if="batchConfig.scope === 'selectedProjects'" label="选择项目">
                    <el-select 
                      v-model="batchConfig.selectedProjects" 
                      multiple 
                      placeholder="选择要导出的项目"
                      style="width: 100%;"
                    >
                      <el-option 
                        v-for="project in availableProjects"
                        :key="project.id"
                        :label="project.name"
                        :value="project.id"
                      />
                    </el-select>
                  </el-form-item>
                  
                  <el-form-item label="压缩格式">
                    <el-radio-group v-model="batchConfig.compressFormat">
                      <el-radio label="zip">ZIP压缩包</el-radio>
                      <el-radio label="rar">RAR压缩包</el-radio>
                      <el-radio label="none">不压缩</el-radio>
                    </el-radio-group>
                  </el-form-item>
                  
                  <el-form-item label="文件命名">
                    <el-input 
                      v-model="batchConfig.namingPattern" 
                      placeholder="使用变量: {project}, {cycle}, {date}, {time}"
                    />
                    <div class="naming-preview">
                      预览: {{ getBatchNamingPreview() }}
                    </div>
                  </el-form-item>
                </el-form>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <div class="footer-left">
            <el-button @click="saveAsTemplate">
              <el-icon><FolderAdd /></el-icon>
              保存为模板
            </el-button>
          </div>
          <div class="footer-right">
            <el-button @click="showOptions = false">取消</el-button>
            <el-button type="info" @click="previewDocument">预览</el-button>
            <el-button 
              type="primary" 
              @click="exportDocument"
              :loading="exportLoading"
            >
              {{ batchConfig.scope !== 'currentProject' && activeTab === 'batch' ? '批量导出' : '导出文档' }}
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>

    <!-- 导出进度对话框 -->
    <el-dialog
      v-model="showProgress"
      title="文档导出进度"
      width="600px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="export-progress-container">
        <!-- 当前任务信息 -->
        <div class="current-task-info">
          <div class="task-header">
            <el-icon class="task-icon"><DocumentCopy /></el-icon>
            <div class="task-details">
              <h4>{{ currentTask.title }}</h4>
              <p>{{ currentTask.description }}</p>
            </div>
            <el-tag :type="getTaskStatusType(currentTask.status)">{{ getTaskStatusText(currentTask.status) }}</el-tag>
          </div>
        </div>
        
        <!-- 总体进度 -->
        <div class="overall-progress">
          <div class="progress-header">
            <span>总体进度</span>
            <span class="progress-text">{{ overallProgress.completed }}/{{ overallProgress.total }}</span>
          </div>
          <el-progress
            :percentage="overallProgress.percentage"
            :stroke-width="12"
            :color="getProgressColor(overallProgress.percentage)"
            class="main-progress-bar"
          />
          <div class="time-info">
            <span>已用时间: {{ formatDuration(overallProgress.elapsed) }}</span>
            <span>预计剩余: {{ formatDuration(overallProgress.remaining) }}</span>
          </div>
        </div>
        
        <!-- 详细步骤 -->
        <div class="step-details">
          <h4>处理步骤</h4>
          <div class="steps-list">
            <div 
              v-for="(step, index) in processingSteps" 
              :key="index"
              :class="['step-item', step.status]"
            >
              <div class="step-indicator">
                <el-icon v-if="step.status === 'completed'" class="step-icon completed"><Check /></el-icon>
                <el-icon v-else-if="step.status === 'processing'" class="step-icon processing is-loading"><Loading /></el-icon>
                <el-icon v-else-if="step.status === 'failed'" class="step-icon failed"><Close /></el-icon>
                <span v-else class="step-number">{{ index + 1 }}</span>
              </div>
              <div class="step-content">
                <div class="step-title">{{ step.title }}</div>
                <div class="step-description">{{ step.description }}</div>
                <div v-if="step.status === 'processing'" class="step-progress">
                  <el-progress 
                    :percentage="step.progress" 
                    :show-text="false"
                    :stroke-width="4"
                  />
                </div>
              </div>
              <div class="step-time">{{ step.duration ? formatDuration(step.duration) : '--' }}</div>
            </div>
          </div>
        </div>
        
        <!-- 导出日志 -->
        <div class="export-logs">
          <div class="log-header">
            <h4>导出日志</h4>
            <el-button size="small" @click="clearLogs">清空日志</el-button>
          </div>
          <div class="log-content">
            <div v-for="log in exportLogs" :key="log.id" :class="['log-item', log.level.toLowerCase()]">
              <span class="log-time">{{ formatLogTime(log.timestamp) }}</span>
              <span class="log-level">{{ log.level }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="progress-footer">
          <el-button @click="cancelExport" :disabled="exportCompleted">
            {{ exportCompleted ? '关闭' : '取消导出' }}
          </el-button>
          <el-button 
            v-if="exportCompleted && exportResult.files && exportResult.files.length > 0"
            type="primary" 
            @click="downloadAllFiles"
          >
            <el-icon><Download /></el-icon>
            下载所有文件 ({{ exportResult.files.length }})
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 文档预览对话框 -->
    <DocumentPreview
      v-model="showPreview"
      :preview-data="previewData"
      :format="exportConfig.format"
      @download="downloadPreview"
    />

    <!-- 模板管理对话框 -->
    <!-- <TemplateManager
      v-model="showTemplateManager"
      type="document"
      @template-updated="refreshTemplates"
    /> -->
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElNotification } from 'element-plus'
import { 
  Document, 
  Edit,
  Setting,
  Avatar,
  Tools,
  FolderAdd,
  DocumentCopy,
  Check,
  Close,
  Loading,
  Download
} from '@element-plus/icons-vue'
import DocumentPreview from './DocumentPreview.vue'
// import TemplateManager from './TemplateManager.vue' // 暂时注释，该组件不存在
import {
  exportComprehensiveDocument,
  getProjectCycleInfo,
  getDocumentTemplates,
  generateDocumentPreview,
  getExportProgress,
  cancelDocumentExport,
  downloadExportedFile,
  batchExportDocuments
} from '@/api/nesma/documentExport'

const props = defineProps({
  cycleId: [String, Number],
  project: Object,
  visible: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:visible', 'export-completed', 'download'])

// 响应式数据
const showOptions = ref(false)
const showProgress = ref(false)
const showPreview = ref(false)
const showTemplateManager = ref(false)
const activeTab = ref('basic')
const exportLoading = ref(false)
const currentExportId = ref(null)

// 导出配置
const exportConfig = reactive({
  type: 'comprehensive',
  format: 'word',
  title: '',
  templateId: null,
  detailLevel: 3,
  language: 'zh-CN',
  sections: [],
  advanced: {
    pageSize: 'A4',
    orientation: 'portrait',
    chartStyle: 'business',
    precision: 2,
    documentOptions: ['includeTableOfContents', 'includePageNumbers'],
    securityOptions: [],
    password: ''
  }
})

// 批量导出配置
const batchConfig = reactive({
  scope: 'currentProject',
  selectedProjects: [],
  compressFormat: 'zip',
  namingPattern: '{project}_{cycle}_{date}'
})

// 模板数据
const documentTemplates = ref([])
const availableProjects = ref([])

// 过滤模板
const filteredTemplates = computed(() => {
  return documentTemplates.value.filter(template => 
    template && 
    (template.id || template.ID) && 
    (template.name || template.title) &&
    (!exportConfig.format || template.format === exportConfig.format)
  )
})

// 章节树数据
const sectionTreeData = ref([
  {
    id: 'project_overview',
    name: '项目概览',
    description: '项目基本信息和背景',
    type: 'overview',
    required: true
  },
  {
    id: 'requirements',
    name: '需求文档',
    description: '详细需求规格说明',
    type: 'requirement',
    required: true,
    children: [
      { id: 'functional_requirements', name: '功能性需求', description: '系统功能描述', type: 'functional' },
      { id: 'non_functional_requirements', name: '非功能性需求', description: '性能、安全等要求', type: 'non_functional' },
      { id: 'business_rules', name: '业务规则', description: '业务逻辑和约束', type: 'business' }
    ]
  },
  {
    id: 'analysis_results',
    name: '分析结果',
    description: 'NESMA分析和评估结果',
    type: 'analysis',
    children: [
      { id: 'function_point_analysis', name: '功能点分析', description: 'NESMA功能点计算', type: 'analysis' },
      { id: 'complexity_assessment', name: '复杂度评估', description: '项目复杂度分析', type: 'analysis' },
      { id: 'quality_metrics', name: '质量指标', description: '代码和设计质量评估', type: 'quality' },
      { id: 'risk_analysis', name: '风险分析', description: '项目风险识别和评估', type: 'risk' }
    ]
  },
  {
    id: 'technical_documentation',
    name: '技术文档',
    description: '技术设计和实现文档',
    type: 'technical',
    children: [
      { id: 'system_architecture', name: '系统架构', description: '整体架构设计', type: 'architecture' },
      { id: 'database_design', name: '数据库设计', description: '数据模型和表结构', type: 'database' },
      { id: 'api_documentation', name: 'API文档', description: '接口规格说明', type: 'api' },
      { id: 'deployment_guide', name: '部署指南', description: '系统部署和配置', type: 'deployment' }
    ]
  },
  {
    id: 'appendices',
    name: '附录',
    description: '补充材料和参考信息',
    type: 'appendix',
    children: [
      { id: 'glossary', name: '术语表', description: '专业术语解释', type: 'glossary' },
      { id: 'references', name: '参考文献', description: '引用资料列表', type: 'reference' },
      { id: 'raw_data', name: '原始数据', description: '分析用原始数据', type: 'data' }
    ]
  }
])

// 进度相关数据
const currentTask = ref({ title: '', description: '', status: 'pending' })
const overallProgress = ref({ completed: 0, total: 0, percentage: 0, elapsed: 0, remaining: 0 })
const processingSteps = ref([])
const exportLogs = ref([])
const exportCompleted = ref(false)
const exportResult = ref({ files: [] })
const previewData = ref(null)

// 计算属性
const defaultTitle = computed(() => {
  if (!props.project) return '项目文档'
  const typeNames = {
    comprehensive: '综合文档',
    requirement_only: '需求文档', 
    analysis_only: '分析报告'
  }
  return `${typeNames[exportConfig.type]}_${props.project.name}_${new Date().toLocaleDateString().replace(/\//g, '')}`
})

const detailLevelMarks = computed(() => ({
  1: '概要',
  2: '简要', 
  3: '标准',
  4: '详细',
  5: '完整'
}))

// 监听props变化
watch(() => props.visible, (newVal) => {
  showOptions.value = newVal
  if (newVal) {
    initializeConfig()
  }
})

watch(showOptions, (newVal) => {
  emit('update:visible', newVal)
})

// 方法实现
const initializeConfig = () => {
  exportConfig.title = defaultTitle.value
  if (props.cycleId) {
    loadProjectCycleInfo()
  }
}

const loadProjectCycleInfo = async () => {
  try {
    const response = await getProjectCycleInfo(props.cycleId)
    if (response.code === 0) {
      const { projectInfo, cycleInfo } = response.data
      exportConfig.title = `${exportConfig.type}_${projectInfo.name}_${cycleInfo.name}`
    }
  } catch (error) {
    console.error('加载项目周期信息失败:', error)
  }
}

const onExportTypeChange = (type) => {
  const defaultSections = {
    comprehensive: ['project_overview', 'requirements', 'analysis_results', 'technical_documentation'],
    requirement_only: ['project_overview', 'requirements', 'business_rules'],
    analysis_only: ['project_overview', 'analysis_results', 'quality_metrics']
  }
  exportConfig.sections = defaultSections[type] || []
}

const onSectionCheck = (data, checked) => {
  const checkedKeys = checked.checkedKeys
  const halfCheckedKeys = checked.halfCheckedKeys
  exportConfig.sections = [...checkedKeys, ...halfCheckedKeys]
}

const selectAllSections = () => {
  const allIds = getAllSectionIds(sectionTreeData.value)
  exportConfig.sections = allIds
  nextTick(() => {
    $refs.sectionTree.setCheckedKeys(allIds)
  })
}

const clearAllSections = () => {
  exportConfig.sections = []
  nextTick(() => {
    $refs.sectionTree.setCheckedKeys([])
  })
}

const resetToDefault = () => {
  onExportTypeChange(exportConfig.type)
  nextTick(() => {
    $refs.sectionTree.setCheckedKeys(exportConfig.sections)
  })
}

const getAllSectionIds = (sections) => {
  let ids = []
  sections.forEach(section => {
    ids.push(section.id)
    if (section.children) {
      ids = ids.concat(getAllSectionIds(section.children))
    }
  })
  return ids
}

const getSectionIcon = (type) => {
  const icons = {
    overview: 'House',
    requirement: 'Document',
    analysis: 'DataAnalysis', 
    technical: 'Tools',
    appendix: 'FolderOpened',
    functional: 'Setting',
    non_functional: 'Shield',
    business: 'Briefcase',
    quality: 'Medal',
    risk: 'Warning',
    architecture: 'Connection',
    database: 'Coin',
    api: 'Link',
    deployment: 'Upload'
  }
  return icons[type] || 'Document'
}

const getDetailLevelDescription = (level) => {
  const descriptions = {
    1: '仅包含关键信息和摘要',
    2: '包含主要内容，省略细节',
    3: '标准详细程度，适合大多数场景',
    4: '包含详细分析和技术细节',
    5: '完整详细，包含所有可用信息'
  }
  return descriptions[level] || ''
}

const getTemplateTypeColor = (type) => {
  const colors = {
    comprehensive: 'primary',
    requirement: 'success',
    analysis: 'warning',
    technical: 'info'
  }
  return colors[type] || 'default'
}

const applyPreset = (presetType) => {
  const presets = {
    management: {
      type: 'comprehensive',
      format: 'pdf',
      detailLevel: 2,
      sections: ['project_overview', 'analysis_results', 'quality_metrics'],
      advanced: {
        chartStyle: 'business',
        documentOptions: ['includeTableOfContents', 'includePageNumbers', 'includeHeader']
      }
    },
    technical: {
      type: 'comprehensive',
      format: 'word',
      detailLevel: 4,
      sections: ['requirements', 'technical_documentation', 'appendices'],
      advanced: {
        chartStyle: 'technical',
        documentOptions: ['includeTableOfContents', 'includePageNumbers']
      }
    },
    audit: {
      type: 'analysis_only',
      format: 'pdf',
      detailLevel: 5,
      sections: ['project_overview', 'analysis_results', 'quality_metrics', 'risk_analysis'],
      advanced: {
        chartStyle: 'academic',
        documentOptions: ['includeTableOfContents', 'includePageNumbers', 'includeWatermark'],
        securityOptions: ['enableDigitalSignature', 'preventCopy']
      }
    }
  }
  
  const preset = presets[presetType]
  if (preset) {
    Object.assign(exportConfig, preset)
    ElMessage.success(`已应用${presetType === 'management' ? '管理' : presetType === 'technical' ? '技术' : '审计'}报告预设`)
  }
}

const getBatchNamingPreview = () => {
  const now = new Date()
  const variables = {
    '{project}': props.project?.name || 'ProjectName',
    '{cycle}': 'CycleName',
    '{date}': now.toLocaleDateString().replace(/\//g, ''),
    '{time}': now.toLocaleTimeString().replace(/:/g, '')
  }
  
  let preview = batchConfig.namingPattern
  Object.keys(variables).forEach(key => {
    preview = preview.replace(key, variables[key])
  })
  
  return `${preview}.${exportConfig.format}`
}

const previewDocument = async () => {
  if (!props.cycleId) {
    ElMessage.error('缺少周期ID')
    return
  }
  
  try {
    const response = await generateDocumentPreview({
      cycleId: props.cycleId,
      exportType: exportConfig.type,
      documentFormat: exportConfig.format,
      sections: exportConfig.sections,
      detailLevel: exportConfig.detailLevel,
      advanced: exportConfig.advanced
    })
    
    previewData.value = response.data
    showPreview.value = true
    
  } catch (error) {
    ElMessage.error('生成预览失败: ' + error.message)
  }
}

const exportDocument = async () => {
  if (!props.cycleId) {
    ElMessage.error('缺少周期ID')
    return
  }
  
  // 验证必填项
  if (!exportConfig.title.trim()) {
    ElMessage.error('请输入文档标题')
    return
  }
  
  if (exportConfig.sections.length === 0) {
    ElMessage.error('请至少选择一个章节')
    return
  }
  
  exportLoading.value = true
  
  try {
    // 判断是批量导出还是单个导出
    const isBatchExport = activeTab.value === 'batch' && batchConfig.scope !== 'currentProject'
    
    let response
    if (isBatchExport) {
      response = await batchExportDocuments({
        scope: batchConfig.scope,
        selectedProjects: batchConfig.selectedProjects,
        exportConfig: exportConfig,
        compressFormat: batchConfig.compressFormat,
        namingPattern: batchConfig.namingPattern
      })
    } else {
      response = await exportComprehensiveDocument({
        cycleId: props.cycleId,
        exportType: exportConfig.type,
        documentFormat: exportConfig.format,
        title: exportConfig.title,
        sections: exportConfig.sections,
        detailLevel: exportConfig.detailLevel,
        language: exportConfig.language,
        templateId: exportConfig.templateId,
        advanced: exportConfig.advanced
      })
    }
    
    if (response.code === 0) {
      currentExportId.value = response.data.exportId
      exportResult.value = response.data
      
      showOptions.value = false
      showProgress.value = true
      
      // 初始化进度数据
      initializeProgressData(response.data)
      
      // 开始轮询进度
      pollExportProgress()
      
      ElMessage.success('文档导出已启动')
    } else {
      throw new Error(response.msg || '导出失败')
    }
    
  } catch (error) {
    ElMessage.error('启动导出失败: ' + error.message)
  } finally {
    exportLoading.value = false
  }
}

const initializeProgressData = (exportData) => {
  currentTask.value = {
    title: exportData.title || exportConfig.title,
    description: `正在导出${exportConfig.type}格式的${exportConfig.format.toUpperCase()}文档`,
    status: 'processing'
  }
  
  overallProgress.value = {
    completed: 0,
    total: exportData.totalSteps || 6,
    percentage: 0,
    elapsed: 0,
    remaining: 0
  }
  
  processingSteps.value = [
    { title: '解析配置', description: '分析导出配置和参数', status: 'pending', progress: 0 },
    { title: '加载数据', description: '从数据库加载项目数据', status: 'pending', progress: 0 },
    { title: '构建文档结构', description: '根据配置构建文档框架', status: 'pending', progress: 0 },
    { title: '生成内容', description: '生成各章节内容', status: 'pending', progress: 0 },
    { title: '格式化文档', description: '应用样式和格式', status: 'pending', progress: 0 },
    { title: '保存文件', description: '保存最终文档文件', status: 'pending', progress: 0 }
  ]
  
  exportLogs.value = []
  exportCompleted.value = false
}

const pollExportProgress = async () => {
  if (!currentExportId.value) return
  
  const interval = setInterval(async () => {
    try {
      const response = await getExportProgress(currentExportId.value)
      const progressData = response.data
      
      // 更新总体进度
      overallProgress.value = {
        completed: progressData.completedSteps || 0,
        total: progressData.totalSteps || 6,
        percentage: progressData.percentage || 0,
        elapsed: progressData.elapsedTime || 0,
        remaining: progressData.estimatedRemaining || 0
      }
      
      // 更新当前任务状态
      currentTask.value.status = progressData.status
      currentTask.value.description = progressData.currentStep || currentTask.value.description
      
      // 更新处理步骤
      if (progressData.steps) {
        processingSteps.value = progressData.steps
      }
      
      // 更新日志
      if (progressData.logs) {
        exportLogs.value = progressData.logs
      }
      
      // 检查是否完成
      if (progressData.status === 'completed') {
        clearInterval(interval)
        exportCompleted.value = true
        currentTask.value.status = 'completed'
        
        if (progressData.exportResult) {
          exportResult.value = progressData.exportResult
        }
        
        ElNotification({
          title: '导出完成',
          message: `文档已成功导出，共生成 ${exportResult.value.files?.length || 1} 个文件`,
          type: 'success',
          duration: 5000
        })
        
        emit('export-completed', {
          exportId: currentExportId.value,
          result: exportResult.value
        })
        
      } else if (progressData.status === 'failed') {
        clearInterval(interval)
        ElMessage.error('文档导出失败: ' + (progressData.errorMessage || '未知错误'))
        showProgress.value = false
      }
      
    } catch (error) {
      clearInterval(interval)
      ElMessage.error('获取导出进度失败: ' + error.message)
      showProgress.value = false
    }
  }, 2000) // 每2秒轮询一次
}

const cancelExport = async () => {
  if (exportCompleted.value) {
    showProgress.value = false
    return
  }
  
  try {
    await ElMessageBox.confirm('确定要取消文档导出吗？', '确认取消', {
      type: 'warning'
    })
    
    if (currentExportId.value) {
      await cancelDocumentExport(currentExportId.value)
    }
    
    showProgress.value = false
    ElMessage.success('已取消文档导出')
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消导出失败: ' + error.message)
    }
  }
}

const downloadAllFiles = async () => {
  if (!exportResult.value.files || exportResult.value.files.length === 0) {
    ElMessage.error('没有可下载的文件')
    return
  }
  
  try {
    for (const file of exportResult.value.files) {
      await downloadExportedFile(currentExportId.value, file.fileName)
      
      // 延迟一下避免同时下载太多文件
      await new Promise(resolve => setTimeout(resolve, 500))
    }
    
    ElMessage.success(`已下载 ${exportResult.value.files.length} 个文件`)
    emit('download', {
      exportId: currentExportId.value,
      files: exportResult.value.files
    })
    
  } catch (error) {
    ElMessage.error('下载文件失败: ' + error.message)
  }
}

const downloadPreview = (previewFile) => {
  emit('download', { type: 'preview', file: previewFile })
}

const saveAsTemplate = async () => {
  try {
    const templateName = await ElMessageBox.prompt('请输入模板名称', '保存为模板', {
      confirmButtonText: '保存',
      cancelButtonText: '取消',
      inputPattern: /^.{1,50}$/,
      inputErrorMessage: '模板名称长度为1-50个字符'
    })
    
    const templateData = {
      name: templateName.value,
      type: exportConfig.type,
      format: exportConfig.format,
      config: {
        sections: exportConfig.sections,
        detailLevel: exportConfig.detailLevel,
        language: exportConfig.language,
        advanced: exportConfig.advanced
      },
      description: `基于当前配置创建的${exportConfig.type}模板`
    }
    
    // 这里应该调用创建模板的API
    ElMessage.success('模板保存成功')
    await refreshTemplates()
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('保存模板失败: ' + error.message)
    }
  }
}

const openTemplateManager = () => {
  showTemplateManager.value = true
}

const refreshTemplates = async () => {
  try {
    const response = await getDocumentTemplates()
    if (response.code === 0) {
      documentTemplates.value = response.data.list || response.data || []
    }
  } catch (error) {
    console.error('刷新模板失败:', error)
    // 使用默认模板
    documentTemplates.value = [
      { id: 1, name: '标准综合文档', type: 'comprehensive', format: 'word' },
      { id: 2, name: '需求规格说明书', type: 'requirement_only', format: 'word' },
      { id: 3, name: 'NESMA分析报告', type: 'analysis_only', format: 'pdf' }
    ]
  }
}

const clearLogs = () => {
  exportLogs.value = []
}

// 辅助方法
const getTaskStatusType = (status) => {
  const types = {
    pending: 'info',
    processing: 'primary', 
    completed: 'success',
    failed: 'danger'
  }
  return types[status] || 'info'
}

const getTaskStatusText = (status) => {
  const texts = {
    pending: '准备中',
    processing: '处理中',
    completed: '已完成', 
    failed: '失败'
  }
  return texts[status] || '未知'
}

const getProgressColor = (percentage) => {
  if (percentage < 30) return '#F56C6C'
  if (percentage < 70) return '#E6A23C' 
  return '#67C23A'
}

const formatDuration = (seconds) => {
  if (!seconds) return '00:00'
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const formatLogTime = (timestamp) => {
  return new Date(timestamp).toLocaleTimeString()
}

// 生命周期
onMounted(() => {
  refreshTemplates()
  onExportTypeChange(exportConfig.type) // 设置默认章节
})

// 暴露方法
defineExpose({
  openOptions: () => { showOptions.value = true },
  exportDocument
})
</script>

<style lang="scss" scoped>
.enhanced-document-generator {
  .document-generator-dialog {
    .document-config-container {
      display: flex;
      gap: 24px;
      min-height: 600px;
      
      .export-type-section {
        flex: 0 0 240px;
        
        h3 {
          margin-bottom: 16px;
          color: #303133;
          font-size: 16px;
          font-weight: 600;
        }
        
        .type-card {
          margin-bottom: 24px;
          
          .el-radio-group {
            display: flex;
            flex-direction: column;
            gap: 8px;
            
            .el-radio-button {
              width: 100%;
              
              :deep(.el-radio-button__inner) {
                width: 100%;
                text-align: left;
                border-radius: 6px;
                transition: all 0.3s;
                
                &:hover {
                  background: #f5f7fa;
                }
              }
            }
          }
        }
        
        .quick-presets {
          h4 {
            margin-bottom: 12px;
            color: #606266;
            font-size: 14px;
          }
          
          .preset-card {
            padding: 12px;
            
            .preset-item {
              display: flex;
              align-items: center;
              gap: 8px;
              padding: 8px 12px;
              margin-bottom: 8px;
              border-radius: 6px;
              cursor: pointer;
              transition: all 0.3s;
              
              &:hover {
                background: #f0f9ff;
                color: #409eff;
              }
              
              .el-icon {
                font-size: 16px;
              }
              
              span {
                font-size: 13px;
              }
            }
          }
        }
      }
      
      .export-options-section {
        flex: 1;
        
        .el-tabs {
          height: 100%;
          
          :deep(.el-tab-pane) {
            max-height: 500px;
            overflow-y: auto;
          }
          
          .template-option {
            display: flex;
            justify-content: space-between;
            align-items: center;
            width: 100%;
            
            .template-name {
              flex: 1;
            }
          }
          
          .template-manage-btn {
            margin-left: 8px;
            font-size: 12px;
          }
          
          .detail-level-desc {
            margin-top: 8px;
            font-size: 12px;
            color: #909399;
            text-align: center;
          }
          
          .form-tip {
            margin-left: 8px;
            font-size: 12px;
            color: #909399;
          }
          
          .option-desc {
            display: block;
            font-size: 12px;
            color: #909399;
            margin-top: 2px;
          }
          
          .content-selection {
            .section-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 16px;
              
              h4 {
                margin: 0;
                color: #303133;
                font-size: 14px;
              }
              
              .section-actions {
                display: flex;
                gap: 8px;
              }
            }
            
            .section-tree {
              max-height: 350px;
              overflow-y: auto;
              border: 1px solid #e4e7ed;
              border-radius: 6px;
              padding: 8px;
              
              .tree-node {
                display: flex;
                align-items: center;
                gap: 8px;
                width: 100%;
                
                .node-icon {
                  color: #409eff;
                  font-size: 14px;
                  flex-shrink: 0;
                }
                
                .node-name {
                  font-weight: 500;
                  color: #303133;
                }
                
                .node-description {
                  color: #909399;
                  font-size: 12px;
                  margin-left: auto;
                  margin-right: 8px;
                }
              }
            }
          }
          
          .batch-export-section {
            .naming-preview {
              margin-top: 8px;
              padding: 8px 12px;
              background: #f5f7fa;
              border-radius: 4px;
              font-size: 12px;
              color: #606266;
              border: 1px solid #e4e7ed;
            }
          }
        }
      }
    }
    
    .dialog-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .footer-left {
        flex: 1;
      }
      
      .footer-right {
        display: flex;
        gap: 12px;
      }
    }
  }
  
  .export-progress-container {
    .current-task-info {
      margin-bottom: 24px;
      
      .task-header {
        display: flex;
        align-items: center;
        gap: 12px;
        padding: 16px;
        background: linear-gradient(135deg, #f8fafc 0%, #e3f2fd 100%);
        border-radius: 8px;
        border: 1px solid #e1e8ed;
        
        .task-icon {
          font-size: 20px;
          color: #409eff;
        }
        
        .task-details {
          flex: 1;
          
          h4 {
            margin: 0 0 4px 0;
            color: #303133;
            font-size: 16px;
          }
          
          p {
            margin: 0;
            color: #606266;
            font-size: 14px;
          }
        }
      }
    }
    
    .overall-progress {
      margin-bottom: 24px;
      
      .progress-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
        
        .progress-text {
          font-weight: 600;
          color: #409eff;
        }
      }
      
      .main-progress-bar {
        margin-bottom: 8px;
      }
      
      .time-info {
        display: flex;
        justify-content: space-between;
        font-size: 12px;
        color: #909399;
      }
    }
    
    .step-details {
      margin-bottom: 24px;
      
      h4 {
        margin-bottom: 16px;
        color: #303133;
        font-size: 14px;
      }
      
      .steps-list {
        .step-item {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 12px 0;
          border-bottom: 1px solid #f0f0f0;
          
          &:last-child {
            border-bottom: none;
          }
          
          .step-indicator {
            width: 32px;
            height: 32px;
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            background: #f5f7fa;
            border: 2px solid #e4e7ed;
            flex-shrink: 0;
            
            .step-number {
              font-size: 12px;
              font-weight: 600;
              color: #909399;
            }
            
            .step-icon {
              font-size: 16px;
              
              &.completed {
                color: #67c23a;
              }
              
              &.processing {
                color: #409eff;
              }
              
              &.failed {
                color: #f56c6c;
              }
            }
          }
          
          &.completed .step-indicator {
            background: #f0f9ff;
            border-color: #67c23a;
          }
          
          &.processing .step-indicator {
            background: #e1f3d8;
            border-color: #409eff;
          }
          
          &.failed .step-indicator {
            background: #fef0f0;
            border-color: #f56c6c;
          }
          
          .step-content {
            flex: 1;
            
            .step-title {
              font-weight: 500;
              color: #303133;
              margin-bottom: 4px;
            }
            
            .step-description {
              font-size: 12px;
              color: #606266;
              margin-bottom: 8px;
            }
            
            .step-progress {
              width: 100%;
            }
          }
          
          .step-time {
            font-size: 12px;
            color: #909399;
            min-width: 60px;
            text-align: right;
          }
        }
      }
    }
    
    .export-logs {
      .log-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;
        
        h4 {
          margin: 0;
          color: #303133;
          font-size: 14px;
        }
      }
      
      .log-content {
        max-height: 200px;
        overflow-y: auto;
        padding: 12px;
        background: #fafbfc;
        border: 1px solid #e4e7ed;
        border-radius: 6px;
        font-family: 'Courier New', monospace;
        font-size: 12px;
        line-height: 1.5;
        
        .log-item {
          display: flex;
          gap: 8px;
          margin-bottom: 4px;
          
          &:last-child {
            margin-bottom: 0;
          }
          
          .log-time {
            color: #909399;
            flex-shrink: 0;
            width: 60px;
          }
          
          .log-level {
            font-weight: 600;
            flex-shrink: 0;
            width: 50px;
            
            &.info {
              color: #409eff;
            }
            
            &.warn {
              color: #e6a23c;
            }
            
            &.error {
              color: #f56c6c;
            }
            
            &.success {
              color: #67c23a;
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
  
  .progress-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
  }
}

// 动画效果
@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}

.is-loading {
  animation: pulse 1.5s ease-in-out infinite;
}
</style>