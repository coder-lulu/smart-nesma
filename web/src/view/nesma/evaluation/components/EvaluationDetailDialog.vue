<template>
  <el-dialog
    v-model="dialogVisible"
    title="评估详情"
    width="80%"
    :close-on-click-modal="false"
    @opened="handleOpened"
  >
    <div v-loading="loading" class="evaluation-detail">
      <div v-if="evaluation" class="detail-content">
        <!-- 基本信息 -->
        <el-card class="info-card">
          <template #header>
            <div class="card-header">
              <span>基本信息</span>
              <el-tag :type="getStatusColor(evaluation.status)">
                {{ getStatusLabel(evaluation.status) }}
              </el-tag>
            </div>
          </template>
          <el-row :gutter="20">
            <el-col :span="8">
              <div class="info-item">
                <span class="label">评估名称：</span>
                <span class="value">{{ evaluation.evaluationName }}</span>
              </div>
              <div class="info-item">
                <span class="label">所属项目：</span>
                <span class="value">{{ evaluation.project?.name || '-' }}</span>
              </div>
              <div class="info-item">
                <span class="label">评估版本：</span>
                <span class="value">{{ evaluation.evaluationVersion }}</span>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="info-item">
                <span class="label">评估类型：</span>
                <span class="value">{{ getEvaluationTypeLabel(evaluation.evaluationType) }}</span>
              </div>
              <div class="info-item">
                <span class="label">NESMA规则：</span>
                <span class="value">{{ evaluation.nesmaRules }}</span>
              </div>
              <div class="info-item">
                <span class="label">评估人：</span>
                <span class="value">{{ evaluation.evaluatorName || '-' }}</span>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="info-item">
                <span class="label">开始时间：</span>
                <span class="value">{{ formatDateTime(evaluation.startTime) }}</span>
              </div>
              <div class="info-item">
                <span class="label">完成时间：</span>
                <span class="value">{{ formatDateTime(evaluation.completionTime) }}</span>
              </div>
              <div class="info-item">
                <span class="label">耗时：</span>
                <span class="value">{{ calculateDuration(evaluation.startTime, evaluation.completionTime) }}</span>
              </div>
            </el-col>
          </el-row>
        </el-card>

        <!-- 评估结果 -->
        <el-card class="result-card" v-if="evaluation.status === 'completed'">
          <template #header>
            <span>评估结果</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="6">
              <div class="result-item">
                <div class="result-number">{{ evaluation.totalFunctionPoints?.toFixed(1) || 0 }}</div>
                <div class="result-label">总功能点</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="result-item">
                <div class="result-number">{{ evaluation.unadjustedFunctionPoints?.toFixed(1) || 0 }}</div>
                <div class="result-label">未调整功能点</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="result-item">
                <div class="result-number">{{ evaluation.adjustmentFactor?.toFixed(3) || 0 }}</div>
                <div class="result-label">调整因子</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="result-item">
                <div class="result-number">{{ evaluation.functionPointCount || 0 }}</div>
                <div class="result-label">功能点条目</div>
              </div>
            </el-col>
          </el-row>
        </el-card>

        <!-- 质量指标 -->
        <el-card class="quality-card" v-if="evaluation.status === 'completed'">
          <template #header>
            <span>质量指标</span>
          </template>
          <el-row :gutter="20">
            <el-col :span="8">
              <div class="quality-item">
                <div class="quality-header">
                  <span>置信度</span>
                  <span class="quality-value">{{ (evaluation.confidenceScore * 100)?.toFixed(1) || 0 }}%</span>
                </div>
                <el-progress
                  :percentage="(evaluation.confidenceScore * 100) || 0"
                  :color="getProgressColor(evaluation.confidenceScore)"
                  :stroke-width="8"
                />
              </div>
            </el-col>
            <el-col :span="8">
              <div class="quality-item">
                <div class="quality-header">
                  <span>准确度</span>
                  <span class="quality-value">{{ (evaluation.accuracyScore * 100)?.toFixed(1) || 0 }}%</span>
                </div>
                <el-progress
                  :percentage="(evaluation.accuracyScore * 100) || 0"
                  :color="getProgressColor(evaluation.accuracyScore)"
                  :stroke-width="8"
                />
              </div>
            </el-col>
            <el-col :span="8">
              <div class="quality-item">
                <div class="quality-header">
                  <span>合规度</span>
                  <span class="quality-value">{{ (evaluation.complianceScore * 100)?.toFixed(1) || 0 }}%</span>
                </div>
                <el-progress
                  :percentage="(evaluation.complianceScore * 100) || 0"
                  :color="getProgressColor(evaluation.complianceScore)"
                  :stroke-width="8"
                />
              </div>
            </el-col>
          </el-row>
        </el-card>

        <!-- 详细内容标签页 -->
        <el-card class="tabs-card">
          <el-tabs v-model="activeTab" @tab-click="handleTabClick">
            <!-- 功能点分析 -->
            <el-tab-pane label="功能点分析" name="functionPoints">
              <FunctionPointsAnalysis
                :evaluation-id="evaluationId"
                :evaluation="evaluation"
              />
            </el-tab-pane>

            <!-- 复杂度指标 -->
            <el-tab-pane label="复杂度指标" name="complexityMetrics">
              <ComplexityMetrics
                :evaluation-id="evaluationId"
                :evaluation="evaluation"
              />
            </el-tab-pane>

            <!-- 验证结果 -->
            <el-tab-pane label="验证结果" name="validationResults">
              <ValidationResults
                :evaluation-id="evaluationId"
                :evaluation="evaluation"
              />
            </el-tab-pane>

            <!-- 改进建议 -->
            <el-tab-pane label="改进建议" name="recommendations">
              <Recommendations
                :evaluation-id="evaluationId"
                :evaluation="evaluation"
              />
            </el-tab-pane>
          </el-tabs>
        </el-card>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="dialogVisible = false">关闭</el-button>
        <el-button 
          type="primary" 
          @click="handleExportReport"
          :loading="exportingReport"
          v-if="evaluation?.status === 'completed'"
        >
          导出报告
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getEvaluation, generateEvaluationReport } from '@/api/nesmaEvaluation'
import FunctionPointsAnalysis from './FunctionPointsAnalysis.vue'
import ComplexityMetrics from './ComplexityMetrics.vue'
import ValidationResults from './ValidationResults.vue'
import Recommendations from './Recommendations.vue'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  evaluationId: {
    type: [String, Number],
    default: null
  }
})

