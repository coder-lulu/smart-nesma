<template>
  <div class="message-input">
    <div class="input-container">
      <!-- 输入区域 -->
      <div class="input-wrapper">
        <el-input
          ref="inputRef"
          :model-value="modelValue"
          type="textarea"
          :rows="inputRows"
          :placeholder="placeholder"
          :disabled="disabled"
          :maxlength="maxLength"
          show-word-limit
          resize="none"
          @input="handleInput"
          @keydown="handleKeydown"
          @focus="handleFocus"
          @blur="handleBlur"
          @paste="handlePaste"
          class="message-textarea"
        />
        
        <!-- 输入工具栏 -->
        <div class="input-toolbar">
          <div class="toolbar-left">
            <!-- 附件上传 -->
            <el-upload
              v-if="enableFileUpload"
              :show-file-list="false"
              :before-upload="handleFileUpload"
              accept=".txt,.doc,.docx,.pdf,.jpg,.jpeg,.png,.gif"
              style="display: inline-block;"
            >
              <el-button text size="small" :icon="Paperclip" title="上传文件">
                文件
              </el-button>
            </el-upload>
            
            <!-- 图片上传 -->
            <el-upload
              v-if="enableImageUpload"
              :show-file-list="false"
              :before-upload="handleImageUpload"
              accept=".jpg,.jpeg,.png,.gif,.webp"
              style="display: inline-block;"
            >
              <el-button text size="small" :icon="Picture" title="上传图片">
                图片
              </el-button>
            </el-upload>
            
            <!-- 语音输入 -->
            <el-button 
              v-if="enableVoiceInput"
              text 
              size="small" 
              :icon="isRecording ? Microphone : Microphone"
              :type="isRecording ? 'danger' : 'default'"
              @click="toggleVoiceInput"
              :title="isRecording ? '停止录音' : '语音输入'"
            >
              {{ isRecording ? '录音中' : '语音' }}
            </el-button>
            
            <!-- 表情选择 -->
            <el-popover
              v-if="enableEmoji"
              placement="top-start"
              :width="320"
              trigger="click"
            >
              <template #reference>
                <el-button text size="small" :icon="Sunny" title="表情">
                  表情
                </el-button>
              </template>
              
              <div class="emoji-picker">
                <div class="emoji-categories">
                  <el-button
                    v-for="category in emojiCategories"
                    :key="category.name"
                    text
                    size="small"
                    @click="currentEmojiCategory = category.name"
                    :type="currentEmojiCategory === category.name ? 'primary' : 'default'"
                  >
                    {{ category.icon }}
                  </el-button>
                </div>
                <div class="emoji-list">
                  <span
                    v-for="emoji in currentEmojis"
                    :key="emoji"
                    class="emoji-item"
                    @click="insertEmoji(emoji)"
                  >
                    {{ emoji }}
                  </span>
                </div>
              </div>
            </el-popover>
            
            <!-- Markdown帮助 -->
            <el-popover
              v-if="enableMarkdown"
              placement="top-start"
              :width="400"
              trigger="hover"
            >
              <template #reference>
                <el-button text size="small" :icon="QuestionFilled" title="Markdown帮助">
                  MD
                </el-button>
              </template>
              
              <div class="markdown-help">
                <h4>Markdown语法帮助</h4>
                <div class="help-item">
                  <code>**粗体**</code> → <strong>粗体</strong>
                </div>
                <div class="help-item">
                  <code>*斜体*</code> → <em>斜体</em>
                </div>
                <div class="help-item">
                  <code>`代码`</code> → <code>代码</code>
                </div>
                <div class="help-item">
                  <code>```代码块```</code> → 代码块
                </div>
                <div class="help-item">
                  <code>[链接](url)</code> → 链接
                </div>
              </div>
            </el-popover>
          </div>
          
          <div class="toolbar-right">
            <!-- 字数统计 -->
            <span class="char-count" :class="{ warning: isNearLimit }">
              {{ characterCount }}/{{ maxLength }}
            </span>
            
            <!-- 发送按钮 -->
            <el-button
              type="primary"
              size="default"
              :loading="sending"
              :disabled="!canSend"
              @click="handleSend"
              :icon="Promotion"
            >
              {{ sendButtonText }}
            </el-button>
          </div>
        </div>
        
        <!-- 快捷操作 -->
        <div v-if="quickActions.length > 0" class="quick-actions">
          <span class="quick-label">快捷操作：</span>
          <el-button
            v-for="action in quickActions"
            :key="action.key"
            size="small"
            text
            @click="handleQuickAction(action)"
          >
            {{ action.label }}
          </el-button>
        </div>
        
        <!-- 上传文件预览 -->
        <div v-if="uploadedFiles.length > 0" class="uploaded-files">
          <div class="files-header">
            <span>已上传文件：</span>
            <el-button text size="small" @click="clearFiles">清空</el-button>
          </div>
          <div class="files-list">
            <div 
              v-for="file in uploadedFiles" 
              :key="file.id"
              class="file-item"
            >
              <el-icon><Document /></el-icon>
              <span class="file-name">{{ file.name }}</span>
              <span class="file-size">{{ formatFileSize(file.size) }}</span>
              <el-button 
                text 
                size="small" 
                @click="removeFile(file.id)"
                :icon="Close"
              />
            </div>
          </div>
        </div>
        
        <!-- 输入提示 -->
        <div v-if="showHints" class="input-hints">
          <div class="hint-item">
            <el-icon><Sunny /></el-icon>
            <span>输入 "/" 查看命令提示</span>
          </div>
          <div class="hint-item">
            <el-icon><Document /></el-icon>
            <span>{{ isMac ? 'Cmd' : 'Ctrl' }} + Enter 发送消息</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, nextTick, onMounted, onUnmounted } from 'vue'
