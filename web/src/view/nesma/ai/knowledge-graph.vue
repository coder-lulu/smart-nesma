<template>
  <div class="knowledge-graph">
    <!-- 顶部工具栏 -->
    <div class="toolbar">
      <div class="toolbar-left">
        <el-button type="primary" icon="Plus" @click="handleAddEntity">
          添加实体
        </el-button>
        <el-button icon="Connection" @click="handleAddRelation">
          添加关系
        </el-button>
        <el-button icon="Upload" @click="handleImportKnowledge">
          导入知识
        </el-button>
        <el-button icon="Download" @click="handleExportKnowledge">
          导出知识
        </el-button>
      </div>
      <div class="toolbar-right">
        <el-input
          v-model="searchQuery"
          placeholder="搜索知识实体..."
          style="width: 300px; margin-right: 10px;"
          clearable
          @keyup.enter="handleSearch"
        >
          <template #prefix>
            <el-icon><search /></el-icon>
          </template>
        </el-input>
        <el-select
          v-model="selectedEntityType"
          placeholder="实体类型"
          style="width: 120px; margin-right: 10px;"
          clearable
          @change="handleSearch"
        >
          <el-option label="全部" value="" />
          <el-option label="功能" value="Function" />
          <el-option label="领域" value="Domain" />
          <el-option label="模式" value="Pattern" />
          <el-option label="规则" value="Rule" />
          <el-option label="案例" value="Case" />
        </el-select>
        <el-button icon="Search" @click="handleSearch">搜索</el-button>
      </div>
    </div>

    <!-- 主要内容区域 -->
    <el-row :gutter="20">
      <!-- 左侧：知识图谱可视化 -->
      <el-col :span="16">
        <el-card title="知识图谱可视化">
          <template #header>
            <div class="card-header">
              <span>知识图谱可视化</span>
              <div class="graph-controls">
                <el-button-group>
                  <el-button 
                    :type="graphView === 'force' ? 'primary' : ''"
                    size="small"
                    @click="setGraphView('force')"
                  >
                    力导图
                  </el-button>
                  <el-button 
                    :type="graphView === 'tree' ? 'primary' : ''"
                    size="small"
                    @click="setGraphView('tree')"
                  >
                    树形图
                  </el-button>
                  <el-button 
                    :type="graphView === 'circular' ? 'primary' : ''"
                    size="small"
                    @click="setGraphView('circular')"
                  >
                    环形图
                  </el-button>
                </el-button-group>
                <el-divider direction="vertical" />
                <el-button size="small" @click="resetGraphView">
                  <el-icon><refresh /></el-icon>
                  重置视图
                </el-button>
                <el-button size="small" @click="centerGraph">
                  <el-icon><aim /></el-icon>
                  居中
                </el-button>
              </div>
            </div>
          </template>
          
          <div class="graph-container">
            <div ref="graphContainer" class="graph-canvas"></div>
            
            <!-- 图谱控制面板 -->
            <div class="graph-control-panel">
              <div class="control-section">
                <div class="control-title">显示选项</div>
                <el-checkbox v-model="showEntityLabels">显示实体标签</el-checkbox>
                <el-checkbox v-model="showRelationLabels">显示关系标签</el-checkbox>
                <el-checkbox v-model="showEntityTypes">按类型着色</el-checkbox>
              </div>
              
              <div class="control-section">
                <div class="control-title">过滤器</div>
                <el-slider
                  v-model="relationThreshold"
                  :min="0"
                  :max="1"
                  :step="0.1"
                  show-input
                  :show-input-controls="false"
                  input-size="small"
                >
                  <template #default>关系权重</template>
                </el-slider>
                <el-slider
                  v-model="entityConfidence"
                  :min="0"
                  :max="1"
                  :step="0.1"
                  show-input
                  :show-input-controls="false"
                  input-size="small"
                  style="margin-top: 10px;"
                >
                  <template #default>实体置信度</template>
                </el-slider>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
      
      <!-- 右侧：实体列表和详情 -->
      <el-col :span="8">
        <el-card title="知识实体">
          <template #header>
            <div class="card-header">
              <span>知识实体 ({{ filteredEntities.length }})</span>
              <el-button size="small" @click="refreshEntities">
                <el-icon><refresh /></el-icon>
              </el-button>
            </div>
          </template>
          
          <div class="entity-list">
            <div 
              v-for="entity in paginatedEntities" 
              :key="entity.id"
              class="entity-item"
              :class="{ 'selected': selectedEntity?.id === entity.id }"
              @click="selectEntity(entity)"
            >
              <div class="entity-header">
                <div class="entity-info">
                  <div class="entity-name">{{ entity.name }}</div>
                  <el-tag :type="getEntityTypeColor(entity.type)" size="small">
                    {{ entity.type }}
                  </el-tag>
                </div>
                <div class="entity-confidence">
                  <el-progress 
                    :percentage="entity.confidence * 100" 
                    :stroke-width="4"
                    :show-text="false"
                    :color="getConfidenceColor(entity.confidence)"
                  />
                  <span class="confidence-text">{{ (entity.confidence * 100).toFixed(0) }}%</span>
                </div>
              </div>
              
              <div class="entity-description">
                {{ entity.description.substring(0, 80) }}{{ entity.description.length > 80 ? '...' : '' }}
              </div>
              
              <div class="entity-metadata">
                <span class="metadata-item">
                  <el-icon><collection /></el-icon>
                  {{ entity.properties ? Object.keys(entity.properties).length : 0 }} 属性
                </span>
                <span class="metadata-item">
                  <el-icon><connection /></el-icon>
                  {{ getEntityRelationCount(entity.id) }} 关系
                </span>
                <span class="metadata-item">
                  <el-icon><time /></el-icon>
                  {{ formatDate(entity.updated_at) }}
                </span>
              </div>
              
              <div class="entity-actions">
                <el-button size="small" @click.stop="editEntity(entity)">编辑</el-button>
                <el-button size="small" @click.stop="viewEntityDetails(entity)">详情</el-button>
                <el-dropdown @command="handleEntityAction">
                  <el-button size="small">
                    更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
                  </el-button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="{action: 'find_similar', entity}">
                        查找相似
                      </el-dropdown-item>
                      <el-dropdown-item :command="{action: 'analyze_relations', entity}">
                        分析关系
                      </el-dropdown-item>
                      <el-dropdown-item :command="{action: 'export_entity', entity}">
                        导出实体
                      </el-dropdown-item>
                      <el-dropdown-item :command="{action: 'delete_entity', entity}" divided>
                        删除实体
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </div>
          
          <!-- 分页 -->
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="currentPage"
              v-model:page-size="pageSize"
              :page-sizes="[10, 20, 50]"
              :total="filteredEntities.length"
              layout="sizes, prev, pager, next"
              small
            />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 底部：关系列表 -->
    <el-row style="margin-top: 20px">
      <el-col :span="24">
        <el-card title="知识关系">
          <template #header>
            <div class="card-header">
              <span>知识关系 ({{ relations.length }})</span>
              <div class="header-actions">
                <el-button size="small" @click="analyzeRelations">
                  <el-icon><data-analysis /></el-icon>
                  关系分析
                </el-button>
                <el-button size="small" @click="optimizeGraph">
                  <el-icon><magic-stick /></el-icon>
                  图谱优化
                </el-button>
              </div>
            </div>
          </template>
          
          <el-table :data="paginatedRelations" style="width: 100%">
            <el-table-column prop="from_entity_name" label="源实体" width="200" />
            <el-table-column prop="relation_type" label="关系类型" width="120">
              <template #default="{ row }">
                <el-tag :type="getRelationTypeColor(row.relation_type)" size="small">
                  {{ getRelationTypeName(row.relation_type) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="to_entity_name" label="目标实体" width="200" />
            <el-table-column prop="weight" label="权重" width="100">
              <template #default="{ row }">
                <el-progress 
                  :percentage="row.weight * 100" 
                  :stroke-width="6"
                  :show-text="false"
                  :color="getWeightColor(row.weight)"
                />
                <span style="margin-left: 8px; font-size: 12px;">
                  {{ row.weight.toFixed(2) }}
                </span>
              </template>
            </el-table-column>
            <el-table-column prop="confidence" label="置信度" width="100">
              <template #default="{ row }">
                {{ (row.confidence * 100).toFixed(0) }}%
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="150">
              <template #default="{ row }">
                {{ formatDate(row.created_at) }}
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120">
              <template #default="{ row }">
                <el-button size="small" @click="editRelation(row)">编辑</el-button>
                <el-button size="small" type="danger" @click="deleteRelation(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
          
          <div class="pagination-container">
            <el-pagination
              v-model:current-page="relationPage"
              v-model:page-size="relationPageSize"
              :page-sizes="[10, 20, 50]"
              :total="relations.length"
              layout="sizes, prev, pager, next"
              small
            />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 实体详情对话框 -->
    <el-dialog
      v-model="entityDetailVisible"
      :title="selectedEntity?.name + ' 详细信息'"
      width="900px"
      :close-on-click-modal="false"
    >
      <div v-if="selectedEntity" class="entity-details">
        <el-tabs v-model="activeDetailTab">
          <el-tab-pane label="基本信息" name="basic">
            <el-descriptions :column="2" border>
              <el-descriptions-item label="实体ID">{{ selectedEntity.id }}</el-descriptions-item>
              <el-descriptions-item label="名称">{{ selectedEntity.name }}</el-descriptions-item>
              <el-descriptions-item label="类型">{{ selectedEntity.type }}</el-descriptions-item>
              <el-descriptions-item label="置信度">{{ (selectedEntity.confidence * 100).toFixed(1) }}%</el-descriptions-item>
              <el-descriptions-item label="来源">{{ selectedEntity.source }}</el-descriptions-item>
              <el-descriptions-item label="访问次数">{{ selectedEntity.access_count }}</el-descriptions-item>
              <el-descriptions-item label="创建时间" :span="2">{{ formatDate(selectedEntity.created_at) }}</el-descriptions-item>
              <el-descriptions-item label="更新时间" :span="2">{{ formatDate(selectedEntity.updated_at) }}</el-descriptions-item>
              <el-descriptions-item label="描述" :span="2">{{ selectedEntity.description }}</el-descriptions-item>
            </el-descriptions>
          </el-tab-pane>
          
          <el-tab-pane label="属性" name="properties">
            <div class="entity-properties">
              <div v-if="selectedEntity.properties" class="properties-list">
                <div v-for="(value, key) in selectedEntity.properties" :key="key" class="property-item">
                  <div class="property-key">{{ key }}</div>
                  <div class="property-value">
                    <template v-if="typeof value === 'object'">
                      <pre>{{ JSON.stringify(value, null, 2) }}</pre>
                    </template>
                    <template v-else>
                      {{ value }}
                    </template>
                  </div>
                </div>
              </div>
              <el-empty v-else description="暂无属性数据" :image-size="80" />
            </div>
          </el-tab-pane>
          
          <el-tab-pane label="关系" name="relations">
            <div class="entity-relations">
              <div class="relations-section">
                <h4>出边关系 ({{ getEntityOutRelations(selectedEntity.id).length }})</h4>
                <div class="relation-list">
                  <div v-for="rel in getEntityOutRelations(selectedEntity.id)" :key="rel.id" class="relation-item">
                    <div class="relation-info">
                      <span class="relation-type">{{ getRelationTypeName(rel.relation_type) }}</span>
                      <el-icon><arrow-right /></el-icon>
                      <span class="target-entity">{{ getEntityName(rel.to_entity) }}</span>
                    </div>
                    <div class="relation-weight">权重: {{ rel.weight.toFixed(2) }}</div>
                  </div>
                </div>
              </div>
              
              <el-divider />
              
              <div class="relations-section">
                <h4>入边关系 ({{ getEntityInRelations(selectedEntity.id).length }})</h4>
                <div class="relation-list">
                  <div v-for="rel in getEntityInRelations(selectedEntity.id)" :key="rel.id" class="relation-item">
                    <div class="relation-info">
                      <span class="source-entity">{{ getEntityName(rel.from_entity) }}</span>
                      <el-icon><arrow-right /></el-icon>
                      <span class="relation-type">{{ getRelationTypeName(rel.relation_type) }}</span>
                    </div>
                    <div class="relation-weight">权重: {{ rel.weight.toFixed(2) }}</div>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>
          
          <el-tab-pane label="相似实体" name="similar">
            <div class="similar-entities">
              <el-button @click="findSimilarEntities" :loading="findingSimilar" style="margin-bottom: 16px;">
                <el-icon><search /></el-icon>
                查找相似实体
              </el-button>
              
              <div v-if="similarEntities.length > 0" class="similar-list">
                <div v-for="similar in similarEntities" :key="similar.id" class="similar-item">
                  <div class="similar-info">
                    <div class="similar-name">{{ similar.name }}</div>
                    <div class="similar-type">{{ similar.type }}</div>
                  </div>
                  <div class="similar-score">
                    相似度: {{ (similar.similarity * 100).toFixed(1) }}%
                  </div>
                </div>
              </div>
              
              <el-empty v-else-if="!findingSimilar" description="暂无相似实体" :image-size="80" />
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>
      
      <template #footer>
        <el-button @click="entityDetailVisible = false">关闭</el-button>
        <el-button type="primary" @click="editEntity(selectedEntity)">编辑实体</el-button>
      </template>
    </el-dialog>

    <!-- 实体编辑对话框 -->
    <el-dialog
      v-model="entityEditVisible"
      :title="editingEntity?.id ? '编辑实体' : '新建实体'"
      width="600px"
      :close-on-click-modal="false"
    >
      <el-form :model="entityForm" :rules="entityRules" ref="entityFormRef" label-width="100px">
        <el-form-item label="实体名称" prop="name">
          <el-input v-model="entityForm.name" placeholder="请输入实体名称" />
        </el-form-item>
        
        <el-form-item label="实体类型" prop="type">
          <el-select v-model="entityForm.type" placeholder="请选择实体类型" style="width: 100%">
            <el-option label="功能" value="Function" />
            <el-option label="领域" value="Domain" />
            <el-option label="模式" value="Pattern" />
            <el-option label="规则" value="Rule" />
            <el-option label="案例" value="Case" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="entityForm.description"
            type="textarea"
            :rows="4"
            placeholder="请输入实体描述"
          />
        </el-form-item>
        
        <el-form-item label="置信度" prop="confidence">
          <el-slider
            v-model="entityForm.confidence"
            :min="0"
            :max="1"
            :step="0.1"
            show-input
            :show-input-controls="false"
            input-size="small"
            style="width: 100%"
          />
        </el-form-item>
        
        <el-form-item label="来源" prop="source">
          <el-input v-model="entityForm.source" placeholder="请输入来源" />
        </el-form-item>
        
        <el-form-item label="属性">
          <div class="property-editor">
            <div v-for="(prop, index) in entityForm.properties" :key="index" class="property-row">
              <el-input v-model="prop.key" placeholder="属性名" style="width: 150px; margin-right: 8px;" />
              <el-input v-model="prop.value" placeholder="属性值" style="flex: 1; margin-right: 8px;" />
              <el-button @click="removeProperty(index)" type="danger" :icon="Close" circle size="small" />
            </div>
            <el-button @click="addProperty" type="primary" :icon="Plus" size="small">添加属性</el-button>
          </div>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="entityEditVisible = false">取消</el-button>
        <el-button type="primary" @click="saveEntity" :loading="saving">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as echarts from 'echarts'
import { 
  getKnowledgeEntities, 
  getKnowledgeRelations,
  createKnowledgeEntity,
  updateKnowledgeEntity,
  deleteKnowledgeEntity,
  searchSimilarEntities
} from '@/api/knowledge-graph'

// 响应式数据
const searchQuery = ref('')
const selectedEntityType = ref('')
const graphView = ref('force')
const showEntityLabels = ref(true)
const showRelationLabels = ref(false)
const showEntityTypes = ref(true)
const relationThreshold = ref(0.5)
const entityConfidence = ref(0.5)

const entities = ref([])
const relations = ref([])
const selectedEntity = ref(null)
const similarEntities = ref([])
const findingSimilar = ref(false)

// 分页
const currentPage = ref(1)
const pageSize = ref(20)
const relationPage = ref(1)
const relationPageSize = ref(20)

// 对话框状态
const entityDetailVisible = ref(false)
const entityEditVisible = ref(false)
const activeDetailTab = ref('basic')
const editingEntity = ref(null)
const saving = ref(false)

// 表单
const entityForm = reactive({
  name: '',
  type: '',
  description: '',
  confidence: 0.8,
  source: 'manual',
  properties: []
})

const entityRules = {
  name: [{ required: true, message: '请输入实体名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择实体类型', trigger: 'change' }],
  description: [{ required: true, message: '请输入实体描述', trigger: 'blur' }]
}

// 图表
const graphContainer = ref(null)
let graphInstance = null

// 计算属性
const filteredEntities = computed(() => {
  let filtered = entities.value

  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(entity => 
      entity.name.toLowerCase().includes(query) ||
      entity.description.toLowerCase().includes(query)
    )
  }

  if (selectedEntityType.value) {
    filtered = filtered.filter(entity => entity.type === selectedEntityType.value)
  }

  if (entityConfidence.value > 0) {
    filtered = filtered.filter(entity => entity.confidence >= entityConfidence.value)
  }

  return filtered
})

const paginatedEntities = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredEntities.value.slice(start, end)
})

