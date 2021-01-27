package mailbox

import (
	"container/list"
)

func New() *Mailbox {
	mb := &Mailbox{
		inChan:  make(chan interface{}),
		outChan: make(chan interface{}),
		queue:   list.New(),
	}

	mb.start()

	return mb
}

type Mailbox struct {
	inChan  chan interface{}
	outChan chan interface{}
	queue   *list.List
}

func (mb *Mailbox) In() chan<- interface{} {
	return mb.inChan
}

func (mb *Mailbox) Out() <-chan interface{} {
	return mb.outChan
}

func (mb *Mailbox) Len() int {
	return mb.queue.Len()
}

func (mb *Mailbox) start() {
	getOutCh := func() chan<- interface{} {
		if mb.queue.Len() == 0 {
			return nil
		}

		return mb.outChan
	}

	getCurVal := func() interface{} {
		if mb.queue.Len() == 0 {
			return nil
		}

		return mb.queue.Front().Value
	}

	go func() {
		for {
			select {
			case elem, ok := <-mb.inChan:
				if ok {
					mb.queue.PushBack(elem)
				} else {
					close(mb.outChan)
					return
				}
			case getOutCh() <- getCurVal():
				mb.queue.Remove(mb.queue.Front())
			}
		}
	}()
}
