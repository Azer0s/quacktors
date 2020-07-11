package test

import (
	"github.com/Azer0s/quacktors"
	"net"
	"testing"
)

const loopback = "127.0.0.1"

//noinspection GoUnhandledErrorResult
func TestGatewayConnection(t *testing.T) {
	quacktors.StartGateway(5521)

	conn, err := net.Dial("udp", ":5521")
	if err != nil {
		t.Error("could not connect to server: ", err)
	}
	defer conn.Close()
}

