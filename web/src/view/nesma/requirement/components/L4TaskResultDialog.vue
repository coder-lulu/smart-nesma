<template>
  <el-dialog 
    v-model="visible" 
    title="L4任务生成结果" 
    width="1200px" 
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div class="l4-task-result">
      <!-- 任务基本信息 -->
      <div class="task-info" v-if="taskDetail">
        <el-descriptions :column="3" size="small" border>
          <el-descriptions-item label="任务ID">{{ taskDetail.ID }}</el-descriptions-item>
          <el-descriptions-item label="任务状态">
            <el-tag :type="getStatusColor(taskDetail.status)">
              {{ getStatusText(taskDetail.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="生成时间">{{ formatDateTime(taskDetail.createdAt) }}</el-descriptions-item>
          <el-descriptions-item label="总计数量">{{ taskDetail.totalCount || 0 }}</el-descriptions-item>
          <el-descriptions-item label="成功数量">{{ taskDetail.successCount || 0 }}</el-descriptions-item>
          <el-descriptions-item label="失败数量">{{ taskDetail.failedCount || 0 }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <!-- 加载状态 -->
      <div class="loading-container" v-if="loading">
        <el-skeleton :rows="5" animated />
      </div>

      <!-- 无结果提示 -->
      <div class="no-results" v-else-if="!loading && (!generationResults || generationResults.length === 0)">
        <el-empty description="暂无生成结果" />
      </div>

      <!-- 生成结果展示 -->
      <div class="generation-results" v-else>
        <div class="result-header">
          <h3>生成结果 ({{ generationResults.length }} 个L3需求，共 {{ totalL4Count }} 个L4功能点)</h3>
        </div>

        <el-collapse v-model="activeResultPanels" accordion>
          <el-collapse-item 
            v-for="(result, index) in generationResults" 
            :key="index"
            :title="`${result.l3Requirement.title} (${result.l4Suggestions.length} 个L4建议)`"
            :name="`result-${index}`"
          >
            <div class="l3-result-content">
              <!-- L3需求信息 -->
              <div class="l3-info">
                <el-descriptions :column="2" size="small" border>
                  <el-descriptions-item label="需求编号">{{ result.l3Requirement.code }}</el-descriptions-item>
                  <el-descriptions-item label="生成策略">{{ result.strategy }}</el-descriptions-item>
                  <el-descriptions-item label="生成时间">{{ formatDateTime(result.createdAt) }}</el-descriptions-item>
                  <el-descriptions-item label="建议数量">{{ result.l4Suggestions.length }} 个</el-descriptions-item>
                </el-descriptions>
              </div>

              <!-- L4建议列表 -->
              <div class="l4-suggestions">
                <h4>L4功能点建议</h4>
                <div class="suggestion-grid">
                  <div 
                    v-for="(suggestion, sIndex) in result.l4Suggestions" 
                    :key="sIndex"
                    class="suggestion-card"
                  >
                    <div class="suggestion-header">
                      <div class="suggestion-title">
                        <span>{{ suggestion.suggested_title }}</span>
                        <el-tag size="small" type="info" style="margin-left: 8px;">
                          {{ suggestion.suggested_code }}
                        </el-tag>
                      </div>
                      <div class="suggestion-meta">
                        <el-tag size="small" :type="getFunctionTypeColor(suggestion.function_type)">
                          {{ suggestion.function_type }}
                        </el-tag>
                        <el-tag size="small" :type="getComplexityType(suggestion.estimated_complexity)">
                          {{ suggestion.estimated_complexity }}
                        </el-tag>
                      </div>
                    </div>
                    
                    <div class="suggestion-content">
                      <p class="suggestion-description">{{ suggestion.suggested_description }}</p>
                      
                      <div class="suggestion-details">
                        <el-row :gutter="12">
                          <el-col :span="6">
                            <div class="detail-item">
                              <span class="label">优先级:</span>
                              <span class="value">{{ suggestion.priority }}</span>
                            </div>
                          </el-col>
                          <el-col :span="6">
                            <div class="detail-item">
                              <span class="label">建议AFP:</span>
                              <span class="value">{{ suggestion.recommended_afp }}</span>
                            </div>
                          </el-col>
                          <el-col :span="6">
                            <div class="detail-item">
                              <span class="label">建议UFP:</span>
                              <span class="value">{{ suggestion.recommended_ufp }}</span>
                            </div>
                          </el-col>
                          <el-col :span="6">
                            <div class="detail-item">
                              <span class="label">置信度:</span>
                              <span class="value">{{ (suggestion.confidence * 100).toFixed(1) }}%</span>
                            </div>
                          </el-col>
                        </el-row>
                      </div>

                      <div class="suggestion-business-value" v-if="suggestion.business_value">
                        <el-text size="small" type="success">
                          <el-icon><InfoFilled /></el-icon>
                          <strong>业务价值:</strong> {{ suggestion.business_value }}
                        </el-text>
                      </div>

                      <div class="suggestion-acceptance-criteria" v-if="suggestion.acceptance_criteria">
                        <el-text size="small" type="warning">
                          <el-icon><InfoFilled /></el-icon>
                          <strong>验收标准:</strong> {{ suggestion.acceptance_criteria }}
                        </el-text>
                      </div>

                      <div class="suggestion-related" v-if="suggestion.related_requirements || suggestion.related_knowledge">
                        <el-row :gutter="12">
                          <el-col :span="12" v-if="suggestion.related_requirements">
                            <div class="detail-item">
                              <span class="label">关联需求:</span>
                              <span class="value">{{ suggestion.related_requirements }}</span>
                            </div>
                          </el-col>
                          <el-col :span="12" v-if="suggestion.related_knowledge">
                            <div class="detail-item">
                              <span class="label">相关知识:</span>
                              <span class="value">{{ suggestion.related_knowledge }}</span>
                            </div>
                          </el-col>
                        </el-row>
                      </div>

                      <div class="suggestion-reason" v-if="suggestion.generation_reason">
                        <el-text size="small" type="info">
                          <el-icon><InfoFilled /></el-icon>
                          <strong>生成原因:</strong> {{ suggestion.generation_reason }}
                        </el-text>
                      </div>

                      <!-- 采纳按钮 -->
                      <div class="suggestion-actions">
                        <el-button 
                          type="primary" 
                          size="small" 
                          @click="adoptL4Suggestion(result.l3Requirement, suggestion)"
                          :disabled="isAdopted(result.l3Requirement.id, suggestion)"
                          :loading="adoptingMap[getSuggestionKey(result.l3Requirement.id, suggestion)]"
                        >
                          <el-icon><Check /></el-icon>
                          {{ isAdopted(result.l3Requirement.id, suggestion) ? '已采纳' : '采纳' }}
                        </el-button>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { InfoFilled, Check } from '@element-plus/icons-vue'
import { getL4TaskDetail, getLevel4GenerationTaskResult, createNesmaRequirement } from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  taskId: {
    type: [Number, String],
    default: null
  }
})

const emit = defineEmits(['update:modelValue'])

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const loading = ref(false)
const taskDetail = ref(null)
const generationResults = ref([])
const activeResultPanels = ref([])

// 采纳相关状态
const adoptedL4Requirements = ref(new Set())
const adoptingMap = ref({}) // 用于跟踪每个建议的采纳状态

const totalL4Count = computed(() => {
  return generationResults.value.reduce((total, result) => {
    return total + (result.l4Suggestions ? result.l4Suggestions.length : 0)
  }, 0)
})

const loadTaskResult = async () => {
  if (!props.taskId) return
  
  loading.value = true
  try {
    // 获取任务详情
    const detailResponse = await getL4TaskDetail(props.taskId)
    if (detailResponse.code === 0) {
      taskDetail.value = detailResponse.data
    }

    // 获取任务结果
    const resultResponse = await getLevel4GenerationTaskResult(props.taskId)
    if (resultResponse.code === 0 && resultResponse.data) {
      // 从返回的数据中提取任务信息和结果
      const taskData = resultResponse.data.task
      const resultData = taskData.result
      
      // 更新任务详情（如果API返回的数据更完整）
      if (taskData) {
        taskDetail.value = {
          ...taskDetail.value,
          ...taskData,
          totalCount: taskData.totalCount,
          successCount: taskData.successCount,
          failedCount: taskData.failedCount,
          status: taskData.status,
          createdAt: taskData.CreatedAt,
          updatedAt: taskData.UpdatedAt
        }
      }

      // 处理L4建议结果
      if (resultData && resultData.l4_suggestions_results && Array.isArray(resultData.l4_suggestions_results)) {
        generationResults.value = resultData.l4_suggestions_results.map(result => ({
          l3Requirement: {
            id: result.l3_requirement.ID,
            code: result.l3_requirement.code || 'N/A',
            title: result.l3_requirement.title || 'N/A',
            description: result.l3_requirement.description || ''
          },
          strategy: taskData.config?.generation_strategy || 'comprehensive',
          createdAt: taskData.CreatedAt,
          l4Suggestions: result.l4_suggestions || []
        }))
      } else {
        generationResults.value = []
      }
    } else {
      // 如果没有结果数据，显示空状态
      generationResults.value = []
    }
  } catch (error) {
    console.error('加载L4任务结果失败:', error)
    ElMessage.error('加载任务结果失败：' + error.message)
    generationResults.value = []
  } finally {
    loading.value = false
  }
}

const getFunctionTypeColor = (type) => {
  const colorMap = {
    'EI': 'primary',
    'EO': 'success',
    'EQ': 'warning',
    'ILF': 'danger',
    'EIF': 'info'
  }
  return colorMap[type] || 'info'
}

const getComplexityType = (complexity) => {
  const typeMap = {
    'low': 'success',
    'moderate': 'warning',
    'high': 'danger'
  }
  return typeMap[complexity] || 'info'
}

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
    running: '生成中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return textMap[status] || status
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

const handleClose = () => {
  visible.value = false
}

// 获取建议的唯一标识
const getSuggestionKey = (l3RequirementId, suggestion) => {
  return `${l3RequirementId}_${suggestion.suggested_title}_${suggestion.function_type}`
}

// 检查建议是否已被采纳
const isAdopted = (l3RequirementId, suggestion) => {
  const key = getSuggestionKey(l3RequirementId, suggestion)
  return adoptedL4Requirements.value.has(key)
}

// 采纳单个L4建议
const adoptL4Suggestion = async (l3Requirement, suggestion) => {
  try {
    // 检查是否已经采纳过
    const suggestionKey = getSuggestionKey(l3Requirement.id, suggestion)
    if (adoptedL4Requirements.value.has(suggestionKey)) {
      ElMessage.warning('该功能点已经采纳过了')
      return
    }

    // 设置加载状态
    adoptingMap.value[suggestionKey] = true

    // 构建L4需求数据
    const l4RequirementData = {
      aiAnalysisStatus: 'completed',
      title: suggestion.suggested_title,
      description: suggestion.suggested_description,
      functionType: suggestion.function_type,
      complexity: suggestion.estimated_complexity,
      afp: suggestion.recommended_afp,
      ufp: suggestion.recommended_ufp,
      level: 4,
      parentId: l3Requirement.id,
      projectId: taskDetail.value?.projectId,
      cycleId: taskDetail.value?.cycleId,
      versionId: taskDetail.value?.versionId,
      status: 'completed',
      source: 'ai_generated',
      businessValue: suggestion.business_value,
      acceptanceCriteria: suggestion.acceptance_criteria,
      notes: `AI生成建议，置信度: ${Math.round(suggestion.confidence * 100)}%，优先级: P${suggestion.priority || 1}`,
      priority: suggestion.priority || 1
    }
    
    ElMessage.info('正在采纳L4功能点建议...')
    const response = await createNesmaRequirement(l4RequirementData)
    
    if (response.code === 0) {
      // 标记为已采纳
      adoptedL4Requirements.value.add(suggestionKey)
      
      ElMessage.success('L4功能点采纳成功！')
    } else {
      throw new Error(response.msg || '采纳失败')
    }
  } catch (error) {
    console.error('采纳L4建议失败:', error)
    ElMessage.error('采纳L4建议失败: ' + (error.message || '未知错误'))
  } finally {
    // 清除加载状态
    const suggestionKey = getSuggestionKey(l3Requirement.id, suggestion)
    adoptingMap.value[suggestionKey] = false
  }
}

// 监听对话框显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.taskId) {
    // 重置采纳状态
    adoptedL4Requirements.value.clear()
    adoptingMap.value = {}
    loadTaskResult()
  }
})
</script>

