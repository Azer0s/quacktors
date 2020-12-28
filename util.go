package quacktors

import "reflect"

func try(err error) {
	if err != nil {
		panic(err)
	}
}

func create(from Message) Message {
	t := reflect.ValueOf(from).Elem()
	typ := t.Type()
	ms := (reflect.New(typ).Elem()).Interface().(Message)

	return ms
}