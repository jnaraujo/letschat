package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/protocol"
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

	http.HandleFunc("/lc", server.handleNewConnection)
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
		if errors.Is(err, ErrConnectionClosed) {
			return
		}
		client.Conn.WritePacket(
			protocol.ServerAuthMessage{
				Status:  "auth_error",
				Content: "failed to auth",
			}.ToPacket(),
		)
		slog.Error("failed to initialize connection", "err", err)
		return
	}

	slog.Info("client authenticated", "username", client.Account.Username, "id", client.Account.ID)

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
	pkt, err := client.Conn.ReadPacket()
	if err != nil {
		return err
	}

	if pkt.Header.PacketType != protocol.PacketTypeAuth {
		return errors.New("expected auth packet")
	}

	authMsg, err := protocol.ClientAuthMessageFromPacket(pkt)
	if err != nil {
		return err
	}

	if len(authMsg.Username) <= 3 {
		return errors.New("username is too short")
	}
	if len(authMsg.Username) >= 16 {
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

	err = client.Conn.WritePacket(
		protocol.ServerAuthMessage{
			Status:  "ok",
			Content: "account authenticated",
			RoomID:  client.RoomID, // TODO: should check whether the room exists
			Account: client.Account,
		}.ToPacket(),
	)
	if err != nil {
		return err
	}

	room.AddClient(client)

	return nil
}

func (s *Server) handleIncomingMessages(client *Client) {
	for {
		pkt, err := client.Conn.ReadPacket()
		if err != nil {
			if errors.Is(err, protocol.ErrProtocolVersionMismatch) {
				slog.Error("protocol version mismatch", "err", err)
				return
			}
			if errors.Is(err, ErrConnectionClosed) {
				return
			}
			slog.Error("error reading message", "err", err)
			break
		}

		if pkt.Header.PacketType == protocol.PacketTypePing {
			client.Conn.Ping()
			continue
		}

		var msg protocol.ChatMessage
		err = json.Unmarshal(pkt.Payload, &msg)
		if err != nil {
			slog.Error("error reading message", "err", err)
			continue
		}

		if len(msg.Content) == 0 || len(msg.Content) > 100 {
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
			protocol.NewChatMessage(
				client.Account, msg.Content, protocol.ChatRoom{
					ID:   room.ID,
					Name: room.Name,
				}, time.Now(),
			),
		)
	}
}

func (s *Server) handleCommand(client *Client, msg *protocol.ChatMessage) {
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
	if strings.HasPrefix(msg.Content, "join") {
		joinRoomCommand(cmdProps)
		return
	}
	if strings.HasPrefix(msg.Content, "new") {
		createRoomCommand(cmdProps)
		return
	}

	client.Conn.WritePacket(
		protocol.NewCommandChatMessage(
			"command not found", time.Now(),
		).ToPacket(),
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
