<template>
  <div>
    <warning-bar title="管理NESMA知识库条目、规则和案例研究" />
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="plus" @click="handleCreate">
          新建知识
        </el-button>
        <el-button icon="upload" @click="handleImport">
          导入知识
        </el-button>
        <el-button 
          icon="delete" 
          type="danger" 
          :disabled="!selectedIds.length"
          @click="handleBatchDelete"
        >
          批量删除
        </el-button>
        <el-button @click="getTableData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
      
      <!-- 搜索和筛选 -->
      <div class="gva-search-box">
        <el-form ref="searchForm" :inline="true" :model="searchForm">
          <el-form-item label="知识类别">
            <el-select v-model="searchForm.category" placeholder="知识类别" clearable @change="getTableData">
              <el-option label="NESMA标准" value="NESMA_STANDARD" />
              <el-option label="最佳实践" value="BEST_PRACTICE" />
              <el-option label="案例研究" value="CASE_STUDY" />
              <el-option label="业务规则" value="RULE" />
            </el-select>
          </el-form-item>
          <el-form-item label="知识领域">
            <el-select v-model="searchForm.domain" placeholder="知识领域" clearable @change="getTableData">
              <el-option label="软件度量" value="软件度量" />
              <el-option label="数据功能" value="数据功能" />
              <el-option label="事务功能" value="事务功能" />
              <el-option label="质量管理" value="质量管理" />
            </el-select>
          </el-form-item>
          <el-form-item label="关键词">
            <el-input v-model="searchForm.keyword" placeholder="搜索知识标题" clearable @keyup.enter="getTableData" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="search" @click="getTableData">查询</el-button>
            <el-button icon="refresh" @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 知识表格 -->
      <el-table
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column align="left" label="创建时间" width="180">
          <template #default="scope">
            <span>{{ formatDate(scope.row.CreatedAt) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="知识标题" prop="title" min-width="200" />
        <el-table-column align="left" label="知识类别" width="120">
          <template #default="scope">
            <el-tag :type="getCategoryType(scope.row.category)" size="small">
              {{ getCategoryText(scope.row.category) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="知识领域" prop="domain" width="120" />
        <!-- <el-table-column align="left" label="重要性" width="100">
          <template #default="scope">
            <el-tag :type="getImportanceType(scope.row.importance)" size="small">
              {{ scope.row.importance }}
            </el-tag>
          </template>
        </el-table-column> -->
        <el-table-column align="left" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" min-width="200">
          <template #default="scope">
            <el-button type="primary" link icon="view" @click="handleView(scope.row)">
              查看
            </el-button>
            <el-button type="primary" link icon="edit" @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button type="danger" link icon="delete" @click="handleDelete(scope.row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- 对话框组件 -->
    <KnowledgeEntryDialog
      v-model="entryDialogVisible"
      :entry="currentEntry"
      @success="handleDialogSuccess"
    />

    <KnowledgeImportDialog
      v-model="importDialogVisible"
      @success="handleImportSuccess"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Refresh, Plus, Upload, Delete, Search, View, Edit, ArrowDown } from '@element-plus/icons-vue'

// 组件导入
import KnowledgeEntryDialog from './components/KnowledgeEntryDialog.vue'
import KnowledgeImportDialog from './components/KnowledgeImportDialog.vue'

// API导入
import {
  getKnowledgeList,
  createKnowledge,
  updateKnowledge,
  deleteKnowledgeEntry
} from '@/api/nesma'

// 响应式数据
const loading = ref(false)
const searchForm = reactive({
  category: '',
  domain: '',
  keyword: ''
})
const selectedIds = ref([])

// 表格数据
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// 对话框状态
const entryDialogVisible = ref(false)
const importDialogVisible = ref(false)
const currentEntry = ref({})

// 方法实现
const handleCreate = () => {
  currentEntry.value = {}
  entryDialogVisible.value = true
}

const handleEdit = (entry) => {
  currentEntry.value = { ...entry }
  entryDialogVisible.value = true
}

const handleView = (entry) => {
  currentEntry.value = entry
  entryDialogVisible.value = true
}

const handleDelete = async (entry) => {
  try {
    await ElMessageBox.confirm('确定要删除此知识条目吗？此操作不可恢复！', '确认删除', {
      type: 'error'
    })
    await deleteKnowledgeEntry(entry.ID)
    ElMessage.success('知识条目已删除')
    getTableData()
  } catch (e) {}
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 个知识条目吗？`, '批量删除', {
      type: 'error'
    })
    // 没有批量接口时循环单删
    await Promise.all(selectedIds.value.map(id => deleteKnowledgeEntry(id)))
    ElMessage.success('知识条目已批量删除')
    getTableData()
  } catch (e) {}
}

const handleImport = () => {
  importDialogVisible.value = true
}

const handleMoreActions = (command) => {
  switch (command) {
    case 'importStandard':
      ElMessage.success('正在导入NESMA标准知识...')
      break
    case 'export':
      ElMessage.success('正在导出Excel...')
      break
    case 'backup':
      ElMessage.success('正在备份知识库...')
      break
  }
}

const handleSelectionChange = (selection) => {
  selectedIds.value = selection.map(item => item.ID)
}

// 搜索相关
const resetSearch = () => {
  searchForm.category = ''
  searchForm.domain = ''
  searchForm.keyword = ''
  page.value = 1
  getTableData()
}

// 辅助方法
const getCategoryType = (category) => {
  const categoryMap = {
    'NESMA_STANDARD': 'danger',
    'BEST_PRACTICE': 'primary',
    'CASE_STUDY': 'success',
    'RULE': 'warning',
    'user_adopted': 'success'
  }
  return categoryMap[category] || 'info'
}

const getCategoryText = (category) => {
  const categoryMap = {
    'NESMA_STANDARD': 'NESMA标准',
    'BEST_PRACTICE': '最佳实践',
    'CASE_STUDY': '案例研究',
    'RULE': '业务规则',
    'user_adopted': '用户采用'
  }
  return categoryMap[category] || '未知'
}

const getImportanceType = (importance) => {
  const importanceMap = {
    'high': 'danger',
    'medium': 'warning',
    'low': 'info'
  }
  return importanceMap[importance] || 'info'
}

const getStatusType = (status) => {
  const statusMap = {
    'active': 'success',
    'draft': 'warning',
    'archived': 'info'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    'active': '已发布',
    'draft': '草稿',
    'archived': '已归档'
  }
  return statusMap[status] || '未知'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN')
}

// 数据加载
const getTableData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      ...searchForm
    }
    
    const response = await getKnowledgeList(params)
    tableData.value = response.data.entries || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载知识列表失败：' + error.message)
  } finally {
    loading.value = false
  }
}

// 对话框事件
const handleDialogSuccess = () => {
  entryDialogVisible.value = false
  getTableData()
}

const handleImportSuccess = () => {
  importDialogVisible.value = false
  getTableData()
}

// 分页事件
const handleSizeChange = (size) => {
  pageSize.value = size
  getTableData()
}

const handleCurrentChange = (currentPage) => {
  page.value = currentPage
  getTableData()
}

// 生命周期
onMounted(() => {
  getTableData()
})
</script>

<style lang="scss" scoped>
// 使用gin-vue-admin的默认样式
</style>