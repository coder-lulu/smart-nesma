<template>
  <div class="requirement-selector">
    <el-card shadow="hover" class="selector-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon class="header-icon"><List /></el-icon>
            <span class="header-title">需求选择</span>
          </div>
          <div class="header-right">
            <el-badge :value="selectedRequirements.length" type="primary" class="selection-badge">
              <el-button text size="small" @click="toggleSelectAll">
                {{ isAllSelected ? '取消全选' : '全选' }}
              </el-button>
            </el-badge>
            <el-dropdown trigger="click" class="action-dropdown">
              <el-button text size="small" :icon="More">操作</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="expandAll">
                    <el-icon><Plus /></el-icon>
                    展开全部
                  </el-dropdown-item>
                  <el-dropdown-item @click="collapseAll">
                    <el-icon><Minus /></el-icon>
                    收起全部
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="selectByLevel">
                    <el-icon><Grid /></el-icon>
                    按级别选择
                  </el-dropdown-item>
                  <el-dropdown-item @click="selectByType">
                    <el-icon><Filter /></el-icon>
                    按类型选择
                  </el-dropdown-item>
                  <el-dropdown-item @click="invertSelection">
                    <el-icon><Switch /></el-icon>
                    反选
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </template>
      
      <!-- 搜索和筛选工具栏 -->
      <div class="toolbar">
        <div class="search-section">
          <el-input
            v-model="searchQuery"
            placeholder="搜索需求..."
            :prefix-icon="Search"
            clearable
            @input="handleSearch"
            class="search-input"
          />
        </div>
        
        <div class="filter-section">
          <el-select
            v-model="levelFilter"
            placeholder="筛选级别"
            multiple
            collapse-tags
            @change="handleLevelFilter"
            style="width: 120px;"
            size="small"
          >
            <el-option label="L1" :value="1" />
            <el-option label="L2" :value="2" />
            <el-option label="L3" :value="3" />
            <el-option label="L4" :value="4" />
          </el-select>
          
          <el-select
            v-model="typeFilter"
            placeholder="筛选类型"
            multiple
            collapse-tags
            @change="handleTypeFilter"
            style="width: 140px;"
            size="small"
          >
            <el-option label="EI-外部输入" value="EI" />
            <el-option label="EO-外部输出" value="EO" />
            <el-option label="EQ-外部查询" value="EQ" />
            <el-option label="ILF-内部文件" value="ILF" />
            <el-option label="EIF-外部文件" value="EIF" />
          </el-select>
          
          <el-select
            v-model="statusFilter"
            placeholder="筛选状态"
            multiple
            collapse-tags
            @change="handleStatusFilter"
            style="width: 120px;"
            size="small"
          >
            <el-option label="未分析" value="pending" />
            <el-option label="已分析" value="analyzed" />
            <el-option label="已优化" value="optimized" />
            <el-option label="已完成" value="completed" />
          </el-select>
        </div>
      </div>
      
      <!-- 需求树 -->
      <div class="tree-container">
        <el-tree
          ref="requirementTreeRef"
          :data="filteredRequirements"
          :props="treeProps"
          show-checkbox
          node-key="id"
          :default-checked-keys="selectedRequirements"
          :default-expanded-keys="expandedNodes"
          @check="handleRequirementCheck"
          @node-expand="handleNodeExpand"
          @node-collapse="handleNodeCollapse"
          :highlight-current="true"
          :filter-node-method="filterNode"
          class="requirement-tree"
        >
          <template #default="{ node, data }">
            <div class="tree-node" :class="getNodeClass(data)">
              <div class="node-content">
                <div class="node-header">
                  <div class="node-title">
                    <el-tag 
                      :type="getRequirementLevelType(data.level)" 
                      size="small"
                      class="level-tag"
                    >
                      L{{ data.level }}
                    </el-tag>
                    <span class="requirement-title">{{ data.title }}</span>
                    <el-tag 
                      v-if="data.function_type"
                      :type="getFunctionTypeTag(data.function_type)"
                      size="small"
                      class="type-tag"
                    >
                      {{ data.function_type }}
                    </el-tag>
                  </div>
                  
                  <div class="node-actions">
                    <el-button 
                      text 
                      size="small" 
                      @click.stop="previewRequirement(data)"
                      :icon="View"
                      title="预览"
                    />
                    <el-button 
                      text 
                      size="small" 
                      @click.stop="editRequirement(data)"
                      :icon="Edit"
                      title="编辑"
                    />
                  </div>
                </div>
                
                <div class="node-meta" v-if="showDetails">
                  <div class="meta-row">
                    <span v-if="data.description" class="description">
                      {{ truncateText(data.description, 100) }}
                    </span>
                  </div>
                  
                  <div class="meta-row" v-if="hasAnalysisData(data)">
                    <div class="analysis-info">
                      <span v-if="data.complexity" class="complexity">
                        <el-icon><TrendCharts /></el-icon>
                        复杂度: {{ data.complexity }}
                      </span>
                      <span v-if="data.afp" class="afp">
                        <el-icon><Coin /></el-icon>
                        AFP: {{ data.afp }}
                      </span>
                      <span v-if="data.ai_confidence_score" class="confidence">
                        <el-icon><DataAnalysis /></el-icon>
                        置信度: {{ Math.round(data.ai_confidence_score * 100) }}%
                      </span>
                    </div>
                  </div>
                  
                  <div class="meta-row">
                    <div class="status-info">
                      <el-tag 
                        :type="getStatusType(data.ai_analysis_status)" 
                        size="small"
                        class="status-tag"
                      >
                        {{ getStatusLabel(data.ai_analysis_status) }}
                      </el-tag>
                      
                      <span v-if="data.updated_at" class="update-time">
                        <el-icon><Clock /></el-icon>
                        {{ formatRelativeTime(data.updated_at) }}
                      </span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </template>
        </el-tree>
        
        <!-- 空状态 -->
        <el-empty v-if="!filteredRequirements.length" description="暂无需求数据">
          <el-button type="primary" @click="$emit('refresh-requirements')">
            刷新数据
          </el-button>
        </el-empty>
        
        <!-- 加载状态 -->
        <div v-if="loading" class="loading-container">
          <el-skeleton :rows="5" animated />
        </div>
      </div>
      
      <!-- 选择统计 -->
      <div class="selection-stats" v-if="selectedRequirements.length > 0">
        <div class="stats-header">
          <span class="stats-title">选择统计</span>
          <el-button text size="small" @click="showDetailedStats = !showDetailedStats">
            {{ showDetailedStats ? '收起' : '详情' }}
          </el-button>
        </div>
        
        <div class="stats-content">
          <div class="stats-overview">
            <div class="stat-item">
              <span class="stat-value">{{ selectedRequirements.length }}</span>
              <span class="stat-label">已选择</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ getSelectedByLevel(3).length }}</span>
              <span class="stat-label">L3需求</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ getSelectedByLevel(4).length }}</span>
              <span class="stat-label">L4需求</span>
            </div>
            <div class="stat-item">
              <span class="stat-value">{{ getTotalFunctionPoints() }}</span>
              <span class="stat-label">功能点</span>
            </div>
          </div>
          
          <div v-if="showDetailedStats" class="stats-details">
            <div class="detail-row">
              <span class="detail-label">功能类型分布:</span>
              <div class="type-distribution">
                <el-tag 
                  v-for="(count, type) in getTypeDistribution()" 
                  :key="type"
                  size="small"
                  :type="getFunctionTypeTag(type)"
                >
                  {{ type }}: {{ count }}
                </el-tag>
              </div>
            </div>
            
            <div class="detail-row">
              <span class="detail-label">复杂度分布:</span>
              <div class="complexity-distribution">
                <el-tag 
                  v-for="(count, complexity) in getComplexityDistribution()" 
                  :key="complexity"
                  size="small"
                  :type="getComplexityTag(complexity)"
                >
                  {{ complexity }}: {{ count }}
                </el-tag>
              </div>
            </div>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { 
  List, More, Plus, Minus, Grid, Filter, Switch, Search, 
  View, Edit, TrendCharts, Coin, DataAnalysis, Clock 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  requirements: {
    type: Array,
    default: () => []
  },
  selectedRequirements: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Boolean,
    default: false
  },
  showDetails: {
    type: Boolean,
    default: true
  },
  treeProps: {
    type: Object,
    default: () => ({
      children: 'children',
      label: 'title'
    })
  }
})

