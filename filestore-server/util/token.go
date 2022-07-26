package util

import (
	"fmt"
	"time"
)

const (
	token_salt = "_tokensalt"
)

// getToken：得到 40 位的 token
func GenToken(username string) string {
	// md5(username + timestamp + token_salt) + timestamp[:8]
	timestamp := fmt.Sprintf("%x", time.Now().Unix())
	tokenPrefix := MD5([]byte(username + timestamp + token_salt))
	return tokenPrefix + timestamp[:8]
}

func IsTokenValid(username string, token string) bool {
	if len(token) != 40 {
		return false
	}
	// TODO: 判断 token 的时效性，是否过期
	// TODO: 从数据库表 tbl_user_token 查询 username 对应的 token 信息
	// TODO: 对比两个 token 是否一致
	return true
}
