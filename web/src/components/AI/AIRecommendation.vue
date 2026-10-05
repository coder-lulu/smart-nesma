<template>
  <div class="ai-recommendation-container">
    <!-- 推荐触发区域 -->
    <div class="recommendation-trigger">
      <el-card class="trigger-card">
        <div class="trigger-header">
          <el-icon><Bulb /></el-icon>
          <span class="trigger-title">AI智能推荐</span>
        </div>
        <p class="trigger-description">
          基于知识库和AI大模型为您提供个性化的需求分析建议
        </p>
        <el-button 
          type="primary" 
          size="large"
          @click="getAIRecommendations"
          :loading="loading"
          :disabled="!canGetRecommendations"
        >
          <el-icon><Magic /></el-icon>
          获取AI推荐
        </el-button>
      </el-card>
    </div>

    <!-- 推荐结果展示 -->
    <div v-if="recommendations.length > 0" class="recommendations-panel">
      <div class="panel-header">
        <h3>AI推荐结果</h3>
        <div class="panel-meta">
          <el-tag type="success" size="small">
            置信度: {{ averageConfidence }}%
          </el-tag>
          <el-tag type="info" size="small">
            知识库引用: {{ knowledgeUsedCount }}条
          </el-tag>
          <el-tag type="warning" size="small">
            处理时间: {{ processingTime }}ms
          </el-tag>
        </div>
      </div>

      <div class="recommendations-list">
        <el-card 
          v-for="(item, index) in recommendations" 
          :key="item.id"
          class="recommendation-item"
          shadow="hover"
        >
          <div class="item-header">
            <div class="item-title">
              <el-icon><Document /></el-icon>
              <h4>{{ item.title }}</h4>
            </div>
            <div class="item-actions">
              <el-button-group>
                <el-button 
                  size="small" 
                  type="success"
                  @click="adoptRecommendation(item)"
                  :loading="adoptingIds.includes(item.id)"
                >
                  <el-icon><Check /></el-icon>
                  采纳
                </el-button>
                <el-button 
                  size="small" 
                  type="danger"
                  @click="rejectRecommendation(item)"
                  :loading="rejectingIds.includes(item.id)"
                >
                  <el-icon><Close /></el-icon>
                  拒绝
                </el-button>
              </el-button-group>
            </div>
          </div>

          <div class="item-content">
            <div class="content-section">
              <h5>推荐内容</h5>
              <p class="content-text">{{ item.content }}</p>
            </div>

            <div v-if="item.improvements && item.improvements.length > 0" class="content-section">
              <h5>改进建议</h5>
              <ul class="improvement-list">
                <li v-for="improvement in item.improvements" :key="improvement">
                  {{ improvement }}
                </li>
              </ul>
            </div>

            <div v-if="item.bestPractices && item.bestPractices.length > 0" class="content-section">
              <h5>最佳实践</h5>
              <ul class="practice-list">
                <li v-for="practice in item.bestPractices" :key="practice">
                  {{ practice }}
                </li>
              </ul>
            </div>

            <div v-if="item.riskWarnings && item.riskWarnings.length > 0" class="content-section">
              <h5>风险提示</h5>
              <ul class="warning-list">
                <li v-for="warning in item.riskWarnings" :key="warning">
                  <el-icon class="warning-icon"><Warning /></el-icon>
                  {{ warning }}
                </li>
              </ul>
            </div>

            <div class="content-section">
              <h5>知识库引用</h5>
              <div class="knowledge-refs">
                <el-tag 
                  v-for="ref in item.knowledgeRefs" 
                  :key="ref.id"
                  size="small"
                  type="info"
                  class="knowledge-tag"
                >
                  {{ ref.title }} ({{ ref.category }})
                </el-tag>
              </div>
            </div>

            <div class="item-footer">
              <div class="confidence-info">
                <span class="confidence-label">置信度: </span>
                <el-rate 
                  v-model="item.displayConfidence" 
                  disabled 
                  show-score
                  text-color="#ff9900"
                  :score-template="`${item.confidence.toFixed(2)}`"
                />
              </div>
              
              <div class="function-type" v-if="item.functionType">
                <el-tag type="primary" size="small">
                  功能类型: {{ item.functionType }}
                </el-tag>
              </div>
            </div>
          </div>
        </el-card>
      </div>
    </div>

    <!-- 采纳反馈对话框 -->
    <el-dialog 
      v-model="showFeedbackDialog" 
      title="采纳反馈" 
      width="500px"
      :before-close="handleDialogClose"
    >
      <div class="feedback-form">
        <p class="feedback-prompt">
          您即将采纳推荐"{{ currentRecommendation?.title }}"，请提供反馈意见：
        </p>
        <el-input
          v-model="feedbackText"
          type="textarea"
          :rows="4"
          placeholder="请输入您的反馈意见（可选）"
          maxlength="500"
          show-word-limit
        />
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showFeedbackDialog = false">取消</el-button>
          <el-button 
            type="primary" 
            @click="confirmAdopt"
            :loading="confirmingAdopt"
          >
            确认采纳
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 拒绝理由对话框 -->
    <el-dialog 
      v-model="showRejectDialog" 
      title="拒绝原因" 
      width="500px"
      :before-close="handleDialogClose"
    >
      <div class="reject-form">
        <p class="reject-prompt">
          为了改进推荐算法，请告诉我们拒绝的原因：
        </p>
        <el-radio-group v-model="rejectReason">
          <el-radio value="not_relevant">推荐内容不相关</el-radio>
          <el-radio value="low_quality">推荐质量较低</el-radio>
          <el-radio value="already_known">内容已知</el-radio>
          <el-radio value="not_applicable">不适用当前场景</el-radio>
          <el-radio value="other">其他原因</el-radio>
        </el-radio-group>
        
        <el-input
          v-if="rejectReason === 'other'"
          v-model="customRejectReason"
          type="textarea"
          :rows="3"
          placeholder="请详细说明原因"
          maxlength="200"
          show-word-limit
          class="custom-reason-input"
        />
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="showRejectDialog = false">取消</el-button>
          <el-button 
            type="danger" 
            @click="confirmReject"
            :loading="confirmingReject"
          >
            确认拒绝
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Bulb, 
  Magic, 
  Document, 
  Check, 
  Close, 
  Warning 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  projectId: {
    type: Number,
    required: true
  },
  projectName: {
    type: String,
    default: ''
  },
  domain: {
    type: String,
    default: ''
  },
  requirementText: {
    type: String,
    default: ''
  }
})

