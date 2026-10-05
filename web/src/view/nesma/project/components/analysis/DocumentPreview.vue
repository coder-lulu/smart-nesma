<template>
  <div class="document-preview">
    <el-dialog
      v-model="visible"
      :title="dialogTitle"
      width="90%"
      :close-on-click-modal="false"
      class="preview-dialog"
    >
      <div class="preview-container">
        <!-- 预览工具栏 -->
        <div class="preview-toolbar">
          <div class="toolbar-left">
            <el-button-group>
              <el-button 
                :type="viewMode === 'html' ? 'primary' : 'default'"
                @click="viewMode = 'html'"
                size="small"
              >
                <el-icon><Monitor /></el-icon>
                HTML预览
              </el-button>
              <el-button 
                :type="viewMode === 'pdf' ? 'primary' : 'default'"
                @click="viewMode = 'pdf'"
                size="small"
              >
                <el-icon><Document /></el-icon>
                PDF预览
              </el-button>
              <el-button 
                :type="viewMode === 'mobile' ? 'primary' : 'default'"
                @click="viewMode = 'mobile'"
                size="small"
              >
                <el-icon><Cellphone /></el-icon>
                移动端
              </el-button>
            </el-button-group>
          </div>
          
          <div class="toolbar-center">
            <el-slider
              v-model="zoomLevel"
              :min="50"
              :max="200"
              :step="10"
              :format-tooltip="(val) => val + '%'"
              style="width: 200px;"
            />
            <span class="zoom-text">{{ zoomLevel }}%</span>
          </div>
          
          <div class="toolbar-right">
            <el-button size="small" @click="printPreview">
              <el-icon><Printer /></el-icon>
              打印
            </el-button>
            <el-button size="small" type="primary" @click="downloadPreview">
              <el-icon><Download /></el-icon>
              下载
            </el-button>
          </div>
        </div>
        
        <!-- 预览内容区域 -->
        <div :class="['preview-content', viewMode]" :style="contentStyle">
          <div v-if="loading" class="preview-loading">
            <el-icon class="is-loading" size="48"><Loading /></el-icon>
            <p>正在生成预览...</p>
          </div>
          
          <div v-else-if="error" class="preview-error">
            <el-icon size="48" color="#f56c6c"><Warning /></el-icon>
            <p>预览生成失败</p>
            <p class="error-message">{{ error }}</p>
            <el-button @click="retryPreview">重试</el-button>
          </div>
          
          <div v-else-if="previewContent" class="preview-wrapper">
            <!-- HTML预览 -->
            <div v-if="viewMode === 'html'" class="html-preview" v-html="previewContent.html"></div>
            
            <!-- PDF预览 -->
            <div v-else-if="viewMode === 'pdf'" class="pdf-preview">
              <iframe 
                v-if="previewContent.pdf" 
                :src="previewContent.pdf" 
                width="100%" 
                height="800px"
                frameborder="0"
              ></iframe>
              <div v-else class="pdf-placeholder">
                <el-icon size="64"><Document /></el-icon>
                <p>PDF预览暂不可用</p>
              </div>
            </div>
            
            <!-- 移动端预览 -->
            <div v-else-if="viewMode === 'mobile'" class="mobile-preview">
              <div class="mobile-frame">
                <div class="mobile-content" v-html="previewContent.mobile || previewContent.html"></div>
              </div>
            </div>
          </div>
          
          <div v-else class="preview-empty">
            <el-empty description="暂无预览内容" />
          </div>
        </div>
        
        <!-- 预览信息面板 -->
        <div v-if="previewData" class="preview-info">
          <el-collapse v-model="activeInfoPanel">
            <el-collapse-item title="文档信息" name="info">
              <div class="info-grid">
                <div class="info-item">
                  <label>文档标题:</label>
                  <span>{{ previewData.title }}</span>
                </div>
                <div class="info-item">
                  <label>文档格式:</label>
                  <span>{{ previewData.format?.toUpperCase() }}</span>
                </div>
                <div class="info-item">
                  <label>页面数:</label>
                  <span>{{ previewData.pageCount || '--' }}</span>
                </div>
                <div class="info-item">
                  <label>文件大小:</label>
                  <span>{{ formatFileSize(previewData.fileSize) }}</span>
                </div>
                <div class="info-item">
                  <label>生成时间:</label>
                  <span>{{ formatTime(previewData.createdAt) }}</span>
                </div>
                <div class="info-item">
                  <label>语言:</label>
                  <span>{{ previewData.language === 'zh-CN' ? '中文' : '英文' }}</span>
                </div>
              </div>
            </el-collapse-item>
            
            <el-collapse-item title="章节概览" name="sections">
              <div class="sections-list">
                <div 
                  v-for="(section, index) in previewData.sections" 
                  :key="index"
                  class="section-item"
                  @click="scrollToSection(section.id)"
                >
                  <span class="section-title">{{ section.title }}</span>
                  <span class="section-page">P{{ section.page || index + 1 }}</span>
                </div>
              </div>
            </el-collapse-item>
            
            <el-collapse-item title="样式设置" name="styles">
              <div class="style-controls">
                <el-form label-width="80px" size="small">
                  <el-form-item label="字体大小">
                    <el-slider 
                      v-model="fontSize" 
                      :min="12" 
                      :max="20" 
                      @change="updatePreviewStyle"
                    />
                  </el-form-item>
                  <el-form-item label="行间距">
                    <el-slider 
                      v-model="lineHeight" 
                      :min="1.2" 
                      :max="2.0" 
                      :step="0.1" 
                      @change="updatePreviewStyle"
                    />
                  </el-form-item>
                  <el-form-item label="页面边距">
                    <el-slider 
                      v-model="pageMargin" 
                      :min="10" 
                      :max="50" 
                      @change="updatePreviewStyle"
                    />
                  </el-form-item>
                </el-form>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <div class="footer-left">
            <el-tag v-if="previewData" type="info">
              最后更新: {{ formatTime(previewData.updatedAt) }}
            </el-tag>
          </div>
          <div class="footer-right">
            <el-button @click="closePreview">关闭</el-button>
            <el-button type="primary" @click="confirmPreview">确认并继续</el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { 
  Monitor, 
  Document, 
  Cellphone, 
  Printer, 
  Download, 
  Loading, 
  Warning 
} from '@element-plus/icons-vue'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  previewData: {
    type: Object,
    default: () => ({})
  },
  format: {
    type: String,
    default: 'html'
  },
  loading: {
    type: Boolean,
    default: false
  },
  error: {
    type: String,
    default: ''
  }
})

