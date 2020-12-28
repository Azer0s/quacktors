package quacktors

type Context struct {

}

func (c *Context) Self() Pid {
	return Pid{}
}

func (c *Context) Children() []Pid {
	return nil
}

func (c *Context) Receive() Message {
	return DebugReceive()
}

func (c *Context) Monitor(process Pid) {

}