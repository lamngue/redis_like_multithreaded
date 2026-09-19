package server

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func newTestStore() *Store {
	return &Store{
		data: make(map[string]Entry),
	}
}

func TestStoreBasicOperations(t *testing.T) {
	store := newTestStore()

	// SET
	store.Set("name", "Lam")

	// GET
	value, exists := store.Get("name")
	if !exists {
		t.Fatal("expected key to exist")
	}
	if value != "Lam" {
		t.Fatalf("expected value %q, got %q", "Lam", value)
	}

	// EXISTS
	if !store.Exists("name") {
		t.Fatal("expected key to exist")
	}

	// TTL without expiration
	ttl := store.TTL("name")
	if ttl != -1 {
		t.Fatalf("expected TTL -1, got %d", ttl)
	}

	// DELETE
	deleted := store.Delete("name")
	if !deleted {
		t.Fatal("expected key to be deleted")
	}

	// Key should no longer exist
	if store.Exists("name") {
		t.Fatal("expected key not to exist after deletion")
	}

	// DELETE nonexistent key
	deleted = store.Delete("name")
	if deleted {
		t.Fatal("expected deleting nonexistent key to return false")
	}
}

func TestStoreExpiration(t *testing.T) {
	store := newTestStore()

	store.Set("foo", "bar")

	// Set a very short expiration.
	if !store.SetExpiration("foo", 1) {
		t.Fatal("expected SetExpiration to succeed")
	}

	ttl := store.TTL("foo")
	if ttl < 0 || ttl > 1 {
		t.Fatalf("expected TTL between 0 and 1, got %d", ttl)
	}

	// Wait until the key should definitely be expired.
	time.Sleep(2 * time.Second)

	// GET should treat it as nonexistent.
	_, exists := store.Get("foo")
	if exists {
		t.Fatal("expected expired key not to exist")
	}

	// EXISTS should also return false.
	if store.Exists("foo") {
		t.Fatal("expected expired key not to exist")
	}

	// TTL should return -2.
	ttl = store.TTL("foo")
	if ttl != -2 {
		t.Fatalf("expected TTL -2 for expired key, got %d", ttl)
	}

	// DELETE should return false because the key is already expired.
	if store.Delete("foo") {
		t.Fatal("expected deleting expired key to return false")
	}
}

func TestStoreConcurrentAccess(t *testing.T) {
	store := newTestStore()

	const goroutines = 100

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()

			key := fmt.Sprintf("key-%d", i)
			value := fmt.Sprintf("value-%d", i)

			store.Set(key, value)

			_, _ = store.Get(key)
			_ = store.Exists(key)

			store.SetExpiration(key, 10)
			_ = store.TTL(key)

			_ = store.Delete(key)
		}(i)
	}

	wg.Wait()
}
