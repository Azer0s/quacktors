package quacktors

func RegisterType(message Message) {
	storeType(message)
}

func Spawn(action func(ctx *Context)) Pid {
	return Pid{}
}

func Send(receiver Pid, value Message) {

}

func NewSystem(name string) (System, error) {
	return System{}, nil
}

func Connect(name string) {

}