package node

import (
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
)

type System struct {
	name       string
	remotePids map[string]pid.Pid
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
