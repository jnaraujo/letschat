package main

import (
	"bufio"
	"fmt"
	"os"

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

	err = client.Write([]byte(username))
	if err != nil {
		fmt.Println("Failed to send message.", err)
		return
	}

	data, err := client.Read()
	if err != nil {
		fmt.Println("Failed to read message.", err)
		return
	}
	if string(data) != "ok" {
		fmt.Println("Failed to login.", string(data))
		return
	}

	go func() {
		for {
			msg, err := client.ReadMessage()
			if err != nil {
				fmt.Println("Failed to read message.", err)
				break
			}
			msg.Show()
		}
	}()

	for scanner.Scan() {
		content := scanner.Text()
		err := client.WriteMessage(
			message.NewClientMessage(content),
		)
		if err != nil {
			fmt.Println("Failed to send message.", err)
			continue
		}
	}
}
