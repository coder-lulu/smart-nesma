<template>
  <div class="validation-category">
    <div class="category-header">
      <h4>{{ title }}</h4>
      <div class="category-stats">
        <el-tag type="success" size="small">{{ passCount }}通过</el-tag>
        <el-tag type="warning" size="small">{{ warningCount }}警告</el-tag>
        <el-tag type="danger" size="small">{{ failCount }}失败</el-tag>
      </div>
    </div>
    
    <div class="validation-items">
      <div
        v-for="item in items"
        :key="item.id"
        class="validation-item"
        :class="item.status"
        @click="handleItemClick(item)"
      >
        <div class="item-header">
          <div class="item-title">
            <el-icon class="item-icon">
              <component :is="getStatusIcon(item.status)" />
            </el-icon>
            <span>{{ item.name }}</span>
          </div>
          <div class="item-score">{{ item.score }}%</div>
        </div>
        
        <div class="item-content">
          <div class="item-rule">{{ item.rule }}</div>
          <div class="item-result">{{ item.result }}</div>
        </div>
        
        <div class="item-footer" v-if="item.issues && item.issues.length > 0">
          <el-tag type="danger" size="small">{{ item.issues.length }}个问题</el-tag>
          <el-tag type="info" size="small">{{ item.recommendations?.length || 0 }}个建议</el-tag>
        </div>
      </div>
    </div>
    
    <div v-if="items.length === 0" class="empty-state">
      <el-empty description="暂无验证项目" />
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { Check, Warning, Close } from '@element-plus/icons-vue'

// Props
const props = defineProps({
  title: {
    type: String,
    required: true
  },
  items: {
    type: Array,
    default: () => []
  }
})

// Emits
const emit = defineEmits(['item-click'])

// 计算属性
const passCount = computed(() => {
  return props.items.filter(item => item.status === 'pass').length
})

const warningCount = computed(() => {
  return props.items.filter(item => item.status === 'warning').length
})

const failCount = computed(() => {
  return props.items.filter(item => item.status === 'fail').length
})

// 方法
const getStatusIcon = (status) => {
  const icons = {
    pass: Check,
    warning: Warning,
    fail: Close
  }
  return icons[status] || Check
}

const handleItemClick = (item) => {
  emit('item-click', item)
}
</script>

<style scoped>
.validation-category {
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

.category-stats {
  display: flex;
  gap: 8px;
}

.validation-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.validation-item {
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: #fff;
  cursor: pointer;
  transition: all 0.3s;
}

.validation-item:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.validation-item.pass {
  border-left: 4px solid #67c23a;
}

.validation-item.warning {
  border-left: 4px solid #e6a23c;
}

.validation-item.fail {
  border-left: 4px solid #f56c6c;
}

.item-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.item-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  color: #303133;
}

.item-icon {
  font-size: 16px;
}

.validation-item.pass .item-icon {
  color: #67c23a;
}

.validation-item.warning .item-icon {
  color: #e6a23c;
}

.validation-item.fail .item-icon {
  color: #f56c6c;
}

.item-score {
  font-size: 16px;
  font-weight: bold;
  color: #409eff;
}

.item-content {
  margin-bottom: 15px;
}

.item-rule {
  font-size: 14px;
  color: #606266;
  margin-bottom: 8px;
  line-height: 1.5;
}

.item-result {
  font-size: 14px;
  color: #909399;
  line-height: 1.5;
}

.item-footer {
  display: flex;
  gap: 8px;
}

.empty-state {
  text-align: center;
  padding: 40px;
}
</style> 