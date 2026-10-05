<template>
  <el-dialog
    v-model="dialogVisible"
    title="项目详情"
    width="1200px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <div v-loading="loading" class="project-detail">
      <div v-if="projectDetail" class="detail-content">
        <!-- 项目基本信息 -->
        <div class="info-section">
          <h3 class="section-title">基本信息</h3>
          <el-row :gutter="20">
            <el-col :span="12">
              <div class="info-item">
                <label>项目名称：</label>
                <span class="value">{{ projectDetail.name }}</span>
              </div>
              <div class="info-item">
                <label>项目ID：</label>
                <span class="value">{{ projectDetail.ID }}</span>
              </div>
              <div class="info-item">
                <label>项目领域：</label>
                <el-tag size="small">{{ getDomainLabel(projectDetail.domain) }}</el-tag>
              </div>
              <div class="info-item">
                <label>项目状态：</label>
                <el-tag :type="getStatusType(projectDetail.status)">
                  {{ getStatusLabel(projectDetail.status) }}
                </el-tag>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="info-item">
                <label>项目负责人ID：</label>
                <span class="value">{{ projectDetail.ownerId || '未设置' }}</span>
              </div>
              <div class="info-item">
                <label>当前周期：</label>
                <span class="value">{{ getActiveCycleName() }}</span>
              </div>
              <div class="info-item">
                <label>领域标签：</label>
                <div class="tag-list">
                  <el-tag 
                    v-for="tag in projectDetail.domainTags" 
                    :key="tag" 
                    size="small" 
                    style="margin-right: 4px;"
                  >
                    {{ tag }}
                  </el-tag>
                  <span v-if="!projectDetail.domainTags || projectDetail.domainTags.length === 0">暂无标签</span>
                </div>
              </div>
              <div class="info-item">
                <label>活跃周期ID：</label>
                <span class="value">{{ projectDetail.activeCycleId || '未设置' }}</span>
              </div>
            </el-col>
          </el-row>
        </div>

        <!-- 时间信息 -->
        <div class="info-section">
          <h3 class="section-title">时间信息</h3>
          <el-row :gutter="20">
            <el-col :span="12">
              <div class="info-item">
                <label>开始时间：</label>
                <span class="value">{{ formatDate(projectDetail.startDate) }}</span>
              </div>
              <div class="info-item">
                <label>结束时间：</label>
                <span class="value">{{ formatDate(projectDetail.endDate) }}</span>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="info-item">
                <label>创建时间：</label>
                <span class="value">{{ formatDate(projectDetail.CreatedAt) }}</span>
              </div>
              <div class="info-item">
                <label>更新时间：</label>
                <span class="value">{{ formatDate(projectDetail.UpdatedAt) }}</span>
              </div>
            </el-col>
          </el-row>
        </div>

        <!-- 项目周期信息 -->
        <div v-if="projectDetail.cycles && projectDetail.cycles.length > 0" class="info-section">
          <h3 class="section-title">项目周期</h3>
          <el-row :gutter="20">
            <el-col :span="24">
              <div class="cycle-list">
                <div 
                  v-for="cycle in projectDetail.cycles" 
                  :key="cycle.ID" 
                  class="cycle-item"
                  :class="{ 'active-cycle': cycle.ID === projectDetail.activeCycleId }"
                >
                  <div class="cycle-header">
                    <el-tag 
                      :type="cycle.ID === projectDetail.activeCycleId ? 'success' : 'info'" 
                      size="small"
                    >
                      {{ cycle.name }}
                    </el-tag>
                    <el-tag v-if="cycle.ID === projectDetail.activeCycleId" type="warning" size="small">
                      当前活跃
                    </el-tag>
                  </div>
                  <div class="cycle-content">
                    <p><strong>描述：</strong>{{ cycle.description || '暂无描述' }}</p>
                    <p><strong>状态：</strong>{{ getCycleStatusLabel(cycle.status) }}</p>
                    <p><strong>阶段：</strong>{{ getCyclePhaseLabel(cycle.phase) }}</p>
                    <p><strong>开始时间：</strong>{{ formatDate(cycle.startDate) }}</p>
                    <p><strong>结束时间：</strong>{{ formatDate(cycle.endDate) || '未设置' }}</p>
                    <p><strong>需求统计：</strong>总计 {{ cycle.requirementCount || 0 }} 个，完成 {{ cycle.completedCount || 0 }} 个</p>
                    <p><strong>进度：</strong>{{ cycle.progress || 0 }}%</p>
                  </div>
                </div>
              </div>
            </el-col>
          </el-row>
        </div>

        <!-- 需求统计 -->
        <div v-if="requirementStats" class="info-section">
          <h3 class="section-title">需求统计</h3>
          <el-row :gutter="20">
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.totalCount || 0 }}</div>
                  <div class="stat-label">需求总数</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.level1 || 0 }}</div>
                  <div class="stat-label">一级模块</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.level2 || 0 }}</div>
                  <div class="stat-label">二级模块</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.level4 || 0 }}</div>
                  <div class="stat-label">功能点</div>
                </div>
              </el-card>
            </el-col>
          </el-row>
          
          <el-row :gutter="20" style="margin-top: 16px;">
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.completed || 0 }}</div>
                  <div class="stat-label">已完成</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.inProgress || 0 }}</div>
                  <div class="stat-label">进行中</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ requirementStats.pending || 0 }}</div>
                  <div class="stat-label">待处理</div>
                </div>
              </el-card>
            </el-col>
            <el-col :span="6">
              <el-card class="stat-card">
                <div class="stat-content">
                  <div class="stat-number">{{ getCompletionRate() }}%</div>
                  <div class="stat-label">完成率</div>
                </div>
              </el-card>
            </el-col>
          </el-row>
        </div>

        <!-- 项目描述 -->
        <div class="info-section">
          <h3 class="section-title">项目描述</h3>
          <div class="description-content">
            <p>{{ projectDetail.description || '暂无描述' }}</p>
          </div>
        </div>

        <!-- 项目进度 -->
        <div class="info-section">
          <h3 class="section-title">项目进度</h3>
          <div class="progress-info">
            <div class="progress-item">
              <label>整体进度：</label>
              <el-progress 
                :percentage="getCompletionRate()" 
                :stroke-width="8"
                :color="getProgressColor()"
              />
            </div>
            <div class="progress-item">
              <label>剩余时间：</label>
              <span class="value">{{ getRemainingDays() }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">关闭</el-button>
        <el-button type="primary" @click="handleEdit">编辑</el-button>
        <el-button type="success" @click="handleViewRequirements">查看需求</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getNesmaProject } from '@/api/nesma'
