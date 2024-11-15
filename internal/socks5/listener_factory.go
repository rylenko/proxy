package socks5

import (
	"fmt"
	"net"

	"github.com/rylenko/proxy/internal/proxy"
)

type ListenerFactory struct {
	port int
}

func (f *ListenerFactory) Create() (net.Listener, error) {
	addrStr := fmt.Sprintf(":%d", f.port)

	listener, err := net.Listen("tcp", addrStr)
	if err != nil {
		return nil, fmt.Errorf("Listen(\"%s\"): %w", addrStr, err)
	}

	return listener, nil
}

func NewListenerFactory(port int) *ListenerFactory {
	return &ListenerFactory{
		port: port,
	}
}

// Ensure that SOCKS5 listener factory implements listener factory interface.
var _ proxy.ListenerFactory = (*ListenerFactory)(nil)
