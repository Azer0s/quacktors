package quacktorstreams

import "github.com/Azer0s/quacktors"

func NewConsumer(consumer Consumer) (*ConsumerActor, *quacktors.Pid) {
	actor := &ConsumerActor{
		Consumer: consumer,
	}
	pid := quacktors.SpawnStateful(actor)

	return actor, pid
}

func NewProducer(producer Producer, topic string) *quacktors.Pid {
	actor := &ProducerActor{
		producer,
	}
	pid := quacktors.SpawnStateful(actor)

	actor.SetTopic(topic)

	return pid
}
