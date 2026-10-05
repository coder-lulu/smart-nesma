<template>
  <div class="project-selector">
    <el-card shadow="hover" class="selector-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon class="header-icon"><Folder /></el-icon>
            <span class="header-title">项目配置</span>
          </div>
          <div class="header-right">
            <el-badge :value="getSelectionSummary()" type="info" class="selection-badge">
              <el-button text size="small">当前选择</el-button>
            </el-badge>
            <el-button 
              text 
              size="small" 
              @click="$emit('refresh-data')"
              :icon="Refresh"
              title="刷新数据"
            />
          </div>
        </div>
      </template>
      
      <el-form :model="formData" label-position="top" class="selector-form">
        <el-row :gutter="20">
          <!-- 项目选择 -->
          <el-col :span="6">
            <el-form-item label="选择项目" class="form-item">
              <el-select 
                v-model="formData.selectedProject" 
                placeholder="请选择项目"
                @change="handleProjectChange"
                :loading="loading.projects"
                filterable
                style="width: 100%"
                size="large"
              >
                <el-option 
                  v-for="project in projects" 
                  :key="project.id" 
                  :label="project.name" 
                  :value="project.id"
                >
                  <div class="project-option">
                    <div class="project-main">
                      <span class="project-name">{{ project.name }}</span>
                      <el-tag size="small" :type="getProjectStatusType(project.status)">
                        {{ project.status }}
                      </el-tag>
                    </div>
                    <div class="project-meta">
                      <span class="project-domain">{{ project.domain }}</span>
                      <span class="project-requirements" v-if="project.requirementCount">
                        {{ project.requirementCount }} 个需求
                      </span>
                    </div>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          
          <!-- 周期选择 -->
          <el-col :span="6">
            <el-form-item label="选择周期" class="form-item">
              <el-select 
                v-model="formData.selectedCycle" 
                placeholder="请选择周期"
                @change="handleCycleChange"
                :disabled="!formData.selectedProject"
                :loading="loading.cycles"
                filterable
                style="width: 100%"
                size="large"
              >
                <el-option 
                  v-for="cycle in cycles" 
                  :key="cycle.id" 
                  :label="cycle.name" 
                  :value="cycle.id"
                >
                  <div class="cycle-option">
                    <div class="cycle-main">
                      <span class="cycle-name">{{ cycle.name }}</span>
                      <el-tag size="small" :type="getCycleStatusType(cycle.status)">
                        {{ cycle.status }}
                      </el-tag>
                    </div>
                    <div class="cycle-meta">
                      <span class="cycle-description">{{ cycle.description }}</span>
                      <span class="cycle-date" v-if="cycle.created_at">
                        {{ formatDate(cycle.created_at) }}
                      </span>
                    </div>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          
          <!-- 版本选择 -->
          <el-col :span="6">
            <el-form-item label="当前版本" class="form-item">
              <el-select 
                v-model="formData.selectedVersion" 
                placeholder="选择版本"
                @change="handleVersionChange"
                :disabled="!formData.selectedCycle"
                :loading="loading.versions"
                style="width: 100%"
                size="large"
              >
                <el-option 
                  v-for="version in versions" 
                  :key="version.id" 
                  :label="version.version" 
                  :value="version.id"
                >
                  <div class="version-option">
                    <div class="version-main">
                      <span class="version-name">{{ version.version }}</span>
                      <el-tag size="small" :type="getVersionTypeTag(version.version_type)">
                        {{ getVersionTypeLabel(version.version_type) }}
                      </el-tag>
                    </div>
                    <div class="version-meta">
                      <span class="version-summary">{{ version.summary }}</span>
                      <span class="version-date" v-if="version.created_at">
                        {{ formatDate(version.created_at) }}
                      </span>
                    </div>
                  </div>
                </el-option>
              </el-select>
            </el-form-item>
          </el-col>
          
          <!-- 智能过滤 -->
          <el-col :span="6">
            <el-form-item label="智能过滤" class="form-item">
              <div class="filter-controls">
                <div class="filter-item">
                  <el-switch 
                    v-model="formData.smartFilter.skipCompleted"
                    active-text="跳过已完成"
                    @change="handleFilterChange"
                    size="large"
                  />
                </div>
                <div class="filter-item">
                  <el-switch 
                    v-model="formData.smartFilter.onlyUnanalyzed"
                    active-text="仅未分析"
                    @change="handleFilterChange"
                    size="large"
                  />
                </div>
                <div class="filter-item">
                  <el-switch 
                    v-model="formData.smartFilter.highlightComplex"
                    active-text="突出复杂项"
                    @change="handleFilterChange"
                    size="large"
                  />
                </div>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        
        <!-- 高级筛选选项 -->
        <el-collapse v-if="showAdvancedFilters" class="advanced-filters">
          <el-collapse-item title="高级筛选选项" name="advanced">
            <el-row :gutter="20">
              <el-col :span="8">
                <el-form-item label="需求级别">
                  <el-checkbox-group v-model="formData.advancedFilter.levels" @change="handleFilterChange">
                    <el-checkbox :label="1">L1 - 业务功能</el-checkbox>
                    <el-checkbox :label="2">L2 - 子功能</el-checkbox>
                    <el-checkbox :label="3">L3 - 功能点</el-checkbox>
                    <el-checkbox :label="4">L4 - 细节功能</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
              </el-col>
              
              <el-col :span="8">
                <el-form-item label="功能类型">
                  <el-checkbox-group v-model="formData.advancedFilter.functionTypes" @change="handleFilterChange">
                    <el-checkbox label="EI">外部输入</el-checkbox>
                    <el-checkbox label="EO">外部输出</el-checkbox>
                    <el-checkbox label="EQ">外部查询</el-checkbox>
                    <el-checkbox label="ILF">内部逻辑文件</el-checkbox>
                    <el-checkbox label="EIF">外部接口文件</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
              </el-col>
              
              <el-col :span="8">
                <el-form-item label="复杂度">
                  <el-checkbox-group v-model="formData.advancedFilter.complexities" @change="handleFilterChange">
                    <el-checkbox label="Low">低复杂度</el-checkbox>
                    <el-checkbox label="Average">中等复杂度</el-checkbox>
                    <el-checkbox label="High">高复杂度</el-checkbox>
                  </el-checkbox-group>
                </el-form-item>
              </el-col>
            </el-row>
          </el-collapse-item>
        </el-collapse>
      </el-form>
      
      <!-- 状态提示 -->
      <div class="status-alerts" v-if="statusMessages.length">
        <el-alert 
          v-for="(msg, index) in statusMessages" 
          :key="index"
          :type="msg.type"
          :title="msg.title"
          :description="msg.description"
          show-icon
          :closable="true"
          @close="removeStatusMessage(index)"
          class="status-alert"
        />
      </div>
      
      <!-- 选择概览 -->
      <div v-if="showSelectionOverview" class="selection-overview">
        <div class="overview-header">
          <span class="overview-title">选择概览</span>
          <el-button text size="small" @click="toggleSelectionOverview">
            {{ showDetailedOverview ? '收起' : '展开' }}
          </el-button>
        </div>
        
        <div v-if="showDetailedOverview" class="overview-content">
          <div class="overview-item">
            <span class="overview-label">项目:</span>
            <span class="overview-value">{{ selectedProjectName }}</span>
          </div>
          <div class="overview-item">
            <span class="overview-label">周期:</span>
            <span class="overview-value">{{ selectedCycleName }}</span>
          </div>
          <div class="overview-item">
            <span class="overview-label">版本:</span>
            <span class="overview-value">{{ selectedVersionName }}</span>
          </div>
          <div class="overview-item">
            <span class="overview-label">筛选条件:</span>
            <span class="overview-value">{{ getFilterSummary() }}</span>
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { Folder, Refresh } from '@element-plus/icons-vue'

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
  selectedProject: {
    type: [String, Number],
    default: null
  },
  selectedCycle: {
    type: [String, Number],
    default: null
  },
  selectedVersion: {
    type: [String, Number],
    default: null
  },
  smartFilter: {
    type: Object,
    default: () => ({
      skipCompleted: false,
      onlyUnanalyzed: true,
      highlightComplex: false
    })
  },
  advancedFilter: {
    type: Object,
    default: () => ({
      levels: [3, 4],
      functionTypes: [],
      complexities: []
    })
  },
  statusMessages: {
    type: Array,
    default: () => []
  },
  loading: {
    type: Object,
    default: () => ({
      projects: false,
      cycles: false,
      versions: false
    })
  },
  showAdvancedFilters: {
    type: Boolean,
    default: false
  },
  showSelectionOverview: {
    type: Boolean,
    default: true
  }
})

