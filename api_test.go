package quacktors

import (
	"encoding/json"
	"fmt"
	"runtime"
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

func TestSpawn(t *testing.T) {
	RegisterType(&TestMessage{})

	s, err := NewSystem("test")

	if err != nil {
		panic(err)
	}

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case TestMessage:
			fmt.Printf("GOOS: %s", m.Foo)
		default:
			fmt.Println("Unrecognized type!")
		}
	})

	s.HandleRemote("printer", p)
	Send(p, TestMessage{Foo: runtime.GOOS})
}

type TestActor struct {
	count int
}

func (t *TestActor) Run(ctx *Context, message Message)  {
	for {
		t.count++
	}
}

func TestActorSpawn(t *testing.T) {
	actor := &TestActor{}

	SpawnStateful(actor)
}
