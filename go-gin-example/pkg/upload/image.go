package upload

import (
	"fmt"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/PeiLeizzz/go-gin-example/pkg/file"
	"github.com/PeiLeizzz/go-gin-example/pkg/logging"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/PeiLeizzz/go-gin-example/pkg/util"
)

// 图片的网络路径 http://127.0.0.1:8000/upload/images/xxx.png
func GetImageFullUrl(name string) string {
	return setting.AppSetting.ImagePrefixUrl + "/" + GetImagePath() + name
}

// 图片的本地路径 upload/images/
func GetImagePath() string {
	return setting.AppSetting.ImageSavePath
}

// 图片的完整本地路径 runtime/upload/images/
func GetImageFullPath() string {
	return setting.AppSetting.RuntimeRootPath + GetImagePath()
}

// MD5 计算后的图片名称
func GetImageName(name string) string {
	ext := file.GetExt(name)
	fileName := strings.TrimSuffix(name, ext)
	fileName = util.EncodeMD5(fileName)

	return fileName + ext
}

func CheckImageExt(fileName string) bool {
	ext := file.GetExt(fileName)
	for _, allowExt := range setting.AppSetting.ImageAllowExts {
		if strings.EqualFold(allowExt, ext) {
			return true
		}
	}

	return false
}

func CheckImageSize(f multipart.File) bool {
	size, err := file.GetSize(f)
	if err != nil {
		log.Println(err)
		logging.Warn(err)
		return false
	}
	return size <= setting.AppSetting.ImageMaxSize
}

// 检查文件夹权限、创建文件夹等
func CheckImage(src string) error {
	dir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("os.Getwd err: %v", err)
	}

	permDenied := file.CheckPermission(dir + "/" + src)
	if permDenied {
		return fmt.Errorf("file.CheckPermission Permission denied src: %s", src)
	}

	err = file.IsNotExistMkDir(dir + "/" + src)
	if err != nil {
		return fmt.Errorf("file.IsNotExistMkDir err: %v", err)
	}

	return nil
}
