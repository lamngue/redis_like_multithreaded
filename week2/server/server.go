package server

import (
	"bufio"
	"io"
	"log"
	"net"
	"strings"
)

var store = map[string]string{}

func HandleConnection(clientConn net.Conn) {
	defer clientConn.Close()
	reader := bufio.NewReader(clientConn)
	for {
		// Read data from the client
		message, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
			} else {
				log.Println(err)
			}
			return
		}
		message = strings.TrimSpace(message)
		log.Printf("Received message: %s", message)
		req, err := parseRequest(message)
		if err != nil {
			log.Println(err)
			return
		}
		response := ""
		switch req.Command {
		case "PING":
			response, err = handlePing()
		case "SET":
			response, err = handleSet(req)
		case "GET":
			response, err = handleGet(req)
		default:
			response = "ERR unknown command"
		}

		// Send a response back to the client
		if err != nil {
			response = "ERR " + err.Error()
		}

		_, err = clientConn.Write([]byte(response + "\n"))
		if err != nil {
			log.Println(err)
			return
		}
	}
}
