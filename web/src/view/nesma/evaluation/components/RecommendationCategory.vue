<template>
  <div class="recommendation-category">
    <div class="category-header">
      <h4>{{ title }}</h4>
      <div class="category-summary">
        <el-tag type="danger" size="small">{{ highPriorityCount }}高优先级</el-tag>
        <el-tag type="warning" size="small">{{ mediumPriorityCount }}中优先级</el-tag>
        <el-tag type="info" size="small">{{ lowPriorityCount }}低优先级</el-tag>
      </div>
    </div>
    
    <div class="recommendation-items">
      <div
        v-for="recommendation in recommendations"
        :key="recommendation.id"
        class="recommendation-item"
        :class="recommendation.priority"
      >
        <div class="item-header">
          <div class="item-title">
            <el-icon class="priority-icon">
              <component :is="getPriorityIcon(recommendation.priority)" />
            </el-icon>
            <span>{{ recommendation.title }}</span>
          </div>
          <div class="item-priority">
            <el-tag :type="getPriorityColor(recommendation.priority)" size="small">
              {{ getPriorityLabel(recommendation.priority) }}
            </el-tag>
          </div>
        </div>
        
        <div class="item-content">
          <p class="item-description">{{ recommendation.description }}</p>
          
          <div class="item-metrics">
            <div class="metric">
              <span class="metric-label">影响程度:</span>
              <span class="metric-value">{{ recommendation.impact }}</span>
            </div>
            <div class="metric">
              <span class="metric-label">预计工时:</span>
              <span class="metric-value">{{ recommendation.effort }}小时</span>
            </div>
          </div>
          
          <div class="item-benefits" v-if="recommendation.benefits && recommendation.benefits.length > 0">
            <div class="benefits-label">预期收益:</div>
            <div class="benefits-list">
              <el-tag
                v-for="benefit in recommendation.benefits"
                :key="benefit"
                type="success"
                size="small"
                effect="plain"
                style="margin-right: 8px; margin-bottom: 4px;"
              >
                {{ benefit }}
              </el-tag>
            </div>
          </div>
        </div>
        
        <div class="item-actions">
          <el-button
            type="primary"
            size="small"
            @click="handleActionClick({ type: 'implement', recommendation })"
          >
            实施建议
          </el-button>
          <el-button
            type="default"
            size="small"
            @click="handleActionClick({ type: 'detail', recommendation })"
          >
            查看详情
          </el-button>
          <el-button
            type="default"
            size="small"
            @click="handleActionClick({ type: 'postpone', recommendation })"
          >
            稍后处理
          </el-button>
        </div>
      </div>
    </div>
    
    <div v-if="recommendations.length === 0" class="empty-state">
      <el-empty description="暂无推荐建议" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { ArrowUp, Warning, InfoFilled } from '@element-plus/icons-vue'

// Props
const props = defineProps({
  title: {
    type: String,
    required: true
  },
  recommendations: {
    type: Array,
    default: () => []
  }
})

// Emits
const emit = defineEmits(['action-click'])

// 计算属性
const highPriorityCount = computed(() => {
  return props.recommendations.filter(rec => rec.priority === 'high').length
})

const mediumPriorityCount = computed(() => {
  return props.recommendations.filter(rec => rec.priority === 'medium').length
})

const lowPriorityCount = computed(() => {
  return props.recommendations.filter(rec => rec.priority === 'low').length
})

// 方法
const getPriorityIcon = (priority) => {
  const icons = {
    high: ArrowUp,
    medium: Warning,
    low: InfoFilled
  }
  return icons[priority] || InfoFilled
}

const getPriorityColor = (priority) => {
  const colors = {
    high: 'danger',
    medium: 'warning',
    low: 'info'
  }
  return colors[priority] || 'info'
}

const getPriorityLabel = (priority) => {
  const labels = {
    high: '高优先级',
    medium: '中优先级',
    low: '低优先级'
  }
  return labels[priority] || priority
}

const handleActionClick = (action) => {
  emit('action-click', action)
}
</script>

<style scoped>
.recommendation-category {
  padding: 20px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fff;
}

.category-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 15px;
  border-bottom: 1px solid #e4e7ed;
}

.category-header h4 {
  margin: 0;
  color: #303133;
  font-size: 16px;
}

.category-summary {
  display: flex;
  gap: 8px;
}

.recommendation-items {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.recommendation-item {
  padding: 24px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: #fff;
  transition: all 0.3s;
}

.recommendation-item:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.recommendation-item.high {
  border-left: 4px solid #f56c6c;
  background: linear-gradient(90deg, #fef2f2 0%, #fff 20%);
}

.recommendation-item.medium {
  border-left: 4px solid #e6a23c;
  background: linear-gradient(90deg, #fefce8 0%, #fff 20%);
}

.recommendation-item.low {
  border-left: 4px solid #409eff;
  background: linear-gradient(90deg, #f0f9ff 0%, #fff 20%);
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.item-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #303133;
  font-size: 16px;
}

.priority-icon {
  font-size: 18px;
}

.recommendation-item.high .priority-icon {
  color: #f56c6c;
}

.recommendation-item.medium .priority-icon {
  color: #e6a23c;
}

.recommendation-item.low .priority-icon {
  color: #409eff;
}

.item-content {
  margin-bottom: 20px;
}

.item-description {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 16px;
  font-size: 14px;
}

.item-metrics {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;
}

.metric {
  display: flex;
  align-items: center;
  gap: 4px;
}

.metric-label {
  color: #909399;
  font-size: 13px;
}

.metric-value {
  color: #303133;
  font-weight: 500;
  font-size: 13px;
}

.item-benefits {
  margin-bottom: 0;
}

.benefits-label {
  color: #606266;
  font-size: 13px;
  margin-bottom: 8px;
}

.benefits-list {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.item-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.empty-state {
  text-align: center;
  padding: 40px;
}
</style> 