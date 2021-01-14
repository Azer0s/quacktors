<img src="assets/quacktor-logo.png" alt="logo" align="left"/>

## quacktors

![Github Action](https://github.com/Azer0s/quacktors/workflows/Go/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/Azer0s/quacktors) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Azer0s/quacktors/blob/master/LICENSE.md)

quacktors or "quick actors" is a Go framework that brings Erlang/Elixir style concurrency to Go! It allows for message passing, actor monitoring and can even deal with remote actors/systems. Furthermore, it comes with plenty of useful standard modules for building actor model systems (like Supervisors, Relays, etc.). Oh and btw: quacktors is super easy to use!

```go
rootCtx := RootContext()

pid := Spawn(func(ctx *Context, message Message) {
    fmt.Println("Hello, quacktors!")
})

rootCtx.Send(pid, EmptyMessage{})
```

<br>

### Getting started

To get started, you'll need an installation of qpmd (see: [qpmd](https://github.com/Azer0s/qpmd)).
The quacktor port mapper daemon is responsible for keeping track of all running systems and quacktor instances on your local machine and acts as a "DNS server" for remote machines that want to connect to a local system.

```go
import . "github.com/Azer0s/quacktors"

foo := NewSystem("foo")

pid := Spawn(func(ctx *Context, message Message) {
    switch m := message.(type) {
    case GenericMessage:
        fmt.Println(m.Value)
    }
})

foo.HandleRemote("printer", pid)

Wait()
```

```go
rootCtx := RootContext()

node := Connect("foo@localhost")
printer, ok := node.Remote("printer")

rootCtx.Send(printer, GenericMessage{Value: "Hello, world"})
```

### Monitoring actors

quacktors can monitor both local, as well as remote actors. As soon as the monitored actor goes down, a `DownMessage` is sent out to the monitoring actor.

```go
pid := Spawn(func(ctx *Context, message Message) {
})

SpawnWithInit(func(ctx *Context) {
    ctx.Monitor(pid)
}, func(ctx *Context, message Message) {
    switch m := message.(type) {
        case DownMessage:
            ctx.Logger.Info("received down message from other actor", 
                "pid", m.String())
            ctx.Quit()
    }
})

Wait()
```

### Tracing

quacktors supports [opentracing](https://opentracing.io/) out of the box! It's as easy as setting the global tracer (and optionally providing a span to the root context).

```go
cfg := jaegercfg.Configuration{
    ServiceName: "TestNewSystemWithHandler",
    Sampler: &jaegercfg.SamplerConfig{
        Type:  jaeger.SamplerTypeConst,
        Param: 1,
    },
    Reporter: &jaegercfg.ReporterConfig{
        LogSpans: true,
    },
}
tracer, closer, _ := cfg.NewTracer()
defer closer.Close()

opentracing.SetGlobalTracer(tracer)

span := opentracing.GlobalTracer().StartSpan("root")
defer span.Finish()
rootCtx := RootContextWithSpan(span)

a1 := SpawnWithInit(func(ctx *Context) {
	ctx.Trace("a1")
}, func(ctx *Context, message Message) {
	ctx.Span().SetTag("message_type", message.Type())
	<-time.After(3 * time.Second)
})

rootCtx.Send(a1, EmptyMessage{})

Wait()
```

### Supervision

quacktors comes with some cool standard components, one of which is the supervisor. The supervisor (as the name implies) supervises one or many named actors and reacts to failures according to a set strategy.

```go
SpawnStateful(component.Supervisor(component.ALL_FOR_ONE_STRATEGY, map[string]Actor{
    "1": &superImportantActor{id: 1},
    "2": &superImportantActor{id: 2},
    "3": &superImportantActor{id: 3},
    "4": &superImportantActor{id: 4},
}))
```

### Location transparency

Sending messages in quacktors is completely location transparent, meaning no more worrying about connections, marshalling, unmarshalling, error handling and all that other boring stuff. Just send what you want to whoever you want to send it to. It's that easy.

### Floating PIDs

PIDs in quacktors are floating, meaning you can send a PID to a remote machine as a message and use that same PID there as you would use any other PID.

```go
foo := NewSystem("foo")

ping := Spawn(func(ctx *Context, message Message) {
    switch m := message.(type) {
    case Pid:
        ctx.Logger.Info("ping")
        <- time.After(1 * time.Second)
        ctx.Send(&m, *ctx.Self())
    }
})

foo.HandleRemote("ping", ping)

Wait()
```

```go
rootCtx := RootContext()

bar := NewSystem("bar")

pong := Spawn(func(ctx *Context, message Message) {
    switch m := message.(type) {
    case Pid:
        ctx.Logger.Info("pong")
        <- time.After(1 * time.Second)
        ctx.Send(&m, *ctx.Self())
    }
})

bar.HandleRemote("pong", pong)

foo := Connect("foo@localhost")
ping := foo.Remote("ping")

rootCtx.Send(ping, *pong)

Wait()
```

### GenServers

As part of the default component set, quacktors supports Elixir style GenServers. The handlers for these are configured via the method names via reflection. So a GenServer with a `Call` handler for a `PrintRequest` would look like so:

```go
type PrintRequest struct {
    //our printing request message
}

func (p PrintRequest) Type() string {
    return "PrintRequest"
}

type Printer struct { 
    //printing magic
}

func (p Printer) HandlePrintRequestCall(ctx *Context, message PrintRequest) Message {
    //print stuff
	
    return EmptyMessage{}
}


pid := Spawn(genserver.New(Printer{}))
res, err := genserver.Call(pid, PrintRequest{})
```

So you don't even have to write your own actors if you don't want to. Cool, isn't it?

### On message order and reception

In quacktors, message order is guaranteed from one actor to another. Meaning that if you send messages from A to B, they will arrive in order. The same is true for remote actors.

For multiple actors (A, B & C send messages to D), we can't make that guarantee because we don't know when each actor will execute.

As with basically all other actor systems, there is no guarantee (or even acknowledgement) that a message has been received. `Send` is a non-blocking call and doesn't return anything (even if the sending procedure failed).

### Configuring quacktors

quacktors has some configuration options which can be set by using the `config` package during `init`.

```go
func init() {
    config.SetLogger(&MyCustomLogger{})
    config.SetQpmdPort(7777)
}
```