import { 
  Paperclip, Picture, Microphone, Sunny, 
  QuestionFilled, Promotion, Document, Close 
} from '@element-plus/icons-vue'

// Props
const props = defineProps({
  modelValue: {
    type: String,
    default: ''
  },
  placeholder: {
    type: String,
    default: '输入您的问题... (Ctrl+Enter发送)'
  },
  disabled: {
    type: Boolean,
    default: false
  },
  sending: {
    type: Boolean,
    default: false
  },
  maxLength: {
    type: Number,
    default: 10000
  },
  enableFileUpload: {
    type: Boolean,
    default: true
  },
  enableImageUpload: {
    type: Boolean,
    default: true
  },
  enableVoiceInput: {
    type: Boolean,
    default: true
  },
  enableEmoji: {
    type: Boolean,
    default: true
  },
  enableMarkdown: {
    type: Boolean,
    default: true
  },
  quickActions: {
    type: Array,
    default: () => [
      { key: 'clear', label: '清空输入' },
      { key: 'template1', label: '分析需求模板' },
      { key: 'template2', label: '功能点计算模板' }
    ]
  },
  autoResize: {
    type: Boolean,
    default: true
  }
})

// Emits
const emit = defineEmits([
  'update:modelValue',
  'send',
  'file-upload',
  'image-upload',
  'voice-input',
  'quick-action',
  'input-focus',
  'input-blur'
])

// 响应式数据
const inputRef = ref()
const isRecording = ref(false)
const isFocused = ref(false)
const uploadedFiles = ref([])
const currentEmojiCategory = ref('smileys')
const inputRows = ref(2)

// 表情数据
const emojiCategories = [
  { name: 'smileys', icon: '😀', emojis: ['😀', '😃', '😄', '😁', '😅', '😂', '🤣', '😊', '😇', '🙂', '😉', '😌', '😍', '🥰'] },
  { name: 'gestures', icon: '👍', emojis: ['👍', '👎', '👌', '✌️', '🤞', '🤟', '🤘', '👏', '🙌', '👐', '🤲', '🤝', '🙏'] },
  { name: 'objects', icon: '💼', emojis: ['💻', '📱', '⌨️', '🖥️', '📄', '📊', '📈', '📉', '📝', '📋', '📌', '📎', '🔗'] }
]

// 计算属性
const characterCount = computed(() => props.modelValue.length)
const isNearLimit = computed(() => characterCount.value > props.maxLength * 0.8)
const canSend = computed(() => {
  return props.modelValue.trim().length > 0 && !props.disabled && !props.sending
})
const sendButtonText = computed(() => {
  if (props.sending) return '发送中...'
  return '发送'
})
const currentEmojis = computed(() => {
  const category = emojiCategories.find(cat => cat.name === currentEmojiCategory.value)
  return category ? category.emojis : []
})
const showHints = computed(() => {
  return !isFocused.value && props.modelValue.length === 0
})
const isMac = computed(() => {
  return navigator.platform.toUpperCase().indexOf('MAC') >= 0
})

// 方法
const handleInput = (value) => {
  emit('update:modelValue', value)
  
  if (props.autoResize) {
    nextTick(() => {
      adjustTextareaHeight()
    })
  }
}

const handleKeydown = (event) => {
  const isCtrlOrCmd = event.ctrlKey || event.metaKey
  
  if (isCtrlOrCmd && event.key === 'Enter') {
    event.preventDefault()
    handleSend()
  } else if (event.key === 'Tab') {
    event.preventDefault()
    insertText('    ') // 插入4个空格
  } else if (event.key === '/' && props.modelValue.length === 0) {
    // 显示命令提示
    showCommandHints()
  }
}

