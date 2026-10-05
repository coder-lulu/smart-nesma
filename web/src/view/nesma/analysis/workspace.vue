<template>
  <div class="workspace-container">
    <!-- 顶部标题区域 -->
    <div class="page-header">
      <div class="header-content">
        <div class="header-title">
          <el-icon class="title-icon"><DataAnalysis /></el-icon>
          <span>NESMA智能分析工作台</span>
        </div>
        <div class="header-status">
          <div class="status-indicator">
            <div class="status-dot"></div>
            <span>{{ systemStatus.text }}</span>
          </div>
          <span class="current-time">{{ currentTime }}</span>
        </div>
      </div>
    </div>

    <!-- 核心统计面板 -->
    <div class="stats-section">
      <div class="stats-grid">
        <div class="stat-card" @click="router.push('/layout/nesma/project')">
          <div class="stat-icon primary">
            <el-icon><Folder /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ projectStats.total }}</div>
            <div class="stat-label">总项目数</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon success">
            <el-icon><DataLine /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ functionPointStats.total ?? '—' }}</div>
            <div class="stat-label">功能点总数</div>
          </div>
        </div>

        <div class="stat-card" @click="router.push('/layout/nesma/requirement')">
          <div class="stat-icon warning">
            <el-icon><Document /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ requirementStats.total ?? '—' }}</div>
            <div class="stat-label">需求条目</div>
          </div>
        </div>

        <div class="stat-card">
          <div class="stat-icon info">
            <el-icon><Timer /></el-icon>
          </div>
          <div class="stat-content">
            <div class="stat-number">{{ analysisStats.running ?? '—' }}</div>
            <div class="stat-label">运行中任务</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 主要工作区域 -->
    <div class="main-content">
      <el-row :gutter="24">
        <!-- 左侧：核心功能 -->
        <el-col :span="12">
          <div class="content-card">
            <div class="card-header">
              <h3 class="card-title">
                <el-icon><MagicStick /></el-icon>
                核心功能
              </h3>
            </div>
            <div class="function-grid">
              <div class="function-item primary" @click="handleCreateProject">
                <div class="function-icon">
                  <el-icon><Plus /></el-icon>
                </div>
                <div class="function-info">
                  <div class="function-name">新建项目</div>
                  <div class="function-desc">创建新的NESMA分析项目</div>
                </div>
              </div>

              <div class="function-item success" @click="handleImportRequirements">
                <div class="function-icon">
                  <el-icon><Upload /></el-icon>
                </div>
                <div class="function-info">
                  <div class="function-name">导入需求</div>
                  <div class="function-desc">批量导入项目需求文档</div>
                </div>
              </div>

              <div class="function-item warning">
                <div class="function-icon">
                  <el-icon><Cpu /></el-icon>
                </div>
                <div class="function-info">
                  <div class="function-name">AI智能分析</div>
                  <div class="function-desc">使用AI模型进行深度分析</div>
                </div>
              </div>

              <div class="function-item info" @click="router.push('/layout/nesma/knowledge')">
                <div class="function-icon">
                  <el-icon><Reading /></el-icon>
                </div>
                <div class="function-info">
                  <div class="function-name">知识库</div>
                  <div class="function-desc">NESMA标准知识库</div>
                </div>
              </div>
            </div>
          </div>
        </el-col>

        <!-- 右侧：最近项目 -->
        <el-col :span="12">
          <div class="content-card">
            <div class="card-header">
              <h3 class="card-title">
                <el-icon><Clock /></el-icon>
                最近项目
              </h3>
              <el-button type="primary" text @click="router.push('/layout/nesma/project')">
                查看全部
              </el-button>
            </div>
            
            <div class="project-list">
              <div 
                v-for="project in recentProjects" 
                :key="project.id"
                class="project-item"
                @click="handleProjectClick(project)"
              >
                <div class="project-status" :class="getProjectStatusClass(project.status)"></div>
                <div class="project-content">
                  <div class="project-header">
                    <span class="project-name">{{ project.name }}</span>
                    <el-tag :type="getProjectStatusType(project.status)" size="small">
                      {{ getProjectStatusText(project.status) }}
                    </el-tag>
                  </div>
                  <div class="project-meta">
                    <span>需求: {{ project.requirementCount || 0 }}</span>
                    <span>功能点: {{ project.functionPoints || 0 }}</span>
                  </div>
                  <div class="project-time">{{ formatTime(project.updatedAt) }}</div>
                </div>
              </div>
              
              <el-empty v-if="!recentProjects.length" description="暂无最近项目" :image-size="60" />
            </div>
          </div>
        </el-col>
      </el-row>

      <!-- 底部：系统状态 -->
      <div class="system-status-section">
        <div class="content-card">
          <div class="card-header">
            <h3 class="card-title">
              <el-icon><Bell /></el-icon>
              系统状态
            </h3>
          </div>
          <div class="status-list">
            <div v-for="activity in systemActivities" :key="activity.id" class="status-item">
              <div class="status-icon" :class="activity.type">
                <el-icon>
                  <SuccessFilled v-if="activity.type === 'success'" />
                  <WarningFilled v-if="activity.type === 'warning'" />
                  <InfoFilled v-if="activity.type === 'info'" />
                </el-icon>
              </div>
              <div class="status-content">
                <div class="status-message">{{ activity.message }}</div>
                <div class="status-time">{{ formatTime(activity.time) }}</div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  DataAnalysis, Folder, DataLine, Document, Timer, Plus, Upload, MagicStick, 
  Clock, Bell, SuccessFilled, WarningFilled, InfoFilled, Reading, Cpu
} from '@element-plus/icons-vue'
import { 
  getDashboardStats,
  getSystemActivities,
  getNesmaProjectList as getProjectList
} from '@/api/nesma'

