package encoding

import (
	"errors"
	"github.com/vmihailenco/msgpack/v5"
	"reflect"
	"sync"
)

type MsgpackEncoder struct {
	registry *sync.Map
}

func NewMsgpackEncoder() *MsgpackEncoder {
	return &MsgpackEncoder{
		registry: &sync.Map{},
	}
}

func (mpe *MsgpackEncoder) Encode(messageType string, value interface{}) ([]byte, error) {
	m := make(map[string]interface{})
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
			v, err := mpe.Encode(messageType, field.Interface())
			if err != nil {
				return nil, err
			}
			m[valType.Field(i).Name] = v
		default:
			m[valType.Field(i).Name] = field.Interface()
		}
	}

	b, err := msgpack.Marshal(m)
	if b == nil {
		b = make([]byte, 0)
	}

	return b, err
}

func (mpe *MsgpackEncoder) Decode(messageType string, buffer []byte) (interface{}, error) {
	m := make(map[string]interface{})
	err := msgpack.Unmarshal(buffer, &m)

	registryVal, ok := mpe.registry.Load(messageType)

	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("no such type " + messageType)
	}

	retType := reflect.ValueOf(registryVal).Type()
	ret := reflect.New(retType).Elem()

	for s, i := range m {
		ret.FieldByName(s).Set(reflect.ValueOf(i))
	}

	return ret.Interface(), nil
}

func (mpe *MsgpackEncoder) RegisterType(messageType string, value interface{}) {
	mpe.registry.Store(messageType, value)
}
