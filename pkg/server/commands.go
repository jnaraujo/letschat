package server

import (
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/message"
	"github.com/jnaraujo/letschat/pkg/utils"
)

type CommandProps struct {
	MessageAuthor *Client
	Msg           *message.ChatMessage
	Server        *Server
}

func lsCommand(props *CommandProps) {
	var res strings.Builder

	room := props.Server.rooms.Find(props.MessageAuthor.RoomID)
	if room == nil {
		props.MessageAuthor.Conn.WriteMessage(
			message.NewCommandChatMessage(
				"You need to be connected to a room to view the list of online clients.",
				time.Now(),
			),
		)
		return
	}

	res.WriteString("==== List of Online Clients ====\n")
	for _, clientID := range sortClientIDsByJoinTime(room.Clients.List()) {
		client := room.Clients.Find(clientID)
		if client == nil {
			continue
		}

		res.WriteString(fmt.Sprintf(" %s (%s) - %s\n",
			client.Account.Username,
			string(client.Account.ID),
			utils.FormatDuration(time.Since(client.JoinedAt)),
		))
	}
	res.WriteString("================================")

	props.MessageAuthor.Conn.WriteMessage(
		message.NewCommandChatMessage(res.String(), time.Now()),
	)
}

func createRoomCommand(props *CommandProps) {
	words := strings.Split(props.Msg.Content, " ")
	if len(words) != 2 {
		return
	}
	name := words[1]

	room := NewRoom(name, props.Msg.Author)
	props.Server.rooms.Add(room)

	props.MessageAuthor.Conn.WriteMessage(message.NewCommandChatMessage(
		fmt.Sprintf("Room \"%s\" created! Join and invite your friends with: /join %s",
			room.Name, room.ID),
		time.Now(),
	))
}

func joinRoomCommand(props *CommandProps) {
	words := strings.Split(props.Msg.Content, " ")
	if len(words) != 2 {
		return
	}

	roomID := words[1]
	if !props.Server.rooms.Has(id.ID(roomID)) {
		return
	}

	props.Server.addClientToRoom(props.MessageAuthor, id.ID(roomID))
}

func pingCommand(props *CommandProps) {
	props.MessageAuthor.Conn.WriteMessage(message.NewCommandChatMessage(
		// in ping commands, the createdAt remains the same as the original sent
		// maybe fix this in the future idk
		"Pong!", props.Msg.CreatedAt,
	))
}

func clientPingCommand(props *CommandProps) {
	props.MessageAuthor.Conn.Ping()
}

func sortClientIDsByJoinTime(clients []*Client) []id.ID {
	slices.SortFunc(clients, func(clientA, clientB *Client) int {
		return clientA.JoinedAt.Compare(clientB.JoinedAt)
	})

	clientIDs := make([]id.ID, 0, len(clients))
	for _, client := range clients {
		clientIDs = append(clientIDs, client.Account.ID)
	}

	return clientIDs
}
