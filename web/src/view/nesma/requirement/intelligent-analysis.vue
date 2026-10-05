<template>
  <div class="intelligent-analysis-container">
    <!-- 头部操作栏 -->
    <div class="analysis-header">
      <el-card shadow="never">
        <div class="header-content">
          <div class="title-section">
            <h2 class="analysis-title">
              <el-icon><MagicStick /></el-icon>
              智能需求分析增强
            </h2>
            <p class="analysis-subtitle">基于NESMA知识图谱的深度需求分析与优化</p>
          </div>
          <div class="action-section">
            <el-button-group>
              <el-button 
                type="primary" 
                @click="startIntelligentAnalysis"
                :loading="analysisLoading"
                :disabled="!selectedRequirements.length || !isConfigValid"
              >
                <el-icon><Lightning /></el-icon>
                开始智能分析
              </el-button>
              <el-button @click="showBatchConfig">
                <el-icon><Setting /></el-icon>
                分析配置
              </el-button>
              <el-button @click="exportAnalysisReport">
                <el-icon><Download /></el-icon>
                导出报告
              </el-button>
            </el-button-group>
          </div>
        </div>
      </el-card>
    </div>

    <!-- 项目选择和配置区域 -->
    <div class="project-selector-section">
      <el-card shadow="hover">
        <template #header>
          <div class="card-header">
            <span>项目配置</span>
            <el-badge :value="getSelectionSummary()" type="info">
              <el-button text>当前选择</el-button>
            </el-badge>
          </div>
        </template>
        
        <el-row :gutter="20">
          <!-- 项目选择 -->
          <el-col :span="6">
            <el-form-item label="选择项目">
              <el-select 
                v-model="selectedProject" 
                placeholder="请选择项目"
                @change="onProjectChange"
                style="width: 100%"
              >
                <el-option 
                  v-for="project in projects" 
                  :key="project.id" 
                  :label="project.name" 
                  :value="project.id"
                >
                  <div class="project-option">
                    <span class="project-name">{{ project.name }}</span>
                    <span class="project-domain">{{ project.domain }}</span>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          
          <!-- 周期选择 -->
          <el-col :span="6">
            <el-form-item label="选择周期">
              <el-select 
                v-model="selectedCycle" 
                placeholder="请选择周期"
                @change="onCycleChange"
                :disabled="!selectedProject"
                style="width: 100%"
              >
                <el-option 
                  v-for="cycle in cycles" 
                  :key="cycle.id" 
                  :label="cycle.name" 
                  :value="cycle.id"
                >
                  <div class="cycle-option">
                    <span class="cycle-name">{{ cycle.name }}</span>
                    <el-tag size="small" :type="getCycleStatusTag(cycle.status)">{{ cycle.status }}</el-tag>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          
          <!-- 版本选择 -->
          <el-col :span="6">
            <el-form-item label="当前版本">
              <el-select 
                v-model="selectedVersion" 
                placeholder="选择版本"
                @change="onVersionChange"
                :disabled="!selectedCycle"
                style="width: 100%"
              >
                <el-option 
                  v-for="version in versions" 
                  :key="version.id" 
                  :label="version.version" 
                  :value="version.id"
                >
                  <div class="version-option">
                    <span class="version-name">{{ version.version }}</span>
                    <el-tag size="small" :type="getVersionTypeTag(version.version_type)">{{ version.version_type }}</el-tag>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          
          <!-- 智能过滤 -->
          <el-col :span="6">
            <el-form-item label="智能过滤">
              <div class="filter-controls">
                <el-switch 
                  v-model="smartFilter.skipCompleted"
                  active-text="跳过已完成"
                  @change="applySmartFilter"
                />
                <el-switch 
                  v-model="smartFilter.onlyUnanalyzed"
                  active-text="仅未分析"
                  @change="applySmartFilter"
                />
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        
        <!-- 状态提示 -->
        <div class="status-alerts" v-if="statusMessages.length">
          <el-alert 
            v-for="(msg, index) in statusMessages" 
            :key="index"
            :type="msg.type"
            :title="msg.title"
            :description="msg.description"
            show-icon
            :closable="false"
            style="margin-bottom: 10px"
          />
        </div>
      </el-card>
    </div>

    <!-- 主要内容区域 -->
    <el-row :gutter="20">
      <!-- 左侧：需求选择与预览 -->
      <el-col :span="12">
        <el-card title="需求选择" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>需求选择</span>
              <el-badge :value="selectedRequirements.length" type="primary">
                <el-button text @click="selectAllRequirements">
                  {{ selectedRequirements.length === requirements.length ? '取消全选' : '全选' }}
                </el-button>
              </el-badge>
            </div>
          </template>
          
          <div class="requirement-selector">
            <el-tree
  ref="requirementTreeRef"
  :data="requirementTree"
  :props="treeProps"
  show-checkbox
  node-key="id"
  :default-checked-keys="selectedRequirements"
  @check="handleRequirementCheck"
  class="requirement-tree"
