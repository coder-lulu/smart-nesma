<template>
  <el-dialog
    v-model="visible"
    title="选择智能分析项目"
    width="500px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
  >
    <el-form
      ref="formRef"
      :model="form"
      :rules="rules"
      label-width="120px"
      label-position="left"
    >
      <!-- 项目信息显示 -->
      <el-alert
        :title="`项目：${project.name || ''}`"
        type="info"
        :closable="false"
        show-icon
        style="margin-bottom: 20px"
      >
        <template #default>
          <p style="margin: 5px 0">
            <strong>当前周期：</strong>{{ project.activeCycleName || '未设置' }}
          </p>
          <p style="margin: 5px 0" v-if="project.description">
            <strong>项目描述：</strong>{{ project.description }}
          </p>
        </template>
      </el-alert>

      <!-- 分析选项 -->
      <el-form-item label="分析项目" prop="selectedOptions">
        <el-checkbox-group v-model="form.selectedOptions">
          <el-space direction="vertical" style="width: 100%">
            <el-checkbox 
              label="description" 
              :disabled="false"
            >
              <div style="display: flex; align-items: center">
                <el-icon style="margin-right: 8px"><Document /></el-icon>
                <div>
                  <div style="font-weight: 500">需求描述优化</div>
                  <div style="font-size: 12px; color: #909399; margin-top: 2px">
                    使用AI优化和丰富需求描述，生成详细的功能文档
                  </div>
                </div>
              </div>
            </el-checkbox>

            <el-checkbox 
              label="level4" 
              :disabled="false"
            >
              <div style="display: flex; align-items: center">
                <el-icon style="margin-right: 8px"><Grid /></el-icon>
                <div>
                  <div style="font-weight: 500">L4功能点生成</div>
                  <div style="font-size: 12px; color: #909399; margin-top: 2px">
                    基于L3需求自动生成详细的四级功能点
                  </div>
                </div>
              </div>
            </el-checkbox>

            <el-checkbox 
              label="mermaid" 
              :disabled="false"
            >
              <div style="display: flex; align-items: center">
                <el-icon style="margin-right: 8px"><Share /></el-icon>
                <div>
                  <div style="font-weight: 500">业务流程图生成</div>
                  <div style="font-size: 12px; color: #909399; margin-top: 2px">
                    生成Mermaid格式的业务流程图和功能关系图
                  </div>
                </div>
              </div>
            </el-checkbox>
          </el-space>
        </el-checkbox-group>
      </el-form-item>

      <!-- L4功能点配置 -->
      <el-form-item 
        v-if="form.selectedOptions.includes('level4')" 
        label="L4配置选项"
      >
        <el-space direction="vertical" style="width: 100%">
          <el-checkbox v-model="form.autoSaveL4">
            <div style="display: flex; align-items: center">
              <el-icon style="margin-right: 8px"><Check /></el-icon>
              <div>
                <div style="font-weight: 500">自动保存L4功能点</div>
                <div style="font-size: 12px; color: #909399; margin-top: 2px">
                  生成后自动入库，否则需要手动确认后保存
                </div>
              </div>
            </div>
          </el-checkbox>

          <el-form-item label="生成策略" style="margin-bottom: 0">
            <el-radio-group v-model="form.l4Strategy" size="small">
              <el-radio label="comprehensive">全面分析</el-radio>
              <el-radio label="focused">重点分析</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="最大生成数量" style="margin-bottom: 0">
            <el-input-number 
              v-model="form.maxL4Count" 
              :min="1" 
              :max="20" 
              size="small"
              style="width: 120px"
            />
            <span style="margin-left: 8px; color: #909399; font-size: 12px">
              每个L3需求最多生成的L4功能点数量
            </span>
          </el-form-item>
        </el-space>
      </el-form-item>

      <!-- 流程图配置 -->
      <el-form-item 
        v-if="form.selectedOptions.includes('mermaid')" 
        label="流程图配置"
      >
        <el-space direction="vertical" style="width: 100%">
          <el-form-item label="图表类型" style="margin-bottom: 0">
            <el-radio-group v-model="form.diagramType" size="small">
              <el-radio label="flowchart">流程图</el-radio>
              <el-radio label="mindmap">思维导图</el-radio>
              <el-radio label="sequence">时序图</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="详细级别" style="margin-bottom: 0">
            <el-radio-group v-model="form.detailLevel" size="small">
              <el-radio label="simple">简单</el-radio>
              <el-radio label="detailed">详细</el-radio>
              <el-radio label="comprehensive">全面</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-checkbox v-model="form.autoLayout">
            <span style="font-size: 14px">自动布局优化</span>
          </el-checkbox>
        </el-space>
      </el-form-item>

      <!-- 分析优先级 -->
      <el-form-item label="分析优先级">
        <el-radio-group v-model="form.priority">
          <el-radio :value="1">高优先级</el-radio>
          <el-radio :value="2">普通</el-radio>
          <el-radio :value="3">低优先级</el-radio>
        </el-radio-group>
        <div style="font-size: 12px; color: #909399; margin-top: 4px">
          高优先级任务将优先执行，但可能消耗更多系统资源
        </div>
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleCancel">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleConfirm"
          :loading="confirmLoading"
          :disabled="form.selectedOptions.length === 0"
        >
          开始智能分析
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { 
  Document, 
  Grid, 
  Share, 
  Check 
} from '@element-plus/icons-vue'

