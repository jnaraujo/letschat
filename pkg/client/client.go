package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/server"
)

type WSClient struct {
	Addr string
	server.WSConnection
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

	wsc.WSConnection.Conn, _, err = dialer.DialContext(ctx, wsc.Addr, nil)
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
