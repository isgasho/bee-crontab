package worker

import (
	"fmt"
	"github.com/segmentio/objconv/json"
	"io/ioutil"
)

// WorkerConfig  Worker节点的配置结构
type WorkerConfig struct {
	EtcdEndponits   []string `json:"etcdEndponits"`
	EtcdDialTimeout int      `json:"etcdDialTimeout"`
	MongoURL	string	`json:"mongo_url"`
}

var (
	// WorkerConf Worker的全局配置单例
	WorkerConf *WorkerConfig
)

// InitConfig 解析Worker配置文件
func InitConfig(filename string) (err error) {
	var (
		content []byte
		config  WorkerConfig
	)

	// 读取文件
	if content, err = ioutil.ReadFile(filename); err != nil {
		return
	}
	// 解析json
	if err = json.Unmarshal(content, &config); err != nil {
		return
	}
	WorkerConf = &config
	fmt.Printf("%+v\n", WorkerConf)
	return
}
