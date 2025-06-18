package client

// Collection is a generic collection for any client type
type Collection[T any] struct {
	clients map[Type][]T
}

// NewCollection creates a new client collection
func NewCollection[T any]() *Collection[T] {
	return &Collection[T]{
		clients: make(map[Type][]T),
	}
}

// Add adds a client to the collection
func (c *Collection[T]) Add(clientType Type, client T) {
	c.clients[clientType] = append(c.clients[clientType], client)
}

// All returns all clients in the collection
func (c *Collection[T]) All() []T {
	var all []T
	for _, clients := range c.clients {
		all = append(all, clients...)
	}
	return all
}

// ByType returns all clients of a specific type
func (c *Collection[T]) ByType(clientType Type) []T {
	return c.clients[clientType]
}

// Count returns the total number of clients
func (c *Collection[T]) Count() int {
	count := 0
	for _, clients := range c.clients {
		count += len(clients)
	}
	return count
}

// CountByType returns the number of clients of a specific type
func (c *Collection[T]) CountByType(clientType Type) int {
	return len(c.clients[clientType])
}

// Types returns all client types present in the collection
func (c *Collection[T]) Types() []Type {
	var types []Type
	for clientType := range c.clients {
		types = append(types, clientType)
	}
	return types
}
