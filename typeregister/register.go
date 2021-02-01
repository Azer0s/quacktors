package typeregister

import "sync"

var typeRegistry = &sync.Map{}

//Store stores a type by name.
func Store(name string, value interface{}) {
	typeRegistry.Store(name, value)
}

//Load loads a type by name.
func Load(name string) (interface{}, bool) {
	return typeRegistry.Load(name)
}
