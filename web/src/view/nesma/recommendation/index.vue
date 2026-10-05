<template>
  <div class="recommendation-container">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1>🤖 AI智能推荐中心</h1>
      <p>基于知识库和大模型为您提供智能化的需求分析建议和最佳实践</p>
    </div>

    <!-- 推荐筛选器 -->
    <div class="filter-section">
      <el-card class="filter-card">
        <el-row :gutter="20">
          <el-col :span="6">
            <el-select v-model="filterForm.type" placeholder="推荐类型" @change="loadRecommendations">
              <el-option label="全部" value="all"></el-option>
              <el-option label="需求模板" value="requirement"></el-option>
              <el-option label="文档模板" value="template"></el-option>
              <el-option label="AI助手" value="agent"></el-option>
              <el-option label="项目" value="project"></el-option>
              <el-option label="知识库" value="knowledge"></el-option>
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-select v-model="filterForm.algorithm" placeholder="推荐算法" @change="loadRecommendations">
              <el-option label="混合推荐" value="hybrid"></el-option>
              <el-option label="协同过滤" value="collaborative"></el-option>
              <el-option label="内容推荐" value="content_based"></el-option>
              <el-option label="热门推荐" value="popular"></el-option>
            </el-select>
          </el-col>
          <el-col :span="6">
            <el-input-number v-model="filterForm.limit" :min="5" :max="50" placeholder="推荐数量" @change="loadRecommendations"></el-input-number>
          </el-col>
          <el-col :span="6">
            <el-button 
              type="primary" 
              @click="loadRecommendations" 
              icon="el-icon-refresh"
              :loading="aiLoading"
            >
              {{ aiLoading ? '🤖 AI分析中...' : '🤖 AI智能推荐' }}
            </el-button>
          </el-col>
        </el-row>
      </el-card>
    </div>

    <!-- 推荐内容展示 -->
    <div class="recommendations-section">
      <el-row :gutter="20">
        <el-col :span="24">
          <el-card class="content-card">
            <template #header>
              <div class="card-header">
                <span>为您推荐 ({{ recommendations.total || 0 }}项)</span>
                <div class="header-right">
                  <el-tag type="success" size="small" v-if="isAIRecommendation">
                    <el-icon><Star /></el-icon>
                    AI驱动
                  </el-tag>
                  <el-tag type="info" size="small" v-else>
                    📚 知识库
                  </el-tag>
                  <span class="algorithm-tag">{{ recommendations.algorithm || '推荐算法' }}</span>
                </div>
              </div>
            </template>
            
            <!-- 推荐项列表 -->
            <div v-loading="loading" class="recommendations-list">
              <!-- 首次加载提示 -->
              <div v-if="!isAIRecommendation && recommendations.items.length === 0" class="welcome-prompt">
                <div class="prompt-content">
                  <el-icon size="48" color="#409EFF"><Document /></el-icon>
                  <h3>欢迎使用智能推荐</h3>
                  <p>当前显示知识库推荐，点击"AI智能推荐"按钮获取基于大模型的个性化推荐</p>
                  <el-button type="primary" size="large" @click="loadRecommendations" :loading="aiLoading">
                    <el-icon><Star /></el-icon>
                    {{ aiLoading ? 'AI分析中...' : '获取AI推荐' }}
                  </el-button>
                </div>
              </div>
              
              <div v-for="item in (recommendations.items || [])" :key="item.itemId" class="recommendation-item">
                <el-card class="item-card" shadow="hover" @click="handleItemClick(item)">
                  <div class="item-header">
                    <div class="item-title">
                      <el-tag :type="getTypeColor(item.type)" size="small">{{ getTypeLabel(item.type) }}</el-tag>
                      <h3>{{ item.title || '未命名' }}</h3>
                    </div>
                    <div class="item-score">
                      <el-rate v-model="item.displayScore" disabled show-score text-color="#ff9900" :max="5"></el-rate>
                    </div>
                  </div>
                  
                  <div class="item-content">
                    <p class="item-description">{{ item.description || '暂无描述' }}</p>
                    <div class="item-meta">
                      <el-tag v-if="item.category" size="small" type="info">{{ item.category }}</el-tag>
                      <el-tag v-for="tag in (item.tags || []).slice(0, 3)" :key="tag" size="small" class="tag-item">{{ tag }}</el-tag>
                    </div>
                  </div>
                  
                  <div class="item-footer">
                    <span class="reason">{{ item.reason || '系统推荐' }}</span>
                    <div class="item-actions">
                      <el-button size="small" type="text" @click.stop="viewDetails(item)">查看详情</el-button>
                      <el-button size="small" type="success" @click.stop="adoptAIRecommendation(item)" v-if="isAIRecommendation">立即采纳</el-button>
                      <el-button size="small" type="primary" @click.stop="editItem(item)">编辑</el-button>
                      <el-button size="small" type="text" @click.stop="provideFeedback(item)">反馈</el-button>
                    </div>
                  </div>
                </el-card>
              </div>
            </div>

            <!-- 空状态 -->
            <el-empty v-if="!loading && (!recommendations.items || recommendations.items.length === 0)" description="暂无推荐内容"></el-empty>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 反馈对话框 -->
    <el-dialog v-model="feedbackDialog.visible" title="推荐反馈" width="500px">
      <el-form :model="feedbackDialog.form" label-width="80px">
        <el-form-item label="满意度">
          <el-rate v-model="feedbackDialog.form.score" show-text></el-rate>
        </el-form-item>
        <el-form-item label="反馈意见">
          <el-input v-model="feedbackDialog.form.comment" type="textarea" :rows="4" placeholder="请输入您的反馈意见"></el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="feedbackDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="submitFeedback">提交反馈</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Star, Document } from '@element-plus/icons-vue'
