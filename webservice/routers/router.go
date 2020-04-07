package routers

import (
	"ggstudy/asd/webservice/controllers"
	"ggstudy/asd/webservice/controllers/websocket"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/api/namespaces", &controllers.MainController{}, "get:Namespaces")
	beego.Router("/api/namespaces/pods", &controllers.MainController{}, "get:NamespacePods")
	beego.Router("/asd/sock", &websocket.WSController{})
}
