<template>
  <div class="nesma-chat">
    <!-- 警告栏 -->
    <warning-bar title="AI智能对话 - 与AI助手深度交流，获得专业的NESMA分析建议" />
    
    <!-- 状态指示器 -->
    <div class="gva-search-box">
      <div class="status-indicators">
        <el-tag 
          :type="wsConnected ? 'success' : 'warning'" 
          class="status-tag"
          effect="light"
        >
          <el-icon><Connection /></el-icon>
          {{ wsConnected ? 'WebSocket已连接' : 'WebSocket未连接' }}
        </el-tag>
        <el-tag type="info" class="status-tag" effect="light">
          <el-icon><Document /></el-icon>
          {{ sessions.length }} 个对话
        </el-tag>
      </div>
    </div>

    <!-- 主聊天区域 -->
    <div class="gva-table-box">
      <div class="chat-container">
        <!-- 左侧会话列表 -->
        <div class="chat-sidebar">
          <div class="sidebar-header">
            <el-text tag="h3" size="medium" style="font-weight:600;">对话历史</el-text>
            <el-button type="primary" size="small" round @click="createNewSession">
              <el-icon><Plus /></el-icon> 新建对话
            </el-button>
          </div>

          <div class="sidebar-content">
            <!-- 搜索框 -->
            <div class="search-section">
              <el-input
                v-model="sessionSearchQuery"
                placeholder="搜索对话..."
                class="search-input"
                clearable
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
              </el-input>
            </div>

            <!-- 会话列表 -->
            <div class="sessions-list">
              <div 
                v-for="session in filteredSessions"
                :key="session.id"
                class="session-item"
                :class="{ active: selectedSession?.id === session.id }"
                @click="selectSession(session)"
              >
                <div class="session-main">
                  <div class="session-title">{{ session.title }}</div>
                  <div class="session-preview">
                    {{ session.lastMessage || '暂无消息' }}
                  </div>
                  <div class="session-meta">
                    <span class="session-time">{{ formatTime(session.updatedAt) }}</span>
                    <el-tag 
                      :type="getModelType(session.modelName)" 
                      size="small" 
                      class="session-model"
                    >
                      {{ getModelName(session.modelName) }}
                    </el-tag>
                  </div>
                </div>
                <div class="session-actions">
                  <el-dropdown trigger="click" @command="handleSessionAction">
                    <el-button 
                      text 
                      size="small" 
                      @click.stop
                      class="action-btn"
                    >
                      <el-icon><MoreFilled /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item :command="{ action: 'edit', session }">
                          <el-icon><Edit /></el-icon>
                          重命名
                        </el-dropdown-item>
                        <el-dropdown-item :command="{ action: 'export', session }">
                          <el-icon><Download /></el-icon>
                          导出
                        </el-dropdown-item>
                        <el-dropdown-item :command="{ action: 'delete', session }" divided>
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
        </div>

        <!-- 右侧聊天区域 -->
        <div class="chat-main">
          <!-- 空状态 -->
          <div v-if="!selectedSession" class="empty-state">
            <div class="empty-content">
              <div class="empty-icon">
                <el-icon><ChatDotRound /></el-icon>
              </div>
              <h3 class="empty-title">开始新的AI对话</h3>
              <p class="empty-description">
                选择一个对话或创建新对话，与AI助手探讨NESMA相关问题
              </p>
              <el-button 
                type="primary" 
                @click="createNewSession"
                size="large"
              >
                <el-icon><Plus /></el-icon>
                创建新对话
              </el-button>
            </div>
          </div>

          <!-- 聊天界面 -->
          <div v-else class="chat-content">
            <!-- 聊天工具栏 -->
            <div class="gva-btn-list">
              <div class="toolbar-main">
                <div class="chat-title-meta">
                  <el-text tag="h3" size="large" style="font-weight:600;">
                    <el-icon style="vertical-align: middle; margin-right: 6px;"><ChatDotRound /></el-icon>
                    {{ selectedSession.title }}
                  </el-text>
                  <div class="chat-meta">
                    <el-tag :type="getModelType(selectedModel)" size="small">
                      {{ getModelName(selectedModel) }}
                    </el-tag>
                    <el-text type="info" size="small" style="margin-left: 8px;">
                      {{ messages.length }} 条消息
                    </el-text>
                  </div>
                </div>
                <div class="toolbar-right">
                  
                  <!-- 项目选择器 -->
                  <el-select 
                    v-model="selectedProject" 
                    label="项目"
                    label-width="50px"
                    placeholder="选择项目" 
                    size="small"
                    class="project-selector"
                    @change="onProjectChange"
                  >
                    <el-option
                      v-for="project in projects"
                      :key="project.ID"
                      :label="project.name"
                      :value="project.ID"
                    />
                  </el-select>
                  
                  <!-- 模型选择器 -->
                  <el-select 
                    v-model="selectedModel" 
                    label="模型"
                    label-width="50px"
                    placeholder="选择模型" 
                    size="small"
                    class="model-selector"
                    @change="onModelChange"
                  >
                    <el-option label="DeepSeek Chat" value="deepseek-chat" />
                    <el-option label="GPT-3.5 Turbo" value="gpt-3.5-turbo" />
                    <el-option label="GPT-4" value="gpt-4" />
                  </el-select>

                  <!-- 操作按钮 -->
                  <el-button 
                    @click="exportCurrentSession"
                    size="small"
                  >
                    <el-icon><Download /></el-icon>
                  </el-button>
                  <el-button 
                    @click="clearMessages"
                    size="small"
                  >
                    <el-icon><Delete /></el-icon>
                  </el-button>
                  <el-button 
                    @click="refreshData"
                    :loading="loading"
                    size="small"
                  >
                    <el-icon><Refresh /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>

            <!-- 消息列表 -->
            <div class="gva-form-box messages-wrapper">
              <div class="messages-container" ref="messagesContainer">
                <div
                  v-for="message in messages"
                  :key="message.id"
                  class="message-item"
                  :class="[message.role, { streaming: message.streaming }]"
                >
                  <div class="message-avatar">
                    <div class="avatar-wrapper" :class="message.role">
                      <el-icon>
                        <User v-if="message.role === 'user'" />
                        <MagicStick v-else />
                      </el-icon>
                    </div>
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
                    ></div>
                    <div v-if="message.role === 'assistant'" class="message-meta">
                      <el-icon><Cpu /></el-icon>
                      <span>Token: {{ message.tokenCount || 0 }}</span>
                    </div>
                  </div>
                </div>
                
                <!-- 正在输入指示器 -->
                <div v-if="isTyping" class="typing-indicator">
                  <div class="message-avatar">
                    <div class="avatar-wrapper assistant">
                      <el-icon><MagicStick /></el-icon>
                    </div>
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
            </div>

            <!-- 输入区域 -->
            <div class="gva-form-box input-section">
              <div class="input-toolbar">
                <div class="input-tools">
                  <el-button text size="small" class="tool-button">
                    <el-icon><Paperclip /></el-icon>
                    文件
                  </el-button>
                  <el-button text size="small" class="tool-button">
                    <el-icon><Picture /></el-icon>
                    图片
                  </el-button>
                  <el-button 
                    text 
                    size="small" 
                    class="tool-button"
                    @click="testMarkdownRendering"
                  >
                    <el-icon><MagicStick /></el-icon>
                    测试
                  </el-button>
                </div>
              </div>
              <div class="input-area">
                <el-input
                  v-model="inputMessage"
                  type="textarea"
                  :rows="3"
                  placeholder="输入您的问题... (Ctrl+Enter发送)"
                  :disabled="isTyping"
                  @keydown.enter.ctrl="sendMessage"
                  class="message-input"
                  resize="none"
                />
                <el-button
                  type="primary"
                  :loading="isTyping"
                  :disabled="!inputMessage.trim()"
                  @click="sendMessage"
                  class="send-button"
                  size="large"
                  round
                >
                  <el-icon><Promotion /></el-icon>
                  {{ isTyping ? '发送中...' : '发送' }}
                </el-button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onBeforeUnmount } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  ChatDotRound, Plus, Search, Edit, Delete, Download, MoreFilled, 
  Paperclip, Picture, Promotion, User, MagicStick, Cpu, Connection,
  Document, Refresh
} from '@element-plus/icons-vue'

