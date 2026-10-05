<template>
  <div class="level3-analyzer">
    <warning-bar title="L3需求智能分析器 - 深度分析和优化L3级功能需求" />
    
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

      <!-- 分析配置区 -->
      <div class="analyzer-config">
        <div class="config-section">
          <h3>分析配置</h3>
          <div class="config-row">
            <el-select v-model="analysisType" placeholder="分析类型" style="width: 180px; margin-right: 16px;">
              <el-option label="完整性分析" value="completeness" />
              <el-option label="优化分析" value="optimization" />
              <el-option label="扩展分析" value="expansion" />
              <el-option label="综合分析" value="comprehensive" />
            </el-select>
            
            <el-select v-model="analysisDepth" placeholder="分析深度" style="width: 150px; margin-right: 16px;">
              <el-option label="快速" value="quick" />
              <el-option label="标准" value="standard" />
              <el-option label="深度" value="deep" />
            </el-select>

            <el-switch
              v-model="useKnowledgeBase"
              active-text="使用知识库"
              inactive-text="不使用知识库"
              style="margin-right: 16px;"
            />

            <el-switch
              v-model="generateExpansions"
              active-text="生成扩展建议"
              inactive-text="仅优化现有"
              style="margin-right: 16px;"
            />

            <el-switch
              v-model="includeNESMAScoring"
              active-text="NESMA评分"
              inactive-text="基础分析"
            />
          </div>
        </div>

        <div class="action-buttons">
          <el-button type="primary" @click="startAnalysis" :loading="analyzing" :disabled="!canAnalyze">
            <el-icon><Search /></el-icon>
            开始分析
          </el-button>
          <el-button type="success" @click="batchAnalyze" :loading="batchAnalyzing" :disabled="!canAnalyze">
            <el-icon><Operation /></el-icon>
            批量分析
          </el-button>
          <el-button @click="showHistory" :disabled="!currentCycle">
            <el-icon><Clock /></el-icon>
            分析历史
          </el-button>
          <el-button @click="showStats" :disabled="!currentCycle">
            <el-icon><DataAnalysis /></el-icon>
            分析统计
          </el-button>
        </div>
      </div>

      <!-- L3需求选择器 -->
      <div class="l3-requirement-selector">
        <h3>选择L3需求进行分析 ({{ selectedL3Requirements.length }} 个已选择)</h3>
        <div class="filter-controls">
          <el-input 
            v-model="requirementFilter" 
            placeholder="搜索需求..." 
            style="width: 300px; margin-right: 16px;"
            clearable
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          
          <el-select v-model="functionTypeFilter" placeholder="功能类型" clearable style="width: 120px; margin-right: 16px;">
            <el-option label="EI" value="EI" />
            <el-option label="EO" value="EO" />
            <el-option label="EQ" value="EQ" />
            <el-option label="ILF" value="ILF" />
            <el-option label="EIF" value="EIF" />
          </el-select>

          <el-select v-model="complexityFilter" placeholder="复杂度" clearable style="width: 120px; margin-right: 16px;">
            <el-option label="简单" value="simple" />
            <el-option label="中等" value="moderate" />
            <el-option label="复杂" value="complex" />
          </el-select>

          <el-button @click="selectAllFiltered">全选</el-button>
          <el-button @click="clearSelection">清空选择</el-button>
        </div>

        <el-table 
          :data="filteredL3Requirements" 
          @selection-change="handleL3Selection"
          row-key="id"
          max-height="400"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="code" label="编号" width="120" />
          <el-table-column prop="title" label="需求名称" min-width="250" />
          <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
          <el-table-column prop="function_type" label="功能类型" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.function_type" size="small" :type="getFunctionTypeColor(row.function_type)">
                {{ row.function_type }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="complexity" label="复杂度" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.complexity" :type="getComplexityType(row.complexity)" size="small">
                {{ row.complexity }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="分析状态" width="120">
            <template #default="{ row }">
              <el-tag v-if="row.analysis_status" :type="getAnalysisStatusType(row.analysis_status)" size="small">
                {{ getAnalysisStatusText(row.analysis_status) }}
              </el-tag>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="analyzeSingle(row)" :loading="row.analyzing">
                分析
              </el-button>
              <el-button size="small" type="info" @click="viewDetails(row)" v-if="row.analysis_result">
                查看详情
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分析结果展示 -->
      <div class="analysis-results" v-if="analysisResults.length > 0">
        <h3>分析结果 ({{ analysisResults.length }} 个L3需求)</h3>
        
        <div class="result-summary">
          <el-row :gutter="20">
            <el-col :span="6">
              <div class="summary-card">
                <div class="summary-number">{{ totalOptimizations }}</div>
                <div class="summary-label">优化建议</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="summary-card">
                <div class="summary-number">{{ totalExpansions }}</div>
                <div class="summary-label">扩展建议</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="summary-card">
                <div class="summary-number">{{ averageScore.toFixed(1) }}</div>
                <div class="summary-label">平均得分</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="summary-card">
                <div class="summary-number">{{ appliedCount }}</div>
                <div class="summary-label">已应用</div>
              </div>
            </el-col>
          </el-row>
        </div>

        <div class="result-actions">
          <el-button type="success" @click="applyAllOptimizations" :disabled="!hasValidOptimizations">
            <el-icon><Check /></el-icon>
            应用全部优化
          </el-button>
          <el-button type="primary" @click="createAllExpansions" :disabled="!hasValidExpansions">
            <el-icon><Plus /></el-icon>
            创建全部扩展
          </el-button>
          <el-button @click="exportAnalysisReport">
            <el-icon><Download /></el-icon>
            导出分析报告
          </el-button>
          <el-button @click="clearResults">
            <el-icon><Delete /></el-icon>
            清空结果
          </el-button>
        </div>

        <el-collapse v-model="activeResultPanels" accordion>
          <el-collapse-item 
            v-for="(result, index) in analysisResults" 
            :key="index"
            :title="getResultTitle(result)"
            :name="`result-${index}`"
          >
            <div class="analysis-result-content">
              <!-- 基本信息 -->
              <div class="result-basic-info">
                <el-descriptions :column="3" size="small" border>
                  <el-descriptions-item label="需求编号">{{ result.requirement.code }}</el-descriptions-item>
                  <el-descriptions-item label="分析类型">{{ result.analysis_type }}</el-descriptions-item>
                  <el-descriptions-item label="分析时间">{{ formatDate(result.analyzed_at) }}</el-descriptions-item>
                  <el-descriptions-item label="分析得分">
                    <el-rate v-model="result.analysis_score" :max="5" disabled show-score />
                  </el-descriptions-item>
                  <el-descriptions-item label="置信度">{{ (result.confidence * 100).toFixed(1) }}%</el-descriptions-item>
                  <el-descriptions-item label="状态">
                    <el-tag :type="getAnalysisStatusType(result.status)">
                      {{ getAnalysisStatusText(result.status) }}
                    </el-tag>
                  </el-descriptions-item>
                </el-descriptions>
              </div>

              <!-- 优化建议 -->
              <div class="optimization-suggestions" v-if="result.optimizations && result.optimizations.length > 0">
                <h4>优化建议</h4>
                <div class="suggestion-list">
                  <div 
                    v-for="(optimization, oIndex) in result.optimizations" 
                    :key="oIndex"
                    class="suggestion-item"
                    :class="{ selected: optimization.selected }"
                  >
                    <div class="suggestion-header">
                      <el-checkbox 
                        v-model="optimization.selected" 
                        @change="updateOptimizationSelection(result, oIndex)"
                      />
                      <span class="suggestion-type">{{ optimization.type }}</span>
                      <el-tag size="small" :type="getSuggestionPriorityType(optimization.priority)">
                        {{ optimization.priority }}
                      </el-tag>
                    </div>
                    
                    <div class="suggestion-content">
                      <p class="suggestion-description">{{ optimization.description }}</p>
                      
                      <div class="suggestion-comparison" v-if="optimization.before && optimization.after">
                        <div class="comparison-section">
                          <h5>优化前:</h5>
                          <div class="before-content">{{ optimization.before }}</div>
                        </div>
                        <div class="comparison-section">
                          <h5>优化后:</h5>
                          <div class="after-content">{{ optimization.after }}</div>
                        </div>
                      </div>

                      <div class="suggestion-impact" v-if="optimization.impact">
                        <el-icon><TrendCharts /></el-icon>
                        <span>预期效果: {{ optimization.impact }}</span>
                      </div>
                    </div>

                    <div class="suggestion-actions">
                      <el-button size="small" @click="editOptimization(result, oIndex)">
                        <el-icon><Edit /></el-icon>
                        编辑
                      </el-button>
                      <el-button size="small" type="success" @click="applySingleOptimization(result, oIndex)" :disabled="!optimization.selected">
                        <el-icon><Check /></el-icon>
                        应用
                      </el-button>
                      <el-button size="small" type="danger" @click="removeOptimization(result, oIndex)">
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 扩展建议 -->
              <div class="expansion-suggestions" v-if="result.expansions && result.expansions.length > 0">
                <h4>扩展建议</h4>
                <div class="expansion-grid">
                  <div 
                    v-for="(expansion, eIndex) in result.expansions" 
                    :key="eIndex"
                    class="expansion-card"
                    :class="{ selected: expansion.selected }"
                  >
                    <div class="expansion-header">
                      <el-checkbox 
                        v-model="expansion.selected" 
                        @change="updateExpansionSelection(result, eIndex)"
                      />
                      <span class="expansion-title">{{ expansion.suggested_title }}</span>
                    </div>
                    
                    <div class="expansion-content">
                      <p class="expansion-description">{{ expansion.suggested_description }}</p>
                      
                      <div class="expansion-meta">
                        <el-row :gutter="12">
                          <el-col :span="8">
                            <div class="meta-item">
                              <span class="label">功能类型:</span>
                              <span class="value">{{ expansion.function_type }}</span>
                            </div>
                          </el-col>
                          <el-col :span="8">
                            <div class="meta-item">
                              <span class="label">预估复杂度:</span>
                              <span class="value">{{ expansion.estimated_complexity }}</span>
                            </div>
                          </el-col>
                          <el-col :span="8">
                            <div class="meta-item">
                              <span class="label">业务价值:</span>
                              <span class="value">{{ expansion.business_value }}</span>
                            </div>
                          </el-col>
                        </el-row>
                      </div>

                      <div class="expansion-reason" v-if="expansion.expansion_reason">
                        <el-icon><InfoFilled /></el-icon>
                        <span>扩展理由: {{ expansion.expansion_reason }}</span>
                      </div>
                    </div>

                    <div class="expansion-actions">
                      <el-button size="small" @click="editExpansion(result, eIndex)">
                        <el-icon><Edit /></el-icon>
                        编辑
                      </el-button>
                      <el-button size="small" type="primary" @click="createSingleExpansion(result, eIndex)" :disabled="!expansion.selected">
                        <el-icon><Plus /></el-icon>
                        创建需求
                      </el-button>
                      <el-button size="small" type="danger" @click="removeExpansion(result, eIndex)">
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- NESMA评分详情 -->
              <div class="nesma-scoring" v-if="result.nesma_scoring">
                <h4>NESMA评分详情</h4>
                <div class="scoring-details">
                  <el-row :gutter="20">
                    <el-col :span="12">
                      <div class="score-section">
                        <h5>功能点评估</h5>
                        <el-descriptions :column="1" size="small">
                          <el-descriptions-item label="当前AFP">{{ result.nesma_scoring.current_afp }}</el-descriptions-item>
                          <el-descriptions-item label="建议AFP">{{ result.nesma_scoring.suggested_afp }}</el-descriptions-item>
                          <el-descriptions-item label="当前UFP">{{ result.nesma_scoring.current_ufp }}</el-descriptions-item>
                          <el-descriptions-item label="建议UFP">{{ result.nesma_scoring.suggested_ufp }}</el-descriptions-item>
                        </el-descriptions>
                      </div>
                    </el-col>
                    <el-col :span="12">
                      <div class="score-section">
                        <h5>质量评估</h5>
                        <el-descriptions :column="1" size="small">
                          <el-descriptions-item label="完整性">{{ result.nesma_scoring.completeness_score }}%</el-descriptions-item>
                          <el-descriptions-item label="一致性">{{ result.nesma_scoring.consistency_score }}%</el-descriptions-item>
                          <el-descriptions-item label="可追溯性">{{ result.nesma_scoring.traceability_score }}%</el-descriptions-item>
                          <el-descriptions-item label="可测试性">{{ result.nesma_scoring.testability_score }}%</el-descriptions-item>
                        </el-descriptions>
                      </div>
                    </el-col>
                  </el-row>
                </div>
              </div>

              <!-- 批量操作 -->
              <div class="batch-actions">
                <el-button @click="selectAllOptimizations(result)">全选优化</el-button>
                <el-button @click="selectAllExpansions(result)">全选扩展</el-button>
                <el-button type="success" @click="applySelectedOptimizations(result)" :disabled="!hasSelectedOptimizations(result)">
                  应用选中优化 ({{ getSelectedOptimizationCount(result) }})
                </el-button>
                <el-button type="primary" @click="createSelectedExpansions(result)" :disabled="!hasSelectedExpansions(result)">
                  创建选中扩展 ({{ getSelectedExpansionCount(result) }})
                </el-button>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>

      <!-- 分析进度对话框 -->
      <el-dialog v-model="progressDialogVisible" title="L3需求分析进度" width="600px" :close-on-click-modal="false">
        <div class="progress-container">
          <el-progress
            :percentage="analysisProgress"
            :status="progressStatus"
            :stroke-width="15"
          />
          <div class="progress-info">
            <p>{{ progressText }}</p>
            <div v-if="progressStatus === 'active'" class="progress-details">
              <div class="current-requirement" v-if="currentProcessingRequirement">
                正在分析: {{ currentProcessingRequirement.title }}
              </div>
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

      <!-- 历史记录对话框 -->
      <el-dialog v-model="historyDialogVisible" title="L3分析历史" width="80%">
        <el-table :data="historyList" max-height="400">
          <el-table-column prop="requirement_title" label="需求名称" min-width="200" />
          <el-table-column prop="analysis_type" label="分析类型" width="120" />
          <el-table-column prop="optimization_count" label="优化建议" width="100" />
          <el-table-column prop="expansion_count" label="扩展建议" width="100" />
          <el-table-column prop="analysis_score" label="分析得分" width="100" />
          <el-table-column prop="created_at" label="分析时间" width="180" :formatter="formatTableDate" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="loadHistoryItem(row)">加载</el-button>
              <el-button size="small" type="info" @click="viewHistoryDetail(row)">详情</el-button>
              <el-button size="small" type="danger" @click="deleteHistoryItem(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-dialog>

      <!-- 统计数据对话框 -->
      <el-dialog v-model="statsDialogVisible" title="L3分析统计" width="70%">
        <div class="stats-container">
          <el-row :gutter="20">
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.total_analyzed }}</div>
                <div class="stat-label">总分析数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.total_optimized }}</div>
                <div class="stat-label">已优化数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.total_expanded }}</div>
                <div class="stat-label">已扩展数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.avg_score }}</div>
                <div class="stat-label">平均分析得分</div>
              </div>
            </el-col>
          </el-row>
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
  Search, Operation, Clock, DataAnalysis, Check, Download, Delete, Edit, Loading,
  Plus, TrendCharts, InfoFilled
} from '@element-plus/icons-vue'

