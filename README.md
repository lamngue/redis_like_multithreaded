# Redis Clone in Go

## What is this?

This project purpose is to (attempt to) create a redis-like in-memory key-value database from scratch using Golang.
I want to understand how high-performance and concurrent server works internally by implementing synchronization, networking, concurrency, operating system I/O, protocol parsing, and data storage from scratch.
This repo will document my progress so far (currently week 5 - TTL and thread-safe storage).
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
- [x] Concurrent key-value store (`sync.RWMutex`)
- [x] Key expiration (TTL)
- [ ] Persistence (RDB)
- [ ] Replication


# Week 1 - Thread-safe Counter
Why do I neeed a mutex? To prevent multiple threads from changing or reading the same data at the same time. (avoiding race conditions and data corruptions) <br>
What race condition am I preventing? Multiple threads (goroutine) are trying to increment a counter value <br>
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

What happens when a client connects? The server listens for incoming TCP connections <br>
Why does `go handleConnection(conn)` allow multiple clients? The `go` keyword starts the function in a new goroutine. This allows the main loop to listening for another client via `Accept()` <br>

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
What problem does the worker pool solve? To prevent creating a new goroutine for every connection <br>
What is the producer and what is the consumer? Producer - generate jobs (in this case incoming connections). Consumer - fixed number of goroutines pulling elements from the jobs channel. <br>
Implemented an alternative server architecture using a fixed-size worker pool. <br>

Instead of creating a new goroutine for every connection, incoming client connections are placed onto a shared job queue and processed by a fixed number of worker goroutines. <br>

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
What exactly is a file descriptor? File descriptors (FDs) are part of the POSIX API and use basic integers to determine state. It is a handle to access IO/file resource at kernel level. <br>
What does `epoll_create1()` create? It creates a new epoll instance in the Linux kernel and returns an integer file descriptor referring to that instance. <br>
Why do I register the listener with epoll? To bypass the kernel-level blocking behavior of synchronous I/O. <br>
Why does `epoll_wait()` tell me which sockets are ready?  So the program can process active connections instantly without wasting CPU time checking thousands of idle sockets one by one <br>
Why don't I need a goroutine per connection? One single thread or loop uses `epoll` to watch many connections at the same time and only wakes up when a connection is actually ready to read or write data <br>
What's the difference between the listener FD and a client FD? Listener FD accepts incoming connection requests from a specific port, client FD is the active, dedicated data channel created for a specific, connected client <br>
What happens when `read()` returns 0? The end-of-file (EOF) has been reached <br>

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

# Week 5 - Key Expiration and Thread-Safe Store

## Objective

Extend the Redis-like server with key expiration and a thread-safe in-memory key-value store.

The main goals were:

- Add key expiration using TTL
- Implement `EXPIRE`, `TTL`, `EXISTS`, and `DEL`
- Automatically remove expired keys when they are accessed
- Protect the shared key-value store from concurrent access using `sync.RWMutex`
- Separate command handling from storage and synchronization logic

---

## Key Expiration

The original key-value store only stored a string value:

```text
key -> value
```
To support expiration, each key now stores an Entry containing both its value and an expiration timestamp.

```text
type Entry struct {
    Value     string
    ExpiresAt int64
}
```
ExpiresAt stores a Unix timestamp in seconds.

A value of 0 means that the key does not expire.

Conceptually:

```text
key
 │
 ▼
Entry
 ├── Value
 └── ExpiresAt
```
---
## Lazy Expiration

The server uses lazy expiration.

Instead of continuously scanning the entire key-value store, the server checks whether a key has expired when it is accessed.

For example:
```text
SET name Lam
EXPIRE name 5
```
After five seconds, the key is considered expired.

When a command accesses the key:

```text
Access key
    │
    ▼
Does it exist?
    │
    ▼
Has it expired?
   / \
 yes  no
  │    │
  ▼    ▼
delete return
  │    key
  ▼
not found
```
Expired entries are removed from the store when they are detected.

