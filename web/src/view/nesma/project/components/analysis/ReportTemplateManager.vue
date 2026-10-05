<template>
  <div class="report-template-manager">
    <el-dialog
      v-model="visible"
      title="报告模板管理"
      width="1200px"
      :close-on-click-modal="false"
      @close="handleClose"
    >
      <div class="template-manager-content">
        <!-- 左侧：模板列表 -->
        <div class="template-list-section">
          <div class="section-header">
            <h3>模板列表</h3>
            <el-button type="primary" @click="createTemplate">
              <el-icon><Plus /></el-icon>
              新建模板
            </el-button>
          </div>
          
          <div class="template-search">
            <el-input
              v-model="searchKeyword"
              placeholder="搜索模板..."
              clearable
              @input="handleSearch"
            >
              <template #prefix>
                <el-icon><Search /></el-icon>
              </template>
            </el-input>
          </div>
          
          <div class="template-filters">
            <el-select v-model="filterType" placeholder="类型筛选" clearable>
              <el-option label="分析报告" value="analysis" />
              <el-option label="摘要报告" value="summary" />
              <el-option label="详细报告" value="detailed" />
              <el-option label="自定义报告" value="custom" />
            </el-select>
            
            <el-select v-model="filterFormat" placeholder="格式筛选" clearable>
              <el-option label="PDF" value="pdf" />
              <el-option label="Word" value="word" />
              <el-option label="Excel" value="excel" />
              <el-option label="HTML" value="html" />
            </el-select>
          </div>
          
          <div class="template-items">
            <div
              v-for="template in filteredTemplates"
              :key="template.id"
              :class="['template-item', { active: selectedTemplate?.id === template.id }]"
              @click="selectTemplate(template)"
            >
              <div class="template-icon">
                <el-icon><Document /></el-icon>
              </div>
              <div class="template-info">
                <h4>{{ template.name }}</h4>
                <p class="template-desc">{{ template.description }}</p>
                <div class="template-meta">
                  <span class="template-type">{{ getTypeLabel(template.type) }}</span>
                  <span class="template-format">{{ template.format.toUpperCase() }}</span>
                  <span class="template-status" :class="template.status">
                    {{ template.status === 'active' ? '启用' : '停用' }}
                  </span>
                </div>
              </div>
              <div class="template-actions">
                <el-button type="text" @click.stop="editTemplate(template)">
                  <el-icon><Edit /></el-icon>
                </el-button>
                <el-button type="text" @click.stop="duplicateTemplate(template)">
                  <el-icon><CopyDocument /></el-icon>
                </el-button>
                <el-button type="text" @click.stop="deleteTemplate(template)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 右侧：模板编辑器 -->
        <div class="template-editor-section">
          <div v-if="selectedTemplate" class="template-editor">
            <div class="editor-header">
              <h3>{{ isEditing ? '编辑模板' : '模板详情' }}</h3>
              <div class="editor-actions">
                <el-button v-if="!isEditing" @click="enableEditing">编辑</el-button>
                <el-button v-if="isEditing" @click="cancelEditing">取消</el-button>
                <el-button v-if="isEditing" type="primary" @click="saveTemplate">保存</el-button>
              </div>
            </div>
            
            <div class="editor-content">
              <el-form
                ref="templateForm"
                :model="editingTemplate"
                :rules="templateRules"
                label-width="100px"
                :disabled="!isEditing"
              >
                <el-form-item label="模板名称" prop="name">
                  <el-input v-model="editingTemplate.name" placeholder="输入模板名称" />
                </el-form-item>
                
                <el-form-item label="模板描述" prop="description">
                  <el-input
                    v-model="editingTemplate.description"
                    type="textarea"
                    :rows="3"
                    placeholder="输入模板描述"
                  />
                </el-form-item>
                
                <el-form-item label="模板类型" prop="type">
                  <el-select v-model="editingTemplate.type" placeholder="选择模板类型">
                    <el-option label="分析报告" value="analysis" />
                    <el-option label="摘要报告" value="summary" />
                    <el-option label="详细报告" value="detailed" />
                    <el-option label="自定义报告" value="custom" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="输出格式" prop="format">
                  <el-select v-model="editingTemplate.format" placeholder="选择输出格式">
                    <el-option label="PDF" value="pdf" />
                    <el-option label="Word" value="word" />
                    <el-option label="Excel" value="excel" />
                    <el-option label="HTML" value="html" />
                  </el-select>
                </el-form-item>
                
                <el-form-item label="模板状态" prop="status">
                  <el-switch
                    v-model="editingTemplate.status"
                    active-value="active"
                    inactive-value="inactive"
                    active-text="启用"
                    inactive-text="停用"
                  />
                </el-form-item>
                
                <el-form-item label="默认章节" prop="defaultSections">
                  <el-transfer
                    v-model="editingTemplate.defaultSections"
                    :data="availableSections"
                    :titles="['可用章节', '已选章节']"
                    filterable
                    :filter-method="filterSections"
                  />
                </el-form-item>
                
                <el-form-item label="模板配置">
                  <el-tabs v-model="configTab" type="border-card">
                    <el-tab-pane label="样式配置" name="style">
                      <StyleConfig v-model="editingTemplate.styleConfig" :disabled="!isEditing" />
                    </el-tab-pane>
                    <el-tab-pane label="布局配置" name="layout">
                      <LayoutConfig v-model="editingTemplate.layoutConfig" :disabled="!isEditing" />
                    </el-tab-pane>
                    <el-tab-pane label="数据配置" name="data">
                      <DataConfig v-model="editingTemplate.dataConfig" :disabled="!isEditing" />
                    </el-tab-pane>
                  </el-tabs>
                </el-form-item>
              </el-form>
            </div>
          </div>
          
          <div v-else class="no-template-selected">
            <el-empty description="选择一个模板进行编辑" />
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Search,
  Document,
  Edit,
  CopyDocument,
  Delete
} from '@element-plus/icons-vue'
import StyleConfig from './StyleConfig.vue'
import LayoutConfig from './LayoutConfig.vue'
import DataConfig from './DataConfig.vue'
import {
  getReportTemplates,
  createReportTemplate,
  updateReportTemplate,
  deleteReportTemplate,
  duplicateReportTemplate
} from '@/api/nesma'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue', 'template-updated'])

