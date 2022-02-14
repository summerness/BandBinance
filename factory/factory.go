package factory

import (
	"errors"
	"reflect"
)

var (
	ServiceMap = make(map[string]*interface{}, 64)
)

func Put(name string, value *interface{}) error {
	ServiceMap[name] = value
	return nil
}

// GetService 反射使用
func GetService(name string) (interface{}, error) {
	value := ServiceMap[name]
	if value == nil {
		return nil, errors.New("没有这个service")
	}
	return reflect.ValueOf(value).Interface(), nil
}
