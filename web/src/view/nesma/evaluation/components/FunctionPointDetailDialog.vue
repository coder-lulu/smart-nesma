<template>
  <el-dialog
    :model-value="modelValue"
    @update:model-value="$emit('update:modelValue', $event)"
    title="功能点详情"
    width="70%"
    :close-on-click-modal="false"
    destroy-on-close
  >
    <div v-if="functionPoint" class="function-point-detail">
      <!-- 基本信息 -->
      <el-card class="detail-section">
        <template #header>
          <div class="section-header">
            <span>基本信息</span>
            <el-tag :type="getFunctionTypeColor(functionPoint.functionType)">
              {{ getFunctionTypeLabel(functionPoint.functionType) }}
            </el-tag>
          </div>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="12">
            <div class="detail-item">
              <span class="detail-label">功能名称：</span>
              <span class="detail-value">{{ functionPoint.functionName }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">功能类型：</span>
              <span class="detail-value">{{ getFunctionTypeLabel(functionPoint.functionType) }}</span>
            </div>
            <div class="detail-item">
              <span class="detail-label">复杂度等级：</span>
              <el-tag :type="getComplexityColor(functionPoint.complexityLevel)" size="small">
                {{ getComplexityLabel(functionPoint.complexityLevel) }}
              </el-tag>
            </div>
            <div class="detail-item">
              <span class="detail-label">功能点值：</span>
              <span class="detail-value fp-value">{{ functionPoint.functionPointValue?.toFixed(1) || 0 }}</span>
            </div>
          </el-col>
          <el-col :span="12">
            <div class="detail-item">
              <span class="detail-label">识别方法：</span>
              <el-tag :type="getIdentificationMethodColor(functionPoint.identificationMethod)" size="small">
                {{ getIdentificationMethodLabel(functionPoint.identificationMethod) }}
              </el-tag>
            </div>
            <div class="detail-item">
              <span class="detail-label">验证状态：</span>
              <el-tag :type="getValidationStatusColor(functionPoint.validationStatus)" size="small">
                {{ getValidationStatusLabel(functionPoint.validationStatus) }}
              </el-tag>
            </div>
            <div class="detail-item">
              <span class="detail-label">置信度：</span>
              <div class="confidence-display">
                <el-progress
                  :percentage="(functionPoint.confidenceLevel * 100) || 0"
                  :stroke-width="8"
                  :show-text="false"
                  :color="getConfidenceColor(functionPoint.confidenceLevel)"
                />
                <span class="confidence-text">{{ (functionPoint.confidenceLevel * 100)?.toFixed(1) || 0 }}%</span>
              </div>
            </div>
            <div class="detail-item">
              <span class="detail-label">创建时间：</span>
              <span class="detail-value">{{ formatDateTime(functionPoint.createdAt) }}</span>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 功能描述 -->
      <el-card class="detail-section">
        <template #header>
          <span>功能描述</span>
        </template>
        
        <div class="description-content">
          <p>{{ functionPoint.description || '暂无描述' }}</p>
        </div>
        
        <div v-if="functionPoint.businessRules" class="business-rules">
          <h4>业务规则</h4>
          <p>{{ functionPoint.businessRules }}</p>
        </div>
      </el-card>

      <!-- 技术细节 -->
      <el-card class="detail-section">
        <template #header>
          <span>技术细节</span>
        </template>
        
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="tech-detail-card">
              <h4>数据元素类型 (DET)</h4>
              <div class="tech-value">{{ functionPoint.dataElementTypes || 0 }}</div>
              <p class="tech-description">用户可识别的、不重复的数据字段</p>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="tech-detail-card">
              <h4>文件类型引用 (FTR)</h4>
              <div class="tech-value">{{ functionPoint.fileTypeReferences || 0 }}</div>
              <p class="tech-description">功能读取或维护的文件类型</p>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="tech-detail-card">
              <h4>记录元素类型 (RET)</h4>
              <div class="tech-value">{{ functionPoint.recordElementTypes || 0 }}</div>
              <p class="tech-description">用户可识别的数据子组</p>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 计算详情 -->
      <el-card class="detail-section">
        <template #header>
          <span>计算详情</span>
        </template>
        
        <div class="calculation-details">
          <div class="calculation-step">
            <h4>复杂度确定</h4>
            <p>基于 DET: {{ functionPoint.dataElementTypes || 0 }}，FTR: {{ functionPoint.fileTypeReferences || 0 }}，RET: {{ functionPoint.recordElementTypes || 0 }}</p>
            <p>复杂度等级：<el-tag :type="getComplexityColor(functionPoint.complexityLevel)" size="small">{{ getComplexityLabel(functionPoint.complexityLevel) }}</el-tag></p>
          </div>
          
          <div class="calculation-step">
            <h4>功能点计算</h4>
            <p>{{ getFunctionTypeLabel(functionPoint.functionType) }} - {{ getComplexityLabel(functionPoint.complexityLevel) }} = {{ functionPoint.functionPointValue?.toFixed(1) || 0 }} 功能点</p>
          </div>
          
          <div class="calculation-step" v-if="functionPoint.adjustmentFactors">
            <h4>调整因子</h4>
            <p>{{ functionPoint.adjustmentFactors }}</p>
          </div>
        </div>
      </el-card>

      <!-- 验证信息 -->
      <el-card class="detail-section" v-if="functionPoint.validationNotes || functionPoint.reviewComments">
        <template #header>
          <span>验证信息</span>
        </template>
        
        <div class="validation-info">
          <div v-if="functionPoint.validationNotes" class="validation-item">
            <h4>验证备注</h4>
            <p>{{ functionPoint.validationNotes }}</p>
          </div>
          
          <div v-if="functionPoint.reviewComments" class="validation-item">
            <h4>评审意见</h4>
            <p>{{ functionPoint.reviewComments }}</p>
          </div>
          
          <div v-if="functionPoint.lastReviewedBy || functionPoint.lastReviewedAt" class="validation-item">
            <h4>最后评审</h4>
            <p>
              评审人：{{ functionPoint.lastReviewedBy || '未知' }}
              <span v-if="functionPoint.lastReviewedAt">
                ，时间：{{ formatDateTime(functionPoint.lastReviewedAt) }}
              </span>
            </p>
          </div>
        </div>
      </el-card>

      <!-- 相关文档 -->
      <el-card class="detail-section" v-if="functionPoint.attachments && functionPoint.attachments.length > 0">
        <template #header>
          <span>相关文档</span>
        </template>
        
        <div class="attachments-list">
          <div 
            v-for="attachment in functionPoint.attachments" 
            :key="attachment.id"
            class="attachment-item"
          >
            <el-icon><Document /></el-icon>
            <span class="attachment-name">{{ attachment.fileName }}</span>
            <span class="attachment-size">{{ formatFileSize(attachment.fileSize) }}</span>
            <el-button size="small" @click="downloadAttachment(attachment)">下载</el-button>
          </div>
        </div>
      </el-card>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="$emit('update:modelValue', false)">关闭</el-button>
        <el-button type="primary" @click="editFunctionPoint">编辑</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { Document } from '@element-plus/icons-vue'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  functionPoint: {
    type: Object,
    default: () => ({})
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'edit'])

