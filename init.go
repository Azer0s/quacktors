package quacktors

import (
	"encoding/gob"
	"fmt"
	"github.com/Azer0s/qpmd"
	"github.com/Azer0s/quacktors/config"
	"github.com/Azer0s/quacktors/logging"
	"github.com/vmihailenco/msgpack/v5"
	"net"
)

var messageGatewayPort = uint16(0)
var gpGatewayPort = uint16(0)

var logger logging.Logger
var qpmdPort uint16

func initQuacktorSystems() {
	logger = config.GetLogger()
	qpmdPort = config.GetQpmdPort()

	initializeGateways()
	initializeQpmdConnection()
	initializeBuiltInMessages()
}

func initializeGateways() {
	var err error

	messageGatewayPort, err = startMessageGateway()
	if err != nil {
		logger.Fatal("there was an error while starting the message gateway",
			"error", err)
	}

	gpGatewayPort, err = startGeneralPurposeGateway()
	if err != nil {
		logger.Fatal("there was an error while starting the general purpose gateway",
			"error", err)
	}
}

func initializeQpmdConnection() {
	failIfConnectionError := func(err error) {
		if err != nil {
			panic("Couldn't connect to qpmd! Is qpmd running?")
		}
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

func initializeBuiltInMessages() {
	gob.RegisterName(Pid{}.Type(), Pid{})
	gob.RegisterName(DownMessage{}.Type(), DownMessage{})
	gob.RegisterName(PoisonPill{}.Type(), PoisonPill{})
	gob.RegisterName(GenericMessage{}.Type(), GenericMessage{})
	gob.RegisterName(DisconnectMessage{}.Type(), DisconnectMessage{})
	gob.RegisterName(KillMessage{}.Type(), KillMessage{})
}