// 导入warning-bar组件
import WarningBar from '@/components/warningBar/warningBar.vue'

// 导入原有的API和工具函数
import {
  getChatSessionList,
  createChatSession,
  getChatMessages,
  sendMessage as sendMessageAPI,
  updateChatSession,
  deleteChatSession,
  clearChatMessages,
  exportChatHistory
} from '@/api/chat'
import { getNesmaProjectList } from '@/api/nesma'
import { chatWebSocket } from '@/utils/websocket'
import { markdownRenderer } from '@/utils/markdown'
import { getToken } from '@/utils/auth'
import 'highlight.js/styles/github-dark.css'

// 响应式数据
const projects = ref([])
const selectedProject = ref(null)
const selectedModel = ref('deepseek-chat')
const sessions = ref([])
const selectedSession = ref(null)
const messages = ref([])
const sessionSearchQuery = ref('')
const inputMessage = ref('')
const isTyping = ref(false)
const loading = ref(false)

// WebSocket相关
const wsConnected = ref(false)
const currentStreamingMessage = ref(null)
const streamingContent = ref('')

// 保存WebSocket事件监听器引用，用于清理
const wsListeners = {
  connected: null,
  disconnected: null,
  message: null,
  complete: null,
  status: null,
  error: null
}

// 引用
const messagesContainer = ref(null)

// 计算属性
const filteredSessions = computed(() => {
  if (!sessionSearchQuery.value) {
    return sessions.value
  }
  return sessions.value.filter(session =>
    session.title.toLowerCase().includes(sessionSearchQuery.value.toLowerCase())
  )
})

// 获取模型类型
const getModelType = (modelName) => {
  const modelMap = {
    'deepseek-chat': 'success',
    'gpt-3.5-turbo': 'primary',
    'gpt-4': 'warning',
    'claude-3-sonnet': 'info'
  }
  return modelMap[modelName] || 'info'
}

// 获取模型显示名称
const getModelName = (modelName) => {
  const nameMap = {
    'deepseek-chat': 'DeepSeek',
    'gpt-3.5-turbo': 'GPT-3.5',
    'gpt-4': 'GPT-4',
    'claude-3-sonnet': 'Claude-3'
  }
  return nameMap[modelName] || modelName
}

// 这里包含原有的所有方法，但保持原有逻辑不变
// 监听项目数据变化，确保默认选择
watch(projects, (newProjects) => {
  if (newProjects.length > 0 && !selectedProject.value) {
    selectedProject.value = newProjects[0].ID
  }
}, { immediate: true })

// 监听模型选择，确保默认为deepseek
watch(selectedModel, (newModel) => {
  if (!newModel || newModel === 'gpt-3.5-turbo') {
    selectedModel.value = 'deepseek-chat'
  }
}, { immediate: true })

// 监听 selectedSession 和 projects 的变化，自动修正 selectedProject
watch([selectedSession, projects], ([session, projList]) => {
  if (session && projList.length) {
    const match = projList.find(p => p.ID === session.projectId);
    selectedProject.value = match ? match.ID : projList[0]?.ID || null;
  }
});

