<template>
  <div class="template-management">
    <div class="template-toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleCreateTemplate">
          新建模板
        </el-button>
        <el-button icon="Upload" @click="handleUploadTemplate">上传模板</el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchKeyword"
          placeholder="搜索模板名称"
          style="width: 200px; margin-right: 10px;"
          clearable
          @keyup.enter="getTemplateList"
        />
        <el-select
          v-model="filterType"
          placeholder="模板类型"
          style="width: 120px; margin-right: 10px;"
          clearable
        >
          <el-option label="Word" value="word" />
          <el-option label="Excel" value="excel" />
          <el-option label="PDF" value="pdf" />
        </el-select>
        <el-select
          v-model="filterFormat"
          placeholder="文档格式"
          style="width: 140px; margin-right: 10px;"
          clearable
        >
          <el-option label="需求规格说明书" value="requirement_spec" />
          <el-option label="NESMA评估报告" value="nesma_report" />
          <el-option label="业务需求汇总表" value="business_summary" />
        </el-select>
        <el-button icon="Search" @click="getTemplateList">搜索</el-button>
        <el-button icon="RefreshLeft" @click="resetSearch">重置</el-button>
      </div>
    </div>

    <el-table
      v-loading="loading"
      :data="templateList"
      style="width: 100%"
      @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />
      <el-table-column prop="ID" label="ID" width="80" />
      <el-table-column prop="name" label="模板名称" min-width="200" show-overflow-tooltip>
        <template #default="scope">
          <div class="template-name">
            <el-icon class="template-icon">
              <Document v-if="scope.row.type === 'word'" />
              <Grid v-else-if="scope.row.type === 'excel'" />
              <DocumentCopy v-else />
            </el-icon>
            <span>{{ scope.row.name }}</span>
            <el-tag v-if="scope.row.isDefault" size="small" type="success">默认</el-tag>
          </div>
        </template>
      </el-table-column>
      <el-table-column prop="type" label="模板类型" width="100">
        <template #default="scope">
          <el-tag :type="getTypeTagType(scope.row.type)" size="small">
            {{ getTypeLabel(scope.row.type) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="format" label="文档格式" width="150">
        <template #default="scope">
          <el-tag size="small" type="info">
            {{ getFormatLabel(scope.row.format) }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="isActive" label="状态" width="100">
        <template #default="scope">
          <el-tag :type="scope.row.isActive ? 'success' : 'info'" size="small">
            {{ scope.row.isActive ? '启用' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="usageCount" label="使用次数" width="100" />
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column prop="createdAt" label="创建时间" width="150">
        <template #default="scope">
          {{ formatDate(scope.row.createdAt) }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="scope">
          <div class="action-buttons">
            <el-button 
              size="small" 
              @click="handleView(scope.row)"
              icon="View"
            >
              查看
            </el-button>
            <el-button 
              size="small" 
              type="primary"
              @click="handleEdit(scope.row)"
              icon="Edit"
            >
              编辑
            </el-button>
            <el-dropdown @command="(command) => handleDropdownCommand(command, scope.row)">
              <el-button size="small" type="info" icon="More" />
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item 
                    command="setDefault" 
                    icon="Star"
                    :disabled="scope.row.isDefault"
                  >
                    设为默认
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="activate" 
                    icon="Check"
                    :disabled="scope.row.isActive"
                  >
                    启用
                  </el-dropdown-item>
                  <el-dropdown-item 
                    command="deactivate" 
                    icon="Close"
                    :disabled="!scope.row.isActive"
                  >
                    停用
                  </el-dropdown-item>
                  <el-dropdown-item command="download" icon="Download">下载</el-dropdown-item>
                  <el-dropdown-item command="copy" icon="CopyDocument">复制</el-dropdown-item>
                  <el-dropdown-item command="delete" icon="Delete" divided>删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <div class="pagination-container">
      <el-pagination
        v-model:current-page="page"
        v-model:page-size="pageSize"
        :page-sizes="[10, 20, 50, 100]"
        :total="total"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="handleSizeChange"
        @current-change="handleCurrentChange"
      />
    </div>

    <!-- 操作按钮 -->
    <div class="action-container">
      <el-button @click="handleCancel">关闭</el-button>
    </div>

    <!-- 模板编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogType === 'create' ? '新建模板' : '编辑模板'"
      width="800px"
      :close-on-click-modal="false"
    >
      <TemplateForm
        ref="templateFormRef"
        :form-data="currentTemplate"
        :dialog-type="dialogType"
        @submit="handleFormSubmit"
        @cancel="handleFormCancel"
      />
    </el-dialog>

    <!-- 模板预览对话框 -->
    <el-dialog
      v-model="previewDialogVisible"
      title="模板预览"
      width="90%"
      :close-on-click-modal="false"
    >
      <TemplatePreview
        ref="templatePreviewRef"
        :template-id="previewTemplateId"
      />
    </el-dialog>

    <!-- 上传表单对话框 -->
    <el-dialog
      v-model="uploadFormDialogVisible"
      title="上传模板"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form
        ref="uploadFormRef"
        :model="uploadFormData"
        :rules="uploadFormRules"
        label-width="100px"
      >
        <el-form-item label="选择文件">
          <div class="file-upload-container">
            <el-upload
              ref="uploadFileRef"
              action=""
              :on-change="handleFileChange"
              :before-upload="handleFileSelect"
              :show-file-list="false"
              accept=".docx,.xlsx,.pdf"
              :auto-upload="false"
            >
              <el-button 
                icon="Upload" 
                :type="uploadFormData.fileName ? 'success' : 'primary'"
                style="width: 100%"
              >
                <el-icon style="margin-right: 8px">
                  <Upload v-if="!uploadFormData.fileName" />
                  <SuccessFilled v-else />
                </el-icon>
                {{ uploadFormData.fileName ? '已选择文件' : '选择文件' }}
              </el-button>
            </el-upload>
            
            <!-- 文件信息显示区域 -->
            <div v-if="uploadFormData.fileName" class="selected-file-info">
              <div class="file-info-header">
                <el-icon class="file-icon">
                  <Document v-if="uploadFormData.type === 'word'" />
                  <Grid v-else-if="uploadFormData.type === 'excel'" />
                  <DocumentCopy v-else />
                </el-icon>
                <span class="file-name">{{ uploadFormData.fileName }}</span>
                <el-button size="small" type="danger" link @click="clearSelectedFile">
                  <el-icon><Close /></el-icon>
                </el-button>
              </div>
              <div class="file-info-details">
                <el-tag size="small" :type="getUploadFileTypeTag(uploadFormData.type)">
                  {{ getUploadFileTypeLabel(uploadFormData.type) }}
                </el-tag>
                <span class="file-size">{{ formatFileSize(uploadFormData.file?.size || 0) }}</span>
              </div>
            </div>
            
            <!-- 没有选择文件时的提示 -->
            <div v-else class="no-file-selected">
              <div class="form-tip">支持 Word (.docx)、Excel (.xlsx)、PDF (.pdf) 格式，文件大小不超过 10MB</div>
            </div>
          </div>
        </el-form-item>
        
        <el-form-item label="模板名称" prop="name">
          <el-input
            v-model="uploadFormData.name"
            placeholder="请输入模板名称"
            clearable
          />
        </el-form-item>
        
        <el-form-item label="模板类型" prop="type">
          <el-select
            v-model="uploadFormData.type"
            placeholder="请选择模板类型"
            style="width: 100%"
          >
            <el-option label="Word" value="word" />
            <el-option label="Excel" value="excel" />
            <el-option label="PDF" value="pdf" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="文档格式" prop="format">
          <el-select
            v-model="uploadFormData.format"
            placeholder="请选择文档格式"
            style="width: 100%"
          >
            <el-option label="需求规格说明书" value="requirement_spec" />
            <el-option label="NESMA评估报告" value="nesma_report" />
            <el-option label="业务需求汇总表" value="business_summary" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="uploadFormData.description"
            type="textarea"
            placeholder="请输入模板描述（可选）"
            :rows="3"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="cancelUploadForm">取消</el-button>
        <el-button type="primary" @click="confirmUpload" :loading="uploading">
          {{ uploading ? '上传中...' : '确认上传' }}
        </el-button>
      </template>
    </el-dialog>

    <!-- 上传进度对话框 -->
    <el-dialog
      v-model="uploadDialogVisible"
      title="上传模板"
      width="500px"
      :close-on-click-modal="false"
      :show-close="false"
    >
      <div class="upload-progress">
        <div class="upload-info">
          <el-icon class="upload-icon"><Upload /></el-icon>
          <div class="upload-text">
            <div class="upload-filename">{{ uploadInfo.filename }}</div>
            <div class="upload-status">{{ uploadInfo.status }}</div>
          </div>
        </div>
        <el-progress
          :percentage="uploadInfo.progress"
          :stroke-width="8"
          :text-inside="true"
        />
      </div>
      <template #footer>
        <el-button @click="cancelUpload" :disabled="uploading">
          {{ uploading ? '上传中...' : '取消' }}
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Document,
  Grid,
  DocumentCopy,
  Plus,
  Upload,
  Delete,
  Search,
  RefreshLeft,
  View,
  Edit,
  More,
  Star,
  Check,
  Close,
  Download,
  CopyDocument,
  SuccessFilled
} from '@element-plus/icons-vue'
import {
  getTemplateList as getTemplateListApi,
  createTemplate,
  updateTemplate,
  deleteTemplate,
  batchDeleteTemplates,
  setDefaultTemplate,
  activateTemplate,
  uploadTemplate
} from '@/api/nesma'
import TemplateForm from './TemplateForm.vue'
import TemplatePreview from './TemplatePreview.vue'

// Emits
const emit = defineEmits(['close', 'cancel'])

// 响应式数据
const loading = ref(false)
const templateList = ref([])
const selectedTemplates = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)

// 搜索过滤
const searchKeyword = ref('')
const filterType = ref('')
const filterFormat = ref('')

// 对话框状态
const dialogVisible = ref(false)
const dialogType = ref('create')
const currentTemplate = ref({})
const previewDialogVisible = ref(false)
const previewTemplateId = ref(null)

// 上传相关
const uploadFormDialogVisible = ref(false)
const uploadDialogVisible = ref(false)
const uploading = ref(false)
const uploadInfo = reactive({
  filename: '',
  status: '',
  progress: 0
})

// 上传表单相关
const uploadFormRef = ref(null)
const uploadFileRef = ref(null)
const uploadFormData = reactive({
  name: '',
  type: '',
  format: '',
  description: '',
  fileName: '',
  file: null
})

const uploadFormRules = {
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' },
    { min: 1, max: 100, message: '模板名称长度为 1-100 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择模板类型', trigger: 'change' }
  ],
  format: [
    { required: true, message: '请选择文档格式', trigger: 'change' }
  ]
}

// 表单引用
const templateFormRef = ref(null)
const templatePreviewRef = ref(null)

// 生命周期
onMounted(() => {
  getTemplateList()
})

// 获取模板列表
const getTemplateList = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      keyword: searchKeyword.value,
      type: filterType.value,
      format: filterFormat.value
    }
    const res = await getTemplateListApi(params)
    if (res.code === 0) {
      templateList.value = res.data.list || []
      total.value = res.data.total || 0
    }
  } catch (error) {
    console.error('获取模板列表失败:', error)
    ElMessage.error('获取模板列表失败')
  } finally {
    loading.value = false
  }
}

