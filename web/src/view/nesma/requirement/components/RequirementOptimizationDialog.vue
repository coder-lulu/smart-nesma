<template>
  <el-dialog
    v-model="visible"
    title="需求智能优化"
    width="800px"
    :close-on-click-modal="false"
  >
    <div class="requirement-optimization">
      <!-- 优化前后对比 -->
      <div class="comparison-section">
        <h3>需求优化对比</h3>
        <el-row :gutter="16">
          <!-- 原始需求 -->
          <el-col :span="12">
            <div class="requirement-panel original">
              <h4>原始需求</h4>
              <div class="requirement-content">
                <div class="field-group">
                  <label>需求标题：</label>
                  <p>{{ originalRequirement.title }}</p>
                </div>
                <div class="field-group">
                  <label>需求描述：</label>
                  <p>{{ originalRequirement.description }}</p>
                </div>
                <div class="field-group">
                  <label>功能类型：</label>
                  <el-tag>{{ originalRequirement.function_type || '未分类' }}</el-tag>
                </div>
                <div class="field-group">
                  <label>复杂度：</label>
                  <el-tag :type="getComplexityTagType(originalRequirement.complexity)">
                    {{ originalRequirement.complexity || '未评估' }}
                  </el-tag>
                </div>
              </div>
            </div>
          </el-col>

          <!-- 优化后需求 -->
          <el-col :span="12">
            <div class="requirement-panel optimized">
              <h4>AI优化建议</h4>
              <div v-if="optimizationResult" class="requirement-content">
                <div class="field-group">
                  <label>优化标题：</label>
                  <p class="optimized-text">{{ optimizationResult.optimized_title || originalRequirement.title }}</p>
                </div>
                <div class="field-group">
                  <label>优化描述：</label>
                  <p class="optimized-text">{{ optimizationResult.optimized_description }}</p>
                </div>
                <div class="field-group">
                  <label>建议功能类型：</label>
                  <el-tag type="success">{{ optimizationResult.suggested_function_type }}</el-tag>
                </div>
                <div class="field-group">
                  <label>建议复杂度：</label>
                  <el-tag :type="getComplexityTagType(optimizationResult.suggested_complexity)" effect="dark">
                    {{ optimizationResult.suggested_complexity }}
                  </el-tag>
                </div>
              </div>
              <div v-else-if="optimizationStatus === 'loading'" class="loading-content">
                <el-loading-container>
                  <p>AI正在分析需求，请稍候...</p>
                </el-loading-container>
              </div>
              <div v-else-if="optimizationStatus === 'error'" class="error-content">
                <el-result
                  icon="error"
                  title="优化失败"
                  :sub-title="errorMessage"
                >
                  <template #extra>
                    <el-button type="primary" @click="retryOptimization">重新优化</el-button>
                  </template>
                </el-result>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 优化建议详情 -->
      <div v-if="optimizationResult" class="optimization-details">
        <h3>优化分析</h3>
        <el-tabs v-model="activeTab" type="card">
          <!-- 改进点 -->
          <el-tab-pane label="改进点" name="improvements">
            <div class="improvements-list">
              <div 
                v-for="(improvement, index) in optimizationResult.improvements" 
                :key="index"
                class="improvement-item"
              >
                <div class="improvement-header">
                  <el-tag :type="getImprovementTagType(improvement.type)">
                    {{ improvement.type || '优化建议' }}
                  </el-tag>
                  <span class="improvement-title">{{ improvement.title }}</span>
                </div>
                <p class="improvement-description">{{ improvement.description }}</p>
              </div>
            </div>
          </el-tab-pane>

          <!-- NESMA分析 -->
          <el-tab-pane label="NESMA分析" name="nesma">
            <div class="nesma-analysis">
              <div class="analysis-item">
                <h4>功能类型分析</h4>
                <p><strong>原分类：</strong>{{ originalRequirement.function_type || '未分类' }}</p>
                <p><strong>建议分类：</strong>{{ optimizationResult.suggested_function_type }}</p>
                <p><strong>分析依据：</strong>{{ optimizationResult.function_type_reasoning }}</p>
              </div>
              <div class="analysis-item">
                <h4>复杂度评估</h4>
                <p><strong>原复杂度：</strong>{{ originalRequirement.complexity || '未评估' }}</p>
                <p><strong>建议复杂度：</strong>{{ optimizationResult.suggested_complexity }}</p>
                <p><strong>评估依据：</strong>{{ optimizationResult.complexity_reasoning }}</p>
              </div>
              <div class="analysis-item">
                <h4>功能点估算</h4>
                <p><strong>估算功能点：</strong>{{ optimizationResult.estimated_function_points || 'N/A' }}</p>
                <p><strong>置信度：</strong>{{ optimizationResult.confidence_score || 'N/A' }}%</p>
              </div>
            </div>
          </el-tab-pane>

          <!-- 知识库引用 -->
          <el-tab-pane label="知识引用" name="knowledge">
            <div class="knowledge-references">
              <div 
                v-for="ref in optimizationResult.knowledge_references" 
                :key="ref.id"
                class="knowledge-item"
              >
                <h5>{{ ref.title }}</h5>
                <p>{{ ref.summary }}</p>
                <div class="knowledge-meta">
                  <el-tag size="small">相关度: {{ ref.relevance }}%</el-tag>
                  <el-tag size="small" type="info">{{ ref.category }}</el-tag>
                </div>
              </div>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <!-- 操作按钮 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button 
          v-if="!optimizationResult && optimizationStatus !== 'loading'" 
          type="primary" 
          @click="startOptimization"
        >
          开始优化
        </el-button>
        <el-button 
          v-if="optimizationResult" 
          type="success" 
          @click="applyOptimization"
        >
          应用优化
        </el-button>
        <el-button 
          v-if="optimizationResult" 
          type="warning" 
          @click="saveAsDraft"
        >
          保存为草稿
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  executeRequirementOptimization,
  getUnifiedAnalysisProgress,
  getDetailedAnalysisResult,
  applyOptimization as applyOptimizationAPI
} from '@/api/nesma'

