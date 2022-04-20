package session

import (
	"errors"
	"geeorm/clause"
	"reflect"
)

// ------------------------ 增删改查部分 ------------------------
// 支持 Hook 机制：Before、After
// 这里的实现是通过方法名来匹配（详见 hooks.go）
// 也可以通过接口来实现
// e.g.
// type IBeforeQuery interface {
//     BeforeQuery(s *Session) error
// }
//
// type IAfterQuery interface {
//     AfterQuery(s *Session) error
// }
//
// ...
//
// hooks.go: CallMethod
// func (s *Session) CallMethod(method interface{}, value interface{}) {
//     ...
//     if m, ok := method.(IBeforeQuery); ok {
//	       value.m(s)
//     }
//	   ...
//	   return
// }

/**
 * INSERT INTO table_name(col1, col2, col3, ...) VALUES
 * 		(A1, A2, A3, ...),
 * 		(B1, B2, B3, ...),
 * 		...
 * => 难点：将对象转换为字段与值
 * s := geeorm.NewEngine("sqlite3", "gee.db").NewSession()
 * u1 := &User{Name: "Tom", Age: 18}
 * u2 := &User{Name: "Sam", Age: 25}
 * ...
 * s.Insert(u1, u2, ...)
 * 如果 hook 函数中的接收者是指针类型，这里就要传引用
 */
func (s *Session) Insert(values ...interface{}) (int64, error) {
	recordValues := make([]interface{}, 0)
	for _, value := range values {
		// 行级 hook
		s.CallMethod(BeforeInsert, value)
		// 下面两句可以简化为只设置一次
		table := s.Model(value).RefTable()
		s.clause.Set(clause.INSERT, table.Name, table.FieldNames)
		// FieldNames 和 Fields 顺序一一对应
		// 注意这里不能 table.RecordValues(value)...
		// 因为 recordValues 应该是 [][]interface
		// 每个元素是一组 value
		recordValues = append(recordValues, table.RecordValues(value))
	}
	s.clause.Set(clause.VALUES, recordValues...)
	sql, vars := s.clause.Build(clause.INSERT, clause.VALUES)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterInsert, nil)
	return result.RowsAffected()
}

/**
 * 难点：从 select 查询结果构造出对象
 * 传入的 values 必须是 slice
 * s := geeorm.NewEngin("sqlite3", "gee.db").NewSession()
 * var users []User
 * s.Find(&users)
 */
func (s *Session) Find(values interface{}) error {
	destSlice := reflect.Indirect(reflect.ValueOf(values))
	destType := destSlice.Type().Elem() // slice 中元素的类型
	table := s.Model(reflect.New(destType).Elem().Interface()).RefTable()
	s.CallMethod(BeforeQuery, nil)

	s.clause.Set(clause.SELECT, table.Name, table.FieldNames)
	// 测试：如果查询结果中的字段顺序和结构体中成员的定义顺序不同，则会错误
	// s.clause.Set(clause.SELECT, table.Name, []string{table.FieldNames[1], table.FieldNames[0]})
	sql, vars := s.clause.Build(clause.SELECT, clause.WHERE, clause.ORDERBY, clause.LIMIT)
	rows, err := s.Raw(sql, vars...).QueryRows()
	if err != nil {
		return err
	}

	for rows.Next() {
		// 每一行记录的实例，可以取地址
		dest := reflect.New(destType).Elem()
		var values []interface{}
		for _, name := range table.FieldNames {
			// 加入指针元素（对应 values 中的引用）
			values = append(values, dest.FieldByName(name).Addr().Interface())
		}
		// 将该行记录每一列的值依次赋值给 values 中的每一个字段
		// values 中存放的是 dest 成员的地址
		// 保存给 values 就代表保存给了 dest 的成员
		// TODO: 是否需要保证表中字段的顺序（select 返回的顺序）和结构体中定义的顺序一致？
		// 如果通过 schema 建的表应该是一致的
		// 测试结果：需要保证，否则会出错，顺序必须一一对应
		if err := rows.Scan(values...); err != nil {
			return err
		}
		// 行级 hook
		s.CallMethod(AfterQuery, dest.Addr().Interface())
		destSlice.Set(reflect.Append(destSlice, dest))
	}
	return rows.Close()
}

/**
 * 要求 s 中已绑定 table
 * 使用：s.Update(kv)
 * support kv : map[string]interface{}
 *  	or kv : []interface{} -> {"Name", "Tom", "Age", 18, ...}
 */
func (s *Session) Update(kv ...interface{}) (int64, error) {
	m, ok := kv[0].(map[string]interface{})
	if !ok {
		m = make(map[string]interface{})
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
	}

	s.CallMethod(BeforeUpdate, nil)
	s.clause.Set(clause.UPDATE, s.RefTable().Name, m)
	sql, vars := s.clause.Build(clause.UPDATE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, nil
	}
	s.CallMethod(AfterUpdate, nil)
	return result.RowsAffected()
}

/**
 * 要求 s 中已绑定 table
 */
func (s *Session) Delete() (int64, error) {
	s.CallMethod(BeforeDelete, nil)
	s.clause.Set(clause.DELETE, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.DELETE, clause.WHERE)
	result, err := s.Raw(sql, vars...).Exec()
	if err != nil {
		return 0, err
	}
	s.CallMethod(AfterDelete, nil)
	return result.RowsAffected()
}

// ------------------------ 快捷调用部分 ------------------------
// 统计满足条件的数据数量、只返回一条查询记录

/**
 * 要求 s 中已绑定 table
 */
func (s *Session) Count() (int64, error) {
	s.clause.Set(clause.COUNT, s.RefTable().Name)
	sql, vars := s.clause.Build(clause.COUNT, clause.WHERE)
	row := s.Raw(sql, vars...).QueryRow()
	var tmp int64
	if err := row.Scan(&tmp); err != nil {
		return 0, err
	}
	return tmp, nil
}

/**
 * 只返回一条查询记录：SELECT + LIMIT(1)
 * e.g.
 *     u := &User{}
 * 	   _ = s.OrderBy("Age DESC").First(u)
 */
func (s *Session) First(value interface{}) error {
	dest := reflect.Indirect(reflect.ValueOf(value))
	destSlice := reflect.New(reflect.SliceOf(dest.Type())).Elem()
	if err := s.Limit(1).Find(destSlice.Addr().Interface()); err != nil {
		return err
	}
	if destSlice.Len() == 0 {
		return errors.New("NOT FOUND")
	}
	dest.Set(destSlice.Index(0))
	return nil
}

// ------------------------ 链式调用部分 ------------------------
// e.g.
//     s := geeorm.NewEngin("sqlite3", "gee.db").NewSession()
//	   var users []User
// 	   s.Where("Age > 18").Limit(3).Find(&users)
// WHERE、LIMIT、ORDER BY 适合于这种模式

func (s *Session) Limit(num int) *Session {
	s.clause.Set(clause.LIMIT, num)
	return s
}

func (s *Session) Where(desc string, args ...interface{}) *Session {
	// var vars []interface{}
	// 将 desc, args 合并传入
	// Change: Set(clause.WHERE, append(append(vars, desc), args...)...))
	s.clause.Set(clause.WHERE, desc, args)
	return s
}

func (s *Session) OrderBy(desc string) *Session {
	s.clause.Set(clause.ORDERBY, desc)
	return s
}
