package mysql

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "root:peilei777@tcp(127.0.0.1:3307)/fileserver?charset=utf8")
	if err != nil {
		log.Fatal("Failed to open mysql, err: %s", err.Error())
	}

	db.SetMaxOpenConns(1000)
	err = db.Ping()
	if err != nil {
		log.Fatal("Failed to connect to mysql, err: %s", err.Error())
	}
	log.Println("mysql connected success")
}

func DBConn() *sql.DB {
	return db
}

func ParseRows(rows *sql.Rows) ([]map[string]interface{}, error) {
	columns, _ := rows.Columns()
	scanArgs := make([]interface{}, len(columns))
	values := make([]interface{}, len(columns))

	for j := range values {
		scanArgs[j] = &values[j]
	}

	record := make(map[string]interface{})
	records := make([]map[string]interface{}, 0)
	for rows.Next() {
		// 将行数据保存到 record 字典中
		err := rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		for i, col := range values {
			if col != nil {
				record[columns[i]] = col
			}
		}

		records = append(records, record)
	}
	return records, nil
}