const emit = defineEmits(['update:modelValue', 'download', 'print', 'confirm', 'retry'])

// 响应式数据
const visible = ref(false)
const viewMode = ref('html')
const zoomLevel = ref(100)
const fontSize = ref(14)
const lineHeight = ref(1.6)
const pageMargin = ref(20)
const activeInfoPanel = ref(['info'])

// 计算属性
const dialogTitle = computed(() => {
  const formatNames = {
    html: 'HTML文档预览',
    pdf: 'PDF文档预览',
    word: 'Word文档预览',
    excel: 'Excel文档预览'
  }
  return formatNames[props.format] || '文档预览'
})

const previewContent = computed(() => {
  if (!props.previewData) return null
  
  return {
    html: props.previewData.htmlContent || props.previewData.content,
    pdf: props.previewData.pdfUrl || props.previewData.downloadUrl,
    mobile: props.previewData.mobileContent
  }
})

const contentStyle = computed(() => {
  const baseStyles = {
    transform: `scale(${zoomLevel.value / 100})`,
    transformOrigin: 'top left'
  }
  
  if (viewMode.value === 'html') {
    return {
      ...baseStyles,
      fontSize: `${fontSize.value}px`,
      lineHeight: lineHeight.value,
      padding: `${pageMargin.value}px`
    }
  }
  
  return baseStyles
})

