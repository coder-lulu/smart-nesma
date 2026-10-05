package nesma

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"go.uber.org/zap"
)

// WebSocketManager WebSocket管理器
type WebSocketManager struct {
	clients    map[string]*WebSocketClient
	broadcast  chan []byte
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mutex      sync.RWMutex
}

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	ID       string
	UserID   uint
	Send     chan []byte
	Manager  *WebSocketManager
}

// AnalysisProgressMessage 分析进度消息
type AnalysisProgressMessage struct {
	Type        string      `json:"type"`
	AnalysisID  uint        `json:"analysisId"`
	ProjectID   uint        `json:"projectId"`
	Progress    float64     `json:"progress"`
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	CurrentStep string      `json:"currentStep"`
	Data        interface{} `json:"data,omitempty"`
	Timestamp   time.Time   `json:"timestamp"`
}

// WebSocketService WebSocket服务
type WebSocketService struct {
	manager *WebSocketManager
}

var (
	// GlobalWebSocketService 全局WebSocket服务实例
	GlobalWebSocketService *WebSocketService
	wsOnce                 sync.Once
)

// GetWebSocketService 获取WebSocket服务实例
func GetWebSocketService() *WebSocketService {
	wsOnce.Do(func() {
		manager := &WebSocketManager{
			clients:    make(map[string]*WebSocketClient),
			broadcast:  make(chan []byte),
			register:   make(chan *WebSocketClient),
			unregister: make(chan *WebSocketClient),
		}
		GlobalWebSocketService = &WebSocketService{
			manager: manager,
		}
		go manager.run()
	})
	return GlobalWebSocketService
}

// run 运行WebSocket管理器
func (m *WebSocketManager) run() {
	for {
		select {
		case client := <-m.register:
			m.mutex.Lock()
			m.clients[client.ID] = client
			m.mutex.Unlock()
			global.GVA_LOG.Info("WebSocket客户端连接", zap.String("clientId", client.ID))

		case client := <-m.unregister:
			m.mutex.Lock()
			if _, ok := m.clients[client.ID]; ok {
				delete(m.clients, client.ID)
				close(client.Send)
			}
			m.mutex.Unlock()
			global.GVA_LOG.Info("WebSocket客户端断开", zap.String("clientId", client.ID))

		case message := <-m.broadcast:
			m.mutex.RLock()
			for clientID, client := range m.clients {
				select {
				case client.Send <- message:
				default:
					delete(m.clients, clientID)
					close(client.Send)
				}
			}
			m.mutex.RUnlock()
		}
	}
}

// RegisterClient 注册客户端
func (s *WebSocketService) RegisterClient(clientID string, userID uint) *WebSocketClient {
	client := &WebSocketClient{
		ID:      clientID,
		UserID:  userID,
		Send:    make(chan []byte, 256),
		Manager: s.manager,
	}
	s.manager.register <- client
	return client
}

// UnregisterClient 注销客户端
func (s *WebSocketService) UnregisterClient(client *WebSocketClient) {
	s.manager.unregister <- client
}

// BroadcastAnalysisProgress 广播分析进度
func (s *WebSocketService) BroadcastAnalysisProgress(analysisID, projectID uint, progress float64, status, message, currentStep string) {
	progressMsg := AnalysisProgressMessage{
		Type:        "analysis_progress",
		AnalysisID:  analysisID,
		ProjectID:   projectID,
		Progress:    progress,
		Status:      status,
		Message:     message,
		CurrentStep: currentStep,
		Timestamp:   time.Now(),
	}

	data, err := json.Marshal(progressMsg)
	if err != nil {
		global.GVA_LOG.Error("序列化进度消息失败", zap.Error(err))
		return
	}

	s.manager.broadcast <- data
}

// SendToUser 发送消息给特定用户
func (s *WebSocketService) SendToUser(userID uint, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		global.GVA_LOG.Error("序列化用户消息失败", zap.Error(err))
		return
	}

	s.manager.mutex.RLock()
	defer s.manager.mutex.RUnlock()

	for _, client := range s.manager.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
				global.GVA_LOG.Debug("消息发送成功", zap.Uint("userId", userID))
			default:
				global.GVA_LOG.Warn("消息发送失败，客户端缓冲区满", zap.Uint("userId", userID))
			}
		}
	}
}

// BroadcastToAll 广播消息给所有客户端
func (s *WebSocketService) BroadcastToAll(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		global.GVA_LOG.Error("序列化广播消息失败", zap.Error(err))
		return
	}

	s.manager.broadcast <- data
}

// GetClientCount 获取当前连接的客户端数量
func (s *WebSocketService) GetClientCount() int {
	s.manager.mutex.RLock()
	defer s.manager.mutex.RUnlock()
	return len(s.manager.clients)
}

// GetConnectedUsers 获取当前连接的用户列表
func (s *WebSocketService) GetConnectedUsers() []uint {
	s.manager.mutex.RLock()
	defer s.manager.mutex.RUnlock()

	userMap := make(map[uint]bool)
	for _, client := range s.manager.clients {
		userMap[client.UserID] = true
	}

	users := make([]uint, 0, len(userMap))
	for userID := range userMap {
		users = append(users, userID)
	}
	return users
}

