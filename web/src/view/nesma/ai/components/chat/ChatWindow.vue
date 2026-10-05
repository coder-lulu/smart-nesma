<template>
  <div class="chat-window">
    <!-- 空状态 -->
    <div v-if="!currentSession" class="empty-state">
      <div class="empty-icon">
        <el-icon size="64"><ChatDotRound /></el-icon>
      </div>
      <h3>开始新的对话</h3>
      <p>选择一个对话或创建新对话来开始</p>
      <el-button 
        type="primary" 
        @click="$emit('create-session')"
        :icon="Plus"
      >
        创建新对话
      </el-button>
    </div>

    <!-- 聊天界面 -->
    <div v-else class="chat-content">
      <!-- 聊天头部 -->
      <div class="chat-header">
        <div class="header-left">
          <h3>{{ currentSession.title }}</h3>
          <div class="header-meta">
            <el-tag size="small" type="info">{{ selectedModel }}</el-tag>
            <el-tag 
              size="small" 
              :type="wsConnected ? 'success' : 'warning'"
              style="margin-left: 8px;"
            >
              {{ wsConnected ? 'WebSocket已连接' : 'WebSocket未连接' }}
            </el-tag>
            <span class="message-count">{{ messages.length }} 条消息</span>
          </div>
        </div>
        <div class="header-right">
          <!-- 项目选择 -->
          <div class="project-selector">
            <el-select 
              :model-value="selectedProject" 
              placeholder="选择项目" 
              style="width: 200px;"
              @change="$emit('project-change', $event)"
            >
              <el-option
                v-for="project in projects"
                :key="project.id"
                :label="project.name"
                :value="project.id"
              />
            </el-select>
          </div>
          
          <!-- 模型选择 -->
          <div class="model-selector">
            <el-select 
              :model-value="selectedModel" 
              placeholder="选择模型" 
              style="width: 180px;"
              @change="$emit('model-change', $event)"
            >
              <el-option label="DeepSeek Chat" value="deepseek-chat" />
              <el-option label="GPT-3.5 Turbo" value="gpt-3.5-turbo" />
              <el-option label="GPT-4" value="gpt-4" />
            </el-select>
          </div>
          
          <el-button 
            @click="$emit('export-session')"
            :icon="Download"
            size="small"
          >
            导出
          </el-button>
          <el-button 
            @click="$emit('clear-messages')"
            :icon="Delete"
            size="small"
          >
            清空
          </el-button>
        </div>
      </div>

      <!-- 消息列表 -->
      <div class="messages-container" ref="messagesContainer">
        <div
          v-for="message in messages"
          :key="message.id"
          class="message-item"
          :class="message.role"
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
                {{ message.role === 'user' ? '您' : 'AI助手' }}
              </span>
              <span class="message-time">{{ formatTime(message.createdAt) }}</span>
            </div>
            <div 
              class="message-text" 
              v-html="formatMessage(message.content)" 
              @click="handleMessageClick"
            />
            <div v-if="message.role === 'assistant'" class="message-meta">
              <span class="token-count">Token: {{ message.tokenCount || 0 }}</span>
              <div class="message-actions">
                <el-button 
                  text 
                  size="small" 
                  @click="$emit('copy-message', message)"
                  :icon="DocumentCopy"
                >
                  复制
                </el-button>
                <el-button 
                  text 
                  size="small" 
                  @click="$emit('regenerate-message', message)"
                  :icon="Refresh"
                >
                  重新生成
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
            <span class="typing-text">AI正在思考...</span>
          </div>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-container">
        <div class="input-wrapper">
          <el-input
            :model-value="inputMessage"
            type="textarea"
            :rows="inputRows"
            placeholder="输入您的问题... (Ctrl+Enter发送)"
            :disabled="isTyping"
            @input="$emit('input-change', $event)"
            @keydown.enter.ctrl="$emit('send-message')"
            class="message-input"
          />
          <div class="input-actions">
            <div class="input-tools">
              <el-button 
                text 
                size="small" 
                :icon="Paperclip"
                @click="$emit('attach-file')"
              >
                文件
              </el-button>
              <el-button 
                text 
                size="small" 
                :icon="Picture"
                @click="$emit('attach-image')"
              >
                图片
              </el-button>
              <el-button 
                text 
                size="small" 
                :icon="Microphone"
                @click="$emit('start-voice-input')"
              >
                语音
              </el-button>
            </div>
            <div class="send-controls">
              <el-popover
                placement="top"
                :width="250"
                trigger="hover"
                content="快捷键：Ctrl+Enter 发送消息"
              >
                <template #reference>
                  <el-button
                    type="primary"
                    :loading="isTyping"
                    :disabled="!inputMessage.trim()"
                    @click="$emit('send-message')"
                    :icon="Promotion"
                  >
                    发送
                  </el-button>
                </template>
              </el-popover>
            </div>
          </div>
        </div>
        
        <!-- 快捷回复 -->
        <div v-if="quickReplies.length > 0" class="quick-replies">
          <span class="quick-label">快捷回复：</span>
          <el-tag
            v-for="reply in quickReplies"
            :key="reply"
            size="small"
            clickable
            @click="$emit('select-quick-reply', reply)"
            class="quick-reply-tag"
          >
            {{ reply }}
          </el-tag>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, nextTick, watch } from 'vue'
