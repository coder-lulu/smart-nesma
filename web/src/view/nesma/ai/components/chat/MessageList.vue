<template>
  <div class="message-list" ref="messageContainer">
    <div class="messages-wrapper">
      <!-- 消息列表 -->
      <div
        v-for="message in messages"
        :key="message.id"
        class="message-item"
        :class="[message.role, { 'has-error': message.error }]"
      >
        <div class="message-avatar">
          <el-avatar
            :size="32"
            :src="message.role === 'user' ? userAvatar : aiAvatar"
          >
            {{ message.role === 'user' ? 'U' : 'AI' }}
          </el-avatar>
        </div>
        
        <div class="message-content">
          <div class="message-header">
            <span class="message-sender">
              {{ message.role === 'user' ? (userName || '您') : 'AI助手' }}
            </span>
            <span class="message-time">{{ formatTime(message.createdAt) }}</span>
            <div class="message-status">
              <el-icon v-if="message.status === 'sending'" class="loading">
                <Loading />
              </el-icon>
              <el-icon v-else-if="message.error" class="error">
                <Warning />
              </el-icon>
              <el-icon v-else-if="message.status === 'sent'" class="success">
                <Check />
              </el-icon>
            </div>
          </div>
          
          <div class="message-body">
            <div 
              class="message-text" 
              v-html="formatMessage(message.content)"
              @click="handleMessageClick"
            />
            
            <!-- 错误信息显示 -->
            <div v-if="message.error" class="message-error">
              <el-alert
                :title="message.error"
                type="error"
                size="small"
                :closable="false"
              />
              <el-button 
                size="small" 
                type="primary" 
                plain
                @click="$emit('retry-message', message)"
                style="margin-top: 8px;"
              >
                重试发送
              </el-button>
            </div>
            
            <!-- AI消息的额外信息 -->
            <div v-if="message.role === 'assistant'" class="message-meta">
              <div class="meta-info">
                <span v-if="message.tokenCount" class="token-count">
                  <el-icon><Coin /></el-icon>
                  {{ message.tokenCount }} tokens
                </span>
                <span v-if="message.model" class="model-info">
                  <el-icon><Cpu /></el-icon>
                  {{ message.model }}
                </span>
                <span v-if="message.responseTime" class="response-time">
                  <el-icon><Timer /></el-icon>
                  {{ message.responseTime }}ms
                </span>
              </div>
              
              <div class="message-actions">
                <el-button-group size="small">
                  <el-button 
                    text 
                    @click="$emit('copy-message', message)"
                    :icon="DocumentCopy"
                  >
                    复制
                  </el-button>
                  <el-button 
                    text 
                    @click="$emit('regenerate-message', message)"
                    :icon="Refresh"
                    :loading="message.regenerating"
                  >
                    重新生成
                  </el-button>
                  <el-button 
                    text 
                    @click="$emit('like-message', message)"
                    :icon="message.liked ? StarFilled : Star"
                    :type="message.liked ? 'primary' : 'default'"
                  >
                    {{ message.liked ? '已点赞' : '点赞' }}
                  </el-button>
                  <el-button 
                    text 
                    @click="$emit('export-message', message)"
                    :icon="Download"
                  >
                    导出
                  </el-button>
                </el-button-group>
              </div>
            </div>
            
            <!-- 用户消息的操作按钮 -->
            <div v-else class="user-message-actions">
              <el-button 
                text 
                size="small"
                @click="$emit('edit-message', message)"
                :icon="Edit"
              >
                编辑
              </el-button>
              <el-button 
                text 
                size="small"
                @click="$emit('delete-message', message)"
                :icon="Delete"
              >
                删除
              </el-button>
            </div>
          </div>
        </div>
      </div>
      
      <!-- 正在输入指示器 -->
      <div v-if="isTyping" class="typing-indicator">
        <div class="typing-avatar">
          <el-avatar :size="32" :src="aiAvatar">AI</el-avatar>
        </div>
        <div class="typing-content">
          <div class="typing-dots">
            <span></span>
            <span></span>
            <span></span>
          </div>
          <span class="typing-text">{{ typingText }}</span>
        </div>
      </div>
      
      <!-- 加载更多历史消息 -->
      <div v-if="hasMoreMessages" class="load-more-container">
        <el-button 
          text 
          @click="$emit('load-more-messages')"
          :loading="loadingMore"
        >
          加载更多历史消息
        </el-button>
      </div>
    </div>
    
    <!-- 滚动到底部按钮 -->
    <transition name="fade">
      <el-button
        v-show="showScrollButton"
        class="scroll-to-bottom"
        type="primary"
        :icon="ArrowDown"
        circle
        @click="scrollToBottom"
      />
    </transition>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch, onMounted, onUnmounted } from 'vue'