>
  <template #default="{ data }">
    <div class="tree-node">
      <div class="node-content">
        <span class="node-label">
          <el-tag :type="getRequirementTypeTag(data.level)" size="small">
            L{{ data.level }}
          </el-tag>
          {{ data.title }}
        </span>
        <div class="node-meta">
          <span class="complexity" v-if="data.complexity">
            复杂度: {{ data.complexity }}
          </span>
          <span class="function-type" v-if="data.function_type">
            类型: {{ data.function_type }}
          </span>
        </div>
      </div>
    </div>
  </template>
</el-tree>
<el-empty v-if="!requirementTree.length" description="暂无需求数据" />
          </div>
        </el-card>
      </el-col>

      <!-- 右侧：分析结果展示 -->
      <el-col :span="12">
        <el-card title="分析结果" shadow="hover">
          <template #header>
            <div class="card-header">
              <span>分析结果</span>
              <el-button-group size="small">
                <el-button 
                  :type="activeResultTab === 'overview' ? 'primary' : ''"
                  @click="activeResultTab = 'overview'"
                >
                  概览
                </el-button>
                <el-button 
                  :type="activeResultTab === 'details' ? 'primary' : ''"
                  @click="activeResultTab = 'details'"
                >
                  详情
                </el-button>
                <el-button 
                  :type="activeResultTab === 'recommendations' ? 'primary' : ''"
                  @click="activeResultTab = 'recommendations'"
                >
                  建议
                </el-button>
              </el-button-group>
            </div>
          </template>

          <div class="result-content">
            <!-- 概览视图 -->
            <div v-if="activeResultTab === 'overview'" class="overview-panel">
              <div v-if="!analysisResult" class="empty-state">
                <el-empty description="请先选择需求并开始分析" />
              </div>
              <div v-else class="analysis-overview">
                <div class="stats-grid">
                  <div class="stat-item">
                    <div class="stat-value">{{ analysisResult.total_requirements }}</div>
                    <div class="stat-label">总需求数</div>
                  </div>
                  <div class="stat-item">
                    <div class="stat-value">{{ analysisResult.optimized_count }}</div>
                    <div class="stat-label">优化数量</div>
                  </div>
                  <div class="stat-item">
                    <div class="stat-value">{{ analysisResult.confidence_score }}%</div>
                    <div class="stat-label">置信度</div>
                  </div>
                  <div class="stat-item">
                    <div class="stat-value">{{ analysisResult.function_points }}</div>
                    <div class="stat-label">功能点</div>
                  </div>
                </div>
                
                <!-- 新增：详细统计信息 -->
                <div class="detailed-stats">
                  <el-row :gutter="20">
                    <el-col :span="8">
                      <el-card class="mini-stat-card">
                        <div class="mini-stat-content">
                          <div class="mini-stat-icon generated">
                            <el-icon><DocumentAdd /></el-icon>
                          </div>
                          <div class="mini-stat-info">
                            <div class="mini-stat-value">{{ analysisResult.generated_count || 0 }}</div>
                            <div class="mini-stat-label">生成的内容</div>
                          </div>
                        </div>
                      </el-card>
                    </el-col>
                    <el-col :span="8">
                      <el-card class="mini-stat-card">
                        <div class="mini-stat-content">
                          <div class="mini-stat-icon improved">
                            <el-icon><TrendCharts /></el-icon>
                          </div>
                          <div class="mini-stat-info">
                            <div class="mini-stat-value">{{ analysisResult.improvement_rate || 0 }}%</div>
                            <div class="mini-stat-label">改进率</div>
                          </div>
                        </div>
                      </el-card>
                    </el-col>
                    <el-col :span="8">
                      <el-card class="mini-stat-card">
                        <div class="mini-stat-content">
                          <div class="mini-stat-icon time">
                            <el-icon><Timer /></el-icon>
                          </div>
                          <div class="mini-stat-info">
                            <div class="mini-stat-value">{{ analysisResult.processing_time || 0 }}s</div>
                            <div class="mini-stat-label">处理时间</div>
                          </div>
                        </div>
                      </el-card>
                    </el-col>
                  </el-row>
                </div>
                
                <div class="charts-section">
                  <div class="chart-container">
                    <div ref="functionTypeChart" class="chart"></div>
                  </div>
                  <div class="chart-container">
                    <div ref="complexityChart" class="chart"></div>
                  </div>
                </div>
                
                <!-- 新增：生成内容概览 -->
                <div class="generated-content-overview" v-if="analysisResult.generated_content">
                  <h4>生成内容概览</h4>
                  <el-row :gutter="10">
                    <el-col 
                      :span="6" 
                      v-for="(content, type) in analysisResult.generated_content" 
                      :key="type"
                    >
                      <div class="content-summary-card">
                        <div class="content-type">{{ getContentTypeName(type) }}</div>
                        <div class="content-count">{{ content.count || 0 }} 项</div>
                      </div>
                    </el-col>
                  </el-row>
                </div>
              </div>
            </div>

            <!-- 详情视图 -->
            <div v-if="activeResultTab === 'details'" class="details-panel">
              <el-scrollbar height="600px">
                <div v-for="(detail, index) in analysisDetails" :key="index" class="detail-item">
                  <el-card shadow="hover" class="detail-card">
                    <template #header>
                      <div class="detail-header">
                        <span class="detail-title">{{ detail.title }}</span>
                        <el-tag :type="getOptimizationTag(detail.optimization_level)">
                          {{ detail.optimization_level }}
                        </el-tag>
                      </div>
                    </template>
                    
                    <div class="detail-content">
                      <div class="original-desc">
                        <strong>原始描述：</strong>
                        <p>{{ detail.original_description }}</p>
                      </div>
                      
                      <div class="optimized-desc" v-if="detail.optimized_description">
                        <strong>优化描述：</strong>
                        <p>{{ detail.optimized_description }}</p>
                      </div>
                      
                      <div class="analysis-meta">
                        <el-row :gutter="10">
                          <el-col :span="8">
                            <div class="meta-item">
                              <span class="meta-label">功能类型：</span>
                              <el-tag size="small">{{ detail.function_type }}</el-tag>
                            </div>
                          </el-col>
                          <el-col :span="8">
                            <div class="meta-item">
                              <span class="meta-label">复杂度：</span>
                              <el-tag size="small" :type="getComplexityTag(detail.complexity)">
                                {{ detail.complexity }}
                              </el-tag>
                            </div>
                          </el-col>
                          <el-col :span="8">
                            <div class="meta-item">
                              <span class="meta-label">置信度：</span>
                              <span class="confidence-score">{{ detail.confidence }}%</span>
                            </div>
                          </el-col>
                        </el-row>
                      </div>
                      
                      <div class="knowledge-refs" v-if="detail.knowledge_references">
                        <strong>知识参考：</strong>
                        <div class="ref-tags">
                          <el-tag 
                            v-for="ref in detail.knowledge_references" 
                            :key="ref.id"
                            size="small"
                            type="info"
                            class="ref-tag"
                          >
                            {{ ref.title }}
                          </el-tag>
                        </div>
                      </div>
                    </div>
                  </el-card>
                </div>
              </el-scrollbar>
            </div>

            <!-- 建议视图 -->
            <div v-if="activeResultTab === 'recommendations'" class="recommendations-panel">
              <div class="recommendations-list">
                <div v-for="(rec, index) in recommendations" :key="index" class="recommendation-item">
                  <el-alert 
                    :type="rec.type"
                    :title="rec.title"
                    :description="rec.description"
                    show-icon
                    :closable="false"
                  />
                  <div class="recommendation-actions">
                    <el-button size="small" @click="applyRecommendation(rec)">
                      采纳建议
                    </el-button>
                    <el-button size="small" type="info" @click="viewRecommendationDetail(rec)">
                      查看详情
                    </el-button>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 分析进度对话框 -->
    <el-dialog 
      v-model="showProgressDialog" 
      title="智能分析进度"
      width="500px"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
    >
      <div class="progress-content">
        <div class="progress-info">
          <el-progress 
            :percentage="analysisProgress"
            :status="progressStatus"
            :stroke-width="12"
          />
          <div class="progress-text">
            <p>{{ currentAnalysisStep }}</p>
            <p class="progress-detail">{{ progressDetail }}</p>
          </div>
        </div>
        
        <div class="progress-logs">
          <el-scrollbar height="200px">
            <div v-for="(log, index) in analysisLogs" :key="index" class="log-item">
              <span class="log-time">{{ formatTime(log.timestamp) }}</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </el-scrollbar>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="cancelAnalysis" v-if="analysisLoading">取消分析</el-button>
          <el-button type="primary" @click="showProgressDialog = false" v-else>关闭</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 批量配置对话框 -->
    <el-dialog v-model="showConfigDialog" title="分析配置" width="600px">
      <el-form :model="analysisConfig" label-width="120px">
        <el-form-item label="分析模式">
          <el-radio-group v-model="analysisConfig.mode">
            <el-radio label="standard">标准模式</el-radio>
            <el-radio label="enhanced">增强模式</el-radio>
            <el-radio label="comprehensive">全面模式</el-radio>
          </el-radio-group>
        </el-form-item>
        
        <el-form-item label="AI模型选择">
          <el-select v-model="analysisConfig.ai_model" placeholder="选择AI模型">
            <el-option label="DeepSeek (推荐)" value="deepseek" />
            <el-option label="OpenAI GPT-4" value="openai-gpt4" />
            <el-option label="Claude 3" value="claude-3" />
            <el-option label="智能选择" value="auto" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="知识库权重">
          <el-slider v-model="analysisConfig.knowledge_weight" :min="0" :max="100" />
        </el-form-item>
        
        <el-form-item label="生成内容">
          <el-checkbox-group v-model="analysisConfig.generate_content">
            <el-checkbox label="optimized_description">优化描述</el-checkbox>
            <el-checkbox label="mermaid_diagram">Mermaid流程图</el-checkbox>
            <el-checkbox label="level4_requirements">四级需求生成</el-checkbox>
            <el-checkbox label="test_cases">测试用例</el-checkbox>
            <el-checkbox label="acceptance_criteria">验收标准</el-checkbox>
            <el-checkbox label="api_specifications">API规格说明</el-checkbox>
            <el-checkbox label="ui_mockups">界面原型</el-checkbox>
          </el-checkbox-group>
        </el-form-item>
        
        <el-form-item label="分析深度">
          <el-radio-group v-model="analysisConfig.analysis_depth">
            <el-radio label="basic">基础分析</el-radio>
            <el-radio label="comprehensive">全面分析</el-radio>
            <el-radio label="deep">深度分析</el-radio>
          </el-radio-group>
          <div class="analysis-depth-desc">
            <span v-if="analysisConfig.analysis_depth === 'basic'">快速优化需求描述和功能分类</span>
            <span v-else-if="analysisConfig.analysis_depth === 'comprehensive'">包含多维度分析和详细建议</span>
            <span v-else-if="analysisConfig.analysis_depth === 'deep'">全方位深度分析，生成完整交付物</span>
          </div>
        </el-form-item>
        
        <el-form-item label="质量要求">
          <el-row :gutter="10">
            <el-col :span="12">
              <div class="quality-item">
                <span class="quality-label">置信度阈值：</span>
                <el-slider v-model="analysisConfig.confidence_threshold" :min="0.5" :max="1" :step="0.05" show-input />
              </div>
            </el-col>
            <el-col :span="12">
              <div class="quality-item">
                <span class="quality-label">覆盖率要求：</span>
                <el-slider v-model="analysisConfig.coverage_requirement" :min="0.7" :max="1" :step="0.05" show-input />
              </div>
            </el-col>
          </el-row>
        </el-form-item>
        
        <el-form-item label="执行模式">
          <el-radio-group v-model="analysisConfig.execution_mode">
            <el-radio label="sequential">顺序执行</el-radio>
            <el-radio label="parallel">并行执行</el-radio>
            <el-radio label="adaptive">智能自适应</el-radio>
          </el-radio-group>
          <div class="execution-mode-desc">
            <span v-if="analysisConfig.execution_mode === 'sequential'">按步骤依次执行，稳定可靠</span>
            <span v-else-if="analysisConfig.execution_mode === 'parallel'">并行处理，速度更快</span>
            <span v-else-if="analysisConfig.execution_mode === 'adaptive'">根据负载智能调度</span>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showConfigDialog = false">取消</el-button>
          <el-button type="primary" @click="saveAnalysisConfig">保存配置</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { MagicStick, Lightning, Setting, Download, DocumentAdd, TrendCharts, Timer } from '@element-plus/icons-vue'
