package quacktors

import (
	"encoding/json"
	"testing"
)

type TestMessage struct {
	Foo string `json:"foo"`
}

func (t TestMessage) Serialize() string {
	b, err := json.Marshal(t)

	if err != nil {
		panic(err)
	}

	return string(b)
}

func (t TestMessage) Deserialize(s string) Message {
	m := TestMessage{}
	_ = json.Unmarshal([]byte(s), &m)
	return m
}

func (t TestMessage) Type() string {
	return "TestMessage"
}

func TestConnect(t *testing.T) {

}

func TestSpawn(t *testing.T) {
	RegisterType(&TestMessage{})

	s, err := NewSystem("test")

	if err != nil {
		panic(err)
	}

	self := DebugPid()

	p := Spawn(func(ctx *Context) {
		if v, ok := ctx.Receive().(TestMessage); ok {
			v.Type()
		}
		Send(self, TestMessage{Foo: "hello"})
		ctx.Self()
		ctx.Children()
	})

	s.HandleRemote("printer", p)
}
