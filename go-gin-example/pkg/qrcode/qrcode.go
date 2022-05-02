package qrcode

import (
	"image/jpeg"

	"github.com/PeiLeizzz/go-gin-example/pkg/file"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

type QrCode struct {
	URL    string
	Width  int
	Height int
	Ext    string
	Level  qr.ErrorCorrectionLevel
	Mode   qr.Encoding
}

const (
	EXT_JPG = ".jpg"
)

func NewQrCode(url string, width, height int, level qr.ErrorCorrectionLevel, mode qr.Encoding) *QrCode {
	return &QrCode{
		URL:    url,
		Width:  width,
		Height: height,
		Level:  level,
		Mode:   mode,
		Ext:    EXT_JPG,
	}
}

// http://127.0.0.1:8000/qrcode/name
func GetQrCodeFullUrl(name string) string {
	return setting.AppSetting.PrefixUrl + "/" + GetQrCodePath() + name
}

// runtime/qrcode/
func GetQrCodeFullPath() string {
	return setting.AppSetting.RuntimeRootPath + GetQrCodePath()
}

// qrcode/
func GetQrCodePath() string {
	return setting.AppSetting.QrCodeSavePath
}

// MD5 计算后的文件名称（不包含扩展名）
func GetQrCodeFileName(value string) string {
	return util.EncodeMD5(value)
}

func (q *QrCode) GetQrCodeExt() string {
	return q.Ext
}

func (q *QrCode) CheckEncode(path string) bool {
	src := path + GetQrCodeFileName(q.URL) + q.GetQrCodeExt()
	return !file.CheckNotExist(src)
}

func (q *QrCode) Encode(path string) (string, error) {
	name := GetQrCodeFileName(q.URL) + q.GetQrCodeExt()
	// 检查图片是否已经存在
	if !q.CheckEncode(path) {
		// 创建二维码
		code, err := qr.Encode(q.URL, q.Level, q.Mode)
		if err != nil {
			return "", err
		}

		// 缩放二维码
		code, err = barcode.Scale(code, q.Width, q.Height)
		if err != nil {
			return "", err
		}

		// 新建存放二维码图片的文件
		f, err := file.MustOpen(name, path)
		if err != nil {
			return "", err
		}
		defer f.Close()

		// 将图像（二维码）以 JPEG 4：2：0 基线格式写入文件
		err = jpeg.Encode(f, code, nil)
		if err != nil {
			file.Delete(name, path) // ignore err
			return "", err
		}
	}

	return name, nil
}
