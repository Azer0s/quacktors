package quacktors

type Message interface {
	Type() string
}

type DownMessage struct {
	Who *Pid
}

func (d DownMessage) Type() string {
	return "DownMessage"
}

type PoisonPill struct {
	//PoisonPill message to kill actor
}

func (p PoisonPill) Type() string {
	return "PoisonPill"
}

type GenericMessage struct {
	Value interface{}
}

func (g GenericMessage) Type() string {
	return "GenericMessage"
}

type EmptyMessage struct {
}

func (e EmptyMessage) Type() string {
	return "EmptyMessage"
}

type KillMessage struct {
	//KillMessage to signal to actor it should go down
}

func (k KillMessage) Type() string {
	return "KillMessage"
}

type DisconnectMessage struct {
	MachineId string
	Address   string
}

func (d DisconnectMessage) Type() string {
	return "DisconnectMessage"
}