const router = useRouter()

// 响应式数据
const currentTime = ref('')
const systemStatus = ref({
  text: '系统运行正常'
})

// 统计数据
const projectStats = ref({
  total: 0
})

const functionPointStats = ref({
  total: null
})

const requirementStats = ref({
  total: null
})

const analysisStats = ref({
  running: null
})

// 项目和活动数据
const recentProjects = ref([])
const systemActivities = ref([])

// 生命周期
onMounted(async () => {
  updateTime()
  await Promise.all([
    loadProjectStats(),
    loadRecentProjects(),
    loadSystemActivities()
  ])
  
  // 定时更新时间
  const timeInterval = setInterval(updateTime, 1000)
  
  onUnmounted(() => {
    clearInterval(timeInterval)
  })
})

// 方法
const updateTime = () => {
  currentTime.value = new Date().toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

const loadProjectStats = async () => {
  try {
    const response = await getDashboardStats({
      time_range: 'all',
      include_cache: true
    })
    
    if (response.data) {
      const data = response.data
      
      // 项目统计
      projectStats.value = {
        total: data.total_projects || 0
      }
      
      // 当前统计接口未提供功能点、需求和运行中任务总数，保留空值展示。

      // 系统状态
      if (data.system_status) {
        systemStatus.value = {
          text: data.system_status.message || '系统运行正常'
        }
      }
    }
  } catch (error) {
    console.error('加载工作台统计失败:', error)
    ElMessage.error('加载工作台数据失败，请稍后重试')
  }
}

const loadRecentProjects = async () => {
  try {

    const response = await getProjectList({
      page: 1,
      pageSize: 5,
      orderBy: 'updated_at DESC'
    })
    
    if (response.data && response.data.list) {
      console.log(response.data.list)
      recentProjects.value = response.data.list.map(project => ({
        ...project,
        requirementCount: project.requirementStats.totalCount || 0,
        functionPoints: project.requirementStats.level4 || 0
      }))
    }
  } catch (error) {
    console.error('加载最近项目失败:', error)
  }
}

const loadSystemActivities = async () => {
  try {
    const response = await getSystemActivities({
      page: 1,
      page_size: 5,
      days: 7
    })
    
    if (response.data && response.data.list) {
      systemActivities.value = response.data.list.map(activity => ({
        id: activity.id,
        type: activity.type,
        message: activity.message || activity.title,
        time: new Date(activity.created_at)
      }))
    }
  } catch (error) {
    console.error('加载系统活动失败:', error)
    // 如果API失败，使用默认数据
    systemActivities.value = []
  }
}



// 事件处理
const handleCreateProject = () => {
  router.push('/layout/nesma/project?action=create')
}

const handleImportRequirements = () => {
  router.push('/layout/nesma/requirement?action=import')
}

const handleAIAnalysis = () => {
  ElMessageBox.confirm(
    '是否启动AI智能分析？这将使用最新的DeepSeek-Reasoner模型进行深度分析。',
    'AI智能分析',
    {
      confirmButtonText: '启动分析',
      cancelButtonText: '取消',
      type: 'info'
    }
  ).then(() => {
    router.push('/layout/nesma/ai/analysis')
    ElMessage.success('AI分析已启动')
  }).catch(() => {})
}

const handleProjectClick = (project) => {
  router.push(`/layout/nesma/project/${project.id}`)
}

// 辅助方法
const getProjectStatusType = (status) => {
  const typeMap = {
    'active': 'success',
    'pending': 'warning', 
    'completed': 'info',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

const getProjectStatusText = (status) => {
  const textMap = {
    'active': '进行中',
    'pending': '待开始',
    'completed': '已完成',
    'cancelled': '已取消'
  }
  return textMap[status] || status
}

const getProjectStatusClass = (status) => {
  const classMap = {
    'active': 'success',
    'pending': 'warning',
    'completed': 'info'
  }
  return classMap[status] || 'info'
}

const formatTime = (time) => {
  if (!time) return ''
  
  const now = new Date()
  const target = new Date(time)
  const diff = now - target
  
  if (diff < 1000 * 60) return '刚刚'
  if (diff < 1000 * 60 * 60) return `${Math.floor(diff / (1000 * 60))} 分钟前`
  if (diff < 1000 * 60 * 60 * 24) return `${Math.floor(diff / (1000 * 60 * 60))} 小时前`
  
  return target.toLocaleDateString('zh-CN')
}
</script>

<style lang="scss" scoped>
.workspace-container {
  padding: 20px;
  background-color: var(--el-bg-color-page);
  min-height: 100vh;
}

// 顶部标题区域
.page-header {
  margin-bottom: 24px;
  background: var(--el-bg-color);
  border-radius: 12px;
  padding: 20px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
  
  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    
    .header-title {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 26px;
      font-weight: 700;
      color: var(--el-text-color-primary);
      
      .title-icon {
        font-size: 32px;
        color: var(--el-color-primary);
      }
    }
    
    .header-status {
      display: flex;
      align-items: center;
      gap: 24px;
      
      .status-indicator {
        display: flex;
        align-items: center;
        gap: 6px;
        padding: 6px 12px;
        border-radius: 20px;
        background: var(--el-color-success-light-9);
        color: var(--el-color-success);
        font-size: 13px;
        font-weight: 500;
        
        .status-dot {
          width: 8px;
          height: 8px;
          border-radius: 50%;
          background: var(--el-color-success);
          animation: pulse 2s infinite;
        }
      }
      
      .current-time {
        font-family: 'Courier New', monospace;
        color: var(--el-text-color-regular);
        font-size: 14px;
        font-weight: 500;
      }
    }
  }
}

// 核心统计面板
.stats-section {
  margin-bottom: 20px;
  
  .stats-grid {
    display: grid;
    gap: 12px;
    grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  }
  
  .stat-card {
    padding: 16px;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    background: var(--el-bg-color);
    border: 1px solid var(--el-border-color-lighter);
    box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04);
    display: flex;
    align-items: center;
    gap: 12px;
    
    &:hover {
      transform: translateY(-1px);
      box-shadow: 0 3px 12px rgba(0, 0, 0, 0.08);
      border-color: var(--el-color-primary-light-7);
    }
    
    .stat-icon {
      width: 40px;
      height: 40px;
      border-radius: 8px;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 16px;
      color: white;
      flex-shrink: 0;
      
      &.primary { background: var(--el-color-primary); }
      &.success { background: var(--el-color-success); }
      &.warning { background: var(--el-color-warning); }
      &.info { background: var(--el-color-info); }
    }
    
    .stat-content {
      flex: 1;
      min-width: 0;
      
      .stat-number {
        font-size: 24px;
        font-weight: 700;
        margin-bottom: 2px;
        color: var(--el-text-color-primary);
        line-height: 1;
      }
      
      .stat-label {
        font-size: 12px;
        color: var(--el-text-color-regular);
        font-weight: 500;
        white-space: nowrap;
        overflow: hidden;
        text-overflow: ellipsis;
      }
    }
  }
}

// 主要工作区域
.main-content {
  .content-card {
    background: var(--el-bg-color);
    border-radius: 12px;
    padding: 24px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    border: 1px solid var(--el-border-color-lighter);
    margin-bottom: 24px;
    
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 20px;
      
      .card-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 18px;
        font-weight: 600;
        color: var(--el-text-color-primary);
        margin: 0;
        
        .el-icon {
          color: var(--el-color-primary);
          font-size: 20px;
        }
      }
    }
  }
}