// 导入warning-bar组件
import WarningBar from '@/components/warningBar/warningBar.vue'

// API导入
import {
  getNesmaProjectList,
  getNesmaRequirementList,
  analyzeLevel3Requirements,
  applyOptimizationSuggestion,
  createExpansionRequirement,
  batchApplyLevel3Optimizations,
  batchCreateExpansions,
  getLevel3AnalysisHistory,
  getLevel3AnalysisStats
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
const l3Requirements = ref([])

// 分析配置
const analysisType = ref('comprehensive')
const analysisDepth = ref('standard')
const useKnowledgeBase = ref(true)
const generateExpansions = ref(true)
const includeNESMAScoring = ref(true)
const selectedL3Requirements = ref([])

// 过滤器
const requirementFilter = ref('')
const functionTypeFilter = ref('')
const complexityFilter = ref('')

// 分析状态
const analyzing = ref(false)
const batchAnalyzing = ref(false)

// 分析结果
const analysisResults = ref([])
const activeResultPanels = ref([])

// 进度监控
const progressDialogVisible = ref(false)
const analysisProgress = ref(0)
const progressStatus = ref('active')
const progressText = ref('')
const currentProcessingRequirement = ref(null)
const progressSteps = ref([
  { key: 'preparing', text: '准备分析环境', active: false, completed: false },
  { key: 'analyzing', text: '深度分析需求', active: false, completed: false },
  { key: 'optimizing', text: '生成优化建议', active: false, completed: false },
  { key: 'expanding', text: '生成扩展建议', active: false, completed: false },
  { key: 'scoring', text: 'NESMA评分', active: false, completed: false }
])

// 历史和统计
const historyDialogVisible = ref(false)
const historyList = ref([])
const statsDialogVisible = ref(false)
const stats = ref({
  total_analyzed: 0,
  total_optimized: 0,
  total_expanded: 0,
  avg_score: 0
})

// 搜索表单
const searchForm = ref({})

// 计算属性
const canAnalyze = computed(() => {
  return currentProject.value && currentCycle.value && currentVersion.value
})

const filteredL3Requirements = computed(() => {
  let filtered = l3Requirements.value

  if (requirementFilter.value) {
    const filter = requirementFilter.value.toLowerCase()
    filtered = filtered.filter(req => 
      req.title.toLowerCase().includes(filter) || 
      req.description.toLowerCase().includes(filter) ||
      req.code.toLowerCase().includes(filter)
    )
  }

  if (functionTypeFilter.value) {
    filtered = filtered.filter(req => req.function_type === functionTypeFilter.value)
  }

  if (complexityFilter.value) {
    filtered = filtered.filter(req => req.complexity === complexityFilter.value)
  }

  return filtered
})

const totalOptimizations = computed(() => {
  return analysisResults.value.reduce((total, result) => 
    total + (result.optimizations ? result.optimizations.length : 0), 0
  )
})

const totalExpansions = computed(() => {
  return analysisResults.value.reduce((total, result) => 
    total + (result.expansions ? result.expansions.length : 0), 0
  )
})

const averageScore = computed(() => {
  if (analysisResults.value.length === 0) return 0
  const totalScore = analysisResults.value.reduce((total, result) => total + result.analysis_score, 0)
  return totalScore / analysisResults.value.length
})

const appliedCount = computed(() => {
  return analysisResults.value.filter(result => result.status === 'applied').length
})

const hasValidOptimizations = computed(() => {
  return analysisResults.value.some(result => 
    result.optimizations && result.optimizations.some(opt => opt.selected)
  )
})

const hasValidExpansions = computed(() => {
  return analysisResults.value.some(result => 
    result.expansions && result.expansions.some(exp => exp.selected)
  )
})

// 方法实现
const handleProjectChange = async () => {
  currentCycle.value = null
  currentVersion.value = null
  cycles.value = []
  versions.value = []
  l3Requirements.value = []
  if (currentProject.value) {
    await loadCycles()
  }
}

const handleCycleChange = async () => {
  currentVersion.value = null
  versions.value = []
  l3Requirements.value = []
  
  if (currentCycle.value) {
    await loadVersions()
  }
}

const handleVersionChange = async () => {
  if (currentVersion.value) {
    await loadL3Requirements()
  }
}

const handleL3Selection = (selection) => {
  selectedL3Requirements.value = selection
}

const selectAllFiltered = () => {
  selectedL3Requirements.value = [...filteredL3Requirements.value]
}

const clearSelection = () => {
  selectedL3Requirements.value = []
}

// 主要分析方法
const startAnalysis = async () => {
  if (selectedL3Requirements.value.length === 0) {
    ElMessage.warning('请先选择要分析的L3需求')
    return
  }

  try {
    analyzing.value = true
    progressDialogVisible.value = true
    
    // 重置进度
    resetProgress()
    updateProgress(0, '准备分析环境...', 'preparing')

    const analysisData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      requirement_ids: selectedL3Requirements.value.map(req => req.id || req.ID),
      analysis_type: analysisType.value,
      analysis_depth: analysisDepth.value,
      use_knowledge_base: useKnowledgeBase.value,
      generate_expansions: generateExpansions.value,
      include_nesma_scoring: includeNESMAScoring.value,
      options: {
        generate_mermaid: false,
        detailed_report: true,
        include_recommendations: true
      }
    }

    updateProgress(20, '开始深度分析需求...', 'analyzing')
    
    const response = await analyzeLevel3Requirements(analysisData)
    
    if (response.code === 0) {
      updateProgress(60, '生成优化建议...', 'optimizing')
      
      // 处理分析结果
      const results = response.data.results.map(result => ({
        requirement: result.requirement,
        analysis_type: result.analysis_type,
        analysis_score: result.analysis_score,
        confidence: result.confidence,
        status: 'analyzed',
        analyzed_at: new Date(),
        optimizations: result.optimizations ? result.optimizations.map(opt => ({
          ...opt,
          selected: true
        })) : [],
        expansions: result.expansions ? result.expansions.map(exp => ({
          ...exp,
          selected: true
        })) : [],
        nesma_scoring: result.nesma_scoring
      }))
      
      analysisResults.value.push(...results)
      
      // 展开第一个结果面板
      if (results.length > 0) {
        activeResultPanels.value = [`result-${analysisResults.value.length - results.length}`]
      }
      
      updateProgress(100, `分析完成！共分析 ${results.length} 个L3需求`, 'scoring')
      
      setTimeout(() => {
        progressDialogVisible.value = false
        ElMessage.success(`L3需求分析完成！共生成 ${totalOptimizations.value} 个优化建议和 ${totalExpansions.value} 个扩展建议`)
      }, 1000)
      
    } else {
      throw new Error(response.msg || '分析失败')
    }

  } catch (error) {
    console.error('L3需求分析失败:', error)
    ElMessage.error('分析失败: ' + (error.message || '未知错误'))
    progressDialogVisible.value = false
  } finally {
    analyzing.value = false
  }
}

