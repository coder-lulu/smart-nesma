import service from '@/utils/request'

// @Tags NesmaProject
// @Summary 创建NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectCreate true "创建NESMA项目"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"创建成功"}"
// @Router /nesma/project [post]
export const createNesmaProject = (data) => {
  return service({
    url: '/nesma/project',
    method: 'post',
    data
  })
}

// @Tags NesmaProject
// @Summary 删除NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectByID true "删除NESMA项目"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"删除成功"}"
// @Router /nesma/project [delete]
export const deleteNesmaProject = (data) => {
  return service({
    url: '/nesma/project',
    method: 'delete',
    data
  })
}

// @Tags NesmaProject
// @Summary 批量删除NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectDeleteByIds true "批量删除NESMA项目"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"批量删除成功"}"
// @Router /nesma/project/delete-batch [delete]
export const deleteNesmaProjectByIds = (data) => {
  return service({
    url: '/nesma/project/delete-batch',
    method: 'delete',
    data
  })
}

// @Tags NesmaProject
// @Summary 更新NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectUpdate true "更新NESMA项目"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"更新成功"}"
// @Router /nesma/project [put]
export const updateNesmaProject = (data) => {
  return service({
    url: '/nesma/project',
    method: 'put',
    data
  })
}

// @Tags NesmaProject
// @Summary 用ID查询NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param ID path int true "项目ID"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"查询成功"}"
// @Router /nesma/project/{ID} [get]
export const findNesmaProject = (ID) => {
  return service({
    url: `/nesma/project/${ID}`,
    method: 'get'
  })
}

// @Tags NesmaProject
// @Summary 分页获取NESMA项目列表
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data query nesmaReq.NesmaProjectSearch true "分页获取NESMA项目列表"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /nesma/project/list [get]
export const getNesmaProjectList = (params) => {
  return service({
    url: '/nesma/project/list',
    method: 'get',
    params
  })
}

// @Tags NesmaProject
// @Summary 获取NESMA项目统计
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Success 200 {string} string "{"success":true,"data":{},"msg":"获取成功"}"
// @Router /nesma/project/stats [get]
export const getNesmaProjectStats = (params) => {
  return service({
    url: '/nesma/project/stats',
    method: 'get',
    params
  })
}

// @Tags NesmaProject
// @Summary 归档NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectByID true "归档NESMA项目"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"归档成功"}"
// @Router /nesma/project/archive [post]
export const archiveNesmaProject = (data) => {
  return service({
    url: '/nesma/project/archive',
    method: 'post',
    data
  })
}

// @Tags NesmaProject
// @Summary 恢复NESMA项目
// @Security ApiKeyAuth
// @accept application/json
// @Produce application/json
// @Param data body nesmaReq.NesmaProjectByID true "恢复NESMA项目"
// @Success 200 {string} string "{"success":true,"data":{},"msg":"恢复成功"}"
// @Router /nesma/project/restore [post]
export const restoreNesmaProject = (data) => {
  return service({
    url: '/nesma/project/restore',
    method: 'post',
    data
  })
} 