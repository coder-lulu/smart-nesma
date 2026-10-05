<template>
  <el-dialog
    v-model="visible"
    title="需求智能细化"
    width="80%"
    :close-on-click-modal="false"
    @closed="handleClose"
  >
    <div class="refinement-container">
      <!-- 左侧：原始需求 -->
      <div class="original-requirement">
        <h4>原始需求</h4>
        <el-card class="requirement-card">
          <template #header>
            <div class="requirement-header">
              <span class="requirement-title">{{ requirement.title }}</span>
              <el-tag :type="getPriorityType(requirement.priority)">
                优先级{{ requirement.priority }}
              </el-tag>
            </div>
          </template>
          <div class="requirement-content">
            <p><strong>需求描述：</strong></p>
            <p class="requirement-description">{{ requirement.description }}</p>
            <p><strong>需求类型：</strong>{{ requirement.category }}</p>
            <p><strong>所属项目：</strong>{{ requirement.projectName }}</p>
          </div>
        </el-card>
      </div>

      <!-- 右侧：细化建议 -->
      <div class="refinement-suggestions">
        <div class="suggestions-header">
          <h4>智能细化建议</h4>
          <div class="header-actions">
            <el-button
              type="primary"
              :loading="analyzing"
              @click="startAnalysis"
            >
              <el-icon><Document /></el-icon>
              {{ analyzing ? '正在分析...' : '重新分析' }}
            </el-button>
            <el-button @click="exportSuggestions">
              <el-icon><Download /></el-icon>
              导出建议
            </el-button>
          </div>
        </div>

        <div v-if="analyzing" class="analysis-loading">
          <el-skeleton :rows="5" animated />
          <div class="loading-text">
            <el-icon class="loading-icon"><Loading /></el-icon>
            AI正在分析需求并生成细化建议...
          </div>
        </div>

        <div v-else-if="suggestions.length > 0" class="suggestions-content">
          <!-- 分析配置 -->
          <div class="analysis-config">
            <el-form :model="analysisOptions" label-width="100px" size="small">
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-form-item label="分析详细度">
                    <el-select v-model="analysisOptions.detailLevel">
                      <el-option label="基础" value="basic" />
                      <el-option label="详细" value="detailed" />
                      <el-option label="全面" value="comprehensive" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="AI模型">
                    <el-select v-model="analysisOptions.aiModel">
                      <el-option label="DeepSeek" value="deepseek" />
                      <el-option label="GPT-4" value="gpt-4" />
                      <el-option label="Claude" value="claude" />
                    </el-select>
                  </el-form-item>
                </el-col>
                <el-col :span="8">
                  <el-form-item label="包含示例">
                    <el-switch v-model="analysisOptions.includeExamples" />
                  </el-form-item>
                </el-col>
              </el-row>
            </el-form>
          </div>

          <!-- 分析结果统计 -->
          <div class="analysis-summary">
            <el-row :gutter="20">
              <el-col :span="6">
                <div class="summary-item">
                  <div class="summary-number">{{ suggestions.length }}</div>
                  <div class="summary-label">建议总数</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="summary-item">
                  <div class="summary-number">{{ highPrioritySuggestions }}</div>
                  <div class="summary-label">高优先级</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="summary-item">
                  <div class="summary-number">{{ getConfidenceScore() }}%</div>
                  <div class="summary-label">置信度</div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="summary-item">
                  <div class="summary-number">{{ processedSuggestions }}</div>
                  <div class="summary-label">已处理</div>
                </div>
              </el-col>
            </el-row>
          </div>

          <!-- 建议列表 -->
          <div class="suggestions-list">
            <div class="filter-tabs">
              <el-radio-group v-model="activeFilter" @change="filterSuggestions">
                <el-radio-button label="all">全部</el-radio-button>
                <el-radio-button label="add_detail">补充细节</el-radio-button>
                <el-radio-button label="clarify">澄清需求</el-radio-button>
                <el-radio-button label="split">拆分需求</el-radio-button>
                <el-radio-button label="merge">合并需求</el-radio-button>
              </el-radio-group>
            </div>

            <div class="suggestions-container">
              <div
                v-for="(suggestion, index) in filteredSuggestions"
                :key="index"
                class="suggestion-item"
                :class="{ 'suggestion-processed': suggestion.processed }"
              >
                <div class="suggestion-header">
                  <div class="suggestion-meta">
                    <el-tag :type="getSuggestionTypeColor(suggestion.type)">
                      {{ getSuggestionTypeLabel(suggestion.type) }}
                    </el-tag>
                    <el-tag :type="getPriorityColor(suggestion.priority)">
                      优先级{{ suggestion.priority }}
                    </el-tag>
                    <el-tag :type="getImpactColor(suggestion.impact)">
                      {{ getImpactLabel(suggestion.impact) }}
                    </el-tag>
                  </div>
                  <div class="suggestion-actions">
                    <el-button
                      type="text"
                      :class="{ 'processed': suggestion.processed }"
                      @click="toggleSuggestionStatus(suggestion)"
                    >
                      <el-icon>
                        {{ suggestion.processed ? '<Check />' : '<Plus />' }}
                      </el-icon>
                      {{ suggestion.processed ? '已处理' : '标记处理' }}
                    </el-button>
                    <el-button
                      type="text"
                      @click="applySuggestion(suggestion)"
                    >
                      <el-icon><Edit /></el-icon>
                      应用建议
                    </el-button>
                  </div>
                </div>
                
                <div class="suggestion-content">
                  <h5>{{ suggestion.title }}</h5>
                  <p class="suggestion-description">{{ suggestion.description }}</p>
                  
                  <div v-if="suggestion.examples && suggestion.examples.length > 0" class="suggestion-examples">
                    <h6>改进示例：</h6>
                    <ul>
                      <li v-for="(example, idx) in suggestion.examples" :key="idx">
                        {{ example }}
                      </li>
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else class="empty-suggestions">
          <div class="empty-icon">
            <el-icon size="48"><Document /></el-icon>
          </div>
          <p>暂无细化建议</p>
          <p>点击"开始分析"按钮生成智能细化建议</p>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="visible = false">取消</el-button>
        <el-button type="primary" @click="saveRefinement">
          <el-icon><Check /></el-icon>
          保存细化结果
        </el-button>
      </div>
    </template>

    <!-- 应用建议对话框 -->
    <el-dialog
      v-model="applySuggestionDialog"
      title="应用建议"
      width="60%"
      append-to-body
    >
      <div class="apply-suggestion-content">
        <h4>{{ currentSuggestion.title }}</h4>
        <p class="suggestion-desc">{{ currentSuggestion.description }}</p>
        
        <el-form :model="refinementForm" label-width="100px">
          <el-form-item label="新需求标题">
            <el-input v-model="refinementForm.title" />
          </el-form-item>
          <el-form-item label="需求描述">
            <el-input
              v-model="refinementForm.description"
              type="textarea"
              :rows="4"
            />
          </el-form-item>
          <el-form-item label="需求类型">
            <el-select v-model="refinementForm.category">
              <el-option label="功能需求" value="functional" />
              <el-option label="非功能需求" value="non-functional" />
              <el-option label="约束需求" value="constraint" />
            </el-select>
          </el-form-item>
          <el-form-item label="优先级">
            <el-select v-model="refinementForm.priority">
              <el-option label="1 - 最高" :value="1" />
              <el-option label="2 - 高" :value="2" />
              <el-option label="3 - 中" :value="3" />
              <el-option label="4 - 低" :value="4" />
              <el-option label="5 - 最低" :value="5" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
      
      <template #footer>
        <el-button @click="applySuggestionDialog = false">取消</el-button>
        <el-button type="primary" @click="confirmApplySuggestion">确定</el-button>
      </template>
    </el-dialog>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, defineProps, defineEmits } from 'vue'
