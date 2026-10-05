package nesma

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	clients     map[string]*WebSocketClient
	broadcast   chan []byte
	register    chan *WebSocketClient
	unregister  chan *WebSocketClient
	mu          sync.RWMutex
	upgrader    websocket.Upgrader
	batchService *nesma.BatchEvaluationService
}

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	id       string
	conn     *websocket.Conn
	send     chan []byte
	taskID   string
	userID   string
	lastPing time.Time
	mu       sync.RWMutex
}

// BatchWebSocketMessage WebSocket消息
type BatchWebSocketMessage struct {
	Type      string                 `json:"type"`
	TaskID    string                 `json:"task_id,omitempty"`
	EventType string                 `json:"event_type,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	MessageID string                 `json:"message_id,omitempty"`
}

// BatchEvaluationEvent 批量评估事件
type BatchEvaluationEvent struct {
	EventType string                 `json:"event_type"`
	TaskID    string                 `json:"task_id"`
	BatchID   string                 `json:"batch_id,omitempty"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// NewWebSocketManager 创建WebSocket管理器
func NewWebSocketManager(batchService *nesma.BatchEvaluationService) *WebSocketManager {
	return &WebSocketManager{
		clients:     make(map[string]*WebSocketClient),
		broadcast:   make(chan []byte),
		register:    make(chan *WebSocketClient),
		unregister:  make(chan *WebSocketClient),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 在生产环境中应该进行更严格的检查
			},
		},
		batchService: batchService,
	}
}

// Run 运行WebSocket管理器
func (manager *WebSocketManager) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-manager.register:
			manager.mu.Lock()
			manager.clients[client.id] = client
			manager.mu.Unlock()

			global.GVA_LOG.Info("WebSocket客户端连接", 
				zap.String("clientID", client.id),
				zap.String("userID", client.userID))

			// 发送连接成功消息
			welcome := BatchWebSocketMessage{
				Type:      "connection",
				EventType: "connected",
				Data: map[string]interface{}{
					"client_id": client.id,
					"user_id":   client.userID,
				},
				Timestamp: time.Now(),
			}
			client.send <- manager.marshalMessage(welcome)

		case client := <-manager.unregister:
			manager.mu.Lock()
			if _, ok := manager.clients[client.id]; ok {
				delete(manager.clients, client.id)
				close(client.send)
			}
			manager.mu.Unlock()

			global.GVA_LOG.Info("WebSocket客户端断开连接", 
				zap.String("clientID", client.id))

		case message := <-manager.broadcast:
			manager.mu.RLock()
			for _, client := range manager.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(manager.clients, client.id)
				}
			}
			manager.mu.RUnlock()

		case <-ticker.C:
			manager.cleanupInactiveClients()
		}
	}
}

// @Tags BatchEvaluationWebSocket
// @Summary 批量评估WebSocket连接
// @Description 建立WebSocket连接以接收实时进度更新
// @Param task_id query string true "任务ID"
// @Param user_id query string true "用户ID"
// @Router /api/v1/nesma/batch-evaluation/ws [get]
func (manager *WebSocketManager) HandleWebSocket(c *gin.Context) {
	taskID := c.Query("task_id")
	userID := c.Query("user_id")

	if taskID == "" || userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "task_id and user_id are required"})
		return
	}

	// 升级连接到WebSocket
	conn, err := manager.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		global.GVA_LOG.Error("WebSocket升级失败", zap.Error(err))
		return
	}

	// 创建客户端
	client := &WebSocketClient{
		id:       fmt.Sprintf("%s_%s_%d", userID, taskID, time.Now().UnixNano()),
		conn:     conn,
		send:     make(chan []byte, 256),
		taskID:   taskID,
		userID:   userID,
		lastPing: time.Now(),
	}

	// 注册客户端
	manager.register <- client

	// 启动goroutine处理消息
	go client.writePump()
	go client.readPump(manager)
}

// readPump 处理来自WebSocket的消息
func (client *WebSocketClient) readPump(manager *WebSocketManager) {
	defer func() {
		manager.unregister <- client
		client.conn.Close()
	}()

	client.conn.SetReadLimit(512)
	client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.conn.SetPongHandler(func(string) error {
		client.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		client.lastPing = time.Now()
		return nil
	})

	for {
		_, message, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				global.GVA_LOG.Error("WebSocket读取错误", zap.Error(err))
			}
			break
		}

		var msg BatchWebSocketMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			global.GVA_LOG.Error("WebSocket消息解析失败", zap.Error(err))
			continue
		}

		// 处理客户端消息
		manager.handleClientMessage(client, &msg)
	}
}

