package app

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/rylenko/proxy/internal/proxy"
)

func Run(ctx context.Context, listenerFactory proxy.ListenerFactory, handler proxy.Handler) error {
	listener, err := listenerFactory.Create(ctx)
	if err != nil {
		return fmt.Errorf("listenerFactory.Create(): %w", err)
	}
	defer listener.Close()
	log.Printf("[App][%s] Listening\n", listener.Addr())

	var wg sync.WaitGroup

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("[App][%s] Failed to accept connection: %v\n", listener.Addr(), err)
			continue
		}
		log.Printf("[App][%s] Accepted connection %s\n", listener.Addr(), conn.RemoteAddr())
		fmt.Printf("new conn\n")

		wg.Add(1)
		go func(ctx context.Context, conn net.Conn) {
			defer wg.Done()
			defer conn.Close()

			if err := handler.Handle(ctx, conn); err != nil {
				log.Printf("[App][%s] Failed to handle connection %s: %v\n", listener.Addr(), conn.RemoteAddr(), err)
				return
			}

			log.Printf("[App][%s] Connection %s handled\n", listener.Addr(), conn.RemoteAddr())
		}(ctx, conn)
	}

	wg.Wait()

	return nil
}
