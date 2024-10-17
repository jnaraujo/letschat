package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/jnaraujo/letschat/pkg/account"
	"github.com/jnaraujo/letschat/pkg/client"
	"github.com/jnaraujo/letschat/pkg/message"
)

func main() {
	fmt.Println("==================== LetsChat ====================")
	fmt.Println("Welcome to LetsChat. Insert your credentials below")
	fmt.Println("to log in.")

	fmt.Println("Trying to connect to the server...")
	client := client.NewWSClient(":3000")
	err := client.Connect()
	if err != nil {
		fmt.Println("Failed to connect to the server.", err)
		return
	}
	fmt.Println("Connected successfully.")

	fmt.Printf("Enter your username: ")
	scanner := bufio.NewScanner(os.Stdin)
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
			incomingMsg.Show()
		}
	}()

	for scanner.Scan() {
		content := scanner.Text()
		content = strings.TrimSpace(content)
		clearLine()
		err := client.WriteMessage(
			&message.ChatMessage{
				Content: content,
			},
		)
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
