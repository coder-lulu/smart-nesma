import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import * as aiApi from '@/api/ai-services'
import * as chatApi from '@/api/chat'
import * as recommendationApi from '@/api/recommendation'

export const useAIStore = defineStore('ai', () => {
  // ==================== 状态管理 ====================
  
  // AI服务状态
  const serviceStatus = ref({
    deepseek: { connected: false, lastCheck: null },
    openai: { connected: false, lastCheck: null },
    overall: 'unknown' // unknown, healthy, degraded, down
  })
  
  // 聊天会话管理
  const chatSessions = ref([])
  const currentSession = ref(null)
  const messages = ref([])
  const isTyping = ref(false)
  const wsConnection = ref(null)
  const wsConnected = ref(false)
  
  // 分析任务管理
  const analysisTasks = ref([])
  const currentAnalysisTask = ref(null)
  const analysisProgress = ref({
    taskId: null,
    status: 'idle', // idle, running, completed, failed
    progress: 0,
    message: '',
    startTime: null,
    estimatedTime: null
  })
  
  // 分析结果
  const analysisResults = ref({})
  const selectedRequirements = ref([])
  const analysisConfig = ref({
    model: 'deepseek-chat',
    temperature: 0.7,
    maxTokens: 2000,
    enableKnowledgeBase: true,
    analysisDepth: 'standard' // basic, standard, deep
  })
  
  // AI推荐系统
  const recommendations = ref([])
  const adoptedRecommendations = ref([])
  const dismissedRecommendations = ref([])
  
  // 项目上下文
  const selectedProject = ref(null)
  const selectedCycle = ref(null)
  const selectedVersion = ref(null)
  
  // UI状态
  const loading = ref({
    sessions: false,
    messages: false,
    analysis: false,
    recommendations: false,
    export: false
  })
  
  const errors = ref({
    connection: null,
    analysis: null,
    chat: null,
    general: null
  })
  
  // ==================== 计算属性 ====================
  
  const isAIServiceHealthy = computed(() => {
    return serviceStatus.value.overall === 'healthy'
  })
  
  const hasActiveSession = computed(() => {
    return currentSession.value !== null
  })
  
  const isAnalysisRunning = computed(() => {
    return analysisProgress.value.status === 'running'
  })
  
  const pendingRecommendations = computed(() => {
    return recommendations.value.filter(rec => 
      !adoptedRecommendations.value.includes(rec.id) &&
      !dismissedRecommendations.value.includes(rec.id)
    )
  })
  
  const analysisStatistics = computed(() => {
    const results = analysisResults.value
    if (!results || Object.keys(results).length === 0) {
      return {
        totalRequirements: 0,
        analyzedRequirements: 0,
        totalFunctionPoints: 0,
        averageConfidence: 0,
        typeDistribution: {},
        complexityDistribution: {}
      }
    }
    
    // 计算统计数据
    const analyzed = Object.values(results)
    return {
      totalRequirements: selectedRequirements.value.length,
      analyzedRequirements: analyzed.length,
      totalFunctionPoints: analyzed.reduce((sum, req) => sum + (req.afp || 0), 0),
      averageConfidence: analyzed.length > 0 
        ? analyzed.reduce((sum, req) => sum + (req.confidence || 0), 0) / analyzed.length 
        : 0,
      typeDistribution: analyzed.reduce((dist, req) => {
        const type = req.function_type || 'unknown'
        dist[type] = (dist[type] || 0) + 1
        return dist
      }, {}),
      complexityDistribution: analyzed.reduce((dist, req) => {
        const complexity = req.complexity || 'unknown'
        dist[complexity] = (dist[complexity] || 0) + 1
        return dist
      }, {})
    }
  })
  
  // ==================== AI服务管理 ====================
  
  const checkAIServiceStatus = async () => {
    try {
      const response = await aiApi.getAIServiceStatus()
      serviceStatus.value = {
        ...response.data,
        overall: response.data.overall || 'healthy'
      }
      errors.value.connection = null
      return true
    } catch (error) {
      console.error('AI服务状态检查失败:', error)
      serviceStatus.value.overall = 'down'
      errors.value.connection = error.message
      return false
    }
  }
  
  const switchAIModel = async (modelId) => {
    try {
      await aiApi.switchAIModel(modelId)
      analysisConfig.value.model = modelId
      ElMessage.success(`已切换到模型: ${modelId}`)
    } catch (error) {
      console.error('切换AI模型失败:', error)
      ElMessage.error('切换AI模型失败')
      throw error
    }
  }
  
  const updateAIConfig = (config) => {
    analysisConfig.value = { ...analysisConfig.value, ...config }
  }
  
  // ==================== 聊天会话管理 ====================
  
  const loadChatSessions = async () => {
    loading.value.sessions = true
    try {
      const response = await chatApi.getChatSessionList()
      chatSessions.value = response.data || []
    } catch (error) {
      console.error('加载聊天会话失败:', error)
      errors.value.chat = error.message
    } finally {
      loading.value.sessions = false
    }
  }
  
  const createChatSession = async (sessionData) => {
    try {
      const response = await chatApi.createChatSession(sessionData)
      const newSession = response.data
      chatSessions.value.unshift(newSession)
      await selectChatSession(newSession)
      return newSession
    } catch (error) {
      console.error('创建聊天会话失败:', error)
      ElMessage.error('创建聊天会话失败')
      throw error
    }
  }
  
  const selectChatSession = async (session) => {
    if (currentSession.value?.id === session.id) return
    
    // 关闭当前WebSocket连接
    if (wsConnection.value) {
      wsConnection.value.close()
    }
    
    currentSession.value = session
    await loadChatMessages(session.id)
    await connectWebSocket(session.id)
  }
  
  const loadChatMessages = async (sessionId) => {
    loading.value.messages = true
    try {
      const response = await chatApi.getChatMessages(sessionId)
      messages.value = response.data || []
    } catch (error) {
      console.error('加载聊天消息失败:', error)
      errors.value.chat = error.message
    } finally {
      loading.value.messages = false
    }
  }
  
  const sendChatMessage = async (content, attachments = []) => {
    if (!currentSession.value) {
      throw new Error('没有激活的聊天会话')
    }
    
    const messageData = {
      session_id: currentSession.value.id,
      content,
      attachments,
      model: analysisConfig.value.model,
      project_context: {
        project_id: selectedProject.value?.id,
        cycle_id: selectedCycle.value?.id,
        version_id: selectedVersion.value?.id
      }
    }
    
    try {
      // 添加用户消息到本地
      const userMessage = {
        id: Date.now(),
        role: 'user',
        content,
        createdAt: new Date().toISOString(),
        status: 'sending'
      }
      messages.value.push(userMessage)
      
      // 设置AI正在输入状态
      isTyping.value = true
      
      // 发送消息
      const response = await chatApi.sendMessage(messageData)
      
      // 更新用户消息状态
      userMessage.status = 'sent'
      userMessage.id = response.data.user_message_id
      
      // AI响应会通过WebSocket接收
      
    } catch (error) {
      console.error('发送消息失败:', error)
      isTyping.value = false
      
      // 更新消息状态为失败
      const lastMessage = messages.value[messages.value.length - 1]
      if (lastMessage.role === 'user') {
        lastMessage.status = 'failed'
        lastMessage.error = error.message
      }
      
      ElMessage.error('发送消息失败')
      throw error
    }
  }
  
  const connectWebSocket = async (sessionId) => {
    try {
      wsConnection.value = aiApi.createAIWebSocketConnection(sessionId)
      
      wsConnection.value.onopen = () => {
        wsConnected.value = true
        console.log('WebSocket连接已建立')
      }
      
      wsConnection.value.onmessage = (event) => {
        const data = JSON.parse(event.data)
        handleWebSocketMessage(data)
      }
      
      wsConnection.value.onclose = () => {
        wsConnected.value = false
        console.log('WebSocket连接已关闭')
      }
      
      wsConnection.value.onerror = (error) => {
        console.error('WebSocket连接错误:', error)
        wsConnected.value = false
      }
      
    } catch (error) {
      console.error('建立WebSocket连接失败:', error)
    }
  }
  
  const handleWebSocketMessage = (data) => {
    switch (data.type) {
      case 'message':
        // 新消息
        messages.value.push(data.message)
        isTyping.value = false
        break
        
      case 'typing':
        // AI正在输入
        isTyping.value = data.typing
        break
        
      case 'message_update':
        // 消息更新（流式响应）
        const messageIndex = messages.value.findIndex(m => m.id === data.message_id)
        if (messageIndex > -1) {
          messages.value[messageIndex] = { ...messages.value[messageIndex], ...data.updates }
        }
        break
        
      case 'error':
        // 错误消息
        console.error('WebSocket错误:', data.error)
        isTyping.value = false
        break
    }
  }
  
  const deleteChatSession = async (sessionId) => {
    try {
      await chatApi.deleteChatSession(sessionId)
      chatSessions.value = chatSessions.value.filter(s => s.id !== sessionId)
      
      if (currentSession.value?.id === sessionId) {
        currentSession.value = null
        messages.value = []
        if (wsConnection.value) {
          wsConnection.value.close()
        }
      }
      
      ElMessage.success('聊天会话已删除')
    } catch (error) {
      console.error('删除聊天会话失败:', error)
      ElMessage.error('删除聊天会话失败')
      throw error
    }
  }
  
  // ==================== 需求分析管理 ====================
  
  const startRequirementAnalysis = async (analysisRequest) => {
    loading.value.analysis = true
    try {
      const response = await aiApi.startRequirementAnalysis(analysisRequest)
      const task = response.data
      
      analysisTasks.value.unshift(task)
      currentAnalysisTask.value = task
      
      // 初始化进度追踪
      analysisProgress.value = {
        taskId: task.id,
        status: 'running',
        progress: 0,
        message: '分析任务已启动...',
        startTime: new Date(),
        estimatedTime: null
      }
      
      // 开始轮询进度
      await pollAnalysisProgress(task.id)
      
      return task
    } catch (error) {
      console.error('启动需求分析失败:', error)
      errors.value.analysis = error.message
      ElMessage.error('启动需求分析失败')
      throw error
    } finally {
      loading.value.analysis = false
    }
  }
  
  const pollAnalysisProgress = async (taskId) => {
    const interval = setInterval(async () => {
      try {
        const response = await aiApi.getAnalysisProgress(taskId)
        const progress = response.data
        
        analysisProgress.value = {
          ...analysisProgress.value,
          progress: progress.progress || 0,
          message: progress.message || '分析进行中...',
          estimatedTime: progress.estimated_time
        }
        
        if (progress.status === 'completed') {
          clearInterval(interval)
          analysisProgress.value.status = 'completed'
          await loadAnalysisResult(taskId)
          ElMessage.success('需求分析完成！')
        } else if (progress.status === 'failed') {
          clearInterval(interval)
          analysisProgress.value.status = 'failed'
          analysisProgress.value.message = progress.error || '分析失败'
          ElMessage.error('需求分析失败')
        }
        
      } catch (error) {
        console.error('获取分析进度失败:', error)
        clearInterval(interval)
        analysisProgress.value.status = 'failed'
        analysisProgress.value.message = '进度查询失败'
      }
    }, 2000) // 2秒轮询一次
  }
  
  const loadAnalysisResult = async (taskId) => {
    try {
      const response = await aiApi.getAnalysisResult(taskId)
      analysisResults.value = response.data || {}
    } catch (error) {
      console.error('加载分析结果失败:', error)
      throw error
    }
  }
  
  const cancelAnalysisTask = async (taskId) => {
    try {
      await aiApi.cancelAnalysisTask(taskId)
      
      if (analysisProgress.value.taskId === taskId) {
        analysisProgress.value.status = 'cancelled'
      }
      
      const taskIndex = analysisTasks.value.findIndex(t => t.id === taskId)
      if (taskIndex > -1) {
        analysisTasks.value[taskIndex].status = 'cancelled'
      }
      
      ElMessage.info('分析任务已取消')
    } catch (error) {
      console.error('取消分析任务失败:', error)
      ElMessage.error('取消分析任务失败')
      throw error
    }
  }
  
  // ==================== AI推荐管理 ====================
  
  const loadRecommendations = async (context = {}) => {
    loading.value.recommendations = true
    try {
      const requestData = {
        project_id: selectedProject.value?.id,
        cycle_id: selectedCycle.value?.id,
        analysis_results: analysisResults.value,
        ...context
      }
      
      const response = await recommendationApi.getAIRecommendations(requestData)
      recommendations.value = response.data || []
    } catch (error) {
      console.error('加载AI推荐失败:', error)
      ElMessage.error('加载AI推荐失败')
    } finally {
      loading.value.recommendations = false
    }
  }
  
  const adoptRecommendation = async (recommendation) => {
    try {
      await recommendationApi.adoptRecommendation({
        recommendation_id: recommendation.id,
        project_id: selectedProject.value?.id
      })
      
      adoptedRecommendations.value.push(recommendation.id)
      ElMessage.success('推荐已采纳')
      
      // 根据推荐类型应用相应的更改
      if (recommendation.type === 'requirement_optimization') {
        // 应用需求优化
        await applyRequirementOptimization(recommendation)
      }
      
    } catch (error) {
      console.error('采纳推荐失败:', error)
      ElMessage.error('采纳推荐失败')
      throw error
    }
  }
  
  const dismissRecommendation = async (recommendation) => {
    try {
      await recommendationApi.rejectRecommendation({
        recommendation_id: recommendation.id,
        reason: 'user_dismissed'
      })
      
      dismissedRecommendations.value.push(recommendation.id)
      ElMessage.info('推荐已忽略')
    } catch (error) {
      console.error('忽略推荐失败:', error)
      ElMessage.error('忽略推荐失败')
      throw error
    }
  }
  
  const applyRequirementOptimization = async (recommendation) => {
    // 应用推荐的需求优化到分析结果
    if (recommendation.target_requirements) {
      recommendation.target_requirements.forEach(reqId => {
        if (analysisResults.value[reqId]) {
          analysisResults.value[reqId] = {
            ...analysisResults.value[reqId],
            ...recommendation.optimizations
          }
        }
      })
    }
  }
  
  // ==================== 项目上下文管理 ====================
  
  const setProjectContext = (project, cycle = null, version = null) => {
    selectedProject.value = project
    selectedCycle.value = cycle
    selectedVersion.value = version
    
    // 清除相关状态
    if (project?.id !== selectedProject.value?.id) {
      selectedRequirements.value = []
      analysisResults.value = {}
      recommendations.value = []
    }
  }
  
  const updateSelectedRequirements = (requirements) => {
    selectedRequirements.value = requirements
  }
  
  // ==================== 数据导出 ====================
  
  const exportAnalysisReport = async (format = 'excel') => {
    loading.value.export = true
    try {
      const response = await aiApi.exportAnalysisReport({
        project_id: selectedProject.value?.id,
        cycle_id: selectedCycle.value?.id,
        results: analysisResults.value,
        format
      })
      
      // 处理文件下载
      const blob = new Blob([response.data])
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `需求分析报告_${new Date().toISOString().slice(0, 10)}.${format}`
      a.click()
      window.URL.revokeObjectURL(url)
      
      ElMessage.success('报告导出成功')
    } catch (error) {
      console.error('导出分析报告失败:', error)
      ElMessage.error('导出分析报告失败')
      throw error
    } finally {
      loading.value.export = false
    }
  }
  
  const exportChatSession = async (sessionId, format = 'json') => {
    try {
      const response = await chatApi.exportChatHistory(sessionId, { format })
      
      const blob = new Blob([response.data])
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `聊天记录_${sessionId}_${new Date().toISOString().slice(0, 10)}.${format}`
      a.click()
      window.URL.revokeObjectURL(url)
      
      ElMessage.success('聊天记录导出成功')
    } catch (error) {
      console.error('导出聊天记录失败:', error)
      ElMessage.error('导出聊天记录失败')
      throw error
    }
  }
  
  // ==================== 错误处理 ====================
  
  const clearError = (type) => {
    if (type) {
      errors.value[type] = null
    } else {
      Object.keys(errors.value).forEach(key => {
        errors.value[key] = null
      })
    }
  }
  
  const clearAllData = () => {
    // 重置所有状态
    chatSessions.value = []
    currentSession.value = null
    messages.value = []
    analysisTasks.value = []
    analysisResults.value = {}
    recommendations.value = []
    selectedRequirements.value = []
    
    // 关闭连接
    if (wsConnection.value) {
      wsConnection.value.close()
    }
    
    // 清除错误
    clearError()
  }
  
  // ==================== 返回状态和方法 ====================
  
  return {
    // 状态
    serviceStatus,
    chatSessions,
    currentSession,
    messages,
    isTyping,
    wsConnected,
    analysisTasks,
    currentAnalysisTask,
    analysisProgress,
    analysisResults,
    selectedRequirements,
    analysisConfig,
    recommendations,
    adoptedRecommendations,
    dismissedRecommendations,
    selectedProject,
    selectedCycle,
    selectedVersion,
    loading,
    errors,
    
    // 计算属性
    isAIServiceHealthy,
    hasActiveSession,
    isAnalysisRunning,
    pendingRecommendations,
    analysisStatistics,
    
    // AI服务管理
    checkAIServiceStatus,
    switchAIModel,
    updateAIConfig,
    
    // 聊天管理
    loadChatSessions,
    createChatSession,
    selectChatSession,
    sendChatMessage,
    deleteChatSession,
    
    // 分析管理
    startRequirementAnalysis,
    cancelAnalysisTask,
    loadAnalysisResult,
    
    // 推荐管理
    loadRecommendations,
    adoptRecommendation,
    dismissRecommendation,
    
    // 上下文管理
    setProjectContext,
    updateSelectedRequirements,
    
    // 导出功能
    exportAnalysisReport,
    exportChatSession,
    
    // 工具方法
    clearError,
    clearAllData
  }
})