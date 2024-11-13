package socks5

import (
	"context"
	"net"

	"github.com/rylenko/proxy/internal/proxy"
)

type Proxy struct {
	port int
}

func (p *Proxy) Handle(ctx context.Context, conn net.Conn) error {
	return nil
}

func (p *Proxy) Listen(ctx context.Context) (net.Listener, error) {
	return nil, nil
}

func NewProxy(port int) *Proxy {
	return &Proxy{
		port: port,
	}
}

// Ensure that SOCKS5 proxy implements proxy interface.
var _ proxy.Proxy = (*Proxy)(nil)
