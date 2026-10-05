<template>
  <div class="document-preview">
    <div class="preview-toolbar">
      <div class="toolbar-left">
        <el-button-group>
          <el-button 
            size="small" 
            @click="handleZoomOut"
            :disabled="zoomLevel <= 50"
            icon="ZoomOut"
          />
          <el-button size="small" disabled>{{ zoomLevel }}%</el-button>
          <el-button 
            size="small" 
            @click="handleZoomIn"
            :disabled="zoomLevel >= 200"
            icon="ZoomIn"
          />
        </el-button-group>
        <el-button size="small" @click="handleFitWidth" icon="FullScreen">
          适应宽度
        </el-button>
        <el-button size="small" @click="handleActualSize" icon="Crop">
          实际大小
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-button size="small" @click="handlePrint" icon="Printer">
          打印
        </el-button>
        <el-button size="small" type="primary" @click="handleDownload" icon="Download">
          下载
        </el-button>
      </div>
    </div>

    <div class="preview-content" v-loading="loading">
      <div v-if="!loading && documentInfo" class="document-info">
        <div class="info-header">
          <div class="document-title">
            <el-icon class="document-icon">
              <Document v-if="documentInfo.type === 'word'" />
              <Grid v-else-if="documentInfo.type === 'excel'" />
              <DocumentCopy v-else />
            </el-icon>
            <span>{{ documentInfo.name }}</span>
          </div>
          <div class="document-meta">
            <el-tag size="small" type="info">{{ getTypeLabel(documentInfo.type) }}</el-tag>
            <el-tag size="small" type="success">{{ documentInfo.version }}</el-tag>
            <span class="meta-item">{{ formatFileSize(documentInfo.fileSize) }}</span>
            <span class="meta-item">{{ formatDate(documentInfo.generatedAt) }}</span>
          </div>
        </div>
      </div>

      <div class="preview-container" :style="{ zoom: zoomLevel / 100 }">
        <!-- PDF预览 -->
        <div v-if="documentInfo?.type === 'pdf'" class="pdf-preview">
          <iframe
            v-if="previewUrl"
            :src="previewUrl"
            width="100%"
            height="800px"
            frameborder="0"
          />
          <div v-else class="preview-placeholder">
            <el-icon><DocumentCopy /></el-icon>
            <p>PDF预览加载中...</p>
          </div>
        </div>

        <!-- Word文档预览 -->
        <div v-else-if="documentInfo?.type === 'word'" class="word-preview">
          <div v-if="previewContent" class="word-content" v-html="previewContent"></div>
          <div v-else class="preview-placeholder">
            <el-icon><Document /></el-icon>
            <p>Word文档预览加载中...</p>
          </div>
        </div>

        <!-- Excel表格预览 -->
        <div v-else-if="documentInfo?.type === 'excel'" class="excel-preview">
          <div v-if="previewContent" class="excel-content">
            <el-tabs v-model="activeSheet" type="card">
              <el-tab-pane
                v-for="(sheet, index) in previewContent.sheets"
                :key="index"
                :label="sheet.name"
                :name="sheet.name"
              >
                <div class="excel-table">
                  <table class="sheet-table">
                    <thead>
                      <tr>
                        <th v-for="(header, colIndex) in sheet.headers" :key="colIndex">
                          {{ header }}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="(row, rowIndex) in sheet.rows" :key="rowIndex">
                        <td v-for="(cell, colIndex) in row" :key="colIndex">
                          {{ cell }}
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </el-tab-pane>
            </el-tabs>
          </div>
          <div v-else class="preview-placeholder">
            <el-icon><Grid /></el-icon>
            <p>Excel表格预览加载中...</p>
          </div>
        </div>

        <!-- 预览失败 -->
        <div v-else-if="previewError" class="preview-error">
          <el-result
            icon="error"
            title="预览失败"
            :sub-title="previewError"
          >
            <template #extra>
              <el-button @click="handleRetry" type="primary">重试</el-button>
              <el-button @click="handleDownload">直接下载</el-button>
            </template>
          </el-result>
        </div>

        <!-- 初始状态 -->
        <div v-else class="preview-placeholder">
          <el-icon><Loading /></el-icon>
          <p>正在加载预览...</p>
        </div>
      </div>
    </div>

    <!-- 页面控制（PDF专用） -->
    <div v-if="documentInfo?.type === 'pdf'" class="page-controls">
      <div class="page-info">
        <el-button-group>
          <el-button 
            size="small" 
            @click="handlePrevPage"
            :disabled="currentPage <= 1"
            icon="ArrowLeft"
          />
          <el-input
            v-model="currentPage"
            size="small"
            style="width: 60px; text-align: center;"
            @keyup.enter="handleGoToPage"
          />
          <span class="page-separator">/</span>
          <span class="total-pages">{{ totalPages }}</span>
          <el-button 
            size="small" 
            @click="handleNextPage"
            :disabled="currentPage >= totalPages"
            icon="ArrowRight"
          />
        </el-button-group>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Document,
  Grid,
  DocumentCopy,
  ZoomOut,
  ZoomIn,
  FullScreen,
  Crop,
  Printer,
  Download,
  Loading,
  ArrowLeft,
  ArrowRight
} from '@element-plus/icons-vue'
import {
  getDocument,
  previewDocument,
  downloadDocument
} from '@/api/nesma'