const batchAnalyze = async () => {
  try {
    batchAnalyzing.value = true
    progressDialogVisible.value = true
    resetProgress()

    const batchData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      analysis_type: analysisType.value,
      analysis_depth: analysisDepth.value,
      use_knowledge_base: useKnowledgeBase.value,
      generate_expansions: generateExpansions.value,
      include_nesma_scoring: includeNESMAScoring.value
    }

    updateProgress(20, '开始批量分析...', 'preparing')
    
    const response = await batchApplyLevel3Optimizations(batchData)
    
    if (response.code === 0) {
      updateProgress(80, '处理批量分析结果...', 'analyzing')
      
      // 处理批量结果
      const results = response.data.results.map(result => ({
        requirement: result.requirement,
        analysis_type: result.analysis_type,
        analysis_score: result.analysis_score,
        confidence: result.confidence,
        status: 'analyzed',
        analyzed_at: new Date(),
        optimizations: result.optimizations ? result.optimizations.map(opt => ({
          ...opt,
          selected: true
        })) : [],
        expansions: result.expansions ? result.expansions.map(exp => ({
          ...exp,
          selected: true
        })) : [],
        nesma_scoring: result.nesma_scoring
      }))
      
      analysisResults.value.push(...results)
      
      updateProgress(100, `批量分析完成！共分析 ${results.length} 个L3需求`, 'scoring')
      
      setTimeout(() => {
        progressDialogVisible.value = false
        ElMessage.success(`批量分析完成！共处理 ${response.data.total_count} 个L3需求`)
      }, 1000)
      
    } else {
      throw new Error(response.msg || '批量分析失败')
    }

  } catch (error) {
    console.error('批量分析失败:', error)
    ElMessage.error('批量分析失败: ' + (error.message || '未知错误'))
    progressDialogVisible.value = false
  } finally {
    batchAnalyzing.value = false
  }
}

