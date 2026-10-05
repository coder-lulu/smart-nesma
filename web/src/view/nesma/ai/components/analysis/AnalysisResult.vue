<template>
  <div class="analysis-result">
    <el-card shadow="hover" class="result-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon class="header-icon"><DataAnalysis /></el-icon>
            <span class="header-title">分析结果</span>
          </div>
          <div class="header-right">
            <el-button-group size="small" class="tab-buttons">
              <el-button 
                :type="activeTab === 'overview' ? 'primary' : ''"
                @click="activeTab = 'overview'"
              >
                概览
              </el-button>
              <el-button 
                :type="activeTab === 'details' ? 'primary' : ''"
                @click="activeTab = 'details'"
              >
                详情
              </el-button>
              <el-button 
                :type="activeTab === 'recommendations' ? 'primary' : ''"
                @click="activeTab = 'recommendations'"
              >
                建议
              </el-button>
              <el-button 
                :type="activeTab === 'comparison' ? 'primary' : ''"
                @click="activeTab = 'comparison'"
              >
                对比
              </el-button>
            </el-button-group>
            
            <el-dropdown trigger="click" class="actions-dropdown">
              <el-button size="small" :icon="More">操作</el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item @click="$emit('export-result')">
                    <el-icon><Download /></el-icon>
                    导出结果
                  </el-dropdown-item>
                  <el-dropdown-item @click="$emit('save-template')">
                    <el-icon><CollectionTag /></el-icon>
                    保存模板
                  </el-dropdown-item>
                  <el-dropdown-item @click="$emit('share-result')">
                    <el-icon><Share /></el-icon>
                    分享结果
                  </el-dropdown-item>
                  <el-dropdown-item divided @click="$emit('refresh-result')">
                    <el-icon><Refresh /></el-icon>
                    刷新数据
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </template>
      
      <div class="result-content">
        <!-- 概览视图 -->
        <div v-show="activeTab === 'overview'" class="overview-tab">
          <div v-if="!hasResults" class="empty-state">
            <el-empty description="暂无分析结果">
              <el-button type="primary" @click="$emit('start-analysis')">
                开始分析
              </el-button>
            </el-empty>
          </div>
          
          <div v-else class="overview-content">
            <!-- 总体统计 -->
            <div class="stats-grid">
              <div class="stat-card">
                <div class="stat-icon success">
                  <el-icon><Check /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ overview.analyzed }}</div>
                  <div class="stat-label">已分析需求</div>
                </div>
              </div>
              
              <div class="stat-card">
                <div class="stat-icon primary">
                  <el-icon><TrendCharts /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ overview.totalFunctionPoints }}</div>
                  <div class="stat-label">总功能点</div>
                </div>
              </div>
              
              <div class="stat-card">
                <div class="stat-icon warning">
                  <el-icon><StarFilled /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ overview.avgConfidence }}%</div>
                  <div class="stat-label">平均置信度</div>
                </div>
              </div>
              
              <div class="stat-card">
                <div class="stat-icon info">
                  <el-icon><Timer /></el-icon>
                </div>
                <div class="stat-content">
                  <div class="stat-value">{{ overview.analysisTime }}</div>
                  <div class="stat-label">分析耗时</div>
                </div>
              </div>
            </div>
            
            <!-- 功能类型分布图表 -->
            <div class="chart-container">
              <h4 class="chart-title">功能类型分布</h4>
              <div ref="functionTypeChart" class="chart"></div>
            </div>
            
            <!-- 复杂度分布图表 -->
            <div class="chart-container">
              <h4 class="chart-title">复杂度分布</h4>
              <div ref="complexityChart" class="chart"></div>
            </div>
            
            <!-- 质量指标 -->
            <div class="quality-metrics">
              <h4 class="section-title">质量指标</h4>
              <div class="metrics-grid">
                <div class="metric-item">
                  <span class="metric-label">分析准确率</span>
                  <el-progress 
                    :percentage="overview.accuracy" 
                    :color="getProgressColor(overview.accuracy)"
                    :stroke-width="8"
                  />
                </div>
                <div class="metric-item">
                  <span class="metric-label">优化建议采纳率</span>
                  <el-progress 
                    :percentage="overview.adoptionRate" 
                    :color="getProgressColor(overview.adoptionRate)"
                    :stroke-width="8"
                  />
                </div>
                <div class="metric-item">
                  <span class="metric-label">需求完整性</span>
                  <el-progress 
                    :percentage="overview.completeness" 
                    :color="getProgressColor(overview.completeness)"
                    :stroke-width="8"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>
        
        <!-- 详情视图 -->
        <div v-show="activeTab === 'details'" class="details-tab">
          <div class="details-toolbar">
            <el-input
              v-model="detailsSearch"
              placeholder="搜索分析结果..."
              :prefix-icon="Search"
              clearable
              style="width: 300px;"
            />
            <el-select v-model="detailsFilter" placeholder="筛选条件" style="width: 150px;">
              <el-option label="全部" value="" />
              <el-option label="已优化" value="optimized" />
              <el-option label="需要关注" value="attention" />
              <el-option label="高置信度" value="high-confidence" />
            </el-select>
          </div>
          
          <div class="details-table">
            <el-table
              :data="filteredDetails"
              stripe
              :header-cell-style="{ background: '#f8fafc' }"
              max-height="400"
            >
              <el-table-column prop="title" label="需求标题" min-width="200">
                <template #default="{ row }">
                  <div class="requirement-info">
                    <el-tag :type="getLevelType(row.level)" size="small">
                      L{{ row.level }}
                    </el-tag>
                    <span class="title">{{ row.title }}</span>
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column prop="function_type" label="功能类型" width="100">
                <template #default="{ row }">
                  <el-tag :type="getFunctionTypeTag(row.function_type)" size="small">
                    {{ row.function_type }}
                  </el-tag>
                </template>
              </el-table-column>
              
              <el-table-column prop="complexity" label="复杂度" width="100">
                <template #default="{ row }">
                  <el-tag :type="getComplexityTag(row.complexity)" size="small">
                    {{ row.complexity }}
                  </el-tag>
                </template>
              </el-table-column>
              
              <el-table-column prop="afp" label="AFP" width="80" />
              
              <el-table-column prop="confidence" label="置信度" width="100">
                <template #default="{ row }">
                  <div class="confidence-cell">
                    <el-progress
                      :percentage="Math.round(row.confidence * 100)"
                      :stroke-width="6"
                      :show-text="false"
                      :color="getConfidenceColor(row.confidence)"
                    />
                    <span class="confidence-text">{{ Math.round(row.confidence * 100) }}%</span>
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag :type="getStatusType(row.status)" size="small">
                    {{ getStatusLabel(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              
              <el-table-column label="操作" width="150" fixed="right">
                <template #default="{ row }">
                  <el-button-group size="small">
                    <el-button text @click="viewDetail(row)" :icon="View">
                      查看
                    </el-button>
                    <el-button text @click="editResult(row)" :icon="Edit">
                      编辑
                    </el-button>
                    <el-button text @click="adoptSuggestion(row)" :icon="Check">
                      采纳
                    </el-button>
                  </el-button-group>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </div>
        
        <!-- 建议视图 -->
        <div v-show="activeTab === 'recommendations'" class="recommendations-tab">
          <div class="recommendations-list">
            <div 
              v-for="recommendation in recommendations" 
              :key="recommendation.id"
              class="recommendation-item"
              :class="recommendation.priority"
            >
              <div class="recommendation-header">
                <div class="header-left">
                  <el-tag 
                    :type="getPriorityType(recommendation.priority)" 
                    size="small"
                    class="priority-tag"
                  >
                    {{ getPriorityLabel(recommendation.priority) }}
                  </el-tag>
                  <span class="recommendation-title">{{ recommendation.title }}</span>
                </div>
                <div class="header-right">
                  <el-button-group size="small">
                    <el-button 
                      type="primary" 
                      @click="adoptRecommendation(recommendation)"
                      :icon="Check"
                    >
                      采纳
                    </el-button>
                    <el-button @click="dismissRecommendation(recommendation)" :icon="Close">
                      忽略
                    </el-button>
                  </el-button-group>
                </div>
              </div>
              
              <div class="recommendation-content">
                <p class="description">{{ recommendation.description }}</p>
                
                <div v-if="recommendation.details" class="details">
                  <h5>详细说明：</h5>
                  <ul>
                    <li v-for="detail in recommendation.details" :key="detail">
                      {{ detail }}
                    </li>
                  </ul>
                </div>
                
                <div v-if="recommendation.impact" class="impact">
                  <h5>预期影响：</h5>
                  <div class="impact-metrics">
                    <span v-if="recommendation.impact.accuracy" class="impact-item">
                      准确率提升: +{{ recommendation.impact.accuracy }}%
                    </span>
                    <span v-if="recommendation.impact.efficiency" class="impact-item">
                      效率提升: +{{ recommendation.impact.efficiency }}%
                    </span>
                    <span v-if="recommendation.impact.quality" class="impact-item">
                      质量提升: +{{ recommendation.impact.quality }}%
                    </span>
                  </div>
                </div>
                
                <div class="recommendation-meta">
                  <span class="confidence">置信度: {{ Math.round(recommendation.confidence * 100) }}%</span>
                  <span class="category">类别: {{ recommendation.category }}</span>
                </div>
              </div>
            </div>
          </div>
          
          <el-empty v-if="!recommendations.length" description="暂无优化建议" />
        </div>
        
        <!-- 对比视图 -->
        <div v-show="activeTab === 'comparison'" class="comparison-tab">
          <div class="comparison-selector">
            <el-select v-model="baselineVersion" placeholder="选择基准版本" style="width: 200px;">
              <el-option 
                v-for="version in availableVersions" 
                :key="version.id"
                :label="version.name" 
                :value="version.id"
              />
            </el-select>
            <span class="vs-text">VS</span>
            <el-select v-model="comparisonVersion" placeholder="选择对比版本" style="width: 200px;">
              <el-option 
                v-for="version in availableVersions" 
                :key="version.id"
                :label="version.name" 
                :value="version.id"
              />
            </el-select>
            <el-button type="primary" @click="generateComparison" :icon="DataAnalysis">
              生成对比
            </el-button>
          </div>
          
          <div v-if="comparisonData" class="comparison-content">
            <div class="comparison-summary">
              <div class="summary-item">
                <span class="label">需求变更:</span>
                <span class="value" :class="getChangeClass(comparisonData.requirementChanges)">
                  {{ comparisonData.requirementChanges > 0 ? '+' : '' }}{{ comparisonData.requirementChanges }}
                </span>
              </div>
              <div class="summary-item">
                <span class="label">功能点变化:</span>
                <span class="value" :class="getChangeClass(comparisonData.functionPointChanges)">
                  {{ comparisonData.functionPointChanges > 0 ? '+' : '' }}{{ comparisonData.functionPointChanges }}
                </span>
              </div>
              <div class="summary-item">
                <span class="label">准确率变化:</span>
                <span class="value" :class="getChangeClass(comparisonData.accuracyChanges)">
                  {{ comparisonData.accuracyChanges > 0 ? '+' : '' }}{{ comparisonData.accuracyChanges }}%
                </span>
              </div>
            </div>
            
            <div class="comparison-chart">
              <div ref="comparisonChart" class="chart"></div>
            </div>
          </div>
          
          <el-empty v-else description="请选择版本进行对比" />
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'
import { 
  DataAnalysis, More, Download, CollectionTag, Share, Refresh, 
  Check, TrendCharts, StarFilled, Timer, Search, View, Edit, Close 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  results: {
    type: Object,
    default: () => ({})
  },
  overview: {
    type: Object,
    default: () => ({
      analyzed: 0,
      totalFunctionPoints: 0,
      avgConfidence: 0,
      analysisTime: '0s',
      accuracy: 0,
      adoptionRate: 0,
      completeness: 0
    })
  },
  details: {
    type: Array,
    default: () => []
  },
  recommendations: {
    type: Array,
    default: () => []
  },
  availableVersions: {
    type: Array,
    default: () => []
  },
  comparisonData: {
    type: Object,
    default: null
  }
})

// Emits
const emit = defineEmits([
  'export-result',
  'save-template',
  'share-result',
  'refresh-result',
  'start-analysis',
  'view-detail',
  'edit-result',
  'adopt-suggestion',
  'adopt-recommendation',
  'dismiss-recommendation',
  'generate-comparison'
])

// 响应式数据
const activeTab = ref('overview')
const detailsSearch = ref('')
const detailsFilter = ref('')
const baselineVersion = ref(null)
const comparisonVersion = ref(null)

// 图表引用
const functionTypeChart = ref()
const complexityChart = ref()
const comparisonChart = ref()

// 计算属性
const hasResults = computed(() => {
  return props.results && Object.keys(props.results).length > 0
})

const filteredDetails = computed(() => {
  let filtered = props.details

  // 搜索过滤
  if (detailsSearch.value) {
    const query = detailsSearch.value.toLowerCase()
    filtered = filtered.filter(item => 
      item.title.toLowerCase().includes(query) ||
      item.description?.toLowerCase().includes(query)
    )
  }

  // 状态过滤
  if (detailsFilter.value) {
    switch (detailsFilter.value) {
      case 'optimized':
        filtered = filtered.filter(item => item.status === 'optimized')
        break
      case 'attention':
        filtered = filtered.filter(item => item.confidence < 0.7)
        break
      case 'high-confidence':
        filtered = filtered.filter(item => item.confidence >= 0.9)
        break
    }
  }

  return filtered
})

// 方法
const getLevelType = (level) => {
  const typeMap = { 1: 'danger', 2: 'warning', 3: 'primary', 4: 'success' }
  return typeMap[level] || 'info'
}

const getFunctionTypeTag = (type) => {
  const typeMap = {
    'EI': 'primary', 'EO': 'success', 'EQ': 'info', 
    'ILF': 'warning', 'EIF': 'danger'
  }
  return typeMap[type] || 'default'
}

const getComplexityTag = (complexity) => {
  const tagMap = {
    'Low': 'success', 'Average': 'warning', 'High': 'danger'
  }
  return tagMap[complexity] || 'info'
}

const getStatusType = (status) => {
  const statusMap = {
    'pending': 'info', 'analyzing': 'warning', 'analyzed': 'success',
    'optimized': 'primary', 'completed': 'success', 'failed': 'danger'
  }
  return statusMap[status] || 'info'
}

const getStatusLabel = (status) => {
  const labelMap = {
    'pending': '待分析', 'analyzing': '分析中', 'analyzed': '已分析',
    'optimized': '已优化', 'completed': '已完成', 'failed': '失败'
  }
  return labelMap[status] || status
}

const getPriorityType = (priority) => {
  const typeMap = {
    'high': 'danger', 'medium': 'warning', 'low': 'info'
  }
  return typeMap[priority] || 'default'
}

const getPriorityLabel = (priority) => {
  const labelMap = {
    'high': '高优先级', 'medium': '中优先级', 'low': '低优先级'
  }
  return labelMap[priority] || priority
}

const getProgressColor = (percentage) => {
  if (percentage >= 80) return '#10b981'
  if (percentage >= 60) return '#f59e0b'
  return '#ef4444'
}

const getConfidenceColor = (confidence) => {
  const percentage = confidence * 100
  if (percentage >= 80) return '#10b981'
  if (percentage >= 60) return '#f59e0b'
  return '#ef4444'
}

const getChangeClass = (value) => {
  if (value > 0) return 'positive'
  if (value < 0) return 'negative'
  return 'neutral'
}

const viewDetail = (row) => {
  emit('view-detail', row)
}

const editResult = (row) => {
  emit('edit-result', row)
}

const adoptSuggestion = (row) => {
  emit('adopt-suggestion', row)
}

const adoptRecommendation = (recommendation) => {
  emit('adopt-recommendation', recommendation)
}

const dismissRecommendation = (recommendation) => {
  emit('dismiss-recommendation', recommendation)
}

const generateComparison = () => {
  if (baselineVersion.value && comparisonVersion.value) {
    emit('generate-comparison', {
      baseline: baselineVersion.value,
      comparison: comparisonVersion.value
    })
  }
}

const initCharts = () => {
  nextTick(() => {
    if (functionTypeChart.value) {
      initFunctionTypeChart()
    }
    if (complexityChart.value) {
      initComplexityChart()
    }
    if (comparisonChart.value && props.comparisonData) {
      initComparisonChart()
    }
  })
}

const initFunctionTypeChart = () => {
  const chart = echarts.init(functionTypeChart.value)
  const option = {
    tooltip: { trigger: 'item' },
    legend: { orient: 'vertical', left: 'left' },
    series: [{
      name: '功能类型',
      type: 'pie',
      radius: '50%',
      data: [
        { value: 10, name: 'EI-外部输入' },
        { value: 8, name: 'EO-外部输出' },
        { value: 12, name: 'EQ-外部查询' },
        { value: 5, name: 'ILF-内部文件' },
        { value: 3, name: 'EIF-外部文件' }
      ],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }]
  }
  chart.setOption(option)
}

const initComplexityChart = () => {
  const chart = echarts.init(complexityChart.value)
  const option = {
    tooltip: { trigger: 'axis' },
    xAxis: {
      type: 'category',
      data: ['低复杂度', '中复杂度', '高复杂度']
    },
    yAxis: { type: 'value' },
    series: [{
      name: '需求数量',
      type: 'bar',
      data: [15, 20, 8],
      itemStyle: {
        color: (params) => {
          const colors = ['#10b981', '#f59e0b', '#ef4444']
          return colors[params.dataIndex]
        }
      }
    }]
  }
  chart.setOption(option)
}

const initComparisonChart = () => {
  if (!comparisonChart.value || !props.comparisonData) return
  
  const chart = echarts.init(comparisonChart.value)
  const option = {
    tooltip: { trigger: 'axis' },
    legend: { data: ['基准版本', '对比版本'] },
    xAxis: {
      type: 'category',
      data: ['功能点总数', '平均复杂度', '分析准确率', '完成度']
    },
    yAxis: { type: 'value' },
    series: [
      {
        name: '基准版本',
        type: 'bar',
        data: [100, 75, 85, 90]
      },
      {
        name: '对比版本',
        type: 'bar',
        data: [120, 80, 92, 95]
      }
    ]
  }
  chart.setOption(option)
}

// 监听器
watch(() => props.results, () => {
  if (hasResults.value) {
    initCharts()
  }
}, { deep: true })

watch(activeTab, (newTab) => {
  if (newTab === 'overview') {
    nextTick(() => {
      initCharts()
    })
  }
})

// 生命周期
onMounted(() => {
  if (hasResults.value) {
    initCharts()
  }
})
</script>

<style lang="scss" scoped>
.analysis-result {
  .result-card {
    border-radius: 12px;
    height: 600px;
    display: flex;
    flex-direction: column;

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;

      .header-left {
        display: flex;
        align-items: center;
        gap: 8px;

        .header-icon {
          color: #8b5cf6;
          font-size: 18px;
        }

        .header-title {
          font-weight: 600;
          color: #1f2937;
        }
      }

      .header-right {
        display: flex;
        align-items: center;
        gap: 12px;

        .tab-buttons {
          .el-button {
            font-size: 12px;
            padding: 6px 12px;
          }
        }
      }
    }

    :deep(.el-card__body) {
      flex: 1;
      overflow: hidden;
    }

    .result-content {
      height: 100%;
      overflow: auto;

      .overview-tab {
        .empty-state {
          display: flex;
          justify-content: center;
          align-items: center;
          height: 300px;
        }

        .overview-content {
          .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 16px;
            margin-bottom: 24px;

            .stat-card {
              display: flex;
              align-items: center;
              padding: 16px;
              background: #f8fafc;
              border-radius: 12px;
              border: 1px solid #e5e7eb;

              .stat-icon {
                width: 48px;
                height: 48px;
                border-radius: 12px;
                display: flex;
                align-items: center;
                justify-content: center;
                margin-right: 12px;
                color: white;
                font-size: 20px;

                &.success { background: #10b981; }
                &.primary { background: #3b82f6; }
                &.warning { background: #f59e0b; }
                &.info { background: #6b7280; }
              }

              .stat-content {
                .stat-value {
                  font-size: 24px;
                  font-weight: 700;
                  color: #1f2937;
                  margin-bottom: 4px;
                }

                .stat-label {
                  font-size: 12px;
                  color: #6b7280;
                  text-transform: uppercase;
                  font-weight: 500;
                }
              }
            }
          }

          .chart-container {
            margin-bottom: 24px;
            padding: 20px;
            background: white;
            border-radius: 12px;
            border: 1px solid #e5e7eb;

            .chart-title {
              margin: 0 0 16px 0;
              font-size: 16px;
              font-weight: 600;
              color: #1f2937;
            }

            .chart {
              height: 300px;
              width: 100%;
            }
          }

          .quality-metrics {
            .section-title {
              margin: 0 0 16px 0;
              font-size: 16px;
              font-weight: 600;
              color: #1f2937;
            }

            .metrics-grid {
              display: grid;
              gap: 16px;

              .metric-item {
                .metric-label {
                  display: block;
                  margin-bottom: 8px;
                  font-size: 14px;
                  color: #374151;
                  font-weight: 500;
                }
              }
            }
          }
        }
      }

      .details-tab {
        .details-toolbar {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 16px;
          gap: 12px;
        }

        .details-table {
          .requirement-info {
            display: flex;
            align-items: center;
            gap: 8px;

            .title {
              font-weight: 500;
            }
          }

          .confidence-cell {
            display: flex;
            align-items: center;
            gap: 8px;

            .confidence-text {
              font-size: 12px;
              font-weight: 500;
            }
          }
        }
      }

      .recommendations-tab {
        .recommendations-list {
          .recommendation-item {
            margin-bottom: 16px;
            padding: 20px;
            border-radius: 12px;
            border: 1px solid #e5e7eb;

            &.high {
              border-left: 4px solid #ef4444;
              background: #fef2f2;
            }

            &.medium {
              border-left: 4px solid #f59e0b;
              background: #fffbeb;
            }

            &.low {
              border-left: 4px solid #10b981;
              background: #f0fdf4;
            }

            .recommendation-header {
              display: flex;
              justify-content: space-between;
              align-items: center;
              margin-bottom: 12px;

              .header-left {
                display: flex;
                align-items: center;
                gap: 8px;

                .recommendation-title {
                  font-weight: 600;
                  color: #1f2937;
                }
              }
            }

            .recommendation-content {
              .description {
                margin: 0 0 12px 0;
                color: #374151;
                line-height: 1.6;
              }

              .details,
              .impact {
                margin-bottom: 12px;

                h5 {
                  margin: 0 0 8px 0;
                  font-size: 14px;
                  font-weight: 600;
                  color: #1f2937;
                }

                ul {
                  margin: 0;
                  padding-left: 20px;

                  li {
                    margin-bottom: 4px;
                    color: #374151;
                    font-size: 14px;
                  }
                }

                .impact-metrics {
                  display: flex;
                  gap: 16px;

                  .impact-item {
                    padding: 4px 8px;
                    background: #e5e7eb;
                    border-radius: 4px;
                    font-size: 12px;
                    color: #374151;
                  }
                }
              }

              .recommendation-meta {
                display: flex;
                justify-content: space-between;
                font-size: 12px;
                color: #6b7280;
              }
            }
          }
        }
      }

      .comparison-tab {
        .comparison-selector {
          display: flex;
          align-items: center;
          gap: 12px;
          margin-bottom: 24px;
          padding: 16px;
          background: #f8fafc;
          border-radius: 8px;

          .vs-text {
            font-weight: 600;
            color: #6b7280;
          }
        }

        .comparison-content {
          .comparison-summary {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
            gap: 16px;
            margin-bottom: 24px;

            .summary-item {
              padding: 12px;
              background: white;
              border-radius: 8px;
              border: 1px solid #e5e7eb;
              text-align: center;

              .label {
                display: block;
                font-size: 12px;
                color: #6b7280;
                margin-bottom: 4px;
              }

              .value {
                font-size: 18px;
                font-weight: 700;

                &.positive { color: #10b981; }
                &.negative { color: #ef4444; }
                &.neutral { color: #6b7280; }
              }
            }
          }

          .comparison-chart {
            .chart {
              height: 400px;
              width: 100%;
            }
          }
        }
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .analysis-result {
    .result-card {
      height: auto;
      min-height: 500px;

      .card-header {
        flex-direction: column;
        gap: 12px;
        align-items: flex-start;

        .header-right {
          width: 100%;
          justify-content: space-between;

          .tab-buttons {
            .el-button {
              padding: 4px 8px;
              font-size: 11px;
            }
          }
        }
      }

      .result-content {
        .overview-tab {
          .overview-content {
            .stats-grid {
              grid-template-columns: repeat(2, 1fr);

              .stat-card {
                padding: 12px;

                .stat-icon {
                  width: 40px;
                  height: 40px;
                  font-size: 16px;
                  margin-right: 8px;
                }

                .stat-content {
                  .stat-value {
                    font-size: 18px;
                  }

                  .stat-label {
                    font-size: 11px;
                  }
                }
              }
            }

            .chart-container {
              .chart {
                height: 250px;
              }
            }
          }
        }

        .details-tab {
          .details-toolbar {
            flex-direction: column;
            gap: 8px;

            .el-input,
            .el-select {
              width: 100% !important;
            }
          }
        }

        .recommendations-tab {
          .recommendations-list {
            .recommendation-item {
              padding: 16px;

              .recommendation-header {
                flex-direction: column;
                gap: 12px;
                align-items: flex-start;
              }

              .recommendation-content {
                .impact {
                  .impact-metrics {
                    flex-direction: column;
                    gap: 8px;
                  }
                }

                .recommendation-meta {
                  flex-direction: column;
                  gap: 4px;
                }
              }
            }
          }
        }

        .comparison-tab {
          .comparison-selector {
            flex-direction: column;
            gap: 8px;

            .el-select {
              width: 100% !important;
            }
          }

          .comparison-content {
            .comparison-summary {
              grid-template-columns: 1fr;
            }

            .comparison-chart {
              .chart {
                height: 300px;
              }
            }
          }
        }
      }
    }
  }
}
</style>