// 生命周期
onMounted(async () => {
  selectedModel.value = 'deepseek-chat'
  
  await Promise.all([
    loadProjects(),
    loadSessions()
  ])
  
  await nextTick()
  if (projects.value.length > 0 && !selectedProject.value) {
    selectedProject.value = projects.value[0].ID
  }
  
  await initWebSocket()
})

onBeforeUnmount(() => {
  cleanupWebSocket()
})

// 原有方法保持不变，这里只列出关键方法的声明
// WebSocket处理
const handleWebSocketMessage = (data) => {
  if (!currentStreamingMessage.value) return
  
  if (data.role === 'assistant') {
    streamingContent.value += data.content
    currentStreamingMessage.value.content = markdownRenderer.renderStreaming(streamingContent.value, false)
    
    nextTick(() => {
      scrollToBottom()
    })
  }
}

const handleStreamComplete = (data) => {
  if (currentStreamingMessage.value) {
    const finalContent = markdownRenderer.render(streamingContent.value)
    currentStreamingMessage.value.content = finalContent
    currentStreamingMessage.value.streaming = false
    currentStreamingMessage.value.tokenCount = data.metadata?.tokenCount || 0
    
    if (selectedSession.value) {
      selectedSession.value.lastMessage = streamingContent.value.slice(0, 50) + (streamingContent.value.length > 50 ? '...' : '')
      selectedSession.value.updatedAt = new Date().toISOString()
    }
    
    currentStreamingMessage.value = null
    streamingContent.value = ''
    isTyping.value = false
    
    nextTick(() => {
      scrollToBottom()
    })
  }
}

const handleStatusMessage = (data) => {
  console.log('状态消息:', data.content)
}

const handleWebSocketError = (data) => {
  console.error('WebSocket错误:', data)
  ElMessage.error(data.error || '发生未知错误')
  
  if (currentStreamingMessage.value) {
    currentStreamingMessage.value.streaming = false
    currentStreamingMessage.value.content += '\n\n[发送失败]'
    currentStreamingMessage.value = null
  }
  
  streamingContent.value = ''
  isTyping.value = false
}

// WebSocket初始化
const initWebSocket = async () => {
  try {
    const token = getToken()
    if (!token) {
      console.warn('未找到用户token，无法建立WebSocket连接')
      return
    }
    
    wsListeners.connected = () => {
      wsConnected.value = true
    }
    
    wsListeners.disconnected = (event) => {
      wsConnected.value = false
    }
    
    wsListeners.message = handleWebSocketMessage
    wsListeners.complete = handleStreamComplete
    wsListeners.status = handleStatusMessage
    wsListeners.error = handleWebSocketError
    
    chatWebSocket.on('connected', wsListeners.connected)
    chatWebSocket.on('disconnected', wsListeners.disconnected)
    chatWebSocket.on('message', wsListeners.message)
    chatWebSocket.on('complete', wsListeners.complete)
    chatWebSocket.on('status', wsListeners.status)
    chatWebSocket.on('error', wsListeners.error)
    
    await chatWebSocket.connect(token)
    
    await new Promise((resolve) => {
      const startTime = Date.now()
      const checkConnection = () => {
        if (chatWebSocket.isConnected()) {
          resolve()
        } else if (Date.now() - startTime > 5000) {
          resolve()
        } else {
          setTimeout(checkConnection, 100)
        }
      }
      checkConnection()
    })
    
  } catch (error) {
    console.error('WebSocket初始化失败:', error)
    wsConnected.value = false
    ElMessage.error('WebSocket连接失败，将使用SSE替代')
  }
}

const cleanupWebSocket = () => {
  if (wsListeners.connected) {
    chatWebSocket.off('connected', wsListeners.connected)
  }
  if (wsListeners.disconnected) {
    chatWebSocket.off('disconnected', wsListeners.disconnected)
  }
  if (wsListeners.message) {
    chatWebSocket.off('message', wsListeners.message)
  }
  if (wsListeners.complete) {
    chatWebSocket.off('complete', wsListeners.complete)
  }
  if (wsListeners.status) {
    chatWebSocket.off('status', wsListeners.status)
  }
  if (wsListeners.error) {
    chatWebSocket.off('error', wsListeners.error)
  }
  
  Object.keys(wsListeners).forEach(key => {
    wsListeners[key] = null
  })
  
  chatWebSocket.disconnect()
  chatWebSocket.removeAllListeners()
  wsConnected.value = false
}

