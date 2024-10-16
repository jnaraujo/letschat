package main

import "github.com/jnaraujo/letschat/pkg/server"

func main() {
	server := server.NewServer()
	err := server.Run(":3000")
	if err != nil {
		panic(err)
	}
}