// 重置搜索
const resetSearch = () => {
  searchKeyword.value = ''
  filterType.value = ''
  filterFormat.value = ''
  page.value = 1
  getTemplateList()
}

// 选择变化
const handleSelectionChange = (selection) => {
  selectedTemplates.value = selection
}

// 分页处理
const handleSizeChange = (val) => {
  pageSize.value = val
  page.value = 1
  getTemplateList()
}

const handleCurrentChange = (val) => {
  page.value = val
  getTemplateList()
}

// 新建模板
const handleCreateTemplate = () => {
  dialogType.value = 'create'
  currentTemplate.value = {}
  dialogVisible.value = true
}

// 编辑模板
const handleEdit = (row) => {
  dialogType.value = 'edit'
  currentTemplate.value = { ...row }
  dialogVisible.value = true
}

// 查看模板
const handleView = (row) => {
  previewTemplateId.value = row.ID
  previewDialogVisible.value = true
}

// 批量删除
const handleBatchDelete = async () => {
  if (!selectedTemplates.value.length) {
    ElMessage.warning('请选择要删除的模板')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedTemplates.value.length} 个模板吗？`,
      '批量删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await batchDeleteTemplates({
      ids: selectedTemplates.value.map(item => item.ID)
    })
    if (res.code === 0) {
      ElMessage.success('删除成功')
      selectedTemplates.value = []
      getTemplateList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('批量删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 下拉菜单命令处理
const handleDropdownCommand = async (command, row) => {
  switch (command) {
    case 'setDefault':
      await handleSetDefault(row)
      break
    case 'activate':
      await handleActivate(row, true)
      break
    case 'deactivate':
      await handleActivate(row, false)
      break
    case 'download':
      await handleDownload(row)
      break
    case 'copy':
      await handleCopy(row)
      break
    case 'delete':
      await handleDelete(row)
      break
  }
}

// 设为默认
const handleSetDefault = async (row) => {
  try {
    const res = await setDefaultTemplate(row.ID)
    if (res.code === 0) {
      ElMessage.success('设置成功')
      getTemplateList()
    }
  } catch (error) {
    console.error('设置默认模板失败:', error)
    ElMessage.error('设置失败')
  }
}

// 启用/停用
const handleActivate = async (row, active) => {
  try {
    const res = await activateTemplate(row.ID, { active })
    if (res.code === 0) {
      ElMessage.success(`${active ? '启用' : '停用'}成功`)
      getTemplateList()
    }
  } catch (error) {
    console.error('操作失败:', error)
    ElMessage.error('操作失败')
  }
}

// 下载模板
const handleDownload = async (row) => {
  try {
    // 实现下载逻辑
    ElMessage.info('下载功能开发中...')
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败')
  }
}

// 复制模板
const handleCopy = async (row) => {
  try {
    const newTemplate = {
      ...row,
      name: `${row.name}_副本`,
      isDefault: false
    }
    delete newTemplate.ID
    delete newTemplate.createdAt
    delete newTemplate.updatedAt
    
    const res = await createTemplate(newTemplate)
    if (res.code === 0) {
      ElMessage.success('复制成功')
      getTemplateList()
    }
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败')
  }
}

// 删除模板
const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${row.name}" 吗？`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    const res = await deleteTemplate(row.ID)
    if (res.code === 0) {
      ElMessage.success('删除成功')
      getTemplateList()
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
      ElMessage.error('删除失败')
    }
  }
}