// 监听属性变化
watch(() => props.modelValue, (newVal) => {
  visible.value = newVal
})

watch(visible, (newVal) => {
  emit('update:modelValue', newVal)
})

watch(() => props.format, (newFormat) => {
  // 根据文档格式调整默认视图模式
  if (newFormat === 'pdf') {
    viewMode.value = 'pdf'
  } else {
    viewMode.value = 'html'
  }
})

// 方法实现
const updatePreviewStyle = () => {
  // 更新预览样式
  nextTick(() => {
    const previewEl = document.querySelector('.html-preview')
    if (previewEl) {
      previewEl.style.fontSize = `${fontSize.value}px`
      previewEl.style.lineHeight = lineHeight.value
      previewEl.style.padding = `${pageMargin.value}px`
    }
  })
}

const scrollToSection = (sectionId) => {
  const element = document.querySelector(`#${sectionId}`)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth' })
  }
}

const printPreview = () => {
  if (viewMode.value === 'html' && previewContent.value?.html) {
    const printWindow = window.open('', '_blank')
    printWindow.document.write(`
      <html>
        <head>
          <title>文档打印</title>
          <style>
            body { 
              font-family: 'Microsoft YaHei', sans-serif;
              font-size: ${fontSize.value}px;
              line-height: ${lineHeight.value};
              margin: ${pageMargin.value}px;
            }
            @media print {
              body { margin: 0; }
              .no-print { display: none; }
            }
          </style>
        </head>
        <body>
          ${previewContent.value.html}
        </body>
      </html>
    `)
    printWindow.document.close()
    printWindow.print()
  } else {
    emit('print', { format: props.format, data: props.previewData })
  }
}

const downloadPreview = () => {
  emit('download', { 
    format: props.format, 
    data: props.previewData,
    content: previewContent.value
  })
}

const retryPreview = () => {
  emit('retry')
}

const closePreview = () => {
  visible.value = false
}

const confirmPreview = () => {
  emit('confirm', {
    format: props.format,
    data: props.previewData,
    viewSettings: {
      fontSize: fontSize.value,
      lineHeight: lineHeight.value,
      pageMargin: pageMargin.value
    }
  })
  closePreview()
}

// 辅助方法
const formatFileSize = (bytes) => {
  if (!bytes) return '--'
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
}

const formatTime = (time) => {
  if (!time) return '--'
  return new Date(time).toLocaleString('zh-CN')
}
</script>