const analyzeSingle = async (requirement) => {
  try {
    requirement.analyzing = true
    
    const analysisData = {
      requirement_ids: [requirement.id],
      analysis_type: analysisType.value,
      analysis_depth: analysisDepth.value,
      use_knowledge_base: useKnowledgeBase.value,
      generate_expansions: generateExpansions.value,
      include_nesma_scoring: includeNESMAScoring.value
    }

    const response = await analyzeLevel3Requirements(analysisData)
    
    if (response.code === 0 && response.data.results.length > 0) {
      const result = response.data.results[0]
      
      const analysisResult = {
        requirement: result.requirement,
        analysis_type: result.analysis_type,
        analysis_score: result.analysis_score,
        confidence: result.confidence,
        status: 'analyzed',
        analyzed_at: new Date(),
        optimizations: result.optimizations ? result.optimizations.map(opt => ({
          ...opt,
          selected: true
        })) : [],
        expansions: result.expansions ? result.expansions.map(exp => ({
          ...exp,
          selected: true
        })) : [],
        nesma_scoring: result.nesma_scoring
      }
      
      analysisResults.value.push(analysisResult)
      requirement.analysis_result = analysisResult
      requirement.analysis_status = 'analyzed'
      
      ElMessage.success('单个需求分析完成')
    } else {
      throw new Error(response.msg || '分析失败')
    }

  } catch (error) {
    console.error('单个需求分析失败:', error)
    ElMessage.error('分析失败: ' + (error.message || '未知错误'))
  } finally {
    requirement.analyzing = false
  }
}

