package main

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go-product/irisDemo/web/controllers"
)

func main() {
	app := iris.New()
	app.Logger().SetLevel("debug")

	// 注册模板
	app.RegisterView(iris.HTML("./web/views", ".html"))
	// 注册控制器
	mvc.New(app.Party("/hello")).Handle(new(controllers.MovieController))

	app.Run(iris.Addr("localhost:8080"))
}
