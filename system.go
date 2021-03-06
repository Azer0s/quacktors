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

//The System struct represents a logical actor system (i.e. a
//collection of PIDs that have been assigned a handler
//name so that remote machines can look them up).
//Furthermore, the System struct also keeps track of the
//connection to the local qpmd and keeps track of the
//system server status (the server which is used to
//look up PIDs by handler names, etc).
type System struct {
	name              string
	handlers          map[string]*Pid
	handlersMu        *sync.RWMutex
	quitChan          chan bool
	heartbeatQuitChan chan bool
	closed            bool
}

//HandleRemote associates a PID with a handler name.
func (s *System) HandleRemote(name string, process *Pid) {
	s.handlersMu.Lock()
	defer s.handlersMu.Unlock()

	s.handlers[name] = process
}

//IsClosed returns true if the connection to the
//local qpmd or the system server were closed.
func (s *System) IsClosed() bool {
	return s.closed
}

//Close closes the connection to the local qpmd
//and quits the system server.
func (s *System) Close() {
	s.closed = true
	s.quitChan <- true
	s.heartbeatQuitChan <- true
}

func (s *System) startServer() (uint16, error) {
	logger.Debug("starting system server",
		"system_name", s.name)

	return startServer(func(portChan chan int, errorChan chan error) {
		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start system server on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portChan <- port

		logger.Debug("started system server successfully",
			"system_name", s.name)

		for {
			select {
			case <-s.quitChan:
				logger.Info("quitting system server",
					"system_name", s.name)
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					logger.Warn("there was an error while accepting an incoming client for system",
						"system_name", s.name,
						"error", err)

					continue
				}

				logger.Info("handling incoming client for system",
					"system_name", s.name,
					"client", conn.RemoteAddr().String())

				go s.handleClient(conn)
			}
		}
	})
}

func (s *System) handleClient(conn net.Conn) {
	c := conn.RemoteAddr().String()

	defer func() {
		recover()

		logger.Debug("closing system server connection to client",
			"system_name", s.name,
			"client", c)

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
		logger.Debug("handling system server hello request",
			"system_name", s.name,
			"client", c)

		err = writeOk(conn, map[string]interface{}{})

		if err != nil {
			logger.Warn("there was an error while sending ok message to client",
				"client", c,
				"error", err)
		}
	case qpmd.REQUEST_LOOKUP:
		s.handlersMu.RLock()
		defer s.handlersMu.RUnlock()

		handlerName := req.Data[handler].(string)

		logger.Debug("handling system server lookup request",
			"system_name", s.name,
			"client", c,
			"handler_name", handlerName)

		h, ok := s.handlers[handlerName]

		if ok {
			err = writeOk(conn, map[string]interface{}{
				pidVal: h,
			})

			if err != nil {
				logger.Warn("there was an error while sending ok message to client",
					"client", c,
					"error", err)
			}

			return
		}

		logger.Warn("couldn't find handler for system server lookup request",
			"system_name", s.name,
			"client", c,
			"handler_name", handlerName)

		err = writeError(conn, fmt.Errorf("couldn't find handler %s", handlerName))

		if err != nil {
			logger.Warn("there was an error while sending error message to client",
				"client", c,
				"error", err)
		}
	}
}
