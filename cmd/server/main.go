package main

import (
	"fmt"

	"github.com/jnaraujo/letschat/pkg/server"
)

func main() {
	fmt.Println("Starting server on port 3000")
	server := server.NewServer()
	err := server.Run(":3000")
	if err != nil {
		panic(err)
	}
}
