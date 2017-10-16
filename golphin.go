package golphin

import (
	"gopkg.in/fatih/set.v0"
)

type GolphinInterface interface {
	Init() error

	BindMemory() error
	ReadMemory() error

	Subscribe(string) error
	Unsubscribe(string) error

	CommitLocations() error
}

type Golphin struct {
	DolphinPath string

	MemoryLocations *set.Set
	MemoryUpdate    chan MemoryPair

	Looping bool

	PlatformContainer interface{}
}

type MemoryAddress string
type MemoryValue []byte

type MemoryPair struct {
	Address MemoryAddress
	Value   MemoryValue
}

func New() *Golphin {
	return &Golphin{
		MemoryLocations: set.New(),
		MemoryUpdate:    make(chan MemoryPair),
		Looping:         true,
	}
}
