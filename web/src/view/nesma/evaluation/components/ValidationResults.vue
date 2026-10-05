<template>
  <div class="validation-results">
    <div v-loading="loading" class="validation-content">
      <!-- 验证概览 -->
      <div class="overview-section">
        <h3>验证概览</h3>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="stat-card pass">
              <div class="stat-value">{{ passedCount }}</div>
              <div class="stat-label">通过项</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card warning">
              <div class="stat-value">{{ warningCount }}</div>
              <div class="stat-label">警告项</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card fail">
              <div class="stat-value">{{ failedCount }}</div>
              <div class="stat-label">失败项</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="stat-card total">
              <div class="stat-value">{{ totalCount }}</div>
              <div class="stat-label">总验证项</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 验证结果汇总 -->
      <div class="summary-section">
        <h3>验证结果汇总</h3>
        <div class="summary-card">
          <div class="summary-header">
            <div class="summary-score">
              <span class="score-value">{{ overallScore }}%</span>
              <span class="score-label">总体得分</span>
            </div>
            <div class="summary-level">
              <el-tag :type="getLevelColor(validationLevel)" size="large">
                {{ validationLevel }}
              </el-tag>
            </div>
          </div>
          <div class="summary-progress">
            <el-progress
              :percentage="overallScore"
              :color="getProgressColor(overallScore)"
              :stroke-width="12"
              :show-text="false"
            />
          </div>
          <div class="summary-description">
            {{ validationDescription }}
          </div>
        </div>
      </div>

      <!-- 详细验证结果 -->
      <div class="detailed-results">
        <h3>详细验证结果</h3>
        <div class="validation-categories">
          <el-tabs v-model="activeCategory" @tab-click="handleCategoryChange">
            <el-tab-pane label="功能点验证" name="functionPoints">
              <ValidationCategory
                :title="'功能点验证'"
                :items="functionPointValidations"
                @item-click="handleItemClick"
              />
            </el-tab-pane>

            <el-tab-pane label="数据一致性" name="dataConsistency">
              <ValidationCategory
                :title="'数据一致性验证'"
                :items="dataConsistencyValidations"
                @item-click="handleItemClick"
              />
            </el-tab-pane>

            <el-tab-pane label="规则合规性" name="ruleCompliance">
              <ValidationCategory
                :title="'规则合规性验证'"
                :items="ruleComplianceValidations"
                @item-click="handleItemClick"
              />
            </el-tab-pane>

            <el-tab-pane label="质量检查" name="qualityCheck">
              <ValidationCategory
                :title="'质量检查'"
                :items="qualityCheckValidations"
                @item-click="handleItemClick"
              />
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>

      <!-- 验证详情对话框 -->
      <el-dialog
        v-model="detailDialogVisible"
        :title="selectedItem?.name"
        width="60%"
        :close-on-click-modal="false"
      >
        <div v-if="selectedItem" class="validation-detail">
          <div class="detail-header">
            <el-tag :type="getStatusColor(selectedItem.status)">
              {{ getStatusLabel(selectedItem.status) }}
            </el-tag>
            <span class="detail-score">得分: {{ selectedItem.score }}%</span>
          </div>
          
          <div class="detail-content">
            <div class="detail-section">
              <h4>验证规则</h4>
              <p>{{ selectedItem.rule }}</p>
            </div>
            
            <div class="detail-section">
              <h4>验证结果</h4>
              <p>{{ selectedItem.result }}</p>
            </div>
            
            <div class="detail-section" v-if="selectedItem.issues && selectedItem.issues.length > 0">
              <h4>发现问题</h4>
              <ul>
                <li v-for="issue in selectedItem.issues" :key="issue.id">
                  {{ issue.description }}
                </li>
              </ul>
            </div>
            
            <div class="detail-section" v-if="selectedItem.recommendations && selectedItem.recommendations.length > 0">
              <h4>改进建议</h4>
              <ul>
                <li v-for="rec in selectedItem.recommendations" :key="rec.id">
                  {{ rec.suggestion }}
                </li>
              </ul>
            </div>
          </div>
        </div>
        
        <template #footer>
          <el-button @click="detailDialogVisible = false">关闭</el-button>
        </template>
      </el-dialog>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import ValidationCategory from './ValidationCategory.vue'

