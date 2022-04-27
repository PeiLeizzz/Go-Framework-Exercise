package jwt

import (
	"net/http"
	"time"

	"github.com/PeiLeizzz/go-gin-example/pkg/e"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}

		code = e.SUCCESS
		token := c.Query("token")
		if token == "" {
			code = e.INVALID_PARMAS
		} else {
			claims, err := util.ParseToken(token)
			// 有问题：超时了直接在 err 里显示了，而不会走下面的分支
			if err != nil {
				code = e.ERROR_AUTH_CHECK_TOKEN_FAIL
			} else if time.Now().Unix() > claims.ExpiresAt {
				code = e.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
			}
		}

		if code != e.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  e.GetMsg(code),
				"data": data,
			})

			c.Abort() // 阻断之后的中间件
			return
		}

		c.Next()
	}
}
