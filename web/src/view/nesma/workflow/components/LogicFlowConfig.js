import { RectNode, CircleNode, PolygonNode } from '@logicflow/core'

// Agent节点基类
class AgentNode extends RectNode {
  static extendKey = 'AgentNode'
  
  constructor(props) {
    super(props)
  }

  getDefaultAnchor() {
    const { x, y, width, height } = this.properties
    return [
      { x: x - width / 2, y, id: `${this.id}_left` },
      { x: x + width / 2, y, id: `${this.id}_right` },
      { x, y: y - height / 2, id: `${this.id}_top` },
      { x, y: y + height / 2, id: `${this.id}_bottom` }
    ]
  }

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      rx: 8,
      ry: 8,
      stroke: '#409eff',
      strokeWidth: 2,
      fill: '#e6f7ff',
      fillOpacity: 0.8
    }
  }

  getTextStyle() {
    const style = super.getTextStyle()
    return {
      ...style,
      fontSize: 12,
      fill: '#303133',
      fontWeight: 'bold'
    }
  }
}

// 需求分析Agent节点
class RequirementAgentNode extends AgentNode {
  static extendKey = 'RequirementAgentNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      fill: '#e6f7ff',
      stroke: '#1890ff'
    }
  }
}

// NESMA评估Agent节点
class NesmaAgentNode extends AgentNode {
  static extendKey = 'NesmaAgentNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      fill: '#f6ffed',
      stroke: '#52c41a'
    }
  }
}

// 知识检索Agent节点
class KnowledgeAgentNode extends AgentNode {
  static extendKey = 'KnowledgeAgentNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      fill: '#fff2e8',
      stroke: '#fa8c16'
    }
  }
}

// 文档生成Agent节点
class DocumentAgentNode extends AgentNode {
  static extendKey = 'DocumentAgentNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      fill: '#f9f0ff',
      stroke: '#722ed1'
    }
  }
}

// 协调管理Agent节点
class CoordinationAgentNode extends AgentNode {
  static extendKey = 'CoordinationAgentNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      fill: '#fff1f0',
      stroke: '#f5222d'
    }
  }
}

// 开始节点
class StartNode extends CircleNode {
  static extendKey = 'StartNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      r: 25,
      fill: '#52c41a',
      stroke: '#389e0d',
      strokeWidth: 2
    }
  }

  getTextStyle() {
    const style = super.getTextStyle()
    return {
      ...style,
      fontSize: 12,
      fill: 'white',
      fontWeight: 'bold'
    }
  }

  getDefaultAnchor() {
    const { x, y } = this.properties
    return [
      { x: x + 25, y, id: `${this.id}_right` },
      { x, y: y + 25, id: `${this.id}_bottom` }
    ]
  }
}

// 结束节点
class EndNode extends CircleNode {
  static extendKey = 'EndNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      r: 25,
      fill: '#f5222d',
      stroke: '#cf1322',
      strokeWidth: 2
    }
  }

  getTextStyle() {
    const style = super.getTextStyle()
    return {
      ...style,
      fontSize: 12,
      fill: 'white',
      fontWeight: 'bold'
    }
  }

  getDefaultAnchor() {
    const { x, y } = this.properties
    return [
      { x: x - 25, y, id: `${this.id}_left` },
      { x, y: y - 25, id: `${this.id}_top` }
    ]
  }
}

// 条件判断节点
class ConditionNode extends PolygonNode {
  static extendKey = 'ConditionNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    const { x, y } = this.properties
    const size = 40
    
    return {
      ...style,
      points: [
        [x, y - size],     // 上
        [x + size, y],     // 右
        [x, y + size],     // 下
        [x - size, y]      // 左
      ],
      fill: '#fff7e6',
      stroke: '#fa8c16',
      strokeWidth: 2
    }
  }

  getTextStyle() {
    const style = super.getTextStyle()
    return {
      ...style,
      fontSize: 12,
      fill: '#303133',
      fontWeight: 'bold'
    }
  }

  getDefaultAnchor() {
    const { x, y } = this.properties
    const size = 40
    return [
      { x: x - size, y, id: `${this.id}_left` },
      { x: x + size, y, id: `${this.id}_right` },
      { x, y: y - size, id: `${this.id}_top` },
      { x, y: y + size, id: `${this.id}_bottom` }
    ]
  }
}

// 并行处理节点
class ParallelNode extends RectNode {
  static extendKey = 'ParallelNode'