// 建议操作方法
const updateOptimizationSelection = (result, index) => {
  // 优化建议选择变更
}

const updateExpansionSelection = (result, index) => {
  // 扩展建议选择变更
}

const selectAllOptimizations = (result) => {
  if (result.optimizations) {
    result.optimizations.forEach(opt => opt.selected = true)
  }
}

const selectAllExpansions = (result) => {
  if (result.expansions) {
    result.expansions.forEach(exp => exp.selected = true)
  }
}

const hasSelectedOptimizations = (result) => {
  return result.optimizations && result.optimizations.some(opt => opt.selected)
}

const hasSelectedExpansions = (result) => {
  return result.expansions && result.expansions.some(exp => exp.selected)
}

const getSelectedOptimizationCount = (result) => {
  return result.optimizations ? result.optimizations.filter(opt => opt.selected).length : 0
}

const getSelectedExpansionCount = (result) => {
  return result.expansions ? result.expansions.filter(exp => exp.selected).length : 0
}

// 应用和创建方法
const applySingleOptimization = async (result, index) => {
  const optimization = result.optimizations[index]
  
  try {
    const applyData = {
      requirement_id: result.requirement.id,
      optimization: optimization
    }

    const response = await applyOptimizationSuggestion(applyData)
    
    if (response.code === 0) {
      ElMessage.success('优化建议已应用')
      optimization.applied = true
    } else {
      throw new Error(response.msg || '应用失败')
    }

  } catch (error) {
    console.error('应用优化建议失败:', error)
    ElMessage.error('应用失败: ' + (error.message || '未知错误'))
  }
}

