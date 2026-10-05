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
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// WorkflowEngine 工作流引擎接口
type WorkflowEngine interface {
	CreateWorkflow(ctx context.Context, workflow *nesma.NesmaWorkflow) error
	GetWorkflow(ctx context.Context, workflowID string) (*nesma.NesmaWorkflow, error)
	UpdateWorkflow(ctx context.Context, workflow *nesma.NesmaWorkflow) error
	DeleteWorkflow(ctx context.Context, workflowID string) error
	ListWorkflows(ctx context.Context, status string) ([]*nesma.NesmaWorkflow, error)
	ExecuteWorkflow(ctx context.Context, workflowID string, inputData map[string]interface{}) (*nesma.NesmaWorkflowExecution, error)
	GetWorkflowExecution(ctx context.Context, executionID string) (*nesma.NesmaWorkflowExecution, error)
	GetWorkflowExecutions(ctx context.Context, workflowID string) ([]*nesma.NesmaWorkflowExecution, error)
	CancelWorkflowExecution(ctx context.Context, executionID string) error
	GetWorkflowStatistics(ctx context.Context) (*WorkflowStatistics, error)
}

// WorkflowEngineImpl 工作流引擎实现
type WorkflowEngineImpl struct {
	db           *gorm.DB
	agentService AgentService
	mu           sync.RWMutex
}

// WorkflowStatistics 工作流统计信息
type WorkflowStatistics struct {
	TotalWorkflows    int64   `json:"totalWorkflows"`
	ActiveWorkflows   int64   `json:"activeWorkflows"`
	TotalExecutions   int64   `json:"totalExecutions"`
	RunningExecutions int64   `json:"runningExecutions"`
	SuccessRate       float64 `json:"successRate"`
	AvgExecutionTime  float64 `json:"avgExecutionTime"`
}

// WorkflowStep 工作流步骤
type WorkflowStep struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	AgentType     string                 `json:"agentType"`
	Order         int                    `json:"order"`
	Dependencies  []string               `json:"dependencies"`
	Configuration map[string]interface{} `json:"configuration"`
	Timeout       int                    `json:"timeout"`
	Retries       int                    `json:"retries"`
	Description   string                 `json:"description"`
}

// NewWorkflowEngine 创建工作流引擎实例
func NewWorkflowEngine(agentService AgentService) WorkflowEngine {
	return &WorkflowEngineImpl{
		db:           global.GVA_DB,
		agentService: agentService,
	}
}

// CreateWorkflow 创建工作流
func (w *WorkflowEngineImpl) CreateWorkflow(ctx context.Context, workflow *nesma.NesmaWorkflow) error {
	if workflow.WorkflowID == "" {
		workflow.WorkflowID = uuid.New().String()
	}

	// 验证工作流定义
	if err := w.validateWorkflowDefinition(workflow.Definition); err != nil {
		return err
	}

	workflow.Status = "active"

	return w.db.WithContext(ctx).Create(workflow).Error
}

// validateWorkflowDefinition 验证工作流定义
func (w *WorkflowEngineImpl) validateWorkflowDefinition(definition datatypes.JSON) error {
	if len(definition) == 0 {
		return errors.New("workflow definition cannot be empty")
	}

	var steps []WorkflowStep
	if err := json.Unmarshal(definition, &steps); err != nil {
		return fmt.Errorf("invalid workflow definition format: %v", err)
	}

	if len(steps) == 0 {
		return errors.New("workflow must have at least one step")
	}

	// 验证步骤依赖关系
	stepMap := make(map[string]bool)
	for _, step := range steps {
		stepMap[step.ID] = true
	}

	for _, step := range steps {
		for _, dep := range step.Dependencies {
			if !stepMap[dep] {
				return fmt.Errorf("step %s depends on non-existent step %s", step.ID, dep)
			}
		}
	}

	return nil
}

// GetWorkflow 获取工作流
func (w *WorkflowEngineImpl) GetWorkflow(ctx context.Context, workflowID string) (*nesma.NesmaWorkflow, error) {
	var workflow nesma.NesmaWorkflow
	err := w.db.WithContext(ctx).Where("workflow_id = ?", workflowID).First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}

// UpdateWorkflow 更新工作流
func (w *WorkflowEngineImpl) UpdateWorkflow(ctx context.Context, workflow *nesma.NesmaWorkflow) error {
	// 验证工作流定义
	if err := w.validateWorkflowDefinition(workflow.Definition); err != nil {
		return err
	}

	return w.db.WithContext(ctx).Where("workflow_id = ?", workflow.WorkflowID).Updates(workflow).Error
}

// DeleteWorkflow 删除工作流
func (w *WorkflowEngineImpl) DeleteWorkflow(ctx context.Context, workflowID string) error {
	// 检查是否有正在运行的执行
	var count int64
	err := w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).
		Where("workflow_id = ? AND status IN ?", workflowID, []string{"running", "pending"}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		return errors.New("cannot delete workflow with running executions")
	}

	return w.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Delete(&nesma.NesmaWorkflow{}).Error
}

// ListWorkflows 获取工作流列表
func (w *WorkflowEngineImpl) ListWorkflows(ctx context.Context, status string) ([]*nesma.NesmaWorkflow, error) {
	var workflows []*nesma.NesmaWorkflow
	query := w.db.WithContext(ctx)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Find(&workflows).Error
	return workflows, err
}

