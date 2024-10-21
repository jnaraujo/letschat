package main

import (
	"fmt"

	"github.com/jnaraujo/letschat/pkg/server"
)

const (
	addr = ":2257"
)

func main() {
	fmt.Printf("Starting server on %s", addr)
	server := server.NewServer()
	err := server.Run(addr)
	if err != nil {
		panic(err)
	}
}
