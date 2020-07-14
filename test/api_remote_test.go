package test

import (
	"fmt"
	"github.com/Azer0s/quacktors"
	"testing"
)

func TestGatewayConnection(t *testing.T) {
	quacktors.StartGateway(5521)
	foo := quacktors.NewSystem("foo")
	fmt.Println(foo)

	node, err := quacktors.Connect("foo@127.0.0.1:5521")

	if err != nil {
		panic(err)
	}

	fmt.Println(node)
}
