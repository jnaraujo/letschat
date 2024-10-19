package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/client"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/message"
)

const (
	defaultAddr = "ws://localhost:3000/ws"
)

func main() {
	fmt.Println("==================== LetsChat ====================")
	fmt.Println("Welcome to LetsChat. Insert your credentials below")
	fmt.Println("to log in.")

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Printf("Server address (defaults to %s): ", defaultAddr)
	scanner.Scan()
	addr := scanner.Text()
	if addr == "" {
		addr = defaultAddr
	}

	fmt.Printf("Trying to connect to %s...\n", addr)
	client := client.NewWSClient(addr)
	err := client.Connect()
	if err != nil {
		fmt.Println("Failed to connect to the server.", err)
		return
	}
	defer client.Close()

	fmt.Println("Connected successfully.")

	fmt.Printf("Your Username: ")
	scanner.Scan()
	username := scanner.Text()

	err = client.WriteMessage(message.AuthMessageClient{
		Username: username,
	})
	if err != nil {
		fmt.Println("Failed to send message.", err)
		return
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

	go func() {
		for {
			var incomingMsg message.ChatMessage
			err := client.ReadMessage(&incomingMsg)
			if err != nil {
				fmt.Println("Failed to read message.", err)
				break
			}

			// TODO: fix this
			if incomingMsg.IsCommand && strings.HasPrefix(incomingMsg.Content, "Pong") {
				incomingMsg.Content = fmt.Sprintf("Pong! %d ms", time.Since(incomingMsg.CreatedAt).Milliseconds())
			}

			incomingMsg.Show()
		}
	}()

	for scanner.Scan() {
		content := scanner.Text()
		content = strings.TrimSpace(content)

		msg := message.NewChatMessage(
			&account, content, message.CharRoom{
				ID: id.ID("ALL"),
			}, time.Now(),
		)
		if strings.HasPrefix(content, "/") {
			msg.Content = msg.Content[1:]
			msg.IsCommand = true
		} else {
			clearLine()
		}

		err := client.WriteMessage(msg)
		if err != nil {
			fmt.Println("Failed to send message.", err)
			continue
		}
	}
}

func clearLine() {
	fmt.Print("\033[1A") // move cursor one line up
	fmt.Print("\033[K")  // clear the line
}
