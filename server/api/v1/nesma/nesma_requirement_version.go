package nesma

import (
	"strconv"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/model/common/response"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NesmaRequirementVersionApi struct{}

// CreateRequirementVersion 创建需求版本
// @Tags NesmaRequirementVersion
// @Summary 创建需求版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaRequirementVersionRequest true "创建需求版本请求"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirementVersion} "创建成功"
// @Router /nesma/requirement-version [post]
func (a *NesmaRequirementVersionApi) CreateRequirementVersion(c *gin.Context) {
	var req nesmaReq.NesmaRequirementVersionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	version, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.CreateRequirementVersion(&req)
	if err != nil {
		global.GVA_LOG.Error("创建需求版本失败!", zap.Error(err))
		response.FailWithMessage("创建需求版本失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(version, "创建需求版本成功", c)
}

// GetRequirementVersions 获取周期的所有版本
// @Tags NesmaRequirementVersion
// @Summary 获取周期的所有版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Success 200 {object} response.Response{data=[]nesma.NesmaRequirementVersion} "获取成功"
// @Router /nesma/requirement-version/cycle/{cycleId} [get]
func (a *NesmaRequirementVersionApi) GetRequirementVersions(c *gin.Context) {
	cycleIdStr := c.Param("cycleId")
	cycleId, err := strconv.ParseUint(cycleIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	versions, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.GetRequirementVersions(uint(cycleId))
	if err != nil {
		global.GVA_LOG.Error("获取需求版本失败!", zap.Error(err))
		response.FailWithMessage("获取需求版本失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(versions, "获取需求版本成功", c)
}

// GetRequirementVersion 根据ID获取版本详情
// @Tags NesmaRequirementVersion
// @Summary 根据ID获取版本详情
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirementVersion} "获取成功"
// @Router /nesma/requirement-version/{id} [get]
func (a *NesmaRequirementVersionApi) GetRequirementVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("版本ID格式错误", c)
		return
	}

	version, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.GetRequirementVersion(uint(id))
	if err != nil {
		global.GVA_LOG.Error("获取版本详情失败!", zap.Error(err))
		response.FailWithMessage("获取版本详情失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(version, "获取版本详情成功", c)
}

// UpdateRequirementVersion 更新需求版本
// @Tags NesmaRequirementVersion
// @Summary 更新需求版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaRequirementVersionRequest true "更新需求版本请求"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirementVersion} "更新成功"
// @Router /nesma/requirement-version [put]
func (a *NesmaRequirementVersionApi) UpdateRequirementVersion(c *gin.Context) {
	var req nesmaReq.NesmaRequirementVersionRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		response.FailWithMessage(err.Error(), c)
		return
	}

	version, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.UpdateRequirementVersion(&req)
	if err != nil {
		global.GVA_LOG.Error("更新需求版本失败!", zap.Error(err))
		response.FailWithMessage("更新需求版本失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(version, "更新需求版本成功", c)
}

// DeleteRequirementVersion 删除需求版本
// @Tags NesmaRequirementVersion
// @Summary 删除需求版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response "删除成功"
// @Router /nesma/requirement-version/{id} [delete]
func (a *NesmaRequirementVersionApi) DeleteRequirementVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("版本ID格式错误", c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.DeleteRequirementVersion(uint(id))
	if err != nil {
		global.GVA_LOG.Error("删除需求版本失败!", zap.Error(err))
		response.FailWithMessage("删除需求版本失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("删除需求版本成功", c)
}

// GetActiveVersion 获取周期的激活版本
// @Tags NesmaRequirementVersion
// @Summary 获取周期的激活版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Success 200 {object} response.Response{data=nesma.NesmaRequirementVersion} "获取成功"
// @Router /nesma/requirement-version/active/{cycleId} [get]
func (a *NesmaRequirementVersionApi) GetActiveVersion(c *gin.Context) {
	cycleIdStr := c.Param("cycleId")
	cycleId, err := strconv.ParseUint(cycleIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	version, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.GetActiveVersion(uint(cycleId))
	if err != nil {
		global.GVA_LOG.Error("获取激活版本失败!", zap.Error(err))
		response.FailWithMessage("获取激活版本失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(version, "获取激活版本成功", c)
}

// SetActiveVersion 设置激活版本
// @Tags NesmaRequirementVersion
// @Summary 设置激活版本
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response "设置成功"
// @Router /nesma/requirement-version/active/{id} [put]
func (a *NesmaRequirementVersionApi) SetActiveVersion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("版本ID格式错误", c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.SetActiveVersion(uint(id))
	if err != nil {
		global.GVA_LOG.Error("设置激活版本失败!", zap.Error(err))
		response.FailWithMessage("设置激活版本失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("设置激活版本成功", c)
}

// GenerateNextVersion 生成下一个版本号
// @Tags NesmaRequirementVersion
// @Summary 生成下一个版本号
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param cycleId path int true "周期ID"
// @Param versionType query string true "版本类型"
// @Success 200 {object} response.Response{data=map[string]string} "生成成功"
// @Router /nesma/requirement-version/next/{cycleId} [get]
func (a *NesmaRequirementVersionApi) GenerateNextVersion(c *gin.Context) {
	cycleIdStr := c.Param("cycleId")
	cycleId, err := strconv.ParseUint(cycleIdStr, 10, 32)
	if err != nil {
		response.FailWithMessage("周期ID格式错误", c)
		return
	}

	versionType := c.DefaultQuery("versionType", "initial")

	nextVersion, err := service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.GenerateNextVersion(uint(cycleId), versionType)
	if err != nil {
		global.GVA_LOG.Error("生成下一个版本号失败!", zap.Error(err))
		response.FailWithMessage("生成下一个版本号失败: "+err.Error(), c)
		return
	}

	response.OkWithDetailed(map[string]string{"nextVersion": nextVersion}, "生成下一个版本号成功", c)
}

// UpdateVersionStats 更新版本统计信息
// @Tags NesmaRequirementVersion
// @Summary 更新版本统计信息
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "版本ID"
// @Success 200 {object} response.Response "更新成功"
// @Router /nesma/requirement-version/stats/{id} [put]
func (a *NesmaRequirementVersionApi) UpdateVersionStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.FailWithMessage("版本ID格式错误", c)
		return
	}

	err = service.ServiceGroupApp.NesmaServiceGroup.NesmaRequirementVersionService.UpdateVersionStats(uint(id))
	if err != nil {
		global.GVA_LOG.Error("更新版本统计失败!", zap.Error(err))
		response.FailWithMessage("更新版本统计失败: "+err.Error(), c)
		return
	}

	response.OkWithMessage("更新版本统计成功", c)
} 