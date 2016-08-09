package stubs

var modelTemplate = `package models

import (
	"errors"
	"reflect"
	{{timePkg}}

	"github.com/qasico/beego/orm"
	"github.com/qasico/beego/helper"
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
	offset int64, limit int64, join []string) (result []interface{}, total int64, err error) {

	o := orm.NewOrm()
	qs := o.QueryTable(new({{modelName}})).SetCond(helper.QueryCondition(query))

	if helper.IsJoin(join) {
		qs = o.QueryTable(new({{modelName}})).SetCond(helper.QueryCondition(query)).RelatedSel(helper.QueryJoin(join))
	}

	if len(sortby) != len(order) && len(order) != 1 {
		return nil, total, errors.New("'sortby', 'order' sizes mismatch or 'order' size is not 1")
	}

	sortFields := helper.SetSorting(sortby, order)
	qs = qs.OrderBy(sortFields...).GroupBy(groupby...)

	total, err = qs.Count()
	if err != nil || total == 0 {
		return nil, total, err
	}

	var l []{{modelName}}
	if _, err := qs.Limit(limit, offset).All(&l, fields...); err == nil {
		if len(fields) == 0 {
			for _, v := range l {
				result = append(result, v)
			}
		} else {
			for _, v := range l {
				m := make(map[string]interface{})
				val := reflect.ValueOf(v)
				for _, fname := range fields {
					m[fname] = val.FieldByName(helper.CamelString(fname)).Interface()
				}
				result = append(result, m)
			}
		}

		return result, total, nil
	}

	return nil, total, err
}

func Update{{modelName}}ById(m *{{modelName}}, keys []string) (err error) {
	_, err = orm.NewOrm().Update(m, keys...)
	return
}

func Delete{{modelName}}(m *{{modelName}}) (err error) {
	if num, _ := orm.NewOrm().Delete(m); num == 0 {
		return errors.New("data not exists")
	}

	return
}`

var modelNoPK = `package models
import (
	{{timePkg}}

	"github.com/qasico/beego/orm"
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