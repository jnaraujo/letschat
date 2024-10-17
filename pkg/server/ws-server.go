package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/message"
)

type Client struct {
	Conn     Connection
	Account  *account.Account
	JoinedAt time.Time
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
		// unauthenticated user
		Account:  account.NewAccount("Anonymous"),
		JoinedAt: time.Now(),
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
			// we need to find a way to detect the type of errors.
			// for now, we are just closing the connection,
			// even for non-critical errors.
			return
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
	cmdProps := &CommandProps{
		MessageAuthor: client,
		Msg:           msg,
		Server:        s,
	}

	// TODO: fix this
	if strings.HasPrefix(msg.Content, "ls") {
		lsCommand(cmdProps)
		return
	}
	if strings.HasPrefix(msg.Content, "ping") {
		pingCommand(cmdProps)
		return
	}

	client.Conn.WriteMessage(
		message.NewCommandChatMessage("command not found", time.Now()),
	)
}

func (s *Server) getSortedClientIDs() []id.ID {
	clientIDs := make([]id.ID, 0, len(s.clients))
	for clientID := range s.clients {
		clientIDs = append(clientIDs, clientID)
	}
	slices.SortFunc(clientIDs, func(a, b id.ID) int {
		return s.clients[a].JoinedAt.Compare(s.clients[b].JoinedAt)
	})
	return clientIDs
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

func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		hours := int(d.Hours())
		return fmt.Sprintf("%d hour%s ago", hours, plural(hours))
	} else if d.Minutes() >= 1 {
		minutes := int(d.Minutes())
		return fmt.Sprintf("%d minute%s ago", minutes, plural(minutes))
	} else {
		seconds := int(d.Seconds())
		return fmt.Sprintf("%d second%s ago", seconds, plural(seconds))
	}
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
