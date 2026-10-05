package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

func main() {
	fmt.Println("=== 测试一键分析功能 ===")

	// 测试API端点
	baseURL := "http://localhost:8888"

	// 1. 测试一键分析启动
	fmt.Println("\n1. 测试一键分析启动...")
	testOneClickAnalysis(baseURL)

	// 2. 测试分析进度查询 
	fmt.Println("\n2. 测试分析进度查询...")
	testAnalysisProgress(baseURL)

	fmt.Println("\n=== 测试完成 ===")
}

func testOneClickAnalysis(baseURL string) {
	// 构建请求数据
	reqData := map[string]interface{}{
		"projectId": 1, // 假设项目ID为1
		"cycleId":   1, // 假设周期ID为1
	}

	jsonData, err := json.Marshal(reqData)
	if err != nil {
		fmt.Printf("❌ JSON序列化失败: %v\n", err)
		return
	}

	// 发送POST请求
	url := baseURL + "/api/v1/nesma/project/one-click-analysis"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-token", "test-token") // 模拟认证token

	fmt.Printf("🔄 发送请求到: %s\n", url)
	fmt.Printf("📦 请求数据: %s\n", string(jsonData))

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📊 响应状态码: %d\n", resp.StatusCode)

	// 读取响应
	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	if resp.StatusCode == 200 {
		fmt.Printf("✅ 一键分析启动成功\n")
		if data, ok := respData["data"].(map[string]interface{}); ok {
			fmt.Printf("📋 任务ID: %.0f\n", data["taskId"])
			fmt.Printf("📋 任务状态: %s\n", data["status"])
			fmt.Printf("📋 项目ID: %.0f\n", data["projectId"])
			fmt.Printf("📋 周期ID: %.0f\n", data["cycleId"])
		}
	} else {
		fmt.Printf("❌ 一键分析启动失败: %v\n", respData["msg"])
	}
}

func testAnalysisProgress(baseURL string) {
	// 假设任务ID为1
	taskId := 1
	url := fmt.Sprintf("%s/api/v1/nesma/analysis/progress/%d", baseURL, taskId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("❌ 创建请求失败: %v\n", err)
		return
	}

	req.Header.Set("x-token", "test-token") // 模拟认证token

	fmt.Printf("🔄 查询分析进度: %s\n", url)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("❌ 请求失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("📊 响应状态码: %d\n", resp.StatusCode)

	// 读取响应
	var respData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		fmt.Printf("❌ 解析响应失败: %v\n", err)
		return
	}

	if resp.StatusCode == 200 {
		fmt.Printf("✅ 进度查询成功\n")
		if data, ok := respData["data"].(map[string]interface{}); ok {
			fmt.Printf("📋 任务状态: %s\n", data["status"])
			fmt.Printf("📋 进度: %.0f%%\n", data["progress"])
			if processedCount, ok := data["processedCount"]; ok {
				fmt.Printf("📋 已处理: %.0f\n", processedCount)
			}
			if totalCount, ok := data["totalCount"]; ok {
				fmt.Printf("📋 总数: %.0f\n", totalCount)
			}
		}
	} else {
		fmt.Printf("⚠️ 进度查询响应: %v\n", respData["msg"])
	}
}