import { 
  getRecommendations, 
  recordUserBehavior, 
  submitRecommendationFeedback,
  getAIRecommendations,
  adoptRecommendation,
  rejectRecommendation,
  getQuickRecommendations
} from '@/api/recommendation'

// 响应式数据
const loading = ref(false)
const aiLoading = ref(false)  // AI推荐加载状态
const isAIRecommendation = ref(false)  // 是否是AI推荐
const filterForm = reactive({
  type: 'all',
  algorithm: 'hybrid',
  limit: 20
})

const recommendations = ref({
  items: [],
  total: 0,
  algorithm: '',
  timestamp: null
})

const feedbackDialog = reactive({
  visible: false,
  currentItem: null,
  form: {
    score: 5,
    comment: ''
  }
})

// 计算属性
const displayScore = computed(() => (score) => {
  return Math.round(score * 5) // 将0-1分数转换为1-5星级
})

// 生命周期
onMounted(() => {
  loadKnowledgeBasedRecommendations()  // 首次加载显示知识库内容
})

// 方法
// 加载知识库推荐（首次加载）
const loadKnowledgeBasedRecommendations = async () => {
  try {
    loading.value = true
    isAIRecommendation.value = false
    
    // 调用传统推荐API获取知识库内容
    const response = await getRecommendations({
      recommendationType: 'knowledge',
      algorithm: 'content_based',
      limit: filterForm.limit,
      diversify: true,
      minScore: 0.1
    })
    
    const data = response.data || {}
    recommendations.value = {
      items: data.items || [],
      total: data.total || 0,
      algorithm: '知识库推荐',
      timestamp: data.timestamp || Date.now()
    }
    
    // 为每个推荐项添加显示分数
    recommendations.value.items.forEach(item => {
      item.displayScore = Math.round((item.score || 0) * 5)
    })
    
    if (recommendations.value.items.length === 0) {
      ElMessage.info('暂无知识库推荐，请点击"刷新推荐"获取AI推荐')
    } else {
      ElMessage.success(`加载了 ${recommendations.value.items.length} 条知识库推荐`)
    }
    
  } catch (error) {
    console.error('加载知识库推荐失败:', error)
    ElMessage.warning('暂无知识库推荐，请点击"刷新推荐"获取AI推荐')
    
    // 设置默认值
    recommendations.value = {
      items: [],
      total: 0,
      algorithm: '知识库推荐',
      timestamp: Date.now()
    }
  } finally {
    loading.value = false
  }
}

