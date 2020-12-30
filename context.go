package quacktors

type Context struct {
	self *Pid
}

func (c *Context) Self() *Pid {
	return c.self
}

func (c *Context) Send(to *Pid, message Message) {
	doSend(to, message)
}

func (c *Context) Kill(pid *Pid) {
	go func() {
		pid.quitChanMu.RLock()
		defer pid.quitChanMu.RUnlock()

		if pid.quitChan == nil {
			return
		}

		pid.quitChan <- true
	}()
}

func (c *Context) Quit() {
	panic("Bye cruel world!")
}

func (c *Context) Monitor(pid *Pid) Abortable {
	pid.monitorChanMu.RLock()
	defer pid.monitorChanMu.RUnlock()

	if pid.monitorChan == nil {
		return &NoopAbortable{}
	}

	pid.monitorChan <- c.self

	return &MonitorAbortable{
		pid:  pid,
		self: c.self,
	}
}
