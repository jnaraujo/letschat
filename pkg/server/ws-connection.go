package server

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSConnection struct {
	Conn *websocket.Conn

	rMutex sync.Mutex
	wMutex sync.Mutex
}

func (wsc *WSConnection) Write(data []byte) error {
	wsc.wMutex.Lock()
	defer wsc.wMutex.Unlock()

	return wsc.Conn.WriteMessage(websocket.TextMessage, data)
}

func (wsc *WSConnection) Read() ([]byte, error) {
	wsc.rMutex.Lock()
	defer wsc.rMutex.Unlock()

	messageType, data, err := wsc.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if messageType != websocket.TextMessage {
		return nil, errors.New("message type should be text")
	}
	return data, nil
}

func (wsc *WSConnection) WriteMessage(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return wsc.Write(data)
}

func (wsc *WSConnection) ReadMessage(msg any) error {
	data, err := wsc.Read()
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, msg)
	if err != nil {
		return err
	}
	return nil
}

func (wsc *WSConnection) Ping() error {
	return wsc.Conn.SetWriteDeadline(time.Now().Add(MaxKeepAlive))
}

func (wsc *WSConnection) Close() error {
	return wsc.Conn.Close()
}
