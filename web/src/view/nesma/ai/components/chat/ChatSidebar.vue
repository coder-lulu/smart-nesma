<template>
  <div class="chat-sidebar">
    <div class="sidebar-header">
      <h3>AI对话</h3>
      <el-button 
        type="primary" 
        size="small" 
        @click="$emit('create-session')"
        :icon="Plus"
      >
        新建对话
      </el-button>
    </div>

    <div class="sidebar-content">
      <!-- 搜索框 -->
      <div class="search-section">
        <el-input
          v-model="searchQuery"
          placeholder="搜索对话..."
          :prefix-icon="Search"
          clearable
          class="search-input"
          @input="handleSearch"
        />
      </div>

      <!-- 会话分类筛选 -->
      <div class="filter-section">
        <el-select 
          v-model="selectedCategory" 
          placeholder="选择分类"
          size="small"
          clearable
          @change="handleCategoryChange"
          style="width: 100%"
        >
          <el-option label="全部对话" value="" />
          <el-option label="需求分析" value="requirement_analysis" />
          <el-option label="技术咨询" value="technical_consultation" />
          <el-option label="项目管理" value="project_management" />
          <el-option label="知识问答" value="knowledge_qa" />
          <el-option label="其他" value="other" />
        </el-select>
      </div>

      <!-- 会话列表 -->
      <div class="sessions-list">
        <div v-if="loading" class="loading-container">
          <el-skeleton :rows="3" animated />
        </div>
        
        <div v-else-if="filteredSessions.length === 0" class="empty-sessions">
          <el-empty description="暂无对话" />
          <el-button 
            type="primary" 
            size="small"
            @click="$emit('create-session')"
          >
            创建第一个对话
          </el-button>
        </div>
        
        <div 
          v-else
          v-for="session in filteredSessions"
          :key="session.id"
          class="session-item"
          :class="{ 
            active: selectedSessionId === session.id,
            unread: session.unreadCount > 0
          }"
          @click="$emit('select-session', session)"
        >
          <div class="session-main">
            <div class="session-header">
              <div class="session-title">{{ session.title }}</div>
              <div class="session-badges">
                <el-badge 
                  v-if="session.unreadCount > 0" 
                  :value="session.unreadCount" 
                  :max="99"
                  class="unread-badge"
                />
                <el-tag 
                  v-if="session.category" 
                  size="small" 
                  :type="getCategoryType(session.category)"
                  class="category-tag"
                >
                  {{ getCategoryLabel(session.category) }}
                </el-tag>
              </div>
            </div>
            
            <div class="session-preview">
              {{ session.lastMessage || '暂无消息' }}
            </div>
            
            <div class="session-meta">
              <div class="session-time">
                <el-icon><Clock /></el-icon>
                {{ formatTime(session.updatedAt) }}
              </div>
              <div class="session-stats">
                <span class="message-count">
                  <el-icon><ChatDotRound /></el-icon>
                  {{ session.messageCount || 0 }}
                </span>
              </div>
            </div>
          </div>
          
          <div class="session-actions">
            <el-dropdown 
              trigger="click" 
              @command="(action) => handleSessionAction(action, session)"
              @click.stop
            >
              <el-button 
                text 
                size="small" 
                class="action-btn"
              >
                <el-icon><More /></el-icon>
              </el-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item command="edit">
                    <el-icon><Edit /></el-icon>
                    重命名
                  </el-dropdown-item>
                  <el-dropdown-item command="copy">
                    <el-icon><DocumentCopy /></el-icon>
                    复制对话
                  </el-dropdown-item>
                  <el-dropdown-item command="export">
                    <el-icon><Download /></el-icon>
                    导出对话
                  </el-dropdown-item>
                  <el-dropdown-item command="pin" :disabled="session.isPinned">
                    <el-icon><Star /></el-icon>
                    置顶
                  </el-dropdown-item>
                  <el-dropdown-item command="unpin" :disabled="!session.isPinned">
                    <el-icon><StarFilled /></el-icon>
                    取消置顶
                  </el-dropdown-item>
                  <el-dropdown-item command="archive">
                    <el-icon><Box /></el-icon>
                    归档
                  </el-dropdown-item>
                  <el-dropdown-item command="delete" divided>
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
          </div>
        </div>
      </div>
    </div>

    <!-- 底部统计信息 -->
    <div class="sidebar-footer">
      <div class="stats-info">
        <div class="stat-item">
          <span class="stat-label">总对话数</span>
          <span class="stat-value">{{ totalSessions }}</span>
        </div>
        <div class="stat-item">
          <span class="stat-label">今日消息</span>
          <span class="stat-value">{{ todayMessages }}</span>
        </div>
      </div>
      
      <div class="footer-actions">
        <el-button 
          size="small" 
          text
          @click="$emit('clear-all')"
          :disabled="totalSessions === 0"
        >
          <el-icon><Delete /></el-icon>
          清空所有
        </el-button>
        <el-button 
          size="small" 
          text
          @click="$emit('refresh-sessions')"
        >
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { 
  Plus, Search, Clock, ChatDotRound, More, Edit, DocumentCopy, 
  Download, Star, StarFilled, Box, Delete, Refresh 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  sessions: {
    type: Array,
    default: () => []
  },
  selectedSessionId: {
    type: [String, Number],
    default: null
  },
  loading: {
    type: Boolean,
    default: false
  }
})

