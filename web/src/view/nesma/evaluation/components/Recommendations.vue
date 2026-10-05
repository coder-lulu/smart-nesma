<template>
  <div class="recommendations">
    <div v-loading="loading" class="recommendations-content">
      <!-- 建议概览 -->
      <div class="overview-section">
        <h3>改进建议概览</h3>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="priority-card high">
              <div class="priority-value">{{ highPriorityCount }}</div>
              <div class="priority-label">高优先级</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="priority-card medium">
              <div class="priority-value">{{ mediumPriorityCount }}</div>
              <div class="priority-label">中优先级</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="priority-card low">
              <div class="priority-value">{{ lowPriorityCount }}</div>
              <div class="priority-label">低优先级</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="priority-card total">
              <div class="priority-value">{{ totalRecommendations }}</div>
              <div class="priority-label">总建议数</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 建议分类 -->
      <div class="categories-section">
        <h3>建议分类</h3>
        <el-tabs v-model="activeCategory" @tab-click="handleCategoryChange">
          <el-tab-pane label="功能点优化" name="functionPoint">
            <RecommendationCategory
              :title="'功能点优化建议'"
              :recommendations="functionPointRecommendations"
              @action-click="handleActionClick"
            />
          </el-tab-pane>

          <el-tab-pane label="数据质量" name="dataQuality">
            <RecommendationCategory
              :title="'数据质量改进'"
              :recommendations="dataQualityRecommendations"
              @action-click="handleActionClick"
            />
          </el-tab-pane>

          <el-tab-pane label="流程优化" name="processOptimization">
            <RecommendationCategory
              :title="'流程优化建议'"
              :recommendations="processOptimizationRecommendations"
              @action-click="handleActionClick"
            />
          </el-tab-pane>

          <el-tab-pane label="技术改进" name="technicalImprovement">
            <RecommendationCategory
              :title="'技术改进建议'"
              :recommendations="technicalImprovementRecommendations"
              @action-click="handleActionClick"
            />
          </el-tab-pane>
        </el-tabs>
      </div>

      <!-- 实施计划 -->
      <div class="implementation-section">
        <h3>实施计划建议</h3>
        <div class="implementation-timeline">
          <el-timeline>
            <el-timeline-item
              v-for="phase in implementationPhases"
              :key="phase.id"
              :icon="phase.icon"
              :type="phase.type"
              :timestamp="phase.timeframe"
            >
              <el-card>
                <h4>{{ phase.title }}</h4>
                <p>{{ phase.description }}</p>
                <div class="phase-recommendations">
                  <el-tag
                    v-for="rec in phase.recommendations"
                    :key="rec.id"
                    :type="getPriorityColor(rec.priority)"
                    size="small"
                    style="margin-right: 8px; margin-bottom: 4px;"
                  >
                    {{ rec.title }}
                  </el-tag>
                </div>
              </el-card>
            </el-timeline-item>
          </el-timeline>
        </div>
      </div>

      <!-- 效果预估 -->
      <div class="impact-section">
        <h3>改进效果预估</h3>
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="impact-card">
              <div class="impact-chart">
                <div ref="impactChart" style="height: 200px;"></div>
              </div>
              <div class="impact-title">质量提升预估</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="impact-card">
              <div class="impact-metrics">
                <div class="metric-item">
                  <span class="metric-label">准确度提升</span>
                  <span class="metric-value">+15%</span>
                </div>
                <div class="metric-item">
                  <span class="metric-label">效率提升</span>
                  <span class="metric-value">+25%</span>
                </div>
                <div class="metric-item">
                  <span class="metric-label">错误减少</span>
                  <span class="metric-value">-30%</span>
                </div>
              </div>
              <div class="impact-title">关键指标改进</div>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="impact-card">
              <div class="cost-benefit">
                <div class="benefit-item">
                  <span class="benefit-label">预计投入</span>
                  <span class="benefit-value cost">40工时</span>
                </div>
                <div class="benefit-item">
                  <span class="benefit-label">预计收益</span>
                  <span class="benefit-value benefit">节省120工时</span>
                </div>
                <div class="benefit-item">
                  <span class="benefit-label">投资回报</span>
                  <span class="benefit-value roi">3:1</span>
                </div>
              </div>
              <div class="impact-title">成本效益分析</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 行动项清单 -->
      <div class="action-items-section">
        <h3>行动项清单</h3>
        <div class="action-items">
          <div
            v-for="item in actionItems"
            :key="item.id"
            class="action-item"
            :class="{ completed: item.completed }"
          >
            <div class="action-header">
              <el-checkbox 
                v-model="item.completed" 
                @change="handleActionItemChange(item)"
              >
                {{ item.title }}
              </el-checkbox>
              <el-tag :type="getPriorityColor(item.priority)">
                {{ getPriorityLabel(item.priority) }}
              </el-tag>
            </div>
            <div class="action-details">
              <p>{{ item.description }}</p>
              <div class="action-meta">
                <span class="action-owner">负责人: {{ item.owner }}</span>
                <span class="action-deadline">截止时间: {{ item.deadline }}</span>
                <span class="action-effort">预计工时: {{ item.effort }}小时</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import RecommendationCategory from './RecommendationCategory.vue'

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
const activeCategory = ref('functionPoint')
const impactChart = ref(null)

