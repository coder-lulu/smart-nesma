<template>
  <div class="workflow-designer">
    <!-- 顶部工具栏 -->
    <div class="designer-header">
      <div class="header-left">
        <el-breadcrumb separator="/">
          <el-breadcrumb-item :to="{ path: '/nesma/workflow' }">工作流管理</el-breadcrumb-item>
          <el-breadcrumb-item>{{ workflowName || '工作流设计器' }}</el-breadcrumb-item>
        </el-breadcrumb>
      </div>
      <div class="header-actions">
        <el-input 
          v-model="workflowName" 
          placeholder="工作流名称"
          style="width: 200px; margin-right: 12px;"
        />
        <el-button @click="saveWorkflow" type="primary" :icon="Check" :loading="saving">
          保存
        </el-button>
        <el-button @click="previewWorkflow" :icon="View">
          预览
        </el-button>
        <el-button @click="executeWorkflow" type="success" :icon="VideoPlay">
          执行
        </el-button>
        <el-button @click="goBack" :icon="Back">
          返回
        </el-button>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <div class="designer-content">
      <!-- 左侧节点面板 -->
      <div class="node-panel">
        <div class="panel-header">
          <h3>节点库</h3>
        </div>
        <div class="node-categories">
          <div v-for="category in nodeCategories" :key="category.name" class="category">
            <div class="category-header" @click="toggleCategory(category.name)">
              <span>{{ category.label }}</span>
              <el-icon><ArrowDown v-if="expandedCategories[category.name]" /><ArrowRight v-else /></el-icon>
            </div>
            <div v-show="expandedCategories[category.name]" class="category-nodes">
              <div 
                v-for="node in category.nodes" 
                :key="node.type"
                class="node-item"
                :draggable="true"
                @dragstart="handleNodeDragStart($event, node)"
              >
                <div class="node-icon" :style="{ background: node.color }">
                  <el-icon><component :is="node.icon" /></el-icon>
                </div>
                <span class="node-name">{{ node.name }}</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- 中间画布区域 -->
      <div class="canvas-area">
        <div class="canvas-toolbar">
          <el-button-group>
            <el-button @click="zoomIn" :icon="ZoomIn" />
            <el-button @click="zoomOut" :icon="ZoomOut" />
            <el-button @click="resetZoom" :icon="Refresh" />
          </el-button-group>
          <el-button @click="autoLayout" :icon="Rank">自动布局</el-button>
          <el-button @click="clearCanvas" :icon="Delete">清空画布</el-button>
        </div>
        <div ref="canvasContainer" class="canvas-container" @drop="handleDrop" @dragover="handleDragOver"></div>
      </div>

      <!-- 右侧属性面板 -->
      <div class="property-panel">
        <div class="panel-header">
          <h3>属性配置</h3>
        </div>
        <div v-if="selectedElement" class="property-content">
          <!-- 节点属性 -->
          <template v-if="selectedElement.type === 'node'">
            <el-form :model="nodeProperties" label-width="80px" size="small">
              <el-form-item label="节点名称">
                <el-input v-model="nodeProperties.name" @change="updateNodeProperties" />
              </el-form-item>
              <el-form-item label="描述">
                <el-input 
                  v-model="nodeProperties.description" 
                  type="textarea" 
                  :rows="3"
                  @change="updateNodeProperties"
                />
              </el-form-item>
              <el-form-item label="超时(秒)">
                <el-input-number 
                  v-model="nodeProperties.timeout" 
                  :min="1" 
                  :max="3600"
                  @change="updateNodeProperties"
                />
              </el-form-item>
              <el-form-item label="重试次数">
                <el-input-number 
                  v-model="nodeProperties.retries" 
                  :min="0" 
                  :max="10"
                  @change="updateNodeProperties"
                />
              </el-form-item>
            </el-form>
          </template>
          <!-- 连线属性 -->
          <template v-else-if="selectedElement.type === 'edge'">
            <el-form :model="edgeProperties" label-width="80px" size="small">
              <el-form-item label="连线名称">
                <el-input v-model="edgeProperties.name" @change="updateEdgeProperties" />
              </el-form-item>
              <el-form-item label="条件">
                <el-input 
                  v-model="edgeProperties.condition" 
                  placeholder="执行条件"
                  @change="updateEdgeProperties"
                />
              </el-form-item>
            </el-form>
          </template>
        </div>
        <div v-else class="no-selection">
          <el-empty description="请选择节点或连线" />
        </div>
      </div>
    </div>

    <!-- 保存对话框 -->
    <el-dialog v-model="saveDialogVisible" title="保存工作流" width="500px">
      <el-form :model="saveForm" label-width="100px">
        <el-form-item label="工作流名称" required>
          <el-input v-model="saveForm.name" />
        </el-form-item>
        <el-form-item label="版本">
          <el-input v-model="saveForm.version" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="saveForm.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="saveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="confirmSave" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { 
  Check, View, VideoPlay, Back, ArrowDown, ArrowRight, ZoomIn, ZoomOut, 
  Refresh, Rank, Delete, Document, Operation, Search, Edit, Connection 
} from '@element-plus/icons-vue'
import LogicFlow from '@logicflow/core'
import '@logicflow/core/dist/index.css'