// Emits
const emit = defineEmits([
  'project-change',
  'cycle-change',
  'version-change',
  'filter-change',
  'refresh-data',
  'remove-status-message'
])

// 响应式数据
const formData = ref({
  selectedProject: props.selectedProject,
  selectedCycle: props.selectedCycle,
  selectedVersion: props.selectedVersion,
  smartFilter: { ...props.smartFilter },
  advancedFilter: { ...props.advancedFilter }
})

const showDetailedOverview = ref(false)

// 计算属性
const selectedProjectName = computed(() => {
  const project = props.projects.find(p => p.id === formData.value.selectedProject)
  return project ? project.name : '未选择'
})

const selectedCycleName = computed(() => {
  const cycle = props.cycles.find(c => c.id === formData.value.selectedCycle)
  return cycle ? cycle.name : '未选择'
})

const selectedVersionName = computed(() => {
  const version = props.versions.find(v => v.id === formData.value.selectedVersion)
  return version ? version.version : '未选择'
})

// 方法
const getSelectionSummary = () => {
  const parts = []
  if (formData.value.selectedProject) parts.push('项目')
  if (formData.value.selectedCycle) parts.push('周期')
  if (formData.value.selectedVersion) parts.push('版本')
  return parts.length || '未选择'
}

const getProjectStatusType = (status) => {
  const statusMap = {
    'active': 'success',
    'completed': 'info',
    'archived': 'warning',
    'planning': 'primary'
  }
  return statusMap[status] || 'default'
}

