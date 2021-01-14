package quacktors

import "github.com/opentracing/opentracing-go"

type localMessage struct {
	message     Message
	spanContext opentracing.SpanContext
}

//The Message interface defines all methods a struct has
//to implement so it can be sent around by actors.
type Message interface {
	//Type returns the the "name" of the Message.
	//This is used to identify and correctly unmarshal
	//messages from remote actors.
	Type() string
}

//The DownMessage is sent to a monitoring Actor whenever
//a monitored Actor goes down.
type DownMessage struct {
	//Who is the PID of the Actor that went down.
	Who *Pid
}

//Type of DownMessage returns "DownMessage"
func (d DownMessage) Type() string {
	return "DownMessage"
}

//A PoisonPill can be sent to an Actor to kill it without
//aborting current (or already queued) messages. Instead,
//it is enqueued into the actors mailbox and when the Actor
//gets to the PoisonPill Message, gracefully shuts it down.
//A PoisonPill never gets passed on to the Run function.
//Instead, it just calls Context.Quit on the current actors
//Context.
type PoisonPill struct {
	//PoisonPill message to kill actor
}

//Type of PoisonPill returns "PoisonPill"
func (p PoisonPill) Type() string {
	return "PoisonPill"
}

//The GenericMessage carries a single Value of type interface{}.
type GenericMessage struct {
	//Value is the value a GenericMessage carries.
	Value interface{}
}

//Type of GenericMessage returns "GenericMessage"
func (g GenericMessage) Type() string {
	return "GenericMessage"
}

//The EmptyMessage struct is an empty message without any
//semantic meaning (i.e. it literally doesn't do anything).
type EmptyMessage struct {
}

//Type of EmptyMessage returns "EmptyMessage"
func (e EmptyMessage) Type() string {
	return "EmptyMessage"
}

//A KillMessage can be sent to an Actor to ask it to shut down.
//It is entirely semantic, meaning this will not shut the Actor
//down automatically. Instead, the Actor should clean up whatever
//it is doing gracefully and then call Context.Quit itself.
type KillMessage struct {
	//KillMessage to signal to actor it should go down
}

//Type of KillMessage returns "KillMessage"
func (k KillMessage) Type() string {
	return "KillMessage"
}

//The DisconnectMessage is sent to a monitoring Actor whenever
//a monitored Machine connection goes down.
type DisconnectMessage struct {
	//MachineId is the ID of the Machine that disconnected.
	MachineId string
	//Address is the remote address of the Machine.
	Address string
}

//Type of DisconnectMessage returns "DisconnectMessage"
func (d DisconnectMessage) Type() string {
	return "DisconnectMessage"
}
