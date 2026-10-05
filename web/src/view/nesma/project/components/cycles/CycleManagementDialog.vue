<template>
  <el-dialog
    v-model="visible"
    title="周期管理"
    width="1100px"
    :close-on-click-modal="false"
    @opened="handleDialogOpened"
  >
    <div class="cycle-management">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-button type="primary" @click="handleAddCycle">
          <el-icon><Plus /></el-icon>
          新建周期
        </el-button>
        <el-button @click="handleRefresh">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
        <div class="project-info">
          <el-tag type="info">项目: {{ project.name }}</el-tag>
        </div>
      </div>

      <!-- 周期列表 -->
      <div class="cycle-list" v-loading="loading">
        <el-table 
          :data="cycles" 
          stripe 
          style="width: 100%"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="name" label="周期名称" width="120">
            <template #default="{ row }">
              <el-tag v-if="row.id === project.activeCycleId" type="success" size="small">
                {{ row.name }}
              </el-tag>
              <span v-else>{{ row.name }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" width="200" show-overflow-tooltip />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusType(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="phase" label="阶段" width="100">
            <template #default="{ row }">
              <el-tag type="info" size="small">
                {{ getPhaseText(row.phase) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="起止时间" width="120">
            <template #default="{ row }">
              <div class="date-range-cell">
                <div class="date-start">
                  <span class="date-label">开始：</span>{{ formatDate(row.startDate) }}
                </div>
                <div class="date-end">
                  <span class="date-label">结束：</span>{{ formatDate(row.endDate) }}
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="80">
            <template #default="{ row }">
              {{ row.progress }}%
            </template>
          </el-table-column>
          <el-table-column label="操作" width="250">
            <template #default="{ row }">
              <div class="action-btns">
                <div class="action-row">
                  <el-button @click="handleEditCycle(row)" size="small">编辑</el-button>
                  <el-button 
                    @click="handleSetActive(row)" 
                    size="small" 
                    type="success"
                    :disabled="row.ID === activeCycleId"
                  >
                    {{ row.ID === activeCycleId ? '已激活' : '设为激活' }}
                  </el-button>
                  <el-button @click="handleDeleteCycle(row)" size="small" type="danger">删除</el-button>
                </div>
                <div class="action-row">
                  <el-button @click="handleDropdownCommand('view-versions', row)" size="small" type="primary">版本管理</el-button>
                </div>
              </div>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 批量操作 -->
      <div class="batch-actions" v-if="selectedCycles.length > 0">
        <el-alert 
          :title="`已选择 ${selectedCycles.length} 个周期`" 
          type="info" 
          show-icon 
          :closable="false"
        />
        <div class="batch-buttons">
          <el-button @click="handleBatchDelete" type="danger" size="small">
            批量删除
          </el-button>
        </div>
      </div>
    </div>

    <!-- 周期表单对话框 -->
    <CycleFormDialog 
      v-model="showCycleForm"
      :form-data="cycleFormData"
      :mode="cycleFormMode"
      :project-id="props.project.ID"
      @save="handleCycleSave"
      @cancel="handleCycleCancel"
    />

    <!-- 版本管理对话框 -->
    <VersionManagementDialog
      v-model="showVersionManagement"
      :cycle="selectedCycle"
      @refresh="loadProjectCycles"
      @close="handleVersionManagementClose"
    />
    
    <template #footer>
      <el-button @click="handleClose">关闭</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Refresh, ArrowDown } from '@element-plus/icons-vue'
import { getProjectCycles, deleteProjectCycle, setActiveProjectCycle } from '@/api/projectCycle'
import CycleFormDialog from './CycleFormDialog.vue'
import VersionManagementDialog from './VersionManagementDialog.vue'

const props = defineProps({
  modelValue: Boolean,
  project: {
    type: Object,
    default: () => ({})
  },
  activeCycleId: {
    type: Number,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'save', 'close', 'update-active-cycle'])

// 响应式数据
const loading = ref(false)
const cycles = ref([])
const selectedCycles = ref([])
const showCycleForm = ref(false)
const cycleFormData = ref({})
const cycleFormMode = ref('create')

// 版本管理相关
const showVersionManagement = ref(false)
const selectedCycle = ref({})

const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 格式化日期
const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString()
}

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    'planning': '',
    'active': 'success',
    'completed': 'info',
    'suspended': 'warning'
  }
  return types[status] || ''
}

// 获取状态文本
const getStatusText = (status) => {
  const texts = {
    'planning': '规划中',
    'active': '进行中',
    'completed': '已完成',
    'suspended': '暂停'
  }
  return texts[status] || status
}

// 获取阶段文本
const getPhaseText = (phase) => {
  const texts = {
    'requirement': '需求分析',
    'design': '系统设计',
    'development': '开发实施',
    'testing': '测试验证',
    'deployment': '部署上线'
  }
  return texts[phase] || phase || '未设置'
}