import * as echarts from 'echarts'
import { 
  getRequirementTree, 
  startIntelligentAnalysis as apiStartAnalysis,
  getAnalysisProgress,
  getAnalysisResult,
  getAnalysisRecommendations
} from '@/api/nesma'

// 响应式数据
const requirements = ref([])
const requirementTree = ref([])
const selectedRequirements = ref([])
const analysisLoading = ref(false)
const showProgressDialog = ref(false)
const showConfigDialog = ref(false)

// 项目配置相关
const projects = ref([])
const cycles = ref([])
const versions = ref([])
const selectedProject = ref(null)
const selectedCycle = ref(null)
const selectedVersion = ref(null)

// 智能过滤
const smartFilter = reactive({
  skipCompleted: true,
  onlyUnanalyzed: false,
  minComplexity: null,
  functionTypes: []
})

// 状态消息
const statusMessages = ref([])

// 分析结果
const analysisResult = ref(null)
const analysisDetails = ref([])
const recommendations = ref([])
const activeResultTab = ref('overview')

// 分析进度
const analysisProgress = ref(0)
const progressStatus = ref('active')
const currentAnalysisStep = ref('准备分析...')
const progressDetail = ref('')
const analysisLogs = ref([])

// 分析配置
const analysisConfig = reactive({
  mode: 'standard',
  ai_model: 'deepseek',
  knowledge_weight: 70,
  generate_content: ['optimized_description', 'mermaid_diagram'],
  analysis_depth: 'comprehensive',
  confidence_threshold: 0.8,
  coverage_requirement: 0.85,
  execution_mode: 'adaptive'
})

