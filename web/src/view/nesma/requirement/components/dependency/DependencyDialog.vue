<template>
  <el-dialog
    :model-value="modelValue"
    title="需求依赖关系管理"
    width="800px"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <div class="dependency-dialog">
      <div v-if="requirement" class="requirement-info">
        <h4>当前需求：{{ requirement.title }}</h4>
        <el-descriptions :column="3" border>
          <el-descriptions-item label="层级">L{{ requirement.level }}</el-descriptions-item>
          <el-descriptions-item label="状态">{{ requirement.status }}</el-descriptions-item>
          <el-descriptions-item label="优先级">{{ requirement.priority }}</el-descriptions-item>
        </el-descriptions>
      </div>

      <el-tabs v-model="activeTab" class="dependency-tabs">
        <el-tab-pane label="依赖的需求" name="dependencies">
          <div class="dependency-section">
            <div class="section-header">
              <span>该需求依赖以下需求完成：</span>
              <el-button size="small" type="primary" @click="showAddDependency = true">
                添加依赖
              </el-button>
            </div>
            
            <el-table :data="dependencies" style="width: 100%">
              <el-table-column prop="title" label="需求标题" />
              <el-table-column prop="level" label="层级" width="80">
                <template #default="{ row }">
                  <el-tag size="small" :type="getLevelType(row.level)">
                    L{{ row.level }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag size="small" :type="getStatusType(row.status)">
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="操作" width="100">
                <template #default="{ row }">
                  <el-button size="small" type="danger" @click="removeDependency(row)">
                    移除
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
        
        <el-tab-pane label="被依赖的需求" name="dependents">
          <div class="dependency-section">
            <div class="section-header">
              <span>以下需求依赖该需求完成：</span>
            </div>
            
            <el-table :data="dependents" style="width: 100%">
              <el-table-column prop="title" label="需求标题" />
              <el-table-column prop="level" label="层级" width="80">
                <template #default="{ row }">
                  <el-tag size="small" :type="getLevelType(row.level)">
                    L{{ row.level }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="status" label="状态" width="100">
                <template #default="{ row }">
                  <el-tag size="small" :type="getStatusType(row.status)">
                    {{ row.status }}
                  </el-tag>
                </template>
              </el-table-column>
            </el-table>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 添加依赖对话框 -->
    <el-dialog
      v-model="showAddDependency"
      title="添加依赖需求"
      width="600px"
      append-to-body
    >
      <el-form :model="addForm" label-width="100px">
        <el-form-item label="搜索需求">
          <el-input
            v-model="searchKeyword"
            placeholder="输入需求标题进行搜索"
            @input="handleSearch"
          />
        </el-form-item>
        <el-form-item label="可选需求">
          <el-table
            :data="availableRequirements"
            @selection-change="handleSelectionChange"
            style="width: 100%"
            max-height="300px"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="title" label="需求标题" />
            <el-table-column prop="level" label="层级" width="80">
              <template #default="{ row }">
                <el-tag size="small" :type="getLevelType(row.level)">
                  L{{ row.level }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="showAddDependency = false">取消</el-button>
        <el-button type="primary" @click="handleAddDependencies">确定</el-button>
      </template>
    </el-dialog>

    <template #footer>
      <el-button @click="$emit('update:modelValue', false)">关闭</el-button>
      <el-button type="primary" @click="handleSave">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  requirement: {
    type: Object,
    default: null
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'dependencies-updated'])

// 响应式数据
const activeTab = ref('dependencies')
const showAddDependency = ref(false)
const searchKeyword = ref('')

const dependencies = ref([
  // 模拟数据
  { id: 1, title: '用户认证模块', level: 3, status: 'completed' },
  { id: 2, title: '权限管理系统', level: 2, status: 'in_progress' }
])

const dependents = ref([
  // 模拟数据
  { id: 3, title: '订单处理流程', level: 4, status: 'pending' },
  { id: 4, title: '支付接口集成', level: 3, status: 'pending' }
])

const availableRequirements = ref([
  // 模拟数据
  { id: 5, title: '数据库设计', level: 3, status: 'completed' },
  { id: 6, title: '接口文档', level: 4, status: 'in_progress' },
  { id: 7, title: '测试用例', level: 4, status: 'pending' }
])

const addForm = reactive({
  selectedRequirements: []
})

// 方法
const getLevelType = (level) => {
  const typeMap = {
    1: 'danger',
    2: 'warning',
    3: 'primary', 
    4: 'success'
  }
  return typeMap[level] || 'info'
}

const getStatusType = (status) => {
  const typeMap = {
    'pending': 'info',
    'in_progress': 'warning',
    'completed': 'success',
    'cancelled': 'danger'
  }
  return typeMap[status] || 'info'
}

const handleSearch = () => {
  // 实现搜索逻辑
  console.log('搜索需求:', searchKeyword.value)
}

const handleSelectionChange = (selection) => {
  addForm.selectedRequirements = selection
}

const handleAddDependencies = () => {
  if (addForm.selectedRequirements.length === 0) {
    ElMessage.warning('请选择要添加的依赖需求')
    return
  }
  
  // 添加到依赖列表
  dependencies.value.push(...addForm.selectedRequirements)
  
  // 重置表单
  addForm.selectedRequirements = []
  showAddDependency.value = false
  
  ElMessage.success('依赖关系添加成功')
}

const removeDependency = (dependency) => {
  const index = dependencies.value.findIndex(d => d.id === dependency.id)
  if (index > -1) {
    dependencies.value.splice(index, 1)
    ElMessage.success('依赖关系已移除')
  }
}

const handleSave = () => {
  emit('dependencies-updated')
  emit('update:modelValue', false)
  ElMessage.success('依赖关系保存成功')
}

// 监听器
watch(() => props.requirement, (newRequirement) => {
  if (newRequirement) {
    // 根据需求ID加载依赖关系数据
    console.log('加载需求依赖关系:', newRequirement.id)
  }
})
</script>

<style lang="scss" scoped>
.dependency-dialog {
  .requirement-info {
    margin-bottom: 20px;
    
    h4 {
      margin-bottom: 12px;
      color: #303133;
    }
  }
  
  .dependency-tabs {
    .dependency-section {
      .section-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
        
        span {
          font-weight: 500;
          color: #606266;
        }
      }
    }
  }
}
</style>