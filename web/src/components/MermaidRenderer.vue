<template>
  <div class="mermaid-renderer">
    <div 
      class="mermaid-container"
      :style="{ transform: `scale(${zoomLevel})` }"
    >
      <div 
        ref="mermaidRef" 
        class="mermaid-content"
        v-if="mermaidCode"
      >
        <!-- Mermaid内容将在这里渲染 -->
      </div>
      <div v-else class="mermaid-placeholder">
        <el-empty description="没有流程图代码">
          <template #image>
            <el-icon size="60" color="#d9d9d9"><Coordinate /></el-icon>
          </template>
        </el-empty>
      </div>
    </div>
    
    <!-- 控制工具栏 -->
    <div class="mermaid-controls" v-if="showControls && mermaidCode">
      <el-button-group size="small">
        <el-button @click="zoomIn" :disabled="zoomLevel >= 2">
          <el-icon><ZoomIn /></el-icon>
          放大
        </el-button>
        <el-button @click="zoomOut" :disabled="zoomLevel <= 0.5">
          <el-icon><ZoomOut /></el-icon>
          缩小
        </el-button>
        <el-button @click="resetZoom">
          <el-icon><FullScreen /></el-icon>
          重置
        </el-button>
        <el-button @click="downloadSVG" type="primary">
          <el-icon><Download /></el-icon>
          下载SVG
        </el-button>
      </el-button-group>
    </div>
    
    <!-- 渲染状态 -->
    <div class="render-status" v-if="renderStatus !== 'success'">
      <el-alert 
        v-if="renderStatus === 'loading'"
        title="正在渲染流程图..." 
        type="info" 
        :closable="false" 
        show-icon
      />
      <el-alert 
        v-else-if="renderStatus === 'error'"
        :title="'渲染失败: ' + errorMessage" 
        type="error" 
        :closable="false" 
        show-icon
      />
    </div>
  </div>
</template>

