<template>
  <el-dialog
    v-model="dialogVisible"
    title="NESMA评估因子配置"
    width="900px"
    :close-on-click-modal="false"
    @closed="handleClosed"
  >
    <div class="evaluation-factors">
      <!-- 评估基础信息 -->
      <el-card class="info-card" shadow="never">
        <template #header>
          <div class="card-header">
            <el-icon><InfoFilled /></el-icon>
            <span>评估基础信息</span>
          </div>
        </template>
        <el-row :gutter="20">
          <el-col :span="8">
            <div class="info-item">
              <label>评估名称:</label>
              <span>{{ evaluationInfo.evaluationName }}</span>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="info-item">
              <label>所属项目:</label>
              <span>{{ evaluationInfo.projectName }}</span>
            </div>
          </el-col>
          <el-col :span="8">
            <div class="info-item">
              <label>NESMA规则:</label>
              <span>{{ evaluationInfo.nesmaRules }}</span>
            </div>
          </el-col>
        </el-row>
      </el-card>

      <!-- 功能点类型权重配置 -->
      <el-card class="factors-card" shadow="never">
        <template #header>
          <div class="card-header">
            <el-icon><Grid /></el-icon>
            <span>功能点类型权重配置</span>
            <el-tooltip content="配置各种功能点类型的权重值" placement="top">
              <el-icon class="help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </div>
        </template>

        <el-tabs v-model="activeFactorTab" class="factor-tabs">
          <!-- 数据功能类型 -->
          <el-tab-pane label="数据功能" name="data">
            <div class="factor-section">
              <h4>内部逻辑文件 (ILF - Internal Logical Files)</h4>
              <el-table :data="dataFunctions.ilf" class="factor-table">
                <el-table-column label="复杂度" prop="complexity" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getComplexityTagType(row.complexity)">
                      {{ getComplexityText(row.complexity) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="权重值" width="120">
                  <template #default="{ row }">
                    <el-input-number
                      v-model="row.weight"
                      :min="1"
                      :max="50"
                      controls-position="right"
                      size="small"
                    />
                  </template>
                </el-table-column>
                <el-table-column label="条件说明" prop="description" />
              </el-table>

              <h4 style="margin-top: 30px;">外部接口文件 (EIF - External Interface Files)</h4>
              <el-table :data="dataFunctions.eif" class="factor-table">
                <el-table-column label="复杂度" prop="complexity" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getComplexityTagType(row.complexity)">
                      {{ getComplexityText(row.complexity) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="权重值" width="120">
                  <template #default="{ row }">
                    <el-input-number
                      v-model="row.weight"
                      :min="1"
                      :max="50"
                      controls-position="right"
                      size="small"
                    />
                  </template>
                </el-table-column>
                <el-table-column label="条件说明" prop="description" />
              </el-table>
            </div>
          </el-tab-pane>

          <!-- 事务功能类型 -->
          <el-tab-pane label="事务功能" name="transaction">
            <div class="factor-section">
              <h4>外部输入 (EI - External Inputs)</h4>
              <el-table :data="transactionFunctions.ei" class="factor-table">
                <el-table-column label="复杂度" prop="complexity" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getComplexityTagType(row.complexity)">
                      {{ getComplexityText(row.complexity) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="权重值" width="120">
                  <template #default="{ row }">
                    <el-input-number
                      v-model="row.weight"
                      :min="1"
                      :max="50"
                      controls-position="right"
                      size="small"
                    />
                  </template>
                </el-table-column>
                <el-table-column label="条件说明" prop="description" />
              </el-table>

              <h4 style="margin-top: 30px;">外部输出 (EO - External Outputs)</h4>
              <el-table :data="transactionFunctions.eo" class="factor-table">
                <el-table-column label="复杂度" prop="complexity" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getComplexityTagType(row.complexity)">
                      {{ getComplexityText(row.complexity) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="权重值" width="120">
                  <template #default="{ row }">
                    <el-input-number
                      v-model="row.weight"
                      :min="1"
                      :max="50"
                      controls-position="right"
                      size="small"
                    />
                  </template>
                </el-table-column>
                <el-table-column label="条件说明" prop="description" />
              </el-table>

              <h4 style="margin-top: 30px;">外部查询 (EQ - External Queries)</h4>
              <el-table :data="transactionFunctions.eq" class="factor-table">
                <el-table-column label="复杂度" prop="complexity" width="120">
                  <template #default="{ row }">
                    <el-tag :type="getComplexityTagType(row.complexity)">
                      {{ getComplexityText(row.complexity) }}
                    </el-tag>
                  </template>
                </el-table-column>
                <el-table-column label="权重值" width="120">
                  <template #default="{ row }">
                    <el-input-number
                      v-model="row.weight"
                      :min="1"
                      :max="50"
                      controls-position="right"
                      size="small"
                    />
                  </template>
                </el-table-column>
                <el-table-column label="条件说明" prop="description" />
              </el-table>
            </div>
          </el-tab-pane>

          <!-- 系统调整因子 -->
          <el-tab-pane label="系统调整因子" name="adjustment">
            <div class="factor-section">
              <div class="adjustment-header">
                <h4>通用系统特征 (GSC - General System Characteristics)</h4>
                <p class="factor-desc">
                  每个特征评分范围：0-5分，其中0=无影响，3=一般影响，5=强烈影响
                </p>
              </div>

              <el-row :gutter="20">
                <el-col :span="12" v-for="(factor, index) in adjustmentFactors" :key="index">
                  <div class="adjustment-factor">
                    <div class="factor-header">
                      <span class="factor-name">{{ factor.name }}</span>
                      <el-tooltip :content="factor.description" placement="top">
                        <el-icon class="help-icon"><QuestionFilled /></el-icon>
                      </el-tooltip>
                    </div>
                    <el-slider
                      v-model="factor.value"
                      :min="0"
                      :max="5"
                      :step="1"
                      show-stops
                      show-input
                      :format-tooltip="formatAdjustmentTooltip"
                    />
                    <div class="factor-labels">
                      <span>无影响(0)</span>
                      <span>一般影响(3)</span>
                      <span>强烈影响(5)</span>
                    </div>
                  </div>
                </el-col>
              </el-row>

              <!-- 调整因子计算结果 -->
              <el-card class="calculation-card" shadow="never">
                <template #header>
                  <div class="card-header">
                    <el-icon><Operation /></el-icon>
                    <span>调整因子计算</span>
                  </div>
                </template>
                <el-row :gutter="20">
                  <el-col :span="8">
                    <div class="calc-item">
                      <label>总影响度值 (TDI):</label>
                      <span class="calc-value">{{ totalDegreeOfInfluence }}</span>
                    </div>
                  </el-col>
                  <el-col :span="8">
                    <div class="calc-item">
                      <label>价值调整因子 (VAF):</label>
                      <span class="calc-value">{{ valueAdjustmentFactor.toFixed(2) }}</span>
                    </div>
                  </el-col>
                  <el-col :span="8">
                    <div class="calc-item">
                      <label>调整后功能点:</label>
                      <span class="calc-value">UFP = UFP × {{ valueAdjustmentFactor.toFixed(2) }}</span>
                    </div>
                  </el-col>
                </el-row>
              </el-card>
            </div>
          </el-tab-pane>

          <!-- 复杂度判定规则 -->
          <el-tab-pane label="复杂度规则" name="complexity">
            <div class="factor-section">
              <el-alert
                title="复杂度判定说明"
                type="info"
                :closable="false"
                show-icon
              >
                <p>复杂度判定基于NESMA标准，主要考虑以下因素：</p>
                <ul>
                  <li><strong>DET (Data Element Types)</strong>: 数据元素类型数量</li>
                  <li><strong>RET (Record Element Types)</strong>: 记录元素类型数量</li>
                  <li><strong>FTR (File Type Referenced)</strong>: 引用的文件类型数量</li>
                </ul>
              </el-alert>

              <!-- 数据功能复杂度规则 -->
              <h4>数据功能复杂度判定规则</h4>
              <el-table :data="complexityRules.dataFunctions" class="rules-table">
                <el-table-column label="功能类型" prop="type" width="120" />
                <el-table-column label="低复杂度" prop="low" />
                <el-table-column label="中复杂度" prop="medium" />
                <el-table-column label="高复杂度" prop="high" />
              </el-table>

              <!-- 事务功能复杂度规则 -->
              <h4 style="margin-top: 30px;">事务功能复杂度判定规则</h4>
              <el-table :data="complexityRules.transactionFunctions" class="rules-table">
                <el-table-column label="功能类型" prop="type" width="120" />
                <el-table-column label="低复杂度" prop="low" />
                <el-table-column label="中复杂度" prop="medium" />
                <el-table-column label="高复杂度" prop="high" />
              </el-table>
            </div>
          </el-tab-pane>
        </el-tabs>
      </el-card>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="handleReset">恢复默认</el-button>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button
          type="primary"
          :loading="saving"
          @click="handleSave"
        >
          保存配置
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { 
  InfoFilled, 
  Grid, 
  QuestionFilled, 
  Operation 
} from '@element-plus/icons-vue'
import {
  getEvaluationFactors,
  saveEvaluationFactors,
  updateEvaluationFactors,
  getDefaultNESMAFactors
} from '@/api/nesmaEvaluation'

// Props
const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false
  },
  evaluationId: {
    type: [Number, String],
    default: null
  },
  evaluationInfo: {
    type: Object,
    default: () => ({
      evaluationName: '',
      projectName: '',
      nesmaRules: 'v2.2'
    })
  }
})

// Emits
const emit = defineEmits(['update:modelValue', 'success'])

// 响应式数据
const saving = ref(false)
const activeFactorTab = ref('data')

// 计算属性
const dialogVisible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val)
})

// NESMA标准权重配置
const dataFunctions = reactive({
  ilf: [
    { complexity: 'low', weight: 7, description: '1个RET，1-19个DET 或者 2-5个RET，1-19个DET' },
    { complexity: 'average', weight: 10, description: '1个RET，20-50个DET 或者 2-5个RET，20-50个DET 或者 6+个RET，1-19个DET' },
    { complexity: 'high', weight: 15, description: '1个RET，51+个DET 或者 2-5个RET，51+个DET 或者 6+个RET，20+个DET' }
  ],
  eif: [
    { complexity: 'low', weight: 5, description: '1个RET，1-19个DET 或者 2-5个RET，1-19个DET' },
    { complexity: 'average', weight: 7, description: '1个RET，20-50个DET 或者 2-5个RET，20-50个DET 或者 6+个RET，1-19个DET' },
    { complexity: 'high', weight: 10, description: '1个RET，51+个DET 或者 2-5个RET，51+个DET 或者 6+个RET，20+个DET' }
  ]
})

const transactionFunctions = reactive({
  ei: [
    { complexity: 'low', weight: 3, description: '1-4个DET，0-1个FTR 或者 5-15个DET，0-1个FTR' },
    { complexity: 'average', weight: 4, description: '1-4个DET，2个FTR 或者 5-15个DET，2个FTR 或者 16+个DET，0-1个FTR' },
    { complexity: 'high', weight: 6, description: '5-15个DET，3+个FTR 或者 16+个DET，2+个FTR' }
  ],
  eo: [
    { complexity: 'low', weight: 4, description: '1-5个DET，0-1个FTR 或者 6-19个DET，0-1个FTR' },
    { complexity: 'average', weight: 5, description: '1-5个DET，2-3个FTR 或者 6-19个DET，2-3个FTR 或者 20+个DET，0-1个FTR' },
    { complexity: 'high', weight: 7, description: '6-19个DET，4+个FTR 或者 20+个DET，2+个FTR' }
  ],
  eq: [
    { complexity: 'low', weight: 3, description: '1-5个DET，0-1个FTR 或者 6-19个DET，0-1个FTR' },
    { complexity: 'average', weight: 4, description: '1-5个DET，2-3个FTR 或者 6-19个DET，2-3个FTR 或者 20+个DET，0-1个FTR' },
    { complexity: 'high', weight: 6, description: '6-19个DET，4+个FTR 或者 20+个DET，2+个FTR' }
  ]
})

// 系统调整因子（14个通用系统特征）
const adjustmentFactors = reactive([
  { name: '数据通信', value: 3, description: '应用程序与其用户直接通信的程度' },
  { name: '分布式数据处理', value: 3, description: '分布式数据和处理功能的特征程度' },
  { name: '性能', value: 3, description: '应用程序性能目标（响应时间、吞吐量）对设计的影响' },
  { name: '系统配置负荷严重', value: 3, description: '处理器利用率对应用程序设计的影响' },
  { name: '交易量', value: 3, description: '预期的交易量对应用程序设计的影响' },
  { name: '在线数据录入', value: 3, description: '通过交互式交易进行在线数据录入的数量' },
  { name: '最终用户效率', value: 3, description: '为用户效率而设计的在线功能' },
  { name: '在线更新', value: 3, description: '应用程序提供的在线更新能力' },
  { name: '复杂的处理', value: 3, description: '应用程序中复杂处理的程度' },
  { name: '可重用性', value: 3, description: '应用程序代码专门为在其他应用程序中重用而开发' },
  { name: '安装容易性', value: 3, description: '转换和安装应用程序的难易程度' },
  { name: '操作容易性', value: 3, description: '应用程序的有效和/或自动启动、备份和恢复过程' },
  { name: '多站点', value: 3, description: '应用程序专门为支持多个站点和组织而设计和开发' },
  { name: '便于变更', value: 3, description: '应用程序专门为便于变更而设计、开发和支持' }
])

// 复杂度判定规则
const complexityRules = reactive({
  dataFunctions: [
    {
      type: 'ILF/EIF',
      low: '1个RET且1-19个DET，或2-5个RET且1-19个DET',
      medium: '1个RET且20-50个DET，或2-5个RET且20-50个DET，或6+个RET且1-19个DET',
      high: '1个RET且51+个DET，或2-5个RET且51+个DET，或6+个RET且20+个DET'
    }
  ],
  transactionFunctions: [
    {
      type: 'EI',
      low: '1-4个DET且0-1个FTR，或5-15个DET且0-1个FTR',
      medium: '1-4个DET且2个FTR，或5-15个DET且2个FTR，或16+个DET且0-1个FTR',
      high: '5-15个DET且3+个FTR，或16+个DET且2+个FTR'
    },
    {
      type: 'EO/EQ',
      low: '1-5个DET且0-1个FTR，或6-19个DET且0-1个FTR',
      medium: '1-5个DET且2-3个FTR，或6-19个DET且2-3个FTR，或20+个DET且0-1个FTR',
      high: '6-19个DET且4+个FTR，或20+个DET且2+个FTR'
    }
  ]
})

// 计算总影响度值
const totalDegreeOfInfluence = computed(() => {
  return adjustmentFactors.reduce((sum, factor) => sum + factor.value, 0)
})

// 计算价值调整因子
const valueAdjustmentFactor = computed(() => {
  return (totalDegreeOfInfluence.value * 0.01) + 0.65
})

// 方法
const getComplexityTagType = (complexity) => {
  const typeMap = {
    'low': 'success',
    'average': 'warning', 
    'high': 'danger'
  }
  return typeMap[complexity] || 'info'
}

const getComplexityText = (complexity) => {
  const textMap = {
    'low': '低',
    'average': '中',
    'high': '高'
  }
  return textMap[complexity] || '未知'
}

const formatAdjustmentTooltip = (value) => {
  const levelMap = {
    0: '无影响',
    1: '偶然影响',
    2: '轻微影响',
    3: '一般影响',
    4: '明显影响',
    5: '强烈影响'
  }
  return `${value} - ${levelMap[value] || '未知'}`
}

// 保存配置
const handleSave = async () => {
  try {
    saving.value = true

    // 收集所有配置数据，使用深拷贝避免引用问题
    const factorConfig = {
      evaluationId: props.evaluationId,
      dataFunctions: JSON.parse(JSON.stringify(dataFunctions)),
      transactionFunctions: JSON.parse(JSON.stringify(transactionFunctions)),
      adjustmentFactors: adjustmentFactors.map(factor => ({
        name: factor.name,
        value: factor.value,
        description: factor.description || ''
      })),
      calculatedValues: {
        totalDegreeOfInfluence: totalDegreeOfInfluence.value,
        valueAdjustmentFactor: valueAdjustmentFactor.value
      }
    }
    
    // 调试：在保存前检查EO数据
    console.log('保存前的EO数据:', {
      eoCount: factorConfig.transactionFunctions.eo?.length || 0,
      eoData: factorConfig.transactionFunctions.eo
    })

    // 调用API保存配置
    if (props.evaluationId) {
      await updateEvaluationFactors(props.evaluationId, factorConfig)
    } else {
      await saveEvaluationFactors(factorConfig)
    }

    ElMessage.success('评估因子配置保存成功')
    emit('success', factorConfig)
    dialogVisible.value = false
  } catch (error) {
    console.error('保存评估因子配置失败:', error)
    ElMessage.error('保存失败: ' + (error.response?.data?.msg || error.message || '未知错误'))
  } finally {
    saving.value = false
  }
}

// 恢复默认配置
const handleReset = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要恢复为NESMA标准默认配置吗？这将清除当前所有自定义设置。',
      '确认恢复默认',
      {
        type: 'warning',
        confirmButtonText: '确定恢复',
        cancelButtonText: '取消'
      }
    )

    // 恢复所有默认值
    resetToDefaults()
    ElMessage.success('已恢复为NESMA标准默认配置')
  } catch (error) {
    // 用户取消，不做任何操作
  }
}

