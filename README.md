# Redis-like Database (Work in Progress) in Go

A Redis-like in-memory key-value store built from scratch in Go as part of a systems programming course.

The goal of this project is to understand how high-performance, concurrent servers work internally by implementing the core building blocks from scratch, including synchronization, networking, concurrency, operating system I/O, and efficient data structures.

---

## Project Roadmap

- [x] Thread-safe counter using `sync.Mutex`
- [x] TCP server (goroutine-per-connection)
- [x] Worker pool TCP server
- [x] I/O multiplexing TCP server using Linux `epoll`
- [ ] Concurrent queue
- [ ] Redis Serialization Protocol (RESP) parser
- [x] Basic command handling (`PING`, `GET`, `SET`)
- [x] Basic in-memory key-value store
- [ ] Concurrent key-value store (`sync.RWMutex`)
- [ ] Key expiration (TTL)
- [ ] Persistence (RDB)
- [ ] Replication

---

# Week 1 – Thread-safe Counter

## Objective

Implement a thread-safe counter that can be safely incremented by multiple goroutines using Go's `sync.Mutex`.

## Concepts Covered

- Goroutines
- `sync.Mutex`
- `sync.WaitGroup`
- Race conditions
- Critical sections
- Mutual exclusion

## Implementation

The counter exposes two methods:

```go
Increment()
Value()
```

The `Increment()` method acquires a mutex before modifying the shared counter, ensuring that only one goroutine can update the value at a time.

### Example

```
Expected Output:
100
```

---

# Week 2 – TCP Servers

## Objective

Learn how concurrent network servers work by implementing two TCP server architectures from scratch.

---

## 1. Goroutine-per-Connection Server

A TCP server that:

- Listens for incoming client connections
- Spawns one goroutine per client
- Supports multiple requests over a persistent TCP connection
- Parses simple text-based commands
- Executes basic Redis-style commands

### Supported Commands

```
PING
SET key value
GET key
```

### Concepts Covered

- TCP sockets
- `net.Listener`
- `net.Conn`
- `bufio.Reader`
- Connection lifecycle
- Request parsing
- Goroutines
- Long-lived TCP connections

### Example

```
PING
PONG

SET name Lam
OK

GET name
Lam
```

---

## 2. Worker Pool TCP Server

Implemented an alternative server architecture using a fixed-size worker pool.

Instead of creating one goroutine per connection, incoming client connections are pushed onto a shared job queue and processed by a pool of worker goroutines.

### Concepts Covered

- Worker pools
- Channels
- Producer-consumer pattern
- Bounded concurrency
- Job queues
- Goroutine coordination

### Architecture

```
Clients
    │
    ▼
TCP Listener
    │
    ▼
Connection Queue (channel)
    │
 ┌──┼──┬──┐
 ▼  ▼  ▼  ▼
W1 W2 W3 W4
 │  │  │  │
 ▼  ▼  ▼  ▼
Handle Connections
```

---

# Week 3 – I/O Multiplexing Server (epoll)

## Objective

Implement an event-driven TCP server using Linux's `epoll` API to understand how high-performance servers such as Redis and Nginx efficiently manage thousands of concurrent connections without dedicating a thread or goroutine to each client.

### Implementation

Implemented an event loop using Linux system calls:

- `epoll_create1`
- `epoll_ctl`
- `epoll_wait`
- `accept`
- `read`
- `write`
- `close`

The server:

- Creates an epoll instance
- Registers the listening socket
- Accepts new client connections
- Registers client sockets with epoll
- Waits for readable events
- Reads incoming requests
- Parses commands
- Executes commands
- Sends responses back to clients
- Cleans up disconnected clients

### Concepts Covered

- Linux system calls
- File descriptors
- Event-driven programming
- I/O multiplexing
- `epoll`
- Readable socket events
- Socket lifecycle
- Event loops
- Kernel event notification
- Operating system networking

### Architecture

```
                epoll
                  │
                  ▼
            epoll_wait()
                  │
        ┌─────────┴─────────┐
        │                   │
Listener Event        Client Event
        │                   │
        ▼                   ▼
 Accept Client         Read Request
        │                   │
        ▼                   ▼
Register Client      Parse Command
        │                   │
        ▼                   ▼
 Return to Loop      Execute Command
                             │
                             ▼
                      Write Response
```

---

## Project Structure

```text
.
├── go.mod
├── README.md
│
├── week1
│   ├── main.go
│   └── thread_safe_counter
│       └── thread_safe_counter.go
│
├── week2
│   ├── server
│   │   ├── handlers.go
│   │   ├── parser.go
│   │   └── server.go
│   │
│   ├── tcp_server
│   │   └── main.go
│   │
│   └── thread_pool
│       └── main.go
│
└── week3
    ├── server
    │   ├── handlers.go
    │   ├── parser.go
    │   └── server.go
    │
    └── io_multiplexing
        └── main.go
```

---

## Running

### Week 1

```bash
go run ./week1
```

### Week 2 — Goroutine-per-Connection Server

```bash
go run ./week2/tcp_server
```

### Week 2 — Worker Pool Server

```bash
go run ./week2/thread_pool
```

### Week 3 — epoll Event-driven Server (Linux / WSL)

```bash
go run ./week3/io_multiplexing
```

---

## What I'm Learning

This repository documents my progress as I build a Redis-like database from scratch.

Topics covered include:

- Concurrent programming in Go
- Goroutines
- Channels
- Synchronization primitives (`sync.Mutex`)
- TCP networking
- Worker pool design
- Event-driven server architecture
- Linux system calls
- I/O multiplexing (`epoll`)
- Operating system networking
- Request parsing
- Redis command execution
- Protocol design
- High-performance server architecture
- Systems programming concepts