// 上传模板处理
const handleUploadTemplate = () => {
  resetUploadForm()
  uploadFormDialogVisible.value = true
}

// 文件选择变化处理
const handleFileChange = (file, fileList) => {
  console.log('文件选择变化:', file, fileList)
  
  if (!file.raw) {
    console.log('没有文件数据')
    return
  }
  
  const rawFile = file.raw
  console.log('原始文件信息:', rawFile.name, rawFile.type, rawFile.size)
  
  const isValidType = ['application/vnd.openxmlformats-officedocument.wordprocessingml.document',
                      'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet',
                      'application/pdf'].includes(rawFile.type)
  const isLt10M = rawFile.size / 1024 / 1024 < 10

  if (!isValidType) {
    ElMessage.error('只能上传 Word、Excel 或 PDF 文件!')
    // 清除文件列表
    if (uploadFileRef.value) {
      uploadFileRef.value.clearFiles()
    }
    return
  }
  if (!isLt10M) {
    ElMessage.error('文件大小不能超过 10MB!')
    // 清除文件列表
    if (uploadFileRef.value) {
      uploadFileRef.value.clearFiles()
    }
    return
  }

  // 保存文件信息
  uploadFormData.file = rawFile
  uploadFormData.fileName = rawFile.name
  
  // 根据文件类型自动设置模板类型
  if (rawFile.type === 'application/vnd.openxmlformats-officedocument.wordprocessingml.document') {
    uploadFormData.type = 'word'
  } else if (rawFile.type === 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet') {
    uploadFormData.type = 'excel'
  } else if (rawFile.type === 'application/pdf') {
    uploadFormData.type = 'pdf'
  }
  
  // 如果没有模板名称，使用文件名（去掉扩展名）
  if (!uploadFormData.name) {
    uploadFormData.name = rawFile.name.replace(/\.[^/.]+$/, '')
  }
  
  console.log('文件信息已设置:', {
    fileName: uploadFormData.fileName,
    type: uploadFormData.type,
    name: uploadFormData.name
  })
  
  // 显示成功消息
  ElMessage.success(`文件 "${rawFile.name}" 已选择`)
}

