package proxy

import (
	"context"
	"net"
)

type ListenerFactory interface {
	Create(context.Context) (net.Listener, error)
}
