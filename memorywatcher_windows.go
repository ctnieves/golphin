// +build windows

package golphin

import (
	"bytes"
	"fmt"
	"log"
)

type MemoryMap map[MemoryAddress]MemoryValue

func (g *Golphin) FetchMemoryUpdates() error {
	platform := g.PlatformContainer.(*Win32Container)

	list := g.MemoryLocations.List()

	for _, a := range list {
		if address, ok := a.(MemoryAddress); ok {
			address_string := fmt.Sprintf("%v", a)
			emulator_val, err := g.ReadProcessMemory(address_string, 8)

			if err != nil || emulator_val == nil {
				return err
			}

			platform.ReferenceMutex.Lock()
			saved, ok := platform.MemoryReference[address]
			platform.ReferenceMutex.Unlock()

			// this address has been stored in the map before
			if ok {
				if bytes.Compare(saved, emulator_val) != 0 {
					g.UpdateMemoryReference(address, emulator_val)
				} else {
					// value hasn't changed, don't update
				}
			} else {
				// this will only get ran once.
				g.UpdateMemoryReference(address, emulator_val)
			}
		}
	}

	return nil
}

func (g *Golphin) UpdateMemoryReference(address MemoryAddress, value []byte) {

	log.Printf("Update: %v\n", address)
	platform := g.PlatformContainer.(*Win32Container)

	platform.ReferenceMutex.Lock()
	platform.MemoryReference[address] = value
	platform.ReferenceMutex.Unlock()

	g.MemoryUpdate <- MemoryPair{address, value}
}
