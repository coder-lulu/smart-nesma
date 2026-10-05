<template>
  <div class="requirement-table">
    <!-- 表格工具栏 -->
    <div class="table-toolbar">
      <div class="toolbar-left">
        <el-checkbox 
          v-model="selectAll" 
          :indeterminate="isIndeterminate"
          @change="handleSelectAll"
        >
          全选
        </el-checkbox>
        <span class="selection-info">
          已选择 {{ selectedRows.length }} 项
        </span>
        <el-button 
          v-if="selectedRows.length > 0" 
          type="danger" 
          size="small"
          @click="handleBatchDelete"
        >
          <el-icon><Delete /></el-icon>
          批量删除
        </el-button>
        <el-button 
          v-if="selectedRows.length > 0" 
          type="primary" 
          size="small"
          @click="handleBatchExport"
        >
          <el-icon><Download /></el-icon>
          批量导出
        </el-button>
        <el-button 
          v-if="selectedRows.length > 0" 
          type="warning" 
          size="small"
          @click="handleBatchAnalysis"
        >
          <el-icon><MagicStick /></el-icon>
          批量分析
        </el-button>
      </div>
      
      <div class="toolbar-right">
        <el-tooltip content="列设置" placement="top">
          <el-button circle size="small" @click="showColumnSettings = true">
            <el-icon><Setting /></el-icon>
          </el-button>
        </el-tooltip>
        <el-tooltip content="刷新" placement="top">
          <el-button circle size="small" @click="$emit('refresh')">
            <el-icon><Refresh /></el-icon>
          </el-button>
        </el-tooltip>
      </div>
    </div>

    <!-- 主表格 -->
    <el-table
      ref="tableRef"
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      row-key="ID"
      :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
      :default-expand-all="defaultExpandAll"
      :expand-row-keys="expandedRows"
      @selection-change="handleSelectionChange"
      @row-click="handleRowClick"
      @row-dblclick="handleRowDoubleClick"
      @sort-change="handleSortChange"
      :height="tableHeight"
    >
      <!-- 选择列 -->
      <el-table-column 
        v-if="showColumns.selection"
        type="selection" 
        width="55" 
        :selectable="(row) => row.status !== 'cancelled'"
      />
      
      <!-- 展开列 -->
      <el-table-column 
        v-if="showColumns.expand"
        type="expand"
        width="55"
      >
        <template #default="{ row }">
          <div class="expand-content">
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="需求描述" :span="2">
                {{ row.description || '无描述' }}
              </el-descriptions-item>
              <el-descriptions-item v-if="row.aiDescription" label="AI优化描述" :span="2">
                <div class="ai-description">{{ row.aiDescription }}</div>
              </el-descriptions-item>
              <el-descriptions-item label="业务价值">
                {{ row.businessValue || '无' }}
              </el-descriptions-item>
              <el-descriptions-item label="验收标准">
                {{ row.acceptanceCriteria || '无' }}
              </el-descriptions-item>
              <el-descriptions-item label="预估工时">
                {{ row.estimateHours || 0 }} 小时
              </el-descriptions-item>
              <el-descriptions-item label="实际工时">
                {{ row.actualHours || 0 }} 小时
              </el-descriptions-item>
              <el-descriptions-item v-if="row.notes" label="备注" :span="2">
                {{ row.notes }}
              </el-descriptions-item>
            </el-descriptions>
          </div>
        </template>
      </el-table-column>

      <!-- 需求编号 -->
      <el-table-column 
        v-if="showColumns.code"
        prop="code" 
        label="编号" 
        width="120"
        sortable="custom"
      />

      <!-- 需求标题 -->
      <el-table-column 
        v-if="showColumns.title"
        prop="title" 
        label="需求标题" 
        min-width="250"
        show-overflow-tooltip
      >
        <template #default="{ row }">
          <div class="title-cell">
            <el-icon class="level-icon" :class="`level-${row.level}`">
              <component :is="getNodeIcon(row.level)" />
            </el-icon>
            <el-link 
              @click="handleView(row)" 
              type="primary"
              :class="{ 'has-ai-optimization': row.aiAnalysisStatus === 'completed' }"
            >
              {{ getDisplayTitle(row) }}
            </el-link>
            <el-tooltip v-if="row.aiAnalysisStatus === 'completed'" content="已AI优化" placement="top">
              <el-icon class="ai-indicator">
                <MagicStick />
              </el-icon>
            </el-tooltip>
          </div>
        </template>
      </el-table-column>

      <!-- 层级 -->
      <el-table-column 
        v-if="showColumns.level"
        prop="level" 
        label="层级" 
        width="100"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag size="small" :type="getLevelType(row.level)">
            {{ getLevelLabel(row.level) }}
          </el-tag>
        </template>
      </el-table-column>

      <!-- 状态 -->
      <el-table-column 
        v-if="showColumns.status"
        prop="status" 
        label="状态" 
        width="100"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag size="small" :type="getStatusType(row.status)">
            {{ getStatusLabel(row.status) }}
          </el-tag>
        </template>
      </el-table-column>

      <!-- 功能类型 -->
      <el-table-column 
        v-if="showColumns.functionType"
        prop="functionType" 
        label="功能类型" 
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag v-if="row.functionType" size="small" type="info">
            {{ row.functionType }}
          </el-tag>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>

      <!-- 优先级 -->
      <el-table-column 
        v-if="showColumns.priority"
        prop="priority" 
        label="优先级" 
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-rate
            v-model="row.priority"
            disabled
            show-score
            text-color="#ff9900"
            score-template="{value}"
            :max="5"
            size="small"
          />
        </template>
      </el-table-column>

      <!-- 复杂度 -->
      <el-table-column 
        v-if="showColumns.complexity"
        prop="complexity" 
        label="复杂度" 
        width="100"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag 
            v-if="row.complexity"
            size="small" 
            :type="getComplexityType(row.complexity)"
          >
            {{ row.complexity }}
          </el-tag>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>

      <!-- 功能点数 -->
      <el-table-column 
        v-if="showColumns.functionPoints"
        label="功能点" 
        width="120"
        sortable="custom"
        sort-by="afp"
      >
        <template #default="{ row }">
          <div v-if="row.afp > 0 || row.ufp > 0" class="function-points">
            <el-tooltip :content="`AFP: ${row.afp || 0}, UFP: ${row.ufp || 0}`" placement="top">
              <el-tag size="small" type="warning">
                {{ (row.afp || row.ufp || 0).toFixed(1) }}
              </el-tag>
            </el-tooltip>
          </div>
          <span v-else class="text-muted">-</span>
        </template>
      </el-table-column>

      <!-- AI分析状态 -->
      <el-table-column 
        v-if="showColumns.aiStatus"
        prop="aiAnalysisStatus" 
        label="AI状态" 
        width="120"
        sortable="custom"
      >
        <template #default="{ row }">
          <el-tag 
            size="small" 
            :type="getAIStatusType(row.aiAnalysisStatus)"
          >
            {{ getAIStatusLabel(row.aiAnalysisStatus) }}
          </el-tag>
        </template>
      </el-table-column>

      <!-- 创建时间 -->
      <el-table-column 
        v-if="showColumns.createdAt"
        prop="createdAt" 
        label="创建时间" 
        width="160"
        sortable="custom"
      >
        <template #default="{ row }">
          {{ formatDateTime(row.createdAt) }}
        </template>
      </el-table-column>

      <!-- 更新时间 -->
      <el-table-column 
        v-if="showColumns.updatedAt"
        prop="updatedAt" 
        label="更新时间" 
        width="160"
        sortable="custom"
      >
        <template #default="{ row }">
          {{ formatDateTime(row.updatedAt) }}
        </template>
      </el-table-column>

      <!-- 操作列 -->
      <el-table-column 
        label="操作" 
        width="280" 
        fixed="right"
      >
        <template #default="{ row }">
          <div class="action-buttons">
            <el-tooltip content="查看详情" placement="top">
              <el-button size="small" circle @click="handleView(row)">
                <el-icon><View /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip content="编辑" placement="top">
              <el-button size="small" type="primary" circle @click="handleEdit(row)">
                <el-icon><Edit /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip v-if="row.level < 4" content="添加子需求" placement="top">
              <el-button size="small" type="success" circle @click="handleCreateChild(row)">
                <el-icon><Plus /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip v-if="row.level >= 3 && !row.aiAnalysisStatus" content="AI分析" placement="top">
              <el-button size="small" type="warning" circle @click="handleAIAnalysis(row)">
                <el-icon><MagicStick /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-dropdown @command="(command) => handleMoreActions(command, row)" trigger="click">
              <el-button size="small" circle>
                <el-icon><More /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="copy" icon="DocumentCopy">复制</el-dropdown-item>
                  <el-dropdown-item command="move" icon="Rank">移动</el-dropdown-item>
                  <el-dropdown-item command="export" icon="Download">导出</el-dropdown-item>
                  <el-dropdown-item command="dependencies" icon="Share">依赖关系</el-dropdown-item>
                  <el-dropdown-item divided />
                  <el-dropdown-item command="delete" icon="Delete">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="table-pagination" v-if="showPagination">
      <el-pagination

        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 列设置对话框 -->
    <el-dialog
      v-model="showColumnSettings"
      title="列设置"
      width="500px"
    >
      <div class="column-settings">
        <el-checkbox-group v-model="selectedColumns">
          <div v-for="(label, key) in availableColumns" :key="key" class="column-item">
            <el-checkbox :value="key" :disabled="key === 'title'">
              {{ label }}
            </el-checkbox>
          </div>
        </el-checkbox-group>
      </div>
      <template #footer>
        <el-button @click="resetColumnSettings">重置</el-button>
        <el-button type="primary" @click="saveColumnSettings">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Delete, Download, MagicStick, Setting, Refresh, View, Edit, Plus, More,
  Document, Files, List, Operation, DocumentCopy, Rank, Share
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  tableData: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  total: {
    type: Number,
    default: 0
  },
  currentPage: {
    type: Number,
    default: 1
  },
  pageSize: {
    type: Number,
    default: 20
  },
  showPagination: {
    type: Boolean,
    default: true
  },
  defaultExpandAll: {
    type: Boolean,
    default: false
  },
  tableHeight: {
    type: [String, Number],
    default: '600px'
  }
})