// 文件选择处理 (before-upload)
const handleFileSelect = (file) => {
  console.log('before-upload处理:', file.name, file.type, file.size)
  return false // 阻止默认上传行为
}

// 清除选中的文件
const clearSelectedFile = () => {
  uploadFormData.file = null
  uploadFormData.fileName = ''
  uploadFormData.name = ''
  uploadFormData.type = ''
  
  // 清除上传组件的文件列表
  if (uploadFileRef.value) {
    uploadFileRef.value.clearFiles()
  }
  
  // 重置表单验证状态
  if (uploadFormRef.value) {
    uploadFormRef.value.clearValidate()
  }
}

// 确认上传
const confirmUpload = async () => {
  if (!uploadFormRef.value) return
  
  try {
    await uploadFormRef.value.validate()
    
    if (!uploadFormData.file) {
      ElMessage.error('请选择要上传的文件')
      return
    }
    
    // 关闭表单对话框，开始上传
    uploadFormDialogVisible.value = false
    startUpload(uploadFormData.file, uploadFormData)
  } catch (error) {
    console.error('表单验证失败:', error)
  }
}

// 取消上传表单
const cancelUploadForm = () => {
  uploadFormDialogVisible.value = false
  resetUploadForm()
}

// 重置上传表单
const resetUploadForm = () => {
  uploadFormData.name = ''
  uploadFormData.type = ''
  uploadFormData.format = ''
  uploadFormData.description = ''
  uploadFormData.fileName = ''
  uploadFormData.file = null
  
  if (uploadFormRef.value) {
    uploadFormRef.value.resetFields()
  }
}

