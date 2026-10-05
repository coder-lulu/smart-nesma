import { ElMessage } from 'element-plus'

class ChatWebSocket {
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
  }

  // 连接WebSocket
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
      // 如果是本地开发，默认使用8888端口
      host = host.split(':')[0] + ':8888'
    }
    
    this.url = `${protocol}//${host}/chat/ws?token=${token}`
    console.log('WebSocket URL:', this.url)

    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.onopen = () => {
          console.log('WebSocket连接已建立')
          this.isConnecting = false
          this.reconnectAttempts = 0
          this.startPing()
          this.emit('connected')
          resolve()
        }

        this.ws.onmessage = (event) => {
          try {
            const message = JSON.parse(event.data)
            this.handleMessage(message)
          } catch (error) {
            console.error('消息解析失败:', error)
          }
        }

        this.ws.onclose = (event) => {
          console.log('WebSocket连接已关闭:', event.code, event.reason)
          this.isConnecting = false
          this.stopPing()
          this.emit('disconnected', { code: event.code, reason: event.reason })

          // 根据关闭代码显示具体错误信息
          if (event.code === 1006) {
            console.error('WebSocket连接异常关闭，可能是网络问题或服务器不可达')
          } else if (event.code === 1002) {
            console.error('WebSocket协议错误')
          } else if (event.code === 1011) {
            console.error('WebSocket服务器错误')
          }

          // 非手动关闭时尝试重连
          if (!this.isManualClose && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.reconnect(token)
          }
        }

        this.ws.onerror = (error) => {
          console.error('WebSocket错误:', error)
          console.error('WebSocket URL:', this.url)
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

  // 重连
  reconnect(token) {
    if (this.isManualClose) return

    this.reconnectAttempts++
    console.log(`尝试重连... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`)

    setTimeout(() => {
      this.connect(token).catch(error => {
        console.error('重连失败:', error)
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
          ElMessage.error('连接失败，请刷新页面重试')
        }
      })
    }, this.reconnectInterval)
  }

  // 断开连接
  disconnect() {
    this.isManualClose = true
    this.stopPing()
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  // 清理所有监听器
  removeAllListeners() {
    this.listeners.clear()
  }

  // 发送消息
  sendMessage(sessionId, message, modelName = 'deepseek-chat', projectId = null) {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      throw new Error('WebSocket未连接')
    }

    const payload = {
      action: 'send_message',
      sessionId: sessionId,
      projectId: projectId,
      message: message,
      modelName: modelName,
      context: {
        timestamp: new Date().toISOString()
      }
    }

    this.ws.send(JSON.stringify(payload))
  }

  // 发送ping
  ping() {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify({ action: 'ping' }))
    }
  }

  // 开始心跳
  startPing() {
    this.pingInterval = setInterval(() => {
      this.ping()
    }, 30000) // 30秒心跳
  }

  // 停止心跳
  stopPing() {
    if (this.pingInterval) {
      clearInterval(this.pingInterval)
      this.pingInterval = null
    }
  }

  // 处理接收到的消息
  handleMessage(message) {
    console.log('收到WebSocket消息:', message)

    switch (message.type) {
      case 'status':
        this.emit('status', message)
        break
      case 'message':
        this.emit('message', message)
        break
      case 'complete':
        this.emit('complete', message)
        break
      case 'error':
        this.emit('error', message)
        ElMessage.error(message.error || '发生未知错误')
        break
      case 'pong':
        // 心跳响应，无需处理
        break
      default:
        console.warn('未知消息类型:', message.type)
    }
  }

  // 事件监听
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, [])
    }
    this.listeners.get(event).push(callback)
  }

  // 移除事件监听
  off(event, callback) {
    if (this.listeners.has(event)) {
      const callbacks = this.listeners.get(event)
      const index = callbacks.indexOf(callback)
      if (index > -1) {
        callbacks.splice(index, 1)
      }
    }
  }

  // 触发事件
  emit(event, data) {
    if (this.listeners.has(event)) {
      this.listeners.get(event).forEach(callback => {
        try {
          callback(data)
        } catch (error) {
          console.error('事件回调执行失败:', error)
        }
      })
    }
  }

  // 获取连接状态
  getReadyState() {
    return this.ws ? this.ws.readyState : WebSocket.CLOSED
  }

  // 是否已连接
  isConnected() {
    return this.ws && this.ws.readyState === WebSocket.OPEN
  }
}

// 全局WebSocket实例
export const chatWebSocket = new ChatWebSocket()

export default ChatWebSocket 
