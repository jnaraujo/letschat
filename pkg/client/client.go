package client

import (
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/server"
)

type WSClient struct {
	Addr string
	server.WSConnection
}

func NewWSClient(addr string) *WSClient {
	u := url.URL{Scheme: "ws", Host: addr, Path: "/ws"}
	return &WSClient{
		Addr: u.String(),
	}
}

func (wsc *WSClient) Connect() (err error) {
	wsc.WSConnection.Conn, _, err = websocket.DefaultDialer.Dial(
		wsc.Addr, nil,
	)
	return err
}
