package node

import (
	"encoding/json"
	"github.com/Azer0s/quacktors/messages"
	"github.com/Azer0s/quacktors/util"
	"net"
	"strconv"
)

var remotePort int

func SetRemotePort(port int) {
	remotePort = port
}

func StartLink() {
	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(remotePort))
	if err != nil {
		panic(err)
	}

	connection, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}

	for {
		buffer := make([]byte, 2048)
		n, addr, err := connection.ReadFromUDP(buffer)
		go handleConnection(n, addr, err, buffer, connection)
	}
}

func handleConnection(n int, addr *net.UDPAddr, err error, buffer []byte, connection *net.UDPConn) {
	if err != nil {
		util.SendErr(connection, addr)
		return
	}

	var request messages.GatewayRequest
	err = json.Unmarshal(buffer[0:n-1], &request)

	if err != nil {
		util.SendErr(connection, addr)
		return
	}

	p, err := GetSystemPort(request.System)

	if err != nil {
		util.SendErr(connection, addr)
		return
	}

	data, err := json.Marshal(messages.GatewayResponse{
		Err:        false,
		SystemPort: p,
	})

	if err != nil {
		util.SendErr(connection, addr)
		return
	}

	_, _ = connection.WriteToUDP(data, addr)
}