const paginatedRelations = computed(() => {
  const start = (relationPage.value - 1) * relationPageSize.value
  const end = start + relationPageSize.value
  return relations.value.slice(start, end)
})

// 生命周期
onMounted(() => {
  initData()
  initGraph()
})

// 数据初始化
const initData = async () => {
  await Promise.all([
    loadEntities(),
    loadRelations()
  ])
  updateGraphData()
}

const loadEntities = async () => {
  try {
    // 模拟数据
    entities.value = [
      {
        id: 'entity_1',
        name: '用户登录功能',
        type: 'Function',
        description: '用户通过用户名和密码登录系统，验证身份后进入系统',
        confidence: 0.95,
        source: 'manual',
        access_count: 125,
        created_at: new Date('2024-01-15'),
        updated_at: new Date('2024-01-20'),
        properties: {
          function_type: 'EI',
          complexity: 'Low',
          points: 3,
          keywords: ['登录', '验证', '用户']
        }
      },
      {
        id: 'entity_2',
        name: '金融领域',
        type: 'Domain',
        description: '金融行业软件系统，包括银行、保险、证券等业务',
        confidence: 1.0,
        source: 'system',
        access_count: 89,
        created_at: new Date('2024-01-10'),
        updated_at: new Date('2024-01-18'),
        properties: {
          characteristics: ['高安全性', '强监管', '复杂计算'],
          patterns: ['账户管理', '交易处理', '风险控制'],
          adjustments: { security_factor: 1.2, compliance_factor: 1.1 }
        }
      },
      {
        id: 'entity_3',
        name: '表单输入模式',
        type: 'Pattern',
        description: '用户通过表单输入数据的通用模式',
        confidence: 0.88,
        source: 'ai_generated',
        access_count: 67,
        created_at: new Date('2024-01-12'),
        updated_at: new Date('2024-01-19'),
        properties: {
          keywords: ['表单', '输入', '提交'],
          structure: ['表单界面', '数据验证', '数据存储'],
          function_type: 'EI'
        }
      }
    ]
  } catch (error) {
    console.error('Failed to load entities:', error)
    ElMessage.error('加载实体数据失败')
  }
}

