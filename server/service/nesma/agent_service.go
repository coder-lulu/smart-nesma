package nesma

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AgentService Agent服务接口
type AgentService interface {
	RegisterAgent(ctx context.Context, agent *nesma.NesmaAgent) error
	GetAgent(ctx context.Context, agentID string) (*nesma.NesmaAgent, error)
	UpdateAgent(ctx context.Context, agent *nesma.NesmaAgent) error
	DeleteAgent(ctx context.Context, agentID string) error
	ListAgents(ctx context.Context) ([]*nesma.NesmaAgent, error)
	GetAvailableAgents(ctx context.Context, agentType string) ([]*nesma.NesmaAgent, error)
	CreateTask(ctx context.Context, task *nesma.NesmaAgentTask) error
	GetTask(ctx context.Context, taskID string) (*nesma.NesmaAgentTask, error)
	UpdateTask(ctx context.Context, task *nesma.NesmaAgentTask) error
	AssignTask(ctx context.Context, taskID string, agentID string) error
	ExecuteTask(ctx context.Context, taskID string) error
	GetTasksByAgent(ctx context.Context, agentID string) ([]*nesma.NesmaAgentTask, error)
	GetTasksByStatus(ctx context.Context, status string) ([]*nesma.NesmaAgentTask, error)
	SendMessage(ctx context.Context, message *nesma.NesmaAgentMessage) error
	GetMessages(ctx context.Context, sessionID string) ([]*nesma.NesmaAgentMessage, error)
	UpdateAgentStatus(ctx context.Context, agentID string, status string) error
	Heartbeat(ctx context.Context, agentID string) error
}

// AgentServiceImpl Agent服务实现
type AgentServiceImpl struct {
	db *gorm.DB
	mu sync.RWMutex
}

// NewAgentService 创建Agent服务实例
func NewAgentService() AgentService {
	return &AgentServiceImpl{
		db: global.GVA_DB,
	}
}

// RegisterAgent 注册Agent
func (a *AgentServiceImpl) RegisterAgent(ctx context.Context, agent *nesma.NesmaAgent) error {
	if agent.AgentID == "" {
		agent.AgentID = uuid.New().String()
	}

	// 检查Agent是否已存在
	var existing nesma.NesmaAgent
	err := a.db.WithContext(ctx).Where("agent_id = ?", agent.AgentID).First(&existing).Error
	if err == nil {
		return errors.New("agent already exists")
	}

	agent.Status = "active"
	agent.LastHeartbeat = time.Now()

	return a.db.WithContext(ctx).Create(agent).Error
}

// GetAgent 获取Agent
func (a *AgentServiceImpl) GetAgent(ctx context.Context, agentID string) (*nesma.NesmaAgent, error) {
	var agent nesma.NesmaAgent
	err := a.db.WithContext(ctx).Where("agent_id = ?", agentID).First(&agent).Error
	if err != nil {
		return nil, err
	}
	return &agent, nil
}

// UpdateAgent 更新Agent
func (a *AgentServiceImpl) UpdateAgent(ctx context.Context, agent *nesma.NesmaAgent) error {
	return a.db.WithContext(ctx).Where("agent_id = ?", agent.AgentID).Updates(agent).Error
}

// DeleteAgent 删除Agent
func (a *AgentServiceImpl) DeleteAgent(ctx context.Context, agentID string) error {
	return a.db.WithContext(ctx).Where("agent_id = ?", agentID).Delete(&nesma.NesmaAgent{}).Error
}

// ListAgents 获取所有Agent
func (a *AgentServiceImpl) ListAgents(ctx context.Context) ([]*nesma.NesmaAgent, error) {
	var agents []*nesma.NesmaAgent
	err := a.db.WithContext(ctx).Find(&agents).Error
	return agents, err
}

// GetAvailableAgents 获取可用的Agent
func (a *AgentServiceImpl) GetAvailableAgents(ctx context.Context, agentType string) ([]*nesma.NesmaAgent, error) {
	var agents []*nesma.NesmaAgent
	query := a.db.WithContext(ctx).Where("status = ?", "active")

	if agentType != "" {
		query = query.Where("agent_type = ?", agentType)
	}

	err := query.Find(&agents).Error
	return agents, err
}