const createSingleExpansion = async (result, index) => {
  const expansion = result.expansions[index]
  
  try {
    const createData = {
      parent_id: result.requirement.id,
      expansion: expansion
    }

    const response = await createExpansionRequirement(createData)
    
    if (response.code === 0) {
      ElMessage.success('扩展需求已创建')
      expansion.created = true
    } else {
      throw new Error(response.msg || '创建失败')
    }

  } catch (error) {
    console.error('创建扩展需求失败:', error)
    ElMessage.error('创建失败: ' + (error.message || '未知错误'))
  }
}

const applySelectedOptimizations = async (result) => {
  const selectedOptimizations = result.optimizations.filter(opt => opt.selected)
  
  if (selectedOptimizations.length === 0) {
    ElMessage.warning('请先选择要应用的优化建议')
    return
  }

  try {
    const applyData = {
      requirement_id: result.requirement.id,
      optimizations: selectedOptimizations
    }

    const response = await batchApplyOptimizations(applyData)
    
    if (response.code === 0) {
      ElMessage.success(`成功应用 ${selectedOptimizations.length} 个优化建议`)
      selectedOptimizations.forEach(opt => opt.applied = true)
    } else {
      throw new Error(response.msg || '批量应用失败')
    }

  } catch (error) {
    console.error('批量应用优化失败:', error)
    ElMessage.error('批量应用失败: ' + (error.message || '未知错误'))
  }
}

const createSelectedExpansions = async (result) => {
  const selectedExpansions = result.expansions.filter(exp => exp.selected)
  
  if (selectedExpansions.length === 0) {
    ElMessage.warning('请先选择要创建的扩展建议')
    return
  }

  try {
    const createData = {
      parent_id: result.requirement.id,
      expansions: selectedExpansions
    }

    const response = await batchCreateExpansions(createData)
    
    if (response.code === 0) {
      ElMessage.success(`成功创建 ${selectedExpansions.length} 个扩展需求`)
      selectedExpansions.forEach(exp => exp.created = true)
    } else {
      throw new Error(response.msg || '批量创建失败')
    }

  } catch (error) {
    console.error('批量创建扩展失败:', error)
    ElMessage.error('批量创建失败: ' + (error.message || '未知错误'))
  }
}

