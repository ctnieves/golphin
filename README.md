# golphin

A Go package for interacting with memory in the Dolphin Emulator.

## Installation
	go get -u github.com/ctnieves/golphin

## Usage

### Setup
```go
import "github.com/ctnieves/golphin"

dolphin := golphin.New()
// optionally prompt the user for their path.
err := dolphin.SetPath("/Users/christian/Dolphin.app/Contents/Resources/User")
err = Dolphin.Init()
// handle error
Dolphin.Subscribe("YOUR_MEMORY_ADDRESS")
Dolphine.WriteLocations() // required to "sync" addresses with Dolphin

```

More usage coming...maybe?

## Notes
Only supports Linux/OSX due to restrictions in Dolphin's MemoryWatcher API. However this will be updated by using win32 libs soon.