// 数据加载方法 - 简化版本，保持原有逻辑
const loadProjects = async () => {
  try {
    console.log('开始加载项目列表...')
    const res = await getNesmaProjectList({
      page: 1,
      pageSize: 100
    })
    
    console.log('API响应:', res)
    
    let projectList = []
    
    if (res.data) {
      if (Array.isArray(res.data)) {
        projectList = res.data
      } else if (Array.isArray(res.data.projects)) {
        projectList = res.data.projects
      } else if (Array.isArray(res.data.list)) {
        projectList = res.data.list
      }
    }
    
    console.log('原始项目列表:', projectList)
    
    projects.value = projectList.filter(project => 
      project && (project.ID != null || project.id != null) && project.name != null
    ).map(project => ({
      ...project,
      id: project.ID || project.id,
    }))
    
    console.log('处理后的项目列表:', projects.value)
    
    if (projects.value.length > 0 && !selectedProject.value) {
      selectedProject.value = projects.value[0].ID
      console.log('设置默认项目:', selectedProject.value)
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
    projects.value = []
  }
}

const loadSessions = async () => {
  try {
    const res = await getChatSessionList({
      page: 1,
      pageSize: 100,
      isActive: true
    })
    
    let sessionList = []
    
    if (res.data) {
      if (Array.isArray(res.data)) {
        sessionList = res.data
      } else if (Array.isArray(res.data.sessions)) {
        sessionList = res.data.sessions
      } else if (Array.isArray(res.data.list)) {
        sessionList = res.data.list
      }
    }
    
    const processedSessions = sessionList.filter(session => 
      session && (session.ID != null || session.id != null) && session.title != null
    ).map(session => ({
      ...session,
      id: session.ID || session.id,
    }))
    
    for (const session of processedSessions) {
      try {
        const messagesRes = await getChatMessages(session.id, {
          page: 1,
          pageSize: 1
        })
        
        let messageList = []
        if (messagesRes.data) {
          if (Array.isArray(messagesRes.data)) {
            messageList = messagesRes.data
          } else if (Array.isArray(messagesRes.data.messages)) {
            messageList = messagesRes.data.messages
          } else if (Array.isArray(messagesRes.data.list)) {
            messageList = messagesRes.data.list
          }
        }
        
        if (messageList.length > 0) {
          const lastMessage = messageList[messageList.length - 1]
          const content = lastMessage.content || lastMessage.Content || ''
          session.lastMessage = content.length > 50 ? content.slice(0, 50) + '...' : content
        } else {
          session.lastMessage = '暂无消息'
        }
      } catch (error) {
        session.lastMessage = '暂无消息'
      }
    }
    
    sessions.value = processedSessions
  } catch (error) {
    console.error('加载会话列表失败:', error)
    sessions.value = []
  }
}

const loadMessages = async (sessionId) => {
  try {
    const res = await getChatMessages(sessionId, {
      page: 1,
      pageSize: 100
    })
    
    let messageList = []
    
    if (res.data) {
      if (Array.isArray(res.data)) {
        messageList = res.data
      } else if (Array.isArray(res.data.messages)) {
        messageList = res.data.messages
      } else if (Array.isArray(res.data.list)) {
        messageList = res.data.list
      }
    }
    
    messages.value = messageList.filter(message => {
      if (!message) return false
      if (!(message.ID != null || message.id != null)) return false
      if (!message.role) return false
      
      const hasContent = message.content != null || message.Content != null || message.text != null || message.Text != null
      if (!hasContent) {
        return false
      }
      
      return true
    }).map(message => {
      let content = message.content || message.Content || message.text || message.Text || ''
      
      return {
        ...message,
        id: message.ID || message.id,
        content: content
      }
    })
    
    await scrollToBottom()
  } catch (error) {
    console.error('加载消息失败:', error)
    messages.value = []
  }
}

// 其他方法保持原有逻辑不变...
const selectSession = async (session) => {
  selectedSession.value = session;
  selectedModel.value = session.modelName || 'deepseek-chat';
  // 如果 projects 还没加载，先加载
  if (!projects.value.length) {
    await loadProjects();
  }
  // 只有在 projects 有数据时再赋值
  if (projects.value.some(p => p.ID === session.projectId)) {
    selectedProject.value = session.projectId;
  } else {
    selectedProject.value = projects.value[0]?.ID || null;
  }
  await loadMessages(session.id);
};

const createNewSession = async () => {
  try {
    if (!selectedProject.value && projects.value.length > 0) {
      selectedProject.value = projects.value[0].ID
    }
    
    if (!selectedModel.value || selectedModel.value === 'gpt-3.5-turbo') {
      selectedModel.value = 'deepseek-chat'
    }
    
    const sessionData = {
      title: `新对话 ${new Date().toLocaleString()}`,
      projectId: selectedProject.value,
      modelName: selectedModel.value
    }
    
    const res = await createChatSession(sessionData)
    
    const newSession = {
      ...res.data,
      id: res.data.ID || res.data.id,
      lastMessage: '暂无消息'
    }
    
    sessions.value.unshift(newSession)
    await selectSession(newSession)
    
  } catch (error) {
    console.error('创建会话失败:', error)
    ElMessage.error('创建会话失败: ' + (error.message || '未知错误'))
  }
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isTyping.value) return
  
  if (!selectedProject.value) {
    ElMessage.warning('请先选择项目')
    return
  }
  
  const message = inputMessage.value.trim()
  inputMessage.value = ''
  isTyping.value = true

  const userMessage = {
    id: Date.now(),
    role: 'user',
    content: message,
    createdAt: new Date().toISOString()
  }
  messages.value.push(userMessage)

  try {
    if (wsConnected.value) {
      const aiMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: '',
        createdAt: new Date().toISOString(),
        streaming: true
      }
      messages.value.push(aiMessage)
      currentStreamingMessage.value = aiMessage
      streamingContent.value = ''

      chatWebSocket.sendMessage(
        selectedSession.value.id,
        message,
        selectedModel.value,
        selectedProject.value
      )
    } else {
      const aiMessage = {
        id: Date.now() + 1,
        role: 'assistant',
        content: '',
        createdAt: new Date().toISOString(),
        streaming: true
      }
      messages.value.push(aiMessage)
      currentStreamingMessage.value = aiMessage
      streamingContent.value = ''

      await sendMessageWithSSE(
        selectedSession.value.id,
        message,
        selectedModel.value,
        selectedProject.value
      )
    }
    
    await scrollToBottom()
    
    if (selectedSession.value) {
      selectedSession.value.lastMessage = message
      selectedSession.value.updatedAt = new Date().toISOString()
    }
    
  } catch (error) {
    console.error('发送消息失败:', error)
    
    if (error.response) {
      ElMessage.error(`发送消息失败: ${error.response.data.msg || error.message}`)
    } else {
      ElMessage.error(`发送消息失败: ${error.message}`)
    }
    
    messages.value = messages.value.filter(m => m.id !== userMessage.id)
    
    if (currentStreamingMessage.value) {
      messages.value = messages.value.filter(m => m.id !== currentStreamingMessage.value.id)
      currentStreamingMessage.value = null
      streamingContent.value = ''
    }
    
    isTyping.value = false
  }
}

