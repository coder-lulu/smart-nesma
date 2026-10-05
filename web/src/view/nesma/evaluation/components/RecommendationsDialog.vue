<template>
  <el-dialog
    v-model="dialogVisible"
    title="改进建议"
    width="70%"
    :close-on-click-modal="false"
    @opened="loadRecommendations"
  >
    <div v-loading="loading" class="recommendations-content">
      <div v-if="recommendations.length === 0 && !loading" class="no-recommendations">
        <el-empty description="暂无改进建议" />
      </div>
      
      <div v-else class="recommendations-list">
        <!-- 建议统计 -->
        <div class="stats-section">
          <el-row :gutter="16">
            <el-col :span="6">
              <div class="stat-card critical">
                <div class="stat-number">{{ getCountByPriority('critical') }}</div>
                <div class="stat-label">紧急建议</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card high">
                <div class="stat-number">{{ getCountByPriority('high') }}</div>
                <div class="stat-label">高优先级</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card medium">
                <div class="stat-number">{{ getCountByPriority('medium') }}</div>
                <div class="stat-label">中优先级</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card low">
                <div class="stat-number">{{ getCountByPriority('low') }}</div>
                <div class="stat-label">低优先级</div>
              </div>
            </el-col>
          </el-row>
        </div>

        <!-- 建议分类 -->
        <div class="categories-section">
          <el-tabs v-model="activeCategory" @tab-click="handleCategoryChange">
            <el-tab-pane label="全部" name="all">
              <div class="category-count">{{ recommendations.length }} 条建议</div>
            </el-tab-pane>
            <el-tab-pane label="质量改进" name="quality">
              <div class="category-count">{{ getCountByCategory('quality') }} 条建议</div>
            </el-tab-pane>
            <el-tab-pane label="架构完整性" name="architecture">
              <div class="category-count">{{ getCountByCategory('architecture') }} 条建议</div>
            </el-tab-pane>
            <el-tab-pane label="风险管理" name="risk">
              <div class="category-count">{{ getCountByCategory('risk') }} 条建议</div>
            </el-tab-pane>
            <el-tab-pane label="合规性" name="compliance">
              <div class="category-count">{{ getCountByCategory('compliance') }} 条建议</div>
            </el-tab-pane>
            <el-tab-pane label="项目规划" name="planning">
              <div class="category-count">{{ getCountByCategory('planning') }} 条建议</div>
            </el-tab-pane>
          </el-tabs>
        </div>

        <!-- 建议列表 -->
        <div class="recommendations-items">
          <div
            v-for="(recommendation, index) in filteredRecommendations"
            :key="index"
            class="recommendation-item"
            :class="{ 'handled': recommendation.handled }"
          >
            <div class="recommendation-header">
              <div class="left-section">
                <el-tag
                  :type="getPriorityColor(recommendation.priority)"
                  size="small"
                  class="priority-tag"
                >
                  {{ getPriorityLabel(recommendation.priority) }}
                </el-tag>
                <el-tag
                  :type="getCategoryColor(recommendation.category)"
                  size="small"
                  class="category-tag"
                >
                  {{ getCategoryLabel(recommendation.category) }}
                </el-tag>
                <span class="recommendation-title">{{ recommendation.title }}</span>
              </div>
              <div class="right-section">
                <el-button
                  v-if="!recommendation.handled"
                  size="small"
                  type="success"
                  @click="markAsHandled(index)"
                >
                  标记为已处理
                </el-button>
                <el-tag v-else type="success" size="small">已处理</el-tag>
              </div>
            </div>
            
            <div class="recommendation-body">
              <div class="description">
                {{ recommendation.description }}
              </div>
              
              <div class="details" v-if="recommendation.details">
                <div class="detail-item" v-for="detail in recommendation.details" :key="detail">
                  <i class="el-icon-arrow-right"></i>
                  {{ detail }}
                </div>
              </div>
              
              <div class="actions" v-if="recommendation.actions && recommendation.actions.length > 0">
                <div class="actions-title">建议措施：</div>
                <div class="action-item" v-for="action in recommendation.actions" :key="action">
                  <i class="el-icon-check"></i>
                  {{ action }}
                </div>
              </div>
              
              <div class="impact" v-if="recommendation.impact">
                <div class="impact-title">预期影响：</div>
                <div class="impact-content">{{ recommendation.impact }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="exportRecommendations">导出建议</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getRecommendations, markRecommendationAsHandled } from '@/api/nesma'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  evaluationId: {
    type: [String, Number],
    default: null
  }
})

// Emits
const emit = defineEmits(['update:modelValue'])

// 响应式数据
const loading = ref(false)
const recommendations = ref([])
const activeCategory = ref('all')

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const filteredRecommendations = computed(() => {
  if (activeCategory.value === 'all') {
    return recommendations.value
  }
  return recommendations.value.filter(item => item.category === activeCategory.value)
})

// 监听
watch(() => props.evaluationId, (newId) => {
  if (newId && props.modelValue) {
    loadRecommendations()
  }
})