// Emits
const emit = defineEmits([
  'selection-change',
  'preview-requirement',
  'edit-requirement',
  'refresh-requirements'
])

// 响应式数据
const requirementTreeRef = ref()
const searchQuery = ref('')
const levelFilter = ref([])
const typeFilter = ref([])
const statusFilter = ref([])
const expandedNodes = ref([])
const showDetailedStats = ref(false)

// 计算属性
const filteredRequirements = computed(() => {
  let filtered = props.requirements
  
  // 级别筛选
  if (levelFilter.value.length > 0) {
    filtered = filterByLevel(filtered, levelFilter.value)
  }
  
  // 类型筛选
  if (typeFilter.value.length > 0) {
    filtered = filterByType(filtered, typeFilter.value)
  }
  
  // 状态筛选
  if (statusFilter.value.length > 0) {
    filtered = filterByStatus(filtered, statusFilter.value)
  }
  
  return filtered
})

const isAllSelected = computed(() => {
  return props.selectedRequirements.length === getAllSelectableIds(props.requirements).length
})

// 方法
const filterByLevel = (data, levels) => {
  return data.map(item => {
    const newItem = { ...item }
    if (levels.includes(item.level)) {
      if (item.children) {
        newItem.children = filterByLevel(item.children, levels)
      }
      return newItem
    } else if (item.children) {
      const filteredChildren = filterByLevel(item.children, levels)
      if (filteredChildren.length > 0) {
        newItem.children = filteredChildren
        return newItem
      }
    }
    return null
  }).filter(Boolean)
}

