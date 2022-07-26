package middleware

import (
	"filestore-server/util"
	"github.com/gin-gonic/gin"
	"net/http"
)

// HTTPInterceptor: 请求拦截器
func HTTPInterceptor(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		var username string
		var token string

		if req.Method == "GET" {
			req.ParseForm()
			username = req.Form.Get("username")
			token = req.Form.Get("token")
			if username == "" {
				username = req.PostFormValue("username")
				token = req.PostFormValue("token")
			}
		} else if req.Method == "POST" {
			username = req.PostFormValue("username")
			token = req.PostFormValue("token")
			if username == "" {
				req.ParseForm()
				username = req.Form.Get("username")
				token = req.Form.Get("token")
			}
		}

		if len(username) < 3 || !util.IsTokenValid(username, token) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h(w, req)
	}
}

// Gin 版拦截器
func GinHTTPInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.String() == "/file/upload/success" {
			c.Next()
			return
		}
		var username string
		var token string

		username = c.Query("username")
		token = c.Query("token")
		if username == "" {
			username = c.PostForm("username")
			token = c.PostForm("token")
		}

		if len(username) < 3 || !util.IsTokenValid(username, token) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"msg":  "wrong token",
				"code": -2,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
