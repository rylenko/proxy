package main

import (
	"context"
	"flag"
	"log"

	"github.com/rylenko/proxy/internal/app"
	"github.com/rylenko/proxy/internal/socks5"
)

var port *int = flag.Int("port", 5555, "port for listening clients")

func main() {
	flag.Parse()

	proxy := socks5.NewProxy(*port)

	if err := app.Run(context.Background(), proxy); err != nil {
		log.Fatal("Application error: ", err)
	}
}
