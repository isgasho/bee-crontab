package worker

import (
	"github.com/sinksmell/bee-crontab/models"
	"os/exec"
	"time"
	"math/rand"
	"sync"
	"github.com/sinksmell/bee-crontab/models/common"
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
		var(
			wg sync.WaitGroup	// 用于等待任务执行结束
			timer *time.Timer	// 任务执行定时器
			sigchan  chan struct{}	// 任务执行结束消息管道
			timeLimit  time.Duration
		)
		wg.Add(1)
		// 为了防止任务定时不精准 宽限10%的时间
		timeLimit=time.Duration(info.Job.TimeOut)*1000*time.Millisecond
		sigchan=make(chan struct{},1)

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
			timer=time.NewTimer(timeLimit)
			// 执行并捕获输出
			// 启动协程执行任务 外部定时
			go func() {
				timer.Reset(timeLimit)
				output, err = cmd.CombinedOutput()
				result.EndTime = time.Now()
				result.Output = output
				result.Err = err
				sigchan<- struct{}{}
				wg.Done()
			}()
			// TODO 根据超时时间判断，如果超时那么强制杀死任务

			for  {
				select {
				case <-timer.C:
					// 定时器到期 任务执行超时
					info.CancelFunc()
					wg.Done()
					result.Type=common.RES_TIMEOUT
					result.Output=[]byte("timeout!")
					goto WAIT
				case <-sigchan:
					// 在限制时间内执行完成
					result.Type=common.RES_SUCCESS
					goto WAIT
				}
			}
			WAIT:
			wg.Wait()
		}

		// 任务执行结束 把结果返回给 scheduler
		// 从执行表中删除对应的记录
		Bee_Scheduler.PushJobResult(result)
	}()
}
