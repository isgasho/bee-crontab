package controllers

import (
	"encoding/json"

	"github.com/astaxie/beego"
	"github.com/sinksmell/bee-crontab/models"
)

type JobController struct {
	beego.Controller
}

//func (c *JobController) URLMapping() {
//	c.Mapping("Save", c.Save)     // update or create job
//	c.Mapping("Delete", c.Delete) // delete job
//	c.Mapping("List", c.List)     // get all of the jobs
//	c.Mapping("Kill", c.Kill)     // kill job
//	c.Mapping("Log", c.Log)
//}

// @Title SaveJob
// @Description create jobs or update jobs
// @Param	body		body 	models.Job	true		"body for Job content"
// @Success 200 {int}
// @Failure 403 body is empty
// @router /save [post]
func (c *JobController) Save() {
	var (
		job  models.Job
		resp models.Response
	)
	json.Unmarshal(c.Ctx.Input.RequestBody, &job)
	if oldJob, err := models.MJobManager.SaveJob(&job); err != nil {
		resp = models.NewResponse(-1, err.Error(), nil)
	} else {
		resp = models.NewResponse(0, "success", oldJob)
	}

	c.Data["json"] = &resp
	c.ServeJSON()
}

// @Title DeleteJob
// @Description delete job
// @Param	body		body 	models.Job	true		"body for Job content"
// @Success 200 {int}
// @Failure 403 body is empty
// @router /delete [post]
func (c *JobController) Delete() {
	var (
		job  models.Job
		resp models.Response
	)
	json.Unmarshal(c.Ctx.Input.RequestBody, &job)
	if oldJob, err := models.MJobManager.DeleteJob(&job); err != nil {
		resp = models.NewResponse(-1, err.Error(), nil)
	} else {
		resp = models.NewResponse(0, "success", oldJob)
	}
	c.Data["json"] = &resp
	c.ServeJSON()
}

// @Title ListJob
// @Description get all of the jobs
// @Success 200 {object} models.Job
// @router /list [get]
func (c *JobController) List() {
	var (
		resp models.Response
	)
	if jobs, err := models.MJobManager.ListJobs(); err != nil {
		resp = models.NewResponse(-1, err.Error(), nil)
	} else {
		resp = models.NewResponse(0, "success", jobs)
	}
	c.Data["json"] = &resp
	c.ServeJSON()
}

// @Title KillJob
// @Description Kill job by  name of job
// @Param	body		body 	models.Job	true		"body for Job content"
// @Success 200 {int}
// @Failure 403 body is empty
// @router /kill [post]
func (c *JobController) Kill() {
	var (
		resp models.Response
		job  models.Job
	)
	json.Unmarshal(c.Ctx.Input.RequestBody, &job)
	if err := models.MJobManager.KillJob(&job); err != nil {
		resp = models.NewResponse(-1, err.Error(), nil)
	} else {
		resp = models.NewResponse(0, "success", nil)
	}

	c.Data["json"] = &resp
	c.ServeJSON()
}

// @Title GetJobLog
// @Description get job execute log by job name
// @Param	name		path 	string	true		"The key for staticblock"
// @Success 200
// @router /log/:name [get]
func (c *JobController) Log() {
	var (
		logs []*models.HTTPJobLog
		resp models.Response
		err  error
	)

	jobName := c.GetString(":name")
	if jobName != "" {
		if logs, err = models.ReadLog(jobName); err != nil {
			resp = models.NewResponse(-1, err.Error(), nil)
		} else {
			resp = models.NewResponse(0, "success", logs)
		}
	}
	c.Data["json"] = &resp
	c.ServeJSON()
}
