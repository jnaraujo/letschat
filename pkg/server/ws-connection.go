package server

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/protocol"
)

type WSConnection struct {
	Conn *websocket.Conn

	rMutex sync.Mutex
	wMutex sync.Mutex
}

func (wsc *WSConnection) Write(data []byte) error {
	wsc.wMutex.Lock()
	defer wsc.wMutex.Unlock()

	err := wsc.Conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		if websocket.IsUnexpectedCloseError(err,
			websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return err
		}
		return ErrConnectionClosed
	}
	return nil
}

func (wsc *WSConnection) Read() ([]byte, error) {
	wsc.rMutex.Lock()
	defer wsc.rMutex.Unlock()

	messageType, data, err := wsc.Conn.ReadMessage()
	if err != nil {
		if websocket.IsUnexpectedCloseError(err,
			websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			return nil, err
		}
		return nil, ErrConnectionClosed
	}
	if messageType != websocket.BinaryMessage {
		return nil, errors.New("message type should be binary")
	}
	return data, nil
}

func (wsc *WSConnection) WritePacket(pkt protocol.Packet) error {
	data, err := pkt.ToBinary()
	if err != nil {
		return err
	}
	return wsc.Write(data)
}

func (wsc *WSConnection) ReadPacket() (protocol.Packet, error) {
	data, err := wsc.Read()
	if err != nil {
		return protocol.Packet{}, err
	}

	return protocol.PacketFromBytes(data)
}

func (wsc *WSConnection) Ping() error {
	return wsc.Conn.SetReadDeadline(time.Now().Add(MaxKeepAlive))
}

func (wsc *WSConnection) Close() error {
	return wsc.Conn.Close()
}
