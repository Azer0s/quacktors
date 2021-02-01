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
			if i%100 == 0 {
				fmt.Println(i)
			}
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

	Run()
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

	Run()
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

	Run()
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

	Run()
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

	Run()
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

	p := SpawnStateful(actor)
	context := RootContext()
	context.Kill(p)

	Run()
}

func TestContext_SendAfter(t *testing.T) {
	rootContext := RootContext()

	p := Spawn(func(ctx *Context, message Message) {
		t.Fail()
	})

	a := rootContext.SendAfter(p, GenericMessage{}, 1*time.Second)
	a.Abort()

	<-time.After(3 * time.Second)

	rootContext.Kill(p)

	Run()
}

func TestNewSystem(t *testing.T) {
	_, err := NewSystem("test")

	if err != nil {
		panic(err)
	}
}

/*
Remote tests are commented out because they can, as of right now, only be run manually

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

	Run()

	<-time.After(1 * time.Second)
}

func TestContext_MonitorMachine(t *testing.T) {
	r, err := Connect("test@localhost")

	if err != nil {
		t.Fail()
		return
	}

	SpawnWithInit(func(ctx *Context) {
		a := ctx.MonitorMachine(r.Machine)
		a.Abort()
	}, func(ctx *Context, message Message) {
		fmt.Println(message)
		ctx.Quit()
	})

	Run()
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

func TestConnectRemoteMonitor(t *testing.T) {
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

	SpawnWithInit(func(ctx *Context) {
		ctx.Monitor(p)
		//a := ctx.Monitor(p)
		//<-time.After(1 * time.Second)
		//a.Abort()
		//<-time.After(1 * time.Second)
	}, func(ctx *Context, message Message) {
		fmt.Println(message)
		ctx.Quit()
	})

	<-time.After(5 * time.Second)

	rootCtx := RootContext()
	rootCtx.Kill(p)

	Run()
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
*/