// 加载项目周期列表
const loadProjectCycles = async () => {
  if (!props.project.ID) {
    console.log('项目ID为空:', props.project)
    return
  }
  if (props.activeCycleId) {
    console.log('活跃周期ID:', props.activeCycleId)
  }
  
  console.log('开始加载项目周期，项目ID:', props.project.ID)
  loading.value = true
  try {
    const res = await getProjectCycles(props.project.ID)
    console.log('API响应:', res)
    if (res.code === 0) {
      cycles.value = res.data || []
      console.log('加载周期成功:', cycles.value)
      // 触发表格刷新
    } else {
      console.error('API返回错误:', res)
      ElMessage.error(res.msg || '加载周期列表失败')
    }
  } catch (error) {
    console.error('加载周期列表失败:', error)
    ElMessage.error('加载周期列表失败')
  } finally {
    loading.value = false
  }
}

// 对话框打开时加载数据
const handleDialogOpened = () => {
  loadProjectCycles()
}

// 刷新列表
const handleRefresh = () => {
  loadProjectCycles()
}

// 添加周期
const handleAddCycle = () => {
  cycleFormData.value = {}
  cycleFormMode.value = 'create'
  showCycleForm.value = true
}

// 编辑周期
const handleEditCycle = (cycle) => {
  cycleFormData.value = { ...cycle }
  cycleFormMode.value = 'edit'
  showCycleForm.value = true
}

// 设置激活周期
const handleSetActive = async (cycle) => {
  try {
    const res = await setActiveProjectCycle({ cycleId: cycle.ID })
    if (res.code === 0) {
      ElMessage.success('设置激活周期成功')
      
      // 通知父组件更新激活周期ID
      emit('update-active-cycle', cycle.ID)
      
      // 刷新周期列表以更新表格显示
      await loadProjectCycles()
    } else {
      ElMessage.error(res.msg || '设置激活周期失败')
    }
  } catch (error) {
    console.error('设置激活周期失败:', error)
    ElMessage.error('设置激活周期失败')
  }
}

// 删除周期
const handleDeleteCycle = async (cycle) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除周期 "${cycle.name}" 吗？`,
      '删除确认',
      {
        type: 'warning'
      }
    )
    
    const res = await deleteProjectCycle(cycle.ID)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      loadProjectCycles()
    } else {
      ElMessage.error(res.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除周期失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 批量删除
const handleBatchDelete = async () => {
  if (selectedCycles.value.length === 0) {
    ElMessage.warning('请先选择要删除的周期')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedCycles.value.length} 个周期吗？`,
      '批量删除确认',
      {
        type: 'warning'
      }
    )
    
    // 批量删除
    const deletePromises = selectedCycles.value.map(cycle => deleteProjectCycle(cycle.ID))
    await Promise.all(deletePromises)
    
    ElMessage.success('批量删除成功')
    selectedCycles.value = []
    loadProjectCycles()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedCycles.value = selection
}

// 下拉菜单命令处理
const handleDropdownCommand = (command, cycle) => {
  selectedCycle.value = cycle
  
  switch (command) {
    case 'view-versions':
      handleViewVersions(cycle)
      break
    case 'delete-versions':
      handleDeleteVersions(cycle)
      break
    default:
      console.warn('未知命令:', command)
  }
}

// 查看版本
const handleViewVersions = (cycle) => {
  selectedCycle.value = cycle
  showVersionManagement.value = true
}

// 删除版本
const handleDeleteVersions = (cycle) => {
  selectedCycle.value = cycle
  showVersionManagement.value = true
  // 传递删除模式标志给版本管理对话框
  selectedCycle.value._deleteMode = true
}

// 版本管理对话框关闭
const handleVersionManagementClose = () => {
  showVersionManagement.value = false
  selectedCycle.value = {}
}

// 周期表单保存
const handleCycleSave = () => {
  showCycleForm.value = false
  loadProjectCycles()
  emit('save') // 通知父组件刷新项目数据
}

// 周期表单取消
const handleCycleCancel = () => {
  showCycleForm.value = false
  cycleFormData.value = {}
}

// 关闭对话框
const handleClose = () => {
  emit('close')
}
</script>

<style lang="scss" scoped>
.cycle-management {
  .toolbar {
    display: flex;
    align-items: center;
    gap: 12px;
    margin-bottom: 16px;
    
    .project-info {
      margin-left: auto;
    }
  }

  .cycle-list {
    margin-bottom: 16px;
  }

  .batch-actions {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 12px;
    background: #f5f7fa;
    border-radius: 6px;
    margin-top: 16px;

    .batch-buttons {
      display: flex;
      gap: 8px;
    }
  }

  .action-btns {
    display: flex;
    flex-direction: column;
    gap: 4px;
    .action-row {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
      margin-bottom: 2px;
    }
  }

  .date-range-cell {
    display: flex;
    flex-direction: column;
    .date-start, .date-end {
      font-size: 12px;
      line-height: 1.2;
      color: #666;
    }
    .date-label {
      color: #999;
      font-size: 11px;
      margin-right: 2px;
    }
    .date-end {
      margin-top: 2px;
    }
  }
}
</style>