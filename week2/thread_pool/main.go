package main

import (
	"fmt"
	"log"
	"net"
	"redis-like-multithreaded/week2/server"
)

const workerCount = 4

func worker(jobs <-chan net.Conn) {
	for conn := range jobs {
		server.HandleConnection(conn)
	}
}

func main() {
	port := ":8080"
	listener, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Can not start server: %v", err)
	}
	defer listener.Close()

	fmt.Printf("Thread Pool Server listening on %s...\n", port)
	jobs := make(chan net.Conn, 100) // Channel to hold incoming connections
	for i := 0; i < workerCount; i++ {
		go worker(jobs) // Start 4 worker goroutines
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		fmt.Printf("New client: %s. Inserting to queue...\n", conn.RemoteAddr().String())
		jobs <- conn // Send the connection to the jobs channel for processing
	}
}
