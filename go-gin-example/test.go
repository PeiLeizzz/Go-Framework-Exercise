package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/PeiLeizzz/go-gin-example/pkg/export"
	"github.com/PeiLeizzz/go-gin-example/pkg/file"
	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
)

func main3() {
	setting.Setup()
	file.IsNotExistMkDir(export.GetExcelFullPath())
	file := export.GetExcelFullPath() + "test.csv"
	fmt.Println(file)
	f, err := os.Create(file)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// 标识文件的编码格式（UTF-8）
	f.WriteString("\xEF\xBB\xBF")

	w := csv.NewWriter(f)
	data := [][]string{
		{"1", "test1", "test1-1"},
		{"2", "test2", "test2-1"},
		{"3", "test3", "test3-1"},
	}

	w.WriteAll(data)
}
