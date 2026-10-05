<template>
  <div class="project-table">
    <el-card class="table-card" :body-style="{ padding: '20px' }">
      <!-- 表格工具栏 -->
      <div class="table-toolbar">
        <div class="toolbar-left">
          <el-checkbox 
            v-model="selectAll" 
            @change="handleSelectAll"
            :indeterminate="isIndeterminate"
          >
            全选
          </el-checkbox>
          <span class="selection-info" v-if="selectedRows.length">
            已选择 {{ selectedRows.length }} 项
          </span>
        </div>
        
        <div class="toolbar-right">
          <el-button-group size="small">
            <el-button @click="refreshTable">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
            <el-button @click="toggleColumnSettings">
              <el-icon><Setting /></el-icon>
              列设置
            </el-button>
            <el-button @click="exportTable">
              <el-icon><Download /></el-icon>
              导出
            </el-button>
          </el-button-group>
        </div>
      </div>

      <!-- 主表格 -->
      <el-table
        ref="tableRef"
        v-loading="loading"
        :data="tableData"
        style="width: 100%"
        @selection-change="handleSelectionChange"
        @sort-change="handleSortChange"
        @row-click="handleRowClick"
        :row-class-name="getRowClassName"
        height="600"
      >
        <!-- 选择列 -->
        <el-table-column 
          type="selection" 
          width="50" 
          :selectable="checkSelectable"
          v-if="showSelection"
        />
        
        <!-- ID列 -->
        <el-table-column 
          prop="ID" 
          label="ID" 
          width="80" 
          sortable="custom"
          v-if="visibleColumns.includes('id')"
        />
        
        <!-- 项目名称 -->
        <el-table-column 
          prop="name" 
          label="项目名称" 
          min-width="180" 
          show-overflow-tooltip
          v-if="visibleColumns.includes('name')"
        >
          <template #default="{ row }">
            <div class="project-name-cell">
              <el-link 
                @click="$emit('view', row)" 
                type="primary"
                class="project-link"
              >
                <el-icon><FolderOpened /></el-icon>
                {{ row.name }}
              </el-link>
              <div class="project-tags" v-if="row.domainTags && row.domainTags.length">

                <el-tag v-if="row.domainTags.length > 2" size="small" type="info">
                  +{{ row.domainTags.length - 2 }}
                </el-tag>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <!-- 项目描述 -->
        <el-table-column 
          prop="description" 
          label="项目描述" 
          min-width="200" 
          show-overflow-tooltip
          v-if="visibleColumns.includes('description')"
        >
          <template #default="{ row }">
            <div class="description-cell">
              {{ row.description || '暂无描述' }}
            </div>
          </template>
        </el-table-column>
        
        <!-- 项目领域 -->
        <el-table-column 
          prop="domain" 
          label="项目领域" 
          width="120"
          v-if="visibleColumns.includes('domain')"
        >
          <template #default="{ row }">
            <el-tag :type="getDomainTagType(row.domain)" size="small">
              {{ getDomainLabel(row.domain) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <!-- 项目状态 -->
        <el-table-column 
          prop="status" 
          label="状态" 
          width="100"
          v-if="visibleColumns.includes('status')"
        >
          <template #default="{ row }">
            <el-tag :type="getStatusType(row.status)" size="small" effect="dark">
              <el-icon><component :is="getStatusIcon(row.status)" /></el-icon>
              {{ getStatusLabel(row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        
        <!-- 需求统计 -->
        <el-table-column 
          label="需求统计" 
          width="180"
          v-if="visibleColumns.includes('requirements')"
        >
          <template #default="{ row }">
            <div class="requirement-stats">
              <div class="stats-main">
                <div class="total-count">
                  <span class="label">总计:</span>
                  <span class="value">{{ row.requirementStats?.totalCount || 0 }}</span>
                </div>
                <div class="completed-count">
                  <span class="label">完成:</span>
                  <span class="value success">{{ row.requirementStats?.completed || 0 }}</span>
                </div>
              </div>
              
              <el-popover
                placement="top"
                :width="200"
                trigger="hover"
                v-if="row.requirementStats"
              >
                <template #reference>
                  <el-link type="primary" :underline="false" class="detail-link">
                    详情
                  </el-link>
                </template>
                <div class="stats-detail">
                  <div class="detail-row">
                    <el-tag size="small" type="danger">L1</el-tag>
                    <span>{{ row.requirementStats.level1 || 0 }}</span>
                  </div>
                  <div class="detail-row">
                    <el-tag size="small" type="warning">L2</el-tag>
                    <span>{{ row.requirementStats.level2 || 0 }}</span>
                  </div>
                  <div class="detail-row">
                    <el-tag size="small" type="info">L3</el-tag>
                    <span>{{ row.requirementStats.level3 || 0 }}</span>
                  </div>
                  <div class="detail-row">
                    <el-tag size="small" type="success">L4</el-tag>
                    <span>{{ row.requirementStats.level4 || 0 }}</span>
                  </div>
                </div>
              </el-popover>
            </div>
          </template>
        </el-table-column>
        
        <!-- 当前周期 -->
        <el-table-column 
          label="当前周期" 
          width="150"
          v-if="visibleColumns.includes('cycles')"
        >
          <template #default="{ row }">
            <div v-if="row.activeCycleId" class="cycle-info">
              <el-tag
                size="small"
                type="success"
                effect="light"
                class="cycle-tag"
              >
                <el-icon><Star /></el-icon>
                {{ getActiveCycleName(row) }}
              </el-tag>
            </div>
            <span v-else class="no-cycle">未设置</span>
          </template>
        </el-table-column>
        
        <!-- 时间信息 -->
        <el-table-column 
          label="时间信息" 
          width="200"
          v-if="visibleColumns.includes('dates')"
        >
          <template #default="{ row }">
            <div class="date-info">
              <div class="date-row">
                <span class="date-label">开始:</span>
                <span class="date-value">{{ formatDate(row.startDate) }}</span>
              </div>
              <div class="date-row">
                <span class="date-label">结束:</span>
                <span class="date-value">{{ formatDate(row.endDate) }}</span>
              </div>
              <div class="date-row">
                <span class="date-label">创建:</span>
                <span class="date-value">{{ formatDate(row.createdAt) }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        
        <!-- 操作列 -->
        <el-table-column 
          label="操作" 
          width="300" 
          fixed="right"
          v-if="visibleColumns.includes('actions')"
        >
          <template #default="{ row }">
            <div class="action-buttons">
              <!-- 基础操作 -->
              <el-button-group size="small" class="basic-actions">
                <el-button @click="$emit('view', row)">
                  <el-icon><View /></el-icon>
                  查看
                </el-button>
                <el-button type="primary" @click="$emit('edit', row)">
                  <el-icon><Edit /></el-icon>
                  编辑
                </el-button>
              </el-button-group>
              
              <!-- 业务操作 -->
              <el-dropdown @command="(action) => handleAction(action, row)" class="action-dropdown">
                <el-button size="small">
                  更多操作
                  <el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="requirements">
                      <el-icon><Document /></el-icon>
                      需求管理
                    </el-dropdown-item>
                    <el-dropdown-item command="cycles">
                      <el-icon><Calendar /></el-icon>
                      周期管理
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="analysis" 
                      :disabled="row.status !== 'active'"
                    >
                      <el-icon><MagicStick /></el-icon>
                      一键分析
                    </el-dropdown-item>
                    <el-dropdown-item command="tasks">
                      <el-icon><List /></el-icon>
                      任务管理
                    </el-dropdown-item>
                    <el-dropdown-item divided command="duplicate">
                      <el-icon><CopyDocument /></el-icon>
                      复制项目
                    </el-dropdown-item>
                    <el-dropdown-item command="export">
                      <el-icon><Download /></el-icon>
                      导出数据
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="archive"
                      v-if="row.status !== 'archived'"
                    >
                      <el-icon><Box /></el-icon>
                      归档项目
                    </el-dropdown-item>
                    <el-dropdown-item 
                      command="restore"
                      v-else
                    >
                      <el-icon><RefreshRight /></el-icon>
                      恢复项目
                    </el-dropdown-item>
                    <el-dropdown-item divided command="delete" class="danger-item">
                      <el-icon><Delete /></el-icon>
                      删除项目
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
          background
        />
      </div>
    </el-card>

    <!-- 列设置对话框 -->
    <el-dialog v-model="columnSettingsVisible" title="列设置" width="400px">
      <el-checkbox-group v-model="visibleColumns">
        <div class="column-setting-item" v-for="column in columnOptions" :key="column.key">
          <el-checkbox :value="column.key">{{ column.label }}</el-checkbox>
        </div>
      </el-checkbox-group>
      <template #footer>
        <el-button @click="columnSettingsVisible = false">取消</el-button>
        <el-button type="primary" @click="saveColumnSettings">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Refresh, Setting, Download, FolderOpened, View, Edit, ArrowDown,
  Document, Calendar, MagicStick, CopyDocument, Box, RefreshRight, Delete,
  Star, VideoPlay, Check, CircleClose, Clock, List
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
  showSelection: {
    type: Boolean,
    default: true
  }
})

// Emits
const emit = defineEmits([
  'view',
  'edit', 
  'selection-change',
  'sort-change',
  'size-change',
  'current-change',
  'action',
  'refresh'
])

// 响应式数据
const tableRef = ref(null)
const selectAll = ref(false)
const selectedRows = ref([])
const columnSettingsVisible = ref(false)

const visibleColumns = ref([
  'id', 'name', 'description', 'domain', 'status', 
  'requirements', 'cycles', 'dates', 'actions'
])

const columnOptions = [
  { key: 'id', label: 'ID' },
  { key: 'name', label: '项目名称' },
  { key: 'description', label: '项目描述' },
  { key: 'domain', label: '项目领域' },
  { key: 'status', label: '状态' },
  { key: 'requirements', label: '需求统计' },
  { key: 'cycles', label: '当前周期' },
  { key: 'dates', label: '时间信息' },
  { key: 'actions', label: '操作' }
]

// 计算属性
const isIndeterminate = computed(() => {
  const count = selectedRows.value.length
  return count > 0 && count < props.tableData.length
})

// 方法
const handleSelectionChange = (rows) => {
  selectedRows.value = rows
  selectAll.value = rows.length === props.tableData.length
  emit('selection-change', rows)
}

const handleSelectAll = (checked) => {
  if (checked) {
    tableRef.value.toggleAllSelection()
  } else {
    tableRef.value.clearSelection()
  }
}

const checkSelectable = (row) => {
  return row.status !== 'archived'
}

const handleSortChange = (sortInfo) => {
  emit('sort-change', sortInfo)
}

const handleSizeChange = (size) => {
  emit('size-change', size)
}

const handleCurrentChange = (page) => {
  emit('current-change', page)
}

const handleRowClick = (row) => {
  // 单击行时的操作，可以扩展
}

const handleAction = (action, row) => {
  emit('action', { action, row })
}

const getRowClassName = ({ row }) => {
  const classes = []
  if (row.status === 'archived') classes.push('archived-row')
  if (row.status === 'completed') classes.push('completed-row')
  return classes.join(' ')
}

const getDomainTagType = (domain) => {
  const typeMap = {
    'web': 'primary',
    'mobile': 'success',
    'desktop': 'warning',
    'data': 'info',
    'ai': 'danger'
  }
  return typeMap[domain] || 'info'
}

const getDomainLabel = (domain) => {
  const labelMap = {
    'web': 'Web应用',
    'mobile': '移动应用',
    'desktop': '桌面应用',
    'data': '数据分析',
    'ai': '人工智能'
  }
  return labelMap[domain] || domain
}

const getStatusType = (status) => {
  const typeMap = {
    'active': 'success',
    'paused': 'warning',
    'completed': 'info',
    'archived': 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    'active': '活跃',
    'paused': '暂停',
    'completed': '完成',
    'archived': '归档'
  }
  return labelMap[status] || status
}

const getStatusIcon = (status) => {
  const iconMap = {
    'active': VideoPlay,
    'paused': Clock,
    'completed': Check,
    'archived': CircleClose
  }
  return iconMap[status] || Clock
}

const getActiveCycleName = (row) => {
  // 根据activeCycleId获取周期名称的逻辑
  for (const cycle of row.cycles) {
    console.log(cycle.ID, row.activeCycleId)
    if (cycle.ID === row.activeCycleId) {
      return cycle.name
    }
  }
  return  `周期${row.activeCycleId}`
}

const formatDate = (date) => {
  if (!date) return '未设置'
  return new Date(date).toLocaleDateString()
}

const refreshTable = () => {
  emit('refresh')
}

const toggleColumnSettings = () => {
  columnSettingsVisible.value = true
}

const saveColumnSettings = () => {
  // 保存列设置到本地存储
  localStorage.setItem('project-table-columns', JSON.stringify(visibleColumns.value))
  columnSettingsVisible.value = false
  ElMessage.success('列设置已保存')
}

const exportTable = () => {
  // 导出表格数据
  emit('action', { action: 'export-table', data: selectedRows.value })
}

// 初始化时从本地存储加载列设置
const loadColumnSettings = () => {
  const saved = localStorage.getItem('project-table-columns')
  if (saved) {
    try {
      visibleColumns.value = JSON.parse(saved)
    } catch (e) {
      console.warn('加载列设置失败:', e)
    }
  }
}

// 组件挂载时加载设置
loadColumnSettings()
</script>

<style lang="scss" scoped>
.project-table {
  .table-card {
    border-radius: 8px;
    box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
  }

  .table-toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid #ebeef5;

    .toolbar-left {
      display: flex;
      align-items: center;
      gap: 12px;

      .selection-info {
        color: #606266;
        font-size: 14px;
      }
    }
  }

  .project-name-cell {
    .project-link {
      align-items: center;
      font-weight: 500;
      margin-bottom: 4px;

      .el-icon {
        margin-right: 6px;
      }
    }

    .project-tags {
      display: flex;
      gap: 4px;
      flex-wrap: wrap;

      .domain-tag {
        font-size: 11px;
      }
    }
  }

  .description-cell {
    color: #606266;
    font-size: 13px;
    line-height: 1.4;
  }

  .requirement-stats {
    .stats-main {
      display: flex;
      gap: 8px;
      margin-bottom: 4px;

      .total-count,
      .completed-count {
        display: flex;
        align-items: center;
        font-size: 12px;

        .label {
          color: #909399;
          margin-right: 4px;
        }

        .value {
          font-weight: 500;
          &.success {
            color: #67c23a;
          }
        }
      }
    }

    .detail-link {
      font-size: 12px;
    }

    .stats-detail {
      .detail-row {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        &:last-child {
          margin-bottom: 0;
        }
      }
    }
  }

  .cycle-info {
    .cycle-tag {
      display: flex;
      align-items: center;

      .el-icon {
        margin-right: 4px;
      }
    }
  }

  .no-cycle {
    color: #c0c4cc;
    font-size: 12px;
  }

  .date-info {
    .date-row {
      display: flex;
      margin-bottom: 2px;
      font-size: 12px;

      .date-label {
        color: #909399;
        width: 36px;
      }

      .date-value {
        color: #606266;
      }
    }
  }

  .action-buttons {
    display: flex;
    gap: 8px;
    align-items: center;

    .basic-actions {
      .el-button {
        padding: 4px 8px;
      }
    }

    .action-dropdown {
      .el-button {
        padding: 4px 12px;
      }
    }
  }

  .pagination-wrapper {
    margin-top: 20px;
    display: flex;
    justify-content: center;
  }

  .column-setting-item {
    margin-bottom: 12px;
  }

  // 表格行样式
  :deep(.el-table) {
    .archived-row {
      background-color: #fafafa;
      color: #c0c4cc;
    }

    .completed-row {
      background-color: #f0f9ff;
    }
  }

  // 下拉菜单危险项样式
  :deep(.el-dropdown-menu) {
    .danger-item {
      color: #f56c6c;

      &:hover {
        background-color: #fef0f0;
        color: #f56c6c;
      }
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .project-table {
    .action-buttons {
      flex-direction: column;
      gap: 4px;
    }
  }
}
</style>