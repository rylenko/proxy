package socks5

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"slices"

	"github.com/rylenko/proxy/internal/proxy"
)

const (
	addrTypeIPv4               byte = 0x01
	addrTypeDomain             byte = 0x03

	authMethodNotRequired      byte = 0x00
	authMethodNoAcceptable     byte = 0xFF

	commandConnect             byte = 0x01

	replyStatusSuccess         byte = 0x00
	replyStatusNetUnreachable  byte = 0x03
	replyStatusConnRefused     byte = 0x05
	replyStatusUnknownCommand  byte = 0x07
	replyStatusUnknownAddrType byte = 0x08

	requiredAuthMethod         byte = authMethodNotRequired
	requiredVersion            byte = 0x05

	ipv4BytesLen               int  = 4
	portBytesLen               int  = 2
)

type Handler struct {}

func (h *Handler) Handle(srcConn net.Conn) error {
	srcReader := bufio.NewReader(srcConn)

	if err := h.handshake(srcReader, srcConn); err != nil {
		return fmt.Errorf("handshake: %w", err)
	}
	log.Printf("Handshaked with %s", srcConn.RemoteAddr())

	destConn, err := h.handleRequest(srcReader, srcConn)
	if err != nil {
		return fmt.Errorf("handle request: %w", err)
	}
	defer destConn.Close()
	log.Printf("Request from %s to %s handled", srcConn.RemoteAddr(), destConn.RemoteAddr())

	if err := h.exchange(srcReader, srcConn, destConn); err != nil {
		return fmt.Errorf("exchange source %s with destination %s: %w", srcConn.RemoteAddr(), destConn.RemoteAddr(), err)
	}
	return nil
}

// TODO: cancel one of goroutine on error in another and return an error.
func (h *Handler) exchange(srcReader *bufio.Reader, srcConn, destConn net.Conn) error {
	errCh := make(chan error)

	go func() {
		_, err := io.Copy(destConn, srcReader)
		if err != nil {
			errCh <- fmt.Errorf("copy source to destination: %w", err)
		}

		close(errCh)
	}()

	_, err := io.Copy(srcConn, destConn)
	if err != nil {
		err = fmt.Errorf("copy destination to source: %w", err)
	}

	err = errors.Join(err, <-errCh)
	return err
}

func (h *Handler) expectByte(reader *bufio.Reader, expected ...byte) (byte, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}

	if !slices.Contains(expected, b) {
		return 0, fmt.Errorf("invalid byte 0x%x, expected one of %v", b, expected)
	}

	return b, nil
}

func (h *Handler) handleRequest(reader *bufio.Reader, conn net.Conn) (net.Conn, error) {
	addrType, err := h.handleRequestUntilAddr(reader, conn)
	if err != nil {
		return nil, fmt.Errorf("handle until addr: %w", err)
	}

	destAddrStr, err := h.readRequestAddr(reader, conn, addrType)
	if err != nil {
		return nil, fmt.Errorf("read address with type 0x%x: %w", addrType, err)
	}

	dest, err := net.Dial("tcp4", destAddrStr)
	if err != nil {
		err = fmt.Errorf("dial with target %s: %w", destAddrStr, err)

		if replyErr := h.replyFail(conn, replyStatusNetUnreachable); replyErr != nil {
			return nil, fmt.Errorf("1. %w\n2. reply fail: %w", err, replyErr)
		}

		return nil, err
	}

	destLocalTCPAddr, ok := dest.LocalAddr().(*net.TCPAddr)
	if !ok {
		dest.Close()
		return nil, fmt.Errorf("get destination local TCP address of %s", dest.LocalAddr())
	}

	if err := h.replySuccess(conn, destLocalTCPAddr); err != nil {
		dest.Close()
		return nil, fmt.Errorf("reply success: %w", err)
	}

	return dest, nil
}