const startUpload = async (file, formData) => {
  uploading.value = true
  uploadDialogVisible.value = true
  uploadInfo.filename = file.name
  uploadInfo.status = '上传中...'
  uploadInfo.progress = 0

  try {
    const uploadData = new FormData()
    uploadData.append('file', file)
    uploadData.append('name', formData.name)
    uploadData.append('type', formData.type)
    uploadData.append('format', formData.format)
    if (formData.description) {
      uploadData.append('description', formData.description)
    }

    // 模拟上传进度
    const progressInterval = setInterval(() => {
      if (uploadInfo.progress < 90) {
        uploadInfo.progress += Math.random() * 10
      }
    }, 500)

    const res = await uploadTemplate(uploadData)
    
    clearInterval(progressInterval)
    uploadInfo.progress = 100
    uploadInfo.status = '上传成功'

    if (res.code === 0) {
      ElMessage.success('上传成功')
      setTimeout(() => {
        uploadDialogVisible.value = false
        resetUploadForm()
        getTemplateList()
      }, 1000)
    }
  } catch (error) {
    console.error('上传失败:', error)
    uploadInfo.status = '上传失败'
    ElMessage.error('上传失败')
  } finally {
    uploading.value = false
  }
}

const cancelUpload = () => {
  if (!uploading.value) {
    uploadDialogVisible.value = false
  }
}

// 表单提交
const handleFormSubmit = () => {
  dialogVisible.value = false
  getTemplateList()
}

const handleFormCancel = () => {
  dialogVisible.value = false
}

// 工具函数
const getTypeLabel = (type) => {
  const labels = {
    word: 'Word',
    excel: 'Excel',
    pdf: 'PDF'
  }
  return labels[type] || type
}

const getTypeTagType = (type) => {
  const types = {
    word: 'primary',
    excel: 'success',
    pdf: 'warning'
  }
  return types[type] || ''
}

// 上传文件相关工具函数
const getUploadFileTypeLabel = (type) => {
  const labels = {
    word: 'Word文档',
    excel: 'Excel表格',
    pdf: 'PDF文件'
  }
  return labels[type] || '未知格式'
}

const getUploadFileTypeTag = (type) => {
  const types = {
    word: 'primary',
    excel: 'success',
    pdf: 'warning'
  }
  return types[type] || 'info'
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const getFormatLabel = (format) => {
  const labels = {
    requirement_spec: '需求规格说明书',
    nesma_report: 'NESMA评估报告',
    business_summary: '业务需求汇总表'
  }
  return labels[format] || format
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

// 取消操作
const handleCancel = () => {
  emit('cancel')
}
</script>

<style scoped>
.template-management {
  padding: 20px;
}

.template-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.toolbar-left {
  display: flex;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.template-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.template-icon {
  font-size: 16px;
  color: var(--el-color-primary);
}

.action-buttons {
  display: flex;
  gap: 4px;
}

.pagination-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  padding: 20px;
}

.action-container {
  display: flex;
  justify-content: center;
  margin-top: 20px;
  padding: 20px;
  border-top: 1px solid var(--el-border-color);
}

.upload-progress {
  padding: 20px;
  text-align: center;
}

.upload-info {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 20px;
}

.upload-icon {
  font-size: 32px;
  color: var(--el-color-primary);
  margin-right: 16px;
}

.upload-text {
  text-align: left;
}

.upload-filename {
  font-size: 16px;
  font-weight: 500;
  color: var(--el-text-color-primary);
}

.upload-status {
  font-size: 14px;
  color: var(--el-text-color-secondary);
  margin-top: 4px;
}

.form-tip {
  margin-top: 8px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
  line-height: 1.4;
}

.file-upload-container {
  width: 100%;
}

.selected-file-info {
  margin-top: 12px;
  padding: 12px;
  background: var(--el-color-success-light-9);
  border: 1px solid var(--el-color-success-light-7);
  border-radius: 6px;
  transition: all 0.2s ease;
}

.selected-file-info:hover {
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.file-info-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.file-icon {
  color: var(--el-color-success);
  font-size: 18px;
}

.file-name {
  flex: 1;
  color: var(--el-text-color-primary);
  font-weight: 500;
  font-size: 14px;
  word-break: break-all;
}

.file-info-details {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-left: 26px;
}

.file-size {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}

.no-file-selected {
  margin-top: 8px;
}
</style> 