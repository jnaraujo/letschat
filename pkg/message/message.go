package message

import "github.com/jnaraujo/letschat/pkg/id"

type BaseMessage struct {
	ID   id.ID `json:"id"`
	Data any   `json:"data"`
}

func NewBaseMessage(data any) *BaseMessage {
	return &BaseMessage{
		ID:   id.NewID(16),
		Data: data,
	}
}
