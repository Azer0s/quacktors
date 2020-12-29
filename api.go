package quacktors

func RegisterType(message Message) {
	storeType(message)
}

func Spawn(action func(ctx *Context, message Message)) Pid {
	return Pid{}
}

func SpawnStateful(actor Actor) Pid {
	return Pid{}
}

func NewSystem(name string) (System, error) {
	return System{}, nil
}

func Send(to Pid, message Message) {

}

func Connect(name string) {

}