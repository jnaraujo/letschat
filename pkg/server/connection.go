package server

import "github.com/jnaraujo/letschat/pkg/message"

type Connection interface {
	Write(data []byte) error
	Read() ([]byte, error)
	WriteMessage(msg *message.BaseMessage) error
	ReadMessage() (*message.BaseMessage, error)
	Close() error
}