// Props
const props = defineProps({
  documentId: {
    type: [String, Number],
    required: true
  }
})

// 响应式数据
const loading = ref(false)
const documentInfo = ref(null)
const previewUrl = ref('')
const previewContent = ref(null)
const previewError = ref('')
const zoomLevel = ref(100)
const currentPage = ref(1)
const totalPages = ref(1)
const activeSheet = ref('')

// 监听documentId变化
watch(() => props.documentId, (newId) => {
  if (newId) {
    loadDocument()
  }
}, { immediate: true })

// 加载文档信息
const loadDocument = async () => {
  if (!props.documentId) return

  loading.value = true
  previewError.value = ''
  
  try {
    // 获取文档信息
    const res = await getDocument(props.documentId)
    if (res.code === 0) {
      documentInfo.value = res.data
      
      // 根据文档类型加载预览
      if (documentInfo.value.status === 'completed') {
        await loadPreview()
      } else {
        previewError.value = '文档还未生成完成，无法预览'
      }
    } else {
      previewError.value = res.msg || '获取文档信息失败'
    }
  } catch (error) {
    console.error('加载文档失败:', error)
    previewError.value = '加载文档失败'
  } finally {
    loading.value = false
  }
}

// 加载预览内容
const loadPreview = async () => {
  if (!documentInfo.value) return

  try {
    const res = await previewDocument({
      documentId: props.documentId,
      type: documentInfo.value.type
    })
    
    if (res.code === 0) {
      switch (documentInfo.value.type) {
        case 'pdf':
          previewUrl.value = res.data.url
          totalPages.value = res.data.totalPages || 1
          break
        case 'word':
          previewContent.value = res.data.htmlContent
          break
        case 'excel':
          previewContent.value = res.data
          if (res.data.sheets && res.data.sheets.length > 0) {
            activeSheet.value = res.data.sheets[0].name
          }
          break
      }
    } else {
      previewError.value = res.msg || '预览加载失败'
    }
  } catch (error) {
    console.error('预览加载失败:', error)
    previewError.value = '预览加载失败'
  }
}

// 缩放控制
const handleZoomIn = () => {
  if (zoomLevel.value < 200) {
    zoomLevel.value += 10
  }
}

const handleZoomOut = () => {
  if (zoomLevel.value > 50) {
    zoomLevel.value -= 10
  }
}

const handleFitWidth = () => {
  zoomLevel.value = 100
}

const handleActualSize = () => {
  zoomLevel.value = 100
}

// 页面控制
const handlePrevPage = () => {
  if (currentPage.value > 1) {
    currentPage.value--
  }
}

const handleNextPage = () => {
  if (currentPage.value < totalPages.value) {
    currentPage.value++
  }
}

