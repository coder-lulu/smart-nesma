<template>
  <div>
    <el-dialog
    v-model="visible"
    :title="dialogTitle"
    width="900px"
    :close-on-click-modal="false"
    :before-close="handleBeforeClose"
  >
    <!-- 分析进度阶段 -->
    <div v-if="analysisStatus.status === 'running'" class="enhanced-analysis-progress">
      <!-- 项目信息 -->
      <div class="project-header">
        <div class="project-title">
          <el-icon class="project-icon"><FolderOpened /></el-icon>
          <span>{{ project.name }}</span>
        </div>
        <el-tag type="info" size="small">{{ taskId }}</el-tag>
      </div>

      <!-- 动画进度区域 -->
      <div class="animated-progress-section">
        <!-- 主进度条 -->
        <div class="main-progress">
          <el-progress 
            :percentage="analysisStatus.progress" 
            :status="getProgressStatus()"
            :stroke-width="18"
            :show-text="false"
          />
          <div class="progress-info">
            <span class="progress-percent">{{ analysisStatus.progress }}%</span>
            <span class="stage-name">{{ analysisStatus.stageDesc || currentStageText }}</span>
          </div>
        </div>

        <!-- 动画图标区域 -->
        <div class="animation-section">
          <div :class="['stage-icon-container', analysisStatus.animationType || 'pulse']">
            <el-icon 
              :class="['stage-icon', getStageIconClass()]" 
              :style="{ color: getStageColor() }"
            >
              <component :is="getStageIcon()" />
            </el-icon>
          </div>
          
          <!-- 状态文本 -->
          <div class="status-text">
            <p class="primary-text">{{ analysisStatus.statusText || '正在处理...' }}</p>
            <p class="secondary-text" v-if="analysisStatus.animationData && analysisStatus.animationData.estimatedRemaining > 0">
              预计剩余时间: {{ analysisStatus.animationData.estimatedRemaining }} 秒
            </p>
          </div>
        </div>

        <!-- 阶段历史时间线 -->
        <div class="stage-timeline" v-if="analysisStatus.stageHistory && analysisStatus.stageHistory.length > 0">
          <h4>分析步骤</h4>
          <el-timeline>
            <el-timeline-item 
              v-for="(stage, index) in analysisStatus.stageHistory" 
              :key="index"
              :type="getTimelineType(stage.status)"
              :hollow="stage.status !== 'completed'"
              :timestamp="formatTime(stage.startTime)"
            >
              <div class="timeline-content">
                <span class="stage-desc">{{ stage.description }}</span>
                <el-tag 
                  v-if="stage.duration" 
                  size="small" 
                  type="info"
                >
                  {{ stage.duration }}ms
                </el-tag>
              </div>
            </el-timeline-item>
          </el-timeline>
        </div>
      </div>

      <!-- 处理统计 -->
      <div class="processing-stats" v-if="analysisStatus.metadata">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-statistic title="总需求数" :value="analysisStatus.metadata.totalRequirements || 0" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="已处理" :value="analysisStatus.metadata.processedRequirements || 0" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="成功优化" :value="analysisStatus.metadata.optimizedRequirements || 0" />
          </el-col>
          <el-col :span="6">
            <el-statistic title="错误次数" :value="analysisStatus.metadata.errors || 0" />
          </el-col>
        </el-row>
      </div>

      <!-- 分析特性说明 -->
      <div class="analysis-features">
        <div class="feature-grid">
          <div class="feature-item">
            <el-icon><MagicStick /></el-icon>
            <span>智能需求分析</span>
          </div>
          <div class="feature-item">
            <el-icon><Collection /></el-icon>
            <span>NESMA功能评估</span>
          </div>
          <div class="feature-item">
            <el-icon><Timer /></el-icon>
            <span>复杂度计算</span>
          </div>
          <div class="feature-item">
            <el-icon><FolderOpened /></el-icon>
            <span>知识库支持</span>
          </div>
        </div>
      </div>

      <!-- 实时日志 -->
      <div class="progress-logs" v-if="progressLogs.length > 0">
        <h4>处理日志</h4>
        <div class="log-container">
          <div 
            v-for="log in progressLogs.slice(-5)" 
            :key="log.timestamp"
            class="log-item"
          >
            <span class="log-time">{{ formatTime(log.timestamp) }}</span>
            <span class="log-message">{{ log.message }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 分析完成阶段 -->
    <div v-else-if="analysisStatus.status === 'completed'" class="analysis-result">
      <div class="result-header">
        <el-result
          icon="success"
          title="分析完成"
          :sub-title="`项目 ${project.name} 的 一键分析完成`"
        />
        <div class="completion-stats">
          <el-row :gutter="16">
            <el-col :span="6">
              <el-statistic 
                title="总用时" 
                :value="parseProcessingTime(apiData?.statistics?.processing_time)" 
                suffix="秒"
              />
            </el-col>
            <el-col :span="6">
              <el-statistic 
                title="处理需求" 
                :value="apiData?.statistics?.processed_count || 0" 
              />
            </el-col>
            <el-col :span="6">
              <el-statistic 
                title="成功率" 
                :value="apiData?.statistics?.success_rate || 0" 
                suffix="%"
              />
            </el-col>
            <el-col :span="6">
              <el-statistic 
                title="平均置信度" 
                :value="Math.round((apiData?.statistics?.average_confidence || 0) * 100)" 
                suffix="%"
              />
            </el-col>
          </el-row>
        </div>
      </div>

      <!-- 详细分析结果 -->
      <div v-if="detailedResult" class="detailed-result">
        <el-tabs v-model="activeResultTab" type="card">
          <!-- 分析摘要 -->
          <el-tab-pane label="分析摘要" name="summary">
            <div class="summary-content">
              <h4>项目分析概览</h4>
              <div class="summary-grid">
                <div class="summary-item">
                  <h5>处理统计</h5>
                  <p>总需求数: {{ apiData?.statistics?.total_requirements || 0 }}</p>
                  <p>成功处理: {{ apiData?.statistics?.processed_count || 0 }}</p>
                  <p>成功率: {{ apiData?.statistics?.success_rate || 0 }}%</p>
                  <p>失败数量: {{ apiData?.statistics?.failed_count || 0 }}</p>
                </div>
                <div class="summary-item">
                  <h5>质量评估总览</h5>
                  <p>平均置信度: {{ Math.round((apiData?.statistics?.average_confidence || 0) * 100) }}%</p>
                  <p>处理时间: {{ parseProcessingTime(apiData?.statistics?.processing_time) }}秒</p>
                  <p>任务状态: {{ apiData?.task_info?.status || '未知' }}</p>
                  <p>任务摘要: {{ apiData?.task_info?.summary || '无' }}</p>
                </div>
                <div class="summary-item">
                  <h5>性能指标</h5>
                  <p>总任务数: {{ apiData?.task_info?.totalCount || 0 }}</p>
                  <p>已处理: {{ apiData?.task_info?.processedCount || 0 }}</p>
                  <p>成功数: {{ apiData?.task_info?.successCount || 0 }}</p>
                  <p>失败数: {{ apiData?.task_info?.failedCount || 0 }}</p>
                </div>
                <div class="summary-item">
                  <h5>分析结果</h5>
                  <p>源版本: {{ apiData?.source_version?.version || '未知' }}</p>
                  <p>目标版本: {{ apiData?.target_version?.version || '未知' }}</p>
                  <p>分析需求数: {{ safeGetArrayLength(apiData?.analyzed_requirements) }}</p>
                  <p>任务完成时间: {{ formatTime(apiData?.task_info?.endTime) }}</p>
                </div>
              </div>
              
              <!-- 详细质量评分展示 -->
              <div class="quality-breakdown" v-if="apiData?.result?.quality_score?.score_breakdown">
                <h4>质量评分细分</h4>
                <el-row :gutter="16">
                  <el-col :span="6" v-for="(score, category) in apiData.result.quality_score.score_breakdown" :key="category">
                    <el-card class="score-card">
                      <div class="score-item">
                        <div class="score-title">{{ getScoreCategoryName(category) }}</div>
                        <div class="score-value" :class="getScoreClass(score)">{{ score }}/100</div>
                      </div>
                    </el-card>
                  </el-col>
                </el-row>
              </div>
            </div>
          </el-tab-pane>

          <!-- 优化建议 -->
          <el-tab-pane label="优化建议" name="optimizations">
            <div class="optimizations-content">
              <!-- 需求分析结果 -->
              <div v-if="apiData?.analyzed_requirements && apiData?.analyzed_requirements.length > 0" class="insights-section">
                <h4>需求分析结果</h4>
                <div 
                  v-for="(requirement, index) in apiData?.analyzed_requirements" 
                  :key="index"
                  class="insight-item"
                >
                  <el-card class="insight-card">
                    <div class="insight-header">
                      <div class="insight-title-area">
                        <h5>{{ requirement.original_requirement?.title || '未知需求' }}</h5>
                        <div class="insight-meta">
                          <el-tag type="info" size="small">
                            级别 {{ requirement.original_requirement?.level || '未知' }}
                          </el-tag>
                          <el-tag type="success" size="small">
                            改进得分 {{ Math.round((requirement.improvement_score || 0) * 100) }}%
                          </el-tag>
                          <el-tag type="warning" size="small">
                            {{ requirement.optimized_requirement?.function_type || '未知' }}
                          </el-tag>
                        </div>
                      </div>
                      <div class="confidence-score">
                        <span class="confidence-label">置信度</span>
                        <span class="confidence-value">{{ Math.round((requirement.optimized_requirement?.aiConfidenceScore || 0) * 100) }}%</span>
                      </div>
                    </div>
                    
                    <div class="insight-content">
                      <div class="requirement-comparison">
                        <div class="original-requirement">
                          <h6>原始需求:</h6>
                          <p>{{ requirement.original_requirement?.description || '无描述' }}</p>
                        </div>
                        <div class="optimized-requirement">
                          <h6>优化后需求:</h6>
                          <p>{{ requirement.optimized_requirement?.description || '无描述' }}</p>
                        </div>
                      </div>
                      
                      <div v-if="requirement.changes && requirement.changes.length > 0" class="changes-section">
                        <h6>主要变更:</h6>
                        <ul>
                          <li v-for="change in requirement.changes" :key="change">{{ change }}</li>
                        </ul>
                      </div>
                    </div>
                  </el-card>
                </div>
              </div>
              
              <!-- 统计信息 -->
              <div v-if="apiData?.statistics" class="recommendations-section">
                <h4>分析统计</h4>
                <el-card class="recommendation-card">
                  <div class="rec-header">
                    <div class="rec-title-area">
                      <h5>分析完成情况</h5>
                      <div class="rec-meta">
                        <el-tag type="success" size="small">
                          成功率 {{ apiData?.statistics?.success_rate || 0 }}%
                        </el-tag>
                        <el-tag type="info" size="small">
                          平均置信度 {{ Math.round((apiData?.statistics?.average_confidence || 0) * 100) }}%
                        </el-tag>
                        <el-tag type="warning" size="small">
                          处理时间 {{ parseProcessingTime(apiData?.statistics?.processing_time) }}秒
                        </el-tag>
                      </div>
                    </div>
                  </div>
                  
                  <div class="rec-content">
                    <div class="statistics-grid">
                      <div class="stat-item">
                        <span class="stat-label">总需求数:</span>
                        <span class="stat-value">{{ apiData?.statistics?.total_requirements || 0 }}</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-label">已处理:</span>
                        <span class="stat-value">{{ apiData?.statistics?.processed_count || 0 }}</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-label">成功数:</span>
                        <span class="stat-value">{{ apiData?.statistics?.success_count || 0 }}</span>
                      </div>
                      <div class="stat-item">
                        <span class="stat-label">失败数:</span>
                        <span class="stat-value">{{ apiData?.statistics?.failed_count || 0 }}</span>
                      </div>
                    </div>
                    
                    <div v-if="apiData?.task_info?.summary" class="task-summary">
                      <h6>任务摘要:</h6>
                      <p>{{ apiData?.task_info?.summary }}</p>
                    </div>
                  </div>
                </el-card>
              </div>
              
              <el-empty v-if="(!apiData?.analyzed_requirements || apiData?.analyzed_requirements.length === 0) && 
                              (!apiData?.statistics)" 
                        description="暂无分析数据" />
            </div>
          </el-tab-pane>

          <!-- NESMA分析 -->
          <el-tab-pane label="NESMA评估" name="nesma">
            <div class="nesma-content">
              <div v-if="detailedResult.nesmaAnalysis">
                <h4>功能点分类统计</h4>
                <el-row :gutter="16" class="nesma-stats">
                  <el-col :span="4" v-for="(count, type) in detailedResult.nesmaAnalysis.functionTypes" :key="type">
                    <el-statistic :title="type" :value="count" />
                  </el-col>
                </el-row>
                
                <h4>复杂度分布</h4>
                <div class="complexity-chart">
                  <el-tag v-for="(count, level) in detailedResult.nesmaAnalysis.complexityDistribution" :key="level">
                    {{ level }}: {{ count }}
                  </el-tag>
                </div>
              </div>
              <el-empty v-else description="暂无NESMA分析数据" />
            </div>
          </el-tab-pane>

          <!-- 知识库引用 -->
          <el-tab-pane label="技术详情" name="knowledge">
            <div class="knowledge-content">
              <h4>分析技术详情</h4>
              
              <!-- AI模型信息 -->
              <div class="tech-section">
                <h5>AI分析引擎</h5>
                <el-descriptions :column="2" border>
                  <el-descriptions-item label="分析版本">
                    {{ apiData?.task_info?.config?.version || 'v1.0' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="分析引擎">
                    {{ apiData?.task_info?.config?.analysis_engine || 'NESMA AI Analyzer' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="AI模型">
                    {{ apiData?.task_info?.config?.ai_model || 'N/A' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="批次大小">
                    {{ apiData?.task_info?.config?.batch_size || 'N/A' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="最大重试次数">
                    {{ apiData?.task_info?.config?.max_retries || 'N/A' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="最大并发数">
                    {{ apiData?.task_info?.config?.max_concurrency || 'N/A' }}
                  </el-descriptions-item>
                </el-descriptions>
              </div>
              
              <!-- 任务信息 -->
              <div class="tech-section">
                <h5>任务执行信息</h5>
                <el-descriptions :column="2" border>
                  <el-descriptions-item label="任务ID">
                    {{ apiData?.task_info?.ID || 'N/A' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="任务类型">
                    {{ apiData?.task_info?.taskType || 'N/A' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="开始时间">
                    {{ formatTime(apiData?.task_info?.startTime) }}
                  </el-descriptions-item>
                  <el-descriptions-item label="结束时间">
                    {{ formatTime(apiData?.task_info?.endTime) }}
                  </el-descriptions-item>
                  <el-descriptions-item label="源版本">
                    {{ apiData?.source_version?.version || 'N/A' }}
                  </el-descriptions-item>
                  <el-descriptions-item label="目标版本">
                    {{ apiData?.target_version?.version || 'N/A' }}
                  </el-descriptions-item>
                </el-descriptions>
              </div>
              
              <!-- 处理统计 -->
              <div class="tech-section">
                <h5>处理性能统计</h5>
                <el-descriptions :column="3" border>
                  <el-descriptions-item label="总需求数">
                    {{ apiData?.statistics?.total_requirements || 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="处理成功">
                    {{ apiData?.statistics?.success_count || 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="处理失败">
                    {{ apiData?.statistics?.failed_count || 0 }}
                  </el-descriptions-item>
                  <el-descriptions-item label="成功率">
                    {{ apiData?.statistics?.success_rate || 0 }}%
                  </el-descriptions-item>
                  <el-descriptions-item label="平均置信度">
                    {{ Math.round((apiData?.statistics?.average_confidence || 0) * 100) }}%
                  </el-descriptions-item>
                  <el-descriptions-item label="处理时间">
                    {{ apiData?.statistics?.processing_time || '0s' }}
                  </el-descriptions-item>
                </el-descriptions>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <!-- 分析失败阶段 -->
    <div v-else-if="apiData?.status === 'failed'" class="analysis-error">
      <el-result
        icon="error"
        title="分析失败"
        :sub-title="apiData?.error || '分析过程中发生未知错误'"
      >
        <template #extra>
          <el-button type="primary" @click="retryAnalysis">重新分析</el-button>
          <el-button @click="handleClose">关闭</el-button>
        </template>
      </el-result>
    </div>

    <!-- 对话框操作按钮 -->
    <template #footer v-if="apiData?.status !== 'failed'">
      <div class="dialog-footer">
        <el-button 
          v-if="apiData?.status === 'running'" 
          type="danger" 
          @click="cancelAnalysis"
        >
          取消分析
        </el-button>
        <el-button @click="handleClose">
          {{ apiData?.status === 'completed' ? '关闭' : '后台运行' }}
        </el-button>
        <el-button 
          v-if="apiData?.status === 'completed'" 
          type="primary" 
          @click="openReportGenerator"
        >
          <el-icon><Document /></el-icon>
          生成报告
        </el-button>
        <!-- <el-button 
          v-if="apiData?.status === 'completed'" 
          type="success" 
          @click="viewFullResult"
        >
          查看完整结果
        </el-button> -->
      </div>
    </template>
  </el-dialog>

  <!-- 报告生成器组件 -->
  <ReportGenerator
    ref="reportGeneratorRef"
    :task-id="taskId"
    :project="project"
    :analysis-data="apiData"
    @download="handleReportDownload"
  />
</div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import { 
  getUnifiedAnalysisProgress, 
  getDetailedAnalysisResult,
  cancelUnifiedAnalysis,
  exportUnifiedAnalysisReport,
  applyOptimization as applyOptimizationApi
} from '@/api/nesma'
import ReportGenerator from './ReportGenerator.vue'
import {
  Document, Upload, Search, Check, Warning, Timer, FolderOpened, 
  MagicStick, Collection, Cpu, TrendCharts, Coin, DataAnalysis, 
  Trophy, Setting, User, Calendar, Star, EditPen, DocumentCopy, 
  ArrowRight, Grid, List, Loading
} from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: Boolean,
  taskId: [String, Number],
  project: Object
})

const emit = defineEmits(['update:modelValue', 'complete', 'close'])

// 响应式数据
const analysisStatus = ref({
  status: 'running', // running, completed, failed
  progress: 0,
  message: '',
  startTime: new Date(),
  endTime: null,
  duration: 0,
  metadata: null,
  error: null,
  // 增强字段
  currentStage: 'initializing',
  stageDesc: '正在初始化...',
  animationType: 'pulse',
  statusText: '分析进行中',
  errorMsg: '',
  stageHistory: [],
  animationData: {}
})

const detailedResult = ref(null)
const analysisResult = ref(null)
const progressLogs = ref([])
const activeResultTab = ref('summary')
const reportGeneratorRef = ref(null)
let progressInterval = null
let initialDelayTimer = null
let retryTimers = new Set() // 用于跟踪所有重试定时器

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const dialogTitle = computed(() => {
  switch (analysisStatus.value.status) {
    case 'running': return '项目智能分析 - 进行中'
    case 'completed': return '项目智能分析 - 已完成'
    case 'failed': return '项目智能分析 - 失败'
    default: return '项目智能分析'
  }
})

const progressStatus = computed(() => {
  if (analysisStatus.value.progress === 100) return 'success'
  if (analysisStatus.value.status === 'failed') return 'exception'
  return undefined
})

const currentStageText = computed(() => {
  const progress = analysisStatus.value.progress
  if (progress < 20) return '准备分析环境'
  if (progress < 40) return '分析需求结构'
  if (progress < 60) return 'AI智能优化'
  if (progress < 80) return 'NESMA评估计算'
  if (progress < 100) return '生成分析报告'
  return '分析完成'
})

// 监听taskId变化
watch(() => props.taskId, (newTaskId) => {
  if (newTaskId && props.modelValue) {
    startProgressPolling()
  }
}, { immediate: true })

// 监听对话框显示状态
watch(() => props.modelValue, (visible) => {
  if (visible && props.taskId) {
    startProgressPolling()
  } else {
    stopProgressPolling()
  }
})

// 方法
const startProgressPolling = () => {
  if (!props.taskId) {
    console.log('启动轮询失败：taskId 为空')
    return
  }
  
  console.log(`开始轮询任务进度，taskId: ${props.taskId}`)
  stopProgressPolling() // 先停止之前的轮询
  
  // 智能延迟轮询：给Redis一些时间保存任务状态
  // 第一次查询延迟1秒，避免"task_id不存在"错误
  initialDelayTimer = setTimeout(() => {
    console.log('执行第一次进度查询')
    fetchProgress()
    
    // 启动定期轮询：每2秒查询一次
    progressInterval = setInterval(() => {
      fetchProgress()
    }, 2000)
    console.log('启动定期轮询，间隔2秒')
  }, 1000)
}

const stopProgressPolling = () => {
  console.log('停止进度轮询，清理所有定时器...')
  
  // 清理初始延迟定时器
  if (initialDelayTimer) {
    console.log('清理初始延迟定时器')
    clearTimeout(initialDelayTimer)
    initialDelayTimer = null
  }
  
  // 清理轮询定时器
  if (progressInterval) {
    console.log('清理轮询定时器')
    clearInterval(progressInterval)
    progressInterval = null
  }
  
  // 清理所有重试定时器
  if (retryTimers.size > 0) {
    console.log(`清理 ${retryTimers.size} 个重试定时器`)
    retryTimers.forEach(timerId => {
      clearTimeout(timerId)
    })
    retryTimers.clear()
  }
  
  console.log('所有定时器已清理完成')
}

const fetchProgress = async (retryCount = 0) => {
  try {
    const response = await getUnifiedAnalysisProgress(props.taskId)
    const data = response.data
    
    // 使用增强的分析状态对象
    analysisStatus.value = enhanceAnalysisStatus(data)

    // 添加进度日志
    if (data.message) {
      progressLogs.value.push({
        timestamp: new Date(),
        message: data.message
      })
    }

    // 如果分析完成，获取详细结果
    if (data.status === 'completed') {
      stopProgressPolling()
      await fetchDetailedResult()
    }
    
    // 如果分析失败，停止轮询
    if (data.status === 'failed') {
      stopProgressPolling()
    }
  } catch (error) {
    console.error('获取分析进度失败:', error)
    
    // 智能重试机制：针对"任务不存在"错误进行特殊处理
    if (error?.response?.data?.msg?.includes('不存在') || 
        error?.response?.data?.msg?.includes('invalid') ||
        error?.response?.status === 404) {
      
      // 如果是前3次轮询且遇到任务不存在错误，进行智能重试
      if (retryCount < 3) {
        console.log(`任务 ${props.taskId} 暂时不存在，进行第 ${retryCount + 1} 次重试...`)
        
        // 递增延迟重试：1秒、2秒、3秒
        const retryTimer = setTimeout(() => {
          retryTimers.delete(retryTimer) // 从集合中移除已执行的定时器
          // 检查组件是否仍然挂载
          if (props.taskId) {
            fetchProgress(retryCount + 1)
          }
        }, (retryCount + 1) * 1000)
        
        retryTimers.add(retryTimer) // 添加到跟踪集合
        console.log(`设置重试定时器，第 ${retryCount + 1} 次重试，延迟 ${(retryCount + 1) * 1000}ms`)
        return
      } else {
        // 超过重试次数，显示友好错误提示
        console.error(`任务 ${props.taskId} 在多次重试后仍不存在，可能任务创建失败`)
        ElMessage.warning('任务状态同步中，请稍候再试或刷新页面')
      }
    }
    
    // 其他类型的错误，不显示频繁提示避免用户困扰
  }
}

const apiData = ref(null)

const fetchDetailedResult = async () => {
  try {
    const response = await getDetailedAnalysisResult(props.taskId, 'detailed')
    apiData.value = response.data
    
    // 解析后端返回的数据结构
    const resultData = apiData.result || apiData
    analysisResult.value = apiData
    
    console.log('API返回的完整数据:', apiData)
    console.log('分析结果数据:', resultData)
    

    
    // 提取核心统计数据
    const totalItems = apiData?.value.statistics?.total_requirements
    const processedItems = apiData?.value.statistics?.processed_count
    const successRate = apiData?.value.statistics?.success_rate
    const avgConfidence = apiData?.value.statistics?.average_confidence
    
    // 统计实际的功能类型分布
    const functionTypeStats = {
      'EI': 0, 'EO': 0, 'EQ': 0, 'ILF': 0, 'EIF': 0
    }
    const complexityStats = {
      'Low': 0, 'Average': 0, 'High': 0
    }
    
    // 从analyzed_requirements中统计实际数据
    if (apiData.value.analyzed_requirements && Array.isArray(apiData.value.analyzed_requirements)) {
      apiData.value.analyzed_requirements.forEach(item => {
        if (item.optimized_requirement?.functionType) {
          const funcType = item.optimized_requirement.functionType
          if (functionTypeStats.hasOwnProperty(funcType)) {
            functionTypeStats[funcType]++
          }
        }
        
        if (item.optimized_requirement?.complexityLevel) {
          const complexity = item.optimized_requirement.complexityLevel
          if (complexityStats.hasOwnProperty(complexity)) {
            complexityStats[complexity]++
          }
        }
      })
    }
    
    // 将后端数据结构转换为前端期望的格式
    detailedResult.value = {
      // 基础统计信息 - 优先使用实际API数据
      functionPointsCount: totalItems,
      qualityScore: safeGetNumber(avgConfidence, 0.85),
      
      // NESMA分析数据 - 基于实际统计结果
      nesmaAnalysis: {
        functionTypes: functionTypeStats,
        complexityDistribution: complexityStats,
        totalFunctionPoints: totalItems,
        avgConfidence: safeGetNumber(avgConfidence, 0.85),
        processingTime: parseProcessingTime(apiData.value.statistics?.processing_time || '0s')
      },
    }
    
    // 数据验证和一致性检查
    if (detailedResult.value.functionPointsCount === 0 && totalItems > 0) {
      console.warn('功能点统计数据可能有误，使用备用数据源')
      detailedResult.value.functionPointsCount = totalItems
    }
    
    if (detailedResult.value.improvementCount === 0 && detailedResult.value.optimizations.length > 0) {
      detailedResult.value.improvementCount = detailedResult.value.optimizations.length
    }
    
    console.log('转换后的详细结果:', detailedResult.value)
    console.log('数据映射验证:', {
      原始总项目数: totalItems,
      映射后功能点数: detailedResult.value.functionPointsCount,
      原始建议数: safeGetArrayLength(apiData.value.analyzed_requirements),
      映射后改进数: detailedResult.value.improvementCount,
      质量评分: detailedResult.value.qualityScore
    })
  } catch (error) {
    console.error('获取详细分析结果失败:', error)
    ElMessage.warning('获取详细分析结果失败，但基本分析已完成')
    
    // 设置最小可用的默认数据，避免页面崩溃
    detailedResult.value = {
      functionPointsCount: apiData.value.metadata?.totalRequirements || 0,
      nesmaAnalysis: {
        functionTypes: { 'EI': 0, 'EO': 0, 'EQ': 0, 'ILF': 0, 'EIF': 0 },
        complexityDistribution: { 'Low': 0, 'Average': 0, 'High': 0 },
        totalFunctionPoints: 0,
        avgConfidence: 0.8,
        processingTime: 0
      }
      
    }
    
    console.warn('使用默认数据避免页面错误:', detailedResult.value)
  }
}

// 安全获取数值的辅助函数
const safeGetNumber = (value, defaultValue = 0) => {
  if (typeof value === 'number' && !isNaN(value)) return value
  if (typeof value === 'string') {
    const parsed = parseFloat(value)
    return !isNaN(parsed) ? parsed : defaultValue
  }
  return defaultValue
}

// 安全获取数组长度
const safeGetArrayLength = (arr) => Array.isArray(arr) ? arr.length : 0

// 解析处理时间字符串为秒数
const parseProcessingTime = (timeStr) => {
  if (!timeStr || typeof timeStr !== 'string') return 0
  
  // 处理格式如 "5m36.092202s" 或 "1h2m30s" 等
  const timeMatch = timeStr.match(/(?:(\d+)h)?(?:(\d+)m)?(?:(\d+(?:\.\d+)?)s)?/)
  if (!timeMatch) return 0
  
  const hours = parseInt(timeMatch[1] || 0)
  const minutes = parseInt(timeMatch[2] || 0)
  const seconds = parseFloat(timeMatch[3] || 0)
  
  return Math.round(hours * 3600 + minutes * 60 + seconds)
}

const getPriorityTagType = (priority) => {
  if (typeof priority === 'number') {
    return priority === 1 ? 'danger' : priority === 2 ? 'warning' : 'info'
  }
  switch (priority?.toLowerCase()) {
    case 'high': case '高': return 'danger'
    case 'medium': case '中': return 'warning'
    case 'low': case '低': return 'info'
    default: return ''
  }
}

const getImpactTagType = (impact) => {
  switch (impact?.toLowerCase()) {
    case 'high': return 'danger'
    case 'medium': return 'warning'
    case 'low': return 'info'
    default: return ''
  }
}

const getScoreCategoryName = (category) => {
  const categoryNames = {
    'ai_confidence': 'AI置信度',
    'data_quality': '数据质量',
    'processing_efficiency': '处理效率',
    'result_accuracy': '结果准确性'
  }
  return categoryNames[category] || category
}

const getScoreClass = (score) => {
  if (score >= 90) return 'score-excellent'
  if (score >= 80) return 'score-good'
  if (score >= 70) return 'score-average'
  return 'score-poor'
}

const viewRecommendationDetail = (recommendation) => {
  const content = `
    <div>
      <p><strong>描述:</strong> ${recommendation.description}</p>
      ${recommendation.rationale ? `<p><strong>原因:</strong> ${recommendation.rationale}</p>` : ''}
      ${recommendation.timeline ? `<p><strong>时间线:</strong> ${recommendation.timeline}</p>` : ''}
      ${recommendation.effort ? `<p><strong>工作量:</strong> ${recommendation.effort}</p>` : ''}
      ${recommendation.acceptance_rate ? `<p><strong>接受率:</strong> ${Math.round(recommendation.acceptance_rate * 100)}%</p>` : ''}
    </div>
  `
  
  ElMessageBox.alert(content, recommendation.title || '建议详情', {
    confirmButtonText: '知道了',
    type: 'info',
    dangerouslyUseHTMLString: true
  })
}

const applyRecommendation = async (recommendation) => {
  try {
    await ElMessageBox.confirm(
      `确定要应用建议 "${recommendation.title}" 吗？`,
      '应用建议',
      {
        confirmButtonText: '确定应用',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    ElMessage.success('建议应用成功')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('应用建议失败: ' + error.message)
    }
  }
}

const viewOptimizationDetail = (optimization) => {
  ElMessageBox.alert(
    optimization.description || optimization.suggestion,
    optimization.title || '优化建议详情',
    {
      confirmButtonText: '知道了',
      type: 'info'
    }
  )
}

const applyOptimization = async (optimization) => {
  try {
    await ElMessageBox.confirm(
      '确定要应用这个优化建议吗？此操作将修改相关需求。',
      '应用优化建议',
      {
        confirmButtonText: '确定应用',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const applyData = {
      task_id: props.taskId,
      optimization_id: optimization.id || optimization.suggestionId,
      suggestion_id: optimization.suggestionId || optimization.id,
      action: 'accept'
    }

    await applyOptimizationApi(applyData)
    ElMessage.success('优化建议已应用')
    
    // 重新获取详细结果
    await fetchDetailedResult()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('应用优化建议失败: ' + (error.response?.data?.message || error.message))
    }
  }
}

const cancelAnalysis = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要取消当前分析吗？已处理的部分结果将会保留。',
      '取消分析',
      {
        confirmButtonText: '确定取消',
        cancelButtonText: '继续分析',
        type: 'warning'
      }
    )

    await cancelUnifiedAnalysis({ task_id: props.taskId })
    ElMessage.success('分析已取消')
    
    stopProgressPolling()
    emit('close')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消分析失败: ' + (error.response?.data?.message || error.message))
    }
  }
}

const retryAnalysis = () => {
  // 确保停止当前轮询
  stopProgressPolling()
  emit('close')
  // 触发重新分析 - 父组件需要处理
  setTimeout(() => {
    // 重新启动分析
    emit('retry', props.project)
  }, 500)
}

const exportReport = async () => {
  try {
    // 显示导出选项对话框
    const exportOptions = await ElMessageBox.prompt(
      '请选择导出格式和选项',
      '导出报告',
      {
        confirmButtonText: '导出',
        cancelButtonText: '取消',
        inputType: 'select',
        inputValue: 'pdf',
        inputPlaceholder: '选择导出格式',
        inputOptions: [
          { label: 'PDF报告', value: 'pdf' },
          { label: 'Excel数据', value: 'excel' },
          { label: 'Word文档', value: 'word' },
          { label: 'HTML网页', value: 'html' }
        ],
        inputValidator: (value) => {
          if (!value) {
            return '请选择导出格式'
          }
          return true
        }
      }
    )

    if (!exportOptions.value) return

    const exportFormat = exportOptions.value
    const includeDetails = await ElMessageBox.confirm(
      '是否包含详细分析数据？',
      '导出选项',
      {
        confirmButtonText: '包含详情',
        cancelButtonText: '仅摘要',
        type: 'info'
      }
    ).then(() => true).catch(() => false)

    // 构建导出数据
    const exportData = {
      task_id: props.taskId,
      export_format: exportFormat,
      include_details: includeDetails,
      include_insights: true,
      include_recommendations: true,
      include_quality_score: true,
      include_metadata: true,
      include_processing_stats: true,
      report_title: `分析报告_${props.project?.name || '项目'}_${new Date().toLocaleDateString()}`,
      custom_sections: [
        'executive_summary',
        'analysis_results',
        'quality_assessment',
        'insights_and_recommendations',
        'technical_details',
        'appendix'
      ]
    }

    // 显示加载状态
    const loadingInstance = ElLoading.service({
      lock: true,
      text: `正在生成${exportFormat.toUpperCase()}报告...`,
      background: 'rgba(0, 0, 0, 0.7)'
    })

    try {
      const response = await exportUnifiedAnalysisReport(exportData)
      
      // 根据响应类型处理下载
      if (response.data instanceof Blob) {
        // 二进制数据，直接下载
        const url = window.URL.createObjectURL(response.data)
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `analysis_report_${props.taskId}_${new Date().getTime()}.${exportFormat}`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
      } else if (response.data && response.data.download_url) {
        // 服务器返回下载链接
        window.open(response.data.download_url, '_blank')
      } else if (response.data && response.data.file_content) {
        // Base64编码的文件内容
        const binaryString = atob(response.data.file_content)
        const bytes = new Uint8Array(binaryString.length)
        for (let i = 0; i < binaryString.length; i++) {
          bytes[i] = binaryString.charCodeAt(i)
        }
        const blob = new Blob([bytes], { type: response.data.mime_type || 'application/octet-stream' })
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', response.data.filename || `analysis_report_${props.taskId}.${exportFormat}`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
      } else {
        // 其他情况，尝试作为JSON处理
        const reportData = response.data
        const blob = new Blob([JSON.stringify(reportData, null, 2)], { type: 'application/json' })
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', `analysis_report_${props.taskId}.json`)
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
      }
      
      ElMessage.success(`报告导出成功！格式: ${exportFormat.toUpperCase()}`)
      
      // 记录导出历史
      console.log('报告导出记录:', {
        taskId: props.taskId,
        format: exportFormat,
        timestamp: new Date().toISOString(),
        includeDetails: includeDetails
      })
      
    } finally {
      loadingInstance.close()
    }
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('导出报告失败:', error)
      ElMessage.error('导出报告失败: ' + (error.response?.data?.message || error.message))
    }
  }
}

const previewReport = async () => {
  try {
    // 构建预览数据
    const previewData = {
      task_id: props.taskId,
      export_format: 'html',
      include_details: true,
      include_insights: true,
      include_recommendations: true,
      include_quality_score: true,
      include_metadata: true,
      include_processing_stats: true,
      preview_mode: true,
      report_title: `分析报告_${props.project?.name || '项目'}_${new Date().toLocaleDateString()}`,
      custom_sections: [
        'executive_summary',
        'analysis_results',
        'quality_assessment',
        'insights_and_recommendations',
        'technical_details'
      ]
    }

    const loadingInstance = ElLoading.service({
      lock: true,
      text: '正在生成预览报告...',
      background: 'rgba(0, 0, 0, 0.7)'
    })

    try {
      const response = await exportUnifiedAnalysisReport(previewData)
      
      // 处理预览响应
      if (response.data && response.data.html_content) {
        // 在新窗口中打开预览
        const newWindow = window.open('', '_blank')
        newWindow.document.write(response.data.html_content)
        newWindow.document.close()
      } else if (response.data && response.data.preview_url) {
        // 打开预览链接
        window.open(response.data.preview_url, '_blank')
      } else {
        // 显示预览对话框
        ElMessageBox.alert(
          `<div style="max-height: 400px; overflow-y: auto;">
            <h3>报告预览</h3>
            <pre style="white-space: pre-wrap; font-size: 12px;">${JSON.stringify(response.data, null, 2)}</pre>
          </div>`,
          '报告预览',
          {
            dangerouslyUseHTMLString: true,
            confirmButtonText: '确定',
            customClass: 'preview-dialog'
          }
        )
      }
      
      ElMessage.success('预览报告生成成功')
      
    } finally {
      loadingInstance.close()
    }
    
  } catch (error) {
    console.error('预览报告失败:', error)
    ElMessage.error('预览报告失败: ' + (error.response?.data?.message || error.message))
  }
}

const openReportGenerator = () => {
  if (reportGeneratorRef.value) {
    reportGeneratorRef.value.openReportOptions()
  }
}

const handleReportDownload = (downloadInfo) => {
  console.log('报告下载完成:', downloadInfo)
  ElMessage.success(`报告下载成功: ${downloadInfo.filename}`)
}

const viewFullResult = () => {
  // 跳转到详细分析结果页面
  emit('view-full-result', {
    taskId: props.taskId,
    project: props.project,
    result: detailedResult.value
  })
}

const handleBeforeClose = (done) => {
  if (analysisStatus.value.status === 'running') {
    ElMessageBox.confirm(
      '分析正在进行中，确定要关闭吗？分析将在后台继续进行。',
      '关闭确认',
      {
        confirmButtonText: '确定关闭',
        cancelButtonText: '继续查看',
        type: 'warning'
      }
    ).then(() => {
      // 用户确认关闭，停止轮询
      stopProgressPolling()
      done()
    }).catch(() => {
      // 取消关闭，继续轮询
    })
  } else {
    // 分析已完成或失败，直接关闭
    stopProgressPolling()
    done()
  }
}

const handleClose = () => {
  // 确保停止所有轮询和定时器
  stopProgressPolling()
  
  if (analysisStatus.value.status === 'completed') {
    emit('complete', {
      taskId: props.taskId,
      result: analysisResult.value
    })
  } else {
    emit('close')
  }
}

const formatTime = (time) => {
  if (!time) return '--'
  const date = new Date(time)
  return date.toLocaleTimeString('zh-CN', { 
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 新增方法：动画效果支持方法
const getProgressStatus = () => {
  if (analysisStatus.value.status === 'completed') return 'success'
  if (analysisStatus.value.status === 'failed') return 'exception'
  return undefined
}

const getStageIcon = () => {
  // 安全的图标映射，确保所有图标都已导入
  const iconMap = {
    'initializing': 'Timer',
    'validating': 'Search',
    'fetching': 'Upload',
    'context_building': 'FolderOpened',
    'ai_analyzing': 'MagicStick',
    'ai_calling': 'Upload',
    'parsing': 'Document',
    'finalizing': 'Check',
    'completed': 'Check',
    'failed': 'Warning'
  }
  
  // 安全获取图标名称
  const currentStage = analysisStatus.value.currentStage || 'initializing'
  const iconName = iconMap[currentStage] || 'Timer'
  
  // 返回实际的图标组件
  const icons = {
    Timer, Search, Upload, FolderOpened, MagicStick, 
    Document, Check, Warning
  }
  
  return icons[iconName] || Timer
}

const getStageIconClass = () => {
  const classMap = {
    'pulse': 'pulse-animation',
    'scan': 'scan-animation',
    'loading': 'loading-animation',
    'network': 'network-animation',
    'brain': 'brain-animation',
    'api': 'api-animation',
    'parse': 'parse-animation',
    'check': 'check-animation',
    'success': 'success-animation',
    'error': 'error-animation'
  }
  return classMap[analysisStatus.value.animationType] || 'default-animation'
}

const getStageColor = () => {
  if (analysisStatus.value.status === 'failed') return '#F56C6C'
  if (analysisStatus.value.status === 'completed') return '#67C23A'
  
  const colorMap = {
    'initializing': '#409EFF',
    'validating': '#E6A23C',
    'fetching': '#409EFF',
    'context_building': '#909399',
    'ai_analyzing': '#722ED1',
    'ai_calling': '#13C2C2',
    'parsing': '#52C41A',
    'finalizing': '#1890FF'
  }
  return colorMap[analysisStatus.value.currentStage] || '#409EFF'
}

const getTimelineType = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'running') return 'primary'
  return 'info'
}

// 增强分析状态对象
const enhanceAnalysisStatus = (data) => {
  // 将后端数据映射为增强的分析状态结构
  return {
    ...analysisStatus.value,
    status: data.status,
    progress: data.progress || 0,
    currentStage: data.current_stage || data.currentStage || 'analyzing',
    stageDesc: data.stage_desc || data.stageDesc || data.message || '正在分析中...',
    animationType: data.animation_type || data.animationType || 'pulse',
    statusText: data.status_text || data.statusText || '分析进行中',
    errorMsg: data.error_msg || data.errorMsg || '',
    stageHistory: data.stage_history || data.stageHistory || [],
    animationData: data.animation_data || data.animationData || {},
    metadata: data.metadata || null,
    message: data.message || '',
    endTime: data.endTime ? new Date(data.endTime) : null,
    duration: data.duration || 0,
    error: data.error || null
  }
}

// 生命周期
onMounted(() => {
  if (props.modelValue && props.taskId) {
    startProgressPolling()
  }
})

onBeforeUnmount(() => {
  stopProgressPolling()
})
</script>

<style lang="scss" scoped>
// 增强的分析进度对话框样式
.enhanced-analysis-progress {
  .project-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding: 16px;
    background: linear-gradient(135deg, #f8fafc 0%, #e3f2fd 100%);
    border-radius: 12px;
    border: 1px solid #e1e8ed;
    
    .project-title {
      display: flex;
      align-items: center;
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      
      .project-icon {
        margin-right: 8px;
        color: #409EFF;
      }
    }
  }
  
  .animated-progress-section {
    .main-progress {
      margin-bottom: 24px;
      
      .progress-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 12px;
        
        .progress-percent {
          font-size: 24px;
          font-weight: 700;
          color: #409EFF;
        }
        
        .stage-name {
          font-size: 14px;
          color: #606266;
          font-weight: 500;
        }
      }
    }
    
    .animation-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin: 32px 0;
      
      .stage-icon-container {
        width: 80px;
        height: 80px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3);
        margin-bottom: 16px;
        transition: all 0.3s ease;
        
        .stage-icon {
          font-size: 36px;
          color: white;
          transition: all 0.3s ease;
        }
        
        &.pulse {
          animation: pulse 2s infinite;
        }
        
        &.scan {
          animation: scan 1.5s ease-in-out infinite;
        }
        
        &.loading {
          animation: rotate 1s linear infinite;
        }
        
        &.network {
          animation: network 2s ease-in-out infinite;
        }
        
        &.brain {
          background: linear-gradient(135deg, #722ED1 0%, #B37FEB 100%);
          animation: brain 1.8s ease-in-out infinite;
        }
        
        &.api {
          background: linear-gradient(135deg, #13C2C2 0%, #36CFC9 100%);
          animation: api 1.2s ease-in-out infinite;
        }
        
        &.success {
          background: linear-gradient(135deg, #52C41A 0%, #73D13D 100%);
          animation: success 0.8s ease-out;
        }
        
        &.error {
          background: linear-gradient(135deg, #F56C6C 0%, #FF7875 100%);
          animation: error 0.5s ease-out;
        }
      }
      
      .status-text {
        text-align: center;
        
        .primary-text {
          font-size: 16px;
          color: #303133;
          margin: 0 0 8px 0;
          font-weight: 500;
        }
        
        .secondary-text {
          font-size: 14px;
          color: #909399;
          margin: 0;
        }
      }
    }
    
    .stage-timeline {
      margin-top: 32px;
      
      h4 {
        font-size: 16px;
        color: #303133;
        margin: 0 0 16px 0;
        text-align: center;
      }
      
      .timeline-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
        
        .stage-desc {
          font-size: 14px;
          color: #606266;
        }
      }
    }
  }
  
  .processing-stats {
    margin: 24px 0;
    padding: 20px;
    background: #f8f9fa;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
  }
  
  .analysis-features {
    margin-top: 32px;
    padding: 20px;
    background: #fafbfc;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
    
    .feature-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;
      
      .feature-item {
        display: flex;
        align-items: center;
        font-size: 14px;
        color: #606266;
        
        .el-icon {
          margin-right: 8px;
          color: #409EFF;
        }
      }
    }
  }
  
  .progress-logs {
    margin-top: 24px;
    
    h4 {
      margin: 0 0 12px 0;
      font-size: 14px;
      color: #303133;
    }
    
    .log-container {
      background: #fafafa;
      border: 1px solid #ebeef5;
      border-radius: 4px;
      padding: 12px;
      max-height: 120px;
      overflow-y: auto;
      
      .log-item {
        display: flex;
        margin-bottom: 8px;
        font-size: 12px;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        .log-time {
          color: #909399;
          margin-right: 8px;
          flex-shrink: 0;
        }
        
        .log-message {
          color: #606266;
        }
      }
    }
  }
}

// 动画定义
@keyframes pulse {
  0%, 100% { transform: scale(1); box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3); }
  50% { transform: scale(1.05); box-shadow: 0 12px 30px rgba(102, 126, 234, 0.5); }
}

@keyframes scan {
  0%, 100% { transform: scale(1) rotate(0deg); }
  25% { transform: scale(1.02) rotate(2deg); }
  75% { transform: scale(1.02) rotate(-2deg); }
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes network {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.03); }
}

@keyframes brain {
  0%, 100% { transform: scale(1); filter: hue-rotate(0deg); }
  33% { transform: scale(1.02); filter: hue-rotate(10deg); }
  66% { transform: scale(1.01); filter: hue-rotate(-10deg); }
}

@keyframes api {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

@keyframes success {
  0% { transform: scale(0.8); opacity: 0.8; }
  50% { transform: scale(1.1); opacity: 1; }
  100% { transform: scale(1); opacity: 1; }
}

@keyframes error {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-2px); }
  75% { transform: translateX(2px); }
}

// 保留原有样式
.analysis-progress {
  .progress-header {
    text-align: center;
    margin-bottom: 24px;
    
    h3 {
      margin: 0 0 8px 0;
      color: #303133;
      font-size: 18px;
    }
    
    .task-info {
      margin: 0;
      color: #909399;
      font-size: 14px;
    }
  }
  
  .overall-progress {
    margin-bottom: 24px;
    
    .progress-text {
      font-weight: 500;
      color: #409EFF;
    }
  }
  
  .progress-details {
    .current-stage {
      background: #f5f7fa;
      padding: 16px;
      border-radius: 8px;
      margin-bottom: 16px;
      
      h4 {
        margin: 0 0 8px 0;
        color: #409EFF;
        font-size: 16px;
      }
      
      p {
        margin: 0;
        color: #606266;
      }
    }
    
    .processing-stats {
      margin-bottom: 16px;
    }
    
    .progress-logs {
      h4 {
        margin: 0 0 12px 0;
        font-size: 14px;
        color: #303133;
      }
      
      .log-container {
        background: #fafafa;
        border: 1px solid #ebeef5;
        border-radius: 4px;
        padding: 12px;
        max-height: 120px;
        overflow-y: auto;
        
        .log-item {
          display: flex;
          margin-bottom: 8px;
          font-size: 12px;
          
          &:last-child {
            margin-bottom: 0;
          }
          
          .log-time {
            color: #909399;
            margin-right: 8px;
            flex-shrink: 0;
          }
          
          .log-message {
            color: #606266;
          }
        }
      }
    }
  }
}

.analysis-result {
  .result-header {
    text-align: center;
    margin-bottom: 24px;
    
    .completion-stats {
      margin-top: 16px;
    }
  }
  
  .detailed-result {
    margin-top: 24px;
    
    .summary-content {
      .quality-breakdown {
        margin-top: 24px;
        
        h4 {
          margin-bottom: 16px;
          color: #333;
          font-weight: 600;
        }
        
        .score-card {
          text-align: center;
          
          .score-item {
            .score-title {
              font-size: 14px;
              color: #666;
              margin-bottom: 8px;
            }
            
            .score-value {
              font-size: 20px;
              font-weight: 600;
              
              &.score-excellent {
                color: #67c23a;
              }
              
              &.score-good {
                color: #409eff;
              }
              
              &.score-average {
                color: #e6a23c;
              }
              
              &.score-poor {
                color: #f56c6c;
              }
            }
          }
        }
      }
      
      .summary-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
        gap: 16px;
        margin-top: 16px;
        
        .summary-item {
          background: #f8f9fa;
          padding: 16px;
          border-radius: 8px;
          border-left: 4px solid #409EFF;
          
          h5 {
            margin: 0 0 8px 0;
            color: #303133;
            font-size: 14px;
          }
          
          p {
            margin: 4px 0;
            color: #606266;
            font-size: 13px;
          }
        }
      }
    }
    
    .optimizations-content {
      .insights-section, .recommendations-section {
        margin-bottom: 24px;
        
        h4 {
          margin-bottom: 16px;
          color: #333;
          font-weight: 600;
        }
        
        .insight-card, .recommendation-card {
          margin-bottom: 16px;
          
          .insight-header, .rec-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 16px;
            
            .insight-title-area, .rec-title-area {
              flex: 1;
              
              h5 {
                margin: 0 0 8px 0;
                color: #333;
                font-size: 16px;
                font-weight: 600;
              }
              
              .insight-meta, .rec-meta {
                display: flex;
                gap: 8px;
                flex-wrap: wrap;
              }
            }
            
            .confidence-score, .acceptance-rate {
              text-align: right;
              
              .confidence-label, .rate-label {
                display: block;
                font-size: 12px;
                color: #666;
                margin-bottom: 4px;
              }
              
              .confidence-value, .rate-value {
                font-size: 18px;
                font-weight: 600;
                color: #409eff;
              }
            }
          }
          
          .insight-content, .rec-content {
            .insight-description, .rec-description {
              margin-bottom: 16px;
              color: #333;
              line-height: 1.5;
            }
            
            .evidence-section, .actions-section, .rationale-section, .implementation-section, .tags-section {
              margin-bottom: 12px;
              
              h6 {
                margin: 0 0 8px 0;
                color: #666;
                font-size: 14px;
                font-weight: 600;
              }
              
              ul, ol {
                margin: 0;
                padding-left: 20px;
                
                li {
                  margin-bottom: 4px;
                  color: #666;
                  line-height: 1.4;
                }
              }
              
              p {
                margin: 0;
                color: #666;
                line-height: 1.4;
              }
              
              .impl-timeline {
                margin-top: 8px;
                padding: 8px;
                background: #f8f9fa;
                border-radius: 4px;
              }
              
              .tag-item {
                margin-right: 8px;
                margin-bottom: 4px;
              }
            }
          }
          
          .rec-actions {
            margin-top: 16px;
            padding-top: 16px;
            border-top: 1px solid #ebeef5;
            display: flex;
            gap: 8px;
          }
        }
      }
      
      .optimization-item {
        border: 1px solid #ebeef5;
        border-radius: 8px;
        padding: 16px;
        margin-bottom: 12px;
        
        .opt-header {
          display: flex;
          align-items: center;
          margin-bottom: 8px;
          
          .opt-title {
            margin-left: 8px;
            font-weight: 500;
            color: #303133;
          }
        }
        
        .opt-description {
          margin: 8px 0;
          color: #606266;
          line-height: 1.5;
        }
        
        .opt-actions {
          text-align: right;
        }
      }
    }
    
    .nesma-content {
      .nesma-stats {
        margin: 16px 0;
      }
      
      .complexity-chart {
        margin-top: 16px;
        
        .el-tag {
          margin-right: 8px;
          margin-bottom: 8px;
        }
      }
    }
    
    .knowledge-content {
      .tech-section {
        margin-bottom: 24px;
        
        h5 {
          margin-bottom: 12px;
          color: #333;
          font-weight: 600;
          font-size: 16px;
        }
        
        .el-descriptions {
          margin-bottom: 16px;
        }
        
        .quality-control-item {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 8px;
          padding: 8px;
          background: #f8f9fa;
          border-radius: 4px;
          
          .control-message {
            flex: 1;
            color: #666;
          }
          
          .control-score {
            font-weight: 600;
            color: #333;
          }
        }
      }
      
      .knowledge-item {
        border-bottom: 1px solid #ebeef5;
        padding: 12px 0;
        
        &:last-child {
          border-bottom: none;
        }
        
        h5 {
          margin: 0 0 8px 0;
          color: #303133;
          font-size: 14px;
        }
        
        p {
          margin: 8px 0;
          color: #606266;
          font-size: 13px;
          line-height: 1.4;
        }
      }
    }
  }
}

.analysis-error {
  text-align: center;
}

.dialog-footer {
  text-align: right;
}
</style>