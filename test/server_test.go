package test

import (
	"testing"
	"time"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/client"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/message"
	"github.com/stretchr/testify/assert"
)

func BenchmarkWriteMessageServer(b *testing.B) {
	client := client.NewWSClient("ws://localhost:3000/ws")
	err := client.Connect()
	assert.Nil(b, err)

	exampleMessage := message.NewChatMessage(
		account.NewAccount("test"), "example", message.CharRoom{
			ID:   id.NewID(22),
			Name: "test",
		}, time.Now(),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.WriteMessage(exampleMessage)
	}
}
