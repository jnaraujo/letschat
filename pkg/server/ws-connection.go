package server

import (
	"encoding/json"
	"errors"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/message"
)

type WSConnection struct {
	Conn *websocket.Conn
}

func (wsc *WSConnection) Write(data []byte) error {
	return wsc.Conn.WriteMessage(websocket.TextMessage, data)
}

func (wsc *WSConnection) Read() ([]byte, error) {
	messageType, data, err := wsc.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if messageType != websocket.TextMessage {
		return nil, errors.New("message type should be text")
	}
	return data, nil
}

func (wsc *WSConnection) WriteMessage(msg *message.BaseMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return wsc.Conn.WriteMessage(websocket.TextMessage, data)
}

func (wsc *WSConnection) ReadMessage() (*message.BaseMessage, error) {
	data, err := wsc.Read()
	if err != nil {
		return nil, err
	}

	var msg message.BaseMessage
	err = json.Unmarshal(data, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (wsc *WSConnection) Close() error {
	return wsc.Conn.Close()
}