// AI智能推荐（点击刷新推荐时调用）
const loadRecommendations = async () => {
  try {
    loading.value = true
    aiLoading.value = true
    isAIRecommendation.value = true
    
    // 获取项目功能点信息（模拟数据，实际应该从API获取）
    const projectRequirements = await getProjectRequirements()
    
    // 调用AI智能推荐API
    const response = await getAIRecommendations({
      projectId: 1, // 暂时使用固定项目ID，实际应该从当前项目获取
      projectName: "智能推荐测试项目",
      domain: "general",
      requirementText: `基于${filterForm.type}类型的推荐需求分析`,
      recommendationType: filterForm.type,
      algorithm: filterForm.algorithm,
      requirements: projectRequirements,
      context: {
        requestType: filterForm.type,
        algorithm: filterForm.algorithm,
        limit: filterForm.limit,
        timestamp: new Date().toISOString(),
        userPreferences: {
          focusArea: filterForm.type,
          analysisDepth: 'detailed',
          includeRisks: true
        }
      }
    })
    
    // 处理AI推荐数据
    const data = response.data || {}
    const aiRecommendations = data.recommendations || []
    
    // 转换AI推荐结果为传统推荐格式以兼容现有UI
    const transformedItems = aiRecommendations.map((aiRec, index) => ({
      itemId: Date.now() + index, // 生成数字ID用于用户行为记录
      originalId: aiRec.id || `ai_rec_${index}`, // 保留原始ID用于采纳推荐
      type: aiRec.category || 'knowledge',
      title: aiRec.title || 'AI智能推荐',
      description: aiRec.content || '基于AI大模型生成的推荐内容',
      score: aiRec.confidence || 0.8,
      reason: `AI推荐 - 置信度${Math.round((aiRec.confidence || 0.8) * 100)}%`,
      category: aiRec.category || 'ai_generated',
      tags: aiRec.bestPractices || [],
      metadata: {
        knowledgeRefs: aiRec.knowledgeRefs || [],
        improvements: aiRec.improvements || [],
        riskWarnings: aiRec.riskWarnings || [],
        functionType: aiRec.functionType
      }
    }))
    
    recommendations.value = {
      items: transformedItems,
      total: transformedItems.length,
      algorithm: 'AI-Powered',
      timestamp: data.createdAt || Date.now()
    }
    
    // 为每个推荐项添加显示分数
    recommendations.value.items.forEach(item => {
      item.displayScore = Math.round((item.score || 0) * 5)
    })
    
    ElMessage.success(`🤖 AI生成了 ${transformedItems.length} 条智能推荐`)
  } catch (error) {
    console.error('加载AI推荐失败:', error)
    ElMessage.error('🤖 AI推荐失败: ' + (error.response?.data?.msg || error.message || '未知错误'))
    
    // 如果AI推荐失败，回退到知识库推荐
    isAIRecommendation.value = false
    await loadKnowledgeBasedRecommendations()
    
  } finally {
    loading.value = false
    aiLoading.value = false
  }
}

