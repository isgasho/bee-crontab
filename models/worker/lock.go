package worker

import (
	"go.etcd.io/etcd/clientv3"
	"context"
	"bee-crontab/models/common"
)

// 分布式锁
// 基于txn事务
type JobLock struct {
	// etcd
	kv         clientv3.KV    //kv
	lease      clientv3.Lease // 租约
	leaseId    clientv3.LeaseID
	jobName    string             // 任务名
	cancelFunc context.CancelFunc //用于取消租约 即释放锁
	isLocked   bool               // 标记上锁是否成功
}

// 初始化
func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (*JobLock) {
	return &JobLock{kv: kv, jobName: jobName, lease: lease}
}

// 尝试上锁 抢乐观锁
func (lock *JobLock) TryLock() (err error) {
	var (
		leaseResp    *clientv3.LeaseGrantResponse
		leaseId      clientv3.LeaseID
		ctx          context.Context
		cancelFunc   context.CancelFunc
		keepRespChan <-chan *clientv3.LeaseKeepAliveResponse
		txn          clientv3.Txn
		txnResp      *clientv3.TxnResponse
		lockKey      string
	)

	// 1.创建租约 续租时间5秒
	if leaseResp, err = lock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	// 2.创建用于取消租约的上下文
	ctx, cancelFunc = context.WithCancel(context.TODO())
	// 3.开启自动续租
	leaseId = leaseResp.ID
	if keepRespChan, err = lock.lease.KeepAlive(ctx, leaseId); err != nil {
		goto FAIL
	}

	// 4.续租成功 开启自动续租应答协程

	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)
		// TODO : 续租应答协程
		// 当通道被关闭时 程序协程自动退出
		for keepResp = range keepRespChan {
			if keepResp==nil{
				return
			}
		}

	}()

	// 5.创建事务
	txn = lock.kv.Txn(context.TODO())
	lockKey = common.JOB_LOCK_PATH + lock.jobName

	// 6.事务抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "locked", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	// 提交事务
	if txnResp, err = txn.Commit(); err != nil {
		// 失败则立即释放租约
		goto FAIL
	}

	//7.事务提交成功 判断结果
	if !txnResp.Succeeded {
		// 锁被占用
		err = common.ERR_LOCK_BUSY
		goto FAIL
	}

	// 抢锁成功
	lock.cancelFunc = cancelFunc
	lock.leaseId = leaseId
	lock.isLocked = true

	return

FAIL:
// 失败释放锁
	cancelFunc()                               //取消续租
	lock.lease.Revoke(context.TODO(), leaseId) // 释放租约
	return
}

// 释放锁
func (lock *JobLock) UnLock() {
	if lock.isLocked {
		// 1.取消自动续租
		lock.cancelFunc()
		// 2.释放租约
		lock.lease.Revoke(context.TODO(), lock.leaseId)
		// 3.重置标记位
		lock.isLocked = false
	}
	return
}