// 核心功能
.function-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
  
  .function-item {
    display: flex;
    align-items: center;
    gap: 16px;
    padding: 20px 24px;
    border-radius: 12px;
    background: var(--el-bg-color-page);
    border: 1px solid var(--el-border-color-lighter);
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    
    &:hover {
      transform: translateY(-3px);
      box-shadow: 0 8px 24px rgba(0, 0, 0, 0.12);
    }
    
    &.primary {
      border-color: var(--el-color-primary-light-7);
      background: var(--el-color-primary-light-9);
      
      &:hover {
        background: var(--el-color-primary-light-8);
        border-color: var(--el-color-primary);
      }
      
      .function-icon {
        color: var(--el-color-primary);
      }
    }
    
    &.success {
      border-color: var(--el-color-success-light-7);
      background: var(--el-color-success-light-9);
      
      &:hover {
        background: var(--el-color-success-light-8);
        border-color: var(--el-color-success);
      }
      
      .function-icon {
        color: var(--el-color-success);
      }
    }
    
    &.warning {
      border-color: var(--el-color-warning-light-7);
      background: var(--el-color-warning-light-9);
      
      &:hover {
        background: var(--el-color-warning-light-8);
        border-color: var(--el-color-warning);
      }
      
      .function-icon {
        color: var(--el-color-warning);
      }
    }
    
    &.info {
      border-color: var(--el-color-info-light-7);
      background: var(--el-color-info-light-9);
      
      &:hover {
        background: var(--el-color-info-light-8);
        border-color: var(--el-color-info);
      }
      
      .function-icon {
        color: var(--el-color-info);
      }
    }
    
    .function-icon {
      font-size: 28px;
    }
    
    .function-info {
      flex: 1;
      
      .function-name {
        font-size: 16px;
        font-weight: 600;
        color: var(--el-text-color-primary);
        margin-bottom: 4px;
      }
      
      .function-desc {
        font-size: 13px;
        color: var(--el-text-color-regular);
        line-height: 1.4;
      }
    }
  }
}

