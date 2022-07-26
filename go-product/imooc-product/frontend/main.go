package main

import (
	"context"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go-product/RabbitMQ"
	"go-product/imooc-product/common"
	"go-product/imooc-product/frontend/middleware"
	"go-product/imooc-product/frontend/web/controllers"
	"go-product/imooc-product/repositories"
	"go-product/imooc-product/services"
	"log"
)

func main() {
	app := iris.New()

	app.Logger().SetLevel("debug")

	template := iris.HTML("./frontend/web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(template)

	app.HandleDir("/public", "./frontend/web/public")
	app.HandleDir("/html", "./frontend/web/htmlProductShow")

	app.OnAnyErrorCode(func(ctx iris.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错！"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})

	db, err := common.NewMysqlConn()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	userRepository := repositories.NewUserManager("user", db)
	userService := services.NewUserService(userRepository)
	user := mvc.New(app.Party("/user"))
	user.Register(ctx, userService)
	user.Handle(new(controllers.UserController))

	productRepository := repositories.NewProductManager("product", db)
	productService := services.NewProductService(productRepository)
	orderRepository := repositories.NewOrderManager("order", db)
	orderService := services.NewOrderService(orderRepository)
	rabbitmq := RabbitMQ.NewRabbitMQSimple("imoocProduct")
	pro := app.Party("/product")
	pro.Use(middleware.AuthConProduct)
	product := mvc.New(pro)
	product.Register(ctx, productService, orderService, rabbitmq)
	product.Handle(new(controllers.ProductController))

	app.Run(
		iris.Addr("localhost:8081"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