// Props
const props = defineProps({
  evaluationId: {
    type: [String, Number],
    required: true
  },
  evaluation: {
    type: Object,
    required: true
  }
})

// 响应式数据
const loading = ref(false)
const activeCategory = ref('functionPoints')
const detailDialogVisible = ref(false)
const selectedItem = ref(null)

// 模拟验证数据
const validationData = ref({
  functionPoints: [
    {
      id: 1,
      name: '内部逻辑文件计数',
      status: 'pass',
      score: 95,
      rule: '每个内部逻辑文件应该有明确的业务目的和数据元素',
      result: '发现12个内部逻辑文件，全部符合NESMA标准',
      issues: [],
      recommendations: []
    },
    {
      id: 2,
      name: '外部输入复杂度',
      status: 'warning',
      score: 75,
      rule: '外部输入应该基于DET和FTR进行复杂度评估',
      result: '发现3个外部输入复杂度评估不准确',
      issues: [
        { id: 1, description: '用户登录功能的DET计算缺少验证字段' },
        { id: 2, description: '数据导入功能的FTR计算不完整' }
      ],
      recommendations: [
        { id: 1, suggestion: '重新计算用户登录功能的数据元素类型' },
        { id: 2, suggestion: '补充数据导入功能的文件类型引用' }
      ]
    }
  ],
  dataConsistency: [
    {
      id: 3,
      name: '数据元素一致性',
      status: 'pass',
      score: 92,
      rule: '所有数据元素应该在不同功能点中保持一致',
      result: '数据元素定义一致，无冲突',
      issues: [],
      recommendations: []
    },
    {
      id: 4,
      name: '文件引用完整性',
      status: 'fail',
      score: 45,
      rule: '所有文件引用应该有对应的数据文件定义',
      result: '发现5个文件引用缺少定义',
      issues: [
        { id: 1, description: '用户权限文件缺少定义' },
        { id: 2, description: '系统配置文件未在ILF中定义' }
      ],
      recommendations: [
        { id: 1, suggestion: '添加用户权限文件的ILF定义' },
        { id: 2, suggestion: '将系统配置文件添加到内部逻辑文件中' }
      ]
    }
  ],
  ruleCompliance: [
    {
      id: 5,
      name: 'NESMA规则合规性',
      status: 'pass',
      score: 88,
      rule: '所有功能点计算应该遵循NESMA 2.2标准',
      result: '符合NESMA 2.2标准要求',
      issues: [],
      recommendations: []
    }
  ],
  qualityCheck: [
    {
      id: 6,
      name: '评估质量检查',
      status: 'warning',
      score: 78,
      rule: '评估应该通过质量检查清单',
      result: '部分检查项需要改进',
      issues: [
        { id: 1, description: '文档完整性需要提升' }
      ],
      recommendations: [
        { id: 1, suggestion: '补充功能点计算的详细说明文档' }
      ]
    }
  ]
})

// 计算属性
const allValidations = computed(() => {
  return [
    ...validationData.value.functionPoints,
    ...validationData.value.dataConsistency,
    ...validationData.value.ruleCompliance,
    ...validationData.value.qualityCheck
  ]
})

const passedCount = computed(() => {
  return allValidations.value.filter(item => item.status === 'pass').length
})

const warningCount = computed(() => {
  return allValidations.value.filter(item => item.status === 'warning').length
})

const failedCount = computed(() => {
  return allValidations.value.filter(item => item.status === 'fail').length
})

const totalCount = computed(() => {
  return allValidations.value.length
})

const overallScore = computed(() => {
  if (allValidations.value.length === 0) return 0
  const totalScore = allValidations.value.reduce((sum, item) => sum + item.score, 0)
  return Math.round(totalScore / allValidations.value.length)
})

