<template>
  <div class="level4-generator">
    <warning-bar title="L4功能点智能生成器 - 基于L3需求自动生成详细功能点" />
    
    <div class="gva-table-box">
      <!-- 项目上下文选择 -->
      <div class="gva-search-box">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="当前项目">
            <el-select v-model="currentProject" placeholder="选择项目" @change="handleProjectChange">
              <el-option
                v-for="project in projects"
                :key="project.ID || project.id"
                :label="project.name"
                :value="project.ID || project.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="项目周期">
            <el-select v-model="currentCycle" placeholder="选择周期" @change="handleCycleChange" :disabled="!currentProject">
              <el-option
                v-for="cycle in cycles"
                :key="cycle.ID"
                :label="cycle.name"
                :value="cycle.ID"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="需求版本">
            <el-select v-model="currentVersion" placeholder="选择版本" @change="handleVersionChange" :disabled="!currentCycle">
              <el-option
                v-for="version in versions"
                :key="version.ID"
                :label="`${version.summary} (${version.version})`"
                :value="version.ID"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- 生成策略配置 -->
      <div class="generator-config">
        <div class="config-section">
          <h3>生成策略</h3>
          <div class="config-row">
            <el-select v-model="generationStrategy" placeholder="选择生成策略" style="width: 200px; margin-right: 16px;">
              <el-option label="智能分解" value="decomposition" />
              <el-option label="场景扩展" value="scenario_expansion" />
              <el-option label="流程细化" value="process_refinement" />
              <el-option label="综合生成" value="comprehensive" />
            </el-select>
            
            <el-select v-model="complexityLevel" placeholder="复杂度级别" style="width: 150px; margin-right: 16px;">
              <el-option label="简单" value="simple" />
              <el-option label="中等" value="moderate" />
              <el-option label="复杂" value="complex" />
            </el-select>

            <el-input-number 
              v-model="maxL4Count" 
              :min="1" 
              :max="20" 
              placeholder="最大生成数量"
              style="width: 150px; margin-right: 16px;"
            />

            <el-switch
              v-model="includeKnowledgeBase"
              active-text="使用知识库"
              inactive-text="不使用知识库"
              style="margin-right: 16px;"
            />
          </div>
        </div>

        <div class="action-buttons">
          <el-button type="primary" @click="generateL4Requirements" :loading="generating" :disabled="!canGenerate">
            <!-- <el-icon><Magic /></el-icon> -->
            生成L4功能点
          </el-button>
          <el-button type="success" @click="batchGenerate" :loading="batchGenerating" :disabled="!canGenerate">
            <el-icon><Operation /></el-icon>
            批量生成
          </el-button>
          <el-button @click="showHistory" :disabled="!currentCycle">
            <el-icon><Clock /></el-icon>
            生成历史
          </el-button>
          <el-button @click="showStats" :disabled="!currentCycle">
            <el-icon><DataAnalysis /></el-icon>
            生成统计
          </el-button>
        </div>
      </div>

      <!-- L3需求选择器 -->
      <div class="l3-requirement-selector">
        <h3>选择L3需求 ({{ selectedL3Requirements.length }} 个已选择)</h3>
        <el-table 
          :data="l3Requirements" 
          @selection-change="handleL3Selection"
          row-key="id"
          max-height="350"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="code" label="编号" width="120" />
          <el-table-column prop="title" label="需求名称" min-width="250" />
          <el-table-column prop="description" label="描述" min-width="300" show-overflow-tooltip />
          <el-table-column prop="function_type" label="功能类型" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.function_type" size="small">{{ row.function_type }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="complexity" label="复杂度" width="100">
            <template #default="{ row }">
              <el-tag v-if="row.complexity" :type="getComplexityType(row.complexity)" size="small">
                {{ row.complexity }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="现有L4" width="100">
            <template #default="{ row }">
              <el-button size="small" text @click="viewExistingL4(row)">
                {{ row.existing_l4_count || 0 }} 个
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 生成结果展示 -->
      <div class="generation-results" v-if="generationResults.length > 0">
        <h3>生成结果 ({{ generationResults.length }} 个L3需求，共 {{ totalL4Count }} 个L4功能点)</h3>
        
        <div class="result-actions">
          <el-button type="success" @click="applyAllSuggestions" :disabled="!hasValidSuggestions">
            <el-icon><Check /></el-icon>
            应用全部建议
          </el-button>
          <el-button @click="exportResults">
            <el-icon><Download /></el-icon>
            导出结果
          </el-button>
          <el-button @click="clearResults">
            <el-icon><Delete /></el-icon>
            清空结果
          </el-button>
        </div>

        <el-collapse v-model="activeResultPanels" accordion>
          <el-collapse-item 
            v-for="(result, index) in generationResults" 
            :key="index"
            :title="`${result.l3Requirement.title} (${result.l4Suggestions.length} 个L4建议)`"
            :name="`result-${index}`"
          >
            <div class="l3-result-content">
              <!-- L3需求信息 -->
              <div class="l3-info">
                <el-descriptions :column="2" size="small" border>
                  <el-descriptions-item label="需求编号">{{ result.l3Requirement.code }}</el-descriptions-item>
                  <el-descriptions-item label="生成策略">{{ result.strategy }}</el-descriptions-item>
                  <el-descriptions-item label="生成时间">{{ formatDate(result.createdAt) }}</el-descriptions-item>
                  <el-descriptions-item label="建议数量">{{ result.l4Suggestions.length }} 个</el-descriptions-item>
                </el-descriptions>
              </div>

              <!-- L4建议列表 -->
              <div class="l4-suggestions">
                <h4>L4功能点建议</h4>
                <div class="suggestion-grid">
                  <div 
                    v-for="(suggestion, sIndex) in result.l4Suggestions" 
                    :key="sIndex"
                    class="suggestion-card"
                    :class="{ selected: suggestion.selected }"
                    @click="toggleSuggestion(result, sIndex)"
                  >
                    <div class="suggestion-header">
                      <div class="suggestion-title">
                        <el-checkbox 
                          v-model="suggestion.selected" 
                          @change="updateSuggestionSelection(result, sIndex)"
                          @click.stop
                        />
                        <span>{{ suggestion.suggested_title }}</span>
                      </div>
                      <div class="suggestion-meta">
                        <el-tag size="small" :type="getFunctionTypeColor(suggestion.function_type)">
                          {{ suggestion.function_type }}
                        </el-tag>
                        <el-tag size="small" :type="getComplexityType(suggestion.estimated_complexity)">
                          {{ suggestion.estimated_complexity }}
                        </el-tag>
                      </div>
                    </div>
                    
                    <div class="suggestion-content">
                      <p class="suggestion-description">{{ suggestion.suggested_description }}</p>
                      
                      <div class="suggestion-details">
                        <el-row :gutter="12">
                          <el-col :span="8">
                            <div class="detail-item">
                              <span class="label">建议AFP:</span>
                              <span class="value">{{ suggestion.recommended_afp }}</span>
                            </div>
                          </el-col>
                          <el-col :span="8">
                            <div class="detail-item">
                              <span class="label">建议UFP:</span>
                              <span class="value">{{ suggestion.recommended_ufp }}</span>
                            </div>
                          </el-col>
                          <el-col :span="8">
                            <div class="detail-item">
                              <span class="label">置信度:</span>
                              <span class="value">{{ (suggestion.confidence * 100).toFixed(1) }}%</span>
                            </div>
                          </el-col>
                        </el-row>
                      </div>

                      <div class="suggestion-reason" v-if="suggestion.generation_reason">
                        <el-icon><InfoFilled /></el-icon>
                        <span>{{ suggestion.generation_reason }}</span>
                      </div>
                    </div>

                    <div class="suggestion-actions">
                      <el-button size="small" @click.stop="editSuggestion(result, sIndex)">
                        <el-icon><Edit /></el-icon>
                        编辑
                      </el-button>
                      <el-button size="small" type="success" @click.stop="applySingleSuggestion(result, sIndex)" :disabled="!suggestion.selected">
                        <el-icon><Check /></el-icon>
                        应用
                      </el-button>
                      <el-button size="small" type="danger" @click.stop="removeSuggestion(result, sIndex)">
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-button>
                    </div>
                  </div>
                </div>
              </div>

              <!-- 批量操作 -->
              <div class="batch-actions">
                <el-button @click="selectAllSuggestions(result)">全选</el-button>
                <el-button @click="deselectAllSuggestions(result)">全不选</el-button>
                <el-button type="success" @click="applySelectedSuggestions(result)" :disabled="!hasSelectedSuggestions(result)">
                  应用选中项 ({{ getSelectedCount(result) }})
                </el-button>
              </div>
            </div>
          </el-collapse-item>
        </el-collapse>
      </div>

      <!-- 生成进度对话框 -->
      <el-dialog v-model="progressDialogVisible" title="L4功能点生成进度" width="600px" :close-on-click-modal="false">
        <div class="progress-container">
          <el-progress
            :percentage="generateProgress"
            :status="progressStatus"
            :stroke-width="15"
          />
          <div class="progress-info">
            <p>{{ progressText }}</p>
            <div v-if="progressStatus === 'active'" class="progress-details">
              <div class="current-l3" v-if="currentProcessingL3">
                正在处理: {{ currentProcessingL3.title }}
              </div>
              <div class="progress-steps">
                <div v-for="step in progressSteps" :key="step.key" 
                     :class="['step-item', { active: step.active, completed: step.completed }]">
                  <el-icon v-if="step.completed"><Check /></el-icon>
                  <el-icon v-else-if="step.active"><Loading /></el-icon>
                  <span>{{ step.text }}</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </el-dialog>

      <!-- 历史记录对话框 -->
      <el-dialog v-model="historyDialogVisible" title="L4生成历史" width="80%">
        <el-table :data="historyList" max-height="400">
          <el-table-column prop="l3_title" label="L3需求" min-width="200" />
          <el-table-column prop="strategy" label="生成策略" width="120" />
          <el-table-column prop="l4_count" label="L4数量" width="100" />
          <el-table-column prop="applied_count" label="已应用" width="100" />
          <el-table-column prop="created_at" label="生成时间" width="180" :formatter="formatTableDate" />
          <el-table-column label="操作" width="200">
            <template #default="{ row }">
              <el-button size="small" @click="loadHistoryItem(row)">加载</el-button>
              <el-button size="small" type="info" @click="viewHistoryDetail(row)">详情</el-button>
              <el-button size="small" type="danger" @click="deleteHistoryItem(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-dialog>

      <!-- 统计数据对话框 -->
      <el-dialog v-model="statsDialogVisible" title="L4生成统计" width="70%">
        <div class="stats-container">
          <el-row :gutter="20">
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.total_generated }}</div>
                <div class="stat-label">总生成数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.total_applied }}</div>
                <div class="stat-label">已应用数</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ ((stats.total_applied / stats.total_generated) * 100).toFixed(1) }}%</div>
                <div class="stat-label">应用率</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="stat-card">
                <div class="stat-number">{{ stats.avg_confidence }}</div>
                <div class="stat-label">平均置信度</div>
              </div>
            </el-col>
          </el-row>
          
          <!-- 更多统计图表 -->
          <div class="stats-charts">
            <!-- TODO: 添加ECharts图表 -->
          </div>
        </div>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRouter } from 'vue-router'
