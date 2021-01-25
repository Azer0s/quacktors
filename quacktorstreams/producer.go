package quacktorstreams

import "github.com/Azer0s/quacktors"

type Producer interface {
	Init() error
	SetTopic(topic string)
	Emit(message quacktors.Message)
}

type ProducerActor struct {
	Producer
}

func (p *ProducerActor) Init(ctx *quacktors.Context) {
	err := p.Producer.Init()
	if err != nil {
		panic(err)
	}
}

func (p *ProducerActor) Run(ctx *quacktors.Context, message quacktors.Message) {
	p.Producer.Emit(message)
}