import { formatDate } from '@/utils/format'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  project: {
    type: Object,
    default: () => ({})
  }
})

const emit = defineEmits(['update:modelValue', 'edit', 'viewRequirements'])

const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

const loading = ref(false)
const projectDetail = ref(null)
const requirementStats = ref(null)

// 监听对话框打开
watch(dialogVisible, (newVal) => {
  if (newVal && props.project && props.project.ID) {
    loadProjectDetail(props.project.ID)
  }
})

// 监听project变化  
watch(() => props.project, (newProject) => {
  if (newProject && newProject.ID && dialogVisible.value) {
    loadProjectDetail(newProject.ID)
  }
}, { immediate: false, deep: true })

// 加载项目详情
const loadProjectDetail = async (projectId) => {
  try {
    loading.value = true
    const response = await getNesmaProject(projectId)
    
    // 根据实际API响应格式处理
    if (response.code === 0) {
      // API返回的数据结构是 {project, requirementStats}
      projectDetail.value = response.data.project
      requirementStats.value = response.data.requirementStats
    } else {
      ElMessage.error(response.msg || '获取项目详情失败')
      // 发生错误时使用列表数据作为降级显示
      projectDetail.value = props.project
    }
  } catch (error) {
    console.error('加载项目详情失败:', error)
    ElMessage.error('加载项目详情失败: ' + error.message)
    // 发生错误时使用列表数据作为降级显示
    projectDetail.value = props.project
  } finally {
    loading.value = false
  }
}


// 获取领域标签
const getDomainLabel = (domain) => {
  const labels = {
    'web': 'Web应用',
    'mobile': '移动应用', 
    'desktop': '桌面应用',
    'data': '数据分析',
    'ai': '人工智能',
    'enterprise': '企业系统',
    'game': '游戏开发',
    'iot': '物联网'
  }
  return labels[domain] || '未知'
}

// 获取状态类型
const getStatusType = (status) => {
  const types = {
    'active': 'success',
    'paused': 'warning',
    'completed': 'success',
    'archived': 'info'
  }
  return types[status] || 'info'
}

