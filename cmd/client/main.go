package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jnaraujo/letschat/pkg/client"
	"github.com/jnaraujo/letschat/pkg/id"
	"github.com/jnaraujo/letschat/pkg/protocol"
)

const (
	defaultAddr = "ws://localhost:2257/lc"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fmt.Printf("Trying to connect to %s...\n", addr)
	client := client.NewWSClient(addr)
	err := client.Connect(ctx)
	if err != nil {
		fmt.Println("Failed to connect to the server.", err)
		return
	}
	defer client.Close()

	fmt.Println("Connected successfully.")

	fmt.Printf("Your Username: ")
	scanner.Scan()
	username := scanner.Text()

	err = client.WritePacket(protocol.ClientAuthMessage{
		Username: username,
	}.ToPacket())
	if err != nil {
		fmt.Println("Failed to send message.", err)
		return
	}

	pkt, err := client.ReadPacket()
	if err != nil {
		fmt.Println("Failed to read message.", err)
		return
	}
	serverAuthMsg, err := protocol.ServerAuthMessageFromPacket(pkt)
	if err != nil {
		fmt.Println("Failed to server auth from packet.", err)
		return
	}
	if serverAuthMsg.Status != "ok" {
		fmt.Println("Failed to login.", serverAuthMsg.Content)
		return
	}

	go func() {
		for {
			inPkt, err := client.ReadPacket()
			if err != nil {
				fmt.Println("Failed to read message.", err)
				break
			}
			incomingMsg, err := protocol.ChatMessageFromPacket(inPkt)
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

		msg := protocol.NewChatMessage(
			serverAuthMsg.Account, content, protocol.ChatRoom{
				ID: id.ID("ALL"),
			}, time.Now(),
		)
		if strings.HasPrefix(content, "/") {
			msg.Content = msg.Content[1:]
			msg.IsCommand = true
		} else {
			clearLine()
		}

		err := client.WritePacket(msg.ToPacket())
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