// 方法
const getFunctionTypeLabel = (type) => {
  const labels = {
    'ILF': '内部逻辑文件',
    'EIF': '外部接口文件',
    'EI': '外部输入',
    'EO': '外部输出',
    'EQ': '外部查询'
  }
  return labels[type] || type
}

const getFunctionTypeColor = (type) => {
  const colors = {
    'ILF': 'primary',
    'EIF': 'success',
    'EI': 'warning',
    'EO': 'danger',
    'EQ': 'info'
  }
  return colors[type] || ''
}

const getComplexityLabel = (level) => {
  const labels = {
    'Low': '低',
    'Average': '中',
    'High': '高'
  }
  return labels[level] || level
}

const getComplexityColor = (level) => {
  const colors = {
    'Low': 'success',
    'Average': 'warning',
    'High': 'danger'
  }
  return colors[level] || ''
}

const getIdentificationMethodLabel = (method) => {
  const labels = {
    'manual': '手工识别',
    'automated': '自动识别',
    'hybrid': '混合识别'
  }
  return labels[method] || method
}

const getIdentificationMethodColor = (method) => {
  const colors = {
    'manual': 'warning',
    'automated': 'success',
    'hybrid': 'primary'
  }
  return colors[method] || ''
}

const getValidationStatusLabel = (status) => {
  const labels = {
    'pending': '待验证',
    'approved': '已通过',
    'rejected': '已拒绝',
    'reviewed': '已评审'
  }
  return labels[status] || status
}

