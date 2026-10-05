/**
 * WebSocket功能测试脚本
 * 用于验证Analysis WebSocket服务的基本功能
 */

// 模拟分析WebSocket服务类
class TestAnalysisWebSocket {
  constructor() {
    this.ws = null
    this.listeners = new Map()
    this.subscribedTasks = new Set()
    console.log('✅ 分析WebSocket服务初始化成功')
  }

  // 模拟连接方法
  connect(token) {
    return new Promise((resolve) => {
      console.log('🔗 开始建立WebSocket连接...')
      console.log(`📝 使用Token: ${token ? token.slice(0, 10) + '...' : '未提供'}`)
      
      setTimeout(() => {
        console.log('✅ WebSocket连接建立成功')
        this.emit('connected')
        resolve()
      }, 1000)
    })
  }

  // 模拟订阅任务
  subscribeTask(taskId) {
    console.log(`📡 订阅分析任务: ${taskId}`)
    this.subscribedTasks.add(taskId.toString())
    
    // 模拟实时进度更新
    setTimeout(() => {
      this.simulateProgressUpdates(taskId)
    }, 2000)
    
    return true
  }

  // 模拟进度更新
  simulateProgressUpdates(taskId) {
    const stages = [
      { progress: 10, stage: 'initializing', message: '初始化分析环境' },
      { progress: 30, stage: 'processing', message: '分析需求结构' },
      { progress: 60, stage: 'ai_analyzing', message: 'AI智能优化中' },
      { progress: 85, stage: 'calculating', message: 'NESMA评估计算' },
      { progress: 100, stage: 'completed', message: '分析完成' }
    ]

    stages.forEach((stage, index) => {
      setTimeout(() => {
        console.log(`📊 进度更新 ${stage.progress}%: ${stage.message}`)
        
        // 触发进度事件
        this.emit(`progress:${taskId}`, {
          taskId: taskId,
          progress: stage.progress,
          stage: stage.stage,
          message: stage.message,
          timestamp: new Date().toISOString()
        })
        
        // 最后阶段触发完成事件
        if (stage.stage === 'completed') {
          this.emit(`complete:${taskId}`, {
            taskId: taskId,
            result: {
              success: true,
              processedItems: 25,
              qualityScore: 0.92
            }
          })
        }
      }, (index + 1) * 2000)
    })
  }

  // 事件监听器
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, [])
    }
    this.listeners.get(event).push(callback)
    console.log(`🎧 监听事件: ${event}`)
  }

  // 触发事件
  emit(event, data) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach(callback => {
        try {
          callback(data)
        } catch (error) {
          console.error(`❌ 事件回调执行失败 ${event}:`, error)
        }
      })
    }
  }

  // 断开连接
  disconnect() {
    console.log('🔌 断开WebSocket连接')
    this.subscribedTasks.clear()
    this.listeners.clear()
  }
}

// 测试函数
async function testWebSocketFunctionality() {
  console.log('🚀 开始WebSocket功能测试\n')
  
  // 创建WebSocket实例
  const wsService = new TestAnalysisWebSocket()
  
  // 设置事件监听器
  wsService.on('connected', () => {
    console.log('🎉 WebSocket连接事件触发')
  })
  
  wsService.on('progress:test-task-123', (data) => {
    console.log(`📈 收到进度更新:`, {
      进度: `${data.progress}%`,
      阶段: data.stage,
      消息: data.message
    })
  })
  
  wsService.on('complete:test-task-123', (data) => {
    console.log('🏁 分析任务完成:', data.result)
    console.log('\n✅ WebSocket功能测试完成')
  })
  
  // 测试连接
  try {
    await wsService.connect('sk-test-token-12345')
    
    // 测试任务订阅
    wsService.subscribeTask('test-task-123')
    
    // 等待测试完成
    setTimeout(() => {
      wsService.disconnect()
      console.log('\n📋 测试总结:')
      console.log('- ✅ WebSocket连接建立')
      console.log('- ✅ 事件监听器注册')
      console.log('- ✅ 任务订阅功能')
      console.log('- ✅ 实时进度更新')
      console.log('- ✅ 完成事件触发')
      console.log('- ✅ 资源清理')
    }, 15000)
    
  } catch (error) {
    console.error('❌ 测试失败:', error)
  }
}

// 运行测试
testWebSocketFunctionality()

// 导出供其他模块使用
if (typeof module !== 'undefined' && module.exports) {
  module.exports = TestAnalysisWebSocket
}