// CreateTask 创建任务
func (a *AgentServiceImpl) CreateTask(ctx context.Context, task *nesma.NesmaAgentTask) error {
	if task.TaskID == "" {
		task.TaskID = uuid.New().String()
	}

	task.Status = "pending"

	return a.db.WithContext(ctx).Create(task).Error
}

// GetTask 获取任务
func (a *AgentServiceImpl) GetTask(ctx context.Context, taskID string) (*nesma.NesmaAgentTask, error) {
	var task nesma.NesmaAgentTask
	err := a.db.WithContext(ctx).Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新任务
func (a *AgentServiceImpl) UpdateTask(ctx context.Context, task *nesma.NesmaAgentTask) error {
	return a.db.WithContext(ctx).Where("task_id = ?", task.TaskID).Updates(task).Error
}

// AssignTask 分配任务
func (a *AgentServiceImpl) AssignTask(ctx context.Context, taskID string, agentID string) error {
	// 检查Agent是否可用
	agent, err := a.GetAgent(ctx, agentID)
	if err != nil {
		return err
	}

	if agent.Status != "active" {
		return errors.New("agent is not active")
	}

	// 检查Agent是否已达到最大并发数
	if agent.ProcessingCount >= agent.MaxConcurrency {
		return errors.New("agent is at maximum concurrency")
	}

	// 更新任务状态
	task, err := a.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.AgentID = agentID
	task.Status = "assigned"

	// 更新Agent处理计数
	agent.ProcessingCount++

	// 事务处理
	return a.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", task.TaskID).Updates(task).Error; err != nil {
			return err
		}
		if err := tx.Where("agent_id = ?", agent.AgentID).Updates(agent).Error; err != nil {
			return err
		}
		return nil
	})
}

// ExecuteTask 执行任务
func (a *AgentServiceImpl) ExecuteTask(ctx context.Context, taskID string) error {
	task, err := a.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	if task.Status != "assigned" {
		return errors.New("task is not assigned")
	}

	// 更新任务状态为处理中
	task.Status = "processing"
	startTime := time.Now()
	task.StartTime = &startTime

	return a.UpdateTask(ctx, task)
}

// GetTasksByAgent 获取Agent的任务
func (a *AgentServiceImpl) GetTasksByAgent(ctx context.Context, agentID string) ([]*nesma.NesmaAgentTask, error) {
	var tasks []*nesma.NesmaAgentTask
	err := a.db.WithContext(ctx).Where("agent_id = ?", agentID).Find(&tasks).Error
	return tasks, err
}

// GetTasksByStatus 根据状态获取任务
func (a *AgentServiceImpl) GetTasksByStatus(ctx context.Context, status string) ([]*nesma.NesmaAgentTask, error) {
	var tasks []*nesma.NesmaAgentTask
	err := a.db.WithContext(ctx).Where("status = ?", status).Find(&tasks).Error
	return tasks, err
}

// SendMessage 发送消息
func (a *AgentServiceImpl) SendMessage(ctx context.Context, message *nesma.NesmaAgentMessage) error {
	if message.MessageID == "" {
		message.MessageID = uuid.New().String()
	}

	return a.db.WithContext(ctx).Create(message).Error
}

