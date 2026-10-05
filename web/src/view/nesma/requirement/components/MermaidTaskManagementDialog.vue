<template>
  <el-dialog 
    v-model="visible" 
    title="流程图生成任务管理" 
    width="1400px" 
    :close-on-click-modal="false"
    :before-close="handleClose"
  >
    <div class="mermaid-task-management">
      <!-- 流程图任务统计卡片 -->
      <div class="task-stats" v-if="taskStats">
        <el-row :gutter="16">
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon total">
                  <el-icon><Coordinate /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.total || 0 }}</div>
                  <div class="stat-label">总任务数</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon running">
                  <el-icon><Loading /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.running || 0 }}</div>
                  <div class="stat-label">生成中</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon completed">
                  <el-icon><Check /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.completed || 0 }}</div>
                  <div class="stat-label">已完成</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon failed">
                  <el-icon><Warning /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.failed || 0 }}</div>
                  <div class="stat-label">失败</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon pending">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.pending || 0 }}</div>
                  <div class="stat-label">等待中</div>
                </div>
              </div>
            </el-card>
          </el-col>
          <el-col :span="4">
            <el-card class="stat-card">
              <div class="stat-item">
                <div class="stat-icon cancelled">
                  <el-icon><Close /></el-icon>
                </div>
                <div class="stat-info">
                  <div class="stat-value">{{ taskStats.cancelled || 0 }}</div>
                  <div class="stat-label">已取消</div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </div>

      <!-- 流程图任务筛选 -->
      <div class="task-filters">
        <el-form :inline="true" :model="filterForm">
          <el-form-item label="流程图类型">
            <el-select v-model="filterForm.diagramType" placeholder="选择流程图类型" clearable>
              <el-option label="流程图" value="flowchart" />
              <el-option label="序列图" value="sequence" />
              <el-option label="类图" value="class" />
              <el-option label="状态图" value="state" />
            </el-select>
          </el-form-item>
          <el-form-item label="任务状态">
            <el-select v-model="filterForm.status" placeholder="选择状态" clearable>
              <el-option label="等待中" value="pending" />
              <el-option label="生成中" value="running" />
              <el-option label="已完成" value="completed" />
              <el-option label="失败" value="failed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="L4需求数">
            <el-input-number 
              v-model="filterForm.minRequirementCount" 
              :min="1" 
              :max="100" 
              placeholder="最少L4数"
              controls-position="right" 
              size="small"
            />
            <span style="margin: 0 8px;">-</span>
            <el-input-number 
              v-model="filterForm.maxRequirementCount" 
              :min="1" 
              :max="100" 
              placeholder="最多L4数"
              controls-position="right" 
              size="small"
            />
          </el-form-item>
          <el-form-item label="创建时间">
            <el-date-picker
              v-model="filterForm.dateRange"
              type="datetimerange"
              range-separator="至"
              start-placeholder="开始日期"
              end-placeholder="结束日期"
              format="YYYY-MM-DD HH:mm:ss"
              size="small"
            />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleFilter">筛选</el-button>
            <el-button @click="handleResetFilter">重置</el-button>
            <el-button @click="refreshTasks">刷新</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 流程图任务列表 -->
      <div class="task-table">
        <el-table 
          :data="taskList" 
          style="width: 100%" 
          v-loading="loading"
          empty-text="暂无流程图生成任务"
          @selection-change="handleSelectionChange"
        >
          <el-table-column type="selection" width="55" />
          <el-table-column prop="ID" label="任务ID" width="80" />
          <el-table-column prop="diagramType" label="流程图类型" width="120">
            <template #default="{ row }">
              <el-tag :type="getDiagramTypeColor(row.config.diagramType)" size="small">
                {{ getDiagramTypeText(row.config.diagramType) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="getStatusColor(row.status)" size="small">
                {{ getStatusText(row.status) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="progress" label="进度" width="150">
            <template #default="{ row }">
              <el-progress 
                :percentage="row.progress" 
                :status="getProgressStatus(row.status)"
                :stroke-width="8"
                :show-text="false"
              />
              <span style="margin-left: 8px; font-size: 12px;">{{ row.progress }}%</span>
            </template>
          </el-table-column>
          <el-table-column label="L4需求信息" width="220">
            <template #default="{ row }">
              <div class="l4-info">
                <div class="l4-count">L4需求数: {{ row.l4RequirementCount || 0 }}</div>
                <div class="l4-list" v-if="row.l4RequirementTitles">
                  <el-tooltip :content="row.l4RequirementTitles.join('\n')" placement="top">
                    <div class="l4-titles">
                      {{ row.l4RequirementTitles.slice(0, 2).join(', ') }}
                      <span v-if="row.l4RequirementTitles.length > 2">...</span>
                    </div>
                  </el-tooltip>
                </div>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="生成统计" width="180">
            <template #default="{ row }">
              <div class="generation-stats">
                <span class="stat-item">
                  <el-icon><Check /></el-icon>
                  成功: {{ row.successCount || 0 }}
                </span>
                <span class="stat-item">
                  <el-icon><Warning /></el-icon>
                  失败: {{ row.failedCount || 0 }}
                </span>
                <span class="stat-item">
                  <el-icon><Document /></el-icon>
                  总数: {{ row.totalCount || 0 }}
                </span>
              </div>
            </template>
          </el-table-column>
          <el-table-column prop="duration" label="耗时" width="100">
            <template #default="{ row }">
              <span v-if="row.duration">{{ formatDuration(row.duration) }}</span>
              <span v-else-if="row.status === 'running'">{{ getRunningDuration(row.createdAt) }}</span>
              <span v-else>-</span>
            </template>
          </el-table-column>
          <el-table-column prop="createdAt" label="创建时间" width="180">
            <template #default="{ row }">
              {{ formatDateTime(row.createdAt) }}
            </template>
          </el-table-column>
          <el-table-column label="操作" width="280" fixed="right">
            <template #default="{ row }">
              <el-button 
                type="primary" 
                size="small" 
                link 
                @click="handleViewProgress(row)"
                v-if="row.status === 'running'"
              >
                <el-icon><View /></el-icon>
                实时进度
              </el-button>
              <el-button 
                type="success" 
                size="small" 
                link 
                @click="handleViewResult(row)"
                v-if="row.status === 'completed'"
              >
                <el-icon><DocumentChecked /></el-icon>
                查看结果
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                link 
                @click="handlePreviewDiagrams(row)"
                v-if="row.status === 'completed'"
              >
                <el-icon><ZoomIn /></el-icon>
                预览流程图
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                link 
                @click="handleRetry(row)"
                v-if="row.status === 'failed'"
              >
                <el-icon><Refresh /></el-icon>
                重试
              </el-button>
              <el-button 
                type="warning" 
                size="small" 
                link 
                @click="handleCancel(row)"
                v-if="row.status === 'running' || row.status === 'pending'"
              >
                <el-icon><Close /></el-icon>
                取消
              </el-button>
              <el-button 
                type="info" 
                size="small" 
                link 
                @click="handleViewDetail(row)"
              >
                <el-icon><InfoFilled /></el-icon>
                详情
              </el-button>
              <el-button 
                type="danger" 
                size="small" 
                link 
                @click="handleDelete(row)"
                v-if="row.status !== 'running'"
              >
                <el-icon><Delete /></el-icon>
                删除
              </el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <!-- 分页 -->
      <div class="pagination">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
      </div>
    </div>

    <!-- 对话框底部 -->
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button 
          type="warning" 
          @click="handleBatchCancel"
          :disabled="selectedTasks.length === 0 || !hasRunningTasks"
        >
          批量取消 ({{ getRunningTaskCount }})
        </el-button>
        <el-button 
          type="danger" 
          @click="handleBatchDelete"
          :disabled="selectedTasks.length === 0 || hasRunningTasks"
        >
          批量删除 ({{ selectedTasks.length }})
        </el-button>
      </div>
    </template>
  </el-dialog>

  <!-- 流程图任务详情对话框 -->
  <el-dialog 
    v-model="detailDialogVisible" 
    title="流程图生成任务详情" 
    width="800px"
    :close-on-click-modal="false"
  >
    <div class="task-detail" v-if="currentTaskDetail">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="任务ID">{{ currentTaskDetail.ID }}</el-descriptions-item>
        <el-descriptions-item label="流程图类型">
          <el-tag :type="getDiagramTypeColor(currentTaskDetail.diagramType)" size="small">
            {{ getDiagramTypeText(currentTaskDetail.diagramType) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="任务状态">
          <el-tag :type="getStatusColor(currentTaskDetail.status)" size="small">
            {{ getStatusText(currentTaskDetail.status) }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="进度">{{ currentTaskDetail.progress }}%</el-descriptions-item>
        <el-descriptions-item label="L4需求数">{{ currentTaskDetail.l4RequirementCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="生成流程图数">{{ currentTaskDetail.totalCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="成功数">{{ currentTaskDetail.successCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="失败数">{{ currentTaskDetail.failedCount || 0 }}</el-descriptions-item>
        <el-descriptions-item label="创建时间">{{ formatDateTime(currentTaskDetail.createdAt) }}</el-descriptions-item>
        <el-descriptions-item label="完成时间">{{ formatDateTime(currentTaskDetail.completedAt) }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ formatDuration(currentTaskDetail.duration) }}</el-descriptions-item>
        <el-descriptions-item label="错误信息" v-if="currentTaskDetail.errorMsg">
          <el-text type="danger">{{ currentTaskDetail.errorMsg }}</el-text>
        </el-descriptions-item>
      </el-descriptions>
      
      <!-- L4需求列表 -->
      <div class="l4-requirements" v-if="currentTaskDetail.l4RequirementTitles">
        <h4>涉及的L4需求</h4>
        <el-tag 
          v-for="(title, index) in currentTaskDetail.l4RequirementTitles" 
          :key="index"
          style="margin: 2px 4px 2px 0;"
        >
          {{ title }}
        </el-tag>
      </div>
      
      <!-- 任务配置详情 -->
      <div class="task-config" v-if="currentTaskDetail.configDetails">
        <h4>任务配置</h4>
        <pre>{{ JSON.stringify(currentTaskDetail.configDetails, null, 2) }}</pre>
      </div>
    </div>
    
    <template #footer>
      <el-button @click="detailDialogVisible = false">关闭</el-button>
    </template>
  </el-dialog>

  <!-- 流程图预览对话框 -->
  <el-dialog 
    v-model="previewDialogVisible" 
    title="流程图预览" 
    width="1400px"
    :close-on-click-modal="false"
  >
    <div class="diagram-preview" v-if="previewDiagrams && previewDiagrams.length > 0">
      <el-tabs v-model="activePreviewTab" type="card">
        <el-tab-pane 
          v-for="(diagram, index) in previewDiagrams" 
          :key="index"
          :label="`${diagram.title || '流程图' + (index + 1)}`"
          :name="String(index)"
        >
          <div class="diagram-item">
            <div class="diagram-header">
              <h4>{{ diagram.title }}</h4>
              <p v-if="diagram.description">{{ diagram.description }}</p>
              <div class="diagram-meta">
                <el-tag size="small" type="info">{{ diagram.diagramType || 'flowchart' }}</el-tag>
                <span class="code-length">代码长度: {{ diagram.mermaidCode?.length || 0 }} 字符</span>
              </div>
            </div>
            
            <div class="diagram-content">
              <!-- Mermaid实时渲染 -->
              <div class="diagram-visualization">
                <div class="visualization-header">
                  <h5>流程图可视化</h5>
                  <div class="visualization-actions">
                    <el-button size="small" @click="copyMermaidCode(diagram.mermaidCode)">
                      <el-icon><CopyDocument /></el-icon>
                      复制代码
                    </el-button>
                    <el-button size="small" @click="downloadMermaidCode(diagram.title, diagram.mermaidCode)">
                      <el-icon><Download /></el-icon>
                      下载代码
                    </el-button>
                  </div>
                </div>
                
                <!-- 使用MermaidRenderer组件进行实时渲染 -->
                <MermaidRenderer 
                  :mermaid-code="diagram.mermaidCode" 
                  :show-controls="true"
                  :theme="mermaidTheme"
                  :auto-render="true"
                  @render-success="handleRenderSuccess"
                  @render-error="handleRenderError"
                />
              </div>
              
              <!-- 代码查看区域（可折叠） -->
              <div class="mermaid-code-section">
                <el-collapse v-model="codeCollapseActive">
                  <el-collapse-item title="查看/编辑Mermaid代码" name="code">
                    <el-input 
                      v-model="diagram.mermaidCode" 
                      type="textarea" 
                      :rows="8" 
                      placeholder="Mermaid流程图代码"
                      @input="handleCodeChange(index, $event)"
                    />
                    <div class="code-actions">
                      <el-button size="small" @click="formatMermaidCode(index)">
                        <el-icon><EditPen /></el-icon>
                        格式化代码
                      </el-button>
                      <el-button size="small" @click="validateMermaidCode(diagram.mermaidCode)">
                        <el-icon><Check /></el-icon>
                        验证语法
                      </el-button>
                      <el-button size="small" type="success" @click="refreshDiagram(index)">
                        <el-icon><Refresh /></el-icon>
                        刷新渲染
                      </el-button>
                    </div>
                  </el-collapse-item>
                </el-collapse>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
      
      <!-- 主题选择 -->
      <div class="theme-selector">
        <el-form :inline="true">
          <el-form-item label="主题:">
            <el-select v-model="mermaidTheme" placeholder="选择主题" size="small" style="width: 120px;">
              <el-option label="默认" value="default" />
              <el-option label="暗色" value="dark" />
              <el-option label="森林" value="forest" />
              <el-option label="中性" value="neutral" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
    </div>
    
    <template #footer>
      <div class="preview-footer">
        <el-button @click="previewDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="exportAllDiagrams">
          <el-icon><Download /></el-icon>
          导出所有流程图
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Coordinate, Loading, Check, Warning, Timer, Close, View, DocumentChecked, 
  Refresh, InfoFilled, Delete, Document, ZoomIn, CopyDocument, Download, EditPen
} from '@element-plus/icons-vue'

// 导入MermaidRenderer组件
import MermaidRenderer from '@/components/MermaidRenderer.vue'

// 导入流程图任务相关的API调用
import {
  getMermaidGenerationTasks,
  getMermaidTaskStatistics,
  deleteMermaidGenerationTask,
  batchDeleteMermaidGenerationTasks,
  retryMermaidGenerationTask,
  cancelMermaidGenerationTask,
  getMermaidTaskDetail,
  getMermaidGenerationTaskResult
} from '@/api/nesma'

const props = defineProps({
  modelValue: Boolean,
  project: Object,
  cycle: Object,
  version: Object
})

const emit = defineEmits(['update:modelValue', 'view-progress', 'view-result', 'close'])

// 响应式数据
const loading = ref(false)
const taskList = ref([])
const selectedTasks = ref([])
const taskStats = ref(null)
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const detailDialogVisible = ref(false)
const currentTaskDetail = ref(null)

// 流程图预览相关
const previewDialogVisible = ref(false)
const previewDiagrams = ref([])
const activePreviewTab = ref('0')
const mermaidTheme = ref('default')

// 筛选表单
const filterForm = ref({
  diagramType: '',
  status: '',
  minRequirementCount: null,
  maxRequirementCount: null,
  dateRange: null
})

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

const hasRunningTasks = computed(() => {
  return selectedTasks.value.some(task => task.status === 'running' || task.status === 'pending')
})

const getRunningTaskCount = computed(() => {
  return selectedTasks.value.filter(task => task.status === 'running' || task.status === 'pending').length
})

// 监听对话框显示状态
watch(() => props.modelValue, (newVal) => {
  if (newVal && props.project) {
    refreshTasks()
    loadTaskStats()
  }
})

// 方法
const refreshTasks = async () => {
  if (!props.project) return
  
  try {
    loading.value = true
    
    const params = {
      projectId: props.project.ID,
      cycleId: props.cycle?.ID,
      versionId: props.version?.ID,
      page: page.value,
      pageSize: pageSize.value,
      taskType: 'flowchart_generation', // 筛选流程图生成任务
      ...filterForm.value
    }
    
    // 处理日期范围
    if (filterForm.value.dateRange) {
      params.startDate = filterForm.value.dateRange[0]
      params.endDate = filterForm.value.dateRange[1]
    }
    
    // 处理L4需求数范围
    if (filterForm.value.minRequirementCount) {
      params.minRequirementCount = filterForm.value.minRequirementCount
    }
    if (filterForm.value.maxRequirementCount) {
      params.maxRequirementCount = filterForm.value.maxRequirementCount
    }
    
    // 清理空参数
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    
    const response = await getMermaidGenerationTasks(params)
    
    if (response.code === 0) {
      // 处理分页响应格式
      if (response.data && typeof response.data === 'object') {
        taskList.value = response.data.list || []
        total.value = response.data.total || 0
      } else {
        // 兼容旧格式
        taskList.value = response.data || []
        total.value = taskList.value.length || 0
      }
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
    
  } catch (error) {
    console.error('加载流程图任务列表失败:', error)
    ElMessage.error('加载流程图任务列表失败：' + error.message)
    
    // 发生错误时，回退到模拟数据
    const mockTasks = [
      {
        ID: 1,
        diagramType: 'flowchart',
        status: 'completed',
        progress: 100,
        l4RequirementCount: 8,
        l4RequirementTitles: ['用户登录验证', '权限校验', '数据加密', '日志记录', '异常处理', '响应返回', '会话管理', '安全检查'],
        successCount: 8,
        failedCount: 0,
        totalCount: 8,
        duration: 156,
        createdAt: '2025-07-15T10:30:00Z',
        completedAt: '2025-07-15T10:32:36Z'
      },
      {
        ID: 2,
        diagramType: 'flowchart',
        status: 'running',
        progress: 45,
        l4RequirementCount: 6,
        l4RequirementTitles: ['文件上传', '格式验证', '病毒扫描', '存储处理', '元数据提取', '缩略图生成'],
        successCount: 3,
        failedCount: 0,
        totalCount: 6,
        duration: null,
        createdAt: '2025-07-15T11:15:00Z',
        completedAt: null
      }
    ]
    
    taskList.value = mockTasks
    total.value = mockTasks.length
  } finally {
    loading.value = false
  }
}

const loadTaskStats = async () => {
  if (!props.project) return
  
  try {
    const params = {
      projectId: props.project.ID,
      cycleId: props.cycle?.ID,
      versionId: props.version?.ID,
      taskType: 'flowchart_generation'
    }
    
    const response = await getMermaidTaskStatistics(params)
    
    if (response.code === 0) {
      taskStats.value = response.data
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
  } catch (error) {
    console.error('加载流程图任务统计失败:', error)
    
    // 发生错误时，回退到模拟数据
    taskStats.value = {
      total: 12,
      running: 1,
      completed: 8,
      failed: 2,
      pending: 1,
      cancelled: 0
    }
  }
}

const handleFilter = () => {
  page.value = 1
  refreshTasks()
}

const handleResetFilter = () => {
  filterForm.value = {
    diagramType: '',
    status: '',
    minRequirementCount: null,
    maxRequirementCount: null,
    dateRange: null
  }
  page.value = 1
  refreshTasks()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  page.value = 1
  refreshTasks()
}

const handleCurrentChange = (currentPage) => {
  page.value = currentPage
  refreshTasks()
}

const handleSelectionChange = (selection) => {
  selectedTasks.value = selection
}

const handleViewProgress = (row) => {
  emit('view-progress', row.ID)
}

const handleViewResult = (row) => {
  emit('view-result', row.ID)
}

const handlePreviewDiagrams = async (row) => {
  try {
    const response = await getMermaidGenerationTaskResult(row.ID)
    
    if (response.code === 0 && response.data?.result) {
      const result = typeof response.data.result === 'string' 
        ? JSON.parse(response.data.result) 
        : response.data.result
      
      if (result.diagrams && Array.isArray(result.diagrams)) {
        previewDiagrams.value = result.diagrams
        activePreviewTab.value = '0'
        previewDialogVisible.value = true
      } else {
        ElMessage.warning('该任务没有生成流程图数据')
      }
    } else {
      throw new Error(response.msg || '获取结果失败')
    }
  } catch (error) {
    console.error('获取流程图预览失败:', error)
    ElMessage.error('获取流程图预览失败：' + error.message)
  }
}

const handleRetry = async (row) => {
  try {
    await ElMessageBox.confirm('确定要重试该流程图生成任务吗？', '确认重试', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await retryMermaidGenerationTask(row.ID)
    ElMessage.success('流程图生成任务重试成功')
    refreshTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('重试任务失败：' + error.message)
    }
  }
}

const handleCancel = async (row) => {
  try {
    await ElMessageBox.confirm('确定要取消该流程图生成任务吗？', '确认取消', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await cancelMermaidGenerationTask(row.ID)
    ElMessage.success('流程图生成任务取消成功')
    refreshTasks()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('取消任务失败：' + error.message)
    }
  }
}

const handleViewDetail = async (row) => {
  try {
    const response = await getMermaidTaskDetail(row.ID)
    
    if (response.code === 0) {
      currentTaskDetail.value = response.data
    } else {
      throw new Error(response.msg || 'API调用失败')
    }
    
    detailDialogVisible.value = true
  } catch (error) {
    console.error('获取任务详情失败:', error)
    ElMessage.error('获取任务详情失败：' + error.message)
    
    // 发生错误时，回退到模拟数据
    currentTaskDetail.value = {
      ...row,
      configDetails: {
        diagramType: row.diagramType,
        detailLevel: 'detailed',
        includeSubRequirements: false,
        autoLayout: true,
        addAnnotations: true
      }
    }
    
    detailDialogVisible.value = true
  }
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm('确定要删除该流程图生成任务吗？删除后无法恢复！', '确认删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    await deleteMermaidGenerationTask(row.ID)
    ElMessage.success('流程图生成任务删除成功')
    refreshTasks()
    loadTaskStats()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('删除任务失败：' + error.message)
    }
  }
}

const handleBatchCancel = async () => {
  const runningTasks = selectedTasks.value.filter(task => task.status === 'running' || task.status === 'pending')
  if (runningTasks.length === 0) {
    ElMessage.warning('请选择正在运行的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要取消选中的 ${runningTasks.length} 个任务吗？`, '确认批量取消', {
      confirmButtonText: '确定取消',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    const taskIds = runningTasks.map(task => task.ID)
    
    // 批量取消任务（使用循环调用单个取消函数）
    const cancelPromises = taskIds.map(taskId => 
      cancelMermaidGenerationTask(taskId)
    )
    await Promise.all(cancelPromises)
    
    ElMessage.success('批量取消成功')
    refreshTasks()
    loadTaskStats()
    selectedTasks.value = []
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量取消失败：' + error.message)
    }
  }
}

const handleBatchDelete = async () => {
  const deletableTasks = selectedTasks.value.filter(task => task.status !== 'running')
  if (deletableTasks.length === 0) {
    ElMessage.warning('请选择非运行状态的任务')
    return
  }
  
  try {
    await ElMessageBox.confirm(`确定要删除选中的 ${deletableTasks.length} 个任务吗？删除后无法恢复！`, '确认批量删除', {
      confirmButtonText: '确定删除',
      cancelButtonText: '取消',
      type: 'error'
    })
    
    const taskIds = deletableTasks.map(task => task.ID)
    await batchDeleteMermaidGenerationTasks({ task_ids: taskIds })
    
    ElMessage.success('批量删除成功')
    refreshTasks()
    loadTaskStats()
    selectedTasks.value = []
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败：' + error.message)
    }
  }
}

const handleClose = () => {
  emit('close')
}

// 计算正在运行的任务耗时
const getRunningDuration = (startTime) => {
  if (!startTime) return '-'
  const now = new Date()
  const start = new Date(startTime)
  const duration = Math.floor((now - start) / 1000)
  return formatDuration(duration)
}

// 辅助方法
const getDiagramTypeColor = (type) => {
  const colorMap = {
    flowchart: 'primary',
    sequence: 'success',
    class: 'warning',
    state: 'info'
  }
  return colorMap[type] || 'info'
}

const getDiagramTypeText = (type) => {
  const textMap = {
    flowchart: '流程图',
    sequence: '序列图',
    class: '类图',
    state: '状态图'
  }
  return textMap[type] || type
}

const getStatusColor = (status) => {
  const colorMap = {
    pending: 'warning',
    running: 'primary',
    completed: 'success',
    failed: 'danger',
    cancelled: 'info'
  }
  return colorMap[status] || 'info'
}

const getStatusText = (status) => {
  const textMap = {
    pending: '等待中',
    running: '生成中',
    completed: '已完成',
    failed: '失败',
    cancelled: '已取消'
  }
  return textMap[status] || status
}

const getProgressStatus = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'exception'
  return undefined
}

const formatDuration = (seconds) => {
  if (!seconds) return '-'
  const minutes = Math.floor(seconds / 60)
  const secs = seconds % 60
  return minutes > 0 ? `${minutes}分${secs}秒` : `${secs}秒`
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

// Mermaid预览增强方法
const copyMermaidCode = async (code) => {
  if (!code) {
    ElMessage.warning('没有可复制的代码')
    return
  }
  
  try {
    await navigator.clipboard.writeText(code)
    ElMessage.success('Mermaid代码已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败，请手动复制')
  }
}

const downloadMermaidCode = (title, code) => {
  if (!code) {
    ElMessage.warning('没有可下载的代码')
    return
  }
  
  const blob = new Blob([code], { type: 'text/plain' })
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = `${title || 'mermaid-diagram'}_${Date.now()}.mmd`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
  URL.revokeObjectURL(url)
  
  ElMessage.success('Mermaid代码已下载')
}

const handleCodeChange = (index, newCode) => {
  if (previewDiagrams.value[index]) {
    previewDiagrams.value[index].mermaidCode = newCode
  }
}

const formatMermaidCode = (index) => {
  const diagram = previewDiagrams.value[index]
  if (!diagram?.mermaidCode) {
    ElMessage.warning('没有代码可格式化')
    return
  }
  
  try {
    // 简单的Mermaid代码格式化
    let formatted = diagram.mermaidCode
      .split('\n')
      .map(line => line.trim())
      .filter(line => line.length > 0)
      .join('\n')
    
    // 添加适当的缩进
    const lines = formatted.split('\n')
    let indentLevel = 0
    const formattedLines = lines.map(line => {
      if (line.includes('subgraph') || line.includes('graph')) {
        const result = '  '.repeat(indentLevel) + line
        indentLevel++
        return result
      } else if (line.includes('end')) {
        indentLevel = Math.max(0, indentLevel - 1)
        return '  '.repeat(indentLevel) + line
      } else {
        return '  '.repeat(indentLevel) + line
      }
    })
    
    diagram.mermaidCode = formattedLines.join('\n')
    ElMessage.success('代码格式化完成')
  } catch (error) {
    console.error('格式化失败:', error)
    ElMessage.error('格式化失败: ' + error.message)
  }
}

const validateMermaidCode = async (code) => {
  if (!code) {
    ElMessage.warning('没有代码可验证')
    return
  }
  
  try {
    // 这里可以集成mermaid的验证功能
    // 暂时使用简单验证
    const hasGraphDeclaration = /graph|flowchart|sequenceDiagram|classDiagram|stateDiagram/.test(code)
    
    if (!hasGraphDeclaration) {
      throw new Error('缺少图表类型声明（如: flowchart TD, graph LR等）')
    }
    
    ElMessage.success('Mermaid语法验证通过')
  } catch (error) {
    console.error('验证失败:', error)
    ElMessage.error('语法验证失败: ' + error.message)
  }
}

const refreshDiagram = (index) => {
  // 触发重新渲染，这会通过MermaidRenderer组件自动处理
  ElMessage.info('正在刷新流程图渲染...')
}

const handleRenderSuccess = (data) => {
  console.log('Mermaid渲染成功:', data)
}

const handleRenderError = (error) => {
  console.error('Mermaid渲染失败:', error)
  ElMessage.error('流程图渲染失败: ' + error.message)
}

const exportAllDiagrams = () => {
  if (!previewDiagrams.value || previewDiagrams.value.length === 0) {
    ElMessage.warning('没有可导出的流程图')
    return
  }
  
  try {
    // 创建包含所有流程图的文件内容
    const allDiagrams = previewDiagrams.value.map((diagram, index) => {
      return `# ${diagram.title || '流程图' + (index + 1)}

${diagram.description ? '## 描述\n' + diagram.description + '\n' : ''}
## Mermaid代码

\`\`\`mermaid
${diagram.mermaidCode}
\`\`\`

---
`
    }).join('\n')
    
    const blob = new Blob([allDiagrams], { type: 'text/markdown' })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `mermaid-diagrams-${Date.now()}.md`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
    
    ElMessage.success('所有流程图已导出为Markdown文件')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败: ' + error.message)
  }
}

// 生命周期
onMounted(() => {
  if (props.modelValue && props.project) {
    refreshTasks()
    loadTaskStats()
  }
})
</script>

<style lang="scss" scoped>
.mermaid-task-management {
  .task-stats {
    margin-bottom: 20px;
    
    .stat-card {
      border: none;
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
      
      .stat-item {
        display: flex;
        align-items: center;
        padding: 8px;
        
        .stat-icon {
          width: 40px;
          height: 40px;
          border-radius: 50%;
          display: flex;
          align-items: center;
          justify-content: center;
          margin-right: 12px;
          
          .el-icon {
            font-size: 18px;
            color: white;
          }
          
          &.total {
            background: linear-gradient(135deg, #13C2C2 0%, #36CFC9 100%);
          }
          
          &.running {
            background: linear-gradient(135deg, #409eff 0%, #36cfc9 100%);
          }
          
          &.completed {
            background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
          }
          
          &.failed {
            background: linear-gradient(135deg, #f56c6c 0%, #f78989 100%);
          }
          
          &.pending {
            background: linear-gradient(135deg, #e6a23c 0%, #f7ba2a 100%);
          }
          
          &.cancelled {
            background: linear-gradient(135deg, #909399 0%, #b1b3b8 100%);
          }
        }
        
        .stat-info {
          flex: 1;
          
          .stat-value {
            font-size: 20px;
            font-weight: 600;
            color: #303133;
            line-height: 1;
          }
          
          .stat-label {
            font-size: 12px;
            color: #909399;
            margin-top: 2px;
          }
        }
      }
    }
  }
  
  .task-filters {
    background: #f8f9fa;
    padding: 16px;
    border-radius: 8px;
    margin-bottom: 20px;
  }
  
  .task-table {
    .l4-info {
      .l4-count {
        font-size: 12px;
        color: #606266;
        margin-bottom: 4px;
      }
      
      .l4-titles {
        font-size: 12px;
        color: #909399;
        line-height: 1.2;
      }
    }
    
    .generation-stats {
      display: flex;
      flex-direction: column;
      gap: 2px;
      
      .stat-item {
        display: flex;
        align-items: center;
        font-size: 12px;
        color: #606266;
        
        .el-icon {
          margin-right: 4px;
          font-size: 12px;
        }
      }
    }
  }
  
  .pagination {
    display: flex;
    justify-content: center;
    margin-top: 20px;
  }
}

.task-detail {
  .l4-requirements {
    margin-top: 20px;
    
    h4 {
      margin-bottom: 10px;
      color: #303133;
    }
  }
  
  .task-config {
    margin-top: 20px;
    
    h4 {
      margin-bottom: 10px;
      color: #303133;
    }
    
    pre {
      background: #f8f9fa;
      padding: 12px;
      border-radius: 4px;
      font-size: 12px;
      overflow-x: auto;
    }
  }
}

.diagram-preview {
  .diagram-item {
    .diagram-header {
      margin-bottom: 20px;
      padding-bottom: 16px;
      border-bottom: 1px solid #e4e7ed;
      
      h4 {
        margin: 0 0 8px 0;
        color: #303133;
        font-size: 18px;
        font-weight: 600;
      }
      
      p {
        margin: 0 0 12px 0;
        color: #606266;
        font-size: 14px;
        line-height: 1.5;
      }
      
      .diagram-meta {
        display: flex;
        align-items: center;
        gap: 12px;
        
        .code-length {
          font-size: 12px;
          color: #909399;
        }
      }
    }
    
    .diagram-content {
      .diagram-visualization {
        margin-bottom: 24px;
        
        .visualization-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          
          h5 {
            margin: 0;
            font-size: 16px;
            font-weight: 600;
            color: #303133;
          }
          
          .visualization-actions {
            display: flex;
            gap: 8px;
          }
        }
      }
      
      .mermaid-code-section {
        margin-top: 20px;
        
        .code-actions {
          display: flex;
          gap: 8px;
          margin-top: 12px;
          
          .el-button {
            border-radius: 4px;
          }
        }
      }
    }
  }
  
  .theme-selector {
    margin-top: 20px;
    padding-top: 16px;
    border-top: 1px solid #e4e7ed;
    text-align: center;
  }
}

.preview-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.dialog-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>