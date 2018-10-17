package registry

import (
	"sync"
)

// Service encapsulates everything necessary to execute a step against a target.
type Service interface {
	Execute() (err error)
	UpdateRequest(values map[string]interface{}) (err error)
}

// Factory is a factory which produces services
type Factory interface {
	Make(name string, settings map[string]interface{}) (Service, error)
}

var (
	mutex    sync.RWMutex
	registry = make(map[string]Factory, 32)
)

// Register registers a service factory
func Register(t string, factory Factory) {
	mutex.Lock()
	defer mutex.Unlock()
	registry[t] = factory
}

// Lookup finds a service factory
func Lookup(t string) Factory {
	mutex.RLock()
	defer mutex.RUnlock()
	return registry[t]
}
