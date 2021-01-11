package genserver

import (
	"encoding/gob"
	"github.com/Azer0s/quacktors"
)

func init() {
	gob.RegisterName(callMessage{}.Type(), callMessage{})
	gob.RegisterName(castMessage{}.Type(), castMessage{})
	gob.RegisterName(ReceivedMessage{}.Type(), ReceivedMessage{})
	gob.RegisterName(ResponseMessage{}.Type(), ResponseMessage{})
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

type ReceivedMessage struct {
}

func (r ReceivedMessage) Type() string {
	return "ReceivedMessage"
}

type ResponseMessage struct {
	quacktors.Message
	Error error
}

func (r ResponseMessage) Type() string {
	return "ResponseMessage"
}