import {
  Operation, Clock, DataAnalysis, Check, Download, Delete, Edit, Loading,
  InfoFilled
} from '@element-plus/icons-vue'

// 导入warning-bar组件
import WarningBar from '@/components/warningBar/warningBar.vue'

// API导入
import {
  getNesmaProjectList,
  getNesmaRequirementList,
  generateLevel4Requirements,
  batchCreateLevel4Requirements,
  createLevel4Requirement,
  validateLevel4Suggestion,
  getLevel4GenerationHistory,
  getLevel4GenerationStats
} from '@/api/nesma'

import { getProjectCycles } from '@/api/projectCycle'
import { getRequirementVersions } from '@/api/nesma/requirementVersion'

const router = useRouter()

// 响应式数据
const currentProject = ref(null)
const currentCycle = ref(null)
const currentVersion = ref(null)
const projects = ref([])
const cycles = ref([])
const versions = ref([])
const l3Requirements = ref([])

// 生成器配置
const generationStrategy = ref('comprehensive')
const complexityLevel = ref('moderate')
const maxL4Count = ref(8)
const includeKnowledgeBase = ref(true)
const selectedL3Requirements = ref([])

// 生成状态
const generating = ref(false)
const batchGenerating = ref(false)

// 生成结果
const generationResults = ref([])
const activeResultPanels = ref([])

