package quacktors

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
)

func TestMonitorWithKill(t *testing.T) {
	rootCtx := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case *GenericMessage:
			fmt.Println(m.Value)
		}
	})

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch m := message.(type) {
		case *DownMessage:
			assert.Equal(t, p.String(), m.Who.String())
			fmt.Println("Actor went down!")
			ctx.Quit()
		}
	})

	rootCtx.Kill(p)

	Wait()
}

func TestMonitorWithPoisonPill(t *testing.T) {
	rootCtx := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case *GenericMessage:
			fmt.Println(m.Value)
		}
	})

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch m := message.(type) {
		case *DownMessage:
			assert.Equal(t, p.String(), m.Who.String())
			fmt.Println("Actor went down!")
			ctx.Quit()
		}
	})

	rootCtx.Send(p, &PoisonPill{})

	Wait()
}

type TestMessage struct {
	Foo string
}

func (t TestMessage) Type() string {
	return "TestMessage"
}

func TestTypeRegistration(t *testing.T) {
	//qpmdLookup("foo", "127.0.0.1", 0)

	RegisterType(&TestMessage{Foo: MachineId()})
	v := getType(TestMessage{}.Type())

	assert.Empty(t, v.(TestMessage).Foo)
}

func TestNewSystem(t *testing.T) {
	_, err := NewSystem("test")

	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func TestNewSystemWithHandler(t *testing.T) {
	s, err := NewSystem("test")

	if err != nil {
		panic(err)
	}

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case *GenericMessage:
			fmt.Println(m.Value)
			ctx.Quit()
		}
	})

	s.HandleRemote("printer", p)

	Wait()
}

func TestConnect(t *testing.T) {
	rootCtx := RootContext()

	r, err := Connect("test@localhost")

	if err != nil {
		t.Fail()
	}

	p, err := r.Remote("printer")

	if err != nil {
		t.Fail()
	}

	rootCtx.Send(p, &GenericMessage{Value: "Hello!"})
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
