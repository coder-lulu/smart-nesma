package nesma

import (
	"encoding/json"
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	"github.com/flipped-aurora/gin-vue-admin/server/model/nesma"
	nesmaService "github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/datatypes"
)

type AgentTestApi struct{}

// TestRequirementAnalysisAgent 测试需求细化Agent
// @Tags AgentTest
// @Summary 测试需求细化Agent
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body string true "需求文本"
// @Success 200 {object} response.Response{data=nesma.NesmaAgentTask} "测试成功"
// @Router /agentTest/requirementAnalysis [post]
func (a *AgentTestApi) TestRequirementAnalysisAgent(c *gin.Context) {
	var req struct {
		RequirementText string `json:"requirement_text" binding:"required"`
		Context         string `json:"context"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 创建测试任务
	task := &nesma.NesmaAgentTask{
		TaskID:    uuid.New().String(),
		AgentID:   "requirement-analysis-agent",
		TaskType:  "requirement_analysis",
		Priority:  2,
		Status:    "pending",
		UserID:    utils.GetUserID(c),
		SessionID: c.GetHeader("Session-ID"),
	}

	// 准备输入数据
	inputData := map[string]interface{}{
		"requirement_text": req.RequirementText,
		"context":          req.Context,
	}
	inputJSON, _ := json.Marshal(inputData)
	task.InputData = datatypes.JSON(inputJSON)

	// 创建任务
	agentService := nesmaService.GetAgentService()
	err = agentService.CreateTask(c.Request.Context(), task)
	if err != nil {
		global.GVA_LOG.Error("创建Agent测试任务失败!", zap.Error(err))
		response.FailWithMessage("创建任务失败: "+err.Error(), c)
		return
	}

	// 分配任务
	err = agentService.AssignTask(c.Request.Context(), task.TaskID, task.AgentID)
	if err != nil {
		global.GVA_LOG.Error("分配Agent任务失败!", zap.Error(err))
		response.FailWithMessage("分配任务失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(task, "需求细化Agent测试任务已创建", c)
}

// TestNESMAEvaluationAgent 测试NESMA评估Agent
// @Tags AgentTest
// @Summary 测试NESMA评估Agent
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body []interface{} true "需求列表"
// @Success 200 {object} response.Response{data=nesma.NesmaAgentTask} "测试成功"
// @Router /agentTest/nesmaEvaluation [post]
func (a *AgentTestApi) TestNESMAEvaluationAgent(c *gin.Context) {
	var req struct {
		Requirements []interface{} `json:"requirements" binding:"required"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 创建测试任务
	task := &nesma.NesmaAgentTask{
		TaskID:    uuid.New().String(),
		AgentID:   "nesma-evaluation-agent",
		TaskType:  "nesma_evaluation",
		Priority:  2,
		Status:    "pending",
		UserID:    utils.GetUserID(c),
		SessionID: c.GetHeader("Session-ID"),
	}

	// 准备输入数据
	inputData := map[string]interface{}{
		"requirements": req.Requirements,
	}
	inputJSON, _ := json.Marshal(inputData)
	task.InputData = datatypes.JSON(inputJSON)

	// 创建和分配任务
	agentService := nesmaService.GetAgentService()
	err = agentService.CreateTask(c.Request.Context(), task)
	if err != nil {
		global.GVA_LOG.Error("创建Agent测试任务失败!", zap.Error(err))
		response.FailWithMessage("创建任务失败: "+err.Error(), c)
		return
	}

	err = agentService.AssignTask(c.Request.Context(), task.TaskID, task.AgentID)
	if err != nil {
		global.GVA_LOG.Error("分配Agent任务失败!", zap.Error(err))
		response.FailWithMessage("分配任务失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(task, "NESMA评估Agent测试任务已创建", c)
}

// TestDocumentGenerationAgent 测试文档生成Agent
// @Tags AgentTest
// @Summary 测试文档生成Agent
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body map[string]interface{} true "文档生成参数"
// @Success 200 {object} response.Response{data=nesma.NesmaAgentTask} "测试成功"
// @Router /agentTest/documentGeneration [post]
func (a *AgentTestApi) TestDocumentGenerationAgent(c *gin.Context) {
	var req struct {
		DocumentType string                 `json:"document_type" binding:"required"`
		Data         map[string]interface{} `json:"data"`
	}

	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	// 创建测试任务
	task := &nesma.NesmaAgentTask{
		TaskID:    uuid.New().String(),
		AgentID:   "document-generation-agent",
		TaskType:  "document_generation",
		Priority:  2,
		Status:    "pending",
		UserID:    utils.GetUserID(c),
		SessionID: c.GetHeader("Session-ID"),
	}

	// 准备输入数据
	inputData := map[string]interface{}{
		"document_type": req.DocumentType,
		"data":          req.Data,
	}
	inputJSON, _ := json.Marshal(inputData)
	task.InputData = datatypes.JSON(inputJSON)

	// 创建和分配任务
	agentService := nesmaService.GetAgentService()
	err = agentService.CreateTask(c.Request.Context(), task)
	if err != nil {
		global.GVA_LOG.Error("创建Agent测试任务失败!", zap.Error(err))
		response.FailWithMessage("创建任务失败: "+err.Error(), c)
		return
	}

	err = agentService.AssignTask(c.Request.Context(), task.TaskID, task.AgentID)
	if err != nil {
		global.GVA_LOG.Error("分配Agent任务失败!", zap.Error(err))
		response.FailWithMessage("分配任务失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(task, "文档生成Agent测试任务已创建", c)
}

// GetTaskResult 获取任务结果
// @Tags AgentTest
// @Summary 获取Agent任务结果
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param taskId path string true "任务ID"
// @Success 200 {object} response.Response{data=nesma.NesmaAgentTask} "获取成功"
// @Router /agentTest/taskResult/{taskId} [get]
func (a *AgentTestApi) GetTaskResult(c *gin.Context) {
	taskID := c.Param("taskId")
	if taskID == "" {
		response.FailWithMessage("任务ID不能为空", c)
		return
	}

	agentService := nesmaService.GetAgentService()
	task, err := agentService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		global.GVA_LOG.Error("获取任务失败!", zap.Error(err))
		response.FailWithMessage("获取任务失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(task, "获取任务结果成功", c)
}

// ListAgents 获取所有内置Agent
// @Tags AgentTest
// @Summary 获取所有内置Agent
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {object} response.Response{data=[]nesma.NesmaAgent} "获取成功"
// @Router /agentTest/agents [get]
func (a *AgentTestApi) ListAgents(c *gin.Context) {
	agentService := nesmaService.GetAgentService()
	agents, err := agentService.ListAgents(c.Request.Context())
	if err != nil {
		global.GVA_LOG.Error("获取Agent列表失败!", zap.Error(err))
		response.FailWithMessage("获取Agent列表失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(agents, "获取Agent列表成功", c)
}

// GetAgentTasks 获取Agent任务列表
// @Tags AgentTest
// @Summary 获取Agent任务列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param page query int false "页码"
// @Param pageSize query int false "每页大小"
// @Param status query string false "任务状态"
// @Success 200 {object} response.Response{data=[]nesma.NesmaAgentTask} "获取成功"
// @Router /agentTest/tasks [get]
func (a *AgentTestApi) GetAgentTasks(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	status := c.Query("status")

	agentService := nesmaService.GetAgentService()

	var tasks []*nesma.NesmaAgentTask
	var err error

	if status != "" {
		tasks, err = agentService.GetTasksByStatus(c.Request.Context(), status)
	} else {
		// 获取用户的所有任务（这里需要扩展AgentService增加这个方法）
		tasks, err = agentService.GetTasksByStatus(c.Request.Context(), "")
	}

	if err != nil {
		global.GVA_LOG.Error("获取任务列表失败!", zap.Error(err))
		response.FailWithMessage("获取任务列表失败: "+err.Error(), c)
		return
	}

	// 简单分页处理
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > len(tasks) {
		tasks = []*nesma.NesmaAgentTask{}
	} else if end > len(tasks) {
		tasks = tasks[start:]
	} else {
		tasks = tasks[start:end]
	}

	result := map[string]interface{}{
		"list":     tasks,
		"total":    len(tasks),
		"page":     page,
		"pageSize": pageSize,
	}

	response.OkWithDetailed(result, "获取任务列表成功", c)
}