// Reactive data
const loading = ref(false)
const recommendations = ref([])
const adoptingIds = ref([])
const rejectingIds = ref([])
const averageConfidence = ref(0)
const knowledgeUsedCount = ref(0)
const processingTime = ref(0)

// Dialog states
const showFeedbackDialog = ref(false)
const showRejectDialog = ref(false)
const currentRecommendation = ref(null)
const feedbackText = ref('')
const rejectReason = ref('')
const customRejectReason = ref('')
const confirmingAdopt = ref(false)
const confirmingReject = ref(false)

// Computed
const canGetRecommendations = computed(() => {
  return props.projectId > 0 && !loading.value
})

// Methods
const getAIRecommendations = async () => {
  loading.value = true
  try {
    const response = await fetch('/api/v1/nesma/recommendation/ai', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        projectId: props.projectId,
        projectName: props.projectName,
        domain: props.domain,
        requirementText: props.requirementText,
        context: {
          timestamp: new Date().toISOString()
        }
      })
    })
    
    if (!response.ok) {
      throw new Error('获取推荐失败')
    }
    
    const data = await response.json()
    if (data.code === 0) {
      recommendations.value = data.data.recommendations.map(item => ({
        ...item,
        displayConfidence: item.confidence * 5 // 转换为5星评分
      }))
      averageConfidence.value = Math.round(data.data.confidence * 100)
      knowledgeUsedCount.value = data.data.knowledgeUsed
      processingTime.value = Math.round(data.data.processingTime / 1000000) // 转换为毫秒
      
      ElMessage.success('AI推荐获取成功')
    } else {
      throw new Error(data.msg || '获取推荐失败')
    }
  } catch (error) {
    console.error('获取AI推荐失败:', error)
    ElMessage.error('获取AI推荐失败: ' + error.message)
  } finally {
    loading.value = false
  }
}

const adoptRecommendation = (item) => {
  currentRecommendation.value = item
  showFeedbackDialog.value = true
}

const confirmAdopt = async () => {
  if (!currentRecommendation.value) return
  
  confirmingAdopt.value = true
  adoptingIds.value.push(currentRecommendation.value.id)
  
  try {
    const response = await fetch('/api/v1/nesma/recommendation/adopt', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        projectId: props.projectId,
        recommendationId: currentRecommendation.value.id,
        title: currentRecommendation.value.title,
        content: currentRecommendation.value.content,
        category: currentRecommendation.value.category,
        confidence: currentRecommendation.value.confidence,
        feedback: feedbackText.value
      })
    })
    
    if (!response.ok) {
      throw new Error('采纳推荐失败')
    }
    
    const data = await response.json()
    if (data.code === 0) {
      ElMessage.success('推荐采纳成功，已加入知识库')
      showFeedbackDialog.value = false
      feedbackText.value = ''
      
      // 从推荐列表中移除已采纳的项目
      const index = recommendations.value.findIndex(r => r.id === currentRecommendation.value.id)
      if (index > -1) {
        recommendations.value.splice(index, 1)
      }
    } else {
      throw new Error(data.msg || '采纳推荐失败')
    }
  } catch (error) {
    console.error('采纳推荐失败:', error)
    ElMessage.error('采纳推荐失败: ' + error.message)
  } finally {
    confirmingAdopt.value = false
    adoptingIds.value = adoptingIds.value.filter(id => id !== currentRecommendation.value.id)
    currentRecommendation.value = null
  }
}

