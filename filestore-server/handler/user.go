package handler

import (
	"filestore-server/db"
	"filestore-server/util"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	pwd_salt = "*#98Xdf"
)

// SignUpHandler: 处理用户注册(GET/POST)
func SignUpHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		data, err := ioutil.ReadFile("./static/view/signup.html")
		if err != nil {
			internelServerError(w, fmt.Sprintf("Failed to read signup page, err: %s\n", err))
			return
		}
		w.Write(data)
		return
	} else if req.Method == "POST" {
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")

		if len(username) < 3 || len(password) < 5 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("invalid parameter"))
			return
		}

		encPassword := util.Sha1([]byte(password + pwd_salt))
		ok, err := db.UserSignUp(username, encPassword)
		if err != nil {
			internelServerError(w, fmt.Sprintf("Failed to save user information, err: %s\n", err))
			return
		} else if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("user has signed up before"))
			return
		}

		w.Write([]byte("user sign up success"))
	}
}

// SignInHandler：处理用户登陆(GET/POST)
func SignInHandler(w http.ResponseWriter, req *http.Request) {
	if req.Method == "GET" {
		http.Redirect(w, req, "/static/view/signin.html", http.StatusFound)
		return
	} else if req.Method == "POST" {
		username := req.PostFormValue("username")
		password := req.PostFormValue("password")
		encPassword := util.Sha1([]byte(password + pwd_salt))

		// 1. 校验用户名及密码
		ok, err := db.UserSignIn(username, encPassword)
		if err != nil {
			internelServerError(w, fmt.Sprintf("Failed to confirm user password, err: %s\n", err))
			return
		} else if !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("No such user or wrong password"))
			return
		}

		// 2. 生成 token
		token := util.GenToken(username)
		_, err = db.UpdateToken(username, token) // ignore bool
		if err != nil {
			internelServerError(w, fmt.Sprintf("Failed to generate token, err: %s\n", err))
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
				Location: "http://" + req.Host + "/static/view/home.html",
				Username: username,
				Token:    token,
			},
		}
		resBytes, err := resp.JSONBytes()
		if err != nil {
			internelServerError(w, fmt.Sprintf("Failed to convert response to bytes, err: %s\n", err))
			return
		}

		w.Write(resBytes)
	}
}

// UserInfoHandler：处理查询用户信息(GET)
func UserInfoHandler(w http.ResponseWriter, req *http.Request) {
	req.ParseForm()
	username := req.Form.Get("username")

	// 查询用户信息
	user, err := db.GetUserInfo(username)
	if err != nil {
		internelServerError(w, fmt.Sprintf("Failed to query user info, err: %s\n", err.Error()))
		return
	}

	// 组装并返回相关信息
	resp := util.RespMsg{
		Code: 200,
		Msg:  "success",
		Data: user,
	}
	resBytes, err := resp.JSONBytes()
	if err != nil {
		internelServerError(w, fmt.Sprintf("Failed to convert response to bytes, err: %s\n", err))
		return
	}

	w.Write(resBytes)
}
