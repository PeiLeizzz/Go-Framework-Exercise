package util

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/unknwon/com"

	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
)

func GetPage(c *gin.Context) int {
	result := 0
	page, err := com.StrTo(c.Query("page")).Int()

	if err != nil {
		log.Fatalf("Failed to convert param 'page'(%s) to int: %v", c.Query("page"), err)
	}

	if page > 0 {
		result = (page - 1) * setting.PageSize
	}
	return result
}
