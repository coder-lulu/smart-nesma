<template>
  <div class="analysis-header">
    <el-card shadow="never" class="header-card">
      <div class="header-content">
        <div class="title-section">
          <h2 class="analysis-title">
            <el-icon class="title-icon"><MagicStick /></el-icon>
            {{ title }}
          </h2>
          <p class="analysis-subtitle">{{ subtitle }}</p>
        </div>
        
        <div class="action-section">
          <el-button-group class="main-actions">
            <el-button 
              type="primary" 
              size="large"
              @click="$emit('start-analysis')"
              :loading="isAnalyzing"
              :disabled="!canStartAnalysis"
              class="primary-action"
            >
              <el-icon><Lightning /></el-icon>
              {{ isAnalyzing ? '分析中...' : '开始智能分析' }}
            </el-button>
            
            <el-button 
              size="large"
              @click="$emit('show-config')"
              :disabled="isAnalyzing"
            >
              <el-icon><Setting /></el-icon>
              分析配置
            </el-button>
            
            <el-button 
              size="large"
              @click="$emit('export-report')"
              :disabled="!hasResults"
            >
              <el-icon><Download /></el-icon>
              导出报告
            </el-button>
          </el-button-group>
          
          <!-- 高级操作 -->
          <el-dropdown 
            v-if="showAdvancedActions"
            trigger="click"
            class="advanced-dropdown"
          >
            <el-button size="large" :icon="More">
              更多操作
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="$emit('batch-analysis')">
                  <el-icon><DataBoard /></el-icon>
                  批量分析
                </el-dropdown-item>
                <el-dropdown-item @click="$emit('import-config')">
                  <el-icon><Upload /></el-icon>
                  导入配置
                </el-dropdown-item>
                <el-dropdown-item @click="$emit('save-template')">
                  <el-icon><CollectionTag /></el-icon>
                  保存模板
                </el-dropdown-item>
                <el-dropdown-item divided @click="$emit('reset-analysis')">
                  <el-icon><RefreshLeft /></el-icon>
                  重置分析
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </div>
      
      <!-- 进度指示器 -->
      <div v-if="showProgress" class="progress-section">
        <div class="progress-info">
          <span class="progress-label">分析进度</span>
          <span class="progress-stats">
            {{ progress.completed }}/{{ progress.total }} 
            ({{ Math.round(progress.percentage) }}%)
          </span>
        </div>
        <el-progress
          :percentage="progress.percentage"
          :status="progress.status"
          :stroke-width="8"
          :show-text="false"
          class="progress-bar"
        />
        <div class="progress-details">
          <span class="current-task">{{ progress.currentTask }}</span>
          <span class="estimated-time">预计剩余: {{ progress.estimatedTime }}</span>
        </div>
      </div>
      
      <!-- 快捷统计 -->
      <div v-if="showStats" class="stats-section">
        <div class="stat-item">
          <div class="stat-value">{{ stats.selectedRequirements }}</div>
          <div class="stat-label">已选需求</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.analyzedRequirements }}</div>
          <div class="stat-label">已分析</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.optimizedRequirements }}</div>
          <div class="stat-label">已优化</div>
        </div>
        <div class="stat-item">
          <div class="stat-value">{{ stats.totalFunctionPoints }}</div>
          <div class="stat-label">功能点总计</div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { 
  MagicStick, Lightning, Setting, Download, More, DataBoard, 
  Upload, CollectionTag, RefreshLeft 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  title: {
    type: String,
    default: '智能需求分析增强'
  },
  subtitle: {
    type: String,
    default: '基于NESMA知识图谱的深度需求分析与优化'
  },
  isAnalyzing: {
    type: Boolean,
    default: false
  },
  canStartAnalysis: {
    type: Boolean,
    default: false
  },
  hasResults: {
    type: Boolean,
    default: false
  },
  showAdvancedActions: {
    type: Boolean,
    default: true
  },
  showProgress: {
    type: Boolean,
    default: false
  },
  showStats: {
    type: Boolean,
    default: true
  },
  progress: {
    type: Object,
    default: () => ({
      completed: 0,
      total: 0,
      percentage: 0,
      status: 'active', // active, success, exception
      currentTask: '',
      estimatedTime: ''
    })
  },
  stats: {
    type: Object,
    default: () => ({
      selectedRequirements: 0,
      analyzedRequirements: 0,
      optimizedRequirements: 0,
      totalFunctionPoints: 0
    })
  }
})

// Emits
const emit = defineEmits([
  'start-analysis',
  'show-config',
  'export-report',
  'batch-analysis',
  'import-config',
  'save-template',
  'reset-analysis'
])

// 计算属性
const progressStatus = computed(() => {
  if (props.progress.percentage === 100) return 'success'
  if (props.progress.status === 'exception') return 'exception'
  return 'active'
})
</script>

