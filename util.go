package quacktors

import (
	uuid "github.com/satori/go.uuid"
	"reflect"
	"strings"
)

func try(err error) {
	if err != nil {
		panic(err)
	}
}

func uuidString() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

func createFromTemplateMessage(from Message) Message {
	t := reflect.ValueOf(from).Elem()
	typ := t.Type()
	ms := (reflect.New(typ).Elem()).Interface().(Message)

	return ms
}
