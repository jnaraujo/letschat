package client

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/protocol"
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
		ticker := time.NewTicker(server.MaxPing)
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
			case <-ticker.C:
				err := wsc.Ping()
				if err != nil {
					return
				}
			}
		}
	}()
}

func (wsc *WSClient) Ping() error {
	wsc.wMutex.Lock()
	defer wsc.wMutex.Unlock()

	return wsc.Conn.WriteMessage(websocket.PingMessage, nil)
}

func (wsc *WSClient) Write(data []byte) error {
	wsc.wMutex.Lock()
	defer wsc.wMutex.Unlock()

	return wsc.Conn.WriteMessage(websocket.BinaryMessage, data)
}

func (wsc *WSClient) Read() ([]byte, error) {
	wsc.rMutex.Lock()
	defer wsc.rMutex.Unlock()

	messageType, data, err := wsc.Conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	if messageType != websocket.BinaryMessage {
		return nil, errors.New("message type should be text")
	}
	return data, nil
}

func (wsc *WSClient) WritePacket(pkt *protocol.Packet) error {
	data, err := pkt.ToBinary()
	if err != nil {
		return err
	}
	return wsc.Write(data)
}

func (wsc *WSClient) ReadPacket() (*protocol.Packet, error) {
	data, err := wsc.Read()
	if err != nil {
		return nil, err
	}

	return protocol.PacketFromBytes(data)
}

func (wsc *WSClient) Close() error {
	return wsc.Conn.Close()
}