// 最近项目
.project-list {
  .project-item {
    display: flex;
    gap: 12px;
    padding: 16px;
    border-radius: 12px;
    cursor: pointer;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    border: 1px solid var(--el-border-color-lighter);
    margin-bottom: 12px;
    background: var(--el-bg-color-page);
    
    &:hover {
      background: var(--el-color-primary-light-9);
      border-color: var(--el-color-primary-light-7);
      transform: translateY(-2px);
      box-shadow: 0 4px 12px rgba(64, 158, 255, 0.15);
    }
    
    &:last-child {
      margin-bottom: 0;
    }
    
    .project-status {
      width: 8px;
      height: 8px;
      border-radius: 50%;
      margin-top: 6px;
      flex-shrink: 0;
      
      &.success { background: var(--el-color-success); }
      &.warning { background: var(--el-color-warning); }
      &.info { background: var(--el-color-info); }
    }
    
    .project-content {
      flex: 1;
      
      .project-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;
        
        .project-name {
          font-weight: 600;
          color: var(--el-text-color-primary);
          font-size: 15px;
        }
      }
      
      .project-meta {
        display: flex;
        gap: 16px;
        margin-bottom: 8px;
        font-size: 13px;
        color: var(--el-text-color-regular);
        
        span {
          display: flex;
          align-items: center;
          gap: 4px;
        }
      }
      
      .project-time {
        font-size: 12px;
        color: var(--el-text-color-secondary);
        font-weight: 500;
      }
    }
  }
}

