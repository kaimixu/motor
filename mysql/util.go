package mysql

import (
	"reflect"

	"github.com/pkg/errors"
)

// Deprecated
// struct转换成map, target参数的格式为：[]struct
func structToMap(target interface{}) ([]map[string]interface{}, error) {
	if target == nil || reflect.ValueOf(target).IsNil() || reflect.TypeOf(target).Kind() != reflect.Slice {
		return nil, errors.New("target should be a non-empty slice")
	}
	if reflect.TypeOf(target).Elem().Kind() != reflect.Struct {
		return nil, errors.New("members of each target should be a struct")
	}

	elem := reflect.ValueOf(target)
	if reflect.ValueOf(target).Len() <= 0 {
		return nil, errors.New("target slice size cannot be less than zero")
	}

	mapSlice := make([]map[string]interface{}, 0)
	for i := 0; i < elem.Len(); i++ {
		m := make(map[string]interface{})
		structType := elem.Index(i).Type()

		for j := 0; j < structType.NumField(); j++ {
			m[structType.Field(j).Tag.Get("ddb")] = elem.Index(i).Field(j).Interface()
		}
		mapSlice = append(mapSlice, m)
	}

	return mapSlice, nil
}
