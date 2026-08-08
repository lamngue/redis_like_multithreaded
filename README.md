# Redis Clone in Go

## What is this?

This project purpose is to (attempt to) create a redis-like in-memory key-value database from scratch using Golang.
I want to understand how high-performance and concurrent server works internally by implementing synchronization, networking, concurrency, operating system I/O, protocol parsing, and data storage from scratch.
This repo will document my progress so far (currently week 4 - parsing RESP).
## Project Roadmap

- [x] Thread-safe counter using `sync.Mutex`
- [x] TCP server (goroutine-per-connection)
- [x] Worker pool TCP server
- [x] I/O multiplexing TCP server using Linux `epoll`
- [x] Basic Redis Serialization Protocol (RESP) parser
- [x] RESP response encoder
- [x] Basic command handling (`PING`, `GET`, `SET`)
- [x] Basic in-memory key-value store
- [x] Basic `redis-cli` communication
- [ ] Per-client buffering / partial TCP reads
- [ ] Concurrent queue
- [ ] Concurrent key-value store (`sync.RWMutex`)
- [ ] Key expiration (TTL)
- [ ] Persistence (RDB)
- [ ] Replication


# Week 1 - Thread-safe Counter
Why do I neeed a mutex? To prevent multiple threads from changing or reading the same data at the same time. (avoiding race conditions and data corruptions)
What race condition am I preventing? Multiple threads (goroutine) are trying to increment a counter value
Why do I need a WaitGroup? To manage the goroutines running. The `Add(1)` method is used to start a new goroutine, `Wait()` ensures that the main goroutine waits for all other goroutines to finish before it proceeds, `Done()` is called when a goroutine is finished.

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

```text
Expected Output:
100
```
---

# Week 2 - TCP Servers

What happens when a client connects? The server listens for incoming TCP connections
Why does `go handleConnection(conn)` allow multiple clients? The `go` keyword starts the function in a new goroutine. This allows the main loop to listening for another client via `Accept()`

## 1. Goroutine-per-Connection Server

Implemented a TCP server using Go's standard networking abstractions.

The server:

- Listens for incoming TCP connections
- Spawns one goroutine per connected client
- Supports multiple requests over persistent connections
- Parses simple text-based commands
- Executes basic Redis-style commands

### Supported Commands

```text
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

```text
PING
PONG

SET name Lam
OK

GET name
Lam
```
## 2. Worker Pool TCP Server
---
What problem does the worker pool solve? To prevent creating a new goroutine for every connection
What is the producer and what is the consumer? Producer - generate jobs (in this case incoming connections). Consumer - fixed number of goroutines pulling elements from the jobs channel.
Implemented an alternative server architecture using a fixed-size worker pool.

Instead of creating a new goroutine for every connection, incoming client connections are placed onto a shared job queue and processed by a fixed number of worker goroutines.

### Concepts Covered

- Worker pools
- Channels
- Producer-consumer pattern
- Bounded concurrency
- Job queues
- Goroutine coordination

### Architecture

```text
Clients
    │
    ▼
TCP Listener
    │
    ▼
Connection Queue
    │
 ┌──┼──┬──┐
 ▼  ▼  ▼  ▼
W1 W2 W3 W4
 │  │  │  │
 ▼  ▼  ▼  ▼
Handle Connections
```
---
# Week 3 - I/O Multiplexing
What exactly is a file descriptor? File descriptors (FDs) are part of the POSIX API and use basic integers to determine state. It is a handle to access IO/file resource at kernel level.
What does epoll_create1() create? It creates a new epoll instance in the Linux kernel and returns an integer file descriptor referring to that instance.
Why do I register the listener with epoll? To bypass the kernel-level blocking behavior of synchronous I/O.
Why does epoll_wait() tell me which sockets are ready?
Why don't I need a goroutine per connection?
What's the difference between the listener FD and a client FD?
What happens when read() returns 0?

## Objective

Implement an event-driven TCP server using Linux's `epoll` API to understand how high-performance servers can efficiently manage many concurrent connections without dedicating a thread or goroutine to every connected client.

## Implementation

Implemented an event loop using low-level Linux system calls including:

- `epoll_create1`
- `epoll_ctl`
- `epoll_wait`
- `socket`
- `bind`
- `listen`
- `accept`
- `read`
- `write`
- `close`

The server:

- Creates a TCP listening socket
- Binds the socket to a port
- Creates an `epoll` instance
- Registers the listening socket with `epoll`
- Waits for socket readiness events
- Accepts incoming client connections
- Registers connected clients with `epoll`
- Reads only from sockets reported as ready
- Processes client requests
- Sends responses
- Detects disconnected clients
- Removes disconnected sockets from `epoll`
- Closes unused file descriptors

Unlike the previous TCP servers, the event-driven server does not create a goroutine for each client connection.

### Concepts Covered

- Linux system calls
- File descriptors
- Event-driven programming
- I/O multiplexing
- `epoll`
- Readiness notification
- Socket lifecycle
- Event loops
- Kernel event notification
- Operating system networking

### Architecture

```text
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
   Accept Client        Read Request
         │                   │
         ▼                   ▼
 Register Client       Parse Command
         │                   │
         ▼                   ▼
  Return to Loop      Execute Command
                             │
                             ▼
                       Write Response
                             │
                             ▼
                       Return to Loop