// 系统状态
.system-status-section {
  margin-top: 24px;
  
  .status-list {
    display: flex;
    flex-direction: column;
    gap: 12px;
  }
  
  .status-item {
    display: flex;
    gap: 12px;
    padding: 16px;
    border-radius: 12px;
    transition: all 0.3s ease;
    background: var(--el-bg-color-page);
    border: 1px solid var(--el-border-color-lighter);
    
    &:hover {
      background: var(--el-color-primary-light-9);
      border-color: var(--el-color-primary-light-7);
      transform: translateY(-1px);
    }
    
    .status-icon {
      width: 40px;
      height: 40px;
      border-radius: 50%;
      display: flex;
      align-items: center;
      justify-content: center;
      font-size: 18px;
      flex-shrink: 0;
      
      &.success {
        background: var(--el-color-success-light-9);
        color: var(--el-color-success);
      }
      
      &.warning {
        background: var(--el-color-warning-light-9);
        color: var(--el-color-warning);
      }
      
      &.info {
        background: var(--el-color-info-light-9);
        color: var(--el-color-info);
      }
    }
    
    .status-content {
      flex: 1;
      
      .status-message {
        font-size: 14px;
        font-weight: 500;
        color: var(--el-text-color-primary);
        margin-bottom: 6px;
        line-height: 1.4;
      }
      
      .status-time {
        font-size: 12px;
        color: var(--el-text-color-regular);
        font-weight: 500;
      }
    }
  }
}

// 对话框样式
.project-option {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.form-hint {
  font-size: 12px;
  color: var(--el-text-color-regular);
  margin-top: 5px;
}

.batch-preview {
  margin-top: 20px;
  
  h4 {
    font-size: 16px;
    font-weight: 600;
    color: var(--el-text-color-primary);
    margin-bottom: 15px;
  }
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

// 动画
@keyframes pulse {
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.5;
    transform: scale(1.1);
  }
}

// 响应式
@media (max-width: 768px) {
  .workspace-container {
    padding: 12px;
  }
  
  .page-header {
    padding: 16px;
    margin-bottom: 16px;
    
    .header-content {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;
      
      .header-title {
        font-size: 22px;
        
        .title-icon {
          font-size: 28px;
        }
      }
      
      .header-status {
        gap: 16px;
      }
    }
  }
  
  .stats-section {
    margin-bottom: 12px;
    
    .stats-grid {
      grid-template-columns: repeat(2, 1fr);
      gap: 8px;
    }
    
    .stat-card {
      padding: 12px;
      gap: 10px;
      
      .stat-content .stat-number {
        font-size: 20px;
      }
      
      .stat-icon {
        width: 32px;
        height: 32px;
        font-size: 14px;
      }
      
      .stat-content .stat-label {
        font-size: 11px;
      }
    }
  }
  
  .main-content {
    .el-col {
      margin-bottom: 16px;
    }
    
    .content-card {
      padding: 16px;
      margin-bottom: 16px;
    }
    
    .card-header .card-title {
      font-size: 16px;
    }
  }
  
  .function-grid {
    grid-template-columns: 1fr;
    gap: 12px;
    
    .function-item {
      padding: 16px 20px;
      gap: 12px;
      
      .function-icon {
        font-size: 24px;
      }
      
      .function-info {
        .function-name {
          font-size: 15px;
        }
        
        .function-desc {
          font-size: 12px;
        }
      }
    }
  }
  
  .project-item {
    padding: 12px !important;
    gap: 8px !important;
  }
  
  .status-item {
    padding: 12px !important;
    gap: 8px !important;
    
    .status-icon {
      width: 32px !important;
      height: 32px !important;
      font-size: 14px !important;
    }
  }
}
</style>