package quacktorstreams

import (
	"encoding/json"
	"github.com/Azer0s/quacktors"
)

type testProducer struct {
	topic string
}

func (p *testProducer) Init() error {
	return nil
}

func (p *testProducer) SetTopic(topic string) {
	p.topic = topic
}

func (p *testProducer) Emit(message quacktors.Message) {
	b, _ := json.Marshal(message)

	testChan <- StreamMessage{
		Bytes: b,
		Topic: p.topic,
		Meta:  nil,
	}
}
