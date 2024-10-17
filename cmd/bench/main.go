package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/client"
	"github.com/jnaraujo/letschat/pkg/message"
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
	fmt.Println(`BenchmarkWriteMessageServer took:`, total.Milliseconds())
}

func handleClient(id, N int) {
	client := client.NewWSClient(":3000")
	err := client.Connect()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	// Auth
	err = client.WriteMessage(message.AuthMessageClient{
		Username: fmt.Sprintf("username-%d", id),
	})
	if err != nil {
		panic(err)
	}

	var serverAuthMsg message.AuthMessageServer
	err = client.ReadMessage(&serverAuthMsg)
	if err != nil {
		fmt.Println("Failed to read message.", err)
		return
	}
	if serverAuthMsg.Status != "ok" {
		fmt.Println("Failed to login.", serverAuthMsg.Content)
		return
	}

	var account account.Account
	err = client.ReadMessage(&account)
	if err != nil {
		fmt.Println("Failed to read message.", err)
		return
	}

	exampleMessage := message.NewChatMessage(
		&account, fmt.Sprintf("example message %d", id), time.Now(),
	)

	go func() {
		for {
			_, err = client.Read() // Ignora o tipo de mensagem e os dados
			if err != nil {
				return
			}
		}
	}()

	for range N {
		client.WriteMessage(exampleMessage)
	}
}