// 重置为默认值
const resetToDefaults = () => {
  // 重置数据功能权重
  dataFunctions.ilf.forEach((item, index) => {
    const defaultWeights = [7, 10, 15]
    item.weight = defaultWeights[index]
  })
  
  dataFunctions.eif.forEach((item, index) => {
    const defaultWeights = [5, 7, 10]
    item.weight = defaultWeights[index]
  })

  // 重置事务功能权重
  transactionFunctions.ei.forEach((item, index) => {
    const defaultWeights = [3, 4, 6]
    item.weight = defaultWeights[index]
  })

  transactionFunctions.eo.forEach((item, index) => {
    const defaultWeights = [4, 5, 7]
    item.weight = defaultWeights[index]
  })

  transactionFunctions.eq.forEach((item, index) => {
    const defaultWeights = [3, 4, 6]
    item.weight = defaultWeights[index]
  })

  // 重置调整因子
  adjustmentFactors.forEach(factor => {
    factor.value = 3
  })
}

const handleClosed = () => {
  // 对话框关闭时的清理工作
  activeFactorTab.value = 'data'
}

// 加载评估因子配置
const loadEvaluationFactors = async (evaluationId) => {
  try {
    const response = await getEvaluationFactors(evaluationId)
    const config = response.data || {}
    
    // 更新数据功能配置
    if (config.dataFunctions) {
      if (config.dataFunctions.ilf) {
        dataFunctions.ilf.forEach((item, index) => {
          if (config.dataFunctions.ilf[index]) {
            item.weight = config.dataFunctions.ilf[index].weight
          }
        })
      }
      if (config.dataFunctions.eif) {
        dataFunctions.eif.forEach((item, index) => {
          if (config.dataFunctions.eif[index]) {
            item.weight = config.dataFunctions.eif[index].weight
          }
        })
      }
    }
    
    // 更新事务功能配置
    if (config.transactionFunctions) {
      ['ei', 'eo', 'eq'].forEach(type => {
        if (config.transactionFunctions[type]) {
          transactionFunctions[type].forEach((item, index) => {
            if (config.transactionFunctions[type][index]) {
              item.weight = config.transactionFunctions[type][index].weight
            }
          })
        }
      })
    }
    
    // 更新调整因子
    if (config.adjustmentFactors) {
      config.adjustmentFactors.forEach((configFactor, index) => {
        if (adjustmentFactors[index]) {
          adjustmentFactors[index].value = configFactor.value
        }
      })
    }
    
    console.log('评估因子配置加载成功')
  } catch (error) {
    console.error('加载评估因子配置失败:', error)
    // 如果加载失败，使用默认配置
    if (error.response?.status === 404) {
      console.log('未找到已保存的配置，使用默认配置')
    } else {
      ElMessage.warning('加载已保存的配置失败，使用默认配置')
    }
  }
}

