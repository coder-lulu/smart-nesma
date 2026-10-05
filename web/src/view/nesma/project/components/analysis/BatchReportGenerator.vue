<template>
  <div class="batch-report-generator">
    <el-dialog
      v-model="visible"
      title="批量报告生成"
      width="900px"
      :close-on-click-modal="false"
      @close="handleClose"
    >
      <div class="batch-generator-content">
        <el-steps :active="currentStep" finish-status="success">
          <el-step title="选择项目" />
          <el-step title="配置报告" />
          <el-step title="生成报告" />
        </el-steps>
        
        <!-- 步骤1：选择项目 -->
        <div v-if="currentStep === 0" class="step-content">
          <div class="project-selection">
            <div class="selection-header">
              <h3>选择要生成报告的项目</h3>
              <div class="selection-actions">
                <el-button @click="selectAll">全选</el-button>
                <el-button @click="selectNone">清空</el-button>
                <el-button @click="selectByFilter">按条件选择</el-button>
              </div>
            </div>
            
            <div class="project-filters">
              <el-input
                v-model="projectSearch"
                placeholder="搜索项目..."
                clearable
                @input="handleProjectSearch"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
              
              <el-select v-model="statusFilter" placeholder="项目状态" clearable>
                <el-option label="激活" value="active" />
                <el-option label="已完成" value="completed" />
                <el-option label="已归档" value="archived" />
              </el-select>
              
              <el-select v-model="domainFilter" placeholder="业务领域" clearable>
                <el-option label="金融" value="finance" />
                <el-option label="制造" value="manufacturing" />
                <el-option label="医疗" value="healthcare" />
                <el-option label="政府" value="government" />
                <el-option label="电商" value="ecommerce" />
              </el-select>
            </div>
            
            <div class="project-list">
              <el-table
                ref="projectTable"
                :data="filteredProjects"
                @selection-change="handleProjectSelection"
                max-height="300"
              >
                <el-table-column type="selection" width="55" />
                <el-table-column prop="name" label="项目名称" min-width="200" />
                <el-table-column prop="domain" label="业务领域" width="120" />
                <el-table-column prop="status" label="状态" width="100">
                  <template #default="{ row }">
                    <el-tag :type="getStatusType(row.status)">
                      {{ getStatusLabel(row.status) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="cycle_count" label="周期数" width="80" />
                <el-table-column prop="requirement_count" label="需求数" width="80" />
                <el-table-column prop="last_analysis" label="最近分析" width="120">
                  <template #default="{ row }">
                    {{ formatDate(row.last_analysis) }}
                  </template>
                </el-table-column>
              </el-table>
            </div>
            
            <div class="selection-summary">
              <p>已选择 {{ selectedProjects.length }} 个项目</p>
            </div>
          </div>
        </div>
        
        <!-- 步骤2：配置报告 -->
        <div v-if="currentStep === 1" class="step-content">
          <div class="batch-config">
            <el-form :model="batchConfig" label-width="120px">
              <el-form-item label="报告类型">
                <el-radio-group v-model="batchConfig.type">
                  <el-radio label="analysis">分析报告</el-radio>
                  <el-radio label="summary">摘要报告</el-radio>
                  <el-radio label="detailed">详细报告</el-radio>
                </el-radio-group>
              </el-form-item>
              
              <el-form-item label="输出格式">
                <el-checkbox-group v-model="batchConfig.formats">
                  <el-checkbox label="pdf">PDF</el-checkbox>
                  <el-checkbox label="word">Word</el-checkbox>
                  <el-checkbox label="excel">Excel</el-checkbox>
                  <el-checkbox label="html">HTML</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
              
              <el-form-item label="报告模板">
                <el-select v-model="batchConfig.templateId" placeholder="选择模板">
                  <el-option 
                    v-for="template in reportTemplates" 
                    :key="template.id"
                    :label="template.name" 
                    :value="template.id" 
                  />
                </el-select>
              </el-form-item>
              
              <el-form-item label="包含内容">
                <el-transfer
                  v-model="batchConfig.sections"
                  :data="availableSections"
                  :titles="['可选内容', '包含内容']"
                  filterable
                />
              </el-form-item>
              
              <el-form-item label="文件命名">
                <el-input v-model="batchConfig.nameTemplate" placeholder="支持变量：{project_name}、{date}、{type}" />
                <div class="name-preview">
                  <span>示例：</span>
                  <span class="preview-text">{{ generateNamePreview() }}</span>
                </div>
              </el-form-item>
              
              <el-form-item label="输出目录">
                <el-input v-model="batchConfig.outputDir" placeholder="报告输出目录" />
              </el-form-item>
              
              <el-form-item label="并发设置">
                <el-slider
                  v-model="batchConfig.concurrency"
                  :min="1"
                  :max="5"
                  :marks="{ 1: '1', 3: '3', 5: '5' }"
                  show-stops
                />
                <div class="concurrency-tip">
                  <span>并发数：{{ batchConfig.concurrency }}（建议3个以下，避免系统负载过高）</span>
                </div>
              </el-form-item>
              
              <el-form-item label="高级选项">
                <el-checkbox-group v-model="batchConfig.options">
                  <el-checkbox label="includeCharts">包含图表</el-checkbox>
                  <el-checkbox label="includeMetadata">包含元数据</el-checkbox>
                  <el-checkbox label="compressOutput">压缩输出</el-checkbox>
                  <el-checkbox label="emailNotify">邮件通知</el-checkbox>
                </el-checkbox-group>
              </el-form-item>
            </el-form>
          </div>
        </div>
        
        <!-- 步骤3：生成报告 -->
        <div v-if="currentStep === 2" class="step-content">
          <div class="batch-progress">
            <div class="overall-progress">
              <h3>总体进度</h3>
              <el-progress
                :percentage="overallProgress"
                :stroke-width="15"
                :color="progressColor"
                show-text
              />
              <div class="progress-stats">
                <span>已完成：{{ completedCount }}/{{ totalCount }}</span>
                <span>成功：{{ successCount }}</span>
                <span>失败：{{ failedCount }}</span>
              </div>
            </div>
            
            <div class="project-progress-list">
              <h3>项目进度</h3>
              <div class="progress-items">
                <div 
                  v-for="item in progressItems" 
                  :key="item.projectId"
                  class="progress-item"
                >
                  <div class="item-header">
                    <span class="project-name">{{ item.projectName }}</span>
                    <span :class="['status-badge', item.status]">
                      {{ getProgressStatusLabel(item.status) }}
                    </span>
                  </div>
                  <el-progress
                    :percentage="item.progress"
                    :stroke-width="6"
                    :color="getProgressColor(item.status)"
                    :show-text="false"
                  />
                  <div class="item-details">
                    <span class="current-step">{{ item.currentStep }}</span>
                    <span class="elapsed-time">{{ item.elapsedTime }}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button v-if="currentStep > 0" @click="prevStep">上一步</el-button>
          <el-button @click="handleClose">取消</el-button>
          <el-button 
            v-if="currentStep < 2" 
            type="primary" 
            @click="nextStep"
            :disabled="!canProceed"
          >
            下一步
          </el-button>
          <el-button 
            v-if="currentStep === 2" 
            type="success" 
            @click="startBatchGeneration"
            :disabled="isGenerating"
          >
            {{ isGenerating ? '生成中...' : '开始生成' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Search } from '@element-plus/icons-vue'
import {
  getReportTemplates,
  startBatchReportGeneration,
  getBatchGenerationProgress
} from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  projects: {
    type: Array,
    default: () => []
  }
})

const emit = defineEmits(['update:modelValue', 'batch-started'])

// 响应式数据
const visible = ref(false)
const currentStep = ref(0)
const projectSearch = ref('')
const statusFilter = ref('')
const domainFilter = ref('')
const selectedProjects = ref([])
const reportTemplates = ref([])
const isGenerating = ref(false)
const batchId = ref(null)

// 批量配置
const batchConfig = reactive({
  type: 'analysis',
  formats: ['pdf'],
  templateId: null,
  sections: ['executive_summary', 'analysis_results'],
  nameTemplate: '{project_name}_{type}_report_{date}',
  outputDir: '/reports',
  concurrency: 3,
  options: ['includeCharts']
})

// 进度数据
const overallProgress = ref(0)
const completedCount = ref(0)
const totalCount = ref(0)
const successCount = ref(0)
const failedCount = ref(0)
const progressItems = ref([])

// 可用章节
const availableSections = ref([
  { key: 'executive_summary', label: '执行摘要' },
  { key: 'analysis_results', label: '分析结果' },
  { key: 'quality_assessment', label: '质量评估' },
  { key: 'technical_details', label: '技术细节' },
  { key: 'recommendations', label: '建议事项' }
])

// 计算属性
const filteredProjects = computed(() => {
  let filtered = props.projects

  if (projectSearch.value) {
    filtered = filtered.filter(project => 
      project.name.toLowerCase().includes(projectSearch.value.toLowerCase())
    )
  }

  if (statusFilter.value) {
    filtered = filtered.filter(project => project.status === statusFilter.value)
  }

  if (domainFilter.value) {
    filtered = filtered.filter(project => project.domain === domainFilter.value)
  }

  return filtered
})

const canProceed = computed(() => {
  if (currentStep.value === 0) {
    return selectedProjects.value.length > 0
  }
  if (currentStep.value === 1) {
    return batchConfig.formats.length > 0 && batchConfig.sections.length > 0
  }
  return true
})

const progressColor = computed(() => {
  if (overallProgress.value < 30) return '#f56c6c'
  if (overallProgress.value < 70) return '#e6a23c'
  return '#67c23a'
})

// 监听器
watch(() => props.modelValue, (val) => {
  visible.value = val
  if (val) {
    loadTemplates()
  }
})

watch(visible, (val) => {
  emit('update:modelValue', val)
})

// 方法
const loadTemplates = async () => {
  try {
    const response = await getReportTemplates()
    reportTemplates.value = response.data
  } catch (error) {
    ElMessage.error('加载模板失败: ' + error.message)
  }
}

const handleProjectSelection = (selection) => {
  selectedProjects.value = selection
}

const selectAll = () => {
  const table = document.querySelector('.project-list .el-table')
  if (table) {
    table.toggleAllSelection()
  }
}

const selectNone = () => {
  const table = document.querySelector('.project-list .el-table')
  if (table) {
    table.clearSelection()
  }
}

const selectByFilter = () => {
  // 实现按条件选择逻辑
  ElMessage.info('按条件选择功能开发中...')
}

const handleProjectSearch = () => {
  // 搜索已通过计算属性处理
}

const nextStep = () => {
  if (currentStep.value < 2) {
    currentStep.value++
  }
}

const prevStep = () => {
  if (currentStep.value > 0) {
    currentStep.value--
  }
}

const generateNamePreview = () => {
  const template = batchConfig.nameTemplate || '{project_name}_{type}_report_{date}'
  return template
    .replace('{project_name}', '示例项目')
    .replace('{type}', batchConfig.type)
    .replace('{date}', new Date().toISOString().split('T')[0])
}

const startBatchGeneration = async () => {
  try {
    isGenerating.value = true
    
    const generationData = {
      projects: selectedProjects.value.map(p => p.id),
      config: batchConfig
    }
    
    const response = await startBatchReportGeneration(generationData)
    batchId.value = response.data.batch_id
    
    // 初始化进度数据
    totalCount.value = selectedProjects.value.length
    completedCount.value = 0
    successCount.value = 0
    failedCount.value = 0
    overallProgress.value = 0
    
    progressItems.value = selectedProjects.value.map(project => ({
      projectId: project.id,
      projectName: project.name,
      status: 'pending',
      progress: 0,
      currentStep: '等待开始',
      elapsedTime: '00:00:00'
    }))
    
    // 开始轮询进度
    pollBatchProgress()
    
    emit('batch-started', {
      batchId: batchId.value,
      total: totalCount.value,
      config: batchConfig
    })
    
  } catch (error) {
    ElMessage.error('启动批量生成失败: ' + error.message)
    isGenerating.value = false
  }
}

const pollBatchProgress = async () => {
  if (!batchId.value) return
  
  const interval = setInterval(async () => {
    try {
      const response = await getBatchGenerationProgress(batchId.value)
      const progress = response.data
      
      overallProgress.value = progress.overall_progress
      completedCount.value = progress.completed_count
      successCount.value = progress.success_count
      failedCount.value = progress.failed_count
      
      // 更新项目进度
      if (progress.project_progress) {
        progressItems.value = progressItems.value.map(item => {
          const projectProgress = progress.project_progress[item.projectId]
          if (projectProgress) {
            return {
              ...item,
              status: projectProgress.status,
              progress: projectProgress.progress,
              currentStep: projectProgress.current_step,
              elapsedTime: projectProgress.elapsed_time
            }
          }
          return item
        })
      }
      
      if (progress.status === 'completed') {
        clearInterval(interval)
        isGenerating.value = false
        ElMessage.success('批量生成完成!')
      } else if (progress.status === 'failed') {
        clearInterval(interval)
        isGenerating.value = false
        ElMessage.error('批量生成失败')
      }
      
    } catch (error) {
      clearInterval(interval)
      isGenerating.value = false
      ElMessage.error('获取生成进度失败')
    }
  }, 3000)
}

const getStatusType = (status) => {
  const types = {
    active: 'success',
    completed: 'info',
    archived: 'warning'
  }
  return types[status] || 'info'
}

const getStatusLabel = (status) => {
  const labels = {
    active: '激活',
    completed: '已完成',
    archived: '已归档'
  }
  return labels[status] || status
}

const getProgressStatusLabel = (status) => {
  const labels = {
    pending: '等待',
    running: '进行中',
    completed: '完成',
    failed: '失败'
  }
  return labels[status] || status
}

const getProgressColor = (status) => {
  const colors = {
    pending: '#909399',
    running: '#409eff',
    completed: '#67c23a',
    failed: '#f56c6c'
  }
  return colors[status] || '#909399'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString()
}

const handleClose = () => {
  if (isGenerating.value) {
    ElMessageBox.confirm('生成正在进行中，确定要关闭吗？', '确认关闭', {
      type: 'warning'
    }).then(() => {
      visible.value = false
      resetState()
    }).catch(() => {
      // 取消关闭
    })
  } else {
    visible.value = false
    resetState()
  }
}

const resetState = () => {
  currentStep.value = 0
  selectedProjects.value = []
  isGenerating.value = false
  batchId.value = null
  overallProgress.value = 0
  completedCount.value = 0
  totalCount.value = 0
  successCount.value = 0
  failedCount.value = 0
  progressItems.value = []
}
</script>

<style lang="scss" scoped>
.batch-report-generator {
  .batch-generator-content {
    padding: 20px 0;
    
    .el-steps {
      margin-bottom: 30px;
    }
    
    .step-content {
      min-height: 400px;
      
      .project-selection {
        .selection-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          
          h3 {
            margin: 0;
            color: #303133;
          }
          
          .selection-actions {
            display: flex;
            gap: 8px;
          }
        }
        
        .project-filters {
          display: flex;
          gap: 12px;
          margin-bottom: 16px;
          
          .el-input {
            flex: 2;
          }
          
          .el-select {
            flex: 1;
          }
        }
        
        .project-list {
          margin-bottom: 16px;
        }
        
        .selection-summary {
          text-align: center;
          padding: 12px;
          background: #f5f7fa;
          border-radius: 4px;
          
          p {
            margin: 0;
            color: #606266;
            font-size: 14px;
          }
        }
      }
      
      .batch-config {
        .name-preview {
          margin-top: 8px;
          font-size: 12px;
          color: #909399;
          
          .preview-text {
            color: #409eff;
            font-weight: 500;
          }
        }
        
        .concurrency-tip {
          margin-top: 8px;
          font-size: 12px;
          color: #909399;
        }
      }
      
      .batch-progress {
        .overall-progress {
          margin-bottom: 30px;
          
          h3 {
            margin: 0 0 16px 0;
            color: #303133;
          }
          
          .progress-stats {
            display: flex;
            justify-content: center;
            gap: 24px;
            margin-top: 12px;
            font-size: 14px;
            color: #606266;
          }
        }
        
        .project-progress-list {
          h3 {
            margin: 0 0 16px 0;
            color: #303133;
          }
          
          .progress-items {
            max-height: 300px;
            overflow-y: auto;
            
            .progress-item {
              padding: 12px;
              border: 1px solid #e4e7ed;
              border-radius: 6px;
              margin-bottom: 8px;
              
              .item-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 8px;
                
                .project-name {
                  font-weight: 500;
                  color: #303133;
                }
                
                .status-badge {
                  padding: 2px 8px;
                  border-radius: 3px;
                  font-size: 12px;
                  
                  &.pending {
                    background: #f0f0f0;
                    color: #909399;
                  }
                  
                  &.running {
                    background: #ecf5ff;
                    color: #409eff;
                  }
                  
                  &.completed {
                    background: #f0f9ff;
                    color: #67c23a;
                  }
                  
                  &.failed {
                    background: #fef0f0;
                    color: #f56c6c;
                  }
                }
              }
              
              .item-details {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-top: 8px;
                font-size: 12px;
                color: #909399;
              }
            }
          }
        }
      }
    }
  }
  
  .dialog-footer {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
  }
}
</style>