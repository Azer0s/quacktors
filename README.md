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