// 进度监控
const progressDialogVisible = ref(false)
const generateProgress = ref(0)
const progressStatus = ref('active')
const progressText = ref('')
const currentProcessingL3 = ref(null)
const progressSteps = ref([
  { key: 'analyzing', text: '分析L3需求结构', active: false, completed: false },
  { key: 'generating', text: '生成L4功能点建议', active: false, completed: false },
  { key: 'validating', text: '验证生成结果', active: false, completed: false },
  { key: 'formatting', text: '格式化输出', active: false, completed: false }
])

// 历史和统计
const historyDialogVisible = ref(false)
const historyList = ref([])
const statsDialogVisible = ref(false)
const stats = ref({
  total_generated: 0,
  total_applied: 0,
  avg_confidence: 0
})

// 搜索表单
const searchForm = ref({})

// 计算属性
const canGenerate = computed(() => {
  return currentProject.value && currentCycle.value && currentVersion.value && selectedL3Requirements.value.length > 0
})

const totalL4Count = computed(() => {
  return generationResults.value.reduce((total, result) => total + result.l4Suggestions.length, 0)
})

const hasValidSuggestions = computed(() => {
  return generationResults.value.some(result => 
    result.l4Suggestions.some(suggestion => suggestion.selected)
  )
})

// 方法实现
const handleProjectChange = async () => {
  currentCycle.value = null
  currentVersion.value = null
  cycles.value = []
  versions.value = []
  l3Requirements.value = []
  
  if (currentProject.value) {
    await loadCycles()
  }
}