// 获取项目功能点信息（模拟）
const getProjectRequirements = async () => {
  // 模拟项目功能点数据，实际应该从API获取
  return [
    {
      id: 1,
      level: 1,
      code: "F001",
      title: "用户管理模块",
      description: "负责用户注册、登录、权限管理等功能",
      children: [
        {
          id: 2,
          level: 2,
          code: "F001.01",
          title: "用户注册",
          description: "用户通过邮箱或手机号注册账户",
          children: [
            {
              id: 3,
              level: 3,
              code: "F001.01.01",
              title: "基本信息填写",
              description: "输入用户名、邮箱、密码等基本信息",
              children: []
            },
            {
              id: 4,
              level: 3,
              code: "F001.01.02",
              title: "验证码验证",
              description: "发送和验证邮箱/短信验证码",
              children: []
            }
          ]
        },
        {
          id: 5,
          level: 2,
          code: "F001.02",
          title: "用户登录",
          description: "用户通过用户名/邮箱登录系统",
          children: [
            {
              id: 6,
              level: 3,
              code: "F001.02.01",
              title: "登录验证",
              description: "验证用户名和密码的正确性",
              children: []
            },
            {
              id: 7,
              level: 3,
              code: "F001.02.02",
              title: "会话管理",
              description: "创建和管理用户登录会话",
              children: []
            }
          ]
        }
      ]
    },
    {
      id: 8,
      level: 1,
      code: "F002",
      title: "数据分析模块",
      description: "提供数据统计、分析和报表功能",
      children: [
        {
          id: 9,
          level: 2,
          code: "F002.01",
          title: "数据统计",
          description: "统计各类业务数据指标",
          children: [
            {
              id: 10,
              level: 3,
              code: "F002.01.01",
              title: "用户行为统计",
              description: "统计用户访问、操作等行为数据",
              children: []
            }
          ]
        }
      ]
    }
  ]
}

const handleItemClick = async (item) => {
  // 记录用户行为
  await recordUserBehavior({
    action: 'view',
    itemType: item.type,
    itemId: item.itemId,
    context: { source: 'recommendation' }
  })
  
  viewDetails(item)
}

const viewDetails = async (item) => {
  try {
    // 记录查看详情行为
    await recordUserBehavior({
      action: 'view_detail',
      itemType: item.type,
      itemId: item.itemId,
      context: { source: 'recommendation', itemTitle: item.title }
    })
    
    // 显示详情对话框
    ElMessageBox.alert(
      `<div style="text-align: left;">
        <h3>${item.title}</h3>
        <p><strong>类型:</strong> ${getTypeLabel(item.type)}</p>
        <p><strong>分类:</strong> ${item.category || '未分类'}</p>
        <p><strong>置信度:</strong> ${Math.round((item.score || 0) * 100)}%</p>
        <p><strong>推荐理由:</strong> ${item.reason || '系统推荐'}</p>
        <p><strong>描述:</strong></p>
        <div style="max-height: 200px; overflow-y: auto; border: 1px solid #eee; padding: 10px; background: #f9f9f9;">
          ${item.description || '暂无描述'}
        </div>
        ${item.metadata?.improvements ? 
          `<p><strong>改进建议:</strong></p>
           <ul>${item.metadata.improvements.map(imp => `<li>${imp}</li>`).join('')}</ul>` : ''}
        ${item.metadata?.riskWarnings ? 
          `<p><strong>风险提示:</strong></p>
           <ul>${item.metadata.riskWarnings.map(warn => `<li style="color: #f56c6c;">${warn}</li>`).join('')}</ul>` : ''}
      </div>`,
      '推荐详情',
      {
        confirmButtonText: '关闭',
        dangerouslyUseHTMLString: true,
        customClass: 'recommendation-details-dialog'
      }
    )
  } catch (error) {
    console.error('查看详情失败:', error)
    ElMessage.error('查看详情失败')
  }
}

