package worker

import (
	"fmt"
	"time"

	"github.com/sinksmell/bee-crontab/models"
	"github.com/sinksmell/bee-crontab/models/common"
)

// Scheduler 调度器 用来调度worker工作
type Scheduler struct {
	JobEventChan      chan *models.JobEvent               // 任务事件管道
	JobPlanTable      map[string]*models.JobSchedulerPlan // 任务计划表
	JobExecTable      map[string]*models.JobExecInfo      // 正在执行的任务
	JobExecResultChan chan *models.JobExecResult          // 任务执行结果
}

var (
	// BeeScheduler worker调度器单例
	BeeScheduler *Scheduler
)

// InitScheduler 初始化调度器单例
func InitScheduler() (err error) {

	BeeScheduler = &Scheduler{
		JobEventChan:      make(chan *models.JobEvent, 1000),
		JobPlanTable:      make(map[string]*models.JobSchedulerPlan),
		JobExecTable:      make(map[string]*models.JobExecInfo),
		JobExecResultChan: make(chan *models.JobExecResult, 1000),
	}

	// 启动调度器协程
	go BeeScheduler.DoScheduler()

	return
}

// DoScheduler 调度监听
func (scheduler *Scheduler) DoScheduler() {
	var (
		event    *models.JobEvent      // 任务事件
		duration time.Duration         // 距离下次任务到期时间
		timer    *time.Timer           // 定时器
		result   *models.JobExecResult // 任务执行结果
	)

	// 获取距离下次任务开始 的时间间隔
	duration = scheduler.TryScheduler()
	// 调度定时器
	timer = time.NewTimer(duration)
	// 这里可以让调度器精准睡眠一会
	time.Sleep(duration)
	// 调度循环
	for {
		select {
		case event = <-scheduler.JobEventChan:
			//任务事件传来
			scheduler.HandleJobEvent(event)
		case <-timer.C:
			// 定时器到期
		case result = <-scheduler.JobExecResultChan:
			//任务处理结果传来
			scheduler.HandleJobResult(result)
		}
		// 尝试下一次调度
		duration = scheduler.TryScheduler()
		// 重置定时器
		timer.Reset(duration)
	}

}

// TryScheduler 尝试调度 返回距离最近到期任务的时间间隔
func (scheduler *Scheduler) TryScheduler() (duration time.Duration) {
	var (
		plan *models.JobSchedulerPlan // 任务计划表
		now  time.Time                // 当前时间
		near *time.Time               //最近任务到期时间

	)
	if len(scheduler.JobPlanTable) == 0 {
		fmt.Println("没有任务呢！")
		// 如果现在没有计划中的任务
		// 就返回一秒  让调度器睡1秒
		duration = time.Second
		return
	}

	now = time.Now()
	// 遍历所有任务
	for _, plan = range scheduler.JobPlanTable {
		if plan.NextTime.Before(now) || plan.NextTime.Equal(now) {
			// 任务到期 尝试执行任务
			// 注意 上次任务可能还没有结束执行
			scheduler.TryRunJob(plan)
			plan.NextTime = plan.Expr.Next(now)
		}

		// 统计一个最近要到期的时间
		if near == nil || plan.NextTime.Before(*near) {
			near = &plan.NextTime
		}
	}

	// 获取时间间隔
	duration = near.Sub(now)
	return
}

// TryRunJob 尝试执行任务
func (scheduler *Scheduler) TryRunJob(plan *models.JobSchedulerPlan) {
	// 调度与执行是两码事
	// 例如 每5秒钟调度一次 但是执行一次需要一分钟
	// 接受了调度 如果上次任务还没有执行结束那么就不能执行该任务

	var (
		info      *models.JobExecInfo // 任务执行信息
		isRunning bool                // 标记任务是否正在执行
	)

	if info, isRunning = scheduler.JobExecTable[plan.Job.Name]; isRunning {
		fmt.Println("任务正在执行,尚未退出! ", plan.Job.Name)
		// 直接退出
		return
	}

	// 构建任务运行信息
	info = models.NewJobExecInfo(plan)
	// 保存到运行表
	scheduler.JobExecTable[plan.Job.Name] = info
	// 执行任务
	BeeCronExecutor.ExecuteJob(info)
}

// PushJobEvent 推送任务变化事件
func (scheduler *Scheduler) PushJobEvent(event *models.JobEvent) {
	scheduler.JobEventChan <- event
}

// HandleJobEvent 处理任务变化事件
func (scheduler *Scheduler) HandleJobEvent(event *models.JobEvent) {

	var (
		plan    *models.JobSchedulerPlan
		info    *models.JobExecInfo
		isExist bool
		err     error
	)

	switch event.EventType {
	case common.JOB_EVENT_SAVE:
		// 保存任务事件
		// 解析job 放到planTable中
		if plan, err = models.NewJobSchedulerPlan(event.Job); err != nil {
			// 说明任务解析cron 表达式可能出问题,直接退出
			//fmt.Println(err)
			return
		}
		scheduler.JobPlanTable[event.Job.Name] = plan
	case common.JOB_EVENT_DELETE:
		// 删除任务事件
		// 如果任务还存在 则从计划表中删除
		if plan, isExist = scheduler.JobPlanTable[event.Job.Name]; isExist {
			delete(scheduler.JobPlanTable, event.Job.Name)
		}
	case common.JOB_EVENT_KILL:
		// 强杀任务事件
		// 如果任务正在运行 则杀死它
		if info, isExist = scheduler.JobExecTable[event.Job.Name]; isExist {
			// 执行 cancelFunc 终止程序运行
			info.CancelFunc()
			fmt.Println("杀死任务 ", info.Job.Name)
		}
	}

}

// PushJobResult 推送任务执行结果
func (scheduler *Scheduler) PushJobResult(result *models.JobExecResult) {
	scheduler.JobExecResultChan <- result
}

// HandleJobResult 处理任务执行结果
func (scheduler *Scheduler) HandleJobResult(result *models.JobExecResult) {

	var (
		log   *models.JobExecLog
		shift int64 = 1000000
	)

	// 从执行表中删除对应的任务
	delete(scheduler.JobExecTable, result.ExecInfo.Job.Name)

	// UnixNano 默认是纳秒 这里/1000转换成微秒
	log = &models.JobExecLog{
		JobName:      result.ExecInfo.Job.Name,
		Command:      result.ExecInfo.Job.Command,
		Output:       string(result.Output),
		PlanTime:     result.ExecInfo.PlanTime.UnixNano() / shift,
		ScheduleTime: result.ExecInfo.RealTime.UnixNano() / shift,
		StartTime:    result.StartTime.UnixNano() / shift,
		EndTime:      result.EndTime.UnixNano() / shift,
	}
	// 错误要单独判断是否为空
	if result.Err != nil {
		log.Err = result.Err.Error()
	} else {
		log.Err = "OK"
	}
	BeeCronLogger.logChan <- log
	fmt.Println("任务执行完成: ")
	fmt.Println(result)
}
