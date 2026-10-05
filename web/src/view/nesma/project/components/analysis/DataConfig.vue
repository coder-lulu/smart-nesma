<template>
  <div class="data-config">
    <el-form :model="dataConfig" label-width="100px" :disabled="disabled">
      <el-form-item label="数据格式">
        <el-radio-group v-model="dataConfig.format">
          <el-radio label="table">表格</el-radio>
          <el-radio label="list">列表</el-radio>
          <el-radio label="tree">树形</el-radio>
        </el-radio-group>
      </el-form-item>
      
      <el-form-item label="数据精度">
        <el-input-number v-model="dataConfig.precision" :min="0" :max="6" />
      </el-form-item>
      
      <el-form-item label="数据过滤">
        <el-form :model="dataConfig.filter" label-width="80px">
          <el-form-item label="空值">
            <el-switch v-model="dataConfig.filter.excludeNull" />
          </el-form-item>
          <el-form-item label="零值">
            <el-switch v-model="dataConfig.filter.excludeZero" />
          </el-form-item>
          <el-form-item label="重复值">
            <el-switch v-model="dataConfig.filter.excludeDuplicate" />
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="数据排序">
        <el-form :model="dataConfig.sort" label-width="80px">
          <el-form-item label="字段">
            <el-select v-model="dataConfig.sort.field" placeholder="选择排序字段">
              <el-option label="名称" value="name" />
              <el-option label="创建时间" value="created_at" />
              <el-option label="更新时间" value="updated_at" />
              <el-option label="功能点" value="function_points" />
              <el-option label="复杂度" value="complexity" />
            </el-select>
          </el-form-item>
          <el-form-item label="顺序">
            <el-select v-model="dataConfig.sort.order">
              <el-option label="升序" value="asc" />
              <el-option label="降序" value="desc" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="数据分组">
        <el-form :model="dataConfig.group" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="dataConfig.group.enabled" />
          </el-form-item>
          <el-form-item label="字段" v-if="dataConfig.group.enabled">
            <el-select v-model="dataConfig.group.field" placeholder="选择分组字段">
              <el-option label="项目" value="project" />
              <el-option label="周期" value="cycle" />
              <el-option label="层级" value="level" />
              <el-option label="类型" value="type" />
              <el-option label="状态" value="status" />
            </el-select>
          </el-form-item>
          <el-form-item label="统计" v-if="dataConfig.group.enabled">
            <el-checkbox-group v-model="dataConfig.group.statistics">
              <el-checkbox label="count">计数</el-checkbox>
              <el-checkbox label="sum">求和</el-checkbox>
              <el-checkbox label="avg">平均</el-checkbox>
              <el-checkbox label="max">最大</el-checkbox>
              <el-checkbox label="min">最小</el-checkbox>
            </el-checkbox-group>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="图表设置">
        <el-form :model="dataConfig.chart" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="dataConfig.chart.enabled" />
          </el-form-item>
          <el-form-item label="类型" v-if="dataConfig.chart.enabled">
            <el-select v-model="dataConfig.chart.type" placeholder="选择图表类型">
              <el-option label="柱状图" value="bar" />
              <el-option label="折线图" value="line" />
              <el-option label="饼图" value="pie" />
              <el-option label="雷达图" value="radar" />
              <el-option label="散点图" value="scatter" />
            </el-select>
          </el-form-item>
          <el-form-item label="大小" v-if="dataConfig.chart.enabled">
            <el-form :model="dataConfig.chart.size" label-width="60px">
              <el-form-item label="宽度">
                <el-input-number v-model="dataConfig.chart.size.width" :min="200" :max="1000" />
              </el-form-item>
              <el-form-item label="高度">
                <el-input-number v-model="dataConfig.chart.size.height" :min="150" :max="800" />
              </el-form-item>
            </el-form>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="数据导出">
        <el-form :model="dataConfig.export" label-width="80px">
          <el-form-item label="原始数据">
            <el-switch v-model="dataConfig.export.includeRawData" />
          </el-form-item>
          <el-form-item label="计算数据">
            <el-switch v-model="dataConfig.export.includeCalculated" />
          </el-form-item>
          <el-form-item label="元数据">
            <el-switch v-model="dataConfig.export.includeMetadata" />
          </el-form-item>
          <el-form-item label="附件">
            <el-switch v-model="dataConfig.export.includeAttachments" />
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="数据缓存">
        <el-form :model="dataConfig.cache" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="dataConfig.cache.enabled" />
          </el-form-item>
          <el-form-item label="时长" v-if="dataConfig.cache.enabled">
            <el-input-number v-model="dataConfig.cache.duration" :min="1" :max="3600" />
            <span style="margin-left: 8px;">秒</span>
          </el-form-item>
          <el-form-item label="策略" v-if="dataConfig.cache.enabled">
            <el-select v-model="dataConfig.cache.strategy">
              <el-option label="LRU" value="lru" />
              <el-option label="LFU" value="lfu" />
              <el-option label="TTL" value="ttl" />
            </el-select>
          </el-form-item>
        </el-form>
      </el-form-item>
      
      <el-form-item label="数据验证">
        <el-form :model="dataConfig.validation" label-width="80px">
          <el-form-item label="启用">
            <el-switch v-model="dataConfig.validation.enabled" />
          </el-form-item>
          <el-form-item label="严格模式" v-if="dataConfig.validation.enabled">
            <el-switch v-model="dataConfig.validation.strict" />
          </el-form-item>
          <el-form-item label="规则" v-if="dataConfig.validation.enabled">
            <el-checkbox-group v-model="dataConfig.validation.rules">
              <el-checkbox label="required">必填验证</el-checkbox>
              <el-checkbox label="type">类型验证</el-checkbox>
              <el-checkbox label="range">范围验证</el-checkbox>
              <el-checkbox label="format">格式验证</el-checkbox>
            </el-checkbox-group>
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
      format: 'table',
      precision: 2,
      filter: {
        excludeNull: true,
        excludeZero: false,
        excludeDuplicate: false
      },
      sort: {
        field: 'created_at',
        order: 'desc'
      },
      group: {
        enabled: false,
        field: 'level',
        statistics: ['count', 'sum']
      },
      chart: {
        enabled: true,
        type: 'bar',
        size: {
          width: 600,
          height: 400
        }
      },
      export: {
        includeRawData: true,
        includeCalculated: true,
        includeMetadata: false,
        includeAttachments: false
      },
      cache: {
        enabled: true,
        duration: 300,
        strategy: 'lru'
      },
      validation: {
        enabled: true,
        strict: false,
        rules: ['required', 'type']
      }
    })
  },
  disabled: {
    type: Boolean,
    default: false
  }
})

const emit = defineEmits(['update:modelValue'])

const dataConfig = reactive({ ...props.modelValue })

watch(dataConfig, (newVal) => {
  emit('update:modelValue', newVal)
}, { deep: true })

watch(() => props.modelValue, (newVal) => {
  Object.assign(dataConfig, newVal)
}, { deep: true })
</script>

<style lang="scss" scoped>
.data-config {
  .el-form {
    .el-form-item {
      margin-bottom: 16px;
    }
  }
}
</style>