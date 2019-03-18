package models

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"github.com/astaxie/beego"
	"context"
	"bee-crontab/models/common"
	"go.etcd.io/etcd/mvcc/mvccpb"
)

// worker管理 用来发现worker
// /cron/worker/
type WorkerMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

// 为了使获取节点列表 得到的信息更加丰富 而不是单纯的ip
// 从而添加的描述节点状态的结构体
type WorkerInfo struct {
	Time string `json:"time"` // 查询时间
	Ip   string `json:"ip"`   // 节点ip
}

var (
	WorkerManager *WorkerMgr
)

func init() {
	InitWorkerMgr()
}

func InitWorkerMgr() (err error) {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
	)

	config = clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 5 * time.Second,
	}

	if client, err = clientv3.New(config); err != nil {
		beego.Error(err)
		return
	}
	// 得到kv 和lease
	kv = clientv3.NewKV(client)
	lease = clientv3.NewLease(client)

	WorkerManager = &WorkerMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}

	return
}

func (workerMgr *WorkerMgr) ListWorkers() (workers []*WorkerInfo, err error) {

	var (
		getResp *clientv3.GetResponse
		kvPair  *mvccpb.KeyValue
		ip      string
		info    *WorkerInfo
	)

	workers = make([]*WorkerInfo, 0)
	// 获取目录下所有的节点 ip
	if getResp, err = workerMgr.kv.Get(context.TODO(), common.JOB_WORKER_PATH, clientv3.WithPrefix()); err != nil {
		return
	}

	// 保存结果
	for _, kvPair = range getResp.Kvs {
		ip = ExtarctWorkerIP(string(kvPair.Key))
		if len(ip) != 0 {
			info = &WorkerInfo{Ip: ip,
			}
			info.Time=time.Now().Format("2006-01-02 15:04:05")
			workers = append(workers, info)
		}
	}

	return
}
