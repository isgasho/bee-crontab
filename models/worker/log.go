package worker

import (
	"go.mongodb.org/mongo-driver/mongo"
	"bee-crontab/models"
	"go.mongodb.org/mongo-driver/mongo/options"
	"context"
	"time"
	"fmt"
)

// mongoDB 日志存储
type Logger struct {
	client        *mongo.Client
	logCollection *mongo.Collection
	logChan       chan *models.JobExecLog
}

var (
	Bee_Cron_Logger *Logger
)

// 初始化
func InitLogger() (err error) {
	var (
		client     *mongo.Client
		collection *mongo.Collection
	)
	client, err = mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	ctx, _ := context.WithTimeout(context.Background(), 5*time.Second)
	if err = client.Connect(ctx); err != nil {
		return
	}

	collection = client.Database("cron").Collection("log")

	// 初始化
	Bee_Cron_Logger = &Logger{
		client:        client,
		logCollection: collection,
		logChan:       make(chan *models.JobExecLog, 1000),
	}

	// 启动日志存储协程
	go Bee_Cron_Logger.writeLoop()

	return
}

// 日志存储
func (logger *Logger) writeLoop() {

	var (
		log         *models.JobExecLog //待写入的日志
		buffer      *models.LogBuffer  // 日志缓冲区
		maxSize     = 100              // 缓冲最大容量
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
				fmt.Println("buffer满了!")
				logger.saveLogs(buffer)
				buffer = nil
				commitTimer.Reset(timeOut)
			}
		case <-commitTimer.C:
			fmt.Println("定时器到期！")
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
