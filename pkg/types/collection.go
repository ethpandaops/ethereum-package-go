package types

// ClientCollection is a generic collection for any client type
type ClientCollection[T any] struct {
	clients map[ClientType][]T
}

// NewClientCollection creates a new client collection
func NewClientCollection[T any]() *ClientCollection[T] {
	return &ClientCollection[T]{
		clients: make(map[ClientType][]T),
	}
}

// Add adds a client to the collection
func (c *ClientCollection[T]) Add(clientType ClientType, client T) {
	c.clients[clientType] = append(c.clients[clientType], client)
}

// All returns all clients in the collection
func (c *ClientCollection[T]) All() []T {
	var all []T
	for _, clients := range c.clients {
		all = append(all, clients...)
	}
	return all
}

// ByType returns all clients of a specific type
func (c *ClientCollection[T]) ByType(clientType ClientType) []T {
	return c.clients[clientType]
}

// Count returns the total number of clients
func (c *ClientCollection[T]) Count() int {
	count := 0
	for _, clients := range c.clients {
		count += len(clients)
	}
	return count
}

// CountByType returns the number of clients of a specific type
func (c *ClientCollection[T]) CountByType(clientType ClientType) int {
	return len(c.clients[clientType])
}

// Types returns all client types present in the collection
func (c *ClientCollection[T]) Types() []ClientType {
	var types []ClientType
	for clientType := range c.clients {
		types = append(types, clientType)
	}
	return types
}