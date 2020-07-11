package node

import (
	"log"
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
	log.Println("-> ", string(buffer[0:n-1]))
	data := []byte("Hello there!\n")

	_, _ = connection.WriteToUDP(data, addr)
}
