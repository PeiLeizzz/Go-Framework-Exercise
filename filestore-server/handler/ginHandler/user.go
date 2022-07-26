package ginHandler

import (
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

func internelServerError(c *gin.Context, msg string) {
	c.String(http.StatusInternalServerError, "%s", msg)
	return
}

// 处理用户注册(POST)
func PostSignUpHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	if len(username) < 3 || len(password) < 5 {
		c.String(http.StatusBadRequest, "%s", "invalid parameter")
		return
	}

	encPassword := util.Sha1([]byte(password + config.Pwd_salt))
	ok, err := db.UserSignUp(username, encPassword)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to save user information, err: %s\n", err))
		return
	} else if !ok {
		c.String(http.StatusBadRequest, "%s", "user has signed up before")
		return
	}

	c.String(http.StatusOK, "%s", "user sign up success")
}

// 处理用户注册(GET)
func GetSignUpHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

// 处理用户登陆(POST)
func PostSignInHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")
	encPassword := util.Sha1([]byte(password + config.Pwd_salt))

	// 1. 校验用户名及密码
	ok, err := db.UserSignIn(username, encPassword)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to confirm user password, err: %s\n", err))
		return
	} else if !ok {
		c.String(http.StatusBadRequest, "%s", "No such user or wrong password")
		return
	}

	// 2. 生成 token
	token := util.GenToken(username)
	_, err = db.UpdateToken(username, token) // ignore bool
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to generate token, err: %s\n", err))
		return
	}

	// 3. 返回相关信息
	resp := util.RespMsg{
		Code: 200,
		Msg:  "success",
		Data: struct {
			Location string
			Username string
			Token    string
		}{
			Location: "http://" + c.Request.Host + "/static/view/home.html",
			Username: username,
			Token:    token,
		},
	}
	resBytes, err := resp.JSONString()
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to convert response to bytes, err: %s\n", err))
		return
	}

	c.String(http.StatusOK, "%s", resBytes)
}

// 处理用户登陆(GET)
func GetSignInHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signin.html")
}

// 处理查询用户信息(POST)
func GetUserInfoHandler(c *gin.Context) {
	username := c.PostForm("username")

	// 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to query user info, err: %s\n", err.Error()))
		return
	}

	// 组装并返回相关信息
	resp := util.RespMsg{
		Code: 200,
		Msg:  "success",
		Data: user,
	}
	resBytes, err := resp.JSONString()
	if err != nil {
		internelServerError(c, fmt.Sprintf("Failed to convert response to bytes, err: %s\n", err))
		return
	}

	c.String(http.StatusOK, "%s", resBytes)
}
