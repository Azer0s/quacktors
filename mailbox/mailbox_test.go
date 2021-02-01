package mailbox

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestMailbox(t *testing.T) {
	mb := New()

	in := mb.In()
	out := mb.Out()

	for i := 0; i < 10_000; i++ {
		in <- ""
	}

	c := 0
	for {
		<-out

		c++

		if c == 10_000 {
			select {
			case <-out:
				t.Fail()
			default:

			}

			return
		}
	}
}

func TestMailboxClose(t *testing.T) {
	mb := New()

	in := mb.In()
	out := mb.Out()

	for i := 0; i < 10_000; i++ {
		in <- ""
	}

	close(in)

	<-time.After(10 * time.Millisecond)

	_, ok := <-out
	assert.False(t, ok)
}
