package handler

import (
	"context"
	"filestore-server/rpc/account/proto"
	"github.com/gin-gonic/gin"
	"github.com/go-micro/plugins/v4/registry/consul"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"log"
	"net/http"
)

var (
	userCli proto.UserService
)

func init() {
	reg := consul.NewRegistry(func(op *registry.Options) {
		op.Addrs = []string{
			"127.0.0.1:8500",
		}
	})
	service := micro.NewService(
		micro.Registry(reg),
	)
	service.Init()

	userCli = proto.NewUserService("go.micro.service.user", service.Client())
}

// 处理用户注册(GET)
func GetSignUpHandler(c *gin.Context) {
	c.Redirect(http.StatusFound, "/static/view/signup.html")
}

// 处理用户注册(POST)
func PostSignUpHandler(c *gin.Context) {
	username := c.PostForm("username")
	password := c.PostForm("password")

	resp, err := userCli.Signup(context.TODO(), &proto.ReqSignup{
		Username: username,
		Password: password,
	})
	if err != nil {
		c.Status(http.StatusInternalServerError)
		log.Println(err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": resp.Code,
		"msg":  resp.Message,
	})
}
