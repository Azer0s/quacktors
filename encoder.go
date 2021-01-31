package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/quacktors/typeregister"
	"reflect"
)

func encodeValue(messageType string, value interface{}) (ret map[string]interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic while encoding value %x", r)
		}
	}()

	ret = make(map[string]interface{})
	val := reflect.ValueOf(value)

	//This should actually never happen
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		switch field.Kind() {
		case reflect.Struct:
			fallthrough
		case reflect.Ptr:
			v, err := encodeValue(messageType, field.Interface())
			if err != nil {
				return nil, err
			}
			ret[valType.Field(i).Name] = v
		default:
			ret[valType.Field(i).Name] = field.Interface()
		}
	}

	return ret, nil
}

func decodeValue(messageType string, data map[string]interface{}) (interface{}, error) {
	registryVal, ok := typeregister.Load(messageType)

	if !ok {
		return nil, errors.New("no such type " + messageType)
	}

	return decodeValueByInterface(registryVal, data)
}

func decodeValueByInterface(template interface{}, data map[string]interface{}) (interface{}, error) {
	retType := reflect.ValueOf(template).Type()
	ret := reflect.New(retType).Elem()

	for s, i := range data {
		val := reflect.ValueOf(i)

		switch ret.FieldByName(s).Kind() {
		case reflect.Struct:
			b, ok := i.(map[string]interface{})
			if !ok {
				return nil, errors.New(s + " is " + reflect.TypeOf(i).String() + " not map[string]interface{}")
			}

			v, err := decodeValueByInterface(reflect.New(ret.FieldByName(s).Type()).Elem().Interface(), b)

			if err != nil {
				return nil, err
			}

			ret.FieldByName(s).Set(reflect.ValueOf(v))
		default:
			ret.FieldByName(s).Set(val)
		}
	}

	return ret.Interface(), nil
}