// 加载默认NESMA配置
const loadDefaultFactors = async (nesmaVersion = 'v2.2') => {
  try {
    const response = await getDefaultNESMAFactors(nesmaVersion)
    const defaultConfig = response.data || {}
    
    // 应用默认配置
    if (defaultConfig.dataFunctions) {
      // 更新数据功能默认权重
      // ...
    }
    
    console.log('默认NESMA配置加载成功')
  } catch (error) {
    console.error('加载默认NESMA配置失败:', error)
    // 使用内置默认值
  }
}

// 监听评估ID变化，加载对应的配置
watch(() => props.evaluationId, (newId) => {
  if (newId && dialogVisible.value) {
    loadEvaluationFactors(newId)
  }
}, { immediate: true })

// 监听对话框打开，加载配置
watch(() => dialogVisible.value, (visible) => {
  if (visible && props.evaluationId) {
    loadEvaluationFactors(props.evaluationId)
  }
})
</script>

<style scoped>
.evaluation-factors {
  max-height: 70vh;
  overflow-y: auto;
}

.info-card {
  margin-bottom: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
}

.help-icon {
  color: #909399;
  cursor: help;
  margin-left: 5px;
}

.info-item {
  margin-bottom: 10px;
}

.info-item label {
  font-weight: 600;
  margin-right: 8px;
  color: #606266;
}

