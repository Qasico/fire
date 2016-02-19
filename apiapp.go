package main

import (
	"os"
	"fmt"
	"strings"
	path "path/filepath"

	"github.com/qasico/fire/stubs"
	"github.com/qasico/fire/generator"
	"github.com/qasico/fire/helper"
)

var connection = "root:@tcp(127.0.0.1:3306)/{{.database}}"
var default_db = "konektifa_app"
var cmdApiapp = &Command{
	UsageLine: "api [appname]",
	Short:     "create an API application",
	Long: `
Create an API application.

fire api [appname] [-database=""] [-tables=""] [-driver=mysql] [-conn=root:@tcp(127.0.0.1:3306)/test]
    -tables: a list of table names separated by ',' (default is empty, indicating all tables)
    -driver: [mysql | postgres | sqlite] (default: mysql)
    -conn:   the connection string used by the driver, the default is '127.0.0.1:3306'
             e.g. for mysql:    root:@tcp(127.0.0.1:3306)/test
             e.g. for postgres: postgres://postgres:postgres@127.0.0.1:5432/postgres
`,
}

func init() {
	cmdApiapp.Run = createapi
	cmdApiapp.Flag.Var(&database, "database", "specify database to generate api")
	cmdApiapp.Flag.Var(&tables, "tables", "specify tables to generate model")
	cmdApiapp.Flag.Var(&driver, "driver", "database driver: mysql, postgresql, etc.")
	cmdApiapp.Flag.Var(&conn, "conn", "connection string used by the driver to connect to a database instance")
}

func createapi(cmd *Command, args []string) int {
	curpath, _ := os.Getwd()
	if len(args) < 1 {
		helper.ColorLog("[ERRO] Argument [appname] is missing\n")
		os.Exit(2)
	}
	if len(args) > 1 {
		cmd.Flag.Parse(args[1:])
	}
	apppath, packpath, err := checkEnv(args[0])
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	if driver == "" {
		driver = "mysql"
	}
	if conn == "" {

	}

	if conn != "" {
		connection = string(conn)
	} else if database != "" {
		default_db = string(database)
		connection = strings.Replace(connection, "{{.database}}", default_db, -1)
	} else {
		connection = strings.Replace(connection, "{{.database}}", args[0], -1)
	}

	//
	// Creating directory stucture
	// ----------------------------
	fmt.Println("create app folder:", apppath)
	os.MkdirAll(apppath, 0755)

	fmt.Println("create controllers:", path.Join(apppath, "controllers"))
	os.Mkdir(path.Join(apppath, "controllers"), 0755)

	fmt.Println("create docs:", path.Join(apppath, "docs"))
	os.Mkdir(path.Join(apppath, "docs"), 0755)

	fmt.Println("create helpers:", path.Join(apppath, "helpers"))
	os.Mkdir(path.Join(apppath, "helpers"), 0755)

	fmt.Println("create tests:", path.Join(apppath, "tests"))
	os.Mkdir(path.Join(apppath, "tests"), 0755)

	fpath := ""

	//
	// Stubbing helpers packages
	// ----------------------------
	fpath = path.Join(apppath, "helpers", "global_function.go")
	writetofile(fpath, stubs.TemplateHelper())
	helper.ColorLog("[INFO] helper => %s\n", fpath)

	fpath = path.Join(apppath, "helpers", "response_formater.go")
	writetofile(fpath, stubs.TemplateResponse())
	helper.ColorLog("[INFO] helper => %s\n", fpath)

	fpath = path.Join(apppath, ".env")
	ac := strings.Replace(stubs.TemplateEnv(), "{{.Appname}}", args[0], -1);
	writetofile(fpath, strings.Replace(ac, "{{.database}}", string(default_db), -1))
	helper.ColorLog("[INFO] .env => %s\n", fpath)

	fpath = path.Join(apppath, "main.go")
	maingoContent := strings.Replace(stubs.TemplateMain(), "{{.Appname}}", packpath, -1)
	maingoContent = strings.Replace(maingoContent, "{{.DriverName}}", string(driver), -1)
	if driver == "mysql" {
		maingoContent = strings.Replace(maingoContent, "{{.DriverPkg}}", `_ "github.com/go-sql-driver/mysql"`, -1)
	} else if driver == "postgres" {
		maingoContent = strings.Replace(maingoContent, "{{.DriverPkg}}", `_ "github.com/lib/pq"`, -1)
	}
	writetofile(fpath, strings.Replace(maingoContent, "{{.conn}}", connection, -1))
	helper.ColorLog("[INFO] main => %s\n", fpath)

	helper.ColorLog("[SUCC] Using '%s' as 'driver'\n", driver)
	helper.ColorLog("[SUCC] Using '%s' as 'conn'\n", connection)
	helper.ColorLog("[SUCC] Using '%s' as 'tables'\n", tables)

	generator.GenerateAppcode(string(driver), string(connection), "all", string(tables), path.Join(curpath, args[0]))

	return 0
}

func checkEnv(appname string) (apppath, packpath string, err error) {
	curpath, err := os.Getwd()
	if err != nil {
		return
	}

	gopath := os.Getenv("GOPATH")
	helper.Debugf("gopath:%s", gopath)
	if gopath == "" {
		err = fmt.Errorf("you should set GOPATH in the env")
		return
	}

	appsrcpath := ""
	haspath := false
	wgopath := path.SplitList(gopath)
	for _, wg := range wgopath {
		wg, _ = path.EvalSymlinks(path.Join(wg, "src"))

		if path.HasPrefix(strings.ToLower(curpath), strings.ToLower(wg)) {
			haspath = true
			appsrcpath = wg
			break
		}
	}

	if !haspath {
		err = fmt.Errorf("can't create application outside of GOPATH `%s`\n" +
		"you first should `cd $GOPATH%ssrc` then use create\n", gopath, string(path.Separator))
		return
	}
	apppath = path.Join(curpath, appname)

	if _, e := os.Stat(apppath); os.IsNotExist(e) == false {
		err = fmt.Errorf("path `%s` exists, can not create app without remove it\n", apppath)
		return
	}
	packpath = strings.Join(strings.Split(apppath[len(appsrcpath) + 1:], string(path.Separator)), "/")
	return
}
