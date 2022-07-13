package models

import (
	"fmt"
	"log"
	"time"

	"github.com/PeiLeizzz/go-gin-example/pkg/setting"
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB

// gorm 默认列名映射为小写下划线命名法（CreatedOn -> created_on）
type Model struct {
	ID         int `gorm:"primary_key" json:"id"`
	CreatedOn  int `json:"created_on"`
	ModifiedOn int `json:"modified_on"`
	DeletedOn  int `json:"deleted_on"`
}

func Setup() {
	dbURL := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		setting.DatabaseSetting.User,
		setting.DatabaseSetting.Password,
		setting.DatabaseSetting.Host,
		setting.DatabaseSetting.Name)
	// 坑：千万不能不定义 err 就 db, err := gorm...... 这样会导致 db 是局部变量，外面的 db 没有被更新！
	var err error
	db, err = gorm.Open(setting.DatabaseSetting.Type, dbURL)

	if err != nil {
		log.Println(err)
	}

	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return setting.DatabaseSetting.TablePrefix + defaultTableName
	}

	db.SingularTable(true) // 配置默认数据库表名 = 结构体名的单数（User -> user）
	db.LogMode(true)
	db.DB().SetMaxIdleConns(10)  // 允许的最大空闲链接数
	db.DB().SetMaxOpenConns(100) // 允许的最大链接数
	db.Callback().Create().Replace("gorm:update_time_stamp", updateTimeStampForCreateCallback)
	db.Callback().Update().Replace("gorm:update_time_stamp", updateTimeStampForUpdateCallback)
	db.Callback().Delete().Replace("gorm:delete", deleteCallback)
}

func CloseDB() {
	defer db.Close()
}

// create 时设置 CreateOn 和 ModifiedOn 两个字段
func updateTimeStampForCreateCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		nowTime := time.Now().Unix()
		if createTimeField, ok := scope.FieldByName("CreatedOn"); ok && createTimeField.IsBlank {
			createTimeField.Set(nowTime)
		}

		if modifyTimeField, ok := scope.FieldByName("ModifiedOn"); ok && modifyTimeField.IsBlank {
			modifyTimeField.Set(nowTime)
		}
	}
}

func updateTimeStampForUpdateCallback(scope *gorm.Scope) {
	// 查找 scope 中含这个字面值的字段属性，如果不存在，则默认更新时间
	if _, ok := scope.Get("gorm:update_column"); !ok {
		scope.SetColumn("ModifiedOn", time.Now().Unix())
	}
}

func deleteCallback(scope *gorm.Scope) {
	if !scope.HasError() {
		extraOption := ""
		// 检查是否手动指定了 delete_option
		if str, ok := scope.Get("gorm:delete_option"); ok {
			extraOption = fmt.Sprint(str)
		}

		deletedOnField, hasDeletedOnField := scope.FieldByName("DeletedOn")
		// 软删除
		if !scope.Search.Unscoped && hasDeletedOnField {
			scope.Raw(fmt.Sprintf(
				"UPDATE %v SET %v=%v%v%v",
				scope.QuotedTableName(),                            // e.g. `blog_tag`
				scope.Quote(deletedOnField.DBName),                 // e.g. `blog_tag`.`deleted_on`
				scope.AddToVars(time.Now().Unix()),                 // AddToVars 可以防 sql 注入
				addExtraSpaceIfExist(scope.CombinedConditionSql()), // e.g. join on ... where ... and ...
				addExtraSpaceIfExist(extraOption),                  // other options
			)).Exec()
			// 硬删除
		} else {
			scope.Raw(fmt.Sprintf(
				"DELETE FROM %v%v%v",
				scope.QuotedTableName(),
				addExtraSpaceIfExist(scope.CombinedConditionSql()),
				addExtraSpaceIfExist(extraOption),
			)).Exec()
		}
	}
}

func addExtraSpaceIfExist(str string) string {
	if str != "" {
		return " " + str
	}
	return ""
}
