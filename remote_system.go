package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"net"
)

//TODO: log

type RemoteSystem struct {
	MachineId string
	Address   string
	Port      uint16
	Machine   *Machine
}

func (r *RemoteSystem) sayHello() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.Address, r.Port))
	if err != nil {
		return err
	}

	err = sendRequest(conn, qpmd.Request{
		RequestType: qpmd.REQUEST_HELLO,
		Data:        make(map[string]interface{}),
	})

	if err != nil {
		return err
	}

	res, err := readResponse(conn)
	if err != nil {
		return err
	}

	if res.ResponseType != qpmd.RESPONSE_OK {
		return errors.New("remote system returned non okay result")
	}

	return nil
}

func (r *RemoteSystem) Remote(handlerName string) (*Pid, error) {
	if !r.Machine.connected {
		return nil, errors.New("remote machine is not connected")
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.Address, r.Port))
	if err != nil {
		return nil, err
	}

	err = sendRequest(conn, qpmd.Request{
		RequestType: qpmd.REQUEST_LOOKUP,
		Data: map[string]interface{}{
			handler: handlerName,
		},
	})

	if err != nil {
		return nil, err
	}

	res, err := readResponse(conn)

	if err != nil {
		return nil, err
	}

	if res.ResponseType != qpmd.RESPONSE_OK {
		return nil, errors.New("remote system returned non okay result")
	}

	pid, err := parsePid(res.Data[pidVal].(map[string]interface{}))

	if err != nil {
		return nil, err
	}

	return pid, nil
}