const loadRelations = async () => {
  try {
    // 模拟数据
    relations.value = [
      {
        id: 'rel_1',
        from_entity: 'entity_1',
        to_entity: 'entity_3',
        from_entity_name: '用户登录功能',
        to_entity_name: '表单输入模式',
        relation_type: 'implements',
        weight: 0.9,
        confidence: 0.85,
        created_at: new Date('2024-01-16')
      },
      {
        id: 'rel_2',
        from_entity: 'entity_1',
        to_entity: 'entity_2',
        from_entity_name: '用户登录功能',
        to_entity_name: '金融领域',
        relation_type: 'belongs_to',
        weight: 0.7,
        confidence: 0.92,
        created_at: new Date('2024-01-17')
      }
    ]
  } catch (error) {
    console.error('Failed to load relations:', error)
    ElMessage.error('加载关系数据失败')
  }
}

// 图谱初始化
const initGraph = () => {
  nextTick(() => {
    if (graphContainer.value) {
      graphInstance = echarts.init(graphContainer.value)
      updateGraphData()
    }
  })
}

const updateGraphData = () => {
  if (!graphInstance) return

  const nodes = entities.value.map(entity => ({
    id: entity.id,
    name: entity.name,
    category: entity.type,
    value: entity.confidence,
    symbolSize: 20 + entity.confidence * 30,
    itemStyle: {
      color: getEntityTypeColor(entity.type, true)
    }
  }))

  const links = relations.value
    .filter(rel => rel.weight >= relationThreshold.value)
    .map(rel => ({
      source: rel.from_entity,
      target: rel.to_entity,
      lineStyle: {
        width: rel.weight * 5,
        opacity: rel.confidence
      },
      label: {
        show: showRelationLabels.value,
        formatter: getRelationTypeName(rel.relation_type)
      }
    }))

  const categories = [
    { name: 'Function', itemStyle: { color: '#5470C6' } },
    { name: 'Domain', itemStyle: { color: '#91CC75' } },
    { name: 'Pattern', itemStyle: { color: '#FAC858' } },
    { name: 'Rule', itemStyle: { color: '#EE6666' } },
    { name: 'Case', itemStyle: { color: '#73C0DE' } }
  ]

  const option = {
    title: {
      text: `知识图谱 (${nodes.length} 实体, ${links.length} 关系)`,
      left: 'center',
      textStyle: { fontSize: 16 }
    },
    legend: {
      show: showEntityTypes.value,
      data: categories.map(cat => cat.name),
      bottom: 20
    },
    series: [{
      type: 'graph',
      layout: graphView.value === 'force' ? 'force' : 'circular',
      data: nodes,
      links: links,
      categories: categories,
      roam: true,
      label: {
        show: showEntityLabels.value,
        position: 'right',
        formatter: '{b}'
      },
      force: {
        repulsion: 1000,
        gravity: 0.1,
        edgeLength: 150,
        layoutAnimation: true
      },
      emphasis: {
        focus: 'adjacency',
        lineStyle: {
          width: 6
        }
      }
    }]
  }

  graphInstance.setOption(option, true)
}

