package controllers

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/mvc"
	"go-product/imooc-product/datamodels"
	"go-product/imooc-product/encrypt"
	"go-product/imooc-product/services"
	"go-product/imooc-product/tool"
	"strconv"
)

type UserController struct {
	Ctx     iris.Context
	Service services.IUserService
}

func (c *UserController) GetRegister() mvc.View {
	return mvc.View{
		Name: "user/register.html",
	}
}

func (c *UserController) PostRegister() {
	var (
		nickName = c.Ctx.FormValue("nickName")
		userName = c.Ctx.FormValue("userName")
		password = c.Ctx.FormValue("password")
	)

	user := &datamodels.User{
		UserName:     userName,
		NickName:     nickName,
		HashPassword: password,
	}

	_, err := c.Service.InsertUser(user)
	if err != nil {
		c.Ctx.Application().Logger().Debug(err)
		c.Ctx.Redirect("/user/error")
		return
	}
	c.Ctx.Redirect("/user/login")
	return
}

func (c *UserController) GetLogin() mvc.View {
	return mvc.View{
		Name: "user/login",
	}
}

func (c *UserController) PostLogin() mvc.Response {
	var (
		userName = c.Ctx.FormValue("userName")
		password = c.Ctx.FormValue("password")
	)

	userID, ok := c.Service.IsPwdSuccess(userName, password)
	if !ok {
		return mvc.Response{
			Path: "user/login",
		}
	}

	uidString := strconv.FormatInt(userID, 10)
	uidEnc, err := encrypt.EnPwdCode([]byte(uidString))
	if err != nil {
		return mvc.Response{
			Path: "/shared/error.html",
		}
	}
	tool.GlobalCookie(c.Ctx, "uid", uidString)
	tool.GlobalCookie(c.Ctx, "sign", uidEnc)

	return mvc.Response{
		Path: "product/",
	}
}