const filterByType = (data, types) => {
  return data.map(item => {
    const newItem = { ...item }
    if (types.includes(item.function_type)) {
      if (item.children) {
        newItem.children = filterByType(item.children, types)
      }
      return newItem
    } else if (item.children) {
      const filteredChildren = filterByType(item.children, types)
      if (filteredChildren.length > 0) {
        newItem.children = filteredChildren
        return newItem
      }
    }
    return null
  }).filter(Boolean)
}

const filterByStatus = (data, statuses) => {
  return data.map(item => {
    const newItem = { ...item }
    if (statuses.includes(item.ai_analysis_status)) {
      if (item.children) {
        newItem.children = filterByStatus(item.children, statuses)
      }
      return newItem
    } else if (item.children) {
      const filteredChildren = filterByStatus(item.children, statuses)
      if (filteredChildren.length > 0) {
        newItem.children = filteredChildren
        return newItem
      }
    }
    return null
  }).filter(Boolean)
}

const getAllSelectableIds = (data) => {
  const ids = []
  const traverse = (items) => {
    items.forEach(item => {
      if (item.level >= 3) { // 只有L3和L4可选
        ids.push(item.id)
      }
      if (item.children) {
        traverse(item.children)
      }
    })
  }
  traverse(data)
  return ids
}

const getSelectedByLevel = (level) => {
  return props.selectedRequirements.filter(id => {
    const requirement = findRequirementById(props.requirements, id)
    return requirement && requirement.level === level
  })
}

const findRequirementById = (data, id) => {
  for (const item of data) {
    if (item.id === id) return item
    if (item.children) {
      const found = findRequirementById(item.children, id)
      if (found) return found
    }
  }
  return null
}

