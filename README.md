<img src="assets/quacktor-logo.png" alt="logo" align="left"/>

## quacktors

![Github Action](https://github.com/Azer0s/quacktors/workflows/Go/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/Azer0s/quacktors) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Azer0s/quacktors/blob/master/LICENSE.md)

quacktors or "quick actors" is a Go framework that brings Erlang/Elixir style concurrency to Go! It allows for message passing, actor monitoring and can even deal with remote actors/systems. Furthermore, it comes with plenty of useful standard modules for building actor model systems (like Supervisors, Genservers, etc.). Oh and there's more: quacktors is super easy to use!

```go
self := quacktors.Self()
pid := quacktors.Spawn(func() {
    quacktors.Send(self, "Hello, quacktors!")
})
msg := quacktors.Receive()
fmt.Println(msg)
```

<br>

### Getting started

```go
quacktors.StartGateway(5521)
foo := quacktors.NewSystem("foo")

pid := quacktors.Spawn(func() {
    for {
        fmt.Println(quacktors.Receive())
    }
})

foo.HandleRemote("printer", pid)
```

```go
node := quacktors.Connect("foo@127.0.0.1:5521")
printer := node.Remote("printer")

quacktors.Send(printer, "Hello, world")
```
