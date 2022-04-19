package clause

import (
	"fmt"
	"strings"
)

type generator func(values ...interface{}) (string, []interface{})

var generators map[Type]generator

func init() {
	generators = make(map[Type]generator)
	generators[INSERT] = _insert
	generators[VALUES] = _values
	generators[SELECT] = _select
	generators[LIMIT] = _limit
	generators[WHERE] = _where
	generators[ORDERBY] = _orderBy
	generators[UPDATE] = _update
	generators[DELETE] = _delete
	generators[COUNT] = _count
}

func genBindVars(num int) string {
	var vars []string
	for i := 0; i < num; i++ {
		vars = append(vars, "?")
	}
	return strings.Join(vars, ", ")
}

/**
 * INSERT INTO $tableName ($fields)
 * (tableName string, fields []string)
 */
func _insert(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("INSERT INTO %s (%v)", tableName, fields), []interface{}{}
}

/**
 * VALUES ($v1), ($v2), ...
 * VALUES (?, ?, ?, ?), (?, ?, ?, ?)
 * (value1 []interface{}, value2 []interface{}, ...)
 * let values [][]interface{}
 */
func _values(values ...interface{}) (string, []interface{}) {
	var bindStr string
	var sql strings.Builder
	var vars []interface{}
	sql.WriteString("VALUES ")
	for i, value := range values {
		// values 是 [][]interface{}
		v := value.([]interface{})
		if bindStr == "" {
			bindStr = genBindVars(len(v))
		}
		sql.WriteString(fmt.Sprintf("(%v)", bindStr))
		if i+1 != len(values) {
			sql.WriteString(", ")
		}
		vars = append(vars, v...)
	}
	return sql.String(), vars
}

/**
 * SELECT $fields FROM $tableName
 * (tableName string, fields []string)
 */
func _select(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	fields := strings.Join(values[1].([]string), ",")
	return fmt.Sprintf("SELECT %v FROM %s", fields, tableName), []interface{}{}
}

/**
 * LIMIT $num
 * num int
 */
func _limit(values ...interface{}) (string, []interface{}) {
	return "LIMIT ?", values
}

/**
 * WHERE $desc
 * (desc string, var1 interface{}, var2 interface{}, ...)
 * let vars = values[1:] (=> []interface{})
 * e.g. desc: "Name = ? and Age = ?", vars: []interface{"Tom", 20}
 */
func _where(values ...interface{}) (string, []interface{}) {
	// desc 中的条件语句类似于 "Name = ?", "Tom"
	// Change: vars := values[1:]
	desc, vars := values[0], values[1]
	return fmt.Sprintf("WHERE %s", desc), vars.([]interface{})
}

/**
 * ORDER BY $field
 * field string
 */
func _orderBy(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("ORDER BY %s", values[0]), []interface{}{}
}

/**
 * UPDATE $tableName SET $field1 = ?, $field2 = ?, ...
 * (tableName string, fields map[string]interface{})
 */
func _update(values ...interface{}) (string, []interface{}) {
	tableName := values[0]
	m := values[1].(map[string]interface{})
	var keys []string
	var vars []interface{}
	for k, v := range m {
		keys = append(keys, k+" = ?")
		vars = append(vars, v)
	}
	return fmt.Sprintf("UPDATE %s SET %s", tableName, strings.Join(keys, ", ")), vars
}

/**
 * DELETE FROM $tableName
 * tableName string
 */
func _delete(values ...interface{}) (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s", values[0]), []interface{}{}
}

/**
 * SELECT COUNT(*) FROM $tableName
 * tableName string
 */
func _count(values ...interface{}) (string, []interface{}) {
	return _select(values[0], []string{"count(*)"})
}
