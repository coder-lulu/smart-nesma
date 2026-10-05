<template>
  <div class="knowledge-statistics">
    <!-- 概览统计 -->
    <el-row :gutter="20" class="overview-stats">
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ statistics?.totalEntries || 0 }}</div>
              <div class="stats-label">知识条目</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon rule">
              <el-icon><SetUp /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ statistics?.totalRules || 0 }}</div>
              <div class="stats-label">业务规则</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon case">
              <el-icon><Management /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ statistics?.totalCases || 0 }}</div>
              <div class="stats-label">案例研究</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card class="stats-card">
          <div class="stats-content">
            <div class="stats-icon usage">
              <el-icon><View /></el-icon>
            </div>
            <div class="stats-info">
              <div class="stats-number">{{ statistics?.totalUsage || 0 }}</div>
              <div class="stats-label">总使用次数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表统计 -->
    <el-row :gutter="20" class="chart-stats">
      <!-- 知识类别分布 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>知识类别分布</span>
              <el-button link @click="refreshStats">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="category-chart">
            <div v-if="!categoryStats.length" class="empty-chart">
              <el-empty description="暂无数据" />
            </div>
            <div v-else class="pie-chart">
              <div
                v-for="(item, index) in categoryStats"
                :key="item.name"
                class="pie-item"
                :style="{ '--color': categoryColors[index % categoryColors.length] }"
              >
                <div class="pie-label">
                  <span class="pie-color"></span>
                  <span class="pie-name">{{ item.name }}</span>
                </div>
                <div class="pie-value">{{ item.value }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 知识领域分布 -->
      <el-col :span="12">
        <el-card class="chart-card">
          <template #header>
            <div class="card-header">
              <span>知识领域分布</span>
            </div>
          </template>
          <div class="domain-chart">
            <div v-if="!domainStats.length" class="empty-chart">
              <el-empty description="暂无数据" />
            </div>
            <div v-else class="bar-chart">
              <div
                v-for="(item, index) in domainStats"
                :key="item.name"
                class="bar-item"
              >
                <div class="bar-label">{{ item.name }}</div>
                <div class="bar-container">
                  <div
                    class="bar-fill"
                    :style="{ 
                      width: `${(item.value / maxDomainValue) * 100}%`,
                      backgroundColor: domainColors[index % domainColors.length]
                    }"
                  ></div>
                </div>
                <div class="bar-value">{{ item.value }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 使用统计 -->
    <el-row :gutter="20" class="usage-stats">
      <!-- 热门知识 -->
      <el-col :span="12">
        <el-card class="usage-card">
          <template #header>
            <div class="card-header">
              <span>热门知识条目</span>
              <el-button link @click="getRecommendedKnowledge">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="popular-knowledge">
            <div v-if="!popularKnowledge.length" class="empty-list">
              <el-empty description="暂无数据" />
            </div>
            <div v-else class="knowledge-list">
              <div
                v-for="(item, index) in popularKnowledge"
                :key="item.id"
                class="knowledge-item"
              >
                <div class="item-rank">{{ index + 1 }}</div>
                <div class="item-content">
                  <div class="item-title">{{ item.title }}</div>
                  <div class="item-meta">
                    <span class="item-category">{{ getCategoryLabel(item.category) }}</span>
                    <span class="item-usage">{{ item.usageCount }}次使用</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>

      <!-- 热门标签 -->
      <el-col :span="12">
        <el-card class="usage-card">
          <template #header>
            <div class="card-header">
              <span>热门标签</span>
              <el-button link @click="getPopularTags">
                <el-icon><Refresh /></el-icon>
              </el-button>
            </div>
          </template>
          <div class="popular-tags">
            <div v-if="!popularTags.length" class="empty-list">
              <el-empty description="暂无数据" />
            </div>
            <div v-else class="tags-cloud">
              <el-tag
                v-for="tag in popularTags"
                :key="tag.name"
                :size="getTagSize(tag.count)"
                :type="getTagType(tag.count)"
                class="tag-item"
              >
                {{ tag.name }} ({{ tag.count }})
              </el-tag>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 时间趋势 -->
    <el-row :gutter="20" class="trend-stats">
      <el-col :span="24">
        <el-card class="trend-card">
          <template #header>
            <div class="card-header">
              <span>创建趋势</span>
              <el-radio-group v-model="trendPeriod" @change="getTrendData">
                <el-radio-button value="7d">最近7天</el-radio-button>
                <el-radio-button value="30d">最近30天</el-radio-button>
                <el-radio-button value="90d">最近90天</el-radio-button>
              </el-radio-group>
            </div>
          </template>
          <div class="trend-chart">
            <div v-if="!trendData.length" class="empty-chart">
              <el-empty description="暂无数据" />
            </div>
            <div v-else class="line-chart">
                           <div
               v-for="item in trendData"
               :key="item.date"
               class="line-item"
             >
                <div class="line-date">{{ formatTrendDate(item.date) }}</div>
                <div class="line-value">{{ item.count }}</div>
              </div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  getKnowledgeStatistics, 
  getPopularTags as getPopularTagsApi,
  getRecommendedKnowledge as getRecommendedKnowledgeApi
} from '@/api/nesma'

