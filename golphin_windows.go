// +build windows

package golphin

import (
	"C"
	"errors"
	"strings"
	"bytes"
	"unicode/utf16"
	"unsafe"
	"strconv"
	"sync"

	j32 "github.com/JamesHovious/w32"
	w32 "golang.org/x/sys/windows"
)

// Platform specific interface definitions are not required but help clear
// up the difference between the types across platforms
type Win32Dolphin interface {
	GolphinInterface

	FetchMemoryUpdates()
	GetProcessByName(string) (w32.Handle, error)
	ReadProcessMemory(string) ([]byte, error)
}

type Win32Container struct {
	MemoryReference MemoryMap
	ReferenceMutex *sync.Mutex
	EmulatorProc    w32.Handle
}

const BASE_ADDRESS = 0x7FFF0000

const (
	ALL_ACCESS                uint32 = 0x001F0FFF
	TERMINATE                        = 0x00000001
	CREATE_THREAD                    = 0x00000002
	VM_OPERATION                     = 0x00000008
	VM_READ                          = 0x00000010
	VM_WRITE                         = 0x00000020
	DUPLICATE_HANDLE                 = 0x00000040
	CREATE_PROCESS                   = 0x000000080
	SET_QUOTA                        = 0x00000100
	SET_INFORMATION                  = 0x00000200
	QUERY_INFORMATION                = 0x00000400
	QUERY_LIMITED_INFORMATION        = 0x00001000
	SYNCHRONIZE                      = 0x00100000
)

func (g *Golphin) Init() error {
	g.PlatformContainer = &Win32Container{
		MemoryReference: make(MemoryMap, 500),
		ReferenceMutex:  &sync.Mutex{},
	}
	err := g.BindMemory()

	return err
}

func (g *Golphin) BindMemory() error {
	platform := g.PlatformContainer.(*Win32Container)

	p, err := GetProcessByName("Dolphin.exe")
	platform.EmulatorProc = p

	return err
}

func (g *Golphin) ReadMemory() (err error) {
	return g.FetchMemoryUpdates()
}

func (g *Golphin) CommitLocations() error {
	return nil
}

func GetProcessByName(name string) (w32.Handle, error) {
	snapshot, err := w32.CreateToolhelp32Snapshot(w32.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return w32.InvalidHandle, err
	}

	defer w32.CloseHandle(snapshot)
	var process w32.ProcessEntry32
	process.Size = uint32(unsafe.Sizeof(process))

	if err = w32.Process32First(snapshot, &process); err != nil {
		return w32.InvalidHandle, err
	}

	for {
		name_bytes := []byte(string(utf16.Decode(process.ExeFile[:])))
		process_name := string(bytes.Trim(name_bytes, "\x00"))

		if process_name == name {
			return w32.OpenProcess(ALL_ACCESS, false, process.ProcessID)
		}

		err = w32.Process32Next(snapshot, &process)

		if err != nil {
			return w32.InvalidHandle, errors.New("Failed to find process. Is Dolphin running?")
		}
	}

	return w32.InvalidHandle, errors.New("Failed to find process. Is Dolphin running?")
}

func (g *Golphin) ReadProcessMemory(address string, size uint) ([]byte, error) {
	platform := g.PlatformContainer.(*Win32Container)
	var address_uint uint32 = BASE_ADDRESS
	// if the address is a pointer
	if strings.Contains(address, " ") {
		split := strings.Split(address, " ")
		base, err := strconv.ParseUint(split[0], 16, 32)
		offset, err := strconv.ParseUint(split[1], 16, 32)

		if err != nil {
			return nil, err
		}

		base_val := uint32(base)
		offset_val := uint32(offset)

		address_uint += base_val + offset_val
	} else {
		temp, err := strconv.ParseUint(address, 16, 32)
		if err != nil {
			return nil, err
		}
		address_uint += uint32(temp)
	}


	memory_bytes, err := j32.ReadProcessMemory(j32.HANDLE(platform.EmulatorProc), address_uint, size)

	return memory_bytes, err
}