const handleCycleChange = async () => {
  currentVersion.value = null
  versions.value = []
  l3Requirements.value = []
  
  if (currentCycle.value) {
    await loadVersions()
  }
}

const handleVersionChange = async () => {
  if (currentVersion.value) {
    await loadL3Requirements()
  }
}

const handleL3Selection = (selection) => {
  selectedL3Requirements.value = selection
}

// 生成L4功能点
const generateL4Requirements = async () => {
  if (selectedL3Requirements.value.length === 0) {
    ElMessage.warning('请先选择L3需求')
    return
  }

  try {
    generating.value = true
    progressDialogVisible.value = true
    
    // 重置进度
    resetProgress()
    updateProgress(0, '开始分析L3需求...', 'analyzing')

    const generateData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      l3_requirement_ids: selectedL3Requirements.value.map(req => req.id || req.ID),
      generation_strategy: generationStrategy.value,
      complexity_level: complexityLevel.value,
      max_l4_count: maxL4Count.value,
      include_knowledge_base: includeKnowledgeBase.value,
      options: {
        auto_validate: true,
        generate_mermaid: false,
        calculate_estimates: true
      }
    }

    updateProgress(25, '正在生成L4功能点建议...', 'generating')
    
    const response = await generateLevel4Requirements(generateData)
    
    if (response.code === 0) {
      updateProgress(75, '验证生成结果...', 'validating')
      
      // 处理生成结果
      const results = response.data.results.map(result => ({
        l3Requirement: result.l3_requirement,
        strategy: result.strategy,
        l4Suggestions: result.l4_suggestions.map(suggestion => ({
          ...suggestion,
          selected: true // 默认选中
        })),
        createdAt: new Date()
      }))
      
      generationResults.value.push(...results)
      
      // 展开第一个结果面板
      if (results.length > 0) {
        activeResultPanels.value = [`result-${generationResults.value.length - results.length}`]
      }
      
      updateProgress(100, `成功生成 ${response.data.total_l4_count} 个L4功能点！`, 'formatting')
      
      setTimeout(() => {
        progressDialogVisible.value = false
        ElMessage.success(`L4功能点生成完成！共生成 ${response.data.total_l4_count} 个建议`)
      }, 1000)
      
    } else {
      throw new Error(response.msg || '生成失败')
    }

  } catch (error) {
    console.error('生成L4功能点失败:', error)
    ElMessage.error('生成失败: ' + (error.message || '未知错误'))
    progressDialogVisible.value = false
  } finally {
    generating.value = false
  }
}