const validationLevel = computed(() => {
  const score = overallScore.value
  if (score >= 90) return '优秀'
  if (score >= 80) return '良好'
  if (score >= 70) return '合格'
  if (score >= 60) return '需改进'
  return '不合格'
})

const validationDescription = computed(() => {
  const score = overallScore.value
  if (score >= 90) return '评估质量优秀，符合所有验证标准'
  if (score >= 80) return '评估质量良好，大部分验证项通过'
  if (score >= 70) return '评估质量合格，部分项目需要改进'
  if (score >= 60) return '评估质量需要改进，存在多个问题'
  return '评估质量不合格，需要重新评估'
})

const functionPointValidations = computed(() => validationData.value.functionPoints)
const dataConsistencyValidations = computed(() => validationData.value.dataConsistency)
const ruleComplianceValidations = computed(() => validationData.value.ruleCompliance)
const qualityCheckValidations = computed(() => validationData.value.qualityCheck)

// 方法
const getLevelColor = (level) => {
  const colors = {
    '优秀': 'success',
    '良好': 'success',
    '合格': 'warning',
    '需改进': 'warning',
    '不合格': 'danger'
  }
  return colors[level] || ''
}

const getProgressColor = (score) => {
  if (score >= 90) return '#67c23a'
  if (score >= 80) return '#95d475'
  if (score >= 70) return '#e6a23c'
  if (score >= 60) return '#f78989'
  return '#f56c6c'
}

const getStatusColor = (status) => {
  const colors = {
    pass: 'success',
    warning: 'warning',
    fail: 'danger'
  }
  return colors[status] || ''
}

const getStatusLabel = (status) => {
  const labels = {
    pass: '通过',
    warning: '警告',
    fail: '失败'
  }
  return labels[status] || status
}

const handleCategoryChange = (tab) => {
  activeCategory.value = tab.name
}

const handleItemClick = (item) => {
  selectedItem.value = item
  detailDialogVisible.value = true
}

// 生命周期
onMounted(() => {
  // 这里可以调用API获取验证结果数据
  // loadValidationResults()
})
</script>

<style scoped>
.validation-results {
  padding: 20px;
}

.validation-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.overview-section h3,
.summary-section h3,
.detailed-results h3 {
  margin-bottom: 20px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.stat-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: #fff;
}

.stat-card.pass {
  background: #f0f9ff;
  border-color: #67c23a;
}

.stat-card.warning {
  background: #fefce8;
  border-color: #e6a23c;
}

.stat-card.fail {
  background: #fef2f2;
  border-color: #f56c6c;
}

.stat-card.total {
  background: #f8f9fa;
  border-color: #409eff;
}

.stat-value {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
}

.stat-card.pass .stat-value {
  color: #67c23a;
}

.stat-card.warning .stat-value {
  color: #e6a23c;
}

.stat-card.fail .stat-value {
  color: #f56c6c;
}

.stat-card.total .stat-value {
  color: #409eff;
}

.stat-label {
  font-size: 16px;
  color: #606266;
}

.summary-card {
  padding: 30px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: #fff;
}

.summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.summary-score {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.score-value {
  font-size: 48px;
  font-weight: bold;
  color: #409eff;
  line-height: 1;
}

.score-label {
  font-size: 14px;
  color: #909399;
  margin-top: 5px;
}

.summary-progress {
  margin-bottom: 15px;
}

.summary-description {
  color: #606266;
  font-size: 14px;
  text-align: center;
}

.validation-detail {
  max-height: 500px;
  overflow-y: auto;
}

.detail-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 10px;
  border-bottom: 1px solid #e4e7ed;
}

.detail-score {
  font-size: 16px;
  font-weight: bold;
  color: #409eff;
}

.detail-section {
  margin-bottom: 20px;
}

.detail-section h4 {
  color: #303133;
  font-size: 16px;
  margin-bottom: 10px;
}

.detail-section p {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 0;
}

.detail-section ul {
  margin: 0;
  padding-left: 20px;
}

.detail-section li {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 5px;
}
</style> 