package models


// 根据任务名查询日志的过滤器
type  JobFilter  struct{
    Name string `bson:"jobName"`
}