// 树形配置
const treeProps = {
  children: 'children',
  label: 'title'
}

// 组件引用
const requirementTreeRef = ref(null)
const functionTypeChart = ref(null)
const complexityChart = ref(null)

// 生命周期
onMounted(async () => {
  await loadProjects()
  await initCharts()
})

// 方法定义
const loadProjects = async () => {
  try {
    // 模拟项目数据 - 实际应该从API获取
    projects.value = [
      { id: 1, name: '智慧园区管理系统', domain: '智慧城市', status: 'active' },
      { id: 2, name: '企业内部管理系统', domain: '企业管理', status: 'active' },
      { id: 3, name: '金融风控平台', domain: '金融科技', status: 'planning' }
    ]
    
    // 如果URL有项目参数，自动选择
    const urlProjectId = getUrlProjectId()
    if (urlProjectId && projects.value.find(p => p.id === urlProjectId)) {
      selectedProject.value = urlProjectId
      await onProjectChange()
    }
    
  } catch (error) {
    ElMessage.error('加载项目数据失败: ' + error.message)
  }
}

const loadRequirements = async () => {
  if (!selectedProject.value) {
    requirementTree.value = []
    requirements.value = []
    return
  }
  
  try {
    const response = await getRequirementTree({
      project_id: selectedProject.value,
      cycle_id: selectedCycle.value,
      include_analyzed: true
    })
    
    requirementTree.value = response.data.tree
    requirements.value = response.data.flat_list
    
    // 应用智能过滤
    applySmartFilter()
    
    ElMessage.success('需求数据加载完成')
  } catch (error) {
    ElMessage.error('加载需求数据失败: ' + error.message)
  }
}

