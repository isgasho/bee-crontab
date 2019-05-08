package worker

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"go.etcd.io/etcd/mvcc/mvccpb"
	"github.com/sinksmell/bee-crontab/models"
	"context"
	"github.com/sinksmell/bee-crontab/models/common"
	"fmt"
)

// worker 任务管理器
type WorkerJobMgr struct {
	client  *clientv3.Client //连接etcd 客户端
	kv      clientv3.KV      //kv
	lease   clientv3.Lease   //租约
	watcher clientv3.Watcher //监听器
}

var (
	WorkerJobManager *WorkerJobMgr
)

func InitJobMgr() (err error) {
	var (
		config  clientv3.Config
		client  *clientv3.Client
		kv      clientv3.KV
		lease   clientv3.Lease
		watcher clientv3.Watcher
	)

	// 初始化配置
	config = clientv3.Config{
		Endpoints:   WorkerConf.EtcdEndponits,
		DialTimeout: time.Duration(WorkerConf.EtcdDialTimeout) * time.Millisecond,
	}

	// 建立连接
	if client, err = clientv3.New(config); err != nil {
		return
	}

	// 创建kv lease watcher
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)
	watcher = clientv3.NewWatcher(client)

	// 初始化单例
	WorkerJobManager = &WorkerJobMgr{
		client:  client,
		kv:      kv,
		lease:   lease,
		watcher: watcher,
	}

	//	fmt.Printf("%+v\n",WorkerJobManager)

	// 启动监听任务
	if err = WorkerJobManager.WatchJobs(); err != nil {
		return
	}

	//fmt.Println("启动任务监听器！")
	// 启动监听killer
	if err=WorkerJobManager.WatchKillers();err!=nil{
		return
	}

	return
}

// 从etcd中读取任务
// 监听kv变化
func (jobMgr *WorkerJobMgr) WatchJobs() (err error) {

	var (
		getResp           *clientv3.GetResponse
		kvPair            *mvccpb.KeyValue
		job               *models.Job
		jobName           string
		watchStartRevison int64
		watchChan         clientv3.WatchChan
		watchResp         clientv3.WatchResponse
		watchEvent        *clientv3.Event
		jobEvent          *models.JobEvent
	)

	// 1.get /cron/jobs/下所有任务 并获取 revision
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_SAVE_PATH, clientv3.WithPrefix()); err != nil {
		return
	}

	// 遍历kv
	for _, kvPair = range getResp.Kvs {
		// 反序列化 value->job
		// 如果某次失败则跳过 提高容错性
		if job, err = models.UnpackJob(kvPair.Value); err == nil {
			// 说明是有效的任务
			// 发送给调度器
			//TODO:构造事件 发送给调度器
			jobEvent = models.NewJobEvent(common.JOB_EVENT_SAVE, job)
			Bee_Scheduler.PushJobEvent(jobEvent)
			//fmt.Println("构造任务事件!")
			fmt.Println(jobEvent)
		}
	}

	//2.从当前revision之后监听变化
	go func() {
		//监听协程
		watchStartRevison = getResp.Header.Revision + 1
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_SAVE_PATH, clientv3.WithPrefix())
		for watchResp = range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					// 任务保存事件
					if job, err = models.UnpackJob(watchEvent.Kv.Value); err != nil {
						// 任务解析失败 跳过
						continue
					}
					// 构造一个更新事件
					jobEvent = models.NewJobEvent(common.JOB_EVENT_SAVE, job)
					fmt.Println(jobEvent)
					// TODO 传给调度器
					Bee_Scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
					// 任务删除事件
					jobName = models.ExtractJobName(string(watchEvent.Kv.Key))
					job = &models.Job{Name: jobName}
					// 构造一个任务删除事件
					jobEvent = models.NewJobEvent(common.JOB_EVENT_DELETE, job)

					//TODO： 推送给调度器
					Bee_Scheduler.PushJobEvent(jobEvent)
				}
			}
		}

	}()

	return
}

// 从etcd读取killer
// 监听kv变化
func (jobMgr *WorkerJobMgr) WatchKillers() (err error) {

	// 监听 /cron/killer/ 目录的变化
	var (
		getResp           *clientv3.GetResponse
		watchChan  clientv3.WatchChan
		watchResp  clientv3.WatchResponse
		watchEvent *clientv3.Event
		jobEvent   *models.JobEvent
		jobName    string
		job        *models.Job
		watchStartRevison int64
	)

	// 1.get /cron/jobs/ 下的所有任务,并获取当前revision
	if getResp, err = jobMgr.kv.Get(context.TODO(), common.JOB_KILLER_PATH, clientv3.WithPrefix()); err != nil {
		return
	}
	//从最新的revision之后监听变化
	go func() {
		watchStartRevison = getResp.Header.Revision
		watchChan = jobMgr.watcher.Watch(context.TODO(), common.JOB_KILLER_PATH, clientv3.WithPrefix(),clientv3.WithRev(watchStartRevison))
		for watchResp=range watchChan {
			for _, watchEvent = range watchResp.Events {
				switch watchEvent.Type {
				case mvccpb.PUT:
					// 杀死某个任务
					// 从key中提取出任务名
					jobName =models.ExtractKillerName(string(watchEvent.Kv.Key))
					job=&models.Job{Name: jobName}
					jobEvent=models.NewJobEvent(common.JOB_EVENT_KILL,job)
					// 事件推送给 schedular
					Bee_Scheduler.PushJobEvent(jobEvent)
				case mvccpb.DELETE:
					// killer 任务过期
				}
				// 变化推送给 scheduler
				// scheduler得知后调用cancelFunc取消对应的任务执行
			}
		}
	}()

	return
}




// 创建分布式锁
func (jobMgr *WorkerJobMgr) NewLock(jobName string) (lock *JobLock) {
	return InitJobLock(jobName, jobMgr.kv, jobMgr.lease)
}
