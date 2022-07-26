package util

import (
	"encoding/json"
	"fmt"
)

type RespMsg struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func NewRespMsg(code int, msg string, data interface{}) *RespMsg {
	return &RespMsg{
		Code: code,
		Msg:  msg,
		Data: data,
	}
}

// JSONBytes: 对象转 json 格式的二进制数组
func (resp *RespMsg) JSONBytes() ([]byte, error) {
	r, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// JSONBytes: 对象转 json 格式的 string
func (resp *RespMsg) JSONString() (string, error) {
	r, err := resp.JSONBytes()
	return string(r), err
}

// GenSimpleRespStream: 不包含 data 的 json 格式二进制数组响应
func GenSimpleRespStream(code int, msg string) []byte {
	return []byte(fmt.Sprintf(`{"code":%d,"msg":"%s"`, code, msg))
}

// GenSimpleRespStream: 不包含 data 的 json 格式 string 响应
func GenSimpleRespString(code int, msg string) string {
	return fmt.Sprintf(`{"code":%d,"msg":"%s"`, code, msg)
}
