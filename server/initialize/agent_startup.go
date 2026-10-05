package initialize

import (
	"context"
	"fmt"
	"time"

	"github.com/flipped-aurora/gin-vue-admin/server/global"
	"github.com/flipped-aurora/gin-vue-admin/server/service/nesma"
	"github.com/flipped-aurora/gin-vue-admin/server/service/system"
)

const initOrderAgentStartup = system.InitOrderExternal + 2

type agentStartup struct{}

// auto run
func init() {
	system.RegisterInit(initOrderAgentStartup, &agentStartup{})
}

func (a *agentStartup) InitializerName() string {
	return "agent_startup"
}

func (a *agentStartup) MigrateTable(ctx context.Context) (context.Context, error) {
	return ctx, nil
}

func (a *agentStartup) TableCreated(ctx context.Context) bool {
	return true
}

func (a *agentStartup) DataInserted(ctx context.Context) bool {
	return true
}

func (a *agentStartup) InitializeData(ctx context.Context) (context.Context, error) {
	// 启动Agent服务
	go func() {
		// 等待数据库初始化完成
		time.Sleep(5 * time.Second)

		// 启动Agent管理器
		manager := nesma.GetAgentManager()

		// 启动Agent监控
		global.GVA_LOG.Info("启动Agent监控服务...")
		go manager.MonitorAgents(context.Background())

		// 启动Agent处理器管理器
		processorManager := nesma.GetAgentProcessorManager()
		global.GVA_LOG.Info("Agent处理器管理器已启动")

		// 启动任务处理协程池
		global.GVA_LOG.Info("启动Agent任务处理服务...")
		go a.startTaskProcessingPool(processorManager)

		global.GVA_LOG.Info("所有Agent服务已成功启动")
	}()

	return ctx, nil
}

// startTaskProcessingPool 启动任务处理协程池
func (a *agentStartup) startTaskProcessingPool(processorManager *nesma.AgentProcessorManager) {
	agentService := nesma.GetAgentService()

	// 创建任务处理协程池
	for i := 0; i < 10; i++ { // 10个工作协程
		go func(workerID int) {
			global.GVA_LOG.Info("启动Agent任务处理工作协程 " + fmt.Sprintf("%d", workerID))

			for {
				// 获取待处理任务
				tasks, err := agentService.GetTasksByStatus(context.Background(), "assigned")
				if err != nil {
					global.GVA_LOG.Error("获取待处理任务失败: " + err.Error())
					time.Sleep(10 * time.Second)
					continue
				}

				// 处理任务
				for _, task := range tasks {
					err := processorManager.ProcessTask(context.Background(), task)
					if err != nil {
						global.GVA_LOG.Error("处理任务失败 - TaskID: " + task.TaskID + ", 错误: " + err.Error())
					}
				}

				// 如果没有任务，等待一段时间
				if len(tasks) == 0 {
					time.Sleep(5 * time.Second)
				}
			}
		}(i)
	}
}