  getShapeStyle() {
    const style = super.getShapeStyle()
    return {
      ...style,
      rx: 5,
      ry: 5,
      fill: '#f0f9ff',
      stroke: '#1890ff',
      strokeWidth: 2,
      strokeDasharray: '5,5'
    }
  }

  getTextStyle() {
    const style = super.getTextStyle()
    return {
      ...style,
      fontSize: 12,
      fill: '#303133',
      fontWeight: 'bold'
    }
  }
}

// 注册自定义节点
export function registerCustomNodes(lf) {
  // 注册Agent节点
  lf.register({
    type: 'requirement-agent',
    view: RequirementAgentNode,
    model: RequirementAgentNode
  })

  lf.register({
    type: 'nesma-agent',
    view: NesmaAgentNode,
    model: NesmaAgentNode
  })

  lf.register({
    type: 'knowledge-agent',
    view: KnowledgeAgentNode,
    model: KnowledgeAgentNode
  })

  lf.register({
    type: 'document-agent',
    view: DocumentAgentNode,
    model: DocumentAgentNode
  })

  lf.register({
    type: 'coordination-agent',
    view: CoordinationAgentNode,
    model: CoordinationAgentNode
  })

  // 注册控制节点
  lf.register({
    type: 'start-node',
    view: StartNode,
    model: StartNode
  })

  lf.register({
    type: 'end-node',
    view: EndNode,
    model: EndNode
  })

  lf.register({
    type: 'condition-node',
    view: ConditionNode,
    model: ConditionNode
  })

  lf.register({
    type: 'parallel-node',
    view: ParallelNode,
    model: ParallelNode
  })

  // 设置默认连线类型
  lf.setDefaultEdgeType('polyline')
}

// 节点图标映射
export const nodeIconMap = {
  'requirement-agent': '📝',
  'nesma-agent': '📊',
  'knowledge-agent': '🔍',
  'document-agent': '📄',
  'coordination-agent': '🔗',
  'start-node': '▶️',
  'end-node': '⏹️',
  'condition-node': '❓',
  'parallel-node': '⚡'
}

// 节点颜色映射
export const nodeColorMap = {
  'requirement-agent': { bg: '#e6f7ff', border: '#1890ff' },
  'nesma-agent': { bg: '#f6ffed', border: '#52c41a' },
  'knowledge-agent': { bg: '#fff2e8', border: '#fa8c16' },
  'document-agent': { bg: '#f9f0ff', border: '#722ed1' },
  'coordination-agent': { bg: '#fff1f0', border: '#f5222d' },
  'start-node': { bg: '#52c41a', border: '#389e0d' },
  'end-node': { bg: '#f5222d', border: '#cf1322' },
  'condition-node': { bg: '#fff7e6', border: '#fa8c16' },
  'parallel-node': { bg: '#f0f9ff', border: '#1890ff' }
}

// 验证工作流合法性
export function validateWorkflow(nodes, edges) {
  const errors = []
  
  // 检查是否有开始节点
  const startNodes = nodes.filter(node => node.type === 'start-node')
  if (startNodes.length === 0) {
    errors.push('工作流必须包含至少一个开始节点')
  } else if (startNodes.length > 1) {
    errors.push('工作流只能包含一个开始节点')
  }
  
  // 检查是否有结束节点
  const endNodes = nodes.filter(node => node.type === 'end-node')
  if (endNodes.length === 0) {
    errors.push('工作流必须包含至少一个结束节点')
  }
  
  // 检查节点是否都有连接
  nodes.forEach(node => {
    if (node.type === 'start-node') {
      const outgoingEdges = edges.filter(edge => edge.sourceNodeId === node.id)
      if (outgoingEdges.length === 0) {
        errors.push(`开始节点"${node.text || node.id}"必须有出边`)
      }
    } else if (node.type === 'end-node') {
      const incomingEdges = edges.filter(edge => edge.targetNodeId === node.id)
      if (incomingEdges.length === 0) {
        errors.push(`结束节点"${node.text || node.id}"必须有入边`)
      }
    } else {
      const incomingEdges = edges.filter(edge => edge.targetNodeId === node.id)
      const outgoingEdges = edges.filter(edge => edge.sourceNodeId === node.id)
      
      if (incomingEdges.length === 0) {
        errors.push(`节点"${node.text || node.id}"必须有入边`)
      }
      if (outgoingEdges.length === 0 && node.type !== 'end-node') {
        errors.push(`节点"${node.text || node.id}"必须有出边`)
      }
    }
  })
  
  // 检查是否有循环依赖
  const visited = new Set()
  const recursionStack = new Set()
  
  function hasCycle(nodeId) {
    if (recursionStack.has(nodeId)) {
      return true
    }
    if (visited.has(nodeId)) {
      return false
    }
    
    visited.add(nodeId)
    recursionStack.add(nodeId)
    
    const outgoingEdges = edges.filter(edge => edge.sourceNodeId === nodeId)
    for (const edge of outgoingEdges) {
      if (hasCycle(edge.targetNodeId)) {
        return true
      }
    }
    
    recursionStack.delete(nodeId)
    return false
  }
  
  for (const node of nodes) {
    if (hasCycle(node.id)) {
      errors.push('工作流中存在循环依赖，请检查连线关系')
      break
    }
  }
  
  return {
    isValid: errors.length === 0,
    errors
  }
}

