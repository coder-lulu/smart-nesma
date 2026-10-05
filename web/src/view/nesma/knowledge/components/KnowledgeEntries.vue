<template>
  <div class="knowledge-entries">
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="ID" width="80" sortable="custom" />
      <el-table-column prop="title" label="知识标题" min-width="200" show-overflow-tooltip>
        <template #default="scope">
          <el-link @click="handleView(scope.row)" type="primary">
            {{ scope.row.title }}
          </el-link>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="知识类别" width="120">
        <template #default="scope">
          <el-tag :type="getCategoryType(scope.row.category)" size="small">
            {{ getCategoryLabel(scope.row.category) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="domain" label="知识领域" width="120" show-overflow-tooltip />
      <el-table-column prop="tags" label="标签" width="150">
        <template #default="scope">
          <div class="tags-container">
            <el-tag
              v-for="tag in getTagsArray(scope.row.tags)"
              :key="tag"
              size="small"
              style="margin-right: 4px; margin-bottom: 2px;"
            >
              {{ tag }}
            </el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="confidenceScore" label="置信度" width="100">
        <template #default="scope">
          <el-progress
            :percentage="Math.round(scope.row.confidenceScore * 100)"
            :stroke-width="6"
            :show-text="false"
            :color="getConfidenceColor(scope.row.confidenceScore)"
          />
          <span style="margin-left: 8px; font-size: 12px;">
            {{ Math.round(scope.row.confidenceScore * 100) }}%
          </span>
        </template>
      </el-table-column>
      <el-table-column prop="usageCount" label="使用次数" width="100" sortable="custom" />
      <el-table-column prop="author" label="作者" width="100" show-overflow-tooltip />
      <el-table-column prop="createdAt" label="创建时间" width="160">
        <template #default="scope">
          {{ formatDate(scope.row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="scope">
          <el-button size="small" @click="handleView(scope.row)">查看</el-button>
          <el-button size="small" type="primary" @click="handleEdit(scope.row)">编辑</el-button>
          <el-button 
            size="small" 
            type="danger" 
            @click="handleDelete(scope.row)"
          >
            删除
          </el-button>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="pagination.page"
        v-model:page-size="pagination.pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="pagination.total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getKnowledgeEntryList } from '@/api/nesma'

const props = defineProps({
  searchForm: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['selection-change', 'edit', 'delete', 'view'])

// 响应式数据
const loading = ref(false)
const tableData = ref([])
const pagination = reactive({
  page: 1,
  pageSize: 20,
  total: 0
})

// 监听搜索条件变化
watch(() => props.searchForm, () => {
  pagination.page = 1
  getTableData()
}, { deep: true })

// 初始化
onMounted(() => {
  getTableData()
})

// 获取表格数据
const getTableData = async () => {
  try {
    loading.value = true
    const params = {
      ...props.searchForm,
      page: pagination.page,
      pageSize: pagination.pageSize
    }
    
    const res = await getKnowledgeEntryList(params)
    tableData.value = res.data.entries || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('获取知识条目列表失败:', error)
    ElMessage.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

// 获取知识类别标签类型
const getCategoryType = (category) => {
  const typeMap = {
    'NESMA_STANDARD': 'danger',
    'BEST_PRACTICE': 'success',
    'CASE_STUDY': 'warning',
    'RULE': 'info',
    '用户采纳': 'primary'
  }
  return typeMap[category] || ''
}

// 获取知识类别标签文本
const getCategoryLabel = (category) => {
  const labelMap = {
    'NESMA_STANDARD': 'NESMA标准',
    'BEST_PRACTICE': '最佳实践',
    'CASE_STUDY': '案例研究',
    'RULE': '业务规则',
    '用户采纳': '用户采纳'
  }
  return labelMap[category] || category
}

// 解析标签数组
const getTagsArray = (tags) => {
  if (!tags) return []
  try {
    return Array.isArray(tags) ? tags : JSON.parse(tags)
  } catch {
    return []
  }
}

// 获取置信度颜色
const getConfidenceColor = (score) => {
  if (score >= 0.8) return '#67c23a'
  if (score >= 0.6) return '#e6a23c'
  return '#f56c6c'
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleString('zh-CN')
}

// 选择项变化
const handleSelectionChange = (selection) => {
  emit('selection-change', selection)
}

// 排序变化
const handleSortChange = ({ column, prop, order }) => {
  // 实现排序逻辑
  console.log('排序变化:', { column, prop, order })
}

// 页面大小变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  pagination.page = 1
  getTableData()
}

// 当前页变化
const handleCurrentChange = (page) => {
  pagination.page = page
  getTableData()
}

// 查看
const handleView = (row) => {
  emit('view', row)
}

// 编辑
const handleEdit = (row) => {
  emit('edit', row)
}

// 删除
const handleDelete = (row) => {
  emit('delete', row)
}

// 暴露方法给父组件
defineExpose({
  getTableData
})
</script>

<style scoped>
.knowledge-entries {
  width: 100%;
}

.tags-container {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

:deep(.el-table__cell) {
  padding: 8px 0;
}

:deep(.el-progress-bar__outer) {
  height: 6px !important;
}

:deep(.el-link) {
  font-weight: 500;
}
</style>