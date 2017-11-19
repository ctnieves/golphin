// +build !windows

package golphin

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"sync"

	"net"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Platform specific interface definitions are not required but help clear
// up the difference between the types across platforms
type UnixlikeDolphin interface {
	GolphinInterface

	SetPath(path string) error
	GetLocationBytes() []byte
}

type UnixlikeContainer struct {
	SocketPath   string
	Socket       *net.UnixConn
	SocketBuffer []byte
	SocketMutex  sync.Mutex
}

func (g *Golphin) Init() error {
	platform := &UnixlikeContainer{
		SocketBuffer: make([]byte, 9096),
		SocketMutex:  sync.Mutex{},
	}
	g.PlatformContainer = platform

	err := g.SetPath("/Users/christian/Desktop/FM/Dolphin.app/Contents/Resources/User")

	if g.DolphinPath == "" {
		return errors.New("Failed to initialize. DolphinPath not set. ")
	} else {
		err = g.CommitLocations()

		if err != nil {
			return err
		}
	}

	err = g.BindMemory()
	return err
}

func (g *Golphin) SetPath(path string) error {
	platform := g.PlatformContainer.(*UnixlikeContainer)

	if platform.SocketPath != "" {
		_ = os.Mkdir(platform.SocketPath, os.ModePerm)
	}

	ex, err := FilepathExists(path)

	if err != nil {
		return err
	}

	if ex {
		g.DolphinPath = path
		platform.SocketPath = filepath.Join(path, "MemoryWatcher/")

		_ = os.Mkdir(platform.SocketPath, os.ModePerm)
		return nil
	}

	return errors.New("The provided path does not exist. ")
}

func (g *Golphin) BindMemory() error {
	platform := g.PlatformContainer.(*UnixlikeContainer)

	if platform.SocketPath != "" {
		_ = os.Mkdir(platform.SocketPath, os.ModePerm)
	} else {
		return errors.New("Failed to create socket path. ")
	}

	p := filepath.Join(platform.SocketPath, "MemoryWatcher")

	syscall.Unlink(p)
	c, err := net.ListenUnixgram("unixgram", &net.UnixAddr{p, "unixgram"})

	if err != nil {
		return err
	}

	platform.Socket = c

	return nil
}

func (g *Golphin) ReadMemory() (err error) {
	if !g.Looping {
		close(g.MemoryUpdate)
		return nil
	}
	platform := g.PlatformContainer.(*UnixlikeContainer)

	platform.SocketMutex.Lock()
	n, err := (*platform.Socket).Read(platform.SocketBuffer[:])
	platform.SocketMutex.Unlock()

	if err != nil {
		return err
	}

	s := strings.Split(string(platform.SocketBuffer[0:n]), "\n")
	padded := fmt.Sprintf("%08s", strings.Replace(s[1], "\x00", "", -1))
	decoded, err := hex.DecodeString(padded)

	if err != nil {
		platform.Socket.Close()
		log.Fatalln(err)
		return
	} else {
		g.MemoryUpdate <- MemoryPair{MemoryAddress(s[0]), decoded}
	}

	return
}

func (g *Golphin) CommitLocations() error {
	platform := g.PlatformContainer.(*UnixlikeContainer)

	locations_name := filepath.Join(platform.SocketPath, "Locations.txt")
	file, err := os.OpenFile(string(locations_name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)

	defer file.Close()

	if err != nil {
		return err
	}

	_, err = file.Write(g.GetLocationsBytes())

	return err
}
