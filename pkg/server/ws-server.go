package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/message"
)

type Client struct {
	Conn    Connection
	Account *account.Account
}

type Server struct {
	clients map[id.ID]*Client
}

func NewServer() *Server {
	server := &Server{
		clients: make(map[id.ID]*Client),
	}
	http.HandleFunc("/ws", server.handleWsConn)
	return server
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, nil)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleWsConn(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading connection", "err", err)
		return
	}

	client := &Client{
		// generate a random account to client
		Account: account.NewAccount("Anonymous"),
		Conn: &WSConnection{
			Conn: conn,
		},
	}
	s.clients[client.Account.ID] = client

	defer func() {
		slog.Info("closing connection", "ID", client.Account.ID)
		delete(s.clients, client.Account.ID)
		if err := client.Conn.Close(); err != nil {
			panic(err)
		}
		s.Broadcast(
			message.NewServerChatMessage(
				fmt.Sprintf("%s (%s) left the chat", client.Account.Username, client.Account.ID),
				time.Now(),
			),
		)
	}()

	err = s.handleInitialConn(client)
	if err != nil {
		client.Conn.Write([]byte("failed to initialize connection"))
		slog.Error("failed to initialize connection", "err", err)
		return
	}
	client.Conn.Write([]byte("ok"))

	s.Broadcast(
		message.NewServerChatMessage(
			fmt.Sprintf("New connection from %s (%s)", client.Account.Username, client.Account.ID),
			time.Now(),
		),
	)

	s.handleIncomingMessages(client)
}

func (s *Server) handleIncomingMessages(client *Client) {
	for {
		var msg message.ChatMessage
		err := client.Conn.ReadMessage(&msg)
		if err != nil {
			slog.Error("error reading message", "err", err)
			continue
		}
		if len(msg.Content) > 100 {
			continue
		}

		s.Broadcast(message.NewChatMessage(
			client.Account, msg.Content, time.Now(),
		))
	}
}

func (s *Server) handleInitialConn(client *Client) error {
	msg, err := client.Conn.Read()
	if err != nil {
		return err
	}

	// TODO: create a default message format for the initial connection
	username := string(msg)
	if len(username) <= 3 {
		return errors.New("username is too short")
	}
	if len(username) >= 16 {
		return errors.New("username is too long")
	}

	client.Account.Username = username
	return nil
}

func (s *Server) Broadcast(msg any) {
	for _, client := range s.clients {
		client.Conn.WriteMessage(msg)
	}
}

func (s *Server) BroadcastExcept(exceptID id.ID, msg any) {
	for _, client := range s.clients {
		if client.Account.ID == exceptID {
			continue
		}
		client.Conn.WriteMessage(msg)
	}
}
