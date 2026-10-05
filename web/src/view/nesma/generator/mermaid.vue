<template>
  <div class="mermaid-generator">
    <warning-bar title="Mermaid流程图智能生成器 - 让需求可视化更简单" />
    
    <div class="gva-table-box">
      <!-- 项目上下文选择 -->
      <div class="gva-search-box">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="当前项目">
            <el-select v-model="currentProject" placeholder="选择项目" @change="handleProjectChange">
              <el-option
                v-for="project in projects"
                :key="project.ID || project.id"
                :label="project.name"
                :value="project.ID || project.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="项目周期">
            <el-select v-model="currentCycle" placeholder="选择周期" @change="handleCycleChange" :disabled="!currentProject">
              <el-option
                v-for="cycle in cycles"
                :key="cycle.ID"
                :label="cycle.name"
                :value="cycle.ID"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="需求版本">
            <el-select v-model="currentVersion" placeholder="选择版本" @change="handleVersionChange" :disabled="!currentCycle">
              <el-option
                v-for="version in versions"
                :key="version.ID"
                :label="`${version.summary} (${version.version})`"
                :value="version.ID"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- 生成器控制区 -->
      <div class="generator-controls">
        <div class="control-section">
          <h3>生成选项</h3>
          <div class="control-row">
            <el-select v-model="generateMode" placeholder="选择生成模式" style="width: 200px; margin-right: 16px;">
              <el-option label="单个流程图" value="single" />
              <el-option label="批量生成" value="batch" />
              <el-option label="分层生成" value="hierarchical" />
            </el-select>
            
            <el-select v-model="diagramType" placeholder="流程图类型" style="width: 180px; margin-right: 16px;">
              <el-option label="业务流程图" value="flowchart" />
              <el-option label="序列图" value="sequence" />
              <el-option label="类图" value="class" />
              <el-option label="状态图" value="state" />
            </el-select>

            <el-select v-model="detailLevel" placeholder="详细程度" style="width: 150px; margin-right: 16px;">
              <el-option label="概要" value="summary" />
              <el-option label="详细" value="detailed" />
              <el-option label="完整" value="complete" />
            </el-select>
          </div>
        </div>

        <div class="action-buttons">
          <el-button type="primary" @click="generateMermaid" :loading="generating" :disabled="!canGenerate">
            <el-icon><Magic /></el-icon>
            生成流程图
          </el-button>
          <el-button type="success" @click="batchGenerate" :loading="batchGenerating" :disabled="!canGenerate">
            <el-icon><Operation /></el-icon>
            批量生成
          </el-button>
          <el-button @click="previewDiagram" :disabled="!currentDiagramCode">
            <el-icon><View /></el-icon>
            预览
          </el-button>
          <el-button @click="showHistory" :disabled="!currentCycle">
            <el-icon><Clock /></el-icon>
            查看历史
          </el-button>
        </div>
      </div>

      <!-- 需求选择区 -->
      <div class="requirement-selector" v-if="generateMode === 'single'">
        <h3>选择需求</h3>
        <el-table 
          :data="requirements" 
          @selection-change="handleRequirementSelection"
          row-key="id"
          max-height="300"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="code" label="编号" width="120" />
          <el-table-column prop="title" label="需求名称" min-width="200" />
          <el-table-column prop="level" label="层级" width="80">
            <template #default="{ row }">
              <el-tag :type="getLevelType(row.level)" size="small">
                L{{ row.level }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
        </el-table>
      </div>

      <!-- 生成结果展示区 -->
      <div class="generation-results" v-if="generationResults.length > 0">
        <h3>生成结果</h3>
        <div class="result-tabs">
          <el-tabs v-model="activeResultTab" type="card">
            <el-tab-pane 
              v-for="(result, index) in generationResults" 
              :key="index"
              :label="`流程图 ${index + 1}`"
              :name="`result-${index}`"
            >
              <div class="result-content">
                <!-- Mermaid预览区 -->
                <div class="mermaid-preview">
                  <div class="preview-header">
                    <span>{{ result.title }}</span>
                    <div class="preview-actions">
                      <el-button size="small" @click="editDiagram(result)">
                        <el-icon><Edit /></el-icon>
                        编辑
                      </el-button>
                      <el-button size="small" type="success" @click="applyDiagram(result)">
                        <el-icon><Check /></el-icon>
                        应用
                      </el-button>
                      <el-button size="small" type="info" @click="downloadDiagram(result)">
                        <el-icon><Download /></el-icon>
                        下载
                      </el-button>
                    </div>
                  </div>
                  
                  <!-- Mermaid图表展示 -->
                  <div class="mermaid-display" :id="`mermaid-${index}`">
                    {{ result.mermaidCode }}
                  </div>
                </div>

                <!-- 源码编辑区 -->
                <div class="code-editor" v-if="result.editing">
                  <el-input
                    v-model="result.mermaidCode"
                    type="textarea"
                    :rows="12"
                    placeholder="Mermaid流程图代码"
                    @input="validateMermaidCode(result)"
                  />
                  <div class="editor-actions">
                    <el-button size="small" @click="previewCode(result)">预览</el-button>
                    <el-button size="small" type="primary" @click="saveCode(result)">保存</el-button>
                    <el-button size="small" @click="cancelEdit(result)">取消</el-button>
                  </div>
                </div>

                <!-- 生成信息 -->
                <div class="generation-info">
                  <el-descriptions :column="2" size="small">
                    <el-descriptions-item label="生成时间">{{ formatDate(result.createdAt) }}</el-descriptions-item>
                    <el-descriptions-item label="图表类型">{{ result.diagramType }}</el-descriptions-item>
                    <el-descriptions-item label="复杂度">{{ result.complexity }}</el-descriptions-item>
                    <el-descriptions-item label="节点数量">{{ result.nodeCount }}</el-descriptions-item>
                  </el-descriptions>
                </div>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>

      <!-- 历史记录对话框 -->
      <el-dialog v-model="historyDialogVisible" title="Mermaid生成历史" width="80%">
        <el-table :data="historyList" max-height="400">
          <el-table-column prop="title" label="标题" min-width="200" />
          <el-table-column prop="diagramType" label="类型" width="100" />
          <el-table-column prop="createdAt" label="生成时间" width="180" :formatter="formatTableDate" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">
                {{ row.status }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="loadHistoryItem(row)">加载</el-button>
              <el-button size="small" type="info" @click="viewHistoryItem(row)">查看</el-button>
              <el-button size="small" type="danger" @click="deleteHistoryItem(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-dialog>

      <!-- 生成进度对话框 -->
      <el-dialog v-model="progressDialogVisible" title="生成进度" width="500px" :close-on-click-modal="false">
        <div class="progress-container">
          <el-progress
            :percentage="generateProgress"
            :status="progressStatus"
            :stroke-width="15"
          />
          <div class="progress-info">
            <p>{{ progressText }}</p>
            <div v-if="progressStatus === 'active'" class="progress-details">
              <div class="progress-steps">
                <div v-for="step in progressSteps" :key="step.key" 
                     :class="['step-item', { active: step.active, completed: step.completed }]">
                  <el-icon v-if="step.completed"><Check /></el-icon>
                  <el-icon v-else-if="step.active"><Loading /></el-icon>
                  <span>{{ step.text }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import {
  MagicStick, Operation, View, Clock, Edit, Check, Download, Loading
} from '@element-plus/icons-vue'

// 导入warning-bar组件
import WarningBar from '@/components/warningBar/warningBar.vue'

// API导入
import {
  getNesmaProjectList,
  getNesmaRequirementList,
  generateMermaidDiagrams,
  batchGenerateMermaidDiagrams,
  applyMermaidDiagram,
  validateMermaidDiagram,
  previewMermaidDiagram,
  getMermaidGenerationHistory,
  getMermaidGenerationStats
} from '@/api/nesma'

import { getProjectCycles } from '@/api/projectCycle'
import { getRequirementVersions } from '@/api/nesma/requirementVersion'

const router = useRouter()

// 响应式数据
const currentProject = ref(null)
const currentCycle = ref(null)
const currentVersion = ref(null)
const projects = ref([])
const cycles = ref([])
const versions = ref([])
const requirements = ref([])

// 生成器状态
const generateMode = ref('single')
const diagramType = ref('flowchart')
const detailLevel = ref('detailed')
const selectedRequirements = ref([])
const generating = ref(false)
const batchGenerating = ref(false)

// 生成结果
const generationResults = ref([])
const activeResultTab = ref('result-0')
const currentDiagramCode = ref('')

// 历史记录
const historyDialogVisible = ref(false)
const historyList = ref([])

// 进度监控
const progressDialogVisible = ref(false)
const generateProgress = ref(0)
const progressStatus = ref('active')
const progressText = ref('')
const progressSteps = ref([
  { key: 'analyzing', text: '分析需求结构', active: false, completed: false },
  { key: 'generating', text: '生成Mermaid代码', active: false, completed: false },
  { key: 'validating', text: '验证图表语法', active: false, completed: false },
  { key: 'rendering', text: '渲染预览', active: false, completed: false }
])

// 搜索表单
const searchForm = ref({})

// 计算属性
const canGenerate = computed(() => {
  return currentProject.value && currentCycle.value && currentVersion.value
})

// 方法实现
const handleProjectChange = async () => {
  currentCycle.value = null
  currentVersion.value = null
  cycles.value = []
  versions.value = []
  
  if (currentProject.value) {
    await loadCycles()
  }
}

const handleCycleChange = async () => {
  currentVersion.value = null
  versions.value = []
  
  if (currentCycle.value) {
    await loadVersions()
  }
}

const handleVersionChange = async () => {
  if (currentVersion.value) {
    await loadRequirements()
  }
}

const handleRequirementSelection = (selection) => {
  selectedRequirements.value = selection
}

// 生成单个流程图
const generateMermaid = async () => {
  if (selectedRequirements.value.length === 0) {
    ElMessage.warning('请先选择需求')
    return
  }

  try {
    generating.value = true
    progressDialogVisible.value = true
    
    // 重置进度
    resetProgress()
    updateProgress(0, '开始分析需求结构...', 'analyzing')

    const generateData = {
      cycleId: currentCycle.value,
      requirementIds: selectedRequirements.value.map(req => req.id || req.ID),
      diagramType: diagramType.value
    }

    updateProgress(25, '正在生成Mermaid代码...', 'generating')
    
    const response = await generateMermaidDiagrams(generateData)
    
    if (response.code === 0) {
      updateProgress(75, '验证图表语法...', 'validating')
      
      // 添加生成结果
      const result = {
        id: Date.now(),
        title: response.data.title || '流程图',
        mermaidCode: response.data.mermaid_code,
        diagramType: response.data.diagram_type,
        complexity: response.data.complexity,
        nodeCount: response.data.node_count,
        createdAt: new Date(),
        editing: false
      }
      
      generationResults.value.push(result)
      activeResultTab.value = `result-${generationResults.value.length - 1}`
      
      updateProgress(100, '生成完成！', 'rendering')
      
      setTimeout(() => {
        progressDialogVisible.value = false
        ElMessage.success('Mermaid流程图生成成功！')
      }, 1000)
      
    } else {
      throw new Error(response.msg || '生成失败')
    }

  } catch (error) {
    console.error('生成流程图失败:', error)
    ElMessage.error('生成失败: ' + (error.message || '未知错误'))
    progressDialogVisible.value = false
  } finally {
    generating.value = false
  }
}

// 批量生成
const batchGenerate = async () => {
  try {
    batchGenerating.value = true
    progressDialogVisible.value = true
    resetProgress()

    const batchData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      diagram_type: diagramType.value,
      detail_level: detailLevel.value,
      batch_options: {
        group_by_level: true,
        include_cross_references: true
      }
    }

    updateProgress(20, '开始批量生成...', 'analyzing')
    
    const response = await batchGenerateMermaidDiagrams(batchData)
    
    if (response.code === 0) {
      updateProgress(80, '处理生成结果...', 'generating')
      
      // 批量添加结果
      response.data.diagrams.forEach((diagram, index) => {
        const result = {
          id: Date.now() + index,
          title: diagram.title,
          mermaidCode: diagram.mermaid_code,
          diagramType: diagram.diagram_type,
          complexity: diagram.complexity,
          nodeCount: diagram.node_count,
          createdAt: new Date(),
          editing: false
        }
        generationResults.value.push(result)
      })
      
      activeResultTab.value = `result-${generationResults.value.length - response.data.diagrams.length}`
      
      updateProgress(100, `成功生成 ${response.data.diagrams.length} 个流程图！`, 'rendering')
      
      setTimeout(() => {
        progressDialogVisible.value = false
        ElMessage.success(`批量生成完成！共生成 ${response.data.diagrams.length} 个流程图`)
      }, 1000)
      
    } else {
      throw new Error(response.msg || '批量生成失败')
    }

  } catch (error) {
    console.error('批量生成失败:', error)
    ElMessage.error('批量生成失败: ' + (error.message || '未知错误'))
    progressDialogVisible.value = false
  } finally {
    batchGenerating.value = false
  }
}

// 预览流程图
const previewDiagram = async () => {
  // TODO: 实现预览功能
  ElMessage.info('预览功能开发中...')
}

// 应用流程图
const applyDiagram = async (result) => {
  try {
    const applyData = {
      requirement_id: selectedRequirements.value[0]?.id,
      mermaid_code: result.mermaidCode,
      diagram_type: result.diagramType,
      title: result.title
    }

    const response = await applyMermaidDiagram(applyData)
    
    if (response.code === 0) {
      ElMessage.success('流程图已应用到需求中')
    } else {
      throw new Error(response.msg || '应用失败')
    }

  } catch (error) {
    console.error('应用流程图失败:', error)
    ElMessage.error('应用失败: ' + (error.message || '未知错误'))
  }
}

// 显示历史记录
const showHistory = async () => {
  try {
    const response = await getMermaidGenerationHistory(currentCycle.value)
    
    if (response.code === 0) {
      historyList.value = response.data.history || []
      historyDialogVisible.value = true
    } else {
      throw new Error(response.msg || '获取历史记录失败')
    }

  } catch (error) {
    console.error('获取历史记录失败:', error)
    ElMessage.error('获取历史记录失败: ' + (error.message || '未知错误'))
  }
}

// 进度更新
const resetProgress = () => {
  generateProgress.value = 0
  progressStatus.value = 'active'
  progressSteps.value.forEach(step => {
    step.active = false
    step.completed = false
  })
}

const updateProgress = (percentage, text, activeStep) => {
  generateProgress.value = percentage
  progressText.value = text
  
  progressSteps.value.forEach(step => {
    if (step.key === activeStep) {
      step.active = true
    } else if (step.completed || generateProgress.value > 50) {
      step.completed = true
      step.active = false
    }
  })
}

// 辅助方法
const getLevelType = (level) => {
  const levelMap = {
    1: 'danger',
    2: 'warning', 
    3: 'primary',
    4: 'success'
  }
  return levelMap[level] || 'info'
}

const getStatusType = (status) => {
  const statusMap = {
    'completed': 'success',
    'processing': 'warning',
    'failed': 'danger'
  }
  return statusMap[status] || 'info'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const formatTableDate = (row, column, cellValue) => {
  return formatDate(cellValue)
}

// 数据加载方法
const loadProjects = async () => {
  try {
    const response = await getNesmaProjectList({ 
      page: 1, 
      pageSize: 100 
    })
    
    if (response.code === 0) {
      projects.value = response.data.list || []
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
    ElMessage.error('加载项目列表失败')
  }
}

const loadCycles = async () => {
  if (!currentProject.value) return
  
  try {
    const response = await getProjectCycles(currentProject.value)
    
    if (response.code === 0) {
      cycles.value = response.data.cycles || []
    }
  } catch (error) {
    console.error('加载周期列表失败:', error)
    ElMessage.error('加载周期列表失败')
  }
}

const loadVersions = async () => {
  if (!currentCycle.value) return
  
  try {
    const response = await getRequirementVersions(currentCycle.value)
    
    if (response.code === 0) {
      versions.value = response.data.versions || []
    }
  } catch (error) {
    console.error('加载版本列表失败:', error)
    ElMessage.error('加载版本列表失败')
  }
}

const loadRequirements = async () => {
  if (!currentVersion.value) return
  
  try {
    const params = {
      version_id: currentVersion.value,
      page: 1,
      pageSize: 1000
    }
    
    const response = await getNesmaRequirementList(params)
    
    if (response.code === 0) {
      requirements.value = response.data.list || []
    }
  } catch (error) {
    console.error('加载需求列表失败:', error)
    ElMessage.error('加载需求列表失败')
  }
}

// 编辑相关方法
const editDiagram = (result) => {
  result.editing = true
}

const saveCode = (result) => {
  result.editing = false
  ElMessage.success('代码已保存')
}

const cancelEdit = (result) => {
  result.editing = false
}

const validateMermaidCode = (result) => {
  // TODO: 实现Mermaid代码验证
}

const downloadDiagram = (result) => {
  const blob = new Blob([result.mermaidCode], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${result.title}.mmd`
  a.click()
  URL.revokeObjectURL(url)
}

// 历史记录操作
const loadHistoryItem = (item) => {
  const result = {
    id: Date.now(),
    title: item.title,
    mermaidCode: item.mermaid_code,
    diagramType: item.diagram_type,
    complexity: item.complexity,
    nodeCount: item.node_count,
    createdAt: new Date(item.created_at),
    editing: false
  }
  
  generationResults.value.push(result)
  activeResultTab.value = `result-${generationResults.value.length - 1}`
  historyDialogVisible.value = false
  
  ElMessage.success('历史记录已加载')
}

const viewHistoryItem = (item) => {
  // TODO: 实现查看详情
  ElMessage.info('查看详情功能开发中...')
}

const deleteHistoryItem = async (item) => {
  try {
    await ElMessageBox.confirm('确定要删除这个历史记录吗？', '确认删除', {
      type: 'warning'
    })
    
    // TODO: 调用删除API
    ElMessage.success('删除成功')
    showHistory() // 重新加载
    
  } catch (error) {
    // 用户取消删除
  }
}

// 组件挂载
onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
.mermaid-generator {
  padding: 20px;
}

.generator-controls {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  margin: 20px 0;
}

.control-section h3 {
  margin: 0 0 15px 0;
  color: #333;
}

.control-row {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.requirement-selector {
  margin: 20px 0;
}

.requirement-selector h3 {
  margin-bottom: 15px;
  color: #333;
}

.generation-results {
  margin: 20px 0;
}

.generation-results h3 {
  margin-bottom: 15px;
  color: #333;
}

.result-content {
  padding: 20px;
}

.mermaid-preview {
  border: 1px solid #e1e5e9;
  border-radius: 8px;
  margin-bottom: 20px;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8f9fa;
  border-bottom: 1px solid #e1e5e9;
  font-weight: 500;
}

.preview-actions {
  display: flex;
  gap: 8px;
}

.mermaid-display {
  padding: 20px;
  background: white;
  min-height: 200px;
  font-family: 'Courier New', monospace;
  white-space: pre-wrap;
  border-radius: 0 0 8px 8px;
}

.code-editor {
  margin: 20px 0;
}

.editor-actions {
  margin-top: 12px;
  display: flex;
  gap: 8px;
}

.generation-info {
  margin-top: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.progress-container {
  padding: 20px;
}

.progress-info {
  margin-top: 20px;
  text-align: center;
}

.progress-steps {
  margin-top: 20px;
}

.step-item {
  display: flex;
  align-items: center;
  margin: 8px 0;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.3s;
}

.step-item.active {
  background: #e3f2fd;
  color: #1976d2;
}

.step-item.completed {
  background: #e8f5e8;
  color: #388e3c;
}

.step-item .el-icon {
  margin-right: 8px;
}
</style>