// Emits
const emit = defineEmits(['update:modelValue'])

// 响应式数据
const loading = ref(false)
const exportingReport = ref(false)
const evaluation = ref(null)
const activeTab = ref('functionPoints')

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// 监听评估ID变化
watch(() => props.evaluationId, (newId) => {
  if (newId && props.modelValue) {
    loadEvaluationDetail()
  }
}, { immediate: true })

// 方法
const loadEvaluationDetail = async () => {
  if (!props.evaluationId) return
  
  loading.value = true
  try {
    const response = await getEvaluation(props.evaluationId)
    evaluation.value = {
      ...response.data,
      id: response.data.id ?? response.data.ID
    }
  } catch (error) {
    console.error('加载评估详情失败:', error)
    ElMessage.error('加载评估详情失败')
  } finally {
    loading.value = false
  }
}

const handleOpened = () => {
  if (props.evaluationId) {
    loadEvaluationDetail()
  }
}

const handleTabClick = (tab) => {
  activeTab.value = tab.name
}

const handleExportReport = async () => {
  if (!evaluation.value) return
  
  exportingReport.value = true
  try {
    const response = await generateEvaluationReport(evaluation.value.id, 'pdf')
    const blob = new Blob([response], { type: 'application/pdf' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `${evaluation.value.evaluationName}_评估报告.pdf`
    link.click()
    window.URL.revokeObjectURL(url)
    ElMessage.success('报告导出成功')
  } catch (error) {
    console.error('导出报告失败:', error)
    ElMessage.error('导出报告失败')
  } finally {
    exportingReport.value = false
  }
}

// 工具函数
const getStatusColor = (status) => {
  const colors = {
    pending: '',
    processing: 'warning',
    completed: 'success',
    failed: 'danger'
  }
  return colors[status] || ''
}

const getStatusLabel = (status) => {
  const labels = {
    pending: '待开始',
    processing: '进行中',
    completed: '已完成',
    failed: '失败'
  }
  return labels[status] || status
}

const getEvaluationTypeLabel = (type) => {
  const labels = {
    initial: '初步评估',
    detailed: '详细评估',
    final: '最终评估'
  }
  return labels[type] || type
}

const getProgressColor = (score) => {
  if (score >= 0.8) return '#67c23a'
  if (score >= 0.6) return '#e6a23c'
  return '#f56c6c'
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

const calculateDuration = (startTime, endTime) => {
  if (!startTime || !endTime) return '-'
  
  const start = new Date(startTime)
  const end = new Date(endTime)
  const duration = Math.floor((end - start) / (1000 * 60)) // 分钟
  
  if (duration < 60) return `${duration}分钟`
  if (duration < 1440) return `${Math.floor(duration / 60)}小时${duration % 60}分钟`
  return `${Math.floor(duration / 1440)}天${Math.floor((duration % 1440) / 60)}小时`
}
</script>

<style scoped>
.evaluation-detail {
  min-height: 400px;
}

.detail-content {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.info-card, .result-card, .quality-card, .tabs-card {
  margin-bottom: 0;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.info-item {
  display: flex;
  margin-bottom: 12px;
  line-height: 1.5;
}

.info-item .label {
  width: 100px;
  color: #909399;
  font-size: 14px;
}

.info-item .value {
  flex: 1;
  color: #303133;
  font-size: 14px;
  font-weight: 500;
}

.result-item {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
}

.result-number {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  line-height: 1;
  margin-bottom: 8px;
}

.result-label {
  font-size: 14px;
  color: #666;
}

.quality-item {
  padding: 16px;
  background: #f8f9fa;
  border-radius: 8px;
}

.quality-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.quality-value {
  font-size: 16px;
  font-weight: bold;
  color: #303133;
}

.dialog-footer {
  text-align: right;
}

:deep(.el-tabs__content) {
  padding-top: 20px;
}

:deep(.el-progress-bar__outer) {
  border-radius: 4px;
}

:deep(.el-progress-bar__inner) {
  border-radius: 4px;
}
</style> 