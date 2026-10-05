<template>
  <el-dialog
    :model-value="modelValue"
    title="版本管理"
    width="900px"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="version-management">
      <!-- 版本列表 -->
      <div class="version-list">
        <h4>当前版本列表</h4>
        <el-table :data="versions" style="width: 100%">
          <el-table-column prop="version" label="版本号" width="120" />
          <el-table-column prop="versionType" label="版本类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getVersionType(row.versionType)">
                {{ getVersionLabel(row.versionType) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="summary" label="版本说明" />
          <el-table-column prop="createdBy" label="创建者" width="100" />
          <el-table-column prop="createdAt" label="创建时间" width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150">
            <template #default="{ row }">
              <el-button 
                v-if="!row.isActive" 
                size="small" 
                type="primary"
                @click="handleActivateVersion(row)"
              >
                激活
              </el-button>
              <el-tag v-else type="success" size="small">当前版本</el-tag>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 创建新版本 -->
      <div class="create-version">
        <h4>创建新版本</h4>
        <el-form :model="newVersionForm" label-width="100px">
          <el-form-item label="版本类型">
            <el-radio-group v-model="newVersionForm.versionType">
              <el-radio value="initial">初始版本</el-radio>
              <el-radio value="analyzed">分析版本</el-radio>
              <el-radio value="optimized">优化版本</el-radio>
            </el-radio-group>
          </el-form-item>
          <el-form-item label="版本说明">
            <el-input 
              v-model="newVersionForm.summary"
              type="textarea"
              :rows="3"
              placeholder="请输入版本说明"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleCreateVersion">
              创建版本
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  cycleId: {
    type: [Number, String],
    default: null
  },
  versions: {
    type: Array,
    default: () => []
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'version-created', 'version-switched'])

// 响应式数据
const newVersionForm = reactive({
  versionType: 'initial',
  summary: ''
})

// 方法
const getVersionType = (type) => {
  const typeMap = {
    'initial': 'info',
    'analyzed': 'primary', 
    'optimized': 'success'
  }
  return typeMap[type] || 'info'
}

const getVersionLabel = (type) => {
  const labelMap = {
    'initial': '初始版本',
    'analyzed': '分析版本',
    'optimized': '优化版本'
  }
  return labelMap[type] || type
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

const handleCreateVersion = () => {
  if (!newVersionForm.summary.trim()) {
    ElMessage.warning('请输入版本说明')
    return
  }
  
  emit('version-created', {
    cycleId: props.cycleId,
    versionType: newVersionForm.versionType,
    summary: newVersionForm.summary
  })
  
  // 重置表单
  newVersionForm.versionType = 'initial'
  newVersionForm.summary = ''
  
  ElMessage.success('版本创建成功')
}

const handleActivateVersion = (version) => {
  emit('version-switched', version)
  ElMessage.success('版本切换成功')
}
</script>

<style lang="scss" scoped>
.version-management {
  .version-list {
    margin-bottom: 24px;
    
    h4 {
      margin-bottom: 16px;
      color: #303133;
    }
  }
  
  .create-version {
    h4 {
      margin-bottom: 16px;
      color: #303133;
    }
  }
}
</style>