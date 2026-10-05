package nesma

import (
	"encoding/json"
	"runtime"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AgentApi struct{}

// RegisterAgent 注册Agent
func (a *AgentApi) RegisterAgent(c *gin.Context) {
	var req nesmaReq.RegisterAgentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	agentService := nesmaService.GetAgentService()

	agent := &nesma.NesmaAgent{
		AgentID:        req.AgentID,
		Name:           req.Name,
		Description:    req.Description,
		AgentType:      req.AgentType,
		Capabilities:   req.Capabilities,
		Configuration:  req.Configuration,
		MaxConcurrency: req.MaxConcurrency,
		Version:        req.Version,
	}

	if err := agentService.RegisterAgent(c.Request.Context(), agent); err != nil {
		global.GVA_LOG.Error("Failed to register agent", zap.Error(err))
		response.FailWithMessage("注册Agent失败: "+err.Error(), c)
		return
	}

	// 注册到通信管理器
	commManager := nesmaService.GetCommunicationManager()
	commManager.RegisterAgent(req.AgentID, req.Endpoint)

	response.OkWithData(gin.H{"agentId": agent.AgentID}, c)
}

// GetAgent 获取Agent信息
func (a *AgentApi) GetAgent(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.FailWithMessage("Agent ID不能为空", c)
		return
	}

	agentService := nesmaService.GetAgentService()
	agent, err := agentService.GetAgent(c.Request.Context(), agentID)
	if err != nil {
		global.GVA_LOG.Error("Failed to get agent", zap.Error(err))
		response.FailWithMessage("获取Agent失败: "+err.Error(), c)
		return
	}

	response.OkWithData(agent, c)
}

