package main

import (
	"context"
	"github.com/kataras/iris/v12"
	irisContext "github.com/kataras/iris/v12/context"
	"github.com/kataras/iris/v12/mvc"
	"go-product/imooc-product/backend/web/controllers"
	"go-product/imooc-product/common"
	"go-product/imooc-product/repositories"
	"go-product/imooc-product/services"
	"log"
)

func main() {
	// 创建 iris 实例
	app := iris.New()

	// 设置错误模式
	app.Logger().SetLevel("debug")

	// 注册模版，根目录为 imooc-product/
	template := iris.HTML("./backend/web/views", ".html").Layout("shared/layout.html").Reload(true)
	app.RegisterView(template)

	// 设置静态目录，url 中 /asserts 会指向本机 ./backend/web/assets 目录
	app.HandleDir("/assets", "./backend/web/assets")

	// 出现异常跳转到指定页面
	app.OnAnyErrorCode(func(ctx *irisContext.Context) {
		ctx.ViewData("message", ctx.Values().GetStringDefault("message", "访问的页面出错"))
		ctx.ViewLayout("")
		ctx.View("shared/error.html")
	})

	// 连接数据库
	db, err := common.NewMysqlConn()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册控制器
	productRepository := repositories.NewProductManager("product", db)
	productService := services.NewProductService(productRepository)
	product := mvc.New(app.Party("/product"))
	product.Register(ctx, productService)
	product.Handle(new(controllers.ProductController))

	orderRepository := repositories.NewOrderManager("order", db)
	orderService := services.NewOrderService(orderRepository)
	order := mvc.New(app.Party("/order"))
	order.Register(ctx, orderService)
	order.Handle(new(controllers.OrderController))

	// 启动服务
	app.Run(
		iris.Addr("localhost:8080"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
