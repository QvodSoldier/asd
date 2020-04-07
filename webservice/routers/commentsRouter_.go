package routers

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context/param"
)

//30873
func init() {

	beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"] = append(beego.GlobalControllerRouter["k8-web-terminal/controllers:MainController"],
		beego.ControllerComments{
			Method:           "Get",
			Router:           `/`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"] = append(beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"],
		beego.ControllerComments{
			Method:           "Namespaces",
			Router:           `/api/namespaces`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"] = append(beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"],
		beego.ControllerComments{
			Method:           "NamespacePods",
			Router:           `/api/namespaces/pods`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

	beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"] = append(beego.GlobalControllerRouter["ggstudy/asd/webservice/controllers:MainController"],
		beego.ControllerComments{
			Method:           "ContainerTerminal",
			Router:           `/container/terminal`,
			AllowHTTPMethods: []string{"get"},
			MethodParams:     param.Make(),
			Filters:          nil,
			Params:           nil})

}
