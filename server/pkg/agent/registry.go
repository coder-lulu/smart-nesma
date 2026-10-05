package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// AgentRegistry Agent注册中心
type AgentRegistry struct {
	mu              sync.RWMutex
	agents          map[string]*RegisteredAgent // 注册的Agent
	typeIndex       map[AgentType][]string      // 按类型索引
	capabilityIndex map[string][]string         // 按能力索引
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup

	// 配置
	healthCheckInterval time.Duration
	agentTimeout        time.Duration
}

// RegisteredAgent 注册的Agent信息
type RegisteredAgent struct {
	Agent           Agent
	Info            AgentInfo
	RegisteredAt    time.Time
	LastHealthCheck time.Time
	HealthStatus    HealthStatus
	IsActive        bool
	mu              sync.RWMutex
}

// RegistryConfig 注册中心配置
type RegistryConfig struct {
	HealthCheckInterval time.Duration `json:"healthCheckInterval"`
	AgentTimeout        time.Duration `json:"agentTimeout"`
}

// NewAgentRegistry 创建Agent注册中心
func NewAgentRegistry(config RegistryConfig, logger *zap.Logger) *AgentRegistry {
	ctx, cancel := context.WithCancel(context.Background())

	return &AgentRegistry{
		agents:              make(map[string]*RegisteredAgent),
		typeIndex:           make(map[AgentType][]string),
		capabilityIndex:     make(map[string][]string),
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
		healthCheckInterval: config.HealthCheckInterval,
		agentTimeout:        config.AgentTimeout,
	}
}

// Start 启动注册中心
func (r *AgentRegistry) Start() error {
	r.logger.Info("Starting agent registry")

	// 启动健康检查
	r.wg.Add(1)
	go r.healthCheckWorker()

	// 启动清理工作器
	r.wg.Add(1)
	go r.cleanupWorker()

	return nil
}

// Stop 停止注册中心
func (r *AgentRegistry) Stop() error {
	r.logger.Info("Stopping agent registry")

	r.cancel()
	r.wg.Wait()

	return nil
}

// Register 注册Agent
func (r *AgentRegistry) Register(agent Agent) error {
	info := agent.GetInfo()

	r.mu.Lock()
	defer r.mu.Unlock()

	// 检查是否已注册
	if _, exists := r.agents[info.ID]; exists {
		return fmt.Errorf("agent %s already registered", info.ID)
	}

	// 创建注册信息
	registered := &RegisteredAgent{
		Agent:           agent,
		Info:            info,
		RegisteredAt:    time.Now(),
		LastHealthCheck: time.Now(),
		HealthStatus:    agent.Health(),
		IsActive:        true,
	}

	// 注册Agent
	r.agents[info.ID] = registered

	// 更新索引
	r.updateTypeIndex(info.Type, info.ID, true)
	r.updateCapabilityIndex(info.Capabilities, info.ID, true)

	r.logger.Info("Agent registered",
		zap.String("id", info.ID),
		zap.String("type", string(info.Type)),
		zap.String("name", info.Name))

	return nil
}

// Unregister 注销Agent
func (r *AgentRegistry) Unregister(agentID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	registered, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	// 从索引中移除
	r.updateTypeIndex(registered.Info.Type, agentID, false)
	r.updateCapabilityIndex(registered.Info.Capabilities, agentID, false)

	// 删除Agent
	delete(r.agents, agentID)

	r.logger.Info("Agent unregistered", zap.String("id", agentID))

	return nil
}

// GetAgent 获取Agent
func (r *AgentRegistry) GetAgent(agentID string) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registered, exists := r.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent %s not found", agentID)
	}

	if !registered.IsActive {
		return nil, fmt.Errorf("agent %s is not active", agentID)
	}

	return registered.Agent, nil
}

// GetAgentInfo 获取Agent信息
func (r *AgentRegistry) GetAgentInfo(agentID string) (AgentInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	registered, exists := r.agents[agentID]
	if !exists {
		return AgentInfo{}, fmt.Errorf("agent %s not found", agentID)
	}

	return registered.Info, nil
}

// ListAgents 列出所有Agent
func (r *AgentRegistry) ListAgents() []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agents := make([]AgentInfo, 0, len(r.agents))
	for _, registered := range r.agents {
		agents = append(agents, registered.Info)
	}

	return agents
}

// FindAgentsByType 按类型查找Agent
func (r *AgentRegistry) FindAgentsByType(agentType AgentType) []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.typeIndex[agentType]
	if !exists {
		return []AgentInfo{}
	}

	agents := make([]AgentInfo, 0, len(agentIDs))
	for _, id := range agentIDs {
		if registered, exists := r.agents[id]; exists && registered.IsActive {
			agents = append(agents, registered.Info)
		}
	}

	return agents
}

// FindAgentsByCapability 按能力查找Agent
func (r *AgentRegistry) FindAgentsByCapability(capability string) []AgentInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.capabilityIndex[capability]
	if !exists {
		return []AgentInfo{}
	}

	agents := make([]AgentInfo, 0, len(agentIDs))
	for _, id := range agentIDs {
		if registered, exists := r.agents[id]; exists && registered.IsActive {
			agents = append(agents, registered.Info)
		}
	}

	return agents
}

// GetHealthyAgent 获取健康的Agent
func (r *AgentRegistry) GetHealthyAgent(agentType AgentType) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.typeIndex[agentType]
	if !exists {
		return nil, fmt.Errorf("no agents of type %s found", agentType)
	}

	// 查找健康的Agent
	for _, id := range agentIDs {
		if registered, exists := r.agents[id]; exists && registered.IsActive {
			if registered.HealthStatus.Status == "healthy" {
				return registered.Agent, nil
			}
		}
	}

	return nil, fmt.Errorf("no healthy agents of type %s found", agentType)
}

