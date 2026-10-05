<template>
  <el-drawer
    v-model="drawerVisible"
    title="工作流设计器"
    size="90%"
    direction="rtl"
    :close-on-click-modal="false"
    :before-close="handleBeforeClose"
  >
    <div class="designer-container">
      <!-- 工具栏 -->
      <div class="designer-toolbar">
        <div class="toolbar-left">
          <el-input 
            v-model="workflowName" 
            placeholder="工作流名称"
            style="width: 200px; margin-right: 12px;"
          />
          <el-select
            v-model="workflowCategory"
            placeholder="选择分类"
            style="width: 140px; margin-right: 12px;"
          >
            <el-option label="需求分析" value="requirement_analysis" />
            <el-option label="NESMA评估" value="nesma_evaluation" />
            <el-option label="文档生成" value="document_generation" />
            <el-option label="质量检查" value="quality_check" />
          </el-select>
        </div>
        <div class="toolbar-right">
          <el-button @click="previewWorkflow" :icon="View" size="small">
            预览
          </el-button>
          <el-button @click="saveWorkflow" type="primary" :icon="Check" :loading="saving" size="small">
            保存
          </el-button>
          <el-button @click="executeWorkflow" type="success" :icon="VideoPlay" size="small">
            测试执行
          </el-button>
        </div>
      </div>

      <!-- 主要设计区域 -->
      <div class="designer-content">
        <!-- 左侧节点面板 -->
        <div class="node-panel">
          <div class="panel-header">
            <h4>节点库</h4>
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
            <el-button-group size="small">
              <el-button @click="zoomIn" :icon="ZoomIn" />
              <el-button @click="zoomOut" :icon="ZoomOut" />
              <el-button @click="resetZoom" :icon="Refresh" />
            </el-button-group>
            <el-button @click="autoLayout" :icon="Rank" size="small">自动布局</el-button>
            <el-button @click="clearCanvas" :icon="Delete" size="small">清空画布</el-button>
            <el-button @click="initLogicFlow" type="warning" size="small" v-if="!lf">重新初始化</el-button>
          </div>
          <!-- LogicFlow 画布 -->
          <div 
            ref="canvasContainer" 
            class="canvas-container" 
            @drop="handleDrop" 
            @dragover="handleDragOver"
          >
            <!-- 画布未初始化时的提示 -->
            <div v-if="!lf" class="canvas-placeholder">
              <el-empty description="画布正在初始化...">
                <el-button @click="initLogicFlow" type="primary">手动初始化</el-button>
              </el-empty>
            </div>
          </div>
        </div>

        <!-- 右侧属性面板 -->
        <div class="property-panel">
          <div class="panel-header">
            <h4>属性配置</h4>
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
                <el-form-item label="Agent类型">
                  <el-select v-model="nodeProperties.agentType" @change="updateNodeProperties">
                    <el-option label="需求分析" value="REQUIREMENT_ANALYSIS" />
                    <el-option label="NESMA评估" value="NESMA_EVALUATION" />
                    <el-option label="知识检索" value="KNOWLEDGE_RETRIEVAL" />
                    <el-option label="文档生成" value="DOCUMENT_GENERATION" />
                  </el-select>
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
    </div>

    <!-- 预览对话框 -->
    <el-dialog v-model="previewDialogVisible" title="工作流预览" width="600px">
      <div class="preview-content">
        <el-descriptions title="基本信息" :column="2" border>
          <el-descriptions-item label="名称">{{ workflowName }}</el-descriptions-item>
          <el-descriptions-item label="分类">{{ getCategoryLabel(workflowCategory) }}</el-descriptions-item>
          <el-descriptions-item label="节点数">{{ getNodeCount() }}</el-descriptions-item>
          <el-descriptions-item label="连线数">{{ getEdgeCount() }}</el-descriptions-item>
        </el-descriptions>
        
        <div class="workflow-preview">
          <h4>流程图</h4>
          <div class="mini-canvas" ref="previewCanvas"></div>
        </div>

        <div class="workflow-definition">
          <h4>JSON定义</h4>
          <el-input 
            v-model="workflowJSON" 
            type="textarea" 
            :rows="10" 
            readonly
            class="json-viewer"
          />
        </div>
      </div>
    </el-dialog>
  </el-drawer>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  Check, View, VideoPlay, ArrowDown, ArrowRight, ZoomIn, ZoomOut, 
  Refresh, Rank, Delete, Document, Operation, Search, Edit, Connection,
  CircleCheck
} from '@element-plus/icons-vue'
import LogicFlow from '@logicflow/core'
import '@logicflow/core/dist/index.css'

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  workflowId: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['update:visible', 'refresh'])

