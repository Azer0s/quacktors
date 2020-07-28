package quacktors

import (
	"github.com/Azer0s/quacktors/actors"
	"github.com/Azer0s/quacktors/node"
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
)

// Self returns the PID of the caller goroutine/actor
func Self() pid.Pid {
	goid := util.GetGoid()
	p, err := actors.GetByGoid(goid)

	if err != nil {
		var id string
		p, id = pid.NewPid()
		actors.StoreByGoid(goid, p, id)
	}

	return p
}

// Spawn spawns an actor by a function and returns the actors PID
func Spawn(action func()) pid.Pid {
	return actors.Spawn(action)
}

// Send sends data to a PID; this is a non-blocking call
func Send(pid pid.Pid, data interface{}) {
	orderingComplete := make(chan interface{})
	go pid.Send(data, orderingComplete)
	<-orderingComplete
}

// Receive receives data sent to the caller goroutine/actor; this is a blocking call
func Receive() interface{} {
	p := Self()
	return util.PidToLocalPid(p).Receive()
}

// Monitor monitors a PID; when the state of the monitored PID goes down, a message is sent to the monitoring actor
func Monitor(toMonitor pid.Pid) {
	p := Self()

	if !toMonitor.Up() {
		panic("Actor to monitor is down!")
	}

	toMonitor.Monitor(p)
}

// StartGateway starts the remote gateway so other actor systems can reach local actor systems
func StartGateway(port int) {
	go func() {
		node.StartGatewayServer(port)
	}()
}

func NewSystem(name string) node.System {
	system := node.NewSystem(name)
	node.StoreSystem(system)
	node.StorePortBinding(system.SetupLink(), name)
	//TODO: Start system server

	return system
}

func Connect(address string) (node.Remote, error) {
	s, a, p, err := util.ParseAddress(address)

	if err != nil {
		return node.Remote{}, err
	}

	return node.ConnectRemote(s, a, p)
}
