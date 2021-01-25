package quacktorstreams

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
	return msg, nil
}