// 批量生成
const batchGenerate = async () => {
  try {
    batchGenerating.value = true
    progressDialogVisible.value = true
    resetProgress()

    const batchData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      generation_strategy: generationStrategy.value,
      complexity_level: complexityLevel.value,
      max_l4_count: maxL4Count.value,
      include_knowledge_base: includeKnowledgeBase.value
    }

    updateProgress(20, '开始批量生成...', 'analyzing')
    
    const response = await batchCreateLevel4Requirements(batchData)
    
    if (response.code === 0) {
      updateProgress(80, '处理批量生成结果...', 'generating')
      
      // 处理批量结果
      const results = response.data.results.map(result => ({
        l3Requirement: result.l3_requirement,
        strategy: result.strategy,
        l4Suggestions: result.l4_suggestions.map(suggestion => ({
          ...suggestion,
          selected: true
        })),
        createdAt: new Date()
      }))
      
      generationResults.value.push(...results)
      
      updateProgress(100, `批量生成完成！共生成 ${response.data.total_l4_count} 个L4功能点`, 'formatting')
      
      setTimeout(() => {
        progressDialogVisible.value = false
        ElMessage.success(`批量生成完成！共处理 ${response.data.l3_count} 个L3需求`)
      }, 1000)
      
    } else {
      throw new Error(response.msg || '批量生成失败')
    }

  } catch (error) {
    console.error('批量生成失败:', error)
    ElMessage.error('批量生成失败: ' + (error.message || '未知错误'))
    progressDialogVisible.value = false
  } finally {
    batchGenerating.value = false
  }
}

// 建议操作方法
const toggleSuggestion = (result, index) => {
  result.l4Suggestions[index].selected = !result.l4Suggestions[index].selected
}

const updateSuggestionSelection = (result, index) => {
  // 复选框变更处理
}

const selectAllSuggestions = (result) => {
  result.l4Suggestions.forEach(suggestion => {
    suggestion.selected = true
  })
}

const deselectAllSuggestions = (result) => {
  result.l4Suggestions.forEach(suggestion => {
    suggestion.selected = false
  })
}

const hasSelectedSuggestions = (result) => {
  return result.l4Suggestions.some(suggestion => suggestion.selected)
}

const getSelectedCount = (result) => {
  return result.l4Suggestions.filter(suggestion => suggestion.selected).length
}

// 应用建议
const applySingleSuggestion = async (result, index) => {
  const suggestion = result.l4Suggestions[index]
  
  try {
    const createData = {
      parent_id: result.l3Requirement.id,
      suggestion: suggestion
    }

    const response = await createLevel4Requirement(createData)
    
    if (response.code === 0) {
      ElMessage.success('L4功能点已创建')
      suggestion.applied = true
    } else {
      throw new Error(response.msg || '创建失败')
    }

  } catch (error) {
    console.error('应用建议失败:', error)
    ElMessage.error('应用失败: ' + (error.message || '未知错误'))
  }
}

const applySelectedSuggestions = async (result) => {
  const selectedSuggestions = result.l4Suggestions.filter(s => s.selected)
  
  if (selectedSuggestions.length === 0) {
    ElMessage.warning('请先选择要应用的建议')
    return
  }

  try {
    const createData = {
      parent_id: result.l3Requirement.id,
      suggestions: selectedSuggestions
    }

    const response = await batchCreateLevel4Requirements(createData)
    
    if (response.code === 0) {
      ElMessage.success(`成功创建 ${selectedSuggestions.length} 个L4功能点`)
      selectedSuggestions.forEach(s => s.applied = true)
    } else {
      throw new Error(response.msg || '批量创建失败')
    }

  } catch (error) {
    console.error('批量应用失败:', error)
    ElMessage.error('批量应用失败: ' + (error.message || '未知错误'))
  }
}

