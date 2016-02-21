package stubs

var mainTemplate = `package main

import (
	"os"
	"time"

	"github.com/qasico/beego"
	"github.com/qasico/beego/orm"
	"github.com/qasico/beego/plugins/cors"

	_ "{{.Appname}}/docs"
	_ "{{.Appname}}/routers"
	_ "github.com/go-sql-driver/mysql"
)

func init() {
	orm.RegisterDataBase("default", "mysql", os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASS") + "@tcp(" + os.Getenv("DB_HOST") + ":3306)/" + os.Getenv("DB_NAME"))
	orm.DefaultTimeLoc = time.UTC
	orm.DefaultRelsDepth = 3

	if os.Getenv("APP_DEBUG") == "true" {
		orm.Debug = true
	}
}

func main() {
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "DELETE", "PUT", "PATCH", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	beego.Run()
}`

func TemplateMain() string {
	return mainTemplate
}