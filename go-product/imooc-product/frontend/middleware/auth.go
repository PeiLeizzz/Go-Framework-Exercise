package middleware

import (
	"github.com/kataras/iris/v12"
	"go-product/imooc-product/encrypt"
)

//var (
//	Session = sessions.New(sessions.Config{
//		Cookie:  "AdminCookie",
//		Expires: 600 * time.Minute,
//	})
//)

func AuthConProduct(ctx iris.Context) {
	//sess := Session.Start(ctx)
	//if isLogin, err := sess.GetBoolean("isLogin"); err != nil {
	//	ctx.Application().Logger().Debug(err)
	//	ctx.Redirect("/user/login")
	//	return
	//} else if !isLogin {
	//	ctx.Application().Logger().Debug("必须先登录")
	//	ctx.Redirect("/user/login")
	//	return
	//}

	uid := ctx.GetCookie("uid")
	sign := ctx.GetCookie("sign")
	if uid == "" || sign == "" {
		ctx.Application().Logger().Debug("必须先登录")
		ctx.Redirect("/user/login")
		return
	}

	signByte, err := encrypt.DePwdCode(sign)
	if err != nil {
		ctx.Application().Logger().Debug(err)
		ctx.Redirect("/user/login")
		return
	}

	if string(signByte) != uid {
		ctx.Application().Logger().Debug("身份认证失败")
		ctx.Redirect("/user/login")
		return
	}
	ctx.Next()
}
