package quacktors

type machine struct {
	machineId     string
	address       string
	gatewayPort   uint16
	gpPort        uint16
	quitChan      chan<- *Pid
	messageChan   chan<- Message
	monitorChan   chan<- remoteMonitorTuple
	demonitorChan chan<- remoteMonitorTuple
	scheduled     map[string]chan bool
}

func (m *machine) monitorConnection(pid *Pid) {
	//connectionDownChan := make(chan bool)

	go func() {

	}()
}

func (m *machine) connect() {

}
