<template>
  <div>
    <warning-bar title="管理3/4级层次化需求结构，支持Excel导入与AI智能分析" />
    <div class="gva-table-box">
      <div class="gva-btn-list">
        <!-- 基础操作 -->
        <el-button type="primary" icon="plus" @click="handleCreate">
          新建需求
        </el-button>
        <el-button type="success" icon="upload" @click="handleBatchImport">
          批量导入
        </el-button>
        
        <!-- 视图切换 -->
        <el-divider direction="vertical" />
        <span class="view-label">视图：</span>
        <el-button-group>
          <el-button 
            :type="viewMode === 'table' ? 'primary' : ''" 
            @click="switchViewMode('table')"
            :loading="viewSwitching && viewMode !== 'table'"
            :disabled="viewSwitching"
          >
            <el-icon><Grid /></el-icon>
          </el-button>
          <!-- <el-button 
            :type="viewMode === 'tree' ? 'primary' : ''" 
            @click="switchViewMode('tree')"
            :loading="viewSwitching && viewMode !== 'tree'"
            :disabled="viewSwitching"
          >
            <el-icon><List /></el-icon>
          </el-button> -->
        </el-button-group>
        
        <!-- 批量分析操作 -->
        <el-divider direction="vertical" />
        <span class="batch-label">批量操作：</span>
        <!-- 批量生成操作 -->
        <el-button type="info" icon="operation" @click="handleBatchL4Generation" :disabled="!hasSelectedL3Requirements">
          批量生成L4
        </el-button>
        <el-button type="info" icon="coordinate" @click="handleBatchMermaidGeneration" :disabled="!hasSelectedL4Requirements">
          批量流程图
        </el-button>
        
        <!-- L4任务管理 - 始终显示 -->
        <el-divider direction="vertical" />
        <span class="task-label">任务管理：</span>
        <el-button 
          type="warning" 
          @click="handleOpenL4TaskManagement" 
          style="margin-right: 8px; font-weight: bold;"
        >
          <el-icon style="margin-right: 4px;"><List /></el-icon>
          L4任务管理
        </el-button>
        <el-button 
          type="info" 
          @click="handleOpenMermaidTaskManagement" 
          style="margin-right: 8px; font-weight: bold;"
        >
          <el-icon style="margin-right: 4px;"><Coordinate /></el-icon>
          流程图任务管理
        </el-button>
        
        <!-- 刷新按钮 -->
        <el-divider direction="vertical" />
        <el-button @click="refreshData" :loading="loading">
          <el-icon><Refresh /></el-icon>
          刷新
        </el-button>
      </div>

      <!-- 项目上下文选择 -->
      <div class="gva-search-box">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="当前项目">
            <el-select v-model="currentProject" placeholder="选择项目" @change="handleProjectChange">
              <el-option
                v-for="project in projects"
                :key="project.ID || project.id"
                :label="project.name"
                :value="project.ID || project.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="项目周期">
            <el-select v-model="currentCycle" placeholder="选择周期" @change="handleCycleChange" :disabled="!currentProject">
              <el-option
                v-for="cycle in cycles"
                :key="cycle.ID"
                :label="cycle.name"
                :value="cycle.ID"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="需求版本">
            <el-select v-model="currentVersion" placeholder="选择版本" @change="handleVersionChange" :disabled="!currentCycle">
              <el-option
                v-for="version in versions"
                :key="version.ID"
                :label="`${version.version}`"
                :value="version.ID"
              />
            </el-select>
          </el-form-item>
        </el-form>
      </div>

      <!-- 高级查询条件 -->
      <div class="gva-search-box">
        <el-form :inline="true" :model="searchForm">
          <el-form-item label="需求标题">
            <el-input 
              v-model="searchForm.title" 
              placeholder="请输入需求标题" 
              clearable 
              style="width: 200px"
            />
          </el-form-item>
          <el-form-item label="需求描述">
            <el-input 
              v-model="searchForm.description" 
              placeholder="请输入描述关键词" 
              clearable 
              style="width: 200px"
            />
          </el-form-item>
          <el-form-item label="需求层级">
            <el-select v-model="searchForm.level" placeholder="选择层级" clearable style="width: 120px">
              <el-option label="L1 - 一级模块" :value="1" />
              <el-option label="L2 - 二级模块" :value="2" />
              <el-option label="L3 - 三级模块" :value="3" />
              <el-option label="L4 - 功能点" :value="4" />
            </el-select>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="searchForm.status" placeholder="选择状态" clearable style="width: 120px">
              <el-option label="待处理" value="pending" />
              <el-option label="进行中" value="in_progress" />
              <el-option label="已完成" value="completed" />
              <el-option label="已取消" value="cancelled" />
            </el-select>
          </el-form-item>
          <el-form-item label="功能类型">
            <el-select v-model="searchForm.functionType" placeholder="选择功能类型" clearable style="width: 120px">
              <el-option label="EI" value="EI" />
              <el-option label="EO" value="EO" />
              <el-option label="EQ" value="EQ" />
              <el-option label="ILF" value="ILF" />
              <el-option label="EIF" value="EIF" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" icon="search" @click="handleSearch">查询</el-button>
            <el-button icon="refresh" @click="resetSearch">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 需求表格 -->
      <transition name="view-fade" mode="out-in">
        <div key="loading" v-if="viewSwitching" class="view-switching-overlay">
          <el-skeleton :rows="8" animated />
          <div class="switching-text">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span v-if="viewMode === 'tree'">正在构建树形结构...</span>
            <span v-else>正在加载表格数据...</span>
          </div>
        </div>
        
        <el-table
          v-else-if="viewMode === 'table'"
          key="table-view"
          :data="tableData"
          style="width: 100%"
          tooltip-effect="dark"
          row-key="id"
          :tree-props="{ children: 'children', hasChildren: 'hasChildren' }"
          @selection-change="handleSelectionChange"
          ref="requirementTable"
          v-loading="loading"
          element-loading-text="正在加载需求数据..."
        >
        <el-table-column 
          type="selection" 
          width="55" 
          :reserve-selection="true"
          :selectable="(row) => true"
        />
        <el-table-column align="left" label="需求编号" prop="code" width="120" />
        <el-table-column align="left" label="需求名称" prop="title" min-width="200" />
        <el-table-column align="left" label="层级" width="80">
          <template #default="scope">
            <el-tag :type="getLevelType(scope.row.level)" size="small">
              L{{ scope.row.level }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="功能类型" prop="functionType" width="120">
          <template #default="scope">
            <el-tag v-if="scope.row.functionType" :type="getFunctionTypeColor(scope.row.functionType)" size="small">
              {{ scope.row.functionType }}
            </el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column align="left" label="状态" width="100">
          <template #default="scope">
            <el-tag :type="getStatusType(scope.row.status)" size="small">
              {{ getStatusText(scope.row.status) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column align="left" label="操作" min-width="300">
          <template #default="scope">
            <!-- 基础操作：查看、编辑、删除 -->
            <el-button type="primary" link icon="view" @click="handleView(scope.row)" size="small">
              查看
            </el-button>
            <el-button type="primary" link icon="edit" @click="handleEdit(scope.row)" size="small">
              编辑
            </el-button>
            
            <!-- 层级特定操作 -->
            <template v-if="scope.row.level === 2">
              <!-- 二级需求：L3需求分析 -->
              <el-button type="warning" link icon="search" @click="handleL3Analysis(scope.row)" size="small">
                L3分析
              </el-button>
            </template>
            
            <template v-if="scope.row.level === 3">
              <!-- 三级需求：L4功能点生成 -->
              <el-button type="success" link icon="operation" @click="handleL4Generation(scope.row)" size="small">
                生成L4
              </el-button>
            </template>
            
            <template v-if="scope.row.level === 4">
              <!-- 四级功能点：Mermaid流程图生成 -->
              <el-button type="info" link icon="coordinate" @click="handleMermaidGeneration(scope.row)" size="small">
                流程图
              </el-button>
            </template>
            
            <!-- 通用AI分析 -->
            <el-button type="success" link icon="magic-stick" @click="handleAnalyze(scope.row)" size="small" v-if="scope.row.level !== 1">
              AI分析
            </el-button>
            
            <!-- 删除操作 -->
            <el-button type="danger" link icon="delete" @click="handleDelete(scope.row)" size="small">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>
      
      <!-- 树形视图 -->
      <RequirementTree
        v-else-if="viewMode === 'tree'"
        key="tree-view"
        :tree-data="convertToTreeData(tableData)"
        :search-conditions="searchForm"
        :filtered-count="total"
        :loading="loading"
        :default-expand-level="3"
        @create="handleCreate"
        @import="handleBatchImport"
        @refresh="refreshData"
        @view="handleView"
        @edit="handleEdit"
        @create-child="handleCreateChild"
        @delete="handleDelete"
        @ai-analysis="handleAnalyze"
        @l3-analysis="handleL3Analysis"
        @l4-generation="handleL4Generation"
        @mermaid-generation="handleMermaidGeneration"
        @node-click="handleNodeClick"
        @node-drop="handleNodeDrop"
        @copy="handleCopy"
        @move="handleMove"
        @export="handleExport"
        @tree-rendered="handleTreeRendered"
      />
      </transition>
      
      <!-- 分页 - 只在表格视图显示 -->
      <div v-if="viewMode === 'table'" class="gva-pagination">
        <el-pagination
          :current-page="page"
          :page-size="pageSize"
          :page-sizes="[10, 30, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @current-change="handleCurrentChange"
          @size-change="handleSizeChange"
        />
      </div>
      
      <!-- 树形视图数据统计 -->
      <div v-else-if="viewMode === 'tree'" class="tree-summary">
        <el-alert
          :title="`当前显示 ${total} 条需求记录`"
          type="info"
          :closable="false"
          show-icon
        >
          <template #default>
            <span>树形结构已展示所有层级数据，可使用搜索条件筛选特定需求</span>
          </template>
        </el-alert>
      </div>
    </div>

    <!-- 需求详情对话框 -->
    <RequirementDetailDialog 
      v-model="viewDialogVisible"
      :requirement-id="currentRequirementId"
      @edit="handleEditFromDetail"
      @viewChild="handleViewChild"
    />

    <!-- 需求编辑表单 -->
    <RequirementForm
      v-model="editDialogVisible"
      :form-data="currentRequirement"
      :is-edit="editDialogMode === 'edit'"
      :projects="projects"
      :cycles="cycles"
      :versions="versions"
      :parent-options="parentOptions"
      :loading="formLoading"
      @submit="handleFormSubmit"
      @cancel="handleFormCancel"
    />

    <!-- AI分析进度对话框 -->
    <el-dialog 
      v-model="analysisDialogVisible" 
      :title="showAnalysisResult ? 'AI分析结果' : 'AI需求分析'"
      :width="showAnalysisResult ? '1200px' : '700px'"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :before-close="() => handleCloseDialog('analysis')"
    >
      <!-- 分析进度阶段 -->
      <div v-if="!showAnalysisResult" class="enhanced-analysis-progress">
        <!-- 需求信息 -->
        <div class="requirement-header">
          <div class="req-title">
            <el-icon class="req-icon"><Document /></el-icon>
            <span>{{ currentRequirement.title }}</span>
          </div>
          <el-tag type="info" size="small">{{ currentRequirement.code }}</el-tag>
        </div>

        <!-- 动画进度区域 -->
        <div class="animated-progress-section">
          <!-- 主进度条 -->
          <div class="main-progress">
            <el-progress 
              :percentage="analysisStatus.progress" 
              :status="getProgressStatus()"
              :stroke-width="18"
              :show-text="false"
            />
            <div class="progress-info">
              <span class="progress-percent">{{ analysisStatus.progress }}%</span>
              <span class="stage-name">{{ analysisStatus.stageDesc || '准备中...' }}</span>
            </div>
          </div>

          <!-- 动画图标区域 -->
          <div class="animation-section">
            <div :class="['stage-icon-container', analysisStatus.animationType]">
              <el-icon 
                :class="['stage-icon', getStageIconClass()]" 
                :style="{ color: getStageColor() }"
              >
                <component :is="getStageIcon()" />
              </el-icon>
            </div>
            
            <!-- 状态文本 -->
            <div class="status-text">
              <p class="primary-text">{{ analysisStatus.statusText || '正在处理...' }}</p>
              <p class="secondary-text" v-if="analysisStatus.animationData.estimatedRemaining > 0">
                预计剩余时间: {{ analysisStatus.animationData.estimatedRemaining }} 秒
              </p>
            </div>
          </div>

          <!-- 阶段历史时间线 -->
          <div class="stage-timeline" v-if="analysisStatus.stageHistory.length > 0">
            <h4>分析步骤</h4>
            <el-timeline>
              <el-timeline-item 
                v-for="(stage, index) in analysisStatus.stageHistory" 
                :key="index"
                :type="getTimelineType(stage.status)"
                :hollow="stage.status !== 'completed'"
                :timestamp="formatTime(stage.startTime)"
              >
                <div class="timeline-content">
                  <span class="stage-desc">{{ stage.description }}</span>
                  <el-tag 
                    v-if="stage.duration" 
                    size="small" 
                    type="info"
                  >
                    {{ stage.duration }}ms
                  </el-tag>
                </div>
              </el-timeline-item>
            </el-timeline>
          </div>
        </div>

        <!-- 分析特性说明 -->
        <div class="analysis-features">
          <div class="feature-grid">
            <div class="feature-item">
              <el-icon><MagicStick /></el-icon>
              <span>智能描述优化</span>
            </div>
            <div class="feature-item">
              <el-icon><Collection /></el-icon>
              <span>NESMA功能分类</span>
            </div>
            <div class="feature-item">
              <el-icon><Timer /></el-icon>
              <span>复杂度评估</span>
            </div>
            <div class="feature-item">
              <el-icon><FolderOpened /></el-icon>
              <span>知识库匹配</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 分析结果显示 -->
      <div v-else class="analysis-result-section">
        <div class="result-header">
          <el-result
            icon="success"
            title="分析完成"
            sub-title="AI已为您的需求生成了优化建议"
          />
        </div>

        <!-- 改进亮点概览 - 紧凑显示 -->
        <div class="improvements-highlight" v-if="analysisResult">
          <el-card class="highlight-card" shadow="never">
            <template #header>
              <div class="highlight-header">
                <div class="highlight-title">
                  <el-icon class="highlight-icon"><Star /></el-icon>
                  <span>改进亮点概览</span>
                </div>
                <el-tag type="success">
                  置信度 {{ (analysisResult.confidence * 100).toFixed(0) }}%
                </el-tag>
              </div>
            </template>
            
            <div class="highlight-content">
              <el-row :gutter="16">
                <el-col :span="8" v-if="analysisResult.optimized.title !== analysisResult.original.title">
                  <div class="highlight-item">
                    <el-icon class="item-icon success"><EditPen /></el-icon>
                    <div class="item-text">
                      <span class="item-title">需求标题优化</span>
                      <span class="item-desc">标题更加准确明确</span>
                    </div>
                  </div>
                </el-col>
                <el-col :span="8" v-if="analysisResult.optimized.functionType && !analysisResult.original.functionType">
                  <div class="highlight-item">
                    <el-icon class="item-icon info"><Collection /></el-icon>
                    <div class="item-text">
                      <span class="item-title">NESMA功能分类</span>
                      <span class="item-desc">确定为：{{ analysisResult.optimized.functionType }}</span>
                    </div>
                  </div>
                </el-col>
                <el-col :span="8" v-if="analysisResult.optimized.description && !analysisResult.original.description">
                  <div class="highlight-item">
                    <el-icon class="item-icon warning"><Document /></el-icon>
                    <div class="item-text">
                      <span class="item-title">功能描述补充</span>
                      <span class="item-desc">补充了详细的功能描述</span>
                    </div>
                  </div>
                </el-col>
              </el-row>
              
              <!-- 采纳建议 -->
              <div class="adoption-recommendation" :class="analysisResult.confidence >= 0.7 ? 'recommend' : 'caution'">
                <div class="adoption-content">
                  <el-icon v-if="analysisResult.confidence >= 0.7" class="adoption-icon success"><Check /></el-icon>
                  <el-icon v-else class="adoption-icon warning"><Warning /></el-icon>
                  <div class="adoption-text">
                    <span v-if="analysisResult.confidence >= 0.7">
                      建议采纳此优化，AI分析置信度高达 {{ (analysisResult.confidence * 100).toFixed(0) }}%
                    </span>
                    <span v-else>
                      建议人工审核，AI分析置信度为 {{ (analysisResult.confidence * 100).toFixed(0) }}%
                    </span>
                  </div>
                </div>
              </div>
            </div>
          </el-card>
        </div>

        <!-- 详细对比结果 -->
        <div class="result-comparison" v-if="analysisResult">
          <el-tabs v-model="activeResultTab" type="card">
            <!-- 对比结果 -->
            <el-tab-pane label="详细对比" name="comparison">
              <div class="comparison-content">
                <!-- 并排对比显示 -->
                <el-row :gutter="16">
                  <el-col :span="11">
                    <el-card class="comparison-card original" shadow="never">
                      <template #header>
                        <div class="card-header">
                          <span>
                            <el-icon><DocumentCopy /></el-icon>
                            原始需求
                          </span>
                          <el-tag type="info" size="small">当前版本</el-tag>
                        </div>
                      </template>
                      
                      <div class="card-content">
                        <el-form label-width="80px" size="small" disabled>
                          <el-form-item label="需求标题">
                            <el-input :value="analysisResult.original.title" readonly />
                          </el-form-item>
                          <el-form-item label="需求描述">
                            <el-input 
                              :value="analysisResult.original.description || '暂无描述'" 
                              type="textarea" 
                              :rows="3"
                              readonly
                            />
                          </el-form-item>
                          <el-form-item label="功能类型">
                            <el-input :value="analysisResult.original.functionType || '未分类'" readonly />
                          </el-form-item>
                          <el-form-item label="业务价值">
                            <el-input :value="analysisResult.original.businessValue || '未评估'" readonly />
                          </el-form-item>
                        </el-form>
                      </div>
                    </el-card>
                  </el-col>
                  
                  <!-- 箭头指示 -->
                  <el-col :span="2" class="arrow-col">
                    <div class="comparison-arrow">
                      <el-icon class="arrow-icon"><ArrowRight /></el-icon>
                      <span class="arrow-text">AI优化</span>
                    </div>
                  </el-col>
                  
                  <el-col :span="11">
                    <el-card class="comparison-card optimized" shadow="never">
                      <template #header>
                        <div class="card-header">
                          <span>
                            <el-icon><MagicStick /></el-icon>
                            AI优化建议
                          </span>
                          <el-tag type="success" size="small">优化版本</el-tag>
                        </div>
                      </template>
                      
                      <div class="card-content">
                        <el-form :model="editableResult" label-width="80px" size="small">
                          <el-form-item label="优化标题">
                            <el-input v-model="editableResult.title" type="text" />
                          </el-form-item>
                          <el-form-item label="优化描述">
                            <el-input 
                              v-model="editableResult.description" 
                              type="textarea" 
                              :rows="3"
                              placeholder="请输入详细的功能描述"
                            />
                          </el-form-item>
                          <el-form-item label="功能类型">
                            <el-select v-model="editableResult.functionType" placeholder="选择功能类型">
                              <el-option label="EI - 外部输入" value="EI" />
                              <el-option label="EO - 外部输出" value="EO" />
                              <el-option label="EQ - 外部查询" value="EQ" />
                              <el-option label="ILF - 内部逻辑文件" value="ILF" />
                              <el-option label="EIF - 外部接口文件" value="EIF" />
                            </el-select>
                          </el-form-item>
                          <el-form-item label="业务价值">
                            <el-input 
                              v-model="editableResult.businessValue" 
                              placeholder="请评估业务价值"
                            />
                          </el-form-item>
                        </el-form>
                      </div>
                    </el-card>
                  </el-col>
                </el-row>

                <!-- 分析摘要 - 优化显示 -->
                <el-card class="analysis-summary" shadow="never" style="margin-top: 16px;">
                  <template #header>
                    <div class="summary-header">
                      <span>
                        <el-icon><DataAnalysis /></el-icon>
                        分析摘要
                      </span>
                      <el-tag type="success" size="small">
                        <el-icon><Check /></el-icon>
                        分析完成
                      </el-tag>
                    </div>
                  </template>
                  
                  <div class="summary-content">
                    <el-row :gutter="16">
                      <el-col :span="6">
                        <div class="summary-item confidence">
                          <div class="item-icon">
                            <el-icon><TrendCharts /></el-icon>
                          </div>
                          <div class="item-info">
                            <div class="item-value">{{ (analysisResult.confidence * 100).toFixed(0) }}%</div>
                            <div class="item-label">AI置信度</div>
                          </div>
                        </div>
                      </el-col>
                      <el-col :span="6">
                        <div class="summary-item business">
                          <div class="item-icon">
                            <el-icon><Coin /></el-icon>
                          </div>
                          <div class="item-info">
                            <div class="item-value">{{ editableResult.businessValue || '未评估' }}</div>
                            <div class="item-label">业务价值</div>
                          </div>
                        </div>
                      </el-col>
                      <el-col :span="6">
                        <div class="summary-item complexity">
                          <div class="item-icon">
                            <el-icon><Setting /></el-icon>
                          </div>
                          <div class="item-info">
                            <div class="item-value">{{ editableResult.complexity || '未评估' }}</div>
                            <div class="item-label">复杂度</div>
                          </div>
                        </div>
                      </el-col>
                      <el-col :span="6">
                        <div class="summary-item improvement">
                          <div class="item-icon">
                            <el-icon><Trophy /></el-icon>
                          </div>
                          <div class="item-info">
                            <div class="item-value">{{ calculateImprovementCount() }}</div>
                            <div class="item-label">改进项数</div>
                          </div>
                        </div>
                      </el-col>
                    </el-row>
                    
                    <div class="analysis-note" v-if="analysisResult.analysisNote" style="margin-top: 16px;">
                      <el-alert 
                        :title="analysisResult.analysisNote" 
                        type="info" 
                        :closable="false"
                        show-icon
                      >
                        <template #title>
                          <div class="alert-content">
                            <strong>分析说明：</strong>
                            {{ analysisResult.analysisNote }}
                          </div>
                        </template>
                      </el-alert>
                    </div>
                    
                    <!-- 处理时间统计 -->
                    <div class="processing-info" style="margin-top: 12px;">
                      <el-row :gutter="16">
                        <el-col :span="8">
                          <div class="info-item">
                            <el-icon class="info-icon"><Timer /></el-icon>
                            <span class="info-text">处理时间: {{ getProcessingTime() }}</span>
                          </div>
                        </el-col>
                        <el-col :span="8">
                          <div class="info-item">
                            <el-icon class="info-icon"><User /></el-icon>
                            <span class="info-text">分析模式: AI智能分析</span>
                          </div>
                        </el-col>
                        <el-col :span="8">
                          <div class="info-item">
                            <el-icon class="info-icon"><Calendar /></el-icon>
                            <span class="info-text">分析时间: {{ formatCurrentTime() }}</span>
                          </div>
                        </el-col>
                      </el-row>
                    </div>
                  </div>
                </el-card>
              </div>
            </el-tab-pane>

            <!-- 改进详情 -->
            <el-tab-pane label="改进详情" name="improvements">
              <div class="improvements-content">
                <el-timeline>
                  <el-timeline-item 
                    v-if="analysisResult.optimized.title !== analysisResult.original.title"
                    icon="EditPen" 
                    type="success"
                    timestamp="需求标题优化"
                  >
                    <el-card shadow="never">
                      <div class="improvement-detail">
                        <h4>需求标题优化</h4>
                        <div class="before-after">
                          <div class="change-item">
                            <span class="label">原标题：</span>
                            <span class="original-text">{{ analysisResult.original.title }}</span>
                          </div>
                          <div class="change-item">
                            <span class="label">优化后：</span>
                            <span class="optimized-text">{{ editableResult.title }}</span>
                          </div>
                        </div>
                        <el-tag type="success" size="small">
                          <el-icon><Check /></el-icon>
                          标题更加准确明确，便于理解和管理
                        </el-tag>
                      </div>
                    </el-card>
                  </el-timeline-item>

                  <el-timeline-item 
                    v-if="editableResult.description && !analysisResult.original.description"
                    icon="Document" 
                    type="primary"
                    timestamp="功能描述补充"
                  >
                    <el-card shadow="never">
                      <div class="improvement-detail">
                        <h4>功能描述补充</h4>
                        <div class="new-content">
                          <span class="label">新增描述：</span>
                          <div class="optimized-text">{{ editableResult.description }}</div>
                        </div>
                        <el-tag type="success" size="small">
                          <el-icon><Check /></el-icon>
                          补充了详细的功能描述，提升需求清晰度
                        </el-tag>
                      </div>
                    </el-card>
                  </el-timeline-item>

                  <el-timeline-item 
                    v-if="editableResult.functionType && !analysisResult.original.functionType"
                    icon="Collection" 
                    type="warning"
                    timestamp="NESMA功能分类"
                  >
                    <el-card shadow="never">
                      <div class="improvement-detail">
                        <h4>NESMA功能分类</h4>
                        <div class="classification">
                          <span class="label">确定类型：</span>
                          <el-tag type="success">{{ editableResult.functionType }}</el-tag>
                        </div>
                        <el-tag type="success" size="small">
                          <el-icon><Check /></el-icon>
                          明确了NESMA功能类型，便于工作量评估
                        </el-tag>
                      </div>
                    </el-card>
                  </el-timeline-item>

                  <el-timeline-item 
                    v-if="editableResult.businessValue && !analysisResult.original.businessValue"
                    icon="Coin" 
                    type="info"
                    timestamp="业务价值评估"
                  >
                    <el-card shadow="never">
                      <div class="improvement-detail">
                        <h4>业务价值评估</h4>
                        <div class="value-assessment">
                          <span class="label">价值评级：</span>
                          <el-tag type="warning">{{ editableResult.businessValue }}</el-tag>
                        </div>
                        <el-tag type="success" size="small">
                          <el-icon><Check /></el-icon>
                          明确了业务价值和效益，便于优先级决策
                        </el-tag>
                      </div>
                    </el-card>
                  </el-timeline-item>
                </el-timeline>
              </div>
            </el-tab-pane>
          </el-tabs>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button 
            v-if="!showAnalysisResult && analysisStatus.status === 'running'"
            @click="handleCloseDialog('analysis')"
          >
            后台运行
          </el-button>
          
          <el-button 
            v-if="showAnalysisResult"
            @click="handleCloseDialog('analysis')"
          >
            取消
          </el-button>
          
          <el-button 
            v-if="showAnalysisResult"
            type="primary" 
            @click="handleApplyAnalysis"
          >
            采纳优化
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- L4功能点生成进度对话框 -->
    <el-dialog 
      v-model="l4GenerationDialogVisible" 
      :title="l4GenerationStatus.status === 'completed' ? 'L4功能点生成完成' : 'L4功能点生成进度'"
      :width="l4GenerationStatus.status === 'completed' ? '1400px' : '700px'"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :before-close="closeL4GenerationDialog"
    >
      <!-- 生成进度阶段 -->
      <div v-if="l4GenerationStatus.status !== 'completed'" class="enhanced-l4-generation-progress">
        <!-- 任务信息 -->
        <div class="requirement-header">
          <div class="req-title">
            <el-icon class="req-icon"><Cpu /></el-icon>
            <span>正在为 {{ l4GenerationTask?.totalCount || 0 }} 个三级需求生成L4功能点</span>
          </div>
          <el-tag :type="l4GenerationStatus.status === 'running' ? 'primary' : 'info'" size="small">
            {{ l4GenerationStatus.status === 'running' ? '处理中' : '准备中' }}
          </el-tag>
        </div>

        <!-- 动画进度区域 -->
        <div class="animated-progress-section">
          <!-- 主进度条 -->
          <div class="main-progress">
            <el-progress 
              :percentage="l4GenerationStatus.progress" 
              :status="getL4ProgressStatus()"
              :stroke-width="18"
              :show-text="false"
            />
            <div class="progress-info">
              <span class="progress-percent">{{ l4GenerationStatus.progress }}%</span>
              <span class="stage-name">{{ l4GenerationStatus.stageDesc || '准备中...' }}</span>
            </div>
          </div>

          <!-- 动画图标区域 -->
          <div class="animation-section">
            <div :class="['stage-icon-container', l4GenerationStatus.animationType || 'brain']">
              <el-icon 
                :class="['stage-icon', getL4StageIconClass()]" 
                :style="{ color: getL4StageColor() }"
              >
                <component :is="getL4StageIcon()" />
              </el-icon>
            </div>
            
            <!-- 状态文本 -->
            <div class="status-text">
              <p class="primary-text">{{ l4GenerationStatus.statusText || '正在处理...' }}</p>
              <p class="secondary-text" v-if="l4GenerationStatus.animationData.estimatedRemaining > 0">
                预计剩余时间: {{ l4GenerationStatus.animationData.estimatedRemaining }} 秒
              </p>
            </div>
          </div>

          <!-- 阶段历史时间线 -->
          <div class="stage-timeline" v-if="l4GenerationStatus.stageHistory.length > 0">
            <h4>生成步骤</h4>
            <el-timeline>
              <el-timeline-item 
                v-for="(stage, index) in l4GenerationStatus.stageHistory" 
                :key="index"
                :type="getTimelineType(stage.status)"
                :hollow="stage.status !== 'completed'"
                :timestamp="formatTime(stage.startTime)"
              >
                <div class="timeline-content">
                  <span class="stage-desc">{{ stage.description }}</span>
                  <el-tag 
                    v-if="stage.duration" 
                    size="small" 
                    type="info"
                  >
                    {{ stage.duration }}ms
                  </el-tag>
                </div>
              </el-timeline-item>
            </el-timeline>
          </div>
        </div>
        <!-- L4生成特性说明 -->
        <div class="analysis-features">
          <div class="feature-grid">
            <div class="feature-item">
              <el-icon><MagicStick /></el-icon>
              <span>AI智能生成</span>
            </div>
            <div class="feature-item">
              <el-icon><Collection /></el-icon>
              <span>功能点分类</span>
            </div>
            <div class="feature-item">
              <el-icon><Timer /></el-icon>
              <span>复杂度评估</span>
            </div>
            <div class="feature-item">
              <el-icon><FolderOpened /></el-icon>
              <span>知识库匹配</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 生成结果阶段 -->
      <div v-else class="l4-generation-result">
        <!-- 生成概况 -->
        <div class="result-summary">
          <el-row :gutter="20">
            <el-col :span="6">
              <el-statistic title="处理的L3需求" :value="l4GenerationResult?.task?.totalCount || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="生成的L4功能点" :value="getCreatedL4Count()" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="成功率" :value="calculateSuccessRate()" suffix="%" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="消耗时间" :value="formatDuration(l4GenerationResult?.task?.duration)" />
            </el-col>
          </el-row>
        </div>

        <!-- 错误信息展示 -->
        <div v-if="hasGenerationErrors()" class="error-summary">
          <el-alert 
            title="生成过程中遇到错误" 
            type="warning" 
            :closable="false"
            show-icon
          >
            <template #default>
              <div class="error-list">
                <div v-for="(error, index) in getGenerationErrors()" :key="index" class="error-item">
                  {{ error }}
                </div>
              </div>
            </template>
          </el-alert>
        </div>

        <!-- L4功能点列表 -->
        <div v-if="getCreatedL4Count() > 0" class="result-table">
          <div class="table-header">
            <h4>生成的L4功能点列表</h4>
            <div class="table-actions">
              <el-button @click="selectAllL4Requirements">全选</el-button>
              <el-button @click="selectNoneL4Requirements">取消全选</el-button>
            </div>
          </div>
          
          <el-table 
            :data="getL4SuggestionsList()" 
            style="width: 100%"
            max-height="400"
            @selection-change="handleL4SelectionChange"
          >
            <el-table-column type="selection" width="55" />
            <el-table-column prop="code" label="编码" width="120" />
            <el-table-column prop="title" label="功能点名称" min-width="200">
              <template #default="{ row }">
                <div class="requirement-title">
                  <span>{{ row.title }}</span>
                  <el-tag v-if="row.source === 'ai_generated'" type="success" size="small">AI生成</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="functionType" label="功能类型" width="100">
              <template #default="{ row }">
                <el-tag :type="getFunctionTypeColor(row.functionType)" size="small">
                  {{ row.functionType }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="complexity" label="复杂度" width="100">
              <template #default="{ row }">
                <el-tag :type="getComplexityColor(row.complexity)" size="small">
                  {{ row.complexity }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="functionPoints" label="功能点数" width="100" />
            <el-table-column prop="confidence" label="置信度" width="100">
              <template #default="{ row }">
                <el-tag :type="getConfidenceColor(row.confidence)" size="small">
                  {{ Math.round(row.confidence * 100) }}%
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="priority" label="优先级" width="100">
              <template #default="{ row }">
                <el-tag :type="getPriorityColor(row.priority)" size="small">
                  P{{ row.priority }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="l3RequirementTitle" label="来源L3需求" min-width="200">
              <template #default="{ row }">
                <div class="l3-requirement-info">
                  <span class="l3-title">{{ row.l3RequirementTitle }}</span>
                  <el-tag size="small" type="info">L3</el-tag>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="{ row }">
                <el-button type="primary" size="small" @click="viewL4Detail(row)">
                  查看详情
                </el-button>
                <el-button 
                  type="success" 
                  size="small" 
                  @click="adoptL4Suggestion(row)"
                  :disabled="isL4SuggestionAdopted(row)"
                  :loading="false"
                >
                  {{ isL4SuggestionAdopted(row) ? '已采纳' : '采纳' }}
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>

        <!-- 无生成结果时的提示 -->
        <div v-else class="no-results">
          <el-empty description="本次没有成功生成L4功能点">
            <template #image>
              <el-icon size="80" color="#d9d9d9"><Warning /></el-icon>
            </template>
            <template #description>
              <p>由于以下原因，本次L4功能点生成未成功：</p>
              <ul class="failure-reasons">
                <li v-for="(error, index) in getGenerationErrors()" :key="index">
                  {{ error }}
                </li>
              </ul>
            </template>
          </el-empty>
        </div>
      </div>

      <!-- 对话框底部按钮 -->
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeL4GenerationDialog">
            {{ l4GenerationStatus === 'completed' ? '关闭' : '取消' }}
          </el-button>
          <el-button 
            v-if="l4GenerationStatus === 'completed'" 
            type="primary" 
            @click="confirmSelectedL4Requirements"
            :disabled="selectedL4Requirements.length === 0"
          >
            确认选中的功能点 ({{ selectedL4Requirements.length }})
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 批量导入对话框 -->
    <ImportDialog
      v-model="importDialogVisible"
      :project-id="currentProject"
      :cycle-id="currentCycle"
      :version-id="currentVersion"
      @import-completed="handleImportCompleted"
    />
    
    <!-- L4任务管理对话框 -->
    <L4TaskManagementDialog
      v-model="l4TaskManagementVisible"
      :project="currentProjectObject"
      :cycle="currentCycleObject"
      :version="currentVersionObject"
      @view-progress="handleViewL4TaskProgress"
      @view-result="handleViewL4TaskResult"
      @close="handleCloseL4TaskManagement"
    />
    
    <!-- L4任务进度追踪对话框 -->
    <L4TaskProgressTracker
      v-model="l4TaskProgressVisible"
      :task-id="currentL4TaskId"
      :auto-refresh="true"
      @task-completed="handleL4TaskCompleted"
      @task-failed="handleL4TaskFailed"
      @task-cancelled="handleL4TaskCancelled"
      @view-result="handleViewL4TaskResult"
    />
    
    <!-- Mermaid流程图生成进度对话框 -->
    <el-dialog
      v-model="showMermaidProgressDialog"
      :title="mermaidGenerationStatus.status === 'completed' ? '流程图生成完成' : '流程图生成进度'"
      :width="mermaidGenerationStatus.status === 'completed' ? '1000px' : '700px'"
      :close-on-click-modal="false"
      :close-on-press-escape="false"
      :before-close="closeMermaidProgressDialog"
    >
      <!-- 生成进度阶段 -->
      <div v-if="mermaidGenerationStatus.status !== 'completed'" class="enhanced-mermaid-generation-progress">
        <!-- 任务信息 -->
        <div class="requirement-header">
          <div class="req-title">
            <el-icon class="req-icon"><Coordinate /></el-icon>
            <span>正在为 {{ mermaidGenerationStatus.totalCount || 0 }} 个L4需求生成Mermaid流程图</span>
          </div>
          <el-tag :type="mermaidGenerationStatus.status === 'running' ? 'primary' : 'info'" size="small">
            {{ mermaidGenerationStatus.status === 'running' ? '生成中' : '准备中' }}
          </el-tag>
        </div>

        <!-- 动画进度区域 -->
        <div class="animated-progress-section">
          <!-- 主进度条 -->
          <div class="main-progress">
            <el-progress 
              :percentage="mermaidGenerationStatus.progress" 
              :status="getMermaidProgressStatus()"
              :stroke-width="18"
              :show-text="false"
            />
            <div class="progress-info">
              <span class="progress-percent">{{ mermaidGenerationStatus.progress }}%</span>
              <span class="stage-name">{{ getMermaidStageDesc() }}</span>
            </div>
          </div>

          <!-- 动画图标区域 -->
          <div class="animation-section">
            <div :class="['stage-icon-container', getMermaidAnimationType()]">
              <el-icon 
                :class="['stage-icon', getMermaidStageIconClass()]" 
                :style="{ color: getMermaidStageColor() }"
              >
                <component :is="getMermaidStageIcon()" />
              </el-icon>
            </div>
            
            <!-- 状态文本 -->
            <div class="status-text">
              <p class="primary-text">{{ getMermaidStatusText() }}</p>
              <p class="secondary-text" v-if="mermaidGenerationStatus.estimatedRemaining > 0">
                预计剩余时间: {{ mermaidGenerationStatus.estimatedRemaining }} 秒
              </p>
            </div>
          </div>

          <!-- 生成统计 -->
          <div class="generation-statistics" v-if="mermaidGenerationStatus.totalCount > 0">
            <el-row :gutter="16">
              <el-col :span="6">
                <div class="stat-item">
                  <el-icon class="stat-icon"><Document /></el-icon>
                  <div class="stat-info">
                    <div class="stat-value">{{ mermaidGenerationStatus.totalCount || 0 }}</div>
                    <div class="stat-label">总需求数</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-item">
                  <el-icon class="stat-icon"><Check /></el-icon>
                  <div class="stat-info">
                    <div class="stat-value">{{ mermaidGenerationStatus.successCount || 0 }}</div>
                    <div class="stat-label">已完成</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-item">
                  <el-icon class="stat-icon"><Warning /></el-icon>
                  <div class="stat-info">
                    <div class="stat-value">{{ mermaidGenerationStatus.failedCount || 0 }}</div>
                    <div class="stat-label">失败</div>
                  </div>
                </div>
              </el-col>
              <el-col :span="6">
                <div class="stat-item">
                  <el-icon class="stat-icon"><Loading /></el-icon>
                  <div class="stat-info">
                    <div class="stat-value">{{ mermaidGenerationStatus.processedCount || 0 }}</div>
                    <div class="stat-label">已处理</div>
                  </div>
                </div>
              </el-col>
            </el-row>
          </div>
        </div>

        <!-- 流程图生成特性说明 -->
        <div class="analysis-features">
          <div class="feature-grid">
            <div class="feature-item">
              <el-icon><Coordinate /></el-icon>
              <span>智能流程图</span>
            </div>
            <div class="feature-item">
              <el-icon><MagicStick /></el-icon>
              <span>AI生成</span>
            </div>
            <div class="feature-item">
              <el-icon><Timer /></el-icon>
              <span>自动布局</span>
            </div>
            <div class="feature-item">
              <el-icon><Setting /></el-icon>
              <span>详细注释</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 生成结果阶段 -->
      <div v-else class="mermaid-generation-result">
        <!-- 生成概况 -->
        <div class="result-summary">
          <el-result
            icon="success"
            title="流程图生成完成"
            sub-title="AI已为您的L4功能点生成了Mermaid流程图"
          />
          <el-row :gutter="20">
            <el-col :span="6">
              <el-statistic title="处理的L4需求" :value="mermaidGenerationStatus.totalCount || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="生成成功" :value="mermaidGenerationStatus.successCount || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="生成失败" :value="mermaidGenerationStatus.failedCount || 0" />
            </el-col>
            <el-col :span="6">
              <el-statistic title="成功率" :value="calculateMermaidSuccessRate()" suffix="%" />
            </el-col>
          </el-row>
        </div>

        <!-- 错误信息展示 -->
        <div v-if="mermaidGenerationStatus.failedCount > 0" class="error-summary">
          <el-alert 
            title="部分流程图生成失败" 
            type="warning" 
            :closable="false"
            show-icon
          >
            <template #default>
              <div class="error-details">
                <p>有 {{ mermaidGenerationStatus.failedCount }} 个流程图生成失败，可能原因：</p>
                <ul>
                  <li>需求描述信息不足</li>
                  <li>网络连接问题</li>
                  <li>AI服务暂时不可用</li>
                </ul>
                <p>建议：您可以稍后重试失败的需求，或通过"流程图任务管理"查看详细错误信息。</p>
              </div>
            </template>
          </el-alert>
        </div>

        <!-- 生成摘要 -->
        <div v-if="mermaidGenerationStatus.summary" class="generation-summary">
          <el-card shadow="never">
            <template #header>
              <div class="summary-header">
                <span>
                  <el-icon><DataAnalysis /></el-icon>
                  生成摘要
                </span>
              </div>
            </template>
            <p>{{ mermaidGenerationStatus.summary }}</p>
          </el-card>
        </div>
      </div>
      
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="closeMermaidProgressDialog">
            {{ mermaidGenerationStatus.status === 'completed' ? '关闭' : '后台运行' }}
          </el-button>
          <el-button 
            v-if="mermaidGenerationStatus.status === 'completed'"
            type="primary" 
            @click="handleOpenMermaidTaskManagement"
          >
            查看任务详情
          </el-button>
        </div>
      </template>
    </el-dialog>

    <!-- Mermaid流程图任务管理对话框 -->
    <MermaidTaskManagementDialog
      v-model="showMermaidTaskManagementDialog"
      :project="currentProjectObject"
      :cycle="currentCycleObject"
      :version="currentVersionObject"
      @view-progress="handleViewMermaidTaskProgress"
      @view-result="handleViewMermaidTaskResult"
      @close="handleCloseMermaidTaskManagement"
    />
    
    <!-- L4功能点详情对话框 -->
    <el-dialog
      v-model="l4DetailDialogVisible"
      title="L4功能点详情"
      width="900px"
      :close-on-click-modal="false"
    >
      <div class="l4-detail-content" v-if="currentL4Detail">
        <!-- 基本信息 -->
        <el-card class="detail-card" shadow="never">
          <template #header>
            <div class="card-header">
              <span class="header-title">
                <el-icon><Document /></el-icon>
                基本信息
              </span>
              <el-tag :type="getFunctionTypeColor(currentL4Detail.functionType)" size="small">
                {{ currentL4Detail.functionType }}
              </el-tag>
            </div>
          </template>
          
          <el-descriptions :column="2" border>
            <el-descriptions-item label="功能点名称" :span="2">
              <span class="detail-title">{{ currentL4Detail.title }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="功能描述" :span="2">
              <div class="detail-description">{{ currentL4Detail.description || '暂无描述' }}</div>
            </el-descriptions-item>
            <el-descriptions-item label="功能类型">
              <el-tag :type="getFunctionTypeColor(currentL4Detail.functionType)" size="small">
                {{ currentL4Detail.functionType }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="复杂度">
              <el-tag :type="getComplexityColor(currentL4Detail.complexity)" size="small">
                {{ currentL4Detail.complexity }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="功能点数">
              <span class="function-points">{{ currentL4Detail.functionPoints }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="置信度">
              <el-tag :type="getConfidenceColor(currentL4Detail.confidence)" size="small">
                {{ Math.round(currentL4Detail.confidence * 100) }}%
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="优先级">
              <el-tag :type="getPriorityColor(currentL4Detail.priority)" size="small">
                P{{ currentL4Detail.priority }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="业务价值">
              <span>{{ currentL4Detail.businessValue || '未评估' }}</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 来源信息 -->
        <el-card class="detail-card" shadow="never" style="margin-top: 16px;">
          <template #header>
            <div class="card-header">
              <span class="header-title">
                <el-icon><FolderOpened /></el-icon>
                来源信息
              </span>
            </div>
          </template>
          
          <el-descriptions :column="1" border>
            <el-descriptions-item label="来源L3需求">
              <div class="source-requirement">
                <span class="source-title">{{ currentL4Detail.l3RequirementTitle }}</span>
                <el-tag size="small" type="info">L3</el-tag>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="生成原因">
              <div class="generation-reason">{{ currentL4Detail.generationReason || '暂无' }}</div>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <!-- 验收标准 -->
        <el-card class="detail-card" shadow="never" style="margin-top: 16px;" v-if="currentL4Detail.acceptanceCriteria">
          <template #header>
            <div class="card-header">
              <span class="header-title">
                <el-icon><Check /></el-icon>
                验收标准
              </span>
            </div>
          </template>
          
          <div class="acceptance-criteria">
            {{ currentL4Detail.acceptanceCriteria }}
          </div>
        </el-card>

        <!-- 相关知识 -->
        <el-card class="detail-card" shadow="never" style="margin-top: 16px;" v-if="currentL4Detail.relatedKnowledge && currentL4Detail.relatedKnowledge.length > 0">
          <template #header>
            <div class="card-header">
              <span class="header-title">
                <el-icon><Collection /></el-icon>
                相关知识
              </span>
            </div>
          </template>
          
          <div class="related-knowledge">
            <el-tag 
              v-for="(knowledge, index) in currentL4Detail.relatedKnowledge" 
              :key="index"
              size="small"
              type="info"
              class="knowledge-tag"
            >
              {{ knowledge }}
            </el-tag>
          </div>
        </el-card>
      </div>

      <template #footer>
        <div class="dialog-footer">
          <el-button @click="l4DetailDialogVisible = false">关闭</el-button>
          <el-button 
            type="success" 
            @click="adoptL4Suggestion(currentL4Detail)"
            :disabled="isL4SuggestionAdopted(currentL4Detail)"
          >
            {{ isL4SuggestionAdopted(currentL4Detail) ? '已采纳' : '采纳该功能点' }}
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch, nextTick } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useRoute, useRouter } from 'vue-router'
import {
  Document, Plus, Upload, Search, Refresh, View, Edit, Delete,
  MagicStick, FolderOpened, Timer, Collection, Check, Warning,
  Star, EditPen, DocumentCopy, ArrowRight, TrendCharts, Coin,
  DataAnalysis, Trophy, Reading, Setting, User, Calendar,
  Coordinate, Operation, Grid, List, Loading, Cpu
} from '@element-plus/icons-vue'

// 导入warning-bar组件
import WarningBar from '@/components/warningBar/warningBar.vue'

// 导入需求对话框组件
import RequirementDetailDialog from './components/RequirementDetailDialog.vue'
import RequirementForm from './components/form/RequirementForm.vue'
import ImportDialog from './components/import/ImportDialog.vue'
import RequirementTree from './components/tree/RequirementTree.vue'
import L4TaskManagementDialog from './components/L4TaskManagementDialog.vue'
import L4TaskProgressTracker from './components/L4TaskProgressTracker.vue'
import MermaidTaskManagementDialog from './components/MermaidTaskManagementDialog.vue'

// API导入
import {
  getNesmaRequirementList,
  createNesmaRequirement,
  updateNesmaRequirement,
  deleteNesmaRequirement,
  getNesmaProjectList,
  applyRequirementAnalysisResult
} from '@/api/nesma'

// 导入项目周期和版本相关API
import { getProjectCycles } from '@/api/projectCycle'
import { getRequirementVersions } from '@/api/nesma/requirementVersion'

// 导入需求相关API
import { 
  analyzeRequirement,
  getRequirementAnalysisProgress,
  getRequirementAnalysisResult,
  applyAnalysisRecommendation,
  startIntelligentAnalysis,
  getAnalysisProgress,
  getAnalysisResult,
  // L3分析相关API
  analyzeLevel3Requirements,
  applyOptimizationSuggestion,
  createExpansionRequirement,
  batchApplyLevel3Optimizations,
  // L4生成相关API
  generateLevel4Requirements,
  batchCreateLevel4Requirements,
  // 异步L4生成API
  generateLevel4Async,
  getLevel4GenerationTaskProgress,
  getLevel4GenerationTaskResult,
  // Mermaid生成相关API
  generateMermaidDiagrams,
  generateMermaidDiagramsAsync,
  getMermaidGenerationTaskProgress,
  getMermaidGenerationTaskResult,
  batchGenerateMermaidDiagrams,
  applyMermaidDiagram
} from '@/api/nesma'

// 初始化route和router
const route = useRoute()
const router = useRouter()

// 响应式数据
const loading = ref(false)
const formLoading = ref(false)

// 项目上下文
const currentProject = ref(null)
const currentCycle = ref(null)
const currentVersion = ref(null)
const projects = ref([])
const cycles = ref([])
const versions = ref([])

// 表格数据
const tableData = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

// 搜索表单
const searchForm = ref({
  title: '',
  description: '',
  level: null,
  status: '',
  functionType: ''
})

// 对话框状态
const viewDialogVisible = ref(false)
const editDialogVisible = ref(false)
const analysisDialogVisible = ref(false)
const importDialogVisible = ref(false)
const l4GenerationDialogVisible = ref(false)
const l4DetailDialogVisible = ref(false) // L4详情对话框状态
const l4TaskManagementVisible = ref(false) // L4任务管理对话框状态
const l4TaskProgressVisible = ref(false) // L4任务进度追踪对话框状态

// L4任务管理相关状态
const currentL4TaskId = ref(null) // 当前查看的L4任务ID
const currentProjectObject = ref(null) // 当前项目对象
const currentCycleObject = ref(null) // 当前周期对象
const currentVersionObject = ref(null) // 当前版本对象

// 当前操作的需求
const currentRequirement = ref({})
const currentRequirementId = ref(null)
const currentL4Detail = ref({}) // 当前L4详情数据
const editDialogMode = ref('edit')
const parentOptions = ref([])
const analysisTask = ref(null)
const analysisProgress = ref(0)

// L4生成任务状态
const l4GenerationTask = ref(null)
const l4GenerationProgress = ref(0)
const l4GenerationStatus = ref({
  taskId: null,
  status: 'pending',
  progress: 0,
  currentStage: '',
  stageDesc: '',
  animationType: '',
  errorMsg: '',
  stageHistory: [],
  animationData: {},
  statusText: ''
})
const l4GenerationResult = ref(null)
const selectedL4Requirements = ref([])
const adoptedL4Requirements = ref(new Set()) // 跟踪已采纳的L4需求

// Mermaid流程图生成任务状态
const showMermaidProgressDialog = ref(false)
const currentMermaidTaskId = ref(null)
const mermaidGenerationStatus = ref({
  taskId: null,
  status: 'pending',
  progress: 0,
  totalCount: 0,
  processedCount: 0,
  successCount: 0,
  failedCount: 0,
  errorMsg: '',
  summary: '',
  // 增强的状态字段
  currentStage: 'initializing',
  estimatedRemaining: 0
})

// 增强的分析状态
const analysisStatus = ref({
  taskId: null,
  status: 'pending',
  progress: 0,
  currentStage: '',
  stageDesc: '',
  animationType: '',
  errorMsg: '',
  stageHistory: [],
  animationData: {},
  statusText: ''
})

// 选中需求状态管理
const selectedRequirements = ref([])
const selectedAnalysisLevel = ref('2,3') // 默认L2+L3分析

// 视图模式管理
const viewMode = ref('table') // 'table' 或 'tree'
const viewSwitching = ref(false) // 视图切换专用加载状态

// 计算属性：是否有选中的需求
const hasSelectedRequirements = computed(() => {
  return selectedRequirements.value.length > 0
})

// 计算属性：是否有选中的L3需求
const hasSelectedL3Requirements = computed(() => {
  return selectedRequirements.value.some(req => req.level === 3)
})

// 计算属性：是否有选中的L4需求  
const hasSelectedL4Requirements = computed(() => {
  return selectedRequirements.value.some(req => req.level === 4)
})

const analysisResult = ref(null)
const showAnalysisResult = ref(false)
const activeResultTab = ref('comparison')
const editableResult = ref({
  title: '',
  description: '',
  functionType: '',
  businessValue: '',
  complexity: ''
})

// 视图切换方法（优化版本 - 解决树形渲染阻塞问题）
const switchViewMode = async (mode) => {
  if (viewMode.value === mode) return
  
  console.log('切换视图模式:', viewMode.value, '->', mode)
  
  try {
    viewSwitching.value = true
    
    if (mode === 'tree') {
      ElMessage.info('正在切换到树形视图，构建层级结构...')
      
      // 设置超时保护
      setTreeRenderTimeout()
      
      viewMode.value = 'tree'
      await nextTick()
      
      // 加载树形数据
      await loadRequirements()
      
      console.log('树形数据加载完成，等待树形组件渲染完成事件...')
      // 不关闭 viewSwitching，等待树形组件的 tree-rendered 事件
      
    } else {
      ElMessage.info('正在切换到表格视图...')
      viewMode.value = 'table'
      await loadRequirements()
      ElMessage.success('表格视图加载完成')
      viewSwitching.value = false
    }
  } catch (error) {
    console.error('视图切换失败:', error)
    ElMessage.error('视图切换失败：' + error.message)
    viewSwitching.value = false
    clearTreeRenderTimeout()
  }
}

// 树形数据转换（性能优化版本 - 支持大数据量，防止主线程阻塞）
const convertToTreeData = (data) => {
  if (!data || data.length === 0) return []
  
  const startTime = performance.now()
  console.log(`开始转换树形数据，数据量: ${data.length}`)
  
  // 如果数据量很大，使用同步批处理避免过度优化导致的复杂性
  // 对于大多数场景，现代浏览器可以处理2000条以内的同步操作
  if (data.length > 2000) {
    console.warn(`数据量过大 (${data.length})，可能会影响性能`)
  }
  
  // 性能优化：使用 Map 提高查找效率
  const nodeMap = new Map()
  const rootNodes = []
  
  // 第一次遍历：创建所有节点的映射
  data.forEach(item => {
    const node = {
      ...item,
      children: [],
      // 确保 id 字段存在且一致
      id: item.id || item.ID,
      // 父级ID标准化
      parentId: item.parent_id || item.parentId || item.ParentID
    }
    nodeMap.set(node.id, node)
  })
  
  // 第二次遍历：构建父子关系
  nodeMap.forEach(node => {
    if (node.parentId && nodeMap.has(node.parentId)) {
      // 有父节点，添加到父节点的children中
      const parent = nodeMap.get(node.parentId)
      parent.children.push(node)
    } else {
      // 没有父节点或父节点不存在，视为根节点
      rootNodes.push(node)
    }
  })
  
  // 批量排序，减少递归深度
  const sortNodes = (nodes) => {
    nodes.sort((a, b) => {
      // 首先按层级排序
      if (a.level !== b.level) {
        return a.level - b.level
      }
      // 然后按编码排序
      if (a.code && b.code) {
        return a.code.localeCompare(b.code)
      }
      // 最后按ID排序
      return a.id - b.id
    })
    
    // 递归排序子节点（限制递归深度）
    nodes.forEach(node => {
      if (node.children && node.children.length > 0) {
        sortNodes(node.children)
      }
    })
  }
  
  // 同步处理排序，现代浏览器能很好处理4层嵌套的排序
  sortNodes(rootNodes)
  
  const endTime = performance.now()
  const duration = endTime - startTime
  console.log(`树形数据转换完成: ${data.length} 条扁平数据 -> ${rootNodes.length} 个根节点，耗时: ${duration.toFixed(2)}ms`)
  
  // 如果转换时间超过50ms，警告用户可能的性能问题
  if (duration > 50) {
    console.warn(`树形转换耗时较长: ${duration.toFixed(2)}ms，建议优化数据结构或启用分页`)
  }
  
  return rootNodes
}

// 异步树形数据转换 - 避免阻塞主线程
const convertToTreeDataAsync = async (data) => {
  if (!data || data.length === 0) return []
  
  const startTime = performance.now()
  console.log(`开始异步转换树形数据，数据量: ${data.length}`)
  
  return new Promise((resolve) => {
    // 使用 setTimeout 将数据转换移到下一个事件循环
    setTimeout(() => {
      const nodeMap = new Map()
      const rootNodes = []
      
      // 批量处理数据，避免一次性处理太多
      const batchSize = 100
      let currentIndex = 0
      
      const processBatch = () => {
        const endIndex = Math.min(currentIndex + batchSize, data.length)
        
        // 处理当前批次的数据
        for (let i = currentIndex; i < endIndex; i++) {
          const item = data[i]
          const node = {
            ...item,
            children: [],
            id: item.id || item.ID,
            parentId: item.parent_id || item.parentId || item.ParentID
          }
          nodeMap.set(node.id, node)
        }
        
        currentIndex = endIndex
        
        if (currentIndex < data.length) {
          // 还有数据需要处理，继续下一批次
          setTimeout(processBatch, 0)
        } else {
          // 所有数据处理完成，构建树结构
          buildTreeStructure()
        }
      }
      
      const buildTreeStructure = () => {
        // 构建父子关系
        nodeMap.forEach(node => {
          if (node.parentId && nodeMap.has(node.parentId)) {
            const parent = nodeMap.get(node.parentId)
            parent.children.push(node)
          } else {
            rootNodes.push(node)
          }
        })
        
        // 排序
        const sortNodes = (nodes) => {
          nodes.sort((a, b) => {
            if (a.level !== b.level) return a.level - b.level
            if (a.code && b.code) return a.code.localeCompare(b.code)
            return a.id - b.id
          })
          
          nodes.forEach(node => {
            if (node.children && node.children.length > 0) {
              sortNodes(node.children)
            }
          })
        }
        
        sortNodes(rootNodes)
        
        const endTime = performance.now()
        console.log(`异步树形数据转换完成: ${data.length} 条扁平数据 -> ${rootNodes.length} 个根节点，耗时: ${(endTime - startTime).toFixed(2)}ms`)
        
        resolve(rootNodes)
      }
      
      // 开始处理第一批数据
      processBatch()
    }, 0)
  })
}

// 树形视图事件处理
const handleCreateChild = (parentNode) => {
  console.log('创建子需求:', parentNode)
  currentRequirement.value = {
    projectId: currentProject.value,
    cycleId: currentCycle.value,
    versionId: currentVersion.value,
    parentId: parentNode.id,
    level: (parentNode.level || 1) + 1
  }
  editDialogMode.value = 'create'
  editDialogVisible.value = true
}

const handleNodeClick = (node) => {
  console.log('节点点击:', node)
  // 可以实现节点点击后的逻辑，比如显示详情
}

const handleNodeDrop = (dropData) => {
  console.log('节点拖拽:', dropData)
  // 实现拖拽重新排序或改变层级关系的逻辑
  ElMessage.info('拖拽重新排序功能开发中...')
}

const handleCopy = (node) => {
  console.log('复制节点:', node)
  ElMessage.info('复制功能开发中...')
}

const handleMove = (node) => {
  console.log('移动节点:', node)
  ElMessage.info('移动功能开发中...')
}

const handleExport = (node) => {
  console.log('导出节点:', node)
  ElMessage.info('导出功能开发中...')
}

// 处理树形视图渲染完成事件
const handleTreeRendered = () => {
  console.log('=== 收到树形组件渲染完成事件 ===')
  
  // 清理超时机制
  clearTreeRenderTimeout()
  
  // 确保在下一帧关闭加载动画
  requestAnimationFrame(() => {
    if (viewSwitching.value) {
      console.log('关闭树形视图加载动画')
      ElMessage.success('树形视图加载完成，已展开到三级节点')
      viewSwitching.value = false
    } else {
      console.log('加载动画已经关闭，无需重复操作')
    }
  })
}

// 树形渲染超时处理
let treeRenderTimeout = null

const clearTreeRenderTimeout = () => {
  if (treeRenderTimeout) {
    clearTimeout(treeRenderTimeout)
    treeRenderTimeout = null
  }
}

const setTreeRenderTimeout = () => {
  clearTreeRenderTimeout()
  treeRenderTimeout = setTimeout(() => {
    console.warn('树形渲染超时，强制关闭加载动画')
    if (viewSwitching.value) {
      ElMessage.warning('树形视图加载超时，但已基本完成')
      viewSwitching.value = false
    }
  }, 8000) // 增加超时时间到8秒，给复杂渲染更多时间
}

// 方法实现
const handleCreate = () => {
  if (!currentProject.value || !currentCycle.value || !currentVersion.value) {
    ElMessage.warning('请先选择项目、周期和版本')
    return
  }
  
  currentRequirement.value = {
    projectId: currentProject.value,
    cycleId: currentCycle.value,
    versionId: currentVersion.value
  }
  editDialogMode.value = 'create'
  editDialogVisible.value = true
}

const handleEdit = (requirement) => {
  console.log('编辑需求:', requirement)
  currentRequirement.value = requirement
  editDialogMode.value = 'edit'
  editDialogVisible.value = true
}

const handleView = (requirement) => {
  console.log('查看需求:', requirement)
  currentRequirementId.value = requirement.id || requirement.ID
  viewDialogVisible.value = true
}

const handleDelete = async (requirement) => {
  try {
    await ElMessageBox.confirm('确定要删除此需求吗？此操作不可恢复！', '确认删除', {
      type: 'error',
      confirmButtonText: '确定删除',
      cancelButtonText: '取消'
    })
    
    console.log('删除需求，ID:', requirement.id || requirement.ID)
    const requirementId = requirement.id || requirement.ID
    
    const response = await deleteNesmaRequirement(requirementId)
    if (response.code === 0) {
      ElMessage.success('需求删除成功')
      await loadRequirements()
    } else {
      throw new Error(response.msg || '删除失败')
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除需求失败:', error)
      ElMessage.error('删除需求失败：' + error.message)
    }
  }
}

// L3需求分析
const handleL3Analysis = async (requirement) => {
  try {
    console.log('L3需求分析:', requirement)
    
    const analysisData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      requirement_ids: [requirement.id || requirement.ID],
      analysis_type: 'comprehensive',
      analysis_depth: 'standard',
      use_knowledge_base: true,
      generate_expansions: true,
      include_nesma_scoring: true
    }

    ElMessage.info('正在启动L3需求分析...')
    const response = await analyzeLevel3Requirements(analysisData)
    
    if (response.code === 0) {
      ElMessage.success('L3需求分析完成')
      // 刷新需求列表以显示更新后的状态
      await loadRequirements()
    } else {
      throw new Error(response.msg || 'L3分析失败')
    }

  } catch (error) {
    console.error('L3需求分析失败:', error)
    ElMessage.error('L3分析失败: ' + (error.message || '未知错误'))
  }
}

// L4功能点生成 - 统一使用异步方法
const handleL4Generation = async (requirement) => {
  try {
    console.log('单个L4功能点生成:', requirement)
    
    // 使用和批量生成相同的异步方法
    const generateData = {
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      versionId: currentVersion.value,
      l3RequirementIds: [requirement.id || requirement.ID],
      generationStrategy: 'comprehensive',
      complexityLevel: 'moderate',
      maxL4Count: 8,
      includeKnowledgeBase: true,
      autoSave: true  // 直接保存到数据库
    }

    ElMessage.info('启动L4功能点生成任务...')
    
    // 调用异步生成API，和批量生成使用相同的接口
    const response = await generateLevel4Async(generateData)
    
    if (response.code === 0) {
      const taskId = response.data.taskId
      ElMessage.success('L4生成任务已启动，正在后台处理...')
      
      // 显示任务进度追踪对话框
      currentL4TaskId.value = taskId
      l4TaskProgressVisible.value = true
      
      // 提供任务管理提示
      setTimeout(() => {
        ElMessage.info('💡 提示：您可以点击右上角"L4任务管理"按钮查看任务进度')
      }, 2000)
      
    } else {
      throw new Error(response.msg || '启动L4生成任务失败')
    }

  } catch (error) {
    console.error('L4功能点生成失败:', error)
    ElMessage.error('L4生成失败: ' + (error.message || '未知错误'))
  }
}

// Mermaid流程图生成  
const handleMermaidGeneration = async (requirement) => {
  try {
    console.log('Mermaid流程图生成:', requirement)
    
    const generateData = {
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      versionId: currentVersion.value,
      requirementIds: [requirement.id || requirement.ID],
      diagramType: 'flowchart',
      detailLevel: 'detailed',
      options: {
        includeSubRequirements: false,
        autoLayout: true,
        addAnnotations: true,
        groupByLevel: false
      }
    }

    ElMessage.info('正在启动流程图生成任务...')
    const response = await generateMermaidDiagramsAsync(generateData)
    
    if (response.code === 0) {
      const taskId = response.data.taskId
      ElMessage.success(`流程图生成任务已启动，任务ID: ${taskId}`)
      
      // 显示进度追踪对话框
      showMermaidProgressDialog.value = true
      currentMermaidTaskId.value = taskId
      startMermaidProgressPolling(taskId)
      
    } else {
      throw new Error(response.msg || 'Mermaid生成失败')
    }

  } catch (error) {
    console.error('Mermaid流程图生成失败:', error)
    ElMessage.error('Mermaid生成失败: ' + (error.message || '未知错误'))
  }
}

const handleAnalyze = async (requirement) => {
  try {
    console.log('启动单个需求分析，需求ID:', requirement.id || requirement.ID)
    const requirementId = requirement.id || requirement.ID
    
    // 验证项目上下文
    if (!currentProject.value || !currentCycle.value || !currentVersion.value) {
      ElMessage.warning('请先选择项目、周期和版本')
      return
    }
    
    // 重置分析状态
    analysisStatus.value = {
      taskId: null,
      status: 'pending',
      progress: 0,
      currentStage: 'initializing',
      stageDesc: '正在初始化分析环境...',
      animationType: 'pulse',
      errorMsg: '',
      stageHistory: [],
      animationData: {},
      statusText: '任务已创建，等待开始...'
    }
    
    analysisResult.value = null
    showAnalysisResult.value = false
    activeResultTab.value = 'comparison' // 重置tab到默认状态
    
    // 显示分析对话框
    currentRequirement.value = requirement
    analysisDialogVisible.value = true
    
    // 构建单个需求分析请求参数
    const analysisData = { 
      requirementId: requirementId,
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      version_id: currentVersion.value,
      analysis_options: {
        optimize_title: true,
        optimize_description: true,
        analyze_function_type: true,
        calculate_complexity: true,
        estimate_function_points: true,
        use_knowledge_base: true,
        generate_mermaid: false // 单个需求暂不生成流程图
      },
      ai_config: {
        model: 'deepseek',
        temperature: 0.7,
        max_tokens: 1500,
        timeout: 60
      }
    }
    
    console.log('启动单个需求AI分析，参数:', analysisData)
    const response = await analyzeRequirement(analysisData)
    
    if (response.code === 0) {
      console.log('单个需求AI分析启动响应:', response.data)
      
      // 获取任务ID
      const taskId = response.data.taskId || response.data.task_id || response.data.id
      
      if (taskId) {
        // 确保taskId是数字类型
        const numericTaskId = typeof taskId === 'string' ? parseInt(taskId) : taskId
        
        analysisStatus.value.taskId = numericTaskId
        analysisTask.value = numericTaskId
        ElMessage.success('单个需求AI分析已启动，任务ID: ' + numericTaskId)
        
        console.log('设置analysisTask.value为:', numericTaskId)
        
        // 开始轮询进度（使用专用的单个需求分析进度接口）
        pollRequirementAnalysisProgress()
      } else {
        console.error('响应中未找到有效的taskId:', response.data)
        ElMessage.error('需求分析启动成功但未获取到任务ID')
        analysisDialogVisible.value = false
      }
    } else {
      console.error('启动单个需求AI分析失败:', response.msg)
      ElMessage.error('启动需求分析失败：' + response.msg)
      analysisDialogVisible.value = false
    }
  } catch (error) {
    console.error('启动单个需求AI分析失败:', error)
    ElMessage.error('启动需求分析失败：' + error.message)
    analysisDialogVisible.value = false
  }
}

// 轮询单个需求分析进度（专用接口）
const pollRequirementAnalysisProgress = async () => {
  if (!analysisTask.value) {
    console.warn('缺少analysisTask，无法轮询单个需求分析进度')
    return
  }
  
  console.log('开始轮询单个需求分析进度，taskId:', analysisTask.value)
  
  try {
    const response = await getRequirementAnalysisProgress(analysisTask.value)
    console.log('单个需求分析进度API响应:', response)
    
    if (response.code === 0) {
      const data = response.data
      
      // 更新分析状态
      analysisStatus.value = {
        taskId: data.task_id || data.taskId,
        status: data.status,
        progress: data.progress || 0,
        currentStage: data.current_stage || data.currentStage || 'analyzing',
        stageDesc: data.stage_desc || data.stageDesc || '正在分析中...',
        animationType: data.animation_type || data.animationType || 'pulse',
        errorMsg: data.error_msg || data.errorMsg || '',
        stageHistory: data.stage_history || data.stageHistory || [],
        animationData: data.animation_data || data.animationData || {},
        statusText: data.status_text || data.statusText || '分析进行中'
      }
      
      // 同步旧的progress值用于简单的进度条
      analysisProgress.value = data.progress || 0
      
      console.log('单个需求分析进度更新:', analysisStatus.value)
      
      if (data.status === 'completed') {
        console.log('单个需求分析完成，获取详细结果')
        
        // 分析完成，获取详细分析结果
        try {
          const resultResponse = await getRequirementAnalysisResult(analysisTask.value)
          console.log('单个需求分析结果:', resultResponse)
          
          if (resultResponse.code === 0) {
            const resultData = resultResponse.data
            analysisResult.value = resultData
            showAnalysisResult.value = true
            activeResultTab.value = 'comparison'
            
            // 初始化可编辑结果数据 - 适配单个需求分析结果格式
            editableResult.value = {
              title: resultData.optimized?.title || resultData.original?.title || '',
              description: resultData.optimized?.description || resultData.original?.description || '',
              functionType: resultData.optimized?.functionType || resultData.original?.functionType || '',
              businessValue: resultData.optimized?.businessValue || resultData.original?.businessValue || '',
              complexity: resultData.optimized?.complexity || resultData.original?.complexity || ''
            }
            
            ElMessage.success('✨ 单个需求分析完成！AI已为您优化需求内容')
          } else {
            console.error('获取分析结果失败:', resultResponse.msg)
            ElMessage.success('分析完成，但获取详细结果失败')
            setTimeout(() => {
              analysisDialogVisible.value = false
              loadRequirements()
            }, 2000)
          }
        } catch (error) {
          console.error('获取分析结果异常:', error)
          ElMessage.success('分析完成')
          setTimeout(() => {
            analysisDialogVisible.value = false
            loadRequirements()
          }, 2000)
        }
      } else if (data.status === 'failed') {
        console.error('单个需求分析失败:', data.error_msg || data.errorMsg)
        ElMessage.error('需求分析失败：' + (data.error_msg || data.errorMsg || '未知错误'))
        setTimeout(() => {
          analysisDialogVisible.value = false
        }, 3000)
      } else if (data.status === 'running' || data.status === 'pending') {
        // 继续轮询
        console.log('单个需求分析继续中，2秒后再次轮询')
        setTimeout(pollRequirementAnalysisProgress, 2000)
      } else {
        console.warn('未知状态:', data.status)
        setTimeout(pollRequirementAnalysisProgress, 2000)
      }
    } else {
      console.error('单个需求分析进度API返回错误:', response.msg)
      ElMessage.error('获取分析进度失败：' + response.msg)
      analysisDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取单个需求分析进度失败:', error)
    ElMessage.error('获取分析进度失败：' + error.message)
    analysisDialogVisible.value = false
  }
}

// 轮询批量分析进度（保留原有功能）
const pollAnalysisProgress = async () => {
  if (!analysisTask.value) {
    console.warn('缺少analysisTask，无法轮询进度')
    return
  }
  
  console.log('开始轮询批量分析进度，taskId:', analysisTask.value)
  
  try {
    const response = await getAnalysisProgress(analysisTask.value)
    console.log('批量分析进度API响应:', response)
    
    if (response.code === 0) {
      const data = response.data
      
      // 更新分析状态
      analysisStatus.value = {
        taskId: data.task_id || data.taskId,
        status: data.status,
        progress: data.progress || 0,
        currentStage: data.current_stage || data.currentStage || 'analyzing',
        stageDesc: data.stage_desc || data.stageDesc || '正在分析中...',
        animationType: data.animation_type || data.animationType || 'pulse',
        errorMsg: data.error_msg || data.errorMsg || '',
        stageHistory: data.stage_history || data.stageHistory || [],
        animationData: data.animation_data || data.animationData || {},
        statusText: data.status_text || data.statusText || '分析进行中'
      }
      
      // 同步旧的progress值用于简单的进度条
      analysisProgress.value = data.progress || 0
      
      console.log('批量分析进度更新:', analysisStatus.value)
      
      if (data.status === 'completed') {
        console.log('批量分析完成，处理结果:', data.result)
        
        // 任务完成，获取分析结果
        if (data.result) {
          analysisResult.value = data.result
          showAnalysisResult.value = true
          activeResultTab.value = 'comparison' // 确保默认选中对比结果tab
          
          // 初始化可编辑结果数据 - 适配不同的结果格式
          const resultData = data.result
          editableResult.value = {
            title: resultData.optimized?.title || resultData.title || '',
            description: resultData.optimized?.description || resultData.description || '',
            functionType: resultData.optimized?.functionType || resultData.functionType || '',
            businessValue: resultData.optimized?.businessValue || resultData.businessValue || '',
            complexity: resultData.optimized?.complexity || resultData.complexity || ''
          }
          
          ElMessage.success('✨ 分析完成！AI已为您优化需求内容')
          
          // 显示结果3秒后自动切换到结果展示
          setTimeout(() => {
            if (analysisDialogVisible.value) {
              console.log('分析结果已准备，等待用户确认')
            }
          }, 2000)
        } else {
          console.log('分析完成但无结果数据，刷新列表')
          ElMessage.success('分析完成')
          setTimeout(() => {
            analysisDialogVisible.value = false
            loadRequirements() // 刷新需求列表
          }, 2000)
        }
      } else if (data.status === 'failed') {
        console.error('批量分析失败:', data.error_msg || data.errorMsg)
        ElMessage.error('分析失败：' + (data.error_msg || data.errorMsg || '未知错误'))
        setTimeout(() => {
          analysisDialogVisible.value = false
        }, 3000)
      } else if (data.status === 'running' || data.status === 'pending') {
        // 继续轮询
        console.log('批量分析继续中，2秒后再次轮询')
        setTimeout(pollAnalysisProgress, 2000)
      } else {
        console.warn('未知状态:', data.status)
        setTimeout(pollAnalysisProgress, 2000)
      }
    } else {
      console.error('批量分析进度API返回错误:', response.msg)
      ElMessage.error('获取分析进度失败：' + response.msg)
      analysisDialogVisible.value = false
    }
  } catch (error) {
    console.error('获取批量分析进度失败:', error)
    ElMessage.error('获取分析进度失败：' + error.message)
    analysisDialogVisible.value = false
  }
}

// 从详情对话框打开编辑
const handleEditFromDetail = (requirement) => {
  viewDialogVisible.value = false
  currentRequirement.value = requirement
  editDialogMode.value = 'edit'
  editDialogVisible.value = true
}

// 查看子需求
const handleViewChild = (child) => {
  currentRequirementId.value = child.id || child.ID
  // 不关闭当前对话框，直接切换到子需求
}

// 表单提交事件
const handleFormSubmit = async (formData) => {
  try {
    formLoading.value = true
    console.log('提交表单数据:', formData)
    
    if (editDialogMode.value === 'edit') {
      // 调用更新API
      console.log('更新需求:', formData)
      const response = await updateNesmaRequirement(formData)
      if (response.code === 0) {
        ElMessage.success('需求更新成功')
      } else {
        throw new Error(response.msg || '更新失败')
      }
    } else {
      // 调用创建API  
      console.log('创建需求:', formData)
      const response = await createNesmaRequirement(formData)
      if (response.code === 0) {
        ElMessage.success('需求创建成功')
      } else {
        throw new Error(response.msg || '创建失败')
      }
    }
    
    editDialogVisible.value = false
    await loadRequirements()
  } catch (error) {
    console.error('提交失败:', error)
    ElMessage.error('提交失败：' + error.message)
  } finally {
    formLoading.value = false
  }
}

// 表单取消事件
const handleFormCancel = () => {
  editDialogVisible.value = false
}

// 关闭对话框
const handleCloseDialog = (type) => {
  if (type === 'analysis') {
    analysisDialogVisible.value = false
    analysisTask.value = null
    showAnalysisResult.value = false
    analysisResult.value = null
    activeResultTab.value = 'comparison' // 重置tab选择
    // 重置可编辑结果
    editableResult.value = {
      title: '',
      description: '',
      functionType: '',
      businessValue: '',
      complexity: ''
    }
  }
}

// 动画效果支持方法
const getProgressStatus = () => {
  if (analysisStatus.value.status === 'completed') return 'success'
  if (analysisStatus.value.status === 'failed') return 'exception'
  return undefined
}

const getStageIcon = () => {
  const iconMap = {
    'initializing': Timer,
    'validating': Search,
    'fetching': Upload,
    'context_building': FolderOpened,
    'ai_analyzing': MagicStick,
    'ai_calling': Upload,
    'parsing': Document,
    'finalizing': Check,
    'completed': Check,
    'failed': Warning
  }
  return iconMap[analysisStatus.value.currentStage] || Timer
}

const getStageIconClass = () => {
  const classMap = {
    'pulse': 'pulse-animation',
    'scan': 'scan-animation',
    'loading': 'loading-animation',
    'network': 'network-animation',
    'brain': 'brain-animation',
    'api': 'api-animation',
    'parse': 'parse-animation',
    'check': 'check-animation',
    'success': 'success-animation',
    'error': 'error-animation'
  }
  return classMap[analysisStatus.value.animationType] || 'default-animation'
}

const getStageColor = () => {
  if (analysisStatus.value.status === 'failed') return '#F56C6C'
  if (analysisStatus.value.status === 'completed') return '#67C23A'
  
  const colorMap = {
    'initializing': '#409EFF',
    'validating': '#E6A23C',
    'fetching': '#409EFF',
    'context_building': '#909399',
    'ai_analyzing': '#722ED1',
    'ai_calling': '#13C2C2',
    'parsing': '#52C41A',
    'finalizing': '#1890FF'
  }
  return colorMap[analysisStatus.value.currentStage] || '#409EFF'
}

const getTimelineType = (status) => {
  if (status === 'completed') return 'success'
  if (status === 'failed') return 'danger'
  if (status === 'running') return 'primary'
  return 'info'
}

const formatTime = (time) => {
  if (!time) return '--'
  const date = new Date(time)
  return date.toLocaleTimeString('zh-CN', { 
    hour12: false,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit'
  })
}

// 应用分析结果
const handleApplyAnalysis = async () => {
  try {
    if (!analysisResult.value || !editableResult.value) {
      ElMessage.error('没有可应用的分析结果')
      return
    }

    await ElMessageBox.confirm(
      '确定要应用AI的优化建议吗？这将更新当前需求的信息并标记为已完成。',
      '确认应用优化',
      {
        confirmButtonText: '确定应用',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )

    // 使用用户编辑后的数据 - 适配单个需求分析结果
    const applyData = {
      requirementId: currentRequirement.value.id || currentRequirement.value.ID,
      optimizedInfo: {
        title: editableResult.value.title,
        description: editableResult.value.description,
        functionType: editableResult.value.functionType,
        businessValue: editableResult.value.businessValue,
        complexity: editableResult.value.complexity
      },
      analysisNote: analysisResult.value.analysisNote || '通过单个需求AI分析优化',
      updateStatus: true, // 标识需要更新状态为已完成
      taskId: analysisTask.value // 记录关联的分析任务
    }

    console.log('应用单个需求分析结果:', applyData)
    const response = await applyRequirementAnalysisResult(applyData)
    
    if (response.code === 0) {
      ElMessage.success('✨ 需求优化已应用，状态已更新为已完成')
      handleCloseDialog('analysis')
      await loadRequirements() // 刷新需求列表
    } else {
      ElMessage.error('应用优化失败：' + response.msg)
    }
  } catch (error) {
    if (error !== 'cancel') {
      console.error('应用单个需求分析结果失败:', error)
      ElMessage.error('应用优化失败：' + error.message)
    }
  }
}

// L4生成动画效果支持方法
const getL4ProgressStatus = () => {
  if (l4GenerationStatus.value.status === 'completed') return 'success'
  if (l4GenerationStatus.value.status === 'failed') return 'exception'
  return undefined
}

const getL4StageIcon = () => {
  const iconMap = {
    'initializing': Timer,
    'validating': Search,
    'fetching': Upload,
    'context_building': FolderOpened,
    'ai_analyzing': MagicStick,
    'ai_calling': Upload,
    'parsing': Document,
    'finalizing': Check,
    'completed': Check,
    'failed': Warning,
    'generating': MagicStick
  }
  return iconMap[l4GenerationStatus.value.currentStage] || MagicStick
}

const getL4StageIconClass = () => {
  const classMap = {
    'pulse': 'pulse-animation',
    'scan': 'scan-animation',
    'loading': 'loading-animation',
    'network': 'network-animation',
    'brain': 'brain-animation',
    'api': 'api-animation',
    'parse': 'parse-animation',
    'check': 'check-animation',
    'success': 'success-animation',
    'error': 'error-animation'
  }
  return classMap[l4GenerationStatus.value.animationType] || 'brain-animation'
}

const getL4StageColor = () => {
  if (l4GenerationStatus.value.status === 'failed') return '#F56C6C'
  if (l4GenerationStatus.value.status === 'completed') return '#67C23A'
  
  const colorMap = {
    'initializing': '#409EFF',
    'validating': '#E6A23C',
    'fetching': '#409EFF',
    'context_building': '#909399',
    'ai_analyzing': '#722ED1',
    'ai_calling': '#13C2C2',
    'parsing': '#52C41A',
    'finalizing': '#1890FF',
    'generating': '#722ED1'
  }
  return colorMap[l4GenerationStatus.value.currentStage] || '#722ED1'
}

// 计算改进项数
const calculateImprovementCount = () => {
  if (!analysisResult.value || !editableResult.value) return 0
  
  let count = 0
  const originalInfo = analysisResult.value.original
  const optimizedInfo = editableResult.value
  
  // 检查各种改进
  if (optimizedInfo.title && optimizedInfo.title !== originalInfo.title) count++
  if (optimizedInfo.description && !originalInfo.description) count++
  if (optimizedInfo.functionType && !originalInfo.functionType) count++
  if (optimizedInfo.businessValue && !originalInfo.businessValue) count++
  if (optimizedInfo.complexity && !originalInfo.complexity) count++
  
  return count
}

// 获取处理时间
const getProcessingTime = () => {
  if (!analysisStatus.value.stageHistory || analysisStatus.value.stageHistory.length === 0) {
    return '0秒'
  }
  
  // 计算总处理时间（从第一个阶段开始到最后一个阶段结束）
  const startTime = analysisStatus.value.stageHistory[0]?.startTime
  const endTime = analysisStatus.value.stageHistory[analysisStatus.value.stageHistory.length - 1]?.endTime
  
  if (!startTime || !endTime) {
    return '计算中...'
  }
  
  const duration = Math.round((new Date(endTime) - new Date(startTime)) / 1000)
  return `${duration}秒`
}

// 格式化当前时间
const formatCurrentTime = () => {
  return new Date().toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false
  })
}

const handleBatchImport = () => {
  if (!currentProject.value || !currentCycle.value) {
    ElMessage.warning('请先选择项目、周期和版本')
    return
  }
  importDialogVisible.value = true
}

// 导入完成处理
const handleImportCompleted = (result) => {
  console.log('导入完成结果:', result)
  ElMessage.success('导入操作完成')
  // 刷新需求列表
  loadRequirements()
  loadVersions()
}

// L4任务管理相关方法
const handleOpenL4TaskManagement = () => {
  console.log('打开L4任务管理对话框')
  console.log('当前项目:', currentProject.value)
  console.log('当前周期:', currentCycle.value)
  console.log('当前版本:', currentVersion.value)
  
  // 设置当前项目、周期、版本对象
  currentProjectObject.value = projects.value.find(p => p.ID === currentProject.value)
  currentCycleObject.value = cycles.value.find(c => c.ID === currentCycle.value)
  currentVersionObject.value = versions.value.find(v => v.ID === currentVersion.value)
  
  console.log('项目对象:', currentProjectObject.value)
  console.log('周期对象:', currentCycleObject.value)
  console.log('版本对象:', currentVersionObject.value)
  
  l4TaskManagementVisible.value = true
  
  ElMessage.info('L4任务管理对话框已打开')
}

const handleCloseL4TaskManagement = () => {
  l4TaskManagementVisible.value = false
}

const handleViewL4TaskProgress = (taskId) => {
  console.log('查看L4任务进度:', taskId)
  currentL4TaskId.value = taskId
  l4TaskProgressVisible.value = true
}

const handleViewL4TaskResult = (taskId) => {
  console.log('查看L4任务结果:', taskId)
  // 这里可以跳转到L4生成结果页面或者显示结果对话框
  // 暂时关闭任务管理对话框，显示L4生成结果对话框
  l4TaskManagementVisible.value = false
  
  // 模拟加载任务结果到L4生成结果对话框
  // 实际应该调用API获取任务结果
  // loadL4GenerationResult(taskId)
  l4GenerationDialogVisible.value = true
}

const handleL4TaskCompleted = (taskInfo) => {
  console.log('L4任务完成:', taskInfo)
  ElMessage.success('L4生成任务已完成！')
  l4TaskProgressVisible.value = false
  
  // 刷新需求列表
  loadRequirements()
  
  // 可选：显示生成结果
  handleViewL4TaskResult(taskInfo.id)
}

const handleL4TaskFailed = (taskInfo) => {
  console.log('L4任务失败:', taskInfo)
  ElMessage.error('L4生成任务失败：' + (taskInfo.errorMsg || '未知错误'))
  l4TaskProgressVisible.value = false
}

const handleL4TaskCancelled = (taskId) => {
  console.log('L4任务取消:', taskId)
  ElMessage.info('L4生成任务已取消')
  l4TaskProgressVisible.value = false
}

// 项目上下文变更
const handleProjectChange = async () => {
  console.log('项目变更，新项目ID:', currentProject.value)
  currentCycle.value = null
  currentVersion.value = null
  cycles.value = []
  versions.value = []
  tableData.value = []
  
  if (currentProject.value) {
    await loadCycles()
  }
}

const handleCycleChange = async () => {
  console.log('周期变更，新周期ID:', currentCycle.value)
  currentVersion.value = null
  versions.value = []
  tableData.value = []
  
  if (currentCycle.value) {
    await loadVersions()
  }
}

const handleVersionChange = async () => {
  console.log('版本变更，新版本ID:', currentVersion.value)
  tableData.value = []
  
  if (currentVersion.value) {
    await loadRequirements()
  }
}

// 搜索和筛选
const handleSearch = () => {
  if (viewMode.value === 'tree') {
    // 树形视图搜索：重新加载数据以保持树结构完整性
    console.log('树形视图搜索，重新加载数据')
    loadRequirements()
  } else {
    // 表格视图搜索：重置分页并加载
    page.value = 1
    loadRequirements()
  }
}

const resetSearch = () => {
  searchForm.value = {
    title: '',
    description: '',
    level: null,
    status: '',
    functionType: ''
  }
  
  if (viewMode.value === 'tree') {
    // 树形视图：重新加载全部数据
    loadRequirements()
  } else {
    // 表格视图：重置分页并加载
    page.value = 1
    loadRequirements()
  }
}

const handleSelectionChange = (selection) => {
  selectedRequirements.value = selection
  console.log('选中需求数量:', selection.length)
}

// 智能优化选中需求
const handleIntelligentOptimization = async () => {
  if (!selectedRequirements.value.length) {
    ElMessage.warning('请先选择需要优化的需求')
    return
  }

  if (!currentProject.value || !currentCycle.value || !currentVersion.value) {
    ElMessage.warning('请先选择项目、周期和版本')
    return
  }

  try {
    // 根据选择的层级过滤需求
    const targetLevels = selectedAnalysisLevel.value.split(',').map(l => parseInt(l.trim()))
    const analyzableRequirements = selectedRequirements.value.filter(req => 
      targetLevels.includes(req.level)
    )

    if (analyzableRequirements.length === 0) {
      const levelText = targetLevels.map(l => `L${l}`).join('和')
      ElMessage.warning(`请选择${levelText}级别的需求进行智能优化`)
      return
    }

    ElMessage.info(`开始分析 ${analyzableRequirements.length} 个${targetLevels.map(l => `L${l}`).join('+')}需求...`)

    // 显示分析进度对话框
    analysisDialogVisible.value = true
    
    // 重置分析状态
    analysisStatus.value = {
      taskId: null,
      status: 'running',
      progress: 0,
      currentStage: 'initializing',
      stageDesc: '正在初始化分析任务...',
      animationType: 'pulse',
      errorMsg: '',
      stageHistory: [],
      animationData: { icon: 'cpu', color: 'primary' },
      statusText: '智能分析进行中'
    }

    // 启动批量智能分析
    const analysisData = {
      project_id: currentProject.value,
      cycle_id: currentCycle.value,
      version_id: currentVersion.value,
      requirement_ids: analyzableRequirements.map(req => req.id || req.ID),
      analysis_type: 'intelligent_optimization',
      options: {
        include_l2_analysis: analyzableRequirements.some(req => req.level === 2),
        include_l3_analysis: analyzableRequirements.some(req => req.level === 3),
        use_knowledge_base: true,
        generate_suggestions: true
      }
    }

    const response = await startIntelligentAnalysis(analysisData)
    
    if (response.data && response.data.task_id) {
      analysisStatus.value.taskId = response.data.task_id
      
      // 开始轮询进度
      pollAnalysisProgress()
      
      ElMessage.success('智能分析任务已启动')
    } else {
      throw new Error('启动分析任务失败')
    }

  } catch (error) {
    console.error('智能优化失败:', error)
    ElMessage.error('智能优化失败: ' + (error.message || '未知错误'))
    
    // 关闭进度对话框
    analysisDialogVisible.value = false
  }
}

// 批量L4生成 - 优化版本，集成任务管理
const handleBatchL4Generation = async () => {
  const l3Requirements = selectedRequirements.value.filter(req => req.level === 3)
  if (l3Requirements.length === 0) {
    ElMessage.warning('请先选择三级需求')
    return
  }
  
  try {
    // 准备异步生成请求数据
    const generateData = {
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      versionId: currentVersion.value,
      l3RequirementIds: l3Requirements.map(req => req.id || req.ID),
      generationStrategy: 'comprehensive',
      batchSize: l3Requirements.length,
      enableOptimization: true,
      confidenceThreshold: 0.7,
      maxL4PerL3: 10,
      useKnowledgeBase: true,
      generateDescription: true,
      generateMermaid: false,
      analysisDepth: 'detailed'
    }
    
    ElMessage.info(`准备为 ${l3Requirements.length} 个L3需求生成L4功能点...`)
    
    // 调用异步生成API
    const response = await generateLevel4Async(generateData)
    
    if (response.code === 0) {
      const taskId = response.data.taskId || response.data.task_id
      
      ElMessage.success('L4生成任务已创建，正在后台处理...')
      
      // 显示任务进度追踪对话框
      currentL4TaskId.value = taskId
      l4TaskProgressVisible.value = true
      
      // 同时记录到任务管理系统
      console.log('L4生成任务已创建:', taskId)
      
      // 清空选中的需求
      selectedRequirements.value = []
      
      // 可选：显示成功提示，告知用户可以通过任务管理查看进度
      setTimeout(() => {
        ElMessage.info('您可以点击"L4任务管理"按钮查看所有L4生成任务')
      }, 3000)
      
    } else {
      throw new Error(response.msg || 'L4生成任务创建失败')
    }

  } catch (error) {
    console.error('批量L4生成失败:', error)
    ElMessage.error('批量L4生成失败: ' + (error.message || '未知错误'))
    
    // 如果API调用失败，回退到原有的同步方式
    handleBatchL4GenerationLegacy()
  }
}

// 原有的批量L4生成逻辑（作为备用）
const handleBatchL4GenerationLegacy = async () => {
  const l3Requirements = selectedRequirements.value.filter(req => req.level === 3)
  if (l3Requirements.length === 0) {
    ElMessage.warning('请先选择三级需求')
    return
  }
  
  try {
    // 准备异步生成请求数据
    const generateData = {
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      versionId: currentVersion.value,
      l3RequirementIds: l3Requirements.map(req => req.id || req.ID),
      generationStrategy: 'comprehensive',
      complexityLevel: 'moderate',
      maxL4Count: 8,
      includeKnowledgeBase: true,
      autoSave: true  // 直接保存到数据库
    }

    ElMessage.info(`启动 ${l3Requirements.length} 个三级需求的L4功能点生成任务...`)
    
    // 调用异步生成API
    const response = await generateLevel4Async(generateData)
    
    if (response.code === 0) {
      const taskId = response.data.taskId
      ElMessage.success('L4生成任务已启动，正在后台处理...')
      
      // 显示进度对话框并开始轮询
      startL4GenerationProgressDialog(taskId, l3Requirements.length)
      
    } else {
      throw new Error(response.msg || '启动L4生成任务失败')
    }

  } catch (error) {
    console.error('批量L4生成失败:', error)
    ElMessage.error('批量L4生成失败: ' + (error.message || '未知错误'))
  }
}

// L4生成进度对话框相关函数
const startL4GenerationProgressDialog = (taskId, totalL3Count) => {
  l4GenerationTask.value = { id: taskId, totalCount: totalL3Count }
  l4GenerationProgress.value = 0
  
  // 重置增强的状态结构
  l4GenerationStatus.value = {
    taskId: taskId,
    status: 'running',
    progress: 0,
    currentStage: 'initializing',
    stageDesc: '正在初始化L4生成环境...',
    animationType: 'pulse',
    errorMsg: '',
    stageHistory: [],
    animationData: {},
    statusText: '任务已创建，等待开始...'
  }
  
  l4GenerationResult.value = null
  l4GenerationDialogVisible.value = true
  
  // 开始轮询进度
  pollL4GenerationProgress(taskId)
}

const pollL4GenerationProgress = async (taskId) => {
  const pollInterval = setInterval(async () => {
    try {
      const response = await getLevel4GenerationTaskProgress(taskId)
      
      if (response.code === 0) {
        const data = response.data
        
        // 更新增强的L4生成状态
        l4GenerationStatus.value = {
          taskId: data.task_id || data.taskId || taskId,
          status: data.status,
          progress: data.progress || 0,
          currentStage: data.current_stage || data.currentStage || 'generating',
          stageDesc: data.stage_desc || data.stageDesc || '正在生成L4功能点...',
          animationType: data.animation_type || data.animationType || 'brain',
          errorMsg: data.error_msg || data.errorMsg || '',
          stageHistory: data.stage_history || data.stageHistory || [],
          animationData: data.animation_data || data.animationData || {},
          statusText: data.status_text || data.statusText || 'L4生成进行中'
        }
        
        // 同步旧的progress值用于简单的进度条
        l4GenerationProgress.value = data.progress || 0
        
        console.log('L4生成进度更新:', l4GenerationStatus.value)
        
        if (data.status === 'completed') {
          clearInterval(pollInterval)
          l4GenerationProgress.value = 100
          ElMessage.success('L4功能点生成完成！')
          
          // 获取生成结果
          await loadL4GenerationResult(taskId)
          
          // 刷新需求列表
          await loadRequirements()
          
        } else if (data.status === 'failed') {
          clearInterval(pollInterval)
          ElMessage.error('L4生成任务失败: ' + (data.error_msg || data.errorMsg || '未知错误'))
          l4GenerationDialogVisible.value = false
        }
      }
    } catch (error) {
      console.error('轮询L4生成进度失败:', error)
      clearInterval(pollInterval)
      ElMessage.error('获取生成进度失败')
    }
  }, 2000) // 每2秒轮询一次
}

const loadL4GenerationResult = async (taskId) => {
  try {
    const response = await getLevel4GenerationTaskResult(taskId)
    if (response.code === 0) {
      l4GenerationResult.value = response.data
    }
  } catch (error) {
    console.error('获取L4生成结果失败:', error)
  }
}

const closeL4GenerationDialog = () => {
  l4GenerationDialogVisible.value = false
  l4GenerationTask.value = null
  l4GenerationProgress.value = 0
  
  // 重置增强的状态结构
  l4GenerationStatus.value = {
    taskId: null,
    status: 'pending',
    progress: 0,
    currentStage: '',
    stageDesc: '',
    animationType: '',
    errorMsg: '',
    stageHistory: [],
    animationData: {},
    statusText: ''
  }
  
  l4GenerationResult.value = null
  selectedL4Requirements.value = []
  adoptedL4Requirements.value.clear() // 重置已采纳状态
}

// L4生成结果对话框辅助函数
const calculateSuccessRate = () => {
  if (!l4GenerationResult.value?.task) return 0
  const task = l4GenerationResult.value.task
  const totalCount = task.totalCount || 0
  const successCount = task.successCount || 0
  return totalCount > 0 ? Math.round((successCount / totalCount) * 100) : 0
}

const formatDuration = (seconds) => {
  if (!seconds) return '0秒'
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return mins > 0 ? `${mins}分${secs}秒` : `${secs}秒`
}


const getComplexityColor = (complexity) => {
  const colors = {
    '简单': 'success',
    '中等': 'warning',
    '复杂': 'danger'
  }
  return colors[complexity] || 'default'
}

const getConfidenceColor = (confidence) => {
  if (confidence >= 0.9) return 'success'
  if (confidence >= 0.7) return 'warning'
  return 'danger'
}

const getPriorityColor = (priority) => {
  const colors = {
    1: 'danger',
    2: 'warning',
    3: 'primary',
    4: 'info',
    5: 'success'
  }
  return colors[priority] || 'info'
}

// 获取L4建议列表
const getL4SuggestionsList = () => {
  const results = l4GenerationResult.value?.task?.result?.l4_suggestions_results || []
  const suggestionsList = []
  
  results.forEach(result => {
    const l3Requirement = result.l3_requirement
    const suggestions = result.l4_suggestions || []
    
    suggestions.forEach(suggestion => {
      suggestionsList.push({
        ...suggestion,
        // 添加L3需求信息
        l3RequirementId: l3Requirement?.ID,
        l3RequirementTitle: l3Requirement?.title,
        // 转换字段名以匹配表格显示
        code: suggestion.suggested_code,
        title: suggestion.suggested_title,
        functionType: suggestion.function_type,
        complexity: suggestion.estimated_complexity,
        functionPoints: suggestion.recommended_ufp,
        description: suggestion.suggested_description,
        businessValue: suggestion.business_value,
        acceptanceCriteria: suggestion.acceptance_criteria,
        confidence: suggestion.confidence,
        priority: suggestion.priority,
        generationReason: suggestion.generation_reason,
        relatedKnowledge: suggestion.related_knowledge,
        relatedRequirements: suggestion.related_requirements,
        source: 'ai_generated'
      })
    })
  })
  
  return suggestionsList
}

// 计算实际生成的L4数量
const getCreatedL4Count = () => {
  const results = l4GenerationResult.value?.task?.result?.l4_suggestions_results || []
  let totalCount = 0
  results.forEach(result => {
    totalCount += result.l4_suggestions?.length || 0
  })
  return totalCount
}

// 判断是否有生成失败的情况
const hasGenerationErrors = () => {
  const task = l4GenerationResult.value?.task
  return task && task.failedCount > 0
}

// 获取失败错误信息
const getGenerationErrors = () => {
  const result = l4GenerationResult.value?.task?.result
  if (result && result.errors) {
    return result.errors
  }
  return []
}

const selectAllL4Requirements = () => {
  selectedL4Requirements.value = [...getL4SuggestionsList()]
}

const selectNoneL4Requirements = () => {
  selectedL4Requirements.value = []
}

const handleL4SelectionChange = (selection) => {
  selectedL4Requirements.value = selection
}

const viewL4Detail = (requirement) => {
  // 查看L4需求详情
  console.log('查看L4详情:', requirement)
  currentL4Detail.value = {
    ...requirement,
    // 确保所有字段都有值
    title: requirement.title || '',
    description: requirement.description || '',
    functionType: requirement.functionType || '',
    complexity: requirement.complexity || '',
    functionPoints: requirement.functionPoints || 0,
    confidence: requirement.confidence || 0,
    priority: requirement.priority || 1,
    businessValue: requirement.businessValue || '',
    acceptanceCriteria: requirement.acceptanceCriteria || '',
    generationReason: requirement.generationReason || '',
    relatedKnowledge: requirement.relatedKnowledge || [],
    relatedRequirements: requirement.relatedRequirements || [],
    l3RequirementTitle: requirement.l3RequirementTitle || '',
    l3RequirementId: requirement.l3RequirementId || ''
  }
  l4DetailDialogVisible.value = true
}

// 采纳单个L4建议
const adoptL4Suggestion = async (suggestion) => {
  try {
    // 检查是否已经采纳过
    const suggestionKey = `${suggestion.l3RequirementId}_${suggestion.title}_${suggestion.functionType}`
    if (adoptedL4Requirements.value.has(suggestionKey)) {
      ElMessage.warning('该功能点已经采纳过了')
      return
    }

    // 构建L4需求数据
    const l4RequirementData = {
      aiAnalysisStatus: 'completed',
      title: suggestion.title,
      description: suggestion.description,
      functionType: suggestion.functionType,
      complexity: suggestion.complexity,
      afp: suggestion.recommended_afp,
      ufp: suggestion.functionPoints,
      level: 4,
      parentId: suggestion.l3RequirementId,
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      versionId: currentVersion.value,
      status: 'completed',
      source: 'ai_generated',
      businessValue: suggestion.businessValue,
      acceptanceCriteria: suggestion.acceptanceCriteria,
      notes: `AI生成建议，置信度: ${Math.round(suggestion.confidence * 100)}%，优先级: P${suggestion.priority}`,
      priority: suggestion.priority || 1
    }
    
    ElMessage.info('正在采纳L4功能点建议...')
    const response = await createNesmaRequirement(l4RequirementData)
    
    if (response.code === 0) {
      // 标记为已采纳
      adoptedL4Requirements.value.add(suggestionKey)
      
      ElMessage.success('L4功能点采纳成功！')
      // 刷新需求列表
      await loadRequirements()
    } else {
      throw new Error(response.msg || '采纳失败')
    }
  } catch (error) {
    console.error('采纳L4建议失败:', error)
    ElMessage.error('采纳L4建议失败: ' + (error.message || '未知错误'))
  }
}

// 检查L4建议是否已采纳
const isL4SuggestionAdopted = (suggestion) => {
  const suggestionKey = `${suggestion.l3RequirementId}_${suggestion.title}_${suggestion.functionType}`
  return adoptedL4Requirements.value.has(suggestionKey)
}

// 批量采纳L4建议
const confirmSelectedL4Requirements = async () => {
  if (selectedL4Requirements.value.length === 0) {
    ElMessage.warning('请至少选择一个功能点')
    return
  }
  
  try {
    ElMessage.info(`正在批量采纳 ${selectedL4Requirements.value.length} 个L4功能点...`)
    
    // 批量创建L4需求
    const createPromises = selectedL4Requirements.value.map(suggestion => {
      const l4RequirementData = {
        title: suggestion.title,
        description: suggestion.description,
        functionType: suggestion.functionType,
        complexity: suggestion.complexity,
        afp: suggestion.recommended_afp,
        ufp: suggestion.functionPoints,
        level: 4,
        parentId: suggestion.l3RequirementId,
        projectId: currentProject.value,
        cycleId: currentCycle.value,
        versionId: currentVersion.value,
        status: 'pending',
        source: 'ai_generated',
        businessValue: suggestion.businessValue,
        acceptanceCriteria: suggestion.acceptanceCriteria,
        notes: `AI生成建议，置信度: ${Math.round(suggestion.confidence * 100)}%，优先级: P${suggestion.priority}`,
        priority: suggestion.priority || 1
      }
      return createNesmaRequirement(l4RequirementData)
    })
    
    const results = await Promise.all(createPromises)
    const successCount = results.filter(result => result.code === 0).length
    
    ElMessage.success(`批量采纳完成，成功采纳 ${successCount} 个L4功能点`)
    closeL4GenerationDialog()
    
    // 刷新需求列表
    await loadRequirements()
  } catch (error) {
    console.error('批量采纳L4功能点失败:', error)
    ElMessage.error('批量采纳失败: ' + (error.message || '未知错误'))
  }
}

// 批量Mermaid流程图生成 - 使用异步流程
const handleBatchMermaidGeneration = async () => {
  const l4Requirements = selectedRequirements.value.filter(req => req.level === 4)
  if (l4Requirements.length === 0) {
    ElMessage.warning('请先选择四级功能点')
    return
  }
  
  try {
    // 构建异步批量生成请求
    const generateData = {
      projectId: currentProject.value,
      cycleId: currentCycle.value,
      versionId: currentVersion.value,
      requirementIds: l4Requirements.map(req => req.id || req.ID),
      diagramType: 'flowchart',
      detailLevel: 'detailed',
      options: {
        includeSubRequirements: false,
        autoLayout: true,
        addAnnotations: true,
        groupByLevel: false
      }
    }

    ElMessage.info(`正在启动 ${l4Requirements.length} 个功能点的流程图生成任务...`)
    const response = await generateMermaidDiagramsAsync(generateData)
    
    if (response.code === 0) {
      const taskId = response.data.taskId
      ElMessage.success(`流程图生成任务已启动，任务ID: ${taskId}`)
      
      // 显示进度追踪对话框
      showMermaidProgressDialog.value = true
      currentMermaidTaskId.value = taskId
      
      // 重置流程图生成状态
      mermaidGenerationStatus.value = {
        taskId: taskId,
        status: 'running',
        progress: 0,
        totalCount: l4Requirements.length,
        processedCount: 0,
        successCount: 0,
        failedCount: 0,
        errorMsg: '',
        summary: '',
        currentStage: 'initializing',
        estimatedRemaining: 120
      }
      
      // 开始轮询进度
      startMermaidProgressPolling(taskId)
      
      // 清空选中的需求
      selectedRequirements.value = []
      
    } else {
      throw new Error(response.msg || '批量Mermaid生成失败')
    }

  } catch (error) {
    console.error('批量Mermaid生成失败:', error)
    ElMessage.error('批量Mermaid生成失败: ' + (error.message || '未知错误'))
  }
}

const refreshData = () => {
  loadRequirements()
}

// 辅助方法
const getLevelType = (level) => {
  const levelMap = {
    1: 'danger',
    2: 'warning', 
    3: 'primary',
    4: 'success'
  }
  return levelMap[level] || 'info'
}

const getFunctionTypeColor = (type) => {
  const typeMap = {
    'EI': 'primary',
    'EO': 'success',
    'EQ': 'warning',
    'ILF': 'info',
    'EIF': 'danger'
  }
  return typeMap[type] || 'info'
}

// ==================== Mermaid流程图生成进度追踪 ====================

// 启动Mermaid进度轮询
const startMermaidProgressPolling = (taskId) => {
  const pollInterval = setInterval(async () => {
    try {
      const response = await getMermaidGenerationTaskProgress(taskId)
      
      if (response.code === 0) {
        const data = response.data
        
        // 更新增强的状态
        mermaidGenerationStatus.value = {
          taskId: data.taskId,
          status: data.status,
          progress: data.progress,
          totalCount: data.totalCount,
          processedCount: data.processedCount,
          successCount: data.successCount,
          failedCount: data.failedCount,
          errorMsg: data.errorMsg,
          summary: data.summary,
          // 增强的状态字段
          currentStage: data.currentStage || determineMermaidStage(data.progress, data.status),
          estimatedRemaining: data.estimatedRemaining || calculateEstimatedTime(data.progress)
        }
        
        // 任务完成时停止轮询
        if (data.status === 'completed' || data.status === 'failed') {
          clearInterval(pollInterval)
          
          if (data.status === 'completed') {
            ElMessage.success(`流程图生成完成！成功 ${data.successCount} 个，失败 ${data.failedCount} 个`)
            // 刷新需求列表以显示新生成的流程图
            await loadRequirements()
          } else {
            ElMessage.error(`流程图生成失败：${data.errorMsg}`)
          }
          
          // 5秒后自动关闭进度对话框
          setTimeout(() => {
            showMermaidProgressDialog.value = false
          }, 5000)
        }
      }
    } catch (error) {
      console.error('获取Mermaid任务进度失败:', error)
      clearInterval(pollInterval)
      ElMessage.error('获取任务进度失败')
      showMermaidProgressDialog.value = false
    }
  }, 2000) // 每2秒轮询一次
}

// 根据进度确定当前阶段
const determineMermaidStage = (progress, status) => {
  if (status === 'failed') return 'failed'
  if (status === 'completed') return 'completed'
  
  if (progress < 10) return 'initializing'
  if (progress < 20) return 'validating'
  if (progress < 30) return 'fetching'
  if (progress < 80) return 'generating'
  if (progress < 95) return 'processing'
  return 'finalizing'
}

// 计算预计剩余时间
const calculateEstimatedTime = (progress) => {
  if (progress <= 0) return 120
  if (progress >= 95) return 5
  
  // 简单的线性估算
  const baseTime = 60 // 基础60秒
  const remaining = (100 - progress) / 100
  return Math.ceil(baseTime * remaining)
}

// Mermaid流程图生成动画效果支持方法
const getMermaidProgressStatus = () => {
  if (mermaidGenerationStatus.value.status === 'completed') return 'success'
  if (mermaidGenerationStatus.value.status === 'failed') return 'exception'
  return undefined
}

const getMermaidStageIcon = () => {
  const iconMap = {
    'initializing': Timer,
    'validating': Search,
    'fetching': Upload,
    'generating': Coordinate,
    'processing': MagicStick,
    'finalizing': Check,
    'completed': Check,
    'failed': Warning
  }
  return iconMap[mermaidGenerationStatus.value.currentStage] || Coordinate
}

const getMermaidStageIconClass = () => {
  const classMap = {
    'pulse': 'pulse-animation',
    'scan': 'scan-animation',
    'loading': 'loading-animation',
    'coordinate': 'coordinate-animation',
    'processing': 'processing-animation',
    'success': 'success-animation',
    'error': 'error-animation'
  }
  return classMap[mermaidGenerationStatus.value.animationType] || 'coordinate-animation'
}

const getMermaidStageColor = () => {
  if (mermaidGenerationStatus.value.status === 'failed') return '#F56C6C'
  if (mermaidGenerationStatus.value.status === 'completed') return '#67C23A'
  
  const colorMap = {
    'initializing': '#409EFF',
    'validating': '#E6A23C',
    'fetching': '#409EFF',
    'generating': '#13C2C2',
    'processing': '#722ED1',
    'finalizing': '#1890FF'
  }
  return colorMap[mermaidGenerationStatus.value.currentStage] || '#13C2C2'
}

const getMermaidAnimationType = () => {
  const animationMap = {
    'initializing': 'pulse',
    'validating': 'scan',
    'fetching': 'loading',
    'generating': 'coordinate',
    'processing': 'processing',
    'finalizing': 'success',
    'completed': 'success',
    'failed': 'error'
  }
  return animationMap[mermaidGenerationStatus.value.currentStage] || 'coordinate'
}

const getMermaidStageDesc = () => {
  const descMap = {
    'initializing': '正在初始化流程图生成环境...',
    'validating': '正在验证L4需求信息...',
    'fetching': '正在获取需求详情...',
    'generating': '正在生成Mermaid流程图...',
    'processing': 'AI正在处理流程图结构...',
    'finalizing': '正在完成最后处理...',
    'completed': '流程图生成完成',
    'failed': '流程图生成失败'
  }
  return descMap[mermaidGenerationStatus.value.currentStage] || '正在生成流程图...'
}

const getMermaidStatusText = () => {
  const statusMap = {
    'initializing': '系统正在准备生成环境',
    'validating': '正在检查需求完整性',
    'fetching': '正在获取相关数据',
    'generating': 'AI正在创建流程图',
    'processing': '正在优化图表结构',
    'finalizing': '即将完成',
    'completed': '所有流程图已生成完成',
    'failed': '生成过程中遇到错误'
  }
  return statusMap[mermaidGenerationStatus.value.currentStage] || '流程图生成进行中'
}

const calculateMermaidSuccessRate = () => {
  const total = mermaidGenerationStatus.value.totalCount || 0
  const success = mermaidGenerationStatus.value.successCount || 0
  return total > 0 ? Math.round((success / total) * 100) : 0
}

// Mermaid任务管理相关
const showMermaidTaskManagementDialog = ref(false)

const handleOpenMermaidTaskManagement = () => {
  console.log('打开Mermaid流程图任务管理对话框')
  console.log('当前项目:', currentProject.value)
  console.log('当前周期:', currentCycle.value)
  console.log('当前版本:', currentVersion.value)
  
  // 设置当前项目、周期、版本对象
  currentProjectObject.value = projects.value.find(p => p.ID === currentProject.value)
  currentCycleObject.value = cycles.value.find(c => c.ID === currentCycle.value)
  currentVersionObject.value = versions.value.find(v => v.ID === currentVersion.value)
  
  console.log('项目对象:', currentProjectObject.value)
  console.log('周期对象:', currentCycleObject.value)
  console.log('版本对象:', currentVersionObject.value)
  
  showMermaidTaskManagementDialog.value = true
  
  ElMessage.info('Mermaid流程图任务管理对话框已打开')
}

const handleCloseMermaidTaskManagement = () => {
  showMermaidTaskManagementDialog.value = false
}

const handleViewMermaidTaskProgress = (taskId) => {
  console.log('查看Mermaid任务进度:', taskId)
  // 这里可以打开进度追踪对话框或跳转到进度页面
  ElMessage.info(`查看Mermaid任务 ${taskId} 的进度`)
}

const handleViewMermaidTaskResult = (taskId) => {
  console.log('查看Mermaid任务结果:', taskId)
  // 这里可以显示结果对话框或跳转到结果页面
  ElMessage.info(`查看Mermaid任务 ${taskId} 的结果`)
}

// 关闭Mermaid进度对话框
const closeMermaidProgressDialog = () => {
  showMermaidProgressDialog.value = false
  currentMermaidTaskId.value = null
}

const getStatusType = (status) => {
  const statusMap = {
    'pending': 'warning',
    'analyzing': 'primary',
    'completed': 'success',
    'review': 'info'
  }
  return statusMap[status] || 'info'
}

const getStatusText = (status) => {
  const statusMap = {
    'pending': '待分析',
    'analyzing': '分析中',
    'completed': '已完成',
    'review': '需要确认'
  }
  return statusMap[status] || '未知'
}

const formatDate = (date) => {
  if (!date) return '-'
  return new Date(date).toLocaleDateString('zh-CN')
}

// 数据加载方法
const loadRequirements = async () => {
  if (!currentVersion.value) return
  
  // 只有在表格视图时才显示常规loading，树形视图使用viewSwitching
  if (viewMode.value === 'table') {
    loading.value = true
  }
  
  try {
    let params
    
    if (viewMode.value === 'tree') {
      // 树形视图：加载全部数据，忽略分页
      params = {
        versionId: currentVersion.value,
        pageSize: 10000, // 设置一个大数值获取全部数据
        page: 1,
        // 只保留层级和状态筛选，移除关键词搜索避免破坏树结构
        level: searchForm.value.level,
        status: searchForm.value.status,
        functionType: searchForm.value.functionType
      }
    } else {
      // 表格视图：正常分页
      params = {
        page: page.value,
        pageSize: pageSize.value,
        versionId: currentVersion.value,
        // 添加搜索参数
        ...searchForm.value
      }
    }
    
    // 清理空参数
    Object.keys(params).forEach(key => {
      if (params[key] === '' || params[key] === null || params[key] === undefined) {
        delete params[key]
      }
    })
    
    console.log('查询参数:', params, '视图模式:', viewMode.value)
    const response = await getNesmaRequirementList(params)
    
    // 统一ID字段，确保复选框正常工作
    const rawData = response.data.list || []
    
    if (viewMode.value === 'tree') {
      // 树形视图：数据返回后不立即关闭loading，等待渲染完成
      console.log('树形视图数据加载完成，准备开始渲染，数据量:', rawData.length)
      
      // 在下一个微任务中设置数据，让加载动画继续显示
      await nextTick()
      
      tableData.value = rawData.map(item => ({
        ...item,
        id: item.ID || item.id,
        ID: item.ID || item.id
      }))
      
      total.value = rawData.length
      
      // 数据设置完成后，等待Vue组件完成DOM更新
      console.log('树形数据已设置，等待DOM完全更新...')
      
      // 等待多个渲染周期，确保树形组件完全初始化
      await nextTick()
      await nextTick()
      await nextTick()
      
      console.log('DOM更新完成，等待树形组件内部的渲染检测完成...')
      // 此时树形组件的checkTreeRenderComplete方法会自动检测渲染状态
      // 并在真正完成时发送tree-rendered事件
      
    } else {
      // 表格视图：正常处理
      tableData.value = rawData.map(item => ({
        ...item,
        id: item.ID || item.id,
        ID: item.ID || item.id
      }))
      
      total.value = response.data.total || 0
    }
    
  } catch (error) {
    ElMessage.error('加载需求列表失败：' + error.message)
  } finally {
    // 只有表格视图才在这里关闭loading
    if (viewMode.value === 'table') {
      loading.value = false
    }
  }
}

const loadCycles = async () => {
  if (!currentProject.value) {
    cycles.value = []
    return
  }
  
  try {
    const response = await getProjectCycles(currentProject.value)
    
    if (response.code === 0) {
      // 统一字段名，确保同时有ID和id字段
      cycles.value = (response.data || []).map(cycle => ({
        ...cycle,
        id: cycle.ID || cycle.id,
        ID: cycle.ID || cycle.id
      }))
      
      // 如果有周期，自动选择第一个
      if (cycles.value.length > 0 && !currentCycle.value) {
        currentCycle.value = cycles.value[0].ID || cycles.value[0].id
        // 加载该周期的版本
        await loadVersions()
      }
    } else {
      ElMessage.error('加载项目周期失败：' + response.msg)
      cycles.value = []
    }
  } catch (error) {
    console.error('加载项目周期失败:', error)
    ElMessage.error('加载项目周期失败：' + error.message)
    cycles.value = []
  }
}

const loadVersions = async () => {
  if (!currentCycle.value) {
    versions.value = []
    return
  }
  
  try {
    const response = await getRequirementVersions(currentCycle.value)
    
    if (response.code === 0) {
      // 统一字段名，确保同时有ID和id字段
      versions.value = (response.data || []).map(version => ({
        ...version,
        id: version.ID || version.id,
        ID: version.ID || version.id
      }))
      
      // 如果有版本，自动选择第一个
      if (versions.value.length > 0 && !currentVersion.value) {
        currentVersion.value = versions.value[0].ID || versions.value[0].id
        console.log('自动选择第一个版本:', currentVersion.value)
        // 加载该版本的需求
        await loadRequirements()
      }
    } else {
      ElMessage.error('加载需求版本失败：' + response.msg)
      versions.value = []
    }
  } catch (error) {
    console.error('加载需求版本失败:', error)
    ElMessage.error('加载需求版本失败：' + error.message)
    versions.value = []
  }
}

// 分页事件
const handleSizeChange = (size) => {
  pageSize.value = size
  loadRequirements()
}

const handleCurrentChange = (currentPage) => {
  page.value = currentPage
  loadRequirements()
}

// 生命周期
onMounted(async () => {
  try {
    // 加载项目列表
    const projectRes = await getNesmaProjectList({ 
      page: 1, 
      pageSize: 100 // 获取所有项目
    })
    
    
    if (projectRes.code === 0) {
      // 处理项目数据，确保ID字段统一
      const rawProjects = projectRes.data.list || []
      projects.value = rawProjects.map(project => ({
        ...project,
        id: project.ID || project.id,
        ID: project.ID || project.id
      }))
    } else {
      ElMessage.error('加载项目列表失败：' + projectRes.msg)
      // 降级到示例数据
      projects.value = []
    }
  } catch (error) {
    console.error('加载项目列表失败:', error)
    ElMessage.error('加载项目列表失败：' + error.message)
    // 降级到示例数据
    projects.value = []
  }
  
  // 检查URL参数中是否有projectId
  const projectIdFromUrl = route.query.projectId
  
  if (projectIdFromUrl) {
    // 将字符串转换为数字
    const projectId = parseInt(projectIdFromUrl)
    
    // 等待项目数据加载完成后再查找
    setTimeout(async () => {
      const project = projects.value.find(p => (p.ID === projectId || p.id === projectId))
      
      if (project) {
        currentProject.value = project.ID || project.id
        ElMessage.success(`已自动选择项目: ${project.name}`)
        
        // 自动触发项目变更逻辑，加载周期和版本
        await handleProjectChange()
      } else {
        ElMessage.warning(`项目ID ${projectId} 不存在`)
        // 如果有项目，选择第一个
        if (projects.value.length > 0) {
          currentProject.value = projects.value[0].ID || projects.value[0].id
          await handleProjectChange()
        }
      }
    }, 100)
  } else {
    // 如果没有URL参数，选择第一个项目
    if (projects.value.length > 0) {
      currentProject.value = projects.value[0].ID || projects.value[0].id
      console.log("当前项目", currentProject.value)
      await handleProjectChange()
    }
  }
})
</script>

<style lang="scss" scoped>
// 需求详情对话框样式
.requirement-detail {
  .requirement-description,
  .ai-description,
  .function-detail {
    max-height: 200px;
    overflow-y: auto;
    padding: 12px;
    background-color: #f8f9fa;
    border-radius: 6px;
    border: 1px solid #e4e7ed;
    white-space: pre-wrap;
    line-height: 1.5;
    font-size: 14px;
    color: #606266;
  }
  
  .ai-description {
    background-color: #f0f9ff;
    border-color: #bfdbfe;
    color: #1e40af;
  }
}

// 增强的分析进度对话框样式
.enhanced-analysis-progress {
  .requirement-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding: 16px;
    background: linear-gradient(135deg, #f8fafc 0%, #e3f2fd 100%);
    border-radius: 12px;
    border: 1px solid #e1e8ed;
    
    .req-title {
      display: flex;
      align-items: center;
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      
      .req-icon {
        margin-right: 8px;
        color: #409EFF;
      }
    }
  }
  
  .animated-progress-section {
    .main-progress {
      margin-bottom: 24px;
      
      .progress-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 12px;
        
        .progress-percent {
          font-size: 24px;
          font-weight: 700;
          color: #409EFF;
        }
        
        .stage-name {
          font-size: 14px;
          color: #606266;
          font-weight: 500;
        }
      }
    }
    
    .animation-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin: 32px 0;
      
      .stage-icon-container {
        width: 80px;
        height: 80px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3);
        margin-bottom: 16px;
        transition: all 0.3s ease;
        
        .stage-icon {
          font-size: 36px;
          color: white;
          transition: all 0.3s ease;
        }
        
        &.pulse {
          animation: pulse 2s infinite;
        }
        
        &.scan {
          animation: scan 1.5s ease-in-out infinite;
        }
        
        &.loading {
          animation: rotate 1s linear infinite;
        }
        
        &.network {
          animation: network 2s ease-in-out infinite;
        }
        
        &.brain {
          background: linear-gradient(135deg, #722ED1 0%, #B37FEB 100%);
          animation: brain 1.8s ease-in-out infinite;
        }
        
        &.api {
          background: linear-gradient(135deg, #13C2C2 0%, #36CFC9 100%);
          animation: api 1.2s ease-in-out infinite;
        }
        
        &.success {
          background: linear-gradient(135deg, #52C41A 0%, #73D13D 100%);
          animation: success 0.8s ease-out;
        }
        
        &.error {
          background: linear-gradient(135deg, #F56C6C 0%, #FF7875 100%);
          animation: error 0.5s ease-out;
        }
      }
      
      .status-text {
        text-align: center;
        
        .primary-text {
          font-size: 16px;
          color: #303133;
          margin: 0 0 8px 0;
          font-weight: 500;
        }
        
        .secondary-text {
          font-size: 14px;
          color: #909399;
          margin: 0;
        }
      }
    }
    
    .stage-timeline {
      margin-top: 32px;
      
      h4 {
        font-size: 16px;
        color: #303133;
        margin: 0 0 16px 0;
        text-align: center;
      }
      
      .timeline-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
        
        .stage-desc {
          font-size: 14px;
          color: #606266;
        }
      }
    }
  }
  
  .analysis-features {
    margin-top: 32px;
    padding: 20px;
    background: #fafbfc;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
    
    .feature-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;
      
      .feature-item {
        display: flex;
        align-items: center;
        font-size: 14px;
        color: #606266;
        
        .el-icon {
          margin-right: 8px;
          color: #409EFF;
        }
      }
    }
  }
}

// 分析结果样式 - 框架一致风格
.analysis-result-section {
  .result-header {
    text-align: center;
    margin-bottom: 16px;
  }
  
  // 改进亮点概览 - 紧凑设计
  .improvements-highlight {
    margin-bottom: 16px;
    
    .highlight-card {
      .highlight-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        
        .highlight-title {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 16px;
          font-weight: 600;
          color: #303133;
          
          .highlight-icon {
            color: #E6A23C;
            font-size: 18px;
          }
        }
      }
      
      .highlight-content {
        .highlight-item {
          display: flex;
          align-items: center;
          gap: 12px;
          padding: 12px;
          background: #f5f7fa;
          border: 1px solid #e4e7ed;
          border-radius: 6px;
          margin-bottom: 8px;
          
          .item-icon {
            font-size: 16px;
            
            &.success {
              color: #67c23a;
            }
            
            &.info {
              color: #409eff;
            }
            
            &.warning {
              color: #e6a23c;
            }
          }
          
          .item-text {
            flex: 1;
            
            .item-title {
              display: block;
              font-size: 14px;
              color: #303133;
              font-weight: 500;
              margin-bottom: 2px;
            }
            
            .item-desc {
              display: block;
              font-size: 12px;
              color: #909399;
            }
          }
        }
        
        .adoption-recommendation {
          margin-top: 12px;
          padding: 12px;
          border-radius: 6px;
          border: 1px solid;
          
          &.recommend {
            background: #f0f9ff;
            border-color: #409eff;
          }
          
          &.caution {
            background: #fdf6ec;
            border-color: #e6a23c;
          }
          
          .adoption-content {
            display: flex;
            align-items: center;
            gap: 8px;
            
            .adoption-icon {
              font-size: 16px;
              
              &.success {
                color: #67c23a;
              }
              
              &.warning {
                color: #e6a23c;
              }
            }
            
            .adoption-text {
              font-size: 14px;
              color: #303133;
            }
          }
        }
      }
    }
  }
  
  // 详细对比结果
  .result-comparison {
    .comparison-content {
      .comparison-card {
        &.original {
          border: 1px solid #dcdfe6;
        }
        
        &.optimized {
          border: 1px solid #409eff;
        }
        
        .card-header {
          span {
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 14px;
            font-weight: 600;
            color: #303133;
          }
        }
        
        .card-content {
          padding: 16px;
        }
      }
      
      .arrow-col {
        display: flex;
        align-items: center;
        justify-content: center;
        
        .comparison-arrow {
          display: flex;
          flex-direction: column;
          align-items: center;
          justify-content: center;
          
          .arrow-icon {
            font-size: 24px;
            color: #409eff;
            margin-bottom: 4px;
          }
          
          .arrow-text {
            font-size: 12px;
            color: #909399;
          }
        }
      }
      
      .analysis-summary {
        margin-top: 16px;
      }
    }
    
    .improvements-content {
      .improvement-detail {
        h4 {
          margin: 0 0 12px 0;
          font-size: 16px;
          color: #303133;
          font-weight: 600;
        }
        
        .before-after {
          margin-bottom: 12px;
          
          .change-item {
            margin-bottom: 8px;
            
            .label {
              font-size: 12px;
              color: #909399;
              font-weight: 500;
              margin-right: 8px;
            }
            
            .original-text {
              color: #606266;
            }
            
            .optimized-text {
              color: #409eff;
              font-weight: 500;
            }
          }
        }
        
        .new-content, .classification, .value-assessment {
          margin-bottom: 12px;
          
          .label {
            font-size: 12px;
            color: #909399;
            font-weight: 500;
            margin-right: 8px;
          }
          
          .optimized-text {
            color: #409eff;
            font-weight: 500;
            margin-top: 4px;
          }
        }
      }
    }
  }
}

// 动画定义
@keyframes pulse {
  0%, 100% { transform: scale(1); box-shadow: 0 8px 25px rgba(102, 126, 234, 0.3); }
  50% { transform: scale(1.05); box-shadow: 0 12px 30px rgba(102, 126, 234, 0.5); }
}

@keyframes scan {
  0%, 100% { transform: scale(1) rotate(0deg); }
  25% { transform: scale(1.02) rotate(2deg); }
  75% { transform: scale(1.02) rotate(-2deg); }
}

@keyframes rotate {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

@keyframes network {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.7; transform: scale(1.03); }
}

@keyframes brain {
  0%, 100% { transform: scale(1); filter: hue-rotate(0deg); }
  33% { transform: scale(1.02); filter: hue-rotate(10deg); }
  66% { transform: scale(1.01); filter: hue-rotate(-10deg); }
}

@keyframes api {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-3px); }
}

@keyframes success {
  0% { transform: scale(0.8); opacity: 0.8; }
  50% { transform: scale(1.1); opacity: 1; }
  100% { transform: scale(1); opacity: 1; }
}

@keyframes error {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-2px); }
  75% { transform: translateX(2px); }
}

// 编辑表单样式优化
:deep(.el-form-item__label) {
  font-weight: 600;
  color: #303133;
}

:deep(.el-input__wrapper) {
  border-radius: 6px;
}

:deep(.el-select .el-input__wrapper) {
  border-radius: 6px;
}

:deep(.el-input-number .el-input__wrapper) {
  border-radius: 6px;
}

// 对话框样式优化
:deep(.el-dialog) {
  border-radius: 12px;
  
  .el-dialog__header {
    padding: 20px 24px 16px;
    border-bottom: 1px solid #e4e7ed;
    
    .el-dialog__title {
      font-size: 18px;
      font-weight: 600;
      color: #303133;
    }
  }
  
  .el-dialog__body {
    padding: 24px;
  }
  
  .el-dialog__footer {
    padding: 16px 24px 20px;
    border-top: 1px solid #e4e7ed;
  }
}

// 保持框架默认按钮样式

// 批量操作标签样式
.batch-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  margin-right: 8px;
}

// 任务管理标签样式
.task-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  margin-right: 8px;
  background-color: #f0f0f0; /* 添加背景色方便调试 */
  padding: 2px 4px;
  border-radius: 4px;
}

