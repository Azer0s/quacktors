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

func init() {
	RegisterType(&DownMessage{})
}