// 模拟推荐数据
const recommendationsData = ref({
  functionPoint: [
    {
      id: 1,
      title: '优化功能点计算精度',
      priority: 'high',
      category: 'accuracy',
      description: '重新评估复杂度较高的功能点，提高计算精度',
      impact: '高',
      effort: 8,
      benefits: ['提高评估准确性', '减少后期调整']
    },
    {
      id: 2,
      title: '标准化功能点识别流程',
      priority: 'medium',
      category: 'process',
      description: '建立标准化的功能点识别和分类流程',
      impact: '中',
      effort: 16,
      benefits: ['提高一致性', '减少人为错误']
    }
  ],
  dataQuality: [
    {
      id: 3,
      title: '完善数据字典',
      priority: 'high',
      category: 'documentation',
      description: '补充和完善数据元素定义，建立统一的数据字典',
      impact: '高',
      effort: 12,
      benefits: ['提高数据一致性', '便于维护']
    }
  ],
  processOptimization: [
    {
      id: 4,
      title: '自动化验证流程',
      priority: 'medium',
      category: 'automation',
      description: '引入自动化工具进行基础验证，减少手工检查',
      impact: '中',
      effort: 24,
      benefits: ['提高效率', '减少错误']
    }
  ],
  technicalImprovement: [
    {
      id: 5,
      title: '升级评估工具',
      priority: 'low',
      category: 'tools',
      description: '升级现有评估工具，支持更多NESMA标准',
      impact: '低',
      effort: 40,
      benefits: ['扩展功能', '提升用户体验']
    }
  ]
})

const implementationPhases = ref([
  {
    id: 1,
    title: '第一阶段：紧急修复',
    description: '解决影响评估准确性的关键问题',
    timeframe: '1-2周',
    type: 'danger',
    icon: 'Warning',
    recommendations: [
      { id: 1, title: '修复功能点计算错误', priority: 'high' },
      { id: 3, title: '补充缺失的数据定义', priority: 'high' }
    ]
  },
  {
    id: 2,
    title: '第二阶段：流程优化',
    description: '优化评估流程，提高工作效率',
    timeframe: '3-4周',
    type: 'warning',
    icon: 'Tools',
    recommendations: [
      { id: 2, title: '标准化识别流程', priority: 'medium' },
      { id: 4, title: '引入自动化验证', priority: 'medium' }
    ]
  },
  {
    id: 3,
    title: '第三阶段：工具升级',
    description: '升级工具和技术栈，提升整体能力',
    timeframe: '5-8周',
    type: 'success',
    icon: 'TrendCharts',
    recommendations: [
      { id: 5, title: '升级评估工具', priority: 'low' }
    ]
  }
])

const actionItems = ref([
  {
    id: 1,
    title: '重新计算高复杂度功能点',
    description: '对标记为高复杂度的功能点进行重新评估和计算',
    priority: 'high',
    owner: '张三',
    deadline: '2024-01-15',
    effort: 8,
    completed: false
  },
  {
    id: 2,
    title: '建立数据字典',
    description: '创建完整的数据元素字典，统一数据定义标准',
    priority: 'high',
    owner: '李四',
    deadline: '2024-01-20',
    effort: 12,
    completed: false
  },
  {
    id: 3,
    title: '制定标准化流程文档',
    description: '编写功能点识别和分类的标准化操作流程',
    priority: 'medium',
    owner: '王五',
    deadline: '2024-01-25',
    effort: 16,
    completed: true
  }
])

// 计算属性
const allRecommendations = computed(() => {
  return [
    ...recommendationsData.value.functionPoint,
    ...recommendationsData.value.dataQuality,
    ...recommendationsData.value.processOptimization,
    ...recommendationsData.value.technicalImprovement
  ]
})

const highPriorityCount = computed(() => {
  return allRecommendations.value.filter(r => r.priority === 'high').length
})

const mediumPriorityCount = computed(() => {
  return allRecommendations.value.filter(r => r.priority === 'medium').length
})

const lowPriorityCount = computed(() => {
  return allRecommendations.value.filter(r => r.priority === 'low').length
})

const totalRecommendations = computed(() => {
  return allRecommendations.value.length
})

const functionPointRecommendations = computed(() => recommendationsData.value.functionPoint)
const dataQualityRecommendations = computed(() => recommendationsData.value.dataQuality)
const processOptimizationRecommendations = computed(() => recommendationsData.value.processOptimization)
const technicalImprovementRecommendations = computed(() => recommendationsData.value.technicalImprovement)

