package encoding

type MessageEncoder interface {
	Encode(messageType string, value interface{}) ([]byte, error)
	Decode(messageType string, buffer []byte) (interface{}, error)
	RegisterType(messageType string, value interface{})
}
