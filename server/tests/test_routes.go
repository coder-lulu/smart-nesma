package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/core"
	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/initialize"
)

func main() {
	fmt.Println("=== 测试一键分析API路由 ===")

	// 初始化配置和日志（最小化）
	global.GVA_VP = core.Viper()
	global.GVA_LOG = core.Zap()

	// 初始化路由器
	routers := initialize.Routers()

	// 查看已注册的路由
	routes := routers.Routes()
	
	fmt.Printf("\n🔍 检查一键分析相关路由:\n")
	analysisRouteFound := false
	progressRouteFound := false
	
	for _, route := range routes {
		if route.Path == "/api/v1/nesma/project/one-click-analysis" && route.Method == "POST" {
			fmt.Printf("✅ 找到一键分析路由: %s %s\n", route.Method, route.Path)
			analysisRouteFound = true
		}
		if route.Path == "/api/v1/nesma/analysis/progress/:taskId" && route.Method == "GET" {
			fmt.Printf("✅ 找到进度查询路由: %s %s\n", route.Method, route.Path)
			progressRouteFound = true
		}
		if route.Path == "/api/v1/nesma/analysis/analyze" && route.Method == "POST" {
			fmt.Printf("✅ 找到分析启动路由: %s %s\n", route.Method, route.Path)
		}
	}

	if analysisRouteFound {
		fmt.Println("✅ 一键分析路由配置正确")
	} else {
		fmt.Println("❌ 一键分析路由未找到")
	}

	if progressRouteFound {
		fmt.Println("✅ 进度查询路由配置正确")
	} else {
		fmt.Println("❌ 进度查询路由未找到")
	}

	// 启动一个临时服务器测试路由可达性
	fmt.Println("\n🚀 启动临时服务器测试路由...")
	
	go func() {
		routers.Run(":8889") // 使用不同端口避免冲突
	}()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试路由可达性
	testRoutes()

	fmt.Println("\n=== 测试完成 ===")
}

func testRoutes() {
	baseURL := "http://localhost:8889"
	
	// 测试一键分析路由是否可达
	testURL := baseURL + "/api/v1/nesma/project/one-click-analysis"
	resp, err := http.Get(testURL)
	if err != nil {
		fmt.Printf("❌ 路由不可达: %v\n", err)
		return
	}
	resp.Body.Close()
	
	// 即使返回405 (Method Not Allowed)，说明路由存在但方法不对
	if resp.StatusCode == 405 {
		fmt.Printf("✅ 一键分析路由可达 (返回405说明需要POST方法)\n")
	} else {
		fmt.Printf("📊 一键分析路由状态码: %d\n", resp.StatusCode)
	}

	// 测试进度查询路由
	progressURL := baseURL + "/api/v1/nesma/analysis/progress/1"
	resp2, err := http.Get(progressURL)
	if err != nil {
		fmt.Printf("❌ 进度查询路由不可达: %v\n", err)
		return
	}
	resp2.Body.Close()
	
	if resp2.StatusCode != 404 {
		fmt.Printf("✅ 进度查询路由可达 (状态码: %d)\n", resp2.StatusCode)
	} else {
		fmt.Printf("❌ 进度查询路由返回404\n")
	}
}