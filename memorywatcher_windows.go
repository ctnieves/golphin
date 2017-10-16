// +build windows

package golphin

import "bytes"

type MemoryMap map[MemoryAddress]MemoryValue

func (g *Golphin) FetchMemoryUpdates() error {
	platform := g.PlatformContainer.(*Win32Container)

	list := g.MemoryLocations.List()

	for _, a := range list {
		address := a.(MemoryAddress)
		emulator_val, err := g.ReadProcessMemory(address, 8)

		if err != nil {
			return err
		}

		if saved, ok := platform.MemoryReference[address]; ok {
			if bytes.Compare(saved, emulator_val) == 0 {
				platform.MemoryReference[address] = emulator_val
				g.MemoryUpdate <- MemoryPair{address, emulator_val}
			}
		} else {
			platform.MemoryReference[address] = emulator_val
			g.MemoryUpdate <- MemoryPair{address, emulator_val}
		}
	}

	return nil
}