// Emits
const emit = defineEmits([
  'view',
  'edit',
  'create-child',
  'delete',
  'ai-analysis',
  'copy',
  'move',
  'export',
  'dependencies',
  'batch-delete',
  'batch-export',
  'batch-analysis',
  'selection-change',
  'row-click',
  'row-dblclick',
  'sort-change',
  'size-change',
  'current-change',
  'refresh'
])

// 响应式数据
const tableRef = ref(null)
const selectedRows = ref([])
const expandedRows = ref([])
const showColumnSettings = ref(false)

// 可用的列配置
const availableColumns = {
  selection: '选择',
  expand: '展开',
  code: '编号',
  title: '标题',
  level: '层级',
  status: '状态',
  functionType: '功能类型',
  priority: '优先级',
  complexity: '复杂度',
  functionPoints: '功能点',
  aiStatus: 'AI状态',
  createdAt: '创建时间',
  updatedAt: '更新时间'
}

// 默认显示的列
const defaultColumns = ['selection', 'code', 'title', 'level', 'status', 'priority', 'complexity', 'createdAt']
const selectedColumns = ref([...defaultColumns])

// 计算属性
const showColumns = computed(() => {
  const result = {}
  Object.keys(availableColumns).forEach(key => {
    result[key] = selectedColumns.value.includes(key)
  })
  return result
})

