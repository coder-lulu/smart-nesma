<template>
  <div class="template-preview">
    <div v-loading="loading" class="preview-container">
      <div v-if="templateInfo" class="template-info">
        <div class="info-header">
          <div class="template-title">
            <el-icon class="template-icon">
              <Document v-if="templateInfo.type === 'word'" />
              <Grid v-else-if="templateInfo.type === 'excel'" />
              <DocumentCopy v-else />
            </el-icon>
            <span>{{ templateInfo.name }}</span>
          </div>
          <div class="template-meta">
            <el-tag size="small" :type="getTypeTagType(templateInfo.type)">
              {{ getTypeLabel(templateInfo.type) }}
            </el-tag>
            <el-tag size="small" type="info">
              {{ getFormatLabel(templateInfo.format) }}
            </el-tag>
            <el-tag v-if="templateInfo.isDefault" size="small" type="success">
              默认模板
            </el-tag>
            <span class="meta-item">{{ templateInfo.version }}</span>
          </div>
        </div>
        <div class="template-description">
          {{ templateInfo.description || '暂无描述' }}
        </div>
      </div>

      <div v-if="templateInfo" class="preview-content">
        <el-tabs v-model="activeTab" type="card">
          <el-tab-pane label="模板内容" name="content">
            <div class="content-section">
              <div class="content-header">
                <span>模板内容预览</span>
                <div class="content-actions">
                  <el-button size="small" @click="handleCopyContent" icon="CopyDocument">
                    复制内容
                  </el-button>
                  <el-button size="small" @click="handleDownloadTemplate" icon="Download">
                    下载模板
                  </el-button>
                </div>
              </div>
              <div class="content-viewer">
                <pre class="template-content">{{ templateInfo.content }}</pre>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="变量配置" name="variables">
            <div class="variables-section">
              <div v-if="!templateInfo.variables || templateInfo.variables.length === 0" class="empty-state">
                <el-empty description="该模板暂无自定义变量" :image-size="80" />
              </div>
              <div v-else class="variables-list">
                <div class="variables-header">
                  <span>模板变量列表 ({{ templateInfo.variables.length }})</span>
                </div>
                <el-table
                  :data="templateInfo.variables"
                  style="width: 100%"
                  border
                >
                  <el-table-column prop="name" label="变量名称" width="150" />
                  <el-table-column prop="label" label="显示标签" width="150" />
                  <el-table-column prop="type" label="数据类型" width="100">
                    <template #default="scope">
                      <el-tag size="small" :type="getVariableTypeTag(scope.row.type)">
                        {{ getVariableTypeLabel(scope.row.type) }}
                      </el-tag>
                    </template>
                  </el-table-column>
                  <el-table-column prop="defaultValue" label="默认值" width="150" show-overflow-tooltip>
                    <template #default="scope">
                      <span v-if="scope.row.defaultValue">{{ scope.row.defaultValue }}</span>
                      <span v-else class="empty-value">-</span>
                    </template>
                  </el-table-column>
                  <el-table-column prop="placeholder" label="提示文本" min-width="150" show-overflow-tooltip>
                    <template #default="scope">
                      <span v-if="scope.row.placeholder">{{ scope.row.placeholder }}</span>
                      <span v-else class="empty-value">-</span>
                    </template>
                  </el-table-column>
                  <el-table-column label="选项配置" width="120">
                    <template #default="scope">
                      <div v-if="scope.row.type === 'select' && scope.row.options">
                        <el-popover
                          placement="left"
                          width="300"
                          trigger="hover"
                        >
                          <template #reference>
                            <el-button size="small" type="primary" text>
                              查看选项 ({{ scope.row.options.length }})
                            </el-button>
                          </template>
                          <div class="options-preview">
                            <div class="options-title">选项列表：</div>
                            <div class="options-list">
                              <div 
                                v-for="(option, index) in scope.row.options" 
                                :key="index"
                                class="option-item"
                              >
                                <span class="option-label">{{ option.label }}</span>
                                <span class="option-value">({{ option.value }})</span>
                              </div>
                            </div>
                          </div>
                        </el-popover>
                      </div>
                      <span v-else class="empty-value">-</span>
                    </template>
                  </el-table-column>
                </el-table>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="使用统计" name="statistics">
            <div class="statistics-section">
              <el-row :gutter="20">
                <el-col :span="8">
                  <el-card class="stat-card">
                    <div class="stat-content">
                      <div class="stat-icon">
                        <el-icon><Document /></el-icon>
                      </div>
                      <div class="stat-info">
                        <div class="stat-number">{{ templateInfo.usageCount || 0 }}</div>
                        <div class="stat-label">使用次数</div>
                      </div>
                    </div>
                  </el-card>
                </el-col>
                <el-col :span="8">
                  <el-card class="stat-card">
                    <div class="stat-content">
                      <div class="stat-icon success">
                        <el-icon><Check /></el-icon>
                      </div>
                      <div class="stat-info">
                        <div class="stat-number">{{ templateInfo.successCount || 0 }}</div>
                        <div class="stat-label">成功生成</div>
                      </div>
                    </div>
                  </el-card>
                </el-col>
                <el-col :span="8">
                  <el-card class="stat-card">
                    <div class="stat-content">
                      <div class="stat-icon info">
                        <el-icon><Calendar /></el-icon>
                      </div>
                      <div class="stat-info">
                        <div class="stat-number">{{ formatDate(templateInfo.lastUsedAt) }}</div>
                        <div class="stat-label">最后使用</div>
                      </div>
                    </div>
                  </el-card>
                </el-col>
              </el-row>

              <div class="usage-chart" style="margin-top: 20px;">
                <el-card>
                  <template #header>
                    <span>使用趋势</span>
                  </template>
                  <div class="chart-placeholder">
                    <el-empty description="暂无使用数据" :image-size="80" />
                  </div>
                </el-card>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane label="模板信息" name="info">
            <div class="info-section">
              <el-descriptions title="模板详细信息" :column="2" border>
                <el-descriptions-item label="模板ID">
                  {{ templateInfo.ID }}
                </el-descriptions-item>
                <el-descriptions-item label="模板名称">
                  {{ templateInfo.name }}
                </el-descriptions-item>
                <el-descriptions-item label="模板类型">
                  <el-tag :type="getTypeTagType(templateInfo.type)">
                    {{ getTypeLabel(templateInfo.type) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="文档格式">
                  <el-tag type="info">
                    {{ getFormatLabel(templateInfo.format) }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="模板版本">
                  {{ templateInfo.version }}
                </el-descriptions-item>
                <el-descriptions-item label="模板状态">
                  <el-tag :type="templateInfo.status === 'active' ? 'success' : 'info'">
                    {{ templateInfo.status === 'active' ? '启用' : '停用' }}
                  </el-tag>
                </el-descriptions-item>
                <el-descriptions-item label="是否默认">
                  <el-tag v-if="templateInfo.isDefault" type="success">是</el-tag>
                  <span v-else>否</span>
                </el-descriptions-item>
                <el-descriptions-item label="创建时间">
                  {{ formatDate(templateInfo.createdAt) }}
                </el-descriptions-item>
                <el-descriptions-item label="更新时间">
                  {{ formatDate(templateInfo.updatedAt) }}
                </el-descriptions-item>
                <el-descriptions-item label="使用次数">
                  {{ templateInfo.usageCount || 0 }}
                </el-descriptions-item>
                <el-descriptions-item label="模板描述" :span="2">
                  {{ templateInfo.description || '暂无描述' }}
                </el-descriptions-item>
              </el-descriptions>
            </div>
          </el-tab-pane>
        </el-tabs>
      </div>

      <div v-if="error" class="error-state">
        <el-result
          icon="error"
          title="加载失败"
          :sub-title="error"
        >
          <template #extra>
            <el-button @click="loadTemplate" type="primary">重试</el-button>
          </template>
        </el-result>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Document,
  Grid,
  DocumentCopy,
  CopyDocument,
  Download,
  Check,
  Calendar
} from '@element-plus/icons-vue'
import {
  getTemplate
} from '@/api/nesma'

// Props
const props = defineProps({
  templateId: {
    type: [String, Number],
    required: true
  }
})

// 响应式数据
const loading = ref(false)
const templateInfo = ref(null)
const error = ref('')
const activeTab = ref('content')

// 加载模板信息
const loadTemplate = async () => {
  if (!props.templateId) return

  loading.value = true
  error.value = ''
  
  try {
    const res = await getTemplate(props.templateId)
    if (res.code === 0) {
      templateInfo.value = res.data
    } else {
      error.value = res.msg || '获取模板信息失败'
    }
  } catch (err) {
    console.error('加载模板失败:', err)
    error.value = '加载模板失败'
  } finally {
    loading.value = false
  }
}

// 监听templateId变化
watch(() => props.templateId, (newId) => {
  if (newId) {
    loadTemplate()
  }
}, { immediate: true })

// 复制内容
const handleCopyContent = async () => {
  if (!templateInfo.value?.content) {
    ElMessage.warning('模板内容为空')
    return
  }

  try {
    await navigator.clipboard.writeText(templateInfo.value.content)
    ElMessage.success('内容已复制到剪贴板')
  } catch (err) {
    // 降级方案
    const textArea = document.createElement('textarea')
    textArea.value = templateInfo.value.content
    document.body.appendChild(textArea)
    textArea.select()
    document.execCommand('copy')
    document.body.removeChild(textArea)
    ElMessage.success('内容已复制到剪贴板')
  }
}

// 下载模板
const handleDownloadTemplate = () => {
  if (!templateInfo.value) return

  const content = templateInfo.value.content
  const blob = new Blob([content], { type: 'text/plain;charset=utf-8' })
  const url = window.URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${templateInfo.value.name}_${templateInfo.value.version}.txt`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  window.URL.revokeObjectURL(url)
  
  ElMessage.success('模板下载成功')
}

// 工具函数
const getTypeLabel = (type) => {
  const labels = {
    word: 'Word',
    excel: 'Excel',
    pdf: 'PDF'
  }
  return labels[type] || type
}

const getTypeTagType = (type) => {
  const types = {
    word: 'primary',
    excel: 'success',
    pdf: 'warning'
  }
  return types[type] || ''
}

const getFormatLabel = (format) => {
  const labels = {
    requirement_spec: '需求规格说明书',
    nesma_report: 'NESMA评估报告',
    business_summary: '业务需求汇总表'
  }
  return labels[format] || format
}

const getVariableTypeLabel = (type) => {
  const labels = {
    string: '文本',
    number: '数字',
    boolean: '布尔',
    select: '选择'
  }
  return labels[type] || type
}

const getVariableTypeTag = (type) => {
  const types = {
    string: '',
    number: 'success',
    boolean: 'warning',
    select: 'info'
  }
  return types[type] || ''
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleString()
}

// 生命周期
onMounted(() => {
  if (props.templateId) {
    loadTemplate()
  }
})
</script>

<style scoped>
.template-preview {
  height: 70vh;
}

.preview-container {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.template-info {
  padding: 16px;
  background: var(--el-bg-color-page);
  border-bottom: 1px solid var(--el-border-color);
}

.info-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.template-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 18px;
  font-weight: 500;
}

.template-icon {
  font-size: 20px;
  color: var(--el-color-primary);
}

.template-meta {
  display: flex;
  align-items: center;
  gap: 8px;
}

.meta-item {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.template-description {
  color: var(--el-text-color-secondary);
  font-size: 14px;
}

.preview-content {
  flex: 1;
  overflow: hidden;
}

.content-section {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--el-border-color);
}

.content-actions {
  display: flex;
  gap: 8px;
}

.content-viewer {
  flex: 1;
  overflow: auto;
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
}

.template-content {
  padding: 16px;
  margin: 0;
  white-space: pre-wrap;
  font-family: monospace;
  font-size: 12px;
  line-height: 1.6;
  background: var(--el-fill-color-light);
}

.variables-section {
  padding: 20px;
}

.empty-state {
  text-align: center;
  padding: 40px;
}

.variables-header {
  margin-bottom: 16px;
  font-size: 16px;
  font-weight: 500;
}

.empty-value {
  color: var(--el-text-color-placeholder);
  font-style: italic;
}

.options-preview {
  font-size: 12px;
}

.options-title {
  font-weight: 500;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}

.options-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.option-item {
  display: flex;
  justify-content: space-between;
}

.option-label {
  color: var(--el-text-color-primary);
}

.option-value {
  color: var(--el-text-color-secondary);
  font-style: italic;
}

.statistics-section {
  padding: 20px;
}

.stat-card {
  border-radius: 8px;
}

.stat-content {
  display: flex;
  align-items: center;
  padding: 10px;
}

.stat-icon {
  width: 50px;
  height: 50px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  background: var(--el-color-primary);
  color: white;
  font-size: 20px;
}

.stat-icon.success {
  background: var(--el-color-success);
}

.stat-icon.info {
  background: var(--el-color-info);
}

.stat-info {
  flex: 1;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  line-height: 1;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 14px;
  color: var(--el-text-color-secondary);
}

.chart-placeholder {
  height: 200px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.info-section {
  padding: 20px;
}

.error-state {
  padding: 40px;
}
</style> 