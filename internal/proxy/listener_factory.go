package proxy

import "net"

type ListenerFactory interface {
	Create() (net.Listener, error)
}
