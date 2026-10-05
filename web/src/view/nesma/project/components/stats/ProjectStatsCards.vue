<template>
  <div class="project-stats-cards">
    <el-row :gutter="20">
      <!-- 总项目数 -->
      <el-col :span="6">
        <el-card class="stats-card total" :body-style="{ padding: '20px' }">
          <div class="stats-content">
            <div class="stats-icon">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ stats.totalProjects || 0 }}</div>
              <div class="stats-label">总项目数</div>
              <div class="stats-trend" v-if="stats.totalProjectsTrend">
                <el-icon class="trend-icon" :class="getTrendClass(stats.totalProjectsTrend)">
                  <component :is="getTrendIcon(stats.totalProjectsTrend)" />
                </el-icon>
                <span class="trend-text">{{ Math.abs(stats.totalProjectsTrend) }}%</span>
              </div>
            </div>
          </div>
          <div class="stats-action">
            <el-button text type="primary" @click="$emit('view-all')">
              查看详情
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 活跃项目 -->
      <el-col :span="6">
        <el-card class="stats-card active" :body-style="{ padding: '20px' }">
          <div class="stats-content">
            <div class="stats-icon">
              <el-icon><VideoPlay /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ activeProjects }}</div>
              <div class="stats-label">活跃项目</div>
            </div>
          </div>
          <div class="stats-action">
            <el-button text type="primary" @click="$emit('filter-active')">
              查看活跃
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 已完成项目 -->
      <el-col :span="6">
        <el-card class="stats-card completed" :body-style="{ padding: '20px' }">
          <div class="stats-content">
            <div class="stats-icon">
              <el-icon><Check /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ completedProjects }}</div>
              <div class="stats-label">已归档</div>
            </div>
          </div>
          <div class="stats-action">
            <el-button text type="primary" @click="$emit('filter-completed')">
              查看归档
            </el-button>
          </div>
        </el-card>
      </el-col>

      <!-- 近期新增 -->
      <el-col :span="6">
        <el-card class="stats-card recent" :body-style="{ padding: '20px' }">
          <div class="stats-content">
            <div class="stats-icon">
              <el-icon><Calendar /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ stats.recentProjects || 0 }}</div>
              <div class="stats-label">近30天新增</div>
            </div>
          </div>
          <div class="stats-action">
            <el-button text type="primary" @click="$emit('filter-recent')">
              查看新增
            </el-button>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 详细统计面板 -->
    <div v-if="showDetails" class="detailed-stats">
      <el-row :gutter="20">
        <!-- 状态分布图表 -->
        <el-col :span="8">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">
                <h4>项目状态分布</h4>
                <el-button text @click="refreshCharts">
                  <el-icon><Refresh /></el-icon>
                </el-button>
              </div>
            </template>
            <div ref="statusChart" class="chart-container"></div>
          </el-card>
        </el-col>

        <!-- 领域分布图表 -->
        <el-col :span="8">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">
                <h4>项目领域分布</h4>
              </div>
            </template>
            <div ref="domainChart" class="chart-container"></div>
          </el-card>
        </el-col>

        <!-- 月度趋势图表 -->
        <el-col :span="8">
          <el-card class="chart-card">
            <template #header>
              <div class="chart-header">
                <h4>项目创建趋势</h4>
                <el-radio-group v-model="trendPeriod" size="small">
                  <el-radio-button value="6m">6个月</el-radio-button>
                  <el-radio-button value="1y">1年</el-radio-button>
                </el-radio-group>
              </div>
            </template>
            <div ref="trendChart" class="chart-container"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 展开/收起按钮 -->
    <div class="toggle-details">
      <el-button 
        text 
        type="primary" 
        @click="toggleDetails"
        class="toggle-btn"
      >
        <el-icon><component :is="showDetails ? 'ArrowUp' : 'ArrowDown'" /></el-icon>
        {{ showDetails ? '收起详细统计' : '展开详细统计' }}
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import { 
  FolderOpened, VideoPlay, Check, Calendar, Refresh,
  ArrowUp, ArrowDown, TrendCharts, CaretTop, CaretBottom
} from '@element-plus/icons-vue'
import * as echarts from 'echarts'

