package encrypt

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"errors"
)

// 16/24/32 位 -> AES-128/AES-192/AES-256
var PwdKey = []byte("1xfa342cvxzaewAr")

// 填充字符串
func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padText := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padText...)
}

// AES 加密
func AesEcrypt(origData []byte, key []byte) ([]byte, error) {
	// 创建加密算法实例
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// 获取块的大小
	blockSize := block.BlockSize()

	// 对加密数据进行填充
	origData = PKCS7Padding(origData, blockSize)

	// 进行 AES 加密（CBC 加密模式）
	blockMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)

	return crypted, nil
}

// base64 编码
func EnPwdCode(pwd []byte) (string, error) {
	result, err := AesEcrypt(pwd, PwdKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(result), nil
}

// 删除填充字符串
func PKSC7UnPadding(origData []byte) ([]byte, error) {
	length := len(origData)
	if length == 0 {
		return nil, errors.New("加密字符串错误")
	}

	// 在填充时我们把填充长度附在了最后
	// TODO: bug: len(userid) % blocksize 时，不会附长度
	unpadding := int(origData[length-1])
	// 截取切片，删除填充字节
	return origData[:(length - unpadding)], nil
}

// AES 解密
func AesDecrypt(cypted []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockSize := block.BlockSize()
	blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	origData := make([]byte, len(cypted))
	blockMode.CryptBlocks(origData, cypted)

	origData, err = PKSC7UnPadding(origData)
	if err != nil {
		return nil, err
	}
	return origData, nil
}

// base64 反编码
func DePwdCode(pwd string) ([]byte, error) {
	pwdByte, err := base64.StdEncoding.DecodeString(pwd)
	if err != nil {
		return nil, err
	}
	return AesDecrypt(pwdByte, PwdKey)
}