// AIAnalysisWebSocketIntegration AI分析WebSocket集成
type AIAnalysisWebSocketIntegration struct {
	wsService *WebSocketService
}

// NewAIAnalysisWebSocketIntegration 创建AI分析WebSocket集成
func NewAIAnalysisWebSocketIntegration() *AIAnalysisWebSocketIntegration {
	return &AIAnalysisWebSocketIntegration{
		wsService: GetWebSocketService(),
	}
}

// NotifyAnalysisStart 通知分析开始
func (integration *AIAnalysisWebSocketIntegration) NotifyAnalysisStart(analysisID, projectID, userID uint) {
	integration.wsService.SendToUser(userID, AnalysisProgressMessage{
		Type:        "analysis_start",
		AnalysisID:  analysisID,
		ProjectID:   projectID,
		Progress:    0.0,
		Status:      "started",
		Message:     "AI分析已启动",
		CurrentStep: "初始化分析引擎",
		Timestamp:   time.Now(),
	})
}

// NotifyAnalysisProgress 通知分析进度
func (integration *AIAnalysisWebSocketIntegration) NotifyAnalysisProgress(analysisID, projectID, userID uint, progress float64, currentStep, message string) {
	integration.wsService.SendToUser(userID, AnalysisProgressMessage{
		Type:        "analysis_progress",
		AnalysisID:  analysisID,
		ProjectID:   projectID,
		Progress:    progress,
		Status:      "processing",
		Message:     message,
		CurrentStep: currentStep,
		Timestamp:   time.Now(),
	})
}

// NotifyAnalysisComplete 通知分析完成
func (integration *AIAnalysisWebSocketIntegration) NotifyAnalysisComplete(analysisID, projectID, userID uint, result interface{}) {
	integration.wsService.SendToUser(userID, AnalysisProgressMessage{
		Type:        "analysis_complete",
		AnalysisID:  analysisID,
		ProjectID:   projectID,
		Progress:    100.0,
		Status:      "completed",
		Message:     "AI分析已完成",
		CurrentStep: "分析完成",
		Data:        result,
		Timestamp:   time.Now(),
	})
}

// NotifyAnalysisError 通知分析错误
func (integration *AIAnalysisWebSocketIntegration) NotifyAnalysisError(analysisID, projectID, userID uint, errorMsg string) {
	integration.wsService.SendToUser(userID, AnalysisProgressMessage{
		Type:        "analysis_error",
		AnalysisID:  analysisID,
		ProjectID:   projectID,
		Progress:    0.0,
		Status:      "failed",
		Message:     errorMsg,
		CurrentStep: "分析失败",
		Timestamp:   time.Now(),
	})
}

// NotifyStepProgress 通知步骤进度
func (integration *AIAnalysisWebSocketIntegration) NotifyStepProgress(analysisID, projectID, userID uint, step string, stepProgress float64, overallProgress float64) {
	data := map[string]interface{}{
		"step":         step,
		"stepProgress": stepProgress,
	}

	integration.wsService.SendToUser(userID, AnalysisProgressMessage{
		Type:        "step_progress",
		AnalysisID:  analysisID,
		ProjectID:   projectID,
		Progress:    overallProgress,
		Status:      "processing",
		Message:     fmt.Sprintf("正在执行: %s", step),
		CurrentStep: step,
		Data:        data,
		Timestamp:   time.Now(),
	})
}

// 消息类型常量
const (
	MessageTypeAnalysisStart    = "analysis_start"
	MessageTypeAnalysisProgress = "analysis_progress"
	MessageTypeAnalysisComplete = "analysis_complete"
	MessageTypeAnalysisError    = "analysis_error"
	MessageTypeStepProgress     = "step_progress"
)

// ProgressStep 进度步骤
type ProgressStep struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Progress    float64 `json:"progress"`
	Status      string  `json:"status"` // pending, processing, completed, failed
}

// AnalysisProgressDetail 详细分析进度
type AnalysisProgressDetail struct {
	AnalysisID      uint           `json:"analysisId"`
	ProjectID       uint           `json:"projectId"`
	OverallProgress float64        `json:"overallProgress"`
	CurrentStep     string         `json:"currentStep"`
	Steps           []ProgressStep `json:"steps"`
	StartTime       time.Time      `json:"startTime"`
	EstimatedTime   time.Duration  `json:"estimatedTime"`
	ElapsedTime     time.Duration  `json:"elapsedTime"`
}

// UpdateDetailedProgress 更新详细进度
func (integration *AIAnalysisWebSocketIntegration) UpdateDetailedProgress(userID uint, detail AnalysisProgressDetail) {
	integration.wsService.SendToUser(userID, AnalysisProgressMessage{
		Type:        "detailed_progress",
		AnalysisID:  detail.AnalysisID,
		ProjectID:   detail.ProjectID,
		Progress:    detail.OverallProgress,
		Status:      "processing",
		Message:     fmt.Sprintf("当前步骤: %s", detail.CurrentStep),
		CurrentStep: detail.CurrentStep,
		Data:        detail,
		Timestamp:   time.Now(),
	})
}