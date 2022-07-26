package main

import "github.com/kataras/iris/v12"

func main() {
	app := iris.New()

	app.HandleDir("/public", "./front/web/public")
	app.HandleDir("/html", "./front/web/htmlProductShow")

	app.Run(
		iris.Addr("localhost:8082"),
		iris.WithoutServerError(iris.ErrServerClosed),
		iris.WithOptimizations,
	)
}