// Emits
const emit = defineEmits([
  'create-session',
  'select-session',
  'session-action',
  'clear-all',
  'refresh-sessions'
])

// 响应式数据
const searchQuery = ref('')
const selectedCategory = ref('')

// 计算属性
const filteredSessions = computed(() => {
  let filtered = props.sessions

  // 关键词搜索
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    filtered = filtered.filter(session => 
      session.title.toLowerCase().includes(query) ||
      session.lastMessage?.toLowerCase().includes(query)
    )
  }

  // 分类筛选
  if (selectedCategory.value) {
    filtered = filtered.filter(session => 
      session.category === selectedCategory.value
    )
  }

  // 置顶排序
  return filtered.sort((a, b) => {
    if (a.isPinned && !b.isPinned) return -1
    if (!a.isPinned && b.isPinned) return 1
    return new Date(b.updatedAt) - new Date(a.updatedAt)
  })
})

const totalSessions = computed(() => props.sessions.length)

const todayMessages = computed(() => {
  const today = new Date().toDateString()
  return props.sessions.reduce((count, session) => {
    const sessionDate = new Date(session.updatedAt).toDateString()
    return sessionDate === today ? count + (session.messageCount || 0) : count
  }, 0)
})

// 方法
const formatTime = (time) => {
  if (!time) return ''
  
  const now = new Date()
  const date = new Date(time)
  const diffInSeconds = Math.floor((now - date) / 1000)
  
  if (diffInSeconds < 60) return '刚刚'
  if (diffInSeconds < 3600) return `${Math.floor(diffInSeconds / 60)}分钟前`
  if (diffInSeconds < 86400) return `${Math.floor(diffInSeconds / 3600)}小时前`
  
  const diffInDays = Math.floor(diffInSeconds / 86400)
  if (diffInDays < 7) return `${diffInDays}天前`
  
  return date.toLocaleDateString('zh-CN')
}

const getCategoryType = (category) => {
  const typeMap = {
    'requirement_analysis': 'primary',
    'technical_consultation': 'success',
    'project_management': 'warning',
    'knowledge_qa': 'info',
    'other': 'default'
  }
  return typeMap[category] || 'default'
}

const getCategoryLabel = (category) => {
  const labelMap = {
    'requirement_analysis': '需求分析',
    'technical_consultation': '技术咨询',
    'project_management': '项目管理',
    'knowledge_qa': '知识问答',
    'other': '其他'
  }
  return labelMap[category] || category
}

const handleSearch = () => {
  // 搜索逻辑由computed filteredSessions处理
}

const handleCategoryChange = () => {
  // 分类筛选逻辑由computed filteredSessions处理
}

const handleSessionAction = (action, session) => {
  emit('session-action', { action, session })
}