<style lang="scss" scoped>
.l4-task-result {
  .task-info {
    margin-bottom: 20px;
  }

  .loading-container {
    padding: 20px;
  }

  .no-results {
    padding: 40px;
    text-align: center;
  }

  .generation-results {
    .result-header {
      margin-bottom: 20px;
      
      h3 {
        margin: 0;
        color: #303133;
        font-size: 16px;
        font-weight: 600;
      }
    }

    .l3-result-content {
      .l3-info {
        margin-bottom: 20px;
      }

      .l4-suggestions {
        h4 {
          margin: 0 0 15px 0;
          color: #303133;
          font-size: 14px;
          font-weight: 600;
        }

        .suggestion-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
          gap: 16px;
        }

        .suggestion-card {
          border: 1px solid #e4e7ed;
          border-radius: 8px;
          padding: 16px;
          background: #fff;
          transition: all 0.3s ease;
          
          &:hover {
            border-color: #409eff;
            box-shadow: 0 2px 8px rgba(64, 158, 255, 0.2);
          }

          .suggestion-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 12px;

            .suggestion-title {
              flex: 1;
              font-weight: 600;
              color: #303133;
              font-size: 14px;
              line-height: 1.4;
              margin-right: 12px;
            }

            .suggestion-meta {
              display: flex;
              gap: 8px;
              flex-shrink: 0;
            }
          }

          .suggestion-content {
            .suggestion-description {
              color: #606266;
              font-size: 13px;
              line-height: 1.5;
              margin: 0 0 12px 0;
              word-break: break-word;
            }

            .suggestion-details {
              margin-bottom: 12px;
              
              .detail-item {
                display: flex;
                align-items: center;
                margin-bottom: 8px;
                font-size: 12px;
                
                .label {
                  color: #909399;
                  margin-right: 8px;
                  font-weight: 500;
                }
                
                .value {
                  color: #303133;
                  font-weight: 600;
                }
              }
            }

            .suggestion-business-value,
            .suggestion-acceptance-criteria,
            .suggestion-reason {
              padding: 8px 12px;
              border-radius: 4px;
              margin-bottom: 8px;
              
              .el-text {
                font-size: 12px;
                line-height: 1.4;
                
                .el-icon {
                  margin-right: 4px;
                }
              }
            }

            .suggestion-business-value {
              background: #f0f9ff;
              border-left: 3px solid #67c23a;
            }

            .suggestion-acceptance-criteria {
              background: #fef7e0;
              border-left: 3px solid #e6a23c;
            }

            .suggestion-reason {
              background: #f8f9fa;
              border-left: 3px solid #409eff;
            }

            .suggestion-actions {
              margin-top: 12px;
              padding-top: 12px;
              border-top: 1px solid #e4e7ed;
              text-align: right;
              
              .el-button {
                min-width: 80px;
                
                &.is-disabled {
                  background-color: #f0f0f0;
                  border-color: #d3d3d3;
                  color: #999;
                }
              }
            }

            .suggestion-related {
              margin-bottom: 8px;
              
              .detail-item {
                display: flex;
                align-items: center;
                margin-bottom: 6px;
                font-size: 12px;
                
                .label {
                  color: #909399;
                  margin-right: 8px;
                  font-weight: 500;
                  min-width: 60px;
                }
                
                .value {
                  color: #303133;
                  font-weight: 600;
                  word-break: break-all;
                }
              }
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
}
</style>