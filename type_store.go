package quacktors

import "sync"

var types = make(map[string]Message)
var typesMu = &sync.RWMutex{}

func storeType(message Message) {
	typesMu.Lock()
	defer typesMu.Unlock()

	types[message.Type()] = message
}

func getType(name string) Message {
	typesMu.RLock()
	defer typesMu.RUnlock()

	return create(types[name])
}