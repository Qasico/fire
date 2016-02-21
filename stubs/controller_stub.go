package stubs


var controllerTemplate = `package controllers

import (
	"strconv"
	"encoding/json"
	"{{pkgPath}}/models"

	"github.com/qasico/beego"
	"github.com/qasico/beego/helper"
)

type {{ctrlName}}Controller struct {
	beego.Controller
}

func (c *{{ctrlName}}Controller) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
}

// @Title Create new data
// @Success 200 {int} models.{{ctrlName}}
// @Failure 403 body is empty
// @router / [post]
func (c *{{ctrlName}}Controller) Post() {
	var v models.{{ctrlName}}
	var response helper.APIResponse

	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		if valid := response.Validator(&v); valid != false {
			if _, err := models.Add{{ctrlName}}(&v); err == nil {
				response.Success(1, v)
			} else {
				response.Failed(400, err.Error())
			}
		}
	} else {
		response.Failed(400, err.Error())
	}

	c.Ctx.Output.SetStatus(response.Code)
	c.Data["json"] = response.GetResponse("POST")
	c.ServeJSON()
}

// @Title Get single data with provided id
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is empty
// @router /:id [get]
func (c *{{ctrlName}}Controller) GetOne() {
	response := helper.APIResponse{}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	if data, err := models.Get{{ctrlName}}ById(id); err == nil {
		response.Success(1, data)
	} else {
		response.Failed(404, err.Error())
	}

	c.Ctx.Output.SetStatus(response.Code)
	c.Data["json"] = response.GetResponse("GET")
	c.ServeJSON()
}

// @Title Get data with parameters query string
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 
// @router / [get]
func (c *{{ctrlName}}Controller) GetAll() {
	response := helper.APIResponse{}

	if data, total, err := models.GetAll{{ctrlName}}(helper.QueryString(c.Input())); err == nil {
		response.Success(total, data)
	} else {
		response.Failed(400, err.Error())
	}

	c.Ctx.Output.SetStatus(response.Code)
	c.Data["json"] = response.GetResponse("GET")
	c.ServeJSON()
}

// @Title Update model with provided key and new values
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is not int
// @router /:id [put]
func (c *{{ctrlName}}Controller) Put() {
	response := helper.APIResponse{}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v := models.{{ctrlName}}{Id: id}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		keys := helper.GetInputKeys(c.Ctx.Input.RequestBody)
		if valid := response.Validator(&v); valid != false {
			if err := models.Update{{ctrlName}}ById(&v, keys); err == nil {
				response.Success(0, nil)
			} else {
				response.Failed(404, err.Error())
			}
		}
	} else {
		response.Failed(400, err.Error())
	}

	c.Ctx.Output.SetStatus(response.Code)
	c.Data["json"] = response.GetResponse("POST")
	c.ServeJSON()
}

// @Title Delete model with provided id
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *{{ctrlName}}Controller) Delete() {
	response := helper.APIResponse{}

	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v := models.{{ctrlName}}{Id: id}
	if err := models.Delete{{ctrlName}}(&v); err == nil {
		response.Success(0, nil)
	} else {
		response.Failed(404, err.Error())
	}

	c.Ctx.Output.SetStatus(response.Code)
	c.Data["json"] = response.GetResponse("POST")
	c.ServeJSON()
}`

func TemplateController() string {
	return controllerTemplate
}