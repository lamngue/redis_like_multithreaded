package main

import (
	"log"
	"redis-like-multithreaded/week3+4/server"
)

func main() {
	server, err := server.NewServer(8080)
	if err != nil {
		log.Fatal(err)
	}
	server.Run()
}