import { 
  Loading, Warning, Check, Coin, Cpu, Timer, DocumentCopy, 
  Refresh, Star, StarFilled, Download, Edit, Delete, ArrowDown 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  messages: {
    type: Array,
    default: () => []
  },
  isTyping: {
    type: Boolean,
    default: false
  },
  typingText: {
    type: String,
    default: 'AI正在思考...'
  },
  userAvatar: {
    type: String,
    default: ''
  },
  aiAvatar: {
    type: String,
    default: ''
  },
  userName: {
    type: String,
    default: ''
  },
  hasMoreMessages: {
    type: Boolean,
    default: false
  },
  loadingMore: {
    type: Boolean,
    default: false
  },
  autoScroll: {
    type: Boolean,
    default: true
  }
})

// Emits
const emit = defineEmits([
  'copy-message',
  'regenerate-message',
  'like-message',
  'export-message',
  'edit-message',
  'delete-message',
  'retry-message',
  'load-more-messages',
  'scroll-to-bottom'
])

// 响应式数据
const messageContainer = ref()
const showScrollButton = ref(false)

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
  
  return date.toLocaleDateString('zh-CN', {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  })
}

const formatMessage = (content) => {
  if (!content) return ''
  
  // 增强的Markdown处理
  let formatted = content
    // 代码块处理
    .replace(/```(\w+)?\n([\s\S]*?)```/g, (match, lang, code) => {
      const language = lang || 'text'
      return `<div class="code-block">
        <div class="code-header">
          <span class="language">${language}</span>
          <button class="copy-code" onclick="copyCode(this)">复制</button>
        </div>
        <pre><code class="language-${language}">${code.trim()}</code></pre>
      </div>`
    })
    // 行内代码
    .replace(/`([^`]+)`/g, '<code class="inline-code">$1</code>')
    // 粗体
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    // 斜体
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    // 链接
    .replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener">$1</a>')
    // 换行
    .replace(/\n/g, '<br>')
    // 表格处理（简化版）
    .replace(/\|(.+)\|/g, (match) => {
      const cells = match.split('|').filter(cell => cell.trim())
      return `<div class="table-row">${cells.map(cell => `<span class="table-cell">${cell.trim()}</span>`).join('')}</div>`
    })
  
  return formatted
}

const handleMessageClick = (event) => {
  const target = event.target
  
  if (target.tagName === 'CODE' && target.classList.contains('inline-code')) {
    // 复制行内代码
    navigator.clipboard.writeText(target.textContent).then(() => {
      ElMessage.success('代码已复制到剪贴板')
    })
  } else if (target.tagName === 'A') {
    // 确认外部链接点击
    const href = target.getAttribute('href')
    if (href && !href.startsWith('#')) {
      event.preventDefault()
      ElMessageBox.confirm(
        `确定要打开外部链接吗？\n${href}`,
        '外部链接确认',
        {
          confirmButtonText: '打开',
          cancelButtonText: '取消',
          type: 'info'
        }
      ).then(() => {
        window.open(href, '_blank', 'noopener,noreferrer')
      }).catch(() => {})
    }
  }
}

const scrollToBottom = (smooth = true) => {
  nextTick(() => {
    if (messageContainer.value) {
      messageContainer.value.scrollTo({
        top: messageContainer.value.scrollHeight,
        behavior: smooth ? 'smooth' : 'auto'
      })
      emit('scroll-to-bottom')
    }
  })
}

const checkScrollPosition = () => {
  if (!messageContainer.value) return
  
  const { scrollTop, scrollHeight, clientHeight } = messageContainer.value
  const isNearBottom = scrollHeight - scrollTop - clientHeight < 100
  showScrollButton.value = !isNearBottom && props.messages.length > 5
}

// 全局复制代码函数
window.copyCode = (button) => {
  const codeBlock = button.closest('.code-block')
  const code = codeBlock.querySelector('code').textContent
  navigator.clipboard.writeText(code).then(() => {
    button.textContent = '已复制'
    setTimeout(() => {
      button.textContent = '复制'
    }, 2000)
  })
}

// 监听器
watch(() => props.messages.length, () => {
  if (props.autoScroll) {
    scrollToBottom()
  }
})

watch(() => props.isTyping, (isTyping) => {
  if (isTyping && props.autoScroll) {
    scrollToBottom()
  }
})

// 生命周期
onMounted(() => {
  if (messageContainer.value) {
    messageContainer.value.addEventListener('scroll', checkScrollPosition)
    scrollToBottom(false) // 初始化时不使用动画
  }
})

onUnmounted(() => {
  if (messageContainer.value) {
    messageContainer.value.removeEventListener('scroll', checkScrollPosition)
  }
})
</script>

<style lang="scss" scoped>
.message-list {
  position: relative;
  height: 100%;
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .messages-wrapper {
    flex: 1;
    overflow-y: auto;
    padding: 16px 20px;
    scroll-behavior: smooth;

    .message-item {
      display: flex;
      align-items: flex-start;
      margin-bottom: 20px;
      opacity: 0;
      animation: fadeInUp 0.4s ease-out forwards;

      &.user {
        flex-direction: row-reverse;

        .message-content {
          background: linear-gradient(135deg, #3b82f6 0%, #1d4ed8 100%);
          color: white;
          margin-right: 12px;
          margin-left: 60px;
          border: none;

          .message-header {
            .message-sender,
            .message-time {
              color: rgba(255, 255, 255, 0.9);
            }
          }

          .user-message-actions {
            border-top: 1px solid rgba(255, 255, 255, 0.2);
            margin-top: 8px;
            padding-top: 8px;

            .el-button {
              color: rgba(255, 255, 255, 0.8);

              &:hover {
                color: white;
                background: rgba(255, 255, 255, 0.1);
              }
            }
          }
        }
      }

      &.assistant {
        .message-content {
          background: white;
          border: 1px solid #e5e7eb;
          margin-left: 12px;
          margin-right: 60px;
          box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
        }
      }

      &.has-error {
        .message-content {
          border-color: #f87171;
          background: #fef2f2;
        }
      }

      .message-avatar {
        flex-shrink: 0;
        position: sticky;
        top: 0;
      }

      .message-content {
        max-width: calc(100% - 80px);
        border-radius: 16px;
        padding: 16px 20px;
        position: relative;
        word-wrap: break-word;

        .message-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;

          .message-sender {
            font-size: 13px;
            font-weight: 600;
          }

          .message-time {
            font-size: 11px;
            opacity: 0.7;
          }

          .message-status {
            .loading {
              color: #3b82f6;
              animation: spin 1s linear infinite;
            }

            .error {
              color: #ef4444;
            }

            .success {
              color: #10b981;
            }
          }
        }

        .message-body {
          .message-text {
            line-height: 1.7;
            font-size: 14px;

            :deep(.inline-code) {
              background: rgba(99, 102, 241, 0.1);
              color: #4f46e5;
              padding: 2px 6px;
              border-radius: 4px;
              font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
              font-size: 0.9em;
              cursor: pointer;
              transition: all 0.2s ease;

              &:hover {
                background: rgba(99, 102, 241, 0.2);
              }
            }

            :deep(.code-block) {
              margin: 12px 0;
              border-radius: 8px;
              overflow: hidden;
              border: 1px solid #e5e7eb;

              .code-header {
                background: #f8fafc;
                padding: 8px 12px;
                display: flex;
                justify-content: space-between;
                align-items: center;
                border-bottom: 1px solid #e5e7eb;

                .language {
                  font-size: 12px;
                  color: #6b7280;
                  font-weight: 500;
                }

                .copy-code {
                  background: #3b82f6;
                  color: white;
                  border: none;
                  padding: 4px 8px;
                  border-radius: 4px;
                  font-size: 11px;
                  cursor: pointer;
                  transition: background 0.2s ease;

                  &:hover {
                    background: #2563eb;
                  }
                }
              }

              pre {
                background: #1f2937;
                color: #f9fafb;
                padding: 16px;
                margin: 0;
                overflow-x: auto;
                font-family: 'SF Mono', 'Monaco', 'Inconsolata', 'Roboto Mono', monospace;
                font-size: 13px;
                line-height: 1.5;

                code {
                  background: none;
                  color: inherit;
                  padding: 0;
                }
              }
            }

            :deep(.table-row) {
              display: flex;
              border: 1px solid #e5e7eb;
              margin: 4px 0;

              &:first-child {
                background: #f8fafc;
                font-weight: 600;
              }

              .table-cell {
                flex: 1;
                padding: 8px 12px;
                border-right: 1px solid #e5e7eb;

                &:last-child {
                  border-right: none;
                }
              }
            }

            :deep(a) {
              color: #3b82f6;
              text-decoration: none;

              &:hover {
                text-decoration: underline;
              }
            }

            :deep(strong) {
              font-weight: 600;
              color: #1f2937;
            }

            :deep(em) {
              font-style: italic;
              color: #6b7280;
            }
          }

          .message-error {
            margin-top: 12px;
          }

          .message-meta {
            margin-top: 12px;
            padding-top: 12px;
            border-top: 1px solid #f3f4f6;
            display: flex;
            justify-content: space-between;
            align-items: center;

            .meta-info {
              display: flex;
              gap: 16px;
              font-size: 11px;
              color: #6b7280;

              span {
                display: flex;
                align-items: center;
                gap: 4px;
              }
            }

            .message-actions {
              opacity: 0;
              transition: opacity 0.3s ease;

              .el-button {
                font-size: 11px;
                padding: 4px 8px;
              }
            }
          }

          .user-message-actions {
            margin-top: 8px;
            padding-top: 8px;
            display: flex;
            gap: 8px;
            opacity: 0;
            transition: opacity 0.3s ease;
          }
        }

        &:hover {
          .message-actions,
          .user-message-actions {
            opacity: 1;
          }
        }
      }
    }

    .typing-indicator {
      display: flex;
      align-items: flex-start;
      margin-bottom: 20px;
      animation: fadeInUp 0.4s ease-out;

      .typing-avatar {
        flex-shrink: 0;
      }

      .typing-content {
        margin-left: 12px;
        background: white;
        border: 1px solid #e5e7eb;
        border-radius: 16px;
        padding: 16px 20px;
        display: flex;
        align-items: center;
        gap: 12px;
        box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);

        .typing-dots {
          display: flex;
          gap: 4px;

          span {
            width: 8px;
            height: 8px;
            background: #6b7280;
            border-radius: 50%;
            animation: typing 1.4s infinite ease-in-out;

            &:nth-child(1) { animation-delay: -0.32s; }
            &:nth-child(2) { animation-delay: -0.16s; }
          }
        }

        .typing-text {
          font-size: 13px;
          color: #6b7280;
          font-style: italic;
        }
      }
    }

    .load-more-container {
      text-align: center;
      padding: 16px 0;
      border-bottom: 1px solid #e5e7eb;
      margin-bottom: 16px;
    }
  }

  .scroll-to-bottom {
    position: absolute;
    right: 20px;
    bottom: 20px;
    z-index: 10;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes typing {
  0%, 80%, 100% {
    transform: scale(0);
  }
  40% {
    transform: scale(1);
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

// 响应式设计
@media (max-width: 768px) {
  .message-list {
    .messages-wrapper {
      padding: 12px 16px;

      .message-item {
        margin-bottom: 16px;

        &.user .message-content {
          margin-left: 40px;
        }

        &.assistant .message-content {
          margin-right: 40px;
        }

        .message-content {
          padding: 12px 16px;
          border-radius: 12px;

          .message-body {
            .message-text {
              font-size: 13px;

              :deep(.code-block) {
                .code-header {
                  padding: 6px 8px;
                  
                  .language {
                    font-size: 11px;
                  }

                  .copy-code {
                    font-size: 10px;
                    padding: 3px 6px;
                  }
                }

                pre {
                  padding: 12px;
                  font-size: 12px;
                }
              }
            }

            .message-meta {
              flex-direction: column;
              gap: 8px;

              .meta-info {
                justify-content: center;
              }

              .message-actions {
                opacity: 1;
              }
            }

            .user-message-actions {
              opacity: 1;
              justify-content: center;
            }
          }
        }
      }
    }

    .scroll-to-bottom {
      right: 16px;
      bottom: 16px;
    }
  }
}
</style>