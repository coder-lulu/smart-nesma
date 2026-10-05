<template>
  <div class="knowledge-rules">
    <el-table
      v-loading="loading"
      :data="tableData"
      style="width: 100%"
      @selection-change="handleSelectionChange"
      @sort-change="handleSortChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="id" label="ID" width="80" sortable="custom" />
      <el-table-column prop="ruleName" label="规则名称" min-width="200" show-overflow-tooltip>
        <template #default="scope">
          <el-link @click="handleView(scope.row)" type="primary">
            {{ scope.row.ruleName }}
          </el-link>
        </template>
      </el-table-column>
      <el-table-column prop="category" label="规则类别" width="120">
        <template #default="scope">
          <el-tag :type="getRuleCategoryType(scope.row.category)" size="small">
            {{ getRuleCategoryLabel(scope.row.category) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="priority" label="优先级" width="100">
        <template #default="scope">
          <el-tag 
            :type="getPriorityType(scope.row.priority)" 
            size="small"
          >
            {{ getPriorityLabel(scope.row.priority) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="status" label="状态" width="100">
        <template #default="scope">
          <el-tag 
            :type="scope.row.status === 'ACTIVE' ? 'success' : 'info'" 
            size="small"
          >
            {{ scope.row.status === 'ACTIVE' ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="conditionExpression" label="条件表达式" min-width="200" show-overflow-tooltip>
        <template #default="scope">
          <code class="condition-code">{{ scope.row.conditionExpression }}</code>
        </template>
      </el-table-column>
      <el-table-column prop="actionExpression" label="动作表达式" min-width="200" show-overflow-tooltip>
        <template #default="scope">
          <code class="action-code">{{ scope.row.actionExpression }}</code>
        </template>
      </el-table-column>
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
import { ElMessage, ElMessageBox } from 'element-plus'
import { getKnowledgeRuleList, deleteKnowledgeRule } from '@/api/nesma'

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
    
    const res = await getKnowledgeRuleList(params)
    tableData.value = res.data.rules || []
    pagination.total = res.data.total || 0
  } catch (error) {
    console.error('获取知识规则列表失败:', error)
    ElMessage.error('获取数据失败')
  } finally {
    loading.value = false
  }
}

// 获取规则类别标签类型
const getRuleCategoryType = (category) => {
  const typeMap = {
    'VALIDATION': 'success',
    'CALCULATION': 'primary',
    'BUSINESS': 'warning',
    'SYSTEM': 'info'
  }
  return typeMap[category] || ''
}

// 获取规则类别标签文本
const getRuleCategoryLabel = (category) => {
  const labelMap = {
    'VALIDATION': '验证规则',
    'CALCULATION': '计算规则',
    'BUSINESS': '业务规则',
    'SYSTEM': '系统规则'
  }
  return labelMap[category] || category
}

// 获取优先级标签类型
const getPriorityType = (priority) => {
  const typeMap = {
    'HIGH': 'danger',
    'MEDIUM': 'warning',
    'LOW': 'info'
  }
  return typeMap[priority] || ''
}

// 获取优先级标签文本
const getPriorityLabel = (priority) => {
  const labelMap = {
    'HIGH': '高',
    'MEDIUM': '中',
    'LOW': '低'
  }
  return labelMap[priority] || priority
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

// 查看规则
const handleView = (row) => {
  emit('view', row)
}

// 编辑规则
const handleEdit = (row) => {
  emit('edit', row)
}

// 删除规则
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除规则"${row.ruleName}"吗？`,
      '确认删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await deleteKnowledgeRule(row.id)
    ElMessage.success('删除成功')
    getTableData()
    emit('delete', row)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除规则失败:', error)
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
.knowledge-rules {
  padding: 0;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.condition-code,
.action-code {
  background-color: #f5f7fa;
  padding: 2px 6px;
  border-radius: 3px;
  font-size: 12px;
  color: #606266;
}

.condition-code {
  color: #67c23a;
}

.action-code {
  color: #409eff;
}
</style> 