// 使用SSE流式API发送消息
const sendMessageWithSSE = async (sessionId, message, modelName, projectId) => {
  try {
    const token = getToken()
    if (!token) {
      throw new Error('用户未登录')
    }

    const baseURL = 'http://localhost:8888'
    const params = new URLSearchParams({
      sessionId: sessionId || '',
      projectId: projectId || '',
      message: message,
      modelName: modelName || 'deepseek-chat',
      token: token
    })
    
    const url = `${baseURL}/api/chat/stream?${params.toString()}`
    const eventSource = new EventSource(url)

    eventSource.onopen = () => {
      console.log('SSE连接已建立')
    }

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)

        if (data.type === 'content' && currentStreamingMessage.value) {
          streamingContent.value += data.content
          currentStreamingMessage.value.content = markdownRenderer.renderStreaming(streamingContent.value, false)
          
          nextTick(() => {
            scrollToBottom()
          })
        } else if (data.type === 'done') {
          if (currentStreamingMessage.value) {
            const finalContent = markdownRenderer.render(streamingContent.value)
            currentStreamingMessage.value.content = finalContent
            currentStreamingMessage.value.streaming = false
            currentStreamingMessage.value.tokenCount = data.tokenCount || 0
            
            if (selectedSession.value) {
              selectedSession.value.lastMessage = streamingContent.value.slice(0, 50) + (streamingContent.value.length > 50 ? '...' : '')
              selectedSession.value.updatedAt = new Date().toISOString()
            }
            
            currentStreamingMessage.value = null
            streamingContent.value = ''
            isTyping.value = false
            
            nextTick(() => {
              scrollToBottom()
            })
          }
          
          eventSource.close()
        }
      } catch (error) {
        console.error('解析SSE消息失败:', error)
      }
    }

    eventSource.onerror = (error) => {
      console.error('SSE连接错误:', error)
      eventSource.close()
      
      if (currentStreamingMessage.value) {
        currentStreamingMessage.value.streaming = false
        currentStreamingMessage.value.content += '\n\n[连接中断]'
        currentStreamingMessage.value = null
      }
      
      streamingContent.value = ''
      isTyping.value = false
      
      ElMessage.error('流式连接中断，请重试')
    }

    setTimeout(() => {
      if (eventSource.readyState !== EventSource.CLOSED) {
        eventSource.close()
        if (isTyping.value) {
          ElMessage.error('请求超时，请重试')
          isTyping.value = false
        }
      }
    }, 120000)

  } catch (error) {
    console.error('SSE请求失败:', error)
    throw error
  }
}

const handleSessionAction = async ({ action, session }) => {
  if (action === 'edit') {
    await editSession(session)
  } else if (action === 'delete') {
    await deleteSession(session)
  } else if (action === 'export') {
    await exportSession(session)
  }
}

const editSession = async (session) => {
  const { value: newTitle } = await ElMessageBox.prompt('请输入新的对话标题', '重命名对话', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    inputValue: session.title
  })
  
  if (newTitle) {
    try {
      await updateChatSession(session.id, { title: newTitle })
      session.title = newTitle
      ElMessage.success('对话重命名成功')
    } catch (error) {
      console.error('更新会话失败:', error)
      ElMessage.error('更新会话失败')
    }
  }
}

const deleteSession = async (session) => {
  try {
    await ElMessageBox.confirm('确定要删除这个对话吗？', '删除确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await deleteChatSession(session.id)
    sessions.value = sessions.value.filter(s => s.id !== session.id)
    
    if (selectedSession.value?.id === session.id) {
      selectedSession.value = null
      messages.value = []
    }
    
    ElMessage.success('对话删除成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除会话失败:', error)
      ElMessage.error('删除会话失败')
    }
  }
}

const exportSession = async (session) => {
  try {
    const res = await exportChatHistory(session.id)
    const blob = new Blob([res.data], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `chat_history_${session.title}.txt`
    a.click()
    URL.revokeObjectURL(url)
    
    ElMessage.success('导出成功')
  } catch (error) {
    console.error('导出失败:', error)
    ElMessage.error('导出失败')
  }
}

const clearMessages = async () => {
  try {
    await ElMessageBox.confirm('确定要清空当前对话的所有消息吗？', '清空确认', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    await clearChatMessages(selectedSession.value.id)
    messages.value = []
    ElMessage.success('消息清空成功')
  } catch (error) {
    if (error !== 'cancel') {
      console.error('清空消息失败:', error)
      ElMessage.error('清空消息失败')
    }
  }
}

const exportCurrentSession = async () => {
  if (!selectedSession.value) {
    ElMessage.warning('请选择一个对话')
    return
  }
  
  await exportSession(selectedSession.value)
}

const refreshData = async () => {
  loading.value = true
  try {
    await Promise.all([
      loadProjects(),
      loadSessions()
    ])
    ElMessage.success('数据刷新成功')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    loading.value = false
  }
}

const formatTime = (time) => {
  if (!time) return ''
  try {
    const date = new Date(time)
    const now = new Date()
    const diffTime = now - date
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24))
    
    if (diffDays === 0) {
      return date.toLocaleTimeString('zh-CN', { 
        hour: '2-digit', 
        minute: '2-digit' 
      })
    } else if (diffDays === 1) {
      return '昨天'
    } else if (diffDays < 7) {
      return `${diffDays}天前`
    } else {
      return date.toLocaleDateString('zh-CN', { 
        month: '2-digit', 
        day: '2-digit' 
      })
    }
  } catch (error) {
    return '时间格式错误'
  }
}

