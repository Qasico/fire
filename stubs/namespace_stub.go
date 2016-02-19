package stubs

var namespaceTemplate = `
		beego.NSNamespace("/{{nameSpace}}",
			beego.NSInclude(
				&controllers.{{ctrlName}}Controller{},
			),
		),
`

func TemplateNamespace() string {
	return namespaceTemplate
}