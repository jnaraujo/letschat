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

const (
	defaultRoomID id.ID = "ALL"
	MaxKeepAlive        = 60 * time.Second
	MaxPing             = MaxKeepAlive / 2
)

type Server struct {
	rooms *RoomList
}

func NewServer() *Server {
	server := &Server{
		rooms: NewRoomList(),
	}

	defaultRoom := NewRoom("ALL", nil)
	defaultRoom.ID = defaultRoomID
	server.rooms.Add(defaultRoom)

	http.HandleFunc("/ws", server.handleNewConnection)
	return server
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, nil)
}

var upgrader = websocket.Upgrader{
	EnableCompression: true,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) handleNewConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("error upgrading connection", "err", err)
		return
	}
	defer conn.Close()

	// unauthenticated user
	client := NewClient(
		account.NewAccount("Anonymous"),
		&WSConnection{
			Conn: conn,
		},
	)

	conn.SetReadDeadline(time.Now().Add(MaxKeepAlive))
	conn.SetPingHandler(func(appData string) error {
		return client.Conn.Ping()
	})

	err = s.handleAuth(client)
	if err != nil {
		slog.Error("failed to initialize connection", "err", err)
		return
	}

	clientRoom := s.rooms.Find(client.RoomID)
	if clientRoom == nil {
		fmt.Println("client room does not exists")
		return
	}
	defer func() {
		room := s.rooms.Find(client.RoomID)
		if room != nil {
			room.RemoveClient(client.Account.ID)
		}
	}()

	s.handleIncomingMessages(client)
}

func (s *Server) handleAuth(client *Client) error {
	var authMsg message.AuthMessageClient
	err := client.Conn.ReadMessage(&authMsg)
	if err != nil {
		client.Conn.WriteMessage(message.AuthMessageServer{
			Status:  "auth_error",
			Content: "failed to read message",
		})
		return err
	}

	if len(authMsg.Username) <= 3 {
		client.Conn.WriteMessage(message.AuthMessageServer{
			Status:  "auth_error",
			Content: "username is too short",
		})
		return errors.New("username is too short")
	}
	if len(authMsg.Username) >= 16 {
		client.Conn.WriteMessage(message.AuthMessageServer{
			Status:  "auth_error",
			Content: "username is too long",
		})
		return errors.New("username is too long")
	}

	client.Account.Username = authMsg.Username

	var room *Room
	if authMsg.RoomID != "" && s.rooms.Has(authMsg.RoomID) {
		room = s.rooms.Find(authMsg.RoomID)
	} else {
		room = s.rooms.Find(defaultRoomID)
		if room == nil {
			return errors.New("default room does not exists")
		}
	}

	err = client.Conn.WriteMessage(message.AuthMessageServer{
		Status:  "ok",
		Content: "account authenticated",
		RoomID:  client.RoomID, // TODO: should check whether the room exists
	})
	if err != nil {
		return err
	}

	err = client.Conn.WriteMessage(client.Account)
	if err != nil {
		return err
	}

	room.AddClient(client)

	return nil
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

		room := s.rooms.Find(client.RoomID)
		if room == nil {
			continue
		}

		room.Broadcast(
			message.NewChatMessage(
				client.Account, msg.Content, message.CharRoom{
					ID:   room.ID,
					Name: room.Name,
				}, time.Now(),
			),
		)
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
	if strings.HasPrefix(msg.Content, "client-ping") {
		clientPingCommand(cmdProps)
		return
	}
	if strings.HasPrefix(msg.Content, "ping") {
		pingCommand(cmdProps)
		return
	}
	if strings.HasPrefix(msg.Content, "join") {
		joinRoomCommand(cmdProps)
		return
	}
	if strings.HasPrefix(msg.Content, "new") {
		createRoomCommand(cmdProps)
		return
	}

	client.Conn.WriteMessage(
		message.NewCommandChatMessage(
			"command not found", time.Now(),
		),
	)
}

func (s *Server) addClientToRoom(client *Client, roomID id.ID) {
	if s.rooms.Has(client.RoomID) {
		s.rooms.Find(client.RoomID).RemoveClient(client.Account.ID)
	}
	if s.rooms.Has(roomID) {
		s.rooms.Find(roomID).AddClient(client)
	}
}