const rejectRecommendation = (item) => {
  currentRecommendation.value = item
  showRejectDialog.value = true
}

const confirmReject = async () => {
  if (!currentRecommendation.value) return
  
  confirmingReject.value = true
  rejectingIds.value.push(currentRecommendation.value.id)
  
  try {
    const reason = rejectReason.value === 'other' ? customRejectReason.value : rejectReason.value
    
    const response = await fetch('/api/v1/nesma/recommendation/reject', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${localStorage.getItem('token')}`
      },
      body: JSON.stringify({
        projectId: props.projectId,
        recommendationId: currentRecommendation.value.id,
        reason: reason
      })
    })
    
    if (!response.ok) {
      throw new Error('拒绝推荐失败')
    }
    
    const data = await response.json()
    if (data.code === 0) {
      ElMessage.success('已记录您的反馈，感谢您的建议')
      showRejectDialog.value = false
      rejectReason.value = ''
      customRejectReason.value = ''
      
      // 从推荐列表中移除已拒绝的项目
      const index = recommendations.value.findIndex(r => r.id === currentRecommendation.value.id)
      if (index > -1) {
        recommendations.value.splice(index, 1)
      }
    } else {
      throw new Error(data.msg || '拒绝推荐失败')
    }
  } catch (error) {
    console.error('拒绝推荐失败:', error)
    ElMessage.error('拒绝推荐失败: ' + error.message)
  } finally {
    confirmingReject.value = false
    rejectingIds.value = rejectingIds.value.filter(id => id !== currentRecommendation.value.id)
    currentRecommendation.value = null
  }
}

const handleDialogClose = () => {
  currentRecommendation.value = null
  feedbackText.value = ''
  rejectReason.value = ''
  customRejectReason.value = ''
}

// Lifecycle
onMounted(() => {
  // 可以在这里添加一些初始化逻辑
})
</script>

<style scoped>
.ai-recommendation-container {
  max-width: 1200px;
  margin: 0 auto;
  padding: 20px;
}

.trigger-card {
  text-align: center;
  margin-bottom: 30px;
}

.trigger-header {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  margin-bottom: 15px;
}

.trigger-title {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.trigger-description {
  color: #666;
  margin-bottom: 20px;
  font-size: 16px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
  padding: 15px 20px;
  background: #f5f7fa;
  border-radius: 8px;
}

.panel-header h3 {
  margin: 0;
  color: #303133;
}

.panel-meta {
  display: flex;
  gap: 10px;
}

.recommendations-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.recommendation-item {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-bottom: 15px;
  border-bottom: 1px solid #f0f2f5;
  margin-bottom: 15px;
}

.item-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.item-title h4 {
  margin: 0;
  color: #303133;
  font-size: 18px;
}

.item-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.content-section h5 {
  margin: 0 0 10px 0;
  color: #409eff;
  font-size: 14px;
  font-weight: 600;
}

.content-text {
  color: #606266;
  line-height: 1.6;
  margin: 0;
}

.improvement-list,
.practice-list,
.warning-list {
  margin: 0;
  padding-left: 20px;
}

.improvement-list li,
.practice-list li {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 5px;
}

.warning-list li {
  color: #e6a23c;
  line-height: 1.6;
  margin-bottom: 5px;
  display: flex;
  align-items: center;
  gap: 5px;
}

.warning-icon {
  color: #e6a23c;
}

.knowledge-refs {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.knowledge-tag {
  cursor: pointer;
}

.item-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 15px;
  border-top: 1px solid #f0f2f5;
}

.confidence-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.confidence-label {
  font-size: 14px;
  color: #909399;
}

.feedback-form,
.reject-form {
  margin-bottom: 20px;
}

.feedback-prompt,
.reject-prompt {
  color: #606266;
  margin-bottom: 15px;
}

.custom-reason-input {
  margin-top: 15px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 768px) {
  .ai-recommendation-container {
    padding: 10px;
  }
  
  .item-header {
    flex-direction: column;
    gap: 15px;
    align-items: flex-start;
  }
  
  .item-footer {
    flex-direction: column;
    gap: 15px;
    align-items: flex-start;
  }
  
  .panel-header {
    flex-direction: column;
    gap: 15px;
    align-items: flex-start;
  }
}
</style>