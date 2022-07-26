package handler

import (
	"context"
	"filestore-server/common"
	"filestore-server/config"
	"filestore-server/db"
	"filestore-server/rpc/account/proto"
	"filestore-server/util"
	"fmt"
)

type User struct{}

func (u *User) Signup(ctx context.Context, req *proto.ReqSignup, res *proto.RespSignup) error {
	username := req.Username
	password := req.Password

	if len(username) < 3 || len(password) < 5 {
		res.Code = common.StatusParamInvalid
		res.Message = "wrong username or password"
		return nil
	}

	encPassword := util.Sha1([]byte(password + config.Pwd_salt))
	ok, err := db.UserSignUp(username, encPassword)
	if err != nil {
		res.Code = common.InternelServerError
		res.Message = fmt.Sprintf("Failed to save user information, err: %s\n", err.Error())
		return err
	} else if !ok {
		res.Code = common.StatusParamInvalid
		res.Message = "user has signed up before"
		return nil
	}

	res.Code = common.StatusOK
	res.Message = "user sign up success"
	return nil
}