const handleFocus = () => {
  isFocused.value = true
  emit('input-focus')
}

const handleBlur = () => {
  isFocused.value = false
  emit('input-blur')
}

const handlePaste = (event) => {
  const items = event.clipboardData?.items
  if (!items) return
  
  for (const item of items) {
    if (item.type.indexOf('image') !== -1) {
      const file = item.getAsFile()
      if (file) {
        handleImageUpload(file)
        event.preventDefault()
      }
    }
  }
}

const handleSend = () => {
  if (canSend.value) {
    emit('send', {
      content: props.modelValue,
      files: uploadedFiles.value
    })
    clearFiles()
  }
}

const handleFileUpload = (file) => {
  if (file.size > 10 * 1024 * 1024) { // 10MB限制
    ElMessage.error('文件大小不能超过10MB')
    return false
  }
  
  const fileData = {
    id: Date.now(),
    name: file.name,
    size: file.size,
    type: file.type,
    file: file
  }
  
  uploadedFiles.value.push(fileData)
  emit('file-upload', fileData)
  return false // 阻止自动上传
}

const handleImageUpload = (file) => {
  if (file.size > 5 * 1024 * 1024) { // 5MB限制
    ElMessage.error('图片大小不能超过5MB')
    return false
  }
  
  const imageData = {
    id: Date.now(),
    name: file.name,
    size: file.size,
    type: file.type,
    file: file
  }
  
  uploadedFiles.value.push(imageData)
  emit('image-upload', imageData)
  return false
}

const toggleVoiceInput = () => {
  isRecording.value = !isRecording.value
  emit('voice-input', isRecording.value)
}

const insertEmoji = (emoji) => {
  insertText(emoji)
}

const insertText = (text) => {
  const textarea = inputRef.value?.textarea
  if (!textarea) return
  
  const start = textarea.selectionStart
  const end = textarea.selectionEnd
  const before = props.modelValue.substring(0, start)
  const after = props.modelValue.substring(end)
  
  const newValue = before + text + after
  emit('update:modelValue', newValue)
  
  nextTick(() => {
    const newPosition = start + text.length
    textarea.setSelectionRange(newPosition, newPosition)
    textarea.focus()
  })
}

const handleQuickAction = (action) => {
  switch (action.key) {
    case 'clear':
      emit('update:modelValue', '')
      break
    case 'template1':
      insertText('请分析以下需求的功能点：\n\n需求描述：\n\n预期结果：\n- 功能类型分类\n- 复杂度评估\n- AFP/UFP计算')
      break
    case 'template2':
      insertText('请帮我计算这个功能的功能点：\n\n功能描述：\n数据元素类型：\n文件类型引用：\n\n请按照NESMA标准进行计算')
      break
    default:
      emit('quick-action', action)
  }
}

const removeFile = (fileId) => {
  uploadedFiles.value = uploadedFiles.value.filter(file => file.id !== fileId)
}

const clearFiles = () => {
  uploadedFiles.value = []
}

const formatFileSize = (bytes) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const adjustTextareaHeight = () => {
  const textarea = inputRef.value?.textarea
  if (!textarea) return
  
  const lines = props.modelValue.split('\n').length
  const newRows = Math.min(Math.max(lines, 2), 8)
  
  if (newRows !== inputRows.value) {
    inputRows.value = newRows
  }
}

const showCommandHints = () => {
  // 这里可以实现命令提示功能
  console.log('显示命令提示')
}

// 监听器
watch(() => props.modelValue, () => {
  if (props.autoResize) {
    nextTick(() => {
      adjustTextareaHeight()
    })
  }
})

// 生命周期
onMounted(() => {
  // 初始化
})

onUnmounted(() => {
  // 清理
})
</script>

