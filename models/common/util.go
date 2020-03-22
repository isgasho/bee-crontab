package common

import (
	"encoding/json"
	"strings"
)

// ExtractJobName 从etcd key中提取出Job Name
func ExtractJobName(key string) string {
	return strings.TrimPrefix(key, JobSavePath)
}

// ExtractKillerName 从etcd killer 的key中提取JobName
func ExtractKillerName(key string) string {
	return strings.TrimPrefix(key, JobKillerPath)
}

// ExtarctWorkerIP 从etcd /cron/worker/ip 中获取 worker 的ip
func ExtarctWorkerIP(key string) string {
	return strings.TrimPrefix(key, JobWorkerPath)
}

// UnpackJob 反序列化得到Job
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