const applyAllOptimizations = async () => {
  const allOptimizations = []
  
  analysisResults.value.forEach(result => {
    if (result.optimizations) {
      result.optimizations.forEach(opt => {
        if (opt.selected) {
          allOptimizations.push({
            requirement_id: result.requirement.id,
            optimization: opt
          })
        }
      })
    }
  })

  if (allOptimizations.length === 0) {
    ElMessage.warning('没有选中的优化建议')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要应用 ${allOptimizations.length} 个优化建议吗？`,
      '批量应用确认',
      { type: 'warning' }
    )

    const response = await batchApplyOptimizations({ optimizations: allOptimizations })
    
    if (response.code === 0) {
      ElMessage.success(`成功应用 ${allOptimizations.length} 个优化建议`)
      
      // 标记为已应用
      analysisResults.value.forEach(result => {
        if (result.optimizations) {
          result.optimizations.forEach(opt => {
            if (opt.selected) {
              opt.applied = true
            }
          })
        }
      })
    } else {
      throw new Error(response.msg || '批量应用失败')
    }

  } catch (error) {
    if (error === 'cancel') return
    console.error('批量应用失败:', error)
    ElMessage.error('批量应用失败: ' + (error.message || '未知错误'))
  }
}

const createAllExpansions = async () => {
  const allExpansions = []
  
  analysisResults.value.forEach(result => {
    if (result.expansions) {
      result.expansions.forEach(exp => {
        if (exp.selected) {
          allExpansions.push({
            parent_id: result.requirement.id,
            expansion: exp
          })
        }
      })
    }
  })

  if (allExpansions.length === 0) {
    ElMessage.warning('没有选中的扩展建议')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要创建 ${allExpansions.length} 个扩展需求吗？`,
      '批量创建确认',
      { type: 'warning' }
    )

    const response = await batchCreateExpansions({ expansions: allExpansions })
    
    if (response.code === 0) {
      ElMessage.success(`成功创建 ${allExpansions.length} 个扩展需求`)
      
      // 标记为已创建
      analysisResults.value.forEach(result => {
        if (result.expansions) {
          result.expansions.forEach(exp => {
            if (exp.selected) {
              exp.created = true
            }
          })
        }
      })
    } else {
      throw new Error(response.msg || '批量创建失败')
    }

  } catch (error) {
    if (error === 'cancel') return
    console.error('批量创建失败:', error)
    ElMessage.error('批量创建失败: ' + (error.message || '未知错误'))
  }
}

// 编辑和删除
const editOptimization = (result, index) => {
  // TODO: 实现优化建议编辑功能
  ElMessage.info('编辑优化建议功能开发中...')
}

const editExpansion = (result, index) => {
  // TODO: 实现扩展建议编辑功能
  ElMessage.info('编辑扩展建议功能开发中...')
}

const removeOptimization = async (result, index) => {
  try {
    await ElMessageBox.confirm('确定要删除这个优化建议吗？', '确认删除', {
      type: 'warning'
    })
    
    result.optimizations.splice(index, 1)
    ElMessage.success('优化建议已删除')
    
  } catch (error) {
    // 用户取消删除
  }
}

const removeExpansion = async (result, index) => {
  try {
    await ElMessageBox.confirm('确定要删除这个扩展建议吗？', '确认删除', {
      type: 'warning'
    })
    
    result.expansions.splice(index, 1)
    ElMessage.success('扩展建议已删除')
    
  } catch (error) {
    // 用户取消删除
  }
}

// 其他操作
const viewDetails = (requirement) => {
  // TODO: 查看需求详情
  ElMessage.info('查看详情功能开发中...')
}

const exportAnalysisReport = () => {
  // TODO: 导出分析报告
  ElMessage.info('导出分析报告功能开发中...')
}

const clearResults = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有分析结果吗？', '确认清空', {
      type: 'warning'
    })
    
    analysisResults.value = []
    activeResultPanels.value = []
    ElMessage.success('分析结果已清空')
    
  } catch (error) {
    // 用户取消
  }
}

// 历史和统计
const showHistory = async () => {
  try {
    const response = await getLevel3AnalysisHistory(currentCycle.value)
    
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

const showStats = async () => {
  try {
    const response = await getLevel3AnalysisStats(currentCycle.value)
    
    if (response.code === 0) {
      stats.value = response.data.stats || {}
      statsDialogVisible.value = true
    } else {
      throw new Error(response.msg || '获取统计数据失败')
    }

  } catch (error) {
    console.error('获取统计数据失败:', error)
    ElMessage.error('获取统计数据失败: ' + (error.message || '未知错误'))
  }
}

// 进度更新
const resetProgress = () => {
  analysisProgress.value = 0
  progressStatus.value = 'active'
  currentProcessingRequirement.value = null
  progressSteps.value.forEach(step => {
    step.active = false
    step.completed = false
  })
}

const updateProgress = (percentage, text, activeStep) => {
  analysisProgress.value = percentage
  progressText.value = text
  
  progressSteps.value.forEach(step => {
    if (step.key === activeStep) {
      step.active = true
    } else if (step.completed || analysisProgress.value > 50) {
      step.completed = true
      step.active = false
    }
  })
}

// 辅助方法
const getComplexityType = (complexity) => {
  const complexityMap = {
    'simple': 'success',
    'moderate': 'warning',
    'complex': 'danger'
  }
  return complexityMap[complexity] || 'info'
}

const getFunctionTypeColor = (type) => {
  const typeMap = {
    'EI': 'primary',
    'EO': 'success', 
    'EQ': 'warning',
    'ILF': 'info',
    'EIF': 'danger'
  }
  return typeMap[type] || 'info'
}

const getAnalysisStatusType = (status) => {
  const statusMap = {
    'analyzed': 'success',
    'analyzing': 'warning',
    'failed': 'danger',
    'applied': 'primary'
  }
  return statusMap[status] || 'info'
}

const getAnalysisStatusText = (status) => {
  const statusMap = {
    'analyzed': '已分析',
    'analyzing': '分析中',
    'failed': '分析失败',
    'applied': '已应用'
  }
  return statusMap[status] || status
}

const getSuggestionPriorityType = (priority) => {
  const priorityMap = {
    'high': 'danger',
    'medium': 'warning',
    'low': 'info'
  }
  return priorityMap[priority] || 'info'
}

const getResultTitle = (result) => {
  const optimizationCount = result.optimizations ? result.optimizations.length : 0
  const expansionCount = result.expansions ? result.expansions.length : 0
  return `${result.requirement.title} (${optimizationCount} 个优化, ${expansionCount} 个扩展)`
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
      cycles.value = response.data || []
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
      versions.value = response.data || []
    }
  } catch (error) {
    console.error('加载版本列表失败:', error)
    ElMessage.error('加载版本列表失败')
  }
}

