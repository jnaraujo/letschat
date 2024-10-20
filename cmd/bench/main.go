package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jnaraujo/letschat/pkg/client"
	"github.com/jnaraujo/letschat/pkg/protocol"
)

func main() {
	maxClients := 1000
	connectionsPerClient := 10

	start := time.Now()
	var wg sync.WaitGroup
	wg.Add(maxClients)

	for i := range maxClients {
		go func() {
			defer wg.Done()
			handleClient(i, connectionsPerClient)
		}()
	}
	wg.Wait()
	total := time.Since(start)

	fmt.Println(`Total: `, maxClients*connectionsPerClient)
	fmt.Println(`Msg/Sec:`, maxClients*connectionsPerClient/int(total.Seconds()))
}

func handleClient(id, N int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	client := client.NewWSClient("ws://localhost:3000/ws")
	err := client.Connect(ctx)
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Auth
	err = client.WritePacket(protocol.ClientAuthMessage{
		Username: fmt.Sprintf("username-%d", id),
	}.ToPacket())
	if err != nil {
		panic(err)
	}

	pkt, _ := client.ReadPacket()
	serverAuthMsg, _ := protocol.ServerAuthMessageFromPacket(pkt)

	exampleMessagePkt := protocol.NewChatMessage(
		serverAuthMsg.Account, fmt.Sprintf("example message %d", id),
		protocol.ChatRoom{
			ID: "ALL",
		}, time.Now(),
	).ToPacket()

	go func() {
		for {
			_, err = client.Read() // Ignora o tipo de mensagem e os dados
			if err != nil {
				return
			}
		}
	}()

	for range N {
		client.WritePacket(exampleMessagePkt)
	}
}
