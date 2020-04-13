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
	beego.Router("/api/namespaces/debugimage", &controllers.MainController{}, "get:DebugImage")
	beego.Router("/api/terminal", &controllers.MainController{}, "get:ContainerTerminal")
	// beego.Include(&controllers.MainController{})
	beego.Router("/api/terminal/asd/sock", &websocket.TSockjs{}, "get:ServeHTTP")
}