// 方法
const loadRecommendations = async () => {
  if (!props.evaluationId) return
  
  loading.value = true
  try {
    const response = await getRecommendations(props.evaluationId)
    recommendations.value = response.data || []
  } catch (error) {
    console.error('加载改进建议失败:', error)
    ElMessage.error('加载改进建议失败')
  } finally {
    loading.value = false
  }
}

const markAsHandled = async (index) => {
  try {
    await markRecommendationAsHandled(props.evaluationId, index)
    recommendations.value[index].handled = true
    ElMessage.success('已标记为已处理')
  } catch (error) {
    console.error('标记失败:', error)
    ElMessage.error('标记失败')
  }
}

const handleCategoryChange = (tab) => {
  activeCategory.value = tab.name
}

const exportRecommendations = () => {
  const content = recommendations.value.map(item => {
    let text = `${getPriorityLabel(item.priority)} - ${getCategoryLabel(item.category)}\n`
    text += `标题: ${item.title}\n`
    text += `描述: ${item.description}\n`
    
    if (item.details && item.details.length > 0) {
      text += `详情:\n${item.details.map(detail => `  • ${detail}`).join('\n')}\n`
    }
    
    if (item.actions && item.actions.length > 0) {
      text += `建议措施:\n${item.actions.map(action => `  • ${action}`).join('\n')}\n`
    }
    
    if (item.impact) {
      text += `预期影响: ${item.impact}\n`
    }
    
    return text
  }).join('\n' + '='.repeat(50) + '\n')
  
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `评估改进建议_${new Date().toISOString().slice(0, 10)}.txt`
  link.click()
  window.URL.revokeObjectURL(url)
  ElMessage.success('导出成功')
}

// 工具函数
const getCountByPriority = (priority) => {
  return recommendations.value.filter(item => item.priority === priority).length
}

const getCountByCategory = (category) => {
  return recommendations.value.filter(item => item.category === category).length
}

const getPriorityLabel = (priority) => {
  const labels = {
    'critical': '紧急',
    'high': '高',
    'medium': '中',
    'low': '低'
  }
  return labels[priority] || priority
}

const getPriorityColor = (priority) => {
  const colors = {
    'critical': 'danger',
    'high': 'warning',
    'medium': 'primary',
    'low': 'info'
  }
  return colors[priority] || ''
}

const getCategoryLabel = (category) => {
  const labels = {
    'quality': '质量改进',
    'architecture': '架构完整性',
    'risk': '风险管理',
    'compliance': '合规性',
    'planning': '项目规划'
  }
  return labels[category] || category
}

const getCategoryColor = (category) => {
  const colors = {
    'quality': 'success',
    'architecture': 'primary',
    'risk': 'warning',
    'compliance': 'info',
    'planning': 'danger'
  }
  return colors[category] || ''
}
</script>

<style scoped>
.recommendations-content {
  min-height: 300px;
}

.no-recommendations {
  text-align: center;
  padding: 40px 0;
}

.stats-section {
  margin-bottom: 24px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  color: white;
}

.stat-card.critical {
  background: linear-gradient(135deg, #f56c6c 0%, #e53e3e 100%);
}

.stat-card.high {
  background: linear-gradient(135deg, #e6a23c 0%, #d69e2e 100%);
}

.stat-card.medium {
  background: linear-gradient(135deg, #409eff 0%, #3182ce 100%);
}

.stat-card.low {
  background: linear-gradient(135deg, #909399 0%, #718096 100%);
}

.stat-number {
  font-size: 32px;
  font-weight: bold;
  line-height: 1;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 14px;
  opacity: 0.9;
}

.categories-section {
  margin-bottom: 24px;
}

.category-count {
  font-size: 12px;
  color: #666;
  margin-left: 8px;
}

.recommendations-items {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.recommendation-item {
  border: 1px solid #e6e6e6;
  border-radius: 8px;
  padding: 16px;
  background: white;
  transition: all 0.2s;
}

.recommendation-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.recommendation-item.handled {
  background: #f8f9fa;
  border-color: #d6d6d6;
  opacity: 0.8;
}

.recommendation-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.left-section {
  display: flex;
  align-items: center;
  gap: 8px;
}

.priority-tag, .category-tag {
  font-size: 11px;
}

.recommendation-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.recommendation-body {
  color: #666;
  line-height: 1.6;
}

.description {
  margin-bottom: 12px;
  font-size: 14px;
}

.details {
  margin-bottom: 12px;
}

.detail-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
}

.actions {
  margin-bottom: 12px;
}

.actions-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.action-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 4px;
  font-size: 13px;
  color: #409eff;
}

.impact {
  background: #f0f9ff;
  padding: 12px;
  border-radius: 4px;
  border-left: 4px solid #409eff;
}

.impact-title {
  font-weight: 600;
  color: #303133;
  margin-bottom: 4px;
}

.impact-content {
  font-size: 13px;
  color: #666;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-tabs__content) {
  padding-top: 16px;
}

:deep(.el-tabs__item) {
  padding: 0 20px;
}
</style> 