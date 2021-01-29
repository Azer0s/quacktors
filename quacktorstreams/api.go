package quacktorstreams

import "github.com/Azer0s/quacktors"

//NewConsumer creates a new ConsumerActor by a consumer implementation
//and returns both a pointer to the ConsumerActor itself and the PID
//of the consumer. The pointer to the ConsumerActor can be used to subscribe
//to topics of the stream.
func NewConsumer(consumer Consumer) (*ConsumerActor, *quacktors.Pid) {
	actor := &ConsumerActor{
		Consumer: consumer,
	}
	pid := quacktors.SpawnStateful(actor)

	return actor, pid
}

//NewProducer creates a new producer actor by a consumer implementation
//and returns the PID of the ProducerActor. When a message is sent to
//the ProducerActor, it is automatically forwarded (i.e. published)
//to the provided topic in the stream.
func NewProducer(producer Producer, topic string) *quacktors.Pid {
	actor := &ProducerActor{
		producer,
	}
	pid := quacktors.SpawnStateful(actor)

	actor.SetTopic(topic)

	return pid
}