// 事件处理
const handleSearch = () => {
  currentPage.value = 1
  // 触发计算属性更新
}

const setGraphView = (view) => {
  graphView.value = view
  updateGraphData()
}

const resetGraphView = () => {
  if (graphInstance) {
    graphInstance.dispatchAction({
      type: 'restore'
    })
  }
}

const centerGraph = () => {
  if (graphInstance) {
    graphInstance.dispatchAction({
      type: 'restore'
    })
  }
}

const selectEntity = (entity) => {
  selectedEntity.value = entity
  // 高亮图谱中的实体
  if (graphInstance) {
    graphInstance.dispatchAction({
      type: 'highlight',
      dataIndex: entities.value.findIndex(e => e.id === entity.id)
    })
  }
}

const viewEntityDetails = (entity) => {
  selectedEntity.value = entity
  entityDetailVisible.value = true
  activeDetailTab.value = 'basic'
}

const editEntity = (entity) => {
  editingEntity.value = entity
  if (entity) {
    entityForm.name = entity.name
    entityForm.type = entity.type
    entityForm.description = entity.description
    entityForm.confidence = entity.confidence
    entityForm.source = entity.source
    entityForm.properties = entity.properties ? 
      Object.entries(entity.properties).map(([key, value]) => ({ key, value: JSON.stringify(value) })) :
      []
  } else {
    Object.assign(entityForm, {
      name: '',
      type: '',
      description: '',
      confidence: 0.8,
      source: 'manual',
      properties: []
    })
  }
  entityEditVisible.value = true
}