import { ElMessage } from 'element-plus'
import { Download, Loading, Check, Plus, Edit, Document } from '@element-plus/icons-vue'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  requirement: {
    type: Object,
    default: () => ({})
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'applied'])

// 响应式数据
const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const analyzing = ref(false)
const suggestions = ref([])
const activeFilter = ref('all')
const processedSuggestions = ref(0)
const applySuggestionDialog = ref(false)
const currentSuggestion = ref({})

// 分析选项
const analysisOptions = reactive({
  detailLevel: 'detailed',
  aiModel: 'deepseek',
  includeExamples: true,
  maxSubRequirements: 10,
  focusAreas: []
})

// 应用建议表单
const refinementForm = reactive({
  title: '',
  description: '',
  category: 'functional',
  priority: 3
})

// 计算属性
const filteredSuggestions = computed(() => {
  if (activeFilter.value === 'all') {
    return suggestions.value
  }
  return suggestions.value.filter(s => s.type === activeFilter.value)
})

const highPrioritySuggestions = computed(() => {
  return suggestions.value.filter(s => s.priority >= 4).length
})

// 方法
const startAnalysis = async () => {
  analyzing.value = true
  try {
    // 模拟AI分析过程
    await new Promise(resolve => setTimeout(resolve, 3000))
    
    // 生成模拟建议
    suggestions.value = [
      {
        type: 'add_detail',
        priority: 5,
        title: '补充验收条件',
        description: '当前需求缺少明确的验收条件，建议添加具体的成功标准和可测试的验收条件。',
        examples: [
          '用户可以在30秒内完成登录操作',
          '系统响应时间不超过2秒',
          '支持同时在线用户数不少于1000人'
        ],
        impact: 'high',
        processed: false
      },
      {
        type: 'clarify',
        priority: 4,
        title: '澄清用户角色',
        description: '需求中提到的"用户"概念过于宽泛，建议明确具体的用户角色和权限。',
        examples: [
          '区分管理员、普通用户、访客等角色',
          '定义各角色的操作权限',
          '明确用户认证和授权机制'
        ],
        impact: 'medium',
        processed: false
      },
      {
        type: 'split',
        priority: 3,
        title: '拆分复杂功能',
        description: '当前需求包含多个功能点，建议拆分为独立的子需求以便更好地管理和实现。',
        examples: [
          '将用户管理拆分为用户注册、用户登录、用户信息管理',
          '将数据导出功能独立为单独需求',
          '将权限管理作为独立的功能模块'
        ],
        impact: 'high',
        processed: false
      }
    ]
    
    ElMessage.success('分析完成')
  } catch (error) {
    console.error('分析失败:', error)
    ElMessage.error('分析失败，请重试')
  } finally {
    analyzing.value = false
  }
}