func (h *Handler) handleRequestUntilAddr(reader *bufio.Reader, conn net.Conn) (byte, error) {
	if _, err := h.expectByte(reader, requiredVersion); err != nil {
		err = fmt.Errorf("version: %w", err)

		if replyErr := h.replyFail(conn, replyStatusConnRefused); replyErr != nil {
			return 0, fmt.Errorf("1. %w\n2. reply fail: %w", err, replyErr)
		}

		return 0, err
	}

	if _, err := h.expectByte(reader, commandConnect); err != nil {
		err = fmt.Errorf("connect command: %w", err)

		if replyErr := h.replyFail(conn, replyStatusUnknownCommand); replyErr != nil {
			return 0, fmt.Errorf("1. %w\n2. reply fail: %w", err, replyErr)
		}

		return 0, err
	}

	if _, err := reader.ReadByte(); err != nil {
		err = fmt.Errorf("reserved: %w", err)

		if replyErr := h.replyFail(conn, replyStatusConnRefused); replyErr != nil {
			return 0, fmt.Errorf("1. %w\n2. reply fail: %w", err, replyErr)
		}

		return 0, err
	}

	addrType, err := h.expectByte(reader, addrTypeIPv4, addrTypeDomain)
	if err != nil {
		err = fmt.Errorf("address type: %w", err)

		if replyErr := h.replyFail(conn, replyStatusUnknownAddrType); replyErr != nil {
			return 0, fmt.Errorf("1. %w\n2. reply fail: %w", err, replyErr)
		}

		return 0, err
	}

	return addrType, nil
}

func (h *Handler) handshake(reader *bufio.Reader, conn net.Conn) error {
	if _, err := h.expectByte(reader, requiredVersion); err != nil {
		return fmt.Errorf("version: %w", err)
	}

	if _, err := h.handshakeAuthMethod(reader, conn); err != nil {
		return fmt.Errorf("auth method: %w", err)
	}

	return nil
}

func (h *Handler) handshakeAuthMethod(reader *bufio.Reader, conn net.Conn) (byte, error) {
	methodsCount, err := reader.ReadByte()
	if err != nil {
		return 0, fmt.Errorf("read count: %w", err)
	}

	methods := make([]byte, methodsCount)
	if _, err := io.ReadFull(reader, methods); err != nil {
		return 0, fmt.Errorf("read %d methods: %w", methodsCount, err)
	}

	if !slices.Contains(methods, requiredAuthMethod) {
		if _, err := conn.Write([]byte{requiredVersion, authMethodNoAcceptable}); err != nil {
			return 0, fmt.Errorf("write no acceptable auth method: %w", err)
		}
		return 0, errors.New("no acceptable auth methods")
	}

	if _, err := conn.Write([]byte{requiredVersion, requiredAuthMethod}); err != nil {
		return 0, fmt.Errorf("write auth method: %w", err)
	}

	return requiredAuthMethod, nil
}

func (h *Handler) readRequestAddr(reader *bufio.Reader, conn net.Conn, addrType byte) (string, error) {
	var addr []byte
	port := make([]byte, 2)

	switch addrType {
	case addrTypeIPv4:
		addr = make([]byte, ipv4BytesLen)
		if _, err := io.ReadFull(reader, addr); err != nil {
			return "", fmt.Errorf("read IPv4 with length %d: %w", len(addr), err)
		}
	case addrTypeDomain:
		addrLen, err := reader.ReadByte()
		if err != nil {
			return "", fmt.Errorf("read addr length: %w", err)
		}

		addr = make([]byte, addrLen)
		if _, err := io.ReadFull(reader, addr); err != nil {
			return "", fmt.Errorf("read domain with length %d: %w", len(addr), err)
		}
	default:
		return "", errors.New("unknown address type")
	}

	if _, err := io.ReadFull(reader, port); err != nil {
		return "", fmt.Errorf("read port with length %d: %w", len(addr), err)
	}

	if addrType == addrTypeIPv4 {
		return fmt.Sprintf("%s:%d", net.IP(addr).String(), binary.BigEndian.Uint16(port)), nil
	}
	return fmt.Sprintf("%s:%d", addr, binary.BigEndian.Uint16(port)), nil
}

func (h *Handler) replyFail(conn net.Conn, status byte) error {
	data := []byte{requiredVersion, status, 0x00, addrTypeIPv4, 0, 0, 0, 0, 0, 0}
	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("write %v: %w", data, err)
	}

	return nil
}

func (h *Handler) replySuccess(conn net.Conn, addr *net.TCPAddr) error {
	data := append(
		[]byte{},
		requiredVersion,
		replyStatusSuccess,
		0x00,
		addrTypeIPv4,
		addr.IP[0],
		addr.IP[1],
		addr.IP[2],
		addr.IP[3],
		byte((addr.Port & 0xFF00) >> 8),
		byte(addr.Port & 0xFF))

	if _, err := conn.Write(data); err != nil {
		return fmt.Errorf("write %v: %w", data, err)
	}
	return nil
}

func NewHandler() *Handler {
	return &Handler{}
}

// Ensure that SOCKS5 handler implements handler interface.
var _ proxy.Handler = (*Handler)(nil)