<style lang="scss" scoped>
.message-input {
  border-top: 1px solid #e5e7eb;
  background: linear-gradient(to bottom, #ffffff, #f8fafc);

  .input-container {
    padding: 16px 20px;

    .input-wrapper {
      .message-textarea {
        :deep(.el-textarea__inner) {
          border: 2px solid #e5e7eb;
          border-radius: 12px;
          padding: 16px 20px;
          font-size: 14px;
          line-height: 1.6;
          resize: none;
          transition: all 0.3s ease;
          background: white;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.05);

          &:focus {
            border-color: #3b82f6;
            box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
            background: #fff;
          }

          &:disabled {
            background: #f5f5f5;
            color: #999;
          }
        }

        :deep(.el-input__count) {
          bottom: 8px;
          right: 12px;
          background: rgba(255, 255, 255, 0.9);
          padding: 2px 6px;
          border-radius: 4px;
          font-size: 11px;
        }
      }

      .input-toolbar {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 12px;

        .toolbar-left {
          display: flex;
          align-items: center;
          gap: 8px;

          .el-button {
            color: #6b7280;
            transition: all 0.2s ease;

            &:hover {
              color: #3b82f6;
              background: rgba(59, 130, 246, 0.1);
            }

            &.is-type-danger {
              color: #ef4444;
              
              &:hover {
                background: rgba(239, 68, 68, 0.1);
              }
            }
          }
        }

        .toolbar-right {
          display: flex;
          align-items: center;
          gap: 12px;

          .char-count {
            font-size: 12px;
            color: #6b7280;
            font-weight: 500;

            &.warning {
              color: #f59e0b;
            }
          }
        }
      }

      .quick-actions {
        margin-top: 8px;
        padding: 8px 0;
        border-top: 1px solid #f3f4f6;
        display: flex;
        align-items: center;
        flex-wrap: wrap;
        gap: 8px;

        .quick-label {
          font-size: 12px;
          color: #6b7280;
          margin-right: 4px;
        }

        .el-button {
          font-size: 12px;
          padding: 4px 8px;
          color: #6b7280;

          &:hover {
            color: #3b82f6;
            background: rgba(59, 130, 246, 0.1);
          }
        }
      }

      .uploaded-files {
        margin-top: 12px;
        padding: 12px;
        background: #f8fafc;
        border-radius: 8px;
        border: 1px solid #e5e7eb;

        .files-header {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-bottom: 8px;
          font-size: 12px;
          font-weight: 500;
          color: #374151;
        }

        .files-list {
          .file-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 6px 8px;
            background: white;
            border-radius: 6px;
            margin-bottom: 4px;
            font-size: 12px;

            &:last-child {
              margin-bottom: 0;
            }

            .file-name {
              flex: 1;
              color: #374151;
              font-weight: 500;
            }

            .file-size {
              color: #6b7280;
              font-size: 11px;
            }
          }
        }
      }

      .input-hints {
        margin-top: 8px;
        padding: 8px 12px;
        background: rgba(59, 130, 246, 0.05);
        border-radius: 6px;
        border: 1px solid rgba(59, 130, 246, 0.1);

        .hint-item {
          display: flex;
          align-items: center;
          gap: 6px;
          font-size: 11px;
          color: #6b7280;
          margin-bottom: 4px;

          &:last-child {
            margin-bottom: 0;
          }

          .el-icon {
            color: #3b82f6;
          }
        }
      }
    }
  }
}

// 表情选择器样式
.emoji-picker {
  .emoji-categories {
    display: flex;
    gap: 4px;
    margin-bottom: 12px;
    padding-bottom: 8px;
    border-bottom: 1px solid #e5e7eb;

    .el-button {
      font-size: 16px;
      padding: 4px;
      min-width: 32px;
    }
  }

  .emoji-list {
    display: grid;
    grid-template-columns: repeat(8, 1fr);
    gap: 4px;
    max-height: 200px;
    overflow-y: auto;

    .emoji-item {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 32px;
      height: 32px;
      font-size: 18px;
      cursor: pointer;
      border-radius: 4px;
      transition: background 0.2s ease;

      &:hover {
        background: #f3f4f6;
      }
    }
  }
}

// Markdown帮助样式
.markdown-help {
  h4 {
    margin: 0 0 12px 0;
    font-size: 14px;
    color: #374151;
  }

  .help-item {
    display: flex;
    align-items: center;
    gap: 8px;
    margin-bottom: 6px;
    font-size: 12px;

    &:last-child {
      margin-bottom: 0;
    }

    code {
      background: #f3f4f6;
      padding: 2px 4px;
      border-radius: 3px;
      font-family: monospace;
      font-size: 11px;
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .message-input {
    .input-container {
      padding: 12px 16px;

      .input-wrapper {
        .input-toolbar {
          flex-direction: column;
          gap: 12px;

          .toolbar-left {
            order: 2;
            justify-content: center;
            flex-wrap: wrap;
          }

          .toolbar-right {
            order: 1;
            width: 100%;
            justify-content: space-between;
          }
        }

        .quick-actions {
          justify-content: center;
          
          .quick-label {
            width: 100%;
            text-align: center;
            margin-bottom: 4px;
          }
        }
      }
    }
  }

  .emoji-picker {
    .emoji-list {
      grid-template-columns: repeat(6, 1fr);
    }
  }
}
</style>