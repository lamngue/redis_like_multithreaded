package server

import "fmt"

type DataType byte

const (
	Array        DataType = '*'
	BulkString   DataType = '$'
	SimpleString DataType = '+'
	Error        DataType = '-'
	Integer      DataType = ':'
)

type ParseResult struct {
	Request  *Request
	Consumed int
}

func ParseRESP(data []byte) (*ParseResult, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	switch DataType(data[0]) {
	case SimpleString:
		return parseSimpleString(data)
	case Error:
		return parseError(data)
	case Integer:
		return parseInteger(data)
	case BulkString:
		return parseBulkString(data)
	case Array:
		return parseArray(data)
	default:
		return nil, fmt.Errorf("unknown RESP type: %c", data[0])
	}
}

func parseSimpleString(data []byte) (*ParseResult, error) {
	if len(data) < 3 || data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return nil, fmt.Errorf("invalid simple string format")
	}

	content := string(data[1 : len(data)-2])
	return &ParseResult{
		Request: &Request{
			Command: content,
			Args:    []string{},
		},
		Consumed: len(data),
	}, nil
}

func parseError(data []byte) (*ParseResult, error) {
	if len(data) < 3 || data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return nil, fmt.Errorf("invalid error format")
	}

	content := string(data[1 : len(data)-2])
	return &ParseResult{
		Request: &Request{
			Command: content,
			Args:    []string{},
		},
		Consumed: len(data),
	}, nil
}

func parseInteger(data []byte) (*ParseResult, error) {
	if len(data) < 3 || data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return nil, fmt.Errorf("invalid integer format")
	}

	content := string(data[1 : len(data)-2])
	return &ParseResult{
		Request: &Request{
			Command: content,
			Args:    []string{},
		},
		Consumed: len(data),
	}, nil
}

func parseBulkString(data []byte) (*ParseResult, error) {
	if len(data) < 5 || data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return nil, fmt.Errorf("invalid bulk string format")
	}

	lengthEnd := 1
	for lengthEnd < len(data) && data[lengthEnd] != '\r' {
		lengthEnd++
	}

	if lengthEnd == len(data) || data[lengthEnd+1] != '\n' {
		return nil, fmt.Errorf("invalid bulk string format")
	}

	lengthStr := string(data[1:lengthEnd])
	var length int
	_, err := fmt.Sscanf(lengthStr, "%d", &length)
	if err != nil {
		return nil, fmt.Errorf("invalid bulk string length: %v", err)
	}

	if length < 0 {
		return &ParseResult{
			Request: &Request{
				Command: "",
				Args:    []string{},
			},
			Consumed: 0,
		}, nil
	}

	if len(data) < lengthEnd+2+length+2 {
		return nil, fmt.Errorf("bulk string data is shorter than expected")
	}
	consumed := lengthEnd + 2 + length + 2

	content := string(data[lengthEnd+2 : lengthEnd+2+length])
	return &ParseResult{
		Request: &Request{
			Command: content,
			Args:    []string{},
		},
		Consumed: consumed,
	}, nil
}

func parseArray(data []byte) (*ParseResult, error) {
	if len(data) < 5 || data[len(data)-2] != '\r' || data[len(data)-1] != '\n' {
		return nil, fmt.Errorf("invalid array format")
	}

	lengthEnd := 1
	for lengthEnd < len(data) && data[lengthEnd] != '\r' {
		lengthEnd++
	}

	if lengthEnd == len(data) || data[lengthEnd+1] != '\n' {
		return nil, fmt.Errorf("invalid array format")
	}

	lengthStr := string(data[1:lengthEnd])
	var length int
	_, err := fmt.Sscanf(lengthStr, "%d", &length)
	if err != nil {
		return nil, fmt.Errorf("invalid array length: %v", err)
	}

	if length < 0 {
		return &ParseResult{
			Request: &Request{
				Command: "",
				Args:    []string{},
			},
			Consumed: 0,
		}, nil
	}

	args := []string{}
	offset := lengthEnd + 2
	for i := 0; i < length; i++ {
		if offset >= len(data) {
			return nil, fmt.Errorf("array data is shorter than expected")
		}
		parseResult, err := ParseRESP(data[offset:])
		if err != nil {
			return nil, fmt.Errorf("failed to parse array element: %v", err)
		}

		args = append(args, parseResult.Request.Command)
		offset += parseResult.Consumed
	}

	return &ParseResult{
		Request: &Request{
			Command: args[0],
			Args:    args[1:],
		},
		Consumed: offset,
	}, nil
}
