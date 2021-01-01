package quacktors

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"sync"
	"testing"
	"time"
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

	<-time.After(50 * time.Millisecond)

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

	<-time.After(50 * time.Millisecond)

	rootCtx.Send(p, &PoisonPill{})

	Wait()
}

func TestMonitorAbortable_Abort(t *testing.T) {
	rootCtx := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case *GenericMessage:
			fmt.Println(m.Value)
		}
	})

	var a Abortable

	q := SpawnWithInit(func(ctx *Context) {
		a = ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch message.(type) {
		case *DownMessage:
			fmt.Println(":(")
			t.Fail()
		case *GenericMessage:
			fmt.Println("Worked")
			ctx.Quit()
		}
	})

	a.Abort()

	<-time.After(50 * time.Millisecond)

	rootCtx.Send(p, &PoisonPill{})

	rootCtx.Send(q, &GenericMessage{Value: ""})

	Wait()
}

func TestMonitorDeadPid(t *testing.T) {
	rootCtx := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
	})

	rootCtx.Kill(p)

	<-time.After(50 * time.Millisecond)

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch message.(type) {
		case *DownMessage:
			ctx.Quit()
		}
	})

	Wait()
}

type TestMessage struct {
	Foo string
}

func (t TestMessage) Type() string {
	return "TestMessage"
}

func TestTypeRegistration(t *testing.T) {
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

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

	p, err := r.Remote("printer")

	if err != nil {
		t.Fail()
	}

	rootCtx.Send(p, &GenericMessage{Value: "Hello!"})
}

func TestConnect2(t *testing.T) {
	rootCtx := RootContext()

	r, err := Connect("test@localhost")

	if err != nil {
		t.Fail()
	}

	wg := sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()

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
