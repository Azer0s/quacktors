package encoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type test struct {
	Foo string
	Bar float32
}

func TestEncode(t *testing.T) {
	encoder := NewMsgpackEncoder()

	val := test{
		Foo: "Gello",
		Bar: 12.453,
	}

	encoder.RegisterType("test", val)
	b, _ := encoder.Encode("test", val)

	valDec, _ := encoder.Decode("test", b)

	assert.Equal(t, val, valDec)
}

type empty struct {
}

func TestEncodeEmpty(t *testing.T) {
	encoder := NewMsgpackEncoder()

	val := empty{}

	encoder.RegisterType("empty", val)
	b, _ := encoder.Encode("empty", val)

	valDec, _ := encoder.Decode("empty", b)

	assert.Equal(t, val, valDec)
}
