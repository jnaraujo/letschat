package protocol

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"time"

	"github.com/fatih/color"
	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
)

type ChatRoom struct {
	ID   id.ID  `json:"id"`
	Name string `json:"name"`
}

type ChatMessage struct {
	ID        id.ID            `json:"id"`
	IsServer  bool             `json:"is_server"`
	Room      ChatRoom         `json:"room"`
	Author    *account.Account `json:"author"`
	Content   string           `json:"content"`
	CreatedAt time.Time        `json:"created_at"`
	IsCommand bool             `json:"is_command"`
}

func NewChatMessage(author *account.Account, content string,
	chatRoom ChatRoom, createdAt time.Time) ChatMessage {
	return ChatMessage{
		ID:        id.NewID(22),
		Author:    author,
		Content:   content,
		CreatedAt: createdAt,
		Room:      chatRoom,
		IsCommand: false,
		IsServer:  false,
	}
}

func NewServerChatMessage(content string, chatRoom ChatRoom, createdAt time.Time) ChatMessage {
	msg := NewChatMessage(&account.Account{
		ID:       "SERVER",
		Username: "SERVER",
	}, content, chatRoom, createdAt)
	msg.IsServer = true
	return msg
}

func NewCommandChatMessage(content string, createdAt time.Time) ChatMessage {
	msg := NewChatMessage(&account.Account{
		ID:       "COMMAND",
		Username: "COMMAND",
	}, content, ChatRoom{
		ID:   id.ID("COMMAND_RESPONSE"),
		Name: "Command Response",
	}, createdAt)
	msg.IsCommand = true
	return msg
}

func ChatMessageFromPacket(pkt Packet) (ChatMessage, error) {
	var msg ChatMessage
	if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func (msg ChatMessage) ToPacket() Packet {
	payload, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return NewPacket(PacketTypeMessage, payload)
}

// FIX: move this to another place
func (msg ChatMessage) Show() {
	if msg.IsServer {
		c := color.New(color.Italic, color.Faint)
		fmt.Printf("[%s] <%s>: %s\n",
			color.HiBlueString(timeFormat(msg.CreatedAt)),
			color.WhiteString(string(msg.Author.ID)),
			c.Sprintf(msg.Content),
		)
		return
	}

	if msg.IsCommand {
		fmt.Println(msg.Content)
		return
	}

	pc := color.New(s2c(string(msg.Author.ID)))

	fmt.Printf("[%s] [%s] <%s> %s: %s\n",
		color.HiBlueString(timeFormat(msg.CreatedAt)),
		color.HiBlueString(string(msg.Room.Name)),
		pc.Sprint(string(msg.Author.ID[:6])),
		pc.Sprint(msg.Author.Username),
		msg.Content)
}

func timeFormat(t time.Time) string {
	if time.Since(t) > 24*time.Hour {
		return t.Format(time.DateTime)
	}
	return t.Format(time.Kitchen)
}

var colors = []color.Attribute{
	color.FgHiBlue,
	color.FgHiRed,
	color.FgHiGreen,
	color.FgHiYellow,
	color.FgHiMagenta,
	color.FgHiCyan,
	color.FgHiWhite,
	color.FgRed,
	color.FgGreen,
	color.FgYellow,
	color.FgBlue,
	color.FgMagenta,
	color.FgCyan,
	color.FgWhite,
}

func s2c(txt string) color.Attribute {
	return colors[int(hash(txt))%len(colors)]
}

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}
