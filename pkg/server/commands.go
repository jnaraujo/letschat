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

	res.WriteString("==== List of Online Clients ====\n")
	for _, clientID := range sortClientIDsByJoinTime(props.Server.clients.List()) {
		client := props.Server.clients.Find(clientID)

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

func pingCommand(props *CommandProps) {
	props.MessageAuthor.Conn.WriteMessage(message.NewCommandChatMessage(
		// in ping commands, the createdAt remains the same as the original sent
		// maybe fix this in the future idk
		"Pong!", props.Msg.CreatedAt,
	))
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
