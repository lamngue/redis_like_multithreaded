package main

import (
	"fmt"
	"redis-like-multithreaded/week1/thread_safe_counter"
	"sync"
)

func main() {
	var counter thread_safe_counter.Counter
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() { // starts a new goroutine
			defer wg.Done()
			counter.Increment()
		}()
	}

	wg.Wait()

	fmt.Println(counter.Value())
}
