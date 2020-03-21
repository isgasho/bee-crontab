package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/sinksmell/bee-crontab/models"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Logger mongoDB 日志存储器
type Logger struct {
	client        *mongo.Client
	logCollection *mongo.Collection
	logChan       chan *models.JobExecLog
}

var (
	// BeeCronLogger 全局日志存储器单例
	BeeCronLogger *Logger
)

// InitLogger 初始化Logger的单例
func InitLogger() (err error) {
	var (
		client     *mongo.Client
		collection *mongo.Collection
	)
	client, err = mongo.NewClient(options.Client().ApplyURI(WorkerConf.MongoURL))
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err = client.Connect(ctx); err != nil {
		return
	}
	collection = client.Database("cron").Collection("log")

	// 初始化
	BeeCronLogger = &Logger{
		client:        client,
		logCollection: collection,
		logChan:       make(chan *models.JobExecLog, 1024),
	}

	// 启动日志存储协程
	go BeeCronLogger.writeLoop()

	return
}

// 日志存储
func (logger *Logger) writeLoop() {

	var (
		log         *models.JobExecLog //待写入的日志
		buffer      *models.LogBuffer  // 日志缓冲区
		maxSize     = 128              // 缓冲最大容量
		commitTimer *time.Timer        // 提交定时器
		timeOut     = 10 * time.Second //超时时间
	)

	for {
		// 使用buffer 和定时器机制
		// 实现定时批量提交
		// 提高吞吐率
		// 减少I/O次数
		if commitTimer == nil {
			commitTimer = time.NewTimer(timeOut)
		}
		if buffer == nil {
			// 初始化缓冲载体
			buffer = &models.LogBuffer{
				Logs: make([]interface{}, 0),
			}
		}

		select {
		case log = <-logger.logChan:
			// 有日志传来
			buffer.Logs = append(buffer.Logs, log)
			if len(buffer.Logs) >= maxSize {
				fmt.Println("log buffer满了!")
				logger.saveLogs(buffer)
				buffer = nil
				commitTimer.Reset(timeOut)
			}
		case <-commitTimer.C:
			fmt.Println("log存储定时器到期！")
			// 定时器到期
			if buffer != nil {
				// 保存日志
				logger.saveLogs(buffer)
				buffer = nil
			}
			commitTimer.Reset(timeOut)
		}
	}

}

// 批量保存日志
func (logger *Logger) saveLogs(buffer *models.LogBuffer) {
	logger.logCollection.InsertMany(context.TODO(), buffer.Logs)
	fmt.Println("写入日志成功!")
}
