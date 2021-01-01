package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"net"
	"sync"
)

const handler = "handler"
const pidVal = "pid"

type System struct {
	name              string
	handlers          map[string]*Pid
	handlersMu        *sync.RWMutex
	quitChan          chan bool
	heartbeatQuitChan chan bool
	closed            bool
}

func (s *System) HandleRemote(name string, process *Pid) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	s.handlers[name] = process
}

func (s *System) IsClosed() bool {
	return s.closed
}

func (s *System) Close() {
	s.closed = true
	s.quitChan <- true
	s.heartbeatQuitChan <- true
}

func (s *System) startServer() (uint16, error) {
	return startServer(func(portChan chan int, errorChan chan error) {
		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start system server on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portChan <- port

		listen := make(chan net.Conn)

		go func() {
			for {
				conn, err := listener.Accept()
				if err != nil {
					continue
				}
				listen <- conn
			}
		}()

		for {
			select {
			case <-s.quitChan:
				return
			case conn := <-listen:
				go s.handleClient(conn)
			}
		}
	})
}

func (s *System) handleClient(conn net.Conn) {
	defer func() {
		recover()

		err := conn.Close()
		if err != nil {
			return
		}
	}()

	req, err := readRequest(conn)

	if err != nil {
		return
	}

	switch req.RequestType {
	case qpmd.REQUEST_HELLO:
		//I'll leave the hello message for now. Maybe it'll be useful in the future
		//(plus it's more consistent to machine to machine communication)
		_ = writeOk(conn, map[string]interface{}{})
	case qpmd.REQUEST_LOOKUP:
		s.handlersMu.RLock()
		defer s.handlersMu.RUnlock()

		h, ok := s.handlers[req.Data[handler].(string)]

		if ok {
			_ = writeOk(conn, map[string]interface{}{
				pidVal: h,
			})

			return
		}

		_ = writeError(conn, errors.New(fmt.Sprintf("couldn't find handler %s", req.Data[handler])))
	}
}