// Props
const props = defineProps({
  stats: {
    type: Object,
    default: () => ({
      totalProjects: 0,
      activeProjects: 0,
      completedProjects: 0,
      recentProjects: 0,
      pausedProjects: 0,
      archivedProjects: 0,
      totalProjectsTrend: 0,
      domainDistribution: {},
      monthlyTrend: []
    })
  },
  loading: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits([
  'view-all',
  'filter-active', 
  'filter-completed',
  'filter-recent',
  'refresh'
])

// 响应式数据
const showDetails = ref(false)
const trendPeriod = ref('6m')
const statusChart = ref(null)
const domainChart = ref(null)
const trendChart = ref(null)

// 计算属性
const activeProjects = computed(() => {
  return props.stats.activeProjects || 0
})

const completedProjects = computed(() => {
  // 如果有completed状态，显示completed，否则显示archived（归档）
  return props.stats.completedProjects || props.stats.archivedProjects || 0
})

const getActiveRate = () => {
  const total = props.stats.totalProjects || 0
  if (total === 0) return 0
  return Math.round((activeProjects.value / total) * 100)
}

const getCompletionRate = () => {
  const total = props.stats.totalProjects || 0
  if (total === 0) return 0
  return Math.round((completedProjects.value / total) * 100)
}

const getRecentChange = () => {
  const trend = props.stats.recentProjectsTrend || 0
  if (trend > 0) return `增长${trend}%`
  if (trend < 0) return `下降${Math.abs(trend)}%`
  return '持平'
}

const getTrendClass = (trend) => {
  if (trend > 0) return 'trend-up'
  if (trend < 0) return 'trend-down'
  return 'trend-stable'
}

const getTrendIcon = (trend) => {
  if (trend > 0) return CaretTop
  if (trend < 0) return CaretBottom
  return TrendCharts
}

// 方法
const toggleDetails = () => {
  showDetails.value = !showDetails.value
  if (showDetails.value) {
    nextTick(() => {
      initCharts()
    })
  }
}

const refreshCharts = () => {
  emit('refresh')
  nextTick(() => {
    initCharts()
  })
}

const initCharts = () => {
  initStatusChart()
  initDomainChart()
  initTrendChart()
}

const initStatusChart = () => {
  if (!statusChart.value) return
  
  const chart = echarts.init(statusChart.value)
  const option = {
    tooltip: {
      trigger: 'item',
      formatter: '{a} <br/>{b}: {c} ({d}%)'
    },
    series: [
      {
        name: '项目状态',
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
            fontSize: '18',
            fontWeight: 'bold'
          }
        },
        labelLine: {
          show: false
        },
        data: [
          { 
            value: activeProjects.value, 
            name: '活跃',
            itemStyle: { color: '#67c23a' }
          },
          { 
            value: completedProjects.value, 
            name: '归档',
            itemStyle: { color: '#409eff' }
          },
          { 
            value: props.stats.pausedProjects || 0, 
            name: '暂停',
            itemStyle: { color: '#e6a23c' }
          }
        ]
      }
    ]
  }
  
  chart.setOption(option)
  window.addEventListener('resize', () => chart.resize())
}

const initDomainChart = () => {
  if (!domainChart.value) return
  
  const chart = echarts.init(domainChart.value)
  const domainData = props.stats.domainDistribution || {}
  
  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'shadow'
      }
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: Object.keys(domainData),
      axisTick: {
        alignWithLabel: true
      }
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '项目数量',
        type: 'bar',
        barWidth: '60%',
        data: Object.values(domainData),
        itemStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: '#83bff6' },
            { offset: 0.5, color: '#188df0' },
            { offset: 1, color: '#188df0' }
          ])
        }
      }
    ]
  }
  
  chart.setOption(option)
  window.addEventListener('resize', () => chart.resize())
}

