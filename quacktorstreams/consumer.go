package quacktorstreams

import (
	"github.com/Azer0s/quacktors"
	"github.com/opentracing/opentracing-go"
)

type StreamMessage struct {
	Bytes []byte
	Topic string
	Meta  interface{}
}

type Consumer interface {
	Init() error
	Subscribe(topic string) error
	NextMessage() (StreamMessage, error)
}

type ConsumerActor struct {
	topicMessageMap map[string]func([]byte) (quacktors.Message, error)
	topicHandlerMap map[string]*quacktors.Pid

	Consumer
}

func (c *ConsumerActor) Init(ctx *quacktors.Context) {
	err := c.Consumer.Init()
	if err != nil {
		panic(err)
	}

	c.topicMessageMap = make(map[string]func([]byte) (quacktors.Message, error))
	c.topicHandlerMap = make(map[string]*quacktors.Pid)

	ctx.Send(ctx.Self(), quacktors.EmptyMessage{})
}

func (c *ConsumerActor) Subscribe(topic string, handler *quacktors.Pid, conv func([]byte) (quacktors.Message, error)) error {
	c.topicMessageMap[topic] = conv
	c.topicHandlerMap[topic] = handler

	return c.Consumer.Subscribe(topic)
}

func (c *ConsumerActor) Run(ctx *quacktors.Context, message quacktors.Message) {
	defer func() {
		if r := recover(); r != nil {
			ctx.Quit()
		}

		ctx.Send(ctx.Self(), quacktors.EmptyMessage{})
	}()

	msg, err := c.Consumer.NextMessage()
	if err != nil {
		return
	}

	if conv, ok := c.topicMessageMap[msg.Topic]; ok {
		qm, err := conv(msg.Bytes)
		if err != nil {
			return
		}

		handler := c.topicHandlerMap[msg.Topic]

		context := quacktors.VectorContext("", opentracing.StartSpan("consumer"+msg.Topic))
		context.Send(handler, qm)
	}
}
