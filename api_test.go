package quacktors

import (
	"fmt"
	"runtime"
	"testing"
)

type TestMessage struct {
	Foo string
}

func (t TestMessage) Type() string {
	return "TestMessage"
}

func TestSpawn(t *testing.T) {
	//qpmdLookup("foo", "127.0.0.1", 0)

	RegisterType(&TestMessage{})
	rootCtx := RootContext()

	s, err := NewSystem("test")

	if err != nil {
		panic(err)
	}

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case TestMessage:
			fmt.Printf("GOOS: %s", m.Foo)
			ctx.Quit()
		default:
			fmt.Println("Unrecognized type!")
		}
	})

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		fmt.Println("Ded")
	})

	s.HandleRemote("printer", p)
	rootCtx.Send(p, TestMessage{Foo: runtime.GOOS})
}

type TestActor struct {
	count int
}

func (t *TestActor) Run(ctx *Context, message Message) {
	for {
		t.count++
	}
}

func (t *TestActor) Init(ctx *Context) {

}

func TestActorSpawn(t *testing.T) {
	actor := &TestActor{}

	SpawnStateful(actor)
}