/**
 * LogicFlow 自定义节点配置
 */

// 节点类型定义
export const NODE_TYPES = {
  // 控制节点
  START: 'start',
  END: 'end',
  CONDITION: 'condition',
  PARALLEL: 'parallel',
  
  // AI Agent节点
  REQUIREMENT_ANALYSIS: 'requirement_analysis',
  NESMA_EVALUATION: 'nesma_evaluation',
  KNOWLEDGE_RETRIEVAL: 'knowledge_retrieval',
  DOCUMENT_GENERATION: 'document_generation',
  COORDINATOR: 'coordinator'
}

// 节点样式配置
export const NODE_STYLES = {
  [NODE_TYPES.START]: {
    fill: '#67C23A',
    stroke: '#5CB85C',
    strokeWidth: 2,
    r: 25
  },
  [NODE_TYPES.END]: {
    fill: '#F56C6C',
    stroke: '#E53E3E',
    strokeWidth: 2,
    r: 25
  },
  [NODE_TYPES.CONDITION]: {
    fill: '#E6A23C',
    stroke: '#D4AC0D',
    strokeWidth: 2,
    width: 80,
    height: 50
  },
  [NODE_TYPES.PARALLEL]: {
    fill: '#909399',
    stroke: '#6C757D',
    strokeWidth: 2,
    width: 80,
    height: 50
  },
  [NODE_TYPES.REQUIREMENT_ANALYSIS]: {
    fill: '#409EFF',
    stroke: '#2E86AB',
    strokeWidth: 2,
    width: 120,
    height: 60
  },
  [NODE_TYPES.NESMA_EVALUATION]: {
    fill: '#722ED1',
    stroke: '#531DAB',
    strokeWidth: 2,
    width: 120,
    height: 60
  },
  [NODE_TYPES.KNOWLEDGE_RETRIEVAL]: {
    fill: '#13C2C2',
    stroke: '#08979C',
    strokeWidth: 2,
    width: 120,
    height: 60
  },
  [NODE_TYPES.DOCUMENT_GENERATION]: {
    fill: '#52C41A',
    stroke: '#389E0D',
    strokeWidth: 2,
    width: 120,
    height: 60
  },
  [NODE_TYPES.COORDINATOR]: {
    fill: '#FA541C',
    stroke: '#D4380D',
    strokeWidth: 2,
    width: 120,
    height: 60
  }
}

// 节点文本配置
export const NODE_TEXT = {
  [NODE_TYPES.START]: '开始',
  [NODE_TYPES.END]: '结束',
  [NODE_TYPES.CONDITION]: '条件判断',
  [NODE_TYPES.PARALLEL]: '并行处理',
  [NODE_TYPES.REQUIREMENT_ANALYSIS]: '需求分析',
  [NODE_TYPES.NESMA_EVALUATION]: 'NESMA评估',
  [NODE_TYPES.KNOWLEDGE_RETRIEVAL]: '知识检索',
  [NODE_TYPES.DOCUMENT_GENERATION]: '文档生成',
  [NODE_TYPES.COORDINATOR]: '协调管理'
}

// 节点描述配置
export const NODE_DESCRIPTIONS = {
  [NODE_TYPES.START]: '工作流开始节点',
  [NODE_TYPES.END]: '工作流结束节点',
  [NODE_TYPES.CONDITION]: '根据条件判断执行路径',
  [NODE_TYPES.PARALLEL]: '并行执行多个任务',
  [NODE_TYPES.REQUIREMENT_ANALYSIS]: '分析用户需求，提取关键信息',
  [NODE_TYPES.NESMA_EVALUATION]: '进行NESMA功能点评估',
  [NODE_TYPES.KNOWLEDGE_RETRIEVAL]: '从知识库检索相关信息',
  [NODE_TYPES.DOCUMENT_GENERATION]: '生成文档或报告',
  [NODE_TYPES.COORDINATOR]: '协调其他Agent的工作'
}

