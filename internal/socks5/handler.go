package socks5

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/rylenko/proxy/internal/proxy"
)

const (
	authMethodNotRequired         byte = 0x00
	authMethodNoAcceptable        byte = 0xFF

	requiredSOCKSVersion          byte = 0x05
)

type Handler struct {}

func (h *Handler) Handle(ctx context.Context, conn net.Conn) error {
	if err := h.handshake(ctx, conn); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}
	log.Printf("[Handler][%s] Handshake was successfully made", conn.RemoteAddr())

	if err := h.processRequest(ctx, conn); err != nil {
		return fmt.Errorf("process request: %w", err)
	}
	log.Printf("[Handler][%s] Request was successfully processed", conn.RemoteAddr())

	return nil
}

func (h *Handler) handshake(ctx context.Context, conn net.Conn) error {
	reader := bufio.NewReader(conn)

	version, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("read version: %w", err)
	}
	log.Printf("[Handler][%s] Version readed", conn.RemoteAddr())

	if version != requiredSOCKSVersion {
		return fmt.Errorf("unknown SOCKS version %d", int(version))
	}

	authMethodsCount, err := reader.ReadByte()
	if err != nil {
		return fmt.Errorf("read auth methods count: %w", err)
	}
	log.Printf("[Handler][%s] Auth methods count readed", conn.RemoteAddr())

	authMethods := make([]byte, int(authMethodsCount))
	if _, err := io.ReadFull(reader, authMethods); err != nil {
		return fmt.Errorf("read %d auth methods: %w", int(authMethodsCount), err)
	}
	log.Printf("[Handler][%s] Auth methods readed", conn.RemoteAddr())

	authMethodRequired := true
	for _, method := range authMethods {
		if method == authMethodNotRequired {
			authMethodRequired = false
			break
		}
	}

	if authMethodRequired {
		if _, err := conn.Write([]byte{requiredSOCKSVersion, authMethodNoAcceptable}); err != nil {
			return fmt.Errorf("write auth method no acceptable: %w", err)
		}
		return errors.New("auth method required")
	}

	if _, err := conn.Write([]byte{requiredSOCKSVersion, authMethodNotRequired}); err != nil {
		return fmt.Errorf("write auth method: %w", err)
	}

	return nil
}

func (h *Handler) processRequest(ctx context.Context, conn net.Conn) error {
	return nil
}

func NewHandler() *Handler {
	return &Handler{}
}

// Ensure that SOCKS5 handler implements handler interface.
var _ proxy.Handler = (*Handler)(nil)