const getCycleStatusType = (status) => {
  const statusMap = {
    'active': 'success',
    'completed': 'info',
    'planning': 'warning'
  }
  return statusMap[status] || 'default'
}

const getVersionTypeTag = (type) => {
  const typeMap = {
    'initial': 'info',
    'analyzed': 'success',
    'optimized': 'primary',
    'draft': 'warning'
  }
  return typeMap[type] || 'default'
}

const getVersionTypeLabel = (type) => {
  const labelMap = {
    'initial': '初始版本',
    'analyzed': '分析版本',
    'optimized': '优化版本',
    'draft': '草稿版本'
  }
  return labelMap[type] || type
}

const formatDate = (dateString) => {
  if (!dateString) return ''
  const date = new Date(dateString)
  return date.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const getFilterSummary = () => {
  const filters = []
  if (formData.value.smartFilter.skipCompleted) filters.push('跳过已完成')
  if (formData.value.smartFilter.onlyUnanalyzed) filters.push('仅未分析')
  if (formData.value.smartFilter.highlightComplex) filters.push('突出复杂项')
  
  const levelCount = formData.value.advancedFilter.levels.length
  if (levelCount > 0 && levelCount < 4) {
    filters.push(`L${formData.value.advancedFilter.levels.join(',L')}级别`)
  }
  
  if (formData.value.advancedFilter.functionTypes.length > 0) {
    filters.push(`${formData.value.advancedFilter.functionTypes.join(',')}类型`)
  }
  
  return filters.length ? filters.join(', ') : '无特殊筛选'
}

const handleProjectChange = (projectId) => {
  formData.value.selectedCycle = null
  formData.value.selectedVersion = null
  emit('project-change', projectId)
}

const handleCycleChange = (cycleId) => {
  formData.value.selectedVersion = null
  emit('cycle-change', cycleId)
}

const handleVersionChange = (versionId) => {
  emit('version-change', versionId)
}

const handleFilterChange = () => {
  emit('filter-change', {
    smartFilter: formData.value.smartFilter,
    advancedFilter: formData.value.advancedFilter
  })
}

const removeStatusMessage = (index) => {
  emit('remove-status-message', index)
}

const toggleSelectionOverview = () => {
  showDetailedOverview.value = !showDetailedOverview.value
}

// 监听器
watch(() => props.selectedProject, (newVal) => {
  formData.value.selectedProject = newVal
})

watch(() => props.selectedCycle, (newVal) => {
  formData.value.selectedCycle = newVal
})

watch(() => props.selectedVersion, (newVal) => {
  formData.value.selectedVersion = newVal
})

watch(() => props.smartFilter, (newVal) => {
  formData.value.smartFilter = { ...newVal }
}, { deep: true })

watch(() => props.advancedFilter, (newVal) => {
  formData.value.advancedFilter = { ...newVal }
}, { deep: true })
</script>

<style lang="scss" scoped>
.project-selector {
  margin-bottom: 20px;

  .selector-card {
    border-radius: 12px;
    overflow: hidden;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .header-left {
        display: flex;
        align-items: center;
        gap: 8px;

        .header-icon {
          color: #3b82f6;
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

    .selector-form {
      .form-item {
        margin-bottom: 0;

        :deep(.el-form-item__label) {
          font-weight: 600;
          color: #374151;
          margin-bottom: 8px;
        }

        .el-select {
          :deep(.el-input__inner) {
            border-radius: 8px;
            border: 2px solid #e5e7eb;
            transition: all 0.3s ease;

            &:focus {
              border-color: #3b82f6;
              box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
            }
          }
        }

        .filter-controls {
          display: flex;
          flex-direction: column;
          gap: 12px;

          .filter-item {
            .el-switch {
              :deep(.el-switch__label) {
                font-size: 13px;
                color: #374151;
              }
            }
          }
        }
      }
    }

    .advanced-filters {
      margin-top: 20px;

      :deep(.el-collapse-item__header) {
        background: #f8fafc;
        padding: 12px 16px;
        border-radius: 8px;
        font-weight: 600;
        color: #374151;
      }

      :deep(.el-collapse-item__content) {
        padding: 16px 0;
      }

      .el-checkbox-group {
        display: flex;
        flex-direction: column;
        gap: 8px;

        .el-checkbox {
          margin: 0;

          :deep(.el-checkbox__label) {
            font-size: 13px;
            color: #374151;
          }
        }
      }
    }

    .status-alerts {
      margin-top: 16px;

      .status-alert {
        margin-bottom: 8px;
        border-radius: 8px;

        &:last-child {
          margin-bottom: 0;
        }
      }
    }

    .selection-overview {
      margin-top: 16px;
      padding: 16px;
      background: #f8fafc;
      border-radius: 8px;
      border: 1px solid #e5e7eb;

      .overview-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .overview-title {
          font-weight: 600;
          color: #374151;
          font-size: 14px;
        }
      }

      .overview-content {
        .overview-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          padding: 4px 0;
          font-size: 13px;

          .overview-label {
            color: #6b7280;
            font-weight: 500;
          }

          .overview-value {
            color: #374151;
            font-weight: 600;
          }
        }
      }
    }
  }
}

// 选项样式
.project-option,
.cycle-option,
.version-option {
  padding: 8px 0;

  .project-main,
  .cycle-main,
  .version-main {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 4px;

    .project-name,
    .cycle-name,
    .version-name {
      font-weight: 600;
      color: #1f2937;
    }
  }

  .project-meta,
  .cycle-meta,
  .version-meta {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: #6b7280;

    .project-domain,
    .cycle-description,
    .version-summary {
      flex: 1;
      margin-right: 8px;
    }

    .project-requirements,
    .cycle-date,
    .version-date {
      font-weight: 500;
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .project-selector {
    .selector-card {
      .card-header {
        flex-direction: column;
        gap: 12px;
        align-items: flex-start;
      }

      .selector-form {
        .el-row {
          .el-col {
            margin-bottom: 16px;
          }
        }

        .filter-controls {
          .filter-item {
            .el-switch {
              :deep(.el-switch__label) {
                font-size: 12px;
              }
            }
          }
        }
      }

      .advanced-filters {
        .el-row {
          .el-col {
            margin-bottom: 16px;
          }
        }
      }
    }
  }
}
</style>