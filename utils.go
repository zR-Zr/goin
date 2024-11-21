package goin

import (
	"reflect"
	"strings"
)

func getTagName(obj any, fieldName string) string {
	// 获取对象的反射类型
	rType := reflect.TypeOf(obj)

	// 如果是指针类型,先获取指向的实际类型
	if rType.Kind() == reflect.Ptr {
		rType = rType.Elem()
	}

	// 便利对象的所有字段
	for i := 0; i < rType.NumField(); i++ {
		field := rType.Field(i)

		// 如果字段名匹配, 则获取 Json标签
		if field.Name == fieldName {
			tag := field.Tag.Get("json")
			if tag == "" {
				tag = field.Tag.Get("uri")
			}

			if tag == "" {
				tag = field.Tag.Get("form")
			}

			if tag != "" && tag != "-" {
				return strings.Split(tag, ",")[0]
			}
		}
	}

	// 没有找到,返回空字符串
	return ""
}
