<template>
  <div class="nesma-main">
    <!-- 主导航头部 -->
    <div class="nesma-header">
      <div class="header-content">
        <div class="header-left">
          <h1 class="nesma-title">
            <el-icon><TrendCharts /></el-icon>
            Smart-NESMA
          </h1>
          <p class="nesma-subtitle">智能化功能点分析系统</p>
        </div>
        <div class="header-right">
          <el-button type="primary" @click="navigateToWorkspace">
            <el-icon><House /></el-icon>
            智能工作台
          </el-button>
          <el-button type="info" @click="navigateToStyleTest" size="small">
            <el-icon><Setting /></el-icon>
            样式测试
          </el-button>
        </div>
      </div>
    </div>

    <!-- 功能模块导航 -->
    <div class="module-navigation">
      <div class="nav-container">
        <h2 class="nav-title">功能模块</h2>
        <div class="module-grid">
          <!-- 项目管理模块 -->
          <div class="module-card primary" @click="navigateTo('/nesma/project')">
            <div class="module-icon">
              <el-icon><FolderOpened /></el-icon>
            </div>
            <div class="module-content">
              <h3>项目管理</h3>
              <p>创建和管理NESMA项目，配置多阶段建设周期，支持项目全生命周期管理</p>
              <div class="module-features">
                <span class="feature-tag">项目创建</span>
                <span class="feature-tag">周期管理</span>
                <span class="feature-tag">权限控制</span>
              </div>
            </div>
            <div class="module-arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>

          <!-- 需求管理模块 -->
          <div class="module-card success" @click="navigateTo('/nesma/requirement')">
            <div class="module-icon">
              <el-icon><Document /></el-icon>
            </div>
            <div class="module-content">
              <h3>需求管理</h3>
              <p>支持4级层次化需求管理，Excel批量导入，需求树状结构可视化编辑</p>
              <div class="module-features">
                <span class="feature-tag">层次管理</span>
                <span class="feature-tag">Excel导入</span>
                <span class="feature-tag">版本控制</span>
              </div>
            </div>
            <div class="module-arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>

          <!-- AI智能分析模块 -->
          <div class="module-card warning" @click="navigateTo('/nesma/ai')">
            <div class="module-icon">
              <el-icon><MagicStick /></el-icon>
            </div>
            <div class="module-content">
              <h3>AI智能分析</h3>
              <p>基于AI驱动的需求优化分析，智能对话助手，自动生成需求描述和流程图</p>
              <div class="module-features">
                <span class="feature-tag">智能对话</span>
                <span class="feature-tag">需求分析</span>
                <span class="feature-tag">自动优化</span>
              </div>
            </div>
            <div class="module-arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>

          <!-- NESMA评估模块 -->
          <div class="module-card info" @click="navigateTo('/nesma/evaluation')">
            <div class="module-icon">
              <el-icon><Checked /></el-icon>
            </div>
            <div class="module-content">
              <h3>NESMA评估</h3>
              <p>标准化功能点评估，EI/EO/EQ/ILF/EIF分类，复杂度分析，合规性检查</p>
              <div class="module-features">
                <span class="feature-tag">功能点分析</span>
                <span class="feature-tag">复杂度评估</span>
                <span class="feature-tag">标准合规</span>
              </div>
            </div>
            <div class="module-arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>

          <!-- 文档生成模块 -->
          <div class="module-card danger" @click="navigateTo('/nesma/document')">
            <div class="module-icon">
              <el-icon><Files /></el-icon>
            </div>
            <div class="module-content">
              <h3>文档生成</h3>
              <p>自动生成需求规格书、评估报告，支持Word/Excel/PDF多格式导出</p>
              <div class="module-features">
                <span class="feature-tag">模板管理</span>
                <span class="feature-tag">批量生成</span>
                <span class="feature-tag">多格式导出</span>
              </div>
            </div>
            <div class="module-arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>

          <!-- 知识库模块 -->
          <div class="module-card" @click="navigateTo('/nesma/knowledge')">
            <div class="module-icon">
              <el-icon><Reading /></el-icon>
            </div>
            <div class="module-content">
              <h3>知识库</h3>
              <p>NESMA标准知识库，最佳实践案例，智能知识检索和推荐系统</p>
              <div class="module-features">
                <span class="feature-tag">标准知识</span>
                <span class="feature-tag">案例库</span>
                <span class="feature-tag">智能检索</span>
              </div>
            </div>
            <div class="module-arrow">
              <el-icon><ArrowRight /></el-icon>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 系统概览 -->
    <div class="system-overview">
      <div class="overview-container">
        <h2 class="overview-title">系统概览</h2>
        <el-row :gutter="30">
          <!-- 快速统计 -->
          <el-col :span="12">
            <div class="overview-card">
              <h3>快速统计</h3>
              <div class="stats-grid">
                <div class="stat-item">
                  <div class="stat-value">{{ systemStats.totalProjects }}</div>
                  <div class="stat-label">项目总数</div>
                </div>
                <div class="stat-item">
                  <div class="stat-value">{{ systemStats.totalRequirements }}</div>
                  <div class="stat-label">需求总数</div>
                </div>
                <div class="stat-item">
                  <div class="stat-value">{{ systemStats.completedAnalysis }}</div>
                  <div class="stat-label">已完成分析</div>
                </div>
                <div class="stat-item">
                  <div class="stat-value">{{ systemStats.totalAFP }}</div>
                  <div class="stat-label">功能点总数</div>
                </div>
              </div>
            </div>
          </el-col>

          <!-- 系统状态 -->
          <el-col :span="12">
            <div class="overview-card">
              <h3>系统状态</h3>
              <div class="status-items">
                <div class="status-item">
                  <el-icon class="status-icon success"><Select /></el-icon>
                  <span class="status-text">AI服务正常</span>
                </div>
                <div class="status-item">
                  <el-icon class="status-icon success"><Select /></el-icon>
                  <span class="status-text">数据库连接正常</span>
                </div>
                <div class="status-item">
                  <el-icon class="status-icon success"><Select /></el-icon>
                  <span class="status-text">知识库可用</span>
                </div>
                <div class="status-item">
                  <el-icon class="status-icon warning"><Warning /></el-icon>
                  <span class="status-text">缓存使用率较高</span>
                </div>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>
    </div>

    <!-- 底部链接 -->
    <div class="footer-links">
      <div class="links-container">
        <el-link type="primary" @click="navigateTo('/nesma/monitor/dashboard')">
          <el-icon><Monitor /></el-icon>
          系统监控
        </el-link>
        <el-link type="primary" @click="navigateTo('/nesma/ai/chat')">
          <el-icon><ChatSquare /></el-icon>
          AI助手
        </el-link>
        <el-link type="primary" href="https://www.nesma.org/" target="_blank">
          <el-icon><Link /></el-icon>
          NESMA官网
        </el-link>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import {
  TrendCharts, House, DataAnalysis, FolderOpened, Document, MagicStick, Checked,
  Files, Reading, ArrowRight, Select, Warning, Monitor, ChatSquare, Link, Setting
} from '@element-plus/icons-vue'
import { getNesmaProjectStats, getNesmaRequirementStats } from '@/api/nesma'

