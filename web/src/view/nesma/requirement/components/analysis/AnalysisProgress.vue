<template>
  <el-dialog
    :model-value="modelValue"
    title="AI分析进度"
    width="600px"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="analysis-progress">
      <div v-if="tasks.length === 0" class="no-tasks">
        <el-empty description="暂无分析任务" />
      </div>
      
      <div v-else class="task-list">
        <div v-for="taskId in tasks" :key="taskId" class="task-item">
          <div class="task-header">
            <span class="task-id">任务 {{ taskId }}</span>
            <el-tag :type="getStatusType(getTaskProgress(taskId).status)">
              {{ getStatusLabel(getTaskProgress(taskId).status) }}
            </el-tag>
          </div>
          
          <el-progress
            :percentage="getTaskProgress(taskId).progress"
            :status="getProgressStatus(getTaskProgress(taskId).status)"
            :stroke-width="15"
          />
          
          <div class="task-info">
            <span>进度: {{ getTaskProgress(taskId).progress }}%</span>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">关闭</el-button>
      <el-button v-if="hasCompletedTasks" type="primary" @click="handleRefresh">
        刷新数据
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed } from 'vue'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  tasks: {
    type: Array,
    default: () => []
  },
  progress: {
    type: Object,
    default: () => ({})
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'analysis-completed'])

// 计算属性
const hasCompletedTasks = computed(() => {
  return props.tasks.some(taskId => {
    const taskProgress = getTaskProgress(taskId)
    return taskProgress.status === 'completed'
  })
})

// 方法
const getTaskProgress = (taskId) => {
  return props.progress[taskId] || { progress: 0, status: 'pending' }
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'running': 'warning',
    'completed': 'success',
    'failed': 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    'pending': '待开始',
    'running': '进行中',
    'completed': '已完成',
    'failed': '失败'
  }
  return labelMap[status] || status
}

const getProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const handleRefresh = () => {
  emit('analysis-completed')
  emit('update:modelValue', false)
}
</script>

<style lang="scss" scoped>
.analysis-progress {
  .task-list {
    .task-item {
      margin-bottom: 20px;
      padding: 16px;
      background: #f8fafc;
      border-radius: 8px;
      
      .task-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;
        
        .task-id {
          font-weight: 500;
          color: #303133;
        }
      }
      
      .task-info {
        margin-top: 8px;
        font-size: 12px;
        color: #909399;
      }
    }
  }
}
</style>