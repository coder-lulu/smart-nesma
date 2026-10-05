<template>
  <div>
    <!-- 警告提示条 -->
    <div class="warning-bar" style="background: #fdf6ec; border: 1px solid #fbe6c8; padding: 12px; margin-bottom: 20px; border-radius: 4px; color: #e6a23c;">
      <i class="el-icon-warning" style="margin-right: 8px;"></i>
      管理项目的NESMA功能点评估，查看评估结果和质量指标
    </div>
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <el-button type="primary" icon="plus" @click="handleCreate">
          新建评估
        </el-button>
        <el-button @click="getTableData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <!-- 筛选搜索 -->
      <div class="gva-search-box">
        <el-form ref="searchFormRef" :inline="true" :model="searchForm">
          <el-form-item label="项目" prop="projectId">
            <el-select v-model="searchForm.projectId" placeholder="选择项目" clearable filterable @change="getTableData">
              <el-option
                v-for="project in projectOptions"
                :key="project.value"
                :label="project.label"
                :value="project.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="评估状态" prop="status">
            <el-select v-model="searchForm.status" placeholder="评估状态" clearable @change="getTableData">
              <el-option label="待开始" value="pending" />
              <el-option label="进行中" value="processing" />
              <el-option label="已完成" value="completed" />
              <el-option label="失败" value="failed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="评估类型" prop="evaluationType">
            <el-select v-model="searchForm.evaluationType" placeholder="评估类型" clearable @change="getTableData">
              <el-option label="初步评估" value="initial" />
              <el-option label="详细评估" value="detailed" />
              <el-option label="最终评估" value="final" />
            </el-select>
          </el-form-item>
          <el-form-item label="评估人" prop="evaluatorId">
            <el-select v-model="searchForm.evaluatorId" placeholder="评估人" clearable filterable @change="getTableData">
              <el-option
                v-for="user in userOptions"
                :key="user.value"
                :label="user.nick_name"
                :value="user.value"
              />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="search" @click="getTableData">查询</el-button>
            <el-button icon="refresh" @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 评估表格 -->
      <el-table
        :data="tableData"
        style="width: 100%"
        tooltip-effect="dark"
        @selection-change="handleSelectionChange"
      >
        <el-table-column type="selection" width="55" />
        <el-table-column align="left" label="创建时间" width="180">
          <template #default="scope">
            <span>{{ formatDate(scope.row.startTime) }}</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="评估名称" prop="evaluationName" min-width="200" />
        <el-table-column align="left" label="项目名称" prop="projectName" width="150" />
        <el-table-column align="left" label="评估类型" width="120">
          <template #default="scope">
            <el-tag :type="getTypeColor(scope.row.evaluationType)" size="small">
              {{ getTypeText(scope.row.evaluationType) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="评估状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="功能点数" width="100">
          <template #default="scope">
            <span v-if="scope.row.totalFunctionPoints">{{ scope.row.totalFunctionPoints }}</span>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="评估人" prop="evaluatorName" width="120" />
        <el-table-column align="left" label="操作" min-width="200">
          <template #default="scope">
            <el-button type="primary" link icon="view" @click="handleView(scope.row)">
              查看
            </el-button>
            <el-button type="primary" link icon="edit" @click="handleEdit(scope.row)">
              编辑
            </el-button>
            <el-button 
              type="warning" 
              link 
              icon="setting" 
              @click="handleConfigureFactors(scope.row)"
            >
              配置因子
            </el-button>
            <el-button 
              v-if="scope.row.status === 'pending'"
              type="success" 
              link 
              icon="video-play" 
              @click="handleStart(scope.row)"
            >
              开始
            </el-button>
            <el-button 
              v-if="scope.row.status === 'processing'"
              type="primary" 
              link 
              icon="clock" 
              @click="handleViewProgress(scope.row)"
            >
              查看进度
            </el-button>
            <el-button 
              v-if="scope.row.status === 'processing'"
              type="danger" 
              link 
              icon="close" 
              @click="handleCancelEvaluation(scope.row)"
            >
              取消
            </el-button>
            <el-button 
              v-if="scope.row.status === 'completed'"
              type="success" 
              link 
              icon="document" 
              @click="handleReport(scope.row)"
            >
              报告
            </el-button>
            <el-button type="danger" link icon="delete" @click="handleDelete(scope.row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 分页 -->
      <div class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
    </div>

    <!-- 对话框组件 -->
    <CreateEvaluationDialog
      v-model="createDialogVisible"
      :edit-data="editEvaluation"
      @success="handleCreateSuccess"
    />

    <EvaluationDetailDialog
      v-model="detailDialogVisible"
      :evaluation-id="currentEvaluation?.id"
      @close="handleDetailClose"
    />

    <EvaluationFactorsDialog
      v-model="factorsDialogVisible"
      :evaluation-id="currentEvaluation?.id"
      :evaluation-info="currentEvaluationInfo"
      @success="handleFactorsSuccess"
    />

    <EvaluationProgressDialog
      v-model="progressDialogVisible"
      :evaluation-id="currentEvaluation?.id"
      :evaluation-name="currentEvaluation?.evaluationName"
      @close="handleProgressClose"
    />
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox, ElLoading } from 'element-plus'
import { Refresh, Plus, Search, View, Edit, Delete, Document } from '@element-plus/icons-vue'

// 组件导入
import CreateEvaluationDialog from './components/CreateEvaluationDialog.vue'
import EvaluationDetailDialog from './components/EvaluationDetailDialog.vue'
import EvaluationFactorsDialog from './components/EvaluationFactorsDialog.vue'
import EvaluationProgressDialog from './components/EvaluationProgressDialog.vue'

// API导入
import {
  getEvaluationList,
  deleteEvaluation,
  startEvaluation,
  cancelEvaluation
} from '@/api/nesmaEvaluation'
import {
  getProjectOptions,
  getUserOptions
} from '@/api/nesma'

// 响应式数据
const loading = ref(false)
const searchFormRef = ref(null) // 表单引用
const searchForm = reactive({
  projectId: '',
  status: '',
  evaluationType: '',
  evaluatorId: ''
})

// 表格数据
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// 选项数据
const projectOptions = ref([])
const userOptions = ref([])

// 对话框状态
const createDialogVisible = ref(false)
const detailDialogVisible = ref(false)
const factorsDialogVisible = ref(false)
const progressDialogVisible = ref(false)
const currentEvaluation = ref({})
const editEvaluation = ref(null)

// 当前评估信息（用于因子配置对话框）
const currentEvaluationInfo = ref({
  evaluationName: '',
  projectName: '',
  nesmaRules: 'v2.2'
})

// 方法实现
const handleCreate = () => {
  editEvaluation.value = null // 清空编辑数据
  createDialogVisible.value = true
}

const handleEdit = (evaluation) => {
  editEvaluation.value = { ...evaluation } // 复制评估数据用于编辑
  createDialogVisible.value = true
}

const handleView = (evaluation) => {
  currentEvaluation.value = evaluation
  detailDialogVisible.value = true
}

const handleConfigureFactors = (evaluation) => {
  currentEvaluation.value = evaluation
  // 设置评估信息用于因子配置对话框
  currentEvaluationInfo.value = {
    evaluationName: evaluation.evaluationName || '',
    projectName: evaluation.projectName || '',
    nesmaRules: evaluation.nesmaRules || 'v2.2'
  }
  factorsDialogVisible.value = true
}

const handleStart = async (evaluation) => {
  try {
    await ElMessageBox.confirm('确定要开始此评估吗？评估过程需要一定时间。', '确认开始评估', {
      type: 'warning',
      confirmButtonText: '开始评估',
      cancelButtonText: '取消'
    })
    
    // 显示加载状态
    const loadingInstance = ElLoading.service({
      lock: true,
      text: '正在启动评估...',
      background: 'rgba(0, 0, 0, 0.7)'
    })
    
    try {
      // 调用开始评估API
      await startEvaluation(evaluation.id)
      ElMessage.success('评估已开始，正在后台处理')
      // 刷新列表以显示更新的状态
      getTableData()
    } catch (error) {
      console.error('开始评估失败:', error)
      ElMessage.error('开始评估失败: ' + (error.response?.data?.msg || error.message))
    } finally {
      loadingInstance.close()
    }
  } catch (error) {
    // 用户取消，不做任何操作
    if (error !== 'cancel') {
      console.error('开始评估操作异常:', error)
    }
  }
}

const handleViewProgress = (evaluation) => {
  currentEvaluation.value = evaluation
  progressDialogVisible.value = true
}

const handleCancelEvaluation = async (evaluation) => {
  try {
    await ElMessageBox.confirm('确定要取消此评估吗？取消后评估将停止执行。', '确认取消评估', {
      type: 'warning',
      confirmButtonText: '确定取消',
      cancelButtonText: '继续评估'
    })
    
    // 显示取消加载状态
    const loadingInstance = ElLoading.service({
      lock: true,
      text: '正在取消评估...',
      background: 'rgba(0, 0, 0, 0.7)'
    })
    
    try {
      // 调用取消评估API
      await cancelEvaluation(evaluation.id, '用户手动取消')
      ElMessage.success('评估已取消')
      // 刷新列表以显示更新的状态
      getTableData()
    } catch (error) {
      console.error('取消评估失败:', error)
      ElMessage.error('取消评估失败: ' + (error.response?.data?.msg || error.message))
    } finally {
      loadingInstance.close()
    }
  } catch (error) {
    // 用户取消，不做任何操作
    if (error !== 'cancel') {
      console.error('取消评估操作异常:', error)
    }
  }
}

const handleReport = (evaluation) => {
  // 对于已完成的评估，显示详情（包含报告）
  handleView(evaluation)
}

const handleDelete = async (evaluation) => {
  try {
    await ElMessageBox.confirm('确定要删除此评估吗？此操作不可恢复！', '确认删除', {
      type: 'error',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
    
    // 显示删除加载状态
    const loadingInstance = ElLoading.service({
      lock: true,
      text: '正在删除评估...',
      background: 'rgba(0, 0, 0, 0.7)'
    })
    
    try {
      // 调用真正的删除API
      await deleteEvaluation(evaluation.id)
      ElMessage.success('评估删除成功')
      // 刷新列表
      getTableData()
    } catch (error) {
      console.error('删除评估失败:', error)
      ElMessage.error('删除评估失败: ' + (error.response?.data?.msg || error.message))
    } finally {
      loadingInstance.close()
    }
  } catch (error) {
    // 用户取消删除，不做任何操作
    if (error !== 'cancel') {
      console.error('删除操作异常:', error)
    }
  }
}

const handleSelectionChange = (selection) => {
  // 处理选择变更
}

// 搜索重置
const resetSearch = () => {
  // 使用表单的resetFields方法重置
  searchFormRef.value?.resetFields()
  // 手动重置所有字段（确保完全重置）
  searchForm.projectId = ''
  searchForm.status = ''
  searchForm.evaluationType = ''
  searchForm.evaluatorId = ''
  page.value = 1
  getTableData()
}

// 辅助方法
const getTypeColor = (type) => {
  const typeMap = {
    'initial': 'warning',
    'detailed': 'primary',
    'final': 'success'
  }
  return typeMap[type] || 'info'
}

const getTypeText = (type) => {
  const typeMap = {
    'initial': '初步评估',
    'detailed': '详细评估',
    'final': '最终评估'
  }
  return typeMap[type] || '未知'
}

const getStatusType = (status) => {
  const statusMap = {
    'pending': 'warning',
    'processing': 'primary',
    'completed': 'success',
    'failed': 'danger',
    'cancelled': 'danger'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    'pending': '待开始',
    'processing': '进行中',
    'completed': '已完成',
    'failed': '失败',
    'cancelled': '已取消'
  }
  return statusMap[status] || '未知'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN')
}

// 数据加载
const getTableData = async () => {
  loading.value = true
  try {
    const params = {
      page: page.value,
      pageSize: pageSize.value,
      ...searchForm
    }
    
    const response = await getEvaluationList(params)
    tableData.value = (response.data.list || []).map(evaluation => ({
      ...evaluation,
      id: evaluation.id ?? evaluation.ID
    }))
    total.value = response.data.total || 0
  } catch (error) {
    ElMessage.error('加载评估列表失败：' + error.message)
  } finally {
    loading.value = false
  }
}

const loadOptions = async () => {
  try {
    const [projectRes, userRes] = await Promise.all([
      getProjectOptions(),
      getUserOptions()
    ])
    projectOptions.value = projectRes.data || []
    userOptions.value = userRes.data || []
  } catch (error) {
    console.error('加载选项失败:', error)
  }
}

// 对话框事件
const handleCreateSuccess = () => {
  createDialogVisible.value = false
  editEvaluation.value = null // 清理编辑数据
  getTableData()
}

const handleDetailClose = () => {
  detailDialogVisible.value = false
  currentEvaluation.value = {} // 清理详情数据
}

const handleFactorsSuccess = (factorConfig) => {
  factorsDialogVisible.value = false
  currentEvaluation.value = {} // 清理评估数据
  currentEvaluationInfo.value = { evaluationName: '', projectName: '', nesmaRules: 'v2.2' }
  // 刷新列表以显示可能更新的评估状态
  getTableData()
  ElMessage.success('评估因子配置保存成功，可以开始评估了')
}

const handleProgressClose = () => {
  progressDialogVisible.value = false
  currentEvaluation.value = {} // 清理评估数据
  // 刷新列表以显示可能更新的状态
  getTableData()
}

// 分页事件
const handleSizeChange = (size) => {
  pageSize.value = size
  getTableData()
}

const handleCurrentChange = (currentPage) => {
  page.value = currentPage
  getTableData()
}

// 生命周期
onMounted(() => {
  getTableData()
  loadOptions()
})
</script>

<style lang="scss" scoped>
// 使用gin-vue-admin的默认样式
</style>