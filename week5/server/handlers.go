package server

import (
	"fmt"
	"strconv"
	"time"
)

func ExecuteCommand(req *Request) (*Response, error) {
	switch req.Command {
	case "PING":
		return handlePing()
	case "SET":
		return handleSet(req)
	case "GET":
		return handleGet(req)
	case "EXPIRE":
		return handleExpire(req)
	case "EXISTS":
		return handleExists(req)
	case "TTL":
		return handleTTL(req)
	default:
		return &Response{}, fmt.Errorf("ERR unknown command")
	}
}

type Entry struct {
	Value     string
	ExpiresAt int64 // Unix timestamp in seconds
}

var store = map[string]Entry{}

func handleSet(req *Request) (*Response, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("SET command requires 2 arguments")
	}

	key := req.Args[0]
	value := req.Args[1]

	store[key] = Entry{
		Value:     value,
		ExpiresAt: 0, // No expiration
	}
	return &Response{
		Type:  BulkStringRes,
		Value: "OK",
	}, nil
}

func checkExpiration(key string) int64 {
	entry, ok := store[key]
	if !ok {
		return -2
	}

	if entry.ExpiresAt == 0 {
		return -1 // No expiration
	}
	now := time.Now().Unix()
	if now >= entry.ExpiresAt {
		delete(store, key)
		return -2 // Key has expired
	}

	ttl := entry.ExpiresAt - now

	return ttl // Key is valid
}

func handleGet(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("GET command requires 1 argument")
	}
	value := checkExpiration(req.Args[0])
	if value == -2 {
		return nil, fmt.Errorf("key not found")
	}

	return &Response{
		Type:  BulkStringRes,
		Value: store[req.Args[0]].Value,
	}, nil
}

func handlePing() (*Response, error) {
	return &Response{
		Type:  BulkStringRes,
		Value: "PONG",
	}, nil
}

func handleExpire(req *Request) (*Response, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("EXPIRE command requires 2 arguments")
	}

	key := req.Args[0]
	seconds, err := strconv.Atoi(req.Args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid expiration time")
	}

	entry, _ := store[key]
	expired := checkExpiration(key)
	if expired == -2 {
		return &Response{
			Type:  IntegerRes,
			Value: "0",
		}, nil
	}

	entry.ExpiresAt = time.Now().Unix() + int64(seconds)
	store[key] = entry

	return &Response{
		Type:  IntegerRes,
		Value: "1",
	}, nil
}

func handleExists(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("EXISTS command requires 1 argument")
	}

	key := req.Args[0]
	value := checkExpiration(key)

	if value == -2 {
		return &Response{
			Type:  IntegerRes,
			Value: "0", // Key does not exist
		}, nil
	}

	return &Response{
		Type:  IntegerRes,
		Value: "1", // Key exists
	}, nil
}

func handleTTL(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("TTL command requires 1 argument")
	}

	key := req.Args[0]
	ttl := checkExpiration(key)

	return &Response{
		Type:  IntegerRes,
		Value: fmt.Sprintf("%d", ttl),
	}, nil
}
