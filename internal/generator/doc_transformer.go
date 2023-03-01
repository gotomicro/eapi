package generator

import (
	"fmt"
	"reflect"
	"strings"
)

func transformValue(v reflect.Value) interface{} {
	switch v.Kind() {
	case reflect.Struct:
		return transformStruct(v)
	case reflect.Pointer:
		if v.IsNil() {
			return nil
		}
		return transformValue(v.Elem())
	case reflect.Map:
		return transformMap(v)
	case reflect.Slice:
		return transformSlice(v)
	case reflect.Array:
		return transformSlice(v)
	}

	if v.CanInterface() {
		return v.Interface()
	}

	fmt.Printf("unsupported type %v", v)
	panic("unexpected type")
}

func transformSlice(v reflect.Value) interface{} {
	var res = make([]interface{}, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		res = append(res, transformValue(v.Index(i)))
	}
	return res
}

func transformMap(v reflect.Value) interface{} {
	var res = make(map[interface{}]interface{})
	for _, key := range v.MapKeys() {
		res[transformValue(key)] = transformValue(reflect.ValueOf(v.MapIndex(key).Interface()))
	}
	return res
}

func transformStruct(v reflect.Value) interface{} {
	var res = make(map[string]interface{})
	fieldNums := v.NumField()
	for i := 0; i < fieldNums; i++ {
		field := v.Field(i)
		name := fieldName(v.Type().Field(i))
		if name == "-" {
			continue
		}
		res[name] = transformValue(field)
	}
	return res
}

func fieldName(field reflect.StructField) string {
	tagName, _, _ := strings.Cut(field.Tag.Get("json"), ",")
	if tagName != "" {
		return tagName
	}
	return field.Name
}
