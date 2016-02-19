package stubs

var modelTemplate = `package models

import (
	"{{pkgPath}}/helpers"

	"fmt"
	"errors"
	"reflect"
	{{timePkg}}
	"github.com/astaxie/beego/orm"
)

{{modelStruct}}

func (t *{{modelName}}) TableName() string {
	return "{{tableName}}"
}

func init() {
	orm.RegisterModel(new({{modelName}}))
}

func Add{{modelName}}(m *{{modelName}}) (id int64, err error) {
	o := orm.NewOrm()
	id, err = o.Insert(m)
	return
}

func Get{{modelName}}ById(id int) (v *{{modelName}}, err error) {
	var m {{modelName}}
	o := orm.NewOrm()

	if err = o.QueryTable(new({{modelName}})).Filter("id", id).RelatedSel().One(&m); err == nil {
		return &m, nil
	}

	return nil, err
}

func GetAll{{modelName}}(query map[int]map[string]string, fields []string, groupby []string, sortby []string, order []string,
	offset int64, limit int64, join []string) (ml []interface{}, err error, totals int64) {

	o := orm.NewOrm()
	qs := o.QueryTable(new({{modelName}})).SetCond(helpers.QueryCondition(query))

	if helpers.IsJoin(join) {
		qs = o.QueryTable(new({{modelName}})).SetCond(helpers.QueryCondition(query)).RelatedSel(helpers.QueryJoin(join))
	}

	cnt, err := qs.Count()
	if err != nil {
		return nil, err, cnt
	}

	var sortFields []string
	if len(sortby) != 0 {
		if len(sortby) == len(order) {
			for i, v := range sortby {
				orderby := ""
				if order[i] == "desc" {
					orderby = "-" + v
				} else if order[i] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]"), cnt
				}
				sortFields = append(sortFields, orderby)
			}
			qs = qs.OrderBy(sortFields...)
		} else if len(sortby) != len(order) && len(order) == 1 {
			for _, v := range sortby {
				orderby := ""
				if order[0] == "desc" {
					orderby = "-" + v
				} else if order[0] == "asc" {
					orderby = v
				} else {
					return nil, errors.New("Error: Invalid order. Must be either [asc|desc]"), cnt
				}
				sortFields = append(sortFields, orderby)
			}
		} else if len(sortby) != len(order) && len(order) != 1 {
			return nil, errors.New("Error: 'sortby', 'order' sizes mismatch or 'order' size is not 1"), cnt
		}
	} else {
		if len(order) != 0 {
			return nil, errors.New("Error: unused 'order' fields"), cnt
		}
	}

	var l []{{modelName}}
	qs = qs.OrderBy(sortFields...).GroupBy(groupby...)
	if _, err := qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				ml = append(ml, v)
			}
		} else {
			// trim unused fields
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(helpers.CamelString(fname)).Interface()
				}
				ml = append(ml, m)
			}
		}

		return ml, nil, cnt
	}
	return nil, err, cnt
}

func Update{{modelName}}ById(m *{{modelName}}, keys []string) (err error) {
	o := orm.NewOrm()
	v := {{modelName}}{Id: m.Id}

	if err = o.Read(&v); err == nil {
		o.Update(m, keys...)
	}
	return
}

func Delete{{modelName}}(id int) (err error) {
	o := orm.NewOrm()
	v := {{modelName}}{Id: id}

	if err = o.Read(&v); err == nil {
		var num int64
		if num, err = o.Delete(&{{modelName}}{Id: id}); err == nil {
			fmt.Println("Number of records deleted in database:", num)
		}
	}
	return
}`

var modelNoPK = `package models
import (
	{{timePkg}}

	"github.com/astaxie/beego/orm"
)

{{modelStruct}}

func GetAll{{modelName}}() (ml []interface{}, err error, totals int64) {

	qb, _ := orm.NewQueryBuilder("mysql")

	qb.Select("*")
	qb.From("{{tableName}}")

	o := orm.NewOrm()
	sql := qb.String()

	var m []{{modelName}}
	if _, err := o.Raw(sql).QueryRows(&m); err == nil {
		for _, v := range m {
			ml = append(ml, v)
		}
	}


	return ml, err, totals
}`

func TemplateModel(pk bool) string {
	if(pk){
		return modelTemplate
	} else {
		return modelNoPK
	}
}