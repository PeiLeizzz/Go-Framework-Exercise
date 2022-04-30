package api

import (
	"net/http"

	"github.com/PeiLeizzz/go-gin-example/pkg/app"
	"github.com/PeiLeizzz/go-gin-example/pkg/e"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
	"github.com/PeiLeizzz/go-gin-example/service/auth_service"
	"github.com/gin-gonic/gin"
)

type auth struct {
	Username string `form:"username" valid:"Required; MaxSize(50)"`
	Password string `form:"password" valid:"Required; MaxSize(50)"`
}

// @Summary 获取 token
// @Accept mpfd
// @Produce json
// @Param username formData string true "username"
// @Param password formData string true "password"
// @Success 200 {object} app.Response
// @Failure 500 {object} app.Response
// @Router /auth [post]
func GetAuth(c *gin.Context) {
	var appG = app.Gin{C: c}
	var form auth

	httpCode, errCode := app.BindAndValid(c, &form)
	if errCode != e.SUCCESS {
		appG.Response(httpCode, errCode, nil)
		return
	}

	authService := auth_service.Auth{
		Username: form.Username,
		Password: form.Password,
	}
	exists, err := authService.Check()
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_AUTH_CHECK_TOKEN_FAIL, nil)
		return
	}
	if !exists {
		appG.Response(http.StatusUnauthorized, e.ERROR_AUTH, nil)
	}

	token, err := util.GenerateToken(form.Username, form.Password)
	if err != nil {
		appG.Response(http.StatusInternalServerError, e.ERROR_AUTH_TOKEN, nil)
		return
	}

	data := map[string]string{"token": token}
	appG.Response(http.StatusOK, e.SUCCESS, data)
}
