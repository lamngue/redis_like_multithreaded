package server

import (
	"fmt"
	"strconv"
	"sync"
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
	case "DEL":
		return handleDel(req)
	default:
		return &Response{}, fmt.Errorf("ERR unknown command")
	}
}

type Entry struct {
	Value     string
	ExpiresAt int64 // Unix timestamp in seconds
}

type Store struct {
	mu   sync.RWMutex
	data map[string]Entry
}

var store = &Store{
	data: make(map[string]Entry),
}

func (s *Store) Set(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = Entry{
		Value:     value,
		ExpiresAt: 0, // No expiration
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	entry, exists := s.data[key]
	if !exists {
		s.mu.Unlock()
		return "", false
	}
	now := time.Now().Unix()
	if entry.ExpiresAt > 0 && now >= entry.ExpiresAt {
		delete(s.data, key)
		s.mu.Unlock()
		return "", false
	}
	s.mu.Unlock()
	return entry.Value, true
}

func (s *Store) Delete(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.data[key]
	if !exists {
		return false
	}
	now := time.Now().Unix()
	if entry.ExpiresAt > 0 && now >= entry.ExpiresAt {
		delete(s.data, key)
		return false
	}
	delete(s.data, key)
	return true
}

func (s *Store) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.data[key]
	if !exists {
		return false
	}
	now := time.Now().Unix()
	if entry.ExpiresAt > 0 && now >= entry.ExpiresAt {
		delete(s.data, key)
		return false
	}
	return true
}

func (s *Store) TTL(key string) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, exists := s.data[key]
	if !exists {
		return -2 // Key does not exist
	}
	if entry.ExpiresAt == 0 {
		return -1 // Key exists but has no expiration
	}
	now := time.Now().Unix()
	ttl := entry.ExpiresAt - now
	if now >= entry.ExpiresAt {
		delete(s.data, key)
		return -2 // Key has expired
	}
	return ttl
}

func (s *Store) SetExpiration(key string, seconds int64) bool {
	s.mu.Lock()
	entry, exists := s.data[key]
	if !exists {
		s.mu.Unlock()
		return false
	}
	now := time.Now().Unix()
	if entry.ExpiresAt > 0 && now >= entry.ExpiresAt {
		delete(s.data, key)
		s.mu.Unlock()
		return false
	}
	entry.ExpiresAt = now + seconds
	s.data[key] = entry
	s.mu.Unlock()
	return true
}

func handleSet(req *Request) (*Response, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("SET command requires 2 arguments")
	}

	key := req.Args[0]
	value := req.Args[1]
	store.Set(key, value)
	return &Response{
		Type:  BulkStringRes,
		Value: "OK",
	}, nil
}

func handleGet(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("GET command requires 1 argument")
	}
	value, exists := store.Get(req.Args[0])
	if !exists {
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

func handleExpire(req *Request) (*Response, error) {
	if len(req.Args) != 2 {
		return nil, fmt.Errorf("EXPIRE command requires 2 arguments")
	}

	key := req.Args[0]
	seconds, err := strconv.Atoi(req.Args[1])
	if err != nil {
		return nil, fmt.Errorf("invalid expiration time")
	}

	set := store.SetExpiration(key, int64(seconds))
	if !set {
		return &Response{
			Type:  IntegerRes,
			Value: "0", // Key does not exist
		}, nil
	}

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
	exists := store.Exists(key)
	return &Response{
		Type:  IntegerRes,
		Value: fmt.Sprintf("%d", boolToInt(exists)),
	}, nil
}

func handleTTL(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("TTL command requires 1 argument")
	}

	key := req.Args[0]

	ttl := store.TTL(key)
	return &Response{
		Type:  IntegerRes,
		Value: fmt.Sprintf("%d", ttl),
	}, nil
}

func handleDel(req *Request) (*Response, error) {
	if len(req.Args) != 1 {
		return nil, fmt.Errorf("DEL command requires 1 argument")
	}

	key := req.Args[0]

	res := store.Delete(key)
	return &Response{
		Type:  IntegerRes,
		Value: fmt.Sprintf("%d", boolToInt(res)),
	}, nil
}

func boolToInt(exists bool) int {
	if exists {
		return 1
	}
	return 0
}
