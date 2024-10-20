package protocol

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var ErrProtocolVersionMismatch = errors.New("protocol version mismatch")

type PacketProtocolVersion uint8

const (
	ProtocolVersion PacketProtocolVersion = 1
)

type PacketType uint8

const (
	PacketTypeAuth PacketType = iota
	PacketTypeMessage
	PacketTypePing
	PacketTypePong
)

type PacketHeader struct {
	Version    PacketProtocolVersion
	PacketType PacketType
	Len        uint16
}

type Packet struct {
	Header  PacketHeader
	Payload []byte
}

func NewPacket(pktType PacketType, payload []byte) Packet {
	return Packet{
		Header: PacketHeader{
			Version:    ProtocolVersion,
			PacketType: pktType,
			Len:        uint16(len(payload)),
		},
		Payload: payload,
	}
}

func PacketFromBytes(data []byte) (Packet, error) {
	buf := bytes.NewReader(data)
	var pkt Packet

	if err := binary.Read(buf, binary.BigEndian, &pkt.Header.Version); err != nil {
		return pkt, err
	}

	if pkt.Header.Version != ProtocolVersion {
		return pkt, ErrProtocolVersionMismatch
	}

	if err := binary.Read(buf, binary.BigEndian, &pkt.Header.PacketType); err != nil {
		return pkt, err
	}

	if err := binary.Read(buf, binary.BigEndian, &pkt.Header.Len); err != nil {
		return pkt, err
	}

	pkt.Payload = make([]byte, pkt.Header.Len)
	if err := binary.Read(buf, binary.BigEndian, &pkt.Payload); err != nil {
		return pkt, err
	}

	return pkt, nil
}

func (pkt Packet) ToBinary() ([]byte, error) {
	var buf bytes.Buffer

	if err := binary.Write(&buf, binary.BigEndian, pkt.Header.Version); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, pkt.Header.PacketType); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, pkt.Header.Len); err != nil {
		return nil, err
	}

	if err := binary.Write(&buf, binary.BigEndian, pkt.Payload); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