const selectAll = computed({
  get: () => {
    // 检查是否所有可选择的行都被选中
    const selectableRows = props.tableData.filter(row => row.status !== 'cancelled')
    return selectableRows.length > 0 && selectedRows.value.length === selectableRows.length
  },
  set: (value) => {
    // 这里会被 handleSelectAll 处理
  }
})

const isIndeterminate = computed(() => {
  const selectableRows = props.tableData.filter(row => row.status !== 'cancelled')
  return selectedRows.value.length > 0 && selectedRows.value.length < selectableRows.length
})

// 方法
const getDisplayTitle = (row) => {
  return row.aiGeneratedTitle || row.title
}

const getNodeIcon = (level) => {
  const iconMap = {
    1: Document,
    2: Files,
    3: List,
    4: Operation
  }
  return iconMap[level] || Document
}

const getLevelType = (level) => {
  const typeMap = {
    1: 'danger',
    2: 'warning',
    3: 'primary',
    4: 'success'
  }
  return typeMap[level] || 'info'
}

const getLevelLabel = (level) => {
  const labelMap = {
    1: 'L1',
    2: 'L2',
    3: 'L3',
    4: 'L4'
  }
  return labelMap[level] || `L${level}`
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'in_progress': 'warning',
    'completed': 'success',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    'pending': '待处理',
    'in_progress': '进行中',
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return labelMap[status] || status
}

const getComplexityType = (complexity) => {
  const typeMap = {
    '简单': 'success',
    '中等': 'warning',
    '复杂': 'danger'
  }
  return typeMap[complexity] || 'info'
}

const getAIStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'analyzing': 'warning',
    'completed': 'success',
    'failed': 'danger'
  }
  return typeMap[status] || 'info'
}

