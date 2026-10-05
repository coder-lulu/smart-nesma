<template>
  <el-dialog 
    v-model="visible" 
    title="L4功能点生成进度追踪" 
    width="900px" 
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div class="l4-progress-tracker">
      <!-- 任务基本信息 -->
      <div class="task-info" v-if="taskInfo">
        <el-card shadow="never" class="info-card">
          <div class="task-header">
            <div class="task-title">
              <h3>{{ taskInfo.title || 'L4功能点生成任务' }}</h3>
              <el-tag :type="getStatusColor(taskInfo.status)" size="large">
                {{ getStatusText(taskInfo.status) }}
              </el-tag>
            </div>
            <div class="task-meta">
              <span class="meta-item">
                <el-icon><Document /></el-icon>
                L3需求: {{ taskInfo.l3Count || 0 }}个
              </span>
              <span class="meta-item">
                <el-icon><Operation /></el-icon>
                策略: {{ getStrategyText(taskInfo.strategy) }}
              </span>
              <span class="meta-item">
                <el-icon><Timer /></el-icon>
                耗时: {{ getRunningDuration() }}
              </span>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 整体进度 -->
      <div class="overall-progress">
        <el-card shadow="never" class="progress-card">
          <div class="progress-header">
            <h4>整体进度</h4>
            <span class="progress-text">{{ Math.round(overallProgress) }}%</span>
          </div>
          <el-progress 
            :percentage="overallProgress" 
            :status="getProgressStatus()"
            :stroke-width="20"
            :show-text="false"
            class="main-progress"
          />
          <div class="progress-stats">
            <div class="stat-item">
              <span class="stat-label">已完成</span>
              <span class="stat-value">{{ completedCount }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">进行中</span>
              <span class="stat-value">{{ runningCount }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">等待中</span>
              <span class="stat-value">{{ pendingCount }}</span>
            </div>
            <div class="stat-item">
              <span class="stat-label">失败</span>
              <span class="stat-value">{{ failedCount }}</span>
            </div>
          </div>
        </el-card>
      </div>

      <!-- 阶段进度 -->
      <div class="stage-progress">
        <el-card shadow="never" class="stages-card">
          <h4>生成阶段</h4>
          <div class="stages-container">
            <div 
              v-for="(stage, index) in stages" 
              :key="index"
              :class="['stage-item', getStageClass(stage)]"
            >
              <div class="stage-icon">
                <el-icon :class="getStageIconClass(stage)">
                  <component :is="getStageIcon(stage)" />
                </el-icon>
              </div>
              <div class="stage-content">
                <div class="stage-title">{{ stage.title }}</div>
                <div class="stage-desc">{{ stage.description }}</div>
                <div class="stage-time" v-if="stage.completedAt">
                  {{ formatTime(stage.completedAt) }}
                </div>
              </div>
              <div class="stage-progress" v-if="stage.progress !== undefined">
                <el-progress 
                  :percentage="stage.progress" 
                  :status="getStageProgressStatus(stage.status)"
                  :stroke-width="6"
                  :show-text="false"
                />
              </div>
            </div>
          </div>
        </el-card>
      </div>

      <!-- L3需求处理详情 -->
      <div class="l3-details">
        <el-card shadow="never" class="details-card">
          <div class="details-header">
            <h4>L3需求处理详情</h4>
            <el-button-group size="small">
              <el-button :type="detailView === 'all' ? 'primary' : ''" @click="detailView = 'all'">
                全部 ({{ l3Requirements.length }})
              </el-button>
              <el-button :type="detailView === 'running' ? 'primary' : ''" @click="detailView = 'running'">
                处理中 ({{ getL3Count('running') }})
              </el-button>
              <el-button :type="detailView === 'completed' ? 'primary' : ''" @click="detailView = 'completed'">
                已完成 ({{ getL3Count('completed') }})
              </el-button>
              <el-button :type="detailView === 'failed' ? 'primary' : ''" @click="detailView = 'failed'">
                失败 ({{ getL3Count('failed') }})
              </el-button>
            </el-button-group>
          </div>
          
          <div class="l3-list">
            <div 
              v-for="l3 in filteredL3Requirements" 
              :key="l3.id"
              class="l3-item"
            >
              <div class="l3-header">
                <div class="l3-title">
                  <el-icon><Folder /></el-icon>
                  {{ l3.title }}
                </div>
                <div class="l3-status">
                  <el-tag :type="getStatusColor(l3.status)" size="small">
                    {{ getStatusText(l3.status) }}
                  </el-tag>
                </div>
              </div>
              
              <div class="l3-progress" v-if="l3.progress !== undefined">
                <el-progress 
                  :percentage="l3.progress" 
                  :status="getProgressStatus(l3.status)"
                  :stroke-width="8"
                  :show-text="false"
                />
                <span class="progress-text">{{ l3.progress }}%</span>
              </div>
              
              <div class="l3-results" v-if="l3.results">
                <div class="result-stats">
                  <span class="result-item">
                    <el-icon><Check /></el-icon>
                    生成: {{ l3.results.generated || 0 }}
                  </span>
                  <span class="result-item">
                    <el-icon><Warning /></el-icon>
                    失败: {{ l3.results.failed || 0 }}
                  </span>
                  <span class="result-item">
                    <el-icon><Clock /></el-icon>
                    耗时: {{ formatDuration(l3.results.duration) }}
                  </span>
                </div>
                
                <!-- L4生成详情 -->
                <div class="l4-details" v-if="l3.results.l4Items && l3.results.l4Items.length > 0">
                  <el-collapse accordion>
                    <el-collapse-item :name="l3.id" :title="`查看生成的L4功能点 (${l3.results.l4Items.length})`">
                      <div class="l4-list">
                        <div 
                          v-for="l4 in l3.results.l4Items" 
                          :key="l4.id"
                          class="l4-item"
                        >
                          <div class="l4-header">
                            <span class="l4-title">{{ l4.title }}</span>
                            <el-tag :type="getFunctionTypeColor(l4.functionType)" size="small">
                              {{ l4.functionType }}
                            </el-tag>
                          </div>
                          <div class="l4-meta">
                            <span class="meta-item">复杂度: {{ l4.complexity }}</span>
                            <span class="meta-item">功能点: {{ l4.functionPoints }}</span>
                            <span class="meta-item">置信度: {{ Math.round(l4.confidence * 100) }}%</span>
                          </div>
                        </div>
                      </div>
                    </el-collapse-item>
                  </el-collapse>
                </div>
              </div>
              
              <div class="l3-error" v-if="l3.status === 'failed' && l3.errorMsg">
                <el-alert
                  :title="l3.errorMsg"
                  type="error"
                  :closable="false"
                  show-icon
                />
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 对话框底部 -->
    <template #footer>
      <div class="dialog-footer">
        <div class="footer-info">
          <span v-if="taskInfo && taskInfo.status === 'running'">
            <el-icon class="rotating"><Loading /></el-icon>
            任务正在进行中...
          </span>
          <span v-else-if="taskInfo && taskInfo.status === 'completed'">
            <el-icon class="success"><Check /></el-icon>
            任务已完成
          </span>
          <span v-else-if="taskInfo && taskInfo.status === 'failed'">
            <el-icon class="error"><Warning /></el-icon>
            任务失败
          </span>
        </div>
        <div class="footer-actions">
          <el-button @click="handleClose">关闭</el-button>
          <el-button 
            type="warning" 
            @click="handleCancel"
            v-if="taskInfo && (taskInfo.status === 'running' || taskInfo.status === 'pending')"
          >
            取消任务
          </el-button>
          <el-button 
            type="primary" 
            @click="handleViewResult"
            v-if="taskInfo && taskInfo.status === 'completed'"
          >
            查看结果
          </el-button>
        </div>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document, Operation, Timer, Check, Warning, Clock, Folder, Loading,
  Search, Upload, FolderOpened, MagicStick, DocumentChecked
} from '@element-plus/icons-vue'
import {
  getL4TaskProgress,
  cancelL4GenerationTask
} from '@/api/nesma'

const props = defineProps({
  modelValue: Boolean,
  taskId: [String, Number],
  autoRefresh: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['update:modelValue', 'task-completed', 'task-failed', 'task-cancelled', 'view-result'])

// 响应式数据
const taskInfo = ref(null)
const overallProgress = ref(0)
const stages = ref([])
const l3Requirements = ref([])
const detailView = ref('all')
const pollInterval = ref(null)
const startTime = ref(null)

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const completedCount = computed(() => {
  return l3Requirements.value.filter(l3 => l3.status === 'completed').length
})

const runningCount = computed(() => {
  return l3Requirements.value.filter(l3 => l3.status === 'running').length
})

const pendingCount = computed(() => {
  return l3Requirements.value.filter(l3 => l3.status === 'pending').length
})

const failedCount = computed(() => {
  return l3Requirements.value.filter(l3 => l3.status === 'failed').length
})

const filteredL3Requirements = computed(() => {
  if (detailView.value === 'all') {
    return l3Requirements.value
  }
  return l3Requirements.value.filter(l3 => l3.status === detailView.value)
})

// 监听对话框显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.taskId) {
    startTracking()
  } else {
    stopTracking()
  }
})

