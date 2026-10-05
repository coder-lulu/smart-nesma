<template>
  <div class="case-studies">
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="ID" width="80" sortable="custom" />
      <el-table-column prop="projectName" label="项目名称" min-width="200" show-overflow-tooltip>
        <template #default="scope">
          <el-link @click="handleView(scope.row)" type="primary">
            {{ scope.row.projectName }}
          </el-link>
        </template>
      </el-table-column>
      <el-table-column prop="domain" label="应用领域" width="120" show-overflow-tooltip />
      <el-table-column prop="organizationType" label="组织类型" width="120">
        <template #default="scope">
          <el-tag size="small" type="info">
            {{ scope.row.organizationType }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="projectScale" label="项目规模" width="120">
        <template #default="scope">
          <el-tag :type="getScaleType(scope.row.projectScale)" size="small">
            {{ getScaleLabel(scope.row.projectScale) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="nesmaResult" label="NESMA结果" width="120">
        <template #default="scope">
          <div class="nesma-result">
            <span class="result-value">{{ scope.row.nesmaResult }}</span>
            <span class="result-unit">FP</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="actualEffort" label="实际工作量" width="120">
        <template #default="scope">
          <div class="effort-result">
            <span class="result-value">{{ scope.row.actualEffort }}</span>
            <span class="result-unit">人天</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="productivityRatio" label="生产率" width="100">
        <template #default="scope">
          <div class="productivity-ratio">
            <span class="ratio-value">{{ scope.row.productivityRatio }}</span>
            <span class="ratio-unit">FP/人天</span>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag 
            :type="getStatusType(scope.row.status)" 
            size="small"
          >
            {{ getStatusLabel(scope.row.status) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="completionDate" label="完成时间" width="160">
        <template #default="scope">
          {{ formatDate(scope.row.completionDate) }}
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { getCaseStudyList, deleteCaseStudy } from '@/api/nesma'

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
    
    const res = await getCaseStudyList(params)
    tableData.value = res.data.cases || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('获取案例研究列表失败:', error)
    ElMessage.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

// 获取项目规模标签类型
const getScaleType = (scale) => {
  const typeMap = {
    'SMALL': 'success',
    'MEDIUM': 'warning',
    'LARGE': 'danger'
  }
  return typeMap[scale] || ''
}

// 获取项目规模标签文本
const getScaleLabel = (scale) => {
  const labelMap = {
    'SMALL': '小型',
    'MEDIUM': '中型',
    'LARGE': '大型'
  }
  return labelMap[scale] || scale
}

// 获取状态标签类型
const getStatusType = (status) => {
  const typeMap = {
    'COMPLETED': 'success',
    'IN_PROGRESS': 'warning',
    'CANCELLED': 'danger',
    'DRAFT': 'info'
  }
  return typeMap[status] || ''
}

// 获取状态标签文本
const getStatusLabel = (status) => {
  const labelMap = {
    'COMPLETED': '已完成',
    'IN_PROGRESS': '进行中',
    'CANCELLED': '已取消',
    'DRAFT': '草稿'
  }
  return labelMap[status] || status
}

// 格式化日期
const formatDate = (date) => {
  if (!date) return ''
  return new Date(date).toLocaleDateString('zh-CN')
}

// 选择项变化
const handleSelectionChange = (selection) => {
  emit('selection-change', selection)
}

// 排序变化
const handleSortChange = ({ column, prop, order }) => {
  console.log('排序变化:', { column, prop, order })
}

// 页面大小变化
const handleSizeChange = (size) => {
  pagination.pageSize = size
  getTableData()
}

// 当前页变化
const handleCurrentChange = (page) => {
  pagination.page = page
  getTableData()
}

// 查看案例
const handleView = (row) => {
  emit('view', row)
}

// 编辑案例
const handleEdit = (row) => {
  emit('edit', row)
}

// 删除案例
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除案例"${row.projectName}"吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteCaseStudy(row.id)
    ElMessage.success('删除成功')
    getTableData()
    emit('delete', row)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除案例失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 暴露方法
defineExpose({
  getTableData
})
</script>

<style scoped>
.case-studies {
  padding: 0;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.nesma-result,
.effort-result,
.productivity-ratio {
  display: flex;
  align-items: center;
  gap: 4px;
}

.result-value,
.ratio-value {
  font-weight: bold;
  color: #409eff;
}

.result-unit,
.ratio-unit {
  font-size: 12px;
  color: #909399;
}
</style> 