<template>
  <div class="style-config">
    <el-form :model="styleConfig" label-width="100px" :disabled="disabled">
      <el-form-item label="主题色">
        <el-color-picker v-model="styleConfig.primaryColor" />
      </el-form-item>
      
      <el-form-item label="字体大小">
        <el-slider
          v-model="styleConfig.fontSize"
          :min="10"
          :max="24"
          :marks="{ 10: '10px', 14: '14px', 18: '18px', 24: '24px' }"
          show-stops
        />
      </el-form-item>
      
      <el-form-item label="字体族">
        <el-select v-model="styleConfig.fontFamily" placeholder="选择字体">
          <el-option label="PingFang SC" value="PingFang SC" />
          <el-option label="Microsoft YaHei" value="Microsoft YaHei" />
          <el-option label="SimHei" value="SimHei" />
          <el-option label="Arial" value="Arial" />
          <el-option label="Times New Roman" value="Times New Roman" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="行间距">
        <el-slider
          v-model="styleConfig.lineHeight"
          :min="1"
          :max="3"
          :step="0.1"
          :marks="{ 1: '1.0', 1.5: '1.5', 2: '2.0', 3: '3.0' }"
          show-stops
        />
      </el-form-item>
      
      <el-form-item label="标题样式">
        <el-form :model="styleConfig.headingStyle" label-width="80px">
          <el-form-item label="颜色">
            <el-color-picker v-model="styleConfig.headingStyle.color" />
          </el-form-item>
          <el-form-item label="字重">
            <el-select v-model="styleConfig.headingStyle.fontWeight">
              <el-option label="正常" value="normal" />
              <el-option label="粗体" value="bold" />
              <el-option label="特粗" value="bolder" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="表格样式">
        <el-form :model="styleConfig.tableStyle" label-width="80px">
          <el-form-item label="边框">
            <el-switch v-model="styleConfig.tableStyle.border" />
          </el-form-item>
          <el-form-item label="斑马纹">
            <el-switch v-model="styleConfig.tableStyle.striped" />
          </el-form-item>
          <el-form-item label="悬停">
            <el-switch v-model="styleConfig.tableStyle.hover" />
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="图表样式">
        <el-form :model="styleConfig.chartStyle" label-width="80px">
          <el-form-item label="主题">
            <el-select v-model="styleConfig.chartStyle.theme">
              <el-option label="默认" value="default" />
              <el-option label="暗色" value="dark" />
              <el-option label="商务" value="business" />
              <el-option label="学术" value="academic" />
            </el-select>
          </el-form-item>
          <el-form-item label="动画">
            <el-switch v-model="styleConfig.chartStyle.animation" />
          </el-form-item>
        </el-form>
      </el-form-item>
    </el-form>
  </div>
</template>

<script setup>
import { reactive, watch } from 'vue'

const props = defineProps({
  modelValue: {
    type: Object,
    default: () => ({
      primaryColor: '#409EFF',
      fontSize: 14,
      fontFamily: 'PingFang SC',
      lineHeight: 1.6,
      headingStyle: {
        color: '#303133',
        fontWeight: 'bold'
      },
      tableStyle: {
        border: true,
        striped: true,
        hover: true
      },
      chartStyle: {
        theme: 'default',
        animation: true
      }
    })
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const styleConfig = reactive({ ...props.modelValue })

watch(styleConfig, (newVal) => {
  emit('update:modelValue', newVal)
}, { deep: true })

watch(() => props.modelValue, (newVal) => {
  Object.assign(styleConfig, newVal)
}, { deep: true })
</script>

<style lang="scss" scoped>
.style-config {
  .el-form {
    .el-form-item {
      margin-bottom: 16px;
    }
  }
}
</style>