const getValidationStatusColor = (status) => {
  const colors = {
    'pending': 'warning',
    'approved': 'success',
    'rejected': 'danger',
    'reviewed': 'primary'
  }
  return colors[status] || ''
}

const getConfidenceColor = (level) => {
  if (level >= 0.8) return '#67c23a'
  if (level >= 0.6) return '#e6a23c'
  return '#f56c6c'
}

const formatDateTime = (dateTime) => {
  if (!dateTime) return '-'
  return new Date(dateTime).toLocaleString('zh-CN')
}

const formatFileSize = (size) => {
  if (!size) return '-'
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  return (size / (1024 * 1024)).toFixed(1) + ' MB'
}

const editFunctionPoint = () => {
  emit('edit', props.functionPoint)
  emit('update:modelValue', false)
}

const downloadAttachment = (attachment) => {
  // 实现文档下载逻辑
  console.log('下载文档:', attachment)
}
</script>

<style scoped>
.function-point-detail {
  max-height: 70vh;
  overflow-y: auto;
}

.detail-section {
  margin-bottom: 20px;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.detail-item {
  display: flex;
  margin-bottom: 12px;
  min-height: 24px;
  align-items: center;
}

.detail-label {
  width: 120px;
  color: #909399;
  font-size: 14px;
}

.detail-value {
  flex: 1;
  color: #303133;
  font-size: 14px;
}

.fp-value {
  font-weight: bold;
  color: #409eff;
  font-size: 16px;
}

.confidence-display {
  display: flex;
  align-items: center;
  gap: 10px;
  flex: 1;
}

.confidence-text {
  font-size: 14px;
  font-weight: bold;
  color: #303133;
  min-width: 40px;
}

.description-content {
  color: #606266;
  line-height: 1.6;
  margin-bottom: 15px;
}

.business-rules {
  border-top: 1px solid #e4e7ed;
  padding-top: 15px;
}

.business-rules h4 {
  color: #303133;
  font-size: 14px;
  margin-bottom: 8px;
}

.business-rules p {
  color: #606266;
  line-height: 1.6;
  margin: 0;
}

.tech-detail-card {
  text-align: center;
  padding: 20px;
  background: #f8f9fa;
  border-radius: 8px;
  border: 1px solid #e4e7ed;
}

.tech-detail-card h4 {
  color: #303133;
  font-size: 14px;
  margin-bottom: 10px;
}

.tech-value {
  font-size: 32px;
  font-weight: bold;
  color: #409eff;
  margin-bottom: 8px;
}

.tech-description {
  color: #909399;
  font-size: 12px;
  margin: 0;
  line-height: 1.4;
}

.calculation-details {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.calculation-step {
  padding: 15px;
  background: #f8f9fa;
  border-radius: 8px;
  border-left: 4px solid #409eff;
}

.calculation-step h4 {
  color: #303133;
  font-size: 14px;
  margin-bottom: 8px;
}

.calculation-step p {
  color: #606266;
  font-size: 14px;
  margin: 5px 0;
  line-height: 1.5;
}

.validation-info {
  display: flex;
  flex-direction: column;
  gap: 15px;
}

.validation-item h4 {
  color: #303133;
  font-size: 14px;
  margin-bottom: 8px;
}

.validation-item p {
  color: #606266;
  font-size: 14px;
  margin: 0;
  line-height: 1.6;
}

.attachments-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.attachment-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px;
  background: #f8f9fa;
  border-radius: 4px;
  border: 1px solid #e4e7ed;
}

.attachment-name {
  flex: 1;
  color: #303133;
  font-size: 14px;
}

.attachment-size {
  color: #909399;
  font-size: 12px;
  min-width: 60px;
}

.dialog-footer {
  text-align: right;
}
</style> 