const initTrendChart = () => {
  if (!trendChart.value) return
  
  const chart = echarts.init(trendChart.value)
  const trendData = props.stats.monthlyTrend || []
  
  const option = {
    tooltip: {
      trigger: 'axis'
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '3%',
      containLabel: true
    },
    xAxis: {
      type: 'category',
      data: trendData.map(item => item.month),
      boundaryGap: false
    },
    yAxis: {
      type: 'value'
    },
    series: [
      {
        name: '新增项目',
        type: 'line',
        stack: 'Total',
        data: trendData.map(item => item.count),
        smooth: true,
        areaStyle: {
          opacity: 0.3
        },
        itemStyle: {
          color: '#409eff'
        }
      }
    ]
  }
  
  chart.setOption(option)
  window.addEventListener('resize', () => chart.resize())
}

// 监听趋势期间变化
watch(trendPeriod, () => {
  nextTick(() => {
    initTrendChart()
  })
})

// 监听统计数据变化
watch(() => props.stats, () => {
  if (showDetails.value) {
    nextTick(() => {
      initCharts()
    })
  }
}, { deep: true })
</script>

<style lang="scss" scoped>
.project-stats-cards {
  margin-bottom: 20px;

  .stats-card {
    height: 140px;
    transition: all 0.3s ease;
    border: 1px solid #e4e7ed;

    &:hover {
      transform: translateY(-2px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    }

    .stats-content {
      display: flex;
      align-items: center;
      margin-bottom: 12px;

      .stats-icon {
        width: 60px;
        height: 60px;
        border-radius: 12px;
        display: flex;
        align-items: center;
        justify-content: center;
        margin-right: 16px;

        .el-icon {
          font-size: 28px;
          color: white;
        }
      }

      .stats-info {
        flex: 1;

        .stats-number {
          font-size: 32px;
          font-weight: 700;
          color: #303133;
          line-height: 1;
          margin-bottom: 4px;
        }

        .stats-label {
          font-size: 14px;
          color: #909399;
          margin-bottom: 4px;
        }

        .stats-trend {
          display: flex;
          align-items: center;
          font-size: 12px;

          .trend-icon {
            margin-right: 4px;

            &.trend-up { color: #67c23a; }
            &.trend-down { color: #f56c6c; }
            &.trend-stable { color: #909399; }
          }

          .trend-text {
            color: #606266;
          }
        }

        .stats-detail {
          .detail-text {
            font-size: 12px;
            color: #909399;
          }
        }
      }
    }

    .stats-action {
      text-align: right;
    }

    // 不同卡片的主题色
    &.total .stats-icon { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
    &.active .stats-icon { background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%); }
    &.completed .stats-icon { background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%); }
    &.recent .stats-icon { background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%); }
  }

  .detailed-stats {
    margin: 20px 0;

    .chart-card {
      .chart-header {
        display: flex;
        justify-content: space-between;
        align-items: center;

        h4 {
          margin: 0;
          font-size: 16px;
          font-weight: 600;
          color: #303133;
        }
      }

      .chart-container {
        height: 200px;
        width: 100%;
      }
    }
  }

  .toggle-details {
    text-align: center;
    margin-top: 16px;

    .toggle-btn {
      padding: 8px 16px;
      font-size: 14px;
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .project-stats-cards {
    .stats-card {
      .stats-content {
        .stats-icon {
          width: 50px;
          height: 50px;

          .el-icon {
            font-size: 24px;
          }
        }

        .stats-info {
          .stats-number {
            font-size: 28px;
          }
        }
      }
    }
  }
}

@media (max-width: 768px) {
  .project-stats-cards {
    .el-col {
      margin-bottom: 12px;
    }

    .stats-card {
      height: auto;

      .stats-content {
        flex-direction: column;
        text-align: center;

        .stats-icon {
          margin-right: 0;
          margin-bottom: 8px;
        }
      }
    }
  }
}
</style>