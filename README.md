![Github Action](https://github.com/Azer0s/quacktors/workflows/Go/badge.svg) ![Go Report Card](https://goreportcard.com/badge/github.com/Azer0s/quacktors) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/Azer0s/quacktors/blob/master/LICENSE.md)


# quacktors

```go
self := quacktors.Self()
pid := quacktors.Spawn(func() {
    quacktors.Send(self, "Hello")
})​
msg := quacktors.Receive()
fmt.Println(msg)
```

```go
quacktors.SetHostname("alex@localhost")

pid := quacktors.Spawn(func() {
    for {
        fmt.Println(quacktors.Receive())
    }
})

quacktors.HandleRemote()
```

```go
quacktors.SetHostname("john@localhost")

node := quacktors.Connect("alex@localhost")
printer := node.Remote("printer")

quacktors.Send(printer, "Hello, world")
```