const editItem = async (item) => {
  try {
    // 记录编辑行为
    await recordUserBehavior({
      action: 'edit',
      itemType: item.type,
      itemId: item.itemId,
      context: { source: 'recommendation', itemTitle: item.title }
    })
    
    // 显示编辑对话框
    ElMessageBox.prompt(
      `编辑推荐内容：`,
      `编辑 - ${item.title}`,
      {
        confirmButtonText: '保存',
        cancelButtonText: '取消',
        inputType: 'textarea',
        inputValue: item.description || '',
        inputPlaceholder: '请输入修改后的内容...',
        inputValidator: (value) => {
          if (!value || value.trim() === '') {
            return '内容不能为空'
          }
          return true
        }
      }
    ).then(async ({ value }) => {
      if (value && value.trim() !== item.description) {
        // 如果是AI推荐，可以直接采纳修改后的内容
        if (isAIRecommendation.value) {
          await adoptRecommendation({
            projectId: 1,
            recommendationId: item.originalId || item.itemId, // 使用原始ID
            title: item.title,
            content: value.trim(),
            category: item.category,
            confidence: item.score,
            feedback: '用户编辑后采纳'
          })
          ElMessage.success('编辑内容已保存并采纳到知识库')
        } else {
          // 非AI推荐，记录编辑反馈
          await recordUserBehavior({
            action: 'edit_content',
            itemType: item.type,
            itemId: item.itemId,
            context: { 
              source: 'recommendation', 
              originalContent: item.description,
              editedContent: value.trim() 
            }
          })
          ElMessage.success('编辑内容已记录')
        }
      } else {
        ElMessage.info('内容未发生变化')
      }
    }).catch(() => {
      // 用户取消编辑
    })
    
  } catch (error) {
    console.error('编辑失败:', error)
    ElMessage.error('编辑失败: ' + (error.message || '未知错误'))
  }
}

const provideFeedback = (item) => {
  feedbackDialog.currentItem = item
  feedbackDialog.form.score = 5
  feedbackDialog.form.comment = ''
  feedbackDialog.visible = true
}

const submitFeedback = async () => {
  try {
    await submitRecommendationFeedback({
      itemType: feedbackDialog.currentItem.type,
      itemId: feedbackDialog.currentItem.itemId,
      score: feedbackDialog.form.score,
      comment: feedbackDialog.form.comment
    })
    
    ElMessage.success('反馈提交成功，感谢您的参与！')
    feedbackDialog.visible = false
  } catch (error) {
    ElMessage.error('反馈提交失败')
  }
}

const adoptAIRecommendation = async (item) => {
  try {
    await ElMessageBox.confirm(
      `您确定要采纳推荐"${item.title}"吗？采纳后将加入知识库供其他用户使用。`,
      '采纳AI推荐',
      {
        confirmButtonText: '确定采纳',
        cancelButtonText: '取消',
        type: 'info',
      }
    )
    
    // 先记录采纳行为（采纳操作之前）
    try {
      await recordUserBehavior({
        action: 'adopt_attempt',
        itemType: item.type,
        itemId: item.itemId,
        context: { 
          source: 'ai_recommendation', 
          confidence: item.score,
          originalId: item.originalId || item.itemId,
          title: item.title
        }
      })
    } catch (behaviorError) {
      console.warn('记录用户行为失败:', behaviorError)
      // 不阻断采纳流程
    }
    
    // 调用采纳API
    await adoptRecommendation({
      projectId: 1, // 暂时使用固定项目ID
      recommendationId: item.originalId || item.itemId, // 使用原始ID进行采纳
      title: item.title,
      content: item.description,
      category: item.category,
      confidence: item.score,
      feedback: '用户主动采纳'
    })
    
    ElMessage.success('推荐已采纳并加入知识库，感谢您的参与！')
    
    // 采纳成功后记录最终行为
    try {
      await recordUserBehavior({
        action: 'adopt_success',
        itemType: item.type,
        itemId: item.itemId,
        context: { 
          source: 'ai_recommendation', 
          confidence: item.score,
          originalId: item.originalId || item.itemId,
          status: 'success'
        }
      })
    } catch (behaviorError) {
      console.warn('记录采纳成功行为失败:', behaviorError)
      // 不影响用户体验
    }
    
  } catch (error) {
    if (error !== 'cancel') {
      console.error('采纳推荐失败:', error)
      ElMessage.error('采纳推荐失败: ' + (error.response?.data?.msg || error.message || '未知错误'))
      
      // 记录采纳失败行为
      try {
        await recordUserBehavior({
          action: 'adopt_failed',
          itemType: item.type,
          itemId: item.itemId,
          context: { 
            source: 'ai_recommendation', 
            error: error.message || '未知错误',
            originalId: item.originalId || item.itemId
          }
        })
      } catch (behaviorError) {
        console.warn('记录采纳失败行为失败:', behaviorError)
      }
    }
  }
}

