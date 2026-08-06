package server

import (
	"fmt"
	"syscall"
)

type Server struct {
	epollFd  int
	listener int
	events   []syscall.EpollEvent
}

func NewServer(port int) (*Server, error) {
	epollFd, err := syscall.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("failed to create epoll: %v", err)
	}
	fmt.Println("Epoll created with fd:", epollFd)

	listener, err := createListener(port)
	if err != nil {
		return nil, fmt.Errorf("failed to create listener: %v", err)
	}

	err = registerListenerToEpoll(epollFd, listener)
	if err != nil {
		return nil, fmt.Errorf("failed to register listener to epoll: %v", err)
	}

	return &Server{
		epollFd:  epollFd,
		listener: listener,
		events:   make([]syscall.EpollEvent, 10),
	}, nil
}

func createListener(port int) (int, error) {
	listener, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, 0)
	err = syscall.Bind(listener, &syscall.SockaddrInet4{Port: port})
	if err != nil {
		panic(err)
	}
	err = syscall.Listen(listener, 128)
	if err != nil {
		panic(err)
	}
	fmt.Println("TCP listener created with fd:", listener)
	return listener, nil
}

func registerListenerToEpoll(epollFd int, listener int) error {
	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(listener),
	}
	err := syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, listener, &event)
	if err != nil {
		panic(err)
	}
	return nil
}

func acceptClient(listener int, epollFd int) (int, error) {
	clientFd, _, err := syscall.Accept(listener)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New connection accepted with fd: %d\n", clientFd)
	// register new connection fd to epoll
	event := syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(clientFd),
	}
	err = syscall.EpollCtl(epollFd, syscall.EPOLL_CTL_ADD, clientFd, &event)
	if err != nil {
		panic(err)
	}
	return clientFd, nil
}

func (s *Server) Run() {
	events := make([]syscall.EpollEvent, 10)
	for {
		n, err := syscall.EpollWait(s.epollFd, events, -1)
		if err != nil {
			panic(err)
		}
		for i := 0; i < n; i++ {
			fmt.Printf("Event received for fd: %d\n", events[i].Fd)
			if events[i].Fd == int32(s.listener) {
				// accept new connection
				s.acceptClient()
			} else {
				s.handleClientEvent(events[i])
			}
		}
	}
}

func (s *Server) acceptClient() {
	clientFd, _, err := syscall.Accept(s.listener)
	if err != nil {
		panic(err)
	}
	fmt.Printf("New connection accepted with fd: %d\n", clientFd)
	event := syscall.EpollEvent{Events: syscall.EPOLLIN, Fd: int32(clientFd)}
	err = syscall.EpollCtl(s.epollFd, syscall.EPOLL_CTL_ADD, clientFd, &event)
	if err != nil {
		panic(err)
	}
}

func (s *Server) handleClientEvent(event syscall.EpollEvent) {
	buf := make([]byte, 1024)
	n, err := syscall.Read(int(event.Fd), buf)
	if err != nil {
		panic(err)
	}
	// if disconnected, remove fd from epoll
	if n == 0 {
		fmt.Printf("Client disconnected: fd %d\n", event.Fd)
		err = syscall.EpollCtl(s.epollFd, syscall.EPOLL_CTL_DEL, int(event.Fd), nil)
		// close socket
		syscall.Close(int(event.Fd))
	}
	fmt.Printf("Read %d bytes from fd %d: %q\n", n, event.Fd, string(buf[:n]))
	message := string(buf[:n])
	req, err := ParseRESP([]byte(message))
	if err != nil {
		fmt.Printf("Error parsing request: %v\n", err)
	}

	res, err := ExecuteCommand(req.Request)
	bytes, err := EncodeResponse(res, err)
	// send response back to client
	_, err = syscall.Write(int(event.Fd), bytes)
	if err != nil {
		fmt.Printf("Error writing response: %v\n", err)
	}
}
