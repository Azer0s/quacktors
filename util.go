package quacktors

import (
	"encoding/json"
	"github.com/Azer0s/qpmd"
	"github.com/gofrs/uuid"
	"github.com/opentracing/opentracing-go"
	"github.com/vmihailenco/msgpack/v5"
	"net"
	"strings"
	"time"
)

type quitAction struct{}

type remoteMonitorTuple struct {
	From *Pid
	To   *Pid
}

type remoteMessageTuple struct {
	To      *Pid
	Message Message
	opentracing.SpanContext
}

func try(err error) {
	if err != nil {
		panic(err)
	}
}

func uuidString() string {
	u, err := uuid.NewV4()

	if err != nil {
		//This really shouldn't EVER happen and if it does you're totally screwed anyways... ¯\_(ツ)_/¯
		panic(err)
	}

	return strings.ReplaceAll(u.String(), "-", "")
}

func parsePid(rawData map[string]interface{}) (*Pid, error) {
	b, err := json.Marshal(rawData)

	if err != nil {
		return nil, err
	}

	pid := &Pid{}
	err = json.Unmarshal(b, pid)

	if err != nil {
		return nil, err
	}

	return pid, nil
}

func parseMachine(rawData map[string]interface{}) (*Machine, error) {
	b, err := json.Marshal(rawData)
	if err != nil {
		return nil, err
	}

	m := &Machine{}
	err = json.Unmarshal(b, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func sendRequest(conn net.Conn, req qpmd.Request) error {
	b, err := msgpack.Marshal(req)

	if err != nil {
		return err
	}

	_, err = conn.Write(b)

	if err != nil {
		return err
	}

	return nil
}

func readResponse(conn net.Conn) (qpmd.Response, error) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)

	if err != nil {
		return qpmd.Response{}, err
	}

	res := qpmd.Response{}
	err = msgpack.Unmarshal(buf[:n], &res)

	if err != nil {
		return qpmd.Response{}, err
	}

	return res, nil
}

func readRequest(conn net.Conn) (qpmd.Request, error) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if n == 0 || err != nil {
		return qpmd.Request{}, err
	}

	req := qpmd.Request{}
	err = msgpack.Unmarshal(buf[:n], &req)

	if err != nil {
		return qpmd.Request{}, err
	}

	return req, nil
}

func sendResponse(client net.Conn, response qpmd.Response) error {
	response.Data[qpmd.TIMESTAMP] = time.Now().Unix()

	b, err := msgpack.Marshal(response)

	if err != nil {
		return err
	}

	_, err = client.Write(b)

	if err != nil {
		return err
	}

	return nil
}

func writeError(client net.Conn, err error) error {
	return sendResponse(client, qpmd.Response{
		ResponseType: qpmd.RESPONSE_ERROR,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	})
}

func writeOk(client net.Conn, data map[string]interface{}) error {
	return sendResponse(client, qpmd.Response{
		ResponseType: qpmd.RESPONSE_OK,
		Data:         data,
	})
}