const handleRequirementCheck = (data, checkedInfo) => {
  selectedRequirements.value = checkedInfo.checkedKeys
}

const selectAllRequirements = () => {
  if (selectedRequirements.value.length === requirements.value.length) {
    selectedRequirements.value = []
    requirementTreeRef.value.setCheckedKeys([])
  } else {
    const allKeys = requirements.value.map(req => req.id)
    selectedRequirements.value = allKeys
    requirementTreeRef.value.setCheckedKeys(allKeys)
  }
}

const startIntelligentAnalysis = async () => {
  if (selectedRequirements.value.length === 0) {
    ElMessage.warning('请先选择需要分析的需求')
    return
  }

  try {
    analysisLoading.value = true
    showProgressDialog.value = true
    analysisProgress.value = 0
    progressStatus.value = 'active'
    currentAnalysisStep.value = '启动智能分析...'
    analysisLogs.value = []

    // 启动分析
    const response = await apiStartAnalysis({
      project_id: getCurrentProjectId(),
      cycle_id: getCurrentCycleId(),
      requirement_ids: selectedRequirements.value,
      config: analysisConfig
    })

    const taskId = response.data.task_id
    
    // 开始轮询进度
    pollAnalysisProgress(taskId)
    
  } catch (error) {
    ElMessage.error('启动分析失败: ' + error.message)
    analysisLoading.value = false
    showProgressDialog.value = false
  }
}

const pollAnalysisProgress = async (taskId) => {
  const interval = setInterval(async () => {
    try {
      const response = await getAnalysisProgress(taskId)
      const progress = response.data
      
      analysisProgress.value = progress.percentage
      currentAnalysisStep.value = progress.current_step
      progressDetail.value = progress.detail
      
      // 更新日志
      if (progress.logs && progress.logs.length > analysisLogs.value.length) {
        analysisLogs.value = progress.logs
      }
      
      if (progress.status === 'completed') {
        clearInterval(interval)
        analysisLoading.value = false
        progressStatus.value = 'success'
        currentAnalysisStep.value = '分析完成！'
        
        // 加载分析结果
        await loadAnalysisResult(taskId)
        
        setTimeout(() => {
          showProgressDialog.value = false
        }, 2000)
        
      } else if (progress.status === 'failed') {
        clearInterval(interval)
        analysisLoading.value = false
        progressStatus.value = 'exception'
        currentAnalysisStep.value = '分析失败: ' + progress.error_message
        
        ElMessage.error('分析失败: ' + progress.error_message)
      }
      
    } catch (error) {
      clearInterval(interval)
      analysisLoading.value = false
      progressStatus.value = 'exception'
      ElMessage.error('获取分析进度失败')
    }
  }, 2000)
}

const loadAnalysisResult = async (taskId) => {
  try {
    const [resultResponse, recommendationsResponse] = await Promise.all([
      getAnalysisResult(taskId),
      getAnalysisRecommendations(taskId)
    ])
    
    analysisResult.value = resultResponse.data.summary
    analysisDetails.value = resultResponse.data.details
    recommendations.value = recommendationsResponse.data.recommendations
    
    // 更新图表
    await nextTick()
    updateCharts()
    
    ElMessage.success('分析结果加载完成')
  } catch (error) {
    ElMessage.error('加载分析结果失败: ' + error.message)
  }
}