// 响应式数据
const visible = ref(false)
const searchKeyword = ref('')
const filterType = ref('')
const filterFormat = ref('')
const templates = ref([])
const selectedTemplate = ref(null)
const isEditing = ref(false)
const editingTemplate = ref({})
const configTab = ref('style')
const templateForm = ref(null)

// 可用章节数据
const availableSections = ref([
  { key: 'executive_summary', label: '执行摘要' },
  { key: 'key_findings', label: '关键发现' },
  { key: 'recommendations', label: '建议事项' },
  { key: 'analysis_results', label: '分析结果' },
  { key: 'requirement_analysis', label: '需求分析' },
  { key: 'function_points', label: '功能点计算' },
  { key: 'complexity_analysis', label: '复杂度分析' },
  { key: 'quality_assessment', label: '质量评估' },
  { key: 'quality_metrics', label: '质量指标' },
  { key: 'risk_analysis', label: '风险分析' },
  { key: 'technical_details', label: '技术细节' },
  { key: 'architecture', label: '架构设计' },
  { key: 'implementation', label: '实现方案' },
  { key: 'appendices', label: '附录' },
  { key: 'raw_data', label: '原始数据' },
  { key: 'references', label: '参考文献' }
])

// 表单验证规则
const templateRules = reactive({
  name: [
    { required: true, message: '请输入模板名称', trigger: 'blur' },
    { min: 2, max: 50, message: '长度在 2 到 50 个字符', trigger: 'blur' }
  ],
  description: [
    { required: true, message: '请输入模板描述', trigger: 'blur' },
    { max: 200, message: '长度不能超过 200 个字符', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择模板类型', trigger: 'change' }
  ],
  format: [
    { required: true, message: '请选择输出格式', trigger: 'change' }
  ]
})

// 计算属性
const filteredTemplates = computed(() => {
  let filtered = templates.value

  if (searchKeyword.value) {
    filtered = filtered.filter(template => 
      template.name.toLowerCase().includes(searchKeyword.value.toLowerCase()) ||
      template.description.toLowerCase().includes(searchKeyword.value.toLowerCase())
    )
  }

  if (filterType.value) {
    filtered = filtered.filter(template => template.type === filterType.value)
  }

  if (filterFormat.value) {
    filtered = filtered.filter(template => template.format === filterFormat.value)
  }

  return filtered
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
    templates.value = response.data
  } catch (error) {
    ElMessage.error('加载模板失败: ' + error.message)
  }
}

const selectTemplate = (template) => {
  selectedTemplate.value = template
  editingTemplate.value = { ...template }
  isEditing.value = false
}

const createTemplate = () => {
  selectedTemplate.value = null
  editingTemplate.value = {
    name: '',
    description: '',
    type: 'analysis',
    format: 'pdf',
    status: 'active',
    defaultSections: ['executive_summary', 'analysis_results'],
    styleConfig: {
      primaryColor: '#409EFF',
      fontSize: 14,
      fontFamily: 'PingFang SC',
      lineHeight: 1.6
    },
    layoutConfig: {
      orientation: 'portrait',
      margin: { top: 20, right: 20, bottom: 20, left: 20 },
      headerHeight: 60,
      footerHeight: 40
    },
    dataConfig: {
      includeCharts: true,
      includeImages: true,
      dataFormat: 'table'
    }
  }
  isEditing.value = true
}

const editTemplate = (template) => {
  selectTemplate(template)
  enableEditing()
}

const enableEditing = () => {
  isEditing.value = true
}

const cancelEditing = () => {
  if (selectedTemplate.value) {
    editingTemplate.value = { ...selectedTemplate.value }
    isEditing.value = false
  } else {
    selectedTemplate.value = null
    editingTemplate.value = {}
    isEditing.value = false
  }
}

const saveTemplate = async () => {
  try {
    await templateForm.value.validate()
    
    if (selectedTemplate.value) {
      // 更新现有模板
      await updateReportTemplate(editingTemplate.value)
      ElMessage.success('模板更新成功')
    } else {
      // 创建新模板
      await createReportTemplate(editingTemplate.value)
      ElMessage.success('模板创建成功')
    }
    
    isEditing.value = false
    await loadTemplates()
    emit('template-updated')
    
  } catch (error) {
    if (error.fields) {
      // 表单验证错误
      return
    }
    ElMessage.error('保存模板失败: ' + error.message)
  }
}

const duplicateTemplate = async (template) => {
  try {
    await duplicateReportTemplate(template.id)
    ElMessage.success('模板复制成功')
    await loadTemplates()
    emit('template-updated')
  } catch (error) {
    ElMessage.error('复制模板失败: ' + error.message)
  }
}

const deleteTemplate = async (template) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除模板 "${template.name}" 吗？`,
      '确认删除',
      { type: 'warning' }
    )
    
    await deleteReportTemplate(template.id)
    ElMessage.success('模板删除成功')
    
    if (selectedTemplate.value?.id === template.id) {
      selectedTemplate.value = null
      editingTemplate.value = {}
      isEditing.value = false
    }
    
    await loadTemplates()
    emit('template-updated')
    
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除模板失败: ' + error.message)
    }
  }
}

const handleSearch = () => {
  // 搜索已通过计算属性处理
}

const filterSections = (query, item) => {
  return item.label.toLowerCase().includes(query.toLowerCase())
}

const getTypeLabel = (type) => {
  const labels = {
    analysis: '分析报告',
    summary: '摘要报告',
    detailed: '详细报告',
    custom: '自定义报告'
  }
  return labels[type] || type
}

const handleClose = () => {
  visible.value = false
  selectedTemplate.value = null
  editingTemplate.value = {}
  isEditing.value = false
}
</script>

<style lang="scss" scoped>
.report-template-manager {
  .template-manager-content {
    display: flex;
    gap: 20px;
    height: 700px;
    
    .template-list-section {
      flex: 0 0 400px;
      display: flex;
      flex-direction: column;
      
      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
        
        h3 {
          margin: 0;
          color: #303133;
          font-size: 16px;
        }
      }
      
      .template-search {
        margin-bottom: 12px;
      }
      
      .template-filters {
        display: flex;
        gap: 12px;
        margin-bottom: 16px;
        
        .el-select {
          flex: 1;
        }
      }
      
      .template-items {
        flex: 1;
        overflow-y: auto;
        padding-right: 8px;
        
        .template-item {
          display: flex;
          align-items: flex-start;
          gap: 12px;
          padding: 12px;
          border: 1px solid #e4e7ed;
          border-radius: 6px;
          margin-bottom: 8px;
          cursor: pointer;
          transition: all 0.2s;
          
          &:hover {
            border-color: #409eff;
            background-color: #f0f9ff;
          }
          
          &.active {
            border-color: #409eff;
            background-color: #ecf5ff;
          }
          
          .template-icon {
            .el-icon {
              font-size: 24px;
              color: #909399;
            }
          }
          
          .template-info {
            flex: 1;
            
            h4 {
              margin: 0 0 4px 0;
              font-size: 14px;
              color: #303133;
            }
            
            .template-desc {
              margin: 0 0 8px 0;
              font-size: 12px;
              color: #606266;
              line-height: 1.4;
            }
            
            .template-meta {
              display: flex;
              gap: 8px;
              
              span {
                font-size: 11px;
                padding: 2px 6px;
                border-radius: 3px;
                background: #f5f7fa;
                color: #606266;
              }
              
              .template-status {
                &.active {
                  background: #f0f9ff;
                  color: #409eff;
                }
                
                &.inactive {
                  background: #fef0f0;
                  color: #f56c6c;
                }
              }
            }
          }
          
          .template-actions {
            display: flex;
            flex-direction: column;
            gap: 4px;
            opacity: 0;
            transition: opacity 0.2s;
            
            .el-button {
              padding: 4px;
              
              .el-icon {
                font-size: 14px;
              }
            }
          }
          
          &:hover .template-actions {
            opacity: 1;
          }
        }
      }
    }
    
    .template-editor-section {
      flex: 1;
      display: flex;
      flex-direction: column;
      
      .template-editor {
        height: 100%;
        display: flex;
        flex-direction: column;
        
        .editor-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          padding-bottom: 12px;
          border-bottom: 1px solid #e4e7ed;
          
          h3 {
            margin: 0;
            color: #303133;
            font-size: 16px;
          }
          
          .editor-actions {
            display: flex;
            gap: 8px;
          }
        }
        
        .editor-content {
          flex: 1;
          overflow-y: auto;
          
          .el-form {
            .el-form-item {
              margin-bottom: 20px;
            }
            
            .el-transfer {
              .el-transfer-panel {
                width: 200px;
              }
            }
          }
        }
      }
      
      .no-template-selected {
        display: flex;
        align-items: center;
        justify-content: center;
        height: 100%;
        
        .el-empty {
          color: #909399;
        }
      }
    }
  }
}
</style>