// 响应式数据
const drawerVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const canvasContainer = ref(null)
const previewCanvas = ref(null)
const workflowName = ref('新建工作流')
const workflowCategory = ref('')
const saving = ref(false)
const previewDialogVisible = ref(false)
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
  agentType: '',
  timeout: 60,
  retries: 3
})

// 连线属性
const edgeProperties = ref({
  name: '',
  condition: ''
})

// 节点分类
const nodeCategories = [
  {
    name: 'control',
    label: '控制节点',
    nodes: [
      {
        type: 'start',
        name: '开始',
        icon: VideoPlay,
        color: '#67c23a'
      },
      {
        type: 'end',
        name: '结束',
        icon: CircleCheck,
        color: '#909399'
      },
      {
        type: 'condition',
        name: '条件判断',
        icon: Operation,
        color: '#e6a23c'
      }
    ]
  },
  {
    name: 'agent',
    label: 'Agent节点',
    nodes: [
      {
        type: 'requirement-analysis',
        name: '需求分析',
        icon: Document,
        color: '#409eff'
      },
      {
        type: 'nesma-evaluation',
        name: 'NESMA评估',
        icon: Connection,
        color: '#67c23a'
      },
      {
        type: 'knowledge-retrieval',
        name: '知识检索',
        icon: Search,
        color: '#e6a23c'
      },
      {
        type: 'document-generation',
        name: '文档生成',
        icon: Edit,
        color: '#f56c6c'
      }
    ]
  }
]

// 计算属性
const workflowJSON = computed(() => {
  if (!lf) return ''
  try {
    const data = lf.getGraphData()
    return JSON.stringify(data, null, 2)
  } catch (error) {
    return ''
  }
})

// 方法
const initLogicFlow = (retryCount = 0) => {
  if (!canvasContainer.value) {
    console.warn('Canvas container not found')
    if (retryCount < 5) {
      setTimeout(() => {
        initLogicFlow(retryCount + 1)
      }, 200)
    }
    return
  }
  
  // 如果已经初始化过，先销毁
  if (lf) {
    lf.destroy()
    lf = null
  }
  
  // 获取画布容器的实际尺寸
  const rect = canvasContainer.value.getBoundingClientRect()
  
  // 确保容器有有效的尺寸
  if (rect.width === 0 || rect.height === 0) {
    console.warn(`Canvas container has no size (${rect.width}x${rect.height}), retrying...`)
    if (retryCount < 10) {
      setTimeout(() => {
        initLogicFlow(retryCount + 1)
      }, 200)
    } else {
      ElMessage.error('画布初始化失败：容器尺寸为0')
    }
    return
  }
  
  try {
    lf = new LogicFlow({
      container: canvasContainer.value,
      width: rect.width,
      height: rect.height,
      grid: {
        size: 20,
        visible: true,
        type: 'dot'
      },
      keyboard: {
        enabled: true
      },
      snapline: true,
      history: true,
      partial: true,
      style: {
        rect: {
          rx: 5,
          ry: 5,
          strokeWidth: 2
        },
        circle: {
          strokeWidth: 2
        },
        ellipse: {
          strokeWidth: 2
        },
        polygon: {
          strokeWidth: 2
        },
        polyline: {
          strokeWidth: 2
        },
        text: {
          color: '#000000',
          fontSize: 12
        }
      }
    })

    // 注册自定义节点
    registerCustomNodes()

    // 绑定事件
    bindEvents()

    // 清空画布
    lf.clearData()

    // 如果有workflowId，加载现有工作流
    if (props.workflowId) {
      loadWorkflow()
    }
    
    console.log('LogicFlow initialized successfully', { width: rect.width, height: rect.height })
    
  } catch (error) {
    console.error('LogicFlow initialization failed:', error)
    ElMessage.error('画布初始化失败')
    lf = null
  }
}

