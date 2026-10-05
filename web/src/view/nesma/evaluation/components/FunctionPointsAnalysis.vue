<template>
  <div class="function-points-analysis">
    <!-- 功能点分布图表 -->
    <div class="charts-section">
      <el-row :gutter="20">
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>功能点类型分布</span>
            </template>
            <div ref="typeDistributionChart" class="chart-container"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card class="chart-card">
            <template #header>
              <span>复杂度分布</span>
            </template>
            <div ref="complexityDistributionChart" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 功能点详细表格 -->
    <el-card class="table-card">
      <template #header>
        <div class="table-header">
          <span>功能点详细列表</span>
          <div class="header-actions">
            <el-button size="small" @click="refreshData">刷新</el-button>
            <el-button size="small" type="primary" @click="exportExcel">导出Excel</el-button>
          </div>
        </div>
      </template>

      <el-table
        v-loading="loading"
        :data="functionPoints"
        style="width: 100%"
        :default-sort="{ prop: 'functionPointValue', order: 'descending' }"
        @sort-change="handleSortChange"
      >
        <el-table-column prop="functionName" label="功能名称" min-width="200" show-overflow-tooltip />
        <el-table-column prop="functionType" label="功能类型" width="100">
          <template #default="scope">
            <el-tag :type="getFunctionTypeColor(scope.row.functionType)" size="small">
              {{ getFunctionTypeLabel(scope.row.functionType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="complexityLevel" label="复杂度" width="100">
          <template #default="scope">
            <el-tag :type="getComplexityColor(scope.row.complexityLevel)" size="small">
              {{ getComplexityLabel(scope.row.complexityLevel) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="functionPointValue" label="功能点值" width="120" sortable="custom">
          <template #default="scope">
            <span class="fp-value">{{ scope.row.functionPointValue?.toFixed(1) || 0 }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="identificationMethod" label="识别方法" width="120">
          <template #default="scope">
            <el-tag :type="getIdentificationMethodColor(scope.row.identificationMethod)" size="small">
              {{ getIdentificationMethodLabel(scope.row.identificationMethod) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="confidenceLevel" label="置信度" width="120">
          <template #default="scope">
            <div class="confidence-display">
              <el-progress
                :percentage="(scope.row.confidenceLevel * 100) || 0"
                :stroke-width="6"
                :show-text="false"
                :color="getConfidenceColor(scope.row.confidenceLevel)"
              />
              <span class="confidence-text">{{ (scope.row.confidenceLevel * 100)?.toFixed(0) || 0 }}%</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="validationStatus" label="验证状态" width="100">
          <template #default="scope">
            <el-tag :type="getValidationStatusColor(scope.row.validationStatus)" size="small">
              {{ getValidationStatusLabel(scope.row.validationStatus) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right">
          <template #default="scope">
            <el-button size="small" @click="viewDetails(scope.row)">详情</el-button>
            <el-button size="small" type="primary" @click="editFunctionPoint(scope.row)">编辑</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-container">
        <el-pagination
          :current-page="currentPage"
          :page-size="pageSize"
          :total="total"
          layout="total, prev, pager, next, jumper"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 统计汇总 -->
    <el-card class="summary-card">
      <template #header>
        <span>统计汇总</span>
      </template>
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="summary-item">
            <div class="summary-label">总功能点数</div>
            <div class="summary-value">{{ summary.totalFunctionPoints?.toFixed(1) || 0 }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-item">
            <div class="summary-label">数据功能点</div>
            <div class="summary-value">{{ summary.dataFunctionPoints?.toFixed(1) || 0 }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-item">
            <div class="summary-label">事务功能点</div>
            <div class="summary-value">{{ summary.transactionFunctionPoints?.toFixed(1) || 0 }}</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="summary-item">
            <div class="summary-label">平均置信度</div>
            <div class="summary-value">{{ (summary.averageConfidence * 100)?.toFixed(1) || 0 }}%</div>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- 功能点详情对话框 -->
    <FunctionPointDetailDialog
      v-model="detailDialogVisible"
      :function-point="selectedFunctionPoint"
    />

    <!-- 编辑功能点对话框 -->
    <EditFunctionPointDialog
      v-model="editDialogVisible"
      :function-point="selectedFunctionPoint"
      @success="refreshData"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, nextTick } from 'vue'
import { ElMessage } from 'element-plus'
import { getFunctionPoints, exportEvaluationReport } from '@/api/nesmaEvaluation'
import * as echarts from 'echarts'
import FunctionPointDetailDialog from './FunctionPointDetailDialog.vue'
import EditFunctionPointDialog from './EditFunctionPointDialog.vue'

// Props
const props = defineProps({
  evaluationId: {
    type: [String, Number],
    required: true
  },
  evaluation: {
    type: Object,
    default: () => ({})
  }
})

// 响应式数据
const loading = ref(false)
const functionPoints = ref([])
const currentPage = ref(1)
const pageSize = ref(20)
const total = ref(0)
const detailDialogVisible = ref(false)
const editDialogVisible = ref(false)
const selectedFunctionPoint = ref(null)

// 图表引用
const typeDistributionChart = ref(null)
const complexityDistributionChart = ref(null)
let typeChart = null
let complexityChart = null

// 统计汇总
const summary = reactive({
  totalFunctionPoints: 0,
  dataFunctionPoints: 0,
  transactionFunctionPoints: 0,
  averageConfidence: 0
})

// 方法
const loadFunctionPoints = async () => {
  loading.value = true
  try {
    const response = await getFunctionPoints({
      evaluationId: props.evaluationId,
      page: currentPage.value,
      pageSize: pageSize.value
    })
    functionPoints.value = response.data?.list || []
    total.value = response.data?.total || 0
    
    // 计算统计汇总
    calculateSummary()
    
    // 更新图表
    updateCharts()
  } catch (error) {
    console.error('加载功能点列表失败:', error)
    ElMessage.error('加载功能点列表失败')
  } finally {
    loading.value = false
  }
}

const calculateSummary = () => {
  const points = functionPoints.value
  summary.totalFunctionPoints = points.reduce((sum, item) => sum + (item.functionPointValue || 0), 0)
  summary.dataFunctionPoints = points
    .filter(item => ['ILF', 'EIF'].includes(item.functionType))
    .reduce((sum, item) => sum + (item.functionPointValue || 0), 0)
  summary.transactionFunctionPoints = points
    .filter(item => ['EI', 'EO', 'EQ'].includes(item.functionType))
    .reduce((sum, item) => sum + (item.functionPointValue || 0), 0)
  summary.averageConfidence = points.length > 0 
    ? points.reduce((sum, item) => sum + (item.confidenceLevel || 0), 0) / points.length
    : 0
}

const updateCharts = () => {
  nextTick(() => {
    updateTypeDistributionChart()
    updateComplexityDistributionChart()
  })
}

const updateTypeDistributionChart = () => {
  if (!typeChart) {
    typeChart = echarts.init(typeDistributionChart.value)
  }
  
  const typeData = {}
  functionPoints.value.forEach(item => {
    if (!typeData[item.functionType]) {
      typeData[item.functionType] = 0
    }
    typeData[item.functionType] += item.functionPointValue || 0
  })
  
  const data = Object.entries(typeData).map(([type, value]) => ({
    name: getFunctionTypeLabel(type),
    value: value
  }))
  
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    series: [{
      name: '功能点类型',
      type: 'pie',
      radius: ['50%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: false,
        position: 'center'
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 20,
          fontWeight: 'bold'
        }
      },
      labelLine: {
        show: false
      },
      data: data
    }]
  }
  
  typeChart.setOption(option)
}

const updateComplexityDistributionChart = () => {
  if (!complexityChart) {
    complexityChart = echarts.init(complexityDistributionChart.value)
  }
  
  const complexityData = {}
  functionPoints.value.forEach(item => {
    if (!complexityData[item.complexityLevel]) {
      complexityData[item.complexityLevel] = 0
    }
    complexityData[item.complexityLevel] += item.functionPointValue || 0
  })
  
  const data = Object.entries(complexityData).map(([level, value]) => ({
    name: getComplexityLabel(level),
    value: value
  }))
  
  const option = {
    tooltip: {
      trigger: 'item'
    },
    xAxis: {
      type: 'category',
      data: data.map(item => item.name)
    },
    yAxis: {
      type: 'value'
    },
    series: [{
      data: data.map(item => ({
        value: item.value,
        itemStyle: {
          color: getComplexityChartColor(item.name)
        }
      })),
      type: 'bar',
      barWidth: '50%',
      itemStyle: {
        borderRadius: [4, 4, 0, 0]
      }
    }]
  }
  
  complexityChart.setOption(option)
}

const refreshData = () => {
  loadFunctionPoints()
}

const exportExcel = async () => {
  try {
    const response = await exportEvaluationReport(props.evaluationId, 'excel')
    const blob = new Blob([response], { type: 'application/vnd.openxmlformats-officedocument.spreadsheetml.sheet' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `功能点分析_${props.evaluation.evaluationName}.xlsx`
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

const viewDetails = (row) => {
  selectedFunctionPoint.value = row
  detailDialogVisible.value = true
}

const editFunctionPoint = (row) => {
  selectedFunctionPoint.value = row
  editDialogVisible.value = true
}

const handleSortChange = ({ prop, order }) => {
  // 处理排序
  loadFunctionPoints()
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadFunctionPoints()
}

// 工具函数
const getFunctionTypeLabel = (type) => {
  const labels = {
    'ILF': '内部逻辑文件',
    'EIF': '外部接口文件',
    'EI': '外部输入',
    'EO': '外部输出',
    'EQ': '外部查询'
  }
  return labels[type] || type
}

const getFunctionTypeColor = (type) => {
  const colors = {
    'ILF': 'primary',
    'EIF': 'success',
    'EI': 'warning',
    'EO': 'danger',
    'EQ': 'info'
  }
  return colors[type] || ''
}

const getComplexityLabel = (level) => {
  const labels = {
    'Low': '低',
    'Average': '中',
    'High': '高'
  }
  return labels[level] || level
}

const getComplexityColor = (level) => {
  const colors = {
    'Low': 'success',
    'Average': 'warning',
    'High': 'danger'
  }
  return colors[level] || ''
}

const getComplexityChartColor = (level) => {
  const colors = {
    '低': '#67c23a',
    '中': '#e6a23c',
    '高': '#f56c6c'
  }
  return colors[level] || '#409eff'
}

const getIdentificationMethodLabel = (method) => {
  const labels = {
    'manual': '手工',
    'ai_assisted': 'AI辅助',
    'rule_based': '规则推导',
    'template': '模板'
  }
  return labels[method] || method
}

const getIdentificationMethodColor = (method) => {
  const colors = {
    'manual': '',
    'ai_assisted': 'primary',
    'rule_based': 'success',
    'template': 'warning'
  }
  return colors[method] || ''
}

const getConfidenceColor = (level) => {
  if (level >= 0.8) return '#67c23a'
  if (level >= 0.6) return '#e6a23c'
  return '#f56c6c'
}

const getValidationStatusLabel = (status) => {
  const labels = {
    'passed': '通过',
    'failed': '失败',
    'pending': '待验证',
    'manual_review': '需人工审核'
  }
  return labels[status] || status
}

const getValidationStatusColor = (status) => {
  const colors = {
    'passed': 'success',
    'failed': 'danger',
    'pending': 'warning',
    'manual_review': 'info'
  }
  return colors[status] || ''
}

// 生命周期
onMounted(() => {
  loadFunctionPoints()
})
</script>

<style scoped>
.function-points-analysis {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.charts-section {
  margin-bottom: 20px;
}

.chart-card {
  margin-bottom: 0;
}

.chart-container {
  height: 300px;
  width: 100%;
}

.table-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.fp-value {
  font-weight: bold;
  color: #409eff;
}

.confidence-display {
  display: flex;
  align-items: center;
  gap: 8px;
}

.confidence-display :deep(.el-progress) {
  flex: 1;
}

.confidence-text {
  font-size: 12px;
  color: #666;
  min-width: 35px;
}

.pagination-container {
  margin-top: 20px;
  text-align: right;
}

.summary-card {
  margin-bottom: 0;
}

.summary-item {
  text-align: center;
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.summary-label {
  font-size: 14px;
  color: #666;
  margin-bottom: 8px;
}

.summary-value {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
}
</style> 