import { 
  ChatDotRound, Plus, Download, Delete, DocumentCopy, 
  Refresh, Paperclip, Picture, Microphone, Promotion 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  currentSession: {
    type: Object,
    default: null
  },
  messages: {
    type: Array,
    default: () => []
  },
  selectedModel: {
    type: String,
    default: 'deepseek-chat'
  },
  selectedProject: {
    type: [String, Number],
    default: null
  },
  projects: {
    type: Array,
    default: () => []
  },
  wsConnected: {
    type: Boolean,
    default: false
  },
  isTyping: {
    type: Boolean,
    default: false
  },
  inputMessage: {
    type: String,
    default: ''
  },
  userAvatar: {
    type: String,
    default: ''
  },
  aiAvatar: {
    type: String,
    default: ''
  },
  quickReplies: {
    type: Array,
    default: () => [
      '请分析这个需求',
      '生成Mermaid图表',
      '优化需求描述',
      '计算功能点',
      '帮我检查语法'
    ]
  }
})

// Emits
const emit = defineEmits([
  'create-session',
  'project-change',
  'model-change',
  'export-session',
  'clear-messages',
  'copy-message',
  'regenerate-message',
  'input-change',
  'send-message',
  'attach-file',
  'attach-image',
  'start-voice-input',
  'select-quick-reply'
])

// 响应式数据
const messagesContainer = ref()
const inputRows = ref(2)

// 计算属性
const messageCount = computed(() => props.messages.length)

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

