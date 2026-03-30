package driver

import (
	"fmt"
	"net"
)

// FreePort asks the OS for an available TCP port and returns it.
// Returns an error if no port can be allocated.
func FreePort() (int, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, fmt.Errorf("find free port: %w", err)
	}
	defer func() { _ = ln.Close() }()
	return ln.Addr().(*net.TCPAddr).Port, nil
}