// Props定义
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  project: {
    type: Object,
    default: () => ({})
  }
})

// Emits定义
const emit = defineEmits(['update:modelValue', 'confirm', 'cancel'])

// 响应式数据
const formRef = ref()
const confirmLoading = ref(false)

// 表单数据
const form = reactive({
  selectedOptions: [],
  autoSaveL4: true,
  l4Strategy: 'comprehensive',
  maxL4Count: 8,
  diagramType: 'flowchart',
  detailLevel: 'detailed',
  autoLayout: true,
  priority: 2
})

// 表单验证规则
const rules = {
  selectedOptions: [
    { required: true, message: '请至少选择一个分析项目', trigger: 'change' }
  ]
}

// 计算属性
const visible = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value)
})

// 监听L4选项变化，自动选中描述优化
watch(
  () => form.selectedOptions,
  (newOptions) => {
    // 如果选择了L4功能点生成，描述优化是必选项
    if (newOptions.includes('level4') && !newOptions.includes('description')) {
      form.selectedOptions.push('description')
      ElMessage.info('选择L4功能点生成时，需求描述优化为必选项')
    }
  },
  { deep: true }
)

// 确认按钮处理
const handleConfirm = async () => {
  try {
    // 表单验证
    await formRef.value.validate()
    
    if (form.selectedOptions.length === 0) {
      ElMessage.warning('请至少选择一个分析项目')
      return
    }

    confirmLoading.value = true

    // 构建分析配置
    const analysisConfig = {
      type: "project_analysis",
      project_id: props.project.ID,
      cycle_id: props.project.activeCycleId,
      priority: form.priority,
      parameters: {
        selectedOptions: form.selectedOptions,
        includeDescription: form.selectedOptions.includes('description'),
        includeLevel4: form.selectedOptions.includes('level4'),
        includeMermaid: form.selectedOptions.includes('mermaid'),
        
        // L4配置
        autoSaveL4: form.autoSaveL4,
        l4Strategy: form.l4Strategy,
        maxL4Count: form.maxL4Count,
        
        // 流程图配置
        diagramType: form.diagramType,
        detailLevel: form.detailLevel,
        autoLayout: form.autoLayout
      }
    }

    // 发送确认事件
    emit('confirm', analysisConfig)
    
  } catch (error) {
    console.error('表单验证失败:', error)
  } finally {
    confirmLoading.value = false
  }
}

// 取消按钮处理
const handleCancel = () => {
  emit('cancel')
  visible.value = false
}

// 重置表单
const resetForm = () => {
  form.selectedOptions = []
  form.autoSaveL4 = true
  form.l4Strategy = 'comprehensive'
  form.maxL4Count = 8
  form.diagramType = 'flowchart'
  form.detailLevel = 'detailed'
  form.autoLayout = true
  form.priority = 2
  
  if (formRef.value) {
    formRef.value.clearValidate()
  }
}

// 监听对话框显示状态，重置表单
watch(visible, (newVisible) => {
  if (newVisible) {
    resetForm()
  }
})
</script>

<style lang="scss" scoped>
.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}

:deep(.el-checkbox-group) {
  .el-checkbox {
    width: 100%;
    margin-right: 0;
    margin-bottom: 16px;
    
    .el-checkbox__label {
      width: 100%;
      white-space: normal;
      line-height: 1.4;
    }
  }
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-alert) {
  .el-alert__content {
    .el-alert__title {
      font-size: 14px;
      margin-bottom: 8px;
    }
  }
}
</style>