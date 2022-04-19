/**
 * 实现对象和表的转换
 * 表名 —— 结构体名
 * 字段名和字短类型 —— 成员名称和类型
 * 约束条件（非空、主键等）—— 通过 Tag 实现
 */
package schema

import (
	"geeorm/dialect"
	"go/ast"
	"reflect"
)

// 代表一个字段
type Field struct {
	Name string
	Type string // 数据库中的类型
	Tag  string
}

type Schema struct {
	Model      interface{} // 对象
	Name       string      // 表名
	Fields     []*Field
	FieldNames []string
	fieldMap   map[string]*Field
}

func (schema *Schema) GetField(name string) *Field {
	return schema.fieldMap[name]
}

// 将任意对象解析为 Schema 实例
func Parse(dest interface{}, d dialect.Dialect) *Schema {
	// 获取指针指向的实例，modelType := reflect.ValueOf(dest).Elem().Type() 可行吗？
	modelType := reflect.Indirect(reflect.ValueOf(dest)).Type()
	// modelType := reflect.ValueOf(dest).Elem().Type()
	schema := &Schema{
		Model:    dest,
		Name:     modelType.Name(),
		fieldMap: make(map[string]*Field),
	}

	for i := 0; i < modelType.NumField(); i++ {
		p := modelType.Field(i)
		// 不是内嵌结构体、并且是导出的成员
		if !p.Anonymous && ast.IsExported(p.Name) {
			field := &Field{
				Name: p.Name,
				// 不可以用 d.DataTypeOf(reflect.ValueOf(p.Type)) / d.DataTypeOf(reflect.ValueOf(p))
				// 可以用 Type: d.DataTypeOf(reflect.ValueOf(dest).Elem().Field(i)),
				Type: d.DataTypeOf(reflect.Indirect(reflect.New(p.Type))),
				// Type: d.DataTypeOf(reflect.ValueOf(dest).Elem().Field(i)),
			}
			if v, ok := p.Tag.Lookup("geeorm"); ok {
				field.Tag = v
			}
			schema.Fields = append(schema.Fields, field)
			schema.FieldNames = append(schema.FieldNames, p.Name)
			schema.fieldMap[p.Name] = field
		}
	}
	return schema
}

func (schema *Schema) RecordValues(dest interface{}) []interface{} {
	destValue := reflect.Indirect(reflect.ValueOf(dest))
	var fieldValues []interface{}
	for _, field := range schema.Fields {
		fieldValues = append(fieldValues, destValue.FieldByName(field.Name).Interface())
	}
	return fieldValues
}
