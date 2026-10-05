package nesma

import (
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

type NesmaProjectCycleApi struct{}

// CreateNesmaProjectCycle 创建项目周期
func (api *NesmaProjectCycleApi) CreateNesmaProjectCycle(c *gin.Context) {
	var req nesmaReq.NesmaProjectCycleRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		global.GVA_LOG.Error("创建项目周期参数绑定失败!", zap.Error(err))
		response.FailWithMessage("参数错误: "+err.Error(), c)
		return
	}

	// 调试日志
	global.GVA_LOG.Info("创建项目周期请求", 
		zap.Uint("projectId", req.ProjectID),
		zap.String("name", req.Name),
		zap.String("description", req.Description),
		zap.String("status", req.Status))

	cycle, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.CreateNesmaProjectCycle(&req)
	if err != nil {
		global.GVA_LOG.Error("创建项目周期失败!", zap.Error(err))
		response.FailWithMessage("创建项目周期失败: "+err.Error(), c)
		return
	}
	response.OkWithData(cycle, c)
}

// GetNesmaProjectCycle 根据ID获取项目周期
func (api *NesmaProjectCycleApi) GetNesmaProjectCycle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	cycle, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.GetNesmaProjectCycle(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取项目周期失败!", zap.Error(err))
		response.FailWithMessage("获取项目周期失败", c)
		return
	}
	response.OkWithData(cycle, c)
}

// GetNesmaProjectCycleList 获取项目周期列表
func (api *NesmaProjectCycleApi) GetNesmaProjectCycleList(c *gin.Context) {
	var req nesmaReq.NesmaProjectCycleSearch
	err := c.ShouldBindQuery(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	list, total, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.GetNesmaProjectCycleList(&req)
	if err != nil {
		global.GVA_LOG.Error("获取项目周期列表失败!", zap.Error(err))
		response.FailWithMessage("获取项目周期列表失败", c)
		return
	}
	response.OkWithDetailed(response.PageResult{
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}, "获取成功", c)
}

// UpdateNesmaProjectCycle 更新项目周期
func (api *NesmaProjectCycleApi) UpdateNesmaProjectCycle(c *gin.Context) {
	var req nesmaReq.NesmaProjectCycleRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	cycle, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.UpdateNesmaProjectCycle(&req)
	if err != nil {
		global.GVA_LOG.Error("更新项目周期失败!", zap.Error(err))
		response.FailWithMessage("更新项目周期失败", c)
		return
	}
	response.OkWithData(cycle, c)
}

// DeleteNesmaProjectCycle 删除项目周期
func (api *NesmaProjectCycleApi) DeleteNesmaProjectCycle(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("ID格式错误", c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.DeleteNesmaProjectCycle(uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除项目周期失败!", zap.Error(err))
		response.FailWithMessage("删除项目周期失败", c)
		return
	}
	response.OkWithMessage("删除成功", c)
}

// GetProjectCycles 获取指定项目的所有周期
func (api *NesmaProjectCycleApi) GetProjectCycles(c *gin.Context) {
	projectIdStr := c.Param("projectId")
	if projectIdStr == "" {
		response.FailWithMessage("项目ID不能为空", c)
		return
	}

	projectId, err := strconv.ParseUint(projectIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("项目ID格式错误", c)
		return
	}

	cycles, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.GetProjectCycles(uint(projectId))
	if err != nil {
		global.GVA_LOG.Error("获取项目周期失败!", zap.Error(err))
		response.FailWithMessage("获取项目周期失败", c)
		return
	}
	response.OkWithData(cycles, c)
}

// UpdateCycleStatus 更新周期状态
func (api *NesmaProjectCycleApi) UpdateCycleStatus(c *gin.Context) {
	var req struct {
		ID     uint   `json:"id" binding:"required"`
		Status string `json:"status" binding:"required"`
	}
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.UpdateCycleStatus(req.ID, req.Status)
	if err != nil {
		global.GVA_LOG.Error("更新周期状态失败!", zap.Error(err))
		response.FailWithMessage("更新周期状态失败", c)
		return
	}
	response.OkWithMessage("状态更新成功", c)
}

// SetActiveProjectCycle 设置项目的激活周期
func (api *NesmaProjectCycleApi) SetActiveProjectCycle(c *gin.Context) {
	var req struct {
		CycleID   uint `json:"cycleId" binding:"required"`
	}
	
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectCycleService.SetActiveProjectCycle(req.CycleID)
	if err != nil {
		global.GVA_LOG.Error("设置激活周期失败!", zap.Error(err))
		response.FailWithMessage("设置激活周期失败", c)
		return
	}
	response.OkWithMessage("激活周期设置成功", c)
}