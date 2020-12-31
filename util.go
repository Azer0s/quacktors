package quacktors

import (
	"github.com/Azer0s/qpmd"
	uuid "github.com/satori/go.uuid"
	"github.com/vmihailenco/msgpack/v5"
	"net"
	"reflect"
	"strings"
	"time"
)

type remoteMonitorTuple struct {
	from *Pid
	to   *Pid
}

func try(err error) {
	if err != nil {
		panic(err)
	}
}

func uuidString() string {
	return strings.ReplaceAll(uuid.NewV4().String(), "-", "")
}

func createFromTemplateMessage(from Message) Message {
	t := reflect.ValueOf(from).Elem()
	typ := t.Type()
	ms := (reflect.New(typ).Elem()).Interface().(Message)

	return ms
}

func readRequest(conn net.Conn) (qpmd.Request, error) {
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if n == 0 || err != nil {
		return qpmd.Request{}, err
	}

	req := qpmd.Request{}
	err = msgpack.Unmarshal(buf[:n], &req)

	return req, nil
}

func writeResponse(client net.Conn, response qpmd.Response) error {
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
	return writeResponse(client, qpmd.Response{
		ResponseType: qpmd.RESPONSE_ERROR,
		Data: map[string]interface{}{
			"error": err.Error(),
		},
	})
}

func writeOk(client net.Conn, data map[string]interface{}) error {
	return writeResponse(client, qpmd.Response{
		ResponseType: qpmd.RESPONSE_OK,
		Data:         data,
	})
}
