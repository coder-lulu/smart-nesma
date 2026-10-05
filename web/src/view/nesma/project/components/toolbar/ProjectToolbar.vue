<template>
  <div class="project-toolbar">
    <!-- 左侧操作按钮 -->
    <div class="toolbar-left">
      <el-button type="primary" @click="$emit('create')">
        <el-icon><Plus /></el-icon>
        新建项目
      </el-button>
      <el-button @click="$emit('batch-import')">
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
      <el-button @click="$emit('export')">
        <el-icon><Download /></el-icon>
        导出数据
      </el-button>
    </div>

    <!-- 右侧搜索筛选 -->
    <div class="toolbar-right">
      <div class="search-controls">
        <el-input
          v-model="searchForm.name"
          placeholder="搜索项目名称"
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
          v-model="searchForm.status"
          placeholder="项目状态"
          clearable
          class="filter-select"
          @change="handleSearch"
        >
          <el-option label="全部状态" value="" />
          <el-option label="活跃" value="active" />
          <el-option label="暂停" value="paused" />
          <el-option label="完成" value="completed" />
          <el-option label="归档" value="archived" />
        </el-select>
        
        <el-select
          v-model="searchForm.domain"
          placeholder="项目领域"
          clearable
          class="filter-select"
          @change="handleSearch"
        >
          <el-option label="全部领域" value="" />
          <el-option label="Web应用" value="web" />
          <el-option label="移动应用" value="mobile" />
          <el-option label="桌面应用" value="desktop" />
          <el-option label="数据分析" value="data" />
          <el-option label="人工智能" value="ai" />
          <el-option label="企业服务" value="enterprise" />
          <el-option label="游戏开发" value="game" />
        </el-select>
        
        <el-date-picker
          v-model="searchForm.dateRange"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          class="date-picker"
          @change="handleSearch"
        />
        
        <div class="search-actions">
          <el-button type="primary" @click="handleSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button @click="resetSearch">
            <el-icon><RefreshLeft /></el-icon>
            重置
          </el-button>
          <el-button 
            text 
            @click="toggleAdvancedSearch"
            :type="showAdvanced ? 'primary' : ''"
          >
            <el-icon><Setting /></el-icon>
            {{ showAdvanced ? '收起筛选' : '高级筛选' }}
          </el-button>
        </div>
      </div>
      
      <!-- 高级搜索面板 -->
      <div v-show="showAdvanced" class="advanced-search">
        <el-row :gutter="16">
          <el-col :span="6">
            <el-select
              v-model="searchForm.owner"
              placeholder="项目负责人"
              clearable
              filterable
              @change="handleSearch"
            >
              <el-option
                v-for="user in userOptions"
                :key="user.id"
                :label="user.name"
                :value="user.id"
              />
            </el-select>
          </el-col>
          
          <el-col :span="6">
            <el-select
              v-model="searchForm.priority"
              placeholder="优先级"
              clearable
              @change="handleSearch"
            >
              <el-option label="高优先级" value="high" />
              <el-option label="中优先级" value="medium" />
              <el-option label="低优先级" value="low" />
            </el-select>
          </el-col>
          
          <el-col :span="6">
            <el-input-number
              v-model="searchForm.minRequirements"
              placeholder="最小需求数"
              :min="0"
              controls-position="right"
              class="number-input"
              @change="handleSearch"
            />
          </el-col>
          
          <el-col :span="6">
            <el-checkbox-group 
              v-model="searchForm.tags"
              @change="handleSearch"
            >
              <el-checkbox value="analyzed">已分析</el-checkbox>
              <el-checkbox value="imported">已导入</el-checkbox>
              <el-checkbox value="archived">已归档</el-checkbox>
            </el-checkbox-group>
          </el-col>
        </el-row>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { 
  Plus, Upload, Delete, Download, Search, RefreshLeft, Setting
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  selectedCount: {
    type: Number,
    default: 0
  },
  userOptions: {
    type: Array,
    default: () => []
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
  'export',
  'search',
  'reset'
])

// 响应式数据
const showAdvanced = ref(false)

const searchForm = reactive({
  name: '',
  status: '',
  domain: '',
  dateRange: null,
  owner: '',
  priority: '',
  minRequirements: null,
  tags: []
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
  
  return params
})

// 方法
const handleSearch = () => {
  emit('search', searchParams.value)
}

const resetSearch = () => {
  Object.assign(searchForm, {
    name: '',
    status: '',
    domain: '',
    dateRange: null,
    owner: '',
    priority: '',
    minRequirements: null,
    tags: []
  })
  
  showAdvanced.value = false
  emit('reset')
}

const toggleAdvancedSearch = () => {
  showAdvanced.value = !showAdvanced.value
}

// 监听搜索参数变化，自动搜索（防抖）
let searchTimer = null
watch(() => searchForm.name, () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    if (searchForm.name !== '') {
      handleSearch()
    }
  }, 500)
})
</script>

<style lang="scss" scoped>
.project-toolbar {
  background: white;
  border-radius: 8px;
  padding: 16px 20px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);

  .toolbar-left {
    display: flex;
    gap: 12px;
    margin-bottom: 16px;

    .el-button {
      height: 36px;
    }
  }

  .toolbar-right {
    .search-controls {
      display: flex;
      align-items: center;
      gap: 12px;
      flex-wrap: wrap;

      .search-input {
        width: 240px;
      }

      .filter-select {
        width: 140px;
      }

      .date-picker {
        width: 240px;
      }

      .search-actions {
        display: flex;
        gap: 8px;
        margin-left: auto;
      }
    }

    .advanced-search {
      margin-top: 16px;
      padding: 16px;
      background: #f8fafc;
      border-radius: 6px;
      border: 1px solid #e5e7eb;

      .number-input {
        width: 100%;
      }

      .el-checkbox-group {
        display: flex;
        flex-wrap: wrap;
        gap: 8px;

        .el-checkbox {
          margin-right: 0;
        }
      }
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .project-toolbar {
    .toolbar-right {
      .search-controls {
        .search-input {
          width: 180px;
        }
        
        .filter-select {
          width: 120px;
        }
        
        .date-picker {
          width: 200px;
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .project-toolbar {
    .toolbar-left {
      flex-wrap: wrap;
    }

    .toolbar-right {
      .search-controls {
        flex-direction: column;
        align-items: stretch;
        gap: 8px;

        .search-input,
        .filter-select,
        .date-picker {
          width: 100%;
        }

        .search-actions {
          margin-left: 0;
          justify-content: flex-end;
        }
      }
    }
  }
}
</style>