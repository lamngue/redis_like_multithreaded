package server

import "fmt"

func ExecuteCommand(req *Request) (*Response, error) {
	switch req.Command {
	case "PING":
		return handlePing()
	case "SET":
		return handleSet(req)
	case "GET":
		return handleGet(req)
	default:
		return &Response{}, fmt.Errorf("ERR unknown command")
	}
}

var store = map[string]string{}

func handleSet(req *Request) (*Response, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("SET command requires 2 arguments")
	}

	key := req.Args[0]
	value := req.Args[1]

	store[key] = value
	return &Response{
		Type:  BulkStringRes,
		Value: "OK",
	}, nil
}

func handleGet(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("GET command requires 1 argument")
	}
	value, ok := store[req.Args[0]]
	if !ok {
		return nil, fmt.Errorf("key not found")
	}

	return &Response{
		Type:  BulkStringRes,
		Value: value,
	}, nil
}

func handlePing() (*Response, error) {
	return &Response{
		Type:  BulkStringRes,
		Value: "PONG",
	}, nil
}