// ListAgents 获取Agent列表
func (a *AgentApi) ListAgents(c *gin.Context) {
	agentService := nesmaService.GetAgentService()
	agents, err := agentService.ListAgents(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("Failed to list agents", zap.Error(err))
		response.FailWithMessage("获取Agent列表失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"agents": agents}, c)
}

// UpdateAgentStatus 更新Agent状态
func (a *AgentApi) UpdateAgentStatus(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.FailWithMessage("Agent ID不能为空", c)
		return
	}

	var req nesmaReq.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	agentService := nesmaService.GetAgentService()
	if err := agentService.UpdateAgentStatus(c.Request.Context(), agentID, req.Status); err != nil {
		global.GVA_LOG.Error("Failed to update agent status", zap.Error(err))
		response.FailWithMessage("更新Agent状态失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("状态更新成功", c)
}

// AgentHeartbeat Agent心跳
func (a *AgentApi) AgentHeartbeat(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.FailWithMessage("Agent ID不能为空", c)
		return
	}

	agentService := nesmaService.GetAgentService()
	if err := agentService.Heartbeat(c.Request.Context(), agentID); err != nil {
		global.GVA_LOG.Error("Failed to update heartbeat", zap.Error(err))
		response.FailWithMessage("心跳更新失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("心跳更新成功", c)
}

// CreateTask 创建任务
func (a *AgentApi) CreateTask(c *gin.Context) {
	var req nesmaReq.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	agentService := nesmaService.GetAgentService()

	task := &nesma.NesmaAgentTask{
		TaskType:   req.TaskType,
		Priority:   req.Priority,
		InputData:  req.InputData,
		SessionID:  req.SessionID,
		UserID:     req.UserID,
		ProjectID:  req.ProjectID,
		MaxRetries: req.MaxRetries,
		Metadata:   req.Metadata,
	}

	if err := agentService.CreateTask(c.Request.Context(), task); err != nil {
		global.GVA_LOG.Error("Failed to create task", zap.Error(err))
		response.FailWithMessage("创建任务失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"taskId": task.TaskID}, c)
}

// AssignTask 分配任务
func (a *AgentApi) AssignTask(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	var req nesmaReq.AssignTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	agentService := nesmaService.GetAgentService()
	if err := agentService.AssignTask(c.Request.Context(), taskID, req.AgentID); err != nil {
		global.GVA_LOG.Error("Failed to assign task", zap.Error(err))
		response.FailWithMessage("分配任务失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("任务分配成功", c)
}

// SendMessage 发送消息
func (a *AgentApi) SendMessage(c *gin.Context) {
	var req nesmaReq.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	agentService := nesmaService.GetAgentService()

	// 转换Content和Metadata为JSON格式
	var contentJSON datatypes.JSON
	var metadataJSON datatypes.JSON

	if req.Content != nil {
		contentJSON, _ = json.Marshal(req.Content)
	}
	if req.Metadata != nil {
		metadataJSON, _ = json.Marshal(req.Metadata)
	}

	message := &nesma.NesmaAgentMessage{
		MessageType:   req.Type,
		SourceAgentID: req.From,
		TargetAgentID: req.To,
		SessionID:     req.SessionID,
		Content:       contentJSON,
		Priority:      req.Priority,
		Metadata:      metadataJSON,
	}

	if err := agentService.SendMessage(c.Request.Context(), message); err != nil {
		global.GVA_LOG.Error("Failed to send message", zap.Error(err))
		response.FailWithMessage("发送消息失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("消息发送成功", c)
}

// ==================== 工作流管理 ====================

// CreateWorkflow 创建工作流
func (a *AgentApi) CreateWorkflow(c *gin.Context) {
	var req nesmaReq.CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	workflowEngine := nesmaService.GetWorkflowEngine()

	stepsData, err := json.Marshal(req.Steps)
	if err != nil {
		response.FailWithMessage("工作流步骤格式错误", c)
		return
	}

	workflow := &nesma.NesmaWorkflow{
		WorkflowID:  req.WorkflowID,
		Name:        req.Name,
		Description: req.Description,
		Definition:  stepsData,
		Status:      "active",
		Version:     req.Version,
	}

	if err := workflowEngine.CreateWorkflow(c.Request.Context(), workflow); err != nil {
		global.GVA_LOG.Error("Failed to create workflow", zap.Error(err))
		response.FailWithMessage("创建工作流失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"workflowId": workflow.WorkflowID}, c)
}

// GetWorkflow 获取工作流
func (a *AgentApi) GetWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		response.FailWithMessage("工作流ID不能为空", c)
		return
	}

	workflowEngine := nesmaService.GetWorkflowEngine()
	workflow, err := workflowEngine.GetWorkflow(c.Request.Context(), workflowID)
	if err != nil {
		global.GVA_LOG.Error("Failed to get workflow", zap.Error(err))
		response.FailWithMessage("获取工作流失败: "+err.Error(), c)
		return
	}

	response.OkWithData(workflow, c)
}

// ListWorkflows 获取工作流列表
func (a *AgentApi) ListWorkflows(c *gin.Context) {
	status := c.Query("status")

	workflowEngine := nesmaService.GetWorkflowEngine()
	workflows, err := workflowEngine.ListWorkflows(c.Request.Context(), status)
	if err != nil {
		global.GVA_LOG.Error("Failed to list workflows", zap.Error(err))
		response.FailWithMessage("获取工作流列表失败: "+err.Error(), c)
		return
	}

	response.OkWithData(gin.H{"workflows": workflows}, c)
}

// ExecuteWorkflow 执行工作流
func (a *AgentApi) ExecuteWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	if workflowID == "" {
		response.FailWithMessage("工作流ID不能为空", c)
		return
	}

	var req nesmaReq.ExecuteWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	workflowEngine := nesmaService.GetWorkflowEngine()
	execution, err := workflowEngine.ExecuteWorkflow(c.Request.Context(), workflowID, req.InputData)
	if err != nil {
		global.GVA_LOG.Error("Failed to execute workflow", zap.Error(err))
		response.FailWithMessage("执行工作流失败: "+err.Error(), c)
		return
	}

	response.OkWithData(execution, c)
}

// GetWorkflowExecution 获取工作流执行状态
func (a *AgentApi) GetWorkflowExecution(c *gin.Context) {
	executionID := c.Param("id")
	if executionID == "" {
		response.FailWithMessage("执行ID不能为空", c)
		return
	}

	workflowEngine := nesmaService.GetWorkflowEngine()
	execution, err := workflowEngine.GetWorkflowExecution(c.Request.Context(), executionID)
	if err != nil {
		global.GVA_LOG.Error("Failed to get workflow execution", zap.Error(err))
		response.FailWithMessage("获取工作流执行状态失败: "+err.Error(), c)
		return
	}

	response.OkWithData(execution, c)
}

// CancelWorkflowExecution 取消工作流执行
func (a *AgentApi) CancelWorkflowExecution(c *gin.Context) {
	executionID := c.Param("id")
	if executionID == "" {
		response.FailWithMessage("执行ID不能为空", c)
		return
	}

	workflowEngine := nesmaService.GetWorkflowEngine()
	if err := workflowEngine.CancelWorkflowExecution(c.Request.Context(), executionID); err != nil {
		global.GVA_LOG.Error("Failed to cancel workflow execution", zap.Error(err))
		response.FailWithMessage("取消工作流执行失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("工作流执行已取消", c)
}

// GetWorkflowStatistics 获取工作流统计信息
func (a *AgentApi) GetWorkflowStatistics(c *gin.Context) {
	workflowEngine := nesmaService.GetWorkflowEngine()
	stats, err := workflowEngine.GetWorkflowStatistics(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("Failed to get workflow statistics", zap.Error(err))
		response.FailWithMessage("获取工作流统计信息失败: "+err.Error(), c)
		return
	}

	response.OkWithData(stats, c)
}

// ==================== 统计和监控 ====================

// GetAgentStatistics 获取Agent统计信息
func (a *AgentApi) GetAgentStatistics(c *gin.Context) {
	agentService := nesmaService.GetAgentService()

	// 获取数据库中的Agent总数
	agents, err := agentService.ListAgents(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("Failed to get agents from database", zap.Error(err))
		response.FailWithMessage("获取Agent统计信息失败: "+err.Error(), c)
		return
	}

	// 计算活跃Agent数量
	activeAgents := 0
	for _, agent := range agents {
		if agent.Status == "online" || agent.Status == "running" {
			activeAgents++
		}
	}

	// 获取正在处理的任务数量
	processingTasks := 0
	for _, agent := range agents {
		tasks, err := agentService.GetTasksByAgent(c.Request.Context(), agent.AgentID)
		if err == nil {
			for _, task := range tasks {
				if task.Status == "running" || task.Status == "assigned" {
					processingTasks++
				}
			}
		}
	}

	// 获取消息总数（可选，如果需要的话）
	// 这里暂时设为0，因为消息统计比较复杂
	totalMessages := 0

	stats := gin.H{
		"totalAgents":     len(agents),
		"activeAgents":    activeAgents,
		"processingTasks": processingTasks,
		"totalMessages":   totalMessages,
	}

	response.OkWithData(stats, c)
}

// GetAgentHealth 获取Agent健康状态
func (a *AgentApi) GetAgentHealth(c *gin.Context) {
	agentService := nesmaService.GetAgentService()

	// 获取数据库中的Agent
	agents, err := agentService.ListAgents(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("Failed to get agents from database", zap.Error(err))
		response.FailWithMessage("获取Agent健康状态失败: "+err.Error(), c)
		return
	}

	// 计算在线Agent数量
	onlineAgents := 0
	var totalResponseTime int
	totalProcessed := 0

	for _, agent := range agents {
		if agent.Status == "online" || agent.Status == "running" {
			onlineAgents++
		}
		totalResponseTime += agent.AverageResponseTime
		totalProcessed += agent.TotalProcessed
	}

	// 计算平均响应时间
	avgResponseTime := 0
	if len(agents) > 0 {
		avgResponseTime = totalResponseTime / len(agents)
	}

	// 计算成功率（简化计算）
	successRate := 95.0 // 默认成功率，实际项目中应该根据任务完成情况计算

	// 获取正在处理的任务数量
	processingTasks := 0
	for _, agent := range agents {
		tasks, err := agentService.GetTasksByAgent(c.Request.Context(), agent.AgentID)
		if err == nil {
			for _, task := range tasks {
				if task.Status == "running" || task.Status == "assigned" {
					processingTasks++
				}
			}
		}
	}

	health := gin.H{
		"onlineAgents":    onlineAgents,
		"processingTasks": processingTasks,
		"avgResponseTime": avgResponseTime,
		"successRate":     successRate,
	}

	response.OkWithData(health, c)
}

// GetSystemMetrics 获取系统性能指标
func (a *AgentApi) GetSystemMetrics(c *gin.Context) {
	// 获取运行时内存信息
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 计算内存使用率（简化计算）
	memoryUsage := float64(m.Alloc) / float64(m.Sys) * 100

	// 获取Goroutine数量作为CPU使用的指标
	goroutines := runtime.NumGoroutine()

	// 计算CPU使用率（简化计算，基于Goroutine数量）
	cpuUsage := float64(goroutines) / 100.0 * 10 // 简化的CPU使用率计算
	if cpuUsage > 100 {
		cpuUsage = 100
	}

	// 网络和磁盘使用率设为合理的默认值
	networkUsage := 15.0 // 默认网络使用率
	diskUsage := 45.0    // 默认磁盘使用率

	metrics := gin.H{
		"cpuUsage":     cpuUsage,
		"memoryUsage":  memoryUsage,
		"networkUsage": networkUsage,
		"diskUsage":    diskUsage,
	}

	response.OkWithData(metrics, c)
}

// GetAgentTasks 获取特定Agent的任务列表
// @Tags Agent
// @Summary 获取特定Agent的任务列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "Agent ID"
// @Param status query string false "任务状态"
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Success 200 {object} response.Response{data=[]nesma.NesmaAgentTask} "获取成功"
// @Router /agent/{id}/tasks [get]
func (a *AgentApi) GetAgentTasks(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.FailWithMessage("Agent ID不能为空", c)
		return
	}

	status := c.Query("status")
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")

	agentService := nesmaService.GetAgentService()
	tasks, err := agentService.GetTasksByAgent(c.Request.Context(), agentID)
	if err != nil {
		global.GVA_LOG.Error("Failed to get agent tasks", zap.Error(err))
		response.FailWithMessage("获取Agent任务失败: "+err.Error(), c)
		return
	}

	// 状态过滤
	var filteredTasks []*nesma.NesmaAgentTask
	if status != "" {
		for _, task := range tasks {
			if task.Status == status {
				filteredTasks = append(filteredTasks, task)
			}
		}
	} else {
		filteredTasks = tasks
	}

	response.OkWithData(gin.H{
		"list":     filteredTasks,
		"total":    len(filteredTasks),
		"page":     page,
		"pageSize": pageSize,
	}, c)
}

// GetAgentLogs 获取特定Agent的日志
// @Tags Agent
// @Summary 获取特定Agent的日志
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path string true "Agent ID"
// @Param level query string false "日志级别"
// @Param limit query int false "日志数量限制"
// @Success 200 {object} response.Response{data=[]map[string]interface{}} "获取成功"
// @Router /agent/{id}/logs [get]
func (a *AgentApi) GetAgentLogs(c *gin.Context) {
	agentID := c.Param("id")
	if agentID == "" {
		response.FailWithMessage("Agent ID不能为空", c)
		return
	}

	level := c.Query("level")
	_ = c.DefaultQuery("limit", "100") // TODO: 实现日志数量限制

	// 模拟日志数据（实际项目中应该从日志系统获取）
	logs := []map[string]interface{}{
		{
			"id":        1,
			"timestamp": "2025-01-07T23:54:56+08:00",
			"level":     "info",
			"message":   "Agent " + agentID + " started successfully",
		},
		{
			"id":        2,
			"timestamp": "2025-01-07T23:54:57+08:00",
			"level":     "info",
			"message":   "Agent " + agentID + " registered to agent manager",
		},
		{
			"id":        3,
			"timestamp": "2025-01-07T23:54:58+08:00",
			"level":     "debug",
			"message":   "Agent " + agentID + " heartbeat sent",
		},
		{
			"id":        4,
			"timestamp": "2025-01-07T23:54:59+08:00",
			"level":     "info",
			"message":   "Agent " + agentID + " ready to receive tasks",
		},
		{
			"id":        5,
			"timestamp": "2025-01-07T23:55:00+08:00",
			"level":     "debug",
			"message":   "Agent " + agentID + " status check completed",
		},
	}

	// 级别过滤
	var filteredLogs []map[string]interface{}
	if level != "" {
		for _, log := range logs {
			if log["level"] == level {
				filteredLogs = append(filteredLogs, log)
			}
		}
	} else {
		filteredLogs = logs
	}

	response.OkWithData(gin.H{
		"list":  filteredLogs,
		"total": len(filteredLogs),
	}, c)
}