const formatMessage = (content) => {
  if (!content) {
    return ''
  }
  
  if (typeof content !== 'string') {
    if (typeof content === 'object') {
      const possibleContent = content.content || content.Content || content.text || content.Text || content.message || content.Message
      if (possibleContent && typeof possibleContent === 'string') {
        return markdownRenderer.render(possibleContent)
      }
      return markdownRenderer.render(JSON.stringify(content, null, 2))
    }
    return markdownRenderer.render(String(content))
  }
  
  try {
    return markdownRenderer.render(content)
  } catch (error) {
    console.error('Markdown渲染失败:', error)
    return content
  }
}

const handleMessageClick = (event) => {
  const target = event.target
  
  const copyBtn = target.closest('.copy-code-btn')
  if (copyBtn) {
    const codeId = copyBtn.getAttribute('data-code-id')
    if (codeId) {
      copyCode(codeId)
    }
  }
}

const copyCode = async (elementId) => {
  try {
    const codeElement = document.getElementById(elementId)
    if (!codeElement) {
      console.error('找不到代码元素:', elementId)
      return
    }
    
    const text = codeElement.textContent || codeElement.innerText
    
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
    } else {
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
    }
    
    const btn = codeElement.parentElement.querySelector('.copy-code-btn')
    if (btn) {
      const originalText = btn.innerHTML
      btn.innerHTML = `
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="20,6 9,17 4,12"></polyline>
        </svg>
        已复制
      `
      btn.style.background = '#10b981'
      btn.style.borderColor = '#059669'
      btn.style.color = '#ffffff'
      btn.style.transform = 'translateY(-1px)'
      btn.style.boxShadow = '0 4px 12px rgba(16, 185, 129, 0.4)'
      
      setTimeout(() => {
        btn.innerHTML = originalText
        btn.style.background = ''
        btn.style.borderColor = ''
        btn.style.color = ''
        btn.style.transform = ''
        btn.style.boxShadow = ''
      }, 2000)
    }
    
    ElMessage.success('代码已复制到剪贴板')
  } catch (error) {
    console.error('复制失败:', error)
    ElMessage.error('复制失败，请手动复制')
  }
}

