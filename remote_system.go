package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/vmihailenco/msgpack/v5"
	"net"
)

type RemoteSystem struct {
	MachineId string
	Address   string
	Port      uint16
}

func (r *RemoteSystem) sayHello() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", r.Address, r.Port))
	if err != nil {
		return err
	}

	b, err := msgpack.Marshal(qpmd.Request{
		RequestType: qpmd.REQUEST_HELLO,
		Data: map[string]interface{}{
			qpmd.MACHINE_ID:           machineId,
			qpmd.MESSAGE_GATEWAY_PORT: messageGatewayPort,
			qpmd.GP_GATEWAY_PORT:      gpGatewayPort,
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

	b, err := msgpack.Marshal(qpmd.Request{
		RequestType: qpmd.REQUEST_LOOKUP,
		Data: map[string]interface{}{
			handler: handlerName,
		},
	})
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(b)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4096)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, err
	}

	res := qpmd.Response{}
	err = msgpack.Unmarshal(buf, &res)
	if err != nil {
		return nil, err
	}

	if res.ResponseType != qpmd.RESPONSE_OK {
		return nil, errors.New("remote system returned non okay result")
	}

	pidData := res.Data[pid].(map[string]interface{})

	return &Pid{
		MachineId: pidData["MachineId"].(string),
		Id:        pidData["Id"].(string),
	}, nil
}