// GetLeastLoadedAgent 获取负载最低的Agent
func (r *AgentRegistry) GetLeastLoadedAgent(agentType AgentType) (Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	agentIDs, exists := r.typeIndex[agentType]
	if !exists {
		return nil, fmt.Errorf("no agents of type %s found", agentType)
	}

	var bestAgent Agent
	var bestResponseTime int64 = -1

	// 选择响应时间最短的Agent
	for _, id := range agentIDs {
		if registered, exists := r.agents[id]; exists && registered.IsActive {
			if registered.HealthStatus.Status == "healthy" {
				responseTime := registered.HealthStatus.ResponseTime
				if bestResponseTime == -1 || responseTime < bestResponseTime {
					bestResponseTime = responseTime
					bestAgent = registered.Agent
				}
			}
		}
	}

	if bestAgent == nil {
		return nil, fmt.Errorf("no healthy agents of type %s found", agentType)
	}

	return bestAgent, nil
}

// UpdateAgentHealth 更新Agent健康状态
func (r *AgentRegistry) UpdateAgentHealth(agentID string, health HealthStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	registered, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	registered.mu.Lock()
	registered.HealthStatus = health
	registered.LastHealthCheck = time.Now()
	registered.mu.Unlock()

	return nil
}

// SetAgentActive 设置Agent活跃状态
func (r *AgentRegistry) SetAgentActive(agentID string, active bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	registered, exists := r.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	registered.mu.Lock()
	registered.IsActive = active
	registered.mu.Unlock()

	r.logger.Info("Agent activity updated",
		zap.String("id", agentID),
		zap.Bool("active", active))

	return nil
}

// GetRegistryStats 获取注册中心统计信息
func (r *AgentRegistry) GetRegistryStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})
	stats["totalAgents"] = len(r.agents)

	// 按类型统计
	typeStats := make(map[string]int)
	for agentType, agentIDs := range r.typeIndex {
		typeStats[string(agentType)] = len(agentIDs)
	}
	stats["agentsByType"] = typeStats

	// 按状态统计
	statusStats := make(map[string]int)
	for _, registered := range r.agents {
		status := registered.HealthStatus.Status
		statusStats[status]++
	}
	stats["agentsByStatus"] = statusStats

	// 活跃Agent数量
	activeCount := 0
	for _, registered := range r.agents {
		if registered.IsActive {
			activeCount++
		}
	}
	stats["activeAgents"] = activeCount

	return stats
}

// 更新类型索引
func (r *AgentRegistry) updateTypeIndex(agentType AgentType, agentID string, add bool) {
	if add {
		r.typeIndex[agentType] = append(r.typeIndex[agentType], agentID)
	} else {
		if agentIDs, exists := r.typeIndex[agentType]; exists {
			for i, id := range agentIDs {
				if id == agentID {
					r.typeIndex[agentType] = append(agentIDs[:i], agentIDs[i+1:]...)
					break
				}
			}
		}
	}
}

// 更新能力索引
func (r *AgentRegistry) updateCapabilityIndex(capabilities []string, agentID string, add bool) {
	for _, capability := range capabilities {
		if add {
			r.capabilityIndex[capability] = append(r.capabilityIndex[capability], agentID)
		} else {
			if agentIDs, exists := r.capabilityIndex[capability]; exists {
				for i, id := range agentIDs {
					if id == agentID {
						r.capabilityIndex[capability] = append(agentIDs[:i], agentIDs[i+1:]...)
						break
					}
				}
			}
		}
	}
}

// 健康检查工作器
func (r *AgentRegistry) healthCheckWorker() {
	defer r.wg.Done()

	ticker := time.NewTicker(r.healthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.performHealthCheck()
		}
	}
}

// 执行健康检查
func (r *AgentRegistry) performHealthCheck() {
	r.mu.RLock()
	agents := make(map[string]*RegisteredAgent)
	for id, registered := range r.agents {
		agents[id] = registered
	}
	r.mu.RUnlock()

	for id, registered := range agents {
		if !registered.IsActive {
			continue
		}

		// 检查Agent健康状态
		health := registered.Agent.Health()

		// 更新健康状态
		registered.mu.Lock()
		registered.HealthStatus = health
		registered.LastHealthCheck = time.Now()
		registered.mu.Unlock()

		// 记录不健康的Agent
		if health.Status != "healthy" {
			r.logger.Warn("Agent health check failed",
				zap.String("id", id),
				zap.String("status", health.Status),
				zap.String("details", health.Details))
		}
	}
}

// 清理工作器
func (r *AgentRegistry) cleanupWorker() {
	defer r.wg.Done()

	ticker := time.NewTicker(time.Minute * 5)
	defer ticker.Stop()

	for {
		select {
		case <-r.ctx.Done():
			return
		case <-ticker.C:
			r.cleanup()
		}
	}
}

// 清理超时的Agent
func (r *AgentRegistry) cleanup() {
	now := time.Now()
	timeoutAgents := make([]string, 0)

	r.mu.RLock()
	for id, registered := range r.agents {
		if registered.IsActive && now.Sub(registered.LastHealthCheck) > r.agentTimeout {
			timeoutAgents = append(timeoutAgents, id)
		}
	}
	r.mu.RUnlock()

	// 标记超时的Agent为不活跃
	for _, id := range timeoutAgents {
		r.SetAgentActive(id, false)
		r.logger.Warn("Agent marked as inactive due to timeout", zap.String("id", id))
	}
}
