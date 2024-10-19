package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/server"
)

type WSClient struct {
	Addr string
	Conn *websocket.Conn

	rMutex sync.Mutex
	wMutex sync.Mutex
}

func NewWSClient(addr string) *WSClient {
	return &WSClient{
		Addr: addr,
	}
}

func (wsc *WSClient) Connect(ctx context.Context) (err error) {
	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true,
	}

	wsc.keepAlive(ctx)

	wsc.Conn, _, err = dialer.DialContext(ctx, wsc.Addr, nil)
	return err
}

func (wsc *WSClient) keepAlive(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(server.MaxKeepAlive - 5)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
			case <-ticker.C:
				err := wsc.Conn.WriteMessage(websocket.PingMessage, nil)
				if err != nil {
					fmt.Println(err)
				}
			}
		}
	}()
}

func (wsc *WSClient) Write(data []byte) error {
	wsc.wMutex.Lock()
	defer wsc.wMutex.Unlock()

	return wsc.Conn.WriteMessage(websocket.TextMessage, data)
}

func (wsc *WSClient) Read() ([]byte, error) {
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

func (wsc *WSClient) WriteMessage(msg any) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	return wsc.Write(data)
}

func (wsc *WSClient) ReadMessage(msg any) error {
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

func (wsc *WSClient) Close() error {
	return wsc.Conn.Close()
}
