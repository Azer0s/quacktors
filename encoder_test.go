package quacktors

import (
	"github.com/Azer0s/quacktors/typeregister"
	"github.com/stretchr/testify/assert"
	"testing"
)

type test struct {
	Foo string
	Bar float32
}

func TestEncode(t *testing.T) {
	val := test{
		Foo: "Hello",
		Bar: 12.453,
	}

	typeregister.Store("test", val)
	b, _ := encodeValue("test", val)

	valDec, _ := decodeValue("test", b)

	assert.Equal(t, val, valDec)
}

type empty struct {
}

func TestEncodeEmpty(t *testing.T) {
	val := empty{}

	typeregister.Store("empty", val)
	b, _ := encodeValue("empty", val)

	valDec, _ := decodeValue("empty", b)

	assert.Equal(t, val, valDec)
}

type nestingStruct struct {
	Foo string
	Val nestedStruct
}

type nestedStruct struct {
	Bar float32
}

func TestNestedEncode(t *testing.T) {
	val := nestingStruct{
		Foo: "Hello",
		Val: nestedStruct{Bar: 12.453},
	}

	typeregister.Store("test", val)
	b, _ := encodeValue("test", val)

	valDec, _ := decodeValue("test", b)

	assert.Equal(t, val, valDec)
}
