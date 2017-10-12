package golphin

import (
	"io"
	"os"
	"path/filepath"
)

func (g *Golphin) Subscribe(address string) error {
	g.MemoryLocations.Add(address)
	return nil
}

func (g *Golphin) Unsubscribe(address string) error {
	g.MemoryLocations.Remove(address)
	return nil
}

func (g *Golphin) WriteLocations() error {
	locations_name := filepath.Join(g.SocketPath, "Locations.txt")
	file, err := os.OpenFile(string(locations_name), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	defer file.Close()

	if err != nil {
		return err
	}

	_, err = file.Write(g.GetLocationsBytes())

	return err
}

func (g *Golphin) GetLocationsBytes() []byte {
	list := g.MemoryLocations.List()
	locations := ""
	for _, addy := range list {
		locations += addy.(string) + "\n"
	}
	return []byte(locations)
}

func FilepathExists(path string) (bool, error) {
	_, err := os.Stat(path)

	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}
