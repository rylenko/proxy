package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/rylenko/proxy/internal/app"
	"github.com/rylenko/proxy/internal/socks5"
)

const (
	logFileDefaultName string      = "proxy-socks5.log"
	logFileFlag        int         = os.O_CREATE | os.O_APPEND | os.O_WRONLY
	logFilePerm        os.FileMode = 0644
	logFlags           int         = log.Ldate | log.Ltime | log.Lshortfile
)

var (
	port        *int    = flag.Int("port", 5555, "port for listening clients")
	logFilePath *string = flag.String("log", os.TempDir() + "/" + logFileDefaultName, "path to the logs file")
)

func initLog() error {
	file, err := os.OpenFile(*logFilePath, logFileFlag, logFilePerm)
	if err != nil {
		return fmt.Errorf("os.OpenFile(\"%s\", %d, %s): %w", *logFilePath, logFileFlag, logFilePerm.String(), err)
	}

	log.SetOutput(file)
	log.SetFlags(logFlags)

	return nil
}

func main() {
	flag.Parse()

	if err := initLog(); err != nil {
		log.Fatal("init log: ", err)
	}

	listenerFactory := socks5.NewListenerFactory(*port)
	handler := socks5.NewHandler()

	if err := app.Run(context.Background(), listenerFactory, handler); err != nil {
		log.Fatal("run app: ", err)
	}
}