const getTotalFunctionPoints = () => {
  let total = 0
  props.selectedRequirements.forEach(id => {
    const requirement = findRequirementById(props.requirements, id)
    if (requirement && requirement.afp) {
      total += requirement.afp
    }
  })
  return total.toFixed(1)
}

const getTypeDistribution = () => {
  const distribution = {}
  props.selectedRequirements.forEach(id => {
    const requirement = findRequirementById(props.requirements, id)
    if (requirement && requirement.function_type) {
      distribution[requirement.function_type] = (distribution[requirement.function_type] || 0) + 1
    }
  })
  return distribution
}

const getComplexityDistribution = () => {
  const distribution = {}
  props.selectedRequirements.forEach(id => {
    const requirement = findRequirementById(props.requirements, id)
    if (requirement && requirement.complexity) {
      distribution[requirement.complexity] = (distribution[requirement.complexity] || 0) + 1
    }
  })
  return distribution
}

const getNodeClass = (data) => {
  return [
    `level-${data.level}`,
    data.ai_analysis_status,
    { 'has-analysis': hasAnalysisData(data) }
  ]
}

const getRequirementLevelType = (level) => {
  const typeMap = {
    1: 'danger',
    2: 'warning', 
    3: 'primary',
    4: 'success'
  }
  return typeMap[level] || 'info'
}

const getFunctionTypeTag = (type) => {
  const typeMap = {
    'EI': 'primary',
    'EO': 'success',
    'EQ': 'info',
    'ILF': 'warning',
    'EIF': 'danger'
  }
  return typeMap[type] || 'default'
}

const getComplexityTag = (complexity) => {
  const tagMap = {
    'Low': 'success',
    'Average': 'warning',
    'High': 'danger'
  }
  return tagMap[complexity] || 'info'
}

const getStatusType = (status) => {
  const statusMap = {
    'pending': 'info',
    'analyzing': 'warning',
    'analyzed': 'success',
    'optimized': 'primary',
    'completed': 'success',
    'failed': 'danger'
  }
  return statusMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    'pending': '未分析',
    'analyzing': '分析中',
    'analyzed': '已分析',
    'optimized': '已优化',
    'completed': '已完成',
    'failed': '分析失败'
  }
  return labelMap[status] || status
}

const hasAnalysisData = (data) => {
  return data.complexity || data.afp || data.ai_confidence_score
}

const truncateText = (text, length) => {
  if (!text) return ''
  return text.length > length ? text.substring(0, length) + '...' : text
}

const formatRelativeTime = (time) => {
  if (!time) return ''
  const now = new Date()
  const date = new Date(time)
  const diffInSeconds = Math.floor((now - date) / 1000)
  
  if (diffInSeconds < 60) return '刚刚'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}分钟前`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}小时前`
  
  const diffInDays = Math.floor(diffInSeconds / 86400)
  if (diffInDays < 7) return `${diffInDays}天前`
  
  return date.toLocaleDateString('zh-CN')
}

const filterNode = (value, data) => {
  if (!value) return true
  return data.title.toLowerCase().includes(value.toLowerCase()) ||
         (data.description && data.description.toLowerCase().includes(value.toLowerCase()))
}

const handleRequirementCheck = (data, checked) => {
  emit('selection-change', {
    requirement: data,
    checked: checked.checkedKeys,
    halfChecked: checked.halfCheckedKeys
  })
}

const handleNodeExpand = (data) => {
  if (!expandedNodes.value.includes(data.id)) {
    expandedNodes.value.push(data.id)
  }
}

const handleNodeCollapse = (data) => {
  const index = expandedNodes.value.indexOf(data.id)
  if (index > -1) {
    expandedNodes.value.splice(index, 1)
  }
}

const handleSearch = () => {
  requirementTreeRef.value.filter(searchQuery.value)
}

const handleLevelFilter = () => {
  // 级别筛选逻辑已在computed中处理
}

const handleTypeFilter = () => {
  // 类型筛选逻辑已在computed中处理
}

