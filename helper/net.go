package helper

import (
	"fmt"
	"net"
)

func NextUsefulPort(port int) int {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return NextUsefulPort(port + 1)
	}
	defer func() { _ = listener.Close() }()
	return port
}