const saveEntity = async () => {
  // 表单验证
  const form = entityFormRef.value
  if (!form) return

  try {
    await form.validate()
    saving.value = true

    // 处理属性
    const properties = {}
    entityForm.properties.forEach(prop => {
      if (prop.key && prop.value) {
        try {
          properties[prop.key] = JSON.parse(prop.value)
        } catch {
          properties[prop.key] = prop.value
        }
      }
    })

    const entityData = {
      name: entityForm.name,
      type: entityForm.type,
      description: entityForm.description,
      confidence: entityForm.confidence,
      source: entityForm.source,
      properties
    }

    if (editingEntity.value) {
      // 更新实体
      const index = entities.value.findIndex(e => e.id === editingEntity.value.id)
      if (index > -1) {
        entities.value[index] = { ...editingEntity.value, ...entityData, updated_at: new Date() }
      }
      ElMessage.success('实体更新成功')
    } else {
      // 创建实体
      const newEntity = {
        id: `entity_${Date.now()}`,
        ...entityData,
        access_count: 0,
        created_at: new Date(),
        updated_at: new Date()
      }
      entities.value.push(newEntity)
      ElMessage.success('实体创建成功')
    }

    entityEditVisible.value = false
    updateGraphData()
  } catch {
    // 验证失败
  } finally {
    saving.value = false
  }
}

