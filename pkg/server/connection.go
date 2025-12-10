package server

import (
	"errors"

	"github.com/jnaraujo/letschat/pkg/protocol"
)

type Connection interface {
	Write(data []byte) error
	Read() ([]byte, error)

	WritePacket(pkt *protocol.Packet) error
	ReadPacket() (*protocol.Packet, error)

	RemoteAddr() string

	Ping() error
	Close() error
}

var ErrConnectionClosed = errors.New("connection closed")
