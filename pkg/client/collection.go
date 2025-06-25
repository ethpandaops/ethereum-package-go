package client

import "sync"

// Collection is a generic collection for any client type
type Collection[T any] struct {
	clients map[Type][]T
	mu      sync.RWMutex
}

// NewCollection creates a new client collection
func NewCollection[T any]() *Collection[T] {
	return &Collection[T]{
		clients: make(map[Type][]T),
	}
}

// Add adds a client to the collection
func (c *Collection[T]) Add(clientType Type, client T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients[clientType] = append(c.clients[clientType], client)
}

// All returns all clients in the collection
func (c *Collection[T]) All() []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var all []T
	for _, clients := range c.clients {
		all = append(all, clients...)
	}

	return all
}

// ByType returns all clients of a specific type
func (c *Collection[T]) ByType(clientType Type) []T {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to avoid data races
	result := make([]T, len(c.clients[clientType]))
	copy(result, c.clients[clientType])

	return result
}

// Count returns the total number of clients
func (c *Collection[T]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	count := 0
	for _, clients := range c.clients {
		count += len(clients)
	}

	return count
}

// CountByType returns the number of clients of a specific type
func (c *Collection[T]) CountByType(clientType Type) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.clients[clientType])
}

// Types returns all client types present in the collection
func (c *Collection[T]) Types() []Type {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var types []Type
	for clientType := range c.clients {
		types = append(types, clientType)
	}

	return types
}
