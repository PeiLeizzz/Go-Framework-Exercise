package routers

import (
	"net/http"

	"github.com/PeiLeizzz/go-gin-example/middleware/jwt"
	"github.com/PeiLeizzz/go-gin-example/pkg/export"
	"github.com/PeiLeizzz/go-gin-example/pkg/qrcode"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/pkg/upload"
	"github.com/PeiLeizzz/go-gin-example/routers/api"
	v2 "github.com/PeiLeizzz/go-gin-example/routers/api/v2"
	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"

	_ "github.com/PeiLeizzz/go-gin-example/docs"
)

func InitRouter() *gin.Engine {
	gin.SetMode(setting.ServerSetting.RunMode)
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.POST("/auth", api.GetAuth)
	r.POST("/upload", api.UploadImage)
	r.StaticFS("/upload/images", http.Dir(upload.GetImageFullPath()))
	r.StaticFS("/export/tags", http.Dir(export.GetExcelFullPath()))
	r.StaticFS("/qrcode", http.Dir(qrcode.GetQrCodeFullPath()))

	apiv2 := r.Group("/api/v2")
	apiv2.Use(jwt.JWT())
	{
		apiv2.GET("/tags", v2.GetTags)
		apiv2.POST("/tags", v2.AddTag)
		apiv2.PUT("/tags/:id", v2.EditTag)
		apiv2.DELETE("/tags/:id", v2.DeleteTag)
		apiv2.POST("/tags/export", v2.ExportTag)
		apiv2.POST("/tags/import", v2.ImportTag)
	}

	{
		apiv2.GET("/articles", v2.GetArticles)
		apiv2.GET("/articles/:id", v2.GetArticle)
		apiv2.POST("/articles", v2.AddArticle)
		apiv2.PUT("/articles/:id", v2.EditArticle)
		apiv2.DELETE("/articles/:id", v2.DeleteArticle)
		apiv2.POST("/articles/poster/generate", v2.GenerateArticlePoster)
	}
	return r
}