const loadL3Requirements = async () => {
  if (!currentVersion.value) return
  
  try {
    const params = {
      version_id: currentVersion.value,
      level: 3, // 只加载L3需求
      page: 1,
      pageSize: 1000
    }
    
    const response = await getNesmaRequirementList(params)
    
    if (response.code === 0) {
      l3Requirements.value = response.data.list || []
    }
  } catch (error) {
    console.error('加载L3需求失败:', error)
    ElMessage.error('加载L3需求失败')
  }
}

// 历史操作
const loadHistoryItem = (item) => {
  // TODO: 加载历史项目
  ElMessage.info('加载历史记录功能开发中...')
}

const viewHistoryDetail = (item) => {
  // TODO: 查看历史详情
  ElMessage.info('查看历史详情功能开发中...')
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
.level3-analyzer {
  padding: 20px;
}

.analyzer-config {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  margin: 20px 0;
}

.config-section h3 {
  margin: 0 0 15px 0;
  color: #333;
}

.config-row {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.l3-requirement-selector {
  margin: 20px 0;
}

.l3-requirement-selector h3 {
  margin-bottom: 15px;
  color: #333;
}

.filter-controls {
  margin-bottom: 15px;
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.analysis-results {
  margin: 20px 0;
}

.analysis-results h3 {
  margin-bottom: 15px;
  color: #333;
}

.result-summary {
  margin-bottom: 20px;
}

.summary-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.summary-number {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.summary-label {
  font-size: 12px;
  color: #666;
}

.result-actions {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.analysis-result-content {
  padding: 20px;
}

.result-basic-info {
  margin-bottom: 20px;
}

.optimization-suggestions h4,
.expansion-suggestions h4,
.nesma-scoring h4 {
  margin: 20px 0 15px 0;
  color: #333;
}

.suggestion-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.suggestion-item {
  border: 1px solid #e1e5e9;
  border-radius: 8px;
  padding: 16px;
  transition: all 0.3s;
  background: white;
}

.suggestion-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.suggestion-item.selected {
  border-color: #409eff;
  background: #f0f8ff;
}

.suggestion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.suggestion-type {
  font-weight: 500;
  margin-left: 8px;
}

.suggestion-content {
  margin-bottom: 16px;
}

.suggestion-description {
  margin: 8px 0;
  color: #666;
  line-height: 1.5;
}

.suggestion-comparison {
  margin: 12px 0;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 6px;
}

.comparison-section {
  margin-bottom: 12px;
}

.comparison-section h5 {
  margin: 0 0 6px 0;
  font-size: 14px;
  color: #666;
}

.before-content {
  padding: 8px;
  background: #fff3cd;
  border-radius: 4px;
  margin-bottom: 8px;
}

.after-content {
  padding: 8px;
  background: #d1ecf1;
  border-radius: 4px;
}

.suggestion-impact {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #666;
  background: #e8f5e8;
  padding: 8px;
  border-radius: 4px;
  margin-top: 8px;
}

.suggestion-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.expansion-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.expansion-card {
  border: 1px solid #e1e5e9;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
  background: white;
}

.expansion-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.expansion-card.selected {
  border-color: #409eff;
  background: #f0f8ff;
}

.expansion-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}

.expansion-title {
  font-weight: 500;
}

.expansion-content {
  margin-bottom: 16px;
}

.expansion-description {
  margin: 8px 0;
  color: #666;
  line-height: 1.5;
}

.expansion-meta {
  margin: 12px 0;
}

.meta-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  margin: 4px 0;
}

.meta-item .label {
  color: #999;
}

.meta-item .value {
  font-weight: 500;
  color: #333;
}

.expansion-reason {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #666;
  background: #f8f9fa;
  padding: 8px;
  border-radius: 4px;
  margin-top: 8px;
}

.expansion-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.nesma-scoring {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
  margin: 20px 0;
}

.scoring-details {
  margin-top: 12px;
}

.score-section h5 {
  margin: 0 0 12px 0;
  color: #333;
}

.batch-actions {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e1e5e9;
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.progress-container {
  padding: 20px;
}

.progress-info {
  margin-top: 20px;
  text-align: center;
}

.current-requirement {
  margin: 12px 0;
  padding: 8px;
  background: #f0f8ff;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
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

.stats-container {
  padding: 20px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}
</style>