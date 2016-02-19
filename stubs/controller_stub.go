package stubs


var controllerTemplate = `package controllers

import (
	"{{pkgPath}}/models"
	"{{pkgPath}}/helpers"

	"strconv"
	"encoding/json"

	"github.com/astaxie/beego"
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

// @Title Post
// @Description create {{ctrlName}}
// @Param	body		body 	models.{{ctrlName}}	true		"body for {{ctrlName}} content"
// @Success 200 {int} models.{{ctrlName}}.Id
// @Failure 403 body is empty
// @router / [post]
func (c *{{ctrlName}}Controller) Post() {
	var v models.{{ctrlName}}
	json.Unmarshal(c.Ctx.Input.RequestBody, &v)

	if res, errData := helpers.Validator(&v); res == false {
		c.Data["json"] = errData
	} else {
		if id, err := models.Add{{ctrlName}}(&v); err == nil {
			helpers.Rf.Success(c.Ctx.Request.Method, int(id))
			c.Data["json"] = helpers.Rf.Data
		} else {
			helpers.Rf.Fail(err.Error())
			c.Data["json"] = helpers.Rf.Data
		}
	}

	c.ServeJson()
}

// @Title Get
// @Description get {{ctrlName}} by id
// @Param	id		path 	string	true		"The key for staticblock"
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is empty
// @router /:id [get]
func (c *{{ctrlName}}Controller) GetOne() {
	idStr := c.Ctx.Input.Params[":id"]
	id, _ := strconv.Atoi(idStr)
	v, err := models.Get{{ctrlName}}ById(id)
	if err != nil {
		c.Data["json"] = nil
	} else {
		c.Data["json"] = v
	}
	c.ServeJson()
}

// @Title Get All
// @Description get {{ctrlName}}
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	join	query	string	false	"Foreign key joined. e.g. col1,col2 ..."
// @Param	groupby	query	string	false	"Group-by fields. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 
// @router / [get]
func (c *{{ctrlName}}Controller) GetAll() {
	l, err, totals := models.GetAll{{ctrlName}}(helpers.QueryString(c.Input()))

	helpers.Rf.Data = make(map[string]interface{})
	helpers.Rf.Data["totals"] = totals

	if err != nil {
		c.Data["json"] = nil
	} else {
		if l == nil {
			c.Data["json"] = nil
		} else {
			helpers.Rf.Success(c.Ctx.Request.Method, 0, l)
			c.Data["json"] = helpers.Rf.Data
		}
	}

	c.ServeJson()
}

// @Title Update
// @Description update the {{ctrlName}}
// @Param	id		path 	string				true		"The id you want to update"
// @Param	body	body 	models.{{ctrlName}}	true		"body for {{ctrlName}} content"
// @Success 200 {object} models.{{ctrlName}}
// @Failure 403 :id is not int
// @router /:id [put]
func (c *{{ctrlName}}Controller) Put() {
	idStr := c.Ctx.Input.Params[":id"]
	id, _ := strconv.Atoi(idStr)
	v := models.{{ctrlName}}{Id: id}
	
	json.Unmarshal(c.Ctx.Input.RequestBody, &v)
	keys := helpers.GetInputKeys(c.Ctx.Input.RequestBody)

	if res, errData := helpers.Validator(&v); res == false {
		c.Data["json"] = errData
	} else {
		if err := models.Update{{ctrlName}}ById(&v, keys); err == nil {
			helpers.Rf.Success(c.Ctx.Request.Method, int(id))
			c.Data["json"] = helpers.Rf.Data
		} else {
			helpers.Rf.Fail(err.Error())
			c.Data["json"] = helpers.Rf.Data
		}
	}
	c.ServeJson()
}

// @Title Delete
// @Description delete the {{ctrlName}}
// @Param	id		path 	string	true		"The id you want to delete"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *{{ctrlName}}Controller) Delete() {
	idStr := c.Ctx.Input.Params[":id"]
	id, _ := strconv.Atoi(idStr)
	if err := models.Delete{{ctrlName}}(id); err == nil {
		c.Data["json"] = "OK"
	} else {
		c.Data["json"] = nil
	}
	c.ServeJson()
}`

func TemplateController() string {
	return controllerTemplate
}