const router = useRouter()
const route = useRoute()

// 响应式数据
const canvasContainer = ref(null)
const workflowName = ref('新建工作流')
const saving = ref(false)
const saveDialogVisible = ref(false)
const selectedElement = ref(null)
const expandedCategories = ref({
  'control': true,
  'agent': true
})

// LogicFlow实例
let lf = null

// 节点属性
const nodeProperties = ref({
  name: '',
  description: '',
  timeout: 60,
  retries: 3
})

// 连线属性
const edgeProperties = ref({
  name: '',
  condition: ''
})

// 保存表单
const saveForm = ref({
  name: '',
  version: '1.0.0',
  description: ''
})

// 节点分类
const nodeCategories = [
  {
    name: 'control',
    label: '控制节点',
    nodes: [
      { type: 'start', name: '开始', icon: 'VideoPlay', color: '#52c41a' },
      { type: 'end', name: '结束', icon: 'Check', color: '#f5222d' },
      { type: 'condition', name: '条件判断', icon: 'Rank', color: '#fa8c16' }
    ]
  },
  {
    name: 'agent',
    label: 'AI Agent',
    nodes: [
      { type: 'requirement', name: '需求分析', icon: 'Document', color: '#1890ff' },
      { type: 'nesma', name: 'NESMA评估', icon: 'Operation', color: '#722ed1' },
      { type: 'knowledge', name: '知识检索', icon: 'Search', color: '#fa541c' },
      { type: 'document', name: '文档生成', icon: 'Edit', color: '#13c2c2' }
    ]
  }
]

// 初始化LogicFlow
const initLogicFlow = () => {
  lf = new LogicFlow({
    container: canvasContainer.value,
    width: canvasContainer.value.offsetWidth,
    height: canvasContainer.value.offsetHeight,
    grid: {
      size: 20,
      type: 'dot'
    },
    keyboard: {
      enabled: true
    }
  })

  // 监听事件
  lf.on('node:click', handleNodeClick)
  lf.on('edge:click', handleEdgeClick)
  lf.on('canvas:click', handleCanvasClick)

  lf.render()
}

// 切换分类展开状态
const toggleCategory = (categoryName) => {
  expandedCategories.value[categoryName] = !expandedCategories.value[categoryName]
}

// 处理节点拖拽开始
const handleNodeDragStart = (event, node) => {
  event.dataTransfer.setData('application/json', JSON.stringify(node))
}

// 处理放置
const handleDrop = (event) => {
  event.preventDefault()
  const nodeData = JSON.parse(event.dataTransfer.getData('application/json'))
  const rect = canvasContainer.value.getBoundingClientRect()
  const x = event.clientX - rect.left
  const y = event.clientY - rect.top
  
  lf.addNode({
    type: 'rect',
    x,
    y,
    text: nodeData.name,
    properties: {
      nodeType: nodeData.type,
      name: nodeData.name,
      timeout: 60,
      retries: 3
    }
  })
}

const handleDragOver = (event) => {
  event.preventDefault()
}

// 处理节点点击
const handleNodeClick = ({ data }) => {
  selectedElement.value = { type: 'node', data }
  nodeProperties.value = { ...data.properties }
}

// 处理连线点击
const handleEdgeClick = ({ data }) => {
  selectedElement.value = { type: 'edge', data }
  edgeProperties.value = { ...data.properties }
}

// 处理画布点击
const handleCanvasClick = () => {
  selectedElement.value = null
}

