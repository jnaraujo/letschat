package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/jnaraujo/letschat/pkg/message"
)

type CommandProps struct {
	MessageAuthor *Client
	Msg           *message.ChatMessage
	Server        *Server
}

func lsCommand(props *CommandProps) {
	var res strings.Builder

	res.WriteString("==== List of Online Clients ====\n")
	for _, clientID := range props.Server.getSortedClientIDs() {
		res.WriteString(fmt.Sprintf(" %s (%s) - %s\n",
			props.Server.clients[clientID].Account.Username,
			string(props.Server.clients[clientID].Account.ID),
			formatDuration(time.Since(props.Server.clients[clientID].JoinedAt)),
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
