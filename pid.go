package quacktors

import (
	"fmt"
	"sync"
)

type Pid struct {
	MachineId           string
	Id                  string
	quitChan            chan<- bool
	quitChanMu          *sync.RWMutex
	messageChan         chan<- Message
	messageChanMu       *sync.RWMutex
	monitorChan         chan<- *Pid
	monitorChanMu       *sync.RWMutex
	demonitorChan       chan<- *Pid
	demonitorChanMu     *sync.RWMutex
	scheduled           map[string]chan bool
	monitorQuitChannels map[string]chan bool
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
	}
}

func (p *Pid) String() string {
	return fmt.Sprintf("%s@%s", p.Id, p.MachineId)
}