// 方法
const getPriorityColor = (priority) => {
  const colors = {
    high: 'danger',
    medium: 'warning',
    low: 'info'
  }
  return colors[priority] || 'info'
}

const getPriorityLabel = (priority) => {
  const labels = {
    high: '高优先级',
    medium: '中优先级',
    low: '低优先级'
  }
  return labels[priority] || priority
}

const handleCategoryChange = (tab) => {
  activeCategory.value = tab.name
}

const handleActionClick = (action) => {
  console.log('执行操作:', action)
  // 这里可以实现具体的操作逻辑
}

const handleActionItemChange = (item) => {
  console.log('行动项状态变更:', item)
  // 这里可以保存状态变更
}

const initImpactChart = () => {
  if (!impactChart.value) return
  
  const chart = echarts.init(impactChart.value)
  const option = {
    tooltip: {
      trigger: 'item'
    },
    series: [
      {
        type: 'gauge',
        startAngle: 180,
        endAngle: 0,
        center: ['50%', '75%'],
        radius: '90%',
        min: 0,
        max: 100,
        splitNumber: 8,
        axisLine: {
          lineStyle: {
            width: 6,
            color: [
              [0.25, '#f56c6c'],
              [0.5, '#e6a23c'],
              [0.75, '#409eff'],
              [1, '#67c23a']
            ]
          }
        },
        pointer: {
          icon: 'path://M12.8,0.7l12,40.1H0.7L12.8,0.7z',
          length: '12%',
          width: 20,
          offsetCenter: [0, '-60%'],
          itemStyle: {
            color: 'auto'
          }
        },
        axisTick: {
          length: 12,
          lineStyle: {
            color: 'auto',
            width: 2
          }
        },
        splitLine: {
          length: 20,
          lineStyle: {
            color: 'auto',
            width: 5
          }
        },
        axisLabel: {
          color: '#464646',
          fontSize: 20,
          distance: -60,
          formatter: function (value) {
            if (value === 87.5) {
              return '优秀'
            } else if (value === 62.5) {
              return '良好'
            } else if (value === 37.5) {
              return '一般'
            } else if (value === 12.5) {
              return '较差'
            }
            return ''
          }
        },
        title: {
          offsetCenter: [0, '-20%'],
          fontSize: 20
        },
        detail: {
          fontSize: 30,
          offsetCenter: [0, '0%'],
          valueAnimation: true,
          formatter: function (value) {
            return Math.round(value) + '%'
          },
          color: 'auto'
        },
        data: [
          {
            value: 85,
            name: '预期改进效果'
          }
        ]
      }
    ]
  }
  chart.setOption(option)
}

// 生命周期
onMounted(() => {
  nextTick(() => {
    initImpactChart()
  })
})
</script>

<style scoped>
.recommendations {
  padding: 20px;
}

.recommendations-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.overview-section h3,
.categories-section h3,
.implementation-section h3,
.impact-section h3,
.action-items-section h3 {
  margin-bottom: 20px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.priority-card {
  text-align: center;
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  background: #fff;
}

.priority-card.high {
  background: #fef2f2;
  border-color: #f56c6c;
}

.priority-card.medium {
  background: #fefce8;
  border-color: #e6a23c;
}

.priority-card.low {
  background: #f0f9ff;
  border-color: #409eff;
}

.priority-card.total {
  background: #f8f9fa;
  border-color: #67c23a;
}

.priority-value {
  font-size: 32px;
  font-weight: bold;
  margin-bottom: 8px;
}

.priority-card.high .priority-value {
  color: #f56c6c;
}

.priority-card.medium .priority-value {
  color: #e6a23c;
}

.priority-card.low .priority-value {
  color: #409eff;
}

.priority-card.total .priority-value {
  color: #67c23a;
}

.priority-label {
  font-size: 16px;
  color: #606266;
}

.implementation-timeline {
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.phase-recommendations {
  margin-top: 10px;
}

.impact-section {
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  padding: 20px;
}

.impact-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.impact-title {
  margin-top: 15px;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.impact-metrics {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.metric-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
}

.metric-label {
  color: #606266;
}

.metric-value {
  font-weight: bold;
  color: #67c23a;
}

.cost-benefit {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.benefit-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px;
  background: #fff;
  border-radius: 4px;
}

.benefit-label {
  color: #606266;
}

.benefit-value.cost {
  color: #f56c6c;
  font-weight: bold;
}

.benefit-value.benefit {
  color: #67c23a;
  font-weight: bold;
}

.benefit-value.roi {
  color: #409eff;
  font-weight: bold;
}

.action-items {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.action-item {
  padding: 20px;
  background: #fff;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
  transition: all 0.3s;
}

.action-item.completed {
  background: #f8f9fa;
  opacity: 0.7;
}

.action-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.action-details p {
  color: #606266;
  margin-bottom: 10px;
  line-height: 1.6;
}

.action-meta {
  display: flex;
  gap: 20px;
  font-size: 14px;
  color: #909399;
}
</style> 