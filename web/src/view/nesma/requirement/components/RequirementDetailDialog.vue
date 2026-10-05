<template>
  <el-dialog
    v-model="dialogVisible"
    title="需求详情"
    width="1200px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div v-loading="loading" class="requirement-detail">
      <div v-if="requirement" class="detail-content">
        <!-- 使用Tab组件来组织内容，为L4需求添加流程图tab -->
        <el-tabs v-model="activeTab" type="card" class="detail-tabs">
          <!-- 基本信息Tab -->
          <el-tab-pane label="基本信息" name="basic">
            <!-- 基本信息 -->
            <div class="info-section">
              <h3 class="section-title">基本信息</h3>
              <el-row :gutter="20">
                <el-col :span="8">
                  <div class="info-item">
                    <label>需求标题：</label>
                    <span class="value">{{ requirement.title }}</span>
                  </div>
                  <div class="info-item">
                    <label>需求编号：</label>
                    <span class="value">{{ requirement.code || '自动生成' }}</span>
                  </div>
                  <div class="info-item">
                    <label>所属项目：</label>
                    <span class="value">{{ requirement.project?.name || '未知项目' }}</span>
                  </div>
                  <div class="info-item">
                    <label>需求层级：</label>
                    <el-tag :type="getLevelTagType(requirement.level)">
                      {{ requirement.levelName || `L${requirement.level}` }}
                    </el-tag>
                  </div>
                </el-col>
                <el-col :span="8">
                  <div class="info-item">
                    <label>需求状态：</label>
                    <el-tag :type="getStatusTagType(requirement.status)">
                      {{ getStatusLabel(requirement.status) }}
                    </el-tag>
                  </div>
                  <div class="info-item">
                    <label>优先级：</label>
                    <el-rate 
                      :model-value="requirement.priority || 3" 
                      disabled 
                      show-text
                      :texts="['很低', '较低', '一般', '较高', '很高']"
                    />
                  </div>
                  <div class="info-item">
                    <label>排序索引：</label>
                    <span class="value">{{ requirement.orderIndex || 0 }}</span>
                  </div>
                  <div class="info-item">
                    <label>复杂度：</label>
                    <el-tag :type="getComplexityTagType(requirement.complexity)">
                      {{ requirement.complexity || '未设置' }}
                    </el-tag>
                  </div>
                </el-col>
                <el-col :span="8">
                  <div class="info-item">
                    <label>预估工时：</label>
                    <span class="value">{{ requirement.estimateHours || '-' }} 小时</span>
                  </div>
                  <div class="info-item">
                    <label>实际工时：</label>
                    <span class="value">{{ requirement.actualHours || '-' }} 小时</span>
                  </div>
                  <div class="info-item">
                    <label>分类：</label>
                    <span class="value">{{ requirement.category || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <label>建设周期：</label>
                    <span class="value">{{ requirement.constructionPeriod || '-' }}</span>
                  </div>
                </el-col>
              </el-row>
            </div>

            <!-- NESMA功能点信息 -->
            <div class="info-section">
              <h3 class="section-title">NESMA功能点信息</h3>
              <el-row :gutter="20">
                <el-col :span="8">
                  <div class="info-item">
                    <label>功能类型：</label>
                    <el-tag v-if="requirement.functionType" :type="getFunctionTypeTagType(requirement.functionType)">
                      {{ requirement.functionType }}
                    </el-tag>
                    <span v-else class="value">-</span>
                  </div>
                  <div class="info-item">
                    <label>重用程度：</label>
                    <span class="value">{{ requirement.reuseLevel || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <label>修改类型：</label>
                    <span class="value">{{ requirement.modificationType || '-' }}</span>
                  </div>
                </el-col>
                <el-col :span="8">
                  <div class="info-item">
                    <label>AFP值：</label>
                    <span class="value number">{{ requirement.afp || 0 }}</span>
                  </div>
                  <div class="info-item">
                    <label>UFP值：</label>
                    <span class="value number">{{ requirement.ufp || 0 }}</span>
                  </div>
                  <div class="info-item">
                    <label>版本：</label>
                    <span class="value">{{ requirement.version || 0 }}</span>
                  </div>
                </el-col>
                <el-col :span="8">
                  <div class="info-item">
                    <label>导入批次：</label>
                    <span class="value">{{ requirement.importBatch || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <label>导入来源：</label>
                    <span class="value">{{ requirement.importSource || '-' }}</span>
                  </div>
                </el-col>
              </el-row>
            </div>

            <!-- AI分析信息 -->
            <div v-if="requirement.aiAnalysisStatus" class="info-section">
              <h3 class="section-title">AI分析信息</h3>
              <el-row :gutter="20">
                <el-col :span="12">
                  <div class="info-item">
                    <label>分析状态：</label>
                    <el-tag :type="getAIAnalysisTagType(requirement.aiAnalysisStatus)">
                      {{ getAIAnalysisLabel(requirement.aiAnalysisStatus) }}
                    </el-tag>
                  </div>
                  <div class="info-item">
                    <label>AI置信度：</label>
                    <span class="value">{{ requirement.aiConfidenceScore || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <label>AI复杂度评分：</label>
                    <span class="value">{{ requirement.aiComplexityScore || '-' }}</span>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="info-item">
                    <label>推荐AFP：</label>
                    <span class="value number">{{ requirement.recommendedAFP || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <label>推荐UFP：</label>
                    <span class="value number">{{ requirement.recommendedUFP || '-' }}</span>
                  </div>
                  <div class="info-item">
                    <label>分析时间：</label>
                    <span class="value">{{ formatDate(requirement.aiAnalysisTime) }}</span>
                  </div>
                </el-col>
              </el-row>
            </div>

            <!-- 需求描述 -->
            <div class="info-section">
              <h3 class="section-title">需求描述</h3>
              <div class="description-content">
                <p>{{ requirement.description || '暂无描述' }}</p>
              </div>
            </div>

            <!-- AI优化描述 -->
            <div v-if="requirement.aiDescription" class="info-section">
              <h3 class="section-title">AI优化描述</h3>
              <div class="description-content ai-content">
                <p>{{ requirement.aiDescription }}</p>
              </div>
            </div>

            <!-- AI生成标题 -->
            <div v-if="requirement.aiGeneratedTitle" class="info-section">
              <h3 class="section-title">AI生成标题</h3>
              <div class="description-content ai-content">
                <p>{{ requirement.aiGeneratedTitle }}</p>
              </div>
            </div>

            <!-- 业务价值 -->
            <div v-if="requirement.businessValue" class="info-section">
              <h3 class="section-title">业务价值</h3>
              <div class="description-content">
                <p>{{ requirement.businessValue }}</p>
              </div>
            </div>

            <!-- 验收标准 -->
            <div v-if="requirement.acceptanceCriteria" class="info-section">
              <h3 class="section-title">验收标准</h3>
              <div class="description-content">
                <p>{{ requirement.acceptanceCriteria }}</p>
              </div>
            </div>

            <!-- 备注 -->
            <div v-if="requirement.notes" class="info-section">
              <h3 class="section-title">备注</h3>
              <div class="description-content">
                <p>{{ requirement.notes }}</p>
              </div>
            </div>

            <!-- 子需求 -->
            <div v-if="children && children.length > 0" class="info-section">
              <h3 class="section-title">子需求</h3>
              <el-table :data="children" style="width: 100%">
                <el-table-column prop="code" label="需求编号" width="120" />
                <el-table-column prop="title" label="需求名称" min-width="200" />
                <el-table-column prop="level" label="层级" width="80">
                  <template #default="scope">
                    <el-tag size="small" :type="getLevelTagType(scope.row.level)">
                      {{ scope.row.levelName || `L${scope.row.level}` }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="scope">
                    <el-tag size="small" :type="getStatusTagType(scope.row.status)">
                      {{ getStatusLabel(scope.row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="120">
                  <template #default="scope">
                    <el-button
                      size="small"
                      type="primary"
                      link
                      @click="handleViewChild(scope.row)"
                    >
                      查看
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>

            <!-- 路径信息 -->
            <div class="info-section">
              <h3 class="section-title">路径信息</h3>
              <div class="path-info">
                <el-breadcrumb separator=">">
                  <el-breadcrumb-item v-for="(path, index) in getPathItems(requirement.fullPath)" :key="index">
                    {{ path }}
                  </el-breadcrumb-item>
                </el-breadcrumb>
              </div>
            </div>

            <!-- 时间信息 -->
            <div class="info-section">
              <h3 class="section-title">时间信息</h3>
              <el-row :gutter="20">
                <el-col :span="12">
                  <div class="info-item">
                    <label>创建时间：</label>
                    <span class="value">{{ formatDate(requirement.CreatedAt) }}</span>
                  </div>
                </el-col>
                <el-col :span="12">
                  <div class="info-item">
                    <label>更新时间：</label>
                    <span class="value">{{ formatDate(requirement.UpdatedAt) }}</span>
                  </div>
                </el-col>
              </el-row>
            </div>
          </el-tab-pane>

          <!-- L4需求流程图Tab -->
          <el-tab-pane v-if="requirement.level === 4" label="流程图" name="flowchart">
            <div class="flowchart-content">
              <!-- 流程图操作区域 -->
              <div class="flowchart-actions">
                <el-alert 
                  title="L4功能点流程图" 
                  type="info" 
                  :closable="false" 
                  show-icon
                  style="margin-bottom: 20px;"
                >
                  <template #default>
                    <p>为此L4功能点生成或查看Mermaid流程图，帮助理解业务逻辑和数据流向。</p>
                  </template>
                </el-alert>
                
                <div class="action-buttons">
                  <el-button 
                    v-if="!hasMermaidDiagram"
                    type="primary" 
                    @click="handleGenerateFlowchart"
                    :loading="flowchartGenerating"
                  >
                    <el-icon><Coordinate /></el-icon>
                    生成流程图
                  </el-button>
                  <el-button 
                    v-if="hasMermaidDiagram"
                    type="success" 
                    @click="handleRefreshFlowchart"
                    :loading="flowchartGenerating"
                  >
                    <el-icon><Refresh /></el-icon>
                    重新生成
                  </el-button>
                  <el-button 
                    v-if="hasMermaidDiagram"
                    type="info" 
                    @click="handleDownloadFlowchart"
                  >
                    <el-icon><Download /></el-icon>
                    下载图片
                  </el-button>
                </div>
              </div>

              <!-- 流程图展示区域 -->
              <div v-if="hasMermaidDiagram" class="flowchart-display">
                <div class="flowchart-header">
                  <h4>{{ requirement.title }} - 业务流程图</h4>
                  <div class="flowchart-meta">
                    <span>更新时间: {{ formatDate(requirement.mermaidUpdatedAt) }}</span>
                    <el-tag v-if="requirement.mermaidStatus" :type="getMermaidStatusTagType(requirement.mermaidStatus)">
                      {{ getMermaidStatusLabel(requirement.mermaidStatus) }}
                    </el-tag>
                  </div>
                </div>

                <!-- Mermaid代码展示 -->
                <div class="mermaid-code-section">
                  <el-collapse v-model="codeCollapseActive">
                    <el-collapse-item title="查看Mermaid代码" name="code">
                      <el-input 
                        v-model="requirement.mermaidDiagram" 
                        type="textarea" 
                        :rows="8" 
                        readonly
                        placeholder="暂无流程图代码"
                      />
                      <div class="code-actions">
                        <el-button size="small" @click="copyMermaidCode">
                          <el-icon><CopyDocument /></el-icon>
                          复制代码
                        </el-button>
                        <el-button size="small" @click="downloadMermaidCode">
                          <el-icon><Download /></el-icon>
                          下载代码
                        </el-button>
                      </div>
                    </el-collapse-item>
                  </el-collapse>
                </div>

                <!-- 流程图可视化展示 -->
                <div class="mermaid-visualization">
                  <div class="visualization-header">
                    <h5>流程图可视化</h5>
                    <div class="visualization-tools">
                      <el-button-group size="small">
                        <el-button @click="toggleTheme">
                          <el-icon><Setting /></el-icon>
                          切换主题
                        </el-button>
                        <el-button @click="refreshMermaidRender">
                          <el-icon><Refresh /></el-icon>
                          刷新渲染
                        </el-button>
                      </el-button-group>
                    </div>
                  </div>
                  
                  <!-- 使用MermaidRenderer组件进行实时渲染 -->
                  <MermaidRenderer 
                    :mermaid-code="requirement.mermaidDiagram" 
                    :show-controls="true"
                    :theme="currentMermaidTheme"
                    :auto-render="true"
                    @render-success="handleMermaidRenderSuccess"
                    @render-error="handleMermaidRenderError"
                  />
                </div>
              </div>

              <!-- 没有流程图时的提示 -->
              <div v-else class="no-flowchart">
                <el-empty description="暂无流程图">
                  <template #image>
                    <el-icon size="60" color="#d9d9d9"><Coordinate /></el-icon>
                  </template>
                  <template #description>
                    <p>此L4功能点还没有生成流程图</p>
                    <p>点击上方"生成流程图"按钮，AI将为您自动生成Mermaid流程图</p>
                  </template>
                </el-empty>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="primary" @click="handleEdit">
          编辑
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Coordinate, Refresh, Download, CopyDocument, Setting
} from '@element-plus/icons-vue'
import { getNesmaRequirement } from '@/api/nesma'
import { generateMermaidDiagramsAsync } from '@/api/nesma'
import MermaidRenderer from '@/components/MermaidRenderer.vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  requirementId: {
    type: Number,
    default: null
  }
})

const emit = defineEmits(['update:modelValue', 'edit', 'viewChild'])

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const loading = ref(false)
const requirement = ref(null)
const children = ref([])

// Tab管理
const activeTab = ref('basic')

// 流程图相关状态
const flowchartGenerating = ref(false)
const codeCollapseActive = ref([])
const currentMermaidTheme = ref('default')

// 计算属性：是否有Mermaid流程图
const hasMermaidDiagram = computed(() => {
  return requirement.value && requirement.value.mermaidDiagram && requirement.value.mermaidDiagram.trim().length > 0
})

// 监听对话框打开
watch(dialogVisible, (newVal) => {
  if (newVal && props.requirementId) {
    loadRequirementDetail()
    // 如果是L4需求且有流程图，默认显示流程图tab
    if (requirement.value?.level === 4 && hasMermaidDiagram.value) {
      activeTab.value = 'flowchart'
    } else {
      activeTab.value = 'basic'
    }
  }
})

// 流程图相关方法
const handleGenerateFlowchart = async () => {
  if (!requirement.value) return
  
  try {
    flowchartGenerating.value = true
    
    const generateData = {
      projectId: requirement.value.projectId,
      cycleId: requirement.value.cycleId,
      versionId: requirement.value.versionId,
      requirementIds: [requirement.value.id || requirement.value.ID],
      diagramType: 'flowchart',
      detailLevel: 'detailed',
      options: {
        includeSubRequirements: false,
        autoLayout: true,
        addAnnotations: true,
        groupByLevel: false
      }
    }

    ElMessage.info('正在为此L4功能点生成流程图...')
    const response = await generateMermaidDiagramsAsync(generateData)
    
    if (response.code === 0) {
      ElMessage.success('流程图生成任务已启动，请稍候...')
      
      // 可以在这里轮询检查生成状态
      setTimeout(() => {
        // 重新加载需求详情以获取最新的流程图
        loadRequirementDetail()
      }, 5000)
      
    } else {
      throw new Error(response.msg || 'Mermaid生成失败')
    }

  } catch (error) {
    console.error('流程图生成失败:', error)
    ElMessage.error('流程图生成失败: ' + (error.message || '未知错误'))
  } finally {
    flowchartGenerating.value = false
  }
}

const handleRefreshFlowchart = () => {
  handleGenerateFlowchart()
}

const handleDownloadFlowchart = () => {
  if (!requirement.value?.mermaidDiagram) {
    ElMessage.warning('没有可下载的流程图')
    return
  }
  
  // 创建下载链接
  const blob = new Blob([requirement.value.mermaidDiagram], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${requirement.value.title}_流程图.mmd`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success('流程图文件已下载')
}

const copyMermaidCode = async () => {
  if (!requirement.value?.mermaidDiagram) {
    ElMessage.warning('没有可复制的代码')
    return
  }
  
  try {
    await navigator.clipboard.writeText(requirement.value.mermaidDiagram)
    ElMessage.success('Mermaid代码已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败，请手动复制')
  }
}

const downloadMermaidCode = () => {
  if (!requirement.value?.mermaidDiagram) {
    ElMessage.warning('没有可下载的代码')
    return
  }
  
  const blob = new Blob([requirement.value.mermaidDiagram], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${requirement.value.title}_mermaid.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success('Mermaid代码已下载')
}

// Mermaid渲染增强方法
const toggleTheme = () => {
  const themes = ['default', 'dark', 'forest', 'neutral']
  const currentIndex = themes.indexOf(currentMermaidTheme.value)
  const nextIndex = (currentIndex + 1) % themes.length
  currentMermaidTheme.value = themes[nextIndex]
  
  ElMessage.success(`已切换到 ${currentMermaidTheme.value} 主题`)
}

const refreshMermaidRender = () => {
  ElMessage.info('正在刷新流程图渲染...')
  // MermaidRenderer组件会自动重新渲染
}

const handleMermaidRenderSuccess = (data) => {
  console.log('需求详情页Mermaid渲染成功:', data)
  ElMessage.success('流程图渲染成功')
}

const handleMermaidRenderError = (error) => {
  console.error('需求详情页Mermaid渲染失败:', error)
  ElMessage.error('流程图渲染失败: ' + error.message)
}

// Mermaid状态相关方法
const getMermaidStatusTagType = (status) => {
  const types = {
    'pending': 'warning',
    'generating': 'primary',
    'completed': 'success',
    'failed': 'danger'
  }
  return types[status] || 'info'
}

const getMermaidStatusLabel = (status) => {
  const labels = {
    'pending': '等待生成',
    'generating': '生成中',
    'completed': '已完成',
    'failed': '生成失败'
  }
  return labels[status] || '未知'
}

// 加载需求详情
const loadRequirementDetail = async () => {
  try {
    loading.value = true
    const res = await getNesmaRequirement(props.requirementId)
    requirement.value = res.data
    children.value = res.data.children || []
  } catch (error) {
    console.error('加载需求详情失败:', error)
    ElMessage.error('加载需求详情失败')
  } finally {
    loading.value = false
  }
}

// 获取层级标签类型
const getLevelTagType = (level) => {
  const types = {
    1: 'danger',
    2: 'warning', 
    3: 'info',
    4: 'success'
  }
  return types[level] || 'info'
}

// 获取层级标签
const getLevelLabel = (level) => {
  const labels = {
    1: 'L1',
    2: 'L2',
    3: 'L3',
    4: 'FP'
  }
  return labels[level] || 'L?'
}

// 获取状态标签类型
const getStatusTagType = (status) => {
  const types = {
    'pending': 'warning',
    'in_progress': 'primary',
    'completed': 'success',
    'cancelled': 'danger'
  }
  return types[status] || 'info'
}

// 获取状态标签
const getStatusLabel = (status) => {
  const labels = {
    'pending': '待处理',
    'in_progress': '进行中',
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return labels[status] || '未知'
}

// 获取复杂度标签类型
const getComplexityTagType = (complexity) => {
  const types = {
    '简单': 'success',
    '中等': 'warning',
    '复杂': 'danger'
  }
  return types[complexity] || 'info'
}

// 获取功能类型标签类型
const getFunctionTypeTagType = (functionType) => {
  const types = {
    'EI': 'primary',
    'EO': 'success',
    'EQ': 'warning',
    'ILF': 'info',
    'EIF': 'danger'
  }
  return types[functionType] || 'info'
}

// 获取AI分析状态标签类型
const getAIAnalysisTagType = (status) => {
  const types = {
    'pending': 'warning',
    'running': 'primary',
    'completed': 'success',
    'failed': 'danger'
  }
  return types[status] || 'info'
}

// 获取AI分析状态标签
const getAIAnalysisLabel = (status) => {
  const labels = {
    'pending': '待分析',
    'running': '分析中',
    'completed': '已完成',
    'failed': '分析失败'
  }
  return labels[status] || '未知'
}

// 获取路径项目
const getPathItems = (fullPath) => {
  if (!fullPath) return []
  return fullPath.split(' > ').filter(item => item.trim())
}

// 时间格式化
const formatDate = (dateString) => {
  if (!dateString) return '-'
  try {
    const date = new Date(dateString)
    return date.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit'
    })
  } catch (error) {
    return '-'
  }
}

// 编辑需求
const handleEdit = () => {
  emit('edit', requirement.value)
  handleClose()
}

// 查看子需求
const handleViewChild = (child) => {
  emit('viewChild', child)
}

// 关闭对话框
const handleClose = () => {
  requirement.value = null
  children.value = []
  emit('update:modelValue', false)
}
</script>

<style scoped>
.requirement-detail {
  max-height: 70vh;
  overflow-y: auto;
}

.detail-content {
  padding: 0 8px;
}

.info-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e4e7ed;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.info-item label {
  min-width: 100px;
  font-weight: 500;
  color: #606266;
}

.info-item .value {
  color: #303133;
  word-break: break-all;
}

.info-item .value.number {
  font-weight: 600;
  color: #409eff;
}

.description-content {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 6px;
  border-left: 4px solid #409eff;
}

.description-content.ai-content {
  background: #f0f9ff;
  border-left-color: #0ea5e9;
}

.description-content p {
  margin: 0;
  line-height: 1.6;
  color: #303133;
  white-space: pre-wrap;
}

.path-info {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 6px;
  border-left: 4px solid #67c23a;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-table th) {
  background: #f5f7fa;
  color: #606266;
  font-weight: 500;
}

:deep(.el-rate) {
  display: flex;
  align-items: center;
}

:deep(.el-rate__text) {
  margin-left: 8px;
  font-size: 12px;
  color: #606266;
}

/* 流程图Tab样式 */
.flowchart-content .flowchart-actions {
  margin-bottom: 24px;
}

.flowchart-content .flowchart-actions .action-buttons {
  display: flex;
  gap: 8px;
  margin-top: 16px;
}

.flowchart-content .flowchart-actions .action-buttons .el-button {
  border-radius: 6px;
}

.flowchart-content .flowchart-display {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  overflow: hidden;
}

.flowchart-content .flowchart-display .flowchart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: linear-gradient(135deg, #f8fafc 0%, #e3f8f8 100%);
  border-bottom: 1px solid #e4e7ed;
}

.flowchart-content .flowchart-display .flowchart-header h4 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.flowchart-content .flowchart-display .flowchart-header .flowchart-meta {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 12px;
  color: #909399;
}

.flowchart-content .flowchart-display .mermaid-code-section {
  margin: 20px;
}

.flowchart-content .flowchart-display .mermaid-code-section .code-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

.flowchart-content .flowchart-display .mermaid-code-section .code-actions .el-button {
  border-radius: 4px;
}

.flowchart-content .flowchart-display .mermaid-visualization {
  margin: 20px;
}

.flowchart-content .flowchart-display .mermaid-visualization .visualization-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.flowchart-content .flowchart-display .mermaid-visualization .visualization-header h5 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: #303133;
}

.flowchart-content .flowchart-display .mermaid-visualization .visualization-header .visualization-tools .el-button-group .el-button {
  border-radius: 4px;
}

.flowchart-content .flowchart-display .mermaid-visualization .visualization-header .visualization-tools .el-button-group .el-button:first-child {
  border-top-right-radius: 0;
  border-bottom-right-radius: 0;
}

.flowchart-content .flowchart-display .mermaid-visualization .visualization-header .visualization-tools .el-button-group .el-button:last-child {
  border-top-left-radius: 0;
  border-bottom-left-radius: 0;
}

.flowchart-content .flowchart-display .mermaid-visualization .mermaid-container {
  min-height: 300px;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  background: #fafbfc;
  overflow: hidden;
  transition: transform 0.3s ease;
}

.flowchart-content .flowchart-display .mermaid-visualization .mermaid-container .mermaid-preview {
  position: relative;
  min-height: 300px;
  padding: 20px;
  background: #f8f9fa;
}

.flowchart-content .flowchart-display .mermaid-visualization .mermaid-container .mermaid-preview pre {
  font-family: 'Monaco', 'Consolas', monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #606266;
  background: transparent;
  border: none;
  padding: 0;
  margin: 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.flowchart-content .flowchart-display .mermaid-visualization .mermaid-container .mermaid-preview .preview-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(255, 255, 255, 0.9);
  backdrop-filter: blur(2px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
}

.flowchart-content .flowchart-display .mermaid-visualization .mermaid-container .mermaid-preview .preview-overlay .preview-icon {
  font-size: 48px;
  color: #13C2C2;
  margin-bottom: 16px;
}

.flowchart-content .flowchart-display .mermaid-visualization .mermaid-container .mermaid-preview .preview-overlay p {
  font-size: 14px;
  color: #606266;
  margin: 0 0 16px 0;
}

.flowchart-content .no-flowchart {
  padding: 40px 20px;
  text-align: center;
}

.flowchart-content .no-flowchart .el-empty .el-icon {
  margin-bottom: 16px;
}
</style> 