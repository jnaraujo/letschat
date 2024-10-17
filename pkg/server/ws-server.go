package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
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

	err = s.handleAuth(client)
	if err != nil {
		slog.Error("failed to initialize connection", "err", err)
		return
	}

	s.Broadcast(
		message.NewServerChatMessage(
			fmt.Sprintf("%s (%s) joined the chat", client.Account.Username, client.Account.ID),
			time.Now(),
		),
	)

	s.handleIncomingMessages(client)
}

func (s *Server) handleAuth(client *Client) error {
	var clientAuth message.AuthMessageClient
	err := client.Conn.ReadMessage(&clientAuth)
	if err != nil {
		client.Conn.WriteMessage(message.AuthMessageServer{
			Status:  "auth_error",
			Content: "failed to read message",
		})
		return err
	}

	if len(clientAuth.Username) <= 3 {
		client.Conn.WriteMessage(message.AuthMessageServer{
			Status:  "auth_error",
			Content: "username is too short",
		})
		return errors.New("username is too short")
	}
	if len(clientAuth.Username) >= 16 {
		client.Conn.WriteMessage(message.AuthMessageServer{
			Status:  "auth_error",
			Content: "username is too long",
		})
		return errors.New("username is too long")
	}

	client.Account.Username = clientAuth.Username

	err = client.Conn.WriteMessage(message.AuthMessageServer{
		Status:  "ok",
		Content: "account authenticated",
	})
	if err != nil {
		return err
	}

	return client.Conn.WriteMessage(client.Account)
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

		if msg.IsCommand {
			s.handleCommand(client, &msg)
			continue
		}

		s.Broadcast(message.NewChatMessage(
			client.Account, msg.Content, time.Now(),
		))
	}
}

func (s *Server) handleCommand(client *Client, msg *message.ChatMessage) {
	if strings.HasPrefix(msg.Content, "ls") {
		var res strings.Builder
		res.WriteString("==== List of Online Clients ====\n")
		for _, client := range s.clients {
			res.WriteString(" ")
			res.WriteString(client.Account.Username)
			res.WriteString(" (")
			res.WriteString(string(client.Account.ID))
			res.WriteString(")")
			res.WriteString("\n")
		}
		res.WriteString("================================")

		client.Conn.WriteMessage(
			message.NewCommandChatMessage(res.String(), time.Now()),
		)
		return
	}
	client.Conn.WriteMessage(
		message.NewCommandChatMessage("command not found", time.Now()),
	)
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
