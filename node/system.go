package node

import (
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
	"net"
	"strconv"
	"strings"
)

type System struct {
	name       string
	remotePids map[string]pid.Pid
	conn *net.UDPConn
}

func (s *System) HandleRemote(name string, handler pid.Pid) {
	s.remotePids[name] = handler
}

func (s *System) GetHandler(name string) (pid.Pid, error) {
	if p, ok := s.remotePids[name]; ok {
		return p, nil
	}

	return nil, util.NoSuchPidInSystemError()
}

func (s *System) SetupLink() int {
	addr, err := net.ResolveUDPAddr("udp", ":0")
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	s.conn = connection

	a := strings.Split(connection.LocalAddr().String(), ":")
	port, err := strconv.Atoi(a[len(a) - 1])
	return port
}