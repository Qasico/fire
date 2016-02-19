package stubs

var helper = `package helpers

import (
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
)

func GetInputKeys(input []byte) []string {
	var objmap map[string]*json.RawMessage
	json.Unmarshal(input, &objmap)

	keys := make([]string, 0, len(objmap))
	for k := range objmap {
		keys = append(keys, k)
	}
	return keys
}

func Validator(model interface{}) (bool, map[string]interface{}) {

	errorData := make(map[string]string)
	valid := validation.Validation{}

	passed, _ := valid.Valid(model)
	if !passed {
		for _, err := range valid.Errors {
			field := strings.Split(err.Key, ".")
			errorData[field[0]] = err.Message
		}
		Rf.Fail(errorData)
		return false, Rf.Data
	} else {
		return true, Rf.Data
	}
}

func QueryString(qs url.Values) (query map[int]map[string]string, fields []string, groupby []string, sortby []string, order []string,
	offset int64, limit int64, join []string) {

	var cq map[string]string = make(map[string]string)
	query = make(map[int]map[string]string)
	limit = 10
	offset = 0


	if param, ok := qs["fields"]; ok {
	    fields = strings.Split(param[0], ",")
	}

	if param, ok := qs["join"]; ok {
	    k := strings.Replace(param[0], ".", "__", -1)
		join = strings.Split(k, ",")
	}

	if param, ok := qs["groupby"]; ok {
	    groupby = strings.Split(param[0], ",")
	}

	if param, ok := qs["sortby"]; ok {
		k := strings.Replace(param[0], ".", "__", -1)
	    sortby = strings.Split(k, ",")
	}

	if param, ok := qs["order"]; ok {
	    order = strings.Split(param[0], ",")
	}

	if param, ok := qs["limit"]; ok {
	    x, _ := strconv.Atoi(param[0])
		limit = int64(x)
	}

	if param, ok := qs["offset"]; ok {
	    x, _ := strconv.Atoi(param[0])
		offset = int64(x)
	}

	if param, ok := qs["query"]; ok {
	    var index int = 0
		for _, cond := range strings.Split(param[0], "|") {

			for _, partcond := range strings.Split(cond, ",") {
				kv := strings.Split(partcond, ":")
				if len(kv) > 1 {
					k, val := kv[0], kv[1]
					cq[k] = val
				}
			}

			index = index + 1
			query[index] = cq

			cq = make(map[string]string)
		}
	}

	return query, fields, groupby, sortby, order, offset, limit, join
}

func QueryDetail(oc *orm.Condition, k string, v string, f string) (cond *orm.Condition){
	if strings.Contains(k, "__in") {
		vArr := strings.Split(v, ".")
		switch f {
		    case "or":
		    	cond = oc.Or(k, vArr)
		    case "ornot":
		    	cond = oc.OrNot(k, vArr)
		    case "andnot":
		    	cond = oc.AndNot(k, vArr)
		    default:
		    	cond = oc.And(k, vArr)
	    }
	} else if strings.Contains(k, "__between") {
		vArr := strings.Split(v, ".")
		switch f {
		    case "or":
		    	cond = oc.Or(k, vArr)
		    case "ornot":
		    	cond = oc.OrNot(k, vArr)
		    case "andnot":
		    	cond = oc.AndNot(k, vArr)
		    default:
		    	cond = oc.And(k, vArr)
	    }
	} else if strings.Contains(k, "__null") {
		k = strings.Replace(k, "__null", "__isnull", -1)
		switch f {
		    case "or":
		    	cond = oc.Or(k, true)
		    case "ornot":
		    	cond = oc.OrNot(k, true)
		    case "andnot":
		    	 cond = oc.AndNot(k, true)
		    default:
		    	cond = oc.And(k, true)
	    }
	} else if strings.Contains(k, "__notnull") {
		k = strings.Replace(k, "__notnull", "__isnull", -1)
		switch f {
		    case "or":
		    	cond = oc.Or(k, false)
		    case "ornot":
		    	cond = oc.OrNot(k, false)
		    case "andnot":
		    	cond = oc.AndNot(k, false)
		    default:
		    	cond = oc.And(k, false)
	    }
	} else {
		switch f {
		    case "or":
		    	cond = oc.Or(k, v);
		    case "ornot":
		    	cond = oc.OrNot(k, v);
		    case "andnot":
		    	cond = oc.AndNot(k, v);
		    default:
		    	cond = oc.And(k, v)
	    }
	}

	return cond
}

func QueryCondition(query map[int]map[string]string) (cond *orm.Condition) {
	cond = orm.NewCondition()

	for _, q := range query {
		condition := orm.NewCondition()
		for k, v := range q {
			if strings.Contains(k, "And.") {
				k = strings.Replace(k, "And.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = QueryDetail(condition, k, v, "and")
			} else if strings.Contains(k, "Ex.") {
				k = strings.Replace(k, "Ex.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = QueryDetail(condition, k, v, "andnot")
			} else if strings.Contains(k, "Or.") {
				k = strings.Replace(k, "Or.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = QueryDetail(condition, k, v, "or")
			} else if strings.Contains(k, "OrNot.") {
				k = strings.Replace(k, "OrNot.", "", -1)
				k = strings.Replace(k, ".", "__", -1)

				condition = QueryDetail(condition, k, v, "ornot")
			} else {
				k = strings.Replace(k, ".", "__", -1)

				condition = QueryDetail(condition, k, v, "and")
			}
		}

		cond = cond.AndCond(condition)
	}

	return cond
}

func QueryJoin(joins []string) (field interface{}) {
	if len(joins) > 0 {
		return joins
	}

	return nil;
}

func IsJoin(joins []string) bool {
	if (len(joins) > 0) && (joins[0] == "none") {
		return false
	}

	return true
}

func SnakeString(s string) string {
	data := make([]byte, 0, len(s)*2)
	j := false
	num := len(s)
	for i := 0; i < num; i++ {
		d := s[i]
		if i > 0 && d >= 'A' && d <= 'Z' && j {
			data = append(data, '_')
		}
		if d != '_' {
			j = true
		}
		data = append(data, d)
	}
	return strings.ToLower(string(data[:len(data)]))
}

func CamelString(s string) string {
	data := make([]byte, 0, len(s))
	j := false
	k := false
	num := len(s) - 1
	for i := 0; i <= num; i++ {
		d := s[i]
		if k == false && d >= 'A' && d <= 'Z' {
			k = true
		}
		if d >= 'a' && d <= 'z' && (j || k == false) {
			d = d - 32
			j = false
			k = true
		}
		if k && d == '_' && num > i && s[i+1] >= 'a' && s[i+1] <= 'z' {
			j = true
			continue
		}
		data = append(data, d)
	}
	return string(data[:len(data)])
}`

var response = `package helpers

import (
	"strings"
)

type ResponseFormat struct {
	Data map[string]interface{}
}

var (
	Rf = ResponseFormat{}
)

func (r *ResponseFormat) Success(httpMethod string, id int, d ...[]interface{}) {
	switch httpMethod {
	case "POST":
		r.Data = make(map[string]interface{})
		r.Data["success"] = true
		r.Data["id"] = id
	case "GET":
		if d != nil {
			r.Data["data"] = d[0]
		}
	default:
		r.Data = make(map[string]interface{})
		r.Data["success"] = true
	}

}

func (r *ResponseFormat) Fail(errorData interface{}) {
	r.Data = make(map[string]interface{})
	r.Data["success"] = false

	switch errorData.(type) {
	case string:
		r.Data["error"] = map[string]string{
			"orm": ClearErrorPrefix(errorData.(string)),
		}
	default:
		r.Data["error"] = errorData
	}
}

func ClearErrorPrefix(s string) string {
	strToRemove := "<QuerySeter> "
	s = strings.TrimPrefix(s, strToRemove)
	return s
}`


func TemplateHelper() string {
	return helper
}

func TemplateResponse() string {
	return response
}