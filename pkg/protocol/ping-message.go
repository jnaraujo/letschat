package protocol

import "encoding/json"

type PingMessage struct {
}

func PingMessageFromPacket(pkt *Packet) (PingMessage, error) {
	var msg PingMessage
	if err := json.Unmarshal(pkt.Payload, &msg); err != nil {
		return msg, err
	}
	return msg, nil
}

func (msg PingMessage) ToPacket() *Packet {
	payload, err := json.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return NewPacket(PacketTypePing, payload)
}
