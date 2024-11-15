package app

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/rylenko/proxy/internal/proxy"
)

func Run(listenerFactory proxy.ListenerFactory, handler proxy.Handler) error {
	listener, err := listenerFactory.Create()
	if err != nil {
		return fmt.Errorf("create listener: %w", err)
	}
	defer listener.Close()
	log.Printf("Listening on %s\n", listener.Addr())

	var wg sync.WaitGroup

	for {
		acceptConns(&wg, listener, handler)
	}

	wg.Wait()
	return nil
}

func acceptConns(wg *sync.WaitGroup, listener net.Listener, handler proxy.Handler) {
	conn, err := listener.Accept()
	if err != nil {
		log.Printf("Listener %s failed to accept connection: %v\n", listener.Addr(), err)
		return
	}
	log.Printf("Listener %s accepted a connection %s\n", listener.Addr(), conn.RemoteAddr())

	wg.Add(1)
	go handleConn(wg, listener.Addr(), conn, handler)
}

func handleConn(wg *sync.WaitGroup, listenerAddr net.Addr, conn net.Conn, handler proxy.Handler) {
	defer wg.Done()
	defer conn.Close()

	if err := handler.Handle(conn); err != nil {
		log.Printf("Listener %s failed to handle %s: %v\n", listenerAddr, conn.RemoteAddr(), err)
		return
	}
	log.Printf("Listener %s done with a connection %s\n", listenerAddr, conn.RemoteAddr())
}
