<template>
  <div class="style-test-page">
    <div class="page-header">
      <h1>样式隔离测试页面</h1>
      <p>这个页面用于测试Smart-NESMA样式是否正确隔离，不影响原框架样式</p>
    </div>

    <div class="test-sections">
      <!-- 原框架样式区域 -->
      <div class="test-section">
        <h2>原框架样式（应保持不变）</h2>
        
        <div class="form-demo">
          <el-button type="primary">原框架主按钮</el-button>
          <el-button type="success">原框架成功按钮</el-button>
          <el-button type="warning">原框架警告按钮</el-button>
          <el-button type="danger">原框架危险按钮</el-button>
        </div>

        <div class="card-demo">
          <el-card class="demo-card" header="原框架卡片">
            <p>这是一个使用原框架样式的卡片，应该保持Element Plus默认样式。</p>
            <el-input v-model="testInput" placeholder="原框架输入框" />
          </el-card>
        </div>

        <div class="table-demo">
          <el-table :data="testData" class="demo-table">
            <el-table-column prop="name" label="姓名" />
            <el-table-column prop="role" label="角色" />
            <el-table-column prop="status" label="状态" />
          </el-table>
        </div>
      </div>

      <!-- Smart-NESMA样式区域 -->
      <div class="test-section smart-nesma-page">
        <h2>Smart-NESMA样式（应使用自定义样式）</h2>
        
        <div class="form-demo">
          <el-button type="primary">Smart-NESMA主按钮</el-button>
          <el-button type="success">Smart-NESMA成功按钮</el-button>
          <el-button type="warning">Smart-NESMA警告按钮</el-button>
          <el-button type="danger">Smart-NESMA危险按钮</el-button>
        </div>

        <div class="card-demo">
          <el-card class="demo-card" header="Smart-NESMA卡片">
            <p>这是一个使用Smart-NESMA样式的卡片，应该有渐变背景和圆角。</p>
            <el-input v-model="testInput2" placeholder="Smart-NESMA输入框" />
          </el-card>
        </div>

        <div class="table-demo">
          <el-table :data="testData" class="demo-table">
            <el-table-column prop="name" label="姓名" />
            <el-table-column prop="role" label="角色" />
            <el-table-column prop="status" label="状态" />
          </el-table>
        </div>

        <div class="timeline-demo">
          <h3>时间线组件</h3>
          <el-timeline>
            <el-timeline-item
              v-for="(activity, index) in activities"
              :key="index"
              :timestamp="activity.timestamp"
            >
              {{ activity.content }}
            </el-timeline-item>
          </el-timeline>
        </div>

        <div class="progress-demo">
          <h3>进度条组件</h3>
          <el-progress :percentage="50" class="demo-progress" />
          <el-progress :percentage="80" status="success" />
          <el-progress :percentage="60" status="warning" />
          <el-progress :percentage="30" status="exception" />
        </div>
      </div>
    </div>

    <div class="navigation-links">
      <el-button @click="$router.push('/nesma')">返回NESMA主页</el-button>
      <el-button type="primary" @click="$router.push('/nesma/dashboard-unified')">
        查看统一仪表盘
      </el-button>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'

const testInput = ref('')
const testInput2 = ref('')

const testData = reactive([
  { name: '张三', role: '开发者', status: '在线' },
  { name: '李四', role: '测试员', status: '忙碌' },
  { name: '王五', role: '产品经理', status: '离线' }
])

const activities = reactive([
  {
    content: '完成项目需求分析',
    timestamp: '2025-01-11 10:00'
  },
  {
    content: '开始开发阶段',
    timestamp: '2025-01-11 14:00'
  },
  {
    content: '完成功能测试',
    timestamp: '2025-01-11 16:30'
  }
])
</script>

<style lang="scss" scoped>
.style-test-page {
  padding: 24px;
  background: #f5f5f5;
  min-height: 100vh;
}

.page-header {
  background: white;
  padding: 24px;
  border-radius: 8px;
  margin-bottom: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

  h1 {
    margin: 0 0 8px 0;
    color: #333;
    font-size: 24px;
  }

  p {
    margin: 0;
    color: #666;
    font-size: 14px;
  }
}

.test-sections {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
  margin-bottom: 24px;

  @media (max-width: 768px) {
    grid-template-columns: 1fr;
  }
}

.test-section {
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

  h2 {
    margin: 0 0 20px 0;
    color: #333;
    font-size: 18px;
    border-bottom: 2px solid #eee;
    padding-bottom: 8px;
  }

  h3 {
    margin: 20px 0 12px 0;
    color: #555;
    font-size: 16px;
  }
}

.form-demo {
  margin-bottom: 20px;

  .el-button {
    margin-right: 12px;
    margin-bottom: 8px;
  }
}

.card-demo {
  margin-bottom: 20px;

  .demo-card {
    max-width: 100%;
  }

  .el-input {
    margin-top: 12px;
  }
}

.table-demo {
  margin-bottom: 20px;

  .demo-table {
    width: 100%;
  }
}

.timeline-demo {
  margin-bottom: 20px;
}

.progress-demo {
  .el-progress {
    margin-bottom: 12px;
  }
}

.navigation-links {
  text-align: center;
  background: white;
  padding: 24px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);

  .el-button {
    margin: 0 12px;
  }
}
</style>