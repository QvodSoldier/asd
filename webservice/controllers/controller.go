package controllers

import (
	"log"

	"ggstudy/asd/webservice/models"

	"github.com/astaxie/beego"
	"k8s.io/apimachinery/pkg/labels"
)

var (
	cache = models.NewCache()
)

type MainController struct {
	beego.Controller
}

func (c *MainController) URLMapping() {
	c.Mapping("Namespaces", c.Namespaces)
	c.Mapping("NamespacePods", c.NamespacePods)
	c.Mapping("ContainerTerminal", c.ContainerTerminal)
}

// @router / [get]
func (c *MainController) Get() {
	log.Println("caonima")
	c.TplName = "index.html"
}

// @router /api/namespaces [get]
func (c *MainController) Namespaces() {
	namespaces, err := cache.NameSpaceLister.List(labels.Everything())
	if err != nil {
		log.Println(err)
	}

	c.Data["json"] = namespaces
	c.ServeJSON()
}

func (c *MainController) NamespacePods() {
	namespace := c.GetString("namespace")
	pods, err := cache.PodLister.Pods(namespace).List(labels.Everything())
	if err != nil {
		log.Println(err)
	}

	c.Data["json"] = pods
	c.ServeJSON()
}

func (c *MainController) ContainerTerminal() {
	c.TplName = "terminal.html"
}
