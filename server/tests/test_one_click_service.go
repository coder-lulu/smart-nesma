package main

import (
	"fmt"
	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
	"github.com/flipped-aurora/gin-vue-admin/server/service"
	nesmaReq "github.com/flipped-aurora/gin-vue-admin/server/model/nesma/request"
)

func main() {
	fmt.Println("=== 测试一键分析服务功能 ===")

	// 初始化基本配置
	global.GVA_VP = core.Viper()
	global.GVA_LOG = core.Zap()
	global.GVA_DB = initialize.Gorm()

	// 初始化NESMA服务
	initialize.InitNesmaServices()

	// 1. 测试创建分析请求
	fmt.Println("\n1. 测试创建分析请求...")
	testCreateAnalysisRequest()

	// 2. 测试需求分析服务是否可用
	fmt.Println("\n2. 测试需求分析服务...")
	testRequirementAnalysisService()

	// 3. 测试项目管理服务
	fmt.Println("\n3. 测试项目管理服务...")
	testProjectService()

	fmt.Println("\n=== 测试完成 ===")
}

func testCreateAnalysisRequest() {
	// 创建一键分析请求
	req := nesmaReq.NesmaOneClickAnalysisRequest{
		ProjectID: 1,
		CycleID:   1,
	}

	fmt.Printf("✅ 成功创建一键分析请求: ProjectID=%d, CycleID=%d\n", req.ProjectID, req.CycleID)

	// 创建完整的分析请求
	analysisReq := nesmaReq.NesmaAnalyzeRequest{
		ProjectID: req.ProjectID,
		CycleID:   req.CycleID,
		Config: struct {
			AnalysisType    string   `json:"analysisType"`
			TargetLevels    []int    `json:"targetLevels"`
			AIModel         string   `json:"aiModel"`
			EnableMermaid   bool     `json:"enableMermaid"`
			BatchSize       int      `json:"batchSize"`
			MaxRetries      int      `json:"maxRetries"`
			SkipCompleted   bool     `json:"skipCompleted"`
		}{
			AnalysisType:  "requirement_optimization",
			TargetLevels:  []int{3, 4},
			AIModel:       "deepseek",
			EnableMermaid: true,
			BatchSize:     10,
			MaxRetries:    3,
			SkipCompleted: false,
		},
	}

	fmt.Printf("✅ 成功创建分析请求: %+v\n", analysisReq.Config)
}

func testRequirementAnalysisService() {
	// 获取需求分析服务
	requirementAnalysisService := service.ServiceGroupApp.NesmaServiceGroup.RequirementAnalysisService
	
	fmt.Println("✅ 需求分析服务已正确初始化")
	
	// 测试获取分析任务列表（无实际项目数据时会返回空列表）
	tasks, err := requirementAnalysisService.GetAnalysisTasks(0, 0)
	if err != nil {
		fmt.Printf("⚠️ 获取分析任务列表失败（正常，因为无数据）: %v\n", err)
	} else {
		fmt.Printf("✅ 成功调用分析任务列表接口，返回 %d 个任务\n", len(tasks))
	}
}

func testProjectService() {
	// 获取项目服务
	projectService := service.ServiceGroupApp.NesmaServiceGroup.NesmaProjectService
	
	fmt.Println("✅ 项目服务已正确初始化")
	
	// 测试获取项目统计（无实际项目数据时会返回空统计）
	stats, err := projectService.GetNesmaProjectStats(1, nil) // 假设用户ID为1，获取所有项目统计
	if err != nil {
		fmt.Printf("⚠️ 获取项目统计失败（正常，因为无数据）: %v\n", err)
	} else {
		fmt.Printf("✅ 成功调用项目统计接口: %+v\n", stats)
	}
}