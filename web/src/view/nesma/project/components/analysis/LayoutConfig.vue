<template>
  <div class="layout-config">
    <el-form :model="layoutConfig" label-width="100px" :disabled="disabled">
      <el-form-item label="页面方向">
        <el-radio-group v-model="layoutConfig.orientation">
          <el-radio label="portrait">竖向</el-radio>
          <el-radio label="landscape">横向</el-radio>
        </el-radio-group>
      </el-form-item>
      
      <el-form-item label="页面大小">
        <el-select v-model="layoutConfig.pageSize" placeholder="选择页面大小">
          <el-option label="A4" value="A4" />
          <el-option label="A3" value="A3" />
          <el-option label="A5" value="A5" />
          <el-option label="Letter" value="Letter" />
          <el-option label="Legal" value="Legal" />
        </el-select>
      </el-form-item>
      
      <el-form-item label="页边距">
        <div class="margin-config">
          <el-form :model="layoutConfig.margin" label-width="60px">
            <el-form-item label="上">
              <el-input-number v-model="layoutConfig.margin.top" :min="0" :max="50" />
            </el-form-item>
            <el-form-item label="右">
              <el-input-number v-model="layoutConfig.margin.right" :min="0" :max="50" />
            </el-form-item>
            <el-form-item label="下">
              <el-input-number v-model="layoutConfig.margin.bottom" :min="0" :max="50" />
            </el-form-item>
            <el-form-item label="左">
              <el-input-number v-model="layoutConfig.margin.left" :min="0" :max="50" />
            </el-form-item>
          </el-form>
        </div>
      </el-form-item>
      
      <el-form-item label="页眉设置">
        <el-form :model="layoutConfig.header" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="layoutConfig.header.enabled" />
          </el-form-item>
          <el-form-item label="高度" v-if="layoutConfig.header.enabled">
            <el-input-number v-model="layoutConfig.header.height" :min="20" :max="100" />
          </el-form-item>
          <el-form-item label="内容" v-if="layoutConfig.header.enabled">
            <el-input v-model="layoutConfig.header.content" placeholder="页眉内容" />
          </el-form-item>
          <el-form-item label="对齐" v-if="layoutConfig.header.enabled">
            <el-select v-model="layoutConfig.header.align">
              <el-option label="左对齐" value="left" />
              <el-option label="居中" value="center" />
              <el-option label="右对齐" value="right" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="页脚设置">
        <el-form :model="layoutConfig.footer" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="layoutConfig.footer.enabled" />
          </el-form-item>
          <el-form-item label="高度" v-if="layoutConfig.footer.enabled">
            <el-input-number v-model="layoutConfig.footer.height" :min="20" :max="100" />
          </el-form-item>
          <el-form-item label="内容" v-if="layoutConfig.footer.enabled">
            <el-input v-model="layoutConfig.footer.content" placeholder="页脚内容，支持变量：{page}、{total}、{date}" />
          </el-form-item>
          <el-form-item label="对齐" v-if="layoutConfig.footer.enabled">
            <el-select v-model="layoutConfig.footer.align">
              <el-option label="左对齐" value="left" />
              <el-option label="居中" value="center" />
              <el-option label="右对齐" value="right" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="目录设置">
        <el-form :model="layoutConfig.toc" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="layoutConfig.toc.enabled" />
          </el-form-item>
          <el-form-item label="位置" v-if="layoutConfig.toc.enabled">
            <el-select v-model="layoutConfig.toc.position">
              <el-option label="开始" value="start" />
              <el-option label="结束" value="end" />
            </el-select>
          </el-form-item>
          <el-form-item label="深度" v-if="layoutConfig.toc.enabled">
            <el-slider v-model="layoutConfig.toc.depth" :min="1" :max="6" show-stops />
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="分页设置">
        <el-form :model="layoutConfig.pagination" label-width="80px">
          <el-form-item label="章节分页">
            <el-switch v-model="layoutConfig.pagination.sectionBreak" />
          </el-form-item>
          <el-form-item label="表格分页">
            <el-switch v-model="layoutConfig.pagination.tableBreak" />
          </el-form-item>
          <el-form-item label="图表分页">
            <el-switch v-model="layoutConfig.pagination.chartBreak" />
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="水印设置">
        <el-form :model="layoutConfig.watermark" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="layoutConfig.watermark.enabled" />
          </el-form-item>
          <el-form-item label="文字" v-if="layoutConfig.watermark.enabled">
            <el-input v-model="layoutConfig.watermark.text" placeholder="水印文字" />
          </el-form-item>
          <el-form-item label="透明度" v-if="layoutConfig.watermark.enabled">
            <el-slider v-model="layoutConfig.watermark.opacity" :min="0" :max="1" :step="0.1" />
          </el-form-item>
          <el-form-item label="角度" v-if="layoutConfig.watermark.enabled">
            <el-slider v-model="layoutConfig.watermark.angle" :min="-90" :max="90" />
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
      orientation: 'portrait',
      pageSize: 'A4',
      margin: {
        top: 20,
        right: 20,
        bottom: 20,
        left: 20
      },
      header: {
        enabled: true,
        height: 60,
        content: '{project_name} - {report_type}',
        align: 'center'
      },
      footer: {
        enabled: true,
        height: 40,
        content: '第 {page} 页，共 {total} 页',
        align: 'center'
      },
      toc: {
        enabled: true,
        position: 'start',
        depth: 3
      },
      pagination: {
        sectionBreak: true,
        tableBreak: false,
        chartBreak: false
      },
      watermark: {
        enabled: false,
        text: '',
        opacity: 0.3,
        angle: -45
      }
    })
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const layoutConfig = reactive({ ...props.modelValue })

watch(layoutConfig, (newVal) => {
  emit('update:modelValue', newVal)
}, { deep: true })

watch(() => props.modelValue, (newVal) => {
  Object.assign(layoutConfig, newVal)
}, { deep: true })
</script>

<style lang="scss" scoped>
.layout-config {
  .margin-config {
    .el-form {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 12px;
      
      .el-form-item {
        margin-bottom: 12px;
      }
    }
  }
}
</style>