const registerCustomNodes = () => {
  // 注册开始节点
  lf.register('start', ({ RectNode, RectNodeModel }) => {
    class StartNodeModel extends RectNodeModel {
      initNodeData(data) {
        super.initNodeData(data)
        this.width = 80
        this.height = 40
        this.radius = 20
        this.text.value = '开始'
      }

      getNodeStyle() {
        const style = super.getNodeStyle()
        style.fill = '#67c23a'
        style.stroke = '#67c23a'
        style.color = '#ffffff'
        return style
      }
    }

    return {
      view: RectNode,
      model: StartNodeModel
    }
  })

  // 注册结束节点
  lf.register('end', ({ RectNode, RectNodeModel }) => {
    class EndNodeModel extends RectNodeModel {
      initNodeData(data) {
        super.initNodeData(data)
        this.width = 80
        this.height = 40
        this.radius = 20
        this.text.value = '结束'
      }

      getNodeStyle() {
        const style = super.getNodeStyle()
        style.fill = '#909399'
        style.stroke = '#909399'
        style.color = '#ffffff'
        return style
      }
    }

    return {
      view: RectNode,
      model: EndNodeModel
    }
  })

  // 注册条件节点
  lf.register('condition', ({ PolygonNode, PolygonNodeModel }) => {
    class ConditionNodeModel extends PolygonNodeModel {
      initNodeData(data) {
        super.initNodeData(data)
        this.width = 80
        this.height = 80
        this.text.value = '条件判断'
        this.points = [
          [40, 0],
          [80, 40],
          [40, 80],
          [0, 40]
        ]
      }

      getNodeStyle() {
        const style = super.getNodeStyle()
        style.fill = '#e6a23c'
        style.stroke = '#e6a23c'
        style.color = '#ffffff'
        return style
      }
    }

    return {
      view: PolygonNode,
      model: ConditionNodeModel
    }
  })

  // 注册Agent节点
  const agentTypes = ['requirement-analysis', 'nesma-evaluation', 'knowledge-retrieval', 'document-generation']
  const agentColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c']
  const agentNames = ['需求分析', 'NESMA评估', '知识检索', '文档生成']

  agentTypes.forEach((type, index) => {
    lf.register(type, ({ RectNode, RectNodeModel }) => {
      class AgentNodeModel extends RectNodeModel {
        initNodeData(data) {
          super.initNodeData(data)
          this.width = 100
          this.height = 60
          this.radius = 10
          this.text.value = agentNames[index]
        }

        getNodeStyle() {
          const style = super.getNodeStyle()
          style.fill = agentColors[index]
          style.stroke = agentColors[index]
          style.color = '#ffffff'
          return style
        }
      }

      return {
        view: RectNode,
        model: AgentNodeModel
      }
    })
  })
}

const bindEvents = () => {
  // 选择事件
  lf.on('selection:selected', (data) => {
    if (data.isMultiple) return
    const element = data.data
    if (element.type === 'node') {
      selectedElement.value = { type: 'node', data: element }
      nodeProperties.value = {
        name: element.text?.value || '',
        description: element.properties?.description || '',
        agentType: element.properties?.agentType || '',
        timeout: element.properties?.timeout || 60,
        retries: element.properties?.retries || 3
      }
    } else if (element.type === 'edge') {
      selectedElement.value = { type: 'edge', data: element }
      edgeProperties.value = {
        name: element.text?.value || '',
        condition: element.properties?.condition || ''
      }
    }
  })

  // 取消选择事件
  lf.on('blank:click', () => {
    selectedElement.value = null
  })
}

const toggleCategory = (categoryName) => {
  expandedCategories.value[categoryName] = !expandedCategories.value[categoryName]
}

const handleNodeDragStart = (event, node) => {
  event.dataTransfer.setData('application/json', JSON.stringify(node))
}

const handleDragOver = (event) => {
  event.preventDefault()
}

const handleDrop = (event) => {
  event.preventDefault()
  try {
    // 检查 LogicFlow 是否已初始化
    if (!lf) {
      ElMessage.warning('画布未初始化，正在重新初始化...')
      initLogicFlow()
      return
    }
    
    const nodeData = JSON.parse(event.dataTransfer.getData('application/json'))
    const { offsetX, offsetY } = event
    
    lf.addNode({
      type: nodeData.type,
      x: offsetX,
      y: offsetY,
      text: nodeData.name,
      properties: {
        agentType: nodeData.type,
        timeout: 60,
        retries: 3
      }
    })
  } catch (error) {
    console.error('Drop node failed:', error)
    ElMessage.error('添加节点失败，请重试')
  }
}

const zoomIn = () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  lf.zoom(true)
}

const zoomOut = () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  lf.zoom(false)
}

const resetZoom = () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  lf.resetZoom()
}

const autoLayout = () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  ElMessage.info('自动布局功能开发中')
}

const clearCanvas = async () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  
  try {
    await ElMessageBox.confirm('确定要清空画布吗？', '确认', {
      type: 'warning'
    })
    lf.clearData()
    selectedElement.value = null
  } catch (error) {
    // 用户取消
  }
}

const updateNodeProperties = () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  
  if (!selectedElement.value || selectedElement.value.type !== 'node') return
  
  const nodeId = selectedElement.value.data.id
  lf.setProperties(nodeId, {
    description: nodeProperties.value.description,
    agentType: nodeProperties.value.agentType,
    timeout: nodeProperties.value.timeout,
    retries: nodeProperties.value.retries
  })
  
  if (nodeProperties.value.name !== selectedElement.value.data.text?.value) {
    lf.updateText(nodeId, nodeProperties.value.name)
  }
}

