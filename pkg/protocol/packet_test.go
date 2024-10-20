package protocol

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPacketToBinary(t *testing.T) {
	originalPacket := NewPacket(PacketTypeMessage, []byte("hello"))

	data, err := originalPacket.ToBinary()
	assert.Nil(t, err)

	expectedLengthInBytes := 1 + 1 + 2 + 5 // Version + Type + Len + Payload
	assert.Equal(t, expectedLengthInBytes, len(data))
}

func TestPacketFromBytes(t *testing.T) {
	originalPacket := NewPacket(PacketTypeMessage, []byte("hello"))

	data, err := originalPacket.ToBinary()
	assert.Nil(t, err)

	parsedPacket, err := PacketFromBytes(data)
	assert.Nil(t, err)

	assert.Equal(t, originalPacket.Header.Version, parsedPacket.Header.Version)
	assert.Equal(t, originalPacket.Header.PacketType, parsedPacket.Header.PacketType)
	assert.Equal(t, originalPacket.Header.Len, parsedPacket.Header.Len)
	assert.Equal(t, originalPacket.Payload, parsedPacket.Payload)
}