const initCharts = async () => {
  await nextTick()
  
  if (functionTypeChart.value) {
    const chart = echarts.init(functionTypeChart.value)
    chart.setOption({
      title: { text: '功能类型分布', textStyle: { fontSize: 14 } },
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: '60%',
        data: []
      }]
    })
  }
  
  if (complexityChart.value) {
    const chart = echarts.init(complexityChart.value)
    chart.setOption({
      title: { text: '复杂度分布', textStyle: { fontSize: 14 } },
      tooltip: { trigger: 'item' },
      series: [{
        type: 'pie',
        radius: '60%',
        data: []
      }]
    })
  }
}

const updateCharts = () => {
  if (analysisResult.value && functionTypeChart.value) {
    const functionChart = echarts.getInstanceByDom(functionTypeChart.value)
    if (functionChart) {
      functionChart.setOption({
        series: [{
          data: analysisResult.value.function_type_distribution
        }]
      })
    }
  }
  
  if (analysisResult.value && complexityChart.value) {
    const complexityChart = echarts.getInstanceByDom(complexityChart.value)
    if (complexityChart) {
      complexityChart.setOption({
        series: [{
          data: analysisResult.value.complexity_distribution
        }]
      })
    }
  }
}

// 辅助函数
const getRequirementTypeTag = (level) => {
  const tagMap = {
    1: 'danger',
    2: 'warning', 
    3: 'success',
    4: 'info'
  }
  return tagMap[level] || 'info'
}

const getOptimizationTag = (level) => {
  const tagMap = {
    'high': 'danger',
    'medium': 'warning',
    'low': 'success',
    'none': 'info'
  }
  return tagMap[level] || 'info'
}

const getComplexityTag = (complexity) => {
  const tagMap = {
    'high': 'danger',
    'medium': 'warning',
    'low': 'success'
  }
  return tagMap[complexity?.toLowerCase()] || 'info'
}

const formatTime = (timestamp) => {
  return new Date(timestamp).toLocaleTimeString()
}

const getCurrentProjectId = () => {
  return selectedProject.value
}

const getCurrentCycleId = () => {
  return selectedCycle.value
}

// 新增方法：项目选择处理
const onProjectChange = async () => {
  selectedCycle.value = null
  selectedVersion.value = null
  cycles.value = []
  versions.value = []
  requirementTree.value = []
  requirements.value = []
  selectedRequirements.value = []
  
  if (selectedProject.value) {
    await loadCycles()
    updateStatusMessages()
  }
}

// 新增方法：周期选择处理
const onCycleChange = async () => {
  selectedVersion.value = null
  versions.value = []
  requirementTree.value = []
  requirements.value = []
  selectedRequirements.value = []
  
  if (selectedCycle.value) {
    await loadVersions()
    await loadRequirements()
    updateStatusMessages()
  }
}

// 新增方法：版本选择处理
const onVersionChange = async () => {
  if (selectedVersion.value) {
    await loadRequirements()
    updateStatusMessages()
  }
}

// 新增方法：加载周期数据
const loadCycles = async () => {
  try {
    // 模拟周期数据 - 实际应该从API获取
    cycles.value = [
      { id: 1, name: '核心功能开发', status: 'active', project_id: selectedProject.value },
      { id: 2, name: '扩展功能开发', status: 'planning', project_id: selectedProject.value },
      { id: 3, name: '系统集成测试', status: 'completed', project_id: selectedProject.value }
    ]
    
    // 自动选择活跃的周期
    const activeCycle = cycles.value.find(c => c.status === 'active')
    if (activeCycle) {
      selectedCycle.value = activeCycle.id
      await onCycleChange()
    }
  } catch (error) {
    ElMessage.error('加载周期数据失败: ' + error.message)
  }
}

// 新增方法：加载版本数据
const loadVersions = async () => {
  try {
    // 模拟版本数据 - 实际应该从API获取
    versions.value = [
      { id: 1, version: 'v1.0', version_type: 'initial', cycle_id: selectedCycle.value },
      { id: 2, version: 'v1.1', version_type: 'analyzed', cycle_id: selectedCycle.value },
      { id: 3, version: 'v1.2', version_type: 'optimized', cycle_id: selectedCycle.value }
    ]
    
    // 自动选择最新版本
    if (versions.value.length > 0) {
      selectedVersion.value = versions.value[versions.value.length - 1].id
    }
  } catch (error) {
    ElMessage.error('加载版本数据失败: ' + error.message)
  }
}