```

---

# Week 4 - RESP

## Objective

Implement the Redis Serialization Protocol (RESP) so the server can understand Redis-compatible wire messages instead of relying on the original plain-text command format.

The goal is to separate networking, protocol parsing, command execution, and response serialization into independent layers.

---

## RESP Request Parsing

Implemented a RESP parser capable of interpreting Redis-style requests.

For example, instead of receiving:

```text
SET name Lam
```

a Redis client sends:

```text
*3\r\n
$3\r\n
SET\r\n
$4\r\n
name\r\n
$3\r\n
Lam\r\n
```

The parser converts this representation into an internal request:

```go
Request{
    Command: "SET",
    Args: []string{"name", "Lam"},
}
```

The rest of the application therefore does not need to know how the request was represented on the network.

### RESP Types

Basic parsing support was implemented for RESP data types including:

- Arrays (`*`)
- Bulk Strings (`$`)
- Simple Strings (`+`)
- Errors (`-`)
- Integers (`:`)

---

## Byte Consumption

RESP values have variable lengths.

For example:

```text
$7\r\nCOMMAND\r\n
```

contains more than just the seven bytes making up `COMMAND`.

The parser therefore tracks how many bytes each RESP value consumes so arrays containing multiple RESP values can be parsed sequentially.

Conceptually:

```text
RESP bytes
    │
    ▼
Parse value
    │
    ├── Parsed value
    │
    └── Bytes consumed
```

This allows the parser to advance through RESP arrays correctly.

---

## RESP Response Encoding

Implemented a separate response encoder to serialize internal server responses back into RESP.

This keeps Redis protocol formatting separate from command execution.

### Simple String

Internal response:

```text
OK
```

RESP:

```text
+OK\r\n
```

### Bulk String

Internal response:

```text
Lam
```

RESP:

```text
$3\r\nLam\r\n
```

### Error

Internal error:

```text
ERR unknown command
```

RESP:

```text
-ERR unknown command\r\n
```

The response layer supports:

- Simple Strings
- Bulk Strings
- Errors
- Integers
- Null responses
- Arrays

---

# Project Structure

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
    │   ├── resp_parser.go
    │   ├── response.go
    │   └── server.go
    │
    └── io_multiplexing
        └── main.go
```

---

# Running

## Week 1

```bash
go run ./week1
```

## Week 2 — Goroutine-per-Connection Server

```bash
go run ./week2/tcp_server
```

## Week 2 — Worker Pool Server

```bash
go run ./week2/thread_pool
```

## Week 3/4 — epoll + RESP Server

The `epoll` implementation requires Linux.

On Windows, it can be run through WSL:

```bash
go run ./week3/io_multiplexing
```

---

# Testing with Redis CLI

Install Redis CLI:

```bash
sudo apt install redis-tools
```

Connect to the server:

```bash
redis-cli -p 8080
```

Basic supported commands include:

```text
PING
SET name Lam
GET name
```

---

# What I'm Learning so far

This repository documents my progress building a Redis-like database while exploring the systems concepts behind high-performance network servers.

Topics covered so far include:

- Concurrent programming in Go
- Goroutines and channels
- Mutex-based synchronization
- Race conditions and critical sections
- TCP networking
- Long-lived TCP connections
- Worker pools
- Producer-consumer architecture
- Linux system calls
- File descriptors
- Event-driven programming
- I/O multiplexing with `epoll`
- Kernel readiness notification
- RESP protocol parsing
- Binary-safe length-prefixed protocols
- Protocol serialization
- TCP stream semantics
- Request/response architecture
- In-memory key-value storage
- Redis command execution
- High-performance server architecture
- Systems programming concepts

---

## Next Steps

The next milestones will focus on making the Redis-like database more robust and feature complete:

- Per-client input buffering
- Handling fragmented TCP requests
- Handling multiple RESP commands per read
- Redis command pipelining
- Concurrent data structures
- Thread-safe key-value storage
- Key expiration (TTL)
- Persistence
- Replication