// 监听任务ID变化
watch(() => props.taskId, (newTaskId) => {
  if (newTaskId && props.modelValue) {
    startTracking()
  }
})

// 方法
const startTracking = () => {
  if (pollInterval.value) {
    clearInterval(pollInterval.value)
  }
  
  startTime.value = new Date()
  loadTaskProgress()
  
  if (props.autoRefresh) {
    pollInterval.value = setInterval(() => {
      loadTaskProgress()
    }, 2000) // 每2秒刷新一次
  }
}

const stopTracking = () => {
  if (pollInterval.value) {
    clearInterval(pollInterval.value)
    pollInterval.value = null
  }
}

const loadTaskProgress = async () => {
  if (!props.taskId) return
  
  try {
    const response = await getL4TaskProgress(props.taskId)
    
    if (response.code === 0) {
      const data = response.data
      
      taskInfo.value = data
      overallProgress.value = data.overallProgress
      stages.value = data.stages
      l3Requirements.value = data.l3Requirements
      
      // 检查任务状态
      if (data.status === 'completed') {
        stopTracking()
        emit('task-completed', data)
      } else if (data.status === 'failed') {
        stopTracking()
        emit('task-failed', data)
      }
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
    
  } catch (error) {
    console.error('获取任务进度失败:', error)
    ElMessage.error('获取任务进度失败：' + error.message)
    
    // 发生错误时，回退到模拟数据
    const mockProgress = {
      taskInfo: {
        id: props.taskId,
        title: 'L4功能点批量生成',
        status: 'running',
        strategy: 'comprehensive',
        l3Count: 5,
        progress: 65
      },
      overallProgress: 65,
      stages: [
        {
          id: 'initializing',
          title: '任务初始化',
          description: '准备生成环境和参数',
          status: 'completed',
          progress: 100,
          completedAt: new Date(Date.now() - 60000)
        },
        {
          id: 'analyzing',
          title: '需求分析',
          description: '分析L3需求内容',
          status: 'completed',
          progress: 100,
          completedAt: new Date(Date.now() - 45000)
        },
        {
          id: 'generating',
          title: 'L4生成',
          description: '生成L4功能点',
          status: 'running',
          progress: 65,
          completedAt: null
        },
        {
          id: 'optimizing',
          title: '结果优化',
          description: '优化生成结果',
          status: 'pending',
          progress: 0,
          completedAt: null
        }
      ],
      l3Requirements: [
        {
          id: 1,
          title: '用户登录功能',
          status: 'completed',
          progress: 100,
          results: {
            generated: 3,
            failed: 0,
            duration: 25,
            l4Items: [
              {
                id: 1,
                title: '用户名密码登录',
                functionType: 'EI',
                complexity: '中等',
                functionPoints: 4,
                confidence: 0.9
              },
              {
                id: 2,
                title: '验证码验证',
                functionType: 'EI',
                complexity: '简单',
                functionPoints: 3,
                confidence: 0.85
              },
              {
                id: 3,
                title: '登录日志记录',
                functionType: 'EO',
                complexity: '简单',
                functionPoints: 3,
                confidence: 0.8
              }
            ]
          }
        },
        {
          id: 2,
          title: '权限管理',
          status: 'running',
          progress: 45,
          results: {
            generated: 1,
            failed: 0,
            duration: 18,
            l4Items: [
              {
                id: 4,
                title: '角色权限分配',
                functionType: 'EI',
                complexity: '复杂',
                functionPoints: 6,
                confidence: 0.88
              }
            ]
          }
        },
        {
          id: 3,
          title: '数据备份',
          status: 'pending',
          progress: 0,
          results: null
        },
        {
          id: 4,
          title: '日志记录',
          status: 'pending',
          progress: 0,
          results: null
        },
        {
          id: 5,
          title: '系统监控',
          status: 'failed',
          progress: 0,
          results: null,
          errorMsg: '生成失败：缺少相关知识库内容'
        }
      ]
    }
    
    taskInfo.value = mockProgress.taskInfo
    overallProgress.value = mockProgress.overallProgress
    stages.value = mockProgress.stages
    l3Requirements.value = mockProgress.l3Requirements
    
    // 检查任务状态
    if (mockProgress.taskInfo.status === 'completed') {
      stopTracking()
      emit('task-completed', mockProgress.taskInfo)
    } else if (mockProgress.taskInfo.status === 'failed') {
      stopTracking()
      emit('task-failed', mockProgress.taskInfo)
    }
  }
}

const handleCancel = async () => {
  try {
    await ElMessageBox.confirm('确定要取消当前L4生成任务吗？', '确认取消', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await cancelL4GenerationTask(props.taskId)
    
    stopTracking()
    emit('task-cancelled', props.taskId)
    ElMessage.success('任务已取消')
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消任务失败：' + error.message)
    }
  }
}

const handleViewResult = () => {
  emit('view-result', props.taskId)
}

const handleClose = () => {
  stopTracking()
  emit('update:modelValue', false)
}

const getRunningDuration = () => {
  if (!startTime.value) return '0秒'
  const now = new Date()
  const duration = Math.floor((now - startTime.value) / 1000)
  return formatDuration(duration)
}

const getL3Count = (status) => {
  return l3Requirements.value.filter(l3 => l3.status === status).length
}

// 辅助方法
const getStatusColor = (status) => {
  const colorMap = {
    pending: 'warning',
    running: 'primary',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return colorMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    pending: '等待中',
    running: '进行中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return textMap[status] || status
}

const getStrategyText = (strategy) => {
  const textMap = {
    comprehensive: '全面生成',
    basic: '基础生成',
    intelligent: '智能生成'
  }
  return textMap[strategy] || strategy
}

const getProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const getStageClass = (stage) => {
  return `stage-${stage.status}`
}

const getStageIcon = (stage) => {
  const iconMap = {
    initializing: Timer,
    analyzing: Search,
    generating: MagicStick,
    optimizing: DocumentChecked
  }
  return iconMap[stage.id] || Timer
}

const getStageIconClass = (stage) => {
  return `icon-${stage.status}`
}

const getStageProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const getFunctionTypeColor = (type) => {
  const colorMap = {
    'EI': 'primary',
    'EO': 'success',
    'EQ': 'warning',
    'ILF': 'info',
    'EIF': 'danger'
  }
  return colorMap[type] || 'info'
}

const formatDuration = (seconds) => {
  if (!seconds) return '0秒'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return minutes > 0 ? `${minutes}分${secs}秒` : `${secs}秒`
}

const formatTime = (time) => {
  if (!time) return '-'
  return new Date(time).toLocaleTimeString('zh-CN')
}

// 生命周期
onMounted(() => {
  if (props.modelValue && props.taskId) {
    startTracking()
  }
})

onUnmounted(() => {
  stopTracking()
})
</script>

<style lang="scss" scoped>
.l4-progress-tracker {
  .task-info {
    margin-bottom: 20px;
    
    .info-card {
      border: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      
      .task-header {
        .task-title {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 12px;
          
          h3 {
            margin: 0;
            color: #303133;
          }
        }
        
        .task-meta {
          display: flex;
          gap: 20px;
          
          .meta-item {
            display: flex;
            align-items: center;
            color: #606266;
            font-size: 14px;
            
            .el-icon {
              margin-right: 4px;
            }
          }
        }
      }
    }
  }
  
  .overall-progress {
    margin-bottom: 20px;
    
    .progress-card {
      border: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      
      .progress-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
        
        h4 {
          margin: 0;
          color: #303133;
        }
        
        .progress-text {
          font-size: 20px;
          font-weight: 600;
          color: #409eff;
        }
      }
      
      .main-progress {
        margin-bottom: 16px;
      }
      
      .progress-stats {
        display: flex;
        justify-content: space-between;
        
        .stat-item {
          text-align: center;
          
          .stat-label {
            display: block;
            font-size: 12px;
            color: #909399;
          }
          
          .stat-value {
            display: block;
            font-size: 18px;
            font-weight: 600;
            color: #303133;
          }
        }
      }
    }
  }
  
  .stage-progress {
    margin-bottom: 20px;
    
    .stages-card {
      border: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      
      h4 {
        margin: 0 0 16px 0;
        color: #303133;
      }
      
      .stages-container {
        display: flex;
        gap: 16px;
        overflow-x: auto;
        
        .stage-item {
          flex: 1;
          min-width: 200px;
          padding: 16px;
          border: 1px solid #e4e7ed;
          border-radius: 8px;
          background: #fafafa;
          
          &.stage-completed {
            border-color: #67c23a;
            background: #f0f9ff;
          }
          
          &.stage-running {
            border-color: #409eff;
            background: #ecf5ff;
          }
          
          &.stage-failed {
            border-color: #f56c6c;
            background: #fef0f0;
          }
          
          .stage-icon {
            margin-bottom: 8px;
            
            .el-icon {
              font-size: 24px;
              
              &.icon-completed {
                color: #67c23a;
              }
              
              &.icon-running {
                color: #409eff;
              }
              
              &.icon-failed {
                color: #f56c6c;
              }
              
              &.icon-pending {
                color: #c0c4cc;
              }
            }
          }
          
          .stage-content {
            .stage-title {
              font-weight: 600;
              color: #303133;
              margin-bottom: 4px;
            }
            
            .stage-desc {
              font-size: 12px;
              color: #606266;
              margin-bottom: 8px;
            }
            
            .stage-time {
              font-size: 12px;
              color: #909399;
            }
          }
          
          .stage-progress {
            margin-top: 8px;
          }
        }
      }
    }
  }
  
  .l3-details {
    .details-card {
      border: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      
      .details-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
        
        h4 {
          margin: 0;
          color: #303133;
        }
      }
      
      .l3-list {
        .l3-item {
          border: 1px solid #e4e7ed;
          border-radius: 8px;
          padding: 16px;
          margin-bottom: 12px;
          background: #fafafa;
          
          .l3-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 12px;
            
            .l3-title {
              display: flex;
              align-items: center;
              font-weight: 600;
              color: #303133;
              
              .el-icon {
                margin-right: 8px;
              }
            }
          }
          
          .l3-progress {
            display: flex;
            align-items: center;
            margin-bottom: 12px;
            
            .el-progress {
              flex: 1;
              margin-right: 12px;
            }
            
            .progress-text {
              font-size: 12px;
              color: #606266;
            }
          }
          
          .l3-results {
            .result-stats {
              display: flex;
              gap: 16px;
              margin-bottom: 12px;
              
              .result-item {
                display: flex;
                align-items: center;
                font-size: 12px;
                color: #606266;
                
                .el-icon {
                  margin-right: 4px;
                }
              }
            }
            
            .l4-details {
              .l4-list {
                .l4-item {
                  border: 1px solid #e4e7ed;
                  border-radius: 4px;
                  padding: 12px;
                  margin-bottom: 8px;
                  background: white;
                  
                  .l4-header {
                    display: flex;
                    justify-content: space-between;
                    align-items: center;
                    margin-bottom: 8px;
                    
                    .l4-title {
                      font-weight: 500;
                      color: #303133;
                    }
                  }
                  
                  .l4-meta {
                    display: flex;
                    gap: 12px;
                    
                    .meta-item {
                      font-size: 12px;
                      color: #606266;
                    }
                  }
                }
              }
            }
          }
          
          .l3-error {
            margin-top: 12px;
          }
        }
      }
    }
  }
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  
  .footer-info {
    display: flex;
    align-items: center;
    color: #606266;
    
    .el-icon {
      margin-right: 4px;
      
      &.rotating {
        animation: rotate 2s linear infinite;
      }
      
      &.success {
        color: #67c23a;
      }
      
      &.error {
        color: #f56c6c;
      }
    }
  }
  
  .footer-actions {
    display: flex;
    gap: 12px;
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>