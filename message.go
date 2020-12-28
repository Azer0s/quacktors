package quacktors

type Message interface {
	Serialize() string
	Deserialize(string) Message
	Type() string
}