const formatMessage = (content) => {
  if (!content) return ''
  
  // 简化的Markdown处理
  let formatted = content
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/\*(.*?)\*/g, '<em>$1</em>')
    .replace(/`([^`]+)`/g, '<code>$1</code>')
    .replace(/```(\w+)?\n([\s\S]*?)```/g, '<pre><code>$2</code></pre>')
    .replace(/\n/g, '<br>')
  
  return formatted
}

const handleMessageClick = (event) => {
  // 处理消息点击事件，如代码复制等
  if (event.target.tagName === 'CODE') {
    navigator.clipboard.writeText(event.target.textContent)
    ElMessage.success('代码已复制到剪贴板')
  }
}

const scrollToBottom = () => {
  nextTick(() => {
    if (messagesContainer.value) {
      messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
    }
  })
}

// 监听器
watch(() => props.messages.length, () => {
  scrollToBottom()
})

watch(() => props.isTyping, (isTyping) => {
  if (isTyping) {
    scrollToBottom()
  }
})

// 监听输入内容变化，动态调整输入框高度
watch(() => props.inputMessage, (newMessage) => {
  const lines = newMessage.split('\n').length
  inputRows.value = Math.min(Math.max(lines, 2), 6)
})
</script>

<style lang="scss" scoped>
.chat-window {
  display: flex;
  flex-direction: column;
  height: 100%;
  background: #ffffff;

  .empty-state {
    flex: 1;
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    text-align: center;
    color: #6b7280;

    .empty-icon {
      margin-bottom: 24px;
      color: #d1d5db;
    }

    h3 {
      margin: 0 0 8px 0;
      font-size: 20px;
      font-weight: 600;
      color: #374151;
    }

    p {
      margin: 0 0 24px 0;
      font-size: 14px;
    }
  }

  .chat-content {
    display: flex;
    flex-direction: column;
    height: 100%;

    .chat-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 16px 20px;
      border-bottom: 1px solid #e5e7eb;
      background: #f9fafb;

      .header-left {
        flex: 1;

        h3 {
          margin: 0 0 4px 0;
          font-size: 16px;
          font-weight: 600;
          color: #1f2937;
        }

        .header-meta {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 12px;
          color: #6b7280;

          .message-count {
            margin-left: 8px;
            padding: 2px 6px;
            background: #e5e7eb;
            border-radius: 4px;
            font-size: 11px;
          }
        }
      }

      .header-right {
        display: flex;
        align-items: center;
        gap: 12px;

        .project-selector,
        .model-selector {
          :deep(.el-select) {
            .el-input__inner {
              font-size: 12px;
            }
          }
        }
      }
    }

    .messages-container {
      flex: 1;
      overflow-y: auto;
      padding: 16px 20px;
      background: #f8fafc;

      .message-item {
        display: flex;
        align-items: flex-start;
        margin-bottom: 16px;
        animation: fadeInUp 0.3s ease-out;

        &.user {
          flex-direction: row-reverse;

          .message-content {
            background: #3b82f6;
            color: white;
            margin-right: 12px;
            margin-left: 60px;

            .message-header {
              .message-sender {
                color: #dbeafe;
              }

              .message-time {
                color: #93c5fd;
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
          }
        }

        .message-avatar {
          flex-shrink: 0;
        }

        .message-content {
          max-width: calc(100% - 80px);
          border-radius: 12px;
          padding: 12px 16px;
          position: relative;

          .message-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 6px;

            .message-sender {
              font-size: 12px;
              font-weight: 500;
            }

            .message-time {
              font-size: 11px;
              opacity: 0.8;
            }
          }

          .message-text {
            line-height: 1.6;
            word-wrap: break-word;

            :deep(code) {
              background: rgba(0, 0, 0, 0.1);
              padding: 2px 4px;
              border-radius: 3px;
              font-family: 'Courier New', monospace;
              font-size: 0.9em;
              cursor: pointer;

              &:hover {
                background: rgba(0, 0, 0, 0.2);
              }
            }

            :deep(pre) {
              background: #f3f4f6;
              padding: 12px;
              border-radius: 6px;
              overflow-x: auto;
              margin: 8px 0;

              code {
                background: none;
                padding: 0;
              }
            }
          }

          .message-meta {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-top: 8px;
            padding-top: 8px;
            border-top: 1px solid #f3f4f6;

            .token-count {
              font-size: 11px;
              color: #6b7280;
            }

            .message-actions {
              display: flex;
              gap: 4px;
              opacity: 0;
              transition: opacity 0.2s ease;
            }
          }

          &:hover .message-actions {
            opacity: 1;
          }
        }
      }

      .typing-indicator {
        display: flex;
        align-items: flex-start;
        margin-bottom: 16px;

        .typing-avatar {
          flex-shrink: 0;
        }

        .typing-content {
          margin-left: 12px;
          background: white;
          border: 1px solid #e5e7eb;
          border-radius: 12px;
          padding: 12px 16px;
          display: flex;
          align-items: center;
          gap: 8px;

          .typing-dots {
            display: flex;
            gap: 4px;

            span {
              width: 6px;
              height: 6px;
              background: #6b7280;
              border-radius: 50%;
              animation: typing 1.4s infinite ease-in-out;

              &:nth-child(1) { animation-delay: -0.32s; }
              &:nth-child(2) { animation-delay: -0.16s; }
            }
          }

          .typing-text {
            font-size: 12px;
            color: #6b7280;
          }
        }
      }
    }

    .input-container {
      border-top: 1px solid #e5e7eb;
      background: white;
      padding: 16px 20px;

      .input-wrapper {
        position: relative;

        .message-input {
          :deep(.el-textarea__inner) {
            border: 1px solid #d1d5db;
            border-radius: 12px;
            padding: 12px 16px;
            font-size: 14px;
            line-height: 1.5;
            resize: none;

            &:focus {
              border-color: #3b82f6;
              box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
            }
          }
        }

        .input-actions {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-top: 8px;

          .input-tools {
            display: flex;
            gap: 8px;

            .el-button {
              color: #6b7280;

              &:hover {
                color: #3b82f6;
              }
            }
          }

          .send-controls {
            display: flex;
            align-items: center;
            gap: 8px;
          }
        }
      }

      .quick-replies {
        margin-top: 12px;
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 8px;

        .quick-label {
          font-size: 12px;
          color: #6b7280;
          margin-right: 4px;
        }

        .quick-reply-tag {
          cursor: pointer;
          transition: all 0.2s ease;

          &:hover {
            background: #3b82f6;
            color: white;
          }
        }
      }
    }
  }
}

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(10px);
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

// 响应式设计
@media (max-width: 768px) {
  .chat-window {
    .chat-content {
      .chat-header {
        flex-direction: column;
        gap: 12px;
        padding: 12px 16px;

        .header-right {
          width: 100%;
          flex-wrap: wrap;
          justify-content: flex-start;

          .project-selector,
          .model-selector {
            width: 100%;
            max-width: 200px;
          }
        }
      }

      .messages-container {
        padding: 12px 16px;

        .message-item {
          &.user .message-content {
            margin-left: 40px;
          }

          &.assistant .message-content {
            margin-right: 40px;
          }
        }
      }

      .input-container {
        padding: 12px 16px;

        .input-wrapper {
          .input-actions {
            flex-direction: column;
            gap: 8px;

            .input-tools {
              order: 2;
              justify-content: center;
            }

            .send-controls {
              order: 1;
              width: 100%;
              justify-content: flex-end;
            }
          }
        }
      }
    }
  }
}
</style>