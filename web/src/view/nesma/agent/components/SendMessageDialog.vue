<template>
  <el-dialog
    v-model="dialogVisible"
    :title="dialogTitle"
    width="600px"
    :before-close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="messageForm"
      :rules="messageRules"
      label-width="100px"
    >
      <el-form-item label="消息类型" prop="messageType">
        <el-select 
          v-model="messageForm.messageType" 
          placeholder="选择消息类型"
          style="width: 100%"
        >
          <el-option label="普通消息" value="normal" />
          <el-option label="任务分配" value="task" />
          <el-option label="配置更新" value="config" />
          <el-option label="系统通知" value="system" />
          <el-option label="健康检查" value="health" />
        </el-select>
      </el-form-item>

      <el-form-item label="优先级" prop="priority">
        <el-select 
          v-model="messageForm.priority" 
          placeholder="选择优先级"
          style="width: 100%"
        >
          <el-option label="低" value="low" />
          <el-option label="普通" value="normal" />
          <el-option label="高" value="high" />
          <el-option label="紧急" value="urgent" />
        </el-select>
      </el-form-item>

      <el-form-item label="消息内容" prop="content">
        <el-input
          v-model="messageForm.content"
          type="textarea"
          :rows="6"
          placeholder="请输入消息内容"
          maxlength="1000"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="附加数据" v-if="messageForm.messageType === 'task' || messageForm.messageType === 'config'">
        <el-input
          v-model="messageForm.data"
          type="textarea"
          :rows="4"
          placeholder="请输入JSON格式的附加数据"
        />
        <div class="form-tip">
                      <el-button type="text" size="small" @click="validateJSON">验证JSON格式</el-button>
            <el-button type="text" size="small" @click="formatJSON">格式化JSON</el-button>
        </div>
      </el-form-item>

      <el-form-item label="是否等待回复" v-if="!isBroadcast">
        <el-switch 
          v-model="messageForm.expectReply" 
          active-text="是" 
          inactive-text="否"
        />
      </el-form-item>

      <el-form-item label="超时时间(秒)" v-if="messageForm.expectReply">
        <el-input-number 
          v-model="messageForm.timeout" 
          :min="10" 
          :max="300"
          style="width: 200px"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleClose">取消</el-button>
        <el-button @click="handleReset">重置</el-button>
        <el-button 
          type="primary" 
          @click="handleSendMessage" 
          :loading="sending"
        >
          {{ isBroadcast ? '广播消息' : '发送消息' }}
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { ElMessage } from 'element-plus'
import { sendAgentMessage } from '@/api/nesma'

// Props
const props = defineProps({
  visible: {
    type: Boolean,
    default: false
  },
  agentId: {
    type: String,
    default: ''
  }
})

// Emits
const emit = defineEmits(['update:visible', 'success'])

// 响应式数据
const dialogVisible = computed({
  get: () => props.visible,
  set: (value) => emit('update:visible', value)
})

const formRef = ref(null)
const sending = ref(false)

// 是否为广播消息
const isBroadcast = computed(() => props.agentId === 'broadcast')

// 对话框标题
const dialogTitle = computed(() => {
  return isBroadcast.value ? '广播消息' : '发送消息'
})

// 表单数据
const messageForm = reactive({
  messageType: 'normal',
  priority: 'normal',
  content: '',
  data: '',
  expectReply: false,
  timeout: 30
})

// 表单验证规则
const messageRules = {
  messageType: [
    { required: true, message: '请选择消息类型', trigger: 'change' }
  ],
  priority: [
    { required: true, message: '请选择优先级', trigger: 'change' }
  ],
  content: [
    { required: true, message: '请输入消息内容', trigger: 'blur' },
    { min: 1, max: 1000, message: '消息内容长度在1到1000个字符', trigger: 'blur' }
  ]
}

// 验证JSON格式
const validateJSON = () => {
  if (!messageForm.data) {
    ElMessage.info('附加数据为空')
    return
  }
  
  try {
    JSON.parse(messageForm.data)
    ElMessage.success('JSON格式正确')
  } catch (error) {
    ElMessage.error('JSON格式错误：' + error.message)
  }
}

// 格式化JSON
const formatJSON = () => {
  if (!messageForm.data) {
    ElMessage.info('附加数据为空')
    return
  }
  
  try {
    const parsed = JSON.parse(messageForm.data)
    messageForm.data = JSON.stringify(parsed, null, 2)
    ElMessage.success('JSON格式化成功')
  } catch (error) {
    ElMessage.error('JSON格式错误，无法格式化')
  }
}

// 重置表单
const handleReset = () => {
  if (formRef.value) {
    formRef.value.resetFields()
  }
  Object.assign(messageForm, {
    messageType: 'normal',
    priority: 'normal',
    content: '',
    data: '',
    expectReply: false,
    timeout: 30
  })
}

// 发送消息
const handleSendMessage = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    
    // 验证附加数据
    let additionalData = null
    if (messageForm.data) {
      try {
        additionalData = JSON.parse(messageForm.data)
      } catch (error) {
        ElMessage.error('附加数据格式错误，请检查JSON格式')
        return
      }
    }
    
    sending.value = true
    
    const requestData = {
      agentId: isBroadcast.value ? null : props.agentId,
      messageType: messageForm.messageType,
      priority: messageForm.priority,
      content: messageForm.content,
      data: additionalData,
      expectReply: messageForm.expectReply,
      timeout: messageForm.expectReply ? messageForm.timeout : undefined,
      isBroadcast: isBroadcast.value
    }
    
    await sendAgentMessage(requestData)
    
    ElMessage.success(isBroadcast.value ? '广播消息发送成功' : '消息发送成功')
    emit('success')
    handleClose()
  } catch (error) {
    ElMessage.error('发送失败：' + (error.message || '未知错误'))
  } finally {
    sending.value = false
  }
}

// 关闭对话框
const handleClose = () => {
  dialogVisible.value = false
  handleReset()
}
</script>

<style scoped>
.form-tip {
  margin-top: 8px;
  display: flex;
  gap: 10px;
  font-size: 12px;
  color: #909399;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.el-input-number {
  width: 200px;
}

.el-textarea {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
}
</style> 