const updateEdgeProperties = () => {
  if (!lf) {
    ElMessage.warning('画布未初始化')
    return
  }
  
  if (!selectedElement.value || selectedElement.value.type !== 'edge') return
  
  const edgeId = selectedElement.value.data.id
  lf.setProperties(edgeId, {
    condition: edgeProperties.value.condition
  })
  
  if (edgeProperties.value.name !== selectedElement.value.data.text?.value) {
    lf.updateText(edgeId, edgeProperties.value.name)
  }
}

const previewWorkflow = () => {
  previewDialogVisible.value = true
}

const saveWorkflow = async () => {
  if (!workflowName.value.trim()) {
    ElMessage.error('请输入工作流名称')
    return
  }

  saving.value = true
  try {
    if (!lf) {
      ElMessage.error('LogicFlow 引擎未初始化')
      return
    }
    
    const graphData = lf.getGraphData()
    
    // 构建保存数据
    const saveData = {
      name: workflowName.value,
      category: workflowCategory.value,
      definition: graphData,
      version: '1.0.0'
    }

    // 这里应该调用API保存工作流
    console.log('保存工作流:', saveData)
    
    ElMessage.success('工作流保存成功')
    emit('refresh')
  } catch (error) {
    console.error('保存失败:', error)
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

const executeWorkflow = () => {
  ElMessage.info('测试执行功能开发中')
}

const loadWorkflow = async () => {
  // 这里应该根据workflowId加载工作流数据
  console.log('加载工作流:', props.workflowId)
}

const getNodeCount = () => {
  if (!lf) return 0
  return lf.getGraphData().nodes.length
}

const getEdgeCount = () => {
  if (!lf) return 0
  return lf.getGraphData().edges.length
}

const getCategoryLabel = (category) => {
  const labelMap = {
    requirement_analysis: '需求分析',
    nesma_evaluation: 'NESMA评估',
    document_generation: '文档生成',
    quality_check: '质量检查'
  }
  return labelMap[category] || category
}

const handleBeforeClose = (done) => {
  ElMessageBox.confirm('确定要关闭设计器吗？未保存的更改将丢失。')
    .then(() => {
      done()
    })
    .catch(() => {
      // 用户取消
    })
}

// 监听器
watch(() => props.visible, (visible) => {
  if (visible) {
    // 等待drawer动画完成后再初始化
    setTimeout(() => {
      nextTick(() => {
        // 多重检查确保容器已准备好
        const checkAndInit = () => {
          if (canvasContainer.value) {
            const rect = canvasContainer.value.getBoundingClientRect()
            if (rect.width > 0 && rect.height > 0) {
              initLogicFlow()
            } else {
              // 如果容器还没有尺寸，再等待一下
              setTimeout(checkAndInit, 100)
            }
          } else {
            setTimeout(checkAndInit, 100)
          }
        }
        checkAndInit()
      })
    }, 350)
  } else {
    // drawer 关闭时清理 LogicFlow 实例
    if (lf) {
      lf.destroy()
      lf = null
    }
  }
})

// 窗口大小变化监听
const handleResize = () => {
  if (lf && canvasContainer.value) {
    const rect = canvasContainer.value.getBoundingClientRect()
    lf.resize(rect.width, rect.height)
  }
}

onMounted(() => {
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  window.removeEventListener('resize', handleResize)
  if (lf) {
    lf.destroy()
    lf = null
  }
})
</script>

<style scoped>
.designer-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.designer-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: white;
  border-bottom: 1px solid #e4e7ed;
}

.toolbar-left {
  display: flex;
  align-items: center;
}

.toolbar-right {
  display: flex;
  align-items: center;
  gap: 8px;
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
  padding: 16px 20px;
  border-bottom: 1px solid #e4e7ed;
  background: #f8fafc;
}

.panel-header h4 {
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
}

.canvas-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: #f8fafc;
  border-bottom: 1px solid #e4e7ed;
}

.canvas-container {
  flex: 1;
  background: white;
  border-radius: 8px;
  margin: 8px;
  border: 1px solid #e4e7ed;
  position: relative;
}

.canvas-placeholder {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.9);
  z-index: 10;
}

.property-panel {
  width: 300px;
  background: white;
  border-left: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.property-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

.no-selection {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px 20px;
}

.preview-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.workflow-preview {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
}

.workflow-preview h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #606266;
}

.mini-canvas {
  height: 200px;
  background: #f8fafc;
  border: 1px solid #e4e7ed;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
}

.workflow-definition {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 16px;
}

.workflow-definition h4 {
  margin: 0 0 12px 0;
  font-size: 14px;
  color: #606266;
}

.json-viewer {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
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
</style> 