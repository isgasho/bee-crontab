package worker

import (
	"bee-crontab/models"
	"os/exec"
	"time"
	"math/rand"
)

// 用于执行命令的 执行器
type Executor struct {
}

var (
	Bee_Cron_Executor *Executor
)

// 初始化

func InitExecutor() (err error) {
	Bee_Cron_Executor = &Executor{}
	return
}

// 执行一个任务
func (executor *Executor) ExecuteJob(info *models.JobExecInfo) {
	var (
		cmd    *exec.Cmd
		output []byte
		err    error
		result *models.JobExecResult
		lock   *JobLock
	)

	// 初始化任务结果
	result = &models.JobExecResult{
		ExecInfo: info,
		Output:   make([]byte, 0),
	}


	// 启动协程来处理任务
	go func() {

		// 获取分布式锁
		// 防止任务被并发地调度
		lock = WorkerJobManager.NewLock(info.Job.Name)
		// 记录开始开始抢锁的时间
		result.StartTime = time.Now()

		// 牺牲一点调度的准确性
		// 防止某台机器时间不准导致的资源独占
		// 再锁定资源前 sleep 随机睡眠一小段时间
		// 这里设置的是0-500ms
		time.Sleep(time.Duration(rand.Intn(500)) * time.Millisecond)

		// 锁定资源 defer 延迟释放锁
		err = lock.TryLock()
		defer lock.UnLock()

		if err != nil {
			// 上锁失败
			result.Err = err
			result.EndTime = time.Now()
		} else {
			// 重置开始时间
			result.StartTime = time.Now()

			// 初始化shell命令
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)

			// 执行并捕获输出
			output, err = cmd.CombinedOutput()

			// 记录任务结束时间
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}

		// 任务执行结束 把结果返回给 scheduler
		// 从执行表中删除对应的记录
		Bee_Scheduler.PushJobResult(result)
	}()
}