const router = useRouter()

// 系统统计数据
const systemStats = reactive({
  totalProjects: 0,
  totalRequirements: 0,
  completedAnalysis: 0,
  totalAFP: 0
})

// 导航到指定路由
const navigateTo = (path) => {
  router.push(path)
}

// 导航到工作台
const navigateToWorkspace = () => {
  router.push('/nesma-workspace')
}

// 导航到样式测试页面
const navigateToStyleTest = () => {
  router.push('/test/style-isolation')
}

// 加载系统统计数据
const loadSystemStats = async () => {
  try {
    const [projectStats, requirementStats] = await Promise.all([
      getNesmaProjectStats(),
      getNesmaRequirementStats({})
    ])
    
    systemStats.totalProjects = projectStats.data?.total || 0
    systemStats.totalRequirements = requirementStats.data?.total || 0
    systemStats.completedAnalysis = requirementStats.data?.analyzed || 0
    systemStats.totalAFP = requirementStats.data?.totalAFP || 0
  } catch (error) {
    console.error('加载系统统计数据失败:', error)
  }
}

onMounted(() => {
  loadSystemStats()
})
</script>

<style lang="scss" scoped>
.nesma-main {
  min-height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}

.nesma-header {
  padding: 40px 20px;
  color: white;

  .header-content {
    max-width: 1200px;
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
  }

  .nesma-title {
    display: flex;
    align-items: center;
    font-size: 36px;
    font-weight: 700;
    margin: 0 0 8px 0;

    .el-icon {
      margin-right: 16px;
      font-size: 40px;
    }
  }

  .nesma-subtitle {
    font-size: 18px;
    margin: 0;
    opacity: 0.9;
  }
}

