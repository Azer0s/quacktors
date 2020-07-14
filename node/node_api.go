package node

import "github.com/Azer0s/quacktors/pid"

func NewSystem(name string) System {
	return System{
		name:       name,
		remotePids: make(map[string]pid.Pid),
	}
}

func StartSystem(system System) {
	//TODO: Start system UDP server
}

func ConnectRemote(system, addr string, port int) (Remote, error) {
	return Remote{}, nil
}
