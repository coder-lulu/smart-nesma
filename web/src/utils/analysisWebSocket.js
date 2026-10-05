import { ElMessage } from 'element-plus'

/**
 * AI分析专用WebSocket服务
 * 用于实时接收分析进度和状态更新
 */
class AnalysisWebSocket {
  constructor() {
    this.ws = null
    this.url = ''
    this.listeners = new Map()
    this.reconnectAttempts = 0
    this.maxReconnectAttempts = 5
    this.reconnectInterval = 3000
    this.pingInterval = null
    this.isConnecting = false
    this.isManualClose = false
    this.subscribedTasks = new Set() // 订阅的任务ID
  }

  /**
   * 连接WebSocket
   * @param {string} token 认证token
   */
  connect(token) {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
      return Promise.resolve()
    }

    this.isConnecting = true
    this.isManualClose = false

    // 构建WebSocket URL
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    let host = window.location.host
    
    // 开发环境端口映射
    if (host.includes(':5173')) {
      host = host.replace(':5173', ':8888')
    } else if (host.includes('localhost') || host.includes('127.0.0.1')) {
      host = host.split(':')[0] + ':8888'
    }
    
    this.url = `${protocol}//${host}/analysis/ws?token=${token}`
    console.log('Analysis WebSocket URL:', this.url)

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('Analysis WebSocket连接已建立')
          this.isConnecting = false
          this.reconnectAttempts = 0
          this.startPing()
          this.emit('connected')
          