const handleEntityAction = async (command) => {
  const { action, entity } = command
  
  switch (action) {
    case 'find_similar':
      await findSimilarEntitiesFor(entity)
      break
    case 'analyze_relations':
      analyzeEntityRelations(entity)
      break
    case 'export_entity':
      exportEntity(entity)
      break
    case 'delete_entity':
      await deleteEntityConfirm(entity)
      break
  }
}

const findSimilarEntities = async () => {
  if (!selectedEntity.value) return
  
  findingSimilar.value = true
  try {
    // 模拟相似实体查找
    await new Promise(resolve => setTimeout(resolve, 1000))
    
    similarEntities.value = [
      {
        id: 'similar_1',
        name: '用户注册功能',
        type: 'Function',
        similarity: 0.85
      },
      {
        id: 'similar_2',
        name: '密码重置功能',
        type: 'Function',
        similarity: 0.72
      }
    ]
  } catch (error) {
    ElMessage.error('查找相似实体失败')
  } finally {
    findingSimilar.value = false
  }
}

const findSimilarEntitiesFor = async (entity) => {
  selectedEntity.value = entity
  entityDetailVisible.value = true
  activeDetailTab.value = 'similar'
  await findSimilarEntities()
}

const addProperty = () => {
  entityForm.properties.push({ key: '', value: '' })
}

