package quacktors

type killControlMessage struct{}

type monitorControlMessage struct {
	Who *Pid
}

type demonitorControlMessage struct {
	Who *Pid
}