// 进度条样式优化
:deep(.el-progress) {
  .el-progress-bar__outer {
    border-radius: 10px;
    background-color: #f0f2f5;
  }
  
  .el-progress-bar__inner {
    border-radius: 10px;
    background: linear-gradient(90deg, #409eff 0%, #67c23a 100%);
    transition: all 0.3s ease;
  }
}

// 描述列表样式优化
:deep(.el-descriptions) {
  .el-descriptions__label {
    font-weight: 600;
    color: #303133;
    background-color: #f8f9fa;
  }
  
  .el-descriptions__content {
    color: #606266;
  }
}

// 标签样式优化
:deep(.el-tag) {
  border-radius: 12px;
  font-weight: 500;
  border: none;
}

// 视图切换标签样式
.view-label {
  font-size: 14px;
  color: #606266;
  font-weight: 500;
  margin-right: 8px;
}

// 树形视图摘要样式
.tree-summary {
  margin: 16px 0;
  
  .el-alert {
    border-radius: 8px;
  }
  
  .el-alert__content {
    font-size: 14px;
  }
}

// 视图切换过渡动画
.view-fade-enter-active,
.view-fade-leave-active {
  transition: all 0.4s cubic-bezier(0.4, 0, 0.2, 1);
}

.view-fade-enter-from {
  opacity: 0;
  transform: translateY(20px) scale(0.98);
}

.view-fade-leave-to {
  opacity: 0;
  transform: translateY(-20px) scale(0.98);
}

.view-fade-enter-to,
.view-fade-leave-from {
  opacity: 1;
  transform: translateY(0) scale(1);
}

// 视图切换加载遮罩
.view-switching-overlay {
  position: relative;
  min-height: 500px;
  padding: 40px 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  border-radius: 12px;
  border: 2px solid #e4e7ed;
  
  .switching-text {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 16px;
    color: #606266;
    font-size: 16px;
    font-weight: 500;
    z-index: 10;
    
    .el-icon {
      font-size: 32px;
      color: #409eff;
      margin-bottom: 8px;
    }
    
    .is-loading {
      animation: rotating 2s linear infinite;
    }
    
    span {
      text-align: center;
      line-height: 1.5;
      color: #303133;
    }
  }
  
  // 为skeleton添加透明度
  :deep(.el-skeleton) {
    opacity: 0.6;
    
    .el-skeleton__item {
      background: linear-gradient(90deg, #f0f2f5 25%, #e6e8eb 50%, #f0f2f5 75%);
      background-size: 400% 100%;
      animation: loading 1.4s ease infinite;
    }
  }
}

@keyframes rotating {
  0% { transform: rotate(0deg); }
  100% { transform: rotate(360deg); }
}

@keyframes loading {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

// L4生成进度对话框样式
.enhanced-l4-generation-progress {
  .requirement-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding: 16px;
    background: linear-gradient(135deg, #f8fafc 0%, #e3f2fd 100%);
    border-radius: 12px;
    border: 1px solid #e1e8ed;
    
    .req-title {
      display: flex;
      align-items: center;
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      
      .req-icon {
        margin-right: 8px;
        color: #722ED1;
      }
    }
  }
  
  .animated-progress-section {
    .main-progress {
      margin-bottom: 24px;
      
      .progress-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 12px;
        
        .progress-percent {
          font-size: 24px;
          font-weight: 700;
          color: #722ED1;
        }
        
        .stage-name {
          font-size: 14px;
          color: #606266;
          font-weight: 500;
        }
      }
    }
    
    .animation-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin: 32px 0;
      
      .stage-icon-container {
        width: 80px;
        height: 80px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: linear-gradient(135deg, #722ED1 0%, #B37FEB 100%);
        box-shadow: 0 8px 25px rgba(114, 46, 209, 0.3);
        margin-bottom: 16px;
        transition: all 0.3s ease;
        
        .stage-icon {
          font-size: 36px;
          color: white;
          transition: all 0.3s ease;
        }
        
        &.brain {
          animation: brain 1.8s ease-in-out infinite;
        }
        
        &.pulse {
          animation: pulse 2s infinite;
        }
        
        &.scan {
          animation: scan 1.5s ease-in-out infinite;
        }
        
        &.loading {
          animation: rotate 1s linear infinite;
        }
        
        &.network {
          animation: network 2s ease-in-out infinite;
        }
        
        &.api {
          background: linear-gradient(135deg, #13C2C2 0%, #36CFC9 100%);
          animation: api 1.2s ease-in-out infinite;
        }
        
        &.success {
          background: linear-gradient(135deg, #52C41A 0%, #73D13D 100%);
          animation: success 0.8s ease-out;
        }
        
        &.error {
          background: linear-gradient(135deg, #F56C6C 0%, #FF7875 100%);
          animation: error 0.5s ease-out;
        }
      }
      
      .status-text {
        text-align: center;
        
        .primary-text {
          font-size: 16px;
          color: #303133;
          margin: 0 0 8px 0;
          font-weight: 500;
        }
        
        .secondary-text {
          font-size: 14px;
          color: #909399;
          margin: 0;
        }
      }
    }
    
    .stage-timeline {
      margin-top: 32px;
      
      h4 {
        font-size: 16px;
        color: #303133;
        margin: 0 0 16px 0;
        text-align: center;
      }
      
      .timeline-content {
        display: flex;
        justify-content: space-between;
        align-items: center;
        
        .stage-desc {
          font-size: 14px;
          color: #606266;
        }
      }
    }
  }
  
  .analysis-features {
    margin-top: 32px;
    padding: 20px;
    background: #fafbfc;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
    
    .feature-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;
      
      .feature-item {
        display: flex;
        align-items: center;
        font-size: 14px;
        color: #606266;
        
        .el-icon {
          margin-right: 8px;
          color: #722ED1;
        }
      }
    }
  }
}

// L4生成进度对话框样式
.l4-generation-progress {
  .task-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding-bottom: 16px;
    border-bottom: 1px solid #f0f0f0;
    
    .task-title {
      display: flex;
      align-items: center;
      font-size: 16px;
      font-weight: 500;
      color: #303133;
      
      .task-icon {
        margin-right: 8px;
        color: #409eff;
        font-size: 18px;
      }
    }
  }
  
  .progress-section {
    .main-progress {
      margin-bottom: 24px;
    }
    
    .progress-info {
      .progress-text {
        text-align: center;
        margin-bottom: 24px;
        color: #606266;
        font-size: 14px;
        
        .error-text {
          color: #f56c6c;
        }
      }
      
      .progress-steps {
        display: flex;
        justify-content: space-between;
        padding: 0 20px;
        
        .step-item {
          display: flex;
          flex-direction: column;
          align-items: center;
          opacity: 0.5;
          transition: all 0.3s ease;
          
          &.active {
            opacity: 1;
            color: #409eff;
          }
          
          .el-icon {
            font-size: 24px;
            margin-bottom: 8px;
          }
          
          span {
            font-size: 12px;
            text-align: center;
          }
        }
      }
    }
  }
}

.l4-generation-result {
  .result-summary {
    margin-bottom: 24px;
    padding: 20px;
    background-color: #f8f9fa;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
  }
  
  .error-summary {
    margin-bottom: 24px;
    
    .error-list {
      margin-top: 8px;
      
      .error-item {
        padding: 4px 0;
        font-size: 14px;
        color: #e6a23c;
        border-left: 3px solid #e6a23c;
        padding-left: 8px;
        margin-bottom: 4px;
      }
    }
  }
  
  .no-results {
    text-align: center;
    padding: 40px 20px;
    
    .failure-reasons {
      text-align: left;
      max-width: 600px;
      margin: 16px auto 0;
      padding: 0;
      list-style: none;
      
      li {
        padding: 8px 0;
        border-bottom: 1px solid #f0f0f0;
        
        &:last-child {
          border-bottom: none;
        }
      }
    }
  }
  
  .result-table {
    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      
      h4 {
        margin: 0;
        color: #303133;
        font-size: 16px;
        font-weight: 500;
      }
      
      .table-actions {
        display: flex;
        gap: 8px;
      }
    }
    
    .l3-requirement-info {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .l3-title {
        flex: 1;
        font-size: 14px;
        color: #606266;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
  }
  
  .result-table {
    .table-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 16px;
      
      h4 {
        margin: 0;
        color: #303133;
        font-size: 16px;
        font-weight: 500;
      }
      
      .table-actions {
        display: flex;
        gap: 8px;
      }
    }
    
    .l3-requirement-info {
      display: flex;
      align-items: center;
      gap: 8px;
      
      .l3-title {
        flex: 1;
        font-size: 14px;
        color: #606266;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }
    }
    
    .requirement-title {
      display: flex;
      align-items: center;
      gap: 8px;
      
      span {
        flex: 1;
      }
    }
  }
}

// 增强的Mermaid流程图生成进度对话框样式
.enhanced-mermaid-generation-progress {
  .requirement-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 24px;
    padding: 16px;
    background: linear-gradient(135deg, #f8fafc 0%, #e3f8f8 100%);
    border-radius: 12px;
    border: 1px solid #e1e8ed;
    
    .req-title {
      display: flex;
      align-items: center;
      font-size: 16px;
      font-weight: 600;
      color: #303133;
      
      .req-icon {
        margin-right: 8px;
        color: #13C2C2;
      }
    }
  }
  
  .animated-progress-section {
    .main-progress {
      margin-bottom: 24px;
      
      .progress-info {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 12px;
        
        .progress-percent {
          font-size: 24px;
          font-weight: 700;
          color: #13C2C2;
        }
        
        .stage-name {
          font-size: 14px;
          color: #606266;
          font-weight: 500;
        }
      }
    }
    
    .animation-section {
      display: flex;
      flex-direction: column;
      align-items: center;
      margin: 32px 0;
      
      .stage-icon-container {
        width: 80px;
        height: 80px;
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        background: linear-gradient(135deg, #13C2C2 0%, #36CFC9 100%);
        box-shadow: 0 8px 25px rgba(19, 194, 194, 0.3);
        margin-bottom: 16px;
        transition: all 0.3s ease;
        
        .stage-icon {
          font-size: 36px;
          color: white;
          transition: all 0.3s ease;
        }
        
        &.coordinate {
          animation: coordinate 1.8s ease-in-out infinite;
        }
        
        &.processing {
          animation: processing 1.5s ease-in-out infinite;
        }
        
        &.pulse {
          animation: pulse 2s infinite;
        }
        
        &.scan {
          animation: scan 1.5s ease-in-out infinite;
        }
        
        &.loading {
          animation: rotate 1s linear infinite;
        }
        
        &.success {
          background: linear-gradient(135deg, #52C41A 0%, #73D13D 100%);
          animation: success 0.8s ease-out;
        }
        
        &.error {
          background: linear-gradient(135deg, #F56C6C 0%, #FF7875 100%);
          animation: error 0.5s ease-out;
        }
      }
      
      .status-text {
        text-align: center;
        
        .primary-text {
          font-size: 16px;
          color: #303133;
          margin: 0 0 8px 0;
          font-weight: 500;
        }
        
        .secondary-text {
          font-size: 14px;
          color: #909399;
          margin: 0;
        }
      }
    }
    
    .generation-statistics {
      margin-top: 32px;
      padding: 20px;
      background: #f8f9fa;
      border-radius: 8px;
      border: 1px solid #e4e7ed;
      
      .stat-item {
        display: flex;
        align-items: center;
        
        .stat-icon {
          margin-right: 12px;
          color: #13C2C2;
          font-size: 18px;
        }
        
        .stat-info {
          .stat-value {
            font-size: 18px;
            font-weight: 600;
            color: #303133;
            line-height: 1;
          }
          
          .stat-label {
            font-size: 12px;
            color: #909399;
            margin-top: 2px;
          }
        }
      }
    }
  }
  
  .analysis-features {
    margin-top: 32px;
    padding: 20px;
    background: #fafbfc;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
    
    .feature-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;
      
      .feature-item {
        display: flex;
        align-items: center;
        font-size: 14px;
        color: #606266;
        
        .el-icon {
          margin-right: 8px;
          color: #13C2C2;
        }
      }
    }
  }
}

// Mermaid生成结果样式
.mermaid-generation-result {
  .result-summary {
    margin-bottom: 24px;
    padding: 20px;
    background-color: #f8f9fa;
    border-radius: 8px;
    border: 1px solid #e4e7ed;
  }
  
  .error-summary {
    margin-bottom: 24px;
    
    .error-details {
      ul {
        margin: 8px 0;
        padding-left: 20px;
        
        li {
          margin-bottom: 4px;
          color: #e6a23c;
        }
      }
    }
  }
  
  .generation-summary {
    margin-top: 20px;
    
    .summary-header {
      span {
        display: flex;
        align-items: center;
        gap: 6px;
        font-size: 14px;
        font-weight: 600;
        color: #303133;
      }
    }
  }
}

// Mermaid特有动画
@keyframes coordinate {
  0%, 100% { transform: scale(1) rotate(0deg); filter: hue-rotate(0deg); }
  33% { transform: scale(1.02) rotate(5deg); filter: hue-rotate(15deg); }
  66% { transform: scale(1.01) rotate(-5deg); filter: hue-rotate(-15deg); }
}

@keyframes processing {
  0%, 100% { transform: translateY(0) scale(1); }
  50% { transform: translateY(-3px) scale(1.02); }
}
</style>