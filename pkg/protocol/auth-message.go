package protocol

import (
	"encoding/json"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/id"
)

type ClientAuthMessage struct {
	Username string `json:"username"`
	RoomID   id.ID  `json:"room_id"`
}

func ClientAuthMessageFromPacket(pkt Packet) (ClientAuthMessage, error) {
	var msg ClientAuthMessage
	if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func (msg ClientAuthMessage) ToPacket() Packet {
	payload, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return NewPacket(PacketTypeAuth, payload)
}

type ServerAuthMessage struct {
	Status  string           `json:"status"`
	Content string           `json:"content"`
	RoomID  id.ID            `json:"room_id"`
	Account *account.Account `json:"account"`
}

func ServerAuthMessageFromPacket(pkt Packet) (ServerAuthMessage, error) {
	var msg ServerAuthMessage
	if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func (msg ServerAuthMessage) ToPacket() Packet {
	payload, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return NewPacket(PacketTypeAuth, payload)
}