This behavior is used by commands such as `GET`, `EXISTS`, `TTL`, `EXPIRE`, and `DEL`.

---
## Supported Commands
### `EXPIRE`

Sets an expiration time for an existing key.
```text
EXPIRE name 10
```
Returns:
```text
(integer) 1
```
if the expiration was successfully set, or:
```text
(integer) 0
```
if the key does not exist.

Calling `EXPIRE` again replaces the previous expiration time.
---
### TTL

Returns the remaining lifetime of a key in seconds.
```text
TTL name
```
The implementation uses:
```text
-1 -> key exists but has no expiration
-2 -> key does not exist or has expired
N  -> approximately N seconds remaining
```
Example:
```text
SET name Lam
EXPIRE name 10
TTL name
```
The returned value decreases as time passes.
---
### EXISTS

Checks whether a key currently exists.
```text
EXISTS name
```
Returns:
```text
(integer) 1
```
for an existing key and:
```text
(integer) 0
```

for a missing or expired key.
---
### DEL

Deletes a key from the store.
```text
DEL name
```
Returns:
```text
(integer) 1
```
when a key is deleted and:
```text
(integer) 0
```
when the key does not exist or has already expired.
---
### Thread-Safe Store
The original implementation used a shared map directly:
```text
var store = map[string]Entry{}
```
This becomes unsafe when multiple goroutines access the map concurrently.

The key-value store was therefore refactored into a `Store` type:
```text
type Store struct {
    mu   sync.RWMutex
    data map[string]Entry
}
```
The store is now responsible for both:

* managing the key-value data
* synchronizing concurrent access

The command handlers no longer directly manipulate the map.

Instead, the architecture is:
```text
Command
   │
   ▼
Handler
   │
   ▼
Store method
   │
   ▼
Mutex
   │
   ▼
Map

```

For example
```text
SET
 │
 ▼
handleSet()
 │
 ▼
store.Set()
 │
 ▼
store.data
```
---
### Synchronization
Operations that modify the store use an exclusive lock:

```text
s.mu.Lock()
defer s.mu.Unlock()
```
Expiration-aware operations also use an exclusive lock because checking an expired key can result in deleting it.

For example:

```text
GET
 │
 ├── read entry
 │
 ├── check expiration
 │
 └── delete if expired
```

Although `RWMutex` supports concurrent readers through `RLock`, the current implementation prioritizes correctness because some operations that appear to be reads can also modify the map through lazy expiration.

---
### Concurrency Testing

The store was tested directly using Go tests rather than only through redis-cli.

The tests cover:

* basic SET / GET operations
* EXISTS
* TTL
* DEL
* key expiration
* resetting expiration
* concurrent access from multiple goroutines

The race detector was used to check for data races:

```text
go test -race ./...
```
The tests completed successfully without race detector warnings.
---
### What I learned
Week 5 showed that making a data structure thread-safe is not just about adding a mutex around individual map operations.

The important part is making related operations atomic.

For example, expiration checking involves:

```text
lookup key
    ↓
check expiration
    ↓
possibly delete key
```
These operations must happen while holding the appropriate lock so another goroutine cannot modify the same key between the check and the deletion.

This also led to separating the key-value store from the command handlers, giving each part a clearer responsibility.

---
# Project Structure
`EXPIRE`
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
│   ├── tcp_server
│   └── thread_pool
│
├── week3+4
│   ├── server
│   │   ├── handlers.go
│   │   ├── parser.go
│   │   ├── resp_parser.go
│   │   ├── response.go
│   │   └── server.go
│   └── io_multiplexing
│       └── main.go
│
└── week5
    ├── server
    │   ├── handlers.go
    │   ├── parser.go
    │   ├── resp_parser.go
    │   ├── response.go
    │   ├── server.go
    │   └── store_test.go
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

The next milestones will focus on making the server more robust and adding database features:

- Per-client input buffering
- Handling fragmented TCP requests
- Handling multiple RESP commands per read
- Redis command pipelining
- Concurrent queue
- Persistence
- Replication