const removeProperty = (index) => {
  entityForm.properties.splice(index, 1)
}

const refreshEntities = () => {
  loadEntities()
  ElMessage.success('实体数据已刷新')
}

// 工具函数
const getEntityTypeColor = (type, returnHex = false) => {
  const colors = {
    'Function': returnHex ? '#5470C6' : 'primary',
    'Domain': returnHex ? '#91CC75' : 'success',
    'Pattern': returnHex ? '#FAC858' : 'warning',
    'Rule': returnHex ? '#EE6666' : 'danger',
    'Case': returnHex ? '#73C0DE' : 'info'
  }
  return colors[type] || (returnHex ? '#909399' : '')
}

const getConfidenceColor = (confidence) => {
  if (confidence >= 0.8) return '#67C23A'
  if (confidence >= 0.6) return '#E6A23C'
  return '#F56C6C'
}

const getRelationTypeColor = (type) => {
  const colors = {
    'contains': 'primary',
    'depends_on': 'warning',
    'similar_to': 'success',
    'implements': 'info',
    'belongs_to': 'primary',
    'conflicts_with': 'danger'
  }
  return colors[type] || ''
}

const getRelationTypeName = (type) => {
  const names = {
    'contains': '包含',
    'depends_on': '依赖',
    'similar_to': '相似',
    'implements': '实现',
    'belongs_to': '属于',
    'conflicts_with': '冲突'
  }
  return names[type] || type
}

const getWeightColor = (weight) => {
  if (weight >= 0.7) return '#67C23A'
  if (weight >= 0.4) return '#E6A23C'
  return '#F56C6C'
}

const getEntityRelationCount = (entityId) => {
  return relations.value.filter(rel => 
    rel.from_entity === entityId || rel.to_entity === entityId
  ).length
}

const getEntityOutRelations = (entityId) => {
  return relations.value.filter(rel => rel.from_entity === entityId)
}

const getEntityInRelations = (entityId) => {
  return relations.value.filter(rel => rel.to_entity === entityId)
}

const getEntityName = (entityId) => {
  const entity = entities.value.find(e => e.id === entityId)
  return entity ? entity.name : entityId
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString()
}

// 占位符实现
const handleAddEntity = () => {
  editEntity(null)
}

const handleAddRelation = () => {
  ElMessage.info('添加关系功能开发中...')
}

const handleImportKnowledge = () => {
  ElMessage.info('导入知识功能开发中...')
}

const handleExportKnowledge = () => {
  ElMessage.info('导出知识功能开发中...')
}

const editRelation = (relation) => {
  ElMessage.info('编辑关系功能开发中...')
}

const deleteRelation = async (relation) => {
  try {
    await ElMessageBox.confirm('确定要删除这个关系吗？', '确认删除', { type: 'warning' })
    const index = relations.value.findIndex(r => r.id === relation.id)
    if (index > -1) {
      relations.value.splice(index, 1)
      updateGraphData()
      ElMessage.success('关系删除成功')
    }
  } catch {
    // 用户取消
  }
}

const analyzeRelations = () => {
  ElMessage.info('关系分析功能开发中...')
}

