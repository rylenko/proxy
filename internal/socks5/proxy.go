package socks5

import "github.com/rylenko/proxy/internal/proxy"

type Proxy struct {
	port int
}

func NewProxy(port int) *Proxy {
	return &Proxy{
		port: port,
	}
}

// Ensure that SOCKS5 proxy implements proxy interface.
var _ proxy.Proxy = (*Proxy)(nil)
