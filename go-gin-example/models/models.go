package models

import (
	"fmt"
	"log"

	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

type Model struct {
	ID          int `gorm:"primary_key" json:"id"`
	CreatedOn   int `json:"created_on"`
	ModeifiedOn int `json:"modified_on"`
}

func init() {
	var (
		err error

		dbType,
		dbName,
		user,
		password,
		host,
		tablePrefix string
	)

	sec, err := setting.Cfg.GetSection("database")
	if err != nil {
		log.Fatal(2, "Fail to get section 'database': %v", err)
	}

	dbType = sec.Key("TYPE").String()
	dbName = sec.Key("NAME").String()
	user = sec.Key("USER").String()
	password = sec.Key("PASSWORD").String()
	host = sec.Key("HOST").String()
	tablePrefix = sec.Key("TABLE_PREFIX").String()

	dbURL := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset-utf8mb4&parseTime=True&loc=Local",
		user,
		password,
		host,
		dbName)
	db, err = gorm.Open(dbType, dbURL)

	if err != nil {
		log.Println(err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return tablePrefix + defaultTableName
	}

	db.SingularTable(true)
	db.LogMode(true)
	db.DB().SetMaxIdleConns(10)  // 允许的最大空闲链接数
	db.DB().SetMaxOpenConns(100) // 允许的最大链接数
}

func CloseDB() {
	defer db.Close()
}
