package geeorm

import (
	"database/sql"
	"fmt"
	"strings"

	"geeorm/dialect"
	"geeorm/log"
	"geeorm/session"
)

type Engine struct {
	db      *sql.DB
	dialect dialect.Dialect
}

func NewEngine(driver, source string) (e *Engine, err error) {
	db, err := sql.Open(driver, source)
	if err != nil {
		log.Error(err)
		return
	}
	// 确保连接还在保持
	if err = db.Ping(); err != nil {
		log.Error(err)
		return
	}

	dial, ok := dialect.GetDialect(driver)
	if !ok {
		log.Errorf("dialect %s Not Found", driver)
		return
	}

	e = &Engine{db: db, dialect: dial}
	log.Info("Connect database success")
	return
}

func (engine *Engine) Close() {
	if err := engine.db.Close(); err != nil {
		log.Error("failed to close database")
	}
	log.Info("Close database success")
}

func (engine *Engine) NewSession() *session.Session {
	return session.New(engine.db, engine.dialect)
}

type TxFunc func(*session.Session) (interface{}, error)

// 将所有操作放入一个回调函数中，如果出错自动回滚
func (engine *Engine) Transaction(f TxFunc) (result interface{}, err error) {
	s := engine.NewSession()
	if err := s.Begin(); err != nil {
		return nil, err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = s.Rollback()
			panic(p)
		} else if err != nil {
			_ = s.Rollback()
		} else {
			// 提交失败的话还需要回滚
			defer func() {
				if err != nil {
					_ = s.Rollback()
				}
			}()
			err = s.Commit()
		}
	}()

	return f(s)
}

// 找出 a 表（结构体）存在的、但 b 表（结构体）不存在的字段
func difference(a []string, b []string) (diff []string) {
	mapB := make(map[string]bool)
	for _, v := range b {
		mapB[v] = true
	}
	for _, v := range a {
		if _, ok := mapB[v]; !ok {
			diff = append(diff, v)
		}
	}
	return
}

/**
 * 实现结构体变更时，数据库表的字段自动迁移
 * 针对最简单的场景：仅支持字段新增和删除，不支持字段类型的变更
 * 并且不涉及约束条件
 *
 * SQLite3 原生新增字段
 *     ALTER TABLE table_name ADD COLUMN col_name, col_type;
 * SQLite3 原生删除字段（三步）
 *  1. CREATE TABLE new_table AS SELECT col1, col2, ... FROM old_table;
 *  2. DROP TABLE old_table;
 *  3. ALTER TABLE new_table RENAME TO old_table;
 * 必须先建新表、删除原表、更名新表才能完成表字段的删除，因此需要通过事务来进行
 */
func (engine *Engine) Migrate(value interface{}) error {
	_, err := engine.Transaction(func(s *session.Session) (result interface{}, err error) {
		// 查看需要修改的同名表是否存在
		if !s.Model(value).HasTable() {
			log.Infof("table %s doesn't exist", s.RefTable().Name)
			return nil, s.CreateTable()
		}
		table := s.RefTable()
		// 这里不使用 s.First() / s.Find() 是因为要通过 rows 获取原表各字段名称
		// 另外注意在 MySQL 中，rows 会获得数据库的链接，之后对数据库的操作会无法执行(Exec)
		// 要先 rows.Close()，否则 tx 无法再从连接池获取当前连接
		// (一条 transaction 里面的所有操作都是同步的，MySQL 为了保证事务的顺序执行，连接池里面只有一个连接)
		rows, _ := s.Raw(fmt.Sprintf("SELECT * FROM %s LIMIT 1", table.Name)).QueryRows()
		columns, _ := rows.Columns() // 原表各字段
		addCols := difference(table.FieldNames, columns)
		delCols := difference(columns, table.FieldNames)
		log.Infof("added cols %v, deleted cols %v", addCols, delCols)

		for _, col := range addCols {
			f := table.GetField(col)
			sqlStr := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s;", table.Name, f.Name, f.Type)
			if _, err = s.Raw(sqlStr).Exec(); err != nil {
				return
			}
		}

		if len(delCols) == 0 {
			return
		}

		tmp := "tmp_" + table.Name
		fieldStr := strings.Join(table.FieldNames, ", ")
		// Go 操作 MySQL 默认不支持多条语句运行，需要在数据库的 URL 中添加 multiStatements = true 的选项
		s.Raw(fmt.Sprintf("CREATE TABLE %s AS SELECT %s FROM %s;", tmp, fieldStr, table.Name))
		s.Raw(fmt.Sprintf("DROP TABLE %s;", table.Name))
		s.Raw(fmt.Sprintf("ALTER TABLE %s RENAME TO %s;", tmp, table.Name))
		_, err = s.Exec()
		return
	})
	return err
}
