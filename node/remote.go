package node

import "net"

type Remote struct {
	system  string
	address net.UDPAddr
}