const handleStatusFilter = () => {
  // 状态筛选逻辑已在computed中处理
}

const toggleSelectAll = () => {
  const allIds = getAllSelectableIds(props.requirements)
  if (isAllSelected.value) {
    emit('selection-change', {
      checked: [],
      halfChecked: []
    })
  } else {
    emit('selection-change', {
      checked: allIds,
      halfChecked: []
    })
  }
}

const expandAll = () => {
  const allIds = []
  const traverse = (items) => {
    items.forEach(item => {
      allIds.push(item.id)
      if (item.children) {
        traverse(item.children)
      }
    })
  }
  traverse(props.requirements)
  expandedNodes.value = allIds
}

const collapseAll = () => {
  expandedNodes.value = []
}

const selectByLevel = () => {
  ElMessageBox.prompt('请输入要选择的级别 (1-4)', '按级别选择', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputPattern: /^[1-4]$/,
    inputErrorMessage: '请输入有效的级别 (1-4)'
  }).then(({ value }) => {
    const level = parseInt(value)
    const ids = []
    const traverse = (items) => {
      items.forEach(item => {
        if (item.level === level && item.level >= 3) {
          ids.push(item.id)
        }
        if (item.children) {
          traverse(item.children)
        }
      })
    }
    traverse(props.requirements)
    emit('selection-change', {
      checked: ids,
      halfChecked: []
    })
  }).catch(() => {})
}

const selectByType = () => {
  ElMessageBox.prompt('请输入要选择的功能类型 (EI/EO/EQ/ILF/EIF)', '按类型选择', {
    confirmButtonText: '确定',
    cancelButtonText: '取消'
  }).then(({ value }) => {
    const type = value.toUpperCase()
    const ids = []
    const traverse = (items) => {
      items.forEach(item => {
        if (item.function_type === type && item.level >= 3) {
          ids.push(item.id)
        }
        if (item.children) {
          traverse(item.children)
        }
      })
    }
    traverse(props.requirements)
    emit('selection-change', {
      checked: ids,
      halfChecked: []
    })
  }).catch(() => {})
}

const invertSelection = () => {
  const allIds = getAllSelectableIds(props.requirements)
  const newSelection = allIds.filter(id => !props.selectedRequirements.includes(id))
  emit('selection-change', {
    checked: newSelection,
    halfChecked: []
  })
}

const previewRequirement = (data) => {
  emit('preview-requirement', data)
}

const editRequirement = (data) => {
  emit('edit-requirement', data)
}

// 监听器
watch(searchQuery, () => {
  handleSearch()
})
</script>