const props = defineProps({
  modelValue: Boolean,
  requirement: Object
})

const emit = defineEmits(['update:modelValue', 'optimization-applied', 'close'])

// 响应式数据
const optimizationStatus = ref('idle') // idle, loading, completed, error
const optimizationResult = ref(null)
const errorMessage = ref('')
const activeTab = ref('improvements')
const currentTaskId = ref(null)

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const originalRequirement = computed(() => props.requirement || {})

// 监听需求变化
watch(() => props.requirement, (newRequirement) => {
  if (newRequirement) {
    // 重置状态
    optimizationStatus.value = 'idle'
    optimizationResult.value = null
    errorMessage.value = ''
    currentTaskId.value = null
  }
}, { immediate: true })

// 方法
const startOptimization = async () => {
  try {
    optimizationStatus.value = 'loading'
    
    const requestData = {
      requirement_id: originalRequirement.value.ID,
      parameters: {
        optimization_type: 'comprehensive',
        include_nesma_analysis: true,
        include_knowledge_search: true,
        quality_threshold: 0.8
      }
    }

    const response = await executeRequirementOptimization(requestData)
    currentTaskId.value = response.data.task_id || response.data.taskId

    // 轮询获取结果
    pollOptimizationResult()
  } catch (error) {
    console.error('启动优化失败:', error)
    optimizationStatus.value = 'error'
    errorMessage.value = error.response?.data?.message || error.message
  }
}

const pollOptimizationResult = async () => {
  if (!currentTaskId.value) return

  try {
    const progressResponse = await getUnifiedAnalysisProgress(currentTaskId.value)
    const progress = progressResponse.data

    if (progress.status === 'completed') {
      // 获取详细结果
      const resultResponse = await getDetailedAnalysisResult(currentTaskId.value, 'detailed')
      optimizationResult.value = resultResponse.data.result
      optimizationStatus.value = 'completed'
      
      console.log('优化结果:', optimizationResult.value)
    } else if (progress.status === 'failed') {
      optimizationStatus.value = 'error'
      errorMessage.value = progress.error || '优化过程中发生错误'
    } else {
      // 继续轮询
      setTimeout(pollOptimizationResult, 2000)
    }
  } catch (error) {
    console.error('获取优化结果失败:', error)
    optimizationStatus.value = 'error'
    errorMessage.value = '获取优化结果失败'
  }
}

