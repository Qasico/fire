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

	json.Unmarshal(c.Ctx.Input.RequestBody, &v)

	if valid := helper.Respond.Validator(&v); valid != false {
		if _, err := models.Add{{ctrlName}}(&v); err == nil {
			helper.Respond.SuccessWithModel("POST", v)
		} else {
			helper.Respond.Fail(err.Error())
		}

	}

	if !helper.Respond.IsSuccess() {
		c.Ctx.Output.SetStatus(404)
	}

	c.Data["json"] = helper.Respond.Format
	c.ServeJSON()
}

// @Title Get single data with provided id
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is empty
// @router /:id [get]
func (c *{{ctrlName}}Controller) GetOne() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)

	v, err := models.Get{{ctrlName}}ById(id)
	if err != nil {
		helper.Respond.Fail(err.Error())
	} else {
		helper.Respond.SuccessWithModel("GET", v)
	}

	if !helper.Respond.IsSuccess() {
		c.Ctx.Output.SetStatus(404)
	}

	c.Data["json"] = helper.Respond.Format
	c.ServeJSON()
}

// @Title Get data with parameters query string
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 
// @router / [get]
func (c *{{ctrlName}}Controller) GetAll() {
	l, total, err := models.GetAll{{ctrlName}}(helper.QueryString(c.Input()))

	if err != nil {
		helper.Respond.Fail(err.Error())
	} else {
		helper.Respond.SuccessWithData(total, l)
	}

	if !helper.Respond.IsSuccess() {
		c.Ctx.Output.SetStatus(404)
	}

	c.Data["json"] = helper.Respond.Format
	c.ServeJSON()
}

// @Title Update model with provided key and new values
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is not int
// @router /:id [put]
func (c *{{ctrlName}}Controller) Put() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v := models.{{ctrlName}}{Id: id}
	
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		keys := helper.GetInputKeys(c.Ctx.Input.RequestBody)
		if valid := helper.Respond.Validator(&v); valid != false {
			if err := models.Update{{ctrlName}}ById(&v, keys); err == nil {
				helper.Respond.SuccessWithModel("GETONE", v)
			} else {
				helper.Respond.Fail(err.Error())
			}
		}
	}  else {
		helper.Respond.Fail(err.Error())
	}

	if !helper.Respond.IsSuccess() {
		c.Ctx.Output.SetStatus(404)
	}

	c.Data["json"] = helper.Respond.Format
	c.ServeJSON()
}

// @Title Delete model with provided id
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *{{ctrlName}}Controller) Delete() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)

	if err := models.Delete{{ctrlName}}(id); err == nil {
		helper.Respond.SuccessWithModel("POST", v)
	} else {
		helper.Respond.Fail(err.Error())
	}

	if !helper.Respond.IsSuccess() {
		c.Ctx.Output.SetStatus(404)
	}

	c.Data["json"] = helper.Respond.Format
	c.ServeJSON()
}`

func TemplateController() string {
	return controllerTemplate
}