// 更新节点属性
const updateNodeProperties = () => {
  if (selectedElement.value?.type === 'node') {
    lf.setProperties(selectedElement.value.data.id, nodeProperties.value)
  }
}

// 更新连线属性
const updateEdgeProperties = () => {
  if (selectedElement.value?.type === 'edge') {
    lf.setProperties(selectedElement.value.data.id, edgeProperties.value)
  }
}

// 画布操作
const zoomIn = () => lf.zoom(true)
const zoomOut = () => lf.zoom(false)
const resetZoom = () => lf.resetZoom()
const autoLayout = () => {
  ElMessage.info('自动布局功能开发中')
}
const clearCanvas = () => {
  lf.clearData()
  selectedElement.value = null
}

// 保存工作流
const saveWorkflow = () => {
  saveForm.value.name = workflowName.value
  saveDialogVisible.value = true
}

// 确认保存
const confirmSave = async () => {
  try {
    saving.value = true
    const graphData = lf.getGraphData()
    
    // 这里添加保存逻辑
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    ElMessage.success('保存成功')
    saveDialogVisible.value = false
  } catch (error) {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

// 预览工作流
const previewWorkflow = () => {
  ElMessage.info('预览功能开发中')
}

// 执行工作流
const executeWorkflow = () => {
  ElMessage.info('执行功能开发中')
}

// 返回
const goBack = () => {
  router.push('/nesma/workflow')
}

// 生命周期
onMounted(() => {
  initLogicFlow()
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
})

const handleResize = () => {
  if (lf && canvasContainer.value) {
    lf.resize(canvasContainer.value.offsetWidth, canvasContainer.value.offsetHeight)
  }
}
</script>

<style scoped>
.workflow-designer {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: #f5f7fa;
}

.designer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 2px 4px rgba(0,0,0,0.04);
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.designer-content {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.node-panel {
  width: 280px;
  background: white;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.panel-header {
  padding: 20px 24px;
  border-bottom: 1px solid #e4e7ed;
  background: #f8fafc;
}

.panel-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.node-categories {
  flex: 1;
  padding: 16px;
  overflow-y: auto;
}

.category {
  margin-bottom: 16px;
}

.category-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f3f4f6;
  border-radius: 6px;
  cursor: pointer;
  font-weight: 500;
  color: #374151;
}

.category-nodes {
  margin-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.node-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  cursor: grab;
  background: white;
  transition: all 0.2s;
}

.node-item:hover {
  border-color: #3b82f6;
  box-shadow: 0 2px 8px rgba(59, 130, 246, 0.15);
  transform: translateY(-1px);
}

.node-item:active {
  cursor: grabbing;
}

.node-icon {
  width: 32px;
  height: 32px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.node-name {
  font-size: 14px;
  font-weight: 500;
  color: #374151;
}

.canvas-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: white;
  margin: 16px;
  border-radius: 12px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  overflow: hidden;
}

.canvas-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 20px;
  background: #f8fafc;
  border-bottom: 1px solid #e5e7eb;
}

.canvas-container {
  flex: 1;
  position: relative;
  background: 
    radial-gradient(circle, #d1d5db 1px, transparent 1px);
  background-size: 20px 20px;
}

.property-panel {
  width: 320px;
  background: white;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.property-content {
  flex: 1;
  padding: 24px;
  overflow-y: auto;
}

.no-selection {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .node-panel {
    width: 240px;
  }
  
  .property-panel {
    width: 280px;
  }
}

@media (max-width: 768px) {
  .designer-header {
    flex-direction: column;
    gap: 12px;
    padding: 12px 16px;
  }
  
  .header-actions {
    width: 100%;
    justify-content: center;
  }
  
  .node-panel,
  .property-panel {
    display: none;
  }
  
  .canvas-area {
    margin: 8px;
  }
}

/* 表单样式增强 */
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #374151;
}

:deep(.el-input__wrapper) {
  border-radius: 6px;
}

:deep(.el-textarea__inner) {
  border-radius: 6px;
}

:deep(.el-button-group .el-button) {
  border-radius: 0;
}

:deep(.el-button-group .el-button:first-child) {
  border-top-left-radius: 6px;
  border-bottom-left-radius: 6px;
}

:deep(.el-button-group .el-button:last-child) {
  border-top-right-radius: 6px;
  border-bottom-right-radius: 6px;
}
</style> 