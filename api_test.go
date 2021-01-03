package quacktors

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMessageOrdering(t *testing.T) {
	rootCtx := RootContext()

	testChan := make(chan string, 40000)

	i := 1

	pid := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case GenericMessage:
			fmt.Println(i)
			i++
			testChan <- m.Value.(string)
		}
	})

	for i := 0; i < 10000; i++ {
		rootCtx.Send(pid, GenericMessage{Value: "Hello"})
		rootCtx.Send(pid, GenericMessage{Value: "Foo"})
		rootCtx.Send(pid, GenericMessage{Value: "Bar"})
	}

	for i := 0; i < 10000; i++ {
		assert.Equal(t, "Hello", <-testChan)
		assert.Equal(t, "Foo", <-testChan)
		assert.Equal(t, "Bar", <-testChan)
	}

	rootCtx.Send(pid, PoisonPill{})

	Wait()
}

func TestMonitorWithKill(t *testing.T) {
	rootCtx := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case GenericMessage:
			fmt.Println(m.Value)
		}
	})

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch m := message.(type) {
		case DownMessage:
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
		case GenericMessage:
			fmt.Println(m.Value)
		}
	})

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch m := message.(type) {
		case DownMessage:
			assert.Equal(t, p.String(), m.Who.String())
			fmt.Println("Actor went down!")
			ctx.Quit()
		}
	})

	<-time.After(50 * time.Millisecond)

	rootCtx.Send(p, PoisonPill{})

	Wait()
}

func TestMonitorAbortable_Abort(t *testing.T) {
	rootCtx := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case GenericMessage:
			fmt.Println(m.Value)
		}
	})

	var a Abortable

	q := SpawnWithInit(func(ctx *Context) {
		a = ctx.Monitor(p)
	}, func(ctx *Context, message Message) {
		switch message.(type) {
		case DownMessage:
			fmt.Println(":(")
			t.Fail()
		case GenericMessage:
			fmt.Println("Worked")
			ctx.Quit()
		}
	})

	a.Abort()

	<-time.After(50 * time.Millisecond)

	rootCtx.Send(p, PoisonPill{})

	rootCtx.Send(q, GenericMessage{Value: ""})

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
		case DownMessage:
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

func TestNewSystem(t *testing.T) {
	_, err := NewSystem("test")

	if err != nil {
		panic(err)
	}
}

func TestNewSystemWithHandler(t *testing.T) {
	s, err := NewSystem("test")

	if err != nil {
		panic(err)
	}

	p := Spawn(func(ctx *Context, message Message) {
		switch m := message.(type) {
		case GenericMessage:
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
		return
	}

	p, err := r.Remote("printer")

	if err != nil {
		t.Fail()
		return
	}

	rootCtx.Send(p, GenericMessage{Value: "Hello!"})

	<-time.After(50 * time.Millisecond)
}

func TestConnectRemoteKill(t *testing.T) {
	rootCtx := RootContext()

	r, err := Connect("test@localhost")

	if err != nil {
		t.Fail()
		return
	}

	p, err := r.Remote("printer")

	if err != nil {
		t.Fail()
		return
	}

	rootCtx.Kill(p)

	<-time.After(50 * time.Millisecond)
}

func TestConnectPoisonPill(t *testing.T) {
	rootCtx := RootContext()

	r, err := Connect("test@localhost")

	if err != nil {
		t.Fail()
		return
	}

	p, err := r.Remote("printer")

	if err != nil {
		t.Fail()
		return
	}

	rootCtx.Send(p, PoisonPill{})

	<-time.After(50 * time.Millisecond)
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
