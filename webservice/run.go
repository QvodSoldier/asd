package webservice

import (
	_ "ggstudy/asd/webservice/routers"

	"github.com/astaxie/beego"
)

// 1.写informer-前端页面展示数据
// 2.k8s webshell
// 3.实时更新debugtask
// 4.修改log文件地址
func Run() {
	beego.BConfig.Listen.HTTPPort = 8081
	beego.SetStaticPath("/assets", "./webservice/static/assets")
	beego.SetStaticPath("/public", "./webservice/static")
	beego.Run()
}
