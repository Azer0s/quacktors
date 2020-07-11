![Github Action](https://github.com/Azer0s/quacktors/workflows/Go/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/Azer0s/quacktors) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Azer0s/quacktors/blob/master/LICENSE.md)


# quacktors

```go
self := quacktors.Self()
pid := quacktors.Spawn(func() {
    quacktors.Send(self, "Hello")
})
msg := quacktors.Receive()
fmt.Println(msg)
```

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