// ExecuteWorkflow 执行工作流
func (w *WorkflowEngineImpl) ExecuteWorkflow(ctx context.Context, workflowID string, inputData map[string]interface{}) (*nesma.NesmaWorkflowExecution, error) {
	// 获取工作流定义
	workflow, err := w.GetWorkflow(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	if workflow.Status != "active" {
		return nil, errors.New("workflow is not active")
	}

	// 转换输入数据为JSON
	inputJSON, err := json.Marshal(inputData)
	if err != nil {
		return nil, err
	}

	// 创建执行记录
	startTime := time.Now()
	execution := &nesma.NesmaWorkflowExecution{
		ExecutionID: uuid.New().String(),
		WorkflowID:  workflowID,
		SessionID:   fmt.Sprintf("session-%s", uuid.New().String()),
		Status:      "running",
		InputData:   inputJSON,
		StartTime:   &startTime,
	}

	if err := w.db.WithContext(ctx).Create(execution).Error; err != nil {
		return nil, err
	}

	// 异步执行工作流
	go w.executeWorkflowAsync(ctx, execution, workflow)

	return execution, nil
}

// executeWorkflowAsync 异步执行工作流
func (w *WorkflowEngineImpl) executeWorkflowAsync(ctx context.Context, execution *nesma.NesmaWorkflowExecution, workflow *nesma.NesmaWorkflow) {
	defer func() {
		if r := recover(); r != nil {
			w.updateExecutionStatus(ctx, execution.ExecutionID, "failed", fmt.Sprintf("workflow execution panic: %v", r))
		}
	}()

	// 解析工作流步骤
	var steps []WorkflowStep
	if err := json.Unmarshal(workflow.Definition, &steps); err != nil {
		w.updateExecutionStatus(ctx, execution.ExecutionID, "failed", fmt.Sprintf("failed to parse workflow definition: %v", err))
		return
	}

	// 模拟执行步骤
	for i, step := range steps {
		// 更新当前步骤
		w.updateCurrentStep(ctx, execution.ExecutionID, step.Name)

		// 模拟步骤执行
		time.Sleep(time.Duration(2+i) * time.Second)

		// 检查执行是否被取消
		currentExecution, err := w.GetWorkflowExecution(ctx, execution.ExecutionID)
		if err != nil || currentExecution.Status == "cancelled" {
			return
		}
	}

	// 更新完成状态
	w.updateExecutionStatus(ctx, execution.ExecutionID, "completed", "")
}

// updateExecutionStatus 更新执行状态
func (w *WorkflowEngineImpl) updateExecutionStatus(ctx context.Context, executionID, status, errorMessage string) {
	updateData := map[string]interface{}{
		"status": status,
	}

	if status == "completed" || status == "failed" || status == "cancelled" {
		endTime := time.Now()
		updateData["end_time"] = &endTime
	}

	if errorMessage != "" {
		updateData["error_message"] = errorMessage
	}

	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).
		Where("execution_id = ?", executionID).
		Updates(updateData)
}

// updateCurrentStep 更新当前步骤
func (w *WorkflowEngineImpl) updateCurrentStep(ctx context.Context, executionID, currentStep string) {
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).
		Where("execution_id = ?", executionID).
		Update("current_step", currentStep)
}

// GetWorkflowExecution 获取工作流执行状态
func (w *WorkflowEngineImpl) GetWorkflowExecution(ctx context.Context, executionID string) (*nesma.NesmaWorkflowExecution, error) {
	var execution nesma.NesmaWorkflowExecution
	err := w.db.WithContext(ctx).Where("execution_id = ?", executionID).First(&execution).Error
	if err != nil {
		return nil, err
	}
	return &execution, nil
}

// GetWorkflowExecutions 获取工作流的所有执行记录
func (w *WorkflowEngineImpl) GetWorkflowExecutions(ctx context.Context, workflowID string) ([]*nesma.NesmaWorkflowExecution, error) {
	var executions []*nesma.NesmaWorkflowExecution
	err := w.db.WithContext(ctx).Where("workflow_id = ?", workflowID).Order("start_time DESC").Find(&executions).Error
	return executions, err
}

// CancelWorkflowExecution 取消工作流执行
func (w *WorkflowEngineImpl) CancelWorkflowExecution(ctx context.Context, executionID string) error {
	w.updateExecutionStatus(ctx, executionID, "cancelled", "")
	return nil
}

// GetWorkflowStatistics 获取工作流统计信息
func (w *WorkflowEngineImpl) GetWorkflowStatistics(ctx context.Context) (*WorkflowStatistics, error) {
	var stats WorkflowStatistics

	// 总工作流数
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflow{}).Count(&stats.TotalWorkflows)

	// 活跃工作流数
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflow{}).Where("status = ?", "active").Count(&stats.ActiveWorkflows)

	// 总执行数
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).Count(&stats.TotalExecutions)

	// 运行中的执行数
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).Where("status = ?", "running").Count(&stats.RunningExecutions)

	// 成功率
	var completedCount, totalCount int64
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).Where("status = ?", "completed").Count(&completedCount)
	w.db.WithContext(ctx).Model(&nesma.NesmaWorkflowExecution{}).Where("status IN ?", []string{"completed", "failed"}).Count(&totalCount)

	if totalCount > 0 {
		stats.SuccessRate = float64(completedCount) / float64(totalCount) * 100
	}

	// 平均执行时间（模拟数据）
	stats.AvgExecutionTime = 45.0

	return &stats, nil
}

// GetWorkflowEngine 获取工作流引擎实例
func GetWorkflowEngine() WorkflowEngine {
	return workflowEngine
}

// 全局工作流引擎实例
var workflowEngine WorkflowEngine

// InitWorkflowEngine 初始化工作流引擎
func InitWorkflowEngine() {
	agentService := GetAgentService()
	workflowEngine = NewWorkflowEngine(agentService)
}
