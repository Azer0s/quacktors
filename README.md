# Quacktors

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