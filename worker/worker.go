package main

import (
	"github.com/sinksmell/bee-crontab/models/worker"
	"fmt"
	"time"
)

func main() {
	var (
		err error
	)
	// 初始化配置
	if err = worker.InitConfig("worker.json"); err != nil {
		goto ERR
	}

	// 启动日志协程
	if err = worker.InitLogger(); err != nil {
		goto ERR
	}

	// 启动执行器
	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	// 启动调度协程
	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}
	// 启动任务管理器
	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	// 启动服务注册
	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(100 * time.Millisecond)
	}

ERR:
	fmt.Println(err)
}
