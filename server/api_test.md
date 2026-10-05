# 分析任务管理API优化测试

## 优化内容

### 1. 后端优化
- ✅ 添加了 `AnalysisTaskListRequest` 请求结构体，支持：
  - 分页：`page`, `pageSize`
  - 任务类型筛选：`taskType` (requirement_analysis, description_generation, flowchart_generation, etc.)
  - 状态筛选：`status` (pending, running, completed, failed, cancelled)
  - 时间范围筛选：`startDate`, `endDate`
  - 项目和周期筛选：`projectId`, `cycleId`

- ✅ 添加了 `TaskStatisticsRequest` 请求结构体
- ✅ 创建了新的Service方法 `GetAnalysisTasksWithPagination`
- ✅ 更新了API接口 `MVPGetAnalysisTasks` 和 `MVPGetTaskStatistics`

### 2. 前端优化
- ✅ 更新了TaskManagementDialog组件以处理新的分页响应格式
- ✅ 修正了参数传递方式（projectId而不是project_id）

## API测试用例

### 1. 获取分析任务列表（分页+筛选）
```bash
# 基础分页查询
curl "http://localhost:8888/api/v1/nesma/analysis/tasks?projectId=7&page=1&pageSize=5"

# 按任务类型筛选
curl "http://localhost:8888/api/v1/nesma/analysis/tasks?projectId=7&taskType=requirement_analysis&page=1&pageSize=10"

# 按状态筛选
curl "http://localhost:8888/api/v1/nesma/analysis/tasks?projectId=7&status=completed&page=1&pageSize=10"

# 时间范围筛选
curl "http://localhost:8888/api/v1/nesma/analysis/tasks?projectId=7&startDate=2025-07-14&endDate=2025-07-15&page=1&pageSize=10"

# 组合筛选
curl "http://localhost:8888/api/v1/nesma/analysis/tasks?projectId=7&taskType=requirement_analysis&status=completed&page=1&pageSize=5"
```

### 2. 获取任务统计信息
```bash
# 项目级统计
curl "http://localhost:8888/api/v1/nesma/analysis/tasks/statistics?projectId=7"

# 周期级统计
curl "http://localhost:8888/api/v1/nesma/analysis/tasks/statistics?projectId=7&cycleId=1"
```

## 预期响应格式

### 任务列表响应
```json
{
  "code": 0,
  "data": {
    "list": [...],
    "total": 25,
    "page": 1,
    "pageSize": 10,
    "totalPages": 3,
    "hasMore": true
  },
  "msg": "success"
}
```

### 统计信息响应
```json
{
  "code": 0,
  "data": {
    "total": 25,
    "running": 2,
    "completed": 20,
    "failed": 2,
    "pending": 1,
    "cancelled": 0
  },
  "msg": "success"
}
```

## 前端使用示例

```javascript
// 获取分页任务列表
const response = await getProjectAnalysisTasks(null, {
  projectId: 7,
  page: 1,
  pageSize: 10,
  taskType: 'requirement_analysis',
  status: 'completed',
  startDate: '2025-07-14',
  endDate: '2025-07-15'
})

// 处理响应
if (response.code === 0) {
  taskList.value = response.data.list || []
  total.value = response.data.total || 0
}
```

## 验证清单
- [ ] 服务器启动正常
- [ ] 分页功能正常工作
- [ ] 任务类型筛选功能正常
- [ ] 状态筛选功能正常  
- [ ] 时间范围筛选功能正常
- [ ] 统计接口参数更新正常
- [ ] 前端页面显示正常
- [ ] 分页组件工作正常