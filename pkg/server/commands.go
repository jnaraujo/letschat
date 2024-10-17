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
	for _, clientID := range sortClientIDs(props.Server.getClients()) {
		res.WriteString(fmt.Sprintf(" %s (%s) - %s\n",
			props.Server.clients[clientID].Account.Username,
			string(props.Server.clients[clientID].Account.ID),
			utils.FormatDuration(time.Since(props.Server.clients[clientID].JoinedAt)),
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

func sortClientIDs(clients map[id.ID]*Client) []id.ID {
	clientIDs := make([]id.ID, 0, len(clients))
	for clientID := range clients {
		clientIDs = append(clientIDs, clientID)
	}
	slices.SortFunc(clientIDs, func(a, b id.ID) int {
		return clients[a].JoinedAt.Compare(clients[b].JoinedAt)
	})
	return clientIDs
}