// 节点图标配置
export const NODE_ICONS = {
  [NODE_TYPES.START]: '▶️',
  [NODE_TYPES.END]: '⏹️',
  [NODE_TYPES.CONDITION]: '❓',
  [NODE_TYPES.PARALLEL]: '⚡',
  [NODE_TYPES.REQUIREMENT_ANALYSIS]: '📋',
  [NODE_TYPES.NESMA_EVALUATION]: '📊',
  [NODE_TYPES.KNOWLEDGE_RETRIEVAL]: '🔍',
  [NODE_TYPES.DOCUMENT_GENERATION]: '📄',
  [NODE_TYPES.COORDINATOR]: '👥'
}

// 连线样式配置
export const EDGE_STYLES = {
  stroke: '#8E8E93',
  strokeWidth: 2,
  fill: 'none'
}

// 自定义节点类
export class CustomNode {
  static getNodeModel(defaultModel) {
    return {
      ...defaultModel,
      setAttributes() {
        const { type } = this
        const style = NODE_STYLES[type] || NODE_STYLES[NODE_TYPES.START]
        
        this.fill = style.fill
        this.stroke = style.stroke
        this.strokeWidth = style.strokeWidth
        
        if (style.r) {
          this.r = style.r
        }
        if (style.width) {
          this.width = style.width
        }
        if (style.height) {
          this.height = style.height
        }
        
        this.text = {
          value: NODE_TEXT[type] || type,
          x: this.x,
          y: this.y,
          fontSize: 12,
          color: '#FFFFFF',
          textAnchor: 'middle',
          dominantBaseline: 'middle'
        }
      }
    }
  }
  
  static getNodeView(defaultView) {
    return {
      ...defaultView,
      getShape() {
        const { model } = this.props
        const { type } = model
        
        const h = this.h
        const attrs = {
          ...model.getNodeStyle(),
          cursor: 'pointer'
        }
        
        // 根据节点类型返回不同的形状
        switch (type) {
          case NODE_TYPES.START:
          case NODE_TYPES.END:
            return h('circle', attrs)
          case NODE_TYPES.CONDITION:
            return h('polygon', {
              ...attrs,
              points: '40,0 80,25 40,50 0,25'
            })
          case NODE_TYPES.PARALLEL:
            return h('polygon', {
              ...attrs,
              points: '0,10 10,0 70,0 80,10 80,40 70,50 10,50 0,40'
            })
          default:
            return h('rect', {
              ...attrs,
              rx: 8,
              ry: 8
            })
        }
      }
    }
  }
}

