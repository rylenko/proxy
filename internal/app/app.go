package app

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/rylenko/proxy/internal/proxy"
)

// TODO: Make logs more readable
// TODO: Add context
// TODO: tests?
func Run(listenerFactory proxy.ListenerFactory, handler proxy.Handler) error {
	listener, err := listenerFactory.Create()
	if err != nil {
		return fmt.Errorf("listenerFactory.Create(): %w", err)
	}
	defer listener.Close()
	log.Printf("Listening on %s\n", listener.Addr())

	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Listener %s failed to accept connection: %v\n", listener.Addr(), err)
			continue
		}
		log.Printf("Listener %s accepted a connection %s\n", listener.Addr(), conn.RemoteAddr())

		wg.Add(1)
		go func(conn net.Conn) {
			defer wg.Done()
			defer conn.Close()

			if err := handler.Handle(conn); err != nil {
				log.Printf("Listener %s failed to handle %s: %v\n", listener.Addr(), conn.RemoteAddr(), err)
				return
			}

			log.Printf("Listener %s done with a connection %s\n", listener.Addr(), conn.RemoteAddr())
		}(conn)
	}

	wg.Wait()

	return nil
}
