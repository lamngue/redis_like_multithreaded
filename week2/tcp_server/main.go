package main

import (
	"fmt"
	"log"
	"net"
	"redis-like-multithreaded/week2/server"
)

func main() {
	port := ":6379"

	// Create a TCP server that listens on port 6379
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("TCP Server listening on %s...\n", port)

	defer listener.Close()
	// Accept incoming connections and handle them in separate goroutines
	for {
		clientConn, err := listener.Accept()
		// For each connection, read data from the client and send a response back
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
			continue
		}
		go server.HandleConnection(clientConn)
	}
}
