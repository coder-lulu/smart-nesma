<template>
  <div>
    <warning-bar title="管理文档生成任务，支持Word/Excel/PDF多格式导出" />
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="document-add" @click="handleCreateDocument">
          新建文档
        </el-button>
        <el-button type="success" icon="tools" @click="handleBatchGenerate">
          批量生成
        </el-button>
        <el-button icon="setting" @click="handleTemplateManage">
          模板管理
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
      
      <!-- 搜索筛选 -->
      <div class="gva-search-box">
        <el-form ref="searchFormRef" :inline="true" :model="searchForm">
          <el-form-item label="文档名称">
            <el-input v-model="searchForm.keyword" placeholder="搜索文档名称" clearable @keyup.enter="getTableData" />
          </el-form-item>
          <el-form-item label="文档类型">
            <el-select v-model="searchForm.type" placeholder="文档类型" clearable @change="getTableData">
              <el-option label="Word" value="word" />
              <el-option label="Excel" value="excel" />
              <el-option label="PDF" value="pdf" />
            </el-select>
          </el-form-item>
          <el-form-item label="文档格式">
            <el-select v-model="searchForm.format" placeholder="文档格式" clearable @change="getTableData">
              <el-option label="需求规格说明书" value="requirement_spec" />
              <el-option label="NESMA评估报告" value="nesma_report" />
              <el-option label="业务需求汇总表" value="business_summary" />
            </el-select>
          </el-form-item>
          <el-form-item label="生成状态">
            <el-select v-model="searchForm.status" placeholder="生成状态" clearable @change="getTableData">
              <el-option label="待生成" value="pending" />
              <el-option label="生成中" value="generating" />
              <el-option label="已完成" value="completed" />
              <el-option label="生成失败" value="failed" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="search" @click="getTableData">查询</el-button>
            <el-button icon="refresh" @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 文档表格 -->
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
        <el-table-column align="left" label="文档名称" prop="name" min-width="200" />
        <el-table-column align="left" label="文档类型" width="100">
          <template #default="scope">
            <el-tag :type="getTypeColor(scope.row.type)" size="small">
              {{ scope.row.type?.toUpperCase() }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="文档格式" width="150">
          <template #default="scope">
            <span>{{ getFormatText(scope.row.format) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="生成状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="文件大小" width="100">
          <template #default="scope">
            <span v-if="scope.row.file_size">{{ formatFileSize(scope.row.file_size) }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="创建人" prop="creator_name" width="120" />
        <el-table-column align="left" label="操作" min-width="200">
          <template #default="scope">
            <el-button type="primary" link icon="view" @click="handleView(scope.row)">
              查看
            </el-button>
            <el-button type="success" link icon="download" @click="handleDownload(scope.row)" :disabled="scope.row.status !== 'completed'">
              下载
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

    <!-- 文档表单对话框 -->
    <el-dialog
      v-model="formDialogVisible"
      :title="currentDocument.ID ? '查看文档详情' : '文档生成'"
      width="80%"
      :before-close="handleDialogClose"
      destroy-on-close
    >
      <DocumentForm
        :form-data="currentDocument"
        :dialog-type="currentDocument.ID ? 'view' : 'create'"
        @submit="handleFormSuccess"
        @cancel="handleFormCancel"
      />
    </el-dialog>

    <!-- 批量生成对话框 -->
    <el-dialog
      v-model="batchDialogVisible"
      title="批量生成文档"
      width="60%"
      :before-close="handleDialogClose"
      destroy-on-close
    >
      <BatchGenerateForm
        @success="handleBatchSuccess"
        @cancel="handleBatchCancel"
      />
    </el-dialog>

    <!-- 模板管理对话框 -->
    <el-dialog
      v-model="templateDialogVisible"
      title="模板管理"
      width="70%"
      :before-close="handleDialogClose"
      destroy-on-close
    >
      <TemplateManagement
        @success="handleTemplateSuccess"
        @cancel="handleTemplateCancel"
      />
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Refresh, Plus, DocumentAdd, Tools, Setting, Delete, Search, 
  View, Edit, Download 
} from '@element-plus/icons-vue'

// 组件导入
import WarningBar from '@/components/warningBar/warningBar.vue'
import DocumentForm from './components/DocumentForm.vue'
import BatchGenerateForm from './components/BatchGenerateForm.vue'
import TemplateManagement from './components/TemplateManagement.vue'

// API导入
import {
  getDocumentList,
  deleteDocument,
  downloadDocument,
  getDocument,
  batchDeleteDocuments
} from '@/api/nesma'

// 响应式数据
const loading = ref(false)
const searchFormRef = ref(null)
const searchForm = reactive({
  keyword: '',
  type: '',
  format: '',
  status: ''
})
const selectedIds = ref([])

// 表格数据
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// 对话框状态
const formDialogVisible = ref(false)
const batchDialogVisible = ref(false)
const templateDialogVisible = ref(false)
const currentDocument = ref({})

// 方法实现
const handleCreateDocument = () => {
  currentDocument.value = {}
  formDialogVisible.value = true
}

const handleBatchGenerate = () => {
  batchDialogVisible.value = true
}

const handleTemplateManage = () => {
  templateDialogVisible.value = true
}

const handleEdit = (document) => {
  currentDocument.value = { ...document }
  formDialogVisible.value = true
}

const handleView = async (document) => {
  try {
    const res = await getDocument(document.ID)
    if (res.code === 0) {
      // 显示文档详情对话框
      currentDocument.value = res.data
      formDialogVisible.value = true
    } else {
      ElMessage.error('获取文档详情失败')
    }
  } catch (error) {
    console.error('获取文档详情失败:', error)
    ElMessage.error('获取文档详情失败')
  }
}

import { downloadFile } from '@/utils/downloadFile'

const path = ref(import.meta.env.VITE_BASE_API)
const handleDownload = async (document) => {
  try {
    // 使用url下载文件
    var fileType = document.type === 'word' ? 'docx' : 
                   document.type === 'excel' ? 'xlsx' : 
                   document.type === 'pdf' ? 'pdf' : 'docx'
    if (document.filePath.indexOf('http://') > -1 || document.filePath.indexOf('https://') > -1) {
      downloadFile(document.filePath, document.name, fileType)
    } else {
      downloadFile(path.value + '/' + document.filePath, document.name, fileType)
    }
    ElMessage.success('下载成功')
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败')
  }
}

const handleDelete = async (document) => {
  try {
    await ElMessageBox.confirm('确定要删除此文档吗？此操作不可恢复！', '确认删除', {
      type: 'error'
    })
    
    const res = await deleteDocument(document.ID)
    if (res.code === 0) {
      ElMessage.success('文档已删除')
      getTableData()
    } else {
      ElMessage.error('删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

const handleBatchDelete = async () => {
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${selectedIds.value.length} 个文档吗？`, '批量删除', {
      type: 'error'
    })
    
    const res = await batchDeleteDocuments({ ids: selectedIds.value })
    if (res.code === 0) {
      ElMessage.success('文档已批量删除')
      getTableData()
    } else {
      ElMessage.error('批量删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('批量删除失败')
    }
  }
}

const handleSelectionChange = (selection) => {
  selectedIds.value = selection.map(item => item.ID)
}

// 搜索重置
const resetSearch = () => {
  searchForm.keyword = ''
  searchForm.type = ''
  searchForm.format = ''
  searchForm.status = ''
  page.value = 1
  getTableData()
}

// 辅助方法
const getTypeColor = (type) => {
  const typeMap = {
    'word': 'primary',
    'excel': 'success', 
    'pdf': 'danger'
  }
  return typeMap[type] || 'info'
}

const getFormatText = (format) => {
  const formatMap = {
    'requirement_spec': '需求规格说明书',
    'nesma_report': 'NESMA评估报告',
    'business_summary': '业务需求汇总表'
  }
  return formatMap[format] || '未知格式'
}

const getStatusType = (status) => {
  const statusMap = {
    'pending': 'warning',
    'generating': 'primary',
    'completed': 'success',
    'failed': 'danger'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    'pending': '待生成',
    'generating': '生成中',
    'completed': '已完成',
    'failed': '生成失败'
  }
  return statusMap[status] || '未知'
}

const formatFileSize = (size) => {
  if (!size) return '-'
  const kb = size / 1024
  if (kb < 1024) return `${kb.toFixed(1)}KB`
  const mb = kb / 1024
  return `${mb.toFixed(1)}MB`
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
    
    const response = await getDocumentList(params)
    tableData.value = response.data.list || []
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载文档列表失败：' + error.message)
  } finally {
    loading.value = false
  }
}

// 对话框事件
const handleFormSuccess = () => {
  formDialogVisible.value = false
  getTableData()
}

const handleFormCancel = () => {
  formDialogVisible.value = false
}

const handleBatchSuccess = () => {
  batchDialogVisible.value = false
  getTableData()
}

const handleBatchCancel = () => {
  batchDialogVisible.value = false
}

const handleTemplateSuccess = () => {
  templateDialogVisible.value = false
}

const handleTemplateCancel = () => {
  templateDialogVisible.value = false
}

const handleDialogClose = () => {
  formDialogVisible.value = false
  batchDialogVisible.value = false
  templateDialogVisible.value = false
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