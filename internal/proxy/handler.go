package proxy

import "net"

type Handler interface {
	Handle(net.Conn) error
}
