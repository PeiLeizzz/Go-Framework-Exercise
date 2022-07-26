package router

import (
	"filestore-server/rpc/apigw/handler"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static/", "./static/")
	router.GET("/user/signup", handler.GetSignUpHandler)
	router.POST("/user/signup", handler.PostSignUpHandler)
	router.POST("/file/upload", handler.PostUploadHandler)
	return router
}
