package quacktors

import (
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/vmihailenco/msgpack/v5"
	"net"
)

var qpmdPort = 7161

func init() {
	failIfConnectionError := func(err error) {
		if err != nil {
			panic("Couldn't connect to qpmd! Is qpmd running?")
		}
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", qpmdPort))
	failIfConnectionError(err)

	b, err := msgpack.Marshal(qpmd.Request{
		RequestType: qpmd.REQUEST_HELLO,
		Data: map[string]interface{}{},
	})
	try(err)

	_, err = conn.Write(b)
	failIfConnectionError(err)

	buf := make([]byte, 4096)
	_, err = conn.Read(buf)
	failIfConnectionError(err)

	res := qpmd.Response{}
	err = msgpack.Unmarshal(buf, &res)
	try(err)
}

func qpmdLookup(system, remoteAddress string, remotePort int) (error) {
	port := qpmdPort

	if remotePort != 0 {
		port = remotePort
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remoteAddress, port))
	if err != nil {
		return err
	}

	b, err := msgpack.Marshal(qpmd.Request{
		RequestType: qpmd.REQUEST_LOOKUP,
		Data: map[string]interface{}{
			"system": system,
		},
	})
	if err != nil {
		return err
	}

	_, err = conn.Write(b)
	if err != nil {
		return err
	}

	buf := make([]byte, 4096)
	_, err = conn.Read(buf)
	if err != nil {
		return err
	}

	res := qpmd.Response{}
	err = msgpack.Unmarshal(buf, &res)
	if err != nil {
		return err
	}

	//TODO: Return response in specific dao

	return nil
}