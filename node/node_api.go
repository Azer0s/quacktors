package node

import (
	"encoding/json"
	"github.com/Azer0s/quacktors/messages"
	"github.com/Azer0s/quacktors/pid"
	"github.com/Azer0s/quacktors/util"
	"net"
	"strconv"
	"sync"
)

func NewSystem(name string) System {
	return System{
		name:       name,
		remotePids: make(map[string]pid.Pid),
	}
}

func StartGatewayServer(port int) {
	wg := sync.WaitGroup{}
	for {
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					//ignored
				}
				wg.Done()
			}()

			StartLink(port)
		}()
		wg.Wait()
	}
}

func StartSystemServer(system System) {
	wg := sync.WaitGroup{}
	for {
		wg.Add(1)
		go func() {
			defer func() {
				if r := recover(); r != nil {
					//ignored
				}
				wg.Done()
			}()
			//TODO: Start system UDP server
		}()
		wg.Wait()
	}
}

func ConnectRemote(system, addr string, port int) (Remote, error) {
	msg := messages.GatewayRequest{System: system}

	conn, err := net.Dial("udp", addr+":"+strconv.Itoa(port))
	if err != nil {
		return Remote{}, err
	}

	//noinspection GoUnhandledErrorResult
	defer conn.Close()

	b, err := json.Marshal(msg)
	if err != nil {
		return Remote{}, err
	}

	_, err = conn.Write(b)
	if err != nil {
		return Remote{}, err
	}

	buff := make([]byte, 2048)
	n, err := conn.Read(buff)
	if err != nil {
		return Remote{}, err
	}

	var res messages.GatewayResponse
	err = json.Unmarshal(buff[0:n], &res)
	if err != nil {
		return Remote{}, err
	}

	if res.Err {
		return Remote{}, util.RemoteConnectError()
	}

	systemAddr, err := net.ResolveUDPAddr("udp", addr+":"+strconv.Itoa(res.SystemPort))
	if err != nil {
		return Remote{}, err
	}

	return Remote{system: system, address: *systemAddr}, nil
}