.factor-section {
  padding: 10px 0;
}

.factor-section h4 {
  margin: 20px 0 15px 0;
  color: #303133;
  font-size: 16px;
  border-left: 4px solid #409eff;
  padding-left: 10px;
}

.factor-table {
  margin-bottom: 20px;
}

.factor-table :deep(.el-table__header) {
  background-color: #f5f7fa;
}

.adjustment-header {
  margin-bottom: 30px;
}

.factor-desc {
  color: #606266;
  margin: 10px 0;
  font-size: 14px;
}

.adjustment-factor {
  margin-bottom: 25px;
  padding: 15px;
  background-color: #fafbfc;
  border-radius: 6px;
  border: 1px solid #ebeef5;
}

.factor-header {
  display: flex;
  align-items: center;
  margin-bottom: 15px;
}

.factor-name {
  font-weight: 600;
  color: #303133;
}

.factor-labels {
  display: flex;
  justify-content: space-between;
  margin-top: 10px;
  font-size: 12px;
  color: #909399;
}

.calculation-card {
  margin-top: 30px;
  background-color: #f0f9ff;
  border: 1px solid #c6e8ff;
}

.calc-item {
  text-align: center;
  padding: 10px;
}

.calc-item label {
  display: block;
  font-weight: 600;
  color: #606266;
  margin-bottom: 5px;
}

.calc-value {
  font-size: 18px;
  font-weight: 700;
  color: #409eff;
}

.rules-table {
  margin-bottom: 20px;
}

.rules-table :deep(.el-table__header) {
  background-color: #f5f7fa;
}

.dialog-footer {
  text-align: right;
}

/* 标签页样式 */
.factor-tabs :deep(.el-tabs__header) {
  margin-bottom: 20px;
}

.factor-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #e4e7ed;
}

.factor-tabs :deep(.el-tabs__active-bar) {
  background-color: #409eff;
}

.factor-tabs :deep(.el-tabs__item.is-active) {
  color: #409eff;
  font-weight: 600;
}

/* 响应式设计 */
@media (max-width: 768px) {
  .evaluation-factors {
    max-height: 60vh;
  }
  
  .adjustment-factor {
    margin-bottom: 15px;
    padding: 10px;
  }
  
  .calc-item {
    margin-bottom: 15px;
  }
}
</style>