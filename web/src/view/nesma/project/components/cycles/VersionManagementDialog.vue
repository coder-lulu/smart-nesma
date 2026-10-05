<template>
  <el-dialog
    v-model="visible"
    title="版本管理"
    width="900px"
    :close-on-click-modal="false"
    @opened="handleDialogOpened"
  >
    <div class="version-management" v-loading="loading">
      <!-- 工具栏 -->
      <div class="toolbar">
        <div class="cycle-info">
          <el-tag type="info">周期: {{ cycle.name }}</el-tag>
          <span class="description">{{ cycle.description }}</span>
        </div>
        <el-button @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <!-- 版本列表 -->
      <div class="version-list">
        <el-table 
          :data="versions" 
          stripe 
          style="width: 100%"
          @selection-change="handleSelectionChange"
          empty-text="该周期暂无版本"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="version" label="版本号" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.isActive" type="success" size="small">
                {{ row.version }}
              </el-tag>
              <span v-else>{{ row.version }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="versionType" label="版本类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getVersionTypeTag(row.versionType)" size="small">
                {{ getVersionTypeText(row.versionType) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="summary" label="版本说明" min-width="200" show-overflow-tooltip />
          <el-table-column prop="requirementCount" label="需求数量" width="100">
            <template #default="{ row }">
              <el-tag type="info" size="small">{{ row.requirementCount || 0 }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="CreatedAt" label="创建时间" width="160">
            <template #default="{ row }">
              {{ formatDateTime(row.CreatedAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="180">
            <template #default="{ row }">
              <!-- <el-button
                @click="handleSetActive(row)"
                size="small"
                type="success"
                :disabled="row.isActive"
              >
                {{ row.isActive ? '当前版本' : '设为当前' }}
              </el-button> -->
              <!-- <el-button
                @click="handleViewRequirements(row)"
                size="small"
                type="primary"
              >
                查看需求
              </el-button> -->
              <el-button
                @click="handleDeleteVersion(row)"
                size="small"
                type="danger"
                :disabled="row.isActive"
              >
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 批量操作 -->
      <div class="batch-actions" v-if="selectedVersions.length > 0">
        <el-alert 
          :title="`已选择 ${selectedVersions.length} 个版本`" 
          type="warning" 
          show-icon 
          :closable="false"
        />
        <div class="batch-buttons">
          <el-button 
            @click="handleBatchDelete" 
            type="danger" 
            size="small"
            :disabled="selectedVersions.some(v => v.isActive)"
          >
            批量删除版本
          </el-button>
          <el-popconfirm
            title="删除版本将同步删除该版本下的所有需求，此操作不可恢复！确定继续吗？"
            confirm-button-text="确定删除"
            cancel-button-text="取消"
            @confirm="confirmBatchDelete"
          >
            <template #reference>
              <el-button 
                type="danger" 
                size="small"
                :disabled="selectedVersions.some(v => v.isActive)"
              >
                删除版本及需求
              </el-button>
            </template>
          </el-popconfirm>
        </div>
      </div>
    </div>
    
    <template #footer>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh } from '@element-plus/icons-vue'
import { 
  getRequirementVersions, 
  deleteRequirementVersion, 
  setActiveVersion 
} from '@/api/nesma/requirementVersion'
import { deleteRequirementsByCondition } from '@/api/nesma'

const props = defineProps({
  modelValue: Boolean,
  cycle: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue', 'refresh', 'close'])

// 响应式数据
const loading = ref(false)
const versions = ref([])
const selectedVersions = ref([])

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 格式化日期时间
const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString()
}

// 获取版本类型标签
const getVersionTypeTag = (type) => {
  const tags = {
    'initial': '',
    'analyzed': 'success',
    'optimized': 'warning',
    'approved': 'success',
    'ai_analyzed': 'success',
    'ai_optimized': 'warning',
    'ai_approved': 'success'
  }
  return tags[type] || ''
}

// 获取版本类型文本
const getVersionTypeText = (type) => {
  const texts = {
    'initial': '初始版本',
    'analyzed': '分析版本',
    'optimized': '优化版本',
    'approved': '审批版本',
    'ai_analyzed': 'AI分析版本',
    'ai_optimized': 'AI优化版本',
    'ai_approved': 'AI审批版本'
  }
  return texts[type] || type
}

// 加载版本列表
const loadVersions = async () => {
  if (!props.cycle.ID) {
    console.log('周期ID为空:', props.cycle)
    return
  }
  
  console.log('开始加载版本，周期ID:', props.cycle.ID)
  loading.value = true
  try {
    const res = await getRequirementVersions(props.cycle.ID)
    console.log('版本API响应:', res)
    if (res.code === 0) {
      versions.value = res.data || []
      console.log('加载版本成功:', versions.value)
    } else {
      console.error('API返回错误:', res)
      ElMessage.error(res.msg || '加载版本列表失败')
    }
  } catch (error) {
    console.error('加载版本列表失败:', error)
    ElMessage.error('加载版本列表失败')
  } finally {
    loading.value = false
  }
}

// 对话框打开时加载数据
const handleDialogOpened = () => {
  loadVersions()
}

// 刷新列表
const handleRefresh = () => {
  loadVersions()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedVersions.value = selection
}

// 设置激活版本
const handleSetActive = async (version) => {
  try {
    const res = await setActiveVersion(version.ID)
    if (res.code === 0) {
      ElMessage.success('设置当前版本成功')
      await loadVersions()
      emit('refresh')
    } else {
      ElMessage.error(res.msg || '设置当前版本失败')
    }
  } catch (error) {
    console.error('设置当前版本失败:', error)
    ElMessage.error('设置当前版本失败')
  }
}

// 查看版本需求
const handleViewRequirements = (version) => {
  // 这里可以跳转到需求页面，并筛选特定版本的需求
  ElMessage.info(`查看版本 ${version.version} 的需求功能开发中...`)
}

// 删除单个版本
const handleDeleteVersion = async (version) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除版本 "${version.version}" 吗？删除版本将同步删除该版本下的所有需求，此操作不可恢复！`,
      '删除版本确认',
      {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消'
      }
    )
    
    loading.value = true
    
    // 先删除该版本的所有需求
    await deleteRequirementsByCondition({
      projectId: props.cycle.projectId,
      cycleId: props.cycle.ID,
      versionId: version.ID
    })
    
    // 再删除版本记录
    const res = await deleteRequirementVersion(version.ID)
    if (res.code === 0) {
      ElMessage.success('版本删除成功')
      await loadVersions()
      emit('refresh')
    } else {
      ElMessage.error(res.msg || '删除版本失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除版本失败:', error)
      ElMessage.error('删除版本失败')
    }
  } finally {
    loading.value = false
  }
}

// 批量删除版本
const handleBatchDelete = async () => {
  if (selectedVersions.value.length === 0) {
    ElMessage.warning('请先选择要删除的版本')
    return
  }
  
  // 检查是否包含激活版本
  const hasActiveVersion = selectedVersions.value.some(v => v.isActive)
  if (hasActiveVersion) {
    ElMessage.warning('不能删除当前激活的版本')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedVersions.value.length} 个版本吗？删除版本将同步删除这些版本下的所有需求，此操作不可恢复！`,
      '批量删除版本确认',
      {
        type: 'warning',
        confirmButtonText: '确定删除',
        cancelButtonText: '取消'
      }
    )
    
    await confirmBatchDelete()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除版本失败:', error)
      ElMessage.error('批量删除版本失败')
    }
  }
}

// 确认批量删除
const confirmBatchDelete = async () => {
  loading.value = true
  
  try {
    // 批量删除版本及其需求
    const deletePromises = selectedVersions.value.map(async (version) => {
      // 先删除该版本的所有需求
      await deleteRequirementsByCondition({
        projectId: props.cycle.projectId,
        cycleId: props.cycle.ID,
        versionId: version.ID
      })
      
      // 再删除版本记录
      return deleteRequirementVersion(version.ID)
    })
    
    await Promise.all(deletePromises)
    
    ElMessage.success('批量删除成功')
    selectedVersions.value = []
    await loadVersions()
    emit('refresh')
  } catch (error) {
    console.error('批量删除失败:', error)
    ElMessage.error('批量删除失败')
  } finally {
    loading.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  emit('close')
}

// 监听周期变化
watch(() => props.cycle, (newCycle) => {
  if (newCycle.ID && visible.value) {
    loadVersions()
  }
}, { immediate: true })
</script>

<style lang="scss" scoped>
.version-management {
  .toolbar {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 16px;
    
    .cycle-info {
      display: flex;
      align-items: center;
      gap: 12px;
      
      .description {
        color: #666;
        font-size: 14px;
      }
    }
  }

  .version-list {
    margin-bottom: 16px;
  }

  .batch-actions {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px;
    background: #fef0f0;
    border: 1px solid #fbc4c4;
    border-radius: 6px;
    margin-top: 16px;

    .batch-buttons {
      display: flex;
      gap: 8px;
    }
  }
}
</style>