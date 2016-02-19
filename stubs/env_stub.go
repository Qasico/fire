package stubs

var envTemplate = `APP_NAME={{.Appname}}
APP_DEBUG=true
API_ADDR=192.168.10.10
API_PORT=8080
API_KEY=123456
API_VERSION=1
API_RUNMODE=dev
DB_HOST=localhost
DB_NAME={{.database}}
DB_USER=root
DB_PASS=`

func TemplateEnv() string {
	return envTemplate
}