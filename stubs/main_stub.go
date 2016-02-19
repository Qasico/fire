package stubs

var mainTemplate = `package main

import (
	_ "{{.Appname}}/docs"
	_ "{{.Appname}}/routers"
	_ "github.com/go-sql-driver/mysql"

	"os"
	"time"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/plugins/cors"
)

func main() {
	if envRunMode := os.Getenv("API_RUNMODE"); envRunMode != "" {
		beego.RunMode = envRunMode
	}

	if os.Getenv("APP_DEBUG") == "true" {
		beego.RunMode = "debug"
		beego.EnableDocs = true
		beego.DirectoryIndex = true
		beego.StaticDir["/swagger"] = "swagger"
		orm.Debug = true
	}

	mysqlServer := os.Getenv("DB_HOST")
	mysqlUser := os.Getenv("DB_USER")
	mysqlPass := os.Getenv("DB_PASS")
	mysqlDb := os.Getenv("DB_NAME")

	orm.RegisterDataBase("default", "mysql", mysqlUser + ":" + mysqlPass + "@tcp(" + mysqlServer + ":3306)/" + mysqlDb)
	orm.DefaultTimeLoc = time.UTC

	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "DELETE", "PUT", "PATCH", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	orm.DefaultRelsDepth = 3

	beego.Run()
}`

func TemplateMain() string {
	return mainTemplate
}