// 获取状态标签
const getStatusLabel = (status) => {
  const labels = {
    'active': '活跃',
    'paused': '暂停',
    'completed': '完成',
    'archived': '归档'
  }
  return labels[status] || '未知'
}


// 获取活跃周期名称
const getActiveCycleName = () => {
  if (!projectDetail.value || !projectDetail.value.activeCycle) return '未设置'
  return projectDetail.value.activeCycle.name || '未设置'
}

// 获取周期状态标签
const getCycleStatusLabel = (status) => {
  const labels = {
    'planning': '规划中',
    'active': '进行中',
    'paused': '暂停',
    'completed': '已完成',
    'archived': '已归档'
  }
  return labels[status] || status
}

// 获取周期阶段标签
const getCyclePhaseLabel = (phase) => {
  const labels = {
    'requirement': '需求阶段',
    'design': '设计阶段',
    'development': '开发阶段',
    'testing': '测试阶段',
    'deployment': '部署阶段',
    'maintenance': '维护阶段'
  }
  return labels[phase] || phase
}

// 获取完成率
const getCompletionRate = () => {
  if (!requirementStats.value || !requirementStats.value.totalCount) return 0
  const completed = requirementStats.value.completed || 0
  const total = requirementStats.value.totalCount
  return Math.round((completed / total) * 100)
}

// 获取进度条颜色
const getProgressColor = () => {
  const rate = getCompletionRate()
  if (rate < 30) return '#f56c6c'
  if (rate < 70) return '#e6a23c'
  return '#67c23a'
}

// 获取剩余天数
const getRemainingDays = () => {
  if (!projectDetail.value || !projectDetail.value.endDate) return '未设置'
  
  const endDate = new Date(projectDetail.value.endDate)
  const today = new Date()
  const diffTime = endDate - today
  const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24))
  
  if (diffDays < 0) {
    return `已逾期 ${Math.abs(diffDays)} 天`
  } else if (diffDays === 0) {
    return '今天截止'
  } else {
    return `${diffDays} 天`
  }
}

// 编辑项目
const handleEdit = () => {
  emit('edit', projectDetail.value)
  handleClose()
}

// 查看需求
const handleViewRequirements = () => {
  emit('viewRequirements', projectDetail.value)
  handleClose()
}

// 关闭对话框
const handleClose = () => {
  projectDetail.value = null
  requirementStats.value = null
  emit('update:modelValue', false)
}
</script>

<style scoped>
.project-detail {
  max-height: 600px;
  overflow-y: auto;
}

.detail-content {
  padding: 0 8px;
}

.info-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 8px;
  border-bottom: 2px solid #e4e7ed;
}

.info-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.info-item label {
  min-width: 100px;
  font-weight: 500;
  color: #606266;
}

.info-item .value {
  color: #303133;
  word-break: break-all;
}

.description-content {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 6px;
  border-left: 4px solid #409eff;
}

.description-content p {
  margin: 0;
  line-height: 1.6;
  color: #303133;
  white-space: pre-wrap;
}

.technology-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.stat-card {
  text-align: center;
  height: 80px;
}

.stat-content {
  display: flex;
  flex-direction: column;
  justify-content: center;
  height: 100%;
}

.stat-number {
  font-size: 24px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #909399;
}

.progress-info {
  background: #f8f9fa;
  padding: 16px;
  border-radius: 6px;
}

.progress-item {
  display: flex;
  align-items: center;
  margin-bottom: 12px;
}

.progress-item:last-child {
  margin-bottom: 0;
}

.progress-item label {
  min-width: 80px;
  font-weight: 500;
  color: #606266;
}

.progress-item .value {
  color: #303133;
  font-weight: 500;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-progress-bar__outer) {
  background-color: #e4e7ed;
}

/* 周期相关样式 */
.cycle-list .cycle-item {
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 16px;
  margin-bottom: 16px;
  background: #fafafa;
}
.cycle-list .cycle-item.active-cycle {
  border-color: #67c23a;
  background: #f0f9ff;
}
.cycle-list .cycle-item:last-child {
  margin-bottom: 0;
}
.cycle-list .cycle-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 12px;
}
.cycle-list .cycle-content p {
  margin: 4px 0;
  font-size: 13px;
  line-height: 1.4;
}
.cycle-list .cycle-content p strong {
  color: #606266;
  font-weight: 500;
}

.tag-list {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
}
</style> 