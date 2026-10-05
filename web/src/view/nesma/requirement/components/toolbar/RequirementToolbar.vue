<template>
  <div class="requirement-toolbar">
    <!-- 项目和周期选择区域 -->
    <div class="project-selection-section">
      <el-card class="selection-card" :body-style="{ padding: '16px' }">
        <div class="selection-content">
          <div class="selection-item">
            <label class="selection-label">选择项目:</label>
            <el-select
              v-model="selectedProjectId"
              placeholder="请选择项目"
              filterable
              clearable
              style="width: 300px"
              @change="handleProjectChange"
            >
              <el-option
                v-for="project in projects"
                :key="project.ID"
                :label="project.name"
                :value="project.ID"
              >
                <div class="project-option">
                  <span class="project-name">{{ project.name }}</span>
                  <el-tag size="small" :type="getProjectStatusType(project.status)">
                    {{ getProjectStatusLabel(project.status) }}
                  </el-tag>
                </div>
              </el-option>
            </el-select>
          </div>

          <div class="selection-item" v-if="selectedProjectId">
            <label class="selection-label">选择周期:</label>
            <el-select
              v-model="selectedCycleId"
              placeholder="请选择周期"
              style="width: 200px"
              @change="handleCycleChange"
            >
              <el-option
                v-for="cycle in cycles"
                :key="cycle.id"
                :label="cycle.name"
                :value="cycle.id"
              >
                <div class="cycle-option">
                  <span>{{ cycle.name }}</span>
                  <el-tag size="small" type="info">{{ cycle.status }}</el-tag>
                </div>
              </el-option>
            </el-select>
          </div>

          <div class="selection-item" v-if="selectedCycleId">
            <label class="selection-label">版本:</label>
            <el-select
              v-model="selectedVersionId"
              placeholder="选择版本"
              style="width: 180px"
              @change="handleVersionChange"
            >
              <el-option
                v-for="version in versions"
                :key="version.id"
                :label="version.version"
                :value="version.id"
              >
                <div class="version-option">
                  <span>{{ version.version }}</span>
                  <el-tag size="small" :type="getVersionType(version.versionType)">
                    {{ getVersionLabel(version.versionType) }}
                  </el-tag>
                </div>
              </el-option>
            </el-select>
            
            <el-button 
              size="small" 
              type="primary" 
              @click="handleVersionManagement"
              style="margin-left: 8px"
            >
              <el-icon><Setting /></el-icon>
              版本管理
            </el-button>
          </div>
        </div>
        
        <!-- 快速操作按钮 -->
        <div class="quick-actions" v-if="selectedProjectId">
          <el-button size="small" @click="handleQuickAnalysis" :disabled="!selectedCycleId">
            <el-icon><MagicStick /></el-icon>
            快速分析
          </el-button>
          <el-button size="small" @click="handleCreateVersion">
            <el-icon><Plus /></el-icon>
            新建版本
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- 主工具栏 -->
    <div class="main-toolbar">
      <div class="toolbar-left">
        <el-button 
          type="primary" 
          @click="$emit('create')"
          :disabled="!selectedCycleId"
        >
          <el-icon><Plus /></el-icon>
          新增需求
        </el-button>
        
        <el-button 
          @click="$emit('batch-import')"
          :disabled="!selectedCycleId"
        >
          <el-icon><Upload /></el-icon>
          批量导入
        </el-button>
        
        <el-button 
          type="danger" 
          :disabled="!selectedCount"
          @click="$emit('batch-delete')"
        >
          <el-icon><Delete /></el-icon>
          批量删除 {{ selectedCount > 0 ? `(${selectedCount})` : '' }}
        </el-button>
        
        <el-dropdown @command="handleBatchAction" :disabled="!selectedCount">
          <el-button>
            批量操作
            <el-icon><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="batch-edit">批量编辑</el-dropdown-item>
              <el-dropdown-item command="batch-move">批量移动</el-dropdown-item>
              <el-dropdown-item command="batch-export">批量导出</el-dropdown-item>
              <el-dropdown-item command="batch-analysis">批量分析</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>

      <div class="toolbar-right">
        <!-- 视图切换 -->
        <el-radio-group v-model="viewMode" size="small" @change="handleViewModeChange">
          <el-radio-button value="tree">
            <el-icon><Connection /></el-icon>
            树形视图
          </el-radio-button>
          <el-radio-button value="table">
            <el-icon><Grid /></el-icon>
            表格视图
          </el-radio-button>
          <el-radio-button value="kanban">
            <el-icon><Menu /></el-icon>
            看板视图
          </el-radio-button>
        </el-radio-group>

        <!-- 搜索和筛选 -->
        <div class="search-controls">
          <el-input
            v-model="searchForm.keyword"
            placeholder="搜索需求标题或描述"
            clearable
            class="search-input"
            @keyup.enter="handleSearch"
            @clear="handleSearch"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          
          <el-select
            v-model="searchForm.level"
            placeholder="需求层级"
            clearable
            class="filter-select"
            @change="handleSearch"
          >
            <el-option label="全部层级" value="" />
            <el-option label="L1 - 业务需求" value="1" />
            <el-option label="L2 - 用户需求" value="2" />
            <el-option label="L3 - 功能需求" value="3" />
            <el-option label="L4 - 功能点" value="4" />
          </el-select>
          
          <el-select
            v-model="searchForm.functionType"
            placeholder="功能类型"
            clearable
            class="filter-select"
            @change="handleSearch"
          >
            <el-option label="全部类型" value="" />
            <el-option label="EI - 外部输入" value="EI" />
            <el-option label="EO - 外部输出" value="EO" />
            <el-option label="EQ - 外部查询" value="EQ" />
            <el-option label="ILF - 内部逻辑文件" value="ILF" />
            <el-option label="EIF - 外部接口文件" value="EIF" />
          </el-select>
          
          <el-button-group size="small">
            <el-button @click="handleSearch">
              <el-icon><Search /></el-icon>
              搜索
            </el-button>
            <el-button @click="handleResetSearch">
              <el-icon><RefreshLeft /></el-icon>
              重置
            </el-button>
            <el-button @click="toggleAdvancedFilter">
              <el-icon><Filter /></el-icon>
              {{ showAdvancedFilter ? '收起' : '高级' }}
            </el-button>
          </el-button-group>
        </div>
      </div>
    </div>

    <!-- 高级筛选面板 -->
    <div v-show="showAdvancedFilter" class="advanced-filter">
      <el-card class="filter-card" :body-style="{ padding: '16px' }">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-form-item label="创建时间">
              <el-date-picker
                v-model="searchForm.dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                size="small"
                @change="handleSearch"
              />
            </el-form-item>
          </el-col>
          
          <el-col :span="6">
            <el-form-item label="功能点范围">
              <el-input-number
                v-model="searchForm.minFunctionPoints"
                placeholder="最小值"
                :min="0"
                size="small"
                style="width: 48%"
                @change="handleSearch"
              />
              <span style="margin: 0 4px">-</span>
              <el-input-number
                v-model="searchForm.maxFunctionPoints"
                placeholder="最大值"
                :min="0"
                size="small"
                style="width: 48%"
                @change="handleSearch"
              />
            </el-form-item>
          </el-col>
          
          <el-col :span="6">
            <el-form-item label="复杂度">
              <el-checkbox-group v-model="searchForm.complexity" @change="handleSearch">
                <el-checkbox value="Low">低</el-checkbox>
                <el-checkbox value="Average">中</el-checkbox>
                <el-checkbox value="High">高</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-col>
          
          <el-col :span="6">
            <el-form-item label="分析状态">
              <el-checkbox-group v-model="searchForm.analysisStatus" @change="handleSearch">
                <el-checkbox value="original">原始</el-checkbox>
                <el-checkbox value="analyzed">已分析</el-checkbox>
                <el-checkbox value="optimized">已优化</el-checkbox>
              </el-checkbox-group>
            </el-form-item>
          </el-col>
        </el-row>
      </el-card>
    </div>

    <!-- 统计信息栏 -->
    <div class="stats-bar" v-if="selectedCycleId">
      <el-row :gutter="20">
        <el-col :span="4">
          <div class="stat-item">
            <span class="stat-label">总需求数:</span>
            <span class="stat-value">{{ stats.total || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="stat-item">
            <span class="stat-label">L1需求:</span>
            <span class="stat-value level1">{{ stats.level1 || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="stat-item">
            <span class="stat-label">L2需求:</span>
            <span class="stat-value level2">{{ stats.level2 || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="stat-item">
            <span class="stat-label">L3需求:</span>
            <span class="stat-value level3">{{ stats.level3 || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="stat-item">
            <span class="stat-label">功能点:</span>
            <span class="stat-value level4">{{ stats.level4 || 0 }}</span>
          </div>
        </el-col>
        <el-col :span="4">
          <div class="stat-item">
            <span class="stat-label">总FP:</span>
            <span class="stat-value fp">{{ stats.totalFunctionPoints || 0 }}</span>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Plus, Upload, Delete, ArrowDown, Setting, MagicStick,
  Connection, Grid, Menu, Search, RefreshLeft, Filter
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  projects: {
    type: Array,
    default: () => []
  },
  cycles: {
    type: Array,
    default: () => []
  },
  versions: {
    type: Array,
    default: () => []
  },
  selectedCount: {
    type: Number,
    default: 0
  },
  stats: {
    type: Object,
    default: () => ({})
  },
  loading: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits([
  'create',
  'batch-import',
  'batch-delete',
  'batch-action',
  'project-change',
  'cycle-change',
  'version-change',
  'search',
  'reset',
  'view-mode-change',
  'version-management',
  'quick-analysis',
  'create-version'
])

// 响应式数据
const selectedProjectId = ref(null)
const selectedCycleId = ref(null)
const selectedVersionId = ref(null)
const viewMode = ref('tree')
const showAdvancedFilter = ref(false)

const searchForm = reactive({
  keyword: '',
  level: '',
  functionType: '',
  dateRange: null,
  minFunctionPoints: null,
  maxFunctionPoints: null,
  complexity: [],
  analysisStatus: []
})

// 计算属性
const searchParams = computed(() => {
  const params = { ...searchForm }
  
  // 处理日期范围
  if (params.dateRange && params.dateRange.length === 2) {
    params.startDate = params.dateRange[0]
    params.endDate = params.dateRange[1]
  }
  delete params.dateRange
  
  // 过滤空值
  Object.keys(params).forEach(key => {
    if (params[key] === '' || params[key] === null || 
        (Array.isArray(params[key]) && params[key].length === 0)) {
      delete params[key]
    }
  })
  
  return {
    ...params,
    projectId: selectedProjectId.value,
    cycleId: selectedCycleId.value,
    versionId: selectedVersionId.value
  }
})

// 方法
const handleProjectChange = (projectId) => {
  selectedCycleId.value = null
  selectedVersionId.value = null
  emit('project-change', projectId)
}

const handleCycleChange = (cycleId) => {
  selectedVersionId.value = null
  emit('cycle-change', cycleId)
}

const handleVersionChange = (versionId) => {
  emit('version-change', versionId)
}

const handleViewModeChange = (mode) => {
  emit('view-mode-change', mode)
}

const handleSearch = () => {
  emit('search', searchParams.value)
}

const handleResetSearch = () => {
  Object.assign(searchForm, {
    keyword: '',
    level: '',
    functionType: '',
    dateRange: null,
    minFunctionPoints: null,
    maxFunctionPoints: null,
    complexity: [],
    analysisStatus: []
  })
  showAdvancedFilter.value = false
  emit('reset')
}

const toggleAdvancedFilter = () => {
  showAdvancedFilter.value = !showAdvancedFilter.value
}

const handleBatchAction = (action) => {
  emit('batch-action', action)
}

const handleVersionManagement = () => {
  emit('version-management')
}

const handleQuickAnalysis = () => {
  emit('quick-analysis')
}

const handleCreateVersion = () => {
  emit('create-version')
}

const getProjectStatusType = (status) => {
  const typeMap = {
    'active': 'success',
    'paused': 'warning',
    'completed': 'info',
    'archived': 'danger'
  }
  return typeMap[status] || 'info'
}

const getProjectStatusLabel = (status) => {
  const labelMap = {
    'active': '活跃',
    'paused': '暂停',
    'completed': '完成',
    'archived': '归档'
  }
  return labelMap[status] || status
}

const getVersionType = (versionType) => {
  const typeMap = {
    'initial': 'info',
    'analyzed': 'primary',
    'optimized': 'success'
  }
  return typeMap[versionType] || 'info'
}

const getVersionLabel = (versionType) => {
  const labelMap = {
    'initial': '初始版本',
    'analyzed': '分析版本',
    'optimized': '优化版本'
  }
  return labelMap[versionType] || versionType
}

// 监听搜索参数变化，自动搜索（防抖）
let searchTimer = null
watch(() => searchForm.keyword, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    if (searchForm.keyword !== '') {
      handleSearch()
    }
  }, 500)
})
</script>

<style lang="scss" scoped>
.requirement-toolbar {
  .project-selection-section {
    margin-bottom: 16px;

    .selection-card {
      .selection-content {
        display: flex;
        align-items: center;
        gap: 24px;
        flex-wrap: wrap;

        .selection-item {
          display: flex;
          align-items: center;
          gap: 8px;

          .selection-label {
            font-size: 14px;
            font-weight: 500;
            color: #606266;
            white-space: nowrap;
          }

          .project-option,
          .cycle-option,
          .version-option {
            display: flex;
            justify-content: space-between;
            align-items: center;
            width: 100%;

            .project-name {
              flex: 1;
            }
          }
        }

        .quick-actions {
          margin-left: auto;
          display: flex;
          gap: 8px;
        }
      }
    }
  }

  .main-toolbar {
    background: white;
    border-radius: 8px;
    padding: 16px 20px;
    margin-bottom: 16px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    display: flex;
    justify-content: space-between;
    align-items: center;
    flex-wrap: wrap;
    gap: 16px;

    .toolbar-left {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }

    .toolbar-right {
      display: flex;
      align-items: center;
      gap: 16px;
      flex-wrap: wrap;

      .search-controls {
        display: flex;
        align-items: center;
        gap: 12px;
        flex-wrap: wrap;

        .search-input {
          width: 250px;
        }

        .filter-select {
          width: 140px;
        }
      }
    }
  }

  .advanced-filter {
    margin-bottom: 16px;

    .filter-card {
      background: #f8fafc;
      border: 1px dashed #d1d5db;
    }
  }

  .stats-bar {
    background: white;
    border-radius: 8px;
    padding: 12px 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

    .stat-item {
      display: flex;
      align-items: center;
      justify-content: center;

      .stat-label {
        font-size: 12px;
        color: #909399;
        margin-right: 4px;
      }

      .stat-value {
        font-size: 16px;
        font-weight: 600;
        color: #303133;

        &.level1 { color: #f56c6c; }
        &.level2 { color: #e6a23c; }
        &.level3 { color: #409eff; }
        &.level4 { color: #67c23a; }
        &.fp { color: #9c27b0; }
      }
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .requirement-toolbar {
    .main-toolbar {
      flex-direction: column;
      align-items: stretch;

      .toolbar-left,
      .toolbar-right {
        justify-content: center;
      }
    }

    .project-selection-section {
      .selection-content {
        flex-direction: column;
        align-items: stretch;

        .selection-item {
          justify-content: space-between;
        }

        .quick-actions {
          margin-left: 0;
          justify-content: center;
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .requirement-toolbar {
    .toolbar-right {
      .search-controls {
        flex-direction: column;
        align-items: stretch;

        .search-input,
        .filter-select {
          width: 100%;
        }
      }
    }

    .stats-bar {
      .el-row {
        .el-col {
          margin-bottom: 8px;
        }
      }

      .stat-item {
        justify-content: flex-start;
      }
    }
  }
}
</style>