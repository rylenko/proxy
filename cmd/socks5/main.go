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
)

var (
	port        *int    = flag.Int("port", 5555, "port for listening clients")
	logFilePath *string = flag.String(
		"log", os.TempDir() + "/" + logFileDefaultName, "path to the logs file")
)

func setLogFile() error {
	file, err := os.OpenFile(*logFilePath, logFileFlag, logFilePerm)
	if err != nil {
		return fmt.Errorf(
			"os.OpenFile(\"%s\", %d, %s): %w",
			*logFilePath,
			logFileFlag,
			logFilePerm.String(),
			err)
	}

	log.SetOutput(file)

	return nil
}

func main() {
	flag.Parse()

	if err := setLogFile(); err != nil {
		log.Fatal("setLogFile(): ", err)
	}

	proxy := socks5.NewProxy(*port)

	if err := app.Run(context.Background(), proxy); err != nil {
		log.Fatal("app.Run(): ", err)
	}
}
