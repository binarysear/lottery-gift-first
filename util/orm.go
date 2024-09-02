package util

import (
	"reflect"
	"strings"
)

// 通过反射获取数据库的字段
func GetGormFields(stc interface{}) []string {
	value := reflect.ValueOf(stc)
	typ := value.Type()

	// 如果是指针类型 就解引用
	if typ.Kind() == reflect.Ptr {
		if value.IsNil() {
			return nil
		} else {
			typ = typ.Elem()
			value = value.Elem()
		}
	}

	// 如果输入的不是一个结构体，则返回nil
	if typ.Kind() != reflect.Struct {
		return nil
	}

	//遍历结构体的字段
	columns := make([]string, 0, value.NumField())
	for i := 0; i < value.NumField(); i++ {
		fieldType := typ.Field(i)

		// 只考虑导出的字段
		if !fieldType.IsExported() {
			continue
		}

		// 跳过带有"gorm"标签设置为"-"的字段  "-"表示不查找
		if fieldType.Tag.Get("gorm") == "-" {
			continue
		}

		// 获取字段的snake格式的名称
		name := Camel2Snake(fieldType.Name)

		// 如果字段带有“gorm”标签的"column"选项，使用该选项代替字段名称
		if len(fieldType.Tag.Get("gorm")) > 0 {
			content := fieldType.Tag.Get("gorm")
			if strings.HasPrefix(content, "column:") {
				content := content[7:]
				pos := strings.Index(content, ";")
				if pos > 0 {
					name = content[0:pos]
				} else if pos < 0 {
					name = content
				}
			}
		}

		// 将该字段名称添加到列名列表中
		columns = append(columns, name)
	}

	return columns
}