// 新增方法：应用智能过滤
const applySmartFilter = () => {
  if (!requirements.value.length) return
  
  let filteredIds = []
  
  // 根据智能过滤条件筛选需求
  requirements.value.forEach(req => {
    let shouldInclude = true
    
    // 跳过已完成的需求
    if (smartFilter.skipCompleted && req.status === 'completed') {
      shouldInclude = false
    }
    
    // 仅显示未分析的需求
    if (smartFilter.onlyUnanalyzed && req.is_analyzed) {
      shouldInclude = false
    }
    
    // 最小复杂度过滤
    if (smartFilter.minComplexity && 
        getComplexityLevel(req.complexity) < getComplexityLevel(smartFilter.minComplexity)) {
      shouldInclude = false
    }
    
    // 功能类型过滤
    if (smartFilter.functionTypes.length > 0 && 
        !smartFilter.functionTypes.includes(req.function_type)) {
      shouldInclude = false
    }
    
    if (shouldInclude) {
      filteredIds.push(req.id)
    }
  })
  
  // 更新选中的需求（保留已选中且符合过滤条件的）
  selectedRequirements.value = selectedRequirements.value.filter(id => 
    filteredIds.includes(id)
  )
  
  // 更新树形组件的选中状态
  if (requirementTreeRef.value) {
    requirementTreeRef.value.setCheckedKeys(selectedRequirements.value)
  }
  
  updateStatusMessages()
}

// 新增方法：更新状态消息
const updateStatusMessages = () => {
  statusMessages.value = []
  
  if (!selectedProject.value) {
    statusMessages.value.push({
      type: 'warning',
      title: '请选择项目',
      description: '请先选择要分析的项目'
    })
    return
  }
  
  if (!selectedCycle.value) {
    statusMessages.value.push({
      type: 'warning', 
      title: '请选择周期',
      description: '请选择项目的构建周期'
    })
    return
  }
  
  if (!requirements.value.length) {
    statusMessages.value.push({
      type: 'info',
      title: '暂无需求数据',
      description: '当前周期下没有需求数据，请检查数据或选择其他周期'
    })
    return
  }
  
  // 显示过滤统计
  const completedCount = requirements.value.filter(r => r.status === 'completed').length
  const analyzedCount = requirements.value.filter(r => r.is_analyzed).length
  
  if (smartFilter.skipCompleted && completedCount > 0) {
    statusMessages.value.push({
      type: 'success',
      title: `已跳过 ${completedCount} 个已完成需求`,
      description: '用户已满意的需求将被自动跳过分析'
    })
  }
  
  if (smartFilter.onlyUnanalyzed && analyzedCount > 0) {
    statusMessages.value.push({
      type: 'info',
      title: `${analyzedCount} 个需求已分析`,
      description: '仅显示未分析的需求项'
    })
  }
}

// 新增方法：获取选择摘要
const getSelectionSummary = () => {
  if (!selectedProject.value) return '未选择项目'
  if (!selectedCycle.value) return '未选择周期'
  if (!selectedVersion.value) return '未选择版本'
  
  const projectName = projects.value.find(p => p.id === selectedProject.value)?.name
  const cycleName = cycles.value.find(c => c.id === selectedCycle.value)?.name
  const versionName = versions.value.find(v => v.id === selectedVersion.value)?.version
  
  return `${projectName} > ${cycleName} > ${versionName}`
}

// 新增方法：验证配置有效性
const isConfigValid = computed(() => {
  return selectedProject.value && selectedCycle.value && selectedVersion.value
})

// 新增方法：辅助函数
const getCycleStatusTag = (status) => {
  const tagMap = {
    'active': 'success',
    'planning': 'warning',
    'completed': 'info',
    'archived': 'info'
  }
  return tagMap[status] || 'info'
}

const getVersionTypeTag = (type) => {
  const tagMap = {
    'initial': 'info',
    'analyzed': 'warning', 
    'optimized': 'success'
  }
  return tagMap[type] || 'info'
}

const getComplexityLevel = (complexity) => {
  const levelMap = {
    'Low': 1,
    'Average': 2,
    'High': 3
  }
  return levelMap[complexity] || 0
}

const getUrlProjectId = () => {
  // 从URL参数获取项目ID
  const urlParams = new URLSearchParams(window.location.search)
  const projectId = urlParams.get('project_id')
  return projectId ? parseInt(projectId) : null
}

// 新增方法：获取内容类型名称
const getContentTypeName = (type) => {
  const typeNames = {
    'optimized_description': '优化描述',
    'mermaid_diagram': '流程图',
    'level4_requirements': '四级需求',
    'test_cases': '测试用例',
    'acceptance_criteria': '验收标准',
    'api_specifications': 'API规格',
    'ui_mockups': '界面原型'
  }
  return typeNames[type] || type
}

const showBatchConfig = () => {
  showConfigDialog.value = true
}

const saveAnalysisConfig = () => {
  showConfigDialog.value = false
  ElMessage.success('配置保存成功')
}