// 验证规则
export const VALIDATION_RULES = {
  // 必须有开始节点
  requireStart: (graphData) => {
    const startNodes = graphData.nodes.filter(node => node.type === NODE_TYPES.START)
    if (startNodes.length === 0) {
      return { valid: false, message: '工作流必须包含开始节点' }
    }
    if (startNodes.length > 1) {
      return { valid: false, message: '工作流只能包含一个开始节点' }
    }
    return { valid: true }
  },
  
  // 必须有结束节点
  requireEnd: (graphData) => {
    const endNodes = graphData.nodes.filter(node => node.type === NODE_TYPES.END)
    if (endNodes.length === 0) {
      return { valid: false, message: '工作流必须包含结束节点' }
    }
    return { valid: true }
  },
  
  // 检查节点连接
  checkConnections: (graphData) => {
    const { nodes, edges } = graphData
    
    // 检查每个节点是否有适当的连接
    for (const node of nodes) {
      if (node.type === NODE_TYPES.START) {
        const outEdges = edges.filter(edge => edge.sourceNodeId === node.id)
        if (outEdges.length === 0) {
          return { valid: false, message: '开始节点必须有输出连接' }
        }
      }
      
      if (node.type === NODE_TYPES.END) {
        const inEdges = edges.filter(edge => edge.targetNodeId === node.id)
        if (inEdges.length === 0) {
          return { valid: false, message: '结束节点必须有输入连接' }
        }
      }
      
      // 检查AI Agent节点
      if ([
        NODE_TYPES.REQUIREMENT_ANALYSIS,
        NODE_TYPES.NESMA_EVALUATION,
        NODE_TYPES.KNOWLEDGE_RETRIEVAL,
        NODE_TYPES.DOCUMENT_GENERATION,
        NODE_TYPES.COORDINATOR
      ].includes(node.type)) {
        const inEdges = edges.filter(edge => edge.targetNodeId === node.id)
        const outEdges = edges.filter(edge => edge.sourceNodeId === node.id)
        
        if (inEdges.length === 0 && node.type !== NODE_TYPES.START) {
          return { valid: false, message: `节点 "${NODE_TEXT[node.type]}" 必须有输入连接` }
        }
        
        if (outEdges.length === 0 && node.type !== NODE_TYPES.END) {
          return { valid: false, message: `节点 "${NODE_TEXT[node.type]}" 必须有输出连接` }
        }
      }
    }
    
    return { valid: true }
  },
  
  // 检查循环依赖
  checkCycles: (graphData) => {
    const { nodes, edges } = graphData
    const visited = new Set()
    const inStack = new Set()
    
    function hasCycle(nodeId) {
      if (inStack.has(nodeId)) return true
      if (visited.has(nodeId)) return false
      
      visited.add(nodeId)
      inStack.add(nodeId)
      
      const outEdges = edges.filter(edge => edge.sourceNodeId === nodeId)
      for (const edge of outEdges) {
        if (hasCycle(edge.targetNodeId)) {
          return true
        }
      }
      
      inStack.delete(nodeId)
      return false
    }
    
    for (const node of nodes) {
      if (hasCycle(node.id)) {
        return { valid: false, message: '工作流中存在循环依赖' }
      }
    }
    
    return { valid: true }
  }
}

// 工作流验证函数
export function validateWorkflow(graphData) {
  const rules = Object.values(VALIDATION_RULES)
  
  for (const rule of rules) {
    const result = rule(graphData)
    if (!result.valid) {
      return result
    }
  }
  
  return { valid: true, message: '工作流验证通过' }
}

// 节点面板配置
export const NODE_PANEL_CONFIG = [
  {
    title: '控制节点',
    nodes: [
      { type: NODE_TYPES.START, text: NODE_TEXT[NODE_TYPES.START], icon: NODE_ICONS[NODE_TYPES.START] },
      { type: NODE_TYPES.END, text: NODE_TEXT[NODE_TYPES.END], icon: NODE_ICONS[NODE_TYPES.END] },
      { type: NODE_TYPES.CONDITION, text: NODE_TEXT[NODE_TYPES.CONDITION], icon: NODE_ICONS[NODE_TYPES.CONDITION] },
      { type: NODE_TYPES.PARALLEL, text: NODE_TEXT[NODE_TYPES.PARALLEL], icon: NODE_ICONS[NODE_TYPES.PARALLEL] }
    ]
  },
  {
    title: 'AI Agent节点',
    nodes: [
      { type: NODE_TYPES.REQUIREMENT_ANALYSIS, text: NODE_TEXT[NODE_TYPES.REQUIREMENT_ANALYSIS], icon: NODE_ICONS[NODE_TYPES.REQUIREMENT_ANALYSIS] },
      { type: NODE_TYPES.NESMA_EVALUATION, text: NODE_TEXT[NODE_TYPES.NESMA_EVALUATION], icon: NODE_ICONS[NODE_TYPES.NESMA_EVALUATION] },
      { type: NODE_TYPES.KNOWLEDGE_RETRIEVAL, text: NODE_TEXT[NODE_TYPES.KNOWLEDGE_RETRIEVAL], icon: NODE_ICONS[NODE_TYPES.KNOWLEDGE_RETRIEVAL] },
      { type: NODE_TYPES.DOCUMENT_GENERATION, text: NODE_TEXT[NODE_TYPES.DOCUMENT_GENERATION], icon: NODE_ICONS[NODE_TYPES.DOCUMENT_GENERATION] },
      { type: NODE_TYPES.COORDINATOR, text: NODE_TEXT[NODE_TYPES.COORDINATOR], icon: NODE_ICONS[NODE_TYPES.COORDINATOR] }
    ]
  }
]

// 导出默认配置
export default {
  NODE_TYPES,
  NODE_STYLES,
  NODE_TEXT,
  NODE_DESCRIPTIONS,
  NODE_ICONS,
  EDGE_STYLES,
  CustomNode,
  VALIDATION_RULES,
  validateWorkflow,
  NODE_PANEL_CONFIG
} 