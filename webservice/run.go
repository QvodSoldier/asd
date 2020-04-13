package webservice

import (
	_ "ggstudy/asd/webservice/config"
	_ "ggstudy/asd/webservice/routers"

	"github.com/astaxie/beego"
)

func Run() {
	beego.BConfig.Listen.HTTPPort = 8081
	beego.SetStaticPath("/assets", "./webservice/static/assets")
	beego.SetStaticPath("/public", "./webservice/static")
	beego.SetViewsPath("webservice/views")
	beego.Run()
}