// GetMessages 获取消息
func (a *AgentServiceImpl) GetMessages(ctx context.Context, sessionID string) ([]*nesma.NesmaAgentMessage, error) {
	var messages []*nesma.NesmaAgentMessage
	err := a.db.WithContext(ctx).Where("session_id = ?", sessionID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}

// UpdateAgentStatus 更新Agent状态
func (a *AgentServiceImpl) UpdateAgentStatus(ctx context.Context, agentID string, status string) error {
	return a.db.WithContext(ctx).Model(&nesma.NesmaAgent{}).Where("agent_id = ?", agentID).Updates(map[string]interface{}{
		"status":         status,
		"last_heartbeat": time.Now(),
	}).Error
}

// Heartbeat Agent心跳
func (a *AgentServiceImpl) Heartbeat(ctx context.Context, agentID string) error {
	return a.db.WithContext(ctx).Model(&nesma.NesmaAgent{}).Where("agent_id = ?", agentID).Updates(map[string]interface{}{
		"last_heartbeat": time.Now(),
	}).Error
}

// AgentManager Agent管理器
type AgentManager struct {
	service AgentService
	mu      sync.RWMutex
}

// NewAgentManager 创建Agent管理器
func NewAgentManager(service AgentService) *AgentManager {
	return &AgentManager{
		service: service,
	}
}

// ScheduleTask 调度任务
func (m *AgentManager) ScheduleTask(ctx context.Context, task *nesma.NesmaAgentTask) error {
	// 创建任务
	err := m.service.CreateTask(ctx, task)
	if err != nil {
		return err
	}

	// 查找可用的Agent
	agents, err := m.service.GetAvailableAgents(ctx, task.TaskType)
	if err != nil {
		return err
	}

	if len(agents) == 0 {
		return errors.New("no available agents")
	}

	// 简单的负载均衡：选择处理任务最少的Agent
	var bestAgent *nesma.NesmaAgent
	for _, agent := range agents {
		if bestAgent == nil || agent.ProcessingCount < bestAgent.ProcessingCount {
			bestAgent = agent
		}
	}

	// 分配任务
	return m.service.AssignTask(ctx, task.TaskID, bestAgent.AgentID)
}

// CompleteTask 完成任务
func (m *AgentManager) CompleteTask(ctx context.Context, taskID string, result interface{}) error {
	task, err := m.service.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	// 更新任务状态
	task.Status = "completed"
	endTime := time.Now()
	task.EndTime = &endTime

	if task.StartTime != nil {
		task.ProcessingTime = int(endTime.Sub(*task.StartTime).Milliseconds())
	}

	// 序列化结果
	if result != nil {
		resultData, err := json.Marshal(result)
		if err != nil {
			return err
		}
		task.OutputData = resultData
	}

	// 更新Agent处理计数
	agent, err := m.service.GetAgent(ctx, task.AgentID)
	if err != nil {
		return err
	}

	agent.ProcessingCount--
	agent.TotalProcessed++

	// 计算平均响应时间
	if agent.TotalProcessed > 0 {
		agent.AverageResponseTime = (agent.AverageResponseTime*(agent.TotalProcessed-1) + task.ProcessingTime) / agent.TotalProcessed
	}

	// 事务处理
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&nesma.NesmaAgentTask{}).Where("task_id = ?", task.TaskID).Updates(task).Error; err != nil {
			return err
		}
		if err := tx.Model(&nesma.NesmaAgent{}).Where("agent_id = ?", agent.AgentID).Updates(agent).Error; err != nil {
			return err
		}
		return nil
	})
}

// FailTask 任务失败
func (m *AgentManager) FailTask(ctx context.Context, taskID string, errorMsg string) error {
	task, err := m.service.GetTask(ctx, taskID)
	if err != nil {
		return err
	}

	task.Status = "failed"
	task.ErrorMessage = errorMsg
	endTime := time.Now()
	task.EndTime = &endTime

	if task.StartTime != nil {
		task.ProcessingTime = int(endTime.Sub(*task.StartTime).Milliseconds())
	}

	// 更新Agent处理计数
	agent, err := m.service.GetAgent(ctx, task.AgentID)
	if err != nil {
		return err
	}

	agent.ProcessingCount--

	// 事务处理
	return global.GVA_DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&nesma.NesmaAgentTask{}).Where("task_id = ?", task.TaskID).Updates(task).Error; err != nil {
			return err
		}
		if err := tx.Model(&nesma.NesmaAgent{}).Where("agent_id = ?", agent.AgentID).Updates(agent).Error; err != nil {
			return err
		}
		return nil
	})
}

// MonitorAgents 监控Agent健康状态
func (m *AgentManager) MonitorAgents(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			agents, err := m.service.ListAgents(ctx)
			if err != nil {
				global.GVA_LOG.Error(fmt.Sprintf("Failed to list agents: %v", err))
				continue
			}

			for _, agent := range agents {
				// 检查Agent心跳
				if time.Since(agent.LastHeartbeat) > 2*time.Minute {
					// Agent超时，标记为离线
					m.service.UpdateAgentStatus(ctx, agent.AgentID, "offline")
				}
			}
		}
	}
}

// 全局服务实例
var (
	agentService AgentService
	agentManager *AgentManager
)

// GetAgentService 获取Agent服务实例
func GetAgentService() AgentService {
	if agentService == nil {
		agentService = NewAgentService()
	}
	return agentService
}

// GetAgentManager 获取Agent管理器实例
func GetAgentManager() *AgentManager {
	if agentManager == nil {
		agentManager = NewAgentManager(GetAgentService())
	}
	return agentManager
}
