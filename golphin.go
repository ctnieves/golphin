package golphin

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"

	"gopkg.in/fatih/set.v0"
)

type Golphin struct {
	DolphinPath string

	SocketPath   string
	Socket       *net.UnixConn
	SocketBuffer []byte
	SocketMutex  sync.Mutex

	MemoryLocations *set.Set
	MemoryUpdate    chan MemoryPair

	Looping bool
}

type MemoryPair struct {
	Address string
	Value   []byte
}

func New() *Golphin {
	g := Golphin{
		SocketBuffer:    make([]byte, 9096),
		SocketMutex:     sync.Mutex{},
		MemoryLocations: set.New(),
		MemoryUpdate:    make(chan MemoryPair),
		Looping:         true,
	}
	return &g
}

func (g *Golphin) Init() error {
	if g.DolphinPath == "" || g.SocketPath == "" {
		return errors.New("Failed to initialize. DolphinPath not set. ")
	} else {
		err := g.WriteLocations()

		if err != nil {
			return err
		}
	}

	if g.SocketPath != "" {
		_ = os.Mkdir(g.SocketPath, os.ModePerm)
	}

	err := g.BindSocket()

	return err
}

func (g *Golphin) SetPath(path string) error {
	ex, err := FilepathExists(path)

	if err != nil {
		return err
	}

	if ex {
		g.DolphinPath = path
		g.SocketPath = filepath.Join(path, "MemoryWatcher/")

		_ = os.Mkdir(g.SocketPath, os.ModePerm)
		return nil
	}

	return errors.New("The provided path does not exist. ")
}

func (g *Golphin) BindSocket() error {
	p := filepath.Join(g.SocketPath, "MemoryWatcher")

	syscall.Unlink(p)
	c, err := net.ListenUnixgram("unixgram", &net.UnixAddr{p, "unixgram"})

	if err != nil {
		return err
	}

	g.Socket = c

	return nil
}

func (g *Golphin) ReadSocket() (err error) {
	err = nil

	g.SocketMutex.Lock()
	n, err := (*g.Socket).Read(g.SocketBuffer[:])
	g.SocketMutex.Unlock()

	if err != nil {
		return err
	}

	s := strings.Split(string(g.SocketBuffer[0:n]), "\n")
	padded := fmt.Sprintf("%08s", strings.Replace(s[1], "\x00", "", -1))
	decoded, err := hex.DecodeString(padded)

	if err != nil {
		g.Socket.Close()
		log.Fatalln(err)
		return
	} else {
		g.MemoryUpdate <- MemoryPair{s[0], decoded}
	}

	return
}
