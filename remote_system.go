package quacktors

import (
	"encoding/json"
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
		Data:        map[string]interface{}{},
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

	res, err := readResponse(conn)

	if err != nil {
		return nil, err
	}

	if res.ResponseType != qpmd.RESPONSE_OK {
		return nil, errors.New("remote system returned non okay result")
	}

	pidData, err := json.Marshal(res.Data[pidVal])

	if err != nil {
		return nil, err
	}

	pid := &Pid{}
	err = json.Unmarshal(pidData, &pid)

	return pid, nil
}
