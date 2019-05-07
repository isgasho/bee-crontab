package common

const (
	// 任务事件
	// 任务保存事件
	JOB_EVENT_SAVE = iota
	// 任务删除事件
	JOB_EVENT_DELETE
	// 杀死任务事件
	JOB_EVENT_KILL
	
	
	// 任务保存目录
	JOB_SAVE_PATH = "/cron/jobs/"
	// job killer 目录
	JOB_KILLER_PATH = "/cron/killer/"
	// 分布式锁路径
	JOB_LOCK_PATH = "/cron/lock/"
	// worker节点注册路径  服务注册
	JOB_WORKER_PATH="/cron/worker/"
	
)