const applyAllSuggestions = async () => {
  const allSelected = []
  
  generationResults.value.forEach(result => {
    result.l4Suggestions.forEach(suggestion => {
      if (suggestion.selected) {
        allSelected.push({
          parent_id: result.l3Requirement.id,
          suggestion: suggestion
        })
      }
    })
  })

  if (allSelected.length === 0) {
    ElMessage.warning('没有选中的建议')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要创建 ${allSelected.length} 个L4功能点吗？`,
      '批量应用确认',
      { type: 'warning' }
    )

    const response = await batchCreateLevel4Requirements({ suggestions: allSelected })
    
    if (response.code === 0) {
      ElMessage.success(`成功创建 ${allSelected.length} 个L4功能点`)
      
      // 标记为已应用
      generationResults.value.forEach(result => {
        result.l4Suggestions.forEach(suggestion => {
          if (suggestion.selected) {
            suggestion.applied = true
          }
        })
      })
    } else {
      throw new Error(response.msg || '批量应用失败')
    }

  } catch (error) {
    if (error === 'cancel') return
    console.error('批量应用失败:', error)
    ElMessage.error('批量应用失败: ' + (error.message || '未知错误'))
  }
}

// 编辑和删除
const editSuggestion = (result, index) => {
  // TODO: 实现编辑功能
  ElMessage.info('编辑功能开发中...')
}

const removeSuggestion = async (result, index) => {
  try {
    await ElMessageBox.confirm('确定要删除这个建议吗？', '确认删除', {
      type: 'warning'
    })
    
    result.l4Suggestions.splice(index, 1)
    ElMessage.success('建议已删除')
    
  } catch (error) {
    // 用户取消删除
  }
}

// 历史和统计
const showHistory = async () => {
  try {
    const response = await getLevel4GenerationHistory(currentCycle.value)
    
    if (response.code === 0) {
      historyList.value = response.data.history || []
      historyDialogVisible.value = true
    } else {
      throw new Error(response.msg || '获取历史记录失败')
    }

  } catch (error) {
    console.error('获取历史记录失败:', error)
    ElMessage.error('获取历史记录失败: ' + (error.message || '未知错误'))
  }
}

const showStats = async () => {
  try {
    const response = await getLevel4GenerationStats(currentCycle.value)
    
    if (response.code === 0) {
      stats.value = response.data.stats || {}
      statsDialogVisible.value = true
    } else {
      throw new Error(response.msg || '获取统计数据失败')
    }

  } catch (error) {
    console.error('获取统计数据失败:', error)
    ElMessage.error('获取统计数据失败: ' + (error.message || '未知错误'))
  }
}

// 其他操作
const exportResults = () => {
  // TODO: 实现导出功能
  ElMessage.info('导出功能开发中...')
}

const clearResults = async () => {
  try {
    await ElMessageBox.confirm('确定要清空所有生成结果吗？', '确认清空', {
      type: 'warning'
    })
    
    generationResults.value = []
    activeResultPanels.value = []
    ElMessage.success('结果已清空')
    
  } catch (error) {
    // 用户取消
  }
}

const viewExistingL4 = (l3Requirement) => {
  // TODO: 查看现有L4功能点
  ElMessage.info('查看现有L4功能点功能开发中...')
}

// 进度更新
const resetProgress = () => {
  generateProgress.value = 0
  progressStatus.value = 'active'
  currentProcessingL3.value = null
  progressSteps.value.forEach(step => {
    step.active = false
    step.completed = false
  })
}

const updateProgress = (percentage, text, activeStep) => {
  generateProgress.value = percentage
  progressText.value = text
  
  progressSteps.value.forEach(step => {
    if (step.key === activeStep) {
      step.active = true
    } else if (step.completed || generateProgress.value > 50) {
      step.completed = true
      step.active = false
    }
  })
}

// 辅助方法
const getComplexityType = (complexity) => {
  const complexityMap = {
    'simple': 'success',
    'moderate': 'warning',
    'complex': 'danger'
  }
  return complexityMap[complexity] || 'info'
}

const getFunctionTypeColor = (type) => {
  const typeMap = {
    'EI': 'primary',
    'EO': 'success', 
    'EQ': 'warning',
    'ILF': 'info',
    'EIF': 'danger'
  }
  return typeMap[type] || 'info'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString('zh-CN')
}

const formatTableDate = (row, column, cellValue) => {
  return formatDate(cellValue)
}

// 数据加载方法
const loadProjects = async () => {
  try {
    const response = await getNesmaProjectList({ 
      page: 1, 
      pageSize: 100 
    })
    
    if (response.code === 0) {
      projects.value = response.data.list || []
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
    ElMessage.error('加载项目列表失败')
  }
}

const loadCycles = async () => {
  if (!currentProject.value) return
  
  try {
    const response = await getProjectCycles(currentProject.value)
    
    if (response.code === 0) {
      cycles.value = response.data || []
    }
  } catch (error) {
    console.error('加载周期列表失败:', error)
    ElMessage.error('加载周期列表失败')
  }
}

const loadVersions = async () => {
  if (!currentCycle.value) return
  
  try {
    const response = await getRequirementVersions(currentCycle.value)
    
    if (response.code === 0) {
      versions.value = response.data || []
    }
  } catch (error) {
    console.error('加载版本列表失败:', error)
    ElMessage.error('加载版本列表失败')
  }
}

const loadL3Requirements = async () => {
  if (!currentVersion.value) return
  
  try {
    const params = {
      version_id: currentVersion.value,
      level: 3, // 只加载L3需求
      page: 1,
      pageSize: 1000
    }
    
    const response = await getNesmaRequirementList(params)
    
    if (response.code === 0) {
      l3Requirements.value = response.data.list || []
    }
  } catch (error) {
    console.error('加载L3需求失败:', error)
    ElMessage.error('加载L3需求失败')
  }
}

// 历史操作
const loadHistoryItem = (item) => {
  // TODO: 加载历史项目
  ElMessage.info('加载历史记录功能开发中...')
}

const viewHistoryDetail = (item) => {
  // TODO: 查看历史详情
  ElMessage.info('查看历史详情功能开发中...')
}

const deleteHistoryItem = async (item) => {
  try {
    await ElMessageBox.confirm('确定要删除这个历史记录吗？', '确认删除', {
      type: 'warning'
    })
    
    // TODO: 调用删除API
    ElMessage.success('删除成功')
    showHistory() // 重新加载
    
  } catch (error) {
    // 用户取消删除
  }
}

// 组件挂载
onMounted(() => {
  loadProjects()
})
</script>

<style scoped>
.level4-generator {
  padding: 20px;
}

.generator-config {
  background: #f8f9fa;
  padding: 20px;
  border-radius: 8px;
  margin: 20px 0;
}

.config-section h3 {
  margin: 0 0 15px 0;
  color: #333;
}

.config-row {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
  flex-wrap: wrap;
  gap: 12px;
}

.action-buttons {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.l3-requirement-selector {
  margin: 20px 0;
}

.l3-requirement-selector h3 {
  margin-bottom: 15px;
  color: #333;
}

.generation-results {
  margin: 20px 0;
}

.generation-results h3 {
  margin-bottom: 15px;
  color: #333;
}

.result-actions {
  margin-bottom: 20px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.l3-result-content {
  padding: 20px;
}

.l3-info {
  margin-bottom: 20px;
}

.l4-suggestions h4 {
  margin: 20px 0 15px 0;
  color: #333;
}

.suggestion-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
  gap: 16px;
  margin-bottom: 20px;
}

.suggestion-card {
  border: 1px solid #e1e5e9;
  border-radius: 8px;
  padding: 16px;
  cursor: pointer;
  transition: all 0.3s;
  background: white;
}

.suggestion-card:hover {
  border-color: #409eff;
  box-shadow: 0 2px 8px rgba(64, 158, 255, 0.1);
}

.suggestion-card.selected {
  border-color: #409eff;
  background: #f0f8ff;
}

.suggestion-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 12px;
}

.suggestion-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 500;
  flex: 1;
}

.suggestion-meta {
  display: flex;
  gap: 6px;
}

.suggestion-content {
  margin-bottom: 16px;
}

.suggestion-description {
  margin: 8px 0;
  color: #666;
  line-height: 1.5;
}

.suggestion-details {
  margin: 12px 0;
}

.detail-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  margin: 4px 0;
}

.detail-item .label {
  color: #999;
}

.detail-item .value {
  font-weight: 500;
  color: #333;
}

.suggestion-reason {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: #666;
  background: #f8f9fa;
  padding: 8px;
  border-radius: 4px;
  margin-top: 8px;
}

.suggestion-actions {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.batch-actions {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid #e1e5e9;
  display: flex;
  gap: 12px;
  align-items: center;
}

.progress-container {
  padding: 20px;
}

.progress-info {
  margin-top: 20px;
  text-align: center;
}

.current-l3 {
  margin: 12px 0;
  padding: 8px;
  background: #f0f8ff;
  border-radius: 4px;
  font-size: 14px;
  color: #333;
}

.progress-steps {
  margin-top: 20px;
}

.step-item {
  display: flex;
  align-items: center;
  margin: 8px 0;
  padding: 8px;
  border-radius: 4px;
  transition: all 0.3s;
}

.step-item.active {
  background: #e3f2fd;
  color: #1976d2;
}

.step-item.completed {
  background: #e8f5e8;
  color: #388e3c;
}

.step-item .el-icon {
  margin-right: 8px;
}

.stats-container {
  padding: 20px;
}

.stat-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #333;
  margin-bottom: 8px;
}

.stat-label {
  font-size: 12px;
  color: #666;
}

.stats-charts {
  margin-top: 30px;
  min-height: 300px;
}
</style>