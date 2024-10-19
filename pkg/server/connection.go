package server

type Connection interface {
	Write(data []byte) error
	Read() ([]byte, error)
	WriteMessage(msg any) error
	ReadMessage(msg any) error
	Ping() error
	Close() error
}
