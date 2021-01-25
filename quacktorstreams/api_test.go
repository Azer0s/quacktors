package quacktorstreams

/*func TestStream(t *testing.T) {
	context := quacktors.RootContext()

	testTopic := "test"

	consumer, _ := NewConsumer(&KafkaConsumer{
		Config: &kafka.ConfigMap{
			"bootstrap.servers": "localhost",
			"group.id":          "default",
		},
	})

	producer := NewProducer(&KafkaProducer{
		Config: &kafka.ConfigMap{
			"bootstrap.servers": "localhost",
		},
	}, testTopic)

	pid := quacktors.Spawn(func(ctx *quacktors.Context, message quacktors.Message) {
		if message.(quacktors.GenericMessage).Value == "exit" {
			os.Exit(0)
		}

		fmt.Println(message)
	})

	_ = consumer.Subscribe("test", pid, func(bytes []byte) (quacktors.Message, error) {
		val := quacktors.GenericMessage{}
		err := json.Unmarshal(bytes, &val)
		return val, err
	})

	context.Send(producer, quacktors.GenericMessage{Value: 1})
	context.Send(producer, quacktors.GenericMessage{Value: 2})
	context.Send(producer, quacktors.GenericMessage{Value: 3})
	context.Send(producer, quacktors.GenericMessage{Value: "exit"})

	context.Send(producer, quacktors.PoisonPill{})

	quacktors.Run()
}
*/
