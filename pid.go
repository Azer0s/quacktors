package quacktors

import (
	"fmt"
	"sync"
)

type Pid struct {
	MachineId       string
	Id              string
	quitChan        chan<- bool
	quitChanMu      *sync.RWMutex
	messageChan     chan<- Message
	messageChanMu   *sync.RWMutex
	monitorChan     chan<- *Pid
	monitorChanMu   *sync.RWMutex
	demonitorChan   chan<- *Pid
	demonitorChanMu *sync.RWMutex
	//Stores channels to scheduled tasks (monitors, SendAfter, monitors the actor itself launches but doesn't consume)
	scheduled map[string]chan bool
	//Stores channels to tell a monitor taks to quit (when a pid is demonitored)
	monitorQuitChannels map[string]chan bool
	//Is locked when `scheduled` or `monitorQuitChannels` is modified
	monitorSetupMu *sync.Mutex
}

func createPid(quitChan chan<- bool, messageChan chan<- Message, monitorChan chan<- *Pid, demonitorChan chan<- *Pid, scheduled map[string]chan bool, monitorQuitChannels map[string]chan bool) *Pid {
	return &Pid{
		MachineId:           machineId,
		Id:                  "",
		quitChan:            quitChan,
		quitChanMu:          &sync.RWMutex{},
		messageChan:         messageChan,
		messageChanMu:       &sync.RWMutex{},
		monitorChan:         monitorChan,
		monitorChanMu:       &sync.RWMutex{},
		demonitorChan:       demonitorChan,
		demonitorChanMu:     &sync.RWMutex{},
		scheduled:           scheduled,
		monitorQuitChannels: monitorQuitChannels,
		monitorSetupMu:      &sync.Mutex{},
	}
}

func (p *Pid) String() string {
	return fmt.Sprintf("%s@%s", p.Id, p.MachineId)
}