<style lang="scss" scoped>
.requirement-selector {
  .selector-card {
    border-radius: 12px;
    height: 600px;
    display: flex;
    flex-direction: column;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .header-left {
        display: flex;
        align-items: center;
        gap: 8px;

        .header-icon {
          color: #10b981;
          font-size: 18px;
        }

        .header-title {
          font-weight: 600;
          color: #1f2937;
        }
      }

      .header-right {
        display: flex;
        align-items: center;
        gap: 8px;

        .selection-badge {
          .el-button {
            font-size: 12px;
          }
        }
      }
    }

    :deep(.el-card__body) {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
    }

    .toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      gap: 16px;

      .search-section {
        flex: 1;

        .search-input {
          max-width: 300px;
        }
      }

      .filter-section {
        display: flex;
        gap: 8px;
      }
    }

    .tree-container {
      flex: 1;
      overflow: auto;
      border: 1px solid #e5e7eb;
      border-radius: 8px;
      padding: 8px;

      .requirement-tree {
        .tree-node {
          width: 100%;
          padding: 8px 0;
          border-bottom: 1px solid #f3f4f6;

          &:last-child {
            border-bottom: none;
          }

          &.level-1 {
            background: rgba(239, 68, 68, 0.05);
          }

          &.level-2 {
            background: rgba(245, 158, 11, 0.05);
          }

          &.level-3 {
            background: rgba(59, 130, 246, 0.05);
          }

          &.level-4 {
            background: rgba(16, 185, 129, 0.05);
          }

          &.analyzed {
            border-left: 3px solid #10b981;
          }

          &.optimized {
            border-left: 3px solid #3b82f6;
          }

          .node-content {
            .node-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 4px;

              .node-title {
                flex: 1;
                display: flex;
                align-items: center;
                gap: 8px;

                .level-tag {
                  font-weight: 600;
                }

                .requirement-title {
                  font-weight: 500;
                  color: #1f2937;
                  line-height: 1.4;
                }

                .type-tag {
                  margin-left: auto;
                }
              }

              .node-actions {
                opacity: 0;
                transition: opacity 0.2s ease;

                .el-button {
                  padding: 4px;
                  margin-left: 4px;
                }
              }
            }

            .node-meta {
              font-size: 12px;
              color: #6b7280;

              .meta-row {
                margin-bottom: 4px;

                &:last-child {
                  margin-bottom: 0;
                }

                .description {
                  line-height: 1.4;
                }

                .analysis-info {
                  display: flex;
                  gap: 12px;

                  span {
                    display: flex;
                    align-items: center;
                    gap: 4px;
                  }

                  .complexity {
                    color: #f59e0b;
                  }

                  .afp {
                    color: #10b981;
                  }

                  .confidence {
                    color: #3b82f6;
                  }
                }

                .status-info {
                  display: flex;
                  justify-content: space-between;
                  align-items: center;

                  .update-time {
                    display: flex;
                    align-items: center;
                    gap: 4px;
                  }
                }
              }
            }

            &:hover .node-actions {
              opacity: 1;
            }
          }
        }
      }

      .loading-container {
        padding: 20px;
      }
    }

    .selection-stats {
      margin-top: 16px;
      padding: 16px;
      background: #f8fafc;
      border-radius: 8px;
      border: 1px solid #e5e7eb;

      .stats-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;

        .stats-title {
          font-weight: 600;
          color: #374151;
        }
      }

      .stats-content {
        .stats-overview {
          display: flex;
          justify-content: space-around;
          margin-bottom: 12px;

          .stat-item {
            text-align: center;

            .stat-value {
              display: block;
              font-size: 18px;
              font-weight: 700;
              color: #1f2937;
              margin-bottom: 2px;
            }

            .stat-label {
              font-size: 12px;
              color: #6b7280;
              text-transform: uppercase;
              font-weight: 500;
            }
          }
        }

        .stats-details {
          .detail-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 8px;
            font-size: 12px;

            &:last-child {
              margin-bottom: 0;
            }

            .detail-label {
              color: #6b7280;
              font-weight: 500;
              min-width: 80px;
            }

            .type-distribution,
            .complexity-distribution {
              display: flex;
              gap: 4px;
              flex-wrap: wrap;
            }
          }
        }
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .requirement-selector {
    .selector-card {
      height: 500px;

      .toolbar {
        flex-direction: column;
        gap: 12px;

        .search-section {
          width: 100%;

          .search-input {
            max-width: none;
          }
        }

        .filter-section {
          width: 100%;
          justify-content: space-between;

          .el-select {
            width: auto !important;
            min-width: 80px;
          }
        }
      }

      .tree-container {
        .requirement-tree {
          .tree-node {
            .node-content {
              .node-header {
                flex-direction: column;
                align-items: flex-start;
                gap: 8px;

                .node-title {
                  width: 100%;
                }

                .node-actions {
                  opacity: 1;
                  align-self: flex-end;
                }
              }
            }
          }
        }
      }

      .selection-stats {
        .stats-content {
          .stats-overview {
            flex-wrap: wrap;
            gap: 16px;

            .stat-item {
              min-width: 80px;
            }
          }

          .stats-details {
            .detail-row {
              flex-direction: column;
              align-items: flex-start;
              gap: 4px;
            }
          }
        }
      }
    }
  }
}
</style>