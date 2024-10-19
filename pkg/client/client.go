package client

import (
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

func (wsc *WSClient) Connect() (err error) {
	dialer := &websocket.Dialer{
		Proxy:             http.ProxyFromEnvironment,
		HandshakeTimeout:  45 * time.Second,
		EnableCompression: true,
	}

	wsc.WSConnection.Conn, _, err = dialer.Dial(wsc.Addr, nil)
	return err
}