// 监听器
watch(searchQuery, () => {
  // 防抖处理可以在这里添加
})
</script>

<style lang="scss" scoped>
.chat-sidebar {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #f8fafc;
  border-right: 1px solid #e5e7eb;

  .sidebar-header {
    padding: 16px;
    border-bottom: 1px solid #e5e7eb;
    background: white;

    h3 {
      margin: 0 0 12px 0;
      font-size: 16px;
      font-weight: 600;
      color: #1f2937;
    }
  }

  .sidebar-content {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;

    .search-section {
      padding: 12px 16px;
      background: white;
      border-bottom: 1px solid #e5e7eb;
    }

    .filter-section {
      padding: 8px 16px 12px;
      background: white;
      border-bottom: 1px solid #e5e7eb;
    }

    .sessions-list {
      flex: 1;
      overflow-y: auto;
      padding: 8px 0;

      .loading-container {
        padding: 16px;
      }

      .empty-sessions {
        padding: 40px 16px;
        text-align: center;
      }

      .session-item {
        display: flex;
        align-items: center;
        padding: 12px 16px;
        margin: 4px 8px;
        border-radius: 8px;
        cursor: pointer;
        transition: all 0.2s ease;
        background: white;
        border: 1px solid transparent;

        &:hover {
          background: #f3f4f6;
          transform: translateY(-1px);
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
        }

        &.active {
          background: #eff6ff;
          border-color: #3b82f6;
          box-shadow: 0 2px 8px rgba(59, 130, 246, 0.15);
        }

        &.unread {
          border-left: 3px solid #10b981;
        }

        .session-main {
          flex: 1;
          min-width: 0;

          .session-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 6px;

            .session-title {
              font-weight: 500;
              color: #1f2937;
              font-size: 14px;
              line-height: 1.4;
              overflow: hidden;
              text-overflow: ellipsis;
              white-space: nowrap;
              flex: 1;
              margin-right: 8px;
            }

            .session-badges {
              display: flex;
              align-items: center;
              gap: 4px;
              flex-shrink: 0;

              .category-tag {
                font-size: 10px;
                height: 18px;
                line-height: 16px;
                padding: 0 4px;
              }
            }
          }

          .session-preview {
            font-size: 12px;
            color: #6b7280;
            line-height: 1.4;
            margin-bottom: 6px;
            overflow: hidden;
            text-overflow: ellipsis;
            white-space: nowrap;
          }

          .session-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            font-size: 11px;
            color: #9ca3af;

            .session-time {
              display: flex;
              align-items: center;
              gap: 2px;
            }

            .session-stats {
              display: flex;
              align-items: center;
              gap: 4px;

              .message-count {
                display: flex;
                align-items: center;
                gap: 2px;
              }
            }
          }
        }

        .session-actions {
          opacity: 0;
          transition: opacity 0.2s ease;
          margin-left: 8px;

          .action-btn {
            width: 24px;
            height: 24px;
            border-radius: 4px;
            color: #6b7280;

            &:hover {
              background: #e5e7eb;
              color: #374151;
            }
          }
        }

        &:hover .session-actions {
          opacity: 1;
        }
      }
    }
  }

  .sidebar-footer {
    padding: 12px 16px;
    border-top: 1px solid #e5e7eb;
    background: white;

    .stats-info {
      display: flex;
      justify-content: space-between;
      margin-bottom: 8px;

      .stat-item {
        text-align: center;

        .stat-label {
          display: block;
          font-size: 10px;
          color: #9ca3af;
          margin-bottom: 2px;
        }

        .stat-value {
          font-size: 14px;
          font-weight: 600;
          color: #1f2937;
        }
      }
    }

    .footer-actions {
      display: flex;
      justify-content: space-between;
      gap: 8px;

      .el-button {
        font-size: 11px;
        padding: 4px 8px;
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .chat-sidebar {
    .session-item {
      padding: 8px 12px;
      margin: 2px 4px;

      .session-main {
        .session-title {
          font-size: 13px;
        }

        .session-preview {
          font-size: 11px;
        }
      }
    }
  }
}
</style>