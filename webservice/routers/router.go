package routers

import (
	"ggstudy/asd/webservice/controllers"
	"ggstudy/asd/webservice/controllers/websocket"

	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/asd/sock", &websocket.WSController{})
}