<style lang="scss" scoped>
.document-preview {
  .preview-dialog {
    .preview-container {
      display: flex;
      flex-direction: column;
      height: 80vh;
      
      .preview-toolbar {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 12px 16px;
        background: #f8f9fa;
        border-radius: 6px;
        margin-bottom: 16px;
        border: 1px solid #e9ecef;
        
        .toolbar-left {
          display: flex;
          align-items: center;
          gap: 12px;
        }
        
        .toolbar-center {
          display: flex;
          align-items: center;
          gap: 12px;
          
          .zoom-text {
            font-size: 12px;
            color: #606266;
            min-width: 40px;
            text-align: center;
          }
        }
        
        .toolbar-right {
          display: flex;
          gap: 8px;
        }
      }
      
      .preview-content {
        flex: 1;
        display: flex;
        gap: 16px;
        min-height: 0;
        
        .preview-wrapper {
          flex: 1;
          border: 1px solid #e4e7ed;
          border-radius: 6px;
          overflow: hidden;
          background: white;
          
          .html-preview {
            padding: 20px;
            overflow-y: auto;
            height: 100%;
            background: white;
            
            :deep(h1), :deep(h2), :deep(h3), :deep(h4), :deep(h5), :deep(h6) {
              color: #303133;
              margin: 1.5em 0 1em 0;
            }
            
            :deep(p) {
              margin: 1em 0;
              color: #606266;
            }
            
            :deep(table) {
              width: 100%;
              border-collapse: collapse;
              margin: 1em 0;
              
              th, td {
                border: 1px solid #e4e7ed;
                padding: 8px 12px;
                text-align: left;
              }
              
              th {
                background: #f5f7fa;
                font-weight: 600;
              }
            }
            
            :deep(img) {
              max-width: 100%;
              height: auto;
              border-radius: 4px;
            }
          }
          
          .pdf-preview {
            height: 100%;
            background: #525659;
            
            iframe {
              width: 100%;
              height: 100%;
              border: none;
            }
            
            .pdf-placeholder {
              display: flex;
              flex-direction: column;
              align-items: center;
              justify-content: center;
              height: 100%;
              color: #909399;
              
              .el-icon {
                margin-bottom: 16px;
              }
            }
          }
          
          .mobile-preview {
            display: flex;
            justify-content: center;
            align-items: flex-start;
            padding: 20px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            
            .mobile-frame {
              width: 375px;
              height: 667px;
              background: white;
              border-radius: 20px;
              box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
              overflow: hidden;
              position: relative;
              
              &::before {
                content: '';
                position: absolute;
                top: 20px;
                left: 50%;
                transform: translateX(-50%);
                width: 60px;
                height: 4px;
                background: #333;
                border-radius: 2px;
                z-index: 10;
              }
              
              .mobile-content {
                height: 100%;
                overflow-y: auto;
                padding: 40px 20px 20px;
                font-size: 14px;
                line-height: 1.6;
              }
            }
          }
        }
        
        &.html .preview-wrapper {
          .html-preview {
            transform-origin: top left;
          }
        }
        
        &.pdf .preview-wrapper {
          .pdf-preview {
            transform-origin: top left;
          }
        }
        
        &.mobile .preview-wrapper {
          justify-content: center;
          
          .mobile-preview {
            transform-origin: center top;
          }
        }
        
        .preview-loading {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 300px;
          color: #909399;
          
          .el-icon {
            margin-bottom: 16px;
          }
          
          p {
            margin: 0;
            font-size: 14px;
          }
        }
        
        .preview-error {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          height: 300px;
          
          .el-icon {
            margin-bottom: 16px;
          }
          
          p {
            margin: 0 0 8px 0;
            
            &.error-message {
              font-size: 12px;
              color: #909399;
              margin-bottom: 16px;
            }
          }
        }
        
        .preview-empty {
          display: flex;
          align-items: center;
          justify-content: center;
          height: 300px;
        }
      }
      
      .preview-info {
        width: 280px;
        flex-shrink: 0;
        
        .el-collapse {
          border: 1px solid #e4e7ed;
          border-radius: 6px;
          
          .info-grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 12px;
            
            .info-item {
              display: flex;
              justify-content: space-between;
              align-items: center;
              padding: 8px 0;
              border-bottom: 1px solid #f0f0f0;
              
              &:last-child {
                border-bottom: none;
              }
              
              label {
                font-weight: 500;
                color: #606266;
                font-size: 12px;
              }
              
              span {
                color: #303133;
                font-size: 12px;
                text-align: right;
              }
            }
          }
          
          .sections-list {
            max-height: 200px;
            overflow-y: auto;
            
            .section-item {
              display: flex;
              justify-content: space-between;
              align-items: center;
              padding: 8px 12px;
              cursor: pointer;
              transition: background 0.3s;
              
              &:hover {
                background: #f5f7fa;
              }
              
              .section-title {
                flex: 1;
                font-size: 12px;
                color: #303133;
              }
              
              .section-page {
                font-size: 11px;
                color: #909399;
              }
            }
          }
          
          .style-controls {
            padding: 8px 0;
          }
        }
      }
    }
    
    .dialog-footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .footer-left {
        flex: 1;
      }
      
      .footer-right {
        display: flex;
        gap: 12px;
      }
    }
  }
}

// 动画效果
@keyframes loading {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

.is-loading {
  animation: loading 1s linear infinite;
}
</style>