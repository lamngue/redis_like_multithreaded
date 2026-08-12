package server

import (
	"fmt"
	"strings"
)

type Request struct {
	Command string
	Args    []string
}

func ParseRequest(message string) (*Request, error) {

	message = strings.TrimSpace(message)
	if message == "" {
		return nil, fmt.Errorf("empty request")
	}

	parts := strings.Fields(message)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	return &Request{
		Command: strings.ToUpper(parts[0]),
		Args:    parts[1:],
	}, nil
}
