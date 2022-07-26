package router

import (
	"filestore-server/handler/ginHandler"
	"filestore-server/middleware"
	"github.com/gin-gonic/gin"
)

func Router() *gin.Engine {
	router := gin.Default()

	router.Static("/static/", "./static")

	fileGroup := router.Group("/file")
	fileGroup.Use(middleware.GinHTTPInterceptor())
	{
		fileGroup.GET("/upload", ginHandler.GetUploadHandler)
		fileGroup.POST("/upload", ginHandler.PostUploadHandler)
		fileGroup.POST("/fastupload", ginHandler.PostTryFastUploadHandler)
		fileGroup.GET("/upload/success", ginHandler.GetUploadSuccessHandler)
		fileGroup.GET("/meta", ginHandler.GetFileMetaHandler)
		fileGroup.POST("/query", ginHandler.GetFileQueryHandler)
		fileGroup.GET("/download", ginHandler.GetDownloadHandler)
		fileGroup.GET("/downloadurl", ginHandler.GetDownloadURLHandler)
		fileGroup.POST("/update", ginHandler.PostFileMetaUpdateHandler)
		fileGroup.GET("/delete", ginHandler.GetFileDeleteHandler)
	}

	mpuploadGroup := fileGroup.Group("/mpupload")
	{
		mpuploadGroup.GET("/init", ginHandler.GetInitialMultipartUploadHandler)
		mpuploadGroup.GET("/uppart", ginHandler.GetUploadPartHandler)
		mpuploadGroup.GET("/complete", ginHandler.GetCompleteUploadHandler)
	}

	userGroup := router.Group("/user")
	{
		userGroup.GET("/signup", ginHandler.GetSignUpHandler)
		userGroup.POST("/signup", ginHandler.PostSignUpHandler)
		userGroup.GET("/signin", ginHandler.GetSignInHandler)
		userGroup.POST("/signin", ginHandler.PostSignInHandler)
	}

	userInfo := router.Group("/user/info")
	userInfo.Use(middleware.GinHTTPInterceptor())
	{
		userInfo.POST("", ginHandler.GetUserInfoHandler)
	}

	return router
}
