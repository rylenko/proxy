package proxy

import (
	"context"
	"net"
)

type Proxy interface {
	Listen(context.Context) (net.Listener, error)
	Handle(context.Context, net.Conn) error
}
