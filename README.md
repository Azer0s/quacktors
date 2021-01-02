<img src="assets/quacktor-logo.png" alt="logo" align="left"/>

## quacktors

![Github Action](https://github.com/Azer0s/quacktors/workflows/Go/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/Azer0s/quacktors) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Azer0s/quacktors/blob/master/LICENSE.md)

quacktors or "quick actors" is a Go framework that brings Erlang/Elixir style concurrency to Go! It allows for message passing, actor monitoring and can even deal with remote actors/systems. Furthermore, it comes with plenty of useful standard modules for building actor model systems (like Supervisors, Genservers, etc.). Oh and btw: quacktors is super easy to use!

```go
rootCtx := quacktors.RootContext()

pid := quacktors.Spawn(func(ctx *Context, message Message) {
    fmt.Println("Hello, quacktors!")
})

rootCtx.Send(pid, EmptyMessage{})
```

<br>

### Getting started

```go
foo := quacktors.NewSystem("foo")

pid := quacktors.Spawn(func(ctx *Context, message Message) {
    switch m := message.(type) {
    case GenericMessage:
        fmt.Println(m.Value)
    }
})

foo.HandleRemote("printer", pid)

quacktors.Wait()
```

```go
rootCtx := quacktors.RootContext()

node := quacktors.Connect("foo@localhost")
printer, ok := node.Remote("printer")

rootCtx.Send(printer, GenericMessage{Value: "Hello, world"})
```
