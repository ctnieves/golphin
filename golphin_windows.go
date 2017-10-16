// +build windows

package golphin

import (
	"C"
	"errors"
	//"fmt"
	"unsafe"
	"bytes"
	"unicode/utf16"
	//"log"
	//"strings"

	w32 "golang.org/x/sys/windows"
	j32 "github.com/JamesHovious/w32"
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
	EmulatorProc w32.Handle
}

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
		MemoryReference: make(MemoryMap, 1),
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
		//fmt.Println(process_name + " IS ",[]byte(process_name))

		if  process_name == name {
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
	return j32.ReadProcessMemory(platform.EmulatorProc, address, size)
}

