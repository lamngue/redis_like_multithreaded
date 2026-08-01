package server

import (
	"fmt"
	"strings"
)

type Request struct {
	Command string
	Args    []string
}

func parseRequest(message string) (*Request, error) {
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	return &Request{
		Command: strings.ToUpper(parts[0]),
		Args:    parts[1:],
	}, nil
}