const getTypeColor = (type) => {
  const colors = {
    requirement: 'primary',
    template: 'success',
    agent: 'warning',
    project: 'info',
    knowledge: 'danger'
  }
  return colors[type] || 'info'
}

const getTypeLabel = (type) => {
  const labels = {
    requirement: '需求',
    template: '模板',
    agent: 'Agent',
    project: '项目',
    knowledge: '知识'
  }
  return labels[type] || type
}
</script>

<style lang="scss" scoped>
.recommendation-container {
  padding: 20px;
  background-color: #f5f7fa;
  min-height: 100vh;
}

.page-header {
  text-align: center;
  margin-bottom: 30px;
  
  h1 {
    color: #2c3e50;
    font-size: 28px;
    margin-bottom: 10px;
  }
  
  p {
    color: #7f8c8d;
    font-size: 14px;
  }
}

.filter-section {
  margin-bottom: 20px;
  
  .filter-card {
    border-radius: 8px;
  }
}

.recommendations-section {
  .content-card {
    border-radius: 8px;
    
    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      
      .header-right {
        display: flex;
        align-items: center;
        gap: 10px;
      }
      
      .algorithm-tag {
        color: #909399;
        font-size: 12px;
      }
    }
  }
}

.recommendations-list {
  .welcome-prompt {
    text-align: center;
    padding: 60px 20px;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    border-radius: 12px;
    margin-bottom: 20px;
    
    .prompt-content {
      color: white;
      
      h3 {
        font-size: 24px;
        margin: 20px 0 10px 0;
      }
      
      p {
        font-size: 16px;
        margin-bottom: 30px;
        opacity: 0.9;
      }
      
      .el-button {
        padding: 12px 30px;
        font-size: 16px;
      }
    }
  }
  
  .recommendation-item {
    margin-bottom: 16px;
    
    .item-card {
      cursor: pointer;
      transition: all 0.3s ease;
      border-radius: 8px;
      
      &:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
      }
      
      .item-header {
        display: flex;
        justify-content: space-between;
        align-items: flex-start;
        margin-bottom: 12px;
        
        .item-title {
          display: flex;
          align-items: center;
          gap: 10px;
          
          h3 {
            margin: 0;
            font-size: 16px;
            color: #2c3e50;
          }
        }
      }
      
      .item-content {
        margin-bottom: 12px;
        
        .item-description {
          color: #5a6c7d;
          font-size: 14px;
          line-height: 1.5;
          margin-bottom: 8px;
          display: -webkit-box;
          -webkit-line-clamp: 2;
          -webkit-box-orient: vertical;
          overflow: hidden;
        }
        
        .item-meta {
          display: flex;
          gap: 6px;
          flex-wrap: wrap;
          
          .tag-item {
            margin-right: 4px;
          }
        }
      }
      
      .item-footer {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding-top: 8px;
        border-top: 1px solid #ebeef5;
        
        .reason {
          color: #909399;
          font-size: 12px;
          font-style: italic;
        }
        
        .item-actions {
          display: flex;
          gap: 8px;
          
          :deep(.el-button--text) {
            color: #409EFF !important;
            padding: 0 8px !important;
            
            &:hover {
              color: #66b1ff !important;
              text-decoration: underline;
            }
          }
        }
      }
    }
  }
}

:deep(.el-rate__text) {
  font-size: 12px;
}

/* 推荐详情对话框样式 */
:deep(.recommendation-details-dialog) {
  .el-message-box__message {
    max-height: 400px;
    overflow-y: auto;
  }
  
  .el-message-box__content {
    padding: 20px;
  }
  
  h3 {
    margin-top: 0;
    color: #2c3e50;
    border-bottom: 2px solid #409EFF;
    padding-bottom: 8px;
  }
  
  ul {
    padding-left: 20px;
    margin: 8px 0;
  }
  
  li {
    margin: 4px 0;
    line-height: 1.5;
  }
}
</style> 