const toggleSuggestionStatus = (suggestion) => {
  suggestion.processed = !suggestion.processed
  processedSuggestions.value = suggestions.value.filter(s => s.processed).length
}

const applySuggestion = (suggestion) => {
  currentSuggestion.value = suggestion
  refinementForm.title = suggestion.title
  refinementForm.description = suggestion.description
  applySuggestionDialog.value = true
}

const confirmApplySuggestion = () => {
  emit('applied', {
    suggestion: currentSuggestion.value,
    refinement: refinementForm
  })
  applySuggestionDialog.value = false
  toggleSuggestionStatus(currentSuggestion.value)
  ElMessage.success('建议已应用')
}

const saveRefinement = () => {
  ElMessage.success('细化结果已保存')
  visible.value = false
}

const exportSuggestions = () => {
  const data = {
    requirement: props.requirement,
    suggestions: suggestions.value,
    analysisOptions: analysisOptions,
    timestamp: new Date().toISOString()
  }
  
  const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `requirement_refinement_${props.requirement.id}.json`
  a.click()
  URL.revokeObjectURL(url)
  
  ElMessage.success('建议已导出')
}

const filterSuggestions = () => {
  // 过滤逻辑已在计算属性中实现
}

const getConfidenceScore = () => {
  if (suggestions.value.length === 0) return 0
  const avg = suggestions.value.reduce((sum, s) => sum + s.priority, 0) / suggestions.value.length
  return Math.round(avg * 20) // 转换为百分比
}

