package worker

import (
	"io/ioutil"
	"github.com/segmentio/objconv/json"
	"fmt"
)

type  WorkerConfig  struct{
	EtcdEndponits []string `json:"etcdEndponits"`
	EtcdDialTimeout int `json:"etcdDialTimeout"`
}

var(
	WorkerConf *WorkerConfig
)

// 解析配置文件
func InitConfig(filename string)(err error){
	var(
		content []byte
		config WorkerConfig
	)

	// 读取文件
	if content,err=ioutil.ReadFile(filename);err!=nil{
		return
	}
	// 解析json
	if err=json.Unmarshal(content,&config);err!=nil{
		return
	}
	WorkerConf=&config
	fmt.Printf("%+v\n",WorkerConf)
	return
}