const retryOptimization = () => {
  startOptimization()
}

const applyOptimization = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要应用这些优化建议吗？这将更新原始需求。',
      '应用优化确认',
      {
        confirmButtonText: '确定应用',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const applyData = {
      task_id: currentTaskId.value,
      optimization_id: optimizationResult.value.id,
      suggestion_id: optimizationResult.value.suggestionId,
      action: 'accept'
    }

    await applyOptimizationAPI(applyData)
    
    ElMessage.success('优化建议已应用')
    emit('optimization-applied', {
      requirement: originalRequirement.value,
      optimization: optimizationResult.value
    })
    handleClose()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('应用优化失败: ' + (error.response?.data?.message || error.message))
    }
  }
}

const saveAsDraft = () => {
  ElMessage.info('保存为草稿功能开发中...')
}

const handleClose = () => {
  emit('close')
}

const getComplexityTagType = (complexity) => {
  switch (complexity?.toLowerCase()) {
    case 'low': case '低': return 'success'
    case 'medium': case '中': return 'warning'
    case 'high': case '高': return 'danger'
    default: return 'info'
  }
}

const getImprovementTagType = (type) => {
  switch (type) {
    case '功能优化': return 'primary'
    case '描述改进': return 'success'
    case '分类调整': return 'warning'
    case '复杂度评估': return 'info'
    default: return ''
  }
}
</script>

<style lang="scss" scoped>
.requirement-optimization {
  .comparison-section {
    margin-bottom: 24px;
    
    h3 {
      margin: 0 0 16px 0;
      color: #303133;
      font-size: 16px;
    }
    
    .requirement-panel {
      border: 2px solid #ebeef5;
      border-radius: 8px;
      padding: 16px;
      height: 350px;
      overflow-y: auto;
      
      &.original {
        border-color: #409EFF;
        background: #f0f9ff;
      }
      
      &.optimized {
        border-color: #67C23A;
        background: #f0f9ff;
      }
      
      h4 {
        margin: 0 0 12px 0;
        color: #303133;
        font-size: 14px;
        font-weight: 600;
      }
      
      .requirement-content {
        .field-group {
          margin-bottom: 12px;
          
          label {
            display: block;
            font-size: 12px;
            color: #909399;
            margin-bottom: 4px;
            font-weight: 500;
          }
          
          p {
            margin: 0;
            color: #606266;
            line-height: 1.4;
            font-size: 13px;
            
            &.optimized-text {
              color: #67C23A;
              font-weight: 500;
            }
          }

        }
      }
      
      .loading-content {
        text-align: center;
        padding: 40px 0;
        color: #909399;
      }
      
      .error-content {
        padding: 20px 0;
      }
    }
  }
  
  .optimization-details {
    h3 {
      margin: 0 0 16px 0;
      color: #303133;
      font-size: 16px;
    }
    
    .improvements-list {
      .improvement-item {
        border: 1px solid #ebeef5;
        border-radius: 6px;
        padding: 12px;
        margin-bottom: 12px;
        
        .improvement-header {
          display: flex;
          align-items: center;
          margin-bottom: 8px;
          
          .improvement-title {
            margin-left: 8px;
            font-weight: 500;
            color: #303133;
            font-size: 14px;
          }
        }
        
        .improvement-description {
          margin: 0;
          color: #606266;
          line-height: 1.5;
          font-size: 13px;
        }
      }
    }
    
    .nesma-analysis {
      .analysis-item {
        background: #f8f9fa;
        border: 1px solid #ebeef5;
        border-radius: 6px;
        padding: 16px;
        margin-bottom: 16px;
        
        h4 {
          margin: 0 0 12px 0;
          color: #409EFF;
          font-size: 14px;
        }
        
        p {
          margin: 6px 0;
          color: #606266;
          font-size: 13px;
          line-height: 1.4;
          
          strong {
            color: #303133;
          }
        }
      }
    }
    
    .knowledge-references {
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
        
        .knowledge-meta {
          display: flex;
          gap: 8px;
          margin-top: 8px;
        }
      }
    }
  }
}

.dialog-footer {
  text-align: right;
}
</style>