const optimizeGraph = () => {
  ElMessage.info('图谱优化功能开发中...')
}

const analyzeEntityRelations = (entity) => {
  ElMessage.info('实体关系分析功能开发中...')
}

const exportEntity = (entity) => {
  ElMessage.info('实体导出功能开发中...')
}

const deleteEntityConfirm = async (entity) => {
  try {
    await ElMessageBox.confirm(`确定要删除实体"${entity.name}"吗？`, '确认删除', { type: 'warning' })
    const index = entities.value.findIndex(e => e.id === entity.id)
    if (index > -1) {
      entities.value.splice(index, 1)
      // 同时删除相关关系
      relations.value = relations.value.filter(rel => 
        rel.from_entity !== entity.id && rel.to_entity !== entity.id
      )
      updateGraphData()
      ElMessage.success('实体删除成功')
    }
  } catch {
    // 用户取消
  }
}
</script>

<style scoped>
.knowledge-graph {
  padding: 20px;
}

.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.toolbar-left, .toolbar-right {
  display: flex;
  gap: 10px;
  align-items: center;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.graph-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.graph-container {
  position: relative;
  height: 600px;
  border: 1px solid #ebeef5;
  border-radius: 8px;
}

.graph-canvas {
  width: 100%;
  height: 100%;
}

.graph-control-panel {
  position: absolute;
  top: 10px;
  right: 10px;
  background: rgba(255, 255, 255, 0.95);
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  min-width: 200px;
  backdrop-filter: blur(10px);
}

.control-section {
  margin-bottom: 16px;
}

.control-section:last-child {
  margin-bottom: 0;
}

.control-title {
  font-size: 14px;
  font-weight: 600;
  color: #333;
  margin-bottom: 8px;
}

.entity-list {
  max-height: 500px;
  overflow-y: auto;
}

.entity-item {
  border: 1px solid #ebeef5;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 12px;
  cursor: pointer;
  transition: all 0.3s;
}

.entity-item:hover {
  border-color: #409EFF;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.entity-item.selected {
  border-color: #409EFF;
  background-color: #f0f8ff;
}

.entity-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
}

.entity-info .entity-name {
  font-size: 16px;
  font-weight: 600;
  color: #333;
  margin-bottom: 4px;
}

.entity-confidence {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 80px;
}

.confidence-text {
  font-size: 12px;
  color: #666;
}

.entity-description {
  font-size: 14px;
  color: #666;
  line-height: 1.4;
  margin-bottom: 12px;
}

.entity-metadata {
  display: flex;
  gap: 16px;
  margin-bottom: 12px;
}

.metadata-item {
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 12px;
  color: #666;
}

.entity-actions {
  display: flex;
  gap: 8px;
}

.pagination-container {
  margin-top: 16px;
  text-align: center;
}

.entity-details {
  min-height: 400px;
}

.entity-properties {
  padding: 20px 0;
}

.properties-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.property-item {
  display: flex;
  gap: 12px;
  align-items: flex-start;
}

.property-key {
  font-weight: 600;
  color: #333;
  min-width: 120px;
}

.property-value {
  flex: 1;
  color: #666;
}

.property-value pre {
  background: #f5f7fa;
  padding: 8px;
  border-radius: 4px;
  font-size: 12px;
  margin: 0;
}

.entity-relations {
  padding: 20px 0;
}

.relations-section h4 {
  margin-bottom: 12px;
  color: #333;
}

.relation-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.relation-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: #f8f9fa;
  border-radius: 4px;
}

.relation-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.relation-type {
  background: #409EFF;
  color: white;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
}

.target-entity, .source-entity {
  font-weight: 500;
  color: #333;
}

.relation-weight {
  font-size: 12px;
  color: #666;
}

.similar-entities {
  padding: 20px 0;
}

.similar-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.similar-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px;
  border: 1px solid #ebeef5;
  border-radius: 6px;
}

.similar-info .similar-name {
  font-weight: 600;
  color: #333;
}

.similar-info .similar-type {
  font-size: 12px;
  color: #666;
}

.similar-score {
  font-size: 12px;
  color: #409EFF;
  font-weight: 600;
}

.property-editor {
  width: 100%;
}

.property-row {
  display: flex;
  align-items: center;
  margin-bottom: 8px;
}

.header-actions {
  display: flex;
  gap: 8px;
}
</style>