const props = defineProps({
  statistics: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['refresh'])

// 响应式数据
const popularKnowledge = ref([])
const popularTags = ref([])
const trendData = ref([])
const trendPeriod = ref('30d')

// 计算属性
const categoryStats = computed(() => {
  const stats = props.statistics?.categoryStats || {}
  return Object.entries(stats).map(([name, value]) => ({
    name: getCategoryLabel(name),
    value
  }))
})

const domainStats = computed(() => {
  const stats = props.statistics?.domainStats || {}
  return Object.entries(stats).map(([name, value]) => ({
    name,
    value
  })).sort((a, b) => b.value - a.value)
})

const maxDomainValue = computed(() => {
  return Math.max(...domainStats.value.map(item => item.value), 1)
})

// 颜色配置
const categoryColors = [
  '#409eff',
  '#67c23a',
  '#e6a23c',
  '#f56c6c',
  '#909399'
]

const domainColors = [
  '#409eff',
  '#67c23a',
  '#e6a23c',
  '#f56c6c',
  '#909399'
]

// 初始化
onMounted(() => {
  getRecommendedKnowledge()
  getPopularTags()
  getTrendData()
})

// 获取推荐知识
const getRecommendedKnowledge = async () => {
  try {
    const res = await getRecommendedKnowledgeApi({ limit: 10 })
    popularKnowledge.value = res.data || []
  } catch (error) {
    console.error('获取推荐知识失败:', error)
  }
}

// 获取热门标签
const getPopularTags = async () => {
  try {
    const res = await getPopularTagsApi({ limit: 20 })
    popularTags.value = res.data || []
  } catch (error) {
    console.error('获取热门标签失败:', error)
  }
}

// 获取趋势数据
const getTrendData = async () => {
  try {
    // 这里应该调用真实的API获取趋势数据
    // 暂时使用模拟数据
    const mockData = []
    const days = trendPeriod.value === '7d' ? 7 : trendPeriod.value === '30d' ? 30 : 90
    
    for (let i = days - 1; i >= 0; i--) {
      const date = new Date()
      date.setDate(date.getDate() - i)
      mockData.push({
        date: date.toISOString().split('T')[0],
        count: Math.floor(Math.random() * 10) + 1
      })
    }
    
    trendData.value = mockData
  } catch (error) {
    console.error('获取趋势数据失败:', error)
  }
}

// 刷新统计数据
const refreshStats = () => {
  emit('refresh')
  getRecommendedKnowledge()
  getPopularTags()
  getTrendData()
}

// 获取知识类别标签文本
const getCategoryLabel = (category) => {
  const labelMap = {
    'NESMA_STANDARD': 'NESMA标准',
    'BEST_PRACTICE': '最佳实践',
    'CASE_STUDY': '案例研究',
    'RULE': '业务规则'
  }
  return labelMap[category] || category
}

// 获取标签大小
const getTagSize = (count) => {
  if (count >= 10) return 'large'
  if (count >= 5) return 'default'
  return 'small'
}

// 获取标签类型
const getTagType = (count) => {
  if (count >= 10) return 'danger'
  if (count >= 5) return 'warning'
  return 'info'
}

// 格式化趋势日期
const formatTrendDate = (date) => {
  return new Date(date).toLocaleDateString('zh-CN', {
    month: '2-digit',
    day: '2-digit'
  })
}
</script>

<style scoped>
.knowledge-statistics {
  padding: 0;
}

.overview-stats {
  margin-bottom: 20px;
}

.stats-card {
  transition: all 0.3s;
}

.stats-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.stats-content {
  display: flex;
  align-items: center;
  gap: 16px;
}

.stats-icon {
  width: 48px;
  height: 48px;
  border-radius: 8px;
  background-color: #409eff;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
  font-size: 24px;
}

.stats-icon.rule {
  background-color: #67c23a;
}

.stats-icon.case {
  background-color: #e6a23c;
}

.stats-icon.usage {
  background-color: #f56c6c;
}

.stats-number {
  font-size: 24px;
  font-weight: bold;
  color: #303133;
}

.stats-label {
  font-size: 14px;
  color: #909399;
  margin-top: 4px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.chart-stats,
.usage-stats,
.trend-stats {
  margin-bottom: 20px;
}

.chart-card,
.usage-card,
.trend-card {
  height: 400px;
}

.empty-chart,
.empty-list {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 300px;
}

.pie-chart {
  padding: 20px;
}

.pie-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.pie-item:last-child {
  border-bottom: none;
}

.pie-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.pie-color {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background-color: var(--color);
}

.pie-name {
  font-size: 14px;
  color: #606266;
}

.pie-value {
  font-weight: bold;
  color: #303133;
}

.bar-chart {
  padding: 20px;
}

.bar-item {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}

.bar-label {
  width: 80px;
  font-size: 12px;
  color: #606266;
  text-align: right;
}

.bar-container {
  flex: 1;
  height: 8px;
  background-color: #f0f0f0;
  border-radius: 4px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  transition: width 0.3s;
}

.bar-value {
  width: 30px;
  font-size: 12px;
  color: #303133;
  text-align: center;
}

.knowledge-list {
  padding: 20px;
}

.knowledge-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f0f0;
}

.knowledge-item:last-child {
  border-bottom: none;
}

.item-rank {
  width: 24px;
  height: 24px;
  background-color: #409eff;
  color: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: bold;
  flex-shrink: 0;
}

.item-content {
  flex: 1;
}

.item-title {
  font-size: 14px;
  color: #303133;
  margin-bottom: 4px;
}

.item-meta {
  display: flex;
  gap: 12px;
  font-size: 12px;
  color: #909399;
}

.tags-cloud {
  padding: 20px;
}

.tag-item {
  margin: 4px;
}

.line-chart {
  padding: 20px;
  display: flex;
  gap: 20px;
  overflow-x: auto;
}

.line-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  min-width: 60px;
}

.line-date {
  font-size: 12px;
  color: #909399;
}

.line-value {
  font-size: 16px;
  font-weight: bold;
  color: #409eff;
}
</style> 