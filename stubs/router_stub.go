package stubs

var routerTemplate = `// @APIVersion 1.0.0
// @Title Application API
// @Description application api
// @License Apache 2.0
// @LicenseUrl http://www.apache.org/licenses/LICENSE-2.0.html
package routers

import (
	"{{pkgPath}}/controllers"

	"github.com/astaxie/beego"
)

func init() {
	ns := beego.NewNamespace("/v1",
		{{nameSpaces}}
	)
	beego.AddNamespace(ns)
}`

func TemplateRouter() string {
	return routerTemplate
}