.module-navigation {
  padding: 40px 20px;
  background: white;

  .nav-container {
    max-width: 1200px;
    margin: 0 auto;
  }

  .nav-title {
    font-size: 28px;
    font-weight: 600;
    color: #2c3e50;
    margin-bottom: 30px;
    text-align: center;
  }

  .module-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(350px, 1fr));
    gap: 24px;
  }

  .module-card {
    display: flex;
    align-items: center;
    padding: 24px;
    background: white;
    border-radius: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
    cursor: pointer;
    transition: all 0.3s ease;
    border-left: 4px solid #e1e6ea;

    &:hover {
      transform: translateY(-4px);
      box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
    }

    &.primary {
      border-left-color: #409eff;
      
      .module-icon {
        background: linear-gradient(135deg, #409eff, #66b1ff);
      }
    }

    &.success {
      border-left-color: #67c23a;
      
      .module-icon {
        background: linear-gradient(135deg, #67c23a, #85ce61);
      }
    }

    &.warning {
      border-left-color: #e6a23c;
      
      .module-icon {
        background: linear-gradient(135deg, #e6a23c, #ebb563);
      }
    }

    &.info {
      border-left-color: #909399;
      
      .module-icon {
        background: linear-gradient(135deg, #909399, #a6a9ad);
      }
    }

    &.danger {
      border-left-color: #f56c6c;
      
      .module-icon {
        background: linear-gradient(135deg, #f56c6c, #f78989);
      }
    }

    .module-icon {
      width: 60px;
      height: 60px;
      background: linear-gradient(135deg, #667eea, #764ba2);
      border-radius: 12px;
      display: flex;
      align-items: center;
      justify-content: center;
      margin-right: 20px;

      .el-icon {
        font-size: 28px;
        color: white;
      }
    }

    .module-content {
      flex: 1;

      h3 {
        font-size: 20px;
        font-weight: 600;
        color: #2c3e50;
        margin: 0 0 8px 0;
      }

      p {
        font-size: 14px;
        color: #7f8c8d;
        line-height: 1.5;
        margin: 0 0 12px 0;
      }

      .module-features {
        display: flex;
        flex-wrap: wrap;
        gap: 6px;

        .feature-tag {
          background: #f8f9fa;
          color: #6c757d;
          padding: 2px 8px;
          border-radius: 4px;
          font-size: 12px;
          font-weight: 500;
        }
      }
    }

    .module-arrow {
      .el-icon {
        font-size: 20px;
        color: #bdc3c7;
      }
    }
  }
}

.system-overview {
  padding: 40px 20px;
  background: #f8f9fa;

  .overview-container {
    max-width: 1200px;
    margin: 0 auto;
  }

  .overview-title {
    font-size: 28px;
    font-weight: 600;
    color: #2c3e50;
    margin-bottom: 30px;
    text-align: center;
  }

  .overview-card {
    background: white;
    padding: 30px;
    border-radius: 12px;
    box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);

    h3 {
      font-size: 20px;
      font-weight: 600;
      color: #2c3e50;
      margin: 0 0 20px 0;
    }
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;

    .stat-item {
      text-align: center;

      .stat-value {
        font-size: 32px;
        font-weight: 700;
        color: #667eea;
        margin-bottom: 8px;
      }

      .stat-label {
        font-size: 14px;
        color: #7f8c8d;
      }
    }
  }

  .status-items {
    .status-item {
      display: flex;
      align-items: center;
      margin-bottom: 16px;

      .status-icon {
        margin-right: 12px;
        font-size: 16px;

        &.success {
          color: #67c23a;
        }

        &.warning {
          color: #e6a23c;
        }
      }

      .status-text {
        font-size: 14px;
        color: #2c3e50;
      }
    }
  }
}

.footer-links {
  padding: 30px 20px;
  background: white;

  .links-container {
    max-width: 1200px;
    margin: 0 auto;
    text-align: center;

    .el-link {
      margin: 0 20px;
      font-size: 16px;

      .el-icon {
        margin-right: 6px;
      }
    }
  }
}

@media (max-width: 768px) {
  .nesma-header {
    .header-content {
      flex-direction: column;
      text-align: center;

      .header-right {
        margin-top: 20px;
      }
    }
  }

  .module-navigation {
    .module-grid {
      grid-template-columns: 1fr;
    }
  }

  .system-overview {
    .stats-grid {
      grid-template-columns: 1fr;
    }
  }

  .footer-links {
    .el-link {
      display: block;
      margin: 8px 0;
    }
  }
}
</style>