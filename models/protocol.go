package models

import (
	"github.com/gorhill/cronexpr"
	"time"
	"context"
	"encoding/json"
	"strings"
	"github.com/sinksmell/bee-crontab/models/common"
	"fmt"
)

// 任务
type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

//任务调度计划
type JobSchedulerPlan struct {
	Job      *Job                 // 要调度的任务
	Expr     *cronexpr.Expression //解析好的 要执行的cronExpr
	NextTime time.Time            // 下次调度时间
}

// 任务执行信息
type JobExecInfo struct {
	Job        *Job               // 正在执行的任务
	PlanTime   time.Time          //	计划调度时间
	RealTime   time.Time          //	实际调度时间
	CancelCtx  context.Context    //用于取消任务的上下文
	CancelFunc context.CancelFunc //取消方法
}

// 开始调度的时间 与开始执行的时间是不一样的
// 调度时间之差 反映了调度器的效率
// 执行开始与结束时间之差 为程序运行时间

// 任务执行结果
type JobExecResult struct {
	ExecInfo  *JobExecInfo // 执行状态
	Output    []byte       // 输出结果
	Err       error        // 错误信息
	StartTime time.Time    // 开始运行时间
	EndTime   time.Time    // 结束时间
}

// 任务执行日志
type JobExecLog struct {
	JobName      string ` json:"jobName" bson:"jobName"`          //任务名
	Command      string ` json:"command" bson:"command"`          //执行命令
	Err          string `json:"err" bson:"err"`                   //错误信息
	Output       string `json:"output" bson:"output"`             //任务输出
	PlanTime     int64  `json:"planTime" bson:"planTime"`         // 计划开始时间 时间戳
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"` // 实际调度时间
	StartTime    int64  `json:"startTime" bson:"startTime"`       // 开始运行时间
	EndTime      int64  `json:"endTime" bson:"endTime"`           // 结束运行时间
}

// 日志缓存 批量插入任务日志 提交吞吐效率
// 当buffer满了或者定时器时间到了 执行插入操作

type LogBuffer struct {
	Logs []interface{} // 任务日志集合
}

// 任务变化事件
type JobEvent struct {
	EventType uint //事件类型 在常量中有定义
	Job       *Job // 任务
}

// 通用的返回类型
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 生产response 的方法
func NewResponse(code int, msg string, data interface{}) Response {
	return Response{code, msg, data}
}

// 反序列化得到Job
func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job Job
	)
	if err = json.Unmarshal(value, &job); err != nil {
		return
	}
	ret = &job
	return
}

//构造任务变化时间
func NewJobEvent(eType uint, job *Job) (*JobEvent) {
	return &JobEvent{eType, job}
}

// 构造执行计划
func NewJobSchedulerPlan(job *Job) (plan *JobSchedulerPlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}
	// 构造调度计划对象
	plan = &JobSchedulerPlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}

	return
}

// 构造一个执行状态
func NewJobExecInfo(plan *JobSchedulerPlan) (info *JobExecInfo) {
	// 创建可以取消的上下文
	ctx, cancelFunc := context.WithCancel(context.TODO())
	info = &JobExecInfo{
		Job:        plan.Job,
		PlanTime:   plan.NextTime,
		RealTime:   time.Now(),
		CancelCtx:  ctx,
		CancelFunc: cancelFunc,
	}
	return
}

//从etcd key中提取出Job Name
func ExtractJobName(key string) string {
	return strings.TrimPrefix(key, common.JOB_SAVE_PATH)
}

//从etcd killer 的key中提取JobName
func ExtractKillerName(key string) string {
	return strings.TrimPrefix(key, common.JOB_KILLER_PATH)
}

// 从etcd /cron/worker/ip 中获取 worker 的ip
func ExtarctWorkerIP(key string) string {
	return strings.TrimPrefix(key, common.JOB_WORKER_PATH)
}

// jobEvent 的toString方法
func (event *JobEvent) String() string {
	return fmt.Sprintf("%d %+v\n", event.EventType, *(event.Job))
}

/*
	ExecInfo  *JobExecInfo // 执行状态
	Output    []byte       // 输出结果
	Err       error        // 错误信息
	StartTime time.Time    // 开始运行时间
	EndTime   time.Time    // 结束时间

*/
//JobExecResult的执行结果
func (result *JobExecResult) String() string {
	return fmt.Sprintf("JobName:%s\n Output:%s\n Err:%v",
		result.ExecInfo.Job.Name,
		string(result.Output),
		result.Err)
}
