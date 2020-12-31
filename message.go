package quacktors

type Message interface {
	Type() string
}

type DownMessage struct {
	Who *Pid
}

func (d *DownMessage) Type() string {
	return "DownMessage"
}

type PoisonPill struct {
	//PoisonPill message to kill actor
}

func (p *PoisonPill) Type() string {
	return "PoisonPill"
}

type GenericMessage struct {
	Value interface{}
}

func (g *GenericMessage) Type() string {
	return "GenericMessage"
}

func init() {
	RegisterType(&DownMessage{})
	RegisterType(&PoisonPill{})
	RegisterType(&GenericMessage{})
}
