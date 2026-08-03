# Redis-like Database (Work in Progress) in Go

A Redis-like in-memory key-value store built from scratch in Go as part of a systems programming course.

The goal of this project is to understand how high-performance, concurrent servers work internally by implementing the core building blocks from scratch, including synchronization, networking, concurrency, and efficient data structures.

## Project Roadmap

- [x] Thread-safe counter using `sync.Mutex`
- [x] TCP server
- [x] Worker pool TCP server
- [ ] Concurrent queue
- [ ] Redis Serialization Protocol (RESP) parser
- [ ] Command handling (`PING`, `GET`, `SET`)
- [ ] In-memory key-value store
- [ ] Concurrent client handling
- [ ] Key expiration (TTL)
- [ ] Persistence
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

Running the demo launches 100 goroutines, each incrementing the shared counter once.

```
Expected Output:
100
```

---

# Week 2 – TCP Servers

## Objective

Learn how concurrent network servers work by implementing two TCP server architectures from scratch.

### 1. Goroutine-per-Connection Server

A TCP server that:

- Listens for incoming client connections.
- Spawns one goroutine for each connected client.
- Allows clients to send multiple requests over the same TCP connection.
- Parses simple text-based commands.
- Supports the following commands:
  - `PING`
  - `SET key value`
  - `GET key`

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

### 2. Worker Pool TCP Server

Implemented an alternative server architecture using a fixed-size worker pool.

Instead of creating one goroutine per client connection, incoming connections are placed onto a shared job queue and processed by a pool of worker goroutines.

### Concepts Covered

- Worker pools
- Channels
- Producer-consumer pattern
- Bounded concurrency
- Job queues
- Goroutine coordination

Architecture:

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

## Project Structure

```text
.
├── go.mod
├── README.md
├── week1
│   ├── main.go
│   └── thread_safe_counter
│       └── thread_safe_counter.go
│
└── week2
    ├── server
    │   ├── handlers.go
    │   ├── parser.go
    │   └── server.go
    │
    ├── tcp_server
    │   └── main.go
    │
    └── thread_pool
        └── main.go
```

---

## Running

### Week 1

```bash
go run ./week1
```

### TCP Server

```bash
go run ./week2/tcp_server
```

### Worker Pool Server

```bash
go run ./week2/thread_pool
```

---

## What I'm Learning

This repository documents my progress as I build a Redis-like server from scratch.

Topics covered include:

- Concurrent programming in Go
- Goroutines and channels
- Synchronization primitives (`sync.Mutex`)
- TCP networking
- Worker pool design
- Producer-consumer patterns
- Request parsing
- Protocol design
- High-performance server architecture
- Systems programming concepts
