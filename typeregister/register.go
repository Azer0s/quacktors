package typeregister

import "sync"

var typeRegistry = &sync.Map{}

func Store(name string, value interface{}) {
	typeRegistry.Store(name, value)
}

func Load(name string) (interface{}, bool) {
	return typeRegistry.Load(name)
}
