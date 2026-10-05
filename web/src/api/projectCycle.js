import service from '@/utils/request'

// @Tags NesmaProjectCycle
// @Summary 创建项目周期
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectCycleRequest true "创建项目周期"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /nesma/project-cycle [post]
export const createProjectCycle = (data) => {
  return service({
    url: '/nesma/project-cycle',
    method: 'post',
    data
  })
}

// @Tags NesmaProjectCycle
// @Summary 删除项目周期
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "周期ID"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /nesma/project-cycle/{id} [delete]
export const deleteProjectCycle = (id) => {
  return service({
    url: `/nesma/project-cycle/${id}`,
    method: 'delete'
  })
}

// @Tags NesmaProjectCycle
// @Summary 更新项目周期
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectCycleRequest true "更新项目周期"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /nesma/project-cycle [put]
export const updateProjectCycle = (data) => {
  return service({
    url: '/nesma/project-cycle',
    method: 'put',
    data
  })
}

// @Tags NesmaProjectCycle
// @Summary 用ID查询项目周期
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param id path int true "周期ID"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /nesma/project-cycle/{id} [get]
export const getProjectCycle = (id) => {
  return service({
    url: `/nesma/project-cycle/${id}`,
    method: 'get'
  })
}

// @Tags NesmaProjectCycle
// @Summary 分页获取项目周期列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query nesmaReq.NesmaProjectCycleSearch true "分页获取项目周期列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /nesma/project-cycle/list [get]
export const getProjectCycleList = (params) => {
  return service({
    url: '/nesma/project-cycle/list',
    method: 'get',
    params
  })
}

// @Tags NesmaProjectCycle
// @Summary 获取指定项目的所有周期
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param projectId path int true "项目ID"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /nesma/project-cycle/project/{projectId} [get]
export const getProjectCycles = (projectId) => {
  return service({
    url: `/nesma/project-cycle/project/${projectId}`,
    method: 'get'
  })
}

// @Tags NesmaProjectCycle
// @Summary 更新周期状态
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "更新周期状态"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"状态更新成功"}"
// @Router /nesma/project-cycle/status [put]
export const updateCycleStatus = (data) => {
  return service({
    url: '/nesma/project-cycle/status',
    method: 'put',
    data
  })
}

// @Tags NesmaProjectCycle
// @Summary 设置项目的激活周期
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body object true "设置激活周期"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"激活周期设置成功"}"
// @Router /nesma/project-cycle/set-active [post]
export const setActiveProjectCycle = (data) => {
  return service({
    url: '/nesma/project-cycle/set-active',
    method: 'post',
    data
  })
}