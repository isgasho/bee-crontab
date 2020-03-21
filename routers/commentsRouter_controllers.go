package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

func init() {

	beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"] = append(beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"],
		beego.ControllerComments{
			Method:           "Delete",
			Router:           `/delete`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"] = append(beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"],
		beego.ControllerComments{
			Method:           "Kill",
			Router:           `/kill`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"] = append(beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"] = append(beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"],
		beego.ControllerComments{
			Method:           "Log",
			Router:           `/log/:name`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"] = append(beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:JobController"],
		beego.ControllerComments{
			Method:           "Save",
			Router:           `/save`,
			AllowHTTPMethods: []string{"post"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:WorkerController"] = append(beego.GlobalControllerRouter["github.com/sinksmell/bee-crontab/controllers:WorkerController"],
		beego.ControllerComments{
			Method:           "List",
			Router:           `/list`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
