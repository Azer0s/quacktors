package quacktors

func RegisterType(message Message) {
	storeType(message)
}

func RootContext() Context {
	return Context{}
}

func Spawn(action func(ctx *Context, message Message)) *Pid {
	return startActor(&StatelessActor{
		initFunction: func(ctx *Context) {},
		receiveFunction: action,
	})
}

func SpawnWithInit(init func(ctx *Context), action func(ctx *Context, message Message)) *Pid {
	return startActor(&StatelessActor{
		initFunction:    init,
		receiveFunction: action,
	})
}

func SpawnStateful(actor Actor) *Pid {
	return startActor(actor)
}

func NewSystem(name string) (System, error) {
	//port, err := startServer()

	//conn, err := qpmdRegister(name, port)
	//qpmdHeartbeat(conn)
	return System{}, nil
}

func Connect(name string) RemoteSystem {
	//disconnectChan := make(chan bool)

	//qpmdLookup()
	//remoteSystemHello()
	//remoteSystemConnection(disconnectChan)

	return RemoteSystem{}
}
