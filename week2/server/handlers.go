package server

import "fmt"

func handleSet(req *Request) (string, error) {
	if len(req.Args) != 2 {
		return "", fmt.Errorf("SET command requires 2 arguments")
	}

	key := req.Args[0]
	value := req.Args[1]

	store[key] = value
	return "OK", nil
}

func handleGet(req *Request) (string, error) {
	if len(req.Args) != 1 {
		return "", fmt.Errorf("GET command requires 1 argument")
	}
	value, ok := store[req.Args[0]]
	if !ok {
		return "", fmt.Errorf("key not found")
	}

	return value, nil
}

func handlePing() (string, error) {
	return "PONG", nil
}