const handleClose = () => {
  suggestions.value = []
  processedSuggestions.value = 0
  activeFilter.value = 'all'
}

// 样式相关方法
const getPriorityType = (priority) => {
  const types = { 1: 'danger', 2: 'warning', 3: 'info', 4: 'success', 5: 'primary' }
  return types[priority] || 'info'
}

const getSuggestionTypeColor = (type) => {
  const colors = {
    add_detail: 'primary',
    clarify: 'warning',
    split: 'success',
    merge: 'info'
  }
  return colors[type] || 'info'
}

const getSuggestionTypeLabel = (type) => {
  const labels = {
    add_detail: '补充细节',
    clarify: '澄清需求',
    split: '拆分需求',
    merge: '合并需求'
  }
  return labels[type] || type
}

const getPriorityColor = (priority) => {
  if (priority >= 4) return 'danger'
  if (priority >= 3) return 'warning'
  return 'success'
}

const getImpactColor = (impact) => {
  const colors = { high: 'danger', medium: 'warning', low: 'success' }
  return colors[impact] || 'info'
}

const getImpactLabel = (impact) => {
  const labels = { high: '高影响', medium: '中影响', low: '低影响' }
  return labels[impact] || impact
}
</script>

<style scoped>
.refinement-container {
  display: flex;
  gap: 20px;
  height: 600px;
}

.original-requirement {
  flex: 1;
}

.requirement-card {
  height: 100%;
}

.requirement-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.requirement-title {
  font-weight: bold;
  font-size: 16px;
}

.requirement-content {
  line-height: 1.6;
}

.requirement-description {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 4px;
  margin: 8px 0;
  border-left: 4px solid #409eff;
}

.refinement-suggestions {
  flex: 2;
  display: flex;
  flex-direction: column;
}

.suggestions-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.suggestions-header h4 {
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.analysis-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.loading-text {
  margin-top: 20px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: #606266;
}

.loading-icon {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.suggestions-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.analysis-config {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 8px;
}

.analysis-summary {
  background: white;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.summary-item {
  text-align: center;
}

.summary-number {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}

.summary-label {
  font-size: 14px;
  color: #606266;
  margin-top: 4px;
}

.suggestions-list {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.filter-tabs {
  margin-bottom: 16px;
}

.suggestions-container {
  flex: 1;
  overflow-y: auto;
}

.suggestion-item {
  background: white;
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  margin-bottom: 12px;
  padding: 16px;
  transition: all 0.2s;
}

.suggestion-item:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.suggestion-processed {
  background: #f0f9ff;
  border-color: #409eff;
}

.suggestion-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.suggestion-meta {
  display: flex;
  gap: 8px;
}

.suggestion-actions {
  display: flex;
  gap: 8px;
}

.suggestion-content h5 {
  margin: 0 0 8px 0;
  color: #303133;
}

.suggestion-description {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 12px;
}

.suggestion-examples {
  background: #f8f9fa;
  padding: 12px;
  border-radius: 4px;
  border-left: 4px solid #67c23a;
}

.suggestion-examples h6 {
  margin: 0 0 8px 0;
  color: #303133;
}

.suggestion-examples ul {
  margin: 0;
  padding-left: 20px;
}

.suggestion-examples li {
  margin-bottom: 4px;
  color: #606266;
}

.empty-suggestions {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  color: #909399;
}

.empty-icon {
  margin-bottom: 16px;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

.apply-suggestion-content {
  padding: 20px 0;
}

.apply-suggestion-content h4 {
  margin: 0 0 8px 0;
  color: #303133;
}

.suggestion-desc {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 20px;
  padding: 12px;
  background: #f8f9fa;
  border-radius: 4px;
}

.processed {
  color: #67c23a !important;
}
</style> 