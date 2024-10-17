package message

import (
	"fmt"
	"hash/fnv"
	"time"

	"github.com/fatih/color"
	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
)

type ChatMessage struct {
	ID        id.ID            `json:"id"`
	IsServer  bool             `json:"is_server"`
	Author    *account.Account `json:"author"`
	Content   string           `json:"content"`
	CreatedAt time.Time        `json:"created_at"`
}

func NewChatMessage(author *account.Account, content string, createdAt time.Time) *ChatMessage {
	return &ChatMessage{
		ID:        id.NewID(16),
		Author:    author,
		Content:   content,
		CreatedAt: createdAt,
		IsServer:  false,
	}
}

func NewServerChatMessage(content string, createdAt time.Time) *ChatMessage {
	msg := NewChatMessage(&account.Account{
		ID:       "SERVER",
		Username: "SERVER",
	}, content, createdAt)
	msg.IsServer = true
	return msg
}

func (msg *ChatMessage) Show() {
	pc := color.New(s2c(string(msg.Author.ID)))

	fmt.Printf("[%s] <%s> %s: %s\n",
		color.HiBlueString(timeFormat(msg.CreatedAt)),
		pc.Sprint(string(msg.Author.ID)),
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