          // 重新订阅之前订阅的任务
          this.resubscribeAllTasks()
          
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('Analysis WebSocket消息解析失败:', error)
          }
        }

        this.ws.onclose = (event) => {
          console.log('Analysis WebSocket连接已关闭:', event.code, event.reason)
          this.isConnecting = false
          this.stopPing()
          this.emit('disconnected', { code: event.code, reason: event.reason })

          // 根据关闭代码显示具体错误信息
          if (event.code === 1006) {
            console.error('Analysis WebSocket连接异常关闭，可能是网络问题或服务器不可达')
          } else if (event.code === 1002) {
            console.error('Analysis WebSocket协议错误')
          } else if (event.code === 1011) {
            console.error('Analysis WebSocket服务器错误')
          }

          // 非手动关闭时尝试重连
          if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnect(token)
          }
        }

        this.ws.onerror = (error) => {
          console.error('Analysis WebSocket错误:', error)
          console.error('Analysis WebSocket URL:', this.url)
          this.isConnecting = false
          this.emit('error', error)
          reject(error)
        }

      } catch (error) {
        this.isConnecting = false
        reject(error)
      }
    })
  }

  /**
   * 重连
   * @param {string} token 认证token
   */
  reconnect(token) {
    if (this.isManualClose) return

    this.reconnectAttempts++
    console.log(`Analysis WebSocket尝试重连... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

    setTimeout(() => {
      this.connect(token).catch(error => {
        console.error('Analysis WebSocket重连失败:', error)
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          ElMessage.error('分析服务连接失败，请刷新页面重试')
        }
      })
    }, this.reconnectInterval * Math.pow(1.5, this.reconnectAttempts - 1)) // 指数退避
  }

  /**
   * 断开连接
   */
  disconnect() {
    this.isManualClose = true
    this.stopPing()
    this.subscribedTasks.clear()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  /**
   * 订阅分析任务进度
   * @param {string|number} taskId 任务ID
   */
  subscribeTask(taskId) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      console.warn('Analysis WebSocket未连接，无法订阅任务:', taskId)
      return false
    }

    const payload = {
      action: 'subscribe',
      taskId: taskId.toString(),
      timestamp: new Date().toISOString()
    }

    this.ws.send(JSON.stringify(payload))
    this.subscribedTasks.add(taskId.toString())
    console.log('已订阅分析任务:', taskId)
    return true
  }

  /**
   * 取消订阅分析任务进度
   * @param {string|number} taskId 任务ID
   */
  unsubscribeTask(taskId) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.subscribedTasks.delete(taskId.toString())
      return false
    }

    const payload = {
      action: 'unsubscribe',
      taskId: taskId.toString(),
      timestamp: new Date().toISOString()
    }

    this.ws.send(JSON.stringify(payload))
    this.subscribedTasks.delete(taskId.toString())
    console.log('已取消订阅分析任务:', taskId)
    return true
  }

  /**
   * 重新订阅所有任务（用于重连后）
   */
  resubscribeAllTasks() {
    if (this.subscribedTasks.size === 0) return

    console.log('重新订阅所有分析任务:', Array.from(this.subscribedTasks))
    
    this.subscribedTasks.forEach(taskId => {
      const payload = {
        action: 'subscribe',
        taskId: taskId,
        timestamp: new Date().toISOString()
      }
      this.ws.send(JSON.stringify(payload))
    })
  }

  /**
   * 发送心跳
   */
  ping() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ action: 'ping' }))
    }
  }

  /**
   * 开始心跳
   */
  startPing() {
    this.pingInterval = setInterval(() => {
      this.ping()
    }, 30000) // 30秒心跳
  }

  /**
   * 停止心跳
   */
  stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  /**
   * 处理接收到的消息
   * @param {Object} message 消息对象
   */
  handleMessage(message) {
    console.log('收到Analysis WebSocket消息:', message)

    switch (message.type) {
      case 'task_progress':
        // 任务进度更新
        this.emit('progress', message)
        this.emit(`progress:${message.taskId}`, message)
        break
        
      case 'task_status':
        // 任务状态更新
        this.emit('status', message)
        this.emit(`status:${message.taskId}`, message)
        break
        
      case 'task_complete':
        // 任务完成
        this.emit('complete', message)
        this.emit(`complete:${message.taskId}`, message)
        // 自动取消订阅已完成的任务
        this.subscribedTasks.delete(message.taskId)
        break
        
      case 'task_error':
        // 任务错误
        this.emit('error', message)
        this.emit(`error:${message.taskId}`, message)
        ElMessage.error(message.error || '分析任务发生错误')
        // 自动取消订阅出错的任务
        this.subscribedTasks.delete(message.taskId)
        break
        
      case 'task_stage':
        // 任务阶段更新
        this.emit('stage', message)
        this.emit(`stage:${message.taskId}`, message)
        break
        
      case 'system_status':
        // 系统状态更新
        this.emit('system', message)
        break
        
      case 'pong':
        // 心跳响应，无需处理
        break
        
      default:
        console.warn('未知的Analysis WebSocket消息类型:', message.type)
    }
  }

  /**
   * 事件监听
   * @param {string} event 事件名
   * @param {Function} callback 回调函数
   */
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, [])
    }
    this.listeners.get(event).push(callback)
  }

  /**
   * 移除事件监听
   * @param {string} event 事件名
   * @param {Function} callback 回调函数
   */
  off(event, callback) {
    if (this.listeners.has(event)) {
      const callbacks = this.listeners.get(event)
      const index = callbacks.indexOf(callback)
      if (index > -1) {
        callbacks.splice(index, 1)
      }
    }
  }

  /**
   * 触发事件
   * @param {string} event 事件名
   * @param {any} data 事件数据
   */
  emit(event, data) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach(callback => {
        try {
          callback(data)
        } catch (error) {
          console.error('Analysis WebSocket事件回调执行失败:', error)
        }
      })
    }
  }

  /**
   * 清理所有监听器
   */
  removeAllListeners() {
    this.listeners.clear()
  }

  /**
   * 获取连接状态
   */
  getReadyState() {
    return this.ws ? this.ws.readyState : WebSocket.CLOSED
  }

  /**
   * 是否已连接
   */
  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN
  }

  /**
   * 获取订阅的任务列表
   */
  getSubscribedTasks() {
    return Array.from(this.subscribedTasks)
  }
}

// 全局Analysis WebSocket实例
export const analysisWebSocket = new AnalysisWebSocket()

export default AnalysisWebSocket