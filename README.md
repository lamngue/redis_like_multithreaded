# Redis-like Database (Work in Progress) in Go

A Redis-like in-memory key-value database built from scratch in Go as part of a systems programming learning project.

The goal of this project is to understand how high-performance, concurrent servers work internally by implementing the core building blocks from scratch, including synchronization, networking, concurrency, operating system I/O, protocol parsing, and data storage.

Rather than relying entirely on Go's high-level networking abstractions, parts of the project explore lower-level Linux APIs such as `epoll` to understand how event-driven servers manage large numbers of connections efficiently.

---

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

```text
Expected Output:
100
```

---

# Week 2 – TCP Servers

## Objective

Learn how concurrent network servers work by implementing two different TCP server architectures.

---

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

---

## 2. Worker Pool TCP Server

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

# Week 3 – I/O Multiplexing Server (`epoll`)

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

# Week 4 – Redis Serialization Protocol (RESP)

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

## Request / Response Architecture

The server now follows a layered request lifecycle:

```text
             TCP Socket
                 │
                 ▼
          syscall.Read()
                 │
                 ▼
            RESP Parser
                 │
                 ▼
              Request
                 │
                 ▼
         ExecuteCommand()
                 │
                 ▼
             Response
                 │
                 ▼
          RESP Encoder
                 │
                 ▼
          syscall.Write()
                 │
                 ▼
              Client
```

This separates four responsibilities:

1. Networking
2. Protocol parsing
3. Command execution
4. Protocol serialization

---

## Redis CLI Integration

The server can now communicate with `redis-cli` using RESP.

When `redis-cli` connects, Redis protocol requests such as:

```text
*2\r\n$7\r\nCOMMAND\r\n$4\r\nDOCS\r\n
```

can be received and parsed by the server.

The implementation also demonstrates how Redis clients may automatically send capability/discovery commands such as:

```text
COMMAND DOCS
INFO SERVER
COMMAND
```

during client initialization.

Not all Redis commands are implemented yet, so unsupported commands return an error response.

---

## TCP Framing Limitation

The current implementation provides basic RESP parsing but does not yet implement full per-client input buffering.

TCP is a byte stream and does not guarantee that one `read()` call corresponds to exactly one Redis command.

For example, a command could arrive as:

```text
Read 1:

*3\r\n$3\r\nSE
```

followed by:

```text
Read 2:

T\r\n$4\r\nname\r\n...
```

Similarly, multiple commands could arrive in a single read.

A future milestone will introduce per-client buffers to correctly handle:

- Partial requests
- Multiple requests in one read
- Pipelined Redis commands

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

# What I'm Learning

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