const handleGoToPage = () => {
  const page = parseInt(currentPage.value)
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
  } else {
    currentPage.value = Math.max(1, Math.min(totalPages.value, page))
  }
}

// 操作按钮
const handlePrint = () => {
  if (documentInfo.value?.type === 'pdf' && previewUrl.value) {
    // 打开PDF打印
    const printWindow = window.open(previewUrl.value, '_blank')
    if (printWindow) {
      printWindow.onload = () => {
        printWindow.print()
      }
    }
  } else {
    // 其他类型的打印
    window.print()
  }
}

import { downloadNesmaDocument } from '@/utils/officeDownload'

// ... existing code ...

const handleDownload = async () => {
  if (!documentInfo.value) return

  try {
    const res = await downloadDocument({ documentId: props.documentId })
    
    // 使用专门的 NESMA 文档下载工具
    downloadNesmaDocument(
      new Blob([res], { type: 'application/octet-stream' }),
      documentInfo.value
    )
    
    ElMessage.success('下载成功')
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败')
  }
}

const handleRetry = () => {
  loadDocument()
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

const getFileExtension = (type) => {
  const extensions = {
    word: 'docx',
    excel: 'xlsx',
    pdf: 'pdf'
  }
  return extensions[type] || 'txt'
}

const formatFileSize = (size) => {
  if (!size) return '-'
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  return (size / (1024 * 1024)).toFixed(1) + ' MB'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

// 生命周期
onMounted(() => {
  if (props.documentId) {
    loadDocument()
  }
})
</script>

<style scoped>
.document-preview {
  height: 80vh;
  display: flex;
  flex-direction: column;
}

.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: var(--el-bg-color);
  border-bottom: 1px solid var(--el-border-color);
}

.toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.toolbar-right {
  display: flex;
  gap: 8px;
}

.preview-content {
  flex: 1;
  overflow: hidden;
  position: relative;
}

.document-info {
  padding: 16px;
  background: var(--el-bg-color-page);
  border-bottom: 1px solid var(--el-border-color);
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.document-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 500;
}

.document-icon {
  font-size: 20px;
  color: var(--el-color-primary);
}

.document-meta {
  display: flex;
  align-items: center;
  gap: 12px;
}

.meta-item {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.preview-container {
  flex: 1;
  overflow: auto;
  padding: 20px;
  background: #f5f5f5;
}

.pdf-preview iframe {
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.word-preview {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 40px;
  margin: 0 auto;
  max-width: 800px;
}

.word-content {
  line-height: 1.6;
  font-size: 14px;
}

.excel-preview {
  background: white;
  border-radius: 8px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  padding: 20px;
}

.excel-table {
  overflow: auto;
  max-height: 600px;
}

.sheet-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}

.sheet-table th,
.sheet-table td {
  border: 1px solid var(--el-border-color);
  padding: 8px 12px;
  text-align: left;
}

.sheet-table th {
  background: var(--el-bg-color);
  font-weight: 500;
  position: sticky;
  top: 0;
  z-index: 1;
}

.sheet-table tr:hover {
  background: var(--el-bg-color-page);
}

.preview-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 400px;
  color: var(--el-text-color-secondary);
}

.preview-placeholder .el-icon {
  font-size: 48px;
  margin-bottom: 16px;
}

.preview-error {
  padding: 40px;
}

.page-controls {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 12px;
  background: var(--el-bg-color);
  border-top: 1px solid var(--el-border-color);
}

.page-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.page-separator {
  margin: 0 8px;
  color: var(--el-text-color-secondary);
}

.total-pages {
  color: var(--el-text-color-secondary);
  min-width: 30px;
  text-align: center;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .preview-toolbar {
    flex-direction: column;
    gap: 12px;
  }
  
  .toolbar-left,
  .toolbar-right {
    justify-content: center;
  }
  
  .document-info .info-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 12px;
  }
  
  .word-preview {
    padding: 20px;
  }
  
  .excel-preview {
    padding: 10px;
  }
}
</style> 