<script setup>
import { ref, watch, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { Coordinate, ZoomIn, ZoomOut, FullScreen, Download } from '@element-plus/icons-vue'

const props = defineProps({
  mermaidCode: {
    type: String,
    default: ''
  },
  showControls: {
    type: Boolean,
    default: true
  },
  theme: {
    type: String,
    default: 'default' // default, dark, forest, neutral
  },
  autoRender: {
    type: Boolean,
    default: true
  }
})

const emit = defineEmits(['render-success', 'render-error'])

// 响应式变量
const mermaidRef = ref(null)
const zoomLevel = ref(1)
const renderStatus = ref('idle') // idle, loading, success, error
const errorMessage = ref('')
let mermaid = null

// 初始化Mermaid
const initMermaid = async () => {
  try {
    // 动态导入mermaid
    const mermaidModule = await import('mermaid')
    mermaid = mermaidModule.default
    
    // 配置mermaid
    mermaid.initialize({
      startOnLoad: false,
      theme: props.theme,
      securityLevel: 'loose',
      fontFamily: 'arial',
      fontSize: 16,
      flowchart: {
        useMaxWidth: true,
        htmlLabels: true,
        curve: 'basis'
      },
      themeVariables: {
        primaryColor: '#409eff',
        primaryTextColor: '#303133',
        primaryBorderColor: '#409eff',
        lineColor: '#909399',
        sectionBkgColor: '#f5f7fa',
        altSectionBkgColor: '#ffffff',
        gridColor: '#e4e7ed',
        secondaryColor: '#67c23a',
        tertiaryColor: '#e6a23c'
      }
    })
    
    console.log('Mermaid初始化成功')
  } catch (error) {
    console.error('Mermaid初始化失败:', error)
    renderStatus.value = 'error'
    errorMessage.value = 'Mermaid库加载失败'
  }
}

// 渲染Mermaid图表
const renderMermaid = async () => {
  if (!props.mermaidCode || !mermaidRef.value || !mermaid) return
  
  try {
    renderStatus.value = 'loading'
    
    // 清空容器
    mermaidRef.value.innerHTML = ''
    
    // 生成唯一ID
    const id = `mermaid-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`
    
    // 验证Mermaid语法
    const isValid = await mermaid.parse(props.mermaidCode)
    if (!isValid) {
      throw new Error('Mermaid语法错误')
    }
    
    // 渲染图表
    const { svg } = await mermaid.render(id, props.mermaidCode)
    
    // 插入SVG到容器
    mermaidRef.value.innerHTML = svg
    
    // 设置SVG样式
    const svgElement = mermaidRef.value.querySelector('svg')
    if (svgElement) {
      svgElement.style.maxWidth = '100%'
      svgElement.style.height = 'auto'
      svgElement.style.background = 'transparent'
    }
    
    renderStatus.value = 'success'
    emit('render-success', { svg, element: mermaidRef.value })
    
    console.log('Mermaid渲染成功')
    
  } catch (error) {
    console.error('Mermaid渲染失败:', error)
    renderStatus.value = 'error'
    errorMessage.value = error.message || '渲染失败'
    
    // 显示错误信息
    mermaidRef.value.innerHTML = `
      <div class="render-error">
        <p>流程图渲染失败</p>
        <p class="error-detail">${error.message}</p>
        <p class="error-hint">请检查Mermaid语法是否正确</p>
      </div>
    `
    
    emit('render-error', error)
  }
}

// 缩放控制
const zoomIn = () => {
  if (zoomLevel.value < 2) {
    zoomLevel.value += 0.1
  }
}

const zoomOut = () => {
  if (zoomLevel.value > 0.5) {
    zoomLevel.value -= 0.1
  }
}

const resetZoom = () => {
  zoomLevel.value = 1
}

// 下载SVG
const downloadSVG = () => {
  const svgElement = mermaidRef.value?.querySelector('svg')
  if (!svgElement) {
    ElMessage.warning('没有可下载的流程图')
    return
  }
  
  try {
    // 获取SVG内容
    const svgData = new XMLSerializer().serializeToString(svgElement)
    const blob = new Blob([svgData], { type: 'image/svg+xml' })
    
    // 创建下载链接
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `mermaid-diagram-${Date.now()}.svg`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success('流程图SVG已下载')
  } catch (error) {
    console.error('下载失败:', error)
    ElMessage.error('下载失败: ' + error.message)
  }
}

// 监听代码变化并重新渲染
watch(() => props.mermaidCode, async () => {
  if (props.autoRender && props.mermaidCode) {
    await nextTick()
    renderMermaid()
  }
}, { immediate: false })

// 监听主题变化
watch(() => props.theme, async () => {
  if (mermaid) {
    await initMermaid()
    if (props.mermaidCode) {
      renderMermaid()
    }
  }
})

// 组件挂载后初始化
onMounted(async () => {
  await initMermaid()
  if (props.autoRender && props.mermaidCode) {
    await nextTick()
    renderMermaid()
  }
})

// 暴露方法给父组件
defineExpose({
  renderMermaid,
  zoomIn,
  zoomOut,
  resetZoom,
  downloadSVG
})
</script>

<style scoped>
.mermaid-renderer {
  width: 100%;
  position: relative;
}

.mermaid-container {
  width: 100%;
  min-height: 300px;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  background: #fafbfc;
  overflow: auto;
  transition: transform 0.3s ease;
  position: relative;
}

.mermaid-content {
  padding: 20px;
  text-align: center;
  min-height: 260px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mermaid-placeholder {
  padding: 40px 20px;
  text-align: center;
  min-height: 260px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.mermaid-controls {
  margin-top: 16px;
  text-align: center;
}

.render-status {
  margin-top: 16px;
}

/* Mermaid SVG样式优化 */
:deep(.mermaid-content svg) {
  max-width: 100%;
  height: auto;
  background: transparent;
}

/* 错误显示样式 */
.render-error {
  padding: 20px;
  text-align: center;
  color: #f56c6c;
}

.render-error p {
  margin: 8px 0;
}

.error-detail {
  font-size: 12px;
  color: #909399;
  background: #f5f7fa;
  padding: 8px 12px;
  border-radius: 4px;
  display: inline-block;
  margin: 8px 0;
}

.error-hint {
  font-size: 12px;
  color: #606266;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .mermaid-container {
    min-height: 200px;
  }
  
  .mermaid-content {
    padding: 12px;
    min-height: 176px;
  }
  
  .mermaid-controls .el-button-group .el-button {
    padding: 6px 8px;
    font-size: 12px;
  }
}
</style>