// writePump 处理发送到WebSocket的消息
func (client *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		client.conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.send:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				client.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			client.conn.WriteMessage(websocket.TextMessage, message)

		case <-ticker.C:
			client.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleClientMessage 处理客户端消息
func (manager *WebSocketManager) handleClientMessage(client *WebSocketClient, msg *BatchWebSocketMessage) {
	switch msg.Type {
	case "ping":
		// 响应ping
		pong := BatchWebSocketMessage{
			Type:      "pong",
			Timestamp: time.Now(),
		}
		client.send <- manager.marshalMessage(pong)

	case "subscribe":
		// 订阅任务状态更新
		taskID := msg.TaskID
		if taskID == "" {
			taskID = client.taskID
		}

		if taskID != "" {
			client.mu.Lock()
			client.taskID = taskID
			client.mu.Unlock()

			// 发送当前任务状态
			manager.sendCurrentTaskStatus(client, taskID)
		}

	case "unsubscribe":
		// 取消订阅
		client.mu.Lock()
		client.taskID = ""
		client.mu.Unlock()

	case "get_status":
		// 获取任务状态
		taskID := msg.TaskID
		if taskID == "" {
			taskID = client.taskID
		}

		if taskID != "" {
			manager.sendCurrentTaskStatus(client, taskID)
		}

	default:
		global.GVA_LOG.Warn("未知的WebSocket消息类型", zap.String("type", msg.Type))
	}
}

// sendCurrentTaskStatus 发送当前任务状态
func (manager *WebSocketManager) sendCurrentTaskStatus(client *WebSocketClient, taskID string) {
	task, err := manager.batchService.GetBatchEvaluationStatus(taskID)
	if err != nil {
		errorMsg := BatchWebSocketMessage{
			Type:      "error",
			EventType: "status_error",
			Data: map[string]interface{}{
				"error": err.Error(),
			},
			Timestamp: time.Now(),
		}
		client.send <- manager.marshalMessage(errorMsg)
		return
	}

	statusMsg := BatchWebSocketMessage{
		Type:      "status",
		TaskID:    taskID,
		EventType: "status_update",
		Data: map[string]interface{}{
			"status":             task.Status,
			"progress":           task.Progress,
			"current_batch":      task.CurrentBatch,
			"total_batches":      task.TotalBatches,
			"processed_count":    task.ProcessedCount,
			"total_requirements": task.TotalRequirements,
			"start_time":         task.StartTime,
			"estimated_end_time": task.EstimatedEndTime,
			"message":            task.StatusMessage,
		},
		Timestamp: time.Now(),
	}

	client.send <- manager.marshalMessage(statusMsg)
}

// BroadcastTaskUpdate 广播任务更新
func (manager *WebSocketManager) BroadcastTaskUpdate(taskID string, event *BatchEvaluationEvent) {
	message := BatchWebSocketMessage{
		Type:      "update",
		TaskID:    taskID,
		EventType: event.EventType,
		Data:      event.Data,
		Timestamp: event.Timestamp,
	}

	messageBytes := manager.marshalMessage(message)

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	for _, client := range manager.clients {
		client.mu.RLock()
		if client.taskID == taskID {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(manager.clients, client.id)
			}
		}
		client.mu.RUnlock()
	}
}

// BroadcastToAll 广播到所有客户端
func (manager *WebSocketManager) BroadcastToAll(message *BatchWebSocketMessage) {
	messageBytes := manager.marshalMessage(*message)
	manager.broadcast <- messageBytes
}

// SendToUser 发送消息给特定用户
func (manager *WebSocketManager) SendToUser(userID string, message *BatchWebSocketMessage) {
	messageBytes := manager.marshalMessage(*message)

	manager.mu.RLock()
	defer manager.mu.RUnlock()

	for _, client := range manager.clients {
		if client.userID == userID {
			select {
			case client.send <- messageBytes:
			default:
				close(client.send)
				delete(manager.clients, client.id)
			}
		}
	}
}

// GetConnectedClients 获取连接的客户端数量
func (manager *WebSocketManager) GetConnectedClients() int {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	return len(manager.clients)
}

