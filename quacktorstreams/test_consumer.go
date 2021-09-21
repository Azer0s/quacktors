package quacktorstreams

import (
	"encoding/json"
	"github.com/Azer0s/quacktors"
)

var testChan = make(chan StreamMessage)

type testConsumer struct {
}

func (t *testConsumer) Init() error {
	return nil
}

func (t *testConsumer) Subscribe(topic string) error {
	return nil
}

func (t *testConsumer) NextMessage() (StreamMessage, error) {
	msg := <-testChan

	val := quacktors.GenericMessage{}
	_ = json.Unmarshal(msg.Bytes, &val)

	if val.Value == "exit" {
		panic("")
	}

	return msg, nil
}