const scrollToBottom = async () => {
  await nextTick()
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

const onProjectChange = (projectId) => {
  selectedProject.value = projectId
}

const onModelChange = (modelValue) => {
  selectedModel.value = modelValue
}

// 测试Markdown渲染
const testMarkdownRendering = () => {
  if (!selectedSession.value) {
    ElMessage.warning('请先选择一个对话')
    return
  }
  
  const testContent = `# 🚀 AI对话测试

## ✨ 基本功能测试

这是一个**测试消息**，用于验证*Markdown渲染*功能。

### 代码示例：
\`\`\`javascript
// 简单的JavaScript代码
function hello() {
  console.log("Hello, NESMA!");
  return "AI助手工作正常";
}
\`\`\`

### 表格示例：
| 功能 | 状态 | 说明 |
|------|------|------|
| WebSocket | ✅ | 实时通信 |
| Markdown | ✅ | 富文本渲染 |
| 代码高亮 | ✅ | 语法高亮 |

> **提示：** 所有功能正常工作！

---

**测试完成** 🎉`

  const testMessage = {
    id: Date.now(),
    role: 'assistant',
    content: testContent,
    createdAt: new Date().toISOString(),
    tokenCount: 150
  }
  
  messages.value.push(testMessage)
  
  nextTick(() => {
    scrollToBottom()
  })
  
  ElMessage.success('测试内容已添加')
}
</script>

<style lang="scss" scoped>
.nesma-chat {
  padding: 20px;
  background: #f5f7fa;
  min-height: calc(100vh - 120px);
  
  // 状态指示器
  .status-indicators {
    display: flex;
    gap: 12px;
    
    .status-tag {
      .el-icon {
        margin-right: 4px;
      }
    }
  }

  // 聊天容器
  .chat-container {
    display: flex;
    height: calc(100vh - 300px);
    overflow: hidden;
  }

  // 左侧会话列表
  .chat-sidebar {
    width: 320px;
    border-right: 1px solid #e4e7ed;
    background: #ffffff;
    display: flex;
    flex-direction: column;
    
    .sidebar-header {
      padding: 16px;
      border-bottom: 1px solid #e4e7ed;
      display: flex;
      justify-content: space-between;
      align-items: center;
      background: #f8f9fa;
      
      .sidebar-title {
        margin: 0;
        color: #303133;
        font-size: 16px;
        font-weight: 600;
      }
    }
    
    .sidebar-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      
      .search-section {
        padding: 12px;
        border-bottom: 1px solid #e4e7ed;
      }
      
      .sessions-list {
        flex: 1;
        overflow-y: auto;
        padding: 8px;
        
        .session-item {
          padding: 12px;
          margin-bottom: 8px;
          background: #ffffff;
          border-radius: 8px;
          cursor: pointer;
          transition: all 0.3s ease;
          border: 1px solid #e4e7ed;
          display: flex;
          justify-content: space-between;
          align-items: flex-start;
          
          &:hover {
            border-color: #409eff;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            transform: translateY(-1px);
          }
          
          &.active {
            background: #409eff;
            color: #ffffff;
            border-color: #409eff;
            
            .session-model {
              background: rgba(255, 255, 255, 0.2);
              border-color: rgba(255, 255, 255, 0.3);
              color: #ffffff;
            }
          }
          
          .session-main {
            flex: 1;
            min-width: 0;
            
            .session-title {
              font-weight: 600;
              font-size: 14px;
              margin-bottom: 4px;
              overflow: hidden;
              text-overflow: ellipsis;
              white-space: nowrap;
            }
            
            .session-preview {
              font-size: 12px;
              opacity: 0.7;
              overflow: hidden;
              text-overflow: ellipsis;
              white-space: nowrap;
              margin-bottom: 8px;
            }
            
            .session-meta {
              display: flex;
              justify-content: space-between;
              align-items: center;
              
              .session-time {
                font-size: 12px;
                opacity: 0.6;
              }
              
              .session-model {
                font-size: 12px;
              }
            }
          }
          
          .session-actions {
            opacity: 0;
            transition: opacity 0.3s ease;
            
            .action-btn {
              padding: 4px !important;
              min-height: auto !important;
            }
          }
          
          &:hover .session-actions {
            opacity: 1;
          }
        }
      }
    }
  }

  // 右侧聊天区域
  .chat-main {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
    
    .empty-state {
      flex: 1;
      display: flex;
      align-items: center;
      justify-content: center;
      
      .empty-content {
        text-align: center;
        max-width: 400px;
        
        .empty-icon {
          font-size: 4rem;
          color: #c0c4cc;
          margin-bottom: 24px;
        }
        
        .empty-title {
          font-size: 24px;
          font-weight: 600;
          color: #303133;
          margin: 0 0 12px 0;
        }
        
        .empty-description {
          font-size: 16px;
          color: #606266;
          margin: 0 0 24px 0;
          line-height: 1.6;
        }
      }
    }
    
    .chat-content {
      flex: 1;
      display: flex;
      flex-direction: column;
      overflow: hidden;
      
      .toolbar-main {
        display: flex;
        align-items: center;
        justify-content: space-between;
        padding: 8px 0;

        .chat-title-meta {
          display: flex;
          align-items: center;
          gap: 16px;

          .chat-title {
            font-size: 18px;
            font-weight: 600;
            margin: 0;
          }
          .chat-meta {
            display: flex;
            align-items: center;
            gap: 12px;
            .message-count {
              font-size: 12px;
              color: #909399;
            }
          }
        }
        
        .toolbar-right {
          display: flex;
          align-items: center;
          gap: 12px;
          
          .project-selector,
          .model-selector {
            min-width: 140px;
            display: inline-block;
            
            :deep(.el-select) {
              width: 100%;
              display: block;
            }
            
            :deep(.el-input__wrapper) {
              background: #ffffff;
              border: 1px solid #dcdfe6;
              border-radius: 4px;
              display: flex;
              align-items: center;
              
              &:hover {
                border-color: #c0c4cc;
              }
              
              &.is-focus {
                border-color: #409eff;
              }
            }
            
            :deep(.el-input__inner) {
              font-size: 14px;
              color: #606266;
              height: 32px;
              line-height: 32px;
            }
            
            :deep(.el-select-dropdown) {
              border-radius: 4px;
              box-shadow: 0 2px 12px 0 rgba(0, 0, 0, 0.1);
              z-index: 2000;
            }
            
            :deep(.el-select-dropdown__item) {
              font-size: 14px;
              padding: 8px 12px;
              
              &:hover {
                background-color: #f5f7fa;
              }
              
              &.selected {
                background-color: #409eff;
                color: #ffffff;
              }
            }
          }
        }
      }
      
      .messages-wrapper {
        flex: 1;
        overflow: hidden;
        
        .messages-container {
          height: 400px;
          overflow-y: auto;
          
          .message-item {
            display: flex;
            margin-bottom: 24px;
            gap: 12px;
            animation: slideIn 0.3s ease;
            
            &.user {
              flex-direction: row-reverse;
              
              .message-content {
                background: #409eff;
                color: #ffffff;
                border-radius: 18px 18px 4px 18px;
              }
              
              .avatar-wrapper {
                background: #409eff;
              }
            }
            
            &.assistant {
              .message-content {
                background: #ffffff;
                border: 1px solid #e4e7ed;
                border-radius: 18px 18px 18px 4px;
              }
              
              .avatar-wrapper {
                background: #67c23a;
              }
            }
            
            &.streaming .message-content {
              position: relative;
              
              &::after {
                content: '';
                display: inline-block;
                width: 2px;
                height: 16px;
                background: #409eff;
                animation: blink 1s infinite;
                margin-left: 4px;
                vertical-align: baseline;
              }
            }
            
            .message-avatar {
              .avatar-wrapper {
                width: 40px;
                height: 40px;
                border-radius: 50%;
                display: flex;
                align-items: center;
                justify-content: center;
                color: #ffffff;
                font-size: 18px;
                box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
              }
            }
            
            .message-content {
              max-width: 70%;
              min-width: 200px;
              padding: 16px;
              word-wrap: break-word;
              box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
              
              .message-header {
                display: flex;
                justify-content: space-between;
                align-items: center;
                margin-bottom: 8px;
                font-size: 12px;
                opacity: 0.7;
                
                .message-sender {
                  font-weight: 600;
                }
              }
              
              .message-text {
                line-height: 1.6;
                font-size: 14px;
                word-break: break-word;
                
                // Markdown样式
                h1, h2, h3, h4, h5, h6 {
                  margin: 16px 0 8px 0;
                  color: inherit;
                  font-weight: 600;
                }
                
                p {
                  margin: 8px 0;
                  line-height: 1.6;
                }
                
                ul, ol {
                  margin: 8px 0;
                  padding-left: 20px;
                }
                
                li {
                  margin: 4px 0;
                }
                
                blockquote {
                  margin: 16px 0;
                  padding: 12px 16px;
                  border-left: 4px solid #409eff;
                  background: rgba(64, 158, 255, 0.1);
                  border-radius: 8px;
                }
                
                table {
                  border-collapse: collapse;
                  width: 100%;
                  margin: 16px 0;
                  font-size: 12px;
                  border-radius: 8px;
                  overflow: hidden;
                  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
                }
                
                th, td {
                  border: 1px solid #e4e7ed;
                  padding: 8px 12px;
                  text-align: left;
                }
                
                th {
                  background: #f8f9fa;
                  font-weight: 600;
                }
                
                // 代码块样式
                .code-block-container {
                  margin: 16px 0;
                  border-radius: 8px;
                  overflow: hidden;
                  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
                  background: #1e1e1e;
                  border: 1px solid #333;
                }
                
                .code-block-header {
                  display: flex;
                  justify-content: space-between;
                  align-items: center;
                  padding: 12px 16px;
                  background: #2d2d30;
                  border-bottom: 1px solid #333;
                  
                  .language-label {
                    font-size: 12px;
                    color: #cccccc;
                    font-weight: 600;
                    text-transform: uppercase;
                    letter-spacing: 0.5px;
                  }
                  
                  .copy-code-btn {
                    display: flex;
                    align-items: center;
                    gap: 4px;
                    padding: 6px 12px;
                    background: #404040;
                    color: #cccccc;
                    border: 1px solid #666;
                    border-radius: 4px;
                    font-size: 12px;
                    cursor: pointer;
                    transition: all 0.3s ease;
                    
                    &:hover {
                      background: #1e40af;
                      color: #ffffff;
                      border-color: #3b82f6;
                      transform: translateY(-1px);
                      box-shadow: 0 4px 12px rgba(59, 130, 246, 0.3);
                    }
                    
                    svg {
                      width: 16px;
                      height: 16px;
                    }
                  }
                }
                
                .code-block {
                  background: #1e1e1e;
                  color: #cccccc;
                  padding: 16px;
                  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
                  font-size: 14px;
                  line-height: 1.6;
                  overflow-x: auto;
                  white-space: pre;
                }
                
                .inline-code {
                  background: rgba(128, 128, 128, 0.1);
                  color: #e74c3c;
                  padding: 2px 4px;
                  border-radius: 4px;
                  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
                  font-size: 0.9em;
                  font-weight: 500;
                }
              }
              
              .message-meta {
                margin-top: 8px;
                font-size: 12px;
                opacity: 0.6;
                display: flex;
                align-items: center;
                gap: 4px;
              }
            }
          }
          
          .typing-indicator {
            display: flex;
            gap: 12px;
            margin-bottom: 24px;
            animation: slideIn 0.3s ease;
            
            .typing-content {
              background: #ffffff;
              border: 1px solid #e4e7ed;
              border-radius: 18px 18px 18px 4px;
              padding: 16px;
              box-shadow: 0 1px 4px rgba(0, 0, 0, 0.1);
              display: flex;
              align-items: center;
              gap: 12px;
              
              .typing-dots {
                display: flex;
                gap: 3px;
                
                span {
                  width: 6px;
                  height: 6px;
                  border-radius: 50%;
                  background: #409eff;
                  animation: typing 1.4s infinite;
                  
                  &:nth-child(2) {
                    animation-delay: 0.2s;
                  }
                  
                  &:nth-child(3) {
                    animation-delay: 0.4s;
                  }
                }
              }
              
              .typing-text {
                color: #909399;
                font-size: 12px;
              }
            }
          }
        }
      }
      
      .input-section {
        .input-toolbar {
          margin-bottom: 12px;
          
          .input-tools {
            display: flex;
            gap: 8px;
            
            .tool-button {
              color: #909399;
              
              &:hover {
                color: #409eff;
                transform: translateY(-1px);
              }
            }
          }
        }
        
        .input-area {
          display: flex;
          gap: 12px;
          align-items: flex-end;
          
          .message-input {
            flex: 1;
          }
          
          .send-button {
            padding: 12px 20px;
            font-weight: 600;
            
            &:hover {
              transform: translateY(-2px);
              box-shadow: 0 4px 12px rgba(64, 158, 255, 0.3);
            }
          }
        }
      }
    }
  }
}

