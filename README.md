# Redis Clone in Go

A Redis-like in-memory key-value store built from scratch in Go as part of a systems programming course.

The goal of this project is to understand how high-performance, concurrent servers work internally by implementing the core building blocks from scratch, including synchronization, networking, and efficient data structures.

## Project Roadmap

- [x] Thread-safe counter using `sync.Mutex`
- [ ] Concurrent queue
- [ ] Thread pool
- [ ] TCP server
- [ ] Redis Serialization Protocol (RESP) parser
- [ ] Command handling (`PING`, `GET`, `SET`)
- [ ] In-memory key-value store
- [ ] Concurrent client handling
- [ ] Key expiration (TTL)
- [ ] Persistence
- [ ] Replication

---

## Week 1 – Thread-safe Counter

### Objective

Implement a thread-safe counter that can be safely incremented by multiple goroutines using Go's `sync.Mutex`.

### Concepts Covered

- Goroutines
- `sync.Mutex`
- `sync.WaitGroup`
- Race conditions
- Critical sections
- Mutual exclusion

### Implementation

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

## Project Structure

```text
.
├── go.mod
├── README.md
└── week1
    ├── main.go
    └── thread_safe_counter
        └── thread_safe_counter.go
```

---

## Running

From the project root:

```bash
go run ./week1
```

---

## What I'm Learning

This repository documents my progress as I build a Redis-like server from scratch. Along the way I'm learning about:

- Concurrent programming in Go
- Synchronization primitives
- Network programming
- Protocol design
- High-performance server architecture
- Systems programming concepts

The repository will continue to evolve as more Redis features are implemented.