// GetTaskSubscribers 获取任务订阅者数量
func (manager *WebSocketManager) GetTaskSubscribers(taskID string) int {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	count := 0
	for _, client := range manager.clients {
		client.mu.RLock()
		if client.taskID == taskID {
			count++
		}
		client.mu.RUnlock()
	}
	return count
}

// cleanupInactiveClients 清理不活跃的客户端
func (manager *WebSocketManager) cleanupInactiveClients() {
	manager.mu.Lock()
	defer manager.mu.Unlock()

	cutoff := time.Now().Add(-5 * time.Minute)
	for id, client := range manager.clients {
		if client.lastPing.Before(cutoff) {
			close(client.send)
			delete(manager.clients, id)
			global.GVA_LOG.Info("清理不活跃的WebSocket客户端", zap.String("clientID", id))
		}
	}
}

// marshalMessage 序列化消息
func (manager *WebSocketManager) marshalMessage(message BatchWebSocketMessage) []byte {
	if message.MessageID == "" {
		message.MessageID = fmt.Sprintf("msg_%d", time.Now().UnixNano())
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		global.GVA_LOG.Error("WebSocket消息序列化失败", zap.Error(err))
		return []byte(`{"type":"error","data":{"error":"message serialization failed"}}`)
	}
	return bytes
}

// BatchEvaluationEventHandler 批量评估事件处理器
type BatchEvaluationEventHandler struct {
	wsManager *WebSocketManager
}

// NewBatchEvaluationEventHandler 创建事件处理器
func NewBatchEvaluationEventHandler(wsManager *WebSocketManager) *BatchEvaluationEventHandler {
	return &BatchEvaluationEventHandler{
		wsManager: wsManager,
	}
}

// HandleTaskCreated 处理任务创建事件
func (h *BatchEvaluationEventHandler) HandleTaskCreated(taskID string, data map[string]interface{}) {
	event := &BatchEvaluationEvent{
		EventType: "task_created",
		TaskID:    taskID,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleTaskStarted 处理任务启动事件
func (h *BatchEvaluationEventHandler) HandleTaskStarted(taskID string, data map[string]interface{}) {
	event := &BatchEvaluationEvent{
		EventType: "task_started",
		TaskID:    taskID,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleBatchStarted 处理批次启动事件
func (h *BatchEvaluationEventHandler) HandleBatchStarted(taskID, batchID string, data map[string]interface{}) {
	event := &BatchEvaluationEvent{
		EventType: "batch_started",
		TaskID:    taskID,
		BatchID:   batchID,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleBatchCompleted 处理批次完成事件
func (h *BatchEvaluationEventHandler) HandleBatchCompleted(taskID, batchID string, data map[string]interface{}) {
	event := &BatchEvaluationEvent{
		EventType: "batch_completed",
		TaskID:    taskID,
		BatchID:   batchID,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleProgressUpdate 处理进度更新事件
func (h *BatchEvaluationEventHandler) HandleProgressUpdate(taskID string, progress float64, message string) {
	event := &BatchEvaluationEvent{
		EventType: "progress_update",
		TaskID:    taskID,
		Data: map[string]interface{}{
			"progress": progress,
			"message":  message,
		},
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleTaskCompleted 处理任务完成事件
func (h *BatchEvaluationEventHandler) HandleTaskCompleted(taskID string, data map[string]interface{}) {
	event := &BatchEvaluationEvent{
		EventType: "task_completed",
		TaskID:    taskID,
		Data:      data,
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleTaskFailed 处理任务失败事件
func (h *BatchEvaluationEventHandler) HandleTaskFailed(taskID string, errorMsg string) {
	event := &BatchEvaluationEvent{
		EventType: "task_failed",
		TaskID:    taskID,
		Data: map[string]interface{}{
			"error": errorMsg,
		},
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// HandleTaskCancelled 处理任务取消事件
func (h *BatchEvaluationEventHandler) HandleTaskCancelled(taskID string, reason string) {
	event := &BatchEvaluationEvent{
		EventType: "task_cancelled",
		TaskID:    taskID,
		Data: map[string]interface{}{
			"reason": reason,
		},
		Timestamp: time.Now(),
	}
	h.wsManager.BroadcastTaskUpdate(taskID, event)
}

// 全局WebSocket管理器实例
var GlobalWebSocketManager *WebSocketManager

// InitWebSocketManager 初始化WebSocket管理器
func InitWebSocketManager(batchService *nesma.BatchEvaluationService) {
	GlobalWebSocketManager = NewWebSocketManager(batchService)
	go GlobalWebSocketManager.Run()
}