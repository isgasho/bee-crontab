package models

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
	"github.com/astaxie/beego"
)

// MasterLogger  master日志管理器
type MasterLogger struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// HTTPJobLog 任务执行日志
// 解析为标准时间的日志结构体
type HTTPJobLog struct {
	JobName      string ` json:"jobName" `     //任务名
	Command      string ` json:"command" `     //执行命令
	Err          string `json:"err" `          //错误信息
	Output       string `json:"output" `       //任务输出
	PlanTime     string `json:"planTime" `     // 计划开始时间
	ScheduleTime string `json:"scheduleTime" ` // 实际调度时间
	StartTime    string `json:"startTime" `    // 开始运行时间
	EndTime      string `json:"endTime" `      // 结束运行时间
}

var (
	// MLogger master节点任务管理器单例
	MLogger *MasterLogger
)

// 初始化
func init() {
	var (
		client     *mongo.Client
		collection *mongo.Collection
		err        error
	)
	mongoUrl := beego.AppConfig.String("mongoURL")
	client, err = mongo.NewClient(options.Client().ApplyURI(mongoUrl))
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err = client.Connect(ctx); err != nil {
		return
	}
	collection = client.Database("cron").Collection("log")
	// 初始化
	MLogger = &MasterLogger{
		client:     client,
		collection: collection,
	}

	return
}

// ReadLog 读取任务的执行日志
func ReadLog(jobName string) (logs []*HTTPJobLog, err error) {

	var (
		log     *JobExecLog
		httpLog *HTTPJobLog
		cursor  *mongo.Cursor
		findOps *options.FindOptions
		filter  *JobFilter
	)

	// 初始化返回结果 防止出现空指针
	logs = make([]*HTTPJobLog, 0)
	// 查找时的选项
	findOps = options.Find()
	findOps.SetLimit(20)
	// 设置过滤器即查找条件
	filter = &JobFilter{
		jobName,
	}
	if cursor, err = MLogger.collection.Find(context.TODO(), filter, findOps); err != nil {
		return
	}
	// 延迟释放游标
	defer cursor.Close(context.TODO())
	// 遍历游标
	for cursor.Next(context.TODO()) {
		log = &JobExecLog{}
		err = cursor.Decode(log)
		if err != nil {
			continue
		}
		httpLog = parseLog(log)
		logs = append(logs, httpLog)
	}
	return
}

func parseLog(jobLog *JobExecLog) (httpLog *HTTPJobLog) {

	// 构造http响应的log
	httpLog = &HTTPJobLog{}
	httpLog.JobName = jobLog.JobName
	httpLog.Command = jobLog.Command
	httpLog.Err = jobLog.Err
	httpLog.Output = jobLog.Output
	// 时间戳转换为时间类型
	// time.Unix (seconds,nanoseconds)
	// 要么传入秒 要么传入纳秒
	// 由于之前获取的时毫秒级别的时间戳 这里将其转换为对应的毫秒
	httpLog.StartTime = time.Unix(0, jobLog.StartTime*int64(time.Millisecond)).String()
	httpLog.EndTime = time.Unix(0, jobLog.EndTime*int64(time.Millisecond)).String()
	httpLog.PlanTime = time.Unix(0, jobLog.PlanTime*int64(time.Millisecond)).String()
	httpLog.ScheduleTime = time.Unix(0, jobLog.ScheduleTime*int64(time.Millisecond)).String()

	return
}
