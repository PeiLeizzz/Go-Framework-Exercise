package export

import "github.com/PeiLeizzz/go-gin-example/pkg/setting"

// 网络路径：http://127.0.0.1:8000/export/name
func GetExcelFullUrl(name string) string {
	return setting.AppSetting.PrefixUrl + "/" + GetExportPath() + name
}

// 本地完整路径：runtime/export/
func GetExcelFullPath() string {
	return setting.AppSetting.RuntimeRootPath + GetExportPath()
}

// 本地路径：export/
func GetExportPath() string {
	return setting.AppSetting.ExportSavePath
}
