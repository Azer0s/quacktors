package quacktors

import "github.com/Azer0s/qpmd"

type Context struct {

}

func (c *Context) Self() Pid {
	request := qpmd.Response{}
	request.ResponseType = qpmd.RESPONSE_TIMEOUT
	return Pid{}
}

func (c *Context) Children() []Pid {
	return nil
}

func (c *Context) Send(to Pid, message Message) {

}

func (c *Context) Message() Message {
	return nil
}

func (c *Context) Monitor(process Pid) {

}