const cancelAnalysis = () => {
  ElMessageBox.confirm('确定要取消当前分析吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(() => {
    analysisLoading.value = false
    showProgressDialog.value = false
    ElMessage.info('已取消分析')
  })
}

const applyRecommendation = (recommendation) => {
  ElMessage.success('建议已采纳: ' + recommendation.title)
}

const viewRecommendationDetail = (recommendation) => {
  ElMessage.info('查看建议详情: ' + recommendation.title)
}

const exportAnalysisReport = () => {
  if (!analysisResult.value) {
    ElMessage.warning('请先完成分析')
    return
  }
  
  ElMessage.success('开始导出分析报告...')
}
</script>

<style scoped>
.intelligent-analysis-container {
  padding: 20px;
}

.analysis-header {
  margin-bottom: 20px;
}

.project-selector-section {
  margin-bottom: 20px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.analysis-title {
  margin: 0;
  font-size: 20px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
}

.analysis-subtitle {
  margin: 5px 0 0 0;
  color: #606266;
  font-size: 14px;
}

.project-option, .cycle-option, .version-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.project-domain, .cycle-status, .version-type {
  font-size: 12px;
  color: #909399;
}

.filter-controls {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.status-alerts {
  margin-top: 15px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.requirement-tree {
  max-height: 500px;
  overflow-y: auto;
}

.tree-node {
  width: 100%;
}

.node-content {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.node-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
}

.node-meta {
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #909399;
}

.result-content {
  min-height: 600px;
}

.empty-state {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 400px;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 20px;
  margin-bottom: 20px;
}

.stat-item {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.stat-label {
  color: #606266;
  font-size: 14px;
  margin-top: 5px;
}

.charts-section {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}

.chart-container {
  height: 300px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.chart {
  width: 100%;
  height: 100%;
}

.detail-item {
  margin-bottom: 15px;
}

.detail-card {
  border: 1px solid #e4e7ed;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-title {
  font-weight: 500;
}

.detail-content {
  line-height: 1.6;
}

.original-desc, .optimized-desc {
  margin-bottom: 15px;
}

.original-desc p, .optimized-desc p {
  margin: 5px 0;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 4px;
}

.optimized-desc p {
  background: #e8f5e8;
  border-left: 3px solid #67c23a;
}

.analysis-meta {
  margin: 15px 0;
  padding: 10px;
  background: #fafafa;
  border-radius: 4px;
}

.meta-item {
  display: flex;
  align-items: center;
  gap: 5px;
}

.meta-label {
  font-weight: 500;
  color: #606266;
}

.confidence-score {
  font-weight: bold;
  color: #409eff;
}

.knowledge-refs {
  margin-top: 15px;
}

.ref-tags {
  margin-top: 8px;
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
}

.ref-tag {
  cursor: pointer;
}

.recommendations-list {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.recommendation-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 15px;
}

.recommendation-actions {
  margin-top: 10px;
  display: flex;
  gap: 10px;
}

.progress-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.progress-info {
  text-align: center;
}

.progress-text {
  margin-top: 15px;
}

.progress-detail {
  color: #909399;
  font-size: 14px;
}

.progress-logs {
  background: #f8f9fa;
  border-radius: 4px;
  padding: 10px;
}

.log-item {
  display: flex;
  gap: 10px;
  margin-bottom: 5px;
  font-size: 12px;
}

.log-time {
  color: #909399;
  min-width: 80px;
}

.log-message {
  color: #606266;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.analysis-depth-desc, .execution-mode-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 5px;
  font-style: italic;
}

.quality-item {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.quality-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
}

.detailed-stats {
  margin: 20px 0;
}

.mini-stat-card {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 0;
}

.mini-stat-content {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px;
}

.mini-stat-icon {
  width: 40px;
  height: 40px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: white;
}

.mini-stat-icon.generated {
  background: linear-gradient(135deg, #409EFF 0%, #1976D2 100%);
}

.mini-stat-icon.improved {
  background: linear-gradient(135deg, #67C23A 0%, #4CAF50 100%);
}

.mini-stat-icon.time {
  background: linear-gradient(135deg, #E6A23C 0%, #FF9800 100%);
}

.mini-stat-info {
  flex: 1;
}

.mini-stat-value {
  font-size: 18px;
  font-weight: bold;
  color: #303133;
  margin-bottom: 2px;
}

.mini-stat-label {
  font-size: 12px;
  color: #909399;
}

.generated-content-overview {
  margin-top: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.generated-content-overview h4 {
  margin: 0 0 12px 0;
  font-size: 16px;
  color: #303133;
}

.content-summary-card {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  padding: 12px;
  text-align: center;
}

.content-type {
  font-size: 12px;
  color: #606266;
  margin-bottom: 4px;
}

.content-count {
  font-size: 16px;
  font-weight: bold;
  color: #409EFF;
}
</style>