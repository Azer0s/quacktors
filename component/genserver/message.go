package genserver

import (
	"github.com/Azer0s/quacktors"
	"github.com/Azer0s/quacktors/typeregister"
)

func init() {
	typeregister.Store(callMessage{}.Type(), callMessage{})
	typeregister.Store(castMessage{}.Type(), castMessage{})
	typeregister.Store(ReceivedMessage{}.Type(), ReceivedMessage{})
	typeregister.Store(ResponseMessage{}.Type(), ResponseMessage{})
}

type callMessage struct {
	sender  *quacktors.Pid
	message quacktors.Message
}

func (c callMessage) Type() string {
	return "CallMessage"
}

type castMessage struct {
	sender  *quacktors.Pid
	message quacktors.Message
}

func (c castMessage) Type() string {
	return "CastMessage"
}

//The ReceivedMessage struct is the acknowledgement
//a Cast operation returns when the GenServer has
//received a message.
type ReceivedMessage struct {
}

//Type of ReceivedMessage returns "ReceivedMessage"
func (r ReceivedMessage) Type() string {
	return "ReceivedMessage"
}

//The ResponseMessage struct is returned as the result
//type of a Call operation on a GenServer.
type ResponseMessage struct {
	quacktors.Message
	Error error
}

//Type of ResponseMessage returns "ResponseMessage"
func (r ResponseMessage) Type() string {
	return "ResponseMessage"
}
