<template>
  <div class="complexity-metrics">
    <div v-loading="loading" class="metrics-content">
      <!-- 复杂度概览 -->
      <div class="overview-section">
        <h3>复杂度概览</h3>
        <el-row :gutter="20">
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ overallComplexity }}</div>
              <div class="metric-label">总体复杂度</div>
              <div class="metric-description">基于功能点的复杂度评级</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ dataComplexity }}</div>
              <div class="metric-label">数据复杂度</div>
              <div class="metric-description">数据文件和元素复杂度</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ transactionComplexity }}</div>
              <div class="metric-label">事务复杂度</div>
              <div class="metric-description">业务交易处理复杂度</div>
            </div>
          </el-col>
          <el-col :span="6">
            <div class="metric-card">
              <div class="metric-value">{{ integrationComplexity }}</div>
              <div class="metric-label">集成复杂度</div>
              <div class="metric-description">系统集成和接口复杂度</div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 复杂度分布 -->
      <div class="distribution-section">
        <h3>复杂度分布</h3>
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="chart-container">
              <div ref="complexityChart" style="height: 300px;"></div>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="complexity-breakdown">
              <div class="breakdown-item" v-for="item in complexityBreakdown" :key="item.name">
                <div class="breakdown-header">
                  <span class="breakdown-name">{{ item.name }}</span>
                  <span class="breakdown-value">{{ item.value }}</span>
                </div>
                <el-progress
                  :percentage="item.percentage"
                  :color="item.color"
                  :stroke-width="8"
                />
              </div>
            </div>
          </el-col>
        </el-row>
      </div>

      <!-- 详细指标 -->
      <div class="detailed-metrics">
        <h3>详细指标</h3>
        <el-table :data="detailedMetrics" style="width: 100%">
          <el-table-column prop="category" label="指标类别" width="150" />
          <el-table-column prop="metric" label="指标名称" width="200" />
          <el-table-column prop="value" label="数值" width="100" align="center" />
          <el-table-column prop="weight" label="权重" width="100" align="center" />
          <el-table-column prop="score" label="得分" width="100" align="center" />
          <el-table-column prop="level" label="复杂度等级" width="120" align="center">
            <template #default="{ row }">
              <el-tag :type="getLevelColor(row.level)">{{ row.level }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="说明" />
        </el-table>
      </div>

      <!-- 复杂度趋势 -->
      <div class="trend-section">
        <h3>复杂度趋势</h3>
        <div class="trend-chart">
          <div ref="trendChart" style="height: 250px;"></div>
        </div>
      </div>

      <!-- 改进建议 -->
      <div class="improvement-section">
        <h3>复杂度优化建议</h3>
        <el-alert
          v-for="suggestion in suggestions"
          :key="suggestion.id"
          :title="suggestion.title"
          :type="suggestion.type"
          :description="suggestion.description"
          show-icon
          style="margin-bottom: 10px;"
        />
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import * as echarts from 'echarts'

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
const complexityChart = ref(null)
const trendChart = ref(null)

// 计算属性
const overallComplexity = computed(() => {
  if (!props.evaluation) return 'N/A'
  const total = props.evaluation.totalFunctionPoints || 0
  if (total < 50) return '低'
  if (total < 200) return '中'
  if (total < 500) return '高'
  return '极高'
})

const dataComplexity = computed(() => {
  // 基于数据文件数量计算
  const fileCount = props.evaluation.dataFileCount || 0
  if (fileCount < 10) return '简单'
  if (fileCount < 25) return '中等'
  return '复杂'
})

const transactionComplexity = computed(() => {
  // 基于事务数量计算
  const transactionCount = props.evaluation.transactionCount || 0
  if (transactionCount < 15) return '简单'
  if (transactionCount < 40) return '中等'
  return '复杂'
})

const integrationComplexity = computed(() => {
  // 基于接口数量计算
  const interfaceCount = props.evaluation.interfaceCount || 0
  if (interfaceCount < 5) return '简单'
  if (interfaceCount < 15) return '中等'
  return '复杂'
})

const complexityBreakdown = computed(() => {
  const total = props.evaluation.totalFunctionPoints || 100
  return [
    {
      name: '数据功能',
      value: Math.round(total * 0.4),
      percentage: 40,
      color: '#409eff'
    },
    {
      name: '事务功能',
      value: Math.round(total * 0.45),
      percentage: 45,
      color: '#67c23a'
    },
    {
      name: '接口功能',
      value: Math.round(total * 0.15),
      percentage: 15,
      color: '#e6a23c'
    }
  ]
})

const detailedMetrics = computed(() => {
  return [
    {
      category: '数据复杂度',
      metric: '内部逻辑文件',
      value: props.evaluation.ilfCount || 0,
      weight: 0.3,
      score: 85,
      level: '中等',
      description: '应用程序维护的逻辑文件数量'
    },
    {
      category: '数据复杂度',
      metric: '外部接口文件',
      value: props.evaluation.eifCount || 0,
      weight: 0.2,
      score: 75,
      level: '简单',
      description: '只读引用的外部文件数量'
    },
    {
      category: '事务复杂度',
      metric: '外部输入',
      value: props.evaluation.eiCount || 0,
      weight: 0.25,
      score: 90,
      level: '高',
      description: '来自外部的数据输入事务'
    },
    {
      category: '事务复杂度',
      metric: '外部输出',
      value: props.evaluation.eoCount || 0,
      weight: 0.25,
      score: 80,
      level: '中等',
      description: '向外部发送的数据输出事务'
    },
    {
      category: '事务复杂度',
      metric: '外部查询',
      value: props.evaluation.eqCount || 0,
      weight: 0.2,
      score: 70,
      level: '简单',
      description: '输入输出组合的查询事务'
    }
  ]
})

const suggestions = computed(() => {
  const suggestions = []
  
  if (overallComplexity.value === '极高') {
    suggestions.push({
      id: 1,
      title: '总体复杂度过高',
      type: 'error',
      description: '建议拆分大型功能模块，降低系统整体复杂度'
    })
  }
  
  if (dataComplexity.value === '复杂') {
    suggestions.push({
      id: 2,
      title: '数据复杂度偏高',
      type: 'warning',
      description: '建议优化数据结构设计，减少冗余数据文件'
    })
  }
  
  if (transactionComplexity.value === '复杂') {
    suggestions.push({
      id: 3,
      title: '事务复杂度较高',
      type: 'warning',
      description: '建议简化业务流程，减少不必要的事务处理'
    })
  }
  
  if (integrationComplexity.value === '复杂') {
    suggestions.push({
      id: 4,
      title: '集成复杂度过高',
      type: 'error',
      description: '建议统一接口标准，减少系统间的耦合度'
    })
  }
  
  return suggestions
})

// 方法
const getLevelColor = (level) => {
  const colors = {
    '简单': 'success',
    '中等': 'warning',
    '高': 'danger',
    '复杂': 'danger'
  }
  return colors[level] || ''
}

const initComplexityChart = () => {
  if (!complexityChart.value) return
  
  const chart = echarts.init(complexityChart.value)
  const option = {
    title: {
      text: '复杂度分布',
      left: 'center'
    },
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    legend: {
      orient: 'vertical',
      left: 'left'
    },
    series: [
      {
        name: '复杂度分布',
        type: 'pie',
        radius: ['40%', '70%'],
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
        data: complexityBreakdown.value.map(item => ({
          value: item.value,
          name: item.name,
          itemStyle: { color: item.color }
        }))
      }
    ]
  }
  chart.setOption(option)
}

const initTrendChart = () => {
  if (!trendChart.value) return
  
  const chart = echarts.init(trendChart.value)
  const option = {
    title: {
      text: '复杂度趋势',
      left: 'center'
    },
    tooltip: {
      trigger: 'axis'
    },
    legend: {
      data: ['总体复杂度', '数据复杂度', '事务复杂度']
    },
    xAxis: {
      type: 'category',
      data: ['第1周', '第2周', '第3周', '第4周', '第5周', '第6周']
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '总体复杂度',
        type: 'line',
        data: [120, 132, 101, 134, 90, 130]
      },
      {
        name: '数据复杂度',
        type: 'line',
        data: [220, 182, 191, 234, 290, 330]
      },
      {
        name: '事务复杂度',
        type: 'line',
        data: [150, 232, 201, 154, 190, 330]
      }
    ]
  }
  chart.setOption(option)
}

// 生命周期
onMounted(() => {
  nextTick(() => {
    initComplexityChart()
    initTrendChart()
  })
})
</script>

<style scoped>
.complexity-metrics {
  padding: 20px;
}

.metrics-content {
  display: flex;
  flex-direction: column;
  gap: 30px;
}

.overview-section h3,
.distribution-section h3,
.detailed-metrics h3,
.trend-section h3,
.improvement-section h3 {
  margin-bottom: 20px;
  color: #303133;
  font-size: 18px;
  font-weight: 600;
}

.metric-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.metric-value {
  font-size: 28px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.metric-label {
  font-size: 16px;
  color: #303133;
  margin-bottom: 4px;
}

.metric-description {
  font-size: 12px;
  color: #909399;
}

.chart-container {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
}

.complexity-breakdown {
  padding: 20px;
}

.breakdown-item {
  margin-bottom: 20px;
}

.breakdown-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.breakdown-name {
  font-size: 14px;
  color: #303133;
}

.breakdown-value {
  font-size: 14px;
  font-weight: bold;
  color: #409eff;
}

.trend-chart {
  border: 1px solid #e4e7ed;
  border-radius: 8px;
  padding: 20px;
}
</style> 