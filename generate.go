package main

import (
	"os"
	"fmt"

	"github.com/qasico/fire/generator"
	"github.com/qasico/fire/helper"
)

var cmdGenerate = &Command{
	UsageLine: "generate [Command]",
	Short:     "source code generator",
	Long: `
fire generate docs
    generate swagger doc file

fire generate test [routerfile]
    generate testcase

fire generate appcode [-mode=all] [-database=test] [-tables=""] [-driver=mysql] [-conn="root:@tcp(127.0.0.1:3306)/test"]
    generate appcode based on an existing database
    -level:  [m | mc | r | all], m = models; mc = models,controllers; r = router; all = models,controllers,router;
    -database: database name
    -tables: a list of table names separated by ',', default is empty, indicating all tables
    -driver: [mysql | postgres | sqlite], the default is mysql
    -conn:   the connection string used by the driver.
             default for mysql:    root:@tcp(127.0.0.1:3306)/test
             default for postgres: postgres://postgres:postgres@127.0.0.1:5432/postgres
`,
}

var driver docValue
var conn docValue
var level docValue
var tables docValue
var fields docValue
var database docValue

func init() {
	cmdGenerate.Run = generateCode
	cmdGenerate.Flag.Var(&database, "database", "specify the database want to use.")
	cmdGenerate.Flag.Var(&tables, "tables", "specify tables to generate model")
	cmdGenerate.Flag.Var(&driver, "driver", "database driver: mysql, postgresql, etc.")
	cmdGenerate.Flag.Var(&conn, "conn", "connection string used by the driver to connect to a database instance")
	cmdGenerate.Flag.Var(&level, "mode", "1 = models only; 2 = models and controllers; 3 = models, controllers and routers")
	cmdGenerate.Flag.Var(&fields, "fields", "specify the fields want to generate.")
}

func generateCode(cmd *Command, args []string) int {
	curpath, _ := os.Getwd()
	if len(args) < 1 {
		helper.ColorLog("[ERRO] command is missing\n")
		os.Exit(2)
	}

	gopath := os.Getenv("GOPATH")
	helper.Debugf("gopath:%s", gopath)
	if gopath == "" {
		helper.ColorLog("[ERRO] $GOPATH not found\n")
		helper.ColorLog("[HINT] Set $GOPATH in your environment vairables\n")
		os.Exit(2)
	}

	gcmd := args[0]
	switch gcmd {
	case "docs":
		generator.GenerateDocs(curpath)
	case "appcode":
		// load config
		err := loadConfig()

		if err != nil {
			helper.ColorLog("[ERRO] Fail to parse fire.json[ %s ]\n", err)
		}

		cmd.Flag.Parse(args[1:])

		if database != "" {
			newConn := fmt.Sprint("root:@tcp(127.0.0.1:3306)/", database)
			conn = docValue(newConn)
		}

		if driver == "" {
			driver = docValue(conf.Database.Driver)
			if driver == "" {
				driver = "mysql"
			}
		}
		if conn == "" {
			conn = docValue(conf.Database.Conn)
			if conn == "" {
				if driver == "mysql" {
					conn = "root:@tcp(127.0.0.1:3306)/test"
				} else if driver == "postgres" {
					conn = "postgres://postgres:postgres@127.0.0.1:5432/postgres"
				}
			}
		}
		if level == "" {
			level = "all"
		}
		helper.ColorLog("[INFO] Using '%s' as 'driver'\n", driver)
		helper.ColorLog("[INFO] Using '%s' as 'conn'\n", conn)
		helper.ColorLog("[INFO] Using '%s' as 'tables'\n", tables)
		helper.ColorLog("[INFO] Using '%s' as 'level'\n", level)
		generator.GenerateAppcode(driver.String(), conn.String(), level.String(), tables.String(), curpath)
	default:
		helper.ColorLog("[ERRO] command is missing\n")
	}
	helper.ColorLog("[SUCC] generate successfully created!\n")
	return 0
}