@keyframes slideIn {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

@keyframes blink {
  0%, 50% { opacity: 1; }
  51%, 100% { opacity: 0; }
}

@keyframes typing {
  0%, 60%, 100% {
    transform: translateY(0);
  }
  30% {
    transform: translateY(-6px);
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .nesma-chat {
    .chat-container {
      height: calc(100vh - 320px);
    }
    
    .chat-sidebar {
      width: 280px;
    }
  }
}

@media (max-width: 768px) {
  .nesma-chat {
    padding: 12px;
    
    .chat-container {
      flex-direction: column;
      height: calc(100vh - 200px);
    }
    
    .chat-sidebar {
      width: 100%;
      height: 200px;
    }
    
    .chat-main {
      .chat-content {
        .gva-btn-list {
          flex-direction: column;
          gap: 12px;
          
          .toolbar-right {
            width: 100%;
            justify-content: space-between;
          }
        }
        
        .messages-wrapper {
          .messages-container {
            height: 300px;
            
            .message-item .message-content {
              max-width: 85%;
              min-width: 150px;
            }
          }
        }
        
        .input-section {
          .input-area {
            flex-direction: column;
            gap: 12px;
            
            .send-button {
              width: 100%;
              align-self: stretch;
            }
          }
        }
      }
    }
  }
}
</style>