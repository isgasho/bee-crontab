package models

import (
	"go.etcd.io/etcd/clientv3"
	"time"
	"github.com/astaxie/beego"
	"encoding/json"
	"context"
	"github.com/sinksmell/bee-crontab/models/common"
)

// 单例全局变量

var (
	MJobManager *MasterJobMgr
)

// 任务管理器
type MasterJobMgr struct {
	client *clientv3.Client
	kv     clientv3.KV
	lease  clientv3.Lease
}

func init() {
	var (
		config clientv3.Config
		client *clientv3.Client
		kv     clientv3.KV
		lease  clientv3.Lease
		err    error
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

	// 组装单例
	MJobManager = &MasterJobMgr{
		client: client,
		kv:     kv,
		lease:  lease,
	}

}

// 添加或者修改一个任务
func (jobMgr *MasterJobMgr) SaveJob(job *Job) (oldJob *Job, err error) {

	var (
		jobKey    string
		bytes     []byte
		putResp   *clientv3.PutResponse
		oldJobObj Job
	)
	// 得到job保存路径
	jobKey = common.JOB_SAVE_PATH + job.Name

	if bytes, err = json.Marshal(job); err != nil {
		return
	}
	// etcd put 操作
	if putResp, err = MJobManager.kv.Put(context.TODO(), jobKey, string(bytes), clientv3.WithPrevKV()); err != nil {
		return
	}
	// 如果prevKV 不为空则返回旧值
	if putResp.PrevKv != nil {
		if err = json.Unmarshal(putResp.PrevKv.Value, &oldJobObj); err != nil {
			// 为了提高容错性
			// 旧值是否正确解析 不影响最终结果
			err = nil
			return
		}
	}
	// 赋值旧值
	oldJob = &oldJobObj
	return
}

// 删除一个任务
func (jobMgr *MasterJobMgr) DeleteJob(job *Job) (oldJob *Job, err error) {

	var (
		jobKey    string
		oldJobObj Job
		delResp   *clientv3.DeleteResponse
	)

	jobKey = common.JOB_SAVE_PATH + job.Name
	if delResp, err = MJobManager.kv.Delete(context.TODO(), jobKey, clientv3.WithPrevKV()); err != nil {
		return
	}
	// 解析原来的旧值
	if len(delResp.PrevKvs) != 0 {
		if err = json.Unmarshal(delResp.PrevKvs[0].Value, &oldJobObj); err != nil {
			// 是否成功解析出来对操作结果没有影响
			err = nil
			return
		}
	}
	oldJob = &oldJobObj
	return
}

//获取所有的任务
func (jobMgr *MasterJobMgr) ListJobs() (jobs []*Job, err error) {
	var (
		jobKey  string
		job     *Job
		getResp *clientv3.GetResponse
	)

	jobKey = common.JOB_SAVE_PATH
	jobs = make([]*Job, 0)
	if getResp, err = jobMgr.kv.Get(context.TODO(), jobKey, clientv3.WithPrefix()); err != nil {
		return
	}

	if len(getResp.Kvs) != 0 {
		for _, kvPair := range getResp.Kvs {
			job = &Job{}
			if err = json.Unmarshal(kvPair.Value, job); err != nil {
				// 容忍了个别任务反序列化失败
				// 正常情况下是可以反序列化的
				err = nil
				continue
			}
			jobs = append(jobs, job)
		}
	}
	return
}

// 杀死一个任务
// 向 /cron/killer/JobName put 一个值
// worker监听变化,强行终止对应的任务
func (jobMgr *MasterJobMgr) KillJob(job *Job) (err error) {

	var (
		killJobKey  = common.JOB_KILLER_PATH + job.Name
		leaseId    clientv3.LeaseID
		grantResp  *clientv3.LeaseGrantResponse
	)

	// 申请一个租约 设置对应的过期时间
	if grantResp, err = jobMgr.lease.Grant(context.TODO(), 2); err != nil {
		return
	}

	leaseId = grantResp.ID
	// 向 /cron/killer/JobName put "kill" 表示杀死对应的任务
	// 租约到期自动删除对应的 k-v
	if _, err = jobMgr.kv.Put(context.TODO(), killJobKey, "kill", clientv3.WithLease(leaseId)); err != nil {
		return
	}

	return
}


