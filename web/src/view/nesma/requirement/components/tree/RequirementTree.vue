<template>
  <div class="requirement-tree">
    <!-- 树形视图头部 -->
    <div class="tree-header">
      <!-- 搜索状态提示 -->
      <div v-if="hasSearchConditions" class="search-status">
        <el-alert
          :title="`当前搜索条件下共找到 ${filteredCount} 条记录`"
          type="info"
          :closable="false"
          show-icon
        >
          <template #default>
            <span>当前搜索条件：</span>
            <el-tag v-if="searchConditions.level" size="small" type="info" style="margin-left: 5px;">
              层级：{{ getLevelLabel(searchConditions.level) }}
            </el-tag>
            <el-tag v-if="searchConditions.status" size="small" type="info" style="margin-left: 5px;">
              状态：{{ getStatusLabel(searchConditions.status) }}
            </el-tag>
            <el-tag v-if="searchConditions.keyword" size="small" type="info" style="margin-left: 5px;">
              关键词：{{ searchConditions.keyword }}
            </el-tag>
            <el-tag v-if="searchConditions.functionType" size="small" type="info" style="margin-left: 5px;">
              功能类型：{{ searchConditions.functionType }}
            </el-tag>
          </template>
        </el-alert>
      </div>
      
      <!-- 树形控制按钮 -->
      <div class="tree-controls">
        <el-button size="small" @click="expandAll">
          <el-icon><FolderOpened /></el-icon>
          全部展开
        </el-button>
        <el-button size="small" @click="collapseAll">
          <el-icon><Folder /></el-icon>
          全部收起
        </el-button>
        
        <!-- 按级别展开 -->
        <el-dropdown @command="expandToLevel" trigger="click">
          <el-button size="small">
            展开到
            <el-icon class="el-icon--right"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item :command="1">L1 - 一级模块</el-dropdown-item>
              <el-dropdown-item :command="2">L2 - 二级模块</el-dropdown-item>
              <el-dropdown-item :command="3">L3 - 三级模块</el-dropdown-item>
              <el-dropdown-item :command="4">L4 - 功能点</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
        
        <el-button size="small" @click="$emit('refresh')">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>

    <!-- 空数据状态 -->
    <el-empty v-if="!treeData.length && !loading" description="暂无需求数据">
      <template #description>
        <div class="empty-description">
          <p>当前项目/周期下暂无需求数据</p>
          <p class="empty-hint">您可以：</p>
          <el-button type="primary" size="small" @click="$emit('create')" style="margin: 8px;">
            <el-icon><Plus /></el-icon>
            新建需求
          </el-button>
          <el-button type="info" size="small" @click="$emit('import')" style="margin: 8px;">
            <el-icon><Upload /></el-icon>
            导入Excel
          </el-button>
        </div>
      </template>
    </el-empty>

    <!-- 加载状态 -->
    <el-skeleton v-else-if="loading" :rows="8" animated />

    <!-- 树形数据 -->
    <el-tree
      v-else-if="treeData.length"
      ref="treeRef"
      :data="treeData"
      :props="treeProps"
      node-key="id"
      :default-expand-all="false"
      :expand-on-click-node="false"
      :filter-node-method="filterTreeNode"
      :allow-drop="allowDropFn"
      :allow-drag="allowDragFn"
      draggable
      @node-click="handleNodeClick"
      @node-drop="handleNodeDrop"
      @node-drag-start="handleDragStart"
      @node-drag-end="handleDragEnd"
    >
      <template #default="{ data }">
        <div class="tree-node" :class="{ 'is-dragging': data.id === draggingNodeId }">
          <div class="node-content">
            <!-- 节点图标 -->
            <el-icon class="node-icon" :class="`level-${data.level}`">
              <component :is="getNodeIcon(data.level)" />
            </el-icon>
            
            <!-- 节点标题 -->
            <span class="node-title" :class="{ 'has-ai-optimization': data.aiAnalysisStatus === 'completed' }">
              {{ getDisplayTitle(data) }}
            </span>
            
            <!-- AI分析状态指示器 -->
            <el-tooltip v-if="data.aiAnalysisStatus === 'completed'" content="已AI优化" placement="top">
              <el-icon class="ai-indicator">
                <MagicStick />
              </el-icon>
            </el-tooltip>
            <el-tooltip v-else-if="data.aiAnalysisStatus === 'analyzing'" content="AI分析中" placement="top">
              <el-icon class="ai-indicator analyzing">
                <Loading />
              </el-icon>
            </el-tooltip>
            
            <!-- 层级标签 -->
            <el-tag size="small" :type="getLevelType(data.level)" class="level-tag">
              {{ getLevelLabel(data.level) }}
            </el-tag>
            
            <!-- 状态标签 -->
            <el-tag size="small" :type="getStatusType(data.status)" class="status-tag">
              {{ getStatusLabel(data.status) }}
            </el-tag>
            
            <!-- 功能类型标签 -->
            <el-tag v-if="data.functionType" size="small" type="info" class="function-type-tag">
              {{ data.functionType }}
            </el-tag>
            
            <!-- 复杂度标签 -->
            <el-tag v-if="data.complexity" size="small" :type="getComplexityType(data.complexity)" class="complexity-tag">
              {{ data.complexity }}
            </el-tag>
            
            <!-- 功能点数 -->
            <span v-if="data.afp > 0 || data.ufp > 0" class="function-points">
              <el-tooltip :content="`AFP: ${data.afp}, UFP: ${data.ufp}`" placement="top">
                <el-tag size="small" type="warning">
                  FP: {{ data.afp || data.ufp || 0 }}
                </el-tag>
              </el-tooltip>
            </span>
          </div>
          
          <!-- 节点操作按钮 -->
          <div class="node-actions" v-show="!isDragging">
            <el-tooltip content="查看详情" placement="top">
              <el-button size="small" circle @click.stop="handleView(data)">
                <el-icon><View /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip content="编辑" placement="top">
              <el-button size="small" type="primary" circle @click.stop="handleEdit(data)">
                <el-icon><Edit /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip v-if="data.level < 4" content="添加子需求" placement="top">
              <el-button size="small" type="success" circle @click.stop="handleCreateChild(data)">
                <el-icon><Plus /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip v-if="data.level >= 3 && !data.aiAnalysisStatus" content="AI分析" placement="top">
              <el-button size="small" type="warning" circle @click.stop="handleAIAnalysis(data)">
                <el-icon><MagicStick /></el-icon>
              </el-button>
            </el-tooltip>
            
            <!-- 层级特定操作按钮 -->
            <el-tooltip v-if="data.level === 2" content="L3分析" placement="top">
              <el-button size="small" type="warning" circle @click.stop="handleL3Analysis(data)">
                <el-icon><Search /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip v-if="data.level === 3" content="生成L4" placement="top">
              <el-button size="small" type="success" circle @click.stop="handleL4Generation(data)">
                <el-icon><Operation /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-tooltip v-if="data.level === 4" content="流程图" placement="top">
              <el-button size="small" type="info" circle @click.stop="handleMermaidGeneration(data)">
                <el-icon><Coordinate /></el-icon>
              </el-button>
            </el-tooltip>
            
            <el-dropdown @command="(command) => handleMoreActions(command, data)" trigger="click">
              <el-button size="small" circle>
                <el-icon><More /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="copy" icon="DocumentCopy">复制</el-dropdown-item>
                  <el-dropdown-item command="move" icon="Rank">移动</el-dropdown-item>
                  <el-dropdown-item command="export" icon="Download">导出</el-dropdown-item>
                  <el-dropdown-item divided />
                  <el-dropdown-item command="delete" icon="Delete">删除</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </template>
    </el-tree>

    <!-- 节点详情弹窗 -->
    <el-drawer
      v-model="showNodeDetails"
      :title="selectedNode?.title || '需求详情'"
      direction="rtl"
      size="40%"
    >
      <div v-if="selectedNode" class="node-details">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="需求ID">{{ selectedNode.id }}</el-descriptions-item>
          <el-descriptions-item label="需求编号">{{ selectedNode.code || '无' }}</el-descriptions-item>
          <el-descriptions-item label="需求标题">{{ selectedNode.title }}</el-descriptions-item>
          <el-descriptions-item label="层级">
            <el-tag :type="getLevelType(selectedNode.level)">
              {{ getLevelLabel(selectedNode.level) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getStatusType(selectedNode.status)">
              {{ getStatusLabel(selectedNode.status) }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.functionType" label="功能类型">
            <el-tag type="info">{{ selectedNode.functionType }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.complexity" label="复杂度">
            <el-tag :type="getComplexityType(selectedNode.complexity)">
              {{ selectedNode.complexity }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="selectedNode.afp || selectedNode.ufp" label="功能点数">
            AFP: {{ selectedNode.afp || 0 }}, UFP: {{ selectedNode.ufp || 0 }}
          </el-descriptions-item>
          <el-descriptions-item label="创建时间">
            {{ formatDateTime(selectedNode.createdAt) }}
          </el-descriptions-item>
          <el-descriptions-item label="更新时间">
            {{ formatDateTime(selectedNode.updatedAt) }}
          </el-descriptions-item>
        </el-descriptions>
        
        <div v-if="selectedNode.description" class="node-description">
          <h4>需求描述</h4>
          <p>{{ selectedNode.description }}</p>
        </div>
        
        <div v-if="selectedNode.aiDescription" class="ai-description">
          <h4>AI优化描述</h4>
          <p>{{ selectedNode.aiDescription }}</p>
        </div>
        
        <div v-if="selectedNode.notes" class="node-notes">
          <h4>备注</h4>
          <p>{{ selectedNode.notes }}</p>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import {
  FolderOpened, Folder, Refresh, Plus, Upload, View, Edit, More,
  MagicStick, Loading, DocumentCopy, Rank, Download, Delete,
  Document, Files, List, Operation, Search, Coordinate, ArrowDown
} from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox } from 'element-plus'

// Props
const props = defineProps({
  treeData: {
    type: Array,
    default: () => []
  },
  searchConditions: {
    type: Object,
    default: () => ({})
  },
  filteredCount: {
    type: Number,
    default: 0
  },
  loading: {
    type: Boolean,
    default: false
  },
  allowDragFn: {
    type: Function,
    default: () => (node) => true
  },
  allowDropFn: {
    type: Function,
    default: () => (draggingNode, dropNode, type) => true
  },
  defaultExpandLevel: {
    type: Number,
    default: 3 // 默认展开到第3级
  }
})

// Emits
const emit = defineEmits([
  'create',
  'import',
  'refresh',
  'view',
  'edit',
  'create-child',
  'delete',
  'ai-analysis',
  'l3-analysis',
  'l4-generation', 
  'mermaid-generation',
  'node-click',
  'node-drop',
  'copy',
  'move',
  'export',
  'tree-rendered' // 新增：树形渲染完成事件
])

// 响应式数据
const treeRef = ref(null)
const showNodeDetails = ref(false)
const selectedNode = ref(null)
const isDragging = ref(false)
const draggingNodeId = ref(null)

// 树形组件配置
const treeProps = {
  children: 'children',
  label: 'title',
  disabled: (data) => data.status === 'cancelled'
}

// 计算属性
const hasSearchConditions = computed(() => {
  const conditions = props.searchConditions
  return !!(conditions.keyword || conditions.level || conditions.status || 
           conditions.functionType || conditions.complexity?.length)
})

// 方法
const expandAll = () => {
  if (treeRef.value) {
    const nodes = treeRef.value.store._getAllNodes()
    nodes.forEach(node => {
      node.expanded = true
    })
  }
}

const collapseAll = () => {
  if (treeRef.value) {
    const nodes = treeRef.value.store._getAllNodes()
    nodes.forEach(node => {
      node.expanded = false
    })
  }
}

// 展开到指定级别
const expandToLevel = (level) => {
  if (treeRef.value) {
    const nodes = treeRef.value.store._getAllNodes()
    nodes.forEach(node => {
      // 严格控制展开级别：小于指定级别的节点展开，大于等于指定级别的节点收起
      if (node.level <= level) {
        node.expanded = true
      } else {
        node.expanded = false
      }
    })
    console.log(`树形视图展开到 L${level} 级别，第${level + 1}级及以上节点已收起`)
  }
}

// 自动展开到默认级别（简化版本，主要逻辑已移到checkTreeRenderComplete）
const autoExpandToDefaultLevel = () => {
  console.log('自动展开方法被调用（兼容性保留）')
  // 这个方法保留作为向后兼容，主要逻辑已经移到checkTreeRenderComplete中
}

const filterTreeNode = (value, data) => {
  if (!value) return true
  const conditions = props.searchConditions
  
  // 关键词匹配
  if (conditions.keyword) {
    const keyword = conditions.keyword.toLowerCase()
    if (!data.title.toLowerCase().includes(keyword) && 
        !data.description?.toLowerCase().includes(keyword)) {
      return false
    }
  }
  
  // 层级匹配
  if (conditions.level && data.level !== conditions.level) {
    return false
  }
  
  // 状态匹配
  if (conditions.status && data.status !== conditions.status) {
    return false
  }
  
  // 功能类型匹配
  if (conditions.functionType && data.functionType !== conditions.functionType) {
    return false
  }
  
  return true
}

const getDisplayTitle = (data) => {
  // 优先显示AI生成的标题，否则显示原标题
  return data.aiGeneratedTitle || data.title
}

const getNodeIcon = (level) => {
  const iconMap = {
    1: Document,
    2: Files,
    3: List,
    4: Operation
  }
  return iconMap[level] || Document
}

const getLevelType = (level) => {
  const typeMap = {
    1: 'danger',
    2: 'warning', 
    3: 'primary',
    4: 'success'
  }
  return typeMap[level] || 'info'
}

const getLevelLabel = (level) => {
  const labelMap = {
    1: 'L1',
    2: 'L2', 
    3: 'L3',
    4: 'L4'
  }
  return labelMap[level] || `L${level}`
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'in_progress': 'warning',
    'completed': 'success',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    'pending': '待处理',
    'in_progress': '进行中',
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return labelMap[status] || status
}

const getComplexityType = (complexity) => {
  const typeMap = {
    '简单': 'success',
    '中等': 'warning',
    '复杂': 'danger'
  }
  return typeMap[complexity] || 'info'
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '无'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// 事件处理
const handleNodeClick = (data) => {
  selectedNode.value = data
  emit('node-click', data)
}

const handleView = (data) => {
  // 不在树形组件内打开详情抽屉，直接发送事件给父组件
  emit('view', data)
}

const handleEdit = (data) => {
  emit('edit', data)
}

const handleCreateChild = (data) => {
  emit('create-child', data)
}

const handleAIAnalysis = (data) => {
  emit('ai-analysis', data)
}

const handleL3Analysis = (data) => {
  emit('l3-analysis', data)
}

const handleL4Generation = (data) => {
  emit('l4-generation', data)
}

const handleMermaidGeneration = (data) => {
  emit('mermaid-generation', data)
}

const handleMoreActions = (command, data) => {
  switch (command) {
    case 'copy':
      emit('copy', data)
      break
    case 'move':
      emit('move', data)
      break
    case 'export':
      emit('export', data)
      break
    case 'delete':
      ElMessageBox.confirm(
        `确定要删除需求"${data.title}"吗？此操作不可撤销。`,
        '确认删除',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      ).then(() => {
        emit('delete', data)
      }).catch(() => {
        // 用户取消删除
      })
      break
  }
}

// 拖拽相关
const handleDragStart = (node, ev) => {
  isDragging.value = true
  draggingNodeId.value = node.data.id
}

const handleDragEnd = (draggingNode, dropNode, dropType, ev) => {
  isDragging.value = false
  draggingNodeId.value = null
}

const handleNodeDrop = (draggingNode, dropNode, dropType) => {
  const dragData = draggingNode.data
  const dropData = dropNode.data
  
  emit('node-drop', {
    dragNode: dragData,
    dropNode: dropData,
    dropType
  })
}

// 监听搜索条件变化，自动过滤树节点
watch(() => props.searchConditions, (newConditions) => {
  if (treeRef.value) {
    treeRef.value.filter(newConditions)
  }
}, { deep: true })

// 监听树形数据变化，启动渲染完成检测
watch(() => props.treeData, (newData, oldData) => {
  console.log('树形数据变化监听器触发', {
    新数据长度: newData?.length || 0,
    旧数据长度: oldData?.length || 0
  })
  
  if (newData && newData.length > 0) {
    console.log('检测到树形数据，开始渲染完成检测流程...')
    
    // 短暂延迟后启动检测，确保组件已经开始渲染
    setTimeout(() => {
      console.log('启动树形渲染完成检测')
      checkTreeRenderComplete()
    }, 150)
  }
}, { immediate: true, deep: false })

// 检测树形渲染是否真正完成
const checkTreeRenderComplete = () => {
  console.log('=== 开始树形渲染完成检测 ===')
  
  let checkCount = 0
  const maxChecks = 80 // 最多检测8秒 (80 * 100ms)
  
  const checkRender = () => {
    checkCount++
    console.log(`第${checkCount}次检测树形渲染状态...`)
    
    // 超过最大检测次数
    if (checkCount > maxChecks) {
      console.warn('检测超时，强制完成')
      emit('tree-rendered')
      return
    }
    
    if (!treeRef.value) {
      console.log('树形引用不存在，继续等待...')
      setTimeout(checkRender, 100)
      return
    }

    // 检测Element树组件内部状态
    const store = treeRef.value.store
    if (!store) {
      console.log('树形store不存在，继续等待...')
      setTimeout(checkRender, 100)
      return
    }

    const allNodes = store._getAllNodes()
    if (!allNodes || allNodes.length === 0) {
      console.log('树形节点尚未生成，继续等待...')
      setTimeout(checkRender, 100)
      return
    }

    // 检测实际的DOM元素
    const treeElement = treeRef.value.$el
    if (!treeElement) {
      console.log('树形DOM元素不存在，继续等待...')
      setTimeout(checkRender, 100)
      return
    }
    
    const nodeElements = treeElement.querySelectorAll('.el-tree-node')
    
    console.log(`检测结果 - 内部节点:${allNodes.length}, DOM节点:${nodeElements.length}, 期望:${props.treeData.length}`)
    
    if (nodeElements.length === 0) {
      console.log('DOM节点尚未渲染，继续等待...')
      setTimeout(checkRender, 100)
      return
    }

    // 检查节点数量是否足够（考虑树形结构可能不是1:1的映射）
    if (nodeElements.length < Math.min(props.treeData.length, 50)) {
      console.log(`节点数量不足，继续等待... (当前${nodeElements.length}个)`)
      setTimeout(checkRender, 100)
      return
    }

    // 检测是否有loading状态的节点
    const loadingNodes = treeElement.querySelectorAll('.el-tree-node.is-loading')
    if (loadingNodes.length > 0) {
      console.log(`存在${loadingNodes.length}个loading节点，继续等待...`)
      setTimeout(checkRender, 100)
      return
    }

    console.log('✓ 树形结构DOM渲染完成，执行展开逻辑')
    
    // 执行展开逻辑
    if (props.defaultExpandLevel > 0) {
      try {
        expandToLevel(props.defaultExpandLevel)
        console.log(`✓ 已展开到L${props.defaultExpandLevel}级别`)
      } catch (error) {
        console.warn('展开失败:', error)
      }
    }

    // 等待展开动画完成后发送完成事件
    setTimeout(() => {
      console.log('=== 树形渲染和展开全部完成，发送完成事件 ===')
      emit('tree-rendered')
    }, 400) // 等待展开动画完成
  }

  // 立即开始第一次检测
  checkRender()
}

// 暴露方法给父组件
defineExpose({
  expandAll,
  collapseAll,
  expandToLevel,
  autoExpandToDefaultLevel,
  filter: (value) => {
    if (treeRef.value) {
      treeRef.value.filter(value)
    }
  }
})
</script>

<style lang="scss" scoped>
.requirement-tree {
  .tree-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 16px;
    gap: 16px;

    .search-status {
      flex: 1;
    }

    .tree-controls {
      display: flex;
      gap: 8px;
      flex-shrink: 0;
    }
  }

  .empty-description {
    text-align: center;
    color: #909399;

    .empty-hint {
      margin: 8px 0;
      font-size: 12px;
    }
  }

  :deep(.el-tree) {
    .el-tree-node {
      .el-tree-node__content {
        height: auto;
        min-height: 40px;
        padding: 8px 0;
      }
    }
  }

  .tree-node {
    display: flex;
    align-items: center;
    justify-content: space-between;
    width: 100%;
    min-height: 40px;
    padding: 4px 8px;
    border-radius: 4px;
    transition: all 0.3s ease;

    &:hover {
      background-color: #f5f7fa;

      .node-actions {
        opacity: 1;
      }
    }

    &.is-dragging {
      opacity: 0.5;
      background-color: #e1f3d8;
    }

    .node-content {
      display: flex;
      align-items: center;
      gap: 8px;
      flex: 1;
      min-width: 0;

      .node-icon {
        font-size: 16px;
        flex-shrink: 0;

        &.level-1 { color: #f56c6c; }
        &.level-2 { color: #e6a23c; }
        &.level-3 { color: #409eff; }
        &.level-4 { color: #67c23a; }
      }

      .node-title {
        font-size: 14px;
        font-weight: 500;
        color: #303133;
        flex: 1;
        min-width: 0;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;

        &.has-ai-optimization {
          color: #409eff;
          font-weight: 600;
        }
      }

      .ai-indicator {
        font-size: 14px;
        color: #409eff;

        &.analyzing {
          animation: rotate 1s linear infinite;
        }
      }

      .level-tag,
      .status-tag,
      .function-type-tag,
      .complexity-tag {
        flex-shrink: 0;
      }

      .function-points {
        font-size: 12px;
        color: #909399;
        flex-shrink: 0;
      }
    }

    .node-actions {
      display: flex;
      gap: 4px;
      opacity: 0;
      transition: opacity 0.3s ease;
      flex-shrink: 0;

      .el-button {
        width: 28px;
        height: 28px;
      }
    }
  }

  .node-details {
    .node-description,
    .ai-description,
    .node-notes {
      margin-top: 20px;

      h4 {
        margin-bottom: 8px;
        color: #303133;
        font-weight: 600;
      }

      p {
        color: #606266;
        line-height: 1.6;
        margin: 0;
      }
    }

    .ai-description {
      p {
        color: #409eff;
        background: #f0f9ff;
        padding: 12px;
        border-radius: 4px;
        border-left: 3px solid #409eff;
      }
    }
  }
}

@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

// 响应式设计
@media (max-width: 768px) {
  .requirement-tree {
    .tree-header {
      flex-direction: column;
      align-items: stretch;

      .tree-controls {
        justify-content: center;
      }
    }

    .tree-node {
      .node-content {
        gap: 4px;

        .node-title {
          font-size: 13px;
        }
      }

      .node-actions {
        .el-button {
          width: 24px;
          height: 24px;
        }
      }
    }
  }
}
</style>