<style lang="scss" scoped>
.analysis-header {
  margin-bottom: 20px;

  .header-card {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border: none;
    color: white;

    :deep(.el-card__body) {
      padding: 24px 32px;
    }

    .header-content {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 20px;

      .title-section {
        flex: 1;

        .analysis-title {
          display: flex;
          align-items: center;
          margin: 0 0 8px 0;
          font-size: 28px;
          font-weight: 700;
          color: white;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);

          .title-icon {
            margin-right: 12px;
            font-size: 32px;
            color: #fbbf24;
            filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
          }
        }

        .analysis-subtitle {
          margin: 0;
          font-size: 16px;
          color: rgba(255, 255, 255, 0.9);
          font-weight: 400;
          line-height: 1.5;
        }
      }

      .action-section {
        display: flex;
        align-items: center;
        gap: 16px;

        .main-actions {
          .el-button {
            height: 48px;
            padding: 0 24px;
            font-size: 15px;
            font-weight: 600;
            border-radius: 12px;
            transition: all 0.3s ease;

            &.primary-action {
              background: linear-gradient(135deg, #10b981 0%, #059669 100%);
              border: none;
              box-shadow: 0 4px 12px rgba(16, 185, 129, 0.4);

              &:hover {
                transform: translateY(-2px);
                box-shadow: 0 6px 20px rgba(16, 185, 129, 0.5);
              }

              &:disabled {
                background: #6b7280;
                transform: none;
                box-shadow: none;
              }
            }

            &:not(.primary-action) {
              background: rgba(255, 255, 255, 0.1);
              border: 1px solid rgba(255, 255, 255, 0.2);
              backdrop-filter: blur(10px);

              &:hover {
                background: rgba(255, 255, 255, 0.2);
                border-color: rgba(255, 255, 255, 0.3);
                transform: translateY(-1px);
              }

              &:disabled {
                background: rgba(255, 255, 255, 0.05);
                border-color: rgba(255, 255, 255, 0.1);
                color: rgba(255, 255, 255, 0.5);
              }
            }
          }
        }

        .advanced-dropdown {
          .el-button {
            height: 48px;
            padding: 0 16px;
            background: rgba(255, 255, 255, 0.1);
            border: 1px solid rgba(255, 255, 255, 0.2);
            color: white;
            backdrop-filter: blur(10px);
            border-radius: 12px;

            &:hover {
              background: rgba(255, 255, 255, 0.2);
              transform: translateY(-1px);
            }
          }
        }
      }
    }

    .progress-section {
      margin-bottom: 20px;
      padding: 16px 20px;
      background: rgba(255, 255, 255, 0.1);
      border-radius: 12px;
      backdrop-filter: blur(10px);

      .progress-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 8px;

        .progress-label {
          font-size: 14px;
          font-weight: 600;
          color: white;
        }

        .progress-stats {
          font-size: 13px;
          color: rgba(255, 255, 255, 0.9);
          font-weight: 500;
        }
      }

      .progress-bar {
        margin-bottom: 8px;

        :deep(.el-progress-bar__outer) {
          background: rgba(255, 255, 255, 0.2);
          border-radius: 6px;
        }

        :deep(.el-progress-bar__inner) {
          background: linear-gradient(90deg, #10b981 0%, #34d399 100%);
          border-radius: 6px;
        }
      }

      .progress-details {
        display: flex;
        justify-content: space-between;
        align-items: center;
        font-size: 12px;

        .current-task {
          color: rgba(255, 255, 255, 0.9);
          font-weight: 500;
        }

        .estimated-time {
          color: rgba(255, 255, 255, 0.7);
        }
      }
    }

    .stats-section {
      display: flex;
      justify-content: space-around;
      align-items: center;
      padding: 16px 0;
      border-top: 1px solid rgba(255, 255, 255, 0.2);

      .stat-item {
        text-align: center;
        padding: 0 16px;

        .stat-value {
          font-size: 24px;
          font-weight: 700;
          color: #fbbf24;
          text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
          margin-bottom: 4px;
        }

        .stat-label {
          font-size: 12px;
          color: rgba(255, 255, 255, 0.8);
          font-weight: 500;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }
      }
    }
  }
}

// 深色主题适配
:deep(.el-dropdown-menu) {
  background: #1f2937;
  border: 1px solid #374151;

  .el-dropdown-menu__item {
    color: #f9fafb;

    &:hover {
      background: #374151;
    }

    .el-icon {
      color: #9ca3af;
      margin-right: 8px;
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .analysis-header {
    .header-card {
      :deep(.el-card__body) {
        padding: 16px 20px;
      }

      .header-content {
        flex-direction: column;
        gap: 20px;

        .title-section {
          .analysis-title {
            font-size: 24px;

            .title-icon {
              font-size: 28px;
            }
          }

          .analysis-subtitle {
            font-size: 14px;
          }
        }

        .action-section {
          width: 100%;
          justify-content: center;
          flex-wrap: wrap;

          .main-actions {
            .el-button {
              height: 40px;
              padding: 0 16px;
              font-size: 14px;
            }
          }
        }
      }

      .stats-section {
        flex-wrap: wrap;
        gap: 16px;

        .stat-item {
          min-width: 120px;

          .stat-value {
            font-size: 20px;
          }
        }
      }
    }
  }
}

@media (max-width: 480px) {
  .analysis-header {
    .header-card {
      .header-content {
        .action-section {
          .main-actions {
            flex-direction: column;
            width: 100%;

            .el-button {
              width: 100%;
              margin: 0 0 8px 0;
            }
          }
        }
      }

      .stats-section {
        .stat-item {
          min-width: 80px;

          .stat-value {
            font-size: 18px;
          }

          .stat-label {
            font-size: 11px;
          }
        }
      }
    }
  }
}
</style>