package quacktors

import (
	"errors"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/vmihailenco/msgpack/v5"
	"go.uber.org/zap"
	"net"
	"time"
)

var qpmdPort = 7161
var messageGatewayPort = uint16(0)
var gpGatewayPort = uint16(0)

func init() {
	failIfConnectionError := func(err error) {
		if err != nil {
			panic("Couldn't connect to qpmd! Is qpmd running?")
		}
	}

	var err error

	messageGatewayPort, err = startMessageGateway()
	if err != nil {
		panic(err)
	}

	gpGatewayPort, err = startGeneralPurposeGateway()
	if err != nil {
		panic(err)
	}

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", qpmdPort))
	failIfConnectionError(err)

	b, err := msgpack.Marshal(qpmd.Request{
		RequestType: qpmd.REQUEST_HELLO,
		Data: map[string]interface{}{
			qpmd.MACHINE_ID:           machineId,
			qpmd.MESSAGE_GATEWAY_PORT: messageGatewayPort,
			qpmd.GP_GATEWAY_PORT:      gpGatewayPort,
		},
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

func qpmdRegister(system *System, systemPort uint16) (net.Conn, error) {
	logger.Debug("registering system to qpmd", zap.String("system_name", system.name))

	conn, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", qpmdPort))
	if err != nil {
		return nil, err
	}

	err = sendRequest(conn, qpmd.Request{
		RequestType: qpmd.REQUEST_REGISTER,
		Data: map[string]interface{}{
			qpmd.SYSTEM_NAME: system.name,
			qpmd.PORT:        systemPort,
			qpmd.MACHINE_ID:  machineId,
		},
	})

	if err != nil {
		return nil, err
	}

	res, err := readResponse(conn)
	if err != nil {
		return nil, err
	}

	if res.ResponseType == qpmd.RESPONSE_OK {
		return conn, nil
	}

	return nil, errors.New("qpmd returned error on registration")
}

func qpmdHeartbeat(conn net.Conn, system *System) {
	quit := func() {
		system.quitChan <- true
		system.closed = true
	}

	go func() {
		for {
			select {
			case <-system.heartbeatQuitChan:
				return
			case <-time.After(25 * time.Second):
				err := sendRequest(conn, qpmd.Request{
					RequestType: qpmd.HEARTBEAT,
					Data:        map[string]interface{}{},
				})
				if err != nil {
					quit()
					return
				}

				res, err := readResponse(conn)

				if err != nil || res.ResponseType != qpmd.RESPONSE_OK {
					quit()
					return
				}
			}
		}
	}()
}

func qpmdLookup(system, remoteAddress string) (*RemoteSystem, error) {
	logger.Debug("looking up remote system in qpmd",
		zap.String("system_name", system),
		zap.String("remote_address", remoteAddress),
	)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", remoteAddress, qpmdPort))
	if err != nil {
		return &RemoteSystem{}, err
	}

	err = sendRequest(conn, qpmd.Request{
		RequestType: qpmd.REQUEST_LOOKUP,
		Data: map[string]interface{}{
			"system": system,
		},
	})
	if err != nil {
		return &RemoteSystem{}, err
	}

	res, err := readResponse(conn)
	if err != nil {
		return &RemoteSystem{}, err
	}

	machineData := res.Data[qpmd.MACHINE].(map[string]interface{})

	m := &Machine{
		Address:            remoteAddress,
		MachineId:          machineData[qpmd.MACHINE_ID].(string),
		MessageGatewayPort: machineData[qpmd.MESSAGE_GATEWAY_PORT].(uint16),
		GeneralPurposePort: machineData[qpmd.GP_GATEWAY_PORT].(uint16),
	}

	return &RemoteSystem{
		Address:   remoteAddress,
		Port:      res.Data[qpmd.PORT].(uint16),
		MachineId: m.MachineId,
		Machine:   m,
	}, nil
}
