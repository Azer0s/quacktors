package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/rs/zerolog/log"
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
	log.Debug().
		Str("system_name", s.name).
		Msg("starting system server")

	return startServer(func(portChan chan int, errorChan chan error) {
		listener, err := net.Listen("tcp", ":0")

		if err != nil {
			errorChan <- errors.New("couldn't start system server on random port")
			return
		}

		port := listener.Addr().(*net.TCPAddr).Port
		portChan <- port

		log.Debug().
			Str("system_name", s.name).
			Msg("started system server successfully")

		for {
			select {
			case <-s.quitChan:
				log.Info().
					Str("system_name", s.name).
					Msg("quitting system server")
				return
			default:
				conn, err := listener.Accept()
				if err != nil {
					log.Warn().
						Str("system_name", s.name).
						Err(err).
						Msg("there was an error while accepting an incoming client for system")

					continue
				}

				log.Info().
					Str("system_name", s.name).
					Str("client", conn.RemoteAddr().String()).
					Msg("handling incoming client for system")

				go s.handleClient(conn)
			}
		}
	})
}

func (s *System) handleClient(conn net.Conn) {
	c := conn.RemoteAddr().String()

	defer func() {
		recover()

		log.Info().
			Str("system_name", s.name).
			Str("client", c).
			Msg("closing system server connection to client")

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
		log.Info().
			Str("system_name", s.name).
			Str("client", c).
			Msg("handling system server hello request")

		err = writeOk(conn, map[string]interface{}{})

		if err != nil {
			log.Warn().
				Str("client", c).
				Err(err).
				Msg("there was an error while sending an ok message to a client")
		}
	case qpmd.REQUEST_LOOKUP:
		s.handlersMu.RLock()
		defer s.handlersMu.RUnlock()

		handlerName := req.Data[handler].(string)

		log.Info().
			Str("system_name", s.name).
			Str("client", c).
			Str("handler_name", handlerName).
			Msg("handling system server lookup request")

		h, ok := s.handlers[handlerName]

		if ok {
			err = writeOk(conn, map[string]interface{}{
				pidVal: h,
			})

			if err != nil {
				log.Warn().
					Str("client", c).
					Err(err).
					Msg("there was an error while sending an ok message to a client")
			}

			return
		}

		log.Warn().
			Str("system_name", s.name).
			Str("client", c).
			Str("handler_name", handlerName).
			Msg("couldn't find handler for system server lookup request")

		err = writeError(conn, errors.New(fmt.Sprintf("couldn't find handler %s", handlerName)))

		if err != nil {
			log.Warn().
				Str("client", c).
				Err(err).
				Msg("there was an error while sending an error message to a client")
		}
	}
}