const getAIStatusLabel = (status) => {
  const labelMap = {
    'pending': '待分析',
    'analyzing': '分析中',
    'completed': '已完成',
    'failed': '失败'
  }
  return labelMap[status] || status
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 事件处理
const handleSelectionChange = (selection) => {
  selectedRows.value = selection
  emit('selection-change', selection)
}

const handleSelectAll = (value) => {
  if (value) {
    // 使用表格API选择所有可选择的行
    const selectableRows = props.tableData.filter(row => row.status !== 'cancelled')
    selectableRows.forEach(row => {
      tableRef.value?.toggleRowSelection(row, true)
    })
  } else {
    // 清除所有选择
    tableRef.value?.clearSelection()
  }
}

const handleRowClick = (row, column, event) => {
  emit('row-click', row, column, event)
}

const handleRowDoubleClick = (row, column, event) => {
  emit('row-dblclick', row, column, event)
}

const handleSortChange = ({ column, prop, order }) => {
  emit('sort-change', { column, prop, order })
}

const handleSizeChange = (size) => {
  emit('size-change', size)
}

const handleCurrentChange = (page) => {
  emit('current-change', page)
}

const handleView = (row) => {
  emit('view', row)
}

const handleEdit = (row) => {
  emit('edit', row)
}

const handleCreateChild = (row) => {
  emit('create-child', row)
}

const handleAIAnalysis = (row) => {
  emit('ai-analysis', row)
}

const handleMoreActions = (command, row) => {
  switch (command) {
    case 'copy':
      emit('copy', row)
      break
    case 'move':
      emit('move', row)
      break
    case 'export':
      emit('export', row)
      break
    case 'dependencies':
      emit('dependencies', row)
      break
    case 'delete':
      ElMessageBox.confirm(
        `确定要删除需求"${row.title}"吗？此操作不可撤销。`,
        '确认删除',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(() => {
        emit('delete', row)
      }).catch(() => {
        // 用户取消删除
      })
      break
  }
}

const handleBatchDelete = () => {
  ElMessageBox.confirm(
    `确定要删除选中的 ${selectedRows.value.length} 个需求吗？此操作不可撤销。`,
    '确认批量删除',
    {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    }
  ).then(() => {
    emit('batch-delete', selectedRows.value)
  }).catch(() => {
    // 用户取消删除
  })
}

const handleBatchExport = () => {
  emit('batch-export', selectedRows.value)
}

const handleBatchAnalysis = () => {
  emit('batch-analysis', selectedRows.value)
}

// 列设置相关
const saveColumnSettings = () => {
  // 保存到localStorage
  localStorage.setItem('requirement-table-columns', JSON.stringify(selectedColumns.value))
  showColumnSettings.value = false
  ElMessage.success('列设置已保存')
}

const resetColumnSettings = () => {
  selectedColumns.value = [...defaultColumns]
  localStorage.removeItem('requirement-table-columns')
  ElMessage.success('列设置已重置')
}

// 监听表格数据变化，重置选择状态
watch(() => props.tableData, () => {
  // 当表格数据变化时，清除选择状态
  selectedRows.value = []
  if (tableRef.value) {
    tableRef.value.clearSelection()
  }
}, { deep: true })

// 组件挂载时加载列设置
onMounted(() => {
  const savedColumns = localStorage.getItem('requirement-table-columns')
  if (savedColumns) {
    try {
      selectedColumns.value = JSON.parse(savedColumns)
    } catch (e) {
      console.warn('加载列设置失败:', e)
    }
  }
})
</script>

<style lang="scss" scoped>
.requirement-table {
  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding: 12px 16px;
    background: #f8fafc;
    border-radius: 8px;

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .selection-info {
        font-size: 14px;
        color: #606266;
      }
    }

    .toolbar-right {
      display: flex;
      gap: 8px;
    }
  }

  .title-cell {
    display: flex;
    align-items: center;
    gap: 8px;

    .level-icon {
      font-size: 16px;

      &.level-1 { color: #f56c6c; }
      &.level-2 { color: #e6a23c; }
      &.level-3 { color: #409eff; }
      &.level-4 { color: #67c23a; }
    }

    .has-ai-optimization {
      color: #409eff;
      font-weight: 600;
    }

    .ai-indicator {
      font-size: 14px;
      color: #409eff;
    }
  }

  .expand-content {
    padding: 16px;
    background: #f8fafc;
    border-radius: 8px;

    .ai-description {
      color: #409eff;
      background: #f0f9ff;
      padding: 8px 12px;
      border-radius: 4px;
      border-left: 3px solid #409eff;
    }
  }

  .action-buttons {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;

    .el-button {
      width: 28px;
      height: 28px;
    }
  }

  .function-points {
    display: flex;
    align-items: center;
    gap: 4px;
  }

  .text-muted {
    color: #c0c4cc;
  }

  .table-pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }

  .column-settings {
    .column-item {
      padding: 8px 0;
      border-bottom: 1px solid #ebeef5;

      &:last-child {
        border-bottom: none;
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .requirement-table {
    .table-toolbar {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;

      .toolbar-left,
      .toolbar-right {
        justify-content: center;
      }
    }

    .action-buttons {
      .el-button {
        width: 24px;
        height: 24px;
      }
    }
  }
}
</style>