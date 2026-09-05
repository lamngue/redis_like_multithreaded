package server

import (
	"bytes"
	"fmt"
)

type ResponseType int

const (
	SimpleStringRes ResponseType = iota
	BulkStringRes
	ErrorRes
	IntegerRes
	Null
)

type Response struct {
	Type  ResponseType
	Value string
}

func EncodeResponse(res *Response, err error) ([]byte, error) {
	if err != nil {
		return EncodeError(err.Error()), nil
	}

	switch res.Type {
	case SimpleStringRes:
		return EncodeSimpleString(res.Value), nil
	case BulkStringRes:
		return EncodeBulkString(res.Value), nil
	case ErrorRes:
		return EncodeError(res.Value), nil
	case IntegerRes:
		return fmt.Appendf(nil, ":%s\r\n", res.Value), nil
	case Null:
		return []byte("$-1\r\n"), nil
	default:
		return nil, fmt.Errorf("unknown response type")

	}
}
func EncodeSimpleString(s string) []byte {
	return []byte("+" + s + "\r\n")
}

func EncodeError(s string) []byte {
	return []byte("-" + s + "\r\n")
}

func EncodeBulkString(s string) []byte {
	return fmt.Appendf(nil, "$%d\r\n%s\r\n", len(s), s)
}

func EncodeArray(arr []string) []byte {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("*%d\r\n", len(arr)))
	for _, s := range arr {
		buf.Write(EncodeBulkString(s))
	}
	return buf.Bytes()
}
